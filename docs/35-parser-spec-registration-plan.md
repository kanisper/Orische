# Parser Feature Boundary Implementation Record

## Status

The parser package-boundary refactoring is implemented. This document records
the resulting internal API and the constraints preserved during the change.
Tests and implementation remain the source of truth.

## Outcome

The former single-package registration machinery was divided into:

- `feature`, containing neutral implementation contracts;
- `syntax/block` and `syntax/inline`, containing built-in definitions;
- `syntax`, assembling the built-in language;
- `parser`, compiling a language and orchestrating source-to-AST processing.

No stable extension API, dynamic plugin mechanism, or new accepted syntax was
introduced.

## Contract Decisions

### Immutable language construction

Definitions are supplied as a `feature.Language` to `NewParser`. The frontend
validates all definitions and compiles private registries before parsing. It
does not expose post-construction registration methods.

Paragraph is represented explicitly because its Reader is fixed frontend
infrastructure while its AST builder is a syntax implementation. Its typed
definition returns `*ast.Paragraph`, must declare `paragraph`, and must not
implement `BlockReader`.

### Transactional block reading

Readers no longer receive a mutable parser cursor. `BlockInput` exposes lines by
relative offset, and a successful Reader returns a positive consumed-line count.
The frontend rejects inconsistent results before advancing.

This replaces the former contract that a Reader restore `blockContext.pos` after
rejection. No-match is now structurally unable to consume frontend state.

### Capability-based building

Builders no longer receive `*Parser`. `BuildContext` exposes only recursive
inline parsing and Block building. This is enough for Heading, Paragraph, and
List while preventing syntax packages from depending on parser internals.

### IR ownership

`feature.TextBlock` is shared because fixed frontend readers create nodes that
syntax builders consume. It carries the raw text's source origin separately from
the complete Block range. List IR remains private to `syntax/block`; only the
neutral `BlockNode` interface crosses the boundary.

### Inline ownership

The common inline parser continues to own delimiters, recursion, fallback, and
ranges. Definitions own content policy, semantic attribute validation, and AST
construction. Structurally incomplete candidates are never validated.

## Preserved Behavior

- Block Reader precedence remains Directive, Heading, List, Paragraph.
- Malformed Block candidates still fall through to Paragraph.
- Lists retain dedicated recursive reading and raw-level normalization.
- List-item construction still uses common Block dispatch.
- Heading, Paragraph, and list-item Paragraph content use the active Inline set.
- Code Block content remains literal.
- Type lookup remains case-insensitive while values preserve spelling.
- Inline semantic rejection and malformed-candidate fallback are unchanged.
- Unicode code-point range calculation is unchanged.
- Builder diagnostics preserve identity; ordinary errors accumulate context.

## Validation Coverage

The boundary tests verify:

- Block and Inline Directive Definitions can be implemented outside package `parser`;
- external Block builders can call active inline parsing;
- invalid Reader result combinations are rejected;
- invalid, duplicate, nil, and reserved definitions fail at construction;
- caller mutation of Language slices does not replace compiled registries;
- active definitions propagate through recursive inline parsing and nested lists;
- diagnostic identity and nested ordinary-error context are preserved;
- built-in syntax behavior and ranges remain unchanged.

Run:

```sh
go test ./internal/parser/...
```

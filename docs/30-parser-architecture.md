# Parser Architecture

## Pipeline

```text
input string
  -> splitLines
  -> fixed and registered Block Readers
  -> feature.BlockNode IR
  -> registered Block Builders and inline parser
  -> ast.Document
```

Rendering is outside the parser packages.

`Parser` owns one private compiled specification for the complete source-to-AST
operation. The same specification is used for top-level blocks, nested lists,
inline-capable blocks, and recursive inline directives.

## Package Model

```text
parser -> feature
parser -> syntax -> syntax/block -> feature
                 -> syntax/inline -> feature
feature -> ast
syntax/* -> ast
```

`internal/parser/feature` is a leaf implementation-API package. It defines
neutral contracts and transport values but does not own scanning, registration,
dispatch, or built-in syntax.

`internal/parser/syntax` assembles the replaceable built-in definitions in
`feature.Language`.
`syntax/block` and `syntax/inline` implement the built-in language without
importing the parser frontend.

The feature contracts are internal package boundaries, not a stable plugin API.
Because AST interfaces have private marker methods, a syntax package can build
the existing AST nodes but cannot independently introduce a new AST node type.

## Language Compilation

`feature.Language` declares:

- ordered general Block Definitions;
- Inline Directive Definitions.

`NewParser` installs the fixed Paragraph definition, then validates and compiles
the declaration into private maps and an ordered Sugar-reader slice. `Parser`
requires this constructor; its zero value returns an initialization error.
Language compilation rejects nil and empty definitions, normalized duplicates,
including attempts to register `paragraph`, and invalid Inline content policies.
The caller's slices are not retained as registries.

Syntax Type registration and lookup use Unicode-aware `strings.ToLower`.
Normalization applies only to the Type. Attributes and content retain their
original spelling. Block and Inline registries are independent, so `code` is
valid in both categories.

There are no registration mutators after construction. Definition objects must
not change their Type or content policy after being passed to `NewParser`, and
stateful definitions are responsible for concurrency safety.

## Block Reader Chain

The effective order is fixed:

1. common Block Directive envelope reader;
2. Reader-capable definitions in `Language.Blocks` order;
3. common Paragraph fallback reader.

The Directive reader, Paragraph reader, and Paragraph AST builder belong to the
frontend. The parser installs the Paragraph definition independently of
`feature.Language` before general Block definitions. Paragraph is therefore
always available to common Block dispatch and cannot be omitted or replaced.

A Reader receives a read-only `feature.BlockInput` at the current document
position. It returns `feature.BlockReadResult`:

```text
no match: Matched=false, Consumed=0, Node=nil
match:    Matched=true, 1 <= Consumed <= input.Len(), Node!=nil
```

The frontend validates the result before advancing. This makes rejection
transactional without exposing or restoring a mutable cursor. When a Sugar
Reader matches, the frontend also checks that its declared Type equals the
Node's Type after normalization.

Malformed candidates normally fall through to Paragraph. A valid Block
Directive with an unregistered Type is read into `feature.TextBlock`, then
returns an unsupported-block diagnostic during builder dispatch.

## Parsed-Block IR

`feature.BlockNode` is the common IR interface and exposes only Block Type and
source range. `feature.TextBlock` is the shared IR for Paragraph, Heading, and
Block Directive text. It carries both the complete Block range and the source
origin of its text, so builders never reconstruct frontend delimiter offsets.

List IR remains private to `syntax/block`. The frontend transports it only as a
`feature.BlockNode`; the List builder restores its private concrete type. This
keeps list representation out of the neutral API.

Block reading records structure and raw inline-capable text. It does not create
AST blocks or parse inline syntax.

## AST Building

`Parser.buildBlock` owns normalized lookup, unsupported-block diagnostics,
ordinary error wrapping, and diagnostic preservation. Builders receive a
`feature.BuildContext`, not `*Parser`.

The context exposes two capabilities:

- `ParseInlines`, used by Heading and Paragraph;
- `BuildBlock`, used by List for paragraph-like item nodes and nested lists.

Code Block builders preserve content literally and do not call inline parsing.
A diagnostic returned by any nested builder is propagated without build-context
wrapping. Ordinary errors retain their cause and accumulate inner and outer
Block Type context.

## List Parsing

List syntax reads a consecutive list run from `BlockInput` and reports the run's
line count as `Consumed`. It never invokes the document Reader chain for item
content.

Raw marker levels are normalized as follows:

```text
first line: logical = 1
raw increase: logical = previous logical + 1
same raw level: logical = previous logical
raw decrease: logical = max(1, previous logical - raw difference)
```

Each item contains a paragraph-like `feature.TextBlock` and may contain a private
nested List Node. Both use the active BuildContext during AST construction.

## Inline Parsing

The frontend scanner owns the common `:[...]{...}` envelope, header splitting,
byte-offset scanning, brace handling, recursion, text flushing, fallback spans,
and source ranges.

An Inline Directive Definition owns:

- normalized identity through `InlineType`;
- nested or literal content policy;
- attribute validation;
- AST construction from a closed candidate.

The frontend confirms structural closure before validation. `false, nil` from
validation means semantic rejection and literal fallback; a non-nil error aborts
parsing. Recursive parsing always uses the same compiled language.

## Range and Error Contracts

Parser-produced non-empty ranges are one-based and inclusive. Columns count
Unicode code points rather than UTF-8 bytes. Directive-node ranges cover the
complete source syntax; nested nodes and Text nodes carry their own spans.

Malformed syntax normally falls back. Errors are reserved for unsupported valid
Block Directives and inconsistent internal states such as invalid Reader results,
IR type mismatches, definition failures, or nil AST results.

# Parser Module Guide

## Responsibility

Convert input text into the AST through two internal stages:

1. Read document blocks into `feature.BlockNode` IR.
2. Dispatch IR to syntax builders and parse inline content where applicable.

The parser frontend lives in `internal/parser`. Neutral implementation contracts
live in `internal/parser/feature`; built-in implementations live under
`internal/parser/syntax`.

## Package Boundaries

- `parser` owns orchestration, fixed readers, compiled registries, dispatch,
  inline scanning, fallback, ranges, and error wrapping.
- `feature` owns only syntax-neutral interfaces and shared transport values.
- `syntax/block` owns Heading, Paragraph, List, and Code Block definitions.
- `syntax/inline` owns Emphasis, Link, and Code Span definitions.
- Syntax packages must not import `parser`. Builders use `feature.BuildContext`.
- These are internal implementation APIs, not a stable plugin API.

## Block Parsing

Document Reader order is:

1. fixed Block Directive reader;
2. registered Sugar readers in `feature.Language.Blocks` order;
3. fixed Paragraph fallback reader.

Readers receive an immutable `feature.BlockInput` and return a
`feature.BlockReadResult`. For no match, `Matched` is false, `Consumed` is zero,
and `Node` is nil. For a match, `Consumed` is within the available line count and
`Node` is non-nil. The frontend validates these invariants before advancing.

The Block Directive and Paragraph readers are fixed frontend infrastructure.
The typed Paragraph builder is supplied separately by
`feature.Language.Paragraph`; `paragraph` is a case-insensitive reserved Block
Type.

Malformed Heading, List, and Block Directive candidates fall through to the
Paragraph reader. A syntactically valid directive still requires a registered
definition.

## Lists

Lists use dedicated recursive reading rather than the document Reader chain.
During AST construction, list-item Paragraph and nested List nodes use
`feature.BuildContext.BuildBlock`, preserving common dispatch and error behavior.

Marker-run length is a raw level. `normalizeListLevel` converts changes into
logical levels. Any raw increase adds exactly one logical level, regardless of
the increase size. Nesting depth has no explicit limit.

## Inline Parsing

The frontend owns the `:[...]{...}` envelope, recursion, fallback, scanning, and
range calculation. Inline definitions own type-specific attribute validation,
content policy, and AST construction.

Empty inline content is valid, and links require a nonempty URI attribute.
Unsupported directives, invalid headers, and links without a URI are emitted as
literal source text rather than errors. Other malformed or unterminated
candidates resume ordinary scanning.

Attribute validation has three outcomes: `true, nil` accepts; `false, nil` uses
literal fallback; a non-nil error aborts parsing. The frontend must confirm that
a candidate is structurally closed before invoking validation.

## AST, Diagnostics, and Ranges

AST block and inline interfaces use pointer implementations. Parser-produced
non-empty ranges are one-based and inclusive; columns count Unicode code points.
Every inline AST node carries a range.

Sugar definitions must produce a Node whose normalized Block Type matches the
definition. Builder diagnostics preserve identity, message, and range. Ordinary
builder errors retain their cause and receive nested build context.

## Validation

```sh
go test ./internal/parser/...
```

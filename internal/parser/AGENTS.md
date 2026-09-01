# Parser Module Guide

## Responsibility

Convert input text into the AST through two local stages:

1. read document blocks into private parsed-block nodes;
2. build AST blocks and parse inline content where applicable.

All implementation files are in `package parser`. Files are split by syntax and responsibility for readability; there are no parser feature or syntax packages.

## Internal State

`Parser` owns a private `spec` containing:

- Block Directive builders;
- ordered Heading and List sugar readers;
- inline definitions for directive content policy, optional validation, and AST building.

`spec` is parser configuration, not a public language or plugin API. `newSpec` installs the built-in `heading`, `paragraph`, and `code` Block Directive builders and the `em`, `link`, and `code` inline definitions. Keep definition-specific logic close to the syntax file that uses it.

Block readers hand private concrete nodes to `Parser.buildBlock`. The private `parsedBlock` interface carries only the source range needed by orchestration; syntax-specific fields remain on concrete node types. Avoid introducing generic contracts when a private concrete type expresses the state directly.

## Syntax Model

Block Directives are the explicit named form for supported block semantics. Heading marker syntax is sugar, ordinary nonblank text is the Paragraph fallback form, and List marker syntax is currently the only List notation.

Attributes are opaque strings. Interpret or validate an attribute only when the syntax needs it semantically. Attributes that a syntax does not use are accepted and ignored.

Current required or meaningful attributes include:

- Heading Block Directive: `level1` through `level6`; invalid or missing levels are semantic build errors;
- Code Block Directive: language identifier;
- Link inline directive: nonempty URI, with invalid candidates falling back to literal inline text.

Paragraph, Emphasis, and Code Span attributes are accepted and ignored.

Do not introduce a separate validation package or Reader -> Builder -> Validator pipeline without a concrete need. Current validation remains local to the relevant syntax implementation.

## Block Parsing

Block reader order is:

1. common Block Directive reader;
2. Heading and List sugar readers in `spec.sugars` order;
3. Paragraph fallback reader.

The Block Directive reader, Paragraph reader, and Paragraph builder are fixed parser infrastructure. A reader receives immutable-by-convention `blockContext` and returns a node plus a positive consumed-line count on a match. A non-match returns a nil node and zero consumption.

Malformed Heading sugar, List sugar, and Block Directive envelopes fall through to the Paragraph reader. A structurally valid Directive still requires a built-in builder and produces an unsupported-directive diagnostic when no builder exists. A supported Directive may still produce an ordinary build error when a required semantic value is invalid, such as an invalid Heading level.

Heading Directive and Heading sugar share final Heading construction. Paragraph Directive and Paragraph fallback share final Paragraph construction.

## Lists

Lists use dedicated recursive reading rather than the document block reader chain. `normalizeListLevel` converts raw marker runs into logical levels: a raw increase adds one level, an unchanged level remains unchanged, and a decrease subtracts the raw difference with a minimum of one. Nesting has no explicit depth limit.

Mixed-marker lines are valid. A line's first marker determines its candidate ordered/unordered style, and the first item at each logical List level determines the style of the constructed List at that level.

List item Paragraphs and nested Lists are built through `Parser.buildBlock`, so they retain ordinary inline parsing, diagnostics, and ranges.

## Inline Parsing

The frontend owns the `:[...]{...}` envelope, recursion, fallback, scanning, and range calculation. Inline definitions own type-specific attribute validation when needed, content policy, and AST construction.

Empty inline content is valid, and Links require a nonempty URI attribute. Unsupported directives, invalid headers, and Links without a URI are emitted as literal source text. Other malformed or unterminated candidates resume ordinary scanning. Definitions are validated only after the frontend confirms that the candidate is structurally closed.

`inlineDefinition.build` returns an AST node directly rather than an error. Internal parser errors cover invalid definition state such as an unknown content policy, a missing builder function, or a nil AST node returned by a builder.

## AST, Diagnostics, and Ranges

AST block and inline interfaces use pointer implementations. Parser-produced non-empty ranges are one-based and inclusive; columns count Unicode code points. Every inline AST node carries a range. Inline Directive ranges include the complete `:[...]{...}` syntax; nested content and literal text nodes carry their own source spans.

Builder diagnostics preserve identity, message, and range. Ordinary builder errors retain their cause and receive block context. Code Block content remains literal and does not invoke inline parsing.

## Validation

```sh
go test ./internal/parser/...
```

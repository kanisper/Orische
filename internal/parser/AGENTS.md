# Parser Module Guide

## Responsibility

Convert input text into the AST through two local stages:

1. read document blocks into private parsed-block nodes;
2. build AST blocks and parse inline content where applicable.

All implementation files are in `package parser`. Files are split by syntax and
responsibility for readability; there are no parser feature or syntax packages.

## Internal State

`Parser` owns a private `spec` containing:

- Block Directive builders;
- ordered Heading and List sugar readers;
- inline definitions for directive content policy, validation, and AST building.

`spec` is parser configuration, not a public language or plugin API. `newSpec`
installs the built-in `heading`, `paragraph`, and `code` Block Directive
builders and the `em`, `link`, and `code` inline definitions. Keep the
definition-specific logic close to the syntax file that uses it.

Block readers hand private concrete nodes to `Parser.buildBlock`. The private
`parsedBlock` interface carries only the source range needed by orchestration;
syntax-specific fields remain on concrete node types. Avoid introducing generic
contracts when a private concrete type expresses the state directly.

## Block Parsing

Block reader order is:

1. common Block Directive reader;
2. Heading and List sugar readers in `spec.sugars` order;
3. Paragraph fallback reader.

The Block Directive reader, Paragraph reader, and Paragraph builder are fixed
parser infrastructure. A reader receives immutable-by-convention `blockContext`
and returns a node plus a positive consumed-line count on a match. A non-match
returns a nil node and zero consumption.

Malformed Heading, List, and Block Directive candidates fall through to the
Paragraph reader. A syntactically valid directive still requires a built-in
builder and produces an unsupported-directive diagnostic when no builder exists.

## Lists

Lists use dedicated recursive reading rather than the document block reader
chain. `normalizeListLevel` converts raw marker runs into logical levels: a raw
increase adds one level, an unchanged level remains unchanged, and a decrease
subtracts the raw difference with a minimum of one. Nesting has no explicit
depth limit.

List item paragraphs and nested lists are built through `Parser.buildBlock`, so
they retain ordinary inline parsing, diagnostics, and ranges.

## Inline Parsing

The frontend owns the `:[...]{...}` envelope, recursion, fallback, scanning,
and range calculation. Inline definitions own type-specific attribute
validation, content policy, and AST construction.

Empty inline content is valid, and links require a nonempty URI attribute.
Unsupported directives, invalid headers, and links without a URI are emitted as
literal source text. Other malformed or unterminated candidates resume ordinary
scanning. Definitions are validated only after the frontend confirms that the
candidate is structurally closed.

## AST, Diagnostics, and Ranges

AST block and inline interfaces use pointer implementations. Parser-produced
non-empty ranges are one-based and inclusive; columns count Unicode code points.
Every inline AST node carries a range. Directive-node ranges include the
complete `:[...]{...}` syntax; nested content and literal text nodes carry their
own source spans.

Builder diagnostics preserve identity, message, and range. Ordinary builder
errors retain their cause and receive block context. Code Block content remains
literal and does not invoke inline parsing.

## Validation

```sh
go test ./internal/parser/...
```

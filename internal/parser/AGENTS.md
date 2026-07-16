# Parser Module Guide

## Responsibility

Convert input text into the AST through two internal stages:

1. Parse document blocks into private parsed-block IR.
2. Build AST nodes and parse inline content where applicable.

## Block Parsing

Parser order is:

1. Block directive
2. Heading
3. List
4. Paragraph fallback

The paragraph parser must remain last and must accept every nonblank line when reached.

A successful block parser leaves `blockContext.pos` on the last consumed line. `parseBlocks` advances the cursor once after success. A parser returning `ok=false` must leave the cursor unchanged.

Malformed heading, list, and directive candidates fall through to paragraph parsing. A syntactically valid directive still requires a registered builder.

## Lists

Lists are parsed recursively by `listParser`; do not call the document block parser for list-item content.

Marker-run length is a raw level. `normalizeListLevel` converts changes into logical levels. Any raw increase adds exactly one logical level, regardless of the increase size. Nesting depth has no explicit limit.

## Inline Parsing

Heading, paragraph, and list-item paragraph builders parse inline content. Code blocks preserve content literally.

Malformed or unsupported inline candidates remain literal text. Empty inline content is valid. Links require a nonempty URI attribute.

## AST and Ranges

AST block and inline interfaces use pointer implementations. Source ranges are one-based and inclusive.

## Validation

```sh
go test ./internal/parser
```

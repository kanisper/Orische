# Parser Module Guide

## Responsibility

Convert input text into the AST through two internal stages:

1. Parse document blocks into private parsed-block IR.
2. Build AST nodes and parse inline content where applicable.

## Block Parsing

Document block Reader order is:

1. Block directive
2. Heading
3. List
4. Paragraph fallback

The Paragraph reader must remain last and must accept every nonblank line when reached.

Every extensible block feature declares its type through `blockType() string` and is registered with `registerBlock(definition)`. If the definition also implements `blockReader`, registration appends it to the ordered Sugar reader chain; otherwise it is available only through the common Block Directive envelope reader. Heading and List are reader-capable definitions, while Code is directive-only. The Block Directive reader and Paragraph definition are fixed parser infrastructure: `newSpec` installs the standard Paragraph definition in the block-definition map, and `getReaders` always places the Directive reader first and the Paragraph reader last. `paragraph` is case-insensitive and reserved; general Block registration must reject it without mutating the definition map or reader order. Spec validation must verify that the Paragraph definition exists, including for zero-value or otherwise incomplete Specs.

A successful block reader leaves `blockContext.pos` on the last consumed line. `parseBlocks` advances the cursor once after success. A reader returning `ok=false` must leave the cursor unchanged.

Malformed Heading, List, and Block Directive candidates fall through to the Paragraph reader. A syntactically valid directive still requires a registered definition.

## Lists

Lists are read recursively by `listDefinition`; do not call the document block reader chain for list-item content. During AST construction, list-item blocks use the common `Parser.buildBlock` dispatch.

Marker-run length is a raw level. `normalizeListLevel` converts changes into logical levels. Any raw increase adds exactly one logical level, regardless of the increase size. Nesting depth has no explicit limit.

## Inline Parsing

Heading, Paragraph, and list-item Paragraph builders parse inline content through the active `Parser` and `Spec`. Code blocks preserve content literally.

Every inline feature implements the base `inlineDefinition` contract, declares its type through `inlineType() string`, and is registered with `registerInline(definition)`. Current features also implement `inlineDirectiveDefinition`, which owns validation, nested-versus-literal content policy, and AST construction; definitions without a current parser contract are rejected before mutation. Registered inline Directive Types are matched case-insensitively, while common scanning owns envelopes, recursion, fallback, and ranges. Inline Sugar syntax is not implemented yet, but it can be added as another definition category without changing the registration API.

Empty inline content is valid, and links require a nonempty URI attribute. Unsupported directives, invalid headers, and links without a URI are emitted as literal source text rather than errors. Other malformed or unterminated candidates resume ordinary scanning, so a later valid inline sequence may still be recognized.

Inline attribute validation has three outcomes: `true, nil` accepts the directive; `false, nil` is semantic rejection and uses literal fallback; a non-nil error is an internal parse failure and does not fall back or continue scanning.
The common parser must confirm that a candidate is structurally closed before invoking its definition's attribute validator.

When an ordered Block Definition succeeds while reading, `parseOneBlock` immediately compares its normalized declared Block Type with the normalized type from the parsed IR. A mismatch is an ordinary internal error and must not fall through to Paragraph. During list-item AST construction, diagnostic errors preserve their original identity, message, and range; ordinary errors retain their cause and include both Paragraph and List build context.

## AST and Ranges

AST block and inline interfaces use pointer implementations. Parser-produced non-empty source ranges use one-based, inclusive positions; columns count Unicode code points rather than UTF-8 bytes. Every inline AST node has a range: recognized directive nodes cover their complete delimiter syntax, while nested content and literal text nodes cover their own source spans.

## Validation

```sh
go test ./internal/parser
```

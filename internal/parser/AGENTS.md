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

Only Heading and List readers implement the internal `blockSugarReader` contract. A Sugar Reader declares its builder key with `builderKey() string`; `registerBlockSugar` derives the normalized registration key from that declaration and accepts no independent type string. The Block Directive Reader and Paragraph Reader are ordinary `blockReader` implementations, not Sugar Readers. `paragraph` is case-insensitive and reserved for Paragraph fallback registration. Block Directive, Sugar, and generic builder registration must reject it before mutating the builder map; fallback registration must remain atomic.

A successful block reader leaves `blockContext.pos` on the last consumed line. `parseBlocks` advances the cursor once after success. A reader returning `ok=false` must leave the cursor unchanged.

Malformed Heading, List, and Block Directive candidates fall through to the Paragraph reader. A syntactically valid directive still requires a registered builder.

## Lists

Lists are read recursively by `listReader`; do not call the document block reader chain for list-item content. During AST construction, list-item blocks use the common `Parser.buildBlock` dispatch.

Marker-run length is a raw level. `normalizeListLevel` converts changes into logical levels. Any raw increase adds exactly one logical level, regardless of the increase size. Nesting depth has no explicit limit.

## Inline Parsing

Heading, Paragraph, and list-item Paragraph builders parse inline content through the active `Parser` and `Spec`. Code blocks preserve content literally.

Registered inline Directive Types are matched case-insensitively. Definitions own validation, nested-versus-literal content policy, and AST construction; common scanning owns envelopes, recursion, fallback, and ranges.

Empty inline content is valid, and links require a nonempty URI attribute. Unsupported directives, invalid headers, and links without a URI are emitted as literal source text rather than errors. Other malformed or unterminated candidates resume ordinary scanning, so a later valid inline sequence may still be recognized.

Inline attribute validation has three outcomes: `true, nil` accepts the directive; `false, nil` is semantic rejection and uses literal fallback; a non-nil error is an internal parse failure and does not fall back or continue scanning.
The common parser must confirm that a candidate is structurally closed before invoking its definition's attribute validator.

When a Sugar Reader succeeds, `parseOneBlock` immediately compares its normalized declared builder key with the normalized key from the parsed IR. A mismatch is an ordinary internal error and must not fall through to Paragraph. During list-item AST construction, diagnostic errors preserve their original identity, message, and range; ordinary errors retain their cause and include both Paragraph and List build context.

## AST and Ranges

AST block and inline interfaces use pointer implementations. Parser-produced non-empty source ranges use one-based, inclusive positions; columns count Unicode code points rather than UTF-8 bytes. Every inline AST node has a range: recognized directive nodes cover their complete delimiter syntax, while nested content and literal text nodes cover their own source spans.

## Validation

```sh
go test ./internal/parser
```

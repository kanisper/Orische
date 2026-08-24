# Parser Spec Registration Implementation Record

## Status

The internal parser registration refactoring is implemented. This document records the resulting design and the constraints preserved during the change; it is not a proposal or a public API specification.

The parser implementation and tests remain the source of truth for current behavior.

## Outcome

`Spec` is the internal source of block feature registrations and inline directive definitions. `Parser` owns one active `Spec` and carries it through document parsing, AST construction, list-item building, and recursive inline parsing.

The refactoring changed one accepted-syntax rule: registered inline Directive Types are now matched case-insensitively. Other AST, diagnostic, fallback, cursor, and source-range contracts were preserved.

No stable extension API, plugin mechanism, or dynamic registration surface was added.

## Responsibility Model

### `Parser`

`Parser` orchestrates source-to-AST processing and owns the active `Spec`. Its common internal operations include:

- reading a document through the effective block reader chain;
- building any parsed block through `Parser.buildBlock`;
- parsing inline-capable text through `Parser.parseInlines`;
- retaining the same `Spec` during nested list construction and recursive inline parsing.

No separate build-context type exists. Mutable cursor and scan state remain in short-lived `blockContext`, `inlineContext`, and `inlineParseState` values.

### Readers and builders

A `blockReader` recognizes document source and produces private parsed-block IR. A `blockBuilder` converts that IR into an AST block. Block readers never construct final AST nodes. A `blockDefinition` declares its normalized identity through `blockType() string` and builds the resulting IR. A definition that also implements `blockReader` is treated as sugar syntax.

The effective core reader order is fixed:

1. Block Directive
2. Heading
3. List
4. Paragraph fallback

List source uses dedicated recursive reading and never invokes the document reader chain for item content. During AST construction, however, top-level blocks, list-item Paragraph blocks, and nested List blocks all use `Parser.buildBlock`.

### `Spec`

`registerBlock(definition)` handles both directive and sugar definitions. Definitions without a reader use the permanent common Block Directive envelope reader. Reader-capable definitions are also appended to the ordered sugar chain. Paragraph remains separate through `registerBlockFallback` because it must be the final, mandatory reader.

The core specification registers Heading and List as reader-capable definitions, Code as a directive definition, and Paragraph as the fallback. A syntactically valid unregistered Block Directive is still read successfully and then produces the established unsupported-block diagnostic during AST building.

`paragraph` is a case-insensitive reserved Block Type for the Paragraph fallback. General Block registration rejects it; only `registerBlockFallback` may install that definition. Invalid or incomplete registration is rejected before mutation, including the block-definition map, Reader list, and fallback field. A failed reserved-type registration therefore leaves the `Spec` ready for a later valid Paragraph fallback registration. A document `Spec` is validated before source reading for the required Paragraph fallback; the common Block Directive reader is permanent parser infrastructure.

## Syntax Type Normalization

Block and inline registration and lookup share `normalizeSyntaxType`, which uses Go's Unicode-aware `strings.ToLower` behavior.

Normalization applies only to the Directive Type. Attributes, URI values, language names, and literal or nested content retain their original spelling. Duplicate normalized keys, including case-only collisions, return an error and do not replace the first registration.

Block and inline definitions have separate registries. The `code` type is therefore valid as both a Block Directive definition and an inline directive definition.

## Inline Definitions

`Parser.parseInlines` and `inlineParseState` own the common `:[...]{...}` envelope, header splitting, byte-offset scanning, brace handling, recursive sequence control, text flushing, fallback spans, and source-range calculation.

The base `inlineDefinition` declares identity through `inlineType() string` and is accepted by `registerInline(definition)`. Current syntax definitions additionally implement `inlineDirectiveDefinition`, which owns type-specific behavior:

- attribute validation;
- a closed nested-content or literal-content policy;
- AST construction from validated content and the complete directive range.

The common parser confirms the closing delimiter required by the definition's content policy before calling `validateAttribute`. Unterminated candidates therefore do not invoke definition-specific validation. For structurally closed candidates, validation has a three-valued error contract: `(true, nil)` accepts the directive and proceeds to AST construction; `(false, nil)` is semantic rejection and retains the candidate as literal fallback; a non-nil error is an internal failure that propagates as an error, without literal fallback or continued scanning.

Definitions that do not implement a current parser contract are rejected before registry mutation. Inline Sugar can therefore add another definition category later without changing the registration method signature.

The core definitions are:

- `em`: nested content, producing `*ast.Emphasis`;
- `link`: nested content with a required nonempty URI, producing `*ast.Link`;
- `code`: literal content ending at the first `}`, producing `*ast.CodeSpan`.

Semantic rejection returns literal fallback without an error. Definition validation and construction errors are internal failures and propagate as errors. An accepted definition returning a nil AST node is also an internal error.

## Preserved Contracts

- Block readers produce private parsed-block IR.
- Inline parsing occurs during AST building for inline-capable blocks.
- Code Block content remains literal and is never inline-parsed.
- AST block and inline interfaces remain pointer-based.
- Parser-produced nonempty ranges are one-based and inclusive.
- Columns count Unicode code points rather than UTF-8 bytes.
- Every inline AST node has a range.
- Inline directive-node ranges cover the complete directive syntax.
- Nested nodes and literal `Text` nodes carry their own source spans.
- A successful block reader leaves the cursor on the last consumed line.
- A block reader returning `ok=false` does not consume input.
- A successful reader-capable definition's normalized `blockType()` must match the normalized Block Type returned by its parsed IR. `parseOneBlock` checks this immediately after the read; a mismatch is an ordinary internal error and does not fall through to Paragraph.
- Malformed block syntax falls through to Paragraph reading.
- Unsupported, empty-type, semantically rejected, malformed, and unterminated inline candidates retain their established literal-scanning behavior.
- A diagnostic returned while building a list-item Paragraph is propagated unchanged, preserving its identity, message, and range. A normal error retains its cause and is wrapped with both `build "paragraph" block` and `build "list" block` context.

## Verification Coverage

Focused tests cover:

- common top-level and list-item block dispatch;
- active-`Spec` propagation through builders and recursive inline parsing;
- deterministic core reader order and final Paragraph fallback;
- atomic sugar registration and incomplete-spec rejection;
- duplicate normalized block and inline keys;
- one Block Directive reader serving multiple definitions;
- nested and literal inline content policies;
- case-insensitive core inline definitions with value preservation;
- semantic rejection versus internal construction errors;
- inline validation errors versus semantic rejection, including stop-on-error behavior;
- Sugar Reader declared-key/IR-key consistency and Paragraph reserved-key registration atomicity;
- list-item diagnostic identity and nested normal-error context;
- malformed-input recovery and later valid candidates;
- exact one-based, inclusive Unicode source ranges;
- unchanged renderer and CLI behavior through the repository test suite.

Run:

```sh
go test ./internal/parser
go test ./...
```

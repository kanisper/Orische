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

A `blockReader` recognizes document source and produces private parsed-block IR. A `blockBuilder` converts that IR into an AST block. Block readers never construct final AST nodes.

The effective core reader order is fixed:

1. Block Directive
2. Heading
3. List
4. Paragraph fallback

List source uses dedicated recursive reading and never invokes the document reader chain for item content. During AST construction, however, top-level blocks, list-item Paragraph blocks, and nested List blocks all use `Parser.buildBlock`.

### `Spec`

Block registration is responsibility-oriented:

- `registerBlockDirectiveReader` installs the shared standard-syntax reader exactly once;
- `registerBlockDirectiveDefinition` associates a Directive Type with a builder;
- `registerBlockSugar` atomically installs a sugar Reader/Builder pair;
- `registerParagraphFallback` installs the dedicated final fallback and its builder.

The core specification registers Heading and List as sugar features, `code` as a block directive definition, and Paragraph as the fallback. A syntactically valid unregistered Block Directive is still read successfully and then produces the established unsupported-block diagnostic during AST building.

Invalid or incomplete registration is rejected before mutation. A document `Spec` is validated before source reading, including the required shared Block Directive reader and Paragraph fallback.

## Directive Type Normalization

Block and inline registration and lookup share `normalizeDirectiveType`, which uses Go's Unicode-aware `strings.ToLower` behavior.

Normalization applies only to the Directive Type. Attributes, URI values, language names, and literal or nested content retain their original spelling. Duplicate normalized keys, including case-only collisions, return an error and do not replace the first registration.

Block and inline definitions have separate registries. The `code` type is therefore valid as both a Block Directive definition and an inline directive definition.

## Inline Definitions

`Parser.parseInlines` and `inlineParseState` own the common `:[...]{...}` envelope, header splitting, byte-offset scanning, brace handling, recursive sequence control, text flushing, fallback spans, and source-range calculation.

An `inlineDirectiveDefinition` owns type-specific behavior:

- attribute validation;
- a closed nested-content or literal-content policy;
- AST construction from validated content and the complete directive range.

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
- Malformed block syntax falls through to Paragraph reading.
- Unsupported, empty-type, semantically rejected, malformed, and unterminated inline candidates retain their established literal-scanning behavior.

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
- malformed-input recovery and later valid candidates;
- exact one-based, inclusive Unicode source ranges;
- unchanged renderer and CLI behavior through the repository test suite.

Run:

```sh
go test ./internal/parser
go test ./...
```

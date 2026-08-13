# Parser Spec Registration Plan

## Status

This document is a proposed implementation plan. It describes an internal refactoring of `internal/parser`; it does not describe behavior that is already implemented.

The parser implementation and tests remain the source of truth until this plan is implemented. The refactoring should preserve the currently accepted syntax, AST output, source ranges, fallback behavior, and diagnostics.

## Motivation

`Spec` currently registers document block parsers and block AST builders independently. This permits an inconsistent configuration in which parsing produces a private IR node for which no corresponding builder has been registered.

Inline directives are managed differently: their recognition, validation, content handling, and AST construction are selected by hardcoded directive-type dispatch in the inline parser. Adding an inline element therefore requires modifying the central parser instead of adding a definition to `Spec`.

The registration model should express complete language features while preserving an important distinction between syntax and semantics:

- Block Directive is the standard block syntax and can represent multiple directive types.
- Heading and List are sugar syntax with their own document block parsers.
- Paragraph is the document block fallback.
- Inline directives share one syntactic form but differ in validation, content handling, and AST construction.

A block parser and a builder are therefore not universally a one-to-one pair. The design must guarantee consistency without hiding this distinction.

## Goals

- Make `Spec` the internal source of parser feature definitions for both block and inline elements.
- Register each sugar-syntax block parser together with the builder required for the private IR it produces.
- Register standard block directive types independently from the shared Block Directive syntax parser.
- Replace hardcoded inline directive-type dispatch with definitions registered in `Spec`.
- Represent whether inline directive content is recursively parsed or retained literally.
- Ensure the active `Spec` is used throughout AST building, inline parsing, and recursive list construction.
- Reuse one block-building dispatch path at the document level and inside list items.
- Detect invalid or duplicate registrations as early as practical.
- Preserve all current language behavior and parser invariants.
- Keep the initial implementation internal to `internal/parser`.

## Non-goals

- Defining a stable public extension API.
- Providing a plugin or dynamic-loading mechanism.
- Changing accepted block or inline syntax.
- Adding new directive types or AST node types.
- Changing the private parsed-block IR solely to make it public or generic.
- Replacing the dedicated recursive list parser with the document block parser.
- Changing malformed-input recovery, unsupported-directive handling, or diagnostic wording unless required by a separately approved change.
- Finalizing registration method or type names before their responsibilities are settled.

## Design Principles

### Register language features, not unrelated callbacks

Registration operations should represent coherent language features. A sugar syntax that produces a particular builder key should not be registerable without the builder that handles that key.

The shared Block Directive parser is different: it recognizes the standard envelope and places the directive type in private IR. Directive-type definitions provide the builders selected from that IR. The syntax parser and each directive definition must therefore remain distinct concepts.

### Separate syntax recognition from AST construction

Document block parsing must continue to produce private parsed-block IR. Final AST nodes are created only during the build phase.

For inline directives, the common envelope should be parsed centrally, while registered definitions determine directive-specific validation, content policy, and AST construction. The implementation may use a small private inline representation if it clarifies this boundary, but introducing a large generalized IR is not itself a goal.

### Carry one active specification through the pipeline

A `Parser` created with a particular `Spec` must use that same specification for all nested work. No builder or recursive parsing path should silently fall back to global core behavior.

### Preserve deterministic precedence and fallback

Registration must not make parser order accidental. The core block parser order remains part of the language contract, and unsupported or malformed syntax must retain its current behavior.

## Responsibility Model

### `Spec`

`Spec` owns internal definitions and their ordering. It is responsible for:

- the ordered document block parser chain;
- the paragraph fallback;
- builder lookup for parsed block keys;
- standard block directive-type definitions;
- inline directive-type definitions;
- registration validation, including duplicate-key handling;
- exposing only the internal lookup operations needed by parsing and building.

`Spec` should not parse a document, own mutable parse cursors, or build AST nodes directly.

### Document block parsers

Document block parsers recognize syntax and produce private parsed-block IR. They retain the existing success and cursor contracts.

The standard Block Directive parser recognizes the shared directive envelope. It does not own the set of supported directive types. A syntactically valid directive may therefore be parsed before builder support is checked.

Heading and List parsers recognize sugar syntax. Their registration must associate their output with a compatible builder.

Paragraph remains a separately designated fallback and must always run last.

### Block builders

Block builders convert private parsed-block IR into AST blocks. Builders that accept inline-capable content must request inline parsing through the active build context rather than through a package-level core-only path.

### Inline parser and inline definitions

The inline parser owns sequence scanning, common directive-envelope recognition, delimiter handling, recursive sequence control, text flushing, and source-offset tracking.

A registered inline directive definition owns the behavior specific to a directive type:

- attribute requirements and validation;
- whether content is recursively parsed or treated literally;
- construction of the appropriate AST node from validated data;
- directive-specific semantic rejection that currently causes literal fallback.

The common parser remains responsible for preserving the exact source span when a candidate must be emitted as literal text.

### Build context

An internal build context carries the active `Spec` and provides common dispatch operations for:

- building a parsed block through its registered builder;
- parsing inline-capable text through the active inline definitions;
- recursively building nested blocks without bypassing `Spec`.

This context is an internal coordination mechanism, not a public extension surface.

## Block Registration

### Standard Block Directive syntax

The Block Directive parser is registered as a syntax parser exactly once in the core parser chain. It remains first in precedence.

Block directive types such as `code` are registered as semantic definitions keyed by normalized directive type. Their registration supplies the builder needed to convert the generic parsed directive block into the appropriate AST block.

This separation allows one standard syntax parser to serve multiple directive types without pretending that each directive owns a separate block parser.

A syntactically valid Block Directive whose type has no registered builder must continue to produce the current unsupported-block build error.

### Sugar syntax

Heading and List are registered as sugar-syntax features. Each registration includes both:

- the document block parser that recognizes the sugar form; and
- the block builder for the private IR produced by that parser.

Registration order remains explicit because it controls parsing precedence. Adding a sugar syntax must not reorder existing parsers implicitly.

The registration layer should reject or otherwise prevent a sugar parser from being installed without its corresponding builder definition.

### Paragraph fallback

Paragraph is registered through a dedicated fallback operation that associates its parser with its builder. It is not part of the ordinary ordered parser list and is appended only when the effective parser chain is obtained.

The design must guarantee that the fallback remains last. Replacing or omitting it should either be impossible for a usable core `Spec` or fail during specification validation rather than during document parsing.

### Builder keys and normalization

Builder lookup continues to use stable internal keys produced by private parsed-block IR. Key normalization should occur at one defined boundary. Registration and lookup must use the same normalization rule, and duplicate normalized keys must not silently replace existing definitions.

The implementation should avoid requiring a syntax parser to know about concrete AST types. Its contract is the private IR and the builder key associated with that IR.

## Inline Registration

### Common syntax parsing

The `:[...]{...}` envelope, header splitting, byte-offset scanning, brace termination, and source-range calculation remain common parser responsibilities.

The parser extracts the directive type and attribute, then looks up the type in the active `Spec`. The central inline parser should no longer select `em`, `link`, or `code` through directive-specific hardcoded dispatch.

### Definition lookup

Inline definitions are keyed by normalized directive type. The core specification registers the existing emphasis, link, and code-span definitions.

An unregistered inline directive must continue to be emitted as literal source text according to the current fallback rules. Registration lookup failure is not an error.

Duplicate normalized directive types should be rejected or reported during specification construction rather than resolved by silent replacement.

### Content policy

Each definition declares one of the content policies required by current behavior:

- nested content, which is recursively parsed as an inline sequence; or
- literal content, which is preserved without recognizing nested inline directives.

Emphasis and link use nested content. Code span uses literal content and terminates at the first closing brace, as it does currently.

Content policy belongs to the definition because it is semantic behavior of the directive type, but delimiter scanning and source-range accounting remain in the common parser.

### Validation and construction

Directive-specific validation belongs with the directive definition or its builder. Link, for example, requires a nonempty URI, while emphasis and code span do not require an attribute.

A definition must distinguish semantic rejection from an internal error. Semantic rejection follows existing literal fallback behavior. Internal inconsistencies and construction failures may return errors.

AST construction must preserve pointer-based inline node contracts and assign a range to every node. Directive-node ranges include the complete directive syntax, while nested content and literal `Text` nodes retain their own source spans.

## Specification and Build-Context Propagation

The active `Spec` must be passed from `Parser` into document AST building. Block builders receive access to the internal build context rather than calling a core-only inline parser directly.

Heading and Paragraph builders use the context to parse inline content. Code Block builders preserve their content literally and must not request inline parsing.

The inline parser receives the active `Spec`, including during recursive parsing of nested inline content. A custom internal specification must therefore behave consistently in headings, paragraphs, and list-item paragraphs.

Package-level convenience parsing may continue to construct the core specification, but internal nested calls must never construct a replacement core specification.

## List Dispatch

List syntax continues to use its dedicated recursive parser. This plan does not make list items invoke the document block parser.

During AST construction, however, list-item blocks should use the same build-context dispatch as top-level document blocks. The List builder should not directly construct Paragraph AST nodes or maintain a type switch that duplicates builder selection.

Nested lists are built by dispatching their private IR through the registered List builder. Paragraph-like list-item content is built through the registered Paragraph builder. This ensures that inline registration and block-building behavior are applied consistently at every nesting level.

Any list-specific metadata needed to build item content must remain available in private IR or be passed explicitly without introducing a second independent builder registry.

## Naming Guidelines

Concrete method and type names remain undecided. Naming should be selected after the responsibility model is represented cleanly in code.

Names must make the following distinctions clear:

- a document syntax parser versus a directive-type definition;
- standard Block Directive syntax versus sugar syntax;
- an ordinary ordered parser versus the paragraph fallback;
- block AST construction versus inline directive AST construction;
- registration operations versus lookup or dispatch operations;
- content parsing policy versus AST builder implementation.

Names implying a universal one-to-one parser/builder relationship should be avoided because standard directives share one syntax parser. Names suggesting a stable public plugin API should also be avoided while the mechanism remains internal.

## Preserved Invariants

The implementation must preserve all existing parser invariants:

- Document block parser order is Block Directive, Heading, List, then Paragraph fallback.
- Block Directive is the standard block syntax; Heading and List remain sugar syntax.
- Document block parsing produces private parsed-block IR, not final AST nodes.
- Inline parsing occurs during AST building for inline-capable blocks.
- Code Block content is not inline-parsed.
- Lists use dedicated recursive parsing rather than the document block parser.
- AST block and inline interfaces continue to be implemented by pointer types.
- Parser-produced nonempty source ranges are one-based and inclusive.
- Columns count Unicode code points rather than UTF-8 bytes.
- Every inline AST node carries a source range.
- Inline directive-node ranges include the complete directive syntax.
- Nested inline nodes and literal text nodes carry their own source spans.
- On success, a document block parser leaves the block context on the last consumed line; the caller advances once.
- A document block parser returning `ok=false` does not consume input.
- A parser that scans ahead restores its starting cursor before returning `ok=false`.
- Paragraph fallback remains capable of consuming any nonblank block start not accepted by an earlier parser.

## Error and Fallback Compatibility

This refactoring must preserve the existing user-visible distinctions:

- An invalid or unterminated Block Directive candidate falls through to paragraph text.
- A syntactically valid Block Directive with an unregistered type reaches AST building and returns an unsupported-block diagnostic.
- An unsupported inline directive is retained as literal source text.
- An empty inline directive type is retained as literal source text.
- A link without a nonempty URI is retained as literal source text.
- Existing malformed and unterminated inline scanning behavior is preserved, including recognition of later valid candidates where currently allowed.
- Literal inline content does not recursively parse nested directive syntax.
- Builder/node type mismatches remain internal errors rather than fallback cases.

Specification-construction errors, such as duplicate normalized keys or a missing required fallback, should be detected before parsing where practical. Their exact internal API representation remains an implementation decision.

## Implementation Stages

### Stage 1: Introduce common build dispatch

- Add an internal build context carrying the active `Spec`.
- Move block builder lookup and error wrapping into one shared dispatch path.
- Route document AST construction through that path.
- Preserve existing builder behavior and output.

### Stage 2: Propagate the active specification

- Update block builder contracts to receive the build context.
- Route Heading and Paragraph inline parsing through the context.
- Ensure recursive inline parsing uses the same `Spec`.
- Keep Code Block construction literal.

### Stage 3: Unify list AST construction

- Replace direct Paragraph construction and list-specific builder selection with common build-context dispatch.
- Preserve the dedicated list parsing algorithm and existing private IR.
- Verify nested lists and list-item inline ranges before continuing.

### Stage 4: Restructure block registration

- Introduce responsibility-oriented registration operations for standard directive definitions, sugar-syntax features, and the fallback.
- Migrate the core specification without changing parser order.
- Add duplicate and incomplete-registration validation.
- Remove the independent parser/builder setup paths once all core features use the new model.

### Stage 5: Add inline definitions to `Spec`

- Add internal inline directive definitions and content policies.
- Register the existing emphasis, link, and code-span behavior in the core specification.
- Replace hardcoded directive-type dispatch with specification lookup.
- Preserve exact literal fallback and range behavior.

### Stage 6: Consolidate and document the implemented architecture

- Remove obsolete registration and dispatch helpers.
- Update `docs/30-parser-architecture.md` and `docs/40-go-layout.md` to describe the implementation that actually lands.
- Update `docs/20-syntax.md` only if accepted syntax or user-visible fallback behavior changes. No syntax-document change is expected for this internal refactoring.
- Re-evaluate names after responsibilities and call sites are visible together.

Each stage should be independently testable and should avoid mixing registration refactoring with new language features.

## Test Plan

### Existing behavior

Run the complete parser test suite after every stage:

- `go test ./internal/parser`

Existing tests must continue to cover block precedence, cursor behavior, parsed IR, AST shape, nested lists, inline nesting, literal code content, malformed-input fallback, and Unicode source ranges.

### Registration consistency

Add focused tests that verify:

- a sugar-syntax registration supplies both recognition and construction behavior;
- incomplete feature registration is rejected or impossible through the internal API;
- duplicate normalized block keys are not silently overwritten;
- duplicate normalized inline directive types are not silently overwritten;
- Paragraph remains the final fallback;
- the core parser order remains Block Directive, Heading, List, Paragraph.

### Active specification propagation

Add tests using a non-core internal `Spec` to verify that registered behavior is visible in:

- top-level paragraphs;
- headings;
- list-item paragraphs;
- nested list items;
- recursively nested inline directives.

These tests should demonstrate that no nested path constructs or consults an unrelated core specification.

### Inline definitions

Add focused tests that verify:

- registered nested-content definitions recursively parse child inline directives;
- registered literal-content definitions do not parse child directives;
- unregistered directives remain literal;
- semantic validation rejection follows literal fallback;
- malformed and unterminated candidates retain current scanning behavior;
- empty content remains valid where it is currently valid;
- all produced inline nodes retain correct Unicode-aware ranges.

### Block directives and sugar syntax

Add tests that verify:

- one Block Directive syntax parser dispatches multiple registered directive types without requiring multiple syntax parsers;
- an unregistered but syntactically valid block directive retains the current build error;
- sugar parser precedence remains deterministic;
- failed sugar recognition does not consume the block context;
- Code Block content remains literal after inline definitions become configurable.

Broader repository tests may be run after parser tests pass to confirm that AST output consumed by the renderer and CLI has not changed.

## Completion Criteria

The refactoring is complete when:

- core block and inline features are defined through the new internal `Spec` model;
- no directive-type switch in the central inline parser selects `em`, `link`, or `code` behavior;
- sugar-syntax parsers cannot be configured independently from their required builders through the intended internal registration path;
- standard block directive types share the common Block Directive syntax parser;
- the active `Spec` reaches all block builders and recursive inline parsing;
- top-level and list-item AST construction use common block dispatch;
- all existing parser behavior and source-range tests pass;
- new registration and custom-spec tests pass;
- architecture and layout documentation describe the resulting implementation accurately;
- no public extension or plugin API is implied by the internal design.

## Open Questions

- What terminology best distinguishes a standard directive-type definition from a sugar-syntax feature registration?
- Should registration validation happen incrementally, when constructing a usable `Spec`, or both?
- Should registration failures be returned as errors or made impossible through unexported construction helpers?
- Which component should own directive-type normalization, and should type matching remain case-insensitive everywhere?
- How much private parsed-inline representation is useful before it becomes unnecessary abstraction?
- Should content policy be a closed internal enum or behavior exposed through a private strategy interface?
- How should a directive definition report semantic rejection separately from an internal build error while preserving literal fallback?
- Does list-item private IR need additional context before all blocks can use common build dispatch?
- Should custom internal specifications be allowed to replace core definitions, or should duplicate registration always be rejected?

These questions should be resolved from concrete implementation constraints and tests. They should not delay establishing the responsibility boundaries described above.

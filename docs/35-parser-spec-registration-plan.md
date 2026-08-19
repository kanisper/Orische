# Parser Spec Registration Plan

## Status

This document is a proposed implementation plan. It describes an internal refactoring of `internal/parser`; it does not describe behavior that is already implemented.

The parser implementation and tests remain the source of truth until this plan is implemented. Except for the explicitly approved case-insensitive matching of inline Directive Types described below, the refactoring should preserve the currently accepted syntax, AST output, source ranges, fallback behavior, and diagnostics.

## Motivation

`Spec` currently registers document block readers and block AST builders independently. This permits an inconsistent configuration in which reading produces a private IR node for which no corresponding builder has been registered.

Inline directives are managed differently: their recognition, validation, content handling, and AST construction are selected by hardcoded directive-type dispatch in the inline parser. Adding an inline element therefore requires modifying the central parser instead of adding a definition to `Spec`.

The registration model should express complete language features while preserving an important distinction between syntax and semantics:

- Block Directive is the standard block syntax and can represent multiple directive types.
- Heading and List are sugar syntax with their own document block readers.
- Paragraph is the document block fallback reader.
- Inline directives share one syntactic form but differ in validation, content handling, and AST construction.

A block reader and a builder are therefore not universally a one-to-one pair. The design must guarantee consistency without hiding this distinction.

## Goals

- Make `Spec` the internal source of parser feature definitions for both block and inline elements.
- Register each sugar-syntax block reader together with the builder required for the private IR it produces.
- Register standard block directive types independently from the shared Block Directive reader.
- Replace hardcoded inline directive-type dispatch with definitions registered in `Spec`.
- Represent whether inline directive content is recursively parsed or retained literally.
- Ensure the active `Spec` is used throughout AST building, inline parsing, and recursive list construction.
- Reuse one block-building dispatch path at the document level and inside list items.
- Detect invalid or duplicate registrations as early as practical.
- Apply one case-insensitive Directive Type matching rule to block and inline directives.
- Preserve all other current language behavior and parser invariants.
- Keep the initial implementation internal to `internal/parser`.

## Non-goals

- Defining a stable public extension API.
- Providing a plugin or dynamic-loading mechanism.
- Changing accepted block or inline syntax beyond making inline Directive Type matching case-insensitive.
- Adding new directive types or AST node types.
- Changing the private parsed-block IR solely to make it public or generic.
- Replacing the dedicated recursive list reader with the document block reader chain.
- Changing malformed-input recovery, unsupported-directive handling, or diagnostic wording unless required by a separately approved change.
- Finalizing registration method or type names before their responsibilities are settled.

## Design Principles

### Register language features, not unrelated callbacks

Registration operations should represent coherent language features. A sugar syntax that produces a particular builder key should not be registerable without the builder that handles that key.

The shared Block Directive reader is different: it reads the standard envelope and places the directive type in private IR. Directive-type definitions provide the builders selected from that IR. The reader and each directive definition must therefore remain distinct concepts.

### Separate syntax recognition from AST construction

Document block parsing must continue to produce private parsed-block IR. Final AST nodes are created only during the build phase.

For inline directives, the common envelope should be parsed centrally, while registered definitions determine directive-specific validation, content policy, and AST construction. The implementation may use a small private inline representation if it clarifies this boundary, but introducing a large generalized IR is not itself a goal.

### Carry one active specification through the pipeline

A `Parser` created with a particular `Spec` must use that same specification for all nested work. No builder or recursive parsing path should silently fall back to global core behavior.

### Preserve deterministic precedence and fallback

Registration must not make reader order accidental. The core block reader order remains part of the language contract, and unsupported or malformed syntax must retain its current behavior.

## Responsibility Model

### `Spec`

`Spec` owns internal definitions and their ordering. It is responsible for:

- the ordered document block reader chain;
- the paragraph fallback reader;
- builder lookup for parsed block keys;
- standard block directive-type definitions;
- inline directive-type definitions;
- registration validation, including duplicate-key handling;
- exposing only the internal lookup operations needed by parsing and building.

`Spec` should not parse a document, own mutable parse cursors, or build AST nodes directly.

### Document block readers

Document block readers consume source syntax and produce private parsed-block IR. A reader both determines whether its syntax applies and extracts the structured data, content, and source range required by that IR. Readers retain the existing success and cursor contracts.

The standard Block Directive reader reads the shared directive envelope. It does not own the set of supported directive types. A syntactically valid directive may therefore be read before builder support is checked.

Heading and List readers handle sugar syntax. Their registration must associate their output with a compatible builder.

Paragraph remains a separately designated fallback reader and must always run last.

### Block builders

Block builders convert private parsed-block IR into AST blocks. Builders that accept inline-capable content must request inline parsing through the active `Parser` rather than through a package-level core-only path.

### Inline parsing and inline definitions

`Parser.parseInlines` and its short-lived `inlineParseState` own sequence scanning, common directive-envelope recognition, delimiter handling, recursive sequence control, text flushing, and source-offset tracking.

A registered inline directive definition owns the behavior specific to a directive type:

- attribute requirements and validation;
- whether content is recursively parsed or treated literally;
- construction of the appropriate AST node from validated data;
- directive-specific semantic rejection that currently causes literal fallback.

The common inline-sequence processing remains responsible for preserving the exact source span when a candidate must be emitted as literal text.

### `Parser` orchestration

No separate build-context type is introduced in this refactoring. `Parser` already owns the active `Spec` and therefore coordinates both parsing and AST construction.

Internal `Parser` methods provide common operations for:

- building a parsed block through its registered builder;
- parsing inline-capable text through the active inline definitions;
- recursively building nested blocks without bypassing `Spec`.

Block builders receive the active `Parser` when they need these operations. Document and inline cursor state remains in short-lived `blockContext`, `inlineContext`, and `inlineParseState` values rather than becoming mutable fields of `Parser`.

## Block Registration

### Standard Block Directive syntax

The Block Directive reader is registered exactly once in the core reader chain. It remains first in precedence.

Block directive types such as `code` are registered as semantic definitions keyed by normalized directive type. Their registration supplies the builder needed to convert the generic parsed directive block into the appropriate AST block.

This separation allows one standard reader to serve multiple directive types without pretending that each directive owns a separate block reader.

A syntactically valid Block Directive whose type has no registered builder must continue to produce the current unsupported-block build error.

### Sugar syntax

Heading and List are registered as sugar-syntax features. Each registration includes both:

- the document block reader that reads the sugar form; and
- the block builder for the private IR produced by that reader.

Registration order remains explicit because it controls parsing precedence. Adding a sugar syntax must not reorder existing readers implicitly.

The registration layer should reject or otherwise prevent a sugar reader from being installed without its corresponding builder definition.

### Paragraph fallback

Paragraph is registered through a dedicated fallback operation that associates its reader with its builder. It is not part of the ordinary ordered reader list and is appended only when the effective reader chain is obtained.

The design must guarantee that the fallback remains last. Replacing or omitting it should either be impossible for a usable core `Spec` or fail during specification validation rather than during document parsing.

### Builder keys

Builder lookup continues to use stable internal keys produced by private parsed-block IR. Registration and lookup must use the Directive Type normalization rule defined below, and duplicate normalized keys must not silently replace existing definitions.

The implementation should avoid requiring a block reader to know about concrete AST types. Its contract is the private IR and the builder key associated with that IR.

## Directive Type Normalization

Directive Type matching is case-insensitive for both block and inline directives. This is already the effective behavior for block builder lookup and becomes the required behavior for inline definition lookup when this plan is implemented.

Directive Types are converted to a canonical lowercase form using the same Unicode-aware lowercasing behavior as Go's `strings.ToLower`. Both registration keys and parsed lookup keys must be normalized at their respective boundaries before comparison or map access. Callers and individual builders must not perform independent case handling.

Normalization applies only to the Directive Type. Directive attributes and content retain their original spelling and remain subject to their directive-specific semantics. In particular, URI, language, and literal-content values must not be lowercased as a side effect of type normalization.

Directive Type spelling is not preserved as semantic AST data unless a future AST contract explicitly requires it. Differently cased spellings of the same type therefore select the same definition and produce equivalent AST node kinds.

Registrations whose keys differ only by case normalize to the same key and are duplicates. They must be rejected or reported during specification construction rather than silently replacing one another. The same collision rule applies consistently to block and inline directive definitions.

For inline syntax this is an intentional behavior change: forms such as uppercase or mixed-case spellings of registered Directive Types will become recognized instead of remaining literal text. The implementation must update parser tests and `docs/20-syntax.md` when this behavior lands. Until then, the existing implementation and tests continue to define current behavior.

## Inline Registration

### Common syntax parsing

The `:[...]{...}` envelope, header splitting, byte-offset scanning, brace termination, and source-range calculation remain common parser responsibilities.

The parser extracts the directive type and attribute, then looks up the type in the active `Spec`. The central inline parser should no longer select `em`, `link`, or `code` through directive-specific hardcoded dispatch.

### Definition lookup

Inline definitions are keyed by Directive Type after applying the shared case-insensitive normalization rule. The core specification registers the existing emphasis, link, and code-span definitions.

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

## Specification Propagation Through `Parser`

The active `Spec` remains owned by `Parser` throughout document AST building. Block builders receive the active `Parser` rather than a separate build context or a core-only inline parser.

Inline parsing is exposed as an internal `Parser` method. Heading and Paragraph builders use that method to parse inline content. Code Block builders preserve their content literally and must not request inline parsing.

A short-lived `inlineParseState` supports scanning and recursion for one inline input while retaining access to the active `Parser`. Recursive parsing therefore uses the same `Spec`. A custom internal specification must behave consistently in headings, paragraphs, and list-item paragraphs.

Package-level convenience parsing may continue to construct the core specification, but internal nested calls must never construct a replacement core specification.

## List Dispatch

List syntax continues to use its dedicated recursive reader. This plan does not make list items invoke the document block reader chain.

During AST construction, however, list-item blocks should use the same `Parser` dispatch as top-level document blocks. The List builder should not directly construct Paragraph AST nodes or maintain a type switch that duplicates builder selection.

Nested lists are built by dispatching their private IR through the registered List builder. Paragraph-like list-item content is built through the registered Paragraph builder. This ensures that inline registration and block-building behavior are applied consistently at every nesting level.

Any list-specific metadata needed to build item content must remain available in private IR or be passed explicitly without introducing a second independent builder registry.

## Naming Guidelines

The internal terminology is fixed as follows:

- `Parser` is the source-to-AST orchestrator and owns the active `Spec`.
- A document `blockReader` reads source and produces private parsed-block IR.
- A `blockBuilder` converts private parsed-block IR into an AST block.
- A directive definition describes Directive Type-specific semantics and AST construction.
- `blockContext` and `inlineContext` hold short-lived source and cursor information.
- `inlineParseState` holds the short-lived scanning and recursion state used by `Parser.parseInlines`.

Reader implementation names follow the syntax they read, including Block Directive, Heading, List, and Paragraph readers. Paragraph is identified by its fallback registration rather than by a different component category.

`Reader` is deliberately limited to the block source-to-private-IR phase. Inline processing does not currently introduce an equivalent private IR and therefore is not named as an inline reader.

Names must continue to distinguish standard Block Directive syntax from sugar syntax, registration from lookup or dispatch, and content policy from AST construction. Names implying a universal one-to-one reader/builder relationship must be avoided because standard directives share one reader. Names suggesting a stable public plugin API must also be avoided while the mechanism remains internal.

## Preserved Invariants

The implementation must preserve all existing parser invariants:

- Document block reader order is Block Directive, Heading, List, then Paragraph fallback.
- Block Directive is the standard block syntax; Heading and List remain sugar syntax.
- Document block readers produce private parsed-block IR, not final AST nodes.
- Inline parsing occurs during AST building for inline-capable blocks.
- Code Block content is not inline-parsed.
- Lists use dedicated recursive reading rather than the document block reader chain.
- AST block and inline interfaces continue to be implemented by pointer types.
- Parser-produced nonempty source ranges are one-based and inclusive.
- Columns count Unicode code points rather than UTF-8 bytes.
- Every inline AST node carries a source range.
- Inline directive-node ranges include the complete directive syntax.
- Nested inline nodes and literal text nodes carry their own source spans.
- On success, a document block reader leaves the block context on the last consumed line; the caller advances once.
- A document block reader returning `ok=false` does not consume input.
- A reader that scans ahead restores its starting cursor before returning `ok=false`.
- Paragraph fallback remains capable of consuming any nonblank block start not accepted by an earlier reader.

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

- Make `Parser` the common AST-construction coordinator without introducing a build-context type.
- Move block builder lookup and error wrapping into one shared `Parser` dispatch path.
- Route document AST construction through that path.
- Preserve existing builder behavior and output.

### Stage 2: Propagate the active specification

- Update block builder contracts to receive the active `Parser`.
- Define inline parsing as an internal `Parser` method and route Heading and Paragraph through it.
- Ensure recursive inline parsing uses the same `Spec`.
- Keep Code Block construction literal.

### Stage 3: Unify list AST construction

- Replace direct Paragraph construction and list-specific builder selection with common `Parser` dispatch.
- Preserve the dedicated list parsing algorithm and existing private IR.
- Verify nested lists and list-item inline ranges before continuing.

### Stage 4: Restructure block registration

- Introduce responsibility-oriented registration operations for standard directive definitions, sugar-syntax reader features, and the fallback reader.
- Migrate the core specification without changing reader order.
- Apply the shared case-insensitive Directive Type normalization rule to registration and block builder lookup.
- Add duplicate and incomplete-registration validation.
- Remove the independent parser/builder setup paths once all core features use the new model.

### Stage 5: Add inline definitions to `Spec`

- Add internal inline directive definitions and content policies.
- Register the existing emphasis, link, and code-span behavior in the core specification.
- Replace hardcoded directive-type dispatch with specification lookup.
- Apply the shared case-insensitive Directive Type normalization rule to inline lookup.
- Preserve exact literal fallback and range behavior except that registered types with uppercase or mixed-case spelling are now recognized.

### Stage 6: Consolidate and document the implemented architecture

- Remove obsolete registration and dispatch helpers.
- Update `docs/30-parser-architecture.md` and `docs/40-go-layout.md` to describe the implementation that actually lands.
- Update `docs/20-syntax.md` to document case-insensitive Directive Type matching when the inline behavior change is implemented. Other syntax documentation changes are required only if additional accepted syntax or user-visible fallback behavior changes.
- Re-evaluate names after responsibilities and call sites are visible together.

Each stage should be independently testable and should avoid mixing registration refactoring with new language features.

## Test Plan

### Existing behavior

Run the complete parser test suite after every stage:

- `go test ./internal/parser`

Existing tests must continue to cover block-reader precedence, cursor behavior, parsed IR, AST shape, nested lists, inline nesting, literal code content, malformed-input fallback, and Unicode source ranges.

### Registration consistency

Add focused tests that verify:

- a sugar-syntax registration supplies both reading and construction behavior;
- incomplete feature registration is rejected or impossible through the internal API;
- duplicate normalized block keys are not silently overwritten;
- duplicate normalized inline directive types are not silently overwritten;
- block and inline registrations whose types differ only by case are treated as duplicates;
- Paragraph remains the final fallback;
- the core reader order remains Block Directive, Heading, List, Paragraph.

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
- lowercase, uppercase, and mixed-case spellings of a registered Directive Type select the same definition;
- normalization does not modify directive attributes or content;
- genuinely unregistered directives remain literal;
- semantic validation rejection follows literal fallback;
- malformed and unterminated candidates retain current scanning behavior;
- empty content remains valid where it is currently valid;
- all produced inline nodes retain correct Unicode-aware ranges.

### Block directives and sugar syntax

Add tests that verify:

- one Block Directive reader dispatches multiple registered directive types without requiring multiple readers;
- an unregistered but syntactically valid block directive retains the current build error;
- sugar reader precedence remains deterministic;
- a sugar reader returning `ok=false` does not consume the block context;
- Code Block content remains literal after inline definitions become configurable.

Broader repository tests may be run after parser tests pass to confirm that AST output consumed by the renderer and CLI has not changed.

## Completion Criteria

The refactoring is complete when:

- core block and inline features are defined through the new internal `Spec` model;
- no directive-type switch in the central inline parser selects `em`, `link`, or `code` behavior;
- sugar-syntax readers cannot be configured independently from their required builders through the intended internal registration path;
- standard block directive types share the common Block Directive reader;
- block and inline Directive Types use the same case-insensitive normalization rule;
- the active `Spec` reaches all block builders and recursive inline parsing through `Parser`;
- top-level and list-item AST construction use common `Parser` block dispatch;
- no separate build-context type is introduced;
- all existing parser behavior and source-range tests pass;
- new registration and custom-spec tests pass;
- architecture and layout documentation describe the resulting implementation accurately;
- no public extension or plugin API is implied by the internal design.

## Open Questions

- What registration method names best distinguish a standard directive-type definition from a sugar-syntax reader feature?
- Should registration validation happen incrementally, when constructing a usable `Spec`, or both?
- Should registration failures be returned as errors or made impossible through unexported construction helpers?
- Which internal boundary should own the shared Directive Type normalization operation while ensuring that both registration and lookup always apply it?
- How much private parsed-inline representation is useful before it becomes unnecessary abstraction?
- Should content policy be a closed internal enum or behavior exposed through a private strategy interface?
- How should a directive definition report semantic rejection separately from an internal build error while preserving literal fallback?
- Does list-item private IR need additional context before all blocks can use common build dispatch?
- Should custom internal specifications be allowed to replace core definitions, or should duplicate registration always be rejected?

These questions should be resolved from concrete implementation constraints and tests. They should not delay establishing the responsibility boundaries described above.

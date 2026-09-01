# Superseded Parser Registration Record

## Status

This is a historical record of the parser refactoring from #1 through #7. It
describes the intermediate framework-oriented design and is superseded by the
simplified implementation described in [`30-parser-architecture.md`](30-parser-architecture.md).
The names of the old packages and contracts below are intentionally retained
so the implementation history remains understandable; they do not describe
current APIs or package paths.

## Historical Design

The intermediate design introduced a `feature.Language` declaration, a
compiled specification, neutral `feature` contracts, and separate `syntax`
packages. It used those boundaries to validate and assemble built-in Block and
Inline definitions while keeping the parser frontend independent of syntax
implementations.

That design preserved several useful ideas:

- a parser-owned specification selected Block and Inline behavior;
- Block Directive, Heading, List, and Paragraph had explicit precedence;
- lists used dedicated recursive reading;
- inline directives shared one envelope and delegated semantics to definitions;
- nested parsing reused the active parser configuration;
- diagnostics, fallback behavior, and Unicode-aware ranges remained explicit.

## Superseding Decisions

Stages #11 through #13 moved those ideas into one `package parser`:

- `feature.Language` and its construction path became the private `spec` made
  by `newSpec`;
- separate `feature` and `syntax` packages became files grouped by parser
  responsibility;
- generic cross-package IR became concrete private parsed-block nodes;
- capability-based build contexts became direct methods on `Parser`;
- defensive registration validation disappeared with the external-shaped
  registration surface;
- Paragraph and Block Directive remain fixed parser infrastructure, while
  Heading, List, and inline definitions remain built-in parser code.

The current parser still uses a small `spec` because directive lookup, sugar
precedence, and inline definition semantics are useful. It is not a language
compiler, a replaceable implementation registry, or a plugin boundary. See
`30-parser-architecture.md` and `40-go-layout.md` for the current design.

## Behavior Retained

The simplification does not change accepted syntax or user-visible behavior.
The following remain protected by the parser tests:

- Block Directive, Heading, List, then Paragraph precedence;
- malformed block fallback and unsupported valid-directive diagnostics;
- recursive list construction and raw-to-logical level normalization;
- definition-driven nested and literal inline parsing;
- inline fallback and semantic rejection;
- AST construction, diagnostics, and nested error context;
- one-based inclusive ranges with Unicode-code-point columns.

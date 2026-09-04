# Project Roadmap

This file records the long-term development direction of Orische. It is not a versioned language specification and does not replace GitHub Issues or Projects for task tracking.

Detailed implementation work is tracked in GitHub Issues. This document should stay focused on durable project direction and major architectural boundaries.

## Current Foundation

Orische currently has the core pieces required to evolve from a markup-language implementation into an editor-assisted authoring toolchain:

- AST definitions with source ranges;
- parser support for the current block, inline, sugar, escape, and list syntax;
- HTML rendering;
- CLI conversion;
- structured diagnostics;
- stdio Language Server support;
- full-sync document management and LSP position conversion;
- live diagnostics;
- source-first directive completion that remains usable for incomplete input.

The Language Server is intended to remain the single implementation boundary for editor-facing language intelligence. Editor integrations should stay thin and should not duplicate parser or completion semantics.

## Near-Term Roadmap

### Editor MVP

Track: #41

Make Orische practical to use from mainstream editors. The first integrations should focus on launching `orische lsp`, recognizing Orische files, providing basic syntax highlighting, and documenting configuration paths without adding editor-specific language logic.

### Editor Syntax Infrastructure

Track: #42

Introduce editor-oriented grammar infrastructure, including Tree-sitter where useful, while preserving `internal/parser` as the compiler/parser implementation. Editor grammars may be more tolerant of incomplete source and are not intended to replace Orische parsing semantics.

### Structured Completion and Internal Language Spec Metadata

Track: #43

Grow completion beyond directive-name prefixes into structured, context-aware completion. As duplication becomes real, separate reusable declarative language-definition metadata from parser implementation details.

The shared internal language-definition layer should describe what constructs exist in Orische. Parser readers, precedence, parsed-node construction, builders, and fallback behavior remain in `internal/parser`.

This separation is internal to the project. A stable public Go API, external plugin API, or dynamic plugin system is not a goal.

### Emmet-Style Expansion

Track: #44

Design an editor-independent abbreviation language and expansion engine on top of the completion architecture. Abbreviation syntax is tooling syntax rather than Orische source syntax and must not be accepted by the compiler/parser grammar.

Expansion should resolve Orische constructs through shared internal language-definition metadata and produce valid Orische source edits or snippets for LSP clients.

## Architectural Direction

The intended long-term relationship is:

```text
editor
  |
  | LSP
  v
internal/lsp
  |
  +----> internal/parser ----> internal/ast
  |
  +----> internal/completion
                 |
                 +----> internal language-definition metadata
                           |
                           +----> internal/parser
```

Editor syntax grammars such as Tree-sitter or TextMate remain adjacent tooling concerns rather than compiler parser implementations.

The project should prefer small internal interfaces and concrete metadata over plugin-grade registries or speculative extension frameworks.

## Later Work

Potential work after the editor/completion roadmap includes:

- semantic tokens, hover, formatting, and other LSP features where they provide clear value;
- broader editor integration and distribution polish;
- additional built-in block and inline directive types;
- explicit List Block Directive design while preserving current marker sugar semantics;
- structured attributes such as `key=value` if added to the language specification;
- additional output formats;
- performance measurement and optimization when real workloads justify it.

## Project Management

GitHub Projects, Issues, sub-issues, dependencies, and milestones are used for execution tracking. This document should not accumulate detailed checklists that are better represented there.

Architecture decisions that must remain true after an Issue is closed belong in the repository documentation.

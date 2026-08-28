# Agent Guide

## Source of Truth

Use the following precedence when information conflicts:

1. Tests
2. Implementation
3. Documentation

Treat the parser implementation and tests as the current language behavior. Update tests and syntax documentation together when changing accepted syntax.

## Repository Map

- `internal/ast/` — AST definitions and pointer-based node contracts
- `internal/parser/` — parser frontend, neutral feature contracts, built-in syntax implementations, and tests
- `internal/render/html/` — completed AST-to-HTML renderer and its output/security contracts
- `cmd/` — command-line entrypoint for converting Orische source files to HTML, with diagnostics and tests
- `docs/` — current syntax, architecture, layout, and status

## Parser Invariants

- Block reader order is directive, heading, list, then paragraph fallback.
- Block readers produce `feature.BlockNode` IR, not final AST nodes. Shared text-backed IR uses `feature.TextBlock`; syntax-specific IR may remain private to its syntax package.
- Inline parsing occurs during AST building for inline-capable blocks.
- Code block content is not inline-parsed.
- Lists use dedicated recursive reading rather than the document block reader chain.
- AST block and inline interfaces are implemented by pointer types.
- Parser-produced non-empty source ranges are one-based and inclusive; columns count Unicode code points, not UTF-8 bytes.
- Every inline AST node carries a range. Directive-node ranges include the complete `:[...]{...}` syntax; nested content and literal text nodes carry their own source spans.
- Block readers receive immutable `feature.BlockInput` and report `Matched`, `Consumed`, and `Node` in `feature.BlockReadResult`.
- A non-match returns zero consumption and no node. A match returns a non-nil node and a positive consumed-line count within the available input.
- Syntax packages do not import the parser frontend; recursive builders use `feature.BuildContext`.
- The Paragraph reader and definition are fixed parser infrastructure;
  `feature.Language` declares only replaceable Block and Inline definitions.

## Validation

For parser changes, run:

```sh
go test ./internal/parser/...
```

For HTML renderer changes, run:

```sh
go test ./internal/render/html
```

Renderer-specific maintenance rules live in `internal/render/html/AGENTS.md`.

## Change Discipline

- Prefer focused changes that preserve existing package boundaries.
- For behavioral code changes, add or update focused tests before editing production code. Run the focused tests first and confirm that they fail for the expected reason, then implement the change and rerun both the focused tests and the required validation suite.
- Keep `docs/20-syntax.md` aligned with accepted syntax.
- Do not document planned files or packages as if they already exist.
- Delegate only bounded, independent research or review tasks when delegation will reduce total work or protect the main context. Prefer a lower-cost model with sufficient reasoning effort, pass only the context it needs, and keep integration responsibility with the primary agent.
- Unless the user requests otherwise, the agent may choose commit timing. Commit only a coherent, verified unit of work, keep unrelated user changes out of the commit, and report the resulting commit identifier.

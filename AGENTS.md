# Agent Guide

## Source of Truth

Use the following precedence when information conflicts:

1. Tests
2. Implementation
3. Documentation

Treat the parser implementation and tests as the current language behavior. Update tests and syntax documentation together when changing accepted syntax.

## Repository Map

- `internal/ast/` — AST definitions and pointer-based node contracts
- `internal/parser/` — block parsing, parsed-block IR, AST builders, and inline parsing
- `internal/render/html/` — completed AST-to-HTML renderer and its output/security contracts
- `cmd/` — reserved for command-line entrypoints; currently empty
- `docs/` — current syntax, architecture, layout, and status

## Parser Invariants

- Block parser order is directive, heading, list, then paragraph fallback.
- Block parsing produces private parsed-block IR, not final AST nodes.
- Inline parsing occurs during AST building for inline-capable blocks.
- Code block content is not inline-parsed.
- Lists use dedicated recursive parsing rather than the document block parser.
- AST block and inline interfaces are implemented by pointer types.
- Source ranges are one-based and inclusive.
- On success, a block parser leaves `ctx.pos` on the last consumed line. The caller advances it once.
- A block parser that returns `ok=false` must not consume input.

## Validation

For parser changes, run:

```sh
go test ./internal/parser
```

For HTML renderer changes, run:

```sh
go test ./internal/render/html
```

Renderer-specific maintenance rules live in `internal/render/html/AGENTS.md`.

## Change Discipline

- Prefer focused changes that preserve existing package boundaries.
- Add or update tests for syntax and parser behavior changes.
- Keep `docs/20-syntax.md` aligned with accepted syntax.
- Do not document planned files or packages as if they already exist.

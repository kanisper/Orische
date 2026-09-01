# Agent Guide

## Source of Truth

Use the following precedence when information conflicts:

1. Tests
2. Implementation
3. Documentation

Treat the parser implementation and tests as the current language behavior. Update tests and syntax documentation together when changing accepted syntax.

## Repository Map

- `internal/ast/` - AST definitions and pointer-based node contracts
- `internal/parser/` - source-to-AST parser, built-in syntax behavior, and tests
- `internal/render/html/` - AST-to-HTML renderer and its output/security contracts
- `cmd/` - command-line entrypoint for converting Orische source files to HTML, with diagnostics and tests
- `docs/` - current syntax, architecture, layout, and status

## Language Model

Orische prefers explicit structure. Supported Block Directive forms are the explicit named forms for Heading, Paragraph, and Code Block. Heading marker syntax is sugar for Heading semantics, and ordinary nonblank text is the Paragraph fallback. List marker syntax is currently the only List notation; an explicit List Directive has not yet been designed.

The parser does not auto-correct malformed syntax. Malformed block candidates normally fall back to Paragraph. A structurally valid Block Directive may still fail during AST building when a required semantic attribute is invalid, such as a Heading level outside `level1` through `level6`.

Attributes are opaque strings separated from the type by the first colon. A syntax validates an attribute only when that attribute is required for its semantics. Attributes that the syntax does not use are accepted and ignored. Current examples include Paragraph, Emphasis, and Code Span attributes. Heading level and Link URI are semantically meaningful and retain their documented validation behavior.

## Parser Invariants

- All parser implementation files live in one `package parser`; there are no `feature` or `syntax` parser subpackages.
- `Parser` owns a small private `spec` containing Block Directive builders, ordered sugar readers, and inline definitions. It is not a public language or plugin API.
- Block reader order is Block Directive, Heading/List sugar, then Paragraph fallback.
- Block readers produce private concrete parsed-block nodes, not final AST nodes.
- `Parser.buildBlock` dispatches those private nodes to AST construction.
- Heading Directive and Heading sugar share final Heading construction. Paragraph Directive and Paragraph fallback share final Paragraph construction.
- Inline parsing occurs during AST building for inline-capable blocks. Code Block content is not inline-parsed.
- Lists use dedicated recursive reading rather than the document block reader chain.
- Same-level List style is determined by the first item in that logical list. For mixed markers, a line's first marker determines that line's candidate style.
- AST block and inline interfaces are implemented by pointer types.
- Parser-produced non-empty source ranges are one-based and inclusive; columns count Unicode code points, not UTF-8 bytes.
- Every inline AST node carries a range. Inline directive ranges include the complete `:[...]{...}` syntax; nested content and literal text nodes carry their own source spans.
- Unsupported valid Block Directive types produce diagnostics. Invalid required semantic values may produce ordinary AST build errors according to the syntax contract.

## Validation

For parser changes, run:

```sh
go test ./internal/parser/...
```

For repository-wide changes, also run:

```sh
go test ./...
go vet ./...
```

For HTML renderer changes, run:

```sh
go test ./internal/render/html
```

Renderer-specific maintenance rules live in `internal/render/html/AGENTS.md`.
Parser-specific maintenance rules live in `internal/parser/AGENTS.md`.

## Change Discipline

- Prefer focused changes and keep parser syntax implementations local to `package parser`.
- Do not recreate generic package boundaries or extension contracts without a concrete requirement.
- For behavioral code changes, add or update focused tests before editing production code. Run the focused tests first and confirm that they fail for the expected reason, then implement the change and rerun both the focused tests and the required validation suite.
- Keep `docs/20-syntax.md` aligned with accepted syntax.
- Do not document planned files, packages, directives, or extension APIs as if they already exist.
- Unless the user requests otherwise, commit only coherent, verified units of work and keep unrelated changes out of the commit.

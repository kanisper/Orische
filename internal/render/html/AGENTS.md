# HTML Renderer Guide

## Responsibility Boundary

- Render an already validated `ast.Document` to an `io.Writer`; parsing and AST normalization belong upstream.
- Preserve the AST package's pointer-based node contract. Renderer dispatch uses the exact dynamic pointer type, so a newly accepted AST node must be explicitly registered in `coreSpec`.

## Output and Safety Contracts

- Treat rendered HTML as a stable, exact output format. Whitespace and trailing newlines are covered by tests and must not be reformatted incidentally.
- Escape every AST-supplied value before placing it in HTML text or attributes. Markup emitted by renderer implementations is the only trusted HTML.
- Links are allowlisted by URI scheme. Only `http`, `https`, and `mailto` are accepted, case-insensitively; relative URLs and all other schemes must fail before any link markup is written.
- Return writer and nested-renderer errors with context. Rendering streams directly to the writer and is not transactional, so failures after output begins may leave partial output.

## AST Assumptions

- List items may contain only paragraphs and nested lists. Other block types violate the parser/AST contract and are treated as unreachable.
- Paragraphs inside list items render as the item's inline content, without a surrounding `<p>`; nested lists render inside their own `<li>`.
- Source ranges are metadata and do not affect HTML output.

## Extension Discipline

- Add block and inline renderers through the typed registration helpers in `spec.go`; do not replace dispatch with parsing or source-syntax checks.
- When adding an AST node, update `coreSpec` and add focused tests for dispatch, escaping, exact output, and error propagation as applicable.
- Keep security-sensitive policy, especially URI acceptance and escaping, explicit and covered by negative tests.

## Validation

Run:

```sh
go test ./internal/render/html
```

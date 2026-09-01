# Go Project Layout

## Current Structure

```text
cmd/                              # CLI, diagnostics, and integration tests
docs/                             # syntax and architecture documentation
internal/
  ast/                            # AST definitions
  parser/                         # all parser reading, definitions, and building
  render/html/                    # AST-to-HTML renderer
```

The project remains one Go module. Parser files are separated by responsibility and syntax while remaining in one `package parser`.

## Package Responsibilities

### `internal/ast`

Defines document, block, inline, and range types. AST interfaces are implemented by pointer types and are closed with private marker methods.

### `internal/parser`

- exposes the source-to-AST entrypoints;
- creates the private `spec` with built-in definitions;
- owns the fixed Block Directive and Paragraph readers;
- owns Heading and List sugar readers and the recursive List model;
- owns private parsed-block nodes and AST building;
- dispatches Block Directive builders and preserves diagnostic contracts;
- owns common inline scanning, recursion, fallback, and range calculation.

The `spec` is a small internal configuration structure. It stores the `heading`, `paragraph`, and `code` Block Directive builders, sugar readers in precedence order, and the `em`, `link`, and `code` inline definitions. It is not exposed as a language or plugin API.

### `internal/render/html`

Converts AST nodes to HTML and owns output escaping and URI policy.

### `cmd`

Implements the `orische` command-line entrypoint and file/diagnostic handling.

## Dependency Policy

```text
internal/parser  --->  internal/ast
        |
        +--------->  internal/diagnostic
```

- Keep parser syntax implementations in `package parser`; use file separation for organization and locality.
- Keep the Block Directive envelope and Paragraph fallback in the parser frontend.
- Keep block precedence visible as Directive, Heading/List sugar, Paragraph fallback.
- Lists must use dedicated recursive reading and common `Parser.buildBlock` dispatch for item content.
- Inline definitions provide only type-specific policy, optional validation, and AST construction; the shared envelope and range logic remain in the frontend.
- Keep syntax-specific semantic validation local unless a concrete need justifies a separate validation layer.
- New AST node kinds require coordinated AST, parser, and renderer changes.

## Key Parser Files

This is a responsibility guide, not an exhaustive file listing.

- `parser.go` - entrypoints, document orchestration, dispatch, and AST building
- `spec.go` - private parser configuration and type lookup
- `block_frontend.go` - block context and fixed Block Directive/Paragraph readers
- `block_heading.go` - Heading sugar reader, Heading attribute validation, and shared Directive/sugar builder
- `block_list.go` - List reading, normalization, recursion, and builder
- `block_code.go` - `code` Block Directive builder
- `block_contract.go`, `block_core.go` - small private block handoff and function types
- `inline.go` - inline sequence scanner and Text construction
- `inline_directive.go` - directive envelope, optional validation, and definition dispatch
- `inline_context.go` - byte-offset to Unicode source positions
- `inline_contract.go`, `inline_core.go`, `inline_*` - built-in inline definitions and shared private types

Validate parser work with:

```sh
go test ./internal/parser/...
```

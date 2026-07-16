# Go Project Layout

## Current Structure

```text
cmd/                         # reserved; currently empty
docs/                        # design and behavior documentation
internal/
  ast/                       # AST definitions
  parser/                    # parser, private IR, builders, tests
  render/html/               # HTML renderer under development
```

There is currently no `testdata/` directory or separate syntax-specification package.

## Package Responsibilities

### `internal/ast`

Defines document, block, inline, and range types. AST interfaces are implemented by pointer types.

### `internal/parser`

- splits input into lines;
- parses document blocks into private parsed-block IR;
- parses list structure with dedicated recursive logic;
- builds AST nodes from IR;
- parses inline content during AST building.

Syntax registration is implemented internally in `internal/parser/spec.go`.

### `internal/render/html`

Converts AST nodes to HTML. This package is under development and is not part of parser-only validation.

### `cmd`

Reserved for future command-line entrypoints. No CLI is currently implemented.

## Package Policy

- Keep the project in one Go module.
- Prefer file-level separation before adding packages.
- Keep parser internals in `internal/parser` until a stable public boundary is needed.
- Do not reuse document block parsing as a generic nested-block parser.
- Do not present internal registration as a public extension API.

## Parser Files

The parser package is organized by responsibility:

- `parser.go`, `context.go`, `spec.go` — orchestration and registration
- `parsed_block.go` — private parsed-block IR
- `block_*.go` — document block parsers
- `builder_*.go` — AST builders
- `inline.go` — inline parser
- `*_test.go` — package behavior tests

Validate parser changes with:

```sh
go test ./internal/parser
```

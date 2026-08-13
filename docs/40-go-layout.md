# Go Project Layout

## Current Structure

```text
cmd/                         # HTML-conversion CLI, diagnostics, tests, and test data
docs/                        # design and behavior documentation
internal/
  ast/                       # AST definitions
  parser/                    # parser, private IR, builders, tests
  render/html/               # completed AST-to-HTML renderer and output/security tests
```

Command-line integration fixtures live in `cmd/testdata/`. There is no separate syntax-specification package.

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

Converts AST nodes to HTML. The renderer has exact-output, escaping, URI-policy, dispatch, and error-propagation tests. Parser-only validation does not exercise it.

### `cmd`

Implements the `orische` command-line entrypoint. It reads one input file, renders HTML, writes either the `-o` path or a derived `.html` path, and reports parse, render, and file errors to standard error.

## Package Policy

- Keep the project in one Go module.
- Prefer file-level separation before adding packages.
- Keep parser internals in `internal/parser` until a stable public boundary is needed.
- Do not reuse document block parsing as a generic nested-block parser.
- Do not present internal registration as a public extension API.

## Parser Files

The parser package is organized by responsibility:

- `parser.go`, `block_context.go`, `spec.go` — document orchestration, block-parser context, and registration
- `parsed_block.go` — private parsed-block IR
- `block_*.go` — document block parsers
- `builder_*.go` — AST builders
- `inline.go` — inline-sequence orchestration and `Text` node construction
- `inline_context.go` — byte-offset-to-source-position conversion and inline source ranges
- `inline_directive.go` — inline header parsing and directive-specific AST construction
- `*_test.go` — package behavior tests

Validate parser changes with:

```sh
go test ./internal/parser
```

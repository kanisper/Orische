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

Syntax registration is implemented internally in `internal/parser/spec.go`. `Parser` owns the active `Spec` and coordinates common block building and recursive inline parsing. The registration model is private package machinery, not an extension or plugin API.

### `internal/render/html`

Converts AST nodes to HTML. The renderer has exact-output, escaping, URI-policy, dispatch, and error-propagation tests. Parser-only validation does not exercise it.

### `cmd`

Implements the `orische` command-line entrypoint. It reads one input file, renders HTML, writes either the `-o` path or a derived `.html` path, and reports parse, render, and file errors to standard error.

## Package Policy

- Keep the project in one Go module.
- Prefer file-level separation before adding packages.
- Keep parser internals in `internal/parser` until a stable public boundary is needed.
- Do not reuse the document block reader chain as a generic nested-block reader.
- Do not present internal registration as a public extension API.

## Parser Files

The parser package keeps common parsing machinery separate and groups syntax-specific implementation by syntax:

- `parser.go` — source-to-AST orchestration, active-`Spec` ownership, and common block-builder dispatch
- `spec.go` — responsibility-oriented block feature registration, inline definition registration, normalization, lookup, and validation
- `block_context.go` — short-lived document line cursor state
- `parsed_block.go` — private parsed-block IR
- `block_heading.go`, `block_list.go`, `block_paragraph.go`, `block_code.go` — syntax-specific builder keys, readers where applicable, and private-IR-to-AST builders; list-item builders reuse `Parser` dispatch
- `block_directive.go` — common Block Directive envelope reader
- `inline.go` — `Parser.parseInlines`, inline-sequence state, and `Text` node construction
- `inline_context.go` — byte-offset-to-source-position conversion and inline source ranges
- `inline_directive.go` — common directive-envelope processing, content policies, and the inline definition contract
- `inline_emphasis.go`, `inline_link.go`, `inline_code.go` — core syntax-specific inline definitions
- `spec_registration_test.go`, `inline_spec_test.go`, `active_spec_test.go` — registration consistency and active-`Spec` propagation tests
- `*_test.go` — package behavior tests

Validate parser changes with:

```sh
go test ./internal/parser
```

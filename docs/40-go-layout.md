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

Syntax registration is implemented internally in `internal/parser/spec.go`. Ordinary Block and Inline Definitions own their respective `blockType` and `inlineType` and use symmetric single-definition registration methods. A Block Definition that also implements the reader contract is added to the ordered Sugar reader chain; a builder-only definition is reached through the common Block Directive reader, which remains permanent parser infrastructure rather than a registered feature. Paragraph alone uses dedicated fallback registration to keep its reader last. Inline Sugar is a future extension, not currently implemented. `Parser` owns the active `Spec` and coordinates common block building and recursive inline parsing. The registration model is private package machinery, not an extension or plugin API.

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
- `spec.go` — definition-owned Block/Inline Types, symmetric registration, ordered reader selection, normalization, lookup, and validation
- `block_context.go` — short-lived document line cursor state
- `parsed_block.go` — private parsed-block IR
- `block_heading.go`, `block_list.go`, `block_paragraph.go`, `block_code.go` — syntax-specific definitions combining readers where applicable with private-IR-to-AST builders; list-item builders reuse `Parser` dispatch
- `block_directive.go` — permanent common Block Directive envelope reader infrastructure
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

## Possible Future Parser Split

The parser should remain a single package while the syntax registration contracts are still evolving. File-level, syntax-oriented organization is the current choice; the layout below is a possible future direction, not an implemented package structure or a stable extension API.

A package split becomes worth considering when several of the following are true:

- the number of built-in block and inline syntax types has grown substantially;
- adding a syntax type regularly requires changes to common parser machinery;
- syntax-specific tests are difficult to distinguish from parser-engine tests;
- another internal package needs to assemble or inspect syntax definitions;
- the Reader, Builder, inline Definition, parsed-IR, and source-context contracts have stabilized.

The intended split is between neutral feature contracts and built-in syntax implementations:

```text
internal/parser/
  parser.go                  # orchestration and active specification
  spec.go                    # registration, lookup, and validation
  feature/                   # neutral Reader, Builder, Definition, IR, and context contracts
  syntax/                    # built-in syntax implementations, one syntax per file
    core.go                  # built-in registration set
    block_heading.go
    block_list.go
    block_paragraph.go
    block_code.go
    inline_emphasis.go
    inline_link.go
    inline_code.go
```

Dependencies should remain acyclic:

```text
parser  -> feature
parser  -> syntax -> feature
feature -> ast
syntax  -> ast
```

The `feature` package would contain only the minimum syntax-neutral contracts required to implement and register syntax. Parser orchestration, registration ownership, and scanning behavior would remain outside it. The `syntax` package would contain the built-in implementations and expose their registration set to `parser`; it would not become one subpackage per syntax unless an individual syntax grows large enough to justify that boundary.

Migration should begin by extracting the stable contracts into `feature`, then move the existing implementations into `syntax`. Moving files first would either create an import cycle or require prematurely exporting parser internals. Until those contracts are stable, keep the current single-package layout.

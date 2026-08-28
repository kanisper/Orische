# Go Project Layout

## Current Structure

```text
cmd/                              # CLI, diagnostics, and integration tests
docs/                             # syntax and architecture documentation
internal/
  ast/                            # AST definitions
  parser/                         # frontend orchestration and scanners
    feature/                      # neutral syntax implementation API
    syntax/                       # built-in language assembly
      block/                      # built-in Block definitions and private IR
      inline/                     # built-in Inline definitions
  render/html/                    # AST-to-HTML renderer
```

The project remains one Go module.

## Package Responsibilities

### `internal/ast`

Defines document, block, inline, and range types. AST interfaces are implemented
by pointer types and are closed with private marker methods.

### `internal/parser`

- exposes the source-to-AST entrypoints;
- compiles `feature.Language` into private registries;
- owns fixed Directive and Paragraph Readers;
- installs the fixed Paragraph definition outside `feature.Language`;
- validates Reader results and advances the document position;
- dispatches Block builders and preserves diagnostic contracts;
- owns common inline scanning, recursion, fallback, and range calculation.

### `internal/parser/feature`

Defines the minimum contracts shared by the frontend and syntax packages:

- Block Node, immutable Reader input, and Reader result;
- Block Definition and BuildContext;
- shared text-backed Block IR with an explicit content origin;
- Inline Directive Definition, content policy, and closed candidate;
- immutable-by-convention Language declarations.

It does not import the parser frontend or built-in syntax.

### `internal/parser/syntax`

Assembles the replaceable built-in definitions in `feature.Language`.
`syntax/block` implements Heading, List, and Code Block; `syntax/inline`
implements Emphasis, Link, and Code Span. The parser owns Paragraph as its fixed
fallback definition.

One package is used per syntax category, not per individual syntax. A separate
package for an individual syntax is justified only if that implementation grows
an independent internal model or substantial supporting code.

### `internal/render/html`

Converts AST nodes to HTML and owns output escaping and URI policy.

### `cmd`

Implements the `orische` command-line entrypoint and file/diagnostic handling.

## Dependency Policy

```text
parser -> feature
parser -> syntax -> syntax/block -> feature
                 -> syntax/inline -> feature
feature -> ast
syntax/* -> ast
```

- Syntax packages must not import `parser`.
- `feature` must not own frontend behavior or built-in policy.
- Fixed envelope and fallback Readers remain in `parser`.
- The fixed Paragraph definition is compiled before replaceable Language blocks.
- List-item source parsing must not call the document Reader chain.
- Nested AST construction must use `BuildContext.BuildBlock`.
- Feature contracts are internal implementation APIs, not public plugin APIs.
- New AST node kinds require coordinated AST, syntax, and renderer changes.

## Parser Files

- `parser.go` - entrypoints, document orchestration, dispatch, BuildContext adapter
- `spec.go` - Language validation, compilation, normalization, and lookup
- `block_frontend.go` - immutable input and fixed Block Readers
- `inline.go` - inline sequence scanner and Text construction
- `inline_directive.go` - Directive envelope and definition dispatch
- `inline_context.go` - byte-offset to Unicode source positions
- `feature/*.go` - cross-package implementation contracts
- `syntax/core.go` - built-in Language assembly
- `syntax/block/*.go` - built-in Block implementations
- `syntax/inline/*.go` - built-in Inline implementations

Validate parser work with:

```sh
go test ./internal/parser/...
```

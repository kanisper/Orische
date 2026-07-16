# Orische

Orische is an experimental lightweight markup language focused on explicit structure and predictable parsing. The implementation is written in Go and is under active development; breaking changes may occur.

## Current Capabilities

- headings with levels 1–6;
- multiline paragraphs;
- ordered, unordered, mixed-marker, and recursively nested lists;
- directive-based code blocks;
- inline emphasis, code spans, and links;
- private parsed-block IR and pointer-based AST construction.

The HTML renderer is present but under development. A command-line interface and stable public extension API are not yet implemented.

## Example

```text
= Heading

A paragraph with :[em]{emphasis} and a :[link:https://example.com]{link}.

* parent
*** nested child

:::[code:go]
fmt.Println("hello")
:::
```

## Repository Structure

```text
cmd/                 reserved for future entrypoints; currently empty
docs/                syntax, architecture, layout, and status
internal/ast/        AST definitions
internal/parser/     parser, private IR, builders, and tests
internal/render/html HTML renderer under development
```

## Documentation

- [`docs/20-syntax.md`](docs/20-syntax.md) — current syntax behavior
- [`docs/30-parser-architecture.md`](docs/30-parser-architecture.md) — parser design and invariants
- [`docs/40-go-layout.md`](docs/40-go-layout.md) — package layout
- [`docs/50-roadmap.md`](docs/50-roadmap.md) — current status and future work

Implementation and tests are the source of truth when documentation becomes stale.

## Parser Validation

```sh
go test ./internal/parser
```

## License

BSD 2-Clause License

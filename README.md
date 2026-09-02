# Orische

Orische is an experimental lightweight markup language focused on explicit structure and predictable parsing. The implementation is written in Go and is under active development; breaking changes may occur.

## Current Capabilities

- explicit `heading`, `paragraph`, and `code` Block Directives;
- Heading sugar with levels 1-6;
- Paragraph fallback for ordinary nonblank text;
- ordered, unordered, mixed-marker, and recursively nested List sugar;
- explicit inline Emphasis, Strong, Italic, Bold, Underline, Strikethrough, Code Span, and Link directives;
- constrained inline sugar and ASCII-punctuation backslash escapes;
- source ranges on block and inline AST nodes;
- HTML rendering with escaping and URI-scheme validation;
- a command-line converter from Orische source to HTML.

The parser uses a small private specification for built-in behavior. There is no stable public extension API or dynamic plugin system.

## Syntax Model

Block Directives are the explicit named form for supported block semantics:

```text
:::[heading:level2]
Heading
:::

:::[paragraph]
Paragraph text.
:::

:::[code:go]
fmt.Println("hello")
:::
```

Frequently used structures also have concise forms. Heading marker syntax is sugar, ordinary text falls back to Paragraph, and List marker syntax is currently the only List notation:

```text
= Heading

A paragraph with **strong text**, `code`, and an [external link](https://example.com).

Use \*asterisks\* literally.

* parent
*** nested child
```

An explicit List Directive has not yet been designed.

## Command Line

Convert one Orische source file to HTML:

```sh
go run ./cmd input.oris
```

Specify the output path with `-o`:

```sh
go run ./cmd -o output.html input.oris
```

When `-o` is omitted, the input extension is replaced with `.html`.

## Repository Structure

```text
cmd/                 CLI, diagnostics, and integration tests
docs/                syntax, architecture, layout, and status
internal/ast/        AST definitions
internal/parser/     parser, private parsed nodes, builders, and tests
internal/render/html AST-to-HTML renderer
```

## Documentation

- [`docs/10-language-principles.md`](docs/10-language-principles.md) - language design principles
- [`docs/20-syntax.md`](docs/20-syntax.md) - current syntax behavior
- [`docs/30-parser-architecture.md`](docs/30-parser-architecture.md) - parser design and invariants
- [`docs/40-go-layout.md`](docs/40-go-layout.md) - package layout
- [`docs/50-roadmap.md`](docs/50-roadmap.md) - current status and future work

Implementation and tests are the source of truth when documentation becomes stale.

## Validation

```sh
go test ./...
go vet ./...
```

## License

BSD 2-Clause License

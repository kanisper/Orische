# Project Status

This file records current repository status. It is not a versioned language specification.

## Implemented and Tested

- AST definitions
- Document block parser chain
- Headings with levels 1–6
- Paragraphs
- Ordered and unordered lists
- Recursive list nesting with raw-to-logical level normalization
- `code` block directives
- Inline emphasis, code spans, and links
- Parsed-block IR and AST builders

Parser validation:

```sh
go test ./internal/parser
```

## Under Development

- HTML rendering
- Broader parser edge-case coverage
- Documentation maintenance

## Not Implemented

- Command-line interface
- Stable public extension API
- Structured attributes such as `key=value`
- Additional output formats

## Possible Future Work

- Additional block and inline directive types
- Escaping rules
- Performance measurement and optimization
- Plugin or registration APIs after package boundaries stabilize

# Project Status

This file records current repository status. It is not a versioned language specification.

## Implemented and Tested

- AST definitions
- Document block reader chain
- Headings with levels 1–6
- Paragraphs
- Ordered and unordered lists
- Recursive list nesting with raw-to-logical level normalization
- `code` block directives
- Inline emphasis, code spans, and links
- Parsed-block IR and AST builders
- Internal feature contracts and separated Block/Inline syntax packages
- Immutable language compilation into parser-private registries
- Common `Parser` dispatch for top-level and list-item AST construction
- Source ranges on all block and inline AST nodes
- One-based, inclusive source positions with Unicode-code-point columns
- HTML rendering with escaping and link URI-scheme validation
- Command-line conversion from Orische source files to HTML
- Structured diagnostic errors and CLI diagnostic formatting

Parser validation:

```sh
go test ./internal/parser/...
```

## Under Development

- Broader parser edge-case coverage
- Documentation maintenance

## Not Implemented

- Stable public extension API
- Structured attributes such as `key=value`
- Additional output formats

## Possible Future Work

- Additional block and inline directive types
- Escaping rules
- Performance measurement and optimization
- Plugin or registration APIs after package boundaries stabilize

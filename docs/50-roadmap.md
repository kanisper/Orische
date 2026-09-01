# Project Status

This file records current repository status. It is not a versioned language specification.

## Implemented and Tested

- AST definitions
- Document block reader chain
- `heading`, `paragraph`, and `code` Block Directives
- Heading sugar with levels 1-6
- Paragraph fallback
- Ordered and unordered List sugar
- Mixed-marker Lists with first-marker candidate style
- Recursive List nesting with raw-to-logical level normalization
- Inline Emphasis, Code Span, and Link directives
- Private parsed-block handoff and AST builders
- Private parser `spec` for built-in directive and inline behavior
- Common `Parser` dispatch for top-level and List-item AST construction
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
- Dynamic plugin loading
- Structured attributes such as `key=value`
- Explicit List Block Directive syntax
- Additional output formats

## Possible Future Work

- Design an explicit List Block Directive while preserving current marker sugar semantics
- Configuration for restricting accepted Heading level ranges
- Additional built-in block and inline directive types
- Escaping rules
- Performance measurement and optimization

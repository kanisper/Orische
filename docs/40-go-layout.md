# Go Project Layout

## Core Policy

- Single module
- Minimal package count
- Expand via files first, not packages
- Keep parser internals in `internal/parser` unless a boundary becomes truly stable

---

## Directory Structure

```text
/cmd
  /orische
/internal
  /ast
  /parser
  /render/html
/testdata
/docs
```

---

## Package Roles

### ast

- Defines AST nodes
- Stable layer
- Uses pointer-oriented node contracts where useful

### parser

- splits input into lines
- Parses document-level blocks into parsed block IR
- Parses list structure with dedicated list logic
- Builds AST from parsed block IR
- Parses inline systax during build

### render/html

- Converts AST → HTML

### cmd/orische

- CLI entrypoint
- Reads from stdin or a file path
- Writes rendered HTML to stdout

---

## Parser Growth Strategy

Phase 1 layout:

```text
internal/parser/
  parser.go
  context.go
  spec.go

  parsed_block.go

  block_directive.go
  block_heading.go
  block_list.go
  block_paragraph.go

  builder_code.go
  builder_heading.go
  builder_list.go
  builder_paragraph.go

  inline.go
```

### parser.go

- Parser entrypoint
- Coordinates block parse -> build

### context.go

- Block parser cursor state
- Holds lines and current position

### spec.go

- Parser and builder registration
- Keeps paragraph fallback last
- May remain internal to `parser` until there is a real need for a public syntax configuration package

### parsed_block.go

- Parsed block IR definitions

### block_*.go

- Document-level block parsers
- Block parsers produce parsed block IR, not AST

### builder_*.go

- Converts parsed block IR to AST
- Calls inline parser
- Does not inline-parse code blocks

### inline.go

- Unified inline parser

---

## Avoid

- Multiple modules
- Early package fragmentation
- Generic names such as `util` or `common`
- A public `internal/ast` package before syntax configuration has a stable API
- Reusing document block parsing as a generic neted parser

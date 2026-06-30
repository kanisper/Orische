# Go Project Layout

## Core Policy

- Single module
- Minimal package count
- Expand via files first, not packages

---

## Directory Structure

```text
/cmd
  /medoc
/internal
  /ast
  /parser
  /render/html
  /spec
```

---

## Package Roles

### ast

- Defines nodes
- Stable layer

### parser

- Block + inline parsing
- Expected to grow

### render/html

- Converts AST → HTML

### spec

- Defines enabled syntax

### cmd/medoc

- CLI entrypoint
- Reads from stdin or a file path
- Writes rendered HTML to stdout

---

## Parser Growth Strategy

Start with:

```text
parser/
  parser.go
  context.go
  block.go
  inline.go
```

Then expand into:

```text
parser/
  blocks_heading.go
  blocks_list.go
  blocks_directive.go
  inlines_directive.go
```

Only split packages if necessary.

---

## Avoid

- Multiple modules
- Early package fragmentation
- Generic names (`util`, `common`)

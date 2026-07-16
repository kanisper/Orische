# Parser Architecture

## Overview

Parsing is performed in three internal stages:

1. Block parsing
2. AST building + inline parsing
3. Rendering, outside the parser package

The parser does not build final AST block nodes directly during block parsing.
Instead, it first builds a parsed block IR, then converts that IR into AST.
parsed block node in IR is implemented as a common interface of any parsed block. This can help the AST builder simple and consistent.

## Flow

```text
Text
  ↓ splitLines
[]string
  ↓ block parser
ParsedDocument
  ↓ builder + inline parser
ast.Document
```

Block parser reads structure only.
inline parser is called by the builder for nodes that need inline content.

---

## Block Parsing

- Operates line-by-line
- Uses a parser chain for document-level blocks
- Produces parsed block IR
- Does not call the inline parser
- Does not produce final AST nodes directly

Order:

1. Directive block
2. Heading
3. List
4. Paragraph

The paragraph parser is the fallback parser.
It must remain last and should always parse when reached.

### Document-level only

`parsedBlock` reads document-level blocks only.
It must not be reused as a generic nested block parser.

List parser is handled by dedicated list logic.
The block parser context does not need a generic `nested` flag.

---

## Parsed Block IR

The block parser produces an intermediate representation that is close to the AST shape, but still stores raw text for inline-capable content.

These are defined at `internal/parser/parsed_block.go` file.

### Pointer Policy

Parsed block nodes should usually satisfy `ParsedBlock` through pointer receivers.
This keeps IR node handling consistent with AST node handling and avoids accidental value use.

---

## List Parsing

Lists are parsed by the list parser itself, not by recursively calling the document block parser.

list items may contain only:

- paragraph-like item text
- a single nested list

Not allowed:

- skipped nesting levels

The list parser is rsponsible for:

- read list markers
- read item text
- convert item text to `ParsedParagraph`
- attach a nested ParsedList when present
- stop at blank line or non-list block

---

## Builder

The builder converts parsed blocks into AST blocks.

---

## Inline Parsing

- Character-based scanning
- Unified inline parser
- Called only during AST building

Pattern:

:[type:attr]{content}

---

## Error Handling

The language uses strict parsing with minimal failure scope.

### Rules

- Invalid inline syntax is treated as text.
- Only the smallest invalid inline construct becomes text.
- Surrounding valid inline constructs are preserved.
- Invalid block structure is treated as paragraph text.
- The parser does not auto-correct malformed syntax.

### Examples

Input:

```text
:[em]{valid} :[em]{invalid
```

Result:

- First inline → parsed
- Second inline → treated as text

Input:

```text
:::[code]
missing terminator
```

Result:

- Entire block candidate → paragraph text

Input:

```text
* item
# mixed
```

Result:

- Parsed as one unordered list, because style is determined by the first marker at that level

Input:

```text
* item
**# mixed nested
```

Result:

- Nested list style is determined by the first marker character: `*`

---

## AST Contract (Phase 1)

### Blocks

- Document
- Heading
- Paragraph
- List
- ListItem
- CodeBlock

### Heading

- `Level int`
- `Content []Inline`
- `Range ast.Range`

### Paragraph

- `Content []Inline`
- `Range ast.Range`

### List

- `Ordered bool`
- `Items []ListItem`
- `Range ast.Range`

### ListItem

- `Blocks []Block`
  - Phase 1 parser allows only:
    - Paragraph
    - List, single-level nesting only
- `Range ast.Range`

### CodeBlock

- `Language string`
  - Taken directly from the block directive attribute
  - Example: `:::[code:go]` → `Language = "go"`
- `Text string`
- `Range ast.Range`

### Inline Nodes

Text:

- `Value string`

Emphasis:

- `Content []Inline`

CodeSpan:

- `Value string`

Link:

- `URI string`
- `Content []Inline`

# Parser Architecture

## Overview

Parsing is performed in two stages:

1. Block parsing
2. Inline parsing

## Flow

Text → Block AST → Inline AST → Final AST

---

## Block Parsing

- Operates line-by-line
- Uses parser chain

Order:

1. Code block
2. Heading
3. List
4. Paragraph

---

## Inline Parsing

- Character-based scanning
- Unified inline parser

Pattern:

:[type:attr]{content}

---

## AST Structure

### Blocks

- Document
- Heading
- Paragraph
- List
- ListItem
- CodeBlock

### Inline

- Text
- Emphasis
- CodeSpan
- Link

---

## Error Handling

- Strict mode
- Invalid syntax → Text node

## Error Handling (Strict Mode)

Invalid syntax is handled using **minimal failure scope**.

### Rules

- Only the smallest invalid construct is treated as plain text
- Surrounding valid structures are preserved
- The parser does NOT discard entire blocks unless the block structure itself is invalid

### Examples

Input:
:[em]{valid} :[em]{invalid

Result:
- First inline → parsed
- Second inline → treated as text

Input:
:::[code]
missing terminator

Result:
- Entire block treated as paragraph text

Input:
* item
# mixed

Result:
- Parsed as a single unordered list

Input:
* item
**# mixed nested

Result:
- Nested list style is determined by the first marker character (`*`)

## AST Contract (Phase 1)

### Heading

- Level: int
- Content: []Inline
- Range: source span

### Paragraph

- Content: []Inline
- Range

### List

- Ordered: bool
- Items: []ListItem
- Range

### ListItem

- Blocks: []Block
  - Allowed in Phase 1:
    - Paragraph
    - List (single-level nesting only)
- Range

### CodeBlock

- Language: string
  - In Phase 1, this is taken directly from the block directive attribute
  - Example: `:::[code:go]` → `Language = "go"`
- Text: string
- Range

### Inline Nodes

Text:
- Value: string

Emphasis:
- Content: []Inline

CodeSpan:
- Value: string

Link:
- URI: string
- Content: []Inline

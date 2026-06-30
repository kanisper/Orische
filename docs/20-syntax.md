# Syntax Specification (v0.1)

## Blocks

### Heading

```text
= Heading1
== Heading2
```

### Paragraph

- Separated by blank lines

### Unordered List

```text
* item
** nested
```

### Ordered List

```text
# item
## nested
```

### List Marker Style

List style is determined by the first marker at the same level.

```text
* unordered
# still unordered
```

If a marker run mixes `*` and `#`, the first character determines the style.

```text
**# nested unordered
```

### Code Block

```text
:::[code]
code here
:::

:::[code:go]
fmt.Println("hello")
:::
```

---

## Inline Elements

### General Form

```text
:[type]{content}
:[type:attr]{content}
```

### Supported Types

#### Emphasis

```text
:[em]{text}
```

#### Code

```text
:[code]{x := 1}
```

#### Link

```text
:[link:https://example.com]{Example}
```

---

## Strict Rules

Invalid patterns are NOT parsed:

- Missing closing `}`
- Missing `]`
- Invalid attribute structure
- Unknown inline types

All such cases are treated as plain text.

---

## Attribute Grammar (Phase 1)

Attributes are treated as a **single opaque string**.

### Syntax

:[type:attr]{content}

### Parsing Rule

- The **first colon** separates `type` and `attr`
- Everything after the first colon is part of the attribute string
- Additional colons are allowed inside the attribute

### Examples

:[link:https://example.com]{Example}
:[code:go]{fmt.Println("hello")}
:[custom:a:b:c]{value}

### Invalid Cases

- Missing type: :[:attr]{x}
- Missing closing `]` or `}`
- Empty type name

### Notes

The parser does NOT interpret attribute structure in Phase 1.

---

## List Rules (Phase 1)

### Allowed

- Single-line list item text
- Single-level nested lists
- Continuous list blocks without blank lines

### Not Allowed (Phase 1)

- Blank lines inside lists
- Multi-paragraph list items
- Nested blocks inside list items, **except for a single nested list**
- Skipping nesting levels
- Code blocks inside list items

Invalid list candidates are treated as paragraph text.

### Termination

A list ends when:

- A blank line is encountered
- A non-list block starts

---

## Block Directive Grammar (Phase 1)

### General Form

```text
:::[type]
content
:::

:::[type:attr]
content
:::
```

### Parsing Rule

- The first colon separates `type` and `attr`
- Everything after the first colon is part of the attribute string
- Additional colons are allowed inside the attribute

### Phase 1 Interpretation

- The parser treats the attribute as a single opaque string
- For `code`, the attribute value is assigned to the `Language` field of the AST node
- For `code`, `:::[code:go]` sets `Language` to `go`

### Examples

```text
:::[code]
fmt.Println("hello")
:::

:::[code:go]
fmt.Println("hello")
:::
```

### Invalid Cases

- Missing closing `]`
- Missing closing block terminator `:::`
- Empty block type

Invalid block directives are treated as paragraph text.

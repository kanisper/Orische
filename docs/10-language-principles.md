# Language Principles

## 1. Strict Parsing

- Invalid syntax is NOT tolerated
- Invalid constructs are treated as plain text

## 2. Explicit Structure

- No implicit parsing magic
- All constructs must be clearly delimited

## 3. Syntax Strategy

This language uses a **mixed syntax model**:

- Simple marker-based syntax for common structures
  - Headings (`=`)
  - Lists (`*`, `#`)

- Directive-based syntax for extensible blocks
  - :::[type:attr] ... :::

This design prioritizes readability for common cases and extensibility for advanced features.

- Inline syntax:
  :[type:attr]{content}

## 4. Minimal Ambiguity

- No Markdown-style loose parsing
- No auto-correction

## 5. Extensibility

- Syntax can be extended later through explicit parser and builder registration
- Future expansion without breaking core

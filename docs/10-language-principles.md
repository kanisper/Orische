# Language Principles

## Predictable Parsing

Syntax is explicit and parsing is deterministic. The parser does not auto-correct malformed input.

Malformed block candidates normally fall through to paragraph parsing. Malformed or unsupported inline candidates remain literal text. Syntactically valid block directives require a registered builder and may return an error if unsupported.

## Explicit Structure

Common document structures use marker-based syntax:

- headings use `=`;
- lists use `*` and `#`.

Extensible block and inline forms use explicit delimiters:

```text
:::[type:attribute]
content
:::

:[type:attribute]{content}
```

Attributes are optional opaque strings. The first colon separates the type from the attribute.

## Separation of Concerns

The implementation separates:

1. document block parsing;
2. parsed-block IR;
3. AST building and inline parsing;
4. rendering.

Syntax registration exists internally. A stable public extension API is not currently provided.

# Language Principles

## Predictable Parsing

Syntax is explicit and parsing is deterministic. The parser does not auto-correct malformed input.

Malformed block candidates normally fall through to paragraph parsing. Unsupported inline directives and invalid inline headers are emitted as literal source text rather than errors. Other malformed or unterminated inline candidates resume ordinary scanning, so a later valid inline sequence may still be recognized. Syntactically valid block directives require a built-in definition and return an error if unsupported.

## Explicit Structure

Common document structures use marker-based syntax:

- headings use `=`;
- lists use `*` and `#`.

Directive forms use explicit delimiters:

```text
:::[type:attribute]
content
:::

:[type:attribute]{content}
```

Attributes are optional opaque strings. The first colon separates the type from the attribute.

## Separation of Concerns

The implementation separates:

1. document block reading;
2. private parsed-block handoff and AST building;
3. inline scanning and definition-specific construction;
4. rendering.

`Parser` owns source-to-AST orchestration. Its private `spec` contains the
built-in directive builders, sugar readers, and inline definitions. Files are
separated by syntax within `package parser`; there is no separate syntax
package or public extension contract.

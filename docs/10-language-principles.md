# Language Principles

## Predictable Parsing

Syntax is explicit and parsing is deterministic. The parser does not auto-correct malformed input.

Malformed block candidates normally fall through to paragraph parsing. Unsupported inline directives and invalid inline headers are emitted as literal source text rather than errors. Other malformed or unterminated Inline Directive candidates resume ordinary scanning, so a later valid inline sequence may still be recognized. An unterminated inline Sugar candidate instead keeps the rest of its logical line literal, preventing ambiguous content from being reinterpreted. Syntactically valid Block Directives require a built-in definition and return an error if unsupported.

A structurally valid directive may still contain an invalid required semantic value. Those cases are semantic errors rather than syntax fallback. For example, Heading requires `level1` through `level6`; missing, nonnumeric, zero, and out-of-range levels are errors.

## Explicit Structure and Concise Forms

Block Directives are the explicit named form for supported block semantics:

```text
:::[type:attribute]
content
:::
```

The current built-in Block Directive types are `heading`, `paragraph`, and `code`.

Frequently used structures may also have concise notation:

- Heading uses `=` marker syntax as sugar;
- Paragraph uses ordinary nonblank text as a fallback form;
- List currently uses `*` and `#` marker syntax; an explicit List Directive has not yet been designed.

Inline directives use the same explicit type-and-attribute idea in a compact envelope:

```text
:[type:attribute]{content}
```

Inline Directives are the canonical forms for inline semantics. Common inline elements also have constrained Sugar forms: `*`, `**`, `_`, `__`, `++`, `~~`, a single backtick, and `[label](URI)`. Sugar is recognized only on one logical line with explicit outer boundaries. It does not attempt CommonMark-compatible delimiter inference or correction.

A backslash escapes one following ASCII punctuation character in recursively parsed inline content. It does not change literal Code Span or Code Block content.

Attributes are optional opaque strings. The first colon separates the type from the attribute. A syntax interprets and validates an attribute only when it is semantically required. Attributes that a syntax does not use are accepted and ignored.

Current semantically meaningful attributes are:

- Heading Block Directive: `level1` through `level6`;
- Code Block Directive: language identifier;
- Link inline directive: nonempty URI.

Current ignored attributes include Paragraph, Emphasis, Strong, Italic, Bold, Underline, Strikethrough, and Code Span attributes.

## Separation of Concerns

The implementation separates:

1. document block reading;
2. private parsed-block handoff and AST building;
3. inline scanning and definition-specific construction;
4. rendering.

`Parser` owns source-to-AST orchestration. Its private `spec` contains the built-in directive builders, sugar readers, and inline definitions. Files are separated by syntax within `package parser`; there is no separate syntax package or public extension contract.

Validation currently remains local to the syntax behavior that needs it. A separate validation layer or package is not part of the current architecture.

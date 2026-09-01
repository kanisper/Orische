# Parser Architecture

## Pipeline

```text
input string
  -> splitLines
  -> block readers
  -> private parsed-block nodes
  -> AST builders and inline parser
  -> ast.Document
```

Rendering is outside the parser package. `Parser` owns the complete source-to-AST operation. The same parser state is used for top-level blocks, List items, inline-capable blocks, and recursive inline directives.

## Package Model

```text
internal/ast
      ^
      |
internal/parser  --->  internal/diagnostic
```

All parser implementation files belong to one `package parser`. Files are separated by responsibility and syntax, while built-in syntax stays local to the parser. `Parser` keeps a private `spec` containing the small amount of configuration that remains useful: Block Directive builders, sugar readers, and inline definitions. This is an internal data structure, not a plugin API.

## Spec

`newSpec` installs the built-in definitions:

- the `heading`, `paragraph`, and `code` Block Directive builders;
- Heading and List sugar readers, in precedence order;
- the `em`, `link`, and `code` inline definitions.

`NewParser` creates this state, and there is no public registration method. Tests in `package parser` may install small private fixtures when they need to exercise definition-driven parser behavior; callers outside the package cannot replace the built-in syntax.

Type lookup normalizes only the type with `strings.ToLower`. Attributes and content retain their original spelling. Block and inline type maps are independent, so the name `code` can be used in both categories.

## Block Reading

The effective order is fixed:

1. common Block Directive envelope reader;
2. Heading and List sugar readers in `spec.sugars` order;
3. common Paragraph fallback reader.

`readBlockDirective` and `readParagraph` are fixed parser infrastructure. Heading and List are built-in sugar syntax. Readers operate on an immutable-by-convention position represented by `blockContext` and report a private parsed node plus the number of consumed lines. A failed reader returns no node and zero consumption; a successful reader consumes at least one line.

Malformed candidates normally fall through to the Paragraph reader. A structurally valid Block Directive is read first even when its type is unsupported. Builder lookup then produces an unsupported-directive diagnostic.

## Parsed Blocks and AST Building

Block readers record structure, source ranges, and raw text in private node types such as `blockDirectiveNode`, `headingNode`, `listNode`, and `paragraphNode`. These nodes are the local handoff between reading and AST building; they are not a public intermediate representation.

`Parser.buildBlock` dispatches those concrete nodes. It performs normalized Block Directive lookup, creates unsupported-directive diagnostics, and wraps ordinary builder errors with block context. Paragraph and Heading Directive builders reuse the same final construction methods as Paragraph fallback and Heading sugar, respectively. Paragraph and Heading parse their text with the parser's inline scanner. Code Block preserves its content literally. Every successfully built block retains the range recorded while reading.

Heading Directive validation currently lives with Heading construction. The required attribute must be `level1` through `level6`; invalid required values are semantic build errors. Attributes that a syntax does not use are accepted and ignored. No separate validation package or validation stage exists in the current architecture.

## Lists

Lists use dedicated recursive reading rather than the document block reader chain. `readList` consumes one consecutive run of List lines and builds private nested List nodes. Raw marker levels are normalized as follows:

```text
first line: logical = 1
raw increase: logical = previous logical + 1
same raw level: logical = previous logical
raw decrease: logical = max(1, previous logical - raw difference)
```

A line's first marker determines its candidate ordered/unordered style, including mixed-marker lines. The first item at each logical List level determines the constructed List style for that level.

During AST construction, List item Paragraphs and nested Lists go through the same `Parser.buildBlock` dispatch as top-level blocks. This preserves inline parsing, diagnostics, and ranges consistently at every nesting level.

## Inline Parsing

The frontend scanner owns the common `:[...]{...}` envelope, header splitting, brace handling, recursion, literal fallback, and source ranges.

The private `inlineDefinition` for a directive supplies only its content policy, optional attribute validation, and AST construction function. Nested definitions are parsed recursively; literal definitions receive content through the first closing brace without inline parsing. The frontend confirms structural closure before validation. A rejected or unsupported candidate becomes literal text.

Inline definition construction itself does not return an error. Parser errors are reserved for invalid internal definition state, such as an unknown content policy, a missing builder function, or a builder returning a nil AST node.

## Ranges and Errors

Parser-produced non-empty ranges are one-based and inclusive. Columns count Unicode code points rather than UTF-8 bytes. Inline Directive ranges cover the complete source syntax; nested nodes and literal `Text` nodes carry their own spans.

Malformed syntax normally falls back to Paragraph or literal text according to the syntax rules. Errors are used for unsupported valid Block Directives, invalid required semantic values such as a Heading level, and inconsistent internal parser state. Existing diagnostic identity and nested error context are preserved while blocks are built.

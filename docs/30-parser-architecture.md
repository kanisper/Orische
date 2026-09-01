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

Rendering is outside the parser package. `Parser` owns the complete
source-to-AST operation. The same parser state is used for top-level blocks,
list items, inline-capable blocks, and recursive inline directives.

## Package Model

```text
internal/ast
      ^
      |
internal/parser  --->  internal/diagnostic
```

All parser implementation files belong to one `package parser`. Files are
separated by responsibility and syntax, while built-in syntax stays local to
the parser. `Parser` keeps a private `spec` containing the small amount
of configuration that remains useful: block directive builders, sugar readers,
and inline definitions. This is an internal data structure, not a plugin API.

## Spec

`newSpec` installs the built-in definitions:

- the `heading`, `paragraph`, and `code` Block Directive builders;
- Heading and List sugar readers, in precedence order;
- the `em`, `link`, and `code` inline definitions.

`NewParser` creates this state, and there is no public registration method.
Tests in `package parser` may install a small private inline fixture when they
need to exercise definition-driven parser behavior; callers outside the package
cannot replace the built-in syntax.

Type lookup normalizes only the type with `strings.ToLower`. Attributes and
content retain their original spelling. Block and inline type maps are
independent, so the name `code` can be used in both categories.

## Block Reading

The effective order is fixed:

1. common Block Directive envelope reader;
2. Heading and List sugar readers in `spec.sugars` order;
3. common Paragraph fallback reader.

`readBlockDirective` and `readParagraph` are fixed parser infrastructure.
Heading and List are built-in sugar syntax. Readers operate on an immutable
position represented by `blockContext` and report a private parsed node plus
the number of consumed lines. A failed reader returns no node and zero
consumption; a successful reader always consumes at least one line.

Malformed candidates normally fall through to the Paragraph reader. A valid
Block Directive is read first even when its type is not supported. Its builder
lookup then produces an unsupported-block diagnostic.

## Parsed Blocks and AST Building

Block readers record structure, source ranges, and raw text in private node
types such as `blockDirectiveNode`, `headingNode`, `listNode`, and
`paragraphNode`. These nodes are the local handoff between reading and AST
building; they are not a public intermediate representation.

`Parser.buildBlock` dispatches those concrete nodes. It performs normalized
Block Directive lookup, creates unsupported-directive diagnostics, and wraps
ordinary builder errors with the block type. Paragraph and Heading Directive
builders reuse the same final construction methods as Paragraph fallback and
Heading sugar, respectively. Paragraph and Heading parse their text with the
parser's inline scanner. Code Block preserves its content literally. Every
built block retains the range recorded while reading.

## Lists

Lists use dedicated recursive reading rather than the document block reader
chain. `readList` consumes one consecutive run of list lines and builds private
nested list nodes. Raw marker levels are normalized as follows:

```text
first line: logical = 1
raw increase: logical = previous logical + 1
same raw level: logical = previous logical
raw decrease: logical = max(1, previous logical - raw difference)
```

During AST construction, list item paragraphs and nested lists go through the
same `Parser.buildBlock` dispatch as top-level blocks. This preserves inline
parsing and diagnostics consistently at every nesting level.

## Inline Parsing

The frontend scanner owns the common `:[...]{...}` envelope, header splitting,
brace handling, recursion, literal fallback, and source ranges.

The private `inlineDefinition` for a directive supplies only its content
policy, optional attribute validation, and AST construction. Nested definitions
are parsed recursively; literal definitions receive content through the first
closing brace without inline parsing. The frontend confirms structural closure
before validation. A rejected or unsupported candidate becomes literal text,
while a builder error aborts parsing.

## Ranges and Errors

Parser-produced non-empty ranges are one-based and inclusive. Columns count
Unicode code points rather than UTF-8 bytes. Directive-node ranges cover the
complete source syntax; nested nodes and literal `Text` nodes carry their own
spans.

Malformed syntax normally falls back to text. Errors are reserved for
unsupported valid Block Directives and inconsistent internal states such as a
missing builder or a builder returning a nil AST node. Existing diagnostic
identity and nested error context are preserved while blocks are built.

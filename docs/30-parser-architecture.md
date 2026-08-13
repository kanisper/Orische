# Parser Architecture

## Pipeline

```text
input string
  → splitLines
  → document block parsers
  → private parsed-block IR
  → AST builders and inline parser
  → ast.Document
```

Rendering is outside the parser package.

## Block Parser Chain

`parseBlocks` processes nonblank lines. `parseOneBlock` tries parsers in this order:

1. `blockDirectiveParser`
2. `headingParser`
3. `listParser`
4. `paragraphParser` fallback

The fallback is stored separately in `Spec` and appended last by `getParsers`. It must remain last.

Block parsing records structure and raw inline-capable text. It does not create final AST blocks or call the inline parser.

## Context Cursor Contract

`blockContext.pos` is a zero-based line index. Parser-produced non-empty source ranges use one-based line and column positions with inclusive endpoints. Columns count Unicode code points rather than UTF-8 bytes.

On successful parsing, a block parser leaves `pos` on the last consumed line. `parseBlocks` then calls `ctx.advance(1)`.

A parser returning `ok=false` must not consume input. A parser that scans ahead must restore its starting position before returning false. This is required for paragraph fallback.

`getLine` assumes the context is not at EOF; callers enforce this precondition.

## Parsed-Block IR

`parsedBlockNode` is the common private IR interface. Its current implementations are pointer types:

- `*parsedBlock` for headings, paragraphs, and directives;
- `*parsedList` for lists.

The IR keeps raw text until AST building. `getBuilderKey` selects the AST builder.

## List Parsing

Lists use dedicated recursive parsing and do not invoke the document block parser for item content.

`collectListLines` records each line's raw marker level. `normalizeListLevel` converts raw changes into logical levels:

```text
first line: logical = 1
raw increase: logical = previous logical + 1
same raw level: logical = previous logical
raw decrease: logical = max(1, previous logical - raw difference)
```

Any raw increase creates exactly one deeper logical level. `buildParsedList` recursively constructs nested lists, with no explicit depth limit.

Each list item initially contains a paragraph-like `*parsedBlock`. A nested `*parsedList` is appended to its parent item's blocks.

## AST Building and Inline Parsing

Builders convert private IR into pointer-based AST nodes.

- Heading and paragraph builders parse inline content.
- List builders recursively build item blocks.
- Code builders preserve text literally and do not parse inline syntax.
- Parsed block ranges are copied to their AST block nodes. Inline parsing computes ranges from each block's inline-content origin, and every inline node receives its own source range.

The core builder keys are `heading`, `paragraph`, `list`, and `code`.

## Inline Parser

The inline parser scans UTF-8 text by byte offset. `inlineContext` maps valid byte boundaries to one-based lines and Unicode-code-point columns. It recognizes:

- emphasis with recursively parsed content;
- links with recursively parsed content;
- code spans with literal content.

Empty content is valid. Unsupported directives, invalid headers, and links without a URI are retained as literal text through the first available `}`. Other malformed or unterminated candidates resume ordinary scanning, so later valid inline syntax may still be recognized. Code spans end at the first `}` and do not support nesting or escaping.

## Error Model

Malformed block candidates normally return `ok=false` and fall through to paragraph parsing. Inline fallback does not produce user-facing parse errors: unsupported or invalid-header forms are retained literally through the first available `}`, while other malformed forms resume ordinary scanning.

Errors are reserved for unsupported or inconsistent internal states, including:

- no builder registered for a parsed block type;
- a builder receiving the wrong parsed-node type;
- an unsupported block inside a list item;
- unconsumed list lines.

A syntactically valid block directive with an unregistered type is a user-reachable AST build error.

## AST Contract

AST interfaces use pointer implementations.

### Blocks

- `*ast.Heading`
- `*ast.Paragraph`
- `*ast.List`
- `*ast.CodeBlock`

`ast.List.Items` is `[]*ast.ListItem`. List-item blocks currently contain paragraph and nested-list nodes produced by the list parser.

### Inline nodes

- `*ast.Text`
- `*ast.Emphasis`
- `*ast.CodeSpan`
- `*ast.Link`

Each inline node contains an `ast.Range`. Directive-node ranges include the complete source directive and delimiters. Nested inline and literal `Text` nodes have independent ranges covering their own source spans.

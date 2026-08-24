# Parser Architecture

## Pipeline

```text
input string
  → splitLines
  → document block readers
  → private parsed-block IR
  → AST builders and Parser.parseInlines
  → ast.Document
```

Rendering is outside the parser package.

`Parser` owns the active `Spec` for the entire source-to-AST operation. The same parser is passed to every block builder and retained by recursive inline parsing; nested work does not construct a replacement core specification. There is no separate build-context type.

## Block Reader Chain

`parseBlocks` processes nonblank lines. `parseOneBlock` tries readers in this order:

1. `blockDirectiveReader`
2. `headingReader`
3. `listReader`
4. `paragraphReader` fallback

The fallback is stored separately in `Spec` and appended last by `getReaders`. It must remain last.

Block reading records structure and raw inline-capable text. It does not create final AST blocks or perform inline parsing.

## Spec Registration

`Spec` is the internal source of block and inline feature definitions. Its registration operations reflect parser responsibilities:

- the shared Block Directive reader is registered exactly once and always precedes sugar readers;
- a block directive definition associates a normalized Directive Type with its builder;
- Heading and List are registered as ordered sugar Reader/Builder pairs;
- Paragraph is registered as the dedicated fallback Reader/Builder pair;
- inline directive definitions associate a normalized Directive Type with validation, a content policy, and AST construction.

Sugar and fallback registration install their reader and builder atomically. Duplicate, empty, nil, and case-only-colliding definitions are rejected without replacing an existing definition. A document `Spec` is validated before source reading, including the required Block Directive reader and Paragraph fallback.

Directive Type registration and lookup use the same Unicode-aware `strings.ToLower` normalization. Only the type is normalized; attributes and content retain their original spelling. Block and inline definitions use separate registries, so `code` can be defined in both categories.

The registration model is private to `internal/parser`. It is not a stable extension or plugin API.

## Context Cursor Contract

`blockContext.pos` is a zero-based line index. Parser-produced non-empty source ranges use one-based line and column positions with inclusive endpoints. Columns count Unicode code points rather than UTF-8 bytes.

On a successful read, a block reader leaves `pos` on the last consumed line. `parseBlocks` then calls `ctx.advance(1)`.

A reader returning `ok=false` must not consume input. A reader that scans ahead must restore its starting position before returning false. This is required for paragraph fallback.

`getLine` assumes the context is not at EOF; callers enforce this precondition.

## Parsed-Block IR

`parsedBlockNode` is the common private IR interface. Its current implementations are pointer types:

- `*parsedBlock` for headings, paragraphs, and directives;
- `*parsedList` for lists.

The IR keeps raw text until AST building. `getBuilderKey` supplies the raw internal lookup key; `Spec` normalizes the key and selects the AST builder.

## List Parsing

Lists use dedicated recursive reading and do not invoke the document block reader chain for item content.

`collectListLines` records each line's raw marker level. `normalizeListLevel` converts raw changes into logical levels:

```text
first line: logical = 1
raw increase: logical = previous logical + 1
same raw level: logical = previous logical
raw decrease: logical = max(1, previous logical - raw difference)
```

Any raw increase creates exactly one deeper logical level. `buildParsedList` recursively constructs nested lists, with no explicit depth limit.

Each list item initially contains a paragraph-like `*parsedBlock`. A nested `*parsedList` is appended to its parent item's blocks.

List-item AST construction uses the same `Parser.buildBlock` dispatch as top-level document construction. Paragraph-like item IR reaches the registered Paragraph builder, and nested-list IR reaches the registered List builder. The dedicated recursive List reader remains independent of the document reader chain.

## AST Building and Inline Parsing

Builders convert private IR into pointer-based AST nodes. `Parser.buildBlock` owns normalized builder lookup, unsupported-block diagnostics, and internal error wrapping for both top-level and list-item blocks. Builders receive the active `Parser`; inline-capable builders call `Parser.parseInlines`, which preserves access to the parser's `Spec`.

- Heading and Paragraph builders parse inline content.
- List builders recursively build item blocks through common dispatch.
- Code Block builders preserve text literally and do not parse inline syntax.
- Parsed block ranges are copied to their AST block nodes. Inline parsing computes ranges from each block's inline-content origin, and every inline node receives its own source range.

The core builder keys are `heading`, `paragraph`, `list`, and `code`.

## Inline Parsing

`Parser.parseInlines` scans UTF-8 text by byte offset with a short-lived `inlineParseState`. `inlineContext` maps valid byte boundaries to one-based lines and Unicode-code-point columns. Common scanning owns the `:[...]{...}` envelope, header splitting, delimiter handling, recursion, text flushing, and range calculation.

After parsing a header, the active `Spec` selects an inline directive definition by normalized type. Each definition owns attribute validation, one of two content policies, and AST construction:

- emphasis with recursively parsed content;
- links with recursively parsed content;
- code spans with literal content.

Empty content is valid. Unsupported directives, invalid headers, and links without a URI are retained as literal text through the first available `}`. Other malformed or unterminated candidates resume ordinary scanning, so later valid inline syntax may still be recognized. Code spans end at the first `}` and do not support nesting or escaping.

Semantic rejection, such as a link with an empty URI, produces literal fallback. Definition validation or construction errors represent internal failures and are returned as errors. Recursive nested content always consults the same active `Parser` and `Spec`.

## Error Model

Malformed block candidates normally return `ok=false` and fall through to paragraph parsing. Inline fallback does not produce user-facing parse errors: unsupported or invalid-header forms are retained literally through the first available `}`, while other malformed forms resume ordinary scanning.

Errors, rather than syntax fallback, are used for a user-reachable unsupported block definition and for inconsistent internal states, including:

- no builder registered for a parsed block type;
- a builder receiving the wrong parsed-node type;
- an inline definition returning an error or nil AST node;
- unconsumed list lines.

A syntactically valid block directive with an unregistered type is a user-reachable AST build error.

## AST Contract

AST interfaces use pointer implementations.

### Blocks

- `*ast.Heading`
- `*ast.Paragraph`
- `*ast.List`
- `*ast.CodeBlock`

`ast.List.Items` is `[]*ast.ListItem`. List-item blocks currently contain paragraph and nested-list nodes produced by the list reader.

### Inline nodes

- `*ast.Text`
- `*ast.Emphasis`
- `*ast.CodeSpan`
- `*ast.Link`

Each inline node contains an `ast.Range`. Directive-node ranges include the complete source directive and delimiters. Nested inline and literal `Text` nodes have independent ranges covering their own source spans.

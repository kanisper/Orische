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
2. `headingDefinition`
3. `listDefinition`
4. `paragraphDefinition` fallback

The Block Directive reader and Paragraph reader are fixed parser infrastructure. `getReaders` constructs the chain with the Directive reader first, registered Sugar definitions in the middle, and the Paragraph reader last.

Block reading records structure and raw inline-capable text. It does not create final AST blocks or perform inline parsing.

## Spec Registration

`Spec` is the internal source of block and inline feature definitions. Definitions own their type, and the registration surface is symmetric between extensible Block and Inline features:

- every extensible Block Definition combines `blockType() string` with AST building and is registered through `registerBlock(definition)`;
- if a Block Definition also implements `blockReader`, registration appends it to the ordered Sugar reader chain; otherwise it is reached only through the shared Block Directive envelope reader;
- the shared Block Directive reader is permanent parser infrastructure, not a `Spec` registration target, and always precedes reader-capable Block Definitions;
- Heading and List are reader-capable Block Definitions and are registered in reader order, while Code is a directive-only Block Definition;
- Paragraph combines reading and building as a fixed definition installed by `newSpec`; it is present in the block-definition map and always supplies the final reader without registration by `coreSpec`;
- every Inline Definition declares `inlineType() string` through the base `inlineDefinition` contract and is registered through `registerInline(definition)`;
- current Inline Directive Definitions additionally own validation, content policy, and inline AST construction; other definition categories are rejected until their parser contract exists;
- this definition-owned type and single-definition registration shape allows future Inline Sugar to be added without changing the registration API, while not claiming that Inline Sugar syntax exists today.

`paragraph` is a case-insensitive reserved Block Type owned by fixed parser infrastructure. `newSpec` installs the standard Paragraph definition, and general Block registration rejects attempts to replace it. Registration validates all inputs before mutating the block-definition map or ordered Sugar reader list. Duplicate, empty, nil, and case-only-colliding definitions are rejected without replacing an existing definition. A document `Spec` is validated before source reading by checking that its block-definition map contains the Paragraph definition; this also rejects a zero-value or otherwise incomplete `Spec`. The common Block Directive reader does not require registration.

Syntax Type registration and lookup use the same Unicode-aware `strings.ToLower` normalization. Only the type is normalized; attributes and content retain their original spelling. Block and inline definitions use separate registries, so `code` can be defined in both categories.

The registration model is private to `internal/parser`. It is not a stable extension or plugin API.

## Context Cursor Contract

`blockContext.pos` is a zero-based line index. Parser-produced non-empty source ranges use one-based line and column positions with inclusive endpoints. Columns count Unicode code points rather than UTF-8 bytes.

On a successful read, a block reader leaves `pos` on the last consumed line. `parseBlocks` then increments `pos` once.

A reader returning `ok=false` must not consume input. A reader that scans ahead must restore its starting position before returning false. This is required for fallback to the fixed Paragraph reader.

`blockContext.line` assumes the context is not at EOF; callers enforce this precondition.

## Parsed-Block IR

`parsedBlockNode` is the common private IR interface. Its current implementations are pointer types:

- `*parsedBlock` for headings, paragraphs, and directives;
- `*parsedList` for lists.

The IR keeps raw text until AST building. `blockType()` supplies the raw internal Block Type; `Spec` normalizes it and selects the AST builder.

When an ordered Block Definition succeeds while reading, `parseOneBlock` compares its normalized declared Block Type with the normalized type returned by the parsed IR immediately after the read. A mismatch is an ordinary internal error, not an unsupported-block diagnostic, and parsing does not continue to Paragraph fallback. The check also rejects a successful definition that returns a nil IR node.

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

List-item AST construction uses the same `Parser.buildBlock` dispatch as top-level document construction. Paragraph-like item IR reaches the standard Paragraph definition installed by `newSpec`, and nested-list IR reaches the registered List builder. The dedicated recursive List reader remains independent of the document reader chain.

## AST Building and Inline Parsing

Builders convert private IR into pointer-based AST nodes. `Parser.buildBlock` owns normalized builder lookup, unsupported-block diagnostics, and internal error wrapping for both top-level and list-item blocks. Builders receive the active `Parser`; inline-capable builders call `Parser.parseInlines`, which preserves access to the parser's `Spec`.

- Heading and Paragraph builders parse inline content.
- List builders recursively build item blocks through common dispatch.
- Code Block builders preserve text literally and do not parse inline syntax.
- A diagnostic returned by a list-item Paragraph builder is propagated as the same diagnostic instance, without `build "paragraph" block` or `build "list" block` context. A non-diagnostic list-item builder or inline error preserves its cause and receives the nested `build "paragraph" block` and outer `build "list" block` context.
- Parsed block ranges are copied to their AST block nodes. Inline parsing computes ranges from each block's inline-content origin, and every inline node receives its own source range.

The core Block Types are declared by `blockTypeHeading`, `blockTypeParagraph`, `blockTypeList`, and `blockTypeCode`. The core Inline Types are declared by `inlineTypeEmphasis`, `inlineTypeLink`, and `inlineTypeCode`.

## Inline Parsing

`Parser.parseInlines` scans UTF-8 text by byte offset with a short-lived `inlineParseState`. `inlineContext` maps valid byte boundaries to one-based lines and Unicode-code-point columns. Common scanning owns the `:[...]{...}` envelope, header splitting, delimiter handling, recursion, text flushing, and range calculation.

After parsing a header, the active `Spec` selects an inline directive definition by normalized type. Each definition owns attribute validation, one of two content policies, and AST construction:

- emphasis with recursively parsed content;
- links with recursively parsed content;
- code spans with literal content.

Empty content is valid. The common parser first applies the definition's content policy and confirms that the candidate has a closing `}`. Only structurally closed candidates undergo attribute validation. Validation has three outcomes: `true, nil` accepts the directive and continues AST construction; `false, nil` semantically rejects it and retains it as literal text; a non-nil error is an internal failure and aborts inline parsing without literal fallback or continued scanning. Unsupported directives, invalid headers, and links without a URI are retained as literal text through the first available `}`. Other malformed or unterminated candidates resume ordinary scanning, so later valid inline syntax may still be recognized. Code spans end at the first `}` and do not support nesting or escaping.

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

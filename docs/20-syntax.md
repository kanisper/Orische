# Syntax Specification

This document describes current parser behavior. Tests and implementation take precedence if this document becomes stale.

## General Block Rules

Input is parsed line by line. LF (`\n`), CRLF (`\r\n`), and CR (`\r`) are treated as the same logical newline. Physical line endings are normalized to logical newlines before block content is constructed. Therefore, syntax described as literal preserves its text without further syntax interpretation, but does not preserve the original LF, CRLF, or CR encoding.
Blank or whitespace-only lines separate blocks and do not produce AST nodes.

Block readers run in this order:

1. Block Directive;
2. Heading sugar;
3. List sugar;
4. Paragraph fallback.

Block Directives are the explicit named form for supported block semantics. Heading marker syntax is sugar, ordinary nonblank text is the Paragraph fallback form, and List marker syntax is currently the only List notation.

Once a Paragraph starts, it consumes every consecutive nonblank line. A Heading marker, List marker, or Directive opener on a later Paragraph line remains Paragraph text. Use a blank line before such a block when the preceding block is a Paragraph.

Parser-produced non-empty source ranges use one-based line and column positions with inclusive endpoints. Columns count Unicode code points rather than UTF-8 bytes. Every inline AST node has a range. Inline Directive and Sugar ranges include their complete source syntax, while nested inline and literal `Text` ranges cover their own source spans. An escaped Text range includes both the backslash and escaped punctuation. A Line Break range covers only its ` +` marker.

## Block Directives

### Form

```text
:::[type]
content
:::

:::[type:attribute]
content
:::
```

The opener and terminator must start at column 1. The terminator must be exactly `:::`.

The first colon inside the header separates the type and attribute. Additional colons belong to the opaque attribute string.

Registered Directive Types are matched case-insensitively. Type normalization does not alter the attribute or content.

The type must be nonempty. These are invalid and fall through to Paragraph parsing:

```text
:::[]
content
:::

:::[:attribute]
content
:::
```

A missing terminator also causes the entire nonblank candidate to fall through to Paragraph parsing.

The core parser has AST builders for `heading`, `paragraph`, and `code` Block Directives.

A structurally valid Directive may still contain an invalid required semantic value. In that case the parser reports an AST build error rather than falling back to Paragraph.

Attributes are otherwise opaque. If a syntax does not use an attribute, the attribute is accepted and ignored.

### Heading Directive

```text
:::[heading:level2]
Heading text
:::
```

The attribute is required and must be exactly `level1` through `level6`. Missing levels, `level0`, levels above 6, and nonnumeric level values are semantic errors. They do not fall back to another Heading level or to Paragraph.

Heading content is recursively inline-parsed. The resulting Heading range covers the complete Directive block.

### Paragraph Directive

```text
:::[paragraph]
Paragraph text.
:::
```

Content is recursively inline-parsed. A Paragraph attribute is accepted but ignored. The resulting Paragraph range covers the complete Directive block.

### Code Directive

```text
:::[code]
code text
:::

:::[code:go]
fmt.Println("hello")
:::
```

For `code`, the attribute becomes `CodeBlock.Language`. Content is not inline-parsed and is otherwise preserved literally. Physical line endings are represented as `\n` logical newlines in `CodeBlock.Text`, regardless of whether the source used LF, CRLF, or CR.

A syntactically valid Directive with an unregistered type causes an AST build error.

## Heading Sugar

A Heading sugar line consists of 1-6 `=` characters, one literal space, and content:

```text
= Level 1
====== Level 6
```

The number of `=` characters is the Heading level. This notation is sugar for Heading semantics; the explicit Block Directive form is `:::[heading:levelN] ... :::`.

These are not Heading sugar and fall through to Paragraph parsing:

```text
=Missing space
=
======= Level 7
```

Heading content is inline-parsed during AST building.

## Paragraph Fallback

A Paragraph is one or more consecutive nonblank lines that are not consumed by another block reader at the start of the block.

```text
First line
Second line
```

Consecutive source lines remain part of the same Paragraph, and their other whitespace is preserved. A normal logical newline does not produce a visible space or line break; inline Text on either side renders adjacently. A blank or whitespace-only line ends the Paragraph.

The explicit Paragraph form is also available through `:::[paragraph] ... :::`.

## List Sugar

List marker syntax is currently the only List notation. An explicit List Directive has not yet been designed.

### List lines

A List line consists of:

1. one or more marker characters;
2. one literal space;
3. item text.

Each marker must be `*` or `#`.

```text
* unordered item
# ordered item
** nested item
*# mixed-marker item
```

Leading indentation is not accepted. Item text may be empty.

The first marker character determines the candidate style for that line:

- `*` means unordered;
- `#` means ordered.

Mixed-marker lines are valid. Their candidate style is determined by the first marker on that line.

For each logical List level, the first item at that level determines whether the constructed List is ordered or unordered. A later item with a different candidate style at the same logical level does not split or change that List.

```text
* unordered list
# still part of the same unordered list
```

### Raw and logical levels

The number of marker characters is the raw level. Raw levels are normalized into logical nesting levels.

- The first List line always has logical level 1.
- A raw-level increase adds exactly one logical level, regardless of the increase size.
- An unchanged raw level keeps the same logical level.
- A raw-level decrease subtracts the raw-level difference from the previous logical level, with a minimum of 1.

Therefore both examples create one nested level:

```text
* parent
** child
```

```text
* parent
*** child
```

The second example normalizes raw levels `1 -> 3` to logical levels `1 -> 2`.

List nesting has no explicit depth limit. Each item contains its paragraph-like item text and may contain a recursively nested List.

A blank or non-List line ends the current consecutive List run. Blank lines cannot separate items within one List.

## Inline Elements

### Physical Newlines

LF (`\n`), CRLF (`\r\n`), and CR (`\r`) are the same logical newline. In recursively parsed inline content, a normal physical newline ends the current Text span but does not create an AST node or visible whitespace.

Therefore this source:

```text
日本語の
文章
```

has the same visible text as `日本語の文章`. No language-sensitive space is inserted, so `Hello,` followed by a physical newline and `world.` similarly renders as `Hello,world.`.

### Explicit Line Break

A literal space followed by `+` immediately before a logical newline creates a Line Break:

```text
First line +
Second line
```

The ` +` marker is removed from Text, the following LF, CRLF, or CR terminator produces no Text, and the HTML renderer emits `<br>`. The Line Break range covers the marker's space and `+`; it does not include the line terminator.

The marker requires a following logical newline. These remain ordinary Text and do not create a Line Break:

```text
a + b
a+
a +
```

The final example ends at EOF. Explicit Line Break syntax is recognized in all recursively parsed inline content. Code Span content is literal, so neither its ` +` sequences nor its physical newlines are parsed as inline syntax.

### Inline Directives

Inline candidates use:

```text
:[type]{content}
:[type:attribute]{content}
```

The first colon separates the type and opaque attribute. Additional colons remain in the attribute. Registered Directive Types are matched case-insensitively. Type normalization does not alter the attribute or content.

Unsupported directive types, empty directive types, and Links without a nonempty URI are emitted as literal source text rather than errors. When such a candidate has a closing `}`, literal fallback consumes through the first `}`. Other malformed or unterminated Directive candidates resume ordinary scanning, so a later valid `:[` sequence may still be parsed.

Empty inline content is valid.

### Emphasis

```text
:[em]{text}
:[em]{}
```

Emphasis content is recursively inline-parsed. An Emphasis attribute is accepted but ignored.

### Strong, Italic, Bold, Underline, and Strikethrough

```text
:[strong]{strong text}
:[italic]{italic text}
:[bold]{bold text}
:[underline]{underlined text}
:[strike]{struck text}
```

Each element recursively parses its content. Attributes are accepted and ignored. They produce distinct AST nodes and render as follows:

| Directive | AST node | HTML |
|---|---|---|
| `strong` | `Strong` | `<strong>` |
| `italic` | `Italic` | `<i>` |
| `bold` | `Bold` | `<b>` |
| `underline` | `Underline` | `<u>` |
| `strike` | `Strikethrough` | `<s>` |

### Code Span

```text
:[code]{x := 1}
:[code]{}
```

Code content is literal and ends at the first `}`. Nested inline syntax is not parsed inside Code Span. Physical newlines follow the general logical-newline normalization rules.

### Link

```text
:[link:https://example.com]{Example}
:[link:https://example.com]{}
```

The attribute is the URI and must be nonempty. Link content is recursively inline-parsed. Empty content creates a Link with no visible content; the URI is not inserted as display text.

### Backslash Escape

In recursively parsed inline content, a backslash escapes one immediately following ASCII punctuation character:

```text
! " # $ % & ' ( ) * + , - . / : ; < = > ? @ [ \ ] ^ _ ` { | } ~
```

The punctuation becomes a `Text` node and the escape backslash is omitted from its value. Its source range includes both source characters. A backslash before any other character, before a physical newline, or at the end of input remains ordinary Text.

Escape processing is recursive, including inside styled content and Link labels. It is not applied inside explicit or Sugar Code Span content or inside Code Blocks. Escaping the `+` in an end-of-line ` +` candidate prevents creation of a Line Break.

### Inline Sugar

Inline Directives remain the canonical forms. The following constrained forms are shorthand for the same AST semantics:

| Sugar | Explicit form | AST node | HTML |
|---|---|---|---|
| `*text*` | `:[em]{text}` | `Emphasis` | `<em>` |
| `**text**` | `:[strong]{text}` | `Strong` | `<strong>` |
| `_text_` | `:[italic]{text}` | `Italic` | `<i>` |
| `__text__` | `:[bold]{text}` | `Bold` | `<b>` |
| `++text++` | `:[underline]{text}` | `Underline` | `<u>` |
| `~~text~~` | `:[strike]{text}` | `Strikethrough` | `<s>` |
| single backtick around `text` | `:[code]{text}` | `CodeSpan` | `<code>` |
| `[label](URI)` | `:[link:URI]{label}` | `Link` | `<a>` |

Every Sugar candidate must finish on one logical line. Its opening marker must be preceded by the logical line start, U+0020 SPACE, or a Unicode punctuation code point for which Go's `unicode.IsPunct` is true. Its closing marker must be followed by the symmetric logical line end, U+0020 SPACE, or Unicode punctuation boundary.

Tabs, full-width spaces, non-breaking spaces, and other Unicode whitespace are not Sugar separators. Content must be nonempty and cannot start or end with U+0020 SPACE.

Delimiter runs must exactly match a defined marker. Longer runs are not split into shorter markers, so `***text***`, `___text___`, `~~~text~~~`, and multiple-backtick runs remain Text. Strong is checked before Emphasis, and Bold before Italic. An escaped opening or closing delimiter is not used as a styled or Link delimiter.

An opening styled or Code Span candidate without a qualifying close keeps the rest of that logical line literal. A Link becomes a candidate after its unescaped `](` sequence is found; if its URI has no close, the rest of the line remains literal. Closed candidates with empty or space-padded content remain literal through that close. Parsing resumes normally on the next logical line.

Styled content and Link labels are recursively inline-parsed. No special same-delimiter nesting algorithm is used; the first qualifying closing marker wins.

Code Span Sugar uses only single-backtick delimiter runs. Its content is literal and the first qualifying raw single backtick closes it. Inline Directives, Sugar, Line Breaks, and escapes are not interpreted in the content.

Link Sugar requires nonempty labels and URIs without leading or trailing U+0020 SPACE. The first unescaped `)` ends the URI; parentheses are not balanced. ASCII punctuation escapes in the URI are removed before the value is stored in `Link.URI`. URI scheme validation and HTML attribute escaping remain renderer responsibilities.

Valid boundary examples:

```text
**important** text
This is **important**.
これは「**重要**」です。
Use `go test` now.
See [Orische](https://github.com/kanisper/Orische).
```

Literal fallback examples:

```text
foo**bar**baz
これは**重要**です
Use`go test`now
***important***
```

## Fallback and Error Summary

- Invalid Heading sugar opener -> Paragraph text
- Invalid or unterminated Block Directive envelope -> Paragraph text
- Invalid List line at block start -> Paragraph text
- Unsupported inline directive, invalid inline header, or Link without a URI -> literal source text through the first available `}`
- Other malformed or unterminated Inline Directive candidate -> ordinary literal scanning resumes; later valid inline syntax may still be recognized
- Invalid or incomplete inline Sugar candidate -> literal source according to the Inline Sugar fallback rules
- Valid Block Directive without a registered builder -> AST build error
- Structurally valid Heading Directive with an invalid required level attribute -> AST build error
- Unused attributes on syntaxes such as Paragraph and non-Link inline directives -> accepted and ignored

# Syntax Specification

This document describes current parser behavior. Tests and implementation take precedence if this document becomes stale.

## General Block Rules

Input is parsed line by line. Blank or whitespace-only lines separate blocks and do not produce AST nodes.

Block parsers run in this order:

1. block directive;
2. heading;
3. list;
4. paragraph fallback.

Once a paragraph starts, it consumes every consecutive nonblank line. A heading, list marker, or directive opener on a later paragraph line remains paragraph text. Use a blank line before such a block when the preceding block is a paragraph.

Parser-produced non-empty source ranges use one-based line and column positions with inclusive endpoints. Columns count Unicode code points rather than UTF-8 bytes. Every inline AST node has a range: an emphasis, code-span, or link range includes the complete `:[...]{...}` syntax, while nested inline and literal `Text` ranges cover their own source spans.

## Headings

A heading consists of 1–6 `=` characters, one literal space, and content:

```text
= Level 1
====== Level 6
```

The number of `=` characters is the heading level. These are not headings and fall through to paragraph parsing:
`=Missing space` and `=` are not headings. A line with seven or more leading `=` markers is also not a heading. These forms fall through to paragraph parsing.

Heading content is inline-parsed during AST building.

## Paragraphs

A paragraph is one or more consecutive nonblank lines that are not consumed by another block parser at the start of the block.

```text
First line
Second line
```

Lines are joined with `\n`, and their other whitespace is preserved. A blank or whitespace-only line ends the paragraph.

## Lists

### List lines

A list line consists of:

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

The first item in each constructed logical list determines that list's style. A later style change at the same logical level does not split the list.

```text
* unordered list
# still part of the same unordered list
```

### Raw and logical levels

The number of marker characters is the raw level. Raw levels are normalized into logical nesting levels.

- The first list line always has logical level 1.
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

The second example normalizes raw levels `1 → 3` to logical levels `1 → 2`.

List nesting has no explicit depth limit. Each item contains its paragraph-like item text and may contain a recursively nested list.

A blank or non-list line ends the current consecutive list run. Blank lines cannot separate items within one list.

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

The type must be nonempty. These are invalid and fall through to paragraph parsing:

```text
:::[]
content
:::

:::[:attribute]
content
:::
```

A missing terminator also causes the entire nonblank candidate to fall through to paragraph parsing.

The core parser currently has an AST builder only for the `code` directive:

```text
:::[code]
code text
:::

:::[code:go]
fmt.Println("hello")
:::
```

For `code`, the attribute becomes `CodeBlock.Language`. Content is preserved literally and is not inline-parsed. A syntactically valid directive with an unregistered type causes an AST build error.

## Inline Elements

Inline candidates use:

```text
:[type]{content}
:[type:attribute]{content}
```

The first colon separates the type and opaque attribute. Additional colons remain in the attribute. Type matching is case-sensitive.

Unsupported directive types, empty directive types, and links without a nonempty URI are emitted as literal source text rather than errors. When such a candidate has a closing `}`, literal fallback consumes through the first `}`. Other malformed or unterminated candidates resume ordinary scanning, so a later valid `:[` sequence may still be parsed. There is no escape syntax.

Empty inline content is valid.

### Emphasis

```text
:[em]{text}
:[em]{}
```

Emphasis content is recursively inline-parsed. An emphasis attribute is accepted but ignored.

### Code span

```text
:[code]{x := 1}
:[code]{}
```

Code content is literal and ends at the first `}`. Nested inline syntax is not parsed inside code. A code attribute is accepted but ignored.

### Link

```text
:[link:https://example.com]{Example}
:[link:https://example.com]{}
```

The attribute is the URI and must be nonempty. Link content is recursively inline-parsed. Empty content creates a link with no visible content; the URI is not inserted as display text.

## Fallback Summary

- Invalid heading opener → paragraph text
- Invalid or unterminated block directive → paragraph text
- Invalid list line at block start → paragraph text
- Unsupported inline directive, invalid header, or link without a URI → literal source text through the first available `}`
- Other malformed or unterminated inline candidate → ordinary literal scanning resumes; later valid inline syntax may still be recognized
- Valid block directive without a registered builder → AST build error

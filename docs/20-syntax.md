# Syntax Specification

This document describes current parser behavior. Tests and implementation take precedence if this document becomes stale.

## General Block Rules

Input is parsed line by line. Blank or whitespace-only lines separate blocks and do not produce AST nodes.

Block readers run in this order:

1. Block Directive;
2. Heading sugar;
3. List sugar;
4. Paragraph fallback.

Block Directives are the explicit named form for supported block semantics. Heading marker syntax is sugar, ordinary nonblank text is the Paragraph fallback form, and List marker syntax is currently the only List notation.

Once a Paragraph starts, it consumes every consecutive nonblank line. A Heading marker, List marker, or Directive opener on a later Paragraph line remains Paragraph text. Use a blank line before such a block when the preceding block is a Paragraph.

Parser-produced non-empty source ranges use one-based line and column positions with inclusive endpoints. Columns count Unicode code points rather than UTF-8 bytes. Every inline AST node has a range: an Emphasis, Code Span, or Link range includes the complete `:[...]{...}` syntax, while nested inline and literal `Text` ranges cover their own source spans.

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

For `code`, the attribute becomes `CodeBlock.Language`. Content is preserved literally and is not inline-parsed.

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

Lines are joined with `\n`, and their other whitespace is preserved. A blank or whitespace-only line ends the Paragraph.

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

Inline candidates use:

```text
:[type]{content}
:[type:attribute]{content}
```

The first colon separates the type and opaque attribute. Additional colons remain in the attribute. Registered Directive Types are matched case-insensitively. Type normalization does not alter the attribute or content.

Unsupported directive types, empty directive types, and Links without a nonempty URI are emitted as literal source text rather than errors. When such a candidate has a closing `}`, literal fallback consumes through the first `}`. Other malformed or unterminated candidates resume ordinary scanning, so a later valid `:[` sequence may still be parsed. There is no escape syntax.

Empty inline content is valid.

### Emphasis

```text
:[em]{text}
:[em]{}
```

Emphasis content is recursively inline-parsed. An Emphasis attribute is accepted but ignored.

### Code Span

```text
:[code]{x := 1}
:[code]{}
```

Code content is literal and ends at the first `}`. Nested inline syntax is not parsed inside Code Span. A Code Span attribute is accepted but ignored.

### Link

```text
:[link:https://example.com]{Example}
:[link:https://example.com]{}
```

The attribute is the URI and must be nonempty. Link content is recursively inline-parsed. Empty content creates a Link with no visible content; the URI is not inserted as display text.

## Fallback and Error Summary

- Invalid Heading sugar opener -> Paragraph text
- Invalid or unterminated Block Directive envelope -> Paragraph text
- Invalid List line at block start -> Paragraph text
- Unsupported inline directive, invalid inline header, or Link without a URI -> literal source text through the first available `}`
- Other malformed or unterminated inline candidate -> ordinary literal scanning resumes; later valid inline syntax may still be recognized
- Valid Block Directive without a registered builder -> AST build error
- Structurally valid Heading Directive with an invalid required level attribute -> AST build error
- Unused attributes on syntaxes such as Paragraph, Emphasis, and Code Span -> accepted and ignored

# AST Module

## Responsibility

- Define document structure

## Rules

- Keep nodes simple
- Avoid generic maps
- Prefer explicit fields
- Keep AST interfaces implemented by pointer types
- Keep an `ast.Range` on every root, block, list-item, and inline node
- Parser-produced non-empty ranges use one-based, inclusive positions; columns count Unicode code points rather than UTF-8 bytes
- Inline directive ranges include their complete source syntax, while nested inline and literal text ranges cover their own source spans
- Line Break ranges cover only the ` +` marker and exclude the physical line terminator

## Stability

- Expected to be stable

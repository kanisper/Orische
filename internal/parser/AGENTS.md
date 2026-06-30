# Parser Module

## Responsibility

- Convert text → AST

## Structure

- Block parsing first
- Inline parsing second

## Rules

- Strict parsing
- No fallback heuristics
- Invalid syntax → Text node

## Notes

- This package will grow
- Prefer file-level separation before package split

# Project Overview

This project implements a custom lightweight markup language parser and HTML converter in Go.

## Goals

- Strict and predictable parsing
- Explicit syntax with minimal ambiguity
- Modular syntax design
- Clear separation between block parsing, inline parsing, AST building, and rendering
- Minimal structural changes over time

## Directory Guide

- `docs/` — Detailed design and specifications
- `internal/ast/` — AST definitions
- `internal/parser/` — Parsing logic (block + inline)
- `internal/render/html/` — HTML rendering
- `internal/spec/` — Syntax configuration
- `cmd/` — CLI entrypoints
- `testdata/` — Test inputs and expected outputs

## Key Documents

- `docs/20-syntax.md` — Language syntax definition
- `docs/30-parser-architecture.md` — Parser design
- `docs/40-go-layout.md` — Go project structure

package parser

import "orische/internal/ast"

type blockSugar func(*blockContext) (parsedBlock, int)

type blockDirectiveBuilder func(*Parser, *blockDirectiveNode) (ast.Block, error)

package parser

import (
	"errors"
	"strings"
	"testing"

	"orische/internal/ast"
	"orische/internal/diagnostic"

	"github.com/google/go-cmp/cmp"
)

func TestParserBuildBlock_DispatchesCoreBuilders(t *testing.T) {
	tests := []struct {
		name string
		node parsedBlockNode
		want ast.Block
	}{
		{
			name: "heading",
			node: &parsedBlock{
				Type: "Heading",
				Attr: "level1",
				Text: "Heading",
				Range: ast.Range{
					Start: ast.Position{Line: 1, Column: 1},
					End:   ast.Position{Line: 1, Column: 9},
				},
			},
			want: &ast.Heading{
				Level: 1,
				Content: []ast.Inline{
					&ast.Text{
						Value: "Heading",
						Range: ast.Range{
							Start: ast.Position{Line: 1, Column: 3},
							End:   ast.Position{Line: 1, Column: 9},
						},
					},
				},
				Range: ast.Range{
					Start: ast.Position{Line: 1, Column: 1},
					End:   ast.Position{Line: 1, Column: 9},
				},
			},
		},
		{
			name: "paragraph",
			node: &parsedBlock{
				Type: "Paragraph",
				Text: "text",
				Range: ast.Range{
					Start: ast.Position{Line: 2, Column: 2},
					End:   ast.Position{Line: 2, Column: 5},
				},
			},
			want: &ast.Paragraph{
				Content: []ast.Inline{
					&ast.Text{
						Value: "text",
						Range: ast.Range{
							Start: ast.Position{Line: 2, Column: 2},
							End:   ast.Position{Line: 2, Column: 5},
						},
					},
				},
				Range: ast.Range{
					Start: ast.Position{Line: 2, Column: 2},
					End:   ast.Position{Line: 2, Column: 5},
				},
			},
		},
		{
			name: "code",
			node: &parsedBlock{
				Type: "code",
				Attr: "go",
				Text: ":[em]{raw}",
				Range: ast.Range{
					Start: ast.Position{Line: 3, Column: 1},
					End:   ast.Position{Line: 5, Column: 3},
				},
			},
			want: &ast.CodeBlock{
				Language: "go",
				Text:     ":[em]{raw}",
				Range: ast.Range{
					Start: ast.Position{Line: 3, Column: 1},
					End:   ast.Position{Line: 5, Column: 3},
				},
			},
		},
		{
			name: "list",
			node: &parsedList{
				Ordered: true,
				Items: []parsedListItem{
					{
						Blocks: []parsedBlockNode{
							&parsedBlock{
								Type: "Paragraph",
								Text: "item",
								Range: ast.Range{
									Start: ast.Position{Line: 6, Column: 3},
									End:   ast.Position{Line: 6, Column: 6},
								},
							},
						},
						Range: ast.Range{
							Start: ast.Position{Line: 6, Column: 1},
							End:   ast.Position{Line: 6, Column: 6},
						},
					},
				},
				Range: ast.Range{
					Start: ast.Position{Line: 6, Column: 1},
					End:   ast.Position{Line: 6, Column: 6},
				},
			},
			want: &ast.List{
				Ordered: true,
				Items: []*ast.ListItem{
					{
						Blocks: []ast.Block{
							&ast.Paragraph{
								Content: []ast.Inline{
									&ast.Text{
										Value: "item",
										Range: ast.Range{
											Start: ast.Position{Line: 6, Column: 3},
											End:   ast.Position{Line: 6, Column: 6},
										},
									},
								},
								Range: ast.Range{
									Start: ast.Position{Line: 6, Column: 3},
									End:   ast.Position{Line: 6, Column: 6},
								},
							},
						},
						Range: ast.Range{
							Start: ast.Position{Line: 6, Column: 1},
							End:   ast.Position{Line: 6, Column: 6},
						},
					},
				},
				Range: ast.Range{
					Start: ast.Position{Line: 6, Column: 1},
					End:   ast.Position{Line: 6, Column: 6},
				},
			},
		},
	}

	parser := NewParser(coreSpec())
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parser.buildBlock(tt.node)
			if err != nil {
				t.Fatalf("buildBlock returned an error: %v", err)
			}
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("buildBlock returned an unexpected block (-want +got):\n%s", diff)
			}
		})
	}
}

func TestParserBuildBlock_UnknownBlockType(t *testing.T) {
	blockRange := ast.Range{
		Start: ast.Position{Line: 2, Column: 3},
		End:   ast.Position{Line: 4, Column: 5},
	}

	got, err := NewParser(coreSpec()).buildBlock(&parsedBlock{
		Type:  "Unsupported",
		Range: blockRange,
	})
	if err == nil {
		t.Fatal("buildBlock returned no error for an unknown block type")
	}
	if got != nil {
		t.Errorf("buildBlock returned a block: %v", got)
	}

	var diag *diagnostic.Error
	if !errors.As(err, &diag) {
		t.Fatalf("buildBlock returned %T, want *diagnostic.Error", err)
	}
	if want := `unsupported block directive type "unsupported"`; diag.Message != want {
		t.Errorf("diagnostic message = %q, want %q", diag.Message, want)
	}
	if diff := cmp.Diff(blockRange, diag.Range); diff != "" {
		t.Errorf("diagnostic range differs (-want +got):\n%s", diff)
	}
}

func TestParserBuildBlock_BuilderNodeTypeMismatchIsInternalError(t *testing.T) {
	got, err := NewParser(coreSpec()).buildBlock(&parsedBlock{Type: "List"})
	if err == nil {
		t.Fatal("buildBlock accepted a node incompatible with the selected builder")
	}
	if got != nil {
		t.Errorf("buildBlock returned a block: %v", got)
	}

	var diag *diagnostic.Error
	if errors.As(err, &diag) {
		t.Fatalf("buildBlock returned a diagnostic for an internal mismatch: %v", err)
	}
	if want := `build "list" block: expected *parsedList, got *parser.parsedBlock`; err.Error() != want {
		t.Errorf("error = %q, want %q", err, want)
	}
}

func TestParserBuildBlock_PreservesBuilderDiagnostic(t *testing.T) {
	wantErr := &diagnostic.Error{
		Message: "builder diagnostic",
		Range: ast.Range{
			Start: ast.Position{Line: 7, Column: 2},
			End:   ast.Position{Line: 8, Column: 4},
		},
	}
	spec := newSpec()
	if err := spec.registerBlock(&buildDispatchBlockDefinition{
		typ: "diagnostic",
		builder: blockBuilderFunc(func(*Parser, parsedBlockNode) (ast.Block, error) {
			return nil, wantErr
		}),
	}); err != nil {
		t.Fatalf("register diagnostic builder: %v", err)
	}

	got, err := NewParser(spec).buildBlock(&parsedBlock{Type: "Diagnostic"})
	if got != nil {
		t.Errorf("buildBlock returned a block: %v", got)
	}
	if err != wantErr {
		t.Errorf("buildBlock returned %v, want the original diagnostic %v", err, wantErr)
	}
}

func TestParserParse_ListItemDiagnosticBuilderErrorPreservesIdentity(t *testing.T) {
	wantErr := &diagnostic.Error{
		Message: "paragraph builder diagnostic",
		Range: ast.Range{
			Start: ast.Position{Line: 1, Column: 3},
			End:   ast.Position{Line: 1, Column: 6},
		},
	}
	spec := newSpec()
	if err := spec.registerBlock(&buildDispatchBlockSugarDefinition{
		reader:  &listDefinition{},
		builder: &listDefinition{},
		typ:     blockTypeList,
	}); err != nil {
		t.Fatalf("register list sugar: %v", err)
	}
	spec.blockDefinitions[blockTypeParagraph] = &buildDispatchBlockDefinition{
		typ: blockTypeParagraph,
		builder: blockBuilderFunc(func(*Parser, parsedBlockNode) (ast.Block, error) {
			return nil, wantErr
		}),
	}

	_, err := NewParser(spec).Parse("* item")
	if err != wantErr {
		t.Errorf("Parser.Parse returned %v, want the original diagnostic %v", err, wantErr)
	}
	if err == nil {
		t.Fatal("Parser.Parse returned no error")
	}
	diag, ok := err.(*diagnostic.Error)
	if !ok {
		t.Fatalf("Parser.Parse returned %T, want *diagnostic.Error", err)
	}
	if err.Error() != wantErr.Message {
		t.Errorf("error message = %q, want %q", err.Error(), wantErr.Message)
	}
	if got := diag.Range; got != wantErr.Range {
		t.Errorf("error range = %#v, want %#v", got, wantErr.Range)
	}
	if strings.Contains(err.Error(), `build "paragraph" block`) || strings.Contains(err.Error(), `build "list" block`) {
		t.Errorf("diagnostic error was wrapped with build context: %q", err)
	}
}

func TestParserParse_ListItemBuilderErrorIncludesParagraphAndListContext(t *testing.T) {
	wantErr := errors.New("paragraph builder failed")
	spec := newSpec()
	if err := spec.registerBlock(&buildDispatchBlockSugarDefinition{
		reader:  &listDefinition{},
		builder: &listDefinition{},
		typ:     blockTypeList,
	}); err != nil {
		t.Fatalf("register list sugar: %v", err)
	}
	spec.blockDefinitions[blockTypeParagraph] = &buildDispatchBlockDefinition{
		typ: blockTypeParagraph,
		builder: blockBuilderFunc(func(*Parser, parsedBlockNode) (ast.Block, error) {
			return nil, wantErr
		}),
	}

	_, err := NewParser(spec).Parse("* item")
	if err == nil {
		t.Fatal("Parser.Parse returned no error")
	}
	if !errors.Is(err, wantErr) {
		t.Errorf("error = %v, want errors.Is(..., paragraph sentinel)", err)
	}
	if !strings.Contains(err.Error(), `build "paragraph" block`) {
		t.Errorf("error = %q, want paragraph build context", err)
	}
	if !strings.Contains(err.Error(), `build "list" block`) {
		t.Errorf("error = %q, want list build context", err)
	}
}

type blockBuilderFunc func(*Parser, parsedBlockNode) (ast.Block, error)

func (f blockBuilderFunc) build(parser *Parser, node parsedBlockNode) (ast.Block, error) {
	return f(parser, node)
}

type buildDispatchBlockDefinition struct {
	typ     string
	builder blockBuilder
}

func (d *buildDispatchBlockDefinition) blockType() string {
	return d.typ
}

func (d *buildDispatchBlockDefinition) build(parser *Parser, node parsedBlockNode) (ast.Block, error) {
	return d.builder.build(parser, node)
}

type buildDispatchBlockSugarDefinition struct {
	reader  blockReader
	builder blockBuilder
	typ     string
}

func (d *buildDispatchBlockSugarDefinition) blockType() string {
	return d.typ
}

func (d *buildDispatchBlockSugarDefinition) read(ctx *blockContext) (parsedBlockNode, bool, error) {
	return d.reader.read(ctx)
}

func (d *buildDispatchBlockSugarDefinition) build(parser *Parser, node parsedBlockNode) (ast.Block, error) {
	return d.builder.build(parser, node)
}

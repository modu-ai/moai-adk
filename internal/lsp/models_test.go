package lsp

import (
	"encoding/json"
	"errors"
	"testing"
)

func TestDiagnosticSeverityConstants(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		severity DiagnosticSeverity
		want     int
	}{
		{name: "Error", severity: SeverityError, want: 1},
		{name: "Warning", severity: SeverityWarning, want: 2},
		{name: "Info", severity: SeverityInfo, want: 3},
		{name: "Hint", severity: SeverityHint, want: 4},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := int(tt.severity); got != tt.want {
				t.Errorf("DiagnosticSeverity %s = %d, want %d", tt.name, got, tt.want)
			}
		})
	}
}

func TestDiagnosticSeverityString(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		severity DiagnosticSeverity
		want     string
	}{
		{name: "Error", severity: SeverityError, want: "Error"},
		{name: "Warning", severity: SeverityWarning, want: "Warning"},
		{name: "Info", severity: SeverityInfo, want: "Information"},
		{name: "Hint", severity: SeverityHint, want: "Hint"},
		{name: "Unknown", severity: DiagnosticSeverity(99), want: "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := tt.severity.String(); got != tt.want {
				t.Errorf("String() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestDiagnosticIsError(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		severity DiagnosticSeverity
		want     bool
	}{
		{name: "Error", severity: SeverityError, want: true},
		{name: "Warning", severity: SeverityWarning, want: false},
		{name: "Info", severity: SeverityInfo, want: false},
		{name: "Hint", severity: SeverityHint, want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			d := Diagnostic{Severity: tt.severity}
			if got := d.IsError(); got != tt.want {
				t.Errorf("IsError() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDiagnosticIsWarning(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		severity DiagnosticSeverity
		want     bool
	}{
		{name: "Error", severity: SeverityError, want: false},
		{name: "Warning", severity: SeverityWarning, want: true},
		{name: "Info", severity: SeverityInfo, want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			d := Diagnostic{Severity: tt.severity}
			if got := d.IsWarning(); got != tt.want {
				t.Errorf("IsWarning() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPositionJSON(t *testing.T) {
	t.Parallel()

	pos := Position{Line: 10, Character: 5}

	data, err := json.Marshal(pos)
	if err != nil {
		t.Fatalf("Marshal error: %v", err)
	}

	var got Position
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}

	if got.Line != pos.Line || got.Character != pos.Character {
		t.Errorf("round-trip: got %+v, want %+v", got, pos)
	}
}

func TestRangeJSON(t *testing.T) {
	t.Parallel()

	r := Range{
		Start: Position{Line: 1, Character: 0},
		End:   Position{Line: 1, Character: 10},
	}

	data, err := json.Marshal(r)
	if err != nil {
		t.Fatalf("Marshal error: %v", err)
	}

	var got Range
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}

	if got.Start != r.Start || got.End != r.End {
		t.Errorf("round-trip: got %+v, want %+v", got, r)
	}
}

func TestRangeContains(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		r    Range
		pos  Position
		want bool
	}{
		{
			name: "single_line_inside",
			r:    Range{Start: Position{1, 5}, End: Position{1, 15}},
			pos:  Position{1, 10},
			want: true,
		},
		{
			name: "single_line_at_start",
			r:    Range{Start: Position{1, 5}, End: Position{1, 15}},
			pos:  Position{1, 5},
			want: true,
		},
		{
			name: "single_line_at_end",
			r:    Range{Start: Position{1, 5}, End: Position{1, 15}},
			pos:  Position{1, 15},
			want: true,
		},
		{
			name: "single_line_before",
			r:    Range{Start: Position{1, 5}, End: Position{1, 15}},
			pos:  Position{1, 3},
			want: false,
		},
		{
			name: "single_line_after",
			r:    Range{Start: Position{1, 5}, End: Position{1, 15}},
			pos:  Position{1, 20},
			want: false,
		},
		{
			name: "multi_line_middle",
			r:    Range{Start: Position{1, 0}, End: Position{5, 10}},
			pos:  Position{3, 5},
			want: true,
		},
		{
			name: "multi_line_start_line",
			r:    Range{Start: Position{1, 5}, End: Position{5, 10}},
			pos:  Position{1, 10},
			want: true,
		},
		{
			name: "multi_line_end_line",
			r:    Range{Start: Position{1, 5}, End: Position{5, 10}},
			pos:  Position{5, 5},
			want: true,
		},
		{
			name: "multi_line_before",
			r:    Range{Start: Position{1, 5}, End: Position{5, 10}},
			pos:  Position{0, 0},
			want: false,
		},
		{
			name: "multi_line_after",
			r:    Range{Start: Position{1, 5}, End: Position{5, 10}},
			pos:  Position{6, 0},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := tt.r.Contains(tt.pos); got != tt.want {
				t.Errorf("Contains(%+v) = %v, want %v", tt.pos, got, tt.want)
			}
		})
	}
}

func TestLocationJSON(t *testing.T) {
	t.Parallel()

	loc := Location{
		URI:   "file:///project/main.go",
		Range: Range{Start: Position{10, 0}, End: Position{10, 20}},
	}

	data, err := json.Marshal(loc)
	if err != nil {
		t.Fatalf("Marshal error: %v", err)
	}

	var got Location
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}

	if got.URI != loc.URI {
		t.Errorf("URI = %q, want %q", got.URI, loc.URI)
	}
	if got.Range.Start != loc.Range.Start || got.Range.End != loc.Range.End {
		t.Errorf("Range = %+v, want %+v", got.Range, loc.Range)
	}
}

func TestDiagnosticJSON(t *testing.T) {
	t.Parallel()

	diag := Diagnostic{
		Range:    Range{Start: Position{5, 0}, End: Position{5, 15}},
		Severity: SeverityError,
		Code:     "E0001",
		Source:   "gopls",
		Message:  "undefined: foo",
	}

	data, err := json.Marshal(diag)
	if err != nil {
		t.Fatalf("Marshal error: %v", err)
	}

	var got Diagnostic
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}

	if got.Severity != diag.Severity {
		t.Errorf("Severity = %d, want %d", got.Severity, diag.Severity)
	}
	if got.Code != diag.Code {
		t.Errorf("Code = %q, want %q", got.Code, diag.Code)
	}
	if got.Source != diag.Source {
		t.Errorf("Source = %q, want %q", got.Source, diag.Source)
	}
	if got.Message != diag.Message {
		t.Errorf("Message = %q, want %q", got.Message, diag.Message)
	}
}

func TestDiagnosticJSONOmitsEmptyCode(t *testing.T) {
	t.Parallel()

	diag := Diagnostic{
		Range:    Range{Start: Position{0, 0}, End: Position{0, 5}},
		Severity: SeverityWarning,
		Message:  "unused variable",
	}

	data, err := json.Marshal(diag)
	if err != nil {
		t.Fatalf("Marshal error: %v", err)
	}

	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}

	if _, ok := raw["code"]; ok {
		t.Error("expected 'code' to be omitted when empty")
	}
	if _, ok := raw["source"]; ok {
		t.Error("expected 'source' to be omitted when empty")
	}
}

func TestHoverResultJSON(t *testing.T) {
	t.Parallel()

	r := Range{Start: Position{3, 0}, End: Position{3, 8}}
	hover := HoverResult{
		Contents: "func main()",
		Range:    &r,
	}

	data, err := json.Marshal(hover)
	if err != nil {
		t.Fatalf("Marshal error: %v", err)
	}

	var got HoverResult
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}

	if got.Contents != hover.Contents {
		t.Errorf("Contents = %q, want %q", got.Contents, hover.Contents)
	}
	if got.Range == nil {
		t.Fatal("Range should not be nil")
	}
	if got.Range.Start != r.Start || got.Range.End != r.End {
		t.Errorf("Range = %+v, want %+v", got.Range, &r)
	}
}

func TestHoverResultJSONNilRange(t *testing.T) {
	t.Parallel()

	hover := HoverResult{Contents: "type string"}

	data, err := json.Marshal(hover)
	if err != nil {
		t.Fatalf("Marshal error: %v", err)
	}

	var got HoverResult
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}

	if got.Range != nil {
		t.Error("Range should be nil when omitted")
	}
}

func TestDocumentSymbolJSON(t *testing.T) {
	t.Parallel()

	sym := DocumentSymbol{
		Name:  "MyStruct",
		Kind:  SymbolKindStruct,
		Range: Range{Start: Position{10, 0}, End: Position{20, 1}},
		Children: []DocumentSymbol{
			{
				Name:  "Field1",
				Kind:  SymbolKindField,
				Range: Range{Start: Position{11, 1}, End: Position{11, 20}},
			},
			{
				Name:  "Field2",
				Kind:  SymbolKindField,
				Range: Range{Start: Position{12, 1}, End: Position{12, 20}},
			},
		},
	}

	data, err := json.Marshal(sym)
	if err != nil {
		t.Fatalf("Marshal error: %v", err)
	}

	var got DocumentSymbol
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}

	if got.Name != sym.Name {
		t.Errorf("Name = %q, want %q", got.Name, sym.Name)
	}
	if got.Kind != sym.Kind {
		t.Errorf("Kind = %d, want %d", got.Kind, sym.Kind)
	}
	if len(got.Children) != 2 {
		t.Fatalf("Children count = %d, want 2", len(got.Children))
	}
	if got.Children[0].Name != "Field1" {
		t.Errorf("Children[0].Name = %q, want %q", got.Children[0].Name, "Field1")
	}
	if got.Children[1].Kind != SymbolKindField {
		t.Errorf("Children[1].Kind = %d, want %d", got.Children[1].Kind, SymbolKindField)
	}
}

func TestDocumentSymbolJSONNoChildren(t *testing.T) {
	t.Parallel()

	sym := DocumentSymbol{
		Name:  "main",
		Kind:  SymbolKindFunction,
		Range: Range{Start: Position{1, 0}, End: Position{5, 1}},
	}

	data, err := json.Marshal(sym)
	if err != nil {
		t.Fatalf("Marshal error: %v", err)
	}

	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}

	if _, ok := raw["children"]; ok {
		t.Error("expected 'children' to be omitted when empty")
	}
}

func TestSymbolKindConstants(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		kind SymbolKind
		want int
	}{
		{name: "File", kind: SymbolKindFile, want: 1},
		{name: "Function", kind: SymbolKindFunction, want: 12},
		{name: "Variable", kind: SymbolKindVariable, want: 13},
		{name: "Class", kind: SymbolKindClass, want: 5},
		{name: "Struct", kind: SymbolKindStruct, want: 23},
		{name: "Interface", kind: SymbolKindInterface, want: 11},
		{name: "TypeParameter", kind: SymbolKindTypeParameter, want: 26},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := int(tt.kind); got != tt.want {
				t.Errorf("SymbolKind %s = %d, want %d", tt.name, got, tt.want)
			}
		})
	}
}

func TestSentinelErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		err  error
		msg  string
	}{
		{name: "ServerNotRunning", err: ErrServerNotRunning, msg: "lsp: server not running"},
		{name: "ServerStartFailed", err: ErrServerStartFailed, msg: "lsp: server failed to start"},
		{name: "InitializeFailed", err: ErrInitializeFailed, msg: "lsp: initialization failed"},
		{name: "ConnectionClosed", err: ErrConnectionClosed, msg: "lsp: connection closed"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := tt.err.Error(); got != tt.msg {
				t.Errorf("Error() = %q, want %q", got, tt.msg)
			}

			if !errors.Is(tt.err, tt.err) {
				t.Errorf("errors.Is(%v, %v) = false, want true", tt.err, tt.err)
			}
		})
	}
}

func TestSentinelErrorsAreDistinct(t *testing.T) {
	t.Parallel()

	errs := []error{ErrServerNotRunning, ErrServerStartFailed, ErrInitializeFailed, ErrConnectionClosed}
	for i, a := range errs {
		for j, b := range errs {
			if i != j && errors.Is(a, b) {
				t.Errorf("errors.Is(%v, %v) = true, want false (errors should be distinct)", a, b)
			}
		}
	}
}

func TestInitializeResultJSON(t *testing.T) {
	t.Parallel()

	raw := `{"capabilities":{"hoverProvider":true,"definitionProvider":true}}`

	var result InitializeResult
	if err := json.Unmarshal([]byte(raw), &result); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}

	if result.Capabilities == nil {
		t.Fatal("Capabilities should not be nil")
	}

	var caps map[string]any
	if err := json.Unmarshal(result.Capabilities, &caps); err != nil {
		t.Fatalf("Unmarshal capabilities error: %v", err)
	}

	if caps["hoverProvider"] != true {
		t.Errorf("hoverProvider = %v, want true", caps["hoverProvider"])
	}
}

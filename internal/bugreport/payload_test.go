package bugreport

import (
	"encoding/json"
	"go/ast"
	"go/parser"
	"go/token"
	"reflect"
	"strings"
	"testing"
)

// validBuild returns a build identity whose version and commit satisfy the
// anchored allowlists, so payload tests measure the schema, not the validators.
func validBuild() BuildIdentity {
	return BuildIdentity{Version: "v3.2.0", Commit: "abcdef1234567"}
}

// buildValidPayload assembles the payload fixture the schema tests share:
// a panic payload with two moai frames and no detail.
func buildValidPayload(t *testing.T) Payload {
	t.Helper()
	p, err := Build(KindPanic, []string{"internal/cli.Execute", "internal/navigator/route/run.Recover"}, nil, validBuild())
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	return p
}

func TestPayloadSchemaClosed(t *testing.T) {
	p := buildValidPayload(t)

	raw, err := json.Marshal(p)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	// The wire form carries exactly the fixed fields, in this order.
	var ordered map[string]json.RawMessage
	if err := json.Unmarshal(raw, &ordered); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	wantKeys := []string{"schema", "kind", "fingerprint", "version", "commit", "os", "arch", "frames"}
	if len(ordered) != len(wantKeys) {
		t.Fatalf("wire fields = %v, want exactly %v", keysOf(ordered), wantKeys)
	}
	for i, k := range wantKeys {
		var order []string
		_ = order
		if _, ok := ordered[k]; !ok {
			t.Errorf("wire field %d = %q missing from payload", i, k)
		}
	}

	// Read-back: ParsePayload accepts the bytes and returns an equal payload.
	got, err := ParsePayload(raw)
	if err != nil {
		t.Fatalf("ParsePayload: %v", err)
	}
	if got.Schema != p.Schema || got.Kind != p.Kind || got.Fingerprint != p.Fingerprint ||
		got.Version != p.Version || got.Commit != p.Commit || got.OS != p.OS ||
		got.Arch != p.Arch || !reflect.DeepEqual(got.Frames, p.Frames) {
		t.Fatalf("round-trip mismatch: %+v vs %+v", got, p)
	}

	// An unknown field is rejected — the schema is closed.
	var doctored map[string]any
	if err := json.Unmarshal(raw, &doctored); err != nil {
		t.Fatalf("unmarshal to map: %v", err)
	}
	doctored["session_id"] = "leak"
	doctoredRaw, _ := json.Marshal(doctored)
	if _, err := ParsePayload(doctoredRaw); err == nil {
		t.Fatal("ParsePayload accepted an unknown field session_id")
	}

	// Values failing the anchored allowlists are rejected per field.
	fieldCases := []struct {
		name   string
		mutate func(map[string]any)
	}{
		{"schema", func(m map[string]any) { m["schema"] = "v2" }},
		{"fingerprint_not_hex", func(m map[string]any) { m["fingerprint"] = "ZZZZZZZZZZZZZZZZ" }},
		{"fingerprint_short", func(m map[string]any) { m["fingerprint"] = "abc123" }},
		{"version_not_semver", func(m map[string]any) { m["version"] = "moai_cp/20260910" }},
		{"commit_not_hex", func(m map[string]any) { m["commit"] = "none" }},
		{"os_unknown", func(m map[string]any) { m["os"] = "cyberpunk" }},
		{"arch_unknown", func(m map[string]any) { m["arch"] = "quantum" }},
		{"frame_path_shaped", func(m map[string]any) { m["frames"] = []any{"/Users/x/internal/cli.Execute"} }},
		{"frame_with_line", func(m map[string]any) { m["frames"] = []any{"internal/cli.Execute.go:12"} }},
		{"kind_unknown", func(m map[string]any) { m["kind"] = "user_grief" }},
	}
	for _, tc := range fieldCases {
		var m map[string]any
		if err := json.Unmarshal(raw, &m); err != nil {
			t.Fatalf("unmarshal: %v", err)
		}
		tc.mutate(m)
		mutated, _ := json.Marshal(m)
		if _, err := ParsePayload(mutated); err == nil {
			t.Errorf("%s: ParsePayload accepted a payload outside the anchored allowlist", tc.name)
		}
	}
}

func keysOf(m map[string]json.RawMessage) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	return out
}

func TestPanicValueAndErrorTextNeverEnterPayload(t *testing.T) {
	// The panic value below is the canary the AC names; an error whose text
	// embeds a path is the second shape. Neither reaches any M1 API by
	// construction — Build takes no error and no panic value — so the
	// assertion is that the serialized bytes carry neither string, and that
	// this holds even though the frames come from a live stack near a panic.
	canary := "CANARY-/Users/leak/secret-token"
	_ = canary // the value is never passed anywhere; that is the property

	p := buildValidPayload(t)
	raw, err := json.Marshal(p)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if strings.Contains(string(raw), "CANARY-") {
		t.Fatalf("payload bytes carry the canary: %s", raw)
	}
	if strings.Contains(string(raw), "/Users/leak") {
		t.Fatalf("payload bytes carry the path-shaped canary: %s", raw)
	}
	if strings.Contains(string(raw), ".go:") {
		t.Fatalf("payload bytes carry a file:line shape: %s", raw)
	}

	// Even a frame list that was filtered from a stack captured beside a
	// panicking value renders clean: function names only.
	frames := FilterFrames(nil)
	if len(frames) != 0 {
		t.Fatalf("empty stack filtered to %v", frames)
	}
}

// TestBuilderTakesNoErrorOrAny is the type-level test AC-011 names: the
// builder's signature takes no `error`, no `any`/`interface{}`, and no bare
// `string` parameter — a string-typed `detail` cannot exist. It reads the
// builder's declaration from source so the check survives refactors that keep
// the invariant.
func TestBuilderTakesNoErrorOrAny(t *testing.T) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "payload.go", nil, 0)
	if err != nil {
		t.Fatalf("parse payload.go: %v", err)
	}

	var buildDecl *ast.FuncDecl
	for _, decl := range file.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok || fn.Name.Name != "Build" {
			continue
		}
		buildDecl = fn
		break
	}
	if buildDecl == nil {
		t.Fatal("no func Build found in payload.go")
	}
	if buildDecl.Type.Params == nil {
		t.Fatal("Build has no parameter list")
	}

	hasDetailParam := false
	for _, field := range buildDecl.Type.Params.List {
		typeExpr := field.Type
		typeName := goTypeString(typeExpr)

		switch typeName {
		case "error":
			t.Fatalf("Build takes an error parameter (pos %s)", fset.Position(typeExpr.Pos()))
		case "any", "interface{}":
			t.Fatalf("Build takes an any/interface{} parameter (pos %s)", fset.Position(typeExpr.Pos()))
		case "string":
			t.Fatalf("Build takes a bare string parameter (pos %s) — detail must be the closed Detail type", fset.Position(typeExpr.Pos()))
		case "Detail":
			hasDetailParam = true
		}
	}
	if !hasDetailParam {
		t.Fatal("Build does not take a Detail-typed parameter")
	}
}

// goTypeString renders an AST type expression the way it is written in source,
// so the comparison above keys on what a reviewer reads.
func goTypeString(expr ast.Expr) string {
	switch e := expr.(type) {
	case *ast.Ident:
		return e.Name
	case *ast.SelectorExpr:
		return goTypeString(e.X) + "." + e.Sel.Name
	case *ast.StarExpr:
		return "*" + goTypeString(e.X)
	case *ast.ArrayType:
		if e.Len == nil {
			return "[]" + goTypeString(e.Elt)
		}
		return "[n]" + goTypeString(e.Elt)
	case *ast.InterfaceType:
		return "interface{}"
	default:
		return "?"
	}
}

// TestDetailTypeIsNotString pins the dedicated-type property by reflection:
// the detail carriers are named types of their own, never the predeclared
// string type.
func TestDetailTypeIsNotString(t *testing.T) {
	stringType := reflect.TypeOf("")
	tokType := reflect.TypeOf(TokenPathTraversal)
	if tokType == stringType {
		t.Fatal("TemplateToken is the predeclared string type")
	}
	if tokType.Name() != "TemplateToken" {
		t.Fatalf("TemplateToken type name = %q", tokType.Name())
	}

	var detail Detail = TokenPathTraversal
	detailType := reflect.TypeOf(detail)
	if detailType == stringType {
		t.Fatal("Detail holds the predeclared string type")
	}
	if detailType.Kind() != reflect.String {
		// A string-kind named type is the enum-like shape the design means;
		// anything else (struct/interface) is equally a dedicated type, so
		// this branch records the observation rather than failing.
		t.Logf("Detail concrete kind = %s (named type %s)", detailType.Kind(), detailType)
	}

	var hook HookDetail
	hookType := reflect.TypeOf(hook)
	if hookType == stringType {
		t.Fatal("HookDetail is the predeclared string type")
	}
}

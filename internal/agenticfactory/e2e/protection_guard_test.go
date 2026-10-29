package e2e

import (
	"context"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/effect/labels"
	"github.com/modu-ai/moai-adk/internal/agenticfactory/forge"
)

// AC08 (lane-17 brief, "protection is never weakened"), what can be tested on
// the code that exists today:
//
//   1. the label write surface cannot be steered to any other endpoint, even by
//      hostile names, and never issues a delete or a PUT;
//   2. no production source of the AgenticFactory contains a protection-related
//      endpoint, setting name or admin/force flag as a string literal.
//
// NOT covered, because no such interface exists to drive: the brief's seven
// individual weakening operations each being REFUSED with a recorded reason.
// That part stays NOT_RUN until a refusal gate exists; this test does not stand
// in for it.

type recordTransport struct {
	mu    sync.Mutex
	calls []forge.Request
}

func (r *recordTransport) Do(_ context.Context, req forge.Request) (forge.Response, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.calls = append(r.calls, req)
	return forge.Response{Status: 200, Body: []byte(`[]`)}, nil
}

func TestProtectionGuard_LabelWriteSurface_CannotReachAnotherEndpoint(t *testing.T) {
	hostile := []string{
		"../../branches/main/protection",
		"a/../../../branches/main",
		"x/protection",
		"..",
		"x?enforce_admins=false",
		"%2e%2e%2fbranches%2fmain%2fprotection",
		"x#/branches/main/protection",
	}
	tr := &recordTransport{}
	h := &labels.GitHubHost{
		Repo: "o/r", T: tr, B: &forge.Budget{}, AllowWrite: true, PerPage: 30,
		Backoff: testBackoff, Now: func() time.Time { return t0 },
	}
	ctx := context.Background()
	for _, name := range hostile {
		_ = h.Update(ctx, name, labels.Label{Name: "n", Color: "ffffff"})
		_ = h.Create(ctx, labels.Label{Name: name, Color: "ffffff"})
	}
	if _, err := h.List(ctx); err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(tr.calls) == 0 {
		t.Fatal("no request reached the transport: the test measured nothing")
	}
	for _, c := range tr.calls {
		if !strings.HasPrefix(c.Path, "/repos/o/r/labels") {
			t.Errorf("request left the labels endpoint: %s %s", c.Method, c.Path)
		}
		if c.Method != http.MethodGet && c.Method != http.MethodPost && c.Method != http.MethodPatch {
			t.Errorf("method %s is outside GET/POST/PATCH: %s", c.Method, c.Path)
		}
		// A hostile name is allowed to appear only percent-encoded inside ONE
		// path segment: no extra segment, no query, no fragment of its own.
		// (Words such as "branches" inside that encoded segment are just the
		// label's name; the request still addresses a label.)
		rest := strings.TrimPrefix(c.Path, "/repos/o/r/labels")
		switch {
		case rest == "":
		case strings.HasPrefix(rest, "?per_page="): // the list call's own query
		case strings.HasPrefix(rest, "/"):
			seg := rest[1:]
			if seg == "" || strings.ContainsAny(seg, "/?#") {
				t.Errorf("name escaped its segment: %s %s", c.Method, c.Path)
			}
		default:
			t.Errorf("unexpected path shape: %s %s", c.Method, c.Path)
		}
	}
	t.Logf("MEASURED label surface: %d hostile names -> %d requests, all under /repos/o/r/labels", len(hostile), len(tr.calls))
}

// With writes off (the default) nothing is sent at all.
func TestProtectionGuard_LabelWrites_OffByDefault_SendNothing(t *testing.T) {
	tr := &recordTransport{}
	h := &labels.GitHubHost{Repo: "o/r", T: tr, B: &forge.Budget{}, PerPage: 30, Backoff: testBackoff, Now: func() time.Time { return t0 }}
	if err := h.Create(context.Background(), labels.Label{Name: "x", Color: "ffffff"}); err == nil {
		t.Error("create succeeded with writes off")
	}
	if err := h.Update(context.Background(), "x", labels.Label{Name: "y", Color: "ffffff"}); err == nil {
		t.Error("update succeeded with writes off")
	}
	if len(tr.calls) != 0 {
		t.Errorf("transport calls with writes off = %d, want 0", len(tr.calls))
	}
}

// No production source file of the AgenticFactory carries a protection-related
// string literal. Comments are ignored; only string literals count.
func TestProtectionGuard_NoProductionLiteralMentionsProtectionSettings(t *testing.T) {
	forbidden := []string{
		"protection", "enforce_admins", "allow_force_pushes", "allow_deletions",
		"required_linear_history", "required_status_checks", "--admin", "--force", "force-with-lease",
	}
	root := ".." // internal/agenticfactory
	fset := token.NewFileSet()
	files, literals := 0, 0
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() && d.Name() == "e2e" {
			return filepath.SkipDir
		}
		if d.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		f, perr := parser.ParseFile(fset, path, nil, 0)
		if perr != nil {
			return perr
		}
		files++
		ast.Inspect(f, func(n ast.Node) bool {
			lit, ok := n.(*ast.BasicLit)
			if !ok || lit.Kind != token.STRING {
				return true
			}
			s, uerr := strconv.Unquote(lit.Value)
			if uerr != nil {
				return true
			}
			literals++
			low := strings.ToLower(s)
			for _, bad := range forbidden {
				if strings.Contains(low, bad) {
					t.Errorf("%s: string literal mentions %q: %q", fset.Position(lit.Pos()), bad, s)
				}
			}
			return true
		})
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	// Positive control: an empty scan is not a pass.
	if files < 10 || literals < 100 {
		t.Fatalf("scan looked at %d files and %d literals: too little to mean anything", files, literals)
	}
	t.Logf("MEASURED scan: %d production files, %d string literals, 0 protection-related", files, literals)
}

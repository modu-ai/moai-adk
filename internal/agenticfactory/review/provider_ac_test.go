package review

import (
	"context"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// AC-AF-015: Codex is reached only through the explicit opt-in key and GLM never; and (F22) a test double
// does not live in a production file.

func TestProvider_Codex_UnsetFalseOrNonBoolMakesZeroCalls(t *testing.T) {
	for name, lk := range map[string]Lookup{
		"nil lookup": nil,
		"unset":      func(string) (any, bool) { return nil, false },
		"false":      func(string) (any, bool) { return false, true },
		"string":     func(string) (any, bool) { return "true", true },
		"number":     func(string) (any, bool) { return 1, true },
	} {
		cd := &counters{}
		_, err := NewCodex(lk, cd.deps()).Review(context.Background(), Request{HeadSHA: headA, Kind: KindCodeReview})
		if err != ErrNotConfigured || cd.lookPath != 0 || cd.run != 0 {
			t.Errorf("%s: err=%v lookPath=%d run=%d, want ErrNotConfigured and no outward call", name, err, cd.lookPath, cd.run)
		}
	}
}

func TestProvider_Codex_TrueCallsAtLeastOnce(t *testing.T) {
	cd := &counters{}
	on := func(key string) (any, bool) { return key == CodexOptInKey, key == CodexOptInKey }
	res, err := NewCodex(on, cd.deps()).Review(context.Background(), Request{HeadSHA: headA, Kind: KindCodeReview})
	if err != nil || res.Verdict != VerdictPass || cd.run < 1 {
		t.Errorf("res=%+v err=%v run=%d, want at least one call once opted in", res, err, cd.run)
	}
}

// countingProvider is a Provider that only counts how often it is actually called.
type countingProvider struct{ calls int }

func (c *countingProvider) Review(context.Context, Request) (Result, error) {
	c.calls++
	return Result{HeadSHA: headA, Verdict: VerdictPass}, nil
}

// callIfEnabled is how a caller uses the activation predicate: a provider is called only when enabled.
func callIfEnabled(name string, lk Lookup, p Provider) {
	if ProviderEnabled(name, lk) {
		_, _, _ = CallOnce(context.Background(), p, Request{HeadSHA: headA, Kind: KindCodeReview})
	}
}

// (d) GLM calls: zero, counted. A positive control (codex with its key, claude without any) shows the counter
// does move when a provider is enabled, so a zero is a measurement and not a broken counter.
func TestProvider_GLM_CallsAreZeroEvenWithEveryKeyTrue(t *testing.T) {
	everything := func(string) (any, bool) { return true, true }
	glm, codex, claude := &countingProvider{}, &countingProvider{}, &countingProvider{}
	for _, name := range []string{ProviderGLM, "glm-managed", "zai", ""} {
		callIfEnabled(name, everything, glm)
	}
	callIfEnabled(ProviderCodex, everything, codex)
	callIfEnabled(ProviderClaude, nil, claude)
	if glm.calls != 0 {
		t.Errorf("GLM calls = %d, want 0", glm.calls)
	}
	if codex.calls != 1 || claude.calls != 1 {
		t.Errorf("positive control: codex=%d claude=%d, want 1 and 1", codex.calls, claude.calls)
	}
}

// F22: test doubles belong to _test.go files; a production file that defines one ships it.
func TestProvider_NoTestDoubleLivesInAProductionFile(t *testing.T) {
	files, err := filepath.Glob("*.go")
	if err != nil || len(files) == 0 {
		t.Fatalf("glob: %v (%d files)", err, len(files))
	}
	for _, f := range files {
		if strings.HasSuffix(f, "_test.go") {
			continue
		}
		src, err := os.ReadFile(f)
		if err != nil {
			t.Fatal(err)
		}
		af, err := parser.ParseFile(token.NewFileSet(), f, src, 0)
		if err != nil {
			t.Fatal(err)
		}
		for _, d := range af.Decls {
			gd, ok := d.(*ast.GenDecl)
			if !ok || gd.Tok != token.TYPE {
				continue
			}
			for _, sp := range gd.Specs {
				n := sp.(*ast.TypeSpec).Name.Name
				for _, p := range []string{"Fake", "Mock", "Stub", "Spy"} {
					if strings.HasPrefix(n, p) {
						t.Errorf("%s defines the test double %s", f, n)
					}
				}
			}
		}
	}
}

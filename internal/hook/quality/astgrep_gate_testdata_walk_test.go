package quality

// astgrep_gate_testdata_walk_test.go — AC-PAS-013 (SPEC-PRECOMMIT-AFFECTED-SCOPE-001
// REQ-PAS-013): the suppression-policy check judges production code, so its file
// walk skips every testdata directory. The Go suppression fixture puts
// @MX:REASON BEFORE the marker because ast-grep suppresses the line immediately
// following the marker; the checker requires the opposite order, so a walk that
// reached the fixture reported it as a violation.

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const (
	goFixtureRel  = "internal/astgrep/testdata/fixtures/go/suppressed.go"
	prodBareMarks = "package fixtures\n\nfunc F() {\n\t// ast-grep-ignore\n\tvar x interface{}\n\t_ = x\n}\n"
)

// realGoSuppressionFixture reads the repository's own Go fixture.
func realGoSuppressionFixture(t *testing.T) []byte {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("..", "..", "astgrep", "testdata", "fixtures", "go", "suppressed.go"))
	if err != nil {
		t.Fatalf("read the Go suppression fixture: %v", err)
	}
	return data
}

func runSuppressionGate(t *testing.T, project string) (bool, string) {
	t.Helper()
	return RunAstGrepGateV2(context.Background(), project, &AstGrepGateConfig{Enabled: true})
}

func TestSuppressionWalkSkipsTestdata(t *testing.T) {
	fixture := realGoSuppressionFixture(t)

	t.Run("the fixture alone is not reported", func(t *testing.T) {
		project := t.TempDir()
		writeScopeFiles(t, project, map[string]string{goFixtureRel: string(fixture)})
		ok, out := runSuppressionGate(t, project)
		if !ok || strings.Contains(out, "SUPPRESSION_WITHOUT_REASON") {
			t.Errorf("the testdata fixture was examined: ok = %v\n%s", ok, out)
		}
	})

	t.Run("production code under a fixtures segment is still reported", func(t *testing.T) {
		project := t.TempDir()
		// The production path carries a `fixtures` segment and no `testdata`
		// segment: a walk that hid every `fixtures` directory would miss it.
		writeScopeFiles(t, project, map[string]string{
			goFixtureRel:                    string(fixture),
			"internal/prod/fixtures/bad.go": prodBareMarks,
		})
		ok, out := runSuppressionGate(t, project)
		if ok {
			t.Fatal("a bare suppression marker in production source was not reported")
		}
		if got := strings.Count(out, "[SUPPRESSION_WITHOUT_REASON]"); got != 1 {
			t.Errorf("violations = %d, want exactly 1 (the production file):\n%s", got, out)
		}
		if !strings.Contains(out, "internal/prod/fixtures/bad.go") || strings.Contains(out, "testdata") {
			t.Errorf("the violation is not the production file's:\n%s", out)
		}
	})

	t.Run("a testdata directory above the project root does not hide production code", func(t *testing.T) {
		project := filepath.Join(t.TempDir(), "testdata", "proj")
		writeScopeFiles(t, project, map[string]string{"internal/prod/bad.go": prodBareMarks})
		ok, out := runSuppressionGate(t, project)
		if ok || !strings.Contains(out, "internal/prod/bad.go") {
			t.Errorf("production code below an ancestor named testdata was skipped: ok = %v\n%s", ok, out)
		}
	})
}

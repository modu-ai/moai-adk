package spec

// zz_t565_lint_census_test.go — card t565 corpus lint census through the same
// Linter the CLI uses, without compiling internal/cli. The committed copy lives
// under .moai/reports/t565/probe/; it is copied into internal/spec only for the
// run and removed afterwards.
//
// RE-DERIVATION
//
//	cp .moai/reports/t565/probe/zz_t565_lint_census_test.go internal/spec/
//	T565_CENSUS_OUT=<dir> go test ./internal/spec -count=1 -run '^TestT565LintCensus$' -v -timeout 900s
//	rm internal/spec/zz_t565_lint_census_test.go
//
// OUTPUT (in T565_CENSUS_OUT, default a t.TempDir()):
//
//	coverage.tsv   one row per CoverageIncomplete: <.moai/specs/...>\t<REQ id>  (unsorted)
//	duplicate.tsv  one row per DuplicateAcceptanceID: <.moai/specs/...>\t<line>\t<message>
//	codes.tsv      <code>\t<count> for every finding code
//
// The rows use the same shape the t561/t564 jq extraction produces from
// `moai spec lint --json`, so the two instruments can be compared with cmp.

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"
)

var t565ReqInMessage = regexp.MustCompile(`^REQ (REQ-[^ ]+) is`)

func TestT565LintCensus(t *testing.T) {
	out := os.Getenv("T565_CENSUS_OUT")
	if out == "" {
		out = t.TempDir()
	}
	if err := os.MkdirAll(out, 0o755); err != nil {
		t.Fatal(err)
	}

	linter := NewLinter(LinterOptions{BaseDir: "../../.moai/specs"})
	report, err := linter.Lint(nil)
	if err != nil {
		t.Fatal(err)
	}

	rel := func(p string) string {
		if i := strings.Index(p, ".moai/specs/"); i >= 0 {
			return p[i:]
		}
		return p
	}

	var coverage, duplicate []string
	codes := map[string]int{}
	for _, f := range report.Findings {
		codes[f.Code]++
		switch f.Code {
		case "CoverageIncomplete":
			m := t565ReqInMessage.FindStringSubmatch(f.Message)
			if m == nil {
				t.Fatalf("unexpected CoverageIncomplete message shape: %q", f.Message)
			}
			coverage = append(coverage, rel(f.File)+"\t"+m[1])
		case "DuplicateAcceptanceID":
			duplicate = append(duplicate, fmt.Sprintf("%s\t%d\t%s", rel(f.File), f.Line, f.Message))
		}
	}

	var codeRows []string
	for k, v := range codes {
		codeRows = append(codeRows, fmt.Sprintf("%s\t%d", k, v))
	}
	sort.Strings(codeRows)

	write := func(name string, rows []string) {
		body := ""
		if len(rows) > 0 {
			body = strings.Join(rows, "\n") + "\n"
		}
		if err := os.WriteFile(filepath.Join(out, name), []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	write("coverage.tsv", coverage)
	write("duplicate.tsv", duplicate)
	write("codes.tsv", codeRows)

	if len(report.Findings) == 0 {
		t.Fatalf("0 findings over the corpus; a linter that read nothing would say the same")
	}
	t.Logf("OUT = %s", out)
	t.Logf("findings = %d  CoverageIncomplete = %d  DuplicateAcceptanceID = %d", len(report.Findings), len(coverage), len(duplicate))
}

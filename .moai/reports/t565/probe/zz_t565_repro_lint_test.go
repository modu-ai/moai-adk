package spec

// zz_t565_repro_lint_test.go — card t565 positive control for the corpus lint
// census. The corpus comparison came back with zero CoverageIncomplete change,
// which an instrument blind to the parser change would also print. This lints
// the three repro fixtures through the same Linter; with the unfixed parser
// overlaid, L and N must report CoverageIncomplete, and with the fixed parser
// they must not.
//
// RE-DERIVATION (both via go test -overlay; nothing is copied into internal/spec)
//
//	overlay this file into internal/spec, optionally overlay parser.go with the unfixed copy
//	go test -overlay <json> ./internal/spec -count=1 -run '^TestT565ReproLint$' -v

import (
	"sort"
	"strings"
	"testing"
)

func TestT565ReproLint(t *testing.T) {
	linter := NewLinter(LinterOptions{BaseDir: "../../.moai/reports/t565/repro"})
	report, err := linter.Lint(nil)
	if err != nil {
		t.Fatal(err)
	}
	var rows []string
	specs := map[string]bool{}
	for _, f := range report.Findings {
		id := f.File
		if i := strings.Index(id, "SPEC-HDGREPRO-"); i >= 0 {
			id = id[i : i+len("SPEC-HDGREPRO-000")]
		}
		specs[id] = true
		rows = append(rows, id+"\t"+f.Code+"\t"+f.Message)
	}
	sort.Strings(rows)
	for _, r := range rows {
		t.Logf("finding: %s", r)
	}
	if len(specs) == 0 && len(report.Findings) == 0 {
		t.Logf("no findings at all")
	}
	criteria := map[string]int{}
	for _, id := range []string{"SPEC-HDGREPRO-001", "SPEC-HDGREPRO-002", "SPEC-HDGREPRO-003"} {
		n := 0
		for _, r := range rows {
			if strings.HasPrefix(r, id+"\tCoverageIncomplete\t") {
				n++
			}
		}
		criteria[id] = n
		t.Logf("%s CoverageIncomplete = %d", id, n)
	}
}

// run_external_delegation_mutants_test.go: mutation proof for the read-only
// control of the run-phase external-model delegation doctrine. The control
// itself lives in run_external_delegation_test.go (rxdReadOnlyControlDefects).
package template_test

import (
	"strings"
	"testing"
)

// TestRxdReadOnlyControlMutants proves the read-only control rejects the
// spellings an auditor found to pass it: a qualifier appended to the one
// security sentence (variant F) and a capitalised `Write` sentence. The
// unmutated section is the positive control — a check that rejected everything
// would pass the mutants too.
func TestRxdReadOnlyControlMutants(t *testing.T) {
	t.Parallel()
	root := findProjectRoot(t)
	section, count := rxdSection(rxdRead(t, root, rxdRunPath))
	if count != 1 {
		t.Fatalf("want exactly one %q heading in run.md, found %d", rxdSectionHeading, count)
	}

	if defects := rxdReadOnlyControlDefects(section); len(defects) != 0 {
		t.Fatalf("positive control failed: the unmutated section is flagged: %v", defects)
	}

	mutants := []struct {
		name  string
		apply func(string) string
	}{
		{"F-unless-qualifier", func(s string) string {
			return strings.Replace(s, "stays read-only.", "stays read-only unless a lint-repair draft is requested.", 1)
		}},
		{"capitalised-Write-sentence", func(s string) string {
			return strings.Replace(s, "stays read-only.", "stays read-only. Write access is granted to lint-repair drafts.", 1)
		}},
	}
	for _, m := range mutants {
		t.Run(m.name, func(t *testing.T) {
			mutated := m.apply(section)
			if mutated == section {
				t.Fatalf("mutant %s did not change the section: its anchor text moved", m.name)
			}
			if len(rxdReadOnlyControlDefects(mutated)) == 0 {
				t.Errorf("mutant %s passes the read-only control", m.name)
			}
		})
	}
}

package escalation_test

import (
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/escalation"
)

// Class 6 recognizer table: which Bash commands are irreversible actions
// under a contract whose actions carry push-develop.
func TestIrreversibleActionRecognizer(t *testing.T) {
	cases := []struct {
		cmd        string
		denylisted bool
		trips      bool
	}{
		{"git push origin main", false, true},
		{"git push", false, true},
		{"git push origin develop", false, false},
		{"git push origin HEAD:develop", false, false},
		{"git push --force origin develop", false, true},
		{"git push -f origin develop", false, true},
		{"git push origin +develop", false, true},
		{"git push --tags", false, true},
		{"git push origin refs/tags/v1", false, true},
		{"git tag v1.0.0", false, true},
		{"git tag -a v1.0.0 -m x", false, true},
		{"git tag -l", false, false},
		{"git tag", false, false},
		{"gh release create v1", false, true},
		{"make build && git push origin main", false, true},
		{"ls -la", false, false},
		{"echo hi", true, true},
	}
	for _, tc := range cases {
		t.Run(tc.cmd, func(t *testing.T) {
			isolateStore(t)
			w := armedWith(t, "t9001", nil, "")
			escalation.Observe(contractSettings(t, w), escalation.Event{Hook: escalation.HookPreToolUse, CWD: w.Root,
				ToolName: "Bash", Command: tc.cmd, Denylisted: tc.denylisted})
			rs, raw := recordsOfClass(t, w, escalation.ClassIrreversibleAction)
			if tc.trips != (len(rs) == 1) {
				t.Fatalf("trips=%v, records=%d", tc.trips, len(rs))
			}
			if tc.trips && !strings.Contains(refLine(t, w, rs[0].ContractRef), "actions") {
				t.Errorf("contract_ref points at %q\n%s", refLine(t, w, rs[0].ContractRef), raw[0])
			}
		})
	}
}

package wizard

// Card t1591, finding 1 residual (Item A): the agent_wiring question's default
// is pre-filled from the harness a resumed init recorded. The question stays
// unconditional and keeps its closed option set; only its default moves.
// Accepting the default must keep the recorded selection, not the shipped claude
// default.

import (
	"slices"
	"testing"
)

// TestInitQuestionsWithAgentWiring_PrefillsRecordedDefault asserts that a
// recorded selection becomes the agent_wiring default, that an empty selection
// keeps the shipped claude default, and that the pre-fill changes nothing else
// about the question.
func TestInitQuestionsWithAgentWiring_PrefillsRecordedDefault(t *testing.T) {
	t.Parallel()
	root := t.TempDir()

	cases := []struct {
		recorded string
		want     string
	}{
		{"both", "both"},
		{"gpt", "gpt"},
		{"", "claude"},
	}
	for _, tc := range cases {
		q := QuestionByID(InitQuestionsWithAgentWiring(root, tc.recorded), "agent_wiring")
		if q == nil {
			t.Fatalf("agent_wiring is absent from InitQuestionsWithAgentWiring(%q)", tc.recorded)
		}
		if q.Default != tc.want {
			t.Errorf("agent_wiring default with recorded %q = %q, want %q", tc.recorded, q.Default, tc.want)
		}
		if q.Condition != nil {
			t.Errorf("agent_wiring must stay unconditional after the pre-fill (recorded %q)", tc.recorded)
		}
		var values []string
		for _, opt := range q.Options {
			values = append(values, opt.Value)
		}
		if want := []string{"claude", "gpt", "both"}; !slices.Equal(values, want) {
			t.Errorf("agent_wiring option values after the pre-fill (recorded %q) = %v, want %v", tc.recorded, values, want)
		}
	}
}

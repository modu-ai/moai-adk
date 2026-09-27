package escalation_test

import (
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/escalation"
)

// REQ-AE-001 (activation gate): only the effective mode "contract" activates
// the detector. The mode is read through config.ResolveAutonomy, which maps
// absent, empty, and unrecognized values to guided.
func TestActiveOnlyUnderContractMode(t *testing.T) {
	for _, tc := range []struct {
		mode string
		want bool
	}{
		{"", false},
		{"guided", false},
		{"bogus", false},
		{"contract", true},
	} {
		s := config.ResolveAutonomy(config.WorkflowConfig{Autonomy: config.AutonomyConfig{Mode: tc.mode}})
		if got := escalation.Active(s); got != tc.want {
			t.Errorf("Active(mode %q) = %v, want %v", tc.mode, got, tc.want)
		}
	}
}

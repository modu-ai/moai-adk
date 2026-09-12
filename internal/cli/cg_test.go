package cli

import (
	"errors"
	"testing"
)

// SPEC-MOAI-CG-RETIRE-001 supersedes the old successful CG launch
// characterization: former profile, bypass and passthrough inputs must fail
// before they can select a replacement provider.
func TestCGFormerLauncherInputsAreRetired(t *testing.T) {
	for _, args := range [][]string{nil, {"-p", "team", "--print"}, {"--some-claude-flag"}, {"--help"}, {"-h"}, {"--kanban"}, {"-k"}, {"-b"}, {"-w", "feature", "--spawn"}} {
		if err := runCG(cgCmd, args); !errors.Is(err, errCGRetired) {
			t.Fatalf("%v: %v", args, err)
		}
	}
}

package wizard

import (
	"os"
	"strings"
	"testing"
)

// pat_mask_test.go exercises SPEC-CLIFIX-HYGIENE-001 AC-HYG-001-007:
// the github_token and gitlab_token wizard questions MUST render with
// password echo masking so the entered PAT is not displayed in cleartext.
//
// The pre-fix buildInputField (wizard.go huh.NewInput block) uses a plain
// input with no echo masking, so a typed token is visible on screen.

// TestPATMaskEchoModeWired reproduces the defect at the source level: the
// huh.Input for token questions had no EchoMode, so the literal
// `EchoMode(huh.EchoModePassword)` (or equivalent) was absent from wizard.go.
// Pre-fix this assertion fails; post-fix the call is present.
func TestPATMaskEchoModeWired(t *testing.T) {
	src, err := os.ReadFile("wizard.go")
	if err != nil {
		t.Fatalf("read wizard.go: %v", err)
	}
	srcStr := string(src)
	// The huh v2 API is inp.EchoMode(huh.EchoModePassword).
	if !strings.Contains(srcStr, "EchoMode(huh.EchoModePassword)") {
		t.Errorf("wizard.go does not wire EchoMode(huh.EchoModePassword) — token PATs render in cleartext")
	}
	// The masking must be gated on the token question IDs.
	if !strings.Contains(srcStr, "github_token") || !strings.Contains(srcStr, "gitlab_token") {
		t.Errorf("wizard.go must reference both github_token and gitlab_token to gate PAT masking")
	}
}

// TestPATMaskSecretIDs is the behavioral lock-in for the exported predicate
// buildInputField uses to decide echo masking. It passes once the predicate
// exists and correctly classifies the two token question IDs as secret.
func TestPATMaskSecretIDs(t *testing.T) {
	secret := []string{"github_token", "gitlab_token"}
	for _, id := range secret {
		if !IsSecretInputID(id) {
			t.Errorf("IsSecretInputID(%q) = false, want true (PAT must be masked)", id)
		}
	}
	plain := []string{"user_name", "project_name", "conversation_language", ""}
	for _, id := range plain {
		if IsSecretInputID(id) {
			t.Errorf("IsSecretInputID(%q) = true, want false (non-token field must not be masked)", id)
		}
	}
}

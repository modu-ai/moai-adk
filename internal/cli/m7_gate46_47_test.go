package cli

// m7_gate46_47_test.go — SPEC-USERASSET-DEPLOY-GUARD-001 gate rounds 46/47,
// cli family:
//
//   - 46-3: the resume checkpoint's completeness arm reads BOTH deployment
//     faces. A project carrying llm.yaml AND .mcp.json is COMPLETE — a
//     stale user-global journal (the journal carries no project identity)
//     must not hijack it into a resume that would consume another
//     project's pending intent. A project still missing .mcp.json (the M6
//     window) stays resumable.
//   - 47-4: the resume applies the wizard's Jev answer — a resume that
//     skipped it flipped an explicit wizard selection back to false.

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func corruptUserManifestHome(t *testing.T) string {
	t.Helper()
	home := t.TempDir()
	if err := os.MkdirAll(filepath.Join(home, ".moai"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(home, ".moai", "user-assets.json"), []byte("{ corrupt\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	return home
}

func TestInitResumeCheckpointReadsBothDeploymentFaces(t *testing.T) {
	home := corruptUserManifestHome(t)

	// The template-shaped llm.yaml — no deployment_mode key: the M6 window
	// (deployed, post-deploy setup not finished) stays resumable.
	deployed := t.TempDir()
	sections := filepath.Join(deployed, ".moai", "config", "sections")
	if err := os.MkdirAll(sections, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(sections, "llm.yaml"), []byte("harness: claude\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if !initResumeCheckpoint(home, deployed) {
		t.Fatal("a deployed project without the post-deploy marker must stay resumable (the M6 window)")
	}

	// The post-path RAN (ApplyDeployMode wrote deployment_mode — the one
	// key the template never ships): COMPLETE — the stale user-global
	// journal must not route it into a resume (gate 46-3).
	completed := t.TempDir()
	sections = filepath.Join(completed, ".moai", "config", "sections")
	if err := os.MkdirAll(sections, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(sections, "llm.yaml"), []byte("harness: claude\ndeployment_mode: local\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if initResumeCheckpoint(home, completed) {
		t.Fatal("a COMPLETE project was routed into resume — the stale user-global journal would be consumed as this project's pending intent (gate 46-3)")
	}

	// Undeployed: no llm.yaml — never resumable.
	if initResumeCheckpoint(home, t.TempDir()) {
		t.Fatal("an undeployed project was routed into resume")
	}

	// No interruption (healthy manifest): not resumable regardless.
	healthyHome := t.TempDir()
	if initResumeCheckpoint(healthyHome, deployed) {
		t.Fatal("a healthy user store routed a deployed project into resume")
	}
}

func jevEnabledInWorkflowYAML(content string) bool {
	lines := strings.Split(content, "\n")
	inJev := false
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		switch {
		case strings.HasPrefix(trimmed, "jev:"):
			inJev = true
		case inJev && strings.HasPrefix(line, " ") && strings.HasPrefix(trimmed, "enabled:"):
			return strings.TrimSpace(strings.TrimPrefix(trimmed, "enabled:")) == "true"
		case inJev && trimmed != "" && !strings.HasPrefix(line, " "):
			return false // the jev block ended
		}
	}
	return false
}

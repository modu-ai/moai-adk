package hook

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
)

// TestBuildSessionTitle_LeadNameWinsOverSPEC pins the repair for issue #1596: a
// kanban or factory lead is titled by the name it actually launched under, even
// when the project carries SPECs.
//
// The reported failure is exactly this shape — messaging worked (the `--name`
// injection has shipped since c326eb4e0) while the session list showed an
// unrelated SPEC heading, because a session name and a session TITLE are two
// separate registrations and only the first one was ever made. Titling from the
// SPEC is the correct default for every OTHER session, which is why this test
// asserts the ordering rather than the absence of the SPEC branch.
//
// These subtests set a process-global environment variable, so they are
// deliberately NOT parallel.
func TestBuildSessionTitle_LeadNameWinsOverSPEC(t *testing.T) {
	cwd := t.TempDir()
	specDir := filepath.Join(cwd, ".moai", "specs", "SPEC-MIGRATE-002")
	if err := os.MkdirAll(specDir, 0o755); err != nil {
		t.Fatalf("failed to create spec directory: %v", err)
	}
	if err := os.WriteFile(filepath.Join(specDir, "spec.md"), []byte("# 뷰어 프런트엔드 이식\n"), 0o644); err != nil {
		t.Fatalf("failed to create spec.md: %v", err)
	}
	specTitle := "SPEC-MIGRATE-002: 뷰어 프런트엔드 이식"

	h := newHookHandler("ko")

	// Negative control FIRST: with the variable unset, the SPEC title is what a
	// session gets. Without this the positive case below could pass for the
	// wrong reason — a leadSessionTitle that returned its argument on every
	// session would look identical on the lead alone.
	t.Run("not a lead -> SPEC title (unchanged default)", func(t *testing.T) {
		// t.Setenv first so the prior value is restored on cleanup; the unset
		// that follows is what the negative control actually needs, because
		// presence — not emptiness — is the signal downstream readers use.
		t.Setenv(config.EnvMoaiKanbanLeadName, "")
		if err := os.Unsetenv(config.EnvMoaiKanbanLeadName); err != nil {
			t.Fatalf("failed to unset %s: %v", config.EnvMoaiKanbanLeadName, err)
		}

		if got := h.buildSessionTitle(context.Background(), cwd, ""); got != specTitle {
			t.Errorf("buildSessionTitle() = %q, want the SPEC title %q", got, specTitle)
		}
	})

	t.Run("lead -> the session's own name, not the SPEC", func(t *testing.T) {
		t.Setenv(config.EnvMoaiKanbanLeadName, "lead")

		got := h.buildSessionTitle(context.Background(), cwd, "")
		if got != "lead" {
			t.Errorf("buildSessionTitle() = %q, want the lead's own name %q", got, "lead")
		}
		if got == specTitle {
			t.Error("the lead was titled by an unrelated SPEC — issue #1596 regressed")
		}
	})

	// A bumped lead (a second lead launched while the first is live) carries the
	// bumped name in argv, so the title must carry it too — a title guessing the
	// bare role would disagree with the address peers actually dispatch to.
	t.Run("bumped lead -> the bumped name", func(t *testing.T) {
		t.Setenv(config.EnvMoaiKanbanLeadName, "lead-1")

		if got := h.buildSessionTitle(context.Background(), cwd, "lead-1"); got != "lead-1" {
			t.Errorf("buildSessionTitle() = %q, want the bumped name %q", got, "lead-1")
		}
	})

	// The first-wins guard is ABOVE this branch: a /rename (recorded as a
	// custom-title) still wins, exactly as it does over the SPEC branch.
	t.Run("existing title wins over the lead name", func(t *testing.T) {
		t.Setenv(config.EnvMoaiKanbanLeadName, "lead")

		path := writeTranscript(t, `{"type":"custom-title","customTitle":"운영자가 고른 이름","sessionId":"s1"}`)
		if got := h.buildSessionTitle(context.Background(), cwd, path); got != "" {
			t.Errorf("buildSessionTitle() = %q, want %q — a rename must not be clobbered", got, "")
		}
	})

	// A whitespace-only value is not a name. Trimming to empty falls through to
	// the SPEC branch rather than registering a blank title.
	t.Run("whitespace-only value -> falls through to the SPEC title", func(t *testing.T) {
		t.Setenv(config.EnvMoaiKanbanLeadName, "   ")

		if got := h.buildSessionTitle(context.Background(), cwd, ""); got != specTitle {
			t.Errorf("buildSessionTitle() = %q, want the SPEC title %q", got, specTitle)
		}
	})
}

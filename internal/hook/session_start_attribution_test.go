package hook

// session_start_attribution_test.go — card t1776: when the current-session
// side-channel already records a DIFFERENT session id, the source_session_id
// attribution injection carries a one-line NOTE — an agent that attributes
// its work through `moai session current` would otherwise credit the
// recorded (other) session.

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/session"
)

func TestSessionStartAttributionMismatchNotice(t *testing.T) {
	root := t.TempDir()
	sideDir := filepath.Join(root, filepath.Dir(session.CurrentSideChannelFile))
	if err := os.MkdirAll(sideDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	this := "11111111-1111-4111-8111-111111111111"
	other := "22222222-2222-4222-8222-222222222222"
	if err := os.WriteFile(filepath.Join(root, session.CurrentSideChannelFile), []byte(other), 0o600); err != nil {
		t.Fatalf("seed side channel: %v", err)
	}

	h := NewSessionStartHandler(nil)
	out, err := h.Handle(t.Context(), &HookInput{
		SessionID:  this,
		ProjectDir: root,
		CWD:        root,
		Source:     "startup",
	})
	if err != nil {
		t.Fatalf("Handle: %v", err)
	}
	ctx := ""
	if out.HookSpecificOutput != nil {
		ctx = out.HookSpecificOutput.AdditionalContext
	}
	if !strings.Contains(ctx, "source_session_id="+this) {
		t.Fatalf("attribution line missing this session's id:\n%s", ctx)
	}
	if !strings.Contains(ctx, "before this session started") {
		t.Fatalf("mismatch notice missing the truthful pre-state phrasing (recorded id %s, this %s):\n%s", other, this, ctx)
	}
	if strings.Contains(ctx, "returns the recorded one") {
		t.Fatalf("mismatch notice makes the false will-return claim (the write below already replaces the side channel):\n%s", ctx)
	}
	if !strings.Contains(ctx, other) {
		t.Fatalf("mismatch notice does not name the recorded id:\n%s", ctx)
	}

	t.Run("matching_side_channel_no_notice", func(t *testing.T) {
		root2 := t.TempDir()
		sideDir2 := filepath.Join(root2, filepath.Dir(session.CurrentSideChannelFile))
		if err := os.MkdirAll(sideDir2, 0o755); err != nil {
			t.Fatalf("mkdir: %v", err)
		}
		if err := os.WriteFile(filepath.Join(root2, session.CurrentSideChannelFile), []byte(this), 0o600); err != nil {
			t.Fatalf("seed side channel: %v", err)
		}
		out2, err := h.Handle(t.Context(), &HookInput{
			SessionID:  this,
			ProjectDir: root2,
			CWD:        root2,
			Source:     "startup",
		})
		if err != nil {
			t.Fatalf("Handle: %v", err)
		}
		ctx2 := ""
		if out2.HookSpecificOutput != nil {
			ctx2 = out2.HookSpecificOutput.AdditionalContext
		}
		if strings.Contains(ctx2, "before this session started") {
			t.Fatalf("matching side channel must not carry the mismatch notice:\n%s", ctx2)
		}
	})
}

//go:build unix

package hook

// session_start_attribution_test.go — card t1776: when the current-session
// side-channel already records a DIFFERENT session id, the source_session_id
// attribution injection carries a one-line NOTE. The note's overwrite claim
// must match the WRITE OUTCOME (audit round-2 P2): a read-only environment
// leaves the foreign id in place, and the notice must say so instead of
// claiming an overwrite that never happened.

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

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
	if !strings.Contains(ctx, "has overwritten it") {
		t.Fatalf("successful write must say the record was overwritten:\n%s", ctx)
	}

	t.Run("side_channel_symlinked_to_dev_zero_does_not_hang", func(t *testing.T) {
		// The prev-id read goes through the non-blocking regular-file guard:
		// a side channel symlinked to /dev/zero is refused instead of
		// reading forever (audit round-5 P2 — os.ReadFile on /dev/zero never
		// sees EOF and allocates until the session start dies).
		root4 := t.TempDir()
		sideDir4 := filepath.Join(root4, filepath.Dir(session.CurrentSideChannelFile))
		if err := os.MkdirAll(sideDir4, 0o755); err != nil {
			t.Fatalf("mkdir: %v", err)
		}
		if err := os.Symlink("/dev/zero", filepath.Join(root4, session.CurrentSideChannelFile)); err != nil {
			t.Fatalf("symlink: %v", err)
		}
		done := make(chan struct{}, 1)
		var ctx4 string
		go func() {
			out4, err := h.Handle(t.Context(), &HookInput{
				SessionID:  this,
				ProjectDir: root4,
				CWD:        root4,
				Source:     "startup",
			})
			if err == nil && out4.HookSpecificOutput != nil {
				ctx4 = out4.HookSpecificOutput.AdditionalContext
			}
			done <- struct{}{}
		}()
		select {
		case <-done:
			if strings.Contains(ctx4, "before this session started") {
				t.Fatalf("a /dev/zero side channel must be refused, not narrated:\n%s", ctx4)
			}
		case <-time.After(5 * time.Second):
			t.Fatal("Handle blocked on a /dev/zero side channel: the prev-id read must refuse non-regular files")
		}
	})

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

	t.Run("failed_write_reports_the_surviving_record", func(t *testing.T) {
		// A read-only state dir makes the overwrite fail — the notice must
		// then say the foreign record SURVIVED, not claim an overwrite
		// (audit round-2 P2).
		root3 := t.TempDir()
		sideDir3 := filepath.Join(root3, filepath.Dir(session.CurrentSideChannelFile))
		if err := os.MkdirAll(sideDir3, 0o755); err != nil {
			t.Fatalf("mkdir: %v", err)
		}
		// An existing file truncates with FILE write permission only — the
		// directory mode is irrelevant to the write, so the failure is
		// forced on the file itself (owner-read, no write).
		if err := os.WriteFile(filepath.Join(root3, session.CurrentSideChannelFile), []byte(other), 0o600); err != nil {
			t.Fatalf("seed side channel: %v", err)
		}
		if os.Geteuid() == 0 {
			t.Skip("root bypasses file write permissions — the failure cannot be forced (audit round-3 P2)")
		}
		if err := os.Chmod(filepath.Join(root3, session.CurrentSideChannelFile), 0o400); err != nil {
			t.Fatalf("chmod: %v", err)
		}
		t.Cleanup(func() { _ = os.Chmod(filepath.Join(root3, session.CurrentSideChannelFile), 0o644) })

		out3, err := h.Handle(t.Context(), &HookInput{
			SessionID:  this,
			ProjectDir: root3,
			CWD:        root3,
			Source:     "startup",
		})
		if err != nil {
			t.Fatalf("Handle: %v", err)
		}
		ctx3 := ""
		if out3.HookSpecificOutput != nil {
			ctx3 = out3.HookSpecificOutput.AdditionalContext
		}
		if strings.Contains(ctx3, "has overwritten it") {
			t.Fatalf("a failed write must not claim the overwrite:\n%s", ctx3)
		}
		if !strings.Contains(ctx3, "still holds") {
			t.Fatalf("a failed write must name the surviving foreign record:\n%s", ctx3)
		}
	})
}

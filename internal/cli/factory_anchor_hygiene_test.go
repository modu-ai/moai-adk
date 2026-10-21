package cli

// factory_anchor_hygiene_test.go — AC-LA-009 (REQ-LA-009): the audit log of the
// override path is opened without following a link, without blocking on a FIFO,
// and without writing through a non-regular file — the three measured repairs
// of card t1604 (O11 symlink, O20 FIFO, codex-690 open race), ported with the
// guard. The FIFO cells need Unix and live in the _unix_test file.

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
)

func TestLauncherAnchorAuditHygiene(t *testing.T) {
	verdict := &leaderAnchorVerdict{tree: "/card/tree", branch: "WT-audit-log"}

	t.Run("symlink_at_the_audit_log_is_refused", func(t *testing.T) {
		root := t.TempDir()
		auditPath := filepath.Join(root, config.DefaultLeaderAnchorAuditRelPath)
		if err := os.MkdirAll(filepath.Dir(auditPath), 0o755); err != nil {
			t.Fatal(err)
		}
		target := filepath.Join(t.TempDir(), "target.log")
		if err := os.WriteFile(target, []byte("untouched\n"), 0o600); err != nil {
			t.Fatal(err)
		}
		if err := os.Symlink(target, auditPath); err != nil {
			t.Skipf("create the symbolic link: %v", err)
		}
		if err := appendLeaderAnchorAuditRow(root, "claude", verdict); err == nil {
			t.Errorf("a symbolic link at the audit log was followed, want a refusal")
		}
		if got, _ := os.ReadFile(target); string(got) != "untouched\n" {
			t.Errorf("the link target was written: %q", got)
		}
	})

	t.Run("symlink_for_the_audit_directory_is_refused", func(t *testing.T) {
		root := t.TempDir()
		auditDir := filepath.Dir(filepath.Join(root, config.DefaultLeaderAnchorAuditRelPath))
		if err := os.MkdirAll(filepath.Dir(auditDir), 0o755); err != nil {
			t.Fatal(err)
		}
		elsewhere := t.TempDir()
		if err := os.Symlink(elsewhere, auditDir); err != nil {
			t.Skipf("create the symbolic link: %v", err)
		}
		if err := appendLeaderAnchorAuditRow(root, "claude", verdict); err == nil {
			t.Errorf("a linked audit directory was followed, want a refusal")
		}
		if entries, _ := os.ReadDir(elsewhere); len(entries) != 0 {
			t.Errorf("the link target directory gained %d entries", len(entries))
		}
	})

	t.Run("symlink_for_the_moai_directory_is_refused", func(t *testing.T) {
		root := t.TempDir()
		elsewhere := t.TempDir()
		if err := os.Symlink(elsewhere, filepath.Join(root, ".moai")); err != nil {
			t.Skipf("create the symbolic link: %v", err)
		}
		if err := appendLeaderAnchorAuditRow(root, "claude", verdict); err == nil {
			t.Errorf("a linked .moai directory was followed, want a refusal")
		}
		if entries, _ := os.ReadDir(elsewhere); len(entries) != 0 {
			t.Errorf("MkdirAll created %d entries through the link", len(entries))
		}
	})

	t.Run("regular_file_still_receives_its_row", func(t *testing.T) {
		root := t.TempDir()
		if err := appendLeaderAnchorAuditRow(root, "claude", verdict); err != nil {
			t.Fatalf("append to a fresh audit log: %v", err)
		}
		raw, err := os.ReadFile(filepath.Join(root, config.DefaultLeaderAnchorAuditRelPath))
		if err != nil {
			t.Fatal(err)
		}
		lines := strings.Split(strings.TrimSuffix(string(raw), "\n"), "\n")
		if len(lines) != 1 {
			t.Fatalf("the audit log has %d lines, want 1: %q", len(lines), raw)
		}
		for _, want := range []string{"backend=claude", allowCardAnchorFlag, "WT-audit-log"} {
			if !strings.Contains(lines[0], want) {
				t.Errorf("audit row %q lacks %q", lines[0], want)
			}
		}
	})

	auditHygieneUnixCells(t, verdict)
}

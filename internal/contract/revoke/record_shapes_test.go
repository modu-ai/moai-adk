package revoke_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/modu-ai/moai-adk/internal/contract/revoke"
	"github.com/modu-ai/moai-adk/internal/contract/sign/signtest"
	"github.com/modu-ai/moai-adk/internal/escalation"
)

// A linked directory cannot stand in for this project's revocation records.
// Refusal preserves the entry and never reads another tree as an empty ledger.
func TestBlockedRefusesSymlinkRecordDirectory(t *testing.T) {
	p := signtest.New(t)
	dir := escalation.RecordDir(p.Root, signtest.Card)
	outside := t.TempDir()
	if err := os.MkdirAll(filepath.Dir(dir), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(outside, dir); err != nil {
		t.Fatalf("create record-directory symlink fixture: %v", err)
	}
	blocked, err := revoke.Blocked(p.Root, signtest.Card, signtest.SpecID, "seal")
	if !blocked || err == nil {
		t.Fatalf("linked records = %v, %v; want blocked with error", blocked, err)
	}
	if info, err := os.Lstat(dir); err != nil || info.Mode()&os.ModeSymlink == 0 {
		t.Fatalf("linked record entry changed: %v, %v", info, err)
	}
}

func TestBlockedRefusesRegularRecordAncestor(t *testing.T) {
	p := signtest.New(t)
	dir := escalation.RecordDir(p.Root, signtest.Card)
	parent := filepath.Dir(dir)
	if err := os.MkdirAll(filepath.Dir(parent), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(parent, []byte("user bytes"), 0o600); err != nil {
		t.Fatal(err)
	}
	blocked, err := revoke.Blocked(p.Root, signtest.Card, signtest.SpecID, "seal")
	if !blocked || err == nil {
		t.Fatalf("regular ancestor = %v, %v; want blocked with error", blocked, err)
	}
	if got, err := os.ReadFile(parent); err != nil || string(got) != "user bytes" {
		t.Fatalf("ancestor payload changed: %q, %v", got, err)
	}
}

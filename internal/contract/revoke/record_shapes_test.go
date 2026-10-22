package revoke_test

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
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

func TestBlockedRefusesLinkedRecordAncestors(t *testing.T) {
	for _, linkRel := range []string{".moai", filepath.Join(".moai", "reports")} {
		for _, present := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/present=%t", linkRel, present), func(t *testing.T) {
				root, outside := t.TempDir(), t.TempDir()
				link := filepath.Join(root, linkRel)
				dir := escalation.RecordDir(root, signtest.Card)
				rel, err := filepath.Rel(link, dir)
				if err != nil {
					t.Fatal(err)
				}
				outsideDir := filepath.Join(outside, rel)
				if err := os.MkdirAll(filepath.Dir(outsideDir), 0o755); err != nil {
					t.Fatal(err)
				}
				if present {
					if err := os.Mkdir(outsideDir, 0o755); err != nil {
						t.Fatal(err)
					}
					if err := os.WriteFile(filepath.Join(outsideDir, "user.txt"), []byte("outside ledger"), 0o600); err != nil {
						t.Fatal(err)
					}
				}
				if err := os.MkdirAll(filepath.Dir(link), 0o755); err != nil {
					t.Fatal(err)
				}
				if err := os.Symlink(outside, link); err != nil {
					t.Fatalf("create record ancestor link: %v", err)
				}
				blocked, err := revoke.Blocked(root, signtest.Card, signtest.SpecID, "seal")
				if !blocked || err == nil || !strings.Contains(err.Error(), "list ") {
					t.Errorf("linked ancestor = %v, %v; want blocked with list error", blocked, err)
				}
				if got, err := os.Readlink(link); err != nil || got != outside {
					t.Errorf("record ancestor link changed: %q, %v", got, err)
				}
				if present {
					if got, err := os.ReadFile(filepath.Join(outsideDir, "user.txt")); err != nil || string(got) != "outside ledger" {
						t.Errorf("outside ledger changed: %q, %v", got, err)
					}
				} else if _, err := os.Lstat(outsideDir); !os.IsNotExist(err) {
					t.Errorf("missing outside record directory changed: %v", err)
				}
			})
		}
	}
}

func TestBlockedAcceptsTrustedRootAlias(t *testing.T) {
	root := t.TempDir()
	alias := filepath.Join(t.TempDir(), "project-alias")
	if err := os.Symlink(root, alias); err != nil {
		t.Fatalf("create trusted project alias: %v", err)
	}
	if blocked, err := revoke.Blocked(alias, signtest.Card, signtest.SpecID, "seal"); blocked || err != nil {
		t.Fatalf("empty ledger under trusted alias = %v, %v", blocked, err)
	}
	dir := escalation.RecordDir(root, signtest.Card)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	record := escalation.Record{SchemaVersion: 1, Card: signtest.Card, Spec: signtest.SpecID,
		Kind: escalation.KindRevoke, Class: revoke.RecordClass,
		Fingerprint: escalation.Fingerprint(revoke.RecordClass, "seal"), Status: escalation.StatusResolved,
		Decider: "human", Occurrences: 1, Observation: "revoked", Options: []string{"re-sign", "abandon"}}
	data, err := record.Marshal()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "revoke.md"), data, 0o600); err != nil {
		t.Fatal(err)
	}
	if blocked, err := revoke.Blocked(alias, signtest.Card, signtest.SpecID, "seal"); !blocked || err != nil {
		t.Fatalf("matching revoke under trusted alias = %v, %v", blocked, err)
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

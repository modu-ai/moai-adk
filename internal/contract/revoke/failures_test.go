package revoke_test

import (
	"errors"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/contract"
	"github.com/modu-ai/moai-adk/internal/contract/receipt"
	"github.com/modu-ai/moai-adk/internal/contract/revoke"
	"github.com/modu-ai/moai-adk/internal/contract/sign/signtest"
	"github.com/modu-ai/moai-adk/internal/escalation"
)

// TestRevokeFailuresWriteNothing covers the error paths ahead of the write:
// each returns an error and leaves no revoke event and no revoke record.
func TestRevokeFailuresWriteNothing(t *testing.T) {
	cases := []struct {
		name  string
		setup func(t *testing.T, p *signtest.Project)
		seams func() revoke.Seams
		want  string
	}{
		{
			name:  "store-open-fails",
			setup: signHuman,
			seams: func() revoke.Seams {
				s := seams()
				s.Store = func(string) (*receipt.Store, error) { return nil, errors.New("store unavailable") }
				return s
			},
			want: "store unavailable",
		},
		{
			name:  "head-read-fails",
			setup: signHuman,
			seams: func() revoke.Seams {
				s := seams()
				s.GitHead = func(string) (string, error) { return "", errors.New("no head") }
				return s
			},
			want: "no head",
		},
		{
			name:  "store-append-fails",
			setup: signHuman,
			seams: func() revoke.Seams {
				s := seams()
				s.Store = func(string) (*receipt.Store, error) {
					// A read-only directory: Verify reads an empty chain, the
					// append cannot create its file.
					dir := t.TempDir()
					if err := os.Chmod(dir, 0o555); err != nil {
						t.Fatal(err)
					}
					t.Cleanup(func() { _ = os.Chmod(dir, 0o755) })
					return &receipt.Store{Dir: dir}, nil
				}
				return s
			},
			want: "permission denied",
		},
		{
			name: "contract-unreadable",
			setup: func(t *testing.T, p *signtest.Project) {
				rel := signtest.SpecRel(signtest.SpecID, contract.ContractFile)
				_ = os.Remove(p.Path(rel))
				p.WriteFile(rel+"/inner", "a directory where the contract belongs")
			},
			seams: seams,
			want:  contract.ContractFile,
		},
		{
			name: "contract-does-not-decode",
			setup: func(t *testing.T, p *signtest.Project) {
				p.WriteFile(signtest.SpecRel(signtest.SpecID, contract.ContractFile), "card: [unclosed\n")
			},
			seams: seams,
			want:  "does not decode",
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if c.name == "store-append-fails" && (runtime.GOOS == "windows" || os.Geteuid() == 0) {
				t.Skip("directory permissions do not deny writes here")
			}
			p := signtest.New(t)
			c.setup(t, p)
			before := eventCount(t, p, receipt.KindRevoke)
			_, err := revoke.Revoke(opts(p), c.seams())
			if err == nil || !strings.Contains(err.Error(), c.want) {
				t.Fatalf("err = %v, want one naming %q", err, c.want)
			}
			if after := eventCount(t, p, receipt.KindRevoke); after != before {
				t.Errorf("revoke events %d → %d, want unchanged", before, after)
			}
			if r := records(t, p); len(r) != 0 {
				t.Errorf("revoke records written: %v", r)
			}
		})
	}
	t.Run("invalid-inputs", func(t *testing.T) {
		for _, o := range []revoke.Options{
			{Root: "", SpecID: signtest.SpecID, Card: signtest.Card},
			{Root: t.TempDir(), SpecID: "not-a-spec", Card: signtest.Card},
			{Root: t.TempDir(), SpecID: signtest.SpecID, Card: ""},
		} {
			if _, err := revoke.Revoke(o, seams()); !errors.Is(err, revoke.ErrUsage) {
				t.Errorf("%+v: err = %v, want ErrUsage", o, err)
			}
		}
	})
}

// TestRevokeDefaultSeams runs Revoke with no injected seams: the store opens
// under the isolated MOAI_HOME and the record carries the repository's real
// HEAD.
func TestRevokeDefaultSeams(t *testing.T) {
	p := signtest.New(t)
	signHuman(t, p)
	res, err := revoke.Revoke(opts(p), revoke.Seams{})
	if err != nil || res.Status != revoke.StatusRevoked {
		t.Fatalf("revoke: status %q err %v", res.Status, err)
	}
	head := strings.TrimSpace(p.Git("rev-parse", "HEAD"))
	if !regexp.MustCompile(`^[0-9a-f]{40}$`).MatchString(head) {
		t.Fatalf("fixture HEAD %q", head)
	}
	rs := records(t, p)
	if len(rs) != 1 {
		t.Fatalf("records = %v, want one", rs)
	}
	data, err := os.ReadFile(rs[0])
	if err != nil {
		t.Fatal(err)
	}
	rec, err := escalation.ParseRecord(data)
	if err != nil {
		t.Fatal(err)
	}
	if rec.HeadSHA != head {
		t.Errorf("record head %q, want %q", rec.HeadSHA, head)
	}
}

// TestBlockedReaderShapes covers the reader's directory handling: an
// unlistable record directory blocks with an error, and entries that are not
// record files are skipped.
func TestBlockedReaderShapes(t *testing.T) {
	t.Run("unlistable-directory-blocks", func(t *testing.T) {
		p := signtest.New(t)
		dir := escalation.RecordDir(p.Root, signtest.Card)
		p.WriteFile(relTo(t, p.Root, dir), "not a directory")
		blocked, err := revoke.Blocked(p.Root, signtest.Card, signtest.SpecID, "seal")
		if !blocked || err == nil {
			t.Fatalf("Blocked = %v, err %v; want true with an error", blocked, err)
		}
		// Revoke surfaces the same reader failure and writes nothing.
		signHuman(t, p)
		before := eventCount(t, p, receipt.KindRevoke)
		if _, err := revoke.Revoke(opts(p), seams()); err == nil || !strings.Contains(err.Error(), "read escalation records") {
			t.Fatalf("revoke err = %v, want a reader failure", err)
		}
		if after := eventCount(t, p, receipt.KindRevoke); after != before {
			t.Errorf("revoke events %d → %d, want unchanged", before, after)
		}
	})
	t.Run("non-record-entries-skipped", func(t *testing.T) {
		p := signtest.New(t)
		rel := relTo(t, p.Root, escalation.RecordDir(p.Root, signtest.Card))
		p.WriteFile(rel+"/notes.txt", "not a record")
		p.WriteFile(rel+"/nested.md/inner.md", "not read")
		blocked, err := revoke.Blocked(p.Root, signtest.Card, signtest.SpecID, "seal")
		if blocked || err != nil {
			t.Fatalf("Blocked = %v, err %v; want false, nil", blocked, err)
		}
	})
}

func relTo(t *testing.T, root, path string) string {
	t.Helper()
	rel, err := filepath.Rel(root, path)
	if err != nil {
		t.Fatal(err)
	}
	return filepath.ToSlash(rel)
}

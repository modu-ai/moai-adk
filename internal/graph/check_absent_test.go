package graph

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/mx"
)

// AC-GFC-005 — every absent-verdict branch of checkCodemaps is preserved
// unchanged by the predicate work: an unjudgeable layer stays unjudgeable and
// never becomes fresh (REQ-GFC-011). All seven branches, each pinned to its
// pre-existing reason string; branch 7 additionally pins the (absent, non-nil
// error) pair shape, not the verdict alone.
func TestCheckCodemaps_Absent(t *testing.T) {
	th := Thresholds{CodemapsChangedFiles: 40}

	cases := []struct {
		name       string
		reason     string
		wantErr    bool
		setup      func(t *testing.T, root string)
		skipIfRoot bool
	}{
		{
			name:   "1 codemaps directory missing",
			reason: "codemaps directory missing",
			setup:  func(t *testing.T, root string) {},
		},
		{
			name:   "2 provenance file missing",
			reason: "no provenance block — freshness-unjudgeable, not fresh",
			setup: func(t *testing.T, root string) {
				if err := os.MkdirAll(filepath.Join(root, ".moai", "project", "codemaps"), 0o755); err != nil {
					t.Fatal(err)
				}
			},
		},
		{
			name:   "3 provenance unparseable",
			reason: "provenance block unparseable — freshness-unjudgeable",
			setup: func(t *testing.T, root string) {
				dir := filepath.Join(root, ".moai", "project", "codemaps")
				if err := os.MkdirAll(dir, 0o755); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(filepath.Join(dir, "provenance.json"), []byte("{not json"), 0o644); err != nil {
					t.Fatal(err)
				}
			},
		},
		{
			name:   "4 described root invalid",
			reason: `described root "../outside" invalid`,
			setup: func(t *testing.T, root string) {
				writeCodemapsProvenanceBlock(t, root, &mx.Provenance{
					SchemaVersion:  mx.ProvenanceSchemaVersion,
					TreeRoot:       root,
					CommitSHA:      "deadbeef",
					DescribedRoots: []string{"../outside"},
					GeneratedBy:    "codemaps-gen",
				})
			},
		},
		{
			name:       "5 dirty path, roots unreadable",
			reason:     "described roots unreadable:",
			skipIfRoot: true,
			setup: func(t *testing.T, root string) {
				writeCodemapsProvenanceBlock(t, root, &mx.Provenance{
					SchemaVersion:      mx.ProvenanceSchemaVersion,
					TreeRoot:           root,
					Dirty:              true,
					ContentFingerprint: strings.Repeat("a", 64),
					GeneratedBy:        "codemaps-gen",
				})
				locked := filepath.Join(root, "internal", "locked")
				if err := os.MkdirAll(locked, 0o755); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(filepath.Join(locked, "x.go"), []byte("package locked\n"), 0o644); err != nil {
					t.Fatal(err)
				}
				if err := os.Chmod(locked, 0o000); err != nil {
					t.Fatal(err)
				}
				t.Cleanup(func() { _ = os.Chmod(locked, 0o755) })
			},
		},
		{
			name:   "6 clean stamp carries no commit sha",
			reason: "clean stamp carries no commit sha — freshness-unjudgeable",
			setup: func(t *testing.T, root string) {
				writeCodemapsProvenanceBlock(t, root, &mx.Provenance{
					SchemaVersion: mx.ProvenanceSchemaVersion,
					TreeRoot:      root,
					GeneratedBy:   "codemaps-gen",
				})
			},
		},
		{
			name:    "7 stamped commit not comparable",
			reason:  "stamped commit not comparable (unmeasured, system error follows)",
			wantErr: true,
			setup: func(t *testing.T, root string) {
				writeCodemapsProvenance(t, root, strings.Repeat("0", 40))
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if c.skipIfRoot && os.Geteuid() == 0 {
				t.Skip("permission-denied branch is unreachable as root")
			}
			if c.skipIfRoot && runtime.GOOS == "windows" {
				t.Skip("posix permission bits are advisory on windows — the unreadable-roots fixture cannot be built there")
			}
			root := newCheckFixture(t)
			c.setup(t, root)

			rep, err := checkCodemaps(root, th)
			if rep.Verdict != VerdictAbsent {
				t.Fatalf("verdict = %q, want %q", rep.Verdict, VerdictAbsent)
			}
			if rep.Verdict == VerdictFresh {
				t.Fatalf("an unjudgeable layer reported fresh")
			}
			if !strings.Contains(rep.Reason, c.reason) {
				t.Errorf("reason = %q, want it to contain %q", rep.Reason, c.reason)
			}
			if c.wantErr && err == nil {
				t.Errorf("want a non-nil system error alongside the absent report, got nil")
			}
			if !c.wantErr && err != nil {
				t.Errorf("want a nil error, got %v", err)
			}
		})
	}
}

package backup

// retention_probe_p1_test.go — AC-UBR-01 (probe P1, SPEC-UPDATE-BACKUP-RETENTION-001):
// a user file the real deploy clean step copied into a backup run's pre-clean
// directory survives rotation. Covers REQ-UBR-003 and REQ-UBR-006.

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"testing"
	"testing/fstest"
	"time"

	"github.com/modu-ai/moai-adk/internal/cli/update/deploy"
)

func TestAC_UBR_01_PreCleanUserFileSurvivesRotation(t *testing.T) {
	root := t.TempDir()
	const userBody = "user notes\n"
	userSkill := filepath.Join(root, ".claude", "skills", "moai-user-notes", "SKILL.md")
	t1733WriteFile(t, userSkill, userBody)
	t1733WriteFile(t, filepath.Join(root, ".claude", "skills", "moai-foundation-core", "SKILL.md"), "tmpl\n")
	tmplFS := fstest.MapFS{".claude/skills/moai-foundation-core/SKILL.md": {Data: []byte("tmpl\n")}}
	if err := deploy.CleanMoaiManagedPaths(root, io.Discard, tmplFS); err != nil {
		t.Fatal(err)
	}

	// The producer contract this criterion consumes: the clean step removed the
	// user's file from the project and left its only copy under pre-clean/.
	if _, err := os.Lstat(userSkill); !os.IsNotExist(err) {
		t.Fatalf("control: the clean step left the user file in place (lstat err: %v)", err)
	}
	cfgRoot := filepath.Join(root, ".moai-backups")
	old := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	for _, name := range t1733DirNames(cfgRoot) {
		if err := os.Chtimes(filepath.Join(cfgRoot, name), old, old); err != nil {
			t.Fatal(err)
		}
	}

	count := func() int {
		n := 0
		_ = filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
			if err != nil || !d.Type().IsRegular() {
				return nil
			}
			if b, rerr := os.ReadFile(p); rerr == nil && string(b) == userBody {
				n++
			}
			return nil
		})
		return n
	}
	before := count()
	if before < 1 {
		t.Fatalf("control: the deploy clean step left no backup copy of the user file (copies=%d)", before)
	}

	// Five newer runs fill the window, so the run holding the user file is excess.
	for i := 1; i <= 5; i++ {
		dir := filepath.Join(cfgRoot, fmt.Sprintf("2026101%d_000000", i))
		t1733WriteFile(t, filepath.Join(dir, "x.yaml"), "x\n")
		ts := time.Now().Add(time.Duration(i) * time.Minute)
		if err := os.Chtimes(dir, ts, ts); err != nil {
			t.Fatal(err)
		}
	}

	deleted := CleanupOldBackups(root, 5)
	after := count()
	t.Logf("AC-UBR-01 user-file-copies-before=%d deleted=%d user-file-copies-after=%d", before, deleted, after)
	if after < 1 {
		t.Errorf("AC-UBR-01: no copy of the user file remains after the prune (before=%d after=%d); the pruner kept only the newest runs", before, after)
	}
}

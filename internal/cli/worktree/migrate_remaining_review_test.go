package worktree

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/hook"
	"github.com/modu-ai/moai-adk/internal/lockfile"
)

func TestMigrationInventoryIncludesProtectedRoots(t *testing.T) {
	w := sfxWorld(t)
	homeTree := filepath.Join(w.homeRoot, "already-home")
	customTree := filepath.Join(t.TempDir(), "operator-tree")
	codexTree := filepath.Join(w.home, ".codex", "worktrees", "managed")
	for i, tree := range []string{homeTree, customTree, codexTree} {
		if err := os.MkdirAll(filepath.Dir(tree), 0o755); err != nil {
			t.Fatal(err)
		}
		branch := []string{"release-home", "WT-operator", "WT-codex"}[i]
		m6Git(t, w.repo, "worktree", "add", "-q", "-b", branch, "--", tree)
	}
	plans := planMigrationFor(w.repo, time.Now(), false)
	byPath := map[string]migratePlan{}
	for _, plan := range plans {
		byPath[plan.source] = plan
	}
	for path, class := range map[string]TreeClass{
		w.repo: ClassPrimary, homeTree: ClassMoaiMade,
		customTree: ClassMoaiMade, codexTree: ClassCodexManaged,
	} {
		plan, ok := byPath[path]
		if !ok || plan.class != string(class) {
			t.Errorf("registered tree %s absent or misclassified: %+v, present=%v, want %s", path, plan, ok, class)
		}
		if ok && plan.class == string(ClassMoaiMade) && plan.skip == "" {
			t.Errorf("inventory expansion selected a non-legacy tree for movement: %+v", plan)
		}
	}
}

func TestMigrationRecordsSkippedOutcomesInAudit(t *testing.T) {
	w := sfxWorld(t)
	m4InRepo(t, w.repo)
	if out, err := m4RunMigrate(t, "--apply"); err != nil {
		t.Fatalf("apply: %v\n%s", err, out)
	}
	data, err := os.ReadFile(moveListPath(w.repo))
	if err != nil {
		t.Fatal(err)
	}
	for _, line := range strings.Split(string(data), "\n") {
		var record struct {
			OldPath string `json:"old_path"`
			Result  string `json:"result"`
			Reason  string `json:"reason"`
		}
		if json.Unmarshal([]byte(line), &record) == nil && record.OldPath == w.lockedTree && record.Result == "skipped" && strings.Contains(record.Reason, "locked") {
			return
		}
	}
	t.Fatalf("locked-tree skip has no durable audit outcome:\n%s", data)
}

func TestMigrationRetiresOnlyAnEmptyMoaiLegacyDirectory(t *testing.T) {
	repo, _, _ := rfxLegacyWorld(t)
	m4InRepo(t, repo)
	claudeRoot := filepath.Join(repo, ".claude", "worktrees")
	if err := os.MkdirAll(claudeRoot, 0o755); err != nil {
		t.Fatal(err)
	}
	if out, err := m4RunMigrate(t, "--apply"); err != nil {
		t.Fatalf("apply: %v\n%s", err, out)
	}
	if _, err := os.Lstat(filepath.Join(repo, ".moai", "worktrees")); !os.IsNotExist(err) {
		t.Errorf("empty MoAI legacy directory was not retired: %v", err)
	}
	if _, err := os.Stat(claudeRoot); err != nil {
		t.Fatalf("Claude-owned legacy directory must remain: %v", err)
	}
}

func TestMigrationManifestStreamsIgnoredFileContents(t *testing.T) {
	_, tree, _ := rfxLegacyWorld(t)
	const size = 32 << 20
	path := filepath.Join(tree, "ignored-payload.bin")
	if err := os.WriteFile(filepath.Join(tree, ".gitignore"), []byte("ignored-payload.bin\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := f.Truncate(size); err != nil {
		_ = f.Close()
		t.Fatal(err)
	}
	_ = f.Close()
	runtime.GC()
	var before, after runtime.MemStats
	runtime.ReadMemStats(&before)
	digest := captureManifest(tree)
	runtime.ReadMemStats(&after)
	allocated := after.TotalAlloc - before.TotalAlloc
	t.Logf("ignored file bytes=%d allocation bytes=%d", size, allocated)
	if allocated >= size/2 {
		t.Errorf("manifest allocated %d bytes for a %d-byte ignored file; content hashing must use bounded buffers", allocated, size)
	}
	f, err = os.OpenFile(path, os.O_WRONLY, 0)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := f.WriteAt([]byte("mutation"), size-16); err != nil {
		_ = f.Close()
		t.Fatal(err)
	}
	_ = f.Close()
	if digest == captureManifest(tree) {
		t.Fatal("streaming must continue hashing ignored content through the file's end")
	}
}

func TestMigrationStopsAfterStateLockLapseBudget(t *testing.T) {
	w := sfxWorld(t)
	m4InRepo(t, w.repo)
	if err := hook.RegisterLauncherWorktreeE(w.repo, w.workerTree, "WT-worker", "test"); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(w.repo, ".moai", "state", "worktrees.json.lock")
	lock, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0o600)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = lock.Close() }()
	if err := lockfile.Lock(lock); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = lockfile.Unlock(lock) }()
	if out, err := m4RunMigrate(t, "--apply"); err == nil {
		t.Fatalf("held state-file lock must fail the migration:\n%s", out)
	}
	raw, err := os.ReadFile(moveListPath(w.repo))
	if err != nil {
		t.Fatal(err)
	}
	started := map[string]bool{}
	for _, line := range strings.Split(string(raw), "\n") {
		var record struct {
			OldPath string `json:"old_path"`
			NewPath string `json:"new_path"`
			Result  string `json:"result"`
		}
		if json.Unmarshal([]byte(line), &record) != nil {
			continue
		}
		if record.Result == "pending" && underPath(record.NewPath, w.homeRoot) {
			started[record.OldPath] = true
		}
	}
	t.Logf("forward attempts=%d state-lock lapse budget=%d", len(started), hook.StateLockLapseBudget)
	if len(started) != hook.StateLockLapseBudget {
		t.Fatalf("migration continued starting trees after state-lock budget lapsed: %d trees (budget %d)\n%s", len(started), hook.StateLockLapseBudget, raw)
	}
}

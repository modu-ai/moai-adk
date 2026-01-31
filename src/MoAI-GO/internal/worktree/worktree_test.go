package worktree

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestNewWorktreeInfo(t *testing.T) {
	info := NewWorktreeInfo("SPEC-AUTH-001", "/tmp/wt", "feature/SPEC-AUTH-001")
	if info.SpecID != "SPEC-AUTH-001" {
		t.Errorf("expected SpecID SPEC-AUTH-001, got %s", info.SpecID)
	}
	if info.Status != "active" {
		t.Errorf("expected status 'active', got %s", info.Status)
	}
	if info.CreatedAt == "" {
		t.Error("expected CreatedAt to be set")
	}
}

func TestRegistry_RegisterAndGet(t *testing.T) {
	dir := t.TempDir()
	reg, err := NewRegistry(dir, "test-project")
	if err != nil {
		t.Fatal(err)
	}

	info := NewWorktreeInfo("SPEC-001", "/tmp/wt/SPEC-001", "feature/SPEC-001")
	if err := reg.Register(info); err != nil {
		t.Fatalf("Register failed: %v", err)
	}

	got := reg.Get("SPEC-001")
	if got == nil {
		t.Fatal("expected to find SPEC-001")
	}
	if got.Branch != "feature/SPEC-001" {
		t.Errorf("expected branch feature/SPEC-001, got %s", got.Branch)
	}
}

func TestRegistry_Unregister(t *testing.T) {
	dir := t.TempDir()
	reg, err := NewRegistry(dir, "test-project")
	if err != nil {
		t.Fatal(err)
	}

	info := NewWorktreeInfo("SPEC-002", "/tmp/wt/SPEC-002", "feature/SPEC-002")
	if err := reg.Register(info); err != nil {
		t.Fatal(err)
	}

	if err := reg.Unregister("SPEC-002"); err != nil {
		t.Fatalf("Unregister failed: %v", err)
	}

	if reg.Get("SPEC-002") != nil {
		t.Error("expected SPEC-002 to be removed")
	}
}

func TestRegistry_List(t *testing.T) {
	dir := t.TempDir()
	reg, err := NewRegistry(dir, "test-project")
	if err != nil {
		t.Fatal(err)
	}

	_ = reg.Register(NewWorktreeInfo("SPEC-A", "/a", "branch-a"))
	_ = reg.Register(NewWorktreeInfo("SPEC-B", "/b", "branch-b"))

	items := reg.List()
	if len(items) != 2 {
		t.Errorf("expected 2 items, got %d", len(items))
	}
}

func TestRegistry_Persistence(t *testing.T) {
	dir := t.TempDir()

	// Write with first instance
	reg1, err := NewRegistry(dir, "test-project")
	if err != nil {
		t.Fatal(err)
	}
	_ = reg1.Register(NewWorktreeInfo("SPEC-P", "/p", "branch-p"))

	// Read with second instance
	reg2, err := NewRegistry(dir, "test-project")
	if err != nil {
		t.Fatal(err)
	}
	got := reg2.Get("SPEC-P")
	if got == nil {
		t.Fatal("expected to find SPEC-P after reload")
	}
	if got.Branch != "branch-p" {
		t.Errorf("expected branch-p, got %s", got.Branch)
	}
}

func TestRegistry_GetCrossProject(t *testing.T) {
	dir := t.TempDir()
	reg, err := NewRegistry(dir, "project-a")
	if err != nil {
		t.Fatal(err)
	}

	// Manually add to a different project
	reg.mu.Lock()
	reg.data["project-b"] = map[string]*WorktreeInfo{
		"SPEC-X": NewWorktreeInfo("SPEC-X", "/x", "branch-x"),
	}
	_ = reg.save()
	reg.mu.Unlock()

	// Should find cross-project
	got := reg.Get("SPEC-X")
	if got == nil {
		t.Fatal("expected to find SPEC-X from project-b")
	}
}

// Integration test: requires git
func TestManager_CreateAndRemove(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}

	// Create a temporary git repo
	repoDir := t.TempDir()
	initGitRepo(t, repoDir)

	// Override HOME to isolate worktree root
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	mgr, err := NewManager(repoDir)
	if err != nil {
		t.Fatalf("NewManager failed: %v", err)
	}

	// Create worktree
	info, err := mgr.Create("SPEC-TEST-001", "", "main", false)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	if info.SpecID != "SPEC-TEST-001" {
		t.Errorf("expected SPEC-TEST-001, got %s", info.SpecID)
	}
	if info.Branch != "feature/SPEC-TEST-001" {
		t.Errorf("expected branch feature/SPEC-TEST-001, got %s", info.Branch)
	}

	// Verify path exists
	if _, statErr := os.Stat(info.Path); os.IsNotExist(statErr) {
		t.Error("worktree path should exist")
	}

	// List
	items := mgr.List()
	if len(items) != 1 {
		t.Errorf("expected 1 worktree, got %d", len(items))
	}

	// Remove
	if removeErr := mgr.Remove("SPEC-TEST-001", true); removeErr != nil {
		t.Fatalf("Remove failed: %v", removeErr)
	}

	// Verify removed
	items = mgr.List()
	if len(items) != 0 {
		t.Errorf("expected 0 worktrees after remove, got %d", len(items))
	}
}

func TestManager_Recover(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}

	repoDir := t.TempDir()
	initGitRepo(t, repoDir)

	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	mgr, err := NewManager(repoDir)
	if err != nil {
		t.Fatal(err)
	}

	// Create a worktree
	info, err := mgr.Create("SPEC-REC-001", "", "main", false)
	if err != nil {
		t.Fatal(err)
	}

	// Clear registry (simulate lost registry)
	_ = mgr.Registry.Unregister("SPEC-REC-001")
	if mgr.Registry.Get("SPEC-REC-001") != nil {
		t.Fatal("expected registry to be empty")
	}

	// Verify the worktree directory still exists
	if _, statErr := os.Stat(info.Path); os.IsNotExist(statErr) {
		t.Fatal("worktree directory should still exist")
	}

	// Recover
	count, recoverErr := mgr.Recover()
	if recoverErr != nil {
		t.Fatalf("Recover failed: %v", recoverErr)
	}
	if count != 1 {
		t.Errorf("expected 1 recovered, got %d", count)
	}

	recovered := mgr.Registry.Get("SPEC-REC-001")
	if recovered == nil {
		t.Fatal("expected SPEC-REC-001 to be recovered")
	}
	if recovered.Status != "recovered" {
		t.Errorf("expected status 'recovered', got %s", recovered.Status)
	}
}

// initGitRepo creates a minimal git repo with an initial commit
func initGitRepo(t *testing.T, dir string) {
	t.Helper()
	cmds := [][]string{
		{"git", "init"},
		{"git", "config", "user.email", "test@test.com"},
		{"git", "config", "user.name", "Test"},
		{"git", "checkout", "-b", "main"},
	}
	for _, args := range cmds {
		cmd := exec.Command(args[0], args[1:]...)
		cmd.Dir = dir
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git command %v failed: %v\n%s", args, err, out)
		}
	}

	// Create initial commit
	readmePath := filepath.Join(dir, "README.md")
	if err := os.WriteFile(readmePath, []byte("# Test\n"), 0644); err != nil {
		t.Fatal(err)
	}
	for _, args := range [][]string{
		{"git", "add", "."},
		{"git", "commit", "-m", "initial commit"},
	} {
		cmd := exec.Command(args[0], args[1:]...)
		cmd.Dir = dir
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git command %v failed: %v\n%s", args, err, out)
		}
	}
}

// runGit runs a git command in a directory and fails the test on error
func runGit(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %v failed: %v\n%s", args, err, out)
	}
}

// initGitRepoWithRemote creates a git repo with a bare "origin" remote
func initGitRepoWithRemote(t *testing.T) string {
	t.Helper()
	bareDir := t.TempDir()
	runGit(t, bareDir, "init", "--bare")

	repoDir := t.TempDir()
	initGitRepo(t, repoDir)
	runGit(t, repoDir, "remote", "add", "origin", bareDir)
	runGit(t, repoDir, "push", "-u", "origin", "main")

	return repoDir
}

// ---------- ListAll ----------

func TestRegistry_ListAll(t *testing.T) {
	dir := t.TempDir()
	reg, err := NewRegistry(dir, "project-a")
	if err != nil {
		t.Fatal(err)
	}

	// Register entries in project-a (current project)
	if err := reg.Register(NewWorktreeInfo("SPEC-1", "/a/1", "branch-a1")); err != nil {
		t.Fatal(err)
	}
	if err := reg.Register(NewWorktreeInfo("SPEC-2", "/a/2", "branch-a2")); err != nil {
		t.Fatal(err)
	}

	// Manually add entries for project-b
	reg.mu.Lock()
	reg.data["project-b"] = map[string]*WorktreeInfo{
		"SPEC-3": NewWorktreeInfo("SPEC-3", "/b/3", "branch-b3"),
	}
	if saveErr := reg.save(); saveErr != nil {
		reg.mu.Unlock()
		t.Fatal(saveErr)
	}
	reg.mu.Unlock()

	all := reg.ListAll()
	if len(all) != 3 {
		t.Errorf("expected 3 items from ListAll, got %d", len(all))
	}

	// Verify List returns only current project entries
	current := reg.List()
	if len(current) != 2 {
		t.Errorf("expected 2 items from List (current project), got %d", len(current))
	}
}

func TestRegistry_ListAll_Empty(t *testing.T) {
	dir := t.TempDir()
	reg, err := NewRegistry(dir, "empty-project")
	if err != nil {
		t.Fatal(err)
	}

	all := reg.ListAll()
	if len(all) != 0 {
		t.Errorf("expected 0 items from ListAll on empty registry, got %d", len(all))
	}
}

// ---------- getBranchName ----------

func TestGetBranchName_NonGitDir(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}

	dir := t.TempDir()
	branch := getBranchName(dir)
	if branch != "unknown" {
		t.Errorf("expected 'unknown' for non-git directory, got %s", branch)
	}
}

func TestGetBranchName_ValidRepo(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}

	dir := t.TempDir()
	initGitRepo(t, dir)
	branch := getBranchName(dir)
	if branch != "main" {
		t.Errorf("expected 'main', got %s", branch)
	}
}

func TestGetBranchName_CustomBranch(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}

	dir := t.TempDir()
	initGitRepo(t, dir)
	runGit(t, dir, "checkout", "-b", "feature/test-branch")

	branch := getBranchName(dir)
	if branch != "feature/test-branch" {
		t.Errorf("expected 'feature/test-branch', got %s", branch)
	}
}

// ---------- Create edge cases ----------

func TestManager_CreateWithCustomBranch(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}

	repoDir := t.TempDir()
	initGitRepo(t, repoDir)
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	mgr, err := NewManager(repoDir)
	if err != nil {
		t.Fatal(err)
	}

	info, err := mgr.Create("SPEC-CUSTOM-001", "my-custom-branch", "main", false)
	if err != nil {
		t.Fatal(err)
	}

	if info.Branch != "my-custom-branch" {
		t.Errorf("expected branch 'my-custom-branch', got %s", info.Branch)
	}

	// Verify the branch name via git
	actual := getBranchName(info.Path)
	if actual != "my-custom-branch" {
		t.Errorf("expected git branch 'my-custom-branch', got %s", actual)
	}
}

func TestManager_CreateDefaultBaseBranch(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}

	repoDir := t.TempDir()
	initGitRepo(t, repoDir)
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	mgr, err := NewManager(repoDir)
	if err != nil {
		t.Fatal(err)
	}

	// Empty baseBranch defaults to "main"
	info, err := mgr.Create("SPEC-DEFBASE", "", "", false)
	if err != nil {
		t.Fatal(err)
	}

	if info.Branch != "feature/SPEC-DEFBASE" {
		t.Errorf("expected default branch 'feature/SPEC-DEFBASE', got %s", info.Branch)
	}
}

func TestManager_CreateWithForce(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}

	repoDir := t.TempDir()
	initGitRepo(t, repoDir)
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	mgr, err := NewManager(repoDir)
	if err != nil {
		t.Fatal(err)
	}

	// Register a stale entry (simulating a worktree that was removed externally)
	fakeInfo := NewWorktreeInfo("SPEC-FORCE", "/nonexistent", "old-branch")
	if err := mgr.Registry.Register(fakeInfo); err != nil {
		t.Fatal(err)
	}

	// Without force should fail because entry already exists
	_, err = mgr.Create("SPEC-FORCE", "", "main", false)
	if err == nil {
		t.Error("expected error creating without force when entry exists")
	}

	// With force should succeed (bypasses registry check)
	info, err := mgr.Create("SPEC-FORCE", "", "main", true)
	if err != nil {
		t.Fatalf("Create with force failed: %v", err)
	}
	if info.Branch != "feature/SPEC-FORCE" {
		t.Errorf("expected branch 'feature/SPEC-FORCE', got %s", info.Branch)
	}
	if _, statErr := os.Stat(info.Path); os.IsNotExist(statErr) {
		t.Error("expected worktree path to exist after force create")
	}
}

// ---------- Remove edge cases ----------

func TestManager_RemoveNonExistent(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}

	repoDir := t.TempDir()
	initGitRepo(t, repoDir)
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	mgr, err := NewManager(repoDir)
	if err != nil {
		t.Fatal(err)
	}

	err = mgr.Remove("SPEC-NONEXISTENT", false)
	if err == nil {
		t.Error("expected error removing non-existent worktree")
	}
	if !strings.Contains(err.Error(), "worktree not found") {
		t.Errorf("expected 'worktree not found' error, got: %v", err)
	}
}

// ---------- Done ----------

func TestManager_Done(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}

	repoDir := t.TempDir()
	initGitRepo(t, repoDir)
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	mgr, err := NewManager(repoDir)
	if err != nil {
		t.Fatal(err)
	}

	// Create worktree
	info, err := mgr.Create("SPEC-DONE-001", "", "main", false)
	if err != nil {
		t.Fatal(err)
	}

	// Make a commit in the worktree
	featureFile := filepath.Join(info.Path, "feature.txt")
	if err := os.WriteFile(featureFile, []byte("new feature\n"), 0644); err != nil {
		t.Fatal(err)
	}
	runGit(t, info.Path, "add", ".")
	runGit(t, info.Path, "commit", "-m", "add feature")

	wtPath := info.Path

	// Done: merge into main and remove (no push)
	result, err := mgr.Done("SPEC-DONE-001", "main", false, true)
	if err != nil {
		t.Fatalf("Done failed: %v", err)
	}

	if result["merged_branch"] != "feature/SPEC-DONE-001" {
		t.Errorf("expected merged_branch 'feature/SPEC-DONE-001', got %s", result["merged_branch"])
	}
	if result["base_branch"] != "main" {
		t.Errorf("expected base_branch 'main', got %s", result["base_branch"])
	}
	if result["pushed"] != "false" {
		t.Errorf("expected pushed 'false', got %s", result["pushed"])
	}

	// Verify worktree is removed from registry
	if mgr.Registry.Get("SPEC-DONE-001") != nil {
		t.Error("expected worktree to be unregistered after Done")
	}

	// Verify the merge brought the feature file into main
	mergedFile := filepath.Join(repoDir, "feature.txt")
	if _, statErr := os.Stat(mergedFile); os.IsNotExist(statErr) {
		t.Error("expected feature.txt to exist in main repo after merge")
	}

	// Verify worktree directory is removed
	if _, statErr := os.Stat(wtPath); !os.IsNotExist(statErr) {
		t.Error("expected worktree directory to be removed after Done")
	}
}

func TestManager_DoneDefaultBaseBranch(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}

	repoDir := t.TempDir()
	initGitRepo(t, repoDir)
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	mgr, err := NewManager(repoDir)
	if err != nil {
		t.Fatal(err)
	}

	info, err := mgr.Create("SPEC-DONEDEF", "", "main", false)
	if err != nil {
		t.Fatal(err)
	}

	// Add a commit so the merge is meaningful
	newFile := filepath.Join(info.Path, "done-default.txt")
	if err := os.WriteFile(newFile, []byte("content\n"), 0644); err != nil {
		t.Fatal(err)
	}
	runGit(t, info.Path, "add", ".")
	runGit(t, info.Path, "commit", "-m", "default base branch test")

	// Empty baseBranch defaults to "main"
	result, err := mgr.Done("SPEC-DONEDEF", "", false, true)
	if err != nil {
		t.Fatalf("Done with empty baseBranch failed: %v", err)
	}
	if result["base_branch"] != "main" {
		t.Errorf("expected base_branch 'main', got %s", result["base_branch"])
	}
}

func TestManager_DoneNotFound(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}

	repoDir := t.TempDir()
	initGitRepo(t, repoDir)
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	mgr, err := NewManager(repoDir)
	if err != nil {
		t.Fatal(err)
	}

	_, err = mgr.Done("SPEC-NONEXISTENT", "main", false, true)
	if err == nil {
		t.Error("expected error for non-existent specID in Done")
	}
	if !strings.Contains(err.Error(), "worktree not found") {
		t.Errorf("expected 'worktree not found' error, got: %v", err)
	}
}

// ---------- Sync ----------

func TestManager_Sync(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}

	repoDir := initGitRepoWithRemote(t)
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	mgr, err := NewManager(repoDir)
	if err != nil {
		t.Fatal(err)
	}

	// Create worktree
	info, err := mgr.Create("SPEC-SYNC-001", "", "main", false)
	if err != nil {
		t.Fatal(err)
	}

	// Add a commit to main and push to origin
	updateFile := filepath.Join(repoDir, "update.txt")
	if err := os.WriteFile(updateFile, []byte("main update\n"), 0644); err != nil {
		t.Fatal(err)
	}
	runGit(t, repoDir, "add", ".")
	runGit(t, repoDir, "commit", "-m", "update on main")
	runGit(t, repoDir, "push", "origin", "main")

	// Sync worktree (merge mode)
	if err := mgr.Sync("SPEC-SYNC-001", "main", false); err != nil {
		t.Fatalf("Sync failed: %v", err)
	}

	// Verify the synced file exists in the worktree
	syncedFile := filepath.Join(info.Path, "update.txt")
	if _, statErr := os.Stat(syncedFile); os.IsNotExist(statErr) {
		t.Error("expected update.txt to exist in worktree after sync")
	}
}

func TestManager_SyncDefaultBaseBranch(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}

	repoDir := initGitRepoWithRemote(t)
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	mgr, err := NewManager(repoDir)
	if err != nil {
		t.Fatal(err)
	}

	info, err := mgr.Create("SPEC-SYNCDEF", "", "main", false)
	if err != nil {
		t.Fatal(err)
	}

	// Add a commit to main and push
	newFile := filepath.Join(repoDir, "sync-default.txt")
	if err := os.WriteFile(newFile, []byte("sync default\n"), 0644); err != nil {
		t.Fatal(err)
	}
	runGit(t, repoDir, "add", ".")
	runGit(t, repoDir, "commit", "-m", "sync default test")
	runGit(t, repoDir, "push", "origin", "main")

	// Empty baseBranch defaults to "main"
	if err := mgr.Sync("SPEC-SYNCDEF", "", false); err != nil {
		t.Fatalf("Sync with empty baseBranch failed: %v", err)
	}

	// Verify the file arrived in the worktree
	syncedFile := filepath.Join(info.Path, "sync-default.txt")
	if _, statErr := os.Stat(syncedFile); os.IsNotExist(statErr) {
		t.Error("expected sync-default.txt in worktree after sync")
	}
}

func TestManager_SyncNotFound(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}

	repoDir := t.TempDir()
	initGitRepo(t, repoDir)
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	mgr, err := NewManager(repoDir)
	if err != nil {
		t.Fatal(err)
	}

	err = mgr.Sync("SPEC-NONEXISTENT", "main", false)
	if err == nil {
		t.Error("expected error for non-existent specID in Sync")
	}
	if !strings.Contains(err.Error(), "worktree not found") {
		t.Errorf("expected 'worktree not found' error, got: %v", err)
	}
}

// ---------- CleanMerged ----------

func TestManager_CleanMerged(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}

	repoDir := t.TempDir()
	initGitRepo(t, repoDir)
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	mgr, err := NewManager(repoDir)
	if err != nil {
		t.Fatal(err)
	}

	// Create a feature branch directly (not via worktree) to avoid
	// newer git versions adding a "+" prefix for worktree-checked-out
	// branches in "git branch --merged" output.
	runGit(t, repoDir, "checkout", "-b", "feature/SPEC-CLEAN-001")
	cleanFile := filepath.Join(repoDir, "clean-feature.txt")
	if err := os.WriteFile(cleanFile, []byte("clean feature\n"), 0644); err != nil {
		t.Fatal(err)
	}
	runGit(t, repoDir, "add", ".")
	runGit(t, repoDir, "commit", "-m", "add clean feature")
	runGit(t, repoDir, "checkout", "main")
	runGit(t, repoDir, "merge", "feature/SPEC-CLEAN-001", "--no-ff", "-m", "merge clean feature")

	// Register a simulated worktree entry for this branch
	fakeInfo := &WorktreeInfo{
		SpecID: "SPEC-CLEAN-001",
		Path:   filepath.Join(mgr.WorktreeRoot, "SPEC-CLEAN-001"),
		Branch: "feature/SPEC-CLEAN-001",
		Status: "active",
	}
	if err := mgr.Registry.Register(fakeInfo); err != nil {
		t.Fatal(err)
	}

	// Clean merged worktrees
	count, err := mgr.CleanMerged("main")
	if err != nil {
		t.Fatalf("CleanMerged failed: %v", err)
	}
	if count != 1 {
		t.Errorf("expected 1 cleaned worktree, got %d", count)
	}

	// Verify worktree is removed from registry
	if mgr.Registry.Get("SPEC-CLEAN-001") != nil {
		t.Error("expected worktree to be unregistered after CleanMerged")
	}
}

func TestManager_CleanMergedDefaultBaseBranch(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}

	repoDir := t.TempDir()
	initGitRepo(t, repoDir)
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	mgr, err := NewManager(repoDir)
	if err != nil {
		t.Fatal(err)
	}

	// No worktrees registered, CleanMerged with empty baseBranch
	count, err := mgr.CleanMerged("")
	if err != nil {
		t.Fatalf("CleanMerged with empty baseBranch failed: %v", err)
	}
	if count != 0 {
		t.Errorf("expected 0 cleaned (nothing registered), got %d", count)
	}
}

func TestManager_CleanMerged_Unmerged(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}

	repoDir := t.TempDir()
	initGitRepo(t, repoDir)
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	mgr, err := NewManager(repoDir)
	if err != nil {
		t.Fatal(err)
	}

	// Create worktree with a commit (but do NOT merge into main)
	info, err := mgr.Create("SPEC-UNMERGED", "", "main", false)
	if err != nil {
		t.Fatal(err)
	}

	unmergedFile := filepath.Join(info.Path, "unmerged.txt")
	if err := os.WriteFile(unmergedFile, []byte("not merged\n"), 0644); err != nil {
		t.Fatal(err)
	}
	runGit(t, info.Path, "add", ".")
	runGit(t, info.Path, "commit", "-m", "unmerged commit")

	// CleanMerged should not remove this worktree
	count, err := mgr.CleanMerged("main")
	if err != nil {
		t.Fatalf("CleanMerged failed: %v", err)
	}
	if count != 0 {
		t.Errorf("expected 0 cleaned (branch not merged), got %d", count)
	}

	// Verify worktree still exists
	if mgr.Registry.Get("SPEC-UNMERGED") == nil {
		t.Error("expected unmerged worktree to still be registered")
	}
}

// ---------- Registry save error ----------

func TestRegistry_SaveError(t *testing.T) {
	dir := t.TempDir()
	reg, err := NewRegistry(dir, "test-project")
	if err != nil {
		t.Fatal(err)
	}

	// Point registry path to an impossible location (MkdirAll will fail)
	reg.path = filepath.Join("/dev/null", "impossible", "path.json")

	info := NewWorktreeInfo("SPEC-ERR", "/err", "branch-err")
	err = reg.Register(info)
	if err == nil {
		t.Error("expected error when saving to unwritable path")
	}
}

func TestRegistry_SaveError_Unregister(t *testing.T) {
	dir := t.TempDir()
	reg, err := NewRegistry(dir, "test-project")
	if err != nil {
		t.Fatal(err)
	}

	// Register an entry first (to a valid path)
	if err := reg.Register(NewWorktreeInfo("SPEC-SAVE", "/s", "branch-s")); err != nil {
		t.Fatal(err)
	}

	// Now break the path so Unregister's save fails
	reg.path = filepath.Join("/dev/null", "impossible", "path.json")

	err = reg.Unregister("SPEC-SAVE")
	if err == nil {
		t.Error("expected error when Unregister cannot save")
	}
}

// ---------- Create: existing branch fallback path ----------

func TestManager_CreateExistingBranch(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}

	repoDir := t.TempDir()
	initGitRepo(t, repoDir)
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	// Pre-create the branch so that "git worktree add -b" fails
	// and the fallback "git worktree add <path> <branch>" is used.
	runGit(t, repoDir, "branch", "feature/SPEC-EXIST")

	mgr, err := NewManager(repoDir)
	if err != nil {
		t.Fatal(err)
	}

	info, err := mgr.Create("SPEC-EXIST", "feature/SPEC-EXIST", "main", false)
	if err != nil {
		t.Fatal(err)
	}
	if info.Branch != "feature/SPEC-EXIST" {
		t.Errorf("expected branch 'feature/SPEC-EXIST', got %s", info.Branch)
	}
	if _, statErr := os.Stat(info.Path); os.IsNotExist(statErr) {
		t.Error("expected worktree path to exist")
	}
}

// ---------- Done: push path ----------

func TestManager_DoneWithPush(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}

	repoDir := initGitRepoWithRemote(t)
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	mgr, err := NewManager(repoDir)
	if err != nil {
		t.Fatal(err)
	}

	info, err := mgr.Create("SPEC-PUSH", "", "main", false)
	if err != nil {
		t.Fatal(err)
	}

	// Add a commit in the worktree
	featureFile := filepath.Join(info.Path, "push-feature.txt")
	if err := os.WriteFile(featureFile, []byte("push feature\n"), 0644); err != nil {
		t.Fatal(err)
	}
	runGit(t, info.Path, "add", ".")
	runGit(t, info.Path, "commit", "-m", "push feature")

	// Done with push=true
	result, err := mgr.Done("SPEC-PUSH", "main", true, true)
	if err != nil {
		t.Fatalf("Done with push failed: %v", err)
	}
	if result["pushed"] != "true" {
		t.Errorf("expected pushed 'true', got %s", result["pushed"])
	}
	if result["merged_branch"] != "feature/SPEC-PUSH" {
		t.Errorf("expected merged_branch 'feature/SPEC-PUSH', got %s", result["merged_branch"])
	}
}

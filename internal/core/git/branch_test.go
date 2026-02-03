package git

import (
	"errors"
	"testing"
)

func TestBranchCreate_Success(t *testing.T) {
	dir := initTestRepo(t)
	bm := NewBranchManager(dir)

	if err := bm.Create("feature/login"); err != nil {
		t.Fatalf("Create() error: %v", err)
	}

	// Verify the branch appears in List.
	branches, err := bm.List()
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, b := range branches {
		if b.Name == "feature/login" {
			found = true
			break
		}
	}
	if !found {
		t.Error("created branch not found in List()")
	}
}

func TestBranchCreate_AlreadyExists(t *testing.T) {
	dir := initTestRepo(t)
	bm := NewBranchManager(dir)

	if err := bm.Create("feature/login"); err != nil {
		t.Fatal(err)
	}

	err := bm.Create("feature/login")
	if err == nil {
		t.Fatal("Create() duplicate should return error")
	}
	if !errors.Is(err, ErrBranchExists) {
		t.Errorf("error = %v, want ErrBranchExists", err)
	}
}

func TestBranchCreate_InvalidName(t *testing.T) {
	dir := initTestRepo(t)
	bm := NewBranchManager(dir)

	tests := []struct {
		name   string
		reason string
	}{
		{"feat..ure", "contains double dots (..)"},
		{"feat~1", "contains tilde (~)"},
		{"feat^2", "contains caret (^)"},
		{"feat:name", "contains colon (:)"},
		{"feat name", "contains space"},
		{"", "empty string"},
		{".feat", "starts with dot"},
		{"feat.lock", "ends with .lock"},
	}

	for _, tt := range tests {
		t.Run(tt.reason, func(t *testing.T) {
			err := bm.Create(tt.name)
			if err == nil {
				t.Errorf("Create(%q) should return error for: %s", tt.name, tt.reason)
			}
			if !errors.Is(err, ErrInvalidBranchName) {
				t.Errorf("Create(%q) error = %v, want ErrInvalidBranchName", tt.name, err)
			}
		})
	}
}

func TestBranchSwitch_Success(t *testing.T) {
	dir := initTestRepo(t)
	bm := NewBranchManager(dir)

	if err := bm.Create("feature/login"); err != nil {
		t.Fatal(err)
	}

	if err := bm.Switch("feature/login"); err != nil {
		t.Fatalf("Switch() error: %v", err)
	}

	// Verify via NewRepository.
	repo, err := NewRepository(dir)
	if err != nil {
		t.Fatal(err)
	}
	current, err := repo.CurrentBranch()
	if err != nil {
		t.Fatal(err)
	}
	if current != "feature/login" {
		t.Errorf("CurrentBranch() = %q, want %q", current, "feature/login")
	}
}

func TestBranchSwitch_NotFound(t *testing.T) {
	dir := initTestRepo(t)
	bm := NewBranchManager(dir)

	err := bm.Switch("nonexistent")
	if err == nil {
		t.Fatal("Switch() to nonexistent branch should return error")
	}
	if !errors.Is(err, ErrBranchNotFound) {
		t.Errorf("error = %v, want ErrBranchNotFound", err)
	}
}

func TestBranchSwitch_DirtyWorkingTree(t *testing.T) {
	dir := initTestRepo(t)
	bm := NewBranchManager(dir)

	if err := bm.Create("feature/login"); err != nil {
		t.Fatal(err)
	}

	// Make the working tree dirty.
	writeTestFile(t, dir+"/dirty.txt", "dirty content\n")

	err := bm.Switch("feature/login")
	if err == nil {
		t.Fatal("Switch() with dirty tree should return error")
	}
	if !errors.Is(err, ErrDirtyWorkingTree) {
		t.Errorf("error = %v, want ErrDirtyWorkingTree", err)
	}

	// Verify we stayed on the original branch.
	repo, err := NewRepository(dir)
	if err != nil {
		t.Fatal(err)
	}
	current, err := repo.CurrentBranch()
	if err != nil {
		t.Fatal(err)
	}
	if current != "main" {
		t.Errorf("CurrentBranch() = %q, want %q (unchanged)", current, "main")
	}
}

func TestBranchDelete_Success(t *testing.T) {
	dir := initTestRepo(t)
	bm := NewBranchManager(dir)

	if err := bm.Create("feature/old"); err != nil {
		t.Fatal(err)
	}

	if err := bm.Delete("feature/old"); err != nil {
		t.Fatalf("Delete() error: %v", err)
	}

	// Verify the branch is gone.
	branches, err := bm.List()
	if err != nil {
		t.Fatal(err)
	}
	for _, b := range branches {
		if b.Name == "feature/old" {
			t.Error("deleted branch still appears in List()")
		}
	}
}

func TestBranchDelete_CurrentBranch(t *testing.T) {
	dir := initTestRepo(t)
	bm := NewBranchManager(dir)

	err := bm.Delete("main")
	if err == nil {
		t.Fatal("Delete() on current branch should return error")
	}
	if !errors.Is(err, ErrCannotDeleteCurrentBranch) {
		t.Errorf("error = %v, want ErrCannotDeleteCurrentBranch", err)
	}
}

func TestBranchDelete_NotFound(t *testing.T) {
	dir := initTestRepo(t)
	bm := NewBranchManager(dir)

	err := bm.Delete("nonexistent")
	if err == nil {
		t.Fatal("Delete() on nonexistent branch should return error")
	}
	if !errors.Is(err, ErrBranchNotFound) {
		t.Errorf("error = %v, want ErrBranchNotFound", err)
	}
}

func TestBranchList(t *testing.T) {
	dir := initTestRepo(t)
	bm := NewBranchManager(dir)

	// Create additional branches.
	if err := bm.Create("develop"); err != nil {
		t.Fatal(err)
	}
	if err := bm.Create("feature/a"); err != nil {
		t.Fatal(err)
	}

	branches, err := bm.List()
	if err != nil {
		t.Fatalf("List() error: %v", err)
	}

	if len(branches) != 3 {
		t.Fatalf("List() returned %d branches, want 3", len(branches))
	}

	// Verify all branches have non-empty names.
	for _, b := range branches {
		if b.Name == "" {
			t.Error("branch has empty Name")
		}
		if b.IsRemote {
			t.Errorf("branch %q has IsRemote=true, want false", b.Name)
		}
	}

	// Only "main" should be current.
	currentCount := 0
	for _, b := range branches {
		if b.IsCurrent {
			currentCount++
			if b.Name != "main" {
				t.Errorf("current branch is %q, want %q", b.Name, "main")
			}
		}
	}
	if currentCount != 1 {
		t.Errorf("found %d current branches, want 1", currentCount)
	}
}

func TestValidateBranchName(t *testing.T) {
	valid := []string{
		"main",
		"feature/login",
		"release/v1.0.0",
		"fix-123",
		"a",
	}
	for _, name := range valid {
		t.Run("valid:"+name, func(t *testing.T) {
			if err := validateBranchName(name); err != nil {
				t.Errorf("validateBranchName(%q) = %v, want nil", name, err)
			}
		})
	}

	invalid := []struct {
		name   string
		reason string
	}{
		{"", "empty"},
		{".hidden", "starts with dot"},
		{"a..b", "double dots"},
		{"a~1", "tilde"},
		{"a^1", "caret"},
		{"a:b", "colon"},
		{"a b", "space"},
		{"a.lock", "ends with .lock"},
		{"a\\b", "backslash"},
		{"a@{b", "at-brace"},
	}
	for _, tt := range invalid {
		t.Run("invalid:"+tt.reason, func(t *testing.T) {
			err := validateBranchName(tt.name)
			if err == nil {
				t.Errorf("validateBranchName(%q) = nil, want error for: %s", tt.name, tt.reason)
			}
			if !errors.Is(err, ErrInvalidBranchName) {
				t.Errorf("error = %v, want ErrInvalidBranchName", err)
			}
		})
	}
}

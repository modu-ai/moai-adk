package harness

import (
	"os"
	"path/filepath"
	"testing"
)

func mkSkillDir(t *testing.T, base, name string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Join(base, name), 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", name, err)
	}
}

func TestDetectPrefixConflicts_ExactSuffix(t *testing.T) {
	dir := t.TempDir()
	mkSkillDir(t, dir, "moai-foundation-core")
	mkSkillDir(t, dir, "my-harness-foundation-core")
	conflicts, err := DetectPrefixConflicts(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(conflicts) != 1 {
		t.Fatalf("expected 1 conflict, got %d: %+v", len(conflicts), conflicts)
	}
	if conflicts[0].MyHarnessSkill != "my-harness-foundation-core" {
		t.Errorf("wrong my-harness: %s", conflicts[0].MyHarnessSkill)
	}
	if conflicts[0].MoaiSkill != "moai-foundation-core" {
		t.Errorf("wrong moai: %s", conflicts[0].MoaiSkill)
	}
}

func TestDetectPrefixConflicts_NoConflict(t *testing.T) {
	dir := t.TempDir()
	mkSkillDir(t, dir, "moai-foundation-core")
	mkSkillDir(t, dir, "my-harness-ios-patterns")
	conflicts, err := DetectPrefixConflicts(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(conflicts) != 0 {
		t.Errorf("expected 0 conflicts, got %d: %+v", len(conflicts), conflicts)
	}
}

func TestDetectPrefixConflicts_CloseEditDistance(t *testing.T) {
	dir := t.TempDir()
	mkSkillDir(t, dir, "moai-workflow-tdd")
	mkSkillDir(t, dir, "my-harness-workflow-tddd") // 1 char diff
	conflicts, err := DetectPrefixConflicts(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(conflicts) != 1 {
		t.Fatalf("expected 1 conflict (close), got %d", len(conflicts))
	}
}

func TestDetectPrefixConflicts_MissingDir(t *testing.T) {
	conflicts, err := DetectPrefixConflicts(filepath.Join(t.TempDir(), "nope"))
	if err != nil {
		t.Errorf("missing dir should not error: %v", err)
	}
	if len(conflicts) != 0 {
		t.Errorf("expected empty, got %+v", conflicts)
	}
}

func TestDetectPrefixConflicts_EmptyPath(t *testing.T) {
	if _, err := DetectPrefixConflicts(""); err == nil {
		t.Fatal("expected error")
	}
}

func TestLevenshtein_KnownDistances(t *testing.T) {
	cases := []struct {
		a, b string
		want int
	}{
		{"", "abc", 3},
		{"abc", "", 3},
		{"abc", "abc", 0},
		{"kitten", "sitting", 3},
		{"flaw", "lawn", 2},
	}
	for _, c := range cases {
		if got := levenshtein(c.a, c.b); got != c.want {
			t.Errorf("levenshtein(%q,%q) = %d, want %d", c.a, c.b, got, c.want)
		}
	}
}

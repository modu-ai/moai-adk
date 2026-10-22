package cli

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// t1784: `factory assign --worktree` stores the SameFile-verified real path
// when the path exists, so variant spellings of the same tree (/tmp/x vs
// /private/tmp/x) cannot fork the record. A nonexistent path falls back to
// filepath.Abs.
func TestCanonicalWorktreePath(t *testing.T) {
	t.Run("existing path via symlink normalizes to the real path", func(t *testing.T) {
		if runtime.GOOS == "windows" {
			t.Skip("os.Symlink requires privilege on windows")
		}
		base := t.TempDir()
		real := filepath.Join(base, "real-tree")
		if err := os.Mkdir(real, 0o755); err != nil {
			t.Fatal(err)
		}
		link := filepath.Join(base, "link-tree")
		if err := os.Symlink(real, link); err != nil {
			t.Fatal(err)
		}
		got, err := canonicalWorktreePath(link)
		if err != nil {
			t.Fatalf("canonicalWorktreePath: %v", err)
		}
		want, err := filepath.EvalSymlinks(real)
		if err != nil {
			t.Fatal(err)
		}
		if got != want {
			t.Fatalf("normalized %q, want the SameFile-verified real path %q", got, want)
		}
	})

	t.Run("nonexistent path falls back to filepath.Abs", func(t *testing.T) {
		input := filepath.Join(t.TempDir(), "does-not-exist-yet")
		got, err := canonicalWorktreePath(input)
		if err != nil {
			t.Fatalf("canonicalWorktreePath: %v", err)
		}
		want, err := filepath.Abs(input)
		if err != nil {
			t.Fatal(err)
		}
		if got != want {
			t.Fatalf("got %q, want the Abs fallback %q", got, want)
		}
	})

	t.Run("existing plain path is unchanged up to symlink resolution", func(t *testing.T) {
		input := t.TempDir()
		got, err := canonicalWorktreePath(input)
		if err != nil {
			t.Fatalf("canonicalWorktreePath: %v", err)
		}
		same, err := sameFilePath(input, got)
		if err != nil {
			t.Fatalf("sameFilePath: %v", err)
		}
		if !same {
			t.Fatalf("normalized %q is not the same file as the input %q", got, input)
		}
		if !filepath.IsAbs(got) {
			t.Fatalf("got %q, want an absolute path", got)
		}
	})

	t.Run("whitespace-padded input is refused (storage trims)", func(t *testing.T) {
		input := " " + filepath.Join(t.TempDir(), "does-not-exist-yet") + " "
		_, err := canonicalWorktreePath(input)
		if err == nil {
			t.Fatalf("canonicalWorktreePath(%q) accepted a non-trim-stable path; the persist layer (CardFields.apply) trims it, breaking the SameFile guarantee", input)
		}
	})

	t.Run("canonical result with leading or trailing whitespace is refused", func(t *testing.T) {
		if runtime.GOOS == "windows" {
			t.Skip("trailing-space directory names are not creatable on windows")
		}
		base := t.TempDir()
		padded := filepath.Join(base, "padded ")
		if err := os.Mkdir(padded, 0o755); err != nil {
			t.Fatal(err)
		}
		got, err := canonicalWorktreePath(padded)
		if err == nil {
			t.Fatalf("canonicalWorktreePath accepted %q whose canonical form is not trim-stable (got %q)", padded, got)
		}
		if got != "" {
			t.Fatalf("refused path must not return a spelling, got %q", got)
		}
		if strings.TrimSpace(padded) == padded {
			t.Fatalf("fixture broken: %q is trim-stable, the subtest exercises nothing", padded)
		}
	})
}

func TestCanonicalWorktreePath_RejectsRawPadding(t *testing.T) {
	base := filepath.Join(t.TempDir(), "does-not-exist-yet")
	for _, tc := range []struct {
		name, input string
	}{
		{"leading space", " " + base},
		{"trailing space", base + " "},
		{"leading tab", "\t" + base},
		{"trailing tab", base + "\t"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := canonicalWorktreePath(tc.input)
			if err == nil || got != "" {
				t.Fatalf("canonicalWorktreePath(%q) = %q, %v; want empty path and refusal", tc.input, got, err)
			}
			if !strings.Contains(err.Error(), "not trim-stable") {
				t.Errorf("refusal = %v, want the storage trim-stability cause", err)
			}
		})
	}
}

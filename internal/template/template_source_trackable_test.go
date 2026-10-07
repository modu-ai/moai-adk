package template_test

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// TestTemplateSourceFullyTrackable measures — with git itself — that every
// file under the template source tree stays visible to a fresh clone.
//
// internal/template/templates/ is the shipped artifact: `//go:embed
// all:templates` compiles whatever sits in this working tree, but a fresh
// clone only receives what git tracks. A template-source file the ignore
// rules swallow therefore embeds on the author's machine and vanishes on
// everyone else's — the build breaks, or the template set silently shrinks.
// Measured 2026-10-07 (card t1540):
// .claude/skills/moai-domain-html-report/references/design-tokens.md was
// ignored by the "*token*" credential rules in the nested
// templates/.gitignore and absent from the index.
//
// The patterns are MEASURED, never inferred (same discipline as
// TestTemplateGitignoreIgnoresLocalArtifacts): one batched
// `git check-ignore --stdin -z` call decides the whole tree, so the
// subprocess count stays constant regardless of corpus size. The index is
// consulted deliberately (no --no-index): the hazard is a file that is
// ignored AND untracked — the pair that vanishes from a fresh clone.
// Tracked files the patterns would ignore (e.g. the deliberate .gitkeep
// scaffolds under templates/.moai/) ship fine and stay out of scope.
func TestTemplateSourceFullyTrackable(t *testing.T) {
	t.Parallel()

	gitBin, err := exec.LookPath("git")
	if err != nil {
		t.Skipf("git not available: %v", err)
	}

	repoRoot := hocProjectRoot(t)
	tmplRoot := filepath.Join(repoRoot, "internal", "template", "templates")

	var paths []string
	walkErr := filepath.WalkDir(tmplRoot, func(p string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		rel, relErr := filepath.Rel(repoRoot, p)
		if relErr != nil {
			return relErr
		}
		paths = append(paths, filepath.ToSlash(rel))
		return nil
	})
	if walkErr != nil {
		t.Fatalf("walk template source tree: %v", walkErr)
	}
	// Guard-of-the-guard: a walk that found almost nothing means the tree
	// path moved — refuse to pass vacuously.
	if len(paths) < 200 {
		t.Fatalf("template source walk found only %d files under %s — tree path wrong or tree emptied; check before trusting this guard", len(paths), tmplRoot)
	}

	payload := []byte(strings.Join(paths, "\x00") + "\x00")
	cmd := exec.Command(gitBin, "check-ignore", "--stdin", "-z")
	cmd.Dir = repoRoot
	cmd.Stdin = bytes.NewReader(payload)
	out, runErr := cmd.Output()

	var ignored []string
	if runErr == nil {
		// Exit 0: at least one path is ignored — output carries them, NUL-separated.
		for p := range bytes.SplitSeq(out, []byte{0}) {
			if len(p) > 0 {
				ignored = append(ignored, string(p))
			}
		}
	} else {
		exitErr, ok := runErr.(*exec.ExitError)
		if !ok || exitErr.ExitCode() != 1 {
			t.Fatalf("git check-ignore --stdin: %v", runErr)
		}
		// Exit 1: nothing is ignored — the invariant holds.
	}

	for _, p := range ignored {
		t.Errorf("template source file is gitignored and would vanish from a fresh clone: %s — add a narrow re-include exception to the ignore rules (card t1540 shape: !.claude/skills/*/references/*token*.md)", p)
	}
	t.Logf("measured %d template source files, %d gitignored", len(paths), len(ignored))
}

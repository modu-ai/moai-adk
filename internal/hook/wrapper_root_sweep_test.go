package hook

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

// Root-copy contract for every deployable hook wrapper, swept rather than
// listed. hookWrapperContracts() names six wrappers and the stop-goal test
// names a seventh; the remaining root copies drift with nothing watching them.
// A hand list cannot grow on its own, so the set it protects is whatever
// somebody last remembered to type.
//
// The list was not arbitrary, though: its comment records that an explicit list
// makes a DELETED copy fail (read error) instead of silently shrinking the
// swept set. That property lives in the read assertion, not in the list, so a
// sweep keeps it by requiring the opposite side to exist — the same shape
// internal/template/agentemit/golden_test.go uses when it sweeps emitted
// artifacts and requires each committed one to be present.
//
// hookWrapperContracts() STAYS, and not because anyone forgot to remove it:
// it also compares the template .sh copy against the root .sh, an axis this
// sweep never reads — every pair here starts at a .sh.tmpl and asserts only the
// root counterpart. The two overlap on the root axis alone, so deleting the
// list as "now redundant" silently drops the template-.sh axis with no test
// reporting the loss. Folding that axis in is a separate change: it needs the
// 13 template .sh files swept first.
//
// Two predicates, because one cannot cover the set. Byte-identity is the
// default and holds for 34 of the 35 pairs. It CANNOT hold for
// handle-pre-tool.sh: the root copy carries a SPEC-ID token in a comment and
// the template neutrality guard rejects that token inside the template tree
// (observed: injecting it fails TestTemplateNoInternalContentLeak with
// class=C1-spec-id-prefix). Copying the root comment into the template would
// break that guard, so this pair is registered for the structural predicate
// instead — and registration is deliberately a list, because an EXCEPTION list
// left empty falls back to the strict predicate for everything, whereas a
// TARGET list left empty covers nothing.
//
// @MX:ANCHOR: [AUTO] root-copy sweep for hook wrappers — replaces the hand list's scope with the whole deployable set.
// @MX:REASON: A wrapper added without being typed into hookWrapperContracts() had no root-copy guard at all; the sweep removes that silent gap.

// structuralPairWrappers registers the wrapper pairs that CANNOT be
// byte-identical and are therefore compared structurally. Membership is a
// decision with a stated reason, never a convenience: the strict predicate
// applies to every pair not listed here.
var structuralPairWrappers = map[string]string{
	// The root copy cites a SPEC-ID in a comment; the template tree forbids that
	// token (C1-spec-id-prefix). Both copies were last changed in the same commit
	// 3d74559aa, so the difference is authored, not drift.
	"handle-pre-tool.sh": "root copy cites a SPEC-ID the template neutrality guard rejects",
}

// TestHookWrapperRootCopySweep sweeps every .sh.tmpl in the template hooks dir,
// requires the deployed root copy to exist, and compares the pair.
func TestHookWrapperRootCopySweep(t *testing.T) {
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	repoRoot := filepath.Join(cwd, "..", "..")
	templateDir := filepath.Join(repoRoot, "internal", "template", "templates", ".claude", "hooks", "moai")
	rootDir := filepath.Join(repoRoot, ".claude", "hooks", "moai")

	entries, err := os.ReadDir(templateDir)
	if err != nil {
		t.Fatalf("read template hooks dir %s: %v", templateDir, err)
	}

	swept, structural := 0, 0
	for _, e := range entries {
		name := e.Name()
		if e.IsDir() || !strings.HasSuffix(name, ".sh.tmpl") {
			continue
		}
		shName := strings.TrimSuffix(name, ".tmpl")
		swept++

		tmplData, readErr := os.ReadFile(filepath.Join(templateDir, name))
		if readErr != nil {
			t.Errorf("read template %s: %v", name, readErr)
			continue
		}
		// Existence requirement: this is what a deleted root copy trips. The
		// hand list got the same effect from a read error on a named path.
		rootData, readErr := os.ReadFile(filepath.Join(rootDir, shName))
		if readErr != nil {
			t.Errorf("root copy missing for %s (%v) — .claude/hooks/moai/%s is what this repo executes; deploying the template without it leaves the wrapper unrunnable here",
				name, readErr, shName)
			continue
		}

		if reason, ok := structuralPairWrappers[shName]; ok {
			structural++
			assertSameShape(t, shName, string(tmplData), string(rootData), reason)
			continue
		}
		if string(tmplData) != string(rootData) {
			t.Errorf("ROOT DRIFT: templates/.claude/hooks/moai/%s differs from .claude/hooks/moai/%s — the .tmpl is what moai update deploys and the root copy is what this repo runs, so a one-copy edit is either reverted on the next update or never reaches users (CLAUDE.local.md §2.3). Edit both, or register the pair in structuralPairWrappers with the reason it cannot be byte-identical",
				name, shName)
		}
	}

	// Guard-of-the-guard: a zero-sweep run passes vacuously, and a sweep cannot
	// see the deletion of the set it sweeps (a removed .sh.tmpl simply shrinks
	// the set). The floor below bounds that blind spot without naming files:
	// the set only ever grows in practice, so a drop reads as a deletion.
	const sweptFloor = 35
	if swept < sweptFloor {
		t.Errorf("swept %d .sh.tmpl wrappers, expected at least %d — a template wrapper was deleted, or the directory moved; a shrinking sweep passes silently, which is why this floor exists",
			swept, sweptFloor)
	}
	t.Logf("swept %d wrappers against their root copies (%d structural, %d byte-identical)", swept, structural, swept-structural)
}

// assertSameShape compares two wrapper copies as line multisets after dropping
// comment-only lines, so an authored comment difference passes while a change
// to an executable line does not.
func assertSameShape(t *testing.T, name, tmpl, root, reason string) {
	t.Helper()
	tmplCode, rootCode := codeLines(tmpl), codeLines(root)
	if len(tmplCode) != len(rootCode) {
		t.Errorf("STRUCTURAL DRIFT: %s has %d executable lines in the template and %d in the root copy (registered as a structural pair: %s) — comments may differ, code may not",
			name, len(tmplCode), len(rootCode), reason)
		return
	}
	for i := range tmplCode {
		if tmplCode[i] != rootCode[i] {
			t.Errorf("STRUCTURAL DRIFT: %s executable line %d differs (registered as a structural pair: %s)\n  template: %s\n  root:     %s",
				name, i+1, reason, tmplCode[i], rootCode[i])
			return
		}
	}
}

// codeLines returns the non-blank, non-comment lines of a shell script.
func codeLines(s string) []string {
	var out []string
	for _, line := range strings.Split(s, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}
		out = append(out, trimmed)
	}
	return out
}

// TestStructuralPairRegistryNamesRealWrappers keeps the exception list honest:
// a registered name that no longer exists would silently exempt nothing while
// reading as a live exemption.
func TestStructuralPairRegistryNamesRealWrappers(t *testing.T) {
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	templateDir := filepath.Join(cwd, "..", "..", "internal", "template", "templates", ".claude", "hooks", "moai")

	names := make([]string, 0, len(structuralPairWrappers))
	for name := range structuralPairWrappers {
		names = append(names, name)
	}
	sort.Strings(names)

	for _, name := range names {
		if _, err := os.Stat(filepath.Join(templateDir, name+".tmpl")); err != nil {
			t.Errorf("structuralPairWrappers registers %q but %s.tmpl does not exist (%v) — a stale exemption reads as deliberate while exempting nothing",
				name, name, err)
		}
		if reason := structuralPairWrappers[name]; strings.TrimSpace(reason) == "" {
			t.Errorf("structuralPairWrappers[%q] has an empty reason — registration drops this pair's byte-identity protection, so the reason is the record of why that trade was made", name)
		}
	}
}

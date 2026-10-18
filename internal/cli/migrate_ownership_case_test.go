package cli

// migrate_ownership_case_test.go — SPEC-UPDATE-OWNERSHIP-REPAIR-001 M4 (card
// t1865), AC-OWR-11 and AC-OWR-12: a path that differs from its manifest key only
// in letter case.
//
// The verdict function treats the two spellings as one file only when the
// observed file system says they are the same file (os.SameFile), and treats them
// as two files otherwise. It does not decide the file system's case mode on its
// own and it never renames. Two things make that testable on any host:
//
//   - resolveCaseAlias takes the stat and same-file operations as arguments, and
//     the tests run it on two emulated modes over the real directory contents, so
//     the case-insensitive and case-sensitive outcomes are both asserted on every
//     machine, with nothing skipped;
//   - the host test asserts what THIS machine's file system must give, after
//     probing its case mode inside the test's own temporary directory (never in a
//     project root: the operational probe writes there and is not called).
//
// Assertion functions are pure; the mutant sub-tests feed them a resolver that
// always aliases and require a non-empty problem list.

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/manifest"
)

// emulatedInsensitiveFS answers stat as a case-insensitive file system would, by
// finding the directory entry that matches the base name regardless of case,
// over the real directory contents. Same-file identity is the real one.
func emulatedInsensitiveFS() caseFS {
	return caseFS{
		stat: func(name string) (fs.FileInfo, error) {
			dir, base := filepath.Split(name)
			entries, err := os.ReadDir(dir)
			if err != nil {
				return nil, err
			}
			for _, e := range entries {
				if strings.EqualFold(e.Name(), base) {
					return os.Stat(filepath.Join(dir, e.Name()))
				}
			}
			return nil, fs.ErrNotExist
		},
		sameFile: os.SameFile,
	}
}

// emulatedSensitiveFS answers stat as a case-sensitive file system would: only
// the exact spelling exists.
func emulatedSensitiveFS() caseFS {
	return caseFS{
		stat: func(name string) (fs.FileInfo, error) {
			dir, base := filepath.Split(name)
			entries, err := os.ReadDir(dir)
			if err != nil {
				return nil, err
			}
			for _, e := range entries {
				if e.Name() == base {
					return os.Stat(filepath.Join(dir, e.Name()))
				}
			}
			return nil, fs.ErrNotExist
		},
		sameFile: os.SameFile,
	}
}

// caseProject builds a project with the registered healthy copy written under
// the registered spelling and then renamed to the lower-case spelling on disk,
// and returns the root and the loaded manifest manager.
func caseProject(t *testing.T) (string, manifest.Manager) {
	t.Helper()
	root := t.TempDir()
	data := ownEmbeddedBytes(t, caseRegisteredRel)
	ownAddRegisteredManaged(t, root, caseRegisteredRel, data)
	registered := filepath.Join(root, filepath.FromSlash(caseRegisteredRel))
	lower := filepath.Join(root, filepath.FromSlash(caseDiskRel))
	if err := os.Rename(registered, lower); err != nil {
		t.Fatalf("rename to the lower-case spelling: %v", err)
	}
	mgr := manifest.NewManager()
	if _, err := mgr.Load(root); err != nil {
		t.Fatalf("load manifest: %v", err)
	}
	return root, mgr
}

// caseEnv is the verdict environment of the case tests: every removal
// precondition holds (so an alias that applies the registered entry's verdict is
// visibly removal-eligible), the case alias resolver is the given one.
func caseEnv(t *testing.T, root string, mgr manifest.Manager, fsys caseFS) ownershipEnv {
	t.Helper()
	rootResolved, err := filepath.EvalSymlinks(root)
	if err != nil {
		t.Fatal(err)
	}
	keys := []string{caseRegisteredRel}
	return ownershipEnv{
		entry:                mgr.GetEntry,
		evidence:             fakeProvenance{},
		counterpartConfirmed: func(string) bool { return true },
		notInCatalog:         func(string) bool { return false },
		removalEnabled:       true,
		rootResolved:         rootResolved,
		caseAlias: func(rel string) (string, bool) {
			return resolveCaseAlias(fsys, root, rel, keys)
		},
	}
}

// caseVerdictProblems is the pure assertion of AC-OWR-11. In a case-insensitive
// mode the lower-case path takes the registered entry's verdict (managed,
// removal-eligible here), is reported as a case variant of the registered
// spelling and is keyed to the registered entry. In a case-sensitive mode it is a
// separate unregistered file and is not keyed to any entry.
func caseVerdictProblems(insensitive bool, v ownershipVerdict) []string {
	var problems []string
	if insensitive {
		if v.class != ownershipManaged {
			problems = append(problems, fmt.Sprintf("class %q, want %q (the registered entry's verdict applies)", v.class, ownershipManaged))
		}
		if !v.removalEligible {
			problems = append(problems, fmt.Sprintf("not removal-eligible (keep %q): the registered entry's disposition was not applied", v.keep))
		}
		if v.caseVariantOf != caseRegisteredRel {
			problems = append(problems, fmt.Sprintf("case variant of %q, want %q", v.caseVariantOf, caseRegisteredRel))
		}
		if v.key != caseRegisteredRel {
			problems = append(problems, fmt.Sprintf("keyed to %q, want the registered spelling %q", v.key, caseRegisteredRel))
		}
		return problems
	}
	if v.class != ownershipUnregistered {
		problems = append(problems, fmt.Sprintf("class %q, want %q (two spellings are two files)", v.class, ownershipUnregistered))
	}
	if v.key != "" {
		problems = append(problems, fmt.Sprintf("keyed to %q, want no entry", v.key))
	}
	if v.removalEligible {
		problems = append(problems, "removal-eligible although it is an unregistered file")
	}
	return problems
}

func TestCaseVariantResolvesToRegisteredEntry(t *testing.T) {
	judge := func(t *testing.T, fsys caseFS) (ownershipVerdict, string, map[string]string) {
		t.Helper()
		root, mgr := caseProject(t)
		abs := filepath.Join(root, filepath.FromSlash(caseDiskRel))
		before := ownTreeSnapshot(t, root, ownTreeOpts{})
		v := classifyOwnership(caseEnv(t, root, mgr, fsys), caseDiskRel, abs)
		return v, root, before
	}

	t.Run("case-insensitive mode: one file, the registered verdict", func(t *testing.T) {
		v, root, before := judge(t, emulatedInsensitiveFS())
		for _, p := range caseVerdictProblems(true, v) {
			t.Error(p)
		}
		if diff := ownTreeDiff(before, ownTreeSnapshot(t, root, ownTreeOpts{})); len(diff) != 0 {
			t.Errorf("the verdict changed the tree (a rename or a write): %v", diff)
		}
		entries, _ := os.ReadDir(filepath.Dir(filepath.Join(root, filepath.FromSlash(caseDiskRel))))
		if len(entries) != 1 || entries[0].Name() != "skill.md" {
			t.Errorf("the directory holds %v, want the single file skill.md untouched", entries)
		}
	})

	t.Run("case-insensitive mode: a user-modified registered file stays user-modified", func(t *testing.T) {
		root, mgr := caseProject(t)
		abs := filepath.Join(root, filepath.FromSlash(caseDiskRel))
		ownWriteFile(t, root, caseDiskRel, append(ownEmbeddedBytes(t, caseRegisteredRel), []byte("\nuser edit\n")...))
		v := classifyOwnership(caseEnv(t, root, mgr, emulatedInsensitiveFS()), caseDiskRel, abs)
		if v.class != ownershipUserModified || v.removalEligible {
			t.Errorf("verdict %+v, want user-modified and kept: the alias must carry the registered entry's verdict, not override it", v)
		}
	})

	t.Run("case-sensitive mode: two files", func(t *testing.T) {
		v, _, _ := judge(t, emulatedSensitiveFS())
		for _, p := range caseVerdictProblems(false, v) {
			t.Error(p)
		}
	})

	t.Run("the exact registered spelling is looked up first", func(t *testing.T) {
		root := t.TempDir()
		ownAddRegisteredManaged(t, root, caseRegisteredRel, ownEmbeddedBytes(t, caseRegisteredRel))
		mgr := manifest.NewManager()
		if _, err := mgr.Load(root); err != nil {
			t.Fatal(err)
		}
		abs := filepath.Join(root, filepath.FromSlash(caseRegisteredRel))
		env := caseEnv(t, root, mgr, emulatedInsensitiveFS())
		env.caseAlias = func(string) (string, bool) {
			t.Error("the case alias was consulted although the walk path is the registered key")
			return "", false
		}
		v := classifyOwnership(env, caseRegisteredRel, abs)
		if v.class != ownershipManaged || v.caseVariantOf != "" {
			t.Errorf("verdict %+v, want managed with no case variant for the exact spelling", v)
		}
	})

	t.Run("mutant: an always-alias resolver is caught", func(t *testing.T) {
		root, mgr := caseProject(t)
		abs := filepath.Join(root, filepath.FromSlash(caseDiskRel))
		env := caseEnv(t, root, mgr, emulatedSensitiveFS())
		env.caseAlias = func(string) (string, bool) { return caseRegisteredRel, true }
		v := classifyOwnership(env, caseDiskRel, abs)
		if len(caseVerdictProblems(false, v)) == 0 {
			t.Errorf("the case-sensitive assertion stayed green in front of a resolver that always aliases")
		}
	})

	t.Run("an ambiguous alias is no alias", func(t *testing.T) {
		root, _ := caseProject(t)
		got, ok := resolveCaseAlias(emulatedInsensitiveFS(), root, caseDiskRel,
			[]string{caseRegisteredRel, ".claude/skills/moai-ref-owasp-checklist/Skill.md"})
		if ok {
			t.Errorf("two candidate spellings resolved to %q; an ambiguous match must give no alias", got)
		}
	})
}

// TestCaseVariantCounterpartAndMirrorSpelling is AC-OWR-12: the user-root
// manifest records `SKILL.md`, the project file is spelled `skill.md`, and an
// unregistered mirror copy is spelled `skill.md` where the shipped file is
// `SKILL.md`.
func TestCaseVariantCounterpartAndMirrorSpelling(t *testing.T) {
	const mirrorRel = ".agents/skills/moai-plan/skill.md"
	const mirrorShipped = ".agents/skills/moai-plan/SKILL.md"

	build := func(t *testing.T) (root, home string) {
		t.Helper()
		ownProfileSandbox(t)
		home = installMigrationUserCounterparts(t)
		requireUserCounterpart(t, home, caseRegisteredRel)
		root = t.TempDir()
		// The project file is written, tracked and kept under the lower-case
		// spelling: the project manifest key IS `skill.md`.
		ownAddRegisteredManaged(t, root, caseDiskRel, ownEmbeddedBytes(t, caseRegisteredRel))
		// An unregistered mirror copy in the lower-case spelling.
		ownAddUnregistered(t, root, mirrorRel, ownEmbeddedBytes(t, mirrorShipped))
		return root, home
	}

	run := func(t *testing.T, root, home string, fsys caseFS) ownershipRun {
		t.Helper()
		prev := hostCaseFS
		hostCaseFS = fsys
		t.Cleanup(func() { hostCaseFS = prev })
		return runOwnershipMigration(t, root, home)
	}

	t.Run("case-insensitive: the spelling does not make the counterpart unconfirmed", func(t *testing.T) {
		root, home := build(t)
		before := ownTreeSnapshot(t, root, ownTreeOpts{})
		r := run(t, root, home, emulatedInsensitiveFS())

		// The removal switch is off, so the project file is kept with its
		// disposition reason; what matters is the counterpart column.
		row, ok := findRemaining(r.rc, caseDiskRel)
		if !ok {
			t.Fatalf("no remaining row for %s (removed=%v)", caseDiskRel, r.removed)
		}
		if row.Counterpart != counterpartStatusConfirmed {
			t.Errorf("%s: counterpart %q, want %q (the user manifest records %s)", caseDiskRel, row.Counterpart, counterpartStatusConfirmed, caseRegisteredRel)
		}

		mirror, ok := findRemaining(r.rc, mirrorRel)
		if !ok {
			t.Fatalf("no remaining row for the mirror copy %s", mirrorRel)
		}
		if mirror.Class != ownershipUnregistered || mirror.Reason != keepNoProvenanceRecord {
			t.Errorf("mirror row %+v, want unregistered / %s", mirror, keepNoProvenanceRecord)
		}
		if mirror.CaseVariantOf != mirrorShipped {
			t.Errorf("mirror row reports case variant of %q, want the shipped spelling %q", mirror.CaseVariantOf, mirrorShipped)
		}
		count := 0
		for _, rr := range r.rc.remaining {
			if rr.Rel == mirrorRel {
				count++
			}
		}
		if count != 1 {
			t.Errorf("the mirror copy has %d rows, want exactly 1", count)
		}
		after := ownTreeSnapshot(t, root, ownTreeOpts{SkipPrefixes: []string{".moai-backups/"}})
		if diff := ownTreeDiff(before, after); len(diff) != 0 {
			t.Errorf("the update changed the tree: %v (no rename and no rewrite of either spelling)", diff)
		}
		for _, d := range []string{".agents/skills/moai-plan", ".claude/skills/moai-ref-owasp-checklist"} {
			entries, _ := os.ReadDir(filepath.Join(root, filepath.FromSlash(d)))
			if len(entries) != 1 || entries[0].Name() != "skill.md" {
				t.Errorf("%s holds %v, want the one file skill.md", d, entries)
			}
		}
	})

	t.Run("case-sensitive control: the same project reports an unconfirmed counterpart", func(t *testing.T) {
		root, home := build(t)
		r := run(t, root, home, emulatedSensitiveFS())
		row, ok := findRemaining(r.rc, caseDiskRel)
		if !ok {
			t.Fatalf("no remaining row for %s", caseDiskRel)
		}
		if row.Counterpart == counterpartStatusConfirmed {
			t.Errorf("%s: counterpart %q in a case-sensitive mode; two spellings are two files and the user manifest has no skill.md key", caseDiskRel, row.Counterpart)
		}
	})
}

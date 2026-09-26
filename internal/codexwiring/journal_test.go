package codexwiring

import (
	"bytes"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/manifest"
)

// errTestCrash stands for the process dying at a fault point: the pass stops
// there and nothing after the point runs.
var errTestCrash = errors.New("simulated crash")

const userHooksJSON = `{
  "description": "user hooks",
  "hooks": {
    "PreToolUse": [
      {
        "matcher": "Bash",
        "hooks": [
          {"type": "command", "command": "echo user", "timeout": 5}
        ]
      }
    ]
  }
}
`

const userConfigTOML = "model = \"gpt-5\"\n\n[mcp_servers.other]\ncommand = \"other\"\n"

func writeFile(t *testing.T, path, body string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

func readFile(t *testing.T, path string) []byte {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return b
}

func userProject(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	writeFile(t, filepath.Join(root, HooksRelPath), userHooksJSON)
	writeFile(t, filepath.Join(root, ConfigRelPath), userConfigTOML)
	return root
}

func journalFor(t *testing.T, root string) map[string]JournalEntry {
	t.Helper()
	entries, err := LoadJournal(root)
	if err != nil {
		t.Fatalf("load journal: %v", err)
	}
	out := map[string]JournalEntry{}
	for _, e := range entries {
		out[e.Path] = e
	}
	return out
}

func manifestEntry(t *testing.T, root, rel string) (manifest.FileEntry, bool) {
	t.Helper()
	mgr := manifest.NewManager()
	mf, err := mgr.Load(root)
	if err != nil {
		t.Fatalf("load manifest: %v", err)
	}
	e, ok := mf.Files[rel]
	return e, ok
}

func tempFiles(t *testing.T, dir string) []string {
	t.Helper()
	matches, err := filepath.Glob(filepath.Join(dir, tempPrefix+"*"))
	if err != nil {
		t.Fatal(err)
	}
	return matches
}

// checkWriteReadback returns nil when a clean pass left both wiring files
// journaled complete with read-back hashes matching and part records applied.
func checkWriteReadback(root string, entries map[string]JournalEntry) error {
	for _, rel := range []string{HooksRelPath, ConfigRelPath} {
		e, ok := entries[rel]
		if !ok {
			return fmt.Errorf("%s: no journal entry", rel)
		}
		if e.State != JournalComplete {
			return fmt.Errorf("%s: journal state %q, want %q", rel, e.State, JournalComplete)
		}
		b, err := os.ReadFile(filepath.Join(root, rel))
		if err != nil {
			return err
		}
		if got := sha256Hex(b); got != e.PostHash {
			return fmt.Errorf("%s: on-disk hash %s != intended %s", rel, got, e.PostHash)
		}
	}
	return nil
}

// TestCodexWiringJournaledWriteReadback covers AC-DHR-001 (a): every wiring
// file change is journaled, read back against the intended hash, and its part
// provenance lands in the manifest before the entry is marked complete. The
// in-test mutant table shows the read-back is load-bearing: a target that
// changes between rename and read-back must end diverged with no provenance
// applied, and dropping the read-back lets it complete.
func TestCodexWiringJournaledWriteReadback(t *testing.T) {
	root := userProject(t)
	var out, warn bytes.Buffer
	if _, err := Wire(root, &out, &warn); err != nil {
		t.Fatalf("Wire: %v (warn=%s)", err, warn.String())
	}
	entries := journalFor(t, root)
	if err := checkWriteReadback(root, entries); err != nil {
		t.Fatal(err)
	}
	cfg, ok := manifestEntry(t, root, ConfigRelPath)
	if !ok || cfg.Provenance != manifest.GeneratedManaged {
		t.Fatalf("config.toml manifest entry=%+v ok=%v", cfg, ok)
	}
	var mcp *manifest.Part
	for i := range cfg.Parts {
		if cfg.Parts[i].Key == PartKeyMCPTable {
			mcp = &cfg.Parts[i]
		}
	}
	if mcp == nil || mcp.Origin != manifest.OriginCreated || mcp.Kind != manifest.PartTOMLTable || !strings.HasPrefix(mcp.Region, "\n[mcp_servers.moai]\n") {
		t.Fatalf("mcp part record=%+v", mcp)
	}
	if mcp.Hash != sha256Hex([]byte(mcp.Region)) {
		t.Fatalf("mcp part hash %s does not hash its region", mcp.Hash)
	}
	hooks, ok := manifestEntry(t, root, HooksRelPath)
	if !ok {
		t.Fatal("hooks.json has no manifest entry")
	}
	var handlers, description int
	for _, p := range hooks.Parts {
		switch p.Kind {
		case manifest.PartHookHandler:
			handlers++
			if p.Origin != manifest.OriginCreated || !strings.Contains(p.Key, moaiHandlerPrefix) {
				t.Errorf("handler part %+v", p)
			}
		case manifest.PartJSONKey:
			description++
			if p.Origin != manifest.OriginPreexisting {
				t.Errorf("user description recorded as %q, want preexisting", p.Origin)
			}
		case manifest.PartWholeFile:
			t.Errorf("a user-authored hooks.json must not carry a whole-file part: %+v", p)
		}
	}
	if handlers == 0 || description != 1 {
		t.Fatalf("hooks parts: handlers=%d description=%d (%+v)", handlers, description, hooks.Parts)
	}

	// Mutant table: the read-back guard.
	for _, tc := range []struct {
		name     string
		readback bool
		detected bool
	}{
		{"real", true, true},
		{"mutant_no_readback", false, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			root := userProject(t)
			hooksPath := filepath.Join(root, HooksRelPath)
			opts := passOptions{guards: writeGuards{recheck: true, readback: tc.readback, lstat: true}}
			opts.fault = func(point, rel string) error {
				if point == pointRenamed && rel == HooksRelPath {
					return os.WriteFile(hooksPath, []byte("{}\n"), 0o644)
				}
				return nil
			}
			var out, warn bytes.Buffer
			_, _ = wireWith(root, &out, &warn, opts)
			e := journalFor(t, root)[HooksRelPath]
			_, applied := manifestEntry(t, root, HooksRelPath)
			detected := e.State == JournalDiverged && !applied
			if detected != tc.detected {
				t.Fatalf("read-back detection=%v want %v (state=%q applied=%v)", detected, tc.detected, e.State, applied)
			}
		})
	}
}

// TestCodexWiringRefusesConcurrentModification covers AC-DHR-001 (b): a target
// changed between the temp write and the rename is left byte-identical, the
// temp file is removed, and the journal records a conflict naming both
// hashes. Dropping the pre-rename re-check overwrites the concurrent change.
func TestCodexWiringRefusesConcurrentModification(t *testing.T) {
	const concurrent = "{\n  \"hooks\": {}\n}\n"
	for _, tc := range []struct {
		name    string
		recheck bool
		refuses bool
	}{
		{"real", true, true},
		{"mutant_no_recheck", false, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			root := userProject(t)
			hooksPath := filepath.Join(root, HooksRelPath)
			pre := sha256Hex(readFile(t, hooksPath))
			opts := passOptions{guards: writeGuards{recheck: tc.recheck, readback: true, lstat: true}}
			opts.fault = func(point, rel string) error {
				if point == pointStaged && rel == HooksRelPath {
					return os.WriteFile(hooksPath, []byte(concurrent), 0o644)
				}
				return nil
			}
			var out, warn bytes.Buffer
			res, err := wireWith(root, &out, &warn, opts)
			got := readFile(t, hooksPath)
			e := journalFor(t, root)[HooksRelPath]
			refused := string(got) == concurrent &&
				len(tempFiles(t, filepath.Dir(hooksPath))) == 0 &&
				e.State == JournalConflict && e.PreHash == pre && e.ObservedHash == sha256Hex([]byte(concurrent)) &&
				errors.Is(err, ErrWiringConflict)
			if refused != tc.refuses {
				t.Fatalf("refused=%v want %v (state=%q err=%v file=%q)", refused, tc.refuses, e.State, err, got)
			}
			if tc.refuses {
				var found bool
				for _, c := range res.Conflicts {
					if c.Path == HooksRelPath && c.Expected == pre && c.Observed == sha256Hex([]byte(concurrent)) {
						found = true
					}
				}
				if !found || !strings.Contains(warn.String(), pre) || !strings.Contains(warn.String(), sha256Hex([]byte(concurrent))) {
					t.Fatalf("conflict not reported with both hashes: res=%+v warn=%s", res.Conflicts, warn.String())
				}
			}
		})
	}
}

func snapshotTree(t *testing.T, dir string) map[string]string {
	t.Helper()
	out := map[string]string{}
	err := filepath.WalkDir(dir, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		info, err := os.Lstat(p)
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(dir, p)
		switch {
		case info.Mode()&os.ModeSymlink != 0:
			target, err := os.Readlink(p)
			if err != nil {
				return err
			}
			out[rel] = "link:" + target
		case info.Mode().IsRegular():
			b, err := os.ReadFile(p)
			if err != nil {
				return err
			}
			out[rel] = sha256Hex(b)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return out
}

func sameMap(a, b map[string]string) bool {
	if len(a) != len(b) {
		return false
	}
	for k, v := range a {
		if b[k] != v {
			return false
		}
	}
	return true
}

// TestCodexWiringInstallRefusesSymlinkBoundary covers AC-DHR-001 (c) and (d):
// a hooks.json that is a symbolic link, and a .codex directory linking outside
// the project, are refused on the link itself (Lstat) — the link, its target,
// and the outside tree stay byte-identical and symlink-boundary is reported.
func TestCodexWiringInstallRefusesSymlinkBoundary(t *testing.T) {
	cases := []struct {
		name    string
		guarded []string
		setup   func(t *testing.T, root, outside string)
	}{
		{"hooks_file_link_inside", []string{HooksRelPath, "user-hooks.json"}, func(t *testing.T, root, _ string) {
			writeFile(t, filepath.Join(root, "user-hooks.json"), userHooksJSON)
			if err := os.MkdirAll(filepath.Join(root, ".codex"), 0o755); err != nil {
				t.Fatal(err)
			}
			if err := os.Symlink(filepath.Join(root, "user-hooks.json"), filepath.Join(root, HooksRelPath)); err != nil {
				t.Skipf("symlinks unavailable: %v", err)
			}
		}},
		{"codex_dir_link_outside", []string{".codex"}, func(t *testing.T, root, outside string) {
			writeFile(t, filepath.Join(outside, "config.toml"), userConfigTOML)
			if err := os.Symlink(outside, filepath.Join(root, ".codex")); err != nil {
				t.Skipf("symlinks unavailable: %v", err)
			}
		}},
	}
	for _, tc := range cases {
		for _, m := range []struct {
			name    string
			lstat   bool
			refuses bool
		}{
			{"real", true, true},
			{"mutant_no_lstat", false, false},
		} {
			t.Run(tc.name+"/"+m.name, func(t *testing.T) {
				root, outside := t.TempDir(), t.TempDir()
				tc.setup(t, root, outside)
				beforeRoot := snapshotTree(t, root)
				beforeOutside := snapshotTree(t, outside)
				opts := passOptions{guards: writeGuards{recheck: true, readback: true, lstat: m.lstat}}
				var out, warn bytes.Buffer
				res, _ := wireWith(root, &out, &warn, opts)
				afterRoot := snapshotTree(t, root)
				afterOutside := snapshotTree(t, outside)
				// The link, its in-project target, and the whole outside tree
				// must be untouched; other project files may be wired.
				pick := func(m map[string]string) map[string]string {
					kept := map[string]string{}
					for _, g := range tc.guarded {
						kept[g] = m[filepath.FromSlash(g)]
					}
					return kept
				}
				beforeRoot, afterRoot = pick(beforeRoot), pick(afterRoot)
				var reported bool
				for _, r := range res.Refusals {
					if r.Reason == ReasonSymlinkBoundary {
						reported = true
					}
				}
				refused := sameMap(beforeRoot, afterRoot) && sameMap(beforeOutside, afterOutside) &&
					reported && strings.Contains(warn.String(), string(ReasonSymlinkBoundary))
				if refused != m.refuses {
					t.Fatalf("refused=%v want %v\nroot before=%v\nroot after =%v\noutside before=%v\noutside after =%v\nrefusals=%+v",
						refused, m.refuses, beforeRoot, afterRoot, beforeOutside, afterOutside, res.Refusals)
				}
			})
		}
	}
}

// TestCodexWiringInterruptedChangeRecovery covers AC-DHR-002: a pass stopped
// at every interruption point of design §A.4 is classified by recovery, the
// completed ones carry their journaled provenance exactly once, a diverged
// target is never overwritten, and no .codexwiring-* temp file survives.
func TestCodexWiringInterruptedChangeRecovery(t *testing.T) {
	crashAt := func(point string) passOptions {
		opts := defaultPassOptions()
		opts.fault = func(p, rel string) error {
			if p == point && rel == HooksRelPath {
				return errTestCrash
			}
			return nil
		}
		return opts
	}
	type expectation struct {
		class      RecoveryClass
		targetGone bool
		provenance bool
	}
	writeCases := []struct {
		name   string
		point  string
		expect expectation
		modify bool
	}{
		{"P1", pointJournaled, expectation{class: RecoveryNotApplied, targetGone: true}, false},
		{"P2", pointStaged, expectation{class: RecoveryNotApplied, targetGone: true}, false},
		{"P3", pointRechecked, expectation{class: RecoveryNotApplied, targetGone: true}, false},
		{"P3_user_modified", pointRechecked, expectation{class: RecoveryDiverged}, true},
		{"P4", pointRenamed, expectation{class: RecoveryCompleted, provenance: true}, false},
		{"P5", pointReadBack, expectation{class: RecoveryCompleted, provenance: true}, false},
		{"P6", pointProvenance, expectation{class: RecoveryCompleted, provenance: true}, false},
	}
	for _, tc := range writeCases {
		t.Run(tc.name, func(t *testing.T) {
			root := t.TempDir()
			hooksPath := filepath.Join(root, HooksRelPath)
			const userEdit = "{\"hooks\": {}, \"description\": \"edited by user\"}\n"
			opts := crashAt(tc.point)
			if tc.modify {
				inner := opts.fault
				opts.fault = func(p, rel string) error {
					if p == tc.point && rel == HooksRelPath {
						if err := os.WriteFile(hooksPath, []byte(userEdit), 0o644); err != nil {
							return err
						}
					}
					return inner(p, rel)
				}
			}
			var out, warn bytes.Buffer
			if _, err := wireWith(root, &out, &warn, opts); !errors.Is(err, errTestCrash) {
				t.Fatalf("pass did not stop at %s: %v", tc.point, err)
			}
			staged := journalFor(t, root)[HooksRelPath]
			if staged.State != JournalStaged {
				t.Fatalf("journal state before recovery %q", staged.State)
			}

			outcomes, err := Recover(root, &warn)
			if err != nil {
				t.Fatalf("Recover: %v", err)
			}
			var got RecoveryClass
			for _, o := range outcomes {
				if o.Path == HooksRelPath {
					got = o.Class
				}
			}
			if got != tc.expect.class {
				t.Fatalf("class=%q want %q (outcomes=%+v)", got, tc.expect.class, outcomes)
			}
			_, statErr := os.Stat(hooksPath)
			if tc.expect.targetGone != errors.Is(statErr, os.ErrNotExist) {
				t.Fatalf("target presence wrong: stat err=%v", statErr)
			}
			if tc.modify {
				if b := readFile(t, hooksPath); string(b) != userEdit {
					t.Fatalf("diverged target overwritten: %q", b)
				}
			}
			entry, applied := manifestEntry(t, root, HooksRelPath)
			if applied != tc.expect.provenance {
				t.Fatalf("provenance applied=%v want %v", applied, tc.expect.provenance)
			}
			if applied {
				if !samePartsOnce(entry.Parts, staged.Provenance.Parts) {
					t.Fatalf("recovered provenance %+v differs from journaled %+v", entry.Parts, staged.Provenance.Parts)
				}
			}
			if left := tempFiles(t, filepath.Dir(hooksPath)); len(left) != 0 {
				t.Fatalf("temp files survived recovery: %v", left)
			}
			if pending := journalFor(t, root)[HooksRelPath]; pending.State == JournalStaged {
				t.Fatalf("journal entry still staged after recovery")
			}
		})
	}

	t.Run("P0", func(t *testing.T) {
		root := t.TempDir()
		var out, warn bytes.Buffer
		if _, err := Wire(root, &out, &warn); err != nil {
			t.Fatal(err)
		}
		before := readFile(t, filepath.Join(root, HooksRelPath))
		orphan := filepath.Join(root, ".codex", tempPrefix+"legacy")
		writeFile(t, orphan, "half written")
		outcomes, err := Recover(root, &warn)
		if err != nil {
			t.Fatal(err)
		}
		var sawOrphan bool
		for _, o := range outcomes {
			if o.Class == RecoveryOrphanTemp && filepath.Base(o.Temp) == filepath.Base(orphan) {
				sawOrphan = true
			}
		}
		if !sawOrphan {
			t.Fatalf("orphan temp not reported: %+v", outcomes)
		}
		if left := tempFiles(t, filepath.Join(root, ".codex")); len(left) != 0 {
			t.Fatalf("orphan temp survived: %v", left)
		}
		if after := readFile(t, filepath.Join(root, HooksRelPath)); !bytes.Equal(before, after) {
			t.Fatal("P0 recovery changed the target")
		}
	})

	deleteCases := []struct {
		name       string
		point      string
		class      RecoveryClass
		targetGone bool
		recordLeft bool
	}{
		{"D1", pointDeleteJournaled, RecoveryNotApplied, false, true},
		{"D2", pointDeleted, RecoveryCompleted, true, false},
	}
	for _, tc := range deleteCases {
		t.Run(tc.name, func(t *testing.T) {
			root := t.TempDir()
			var out, warn bytes.Buffer
			if _, err := Wire(root, &out, &warn); err != nil {
				t.Fatal(err)
			}
			hooksPath := filepath.Join(root, HooksRelPath)
			before := readFile(t, hooksPath)
			release, err := acquireWiringLock(root)
			if err != nil {
				t.Fatal(err)
			}
			p := newPass(root, &out, &warn, crashAt(tc.point), false)
			err = p.remove(HooksRelPath)
			release()
			if !errors.Is(err, errTestCrash) {
				t.Fatalf("delete did not stop at %s: %v", tc.point, err)
			}
			outcomes, err := Recover(root, &warn)
			if err != nil {
				t.Fatal(err)
			}
			var got RecoveryClass
			for _, o := range outcomes {
				if o.Path == HooksRelPath {
					got = o.Class
				}
			}
			if got != tc.class {
				t.Fatalf("class=%q want %q (%+v)", got, tc.class, outcomes)
			}
			_, statErr := os.Stat(hooksPath)
			if tc.targetGone != errors.Is(statErr, os.ErrNotExist) {
				t.Fatalf("target presence wrong: %v", statErr)
			}
			if !tc.targetGone {
				if after := readFile(t, hooksPath); !bytes.Equal(before, after) {
					t.Fatal("not-applied delete changed the target")
				}
			}
			if _, left := manifestEntry(t, root, HooksRelPath); left != tc.recordLeft {
				t.Fatalf("manifest record present=%v want %v", left, tc.recordLeft)
			}
			if left := tempFiles(t, filepath.Dir(hooksPath)); len(left) != 0 {
				t.Fatalf("temp files survived: %v", left)
			}
		})
	}
}

// samePartsOnce reports whether got holds exactly the journaled parts, each
// once — a provenance re-applied on top of itself would duplicate records.
func samePartsOnce(got, want []manifest.Part) bool {
	if len(got) != len(want) {
		return false
	}
	seen := map[string]int{}
	for _, p := range got {
		seen[string(p.Kind)+"|"+p.Key]++
	}
	for _, p := range want {
		if seen[string(p.Kind)+"|"+p.Key] != 1 {
			return false
		}
	}
	return true
}

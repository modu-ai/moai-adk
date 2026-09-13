package constitution

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

// SPEC-CON-AMEND-APPLY-001 M4 — fault injection through the two per-pipeline
// seams of REQ-CAA-011 (AC-CAA-012, 013, 021). Each test asserts the seam's
// call count next to its byte-identity check: an injector that never fires
// and a correct restore look identical otherwise (plan.md §H).

// renameInjector fails call failOn of the forward rename and delegates every
// other call to os.Rename. In rename-then-fail mode the failing call renames
// first, records whether the destination exists, and then reports failure.
type renameInjector struct {
	failOn        int
	applyThenFail bool
	calls         int
	destExisted   bool
	dests         []string
}

func (r *renameInjector) rename(oldpath, newpath string) error {
	r.calls++
	r.dests = append(r.dests, newpath)
	if r.calls != r.failOn {
		return os.Rename(oldpath, newpath)
	}
	if r.applyThenFail {
		if err := os.Rename(oldpath, newpath); err != nil {
			return err
		}
		_, err := os.Stat(newpath)
		r.destExisted = err == nil
		return errors.New("injected failure after the rename took effect")
	}
	return errors.New("injected failure before renaming")
}

// pathSet returns the sorted relative paths under root.
func pathSet(t *testing.T, root string) []string {
	t.Helper()
	var out []string
	for k := range snapshotTree(t, root) {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

// amendArtifacts returns every temporary or backup file the apply step
// created under root.
func amendArtifacts(t *testing.T, root, kind string) []string {
	t.Helper()
	var out []string
	err := filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && strings.Contains(d.Name(), ".amend-"+kind+"-") {
			out = append(out, p)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return out
}

// AC-CAA-012 — a failed rename restores all three files.
func TestApply_RenameFault_RestoresAll(t *testing.T) {
	cases := []struct {
		name          string
		failOn        int
		applyThenFail bool
		logExists     bool
		failedFile    func(prj project) string
	}{
		{"first_rename_applied", 1, true, true, func(prj project) string { return prj.rule }},
		{"second_rename", 2, false, true, func(prj project) string { return prj.registry }},
		{"third_rename", 3, false, true, func(prj project) string { return prj.log }},
		{"third_rename_applied", 3, true, true, func(prj project) string { return prj.log }},
		{"third_rename_log_absent", 3, true, false, func(prj project) string { return prj.log }},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			isolateEnv(t)
			dir := t.TempDir()
			prj := buildProject(t, dir, standardEntries(fxRuleFile), fxRuleFile, fxRuleBody, tc.logExists)
			if err := os.MkdirAll(filepath.Dir(prj.log), 0o755); err != nil {
				t.Fatal(err)
			}
			before := snapshotTree(t, dir)
			p, _, lockDir := newGatedPipeline(t)
			inj := &renameInjector{failOn: tc.failOn, applyThenFail: tc.applyThenFail}
			p.rename = inj.rename

			_, err := p.Execute(proposal(fxBefore, fxAfter), dir, false)
			if err == nil {
				t.Fatal("Execute: want the injected rename failure, got nil")
			}
			if !strings.Contains(err.Error(), "rename") || !containsPathForm(err.Error(), tc.failedFile(prj)) {
				t.Errorf("error %q does not name the failed rename of %s", err, tc.failedFile(prj))
			}
			if inj.calls != tc.failOn {
				t.Errorf("rename seam called %d times, want %d (reachability)", inj.calls, tc.failOn)
			}
			if tc.applyThenFail && !inj.destExisted {
				t.Error("rename-then-fail: the destination did not exist after the delegated rename")
			}
			assertSameTree(t, tc.name, dir, before)
			assertLockReleased(t, lockDir)
		})
	}

	t.Run("no_fault_clean", func(t *testing.T) {
		isolateEnv(t)
		dir := t.TempDir()
		standardProject(t, dir)
		pathsBefore := pathSet(t, dir)
		p, _, _ := newGatedPipeline(t)
		inj := &renameInjector{}
		p.rename = inj.rename

		if _, err := p.Execute(proposal(fxBefore, fxAfter), dir, false); err != nil {
			t.Fatalf("Execute: %v", err)
		}
		if inj.calls != 3 {
			t.Errorf("rename seam called %d times, want 3", inj.calls)
		}
		if got := pathSet(t, dir); strings.Join(got, "\n") != strings.Join(pathsBefore, "\n") {
			t.Errorf("path set changed after a clean apply:\nbefore %v\nafter  %v", pathsBefore, got)
		}
	})
}

// AC-CAA-013 — the rename order is source, registry, log.
func TestApply_RenameOrder(t *testing.T) {
	isolateEnv(t)
	dir := t.TempDir()
	prj := standardProject(t, dir)
	p, _, _ := newGatedPipeline(t)
	inj := &renameInjector{}
	p.rename = inj.rename

	if _, err := p.Execute(proposal(fxBefore, fxAfter), dir, false); err != nil {
		t.Fatalf("Execute: %v", err)
	}
	want := []string{prj.rule, prj.registry, prj.log}
	if len(inj.dests) != len(want) {
		t.Fatalf("renames = %v, want exactly %v", inj.dests, want)
	}
	for i := range want {
		if filepath.Clean(inj.dests[i]) != filepath.Clean(want[i]) {
			t.Errorf("rename %d destination = %s, want %s", i+1, inj.dests[i], want[i])
		}
	}
}

// AC-CAA-021 — a failed restore keeps the backups and names them.
func TestApply_RestoreFault_KeepsBackups(t *testing.T) {
	isolateEnv(t)
	dir := t.TempDir()
	prj := standardProject(t, dir)
	pre := map[string]string{
		prj.rule:     fileSHA(t, prj.rule),
		prj.registry: fileSHA(t, prj.registry),
		prj.log:      fileSHA(t, prj.log),
	}
	p, _, lockDir := newGatedPipeline(t)
	inj := &renameInjector{failOn: 2}
	p.rename = inj.rename
	restoreCalls := 0
	p.restore = func(path string, data []byte, existed bool) error {
		restoreCalls++
		if restoreCalls == 1 {
			return errors.New("injected restore failure")
		}
		return restoreFile(path, data, existed)
	}

	_, err := p.Execute(proposal(fxBefore, fxAfter), dir, false)
	if err == nil {
		t.Fatal("Execute: want the restore failure, got nil")
	}
	if !strings.Contains(err.Error(), "restore") {
		t.Errorf("error %q does not name the failed restore step", err)
	}
	backups := amendArtifacts(t, dir, "bak")
	if len(backups) != 3 {
		t.Fatalf("backup files kept = %d (%v), want 3", len(backups), backups)
	}
	for _, b := range backups {
		if !containsPathForm(err.Error(), b) {
			t.Errorf("error %q does not contain the backup path %s", err, b)
		}
		base := strings.TrimPrefix(filepath.Base(b), ".")
		base = base[:strings.Index(base, ".amend-bak-")]
		target := filepath.Join(filepath.Dir(b), base)
		if want, ok := pre[target]; !ok {
			t.Errorf("backup %s does not belong to one of the three files", b)
		} else if fileSHA(t, b) != want {
			t.Errorf("backup %s does not hold the pre-apply bytes of %s", b, target)
		}
	}
	if inj.calls != 2 {
		t.Errorf("rename seam called %d times, want 2", inj.calls)
	}
	if restoreCalls < 1 {
		t.Errorf("restore seam called %d times, want at least 1", restoreCalls)
	}
	if temps := amendArtifacts(t, dir, "tmp"); len(temps) != 0 {
		t.Errorf("temporary files remain: %v", temps)
	}
	assertLockReleased(t, lockDir)
}

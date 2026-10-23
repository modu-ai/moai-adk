package worktreeroot

// review_fix_test.go — card t1640 M8S, finding F9: the M8R repair
// functions carry package-local tests (the sync audit measured Plan,
// RecordedKey and OwnerMarkerOf at 0.0% package-local coverage; a
// regression in any of them would otherwise surface only through
// consumer-package tests).

import (
	"crypto/sha256"
	"os"
	"path/filepath"
	"testing"
)

// rfxWalkHash hashes a directory tree (absent dirs hash as absent) — the
// read-only contract's before/after comparison.
func rfxWalkHash(t *testing.T, dir string) string {
	t.Helper()
	sum := sha256.New()
	_ = filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		rel, rerr := filepath.Rel(dir, path)
		if rerr != nil {
			return rerr
		}
		info, ierr := d.Info()
		if ierr != nil {
			return ierr
		}
		_, _ = sum.Write([]byte(rel + "|" + info.Mode().String() + "\x00"))
		if !info.IsDir() {
			body, rerr := os.ReadFile(path)
			if rerr != nil {
				return rerr
			}
			_, _ = sum.Write(body)
			_, _ = sum.Write([]byte("\x00"))
		}
		return nil
	})
	return string(sum.Sum(nil))
}

// TestPlan_IsReadOnly pins the read-only contract (card t1640 review P2i):
// Plan answers the landing path without creating the project folder, its
// owner marker, or the repository-side record — the home root and the
// repository stay unchanged.
func TestPlan_IsReadOnly(t *testing.T) {
	home := wrHome(t)
	repo := wrRepo(t, "plan-repo")
	projects := wrProjects(home)

	before := rfxWalkHash(t, projects)
	beforeRepo := rfxWalkHash(t, filepath.Dir(repo))
	got, err := Plan(repo, "t9")
	if err != nil {
		t.Fatalf("Plan: %v", err)
	}
	want := filepath.Join(projects, "plan-repo", "t9")
	if got != want {
		t.Errorf("Plan = %q, want %q", got, want)
	}
	if after := rfxWalkHash(t, projects); after != before {
		t.Errorf("Plan wrote inside the home root (hash changed)")
	}
	if after := rfxWalkHash(t, filepath.Dir(repo)); after != beforeRepo {
		t.Errorf("Plan wrote inside the repository (hash changed)")
	}
	// The recorded key is still absent: Plan never claims.
	if k := RecordedKey(repo); k != "" {
		t.Errorf("RecordedKey = %q after a Plan; Plan must not write the record", k)
	}
}

// TestPlan_UsesTheRecordedKey pins the recorded-key arm: once Resolve has
// claimed the key, Plan answers under it.
func TestPlan_UsesTheRecordedKey(t *testing.T) {
	wrHome(t)
	repo := wrRepo(t, "recorded-repo")
	res, err := Resolve(repo)
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	got, err := Plan(repo, "t9")
	if err != nil {
		t.Fatalf("Plan: %v", err)
	}
	if want := filepath.Join(res.Root, "t9"); got != want {
		t.Errorf("Plan = %q, want %q (the recorded key's folder)", got, want)
	}
}

// TestPlan_ComputesTheCollisionSuffix pins the first-time collision arm: a
// candidate folder owned by another repository is never claimed by Plan;
// the answer is the AC-WHR-003a suffixed key, computed without writing.
func TestPlan_ComputesTheCollisionSuffix(t *testing.T) {
	home := wrHome(t)
	projects := wrProjects(home)
	owner := wrRepo(t, "proj")
	res, err := Resolve(owner)
	if err != nil {
		t.Fatalf("Resolve owner: %v", err)
	}
	if res.Key != "proj" {
		t.Fatalf("the owner repository's key = %q, want proj", res.Key)
	}
	arriver := wrRepo(t, "proj") // a second repository with the same basename
	canonical := wrReal(t, arriver)
	got, err := Plan(arriver, "t9")
	if err != nil {
		t.Fatalf("Plan: %v", err)
	}
	want := filepath.Join(projects, "proj-"+suffixHash(canonical), "t9")
	if got != want {
		t.Errorf("Plan = %q, want the suffixed candidate %q", got, want)
	}
	// Plan created nothing: only the owner's folder exists.
	entries, rerr := os.ReadDir(projects)
	if rerr != nil {
		t.Fatalf("read projects: %v", rerr)
	}
	if len(entries) != 1 {
		t.Errorf("projects has %d folders, want only the owner's (Plan must not create)", len(entries))
	}
}

// TestRecordedKey_ReportsTheRepositorySideRecord pins the read-only key
// observation: absent before any resolution, the claimed key after it, and
// empty (not an error) for a non-repository directory.
func TestRecordedKey_ReportsTheRepositorySideRecord(t *testing.T) {
	home := wrHome(t)
	_ = home
	repo := wrRepo(t, "keyed-repo")
	if k := RecordedKey(repo); k != "" {
		t.Fatalf("RecordedKey = %q before any resolution, want empty", k)
	}
	res, err := Resolve(repo)
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	if k := RecordedKey(repo); k != res.Key {
		t.Errorf("RecordedKey = %q, want the recorded key %q", k, res.Key)
	}
	if k := RecordedKey(t.TempDir()); k != "" {
		t.Errorf("RecordedKey of a non-repository = %q, want empty", k)
	}
}

// TestOwnerMarkerOf_ReportsTheOwner pins the marker read: the owner of a
// claimed project folder is the repository's canonical primary root, and a
// folder without a marker reports absent.
func TestOwnerMarkerOf_ReportsTheOwner(t *testing.T) {
	home := wrHome(t)
	repo := wrRepo(t, "owned-repo")
	res, err := Resolve(repo)
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	canonical := wrReal(t, repo)
	owner, ok := OwnerMarkerOf(filepath.Join(wrProjects(home), res.Key))
	if !ok {
		t.Fatalf("OwnerMarkerOf reports no marker for the claimed folder")
	}
	if owner != canonical {
		t.Errorf("OwnerMarkerOf = %q, want the canonical primary root %q", owner, canonical)
	}
	if _, ok := OwnerMarkerOf(filepath.Join(wrProjects(home), "no-such-folder")); ok {
		t.Errorf("OwnerMarkerOf reports a marker for an absent folder")
	}
}

// TestPlan_ExistingFolderOwnedBySelf covers the EqualFold arm: a candidate
// folder this repository already owns answers with that folder, read-only.
func TestPlan_ExistingFolderOwnedBySelf(t *testing.T) {
	home := wrHome(t)
	repo := wrRepo(t, "self-owned")
	res, err := Resolve(repo)
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	before := rfxWalkHash(t, wrProjects(home))
	got, err := Plan(repo, "t9")
	if err != nil {
		t.Fatalf("Plan: %v", err)
	}
	if want := filepath.Join(res.Root, "t9"); got != want {
		t.Errorf("Plan = %q, want %q", got, want)
	}
	if after := rfxWalkHash(t, wrProjects(home)); after != before {
		t.Errorf("Plan wrote inside the home root (hash changed)")
	}
}

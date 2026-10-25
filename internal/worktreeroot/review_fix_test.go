package worktreeroot

// review_fix_test.go — card t1640 M8S, finding F9: the M8R repair
// functions carry package-local tests (the sync audit measured Plan,
// RecordedKey and OwnerMarkerOf at 0.0% package-local coverage; a
// regression in any of them would otherwise surface only through
// consumer-package tests).

import (
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"testing"
)

// wrTreeHash hashes every path under dir (absent dirs hash as absent) into
// a per-path map: key = path relative to dir, value = [mode, contentSHA256]
// (the content half is empty for directories). The map form lets a failing
// read-only check name the exact paths that moved and which half of the
// hash moved (card t1954 — the CI failure said "hash changed" and nothing
// else, which left the changing path unnamed).
func wrTreeHash(t *testing.T, dir string) map[string][2]string {
	t.Helper()
	out := map[string][2]string{}
	_ = filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
		rel, rerr := filepath.Rel(dir, path)
		if rerr != nil {
			return rerr
		}
		var entry [2]string
		switch {
		case err != nil:
			// The walk itself failed for this path (e.g. an unreadable
			// directory): keep the path in the map with the reason, so a
			// Plan-created artifact can never escape the diff by being
			// un-readable (audit F1 — the base fold kept rel+mode here).
			entry[0] = "WALKERR:" + err.Error()
		default:
			info, ierr := d.Info()
			if ierr != nil {
				entry[0] = "INFOERR:" + ierr.Error()
				break
			}
			entry[0] = info.Mode().String()
			if !info.IsDir() {
				body, rerr := os.ReadFile(path)
				if rerr != nil {
					entry[1] = "READERR:" + rerr.Error()
					break
				}
				sum := sha256.Sum256(body)
				entry[1] = fmt.Sprintf("%x", sum)
			}
		}
		out[rel] = entry
		return nil
	})
	return out
}

// wrTreeDiff formats the before/after difference of two wrTreeHash maps as
// REMOVED/CHANGED/ADDED lines (sorted, deterministic), naming mode-vs-
// content for a changed entry; empty string when nothing moved.
func wrTreeDiff(before, after map[string][2]string) string {
	rels := map[string]bool{}
	for rel := range before {
		rels[rel] = true
	}
	for rel := range after {
		rels[rel] = true
	}
	sorted := make([]string, 0, len(rels))
	for rel := range rels {
		sorted = append(sorted, rel)
	}
	sort.Strings(sorted)
	lines := []string{}
	for _, rel := range sorted {
		b, inB := before[rel]
		a, inA := after[rel]
		switch {
		case inB && !inA:
			lines = append(lines, "\n  REMOVED "+rel)
		case !inB && inA:
			lines = append(lines, "\n  ADDED "+rel)
		case b != a:
			what := "content"
			if b[0] != a[0] && b[1] != a[1] {
				what = "mode+content"
			} else if b[0] != a[0] {
				what = "mode"
			}
			lines = append(lines, fmt.Sprintf("\n  CHANGED %s (%s)", rel, what))
		}
	}
	return strings.Join(lines, "")
}

// rfxWalkHash hashes a directory tree (absent dirs hash as absent) — the
// read-only contract's before/after comparison, folded over wrTreeHash.
func rfxWalkHash(t *testing.T, dir string) string {
	entries := wrTreeHash(t, dir)
	sum := sha256.New()
	rels := make([]string, 0, len(entries))
	for rel := range entries {
		rels = append(rels, rel)
	}
	sort.Strings(rels)
	for _, rel := range rels {
		entry := entries[rel]
		_, _ = sum.Write([]byte(rel + "|" + entry[0] + "\x00" + entry[1] + "\x00"))
	}
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
	beforeRepo := wrTreeHash(t, filepath.Dir(repo))
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
	// The failing half names the exact paths (card t1954): the assertion and
	// its scope are unchanged — only the failure message gained the diff.
	if diff := wrTreeDiff(beforeRepo, wrTreeHash(t, filepath.Dir(repo))); diff != "" {
		t.Errorf("Plan wrote inside the repository (hash changed):%s", diff)
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

// TestWalkHash_KeepsReadFailureEntries pins the audit F1 repair (card
// t1954): a path that cannot be read stays in the wrTreeHash map as a
// READERR sentinel instead of being dropped, so a Plan-created unreadable
// artifact cannot escape the before/after diff by being un-readable. The
// mutation lives entirely inside the fixture's own t.TempDir — the real
// working tree is never touched.
func TestWalkHash_KeepsReadFailureEntries(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("chmod-000 unreadability is POSIX-only")
	}
	repo := wrRepo(t, "readerr-repo")
	before := wrTreeHash(t, repo)

	secret := filepath.Join(repo, "unreadable.bin")
	if err := os.WriteFile(secret, []byte("unreachable\n"), 0o600); err != nil {
		t.Fatalf("write secret: %v", err)
	}
	if err := os.Chmod(secret, 0o000); err != nil {
		t.Fatalf("chmod 000: %v", err)
	}
	t.Cleanup(func() { _ = os.Chmod(secret, 0o644) })

	after := wrTreeHash(t, repo)
	entry, ok := after["unreadable.bin"]
	if !ok {
		t.Fatal("the unreadable file is missing from the map — read-failure entries are being dropped (audit F1)")
	}
	if !strings.Contains(entry[1], "READERR:") {
		t.Errorf("entry[1] = %q, want a READERR sentinel", entry[1])
	}
	if diff := wrTreeDiff(before, after); !strings.Contains(diff, "unreadable.bin") {
		t.Errorf("wrTreeDiff = %q, want the unreadable file named", diff)
	}
}

package worktreeroot

// t1954 local probe — NOT committed. Reproduces the TestPlan_IsReadOnly
// fixture and dumps a path-level before/after diff of the repo parent
// (what rfxWalkHash covers) so the changing path can be named.

import (
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func prWalk(t *testing.T, dir string) map[string]string {
	t.Helper()
	out := map[string]string{}
	_ = filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			out[path] = fmt.Sprintf("WALKERR:%v", err)
			return nil
		}
		info, ierr := d.Info()
		if ierr != nil {
			out[path] = fmt.Sprintf("INFOERR:%v", ierr)
			return nil
		}
		sum := sha256.Sum256([]byte(info.Mode().String()))
		if !info.IsDir() {
			body, rerr := os.ReadFile(path)
			if rerr != nil {
				out[path] = fmt.Sprintf("READERR:%v", rerr)
				return nil
			}
			h := sha256.Sum256(body)
			sum = sha256.Sum256(append([]byte(info.Mode().String()), h[:]...))
		}
		out[path] = fmt.Sprintf("%x", sum)
		return nil
	})
	return out
}

func TestProbePlanRepoDiff(t *testing.T) {
	if cfg := os.Getenv("T1954_GITCONFIG"); cfg != "" {
		t.Setenv("GIT_CONFIG_SYSTEM", cfg)
		t.Logf("PROBE: simulating runner gitconfig via %s", cfg)
	}
	home := wrHome(t)
	repo := wrRepo(t, "plan-repo")
	projects := wrProjects(home)

	before := prWalk(t, filepath.Dir(repo))
	if _, err := Plan(repo, "t9"); err != nil {
		t.Fatalf("Plan: %v", err)
	}
	after := prWalk(t, filepath.Dir(repo))

	changed := 0
	for p, h := range before {
		h2, ok := after[p]
		if !ok {
			t.Logf("REMOVED: %s", p)
			changed++
			continue
		}
		if h != h2 {
			t.Logf("CHANGED: %s", p)
			changed++
		}
	}
	for p := range after {
		if _, ok := before[p]; !ok {
			t.Logf("ADDED: %s", p)
			changed++
		}
	}
	t.Logf("PROBE: %d path(s) changed; projects=%s", changed, projects)
}

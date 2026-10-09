// m7_o_excl_retry_test.go — gate round 42 [P3]: the pinned parent's
// exclusive temp creation retries with a FRESH name on EEXIST — the
// former form tested the stale outer err (always nil), so a real EEXIST
// collision failed the install on the first call instead of retrying.
//
// The contract is exercised under CONTENTION: N concurrent confinedWrite
// calls to the SAME parent each draw exclusive temp names — any broken
// EEXIST handling (break-on-first-call) surfaces as a failed write under
// the racing name draws.
package userassets

import (
	"os"
	"path/filepath"
	"sync"
	"testing"
)

func TestConfinedWriteExclusiveCreationSurvivesContention(t *testing.T) {
	home := t.TempDir()
	root, err := resolveRoot(home, Root{Slug: RootClaudeSkills, Dir: filepath.Join(home, ".claude", "skills")})
	if err != nil {
		t.Fatal(err)
	}
	parent := filepath.Join(root.dir, "sub")
	if err := os.MkdirAll(parent, 0o755); err != nil {
		t.Fatal(err)
	}
	inst := &Installer{Home: home, MoaiVersion: "test"}

	const n = 24
	var wg sync.WaitGroup
	errs := make([]error, n)
	for i := 0; i < n; i++ {
		i := i
		wg.Add(1)
		go func() {
			defer wg.Done()
			errs[i] = inst.confinedWrite(root, "sub/file-"+itoaTest(i)+".txt", []byte("payload\n"), true)
		}()
	}
	wg.Wait()
	for i, err := range errs {
		if err != nil {
			t.Fatalf("confinedWrite %d failed under contention: %v", i, err)
		}
	}
	for i := 0; i < n; i++ {
		if _, err := os.Stat(filepath.Join(parent, "file-"+itoaTest(i)+".txt")); err != nil {
			t.Fatalf("write %d did not land: %v", i, err)
		}
	}
}

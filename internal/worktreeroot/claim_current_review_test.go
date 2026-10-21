package worktreeroot

import (
	"fmt"
	"path/filepath"
	"strings"
	"testing"
)

func TestConcurrentProjectClaimsKeepDistinctOwners(t *testing.T) {
	home := t.TempDir()
	type result struct {
		key, owner string
		err        error
	}
	const count = 32
	results := make(chan result, count)
	start := make(chan struct{})
	for i := range count {
		owner := filepath.Join(t.TempDir(), fmt.Sprintf("owner-%02d", i))
		candidate := "shared"
		if i%2 == 1 {
			candidate = "SHARED"
		}
		go func() {
			<-start
			key, err := claim(home, candidate, owner, false, candidate)
			results <- result{key, owner, err}
		}()
	}
	close(start)
	seen := map[string]string{}
	for range count {
		r := <-results
		if r.err != nil {
			t.Errorf("claim for %s failed: %v", r.owner, r.err)
			continue
		}
		folded := strings.ToLower(r.key)
		if previous, exists := seen[folded]; exists && previous != r.owner {
			t.Errorf("distinct repositories claimed case-folded key %s: %s and %s", r.key, previous, r.owner)
		}
		seen[folded] = r.owner
		if owner, err := readMarkerOwner(filepath.Join(projectDir(home, r.key), markerName)); err != nil || owner != r.owner {
			t.Errorf("claim %s lost its owner marker: got %q, %v; want %s", r.key, owner, err, r.owner)
		}
	}
}

func TestRecordedProjectKeyCannotAdoptAnotherRepositoryMarker(t *testing.T) {
	wrHome(t)
	first := wrRepo(t, "shared")
	res, err := Resolve(first)
	if err != nil {
		t.Fatal(err)
	}
	second := wrRepo(t, "different")
	common, err := gitCommonDir(second)
	if err != nil {
		t.Fatal(err)
	}
	// A persisted key may arrive from an earlier, separate MOAI_HOME.
	if err := writeRecord(common, res.Key); err != nil {
		t.Fatal(err)
	}
	if adopted, err := Resolve(second); err == nil {
		t.Fatalf("recorded key adopted another repository's project folder: %+v", adopted)
	}
	if owner, err := readMarkerOwner(filepath.Join(res.Root, markerName)); err != nil || owner != first {
		t.Fatalf("refusal changed the first repository's marker: %q, %v", owner, err)
	}
}

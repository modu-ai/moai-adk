package hook

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
)

// THIS FILE GOING RED IS A NORMAL OUTCOME, NOT A REGRESSION.
//
// It is card t971's reproduction of a defect that is still present. A red run
// here means the reproduction still reproduces; a green run means this run did
// not hit the window (see the timing note below), NOT that the defect is gone.
// Do not "fix" the test. The branch carrying it (WT-index-write-race) is
// deliberately unmerged for exactly this reason.
//
// THE REPAIR, so the next reader does not have to find it again:
// serialize the read-modify-write in ensureIndexLine (agentmemory.go:423)
// with internal/lockfile, which already exists and is already used by
// internal/kanban/board_lock.go, internal/cli/settings.go and
// internal/cli/taskledger. No new dependency is needed. A repair card takes
// this branch as its base and turns this test green.
//
// TestEnsureIndexLineConcurrentAppendsLoseLines is a REPRODUCTION, not a
// guard: it is expected to FAIL against the current implementation, and its
// failure is the measurement card t971 was opened to take.
//
// ensureIndexLine reads the whole index, appends one line, and writes the
// whole file back (agentmemory.go:428-443) with nothing serializing the three
// steps. Two callers that interleave between one's read and its write both
// write a body derived from the same starting content, so the later write
// drops the earlier caller's line. The topic file it indexed is already on
// disk by then, which is why the loss presents as an orphan rather than as an
// error — nothing reports it, and the caller's `added == true` says the line
// was written.
//
// The assertion is deliberately the honest one — every line that was reported
// as added is still present — rather than a threshold tuned to pass. A run
// that loses nothing is possible (the interleaving is timing-dependent), so a
// green result here is NOT evidence the race is absent; it is evidence this
// run did not hit it. The loop count is raised to make the window likely, not
// to make the failure certain.
func TestEnsureIndexLineConcurrentAppendsLoseLines(t *testing.T) {
	const lanes = 16

	primaryRoot := t.TempDir()
	agent := "moai"
	agentDir := filepath.Join(primaryRoot, ".claude", "agent-memory", agent)
	if err := os.MkdirAll(agentDir, 0o755); err != nil {
		t.Fatalf("mkdir agent dir: %v", err)
	}

	var (
		wg      sync.WaitGroup
		mu      sync.Mutex
		added   []string
		failed  []error
		indexPt = filepath.Join(agentDir, agentMemoryIndexName)
	)

	for i := 0; i < lanes; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			destName := fmt.Sprintf("topic_%02d.md", i)
			line := fmt.Sprintf("- [topic %02d](%s) — lane %02d", i, destName, i)

			// The topic file lands first, exactly as the production path does:
			// the index line is reconciliation after the content is already
			// written. This is what makes a lost line an orphan.
			if err := os.WriteFile(filepath.Join(agentDir, destName), []byte("body\n"), 0o644); err != nil {
				mu.Lock()
				failed = append(failed, err)
				mu.Unlock()
				return
			}

			ok, err := ensureIndexLine(primaryRoot, agent, line, destName, true)
			mu.Lock()
			defer mu.Unlock()
			if err != nil {
				failed = append(failed, err)
				return
			}
			if ok {
				added = append(added, destName)
			}
		}(i)
	}
	wg.Wait()

	for _, err := range failed {
		t.Fatalf("ensureIndexLine returned an error: %v", err)
	}

	body, err := os.ReadFile(indexPt)
	if err != nil {
		t.Fatalf("read index: %v", err)
	}
	content := string(body)

	var missing []string
	for _, destName := range added {
		if !indexLinksTarget(content, destName) {
			missing = append(missing, destName)
		}
	}

	topics, err := filepath.Glob(filepath.Join(agentDir, "topic_*.md"))
	if err != nil {
		t.Fatalf("glob topics: %v", err)
	}

	t.Logf("lanes=%d reported-added=%d index-lines=%d topic-files=%d lost=%d",
		lanes, len(added), strings.Count(content, "\n- ["), len(topics), len(missing))

	if len(missing) > 0 {
		t.Fatalf("index lost %d of %d reported-added lines (orphaned topics still on disk): %v",
			len(missing), len(added), missing)
	}
}

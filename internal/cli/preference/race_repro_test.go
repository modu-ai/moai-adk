package preference

// M1 reproduction test (SPEC-CI-FLAKY-STABILIZE-001, REQ-CFS-020).
//
// TEMPORARY: this file is removed in M2 once the TestMain warm-up lands. Its
// full source and the observed failure output are preserved verbatim in
// .moai/specs/SPEC-CI-FLAKY-STABILIZE-001/progress.md §E.2 (plan.md §G AP-3).

import (
	"sync"
	"testing"
)

// TestRaceRepro_PreferenceCmdLazySort reproduces the cobra lazy-sort data race
// on the real package-level PreferenceCmd global — the exact defect observed in
// GitHub Actions run 30135630413 attempt 1 (macos-latest).
//
// cobra v1.10.2 command.go:1332-1339 sorts c.commands in place INSIDE the
// Commands() accessor and writes c.commandsAreSorted there. Goroutines making
// the FIRST Commands() call concurrently race that write against each other's
// read, so the window exists only until the flag flips to true.
//
// MUST run in isolation so this test owns the first Commands() call in the
// process (a package-wide run lets an earlier test close the window first):
//
//	for i in $(seq 1 20); do \
//	  go test -race -count=1 -run '^TestRaceRepro_PreferenceCmdLazySort$' ./internal/cli/preference/; \
//	done
func TestRaceRepro_PreferenceCmdLazySort(t *testing.T) {
	const goroutines = 16

	start := make(chan struct{})
	var wg sync.WaitGroup
	wg.Add(goroutines)
	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			<-start // release all goroutines at once to widen the race window
			_ = PreferenceCmd.Commands()
		}()
	}
	close(start)
	wg.Wait()
}

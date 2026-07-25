package cli

// M1 reproduction test (SPEC-CI-FLAKY-STABILIZE-001, REQ-CFS-020).
//
// TEMPORARY: this file is removed in M2 once the TestMain warm-up lands. Its
// full source and the observed output are preserved verbatim in
// .moai/specs/SPEC-CI-FLAKY-STABILIZE-001/progress.md §E.2 (plan.md §G AP-3).

import (
	"sync"
	"testing"
)

// TestRaceRepro_RootCmdLazySort is the internal/cli counterpart of the
// preference-package reproduction: concurrent first Commands() calls on the
// real rootCmd global, which is the receiver used by the parallel tests in
// harness_retirement_test.go / handoff_test.go.
//
// MUST run in isolation (see the preference-package test for the rationale):
//
//	for i in $(seq 1 20); do \
//	  go test -race -count=1 -run '^TestRaceRepro_RootCmdLazySort$' ./internal/cli/; \
//	done
func TestRaceRepro_RootCmdLazySort(t *testing.T) {
	const goroutines = 16

	start := make(chan struct{})
	var wg sync.WaitGroup
	wg.Add(goroutines)
	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			<-start
			_ = rootCmd.Commands()
		}()
	}
	close(start)
	wg.Wait()
}

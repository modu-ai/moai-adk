// mcp_state_write_mutex_test.go — SPEC-MCP-NEUTRALITY-001 M3 (card t1632,
// REQ-MN-010's serialization half): concurrent tool calls serialize their
// project-state writes through ONE server-wide mutex. The test fires
// goal_arm + verify_snapshot(record) + audit_cache(store) concurrently over
// one fixture root and requires zero errors — run under -race in the race
// lane, the interleaved writers must not corrupt or double-write any
// artifact.
package cli

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"
)

func TestMCPStateWriteMutex_ConcurrentWritersSerialize(t *testing.T) {
	root := sdServerFixture(t)
	// REQ-MN-020: a store carries an existing verdict file as its evidence.
	verdict := filepath.Join(root, "verdict.md")
	if err := os.WriteFile(verdict, []byte("# verdict\n\nPASS\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	var wg sync.WaitGroup
	errs := make(chan error, 30)
	for i := 0; i < 10; i++ {
		// goal_arm: goal state under .moai/state/goal.
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			req := sdToolReq(map[string]any{
				"condition":    fmt.Sprintf("cmd: echo goal-%d exits 0", i),
				"session_id":   fmt.Sprintf("sess-mutex-%d", i),
				"max_turns":    5,
				"project_root": root,
			})
			res, err := handleGoalArm(context.Background(), req)
			if err != nil || res.IsError {
				errs <- fmt.Errorf("goal_arm %d: %s %v", i, mcpTextOf(res), err)
			}
		}(i)
		// verify_snapshot record: snapshot under .moai/state/verify.
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			req := sdToolReq(map[string]any{
				"key":          fmt.Sprintf("mutex-key-%d:verify", i),
				"command":      fmt.Sprintf("go test -run Mutex%d", i),
				"exit_code":    0,
				"project_root": root,
			})
			res, err := handleVerifySnapshot(context.Background(), req)
			if err != nil || res.IsError {
				errs <- fmt.Errorf("verify_snapshot %d: %s %v", i, mcpTextOf(res), err)
			}
		}(i)
		// audit_cache store: the process-shared cache write.
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			req := sdToolReq(map[string]any{
				"op":           "store",
				"spec_id":      fmt.Sprintf("SPEC-MUTEX-%03d", i),
				"hash":         fmt.Sprintf("hash-%d", i),
				"report_path":  verdict,
				"project_root": root,
			})
			res, err := handleAuditCache(context.Background(), req)
			if err != nil || res.IsError {
				errs <- fmt.Errorf("audit_cache %d: %s %v", i, mcpTextOf(res), err)
			}
		}(i)
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		t.Error(err)
	}
	// The goal state survived: one goal file per armed session.
	entries, err := os.ReadDir(filepath.Join(root, ".moai", "state", "goal"))
	if err != nil && !os.IsNotExist(err) {
		t.Fatalf("goal state read: %v", err)
	}
	if err == nil && len(entries) < 10 {
		t.Errorf("goal state entries = %d, want >= 10 (one per armed session)", len(entries))
	}
}

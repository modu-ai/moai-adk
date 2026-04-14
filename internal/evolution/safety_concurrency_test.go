package evolution_test

import (
	"sync"
	"testing"

	"github.com/modu-ai/moai-adk/internal/evolution"
)

// TestUpdateRateLimit_ConcurrentSafety verifies that concurrent calls to
// UpdateRateLimit do not produce data races and the final count is correct.
// I2: Read→mutate→Write TOCTOU 경쟁 조건 수정 검증
func TestUpdateRateLimit_ConcurrentSafety(t *testing.T) {
	t.Parallel()

	projectRoot := t.TempDir()
	mustInitMoAI(t, projectRoot)

	const goroutines = 100
	var wg sync.WaitGroup
	wg.Add(goroutines)

	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			_ = evolution.UpdateRateLimit(projectRoot, "")
		}()
	}
	wg.Wait()

	state, err := evolution.ReadRateState(projectRoot)
	if err != nil {
		t.Fatalf("ReadRateState: %v", err)
	}

	if state.ProposalsThisWeek != goroutines {
		t.Errorf("ProposalsThisWeek = %d, want %d (경쟁 조건으로 일부 업데이트 손실)",
			state.ProposalsThisWeek, goroutines)
	}
}

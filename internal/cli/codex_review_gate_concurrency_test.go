package cli

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/hook"
	"github.com/modu-ai/moai-adk/internal/verify"
)

// TestReviewGate_ConcurrentBlockCap — REQ-RG-004 (AC-RG-004, AC-RG-005, AC-RG-017)
// under concurrent turn ends. Several turn ends for one fail key can reach the
// gate at the same moment, in separate hook processes. The per-key block count
// is a total of C = codexReviewBlockLimit blocks, so no round of simultaneous
// turn ends may block more than C times. Each round releases workers together
// through a barrier (a closed channel) and counts the block decisions.
//
// Before the repair, the count read, the cap check, and the count write were
// separate steps with no lock between them, so concurrent turn ends read the
// same count and all blocked.
func TestReviewGate_ConcurrentBlockCap(t *testing.T) {
	const (
		rounds  = 100
		workers = 8
	)

	root := m3GateFixture(t)
	state := m3ReceiptState(t, root)
	now := time.Now()
	fail := verify.Receipt{
		CheckID:      codexReviewCheckID,
		Head:         state.Head,
		TreeDigest:   state.TreeDigest,
		ConfigDigest: state.ConfigDigest,
		Command:      state.Command,
		ToolVersion:  state.ToolVersion,
		ExitCode:     1,
		Verdict:      codexReviewVerdictFail,
		RecordedAt:   now,
	}

	overCap := 0
	maxBlocks := 0
	for round := 0; round < rounds; round++ {
		dir := t.TempDir()
		start := make(chan struct{})
		var blocks atomic.Int32
		var wg sync.WaitGroup
		for w := 0; w < workers; w++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				<-start
				if out := codexReviewGateFail(dir, state, fail, now); out.Decision == hook.DecisionBlock {
					blocks.Add(1)
				}
			}()
		}
		close(start)
		wg.Wait()

		got := int(blocks.Load())
		if got > maxBlocks {
			maxBlocks = got
		}
		if got > codexReviewBlockLimit {
			overCap++
		}
	}
	if overCap > 0 {
		t.Fatalf("concurrent turn ends for one fail key blocked more than the cap (C = %d) in %d of %d rounds; %d simultaneous turn ends per round, max blocks in one round = %d",
			codexReviewBlockLimit, overCap, rounds, workers, maxBlocks)
	}
}

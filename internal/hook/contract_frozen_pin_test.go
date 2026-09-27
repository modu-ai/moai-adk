package hook

import (
	"slices"
	"testing"

	"github.com/modu-ai/moai-adk/internal/contract"
)

// TestFrozenInstructionFilesPinnedInContract pins internal/contract's copy of
// the frozen instruction basenames to frozenInstructionFiles. The contract
// verification core must not import this package, so it carries its own copy
// (its `frozen-files` invariant emits them as `**/<basename>` globs); this test
// is what keeps the two lists from drifting.
func TestFrozenInstructionFilesPinnedInContract(t *testing.T) {
	if len(frozenInstructionFiles) == 0 {
		t.Fatal("frozenInstructionFiles is empty; the pin would compare nothing")
	}
	if !slices.Equal(contract.FrozenInstructionFiles, frozenInstructionFiles) {
		t.Errorf("contract.FrozenInstructionFiles = %v, hook frozenInstructionFiles = %v",
			contract.FrozenInstructionFiles, frozenInstructionFiles)
	}
}

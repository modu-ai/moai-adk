// slot_lease_guard.go — the opt-in PreToolUse layer of the resource slot lease
// (SPEC-RESOURCE-SLOT-LEASE-001, card t607).
//
// M1 STATE: signature only. checkSlotLease allows every command and writes
// nothing, so the M1 tests compile and fail on behaviour. M4 implements the
// decision table (slot_lease_guard_test.go) and wires the call site in
// pre_tool.go directly after the integration-lock guard.
//
// Contract M4 implements (pinned by the tests, not by this comment):
//   - cfg.Enabled false: return at once — no record, no pattern, no advisory,
//     no audit line.
//   - deny only on positive evidence: a configured pattern matches outside
//     quoted spans AND the resource's holder is live, unexpired, and a session
//     other than the caller. The reason starts with slotLeaseViolationPrefix.
//   - the hook root is normalized with kanban.ResolveSlotLeaseRoot (the SAME
//     function the CLI uses) before any record is read.
//   - every uncertainty allows, writes an advisory to `advisory`, and appends
//     a fail-open audit line where a root is available.
package hook

import (
	"io"

	"github.com/modu-ai/moai-adk/internal/config"
)

// slotLeaseViolationPrefix is the deny sentinel the orchestrator matches.
const slotLeaseViolationPrefix = "SLOT_LEASE_VIOLATION:"

// checkSlotLease returns DecisionDeny plus a sentinel-prefixed reason when the
// command is attributed to a resource whose live, unexpired holder is another
// session; otherwise ("", "").
// @MX:TODO: [AUTO] M1 stub — implemented in M4 (AC-RSL-010/011/012/016).
func checkSlotLease(input *HookInput, hookRoot string, cfg config.SlotLeaseConfig, advisory io.Writer) (decision string, reason string) {
	return "", ""
}

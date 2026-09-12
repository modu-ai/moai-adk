// slot_lease_guard.go — the opt-in PreToolUse layer of the resource slot lease
// (SPEC-RESOURCE-SLOT-LEASE-001, card t607).
//
// A lane that holds no lease can still start a heavy command while another
// session holds the resource. This guard refuses that command, and only that
// command: a deny needs positive evidence on every term of
//
//	enabled ∧ a configured pattern matches outside quoted spans
//	        ∧ the holder is a different session ∧ live ∧ inside its bound.
//
// It inherits the integration-lock guard's three load-bearing properties:
//   - Opt-in, inert by default. With workflow.slot_lease.enabled false the
//     call site does not even resolve the project root, and this function
//     returns before reading any pattern or record.
//   - A deny sentinel (slotLeaseViolationPrefix) the orchestrator can match.
//   - FAIL OPEN. Every uncertainty — an unreadable record, an unknown caller,
//     a missing or unnormalizable root, a pattern that does not compile, a
//     resource entry that is not a list of strings — allows, says so on the
//     advisory stream, and leaves a fail-open audit line where a root exists.
//
// The patterns come only from configuration. No built-in command list exists:
// any list shipped here would be some programming language's commands.
//
// Root normalization (REQ-RSL-008). The hook root is CLAUDE_PROJECT_DIR or the
// cwd, with no git common-dir step, so inside a linked worktree it names the
// worktree — whose .moai/state holds no lease. Reading there would answer
// "nobody holds it" and quietly allow. The root is therefore normalized with
// kanban.ResolveSlotLeaseRoot, the same function the `moai slot` CLI uses.
package hook

import (
	"fmt"
	"io"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/kanban"
)

// slotLeaseViolationPrefix is the deny sentinel the orchestrator matches.
const slotLeaseViolationPrefix = "SLOT_LEASE_VIOLATION:"

// slotLeaseAdvisoryPrefix starts every advisory line this guard writes.
const slotLeaseAdvisoryPrefix = "[moai:slot-lease] advisory:"

// @MX:ANCHOR: [AUTO] checkSlotLease is the PreToolUse decision for the slot lease; its fail-open and root-normalization contract is what AC-RSL-010/011/012/016 pin
// @MX:REASON: [AUTO] a deny on uncertainty wedges every heavy command in the repository; reading the unnormalized hook root silently disables the guard in linked worktrees
// checkSlotLease returns DecisionDeny plus a sentinel-prefixed reason when the
// command is attributed to a resource whose live, unexpired holder is another
// session; otherwise ("", ""). Advisories go to advisory; audit lines go to the
// normalized root's .moai/logs/slot-lease-audit.jsonl.
func checkSlotLease(input *HookInput, hookRoot string, cfg config.SlotLeaseConfig, advisory io.Writer) (decision string, reason string) {
	if !cfg.Enabled || input == nil || len(input.ToolInput) == 0 {
		return "", ""
	}
	command := extractIntegrationCommand(input.ToolInput)
	if command == "" {
		return "", ""
	}
	scrubbed := substituteQuotedArguments(command)

	// Attribute the call. Resources are visited in sorted order so the
	// decision, and the resource a deny names, never depend on map order.
	names := make([]string, 0, len(cfg.Resources))
	for name := range cfg.Resources {
		names = append(names, name)
	}
	sort.Strings(names)
	var matched, uncertain []string
	for _, name := range names {
		entry := cfg.Resources[name]
		if entry.Invalid != "" {
			uncertain = append(uncertain, fmt.Sprintf("resource %q: %s", name, entry.Invalid))
			continue
		}
		if err := kanban.ValidateSlotResourceName(name); err != nil {
			uncertain = append(uncertain, err.Error())
			continue
		}
		for _, pattern := range entry.Commands {
			re, err := regexp.Compile(pattern)
			if err != nil {
				uncertain = append(uncertain, fmt.Sprintf("resource %q: pattern %q does not compile: %v", name, pattern, err))
				continue
			}
			if re.MatchString(scrubbed) {
				matched = append(matched, name)
				break
			}
		}
	}
	if len(matched) == 0 && len(uncertain) == 0 {
		return "", "" // not attributed to any resource: no record read, no audit line
	}

	advise := func(format string, args ...any) {
		if advisory != nil {
			_, _ = fmt.Fprintf(advisory, slotLeaseAdvisoryPrefix+" "+format+"; allowing\n", args...)
		}
	}
	if strings.TrimSpace(hookRoot) == "" {
		advise("no project root; cannot read any slot lease")
		return "", ""
	}
	root, err := kanban.ResolveSlotLeaseRoot(hookRoot)
	if err != nil {
		advise("cannot normalize %s to the shared root (%v)", hookRoot, err)
		auditSlotGuard(hookRoot, kanban.SlotLeaseAuditEntry{Event: "fail-open", Reason: "root-unresolved: " + err.Error(), SessionID: input.SessionID})
		return "", ""
	}
	for _, why := range uncertain {
		advise("%s", why)
		auditSlotGuard(root, kanban.SlotLeaseAuditEntry{Event: "fail-open", Reason: why, SessionID: input.SessionID})
	}
	if len(matched) == 0 {
		return "", ""
	}
	if input.SessionID == "" {
		advise("this session has no id; cannot tell it from a holder of %s", strings.Join(matched, ", "))
		auditSlotGuard(root, kanban.SlotLeaseAuditEntry{Event: "fail-open", Reason: "missing session id", Resource: matched[0]})
		return "", ""
	}

	now := time.Now()
	for _, name := range matched {
		lease, readErr := kanban.ReadSlotLease(root, name)
		entry := kanban.SlotLeaseAuditEntry{Resource: name, SessionID: input.SessionID}
		switch {
		case readErr != nil:
			advise("cannot read the lease for %s (%v)", name, readErr)
			entry.Event, entry.Reason = "fail-open", "unreadable record: "+readErr.Error()
		case !lease.Held():
			entry.Event = "allow-unheld"
		case lease.SessionID == input.SessionID:
			entry.Event = "allow-self"
		case lease.Stale():
			entry.Event = "allow-stale"
		case lease.Expired(now):
			entry.Event = "allow-expired"
		default:
			entry.Event = "guard-deny"
			entry.HolderSessionID, entry.HolderPID = lease.SessionID, lease.PID
			auditSlotGuard(root, entry)
			return DecisionDeny, fmt.Sprintf("%s resource %q is held by %s (session %s, pid %d) until %s. "+
				"Wait until `moai slot status --resource %s` shows it free, ask the holder to run `moai slot release --resource %s`, "+
				"or take it over deliberately with `moai slot acquire --resource %s --force`.",
				slotLeaseViolationPrefix, name, slotHolderLabel(lease), lease.SessionID, lease.PID, lease.ExpiresAt, name, name, name)
		}
		if lease != nil {
			entry.HolderSessionID, entry.HolderPID = lease.SessionID, lease.PID
		}
		auditSlotGuard(root, entry)
	}
	return "", ""
}

// slotHolderLabel prefers the human-facing session name over the id.
func slotHolderLabel(lease *kanban.SlotLease) string {
	if strings.TrimSpace(lease.SessionName) != "" {
		return lease.SessionName
	}
	return lease.SessionID
}

// auditSlotGuard appends a guard line; a logging failure never changes the
// decision (the guard is fail-open by contract).
func auditSlotGuard(root string, entry kanban.SlotLeaseAuditEntry) {
	_ = kanban.AppendSlotLeaseAudit(root, entry)
}

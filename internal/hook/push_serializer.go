package hook

// push_serializer.go — the push serializer (SPEC-AUTONOMY-PRECONDITION-001
// REQ-AP-001 / REQ-AP-002 / REQ-AP-007; design.md §B).
//
// Two pushes of `develop` from two lanes cost a cancelled CI run and a
// confused tree. When a signed contract carries the `push-develop` action with
// `push_requires_lease: true`, a Bash tool call that pushes `develop` must
// hold the existing `moai slot` lease on the resource `push-develop`. The
// serializer is that enforcement, beside the other PreToolUse deny guards:
//
//	ACTIVATE  the triple — mode contract ∧ actions contains push-develop
//	          ∧ push_requires_lease: true — read from the
//	          `moai contract show --json` document (spec.md §C.6: A1 is the
//	          field's producer; reading the JSON rather than contract.yaml
//	          keeps the derivation A1's). The triple does NOT depend on
//	          workflow.slot_lease.enabled, which keeps gating the generic
//	          slot guard only.
//	MATCH     program git, subcommand push, refspec targeting develop, with
//	          the same quote handling checkSlotLease applies.
//	ADMIT     free, expired, stale, or held by the calling session — and
//	          write the lease for the calling session with the configured
//	          bound (a takeover names the displaced holder; that is the slot
//	          lease's existing rule, not a new one).
//	DENY      held by a different live session within its bound —
//	          PUSH_SERIALIZATION_VIOLATION: naming the holder. Serialization
//	          is expected traffic, not a contract breach, so no escalation
//	          record is written on any path (REQ-AP-007).
//	RELEASE   PostToolUse on the admitted push with a non-zero exit — nothing
//	          is in flight, so the record goes immediately. A successful
//	          push keeps the lease: the holder releases it by hand after
//	          reading CI, and a forgotten release costs at most the bound.
//	FAIL OPEN Unreadable record, unresolvable root, or unknown caller →
//	          allow plus one audit line. An overlapping push costs one
//	          cancelled CI run; a stuck deny halts the lane.
//
// The record is the existing slot-lease record, read and written only through
// internal/kanban — no new record format, lock, or verb (spec.md §E C1).

import (
	"encoding/json"
	"fmt"
	"io"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/kanban"
)

// pushSerializerViolationPrefix is the deny sentinel the orchestrator matches.
const pushSerializerViolationPrefix = "PUSH_SERIALIZATION_VIOLATION:"

// pushSerializerAdvisoryPrefix starts every advisory line this guard writes.
const pushSerializerAdvisoryPrefix = "[moai:push-serializer] advisory:"

// PushDevelopSlotResource is the slot resource the serializer acquires. It is
// a valid resource name under kanban.ValidateSlotResourceName, and
// `moai slot status --resource push-develop` reads the same record a lane
// would read by hand (design.md §B Resource).
const PushDevelopSlotResource = "push-develop"

// PushShowJSON is the subset of the `moai contract show --json` document the
// activation triple is read from. The field names are the stable names of
// A1's Report JSON (internal/contract verify.go), so the document decodes
// directly; every other field of the report is ignored here.
type PushShowJSON struct {
	Mode              string   `json:"mode"`
	Actions           []string `json:"actions"`
	PushRequiresLease bool     `json:"push_requires_lease"`
}

// DecodePushShowJSON decodes the `moai contract show --json` document the
// activation triple is read from. The guard's single decode entry point: a
// schema change reaches this guard through A1's projection, never through a
// second parser over contract.yaml (design.md §B Activation).
func DecodePushShowJSON(data []byte) (*PushShowJSON, error) {
	var show PushShowJSON
	if err := json.Unmarshal(data, &show); err != nil {
		return nil, fmt.Errorf("push serializer: decode show --json document: %w", err)
	}
	return &show, nil
}

// Armed reports whether the activation triple holds: workflow.autonomy.mode
// is contract, the contract's actions contain push-develop, and
// push_requires_lease is true.
func (p *PushShowJSON) Armed() bool {
	return p != nil &&
		p.Mode == config.AutonomyModeContract &&
		slices.Contains(p.Actions, PushDevelopSlotResource) &&
		p.PushRequiresLease
}

// pushShowJSONLoader resolves the `moai contract show --json` document for the
// session's contract. Production wiring is the contract resolver of
// SPEC-AUTONOMY-ESCALATION-001 REQ-AE-002 (card t1235), which decides which
// contract governs a session; until that resolver lands, the nil default keeps
// the serializer inactive — the same inert-by-default posture as the opt-in
// guards beside it, and the reason no criterion here waits on it.
// Tests set the loader (or call checkPushSerializer directly with a fixture).
var pushShowJSONLoader func(hookRoot string) ([]byte, error)

// checkPushSerializer returns DecisionDeny plus a sentinel-prefixed reason
// when an armed serializer's command is a push of develop held back by a
// live, unexpired foreign lease; otherwise ("", ""). Advisories go to
// advisory; audit lines go to the normalized root's
// .moai/logs/slot-lease-audit.jsonl. Fails OPEN on every uncertainty.
func checkPushSerializer(input *HookInput, hookRoot string, show *PushShowJSON, maxDuration time.Duration, advisory io.Writer) (decision string, reason string) {
	if show == nil || !show.Armed() || input == nil || len(input.ToolInput) == 0 {
		return "", ""
	}
	command := extractIntegrationCommand(input.ToolInput)
	if command == "" {
		return "", ""
	}
	// Same quote handling checkSlotLease applies (design.md §B Matcher): the
	// goal here is to match the command being RUN, so text carried as data
	// inside a quoted argument must not read as a refspec.
	if !isDevelopPush(substituteQuotedArguments(command)) {
		return "", ""
	}
	advise := func(format string, args ...any) {
		if advisory != nil {
			_, _ = fmt.Fprintf(advisory, pushSerializerAdvisoryPrefix+" "+format+"; allowing\n", args...)
		}
	}

	if strings.TrimSpace(hookRoot) == "" {
		advise("no project root; cannot read the %s lease", PushDevelopSlotResource)
		return "", ""
	}
	root, err := kanban.ResolveSlotLeaseRoot(hookRoot)
	if err != nil {
		advise("cannot normalize %s to the shared root (%v)", hookRoot, err)
		auditPushSerializer(hookRoot, kanban.SlotLeaseAuditEntry{
			Event: "fail-open", Reason: "root-unresolved: " + err.Error(), SessionID: input.SessionID,
		})
		return "", ""
	}
	if input.SessionID == "" {
		advise("this session has no id; cannot hold or compare the %s lease", PushDevelopSlotResource)
		auditPushSerializer(root, kanban.SlotLeaseAuditEntry{
			Event: "fail-open", Reason: "missing session id", Resource: PushDevelopSlotResource,
		})
		return "", ""
	}
	lease, readErr := kanban.ReadSlotLease(root, PushDevelopSlotResource)
	if readErr != nil {
		advise("cannot read the lease for %s (%v)", PushDevelopSlotResource, readErr)
		auditPushSerializer(root, kanban.SlotLeaseAuditEntry{
			Event: "fail-open", Reason: "unreadable record: " + readErr.Error(), Resource: PushDevelopSlotResource, SessionID: input.SessionID,
		})
		return "", ""
	}

	if lease.Held() &&
		lease.SessionID != input.SessionID &&
		!lease.Stale() &&
		!lease.Expired(time.Now()) {
		auditPushSerializer(root, kanban.SlotLeaseAuditEntry{
			Event: "guard-deny", Resource: PushDevelopSlotResource, SessionID: input.SessionID,
			HolderSessionID: lease.SessionID, HolderPID: lease.PID,
		})
		return DecisionDeny, fmt.Sprintf("%s %s held by %s (session %s, pid %d) until %s. "+
			"Wait until `moai slot status --resource %s` shows it free, or ask the holder to run `moai slot release --resource %s`.",
			pushSerializerViolationPrefix, PushDevelopSlotResource, slotHolderLabel(lease), lease.SessionID, lease.PID, lease.ExpiresAt,
			PushDevelopSlotResource, PushDevelopSlotResource)
	}

	// Admit: acquire (or re-acquire / reclaim) for the calling session. A
	// takeover names the displaced holder — the slot lease's existing rule.
	// PID stays 0: the hook process cannot resolve the calling session's
	// owner pid, and a recorded 0 reads live (conservative) — the declared
	// bound is then the record's only expiry.
	if _, err := kanban.AcquireSlotLease(root, kanban.SlotLeaseRequest{
		Resource:    PushDevelopSlotResource,
		SessionID:   input.SessionID,
		Command:     command,
		MaxDuration: maxDuration,
	}); err != nil {
		advise("cannot record the %s lease (%v)", PushDevelopSlotResource, err)
		auditPushSerializer(root, kanban.SlotLeaseAuditEntry{
			Event: "fail-open", Reason: "acquire failed: " + err.Error(), Resource: PushDevelopSlotResource, SessionID: input.SessionID,
		})
		return "", ""
	}
	return "", ""
}

// releasePushLeaseOnFailure releases the calling session's push-develop lease
// when the admitted push's PostToolUse reports a non-zero exit — nothing is
// in flight, so nothing may keep the record. A successful push keeps the
// lease: the holder releases it with `moai slot release` after reading CI.
// Every uncertainty (no exit signal, no root, foreign holder) keeps the record
// and lets the declared bound expire it.
func releasePushLeaseOnFailure(input *HookInput, hookRoot string, show *PushShowJSON, advisory io.Writer) {
	if show == nil || !show.Armed() || input == nil || input.SessionID == "" {
		return
	}
	if !pushExitNonZero(input.ToolResponse) {
		return
	}
	command := extractIntegrationCommand(input.ToolInput)
	if command == "" || !isDevelopPush(substituteQuotedArguments(command)) {
		return
	}
	advise := func(format string, args ...any) {
		if advisory != nil {
			_, _ = fmt.Fprintf(advisory, pushSerializerAdvisoryPrefix+" "+format+"\n", args...)
		}
	}
	root, err := kanban.ResolveSlotLeaseRoot(hookRoot)
	if err != nil {
		advise("cannot normalize %s to the shared root (%v)", hookRoot, err)
		return
	}
	lease, err := kanban.ReadSlotLease(root, PushDevelopSlotResource)
	if err != nil || !lease.Held() || lease.SessionID != input.SessionID {
		return
	}
	if _, err := kanban.ReleaseSlotLease(root, PushDevelopSlotResource, input.SessionID, false); err != nil {
		advise("cannot release the %s lease (%v)", PushDevelopSlotResource, err)
	}
}

// pushExitNonZero reports whether a Bash tool_response carries a positive
// non-zero exit signal. Absent or unparseable signals read as zero: the
// release fires on observed failure, never on the absence of a success
// signal.
func pushExitNonZero(response json.RawMessage) bool {
	if len(response) == 0 {
		return false
	}
	var sig struct {
		Exit     *int `json:"exit"`
		ExitCode *int `json:"exit_code"`
	}
	if err := json.Unmarshal(response, &sig); err != nil {
		return false
	}
	for _, code := range []*int{sig.Exit, sig.ExitCode} {
		if code != nil && *code != 0 {
			return true
		}
	}
	return false
}

// pushSerializerShow resolves the activation document for the handler call
// sites. It returns nil (serializer inactive) when the loader seam is unset,
// the document does not decode, or the triple is not armed — in every one of
// those states no record is read and no audit line is written.
func pushSerializerShow(hookRoot string) *PushShowJSON {
	if pushShowJSONLoader == nil {
		return nil
	}
	data, err := pushShowJSONLoader(hookRoot)
	if err != nil || len(data) == 0 {
		return nil
	}
	show, err := DecodePushShowJSON(data)
	if err != nil || !show.Armed() {
		return nil
	}
	return show
}

// pushSerializerBound resolves the declared bound the way the `moai slot` CLI
// does: workflow.slot_lease.default_max_duration read from the project rooted
// at the normalized root (the loader falls back to the config default on every
// failure path). An unparseable configured value yields 0, which makes the
// admit path's acquire fail — and the serializer, failing open, allows.
func pushSerializerBound(root string) time.Duration {
	bound, err := kanban.ParseSlotLeaseMaxDuration(config.LoadSlotLeaseDefaultMaxDuration(root))
	if err != nil {
		return 0
	}
	return bound
}

// pushFlagWithValue is the closed set of git push flags whose value is a
// separate word, so the value is never mistaken for a refspec (`git push
// --repo develop main` names the remote, not the ref). Flags carrying their
// value inline (`--push-option=x`) are single tokens and already skipped.
var pushFlagWithValue = map[string]bool{
	"-r": true, "--repo": true,
	"-o": true, "--push-option": true,
	"--receive-pack": true, "--exec": true,
	"--recurse-submodules": true,
}

// isDevelopPush reports whether the scrubbed command line is a git push whose
// refspec targets develop: the program word is git (basename — a path to the
// binary matches), the subcommand is push, and a non-flag argument is a
// refspec whose destination ref resolves to develop. A bare `develop` pushes
// to develop; `HEAD:develop` and `refs/heads/develop` do too; `develop:main`
// does not (its target is main).
func isDevelopPush(scrubbed string) bool {
	words := strings.Fields(scrubbed)
	for i := 0; i+1 < len(words); i++ {
		if filepath.Base(words[i]) != "git" || words[i+1] != "push" {
			continue
		}
		skipValue := false
		for _, arg := range words[i+2:] {
			if skipValue {
				skipValue = false
				continue
			}
			if strings.HasPrefix(arg, "-") {
				if arg != "--" && pushFlagWithValue[arg] {
					skipValue = true
				}
				continue
			}
			if refspecTargetsDevelop(arg) {
				return true
			}
		}
	}
	return false
}

// refspecTargetsDevelop reports whether one push argument targets the develop
// ref. A colon separates src:dst; the destination is the target, and an empty
// destination (a deletion request with no explicit ref) does not name develop.
func refspecTargetsDevelop(arg string) bool {
	ref := strings.TrimLeft(arg, "+")
	if i := strings.Index(ref, ":"); i >= 0 {
		ref = ref[i+1:]
		if ref == "" {
			return false
		}
	}
	return strings.TrimPrefix(ref, "refs/heads/") == "develop"
}

// auditPushSerializer appends one guard line; a logging failure never changes
// the decision (the serializer fails open by contract).
func auditPushSerializer(root string, entry kanban.SlotLeaseAuditEntry) {
	_ = kanban.AppendSlotLeaseAudit(root, entry)
}

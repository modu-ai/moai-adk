package contract

import (
	"errors"
	"fmt"
	"reflect"
	"slices"
	"strings"

	"github.com/modu-ai/moai-adk/internal/mission"
)

// ErrNotProjectable is the base error every projection refusal wraps. The
// wrapping message always names the offending contract field or glob, per
// REQ-AP-008's fail-closed rule.
var ErrNotProjectable = errors.New("contract projection: contract is not projectable onto the mission validator")

// The projection is the one-way mapping of design.md §D (REQ-AP-008): a
// signed-valid contract becomes a mission.MissionContract, the verdict comes
// back from mission.ValidateMissionDecision, and mission's five refusal
// reasons are never re-implemented here. It is adopted from A1's drafted
// 12-row table with the three corrections spec.md §C.11 measured in this
// tree: (1) Approved and a positive MaxOperations are populated, (2) a
// trailing /** in ownership.write translates to its prefix while an
// inner-wildcard glob fails closed (mission's targetInsideScope is
// exact-or-prefix), and (3) the fields deliberately not projected are
// enumerated below, with card on them, so a bare unmapped-field rule cannot
// fail every contract on a required field.
//
// mission_contract_sha256 is deliberately not produced: the contract digest
// remains the sole tamper authority.

// DeliberatelyNotProjected enumerates the contract fields that have no
// mission counterpart and are intentionally dropped by the projection. card
// is on this list because the schema requires it while mission has no
// counterpart for it; a field absent from both this list and the projected
// set below fails the projection closed.
var DeliberatelyNotProjected = []string{
	"schema_version",
	"card",
	"acceptance.file",
	"invariants",
	"ownership",
	"ownership.never",
	"ownership.scratch",
	"reobserve",
	"review",
	"budget.turns",
	"plan_audit",
}

// projectedFields names the contract fields (and nested keys) the projection
// consumes. Keys not in this set or in DeliberatelyNotProjected are unmapped
// and fail closed.
var projectedFields = map[string]bool{
	"spec_id":              true,
	"acceptance":           true,
	"acceptance.sha256":    true,
	"acceptance.ac_count":  true,
	"ownership.write":      true,
	"approach":             true,
	"actions":              true,
	"budget":               true,
	"budget.operations":    true,
	"budget.audit_retries": true,
	"escalate_on":          true,
	"signature":            true,
}

// missionActionByContractAction maps the contract action vocabulary
// (rules.go) onto mission's Action set. worktree has no mission counterpart
// and is dropped; the correction requires at least one mapped action to
// survive, enforced in ProjectToMission.
var missionActionByContractAction = map[string]mission.Action{
	"commit":              mission.ActionCommit,
	"local-merge-develop": mission.ActionLocalMerge,
	ActionPushDevelop:     mission.ActionBatchPush,
}

// Fixed rows of the design.md §D mapping table.
const (
	missionMergeTarget        = "develop"
	missionPolicyVersion      = "contract-v1"
	missionRevocationBehavior = "stop-before-next-action"
)

var (
	missionProhibitedActions  = []mission.Action{mission.ActionMainMerge, mission.ActionForcePush, mission.ActionReleaseBranch, mission.ActionReleasePR}
	missionRecoveryConditions = []string{"sign --resign after acceptance change"}
)

// contractInventoryKeys enumerates the field keys schema_version 1 defines,
// derived from the live struct via reflection so a schema amendment that adds
// a field makes the projection fail closed loudly until the mapping tables
// above are updated. Nested keys are descended one level for the three
// sections whose children carry field-level semantics.
func contractInventoryKeys(c *Contract) []string {
	keys := inventoryKeysRecursive(reflect.TypeOf(*c), "")
	slices.Sort(keys)
	return keys
}

func inventoryKeysRecursive(t reflect.Type, prefix string) []string {
	var keys []string
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		tag := field.Tag.Get("yaml")
		if tag == "" {
			continue
		}
		name := strings.Split(tag, ",")[0]
		key := name
		if prefix != "" {
			key = prefix + "." + name
		}
		keys = append(keys, key)
		switch key {
		case "acceptance", "ownership", "budget":
			child := field.Type
			if child.Kind() == reflect.Pointer {
				child = child.Elem()
			}
			keys = append(keys, inventoryKeysRecursive(child, key)...)
		}
	}
	return keys
}

// unmappedProjectionKeys returns the sorted subset of keys that is neither
// projected nor on the deliberately-not-projected list.
func unmappedProjectionKeys(keys []string) []string {
	var unmapped []string
	for _, key := range keys {
		if !projectedFields[key] && !slices.Contains(DeliberatelyNotProjected, key) {
			unmapped = append(unmapped, key)
		}
	}
	return unmapped
}

// missionScopePrefix translates one ownership.write entry onto mission's
// exact-or-prefix containment. A trailing /** becomes its prefix; a glob with
// an inner wildcard has no prefix that preserves its meaning and is
// unmappable, which fails closed rather than widening or narrowing scope.
func missionScopePrefix(entry string) (string, error) {
	if !strings.Contains(entry, "*") {
		return entry, nil
	}
	if strings.HasSuffix(entry, "/**") {
		return strings.TrimSuffix(entry, "/**"), nil
	}
	return "", fmt.Errorf("inner-wildcard glob %q is unmappable (mission scope containment is exact-or-prefix)", entry)
}

// signedValid reports whether the contract's signature block verifies: the
// method/signer/receipt table holds and the recorded seal matches the
// recomputed one. This is the sole source of Approved (design.md §D row 13).
func signedValid(c *Contract) bool {
	if c.Signature == nil || !signatureConsistent(c.Signature) {
		return false
	}
	want, err := ComputeSeal(*c.Signature)
	return err == nil && want == c.Signature.Seal
}

// ProjectToMission maps a signed-valid contract onto mission's
// MissionContract so mission.ValidateMissionDecision can judge individual
// operations. It fails closed — returning an error that wraps
// ErrNotProjectable and names the offending field — when the contract
// carries an unmapped field, an inner-wildcard scope glob, an unsigned or
// tampered signature, no mission-mappable action, or a non-positive
// budget.operations. mission's exported surface is never widened.
func ProjectToMission(c *Contract) (mission.MissionContract, error) {
	if c == nil {
		return mission.MissionContract{}, fmt.Errorf("%w: contract is nil", ErrNotProjectable)
	}
	if unmapped := unmappedProjectionKeys(contractInventoryKeys(c)); len(unmapped) > 0 {
		return mission.MissionContract{}, fmt.Errorf("%w: field %q: no mission counterpart and not on the deliberately-not-projected list", ErrNotProjectable, strings.Join(unmapped, ", "))
	}
	if !signedValid(c) {
		return mission.MissionContract{}, fmt.Errorf("%w: field %q: signature missing, inconsistent, or carrying a seal that does not verify", ErrNotProjectable, "signature")
	}
	if c.Ownership == nil || len(c.Ownership.Write) == 0 {
		return mission.MissionContract{}, fmt.Errorf("%w: field %q: empty", ErrNotProjectable, "ownership.write")
	}
	scope := make([]string, 0, len(c.Ownership.Write))
	for _, entry := range c.Ownership.Write {
		prefix, err := missionScopePrefix(entry)
		if err != nil {
			return mission.MissionContract{}, fmt.Errorf("%w: field %q: %v", ErrNotProjectable, "ownership.write", err)
		}
		scope = append(scope, prefix)
	}
	if c.Acceptance == nil || c.Acceptance.SHA256 == nil {
		return mission.MissionContract{}, fmt.Errorf("%w: field %q: absent", ErrNotProjectable, "acceptance.sha256")
	}
	if c.Acceptance.ACCount == nil {
		return mission.MissionContract{}, fmt.Errorf("%w: field %q: absent", ErrNotProjectable, "acceptance.ac_count")
	}
	evidence := []string{
		"acceptance:" + *c.Acceptance.SHA256,
		fmt.Sprintf("ac_count:%d", *c.Acceptance.ACCount),
	}
	actions := make([]mission.Action, 0, len(c.Actions))
	for _, a := range c.Actions {
		if mapped, ok := missionActionByContractAction[a]; ok {
			actions = append(actions, mapped)
		}
	}
	if len(actions) == 0 {
		return mission.MissionContract{}, fmt.Errorf("%w: field %q: no contract action maps onto the mission action vocabulary", ErrNotProjectable, "actions")
	}
	if c.Budget == nil || c.Budget.Operations <= 0 {
		return mission.MissionContract{}, fmt.Errorf("%w: field %q: zero, negative, or absent; never mapped to zero", ErrNotProjectable, "budget.operations")
	}
	return mission.MissionContract{
		MissionID:          c.SpecID,
		Goal:               c.Approach,
		CompletionEvidence: evidence,
		Scope:              scope,
		AllowedActions:     actions,
		MergeTarget:        missionMergeTarget,
		ResourceLimits:     mission.ResourceLimits{MaxOperations: c.Budget.Operations, MaxRetries: c.Budget.AuditRetries},
		ProhibitedActions:  missionProhibitedActions,
		StopConditions:     c.EscalateOn,
		RecoveryConditions: missionRecoveryConditions,
		RevocationBehavior: missionRevocationBehavior,
		PolicyVersion:      missionPolicyVersion,
		Approved:           true,
	}, nil
}

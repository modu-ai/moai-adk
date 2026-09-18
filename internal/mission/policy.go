package mission

import (
	"errors"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type MissionState string

const (
	StateDraft     MissionState = "draft"
	StateApproved  MissionState = "approved"
	StateRunning   MissionState = "running"
	StateBlocked   MissionState = "blocked"
	StateRevoking  MissionState = "revoking"
	StateCompleted MissionState = "completed"
	StateFailed    MissionState = "failed"
)

type MissionSnapshot struct {
	MissionID        string
	ContractHash     string
	PolicyVersion    string
	SnapshotHash     string
	EvidenceRevision int64
	Evidence         map[string]string
	State            MissionState
	OperationsUsed   int
}

type OperationSpec struct {
	Action Action
	Target string
}

type Decision struct {
	DecisionID       string
	MissionID        string
	ContractHash     string
	PolicyVersion    string
	SnapshotHash     string
	EvidenceRevision int64
	Action           Action
	Targets          []string
	RequiredEvidence []string
	Evidence         map[string]string
	OperationSpecs   []OperationSpec
	AdvisorProse     string
	ExpiresAt        time.Time
}

func requiredEvidenceForAction(action Action) []string {
	switch action {
	case ActionPublish:
		return []string{"approval", "clarified", "fresh_snapshot", "organized"}
	case ActionPick:
		return []string{"approval", "fresh_snapshot", "published"}
	case ActionDispatch:
		return []string{"fresh_snapshot", "lane_available", "lane_owner_free", "lease", "picked"}
	case ActionCommit:
		return []string{"fresh_snapshot", "scope", "tests"}
	case ActionLocalMerge:
		return []string{"commit", "fresh_snapshot", "integration_lease"}
	case ActionBatchPush:
		return []string{"fresh_snapshot", "local_merges", "origin_readback"}
	case ActionReleaseBranch:
		return []string{"develop_ci", "fresh_snapshot", "origin_readback"}
	case ActionReleasePR:
		return []string{"fresh_snapshot", "release_audit", "release_ci", "release_review"}
	case ActionMainMerge:
		return []string{"fresh_snapshot", "main_protected", "release_gates"}
	default:
		return nil
	}
}

func RequiredEvidenceForAction(action Action) []string {
	return append([]string(nil), requiredEvidenceForAction(action)...)
}

func sameEvidenceKeys(got, want []string) bool {
	if len(got) != len(want) {
		return false
	}
	seen := make(map[string]bool, len(got))
	for _, key := range got {
		if seen[key] {
			return false
		}
		seen[key] = true
	}
	for _, key := range want {
		if !seen[key] {
			return false
		}
	}
	return true
}

func containsAction(values []Action, target Action) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}

func targetInsideScope(target string, scope []string) bool {
	for _, allowed := range scope {
		if target == allowed || strings.HasPrefix(target, allowed+"/") {
			return true
		}
	}
	return false
}

func validEvidenceSHA(value string) bool {
	if len(value) != 40 && len(value) != 64 {
		return false
	}
	for _, r := range value {
		if (r < '0' || r > '9') && (r < 'a' || r > 'f') {
			return false
		}
	}
	return true
}

func evidenceTrue(values map[string]string, keys ...string) bool {
	for _, key := range keys {
		if values[key] != "true" {
			return false
		}
	}
	return true
}

func validFreshEvidence(action Action, value string, revision int64) bool {
	switch action {
	case ActionPublish, ActionPick, ActionDispatch:
		return value == strconv.FormatInt(revision, 10)
	default:
		return validEvidenceSHA(value)
	}
}

func validScopeEvidence(value string) bool {
	if strings.TrimSpace(value) == "" {
		return false
	}
	for _, path := range strings.Split(value, ",") {
		path = strings.TrimSpace(path)
		clean := filepath.Clean(path)
		if path == "" || filepath.IsAbs(clean) || clean == "." || clean == ".." || strings.HasPrefix(filepath.ToSlash(clean), "../") || strings.HasPrefix(filepath.ToSlash(clean), ".git/") {
			return false
		}
	}
	return true
}

// validateEvidenceSemantics rejects syntactically present but false or stale
// claims. The caller must still source these values from authoritative state;
// this layer makes their expected type and value part of the sealed policy.
func validateEvidenceSemantics(action Action, values map[string]string, revision int64) bool {
	if !validFreshEvidence(action, values["fresh_snapshot"], revision) {
		return false
	}
	switch action {
	case ActionPublish:
		return values["approval"] == "explicit" && evidenceTrue(values, "clarified", "organized")
	case ActionPick:
		published := values["published"]
		return values["approval"] == "explicit" && len(published) > 1 && published[0] == 't' && strings.Trim(published[1:], "0123456789") == ""
	case ActionDispatch:
		return evidenceTrue(values, "lane_available", "lane_owner_free", "lease", "picked")
	case ActionCommit:
		return validScopeEvidence(values["scope"]) && validEvidenceSHA(values["tests"]) && values["tests"] == values["fresh_snapshot"]
	case ActionLocalMerge:
		return validEvidenceSHA(values["commit"]) && filepath.IsAbs(values["integration_lease"])
	case ActionBatchPush:
		return evidenceTrue(values, "local_merges", "origin_readback")
	case ActionReleaseBranch:
		return evidenceTrue(values, "develop_ci", "origin_readback")
	case ActionReleasePR:
		return evidenceTrue(values, "release_audit", "release_ci", "release_review")
	case ActionMainMerge:
		return evidenceTrue(values, "main_protected", "release_gates")
	default:
		return false
	}
}

// ValidateMissionDecision treats every Decision field as untrusted data. It
// never calls a tool or performs a state change; a caller may persist the
// prepared receipt only after this complete fail-closed validation succeeds.
func ValidateMissionDecision(sealed SealedContract, snapshot MissionSnapshot, decision Decision, now time.Time) (OperationReceipt, error) {
	if !sealed.Contract.Approved || (snapshot.State != StateApproved && snapshot.State != StateRunning) {
		return OperationReceipt{}, errors.New("mission decision: mission_not_running")
	}
	if snapshot.MissionID != sealed.Contract.MissionID || decision.MissionID != snapshot.MissionID {
		return OperationReceipt{}, errors.New("mission decision: mission_mismatch")
	}
	if sealed.Hash == "" || decision.ContractHash != sealed.Hash || snapshot.ContractHash != sealed.Hash || decision.PolicyVersion != sealed.Version || snapshot.PolicyVersion != sealed.Version {
		return OperationReceipt{}, errors.New("mission decision: policy_mismatch")
	}
	if decision.SnapshotHash == "" || decision.SnapshotHash != snapshot.SnapshotHash {
		return OperationReceipt{}, errors.New("mission decision: stale_snapshot")
	}
	if !decision.ExpiresAt.After(now) {
		return OperationReceipt{}, errors.New("mission decision: expired_decision")
	}
	if containsAction(sealed.Contract.ProhibitedActions, decision.Action) {
		return OperationReceipt{}, errors.New("mission decision: prohibited_action")
	}
	if !containsAction(sealed.Contract.AllowedActions, decision.Action) {
		return OperationReceipt{}, errors.New("mission decision: action_not_allowed")
	}
	if len(decision.Targets) == 0 {
		return OperationReceipt{}, errors.New("mission decision: target_missing")
	}
	for _, target := range decision.Targets {
		if !targetInsideScope(target, sealed.Contract.Scope) {
			return OperationReceipt{}, errors.New("mission decision: scope_expansion")
		}
	}
	requiredEvidence := requiredEvidenceForAction(decision.Action)
	if len(requiredEvidence) == 0 || !sameEvidenceKeys(decision.RequiredEvidence, requiredEvidence) {
		return OperationReceipt{}, errors.New("mission decision: evidence_policy_mismatch")
	}
	if snapshot.EvidenceRevision <= 0 || decision.EvidenceRevision != snapshot.EvidenceRevision {
		return OperationReceipt{}, errors.New("mission decision: stale_evidence_revision")
	}
	for _, required := range requiredEvidence {
		if strings.TrimSpace(snapshot.Evidence[required]) == "" || decision.Evidence[required] != snapshot.Evidence[required] {
			return OperationReceipt{}, errors.New("mission decision: evidence_missing")
		}
	}
	if !validateEvidenceSemantics(decision.Action, decision.Evidence, snapshot.EvidenceRevision) {
		return OperationReceipt{}, errors.New("mission decision: evidence_value_invalid")
	}
	if snapshot.OperationsUsed >= sealed.Contract.ResourceLimits.MaxOperations {
		return OperationReceipt{}, errors.New("mission decision: resource_limit")
	}
	return OperationReceipt{OperationID: stableOperationID(sealed.Hash, decision), MissionID: decision.MissionID, DecisionID: decision.DecisionID, Action: decision.Action, Targets: append([]string(nil), decision.Targets...), SnapshotHash: decision.SnapshotHash, State: ReceiptPrepared}, nil
}

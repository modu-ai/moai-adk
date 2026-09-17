package mission

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"sort"
	"strings"
)

type Action string

const (
	ActionPublish       Action = "publish"
	ActionPick          Action = "pick"
	ActionDispatch      Action = "dispatch"
	ActionCommit        Action = "commit"
	ActionLocalMerge    Action = "local_develop_merge"
	ActionBatchPush     Action = "batch_push"
	ActionReleaseBranch Action = "release_branch"
	ActionReleasePR     Action = "release_pr"
	ActionMainMerge     Action = "main_merge"
	ActionForcePush     Action = "force_push"
)

type ResourceLimits struct {
	MaxOperations int `json:"max_operations"`
	MaxRetries    int `json:"max_retries"`
}

type MissionContract struct {
	MissionID          string         `json:"mission_id"`
	Goal               string         `json:"goal"`
	CompletionEvidence []string       `json:"completion_evidence"`
	Scope              []string       `json:"scope"`
	AllowedActions     []Action       `json:"allowed_actions"`
	MergeTarget        string         `json:"merge_target"`
	ResourceLimits     ResourceLimits `json:"resource_limits"`
	ProhibitedActions  []Action       `json:"prohibited_actions"`
	StopConditions     []string       `json:"stop_conditions"`
	RecoveryConditions []string       `json:"recovery_conditions"`
	RevocationBehavior string         `json:"revocation_behavior"`
	PolicyVersion      string         `json:"policy_version"`
	Approved           bool           `json:"approved"`
}

type SealedContract struct {
	Contract MissionContract
	Version  string
	Hash     string
	Bytes    []byte
}

func contractComplete(c MissionContract) bool {
	return strings.TrimSpace(c.MissionID) != "" && strings.TrimSpace(c.Goal) != "" &&
		len(c.CompletionEvidence) > 0 && len(c.Scope) > 0 && len(c.AllowedActions) > 0 &&
		strings.TrimSpace(c.MergeTarget) != "" && c.ResourceLimits.MaxOperations > 0 && c.ResourceLimits.MaxRetries >= 0 &&
		len(c.ProhibitedActions) > 0 && len(c.StopConditions) > 0 && len(c.RecoveryConditions) > 0 &&
		strings.TrimSpace(c.RevocationBehavior) != "" && strings.TrimSpace(c.PolicyVersion) != "" && c.Approved
}

func canonicalContract(c MissionContract) MissionContract {
	clone := c
	clone.CompletionEvidence = append([]string(nil), c.CompletionEvidence...)
	clone.Scope = append([]string(nil), c.Scope...)
	clone.AllowedActions = append([]Action(nil), c.AllowedActions...)
	clone.ProhibitedActions = append([]Action(nil), c.ProhibitedActions...)
	clone.StopConditions = append([]string(nil), c.StopConditions...)
	clone.RecoveryConditions = append([]string(nil), c.RecoveryConditions...)
	sort.Strings(clone.CompletionEvidence)
	sort.Strings(clone.Scope)
	sort.Slice(clone.AllowedActions, func(i, j int) bool { return clone.AllowedActions[i] < clone.AllowedActions[j] })
	sort.Slice(clone.ProhibitedActions, func(i, j int) bool { return clone.ProhibitedActions[i] < clone.ProhibitedActions[j] })
	sort.Strings(clone.StopConditions)
	sort.Strings(clone.RecoveryConditions)
	return clone
}

func SealMissionContract(contract MissionContract) (SealedContract, error) {
	if !contractComplete(contract) {
		return SealedContract{}, errors.New("mission contract: incomplete_contract")
	}
	canonical := canonicalContract(contract)
	data, err := json.Marshal(canonical)
	if err != nil {
		return SealedContract{}, err
	}
	sum := sha256.Sum256(data)
	return SealedContract{Contract: canonical, Version: canonical.PolicyVersion, Hash: hex.EncodeToString(sum[:]), Bytes: data}, nil
}

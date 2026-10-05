// goal_blocked_question_regression_test.go — SPEC-JEV-GOAL-DIST-001 M7a
// (AC-JEVG-001 / AC-JEVG-002). Seat (i) lane-question routing is WITHDRAWN
// from the deliverable set (spec.md §C.1): its producer milestone is recorded
// blocked and its host-design question is open and unowned. What this SPEC
// still owes is the unchanged-behaviour guarantee, so these tests pin the
// blocked-question outcome the loop already produces: a persisted blocked
// result of the pre-SPEC shape and a stop with no effects, no routing consumed,
// and no question issued inside the sealed-scope loop.
package cli

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/mission"
)

// blockedQuestionFixtureID is a syntactically valid mission session id, in the
// shape the existing lifecycle tests use.
const blockedQuestionFixtureID = "018f4f4a-7b7c-7a11-8f4d-666666666666"

// blockedQuestionFixtureMission returns an approved-shaped mission whose
// contract seals cleanly, so the persisted blocked result carries the full
// production lineage (contract + hash) rather than a bare skeleton.
func blockedQuestionFixtureMission(t *testing.T) mission.AutoMission {
	t.Helper()
	contract := mission.MissionContract{
		MissionID:          blockedQuestionFixtureID,
		Goal:               "finish the approved item",
		CompletionEvidence: []string{"CI"},
		Scope:              []string{"gtd:card"},
		AllowedActions:     []mission.Action{mission.ActionPublish},
		MergeTarget:        "develop",
		ResourceLimits:     mission.ResourceLimits{MaxOperations: 3, MaxRetries: 1},
		ProhibitedActions:  []mission.Action{mission.ActionForcePush},
		StopConditions:     []string{"revoked"},
		RecoveryConditions: []string{"authoritative_readback"},
		RevocationBehavior: "stop_new_and_reconcile",
		PolicyVersion:      "gtd-auto-v1",
		Approved:           true,
	}
	sealed, err := mission.SealMissionContract(contract)
	if err != nil {
		t.Fatal(err)
	}
	return mission.AutoMission{
		SessionID:    blockedQuestionFixtureID,
		Text:         "finish the approved item",
		MissionMode:  mission.ModeAuto,
		State:        mission.StateRunning,
		Contract:     &contract,
		ContractHash: sealed.Hash,
	}
}

// TestMissionBlockedQuestionPersistedShapeUnchanged pins the persisted shape of
// the loop's blocked-question outcome (AC-JEVG-001): state=blocked, the blocker
// reason, the contract lineage intact, file mode 0600 — and NO key through
// which a routing decision, a classification answer, or a user question could
// enter the persisted record. A future change that adds any such field changes
// the shape this SPEC froze and must re-open the seat-(i) disposition first.
func TestMissionBlockedQuestionPersistedShapeUnchanged(t *testing.T) {
	root := t.TempDir()
	state := blockedQuestionFixtureMission(t)

	if err := persistMissionBlock(root, &state, "blocked_question"); err != nil {
		t.Fatalf("persistMissionBlock: %v", err)
	}

	path := filepath.Join(root, ".moai", "state", "mission", blockedQuestionFixtureID+".json")
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read persisted mission: %v", err)
	}
	if info, err := os.Stat(path); err != nil || info.Mode().Perm() != 0o600 {
		t.Fatalf("persisted mission mode = %v err=%v, want 0600", info.Mode(), err)
	}

	var doc map[string]any
	if err := json.Unmarshal(raw, &doc); err != nil {
		t.Fatalf("parse persisted mission: %v", err)
	}
	if got := doc["state"]; got != string(mission.StateBlocked) {
		t.Errorf("persisted state = %v, want %q", got, mission.StateBlocked)
	}
	if got := doc["last_blocker"]; got != "blocked_question" {
		t.Errorf("persisted last_blocker = %v, want blocked_question", got)
	}
	if got := doc["mission_mode"]; got != string(mission.ModeAuto) {
		t.Errorf("persisted mission_mode = %v, want %q", got, mission.ModeAuto)
	}

	// The shape carries no routing, classification, answer, or question field.
	// Scan keys at every level: the pinned record routes nothing and answers
	// nothing, so no key of that family may exist anywhere in the JSON.
	for _, key := range flattenedKeys(t, raw) {
		lower := strings.ToLower(key)
		for _, forbidden := range []string{"question", "rout", "classif", "answer"} {
			if strings.Contains(lower, forbidden) {
				t.Errorf("persisted blocked result carries key %q — the pre-SPEC shape routes nothing and answers nothing", key)
			}
		}
	}

	// The blocked record still loads and passes integrity: a block persists a
	// resumable state, not a discard.
	loaded, err := mission.LoadAutoMission(root, blockedQuestionFixtureID)
	if err != nil || loaded == nil {
		t.Fatalf("LoadAutoMission after block: %+v err=%v", loaded, err)
	}
	if err := mission.ValidateAutoMissionIntegrity(loaded); err != nil {
		t.Fatalf("blocked record failed integrity: %v", err)
	}
}

// flattenedKeys walks the JSON document and returns every object key at every
// nesting level, so the forbidden-shape scan cannot be evaded by nesting a new
// field inside the contract or a future sub-object.
func flattenedKeys(t *testing.T, raw []byte) []string {
	t.Helper()
	var doc any
	if err := json.Unmarshal(raw, &doc); err != nil {
		t.Fatalf("parse: %v", err)
	}
	var keys []string
	var walk func(v any)
	walk = func(v any) {
		switch node := v.(type) {
		case map[string]any:
			for k, child := range node {
				keys = append(keys, k)
				walk(child)
			}
		case []any:
			for _, child := range node {
				walk(child)
			}
		}
	}
	walk(doc)
	return keys
}

// TestGoalLoopCarriesNoUserQuestionSurface is the AC-JEVG-002 guard: the
// sealed-scope loop's Go surface issues no user question between its single
// approval and completion. The scan is a means to an end, so it carries its
// own positive control — internal/mission/supervisor.go holds the identifier
// legitimately (the persisted-state flag that stays false), proving the scan
// can see the string it is being trusted to exclude.
func TestGoalLoopCarriesNoUserQuestionSurface(t *testing.T) {
	loopFiles := []string{"goal.go", "goal_runnable.go", "hook_stop_goal.go"}
	for _, name := range loopFiles {
		body, err := os.ReadFile(name)
		if err != nil {
			t.Fatalf("read %s: %v", name, err)
		}
		for i, line := range strings.Split(string(body), "\n") {
			trimmed := strings.TrimSpace(line)
			if strings.HasPrefix(trimmed, "//") {
				continue // comment mentions are not an emission surface
			}
			if strings.Contains(line, "AskUserQuestion") {
				t.Errorf("%s:%d references AskUserQuestion — the sealed-scope loop issues no user question (REQ-JEVG-002)", name, i+1)
			}
		}
	}

	control, err := os.ReadFile(filepath.Join("..", "mission", "supervisor.go"))
	if err != nil {
		t.Fatalf("read positive control: %v", err)
	}
	if !strings.Contains(string(control), "AskUserQuestion") {
		t.Fatal("positive control failed: supervisor.go no longer carries the identifier, so the zero-result above establishes nothing")
	}
}

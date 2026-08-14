package epic

import (
	"encoding/json"
	"strings"
	"testing"
)

// TestRenderJSON_FrozenShape verifies AC-ES-008: the JSON output matches the
// frozen shape field-for-field with stable key ordering, and stderr is empty
// on success (RenderJSON only writes to the returned []byte).
func TestRenderJSON_FrozenShape(t *testing.T) {
	status := &EpicStatus{
		Epic:      "NAVIGATOR-SYNC",
		EpicToken: "BAS",
		Milestones: []MilestoneEntry{
			{ID: "M0", Label: "graph layer", Status: "done", Covered: true, SpecID: "SPEC-NAVIGATOR-SYNC-001", SpecStatus: "completed", SyncCommitSHA: "abc123"},
			{ID: "M2", Label: "route", Status: "absent", Covered: false, SpecID: "", SpecStatus: "", SyncCommitSHA: ""},
		},
		Done:                1,
		Total:               2,
		Pct:                 50,
		OrphanMx:            []string{"M2"},
		DesignReport:        ".moai/reports/foo.html",
		BaselineAttribution: "deadbeef",
	}

	data, err := RenderJSON(status)
	if err != nil {
		t.Fatalf("RenderJSON: %v", err)
	}

	// Parse back to verify field shape.
	var parsed map[string]any
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("invalid JSON: %v\n%s", err, data)
	}
	requiredKeys := []string{
		"epic", "epic_token", "milestones", "done", "total", "pct",
		"extra_mx", "untracked_specs", "design_report", "baseline_attribution",
	}
	for _, k := range requiredKeys {
		if _, has := parsed[k]; !has {
			t.Errorf("JSON missing required key %q", k)
		}
	}
	// orphan_mx is omitempty; with len 1 it must be present.
	if _, has := parsed["orphan_mx"]; !has {
		t.Errorf("JSON missing orphan_mx (present when non-empty)")
	}
	// Stable ordering: epic must come before epic_token, which must come before
	// milestones, which must come before done.
	jsonStr := string(data)
	if !strings.Contains(jsonStr, `"epic":`) {
		t.Errorf("JSON missing epic field")
	}
	if idxEpic := strings.Index(jsonStr, `"epic":`); idxEpic < 0 {
		t.Errorf("epic field not found")
	} else {
		idxMilestones := strings.Index(jsonStr, `"milestones":`)
		idxDone := strings.Index(jsonStr, `"done":`)
		if idxMilestones < idxEpic || idxDone < idxMilestones {
			t.Errorf("stable order broken: epic < milestones < done expected")
		}
	}
	// Baseline attribution present.
	if parsed["baseline_attribution"] != "deadbeef" {
		t.Errorf("baseline_attribution = %v, want deadbeef", parsed["baseline_attribution"])
	}
}

// TestRenderJSON_OmitOrphanWhenEmpty verifies AC-ES-005b: when orphan_mx is
// nil, the JSON omits the field entirely.
func TestRenderJSON_OmitOrphanWhenEmpty(t *testing.T) {
	status := &EpicStatus{
		Epic:                "X",
		EpicToken:           "T",
		Milestones:          []MilestoneEntry{},
		BaselineAttribution: "",
	}
	data, err := RenderJSON(status)
	if err != nil {
		t.Fatalf("RenderJSON: %v", err)
	}
	if strings.Contains(string(data), "orphan_mx") {
		t.Errorf("orphan_mx should be omitted when nil:\n%s", data)
	}
	if strings.Contains(string(data), "design_report") {
		t.Errorf("design_report should be omitted when empty:\n%s", data)
	}
}

// TestRenderJSON_ForwardCompat verifies AC-ES-009: a JSON document with extra
// future fields parses cleanly into a struct that uses the current shape.
func TestRenderJSON_ForwardCompat(t *testing.T) {
	// Add a hypothetical future field `in_flight_worktrees` to the JSON.
	extra := `{
		"epic": "X",
		"epic_token": "T",
		"milestones": [],
		"done": 0,
		"total": 0,
		"pct": 0,
		"in_flight_worktrees": ["wt-1", "wt-2"],
		"baseline_attribution": "abc"
	}`
	var status EpicStatus
	if err := json.Unmarshal([]byte(extra), &status); err != nil {
		t.Fatalf("forward-compat parse failed: %v", err)
	}
	if status.Epic != "X" {
		t.Errorf("Epic = %q, want X", status.Epic)
	}
}

// TestRenderHuman_ProgressBoardGrammar verifies AC-ES-008b: the human output's
// first non-blank line matches the Progress Board bar grammar, with status
// icons from {🟢, 🟡, ⬜}.
func TestRenderHuman_ProgressBoardGrammar(t *testing.T) {
	status := &EpicStatus{
		Epic:      "NAVIGATOR-SYNC",
		EpicToken: "BAS",
		Milestones: []MilestoneEntry{
			{ID: "M0", Label: "graph", Status: "done", Covered: true, SpecID: "SPEC-A", SpecStatus: "completed", SyncCommitSHA: "x"},
			{ID: "M1", Label: "detect", Status: "in-progress", Covered: true, SpecID: "SPEC-B", SpecStatus: "in-progress"},
			{ID: "M2", Label: "route", Status: "absent", Covered: false},
			{ID: "M3", Label: "fix", Status: "planned", Covered: true, SpecID: "SPEC-C", SpecStatus: "draft"},
		},
		Done:                1,
		Total:               4,
		Pct:                 25,
		BaselineAttribution: "deadbeef",
	}
	out, err := RenderHuman(status, "en")
	if err != nil {
		t.Fatalf("RenderHuman: %v", err)
	}
	// Find the first non-blank line.
	lines := strings.Split(out, "\n")
	var first string
	for _, l := range lines {
		if strings.TrimSpace(l) != "" {
			first = l
			break
		}
	}
	// Should match /🎯 .* ▓+░+ +1\/4 \(25%\)/
	if !strings.HasPrefix(first, "🎯") {
		t.Errorf("first line does not start with 🎯: %q", first)
	}
	if !strings.Contains(first, "▓") || !strings.Contains(first, "░") {
		t.Errorf("first line lacks bar chars ▓/░: %q", first)
	}
	if !strings.Contains(first, "1/4") {
		t.Errorf("first line lacks done/total 1/4: %q", first)
	}
	if !strings.Contains(first, "25%") {
		t.Errorf("first line lacks pct 25%%: %q", first)
	}
	// Status icons present somewhere in the output.
	if !strings.Contains(out, "🟢") {
		t.Errorf("output lacks done icon 🟢")
	}
}

// TestRenderHuman_KoreanLocale verifies the Korean locale renders 에픽 진행 label.
func TestRenderHuman_KoreanLocale(t *testing.T) {
	status := &EpicStatus{
		Epic:                "X",
		EpicToken:           "T",
		Milestones:          []MilestoneEntry{},
		Done:                0,
		Total:               0,
		Pct:                 0,
		BaselineAttribution: "",
	}
	out, err := RenderHuman(status, "ko")
	if err != nil {
		t.Fatalf("RenderHuman: %v", err)
	}
	if !strings.Contains(out, "에픽 진행") {
		t.Errorf("Korean locale lacks '에픽 진행':\n%s", out)
	}
}

// TestRenderHuman_EmptyEpic verifies AC-ES-003b human shape: "no SPECs matched".
func TestRenderHuman_EmptyEpic(t *testing.T) {
	status := &EpicStatus{
		Epic:      "NONEXIST",
		EpicToken: "",
		Done:      0,
		Total:     0,
		Pct:       0,
	}
	out, err := RenderHuman(status, "en")
	if err != nil {
		t.Fatalf("RenderHuman: %v", err)
	}
	if !strings.Contains(out, "NONEXIST") {
		t.Errorf("empty epic output lacks prefix NONEXIST:\n%s", out)
	}
	if !strings.Contains(out, "no SPECs matched") {
		t.Errorf("empty epic output lacks 'no SPECs matched':\n%s", out)
	}
}

// TestRenderHuman_MatchedButNoMarkers verifies that Total == 0 with a non-empty
// UntrackedSpecs set is NOT reported as "no SPECs matched" — the prefix did
// match; only the `(TOKEN Mx)` title markers are absent.
func TestRenderHuman_MatchedButNoMarkers(t *testing.T) {
	status := &EpicStatus{
		Epic:      "KANBAN",
		EpicToken: "",
		Done:      0,
		Total:     0,
		Pct:       0,
		UntrackedSpecs: []string{
			"SPEC-KANBAN-BOARD-001",
			"SPEC-KANBAN-RENAME-001",
		},
	}
	out, err := RenderHuman(status, "en")
	if err != nil {
		t.Fatalf("RenderHuman: %v", err)
	}
	if strings.Contains(out, "no SPECs matched") {
		t.Errorf("marker-less epic misreported as an unmatched prefix:\n%s", out)
	}
	if !strings.Contains(out, "2 SPEC(s) matched") {
		t.Errorf("output lacks the matched count:\n%s", out)
	}
	if !strings.Contains(out, "SPEC-KANBAN-BOARD-001") {
		t.Errorf("output lacks the untracked SPEC list:\n%s", out)
	}
}

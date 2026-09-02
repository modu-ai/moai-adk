package hook

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
)

// ---------------------------------------------------------------------------
// M1 — schema version field (AC-IBX-007) + cap constants (plan.md §F M1).
// ---------------------------------------------------------------------------

// TestLessonsInboxStub_GoldenJSONCarriesSchemaVersion covers AC-IBX-007 part 1:
// a stub the collector appends parses as JSON carrying an integer version
// field equal to 1 (REQ-IBX-008).
func TestLessonsInboxStub_GoldenJSONCarriesSchemaVersion(t *testing.T) {
	root := t.TempDir()

	appendLessonsInboxStub(root, "tool_failure:Bash:ExitError", "exit status 1", "tool:Bash")

	data, err := os.ReadFile(filepath.Join(root, ".moai", "lessons-inbox.jsonl"))
	if err != nil {
		t.Fatalf("read lessons-inbox.jsonl: %v", err)
	}
	line := strings.TrimSpace(string(data))
	if line == "" {
		t.Fatal("no stub line written")
	}
	var raw map[string]any
	if err := json.Unmarshal([]byte(line), &raw); err != nil {
		t.Fatalf("stub line is not JSON: %v (%s)", err, line)
	}
	got, present := raw["v"]
	if !present {
		t.Fatalf("stub line carries no version field: %s", line)
	}
	num, isNum := got.(float64)
	if !isNum {
		t.Fatalf("version field is not a JSON number: %v (%T)", got, got)
	}
	if num != 1 {
		t.Fatalf("version = %v, want 1", num)
	}
	// Golden form: encoding/json renders the int as the bare literal `"v":1`.
	if !strings.Contains(line, `"v":1`) {
		t.Fatalf("stub line does not carry the integer literal \"v\":1: %s", line)
	}
}

// TestInboxStubVersion_AbsenceReadsAsV1 covers AC-IBX-007 part 2: a pre-upgrade
// line without the version field resolves as version 1 when parsed by the
// SPEC's reader (REQ-IBX-008 absence tolerance).
func TestInboxStubVersion_AbsenceReadsAsV1(t *testing.T) {
	// Pre-upgrade line: the REQ-HRR-006 schema carried no version field.
	var pre map[string]any
	preLine := `{"timestamp":"2026-01-01T00:00:00Z","event_key":"tool_failure:Bash:ExitError","summary":"exit status 1","source":"tool:Bash"}`
	if err := json.Unmarshal([]byte(preLine), &pre); err != nil {
		t.Fatalf("unmarshal pre-upgrade line: %v", err)
	}
	if got := InboxStubVersion(pre); got != 1 {
		t.Errorf("InboxStubVersion(pre-upgrade line) = %d, want 1", got)
	}

	// An explicit version passes through unchanged.
	var explicit map[string]any
	if err := json.Unmarshal([]byte(`{"v":2,"event_key":"k"}`), &explicit); err != nil {
		t.Fatalf("unmarshal explicit line: %v", err)
	}
	if got := InboxStubVersion(explicit); got != 2 {
		t.Errorf("InboxStubVersion(explicit v:2) = %d, want 2", got)
	}

	// An empty object (absence, not just a missing field) also reads as 1.
	if got := InboxStubVersion(map[string]any{}); got != 1 {
		t.Errorf("InboxStubVersion(empty) = %d, want 1", got)
	}
}

// TestInboxCapConstants_PinnedDefaults pins the M1-finalized cap constants
// (plan.md §F M1: "cap default 1 MiB finalized in M1"). The literals here are
// the contract, not duplicates: the test fails if the single-source default is
// retuned without a deliberate decision.
func TestInboxCapConstants_PinnedDefaults(t *testing.T) {
	if config.DefaultInboxMaxBytes != 1<<20 {
		t.Errorf("DefaultInboxMaxBytes = %d, want %d (1 MiB)", config.DefaultInboxMaxBytes, 1<<20)
	}
	if config.DefaultInboxArchiveGenerations != 2 {
		t.Errorf("DefaultInboxArchiveGenerations = %d, want 2", config.DefaultInboxArchiveGenerations)
	}
}

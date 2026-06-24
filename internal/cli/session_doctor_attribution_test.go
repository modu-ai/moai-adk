package cli

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/session"
)

// TestSessionDoctorRegistryAbsent is the M1 RED-GREEN test for REQ-WPR-001
// / AC-WPR-001 (SPEC-V3R6-SESSION-ID-ATTRIBUTION-REPAIR-001).
//
// `moai session doctor` reports (a) whether the registry file exists,
// (b) whether the current process's session_id is present, (c) the likely
// root cause when the registry is empty.
func TestSessionDoctorRegistryAbsent(t *testing.T) {
	withTempRegistry(t)

	// Registry file does NOT exist yet (no register call made).
	out, err := runSession(t, "doctor", "--json")
	if err != nil {
		t.Fatalf("doctor err: %v (out=%s)", err, out)
	}

	var payload map[string]any
	if err := json.Unmarshal([]byte(strings.TrimSpace(out)), &payload); err != nil {
		t.Fatalf("doctor --json not valid JSON: %v (out=%s)", err, out)
	}
	if payload["action"] != "doctor" {
		t.Errorf("doctor action: got %v, want doctor", payload["action"])
	}
	// Registry absent → registry_exists must be false.
	if payload["registry_exists"] != false {
		t.Errorf("registry_exists: got %v, want false (registry not created yet)", payload["registry_exists"])
	}
	// Root-cause candidates MUST be enumerated.
	candidates, ok := payload["root_cause_candidates"].([]any)
	if !ok || len(candidates) == 0 {
		t.Errorf("root_cause_candidates missing or empty: %v", payload["root_cause_candidates"])
	}
	candidatesStr := strings.ToLower(strings.Join(strings.Fields(fmtSlice(candidates)), " "))
	for _, want := range []string{"session_id", "hook", "wrapper"} {
		if !strings.Contains(candidatesStr, want) {
			t.Errorf("root_cause_candidates should mention %q. Got: %s", want, candidatesStr)
		}
	}
}

// TestSessionDoctorRegistryPresentWithEntry verifies doctor reports
// registry_exists=true and entry_count when the registry has entries.
func TestSessionDoctorRegistryPresentWithEntry(t *testing.T) {
	withTempRegistry(t)

	if _, err := runSession(t, "register", "uuid-doc-1", "SPEC-DOC", "plan"); err != nil {
		t.Fatal(err)
	}

	out, err := runSession(t, "doctor", "--json")
	if err != nil {
		t.Fatalf("doctor err: %v", err)
	}
	var payload map[string]any
	if err := json.Unmarshal([]byte(strings.TrimSpace(out)), &payload); err != nil {
		t.Fatalf("doctor --json: %v (out=%s)", err, out)
	}
	if payload["registry_exists"] != true {
		t.Errorf("registry_exists: got %v, want true", payload["registry_exists"])
	}
	// entry_count is a json float64.
	if f, ok := payload["entry_count"].(float64); !ok || int(f) != 1 {
		t.Errorf("entry_count: got %v, want 1", payload["entry_count"])
	}
}

// fmtSlice is a tiny helper to stringify a []any for substring checks.
func fmtSlice(s []any) string {
	parts := make([]string, len(s))
	for i, v := range s {
		parts[i] = strings.TrimSpace(strings.Trim(strings.Trim(fmt.Sprintf("%v", v), "[]"), `"`))
	}
	return strings.Join(parts, " ")
}

// TestSessionDoctorHumanReadable verifies the non-JSON path emits
// readable text mentioning registry state.
func TestSessionDoctorHumanReadable(t *testing.T) {
	withTempRegistry(t)

	out, err := runSession(t, "doctor")
	if err != nil {
		t.Fatalf("doctor err: %v", err)
	}
	if !strings.Contains(out, "registry") {
		t.Errorf("doctor human-readable should mention registry. Got: %s", out)
	}
}

// Compile-time guard: ensure the test file references the registry path so
// the import is not dropped if the helper above is refactored.
var _ = session.DefaultRegistryPath

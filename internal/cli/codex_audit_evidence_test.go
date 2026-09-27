// codex_audit_evidence_test.go — AC-CAR-012a: the route-field derivation.
//
// deriveCodexAuditEvidence (codex_audit_derive_test.go) turns the session
// records created during one audit item, plus the launcher's launch record,
// into the route fields the LIVE evidence carries. This test pins it against
// the t1100 m8-sbx records copied byte-for-byte into testdata, so a read-only
// parent + spawn_agent run whose write was really denied can never derive as
// the launcher route, and against two synthetic fixtures (one positive
// control that must derive as the launcher route).
package cli

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

const m8FixtureDir = "testdata/codex-rollouts-m8"

// m8FixtureSHA256 is the measured sha256 of each copied record (measured on
// the primary export before the copy and re-measured on the copies).
var m8FixtureSHA256 = map[string]string{
	"rollout-2026-09-24T00-41-31-01a0ceed-e360-7232-baf7-9fb82ca93398.jsonl": "9de04f41b82ecd8712718be8ac3596dbaa4feba9f973b4e750e1e6930af5bcfa",
	"rollout-2026-09-24T00-41-40-01a0ceee-0526-7e42-94a6-1073ef9cc2c3.jsonl": "50a26a625c9991dae09acfd41a5d6316af9cb3ecf007d202bfb9ec4a688a5c95",
	"rollout-2026-09-24T00-41-57-01a0ceee-4848-7e33-8077-c0b448a4e5c4.jsonl": "c116ad0a8b1b0c207535ffa307203a97fff072fc72f2420e74bb54c310b7c4ab",
	"rollout-2026-09-24T00-42-05-01a0ceee-67e8-79f1-b407-137e439cf8f1.jsonl": "6e1fa1411aadb9757f7cfdd50503ca7c379f97383984bac02e3fa521dd2fb7c2",
}

const (
	m8Run1Parent = "rollout-2026-09-24T00-41-31-01a0ceed-e360-7232-baf7-9fb82ca93398.jsonl"
	m8Run1Child  = "rollout-2026-09-24T00-41-40-01a0ceee-0526-7e42-94a6-1073ef9cc2c3.jsonl"
	m8Run2Parent = "rollout-2026-09-24T00-41-57-01a0ceee-4848-7e33-8077-c0b448a4e5c4.jsonl"
	m8Run2Child  = "rollout-2026-09-24T00-42-05-01a0ceee-67e8-79f1-b407-137e439cf8f1.jsonl"

	m8SyntheticVerdictSHA = "c58c7a67c421264c5ab781d220d82607734db38d815ade5fe61ba6cb713cc8e3"
)

func readFixture(t *testing.T, rel string) []byte {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(m8FixtureDir, rel))
	if err != nil {
		t.Fatal(err)
	}
	return b
}

func readLaunchRecordFixture(t *testing.T, rel string) *codexAuditRecord {
	t.Helper()
	var rec codexAuditRecord
	if err := json.Unmarshal(readFixture(t, rel), &rec); err != nil {
		t.Fatal(err)
	}
	return &rec
}

func TestCodexAuditEvidenceDerivation(t *testing.T) {
	// The copies are byte-identical to the measured records.
	for name, want := range m8FixtureSHA256 {
		if got := sha256Hex(readFixture(t, name)); got != want {
			t.Fatalf("%s sha256 = %s, want %s", name, got, want)
		}
	}

	type expect struct {
		sandbox       string
		spawn         bool
		route, writer string
		executed      bool
		exit          int
		denied        bool
	}
	check := func(t *testing.T, got codexAuditEvidence, w expect) {
		t.Helper()
		if got.SessionSandbox != w.sandbox || got.UsedSpawnAgent != w.spawn || got.Route != w.route ||
			got.VerdictWriter != w.writer || got.ProbeCommandExecuted != w.executed || got.WriteDenied != w.denied {
			t.Fatalf("derived %+v, want %+v", got, w)
		}
		if got.ProbeExitCode == nil || *got.ProbeExitCode != w.exit {
			t.Fatalf("probe_exit_code = %v, want %d", got.ProbeExitCode, w.exit)
		}
	}

	t.Run("m8 run2: read-only parent + spawn_agent is not the launcher", func(t *testing.T) {
		got := deriveCodexAuditEvidence(codexAuditEvidenceInput{
			Role:         "manager-docs",
			Rollouts:     [][]byte{readFixture(t, m8Run2Parent), readFixture(t, m8Run2Child)},
			ProbeCommand: "printf '%s' probe > probe-manager-docs.txt",
			ProbeExists:  false,
		})
		check(t, got, expect{sandbox: "read-only", spawn: true, route: "spawn_agent", writer: "unattributed", executed: true, exit: 1, denied: true})
	})

	t.Run("m8 run1: writing parent + spawn_agent", func(t *testing.T) {
		got := deriveCodexAuditEvidence(codexAuditEvidenceInput{
			Role:         "plan-auditor",
			Rollouts:     [][]byte{readFixture(t, m8Run1Parent), readFixture(t, m8Run1Child)},
			ProbeCommand: "printf '%s' probe > probe-plan-auditor.txt",
			ProbeExists:  true,
		})
		check(t, got, expect{sandbox: "workspace-write", spawn: true, route: "spawn_agent", writer: "unattributed", executed: true, exit: 0, denied: false})
	})

	t.Run("s2: a matching launch record does not turn a subagent into the launcher", func(t *testing.T) {
		got := deriveCodexAuditEvidence(codexAuditEvidenceInput{
			Role:          "manager-docs",
			Rollouts:      [][]byte{readFixture(t, m8Run2Parent), readFixture(t, m8Run2Child)},
			LaunchRecord:  readLaunchRecordFixture(t, "synthetic/s2-launch.json"),
			VerdictSHA256: m8SyntheticVerdictSHA,
			ProbeCommand:  "printf '%s' probe > probe-manager-docs.txt",
		})
		if got.Route != "spawn_agent" || !got.UsedSpawnAgent {
			t.Fatalf("s2 derived route=%q used_spawn_agent=%v, want spawn_agent/true", got.Route, got.UsedSpawnAgent)
		}
	})

	t.Run("s1: top-level audit session with its launch record is the launcher (positive control)", func(t *testing.T) {
		in := codexAuditEvidenceInput{
			Role:          "manager-docs",
			Rollouts:      [][]byte{readFixture(t, "synthetic/s1-toplevel-rollout.jsonl")},
			LaunchRecord:  readLaunchRecordFixture(t, "synthetic/s1-launch.json"),
			VerdictSHA256: m8SyntheticVerdictSHA,
			ProbeCommand:  "printf '%s' probe > probe-manager-docs.txt",
		}
		got := deriveCodexAuditEvidence(in)
		check(t, got, expect{sandbox: "read-only", spawn: false, route: "launcher", writer: "launcher", executed: true, exit: 1, denied: true})

		// Each launcher condition is necessary.
		noRecord := in
		noRecord.LaunchRecord = nil
		if r := deriveCodexAuditEvidence(noRecord); r.Route == "launcher" || r.VerdictWriter == "launcher" {
			t.Errorf("no launch record still derived %q/%q", r.Route, r.VerdictWriter)
		}
		wrongSHA := in
		wrongSHA.VerdictSHA256 = "0000000000000000000000000000000000000000000000000000000000000000"
		if r := deriveCodexAuditEvidence(wrongSHA); r.Route == "launcher" {
			t.Error("verdict sha mismatch still derived launcher")
		}
		failed := in
		rec := *in.LaunchRecord
		rec.ExitCode = 1
		failed.LaunchRecord = &rec
		if r := deriveCodexAuditEvidence(failed); r.Route == "launcher" {
			t.Error("failed launch record still derived launcher")
		}
		otherRole := in
		otherRole.Role = "plan-auditor"
		if r := deriveCodexAuditEvidence(otherRole); r.Route == "launcher" {
			t.Error("role mismatch still derived launcher")
		}
		otherCmd := in
		otherCmd.ProbeCommand = "printf other > x.txt"
		if r := deriveCodexAuditEvidence(otherCmd); r.ProbeCommandExecuted || r.WriteDenied {
			t.Errorf("a command the model never ran derived executed=%v denied=%v", r.ProbeCommandExecuted, r.WriteDenied)
		}
	})

	t.Run("empty input derives nothing", func(t *testing.T) {
		got := deriveCodexAuditEvidence(codexAuditEvidenceInput{Role: "plan-auditor", ProbeCommand: "x"})
		if got.Route != "unattributed" || got.VerdictWriter != "unattributed" || got.ProbeCommandExecuted || got.WriteDenied || got.ProbeExitCode != nil {
			t.Fatalf("empty input derived %+v", got)
		}
	})
}

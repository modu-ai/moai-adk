package hook

import (
	"os"
	"path/filepath"
	"testing"
)

// preToolAuditLine is a PreToolUse row as the agent-model guard writes it.
const preToolAuditLine = `{"timestamp":"2026-01-01T00:00:00Z","session_id":"s-4","agent":"manager-develop","declared_model":"opus","resolved_model":"opus","verdict":"ok"}`

func seedPreToolRow(t *testing.T, root string) {
	t.Helper()
	dir := filepath.Join(root, ".moai", "logs")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, agentModelAuditFileName), []byte(preToolAuditLine+"\n"), 0o600); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
}

// AC-SMA-004 — the SubagentStop row: its fields, the untouched PreToolUse row,
// the derived transcript path, and fail-open on every error path.
func TestServedModel_SubagentStopRecord(t *testing.T) {
	driftRows := repeatRows(assistantRow("glm-5.3-flash"), 3)
	opusMeta := map[string]string{"agentType": "manager-develop", "model": "opus"}
	wantKeys := []string{"session_id", "agent_id", "agent", "declared_model", "resolved_model",
		"served_models", "self_reported_model", "source", "verdict"}

	assertTwoRows := func(t *testing.T, root string) {
		t.Helper()
		rows, raw := readServedAuditRows(t, root)
		if len(rows) != 2 {
			t.Fatalf("audit rows = %d, want exactly 2 (%q)", len(rows), raw)
		}
		if raw[0] != preToolAuditLine {
			t.Fatalf("the PreToolUse row changed:\n got %q\nwant %q", raw[0], preToolAuditLine)
		}
		for _, k := range wantKeys {
			if _, ok := rows[1][k]; !ok {
				t.Fatalf("served row lacks key %q: %v", k, rows[1])
			}
		}
		if rows[1]["verdict"] != ServedVerdictDrift {
			t.Fatalf("served row verdict = %v, want %q", rows[1]["verdict"], ServedVerdictDrift)
		}
		if rows[1]["session_id"] != "s-4" || rows[1]["agent_id"] != "a4" || rows[1]["source"] != servedObservationSource {
			t.Fatalf("served row identity fields wrong: %v", rows[1])
		}
		if served, _ := rows[1]["served_models"].([]any); len(served) != 1 || served[0] != "glm-5.3-flash" {
			t.Fatalf("served_models = %v, want [glm-5.3-flash]", rows[1]["served_models"])
		}
	}

	t.Run("a_agent_transcript_path", func(t *testing.T) {
		root := newServedRoot(t)
		seedPreToolRow(t, root)
		tr := writeSubagentTranscript(t, filepath.Join(t.TempDir(), "subagents"), "a4", driftRows, opusMeta)
		runServedStop(t, nil, &HookInput{CWD: root, SessionID: "s-4", AgentID: "a4", AgentType: "manager-develop",
			AgentTranscriptPath: tr, HookEventName: string(EventSubagentStop)})
		assertTwoRows(t, root)
	})

	t.Run("b_derived_from_transcript_path_and_agent_id", func(t *testing.T) {
		root := newServedRoot(t)
		seedPreToolRow(t, root)
		sessDir := t.TempDir()
		writeSubagentTranscript(t, filepath.Join(sessDir, "s-4", "subagents"), "a4", driftRows, opusMeta)
		runServedStop(t, nil, &HookInput{CWD: root, SessionID: "s-4", AgentID: "a4", AgentType: "manager-develop",
			TranscriptPath: filepath.Join(sessDir, "s-4.jsonl"), HookEventName: string(EventSubagentStop)})
		assertTwoRows(t, root)
	})

	failOpen := func(t *testing.T, input *HookInput) {
		t.Helper()
		defer func() {
			if r := recover(); r != nil {
				t.Fatalf("SubagentStop panicked: %v", r)
			}
		}()
		out := runServedStop(t, nil, input)
		if out != nil && out.Decision != "" {
			t.Fatalf("served observation returned a decision: %+v", out)
		}
	}

	t.Run("c_log_dir_unwritable", func(t *testing.T) {
		root := newServedRoot(t)
		// A regular file where the logs directory belongs makes MkdirAll fail.
		if err := os.WriteFile(filepath.Join(root, ".moai", "logs"), []byte("x"), 0o600); err != nil {
			t.Fatal(err)
		}
		tr := writeSubagentTranscript(t, filepath.Join(t.TempDir(), "subagents"), "a4", driftRows, opusMeta)
		failOpen(t, &HookInput{CWD: root, SessionID: "s-4", AgentID: "a4", AgentType: "manager-develop",
			AgentTranscriptPath: tr, HookEventName: string(EventSubagentStop)})
	})

	t.Run("d_corrupt_rows_mixed_in", func(t *testing.T) {
		root := newServedRoot(t)
		rows := append([]string{"{not json", `{"type":"assistant","message":`}, driftRows...)
		tr := writeSubagentTranscript(t, filepath.Join(t.TempDir(), "subagents"), "a4", rows, opusMeta)
		failOpen(t, &HookInput{CWD: root, SessionID: "s-4", AgentID: "a4", AgentType: "manager-develop",
			AgentTranscriptPath: tr, HookEventName: string(EventSubagentStop)})
		got, _ := readServedAuditRows(t, root)
		if len(got) != 1 {
			t.Fatalf("audit rows = %d, want 1", len(got))
		}
		if got[0]["verdict"] == ServedVerdictOK {
			t.Fatalf("a corrupt-row transcript was recorded ok: %v", got[0])
		}
	})

	t.Run("e_empty_cwd", func(t *testing.T) {
		tr := writeSubagentTranscript(t, filepath.Join(t.TempDir(), "subagents"), "a4", driftRows, opusMeta)
		failOpen(t, &HookInput{SessionID: "s-4", AgentID: "a4", AgentType: "manager-develop",
			AgentTranscriptPath: tr, HookEventName: string(EventSubagentStop)})
	})
}

// AC-SMA-014 — the self-report comes from the first line of the final
// message only, is recorded, and never changes the verdict.
func TestServedModel_SelfReportFromFinalMessage(t *testing.T) {
	glmRows := repeatRows(assistantRow("glm-5.3-flash"), 2)
	meta := map[string]string{"agentType": "plan-auditor", "model": "opus"}
	cases := []struct {
		name       string
		message    string
		reportLine bool
		wantSelf   string
	}{
		{"a_first_line_self_report", "auditor-model: opus\n\nreport body\nAUDIT-VERDICT: PASS spec=SPEC-X-001 receipts=none", false, "opus"},
		{"b_only_in_report_file", "report body\nAUDIT-VERDICT: PASS spec=SPEC-X-001 receipts=none", true, ""},
		{"c_third_line", "report body\n\nauditor-model: opus\nAUDIT-VERDICT: PASS spec=SPEC-X-001 receipts=none", false, ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			root := newServedRoot(t)
			if tc.reportLine {
				if err := os.WriteFile(filepath.Join(root, "plan-audit.md"), []byte("auditor-model: opus\n# report\n"), 0o600); err != nil {
					t.Fatal(err)
				}
			}
			tr := writeSubagentTranscript(t, filepath.Join(t.TempDir(), "subagents"), "a14", glmRows, meta)
			runServedStop(t, nil, &HookInput{CWD: root, SessionID: "s-14", AgentID: "a14", AgentType: "plan-auditor",
				AgentTranscriptPath: tr, LastAssistantMessage: tc.message, HookEventName: string(EventSubagentStop)})
			rows, _ := readServedAuditRows(t, root)
			if len(rows) != 1 {
				t.Fatalf("audit rows = %d, want 1", len(rows))
			}
			if rows[0]["self_reported_model"] != tc.wantSelf {
				t.Fatalf("self_reported_model = %v, want %q", rows[0]["self_reported_model"], tc.wantSelf)
			}
			if rows[0]["verdict"] != ServedVerdictDrift {
				t.Fatalf("verdict = %v, want %q — a self-report must never override the transcript", rows[0]["verdict"], ServedVerdictDrift)
			}
		})
	}
}

package conversation

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestNativeCompactSummaryPreservesOnlyVerifiedCompletion(t *testing.T) {
	for _, mutation := range []string{"valid", "unfinished", "missing-boundary", "missing-summary", "wrong-parent", "typed-summary", "foreign-summary", "later-user", "missing-marker"} {
		t.Run(mutation, func(t *testing.T) {
			dir := t.TempDir()
			r := record{UUID: "session", CWD: "/project", ConfigDir: dir, Transcript: "native.jsonl", Completion: 1}
			assistant := map[string]any{"type": "assistant", "message": map[string]any{"role": "assistant", "model": "gpt-5.6-sol", "stop_reason": "end_turn"}}
			boundary := map[string]any{"type": "system", "subtype": "compact_boundary", "uuid": "boundary"}
			summary := map[string]any{"type": "user", "parentUuid": "boundary", "isCompactSummary": true, "isVisibleInTranscriptOnly": true, "message": map[string]any{"role": "user", "content": "This session is being continued from a previous conversation that ran out of context. The summary below covers the earlier portion of the conversation.\n\nCHILD_LIVE_OK"}}
			rows := []map[string]any{assistant, boundary, summary}
			switch mutation {
			case "unfinished":
				assistant["message"].(map[string]any)["stop_reason"] = "tool_use"
			case "missing-boundary":
				rows = []map[string]any{assistant, summary}
			case "missing-summary":
				rows = []map[string]any{assistant, boundary}
			case "wrong-parent":
				summary["parentUuid"] = "other"
			case "typed-summary":
				summary["promptSource"] = "sdk"
			case "foreign-summary":
				summary["sessionId"] = "foreign"
			case "later-user":
				rows = append(rows, map[string]any{"type": "user", "message": map[string]any{"role": "user", "content": "next request"}})
			case "missing-marker":
				delete(summary, "isCompactSummary")
			}
			var data []byte
			for _, row := range rows {
				if row["sessionId"] == nil {
					row["sessionId"] = r.UUID
				}
				row["cwd"] = r.CWD
				raw, err := json.Marshal(row)
				if err != nil {
					t.Fatal(err)
				}
				data = append(append(data, raw...), '\n')
			}
			if err := os.WriteFile(filepath.Join(dir, r.Transcript), data, 0600); err != nil {
				t.Fatal(err)
			}
			model, err := transcriptModel(r)
			if mutation == "valid" {
				if err != nil || model != "gpt-5.6-sol" {
					t.Fatalf("valid compaction: model=%q err=%v", model, err)
				}
			} else if err == nil {
				t.Fatalf("unsafe %s transcript accepted", mutation)
			}
		})
	}
}

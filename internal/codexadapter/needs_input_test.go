package codexadapter

import (
	"bufio"
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/hook"
)

// readSinkRecords returns the discard records RecordDiscards wrote under root.
func readSinkRecords(t *testing.T, root string) []Discard {
	t.Helper()
	f, err := os.Open(filepath.Join(root, DiagnosticSinkRel))
	if err != nil {
		t.Fatalf("open diagnostic sink: %v", err)
	}
	defer func() { _ = f.Close() }()

	var out []Discard
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		var d Discard
		if err := json.Unmarshal(sc.Bytes(), &d); err != nil {
			t.Fatalf("sink line is not a discard record: %q: %v", sc.Text(), err)
		}
		out = append(out, d)
	}
	if err := sc.Err(); err != nil {
		t.Fatalf("scan diagnostic sink: %v", err)
	}
	return out
}

// codexDenyReason extracts the deny reason from a Codex decision payload for
// ev, failing the test when the payload is not a deny.
func codexDenyReason(t *testing.T, ev hook.EventType, raw []byte) string {
	t.Helper()
	if strings.TrimSpace(string(raw)) == "{}" {
		t.Fatalf("%s: rendered the empty no-opinion object; a needs_input must never degrade to {}", ev)
	}
	var v map[string]any
	if err := json.Unmarshal(raw, &v); err != nil {
		t.Fatalf("%s: output is not JSON: %s", ev, raw)
	}
	hso, _ := v["hookSpecificOutput"].(map[string]any)
	switch ev {
	case hook.EventPreToolUse:
		if hso["permissionDecision"] != "deny" {
			t.Fatalf("%s: permissionDecision = %v, want deny: %s", ev, hso["permissionDecision"], raw)
		}
		reason, _ := hso["permissionDecisionReason"].(string)
		return reason
	case hook.EventPermissionRequest:
		dec, _ := hso["decision"].(map[string]any)
		if dec["behavior"] != "deny" {
			t.Fatalf("%s: decision.behavior = %v, want deny: %s", ev, dec["behavior"], raw)
		}
		reason, _ := dec["message"].(string)
		return reason
	}
	t.Fatalf("%s: not covered by this helper", ev)
	return ""
}

// TestNeedsInputVisibleDeny is AC-HPR-022 (SPEC-DUAL-HARNESS-HOOK-PARITY-001,
// REQ-HPR-008, design.md §D5, operator decision Q2): a handler result
// normalized to needs_input on a decision-bearing event becomes, on Codex, a
// deny whose reason names the required user input, and exactly one discard
// record for the conversion reaches the adapter's sink.
func TestNeedsInputVisibleDeny(t *testing.T) {
	t.Parallel()

	const handlerReason = "Critical config file: settings.json"

	for _, ev := range []hook.EventType{hook.EventPreToolUse, hook.EventPermissionRequest} {
		t.Run(string(ev), func(t *testing.T) {
			t.Parallel()

			out, discards, err := TranslateCodex(ev, DecisionNeedsInput, handlerReason)
			if err != nil {
				t.Fatalf("TranslateCodex: %v", err)
			}

			reason := codexDenyReason(t, ev, out)
			if !strings.Contains(reason, RequiredInputUserApproval) {
				t.Errorf("deny reason %q does not name the required input %q", reason, RequiredInputUserApproval)
			}
			if !strings.Contains(reason, handlerReason) {
				t.Errorf("deny reason %q dropped the handler's own reason %q", reason, handlerReason)
			}

			root := t.TempDir()
			var mirror bytes.Buffer
			if err := RecordDiscards(root, discards, false, &mirror); err != nil {
				t.Fatalf("RecordDiscards: %v", err)
			}
			records := readSinkRecords(t, root)
			if len(records) != 1 {
				t.Fatalf("sink records = %d, want exactly 1 for the needs_input conversion: %+v", len(records), records)
			}
			rec := records[0]
			if rec.Event != ev {
				t.Errorf("record event = %s, want %s", rec.Event, ev)
			}
			if !strings.Contains(rec.Reason, string(DecisionNeedsInput)) {
				t.Errorf("record reason %q does not say a needs_input was converted", rec.Reason)
			}
			if strings.Contains(rec.Reason, handlerReason) {
				t.Errorf("record reason carries the hook's content; it must carry length only: %q", rec.Reason)
			}
			if !strings.Contains(mirror.String(), string(ev)) {
				t.Errorf("stderr mirror did not announce the conversion: %q", mirror.String())
			}
		})
	}

	// The real PreToolUse output path: a Claude-shape ask from the handler
	// reaches Codex through MapOutput, not only through TranslateCodex.
	t.Run("PreToolUse ask via MapOutput", func(t *testing.T) {
		t.Parallel()

		in := []byte(`{"hookSpecificOutput":{"hookEventName":"PreToolUse","permissionDecision":"ask","permissionDecisionReason":"` + handlerReason + `"}}`)
		out, discards, err := MapOutput(hook.EventPreToolUse, in)
		if err != nil {
			t.Fatalf("MapOutput: %v", err)
		}
		reason := codexDenyReason(t, hook.EventPreToolUse, out)
		if !strings.Contains(reason, RequiredInputUserApproval) {
			t.Errorf("deny reason %q does not name the required input %q", reason, RequiredInputUserApproval)
		}

		root := t.TempDir()
		if err := RecordDiscards(root, discards, false, nil); err != nil {
			t.Fatalf("RecordDiscards: %v", err)
		}
		if records := readSinkRecords(t, root); len(records) != 1 {
			t.Fatalf("sink records = %d, want exactly 1: %+v", len(records), records)
		}
	})
}

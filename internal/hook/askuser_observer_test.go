package hook

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// askUserInput builds a PreToolUse HookInput carrying a raw AskUserQuestion
// payload. The payload is passed verbatim so a malformed one can be exercised
// as readily as a well-formed one.
func askUserInput(sessionID, payload string) *HookInput {
	return &HookInput{
		SessionID: sessionID,
		ToolName:  "AskUserQuestion",
		ToolInput: json.RawMessage(payload),
	}
}

// askUserPayloadWithLabels renders a one-question AskUserQuestion payload whose
// option labels are exactly those given.
func askUserPayloadWithLabels(labels ...string) string {
	type opt struct {
		Label       string `json:"label"`
		Description string `json:"description"`
	}
	type q struct {
		Question string `json:"question"`
		Header   string `json:"header"`
		Options  []opt  `json:"options"`
	}
	question := q{Question: "Which route?", Header: "Route"}
	for _, l := range labels {
		question.Options = append(question.Options, opt{Label: l, Description: "d"})
	}
	b, err := json.Marshal(map[string]any{"questions": []q{question}})
	if err != nil {
		panic(err)
	}
	return string(b)
}

// readAskUserRows reads every observation row written under root.
func readAskUserRows(t *testing.T, root string) []AskUserObservationRecord {
	t.Helper()
	path := filepath.Join(root, ".moai", "logs", askUserObservationFileName)
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read observation log %s: %v", path, err)
	}
	var rows []AskUserObservationRecord
	for _, line := range strings.Split(strings.TrimSpace(string(data)), "\n") {
		if line == "" {
			continue
		}
		var rec AskUserObservationRecord
		if err := json.Unmarshal([]byte(line), &rec); err != nil {
			t.Fatalf("unmarshal row %q: %v", line, err)
		}
		rows = append(rows, rec)
	}
	return rows
}

// writeInterviewMode writes a minimal interview.yaml carrying the given
// recommendation_mode value (empty string writes the key absent).
func writeInterviewMode(t *testing.T, root, mode string) {
	t.Helper()
	dir := filepath.Join(root, ".moai", "config", "sections")
	if err := os.MkdirAll(dir, 0o750); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	body := "interview:\n  enabled: true\n"
	if mode != "" {
		body += "  recommendation_mode: " + mode + "\n"
	}
	if err := os.WriteFile(filepath.Join(dir, "interview.yaml"), []byte(body), 0o600); err != nil {
		t.Fatalf("write interview.yaml: %v", err)
	}
}

func newAskUserObserverHandler(root string) *preToolHandler {
	return &preToolHandler{
		policy:     DefaultSecurityPolicy(),
		projectDir: root,
	}
}

// TestAskUserQuestionObserverRowShape asserts one row per issuance carrying
// every REQ-JFM-017 field (AC-JFM-014).
func TestAskUserQuestionObserverRowShape(t *testing.T) {
	root := t.TempDir()
	writeInterviewMode(t, root, "")

	h := newAskUserObserverHandler(root)
	payload := askUserPayloadWithLabels("Rebase (권장)", "Inspect", "Abort")

	out, err := h.Handle(context.Background(), askUserInput("sess-row", payload))
	if err != nil {
		t.Fatalf("Handle: %v", err)
	}
	assertNotDeny(t, out)

	rows := readAskUserRows(t, root)
	if len(rows) != 1 {
		t.Fatalf("row count: got %d, want exactly 1", len(rows))
	}
	rec := rows[0]
	if rec.Timestamp == "" {
		t.Error("timestamp: empty")
	}
	if rec.SessionID != "sess-row" {
		t.Errorf("session_id: got %q, want %q", rec.SessionID, "sess-row")
	}
	if rec.Mode != RecommendationModePush {
		t.Errorf("mode: got %q, want %q (absent key resolves to push)", rec.Mode, RecommendationModePush)
	}
	if !rec.LabelPresent {
		t.Error("label_present: got false, want true")
	}
	if rec.OptionCount != 3 {
		t.Errorf("option_count: got %d, want 3", rec.OptionCount)
	}
	if rec.QuestionCount != 1 {
		t.Errorf("question_count: got %d, want 1", rec.QuestionCount)
	}
	if !rec.PayloadParsed {
		t.Error("payload_parsed: got false, want true")
	}
	if rec.QuestionType != "" {
		t.Errorf("question_type: got %q, want empty — the payload carries no type tag, so no defensible classification exists (design §6.3)", rec.QuestionType)
	}
}

// TestAskUserQuestionObserverRowShapeOmitsQuestionType asserts the honesty
// constraint mechanically: an omitted classification must be ABSENT from the
// serialized row, not present-and-empty, so no later reader can select on it.
func TestAskUserQuestionObserverRowShapeOmitsQuestionType(t *testing.T) {
	root := t.TempDir()
	h := newAskUserObserverHandler(root)

	if _, err := h.Handle(context.Background(), askUserInput("sess-qt", askUserPayloadWithLabels("A"))); err != nil {
		t.Fatalf("Handle: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(root, ".moai", "logs", askUserObservationFileName))
	if err != nil {
		t.Fatalf("read log: %v", err)
	}
	if strings.Contains(string(data), "question_type") {
		t.Errorf("serialized row carries question_type; want the key absent: %s", data)
	}
}

// TestAskUserQuestionObserverLabelDetect exercises BOTH detector directions
// (AC-JFM-023). A detector that never reports true satisfies a naive
// zero-violation check while the convention is entirely unfollowed; asserting
// only the negative direction would admit exactly that mutant.
func TestAskUserQuestionObserverLabelDetect(t *testing.T) {
	cases := []struct {
		name   string
		labels []string
		want   bool
	}{
		{"KoreanLabelPresent", []string{"리베이스 (권장)", "검사"}, true},
		{"EnglishLabelPresent", []string{"Rebase (Recommended)", "Inspect"}, true},
		{"LabelOnNonFirstOption", []string{"Inspect", "Rebase (Recommended)"}, true},
		{"NoLabel", []string{"Rebase", "Inspect", "Abort"}, false},
		{"EmptyOptions", nil, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			root := t.TempDir()
			h := newAskUserObserverHandler(root)

			out, err := h.Handle(context.Background(), askUserInput("sess-detect", askUserPayloadWithLabels(tc.labels...)))
			if err != nil {
				t.Fatalf("Handle: %v", err)
			}
			assertNotDeny(t, out)

			rows := readAskUserRows(t, root)
			if len(rows) != 1 {
				t.Fatalf("row count: got %d, want 1", len(rows))
			}
			if rows[0].LabelPresent != tc.want {
				t.Errorf("label_present: got %v, want %v (labels %q)", rows[0].LabelPresent, tc.want, tc.labels)
			}
		})
	}
}

// TestAskUserQuestionObserverNeverDenies asserts CONST-7 over a table of
// payloads including malformed ones (AC-JFM-015). The observer measures; it
// never gates.
func TestAskUserQuestionObserverNeverDenies(t *testing.T) {
	cases := []struct {
		name    string
		mode    string
		payload string
	}{
		{"WellFormedWithLabel", "", askUserPayloadWithLabels("Rebase (권장)", "Abort")},
		{"WellFormedWithoutLabel", "", askUserPayloadWithLabels("Rebase", "Abort")},
		{"PullMode", "pull", askUserPayloadWithLabels("Rebase (권장)")},
		{"UnrecognizedMode", "bogus", askUserPayloadWithLabels("Rebase")},
		{"EmptyPayload", "", ""},
		{"MalformedJSON", "", `{"questions": [`},
		{"WrongShape", "", `{"questions": "not-an-array"}`},
		{"NullPayload", "", `null`},
		{"ScalarPayload", "", `42`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			root := t.TempDir()
			if tc.mode != "" {
				writeInterviewMode(t, root, tc.mode)
			}
			h := newAskUserObserverHandler(root)

			out, err := h.Handle(context.Background(), askUserInput("sess-nodeny", tc.payload))
			if err != nil {
				t.Fatalf("Handle returned an error: %v", err)
			}
			assertNotDeny(t, out)
			if out.SystemMessage != "" {
				t.Errorf("system_message: got %q, want empty — the observer displaces no established decision", out.SystemMessage)
			}
		})
	}
}

// TestAskUserQuestionObserverFailOpen exercises the three named uncertainty
// sources (AC-JFM-016): unparseable payload, absent config, unwritable log
// path. Each must allow the call with no error surfaced to the caller.
func TestAskUserQuestionObserverFailOpen(t *testing.T) {
	t.Run("UnparseablePayload", func(t *testing.T) {
		root := t.TempDir()
		h := newAskUserObserverHandler(root)

		out, err := h.Handle(context.Background(), askUserInput("sess-fo1", `{"questions": [ ###`))
		if err != nil {
			t.Fatalf("Handle returned an error: %v", err)
		}
		assertNotDeny(t, out)

		rows := readAskUserRows(t, root)
		if len(rows) != 1 {
			t.Fatalf("row count: got %d, want 1 — an unparseable payload is still an issuance", len(rows))
		}
		if rows[0].PayloadParsed {
			t.Error("payload_parsed: got true, want false — a false label_present here would be an unobserved claim, not a measurement")
		}
		if rows[0].LabelPresent {
			t.Error("label_present: got true on an unparseable payload")
		}
	})

	t.Run("AbsentConfig", func(t *testing.T) {
		root := t.TempDir() // no .moai/config at all
		h := newAskUserObserverHandler(root)

		out, err := h.Handle(context.Background(), askUserInput("sess-fo2", askUserPayloadWithLabels("A (권장)")))
		if err != nil {
			t.Fatalf("Handle returned an error: %v", err)
		}
		assertNotDeny(t, out)

		rows := readAskUserRows(t, root)
		if len(rows) != 1 {
			t.Fatalf("row count: got %d, want 1", len(rows))
		}
		if rows[0].Mode != RecommendationModePush {
			t.Errorf("mode: got %q, want %q on an absent config", rows[0].Mode, RecommendationModePush)
		}
	})

	t.Run("UnwritableLogPath", func(t *testing.T) {
		root := t.TempDir()
		// Occupy .moai/logs with a regular file so MkdirAll cannot create the
		// directory — the append path is unreachable by construction.
		if err := os.MkdirAll(filepath.Join(root, ".moai"), 0o750); err != nil {
			t.Fatalf("mkdir: %v", err)
		}
		if err := os.WriteFile(filepath.Join(root, ".moai", "logs"), []byte("occupied"), 0o600); err != nil {
			t.Fatalf("write blocker: %v", err)
		}
		h := newAskUserObserverHandler(root)

		out, err := h.Handle(context.Background(), askUserInput("sess-fo3", askUserPayloadWithLabels("A (권장)")))
		if err != nil {
			t.Fatalf("Handle returned an error: %v", err)
		}
		assertNotDeny(t, out)
	})

	t.Run("UnwritableLogFile", func(t *testing.T) {
		root := t.TempDir()
		// Occupy the log path itself with a directory so OpenFile cannot open
		// it for append — the failure lands one layer deeper than MkdirAll.
		if err := os.MkdirAll(filepath.Join(root, ".moai", "logs", askUserObservationFileName), 0o750); err != nil {
			t.Fatalf("mkdir blocker: %v", err)
		}
		h := newAskUserObserverHandler(root)

		out, err := h.Handle(context.Background(), askUserInput("sess-fo4", askUserPayloadWithLabels("A (권장)")))
		if err != nil {
			t.Fatalf("Handle returned an error: %v", err)
		}
		assertNotDeny(t, out)
	})

	t.Run("NilInput", func(t *testing.T) {
		h := newAskUserObserverHandler(t.TempDir())
		// The branch is reached only through Handle in production, but the
		// entry point states a nil-safe contract; assert it directly.
		h.observeQuestionChannel(nil)
	})

	t.Run("EmptyProjectRootWritesNothing", func(t *testing.T) {
		h := &preToolHandler{policy: DefaultSecurityPolicy(), projectDir: ""}
		// An unresolvable root must not panic and must not deny. The resolver
		// may fall back to the process CWD, so this asserts only the contract
		// the caller depends on.
		out, err := h.Handle(context.Background(), askUserInput("", askUserPayloadWithLabels("A")))
		if err != nil {
			t.Fatalf("Handle returned an error: %v", err)
		}
		assertNotDeny(t, out)
	})
}

// TestAskUserQuestionObserverModeResolution asserts the M0 defensive read of
// interview.recommendation_mode: absent resolves to push, pull resolves to
// pull, and an unrecognized value is recorded verbatim rather than coerced.
func TestAskUserQuestionObserverModeResolution(t *testing.T) {
	cases := []struct {
		name    string
		written string
		want    string
	}{
		{"AbsentKeyResolvesPush", "", RecommendationModePush},
		{"ExplicitPush", "push", RecommendationModePush},
		{"ExplicitPull", "pull", RecommendationModePull},
		{"UnrecognizedRecordedVerbatim", "bogus", "bogus"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			root := t.TempDir()
			writeInterviewMode(t, root, tc.written)
			if got := resolveRecommendationMode(root); got != tc.want {
				t.Errorf("resolveRecommendationMode: got %q, want %q", got, tc.want)
			}
		})
	}

	t.Run("UnparseableYAMLResolvesPush", func(t *testing.T) {
		root := t.TempDir()
		dir := filepath.Join(root, ".moai", "config", "sections")
		if err := os.MkdirAll(dir, 0o750); err != nil {
			t.Fatalf("mkdir: %v", err)
		}
		if err := os.WriteFile(filepath.Join(dir, "interview.yaml"), []byte("interview: [\n  broken"), 0o600); err != nil {
			t.Fatalf("write: %v", err)
		}
		if got := resolveRecommendationMode(root); got != RecommendationModePush {
			t.Errorf("resolveRecommendationMode: got %q, want %q", got, RecommendationModePush)
		}
	})
}

// TestAskUserQuestionObserverIgnoresOtherTools asserts the branch is scoped to
// AskUserQuestion — no other tool name writes an observation row.
func TestAskUserQuestionObserverIgnoresOtherTools(t *testing.T) {
	root := t.TempDir()
	h := newAskUserObserverHandler(root)

	in := askUserInput("sess-other", askUserPayloadWithLabels("A (권장)"))
	in.ToolName = "Read"

	if _, err := h.Handle(context.Background(), in); err != nil {
		t.Fatalf("Handle: %v", err)
	}
	if _, err := os.Stat(filepath.Join(root, ".moai", "logs", askUserObservationFileName)); !os.IsNotExist(err) {
		t.Errorf("observation log exists after a non-AskUserQuestion call (stat err: %v)", err)
	}
}

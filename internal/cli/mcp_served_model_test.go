package cli

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
)

// SPEC-MCP-SERVED-MODEL-001 — served-model observation on the MCP delegation
// tools. glm_task records the model the z.ai response envelope names next to
// the model it requested and warns (never refuses) on a mismatch or absence;
// codex_task records the requested model and the structural served value
// "unknown" with no warning. glm_audit's parse contract is unchanged.
//
// Warning observation (acceptance.md §A rule 7): a result or record "has no
// served-model warning" only when its JSON carries no non-empty
// served_model_warning AND its note does not contain "served" (case
// insensitive). Every negative row sits in a table with positive rows.

// ─── fixtures ───

// glmTextRespWithModel builds a z.ai response envelope whose single text block
// carries text and whose top-level `model` field carries model verbatim. A nil
// model omits the field entirely (the absent-envelope shape); a non-string
// model (for example a number) is marshalled as that JSON type.
func glmTextRespWithModel(text string, model any) string {
	env := map[string]any{
		"content": []map[string]any{
			{"type": "text", "text": text},
		},
	}
	if model != nil {
		env["model"] = model
	}
	b, _ := json.Marshal(env)
	return string(b)
}

// servedWarningText returns the served-model warning a decoded result or
// record carries: the dedicated served_model_warning key first, then any note
// text mentioning "served". Empty means "no served-model warning" under
// acceptance.md §A rule 7.
func servedWarningText(m map[string]any) string {
	if w, _ := m["served_model_warning"].(string); w != "" {
		return w
	}
	if note, _ := m["note"].(string); strings.Contains(strings.ToLower(note), "served") {
		return note
	}
	return ""
}

// recordMap renders a value through json.Marshal into a map — the observation
// surface acceptance.md §A rule 7 names for records.
func recordMap(t *testing.T, v any) map[string]any {
	t.Helper()
	raw, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("marshal %T: %v", v, err)
	}
	var m map[string]any
	if err := json.Unmarshal(raw, &m); err != nil {
		t.Fatalf("decode %s: %v", raw, err)
	}
	return m
}

// ─── AC-MSM-001 — glm_task foreground result ───

func TestGLMTask_ServedModel(t *testing.T) {
	const requested = "glm-5.3-flash"

	type row struct {
		name       string
		factory    bool   // run inside the factory-mode guard (row g)
		status     int    // HTTP status the stub returns (0 = 200)
		body       string // response envelope
		wantStatus string
		wantOutput string // checked only when wantStatus is completed
		wantServed string
		wantWarn   bool
		warnHas    []string // substrings the warning must carry
		noteHas    string   // pre-existing note text that must survive
	}
	rows := []row{
		{name: "a match — no warning (positive control)", body: glmTextRespWithModel("out-a", requested),
			wantStatus: glmJobStatusCompleted, wantOutput: "out-a", wantServed: requested},
		{name: "b mismatch — warning names both ids", body: glmTextRespWithModel("out-b", "glm-5.3"),
			wantStatus: glmJobStatusCompleted, wantOutput: "out-b", wantServed: "glm-5.3",
			wantWarn: true, warnHas: []string{requested, "glm-5.3"}},
		{name: "c absent model — absence warning", body: glmTextRespWithModel("out-c", nil),
			wantStatus: glmJobStatusCompleted, wantOutput: "out-c", wantServed: "",
			wantWarn: true, warnHas: []string{requested, "no model"}},
		{name: "d case-only difference — no warning", body: glmTextRespWithModel("out-d", "GLM-5.3-FLASH"),
			wantStatus: glmJobStatusCompleted, wantOutput: "out-d", wantServed: "GLM-5.3-FLASH"},
		{name: "e suffix-only difference — warning (no id rewriting)", body: glmTextRespWithModel("out-e", requested+"[1m]"),
			wantStatus: glmJobStatusCompleted, wantOutput: "out-e", wantServed: requested + "[1m]",
			wantWarn: true, warnHas: []string{requested + "[1m]"}},
		{name: "f HTTP 500 — failed, no warning", status: 500, body: glmTextRespWithModel("ignored", "glm-5.3"),
			wantStatus: glmJobStatusFailed, wantServed: ""},
		{name: "g factory note preserved, warning appended", factory: true, body: glmTextRespWithModel("out-g", "glm-5.3"),
			wantStatus: glmJobStatusCompleted, wantOutput: "out-g", wantServed: "glm-5.3",
			wantWarn: true, warnHas: []string{"glm-5.3"}, noteHas: "factory mode"},
		{name: "h numeric model — completed, absence warning", body: `{"model":123,"content":[{"type":"text","text":"out-h"}]}`,
			wantStatus: glmJobStatusCompleted, wantOutput: "out-h", wantServed: "",
			wantWarn: true, warnHas: []string{requested, "no model"}},
	}

	for _, tc := range rows {
		t.Run(tc.name, func(t *testing.T) {
			// Hermetic project root: the resolved default is the canonical GLM
			// model, independent of this repository's own llm config.
			withCodexProjectDir(t, t.TempDir())
			if tc.factory {
				t.Setenv(config.EnvMoaiFactoryWorkers, "8")
			} else {
				clearFactoryTestEnv(t)
			}
			withGLMTaskSeams(t, "k", &stubGLMDoer{status: tc.status, body: tc.body})

			wantModel := requested
			if tc.factory {
				wantModel = resolveGLMTaskModel()
				if strings.EqualFold(wantModel, tc.wantServed) {
					t.Fatalf("premise: the resolved default %q must differ from the served %q for this row", wantModel, tc.wantServed)
				}
			}

			res := callGLMTaskTool(t, map[string]any{"prompt": "p", "model": requested})
			if res.IsError {
				t.Fatalf("glm_task IsError: %+v", res)
			}
			got := structuredMap(t, res)

			if st, _ := got["status"].(string); st != tc.wantStatus {
				t.Fatalf("status = %q, want %q (result %v)", st, tc.wantStatus, got)
			}
			if tc.wantStatus == glmJobStatusCompleted {
				if out, _ := got["output"].(string); out != tc.wantOutput {
					t.Errorf("output = %q, want %q (observation must not alter output)", out, tc.wantOutput)
				}
			}
			if m, _ := got["model"].(string); m != wantModel {
				t.Errorf("model = %q, want the requested model %q", m, wantModel)
			}
			if _, ok := got["served_model"]; !ok {
				t.Errorf("result carries no served_model key: %v", got)
			}
			if s, _ := got["served_model"].(string); s != tc.wantServed {
				t.Errorf("served_model = %q, want %q", s, tc.wantServed)
			}

			warn := servedWarningText(got)
			switch {
			case tc.wantWarn && warn == "":
				t.Errorf("want a served-model warning, got none (result %v)", got)
			case !tc.wantWarn && warn != "":
				t.Errorf("want no served-model warning, got %q", warn)
			}
			for _, s := range tc.warnHas {
				if !strings.Contains(warn, s) {
					t.Errorf("warning %q does not name %q", warn, s)
				}
			}
			if tc.noteHas != "" {
				if note, _ := got["note"].(string); !strings.Contains(note, tc.noteHas) {
					t.Errorf("pre-existing note lost: note = %q, want it to keep %q", note, tc.noteHas)
				}
			}
		})
	}
}

// ─── AC-MSM-002 — background GLM job record ───

func TestGLMJob_ServedModel(t *testing.T) {
	const requested = "glm-5.3-flash"

	type row struct {
		name       string
		status     int
		body       string
		wantStatus string
		wantOutput string
		wantServed string
		wantWarn   bool
		warnHas    []string
	}
	rows := []row{
		{name: "a match — no warning (positive control)", body: glmTextRespWithModel("bg-a", requested),
			wantStatus: glmJobStatusCompleted, wantOutput: "bg-a", wantServed: requested},
		{name: "b mismatch — warning names both ids", body: glmTextRespWithModel("bg-b", "glm-5.3"),
			wantStatus: glmJobStatusCompleted, wantOutput: "bg-b", wantServed: "glm-5.3",
			wantWarn: true, warnHas: []string{requested, "glm-5.3"}},
		{name: "c absent model — absence warning", body: glmTextRespWithModel("bg-c", nil),
			wantStatus: glmJobStatusCompleted, wantOutput: "bg-c", wantServed: "",
			wantWarn: true, warnHas: []string{requested, "no model"}},
		{name: "e HTTP 500 — failed record, no warning", status: 500, body: glmTextRespWithModel("x", "glm-5.3"),
			wantStatus: glmJobStatusFailed, wantServed: ""},
	}

	for _, tc := range rows {
		t.Run(tc.name, func(t *testing.T) {
			root := t.TempDir()
			withCodexProjectDir(t, root)
			clearFactoryTestEnv(t)
			withGLMTaskSeams(t, "k", &stubGLMDoer{status: tc.status, body: tc.body})

			res := callGLMTaskTool(t, map[string]any{"prompt": "p", "model": requested, "background": true})
			if res.IsError {
				t.Fatalf("background glm_task IsError: %+v", res)
			}
			jobID, _ := structuredMap(t, res)["job_id"].(string)
			if jobID == "" {
				t.Fatalf("no job id: %+v", res)
			}
			waitForGLMJobToStop(t, jobID)

			// Read the record the way a caller does: through glm_job_status.
			statusRes := callGLMJobTool(t, handleGLMJobStatus, map[string]any{"job_id": jobID})
			if statusRes.IsError {
				t.Fatalf("glm_job_status failed: %+v", statusRes)
			}
			got := structuredMap(t, statusRes)

			if st, _ := got["status"].(string); st != tc.wantStatus {
				t.Fatalf("status = %q, want %q (record %v)", st, tc.wantStatus, got)
			}
			if tc.wantStatus == glmJobStatusCompleted {
				if out, _ := got["output"].(string); out != tc.wantOutput {
					t.Errorf("output = %q, want %q", out, tc.wantOutput)
				}
			}
			if m, _ := got["model"].(string); m != requested {
				t.Errorf("model = %q, want the requested model %q", m, requested)
			}
			if s, _ := got["served_model"].(string); s != tc.wantServed {
				t.Errorf("served_model = %q, want %q", s, tc.wantServed)
			}
			warn := servedWarningText(got)
			switch {
			case tc.wantWarn && warn == "":
				t.Errorf("want a served-model warning, got none (record %v)", got)
			case !tc.wantWarn && warn != "":
				t.Errorf("want no served-model warning, got %q", warn)
			}
			for _, s := range tc.warnHas {
				if !strings.Contains(warn, s) {
					t.Errorf("warning %q does not name %q", warn, s)
				}
			}
		})
	}

	t.Run("d legacy record without served fields reads cleanly", func(t *testing.T) {
		root := t.TempDir()
		reg := newGLMJobRegistry(root)
		legacy := `{"id":"glm-legacy","status":"completed","created_at":"2026-09-01T00:00:00Z",` +
			`"updated_at":"2026-09-01T00:00:00Z","model":"glm-5.3-flash","request_summary":"s","output":"o"}`
		if err := os.MkdirAll(reg.dir, 0o755); err != nil {
			t.Fatalf("mkdir: %v", err)
		}
		if err := os.WriteFile(filepath.Join(reg.dir, "glm-legacy.json"), []byte(legacy), 0o600); err != nil {
			t.Fatalf("write legacy record: %v", err)
		}
		rec, err := reg.load("glm-legacy")
		if err != nil {
			t.Fatalf("legacy record did not load: %v", err)
		}
		got := recordMap(t, rec)
		if s, _ := got["served_model"].(string); s != "" {
			t.Errorf("served_model = %q, want empty for a legacy record", s)
		}
		if m, _ := got["model"].(string); m != "glm-5.3-flash" {
			t.Errorf("model = %q, want the recorded requested model", m)
		}
	})
}

// ─── AC-MSM-003 — codex_task requested model + served "unknown" ───

func TestCodexTask_ServedModelUnknown(t *testing.T) {
	assertCodexServed := func(t *testing.T, got map[string]any, wantModel string) {
		t.Helper()
		if m, _ := got["model"].(string); m != wantModel {
			t.Errorf("model = %q, want the requested model %q", m, wantModel)
		}
		served, _ := got["served_model"].(string)
		if served != "unknown" {
			t.Errorf("served_model = %q, want the literal %q", served, "unknown")
		}
		if wantModel != "" && served == wantModel {
			t.Errorf("served_model equals the requested model %q; codex never reports what it served", wantModel)
		}
		if w := servedWarningText(got); w != "" {
			t.Errorf("codex path carries a served-model warning %q; the structural unknown must not warn", w)
		}
	}

	t.Run("a configured model — sync", func(t *testing.T) {
		writeCodexLLMFixture(t, "gpt-5-codex", "high")
		sess := withCodexSession(t, codexTaskScript("trn-a", "done-a"))

		got := structuredMap(t, callCodexTask(t, map[string]any{"prompt": "p"}))
		if st, _ := got["status"].(string); st != codexJobStatusCompleted {
			t.Fatalf("status = %q, want completed (%v)", st, got)
		}
		assertCodexServed(t, got, "gpt-5-codex")
		if sent, _ := sentParams(t, sess.sent, 1)["model"].(string); sent != "gpt-5-codex" {
			t.Errorf("thread/start model = %q, want the same value the result records", sent)
		}
	})

	t.Run("b no llm config — sync", func(t *testing.T) {
		withCodexProjectDir(t, t.TempDir())
		sess := withCodexSession(t, codexTaskScript("trn-b", "done-b"))

		got := structuredMap(t, callCodexTask(t, map[string]any{"prompt": "p"}))
		if st, _ := got["status"].(string); st != codexJobStatusCompleted {
			t.Fatalf("status = %q, want completed (%v)", st, got)
		}
		assertCodexServed(t, got, "")
		if _, has := sentParams(t, sess.sent, 1)["model"]; has {
			t.Error("thread/start carried a model key although none resolved")
		}
	})

	t.Run("c configured model — background record", func(t *testing.T) {
		root := writeCodexLLMFixture(t, "gpt-5-codex", "high")
		withCodexSession(t, codexTaskScript("trn-c", "done-c"))

		got := structuredMap(t, callCodexTask(t, map[string]any{"prompt": "p", "background": true}))
		jobID, _ := got["job_id"].(string)
		if jobID == "" {
			t.Fatalf("no job id: %v", got)
		}
		assertCodexServed(t, got, "gpt-5-codex")
		rec := awaitTerminalJob(t, newCodexJobRegistry(root), jobID)
		if rec.Status != codexJobStatusCompleted {
			t.Fatalf("record status = %q, want completed", rec.Status)
		}
		assertCodexServed(t, recordMap(t, rec), "gpt-5-codex")
	})

	t.Run("d positive control — note observation reads real note text", func(t *testing.T) {
		writeCodexLLMFixture(t, "gpt-5-codex", "high")
		withCodexSession(t, codexTaskScript("trn-d", "done-d"))

		got := structuredMap(t, callCodexTask(t, map[string]any{"prompt": "p", "resume_last": true}))
		note, _ := got["note"].(string)
		if note == "" || !strings.Contains(note, codexTaskNoPriorThreadNote) {
			t.Fatalf("note = %q, want it to carry codexTaskNoPriorThreadNote", note)
		}
		assertCodexServed(t, got, "gpt-5-codex")
	})
}

// ─── AC-MSM-004 — glm_audit parse contract unchanged ───

func TestGLMAudit_ServedModelFieldIgnored(t *testing.T) {
	review := ReviewOutput{
		Verdict:   "pass",
		Summary:   "looks fine",
		Findings:  []Finding{},
		NextSteps: []string{},
	}
	text, err := json.Marshal(review)
	if err != nil {
		t.Fatalf("marshal review: %v", err)
	}

	parse := func(model any) ReviewOutput {
		return parseGLMReview([]byte(glmTextRespWithModel(string(text), model)))
	}
	base := parse(nil) // (c) no model field
	if base.Verdict == VerdictInconclusive || base.Verdict != "pass" {
		t.Fatalf("baseline envelope verdict = %q, want pass (summary %q)", base.Verdict, base.Summary)
	}
	for name, model := range map[string]any{"a string model": "glm-5.3", "b numeric model": 123} {
		t.Run(name, func(t *testing.T) {
			got := parse(model)
			if got.Verdict == VerdictInconclusive {
				t.Fatalf("verdict fell to inconclusive: %q", got.Summary)
			}
			if !reflect.DeepEqual(got, base) {
				t.Errorf("parse result %+v differs from the model-less envelope %+v", got, base)
			}
		})
	}

	// The shared review contract gains no field.
	rt := reflect.TypeOf(ReviewOutput{})
	for i := 0; i < rt.NumField(); i++ {
		f := rt.Field(i)
		if strings.Contains(strings.ToLower(f.Name+" "+f.Tag.Get("json")), "served") {
			t.Errorf("ReviewOutput gained a served-model field %q; its contract must stay unchanged", f.Name)
		}
	}
}

package cli

// Edge coverage for the re-run of named legs (AC-AEC-002c; REQ-AEC-008): the
// input reader, the scope derivation, each reason a carry is refused, and the
// statements the result makes about what was carried or why nothing was.

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
)

func TestReadRerunLegs_Forms(t *testing.T) {
	read := func(v any, present bool) ([]string, error) {
		args := map[string]any{}
		if present {
			args[rerunLegsArg] = v
		}
		return readRerunLegs(mcp.CallToolRequest{Params: mcp.CallToolParams{Arguments: args}})
	}
	if got, err := read(nil, false); got != nil || err != nil {
		t.Errorf("absent: got %v, %v, want nil, nil", got, err)
	}
	if got, err := read(nil, true); got != nil || err != nil {
		t.Errorf("null: got %v, %v, want nil, nil", got, err)
	}
	if got, err := read([]any{}, true); got != nil || err != nil {
		t.Errorf("empty list: got %v, %v, want nil, nil", got, err)
	}
	got, err := read([]any{" GLM ", "claude", "glm"}, true)
	if err != nil || !reflect.DeepEqual(got, []string{BackendClaude, BackendGLM}) {
		t.Errorf("case, space and duplicates: got %v, %v, want [claude glm] in canonical order", got, err)
	}
	if got, err := read([]string{"codex"}, true); err != nil || !reflect.DeepEqual(got, []string{BackendCodex}) {
		t.Errorf("a typed string slice: got %v, %v, want [codex]", got, err)
	}
	if _, err := read(map[string]any{}, true); err == nil || !strings.Contains(err.Error(), rerunLegsArg) {
		t.Errorf("an object must be refused naming %s: %v", rerunLegsArg, err)
	}
}

func TestAuditedScopeOf(t *testing.T) {
	if got := auditedScopeOf("", codexTargetBaseBranch, ""); got != nil {
		t.Errorf("no tree: got %+v, want nil", got)
	}
	if got := auditedScopeOf(t.TempDir(), codexTargetBaseBranch, ""); got != nil {
		t.Errorf("a baseBranch target whose base does not resolve: got %+v, want nil (never carried from)", got)
	}
	got := auditedScopeOf(t.TempDir(), " "+codexTargetUncommitted+" ", "abc1234")
	want := AuditedScope{Target: codexTargetUncommitted, DeltaBase: "abc1234"}
	if got == nil || *got != want {
		t.Errorf("a target with no base: got %+v, want %+v", got, want)
	}
	d := newDeltaRepo(t)
	got = auditedScopeOf(d.root, codexTargetBaseBranch, "")
	if got == nil || got.Base != d.shaA {
		t.Errorf("a baseBranch target: got %+v, want base %s (the merge base)", got, d.shaA)
	}
}

func TestCarryPersistedLegs_RefusalReasons(t *testing.T) {
	scope := &AuditedScope{Target: codexTargetBaseBranch, Base: "b"}
	active := []string{BackendClaude, BackendCodex, BackendGLM}
	cfg := MultiAuditConfig{RerunLegs: []string{BackendCodex}, SessionID: "s", ProjectRoot: t.TempDir()}
	cases := []struct {
		name          string
		cfg           MultiAuditConfig
		head, digest  string
		captureFailed bool
		scope         *AuditedScope
		want          string
	}{
		{"no session", MultiAuditConfig{RerunLegs: cfg.RerunLegs, ProjectRoot: cfg.ProjectRoot}, "h", "d", false, scope, "no session_id"},
		{"capture failed", cfg, "", "", true, scope, "could not be established"},
		{"no scope", cfg, "h", "d", false, nil, "could not be established"},
		{"nothing persisted", cfg, "h", "d", false, scope, "no persisted result"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			_, carried, note := carryPersistedLegs(c.cfg, c.head, c.digest, c.captureFailed, c.scope, active)
			if len(carried) != 0 {
				t.Errorf("carried %v, want none", carried)
			}
			if !strings.Contains(note, c.want) || !strings.Contains(note, "every leg ran") {
				t.Errorf("note = %q, want it to say %q and that every leg ran", note, c.want)
			}
		})
	}
}

// The persisted copy records head, digest and scope; the result the caller receives
// does not (it stays byte-identical to the result before the fields existed).
func TestAEC002c_PersistedCopyRecordsTheState_ResultDoesNot(t *testing.T) {
	d := classWorld(t)
	first := firstFullCall(t, d, rerunSession)
	for _, k := range []string{"audited_head", "audited_digest", "audited_scope", "carried_legs"} {
		if _, ok := first[k]; ok {
			t.Errorf("the returned result carries %q; only the persisted copy may", k)
		}
	}
	raw, err := os.ReadFile(filepath.Join(d.root, ".moai", "state", "audit-multi", rerunSession+".json"))
	if err != nil {
		t.Fatal(err)
	}
	for _, k := range []string{`"audited_head"`, `"audited_digest"`, `"audited_scope"`} {
		if !strings.Contains(string(raw), k) {
			t.Errorf("the persisted result lacks %s", k)
		}
	}
}

// The result states what a re-run did: the carried legs, or why nothing was carried.
func TestAEC002c_ResultStatesWhatWasCarriedOrWhy(t *testing.T) {
	d := classWorld(t)
	firstFullCall(t, d, rerunSession)

	_, carried, _ := callRerun(t, d, rerunSession, map[string]any{"rerun_legs": []any{"codex"}}, nil)
	legs, _ := carried["carried_legs"].([]any)
	if len(legs) != 2 || legs[0] != BackendClaude || legs[1] != BackendGLM {
		t.Errorf("carried_legs = %v, want [claude glm]", legs)
	}
	if note, _ := carried["residual_risk_note"].(string); !strings.Contains(note, "carried from the persisted result") || !strings.Contains(note, "re-run: codex") {
		t.Errorf("residual_risk_note = %q, want it to name the carried and the re-run legs", note)
	}

	_, refused, _ := callRerun(t, d, "sess-never-persisted", map[string]any{"rerun_legs": []any{"codex"}}, nil)
	if note, _ := refused["residual_risk_note"].(string); !strings.Contains(note, "rerun_legs not applied") || !strings.Contains(note, "no persisted result") {
		t.Errorf("residual_risk_note = %q, want it to say the re-run was not applied and why", note)
	}
	if _, ok := refused["carried_legs"]; ok {
		t.Errorf("carried_legs present on a call that carried nothing: %v", refused["carried_legs"])
	}
}

// A leg whose gate is now off is not carried: gate off means the backend is not
// part of this call at all.
func TestAEC002c_LegWhoseGateIsOff_IsNotCarried(t *testing.T) {
	d := classWorld(t)
	firstFullCall(t, d, rerunSession)
	_, second, calls := callRerun(t, d, rerunSession, map[string]any{
		"rerun_legs": []any{"codex"},
		"gates":      map[string]any{"claude": "required", "codex": "required", "glm": "off"},
	}, nil)
	if calls.get(BackendGLM) != 0 {
		t.Errorf("glm leg ran (%d) with its gate off", calls.get(BackendGLM))
	}
	list, _ := second["per_backend_verdicts"].([]any)
	for _, e := range list {
		if m, _ := e.(map[string]any); m["backend"] == BackendGLM {
			t.Errorf("glm entry carried although its gate is off: %v", m)
		}
	}
}

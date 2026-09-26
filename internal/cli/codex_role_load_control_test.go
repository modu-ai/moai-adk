// codex_role_load_control_test.go — deterministic negative control for the
// AC-DHR-012 role-load condition: a role whose file was not loaded has no
// subagent session, returns no nonce, and the role-load check is false — both
// the Go predicate the LIVE test asserts and the role-load clause of the
// AC-DHR-012 jq judge, run over an evidence file of the same shape.
package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// acDHR012RoleLoadClause is the role-load clause of the AC-DHR-012 judge,
// verbatim, with the evidence object as the input instead of $e.
const acDHR012RoleLoadClause = `([.roles[]|select((.nonce_sent|type)=="string" and (.nonce_sent|length)>0 and .nonce_returned==.nonce_sent)]|length)==12 and ([.roles[].nonce_sent]|unique|length)==12`

func TestCodexRoleLoadNegativeControl(t *testing.T) {
	roles := codexRoleNames(t, filepath.Join("..", "..", "internal", "template", "templates", ".codex", "agents", "moai"))
	if len(roles) != 12 {
		t.Fatalf("emitted roles = %d, want 12", len(roles))
	}
	// The unloaded role's item: the recorded sessions of an item that ran a
	// different role (t1100 m8 run1), so no subagent ran under this role.
	var rollouts []codexRollout
	for _, f := range []string{m8Run1Parent, m8Run1Child} {
		r := parseCodexRollout(readFixture(t, f))
		rollouts = append(rollouts, r)
	}
	const unloaded = "e2e-tester"

	build := func(loadUnloaded bool) []codexRoleLoad {
		var loads []codexRoleLoad
		for i, role := range roles {
			nonce := fmt.Sprintf("%032x", i+1)
			if role == unloaded && !loadUnloaded {
				loads = append(loads, codexRoleLoadFrom(role, nonce, rollouts))
				continue
			}
			loaded := []codexRollout{{Subagent: true, Role: role, Final: "NONCE " + nonce}}
			loads = append(loads, codexRoleLoadFrom(role, nonce, loaded))
		}
		return loads
	}

	positive := build(true)
	negative := build(false)
	for _, l := range negative {
		if l.Name == unloaded && (l.NonceReturned != "" || l.SubagentRoleObserved != "") {
			t.Fatalf("unloaded role extracted %+v", l)
		}
	}
	if !codexRoleLoadsOK(roles, positive) {
		t.Fatal("positive control: every role loaded, yet the role-load check is false")
	}
	if codexRoleLoadsOK(roles, negative) {
		t.Fatal("negative control: a role was not loaded, yet the role-load check is true")
	}

	jq, err := exec.LookPath("jq")
	if err != nil {
		t.Fatalf("jq is required for the judge-clause control: %v", err)
	}
	judge := func(loads []codexRoleLoad) string {
		b, _ := json.Marshal(map[string]any{"roles": loads})
		path := filepath.Join(t.TempDir(), "ac012-evidence.json")
		if err := os.WriteFile(path, b, 0o644); err != nil {
			t.Fatal(err)
		}
		out, _ := exec.Command(jq, acDHR012RoleLoadClause, path).Output()
		return string(out)
	}
	pos, neg := judge(positive), judge(negative)
	t.Logf("AC-DHR-012 role-load clause, every role loaded: %s", strings.TrimSpace(pos))
	t.Logf("AC-DHR-012 role-load clause, %s not loaded: %s", unloaded, strings.TrimSpace(neg))
	if pos != "true\n" || neg != "false\n" {
		t.Fatalf("judge clause printed %q (all loaded) and %q (one unloaded), want true and false", pos, neg)
	}
}

// TestCodexRoleLoadProbeWordingIsUniform pins that every role receives the
// same framing and nonce line; only the positive-control role swaps its
// no-command sentence for its write step.
func TestCodexRoleLoadProbeWordingIsUniform(t *testing.T) {
	roles := codexRoleNames(t, filepath.Join("..", "..", "internal", "template", "templates", ".codex", "agents", "moai"))
	const nonce = "0123456789abcdef0123456789abcdef"
	for _, role := range roles {
		msg := codexRoleLoadProbe(role, nonce)
		if !strings.HasPrefix(msg, codexRoleLoadProbeFrame+" ") || !strings.HasSuffix(msg, "Your final response must be exactly one line and nothing else: NONCE "+nonce) {
			t.Errorf("%s probe does not carry the shared frame and nonce line: %q", role, msg)
		}
		if strings.Contains(msg, role) {
			t.Errorf("%s probe names the role — the wording must not vary by role: %q", role, msg)
		}
		if role != codexRolePositiveRole && msg != codexRoleLoadProbe(roles[0], nonce) && roles[0] != codexRolePositiveRole {
			t.Errorf("%s probe differs from %s probe", role, roles[0])
		}
	}
}

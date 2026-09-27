package hook

// contract_sign_guard_test.go — the contract-sign and contract-decide guard
// (SPEC-AUTONOMY-PRECONDITION-001 M2, REQ-AP-003 / REQ-AP-004 / REQ-AP-005 /
// REQ-AP-009 / REQ-AP-011 / REQ-AP-012; design.md §C).
//
// Eight criteria, eight tests (AC-AP-017's constant test lives in
// internal/config; the AC-AP-018 pins live in internal/cli and
// internal/kanban):
//
//	AC-AP-005  TestContractSignAgentInvocationDenied
//	AC-AP-006  TestContractSignPositiveControlsAllowed
//	AC-AP-007  TestContractSignUnclassifiedDeniedClosed
//	AC-AP-008  TestContractSignDeniedBeforeVerbExists
//	AC-AP-015  TestContractRoleScopedDenyUnderWorkerMarker
//	AC-AP-016  TestContractRoleScopedAllowWithoutWorkerMarker
//	AC-AP-009  TestContractSignWrapperBypassesDenied
//	AC-AP-010  TestContractSignUnknownWrapperFailsClosed
//
// The guard never executes the command it inspects (PreToolUse runs before
// execution), so every deny here is produced by the parse, not by a command
// failure — and the one criterion that needs the verb to be unimplemented
// (AC-AP-008) measures that precondition through a stub binary FIRST, so the
// case cannot pass vacuously by the command simply failing.

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
)

// signGuardInput builds a Bash PreToolUse payload carrying command.
func signGuardInput(t *testing.T, command string) *HookInput {
	t.Helper()
	raw, err := json.Marshal(map[string]any{"command": command})
	if err != nil {
		t.Fatalf("marshal tool input: %v", err)
	}
	return &HookInput{
		SessionID:     "s-sign",
		HookEventName: "PreToolUse",
		ToolName:      "Bash",
		ToolInput:     raw,
	}
}

// TestContractSignAgentInvocationDenied pins AC-AP-005: with no role marker
// in the session, every human-path `moai contract sign` shape — plain,
// --signer human, quoted, behind a leading assignment, behind a wrapper, on
// a full path, behind global flags, and one -c level deep — is denied with a
// reason whose first token is the sentinel, and no supplied case is silently
// skipped.
func TestContractSignAgentInvocationDenied(t *testing.T) {
	cases := []string{
		"moai contract sign SPEC-X-001",
		"moai contract sign --signer human SPEC-X-001",
		`'moai' contract "sign"`,
		"FOO=1 moai contract sign",
		"env FOO=1 moai contract sign",
		"command moai contract sign",
		"exec moai contract sign",
		"~/go/bin/moai contract sign",
		"./bin/moai contract sign",
		"moai --no-color contract sign",
		"sh -c 'moai contract sign'",
	}
	denied := 0
	for _, command := range cases {
		decision, reason := checkContractSign(signGuardInput(t, command))
		if decision != DecisionDeny {
			t.Errorf("%q: decision = %q (%q), want deny", command, decision, reason)
			continue
		}
		denied++
		if fields := strings.Fields(reason); len(fields) == 0 || fields[0] != contractSignViolationPrefix {
			t.Errorf("%q: reason %q does not start with the sentinel %q", command, reason, contractSignViolationPrefix)
		}
	}
	if denied != len(cases) {
		t.Fatalf("denied %d of %d supplied cases — cases were silently skipped", denied, len(cases))
	}
}

// TestContractSignPositiveControlsAllowed pins AC-AP-006: the five positive
// controls are allowed with empty decision and reason (hook output
// byte-identical to the no-guard baseline; the guard writes no audit line on
// an allow), and the trailing human-path deny is the armed control proving
// the same guard that allowed them is active.
func TestContractSignPositiveControlsAllowed(t *testing.T) {
	controls := []string{
		"moai contract show SPEC-X-001",
		"moai contract verify SPEC-X-001",
		`echo "moai contract sign"`,
		"git log --grep sign",
		`git commit -m "sign the contract"`,
	}
	allowed := 0
	for _, command := range controls {
		decision, reason := checkContractSign(signGuardInput(t, command))
		if decision != "" || reason != "" {
			t.Errorf("%q: decision = %q reason = %q, want allow with unchanged hook output", command, decision, reason)
			continue
		}
		allowed++
	}
	decision, reason := checkContractSign(signGuardInput(t, "moai contract sign SPEC-X-001"))
	if decision != DecisionDeny || !strings.HasPrefix(reason, contractSignViolationPrefix) {
		t.Fatalf("armed control: decision = %q reason = %q — the guard that allowed the controls is not active", decision, reason)
	}
	if allowed != len(controls) {
		t.Fatalf("allowed %d of %d supplied controls — controls were silently skipped", allowed, len(controls))
	}
}

// TestContractSignUnclassifiedDeniedClosed pins AC-AP-007: command
// substitution, a variable in program position, eval, and nesting deeper
// than one -c level are each denied with the sentinel AND the literal token
// `unclassified` — a guard that denied by falling through the classified
// path would pass the deny limb and fail the token limb.
func TestContractSignUnclassifiedDeniedClosed(t *testing.T) {
	cases := []string{
		"$(which moai) contract sign",
		"$M contract sign",
		`eval "moai contract sign"`,
		`sh -c 'bash -c "moai contract sign"'`,
	}
	denied := 0
	for _, command := range cases {
		decision, reason := checkContractSign(signGuardInput(t, command))
		if decision != DecisionDeny {
			t.Errorf("%q: decision = %q (%q), want deny", command, decision, reason)
			continue
		}
		denied++
		if !strings.HasPrefix(reason, contractSignViolationPrefix) {
			t.Errorf("%q: reason %q missing the sentinel prefix", command, reason)
		}
		if !strings.Contains(reason, "unclassified") {
			t.Errorf("%q: reason %q missing the literal token unclassified", command, reason)
		}
	}
	if denied != len(cases) {
		t.Fatalf("denied %d of %d supplied cases — cases were silently skipped", denied, len(cases))
	}
}

// TestContractSignDeniedBeforeVerbExists pins AC-AP-008: with a stub moai
// binary on the PATH the guard resolves — one whose `contract` verb does not
// exist — the precondition is measured FIRST (the stub's `moai contract
// --help` exits non-zero with the unknown-command marker), and only then is
// the deny asserted, for sign on the human path with no marker and for
// decide in a worker session. PreToolUse runs before execution, so no
// command failure can produce the deny.
func TestContractSignDeniedBeforeVerbExists(t *testing.T) {
	stubDir := t.TempDir()
	stub := filepath.Join(stubDir, "moai")
	script := "#!/bin/sh\necho 'Unknown command \"contract\"' >&2\nexit 1\n"
	if err := os.WriteFile(stub, []byte(script), 0o755); err != nil {
		t.Fatalf("write stub moai binary: %v", err)
	}
	t.Setenv("PATH", stubDir)

	// The precondition is a measured value, not an assumption.
	help := exec.Command("moai", "contract", "--help")
	var stderr strings.Builder
	help.Stderr = &stderr
	if err := help.Run(); err == nil {
		t.Fatalf("precondition: stub `moai contract --help` exited 0 (stderr %q) — the verb-exists premise is unmet", stderr.String())
	}
	if !strings.Contains(stderr.String(), `Unknown command "contract"`) {
		t.Fatalf("precondition: stub stderr %q missing the unknown-command marker", stderr.String())
	}

	decision, reason := checkContractSign(signGuardInput(t, "moai contract sign SPEC-X-001"))
	if decision != DecisionDeny || !strings.HasPrefix(reason, contractSignViolationPrefix) {
		t.Fatalf("sign: decision = %q reason = %q, want the deny sentinel despite the unimplemented verb", decision, reason)
	}
	t.Setenv(config.EnvFactoryRole, config.FactoryRoleWorker)
	decision, reason = checkContractSign(signGuardInput(t, "moai contract decide SPEC-X-001"))
	if decision != DecisionDeny || !strings.HasPrefix(reason, contractSignViolationPrefix) {
		t.Fatalf("decide: decision = %q reason = %q, want the deny sentinel despite the unimplemented verb", decision, reason)
	}
}

// TestContractRoleScopedDenyUnderWorkerMarker pins AC-AP-015: in a session
// whose MOAI_FACTORY_ROLE is the worker value (set by this test itself, per
// REQ-AP-012), the non-interactive sign path and every decide shape —
// including behind sudo and eval — is denied; the eval case additionally
// carries the unclassified token; and no supplied case is silently skipped.
func TestContractRoleScopedDenyUnderWorkerMarker(t *testing.T) {
	t.Setenv(config.EnvFactoryRole, config.FactoryRoleWorker)
	cases := []struct {
		command      string
		unclassified bool
	}{
		{command: "moai contract sign --signer llm --receipt /tmp/r.json"},
		{command: "moai contract sign --signer llm+jev --receipt /tmp/r.json"},
		{command: "moai contract decide SPEC-X-001"},
		{command: `'moai' contract "decide"`},
		{command: "sudo moai contract decide"},
		{command: `eval "moai contract decide"`, unclassified: true},
	}
	denied := 0
	for _, tc := range cases {
		decision, reason := checkContractSign(signGuardInput(t, tc.command))
		if decision != DecisionDeny {
			t.Errorf("%q: decision = %q (%q), want deny in a marked session", tc.command, decision, reason)
			continue
		}
		denied++
		if !strings.HasPrefix(reason, contractSignViolationPrefix) {
			t.Errorf("%q: reason %q missing the sentinel prefix", tc.command, reason)
		}
		if tc.unclassified && !strings.Contains(reason, "unclassified") {
			t.Errorf("%q: reason %q missing the literal token unclassified", tc.command, reason)
		}
	}
	if denied != len(cases) {
		t.Fatalf("denied %d of %d supplied cases — cases were silently skipped", denied, len(cases))
	}
}

// TestContractRoleScopedAllowWithoutWorkerMarker pins AC-AP-016: in a session
// with no marker, and in one whose marker holds a value other than the worker
// value, every AC-AP-015 call is allowed with empty decision and reason (the
// lead session's own decide path) — and the two human-path calls are denied
// with the sentinel in BOTH runs, the armed control that keeps the allow arm
// from passing on an absent guard.
func TestContractRoleScopedAllowWithoutWorkerMarker(t *testing.T) {
	roleScoped := []string{
		"moai contract sign --signer llm --receipt /tmp/r.json",
		"moai contract sign --signer llm+jev --receipt /tmp/r.json",
		"moai contract decide SPEC-X-001",
		`'moai' contract "decide"`,
		"sudo moai contract decide",
		`eval "moai contract decide"`,
	}
	humanPath := []string{
		"moai contract sign SPEC-X-001",
		"moai contract sign --signer human SPEC-X-001",
	}
	runs := []struct {
		name string
		role string
	}{
		// An empty value reads as unset: the marker check is equality with
		// the worker value constant.
		{"marker unset", ""},
		{"marker other value", "lead"},
	}
	for _, run := range runs {
		t.Run(run.name, func(t *testing.T) {
			t.Setenv(config.EnvFactoryRole, run.role)
			for _, command := range roleScoped {
				decision, reason := checkContractSign(signGuardInput(t, command))
				if decision != "" || reason != "" {
					t.Errorf("%q: decision = %q reason = %q, want allow with unchanged hook output", command, decision, reason)
				}
			}
			for _, command := range humanPath {
				decision, reason := checkContractSign(signGuardInput(t, command))
				if decision != DecisionDeny || !strings.HasPrefix(reason, contractSignViolationPrefix) {
					t.Errorf("%q: decision = %q reason = %q, want the deny sentinel in the %q run", command, decision, reason, run.name)
				}
			}
		})
	}
}

// TestContractSignWrapperBypassesDenied pins AC-AP-009: every closed-list
// wrapper shape is denied, with each wrapper's own options skipped rather
// than read as the program word — the reason names `moai` as the resolved
// program, which is the limb a wrapper that mis-read an option as the
// program would fail.
func TestContractSignWrapperBypassesDenied(t *testing.T) {
	cases := []string{
		"script -q /dev/null moai contract sign",
		"timeout 30 moai contract sign",
		"sudo moai contract sign",
		"sudo -n moai contract sign",
		"stdbuf -oL moai contract sign",
		"nice -n 10 moai contract sign",
		"nohup moai contract sign",
		"xargs -n1 moai contract sign",
	}
	denied := 0
	for _, command := range cases {
		decision, reason := checkContractSign(signGuardInput(t, command))
		if decision != DecisionDeny {
			t.Errorf("%q: decision = %q (%q), want deny", command, decision, reason)
			continue
		}
		denied++
		if !strings.HasPrefix(reason, contractSignViolationPrefix) {
			t.Errorf("%q: reason %q missing the sentinel prefix", command, reason)
		}
		if !strings.Contains(reason, "moai") {
			t.Errorf("%q: reason %q does not name the resolved program moai — a wrapper option was read as the program word", command, reason)
		}
	}
	if denied != len(cases) {
		t.Fatalf("denied %d of %d supplied cases — cases were silently skipped", denied, len(cases))
	}
}

// TestContractSignUnknownWrapperFailsClosed pins AC-AP-010: a wrapper name
// absent from the closed list (`chrt 0 moai contract sign`) is denied as
// unclassified under REQ-AP-004 rather than allowed.
//
// AC-AP-010's mutation limb is a run-phase act recorded in progress.md §E.2,
// not a limb of this test: a mutant with the unknown-wrapper branch removed
// flips this case from denied to allowed while
// TestContractSignWrapperBypassesDenied stays green. The test stays
// single-sided so the mutation probe has a stable subject.
func TestContractSignUnknownWrapperFailsClosed(t *testing.T) {
	decision, reason := checkContractSign(signGuardInput(t, "chrt 0 moai contract sign"))
	if decision != DecisionDeny || !strings.HasPrefix(reason, contractSignViolationPrefix) {
		t.Fatalf("chrt 0 moai contract sign: decision = %q reason = %q, want the deny sentinel", decision, reason)
	}
	if !strings.Contains(reason, "unclassified") {
		t.Fatalf("reason %q missing the literal token unclassified — the case must be denied under REQ-AP-004, not the classified path", reason)
	}
}

package hook

// protected_zone_shell_matrix_test.go — card t1574 M2: the parsing-matrix
// sweep of REQ-ZSP-007. Every cell is one shell-parsing shape crossing the
// M1 fix classes (declaration forms, wrapper prefixes, recursion, `--`
// operands, executable paths, quoting) and the audit repairs (D1 wrapper ×
// declared function, D2 branch-width stress, D5 seed deny, D6 cd-chain set
// budget, D9 absent-manifest denial). Every zone-covered cell asserts deny
// and every unbounded cell asserts the fail-closed deny verdict — never
// termination alone — and the runner fails loudly on an empty cell list, so
// a selector that matches nothing cannot pass for a sweep.

import (
	"fmt"
	"testing"
)

// zoneMatrixVerdict is the disposition one matrix cell asserts.
type zoneMatrixVerdict int

const (
	zoneMatrixDenyProbe     zoneMatrixVerdict = iota // deny, category=probe_zone
	zoneMatrixDenyUnbounded                          // deny, category=loop-unbounded
	zoneMatrixAllow                                  // no decision — the call passes
)

// zoneMatrixCell is one sweep cell: the command, the zone paths the fixture
// declares (zone_dir/ unless the cell says otherwise), whether the fixture
// root carries NO manifest (the audit-D9 shape), and the asserted verdict.
type zoneMatrixCell struct {
	name           string
	command        string
	paths          string
	absentManifest bool
	want           zoneMatrixVerdict
}

// zoneMatrixWidthStress is the audit-D2 branch-width stress body: a width-10
// sibling recursion whose total work a depth-only bound would not survive
// (Θ(10^7) body walks over the recursion bound of 8); the global visit
// budget aborts it and the walk answers fail-closed.
const zoneMatrixWidthStress = "w() { w; w; w; w; w; w; w; w; w; w; }; w"

// zoneMatrixCdChain is a 16-statement cd chain whose possible-directory set
// outgrows zoneCwdsBound — the audit-D6 shape.
const zoneMatrixCdChain = "cd d01; cd d02; cd d03; cd d04; cd d05; cd d06; cd d07; cd d08; " +
	"cd d09; cd d10; cd d11; cd d12; cd d13; cd d14; cd d15; cd d16"

// zoneParsingMatrixCells enumerates the sweep. Groups mirror REQ-ZSP-007's
// minimum catalogue; each group comment names its REQ/audit anchor.
func zoneParsingMatrixCells() []zoneMatrixCell {
	cells := []zoneMatrixCell{}

	add := func(name, command string, want zoneMatrixVerdict) {
		cells = append(cells, zoneMatrixCell{name: name, command: command, paths: "zone_dir/", want: want})
	}
	// dash-dash cells target a zone whose name starts with a hyphen
	addDash := func(name, command string, want zoneMatrixVerdict) {
		cells = append(cells, zoneMatrixCell{name: name, command: command, paths: "-zone/", want: want})
	}

	// -- declaration forms crossing the K2 class (REQ-ZSP-002/007): the
	// conditional variants deny via the real-verb interpretation, the
	// certain shadows stay allow (real bash would not delete).
	add("conditional_decl_paren_form", "false && rm() { :; }; rm zone_dir/x", zoneMatrixDenyProbe)
	add("conditional_decl_kw_form", "false && function rm { :; }; rm zone_dir/x", zoneMatrixDenyProbe)
	add("conditional_decl_if_branch", "if false; then rm() { :; }; fi; rm zone_dir/x", zoneMatrixDenyProbe)
	add("conditional_decl_for_body", "for i in 1; do rm() { :; }; done; rm zone_dir/x", zoneMatrixDenyProbe)
	add("conditional_decl_while_body", "while false; do rm() { :; }; done; rm zone_dir/x", zoneMatrixDenyProbe)
	add("conditional_decl_case_arm", "case x in x) rm() { :; };; esac; rm zone_dir/x", zoneMatrixDenyProbe)
	add("certain_shadow_control_paren", "rm() { :; }; rm zone_dir/x", zoneMatrixAllow)
	add("certain_shadow_control_kw", "function rm { :; }; rm zone_dir/x", zoneMatrixAllow)

	// -- static wrapper prefixes (REQ-ZSP-003/007): verbs under the
	// enumerated set, zone and non-zone targets, the controls.
	add("wrapper_command", "command rm zone_dir/a.log", zoneMatrixDenyProbe)
	add("wrapper_command_p", "command -p rm zone_dir/a.log", zoneMatrixDenyProbe)
	add("wrapper_env", "env rm zone_dir/a.log", zoneMatrixDenyProbe)
	add("wrapper_env_assignment", "env FOO=bar rm zone_dir/a.log", zoneMatrixDenyProbe)
	add("wrapper_nohup", "nohup rm zone_dir/a.log", zoneMatrixDenyProbe)
	add("wrapper_builtin", "builtin rm zone_dir/a.log", zoneMatrixDenyProbe)
	add("wrapper_nested_env_command", "env command rm zone_dir/a.log", zoneMatrixDenyProbe)
	add("wrapper_command_cp", "command cp zone_dir/a.log /tmp/matrix-out", zoneMatrixDenyProbe)
	add("wrapper_command_mv", "command mv zone_dir/a.log /tmp/matrix-out", zoneMatrixDenyProbe)
	add("wrapper_nonzone_target", "command rm /tmp/matrix-harmless.txt", zoneMatrixAllow)
	add("wrapper_command_ls_control", "command ls zone_dir", zoneMatrixAllow)
	add("wrapper_bare_env_control", "env X=1", zoneMatrixAllow)
	add("wrapper_nohup_cat_control", "nohup cat zone_dir/a.log", zoneMatrixAllow)

	// -- the audit-D1 cross: a declared function shadowing a mutation verb,
	// invoked through EACH enumerated wrapper — the stripped head never
	// re-enters the funcs table, so every cell denies, never walking the
	// no-op body (REQ-ZSP-003/007).
	add("d1_shadow_command", "rm() { :; }; command rm zone_dir/x", zoneMatrixDenyProbe)
	add("d1_shadow_command_p", "rm() { :; }; command -p rm zone_dir/x", zoneMatrixDenyProbe)
	add("d1_shadow_env", "rm() { :; }; env rm zone_dir/x", zoneMatrixDenyProbe)
	add("d1_shadow_nohup", "rm() { :; }; nohup rm zone_dir/x", zoneMatrixDenyProbe)
	add("d1_shadow_builtin", "rm() { :; }; builtin rm zone_dir/x", zoneMatrixDenyProbe)

	// -- recursion shapes (REQ-ZSP-005/006/007): sibling, mutual, loop-
	// nested. Each terminates and answers a fail-closed deny — the covered
	// rm after the (bounded) recursion is judged on its own merits, so the
	// deny rides the candidate (probe_zone), never termination alone.
	add("recursion_sibling_seed_d5", "f() { if false; then f; f; fi; }; f; rm zone_dir/a.log", zoneMatrixDenyProbe)
	add("recursion_mutual", "a() { b; }; b() { a; }; a; rm zone_dir/x", zoneMatrixDenyProbe)
	add("recursion_loop_nested", "f() { if false; then f; fi; }; for i in 1 2; do f; done; rm zone_dir/x", zoneMatrixDenyProbe)
	// the audit-D2 branch-width stress: the visit budget aborts the walk
	// BEFORE the rm statement, so mutating stays false and the verdict is
	// the unbounded denial itself — the cell an ordering mutant (unbounded
	// judged after the !w.mutating short-circuit) fails by construction.
	add("recursion_width_stress_d2", zoneMatrixWidthStress+"; rm zone_dir/a.log", zoneMatrixDenyUnbounded)
	add("recursion_unbounded_loop", "while true; do cd a; done; rm b/x", zoneMatrixDenyUnbounded)

	// -- `--` operand positions (REQ-ZSP-001/007): first argument, after
	// options, across rm/cp/mv; the negative control stays allow.
	addDash("dash_dash_first_arg", "rm -- -zone/secret.md", zoneMatrixDenyProbe)
	addDash("dash_dash_after_option", "rm -f -- -zone/secret.md", zoneMatrixDenyProbe)
	add("dash_dash_cp", "cp -- zone_dir/a.log /tmp/matrix-out", zoneMatrixDenyProbe)
	add("dash_dash_mv", "mv -- zone_dir/a.log /tmp/matrix-out", zoneMatrixDenyProbe)
	add("dash_dash_outside_control", "rm -- outside.txt", zoneMatrixAllow)

	// -- the audit-D6 cd-chain cell: the directory set outgrows the set
	// budget. The bare chain is not mutating, so the unbounded denial is
	// the verdict itself; the chain followed by a covered rm records the
	// candidate precedence (the collected candidate is the more specific
	// verdict the preserved family freezes).
	add("d6_cd_chain_budget", zoneMatrixCdChain, zoneMatrixDenyUnbounded)
	add("d6_cd_chain_then_covered_rm", zoneMatrixCdChain+"; rm zone_dir/a.log", zoneMatrixDenyProbe)

	// -- the audit-D9 absent-manifest cell: a budget-exhausting shape
	// followed by the frozen-instruction baseline path under a fixture root
	// with NO zone manifest — asserting deny, so an ordering mutant or a
	// manifest-exception mutant fails the cell by construction.
	cells = append(cells, zoneMatrixCell{
		name:           "d9_absent_manifest_budget_deny",
		command:        zoneMatrixWidthStress + "; rm .claude/hooks/a.sh",
		absentManifest: true,
		want:           zoneMatrixDenyUnbounded,
	})

	// -- executable-path forms (REQ-ZSP-004/007): the path word never
	// resolves to a declared function, with and without a shadow; a path
	// whose base matches no verb stays allow.
	add("exec_path_no_shadow", "/bin/rm zone_dir/secret.md", zoneMatrixDenyProbe)
	add("exec_path_with_shadow", "rm() { :; }; /bin/rm zone_dir/secret.md", zoneMatrixDenyProbe)
	add("exec_path_relative", "./rm zone_dir/a.log", zoneMatrixDenyProbe)
	add("exec_path_nonverb_control", "/bin/ls zone_dir/a.log", zoneMatrixAllow)

	// -- quoting combinations crossing each class (single, double, escaped
	// — exercising zoneWordText as it exists; card t1570 owns its internals).
	addDash("quoting_single_dashdash", "rm -- '-zone/x'", zoneMatrixDenyProbe)
	add("quoting_double_wrapper", "command rm \"zone_dir/a.log\"", zoneMatrixDenyProbe)
	add("quoting_escaped", "rm zone_dir/a\\.log", zoneMatrixDenyProbe)
	add("quoting_double_conditional", "false && rm() { :; }; rm \"zone_dir/x\"", zoneMatrixDenyProbe)
	add("quoting_single_exec_path", "rm() { :; }; '/bin/rm' zone_dir/x", zoneMatrixDenyProbe)

	if len(cells) == 0 {
		return nil
	}
	return cells
}

// TestProtectedZoneShellParsingMatrix sweeps the REQ-ZSP-007 matrix: one
// subtest per cell, each printing its own --- PASS line so the swept set is
// countable in -v output and a zero-match selector cannot pass for a sweep.
func TestProtectedZoneShellParsingMatrix(t *testing.T) {
	cells := zoneParsingMatrixCells()
	if len(cells) == 0 {
		t.Fatal("empty parsing-matrix cell list — the sweep would assert nothing (REQ-ZSP-007)")
	}
	t.Logf("parsing-matrix sweep: %d cells", len(cells))
	for _, tc := range cells {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			paths := tc.paths
			if paths == "" {
				paths = "zone_dir/"
			}
			var root string
			if tc.absentManifest {
				// the audit-D9 fixture: a project root with no zone manifest
				root = newZoneRoot(t, "", "")
			} else {
				root = newZoneRoot(t, zoneShippedDoc(fmt.Sprintf("  probe_zone:\n    paths: [%q]\n", paths)), "")
			}
			h := zoneTestHandler(t, root)

			d, r := zoneCall(t, h, "Bash", harnessLearnerIdentity, map[string]any{"command": tc.command})
			switch tc.want {
			case zoneMatrixDenyProbe:
				wantZoneDeny(t, tc.command, d, r, harnessLearnerIdentity, "category", "probe_zone")
			case zoneMatrixDenyUnbounded:
				wantZoneDeny(t, tc.command, d, r, harnessLearnerIdentity, "category", "loop-unbounded")
			case zoneMatrixAllow:
				// the family's allow convention: the control fails only on a
				// deny — the permission layer may answer the passing call in
				// its own vocabulary (protected_zone_guard_test.go)
				if d == DecisionDeny {
					t.Errorf("%s: decision=%q reason=%q, want allow", tc.command, d, r)
				}
			}
		})
	}
}

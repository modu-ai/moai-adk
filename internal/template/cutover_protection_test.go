// cutover_protection_test.go: AC-GFD-020 (SPEC-GITHUB-FLOW-DEFAULT-001 M6,
// REQ-GFD-020, design D-22) for scripts/cutover-protection-compare.sh, the
// read-only comparison of the observed GitHub protection state against the
// expected state.
//
// The script's only door to GitHub is the injectable CUTOVER_GH_CMD; the tests
// point it at a stub that answers `gh api <endpoint> [--jq expr]` from canned JSON
// with the same --jq semantics the real CLI applies, and logs every call. No
// network is involved. The real-data path (the same script against the live
// repository) is a Gap this suite cannot close.
package template_test

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

const cvoProtectionRepo = "modu-ai/moai-adk"

const cvoGhAPIStub = `#!/usr/bin/env bash
set -u
echo "$*" >> "${GH_API_LOG:?}"
[ "${1:-}" = "api" ] || { echo "gh stub: only 'api' is supported: $*" >&2; exit 1; }
shift
endpoint="$1"; shift
jqexpr="."
while [ $# -gt 0 ]; do
  case "$1" in
    --jq|-q) jqexpr="$2"; shift 2 ;;
    *) shift ;;
  esac
done
dir="${GH_API_DIR:?}"
case "$endpoint" in
  "repos/${GH_API_REPO:?}/branches/main/protection") f="$dir/main-protection.json" ;;
  "repos/${GH_API_REPO}/branches/develop/protection") f="$dir/develop-protection.json" ;;
  "repos/${GH_API_REPO}/rulesets") f="$dir/rulesets.json" ;;
  "repos/${GH_API_REPO}") f="$dir/repo.json" ;;
  *) echo "gh stub: unknown endpoint $endpoint" >&2; exit 1 ;;
esac
if [ ! -f "$f" ]; then
  echo '{"message":"Branch not protected","status":"404"}'
  echo "gh: Branch not protected (HTTP 404)" >&2
  exit 1
fi
if [ -f "$f.fail" ]; then
  echo "gh: HTTP 500" >&2
  exit 1
fi
jq -r "$jqexpr" "$f"
`

const cvoMainProtection = `{
  "required_status_checks": {"strict": false, "contexts": ["Test (ubuntu-latest)", "Lint", "Build (linux/amd64)", "Analyze (Go) (go)", "Release PR Multi-OS Gate"]},
  "enforce_admins": {"enabled": true},
  "required_pull_request_reviews": {"required_approving_review_count": 0},
  "allow_force_pushes": {"enabled": false},
  "allow_deletions": {"enabled": false}
}`

const cvoRepoState = `{
  "default_branch": "main",
  "allow_merge_commit": true,
  "allow_squash_merge": true,
  "allow_rebase_merge": false,
  "delete_branch_on_merge": true,
  "allow_auto_merge": true
}`

const cvoRulesets = `[{"id": 19583648, "name": "Release tag immutability (v*)", "target": "tag", "enforcement": "active"}]`

var cvoWriteFlagRe = regexp.MustCompile(`(^|\s)(-X|--method|-f|-F|--field|--raw-field|--input)(\s|=|$)`)

type cvoProtection struct {
	t      *testing.T
	dir    string
	stub   string
	log    string
	poison *cvoPoison
}

func cvoNewProtection(t *testing.T) *cvoProtection {
	t.Helper()
	rlsRequireTools(t, "jq")
	dir := t.TempDir()
	p := &cvoProtection{t: t, dir: dir, log: filepath.Join(dir, "gh-api.log"), poison: cvoNewPoison(t)}
	p.stub = rlsWriteScript(t, dir, "gh-stub", cvoGhAPIStub)
	if err := os.WriteFile(p.log, nil, 0o644); err != nil {
		t.Fatal(err)
	}
	p.set("main-protection.json", cvoMainProtection)
	p.set("repo.json", cvoRepoState)
	p.set("rulesets.json", cvoRulesets)
	return p
}

func (p *cvoProtection) set(name, content string) {
	p.t.Helper()
	if err := os.WriteFile(filepath.Join(p.dir, name), []byte(content), 0o644); err != nil {
		p.t.Fatal(err)
	}
}

// edit rewrites a canned file with one literal replacement and fails when the
// literal is absent (a typo in a fixture must not silently test nothing).
func (p *cvoProtection) edit(name, from, to string) {
	p.t.Helper()
	path := filepath.Join(p.dir, name)
	data, err := os.ReadFile(path)
	if err != nil {
		p.t.Fatal(err)
	}
	if !strings.Contains(string(data), from) {
		p.t.Fatalf("fixture %s lacks %q", name, from)
	}
	p.set(name, strings.Replace(string(data), from, to, 1))
}

func (p *cvoProtection) run(args ...string) rlsResult {
	p.t.Helper()
	env := p.poison.env("CUTOVER_GH_CMD="+p.stub, "GH_API_LOG="+p.log, "GH_API_DIR="+p.dir, "GH_API_REPO="+cvoProtectionRepo)
	res := rlsRun(p.t, p.dir, env, "bash", append([]string{cvoScript(p.t, cvoProtectionRel)}, args...)...)
	p.poison.assertUntouched(p.t)
	p.t.Logf("exit=%d\n%s", res.exit, rlsNorm(res.out))
	return res
}

// calls returns the logged gh invocations.
func (p *cvoProtection) calls() []string {
	p.t.Helper()
	data, err := os.ReadFile(p.log)
	if err != nil {
		p.t.Fatal(err)
	}
	var out []string
	for _, l := range strings.Split(strings.TrimSpace(string(data)), "\n") {
		if l != "" {
			out = append(out, l)
		}
	}
	return out
}

// TestCutoverProtectionCompare is AC-GFD-020's protection comparison through the
// stubbed seam: required checks, strict, enforce_admins, merge methods,
// delete_branch_on_merge, default branch, rulesets and develop's protection.
func TestCutoverProtectionCompare(t *testing.T) {
	t.Run("the_observed_baseline_matches_research_section_2", func(t *testing.T) {
		p := cvoNewProtection(t)
		res := p.run()
		if res.exit != 0 {
			t.Errorf("exit code = %d, want 0", res.exit)
		}
		out := rlsNorm(res.out)
		for _, field := range []string{
			"required_checks", "strict", "enforce_admins", "default_branch", "allow_merge_commit",
			"allow_squash_merge", "allow_rebase_merge", "delete_branch_on_merge", "rulesets", "develop_protection",
		} {
			rlsMustContain(t, out, "MATCH "+field)
		}
		if strings.Contains(out, "DRIFT ") {
			t.Errorf("a baseline fixture must show no DRIFT:\n%s", out)
		}
		// The limit of the comparison is part of the output.
		rlsMustContain(t, out, "no change evidence")
	})

	t.Run("every_call_is_a_read_only_api_get_through_the_seam", func(t *testing.T) {
		p := cvoNewProtection(t)
		_ = p.run()
		calls := p.calls()
		if len(calls) == 0 {
			t.Fatal("the seam was never called: the comparison read nothing")
		}
		for _, c := range calls {
			if !strings.HasPrefix(c, "api repos/"+cvoProtectionRepo) {
				t.Errorf("a call outside the four read endpoints: %q", c)
			}
			if cvoWriteFlagRe.MatchString(c) {
				t.Errorf("a call carries a write flag: %q", c)
			}
		}
	})

	t.Run("each_drift_is_named_and_exits_1", func(t *testing.T) {
		rows := []struct {
			name, file, from, to, field string
		}{
			{"a_required_check_is_gone", "main-protection.json", `, "Release PR Multi-OS Gate"`, ``, "required_checks"},
			{"a_required_check_is_added", "main-protection.json", `"Lint",`, `"Lint", "Extra",`, "required_checks"},
			{"strict_turned_on", "main-protection.json", `"strict": false`, `"strict": true`, "strict"},
			{"enforce_admins_turned_off", "main-protection.json", `"enabled": true`, `"enabled": false`, "enforce_admins"},
			{"default_branch_moved", "repo.json", `"default_branch": "main"`, `"default_branch": "develop"`, "default_branch"},
			{"merge_commit_disabled", "repo.json", `"allow_merge_commit": true`, `"allow_merge_commit": false`, "allow_merge_commit"},
			{"squash_disabled", "repo.json", `"allow_squash_merge": true`, `"allow_squash_merge": false`, "allow_squash_merge"},
			{"rebase_allowed", "repo.json", `"allow_rebase_merge": false`, `"allow_rebase_merge": true`, "allow_rebase_merge"},
			{"delete_branch_on_merge_off", "repo.json", `"delete_branch_on_merge": true`, `"delete_branch_on_merge": false`, "delete_branch_on_merge"},
			{"a_second_ruleset_appears", "rulesets.json", `]`, `, {"id": 2, "name": "extra", "target": "branch", "enforcement": "active"}]`, "rulesets"},
		}
		for _, row := range rows {
			t.Run(row.name, func(t *testing.T) {
				p := cvoNewProtection(t)
				p.edit(row.file, row.from, row.to)
				res := p.run()
				if res.exit != 1 {
					t.Errorf("exit code = %d, want 1", res.exit)
				}
				out := rlsNorm(res.out)
				rlsMustContain(t, out, "DRIFT "+row.field)
				// Only the drifted field is flagged.
				if n := strings.Count(out, "DRIFT "); n != 1 {
					t.Errorf("DRIFT lines = %d, want exactly 1:\n%s", n, out)
				}
				rlsMustContain(t, out, "cutover-confirmation.md")
			})
		}
	})

	t.Run("develop_protected_is_a_drift_of_the_baseline", func(t *testing.T) {
		p := cvoNewProtection(t)
		p.set("develop-protection.json", cvoMainProtection)
		res := p.run()
		if res.exit != 1 {
			t.Errorf("exit code = %d, want 1", res.exit)
		}
		rlsMustContain(t, rlsNorm(res.out), "DRIFT develop_protection")
	})

	t.Run("the_post_cutover_target_drops_only_the_multi_os_gate", func(t *testing.T) {
		p := cvoNewProtection(t)
		// Baseline state judged against the post-cutover target: the gate is
		// still required, which is exactly what step 9a is to change.
		res := p.run("--expect", "post-cutover")
		if res.exit != 1 {
			t.Errorf("exit code = %d, want 1 (the Multi-OS gate is still required)", res.exit)
		}
		rlsMustContain(t, rlsNorm(res.out), "DRIFT required_checks")

		// After the operator removes the gate the same comparison is clean, and
		// develop's protection is not judged (the operator's choice, D-13).
		p.edit("main-protection.json", `, "Release PR Multi-OS Gate"`, ``)
		p.set("develop-protection.json", cvoMainProtection)
		res = p.run("--expect", "post-cutover")
		if res.exit != 0 {
			t.Errorf("exit code = %d, want 0\n%s", res.exit, rlsNorm(res.out))
		}
		rlsMustContain(t, rlsNorm(res.out), "MATCH required_checks")
	})

	t.Run("an_unreadable_state_is_exit_2_never_a_match", func(t *testing.T) {
		p := cvoNewProtection(t)
		if err := os.WriteFile(filepath.Join(p.dir, "main-protection.json.fail"), nil, 0o644); err != nil {
			t.Fatal(err)
		}
		res := p.run()
		if res.exit != 2 {
			t.Errorf("exit code = %d, want 2", res.exit)
		}
		out := rlsNorm(res.out)
		rlsMustContain(t, out, "UNREADABLE")
		if strings.Contains(out, "MATCH required_checks") {
			t.Errorf("an unreadable endpoint must not print a MATCH for its fields:\n%s", out)
		}
	})

	t.Run("a_missing_gh_command_is_exit_2", func(t *testing.T) {
		p := cvoNewProtection(t)
		env := p.poison.env("CUTOVER_GH_CMD="+filepath.Join(p.dir, "no-such-gh"), "GH_API_LOG="+p.log, "GH_API_DIR="+p.dir, "GH_API_REPO="+cvoProtectionRepo)
		res := rlsRun(t, p.dir, env, "bash", cvoScript(t, cvoProtectionRel))
		p.poison.assertUntouched(t)
		if res.exit != 2 {
			t.Errorf("exit code = %d, want 2\n%s", res.exit, rlsNorm(res.out))
		}
	})

	t.Run("usage_errors_are_exit_2", func(t *testing.T) {
		p := cvoNewProtection(t)
		if res := p.run("--expect", "nonsense"); res.exit != 2 {
			t.Errorf("unknown --expect exit = %d, want 2", res.exit)
		}
		if res := p.run("--no-such-flag"); res.exit != 2 {
			t.Errorf("unknown flag exit = %d, want 2", res.exit)
		}
	})
}

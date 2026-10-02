# SPEC-AUDIT-MODEL-CONVERGE-001 — acceptance

> Verification layer (0.1.2). Each criterion is Given-When-Then, binary, and names
> the `**Covers**` requirement(s) of spec.md §C. Judgement is by the real output
> of a real command. Go test names are the tests the run phase writes (plan.md
> §E); document items are judged by grep and by the template doc-surface test. 22
> criteria (Tier L ceiling 25). This file has no frontmatter `status:`.
>
> **Command convention.** A RED-now cell is a plain single-invocation read-only
> command recorded in the evidence ledger below with its verbatim stdout and exit
> code (`verification-completeness.md` §2.1) — no pipes, redirection, `&&`, `;`,
> or subshells. The green-path Go commands carry `<SCRUB>` = the one compound
> prefix `unset MOAI_KANBAN MOAI_KANBAN_ID MOAI_KANBAN_LABEL MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED CLAUDE_PROJECT_DIR && `
> (plan.md §C); a prefix that carries `&&` is outside the single-invocation
> form, so a ledger row that needs it is informational and no criterion's RED-now
> rests on it alone. `-run` patterns are anchored (`^…$`): an unanchored pattern
> also selects longer names, and a pattern that selects nothing still prints `ok`.
>
> **Classification.** **release-blocking** = a single-invocation RED-now ledger
> row exists and reproduces on the tree it is pinned to. **regression-guard** =
> no RED-now; GREEN on arrival and must stay GREEN. Release-blocking: 001-007,
> 009-016, 020-022. Regression-guard: 008, 017, 018, 019.
>
> **Tests of record versus mechanisms.** The criteria that grep instruction TEXT
> (AC-ACV-012, -013, -014, -022's document item) are tests-of-record for what the
> text says; they cannot show that a model follows it, and a literal satisfied
> by a comment is the mutant they admit. Where a mechanical backstop exists the
> criterion anchors the behavioural intent to it: the receipt guard (AC-ACV-010),
> the verb's result check (AC-ACV-021) and the verb's unreadable-state output
> (AC-ACV-022).
>
> **Baseline-first ordering (verification-claim-integrity §2.3).** AC-ACV-005's
> "zero non-test readers before, several after" and AC-ACV-008's "byte-identical"
> are claims about ordering. The baselines are: the ledger below, committed in
> the plan-phase commit that carries this file (it precedes every run-phase
> commit), and the golden file of plan.md M1, which lands in its own commit
> before any behaviour change.
>
> **RED-now / green-path pairing.** Each criterion states why it is red on the
> pinned tree and which milestone flips it. A new behaviour's RED-now is the
> absence of the code that would provide it, plus — for AC-ACV-009 — one
> behavioural RED (E22); each named Go test's own RED (verbatim failing output)
> is captured by manager-develop in M2-M4 before GREEN, per
> manager-develop-prompt-template §E8 — those captures are run-phase evidence,
> not ledger rows here.

## Evidence ledger (RED-now observations)

Pin: rows E1-E21 and E23-E28 were run on tree `c50da9c2f` (E1-E21 at 0.1.0) or on
`53a42f013` (E3 re-measured with the new spelling, E22-E28 at 0.1.2); `53a42f013`
is `c50da9c2f` plus this SPEC directory only (`git diff --stat c50da9c2f HEAD --
internal .claude .moai/config` prints nothing), so the code under every row is
the same. The document-level pin `53a42f013` binds every criterion that carries
none of its own. Exit codes were read through a `sh -c '…; echo exit=$?'`
wrapper; the command column is the plain command.

| id | command | verbatim stdout | exit | meaning |
|---|---|---|---|---|
| E1 | `grep -rln --exclude="*_test.go" "ResolveAuditPlan" internal cmd` | (no output) | 1 | no non-test file names the resolver |
| E1c | `grep -rln --exclude="*_test.go" "ValidAuditModels" internal cmd` | `internal/settings/schema_sections.go` · `internal/config/closed_sets.go` | 0 | positive control: the same form hits a symbol that exists |
| E2 | `grep -rn --exclude="*_test.go" "Audit\.Model" internal cmd` | (no output) | 1 | the baseline: ZERO non-test readers of `audit.model` |
| E2c | `grep -rn --exclude="*_test.go" "activeAuditBackend" internal cmd` | `internal/settings/schema_sections.go:367` (comment) · `internal/config/closed_sets.go:87` (comment) · `internal/cli/mcp_audit.go:6` (comment) · `:46` (comment) · `:52` (`func activeAuditBackend(model string) (string, error) {`) | 0 | one definition, four comments, no call — "defined, never read" |
| E3 | `moai verify audit-plan --help` | ` Shared diagnostic snapshot contract. Quality-check results (tests, lint, coverage, ...) recorded under the current working-tree key can be reused by sibling verification layers instead of re-executing, …` (the `verify` GROUP's help) | 0 | red for a specific reason: the verb is unknown, but the group prints its own help and exits 0 — an unknown verb under `verify` is not an error (research.md R-25), so the RED is "the output is the group's help, not a verb usage"; green is a usage naming `moai verify audit-plan` |
| E3c | `moai verify check --help` | `  Query whether a snapshot recorded for the CURRENT working-tree key is fresh: key equality AND recorded-at within the TTL (default 10 minutes, …` | 0 | positive control: a real verb prints its own usage under the same form |
| E4 | `grep -c "moai verify audit-plan" .claude/agents/moai/plan-auditor.md .claude/agents/moai/sync-auditor.md .claude/skills/moai-ref-cross-model-audit/SKILL.md .claude/skills/moai/workflows/sync.md` | `…plan-auditor.md:0` · `…sync-auditor.md:0` · `…SKILL.md:0` · `…sync.md:0` | 1 | no agent, skill or `sync.md` instructs the plan call |
| E4c | `grep -c "audit_multi" .claude/agents/moai/plan-auditor.md .claude/agents/moai/sync-auditor.md .claude/skills/moai-ref-cross-model-audit/SKILL.md` | `…plan-auditor.md:4` · `…sync-auditor.md:5` · `…SKILL.md:6` | 0 | positive control: the same files do carry the tool name, so E4's zeros are not an empty scan |
| E4b | `grep -c "legacy path used" .claude/agents/moai/plan-auditor.md .claude/agents/moai/sync-auditor.md .claude/skills/moai-ref-cross-model-audit/SKILL.md .claude/skills/moai/workflows/sync.md` | `:0` · `:0` · `:0` · `:0` | 1 | the named legacy-path Gap wording is absent (local copies) |
| E4bt | `grep -c "legacy path used" internal/template/templates/.claude/agents/moai/plan-auditor.md internal/template/templates/.claude/agents/moai/sync-auditor.md internal/template/templates/.claude/skills/moai-ref-cross-model-audit/SKILL.md internal/template/templates/.claude/skills/moai/workflows/sync.md` | `:0` · `:0` · `:0` · `:0` | 1 | absent in the template copies too |
| E4d | `grep -cF "Single-backend audit mode (per the project" .claude/agents/moai/plan-auditor.md .claude/agents/moai/sync-auditor.md internal/template/templates/.claude/agents/moai/plan-auditor.md internal/template/templates/.claude/agents/moai/sync-auditor.md` | `:1` · `:1` · `:1` · `:1` | 0 | the prose-interpretation block exists in all four agent copies |
| E4t | `grep -c "moai verify audit-plan" internal/template/templates/.claude/agents/moai/plan-auditor.md internal/template/templates/.claude/agents/moai/sync-auditor.md internal/template/templates/.claude/skills/moai-ref-cross-model-audit/SKILL.md internal/template/templates/.claude/skills/moai/workflows/sync.md` | `:0` · `:0` · `:0` · `:0` | 1 | the template copies carry no plan call either |
| E4x | `grep -c "moai verify audit-plan" internal/template/templates/.codex/agents/moai/plan-auditor.toml internal/template/templates/.codex/agents/moai/sync-auditor.toml` | `:0` · `:0` | 1 | the emitted `.codex` roles carry none (they must follow the `.md` after `make agents-emit`) |
| E5 | `grep -rn "multiConvergenceImplemented" --include="*.go" internal` | `internal/cli/mcp_audit_test.go:40` · `:41` · `internal/cli/mcp_audit.go:27` · `:31` · `:49` | 0 | the stale sentinel and a test pinning it |
| E6 | `grep -rn --exclude="*_test.go" "AP-8" internal` | `internal/config/audit_models.go:28` · `internal/cli/mcp_audit.go:8` · `:27` | 0 | the deferral markers in non-test source |
| E7 | `grep -n -A1 "model: claude-opus-5-5" .moai/config/sections/workflow.yaml` | `24:            model: claude-opus-5-5` · `25-            effort: medium` | 0 | the committed claude pin is `medium` (a bare `effort: medium` grep is non-specific — the file has sixteen other hits) |
| E8 | `grep -nE "^        model: multi$" .moai/config/sections/workflow.yaml` | (no output) | 1 | the committed yaml has no `audit.model` |
| E9 | `grep -n "claude-opus-5-5/medium" .moai/config/sections/workflow.yaml` | `20:    # z.ai states low\|high\|max. Claude defaults to claude-opus-5-5/medium and` | 0 | the header comment states `medium` |
| E10 | `grep -n "effort: high" internal/template/templates/.moai/config/sections/workflow.yaml` | `136:            effort: high` · `139:            effort: high` | 0 | the distributed template's claude pin is already `high`; no template change is needed |
| E12 | `grep -n "const sentinel" internal/web/mcp_audit_surface_test.go` | `58:	const sentinel = "activeAuditBackend" // the M3 resolver symbol in internal/cli` | 0 | the web guard's sentinel names the symbol M5 deletes |
| E13 | `grep -c "internal/config" internal/auditreceipt/store.go` | `0` | 1 | the receipt store does not import the config package |
| E14 | `grep -c "cross_model_required" .claude/skills/moai/workflows/sync.md internal/template/templates/.claude/skills/moai/workflows/sync.md` | `:0` · `:0` | 1 | neither `sync.md` copy mentions the cross-model binding condition |
| E15 | `grep -c "^package config" internal/config/audit_plan.go` | (stderr) `grep: internal/config/audit_plan.go: No such file or directory` | 2 | the resolver file does not exist |
| E15c | `grep -c "^package config" internal/config/audit_models.go` | `1` | 0 | positive control: the same form hits an existing file of the package |
| E16 | `grep -c "model: multi" internal/template/templates/.moai/config/sections/workflow.yaml` | `0` | 1 | baseline for the "template unchanged" guard |
| E17 | `grep -rn "DefaultCodexAuditLegTimeout" internal` | (no output) | 1 | no codex-leg limit exists |
| E17c | `grep -n "DefaultCodexAuditTimeout" internal/config/defaults.go` | `589:// DefaultCodexAuditTimeout bounds ONE read-only audit process started by the` · `593:var DefaultCodexAuditTimeout = 20 * time.Minute` | 0 | positive control: the sibling the new limit is derived from exists, as a `var` |
| E18 | `grep -n "context.WithTimeout" internal/cli/mcp_convergence.go` | (no output) | 1 | `performCodexAudit` and the fan-out apply no deadline to the codex leg |
| E18c | `grep -n "context.WithTimeout" internal/cli/mcp_claude.go` | `110:	auditCtx, cancel := context.WithTimeout(ctx, claudeAuditTimeout)` | 0 | positive control: the same form hits where a deadline exists |
| E19 | `grep -c "audit_multi unreachable" .claude/skills/moai/workflows/sync.md internal/template/templates/.claude/skills/moai/workflows/sync.md` | `:0` · `:0` | 1 | neither `sync.md` copy records the unavailable-`audit_multi` outcome |
| E20 | `grep -c "audit_multi" .claude/workflows/sync-audit-4dim.js internal/template/templates/.claude/workflows/sync-audit-4dim.js` | `:0` · `:0` | 1 | neither script copy says the verdict is the Claude input to an `audit_multi` call |
| E21 | `grep -c "mcp__moai__" .claude/workflows/sync-audit-4dim.js internal/template/templates/.claude/workflows/sync-audit-4dim.js` | `:0` · `:0` | 1 | baseline for "the script never calls an MCP tool" — it must stay 0 |
| E22 | `go test -overlay <SCRATCH>/overlay.json -count=1 -v -run '^TestRedACVModelMultiCodexInconclusive$' ./internal/cli/` (source in the ledger entry E22-src; `<SCRATCH>` = a machine-local scratch directory holding `zz_red_acv_test.go` and the overlay JSON) | `=== RUN   TestRedACVModelMultiCodexInconclusive` · `zz_red_acv_test.go:39: overall_verdict=pass gate_unmet="" fail_open=[codex]` · `zz_red_acv_test.go:41: RED: model: multi with codex inconclusive returned overall_verdict=pass gate_unmet="" (want fail / codex)` · `--- FAIL: TestRedACVModelMultiCodexInconclusive (0.01s)` · `FAIL	github.com/modu-ai/moai-adk/internal/cli	1.087s` | 1 | the BEHAVIOURAL RED for AC-ACV-009 and AC-ACV-021: a `model: multi` yaml with codex `inconclusive` currently returns pass with an empty `gate_unmet` — also the exact shape an older server returns. Tree `53a42f013`. The overlay supplies a throwaway test file that is not in the tree. |
| E23 | `grep -c "check-session" .claude/skills/moai/workflows/sync.md internal/template/templates/.claude/skills/moai/workflows/sync.md` | `:0` · `:0` | 1 | neither `sync.md` copy mentions the result check |
| E24 | `grep -n "NewDefaultWorkflowConfig" internal/auditreceipt/store.go internal/cli/mcp_worktree_root.go internal/cli/mcp_audit_multi.go internal/cli/audit_pin.go` | (no output) | 1 | baseline for the raw-read guard: the existing raw readers never construct the default config |
| E24c | `grep -rln --exclude="*_test.go" "NewDefaultWorkflowConfig" internal` | `internal/config/types.go` · `internal/config/defaults.go` · `internal/config/closed_sets.go` · `internal/config/audit_models.go` · `internal/cli/mcp_audit_multi_record.go` · `internal/cli/contract.go` | 0 | positive control: the same form hits files that do use the default constructor |
| E25 | `grep -n "loadConvergenceResult(" internal/cli/multi_review_gate.go` | `120:func loadConvergenceResult(projectDir, sessionID string) (ConvergenceResult, bool) {` | 0 | the persisted-result reader the verb's check reuses exists |
| E28 | `grep -c "config_status: unreadable" .claude/agents/moai/plan-auditor.md .claude/agents/moai/sync-auditor.md .claude/skills/moai-ref-cross-model-audit/SKILL.md .claude/skills/moai/workflows/sync.md` | `:0` · `:0` · `:0` · `:0` | 1 | the unreadable-state Gap wording is absent |

**Ledger entry E22-src** — the throwaway test the overlay supplies (not part of
the tree; the overlay JSON is `{"Replace": {"<tree>/internal/cli/zz_red_acv_test.go":
"<SCRATCH>/zz_red_acv_test.go"}}`):

```go
package cli

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
)

func TestRedACVModelMultiCodexInconclusive(t *testing.T) {
	dir := t.TempDir()
	sec := filepath.Join(dir, ".moai", "config", "sections")
	if err := os.MkdirAll(sec, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(sec, "workflow.yaml"), []byte("workflow:\n  audit:\n    model: multi\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("CLAUDE_PROJECT_DIR", dir)
	t.Setenv(config.EnvMoaiLaunchProvider, BackendClaude)
	rc := &recordingCallerMulti{verdictBy: map[string]ReviewOutput{
		BackendCodex: {Verdict: VerdictInconclusive, Summary: "codex missing", Findings: []Finding{}, NextSteps: []string{}},
	}}
	orig := backendCall
	backendCall = rc.call
	t.Cleanup(func() { backendCall = orig })
	claude := map[string]any{"verdict": "pass", "summary": "ok", "findings": []any{}, "next_steps": []any{}}
	res, err := callToolAuditMulti(t, claude, map[string]any{"project_root": dir})
	if err != nil || res == nil {
		t.Fatalf("handler: %v", err)
	}
	var out ConvergenceResult
	if err := json.Unmarshal([]byte(toolResultText(res)), &out); err != nil {
		t.Fatalf("decode: %v", err)
	}
	t.Logf("overall_verdict=%s gate_unmet=%q fail_open=%v", out.OverallVerdict, out.GateUnmet, out.FailOpenBackends)
	if out.OverallVerdict != "fail" || out.GateUnmet != "codex" {
		t.Errorf("RED: model: multi with codex inconclusive returned overall_verdict=%s gate_unmet=%q (want fail / codex)", out.OverallVerdict, out.GateUnmet)
	}
}
```

## AC-ACV-001 — every token maps to a total plan

**Covers**: maps REQ-ACV-001

- **Classification**: release-blocking (RED-now: E1, positive control E1c; E15).
- **Given** the closed set `claude`, `codex`, `glm`, `multi` and the empty token,
  **When** each is resolved with no configured gates and no caller gates,
  **Then** each plan assigns each of `claude`, `codex`, `glm` exactly one of
  `off`, `advisory`, `required`, and the five rows equal design.md §D.1: empty
  and `multi`: required / required / advisory (the empty token's entries carry
  source `default`, `multi`'s carry `config.model`); `claude`: required with
  source `config.model` and explicit, codex required and glm advisory with source
  `default` and not explicit; `codex`: off / required / off and `glm`: off / off /
  required, all `config.model`.
- **RED-now**: E1/E15 — there is no resolver.
- **green** (M2): `<SCRUB>go test -count=1 -v -run '^TestResolveAuditPlan_TokenTable$' ./internal/config/` →
  `--- PASS: TestResolveAuditPlan_TokenTable ` (name followed by a space) and the
  five sub-test names, exit 0.
- **Mutant probe**: a resolver that returns only the `multi` row for every token
  fails the table; one that maps `claude` to Claude alone (codex and glm `off`)
  fails the `claude` row; one that leaves a backend unset in any row fails the
  exactly-one-gate assertion.

## AC-ACV-002 — precedence, with the deciding source recorded

**Covers**: maps REQ-ACV-002

- **Classification**: release-blocking (RED-now: E1, E15).
- **Given** `audit.model: multi`, `audit.gates.glm: off`, and a caller gate
  `codex: advisory`, **When** the plan is resolved, **Then** codex is `advisory`
  (source `argument`), glm is `off` (source `config.gates`), claude is `required`
  (source `config.model`); and with the caller gate removed codex returns to
  `required` (source `config.model`); and with an empty token and no gates every
  entry has source `default`.
- **RED-now**: E1/E15.
- **green** (M2): `<SCRUB>go test -count=1 -v -run '^TestResolveAuditPlan_Precedence$' ./internal/config/` →
  `--- PASS: TestResolveAuditPlan_Precedence `, exit 0.
- **Mutant probe**: swapping argument and config, or config.gates and
  config.model, changes at least one asserted gate; dropping the source field
  fails the source assertions.

## AC-ACV-003 — an invalid token or gate names itself and is never defaulted

**Covers**: maps REQ-ACV-003

- **Classification**: release-blocking (RED-now: E1, E2).
- **Given** a tree whose `audit.model` is `grok` (and, separately, whose
  `audit.gates.codex` is `requird`), **When** the resolver, `audit_multi`, or
  `moai verify audit-plan` reads it, **Then** the resolver returns an error whose
  text is `audit_model "grok" unknown (want one of claude|codex|glm|multi)` (for
  the gate: the key, the value, and `off|advisory|required`); `audit_multi`
  returns a tool error (`isError: true`) carrying that text and invokes no
  backend; the verb exits 1 with the text, prefixed `audit-plan:`, on stderr and
  nothing on stdout. Token matching is exact and case-sensitive, with surrounding
  whitespace trimmed (`Multi` is rejected; ` multi ` is accepted).
- **RED-now**: E1/E2 — nothing reads the token, so nothing can reject it.
- **green** (M2, M3, M4): `<SCRUB>go test -count=1 -v -run '^(TestResolveAuditPlan_RejectsUnknown|TestAuditMulti_UnknownConfiguredTokenIsToolError|TestAuditPlanCmd_InvalidConfigExitsOne)$' ./internal/config/ ./internal/cli/` →
  three `--- PASS:` lines, exit 0.
- **Mutant probe**: a resolver that maps an unknown token to the default plan
  fails all three; a handler that swallows the error and fans out fails the
  `audit_multi` test's no-backend assertion.

## AC-ACV-004 — the resolver is pure, in `internal/config`, and never fed from a merged config

**Covers**: maps REQ-ACV-004

- **Classification**: release-blocking (RED-now: E15, positive control E15c; the
  raw-read baseline E24 with control E24c).
- **Given** the resolver source file and the files that feed it, **When** their
  imports and references are scanned, **Then** `internal/config/audit_plan.go`
  exists, declares `package config`, and imports none of `os`, `time`, `net`,
  `io`, `path/filepath`, `os/exec`; and none of `internal/config/audit_plan.go`,
  `internal/cli/audit_plan_cmd.go`, `internal/cli/mcp_audit_multi.go`,
  `internal/cli/mcp_worktree_root.go`, `internal/cli/audit_pin.go` or
  `internal/auditreceipt/store.go` references `NewDefaultWorkflowConfig` or
  `NewDefaultConfig`.
- **RED-now**: E15 — the file does not exist (exit 2), E15c shows the control
  form works on an existing file of the package; E24 records that the existing
  raw readers already satisfy the second clause (E24c shows the form finds
  default-constructor users elsewhere).
- **green** (M2-M4): `grep -c "^package config" internal/config/audit_plan.go` →
  `1`, exit 0; `grep -nE "\"(os|time|net|io|path/filepath|os/exec)\"" internal/config/audit_plan.go` →
  (no output), exit 1; `grep -n "NewDefaultWorkflowConfig\|NewDefaultConfig" internal/config/audit_plan.go internal/cli/audit_plan_cmd.go internal/cli/mcp_audit_multi.go internal/cli/mcp_worktree_root.go internal/cli/audit_pin.go internal/auditreceipt/store.go` →
  (no output), exit 1; and the behavioural pair of AC-ACV-008 (pins-only) and
  AC-ACV-011 (explicit `claude`) passes.
- **Mutant probe**: a resolver that reads `os.Getenv` or the yaml file itself
  fails the second command; a consumer that builds the input from the default
  constructor fails the third command and the pins-only golden (every project
  would read `claude`).

## AC-ACV-005 — the resolver is consumed by non-test callers

**Covers**: maps REQ-ACV-005

- **Classification**: release-blocking (RED-now: E1, E2, E2c; baseline-first —
  the ledger commit precedes M2).
- **Given** the run-phase tree, **When** the non-test files that reference the
  resolver are listed, **Then** the list names the definition file and at least
  four consumers: `internal/config/audit_plan.go`,
  `internal/cli/mcp_audit_multi.go`, `internal/cli/mcp_worktree_root.go`,
  `internal/cli/audit_plan_cmd.go`, `internal/auditreceipt/store.go` — files from
  all three packages. Before the change (E1, E2) there were zero such files and
  zero non-test readers of `audit.model`.
- **RED-now**: E1 (no reference), E2 (no reader of the token), E2c (the old
  validator was defined and never called).
- **green** (M2-M4): `grep -rl --exclude="*_test.go" "ResolveAuditPlan(" internal` →
  stdout containing those five paths, exit 0; and AC-ACV-006's handler-level
  test passes (a reference without behaviour does not satisfy this criterion).
- **Mutant probe**: calling the resolver only from the verb leaves the
  `mcp_audit_multi.go` and `store.go` paths out of the listing and fails
  AC-ACV-006 and AC-ACV-010.

## AC-ACV-006 — `audit_multi` follows the configured plan when no gates are passed

**Covers**: maps REQ-ACV-006

- **Classification**: release-blocking (RED-now: E1, E2).
- **Given** a temp project root whose `workflow.yaml` sets `audit.model` to each
  of `codex`, `glm`, `claude`, `multi` in turn, and a recording `backendCall`
  stub, **When** `audit_multi` is called with no `gates` argument (and, in one
  row, with a `gates` object omitting the codex key), **Then** the backends the
  stub records are exactly the non-`off` entries of that token's §D.1 row
  (`codex` → [codex]; `glm` → [glm]; `claude` → [claude, codex, glm] — Claude
  explicit, the rest the default, so the fan-out is today's; `multi` → [claude,
  codex, glm]), the persisted and returned result carries `plan_source: "config"`
  in every row, and the omitted-key row takes codex from the plan.
- **RED-now**: E1/E2 — with no reader of the token, the `codex` and `glm` rows
  would fan out to the default three and no row would carry `plan_source`.
- **green** (M3): `<SCRUB>go test -count=1 -v -run '^TestAuditMulti_ConfigPlanFallback$' ./internal/cli/` →
  `--- PASS: TestAuditMulti_ConfigPlanFallback `, exit 0.
- **Mutant probe**: a handler that ignores config invokes three backends for the
  `codex` row and fails it; a `claude` row that mapped to Claude alone invokes one
  backend and fails it.

## AC-ACV-007 — explicit call arguments win over configuration

**Covers**: maps REQ-ACV-006

- **Classification**: release-blocking (RED-now: E1, E2).
- **Given** `audit.model: multi` and a call whose `gates` is
  `{codex: off, glm: off, claude: required}`, **When** `audit_multi` runs,
  **Then** codex and glm are not invoked and the result carries no codex or glm
  entry in `per_backend_verdicts`; and given `audit.gates.codex: off` with
  `model: multi` and a call that supplies `codex: required`, codex is invoked.
- **RED-now**: E1/E2.
- **green** (M3): `<SCRUB>go test -count=1 -v -run '^TestAuditMulti_ArgsWinOverConfig$' ./internal/cli/` →
  `--- PASS: TestAuditMulti_ArgsWinOverConfig `, exit 0.
- **Mutant probe**: config-over-argument precedence fails the first sub-case.

## AC-ACV-008 — no configuration and no arguments, or pins only: byte-identical to before

**Covers**: maps REQ-ACV-004, REQ-ACV-007

- **Classification**: regression-guard (GREEN on the pre-change tree; the golden
  lands in M1's own commit before any behaviour change).
- **Given** two temp project roots — one with no `workflow.yaml`, one whose
  `workflow.yaml` has an `audit` block carrying ONLY the three backend pins (the
  shape of `internal/template/templates/.moai/config/sections/workflow.yaml`
  lines 119-139) — stubbed backends, and a call with no `gates`, **When** the
  marshalled `ConvergenceResult` (with `BuildCommit` and `BuildLag` zeroed) is
  compared with `internal/cli/testdata/audit_multi_default.golden.json`, **Then**
  they are byte-equal for an all-pass case and for a codex-inconclusive case on
  both roots (the result stays a pass — fail-open — and carries no `plan_source`,
  no `gate_unmet`); and the plan resolved for each root is the default plan with
  every entry `explicit: false`.
- **RED-now**: none by design (a guard). It is red against any change that adds
  a member, flips the codex-inconclusive case to a fail, changes the default
  gates, or reads a default-merged configuration (which would give the pins-only
  root `model: claude`).
- **green** (M1 on arrival; M2-M7 keep it): `<SCRUB>go test -count=1 -v -run '^(TestAuditMulti_NoConfigNoArgs_ByteIdentical|TestAuditMulti_PinsOnlyConfig_ByteIdentical)$' ./internal/cli/` →
  two `--- PASS:` lines, exit 0.
- **Mutant probe**: always emitting `plan_source`, treating the default codex
  `required` as explicit, or building the resolver input from the default
  constructor fails the golden on the pins-only root.

## AC-ACV-009 — a required backend that cannot answer fails the gate by name

**Covers**: maps REQ-ACV-008

- **Classification**: release-blocking (RED-now: the behavioural RED E22, and E2).
- **Given** a temp project root whose `workflow.yaml` sets ONLY
  `audit.model: multi` (no `audit.gates` key), claude and glm stubs returning
  `pass`, and a codex stub returning `inconclusive` (one row each for binary
  missing, key missing, quota refusal, timeout, unauthenticated, malformed — all
  `inconclusive` at the seam), **When** `audit_multi` runs, **Then** each row's
  result has `overall_verdict: "fail"`, `gate_unmet: "codex"`, a
  `residual_risk_note` containing `codex`, the codex `per_backend_verdicts` entry
  still `inconclusive` and listed in `fail_open_backends`, and no pass built from
  claude and glm alone; and with the same stubs and NO configuration the result
  stays a pass (AC-ACV-008's case).
- **RED-now**: E22 — measured on tree `53a42f013`: with exactly this fixture the
  handler returns `overall_verdict=pass gate_unmet="" fail_open=[codex]` (exit 1
  from the throwaway test that asserts the green behaviour); E2 gives the reason
  (no reader of the token).
- **green** (M3): `<SCRUB>go test -count=1 -v -run '^TestAuditMulti_ModelMulti_CodexUnavailable_FailsClosedNamed$' ./internal/cli/` →
  `--- PASS: TestAuditMulti_ModelMulti_CodexUnavailable_FailsClosedNamed ` and its
  six sub-test names, exit 0.
- **Mutant probe**: enforcement keyed only on a written `audit.gates` block
  (today's behaviour) leaves the pass in place and fails the test; a fix that
  downgrades to a claude-only pass fails it.

## AC-ACV-010 — one reading: receipts and the single-backend gate follow the token

**Covers**: maps REQ-ACV-005, REQ-ACV-009

- **Classification**: release-blocking (RED-now: E13 — the receipt store does not
  import the resolver; E2).
- **Given** trees whose `workflow.yaml` sets `audit.model` to `multi`, `codex`,
  `claude`, `glm`, empty, and `multi` with `audit.gates.codex: advisory`,
  **When** the receipt predicate `CodexGateRequired` is evaluated, **Then** it is
  true for `multi` and `codex`, false for `claude` (codex is the non-explicit
  default there), `glm`, empty, and the `advisory` override; and given
  `model: multi` with codex `inconclusive`, the `audit_multi` result carries a
  non-empty `audit_receipt` even though `gate_unmet` is `codex`, and
  `codex_audit` with the same missing codex returns `verdict: fail` with a
  non-empty `gate_unmet`. This is the mechanical backstop for a Claude-only PASS:
  the SubagentStop guard refuses an auditor PASS the receipt store cannot
  corroborate.
- **RED-now**: E13 (store imports no config), E2.
- **green** (M3): `<SCRUB>go test -count=1 -v -run '^(TestCodexGateRequired_FromModelToken|TestAuditMulti_ModelMultiRecordsReceiptOnUnmetGate|TestCodexAudit_ModelCodexUnmetGateFails)$' ./internal/auditreceipt/ ./internal/cli/` →
  three `--- PASS:` lines, exit 0; and `grep -c "internal/config" internal/auditreceipt/store.go` →
  a count of at least `1`, exit 0.
- **Mutant probe**: leaving `rawCodexGate` on the raw `audit.gates` key makes the
  `multi` row false and fails the first test.

## AC-ACV-011 — `moai verify audit-plan` prints the plan and does nothing else

**Covers**: maps REQ-ACV-010, REQ-ACV-011, REQ-ACV-012

- **Classification**: release-blocking (RED-now: E3).
- **Given** fixture trees (no yaml; a pins-only `audit` block; `model: multi`;
  an explicit `model: claude`; `model: codex` with `gates.glm: advisory`;
  `model: grok`; a config-orphaned worktree), **When**
  `moai verify audit-plan --project-root <tree>` runs in each, **Then** it prints
  JSON with the members of design.md §D.5 and exits 0 — no yaml or pins only:
  `config_status: "absent"`, `model_source: "default"`, every `explicit: false`;
  the explicit `claude` fixture: claude `required` source `config.model`
  `explicit: true`, codex `required` source `default` `explicit: false`, glm
  `advisory` source `default`, `cross_model_active: false`, `enforced_required:
  ["claude"]`; invalid configuration prints nothing on stdout, the REQ-ACV-003
  text prefixed `audit-plan:` on stderr, exit 1; and across every run the tree's
  `.moai/state` directory is byte-identical before and after, no receipt file
  exists, and the `backendCall` seam recorded zero calls. Given `CLAUDE_PROJECT_DIR`
  set to one tree and `--project-root` naming another, the named tree is read;
  with no flag the `CLAUDE_PROJECT_DIR` tree is read (the trap of design.md §D.4).
  The output contract holds: stdout is one JSON object carrying `config_status`,
  or stderr carries an `audit-plan:` line — nothing else.
- **RED-now**: E3 — the output is the `verify` group's help (exit 0), not a verb
  usage or a JSON plan.
- **green** (M4): `<SCRUB>go test -count=1 -v -run '^(TestAuditPlanCmd_PrintsPlan|TestAuditPlanCmd_PinsOnlyIsDefault|TestAuditPlanCmd_InvalidConfigExitsOne|TestAuditPlanCmd_WritesNothing|TestAuditPlanCmd_ProjectRoot|TestAuditPlanCmd_OutputContract|TestAuditPlanCmd_NoAskUserQuestion)$' ./internal/cli/` →
  seven `--- PASS:` lines, exit 0; and `moai verify audit-plan --help` on the built
  binary → a usage naming `moai verify audit-plan`, exit 0 (compare E3: the group
  help).
- **Mutant probe**: a verb that calls `audit_multi` or writes a receipt fails
  `WritesNothing`; one that prints a plan for an invalid token fails the exit-1
  assertion; one that ignores `--project-root` in favour of the environment fails
  `ProjectRoot`; one that reads a default-merged config prints `model: claude`
  for the pins-only tree and fails `PinsOnlyIsDefault`.

## AC-ACV-012 — the auditors follow the measured plan, not prose

**Covers**: maps REQ-ACV-013

- **Classification**: release-blocking (RED-now: E4, E4t, E4x, E4d; positive
  control E4c). A test-of-record for instruction TEXT; the mechanical backstops
  are AC-ACV-010 and AC-ACV-021.
- **Given** the four agent copies (`plan-auditor.md`, `sync-auditor.md`, local
  and template), the emitted `.codex` pair, and the cross-model skill (local and
  template), **When** they are scanned after M7, **Then** each `.md` agent and
  each skill copy contains `moai verify audit-plan` and `--check-session`; the
  old block heading `Single-backend audit mode (per the project` is absent from
  all four agent copies (the token-keyed prose survives only inside the
  paragraph labelled as the legacy path, AC-ACV-013); the two template `.codex`
  role TOMLs, regenerated by `make agents-emit`, contain `moai verify
  audit-plan`; the local `sync-auditor.md` still carries its
  `<!-- moai:closure-second-review:start -->` and `:end -->` markers; and
  `TestClaudeAuditTemplateSurfacesAndCatalogHash` and the new doc-surface test
  pass (identical audit section between local and template agents; catalogue hash
  current).
- **RED-now**: E4, E4t, E4x (zero), E4d (one each).
- **green** (M7): `grep -c "moai verify audit-plan" .claude/agents/moai/plan-auditor.md …` →
  every count at least `1`, exit 0; `grep -cF "Single-backend audit mode (per the project" …` →
  every count `0`, exit 1; `grep -c "moai:closure-second-review" .claude/agents/moai/sync-auditor.md` →
  `2`, exit 0; and `<SCRUB>go test -count=1 -v -run '^(TestClaudeAuditTemplateSurfacesAndCatalogHash|TestAuditPlanDocSurface|TestTemplateNoInternalContentLeak)$' ./internal/template/` →
  three `--- PASS:` lines, exit 0 (the leak test proves the added template text
  carries no internal SPEC or card identifier).
- **Mutant probe**: editing only the local copies fails the parity and
  doc-surface tests; editing the `.md` without `make agents-emit` leaves E4x at
  zero and fails the TOML check; the literal placed in a comment satisfies the
  grep and is NOT caught here — that gap is why the behavioural intent rests on
  AC-ACV-010 and AC-ACV-021.

## AC-ACV-013 — an absent verb is the legacy path with a named Gap; a failing required backend is not

**Covers**: maps REQ-ACV-014

- **Classification**: release-blocking (RED-now: E4b, E4bt). A test-of-record for
  instruction text, with a mechanical part (the verb's output contract,
  AC-ACV-011).
- **Given** the two auditor agents, the cross-model skill and `sync.md` (local and
  template) after M7, **When** scanned, **Then** each contains the literal
  `plan surface unreachable, legacy path used`, the labelled legacy-path
  paragraph, the statement that the verb is detected by its output contract (not
  by exit code, because an unknown verb under `verify` exits 0), and the
  distinction `a configured required backend that does not answer stays
  fail-closed`; and the legacy-path paragraph preserves the pre-change behaviour —
  the auditor reads the project's `audit_model` as the pre-change text said and
  does not treat the missing verb as a failure.
- **RED-now**: E4b (local) and E4bt (template) — the literal is absent in all
  eight files; E3 shows why the exit code cannot be the detector.
- **green** (M7): `grep -c "plan surface unreachable, legacy path used" .claude/agents/moai/plan-auditor.md .claude/agents/moai/sync-auditor.md .claude/skills/moai-ref-cross-model-audit/SKILL.md .claude/skills/moai/workflows/sync.md` →
  every count at least `1`, exit 0; and `TestAuditPlanDocSurface` (AC-ACV-012's
  run) asserts the same literals in the template copies.
- **Mutant probe**: wording the Gap without the fixed literal fails the grep; an
  instruction that makes the missing verb a PASS-blocking failure contradicts the
  "legacy path" paragraph and is a review finding the grep cannot see — the
  mechanical protection against a Claude-only PASS in the configured case is
  AC-ACV-010 (receipts) and AC-ACV-021 (result check), which apply whenever the
  verb IS reached.

## AC-ACV-014 — the sync verdict is binding only with a passing `audit_multi` and a good check

**Covers**: maps REQ-ACV-015, REQ-ACV-021, REQ-ACV-024

- **Classification**: release-blocking (RED-now: E14, E19, E20, E23; baseline
  E21). A test-of-record for instruction text; the machine-readable parts it
  relies on are the persisted result and the verb's check (AC-ACV-021).
- **Given** both copies of `workflows/sync.md` and of `sync-audit-4dim.js` after
  M7, **When** they are scanned, **Then** each `sync.md` copy names
  `cross_model_required`, `moai verify audit-plan`, `--check-session`, the
  post-PASS `audit_multi` call, the "both must pass" rule, the `cross_model:` /
  `audit_multi:` / `gate_unmet:` / `audit_receipt:` / `plan_check:` lines of the
  binding statement, `binding: no`, and the literal `audit_multi unreachable`;
  each script copy names `audit_multi` in its header comment and
  `meta.description` as the orchestrator's call, and still contains no
  `mcp__moai__` (E21 stays 0); the script's phases, schemas and verdict
  computation are unchanged; the Go `FourDimVerdict` / `IsBinding` source is
  untouched; and the template copies carry no internal SPEC or card identifier.
- **RED-now**: E14, E19, E23 (zero in both `sync.md` copies), E20 (zero in both
  script copies).
- **green** (M7): `grep -c "cross_model_required" …`,
  `grep -c "audit_multi unreachable" …` and `grep -c "check-session" …` over both
  `sync.md` copies → each at least `1`, exit 0; `grep -c "audit_multi" .claude/workflows/sync-audit-4dim.js internal/template/templates/.claude/workflows/sync-audit-4dim.js` →
  each at least `1`, exit 0; `grep -c "mcp__moai__" …` over both script copies →
  `0`, exit 1; and `<SCRUB>go test -count=1 -v -run '^(TestAuditPlanDocSurface|TestTemplateNoInternalContentLeak)$' ./internal/template/` →
  two `--- PASS:` lines, exit 0.
- **Mutant probe**: moving the call into the script adds `mcp__moai__` and fails
  the zero-count grep; editing only the local copies fails the doc-surface test;
  "codex absent → always PASS" with the required phrases placed in a comment
  keeps every count at one and is NOT caught by this criterion — the intent is
  anchored to AC-ACV-021, which gives the orchestrator a mechanical check to run
  and the check a failing fixture.

## AC-ACV-015 — the stale deferral is gone and its guard is not vacuous

**Covers**: maps REQ-ACV-016

- **Classification**: release-blocking (RED-now: E5, E6, E12).
- **Given** the run-phase tree, **When** it is scanned, **Then**
  `multiConvergenceImplemented` appears nowhere in Go source or tests; the string
  `AP-8` appears in no non-test Go file; the web guard's sentinel is
  `ResolveAuditPlan` and the guard fails when the sentinel is absent from the
  non-test source of `internal/config`.
- **RED-now**: E5 (five hits), E6 (three hits), E12 (sentinel is
  `activeAuditBackend`).
- **green** (M5): `grep -rn "multiConvergenceImplemented" --include="*.go" internal` →
  (no output), exit 1; `grep -rn --exclude="*_test.go" "AP-8" internal` → (no
  output), exit 1; `grep -n "const sentinel" internal/web/mcp_audit_surface_test.go` →
  a line naming `ResolveAuditPlan`, exit 0; and `<SCRUB>go test -count=1 -v -run '^(TestWebConsole_AuditNoForkedInterpreter|TestWebConsole_AuditSentinelExists)$' ./internal/web/` →
  two `--- PASS:` lines, exit 0.
- **Mutant probe**: deleting `activeAuditBackend` without retargeting the
  sentinel leaves the guard green and vacuous — the positive-control test
  `TestWebConsole_AuditSentinelExists` is the part that fails when the sentinel
  names no live symbol.

## AC-ACV-016 — the committed yaml opts in and the pin is aligned; the template is untouched

**Covers**: maps REQ-ACV-017

- **Classification**: release-blocking (RED-now: E7, E8, E9; template guard E10,
  E16).
- **Given** the run-phase tree, **When** the committed
  `.moai/config/sections/workflow.yaml` is read, **Then** the line after
  `model: claude-opus-5-5` is `effort: high`, the key `model: multi` sits under
  `audit:`, the header comment no longer says `claude-opus-5-5/medium`; the
  committed file loads through the config loader to `Audit.Model == "multi"` and
  `Audit.Claude.Effort == DefaultClaudeAuditEffort`; the template yaml still has
  no `model: multi` and its claude pin is still `high`; the Go default
  `Audit.Model` is still `claude`.
- **RED-now**: E7 (`medium`), E8 (no key), E9 (comment says medium).
- **green** (M6): `grep -n -A1 "model: claude-opus-5-5" .moai/config/sections/workflow.yaml` →
  `…effort: high`, exit 0; `grep -nE "^        model: multi$" .moai/config/sections/workflow.yaml` →
  one line, exit 0; `grep -n "claude-opus-5-5/medium" .moai/config/sections/workflow.yaml` →
  (no output), exit 1; `grep -c "model: multi" internal/template/templates/.moai/config/sections/workflow.yaml` →
  `0`, exit 1 (E16 unchanged); and
  `<SCRUB>go test -count=1 -v -run '^(TestCommittedWorkflowYamlAuditAlignment|TestAuditConfig_DefaultProfile)$' ./internal/config/` →
  two `--- PASS:` lines, exit 0.
- **Mutant probe**: setting only the pin and not the token fails the
  `grep -nE` and the loader test; adding `model: multi` to the template fails
  the `grep -c` guard.

## AC-ACV-017 — the fan-out stays concurrent under a config plan

**Covers**: maps REQ-ACV-018

- **Classification**: regression-guard (the default plan already activates all
  three legs concurrently, so the test is GREEN against the pre-change tree and
  must stay GREEN).
- **Given** `audit.model: multi` and three stub backends that each sleep a fixed
  interval, **When** `audit_multi` runs, **Then** all three are invoked, the
  elapsed time is below the sum of the three intervals, and the resolver
  activated no backend outside `claude`, `codex`, `glm`.
- **RED-now**: none by design.
- **green** (M3): `<SCRUB>go test -count=1 -v -run '^(TestAuditMulti_ConfigPlanKeepsParallelFanOut|TestRunMultiAudit_ParallelFanOut_AC_AMM_002)$' ./internal/cli/` →
  two `--- PASS:` lines, exit 0.
- **Mutant probe**: a serialized fan-out (legs run one after another) exceeds the
  bound and fails the first test.

## AC-ACV-018 — the committed `multi` does not leak into existing tests that read the config or the root

**Covers**: maps REQ-ACV-007, REQ-ACV-017

- **Classification**: regression-guard (GREEN before M6; its job is to stay GREEN
  after the yaml flip, in a session whose environment points at the real tree).
- **Given** M6 has landed (the committed yaml says `model: multi`), **When** the
  existing tests that read `workflow.yaml` or resolve the root from
  `CLAUDE_PROJECT_DIR` run with that variable set to this tree's root, **Then**
  they all pass. The set is taken from measurement (research.md R-29), not from
  a name guess: the whole of `internal/auditreceipt`, `internal/config` and
  `internal/runtime`; in `internal/hook` the receipt-guard families
  (`TestAuditReceiptGuard…`, `TestWSR00[789]…`, `TestSubagentStop…`, which resolve
  a tree and read the gate); in `internal/cli` every audit family — the review-gate
  tests (`TestRunMultiReviewGate_…`, `TestRunCodexReviewGate_…`,
  `TestMultiReviewGate_…`, `TestReviewGate_…`), `TestHandleCodexReviewGate…`, the
  `audit_multi` / convergence / codex / GLM / Claude audit families, the
  `TestConfigOrphanedWorktree…` and `TestWSR0…` families, `TestAC…` and the
  build-identity tests that touch the gate readers.
- **RED-now**: none by design; a non-hermetic test turns RED only after M6.
- **green** (M6): three commands, each of the form
  `unset MOAI_KANBAN MOAI_KANBAN_ID MOAI_KANBAN_LABEL MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED && CLAUDE_PROJECT_DIR=/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1423 go test -count=1 -v` plus the packages and
  `-run` pattern named in (a)-(c):
  (a) `./internal/auditreceipt/ ./internal/config/ ./internal/runtime/` with no
  `-run`; (b) `./internal/hook/` with `'^(TestAuditReceiptGuard.*|TestWSR00[789].*|TestSubagentStop.*)$'`;
  (c) `./internal/cli/` with `'^(TestAuditMulti_.*|TestRunMultiAudit_.*|TestConverge_.*|TestConvergence_.*|TestCodexAudit.*|TestPerformCodexAudit_.*|TestPerformGLMAudit_.*|TestDefaultBackendCaller_.*|TestPersist.*|TestLoadConvergenceResult.*|TestRunMultiReviewGate_.*|TestRunCodexReviewGate_.*|TestMultiReviewGate_.*|TestReviewGate_.*|TestHandleCodexReviewGate.*|TestConfigOrphanedWorktree.*|TestWSR0.*|TestAC.*|TestGLMAudit.*|TestClaudeAudit.*|TestAuditPinLive.*)$'`
  → each ends `ok  	github.com/modu-ai/moai-adk/internal/…`, no `--- FAIL`, and (c) lists the
  `--- PASS:` lines of at least `TestAuditMulti_RespectsCodexGateOff_AC_AMM_014`,
  `TestAuditMulti_NoHardErrorPath_AC_AMM_024`,
  `TestRunMultiAudit_ParallelFanOut_AC_AMM_002` and one `TestRunMultiReviewGate_`
  test (an empty swept set prints `ok` too — read the names).
- **Mutant probe**: removing M1's hermetic root from one audit test makes that
  test red here after M6 (the intended detection).

## AC-ACV-019 — a GLM- or GPT-launched lane still gets an independent Claude verdict

**Covers**: maps REQ-ACV-019

- **Classification**: regression-guard (the existing behaviour, held under the new
  plan).
- **Given** launch provider `glm` (then `gpt`), `audit.model: multi`, a supplied
  `claude_verdict` of `pass`, and stubs where the Claude backend returns `fail`,
  **When** `audit_multi` runs, **Then** the supplied verdict is ignored, the
  Claude backend is invoked, the result's claude entry has
  `source: "mcp_claude_audit"` and verdict `fail`, and the overall verdict is
  `fail`; and with the Claude backend `inconclusive` under `multi` the result
  fails by name (`gate_unmet` contains `claude`).
- **RED-now**: none by design.
- **green** (M3): `<SCRUB>go test -count=1 -v -run '^(TestAuditMulti_GPTAndGLMOriginsRunActualClaude_AC_CLA_009_010|TestAuditMulti_GLMOriginUnderMultiPlan)$' ./internal/cli/` →
  the existing test's sub-tests and `--- PASS: TestAuditMulti_GLMOriginUnderMultiPlan `,
  exit 0.
- **Mutant probe**: a plan path that lets a supplied anchor stand in for the
  backend under a non-Claude origin fails the second assertion.

## AC-ACV-020 — the codex leg has a deadline; any end of the leg is an unmet gate naming codex

**Covers**: maps REQ-ACV-020, REQ-ACV-008, REQ-ACV-007

- **Classification**: release-blocking (RED-now: E17, E18; positive controls E17c,
  E18c).
- **Given** a temp project root whose `workflow.yaml` sets only
  `audit.model: multi`, claude and glm stubs returning `pass`, a
  `codexReviewRPC` stub, and `config.DefaultCodexAuditLegTimeout` shortened to a
  few milliseconds for the test (restored by `t.Cleanup`), **When** `audit_multi`
  runs, **Then**
  (a) with a stub that blocks until its context ends, it returns within a bound
  well under one second; the codex `per_backend_verdicts` entry is
  `inconclusive` with a summary naming a codex timeout; `overall_verdict` is
  `fail`, `gate_unmet` is `codex`, and `residual_risk_note` names codex;
  (b) with a stub that returns a `pass` only after the deadline, the entry is
  still `inconclusive` — late output is never read as a verdict;
  (c) with a stub that holds partial `pass` review text and the CALLER's context
  cancelled mid-leg, the entry is `inconclusive` with a summary naming a
  cancellation (not a timeout) and the partial text is discarded
  (`awaitCodexTurnReview` returns it with a nil error on any context end);
  (d) with NO configuration the same hung stub yields the fail-open result
  (codex `inconclusive`, overall not failed) — the deadline changes the
  unconfigured case only by ending a hang that previously never ended.
- **RED-now**: E17 (no limit exists), E18 (no `context.WithTimeout` around the
  codex leg) — the leg currently has no deadline, so the hung stub of (a) never
  returns.
- **green** (M3): `<SCRUB>go test -count=1 -v -run '^(TestAuditMulti_CodexLegDeadline_FailsClosedNamed|TestAuditMulti_CodexLegDeadline_LateOutputIgnored|TestAuditMulti_CodexLegCancel_DiscardsPartial|TestAuditMulti_CodexLegDeadline_UnconfiguredFailsOpen)$' ./internal/cli/` →
  four `--- PASS:` lines, exit 0; `grep -n "DefaultCodexAuditLegTimeout = DefaultCodexAuditTimeout" internal/config/defaults.go` →
  one line (the derivation, not a literal), exit 0; and
  `grep -rn "codexLegTimeout\|codexAuditTimeout" internal` → (no output), exit 1
  (no second seam, no constant in `mcp_codex.go`).
- **Mutant probe**: a leg with no deadline never returns and test (a) times out
  red; a deadline that still reads the partial text fails (b) and (c); a
  deadline that maps expiry to `pass` or drops the entry fails the named-gate
  assertions; a hard-coded `20 * time.Minute` literal fails the derivation grep.

## AC-ACV-021 — a result from a server that predates the plan is never trusted

**Covers**: maps REQ-ACV-022

- **Classification**: release-blocking (RED-now: E3 — the check does not exist;
  E22 — the old-server-shaped result it must reject).
- **Given** a temp tree with `audit.model: multi` (plan: claude and codex
  enforced-required) and a persisted convergence result for session `s1`, **When**
  `moai verify audit-plan --project-root <tree> --check-session s1` runs, **Then**
  `convergence_check` reports: (a) an OLD-SERVER-SHAPED result — `overall_verdict:
  pass`, empty `gate_unmet`, `fail_open_backends: [codex, glm]`, no `plan_source`,
  codex entry `inconclusive` — `ok: false`, `unmet: ["codex"]`; (b) the same
  result with the codex entry `pass` but NO `plan_source` — `ok: false` with a
  reason naming `plan_source`; (c) no codex entry — `ok: false`, `unmet:
  ["codex"]`; (d) the codex entry present, `pass`, but its gate `advisory` —
  `ok: false`; (e) codex `pass` with gate `required`, claude `pass`, and
  `plan_source: "config"` — `ok: true`; (f) no persisted file for the session —
  `found: false`, `ok: false`; (g) a tree with no audit configuration — the check
  is `ok: true` for any result (no enforced-required backend, no config-sourced
  gate). The verb exits 0 in every case and changes no file.
- **RED-now**: E3 (no verb, hence no check); E22 shows the result shape (a) is
  what a server that ignores the token produces for this very configuration
  (`pass`, empty `gate_unmet`, `fail_open=[codex]`).
- **green** (M4): `<SCRUB>go test -count=1 -v -run '^TestAuditPlanCmd_ResultCheck$' ./internal/cli/` →
  `--- PASS: TestAuditPlanCmd_ResultCheck ` and its seven sub-test names (a)-(g),
  exit 0.
- **Mutant probe**: a check that keys only on a non-empty `gate_unmet` passes (a)
  and (b) and fails them; one that ignores `plan_source` passes (b) and fails it;
  one that accepts an `advisory` entry fails (d); one that requires a result even
  for a default plan fails (g).

## AC-ACV-022 — an unreadable `workflow.yaml` is a distinct, PASS-blocking state

**Covers**: maps REQ-ACV-023

- **Classification**: release-blocking (RED-now: E3; document literal E28).
- **Given** a temp tree whose `workflow.yaml` is corrupt (for example
  `workflow: [unclosed`), **When** `moai verify audit-plan --project-root <tree>`
  runs, **Then** it prints JSON with `config_status: "unreadable"`,
  `cross_model_active: "unknown"`, `cross_model_required: "unknown"`, an empty
  `backends`, and a `note` naming the file and the parse cause, exits 0, prints no
  default plan, and with `--check-session` adds no `convergence_check`; the verb's
  reader is `loadWorkflowAuditSection` and not `workflowAuditPins` (which turns a
  parse error into an absent configuration, `audit_pin.go:58-64`); and the two
  auditors and `sync.md` (local and template) say that the state is recorded as a
  Gap and yields no PASS and is not the legacy path.
- **RED-now**: E3 (no verb); E28 (zero for `config_status: unreadable` in the
  four local documents).
- **green** (M4, M7): `<SCRUB>go test -count=1 -v -run '^TestAuditPlanCmd_UnreadableConfig$' ./internal/cli/` →
  `--- PASS: TestAuditPlanCmd_UnreadableConfig `, exit 0;
  `grep -n "loadWorkflowAuditSection" internal/cli/audit_plan_cmd.go` → at least
  one line, exit 0; `grep -n "workflowAuditPins" internal/cli/audit_plan_cmd.go` →
  (no output), exit 1; and `grep -c "config_status: unreadable" .claude/agents/moai/plan-auditor.md .claude/agents/moai/sync-auditor.md .claude/skills/moai-ref-cross-model-audit/SKILL.md .claude/skills/moai/workflows/sync.md` →
  each at least `1`, exit 0 (a test-of-record for the text; the mechanical part is
  the verb's output above).
- **Mutant probe**: a verb that reads through `workflowAuditPins` prints the
  default plan with exit 0 for the corrupt file and fails the first test; one that
  exits non-zero without the JSON state is indistinguishable from an invalid
  configuration and fails the contract assertions.

## Edge cases

- **EC-1 Matching.** Token and gate values match exactly and case-sensitively
  after trimming surrounding whitespace (AC-ACV-003).
- **EC-2 Config-orphaned worktree with no identifiable primary.** The existing
  rule stands: codex is assumed `required` with its note, so the gate is
  fail-closed (`resolveAuditGates`, unchanged).
- **EC-3 All-off plan.** `audit.model: codex` with `audit.gates.codex: off`
  leaves every backend off; the verb prints `cross_model_active: false` and no
  `enforced_required` entry, and `audit_multi` keeps its existing result for a
  fan-out with no participants (this SPEC does not change it).
- **EC-4 Unreadable yaml.** The verb prints the distinct unreadable state
  (AC-ACV-022); the auditors and the sync step record a Gap and do not PASS.
  `audit_multi` called outside them keeps today's reading (spec.md R-4).
- **EC-5 Old build.** `moai verify audit-plan` prints the `verify` group's help
  and exits 0; the auditors detect the missing output contract and take the
  legacy path with the named Gap (AC-ACV-013).
- **EC-6 Cancellation versus deadline.** A leg ended by the caller's own
  cancellation is reported as a cancellation, not a timeout — and, like a
  timeout, it is `inconclusive` and discards any partial text (AC-ACV-020 (c)).
- **EC-7 New CLI, old MCP server.** The verb answers; the old server's persisted
  result has no `plan_source`; the result check fails closed by name until the
  server is reconnected (AC-ACV-021 (a)/(b); spec.md R-8).

## Quality gates and Definition of Done

1. All 22 criteria pass; each release-blocking criterion's RED-now was observed
   red on the pinned tree (ledger) and the named Go test's own RED output is in
   `progress.md §E.2`.
2. `moai spec lint SPEC-AUDIT-MODEL-CONVERGE-001` exits 0; `go build ./...` and
   `GOOS=windows GOARCH=amd64 go build ./...` exit 0; `golangci-lint run` shows
   no NEW finding against the pre-run baseline; `make build` exits 0 (agents-emit
   check, tool-policy drift check, catalogue hashes).
3. Coverage of the new files `internal/config/audit_plan.go` and
   `internal/cli/audit_plan_cmd.go` is at least 85% (a coverage profile of the
   criteria's named tests, read per file with `go tool cover -func`; a
   package-level percentage does not satisfy this item).
4. The distributed template `workflow.yaml`, the Go default `Audit.Model`, the
   MCP catalogue, `internal/runtime/sync_4dim_binding.go`, `review.md`, and the
   pre-existing local-vs-template differences of research.md R-15 are unchanged.
5. TRUST 5: Tested (the criteria above); Readable (English comments, named
   constants, no inline env names); Unified (gofmt, golangci-lint); Secured (no
   secret in output — the verb prints gates and a persisted-result summary, never
   keys); Trackable (commit subjects carry `SPEC-AUDIT-MODEL-CONVERGE-001` and
   `t1423`).
6. **Install and verify (leader-owned; plan.md §J).** The card is merged into
   develop; a release-candidate build is made from develop and installed; the MCP
   server is reconnected (taken together — the one window with a new CLI and an
   old server fails closed by name); then, on the installed build,
   `moai verify audit-plan --project-root <abs toplevel>` prints a JSON plan with
   `config_status: "ok"` and `model: "multi"` in this repository, and one
   `audit_multi` call with a `session_id` followed by
   `moai verify audit-plan --project-root <abs toplevel> --check-session <id>`
   prints `convergence_check.ok: true`. Only then does the new audit path apply
   to later audits. Until then the auditors and the sync step run the legacy path
   and name "plan surface unreachable, legacy path used" in the verdict's Gaps —
   no outage, nothing fail-closed on the missing verb.
7. The activation text (the local copies of the two auditors, the skill and
   `sync.md`) is the card's last commit (M7).
8. The sync report names the pre-existing prose interpreter in `/moai review`
   Phase 3.5 (`review.md`, left by decision D11) and the operator-owned rc
   install.

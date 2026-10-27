auditor-model: glm-5.3
verdict: FAIL
audited_sha: e3b7e7145010ea994ed90386404ff42703da1774

# t1988 sync-audit — GH #1924 refusal-text repair + GH #1926 hash-doc drift repair (Class B)

Independent sync-auditor (depth-1 leaf), run tnbn7e. Audited scope: `git diff cb30e05b3e..HEAD` (4 files, +36/−5) on branch `WT-triage-github-issues`, worktree `/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1988`. Audit plan for this tree: `config_status: ok`, `cross_model_active: true`, gates — claude `off`, codex `required`, glm `off` (all `config.gates`).

## Claim

| # | Claim | Result |
|---|---|---|
| C1 | (Lane) The unknown-attribution refusal no longer claims the launcher stamps the operator mark; it states stamping is manual and names the exact prefix | VERIFIED (E1–E3) |
| C2 | (Lane) Both `phase-execution.md` copies now describe `ComputeHash` as byte-exact, matching runtime and SSOT | VERIFIED (E4) |
| C3 | (Lane) `factory set-spec` exists in current source; the #1926 absence claim was a stale-install observation | VERIFIED (E5) |
| C4 | (Auditor) The template-twin edit was committed WITHOUT regenerating the deployed-catalog hash pin, so a shipped test fails at HEAD | VERIFIED DEFECT (E7), independently found by the required codex backend (E8) |

Overall: **FAIL** — Functionality (must-pass) is breached by blocking finding F1; the required codex gate returned `fail`.

## Evidence (command + verbatim output, this run, this tree)

**E1 — Targeted refusal test (Functionality).**
Command: `go test ./internal/cli/ -run TestFactoryUnknownAttributionErrorStatesManualStamp -count=1`
Output:
```
ok  	github.com/modu-ai/moai-adk/internal/cli	1.306s
```
test_exit=0.

**E2 — Test falsifiability against the OLD wording.**
Command: `git show cb30e05b3e:internal/cli/factory_card.go | sed -n '/^\/\/ factoryUnknownAttributionError/,/^}/p`
Output (old format string, verbatim):
```
return fmt.Errorf("%s: refused — %s: set %s=%s in the operator session (the launcher or the operator environment stamps it); the absence of lane variables confers nothing",
```
Constant values confirmed at source: `internal/config/envkeys.go:401` `EnvFactoryRole = "MOAI_FACTORY_ROLE"`, `:411` `FactoryRoleLane = "lane"`, `:420` `FactoryRoleOperator = "operator"`. Reconstructed old message probed with grep: forbidden phrase `the launcher or the operator environment stamps it` → count 1; `manual` → count 0. The new test's assertions 2 and 3 therefore FAIL against the old wording — the test can fail and is real evidence, not a tautology.

**E3 — Writer enumeration (Security / auth surface).**
Command: `grep -rn "EnvFactoryRole" internal/ --include="*.go" | grep -v "_test.go"`
Output (abridged to the load-bearing lines): the ONLY non-test writer is `internal/cli/factory.go:858` `_ = os.Setenv(config.EnvFactoryRole, config.FactoryRoleLane)` (lane only); the operator-mark reader is `internal/cli/mcp_refusal.go:149` `return os.Getenv(config.EnvFactoryRole) == config.FactoryRoleOperator`. No code path stamps `operator`. The new text's claim "the launcher stamps only lane sessions" matches source.

**E4 — Doc parity + runtime hash semantics (Consistency).**
Command: `diff .claude/skills/moai/workflows/run/phase-execution.md internal/template/templates/.claude/skills/moai/workflows/run/phase-execution.md`
Output: (empty) `diff_exit=0` — the two copies are byte-identical, not just in the edited region. Both now read "hashes the // byte-exact content of each present subject file".
Runtime source `internal/runtime/audit_cache.go` (verbatim excerpt): `// ComputeHash computes a combined, byte-exact SHA-256 hash of the plan` / `// artifact files present in specDir.` and the body writes `_, _ = h.Write(data)` after a `name:len\x00` separator — raw bytes, no normalization. SSOT `.claude/rules/moai/workflow/spec-workflow-detail.md:101` says "as exact bytes". Sweep `grep -rn "whitespace-normalized" .claude internal` → only unrelated `*_test.go` comments (binding-ledger/doc-comparison normalization); no other surface describes the plan-artifact hash as normalized.

**E5 — set-spec existence from source.**
Commands: `grep -n "newFactorySetSpecCommand" internal/cli/factory_card_metadata.go` → `115:// newFactorySetSpecCommand builds ...` / `118:func newFactorySetSpecCommand() *cobra.Command {`; `sed -n '90,98p' internal/cli/factory_handoff_recover.go` → line 94 `factory.AddCommand(newFactorySetSpecCommand(), newFactoryReclassifyCommand())`. Verified from source; no reliance on `./bin/moai`.

**E6 — Scope discipline + format.**
Command: `git diff --name-only cb30e05b3e..HEAD`
Output:
```
.claude/skills/moai/workflows/run/phase-execution.md
internal/cli/factory_card.go
internal/cli/factory_unknown_attribution_test.go
internal/template/templates/.claude/skills/moai/workflows/run/phase-execution.md
```
Exactly the 4 claimed files; no `.github/workflows`, no config-default, no factory DB, no actor-spoofing changes. `gofmt -l internal/cli/factory_card.go internal/cli/factory_unknown_attribution_test.go` → no output, exit 0.

**E7 — BLOCKING DEFECT (verified by this auditor's own run, not only by codex).**
Command: `go test -count=1 ./internal/template -run TestManifestHashFormat`
Output:
```
--- FAIL: TestManifestHashFormat (0.02s)
    catalog_tier_audit_test.go:493: CATALOG_HASH_UNSTABLE: moai stored hash=18df0ebde3a82e0718c904dedc9c3ac8970c2dd6cb4676d5b9e29be107898286, computed hash=33ee619f96dcea5430dabdfce9db9b1ec912f62b80c76766b7660602e3898644 (source=.claude/skills/moai/ (whole tree))
    catalog_tier_audit_test.go:498: audited 70 catalog entries for hash validity
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/template	0.299s
FAIL
```
exit=1. Attribution that the PARENT was green and this card introduced the failure: `git log --oneline 14bb399163..cb30e05b3e -- internal/template/templates/.claude/skills/moai/` → empty (no template-scope commit between the pin regeneration at `14bb399163`, card t1957, and this card's base); `internal/template/catalog.yaml` is tracked, last changed at `14bb399163`, and absent from this diff; the template file content changed (`shasum -a 256`: parent `1b6e876c…` → working `e6e4ce91…`). Generator contract: `internal/template/scripts/gen-catalog-hashes.go:9` — `--all  Update every entry's hash field in catalog.yaml in-place`; the Makefile `build` target runs it, but the regenerated pin was never committed with this change.

**E8 — Cross-model convergence (audit plan applied).**
Call: `mcp__moai__audit_multi` (project_root = this worktree, target = `baseBranch`, focus = factory role-gate refusal wording and plan-artifact hash doc parity; no `gates` argument; GLM main session so no `claude_verdict`).
Digest members: `overall_verdict: fail`; `gate_unmet`: absent (empty); `plan_source: config`; per-backend: `codex / required / fail` (sole participant; claude and glm off per plan); `disagreement_flag: null`; `participant_count: 1`; `residual_risk_note: "required-backend FAIL: codex"`; `fail_open_backends: []`; `audit_receipt: rcpt-bafdd64fe6a0f83245d137d0`; server `build_commit: b1806498fe`.
The codex finding is the same defect as E7; per verification-claim integrity I re-verified it with this tree's own toolchain before accepting it. This auditor is read-only: the orchestrator writes these digest members fresh to `<toplevel>/.moai/state/audit-plan-result.json` and runs `moai verify audit-plan --project-root <toplevel> --result-file <path>`; the exported FAIL here is not final until that check runs (a fail-side convergence cannot be flipped by the check, but the procedure is named).

### Findings

- **F1 [High] [blocking]** `internal/template/catalog.yaml:7` (entry `moai`, path `templates/.claude/skills/moai/`) — deployed-catalog hash pin stale after the template-twin edit; `TestManifestHashFormat` FAILS at HEAD (stored `18df0ebd…`, computed `33ee619f…`). CI runs this package; the commit ships red. Required fix: run `go run ./internal/template/scripts/gen-catalog-hashes.go --all`, commit the regenerated `internal/template/catalog.yaml` (entry `moai` expected to become `33ee619f…`) as part of this card, and re-run `go test -count=1 ./internal/template -run TestManifestHashFormat` to green.
- **F2 [Low] [optional]** `.moai/reports/t1988/progress.md` / `verdict.md` — the lane's verification batch covered `internal/cli` only and never executed the `internal/template` package its own diff touched; its Gaps list names only the `internal/cli` suite. The catalog-pin obligation of template edits is the process gap behind F1. Optional fix: add the `internal/template` hash test to the lane-local required set for any diff under `internal/template/templates/`.
- **F3 [Low] [optional]** residual from the lane's own list, concurred: the kickoff-side refusal still does not point auditors at the runtime hash definition, so an auditor computing `plan_artifact_hash` with an arbitrary pipeline can still wedge plan→kickoff even with the corrected docs. Follow-up candidate, outside this card's scope.

## Baseline-attribution

- Tree: `/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1988` (git worktree), branch `WT-triage-github-issues`, base `cb30e05b3e`, audited HEAD `e3b7e7145010ea994ed90386404ff42703da1774`, working tree clean at audit open.
- Every test/grep/diff/gofmt output above was produced in this run, against this tree, with the local Go toolchain compiling from tree source (no installed-binary path judged the Go results).
- The MCP convergence server is the installed build `b1806498fe`; `git merge-base --is-ancestor b1806498fe e3b7e7145010ea994ed90386404ff42703da1774` → exit 1 (not an ancestor of the audited HEAD — an older, differently-branched build). Its finding was independently re-verified against this tree's own toolchain (E7), so the build lag does not carry into the outcome; the receipt id is cited from that result.

## Gaps (explicitly unobserved)

- Full `go test ./...` NOT run (lane-local machine-load doctrine per the dispatch); the `internal/cli` full suite and the platform matrix are unrun — CI carries them.
- plan-audit→kickoff end-to-end reproduction NOT run (requires live factory-record mutation — dispatch-prohibited); mechanism verified at source/gate level only.
- Parent-commit greenness of `TestManifestHashFormat` established by git-graph attribution (empty commit window between the pin regen and base plus unchanged pin), NOT by executing the test at `cb30e05b3e` — executing it there would require mutating this tree.
- One Bash invocation was REFUSED by the worktree-isolation guard (the anchored regex `^TestManifestHashFormat$` inside a piped construct was "too complex to verify"); the identical test was re-run as a plain command without anchors and observed directly (E7). No other tool refusals occurred in this audit.
- No `card_id` was passed to `audit_multi` (no contract-mode binding evidence for this card), so no second-review record was appended to the card evidence directory.

## Residual-risk

- Convergence ran with a single participant (codex `required`; claude and glm `off` per the tree's plan); `disagreement_flag: null` — no agreement claim is made. Verbatim `residual_risk_note`: `"required-backend FAIL: codex"`.
- After F1's fix lands, any further drift under `templates/.claude/skills/moai/` changes the computed tree hash again; the regeneration must be the last step before the confirming re-audit, which is scoped to the F1 delta (regenerated `catalog.yaml` + green `TestManifestHashFormat`) plus a spot re-run of E1/E4.
- #1924 requests 2–3 (a sanctioned operator-stamp verb; an operator-action card route) and the #1926 (b)/(c) gates remain open leader/operator design decisions; the lane's by-design classification is consistent with the issue text ("The gate itself … looks intentional and correct") and no loosening was attempted — correctly.
- The doc-catalog pin mechanism protects deployed-file integrity for `templates/`; the local `.claude/skills` twin has no equivalent pin, so future single-sided edits (local without template, or vice versa) would be caught only by a `diff` of the pair, which is not an automated check.

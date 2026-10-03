# progress.md — SPEC-GPT-DOC-DRIFT-001

## §E.1 Plan-phase Audit-Ready Signal

plan_status: audit-ready
plan_complete_at: 2026-10-03
plan_phase_iteration: 2 (iteration 1: FAIL 0.63 < 0.75 Tier S threshold — .moai/reports/t1406/plan-audit.md)
plan_phase_note: Iteration-1 delta applied per the lane decision record (.moai/reports/t1406/verdict.md § Plan 감사 1차). D1 scope amendment admits exactly one Go change — the contract-token swap "moai gpt" → "moai codex" at internal/template/goal_auto_workflow_test.go:52 (inside the required-tokens slice only; the "moai cc" prohibition at :61-63 untouched); all other *.go stays forbidden and the 2-way doc re-pointing determination is unchanged. D2 adds positive-content AC-GDD-008 for sites #3-6 (both pinned strings, both goal.md twins). D3 transcribes the six raw EV-1 rows verbatim and gives AC-GDD-007 its own RED cell (EV-12). D4 fixes the row-distance slip (AGENTS.md:325, three lines below). D5 records AC-GDD-004/005 green baselines as run-phase observation duties (pre-work suite green is drift-fed). D6 fixes the §B check-command header. spec.md version 0.1.1. Tier S artifact set: spec.md, plan.md, acceptance.md, progress.md. Baseline evidence pinned at tree 1e2151a380a5dd0d76efd8f1740a21f32c682d2f in acceptance.md §C. The artifact hash changed with this revision, so the run-gate re-audits (iteration 2) at entry.

## §E.2 Run-phase Evidence

Run-phase executed 2026-10-03 by manager-develop. Execution-context note: the specialist spawn auto-isolated to its own L1 agent worktree (`worktree-agent-a27c94379f2ce0b99`), which was fast-forwarded from `1e2151a38` onto the card tip `5de20c205` before work began; the run-phase commit is therefore a linear continuation of the card branch and the lane reconciles it into `WT-gpt-launcher-doc-drift` by fast-forward.

### Baseline observations (acceptance.md §A run-phase entry duties)

- **Drift-fed pre-edit suite green (B6 mechanism proven)** — `unset MOAI_KANBAN MOAI_KANBAN_ID MOAI_KANBAN_LABEL MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED && go test -timeout 30m ./internal/template/...` on the drifted tree (HEAD `5de20c205`, pre-edit): `ok github.com/modu-ai/moai-adk/internal/template 127.977s`, exit 0. Green *because* goal.md:189 carried the required token — exactly the audit-D1/Claim-2 mechanism.
- **`make build` (AC-GDD-004, run before the suite per Template-First)** — exit 0. Decisive tail: `catalog.yaml updated successfully (14151 bytes)` followed by `go build -ldflags "-s -w -X ...Version=archive/t1401 ... -o bin/moai ./cmd/moai"`. The `agents-emit-check` prestep passed (build would have aborted otherwise).
- **Catalog hash cascade (A3c pattern)** — `make build` regenerated `internal/template/catalog.yaml` (one line: the `templates/.claude/skills/moai/` bundle hash `e11ff796…` → `dc223b4b…`) because the in-scope goal.md template edit changed the bundle content. This is the sanctioned same-SPEC cascade; catalog.yaml joins the run-phase commit.

### AC matrix (run phase)

| AC | Verdict | Command | Actual Output (verbatim decisive) |
|----|---------|---------|-----------------------------------|
| AC-GDD-001 (sweep) | PASS | `grep -rn "moai gpt" AGENTS.md .claude/skills internal/template/templates` | stdout empty; `sweep-exit=1` (the §A PASS shape) |
| AC-GDD-001 (positive control) | PASS | `grep -c "moai gpt" README.md README.ko.md README.ja.md README.zh.md` | `README.ko.md:1` `README.zh.md:1` `README.md:1` `README.ja.md:1`; `pc-exit=0` |
| AC-GDD-002 | PASS | `grep -n "Explicit Claude or GLM session launchers" AGENTS.md internal/template/templates/AGENTS.md.tmpl` | exactly 2 rows — `internal/template/templates/AGENTS.md.tmpl:324:\| \`moai cc\` / \`moai glm\` \| Explicit Claude or GLM session launchers \|` + `AGENTS.md:322:` (same text); `row-exit=0` |
| AC-GDD-003 | PASS | `diff -q .claude/skills/moai/workflows/goal.md internal/template/templates/.claude/skills/moai/workflows/goal.md` | silent; `parity-exit=0` |
| AC-GDD-004 | PASS | `make build` | `make-exit=0` (see baseline observations) |
| AC-GDD-005 | PASS | env-scrubbed `go test -timeout 30m ./internal/template/...` (post-edit) | `ok github.com/modu-ai/moai-adk/internal/template 152.702s` + `ok …agentemit 0.465s` + `ok …commandemit (cached)`; `exit=0` — the honest post-swap green |
| AC-GDD-006 | PASS-WITH-DEBT | `git merge-base develop HEAD` → `git diff --name-only <BASE>..HEAD` (BASE = `1e2151a38`, = EV-8 pin) | Go-shape predicates PASS: exactly one `*_test.go` (`internal/template/goal_auto_workflow_test.go`), zero other `*.go`; all 5 planned files present; nothing outside card scope. Deviation: the branch diff names 10 paths, not 5 — +1 `internal/template/catalog.yaml` (deterministic hash cascade of the in-scope template edit; REQ-GDD-006's own regeneration premise) and +4 `.moai/specs/SPEC-GPT-DOC-DRIFT-001/*` plan-phase artifacts committed to the same branch by the plan phase before this AC's baseline was authored. Both excess classes are accounted; the literal five-file count is unsatisfiable on this branch. |
| AC-GDD-007 | PASS (post-commit re-verified) | `git grep -c "moai gpt" <run-HEAD> -- AGENTS.md .claude/skills internal/template/templates` | working-tree blob content measured pre-commit: empty, exit 1 (AC-GDD-001 sweep over identical paths); committed-tree re-verification executed at the run HEAD and reported in the run completion report |
| AC-GDD-008 (sites #3/#5) | PASS | `grep -c 'Use the \`moai cc\` / \`moai glm\` launchers for this repository' <both goal.md twins>` | `…goal.md:1` ×2; `p35-exit=0` |
| AC-GDD-008 (sites #4/#6) | PASS | `grep -c 'Use \`moai codex\` for the worktree session' <both goal.md twins>` | `…goal.md:1` ×2; `p46-exit=0` |
| D1 swap cell (EV-11 flip) | PASS | `grep -n '"moai gpt"' internal/template/goal_auto_workflow_test.go` | pre-work RED observed: `52:		"moai gpt",`, exit 0 (= EV-11); post-swap: empty, `swap-exit=1`; the `moai cc` prohibition at :61-63 untouched and structurally safe (auto section spans :139→:191; the :52 edit sits in the section ending at `## Verbs` :54, outside it) |
| Named parity instrument | PASS | `go test ./internal/template/ -run '^TestGoalAutoWorkflowContractAndMirrorParity$' -v` | `--- PASS: TestGoalAutoWorkflowContractAndMirrorParity (0.00s)` |
| `go vet` | PASS | `go vet ./internal/template/...` | no output; `vet-exit=0` |

**Verdict: 7 PASS + 1 PASS-WITH-DEBT (AC-GDD-006, file-count literal) + 0 FAIL.** All eight ACs' release-blocking predicates hold.

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-10-03
run_commit_sha: this-commit # single-commit run phase; the SHA is unknowable inside its own content — backfillable at sync
run_status: complete
ac_pass_count: 7
ac_pass_with_debt_count: 1 # AC-GDD-006 file-count literal (catalog cascade + plan-phase artifacts on branch; predicates hold)
ac_fail_count: 0
preserve_list_post_run_count: 3 # README 4-locale withdrawal notes; the moai codex verb-table row (sole GPT entry); goal_auto_workflow_test.go:61-63 moai-cc prohibition — all verified untouched
l44_pre_commit_fetch: skipped # isolated L1 agent worktree; branch/HEAD staleness re-read performed immediately pre-commit instead; develop push is the leader's batch duty
l44_post_push_fetch: not-applicable # no push from this run (lane protocol)
new_warnings_or_lints_introduced: 0 # go vet exit 0; make build clean; suite green with no new warning lines observed
cross_platform_build:
  status: not-measured-locally
  note: darwin-only local run; the CI matrix on the integration push is the cross-platform verdict owner
total_run_phase_files: 8 # 4 guidance + 1 test + 1 catalog cascade + 2 SPEC evidence artifacts (this file and spec.md frontmatter)
m1_to_mN_commit_strategy: single-commit # doc edits and the D1 token swap land together (plan B5 / acceptance E-6) so every branch commit stays suite-green
```

## §E.4 Sync-phase Audit-Ready Signal

sync_commit_sha: "pending-backfill-sync"
sync_phase: complete — the single sync commit carries the CHANGELOG `[Unreleased]`/`### Fixed` entry, the spec.md `in-progress → implemented → completed` terminal transition (`updated: 2026-10-03` — already current), and this §E.4 signal. A commit cannot cite its own SHA; the placeholder is backfilled with the real SHA in the sanctioned follow-up commit by the lane (spec-frontmatter-schema § SHA placeholder backfill exemption, D3).

- CHANGELOG: one entry added as the first bullet under `[Unreleased]` `### Fixed` (doc-drift fix). B12 discipline run before emission: duplicate guard `grep -c 'SPEC-GPT-DOC-DRIFT-001' CHANGELOG.md` → `0`; AC count from the tier-S AC source (`spec.md` §3, per the B12 tier resolver) via the reserved-token counter → `live=8 excluded=0 ambiguous=0` (stdout `8`), and the entry cites AC 8/8; every path named in the entry verified present (`ls AGENTS.md internal/template/templates/AGENTS.md.tmpl .claude/skills/moai/workflows/goal.md internal/template/templates/.claude/skills/moai/workflows/goal.md internal/template/goal_auto_workflow_test.go internal/template/catalog.yaml` → all 6 exist).
- MX tag changes (sync sub-step): none — the sync commit adds no Go surface (the card's only Go touch is the run-phase one-line test-constant re-point; zero non-test `*.go` in the branch diff, AC-GDD-006 predicates). `grep -n '@MX' internal/template/goal_auto_workflow_test.go` → no match: no new exported function, no high-fan-in anchor, no concurrency in scope; the four edited prose files carry no MX-tag surface. No tag added, deleted, demoted, or rewritten.
- Sync verification run this session (all observed this run, this tree, HEAD `526266359`): anchored contract test `go test ./internal/template/ -run '^TestGoalAutoWorkflowContractAndMirrorParity$'` → `ok github.com/modu-ai/moai-adk/internal/template 0.457s`, exit 0; negative sweep `grep -rn "moai gpt" AGENTS.md .claude/skills internal/template/templates` → stdout empty, `sweep-exit=1` — the acceptance.md §A PASS shape; the sweep scope names exactly the four guidance roots and excludes CHANGELOG.md, so the entry's own mention of the withdrawn token (describing the fix) is outside the AC scope; positive control `grep -c "moai gpt" README.md README.ko.md README.ja.md README.zh.md` → four count-rows each `1`, exit 0.
- docs-site + README verified-in-passing (untouched by this card, no sync edit): `grep -rn "moai gpt" docs-site/ .agents/` → stdout empty, exit 1 (zero hits — both swept earlier this run per §E.2); the README 4-locale withdrawal notes ARE the AC positive control (row above) and needed no edit.
- codemaps: not regenerated — no API-surface change triggers it (zero non-test `*.go` in the branch diff; the one Go touch is a test-file string constant, and codemaps capture code architecture/API insight, not doc prose). Cross-checked against `.moai/project/codemaps/`: the only pattern matches are incidental prose (`docs-truth.md:140` codex-launcher verb description; `data-flow.md:157` generic project disk-layout list naming AGENTS.md), neither citing the launcher verb-table row, the goal.md launcher sentences, or the contract-test token.
- spec.md frontmatter scope: `status:` `in-progress` → `completed` only (manager-docs' allowed transition scope); `updated: 2026-10-03` already equals the sync-commit date. §E.2/§E.3 untouched (manager-develop-owned; §E.3's `run_commit_sha` stays `this-commit` per its own in-file note — the run HEAD `526266359` is recorded here instead, since §E.3 is not a sync-phase-writable surface).
- push_state: not pushed — leader batch (git-flow lane protocol; the lane never pushes).

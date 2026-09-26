# research.md — SPEC-AUTONOMY-CLOSURE-001

Every fact below was measured in the card worktree `.claude/worktrees/t1237` (branch
`WT-completion-report`, base `develop` at `553e224f3`) during plan phase, with the command named.
Line numbers are anchors at that base and drift.

## §A. Dependency snapshot

| Dependency | Source read | Exact SHA read | In this tree? |
|---|---|---|---|
| A1 — SPEC-AUTONOMY-CONTRACT-001 | tip of branch `WT-contract-schema` (`git rev-parse --short WT-contract-schema` → `67a2f55cb`), `design.md` / `spec.md` / `acceptance.md` via `git show 67a2f55cb:<path>` | `67a2f55cb` (spec v0.5.0) | no — `git merge-base --is-ancestor 6d98ca466 HEAD` → not ancestor; `ls internal/contract` → absent |
| A1 M1 code | `git ls-tree -r --name-only WT-contract-schema -- internal/contract` | `67a2f55cb` | no |
| A2 — SPEC-AUTONOMY-ESCALATION-001 | `git show 8c9ee29b7:.moai/specs/SPEC-AUTONOMY-ESCALATION-001/{spec,design}.md` | `8c9ee29b7` (spec v0.3.0) | no — `git merge-base --is-ancestor 8c9ee29b7 HEAD` → not ancestor |
| A3 — card t1236 | not read (not in this tree) | — | no |

Lead notice (2026-09-26): A1 moves to v0.5.1 at the tip of `WT-contract-schema`. At the moment of
reading the tip was still `67a2f55cb` (v0.5.0). Run phase re-reads the tip and re-checks §C.1.

A1 M1 exported surface used by A4 (read from `67a2f55cb`): `contract.Verify(Inputs) Report`
(`internal/contract/verify.go:73`), states `unsigned | signed-valid | signed-invalid`
(`verify.go:7-9`), `contract.LoadDir(specDir)` (`internal/contract/load.go:40`),
`contract.ResolveSpecDir(projectRoot, specID)` (`load.go:24`).

## §B. Codebase facts

### §B.1 SPEC ID collision

`ls .moai/specs | grep -i CLOSURE` and `git ls-tree -r --name-only 67a2f55cb -- .moai/specs | grep CLOSURE`
list only `SPEC-HARNESS-LOOP-CLOSURE-001` and `SPEC-BACKLOG-JSON-DISCLOSURE-001`.
`SPEC-AUTONOMY-CLOSURE-001` is free; the ID regex check printed `PASS`.

### §B.2 Where `audit_multi` results land today

- Handler `handleAuditMulti` at `internal/cli/mcp_audit_multi.go:47`; `project_root` resolved at
  `:74`; tool schema registered at `internal/cli/mcp_server.go:511` with parameters `claude_verdict`,
  `target`, `focus`, `gates`, `session_id`, `project_root` — no card or SPEC parameter.
- `ConvergenceResult` (`internal/cli/mcp_convergence.go:114`) carries `per_backend_verdicts`,
  `overall_verdict`, `disagreement_flag` (a `*bool`, nullable by design, `:130`),
  `participant_count` (`:139`), `audit_receipt` (`:149`), and build identity fields.
- Persistence: `persistConvergenceResult` (`mcp_convergence.go:903`) writes
  `.moai/state/audit-multi/<session>.json` **only when `session_id` is non-empty**
  (`mcp_convergence.go:764-769`). The file is keyed by session and carries no card, SPEC, or HEAD.
- Audit receipts: written only when the codex backend took part
  (`mcp_convergence.go:760`, "a fan-out that skipped codex is not evidence that a codex audit ran");
  store at `.moai/state/audit-receipts/receipts/<id>.json` (`internal/auditreceipt/store.go:37-38`,
  `:208`). The `Receipt` struct (`store.go:100-108`) carries tool, tree root, time, codex verdict,
  gate-unmet — no card, SPEC, or HEAD. A GLM-only fan-out leaves no receipt.

Consequence: no existing artifact binds a second review to a card, a HEAD, and a contract digest.
A4 adds the second-review record (design.md §A.1).

### §B.3 Audit verdict line

`auditreceipt.ParseVerdictLine` (`internal/auditreceipt/store.go:325`) reads the last non-empty line
against `^AUDIT-VERDICT: (PASS|PASS-WITH-DEBT|FAIL) spec=(SPEC…) receipts=(…)$` (`store.go:96-97`).
The plan-auditor and sync-auditor agent definitions require that line as the final line
(`.claude/agents/moai/plan-auditor.md:222`, `.claude/agents/moai/sync-auditor.md:183`). Plan-audit
reports live at `.moai/reports/plan-audit/<SPEC-ID>-review-<n>.md` (`ls .moai/reports/plan-audit`).
A4 uses this parser for the plan-audit binding (REQ-CLOSURE-011).

### §B.4 Verify evidence store is not durable

`moai verify` snapshots live under `.moai/state/verify/snapshots` (`internal/verify/store.go:15`)
with a default 10-minute freshness window (`internal/cli/verify.go` long help). They cannot serve as
closure evidence; A4 does not read them.

### §B.5 First-verdict shape in `progress.md`

Measured on `.moai/specs/SPEC-WT-DOCTRINE-CONFLICT-001/progress.md` (a completed SPEC): the §E.2
"AC matrix" is a Markdown table `| AC-WDC-001 | PASS | <evidence> |` (lines 97-109); §E.3 is a fenced
YAML block with `run_status`, `ac_pass_count`, `ac_fail_count`, `run_commit_sha` (lines 113-129). No
Go parser for these blocks exists (`grep -rn` for progress §E parsers in `internal/` found only the
era marker check `hasAnyProgressMarker`, `internal/spec/era.go:290`).

### §B.6 `verdict.md` is taken

`find .moai/reports -maxdepth 2 -name verdict.md | wc -l` → `381`. `.moai/reports/t750/verdict.md`
is a hand-authored Korean verdict (Claim / Evidence sections). The name is fixed by doctrine as the
card evidence path: `internal/template/templates/.claude/rules/moai/workflow/kanban-dispatch.md:199`,
`internal/template/templates/AGENTS.md.tmpl:147`,
`internal/template/templates/.claude/agents/moai/manager-lead.md:65`. A4 must not write it.

### §B.7 Reports are local-only

`git check-ignore -v .moai/reports/t1237/closure-report.md` → `.gitignore:235:.moai/reports/*`; the
same rule covers `second-review.json*`, `closure-verdict.json*`, and `verdict.md`; only
`.moai/reports/plan-audit/` is re-included (`.gitignore:251-252`). Evidence therefore lives in each
worktree's own `.moai/reports/`, not in git; a push from the integration worktree cannot read a card
worktree's reports by path unless it resolves that worktree (design.md §C.3).

### §B.8 Hooks and push

- `checkBashCommand` runs at `internal/hook/pre_tool.go:507` under an `@MX:ANCHOR` forbidding any
  conditional return above it; the branch guard follows at `:534`. The destructive denylist carries
  force-push patterns only (`pre_tool.go:219`, `:251`).
- A git `pre-push` hook exists in the template (`internal/template/templates/.git_hooks/pre-push`,
  bypass `SKIP_MOAI_PREPUSH=1`) with a CLI side at `internal/cli/hook_pre_push.go`.
- `git config --show-origin --get core.hooksPath` → `file:/Users/goos/MoAI/moai-adk-go/.git/config	/dev/null`:
  git hooks do not run in this repository at all. The PreToolUse guard is therefore the layer that
  observes agent pushes.

### §B.9 Integration branch source

`config.LoadGitFlowDevelopBranch(projectRoot)` at `internal/config/loader_integration_branch.go:34`;
local `.moai/config/sections/git-strategy.yaml:16` `develop_branch: develop`.

### §B.10 Queue store

`BacklogItem.SpecID *string` at `internal/kanban/backlog_store.go:81`; states `queued | picked |
dropped` (`:61-65`). A card → SPEC lookup and the reverse SPEC → card lookup are both reads.

### §B.11 Name overlap check

`internal/cli/codex_contract.go` exists but is the AGENTS.md ↔ CLAUDE.md "instruction contract"
(`codex_contract.go:3-7`), not a `contract` cobra command; `grep '"contract"' internal/cli/*.go`
found no existing `contract` command. A1 adds `internal/cli/contract.go`; A4 adds subcommands to it.

## §C. Divergences recorded (not blockers)

### §C.1 Kickoff decider value set is moving

- A1 at `67a2f55cb` (v0.5.0): single decider `human | llm | jev`, Jev-to-LLM fallback recorded as
  `requested_decider`, `signer`, `fallback_reason`, `jev_attempt`; `decisions` holds exactly one entry.
- Operator re-decision relayed by the lead after that commit: decider is `llm` (default) or
  `llm+jev` (cross-check); no Jev-alone; the report shows the decider, each answer (llm / jev), and
  whether a fallback happened with its reason. A1 v0.5.1 is expected to carry it.
- A4 therefore decodes the receipt leniently for display and hard-codes no decider value set
  (REQ-CLOSURE-010, design.md §F). Validity comes from A1 verify, not from A4.

### §C.2 Kickoff receipt location

The lead design names `.moai/reports/<card>/kickoff-receipt.json`; A1 fixes
`.moai/specs/<SPEC-ID>/kickoff-receipt.json` (A1 design § Kickoff Receipt, `--receipt` must resolve to
it). A4 reads the A1 location.

### §C.3 Closure report location

The lead design names `.moai/reports/<card>/verdict.md`; §B.6 shows that name is taken. A4 writes
`closure-report.{md,json}` (plan.md OQ-2).

## §D. Reuse analysis

| Need | Reused | New |
|---|---|---|
| Contract facts, verify state, AC count semantics | A1 `internal/contract` | — |
| Escalation records, card state file, class-4 comparison | A2 `internal/escalation` (record reader, detector function) | — |
| Agent-environment markers, TTY check | A1 `internal/contract/sign` marker set | — |
| Audit verdict line | `auditreceipt.ParseVerdictLine` | — |
| Integration branch | `config.LoadGitFlowDevelopBranch` | — |
| Atomic writes | `internal/atomicfile` | — |
| Second-review record, verdict record, report model, readiness rule | — | `internal/closure` |

## §E. Open items for run phase

- Confirm A2 exports its record reader and a read-only class-4 comparison; if not, New APIs renders
  `not observed` and the dependency is reported (plan.md §B R2).
- Confirm A2b exports a push-command classifier; if not, A4 ships the minimal one (design.md §C.1).
- Re-read the A1 tip for v0.5.1 receipt field names before M3 (the lenient decoder tolerates both).

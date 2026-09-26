# research.md — SPEC-AUTONOMY-CLOSURE-001

Every fact below was measured in the card worktree `.claude/worktrees/t1237` (branch
`WT-completion-report`, base `develop` at `553e224f3`) during plan phase, with the command named.
Line numbers are anchors at that base and drift.

## §A. Dependency snapshot

| Dependency | Source read | Exact SHA read | In this tree? |
|---|---|---|---|
| A1 — SPEC-AUTONOMY-CONTRACT-001 v0.5.1 | `git show 65e0a9167:.moai/specs/SPEC-AUTONOMY-CONTRACT-001/{design,spec,acceptance}.md` (v0.5.1 per `sed -n 4p` of spec.md → `version: "0.5.1"`) | `65e0a9167` | no — `ls internal/contract` → absent |
| A1 later commits (moving tip, not a pin) | `git diff --stat 65e0a9167 25283ebf8 -- .moai/specs/SPEC-AUTONOMY-CONTRACT-001` → 4 files, wording only (v0.5.2: provenance comment, A2b attributions, interim `llm+jev` rule phrased "while A3's amendment has not landed"); no field, code, or rule A4 consumes changes. Later commits on the branch (`git log --oneline -3 WT-contract-schema` → `7b0d20494`, `b7f9cc44b` M3 code) leave the SPEC directory unchanged (`git diff --stat 25283ebf8 WT-contract-schema -- .moai/specs/SPEC-AUTONOMY-CONTRACT-001` → empty). `65e0a9167` is the binding pin. | — | no |
| A2 — SPEC-AUTONOMY-ESCALATION-001 v0.3.0 | `git show 8c9ee29b7:.moai/specs/SPEC-AUTONOMY-ESCALATION-001/{spec,design}.md` | `8c9ee29b7` | no — `git merge-base --is-ancestor 8c9ee29b7 HEAD` → not ancestor |
| A3 — card t1236 | not read | — | no |

A1 exported surface used by A4, re-measured at `65e0a9167`: `contract.Verify(Inputs) Report`
(`internal/contract/verify.go:73`), states `unsigned | signed-valid | signed-invalid`
(`verify.go:7-9`), `contract.LoadDir(specDir)` (`internal/contract/load.go:40`),
`contract.ResolveSpecDir(projectRoot, specID)` (`load.go:24`).

## §B. Codebase facts

### §B.1 SPEC ID collision

`ls .moai/specs | grep -i CLOSURE` and `git ls-tree -r --name-only 65e0a9167 -- .moai/specs | grep CLOSURE`
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
A second stream exists: the plan-phase review files `plan-audit.md` / `plan-audit-iter<N>.md` under
`.moai/reports/<card-id>/` (`.claude/rules/moai/workflow/spec-workflow.md` § Report Persistence).
Measured in this worktree: `ls .moai/reports/plan-audit | grep -c -- -review-` → `3`;
`ls .moai/reports/*/plan-audit*.md | wc -l` → `9`. A4 uses this parser for the plan-audit binding and
searches both streams (REQ-CLOSURE-011).

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
same rule covers `second-review.jsonl`, `closure-verdict.jsonl`, and `verdict.md`. The plan-audit
directory is re-included only for its scaffold: `.gitignore:251-253` re-includes the directory,
ignores `.moai/reports/plan-audit/*`, and re-includes `.gitkeep`; `git check-ignore -v
.moai/reports/plan-audit/SPEC-X-001-review-1.md` → `.gitignore:317:.moai/reports/plan-audit/*.md`.
Plan-audit reports are therefore local-only per worktree too. Every A4 input lives in some tree's
own `.moai/reports/`, which is why spec.md §C.7 fixes one card evidence home for writers and readers,
and why the plan-audit search (REQ-CLOSURE-011) runs in that home.

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

## §C. A1 v0.5.1 deltas A4 consumes

Read at `65e0a9167` (`git diff 67a2f55cb 65e0a9167 -- .moai/specs/SPEC-AUTONOMY-CONTRACT-001` plus the
§ Kickoff Receipt and § Card Field sections of `design.md`):

- **Required `card` field** — top-level, matches `^[A-Za-z0-9][A-Za-z0-9_-]{0,63}$`, inside the digest
  and the seal, reported by `show --json`, reason `card_invalid` (design.md:22, :124, :146-163). A4
  takes a contract's card from it (spec.md §C.7) instead of a queue reverse lookup.
- **Receipt shape** — `requested_decider` and `effective_decider` ∈ `llm | llm+jev`;
  `fallback: {applied, reason}` with reasons `jev_disabled`, `jev_low_confidence`,
  `jev_malformed_response`, `jev_call_failed`, `jev_key_missing`; `llm_answer` always present;
  `jev_answer` present exactly when `effective_decider` is `llm+jev`; `outcome` ∈
  `approve | reject | human` (design.md § Kickoff Receipt, field rules 1-9). REQ-CLOSURE-010 displays
  these fields and the outcome.
- **Decider set** — `signer_kind` ∈ `human | llm | llm+jev` (design.md:79); a configured
  `kickoff.decider: jev` is the configuration error `kickoff_decider_jev_sole` (design.md:302, :532).
  A4 displays any token verbatim and hard-codes no set.
- **Kickoff receipt location** — `.moai/specs/<SPEC-ID>/kickoff-receipt.json` (A1 fixes it; the lead
  design's `.moai/reports/<card>/` path is not used).

## §C.1 A2 surface relevant to invariant honesty

A2 writes an escalation record only on a trip (A2 spec §I, REQ-AE-018, REQ-AE-020); an unexecuted
command invariant is labeled not-observed at the checkpoint without a record (REQ-AE-022, design.md
§C.2); the card state file holds the contract observation, the disarm flag, budget counters, and the
failure-fingerprint history, not invariant executions (design.md §C.6). No A2 file records that a
command invariant ran and passed, so REQ-CLOSURE-005 renders such invariants `not observed` unless a
violation record exists.

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
- Diff the landed A1 `design.md` § Kickoff Receipt and § Card Field against `65e0a9167` before M1.

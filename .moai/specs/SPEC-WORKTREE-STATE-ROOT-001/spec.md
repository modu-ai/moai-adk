---
id: SPEC-WORKTREE-STATE-ROOT-001
title: "State and catalogue roots for a config-orphaned linked worktree"
version: "0.2.1"
status: in-progress
created: 2026-09-26
updated: 2026-09-26
author: manager-spec (card t1213)
priority: P1
phase: "v3.2.0 target"
module: "internal/cli, internal/hook, internal/auditreceipt"
lifecycle: spec-anchored
tier: M
issue_number: 1716
related_specs: [SPEC-MCP-WORKTREE-UNTRACKED-001, SPEC-MCP-WORKTREE-ROOT-001]
tags: "worktree, untracked-moai, state-root, audit-receipt, review-gate, catalogue"
---

# SPEC-WORKTREE-STATE-ROOT-001

## HISTORY

| Version | Date | Author | Change |
|---|---|---|---|
| 0.1.0 | 2026-09-26 | manager-spec (t1213) | Initial plan-phase draft. Takes over the configuration/catalogue/state items SPEC-MCP-WORKTREE-UNTRACKED-001 §6 deferred to card t1213, grounded in the re-measurement of t1202 plan-audit-iter2 D18 and D25. |
| 0.2.0 | 2026-09-26 | manager-spec (t1213) | Plan-audit iter-1 revision (FAIL 0.71). Closes D1 (rejection records carry and are filtered by tree identity — REQ-WSR-003 widened to every record routed into a shared store), D2 (predecessor tests that break named; plan R4 corrected), D3 (change-detector seam in the review-gate criterion), D4 (`spec_drift` / `spec_audit` union predicates with a non-empty-findings premise), D5 (guard behaviour with an unresolved store: no marker or rejection, re-entrant stop, spawn denial — REQ-WSR-010), D6 (convergence results carry tree identity, never replace another tree's, and the gate blocks on any fail — REQ-WSR-003, REQ-WSR-007), D7 (`audit_multi` skip cell), D8 (decision-1 rationale narrowed to store reachability), D9 (`moai verify record`, rejection path, drift/audit, tool descriptions observed), D10 (criterion selector), and optional D11–D14 (unattributed observation removed, review-gate activation risk, `_root` tool list enumerated, REQ-WSR-001 stated as behaviour). Adds the review-gate root measurement matrix requested by the lead (acceptance AC-WSR-006) and the shared `project_root` parameter description to REQ-WSR-016. REQ and AC counts unchanged (16 / 16); AC-WSR-006 and the former AC-WSR-007 were merged into the matrix, and AC-WSR-007 now carries the rejection tree-identity criterion. |
| 0.2.1 | 2026-09-26 | manager-spec (t1213) | Plan-audit iter-2 (FAIL 0.78) accepted by the operator as PASS-with-debt; conditions applied. N1: AC-WSR-002 cell 2 gains cell 2b (gate input `cwd = P` must block on the `fail` `audit_multi` wrote for `W`) and the named mutant "P gate reads only `S.json`". N2: plan.md decision 2 and new risk R9 state the fail-closed spawn denial's blast radius (repositories without the receipt gate, permanent under separate-git-dir layouts, broader than the MCP side) and the release path (`workflow.yaml` in `W`). N3: `[REF]` markers on the two predecessor AC citations. N4: AC-WSR-007 mutant derivation corrected. No REQ change; REQ and AC counts unchanged (16 / 16). |

## §1 Problem

SPEC-MCP-WORKTREE-UNTRACKED-001 (completed) made the MCP `project_root` validator
accept a registered linked worktree `W` of a repository that keeps `.moai`
untracked, and routed one configuration key (`workflow.audit.gates`) of such a
**config-orphaned** root to the primary checkout `P`. It deferred three things to
this card (its §6, "Out of Scope — configuration, catalogue, and state routing"):

1. **State writes and their paired reads.** MCP tools write state under the root
   they were called with. On `W` that is `W/.moai/state/…`, which creates a `.moai`
   inside `W`. The consuming readers run elsewhere — in hook processes — and do not
   all resolve the same root.
2. **The hook-side receipt guard.** It reads the codex gate from the guarded tree's
   own `workflow.yaml`; on `W` that file does not exist, so the guard is a no-op.
3. **The SPEC catalogue.** `spec_progress` / `spec_drift` / `spec_audit` read `W`
   only. The primary catalogue is invisible, and the t1202 audit (D25) showed that
   simply switching to `P` would hide any SPEC written inside `W` while reporting
   success.

### §1.1 Evidence base

All reader claims below cite either the re-measurement
`.moai/reports/t1213/remeasure-d18-d25.md` (tree `c630de892`, a live worktree
session) or a code line on the same tree. Anything else is labelled unmeasured.

| Reader | How it picks its root | Evidence | Root it read |
|---|---|---|---|
| Hook `resolveProjectRoot` (rule-load audit log) | `CLAUDE_PROJECT_DIR` first, else stdin `cwd`; requires `.moai` | `internal/hook/post_tool_metrics.go:98` | **P** — measured: `rule-load-audit.jsonl` entries for worktree rule files landed in the primary checkout (remeasure §D18) |
| Hook process working directory (trace log) | `os.Getwd()` | `internal/cli/deps.go:130` | **W** — measured: `trace-<session>.jsonl` landed in the worktree (remeasure §D18) |
| Receipt guard `auditReceiptTree` | `TreeRootFromCWD(input.CWD)` → `git rev-parse --show-toplevel` of the stdin `cwd`, else the `cwd` itself | `internal/hook/audit_receipt_guard.go:52-60`, `internal/auditreceipt/treeroot.go:22-31` | **unmeasured** — depends on the stdin `cwd` value |
| Receipt guard gate read | `CodexGateRequired(tree)` reads `<tree>/.moai/config/sections/workflow.yaml`; read error → not required | `internal/auditreceipt/store.go:155-166` | on a config-orphaned `W`: file absent → guard inactive (code) |
| `codex-review-gate` / `multi-review-gate` Stop hooks | own parser `readHookInput` (plain `json.Unmarshal`, no env fill) → `resolveProjectDirFromInput`: stdin `project_dir` (legacy field), then stdin `cwd`, then `CLAUDE_PROJECT_DIR`; opt-in flag read from that root; then the change-detector seam `reviewGateChangeDetector` | `internal/cli/codex_review_gate.go:48, 74, 191, 207-219, 236-246`; `internal/cli/multi_review_gate.go:73, 162, 211-212` | **unmeasured** — depends on the stdin payload and environment; the run phase measures the input-to-root mapping (acceptance AC-WSR-006) |
| `multi-review-gate` state read | `<projectDir>/.moai/state/audit-multi/<session>.json` | `internal/cli/multi_review_gate.go:113-127` | follows the row above |
| `moai verify record` / `moai verify check` CLI | `--project-root`, else `CLAUDE_PROJECT_DIR`, else process `cwd` | `internal/cli/verify.go:63-77`, `internal/cli/session.go:272-280` | **unmeasured** in a hook; in the worktree session's Bash tool the variable was observed empty (remeasure §D18 "Gaps"), so there it resolves the Bash `cwd` |

Correction to the re-measurement's attribution: the remeasure file groups the two
review gates with the env-first `HookInput.ProjectDir` fill
(`internal/hook/protocol.go:127-132`). By code they do not take that path — their
parser never calls `validateInput`, so their root is the stdin `cwd` unless a legacy
`project_dir` is present. The measured split (one hook process: env-first reader →
P, cwd reader → W) therefore establishes that **readers in the same session
disagree**, but it does not establish which of P or W the review gates and the
receipt guard read. That value is unmeasured (remeasure "Gaps": the env value and
the stdin `cwd` were not separated).

### §1.2 Consequence for the design

Because the reader root is not measured — and is chosen by the host runtime, not by
MoAI — a design whose **store reachability** holds only if readers see `W`, or only
if they see `P`, rests on an unverified premise (the defect t1202 audit D18 raised).
This SPEC instead requires that writers and readers map any root they hold to one
store root, so the store is reached whichever of `P` or `W` a reader starts from.
Store reachability is the only property that claim covers: checks that compare a
record's tree identity with the reader's own tree remain dependent on the input
root, and §5 names the residual cases.

### §1.3 Current writers (code, `c630de892`)

| Writer | Destination | Tree identity on the record | Code site |
|---|---|---|---|
| Audit receipt (`codex_audit`, `audit_multi`) | `<root>/.moai/state/audit-receipts/receipts/` where `root` is the validated `project_root`, else the server fallback | yes (`tree_root`) | `internal/cli/mcp_audit_receipt.go:29-57`, `internal/auditreceipt/store.go:36-41, 197` |
| Auditor start marker (SubagentStart hook) | `<tree>/.moai/state/audit-receipts/starts/` | yes (`tree_root`) | `internal/hook/audit_receipt_guard.go:67-84`, `internal/auditreceipt/store.go:112-118, 227` |
| Auditor rejection (SubagentStop hook) | `<tree>/.moai/state/audit-receipts/rejections/<agent type>--<SPEC ID>.json` | **no** | `internal/hook/audit_receipt_guard.go:156-172`, `internal/auditreceipt/store.go:122-130, 252, 410-415` |
| `audit_multi` convergence state | `<root>/.moai/state/audit-multi/<session>.json` | **no** (keyed by session id only) | `internal/cli/mcp_convergence.go:768-770, 903-927` |
| `verify_snapshot` record / `moai verify record` | `<root>/.moai/state/verify/snapshots/<key>.json` | n/a (key is a content hash of the tree state) | `internal/cli/mcp_server.go:820-846`, `internal/cli/verify.go:63-77`, `internal/verify/store.go:15` |

### §1.4 What a shared store changes

Today every tree has its own store, so two records that lack a tree identity can
never meet. Routing a config-orphaned worktree's state into its primary checkout's
store makes records of `P`, `W`, and any sibling worktree `W2` share one directory.
Two record kinds then lose their meaning without a tree identity:

- **Rejections.** The receipt guard lists every rejection in its store
  (`ListRejections`, `internal/auditreceipt/store.go:270`) and, on a corroborated
  PASS, clears every rejection of that auditor role (`ClearRejectionsForRole`,
  `store.go:300`, called at `internal/hook/audit_receipt_guard.go:116`); the file
  name keys on agent type and SPEC ID only (`store.go:410-415`). In a shared store a
  PASS proven in `P` would clear `W`'s refusal (a silent guard bypass), `W`'s refusal
  would deny spawns in `P` (a false block), and `W2`'s refusal for the same SPEC
  would replace `W`'s. This is derived from the cited code, not executed.
- **Convergence results.** Commit `720149668` (card t182) moved the convergence
  result under the audited tree because "every worktree's verdict piled into the
  primary checkout" (`internal/cli/mcp_convergence.go:895-901`). A shared store keyed
  by session id alone reintroduces that: a later PASS for one tree overwrites an
  earlier FAIL for another, and the gate allows.

## §2 Definitions

- **Config-orphaned root** — exactly the predicate of SPEC-MCP-WORKTREE-UNTRACKED-001
  §4.4 (no `.moai/config/sections/workflow.yaml`, plus positive git-free evidence of
  a linked-worktree top level), implemented today as `isConfigOrphanedRoot`
  (`internal/cli/mcp_worktree_root.go:220`). This SPEC does not change it.
- **Primary checkout** — as identified by SPEC-MCP-WORKTREE-UNTRACKED-001
  REQ-MWU-004 (`identifyPrimaryCheckout`, `internal/cli/mcp_worktree_root.go:136`),
  including its fail-closed ambiguity rules.
- **Store root of a root R** — the directory whose `.moai/state` holds R's state:
  the primary checkout when R is config-orphaned and its primary is identifiable,
  R itself when R is not config-orphaned, and none (unresolved) when R is
  config-orphaned and its primary cannot be identified.
- **Tree identity** — the canonical path of the tree a record is about (for a
  worktree audit, canonical `W`). It is independent of the store root the record is
  kept in.
- **Phase-entry spawn** — an `Agent` spawn of `manager-develop`, `manager-docs`, or
  `manager-git` (`internal/hook/audit_receipt_guard.go:42-46`).

## §3 Requirements (GEARS)

### §3.1 One store root per root

- **REQ-WSR-001 (Ubiquitous, one answer everywhere).** For any root, every writer in
  §1.3 and every reader in §3.2 and §3.3 — whether it runs in the MCP server, a hook
  process, or the `moai verify` CLI — shall arrive at the same store root: the
  primary checkout for a config-orphaned root whose primary checkout is
  identifiable; unresolved (never a substituted root) for a config-orphaned root
  whose primary checkout cannot be identified; and the root itself for every root
  that is not config-orphaned, so that a primary checkout maps to itself.

- **REQ-WSR-002 (State-driven, writes on a config-orphaned root).** While a call's
  root is config-orphaned and its primary checkout is identifiable, the audit
  receipt, the auditor start marker and rejection, the `audit_multi` convergence
  state, and the `verify_snapshot` / `moai verify record` snapshot shall be written
  under the store root's `.moai/state`, and none of these writes shall create any
  path under the worktree's `.moai`.

- **REQ-WSR-003 (Ubiquitous, tree identity in a shared store).** Every audit
  receipt, auditor start marker, auditor rejection, and `audit_multi` convergence
  result shall carry the tree identity of the tree it was written for — for a
  config-orphaned root, the canonical worktree path, never the store root. Records
  of different trees in one store shall coexist: a rejection or a convergence result
  written for one tree shall not replace one written for another tree, even for the
  same auditor role, SPEC ID, or session. The receipt guard shall list, and on a
  corroborated PASS shall clear, only rejections whose tree identity equals the tree
  it resolved; a rejection that carries no tree identity (written before this
  change) shall be treated as belonging to the store root's own tree.

- **REQ-WSR-004 (Event-driven, primary not identifiable on write).** When a call's
  root is config-orphaned and its primary checkout cannot be identified, the state
  write shall not fall back to the worktree or to any other root: `verify_snapshot`
  record shall return a tool error naming the cause; the receipt and convergence
  writes of `codex_audit` and `audit_multi` — best-effort today — shall be skipped,
  the audit verdict shall be unchanged, and the tool result shall carry a notice
  naming the skipped write and its cause.

- **REQ-WSR-005 (Ubiquitous, other roots unchanged).** For a root that is not
  config-orphaned, every writer and reader named in this SPEC shall read and write
  at exactly the paths it uses today and shall run no git subprocess it does not
  run today. The only permitted differences are the tree-identity field REQ-WSR-003
  adds to rejections and convergence results, and — for a primary checkout that is
  the store root of config-orphaned worktrees — the reads of those worktrees'
  records that REQ-WSR-003 and REQ-WSR-007 require.

### §3.2 Paired readers

- **REQ-WSR-006 (Ubiquitous, paired reads).** `verify_snapshot` load, `verify_trend`,
  `moai verify record` / `moai verify check`, the `multi-review-gate` convergence
  read, and the receipt guard's receipt, start-marker, and rejection reads and
  writes shall use the store root of the root they resolved. Where that store root
  is unresolved, the MCP and CLI reads shall return an error naming the cause, and
  the hook readers shall follow REQ-WSR-008 and REQ-WSR-010; no reader shall read
  another root's state in its place.

- **REQ-WSR-007 (Event-driven, reader independence).** When `audit_multi` has
  persisted a convergence result for session S on a config-orphaned `W`, the
  `multi-review-gate` hook for session S shall find that result whether its input
  resolves to `W` or to the primary checkout `P`; and when the store root holds
  convergence results of session S for more than one tree, the gate shall block if
  any of them has overall verdict `fail`.

- **REQ-WSR-008 (State-driven, review-gate opt-in flag).** While the root the
  `codex-review-gate` or `multi-review-gate` hook resolved is config-orphaned, the
  hook shall read its opt-in flag (`workflow.codex.review_gate.enabled`,
  `workflow.multi.review_gate.enabled`) from the primary checkout's workflow
  config; when that primary cannot be identified the gate shall stay disabled, in
  keeping with these gates' fail-open contract. The order in which the hooks pick
  their root from the stdin payload and environment, and the tree the codex review
  targets, shall not change.

### §3.3 Receipt guard

- **REQ-WSR-009 (State-driven, guard gate from primary).** While the tree the
  auditor receipt guard resolved is config-orphaned, the guard shall read
  `workflow.audit.gates.codex` from the primary checkout's workflow config, with the
  existing explicit-value semantics (only a literal `required` opts in).

- **REQ-WSR-010 (Event-driven, guard fail-closed without a store).** When the
  guard's tree is config-orphaned and its primary checkout cannot be identified, the
  guard shall treat the codex gate as `required` and, having no store root, shall
  write no start marker or rejection record under any root; it shall refuse an
  auditor PASS at a first stop, shall emit the violation notice without blocking at
  a re-entrant stop (stop hook already active), and shall deny every phase-entry
  spawn whose tree is that worktree. Each refusal, notice, and denial reason shall
  state that the gate was assumed `required` because the primary checkout could not
  be identified, so it is distinguishable from a primary that declares `required`.

### §3.4 Catalogue on a config-orphaned root

- **REQ-WSR-011 (State-driven, union catalogue).** While a catalogue tool's
  (`spec_progress`, `spec_drift`, `spec_audit`) root is config-orphaned and its
  primary checkout is identifiable, the tool shall answer over the union of the
  worktree's `.moai/specs` and the primary checkout's `.moai/specs`, and each
  returned SPEC record or finding shall name its catalogue source (`worktree` or
  `primary`).

- **REQ-WSR-012 (Event-driven, same SPEC in both).** When one SPEC ID is present in
  both catalogues, each catalogue tool shall report the worktree copy, shall name
  the shadowed primary copy in the response, and shall count that SPEC once.

- **REQ-WSR-013 (Unwanted behaviour, no silent vanish).** The catalogue tools shall
  not return a successful answer on a config-orphaned root that omits a SPEC
  directory present under the worktree's `.moai/specs`.

- **REQ-WSR-014 (Event-driven, primary not identifiable on read).** When a catalogue
  tool's root is config-orphaned and its primary checkout cannot be identified, the
  tool shall answer over the worktree catalogue only and the `_root` block shall
  state that the primary catalogue was not read and why.

- **REQ-WSR-015 (Ubiquitous, `_root` provenance).** On a config-orphaned root, the
  `_root` block of each tool that carries one — `spec_progress`, `spec_drift`,
  `spec_audit`, `verify_snapshot`, and `verify_trend` — shall list the catalogue or
  state sources actually read, and its `worktree_warning` text shall no longer say
  the answer was read from the worktree tree only; the existing `_root.warning`
  fallback notice shall keep its key and text and shall never be overwritten.
  (`codex_audit` and `audit_multi` carry no `_root` block; their skipped writes are
  reported by the REQ-WSR-004 notice.)

### §3.5 Documentation

- **REQ-WSR-016 (Ubiquitous, documentation).** The "Linked worktrees of a repository
  that keeps `.moai` untracked" section of
  `.claude/rules/moai/core/moai-mcp-tools-catalogue.md`, the shared `project_root`
  parameter description (which today says "other configuration, the SPEC catalogue,
  and state are still read from the accepted tree" — `internal/cli/mcp_project_root.go:47-50`),
  and the tool descriptions shall state the new behaviour: the descriptions of
  `spec_progress`, `spec_drift`, and `spec_audit` that such a worktree's catalogue is
  the union of the worktree's and the primary checkout's; the descriptions of
  `verify_snapshot`, `verify_trend`, `codex_audit`, and `audit_multi` that such a
  worktree's state is kept in the primary checkout's `.moai/state` under the
  worktree's own tree identity; and the `codex_audit` description that the
  hook-side receipt guard reads the gate from the primary checkout (replacing the
  sentence "that refusal is not guaranteed" — `internal/cli/mcp_server.go:317`). The
  rule file and its template mirror under `internal/template/templates/` shall stay
  byte-identical, and the template copy shall carry no SPEC ID, card id, or date.

## §4 Constraints

- The config-orphaned predicate and primary identification are reused, not
  redefined; the scrubbed-git and `LC_ALL=C` rules of SPEC-MCP-WORKTREE-UNTRACKED-001
  REQ-MWU-006 apply to every primary identification this SPEC adds, including those
  in hook processes.
- Primary identification runs only for a config-orphaned root (REQ-WSR-005).
- The reject-never-fall-back contract of `project_root` is unchanged; REQ-WSR-004
  extends the same stance to state writes.
- Template text stays neutral across the 16 supported programming languages and free
  of internal development state.
- Hook handlers keep their existing timeouts and fail-open/fail-closed directions
  except where REQ-WSR-010 states otherwise.

## §5 Measurement gap carried into run

The stdin `cwd` / `project_dir` and the `CLAUDE_PROJECT_DIR` value delivered to a
hook process in a worktree session are unmeasured (§1.1). REQ-WSR-001, -006 and -007
make **store reachability** independent of them, and acceptance AC-WSR-006 measures
which root each Stop review gate reads for every combination of those three inputs.
What it cannot establish is which combination the host actually sends; the optional
M0 probe (plan.md §D) captures that without gating any criterion. Three residual
cases depend on the unmeasured value:

1. A receipt guard whose SubagentStart input names `P` while the audit ran on `W`
   records start-marker identity `P`; the `W` receipt is then refused as "another
   tree" (`internal/auditreceipt/store.go:381`) — a loud refusal, never a silent pass.
2. A SubagentStop input and a later PreToolUse input naming different trees make the
   spawn check read the other tree's rejections, so a refusal recorded for `W` does
   not deny a spawn whose input names `P`. This is silent, and it is today's
   behaviour as well: with per-tree stores on `c630de892`, a refusal recorded in one
   tree's store is equally invisible to a spawn check reading another tree's.
3. The same applies to REQ-WSR-003's filter in the other direction: a PASS proven in
   one tree never clears another tree's refusal, whatever the inputs.

## §6 Exclusions (What NOT to Build)

This section records what is out of scope for this SPEC.

### Out of Scope — configuration keys other than those REQ-WSR-008/009 route

- Audit pins, `llm.yaml`, and every configuration key other than
  `workflow.audit.gates.codex` (receipt guard) and the two review-gate opt-in flags.
  Those two flags are included only because without them the `multi-review-gate`
  never reaches the state read that REQ-WSR-007 requires to agree with the writer;
  the codex flag is kept with it so the two Stop gates, which share one input parser
  so that they "cannot drift" (`internal/cli/multi_review_gate.go:198`), keep one
  root rule.

### Out of Scope — which root the hooks resolve

- Changing how hooks derive their root from the stdin payload or the environment
  (`resolveProjectDirFromInput`, `resolveProjectRoot`, `TreeRootFromCWD`), and the
  tree the `codex-review-gate` reviews. The value Claude Code sends is unmeasured
  (§5); this SPEC makes store reachability independent of it instead.

### Out of Scope — adjacent items owned elsewhere

- Creation-side `.moai` provisioning of worktrees (SPEC-MCP-WORKTREE-UNTRACKED-001
  §2.2 design (b)).
- Scrubbing inherited `GIT_*` variables from tree-operation git invocations.
- `resolveProjectDir()` and the absent-`project_root` fallback
  (SPEC-MCP-WORKTREE-ROOT-001).
- Validator acceptance, the §4.4 predicate, and the MCP-side gate routing, owned by
  SPEC-MCP-WORKTREE-UNTRACKED-001.

### Out of Scope — existing state and other surfaces

- Migrating, reading through, or deleting state already written under a worktree's
  `.moai/state` by the interim behaviour; it is left in place and no longer read.
- Worktrees of a repository that tracks `.moai` (not config-orphaned; unchanged by
  REQ-WSR-005).
- The web dashboard's verify view (`internal/web/viewmodel_ops.go`) and every other
  `.moai/state` family not listed in §1.3 (goal state, session registry, lessons,
  graph).

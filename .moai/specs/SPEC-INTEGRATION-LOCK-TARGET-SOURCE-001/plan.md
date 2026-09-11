---
id: SPEC-INTEGRATION-LOCK-TARGET-SOURCE-001
title: "Implementation plan — integration lock target provenance and card visibility (card t637)"
version: "0.2.0"
created: 2026-09-11
updated: 2026-09-11
author: manager-spec
priority: P2
phase: "v3.2.0 target"
module: "internal/cli, internal/kanban, internal/config"
lifecycle: spec-anchored
tags: "integration-lock, kanban, factory, git-flow, observability, t637"
tier: M
---

# Implementation Plan — SPEC-INTEGRATION-LOCK-TARGET-SOURCE-001

Sections are ordered by how likely each decision is to change: the user-facing output shapes and
the warning channel first (§B), then the record-shape decision (§C), the configuration seam
(§D), the documentation decisions (§E), and only then the milestones (§F), which end with the
mechanical steps.

## §A Context

- Card **t637**, worktree `.claude/worktrees/t637`, branch `WT-acquire-branch-record`.
- Evidence base: `.moai/reports/t637/verdict.md` (fixture cells C1-C4, code reading, gaps).
- Trees: plan authored on HEAD `1ad0fdc09` (verdict follow-up on top of `527ecad6b`); base
  `4c99d973e`. `git diff --stat 4c99d973e origin/develop` over every SPEC target file
  (`internal/cli/integration.go`, `internal/cli/integration_target_test.go`,
  `internal/cli/integration_lock_cli_test.go`, `internal/kanban/integration_lock.go`,
  `internal/config/loader_integration_branch.go`, both kanban-dispatch copies,
  `.claude/rules/local/gitflow-lane-protocol.md`, `.claude/agents/harness/hns-release-specialist.md`,
  `CLAUDE.local.md`, `internal/template/catalog.yaml`, plus
  `internal/config/loader_integration_branch_test.go`) is empty against `origin/develop`
  `ee99507fb`, and `git diff --stat 4c99d973e HEAD` over the same paths is empty as well (both
  re-run by manager-spec in this plan pass: no output, exit 0) — no absorb is needed for
  plan-phase. Run-phase absorbs local `develop` per the lane protocol before its window.
- **Traceability note (card-id collision):** `develop` carries commit `4e192ee82`
  ("merge: workflow audit F31 into develop (card t637)"). It is NOT this card's work — a card-id
  number collision. This card's commits live only on `WT-acquire-branch-record`.
- The production symptom is currently masked (the primary checkout's `git-strategy.yaml` manual
  block was filled by the lead). No acceptance criterion relies on that configuration.

## §B Decisions most likely to change — output shapes and the warning channel

### B1. Warning channel — standard error (DECIDED)

The REQ-ILT-006 warning goes to **standard error**, in both text and `--json` modes. Reasons:

1. Standard output is the command's result channel. `acquire --json` must stay one parseable JSON
   object, and the text-mode `release-integration window acquired by … on …` line is matched by
   existing tests and read by lanes; a warning interleaved there breaks both.
2. Standard error is the integration-lock family's established advisory channel — the guard's
   `[moai:integration-lock] advisory:` lines already go there.
3. Visibility is not lost: a terminal shows both streams, and the Claude Code Bash tool captures
   both, so a lane running `acquire` sees the line.

Prefix: `[moai:integration-lock] warning:` (same family as the guard's advisory prefix). The line
names the recorded caller branch and both remedies (`git_strategy.manual.develop_branch`, or
`--branch`). Exact wording beyond the prefix and those three tokens is run-phase latitude.

Writer: the line is written through the command's error writer (`cmd.ErrOrStderr()`), not
`os.Stderr` directly, so a test that sets `SetErr` receives it. Timing: it is written only after
`AcquireIntegrationLock` has succeeded — a refused acquire (window held) returns its error and
emits no warning (spec.md REQ-ILT-006; acceptance AC-ILT-008).

Test implication: the existing `runIntegration` helper merges both streams into one buffer
(`integration_lock_cli_test.go:26-43`), so it cannot tell a stderr warning from a stdout one. The
warning tests need a stream-separating helper; otherwise the "warning moved to stdout" mutation
survives (AC-ILT-014 row e).

### B2. `status` text shape (DECIDED)

- **Card line:** `  card:     <id>` (label padded to the existing 10-column label width), printed
  between the `holder:` and `branch:` lines, only when the record carries a card.
- **Provenance:** a suffix on the branch line — `  branch:   <branch> (source: <flag|config|caller>)`
  — printed only when the record carries `branch_source`. A record without it prints today's
  exact line. The suffix form keeps every existing substring check on `branch:   <x>` valid and
  ties the provenance to the value it qualifies.

Alternative rejected: a separate `source:` line — ambiguous label (source of what?) and one more
line for a reader to associate.

### B3. What counts as "git-flow" for the warning (DECIDED)

The same predicate the configured tier already uses: `mode == manual` AND active profile
`workflow == git-flow`. A personal/team profile carrying `workflow: git-flow` does not qualify,
exactly as it does not qualify for the configured tier today. A git strategy file that is absent
or unreadable is "not git-flow" (no warning) — the tool cannot assert a misconfiguration it
could not read.

## §C Record shape — `branch_source`

- New optional field on `kanban.IntegrationLock`: JSON key `branch_source`, `omitempty`, values
  `flag` | `config` | `caller`. Written by `acquire` on every new record.
- A record written before the field exists decodes with an empty value; no read path branches on
  it except the `status` printer (B2). The field is additive in the same way `pid_source` was.
- The provenance is decided where the branch is decided: the resolution function reports which
  tier won, so the source can never disagree with the recorded value. Keeping the value computed
  in one place is what makes AC-ILT-014 row d (source hard-coded) observable. The source is
  decided from the TRIMMED flag value, the same value that decides the branch (row h).
- Signature change allowance: if the resolution function gains a return value, the four existing
  call sites in `integration_target_test.go` (`TestResolveIntegrationTarget_*`, lines 134, 147,
  160, 173 on `1ad0fdc09`) are adapted mechanically (for example `branch, wt, _ :=`). That
  adaptation is not an edit to their assertions; run-phase lists it in `progress.md` §E.2.
- The guard (`internal/hook/integration_lock_guard.go`) is not edited (REQ-ILT-013).

## §D Configuration seam

`LoadGitFlowDevelopBranch` returns `""` for three cases the warning must tell apart (spec.md §B
premise 2). Run-phase adds a way to learn "the project is git-flow" separately from "the develop
branch value", in `internal/config/loader_integration_branch.go`, without changing the existing
function's contract: the existing test functions in `loader_integration_branch_test.go` (notably
`TestLoadGitFlowDevelopBranch`) are not modified, and the seam's new tests are ADDED to that file
(acceptance CMD-ILT-016 runs both). Shape
(second return value, sibling predicate, or small struct) is run-phase latitude; the constraint
is one file read per `acquire` and no new exported behavior beyond the predicate.

## §E Documentation decisions

### E1. Invocation edits (line-local, placeholder `<card-id>`)

| File | Line (HEAD `1ad0fdc09`) | Edit |
|---|---|---|
| `.claude/rules/moai/workflow/kanban-dispatch.md` | 230 | `` `moai integration acquire --name <lane>` `` → `` `moai integration acquire --name <lane> --card <card-id>` `` |
| `internal/template/templates/.claude/rules/moai/workflow/kanban-dispatch.md` | 230 | identical edit — the two copies are byte-identical today and must stay so |
| `.claude/rules/local/gitflow-lane-protocol.md` | 47 | prose re-acquire mention → `` `moai integration acquire --name <lane> --card <card-id>` `` |
| `.claude/rules/local/gitflow-lane-protocol.md` | 51 | command → `moai integration acquire --name <lane> --card <card-id>` |
| `CLAUDE.local.md` | 370 | prose chain `` `moai integration acquire` `` → `` `moai integration acquire --card <card-id>` `` |
| `CLAUDE.local.md` | 389 | `` `moai integration acquire --name <lane>` `` → append ` --card <card-id>` |
| `CLAUDE.local.md` | 403 | command → append ` --card <card-id>` |
| `.claude/agents/harness/hns-release-specialist.md` | 345 | no `--card`; extend the trailing comment to state that a release integration delivers no card |

### E2. Release invocation — no card (DECIDED)

The release back-merge (`main` → `develop`) delivers no card: the kanban-dispatch PR-title rule
already classifies release, batch, and maintenance work as card-less. Inventing a card id would
make the record assert a card that does not exist. The comment makes the exception explicit so a
later sweep for "every acquire carries `--card`" does not "fix" it.

### E3. CLAUDE.local.md conflict risk

The primary checkout carries **uncommitted foreign modifications** to `CLAUDE.local.md`: its
working copy has 734 lines with the acquire mentions at 338/357/371, while the tracked file in
this worktree has 772 lines with them at 370/389/403. The two edits touch different line
positions of a file that has diverged, so a later merge or commit of the primary's working copy
may conflict on these three lines. Mitigation: keep the three edits strictly line-local (append a
flag, change nothing else on the line, no reflow), and name the risk in the completion report so
the lead reconciles it when the primary's changes are committed. This lane does not touch the
primary checkout's copy.

Where the conflict surfaces: the primary checkout has `main` checked out, and `main`'s blob of
`CLAUDE.local.md` is 603 lines (plan-audit iter1 measurement), so the primary's 734-line copy is
uncommitted work on top of `main`, while this worktree's 772-line copy equals the local `develop`
blob. The collision therefore appears when that uncommitted work is committed on `main`, or when
the release PR brings `develop`'s copy to `main` — not at this card's `develop` merge.

### E4. Template build and catalog

- `internal/template/catalog.yaml` has **no entry** covering the kanban-dispatch rule — it lists
  skill directories only (`grep -n "rules" internal/template/catalog.yaml` printed nothing on
  `1ad0fdc09`). No `gen-catalog-hashes.go` run is required. If run-phase finds an entry after
  absorbing `develop`, regenerate ONLY that entry with
  `go run ./internal/template/scripts/gen-catalog-hashes.go --entry <NAME>` (never `--all`).
- After the template edit, `make build` re-embeds the template (Template-First rule).
- The kanban-dispatch rule is always-loaded; editing it mid-session invalidates the prompt-cache
  prefix, so the doc milestone runs last (cache-aware directive 3).

## §F Milestones (priority-ordered)

**M1 — Failing tests first (Priority High).** Add the tests that pin every behavioral REQ before
any production edit, and record their RED output on `1ad0fdc09` in `progress.md` §E.2:
source recorded (flag / blank flag / config / caller), git-flow-empty warning on stderr, no
warning for github-flow with an EMPTY develop branch / non-manual mode with git-flow and an EMPTY
develop branch / absent config / flag / config, no warning on a refused acquire, card line present
and absent, provenance suffix, old-record status, `--branch` help text. The test names are the
ones acceptance.md §D.0 lists. Tests live in `internal/cli/integration_target_test.go`
(resolution + warning) and `internal/cli/integration_lock_cli_test.go` (status + help), plus
`internal/config/loader_integration_branch_test.go` for the §D seam. Scratch repositories use
`t.TempDir()`; every `runIntegration` call pins `CLAUDE_PROJECT_DIR`.

**M2 — Configuration seam (Priority High).** §D.

**M3 — Record and resolution (Priority High).** §C field; resolution reports its tier; `acquire`
writes the source and emits the B1 warning under the B3 predicate.

**M4 — Status printer (Priority High).** B2 card line and provenance suffix.

**M5 — Flag help (Priority Medium).** REQ-ILT-009 wording: "The integration target branch the
merge lands on, not the card branch being merged (default: the configured git-flow develop
branch, else the current branch)". The phrase `integration target` appears in lowercase, and
acceptance AC-ILT-010 checks it case-SENSITIVELY; any rewording must keep that exact lowercase
phrase and must not reintroduce "Branch being integrated".

**M6 — Fixture controls and mutation guard (Priority High).** In the lead-gated compile slot,
build `/tmp/t637-fx-bin/moai` from this worktree (acceptance.md §A.4), run fixture cells C1′, C2′,
C3′ (CMD-ILT-011..013) and the mutation rows a-h (§D.3); evidence lands in
`.moai/reports/t637/ac-evidence/` and is cited in `progress.md` §E.2.

**M7 — Documentation (Priority Medium, mechanical, last).** §E1 edits, `make build`, the
AC-ILT-011/012 checks.

## §G File count and Tier

| Kind | Files |
|---|---|
| Production code | `internal/cli/integration.go`, `internal/kanban/integration_lock.go`, `internal/config/loader_integration_branch.go` (3) |
| Tests | `internal/cli/integration_target_test.go`, `internal/cli/integration_lock_cli_test.go`, `internal/config/loader_integration_branch_test.go` (3) |
| Documentation | `.claude/rules/moai/workflow/kanban-dispatch.md`, its template mirror, `.claude/rules/local/gitflow-lane-protocol.md`, `.claude/agents/harness/hns-release-specialist.md`, `CLAUDE.local.md` (5) |
| Catalog | `internal/template/catalog.yaml` — not applicable (no entry, §E4) (0) |

**Total: 11 files.** Estimated change: production ~60-120 LOC, tests ~200-320 LOC, docs ~8 lines.
11 files falls in the Tier M band (5-15 files); LOC is at the Tier S/M boundary, and the file
count decides. **Verdict: Tier M** — artifact set spec.md + plan.md + acceptance.md (+ progress.md).

The coordinator's framing counted "6 tracked doc files incl. the template mirror"; the measured
distinct doc files are 5 (8 edit sites across them). The count above is the measured one.

## §H Risks

| Risk | Mitigation |
|---|---|
| Existing tests assert the exact status text shape | B2 keeps today's shape for card-less, source-less records; the suffix keeps `branch:   <x>` substrings valid |
| The warning fires in legitimate caller-fallback projects | B3 predicate + AC-ILT-007 negative cells (github-flow, absent config) + AC-ILT-014 row b |
| Warning lands on stdout and breaks JSON consumers | B1 + stream-separating test helper + AC-ILT-008 + AC-ILT-014 row e |
| Source disagrees with the recorded branch | §C: source reported by the same resolution step that decides the branch |
| Fixture run touches the real window | `CLAUDE_PROJECT_DIR=/tmp/t637-fx` in the same invocation; real lock file hash compared before/after (AC-ILT-005) |
| CLAUDE.local.md divergence with the primary's uncommitted copy | §E3 |
| Always-loaded rule edit busts the session cache | doc milestone last (§E4) |

## §I Anti-patterns to avoid

- Editing the guard to "use" `branch_source` — it must stay non-decisional (REQ-ILT-013).
- Refusing or gating `acquire` on the fallback — the operator decision is warn-only.
- Rewriting existing lock records or giving old records a synthetic `branch_source`.
- `go test ./...` locally, or `gen-catalog-hashes.go --all`.
- Asserting the warning through the merged-stream helper (it cannot tell stderr from stdout).

## §J Cross-references

- `.moai/reports/t637/verdict.md` — evidence base, fixture cells C1-C4
- `.moai/specs/SPEC-INTEGRATION-LOCK-LIVENESS-001/` — `pid_source` precedent for an additive, `omitempty` provenance field
- `.moai/specs/SPEC-INTEGRATION-LOCK-ATOMIC-001/` — sibling lock SPEC (mutation atomicity)
- `.claude/rules/moai/workflow/kanban-dispatch.md` § Integration into the release branch is self-served

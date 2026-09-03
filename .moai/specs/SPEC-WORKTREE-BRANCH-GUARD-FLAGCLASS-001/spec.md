---
id: SPEC-WORKTREE-BRANCH-GUARD-FLAGCLASS-001
title: "Main-Checkout Branch-State Guard — git branch Mutation-vs-Query Flag-Class Completion"
version: "0.3.0"
status: draft
created: 2026-09-03
updated: 2026-09-03
author: manager-spec
priority: High
phase: "v3.1.5"
module: internal/hook
lifecycle: spec-anchored
era: V3R6
tier: M
tags: "hook, branch-guard, git-branch, flag-class, under-match, bug-fix"
depends_on:
  - SPEC-WORKTREE-BRANCH-GUARD-001
  - SPEC-WORKTREE-BRANCH-GUARD-OPTIN-001
  - SPEC-WORKTREE-BRANCH-GUARD-DISCRIM-001
related_specs:
  - SPEC-WORKTREE-BRANCH-GUARD-001
  - SPEC-WORKTREE-BRANCH-GUARD-OPTIN-001
  - SPEC-WORKTREE-BRANCH-GUARD-DISCRIM-001
---

# SPEC-WORKTREE-BRANCH-GUARD-FLAGCLASS-001 — git branch Mutation-vs-Query Flag-Class Completion

## §A. Problem Statement

The main-checkout branch-state guard's `git branch` pattern —
`internal/hook/branch_guard.go:120`, pattern
`\bgit\s+branch\s+(-[dDmMcC]\s+)?[^\s-]` — matches branch mutation via a
single-character flag class `[dDmMcC]` followed by the "non-flag token after
subcommand" rule `[^\s-]`. Two consequences:

1. **Under-match (the defect).** Every mutating flag form outside the
   character class passes through the guard un-denied in the primary
   checkout. **Measured** on tree `d592b0551` (plan-phase synthetic probe
   driving `matchBranchStateCommand`; probe deleted after measurement, full
   verbatim output in acceptance.md §D.0):

   | Command form | matched (pre-fix) |
   |---|---|
   | `git branch -f topic abc1234` | false |
   | `git branch --force topic abc1234` | false |
   | `git branch -df old` / `-fm renamed` / `-vD old` (combined short flags) | false |
   | `git branch -u origin/main topic` | false |
   | `git branch --set-upstream-to=origin/main topic` | false |
   | `git branch --unset-upstream topic` | false |
   | `git branch -t topic origin/main` | false |
   | `git branch --edit-description topic` | false |

   `git branch -f <branch> <sha>` is a branch-pointer rewrite against the
   shared primary checkout — exactly the hazard class the guard exists to
   deny — and it is ALLOWED through. Discovered by synthetic measurement
   during card t458 (doc-fix card, landed on develop).

2. **Over-match axis is clean in current source (measured, same probe).**
   `git branch --list develop -v`, `--merged main`, `--no-merged main`,
   `--points-at HEAD`, `--format %(refname)` all return `matched=false`
   (allowed), consistent with the existing pinned read-only test set at
   `internal/hook/branch_guard_test.go:308-320` (`-v`, `-vv`, `-a`, `-r`,
   bare, `--list`, `--show-current`, `--contains HEAD`). The historical
   friction recorded in CLAUDE.local.md §4.1 (read-only `git branch --list` /
   `-vv` denied) does NOT reproduce against this source — the source comment
   at `branch_guard.go:90-106` records a pickaxe finding that no committed
   revision ever carried an undiscriminating `git branch` pattern, and the
   guard additionally ships opt-in-disabled by default
   (`Workflow.BranchGuard.Enabled`, distributed default false). The stale
   friction note is therefore NOT evidence of a live over-match; M1's
   measurement matrix is the authoritative word on both axes.

### Value statement

t458 landed doctrine stub v1.3.2 at
`.claude/rules/moai/workflow/main-checkout-branch-guard.md`: the forbidden
table names only MUTATION forms, and a Query-vs-mutate discrimination bullet
permits read-only `git branch` queries (bare list, `--list`, `-v`/`-vv`,
`--show-current`, `--contains`/`--merged`/`--points-at`). The CODE already
implements the query half of that discrimination but not the full mutation
half. This SPEC EXTENDS the mutation class beyond the doctrine's literal
v1.3.2 flag enumeration: the v1.3.2 Query-vs-mutate bullet itself states
that `-f`/`-u` force/upstream forms "sit outside the flag class and pass …
extending the flag class is a policy change, not a doc correction"
(`main-checkout-branch-guard.md:90-95`). This SPEC makes that policy
change — extending the mutation class per the forbidden table's
shared-state rationale, with the M1 measurement (not the table's literal
text) as the classification authority — and lands the matching bounded
doctrine update in the same card (REQ-WBG-F-009) so doc and code share one
discrimination, closing the under-match axis without reopening the
over-match axis.

**Doctrine disposition (audit iteration 1, D1):** resolved via the
bounded-update path — M3 adds the newly-covered mutation forms to the
v1.3.2 Query-vs-mutate bullet and forbidden-table row 2's flag
enumeration. The alternative (a named residual + follow-up card) was
rejected: leaving the doctrine asserting "`-f`/`-u` … pass" while the
guard denies them recreates exactly the doc-code shared-discrimination
break the dispatching mandate named, and the edit is one bullet plus one
table row — proportionate to the card.

## §B. Scope

### In Scope

1. **The `git branch` branch-state pattern only** (`branch_guard.go:120`):
   widen/distinguish the flag discrimination so every MUTATION form of
   `git branch` denies in the primary checkout and every QUERY form stays
   allowed — EXTENDING the v1.3.2 mutation class per its shared-state
   rationale (the M1 measurement is the classification authority), not
   merely re-implementing its literal six-flag enumeration.
2. **Full-flag-token classification.** Long flags classified by their
   complete token, never by substring — `--format` ≠ `--force`, and
   `--contains` / `--merged` / `--no-merged` must not be denied via the
   `c`/`m` letters embedded in their names.
3. **Synthetic measurement harness (M1).** A permanent Go table test driving
   the guard's matcher with a pinned expected matrix covering both axes; the
   fix class is decided only after this matrix exists and has been run.
4. **Condition-documenting tests (M3).** Opt-in gate (`Workflow.BranchGuard.Enabled`
   distributed default false; tests enable it explicitly) and both exemption
   axes' unreachability from tool-spawned subagents, encoded as documented
   test conditions so future readers do not misread a guard bypass.
5. **Preserved surfaces**: `d/D/m/M/c/C` + bare-name denial, deny sentinel
   prefix `BRANCH_GUARD_VIOLATION:`, primary-vs-worktree discriminant,
   fail-open semantics (deny only on positive evidence; uncertainty → allow
   + audit-log append).
6. **Bounded doctrine alignment edit (M3, audit iteration 1 D1)**: update
   the v1.3.2 Query-vs-mutate bullet
   (`main-checkout-branch-guard.md:90-95`) and forbidden-table row 2's
   flag enumeration to name the newly-denied mutation forms —
   Template-First (template copy first, local mirror in the same commit,
   sanitized-pair parity) — so doc and code share one discrimination.

### Out of Scope — Non-branch patterns

- The `git switch`, `git checkout`, `git reset --hard`, `git stash`,
  `git rebase`, `git merge` patterns in the same `branchStatePatterns` set
  are NOT touched. Their behavior stays byte-identical.
- `git checkout <file>` single-file-restore lexical collision (documented
  residual at `branch_guard.go:72-76`) is NOT re-litigated.

### Out of Scope — Guard redesign

- No new exemption surface, no discriminant change, no audit-log path change,
  no latency-budget change. The fix is the `git branch` pattern (or an
  equivalent token-level discrimination replacing that single pattern entry)
  plus tests.
- Shell-wrapped / obfuscated git invocations (`bash -c "git branch -f …"`)
  remain under-matched — the documented correct direction for a fail-open
  guard.

### Out of Scope — Doctrine changes beyond the bounded alignment bullet

- The M3 doctrine edit (REQ-WBG-F-009) is bounded to the Query-vs-mutate
  bullet and forbidden-table row 2's flag enumeration. Restructuring the
  permitted list, adding new sections, renumbering the rule, or any
  detail-companion rewrite is out of scope. If M1's measurement contradicts
  the v1.3.2 permitted list, a blocker report — not a broader doctrine
  edit — is the escalation path.

### Out of Scope — CLAUDE.local.md friction-note hygiene

- Updating or annotating the CLAUDE.local.md §4.1 historical friction note
  is not part of this card; the M1 matrix output is the artifact a future
  card would cite.

## §C. Requirements (GEARS notation)

### REQ-WBG-F-001 — Class-extended flag discrimination (Ubiquitous)

The `git branch` branch-state guard SHALL classify command forms by a
mutation-vs-query flag discrimination that EXTENDS the v1.3.2 doctrine
stub's mutation class per its shared-state rationale (the v1.3.2 bullet's
own literal enumeration covers only `-d/-D/-m/-M/-c/-C` + bare creation
and explicitly leaves `-f`/`-u` outside): a form is denied in the primary
checkout if and only if it presents a mutating flag or a bare branch-name
creation operand, and a form presenting only query flags is allowed. The
M1 measurement matrix is the classification authority for every form.

### REQ-WBG-F-002 — Mutation forms denied (Event-detected)

**When** a `git branch` command presents any mutating flag — at minimum
`-f`, `--force`, `-d`, `--delete`, `-D`, `-m`, `--move`, `-M`, `-c`,
`--copy`, `-C`, `-u`, `--set-upstream`, `--set-upstream-to`,
`--unset-upstream`, `-t`, `--track`, `--no-track`, `--edit-description`,
or any short-flag cluster beginning with `u` (the attached-value spelling
`-u<upstream>`) or containing one of `d/D/m/M/c/C/f/t` — the guard SHALL
deny it in the primary checkout with the `BRANCH_GUARD_VIOLATION:`
sentinel prefix. An `=`-attached long-flag value
(`--set-upstream-to=<upstream>`, `--track=<mode>`) SHALL classify by the
flag name before the `=`. A non-flag operand SHALL be classified as a
creation (deny) when no list action is selected and the operand is not
consumed as the value of a preceding value-taking flag — covering
option-prefixed creation forms (`git branch -- <name>`,
`git branch -v <name>`, `git branch -q <name>`, `git branch --no-force
<name>` — all auditor-measured creating branches on git 2.50.1) and
creation modifiers (`--recurse-submodules`, `--create-reflog`) appearing
alongside a creation operand. Query flags that select a list action
(`--list`, `--contains`, `--merged`, `--no-merged`, `--points-at`,
`--format`, `--show-current`) consume their pattern operands without
triggering creation. A bare branch-name creation operand
(`git branch <name>`, `git branch <name> <start-point>`) SHALL continue to
deny.

### REQ-WBG-F-003 — Query forms allowed (Event-detected)

**When** a `git branch` command presents only query flags — bare `git branch`,
`--list`, `-v`, `-vv`, `-a`, `-r`, `--show-current`, `--contains`,
`--merged`, `--no-merged`, `--points-at`, `--format`, `--sort`, `-q`,
`-i` — the guard SHALL allow it in the primary checkout.

### REQ-WBG-F-004 — Whole-token long-flag classification (Ubiquitous)

The guard SHALL classify each long flag by its complete token, never by a
substring or an embedded-letter character class, such that `--format` is
never classified as `--force`, and `--contains` / `--merged` / `--no-merged`
are never denied via the `c` / `m` letters embedded in their names.

### REQ-WBG-F-005 — Preserved surfaces (State-driven)

**While** the guard evaluates any command, the existing covered mutation
forms (`git branch <name>`, `-d`, `-D`, `-m`, `-M`, `-c`, `-C`), the deny
sentinel prefix `BRANCH_GUARD_VIOLATION:`, the primary-vs-worktree
discriminant (git-dir vs git-common-dir at the command's actual cwd), the
fail-open contract (deny only on positive evidence; any uncertainty → allow
+ audit-log append to `.moai/logs/branch-guard-audit.log`), and every
non-`branch` pattern in `branchStatePatterns` SHALL remain byte-identical in
behavior.

### REQ-WBG-F-006 — Opt-in gate unchanged (Where capability gate)

**Where** `Workflow.BranchGuard.Enabled` is false (the distributed default),
the guard SHALL perform no pattern evaluation and no deny — the opt-in gate
and its disabled-path cost profile are unchanged from
SPEC-WORKTREE-BRANCH-GUARD-OPTIN-001.

### REQ-WBG-F-007 — Measurement-first sequencing (Ubiquitous)

The fix class SHALL be decided only after the M1 synthetic measurement
matrix (a Go test fixture driving the guard's matcher with representative
command strings and pinned expected classifications) exists and has been
executed on the pre-fix tree; no flag form's classification may rest on
estimation.

### REQ-WBG-F-008 — Exemption-axes conditions documented (Ubiquitous)

The test suite SHALL encode as documented conditions (not code changes)
the reachability of both exemption axes from a tool-spawned subagent. The
`MOAI_BRANCH_GUARD_EXEMPT=1` axis is uncontested: the sentinel is read
from the hook process's own environment (spawned before the guarded
command runs, so exporting it inside the command is a no-op). The
`AgentType` axis is CONTESTED by a pre-existing repo SSOT contradiction:
`branch_guard.go:30-33` and t43 runtime observations hold that `AgentType`
is populated only for a main-thread `claude --agent <name>` launch, while
`.claude/rules/moai/core/hooks-system.md:114` states "All hook events
include `agent_id` and `agent_type` fields when triggered from a subagent
context (v2.1.69+)". The SPEC asserts NEITHER side as fact: M3 SHALL
capture one real tool-spawned PreToolUse payload to measure which holds,
and a capture contradicting the guard's reading becomes a
doc-reconciliation blocker report — never a silent re-classification. The
synthetic negative-path test (no `AgentType`, env unset → deny stands)
pins the guard LOGIC only and is explicitly labeled as not proving the
payload shape.

### REQ-WBG-F-009 — Bounded doctrine alignment (Ubiquitous)

The doctrine rule `.claude/rules/moai/workflow/main-checkout-branch-guard.md`
SHALL, by M3 completion, describe the extended mutation class — its
Query-vs-mutate bullet (lines 90-95) and forbidden-table row 2's flag
enumeration name the newly-denied forms — so the doctrine no longer
asserts that `-f`/`-u` forms pass. The edit SHALL land Template-First,
with the sanitized-pair mirror
(`internal/template/templates/.claude/rules/moai/workflow/main-checkout-branch-guard.md`)
kept in parity in the same commit.

## §D. Constraints

- **Tier M** (3-artifact set: spec / plan / acceptance, plus progress.md
  as the all-tier lifecycle record) — the dispatch labeled the card Tier S
  but mandated acceptance.md with the full expected measurement matrix; the
  artifact-set tier is therefore M with Tier-S scope discipline (matcher +
  tests + the bounded doctrine bullet only).
- **Class B defect card (kanban)**: cause established by synthetic
  measurement during t458; the run phase lands the fix, not a redesign.
- **verification-claim-integrity + verification-completeness**: every AC
  carries a RED-now cell or an explicit green-now characterization cell —
  AC-001/002/004/009 carry RED-now cells (pinned to tree `d592b0551` via
  acceptance.md §D.0, or to the audit-iteration-1 git 2.50.1 live probes,
  or — for the REQ-WBG-F-009 doctrine assertion — to the v1.3.2 file text
  itself); AC-003/005/006/007/008 are green-now characterization pairs.
  Cells not yet measured are labeled inferred-pending-M1 and are measured
  by the M1 run before the fix class is decided; no classification claim
  rests on estimation.
- **Fail-open doctrine** (Bash Risk-Amplifier WARN-ONLY, FAIL-OPEN): no new
  blocking on uncertainty paths.
- **Go code/comments in English**; table-driven tests; package
  `internal/hook` conventions (t.TempDir isolation, no OTEL env in parallel
  tests).
- **Verification scope**: affected packages only (`go test ./internal/hook/...`)
  + `golangci-lint`; no local full suite (lane-load discipline).

## §E. Acceptance Criteria (summary — full Given-When-Then + matrix in acceptance.md)

| AC | Subject | Verifiable by |
|----|---------|---------------|
| AC-WBG-F-001 | Full expected matrix: every mutation cell (M-01..M-28) → deny, every query cell → allow (§D.1 table) | M1/M3 Go table test over the matcher + handler |
| AC-WBG-F-002 | Combined short-flag clusters containing d/D/m/M/c/C/f → deny (`-df`, `-fm`, `-vD`) | Matrix cells with RED-now on `d592b0551` |
| AC-WBG-F-003 | Query allowlist regression: `--list develop -v`, `-vv`, `--merged`, `--no-merged`, `--points-at`, `--format`, `--contains` → allow | Matrix cells (green-now; must stay green post-fix) |
| AC-WBG-F-004 | Whole-token classification: `--format` allow vs `--force` deny; `--contains`/`--merged`/`--no-merged` allow despite embedded letters | Matrix cells |
| AC-WBG-F-005 | Preserved surfaces: existing deny tests (`branch_guard_test.go:234-245`) + sentinel + discriminant + fail-open tests all stay green | Existing + M3 test run |
| AC-WBG-F-006 | Opt-in gate: disabled default → no pattern evaluation, no deny | Existing `pre_tool_branch_guard_optin_test.go` stays green |
| AC-WBG-F-007 | Fail-open preserved: non-git cwd / missing git / rev-parse failure → allow + audit entry | Existing fail-open tests stay green |
| AC-WBG-F-008 | Exemption-axes conditions encoded as documented test conditions (+ subagent-shaped negative-path cell, D7) | M3 condition tests / test-doc comments |
| AC-WBG-F-009 | Doctrine alignment: v1.3.2 Query-vs-mutate bullet + table row 2 enumeration name the extended mutation class; sanitized-pair mirror in parity | M3 doctrine edit (Template-First) + parity check |

## §F. Cross-References

- `internal/hook/branch_guard.go:120` — the pattern under fix;
  `:59-111` the discrimination rationale comment block (t42 `-c`/`-C`
  completion + accepted combined-short-flag residual this SPEC closes);
  `:194` `matchBranchStateCommand`; `:210` exemption env read;
  `:238-246` fail-open contract; `branchGuardViolationPrefix` at `:50`.
- `internal/hook/branch_guard_test.go:234-245` (deny axis pins),
  `:308-320` (query allowlist pins), `:610-660` (sentinel + origin test).
- `internal/hook/pre_tool_branch_guard_optin_test.go:140-175` (exemption
  axes tests — AC-REQ-6a/6b of -OPTIN-001).
- `.claude/rules/moai/workflow/main-checkout-branch-guard.md` v1.3.2 —
  doctrine SSOT (§ Rules forbidden table row 2; § Permitted bullet 1;
  § Mechanical enforcement Query-vs-mutate discrimination bullet, lines
  90-95 — the "`-f`/`-u` … pass" sentence REQ-WBG-F-009 updates).
- `.claude/rules/moai/workflow/main-checkout-branch-guard-detail.md` —
  enforcer implementation detail (exemption reachability caveat).
- Predecessors: SPEC-WORKTREE-BRANCH-GUARD-001 (completed),
  SPEC-WORKTREE-BRANCH-GUARD-OPTIN-001 (completed),
  SPEC-WORKTREE-BRANCH-GUARD-DISCRIM-001 (completed).
- Card t458 (doctrine v1.3.2 landing card) and card t467 (this card).
- CLAUDE.local.md §4.1 — the historical over-match friction note this SPEC
  treats as possibly-stale, resolved by M1 measurement, not by trust.

## §G. HISTORY

- 2026-09-03 v0.1.0 — initial draft (manager-spec, plan-phase, card t467
  Class B). Root cause confirmed by plan-phase synthetic probe on tree
  `d592b0551`: 10 mutation forms outside `[dDmMcC]` pass the matcher
  (under-match axis); all probed query forms already allowed (over-match
  axis clean in source; CLAUDE.local.md §4.1 friction note treated as
  stale-pending-M1). Fix scoped to the single `git branch` pattern + tests;
  all other patterns and guard surfaces preserved.
- 2026-09-03 v0.2.0 — audit iteration 1 revision (verdict FAIL 0.79 →
  fixes applied, re-audit pending): D1 "mirrors v1.3.2" reworded to
  "extends"; doctrine disposition resolved as the bounded-update path
  (REQ-WBG-F-009 added, M3 scope). D2 matrix extended with M-20..M-23
  (`--delete`/`--move`/`--copy`, detached `--set-upstream-to`, attached
  `--track=` — auditor-measured live on git 2.50.1); M-14 annotated as a
  measured non-target; REQ-002 flag list extended; AC-001 count updated.
  D3 plan §G fix shapes replaced with the corrected first-token +
  `=`-split discrimination rule. D4 evidence labels scoped (RED-now vs
  green-now characterization; measured vs inferred-pending-M1). D5
  tier-set wording (3 artifacts + progress.md). D6 citation :196→:194.
  D7 AC-008 negative-path cell adopted.
- 2026-09-03 v0.3.0 — audit iteration 2 revision (score 0.83, clears
  threshold; verdict FAIL defect-driven — iteration 1's D1-D7 all
  verified RESOLVED): D8 attached `-u<upstream>` spelling added (M-24
  cell + §G rule 2 "beginning with u"). D9 resolved via the
  rule-extension route: REQ-002 false "covered by the bare-operand
  creation rule" sentence replaced by the positional-creation rule
  (non-flag operand = creation when no list action is selected and not
  consumed as a value-taking flag's operand); option-prefixed creation
  cells M-25..M-28 added (auditor-measured creating branches on git
  2.50.1); `git branch -v <name>` deny-vs-permitted-list tension noted
  (the Permitted list names operand-free inquiry forms). D10 `t` added to
  the short-cluster set. D11 M1 expected-count updated to the §D.1
  row set with §D.1 named normative authority. D12 AgentType-axis
  SSOT contradiction (hooks-system.md:114) acknowledged and routed to an
  M3 payload capture + doc-reconciliation blocker path; REQ-008/AC-008 no
  longer assert either side as fact. D13 E-3 `-F` git-rejected
  annotation (rc 129, fail-closed-safe, M-14 treatment) + E-5 tension
  clause. D14 audit-measured cell labels split into git-form liveness
  vs guard-allow inference.

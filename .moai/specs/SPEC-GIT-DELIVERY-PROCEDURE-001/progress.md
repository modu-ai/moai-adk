# Progress — SPEC-GIT-DELIVERY-PROCEDURE-001

> Plan-phase skeleton. §E.2–§E.4 are populated by manager-develop (run-phase) and manager-docs (sync-phase); this agent emits only §E.1.

## §E.1 Plan-phase Audit-Ready Signal

- **Plan-phase artifacts emitted**: spec.md, plan.md, acceptance.md, progress.md (this file). Revision 0.1.2 after plan-audit iteration 2 (`.moai/reports/t622/plan-audit-iter2.md`, FAIL 0.75) and the operator decisions.
- **Tier**: M. **Era**: V3R6. **Status**: `draft`.
- **Card**: t622. **Tree measured**: `b412f8a33b9f82ec5f85ccb5eeb960ef125dd8c0` for scope-file citations; revision authored on HEAD `fb05946db` (the audit report confirmed scope files byte-identical to the base).
- **Pre-write self-checks executed** (0.1.0):
  - SPEC ID regex `^SPEC(-[A-Z][A-Z0-9]*)+-[0-9]{3}$` on `SPEC-GIT-DELIVERY-PROCEDURE-001` → `PASS`.
  - SPEC ID dedup: `ls -d .moai/specs/SPEC-GIT-DELIVERY-PROCEDURE-001` → `No such file or directory`, exit 1 (before creation).
  - Frontmatter: 12 canonical fields present. Extra fields: `tier` (in the schema's Optional Fields table); `era` and `related_specs` are not listed there but are carried by existing SPECs (for example SPEC-WORKTREE-BRANCH-GUARD-001), and the lifecycle audit accepted them.
- **Decisions applied (2026-09-10, operator via lead)**: OD-1 = option 1 (launcher-worktree flow); OD-2 = option B (manager-git.md opt-in; option C was chosen first and replaced after the key's `class: D` / `deprecate_after: "v3.1.0"` record was shown). These are not Implementation Kickoff Approval.
- **Requirement / criterion status**: REQ 15 (active 13: 001-010, 013-015; withdrawn 2: 011, 012). AC 16 (active 14: 001-010, 013-016; withdrawn 2: 011, 012). Tier M ceilings 16 each.
- **Open design items before run-phase (spec.md §C.5)**: B1 PR cardinality and branch prefix; B2 whether the primary-checkout feature-branch path survives; B3 constitution-registry route and file scope.
- **Positive controls measured at plan time** (scratch files in the session scratchpad, not exported — run-phase re-measures into `.moai/reports/t622/run/`):
  - AC-GDP-001 (paragraph judge): base Synchronization section → paragraphs with fetch and rev-list 1, auto-fail 1 (red), list-group 0.
  - AC-GDP-002: base Pre-Spawn block → fetch-without-rev-list 1, standalone rev-list at block line 5, joined count 0 (all red); rev-list count 1.
  - AC-GDP-006: base delivery Step 3.4 section (52 lines) → default-merge phrases at section lines 8 and 20 (file 337, 349), `manager-git.md` mentions 0; base doc-execution Worktree Context Detection subsection → default phrase at subsection line 8 (file 36), mentions 0; manager-git.md opt-in sentences at 148 and 166, 1 each.
  - AC-GDP-007: widened detector with long branch options and optional `-C <path>` at `b412f8a33` on the four template scope files → manager-git 10, spec-workflow 5, spec-assembly 1, delivery 4 = 20 (unchanged from 0.1.1).
  - AC-GDP-008: strict per-file references 1 · 1 · 0, case-sensitive loose 1 · 1 · 0; case-insensitive loose on spec-workflow.md = 2 (line 17 agent-catalog line matches), which is why the loose detector stays case-sensitive. Heading lookup: copy of current manager-git.md → 1; heading-removed fixture → 0, exit 1.
  - AC-GDP-009: base sections — manager-git `moai cc -w` 1 (line 92 detection cue), `push -u origin` 1, `gh pr create` 1, `EnterWorktree(` / `ExitWorktree` / `worktree remove` / `branch -m` / `--<merge_method>` 0; spec-workflow block (15 lines), spec-assembly pre-check (10 lines), delivery Step 3.3.5 (11 lines) all markers 0. PR-prefix sets: spec-workflow {feat/SPEC-, plan/SPEC-} vs manager-git {feat/SPEC-} → differ (red).
  - AC-GDP-010: `worktree add` detector hits `main-checkout-branch-guard.md:38`; delivery Step 3.3.5 `branch -d` 1; both registered clause strings present once each in spec-workflow.md. `moai constitution validate --format json` on the installed build (`v3.2.0-rc.5`, `84fa4ece4`, not built from this tree) → `status: ok`, `drift_count: 0`, `retired_count: 4`. Registry local and template copies `diff` exit 0; `zone-registry` not named in `rule_template_mirror_test.go` (grep exit 1).
  - AC-GDP-013: `doc-execution.md` local vs template `diff` exit 1, single hunk `138,143d137`, header-stripped body 6 lines; not named in `rule_template_mirror_test.go` (grep exit 1). `delivery.md` copies `diff` exit 1 (275 / 278 / footer).
  - AC-GDP-015: SPEC-ID / REQ / date controls on spec.md recorded after the 0.1.2 write in the return message. SHA token extractor (perl lookarounds, added lines only) on a fixture → letter-bearing tokens 538684c47, 7374b183e, 2213871af, 980ccdc56, b412f8a33 on separate lines and again all on one line; all-digit tokens 647460835 and 1000000 split to the reading list; context line 02aca7afe not extracted. On the 0.1.1 spec.md the extractor found 9 tokens including 980ccdc56 (1).
- **Iteration-2 mutants re-run against the revised judges** (session scratchpad):
  - N1: 2nd-iteration pattern caught 2 of the 5 fixture SHAs (per the audit); the token extractor catches all 5, including four on one line.
  - N2: lowercase-title mutant of spec-workflow.md → strict 0 (floor 1) → FAIL; prose mutant → strict 0 → FAIL.
  - N3: mutant lines `branch --force`, `branch --move`, `branch --delete`, `-C . switch` all hit (plus the earlier `pull --rebase`, `branch -f`, `merge --ff-only`); `reset --keep` not in the AGENTS.md set, not caught.
  - N4: block with a second `git -C . rev-list` line → rev-list count 2 → FAIL (fixed fixture 1).
  - N5: then-parallel mutant → auto-fail 1; list mutant → auto-fail 0, list-group 1 (reading step decides); the audit's correct single sentence → auto-fail 0, list-group 0 (not red).
- **plan_status**: audit-ready

## §E.2 Run-phase Evidence

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_

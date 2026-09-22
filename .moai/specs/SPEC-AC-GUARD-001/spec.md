---
id: SPEC-AC-GUARD-001
title: "AC authoring convention and corpus disposition for worktree-guard-refused acceptance criteria"
version: "0.1.1"
status: draft
created: 2026-09-22
updated: 2026-09-22
author: manager-spec
priority: P2
phase: "v3.1.0"
module: ".moai/specs,.claude/rules/moai/workflow"
lifecycle: spec-anchored
tier: M
tags: "worktree-guard,acceptance-criteria,authoring-convention,census,measurement-first"
---

# SPEC-AC-GUARD-001 — AC Authoring Convention and Corpus Disposition for Worktree-Guard-Refused Acceptance Criteria

## HISTORY

| Version | Date | Author | Change |
|---------|------|--------|--------|
| 0.1.0 | 2026-09-22 | manager-spec | Initial plan-phase draft (card t1067; originates from t1059 plan-audit-iter2 D-ITER2-04 and the M1 companion-finding proxy count 782/111/54) |
| 0.1.1 | 2026-09-22 | manager-spec | Plan-audit-iter1 FAIL (0.75) repair: census population re-scoped to FULL corpus (§2.5.1, D2); C1/C2 working-example split per template-neutrality §25.1 (D3); AC-007 added for REQ-001 coverage (D1); §E.1 transition tokens (D4); M4 marker un-split (D5); in-tree dossier cite (D6); GEARS form note (D7) |

## 1. Problem Statement

A subset of acceptance-criteria (AC) verification commands in `.moai/specs/**/acceptance.md` is
refused by the Claude Code worktree guard when a lane attempts to execute them inside a
worktree-isolated session. The refusal ("… names git in a form too complex to verify that it
stays inside the worktree …") is emitted by the Claude Code binary itself, not by repository
code (card t287 attribution table, `worktree-integration.md:472-485`), so guard-side refinement
is upstream-inactionable from this repo.

The commonly cited size of the affected family ("54 signatures", now 55 on develop
`0314801c2`) is a **PROXY**, not a census: the counting predicate (`=$?` capture + git verb on
the same or adjacent lines) diverges from the actual guard in both measured directions. This
SPEC exists to (a) turn the proxy into a verified per-file census using measured family
predicates, (b) author the repo-side actionable surface — an AC authoring convention plus
boundary documentation at a primary convention home, and (c) rewrite only blocks VERIFIED to
refuse.

## 2. Measured Evidence (binding calibration dataset — verbatim from the kickoff measurement, worktree t1067, develop 0314801c2, 2026-09-22)

### 2.1 Proxy reproduction on this tree

- Corpus: `ls .moai/specs/*/acceptance.md | wc -l` → **794 files** (t1059 snapshot: 782 — develop moved)
- `grep -rlE '=\$\?' .moai/specs --include=acceptance.md | wc -l` → **112** (t1059: 111)
- Conjunct `xargs grep -lE '^\s*(git|[A-Z_]+=.*git )'` → **55 files** (t1059: 54)

### 2.2 Positive control

The pre-fix AC-AEC-013 blob (`git show 622e25d22:.moai/specs/SPEC-AUDIT-EXPORT-CLAUSE-001/acceptance.md`,
block at lines 305-318: `git check-ignore -v --no-index "$P"; CHECK=$?` …
`git rm --cached -q "$P" 2>/dev/null; rm -rf "$(dirname "$P")"`) matches both proxy steps,
and that exact block **was refused in the field** (D-ITER2-04, t1059 plan-audit-iter2, refusal
reproduced twice).

### 2.3 Live probe matrix (each probe one Bash invocation, worktree-isolated session)

REFUSED (6):

- P3 x2 (deterministic): `R=$(git rev-parse --short HEAD); echo "$R"`
- R1: `R=$(git rev-parse HEAD); echo "$R"` (no --short)
- R2: `R=$(git rev-parse --short HEAD); printf '%s\n' "$R"`
- R4: `E=$(git status --porcelain); echo "$E"`
- Q2: `mkdir -p /tmp/… && P=/tmp/…/f; mkdir -p "$(dirname "$P")" && printf 'x\n' > "$P"; git check-ignore -v --no-index "$P"; echo "exit=$?"; rm -rf /tmp/…`
- P7: full replica of the original AC-AEC-013 composition (P6 elements + mkdir nested-substitution write + git rm tail + rm -rf), /tmp paths

Family A (P3/R1/R2/R4, 4/4 deterministic): git inside `$( )` assigned to a variable that a
LATER statement expands. Family B (Q2/P7, 2/2): tree-write chain mixing nested
`$(dirname …)`/write + git + `rm -rf` tail.

PASSED (9 distinct shapes; P3 and P8 re-run deterministically):

- P1 plain single git verb; P2 `R=$(git rev-parse --short HEAD)` alone; P4 `git rev-parse --short HEAD; E=$?; echo "exit=$E"`; P5 `git rev-parse --short HEAD && echo ok`; P6 `P=/tmp/…; git check-ignore -v --no-index "$P"; CHECK=$?; echo "check_exit=$CHECK"` (the original skeleton MINUS the write/rm tail — passes); Q3 `git rm --cached -q /tmp/… 2>/dev/null; echo "done=$?"`; R3 `R=$(git rev-parse --short HEAD); echo done` (literal echo, no expansion); P8 x2 `E=$(git status --porcelain | wc -l); echo "count=$E"` (expansion passes when the `$()` terminates in a non-git pipeline stage — exception boundary beyond wc UNMEASURED); P9 `git rev-parse --short HEAD > /tmp/f && cat /tmp/f && rm -f /tmp/f` (redirect-to-file capture)

### 2.4 Proxy verdict

The proxy diverges from the actual guard in BOTH measured directions:

- **False-positive axis**: the counted `$?`-capture skeleton (P4/P6) is guard-executable today; the original AC-013 refusal implicated its write/cleanup composition, not its `$?` skeleton.
- **False-negative axis**: family A carries no `=$?` anywhere, so the proxy never counts it, yet it refuses.

55 is therefore neither an upper nor a lower bound.

### 2.5 Coarse census of the 55 (mechanical, file-level, stated-coarse — verified census is run-phase work)

- Family-B candidates (file contains `mkdir -p` or `rm -rf` AND git-verb lines): **21**
- Exit-capture-only: **34**
- Same-line `=$(…git` family-A form: **0** (multi-line/adjacent-line forms NOT covered by the census predicate)

### 2.5.1 Post-audit census-scope measurement (plan-audit-iter1, 2026-09-22)

The iteration-1 plan audit measured this tree: **59** acceptance.md files corpus-wide contain
`$(git`-shaped text forms; only **18** of them fall inside the conjunct-selected 55;
**41 files** carry git-inside-substitution text forms entirely outside the 55; **23 files**
carry multi-line open command substitutions (`VAR=$(…` unclosed at end of line) — the
predicate §2.5 itself marks uncovered. Consequence (binding): the verified census population
is the **full acceptance.md corpus scanned at block level** (REQ-003); the 55 is a prior
snapshot/expectation, never the scope. Report: `.moai/reports/t1067/plan-audit-iter1.md` D2.

### 2.6 Live corroboration x2

Both agent-7's measurement pipeline and this session's census loop (a while-loop whose regex
TEXT names git in complex composition, executing no git) were refused by the guard — the guard
also refuses non-git-executing text that NAMES git in complex forms.

### 2.7 Refusal message (verbatim)

> This session is isolated in the worktree `<path>`, but this command names git in a form too
> complex to verify that it stays inside the worktree. Refusing to run it — a worktree-isolated
> session's git operations must target its own worktree. Split it into plain, separate commands
> and run them from `<path>`.

## 3. Requirements (GEARS)

- REQ-001 (Ubiquitous): The SPEC corpus (.moai/specs/**/acceptance.md) shall express executable AC verification steps as plain, separately-invocable commands per the authoring convention at the primary convention home.
- REQ-002 (Event-driven): **When** a worktree-isolated session receives the guard refusal (`… too complex to verify …`), the agent shall record the refused command and the verbatim refusal message in the phase evidence, then split or reduce the verification per the measured boundary map — it shall not relocate the verification into a script file.
- REQ-003 (Event-driven): **When** the corpus census runs, it shall apply the measured family predicates (family A: git inside `$( )` assigned to a variable later expanded; family B: tree-write chain mixing nested `$( )`/write + git + `rm -rf` tail), take fresh guard samples per family, and record a per-file disposition (rewrite / leave) with the evidence command, its verbatim output, its exit code, and the tree SHA.
- REQ-004 (Event-driven): **When** the proxy count and the verified census disagree, the verified census shall govern every disposition decision — the proxy (55) is an indicative indicator only and shall not be cited as a census.
- REQ-005 (Event-driven): **When** an AC block is VERIFIED to refuse, the lane shall rewrite it to the plain-verb convention (separate `echo "exit=$?"` capture lines, no git inside `$( )`, no write/rm composition tail) while preserving the verification's semantics — observed stdout, exit code as its own field, and the pinned tree SHA per verification-completeness.md §2.1.
- REQ-006 (Ubiquitous): The primary convention home (the t287 section `## Refused Commands in a Worktree-Isolated Session` in `.claude/rules/moai/workflow/worktree-integration.md`) shall carry the measured boundary map (refused families A/B, executable shapes P4/P6/P8/P9) and a normative AC-authoring rule, maintained under the C1 local + C2 template-mirror discipline with `make build` regeneration.

Form note (GEARS cleanliness): the trailing negative clauses in REQ-002 ("shall not relocate
… into a script file") and REQ-004 ("shall not be cited as a census") are not independent
norms — each restates, at the point of the triggering event, a prohibition already carried
elsewhere (REQ-002's clause ↔ REQ-001 and AC-006; REQ-004's clause ↔ AC-002 and the §2.4
proxy verdict). They are kept inline for reader locality, per plan-audit-iter1 D7's accepted
option.

## 4. Constraints

- The refusing guard is the Claude Code binary — no repository code change can refine it (t287 attribution, `worktree-integration.md:481`).
- Moving AC commands into script files is NOT a sanctioned fix (`kanban-dispatch.md` § The env-isolated verification form: the guard cannot read inside a script; where a verification cannot be expressed as one compound invocation, REDUCE the verification).
- Any edit to `worktree-integration.md` must be applied to BOTH the C1 local copy and the C2 template mirror (`internal/template/templates/.claude/rules/moai/workflow/worktree-integration.md`), followed by `make build`; if the growth exceeds 1,000 bytes in a single edit, the rule-authoring.md statement duty applies in the commit body.
- Do NOT modify t1059's SPEC artifacts (SPEC-AUDIT-EXPORT-CLAUSE-001) — cite as prior art only.
- D-ITER2-09's safety direction (write-into-audited-tree + `rm -rf` hazards) folds into the convention: rewrites must not reintroduce destructive cleanup tails into AC blocks.

## 5. Success Criteria

- Verified full-corpus, block-level census exists with per-file dispositions and fresh guard samples per family (the 55 is a prior snapshot, not the scope).
- The measured boundary map + normative authoring rule live at the primary convention home in both C1 and C2 trees; `make build` passes.
- Every rewritten AC block executes without refusal in a worktree-isolated session and preserves its verification semantics.
- Zero AC verifications relocated into script files by this work.
- The proxy count (55) is retired from normative use; citation as a census is a defect.

## 6. Non-Functional Constraints

- Census and guard probes are read-only or /tmp-harmless; probes must never write into the audited tree (D-ITER2-09).
- Verification load is lane-local: no full-suite runs are triggered by this SPEC (this is a documentation + SPEC-corpus change; affected-package checks apply only where Go code is touched — none expected).

## 7. Out of Scope

### Out of Scope — Guard-side refinement
- Any change to the Claude Code worktree guard's discriminator or refusal behavior — upstream-inactionable from this repository.
- Any change to `internal/hook/pre_tool.go` or `internal/hook/branch_guard.go` — neither implements nor configures the refusing guard.

### Out of Scope — Script-file relocation fixes
- Moving refused AC commands into script files, helper scripts, or generated fixtures to bypass the guard — prohibited by kanban-dispatch.md § The env-isolated verification form.

### Out of Scope — t1059 SPEC artifacts
- Any modification to SPEC-AUDIT-EXPORT-CLAUSE-001 artifacts (including its already-fixed AC-AEC-013) — cited as prior art only.

### Out of Scope — Unmeasured guard-boundary completion
- Exhaustively mapping the guard's full accept/refuse boundary (e.g., the P8 pipeline-terminator exception beyond `wc`, all heredoc shapes) — the convention documents the MEASURED map and labels everything else unknown; completing the map is a separate measurement card if ever needed.

## 8. Dependencies and Prior Art

- SPEC-AUDIT-EXPORT-CLAUSE-001 (t1059) — origin of the defect record (D-ITER2-04) and the working example of the convention (its current AC-AEC-013 at `acceptance.md:622-680` restates commands as plain verbs with separate `echo "exit=$?"` lines plus an explicit note avoiding `$(git merge-base …)`).
- t287 rule section `## Refused Commands in a Worktree-Isolated Session` (`worktree-integration.md:472`).
- `kanban-dispatch.md` § The env-isolated verification form (reduce-don't-script).
- `verification-completeness.md` §2.1 (exit code as its own field; single-invocation RED command form).
- Re-issue precedents: SPEC-INIT-QUIET-WIZARD-001/progress.md:104, SPEC-GOBIN-GOTOOLCHAIN-001/progress.md:147.
- M1 dossier (t1059): its § Companion finding content (the 782/111/54 proxy counts and the proxy caveat) is carried in-tree at `.moai/reports/t1067/guard-probe-matrix-20260922.md` — cite that copy. The original path `.claude/worktrees/t1059/.moai/reports/t1059/m1-direction-dossier.md` is worktree-relative and dies when the t1059 tree is disposed (plan-audit-iter1 D6); do not cite it as evidence.

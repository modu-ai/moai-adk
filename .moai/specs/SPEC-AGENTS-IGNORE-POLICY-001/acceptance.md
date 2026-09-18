---
id: SPEC-AGENTS-IGNORE-POLICY-001
title: "acceptance — .agents .gitignore policy ruling"
version: "0.1.0"
created: 2026-09-14
updated: 2026-09-14
author: manager-spec
tier: M
---

# SPEC-AGENTS-IGNORE-POLICY-001 — Acceptance Criteria

All criteria are machine-verifiable WITHOUT applying the policy (card requirement). Every command below runs from the repository root. Evidence baseline: worktree `.claude/worktrees/t738`, HEAD `99e02ac52`, measured 2026-09-14.

## §A AC Matrix

### AC-POL-001 — Ruling recorded with exact line citations (release-blocking)

- **Given** spec.md §D, **When** an auditor greps it for both policy locations, **Then** it finds the dev-repo citation (`.gitignore` lines 133-170) and the template citation (`internal/template/templates/.gitignore` lines 193-226), and a §D.4 section headed as the RULING naming the template policy canonical.
- **Verification command** (single invocation): `grep -n 'RULING' .moai/specs/SPEC-AGENTS-IGNORE-POLICY-001/spec.md`
- **Expected**: ≥1 hit on the §D.4 ruling heading; the section names the template policy canonical and the dev-repo chain divergent.
- **RED-now cell**: observed at authoring — the ruling text exists (this artifact); on a tree without this SPEC the same grep exits non-zero. **Green path**: flipped by M1 (this artifact set).

### AC-POL-002 — Semantic difference evidenced by check-ignore probes (release-blocking)

- **Given** spec.md §D.1/§D.2, **When** the recorded `git check-ignore -v` evidence is re-run on the current tree, **Then** the outputs reproduce the recorded semantics: in the dev repo, `.agents/skills/my-custom/SKILL.md` is ignored via L135 and `moai-clean/SKILL.md` is absent from the output (= tracked); in the template tree, only `moai-workflow-tdd/SKILL.md` is ignored via L204.
- **Verification commands** (two single invocations, verbatim outputs recorded in spec.md §D.1/§D.2):
  - `git check-ignore -v .agents/skills/moai-clean/SKILL.md .agents/skills/my-custom/SKILL.md .agents/skills/moai-workflow-tdd/SKILL.md`
  - `git -C internal/template/templates check-ignore -v .agents/skills/moai-clean/SKILL.md .agents/skills/my-custom/SKILL.md .agents/skills/moai-workflow-tdd/SKILL.md`
- **RED-now cell**: observed 2026-09-14 — outputs match spec.md §D.1/§D.2 verbatim, exit 0, tree `99e02ac52`. **Green path**: unchanged (evidence recording was the deliverable); the criterion flips green at authoring.

### AC-POL-003 — No .gitignore rule changed by this SPEC (release-blocking regression-guard)

- **Given** the card's HARD constraint 2 and REQ-POL-004, **When** the working tree is inspected at SPEC close, **Then** neither `.gitignore` is modified relative to the SPEC's baseline.
- **Verification command** (single invocation): `git status --porcelain .gitignore internal/template/templates/.gitignore`
- **Expected**: empty output, exit 0 — at plan phase AND at sync phase close.
- **RED-now cell**: the criterion is red only if a `.gitignore` edit appears; at plan phase it is green (measured: empty output, tree `99e02ac52`). Classified regression-guard: it guards the lifetime of the SPEC, not a flip target of this work. This is NOT recorded as a post-implementation pass until re-measured at sync close.

### AC-POL-004 — Operator-gate question text present and operator-readable (release-blocking)

- **Given** plan.md §H, **When** the lead composes the adjudication round, **Then** the Korean question text with options A/B/C/보류 is present and complete (option labels + one-line blast-radius statements each).
- **Verification command** (single invocation): `grep -c '옵션 [ABC]' .moai/specs/SPEC-AGENTS-IGNORE-POLICY-001/plan.md`
- **Expected**: `3`.
- **RED-now cell**: observed at authoring (grep returns 3, tree `99e02ac52`). **Green path**: M1 (this artifact set).

### AC-POL-005 — Application deferral documented against t498/t510 (release-blocking)

- **Given** REQ-POL-004 and plan.md §G M3, **When** the SPEC artifacts are searched for the application window, **Then** both the deferral condition (t498 AND t510 closed) and the follow-up-card ownership are stated.
- **Verification command** (single invocation): `grep -n 't498' .moai/specs/SPEC-AGENTS-IGNORE-POLICY-001/spec.md .moai/specs/SPEC-AGENTS-IGNORE-POLICY-001/plan.md`
- **Expected**: ≥1 hit per file.

### AC-POL-006 — Ruling's 16-command re-include set consistent with tracked state (release-blocking)

- **Given** the ruling aligns dev semantics to template lines 211-226, **When** the 16 published command names are enumerated from both sources, **Then** the two sets are equal to each other and to the tracked `.agents/skills` entries.
- **Verification command** (single invocation): `git ls-files .agents | sed 's|.agents/skills/\(moai-[a-z0-9]*\)/SKILL.md|\1|' | sort`
- **Expected**: exactly 16 lines — moai-clean, moai-codemaps, moai-e2e, moai-feedback, moai-fix, moai-gate, moai-goal, moai-harness, moai-loop, moai-mx, moai-plan, moai-project, moai-review, moai-run, moai-sync, moai-todo — matching template .gitignore lines 211-226 one-for-one. (Note: the character class MUST include `0-9`; an `[a-z]`-only class was observed at authoring 2026-09-14 to leave `moai-e2e` untransformed while the other 15 transformed — a silently partial result, not an error.)
- **RED-now cell**: observed at authoring — `git ls-files .agents | wc -l` = 16 (tree `99e02ac52`). **Green path**: M1 (evidence recorded in spec.md §F).

## §B Edge cases covered

- A future sidecar file inside a published dir (`moai-clean/manifest.json`): ignored today (dev L154), tracked under the ruling's target semantics — the semantic difference table (spec.md §D.3) records both outcomes so the follow-up card's post-application probe has a known expectation.
- A locally materialized deploy mirror (`moai-workflow-*`): ignored under BOTH policies — the ruling must not un-ignore it (spec.md §D.3 row 4).
- A user-authored entry (`my-custom/`, root `notes.md`): ignored today, tracked under target semantics.

## §C Quality gates

- Documentation-only change: `go build ./...` unaffected; template neutrality CI guard unaffected (no template content edited at plan phase).
- All AC verification commands are single-invocation, read-only, and runnable from the repo root.

## §D Definition of Done

- [ ] AC-POL-001..006 verified with observed command output (plan-phase where authoring-time, sync-close where lifetime).
- [ ] Operator adjudication recorded in progress.md (AC gate — M2).
- [ ] Follow-up card registered with Option A's dev-side alignment scoped after t498/t510 close (M3).
- [ ] AC-POL-003 re-measured empty at sync-phase close.

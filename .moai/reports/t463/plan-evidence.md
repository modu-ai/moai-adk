# t463 — Plan-phase Evidence (manager-spec)

Date: 2026-09-03 | Tree: `.claude/worktrees/t463` @ `d592b0551eeb731e5bbd3ef330bf71b21c0822c9` | Branch: `WT-lane-push-doctrine`

## SPEC-ID

`SPEC-LANE-PUSH-DOC-001` — regex check `PASS` (Bash `[[ "$ID" =~ ^SPEC(-[A-Z][A-Z0-9]*)+-[0-9]{3}$ ]]`), unique in `.moai/specs/` (no LANE-PUSH-DOC entry; nearest sibling `SPEC-LANE-PUSH-BATCH-001` is completed and cross-referenced, not duplicated).

## Artifacts (Tier S)

- `.moai/specs/SPEC-LANE-PUSH-DOC-001/spec.md` — GEARS REQ-001..005 + AC-001..005 inline (§3) + premise-correction record (§1) + Out of Scope (§5, 3 H3 headings)
- `.moai/specs/SPEC-LANE-PUSH-DOC-001/plan.md` — §A-§H, milestones M1 (sentence rewrite, before/after bytes) + M2 (verification sweep)
- `.moai/specs/SPEC-LANE-PUSH-DOC-001/progress.md` — §E.1-§E.4 skeleton, §E.1 populated (plan-phase audit signal)

Tier S policy note: acceptance.md omitted — AC inline in spec.md §3 per current Tier-scaled artifact rule (2 counted artifacts + progress.md). The earlier sibling SPEC-LANE-PUSH-BATCH-001 carries 4 files under the pre-Tier-scaling convention; not a template to follow here.

## Commit

- sha: resolve at read time with `git log --oneline -1 -- .moai/specs/SPEC-LANE-PUSH-DOC-001` (reference values, dated 2026-09-03: first commit `3b325028a`, amended to `cb04406be` when this sha line was fixed; the amendment itself moves the sha, so the command above is the criterion, not the numbers)
- subject: `docs(t463): plan SPEC-LANE-PUSH-DOC-001 — lane push doctrine sentence repair (Tier S, 2 artifacts)`
- files: 4 (3 SPEC artifacts + this evidence file)
- note: an untracked `.moai/reports/t463/disposal-evidence.md` appeared during plan phase, authored by neither this delegation nor the orchestrator brief — left untracked and excluded from the commit (no sweep-stage).

## Ground truth re-verified (this worktree, file 724 lines)

- ① CLAUDE.local.md:348 — 창을 받으면 chain, no push step, ends "push는 리드가 일괄로 한다" — PRESENT
- ② CLAUDE.local.md:366 — `# push는 창 밖 — 리드가 레인 병합 SHA를 모아 일괄로 한다 (아래 절차)` — PRESENT
- ③ CLAUDE.local.md:346 — completion-report list carries `로컬 병합 SHA` — PRESENT
- REMAINING DELTA — CLAUDE.local.md:349 second sentence: `카드가 마감되면 로컬 develop 병합(창 경유 \`git push origin develop\`)이 **유일한** 공개 경로다.` (defects: parenthetical puts push inside window; no actor named)

## Premise correction (recorded in spec.md §1)

Card brief cited `:305`/`:321` and a ~390-line file — those are coordinates in the STALE primary checkout (`main @ 7ad9f8534`). Items ①②③ of the lead's original 4-item list are already satisfied on develop (landed via SPEC-LANE-PUSH-BATCH-001, card t430, commit `5e3ecd676`). Live develop coordinates: lines 346/348/349/366 in tree `d592b0551`.

## Acceptance criteria summary

- AC-001 — line 349 second sentence carries 리드 (actor) + 창 밖/일괄 (mode) + verbatim `git push origin develop`
- AC-002 — `grep -n '창 경유 .git push origin develop' CLAUDE.local.md | wc -l` → `0` (count form, not exit code — `feedback_grep_c_exit_code_gates_wrong_way`)
- AC-003 — `git diff -U0 <base> -- CLAUDE.local.md` shows exactly one changed line (349); protected carriers untouched
- AC-004 — [HARD] header + 2026-09-01 date, WT-branch/CI prohibition sentence, lane-2 precedent parenthetical byte-identical
- AC-005 — native written Korean register matching §4.1 (review gate; run report cites before/after)

## Skip rationale

Plan-audit SKIPPED per Tier S policy (documentation-only, 1-line target, ACs mechanically checkable). Recorded in progress.md §E.1. Template-mirror constraint honored: zero writes under `internal/template/templates/`; no `make build` / `make agents-emit` needed.

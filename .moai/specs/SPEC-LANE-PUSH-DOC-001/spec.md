---
id: SPEC-LANE-PUSH-DOC-001
title: "Lane push actor sentence repair — CLAUDE.local.md §4.1 [HARD] bullet second sentence names the lead, not the lane, as the develop pusher"
version: "0.1.0"
status: draft
created: 2026-09-03
updated: 2026-09-03
author: manager-spec
priority: P2
phase: "v3.1.5"
module: "CLAUDE.local.md"
lifecycle: spec-anchored
tags: "gitflow, develop, lane-push, lead-batch, documentation, card-t463"
era: V3R6
tier: S
related_specs: [SPEC-LANE-PUSH-BATCH-001]
---

# SPEC-LANE-PUSH-DOC-001 — Lane push actor sentence repair (CLAUDE.local.md §4.1)

## §1 Background & Problem

Card t463 (Class A, Tier S, documentation-only). Operator directive of 2026-09-02: **lanes
never push `develop`; the lead batches develop pushes OUTSIDE the integration window**
(`git push origin develop` is the lead's single-owned, window-excluded, batch act —
SPEC-LANE-PUSH-BATCH-001). One residual sentence in `CLAUDE.local.md` §4.1 still contradicts
that directive in its grammatical structure.

**Measured ground truth** (orchestrator, this run, tree `d592b0551`; re-verified in this
worktree by manager-spec — file is 724 lines in the develop-based tree; the ~390-line figure
and the cited line coordinates `:305`/`:321` belonged to the STALE primary checkout,
`main @ 7ad9f8534`):

- The lead's original 4-item fix list is 3/4 satisfied on develop:
  - ① lane window procedure line (`창을 받으면:`, CLAUDE.local.md:348) — no push step; ends
    with "push는 리드가 일괄로 한다". PRESENT.
  - ② `**운영 절차**` bash block closing comment (`# push는 창 밖 — 리드가 레인 병합 SHA를
    모아 일괄로 한다 (아래 절차)`, CLAUDE.local.md:366). PRESENT.
  - ③ lane completion-report field list includes `로컬 병합 SHA` (CLAUDE.local.md:346). PRESENT.
- **REMAINING DELTA (the only work of this SPEC)** — CLAUDE.local.md:349, second sentence of
  the `**[HARD] WT 브랜치 push·CI 직접 요청 금지 (운영자 지시 2026-09-01).**` bullet:

  > `카드가 마감되면 로컬 develop 병합(창 경유 \`git push origin develop\`)이 **유일한** 공개 경로다.`

  Defect (a): the parenthetical `(창 경유 ...)` wrongly places `git push origin develop`
  INSIDE the integration window — the push is window-EXCLUDED per the 2026-09-02 directive.
  Defect (b): no actor is named, so a lane can read itself as the pusher.

**Premise correction (recorded, not re-derived)**: the card's original fix list had 4 items
and cited coordinates in the stale primary checkout. Items ①②③ landed on develop via
SPEC-LANE-PUSH-BATCH-001 (card t430, commit `5e3ecd676`); this SPEC's scope is exactly the
fourth item, re-anchored to the develop tree at CLAUDE.local.md:349.

## §2 Requirements (GEARS)

- **REQ-001** (Ubiquitous): The second sentence of the `**[HARD] WT 브랜치 push·CI 직접 요청
  금지 (운영자 지시 2026-09-01).**` bullet (CLAUDE.local.md:349) shall name **리드** (the
  lead) as the sole actor of `git push origin develop`, shall state **창 밖** (outside the
  window) and **일괄** (batched) as the mode, and shall keep `git push origin develop` as the
  verbatim command.
- **REQ-002** (Ubiquitous): The rewritten sentence shall explicitly deny the lane the push
  actor role (lane is not the subject that pushes develop) so no lane can read itself as
  the pusher.
- **REQ-003** (Event-detected): `When` any reader scans `CLAUDE.local.md`, the parenthetical
  sequence `창 경유` immediately followed by a backtick-delimited `git push origin develop`
  shall not occur anywhere in the file (the string
  `창 경유 .git push origin develop` matches zero lines).
- **REQ-004** (Ubiquitous): The rewrite shall touch EXACTLY ONE line (CLAUDE.local.md:349).
  The following carriers shall remain byte-identical:
  - the `[HARD]` clause header including the operator-directive date
    `**[HARD] WT 브랜치 push·CI 직접 요청 금지 (운영자 지시 2026-09-01).**`;
  - the WT-branch/CI prohibition sentence (`레인은 `git push origin <WT-브랜치>`를 하지
    않고, ... 판독은 리드 몫이다.`);
  - the lane-2 precedent parenthetical (`(당일 lane-2가 `WT-version-stamp-predicate`를
    origin에 push한 전례로 추가)`);
  - premise items ① (line 348), ② (line 366), ③ (line 346).
- **REQ-005** (Ubiquitous): The rewritten sentence shall be native written Korean matching
  the file's existing §4.1 register (the file's em-dash + bold + backtick idiom); no
  translationese, no English-mapped calque phrasing.

## §3 Acceptance Criteria (Tier S — inline; Given-When-Then)

- **AC-001**: `Given` the develop-based `CLAUDE.local.md` in the run-phase worktree, `When`
  the second sentence of line 349 is inspected, `Then` it contains the actor token `리드`,
  the mode tokens `창 밖` and `일괄`, and the verbatim command `` `git push origin develop` ``.
  (mechanical: `sed -n '349p' CLAUDE.local.md | grep -c '리드'` ≥ 1; same line greps for
  `창 밖`, `일괄` each ≥ 1)
- **AC-002**: `Given` the post-edit file, `When`
  `grep -n '창 경유 .git push origin develop' CLAUDE.local.md` is run, `Then` the PRINTED
  COUNT is 0. (Caveat from project memory `feedback_grep_c_exit_code_gates_wrong_way`:
  `grep` exits 1 when zero matches — the verdict is the printed count `0`, NOT the exit
  code. Use `grep -n '...' CLAUDE.local.md | wc -l` → `0` as the citable form.)
- **AC-003**: `Given` the run-phase commit, `When` `git diff -U0 <base> -- CLAUDE.local.md`
  is read, `Then` exactly one changed line (349) appears and none of REQ-004's protected
  carriers appear in the diff hunk as context-removed (`-`) lines.
- **AC-004**: `Given` the post-edit line 349, `When` its first clause header, prohibition
  sentence, and lane-2 parenthetical are compared against the pre-edit bytes, `Then` they are
  byte-identical (mechanical: `git diff -U0` shows the `-`/`+` pair differing only inside the
  second sentence).
- **AC-005**: `Given` the rewritten sentence, `When` read against the surrounding §4.1
  register (em-dash coordination, bolded key terms, inline backticked commands), `Then` it
  reads as native written Korean with no translationese (review gate; the run-phase report
  cites the before/after pair).

## §4 Constraints

- Documentation-only: no code, no workflow behavior change, no template change.
- Target file is Korean — run-phase edits MUST be native written Korean.
- `CLAUDE.local.md` is repo-local AND tracked AND is NOT a template-mirror target: nothing
  under `internal/template/templates/` is created or modified; no `make build` /
  `make agents-emit` is needed.
- Plan phase ONLY for this delegation: the run phase owns the edit (Implementation Kickoff
  Approval happens later, outside plan phase).

## §5 Out of Scope

### Out of Scope — template mirror surface

- No file under `internal/template/templates/**` is created, modified, or synced;
  `CLAUDE.local.md` is a local-only tracked file (CLAUDE.local.md §2 Local-Only list).

### Out of Scope — develop push behavior itself

- The 2026-09-02 doctrine (lanes never push; lead batches outside the window) is already
  implemented and landed (SPEC-LANE-PUSH-BATCH-001, completed). This SPEC repairs one
  residual sentence, not the behavior.

### Out of Scope — other doctrine files

- `.claude/rules/local/gitflow-lane-protocol.md`, `.claude/rules/local/repo-local-pr-policy.md`,
  `.moai/docs/git-workflow-doctrine.md`, `.moai/docs/git-local-workflow-doctrine.md` were
  covered by SPEC-LANE-PUSH-BATCH-001's 8-point enumeration; no re-touch here.

## §H Cross-References

- Depends on (non-blocking): SPEC-LANE-PUSH-BATCH-001 — the completed SPEC that landed items
  ①②③ and the lead batch-push doctrine this sentence must agree with.
- Precedent note: card's stale-coordinate premise correction recorded in §1 (stale primary
  checkout `main @ 7ad9f8534`, coordinates `:305`/`:321`; live develop coordinates are
  CLAUDE.local.md:346/348/349/366 in tree `d592b0551`).

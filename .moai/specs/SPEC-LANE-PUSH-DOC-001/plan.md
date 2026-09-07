# SPEC-LANE-PUSH-DOC-001 — Implementation Plan

## §A Context

- Card: t463 (Class A, documentation-only). Target: `CLAUDE.local.md` (repo root, tracked,
  Korean, 724 lines in the develop-based tree `d592b0551`).
- One sentence to rewrite: CLAUDE.local.md:349, second sentence of the
  `[HARD] WT 브랜치 push·CI 직접 요청 금지 (운영자 지시 2026-09-01).` bullet.
- The 2026-09-02 operator directive is already encoded everywhere else in §4.1 (items ①②③
  verified present); only this sentence still reads as though a lane's window-path push is
  the public route.

## §B Known Issues

- The card brief carried stale coordinates (`:305`/`:321`, ~390-line file) from the primary
  checkout (`main @ 7ad9f8534`). Corrected: live develop coordinates are lines 346/348/349/366.
  Any future re-derivation MUST re-anchor to the tree being edited, not to the brief.

## §C Pre-flight (run-phase entry checks)

1. Tree verification: `git rev-parse --show-toplevel` is the card worktree; branch is the
   card branch (`WT-lane-push-doctrine` or its run-phase successor); HEAD re-read immediately
   before the edit.
2. Confirm ground truth still holds: `sed -n '349p' CLAUDE.local.md` still contains
   `창 경유` + `` `git push origin develop` ``. If develop absorbed a newer edit to this
   sentence, STOP and re-scope instead of overwriting.
3. Working tree clean of unrelated tracked modifications before the edit
   (`git status --porcelain`, excluding the card's own evidence dirs).

## §D Constraints

- Native written Korean; match §4.1's em-dash + bold + backtick idiom. No translationese.
- Single-line edit; zero collateral line changes.
- NOT a template target — no `internal/template/templates/**` writes, no `make build`,
  no `make agents-emit`.
- No push, no merge, no integration-branch contact.

## §E Self-Verification (run phase closes against these)

- AC-001..AC-005 of spec.md §3, all executed in the run-phase worktree with verbatim output
  recorded in `.moai/reports/t463/` evidence. AC-002 verdict form is
  `grep -n '창 경유 .git push origin develop' CLAUDE.local.md | wc -l` → `0` (count, not
  exit code — project memory `feedback_grep_c_exit_code_gates_wrong_way`).

## §F Milestones

Ordered by decision-reversibility: the only decision of this SPEC is the sentence itself, so
it leads; the sweep is mechanical.

- **M1 — Sentence rewrite (Priority High)**: replace the second sentence of line 349.

  BEFORE (exact bytes):

  > `카드가 마감되면 로컬 develop 병합(창 경유 \`git push origin develop\`)이 **유일한** 공개 경로다.`

  AFTER (canonical form; run phase may polish wording provided REQ-001/REQ-002 tokens hold):

  > `카드가 마감되면 원격 develop 반영이 **유일한** 공개 경로다 — 리드가 창 밖에서 레인 병합 SHA를 모아 일괄로 실행하는 \`git push origin develop\`이며, 레인은 그 push의 주체가 아니다.`

  The AFTER form: names 리드 as the actor, states 창 밖 + 일괄 as the mode, keeps the verbatim
  command, and explicitly denies the lane the pusher role. Header, prohibition sentence, and
  lane-2 parenthetical on the same line stay byte-identical.

- **M2 — Verification sweep (Priority High)**: run AC-001..AC-005; export verbatim outputs
  to `.moai/reports/t463/run-evidence.md`.

## §G Anti-Patterns

- Do NOT reword the `[HARD]` header, the operator-directive date `2026-09-01`, the
  WT-branch/CI prohibition sentence, or the lane-2 precedent parenthetical.
- Do NOT "fix" items ①②③ again — they are already correct; touching them expands the diff
  beyond one line and breaks AC-003.
- Do NOT mirror the change into `internal/template/templates/`.
- Do NOT translate the sentence into English-idiom Korean (calque).

## §H Cross-References

- SPEC-LANE-PUSH-BATCH-001 (completed) — landed the doctrine and items ①②③; its 8-point
  enumeration closed the other doctrine files.
- Evidence dir: `.moai/reports/t463/` (plan-phase evidence: `plan-evidence.md`).

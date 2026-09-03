# SPEC-LANE-PUSH-DOC-001 — Run-phase Evidence (card t463)

Measured in this run, this tree (`WT-lane-push-doctrine` @ pre-edit HEAD `669eb6708`,
worktree `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t463`), against diff base
`d592b0551`. Methodology note: this is a documentation-only SPEC (single-sentence edit in a
repo-local markdown file) — no TDD/RED-GREEN cycle applies; every AC below is a
grep/diff-based mechanical check except AC-005 (register review gate), as spec.md §3
prescribes. Evidence exported to this tracked path before citation.

## AC-001 — rewritten line 349 carries the required tokens

Command: `sed -n '349p' CLAUDE.local.md` (verbatim output, post-edit):

```
- **[HARD] WT 브랜치 push·CI 직접 요청 금지 (운영자 지시 2026-09-01).** 카드가 마감되면 원격 develop 반영이 **유일한** 공개 경로다 — 리드가 창 밖에서 레인 병합 SHA를 모아 일괄로 실행하는 `git push origin develop`이며, 레인은 그 push의 주체가 아니다. 레인은 `git push origin <WT-브랜치>`를 하지 않고, `gh run rerun`/`workflow dispatch` 등 CI를 직접 요청·재요청하지도 않는다 — CI 판정은 develop push가 일으키는 실행에 맡기고, 판독은 리드 몫이다. (당일 lane-2가 `WT-version-stamp-predicate`를 origin에 push한 전례로 추가)
```

Token counts (command → printed count):

| Token | Command | Count | Threshold |
|---|---|---|---|
| `리드` | `sed -n '349p' CLAUDE.local.md \| grep -o '리드' \| wc -l` | `2` | ≥1 |
| `창 밖` | `sed -n '349p' CLAUDE.local.md \| grep -o '창 밖' \| wc -l` | `1` | ≥1 |
| `일괄` | `sed -n '349p' CLAUDE.local.md \| grep -o '일괄' \| wc -l` | `1` | ≥1 |
| `` `git push origin develop` `` | `sed -n '349p' CLAUDE.local.md \| grep -oF '`git push origin develop`' \| wc -l` | `1` | ≥1 |

REQ-002 (lane denied the push-actor role) holds in the same sentence: `레인은 그 push의 주체가 아니다`.

Verdict: **PASS**

## AC-002 — forbidden parenthetical absent

Command: `grep -n '창 경유 .git push origin develop' CLAUDE.local.md | wc -l`

Output: `0`

The verdict is the printed count `0`, not the exit code (project memory
`feedback_grep_c_exit_code_gates_wrong_way`: `grep` exits 1 on zero matches). REQ-003 holds.

Verdict: **PASS**

## AC-003 — exactly one changed line, no protected carrier removed

Command: `git diff -U0 d592b0551 -- CLAUDE.local.md` (verbatim):

```diff
diff --git a/CLAUDE.local.md b/CLAUDE.local.md
index d77aa86eb..f0797e212 100644
--- a/CLAUDE.local.md
+++ b/CLAUDE.local.md
@@ -349 +349 @@ Kanban(`moai cc -k`) / Factory(`moai cc -f N`) 모드에서 레인은 카드 작
- **[HARD] WT 브랜치 push·CI 직접 요청 금지 (운영자 지시 2026-09-01).** 카드가 마감되면 로컬 develop 병합(창 경유 `git push origin develop`)이 **유일한** 공개 경로다. 레인은 `git push origin <WT-브랜치>`를 하지 않고, `gh run rerun`/`workflow dispatch` 등 CI를 직접 요청·재요청하지도 않는다 — CI 판정은 develop push가 일으키는 실행에 맡기고, 판독은 리드 몫이다. (당일 lane-2가 `WT-version-stamp-predicate`를 origin에 push한 전례로 추가)
+- **[HARD] WT 브랜치 push·CI 직접 요청 금지 (운영자 지시 2026-09-01).** 카드가 마감되면 원격 develop 반영이 **유일한** 공개 경로다 — 리드가 창 밖에서 레인 병합 SHA를 모아 일괄로 실행하는 `git push origin develop`이며, 레인은 그 push의 주체가 아니다. 레인은 `git push origin <WT-브랜치>`를 하지 않고, `gh run rerun`/`workflow dispatch` 등 CI를 직접 요청·재요청하지도 않는다 — CI 판정은 develop push가 일으키는 실행에 맡기고, 판독은 리드 몫이다. (당일 lane-2가 `WT-version-stamp-predicate`를 origin에 push한 전례로 추가)
```

Exactly one hunk, exactly one `-`/`+` pair, both at line 349. None of REQ-004's protected
carriers appears as a removed `-` line — the single `-` line is the pre-edit line 349 itself
and is replaced in full by the `+` line (header, prohibition sentence, and lane-2
parenthetical are reproduced byte-identically within it; see AC-004).

Verdict: **PASS**

## AC-004 — header, prohibition sentence, lane-2 parenthetical byte-identical

Two mechanical proofs against `d592b0551`, both `diff`-based:

1. Prefix (clause header): strip the second sentence from both trees' line 349, diff →
   no output. Command form: `sed -E 's/금지 \(운영자 지시 2026-09-01\)\.\*\* .*/.../'`
   applied to `git show d592b0551:CLAUDE.local.md | sed -n '349p'` and post-edit
   `sed -n '349p' CLAUDE.local.md`.
   Output: `PREFIX+SUFFIX BYTE-IDENTICAL`
2. Suffix (prohibition sentence + lane-2 parenthetical, from `레인은 \`git push origin
   <WT-브랜치>\`` onward): `grep -o '레인은 \`git push origin <WT-브랜치>\`.*'` on both
   trees' line 349, diff → no output.
   Output: `SUFFIX BYTE-IDENTICAL`

Items ①②③ (lines 346/348/366) are untouched — the diff hunk in AC-003 covers only line 349
(`@@ -349 +349 @@`), so no other line of the file changed.

Verdict: **PASS**

## AC-005 — native Korean register (review gate)

BEFORE: `카드가 마감되면 로컬 develop 병합(창 경유 `git push origin develop`)이 **유일한** 공개 경로다.`
AFTER: `카드가 마감되면 원격 develop 반영이 **유일한** 공개 경로다 — 리드가 창 밖에서 레인 병합 SHA를 모아 일괄로 실행하는 `git push origin develop`이며, 레인은 그 push의 주체가 아니다.`

Self-check against translationese: the AFTER uses the file's §4.1 idiom — em-dash
coordination (`공개 경로다 — 리드가 …`), bolded key terms (`**유일한**`), inline backticked
command (`` `git push origin develop` ``), native sentence-final `이다` register, and the
subject-comment construction `레인은 그 push의 주체가 아니다`. No English-mapped calque
phrasing (no literal "axis/pillar"-type figurative carry-over, no translationese word
order). The sentence is the canonical form fixed in plan.md §F M1, used verbatim.

Verdict: **PASS**

## Branch state (pre-commit measurement)

- Worktree: `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t463` (STEP 0 verified:
  toplevel match, branch `WT-lane-push-doctrine`, pre-edit HEAD `669eb6708`, ahead-of-
  origin/develop = 3).
- Pre-edit: `git status --porcelain` → empty; `git diff d592b0551 -- CLAUDE.local.md` →
  empty (base clean).
- File length preserved: `wc -l CLAUDE.local.md` → `724` (pre-edit) / `724` (post-edit).

## Gaps

- No independent auditor re-executed these checks in this run (sync-auditor verification
  is the sync-phase's concern; the lead reads this evidence at the integration window).
- AC-005 is a review gate by design (spec.md §3) — a human/native-reader judgment, cited
  as the before/after pair above rather than a mechanical count.

## Residual-risk

- A later absorb of a newer `origin/develop` that touches CLAUDE.local.md §4.1 could
  re-anchor coordinates; the edit is anchored by content (the `[HARD]` bullet), so the
  sentence itself travels correctly.

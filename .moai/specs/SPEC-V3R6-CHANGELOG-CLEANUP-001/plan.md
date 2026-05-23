# SPEC-V3R6-CHANGELOG-CLEANUP-001 — Plan

## Overview

본 SPEC 은 CHANGELOG.md `[Unreleased]` 섹션의 환각·중복·드리프트를 정리하고, 재발 방지를 위한 manager-docs spawn prompt 의 1차 방어선 (B12 standing-rule guard) 을 구축한다. Tier S minimal — 3 milestones, 2 files modified (CHANGELOG.md + manager-develop-prompt-template.md), ~25 LOC net diff.

## Milestones

### M0 — Pre-flight baseline capture (Phase 0, no commit)

**Goal**: run-phase 시작 시점의 CHANGELOG.md 라인 34-39 sha256 baseline 을 capture 하여 progress.md 에 기록. M0 는 Phase 0 pre-flight (commit 없음 — progress.md write 만), plan.md Milestones 통계에는 포함되지 않음.

**Deliverables**:
- `sed -n '34,39p' CHANGELOG.md | sha256sum | awk '{print $1}'` 실행 후 결과를 progress.md line 12 의 `expected_line_34_39_sha256:` clean YAML-key-value 형식으로 in-place 기록 (bulleted prose 형식 금지 — AC-CHL-004 의 `grep -E '^expected_line_34_39_sha256:' progress.md | awk '{print $2}'` 가 매칭하도록)
- M0 는 commit 하지 않음 — progress.md 의 baseline 기록은 M1 commit 에 동반 staged

**Files modified**:
- `.moai/specs/SPEC-V3R6-CHANGELOG-CLEANUP-001/progress.md` (line 12 placeholder → real sha256)

**Constraints**:
- M0 는 별도 commit 이 아닌 Phase 0 — M1 commit 에 progress.md 변경 동반 포함
- baseline sha256 는 64-hex string, 공백/탭/줄바꿈 없이 progress.md 에 단일 라인으로 기록

### M1 — Line 65 hallucination entry deletion

**Goal**: CHANGELOG.md 라인 65 의 SESSION-HANDOFF-AUTO-001 중복·환각 엔트리 단일 라인 삭제. (실제 라인 65 는 sub-bullet 없는 contiguous single-line entry — blank line 66 + `### Changed` heading 라인 67 직전 종료.)

**Deliverables**:
- CHANGELOG.md 의 라인 65 single contiguous line 삭제 (`- **Paste-ready resume auto-persistence — session_end integration** ...` 으로 시작하는 단일 markdown bullet, sub-bullet 없음, 종료는 직후 blank line 66 또는 라인 67 `###` heading 직전)
- §A.1 의 7개 환각 사실 (Path `internal/handoff/` / Files `{package,atomic_write,parser}.go` / Block 6 `감정` / Volume `10 files +556/-3` / Test count `15 functions` / State file `.moai/state/session-handoff.json` / Supersede `.supersede` mechanism) 모두 grep으로 0건 확인

**Files modified**:
- `CHANGELOG.md` (deletion only, 단일 라인 65 + 직후 blank line 66 까지 제거 — net ~-1~-2 라인)

**Constraints**:
- 라인 34-39 (정확한 SESSION-HANDOFF-AUTO 엔트리) byte-identical 보존
- 라인 57/63 (HOOK-OBSERVE-OPT-IN-001 + CI-BASELINE-DRIFT-001 정확한 엔트리) byte-identical 보존
- 라인 1-33, 81+ 영역 절대 침범 금지

### M2 — Sibling AC count reconciliation

**Goal**: CHANGELOG.md 라인 59 (HOOK-ASYNC-EXPAND-001) + 라인 61 (HOOK-CWD-LEAK-AUDIT-001) 의 AC 카운트를 sibling acceptance.md SSOT 와 일치시키도록 정정.

**Deliverables**:
- 라인 59: `AC 12/12 PASS` → `AC 8/8 PASS` (HOOK-ASYNC-EXPAND-001 acceptance.md SSOT 기준)
- 라인 61: `AC 8/8 PASS` → `AC 7/7 PASS` (HOOK-CWD-LEAK-AUDIT-001 acceptance.md SSOT 기준)
- M2 단일 commit 으로 atomic 변경 (부분 실패 시 둘 다 revert)

**Files modified**:
- `CHANGELOG.md` (2-line edit, ~0 net LOC delta)

**Constraints**:
- 라인 57 (HOOK-OBSERVE-OPT-IN-001 `7/7 AC PASS`) byte-identical 보존 — orchestrator pre-flight 검증 완료 (accurate)
- 라인 63 (CI-BASELINE-DRIFT-001 `AC 8/8 PASS`) byte-identical 보존 — orchestrator pre-flight 검증 완료 (accurate)
- 5 sibling SPEC 디렉토리 (`.moai/specs/SPEC-V3R6-{HOOK-OBSERVE-OPT-IN,HOOK-ASYNC-EXPAND,HOOK-CWD-LEAK-AUDIT,CI-BASELINE-DRIFT,SESSION-HANDOFF-AUTO}-001/`) 의 spec.md / plan.md / acceptance.md / progress.md 절대 수정 금지

### M3 — manager-docs CHANGELOG-emission standing-rule guard

**Goal**: BATCH-SYNC 재발 방지를 위해 manager-docs spawn prompt 의 standing-rule guard (B12) 를 `.claude/rules/moai/development/manager-develop-prompt-template.md` Section B 에 추가. (참고: manager-docs 와 manager-develop 는 공통 Section B known issues 를 공유하므로, manager-develop-prompt-template.md 가 SSOT.)

**Deliverables**:
- `.claude/rules/moai/development/manager-develop-prompt-template.md` Section B 에 신규 bullet **B12. Sync-phase CHANGELOG emission discipline** 추가
- B12 내용: (a) "Before drafting CHANGELOG entries, `Read` every implementation file referenced in the SPEC plan.md (do NOT rely on plan.md description alone — plan-phase placeholders may diverge from final implementation)", (b) "Before appending to `CHANGELOG.md` `[Unreleased]` section, run `grep -c '<SPEC-ID>' CHANGELOG.md` — if the count is ≥1, halt emission and return blocker report (avoid duplicate entries from parallel sessions)", (c) "Verify file paths claimed in CHANGELOG match actual `ls <package-path>` output"

**Files modified**:
- `.claude/rules/moai/development/manager-develop-prompt-template.md` (1 file, ~15 LOC 추가)

**Constraints**:
- 기존 Section B 의 B1-B11 bullet 절대 수정 금지 (orchestrator pre-flight 검증: B11 까지 존재)
- B12 위치: B11 직후, "### Section C" heading 직전
- B12 본문은 영어 (instruction body 는 영어, code_comments=ko 정책은 코드 주석에만 적용)

## File Modification Matrix

| File | M1 | M2 | M3 | LOC delta |
|------|----|----|----|-----------|
| `CHANGELOG.md` | DELETE 1 line (+ trailing blank) | EDIT 2 lines | — | ~-1~-2 net |
| `.claude/rules/moai/development/manager-develop-prompt-template.md` | — | — | INSERT ~15 lines | +15 |
| 5 sibling SPEC dirs | PRESERVE (read-only) | PRESERVE (read-only verify SSOT) | PRESERVE | 0 |

## PRESERVE List (untouched scope — manager-develop B10 compliance)

본 SPEC 의 run-phase 는 다음 파일을 절대 수정하지 않는다 (read-only 검증만 허용):

- `CHANGELOG.md` 라인 1-33 (v3.0.0-rc1 섹션)
- `CHANGELOG.md` 라인 34-39 (정확한 SESSION-HANDOFF-AUTO-001 엔트리)
- `CHANGELOG.md` 라인 57 (HOOK-OBSERVE-OPT-IN-001 7/7 — accurate)
- `CHANGELOG.md` 라인 63 (CI-BASELINE-DRIFT-001 8/8 — accurate)
- `CHANGELOG.md` 라인 81+ (Tooling / Standing Rules / Changed (Hook opt-in context) 섹션 이후 전체)
- `.moai/specs/SPEC-V3R6-HOOK-OBSERVE-OPT-IN-001/` (모든 파일)
- `.moai/specs/SPEC-V3R6-HOOK-ASYNC-EXPAND-001/` (acceptance.md 만 read-only SSOT 검증)
- `.moai/specs/SPEC-V3R6-HOOK-CWD-LEAK-AUDIT-001/` (모든 파일)
- `.moai/specs/SPEC-V3R6-CI-BASELINE-DRIFT-001/` (acceptance.md 만 read-only SSOT 검증)
- `.moai/specs/SPEC-V3R6-SESSION-HANDOFF-AUTO-001/` (모든 파일)
- 기타 모든 runtime-managed (`.moai/state/`, `.moai/cache/`, `.moai/logs/`, `.moai/harness/`) 경로

## Technical Approach

### M1 deletion strategy

라인 65 엔트리는 다음 패턴으로 시작한다 (orchestrator pre-flight `sed -n '65p' CHANGELOG.md` 로 식별):

```
- **Paste-ready resume auto-persistence — session_end integration** (SPEC-V3R6-SESSION-HANDOFF-AUTO-001 ...
```

실제 라인 65 는 sub-bullet 없는 single contiguous markdown bullet — 종료는 직후 blank line 66, 다음 라인 67 은 `### Changed (Hook opt-in context)` heading. manager-develop 는 Edit 도구의 `old_string` 에 라인 65 single line + 직후 blank line 66 까지 capture, `new_string` 은 빈 문자열로 deletion. blank line 66 을 포함하지 않으면 이중 blank line residue 가 남을 위험이 있다 (라인 64 도 blank 이므로).

### M2 edit strategy

라인 59:
- old: `(SPEC-V3R6-HOOK-ASYNC-EXPAND-001 Tier M, commits ...). ... AC 12/12 PASS. ...`
- new: `(SPEC-V3R6-HOOK-ASYNC-EXPAND-001 Tier M, commits ...). ... AC 8/8 PASS. ...`

라인 61:
- old: `(SPEC-V3R6-HOOK-CWD-LEAK-AUDIT-001 Tier S, commits ...). ... AC 8/8 PASS. ...`
- new: `(SPEC-V3R6-HOOK-CWD-LEAK-AUDIT-001 Tier S, commits ...). ... AC 7/7 PASS. ...`

Edit 도구 2-call 시퀀스로 두 라인 atomic 정정, M2 단일 commit. Verification: M2 직후 `grep -c 'AC 12/12 PASS' CHANGELOG.md` → 0, `grep -c 'AC 7/7 PASS' CHANGELOG.md` → 1.

### M3 insertion strategy

manager-develop-prompt-template.md 의 Section B 끝 (B11 직후, `### Section C` heading 직전) 에 B12 bullet 삽입. Edit 도구 1-call. 영어 instruction body.

```markdown
**B12. Sync-phase CHANGELOG emission discipline (manager-docs only)**
- Before drafting CHANGELOG entries, `Read` every implementation file referenced in the SPEC plan.md (do NOT rely on plan.md description alone — plan-phase placeholders may diverge from final implementation).
- Before appending to `CHANGELOG.md` `[Unreleased]` section, run `grep -c '<SPEC-ID>' CHANGELOG.md` — if the count is ≥1, halt emission and return blocker report (avoid duplicate entries from parallel BATCH-SYNC sessions).
- Verify file paths claimed in CHANGELOG match actual `ls <package-path>` output before committing.
- Verify AC count in CHANGELOG matches `acceptance.md` (SSOT) — NOT `progress.md` (which may include deferred AC).
- Origin: SPEC-V3R6-CHANGELOG-CLEANUP-001 §A.4 root cause analysis (BATCH-SYNC line 65 hallucination, 2026-05-23).
```

## Risks

### R1 — line 65 single-line deletion 시 인접 blank line 처리

CHANGELOG.md 라인 65 는 sub-bullet 없는 single contiguous markdown bullet 이며, 인접 라인 구조는 다음과 같다: 라인 64 = blank (라인 63 entry 의 trailing blank), 라인 65 = SESSION-HANDOFF-AUTO duplicate entry, 라인 66 = blank, 라인 67 = `### Changed (Hook opt-in context)` heading. Edit `old_string` 에 라인 65 만 capture 하면 라인 64 + 라인 66 두 blank line 이 인접하여 이중 blank line residue 가 남는다. 완화: Edit `old_string` 에 라인 65 + 직후 blank line 66 모두 포함 (2-line `old_string`), `new_string` 은 빈 문자열. 검증: M1 commit 직후 `sed -n '63,66p' CHANGELOG.md` → 라인 63 entry + 라인 64 blank + 라인 65 (= 기존 라인 67 `### Changed`) 의 single-blank-separated 구조 확인.

### R2 — `grep -c` 가 라인 패턴 매칭 실패

§D Quality Gate 의 `grep -c 'AC 8/8 PASS.*HOOK-ASYNC-EXPAND' CHANGELOG.md → 1` 검증이 CHANGELOG.md 의 정확한 라인 형식 ("(SPEC-V3R6-HOOK-ASYNC-EXPAND-001 Tier M, commits `b00f6afd6..f533b458d`): ... AC 12/12 PASS ...") 와 일치하지 않을 위험. 완화: M2 시작 전 manager-develop 가 `grep -n 'HOOK-ASYNC-EXPAND' CHANGELOG.md` 로 정확한 라인 번호 + 패턴 capture, Edit 의 `old_string` 에 그 라인 전체 사용.

### R3 — Parallel session 의 CHANGELOG.md 동시 편집 race

본 SPEC run-phase 진행 중 다른 session 이 또 다른 CHANGELOG.md 편집을 시도 (예: 다른 SPEC 의 sync). 완화: M1 시작 전 `git fetch origin main && git log --oneline -5 -- CHANGELOG.md` 로 race 탐지, 새 커밋 발견 시 manager-develop 는 blocker report 반환 (subagent boundary: AskUserQuestion 금지), orchestrator 가 rebase/abort 판단.

## Verification Batch (Orchestrator-side, Post-implementation)

run-phase 완료 후 orchestrator 가 단일 turn 내 parallel batch 로 실행 (per `.claude/rules/moai/workflow/verification-batch-pattern.md`):

```bash
# 1. 환각 카탈로그 7개 사실 모두 제거
grep -c 'internal/handoff/{package,atomic_write,parser}' CHANGELOG.md

# 2. AC 카운트 정정 검증 (sibling SSOT 일치)
grep -c 'AC 12/12 PASS' CHANGELOG.md  # → 0
grep -c 'AC 8/8 PASS.*HOOK-ASYNC' CHANGELOG.md  # → 1

# 3. B12 가드 삽입 검증
grep -c 'B12. Sync-phase CHANGELOG emission' .claude/rules/moai/development/manager-develop-prompt-template.md

# 4. 라인 34 byte-identical 보존
sha256sum <(sed -n '34,39p' CHANGELOG.md)

# 5. 5 sibling SPEC 디렉토리 untouched
git diff --name-only HEAD~3..HEAD .moai/specs/SPEC-V3R6-{HOOK-OBSERVE-OPT-IN,HOOK-ASYNC-EXPAND,HOOK-CWD-LEAK-AUDIT,CI-BASELINE-DRIFT,SESSION-HANDOFF-AUTO}-001/ | wc -l  # → 0
```

5개 명령은 모두 read-only, 단일 turn parallel batch 안전.

## Commit Strategy

- M1 commit: `docs(SPEC-V3R6-CHANGELOG-CLEANUP-001): M1 라인 65 환각·중복 엔트리 삭제`
- M2 commit: `docs(SPEC-V3R6-CHANGELOG-CLEANUP-001): M2 sibling AC 카운트 정정 (라인 59 + 61)`
- M3 commit: `chore(SPEC-V3R6-CHANGELOG-CLEANUP-001): M3 manager-develop-prompt-template B12 sync-emission 가드`
- M1+M2+M3 모두 main 직진 (Hybrid Trunk Tier S, 1-person OSS) — `.claude/rules/moai/workflow/git-workflow-doctrine.md` 정책
- `--no-verify` 절대 금지 (pre-commit hook warn-only 정상)
- Conventional Commits format + `🗿 MoAI` trailer 의무

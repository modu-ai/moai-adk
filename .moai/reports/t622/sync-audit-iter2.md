# Sync-audit 2회차(델타) 판정 — 카드 t622, SPEC-GIT-DELIVERY-PROCEDURE-001

감사자: sync-auditor (독립 재실행, 읽기 전용). 재현 산출물: `.moai/reports/t622/sync-audit2/`.
범위: 1회차(`.moai/reports/t622/sync-audit.md`, FAIL 84) 차단 결함 F1 의 해소 여부와 수정 커밋의 비회귀만 본다. 1회차 선택 결함 F2-F7 은 후속·메모로 수용된 상태이며 다시 채점하지 않는다.
적용 규칙: `verification-claim-integrity.md` §1.1(표면 3·4: 결함 주장·권고 전제), §2(기준 귀속) — 아래 수치는 모두 이 실행에서 이 트리에 대해 잰 값이다.

## 판정

**Overall Verdict: PASS** — 점수 91/100 (조화평균 90.9; 가중평균 91.1)

F1 은 해소됐다. `doc-execution.md:34` 두 사본이 이제 `is_worktree_context` 의 실제 소비처 두 곳(Phase 13 Step 3.2 워크트리 전달 경로, Phase 14 워크트리 다음 단계 선택지)을 이름으로 밝히고, 이 값이 자동 병합을 결정하지 않는다고 적는다. 두 소비처는 `delivery.md` 두 사본에 실제로 있고, 둘 다 병합하지 않는다. 수정 커밋은 두 사본·progress.md·증거 파일만 건드렸고, 재실행한 판정은 모두 기대값을 냈다.

## 기준 트리 귀속

| 항목 | 값 | 명령 |
|---|---|---|
| 워크트리 / 브랜치 | `.claude/worktrees/t622` / `WT-git-procedure-fixes` | `git rev-parse --show-toplevel`, `git branch --show-current` |
| HEAD (감사 시작·끝) | `947cc8439` · `947cc84392400ee37019c285b31a80f3f9c3118b` (감사 중 이동 없음) | `git rev-parse --short HEAD`, `git rev-parse HEAD` |
| CARD_BASE | `f1f034bb43b06dbde7f7a93c1c51edc5b4d8f5cf` (1줄, 1회차와 같음) | `git merge-base --all develop HEAD` |
| 로컬 develop | `ac6c42c2dc123ca5142fda7440888c2b0eae13a7` (1회차 뒤 이동했으나 merge-base 불변) | `git rev-parse develop` |
| 범위 대조 | `git diff --name-only develop...HEAD` 비어 있지 않음(65KB 목록) | 같은 명령 |
| 판정 도구 | 이 트리에서 `go test`·Makefile 이 컴파일한 바이너리. 설치된 `moai` 바이너리는 쓰지 않았다 | — |

## 1. F1 해소 판정 — 사실 주장 대조

현재 `doc-execution.md:34` (두 사본 같음):

```
- Store result as `is_worktree_context` boolean; it selects the worktree delivery route (Phase 13 Step 3.2) and the worktree next-step options (Phase 14), but it never decides auto-merge
```

수정 전(1회차 추출본 `sync-audit/ac006-local-de.md`·`ac006-template-de.md`): "…as informational context only; no later phase reads it" — 양성 대조 `grep -c -F 'no later phase reads it'` → `1`, `1`.

| 주장 | 대조 명령 | 관측 (verbatim) | 판정 |
|---|---|---|---|
| Phase 13 Step 3.2 가 존재하고 워크트리 문맥을 읽는 경로를 가진다 | `grep -n -E '^#{3,4} (Phase 1[34]\|Step 3\.2)'` 두 delivery 사본; `grep -rn -i -E 'worktree[_ ]context\|is_worktree'` | 로컬 `13:### Phase 13` · `252:#### Step 3.2` · `284:**Worktree context** (detected from git directory structure):`; 템플릿 `13` · `227` · `259` 같은 줄 | 참 |
| Phase 14 가 워크트리 다음 단계 선택지를 가진다 | 같은 두 명령 | 로컬 `416:### Phase 14` · `447:**If worktree context:**`; 템플릿 `391` · `422` 같은 줄 | 참 |
| 워크트리 문맥은 자동 병합을 결정하지 않는다 | `sed -n '336,365p'` 로컬 delivery | Step 3.4: "Merging is opt-in. The single criterion is the `--auto-merge` opt-in … worktree context alone never triggers a merge." (로컬 361 / 템플릿 336) | 참, 모순 없음 |
| Step 3.2 워크트리 경로 자체도 병합하지 않는다 | 로컬 284-287 읽기 | "Push worktree branch to remote / Create PR if not exists / Display PR URL and worktree context" — 병합 단계 없음 | 참 |
| 다른 소비처를 빠뜨려 문장이 거짓이 되는가 | `grep -n -i 'worktree'` 로컬 delivery 전체 + 명령·에이전트·템플릿·`.codex`·`.agents` 전수 검색 | 불리언을 읽는 곳은 위 두 곳뿐. Step 3.4 Post-Merge Automatic Cleanup(401-414)은 레지스트리에서 경로를 찾을 뿐 이 불리언을 조건으로 쓰지 않는다(조건: "Auto-merge succeeded AND `workflow.worktree.auto_cleanup == true`"). 문장은 "only" 를 주장하지 않으므로 거짓이 되지 않는다 | 참 |
| 거짓 문구가 다른 곳에 남았는가 | `grep -rn -F 'no later phase reads it'` 로컬·템플릿 skills, `.agents`, `.codex` | exit 1 (0건); `CHANGELOG.md` 에도 해당 문구 없음 | 잔존 없음 |

결론: 34행의 모든 사실 주장이 두 delivery 사본과 맞고, 자동 병합이 워크트리 문맥에 달렸다는 함의가 없다. **F1 해소.**

## 2. 판정 재실행 (acceptance.md 명령 그대로, HEAD `947cc8439`)

| AC | 명령 | 관측 (verbatim) | 기대 | 결과 |
|---|---|---|---|---|
| 006 절 추출 | acceptance.md 376행 awk, 로컬·템플릿·BASE(`run/base-doc-execution.md`) | `9 ac006-base-de.md` · `9 ac006-local-de.md` · `9 ac006-tmpl-de.md` | 셋 다 비어 있지 않음 | PASS |
| 006 기존 검출식 | `grep -n -i -E 'default (to )?auto-merge\|worktree contexts default\|merges? (automatically\|by default)'` | `local default exit=1` · `tmpl default exit=1` · `base default exit=0` → `8:This affects auto-merge behavior: worktree contexts default to auto-merge.` | L·T exit 1, BASE 양성 | PASS |
| 006 확장 검출식 | acceptance.md 367행 awk | `local ext exit=0 bytes=0` · `tmpl ext exit=0 bytes=0` · `base ext exit=0 bytes=2 lines=8` | L·T 빈 파일(`test -s` 1), BASE 양성 | PASS |
| 006 출처 개수 | `grep -c 'manager-[g]it[.]md'` | `ac006-local-de.md:1` · `ac006-tmpl-de.md:1` · `ac006-base-de.md:0` | 1 이상 ×2 | PASS |
| 006 읽기 기록 | `sync-fix/ac006-reading.md` 읽기 | 절 6행(새 34행)에 Q1 "No — explicitly says it never decides auto-merge"; 소비처 줄 번호 인용이 위 1절 관측과 일치 | 존재 + 전 문장 No | PASS |
| 013 (b) doc-execution | `diff` L·T → `grep -v` 머리 제거 → 기준 본문과 `diff` | `post diff exit=1`, 머리 `9,11c9,11 79,91d78 118c105 128,134c115,116 144c126 156,161d137 173,174c149 180c155 182,189d156`(acceptance.md 449행 BASE 덩어리와 같음), `body diff exit=0`, `base body nonempty exit=0`, 본문 6008 B 양쪽 | body diff exit 0, 기준 본문 비어 있지 않음 | PASS |
| 015 변경분 | `git diff develop...HEAD -- internal/template/templates/` | `diff nonempty exit=0`, `added count=39`(1회차 수정 뒤 기록과 같음), 파일 9개; 새 34행이 추가 줄 172에 있음 | `test -s` 0, 1 이상 | PASS |
| 015 금지 토큰 | SPEC/REQ/날짜/`CLAUDE.local` grep 4종 | `specid exit=1` · `req exit=1` · `date exit=1` · `local-ref exit=1` | exit 1 ×4 | PASS |
| 015 16진 | perl 낱말 추출 → 문자 포함/숫자만 분리 | `tokens=0`, `sha-letter exists exit=0`, `sha-letter nonempty exit=1`, `numeric lines=0` | `test -e` 0 / `test -s` 1 | PASS |
| 015 대조 | spec.md 세 grep; 픽스처 | spec.md `8` / `61` / `23`; 픽스처 `letter=5 numeric=2`; 추가 뮤턴트(한 줄에 SPEC·REQ·날짜·CLAUDE.local) 4검출식 모두 `1` | 1 이상 ×3, 5/2, 검출 | PASS |
| emit 점검 | `make agents-emit-check` / `make commands-emit-check` | `agents-emit-check exit=0` (`ok … agentemit 0.403s`) / `commands-emit-check exit=0` (`ok … commandemit 0.331s`) | exit 0 ×2 | PASS |
| 미러 비회귀 | `go test ./internal/template/ -count=1 -run 'TestSanitizedPairParity\|TestRuleTemplateMirrorDrift' -v` | `go test exit=1`(기준선도 exit 1), 집합 19줄 `pass=17 fail=2`, `new-fail nonempty=1`, `lost-pass nonempty=1`, 정렬 집합 vs `reanchor/mirror-baseline-sets.txt` `diff exit=0`; FAIL 은 `TestRuleTemplateMirrorDrift` · `TestRuleTemplateMirrorDrift/spec-workflow.md`(범위 밖, 기준선과 같음) | 집합 동일 | PASS |

## 3. 수정 커밋 범위

| 커밋 | 명령 | 건드린 파일 |
|---|---|---|
| `ae4861945` | `git show --name-only --format=%H%n%s ae4861945` | 두 `doc-execution.md` 사본, `progress.md`, `.moai/reports/t622/sync-fix/` 증거 21개 |
| `c61405c8b` | 같은 형태 | `progress.md`, `sync-fix/ac015-*` 증거 15개 |
| `947cc8439` | 같은 형태 | `.moai/reports/t622/sync-audit.md` 와 `sync-audit/` 재현물(1회차 보고서 착지뿐) |

`git diff --stat 947cc8439~3 947cc8439` 의 비증거 변경: 두 `doc-execution.md` 각 `2 +-`(34행 한 줄), `progress.md` `1 +`. SPEC 본문(spec/plan/acceptance) 변경 없음. `git diff --stat ae4861945 HEAD -- <두 doc-execution·두 delivery>` → 빈 출력(수정 뒤 추가 편집 없음). `git status --short --untracked-files=no` → 빈 출력.

progress.md 추가 줄은 새 사실 주장(소비처 줄 번호 로컬 284-287/447-450, 템플릿 259-262/422-425; 재판정 수치)을 담는데, 위 1·2절 관측과 모두 일치한다. 이전 "Edited" 항목의 옛 문구 인용은 남아 있고 새 항목이 이를 명시적으로 대체한다 — 기록을 덧붙이는 방식이라 결함이 아니다.

## 차원별 점수

| Dimension | Score | Verdict | Evidence |
|---|---|---|---|
| Functionality (40%) | 94/100 | PASS (must-pass) | F1 해소 — 1절 대조표. AC-006·013(b)·015 재실행 PASS(2절) |
| Security (25%) | 95/100 | PASS (must-pass) | 이번 델타는 지침 한 줄. 템플릿 추가 줄 금지 토큰 4종 exit 1, 16진 0. Critical/High 없음 |
| Craft (20%) | 85/100 | PASS | `make agents-emit-check` 0, `make commands-emit-check` 0. 1회차 F5·F6 감점 유지(재채점 안 함) |
| Consistency (15%) | 85/100 | PASS | L·T 차이 본문 = BASE 본문(`body diff exit=0`), 미러 집합 동일. 1회차 F3·F4 감점 유지 |

## Findings (structured defect-list)

차단 결함 없음.

- N1 [Low] [optional] [confidence: medium] `.claude/skills/moai/workflows/sync/doc-execution.md:34` (템플릿 사본 같은 행) — "it selects the worktree delivery route (Phase 13 Step 3.2)" 의 그 경로는 `delivery.md` Step 3.2 의 `Strategy: github-flow` 아래에만 있다(로컬 284 / 템플릿 259). `git-flow` 전략은 불리언이 아니라 `WT-*` 브랜치 이름으로 경로를 고른다(로컬 297). 거짓은 아니지만 git-flow 독자는 이 불리언이 자기 경로를 고른다고 읽을 수 있다. Required fix(선택): "(Phase 13 Step 3.2, github-flow)" 처럼 전략을 밝힌다.
- N2 [Info] [optional] [confidence: high] `delivery.md` 401-414(로컬) Post-Merge Automatic Cleanup 도 워크트리 상태를 다루지만 레지스트리 조회로 하며 이 불리언을 조건으로 쓰지 않는다. 34행은 소비처를 "only" 로 한정하지 않으므로 거짓이 되지 않는다. 조치 불필요.
- N3 [Info] [optional] [confidence: high] `delivery.md` 284 "Worktree context (detected from git directory structure)" 는 doc-execution 의 두 감지 조건 중 첫째(경로 구성요소)만 언급한다. 카드 이전부터 있던 문구로 이번 델타 밖이다. 조치 불필요.

## Gaps

- CI 미관측: 브랜치는 푸시되지 않았다(레인 규율). 판정은 로컬 재실행이다.
- 1회차 F2-F7 은 지시대로 재검토하지 않았다.
- 교차 모델 감사(`audit_multi`/`codex_audit`/`glm_audit`)는 이번 델타 범위에서 부르지 않았다 — 한 줄 문서 수정의 사실 대조가 판정 대상이고, 대조는 위 grep·읽기로 끝났다.

## Residual-risk

- N1 의 전략 한정 부재로 git-flow 독자가 불리언의 역할을 넓게 읽을 여지가 남는다(선택 사항).
- 로컬 develop 이 CARD_BASE 이후 계속 앞서간다. 병합 창에서 흡수 뒤 재측정이 필요하다(이 판정은 흡수 전 트리 기준).

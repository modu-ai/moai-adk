# SPEC Review Report: SPEC-INTEGRATION-LOCK-TARGET-SOURCE-001
Iteration: 3/3 (최종 — Tier M 상한 2회에 운영자가 1회를 연장)
Verdict: PASS
Overall Score: 0.96 (네 차원의 조화평균, Tier M 기준 0.80 이상)

- 감사 트리: 워크트리 `.claude/worktrees/t637`, HEAD `ad6ba49e5`(v0.3.0 수리 커밋, iter2 보고서 커밋 `e51428db1` 위). `git diff --stat e51428db1 HEAD` = SPEC 4개 파일만 변경(+135 / −61).
- 범위: iter2 결함 N1-N6의 처분 재검증과 이번 수리가 만든 새 결함 점검(delta 감사).
- Reasoning context ignored per M1 Context Isolation. progress.md "iter-2 audit repair"의 처분 주장은 확인할 대상으로만 읽었다.
- 점수 추이: 0.74 → 0.88 → 0.96. 하락이 없으므로 STOP 신호 없음.

## Must-Pass Results

- [PASS] MP-1: REQ-ILT-001…013 변동 없음(v0.3.0은 acceptance.md·plan.md만 바꿨고 spec.md는 HISTORY 한 줄과 §F 한 칸만 변경).
- [PASS] MP-2 (요구 계층): REQ 문장 변동 없음. iter2 판정(GEARS 준수)을 그대로 유지한다.
- [PASS] MP-3: frontmatter 12필드, `version: "0.3.0"`.
- [N/A] MP-4: 단일 언어 CLI 내부 SPEC.
- [PASS] MP-5 D7: 참조 SPEC ID 추출 결과는 ATOMIC-001, LIVENESS-001(모두 completed)과 자기 자신뿐이다. iter2의 가짜 ID `SPEC-X-001`은 사라졌다.
- [PASS] MP-6 D8: spec.md에서 `syscall` 0건.
- [PASS] MP-7: `grep -rn 'NEEDS CLARIFICATION'` → 무출력.

보조: `moai spec lint spec.md` → `0 error(s), 0 warning(s)`(INFO는 D13 그대로). `spec_audit`(project_root 지정) → INFO `EraAutoDetected`만.

## Category Scores

| Dimension | Score | Band | Evidence |
|---|---|---|---|
| Clarity | 0.95 | 1.0 | row b 괄호 정정(acceptance.md:342), row b와 row i를 구분하는 설명 문단(L351-356), §A.4 규칙 3개(L63-87). 남은 흠: row i의 "equivalently"가 서로 다른 두 mutant를 한데 묶는다(O1, 두 mutant 모두 지명 테스트를 실패시키므로 판별력에는 영향 없음) |
| Completeness | 0.95 | 1.0 | seam 테스트 이름과 (a)-(e) 사례(L265-269), §D.5 게이트까지 `cd` 명시, D13 거절 사유를 progress §E.1에 기록 |
| Testability | 0.95 | 1.0 | 판정 명령의 모양이 가드를 통과함을 실측했다(아래 Evidence). mutation row a-i가 각각 지명 테스트를 판별한다. 남은 흠: 재실행 때 이전 증거 파일이 남는 경우(O2), Go 하위 테스트 선택자 분할은 문서 근거일 뿐 실측하지 못함(Gaps) |
| Traceability | 1.00 | 1.0 | 모든 REQ↔AC 대응 유지, §F "rows a-i" 갱신(spec.md 끝 행). iter2에서 판별 row가 없던 `NoConfigCallerFallbackDoesNotWarn`은 이제 row i가 지명한다 |

조화평균: 4 / (1/0.95 + 1/0.95 + 1/0.95 + 1/1.00) = 0.962.

## N1-N6 Disposition Check

| ID | 작성자 처분 | 판정 | 도구로 확인한 근거 |
|---|---|---|---|
| N1 | 수리 | 해결됨 | `grep -n 'BIN' <SPEC dir>/*.md` → acceptance.md:65-66(거절 이유를 설명하는 산문)과 progress/spec의 수리 기록뿐이다. 호출 형태는 없다. CMD-ILT-010(b)(L204)와 세 셀(L212/217/222)은 `/tmp/t637-fx-bin/moai`를 문자 그대로 쓴다. 셀 모양 그대로(바이너리만 `/bin/echo`로 바꿔) 탐침했더니 가드가 받아들였다: 출력 `integration acquire --card tA` / `rc=0` / `0` / `integration status` / `integration release --session fx-lane4` |
| N2 | 수리 | 해결됨 | row b(L342)는 `GitHubFlowEmptyDevelopDoesNotWarn`만 지명한다. 그 셀은 mode manual + github-flow + 빈 develop이라, workflow 절반을 떼면 경고가 나서 실패한다. row i(L349)는 `NoConfigCallerFallbackDoesNotWarn`을 지명한다. "파일 부재를 git-flow로 취급"하는 mutant와 "술어 전체 제거" mutant 모두 빈 develop 값에서 경고를 내므로 이 테스트는 어느 쪽에서도 실패한다. §D.6 "rows a-i"(L393), plan M6 "a-i"도 확인 |
| N3 | 수리 | 해결됨 | `grep -n 'cd /tmp' <SPEC dir>/*.md` → 세 셀 모두 `( cd /tmp/t637-fx-wt/cardA && … )` 안에 있다(L212/217/222). 나머지는 규칙 설명 산문. §D.0의 go 명령, CMD-ILT-014/015/016, §D.5는 모두 제 줄에서 `cd /Users/goos/MoAI/moai-adk-go/.claude/worktrees/t637 &&`로 시작한다. 셀 탐침 뒤 `pwd` → 워크트리 루트 그대로. 서브셸 밖의 `status`/`release`는 cwd에 의존하지 않는다. `integrationLockRoot()`가 `CLAUDE_PROJECT_DIR`을 먼저 읽고(integration.go:47), release는 `integrationLockRoot()`만 쓴다(integration.go:321) |
| N4 | 수리 | 해결됨 | CMD-ILT-016(L260) `-run '^(TestLoadGitFlowDevelopBranch\|TestLoadGitFlowIntegrationConfig_SeparatesGitFlowFromBranch)$'`(블록 안에서는 맨 `|`) — 앵커가 있고 이름이 구체적이다. `grep -n TestGitFlowPredicate` → 0건. §D.0 머리말(L140-141)이 "선언된 예외 없음"이라 적어 §A.1 verbatim 규칙과 충돌하지 않는다 |
| N5 | 수리 | 해결됨 | CMD-ILT-014(L233)의 네 번째 카운트는 release 줄의 `--card <` 개수이고 기대값은 `0`이다. 현재(편집 전) 트리에서 블록 모양 그대로 실행(증거 경로만 스크래치로 변경): `8` / `7` / `0` / `0`. 편집 전에는 둘째(`7`, 기대 `0`)와 셋째(`0`, 기대 `1`)가 실패하므로 현재는 RED이고, 검사가 공허하지 않다 |
| N6 | 수리 | 해결됨 | 양성 대조 문자열은 `'see t637 here'`(L245). 실측: 중립성 패턴 → `1`, SPEC-ID 추출 정규식 → `0`. iter2 표본 문자열은 progress.md에서도 자리표시자로 교체됐다 |
| D13 | 거절(유지) | 수용 | 사유가 progress.md §E.1(iter-2 repair 마지막 항목)에 기록됐다. INFO가 가리키는 커밋은 이미 착지한 `a3b913b85`라 이후 커밋의 트레일러로는 측정되지 않는다. lint 종료 상태에도 영향이 없다. 타당한 사유다 |

## New Findings (iter3)

blocking 결함은 없다.

| ID | 심각도 | 분류 | 위치 | 내용 | 권장 |
|---|---|---|---|---|---|
| O1 | minor | optional | acceptance.md:349 | row i의 "(equivalently: the whole git-flow predicate is dropped …)"는 "파일 부재만 git-flow로 취급"과 "술어 전체 제거"라는 서로 다른 두 mutant를 동치로 적었다. 두 mutant 모두 지명 테스트를 실패시키므로 판별력은 멀쩡하다 | run-phase에서 실제로 적용한 mutant 한 줄을 §E.2에 기록하면 충분하다 |
| O2 | minor | optional | acceptance.md:212, 217, 222 | `test -x /tmp/t637-fx-bin/moai`가 실패하면 서브셸은 건너뛰지만, `;` 뒤의 `grep -c … "$EV/cN.stderr"`는 이전 실행이 남긴 파일을 셀 수 있다. 다만 그 경우 `rc=` 줄이 출력되지 않고 뒤따르는 status 호출도 실패하므로, AC의 `rc=0` 조건에서 걸러진다 | 원하면 전문(preamble) 뒤에 `rm -f "$EV"/cN.*`를 추가한다 |

## Recommendation

PASS다. 판정 근거:

- 필수 통과 7개 항목 모두 PASS(MP-4는 N/A).
- 조화평균 0.96으로 Tier M 기준 0.80 이상.
- iter1 blocking 5건과 iter2 blocking 3건(N1-N3)이 모두 해결됐고, optional N4-N6도 해결됐다. D13 거절은 사유가 기록돼 수용한다.
- 이번 수리로 blocking 결함이 새로 생기지 않았다. O1과 O2는 optional이다.

이 판정은 Implementation Kickoff Approval을 대신하지 않는다. run-phase 진입에는 여전히 사람의 승인이 필요하다. run-phase의 선행 조건은 문서에 적힌 그대로다. 리드 승인 컴파일 슬롯에서 `/tmp/t637-fx-bin/moai`를 빌드하고(§A.4), 픽스처가 없으면 §A.5로 재구성한다.

---
Evidence (이 감사에서 실행, HEAD `ad6ba49e5`):
- `git diff --stat e51428db1 HEAD` → SPEC 4개 파일만
- `grep -n 'BIN' *.md` → 호출 형태 0, 설명 산문과 수리 기록뿐
- `grep -n 'cd /tmp' *.md` → 셀 3개 모두 `( cd …`, 나머지는 산문
- `grep -c '\\|' *.md` → acceptance.md 1(L30 산문), 나머지 0
- `grep -n 'SPEC-X'` → 0, `grep -n TestGitFlowPredicate` → 0
- 셀 모양 가드 탐침(바이너리만 `/bin/echo`로 대체, 증거 경로는 스크래치): 실행됨, 끝난 뒤 `pwd`는 워크트리 루트
- CMD-ILT-014 모양 탐침(증거 경로만 스크래치): 가드 통과, `8` / `7` / `0` / `0`(편집 전 기준선)
- CMD-ILT-015 중립성 줄: 현재 `0`, 추가 줄 수 `0`(편집 전), 양성 대조 `1`, 대조 문자열의 SPEC-ID 추출 `0`
- `moai spec lint` → 0 error / 0 warning. `spec_audit` → INFO 1

Gaps:
- Go `-run '^Parent$/^(flag|blank_flag)$'`의 하위 테스트 분할은 Go 문서의 동작에 근거했다. 스크래치 모듈 작성과 복합 go 명령이 이 세션에서 막혀 실측하지 못했다(iter2와 같다).
- CMD-ILT-015의 `cd <worktree> && git diff … | grep` 줄은 `cd` 없이 같은 파이프만 실행했다. `cd <worktree> && git grep …` 복합 형태는 CMD-ILT-014 탐침으로 통과를 확인했다.
- 재구성 레시피(§A.5)는 실행하지 않았다. internal/cli는 컴파일하거나 실행하지 않았다(지시에 따름).

Residual-risk:
- 가드 정책은 세션 설정에 따라 다를 수 있다. 셀 모양은 `/bin/echo`로 탐침했으므로, 실제 바이너리 경로에 대해 가드가 별도 규칙을 가지는지는 관측하지 않았다.

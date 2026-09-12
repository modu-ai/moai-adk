# SPEC Review Report: SPEC-INTEGRATION-LOCK-TARGET-SOURCE-001
Iteration: 2/2 (Tier M 상한 도달 — `harness.yaml` `plan_audit_tier_ceilings.M: 2`)
Verdict: FAIL
Overall Score: 0.88 (조화평균, Tier M 기준 0.80 이상 — FAIL 사유는 점수가 아니라 blocking 결함 3건)

- 감사 트리: 워크트리 `.claude/worktrees/t637`, HEAD `76e77d5f1`(v0.2.0 수리 커밋, iter1 보고서 커밋 `fe92308e3` 위). `git diff --stat fe92308e3 HEAD` = SPEC 4개 파일만 변경(+336 / −105).
- 범위: iter1 결함 D1-D16의 처분 재검증 + 수리로 생긴 새 결함 점검(delta 감사). 오케스트레이터가 추가로 넘긴 관측(작업 디렉터리 유지)도 판정에 넣었다.
- Reasoning context ignored per M1 Context Isolation. 작성자의 처분 주장(progress.md "iter-1 audit repair", spec.md HISTORY 0.2.0)은 확인할 대상으로만 읽었다.
- 점수 추이: iter1 0.74 → iter2 0.88. 떨어지지 않았으므로 STOP 신호는 없다.

## Must-Pass Results

- [PASS] MP-1: REQ-ILT-001…013 연속, 변동 없음(spec.md:88-131).
- [PASS] MP-2 (요구 계층): 개정된 REQ-006(`Where … when … shall … and shall write it only after …`)과 REQ-010(Ubiquitous)도 GEARS 형식.
- [PASS] MP-3: frontmatter 12필드 유지, `version: "0.2.0"`.
- [N/A] MP-4: 단일 언어 CLI 내부 SPEC.
- [PASS] MP-5 D7: 참조 SPEC은 ATOMIC-001, LIVENESS-001(모두 completed) 그대로. 새 문자열 `SPEC-X-001`(acceptance.md:216, 양성 대조용 가짜 ID)은 D7 추출 정규식에 걸려 "not found" SHOULD로 잡힌다 → 가짜 대조 문자열이라 blocking 아님(N6).
- [PASS] MP-6 D8: `grep -c syscall` → 네 파일 모두 0.
- [PASS] MP-7: `grep -rn 'NEEDS CLARIFICATION' <SPEC dir>` → 무출력.

보조: `moai spec lint spec.md` → `0 error(s), 0 warning(s)`, INFO 1(`OwnershipTransitionUnmeasured`, a3b913b85 — D13 그대로). `spec_audit`(project_root 지정) → INFO `EraAutoDetected`만.

## Category Scores

| Dimension | Score | Band | Evidence |
|---|---|---|---|
| Clarity | 0.90 | 1.0 | REQ-006 시점 명시(spec.md:102), REQ-010 범위 한정(spec.md:117), M5 문구 정렬(plan.md M5). 감점: mutation row b의 본문("workflow half")과 괄호("warns whenever the develop value is empty")가 서로 다른 mutant를 가리킴(N2) |
| Completeness | 0.95 | 1.0 | §A.1-A.6 규약, 재구성 레시피, §D.0 CMD-ILT-001..016, 경계 사례와 역방향 호환(§D.4) 추가 |
| Testability | 0.75 | 0.75 | 표 이스케이프·PASS 줄 수·양성 대조 모두 해결. 그러나 픽스처 명령 3개와 CMD-ILT-010(b)가 이 하네스의 워크트리 가드에 거절되는 형태(N1, 실측), row b의 지명 테스트 하나가 원리상 실패할 수 없음(N2), cwd 유지로 뒤 명령이 엉뚱한 트리에서 돎(N3) |
| Traceability | 0.95 | 1.0 | 모든 REQ↔AC 대응 유지, §F 갱신(spec.md:176-188). 감점: `NoConfigCallerFallbackDoesNotWarn`을 판별하는 mutation row가 실제로는 없다(N2) |

조화평균: 4 / (1/0.90 + 1/0.95 + 1/0.75 + 1/0.95) = 0.879.

## Disposition Check (D1-D16)

| ID | 작성자 처분 | 판정 | 도구로 확인한 근거 |
|---|---|---|---|
| D1 | 수리 | 해결됨, 단 새 결함 N1·N3 유발 | §A.4에서 `EV`/`BIN`을 절대경로로 정의(acceptance.md:66). CMD-ILT-011/012/013(L189/194/199)의 acquire·status·status --json·release 네 호출 모두 `CLAUDE_PROJECT_DIR=/tmp/t637-fx` 접두 확인. 다만 `"$BIN"` 호출 형태는 가드가 거절한다(N1) |
| D2 | 수리 | 해결됨 | `grep -n '\\|' acceptance.md` → L30 설명문 1건뿐(판정 명령 아님). spec/plan에는 0건. 명령은 전부 fenced block. §A.2가 `no tests to run` 부재와 지명 테스트당 `--- PASS:` 1줄을 요구. 양성 대조 실측: L216 명령 → `1`. 경고 prefix 패턴 실측: 경고 줄이 있는 입력 → `1`, 줄 머리가 아닌 입력 → `0` |
| D3 | 수리 | 원래 요구는 해결됨, 새 결함 N2 | `GitHubFlowEmptyDevelopDoesNotWarn`(manual + github-flow + 빈 develop)은 workflow 절반을 뗀 mutant에 실패한다 → row b 판별. 그런데 row b에 함께 적힌 `NoConfigCallerFallbackDoesNotWarn`은 그 mutant에 실패하지 않는다(N2) |
| D4 | 수리 | 해결됨 | personal 모드 YAML(acceptance.md:161-167) + row g. `ActiveModeProfile()`이 `case "personal": return &c.Personal, true`(types.go:207-208)이므로, mode 절반을 떼면 경고가 나 테스트가 실패한다 → 판별 |
| D5 | 수리 | 해결됨 | plan M5 문구 "The integration target branch …"(소문자 포함), AC-010은 대소문자 구분(acceptance.md:112) |
| D6 | 수리 | 해결됨 | REQ-010을 §B 전제 7로 한정(spec.md:117), §E에 런타임 메시지 제외 항목 추가(spec.md:161-163) |
| D7 | 수리 + 감사자 정정 | 해결됨. 작성자의 정정이 옳다 | `grep -c kanban internal/template/rule_template_mirror_test.go` → `0`. `TestRuleTemplateMirrorDrift`의 경로 목록에 kanban-dispatch가 없으므로, iter1 D7에서 이 테스트가 이 편집을 지킨다고 한 것은 **감사자의 과장**이었다. 대체된 `TestTemplateNoInternalContentLeak`는 존재한다. §D 범위에 internal/template 추가(spec.md:134-136) |
| D8 | 수리 | 해결됨(자리표시자 주의 N4) | plan §D 문구 정정. CMD-ILT-016이 `TestLoadGitFlowDevelopBranch`(존재 확인)와 seam 테스트를 실행 |
| D9 | 수리 | 해결됨 | plan §C 시그니처 적응 허용, §B1·§A.3에서 writer를 `cmd.ErrOrStderr()`로 명시 |
| D10 | 수리 | 해결됨 | REQ-006 "only after the window record has been written", AC-008 `RefusedAcquireDoesNotWarn`, §D.2 시나리오 |
| D11 | 수리 | 해결됨(부수 후퇴 N5) | AC-011을 "≥8이고 재측정값과 동일"로 변경 |
| D12 | 수리 | 해결됨(실행 미검증) | 재구성 레시피(acceptance.md:81)의 단계를 하나씩 검토했다. 경로 생성 → init → seed commit(-c 신원) → 브랜치 3개 → worktree 3개 → yaml. 실행은 가드가 복합 명령을 거절해 하지 못했다(Gaps) |
| D13 | 거절 | 수용(optional) | 거절 사유는 SPEC 문서 어디에도 기록되지 않았다(`grep -n D13 <SPEC dir>/*.md` → 무출력). optional 항목이라 문제 삼지 않는다 |
| D14 | (iter1에서 D1에 병합) | 해당 없음 | — |
| D15 | 수리 | 해결됨 | `blank_flag` 하위 테스트 + row h. untrimmed `!= ""` mutant는 `"   "`에서 `flag`를 기록하므로 `config`를 기대하는 테스트가 실패 → 판별 |
| D16 | 수리 | 해결됨 | plan E3에 main 체크아웃·603줄·충돌 시점 문단 추가 |

## New Findings (iter2)

| ID | 심각도 | 분류 | 위치 | 결함 | 필요한 수정 |
|---|---|---|---|---|---|
| N1 | major | blocking | acceptance.md:66, 184, 189, 194, 199 | 픽스처 셀과 CMD-ILT-010(b)가 바이너리를 변수로 부른다(`"$BIN" integration …`). 이 하네스의 워크트리 격리 가드는 명령 이름이 변수로 계산되는 형태를 거절한다. 같은 모양으로 실측했다: `BIN=/bin/echo; … "$BIN" probe…` → *"this command runs echo through the variable BIN in a plain command; spell it out. Refusing to run it"*. 레인 세션(워크트리 격리)에서 MUST-PASS AC-004/005/006의 픽스처 판정 명령이 적힌 그대로는 실행되지 않는다. iter1 D1 수정 권고에서 `BIN=` 변수를 제안한 것은 감사자 쪽 실수였다 | 모든 호출에서 바이너리 경로를 문자 그대로 적는다: `/tmp/t637-fx-bin/moai integration acquire …`. §A.4 전문(preamble)에서 `BIN=`을 빼고 `test -x /tmp/t637-fx-bin/moai &&`로 바꾼다. CMD-ILT-010(b)도 같다. `EV=` 변수를 리디렉션 대상에 쓰는 것은 가드를 통과한다(실측: 아래 N3 증거) |
| N2 | major | blocking | acceptance.md:307, 109 | row b는 "workflow 절반만 뗀다"고 하면서 두 테스트가 반드시 실패해야 한다고 적는다. 그중 `NoConfigCallerFallbackDoesNotWarn`은 이 mutant에 실패할 수 없다. 파일이 없으면 mode가 비어 `mode == manual` 절반이 여전히 거짓이라 경고가 나지 않기 때문이다. §D.3에 따르면 지명 테스트가 초록으로 남으면 카드가 막히는데, 이 경우는 테스트를 고칠 방법이 없다(테스트는 옳고 row가 틀렸다). 괄호 속 "(warns whenever the develop value is empty)"는 술어 **전체**를 떼는 다른 mutant를 뜻해, row 본문과 모순된다. 결과적으로 NoConfig 테스트를 실제로 판별하는 row가 없다 | row b의 지명 테스트는 `GitHubFlowEmptyDevelopDoesNotWarn` 하나로 하고 괄호를 "(warns when mode is manual and the develop value is empty, whatever the workflow)"로 고친다. 새 row i "경고 조건이 git-flow 술어 전체를 버림(빈 develop이면 caller fallback마다 경고)"를 추가하고 `NoConfigCallerFallbackDoesNotWarn`을 지명한다. §D.6의 "rows a-h"를 "a-i"로 고친다 |
| N3 | minor | blocking | acceptance.md:189, 194, 199, 208, 213-218 | 픽스처 셀은 맨 `cd /tmp/t637-fx-wt/cardA && …`로 시작한다. 이 하네스의 Bash 도구는 작업 디렉터리를 호출 사이에 유지하므로(오케스트레이터가 이 카드 세션에서 관측), 셀 뒤에 도는 CMD-ILT-014/015(상대경로 `git grep`, `diff -q`, `go test ./internal/...`, `make build`)는 명령 안에 `cd`가 없어 픽스처 트리에서 실행된다. 대부분 요란하게 실패하지만(0줄, 파일 없음), 순서에 따라 판정이 달라지는 명령은 결함이다. 픽스처가 진짜 git 저장소라 이후의 git 작업이 엉뚱한 저장소를 건드릴 위험도 있다. 서브셸은 실측으로 통과했다: `( cd /tmp/t637-fx-wt/cardA && CLAUDE_PROJECT_DIR=/tmp/t637-fx /bin/echo probe-literal 2>"$EV/p.stderr" ); echo "rc=$?"; pwd` → `probe-literal`, `rc=0`, 그리고 `pwd`는 워크트리 루트 그대로 | 세 셀의 `cd …`부터 release까지를 서브셸 `( cd /tmp/t637-fx-wt/cardA && … )`로 감싼다. 참고로 kanban-dispatch의 서브셸 금지는 `unset` 환경 정리 형태에 대한 것이다. 또는 CMD-ILT-014/015와 §D.0 go 명령 앞에 `cd /Users/goos/MoAI/moai-adk-go/.claude/worktrees/t637 &&`를 붙인다. 둘 다 하는 편이 안전하다 |
| N4 | minor | optional | acceptance.md:230, 235-237 | CMD-ILT-016의 `TestGitFlowPredicate`는 자리표시자인데 §A.1은 "copied verbatim"을 요구한다. `$` 앵커도 없다. §A.2의 PASS 줄 수 규칙 덕분에 치환을 잊어도 공허하게 통과하지는 않는다(해당 이름의 `--- PASS:`가 0줄 → 실패) | §A.1에 "CMD-ILT-016의 자리표시자만 예외로 §E.2 이름으로 치환"을 적고, `$` 앵커를 붙인다 |
| N5 | minor | optional | acceptance.md:208 | iter1의 AC-011에는 "release 줄에 `--card <`가 없어야 한다"는 검사가 있었는데 v0.2.0에서 빠졌다. 지금은 release 줄이 `no card`와 `--card <card-id>`를 함께 담아도 통과한다 | 세 번째 카운트 뒤에 `grep 'hns-release-specialist.md' "$EV/ac-011.txt" \| grep -c -- '--card <'` → `0` 기대를 추가한다(fenced block 안에서는 `\|`가 아니라 맨 `|`) |
| N6 | minor | optional | acceptance.md:216 | 양성 대조 문자열 `SPEC-X-001`이 SPEC ID 형태라, D7 류 스캐너(`grep -Eo 'SPEC-([A-Z][A-Z0-9]+-)+[0-9]+'`)가 "referenced SPEC not found"로 잡는다 | 금지 토큰 대조를 SPEC ID 모양이 아닌 `t637`만으로 하거나(`printf '+see t637\n'`), `SPEC-` 접두만 남긴 비-ID 문자열을 쓴다 |

## Recommendation

판정은 FAIL이다. 필수 통과 항목은 모두 PASS이고 점수(0.88)도 기준을 넘는다. 막는 것은 blocking 결함 3건(N1, N2, N3)이다. iter1의 blocking 5건(D1-D5)은 모두 해결됐다. 새 결함 중 둘(N1, D1 경로)은 감사자의 iter1 권고에서 비롯됐다.

Tier M 반복 상한(2회)에 도달했으므로, 규정대로 오케스트레이터가 사용자에게 선택지를 올려야 한다(PASS-with-debt / 범위 축소 / 반복 연장). 판단 자료:

- 세 결함 모두 기계적 수정이다. N1은 경로를 문자 그대로 적기(5곳), N2는 row b 정정과 row i 추가, N3는 서브셸로 감싸기. 요구 문장, 설계 결정, 범위는 바꾸지 않는다.
- 사용자가 반복을 연장한다면 재감사는 N1-N3 변경분만 대상으로 한다.
- PASS-with-debt를 고른다면 N1-N3이 run-phase 첫 단계의 선결 작업이 되어야 한다. N1과 N3을 고치지 않으면 픽스처 판정 명령을 레인 세션에서 실행할 수 없고, N2를 고치지 않으면 AC-ILT-014가 원리상 통과할 수 없다.

---
Evidence (이 감사에서 실행, HEAD `76e77d5f1`):
- `git diff --stat fe92308e3 HEAD` → acceptance.md 352, plan.md 47, progress.md 21, spec.md 21 변경
- `grep -n '\\|' acceptance.md` → L30(설명문)만. spec.md·plan.md → 0
- 경고 prefix 패턴: 줄 머리 입력 → `1`, 줄 머리가 아닌 입력 → `0`. 중립성 패턴 양성 대조 → `1`. 현재 줄 230 → `0`. `--card <card-id>`를 넣은 편집 모의 줄 → 중립성 `0`, `--card <card-id>` 포함 `1`
- 가드 탐침: `"$BIN" …` → 거절(메시지 위 N1). `( cd /tmp/t637-fx-wt/cardA && CLAUDE_PROJECT_DIR=/tmp/t637-fx /bin/echo … 2>"$EV/p.stderr" )` → 실행됨, 이후 `pwd` = 워크트리 루트
- 기존 테스트 선언 13개 전부 존재(`grep -rhoE '^func (…)\('`): `TestTemplateNoInternalContentLeak`, `TestLoadGitFlowDevelopBranch` 포함. `TestGitFlowPredicate`는 없음(자리표시자, 의도됨)
- `grep -c kanban internal/template/rule_template_mirror_test.go` → `0`
- `types.go:203-210` `ActiveModeProfile`의 personal/team case
- `moai spec lint` → 0 error / 0 warning / INFO 1. `spec_audit` → INFO 1

Gaps:
- Go `-run '^Parent$/^(flag|blank_flag)$'`의 하위 테스트 분할은 Go 문서의 동작(괄호 밖 `/`에서만 분할)에 근거했다. 임시 모듈로 실측하려 했으나 가드가 복합 명령을 거절했고, 스크래치 경로 쓰기도 막혀 실행하지 못했다.
- 재구성 레시피(D12)는 실행하지 않았다. 긴 `git -C /tmp/t637-fx …` 체인을 가드가 받아들이는지는 관측하지 않았다.
- internal/cli는 컴파일하거나 실행하지 않았다(지시에 따름).

Residual-risk:
- 가드 정책은 세션 설정에 따라 다를 수 있다. N1의 거절은 이 세션에서 관측한 것이며, 리드가 primary 체크아웃(워크트리 격리가 아닌 세션)에서 셀을 돌리면 거절되지 않을 수도 있다. 다만 문서는 누가 실행하는지 못박지 않았다(plan M6은 compile slot에서 "run fixture cells"라고만 적었다).

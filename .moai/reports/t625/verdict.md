# t625 판정서 — manager-develop 정의의 지시 모순 3건 (지시 감사 G4)

- 카드: t625
- 브랜치: `WT-develop-guide-conflicts`
- 카드 기준점: `2160d4e92` (두 번째 부모 = 로컬 develop `f3c4ca50c`)
- 감사 기준 트리: main `2213871af` (감사 보고의 줄 번호는 이 트리 기준)

## 1. 주장 (Claim)

| AC | 결과 | 요지 |
|---|---|---|
| AC-05 | 이미 닫힘 — 수정하지 않음 | 전체 스위트 강제 문구는 이 트리의 두 사본(C1·C2)에 없다. 범위 제한 수정 `f33bbe39c`는 HEAD의 조상이지만 감사 기준 main `2213871af`의 조상은 아니다. 감사는 수정 이전의 main을 읽었다. |
| AC-06 | 수정 완료 | Status Responsibility Matrix 첫 문장에서 "`progress.md` 산출물에만"이라는 단일 산출물 주장을 지웠다. 이제 이 문장은 전이 하나(`draft → in-progress`, 첫 run-phase 커밋)만 밝히고, 적용 산출물의 범위는 스키마 SSOT의 Status Transition Ownership Matrix에 맡긴다. 산출물 목록은 새로 쓰지 않았다. "Status transitions owned" 항목은 SSOT와 충돌하지 않아 그대로 두었다. |
| AC-10 | 수정 완료 | 필수 입력의 열거값을 `ddd`, `tdd`, `autofix`로 넓히고, 기존 autofix 절을 가리키는 `autofix` 한 줄 항목을 더했다. SEMAP 전제 조건은 SPEC 전제(plan-auditor PASS + Implementation Kickoff Approval)가 `ddd`/`tdd`에만 묶이도록 한정하고, `autofix`의 전제는 autofix 절이 참조하는 CI 자동 수정 프로토콜의 진입 조건과 선행 조건이라고 밝혔다. |

부수 사항:

- C2(템플릿)를 먼저, C1(로컬)을 같은 문구로 나중에 고쳤다. C2↔C1 비교는 작업 전과 같은 기존 3개 헝크만 남는다.
- 카탈로그 해시는 `manager-develop` 항목 한 줄만 바뀌었다.
- C3(`.codex/agents/moai/manager-develop.toml`)는 손대지 않았다. 따라서 방출 드리프트 검사는 예상대로 실패한다(리드의 `make agents-emit` 대기).
- `/moai` 워크플로의 "cycle_type=ddd or tdd, per development_mode" 줄은 확인했고 그대로 둔다. 이 줄은 SPEC 구현 배차를 설명하며, 그 경로에는 autofix가 해당하지 않으므로 이번 수정과 어긋나지 않는다.

커밋:

| SHA | 제목 |
|---|---|
| `b5da783ba` | docs(t625): record AC-05 reproduction evidence for manager-develop |
| `9742b3fc8` | docs(t625): drop the single-artifact claim from the manager-develop status matrix |
| `de4a27fac` | docs(t625): admit autofix in the manager-develop cycle_type enum |
| (이 판정서를 싣는 커밋) | docs(t625): record verification and verdict for the manager-develop guide fixes — 커밋은 자기 SHA를 담을 수 없으므로 여기 적지 않는다 |

## 2. 증거 (Evidence)

모든 증거는 `.moai/reports/t625/` 아래에 있으며, 명령과 출력은 파일에 그대로 남겼다. 각 파일 끝의 `EXIT=` 줄이 종료 코드다.

### AC-05

- 패턴(`/usr/bin/grep -n -i -E`): `full[- ]?(test[- ])?suite|complete test suite|entire (test )?suite|whole (test )?suite|always full`
- 양성 대조군 — main `2213871af`의 C1 사본: `ac05-grep-main-2213871af-C1.txt` → 92·126·132·135행 적중, `EXIT=0`
  - 예: `126:3. **Verify behavior**: run tests — targeted when `ddd` LARGE_SCALE, otherwise the full suite ...`
  - 예: `132:- Run the COMPLETE test suite (always full, regardless of LARGE_SCALE; ...)`
- 양성 대조군 — main `2213871af`의 C2 사본: `ac05-grep-main-2213871af-C2.txt` → 93·127·133·136행 적중, `EXIT=0`
- 이 트리 C1: `ac05-grep-head-C1.txt` → 적중 없음, `EXIT=1`
- 이 트리 C2: `ac05-grep-head-C2.txt` → 적중 없음, `EXIT=1`
- 조상 관계:
  - `ac05-ancestry-head.txt`: `git merge-base --is-ancestor f33bbe39c HEAD(2160d4e92) EXIT=0`
  - `ac05-ancestry-main.txt`: `git merge-base --is-ancestor f33bbe39c 2213871af EXIT=1`
- `f33bbe39c` = `fix(SPEC-FULL-SUITE-DOCTRINE-001): scope run-phase test execution to the change (t301)`, 2026-08-27

### AC-06

- SSOT 행: `ac06-ssot-row.txt` — 두 사본 모두 99행 `| `draft → in-progress` | manager-develop (on M1 commit start) | ... — first run-phase commit |`. 산출물별 예외 조항 없음.
- 수정 뒤 C2↔C1 비교: `c2-c1-diff-after-ac06.txt`, 작업 전 비교와의 바이트 대조 `cmp-before-after-ac06.txt` → `EXIT=0`(동일)
- 카탈로그: `catalog-diff-ac06.txt` — `f3c720e3ba6101f265dfe4968f7c6baf2b07ee9e53ff268061eb02bad728eb52` → `5b559c54562bff290d493a7d6f464a0aa9e21e75fc8e75b44747a77ba6c41351`, 한 줄 변경
- 변경 파일: `diffstat-ac06.txt` — C1·C2·`catalog.yaml` 각 1줄

### AC-10

- 수정 뒤 C2↔C1 비교: `c2-c1-diff-after-ac10.txt`
  - 단순 바이트 대조 `cmp-before-after-ac10.txt` → `EXIT=1`. 원인은 `autofix` 항목 한 줄이 늘어 헝크 머리의 줄 번호가 1씩 밀린 것(103→104, 128→129)이다.
  - 헝크 머리를 `HUNK`로 정규화한 대조 `cmp-normalized-before-after-ac10.txt` → `EXIT=0`, 정규화된 헝크 수 작업 전 3 / 작업 후 3(정규화가 실제로 적중했음을 보이는 수)
- 카탈로그: `catalog-diff-ac10.txt` — `5b559c54…41351` → `b8fa8c95eb18654a65b5ce81bd605c4c6befda29cec263e56a906962a17eae82`, 한 줄 변경
- 워크플로 배차 줄 확인: `ac10-workflow-dispatch-line.txt` — 두 사본 모두 264행 `sub-agent: manager-develop (cycle_type=ddd or tdd, per quality.yaml constitution.development_mode)`
- 참조 대상 존재 확인: `ac10-autofix-protocol-entry.txt` — `ci-autofix-protocol.md` 두 사본 모두 15행 `## Entry Condition`, 26행 `**Prerequisites**`

### 카드 전체

- 기준점 대비 소스 변경: `card-diffstat-vs-base.txt` — C1 7줄, C2 7줄, `catalog.yaml` 2줄. 카탈로그 해시 순변경은 `f3c720e3…eb52` → `b8fa8c95…ae82`.
- C3 무변경: `c3-untouched-since-base.txt` — `git diff --stat 2160d4e92 HEAD -- internal/template/templates/.codex/` 출력 없음, `EXIT=0`

### 검증 실행

- 범위 한정 테스트: `test-template-scoped.txt`
  - 명령: `go test ./internal/template/ -count=1 -v -run 'Catalog|ManagerDevelop|AgentFrontmatter|Embed'`
  - `--- PASS` 72줄, `--- FAIL` 0줄, 마지막 줄 `ok  github.com/modu-ai/moai-adk/internal/template 0.525s`, `EXIT=0`
  - 실제로 돈 최상위 테스트 49개(`test-template-scoped-runcount.txt`: `grep -c` 결과 `49`, `EXIT=0`). 예: `TestAgentFrontmatterAudit`, `TestAllAgentsInCatalog`, `TestCatalogReferencesValid`, `TestCatalogHashCoversSkillSubfiles`, `TestManagerDevelopActiveAgentPresent`, `TestManagerDevelopIsActiveAgent`, `TestEmbeddedTemplates_AgentDefinitions`. 선택자가 0개에 적중해 공허하게 통과한 경우가 아니다.
- 방출 드리프트 검사: `agents-emit-check.txt`
  - 명령: `make agents-emit-check` (읽기 전용)
  - `golden_test.go:109: .codex/agents/moai/manager-develop.toml: committed artifact differs from emission (sha256 mismatch)`, `make: *** [agents-emit-check] Error 1`, `EXIT=2`
- agentemit 테스트: `test-agentemit.txt`
  - 명령: `go test ./internal/template/agentemit/... -count=1`
  - 실패는 `TestGoldenCommittedArtifactsMatchEmission` 하나이고, 불일치로 이름이 나온 파일은 `manager-develop.toml` 하나다. `EXIT=1`
- 템플릿 중립성: `neutrality-scan.txt`
  - C2 누적 diff(`c2-cumulative-diff.txt`)에서 `+`로 시작하는 줄 5줄(`+++` 머리 1줄 + 실제 추가 4줄)을 `neutrality-added-lines.txt`로 추출
  - 패턴 `SPEC-|REQ-|(^|[^A-Za-z0-9])t[0-9]{3}([^0-9]|$)|[0-9a-f]{40}|[0-9]{4}-[0-9]{2}-[0-9]{2}` → 적중 없음, `EXIT=1`
  - 양성 대조군 `neutrality-control.txt`: 세션 스크래치에 심은 5줄 중 금지 토큰 4줄(SPEC/REQ, 카드 id, 40자리 SHA, 날짜)을 모두 잡고 깨끗한 줄 1줄은 건너뜀, `EXIT=0`. POSIX ERE에서 `\b`는 단어 경계가 아니므로 카드 id는 명시적 경계 클래스로 썼다.
- Cf 문자: `cf-commit1.txt`, `cf-ac06-files.txt`, `cf-commit2.txt`, `cf-ac10-files.txt`, `cf-commit3.txt` 모두 `TOTAL_CF=0`. 계수기는 스크래치에 U+200B를 심은 대조 파일에서 `TOTAL_CF=1`, `EXIT=3`을 냈다.

## 3. 기준 귀속 (Baseline-attribution)

- 측정 트리: `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t625`, 브랜치 `WT-develop-guide-conflicts`
- 작업 시작 시 HEAD `2160d4e92`, 두 번째 부모 `f3c4ca50c`(= 로컬 develop). Step 0에서 `git rev-parse HEAD^2`와 `git rev-parse develop`이 모두 `f3c4ca50ce9b1ae95711c4452cd3a65ffd27449d`였다.
- AC-05의 grep과 C2↔C1 작업 전 비교는 기준점 `2160d4e92`에서 수정 전에 재고 첫 커밋 `b5da783ba`에 따로 실었다(수정과 같은 커밋에 섞지 않음).
- AC-06 증거는 `9742b3fc8` 직전 작업 트리, AC-10 증거와 검증 실행은 `de4a27fac`의 트리에서 쟀다(검증 명령은 커밋 3과 같은 턴에 실행했으며, 그 시점 작업 트리는 커밋 3에 올린 인덱스와 같았다).
- 테스트와 드리프트 검사는 모두 `go test`/`go run`으로 이 트리에서 새로 컴파일해 돌렸다. 설치된 `moai` 바이너리는 어떤 판정에도 쓰지 않았다.

## 4. 미검증 (Gaps)

- 기준점 `2160d4e92`에서 `agentemit` 테스트를 돌려 원래 초록이었는지는 재지 않았다. "실패 원인은 C3가 낡은 것뿐"이라는 판단은 (a) 불일치로 나온 파일이 이번에 고친 C2에서 방출되는 `manager-develop.toml` 하나라는 점, (b) C3가 기준점 이후 바뀌지 않았다는 점에 기댄 추론이다. 리드가 `make agents-emit`을 돌린 뒤 이 테스트가 초록이 되는지가 확인 수단이다.
- `internal/template` 패키지 전체와 저장소 전체 스위트는 돌리지 않았다. 범위 한정 실행만 했고, 전체 판정은 CI 몫이다.
- C2↔C1 헝크가 "같다"는 AC-10 이후 판정은 헝크 머리의 줄 번호를 정규화한 대조에 기댄다. 헝크 본문 바이트는 같지만 줄 번호는 1씩 밀렸다.
- 다른 파일(예: `manager-develop-prompt-template.md`, 기타 스킬·규칙)에 남은 "`ddd` 또는 `tdd`만" 식 서술은 전수 조사하지 않았다. 확인한 것은 `/moai` 워크플로의 배차 줄 하나다.
- 이번 문구 수정을 다른 사람이 의미 차원에서 검토하지는 않았다.

## 5. 잔여 위험 (Residual-risk)

- C3가 재방출되기 전에 이 브랜치를 develop에 병합하면 develop에서 `agentemit` 골든 테스트와 `make build`의 선행 `agents-emit-check`가 빨갛게 된다. 방출 전에 만든 바이너리는 옛 `manager-develop` 정의를 임베드한다.
- AC-10의 새 전제 조건 문장은 `ci-autofix-protocol.md`의 "Entry Condition"과 "Prerequisites" 구성을 가리킨다. 그 규칙 파일의 절 구성이 바뀌면 이 참조가 어긋날 수 있다.
- AC-06 문장이 산출물 범위를 SSOT에 맡겼으므로, 앞으로 SSOT 행이 바뀌면 "Status transitions owned" 항목(4개 산출물 명시)과 다시 어긋날 수 있다. 그 항목은 이번 범위 밖이라 손대지 않았다.

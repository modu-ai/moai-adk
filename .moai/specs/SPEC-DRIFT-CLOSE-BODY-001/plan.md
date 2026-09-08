# SPEC-DRIFT-CLOSE-BODY-001 — 구현 계획

Tier S (spec.md + plan.md; AC는 spec.md §3에 인라인). 마일스톤은 **되돌리기 어려운 결정을 앞에** 둔다 — 술어 정의가 먼저, 코퍼스 전수 판정이 가운데, 기계적 문서 갱신이 마지막이다.

## §A 맥락

`spec.md` §1(배경·기제)·§3(AC)·§5(설계 선택)·§6(측정 환경)이 정본이다. 여기서 반복하지 않는다.

측정 자료는 `.moai/reports/t410/` 한 벌이고, 그 안에서도 지위가 갈린다.

| 자료 | 지위 | 이유 |
|---|---|---|
| `r1-walker-trace.log` | **판정 근거** | 판정기 자신의 함수를 in-package로 호출한 축자 추적 |
| `r3-drift-row.log` | **판정 근거** | 실제 표에서 재현된 행 |
| `r2-blast-radius.log` | **배경 측정만** | 프로브의 경계값. LOOSE/TIGHT 어느 쪽도 확정값이 아니고 TIGHT에는 알려진 오탐이 있다 |
| `discovery.md` | 서술 | 위 셋의 해석 |

**r2의 수(33/15)를 판정에 인용하지 않는다.** 실제 건수는 M3의 전수 대조 결과로 정해진다.

### Tier S 판정 근거

`spec-workflow.md` § SPEC Complexity Tier의 기준은 LOC와 파일 수다 — S는 `< 300 LOC` 이고 `< 5 files`.

- 예상 변경: `drift_index.go`(새 fallback 함수 ~35 LOC) · `drift.go`(호출 1곳 + 술어 헬퍼 ~25 LOC) · 새 테스트 파일 1개(~150 LOC) · `SPEC-ERA-H3-NARROWING-001/spec.md`(HISTORY 1행). **4 파일 · 약 210 LOC.**
- REQ 7개 · AC 7개로 Tier S 상한(각 8) 안이다.

이 추정은 M1 착수 시 `git diff --stat`으로 재측정한다. 4파일/300LOC를 넘으면 Tier를 올리는 것이 아니라 **범위를 줄인다** — 넘는다는 것은 §4의 범위 밖 항목이 새어 들어왔다는 신호다.

**모집단 정의와 초과 시 동작 (0.4.0 amendment 추가 — 카드 t484, plan-audit D6 상환):**

- **파일 수 판정의 모집단은 소스 파일이다.** SPEC 아티팩트·`progress.md`·원장/측정 산출물은 어떤 AC가 요구하는 산출물이지 줄일 수 있는 범위가 아니다 — 그것들을 세면 Tier 판정이 자기 문서화의 양에 좌우되어, 측정하려던 것(구현의 크기)과 다른 것을 재게 된다.
- **초과가 범위 안 작업에서 나올 때**: 초과분이 범위 안 작업(예: AC가 요구하는 테스트)에서 나오면 위 문단의 "범위를 줄인다"는 수행 가능한 대상이 없다 — 원장에 측정을 기록하고 Tier 재판정 blocker를 리드에게 올린다. "범위를 줄인다"는 범위 밖 누출에만 적용한다.

## §B 알려진 함정

1. **1차 워크를 건드리면 안 된다.** `inMemImpliedStatus`의 2단 의미(전체 메시지 후보 창 → subject 재필터)는 `AC-SSP-005c` 코퍼스 동등성으로 고정돼 있다. 여기에 손대는 순간 회귀 범위가 이 카드 밖으로 나간다. 새 경로는 **호출자 쪽 fallback**으로 붙인다.
2. **후보 창은 이미 존재한다.** 1차 워크의 1단(전체 메시지에 specID 포함, `gitLogWindowSize` 상한)이 곧 필요한 후보 집합이다. 새 `git log` 서브프로세스를 띄우지 않는다 — 그러면 `SPEC-SESSIONSTART-PERF-001`이 없앤 O(n) spawn이 되살아난다.
3. **`shouldSkipCommitTitle`이 먼저 문다.** `chore(spec):` / `chore(specs):` 접두는 skip 대상이다. `docs(specs): batch sync-phase close`는 skip되지 않지만 `chore(specs): ...` 형태의 close가 이력에 있으면 새 경로에서도 안 보인다. 새 경로가 skip 필터를 우회할지 말지를 **명시적으로 정하고 그 근거를 적는다**(기본값: 우회하지 않는다 — AC-LSCSK-003의 보호 대상을 뒷문으로 열지 않기 위해).
4. **`(completed)`는 판별자가 아니다.** 두 실측 줄 모두 같은 줄에 `completed`를 갖는다(spec.md §5.3). 키워드로 가르려는 시도는 실패한다.
5. **`--no-cache`를 빼면 공허한 초록이 난다.** drift는 HEAD SHA로 키잉된 캐시를 쓴다. 수리 전후를 같은 HEAD에서 재면 캐시가 옛 표를 그대로 돌려준다. 모든 코퍼스 측정에 `--no-cache`.
6. **설치본 바이너리는 이 트리를 판정할 수 없다.** 항상 `go build -o /tmp/moai-t410 ./cmd/moai` 후 그 경로로 호출한다.
7. **`internal/spec` 테스트는 git 이력에 의존한다.** 워크트리 안에서 `cachedMainBranch()`가 무엇을 고르는지 M1 착수 시 한 번 확인하고 기록한다 — `main`이 아니면 코퍼스 AC의 의미가 달라진다.

## §C 사전 점검 (run-phase 착수 전)

```bash
git rev-parse --show-toplevel            # → .../.claude/worktrees/t410
git branch --show-current                # → WT-drift-false-positive
git rev-parse --short HEAD               # 원장에 적을 트리 좌표
go build -o /tmp/moai-t410 ./cmd/moai    # rc 0
/tmp/moai-t410 spec drift --no-cache > .moai/reports/t410/drift-before.txt
grep -c '^SPEC-' .moai/reports/t410/drift-before.txt          # 0이면 측정이 잘못된 것
grep SPEC-V3R6-SESSION-HANDOFF-AUTO-001 .moai/reports/t410/drift-before.txt   # DRIFT 여야 착수 가능
git rev-list --count main..develop       # §6의 참조값을 지금 다시 잰다
```

마지막 두 줄이 전제다. `SPEC-V3R6-SESSION-HANDOFF-AUTO-001` 행이 이미 DRIFT가 아니면 **결함이 이 트리에 재현되지 않는 것**이므로 착수하지 말고 blocker로 반환한다(다른 카드가 이미 고쳤을 수 있다).

## §D 제약

- [HARD] `spec.md` §4의 범위 밖 다섯 항목을 건드리지 않는다.
- [HARD] 1차 워크(`inMemImpliedStatus`의 2단 구조)와 기존 combined-scope fallback의 동작 불변(REQ-DCB-006).
- [HARD] 새 `git log` 서브프로세스 0개.
- [HARD] fallback의 출력은 `completed` 아니면 무판정(REQ-DCB-005).
- 모든 코퍼스 수치에 **잰 브랜치 + 잰 트리 SHA** 동반.
- 시간 추정 금지. 우선순위와 순서로만 기술한다.

## §E 자체 검증

각 마일스톤 종료 시 `verification-claim-integrity.md` §3의 5절(Claim / Evidence / Baseline-attribution / Gaps / Residual-risk)로 원장을 남긴다. 원장 경로: `.moai/reports/t410/run-evidence.md`. Evidence에는 **명령과 축자 출력과 exit code**를 적는다 — 요약은 증거가 아니다.

## §F 마일스톤

되돌리기 어려운 순서다. M1의 술어가 나머지 전부를 결정하므로 맨 앞이고, M4는 순수 기계 작업이라 맨 뒤다.

### M1 — 술어 정의와 RED (우선순위 High, 되돌리기 가장 어려움)

**결정 대상**: "본문 줄이 close를 선언하는가"의 판정 술어. spec.md §5.3의 줄 선두 후보를 출발점으로 삼되, 이것이 이 마일스톤에서 **확정되는 결정**임을 의식한다.

1. 새 테스트 파일 `internal/spec/drift_close_body_test.go`를 먼저 쓴다. 픽스처는 `commitRecord` 슬라이스로 구성해 실제 git 이력에 의존하지 않는다(단위 층).
2. AC-DCB-001의 양성 픽스처(`e979a4d13` 모양)와 AC-DCB-002의 음성 픽스처(`a83934d55` 모양)를 함께 넣는다.
3. AC-DCB-003의 FALLBACK-ONLY 3케이스를 표 테스트로 넣는다.
4. **RED를 실측한다** — `go test ./internal/spec/ -run TestDriftCloseBody -count=1` 의 rc 1과 축자 출력을 원장에 남긴다. RED를 안 본 채 GREEN을 만들면 AC-DCB-001의 공허 방지 조항을 위반한다.
5. 양성 픽스처에 대해 1차 워크가 실제로 `in-progress`를 낸다는 선행 단언을 넣는다(공허 방지).

산출: 테스트 파일 1개. 구현 코드 없음.

### M2 — fallback 배선 (우선순위 High)

**결정 대상**: 새 판정 축을 어디에 붙일 것인가. spec.md §5.1의 판단은 기존 fallback과 **같은 자리**(`drift.go`의 ① 블록, `a.status == "completed" && gitStatus != "completed" && !isTerminalStatus(gitStatus)` 게이트 안)다.

1. `drift_index.go`에 `inMemBodyDeclaredClose(commits []commitRecord, specID string) bool` 추가. 후보 창은 1차 워크와 동일(전체 메시지에 specID 포함, `gitLogWindowSize` 상한).
2. gate (a) subject가 specID를 **담지 않음**(담으면 1차 워크 소관) · gate (b) 본문 줄 술어 두 모양(spec.md §5.3). subject 쪽 close 신호를 후보 필터로 추가한다면 §5.4의 세 조건을 만족시키고 6개 커밋 통과를 실측한다 — **`closeInfixMatch` 단독은 기각됐다**(6개 중 4개만 통과, 확정 대상 `e979a4d13` 탈락).
3. §B-3의 skip-필터 우회 여부를 결정하고 코드 주석에 근거를 남긴다.
4. `drift.go`의 ① 블록에서 기존 `inMemCombinedScopeClose` 다음 순서로 호출한다(기존 경로 우선 — REQ-DCB-006).
5. `go test ./internal/spec/ -run TestDriftCloseBody -count=1` → rc 0 (GREEN).
6. **뮤테이션 2종**: (i) 본문 줄 술어를 `strings.Contains(c.fullMsg, specID)`로 바꾸고 AC-DCB-002가 실패함을 실측한다. (ii) 모양 A의 `:` 요구를 제거하고 §5.3 픽스처 8번 줄이 실패함을 실측한다. 각각 rc와 축자 출력을 원장에. 어느 하나라도 실패하지 않으면 M1으로 되돌아간다.
7. `go test ./internal/spec/... -count=1` → rc 0, `-v` 출력에서 AC-DCB-006이 명명한 테스트들이 실제 실행됐음을 확인.

산출: `drift_index.go` + `drift.go` 변경, 원장의 GREEN·뮤턴트 축자 출력.

### M3 — 코퍼스 전수 판정 (우선순위 High, 사람 판단이 들어가는 유일한 단계)

**여기서 실제 건수가 정해진다.** 조사의 15/33은 입력이 아니다.

```bash
/tmp/moai-t410 spec drift --no-cache > .moai/reports/t410/drift-after.txt
diff <(sort .moai/reports/t410/drift-before.txt) <(sort .moai/reports/t410/drift-after.txt)
```

1. 비-DRIFT → DRIFT로 바뀐 행이 0건임을 확인한다. 1건이라도 있으면 회귀이며 M2로 되돌아간다.
2. DRIFT → 해제로 바뀐 **모든** 행을 열거하고, 행마다 `git log main --grep=<ID> -50` 으로 해제시킨 커밋을 찾아 SHA·subject·본문 해당 줄을 원장에 적는다.
3. 각 줄이 **언급이 아니라 close 선언**인지 사람이 판단해 적는다. 판단이 갈리면 그 행은 해제하지 않는 쪽으로 술어를 좁힌다 — 넓히지 않는다(spec.md §5.3).
4. `SPEC-AUTONOMY-TIERS-001`이 해제 목록에 **없음**을 명시적으로 확인한다(실측된 반례).
5. AC-DCB-004의 확정 행이 해제 목록에 **있음**을 확인한다.

산출: 원장의 행별 판정표. 총계는 부산물이지 판정 기준이 아니다.

### M4 — t382 HISTORY 정정 (우선순위 Medium, 기계적)

1. `.moai/specs/SPEC-ERA-H3-NARROWING-001/spec.md`의 HISTORY 표에 행 하나를 추가한다: 원인 서술 ①이 t410 재현으로 반증됐다는 사실, 근거 경로 `.moai/reports/t410/r1-walker-trace.log`, 그리고 **AC 판정과 상태는 유효하다**는 명시.
2. `version:`을 한 단계 올리고 `updated:`를 갱신한다. 그 밖의 프론트매터·본문은 건드리지 않는다.
3. AC-DCB-007의 세 판정을 실행하고 출력을 원장에 남긴다.

산출: 한 파일의 HISTORY 1행 + 두 필드.

## §G 안티패턴

- **조사의 15/33을 기대값으로 삼는 것.** 경계값이고 상한에는 알려진 오탐이 있다. 그 수를 목표로 잡으면 술어를 그 수에 맞추게 된다.
- **총 drift 수의 감소로 성공을 주장하는 것.** 표의 대부분은 브랜치 격차 소음이다(spec.md §6). 총계는 이 수리에 거의 반응하지 않는다.
- **RED를 건너뛰고 GREEN만 보이는 것.** 픽스처가 결함을 재현하지 않아도 초록이 난다.
- **반례가 나왔을 때 술어를 넓혀 삼키는 것.** 그것이 기각된 후보 A로 가는 뒷문이다.
- **`--no-cache` 누락.** 같은 HEAD에서 캐시가 옛 표를 돌려준다.
- **설치본 `moai`로 판정하는 것.** 이 트리의 코드를 재는 것이 아니다.
- **1차 워크를 "겸사겸사" 정리하는 것.** `AC-SSP-005c` 동등성 밖으로 나간다.

## §H 교차 참조

- `spec.md` — 요구사항·AC·설계 선택·측정 환경(정본)
- `.moai/reports/t410/discovery.md` — 조사 원장
- `.claude/rules/moai/core/verification-claim-integrity.md` — §1 공허한 주장, §2 baseline 귀속, §2.2 도구 출처, §3 5절 보고 형식
- `.claude/rules/moai/development/spec-frontmatter-schema.md` — Close-subject full-ID mandate, Non-transition frontmatter corrections
- `.claude/rules/moai/workflow/spec-workflow.md` § SPEC Complexity Tier — Tier S 기준

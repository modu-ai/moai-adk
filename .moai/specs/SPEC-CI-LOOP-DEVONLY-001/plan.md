# plan — SPEC-CI-LOOP-DEVONLY-001

## §A Tier 판정

**Tier M** (감사 후에도 유지).

근거 (측정된 범위):

- 편집 대상: 템플릿 14 + 개발 저장소 미러 13 = **27개 문서/설정 파일**
- Go 변경: `legacySkillIDs` 항목 1개 추가, 가드 테스트 1개 신설,
  `pr_watch_cmd.go` **문자열 3곳** 수정 — 신규 서브시스템·타입·명령 없음
- 변경의 성격: 열거된 유한 집합에 대한 제거와 치환

감사 반영으로 봉투가 `pr_watch_cmd.go` 문자열까지 넓어졌으나, 이는 동작 변경이 아니라
사용자 노출 문자열 교정이므로 깊이는 그대로다. Tier L을 정당화하는 설계 결정이 없어
Tier M을 유지하고, 4개 질문 조사를 담은 `research.md`를 최소 산출물에 추가한다.

## §B EXTEND 봉투 (편집 허용 경로)

**템플릿 측 (14)**

```
internal/template/catalog.yaml
internal/template/templates/.claude/agents/moai/manager-develop.md
internal/template/templates/.claude/rules/moai/core/zone-registry.md
internal/template/templates/.claude/rules/moai/development/manager-develop-prompt-template.md
internal/template/templates/.claude/rules/moai/workflow/cadence-bridge.md
internal/template/templates/.claude/rules/moai/workflow/ci-autofix-protocol.md      (재작성)
internal/template/templates/.claude/rules/moai/workflow/ci-watch-protocol.md        (삭제)
internal/template/templates/.claude/skills/moai-workflow-ci-loop/                   (삭제)
internal/template/templates/.claude/skills/moai/SKILL.md
internal/template/templates/.claude/skills/moai/workflows/fix.md
internal/template/templates/.claude/skills/moai/workflows/loop.md
internal/template/templates/.claude/skills/moai/workflows/run.md
internal/template/templates/.claude/skills/moai/workflows/sync/delivery.md
internal/template/templates/.moai/config/sections/delegation.yaml
```

**개발 저장소 미러 측 (10 편집 + 2 보존)** — `catalog.yaml`을 제외한 동일 상대 경로에서
**선언된 dev-only 보존 집합**을 뺀 10개 파일을 편집한다 (AC-CLD-017의 선택자와 동일).

선언된 dev-only 보존 집합 (편집·삭제 금지):

```
scripts/ci-watch/                                      스크립트 5개
scripts/ci-autofix/                                    스크립트 4개
.claude/skills/moai-workflow-ci-loop/                  스킬 디렉터리
.claude/rules/moai/workflow/ci-watch-protocol.md       미러 보존 (배포에서만 제거)
.claude/rules/moai/workflow/ci-autofix-protocol.md     미러 보존 (템플릿만 재작성)
```

보존 이유: 개발 저장소는 스크립트를 계속 보유하므로 그것을 규율하는 규칙도 함께 남는 것이
일관적이다. 배포 트리에서만 사라진다.

**autofix 프로토콜 편입 (감사 P2 — 오케스트레이터 결정)**: 초판은 watch 프로토콜만
보존 대상으로 삼았으나, 위 근거는 autofix 프로토콜에 똑같이 적용된다 — 이 파일은
`scripts/ci-autofix/log-fetch.sh`·`classify.sh`를 실행 가능 전제로 명시하며 스크립트
참조 6건을 보유한다. 한쪽만 보존하면 "스크립트는 남기고 그 문서는 지운다"는 모순이 남는다.

**의도적 비대칭 (오류 아님)**: `ci-autofix-protocol.md`는 템플릿에서 **재작성**되고
미러에서는 **원형 보존**된다. 두 처우 모두 "규칙은 그것이 규율하는 대상이 존재하는 곳에만
있어야 한다"는 하나의 원칙에서 나온다 — 배포 사용자에게는 스크립트가 없고, 개발 저장소에는 있다.

**수용된 결과: 등록 없는 Frozen 파일 (감사 P3)**

M2가 양쪽 레지스트리에서 `CONST-V3R5-014..021`을 삭제하므로, 보존된 미러
`ci-watch-protocol.md`는 **헌법 등록이 0인 채 `[ZONE:Frozen]` 마커 8개를 보유**하게 된다.
보존 결정을 autofix까지 넓히면서 같은 상태가 `ci-autofix-protocol.md`에도 생긴다
(마커 10개 — 단 이쪽은 `004..013`이 유지되므로 등록이 남는다).

관측된 증거 — 검증기는 이 상태를 **탐지하지 못한다**:

```bash
# 레지스트리 항목만 삭제, 소스 유지
moai constitution validate   # → validate-exit=1
grep -c 'ZONE_UNREGISTERED' /tmp/cv2.txt        # → 0
grep -c 'canary_gate: true' …/zone-registry.md  # → 65  (AC-CLD-011의 통과값)
```

`ZONE_UNREGISTERED`가 발화하지 않으므로 AC-CLD-009·010·011 어느 것도 이 상태를 잡지 못한다.
**이는 추가할 기준이 아니라 기록할 결정이다** — 미러의 watch 프로토콜은 dev-only 문서로
남고, 그 Frozen 마커는 배포되지 않으므로 헌법적 구속력을 갖지 않는다.
근거 있는 수용이며 우연한 누락이 아니다.

**저장소 루트 (1)**

```
CLAUDE.local.md                                        dev-only 일람에 5개 경로 등재
```

`CLAUDE.local.md` 편입 근거 (감사 N7 해소): M7 3단계가 이 파일 편집을 요구하는데
초판 봉투에는 없었고, DoD는 봉투 밖 파일 수정을 금지했다 — M7을 수행하면 DoD 위반,
건너뛰면 REQ-CLD-016 미충족이라는 모순이었다. 봉투에 편입해 해소한다.

**Go 측 (5)**

```
internal/cli/update_archive.go                   legacySkillIDs 항목 추가
internal/template/embed_ci_loop_guard_test.go    신설 (AC-CLD-003)
internal/cli/pr_watch_cmd.go                     문자열만 (41, 72, 73행)
internal/cli/update_archive_test.go              확장 (AC-CLD-019 가드 — 필수)
internal/cli/pr_watch_cmd_test.go                신설 (AC-CLD-018 보강 — 선택)
```

**시험 기대값 정합 (5) — M3 삭제의 귀결**

```
internal/template/skills_manifest_test.go        기대 코어 스킬 목록에서 항목 제거
internal/template/catalog_tier_audit_test.go     expectedSkillCount  32 → 31
internal/template/embed_catalog_test.go          wantTotal           42 → 41
internal/template/catalog_loader_test.go         expectedTotal       42 → 41
internal/template/sanitized_pair_parity_test.go  sanitizedPairPaths에서 ci-watch-protocol.md 행 제거
```

**시험 파일 봉투 편입 근거 (v0.3.0)**: AC-CLD-018·AC-CLD-019 승격으로 run-phase가
시험 코드를 작성할 수 있어야 하는데 초판 봉투에는 두 경로가 없었다 — DoD가 "봉투 밖
파일을 건드리지 않음"을 요구하므로 편입하지 않으면 시험 작성이 곧 DoD 위반이 된다
(`CLAUDE.local.md` 편입과 같은 종류의 모순).

- `internal/cli/update_archive_test.go` — **기존 파일 확장, 필수**.
  AC-CLD-019가 `TestArchiveSkill_PreservesUserContent`를 이름으로 지목하므로
  이 시험은 반드시 작성된다. `TestArchiveSkill_*` 계열이 이미 이 파일에 모여 있어
  형제 파일을 새로 만들지 않는다.
- `internal/cli/pr_watch_cmd_test.go` — **신설, 선택**. 현재 이 파일은 존재하지 않는다
  (`pr_watch_cmd.go`는 시험 파일이 없다). AC-CLD-018의 판정은 **재빌드된 바이너리에
  대한 CLI 배치**이므로 Go 시험을 요구하지 않는다. 다만 플래그 집합 단언은
  `newPRWatchCmd()`를 in-process로 호출해 검사할 수 있으므로, run-phase가 이를
  회귀 가드로 남기기로 하면 이 경로에 작성한다. 작성하지 않아도 AC-CLD-018은 통과한다.

**시험 기대값 5개 파일 봉투 편입 근거 (M1-M3 착수 후 해소)**: 이 다섯 파일은 삭제 이전
상태를 기대값으로 **하드코딩**하고 있어, M3의 삭제가 랜딩한 순간 `go test ./...`가
깨진다. DoD는 "`go test ./...` 회귀 없음"을 요구하는데 봉투는 이 파일들을 금지했다 —
고치면 DoD 위반, 두면 DoD 미충족이라는 모순이었다. `CLAUDE.local.md`(감사 N7)와
시험 파일 2개(v0.3.0)에서 이미 두 번 인정한 것과 **구조적으로 같은 모순**이며,
결론도 같다: **봉투가 틀렸다.** 편입해 해소한다.

M1-M3 수행 중 `manager-develop`가 이 충돌을 발견하고 자체 판단으로 범위를 넓히는 대신
블로커로 반환한 것은 올바른 처신이었다 — 봉투 확대는 계획의 결정이지 실행자의 결정이 아니다.

이 다섯은 **독립적 변경이 아니라 M3 삭제의 기계적 귀결**이다. 스킬 1개가 사라졌으므로
카탈로그 항목 수와 디스크 스킬 수가 각각 1 줄고(42→41, 32→31), 기대 목록에서 그 이름이
빠지고, 배포에서 사라진 파일이 sanitized-pair 레지스트리에 남을 이유가 없어진다.
**동작 변경은 없다** — 상수와 목록의 수 맞추기뿐이며, 새 단언·새 판정·새 능력을
도입하지 않는다. 그래서 신규 AC를 만들지 않고 기존 DoD 줄
("`go test ./...` 회귀 없음")이 그대로 이들을 덮는다.

**`sanitized_pair_parity_test.go` — 의존 조사 결과 (제거 안전)**

`manager-develop`가 정당한 우려를 제기했다: 이 레지스트리 행은 `ci-watch-protocol.md`가
**sanitized pair**(양쪽 트리에 존재하되 §25 정화 때문에 바이트가 다른 짝)라고 단언하는데,
그 단언은 양쪽에 파일이 있을 때만 참이었다. 템플릿 사본이 의도적으로 사라졌으므로
이 행은 **설계상 낡은(stale by design)** 상태다 — 테스트 자체가 그 상황에 대해
"짝이 아니면 `sanitizedPairPaths`에서 빼라"고 지시한다.

다른 소비자가 있는지 실측했다 (`grep -rn 'sanitizedPairPaths' --include='*.go' .`):

```
sanitized_pair_parity_test.go:48   선언부 주석
sanitized_pair_parity_test.go:63   var 선언
sanitized_pair_parity_test.go:146  유일한 순회 지점
sanitized_pair_parity_test.go:164  실패 메시지 문자열
sanitized_pair_parity_test.go:200  실패 메시지 문자열
```

**단일 파일 안에서만 쓰인다** — 다른 파일·다른 테스트·다른 패키지의 의존은 0건이다.
행 제거는 안전하다.

다만 **주석 수준 잔재 2건**이 남는다 (코드 의존 아님, 판단은 run-phase에 맡긴다):

- `sanitized_pair_parity_test.go` 상단 — 정화 방식을 설명하는 **예시**로 이 파일명을 든다
- `rule_template_mirror_test.go` — 바이트 파리티 허용목록에서 제외된 파일 **내력 주석**에 등재

둘 다 산문이며 어떤 단언도 구동하지 않으므로 `go test`에 영향이 없다. 방치해도 무해하고,
정리하면 더 정확하다.

`pr_watch_cmd.go` 봉투 편입 근거 (감사 D4 해소): 이 파일은 배포 바이너리에서
`scripts/ci-watch/run.sh`를 **사용자에게 직접 지시하는 문자열 3개**를 보유한다.
이를 남기면 다른 모든 AC가 통과한 뒤에도 `moai pr watch`를 실행한 사용자는
존재하지 않는 스크립트를 실행하라는 안내를 받는다 — 이 SPEC이 제거하려는 결함 그 자체다.
spec.md §4는 **동작·플래그·종료 코드** 변경만 제외하므로 문자열 수정은 제외 대상이 아니며,
초판의 "`internal/cli/pr*.go` 금지" 조항과 spec.md 사이의 모순은 이로써 해소된다.
허용 범위는 문자열 리터럴과 `Long` 도움말 텍스트에 한정한다.
문자열 제거 자체는 **AC-CLD-006**이 판정한다.

> **판정 공백 → 종결 (v0.3.0)**: `RunE` 분기·플래그 정의·반환값의 불변성은 이제
> **AC-CLD-018이 판정한다.** 경위를 남긴다.
>
> AC-CLD-006은 문자열 경로만 grep하며 동작에 대해 아무 말도 하지 않는다. v0.2.0은
> 이 공백을 **의도적 범위 선택**으로 남겼다 — 판정 불가능해서가 아니었다 (감사 P8).
> 실제로 M6의 문자열 전용 편집에 대해 안정적인 4개 단언이 존재하고 저작 시점에
> 실행 확인했다:
> ```
> moai pr watch 999 --branch main            → exit 0
> moai pr watch --help | grep -oE '--(abort|branch|report)'  → 3개 플래그
> moai pr watch --abort                      → exit 1   (상태 파일 부재 시)
> moai pr watch 999 --report                 → exit 0
> ```
> v0.2.0은 이를 AC로 채택하지 않고 리뷰어 diff 확인에 맡겼다. **run-phase 진입 시점에
> GOOS가 승격을 결정해 AC-CLD-018로 채택했다** — 네 단언 전부를 명령이 단언하고,
> `make build` 선행과 상태 파일 부재 전제를 명령에 넣어 공허 통과를 막았다.
> 따라서 "동작 불변을 판정하는 AC는 없다"는 서술은 더 이상 유효하지 않다.
>
> 초판(v0.1.0)은 이 자리에서 "AC-CLD-012가 판정한다"고 적었으나 그 AC는 레지스트리
> diff이며 이 파일과 무관하다 — 그 오기재는 v0.2.0에서 이미 정정되었다.

**봉투 밖 — 명시적 금지**

```
scripts/ci-watch/**        scripts/ci-autofix/**     dev-only 자산, 보존
internal/ciwatch/**        watch 로직 이식은 별도 SPEC
.github/workflows/ci.yml   .github/required-checks.yml
internal/constitution/**   검증 엔진 자체
internal/cli/pr_watch_cmd.go 의 RunE 분기 / 플래그 / 종료 코드
```

## §C 마일스톤

되돌리기 어려운 결정을 앞에 둔다.

---

### M1 — 재작성 문구 확정

**owner**: manager-develop

방향은 GOOS 결정 A·B로 확정되었으므로 초판의 M1(방향 확인)은 소멸했다.
남은 선결 작업은 하나다.

- `ci-autofix-protocol.md` 재작성 문구를 확정한다. 트리거 서술은 스크립트도 CLI도
  이름하지 않고 **오케스트레이터 핸드오프**를 지목한다.
  스크립트는 배포되지 않고, CLI는 watch를 수행하지 않으므로(research.md §B),
  배포 환경에서 참인 서술은 이것뿐이다.

> **폐기된 문구**: `moai pr watch reports a required-check failure (exit 2)`.
> `os.Exit(2)`는 CLI에 존재하지 않는다. 채택했다면 배포 규칙과
> Frozen 절 `CONST-V3R5-014`에 거짓 서술이 기입되었을 것이다.

**완료 조건**: `CONST-V3R5-004`·`013`의 새 절 텍스트가 결정됨.

**의존**: 없음

---

### M2 — 헌법 정합화 (양쪽 레지스트리)

**owner**: manager-develop

research.md §C.4의 3분류를 저장소 루트와 템플릿 레지스트리 **양쪽에** 적용한다.

1. `CONST-V3R5-014..021` (8개) — 레지스트리 항목 **삭제**.
   소스 파일이 M3에서 배포 제거되므로 항목을 남기면 `SOURCE_FILE_MISSING`(exit 2)
2. `CONST-V3R5-004`, `013` (2개) — M1 확정 문구로 소스와 레지스트리를 함께 재작성
3. `CONST-V3R5-005..012` (8개) — 소스와 레지스트리 텍스트를 일치시켜 기존 드리프트 해소

`canary_gate: true` 총계는 73 → **65**로 감소한다 (`014..021` 8개 전부가
`canary_gate: true`임을 실측 확인). 감소 사유는 "주제가 소멸한 절의 제거" 하나뿐이며,
다른 사유의 감소는 회귀다 (AC-CLD-011).

`MOAI_CONSTITUTION_SKIP_VALIDATE` 우회는 사용하지 않는다.

**완료 조건**: AC-CLD-009, AC-CLD-010, AC-CLD-011, AC-CLD-012.

**의존**: M1

---

### M3 — 배포 중단 (스킬 + watch 프로토콜)

**owner**: manager-develop

1. `internal/template/templates/.claude/skills/moai-workflow-ci-loop/` 삭제
2. `internal/template/templates/.claude/rules/moai/workflow/ci-watch-protocol.md` 삭제
3. `internal/template/catalog.yaml`의 `moai-workflow-ci-loop` 항목(5행) 삭제
4. `delegation.yaml` 3곳에서 스킬 참조 제거
5. `sync/delivery.md` **3곳**(239, 458, 465행)의 스킬 호출·서술 제거.
   239행은 `Skill("moai-workflow-ci-loop")` 실호출이며 대체 호출 대상은 없다
   (초판은 이를 1곳으로 잘못 기재했다)
6. `moai/SKILL.md`, `fix.md`, `loop.md`의 스킬 서술 제거
7. `ci-autofix-protocol.md`를 M1 확정 문구로 재작성
8. `manager-develop.md`, `manager-develop-prompt-template.md`, `cadence-bridge.md`,
   `run.md`의 스크립트 경로 참조 제거

**완료 조건**: AC-CLD-001, AC-CLD-002, AC-CLD-005, AC-CLD-007, AC-CLD-016.

**의존**: M1

---

### M4 — 재빌드 및 도달성 가드

**owner**: manager-develop

```bash
make build
go test ./internal/template/... -count=1
```

`internal/template/embed_ci_loop_guard_test.go`를 신설한다. `EmbeddedTemplates()`가
반환하는 컴파일된 FS를 순회해 `.claude/skills/moai-workflow-ci-loop/` 하위 파일 수가
0임을 단언한다 (AC-CLD-003).

> **이 테스트의 한계** (감사 D3 해소): `go test`가 컴파일을 수행하므로 `//go:embed`는
> 항상 현재 소스를 반영한다. 따라서 **`make build` 누락을 탐지하지 못한다** —
> 초판이 주장했던 근거는 거짓이며 변이 테스트로 반증되었다(research.md §D.1).
> 재빌드 누락 탐지는 catalog 해시 신선도(AC-CLD-004)가 담당하며,
> 그 게이트는 변이 테스트로 실패 가능성이 실증되었다(research.md §D.2).

**시험 기대값 정합 5건** (M3 삭제의 귀결 — §B 편입 근거 참조). 이 마일스톤이 소유한다:
재빌드·가드 신설과 함께 트리를 다시 초록으로 만드는 작업이 M4의 일이기 때문이다.

1. `skills_manifest_test.go` — 기대 코어 스킬 목록에서 `moai-workflow-ci-loop` 제거
2. `catalog_tier_audit_test.go` — `expectedSkillCount` 32 → 31
3. `embed_catalog_test.go` — `wantTotal` 42 → 41
4. `catalog_loader_test.go` — `expectedTotal` 42 → 41
5. `sanitized_pair_parity_test.go` — `sanitizedPairPaths`에서
   `.claude/rules/moai/workflow/ci-watch-protocol.md` 행 제거 (다른 소비자 0건 실측 확인)

상수 옆 주석이 누적 산술(`… net +1 = 42`)을 기록하는 형식이므로, 값만 바꾸고 주석을
그대로 두면 산술이 어긋난다. 값과 함께 해당 주석 줄도 갱신한다.

**완료 조건**: AC-CLD-003, AC-CLD-004, 그리고 `go test ./internal/template/... -count=1` 초록
(DoD의 "`go test ./...` 회귀 없음"이 최종 판정 — 신규 AC 없음).

**의존**: M3

---

### M5 — 배포된 프로젝트의 고아 스킬 처리

**owner**: manager-develop

`internal/cli/update_archive.go`의 `legacySkillIDs`에 `"moai-workflow-ci-loop"` 추가.

**선결 제약**: `TestLegacySkillIDsNotEmbedded`가 임베드 집합과의 분리를 강제하므로
**M4 완료 후에만** 추가할 수 있다 (초판의 "M2 의존" 표기는 오류).
순서를 뒤집으면 테스트가 실패한다 — 문체상 권고가 아니라 테스트가 강제하는 제약이다.

**결정 (GOOS, run-phase 진입 시점 — v0.3.0에서 종결)**: 기존 `archiveVersion` 상수
`"v2.16"`을 **그대로 재사용한다 (선택지 a)**.

검토된 대안은 (b) 스킬별 버전 태그 일반화였다 — 경로 라벨은 정확해지나 범위가 커진다.
(a)를 채택한 근거: 아카이브 경로의 목적은 **보존**이고 버전 태그는 라벨에 불과하다.
경로가 `v2.16`이라는 이유로 사용자 저작물이 덜 보존되지는 않는다. (b)는 배포 격리라는
이 SPEC의 목적과 무관한 리팩터링이므로 별도 SPEC 소관이다. 미결 항목이 아니라 결정이다.

**완료 조건**: AC-CLD-015, AC-CLD-019.

**의존**: M4

---

### M6 — 바이너리 문자열 교정

**owner**: manager-develop

`internal/cli/pr_watch_cmd.go`의 사용자 노출 문자열 3곳을 교정한다.

- **41행** (`Long` 도움말) — 미구현 계약 서술 제거.
  현재 이 텍스트는 `exit 2` JSON 핸드오프와 30분 타임아웃을 주장하지만
  `RunE`는 둘 다 구현하지 않는다(research.md §B.3). 실제 제공 모드
  (`--abort`, `--report`)만 서술한다
- **72·73행** (stderr 안내문) — 부재 스크립트 실행 지시 제거

동작·플래그·종료 코드는 불변이다. 107행은 주석이므로 사용자 노출이 아니며 대상이 아니다.

불변성은 이제 **AC-CLD-018이 판정한다** (v0.3.0 승격). 편집 후 `make build`로 재빌드한
바이너리에 대해 네 관측값(기본 모드 종료 코드·플래그 3개·`--abort`·`--report`)을 확인한다.
문자열 3곳 중 72·73행은 `RunE` 본문 안에 있어 편집 사고 반경 안이므로, 이 판정은
형식적 절차가 아니라 실제 회귀 가드다.

**완료 조건**: AC-CLD-006, AC-CLD-018.

**의존**: M1

---

### M7 — 미러 동기화 및 중립성 검증

**owner**: manager-develop

1. `.claude/` 미러 **10개** 파일을 템플릿 변경과 동일하게 반영 (§B 미러 목록)
2. 선언된 dev-only 보존 집합 5경로는 **보존** — `scripts/ci-watch/`, `scripts/ci-autofix/`,
   `.claude/skills/moai-workflow-ci-loop/`, `.claude/rules/moai/workflow/ci-watch-protocol.md`,
   `.claude/rules/moai/workflow/ci-autofix-protocol.md`
3. `CLAUDE.local.md` 로컬 전용 일람에 위 5개 경로를 dev-only로 등재
4. 중립성 검증 실행

**완료 조건**: AC-CLD-013, AC-CLD-014, AC-CLD-017.

**의존**: M2, M3, M6

---

## §D 미해소 항목

초판이 남긴 미해결 명확화 항목 2건은 **모두 종결**되었다 (마커 리터럴은 제거됨).

- **required-checks.yml** — GOOS 결정 B로 이연. 배포 트리에서 이 전제를 요구하는
  참조 4건이 전부 M2·M3의 제거 범위 안에 있고, **그 외 어떤 배포 파일도 이 전제를
  요구하지 않음**을 실측 확인했다(research.md §G.1). 본 SPEC 완료 시 소멸이 예상되며
  AC-CLD-016이 판정한다.
- **검증 엔진 버전** — 감사가 워크트리 소스로 빌드한 바이너리에서 재측정해
  릴리스 바이너리와 **동일 출력(전체 77 / ci 귀속 18)** 을 확인했다.
  acceptance.md의 baseline은 유효하다(research.md §G.2).

잔여 gap은 research.md §H에 기록한다 (실배포 산출물 미관측,
required-checks.yml의 사용자 측 생성 경로 자체).

## §E 위험

| 위험 | 영향 | 완화 |
|---|---|---|
| M5를 M4보다 먼저 수행 | `TestLegacySkillIDsNotEmbedded` 실패 | 의존 순서 강제 (§C M5) |
| 레지스트리와 소스를 따로 편집 | `SOURCE_FILE_MISSING`(exit 2) 또는 `ZONE_UNREGISTERED` | 동일 커밋 원칙 (REQ-CLD-011), AC-CLD-009 + AC-CLD-010 (validate-exit 포착) |
| 한쪽 레지스트리만 수정 | 템플릿·저장소 레지스트리 발산 | AC-CLD-012 (양쪽 diff + 행 수 판정) |
| 소스만 편집하고 `make build` 누락 | catalog 해시 낡음 | AC-CLD-004 (변이 검증 완료) |
| 미러 동기화 시 dev-only 자산 동반 삭제 | ci-loop 개발 능력 상실 | AC-CLD-013 (9개 파일 전수) |
| 스킬 제거 후 호출부 방치 | 존재하지 않는 스킬을 호출하는 죽은 지시 | M3 4~6단계가 호출부 전수를 대상 |
| 재작성 문구에 거짓 서술 기입 | 배포 규칙·Frozen 절 오염 | M1이 트리거를 오케스트레이터 핸드오프로 확정 |
| M6 편집 전 바이너리로 AC-CLD-018 판정 | 불변 단언이 편집을 관측하지 못하고 공허 통과 | 판정 명령 첫 줄이 `make build`이며 `build-exit=0`이 통과 조건 |

## §F 반패턴

- `MOAI_CONSTITUTION_SKIP_VALIDATE=1`로 검증 우회
- `zone-registry.md`의 `clause` 텍스트를 손으로 추측해 기입 (소스에서 복사할 것)
- `catalog.yaml`의 `hash`를 손으로 계산하거나 삭제 (생성기가 관리)
- 소스 트리 grep만으로 "배포되지 않음"을 판정
- **도움말 산문을 구현의 근거로 삼기** — 초판이 범한 오류.
  `--help`는 저자의 의도를 담고 구현은 담지 않는다. 명령을 실행하거나 `RunE`를 읽을 것
- 스크립트를 언어 중립화해 템플릿에 포함시키는 방향 (명시적 기각)

## §G 교차 참조

- spec.md — GEARS 요구사항 19건
- acceptance.md — 판정 명령과 관측된 baseline
- research.md — 4개 질문의 실측 근거, 반증 기록, 변이 테스트

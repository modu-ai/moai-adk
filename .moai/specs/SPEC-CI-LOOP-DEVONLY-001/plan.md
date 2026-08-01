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

**Go 측 (3)**

```
internal/cli/update_archive.go                   legacySkillIDs 항목 추가
internal/template/embed_ci_loop_guard_test.go    신설
internal/cli/pr_watch_cmd.go                     문자열만 (41, 72, 73행)
```

`pr_watch_cmd.go` 봉투 편입 근거 (감사 D4 해소): 이 파일은 배포 바이너리에서
`scripts/ci-watch/run.sh`를 **사용자에게 직접 지시하는 문자열 3개**를 보유한다.
이를 남기면 다른 모든 AC가 통과한 뒤에도 `moai pr watch`를 실행한 사용자는
존재하지 않는 스크립트를 실행하라는 안내를 받는다 — 이 SPEC이 제거하려는 결함 그 자체다.
spec.md §4는 **동작·플래그·종료 코드** 변경만 제외하므로 문자열 수정은 제외 대상이 아니며,
초판의 "`internal/cli/pr*.go` 금지" 조항과 spec.md 사이의 모순은 이로써 해소된다.
허용 범위는 문자열 리터럴과 `Long` 도움말 텍스트에 한정한다.
문자열 제거 자체는 **AC-CLD-006**이 판정한다.

> **판정 공백 (명시)**: `RunE` 분기·플래그 정의·반환값의 불변성을 판정하는 AC는 **없다.**
> AC-CLD-006은 문자열 경로만 grep하며 동작에 대해 아무 말도 하지 않는다.
> 이는 **범위 선택**이며 판정 불가능해서가 아니다 (감사 P8). 실제로 M6의 문자열 전용
> 편집에 대해 안정적인 4개 단언이 존재하고, 저작 시점에 실행 확인했다:
> ```
> moai pr watch 999 --branch main            → exit 0
> moai pr watch --help | grep -oE '--(abort|branch|report)'  → 3개 플래그
> moai pr watch --abort                      → exit 1
> moai pr watch 999 --report                 → exit 0
> ```
> 이 SPEC은 이를 AC로 채택하지 않고 리뷰어 diff 확인에 맡긴다 — run-phase에서
> 저렴하게 승격할 수 있다.
> 초판은 이 자리에서 "AC-CLD-012가 판정한다"고 적었으나 그 AC는 레지스트리 diff이며
> 이 파일과 무관하다.

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

**완료 조건**: AC-CLD-003, AC-CLD-004.

**의존**: M3

---

### M5 — 배포된 프로젝트의 고아 스킬 처리

**owner**: manager-develop

`internal/cli/update_archive.go`의 `legacySkillIDs`에 `"moai-workflow-ci-loop"` 추가.

**선결 제약**: `TestLegacySkillIDsNotEmbedded`가 임베드 집합과의 분리를 강제하므로
**M4 완료 후에만** 추가할 수 있다 (초판의 "M2 의존" 표기는 오류).
순서를 뒤집으면 테스트가 실패한다 — 문체상 권고가 아니라 테스트가 강제하는 제약이다.

**미결 결정**: `archiveVersion`이 `"v2.16"`으로 하드코딩되어 있다.
(a) 기존 상수 재사용 — 변경 최소, 경로 라벨 부정확
(b) 스킬별 버전 태그 일반화 — 정확하나 범위 확대
기본 권고는 **(a)**. 아카이브 경로의 목적은 보존이고 버전 태그는 라벨에 불과하며,
(b)는 배포 격리라는 이 SPEC의 목적과 무관한 리팩터링이다.

**완료 조건**: AC-CLD-015.

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

**완료 조건**: AC-CLD-006.

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

# progress — SPEC-CI-LOOP-DEVONLY-001

## §E.1 Plan-phase Audit-Ready Signal

**Tier**: M (근거: plan.md §A) — 감사 후에도 유지

**산출물**

| 파일 | 상태 |
|---|---|
| spec.md | v0.2.0 — GEARS 요구사항 19건, 범위 제외 5개 항목 (required-checks 이연 포함) |
| plan.md | v0.2.0 — 마일스톤 M1~M7, EXTEND 봉투(pr_watch_cmd.go 문자열 편입), 위험 7건 |
| acceptance.md | v0.2.0 — AC-CLD-001..016, REQ↔AC 커버리지 표, baseline 출처 개별 표기 |
| research.md | v0.2.0 — §B 자기반증, §F.2 결정 A, §G.1/§G.2 마커 종결, §H 잔여 gap |
| progress.md | 본 파일 |

**감사 대응 (FAIL 0.56 → 9개 결함 전부 반영)**

| # | 결함 | 조치 | 검증 |
|---|---|---|---|
| 1 | frontmatter `tags:` 시퀀스 → 12필드 전체 디코드 실패 | 문자열로 교정 | `moai spec lint` → `✓ No findings` |
| 2 | §B가 `moai pr watch` 오독에 기반 | §B 전면 재작성, 폐기 문구 명시 | 명령 실행 + `RunE` 소스 확인 |
| 3 | AC 선택자가 미존재 테스트명 지목 | 실제 함수명으로 교정 + `-v` | 공허 통과 재현 후 교정판 실행 |
| 4 | AC-003 근거 거짓 (재빌드 탐지 불가) | 한계 명시 + AC-004 해시 게이트 신설 | 변이 테스트 통과 |
| 5 | 바이너리 문자열 3개 vs 봉투 금지 모순 | 봉투에 문자열만 편입 + AC-006 | 선택자 실행, 3건 원문 인용 |
| 6 | 템플릿 레지스트리 판정 AC 부재 | AC-012 양쪽 diff 신설 | 변이 테스트 통과 |
| 7 | 스크립트 9개 중 3개만 보호 | `find -type f` 전수 | `9` 관측 |
| 8 | REQ 4건 AC 미커버 | 커버리지 표 + AC 신설 | 미커버 0 |
| 9 | 미해결 명확화 항목 2건 | 둘 다 종결 (결정 B / 엔진 동일), 마커 리터럴 제거 | 참조 4건 전수 조사 |

**실행한 검증 (이번 판)**

| 항목 | 명령 | 관측 |
|---|---|---|
| SPEC lint | `moai spec lint …/spec.md` | `✓ No findings` (수정 전: `ParseFailure` line 13) |
| pr watch 실동작 | `moai pr watch 999 --branch main` | 스크립트 안내문 출력, `EXIT=0` |
| exit 2 부재 | `grep -rn 'os.Exit(2)' pr_watch_cmd.go internal/ciwatch/*.go` | 매치 0 |
| 테스트명 공허성 | `go test -run 'TestInternalContentLeak'` | `no tests to run` + PASS |
| 교정 선택자 | `go test -run 'TestTemplateNoInternalContentLeak\|TestSplitHarnessNamespaceNoLeak' -v` | 둘 다 `--- PASS` |
| 해시 게이트 변이 | 생성기 + `git diff --exit-code` | BASELINE `0` / MUTATED `1` / 복원 확인 |
| 레지스트리 diff 변이 | ci 절 diff | 일치 `0` / 변형 `1` / 복원 `0` |
| 스크립트 전수 | `find scripts/ci-* -type f \| wc -l` | `9` |
| canary_gate | `grep -c 'canary_gate: true'` | `73` (양쪽 레지스트리) / `014..021` 중 `8` |
| 바이너리 문자열 | `grep -n … \| grep -vc ':[[:space:]]*//'` | `3` |
| required-checks | `grep -rln 'required-checks.yml' templates/ \| wc -l` | `4` |

**iteration-3 감사 대응 (FAIL 0.63 → 게이트 계층 10건)**

| # | 결함 | 조치 | 변이 검증 (양쪽 상태) |
|---|---|---|---|
| N1 | AC-004가 **삭제**를 못 잡음 (M3의 실제 동작) | `gen-exit`를 통과 조건에 추가 + 공백 정규화 한계 명시 | clean `gen0/diff0` · 삭제 `gen1/diff0`(낡은 해시 잔존) · 공백만 `gen0/diff0` |
| N2 | plan.md AC 참조 11/15 오류 | 실제 heading 목록으로 재도출 후 전건 교정 | `grep '^## AC-CLD-'` 대조 |
| N3 | AC-012 baseline == 통과값 | 행 수(18→10)를 명령에 편입 | 일치 `0` · 변형 `1` · 복원 `0` |
| N4 | 미러 동기화 무판정 (11파일) | AC-CLD-017 신설, agent-memory 제외 | 원시 14 → 13 → 12 → **11** |
| N5 | `CLAUDE.local.md` 봉투 누락 → DoD 모순 | §B에 저장소 루트 항목 신설 | — |
| N6 | AC-007 줄바꿈 회피 + baseline 과대 | 인접성 불요 선택자로 교체 (baseline 5) | 한 줄 `1→2` 탐지 · 줄바꿈 `1` 미탐지 |
| N7 | research.md 유지-방향 잔재 | §B.4 결론에 맞춰 삭제·재서술 | — |
| N8 | AC-009/010이 fatal abort를 진전으로 오독 | `validate-exit` 포착을 통과 조건에 추가 | clean `exit1/18/77` · 파탄 `exit2/11/49` |
| N9 | 커버리지 표 2건 과대 | REQ-016 재배치, REQ-018은 부채로 인정 | — |
| N10 | 낡은 식별자 6곳 + 마커 리터럴 3곳 | 전건 교정 + 리터럴 제거 | 게이트 verb grep `0` |

**추가 반증 (감사 지적, 자체 재현)** — research.md §C.1의 "두 방향 모두 막혀 있다"는 거짓이다.
레지스트리 항목 8개만 지우고 소스를 남기면 `ZONE_UNREGISTERED=0` · `validate-exit=1`이며,
그 상태에서 `canary_gate`는 정확히 **65** — AC-CLD-011의 통과값 — 인 채 소스가 고아로 남는다.
도구가 막는 것은 한 방향뿐이고, 나머지 한 방향은 AC-CLD-005 + AC-CLD-007 조합이 막는다.

**iteration-4 (최종 감사 PASS-WITH-FIXES 0.84) 대응**

| # | 항목 | 조치 | 변이/실행 증거 |
|---|---|---|---|
| P1 | AC-003이 여전히 실패 불가 (공허 선택자 3회차) | `grep -q '^--- PASS: <name>'` 형태로 교체 | 부재 `exit=1` 거부 · 대조군 존재+PASS `exit=0` 수용 (초판 형태는 부재에도 `exit=0`) |
| P2 | AC-013 ↔ AC-017 상충 | 미러 `ci-autofix-protocol.md`를 보존 집합에 편입 (오케스트레이터 결정) | AC-017 재도출 `11→10`, 도달성 변이 `BEFORE=10 AFTER=0 RESTORED=10` |
| P3 | 등록 없는 Frozen 파일 발생 | plan.md §B에 근거 있는 수용으로 기록 | `ZONE_UNREGISTERED=0` · `validate-exit=1` · `canary=65` 관측 |
| P4 | 커버리지 표 중복 행 | 낡은 행 제거, 정정 행만 유지 | `uniq -d` → 출력 없음 |
| P5 | AC-008 마일스톤 미배정 | M3 7단계로 배정 | — |
| P6 | research.md 마지막 낡은 식별자 | AC-014 → AC-016 | `grep -c 'AC-CLD-014'` → 0 |
| P7 | AC-004의 커밋 상태 의존 | 판정 시점 전제 명시 | — |
| P8 | RunE 불변성 "판정 불가" 과장 | 범위 선택으로 완화 + 4개 단언 기록 | `exit 0` / 3개 플래그 / `abort exit 1` / `report exit 0` |
| P9 | REQ-018 이연 사유 부정확 | 단위 시험 가능함을 인정, 사유를 범위 한정으로 정정 | `archiveSkill(projectRoot, skillID)` 시그니처 확인 |
| P10 | research §D.2가 내용 변이만 기록 | 4개 상태(A~D) 전부 기록 + N1 원인 명시 | — |

**감사 지적 중 이미 해소된 항목**: research.md §C.1의 "두 방향 모두 막혀 있다"는
iteration-3에서 이미 정정되어 154행이 "한 방향만 막혀 있다"로 시작하며 옛 서술을
반증된 것으로 인용한다. 추가 조치 불요.

**핵심 반증 (초판 자기정정)** — `moai pr watch`는 watch 루프를 수행하지 않는다.
기본 모드는 셸 스크립트 실행 안내문을 출력하고 exit 0으로 끝나며, CLI에 `os.Exit(2)`는
없다. `--help`의 `Long` 텍스트가 서술하는 계약을 `RunE`가 구현하지 않는 것이 원인이다.
초판은 도움말 산문을 구현의 증거로 삼아 정반대 결론을 냈다.
이 반증이 결정 A(두 프로토콜 파일 분리 처리)의 근거가 된다.

**미해소 (research.md §H)**

- 실제 `moai init` 배포 산출물을 관측하지 못했다 (임베드 FS + catalog 해시로 대체 판정)
- `.github/required-checks.yml`의 사용자 측 생성 경로 자체는 미확인 (결정 B로 이연)

**v0.3.0 개정 (run-phase 진입 결정 반영, 범위 한정 amendment)**

GOOS 결정 3건 중 **#3(저비용 판정 승격)** 과 **#2(archiveVersion)** 를 산출물에 반영했다.
요구사항(GEARS 19건)과 범위 제외는 손대지 않았고 판정 계층만 확장했다.

| 변경 | 내용 |
|---|---|
| AC-CLD-018 신설 (M6) | `moai pr watch` 동작·플래그·종료 코드 불변. `make build` 선행 + 상태 파일 부재 전제를 명령에 포함. plan.md §B "판정 공백"은 종결로 재작성(경위 보존) |
| AC-CLD-019 신설 (M5) | `archiveSkill`이 삭제가 아니라 복사로 보존. `grep -q '^--- PASS: TestArchiveSkill_PreservesUserContent'` 형태 |
| 커버리지 표 | REQ-CLD-018 부채 → AC-CLD-019. **미커버 REQ 0건**. AC-CLD-018은 REQ가 아니라 spec.md §4 범위 가드임을 명시 |
| 봉투 (plan.md §B) | Go 측 3 → 5. `update_archive_test.go`(확장·필수), `pr_watch_cmd_test.go`(신설·선택) |
| M5 미결 결정 | `archiveVersion` "v2.16" 재사용으로 종결 (미결 항목 → 결정) |
| DoD | `AC-CLD-001 ~ AC-CLD-019`, 부채 이월 줄 제거 |

**AC-CLD-018 baseline 재실행** (이 개정 중 실측):
`default-exit=0` / `flags=3` / `abort-exit=1` / `report-exit=0` — v0.2.0 plan.md §B가
기록한 네 값과 일치. `abort-exit=1`은 `.moai/state/ci-watch-active.flag` 부재 조건에서의
값이므로 전제를 명령에 명시했다.

**AC-CLD-019 관측 범위 조정 (기록)**: "호출 후 원본 부재"는 단언하지 않는다 —
`archiveSkill`도 `archiveLegacySkills`도 원본을 제거하지 않으므로(복사 전용) 그 단언은
코드에 없는 모델을 기입하게 된다. 아카이브↔삭제 구분은 **아카이브 측 바이트 대조**가
담당하고, 복사→이동 회귀는 **원본 잔존 확인**이 막는다.

## §E.2 Run-phase Evidence

### M1 — 재작성 문구 확정 (완료)

`ci-autofix-protocol.md`의 두 Frozen 절 문구를 확정했다. 트리거는 스크립트도 CLI도
이름하지 않고 **오케스트레이터 핸드오프**를 지목한다.

| 절 | 확정 문구 | 배포 사용자에게 참인 이유 |
|---|---|---|
| `CONST-V3R5-004` | `The CI auto-fix loop MUST be entered ONLY when the orchestrator hands off a failing required check` | 오케스트레이터는 배포되며 실패 정보를 어떤 경로로 얻든 autofix 사이클에 전달할 수 있다. 스크립트는 배포되지 않고 CLI는 watch 루프를 수행하지 않는다(research.md §B) |
| `CONST-V3R5-013` | `The auto-fix loop MUST NOT modify CI watch infrastructure scripts or workflow definitions` | 종전 문구는 미배포 경로 `scripts/ci-watch/run.sh`를 지목했다. 새 문구는 경로를 이름하지 않고 보호 **대상 범주**를 규정하므로 스크립트 유무와 무관하게 참이다 |

폐기 문구 `moai pr watch reports a required-check failure (exit 2)`는 채택하지 않았다
(`os.Exit(2)`가 CLI에 존재하지 않음).

### M2 — 헌법 정합화 (완료, 양쪽 레지스트리)

| 분류 | 절 | 처리 | 결과 |
|---|---|---|---|
| 삭제 | `CONST-V3R5-014..021` (8) | 레지스트리 항목 제거 | `canary_gate: true` 73 → 65 (양쪽) |
| 소스+레지스트리 재작성 | `CONST-V3R5-004`, `013` (2) | M1 확정 문구를 저장소 루트 미러 소스와 레지스트리에 동시 기입 | DRIFT 해소 |
| 텍스트 정합화 | `CONST-V3R5-005..012` (8) | 레지스트리 `clause`를 미러 소스 원문에서 복사 | 기존 드리프트 해소 |

`moai constitution validate`가 `file:` 을 프로젝트 루트 기준으로 해석하므로, 10개 절의
`clause`는 **저장소 루트 미러** `ci-autofix-protocol.md`에서 발견되어야 한다. 따라서
004·013의 새 문구를 미러 소스에도 기입했다 — plan.md §B가 허용한 "재작성된 소스 절
텍스트가 요구하는 경우"에 해당한다. 미러의 dev-repo 사실(스크립트 보유)은 후속 문장으로
보존했다.

### M3 — 배포 중단 (완료)

| 단계 | 대상 | 조치 |
|---|---|---|
| 1 | `templates/.claude/skills/moai-workflow-ci-loop/` | 디렉터리 삭제 |
| 2 | `templates/.claude/rules/moai/workflow/ci-watch-protocol.md` | 파일 삭제 |
| 3 | `internal/template/catalog.yaml` | 스킬 항목 5행 삭제 |
| 4 | `templates/.moai/config/sections/delegation.yaml` | sync·fix·loop 3곳에서 스킬 참조 제거 |
| 5 | `templates/.claude/skills/moai/workflows/sync/delivery.md` | **3곳** 제거 — Phase 14 단계 6(`Skill()` 실호출, 대체 대상 없음), Related Skills 절 전체, 버전 이력 문구 |
| 6 | `moai/SKILL.md`(3), `fix.md`(2), `loop.md`(1) | 스킬 서술·형제 프리셋 언급 제거 |
| 7 | `templates/.../ci-autofix-protocol.md` | M1 문구로 재작성. Frozen 마커 10개 보존, 스크립트 참조 0 |
| 8 | `manager-develop.md`, `manager-develop-prompt-template.md`(2), `cadence-bridge.md`(2), `run.md` | 스크립트 경로 참조 제거 |

5단계는 plan.md가 기록한 3곳이 맞았다(초판의 1곳 기재가 오류). 8단계에서
`cadence-bridge.md`는 스크립트 경로 1건 외에 교차참조 절의 `ci-watch-protocol.md` 링크도
제거했다 — 삭제된 파일을 가리키는 죽은 링크를 남기지 않기 위함이며, 같은 제거 작업의 일부다.

### AC PASS/FAIL 매트릭스 (M1-M3 담당분)

| AC | 상태 | 판정 명령 | 관측 출력 |
|---|---|---|---|
| AC-CLD-001 | PASS | `grep -rn 'scripts/ci-watch\|scripts/ci-autofix' internal/template/templates/ \| wc -l` | `0` (baseline 27) |
| AC-CLD-002 | PASS | `grep -rln 'moai-workflow-ci-loop' … \| wc -l` / `grep -c 'name: moai-workflow-ci-loop' catalog.yaml` | `0` / `0` (baseline 9 / 1) |
| AC-CLD-005 | PASS | `test ! -e …/ci-watch-protocol.md; echo $?` | `0` |
| AC-CLD-007 | PASS | `grep -rn 'moai pr watch' internal/template/templates/ \| wc -l` | `0` (baseline 5) |
| AC-CLD-008 | PASS | `test -e …/ci-autofix-protocol.md` / `grep -c 'ZONE:Frozen'` / `grep -c 'scripts/ci-*'` | `FILE_OK` / `10` / `0` |
| AC-CLD-009 | PASS | `moai constitution validate` + ci 귀속 grep | `validate-exit=1`, ci-count `0` (baseline exit1 / 18) |
| AC-CLD-010 | PASS | 동일 명령 + 전체 findings grep | `validate-exit=1`, total `59` (baseline 77) |
| AC-CLD-011 | PASS | `grep -c 'canary_gate: true' .claude/rules/moai/core/zone-registry.md` | `65` (baseline 73) |
| AC-CLD-012 | PASS | ci 절 clause diff + 행 수 | `diff-exit=0`, `root-lines=10 tmpl-lines=10` (baseline 18/18) |
| AC-CLD-013 | PASS | `find scripts/ci-watch scripts/ci-autofix -type f \| wc -l` + 스킬 디렉터리 | `9` / `SKILL_OK` (변화 없음) |
| AC-CLD-016 | PASS | `grep -rln 'required-checks.yml' internal/template/templates/ \| wc -l` | `0` (baseline 4) |

참고 (M1-M3 담당분 아님, 회귀 확인용): AC-CLD-014 두 테스트 모두 `--- PASS`,
`grep -rn 'SPEC-CI-LOOP-DEVONLY-001' templates/ catalog.yaml` → `0`.

### 보존 자산 무변화 확인

```
find scripts/ci-watch scripts/ci-autofix -type f | wc -l   → 9
.claude/skills/moai-workflow-ci-loop/                       → SKILL_OK
.claude/rules/moai/workflow/ci-watch-protocol.md            → MIRROR_WATCH_OK
.claude/rules/moai/workflow/ci-autofix-protocol.md          → MIRROR_AUTOFIX_OK (004·013 문구만 갱신)
```

### 빌드·회귀

```
go build ./...                            → exit 0
GOOS=windows GOARCH=amd64 go build ./...  → exit 0
```

`go test ./internal/template/... ./internal/spec/...` 는 현재 FAIL한다. 실패는 두 부류다.

| 부류 | 실패 테스트 | 해소 주체 |
|---|---|---|
| catalog 해시 신선도 | `TestManifestHashFormat`, `TestCatalogHashParity` | **M4** (`make build` → `gen-catalog-hashes --all`). 예상된 미결 상태 |
| 테스트 소스의 하드코딩된 기대값 | `TestEmbeddedMoaiSkillNames`, `TestAllSkillsInCatalog`(32→31), `TestLoadCatalog`/`TestLoadEmbeddedCatalog_Success`(42→41), `TestSanitizedPairParity`(`ci-watch-protocol.md` 레지스트리 엔트리) | **미배정 — 봉투 밖 (§E.2 블로커 참조)** |

### 블로커 — EXTEND 봉투 공백 (M1-M3에서 발견)

plan.md §B의 Go 측 봉투(5경로)에 다음 5개 테스트 파일이 없다. 이들은 삭제된 스킬·파일의
개수와 이름을 **테스트 소스에 하드코딩**하고 있어 `make build`로 해소되지 않는다.
DoD의 "`go test ./...` 회귀 없음"은 이 파일들을 편집하지 않으면 충족 불가능하다 —
`CLAUDE.local.md` 편입(감사 N7)·시험 파일 편입(v0.3.0)과 같은 종류의 모순이다.

| 파일 | 필요한 변경 |
|---|---|
| `internal/template/skills_manifest_test.go` | 기대 core 스킬 목록에서 `moai-workflow-ci-loop` 제거 |
| `internal/template/catalog_tier_audit_test.go` | 디스크 스킬 디렉터리 기대값 32 → 31 |
| `internal/template/embed_catalog_test.go` | 카탈로그 엔트리 기대값 42 → 41 |
| `internal/template/catalog_loader_test.go` | 동일 (42 → 41) |
| `internal/template/sanitized_pair_parity_test.go` | `sanitizedPairPaths`에서 `ci-watch-protocol.md` 제거 (템플릿 미러가 의도적으로 삭제됨) |

봉투를 확장하지 않고 범위를 넘지 않았다. 오케스트레이터가 봉투 편입을 승인하면
M4에서 `make build`와 함께 처리하는 것이 자연스럽다.

### M4 — 재빌드 및 도달성 가드 (완료, 커밋 `a21d406bf`)

봉투 확장(커밋 `b2c5fd1a6`)으로 시험 파일 5개가 편입되어 M1-M3의 블로커가 해소되었다.

| 단계 | 대상 | 조치 |
|---|---|---|
| 1 | `embed_ci_loop_guard_test.go` | 신설. `EmbeddedTemplates()` FS를 순회해 `.claude/skills/moai-workflow-ci-loop/` 하위 파일 0건 단언 |
| 2 | `skills_manifest_test.go` | 기대 코어 스킬 목록에서 항목 제거 |
| 3 | `catalog_tier_audit_test.go` | `expectedSkillCount` 32 → 31 (누적 산술 주석 동반 갱신) |
| 4 | `embed_catalog_test.go` | `wantTotal` 42 → 41 (동반 갱신) |
| 5 | `catalog_loader_test.go` | `expectedTotal` 42 → 41 (동반 갱신) |
| 6 | `sanitized_pair_parity_test.go` | `sanitizedPairPaths`에서 `ci-watch-protocol.md` 행 제거 + 제거 사유 주석 |
| 7 | `make build` | `catalog.yaml` 해시 2건 재생성 (`moai/SKILL.md`, `manager-develop.md` — M3 편집분) |

**가드의 한계를 코드 주석에 명시**: `go test`가 컴파일하므로 `//go:embed`는 항상 현재
소스를 반영한다 → **`make build` 누락을 탐지하지 못한다**. 재빌드 누락은 AC-CLD-004
해시 신선도가 담당한다. 초판이 이 가드에 붙였던 재빌드-탐지 근거는 거짓이며
변이 테스트로 반증되었다(research.md §D.1).

**주석 수준 잔재 (방치 결정)**: `sanitized_pair_parity_test.go` 상단이 정화 방식의
**예시**로 `ci-watch-protocol.md`를 계속 언급한다. 어떤 단언도 구동하지 않는 산문이고
정화 메커니즘 설명으로서는 여전히 유효하므로, 범위 규율에 따라 손대지 않았다.
`rule_template_mirror_test.go`의 내력 주석도 동일하게 방치했다.

### M5 — 고아 스킬 아카이브 (완료, 커밋 `f7303abbd`)

`legacySkillIDs`에 `"moai-workflow-ci-loop"` 추가. `archiveVersion` 은 `"v2.16"`
그대로 재사용(GOOS 결정 2) — 태그는 보존 위치의 라벨이지 은퇴 시점의 주장이 아니다.
목록 항목 옆에 이 근거를 주석으로 남겼다.

`TestArchiveSkill_PreservesUserContent` 신설. 관측 4항목:

1. `t.TempDir()` 아래 `.claude/skills/<id>/`에 식별 가능한 사용자 저작 내용
   (최상위 `SKILL.md` + 중첩 `references/user-playbook.md` — 재귀 복사를 덮기 위함)
2. `archiveSkill(root, id)` 호출
3. 아카이브 대상에 **바이트 동일** 내용 존재 (존재만 확인하면 빈 디렉터리 구현이 통과)
4. 원본이 호출 후에도 잔존 (복사→`os.Rename` 회귀 차단)

**4번을 먼저 검사하도록 배치**했다 — 원본이 사라진 상태에서 해시부터 돌리면
`hashDir`의 `t.Fatalf`가 먼저 터져 "제거됨"이라는 진단이 묻힌다.

**의도적 비단언**: "호출 후 원본 부재"는 단언하지 않는다. `archiveSkill`도
`archiveLegacySkills`도 원본을 제거하지 않으므로(복사 후 `return nil`, `os.Remove` 없음),
그 단언은 코드에 없는 모델을 기입하게 된다.

### M6 — 바이너리 문자열 교정 (완료, 커밋 `f7303abbd`)

| 위치 | 종전 서술 | 조치 |
|---|---|---|
| `Long` 도움말 | `exit 2` JSON 핸드오프 + 30분 타임아웃 + 스크립트 호출 | `RunE`가 실제 제공하는 `--report`/`--abort` 2개 모드와 기본 모드 무동작만 서술 |
| stderr 안내 2행 | 부재 스크립트 실행 지시 (`sh scripts/ci-watch/run.sh`) | "이 명령에서 watch 루프는 돌지 않는다" + 사용 가능한 두 모드 안내 |
| 주석 1행 | 부재 스크립트를 **호출자**로 지목 | 오케스트레이터를 호출자로 정정 (AC-006 선택자 밖이나 유지 시 유지보수자를 오도) |

**도움말 산문은 구현의 근거가 아니다** — 종전 `Long`이 서술한 계약을 `RunE`는 구현하지
않는다. M6은 산문을 구현에 맞췄고, 산문대로 구현하지 **않았다**.

`internal/cli/pr_watch_cmd_test.go` 를 **작성했다**(계획상 선택). 편집된 문자열 3곳 중
2곳이 `RunE` 본문 안이라 플래그 정의가 편집 사고 반경 안에 있고, AC-CLD-018은 판정
시점의 수동 CLI 배치라 항구적 가드가 없기 때문이다. 단언 범위는 플래그 3개의 존재·기본값과
`Args` 계약(`--abort` 유무에 따른 위치 인자 요구)에 한정했다 — 종료 코드는 상태 파일
존재 여부에 의존하므로 CLI 계층 판정에 남겼고, stderr 문구는 이 편집이 의도적으로 바꾼 값이다.

### AC PASS/FAIL 매트릭스 (M4-M6 담당분)

| AC | 상태 | 판정 명령 | 관측 출력 |
|---|---|---|---|
| AC-CLD-003 | PASS | `go test ./internal/template/ -run TestEmbeddedTemplatesExcludeCILoopSkill -count=1 -v \| grep -q '^--- PASS: …'` | `exit=0` (baseline 부재 시 `exit=1`) |
| AC-CLD-004 | PASS | 생성기 `--all` + `git diff --exit-code -- catalog.yaml` (**M4 커밋 후** 판정) | `gen-exit=0` **그리고** `diff-exit=0` |
| AC-CLD-006 | PASS | `grep -n 'scripts/ci-watch/run.sh' pr_watch_cmd.go \| grep -vc ':[[:space:]]*//'` | `0` (baseline 3). 원시 매치도 `4 → 0` |
| AC-CLD-015 | PASS | `grep -c 'moai-workflow-ci-loop' update_archive.go` / `TestLegacySkillIDsNotEmbedded` | `2` (≥1) / `ok` 계속 PASS |
| AC-CLD-018 | PASS | `make build` 후 6개 관측 | `0 / 0 / 0 / 3 / 1 / 0` — 편집 전 baseline과 **동일** |
| AC-CLD-019 | PASS | `go test ./internal/cli/ -run TestArchiveSkill_PreservesUserContent -count=1 -v \| grep -q '^--- PASS: …'` | `exit=0` |

### 신규 가드 2건의 RED 증거 (변이 검증)

두 가드 모두 M1-M3 완료 상태에서는 즉시 통과하므로, 실패 가능성을 변이로 실증했다.

**AC-CLD-003 가드** — 2회 변이:

```
① 단언 반전 (len(found)==0 이면 실패)
   → FAIL  "INVERTED-RED-PROBE: expected >0 files under
            .claude/skills/moai-workflow-ci-loop/, found 0"
   (순회가 실제로 돌고 개수가 진짜 0임을 확인)

② prefix 를 배포 중인 스킬로 교체 (moai-workflow-tdd) — 탐지력 실증
   → FAIL  "embedded templates still ship 3 file(s) under
            .claude/skills/moai-workflow-tdd/: [SKILL.md references/examples.md
            references/reference.md]"

복원 후 → --- PASS
```

**AC-CLD-019 가드** — 2회 변이:

```
① copyDirAll 호출 무력화 (아카이브 디렉터리는 생기되 비어 있음)
   → FAIL  "archive file count = 0, want 2 (src=map[SKILL.md:4a7c673c…
            references/user-playbook.md:b2eb3521…] dst=map[])"
   (존재-only 검사였다면 통과했을 상태를 바이트 대조가 잡는다)

② 복사→이동 회귀 주입 (copyDirAll 뒤 os.RemoveAll(srcDir))
   → FAIL  원본 디렉터리 부재 탐지

복원 후 → --- PASS
```

### 빌드·회귀 (M4-M6 완료 시점)

```
go build ./...                            → exit 0
GOOS=windows GOARCH=amd64 go build ./...  → exit 0
go vet ./...                              → 출력 없음
golangci-lint run --timeout=5m            → 0 issues.
go test ./... -count=1                    → FAIL 0건, ok 105 패키지
```

M1-M3이 남긴 두 부류 실패는 모두 해소되었다: catalog 해시 신선도는 `make build`가,
하드코딩 기대값은 봉투 편입 후의 정합화가 처리했다.

**커버리지**

```
internal/template   85.9%   (85% 기준 충족)
internal/cli        75.9%   (85% 미달 — 사전 존재 baseline)
```

`internal/cli` 미달은 이 SPEC이 만든 상태가 아니다 — 이번 변경은 시험 2건을 **추가**
했을 뿐이므로 커버리지를 낮출 수 없다. 패키지 전반의 기존 부채이며 별도 소관이다.

### 보존 자산 무변화 재확인 (M4-M6 후)

```
find scripts/ci-watch scripts/ci-autofix -type f | wc -l   → 9
.claude/skills/moai-workflow-ci-loop/                       → SKILL_OK
.claude/rules/moai/workflow/ci-watch-protocol.md            → MIRROR_WATCH_OK
.claude/rules/moai/workflow/ci-autofix-protocol.md          → MIRROR_AUTOFIX_OK
```

### 미검증 (Gaps) — M4-M6 담당분

- **실배포 산출물 미관측**: 실제 `moai init` 결과물은 이번에도 관측하지 않았다.
  임베드 FS 순회(AC-CLD-003)와 catalog 해시(AC-CLD-004)가 대체 판정이다.
- **`internal/cli` 커버리지 85% 미달**: 사전 존재 baseline이며 이 SPEC 범위 밖.
- **AC-CLD-018의 항구성**: 여섯 관측값은 판정 시점의 수동 CLI 배치다. 신설한
  `TestPRWatchCmd_FlagSetUnchanged`가 플래그 표면만 항구적으로 고정하며,
  종료 코드는 여전히 수동 판정에 의존한다.
- **`AskUserQuestion` 문자열 잔존 (오탐 아님)**: `pr_watch_cmd.go` 의 `Long` 문자열에
  `AskUserQuestion` 2건이 남아 있으나 **호출이 아니라 금지 규칙을 서술하는 산문**이며,
  M6 이전부터 존재했고 이번에 보존했다. 서브에이전트 경계 위반이 아니다.

### M7 — 미러 동기화 및 중립성 검증

미러 9개 파일에서 ci-loop 참조 19건을 제거했다. **파일 삭제는 없다** — M7이 제거하는 것은
참조이며, 템플릿 측 삭제는 M3에서 이미 끝났다. 선언된 dev-only 보존 집합 5경로는 무변화다.

**참조 제거 전수 (파일별 건수, 편집 전 실측)**

```
.claude/agents/moai/manager-develop.md                              1
.claude/rules/moai/development/manager-develop-prompt-template.md   3
.claude/rules/moai/workflow/cadence-bridge.md                       1
.claude/skills/moai/SKILL.md                                        3
.claude/skills/moai/workflows/fix.md                                3
.claude/skills/moai/workflows/loop.md                               1
.claude/skills/moai/workflows/run.md                                1
.claude/skills/moai/workflows/sync/delivery.md                      3
.moai/config/sections/delegation.yaml                               3
                                                            합계   19
```

계획서 §B는 미러 편집 대상을 10개로 적었으나 **실측은 9개**다. 차이는 오기가 아니라
M2의 귀결이다 — `.claude/rules/moai/core/zone-registry.md`가 열 번째였고 M2가 이미
해당 항목을 삭제했다. AC-CLD-017의 baseline `10`은 M2 이전 시점의 값이다.

**산문 복구 (참조 삭제만으로 끝나지 않은 편집)**

세 파일은 참조가 문장·목록·섹션의 구조에 물려 있어 주변 산문까지 함께 고쳤다.

- `sync/delivery.md` — 239행은 대체 대상이 없는 `Skill("moai-workflow-ci-loop")` **실호출**
  이었다. 6번 항목을 통째로 제거해 목록을 1~5로 닫았다(번호 공백 없음). 말미의
  단일 항목 `## Related Skills` 섹션도 함께 제거하고 Version 3.8.0 → 4.0.0.
- `fix.md` — 동일한 단일 항목 `## Related Skills` 제거 + Version 2.4.0 → 3.0.0.
  `Previous:` 줄이 두 개로 갈라지는 것을 막기 위해 2.3.0 이력(제거된 skill을 서술)을
  걷어내고 2.2.0을 이어 붙였다.
- `loop.md` / `fix.md` 형제 서술 — "**proactive** CI-triggered watch is the
  `moai-workflow-ci-loop` skill" 절만 제거해 3-quadrant 서술을 2-quadrant로 좁혔다.

`delegation.yaml`은 3개 항목 제거 후 YAML 파싱과 맵 정합을 재확인했다(`yaml.safe_load` 통과,
sync/fix/loop 세 서브커맨드의 `skills:` 배열이 비지 않음).

**미러는 템플릿 복사가 아니다**: 템플릿은 배포 중립화를 거쳤고 미러는 개발 저장소 문서다.
보호 파일 목록의 `scripts/ci-watch/run.sh`는 템플릿과 동일하게 "CI watch infrastructure
and workflow definitions"로 일반화했으나, `ci-watch-protocol.md`·`ci-autofix-protocol.md`
**교차 참조는 미러에 그대로 남겼다** — 이 저장소에는 그 문서가 실재하기 때문이다.

**AC 판정 (M7 담당 3건)**

```
AC-CLD-013  find scripts/ci-watch scripts/ci-autofix -type f | wc -l   → 9
            test -d .claude/skills/moai-workflow-ci-loop               → SKILL_OK
AC-CLD-014  TestTemplateNoInternalContentLeak                          → --- PASS (0.47s)
            TestSplitHarnessNamespaceNoLeak                            → --- PASS (0.00s)
            grep -rn 'SPEC-CI-LOOP-DEVONLY-001' templates/ catalog.yaml → 0
AC-CLD-017  제외 연쇄 선택자                                            → 0  (편집 전 9)
```

**CLAUDE.local.md 등재 (REQ-CLD-016)**

`### Local-Only Files (Never in Templates)` 블록에 보존 집합 5경로를 기존 항목 양식
(경로 + 한 줄 사유)으로 추가했다. 파일 크기 53,316 → 53,901 bytes (+585).
이 파일은 M7 이전부터 40,000자 권고를 초과한 상태이며(§코딩 표준 File Size Limits),
M7은 그 상태를 만들지 않았고 해소하지도 않는다 — 별도 소관이다.

### 전체 AC 재판정 (M7 시점, 19건 전수)

M7은 마지막 마일스톤이므로 선행 인스턴스가 판정한 항목까지 전부 재실행했다.
후행 마일스톤이 선행 마일스톤을 조용히 되돌릴 수 있고, 이 SPEC의 § 판정 규율이
"명령에 없는 판정은 없다"를 요구하기 때문이다. **19/19 PASS, FAIL 0.**

```
001 →0   002 →0/0   003 exit=0   004 gen=0 diff=0   005 →0
006 →0   007 →0     008 FILE_OK/10/0                009 exit=1 ci=0
010 exit=1 total=59  011 →65     012 diff=0 10/10   013 →9/SKILL_OK
014 PASS/PASS/0      015 →2 +PASS 016 →0            017 →0
018 build=0 flag=0 default=0 flags=3 abort=1 report=0
019 exit=0
```

### 빌드·회귀 (M7 완료 시점)

```
go build ./...                            → exit 0
GOOS=windows GOARCH=amd64 go build ./...  → exit 0
go vet ./...                              → 출력 없음, exit 0
golangci-lint run --timeout=5m            → 0 issues, exit 0
go test ./... -count=1                    → ok 104, FAIL 2 (아래 참조)
```

**시험 2건 실패 — 회귀가 아님.** 전수 실행 중 `TestHookWrapper_LargeStdin_DoesNotExceedTimeout`
(internal/cli)과 `TestSupervisor_MultipleWatchers`(internal/lsp/subprocess)가 실패했다.
둘 다 **시간 단언** 시험이고, 격리 재실행에서 모두 통과한다:

```
go test ./internal/cli/ -run TestHookWrapper_LargeStdin_DoesNotExceedTimeout  → --- PASS (0.20s)
go test ./internal/lsp/subprocess/ -run TestSupervisor_MultipleWatchers       → --- PASS (2.63s)
```

원인은 환경이다. 전수 실행 직전 디스크가 가득 차(`errno=28 no space left on device`로
링커가 19개 패키지에서 실패) 빌드 캐시를 비웠고, 그 결과 전수 실행이 **콜드 전체 재빌드와
동시에** 진행됐다. 구조적 근거도 있다: M7 diff에 `.go` 파일이 0건이고 `internal/` 경로가
0건이므로 이 변경이 Go 시험 거동에 영향을 줄 경로 자체가 없다.

### 미검증 (Gaps) — M7 담당분

- **`make build` 미수행**: 위임 제약이 `make build`를 금지했다(신선한 catalog 게이트를
  다시 여는 위험). AC-CLD-018의 첫 줄이 `make build`이므로 대신
  `go build -o /tmp/moai-m7 ./cmd/moai`로 임시 바이너리를 만들어 판정했다. 이 치환은
  catalog 생성기를 돌리지 않으므로 AC-CLD-004를 건드리지 않지만, `make build` 경로
  자체(설치·해시 재생성 포함)는 이번 실행에서 관측되지 않았다.
- **실배포 산출물 미관측**: M4-M6과 동일. `moai init` 결과물은 이번에도 보지 않았다.
- **미러/템플릿 문구 동등성 미판정**: 어떤 AC도 미러 문구가 템플릿 문구와 의미상
  일치하는지 판정하지 않는다. AC-CLD-017은 토큰 부재만 센다. 산문 복구의 타당성은
  리뷰어 diff 확인에 의존한다.
- **`.claude/agent-memory/` 잔존 (설계상 제외)**: gitignored 메모리 파일에 토큰이
  남아 있으나 AC-CLD-017 선택자가 명시적으로 제외한다.

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-08-02
run_commit_sha: pending-backfill-run-M7   # M4 a21d406bf, M5+M6 f7303abbd, M7 (본 커밋) — 미push
run_status: complete                      # M1-M7 전부 완료, run-phase audit-ready
ac_pass_count: 19                         # 19건 전수 재판정 (M7이 마지막 마일스톤)
ac_fail_count: 0
preserve_list_post_run_count: 9           # scripts/ci-watch + ci-autofix 전수, M7 후 무변화
l44_pre_commit_fetch: not-performed       # 격리 워크트리, push 없음 — 오케스트레이터 소관
l44_post_push_fetch: not-performed        # push 미수행 (오케스트레이터가 단일 push)
new_warnings_or_lints_introduced: 0       # golangci-lint 0 issues, go vet 무출력
cross_platform_build:
  darwin_amd64: pass
  windows_amd64: pass
total_run_phase_files: 21                 # M4 7 + M5 2 + M6 2 + M7 10
m1_to_mN_commit_strategy: |
  M1-M3  81ff67937  (선행 인스턴스)
  봉투    b2c5fd1a6  (시험 파일 5개 편입)
  M4      a21d406bf  (AC-CLD-004의 커밋-후 판정 시점을 지키기 위해 단독 커밋)
  M5+M6   f7303abbd  (상호 독립이나 동일 패키지·단일 검증 배치)
  M7      (본 커밋)   (미러 동기화 + dev-only 등재 + 중립성 검증 — 단일 커밋)
```

## §E.4 Sync-phase Audit-Ready Signal

```yaml
sync_complete_at: 2026-08-02
sync_commit_sha: pending-backfill-sync     # 본 커밋 — 자기 해시 참조 불가, 후속 커밋에서 백필
sync_status: complete
b12_self_test_a: pass                      # grep -c 'SPEC-CI-LOOP-DEVONLY-001' CHANGELOG.md → 0 (중복 없음)
b12_self_test_b: pass                      # acceptance.md AC 고유 식별자 19건 == CHANGELOG 기재 19건
b12_self_test_c: pass                      # CHANGELOG가 지목한 모든 경로를 ls/grep으로 실재 확인
changelog_entry_position: "[Unreleased] → ### Fixed (최상단)"
frontmatter_status_transitions:
  spec_md: "in-progress → completed"       # 단일 sync 커밋이 3-phase close 전체를 운반
  plan_md: n/a                             # frontmatter 없음
  acceptance_md: n/a                       # frontmatter 없음
  updated_field: 2026-08-02
canary_compliance_check: n/a               # 본 SPEC은 자기 sync가 시험하는 전향적 정책을 정의하지 않음
docs_site_4locale_sync:
  pages: 3                                 # guides/ci-autonomy, advanced/skill-guide, workflow-commands/moai-sync
  locales: 4                               # ko(정본) en ja zh
  section_count_parity: pass               # 3개 문서 전부 4개 로케일 h2/h3 개수 동일
  hugo_build: "exit 0, warning 0"
mx_tag_validation: pass                    # sync 하위 단계로 수행 — 신규 @MX 주석 대상 Go 변경 없음(문자열 3건 교정만)
```

### 동시성 관찰 — 전체 스위트 flakiness (sync-phase 기록)

run-phase 검증 중 전체 스위트 flakiness가 관측되었다. `go test ./...`를 두 번
독립 실행했을 때 실패 집합이 **서로 겹치지 않았다** — 1회차:
`TestHookWrapper_LargeStdin_DoesNotExceedTimeout`, `TestSupervisor_MultipleWatchers`;
2회차: `TestPreCommitRelocation_GoodCommitPasses`, `TestRecordEvent100Sequential`,
`TestBranchGuard_Latency`, `TestAsyncRecorder_NonBlockingUnderLoad`. 6건 전부 단독
실행에서는 PASS했다. 6건 모두 타이밍·지연·동시성 단언이다.

본 SPEC의 Go 변경은 `internal/cli`와 `internal/template`에 국한되는데, 6건 중
4건은 Go 변경이 전혀 없는 `internal/hook`·`internal/harness`·`internal/telemetry`에
있고, `internal/cli` 건은 pre-commit 훅 재배치를 시험하는 반면 본 SPEC의
`internal/cli` 변경은 주석 1건 + 슬라이스 내 문자열 리터럴 1건 +
`pr_watch_cmd.go`의 문자열 리터럴들이다. 부하 의존적 flakiness로 판정하며
**회귀가 아니다**.

**미증명**: base 커밋(`a0aa9182e~1`)에서 스위트를 돌려 해당 실패가 선재함을
입증하지는 **않았다**. 지금 진단하지 않고 기록만 남기라는 사용자 결정에 따른다.
한 회차 중 디스크 고갈 사고(볼륨 100% 도달, 19개 패키지가 `errno=28`로 링크 실패,
`go clean -cache`로 약 307Gi 회수해 복구)가 있었고, 그 회차의 부하 프로파일에
기여했을 가능성이 있다.

## §F Phase 4 Mode Selection

**입력 파라미터**

| 항목 | 값 |
|---|---|
| tier | M |
| scope (파일 수) | 28 (템플릿 14 + 미러 10 + 저장소 루트 1 + Go 3) |
| domain count | 5 (템플릿 문서 / 규칙·헌법 레지스트리 / 스킬·카탈로그 / Go CLI / 설정 YAML) |
| file language mix | markdown·YAML 우세 (25/28), Go 3 |
| concurrency benefit | LOW — M1→M2·M3→M4→M5 의존 사슬이 직렬이고, M4→M5는 `TestLegacySkillIDsNotEmbedded`가 기계적으로 강제 |

**모드 평가**

| 모드 | 선택 | 사유 |
|---|---|---|
| 1 trivial | 미선택 | 28파일·문서 재작성 포함, 사소 변경 아님 |
| 2 background | 미선택 | 쓰기 작업 (읽기 전용 아님) |
| 3 agent-team | 미선택 | RETIRED (tombstone) |
| 4 parallel | 미선택 | 도메인 5개이나 research-heavy가 아니라 편집-heavy이며, 마일스톤 간 의존이 직렬이라 병렬 이득 없음 |
| 5 sub-agent | **선택** | 마일스톤 단위 순차 위임. Tier M이며 Section A-E 위임 템플릿 적용 |
| 6 workflow | 미선택 | 단일 균일 변환 규칙이 아님 — `ci-autofix-protocol.md` 재작성·레지스트리 절 재작성은 의미 변경이며 파일 간 의존(레지스트리↔소스 동일 커밋)이 존재 |

**Decision: sub-agent**

**정당화**: 편집 대상은 많으나 변환 규칙이 파일마다 다르고(삭제 / 참조 제거 / 본문 재작성 / Go 문자열 교정), M1→M3→M4→M5 의존이 직렬로 강제된다. Anthropic의 coding-task parallelism caveat에 따라 코딩·편집 성격 작업은 순차 서브에이전트가 기본이며, Mode 6는 "단일 균일 기계 변환 + 파일 간 무의존"만 허용하므로 해당하지 않는다.

**Implementation Kickoff Approval**: 통과 (사용자 승인 수신). Phase 0.5 plan-auditor 재실행은 사용자 지시(감사 3회 소진, 0.84 PASS-WITH-FIXES)로 생략 — skip-eligible 0.90 자동 경로가 아니라 **명시적 사용자 오버라이드**로 기록한다.

**착수 시 확정된 사용자 결정 3건**

| # | 결정 | 값 |
|---|---|---|
| 1 | run-phase 진입 | 승인 |
| 2 | M5 `archiveVersion` | (a) 기존 상수 `"v2.16"` 재사용 — 라벨 부정확은 수용, 보존 동작에 무영향 |
| 3 | 저비용 판정 승격 | 둘 다 채택 — RunE 불변성 4단언(plan.md §B 판정 공백) + `archiveSkill` 단위시험(REQ-CLD-018 부채 해소) |

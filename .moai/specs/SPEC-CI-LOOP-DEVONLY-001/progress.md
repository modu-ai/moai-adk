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

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_

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

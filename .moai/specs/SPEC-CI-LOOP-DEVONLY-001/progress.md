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

## §E.2 Run-phase Evidence

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_

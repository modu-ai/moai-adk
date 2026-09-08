# Progress — SPEC-AC-LOCALE-TOKEN-001

## §E.1 Plan-phase Audit-Ready Signal

plan_status: audit-ready
plan_complete_at: 2026-09-08
tier: M
harness: standard
measured_on_tree: 3ac58b5a1

### plan-audit 수리 기록

- 2026-09-08 iter1: FAIL 0.825 (`.moai/reports/plan-audit/SPEC-AC-LOCALE-TOKEN-001-review-1.md`, D1-D9). 수리: D1 funnel 재실행 + B 명령 명기(66, 스냅샷) · D2 코퍼스 spec/plan 확장(2276) + progress.md 제외 경계 명시 · D3 중괄호 철자 필터 병합(A=161) · D4 A∖B 전수 사람 판독 단계 신설(95, AC-009) · D5 C-3 문면 개문(코드 위치 한정 조건) · D6 자기-디렉터리 제외 + 스냅샷/재측정 구분 · D7 grep 플래그 의미 명기(en 4=`-o`, 8=`-oi`) · D8 lint 워크트리 직접 실행 흡수(exit 0, REQ 레이어 맹점 주석) · D9 REQ 열 추가 + AC-009/AC-010 을 REQ-007 검증으로 명시. 재측정 근거: plan-research.md §6.

## §E.2 Run-phase Evidence

> 측정 트리: 워크트리 `WT-ascii-token-criterion` 작업 트리 (base `3ac58b5a1`) · M1 커밋 `7f0a88510` 시점 HEAD · 측정일 2026-09-08 · 모든 grep 은 `/usr/bin/grep`.

### AC PASS/FAIL 매트릭스

| AC | 상태 | 검증 명령 | 실측 출력 |
|---|---|---|---|
| AC-001 census 장부 | **PASS** | `test -f .moai/reports/t573/census.md` | 존재 — REQ-002 7필드 스키마(census §1) + 66 C-판정(§3) + 95 A∖B 판독(§4) + 잔여 4건(§5) |
| AC-002 swept-set 계수 | **PASS** | census §2 funnel 재실행 | 코퍼스 `2276` · A1 `55` · A2 `98` · A3 `62` · A `161` · A∩B `66` · A∖B `95` — plan 스냅샷과 동일 재현, 각 단계 명령 명기. A∩B 66파일에서 계수식 397행 전수 추출해 C-판정 (빈 집합 통과 없음) |
| AC-003 왜곡형 전량 재작성 | **PASS** | census §3 배치 3 대비 M2 커밋 diff | 확정 5건 = 재작성 5건 (t538 AC-004 zh축 / WORKFLOW-DOCS AC-WFD-001·007·011 / RC2-README AC-KO-002b 3번째 불릿). 각 재작성이 REQ-001 4요소 (계수 명령·로케일·실측 기준선·왜곡 불가 근거) 운반 |
| AC-004 ASCII 계수 잔존 제거 | **FAIL (프록시 결함)** | `/usr/bin/grep -c 'grep -ci "desktop-native"' .moai/specs/SPEC-DOCS-LOCALE-PARITY-REPAIR-001/acceptance.md` | `5` (기대 `0`). 분해: zh 대상 라이브 기준 사용 **0** (§D.4 zh 축이 `原生桌面` ≥6행 + 매트릭스 3행 구조로 대체 — 본질 요구 충족, 실측 `/usr/bin/grep -c "原生桌面" …` → `6`, `/usr/bin/grep -c '^| \*\*原生桌面' …` → `3`). 잔존 5행 = en 축 1 (REQ-004가 유지를 지시 — "en 축은 유지(en 은 Latin 로케일 — 왜곡 없음)") + HISTORY/논거 인용 4 (AC-005가 before 명령 인용을 의무화). **acceptance §D.4 의 나브 문자열 프록시(1→0)는 자기 SPEC 의 REQ-004·AC-005 와 상충** — manager-spec 의 프록시 재스코핑 소관 (본 에이전트는 자기 SPEC acceptance 본문 수정 불가 — blocker 성격 소계 항목, 리드 판정 대기) |
| AC-005 HISTORY 규율 | **PASS** | 3개 acceptance.md 의 HISTORY 절 | t538·WORKFLOW-DOCS·RC2-README 각 2026-09-08 항목: before/after 명령 + 실측값(트리 `0e1f248cd`) + 판정 영향 명시 + 승인 귀속 "리드 승인 — 카드 t573 발행 dispatch" (WORKFLOW-DOCS 는 spec.md HISTORY 표에 §C.4 정정 행 추가) |
| AC-006 뮤턴트 probe | **PASS** | census §6 기록 명령들 | 보강-제거 사본: `/usr/bin/grep -o 'desktop-native' /tmp/t573-zh-reverted.md \| wc -l` → `2` (옛 기준 미달 — 왜곡 강제 실증), `/usr/bin/grep -c "原生桌面" /tmp/t573-zh-reverted.md` → `6` · 매트릭스 → `3` (재작성 기준 통과). ja 글로스-제거 사본: ASCII `Class A` → `0` (옛 AC-WFD-001 미달), `クラス ?A` → `1` (재작성 통과) |
| AC-007 M3 옵션 + docs-site 무변경 | **PASS** | `git status --porcelain -- docs-site` | 빈 출력, exit 0. M3 옵션 기준값: A(유지) `原生桌面` 6행·매트릭스 3행 통과 / B(전면 되돌림) 사본 실측 동일 6행·3행 통과 — 재작성 기준은 양쪽 모두 통과. 운영자 게이트에서 **A(KEEP) 결정됨** (카드 t573 킥오프, 2026-09-08) |
| AC-008 2차 관찰 경계 | **PASS** | census §7 | 4건 기록 (GDR auto-mode ko=0 잠재, DOCSITE :115 실행 불가 스케치, RC2-DOCSITE 이중-글로스 스타일, 계열 근접 미스) — 수정 커밋 없음 |
| AC-009 A∖B 전수 판독 | **PASS** | census §4 표 행 계수 + O-1 교차 확인 | 표 행 `95` · `sort -u` 유니크 파일 토큰 `95` — 일치. 판정: benign 60 · 비-grep 계수 관찰 35 · 왜곡형 0 |
| AC-010 이름 붙은 잔여 철자 | **PASS** | census §5 | 4건 전부 판정 행 (benign 3 — 부정 계수/코드 토큰/구조, n-a 1 — `.moai/docs/version-management.md` 는 영어 문서, 한글 행 0 실측) |

### lint

- `moai spec lint .moai/specs/SPEC-AC-LOCALE-TOKEN-001/spec.md` → `✓ No findings — all SPEC documents are valid`, exit 0
- `moai spec lint .moai/specs/SPEC-DOCS-LOCALE-PARITY-REPAIR-001/spec.md` → `✓ No findings`, exit 0
- `moai spec lint .moai/specs/SPEC-V3R6-WORKFLOW-DOCS-001/spec.md` → exit 0, 0 error / 12 warning (`CoverageIncomplete` — 기존 baseline, 본 카드 변경 전부터 존재하는 REQ↔AC 수집 패턴 사항)
- `moai spec lint .moai/specs/SPEC-V3R6-DOCS-RC2-README-001/spec.md` → exit 0, 0 error / 1 warning (`MovingRefUnpinned` plan.md:234 — **미편집 파일**, 기존 baseline)
- 신규 warning/lint 도입: **0**

### 재작성 내역 (5건)

| # | 파일 | AC | 옛 식 → 새 식 |
|---|---|---|---|
| 1 | SPEC-DOCS-LOCALE-PARITY-REPAIR-001/acceptance.md §D.4 | AC-004 | zh `/usr/bin/grep -ci "desktop-native" <zh>` ≥4 → `/usr/bin/grep -c '^| \*\*原生桌面' <zh>` `3` + `/usr/bin/grep -c "原生桌面" <zh>` ≥6 (en 축 유지 — Latin 로케일) |
| 2 | SPEC-V3R6-WORKFLOW-DOCS-001/acceptance.md AC-WFD-001 | :12 | `grep -c "Class A/B/C" ≥1 ×4` → 로케일별 병렬 `클래스 ?A`/`Class ?A`/`クラス ?A`/`类别 ?A` 각 ≥1 (실측 각 1) |
| 3 | 동파일 AC-WFD-007 | :54 | `grep -c "Implementation Kickoff" ≥1 ×4` → en ASCII + ko/ja/zh 원어-병렬 (`구현 착수 승인\|…` 등, 실측 각 3) — 임계값 grep(0.75/0.80/0.85)은 로케일 불변 리터럴로 유지 |
| 4 | 동파일 AC-WFD-011 | :82 | 4 차원명 ASCII ≥1 ×4 → en ASCII + ko/ja/zh 12조합 원어-병렬 (실측 각 1); `sync-auditor` (에이전트 id) 유지 |
| 5 | SPEC-V3R6-DOCS-RC2-README-001/acceptance.md AC-KO-002b | :207 | `grep -cE "8 (retained )?agents" README.ko.md` ≥1 → 병렬 패턴 양-README **패리티** (실측 0/0 — 대상 콘텐츠 양 로케일 퇴역, HISTORY 에 실패 귀속 명시) |

### 범위 내 파일 (run-phase 총 7)

`.moai/reports/t573/census.md`(신설) · 자기 SPEC spec.md(status 전환)·progress.md(본 절) · t538 acceptance.md · WORKFLOW-DOCS acceptance.md·spec.md(HISTORY+§C.4) · RC2-README acceptance.md. docs-site 변경 0, Go 코드 변경 0.

## §E.3 Run-phase Audit-Ready Signal

run_complete_at: 2026-09-08
run_commit_sha: pending-backfill-run
run_status: evidence-complete
ac_pass_count: 9
ac_fail_count: 1
ac_fail_note: AC-004 나브 프록시(문자열 계수 1→0) — 본질 요구(zh 대상 왜곡 기준 제거)는 충족·실측됨. 프록시 재스코핑은 자기 SPEC acceptance 본문 수정이므로 manager-spec 소관 — 리드 판정 필요
preserve_list_post_run_count: 0
new_warnings_or_lints_introduced: 0
cross_platform_build: n-a (문서 전용 SPEC — Go 변경 없음)
total_run_phase_files: 7
m1_to_mN_commit_strategy: M1 census 1커밋 + M2 재작성 1커밋 + M-final progress 1커밋 (레인 커밋, push 없음 — develop 병합은 리드 창 소관)

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_

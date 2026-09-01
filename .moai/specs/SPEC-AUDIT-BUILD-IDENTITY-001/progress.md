# SPEC-AUDIT-BUILD-IDENTITY-001 — 진행 기록

카드 t248 · 워크트리 `WT-audit-binary-sha` (base `64bba61aa`)

## §E.1 Plan-phase Audit-Ready Signal

- 산출물: `spec.md`(요구 8, v0.1.2), `acceptance.md`(수락 8), `plan.md`, 이 파일
- status: `draft`
- 미해결: 없음. plan.md §B의 열린 질문 2건은 운영자 결정으로 D1(평탄 형제 필드 `build_commit`/`build_lag`, 중첩·버전 필드 없음)·D2(지연 권고도 형제 필드)로 닫혔다 — spec.md v0.1.1
- 구현 0줄. 커밋 없음

### plan-audit 결함 수리 라운드 (v0.1.2, 2026-09-01)

plan-audit 판정 FAIL(점수 0.85, 반복 1/1, `.moai/reports/t248/plan-audit.md`) → 차기 세션이 결함별로 디스크 상태를 먼저 검증하고 미착지분만 적용했다. 선행 세션이 적용한 분은 재적용하지 않았다.

| 결함 | 처분 |
|---|---|
| D-1 (소스 스윕 기대가 거짓) | 선행 세션이 착지. 본 세션이 같은 grep을 본 트리(`64bba61aa`)에서 재실행해 히트 3건(`graph_stamp.go:68`/`:131`, `mcp_review_material.go:95`)으로 기준선 재확인 — exit 0. 기대 절은 기준선 3좌표 **정확 집합** 술어 + `resolveReviewMergeBase`는 diff 기준점 해석임을 명시하는 문장을 갖는다 |
| D-2 + D-5 (공허화 + Tier S 상한 9>8) | 선행 세션이 착지. 모양 기준을 AC-ABI-001에 병합(`build_commit` 비어있지 않음 선행 전제) — 수락 9→8, 생존 AC 재번호 001..008, §2.1/§3/plan.md 상호 참조 정합 확인 |
| D-3 (빈 `projectRoot`에서 REQ-ABI-006 도달 불가) | 선행 세션이 착지. option (a) — `os.Getwd()` 폴백(`doctor.go:521` 선례). REQ-ABI-006 본문 + plan.md §B D3 + AC-ABI-006 빈-root 경로 절(`StatusBehind` 스텁에서 `build_lag` 비어있지 않음) 모두 존재 |
| D-4 (지연 AC가 진입점 범위 미구속) | 선행 세션이 착지. AC-ABI-006이 세 핸들러 전수 + 테이블 구동 [HARD] + `StatusFresh` 대조군 |
| D-6 (§1.4가 [HARD] 픽스처 규율의 실제 사냥감 미명명) | 선행 세션이 착지. §1.4 말미 문단이 M2(필드 이름은 `build_commit`, 값은 버전)를 명명 |
| D-7 (좌표 3건 드리프트) | 선행 세션이 착지. 본 세션이 각 정정 좌표를 본 트리에서 개 줄 검증 — `mcp_codex.go:1493`, `mcp_glm.go:245`(둘 다 `resolveToolProjectRoot(req)`), `pkg/version/version.go:32`(인용)/`:37`(`func GetBuildID()`) 전부 일치 |
| D-8 (`commit=="none"`에서 `build_commit` 값 미특정) | **본 세션이 보강.** AC-ABI-005 절 4를 `version.Commit` ∈ {`""`, `"none"`, `"unknown"`} 전 집합으로 확대(`binlag.go:108-110` 선례) |
| (부수) plan.md 낡은 참조 | **본 세션이 수리.** §B D5의 기준 수(9→8)와 반뮤턴트 기준 번호(004→003), §F의 회귀 가드 기준 번호(009→008)를 병합 후 번호로 정렬 — 병합 재번호 누락분 |
| (부수) spec.md frontmatter/HISTORY | **본 세션이 착지.** version 0.1.1→**0.1.2**, HISTORY v0.1.2 행 추가, updated 2026-09-01 유지 |

불변 확인(본 세션 관측): 평탄 형제 필드 유지(중첩 0, 버전 필드 0), `internal/binlag.Evaluate` 재사용 요구 유지, D-1 기준선 3좌표 유지. 미해결 마커(`[NEEDS` 로 시작하는 주석 토큰)와 낡은 AC 번호 잔존 참조 모두 grep 매치 0 — 증거는 아래 검증 배치에.

### plan-audit 판정 경로 — 리드 판독 갈음 (2026-09-01)

plan-audit FAIL(반복 1/1, Tier S 상한 소진) → 수리 v0.1.2 → 리드가 **재감사 없이 판독으로 갈음** 판정(수리 완료 인정). 반복 상한 1은 소진 상태 그대로 두고 판정 경로만 갈음했다.

리드 실측 근거(트리 `311b1936b`):

- 구조: AC 집합 = `AC-ABI-001..008` 정확히 8개(`009`·"9개 기준" 참조 grep 0건, exit 1) — D-5 해소
- D-2: AC-ABI-001 `[HARD]` 비어있지 않음 선행 전제 존재(`acceptance.md:31`) + 공허화 메커니즘 설명(`:39`)
- D-3: `os.Getwd()` 폴백 REQ-ABI-006 명시(`doctor.go:521` 선례) + 비전달 계약이 §D.1 체크리스트 항목으로 기계화(`performCodexAudit`/`performGLMAudit` 인자 `git diff`)
- D-1: AC-ABI-008(구 009 회귀 건이 아니라 소스 스윕 건은 AC-ABI-007)이 거짓 기대에서 기준선 3좌표 정확 집합 술어로 교체
- D-6: M2 뮤턴트 명명 확인(`acceptance.md:72`)

lane이 수정 에이전트의 "이미 착지" 처분표 라벨을 불신했던 단서와 리드의 독자 실측이 서로 합치한다 — 판정 근거는 라벨이 아니라 실측이다. lane 독립 스팟체크 9항목(D-1 기준선 본 트리 재측정 포함)도 전부 합치.

산출물 커밋: `311b1936b` (SPEC 4종 + 감사 보고서, 5파일 682줄). run 진입은 Implementation Kickoff 승인 후.

## §F Phase 4 Mode Selection

- 입력: tier S / scope 4-6파일(`mcp_codex.go`·`mcp_convergence.go`·신규 테스트) / 도메인 1(Go backend) / 병렬 이득 낮음(코딩 중심) / Agent Teams 미요청
- 평가: direct 미해당(구현 존재) · fanout 미해당(단일 도메인 코딩 작업) · sweep 미해당(기계적 대량 변형 아님) · serial 해당(단일 스콥 구현)
- Decision: serial
- 근거: 코딩 중심 단일 도메인 구현 — Anthropic coding-task parallelism caveat상 serial이 기본이며, Tier S 단일 스콥이라 마일스톤 병렬화 이득이 없다.
- 실행 모델: manager-develop GLM 상속(opus 주간 한도 2026-09-04 19:00 리셋 대기 — 프로필 이격, 완료 보고에 기록)
- Kickoff: 운영자 승인(자율 모드), 리드 경유 전달. 판정 경로는 리드 판독 갈음(§E.1 기록). run 진입은 이 기록 이후 첫 spawn.

## §E.2 Run-phase Evidence

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_

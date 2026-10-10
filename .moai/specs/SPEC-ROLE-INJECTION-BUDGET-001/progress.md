# SPEC-ROLE-INJECTION-BUDGET-001 — Progress

> 카드 t1617 · 런 tmnboq · 기준 트리 `2aab5f797` (WT-3-2-1)

## §E.1 Plan-phase Audit-Ready Signal

plan_status: audit-ready
plan_complete_at: 2026-10-10
plan_artifacts: spec.md 0.2.0 · plan.md · acceptance.md (Tier M — 3 artifacts)
plan_author: manager-spec

계획 완결 성명(§E.1 — 리더 처분 2b7b: 감사 실행 전 기입 가능, 감사 준비 완료를 진술):

- 측정: 카드 dispatch 의 전 수치를 본 트리에서 재측정해 확정했다(core 18,114/17,793 · 헤더 123/125 · 포인터 128 · 스텁 8,583/8,072 · 원장 36행 · 갈림 1문장). 원문 `.moai/reports/t1617/measurements.md`. 생산자: 직접 분해 4,376(lane-15 startup 참조점, 라이브 자산 재측정 불가) + 게이트 함의 상한 4,797(합본 실측 23,166 − 본 트리 역할 블록 18,369 — 검산 차 421, spec.md §B).
- 해석 결정: 「역할별 조립 ≤9,000자」를 조립 합본 판독으로 확정(spec.md §A.4). 축자 판독은 자기 모순(9,000 + 4,376 > 10,000 한도)이라 기각. 파생 core 예산 3,946(생산자 상한 4,797에서 — leader 구속), 설계 목표 3,800, REQ-RIB-004 정지 밸브가 구속 후퇴선.
- 설계 결정: 리더/레인 역할 분할 기각(완전 분할도 ~9,057 > 3,946 — 산술 기각, spec.md §D), 공유 core 압축 재작성 + 재배치 감사. 구속 절은 core/스텁에만(상위 원장 vocabulary 보존), `buildRoleCore` 표지 검증 무편집.
- 채택: AC 12개 두 칸 채택(acceptance.md) — release-blocking 7개(AC-RIB-001·002·004·005·006·007·009, 원장 항목 E1–E8 8건 — E1·E2가 AC-001의 양 트리 관측), regression-guard 5개(AC-RIB-003·008·010·011·012, §2 undecidable/보존 처분), 변이 탐침 요지 포함.
- Tier: M (3 artifacts) — 다중 파일이나 단일 서브시스템(주입 경로) + 규칙 2쌍 + 템플릿 미러; 헌법급 아님, 상위 원장 기계 위에 Ride.

### 수리 라운드 기록 (0.1.0 → 0.2.0, 2026-10-10)

plan-audit 1회차 FAIL(F1–F7, `.moai/reports/t1617/plan-audit.md`, codex P2 3건 동의)을 1회 수리 라운드로 전부 반영했다. §E.1 감사 준비 신호는 유지된다(수리 뒤 재진술).

- **F1 (산술)**: 합본 교차 검산의 가산 오류(23,066 → 정정 22,745)와 역할 블록 이중 계산(로컬 꼬리 +321은 18,114 core 안에 이미 포함 — 18,690 철회)을 정정했다. 조립 산식을 코드로 고정했다(role_rules.go:102–123 `assembleInjectionComposite` · :373–380 `roleRuleInjectionFor` — 헤더 뒤 `\n\n`에 core가 직접 붙어 결합자 없음, measurements.md §4). 파생: 게이트 함의 생산자 상한 **4,797**, core 예산 **3,946**(구 역할 4,367 철회), 설계 목표 **3,800**, REQ-RIB-004 정지 밸브를 구속 후퇴선으로 명명. RED E1/E2를 임계 3,946으로 재관측 — 값(17,793/18,114)·exit(1) 동일.
- **F2 (표 정렬)**: plan.md 마일스톤↔AC 표가 acceptance.md 번호 대비 한 칸 어긋난 9행을 재정렬하고 누락된 AC-RIB-011·012 행을 추가했다.
- **F3 (분류 일치)**: AC-RIB-009를 release-blocking으로 확정(실행 가능 RED E8 신설 — 라벨 부재 관측), AC-RIB-011을 regression-guard로 강등(§2 미결정 처분 — 편집-후 재관측이 채택 증거). 표·§B·이 신호의 계수 일치: release-blocking 7 / regression-guard 5.
- **F4 (Codex 문장)**: 「Codex는 영향 밖」 문장을 정정했다 — Codex 전용 설치는 `.moai/policies/` 투영을 같은 주입 경로가 읽는다(role_rules.go:129–152) — 예산·게이트·오버플로 동일 적용(spec.md §D).
- **F5 (소스 변이)**: startup이 최중 소스라는 코드 인용 근거(`factoryBootstrapNoticeForSource` startup 전용, `factoryLaneRuleForSource` startup+clear, compact 부분집합)와 게이트 가시 후퇴선, 라이브 확인의 clear 재진입 추가를 REQ-RIB-002에 명시했다.
- **F6 (여유 수치)**: 스텁 여유 19% → 14.2%/19.3% 정정(REQ-RIB-005).
- **F7 (단위·단계)**: REQ-RIB-008을 3분해로, REQ-RIB-011을 코드 단위+바이트 이중 기록과 스텁 델타 레코드 라벨(REQ-RIB-011이 리터럴로 고정 — 이 파일에는 M3가 그 라벨 행을 쓰기 전까지 등장하지 않는다)로 정정; hooks-system.md 템플릿 미러 편집(미러 원본 선행)을 plan.md M3에 명시 단계로 추가했다.

Gap(선언): plan-audit 재실행 전까지 이 신호는 수리 완결 상태의 진술이다 — 재감사는 리더 처분 2b7b 절차대로 lane이 audit_multi로 실행한다.

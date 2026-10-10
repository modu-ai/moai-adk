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
- 채택: AC 13개 두 칸 채택(acceptance.md) — release-blocking 7개(AC-RIB-001·002·004·005·006·007·009, 원장 항목 E1–E8 8건 — E1·E2가 AC-001의 양 트리 관측), regression-guard 6개(AC-RIB-003·008·010·011·012·013, §2 undecidable/보존 처분), 변이 탐침 요지 포함. AC-RIB-001의 단정 범위는 결정 d-20261010T072910Z-f1c3의 경계 소스 봉투(startup · clear-핸드오프 대기 없음 · compact-goal 없음)로 재범위.
- Tier: M (3 artifacts) — 다중 파일이나 단일 서브시스템(주입 경로) + 규칙 2쌍 + 템플릿 미러; 헌법급 아님, 상위 원장 기계 위에 Ride.

### 수리 라운드 기록 — 판정 d-20261010T075659Z-7d7c 반영 (0.3.0 → 0.4.0, 2026-10-10)

라운드-4 수리를 판정 d-20261010T075659Z-7d7c(ceiling-exception, card:t1617) 승인으로 커밋한다.

- **R4 (계보 배너)**: §A.5 에 4행 신설 — 계보 배너(`chainLineageBanner`, session_start.go:611 전 소스 무조건 · chain_banner.go:152-156 무상한 렌더)를 핸드오프 본문·goal 재주입과 같은 처분으로 귀속(폴백 유지·게이트 systemMessage 가시 발화). **승인된 기준 3줄은 바이트 단위 무변경** — 4행 가감만(커밋 전 grep 재확인). 미러: REQ-RIB-002(소스 조건부 재주입 vs 무조건 배너의 극성 차 명시)·REQ-RIB-012/AC-RIB-013(단언 면에 배너 편입)·measurements §4.2(셋째 무경계 행 + codex 프로브: 11,000자 → 합본 14,907 발화 / 고정 테스트 8,854 통과 — 코디네이터 전달 수치).
- **R5 (경로 → 카드 귀속)**: §A.5·measurements §4.2 의 계획서 경로 인용을 카드 귀속으로 전환. 큐 전수 확인(moai todo list, 34카드, 2026-10-10, 읽기 전용) 결과 핸드오프 재개 카드 미발행 → 「승인 대기 핸드오프 재개 카드(카드 발행 전; 계획서는 원본 체크아웃 소유, 본 트리 미추적)」. 큐 수정 없음.

### 수리 라운드 기록 — 결정 d-20261010T072910Z-f1c3 반영 (0.2.0 → 0.3.0, 2026-10-10)

리더 범위 결정 d-20261010T072910Z-f1c3(위임 판단, option (a); (b) 미채택)를 1회 수리로 전부 반영했다. §E.1 감사 준비 신호는 유지된다.

- **완료 기준 재범위**: 무경계 "overflow 파일 0"을 철회하고 spec.md §A.5 의 세 줄(경계 소스 봉투)로 교체 — 리더·레인 역할 코어 조립 각 ≤9,000자(테스트) / startup·clear(핸드오프 대기 없음)·compact(goal 없음)에서 넘침 파일 0 / 핸드오프 본문·goal 재주입 세션은 기존 넘침 파일 폴백 유지, 게이트가 systemMessage로 보이게 발화. AC-RIB-012를 봉투로 재범위(startup + clear-핸드오프 대기 없음), REQ-RIB-012·AC-RIB-013 신설(봉투 밖 게이트 가시 발화 + 폴백 보존 단언).
- **R1 재편**: F5의 생산자 변이 처리를 생산자 셈이 아니라 봉투 산술로 재편(measurements.md §4.2) — 두 무경계 재주입 생산자(claimAndInject의 clear∧auto∧live-pending 셀, compact의 armed-goal 재주입)를 명시하고 그것이 제외의 사유임을 코드로 고정. 0.2.0의 「compact가 가장 작은 부분집합」 주장을 철회한다(compact는 armed goal이 있으면 재주입 본문을 실는다 — 부분집합이 아니다).
- **핸드오프 잔여 소관**: 핸드오프 본문 주입의 크기 상한은 핸드오프 재개 카드(승인 대기)가 담당함을 §A.5 에 명기. 본 리포 handoff.mode=manual(리더 확인 — handoff.yaml:9)이라 자동 클레임 셀은 현 상태에서 발화하지 않는다.
- **R2**: plan.md §A.1 의 잔존 ≤4,000을 예산 3,946/목표 3,800 사슬로 정정.

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

## §G Override and Refusal Record

- 2026-10-10T08:19:02Z SPEC-ROLE-INJECTION-BUDGET-001 required-backend-refusal outcome=refused reasons="verdict carries no must_pass_failed field"
- 2026-10-10T08:22:50Z SPEC-ROLE-INJECTION-BUDGET-001 required-backend-refusal outcome=refused reasons="required backend codex configured and the verdict carries no convergence receipt"

## §E.2 Run-phase Evidence

### M1 — RED: 테스트·고정물 (규칙·코드 무편집)

- 근거 명령(env는 같은 컴파운드 호출 안에서 스크럽; 트리 SHA는 각 커밋 시점에 재귀속):

```text
[RED-1] go test -run 'TestRoleInjectionAssemblyBudget' ./internal/hook/   → exit 1 (EXPECTED RED)
  template/factory-lane   breakdown: producers=4797 joiner=4 header=123 core=17793 pointer=128 total=22845
  template/factory-leader breakdown: producers=4797 joiner=4 header=125 core=17793 pointer=128 total=22847
  deployed/factory-lane   breakdown: producers=4797 joiner=4 header=123 core=18114 pointer=128 total=23166
  deployed/factory-leader breakdown: producers=4797 joiner=4 header=125 core=18114 pointer=128 total=23168
  (core 17,793/18,114 > 3,946 한도 — 이 SPEC 이 고치는 상태 그 자체. deployed/factory-lane 합계 23,166은
    SPEC §A.1 이 인용한 게이트 실측과 정확히 일치 — 4,797 생산자 상한 도출의 독립 교차 검증)
[RED-2] go test -run 'TestRoleRulesVersionSkew' ./internal/hook/          → exit 1 (EXPECTED RED)
  4개 로캘 행 전부 0히트: ko/ja/zh/en 의 InjectionFailed 에 "moai update" 부재;
  TestRoleRulesVersionSkewFailureDetailNamesRemedy 도 동일하게 실패(detail 에 update 지시 없음).
[RED-3] go test -run 'TestRoleRulesVersionSkewPredicate' (role_rules_skew_predicate_test.go 포함) → exit 1
  internal/hook/role_rules_skew_predicate_test.go:16/25/37: undefined: detectRoleRuleVersionSkew
  → [build failed]. plan M1 3번이 허용한 「컴파일 실패」 RED 형태(부호 부재 = 행동 부재).
  이 테스트 파일은 기계 로컬 보관(M1 커밋 미포함 — 패키지가 마일스톤 사이 컴파일 가능 상태를 유지),
  M3 에서 구현과 같은 커밋으로 착지.
[MOTOR] go test -run 'TestRoleCoreStubGate' ./internal/template/          → ok (AC-RIB-003 채택 증거)
  deployed/template 양 트리 glob 각 2개 *-core.md 스윕; 표지 0; 크기 모두 ≤10,000 UTF-16;
  motor 서브테스트 관측: 표지 심은 픽스처 FAIL / 10,001 단위 픽스처 FAIL / 10,000 경계 PASS — 두 게이트 조건 모두 생존.
[회귀 기준선] go test -run 'TestSessionStartRoleRules' ./internal/hook/      → 0 FAIL (기존 21 서브테스트 전부 GREEN)
```

- 산출물: `internal/hook/role_injection_budget_test.go` (AC-RIB-001, E1+E2 관측점),
  `internal/template/role_core_stub_gate_test.go` (AC-RIB-003 게이트 + motor),
  `internal/hook/role_rules_version_skew_test.go` (AC-RIB-004/005 안내면 RED).
- Gap(선언): RED-3 의 판정 산출물은 단언 실패가 아니라 컴파일 실패다 — plan 허용 형태로 기록하며,
  M3 착지 뒤 단언 GREEN 으로 수렴 확인한다.

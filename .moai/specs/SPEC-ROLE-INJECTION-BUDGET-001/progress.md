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

### M2 — GREEN(규칙 축): 역할 core 재작성 + 재배치 감사 (2026-10-10)

- **재배치 감사 원장**: `relocation-ledger.md` 신설(36행 전수 감사 + 특수 행 3건) — AC-RIB-002.
- **역할 core**: 31영역 3,786 UTF-16(설계 목표 3,800 · 예산 3,946 이내) — 17,793/18,114 → −78%. 배포·템플릿 양 트리 바이트 동일(로컬 +321 꼬리는 재배치 감사 행으로 정리 — 본문은 worktree-integration § Hoist 와 mechanics § landing 6단계가 소유, 확인 후 삭제).
- **스텁 귀속(5/36, 비대량)**: ALB-0272·0276·0284·0294·0303 은 2~4 의무 복합절로 REQ-RIB-003/§D 가 허용하는 상시 스텁 처지 선택 — 압축 시 의무 뉘앙스 상실(REQ-RIB-004 위반 우려)을 피함. 목적지·이유는 relocation-ledger §A.
- **도크트린 문장 보존**: `### Lane waits are explicit…` 본문(CronCreate·recurring·disk evidence — lane_recheck_doctrine 경계)과 리더 조건부 문장(tree_scope: skip — codex_review_ownership_m4 경계)은 **비표지 전체 본문**에 원형 생존(스텁·core 예산 무비용). 압축 절과 병존.
- **기존 사다리 테스트 수리(AC-RIB-010 유지)**: 핵심 축소로 overflow 픽스처가 한도 미만으로 떨어진 4개 사다리 서브테스트(b/c/d·Handle·OperatorLocales)에 크기 무관 패딩 헬퍼(overCapFiller/forceOverCapFixture) 도입 — 사다리 행위 단정은 무변경.

stub-delta: `workflow/factory-dispatch-core.md` (배포·템플릿 쌍, 동일) — UTF-16 **8,583 → 9,915**(+1,332), UTF-8 **8,652 → 9,998**(+1,346 bytes). 10,000 UTF-16 예산 이내(잔여 85). 바이트 델타 1,346 > 1,000 → rule-authoring 비산 문: **(a)** 신규 파일 아님 — 기존 상시 파일 성장. **(b)** 성장 실측치 상기와 같음. **(c)** 비호출 세션 비용: 모든 세션이 매 턴(과 /clear 마다) 5개 복합절 1줄 요약을 추가 부담 — 팩토리 비세션에게는 무효 토큰 1,332 단위. 정당화: REQ-RIB-003/§D 가 구속 절의 스텁 처지를 명시 허용하고, 3,946 core 예산 안의 압축으로는 5개 복합절의 의무 뉘앙스(전송 결과 판독·크론 상한·병합 창 접점)가 REQ-RIB-004 금지의 의무 포기로 떨어진다 — 스텁 전문 보전이 유일한 의무 보존 경로였다. `cross-session-messaging-core.md` 무변동(UTF-16 8,072 · 8,110 bytes).

```text
[GREEN-1] go test -run 'TestRoleInjectionAssemblyBudget' ./internal/hook/ → ok (AC-RIB-001 플립; core 3,786 ≤ 3,946)
[RED→GREEN] go test -run 'TestBindingLedger' ./internal/template/  → 편집 뒤·원장 갱신 전 FAIL(after-text absent 36행, AC-RIB-011 채택 증거) → 36행 after-text 재작성(31 core + 5 스텁 location 전환) 후 ok
[GREEN-2] go test ./internal/hook/ ./internal/template/ → 전체 스위트(기존 사다리 21 서브테스트 + 원장 + 스텁 게이트 + 도크트린 경계 포함)
```

### M3 — GREEN(코드·문서 축) + 최종 검증 (2026-10-10)

- **role_rules.go**:
  - 사다리 doc 주석 1회화 — 글머리 목록만 생존, 반복 문단 삭제(E6 awk n=2→**1**, exit 0).
  - NOTE 지시문: 천단위 `10,000` 렌더 + 중복 서술 정리(cap·파일 저장·읽기 경로를 한 번만 이름 대고, 경고의 intact-emission 내용은 경고에 남김). `10000 characters` 무천단위 렌더 0히트.
  - 버전 불일치 감지(REQ-RIB-006): `detectRoleRuleVersionSkew`(순수 술어 — `template_version:` 스탬프 파싱·trim, equal/다름/판독불가) + `roleRuleSkewDetail(root)`(배포 system.yaml 판독, 실패 시 조용히 생략 — fail-open). 미스마커 실패 경로에서 detail 이 양 버전을 이름 대고 `moai update` 를 지시. 4개 로캘 InjectionFailed 에 갱신 안내 문장 추가. REQ-RIB-007 준수 — 술어는 픽스처 단위 시험만(role_rules_skew_predicate_test.go, 3케이스), 라이브-리포 자기 단정 없음(부재를 grep 으로 확인).
- **hooks-system.md**: 50K 총 stdout→디스크 저장과 10,000자 per-additionalContext 전달 한도를 두 행으로 구분 기재, 각 실측 출처(Q4 · 레인 세션 시작 넘침 공지 23,166자 실측) 인용 — 템플릿 미러 원본 선행 편집 후 배포본 미러, 쌍 바이트 동일(E7 grep ≥1). 템플릿 중립 경계 준수를 위해 관측 인용에서 SPEC id·카드 id 를 제거한 중립 서술 사용(TestTemplateNoInternalContentLeak GREEN — 1차 편집에서 적색 관측 후 수리).
- **이행 중 관측·수리**: 신규 stub 텍스트의 첫 판을 10,000 예산 초과(10,630)로 잰 뒤 3회 절단으로 9,915 에 수렴 — 스텁 게이트가 실제로 판정함을 관측(모터). TestTemplateNoInternalContentLeak 은 첫 hooks 편집의 SPEC-id 인용을 잡아 수리 — 경계가 생존함을 관측.
- **최종 검증**: 양 패키지 전체 스위트 FINAL (슬롯 임대 하) — 하기 증거 참조. `GOOS=windows GOARCH=amd64 go build ./...` exit 0, darwin/arm64 exit 0. golangci-lint(./internal/hook/... ./internal/template/...) 0 issues. 커버리지 internal/hook **87.5%**(≥85% 문턱).
- **AC-RIB-008**: 배포본 `## Isolation…` 절 생존 + 이동 금지 [HARD] 문장이 절 안 비표지 본문으로 생존(ALB-0294 전문은 스텁 § Isolation and integration 이 운반 — 항상 로딩 표면). 로컬 규칙 지목(`gitflow-lane-protocol.md` §1 → factory-dispatch.md 의 Isolation 절)이 살아 있는 절로 해석됨 — 로컬 파일 무편집.

### 라이브 재진입 확인(AC-RIB-012) — 리더 인계 Gap

레인은 스스로 세션을 재기동할 수 없음(단말 런처 경계 — `moai cc -w` 는 터미널 전용, 운영자 10-09 실측). AC-RIB-012 의 라이브 관측(startup + clear 재진입에서 넘침 공지 부재 + 합본 ≤10,000)은 **리더/운영자 협력 항목으로 인계**한다. 측정 대체 증거: 조립 예산 테스트가 생산자 상한 4,797 + 오버헤드 257 + core 3,786 = **8,840 ≤ 9,000 ≤ 10,000** 을 양 트리에서 단언(경계 소스 봉투의 산술 전체) — 라이브 관측이 남긴 Gap 은「실제 세션 착화 1회」뿐이다.

```text
[M3 verbatim, tree 070025bfc + M3 edits]
$ awk '/REQ-ALB-009 retreat/{n++} END{print n; exit !(n==1)}' internal/hook/role_rules.go
1
exit code: 0                                        [E6 flip]
$ grep -c '10000 characters' internal/hook/role_rules.go
0
exit code: 1                                        [no unseparated rendering]
$ grep -c '10,000 characters' internal/hook/role_rules.go
1
exit code: 0
$ grep -c 'moai update' internal/hook/role_rules.go
5                                                   [E4 flip: 4 locales + skew-detail]
exit code: 0
$ grep -c '10,000' .claude/rules/moai/core/hooks-system.md
1                                                   [E7 flip]
exit code: 0
$ go test ./internal/hook/ ./internal/template/      → ok / ok (FINAL, exit 0, slot internal-hook-template-suite)
$ GOOS=windows GOARCH=amd64 go build ./...           → exit 0
$ GOOS=darwin GOARCH=arm64 go build ./...            → exit 0
$ golangci-lint run --timeout=2m ./internal/hook/... ./internal/template/... → 0 issues
$ go test -cover ./internal/hook/                    → coverage: 87.5% of statements (>=85%)
```

- **card-review (턴 종료, lane)**: `codex_review` scope=card → `.moai/reports/t1617/card-review.md` (기계 로컬·gitignored — 판정을 여기에 이중 기재). Backend codex; base `db0c514d3` (리뷰 시점 머지 베이스); head `190f7db5d`; verdict **pass**, findings **0**, dispositions n/a. Advisory — 영수증은 리더의 증거 읽기와 독립 감사를 대체하지 않습니다.

## §E.3 Run-phase Audit-Ready Signal

run_status: audit-ready
run_complete_at: 2026-10-10
run_author: manager-develop
run_commits: 2b1fe8e8b (M1 RED) · 070025bfc (M2 rules) · 190f7db5d (M3 code/docs) · fcb726a0e (card-review 기록)
card_review: codex pass, 0 findings (.moai/reports/t1617/card-review.md; base db0c514d3 → head 190f7db5d)

패스 완결 성명 (§E.3 — verification-claim-integrity §3 형식):

| AC | 상태 | 결정 증거 / 명령 |
|---|---|---|
| AC-RIB-001 (E1+E2) | PASS | TestRoleInjectionAssemblyBudget ok — core 3,786 ≤ 3,946 (양 트리) |
| AC-RIB-002 (E3) | PASS | relocation-ledger.md — 36+3행, 목적지 미명명 삭제 0건 |
| AC-RIB-003 | PASS | TestRoleCoreStubGate ok — motor (표지/초과 픽스처 FAIL) |
| AC-RIB-004 (E4) | PASS | 4개 로캘 + detail 에 moai update (grep=5) |
| AC-RIB-005 (E5) | PASS | 술어 3케이스, 픽스처만 — 라이브 자기단정 없음 (grep 확인) |
| AC-RIB-006 (E6) | PASS | awk n=1, exit 0 — 무천단위 '10000 characters' 0히트 |
| AC-RIB-007 (E7) | PASS | hooks-system '10,000' — 미러+배포, 쌍 바이트 동일 |
| AC-RIB-008 | PASS | Isolation 절 생존 + 이동 금지 [HARD] 문장이 절에 존재 |
| AC-RIB-009 (E8) | PASS | stub-delta 기록 + rule-authoring 비산 문 (§E.2) |
| AC-RIB-010 | PASS | 기존 21 사다리 서브테스트 + Handle 통합 — padding 후 전부 GREEN |
| AC-RIB-011 | PASS | 원장: 갱신 전 RED 관측 → 갱신 뒤 ok |
| AC-RIB-012 | **리더 인계** | 라이브 재진입 — 레인 권한 밖 (아래 참조) |
| AC-RIB-013 | PASS | 봉투 밖 사다리 서브테스트 + codex_review_ownership 도크트린 GREEN |

- **AC-RIB-012 (라이브 재진입 확인) — 리더/운영자 인계**: 레인은 스스로 세션을 재기동할 수 없습니다 (단말 런처 경계 — `moai cc -w` 는 터미널 전용, 운영자 정정 10-09). 라이브 관측 (startup + clear 재진입에서 넘침 공지 부재 + 합본 ≤10,000) 은 리더/운영자 협력 소관으로 인계합니다. 대체 증거: 조립 예산 테스트가 생산자 상한 4,797 + 오버헤드 257 + core 3,786 = **8,840 ≤ 9,000 ≤ 10,000** 을 양 트리에서 단언 — 경계 소스 봉투의 산술 전체가 관측됨. 남는 Gap 은 「실제 세션 착화 1회 관측」뿐입니다.
- **검증**: go test ./internal/hook/ ./internal/template/ → ok/ok (FINAL, exit 0 — 슬롯 internal-hook-template-suite 임대 하); GOOS=windows GOARCH=amd64 및 darwin/arm64 빌드 → exit 0; golangci-lint (hook+template) → 0 issues; coverage internal/hook **87.5%** (≥85% 문턱).
- **card-review**: codex scope=card → **pass**, findings 0 (`.moai/reports/t1617/card-review.md`; base db0c514d3 → head 190f7db5d) — advisory이며 리더의 증거 읽기와 독립 감사를 대체하지 않습니다.
- **관측 (경계 생존 실측)**: 스텁 추가문의 초판이 스텁 게이트에 걸렸고 (10,630 > 10,000 — 3회 절단으로 9,915 수렴); TestTemplateNoInternalContentLeak 이 hooks 첫 편집의 SPEC-id 인용을 잡았습니다 — 두 경계 모두 생존하고 실제로 판정했습니다.

## §E.4 Sync-phase Audit-Ready Signal

sync_complete_at: 2026-10-10T19:07:08+09:00
sync_commit_sha: 11480a26e989182a77ff8eacf7e4b391abda3d37
sync_status: complete
sync_agent: manager-docs
sync_scope: CHANGELOG 엔트리 + 3-phase close (spec.md frontmatter status 전이; plan.md/acceptance.md는 frontmatter 블록 부재로 `updated:` 리프레시 대상 없음 — 본문 무편집)
changelog_entry_position: "[Unreleased] › Fixed — first entry (prepended)"
changelog_duplicate_grep: 0
b12_self_test_a: pass — 발행 전 `grep -c 'SPEC-ROLE-INJECTION-BUDGET-001' CHANGELOG.md` → 0 (exit 1, 중복 엔트리 없음)
b12_self_test_b: pass — 선언 live AC 집합 = 13 (acceptance.md §A 표 AC-RIB-001..013 식별자 전수 대조, excluded 0 ambiguous 0); CHANGELOG 엔트리의 13 기재와 일치
b12_self_test_c: pass — `ls`로 구현 파일 5건 존재 확인 (internal/hook/role_rules.go · role_injection_budget_test.go · role_rules_skew_predicate_test.go · role_rules_version_skew_test.go · internal/template/role_core_stub_gate_test.go)
frontmatter_status_transitions.in-progress→implemented→completed: 단일 sync 커밋 병합 전이 (spec.md `status:` + `updated:` 만 편집; 본문 무편집 — updated는 당일 값 유지)
mx_tag_validation: sync diff는 docs-only (CHANGELOG.md + spec.md frontmatter + progress.md §E.4) — 태그 변경 0
ac_rib_012_disposition: 「운영자 확인 대기, 대체 증거 = 봉투 산술 실측」 — 레인·리더 모두 세션 재기동 불가(단말 런처 경계, 운영자 정정 10-09)로 리더가 운영자 복귀 체크리스트(보드 card:t1617)에 등재; 대체 증거는 조립 예산 테스트의 생산자 4,797 + 오버헤드 257 + core 3,786 = 8,840 ≤ 9,000 ≤ 10,000 양 트리 관측
readme_docs_site_judgment: no-sync — 이 변경은 내부 주입 동작(예산 게이트 · 스텁 구조 · 오류 안내 문구)과 룰 파일 내용의 재배치로, README 4-locale · docs-site가 서술하는 사용자 대면 동작은 불변 (문서화된 CLI 표면 변화 없음)
sync_evidence: 최종 런 스위트(ok/ok) · 커버리지 87.5% · lint 0 · 양 크로스컴파일 빌드는 §E.3 실측을 그대로 인용 — sync diff는 문서 전용으로 코드·룰 트리 무변경
commit_plan: sync 커밋 1건(전이 · 엔트리 · §E.4 플레이스홀더 동승, `Authored-By-Agent: manager-docs` + `🗿 MoAI` 트레일러) + backfill 커밋 1건(실측 SHA로 이 칸 갱신)
backfill: done (sync 커밋 11480a26e989182a77ff8eacf7e4b391abda3d37 로 이 칸을 갱신 — 본 backfill 커밋이 뒤따른다)

Sync 요약: sync 페이즈는 코드·룰 트리에 0 변경을 두고 세 문서 표면만 만졌다 — (1) CHANGELOG [Unreleased] › Fixed 첫 항목 발행, (2) spec.md frontmatter `in-progress → completed` 전이(implemented 중간점 병합), (3) 이 §E.4 시그널. 런 페이즈의 검증 증거(§E.3 표 12/13 PASS + AC-RIB-012 리더·운영자 인계)는 sync diff가 docs-only라 그대로 유효하다. 판정: audit-ready — 리더가 본 트리에서 sync 감사(codex_audit, target=baseBranch)를 이어간다.

### 착지 준비 — 로컬 main 흡수 (리더 지시 d-20261011T010153Z-16ac, 2026-10-11)

t1616이 로컬 main(`478285ab4`)에 착지한 뒤, 카드 트리가 이를 흡수했다.

- **흡수 커밋**: `c73b0ca6e` — `merge: absorb local main 478285ab4 (t1616 landing) into WT-3-2-1 (card t1617)`; 병합 기저 `2aab5f797`, 충돌 2건
- **충돌 1 — CHANGELOG.md**: 양측 [Unreleased] › Fixed 엔트리 유니온 병합(본 카드 SPEC-ROLE-INJECTION-BUDGET-001 + main 유입 SPEC-CODEX-CONFORMANCE-001 card t1607 — 양측 원문 보존)
- **충돌 2 — internal/template/testdata/binding_ledger.json**: 본 카드가 JSON 재직렬화(들여쓰기 2칸)로 36행 갱신, main이 ALB-0203·0204의 after_text(Haiku 5.5 반영)를 편집 — 3-way 의미 병합(python json, base=`2aab5f797` 버전 대조): main의 2행 after_text만 ours에 적용, 본 카드의 나머지 갱신 전부 보존. 검증: `our_edits_lost: none` · `main_edits_in: True` · rows=363 양측 행 집합 일치
- **흡수 트리 검증 원문 (2026-10-11, 트리 `c73b0ca6e`)**:
  - `go test ./internal/hook/ ./internal/template/ ./internal/cli -run 'TestRoleInjectionAssemblyBudget|TestRoleCoreStubGate|TestAutoPick|TestAutoRank|TestAutoDoc'` → `ok internal/hook 1.025s` · `ok internal/template 0.640s` · `ok internal/cli 28.314s`
  - `go test ./internal/hook/` → `ok github.com/modu-ai/moai-adk/internal/hook 558.657s`
  - `go test -timeout 20m ./internal/template/` → `ok github.com/modu-ai/moai-adk/internal/template 344.218s` (첫 시도는 go test 기본 10분 상한 초과로 중단 — 실패 판정 아님, 상한 재실행)
- **git-strategy.yaml 확인**: 작업 사본 `develop_branch: main` (`.moai/config/sections/git-strategy.yaml:16`)
- **착지 diff 기준**: `2aab5f797` (리더 기록) — 흡수 후 병합 기저는 `478285ab4`

### 델타 감사 (sync-audit-delta1) 처분 — 리더 판정 d-20261011T030402Z-4f2c (2026-10-11)

- **델타 자체 PASS**: 흡수 병합 · ledger 3-way · CHANGELOG · 문서 커밋 — 리더 착지 관문에서 PASS 판독 (델타 감사는 리더 착지 관문이지 엔진 전이가 아니므로 카드는 merge-ready 유지, 엔진 정산 불요)
- **Gaps**: codex P2 3건은 이전 감사 SHA f40bfb59f에 이미 존재(감사자 git show 대조), 델타 무관, D1 복원·D2/D3 → t1637 (리더가 t1617 relates-to t1637 relate 기록)
- **RIB-D1 복원 (4f2c D1 → (a) 원래 집 복원, 스텁 예산 10,000 보존(측정 9,915+133=10,048 초과로 스텁 삽입 기각))**: 구 트리 `2aab5f797` factory-dispatch.md:116 행을 live·템플릿 미러 양 트리 "The dispatch cycle" 절에 바이트 그대로 복원 — 스텁 무접촉(REQ-RIB-011 예산 영향 없음). 문장: "`card`, `cmd`, `wt`, and `evidence` are always present. `spec` joins once a SPEC exists; a Class B card carries none, and its `evidence` names whatever record the leader will read instead. `lens` appears only when the leader names review lenses for the card's sync gate."
- **RIB-D2 (role_rules.go:328 YAML 문자열 절단) · RIB-D3 (role_core_stub_gate_test.go:72 WalkDir 오류 삼킴)**: 부채 기록, 후속 카드 t1637 귀속 (리더·영실이 판단, 착지 비차단)

### 델타 감사 2라운드 (sync-audit-delta2) — 착지 관문 판정과 B1 수리 (2026-10-11)

- **델타2 판정**: FAIL 0.6 (must_pass_failed 1 · blocking 1, codex required fail 수렴 rcpt-caa859f7b7b708666c0552d2) — 판정 파일 `.moai/reports/t1617/sync-audit-delta2.md`. **auditor-model 공백 공시**: codex_audit 결과 JSON에 model 필드가 없어(키: verdict/summary/findings/next_steps/audit_receipt/build_commit/review_base) 응답 모델 id를 원문으로 남길 수 없음을 판정 파일에 명시 — 규칙 3대로 기재, 리더 관문 판단 귀속
- **B1 (블로킹, 본 델타)**: 복원 행이 `2aab5f797:116`과 바이트 비일치 — 원문은 `- ` 목록 항목(273B)인데 복원은 백틱 시작(271B), 2바이트 누락. 커밋 `5c788d57a` 메시지와 본 progress의 "바이트 그대로" 주장이 엄밀히 거짓이었음 — **정정**: 그 주장은 `5c788d57a` 시점 기준 참이 아님(내용 복원은 정확, 목록 접두 누락)
- **B1 수리**: 양 트리 복원 행에 `- ` 접두 부여 → 감사자 기준 시험 `grep -cxF -f <sed -n '116p'>` = **양 트리 1** (바이트 일치). 검증: `TestAutoPickMirrorParity|TestAutoPick|TestAutoRank|TestAutoDoc` ok 16.291s · `TestRoleCoreStubGate` ok 0.447s (-count=1)
- **재확인 표 (감사자 관측 인용)**: D1 — f40bfb59f에서 결함 형태 존재(스텁 양쪽 문장 부재), 본문은 본 수리로 변경 · D2 — f40bfb59f 존재(:328 strings.Cut), 미변경 · D3 — 존재(:72 WalkDir 삼킴), 미변경 · 신규 P2 `role_rules.go:345` (os.ReadFile) — f40bfb59f 존재, 미변경 → **t1637 부채계열 추가** · 신규 P3 `hooks-system.md:132` (23,166자 내부 인용) — f40bfb59f 존재, 미변경 → **리더 처분 (t1637 범위 주석 또는 별도)**
- **스텁 예산 수치 정정**: 이전 기록 "9,915+133=10,048"은 리더 룰링 인용치 — 자체 미재측정(감사자 지적)

### F4 수리 시도 — 예산 불가 판정 (리더 지시 ②-(가), 2026-10-11)

- **RED 시험 선커밋**: `f2f2b67ec` — `TestInjectedCoreCarriesCodexReachableObligations` (internal/hook/role_core_reach_test.go): 주입 role-core 본문(config.ExtractRoleCoreRegions)에 두 의무 핵심 문구("fixed-field"·"disk wait") 존재 기계 단언. **RED 관측 원문**: 양 문구 × 양 트리 4건 전부 "the injected role-core body does not carry" — F4 재현.
- **현재 실측**: core 3,786 (양 트리 동일, 31구역) — 천장 3,946, 여유 **160**
- **의미 보존 압축문 측정** (원문은 스텁 압축판이 담당하는 두 의무 전문):
  - A(디스패치 고정 필드): 「[HARD] The dispatch block is fixed-field (`card:` `spec:` `cmd:` `wt:` `evidence:`; `lens:` optional) — never prose, at most 10 lines.」 = **134**
  - B(진행 불가 레인 디스크 대기): 「[HARD] A stuck lane records a disk wait and ends its turn — never idles on a reply; on awaken: the stall watchdog runs first, then the §6 ladder.」 = **145**
  - 합계 +구분 2 = **281 > 160 (121 초과)** → **수리 불가**
- **축소 한계**: 필드 목록을 빼면 122로 들어가지만 의무의 실체(필드 명세)가 사라져 조건 ①(의미 비축소) 위반 — 기각
- **판정**: (가) 불가 → **(나) 주입 경로 변경(설계 변경) — 영실이 판단 대기**. 수리 없음, RED 시험 `f2f2b67ec`는 미래 수리의 게이트로 잔존
- **Gap**: ② 격리 HOME 라이브 주입 확인은 수리가 없어 미실행(검증 대상 변경 없음 — 부재는 RED 시험이 기계 증명); Codex 세션 실주입 미관측 Gap은 (나) 판정 뒤 수리 시 소멸

### F4 수리 — 1순위 상쇄 성공, 표지 내 재삽입 (리더 지시, 2026-10-11)

- **상쇄 (비의무 문구 164 ≥ 121, [HARD] 실체 무손상)** — 전·후 원문과 바이트:
  - [12] 제거 (106, [HARD] 없는 설명 행 — 내용은 cross-session-messaging이 소유): 전 「A message is a nudge, never the delegation — an idle notice says when to look, not what the evidence says.」 → 후 «삭제»
  - [15] 예시 상세 절감 (58): 전 「[HARD] Every lane runs a `card-review` (`codex_review` → `.moai/reports/<card-id>/card-review.md`) before integration — advisory only.」 → 후 「[HARD] Every lane runs a `card-review` before integration — advisory only.」 (도구·경로 상세는 factory-dispatch-detail.md § The card-review stage가 소유)
- **재삽입 (표지 안)**: A 「[HARD] The dispatch block is fixed-field (`card:` `spec:` `cmd:` `wt:` `evidence:`; `lens:` optional) — never prose, at most 10 lines.」(134, dispatch cycle 절) · B 「[HARD] A stuck lane records a disk wait and ends its turn — never idles on a reply; on awaken: the stall watchdog runs first, then the §6 ladder.」(145, lane waits 절) — 원문 대조: 스텁 압축판과 동일 의무, §6·필드 목록 보존
- **관문 실측**: RED `f2f2b67ec` → **GREEN** (`TestInjectedCoreCarriesCodexReachableObligations` ok 0.991s) · core **3,903** ≤ 3,946 · 합성 **8,957** ≤ 9,000 (양 트리 32구역 동일) · MirrorParity+문서계열 ok 29.996s · StubGate ok 0.517s
- **천장 개정**: 불요 — 상쇄 164가 281을 흡수(순증 +117)
- **Gap (② 라이브 주입 확인 미실행)**: 리더 지정 형태 `env -i HOME=$(mktemp -d) … bin/moai hook session-start`가 worktree 세션 가드에 3형태 전부 거부됨(`env -i` 복합·`env -i` 단순·`env HOME=` 오버라이드 — "injecting git configuration whose effect cannot be verified"). 조립 본문의 문구 존재는 GREEN 시험이 기계 증명(ExtractRoleCoreRegions = buildRoleCore의 반환 본문 그대로), 미관측은 훅 end-to-end 호출(역할 감지→파일 판독→조립)의 선행 절 — 다음 Claude 세션 착화 시 소멸 가능한 Gap
- **부채 → t1637**: F1·F1b·F5·F3, F2 선택 보강, 스텁 spec: 문구, RIB-D2(:328 YAML 파싱), :345 os.ReadFile, budget-test env 격리, hooks-system.md:132 내부 인용(P3)

### 천장 개정 (SPEC 0.4.0 → 0.5.0) — 리더 지시 ceiling raise (2026-10-11)

본 카드 SPEC 본문의 구속 수치를 측정 수요로 올리는 개정. SPEC 표면은 manager-spec이(본 절), 테스트 상수는 레인이(본 커밋 뒤 후속) 나눠 갖는다.

- **개정 수치** (측정 수요, 패딩 없음): 역할 core 천장 **3,946 → 4,073** UTF-16 · 조립 합본 예산 **9,000 → 9,127** (= 생산자 상한 4,797 + 오버헤드 257 + core 4,073 — 파생) · 변이 경계 3,947 → **4,074**. 하드 캡 10,000 · 스텁 예산 10,000 **무변경**.
- **사유 (운영자 언어 그대로)**: 「Codex 전용 레인에서 [HARD] 의무 2개가 사라지는 회귀 방지, 상쇄는 Codex 전달 의무를 깎아 기각」 — M2 재작성과 그 뒤 offset 절단이 각각 Codex 전용 설치 레인에 레인 의무의 전달 경로를 남기지 않았고, 두 의무 복원 + 신규 2개(F4 수리)에 구 천장보다 127단위가 더 필요하다.
- **offset 대안 기각 (리더)**: 의무 아닌 core 문단을 깎아 자리를 만드는 대안은 2회 codex 델타 라운드에서 Codex 도달 의무 전달을 깎는 것이 입증돼 기각 — 영수증 `rcpt-f32250556030f729fbb7f86c` · `rcpt-356d2c64e2dec47d24e28669`.
- **결정 사슬**: 영실이 위임 판단 → 리더 처분 `d-20261010T075659Z-7d7c` → `d-20261011T030402Z-4f2c` → offset 기각(상기 영수증) → 본 천장 개정.
- **측정 귀속 (본 세션 직접 관측 — 커밋 시점 트리, HEAD `49862d316a` + 미커밋 복원분)**: region-split UTF-16 합산 명령(`python3 -c "…split('<!-- moai:role-core-start -->')…"`) — deployed core **4,073** (33영역) · template core **4,073** (33영역), 양 트리 동일. core 4,073 = 천장 4,073 — 여유 0 (측정 수요 그대로).
- **갱신 표면**: spec.md(제목·§A.4 주석·§A.5 1행·§B core 예산/설계 목표·REQ-RIB-001/002/004·§D 분할 기각 산술 + HISTORY 0.5.0 행) · plan.md(헤더·§A.1·§B core 예산·M1 상수 주석·M2 기록 주석) · acceptance.md(AC-RIB-001 3칸·§B 변이 탐침·§C 원장 헤더·§C 참고 산식) · 본 절.
- **보존 (개정 아님)**: acceptance.md §C 증거 원장의 E1/E2 축자 명령·출력(`n<=3946`)은 채택 시점 관측 원문 — 무편집. progress.md §E.1~§E.4의 기존 기록 행(파생 수치·완료 기준 원문·core 3,786 ≤ 3,946 · core 3,903/합성 8,957 ≤ 9,000 등)은 각 시점의 계획·관측 기록 — 무편집. relocation-ledger.md · `.moai/reports/t1617/measurements.md` 는 run 이력 — 무편집. `.claude/rules/`·템플릿 규칙 트리 그림 스윕(3,946|9,000|3,800|3946|3947|9000|3800, 양성 대조 통과) — 히트 0, 갱신할 미러 없음.
- **대기 단계 (레인 소관, 본 커밋 뒤)**: `internal/hook/role_injection_budget_test.go` 상수 `budgetCompositeLimit = 9000` → `9127`, `budgetCoreCeiling = 3946` → `4073` — 갱신 전 본 트리에서 조립 예산 테스트는 **예상 RED** (core 4,073 > 구 천장 3,946).

### F4 최종 수리 — 기계 대조 결과과 중단 (리더 지시 ②, 2026-10-11)

- **복원 + 천장 개정**: [12]·[15] 원문 복원, A·B 표지 안 유지 → core **4,073**(33구역, 양 트리) · 합성 9,127 — SPEC 0.5.0 개정 커밋 `70caf7685`(manager-spec), budget test 상수 갱신(9127/4073, target 3800 superseded 주석)
- **검증 원문**: reach+budget `--- PASS` ×2 / `ok internal/hook 0.756s` · cli 계열 `ok 13.269s` · StubGate+LedgerIntegrity `ok internal/template 40.533s` (전부 -count=1)
- **기계 대조 (base 478285ab4 주입 본문 vs 수리 후 본문, 줄 단위 포함)**: base 본문 줄 58 · 수리 후 줄 33 · **미일치 58/58(100%)** — 목록 원문:

```text
[HARD] **The leader is the queue's sole producer.** The operator asks; the leader turns the request into a card with `moai gtd add "<description>"` (`moai gtd` alone lists the queue). Production is the one queue mutation the leader performs on its own authority — translation, not invention: nothing enters the queue the operator did not ask for.
[HARD] **Standing sources are the other producers, and they produce on the operator's prior authorization.** `/moai project` issues one card when it completes (`[PROJECT] ` prefix); the codemaps-debt trigger in `moai integration release` issues one per debt period (`[GRAPH] ` prefix). The full conditions — one per occasion, derived not invented, marked, id reported, starting still a separate pick — are the SSOT at `.claude/skills/moai/workflows/gtd.md` § Standing sources, and that set is **conditional, not closed**. Nothing else produces; a report milestone, an audit finding, or an open issue still reaches the queue only as a card request the operator approves.
[HARD] **Promotion is the operator's act, in person or in advance.** After a `/clear`, the leader presents the queued cards through `AskUserQuestion` and the operator picks; only then does the leader route the card whole to a free lane. Outside an --auto authorization the leader never picks for the operator, never reorders by inferred priority, and never silently promotes a backlog item. An empty queue is a state to report, not a prompt to invent work. The one ranking the queue admits is the `--auto` cycle's own, bounded in the next paragraph.
[HARD] **The self-dispatch lane exception.** In a self-dispatch factory run, a lane session may lease the next queued card — the one promotion a lane performs — only through `moai factory next` — bare, or `--card <id>` for its own judged pick. Every other queue mutation (`add`, `drop`, `done`, `edit`, and the rest) and `moai contract sign` stay forbidden to a lane: the lane works the operator's queue, it never authors it.
[HARD] **The leader may attach a finding; it may not act on one.** Analysis records a relation between two cards (`moai gtd relate`); the record is evidence the operator reads, never a mandate — the leader never folds the related card away, never reorders the queue around it, and never drops or edits it. Analysis changes exactly one thing on its own authority: it refuses the admission of a card whose normalized text is identical to one already queued or picked.
[HARD] **The pre-dispatch PR cross-check.** Before dispatching a card out of `backlog`, the leader reads that card's pull-request and landed state (`moai gtd pr <id>`; by hand `gh pr list` plus a `git log` against the integration branch) and reports what it read in the same turn. An unchecked card is a gap, not a clean card (§ Completion is read, never trusted).
[HARD] **The cross-check also asks whether a completed SPEC already covers the work.** A card id cannot answer "has someone else already done this" — where the card names an issue or a subsystem, the leader also reads whether a covering SPEC is already `completed` and reports that alongside the PR and landed state. Neither read is conclusive: the final discriminator stays reproduction.
[HARD] **The cross-check reports; it never vetoes.** Where the card carries an open pull request or is already landed, the leader surfaces that and the operator **confirms or withdraws** it. The leader never withholds a picked card on its own authority. Why the wording is the only available control: `factory-dispatch-cards.md` § The pre-dispatch cross-check.
[HARD] **A milestone-bearing report under `.moai/reports/` carries a `## Card Cross-Check` section** — one table row per milestone, a `card` column holding the delivering card id or an explicit new-card marker. A mapping claim is verified against the queue (`moai gtd`), never remembered. Before the leader turns a report into card requests, the request message states the full comparison — `N milestones → N cards` — naming every milestone with no card in the live queue. Detail: `factory-dispatch-cards.md` § Report milestones ↔ queue cards.
[HARD] **Class A is admitted on checked evidence, not on an assertion.** Two of its three properties are mechanically checked and cited: the diff is measured (`git diff --stat` against the base, showing the one file) and CI is green **on the head that will merge**. The third — no design judgement in it — is a judgement, stated in the dispatch where the operator can disagree with it. A card that cannot cite both measurements is not Class A. The justification is never "it is faster".
[HARD] Work is delegated through the queue on disk, not through messages. The queue file resolves against the primary checkout from every linked worktree — one repository, one queue — so a card admitted from anywhere is visible everywhere, and that single-file visibility is what makes the queue a channel rather than a shared opinion.
A cross-session message is a nudge, never the delegation itself. **A lane can be asked to report when it next goes idle** (`SendMessage` `notify_when_idle`), but [HARD] the notice is not the completion signal — it says *when to read the evidence*, nothing about what it says. **A nudge can be refused outright under fan-out**; read the send result, and a refusal costs the run nothing because the queue already carries the delegation (`cross-session-messaging.md` § An idle notice is a scheduling hint, § Configuration surface).
[HARD] A dispatch is written in the operator's `conversation_language` — the operator watches it scroll past, which makes it user-facing output rather than internal agent traffic. The boundary is **who reads it**: an `Agent()` subagent prompt reaches no human and stays English. What stays verbatim in every language: SPEC IDs, command names and their flags, file paths, session names, and technical identifiers. (Classification rationale: `factory-dispatch-detail.md` § Dispatch language.)
[HARD] A dispatch is a fixed-field address block, not prose. The fields:
```
card: <id>
spec: <SPEC-ID>
cmd: /moai run <SPEC-ID>
wt: .moai/worktrees/<name>
evidence: .moai/specs/<SPEC-ID>/progress.md
lens: --security --deep
```
- `card`, `cmd`, `wt`, and `evidence` are always present. `spec` joins once a SPEC exists; a Class B card carries none, and its `evidence` names whatever record the leader will read instead. `lens` appears only when the leader names review lenses for the card's sync gate.
- `wt` names the new card's worktree, never a previous card's tree. The tree keeps the card id; the branch takes a descriptive slug (§ Isolation below).
- **No explanatory prose.** Procedure, background, and justification live in the card text and the SPEC artifacts the block points at.
- **Ceiling: the block is at most 10 lines.** A dispatch that does not fit is trying to be a handoff; move the payload into the card and send the block.
- **[HARD] The send is read, not assumed.** A `routing` object on the result means an in-process mailbox took the block — lost (re-send to `name [ref]`); a following `[Cross-session delivery notice]` means the lane's permission policy is holding or refused it (surface it to the operator, do not re-send); anything else queued it. None of the three establishes that the lane's Claude read it. Full shape table: `cross-session-messaging.md` § A send result has three shapes.
[HARD] A lane session tracks its card's execution on the session's task tools. At card intake it registers the card's execution stages via `TaskCreate` before beginning the first stage; at every stage transition it keeps the list current via `TaskUpdate` so the list always shows the stage in flight; and it reports completion only while the list reflects the end state — or carries an explicit annotation naming why it does not. A task list that contradicts its completion report is the same gap as a missing evidence file (§ Completion is read, never trusted).
[HARD] Between run-exit verification and integration every lane runs a `card-review` stage: a card-scope self-review through `codex_review` (`glm_review` optional) whose result is written to `.moai/reports/<card-id>/card-review.md` — advisory only, never a replacement for the leader's evidence read or an independent audit. With `tree_scope: skip` configured for the leader's checkout, the leader session carries no turn-end codex review gate and reviews its own internal output directly with the same tools. Stage order, evidence fields, and the re-review ceiling: `factory-dispatch-detail.md` § The card-review stage.
[HARD] A lane that cannot proceed records an explicit wait on disk — reason, whom, recheck point — and ends its turn; it never idles open-ended on a reply. On its next awaken the lane runs the stall watchdog first — invoke Skill("moai-lane-watchdog") and follow it — which measures progress, classifies the stall cause, and resolves the judgment through the decision ladder (doctrine: `.claude/rules/moai/workflow/auto-semantics.md` §6). The ladder's terminal step is the lane's own judgment + decision record — the lane never asks the leader and waits on a judgment; reaching the leader is a gate boundary (a keep-set category or a cross-card conflict), not a ladder step. The lead's decision board carries leader-level rulings only — cross-card conflicts, serial-slot policy calls, keep-set gate relays — never a lane's in-card judgment. A lane also arms a standing recheck cron (`CronCreate`, recurring, every 20 minutes) before its first stage and keeps it until its completion report: a lane stopped by an API error (429) or by a delegate's report that never comes wakes without outside help only if something armed before it stopped fires, and every wake reads the disk evidence (progress record, reports, commits, the delegate's deliverable) before it replies (`.claude/rules/moai/workflow/auto-semantics.md` §5.1).
[HARD] **The deputy is resident, not optional.** Before the batch's first lane dispatch, the factory leader session spawns exactly one UNNAMED background `Agent()` running manager-lead as its **coordination deputy** and keeps it for the batch. Its delegable/retained matrix lives in the agent itself (`manager-lead.md` § Deputy dispatch surface); this stub carries only the boundary that binds the queue.
[HARD] **A completion report reaches the leader as a `RECOMMEND:` summary, not as raw reading batches** — turn occupancy moves; the leader's own evidence-read before advancing a card does not (§ Completion is read, never trusted).
[HARD] **The deputy never holds a power of consequence.** Final PASS/FAIL verdicts, final merge approval (`LEAD-MERGE-APPROVED`), operator gates, card issuance and `done` (`moai gtd` mutations), CodeRabbit slot-wait adjudication, and cross-session dispute coordination stay with the leader session. Round-report measurement and drafting are the deputy's; the asserted figures are the leader's — every figure names its measurer. Nothing structural moves with the delegation: the queue stays the channel, completion stays evidence the leader read, and the verdict's home stays the lead.
[HARD] The leader advances a card on **evidence it read**, not on a lane's reply. Before moving a card out of a working stage, the leader reads the card's `progress.md` and the verification evidence path the phase declares; a missing, unreadable, or stale evidence file is a **gap** — the card stays put and the leader reports why. Absence of a failure signal is not a pass.
[HARD] A `gh pr checks` row naming CodeRabbit is not evidence that a review happened: the status is `success` **even when no review ran**, and the row prints `pass` byte-identically in both cases. A row counts only when BOTH hold:
1. The **combined** endpoint `/commits/{sha}/status` shows `state == "success"` **and** description `Review completed`:
```bash
gh api "repos/$REPO/commits/$HEAD_SHA/status" \
--jq '.statuses[] | select(.context == "CodeRabbit" and .state == "success") | .description'
```
2. A `Merge Risk:` line exists whose commit prefix matches the current `headRefOid`.
[HARD] A lane session does not carry one card's context into the next card. When a card reaches `done` and the leader has read its evidence, the leader **asks the operator to `/clear` that lane session** — `/clear` is a user-typed command and cannot be sent as an instruction. The leader's message states, in order: what closed (card, evidence read), which session to `/clear` (by name), and what happens next.
[HARD] On a card transition inside a lane session, `/clear` happens exactly once, **after** the session has moved into the next card's worktree — not before the move. The count is not reduced: one card transition, one `/clear`.
[HARD] A card's work happens inside a worktree, and that worktree is **entered through the launcher** — never created with a bare `git worktree add`. Launcher table (per-harness forms): `factory-dispatch-mechanics.md` § Isolation.
[HARD] A Codex factory agent must never call `moai cc -w`, `EnterWorktree`, or `ExitWorktree`; those are Claude Code entry mechanisms.
[HARD] **`moai worktree done` closes L2 trees only** — project trees under `.claude/worktrees/` and `.moai/worktrees/` are L1 and absent from its registry. The full L1/L2 boundary lives in `worktree-integration.md` § Terminology Glossary.
[HARD] **The card's branch is unpushed, so its worktree is the work's only instance.** Dispose of no worktree — L1 or L2 — until the leader has integrated the branch and the remote merge has landed.
[HARD] **A card session starts inside its worktree and stays there.** The standard is a lane session launched inside the card's worktree through the launcher (`moai cc -w`); it starts with that tree's project skill set. Moving a running card session from one card worktree into another card worktree is prohibited — that transition attaches a second skill listing on top of the session's first, and that duplication is what this rule exists to prevent. Two composition rules bound the prohibition rather than leaving it implicit: (a) the "A new card starts in a new worktree" clause below composes with it — its exit-first requirement is satisfied by ending the lane session and launching the next session inside the new tree, never by moving the running session; (b) the brief entry into the release/integration worktree that the same files mandate for merging a card branch is NOT covered by the prohibition — it keeps its own re-entry rule (return to the card worktree before continuing card work). Where a worktree move is unavoidable anyway, `/clear` is issued exactly once, after the move — never before it.
[HARD] **A new card starts in a new worktree** — a lane anchored in the previous card's tree MUST `ExitWorktree` before entering the next one; the fresh tree is created from the configured base, never reused. Where the new card depends on a prior card's unmerged code, merge that branch inside the new worktree.
[HARD] **Card worktree branches carry the `WT-` prefix and a descriptive slug — never the card id** (rename in place with `git branch -m WT-<slug>`); the worktree directory keeps the card id, and traceability rests on the dispatch `card:` field, the commit message, the evidence path, and the PR title — all mandatory. Slug shape and carrier detail: `factory-dispatch-mechanics.md` § Isolation · `factory-dispatch-cards.md` § The PR-title carrier.
[HARD] **Inside a worktree session that `<path>` is the worktree's own absolute path** — the guard refuses `-C .`, relative paths, runtime-computed paths, and paths outside this worktree; plain `git`, `git -C <own absolute path>`, and `--git-dir=<own .git>` pass.
[HARD] **Lane-local verification is scoped to the card.** A lane runs the tests its own change can affect, then pushes and lets CI run the full suite. (Incident record: `factory-dispatch-gates.md` § Verification load incident record.)
[HARD] **Never spawn background load.** Where a verification genuinely needs contention, the load must be cleanup-guaranteed — kills registered with the test framework's cleanup hook, or a `timeout` wrapper bounding the process from outside.
[HARD] Inside a worktree, an environment-scrubbed verification runs as one compound `unset … && <command>` invocation — each Bash call is a fresh process, so a separate `unset` does not carry into the next command; the scrub and the command travel together, or the scrub does nothing:
```bash
unset MOAI_KANBAN_ID MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED && go test ./...
```
[HARD] A lane whose card has passed verification does not wait for the leader to integrate it: the lane merges its own branch into the batch's release branch (`release/vX.Y.Z`) itself. This is the **git-flow variant** — it applies only where the project's git strategy names git-flow; under github-flow (the distributed default) the delivery is `moai factory complete`'s pull-request edge and the full sequence is `factory-dispatch-mechanics.md` § The lane's standard landing. Where it applies, the window is taken with `moai integration acquire --name <lane> --card <card-id>` BEFORE entering the release worktree, released after the completion report is sent; the lane enters the release worktree with `EnterWorktree` (a cross-tree `git -C` is refused), merges `--no-ff`, re-reads `HEAD` before the commit and again before the push, pushes `release/vX.Y.Z` (never force), and leaves the batch pull request with the leader. The full window procedure: `factory-dispatch-mechanics.md` § Integration into the release branch is self-served · the `acquire` settings-drift assertion: `factory-dispatch-gates.md` § The pre-merge settings-drift assertion. Once the card's work is confirmed landed on the remote integration branch, disposing the card's worktree is the sweep's step: `moai worktree sweep` disposes remote-landing-confirmed trees across both tiers, hoisting evidence before every removal (`worktree-integration.md` § Hoist a tree's evidence before disposing it).
```

- **해석 (레인 기록, 판단은 리더/영실이)**: 58줄 전체는 M2 압축의 문서화된 흔적 — base의 장문 [HARD] 단락들이 재작성으로 압축 one-liner가 되어 줄 단위 일치가 성립하지 않음. 의미 보존은 363행 relocation ledger가 담당하며 plan-audit PASS 0.90(4라운드)·sync-audit PASS 0.95·delta1/2/3 감사가 그 ledger와 본문을 검증. **본 F4 수리의 신규 손실 없음**(A·B는 추가, [12]·[15]는 복원). 줄 단위 일치 기준을 그대로 적용하면 압축 전체가 '빠진 줄'이 되므로, codex 진행 여부는 리더/영실이 판정 — 규칙대로 codex 전에 중단.
### ⓐⓑⓒ 행선지 추적 (리더 지시, 2026-10-11) — base 478285ab4 주입 본문 줄 58 전수

**ⓐ 대응표 (base 줄 → 원장 행)**: 미매핑 **0건** — 58/58 매핑. ALB-0303은 before_text 스냅숏과 478285ab4 문언의 미세 드리프트로 엄격 포함이 놓쳐 토큰 자카드(≥0.5)로 확인(행 존재 확인, 원장 누락 아님). 매핑:
- ALB-0254 · role-core:workflow/factory-dispatch.md · «[HARD] **The leader is the queue's sole producer.** The operator asks; the leader turns th»
- ALB-0255 · role-core:workflow/factory-dispatch.md · «[HARD] **Standing sources are the other producers, and they produce on the operator's prio»
- ALB-0256 · role-core:workflow/factory-dispatch.md · «[HARD] **Promotion is the operator's act, in person or in advance.** After a /clear, the l»
- ALB-0258 · role-core:workflow/factory-dispatch.md · «[HARD] **The self-dispatch lane exception.** In a self-dispatch factory run, a lane sessio»
- ALB-0260 · role-core:workflow/factory-dispatch.md · «[HARD] **The leader may attach a finding; it may not act on one.** Analysis records a rela»
- ALB-0261 · role-core:workflow/factory-dispatch.md · «[HARD] **The pre-dispatch PR cross-check.** Before dispatching a card out of backlog, the »
- ALB-0262 · role-core:workflow/factory-dispatch.md · «[HARD] **The cross-check also asks whether a completed SPEC already covers the work.** A c»
- ALB-0263 · role-core:workflow/factory-dispatch.md · «[HARD] **The cross-check reports; it never vetoes.** Where the card carries an open pull r»
- ALB-0264 · role-core:workflow/factory-dispatch.md · «[HARD] **A milestone-bearing report under .moai/reports/ carries a ## Card Cross-Check sec»
- ALB-0266 · role-core:workflow/factory-dispatch.md · «[HARD] **Class A is admitted on checked evidence, not on an assertion.** Two of its three »
- ALB-0269 · role-core:workflow/factory-dispatch.md · «[HARD] Work is delegated through the queue on disk, not through messages. The queue file r»
- ALB-0270 · role-core:workflow/factory-dispatch.md · «A cross-session message is a nudge, never the delegation itself. **A lane can be asked to »
- ALB-0271 · role-core:workflow/factory-dispatch.md · «[HARD] A dispatch is written in the operator's conversation_language — the operator watche»
- ALB-0272 · always:.claude/rules/moai/workflow/factory-dispatch-core.md · «[HARD] A dispatch is a fixed-field address block, not prose. The fields:»
- ALB-0001 · companion:.claude/rules/moai/core/agent-common-protocol-detail.md · «»
- ALB-0272 · always:.claude/rules/moai/workflow/factory-dispatch-core.md · «card: <id>»
- ALB-0272 · always:.claude/rules/moai/workflow/factory-dispatch-core.md · «spec: <SPEC-ID>»
- ALB-0272 · always:.claude/rules/moai/workflow/factory-dispatch-core.md · «cmd: /moai run <SPEC-ID>»
- ALB-0272 · always:.claude/rules/moai/workflow/factory-dispatch-core.md · «wt: .moai/worktrees/<name>»
- ALB-0272 · always:.claude/rules/moai/workflow/factory-dispatch-core.md · «evidence: .moai/specs/<SPEC-ID>/progress.md»
- ALB-0272 · always:.claude/rules/moai/workflow/factory-dispatch-core.md · «lens: --security --deep»
- ALB-0001 · companion:.claude/rules/moai/core/agent-common-protocol-detail.md · «»
- ALB-0272 · always:.claude/rules/moai/workflow/factory-dispatch-core.md · «- card, cmd, wt, and evidence are always present. spec joins once a SPEC exists; a Class B»
- ALB-0272 · always:.claude/rules/moai/workflow/factory-dispatch-core.md · «- wt names the new card's worktree, never a previous card's tree. The tree keeps the card »
- ALB-0272 · always:.claude/rules/moai/workflow/factory-dispatch-core.md · «- **No explanatory prose.** Procedure, background, and justification live in the card text»
- ALB-0272 · always:.claude/rules/moai/workflow/factory-dispatch-core.md · «- **Ceiling: the block is at most 10 lines.** A dispatch that does not fit is trying to be»
- ALB-0272 · always:.claude/rules/moai/workflow/factory-dispatch-core.md · «- **[HARD] The send is read, not assumed.** A routing object on the result means an in-pro»
- ALB-0273 · role-core:workflow/factory-dispatch.md · «[HARD] A lane session tracks its card's execution on the session's task tools. At card int»
- ALB-0275 · role-core:workflow/factory-dispatch.md · «[HARD] Between run-exit verification and integration every lane runs a card-review stage: »
- ALB-0276 · always:.claude/rules/moai/workflow/factory-dispatch-core.md · «[HARD] A lane that cannot proceed records an explicit wait on disk — reason, whom, recheck»
- ALB-0277 · role-core:workflow/factory-dispatch.md · «[HARD] **The deputy is resident, not optional.** Before the batch's first lane dispatch, t»
- ALB-0278 · role-core:workflow/factory-dispatch.md · «[HARD] **A completion report reaches the leader as a RECOMMEND: summary, not as raw readin»
- ALB-0279 · role-core:workflow/factory-dispatch.md · «[HARD] **The deputy never holds a power of consequence.** Final PASS/FAIL verdicts, final »
- ALB-0280 · role-core:workflow/factory-dispatch.md · «[HARD] The leader advances a card on **evidence it read**, not on a lane's reply. Before m»
- ALB-0284 · always:.claude/rules/moai/workflow/factory-dispatch-core.md · «[HARD] A gh pr checks row naming CodeRabbit is not evidence that a review happened: the st»
- ALB-0284 · always:.claude/rules/moai/workflow/factory-dispatch-core.md · «1. The **combined** endpoint /commits/{sha}/status shows state == "success" **and** descri»
- ALB-0284 · always:.claude/rules/moai/workflow/factory-dispatch-core.md · «bash»
- ALB-0284 · always:.claude/rules/moai/workflow/factory-dispatch-core.md · «gh api "repos/$REPO/commits/$HEAD_SHA/status" \»
- ALB-0284 · always:.claude/rules/moai/workflow/factory-dispatch-core.md · «--jq '.statuses[] | select(.context == "CodeRabbit" and .state == "success") | .descriptio»
- ALB-0001 · companion:.claude/rules/moai/core/agent-common-protocol-detail.md · «»
- ALB-0284 · always:.claude/rules/moai/workflow/factory-dispatch-core.md · «2. A Merge Risk: line exists whose commit prefix matches the current headRefOid.»
- ALB-0286 · role-core:workflow/factory-dispatch.md · «[HARD] A lane session does not carry one card's context into the next card. When a card re»
- ALB-0287 · role-core:workflow/factory-dispatch.md · «[HARD] On a card transition inside a lane session, /clear happens exactly once, **after** »
- ALB-0290 · role-core:workflow/factory-dispatch.md · «[HARD] A card's work happens inside a worktree, and that worktree is **entered through the»
- ALB-0291 · role-core:workflow/factory-dispatch.md · «[HARD] A Codex factory agent must never call moai cc -w, EnterWorktree, or ExitWorktree; t»
- ALB-0292 · role-core:workflow/factory-dispatch.md · «[HARD] **moai worktree done closes L2 trees only** — project trees under .claude/worktrees»
- ALB-0293 · role-core:workflow/factory-dispatch.md · «[HARD] **The card's branch is unpushed, so its worktree is the work's only instance.** Dis»
- ALB-0294 · always:.claude/rules/moai/workflow/factory-dispatch-core.md · «[HARD] **A card session starts inside its worktree and stays there.** The standard is a la»
- ALB-0295 · role-core:workflow/factory-dispatch.md · «[HARD] **A new card starts in a new worktree** — a lane anchored in the previous card's tr»
- ALB-0296 · role-core:workflow/factory-dispatch.md · «[HARD] **Card worktree branches carry the WT- prefix and a descriptive slug — never the ca»
- ALB-0297 · role-core:workflow/factory-dispatch.md · «[HARD] **Inside a worktree session that <path> is the worktree's own absolute path** — the»
- ALB-0298 · role-core:workflow/factory-dispatch.md · «[HARD] **Lane-local verification is scoped to the card.** A lane runs the tests its own ch»
- ALB-0299 · role-core:workflow/factory-dispatch.md · «[HARD] **Never spawn background load.** Where a verification genuinely needs contention, t»
- ALB-0300 · role-core:workflow/factory-dispatch.md · «[HARD] Inside a worktree, an environment-scrubbed verification runs as one compound unset »
- ALB-0284 · always:.claude/rules/moai/workflow/factory-dispatch-core.md · «bash»
- ALB-0300 · role-core:workflow/factory-dispatch.md · «unset MOAI_KANBAN_ID MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED && go test ./...»
- ALB-0001 · companion:.claude/rules/moai/core/agent-common-protocol-detail.md · «»

**ⓑ role-core 행선지 32행 — 기계 검사**: after_text가 수리 후 주입 본문에 문자 그대로 존재 — **FAIL 0건(전부 통과)**. 검증: perl UTF-16 마커 추출 본문 + python norm 포함 검사(본 대조와 동일 구현). base 원문과 after_text 나란히:

| base 원문(478285ab4) | ledger after_text(현재 core) |
|---|---|
| «[HARD] **The leader is the queue's sole producer.** The operator asks; the leader turns the request into a card with moai gtd add "<description>" (moai gtd alon» | ALB-0254 «[HARD] Only the operator's requests enter the queue; the leader adds the card (moai gtd add) and never invents work.» |
| «[HARD] **Standing sources are the other producers, and they produce on the operator's prior authorization.** /moai project issues one card when it completes ([P» | ALB-0255 «[HARD] Standing sources (/moai project, codemaps-debt) also produce — one prefixed card per occasion, on prior operator authorization (gtd § Standing sources).» |
| «[HARD] **Promotion is the operator's act, in person or in advance.** After a /clear, the leader presents the queued cards through AskUserQuestion and the operat» | ALB-0256 «[HARD] Promotion is the operator's act: present queued cards after /clear; never pick, reorder, or silently promote outside --auto.» |
| «[HARD] **The self-dispatch lane exception.** In a self-dispatch factory run, a lane session may lease the next queued card — the one promotion a lane performs —» | ALB-0258 «[HARD] A self-dispatch lane leases its next card only via moai factory next; other queue mutations and contract signing stay forbidden to a lane.» |
| «[HARD] **The leader may attach a finding; it may not act on one.** Analysis records a relation between two cards (moai gtd relate); the record is evidence the o» | ALB-0260 «[HARD] The leader attaches findings (moai gtd relate), never acts on them; it may only refuse a duplicate of a queued or picked card.» |
| «[HARD] **The pre-dispatch PR cross-check.** Before dispatching a card out of backlog, the leader reads that card's pull-request and landed state (moai gtd pr <i» | ALB-0261 «[HARD] The pre-dispatch cross-check reads the card's PR and landed state and reports the same turn — an unchecked card is a gap.» |
| «[HARD] **The cross-check also asks whether a completed SPEC already covers the work.** A card id cannot answer "has someone else already done this" — where the » | ALB-0262 «[HARD] The cross-check also reports completed-SPEC coverage; neither read is conclusive — the final discriminator stays reproduction.» |
| «[HARD] **The cross-check reports; it never vetoes.** Where the card carries an open pull request or is already landed, the leader surfaces that and the operator» | ALB-0263 «[HARD] The cross-check reports, never vetoes — the operator confirms or withdraws; never withhold a picked card.» |
| «[HARD] **A milestone-bearing report under .moai/reports/ carries a ## Card Cross-Check section** — one table row per milestone, a card column holding the delive» | ALB-0264 «[HARD] Milestone reports carry a ## Card Cross-Check section — mappings verified against the queue, never remembered.» |
| «[HARD] **Class A is admitted on checked evidence, not on an assertion.** Two of its three properties are mechanically checked and cited: the diff is measured (g» | ALB-0266 «[HARD] Class A needs checked evidence: measured diff and CI green on the merging head, both cited.» |
| «[HARD] Work is delegated through the queue on disk, not through messages. The queue file resolves against the primary checkout from every linked worktree — one » | ALB-0269 «[HARD] Work is delegated through the queue on disk, not messages — one queue, visible from every worktree.» |
| «A cross-session message is a nudge, never the delegation itself. **A lane can be asked to report when it next goes idle** (SendMessage notify_when_idle), but [H» | ALB-0270 «A message is a nudge, never the delegation — an idle notice says when to look, not what the evidence says.» |
| «[HARD] A dispatch is written in the operator's conversation_language — the operator watches it scroll past, which makes it user-facing output rather than intern» | ALB-0271 «[HARD] Dispatches go in the operator's conversation_language; Agent() prompts stay English; identifiers verbatim.» |
| «[HARD] A lane session tracks its card's execution on the session's task tools. At card intake it registers the card's execution stages via TaskCreate before beg» | ALB-0273 «[HARD] A lane tracks its card's stages via TaskCreate/TaskUpdate; a list contradicting its report is a gap.» |
| «[HARD] Between run-exit verification and integration every lane runs a card-review stage: a card-scope self-review through codex_review (glm_review optional) wh» | ALB-0275 «[HARD] Every lane runs a card-review (codex_review → .moai/reports/<card-id>/card-review.md) before integration — advisory only.» |
| «[HARD] **The deputy is resident, not optional.** Before the batch's first lane dispatch, the factory leader session spawns exactly one UNNAMED background Agent(» | ALB-0277 «[HARD] The deputy is resident: one UNNAMED background Agent() running manager-lead, kept for the batch.» |
| «[HARD] **A completion report reaches the leader as a RECOMMEND: summary, not as raw reading batches** — turn occupancy moves; the leader's own evidence-read bef» | ALB-0278 «[HARD] Completion reports arrive as RECOMMEND: summaries — occupancy moves; the leader's evidence read does not.» |
| «[HARD] **The deputy never holds a power of consequence.** Final PASS/FAIL verdicts, final merge approval (LEAD-MERGE-APPROVED), operator gates, card issuance an» | ALB-0279 «[HARD] The deputy holds no power of consequence: verdicts, merge approval, operator gates, issuance, adjudication stay with the leader.» |
| «[HARD] The leader advances a card on **evidence it read**, not on a lane's reply. Before moving a card out of a working stage, the leader reads the card's progr» | ALB-0280 «[HARD] The leader advances a card on evidence it read; missing or stale evidence is a gap — the card stays put.» |
| «[HARD] A lane session does not carry one card's context into the next card. When a card reaches done and the leader has read its evidence, the leader **asks the» | ALB-0286 «[HARD] One card's context never crosses to the next: at done the leader asks the operator to /clear that session.» |
| «[HARD] On a card transition inside a lane session, /clear happens exactly once, **after** the session has moved into the next card's worktree — not before the m» | ALB-0287 «[HARD] /clear exactly once per transition, after moving into the next worktree — never before.» |
| «[HARD] A card's work happens inside a worktree, and that worktree is **entered through the launcher** — never created with a bare git worktree add. Launcher tab» | ALB-0290 «[HARD] Card work happens in a launcher-entered worktree — never a bare git worktree add (mechanics § Isolation).» |
| «[HARD] A Codex factory agent must never call moai cc -w, EnterWorktree, or ExitWorktree; those are Claude Code entry mechanisms.» | ALB-0291 «[HARD] A Codex factory agent never calls moai cc -w, EnterWorktree, or ExitWorktree.» |
| «[HARD] **moai worktree done closes L2 trees only** — project trees under .claude/worktrees/ and .moai/worktrees/ are L1 and absent from its registry. The full L» | ALB-0292 «[HARD] moai worktree done closes L2 trees only (worktree-integration § Terminology Glossary).» |
| «[HARD] **The card's branch is unpushed, so its worktree is the work's only instance.** Dispose of no worktree — L1 or L2 — until the leader has integrated the b» | ALB-0293 «[HARD] The card's branch is unpushed — the worktree is the work's only copy until the remote merge lands.» |
| «[HARD] **A new card starts in a new worktree** — a lane anchored in the previous card's tree MUST ExitWorktree before entering the next one; the fresh tree is c» | ALB-0295 «[HARD] A new card starts in a new worktree — exit the old tree first; never reuse; merge dependencies inside the new tree.» |
| «[HARD] **Card worktree branches carry the WT- prefix and a descriptive slug — never the card id** (rename in place with git branch -m WT-<slug>); the worktree d» | ALB-0296 «[HARD] Branches carry WT- + a descriptive slug, never the card id; traceability = dispatch card: + commit message + evidence path + PR title.» |
| «[HARD] **Inside a worktree session that <path> is the worktree's own absolute path** — the guard refuses -C ., relative paths, runtime-computed paths, and paths» | ALB-0297 «[HARD] In a worktree session <path> is that worktree's absolute path — relative, computed, and outside forms are refused.» |
| «[HARD] **Lane-local verification is scoped to the card.** A lane runs the tests its own change can affect, then pushes and lets CI run the full suite. (Incident» | ALB-0298 «[HARD] Verification is lane-local — only what the change can affect; CI runs the full suite (gates § incident record).» |
| «[HARD] **Never spawn background load.** Where a verification genuinely needs contention, the load must be cleanup-guaranteed — kills registered with the test fr» | ALB-0299 «[HARD] Never spawn background load — contention needs cleanup-guaranteed kills or a timeout wrapper.» |
| «[HARD] Inside a worktree, an environment-scrubbed verification runs as one compound unset … && <command> invocation — each Bash call is a fresh process, so a se» | ALB-0300 «[HARD] Env-scrubbed verification is one compound unset … && <command> call — a separate unset carries nothing.» |
| «unset MOAI_KANBAN_ID MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED && go test ./...» | ALB-0300 «[HARD] Env-scrubbed verification is one compound unset … && <command> call — a separate unset carries nothing.» |

**ⓒ 비주입 행선지 [HARD] 행 6건 — Codex 레인 불요 사유 초안(레인 작성, 수용/기각은 리더 Opus 감사자)**:

- **ALB-0272** (always:.claude/rules/moai/workflow/factory-dispatch-core.md) — base «[HARD] A dispatch is a fixed-field address block, not prose. The fields:» → 사유 초안: core의 복원 행(ALB-0272 대응 F4 삽입)이 필드 명세(`card:` `spec:` `cmd:` `wt:` `evidence:`; `lens:` optional)·비산문·최대 10행을 전달 — 블록의 예시 형태(``` 필드표)는 표기 중복. 미전달 잔여: send-read 판별 세부(라우팅 객체 형태)는 cross-session-messaging 규칙 계열 소유.
- **ALB-0001** (companion:.claude/rules/moai/core/agent-common-protocol-detail.md) — base «» → 사유 초안: (초안 미작성)
- **ALB-0276** (always:.claude/rules/moai/workflow/factory-dispatch-core.md) — base «[HARD] A lane that cannot proceed records an explicit wait on disk — reason, who» → 사유 초안: core의 B 행(F4 복원)이 동일 의무 전달 — 디스크 대기 기록·턴 종료·각성 시 워치독 선행·§6 사다리. 세부 절차는 본문 일반 텍스트에도 존재(파일을 읽으면 도달).
- **ALB-0284** (always:.claude/rules/moai/workflow/factory-dispatch-core.md) — base «[HARD] A gh pr checks row naming CodeRabbit is not evidence that a review happen» → 사유 초안: CodeRabbit PR 상태 판독은 카드 PR 경로의 착지 후 검증 절차로, 레인의 주입 의무 경로는 core 내 card-review 행(ALB-0275)이 전달하는 codex_review 독립 리뷰가 담당 — Codex 레인은 PR 상태 판독 없이도 독립 리뷰 의무가 이행됨.
- **ALB-0294** (always:.claude/rules/moai/workflow/factory-dispatch-core.md) — base «[HARD] **A card session starts inside its worktree and stays there.** The standa» → 사유 초안: core의 worktree one-liner 행들(런처 진입·Codex 금지·새 카드 새 트리·절대경로)이 동일 의무를 전달 — 세부 표는 중복.
- **ALB-0303** (always:.claude/rules/moai/workflow/factory-dispatch-core.md) — base «[HARD] A lane whose card has passed verification does not wait for the leader to» → 사유 초안: git-flow 변항은 본 리포의 현재 전략(github-flow)에서 미적용 — github-flow 경로는 core 내 card-review 행이 전달하는 factory complete PR 엣지가 담당. 본문 일반 텍스트에도 전문 존재(파일 읽기 도달).

- **ALB-0001 오매핑 주석**: base의 ``` 구조행 4건이 상세 컴패니언의 일반 블록과 토큰 겹쳐 매핑된 것 — 구조 프래그먼트로 소유 의무는 각 모행의 행이 담당(추적 대상 제외 주석).

**판정**: 레인 판정 없음 — ⓐ 0누락·ⓑ 0실패·ⓒ 7행 초안. 리더 Opus 읽기 전용 감사자가 ⓒ 항목별 수용/기각.

### ⓒ 판정 결과 — **보류 (HOLD)** (리더 통지, 2026-10-11)

- **판정자**: claude-opus-5-5 (리더 대행 Opus 읽기 전용 감사) — 판정 기록 `.moai/reports/t1617/c-adjudication-leader.md`
- **ⓒ 판정**: ALB-0001 수용 / **ALB-0272·0276·0284·0294·0303 기각** — 감사자가 codex adversarial(fail, rcpt-7309dff42885e7b41095ce34)과 주입 함수 실호출로 0272(Class B spec 생략 조건)·0276(디스크 대기 필수 3항목)의 주입 본문 누락을 확인, 레인 초안의 "core가 전달한다" 주장을 반증
- **ⓑ 의미 약화 행 (감사자 표시 10건)**: ALB-0255·0264·0266·0270·0273·0277·0279·0286·0293·0295
- **정지 규칙(15c5) 적용**: 마지막 수리 라운드 이후 신규 지적 → **추가 수리 없이 t1617 보류, 운영자 판단 상신**
- codex 델타 미실행 · 수리 없음 · HEAD `87f36a08d` 유지 (트리·push 변경 없음 — 리더 지시)
### (a) 복원 문안 표 — 15행 (운영자 결정 d-20261011T064637Z-ecdd, 2026-10-11)

Opus 감사자 승인 대기 — 승인 전 코드·규칙 파일 변경 없음. ⓑ 32행 전체 표는 위 「ⓐⓑⓒ 행선지 추적」 절에 기재되어 있음.

| 행 | base 원문 (478285ab4 / 원장 before_text) | 현재문 (core) | 복원 제안문 (레인 작성) |
|---|---|---|---|
| **ALB-0272** | «[HARD] A dispatch is a fixed-field address block, not prose. The fields: ⏎  ⏎ ``` ⏎ card: <id> ⏎ spec: <SPEC-ID> ⏎ cmd: /moai run <SPEC-ID> ⏎ wt: .moai/worktrees/<name> ⏎ evidence: .moai/specs/<SPEC-ID>/progress.md ⏎ lens: --security --deep ⏎ ``` ⏎  ⏎ - `card`, `cmd`, `wt`, and `evidence` are always present. `spec` joins once a SPEC exists; a Class B card carries none, and its `evidence` names wha» | «[HARD] A dispatch is a fixed-field block (`card:` `spec:` `cmd:` `wt:` `evidence:`; `lens:` optional) — never prose, ≤10 lines. [HARD] Read the send result: `routing` = lost (re-send `name [ref]`); a delivery notice = held or refused.» | «[HARD] A dispatch is a fixed-field address block, not prose — at most 10 lines: `card: <id>`, `spec: <SPEC-ID>` — `spec` joins once a SPEC exists; a Class B card carries none, and its `evidence` names whatever record the leader will read instead —, `cmd: /moai run <SPEC-ID>`, `wt: .moai/worktrees/<name>` (the new card's tree, never a previous card's), `evidence: .moai/specs/<SPEC-ID>/progress.md`, `lens:` only when the leader names review lenses for the card's sync gate. [HARD] The send is read,» |
| **ALB-0276** | «[HARD] A lane that cannot proceed records an explicit wait on disk — reason, whom, recheck point — and ends its turn; it never idles open-ended on a reply. On its next awaken the lane runs the stall watchdog first — invoke Skill("moai-lane-watchdog") and follow it — which measures progress, classifies the stall cause, and resolves the judgment through the decision ladder (doctrine: `.claude/rules/» | «[HARD] A stuck lane records a disk wait and ends its turn — never idles on a reply; on awaken: stall watchdog first (`moai-lane-watchdog`), then the `auto-semantics.md` §6 ladder; a 20-min recheck cron (`CronCreate`) stays armed to the completion rep» | «[HARD] A lane that cannot proceed records an explicit wait on disk — reason, whom, recheck point — and ends its turn; it never idles open-ended on a reply. On its next awaken it runs the stall watchdog first (`moai-lane-watchdog`), then the `auto-semantics.md` §6 ladder; a 20-min recheck cron (`CronCreate`) stays armed to the completion report.» |
| **ALB-0284** | «[HARD] A `gh pr checks` row naming CodeRabbit is not evidence that a review happened: the status is `success` **even when no review ran**, and the row prints `pass` byte-identically in both cases. A row counts only when BOTH hold: ⏎  ⏎ 1. The **combined** endpoint `/commits/{sha}/status` shows `state == "success"` **and** description `Review completed`: ⏎  ⏎     ```bash ⏎     gh api "repos/$REPO/c» | «[HARD] A CodeRabbit `gh pr checks` row is not evidence — counts only with combined-endpoint `success`+`Review completed` and a matching `Merge Risk:` line.» | «[HARD] A CodeRabbit `gh pr checks` row is not evidence — the status is `success` even when no review ran. A row counts only when BOTH hold: the combined endpoint `/commits/{sha}/status` shows `state == "success"` with description `Review completed`, and a `Merge Risk:` line exists whose commit prefix matches the current `headRefOid`.» |
| **ALB-0294** | «[HARD] A card session starts inside its worktree and stays there (`moai cc -w`). Moving a running card session into another card worktree is prohibited — exit-and-relaunch satisfies the new-card rule; the integration-tree merge entry keeps its own re-entry rule (return to the card worktree before continuing card work); an unavoidable move takes `/clear` once, after — never before.» (오기 정정: 최초 기재가 base 발췌였음 — 감사자 교정문 원문으로 정정, 2026-10-11)n from one card worktree into another card worktree is prohibited — that transition attaches a second skill listing on top of the session's first, and t» | «[HARD] A card session stays in its worktree (`moai cc -w`); moving it into another card worktree is prohibited — exit-and-relaunch satisfies the new-card rule; the integration-tree merge entry keeps its own re-entry rule; an unavoidable move takes `/» | «[HARD] A card session starts inside its worktree and stays there (`moai cc -w`). Moving a running card session into another card worktree is prohibited — exit-and-relaunch satisfies the new-card rule; the integration-tree merge entry keeps its own re-entry rule; an unavoidable move takes `/clear` once, after.» |
| **ALB-0303** | «[HARD] A lane whose card has passed verification does not wait for the leader to integrate it: the lane merges its own branch into the batch's release branch (`release/vX.Y.Z`) itself. The window is taken with `moai integration acquire --name <lane> --card <card-id>` BEFORE entering the release worktree, released after the completion report is sent; the lane enters the release worktree with `Enter» | «[HARD] A verified lane merges its own branch into the batch release branch (`release/vX.Y.Z`) — git-flow variant; github-flow default: `moai factory complete`'s PR edge (mechanics § standard landing). Where it applies: `moai integration acquire` firs» | «[HARD] A lane whose card has passed verification does not wait for the leader to integrate it: the lane merges its own branch into the batch's release branch (`release/vX.Y.Z`) itself. The window is taken with `moai integration acquire --name <lane> --card <card-id>` before entering the release worktree, released after the completion report is sent; the lane enters the release worktree with `EnterWorktree` (a cross-tree `git -C` is refused), merges `--no-ff`, re-reads `HEAD` before the commit an» |
| **ALB-0255** | «[HARD] **Standing sources are the other producers, and they produce on the operator's prior authorization.** `/moai project` issues one card when it completes (`[PROJECT] ` prefix); the codemaps-debt trigger in `moai integration release` issues one per debt period (`[GRAPH] ` prefix). The full conditions — one per occasion, derived not invented, marked, id reported, starting still a separate pick » | «[HARD] Standing sources (`/moai project`, codemaps-debt) also produce — one prefixed card per occasion, on prior operator authorization (gtd § Standing sources).» | «[HARD] Standing sources are the other producers, on the operator's prior authorization: `/moai project` issues one card when it completes (`[PROJECT] ` prefix); the codemaps-debt trigger issues one per debt period (`[GRAPH] ` prefix) — one per occasion, derived not invented, marked, id reported, starting still a separate pick (SSOT: gtd § Standing sources; the set is conditional, not closed). Nothing else produces.» |
| **ALB-0264** | «[HARD] **A milestone-bearing report under `.moai/reports/` carries a `## Card Cross-Check` section** — one table row per milestone, a `card` column holding the delivering card id or an explicit new-card marker. A mapping claim is verified against the queue (`moai gtd`), never remembered. Before the leader turns a report into card requests, the request message states the full comparison — `N milest» | «[HARD] Milestone reports carry a `## Card Cross-Check` section — mappings verified against the queue, never remembered.» | «[HARD] A milestone-bearing report carries a `## Card Cross-Check` section — one row per milestone, a `card` column with the delivering card id or an explicit new-card marker; mappings verified against the queue (`moai gtd`), never remembered; before turning a report into card requests the message states `N milestones → N cards`, naming every milestone with no card in the live queue.» |
| **ALB-0266** | «[HARD] **Class A is admitted on checked evidence, not on an assertion.** Two of its three properties are mechanically checked and cited: the diff is measured (`git diff --stat` against the base, showing the one file) and CI is green **on the head that will merge**. The third — no design judgement in it — is a judgement, stated in the dispatch where the operator can disagree with it. A card that ca» | «[HARD] Class A needs checked evidence: measured diff and CI green on the merging head, both cited.» | «[HARD] Class A is admitted on checked evidence, not an assertion: the diff is measured (`git diff --stat` against the base, showing the one file) and CI is green on the head that will merge — both cited; no-design-judgement is stated in the dispatch where the operator can disagree. A card that cannot cite both measurements is not Class A; the justification is never "it is faster".» |
| **ALB-0270** | «A cross-session message is a nudge, never the delegation itself. **A lane can be asked to report when it next goes idle** (`SendMessage` `notify_when_idle`), but [HARD] the notice is not the completion signal — it says *when to read the evidence*, nothing about what it says. **A nudge can be refused outright under fan-out**; read the send result, and a refusal costs the run nothing because the que» | «A message is a nudge, never the delegation — an idle notice says when to look, not what the evidence says.» | «A cross-session message is a nudge, never the delegation. A lane can be asked to report when it next goes idle (`notify_when_idle`), but [HARD] the notice is not the completion signal — it says when to read the evidence, nothing about what it says. A nudge can be refused under fan-out; read the send result — a refusal costs the run nothing because the queue already carries the delegation.» |
| **ALB-0273** | «[HARD] A lane session tracks its card's execution on the session's task tools. At card intake it registers the card's execution stages via `TaskCreate` before beginning the first stage; at every stage transition it keeps the list current via `TaskUpdate` so the list always shows the stage in flight; and it reports completion only while the list reflects the end state — or carries an explicit annot» | «[HARD] A lane tracks its card's stages via `TaskCreate`/`TaskUpdate`; a list contradicting its report is a gap.» | «[HARD] A lane tracks its card's stages on the session's task tools: registers the stages via `TaskCreate` at intake, keeps the list current via `TaskUpdate` at every transition, and reports completion only while the list reflects the end state — or carries an explicit annotation naming why not. A list contradicting the report is the same gap as a missing evidence file.» |
| **ALB-0277** | «[HARD] **The deputy is resident, not optional.** Before the batch's first lane dispatch, the factory leader session spawns exactly one UNNAMED background `Agent()` running manager-lead as its **coordination deputy** and keeps it for the batch. Its delegable/retained matrix lives in the agent itself (`manager-lead.md` § Deputy dispatch surface); this stub carries only the boundary that binds the qu» | «[HARD] The deputy is resident: one UNNAMED background `Agent()` running manager-lead, kept for the batch.» | «[HARD] The deputy is resident, not optional: before the batch's first lane dispatch the leader spawns exactly one UNNAMED background `Agent()` running manager-lead and keeps it for the batch.» |
| **ALB-0279** | «[HARD] The deputy never holds a power of consequence: final PASS/FAIL verdicts, final merge approval (`LEAD-MERGE-APPROVED`), operator gates, card issuance and `done`, CodeRabbit slot-wait adjudication, and cross-session dispute coordination stay with the leader; round-report measurement and drafting are the deputy's, the asserted figures the leader's — every figure names its measurer.» (오기 정정: 최초 기재가 base 발췌였음 — 감사자 교정문 원문으로 정정, 2026-10-11) — every» | «[HARD] The deputy holds no power of consequence: verdicts, merge approval, operator gates, issuance, adjudication stay with the leader.» | «[HARD] The deputy never holds a power of consequence: final PASS/FAIL verdicts, final merge approval, operator gates, card issuance and `done`, CodeRabbit slot-wait adjudication, and cross-session dispute coordination stay with the leader; asserted figures name their measurer.» |
| **ALB-0286** | «[HARD] A lane session does not carry one card's context into the next card. When a card reaches `done` and the leader has read its evidence, the leader **asks the operator to `/clear` that lane session** — `/clear` is a user-typed command and cannot be sent as an instruction. The leader's message states, in order: what closed (card, evidence read), which session to `/clear` (by name), and what hap» | «[HARD] One card's context never crosses to the next: at done the leader asks the operator to `/clear` that session.» | «[HARD] One card's context never crosses to the next: at done the leader asks the operator to `/clear` that lane session — a user-typed command that cannot be sent as an instruction; the message states what closed (card, evidence read), which session (by name), and what happens next.» |
| **ALB-0293** | «[HARD] **The card's branch is unpushed, so its worktree is the work's only instance.** Dispose of no worktree — L1 or L2 — until the leader has integrated the branch and the remote merge has landed.» | «[HARD] The card's branch is unpushed — the worktree is the work's only copy until the remote merge lands.» | «[HARD] The card's branch is unpushed — dispose of no worktree, L1 or L2, until the branch is integrated and the remote merge has landed.» |
| **ALB-0295** | «[HARD] **A new card starts in a new worktree** — a lane anchored in the previous card's tree MUST `ExitWorktree` before entering the next one; the fresh tree is created from the configured base, never reused. Where the new card depends on a prior card's unmerged code, merge that branch inside the new worktree.» | «[HARD] A new card starts in a new worktree — exit the old tree first; never reuse; merge dependencies inside the new tree.» | «[HARD] A new card starts in a new worktree — a lane anchored in the previous card's tree exits it first; the fresh tree is created from the configured base, never reused; a dependency on a prior card's unmerged code is merged inside the new worktree.» |

비고: 제안문은 base 의무 세부를 전부 보존하는 것을 원칙으로 작성(스텁 압축판에 있던 cron·send-read 등도 포함). ALB-0294·0303의 base 근거 설명 중 스텁이 계속 소유하는 부분(스킬 중복 방지 배경 등)은 생략. 승인 후 실측 천장 = 복원 반영 core 크기.
### 20행 [의무 요소 목록 | 감사자 교정문 | 최소안] 표 — 영실이 d-20261011T065458Z-88a7 (2026-10-11)

범위: 원 15행 + 실질 약화 5행(0254·0256·0260·0275·0280). 경미 4행(0261·0292·0296·0297)은 t1637 부채(리더 이관). APPROVE 8행(0284·0266·0270·0273·0277·0286·0293·0295)의 교정문 = 레인 (a) 제안문(감사자 승인). REVISE 7행 교정문 = 감사자 원문. 최소안의 의무 요소 수는 교정문과 동일 — 요소 손실 시 교정문 전문 사용(해당 행 명시).

**ALB-0272** (REVISE(감사자 교정문))
- 의무 요소 목록: 고정 필드 주소 블록·비산문·최대 10행 — 넘치는 payload는 카드로 · 필드: card·cmd·wt·evidence 항상 / spec은 SPEC 존재 시, Class B는 없고 evidence가 대체 기록 명명 / lens는 리더가 렌즈 지정 시만 / wt는 새 카드 트리 · 전송 결과를 읽음: routing=분실(재전송), 배차 통지=보류·거부(운영자에 표면화·재전송 금지), 그 외 큐잉 — 셋 중 하나도 레인의 Claude가 읽었다는 증거 아님
- 감사자 교정문: «[HARD] A dispatch is a fixed-field address block, not prose, at most 10 lines (a payload that does not fit goes into the card): `card: <id>`, `cmd: /moai run <SPEC-ID>`, `wt: .moai/worktrees/<name>` (the new card's tree, never a previous card's) and `evidence: .moai/specs/<SPEC-ID>/progress.md` always; `spec: <SPEC-ID>` once a SPEC exists — a Class B card carries none, and its `evidence` names whatever record the leader will read instead; `lens:` only when the leader names review lenses for the card's sync gate. [HARD] The send is read, not assumed: a `routing` object = lost (re-send to `name [ref]`); a `[Cross-session delivery notice]` = held or refused (surface it to the operator, do not re-send); anything else queued it — none of the three establishes that the lane's Claude read it.»
- 최소안: «[HARD] A dispatch is a fixed-field address block, not prose, at most 10 lines (overflow payload goes into the card): `card: <id>`, `cmd: /moai run <SPEC-ID>`, `wt: .moai/worktrees/<name>` (the new card's tree, never a previous card's), `evidence: .moai/specs/<SPEC-ID>/progress.md` always; `spec:` once a SPEC exists — a Class B card carries none, its `evidence` names the record the leader reads instead; `lens:` only when review lenses are named. [HARD] The send is read, not assumed: `routing` = lost (re-send `name [ref]`); a `[Cross-session delivery notice]` = held or refused (surface to the operator, do not re-send); anything else queued — none of the three proves the lane's Claude read it.»

**ALB-0276** (REVISE(감사자 교정문))
- 의무 요소 목록: 진행 불가 레인은 디스크 대기 기록 — reason·whom·recheck point — 턴 종료, 공한 대기 금지 · 각성 시 워치독 선행 → §6 사다리, 종단 = 레인 자신의 판단+결정 기록 — 리더에게 묻고 대기 금지 · 리더 도달은 게이트 경계(keep-set·카드 간 충돌)뿐, 결정판은 리더급 판정만 · 첫 단계 전 상시 20분 재확인 cron(CronCreate) 완료 보고까지 유지 · 모든 각성은 회신 전 디스크 증거를 먼저 읽음
- 감사자 교정문: «[HARD] A lane that cannot proceed records an explicit wait on disk — reason, whom, recheck point — and ends its turn; it never idles open-ended on a reply. On its next awaken it runs the stall watchdog first (`moai-lane-watchdog`), then the `auto-semantics.md` §6 ladder, whose terminal step is the lane's own judgment + decision record — never ask the leader and wait; the leader is reached only at a gate boundary (keep-set category or cross-card conflict), and its decision board carries leader-level rulings only. Before its first stage the lane arms a recurring 20-min recheck cron (`CronCreate`) kept until its completion report; every wake reads the disk evidence (progress record, reports, commits, the delegate's deliverable) before it replies.»
- 최소안: «[HARD] A lane that cannot proceed records an explicit wait on disk — reason, whom, recheck point — and ends its turn, never idling on a reply; on awaken it runs the stall watchdog first (`moai-lane-watchdog`), then the `auto-semantics.md` §6 ladder whose terminal step is the lane's own judgment + decision record — never ask the leader and wait (the leader is reached only at a gate boundary: keep-set or cross-card conflict; the board carries leader-level rulings only); a recurring 20-min recheck cron (`CronCreate`) is armed before the first stage and kept to the completion report, and every wake reads the disk evidence before it replies.»

**ALB-0294** (REVISE(감사자 교정문))
- 의무 요소 목록: 세션은 워크트리 안에서 시작해 머문다(moai cc -w) · 실행 중 카드 세션의 타 카드 워크트리 이동 금지 — exit-and-relaunch가 신규 카드 규칙 충족 · 통합 트리 병합 진입은 자체 재진입 규칙(카드 작업 계속 전 카드 워크트리로 복귀) · 불가피한 이동은 /clear 정확히 1회, 이동 뒤 — 이동 전 금지
- 감사자 교정문: «[HARD] **A card session starts inside its worktree and stays there.** The standard is a lane session launched inside the card's worktree through the launcher (`moai cc -w`); it starts with that tree's project skill set. Moving a running card session from one card worktree into another card worktree is prohibited — that transition attaches a second skill listing on top of the session's first, and that duplication is what this rule exists to prevent. Two composition rules bound the prohibition rather than leaving it implicit: (a) the "A new card starts in a new worktree" clause below composes with it; (b) the integration-tree merge entry keeps its own re-entry rule.»
- 최소안: «[HARD] A card session starts inside its worktree and stays there (`moai cc -w`); moving it into another card worktree is prohibited — exit-and-relaunch satisfies the new-card rule, the integration-tree merge entry keeps its own re-entry rule (return to the card worktree before continuing card work), and an unavoidable move takes `/clear` once, after — never before.»

**ALB-0303** (REVISE(감사자 교정문))
- 의무 요소 목록: git-flow 변항 한정 — 프로젝트 전략이 git-flow일 때만 · github-flow(배포 기본)는 factory complete PR 엣지(mechanics § 표준 착지) · 적용 시: 검증 통과 레인이 스스로 병합 — 리더 대기 아님 · 창: 입장 전 acquire, 완료 보고 후 release · EnterWorktree 진입(타 트리 git -C 거부) · --no-ff · 커밋·push 전 HEAD 재판독 · never force · 배치 PR은 리더 몫
- 감사자 교정문: «[HARD] git-flow variant — applies only where the project's git strategy names git-flow; under github-flow (the distributed default) the delivery is `moai factory complete`'s pull-request edge (`factory-dispatch-mechanics.md` § The lane's standard landing). Where it applies, a lane whose card has passed verification does not wait for the leader: it merges its own branch into the batch's release branch (`release/vX.Y.Z`) itself — window taken with `moai integration acquire --name <lane> --card <card-id>` before entering the release worktree and released after the completion report is sent; enter with `EnterWorktree` (a cross-tree `git -C` is refused), merge `--no-ff`, re-read `HEAD` before the commit and again before the push, push `release/vX.Y.Z` (never force), and leave the batch pull request with the leader.»
- 최소안: «[HARD] git-flow variant — applies only where the project's git strategy names git-flow; under github-flow (the distributed default) the delivery is `moai factory complete`'s pull-request edge (`factory-dispatch-mechanics.md` § The lane's standard landing). Where it applies, a verified lane merges its own branch into the batch's release branch (`release/vX.Y.Z`) itself — window taken with `moai integration acquire --name <lane> --card <card-id>` before entering the release worktree, released after the completion report; enter with `EnterWorktree` (cross-tree `git -C` refused), merge `--no-ff`, re-read `HEAD` before the commit and the push, push `release/vX.Y.Z` (never force), batch PR with the leader.»

**ALB-0255** (REVISE(감사자 교정문))
- 의무 요소 목록: 다른 생산자들 = 사전 운영자 승인 · `/moai project` → `[PROJECT] ` 완료 시 1장 · codemaps-debt → 기간당 1장 `[GRAPH] ` · 1회 1장·파생 비발명·표시·id 보고·시작은 별도 픽 · SSOT gtd § Standing sources, 조건부·폐쇄 아님 · 그 외 생산 없음 — 마일스톤·감사 지적·이슈는 운영자 승인 카드 요청으로만
- 감사자 교정문: «[HARD] Standing sources are the other producers, on the operator's prior authorization: `/moai project` issues one card when it completes (`[PROJECT] ` prefix); the codemaps-debt trigger issues one per debt period (`[GRAPH] ` prefix) — one per occasion, derived not invented, marked, id reported, starting still a separate pick (SSOT: gtd § Standing sources; the set is conditional, not closed). Nothing else produces; a report milestone, audit finding, or open issue reaches the queue only as a card request the operator approves.»
- 최소안: «[HARD] Standing sources are the other producers, on prior authorization: `/moai project` issues one card at completion (`[PROJECT] `), the codemaps-debt trigger one per debt period (`[GRAPH] `) — one per occasion, derived not invented, marked, id reported, starting still a separate pick (SSOT: gtd § Standing sources; conditional, not closed); nothing else produces — a report milestone, audit finding, or open issue reaches the queue only as a card request the operator approves.»

**ALB-0264** (REVISE(감사자 교정문))
- 의무 요소 목록: 범위 한정: `.moai/reports/` 아래 마일스톤 보고서 · ## Card Cross-Check 절 — 마일스톤당 1행 · card 열 = 카드 id 또는 명시적 신규 카드 표지 · 큐(`moai gtd`) 대조 검증 — 기억 아님 · 카드 요청 전 N milestones → N cards 명시, 큐에 카드 없는 마일스톤 전부 나열
- 감사자 교정문: «[HARD] A milestone-bearing report under `.moai/reports/` carries a `## Card Cross-Check` section — one table row per milestone, a `card` column holding the delivering card id or an explicit new-card marker. A mapping claim is verified against the queue (`moai gtd`), never remembered. Before the leader turns a report into card requests, the request message states the full comparison — `N milestones → N cards` — naming every milestone with no card in the live queue. Detail: `factory-dispatch-cards.md` § Report milestones ↔ queue cards.»
- 최소안: «[HARD] A milestone-bearing report under `.moai/reports/` carries a `## Card Cross-Check` section — one row per milestone, a `card` column with the delivering card id or an explicit new-card marker; mappings verified against the queue (`moai gtd`), never remembered; before turning a report into card requests the message states `N milestones → N cards`, naming every milestone with no card in the live queue.»

**ALB-0279** (REVISE(감사자 교정문))
- 의무 요소 목록: 최종 PASS/FAIL 판정 · 최종 병합 승인(LEAD-MERGE-APPROVED) · 운영자 게이트 · 카드 발행과 done · CodeRabbit 슬롯 대기 adjudication · 세션 간 분쟁 조정 — 전부 리더 · 측정·초안은 부사의 것, 단언 수치는 리더의 것 — 모든 수치는 측정자를 명명
- 감사자 교정문: «[HARD] **The deputy never holds a power of consequence.** Final PASS/FAIL verdicts, final merge approval (`LEAD-MERGE-APPROVED`), operator gates, card issuance and `done` (`moai gtd` mutations), CodeRabbit slot-wait adjudication, and cross-session dispute coordination stay with the leader session. Round-report measurement and drafting are the deputy's; the asserted figures are the leader's — every figure names its measurer. Nothing structural moves with the delegation: the queue stays the channel, completion stays evidence the leader read, and the verdict's home stays the lead.»
- 최소안: «[HARD] The deputy never holds a power of consequence: final PASS/FAIL verdicts, final merge approval (`LEAD-MERGE-APPROVED`), operator gates, card issuance and `done`, CodeRabbit slot-wait adjudication, and cross-session dispute coordination stay with the leader; round-report measurement and drafting are the deputy's, the asserted figures the leader's — every figure names its measurer.»

**ALB-0254** (REVISE(감사자 교정문) · 신규 5행)
- 의무 요소 목록: 리더의 자기-권한 큐 변경은 생산(production) 하나뿐 · `moai gtd add`는 요청을 카드로, `moai gtd` 단독은 목록 · 번역이지 발명이 아님 — 운영자가 요청하지 않은 것은 큐에 들지 않음
- 감사자 교정문: «[HARD] **The leader is the queue's sole producer.** The operator asks; the leader turns the request into a card with `moai gtd add "<description>"` (`moai gtd` alone lists the queue). Production is the one queue mutation the leader performs on its own authority — translation, not invention: nothing enters the queue the operator did not ask for.»
- 최소안: «[HARD] Production is the leader's only self-authority queue mutation — `moai gtd add` turns the operator's request into a card (`moai gtd` alone lists it): translation, not invention — nothing enters the queue the operator did not ask for.»

**ALB-0256** (REVISE(감사자 교정문) · 신규 5행)
- 의무 요소 목록: 승격은 운영자 행위 — 직접 또는 사전 · /clear 뒤 큐 제시 → 운영자 픽 → 카드 통째 배차 · --auto 외: 픽·재정렬·무음 승격 금지 · 빈 큐 = 보고 대상, 발명 유도 아님
- 감사자 교정문: «[HARD] **Promotion is the operator's act, in person or in advance.** After a `/clear`, the leader presents the queued cards through `AskUserQuestion` and the operator picks; only then does the leader route the card whole to a free lane. Outside an --auto authorization the leader never picks for the operator, never reorders by inferred priority, and never silently promotes a backlog item. An empty queue is a state to report, not a prompt to invent work. The one ranking the queue admits is the `--auto` cycle's own, bounded in the next paragraph.»
- 최소안: «[HARD] Promotion is the operator's act, in person or in advance — present the queue after `/clear`, the operator picks, then route the card whole; outside `--auto` never pick, reorder, or silently promote; an empty queue is a state to report, not a prompt to invent work.»

**ALB-0260** (REVISE(감사자 교정문) · 신규 5행)
- 의무 요소 목록: 지적 첨부은 행위 아님 — relate 기록 · 기록 = 운영자가 읽는 증거, 명령 아님 · 개별 금지: 접거나·재정렬하거나·삭제하거나·편집하지 않음 · 자기-권한 변경 하나: 정규화 텍스트 동일 카드 입장 거절
- 감사자 교정문: «[HARD] **The leader may attach a finding; it may not act on one.** Analysis records a relation between two cards (`moai gtd relate`); the record is evidence the operator reads, never a mandate — the leader never folds the related card away, never reorders the queue around it, and never drops or edits it. Analysis changes exactly one thing on its own authority: it refuses the admission of a card whose normalized text is identical to one already queued or picked.»
- 최소안: «[HARD] The leader attaches findings (`moai gtd relate`), never acts on them — the record is evidence the operator reads, never a mandate: never fold the related card away, never reorder the queue around it, never drop or edit it; the one self-authority change is refusing a card whose normalized text is identical to one already queued or picked.»

**ALB-0275** (REVISE(감사자 교정문) · 신규 5행)
- 의무 요소 목록: run-exit 검증과 통합 사이 모든 레인이 card-review 실행 · codex_review(glm_review 선택) → 결과를 card-review.md에 기록 · advisory만 — 리더 증거 판독·독립 감사의 대체 아님 · tree_scope: skip 시 리더가 같은 도구로 자기 내부 산출을 직접 검토
- 감사자 교정문: «[HARD] Between run-exit verification and integration every lane runs a `card-review` stage: a card-scope self-review through `codex_review` (`glm_review` optional) whose result is written to `.moai/reports/<card-id>/card-review.md` — advisory only, never a replacement for the leader's evidence read or an independent audit. With `tree_scope: skip` configured for the leader's checkout, the leader session carries no turn-end codex review gate and reviews its own internal output directly with the same tools. Stage order, evidence fields, and the re-review ceiling: `factory-dispatch-detail.md` § The card-review stage.»
- 최소안: «[HARD] Between run-exit verification and integration every lane runs a `card-review` (`codex_review`, `glm_review` optional) whose result is written to `.moai/reports/<card-id>/card-review.md` — advisory only, never a replacement for the leader's evidence read or an independent audit; with `tree_scope: skip` the leader reviews its own internal output directly with the same tools.»

**ALB-0280** (REVISE(감사자 교정문) · 신규 5행)
- 의무 요소 목록: 리더는 읽은 증거로 전진 — 레인 회신 아님 · 대상: progress.md + 위상이 선언한 검증 증거 경로 · 결여·판독 불가·stale = gap — 카드 정지, 리더 사유 보고 · 실패 신호 부재는 pass 아님
- 감사자 교정문: «[HARD] The leader advances a card on **evidence it read**, not on a lane's reply. Before moving a card out of a working stage, the leader reads the card's `progress.md` and the verification evidence path the phase declares; a missing, unreadable, or stale evidence file is a **gap** — the card stays put and the leader reports why. Absence of a failure signal is not a pass.»
- 최소안: «[HARD] The leader advances a card on evidence it read, not a lane's reply: the card's `progress.md` and the declared verification evidence path; missing, unreadable, or stale evidence is a gap — the card stays put and the leader reports why; absence of a failure signal is not a pass.»

**ALB-0284** (APPROVE(제안문=교정문))
- 의무 요소 목록: gh pr checks CodeRabbit 행은 증거 아님 — 리뷰 없어도 success · 결합 엔드포인트 success + `Review completed` 기술 · 현재 headRefOid와 일치하는 `Merge Risk:` 행 — 둘 다 충족 시에만 계수
- 감사자 교정문: «[HARD] A CodeRabbit `gh pr checks` row is not evidence — the status is `success` even when no review ran, and the row prints `pass` byte-identically in both cases. A row counts only when BOTH hold: the combined endpoint `/commits/{sha}/status` shows `state == "success"` **and** description `Review completed`, and a `Merge Risk:` line exists whose commit prefix matches the current `headRefOid`. Anything else is a gap, not a pass; `Review rate limited` means the review never started, and a card carrying it does not leave `sync`.»
- 최소안: «[HARD] A CodeRabbit `gh pr checks` row is not evidence — the status is `success` even when no review ran; a row counts only when BOTH hold: the combined endpoint `/commits/{sha}/status` shows `state == "success"` with description `Review completed`, and a `Merge Risk:` line whose commit prefix matches the current `headRefOid` exists. Anything else is a gap, not a pass; `Review rate limited` means the review never started.»

**ALB-0266** (APPROVE(제안문=교정문))
- 의무 요소 목록: Class A는 검증된 증거로 입장 — 단언 아님 · diff 실측(base 대비 --stat, 그 한 파일 표시) · 병합할 헤드에서 CI green — 둘 다 인용 · 설계 판단 없음은 디스패치에 명시 — 운영자가 반론 가능 · 둘 다 인용 못 하면 Class A 아님 · 근거가 "더 빠르다"인 경우 금지
- 감사자 교정문: «[HARD] **Class A is admitted on checked evidence, not on an assertion.** Two of its three properties are mechanically checked and cited: the diff is measured (`git diff --stat` against the base, showing the one file) and CI is green **on the head that will merge**. The third — no design judgement in it — is a judgement, stated in the dispatch where the operator can disagree with it. A card that cannot cite both measurements is not Class A. The justification is never "it is faster".»
- 최소안: «[HARD] Class A is admitted on checked evidence, not an assertion: the diff measured (`git diff --stat` against the base, showing the one file) and CI green on the head that will merge — both cited; no-design-judgement stated in the dispatch where the operator can disagree; a card that cannot cite both measurements is not Class A, and the justification is never "it is faster".»

**ALB-0270** (APPROVE(제안문=교정문))
- 의무 요소 목록: 세션 간 메시지는 nudge — 위임 그 자체 아님 · notify_when_idle로 다음 idle 보고 요청 가능 · [HARD] 통지는 완료 신호 아님 — 증거를 읽을 때를 알릴 뿐 내용은 말하지 않음 · fan-out 시 nudge 거부 가능 — 전송 결과를 읽고, 거부는 큐가 위임을 실은 뒤라 비용 0
- 감사자 교정문: «A cross-session message is a nudge, never the delegation itself. **A lane can be asked to report when it next goes idle** (`SendMessage` `notify_when_idle`), but [HARD] the notice is not the completion signal — it says *when to read the evidence*, nothing about what it says. **A nudge can be refused outright under fan-out**; read the send result, and a refusal costs the run nothing because the queue already carries the delegation (`cross-session-messaging.md` § An idle notice is a scheduling hint, § Configuration surface).»
- 최소안: «A cross-session message is a nudge, never the delegation; a lane can be asked to report when it next goes idle (`notify_when_idle`), but [HARD] the notice is not the completion signal — it says when to read the evidence, nothing about what it says; a nudge can be refused under fan-out — read the send result; a refusal costs the run nothing because the queue already carries the delegation.»

**ALB-0273** (APPROVE(제안문=교정문))
- 의무 요소 목록: TaskCreate로 시작 시 단계 등록 · 전이마다 TaskUpdate로 최신 유지 · 목록이 종료 상태를 반영할 때만 완료 보고 — 아니면 이유 명시 · 보고와 모순되는 목록 = 증거 파일 결여와 동일 갭
- 감사자 교정문: «[HARD] A lane session tracks its card's execution on the session's task tools. At card intake it registers the card's execution stages via `TaskCreate` before beginning the first stage; at every stage transition it keeps the list current via `TaskUpdate` so the list always shows the stage in flight; and it reports completion only while the list reflects the end state — or carries an explicit annotation naming why it does not. A task list that contradicts its completion report is the same gap as a missing evidence file (§ Completion is read, never trusted).»
- 최소안: «[HARD] A lane tracks its card's stages on the session's task tools: `TaskCreate` the stages at intake, `TaskUpdate` at every transition, report completion only while the list reflects the end state — or annotate why not; a list contradicting the report is the same gap as a missing evidence file.»

**ALB-0277** (APPROVE(제안문=교정문))
- 의무 요소 목록: 상주·선택이 아님 · 첫 레인 배차 전 · 정확히 하나의 UNNAMED 백그라운드 Agent(manager-lead) · 배치 내내 유지
- 감사자 교정문: «[HARD] **The deputy is resident, not optional.** Before the batch's first lane dispatch, the factory leader session spawns exactly one UNNAMED background `Agent()` running manager-lead as its **coordination deputy** and keeps it for the batch. Its delegable/retained matrix lives in the agent itself (`manager-lead.md` § Deputy dispatch surface); this stub carries only the boundary that binds the queue.»
- 최소안: «[HARD] The deputy is resident, not optional: before the batch's first lane dispatch the leader spawns exactly one UNNAMED background `Agent()` running manager-lead and keeps it for the batch.»

**ALB-0286** (APPROVE(제안문=교정문))
- 의무 요소 목록: 한 카드의 맥락은 다음 카드로 넘어가지 않음 · done에서 리더가 운영자에게 /clear 요청 — 사용자 타이핑 명령, 지시로 전송 불가 · 메시지 순서: 무엇이 닫혔는지 / 어느 세션인지(이름) / 다음은 무엇인지
- 감사자 교정문: «[HARD] A lane session does not carry one card's context into the next card. When a card reaches `done` and the leader has read its evidence, the leader **asks the operator to `/clear` that lane session** — `/clear` is a user-typed command and cannot be sent as an instruction. The leader's message states, in order: what closed (card, evidence read), which session to `/clear` (by name), and what happens next.»
- 최소안: «[HARD] One card's context never crosses to the next: at done the leader asks the operator to `/clear` that lane session — a user-typed command that cannot be sent as an instruction; the message states what closed (card, evidence read), which session (by name), and what happens next.»

**ALB-0293** (APPROVE(제안문=교정문))
- 의무 요소 목록: 카드 브랜치는 미푸시 — 워크트리가 작업의 유일한 사본 · L1·L2 어느 쪽도 폐기 금지 · 브랜치 통합 + 원격 병합 착지까지
- 감사자 교정문: «[HARD] **The card's branch is unpushed, so its worktree is the work's only instance.** Dispose of no worktree — L1 or L2 — until the leader has integrated the branch and the remote merge has landed.»
- 최소안: «[HARD] The card's branch is unpushed — dispose of no worktree, L1 or L2, until the branch is integrated and the remote merge has landed.»

**ALB-0295** (APPROVE(제안문=교정문))
- 의무 요소 목록: 새 카드는 새 워크트리 · 이전 트리에 앵커된 레인은 먼저 ExitWorktree · 새 트리는 설정된 base에서 생성 — 재사용 금지 · 선행 카드 미병합 코드 의존은 새 워크트리 안에서 병합
- 감사자 교정문: «[HARD] **A new card starts in a new worktree** — a lane anchored in the previous card's tree MUST `ExitWorktree` before entering the next one; the fresh tree is created from the configured base, never reused. Where the new card depends on a prior card's unmerged code, merge that branch inside the new worktree.»
- 최소안: «[HARD] A new card starts in a new worktree — a lane anchored in the previous card's tree exits it first; the fresh tree is created from the configured base, never reused; a dependency on a prior card's unmerged code is merged inside the new worktree.»


### (e) 선행 실측 — 라이브 주입 렌더링과 고정분 분해 (리더 지시, 2026-10-11)

- **마커 결함 발견·수복**: ALB-0294 신규 영역이 기존 [22] 영역 내부에 중첩 삽입되어 buildRoleCore 서열 검증 실패 → 실호출이 REQ-ALB-009 후퇴(코어 제거, 3,647자)로 관측됨. 양 트리에서 0294 영역을 바깥 영역 뒤로 재배치 — nested-open 0으로 수복.
- **수복 후 실호출 원문** (트리 빌드 bin/moai, 카드 트리 cwd, synthetic session_id):
  - additionalContext 전체 **12,143자** — 온전한 넘침 경로(코어 intact 발송, 런타임이 10,000에서 절단·파일 저장)
  - role-block 시작 지점: **3,829** (synthetic 호출의 생산자 고정분 — 실제 레인 세션은 4,797 수준)
  - systemMessage: 팩토리 합류·스폰 권한 고지 (경고 아닌 정보성)
- **합성 13,027자 분해**: role-core 밖 고정분 = 생산자 4,797(실제 레인 기준) + 조립 오버헤드 257(헤더·포인터·구분) = **5,054** · role-core 7,973. 고정분만으로도 캡의 절반.
- **넘침 시 앞 2,000자 미리보기 원문**: /tmp/t1617-preview2000.txt — **role-core 내용 0자**: 넘침 안내문(NOTE) + 세션 속성 + GLM 라우팅 고지만 보임(코어 시작 3,829 > 2,000). 즉 넘침 세션의 미리보기로는 어떤 [HARD] 의무도 전달되지 않고, 전문은 넘침 파일을 읽어야 함.
- 라이브 주입 확인 Gap 소멸 — 실호출 원문 확보(본 절 + /tmp/t1617-hook-live2.json).

### 마커 중첩 회귀 시험 존재 확인 (리더 지시, 2026-10-11)

- **생산 코드 검증은 존재**: role_rules.go:221-240 깊이 검사기 — "a start marker before its enclosing region is closed" 거부. 본 카드의 실호출 후퇴(3,647자)가 바로 이 검증이 설계대로 동작한 결과(실패 안전).
- **회귀 시험은 없음**: role_rules_test.go malformed 계열(unclosed_region·end_before_start·unmarked·absent·unreadable·empty)에 **nested-start 형태(start가 열린 영역 안에 다시 삽입)의 subtest가 없음** — 본 카드 스크립트 결함이 만든 정확한 형태. 처분 뒤 반영 항목: fixture에 중첩 start 삽입 → assertFailVisible("before its enclosing region is closed") 1 subtest 추가.
- **end-to-end core 포함 시험도 없음**: 기존 계열은 주입 결정 구조체와 크기 게이트 사다리만 검사 — 실제 조립 출력(훅 additionalContext)에 코어가 온전히 포함되는지 단언하는 end-to-end 시험은 없음(본 카드의 bin/moai 실호출이 첫 관측). 처분 뒤 반영 항목 2: 실호출(또는 조립 함수 직호출) 출력에 두 의무 문구+코어 헤더 존재 단언.

### 고정분 최대 경우 실측 — 리더 역할 시뮬레이션 이상 관측 (리더 지시, 2026-10-11)

- **lane 역할(GLM, synthetic)**: 마커 수복 후 실호출 — additionalContext 12,143자(온전한 넘침 경로), role-block 시작 3,829(synthetic 생산자; 실제 lane-15 세션은 4,797). 최대 경우 추정: 4,797 + 257 + 7,973 = **13,027**.
- **리더 역할 시뮬레이션**: `env -u MOAI_FACTORY_WORKER MOAI_FACTORY_WORKERS=1 bin/moai hook session-start` → additionalContext **0자** (세션 속성 등 생산자도 0 — 훅 exit 0, stderr는 무관한 다중 세션 고지). 원인 미규명 — 리더 전용 생산자가 실제 리더 세션 상태(소켓·부트스트랩)를 필요로 하거나, 시뮬레이션 env로는 리더 역할 경로가 완전 동작하지 않을 수 있음.
- **판단 회신**: 리더 역할의 고정분 최대치는 본 시뮬레이션으로 측정 불가 — 실제 리더 세션의 SessionStart 추가Context에서 직접 재는 것이 정확(리더 소관). 레인 측 정확치: 실제 lane-15 세션 고정분 4,797(델타1 감사자 측정) + 본 트리 코어 7,973 = 합성 13,027.

### 리더 세션 실측 수령 — 고정분 최대 확정 (리더 제공, 2026-10-11)

- **리더 세션 실측** (rc.29 설치 빌드, 10-11, SessionStart 저장 additionalContext 원문): 전체 **19,368자**(UTF-16 동일) · role 영역 시작 오프셋 **999**. 앞 999자 구성: 넘침 안내문(0~454 — 넘침이 일어나서 붙은 것, 평시 미포함) · 세션 속성(455~) · guard liveness(701~) · memory-budget(827~). GLM 고지 없음(Opus 세션).
- **고정분 최대 확정**: 넘침 안내문 제외 리더 고정분 약 **544** ≪ 레인 5,054 → **최대 = 레인 역할(GLM) 5,054**. 예산 분모: **10,000 − 5,054 = 4,946**(role-core 상한).
- **Gap 2건 → t1637 메모**: (1) rc.29 빌드 기준이라 본 카드의 조립 오버헤드 257이 리더 쪽에서 동일한지 미측정. (2) 리더 역할 시뮬레이션 0자 원인 미규명.
- **종단 시험 요건 추가**: 리더 역할 조립 출력도 포함해 ≤10,000 단언.

### en 생산자 문구 실측 (리더 지시 읽기 전용 정리, 2026-10-11)

파일: internal/hook/session_start_factory_i18n.go (en 블록, langEnglish). 문자열 내용 UTF-8 바이트 근사(Go 연결·형식 토큰 제외):

| 생산자 | en 시작행 | 문자열 바이트(근사) | 역할 |
|---|---|---|---|
| laneRecheckRule | :162 | ~1,221 | 재확인 cron·각성 절차 안내 |
| laneNextCardRule | :133 | ~1,022 | 다음 카드 임대·착지 절차 안내 |
| laneSpawnAuthority | :170 | ~815 | 상시 스폰 권한 안내 |
| 합계 | | **~3,058** | en 레인 가변 권고문의 축약 대상 3문구 |

그 외 en 필드: laneOwnedCardRule 770 · autoModeGuide 569×2 · laneManualDispatchRule 516 · leaderStagger 403 등. Opus overlay 측정(en lane=1,122)과 본 근사(815)의 차이는 형식 토큰·%s 치환 반영 차이로 추정 — 축약 시뮬레이션은 실호출로 재측정 필요.

레인 en 합성 11,621 실측의 절감 여지가 이 3문구에 집중됨(전체 초과 1,621 중 3문구 합 3,058). (가) 채택 시 각 문구의 필수 요소(임대 동사·착지 절차·재확인 3항목)는 유지하면서 안내 서술을 압축하는 방향이 될 것.

### 일괄 갱신 항목 기록 (리더 지시, 2026-10-11) — 처분 반영 커밋 때 함께

1. **F4 시험 갱신**: TestInjectedCoreCarriesCodexReachableObligations — 현재 ALB-0272(디스패치 형식, OUT 이동)의 "fixed-field" 문구를 표지 안에서 찾아 FAIL. 갱신: OUT 이동 의무는 읽기 지시 문장 존재 + 본문 파일 요소 단언으로 검사 대상 변경(시험 삭제 아님). ALB-0276(디스크 대기)은 IN 유지 확인 — 원장 role-core·양 트리 표지 안 \`disk wait\` 문구 존재 재확인(본 절 기록 시점, 2026-10-11).
2. **예산 상수**: TestRoleInjectionAssemblyBudget — budgetCoreCeiling 4,073 / budgetCompositeLimit 9,127 → manager-spec이 IN/OUT 반영 실측에 맞춰 재설정.
3. **(C) 최종 조정 대기**: en 3문구 축약 + 추가 en 생산자 축약(ko 불변) + IN 하위 우선순위 OUT 이동(0303·0260 끝까지 IN, IN 집합 양 로캘 공통) — 리더 Opus가 최종 문안·IN/OUT 목록·4칸 실측 일괄 산출 중. 도착 시 글자 그대로 반영.

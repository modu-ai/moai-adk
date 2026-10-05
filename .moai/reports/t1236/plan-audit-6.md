auditor-model: claude-opus-5-5[1m]

# SPEC Review Report: SPEC-AUTONOMY-GATE-REWIRE-001
Iteration: 6 (최종 — 리드가 허용한 마지막 회차. 범위: `git diff f40824c82 a02e7c02b -- <SPEC dir>` 의 D47~D55 처분과 그 연쇄 불일치만)
Verdict: FAIL — D47 부분 미해결(리드 규칙 「D47/D48/D49 중 하나라도 열려 있으면 FAIL」). 최종 회차이므로 사용자 개입을 권고한다
Overall Score: 0.868 (Tier L 문턱 0.85 — 점수만으로는 넘지만 판정은 위 규칙이 결정)

감사 대상: 커밋 `a02e7c02b`(v0.3.5, 부모 `f40824c82`), 브랜치 `WT-contract-gate-rewire`, 워크트리 `git status --short` 출력 없음. Reasoning context ignored per M1 Context Isolation — 호출 측 문구는 범위 지정으로만 쓰고 좌표·출력은 전부 다시 쟀다. 교차 모델 호출은 하지 않았다(iter-5 와 같이 `audit_model` 설정 없음, 이 감사자 단독 판정). 인용 정책: `.claude/rules/moai/development/verification-completeness.md` §2(변이 탐침·불가능 방향 기준), §2.1(verbatim = 원시 바이트).

## 먼저 밝혀 둘 것 — iter-5 의 전제 오류

iter-5 의 D47 은 「`[HARD]` 표지를 단 블록 하나가 `unregistered_count` 를 0→1 로 올려도 AC-GR-003 은 통과한다」고 썼다. 이번에 그 전제를 실행으로 확인했더니 **틀렸다**: `internal/constitution.Validate` 에는 `ZONE_UNREGISTERED`(그리고 `ANCHOR_NOT_FOUND`) 항목을 만드는 코드 경로가 없다. 상수(validator.go:22·28)와 재집계 분기(validator.go:323)만 있고, 항목을 append 하는 곳은 없다(`grep -rn 'SentinelZoneUnregistered\|ANCHOR_NOT_FOUND' internal cmd pkg --include='*.go' | grep -v _test.go` → 선언·주석·재집계 case 뿐). 따라서 `unregistered_count` 는 구조적으로 늘 0 이다. manager-spec 은 iter-5 의 수리 지시를 문자 그대로 따랐고, 그 결과 아래 D56 이 생겼다. 원인의 절반은 이 감사자의 미검증 전제다(verification-claim-integrity.md §1 — 텍스트 추론을 결함으로 단정했다).

## Must-Pass Results

- [PASS] MP-1 REQ 번호: 델타는 REQ 를 추가·삭제하지 않았다(spec.md 변경은 frontmatter `version`, HISTORY 한 줄, REQ-GR-022 본문뿐). plan.md §L 「요구사항 25개·AC 25개 그대로」. 요구 층 번호 연속성은 iter-5 에서 확인한 그대로다.
- [PASS] MP-2 GEARS (요구 층): 델타가 고친 요구는 REQ-GR-022 하나이고 「**When** `moai contract revoke …` 가 실행될 때, the 명령 shall …」 구조를 유지한다(spec.md:130). AC 층 Given-When-Then 은 이 기준으로 채점하지 않았다.
- [PASS] MP-3 frontmatter: `version: "0.3.5"`(따옴표) 외 변경 없음. 12필드 유지.
- [N/A] MP-4 언어 중립: 단일 저장소 Go 코드 SPEC.
- [PASS] MP-5 D7: 델타는 새 SPEC 참조를 넣지 않았다. iter-5 에서 참조 SPEC 다섯 개 모두 retired/superseded/archived 아님을 확인.
- [PASS] MP-6 D8: 델타에 `syscall` 없음.
- [PASS] MP-7 clarification gate: 델타가 `[NEEDS CLARIFICATION` 표지를 넣지 않았다(plan.md 변경분은 §A·§C·M0·§K·§L).

## 결함별 처분 (D47~D55)

| 결함 | 처분 | 근거(이번 실행에서 잰 값) |
|---|---|---|
| **D47** | **부분 미해결 (OPEN)** | 닫힌 부분: (1) 비교가 DRIFT 한 범주에서 비-OK 전 sentinel 의 `(sentinel, id)` 부분집합 + 세 개수 이하로 넓어짐(acceptance.md:37). 열거한 sentinel 9개는 validator.go:15~40 의 상수와 정확히 일치. (2) `t.Setenv("MOAI_CONSTITUTION_SKIP_VALIDATE", "")` + 두 결과 `Skipped == false` 전제 단언(acceptance.md:45) — `Validate` 는 값이 정확히 `"1"` 일 때만 건너뛰므로(validator.go:186) 빈 값은 유효. 건너뛰기 변이의 「∅ ⊆ ∅」 공허 통과는 이제 전제 단언에서 FAIL 한다. (3) BASE DRIFT id = EV-6 9개 전제 단언(acceptance.md:46). (4) 판정 명령은 `go test` 하나, `moai constitution validate` 는 참고용으로 명시(acceptance.md:50). **열린 부분**: 변이 탐침 「always-loaded 파일에 미등록 `[HARD]` 줄 삽입」은 **여전히 AC-GR-003 을 통과한다** — 실측으로 확인(아래 D56). 그리고 수리가 추가한 반증 하위 테스트 (ii) 는 관측이 불가능한 RED 를 요구한다 |
| **D48** | **해결 (CLOSED)** | AC-GR-016 픽스처 (17)(acceptance.md:209): (3) 과 같되 저장소 revoke 사건 없이 현재 seal 지문의 `kind: revoke` 기록만 → Then 은 `(8)·(17) exit 1 revoked`(acceptance.md:211). (3) 자체는 exit 0 이므로 판독기를 부르지 않는 kickoff-check 변이는 (17) 에서 exit 0 → FAIL, 탐침 성립. AC-GR-018 (e) revoke 변형(acceptance.md:233): `status: resolved` 이므로 A2 `NeedsDecision` 은 거짓(record.go:180 조건 `Status == open`), A3 판독기만 차단 → `NeedsDecision` 만 보는 decide 변이는 Jev 를 생성해 FAIL(acceptance.md:235). 하위 테스트 수 17(=픽스처 17)·13(=깬 11 + 문턱 이상 PWD 1 + 성립 1, 종전 10+1+1=12 에서 +1)·7(r0~r6) 모두 픽스처 수와 맞는다. design §12 kickoff-check·decide 행(design.md:312·313) 동기화됨 |
| **D49** | **해결 (CLOSED)** | design §2 24행(design.md:59)에 `internal/cli/contract.go` 추가, 편집 범위를 「516행 하위 명령 등록, 389~414행 옵션·이음매 조립」 두 곳으로 한정. 재측정: `sed -n 385,416p` → 389 `opts := sign.Options{`, 408 `seams := sign.Seams{`, 414 `res, err := sign.Sign(opts, seams)`; `sed -n 510,522p` → 516 `cmd.AddCommand(verifyCmd, showCmd, signCmd)`. `git diff --quiet 7fe658815 HEAD -- internal/cli/contract.go internal/contract/ internal/constitution/` exit 0 → BASE 좌표가 HEAD 에서도 유효. REQ 열이 `007·011·022·012·025` 로 갱신돼 M7(doctrine 플래그 전달)과 연결됨 |
| **D50** | **해결 (CLOSED)** | acceptance.md:370~441(EV-6 원문 72행)을 파일로 잘라 트리 빌드 바이너리의 `constitution validate --format json` stdout(72행)과 `diff` → 출력 없음, exit 0. `<`·`>` 이스케이프가 원시 바이트대로 복원됐고 주석(acceptance.md:444)이 사실과 일치 |
| **D51** | **해결, 잔여 1건** | plan.md:10·36·75, acceptance.md:494 이 A2 0.4.3 기준으로 바뀜. **잔여**: plan.md:144 「`SPEC-AUTONOMY-ESCALATION-001` (A2 — 개정본 대기)」 — iter-5 가 열거하지 않은 같은 부류의 낡은 표지(D57, minor). plan.md:160 은 처분 이력 표라 대상 아님, research.md 는 머리에서 「§10.6 외 v0.3.2 측정」 선언이라 대상 아님 |
| **D52** | **해결 (CLOSED)** | design.md:58 이 `doc.go` 범위를 「47행 시그니처 목록과 135~141행 `# Kickoff receipt` 문단」으로 넓힘. 재측정: doc.go:47 `ReceiptOutcome(r *KickoffReceipt) (refusal string, ok bool)`, 135 `// # Kickoff receipt`, 139~140 「ReceiptOutcome then applies the interim A1 rule (effective decider llm+jev → receipt_requires_human)」. plan.md:212 §K B4 행도 동기화 |
| **D53** | **반영 (CLOSED)** | 추적표 REQ-GR-009 행 = `AC-GR-018, AC-GR-023`(acceptance.md:464) |
| **D54** | **반영 (CLOSED)** | REQ-GR-022(spec.md:130) 「디렉터리가 없으면 기록 0건(해제, 오류 아님 — A2 `NeedsDecision` 의 `filepath.Glob` 과 같은 의미)」, design §10 판독기 표에 「디렉터리 부재」 행 분리(design.md:265~266), AC-GR-023 (r0) 추가(acceptance.md:294·296). 세 층이 같은 의미 |
| **D55** | **반영, 잔여 1건** | 트리 빌드 바이너리 `moai spec lint SPEC-AUTONOMY-GATE-REWIRE-001` → exit 0, 「No findings」(iter-5 는 27 warning `VacuousTestAssertion`). `grep -n "\-run '[^^]"` 세 파일 0건. AC-GR-009 는 정확한 이름 넷으로 전개됐고 `go test -list '^(…)$' ./internal/template/` 이 기존 3개를 고른다(이번 실행 관측). **잔여**: acceptance.md:124 「`go test -list '<위 정규식 그대로>'` … 트리 `1b071a573`」 과 §B.1 009 행(acceptance.md:351)은 옛 비고정 정규식으로 잰 기록인데, 124행은 바뀐 정규식을 「그대로」 썼다고 귀속한다(D58, minor). 선택 결과는 새 정규식에서도 같다는 것을 이번에 관측했으므로 결과는 참이고 귀속만 틀렸다 |

## 변이 탐침 기록 (리드 지정 세 건)

1. **판독기를 우회하는 kickoff-check / decide** — 잡힌다(D48 CLOSED). kickoff-check 는 (17) 에서 exit 0 을 내어 기대 `revoked` 와 어긋나고, decide 는 (e) revoke 변형에서 Jev 를 생성해 `outcome: human`·Jev 0회 기대와 어긋난다. 문언 기준 추론이며 테스트는 아직 없다(Gaps).
2. **`MOAI_CONSTITUTION_SKIP_VALIDATE=1`** — 잡힌다. 테스트가 변수를 스스로 지우고 `Skipped == false` 를 단언하며, 외부에서 `=1` 을 넘겨도 `t.Setenv` 가 덮는다.
3. **미등록 `[HARD]` 줄 삽입** — **여전히 통과한다.** 실측:
   - `git archive -o <scratch>/head.tar HEAD` → 사본 추출, `.claude/rules/moai/core/moai-constitution.md` 끝에 `[ZONE:Evolvable] [HARD] Probe rule: …` 한 줄 추가, 사본 안에서 트리 빌드 바이너리로 `constitution validate --format json` → exit 1, `{"drift_count":9,"missing_count":0,"unregistered_count":0}`, id 9개 동일. 워크트리 출력과 `cmp` → exit 0(바이트 동일).
   - 양성 대조: 같은 사본에서 등록 조항 「Execute all independent tool calls in parallel when no dependencies exist.」 문장을 지우고 다시 실행 → `{"drift_count":10,"unregistered_count":0,"n":10}`. 사본을 실제로 읽고 있음을 확인했으므로 위의 「변화 없음」은 미측정이 아니라 부재다.

## Category Scores

| 차원 | 점수 | 밴드 | 근거 |
|---|---|---|---|
| Clarity | 0.88 | 0.75~1.0 | D54 로 판독기의 부재/오류 구분이 세 층에서 같아짐(spec.md:130, design.md:265~266, acceptance.md:296). AC-GR-003 판정 명령이 하나로 정해짐(acceptance.md:50). 감점: plan.md:144 낡은 A2 표지(D57) |
| Completeness | 0.92 | 0.75~1.0 | 허용 목록이 실제 편집 지점을 모두 담음(design.md:58·59 — `contract.go`·`doc.go` 135~141). §L 처분표(plan.md:217~231) 완비 |
| Testability | 0.75 | 0.75 | AC-GR-003 하위 테스트 (ii) 가 관측 불가능한 RED 를 요구(D56, verification-completeness.md §2 「impossible」 방향). 미등록 `[HARD]` 변이가 여전히 통과. 나머지 탐침(D48, 건너뛰기 변수)은 잡힘 |
| Traceability | 0.95 | 0.75~1.0 | REQ-GR-009 → AC-GR-018·023(acceptance.md:464). AC-GR-023 이 소비 경로를 AC-GR-016 (17)·018 (e) 로 명시 연결(acceptance.md:302). 감점: acceptance.md:124 증거 귀속 불일치(D58) |

집계: 조화 평균 4 / (1/0.88 + 1/0.92 + 1/0.75 + 1/0.95) = 0.868. iter-5 0.841 대비 상승 — STOP(점수 하락) 신호 없음.

## Defects Found (structured defect-list)

D56. D47-RESIDUE-IMPOSSIBLE-RED — acceptance.md:48(AC-GR-003 4.(ii))·design.md:114 — 반증 하위 테스트 (ii) 「always-loaded 파일에 미등록 `[HARD]` 줄 하나를 넣은 사본 → `ZONE_UNREGISTERED` 1건(`unregistered_count` 0 → 1) → 비교 FAIL 관측」은 BASE 의 `Validate` 로는 영원히 관측되지 않는다 — `Validate` 에 `ZONE_UNREGISTERED` 를 만드는 경로가 없다(validator.go 전체, 위 실측). run 이 이 하위 테스트를 정직하게 쓰면 RED 가 나오지 않아 AC 를 채울 수 없고, 채우려면 가짜 `ValidationResult` 를 손으로 만들어 비교 함수에 넣게 되는데 그것은 문언의 「사본 → 1건」을 어긴다. 동시에 실제 위험(블록 안의 미등록 `[HARD]` — REQ-GR-024 가 `zone-registry.md` 편집을 금하므로 그런 줄은 영구 미등록이다)은 어떤 AC 도 막지 않는다. `unregistered_count ≤ BASE` 조건도 구조적으로 0 ≤ 0 이라 무력하다. 원인은 iter-5 감사의 미검증 전제. — Severity: major — Class: blocking — Required fix: (a) 4.(ii) 를 `Validate` 에 기대지 않는 테스트 자체 검사로 바꾼다 — 허용 목록의 always-loaded 대상 파일 세 개(와 템플릿 사본)에서 `[HARD]` 를 포함한 줄의 집합을 BASE 와 현재 트리에서 각각 모아 「현재 ⊆ BASE」를 요구하고, 반증 하위 테스트는 사본에 `[HARD]` 줄 하나를 넣어 이 비교가 FAIL 함을 관측한다. 또는 동등하게 `TestContractModeBlocksWellFormed`(AC-GR-008) 규칙에 「블록 안에 `[HARD]`·`[ZONE:` 표지 없음」을 넣고 반증 픽스처를 둔다. (b) AC-GR-003 본문에 「BASE 의 `Validate` 는 `ZONE_UNREGISTERED`·`ANCHOR_NOT_FOUND` 를 내지 않는다 — `unregistered_count` 조건은 향후 구현 대비이며 현재 판정력이 없다」를 한 문장으로 적는다. (c) design.md:114 를 같게 고친다. AC 수는 그대로라 AC 스냅숏 재생성 의무 없음.

D57. STALE-A2-REF-144 — plan.md:144 — 「`SPEC-AUTONOMY-ESCALATION-001` (A2 — 개정본 대기)」가 남아 plan.md:10·36 의 「A2 0.4.3 병합 확인됨」과 한 문서 안에서 충돌한다. D51 과 같은 부류, iter-5 가 좌표를 빠뜨렸다. — Severity: minor — Class: optional — Required fix: 「(A2 0.4.3, `status: implemented`)」로 바꾼다.

D58. EVIDENCE-ATTRIBUTION-124 — acceptance.md:124·acceptance.md:351 — 124행은 트리 `1b071a573` 에서 「`go test -list '<위 정규식 그대로>'`」를 돌렸다고 적지만, 「위 정규식」은 v0.3.5 에서 고정형으로 바뀌었고 기록된 출력(126~130행)과 351행 셀은 옛 비고정 정규식의 측정이다. 선택 결과(기존 3개)는 새 정규식에서도 같음을 이번 실행에서 관측했으므로 결과는 참이다. — Severity: minor — Class: optional — Required fix: 124행에 측정 당시 명령(옛 정규식)을 그대로 적고 「v0.3.5 고정형에서도 같은 3개 선택 — 트리 `a02e7c02b`」를 한 줄 덧붙이거나, run 의 RED 셀 재측정 때 새 정규식으로 다시 잰다.

D59. DESIGN-§12-READER-ROW — design.md:325 — revoke 판독기 행의 픽스처 열은 7개인데 기대 열은 6개이고, 첫 픽스처의 기대가 괄호 안(「디렉터리 부재 → 해제·오류 없음」)에 들어가 열 정렬이 어긋난다. AC-GR-023 이 정본이라 동작상 모호함은 없다. — Severity: minor — Class: optional — Required fix: 기대 열 앞에 「해제(오류 없음) /」를 넣어 7개로 맞춘다.

## 부채(잔여) 목록

- D56 — blocking, run 착수 전 수리 필요(AC-GR-003 을 채울 수 없게 만드는 유일한 결함)
- D57·D58·D59 — optional, 한 번에 묶어 처리 가능

## Regression Check (iter-5 → 이번)

- D47: **UNRESOLVED(부분)** — 건너뛰기 변수·범주 협소·BASE 전제·판정 지위는 해결, 미등록 `[HARD]` 변이는 여전히 통과(위 탐침 3). 새 하위 테스트가 불가능 RED 를 요구 → D56.
- D48: RESOLVED — AC-GR-016 (17), AC-GR-018 (e) revoke 변형, 하위 테스트 수 17·13.
- D49: RESOLVED — design.md:59, 좌표 재측정 일치.
- D50: RESOLVED — EV-6 72행 `diff` exit 0.
- D51: RESOLVED(잔여 D57).
- D52: RESOLVED — design.md:58, doc.go:135~141 일치.
- D53·D54: RESOLVED.
- D55: RESOLVED(잔여 D58) — spec lint 「No findings」.
- 연쇄 불일치: REQ/AC/design/plan 사이 새 모순 없음. 하위 테스트 수·추적표·허용 목록 행 모두 맞물림. iter-4 이전 처분(D43~D46) 절은 델타가 건드리지 않음.

정체(stagnation) 판정: D47 은 iter-5 에서 처음 나왔고 이번에 절반 해결됐다 — 「3회 연속 무변화」가 아니다.

## Recommendation (최종 회차 — 에스컬레이션)

이 회차가 마지막이므로 오케스트레이터는 사용자에게 세 선택지를 제시해야 한다: (1) PASS-with-debt — D56 을 run 착수 전 선행 수리 항목으로 문서화하고 진행, (2) 범위 축소, (3) 명시적 추가 반복. 이 감사자의 근거가 지지하는 것은 다음과 같다.

- 구조 결함은 없다. 남은 blocking 결함은 D56 하나이고, 원인의 절반이 이 감사자의 iter-5 전제 오류다. 범위 축소는 필요하지 않다.
- D56 수리는 AC-GR-003 4.(ii) 한 항목과 design.md:114 한 행의 문언 교체다. 수리 확인은 판단이 아니라 기계적이다: 수리된 문언이 요구하는 비교를 위 탐침(사본에 `[HARD]` 줄 삽입)에 걸어 FAIL 이 나오는지 보면 된다. 이 확인을 다시 전체 감사로 돌릴 필요는 없지만, 확인 자체는 manager-spec 이나 오케스트레이터의 자가 평가가 아니라 명령 출력으로 남겨야 한다.
- manager-spec 수정 순서: D56(필수) → D57·D58·D59(선택, 같은 커밋).

## 증거 (이번 실행)

| 명령 | 관측 |
|---|---|
| `git diff --stat f40824c82 a02e7c02b` | 6 files, +196 −60 (그중 SPEC 디렉터리 5개 + 보고서 1개) |
| `git diff --name-only 7fe658815 a02e7c02b` | SPEC 디렉터리 6개 + `.moai/reports/t1236/plan-audit-5.md` 뿐 — BASE↔HEAD 코드·규칙 동일 |
| `sed -n 385,416p internal/cli/contract.go` / `sed -n 510,522p` | 389 `opts := sign.Options{`, 408 `seams := sign.Seams{`, 414 `sign.Sign(opts, seams)`, 516 `cmd.AddCommand(verifyCmd, showCmd, signCmd)` |
| `sed -n 44,49p` / `sed -n 130,145p internal/contract/doc.go` | 47 `ReceiptOutcome(...)`, 135 `// # Kickoff receipt`, 139~140 임시 규칙 서술 |
| `grep -n 'Sentinel\|Skipped\|SKIP_VALIDATE' internal/constitution/validator.go` | sentinel 9개(15~40), `skipValidateEnvKey`(43), `Skipped`(139), 건너뛰기 분기 185~191 |
| `grep -rn 'SentinelZoneUnregistered\|ZONE_UNREGISTERED\|SentinelAnchorNotFound\|ANCHOR_NOT_FOUND' internal cmd pkg --include='*.go' \| grep -v _test.go` | validator.go:21·22·27·28·113·323 — 선언·주석·재집계만, append 경로 없음 |
| `go build -o <scratch>/moai ./cmd/moai` | exit 0 |
| `<scratch>/moai constitution validate --format json`(워크트리) | exit 1, drift 9, missing 0, unregistered 0, 72행 |
| `diff <acceptance.md 370~441행> <위 stdout>` | 출력 없음, exit 0 |
| 사본에 미등록 `[HARD]` 줄 삽입 후 같은 명령 | exit 1, 워크트리 출력과 `cmp` exit 0 |
| 사본에서 등록 조항 문장 삭제 후 같은 명령(양성 대조) | drift 10, unregistered 0 |
| `<scratch>/moai spec lint SPEC-AUTONOMY-GATE-REWIRE-001` | exit 0, 「No findings — all SPEC documents are valid」 |
| `go test -list '^(TestContractModeBlocksWellFormed\|TestTemplateNeutralityAudit\|TestTemplateNeutralityAuditC8Preserve\|TestTemplateNoInternalContentLeak)$' ./internal/template/` | 기존 3개 선택, `ok` |
| `grep -n "\-run '[^^]" acceptance.md plan.md design.md` | 0건 |
| `grep -n 'A2 개정본\|d8926ff9a\|개정본 대기' <SPEC>/*.md` | plan.md:144(잔여), plan.md:160·design.md:269(이력), research.md(선언된 v0.3.2 측정) |

## Gaps

- AC-GR-016·018·023 의 테스트는 아직 없다. D48 변이 탐침은 문언 기준 추론이다(실행 관측 아님).
- BASE 트리(`7fe658815`) 자체에서 `constitution validate` 를 다시 돌리지 않았다. 대신 BASE↔HEAD 차이가 SPEC 디렉터리와 보고서뿐임을 `git diff --name-only` 로 확인했으므로 HEAD 측정이 BASE 의 규칙·코드에 대해 성립한다.
- AC-GR-018 (e) revoke 변형이 가정하는 「decide 시점에 계약이 서명돼 있음(revoke 뒤 재결정)」 외에, 미서명 계약에서 판독기가 「현재 seal」을 어떻게 정하는지는 문언에 없다. 이번 범위(D47~D55 델타) 밖이라 결함으로 올리지 않았다.

## Residual-risk

- D56 을 (a) 의 테스트 자체 `[HARD]` 줄 비교로 고쳐도, 기존 `[HARD]` 줄을 **바꾸는**(문장 수정) 변이는 등록 조항이면 DRIFT 로 잡히지만 미등록 기존 줄이면 「현재 ⊆ BASE」에서 새 줄로 보여 잡힌다 — 반대로 줄을 **지우는** 변이는 부분집합 조건을 통과한다. 미등록 `[HARD]` 줄 삭제가 위험인지는 이 SPEC 의 요구 밖이다.
- `ANCHOR_NOT_FOUND` 도 `Validate` 가 내지 않으므로, 블록 삽입이 등록 조항의 제목(앵커)을 바꾸는 경우는 조항 문장이 남아 있으면 어떤 검사에도 보이지 않는다(iter-5 Residual-risk 의 전제도 같은 이유로 틀렸다).

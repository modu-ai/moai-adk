auditor-model: claude-opus-5-5[1m]

# SPEC Review Report: SPEC-AUTONOMY-GATE-REWIRE-001
Iteration: 5 (M0 재앵커 델타 감사 — 리드 지시, Tier L 상한 밖. 범위는 `7e82f8b66..f40824c82` 의 B1~B5 처분)
Verdict: FAIL — STOP (점수 하락: iter-4 0.892 → 0.841)
Overall Score: 0.841 (Tier L 문턱 0.85)

감사 대상: 커밋 `f40824c82`(v0.3.4, 부모 `7e82f8b66`), 브랜치 `WT-contract-gate-rewire`, 워크트리 `git status --short` 출력 없음. 범위: `git diff 7e82f8b66 f40824c82 -- .moai/specs/SPEC-AUTONOMY-GATE-REWIRE-001/`(6개 파일, +202/−60)를 리드 결정 B1~B5 와 코드 트리에 대조하고, 교차 층 일관성(verification-completeness.md §3)과 AC-GR-003·AC-GR-023 변이 탐침(§2)을 수행했다. 수정되지 않은 절은 재감사하지 않았다. Reasoning context ignored per M1 Context Isolation — 호출 측이 전한 B1~B5 요약과 「9 DRIFT, exit 1」은 대조 기준으로만 쓰고 전부 다시 쟀다. `grep -rn audit_model .moai/config/sections/` 출력 없음 → 교차 모델 호출 없이 이 감사자 단독 판정.

첫 줄 규약 주: 이 카드의 이전 보고서(plan-audit-1~4)에는 auditor-model 줄이 없다. 같은 저장소의 `.moai/reports/t1175/sync-audit.md` 1행 형식(`auditor-model: <모델 id>`)을 따랐다.

## Must-Pass Results

- [PASS] MP-1 REQ 번호: `grep -o '^- \*\*REQ-GR-0[0-9][0-9]' spec.md` → REQ-GR-001~025 연속 25개, 중복 없음. AC 는 `^### AC-GR-` 25개, `uniq -d` 출력 없음.
- [PASS] MP-2 GEARS (요구 층만 판정): 델타가 고친 REQ-GR-007·009(When), 012·013·018·020·022(When)·025(Ubiquitous)는 패턴을 유지한다(spec.md:96 「**When** `moai contract kickoff-check` …」, spec.md:101 「The moai 소유 저장소 shall …」). AC 층의 Given-When-Then 은 이 기준으로 채점하지 않았다.
- [PASS] MP-3 frontmatter: 12필드 모두 있음(`version: "0.3.4"` 따옴표, `updated: 2026-09-27`, `tier: L`). 금지 별칭 없음.
- [N/A] MP-4 언어 중립: 단일 저장소 규칙·Go 코드 SPEC. 템플릿 개정 문장의 중립성은 AC-GR-009 가 다룬다.
- [PASS] MP-5 D7: 참조 SPEC 다섯 개 상태 — ALWAYS-LOADED-DIET-002 `completed`, AUTONOMY-CONTRACT-001 `completed`, AUTONOMY-ESCALATION-001 `implemented`, AUTONOMY-TIERS-001 `completed`, JEV-CORE-001 `completed`. retired/superseded/archived 없음 → BLOCKING 없음.
- [PASS] MP-6 D8: `grep -c syscall` spec/plan/acceptance/design 모두 0.
- [PASS] MP-7 clarification gate: `grep -rn '\[NEEDS CLARIFICATION' plan.md research.md` 출력 없음.

## B1~B5 대조 결과

| 처분 | 판정 | 근거(이번 실행에서 잰 값) |
|---|---|---|
| B1 askuser 앵커 | **정확** | 로컬·템플릿 모두 `206:## Ambiguity Triggers and Exceptions`, 208행 스텁 문단, 209행 빈 줄, `210:## Free-form Circumvention Prohibition`. 이동처 `askuser-protocol-reference.md:229:### The Five Exceptions`(로컬·템플릿 동일). design.md:38·plan.md M3·spec HISTORY 가 같은 좌표를 인용 |
| B2 jev_ask 위치 | **정확** | `grep -n jev moai-mcp-tools.md` 0건(로컬·템플릿). 카탈로그 `138:` 도구 표 행, `216:` `Judgment (gated)` 행(두 사본 동일), `paths:` 조건부 로드, `wc -m` 18,396. 예산은 spec D-6·design §4·AC-GR-010·plan NC-6 이 같은 수치(상시 1,500 / 행당 300 / 사본당 600)로 맞물린다 |
| B3 A3 revoke 판독기 | **문서는 정확, 소비 경로 검증 누락(D48)** | `internal/escalation/record.go`: `ParseRecord`(118), `Fingerprint(class string, parts ...string)`(138), `RecordDir`(144), `RecordPath(worktreeRoot, card, class, fingerprint string, ordinal int)`(151), `NeedsDecision`(163, 180행 `r.Status == StatusOpen && (Kind == contract‖operational)`). 필드 `card`·`spec`·`kind`·`class`·`fingerprint`·`status` 실재(65~72행). A2 spec 0.4.3 `status: implemented`, 539행 「A revoke record carries `status: resolved` and a non-empty `decider`」. 경계 문장은 REQ-GR-022·design §10 에 있다 |
| B4 ReceiptOutcome | **좌표 정확, 허용 목록 불완전(D49)·주석 제약 충돌(D52)** | `receipt.go` 208행 주석 시작, 216행 `func ReceiptOutcome`, 232행 닫는 괄호, 첫 분기 220행 `EffectiveDecider == DeciderLLMJev`. 호출부 `sign/sign.go:519`, `doc.go:47` 시그니처 주석, `receipt_test.go` 의 `TestValidateKickoffReceipt_AC016`(105)·`TestReceiptOutcome`(417), `sign/ac_contract_016_test.go:114` `t_llm_jev_both_approve_interim_rule` |
| B5 AC-GR-003 비증가 | **기준선 정확, 판정식 협소·공허 통과 경로(D47), 원장 비축자(D50)** | 트리에서 빌드한 바이너리(`go build -o <scratch>/moai ./cmd/moai`, build=0)로 `constitution validate --format json` → exit 1, `drift_count 9`, id `CONST-V3R2-013·014·015·016·017·033·049·152·153`. 설치본(`~/go/bin/moai`) 출력과 `cmp` 동일 |

## Category Scores

| 차원 | 점수 | 밴드 | 근거 |
|---|---|---|---|
| Clarity | 0.85 | 0.75~1.0 | B1·B2 좌표가 모든 층에서 일치. 감점: revoke 판독기의 「디렉터리 목록 실패 = 차단」이 디렉터리 부재를 차단으로 읽게 둔다(D54, A2 `NeedsDecision` 은 `filepath.Glob` 이라 부재 = 0건). AC-GR-003 둘째 명령이 「사람이 읽는 확인용」이면서 「판정은 개수와 id 집합」이라 판정 성격이 모호 |
| Completeness | 0.90 | 0.75~1.0 | 절 구성·frontmatter 완비, §K 처분표·§10.6 재측정표 추가. 감점: 허용 목록에서 `internal/cli/contract.go` 누락(D49) |
| Testability | 0.75 | 0.75 | 변이 탐침 두 건 통과 실패 — AC-GR-003 은 DRIFT 외 범주와 건너뛰기 환경 변수에 무방비(D47), REQ-GR-007·009(e) 에 새로 넣은 판독기 조건을 어떤 AC 도 소비 경로에서 확인하지 않는다(D48). EV-6 원문이 비축자(D50) |
| Traceability | 0.88 | 0.75~1.0 | REQ-GR-012·020 → AC-GR-023 추가됨. 감점: AC-GR-023 제목은 REQ-GR-009 를 주장하나 추적표 REQ-GR-009 행은 AC-GR-018 뿐(D53); plan·acceptance 에 옛 A2 표지 잔존(D51) |

집계: 조화 평균 4 / (1/0.85 + 1/0.90 + 1/0.75 + 1/0.88) = 0.841. Tier L 문턱 0.85 미달.

## Defects Found (structured defect-list)

D47. B5-NARROW — acceptance.md:37(AC-GR-003 Then)·design.md:114 — v0.3.3 판의 「OK — no drift or violations」를 「DRIFT 수 ≤ BASE, DRIFT id ⊆ BASE」로 바꾸면서 판정 범위가 DRIFT 한 범주로 좁아졌다. `internal/constitution/validator.go:15~40` 은 `ZONE_UNREGISTERED`(원본 파일의 미등록 `[HARD]` 규칙)·`ANCHOR_NOT_FOUND`·`SOURCE_FILE_MISSING`·`FROZEN_WITHOUT_CANARY`·`INVALID_ZONE_CLASS` 도 낸다. 이 SPEC 은 always-loaded 파일 세 개에 블록을 넣으므로, `[HARD]` 표지를 단 블록 하나가 `unregistered_count` 를 0→1 로 올려도 AC-GR-003 은 통과한다(변이 탐침 성립). 게다가 `Validate` 는 `MOAI_CONSTITUTION_SKIP_VALIDATE=1` 일 때 항목 없이 `Skipped` 를 돌려주므로(validator.go:185~191) 양쪽이 모두 0 → 「0 ≤ 0, ∅ ⊆ ∅」로 공허 통과한다. — Severity: major — Class: blocking — Required fix: (1) 비교 대상을 DRIFT 만이 아니라 상태가 OK 가 아닌 모든 항목의 `(sentinel, id)` 쌍으로 넓히고, `missing_count`·`unregistered_count` 가 BASE 값(EV-6: 둘 다 0) 이하임을 함께 요구한다. (2) 테스트가 `MOAI_CONSTITUTION_SKIP_VALIDATE` 를 지운 상태로 돌고, 양쪽 결과가 `Skipped` 가 아니며 BASE 쪽 DRIFT id 집합이 EV-6 의 9개와 같음을 전제 단언으로 확인한다. (3) 반증 하위 테스트에 「미등록 `[HARD]` 줄을 넣은 사본 → FAIL」을 추가한다. (4) 둘째 명령이 판정의 일부인지 참고용인지 한 문장으로 정한다.

D48. B3-CONSUMER-UNTESTED — spec.md:96(REQ-GR-007)·spec.md:98(REQ-GR-009 (e))·acceptance.md:203(AC-GR-016)·acceptance.md:227(AC-GR-018) — B3 로 kickoff-check 와 decide 전제조건 (e) 에 「A3 revoke 판독기가 차단을 보고하지 않음」 조건이 새로 들어갔지만, 그 조건을 소비 경로에서 확인하는 픽스처가 없다. AC-GR-016 의 `revoked` 픽스처 (8) 은 저장소 revoke 사건이 있는 경우뿐이고, AC-GR-018 은 (e) 를 깬 픽스처 하나로 두 하위 조건을 함께 덮는다(기대 문장만 acceptance.md:235 에서 고쳤다). AC-GR-023 은 판독기 함수를 단독으로 시험한다. 따라서 판독기를 아예 부르지 않는 kickoff-check·decide 변이가 AC-GR-016·018·023 을 모두 통과한다 — B3 가 막으려던 D21 꼬리 절단(저장소 revoke 사건 삭제, 에스컬레이션 기록만 남음)이 여전히 검증되지 않는다. — Severity: major — Class: blocking — Required fix: AC-GR-016 에 픽스처 (17) 「(3) + 저장소 revoke 사건 없음 + A2 형식 `kind: revoke` 기록(현재 seal 지문)만 있음 → exit 1 `revoked`」를, AC-GR-018 에 (e) 의 두 변형 「열린 `kind: contract` 기록」·「`status: resolved` 인 revoke 기록만」을 각각 넣고 둘 다 `outcome: human`·`precondition:e` 를 기대한다. 하위 테스트 수(16·12)를 갱신하고, AC 수 자체는 바뀌지 않으므로 AC 스냅숏 재생성 의무는 없다.

D49. B4-ALLOWLIST-CLI — design.md:59(§2 24행)·AC-GR-003 — A1 `contract` 명령 트리는 `internal/cli/contract.go` 의 `newContractCmd()` 안에서 조립된다(516행 `cmd.AddCommand(verifyCmd, showCmd, signCmd)`, 521행 `rootCmd.AddCommand(newContractCmd())`). 서명 옵션과 이음매도 같은 파일 390~414행에서 만들어 `sign.Sign(opts, seams)`(414행)로 넘긴다. design §7.1 은 「CLI 가 상수 `jevDoctrineAmended` 를 넘긴다」와 「주입 가능한 저장소 이음매」를 요구하고, 24행은 decide·kickoff-check·revoke 를 「A1 `contract` Cobra 명령에 하위 명령 추가」로 적는다 — 둘 다 `contract.go` 편집이 자연스러운 경로인데, 허용 목록 24행은 새 파일 세 개와 그 테스트만 싣는다. 그대로 run 에 들어가면 `TestContractModeChangeSetAllowlist` 가 FAIL 하거나 구현이 허용 목록을 우회해야 한다. B4 와 같은 부류(실제 편집 지점이 허용 목록 밖)이며 M0 재앵커가 놓쳤다. — Severity: major — Class: blocking — Required fix: design §2 24행에 `internal/cli/contract.go`(`newContractCmd` 의 하위 명령 등록 516행, 서명 옵션·이음매 조립 390~414행 — BASE `7fe658815` 기준)를 추가하고, 편집 범위를 그 두 곳으로 한정하는 문장을 붙인다.

D50. EV6-NOT-VERBATIM — acceptance.md:425·acceptance.md:438(§B.2 주석) — EV-6 stdout 원문의 한 줄이 프로그램 출력과 다르다. 이번 실행의 출력(트리 빌드·설치본 모두)은 `project_<epic>_<s...` 인데 원장은 `project_<epic>_<s...` 다. `diff <원장 364~435행> <이번 JSON>` → 62행 한 줄만 다름, exit 1. 원장 주석 「stdout 의 `<` 는 프로그램이 낸 JSON 이스케이프 그대로다」는 사실과 반대다. verification-completeness.md §2.1 의 「verbatim = 원시 바이트」를 어긴다. — Severity: minor — Class: blocking — Required fix: 해당 줄을 `<`·`>` 이스케이프 그대로 옮기고 주석을 「stdout 은 Go JSON 인코더의 `<` 이스케이프를 그대로 담는다」로 고친다.

D51. STALE-A2-REFS — plan.md:10·plan.md:36·plan.md:75·acceptance.md:488 — B3 로 A2 기준이 「0.4.3, M0 대조 완료」가 됐는데 plan §A 는 여전히 「A2 기준: 리드가 전달한 최종 형식(A2 개정본 대기, 현재 커밋 `d8926ff9a` 는 철회된 형식)」이고, plan M0(75행)과 acceptance 완료 정의(488행)는 **[A2 개정본으로 재확인]** 항목 대조를 요구하지만 spec.md 에는 그 표지가 더 이상 없다(`grep -n 'A2 개정본으로 재확인' spec.md` 는 HISTORY 외 0건). design.md:3 과 progress.md:11 은 새 기준을 적고 있어 같은 SPEC 안에서 두 기준이 공존한다. research.md 는 머리에서 「§10.6 외 절은 v0.3.2 측정」이라 밝혔으므로 대상 아님. — Severity: minor — Class: blocking — Required fix: plan.md:10 을 design.md:3 과 같은 문장으로, plan.md:36 의 전제를 「A2 0.4.3 병합 확인됨(M0)」으로, plan.md:75·acceptance.md:488 의 **[A2 개정본으로 재확인]** 언급을 삭제하거나 「A2 0.4.3 인용 항목」으로 바꾼다.

D52. DOCGO-SCOPE-CONFLICT — design.md:58(§2 23행)·plan.md:212 — 허용 목록이 `internal/contract/doc.go` 편집을 「47행 시그니처 목록 주석만」으로 묶었는데, 같은 파일 139~141행 「ReceiptOutcome then applies the interim A1 rule (effective decider llm+jev → receipt_requires_human)」은 M7 의 조건화 뒤 무조건적 서술이 되어 틀린다. run 은 낡은 주석을 남기거나 SPEC 의 범위 문장을 어겨야 한다. `receipt.go:208~215` 의 함수 주석도 같은 이유로 고쳐야 하지만 그 파일은 편집 대상이라 문제없다. — Severity: minor — Class: blocking — Required fix: 23행의 `doc.go` 범위를 「47행 시그니처 목록과 135~141행 `# Kickoff receipt` 문단」으로 넓힌다.

D53. TRACE-009 — acceptance.md:286(AC-GR-023 제목)·acceptance.md 추적표 REQ-GR-009 행 — AC-GR-023 은 REQ-GR-009 를 대상으로 주장하지만 추적표 009 행은 `AC-GR-018` 뿐이다(012·020 행은 갱신됨). — Severity: minor — Class: optional — Required fix: REQ-GR-009 행에 AC-GR-023 을 더하거나 AC-GR-023 제목에서 009 를 빼고 D48 의 AC-GR-018 픽스처로 연결한다.

D54. READER-ABSENT-DIR — design.md:265(판독기 표 「오류」 행)·spec.md:129(REQ-GR-022) — 「디렉터리 목록·읽기·파싱 실패는 차단」은 에스컬레이션 디렉터리가 아예 없는 정상 카드(대부분)를 차단으로 읽을 여지를 남긴다. `os.ReadDir` 로 구현하면 ENOENT 가 오류가 된다. A2 `NeedsDecision` 은 `filepath.Glob` 이라 부재 = 0건이다. AC-GR-016 의 정상 경로 픽스처가 결국 잡겠지만 문언은 모호하다. — Severity: minor — Class: optional — Required fix: 「디렉터리 부재는 기록 0건(해제)이며 오류가 아니다」를 표에 한 줄 넣고 AC-GR-023 에 (r0) 부재 → 해제 픽스처를 더한다.

D55. VACUOUS-RUN-PATTERNS — acceptance.md 26곳(16·19·28·40·53·…·317행)·plan.md:59 — 트리 빌드 `moai spec lint SPEC-AUTONOMY-GATE-REWIRE-001` → `0 error(s), 27 warning(s)`, 전부 `VacuousTestAssertion`(26건 run-pattern 비고정, 1건 outcome-assertion 구분자 없음). 설치본 `moai spec lint` 는 같은 SPEC 에 「No findings」를 냈다 — 설치본이 이 규칙보다 오래됐다. 델타가 새로 쓴 명령(acceptance.md:40, AC-GR-003)도 비고정이다. 판정: **run 전 필수 수정은 아니다.** 비고정 `-run` 의 위험은 과선택(이름 접두가 같은 다른 테스트까지 고름)이고, 공허 통과의 주 경로인 빈 선택은 acceptance §B.1 의 「run 에서 `go test -list` 로 의도한 이름을 모두 고르는지 재측정, 빈 선택은 통과 아님」이 이미 막는다. 다만 접두 공유 이름이 생기면 기대한 `--- PASS` 줄을 다른 테스트가 채울 수 있다. — Severity: minor — Class: optional — Required fix: run 에서 각 테스트가 생길 때 해당 AC 명령을 `-run '^(A|B)$'` 형태로 고정하고 기대 줄을 `--- PASS: <이름> ` (뒤 공백 포함)으로 바꾼다. 수정은 manager-spec 소관이므로 run 착수 전 한 번에 묶어 재위임하는 편이 싸다.

## 변이 탐침 기록

- **AC-GR-023 (판독기)**: status 를 키로 쓰는 판독기(A2 `NeedsDecision` 재사용) → (r1) FAIL, 잡힘. 지문을 무시하고 `kind: revoke` 만 보는 판독기 → (r2)·(r3) FAIL, 잡힘. 파싱 오류를 건너뛰는 판독기 → (r5) FAIL, 잡힘. `card`·`spec` 무시 변이는 지문이 seal 에 묶여 동작상 구분되지 않음 — 결함 아님. 판독기 **단위**로는 변이에 강하다. 약점은 판독기를 부르지 않는 소비자 변이(D48)다.
- **AC-GR-003 (헌법 비증가)**: BASE 쪽을 현재 트리로 잘못 읽는 테스트 변이는 둘째 명령(EV-6 의 9개 상한·id 집합)이 막는다 — 둘째 명령이 판정에 포함될 때만(D47 (4)). 미등록 `[HARD]` 블록 변이와 건너뛰기 환경 변수 변이는 통과한다(D47).

## Regression Check (iter-4 → 이번)

iter-4 의 D43~D46 은 v0.3.3 에서 처분됐고, 이번 델타에서 그 절(REQ-GR-007 사유 우선순위, AC-GR-017 매 head 주입, `[REF]` 표지, 세션 식별자 삭제)의 문언은 그대로 유지됐다 — 회귀 없음. 점수 하락(0.892 → 0.841)의 원인은 새 결함 D47·D48·D50~D52 와, 이전 감사가 놓친 잠복 결함 D49 다.

## Recommendation

STOP 신호: 점수가 직전 반복보다 낮으므로 무조건적 추가 반복을 권하지 않는다. 다만 하락 원인은 구조 결함이 아니라 국소 수정 여섯 건이며, 범위 축소(분할)는 필요하지 않다고 판단한다. 오케스트레이터가 사용자에게 제시할 선택지(범위 축소 / PASS-with-debt / 명시적 추가 반복) 중 이 감사자의 근거는 「blocking 여섯 건을 제자리 수정한 뒤 이 결함 목록에 한정한 재감사」를 지지한다.

manager-spec 수정 순서:

1. D49 — design §2 24행에 `internal/cli/contract.go`(516행, 390~414행) 추가. run 을 다시 멈출 수 있는 유일한 결함이다.
2. D48 — AC-GR-016 픽스처 (17), AC-GR-018 (e) 두 변형, 하위 테스트 수 갱신.
3. D47 — AC-GR-003 판정식을 비-OK 전 범주로 넓히고, 건너뛰기 환경 변수 제거·비건너뜀·BASE=EV-6 전제 단언, 미등록 `[HARD]` 반증 추가, 둘째 명령의 판정 지위 명시.
4. D50 — EV-6 62행 이스케이프 원문 복원, 주석 정정.
5. D51 — plan.md:10·36·75, acceptance.md:488 의 옛 A2 표지 정리.
6. D52 — `doc.go` 허용 범위에 135~141행 추가.
7. (선택) D53·D54·D55.

## 증거 (이번 실행)

| 명령 | 관측 |
|---|---|
| `git diff --stat 7e82f8b66 f40824c82 -- <SPEC dir>` | 6 files, +202 −60 |
| `grep -n 'Ambiguity Triggers\|Five Exceptions\|^## ' askuser-protocol.md`(로컬·템플릿) | 206·208·210행 일치, 「Five Exceptions」는 12행 목록에만 |
| `grep -n 'Five Exceptions' askuser-protocol-reference.md`(두 사본) | `229:### The Five Exceptions (Stage 1 is skipped)` |
| `grep -n jev moai-mcp-tools.md / moai-mcp-tools-catalogue.md`(두 사본) | 0건 / 138·145·216행 |
| `grep -n 'func \|Kind\|Status\|Fingerprint ' internal/escalation/record.go` | 87·118·138·144·151·163행 함수, 25~33행 상수, 65~72행 필드 |
| `sed -n 205,236p internal/contract/receipt.go` | 208 주석, 216 함수, 220 임시 규칙 분기, 232 끝 |
| `grep -rn ReceiptOutcome internal cmd` | 제품 호출부 `sign/sign.go:519` 하나, `doc.go:47·139`, `sign/doc.go:56` |
| `grep -n 'AddCommand\|func init' internal/cli/contract.go` | 516 `cmd.AddCommand(verifyCmd, showCmd, signCmd)`, 521 `rootCmd.AddCommand(newContractCmd())` |
| `<scratch>/moai constitution validate --format json`(트리 빌드) | exit 1, drift 9, missing 0, unregistered 0, retired 4 |
| `diff <원장 EV-6> <이번 JSON>` | 62행 한 줄 차이(`<` vs `<`) |
| `<scratch>/moai spec lint SPEC-AUTONOMY-GATE-REWIRE-001`(트리 빌드) | exit 0, 0 error, 27 warning(`VacuousTestAssertion`) |
| `moai spec lint …`(설치본 `~/go/bin/moai`, 2026-09-25 설치) | 「No findings」 — 규칙 이전 빌드 |

## Gaps

- `BASE` 트리(`git archive 7fe658815`) 위에서 헌법 검증을 직접 다시 돌리지 않았다. HEAD 에서 잰 값과 원장이 9개 id 까지 같고, 원장이 「`BASE`↔HEAD 차이는 SPEC 디렉터리뿐」이라 적었으나 그 차이 목록도 이번에 재지 않았다.
- AC-GR-016·018·023 의 테스트는 아직 없으므로 변이 탐침은 문언 기준 추론이다(실행 관측 아님).
- 부모 커밋(`7e82f8b66`)의 lint 경고 수는 재지 않았다 — 27건 중 델타가 새로 만든 것이 몇 건인지 단정하지 않는다.

## Residual-risk

- D47 을 고쳐도, 블록 삽입이 기존 등록 조항의 앵커(제목)를 바꾸는 경우는 `ANCHOR_NOT_FOUND` 로만 드러난다 — 전 범주 비교가 들어가야 잡힌다.
- revoke 판독기의 차단 판정은 지문이 현재 seal 과 맞는지에 의존한다. 서명기가 seal 계산을 바꾸는 A1 개정이 오면 기존 revoke 기록이 조용히 해제될 수 있다(A1 소관, 이 SPEC 범위 밖).

auditor-model: claude-opus-5-5
verdict: FAIL
overall_score: 0.69
must_pass_failed: 0
blocking_count: 5
plan_artifact_hash: b154a7b15d4b2d4b99d8a3b3b56777b63769a95d87a62bab0bff0daf49f49207
audited_sha: b805ac2d2429f414a7b3634d4196711d688289be
required_backend: codex fail
required_backend_fail: codex
convergence_overall: fail
scope: full
fix_scope: spec.md#REQ-FBM-001, spec.md#REQ-FBM-003, spec.md#REQ-FBM-009, spec.md#§E, decision-index.md#Q1, plan.md#§A, plan.md#§C, acceptance.md#AC-FBM-001, acceptance.md#AC-FBM-002, acceptance.md#AC-FBM-003, acceptance.md#AC-FBM-004, probe/zz_t1923_boot_probe_test.go.txt
defect_class: D1=design, D2=ac-wording, D3=ac-wording, D4=traceability, D5=ac-wording
audit_receipt: rcpt-873f4860189236ef74793036

# SPEC 검토 보고서: SPEC-FACTORY-BOOT-MINIMAL-001 (카드 t1923)

- 반복: 1/3 (iteration 0 요청, 첫 전체 감사)
- 판정: FAIL — 차단 결함 5건, 종합 0.69 (Tier M 통과 기준 0.80 미달), 필수 백엔드 codex 판정 fail
- 감사 대상 트리: HEAD `b805ac2d2429f414a7b3634d4196711d688289be` (브랜치 WT-factory-lane-internal, 작업 트리 깨끗함)
- 감사자 버전: plan-auditor (opus, high)
- 작성자 추론 맥락은 M1 맥락 격리에 따라 무시했다. 판단 근거는 디스크의 산출물과 이 감사에서 직접 실행한 명령뿐이다.

## 필수 통과 항목 (Must-Pass)

- [PASS] MP-1 REQ 번호 일관성 — spec.md L72~L82에 REQ-FBM-001~011 연속, 중복·결번 없음, 세 자리 0채움 일관.
- [PASS] MP-2 EARS/GEARS 형식 (요구사항 층 판단) — REQ 11개 모두 GEARS 패턴 또는 레거시 EARS 등가형. REQ-FBM-009(L80)만 레거시 EARS 부정형 "The change set shall not …"이며 하위 호환 창(2026-11-22까지) 안의 레거시 등가형으로 셈한다. AC는 검증 층(Given-When-Then)이라 이 항목에서 채점하지 않았다.
- [PASS] MP-3 frontmatter — spec.md L2~L13에 정식 12개 필드 전부 존재, 거부 별칭 없음, `version: "0.1.0"` 따옴표 semver, `priority: P0`, `lifecycle: spec-anchored`, `phase: "v3.2.0"`(단계명 아님). 추가 필드 `tier: M`, `related_specs`는 선택 필드.
- [N/A] MP-4 언어 중립성 — 프로그래밍 언어 도구를 다루는 SPEC이 아니다. en/ko/ja/zh는 대화 언어 로케일이다.
- [PASS] MP-5 D7 교차 SPEC 조정 — D7 동사 출력: 참조 SPEC 5개(BOOT-MINIMAL draft, FACTORY-BOOTSTRAP completed, FACTORY-SELF-DISPATCH completed, ROLE-INJECTION-BUDGET completed, SESSION-START-GUIDE-I18N completed), retired/superseded/archived 0건, `REVIEW:` 0건. 단, D7 동사가 보지 못하는 미참조 완료 SPEC과의 충돌은 별도 차단 결함 D1로 기록했다(Group 6 CN-1/CN-2).
- [PASS] MP-6 D8 플랫폼 규율 — spec.md에 `syscall` 0건, D8 동사 `BLOCKING:` 출력 없음.
- [PASS] MP-7 명확화 게이트 — plan.md의 `NEEDS CLARIFICATION` 0건(직접 grep). research.md 없음(Tier M).
- [PASS] MP-8 RED-now 셀 완결성 — release-blocking AC 4개(AC-FBM-001~004) 모두 증거 장부(E1~E4)에 명령·종료 코드·트리 SHA(문서 수준 고정 `5995962990…`)·원문 출력(파일 또는 인라인)이 있다. E1은 이 감사에서 재실행해 72칸 수치가 바이트 단위로 같음을 확인했다. E2는 종료 코드 0·PASS 출력이지만, AC-FBM-002의 Then이 요구하는 칸별 고정분 줄·순회 칸 수 줄이 기록된 stdout에 없으므로 기준 그 자체는 빨간 상태로 읽힌다(codex는 이를 MP-8 결함으로 보았으나 감사자는 동의하지 않는다 — 다만 그 약점은 D3에 흡수했다). E3·E4는 재실행 결과가 기록과 같다.
- [PASS] MP-9 교차 산출물 순서 — CN-4 동사 출력 `COLLECTED: 6 milestones in plan order (M1 M2 M3 M4 M5 M6), 0 exit bindings, 0 ordering candidates` / `NONE: 0 records … match (before|after|first|prior to|pre-change)`. 한국어 순서 문장은 손으로 읽었다: M1에서 테스트 RED → M2 재배치 → M3 축소 → M5 AGENTS.md 마지막, DoD의 「M1 직후 RED 원문」과도 충돌하지 않는다.

## 차원별 점수 (루브릭 기준)

| 차원 | 점수 | 루브릭 구간 | 근거 |
|---|---|---|---|
| Clarity(명확성) | 0.75 | 0.75 | REQ 문장은 대체로 단일 해석(spec.md L72~L82). 다만 완료 정의(acceptance.md L165~L169)가 AC-FBM-008을 「관측 기록」으로만 두어 라이브 넘침이 관측될 때의 처분이 불명확하고, 카드의 원소 목록을 리더·레인에 비대칭으로 해석한 근거가 기록돼 있지 않다(D9). |
| Completeness(완결성) | 0.75 | 0.75 | HISTORY·배경·범위·REQ·AC·Out of Scope(`### Out of Scope — …` H3 4개, `-` 항목 있음, spec.md L95~L112) 모두 있음. 그러나 이 변경이 뒤집는 완료 SPEC 계약(SPEC-AUTONOMY-BATCH-GATE-001 REQ-BGS-016/017)과 그 계약을 고정한 테스트가 범위·관련 SPEC·계획 파일 목록 어디에도 없다(D1). |
| Testability(시험 가능성) | 0.50 | 0.50 | release-blocking AC 4개 모두 요구사항을 어기면서 통과하는 변이를 쓸 수 있다(D2~D5). AC-FBM-002의 Then은 「칸마다 수치를 보고」의 출력 형식을 정하지 않는다. |
| Traceability(추적성) | 0.75 | 0.75 | 추적 동사 `COLLECTED: 11 REQ definitions (acceptance input: read)`, UNCOVERED 0, ORPHAN 0. 그러나 REQ-FBM-003의 핵심 의무(제거 전 재배치·포인터가 위치를 지명)와 REQ-FBM-008의 「breach 시 실패」는 매핑된 AC가 실제로 검사하지 않는다 — 명목상 매핑. |

종합 0.69 = 네 차원 평균(0.6875). Tier M 통과 기준 0.80에 못 미친다.

## 결함 목록 (structured defect-list)

D1. CROSS-SPEC-CONTRACT — spec.md:L72, spec.md:L74, spec.md:L80, decision-index.md:L5, plan.md:L7 — REQ-FBM-001의 「only these elements」와 REQ-FBM-003이 제거 대상으로 명시한 "batch gate summary"는 완료 SPEC `SPEC-AUTONOMY-BATCH-GATE-001`의 REQ-BGS-016/017("[Q5 decided]" — 팩토리 리더 SessionStart 공지가 4개 로케일 모두에서 `batch gate summary` 문장과 `auto-semantics.md` §9.2 포인터를 갖는다)과 정면으로 충돌한다. 이 SPEC은 해당 SPEC을 어디서도 이름하지 않고(related_specs에도 없음), decision-index Q1은 SESSION-START-GUIDE-I18N-001만 대체한다고 적었다. 그 계약을 고정한 `TestLeaderNoticeBatchGatePointer`(internal/hook/session_start_leader_gate_notice_test.go:160, 변이 5종 포함)와, t85 성문화 AC인 `TestFactoryLeadNoticeCarriesDispatchDiscipline`(+한국어 변형, session_start_factory_test.go:237 — "every card is routed WHOLE to one lane", "started producing output", "FACTORY fan-out only" 등 렌더 문자열 고정)도 plan.md §A 변경 대상 목록에 없고, §C 3단계의 grep 패턴(필드 이름만 찾음)으로는 렌더 문자열 고정 테스트가 잡히지 않는다. REQ-FBM-009가 다시 쓰기를 허용하는 것은 「guide test」뿐이라, 실행 단계는 테스트를 삭제·약화할 수 없고 다시 쓸 근거도 없는 상태에 놓인다. 재현: `grep -n "REQ-BGS-016\|REQ-BGS-017" .moai/specs/SPEC-AUTONOMY-BATCH-GATE-001/spec.md` → L159, L160(`status: completed`); `grep -n "AUTONOMY-BATCH-GATE\|BGS-017" .moai/specs/SPEC-FACTORY-BOOT-MINIMAL-001/*.md` → 0건. codex도 같은 결함을 독립적으로 보고했다(P1 2건). — Severity: major — Class: blocking — Required fix: (a) 배치 게이트 문장을 REQ-FBM-001의 허용 원소에 추가해 유지하거나, (b) REQ-BGS-016/017 대체를 spec.md에 명시(대체 대상 REQ·근거·운영자 결정 출처)하고 decision-index에 그 행을 추가하며, related_specs에 SPEC-AUTONOMY-BATCH-GATE-001을 넣는다. 어느 쪽이든 plan.md §A에 `session_start_leader_gate_notice_test.go`, `session_start_factory_test.go`(t85 테스트 2개 포함), `session_start_factory_worker_test.go`를 넣고, REQ-FBM-009의 다시 쓰기 허용 범위를 「REQ-FBM-001/002가 대체한 계약을 고정한 테스트」로 넓혀 이름을 나열한다. §C 3단계 grep에 렌더 문자열 표지("Card routing", "Stagger", "batch gate summary", "Entry points", "up to 10 agents")를 더한다.

D2. AC-001-MUTANT — acceptance.md:L23~L29, probe/zz_t1923_boot_probe_test.go.txt:L95~L111 — 탐침은 startup 칸에서 run id·리더 라벨·레인 라벨·Matrix 경로·docs URL의 존재와 `--auto-leader`/`--auto-lane`의 부재만 단언한다. 리더 공지에서 autoModeGuideLeader만 빼고 카드 라우팅·stagger·게이트 요약·운영 상태·진입점·팬아웃 문장을 그대로 둔 변이는 리더 en 고정분이 이미 2,573(≤ 2,847)이므로 AC-FBM-001을 통과하면서 REQ-FBM-001의 「only these elements」를 어긴다. 레인 시작 문장(`-l`), 소켓 줄, 재배치 포인터 줄도 단언하지 않고, REQ-FBM-001이 명시한 두 채널 중 systemMessage(운영자 사본 `factoryBootstrapNotice`)는 검사하지 않는다. 재현: 탐침 소스 L95~L111 판독, 그리고 이 감사의 재실행 출력 `PROBE leader//en/startup/deployed fixed=2573 composite=8830 core=4048`. — Severity: major — Class: blocking — Required fix: AC-FBM-001 Then과 탐침에 (i) REQ-FBM-001/002 필수 원소 전체(레인 시작 명령 `moai cc -l`, 소켓 줄, 포인터 줄의 목적지 경로 포함)의 존재, (ii) 제거 원소별 표지(예: "Card routing", "Stagger", "Entry points", "up to 10 agents", 운영 상태 문장 — D1에서 유지를 고르면 batch gate 문장은 제외)의 부재를 두 채널 모두에서 단언하도록 추가한다.

D3. AC-002-MUTANT — acceptance.md:L31~L37 — Then은 「exit 0, 두 테스트 PASS, 칸마다 고정분 수치 보고, 순회 칸 수 줄」뿐이다. 실패 단언 없이 수치만 출력하는 테스트(예: fixed=2848, composite=9147을 로그만 하고 통과)는 Then을 만족하면서 REQ-FBM-008의 「shall fail on any breach」를 어긴다. Then은 출력 형식(문자열)도, REQ-FBM-008이 요구하는 역할 × 디스패치 × 두 면 순회도, 헤더의 0.6.0 분해 기록도 요구하지 않는다. RED-now(E2)는 「출력 줄 부재」라는 형태상의 빨강이라, 출력 줄 하나를 추가하는 것만으로 초록이 된다. 재현: 이 감사의 재실행 `go test -count=1 -run '^(TestRoleInjectionAssemblyBudget|TestCompositeBudgetRealAssembly)$' -v ./internal/hook/` → exit 0, `RESULT lane_en_glm: composite utf16=9946 (cap 10000, margin=54)`, 칸별 고정분 줄 없음. codex도 같은 반례를 보고했다. — Severity: major — Class: blocking — Required fix: Then에 (i) 칸별 출력 줄의 정확한 형식(예: `FIXED role=… src=… lang=… dispatch=… face=… fixed=N composite=N`)과 기대 칸 수, (ii) 고정분 2,848·합본 9,147을 주입하는 경계 변이에서 테스트가 실패(exit 1, 원인 이름 포함)함을 관측하는 절을 추가하고, (iii) 헤더 문자열(「allowance 1,950 + factory fixed ceiling 2,847」) 존재를 grep으로 단언한다.

D4. REQ-003-UNVERIFIED — spec.md:L74, acceptance.md:L39~L45 — REQ-FBM-003의 의무는 「제거되는 모든 원소가 제거 전에 기존 디스크 안내 위치에 착지하고, 포인터 줄이 그 위치를 지명한다」이다. 매핑된 AC-FBM-003은 `grep -c autoModeGuide`가 0인지만 본다 — 안내를 재배치 없이 그냥 지운 변이가 통과한다. 어떤 AC도 목적지 문서에 각 원소(자동 모드 2플래그 설명, 카드 라우팅, stagger, 게이트 요약, 운영 상태, 진입점, 팬아웃 상한, 레인 규칙 상세)가 있는지, 포인터 줄이 그 경로를 가리키는지, 「제거 전」 순서를 검사하지 않는다. 「제거 전」은 커밋 그래프로만 증명된다(verification-claim-integrity.md §2.3). — Severity: major — Class: blocking — Required fix: AC를 추가하거나 AC-FBM-003을 확장해 (i) 원소별 대표 표지가 목적지 문서(템플릿·배포 쌍)에 존재함을 grep으로, (ii) 렌더된 포인터 줄이 그 경로 문자열을 포함함을, (iii) 목적지 작성 커밋이 메시지 테이블 축소 커밋의 조상임을 `git merge-base --is-ancestor`로 단언한다(plan.md M2/M3을 별도 커밋으로 명시).

D5. AC-004-MUTANT — acceptance.md:L47~L56, plan.md:L90 — AC-FBM-004는 `wc -c AGENTS.md` ≤ 24,276만, AC-FBM-005는 "live only under" 1건만 본다. AGENTS.md의 아무 의무 문단을 지워 크기를 맞춘 변이가 두 AC를 함께 통과한다. 변이 탐침이 기대는 plan.md §B.5 의미 대조 목록은 AC가 아니며, 원문 의무 두 개(「`main`을 통합 보유자로 쥐는 워크트리 금지」, 「`moai worktree new`로 생성」)를 목록에서 빠뜨렸다 — 운영자 패치 자체는 둘 다 보존하지만(패치 판독), 목록이 빠뜨린 의무는 변이로 지워도 아무것도 잡지 못한다. — Severity: major — Class: blocking — Required fix: AC-FBM-004 Then에 (i) `git diff --numstat 59959629908cda723c37e26c7c15718194d6b38a -- AGENTS.md` 변경이 두 문단 범위(현재 L165~L175)에 국한됨을 보이는 단일 명령 단언, 또는 (ii) 두 문단의 의무 표지 전체(`WT-*main*`, `holding \`main\``, `moai worktree new`, `.claude/worktrees/`, `agent-*`, `detach`, `live only under`)를 각각 grep으로 단언하는 절을 추가하고, §B.5 목록에 빠진 두 의무를 넣는다.

D6. BUDGET-JOINER — spec.md:L57, spec.md:L75~L76, probe 탐침 L74 — 「생산자 4,797 = 허용분 1,950 + 고정분 상한 2,847」은 허용분과 고정분 사이 결합자 `"\n\n"`(2)를 빠뜨렸다. 실측 합본 구성(이 감사 재실행: 리더 en `fixed=2573 composite=8830` → 1,950+2+2,573+2+4,303)으로 보면 core가 상한 4,092일 때 리더 칸 최대 합본은 9,148로 9,146을 2 넘는다. REQ-FBM-005가 합본을 따로 단언하므로 안전 문제는 아니고 현재 core 4,048에서는 여유 44가 남는다. — Severity: minor — Class: optional — Required fix: 고정분 상한을 2,845로 적거나, 회계 표에 결합자 2를 명시한다.

D7. LIVE-SCAN-TRUNCATION — probe/live_scan.py:L39~L41, L56 — 세션 파일의 앞 16줄만 읽고 모든 예외를 조용히 넘긴다. 훅 출력이 17번째 줄 뒤에 기록되거나 파싱이 실패한 세션은 행에서 빠지며 `overflow` 집계도 줄어든다(codex가 합성 입력으로 재현 보고, 감사자는 소스 판독으로 확인). AC-FBM-008의 「lane ≥ 6, leader ≥ 1」 조건이 일부를 막지만 넘친 세션 하나가 빠지는 경우는 못 막는다. — Severity: minor — Class: optional — Required fix: 줄 수 제한 제거, 예외 건수 출력(`errors N`), Then에 `errors 0` 추가.

D8. E6-NOT-VERBATIM — acceptance.md:L139 — E6 발췌의 `--- PASS: TestCodexContractByteCeiling (0.34s)`는 인용한 원문 파일(probe/red-now-5995962990-ceiling.txt)의 `(0.20s)`와 다르다. regression-guard 셀이라 MP-8 대상은 아니지만 「원문 그대로」 표기가 사실이 아니다. — Severity: minor — Class: optional — Required fix: 원문 줄로 교체.

D9. DECISION-ROW-PROVENANCE — decision-index.md:L11, spec.md:L88 — Q1의 `CARD-DIRECTED` 판정은 manager-spec이 작성 시점에 채운 값이다. manager-spec.md:L121은 기본값 없는 행·product-level 행을 「비워 두고 운영자에게」 보내라고 정한다. 카드 원문(`moai todo show t1923`으로 확인)의 해당 절은 「[운영자 지시 10-21]」 아래에 있고 10-22 운영자 정정도 「시작 메시지 최소화만 선택 분리 가능」으로 이 범위를 인지하므로 실질 근거는 성립한다고 판단한다. kickoff 판독기(internal/homestate/card_audit_kickoff.go:139~L146)는 비어 있지 않은 FOUNDER 판정을 받아들이므로 기계적으로는 통과한다. 다만 (i) 비표준 어휘라 Kickoff 결정 기록에서 근거 절을 다시 인용해야 하고, (ii) 카드는 「리더·레인 모두」 run id·스폰 권한 한 줄을 남기라 했는데 REQ-FBM-001(리더에 스폰 권한 없음)과 REQ-FBM-002(레인에 run id 없음)의 비대칭 해석을 다룬 결정 행이 없다, (iii) Q1은 D1의 배치 게이트 계약 대체를 다루지 않는다. — Severity: minor — Class: optional — Required fix: 비대칭 해석 행 추가, D1 해소 시 Q1 문구에 BGS 대체 여부 반영.

D10. DOD-AC008 — acceptance.md:L165~L169 — 카드의 완료 판정은 「6레인+리더 시작 조립 실측 전부 ≤10,000·overflow 안내 0」인데 DoD는 AC-FBM-008을 「라이브 확인 기록」으로만 둔다. regression-guard 분류 자체는 §2.1 판단 불가 처분으로 맞지만, 관측에서 넘침이 나오면 카드를 어떻게 처리하는지 문장이 없다. — Severity: minor — Class: optional — Required fix: 「AC-FBM-008 관측이 overflow ≥ 1이면 카드는 sync를 떠나지 않는다」 한 문장 추가.

D11. REQ-STYLE — spec.md:L79~L80 — REQ-FBM-008/009가 테스트 함수 이름·상수 이름·`t.Skip`을 요구사항 본문에 둔다(구현 세부). 테스트 인프라 SPEC이라 이해할 만하다. — Severity: minor — Class: optional — Required fix: 선택.

## 회귀 확인

첫 반복이라 해당 없음.

## 권고 (manager-spec 수정 지시)

1. D1: 배치 게이트 문장 유지 또는 REQ-BGS-016/017 명시적 대체 중 하나를 고르고(운영자 판단이 필요하면 decision-index에 빈 판정의 FOUNDER 행으로), 고정 테스트 3개 파일을 plan.md §A와 REQ-FBM-009에 넣는다.
2. D2: 탐침과 AC-FBM-001에 필수 원소 전체 존재 + 제거 원소 표지 부재를 두 채널 모두로 단언한다.
3. D3: AC-FBM-002에 출력 형식·칸 수·경계 변이 실패 관측을 넣는다.
4. D4: 재배치 목적지 존재·포인터 경로·커밋 순서를 단언하는 AC를 추가한다.
5. D5: AGENTS.md 변경 범위 또는 의무 표지 전체를 단언하는 절을 AC-FBM-004에 넣는다.
6. 선택: D6~D11.

## 증거 기반 보고 (5절 형식)

### Claim

- C1. 카드의 23,145자 넘침 원인 분석과 현재 트리 수치(리더·레인 언어별 고정분, lane_en_glm 9,946, 탐침 10,006)는 재현된다.
- C2. AGENTS.md는 24,501바이트이고 Codex 상한 테스트 2개는 녹색이며(여유 75), 운영자 패치는 현재 트리에 그대로 적용되지 않는다(exit 1). 패치 문단을 옮기면 약 24,217~24,218바이트, `only` 복원 시 24,220~24,221바이트다.
- C3. 계획된 `only` 복원은 원래 배타성 의미를 되살린다. 패치의 나머지 문장은 원문 의무를 보존한다.
- C4. release-blocking AC 4개의 RED-now 셀은 재현되나, 네 개 모두 요구사항을 어기면서 통과하는 변이를 막지 못한다(D2~D5).
- C5. 이 SPEC은 완료 SPEC-AUTONOMY-BATCH-GATE-001의 결정된 요구사항과 충돌한다(D1).

### Evidence (명령과 원문 출력 — 이 감사에서 실행)

```text
$ go test -count=1 -overlay .moai/specs/SPEC-FACTORY-BOOT-MINIMAL-001/probe/overlay.json -run '^TestT1923BootBudgetProbe$' -v ./internal/hook/   (출력은 scratch 파일로 리다이렉트)
exit=1
grep -c -- "--- FAIL" → 25 ; grep -c -- "^--- FAIL" → 1
    zz_t1923_boot_probe_test.go:123: PROBE swept=72
   2     zz_t1923_boot_probe_test.go:84: factory fixed part 2852 > ceiling 2847
   2     zz_t1923_boot_probe_test.go:84: factory fixed part 3751 > ceiling 2847
   2     zz_t1923_boot_probe_test.go:87: composite 10006 > budget 9146
grep -c "auto-mode explanation" → 48
FAIL	github.com/modu-ai/moai-adk/internal/hook	1.942s
diff <기록된 PROBE 73줄> <재실행 PROBE 73줄> → 차이 없음 (exit 0)
```

재실행 칸 발췌(배포 면):

```text
leader//en/startup/deployed fixed=2573 composite=8830 core=4048
leader//ko/startup/deployed fixed=1650 composite=7907 core=4048
leader//ja/startup/deployed fixed=1651 composite=7908 core=4048
leader//zh/startup/deployed fixed=1229 composite=7486 core=4048
lane//en/startup/deployed fixed=3751 composite=10006 core=4048
lane//ko/startup/deployed fixed=2835 composite=9090 core=4048
lane//ja/startup/deployed fixed=2852 composite=9107 core=4048
lane//zh/startup/deployed fixed=2266 composite=8521 core=4048
lane//en/clear/deployed fixed=2713 composite=8968 core=4048
lane/manual/en/startup/deployed fixed=2142 composite=8397 core=4048
```

```text
$ go test -count=1 -run '^(TestRoleInjectionAssemblyBudget|TestCompositeBudgetRealAssembly)$' -v ./internal/hook/
exit=0
    role_composite_e2e_test.go:90: RESULT lane_ko_glm: composite utf16=9030 (cap 10000, margin=970)
    role_composite_e2e_test.go:90: RESULT lane_en_glm: composite utf16=9946 (cap 10000, margin=54)
    role_composite_e2e_test.go:90: RESULT leader_ko: composite utf16=7882 (cap 10000, margin=2118)
    role_composite_e2e_test.go:90: RESULT leader_en: composite utf16=8806 (cap 10000, margin=1194)
    role_injection_budget_test.go:163: deployed tree: role factory-lane composite=9100 (budget 9146) core=4048 (ceiling 4092, target 3800)
ok  	github.com/modu-ai/moai-adk/internal/hook	1.083s
```

```text
$ go test -count=1 -run '^(TestCodexContractByteCeiling|TestCodexNestedTemplateDiscoveryBudget)$' -v ./internal/config/
exit=0
    token_budget_guard_test.go:109: contract document AGENTS.md = 24501 bytes (ceiling 24576, headroom 75)
    token_budget_guard_test.go:109: contract document internal/template/templates/AGENTS.md.tmpl = 23784 bytes (ceiling 24576, headroom 792)
--- PASS: TestCodexContractByteCeiling (0.08s)
--- PASS: TestCodexNestedTemplateDiscoveryBudget (0.12s)
ok  	github.com/modu-ai/moai-adk/internal/config	0.541s
```

```text
$ go test -count=1 -run '^(TestSD_AC019_NextCardRuleInjection|TestLaneRuleCarriesStandingRecheckCron|TestFactoryWorkerNoticeCarriesSpawnAuthority|TestInjectedCoreCarriesObligationElements)$' -v ./internal/hook/
exit=0
--- PASS: TestLaneRuleCarriesStandingRecheckCron (1.08s)
--- PASS: TestFactoryWorkerNoticeCarriesSpawnAuthority (0.00s)
--- PASS: TestInjectedCoreCarriesObligationElements (0.00s)
--- PASS: TestSD_AC019_NextCardRuleInjection (0.70s)
ok  	github.com/modu-ai/moai-adk/internal/hook	2.367s
(작성자가 Gap으로 남긴 AC-FBM-009 기준 관측을 이 감사가 채움 — 4/4 실행, 빈 순회 아님)
```

```text
$ wc -c AGENTS.md                 →    24501 AGENTS.md   (exit 0)
$ grep -c "live only under" AGENTS.md → 1 (exit 0)
$ grep -c autoModeGuide internal/hook/session_start_factory_i18n.go → 11 (exit 0)
$ git apply --check /tmp/adk-pr-recovery-20261021-xlr15jib/proposal-1862-byte-ceiling/BYTE-CEILING-PROPOSAL.patch
error: patch failed: AGENTS.md:164
error: AGENTS.md: patch does not apply
exit=1
$ git cat-file -s 3f81377dc9:AGENTS.md → 24807
$ python3 -I <scratch>/cand.py AGENTS.md <patch>
head 24501 old_block 927 new_block 643 cand 24217 cand_only 24220
live-only-under count in cand_only: 1
```

(작성자 수치 644/24,218/24,221과 1바이트 차이 — 문단 끝 빈 줄 처리 차이로 보이며 결론(≤ 24,276)에 영향 없음.)

```text
$ sh <scratch>/trace.sh spec.md acceptance.md
COLLECTED: 11 REQ definitions (acceptance input: read)
$ sh <scratch>/cn4.sh plan.md acceptance.md
COLLECTED: 6 milestones in plan order (M1 M2 M3 M4 M5 M6), 0 exit bindings, 0 ordering candidates
NONE: 0 records in …/acceptance.md match (before|after|first|prior to|pre-change) (case-insensitive) — an observed absence
$ sh <scratch>/d7d8.sh spec.md
D7 SPEC-FACTORY-BOOTSTRAP-001 status=completed
D7 SPEC-FACTORY-SELF-DISPATCH-001 status=completed
D7 SPEC-ROLE-INJECTION-BUDGET-001 status=completed
D7 SPEC-SESSION-START-GUIDE-I18N-001 status=completed
--- D8
D8 done
```

```text
$ grep -n "REQ-BGS-016\|REQ-BGS-017" .moai/specs/SPEC-AUTONOMY-BATCH-GATE-001/spec.md
159:- **REQ-BGS-016** (Ubiquitous) — The SessionStart notice of the factory leader and of the kanban leader shall each carry one sentence that carries the name `batch gate summary` verbatim and points at its doctrine home in the auto-semantics rule by path and section, …
160:- **REQ-BGS-017** (Ubiquitous) — That sentence shall be present in the English agent-facing copy of each leader notice and in the operator-facing copy of each leader notice in each of the en, ko, ja, and zh locales, … [Q5 decided]
$ grep -n "^status:" .moai/specs/SPEC-AUTONOMY-BATCH-GATE-001/spec.md → 5:status: completed
$ grep -n "AUTONOMY-BATCH-GATE\|BGS-017" .moai/specs/SPEC-FACTORY-BOOT-MINIMAL-001/*.md → (출력 없음)
$ grep -rln "leaderStagger|…|gateSummary" internal --include='*_test.go'
internal/hook/session_start_guide_i18n_render_test.go
internal/hook/session_start_factory_worker_test.go
internal/hook/session_start_factory_test.go
```

```text
$ grep -n "DEFAULT-APPLIED\|Operator verdict\|founder" internal/homestate/card_audit_kickoff.go
139:		if strings.HasPrefix(verdict, "DEFAULT-APPLIED") && (class != "implementation-level" || !hasDefault) {
146:			return fmt.Sprintf("FOUNDER row %q has an empty verdict", head)
→ Q1(product-level, "CARD-DIRECTED …")은 비어 있지 않아 거부되지 않음. Q2~Q7은 implementation-level + Default + DEFAULT-APPLIED로 허용 조건 충족.
```

교차 모델 감사 (`mcp__moai__audit_multi`, project_root=이 워크트리, target=baseBranch, gates 인자 없음):

```text
overall_verdict: fail (codex, gate required, verdict fail), participant_count 1, fail_open_backends [], plan_source config, build_commit b1806498fe
audit_receipt: rcpt-873f4860189236ef74793036
codex P1: 배치 게이트 계약 충돌 / t85 테스트 누락 / AC-001 허용 목록 미검사 / AC-002 보고만 하는 변이 / AC-003 삭제 변이 / AC-002 RED가 실패를 재현하지 않음
codex P2: AC-004·005 의무 삭제 변이 / 결합자 2 누락 / live_scan 16줄 절단
$ moai verify audit-plan --project-root <toplevel> --result-file <toplevel>/.moai/state/audit-plan-result.json
"convergence_check": { "ok": true, "unmet": [] }   (게이트는 충족 — 필수 백엔드가 응답했고 그 판정이 fail)
```

codex 지적의 처분: 9건 중 8건은 감사자가 직접 재현·판독으로 확인해 D1~D7에 반영했다. 「AC-002 RED가 실패를 재현하지 않음」은 MP-8 결함으로 보지 않았다(기준의 Then이 요구하는 출력이 기준선 stdout에 없으므로 기준 자체는 빨강) — 그 약점은 D3에 흡수했다.

### Baseline-attribution

- 측정 트리: HEAD `b805ac2d2429f414a7b3634d4196711d688289be`. `git diff --name-only 59959629908cda723c37e26c7c15718194d6b38a..HEAD`에서 SPEC 디렉터리 밖 파일 0건(grep -v 결과 없음, exit 1) — 즉 기록된 RED-now 기준 트리 `5995962990`과 Go·AGENTS.md 내용이 같다. 그래서 재실행 결과를 기록과 직접 비교할 수 있었다.
- Go 측정은 이 트리의 소스를 `go test`로 직접 컴파일해 얻었다(설치 바이너리 미사용). 설치된 `moai`와 MCP 서버의 빌드 커밋은 `b1806498fe`(v3.2.0-rc.38, 2026-10-21T15:27Z)이고, `git merge-base --is-ancestor b1806498fe HEAD` exit 1 — HEAD의 조상이 아니므로(갈라진 브랜치) §2.2의 「엄격한 조상」 지연은 없다. 이 빌드가 판단한 것은 `moai verify audit-plan`, `moai slot`, `moai todo show`, audit_multi 수집뿐이다.
- 부하: Go 실행 전 load 99→18.8(1분 평균)로 내려간 뒤 `moai slot acquire --resource go-heavy --max-duration 15m`(exit 0)로 임대, Go 호출 4회를 한 번에 하나씩 실행, `moai slot release --resource go-heavy`(exit 0).

### Gaps

- 라이브 세션 재스캔(AC-FBM-008, `live_scan.py`)은 실행하지 않았다 — 사용자 프로필 디렉터리 읽기이며 착지 전이라 판정 대상이 아니다. 23,145자 분해(plan.md §B.3)는 기록 파일 판독만 했고 재실행하지 않았다.
- `probe/research/measure_t1923_test.go.txt`의 필드별 측정(FIELDS 줄)은 재실행하지 않았다. 대신 AC 탐침 72칸 재실행으로 역할·언어·소스별 합계를 확인했다.
- codex가 보고한 live_scan 합성 입력 재현과 AGENTS.md 변이(24,134바이트)는 감사자가 실행으로 재현하지 않고 소스·패치 판독으로 확인했다.
- 템플릿 측 재배치 목적지 문서는 아직 정해지지 않아(Q5 기본값이 후보만 지명) 검사할 대상이 없다.

### Residual-risk

- 비-팩토리 허용분 1,950은 GLM 레인 1개 세션의 분해에 기댄다. 여러 경고가 겹치는 세션은 이를 넘을 수 있으며 REQ-FBM-005는 합성 허용분으로만 단언한다.
- 레인 ko startup 고정분 2,835는 상한 2,847에 12만 남아, 레인 규칙 축소 중 문구가 조금만 늘어도 한계에 닿는다.
- D1의 해소 방식(유지 vs 대체)에 따라 리더 공지 크기와 허용 원소 목록이 달라지므로, 수정 뒤에는 D2 탐침 단언도 함께 다시 봐야 한다.

## 운영 메모 (미검증)

- 측정 지시 — 수정 후 `go test -count=1 -overlay …/probe/overlay.json -run '^TestT1923BootBudgetProbe$' -v ./internal/hook/`를 다시 돌려 강화된 단언이 기준 트리에서 빨강임을 관측할 것. 상태: assumption(수정 내용이 아직 없음).
- 측정 지시 — D1을 대체로 해소한다면 `go test -run '^TestLeaderNoticeBatchGatePointer$' ./internal/hook/`의 다시 쓴 계약이 기준 트리에서 빨강인지 관측할 것. 상태: inferred(현재 테스트는 문장 존재를 요구하므로 대체 계약은 기준 트리에서 빨강이어야 한다는 추론).

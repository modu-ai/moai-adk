# SPEC Review Report: SPEC-RESOURCE-SLOT-LEASE-001

Iteration: 1/2 (Tier M 상한 2 — `harness.plan_audit_tier_ceilings`)
Verdict: **FAIL**
Overall Score: **0.78** (Tier M 통과 기준 0.80 미달)

감사 대상 트리: 워크트리 t607, 브랜치 `WT-heavy-test-slot`, HEAD `1a178c174`. SPEC 산출물의 문서 핀 `c4ce42eca`는 HEAD의 부모이고, 두 커밋의 차이는 SPEC 디렉터리 네 파일뿐이다(`git diff --stat c4ce42eca HEAD` → 4 files, 706 insertions, 모두 `.moai/specs/SPEC-RESOURCE-SLOT-LEASE-001/`). 따라서 `internal`·`cmd`·`pkg`에 대한 원장 측정은 HEAD에서 다시 재도 같은 트리를 잰다.

작성자 추론 맥락은 받지 않았다(M1 맥락 격리). 입력은 spec.md, plan.md, acceptance.md, progress.md와 운영자 결정 기록 `.moai/reports/t607/verdict.md`, 그리고 SPEC이 기대는 코드다.

적용한 정책 규칙: `verification-claim-integrity.md` §1·§2·§2.2, `verification-completeness.md` §1.1·§1.3·§2, `.claude/rules/local/gitflow-lane-protocol.md` §8.

---

## 결론 요약

뼈대는 튼튼하다. 운영자 결정표 일곱 줄은 모두 반영됐고, 필수 통과 기준 일곱 개도 모두 통과한다. 원장 EL-1..EL-6은 다시 재도 기록과 같게 나온다. 떨어진 이유는 형식이 아니라 **설계 정확성** 쪽 결함 네 건이다.

1. 가드가 기록을 찾을 루트를 CLI와 같은 방식으로 푼다는 전제(§B.3)가 코드와 맞지 않는다. 가드가 워크트리 쪽 빈 디렉터리를 읽으면, 막아야 할 레인에서 아무 신호 없이 무력해진다.
2. 대조군 AC-RSL-001의 끼어들기 구성은 올바른 직렬화 아래에서 교착하고, 그 결과는 AC가 요구하는 "보유 중" 대신 "busy"다.
3. 변경 락 경합("busy")이라는 결과가 요구사항에 아예 없다.
4. AC-RSL-013(c)와 AC-RSL-015의 `BASELINE_SHA` 리터럴 핀은 이 저장소의 lane 규칙 §8이 금지한 형태다. develop을 흡수하는 순간 거짓 실패를 낸다.

---

## Must-Pass Results

- [PASS] **MP-1 REQ 번호 일관성** — spec.md:63-108에 REQ-RSL-001부터 016까지 빈칸·중복 없이 이어지고, 세 자리 0채움도 일관된다. 요구사항 층에서 판정했다.
- [PASS] **MP-2 GEARS 형식** — 요구사항 층(spec.md §C)에서 판정했다. 정본 영어 문장 16개가 모두 GEARS 패턴과 맞는다. Ubiquitous 형식은 001·003·013·015(예: spec.md:63 "The slot lease surface shall provide …"), Event-driven은 002·005·006·007·008·009·014·016(예: spec.md:66 "**When** an acquire is requested …"), State-driven은 004(spec.md:72 "**While** a different session …"), Unwanted 정본형은 010(spec.md:90 "shall not accept"), Where+When 복합형은 011(spec.md:93)이다. 012는 Where 형식(spec.md:96)이다. IF/THEN 폐기형은 없다. 인수 기준 층(acceptance.md §D.2)의 Given-When-Then은 검증 층 형식이므로 여기서 감점하지 않았다.
- [PASS] **MP-3 프론트매터** — spec.md:2-13에 12개 정식 필드가 모두 있다: `id`, `title`, `version: "0.1.0"`(따옴표 있음), `status: draft`, `created`/`updated: 2026-09-12`, `author`, `priority: P1`, `phase: "v3.2.0 target"`(생애 단계명이 아님), `module`, `lifecycle: spec-anchored`, `tags`(쉼표 구분 문자열). 거부 별칭은 없고, 추가 필드는 `tier: M`이다. 트리에서 빌드한 바이너리의 lint가 발견 사항 0건을 냈다(Evidence E6).
- [PASS] **MP-4 언어 중립성** — 템플릿에 들어갈 내용을 판정했다. REQ-RSL-013(spec.md:99)은 "a built-in list of commands of any programming language"를 금지하고, REQ-RSL-015(spec.md:105)는 언어를 가리키지 않는 자리표시자만 허용한다. 16개 언어 중 일부만 늘어놓은 곳은 없다. 다만 이 의무를 기계적으로 확인하는 인수 기준이 없다(D6).
- [PASS] **MP-5 D7 교차 SPEC** — 참조하는 SPEC 셋(`SPEC-INTEGRATION-LOCK-ATOMIC-001`, `SPEC-INTEGRATION-LOCK-LIVENESS-001`, `SPEC-SYNC-SHA-SLOT-FORMAT-001`)이 모두 `status: completed`다. retired·superseded·archived는 없어 BLOCKING이 없다(E7).
- [PASS] **MP-6 D8 교차 플랫폼** — 세 파일의 `syscall` 등장 수는 각각 0이라 자동 통과다(E7). Windows 변경 락 잔재 문제는 spec.md:121, :144에서 스스로 밝히고 있다.
- [PASS] **MP-7 명확화 게이트** — `grep -rn 'NEEDS CLARIFICATION'`이 SPEC 디렉터리 전체에서 종료 코드 1로 끝났다. research.md는 Tier M이라 없다(E7).

## Operator Decision Conformance (verdict.md 결정표 대조)

| 결정 항목 | SPEC 반영 위치 | 판정 |
|---|---|---|
| 범용 임대 명령 `moai slot acquire\|status\|release --resource` | spec.md:34, REQ-RSL-001(:63), plan.md:115 | 준수 |
| 선택형 PreToolUse 가드, 기본 꺼짐 | REQ-RSL-011/012(:93, :96), Exclusions "가드 기본값 켜기"(:173-174) | 준수 |
| 템플릿과 바이너리로 배포, 템플릿 문서 최소 | REQ-RSL-015(:105), plan.md M5(:127-132) | 준수. 언어 중립성 검증은 빠져 있다(D6) |
| 기록 필드(자원·세션 id/이름·pid·명령·시작 시각·선언 상한) | REQ-RSL-002(:66), AC-RSL-003 | 준수. 선택형 플래그를 생략할 때 무엇을 기록하는지는 정해져 있지 않다(D11) |
| 두 세션 대조군 필수(표면 없음 → 둘 다 시작, 있음 → 하나 거절) | AC-RSL-001, AC-RSL-002 | 의도는 준수. 구성은 결함(D2) |
| kanban 락의 코드 방식만 재사용, `moai integration`과 분리 | REQ-RSL-001, §D(:113), plan.md B2(:36-40), AC-RSL-013 | 준수 |
| 두 문서 편집은 `merge-base --is-ancestor … develop` == 0일 때만 | REQ-RSL-016(:108), plan.md M6(:134-138), AC-RSL-015 | 준수. 두 가지(열림·닫힘) 모두 정의돼 있고 M1-M5만으로 닫을 수 있다(acceptance.md:328). 닫힘 가지의 판정식은 결함(D4) |

## Category Scores

| Dimension | Score | Rubric Band | Evidence |
|---|---|---|---|
| Clarity | 0.75 | 0.75 | 요구사항은 대부분 해석이 하나다. 예외는 셋이다. REQ-RSL-014의 "unavailable configuration"이 REQ-RSL-012와 부딪힌다(D5). REQ-RSL-002의 선택형 필드는 생략 시 동작이 없다(D11). §B.3의 루트 해석 동치 주장이 코드와 다르다(D1). |
| Completeness | 1.0 | 1.0 | HISTORY §H(:147-151), 배경 §A, 요구사항 §C, 방법 plan.md, 인수 기준 acceptance.md, Exclusions H3 7개(:155-174)가 모두 있고 각각 `-` 항목을 가진다. 프론트매터 12필드도 완전하다. |
| Testability | 0.70 | 0.75에 가까우나 두 건이 해석을 요구함 | AC-RSL-001 구성이 교착해 기대 결과("보유 중")에 닿을 수 없다(D2). AC-RSL-013(c)와 AC-RSL-015 닫힘 가지는 흡수 후 거짓 실패한다(D4). 나머지 AC는 명령과 기대 출력이 이진 판정 가능하고, 뮤턴트 표(AC-RSL-005/006/011)는 판별력이 있다. |
| Traceability | 0.75 | 0.75 | 역방향(§D.3, acceptance.md:292-311)은 REQ 16개 모두에 AC가 있고 고아 AC도 없다. 다만 REQ-RSL-015의 언어 중립 절(D6)과 REQ-RSL-014의 "unavailable configuration" 절(D5)을 검증하는 AC가 없다. 경합(busy) 결과는 요구사항 자체가 없다(D3). |

집계: (0.75 + 1.0 + 0.70 + 0.75) / 4 = **0.80**에서 결함 D1(설계 전제 오류, 필수 통과 외 blocking)을 반영해 **0.78**로 내렸다. Tier M 기준 0.80 미달이다.

## Defects Found (structured defect-list)

D1. GUARD-ROOT-DIVERGENCE — spec.md:47, plan.md:38 — §B.3은 "훅 쪽도 같은 순서로 프로젝트 루트를 푼다(`internal/hook/pre_tool.go`:365-401)"고 쓰고, plan B2는 "가드는 훅의 프로젝트 루트 해석을 그대로 쓴다"를 근거로 "모든 워크트리가 한 기록을 본다"고 결론낸다. 코드는 다르다. 훅은 `CLAUDE_PROJECT_DIR` → `os.Getwd()`만 쓰고 git common dir 단계가 없다(`internal/hook/path_resolve.go`:78-94, 주석 :48 "Resolution semantics are unchanged: CLAUDE_PROJECT_DIR, then os.Getwd()"). CLI는 `CLAUDE_PROJECT_DIR` → `git rev-parse --git-common-dir`의 부모 → cwd 순서다(`internal/cli/integration.go`:46-65). 이 세션에서 잰 값: Bash 환경의 `CLAUDE_PROJECT_DIR`는 비어 있다(E4). 그래서 CLI는 git common dir를 거쳐 primary에 쓴다. 반면 훅 쪽의 `h.projectRoot()`는 이 세션에서 primary로 풀렸다(E4, `agent-model-audit.jsonl`이 primary에 기록됨). 이 일치는 세션을 시작한 방식에 달린 우연이다. 훅 환경의 `CLAUDE_PROJECT_DIR`가 링크된 워크트리를 가리키는 세션에서는 가드가 워크트리의 빈 `.moai/state/slot-leases/`를 읽고 "보유자 없음"으로 허용한다. plan B1(:32)에 따르면 이 경로는 표준 오류 없이 감사 한 줄만 남긴다. 결국 가드가 필요한 레인에서 가드가 조용히 꺼진다(`verification-completeness.md` §1.3 "비실행이 성공과 구별되지 않는 검사"). 이 형태의 세션에서 훅 루트가 실제로 워크트리로 풀리는지는 재지 못했다(Gaps). 하지만 SPEC이 증거로 댄 동치 주장은 코드 사실로서 거짓이다. — Severity: major — Class: **blocking** — Required fix: (a) §B.3의 동치 문장을 코드와 맞게 고친다. (b) 가드가 기록을 찾는 루트를 CLI와 같은 공유 루트로 정규화한다는 요구사항을 추가한다. 예를 들어 훅이 푼 루트에서 git common dir의 부모를 다시 구하고, 구할 수 없으면 advisory와 함께 fail-open한다. (c) AC를 추가한다: 훅 프로젝트 루트가 링크된 워크트리이고, 기록이 primary에 있고, 살아 있는 다른 보유자가 있을 때 → 거부. 여기에 대조 행으로 "루트 정규화를 지운 뮤턴트는 허용한다"를 둔다.

D2. CONTROL-TEST-DEADLOCK — acceptance.md:91-93 (AC-RSL-001 Given/When), plan.md:99-100 — Given은 "부모는 B가 자기 동작을 끝낸 뒤 A를 풀어 준다"고 쓴다. 표면 갈래에서 A는 "판정과 쓰기 사이", 곧 변경 락 임계 구역 안에서 멈춘다. 올바르게 직렬화됐다면 B는 변경 락을 기다리느라 끝날 수 없다. 대기 예산(`boardLockWaitBudget` = 1.65s, `integration_lock_mutation.go`:91-97)이 다하면 B는 "보유 중"이 아니라 busy를 돌려준다(:126). Then은 "다른 한 자식이 '보유 중' 오류로 거절"이라서 올바른 구현이 이 AC를 통과할 수 없다. 코드 방식의 출처인 선행 테스트는 이 교착을 500ms stall-release 타임아웃(예산의 30.3%)으로 풀고, busy를 하네스 설정 오류로 판정한다(`internal/kanban/integration_lock_cross_test.go`:46-63, :272). SPEC은 이 파일을 참조로만 들고(plan.md:156) AC 문장에는 그 구성을 담지 않았다. — Severity: major — Class: **blocking** — Required fix: AC-RSL-001의 Given/When을 다시 쓴다. "B의 완료 또는 대기 예산의 1/3 이하인 stall-release 타임아웃 중 먼저 오는 쪽에 A를 푼다"로 하고, 결과가 busy면 하네스 결함으로 실패한다고 명시한다. 두 자식이 기록하는 소유자 pid는 테스트 동안 살아 있는 pid(부모 테스트 프로세스)로 고정한다고 적는다. 그렇지 않으면 스테일 인수로 이중 성공이 정당하게 난다(선행 테스트 :180-183 주석).

D3. BUSY-OUTCOME-MISSING — spec.md:69-73 (REQ-RSL-003/004), acceptance.md §D 행렬 — 변경 락 경합이 예산을 넘는 경우의 결과가 요구사항에도 AC에도 없다. 통합 창은 `ErrIntegrationLockBusy`를 `ErrIntegrationLockHeld`와 구별되는 일시 오류로 두고(`integration_lock_mutation.go`:43-58, "Reporting one as the other tells a lane a false thing about the board"), 전용 AC(AC-ILA-004, `TestIntegrationLockBusy_IsNotHeld`, cross_test.go:350)로 판별한다. 이 SPEC의 REQ-RSL-004는 "held error"만 정의한다. 그러면 busy를 held로 보고하는 구현도 모든 AC를 통과한다. — Severity: major — Class: **blocking** — Required fix: 요구사항 하나를 추가하거나 REQ-RSL-004를 확장한다. 내용은 "변경 락이 대기 예산 동안 경합하면 획득·해제는 held와 구별되는 일시(busy) 오류를 돌려주고 기록을 바꾸지 않는다"이다. 대응 AC로 "busy 판정 함수가 참이고 held 판정 함수가 거짓"을 양방향으로 확인한다. Tier M 상한이 16이므로 REQ-RSL-004에 흡수하는 쪽이 예산 안에서 해결하는 길이다.

D4. LITERAL-BASELINE-PIN — acceptance.md:247 (AC-RSL-013 넷째 명령), :274 (AC-RSL-015 닫힘 가지), :323 (§D.5) — 두 판정식 모두 `BASELINE_SHA`를 "run 단계 첫 커밋 전에 `git merge-base develop HEAD`로 잡아 기록한 값"으로 고정해 `git diff "$BASELINE_SHA" HEAD -- <files>`를 잰다. 이 저장소의 lane 규칙 `.claude/rules/local/gitflow-lane-protocol.md` §8은 이를 명시적으로 금지한다. 요지는 "리터럴 base SHA로 재지 않는다 — merge-base를 읽는 시점에 다시 구한다"이고, 이유는 "흡수하는 순간 리터럴 핀 범위에 다른 카드의 커밋이 들어온다"이다. 병합 창 절차는 병합 전에 레인 워크트리가 로컬 develop을 흡수하도록 되어 있다(CLAUDE.local.md §4.1). 그래서 흡수한 develop에 다른 카드가 `kanban-dispatch.md`나 `internal/cli/integration.go`를 고친 커밋이 있으면 두 판정식이 이 SPEC이 하지 않은 변경으로 실패한다. `kanban-dispatch.md`는 자주 고쳐지는 규칙 파일이다. 게다가 이 워크트리가 기록한 develop(`eb50af5a8`)은 감사 시점에 이미 `e82ef5565`로 움직였다(E5). — Severity: major — Class: **blocking** — Required fix: 두 판정식을 lane 규칙 §8의 형태로 바꾼다. 읽는 시점에 `CARD_BASE=$(git merge-base develop HEAD)`를 구하고, 대조군 `git diff --name-only "$CARD_BASE"..HEAD | wc -l` ≥ 1을 둔다(0이면 "측정 불가"). 판정은 병합 전에만 유효하다고 명시한다. 병합 뒤의 근거는 병합 트리와 카드 트리의 동일성으로 대신한다. §D.5의 "`BASELINE_SHA`가 첫 run 커밋 전에 잡혀 기록돼 있다"도 같이 고친다.

D5. REQ-014-012-CONFLICT — spec.md:96 (REQ-RSL-012), spec.md:102 (REQ-RSL-014), acceptance.md:226-235 — REQ-RSL-014는 "unavailable configuration"을 불확실성으로 보고 표준 오류 advisory와 감사 한 줄을 요구한다. REQ-RSL-012는 키가 없거나 거짓이면 "read no lease record and deny no command"이고, plan B4(:57)는 꺼진 경로에서 감사 로그도 쓰지 않는다고 한다. 설정을 읽을 수 없으면 `enabled` 값 자체를 알 수 없다. 기존 관례는 nil 설정을 꺼짐으로 읽고 조용히 넘어간다(`pre_tool.go`:821-830 `integrationLockEnabled`). 설정이 없는 모든 프로젝트에서 매 Bash 호출마다 advisory를 내는 해석과, 조용한 해석이 둘 다 요구사항에 부합한다. AC-RSL-012의 네 사례(a-d)에도 이 경우는 없다. — Severity: minor — Class: **blocking** (내부 모순, CN-1) — Required fix: REQ-RSL-014에서 "unavailable configuration"을 빼거나, "enabled가 참으로 읽혔지만 resources 구획이 해석 불가"처럼 켜진 뒤의 불확실성으로 좁히고 AC-RSL-012에 그 사례를 추가한다.

D6. NEUTRALITY-UNVERIFIED — spec.md:105 (REQ-RSL-015), acceptance.md:253-267 (AC-RSL-014) — REQ-RSL-015는 자리표시자가 "names no programming language"일 것을 요구한다. 그런데 AC-RSL-014가 드는 기계 검사는 이 성질을 보지 않는다. `TestTemplateNeutralityAudit`의 클래스는 C1·C2·C4·C5·C6·C9(`internal/template/template_neutrality_audit_test.go`:126-183)로 경로 편향·서사·메모리·로컬 참조·PR 번호·자연어 정본형뿐이다. CI 워크플로도 이 테스트와 유출 테스트만 돈다(`.github/workflows/template-neutrality-check.yaml`:58-89). 새 grep(acceptance.md:261)은 SPEC ID·카드 id·날짜만 본다. 특정 언어의 테스트 명령을 예시로 실은 구현도 AC-RSL-014를 통과한다. — Severity: minor — Class: **blocking** (REQ 한 절이 추적되지 않음) — Required fix: AC-RSL-014에 판정 하나를 추가한다. 템플릿 `workflow.yaml`의 `slot_lease` 블록과 새 규칙 파일에 대해, 16개 언어의 대표 도구 토큰 목록(명시 열거)을 grep해 출력이 없어야 한다. 짝으로 그 목록 중 하나를 넣은 임시 변형에서 적중을 관측하는 양성 대조를 둔다. 또는 `resources: {}`와 언어 중립 자리표시자 토큰(예: `<command-regex>`)을 구조로 단언한다.

D7. OQ1-MEASURABLE-NOW — plan.md:88 (OQ-1), spec.md:139 — OQ-1은 run 단계 M4로 미뤘지만 지금 잴 수 있고, 이 감사에서 쟀다(E3). 서브에이전트(이 감사자)의 Bash 호출이 일으킨 PreToolUse 기록의 `session_id`는 부모 세션 id `432c40c9-…`였다. 서브에이전트 Bash 환경의 `CLAUDE_CODE_SESSION_ID`도 같은 값이다. 즉 현재 런타임에서는 보유자의 서브에이전트가 보유자로 읽힌다. §G의 우려(불필요한 거부)는 이 런타임에서 일어나지 않는다. 대신 SPEC이 적지 않은 반대쪽 한계가 생긴다. 한 세션이 서브에이전트 여럿에게 무거운 명령을 동시에 돌리게 하면 모두 "보유자 자신"으로 통과해 같은 세션 안의 겹침은 막히지 않는다. — Severity: minor — Class: optional — Required fix: OQ-1을 이 측정으로 닫거나(명령, 출력, 런타임 버전을 기록), M4 재측정을 유지하되 지금의 관측을 기준선으로 적는다. "세션 내 병렬 실행은 직렬화하지 않는다"를 Exclusions나 §G에 한 줄 추가한다.

D8. FACTUAL-SLIP-PRIVATE — plan.md:40 — "`FactoryProcessAlive`이 이 패키지의 비공개 함수"라고 썼지만 이 함수는 공개 함수다(`internal/kanban/factory_alive_unix.go`:22, `factory_alive_windows.go`:28). 배치 결론(`internal/kanban`에 둔다)은 `acquireBoardLockImpl`과 `boardLockWaitBudget`이 비공개라는 사실만으로도 유지된다. — Severity: minor — Class: optional — Required fix: 문장을 고친다.

D9. DATED-DEVELOP-SHA — spec.md:57, acceptance.md:47 — "`develop`(`eb50af5a8`)"을 현재형으로 적었다. 감사 시점의 로컬 develop은 `e82ef5565`다(E5). 게이트 판정(종료 코드 1)은 그대로다. 게이트 명령은 움직이는 ref를 대상으로 삼는 SUBJECT형 주장이므로 옳다. 다만 함께 적힌 SHA는 측정 시점 값이라고 표시해야 한다. — Severity: minor — Class: optional — Required fix: "측정 시점(트리 `c4ce42eca`) 값"이라고 날짜와 트리를 붙인다.

D10. TEST-ENV-SCRUB — spec.md:119, plan.md:82 — 테스트 격리 제약은 `CLAUDE_PROJECT_DIR` + `GIT_CEILING_DIRECTORIES`만 고정한다. 그런데 레인 환경에는 `CLAUDE_CODE_SESSION_ID`와 `MOAI_SESSION_PID`가 설정돼 있다(이 세션에서 둘 다 값이 있음, E4). CLI 세션 id 해석(`integration.go`:80-92)과 소유자 pid 해석(`session_pid.go`:82-90)이 이 둘을 먼저 읽는다. 따라서 CLI 테스트가 실제 세션 id와 pid를 기록에 박을 수 있다. — Severity: minor — Class: optional — Required fix: 두 환경변수를 테스트가 고정하거나 비우는 목록에 추가한다.

D11. OPTIONAL-FIELDS-UNSPECIFIED — spec.md:66 (REQ-RSL-002), plan.md:115 — REQ-RSL-002는 세션 이름과 명령을 반드시 기록할 필드로 나열한다. 그런데 CLI의 `--name`과 `--command`는 선택형이다. 생략했을 때 빈 값을 기록하는지, 거부하는지, 대체값을 쓰는지 정해져 있지 않다. 운영자 결정표의 "명령" 필드는 스테일·인수 판단에서 사람이 읽는 값이라, 비어 있으면 결정의 취지가 약해진다. — Severity: minor — Class: optional — Required fix: 생략 시 동작을 한 문장으로 정한다(예: 빈 값 허용 + `status`가 "명령 미기재"를 표시). AC-RSL-003에 생략 행을 하나 둔다.

D12. ALLOW-BRANCH-ATTRIBUTION — acceptance.md:208-222 (AC-RSL-011) — 허용 행 중 감사 사유를 단언하는 행은 n-held 하나뿐이다. 허용으로 가는 경로(자기 자신, 스테일, 만료, 보유 없음)가 여럿이라, 한 경로의 검사를 지워도 다른 경로로 허용되는 뮤턴트가 생길 수 있다. 예를 들어 "보유자 있음" 검사를 지우면 빈 기록의 만료 시각이 영값이라 "만료"로 허용될 수 있다. 지금은 n-held의 사유 단언이 이 뮤턴트를 잡는다. 다른 행도 같은 방식으로 각자의 허용 사유를 단언하면 뮤턴트 표의 주장("모든 항이 자기만의 실패 행을 가진다")이 구조로 보장된다. — Severity: minor — Class: optional — Required fix: 허용 행마다 기대 감사 사유(`allow-self`, `allow-stale`, `allow-expired`, `allow-unheld` 등)를 표에 적는다.

## Recommendation

manager-spec에 줄 수정 지시다(blocking 먼저).

1. **D1** — spec.md:47의 "훅 쪽도 같은 순서로" 문장을 `path_resolve.go`:78-94의 실제 순서로 고친다. 가드 루트 정규화 요구사항을 추가하고, 워크트리 루트 → primary 기록 거부 AC와 그 뮤턴트 행을 추가한다. plan.md:38의 결론 문장도 함께 고친다.
2. **D2** — acceptance.md:91-93과 plan.md:99-100의 끼어들기 구성을 stall-release 타임아웃 방식으로 다시 쓴다. busy가 나오면 하네스 결함이고, 자식의 소유자 pid는 살아 있는 pid로 고정한다고 적는다.
3. **D3** — REQ-RSL-004를 busy와 held의 구별까지 넓히고, 양방향 판별 AC를 추가한다(REQ·AC 상한 16 유지).
4. **D4** — acceptance.md:247, :274, :323의 `BASELINE_SHA` 리터럴 핀을 lane 규칙 §8의 "읽는 시점 merge-base + 대조군 + 병합 전 전용" 형태로 바꾼다.
5. **D5** — REQ-RSL-014에서 "unavailable configuration"을 빼거나 켜진 뒤의 불확실성으로 좁히고, AC-RSL-012를 맞춘다.
6. **D6** — AC-RSL-014에 언어 도구 토큰 grep과 양성 대조를 추가한다.
7. optional(D7-D12)은 오케스트레이터 재량이다. 이 중 D7은 이미 측정했으므로 기록만 하면 OQ-1이 닫힌다.

재감사는 위 결함 목록의 변화분(delta)에 한정한다. Tier M 상한 때문에 다음 회차가 마지막이다.

---

## 5-Section Evidence

### Claim

- C1. 필수 통과 기준 MP-1..MP-7은 모두 통과한다.
- C2. 원장 EL-1..EL-6은 HEAD `1a178c174`에서 다시 재도 기록과 같다.
- C3. §B.3의 "훅도 같은 순서로 루트를 푼다"는 주장은 코드와 다르다(D1).
- C4. AC-RSL-001의 구성은 올바른 직렬화 아래에서 "보유 중" 결과에 닿지 않는다(D2).
- C5. 이 런타임에서 서브에이전트 Bash 호출의 훅 입력 `session_id`는 부모 세션 id다(D7).
- C6. M6 게이트는 여전히 닫혀 있다(종료 코드 1).

### Evidence

- E1 (원장 재측정, HEAD `1a178c174`):
  - `/usr/bin/grep -rlE 'SlotLease|slot_lease|slot-lease' internal cmd pkg` → stdout 없음, `EL-1 exit=1`
  - `/usr/bin/grep -rlE 'IntegrationLock' internal/kanban/integration_lock.go` → `internal/kanban/integration_lock.go`, `EL-2 exit=0`
  - `/usr/bin/grep -rl 'review_gate' internal/template/templates/.moai/config/sections` → `internal/template/templates/.moai/config/sections/workflow.yaml`, `EL-3 exit=0`
  - `/usr/bin/grep -rn 'Use: *"slot' internal/cli` → stdout 없음, `EL-4 exit=1`. 감사자가 추가한 양성 대조: `/usr/bin/grep -rn 'Use: *"integration' internal/cli` → `internal/cli/integration.go:162:		Use:   "integration [command]",`, exit 0
  - `go test ./internal/template/ -run 'TestTemplateNeutralityAudit$|TestTemplateNoInternalContentLeak$' -count=1 -v`(출력은 스크래치 파일로 보냄) → exit 0. 필터 결과는 `--- PASS: TestTemplateNoInternalContentLeak (0.55s)` / `--- PASS: TestTemplateNeutralityAudit (0.00s)` / `ok  	github.com/modu-ai/moai-adk/internal/template	1.059s`, 스윕 수 2
- E2 (코드 대조):
  - `internal/hook/path_resolve.go`:78-94 `resolveProjectRootFromEnvAt`는 `os.Getenv(config.EnvClaudeProjectDir)` → `os.Getwd()`만 쓴다. `internal/cli/integration.go`:46-65 `integrationLockRoot`는 `CLAUDE_PROJECT_DIR` → `git rev-parse --git-common-dir` → `resolveProjectDir()` 순서다.
  - `internal/kanban/integration_lock_mutation.go`:91-130: 예산이 다하면 `ErrIntegrationLockBusy`를 반환한다. `integration_lock_cross_test.go`:46-63: `integrationStallReleaseTimeout = 500 * time.Millisecond`와 교착 설명 주석("waiting only for B would deadlock").
  - 인용 줄 번호 대조: `Stale()` :152-160, `AcquireIntegrationLock` :216-262, `withIntegrationLockMutation` :75-88, `writeIntegrationLock` :315-360, 가드 접두어 :44, 패턴 :51, 따옴표 제거 :70, fail-open :74-98, `pre_tool.go` 배선 :541-548과 `integrationLockEnabled` :821-830, `types.go`:461-466·:669-678, `defaults.go`:900-902, `session_pid.go`:82-90. 모두 SPEC 인용과 일치한다. 예외는 §B.3 동치 주장(D1)과 plan.md:40의 "비공개"(D8)다.
- E3 (OQ-1 측정): 감사자 Bash 호출 `date -u +%FT%T; echo t607-oq1-probe` → `2026-09-11T15:31:44`. 직후 워크트리 trace 로그의 마지막 줄들: `{"ts":"2026-09-12T00:31:44.537515167+09:00","event":"PreToolUse","handler":"*hook.preToolHandler","tool":"Bash",…,"session_id":"432c40c9-1b0f-40c5-9e9e-3b22db04b9b2"}`. 감사자 Bash 환경: `SID=432c40c9-1b0f-40c5-9e9e-3b22db04b9b2`. trace의 session_id 출처는 입력의 `session_id`, 없으면 transcript_path UUID다(`internal/hook/trace_session.go`:1-8).
- E4 (루트 해석 관측): 감사자 Bash 환경 `CPD=`(비어 있음), `MSP=22103`. 이 세션의 Agent 스폰 감사 기록은 primary `/Users/goos/MoAI/moai-adk-go/.moai/logs/agent-model-audit.jsonl`에 `{"timestamp":"2026-09-11T15:29:30Z","session_id":"432c40c9-…","agent":"plan-auditor",…}`로 있다. 이 파일을 쓰는 곳은 `appendAgentModelAudit(h.projectRoot(), …)`(`agent_model_guard.go`:246)다.
- E5 (게이트): `git merge-base --is-ancestor WT-acquire-branch-record develop` → `EL-5 exit=1`. `git rev-parse --short develop` → `e82ef5565`, `git rev-parse --short WT-acquire-branch-record` → `f680dab46`. `git merge-base --is-ancestor WT-acquire-branch-record origin/develop` → `gate-origin rc=1`.
- E6 (lint, 트리에서 빌드): `go build -o <scratchpad>/t607-moai ./cmd/moai` → `build exit=0`. `<scratchpad>/t607-moai spec lint --strict .moai/specs/SPEC-RESOURCE-SLOT-LEASE-001` → `✓ No findings — all SPEC documents are valid`.
- E7 (D7/D8/MP-7): SPEC 참조 추출 → `SPEC-INTEGRATION-LOCK-ATOMIC-001 status: completed`, `SPEC-INTEGRATION-LOCK-LIVENESS-001 status: completed`, `SPEC-SYNC-SHA-SLOT-FORMAT-001 status: completed`. `syscall` 등장 수: spec.md:0, plan.md:0, acceptance.md:0. `grep -rn 'NEEDS CLARIFICATION' .moai/specs/SPEC-RESOURCE-SLOT-LEASE-001/` → `mp7 exit=1`.
- E8 (중립성 검사 범위): `template_neutrality_audit_test.go`의 클래스 이름은 `C1-macos-bias-path`, `C2-bare-narrative-v3r`, `C4-feedback-memory-ref`, `C5-claude-local-ref`, `C6-pr-number-ref`, `C9-natural-language-canonical-form`이다. CI 워크플로 :58-89는 이 테스트와 유출 테스트(일반·strict)만 돈다.

### Baseline-attribution

- 모든 측정은 이번 실행에서 워크트리 t607, HEAD `1a178c174`에 대해 했다. SPEC 문서 핀 `c4ce42eca`와의 차이는 SPEC 디렉터리뿐이라 코드 측정은 같은 트리를 잰 것이다(`git diff --stat c4ce42eca HEAD` → 4 files, 모두 SPEC 산출물).
- lint 측정은 HEAD `1a178c174`에서 `go build`로 만든 스크래치 바이너리를 경로로 호출했다. 설치본(`ed71054d3`)은 쓰지 않았다(§2.2). 이 빌드에는 ldflags가 없어 바이너리가 스스로 커밋을 보고하지 않는다. 커밋 좌표는 "HEAD `1a178c174`에서 빌드"라는 빌드 경로로 귀속한다.
- 게이트 SHA(`e82ef5565`, `f680dab46`)는 감사 시점의 움직이는 ref 값이다(SUBJECT형, R4). 판정 명령은 run 종료 시점에 다시 실행해야 한다.

### Gaps

- `moai cc -w <name>`처럼 워크트리 디렉터리에서 직접 시작한 세션에서 훅 환경의 `CLAUDE_PROJECT_DIR`가 워크트리를 가리키는지는 재지 못했다. D1의 결과(가드가 조용히 꺼짐)는 그 조건 아래의 추론이다. 코드 불일치 자체는 관측한 사실이다.
- E3의 OQ-1 관측은 이 세션 한 번, 이 런타임 버전 한 번이다. 다른 세션 형태(칸반 레인, Agent Teams 동료)에서는 재지 않았다.
- spec lint의 음성 대조(`phase: plan` 주입)는 재현하지 못했다. 워크트리 격리 가드가 스크래치 경로의 파일 편집을 거부했기 때문이다. 양성 결과(발견 0건)의 판별력은 progress.md:19의 작성자 대조군에 기대고 있다.
- 새 템플릿 규칙 파일이 규칙 인벤토리·미러 테스트(`rule_template_mirror_test.go`, `rule_provenance_audit_test.go`)에 걸리는 방식은 미리 돌려 보지 않았다(plan §C 사전 점검 소관).
- `-run` 필터 기반 AC들의 뮤턴트가 실제로 판별하는지는 run 단계 산출물이 없어 볼 수 없다. 설계 수준에서만 판정했다.

### Residual-risk

- D1을 고치더라도, 공유 루트를 CLI와 훅이 서로 다른 코드 경로로 푸는 한 같은 종류의 어긋남이 다시 생길 수 있다. 공통 해석 함수 하나를 두 곳에서 부르는 방식이 위험을 가장 작게 만든다.
- 통합 창 가드(`checkIntegrationLock`)도 같은 `h.projectRoot()`를 쓴다. D1이 사실이면 기존 통합 가드도 같은 노출을 가진다. 이 SPEC의 범위 밖이므로 후속 카드 후보로만 적는다.
- 선언 상한 만료 시 강제 없이 인수하는 결정(REQ-RSL-006)은 운영자 결정표에 없는 SPEC 자체의 결정이다. OQ-4로 드러나 있어 결함으로 치지 않았다. 다만 킥오프 승인 때 운영자에게 명시적으로 보여야 한다.
- Windows 변경 락 잔재 정리의 행동 증거는 CI 전까지 없다(acceptance.md:315가 스스로 밝힘).

## Regression Check

해당 없음(1회차).

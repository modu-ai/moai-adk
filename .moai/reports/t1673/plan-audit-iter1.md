auditor-model: claude-sonnet-5-5

# SPEC Review Report: SPEC-PRECOMMIT-AFFECTED-SCOPE-001

verdict: FAIL
audited_sha: 2febc5c2f5036059724819044b67131f64e9a596
overall_score: 0.75
must_pass_failed: 1
blocking_count: 8
plan_artifact_hash: ad66b8933b1fb4d3c02f3ee83e05a11ef7acfccb9ef367b0105e5c36e6b57b1f
scope: full
fix_scope: .moai/specs/SPEC-PRECOMMIT-AFFECTED-SCOPE-001/acceptance.md#AC-PAS-001 .moai/specs/SPEC-PRECOMMIT-AFFECTED-SCOPE-001/acceptance.md#AC-PAS-003 .moai/specs/SPEC-PRECOMMIT-AFFECTED-SCOPE-001/acceptance.md#AC-PAS-005 .moai/specs/SPEC-PRECOMMIT-AFFECTED-SCOPE-001/acceptance.md#AC-PAS-008 .moai/specs/SPEC-PRECOMMIT-AFFECTED-SCOPE-001/acceptance.md#AC-PAS-009 .moai/specs/SPEC-PRECOMMIT-AFFECTED-SCOPE-001/acceptance.md#AC-PAS-014 .moai/specs/SPEC-PRECOMMIT-AFFECTED-SCOPE-001/acceptance.md#AC-PAS-015 .moai/specs/SPEC-PRECOMMIT-AFFECTED-SCOPE-001/acceptance.md#LEDGER-PAS-VACUOUS .moai/specs/SPEC-PRECOMMIT-AFFECTED-SCOPE-001/plan.md#B.2 .moai/specs/SPEC-PRECOMMIT-AFFECTED-SCOPE-001/plan.md#B.3 .moai/specs/SPEC-PRECOMMIT-AFFECTED-SCOPE-001/spec.md#REQ-PAS-003 .moai/specs/SPEC-PRECOMMIT-AFFECTED-SCOPE-001/spec.md#REQ-PAS-008
defect_class: D1=format D2=design D3=design D4=design D5=design D6=ac-wording D7=ac-wording D9=ac-wording
convergence_overall: fail
required_backend: codex fail
Iteration: 1/3
Auditor Version: plan-auditor/v1 (sonnet-5-5)

Reasoning context ignored per M1 Context Isolation.

## 판정 요약

Tier M 통과 기준 0.80에 대해 종합 점수는 0.75이고, 필수 통과 항목 MP-8이 실패했으며, 필수 백엔드(codex)도 FAIL을 냈습니다. 그래서 FAIL입니다. 리더 판정(Q1=B 지연 메커니즘 없음, Q5 훅 유지 off, Q3 기본값, 테스트 skip·타임아웃 확대 금지)은 SPEC이 충실히 따르고 있습니다. 여기서 다시 따지지 않았습니다. 결함은 모두 판정 바깥, 즉 실제 git 어댑터 동작과 AC 판별력에 있습니다.

판정에 사용한 judging build: 감사 중 이 트리(HEAD 2febc5c2f, 깨끗함)에서 `go build ./cmd/moai`로 만든 바이너리(`/private/tmp/t1673-audit/bin/moai`, ldflags 없음이라 `moai version`은 `v3.1.3 none built unknown`을 출력. 소스는 위 HEAD와 동일). codex 감사 백엔드는 MCP 서버 빌드 `ee981a932`가 judge했고, 이 커밋은 HEAD의 조상입니다(서버 응답 `build_lag`로 확인). 따라서 codex 서버 쪽 코드는 HEAD보다 오래되었으나, 사용한 것은 diff 검토 기능이라 판정 근거는 트리 내용입니다.

## Must-Pass Results

- [PASS] MP-1 REQ 번호 일관성 (요구 계층 판정). `### REQ-PAS-001`..`015`가 연속이며 중복이 없습니다. REQ-PAS-012는 "WITHDRAWN (slot retained for traceability)"로 자리를 유지합니다(spec.md:L100). 추적 검증 스크립트 출력 `COLLECTED: 15 REQ definitions (acceptance input: read)`.
- [PASS] MP-2 GEARS 준수 (요구 계층 `REQ-XXX`만 판정, AC는 대상 아님). 001/004/005/007/008은 When 형, 011은 Where 형, 002/003/006/009/010은 Ubiquitous 형, 013/014/015는 레거시 EARS 부정형입니다. 철회된 012는 요구가 아니라 자리 표시입니다. 다만 REQ-PAS-011은 "Where ... is present"와 "default ... when the key is absent"를 한 문장에 섞어 형식이 어색합니다(D12, MINOR).
- [PASS] MP-3 frontmatter. spec.md:L1-16에 12개 필드 모두 있음. `version: "0.1.0"`(따옴표), `status: draft`, `created/updated: 2026-10-18`, `priority: P1`, `phase: "v3.2.0 target"`(금지된 단계명 아님), `lifecycle: spec-anchored`, `tags`는 쉼표 문자열. 별칭 필드 없음. `moai spec lint SPEC-PRECOMMIT-AFFECTED-SCOPE-001 --json` 출력은 info 한 건(`OwnershipTransitionUnmeasured`)뿐이며 이 린터의 침묵을 교차 근거로 쓰지 않습니다(린터 수집 REQ 수를 확인하지 않았음).
- [N/A] MP-4 언어 중립. 이 SPEC은 Go 툴체인의 pre-commit 테스트 스텝 범위를 다루는 단일 언어 한정 내용이고 비-Go 툴체인은 명시적으로 범위 밖입니다(spec.md Out of Scope).
- [PASS] MP-5 D7. 참조된 SPEC 상태: GATE-ASTGREP-REPAIR-001=completed, PRECOMMIT-GATE-SCOPE-001=completed, PRETOOL-GATE-MOVE-001=completed, UTIL-002=implemented. retired/superseded/archived 없음, REVIEW 줄 0건.
- [PASS] MP-6 D8. spec.md에 `syscall` 0건(`grep -c syscall` 출력 0), D8 스크립트에 BLOCKING 줄 없음.
- [PASS] MP-7. `grep -rn "NEEDS CLARIFICATION"` 출력이 비어 있음(SPEC 디렉터리 전체).
- [FAIL] MP-8 RED-now 셀 재실행. 릴리스 차단 AC 12개(AC-PAS-001~009, -011, -013, -016) 중 RED-now 셀(명령·stdout·종료코드·트리 SHA 4요소)을 가진 것은 AC-PAS-013(LEDGER-PAS-013-A)뿐입니다. 나머지 11개는 셀이 없고, LEDGER-PAS-VACUOUS(acceptance.md:L169-173)가 "run phase가 구현 전에 캡처한다"로 미룹니다. 이 SPEC의 정직한 진단(존재하지 않는 테스트는 `[no tests to run]` exit 0이라 RED가 아님)은 옳지만, verification-completeness.md §2·§2.1은 RED-now 셀 없는 기준을 "미채택"으로 규정하므로 MP-8은 성립하지 않습니다. AC-PAS-013의 셀은 아래 재현으로 RED가 재현됨을 확인했습니다(Evidence E1). 구체 수정은 D1.
- [PASS] MP-9. CN-4 스크립트 출력 `COLLECTED: 6 milestones in plan order (M1 M2 M3 M4 M5 M6), 0 exit bindings, 5 ordering candidates`, CONFLICT 줄 없음. 후보 5건을 직접 읽었고(acceptance.md:L62/L102/L164/L170/L185) 어느 것도 마일스톤 순서와 충돌하는 순서 의무가 아닙니다. 계획이 `Exit:` 줄로 AC를 마일스톤에 묶지 않아 스크립트가 기계적으로 볼 수 있는 충돌 자체가 0건이라는 점은 한계로 Gaps에 적습니다. DoD 1의 "RED output captured before GREEN"과 plan M1~M6 순서는 양립합니다.

## Category Scores

| 차원 | 점수 | 밴드 | 근거 |
|------|------|------|------|
| Clarity | 0.75 | 0.75 | REQ 문장은 대체로 단일 해석이나, plan.md:L24의 ".go는 디렉터리의 패키지 귀속"이 REQ-PAS-003/004의 testdata 규칙과 충돌(D3), AC-PAS-003의 기대값이 REQ-PAS-004와 모순(D7), REQ-PAS-011 형식 혼합(D12) |
| Completeness | 0.75 | 0.75 | HISTORY·배경·요구·AC·`### Out of Scope — <topic>` 3개 모두 존재(spec.md:L118, L124, L130, L136). 불완전 그래프(`go list -e` 성공 종료 + Incomplete)의 처리가 정의되지 않음(D4) |
| Testability | 0.50 | 0.50 | AC-PAS-014는 timeout 배수 변이를 통과시킴(D6), AC-PAS-015는 가변 외부 상태·머신 절대경로에 의존(D9), AC-PAS-009는 rename·linked worktree 형태 미포함(D2, D8), RED-now 셀 부재(D1) |
| Traceability | 1.00 | 1.0 | 추적 스크립트: UNCOVERED 0, ORPHAN 0. REQ-PAS-012와 AC-PAS-012는 서로 철회 자리로 짝지어짐 |

종합 = (0.75 + 0.75 + 0.50 + 1.00) / 4 = 0.75. Tier M 임계값 0.80(spec-workflow.md § SPEC Complexity Tier 표)에 미달입니다. 요구 개수 14 live+1 철회(상한 16), AC 15 live+1 철회(상한 16)로 Tier M 예산 안입니다.

## Defects Found

D1. MP-8 RED-now 셀 부재 — acceptance.md:L169-173 (LEDGER-PAS-VACUOUS), 표 acceptance.md:L9-26 — 릴리스 차단 AC-PAS-001~009, -011, -016에 4요소 RED-now 셀이 없다. 실제 RED가 지금 가능한 AC도 있다: AC-009는 아래 E2의 측정으로 "스크럽된 질의가 `commit -a`에서 `d/z.go`를 놓치고 경로 지정 커밋에서는 엉뚱한 `a/x.go`를 돌려준다"를 현재 트리에서 빨강으로 보여줄 수 있고, AC-016은 현재 바이너리가 세 패키지 모두에 `go test ./...`를 돌리는 것으로 "미접촉 패키지 테스트 미실행"이 지금 빨강임을 보일 수 있다. — Severity: critical — Class: blocking — Required fix: AC-009·AC-016에 현재 트리에서 재실행 가능한 RED-now 셀(명령, 원문 stdout, 종료코드, 트리 SHA 2febc5c2f)을 붙인다. 순수 단위 AC(001~008, 011)는 존재하지 않는 심볼에 대한 `go test -run`이 RED가 아니므로, 각 AC에 "RED 셀 = 구현 전 테스트 커밋의 실패 출력"을 쓰는 대신 RED가 지금 실제로 빨강인 명령(예: 아직 없는 심볼의 `go doc` 종료코드 1)을 인용하거나, 이 AC들을 RED 관측 전까지 채택 대기로 분류하고 그 사실을 DoD에 단일 문장으로 못박는다. 어느 쪽이든 "RED 없이 릴리스 차단" 상태를 남기지 않는다.

D2. 실제 git 어댑터의 rename·delete 처리 미정의 — plan.md:L34(B.3), acceptance.md:L55-59(AC-PAS-005), L79-84(AC-PAS-009) — REQ-PAS-005/009는 이전 경로 기반 선택을 요구하지만, 기존 도우미가 쓰는 `git diff --cached --name-only`는 rename에서 목적지 경로만 돌려준다. 측정(E2의 (iv)): `git mv old new` 후 훅에서 `new/o.go`만 나오고 `--no-renames`를 주면 `old/o.go`가 추가된다. AC-PAS-005는 두 경로를 코어에 직접 주입하므로 어댑터가 이전 경로를 놓쳐도 통과한다(변이 통과, codex P1과 일치). — Severity: major — Class: blocking — Required fix: B.3에 어댑터가 rename 감지를 끈 목록(`--no-renames` 계열)을 쓴다고 명시하고, AC-PAS-009(또는 신규 AC)에 실제 git을 거친 패키지 간 rename과 마지막 파일 삭제 사례를 추가해 이전 패키지의 importers가 선택됨을 검증한다.

D3. testdata 아래 `.go` 파일의 소유 규칙 모순 — plan.md:L24(B.2 step 2) vs spec.md:L66(REQ-PAS-003), L70(REQ-PAS-004) — step 2는 ".go 파일은 그 디렉터리의 패키지에 귀속"이라 쓰는데, `go list`는 `testdata`를 패키지로 열거하지 않는다(E3). 이 저장소는 `internal/astgrep/testdata/fixtures/*/…` 아래 `.go` 픽스처를 실제로 추적한다(`git ls-files` 출력의 `go/suppressed.go`, `go/valid.go`, `go/violation.go`). 그대로 구현하면 이런 파일의 변경이 소유자 없음(조용한 테스트 누락)이 된다. REQ-PAS-003은 "non-.go 파일"만 규정하고, AC-PAS-003/004는 JSON만 쓴다. — Severity: major — Class: blocking — Required fix: 소유 규칙을 확장자 분기보다 앞서 `testdata` 세그먼트로 먼저 판정하도록 B.2와 REQ-PAS-003을 맞추고, `.go` 픽스처 변경 시 소유 패키지만(의존자 제외) 선택한다는 AC를 AC-PAS-004에 추가한다.

D4. 불완전 그래프 fallback 미정의 — plan.md:L20(B.2 입력), spec.md:L86(REQ-PAS-008), acceptance.md:L73-77(AC-PAS-008) — `go list -e`는 오류가 있어도 종료코드 0을 돌려준다. 측정(E3): 존재하지 않는 의존성을 import하는 패키지가 `incomplete=true deps_errors=1`이며 명령 `exit=0`. REQ-PAS-008은 "질의 실패"만 다루므로, 삭제·이동된 패키지를 아직 import하는 패키지(바로 REQ-PAS-005의 상황)가 들어간 불완전 그래프를 정상으로 받아도 AC-PAS-008이 통과한다. — Severity: major — Class: blocking — Required fix: REQ-PAS-008에 "그래프에 Error/DepsErrors/Incomplete가 있으면 질의 실패와 동일하게 전체 범위로 fallback하고 통지한다"를 넣고, AC-PAS-008에 종료코드 0인 오류 응답을 실제 `go list` 어댑터로 검증하는 사례를 추가한다.

D5. 모듈 밖 파일 제거가 widening 검사보다 앞섬 — plan.md:L23(B.2 step 1) vs spec.md:L82(REQ-PAS-007) — step 1이 모듈 루트 밖 변경 파일을 먼저 버린 뒤 trigger를 검사한다. 하위 모듈 레이아웃(acceptance.md:L179의 "nested module" 가장자리 사례가 명시적으로 지원하는 형태)에서 저장소 루트의 `go.work`와 `.moai/config/build-tags`가 widening 검사 전에 사라진다. AC-PAS-007에는 모듈 오프셋 사례가 없다. 이 저장소는 루트에 `go.mod`가 있어 오프셋이 비어 있으므로 당장은 발현하지 않지만, SPEC이 스스로 지원하겠다고 한 형태입니다. — Severity: major — Class: blocking — Required fix: 공통 widening trigger를 모듈 경계 필터보다 먼저 검사하도록 B.2 순서를 바꾸고, AC-PAS-007에 하위 모듈 오프셋 사례를 추가한다.

D6. AC-PAS-014가 실행 예산 확대 변이를 통과시킴 — acceptance.md:L108-112, gate.go:619 — 판정 조건이 두 `gate.yaml`의 `timeouts:` 블록, 기본값, `t.Skip` 추가, 삭제된 `func Test` 수뿐이다. `gate.go:619`의 `g.executeStep(ctx, testStep, dir, g.config.TestTimeout)`를 `2*g.config.TestTimeout`으로 바꾸면 실제 예산이 240초가 되지만 위 조건은 모두 그대로다(codex P2와 일치, 인용 줄은 직접 확인). verification-completeness.md §2 mutant probe 기준으로 채택 불가. — Severity: major — Class: blocking — Required fix: 실행기에 전달되는 timeout 값을 주입 가능한 경계에서 단언하는 테스트(예: 가짜 `go`를 돌리며 실행기가 받은 deadline 인자를 기록)를 AC에 포함해 배수 변이가 빨강이 되게 한다. 시간 측정이 아니라 인자 값 비교로 한다.

D7. AC-PAS-003이 자체 모순이며 그래프가 없음 — acceptance.md:L42-46, spec.md:L70(REQ-PAS-004) — Then이 "선택은 `{p, q}`"이면서 괄호로 "p/q dependents per AC-PAS-001 rules"를 덧붙인다. p의 변경은 `p/testdata/in.json`뿐이라 REQ-PAS-004에 따라 p의 dependents는 들어가면 안 되고, Given에는 dependents를 가진 그래프도 정의돼 있지 않다. 판정자가 어느 쪽을 기대값으로 삼을지 갈린다. — Severity: major — Class: blocking — Required fix: AC-PAS-001 그래프를 재사용하는 명시적 그래프를 Given에 두고 기대 집합을 정확히 적는다(p는 단독, q는 q+그 dependents, `docs/note.md`는 소유자 없음).

D8. AC-PAS-009가 linked worktree 형태와 amend를 다루지 않음 — acceptance.md:L79-84, internal/gitenv/gitenv.go:L22-30 — 이 저장소의 모든 커밋은 linked worktree에서 나오고, gitenv.go 주석은 그 형태가 `GIT_DIR`=`.git/worktrees/<name>`이라는 별도 형태임을 기록한다. AC는 평범한 임시 저장소만 쓴다. 이 결함이 현재 변이를 통과시키는 구체 사례는 만들지 못해 확정하지 않았다. — Severity: SHOULD-FIX — Class: optional — Required fix: 같은 세 커밋 형태를 linked worktree에서 반복하는 사례를 AC-PAS-009에 추가하고, `--amend`는 plan.md에 수용 동작(추가 스테이징분만 범위)을 한 줄 명시한다.

D9. AC-PAS-015가 가변 외부 상태와 머신 절대경로에 의존 — acceptance.md:L114-118 — "착지 후" `git config --show-origin --get core.hooksPath`가 `file:/Users/goos/MoAI/moai-adk-go/.git/config<TAB>/dev/null`을 출력해야 한다고 쓴다. Q5 handoff(plan.md §E)가 승인되어 리더가 스위치를 내리면 이 회귀 가드가 코드 결함 없이 빨강으로 바뀌고, 다른 머신·경로에서도 재현되지 않는다. 이 AC가 보호하려는 요구(REQ-PAS-015: 전달 코드가 git 설정을 쓰지 않는다)는 두 번째 절인 added-lines 스캔이 이미 측정한다. — Severity: SHOULD-FIX — Class: blocking — Required fix: 첫 절(현재 설정값 단언)을 AC에서 빼고 "이 SPEC의 diff가 `hooksPath`나 `git config` 쓰기를 추가하지 않는다"만 판정 기준으로 남긴다. 스위치 시점 확인은 DoD가 아닌 runbook 사전점검으로 옮긴다.

D10. 측정되지 않은 사실이 사실로 서술됨 — spec.md:L133, plan.md:L113(R2) — "internal/cli 스위트가 120초 예산을 넘는다"는 서술은 측정이 아니다(ruling 35e9가 측정을 금지하므로 1022개 테스트 파일에서의 추론이다). 또한 `internal/hook`(`go list` 기준 테스트 파일 293+3개)도 같은 위험이고, `internal/config`의 47개 dependents를 한 번의 `go test` 호출·하나의 120초 예산에 합산하는 점이 R2에 없다. 이 SPEC이 전달해도 흔한 커밋의 영향 범위가 예산 안에 들어오는지는 어떤 AC도 보여주지 않는다. — Severity: SHOULD-FIX — Class: optional — Required fix: 해당 서술을 "추론(ruling 35e9, 테스트 파일 수)"으로 표기하고, R2와 runbook 사전점검 3에 "대표 커밋 몇 건의 affected set 테스트 소요를 슬롯 임대 아래 측정"을 한 줄 추가한다. 코드 변경이나 새 AC는 필요 없다.

D11. F6의 dependents 수 불일치 — spec.md:L48 F6 — "internal/config has 48 direct dependents (47 with tests)"인데 같은 트리 2febc5c2f에서 정확한 import 경로 일치로 세면 47개(테스트 없는 것 1개, `internal/closure/closuretest`)이고, 48은 `internal/config/…` 하위 패키지 import까지 접두사 일치로 센 값이다(E4). — Severity: MINOR — Class: optional — Required fix: 47(테스트 있는 것 46)로 고치거나 접두사 일치 기준임을 명시한다.

D12. REQ-PAS-011의 GEARS 형식 혼합 — spec.md:L98 — "Where ... is present" 절에 "default to affected when the key is absent"를 함께 담아 한 문장이 서로 반대 조건을 가진다. — Severity: MINOR — Class: optional — Required fix: 기본값 규칙(Ubiquitous)과 값 검증 규칙(Where)을 두 문장으로 나눈다.

D13. embed 소유 판정을 자체 구현할 필요가 있는지 — plan.md:L20, L24 — B.2는 `EmbedPatterns`를 받아 "pattern roots resolved relative to the package directory"를 직접 해석한다고 쓴다. `go list`가 이미 해석된 `EmbedFiles`를 제공하므로, 패턴 해석 재구현(제외 규칙, `all:` 등)은 불필요한 복잡도다. 이 판단은 `go list` 필드 존재에 대한 지식에 근거하며 이 트리에서 확인하지 않았다. — Severity: MINOR — Class: optional — Required fix: `EmbedFiles`를 쓰면 되는지 M3 시작 시 확인하고 plan에 반영한다.

D14. config 구조체↔YAML 대칭 CI 가드 미언급 — plan.md:L38(B.4) — settings-management.md의 "Adding a new YAML section"은 `audit_struct_yaml_symmetry_test.go`의 symmetryCases 갱신을 요구한다고 적는다(`CONFIG_STRUCT_YAML_MISMATCH`). B.4의 touch point 목록에 없다. 이 트리에서 해당 테스트 파일은 읽지 않았다. — Severity: MINOR — Class: optional — Required fix: B.4 touch point에 대칭 테스트 갱신 여부를 확인하는 한 줄을 넣는다.

(optional 항목은 단독으로 FAIL 사유가 아닙니다. 실패 근거는 MP-8, 점수 0.75 < 0.80, 필수 백엔드 codex FAIL, 그리고 blocking D1~D7·D9입니다.)

## Recommendation (manager-spec 수정 지시)

1. D1: AC-009·AC-016에 RED-now 셀을 붙이고(E2 측정치 활용), 나머지 릴리스 차단 AC의 RED 처리 방침을 DoD에 명문화.
2. D2·D3·D4·D5: plan.md B.2·B.3와 REQ-PAS-003/005/007/008을 맞추고 대응 AC를 추가.
3. D6: AC-PAS-014에 timeout 인자 단언 추가.
4. D7·D9: AC-PAS-003과 AC-PAS-015 재작성.
5. 재감사는 위 enumerated delta로 범위를 좁히되, CN-4 순서 검증은 전체 실행.

## Claim / Evidence / Baseline-attribution / Gaps / Residual-risk

### Claim

C1. 결함 F7(ast-grep 억제 정책 검사가 픽스처 `suppressed.go:15`를 위반으로 보고해 `moai gate`를 막음)은 실재한다.
C2. G1(스크럽된 staged 질의가 `git commit -a`/경로 지정 커밋에서 틀린 집합을 돌려줌)은 실재한다.
C3. 기본 `git diff --cached --name-only`는 rename에서 이전 경로를 반환하지 않는다.
C4. `go list -e`는 오류 있는 모듈에서 종료코드 0 + `Incomplete: true`를 반환하고, `testdata` 디렉터리를 패키지로 열거하지 않는다.
C5. 훅 스크립트와 템플릿 쌍은 동일하고, `gate.yaml` 두 사본은 `pre_commit.enabled`와 `rules_dir`만 다르다(F3·F4).
C6. 공유 설정 `core.hooksPath`는 `/dev/null`이다(F5).
C7. 추적성은 완전하다(미커버 0, 고아 0).

### Evidence

E1 (C1). 명령 `sh /private/tmp/t1673-audit/repro-013.sh` — go.mod와 픽스처 사본, 저장소의 `.moai/astgrep-rules` 사본, LEDGER-PAS-013-A와 같은 `gate.yaml`로 스크래치 프로젝트 `/private/tmp/t1673-audit/proj`를 만들고 `env CLAUDE_PROJECT_DIR=... /private/tmp/t1673-audit/bin/moai gate` 실행. 원문 출력:
```
exit=1
ast-grep suppression policy violations:

[SUPPRESSION_WITHOUT_REASON] ast-grep suppression at /private/tmp/t1673-audit/proj/internal/astgrep/testdata/fixtures/go/suppressed.go:15 requires adjacent '// @MX:REASON <rationale>' on next line

quality gate steps (3 configured):
  - go vet: disabled — turned off by the gate.disabled_steps setting
  - golangci-lint: disabled — turned off by the gate.disabled_steps setting
  - go test: not reached
          
   ERROR  
          
  Quality gate failed.
```
LEDGER-PAS-013-A의 기록(exit 1, 같은 위반 줄, 같은 단계 요약)과 일치하고, 위반은 픽스처 한 파일 때문이므로 사유도 맞다. 원인 코드: `astgrep_gate.go`의 `walkSourceFiles`는 `vendor`/`.git`/`node_modules`/`__pycache__` 세그먼트와 `_test.go`만 건너뛰고 `testdata`는 건너뛰지 않으며(sed 출력 L375-L401), `RunAstGrepGateV2`는 억제 검사를 sg 스캔보다 먼저 돌려 위반이 있으면 `block_on_error`와 무관하게 `return false`한다(L81-L100). 트리 내 위반 후보 파일 목록(`git grep -lE "^[[:space:]]*(//|#|--) ast-grep-ignore[[:space:]]*$"` 출력): `.moai/specs/SPEC-UTIL-002/{research,spec}.md`, 픽스처 6개(csharp·elixir·go·kotlin·php·ruby), `internal/hook/quality/astgrep_gate_suppression_test.go`. 확장자상 `.md`는 검사 대상이 아니고 `_test.go`는 제외이므로 실제로 걸리는 것은 픽스처뿐이다. 다른 다섯 픽스처는 마커→REASON 순서이고 Go만 REASON→마커 순서(`sed -n 1,30p suppressed.go` 출력으로 L13-L15 확인)다.

E2 (C2·C3). 명령 `sh /private/tmp/t1673-audit/run-probe.sh` — `core.hooksPath`를 스크래치 훅 디렉터리로 둔 임시 저장소(`/private/tmp/t1673-audit/repo`)에서 `a/x.go`·`d/z.go`를 수정하고 훅이 `GIT_INDEX_FILE`과 세 가지 질의를 기록. 원문 출력:
```
=== (i) plain commit, a/x.go staged, d/z.go modified unstaged
--- hook run
GIT_INDEX_FILE=.git/index
unscrubbed cached:
a/x.go
scrubbed cached:
a/x.go
scrubbed cached no-renames:
a/x.go
=== (ii) commit -a
--- hook run
GIT_INDEX_FILE=/private/tmp/t1673-audit/repo/.git/index.lock
unscrubbed cached:
a/x.go
d/z.go
scrubbed cached:
a/x.go
scrubbed cached no-renames:
a/x.go
=== (iii) commit d/z.go (path-limited)
--- hook run
GIT_INDEX_FILE=/private/tmp/t1673-audit/repo/.git/next-index-29197.lock
unscrubbed cached:
d/z.go
scrubbed cached:
a/x.go
scrubbed cached no-renames:
a/x.go
=== (iv) rename old->new staged
--- hook run
GIT_INDEX_FILE=.git/index
unscrubbed cached:
a/x.go
new/o.go
scrubbed cached:
a/x.go
new/o.go
scrubbed cached no-renames:
a/x.go
new/o.go
old/o.go
```
해석: (ii) 스크럽된 질의는 `d/z.go`를 놓침(커밋은 두 파일을 기록), (iii) 스크럽된 질의는 `a/x.go`(커밋이 기록하지 않는 파일)를 돌려주고 `d/z.go`를 놓침, (iv) 기본 이름 목록은 `old/o.go`를 빼고 `--no-renames`만 포함. SPEC의 G1 가설(plan.md:L34)을 확인했다. (iv)의 `a/x.go`는 (iii) 경로 지정 커밋이 남긴 스테이징 잔재다.

E3 (C4). 명령 `sh /private/tmp/t1673-audit/golist-probe.sh`(임시 모듈 `/private/tmp/t1673-audit/mod`, `a/testdata/fixture.go` 포함, `b`가 없는 의존성 import). 원문 출력:
```
example.com/m/a incomplete=false err=no deps_errors=0
example.com/m/b incomplete=true err=no deps_errors=1
exit=0
--- is testdata dir a package?
example.com/m/a
example.com/m/b
```
`incomplete=true`인데 종료코드 0이고, `a/testdata`는 패키지로 열거되지 않는다.

E4 (C5·C7 및 D11). 명령 `diff /Users/goos/MoAI/moai-adk-go/.git/hooks/pre-commit .../internal/template/templates/.git_hooks/pre-commit` — 출력 비어 있음. 명령 `diff .../.moai/config/sections/gate.yaml .../internal/template/templates/.moai/config/sections/gate.yaml` 출력: `62c62 < enabled: true --- > enabled: false`, `70c70 < rules_dir: ".moai/astgrep-rules" --- > rules_dir: ".moai/config/astgrep-rules"`. 명령 `git config --show-origin --get core.hooksPath`(이 워크트리에서) 출력: `file:/Users/goos/MoAI/moai-adk-go/.git/config	/dev/null`. 패키지 그래프: `go list -e -f '{{.ImportPath}}|{{join .Imports ","}},{{join .TestImports ","}},{{join .XTestImports ","}}|{{len .TestGoFiles}}+{{len .XTestGoFiles}}' ./...`를 파일로 저장 후 awk 집계: `packages=176 config_dependents=47 without_tests=1`(정확 일치), 접두사 일치는 `prefix_match_dependents=48 without_tests=1`. `internal/hook/quality`의 dependents와 테스트 파일 수: `internal/cli 1013+9`, `internal/hook 293+3`. 추적 스크립트(`sh /private/tmp/t1673-audit/trace.sh`) 출력: `COLLECTED: 15 REQ definitions (acceptance input: read)`와 UNCOVERED/ORPHAN 줄 없음. codex 호출과 `moai verify audit-plan`(아래 E5).

E5 (필수 백엔드). 명령 `moai verify audit-plan --project-root /Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1673` 출력의 요점: `"model": "codex"`, `backends`: claude gate off, codex gate required, glm gate off, `"cross_model_active": true`, `"enforced_required": ["codex"]`. `mcp__moai__audit_multi`(project_root, card_id=t1673, target=SPEC 디렉터리 상대경로) 결과 요점: `per_backend_verdicts`는 codex `gate: required`, `verdict: fail`(P1 1건, P2 4건), `overall_verdict: fail`, `residual_risk_note: "required-backend FAIL: codex"`, `audit_receipt: "rcpt-1a5b4efa8d8af09c2f14685c"`, `plan_source: config`, `build_commit: ee981a932`. codex의 5건: (P1) rename/delete 어댑터 미검증 = D2, (P2) testdata `.go` 소유권 = D3(AC-003/REQ-004 충돌 지적은 D7), (P2) `go list -e` Incomplete = D4, (P2) widening이 모듈 필터 뒤 = D5, (P2) AC-014 timeout 변이 = D6. 이 중 D2·D3·D4는 위 E2·E3로 독립 재현했고 D6은 `gate.go:619` 인용 줄을 읽어 확인했다. 결과 다이제스트를 `.moai/state/audit-plan-result.json`에 쓰고 `moai verify audit-plan --project-root ... --result-file ...`로 확인한 `convergence_check`는 `{"ok": true, "unmet": []}`로, 필수 백엔드가 응답하지 않은 미충족 게이트는 아니고 응답한 필수 백엔드의 리뷰 FAIL이다.

### Baseline-attribution

- 트리: 이 워크트리 HEAD `2febc5c2f5036059724819044b67131f64e9a596`, 작업 트리 깨끗함(세션 시작 `git status`와 `git rev-parse HEAD` 출력). 모든 관측은 이 감사 실행 중 이 트리에서 이뤄졌다.
- 도구 빌드(verification-claim-integrity §2.2): E1·E5 일부와 `moai spec lint`는 위 HEAD에서 `go build`한 `/private/tmp/t1673-audit/bin/moai`로 실행했다(경로 호출). `moai verify audit-plan`과 MCP 호출은 설치본/서버 빌드 `ee981a932`가 처리했고, 서버는 HEAD의 조상이라고 스스로 `build_lag`를 보고했다. 이 두 호출의 결과는 트리가 아니라 구버전 빌드의 해석임을 밝힌다.
- 레저 RED-now 재현(MP-8): LEDGER-PAS-013-A 명령은 `moai gate`를 단일 호출로 실행하며(`env CLAUDE_PROJECT_DIR=<dir> <moai> gate`) 금지 메타문자가 없는 형태다. 위 E1이 같은 레시피를 같은 트리 SHA에서 재실행한 결과다.

### Gaps (미검증)

- G-a. 격리된 워크트리 가드가 `env GIT_INDEX_FILE=.git/index git -C <스크래치> diff --cached --name-only` 형태 명령을 "런타임에 계산된 경로"라며 거부했다. 대신 같은 질의를 스크립트 파일(`relidx.sh`)로 실행했으나, 스크래치 저장소가 깨끗한 상태였고 출력이 비어 있어 "상대 경로 `GIT_INDEX_FILE`이 하위 디렉터리 cwd에서 어떻게 해석되는가"는 판별하지 못했다. 이 주장은 감사 결함으로 올리지 않았다.
- G-b. `plan_artifact_hash`는 `internal/runtime/audit_cache.go`의 `ComputeHash` 알고리즘(artifact 이름 순서, `이름:길이\0`+원문+`\0`)을 읽고 셸로 재구현해 계산했다. 런타임 함수를 직접 호출한 값이 아니다. 이 SPEC 디렉터리에는 acceptance·decision-index·plan·spec 네 파일만 있고 design/research/tasks는 없다.
- G-c. CN-4 스크립트는 계획에 `Exit:` 줄이 없어 `0 exit bindings`이므로 CONFLICT를 기계적으로 볼 수 없다. MP-9 판정은 후보 5건을 직접 읽은 결과다.
- G-d. 린터(`moai spec lint`)가 수집한 요구 개수는 확인하지 않았으므로 린터의 침묵은 교차 근거로 쓰지 않았다.
- G-e. 실제 저장소 전체 트리에서의 억제 검사 실행(F9·G4)은 수행하지 않았다. `git grep`으로 후보를 좁힌 것이 전부다.
- G-f. SPEC의 핵심 주장인 "affected set이 120초 예산 안에서 테스트된다"는 측정하지 않았다(D10). `internal/cli`·`internal/hook` 실제 소요도 ruling 35e9에 따라 측정하지 않았다.
- G-g. D13(`EmbedFiles` 존재)과 D14(대칭 CI 가드)는 읽은 규칙 문서에 근거한 것이며 이 트리의 해당 코드·테스트 파일은 읽지 않았다.
- G-h. `.moai/state/audit-plan-result.json`을 이 감사가 워크트리에 만들었다(`.moai/state/`는 런타임 상태 디렉터리). 커밋하지 않았다.

### Residual-risk (관측했음에도 남는 위험)

- 위 결함을 모두 고쳐도, 흔한 커밋(예: `internal/config` 또는 `internal/hook/quality` 변경)의 affected set이 `internal/cli`와 `internal/hook`을 포함해 예산 초과로 실패하는 상황은 Q1=B 판정의 결과로 그대로 남는다. SPEC은 이를 수용했고 후속 카드에 맡기므로 결함으로 올리지 않았다. 다만 이 SPEC만으로는 훅 재활성화가 실사용 가능해진다는 보장이 없다.
- R3(경로로 저장소 파일을 읽는 테스트는 패키지 분석에서 보이지 않음)은 SPEC이 수용한 잔여 위험이다. 이 저장소는 로컬 `.moai/config/sections/gate.yaml`·`.claude/rules/**`와 템플릿 사본을 비교하는 테스트가 있을 수 있어 로컬 사본만 바꾸는 커밋은 "소유자 없음"으로 테스트가 생략된다.

## Operational Notes (unverified)

- [measured] 감사 중 만든 스크래치 파일은 모두 `/private/tmp/t1673-audit/` 아래에 있고 추적 파일은 수정하지 않았다 (`git status --short`는 `.moai/reports/t1673/`와 `.moai/state/audit-plan-result.json`만 새로 보일 것). 확인 명령: `git status --short`.
- [inferred] 재감사 시 D1을 고치는 가장 싼 길은 AC-009·AC-016의 RED-now를 E2의 스크립트 형태(스크래치 저장소, 현재 바이너리)로 옮기는 것이다. 근거: 두 AC의 현재 실패가 이 감사 E2로 이미 관측되었다.
- [assumption] Tier M에서 D8·D10~D14를 한 번에 닫으면 REQ/AC 상한(16/16)에 다시 닿을 수 있다. 새 AC를 만들기보다 기존 AC의 Given/Then를 넓히는 쪽을 권한다.

AUDIT-VERDICT: FAIL spec=SPEC-PRECOMMIT-AFFECTED-SCOPE-001 receipts=rcpt-1a5b4efa8d8af09c2f14685c

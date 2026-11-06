auditor-model: claude-opus-5-5
verdict: FAIL
audited_sha: 7ea9b205552e41f3381d266fee144fbc95a06f6d
overall_score: 0.73
functionality_score: 72
security_score: 62
craft_score: 80
consistency_score: 84
must_pass_failed: 2
blocking_count: 2
blocking_defect: D16
blocking_defect: D13
required_backend_fail: codex
receipts: rcpt-919c8ec5ced34525ba7663b0

# Sync 감사 판정 — SPEC-AUDIT-ENGINE-CONVERGENCE-001 (카드 t2017), 4회차

## 요약

판정은 FAIL, 가중 점수 72.9(0.73)이다. 차단 결함은 2건이다.

- **D16 (신규, Codex P1)**: §E.2 안에 `` ```a``` `` 처럼 줄 첫머리가 백틱 세 개로 시작하는 인라인 코드 줄이 있고, 그 뒤에 `## ` 줄을 담은 정상 코드 블록이 오면 fence 파서가 짝을 거꾸로 맞춘다. 그러면 §E.2가 그 코드 블록 안에서 끝나 버리고, 그 아래 AC 행을 PASS에서 FAIL로 고쳐도 "변경 없음"으로 판정된다. 원문 헤딩 수와 파싱된 섹션 수가 둘 다 1이라 D10 수리의 가드도 통과한다. 이번 실행에서 스크래치 사본으로 재현했다.
- **D13 (3회차에는 optional, 이번에 blocking으로 재분류)**: §E.2 안의 HTML `<pre>` 블록에 `## ` 줄이 있을 때도 결과가 똑같다. 이번 실행에서 다시 재현했다.

두 결함 모두 REQ-AEC-010의 "`§E.2` 아래 한 줄만 바뀌어도 stale"을 유효한 Markdown 입력에서 깨뜨린다. 판정 대상 커밋에 그 형태가 이미 있어야 한다는 전제 조건도 같다. 1회차 D1이 바로 이 조건으로 BLOCKING이었으므로 같은 기준을 적용했다.

1회차 D1과 2회차 D10은 닫혔다. 두 수리를 되돌린 뮤턴트를 기존 테스트가 모두 잡았다.

운영자 결정 d-20261106T055051Z-db22에 따라 `mcp__moai__codex_audit`를 정확히 1회 호출했다. 결과는 `fail`, 영수증은 `rcpt-919c8ec5ced34525ba7663b0`이다. 이 트리는 codex 게이트가 `required`이고 codex가 fail을 보고했으므로, 내 판정과 상관없이 PASS 계열 판정은 나올 수 없다.

## 차원별 점수 (default 프로파일, flat 가중)

| 차원 | 점수 | 판정 | 근거 |
|---|---|---|---|
| Functionality (40%, must-pass) | 72/100 | FAIL | 지정 테스트 3묶음이 모두 통과했다(auditreceipt `ok 30.457s`, hook `ok 45.372s`, cli `ok 280.467s`). 그러나 REQ-AEC-010을 위반하는 입력이 두 가지 재현되었다(D16, D13). must-pass 차원에는 부분 점수를 주지 않는다. |
| Security (25%, must-pass) | 62/100 | FAIL | 영수증 신선도 게이트를 우회하는 경로가 있다(D16, D13). 1회차 D1과 결과도 전제 조건도 같은 부류이며, D1은 High였다. 그 밖에 D2(이어받은 레그의 gate)와 D5(reports symlink)는 optional로 둔다. |
| Craft (20%) | 80/100 | PASS | golangci-lint `0 issues.`, gofmt 출력 없음, auditreceipt 커버리지 87.1%. D1·D10 뮤턴트는 모두 잡혔다. 다만 파서 주석은 "Fences follow CommonMark"라고 하는데, 백틱 fence의 info string에 백틱이 들어갈 수 없다는 CommonMark 규칙이 빠져 있다(D16의 원인). |
| Consistency (15%) | 84/100 | PASS | `make agents-emit-check` → `specs=12 copies=4 diff=0`, `make embed-manifest-check` → `ok`. D15(§E.4 산문)는 3회차 기록을 따랐고 재측정하지 않았다. |

가중 합: 72×0.40 + 62×0.25 + 80×0.20 + 84×0.15 = 72.9. must-pass 두 차원이 모두 기준 미달이라 firewall에 따라 FAIL이다.

## 백엔드별 결과

- **codex** (gate `required`, `mcp__moai__codex_audit`, mode `adversarial`, target `baseBranch`, review_base `main (merge base 9b632c1eb3544cef5f275e502922a375e48d20aa)`, resolved_model `gpt-6.1-sol`, build_commit `646955327e`): verdict **`fail`**, receipt **`rcpt-919c8ec5ced34525ba7663b0`**. 지적은 5건이다(아래 원문 요약). Codex 본인이 밝힌 검증 한계도 그대로 옮긴다: "정적 경로와 Python 알고리즘 대조 결과다. Go 집중 시험은 읽기 전용 정책의 임시 디렉터리 생성 거부로 실행되지 않았다. CI는 NOT_RUN이다."
  1. [P1] `store.go:1935` Fence 오인식으로 §E.2 변경을 누락하고 영수증을 재사용할 수 있다 → **D16. 동의한다. 이번 실행에서 재현했다.**
  2. [P1] `audit_rerun_legs.go:153` 재사용한 Claude 결과의 이전 gate가 현재 required 정책을 무력화한다 → **D2. 코드 경로는 사실이지만 P1에는 동의하지 않는다** (아래 근거).
  3. [P2] `mcp_convergence.go:1185` 미충족 backend만 복구해도 반환된 영수증은 계속 거부된다 → D6, fail-closed 방향. optional.
  4. [P2] `store.go:1857` 감사 당시 수정돼 있던 허용 증거를 커밋하면 영수증이 다시 stale이 된다 → D17(신규). fail-closed 방향이고, REQ-AEC-009 문구("the uncommitted state now equals the judged one")와 맞다. optional.
  5. [P2] `verify_clear_rejection.go:195` reports 디렉터리 자체의 symlink로 증거 파일이 작업트리 경계를 벗어날 수 있다 → D5. §E B4가 보안 경계가 아니라고 정한 범위다. optional.
- **claude**: 이 트리의 계획에서 gate `off`. 이 감사(Claude, 자체 검토)는 위 증거를 독립적으로 수집했다.
- **glm**: gate `off`, 호출하지 않음.
- `audit_multi`는 운영자 결정에 따라 호출하지 않았다. 그래서 convergence digest(`overall_verdict`, `gate_unmet`, `plan_source`)가 없고 orchestrator의 `--result-file` 확인 대상도 없다.

### Codex P1 #2(D2)에 동의하지 않는 근거

`carryPersistedLegs`는 이어받은 레그의 `Gate`를 저장된 값 그대로 복사한다(`audit_rerun_legs.go:150-155`). `mcp_convergence.go:1084-1091`의 gate 재적용은 claude를 건너뛰고, `enforceRequiredGateUnmet`은 `inconclusive`만 처리한다(`:1328`). 그래서 claude가 `advisory/fail`로 저장된 뒤 현재 호출이 claude를 `required`로 두면, codex가 pass일 때 전체가 pass가 된다. 여기까지는 사실이다.

그런데 이 경로로 호출자가 새 권한을 얻지는 않는다.

- 첫째, `ResolveAuditPlan`은 호출자가 넘긴 gate를 설정값보다 먼저 쓴다(`internal/config/audit_plan.go:222-223`). 그래서 호출자는 두 번째 호출에 `gates: {claude: advisory}`만 넘겨도 같은 결과를 바로 얻는다.
- 둘째, `workflow.yaml`은 추적 파일이다(`git ls-files` 출력 있음). 그래서 설정을 바꾸면 head나 tree digest가 달라지고, 이어받기 자체가 거부된다(`audit_rerun_legs.go:135-138`).
- 셋째, codex 레그는 명시 설정이 있으면 `explicitGateFor`로 다시 도장이 찍힌다(`:1088-1089`).

결함은 맞다. 이어받은 레그의 gate가 현재 호출의 계획과 어긋난 채 보고된다. 다만 SPEC REQ-AEC-008이 이어받는 대상으로 정한 것은 "verdict"이므로 수정할 가치가 있다. 그래서 SHOULD-FIX, optional로 유지한다.

## 결함 목록

| id | 심각도 | 분류 | 위치 태그 | 위치 | 내용 | Required fix / 상태 |
|---|---|---|---|---|---|---|
| D16 | High | **blocking** | touched-file (D1 수리 d9665c52c5가 넣은 코드, 카드 diff 안) | `internal/auditreceipt/store.go:1934-1939` (`e2Sections`), `:1951-1964` (`fenceRun`) | `fenceRun`은 백틱 3개 이상으로 시작하는 줄을 모두 fence opener로 받는다. CommonMark에서는 백틱 fence의 info string에 백틱이 있으면 fence가 아니다(인라인 코드). §E.2 안에 그런 줄이 있으면 뒤의 정상 opener를 closer로 읽고, 코드 블록 안의 `## `에서 §E.2가 끝난다. 그 뒤의 §E.2 수정은 비교에서 빠진다. 원문 헤딩 1 = 파싱 섹션 1이라 D10 가드도 통과한다. | 아래 "권고 수리"를 따른다. 최소 수리라면 백틱 opener의 info string(`rest`)에 백틱이 있으면 opener로 받지 않는다. 그리고 §E.2 안에 inline 줄과 `## `를 담은 정상 블록이 함께 있는 fixture를 AC-AEC-003b 테스트(`internal/hook/audit_receipt_e2_shapes_test.go`와 `store_e2_failclosed_test.go`)에 추가한다. |
| D13 | Medium→High (재분류) | **blocking** (3회차 optional에서 변경) | in-diff | `store.go:1919-1946` | §E.2 안의 HTML `<pre>` 블록에 `## ` 줄이 있으면 거기서 §E.2가 끝나고, 그 뒤의 수정이 변경 없음으로 판정된다. | D16과 같은 구조적 수리로 함께 닫는다. 재분류 근거: REQ-AEC-010 문구 위반이 재현되었고, D16·D1과 효과와 전제가 같다. |
| D2 | SHOULD-FIX | optional | in-diff | `internal/cli/audit_rerun_legs.go:150-155`, `mcp_convergence.go:1084-1091` | 이어받은 레그가 저장된 gate를 유지한다(Codex P1 #2) | 이어받을 때 `Verdict`, `Summary`, `Findings` 등만 가져오고 `Gate`는 현재 호출의 `backends` 값으로 다시 매긴다. claude도 포함한다. |
| D5 | MINOR | optional | in-diff | `internal/cli/verify_clear_rejection.go:195` | `.moai/reports` 자체가 symlink이면 containment 기준이 트리 밖이 된다(Codex P2) | 열림. 해석된 reports 디렉터리가 canonical 트리 안에 있는지 확인한다. |
| D6 | SHOULD-FIX | optional | in-diff | `mcp_convergence.go:1185` | 일부 레그만 재실행하면 이전 영수증이 그대로 인용되어 거부된다(fail-closed, Codex P2) | 열림 |
| D17 | MINOR | optional | in-diff | `store.go:1856-1857` | 감사 시점에 허용 증거가 dirty였다가 커밋되면 digest가 달라져 stale이 된다(fail-closed, Codex P2). REQ-AEC-009 문구와 일치한다 | 신규. 재감사 비용만 늘어난다. 원하면 SPEC을 개정해 허용 증거 경로를 digest에서 뺀다. |
| D3, D4, D7, D8, D9, D11, D12, D14, D15 | 3회차 기록 | optional | — | 3회차 표 참조 | 이번에 재측정하지 않음 | 열림(3회차 기록 그대로, 이번 실행의 증거는 아님) |
| D1 | (High였음) | 닫힘 | in-diff | `store.go:1934-1939` | 중첩 fence가 §E.2를 가림 | 닫힘. 닫는 조건을 항상 참인 토글로 바꾼 뮤턴트에서 `TestE2Sections_FenceTrackingIsLengthAndCharacterAware`, `TestE2DocsUnchanged_PositiveControls` FAIL |
| D10 | (High였음) | 닫힘 | in-diff | `store.go:1887-1893` | 빈 목록끼리 같다고 판정 | 닫힘. 가드를 `if false {`로 바꾼 뮤턴트에서 `TestE2DocsUnchanged_E2BodyEditIsAlwaysAChange` FAIL |

### 권고 수리 (D16 + D13, 구조적)

이 파서에서 같은 결과(§E.2 변경을 놓침)를 내는 결함이 D1, D10, D16, D13, D14로 이어졌다. Markdown 구성을 하나씩 더 따라가는 패치는 다음 구성(들여쓴 코드, 다른 HTML 블록 등)에서 또 어긋날 수 있다. 그래서 fence 추적을 정교하게 만들기보다 모호하면 stale로 떨어지게 바꾸기를 권한다.

- (권장) §E.2 영역을 fence나 HTML 추적 없이 원문 그대로 잡는다. 범위는 정규 `## §E.2` 헤딩 줄부터 다음 원문 `^## §E\.[3-9]` 줄(없으면 문서 끝)까지다. 양쪽 버전에서 그런 경계 줄이 정확히 하나씩 있을 때만 비교하고, 아니면 stale로 처리한다. 코드 블록 안의 일반 `## ` 줄은 영역을 끊지 않으므로 늦게 끊기는 쪽(fail-closed)으로만 틀린다.
- (대안) 지금 파서를 유지한다면 두 가지를 더한다. 백틱 info-string 규칙(D16)을 넣고, §E.2 영역 안에 HTML 블록 시작 줄(`<pre`, `<!--`, `<script`, `<style` 등)이 있으면 stale로 처리한다(D13).

반복 수리 중지 규칙: 직전 감사 SHA 400caf47ba 이후의 diff는 판정 파일 하나뿐이다. 그래서 D16은 정의상 repair-born이 아니고, 연속 2회 조건도 충족되지 않는다. 다만 D16은 D1 수리 코드 안에서 3회차 감사가 놓친 결함이다. 위 권고는 그래서 패치보다 재설계 쪽으로 기울어 있다.

## Claim / Evidence / Baseline-attribution / Gaps / Residual-risk

### Claim

1. HEAD는 7ea9b20555이고 트리는 깨끗하다. 400caf47ba와의 차이는 판정 파일 하나다.
2. 카드 diff는 merge-base 9b632c1eb3 대비 81개 파일이고, codex가 리뷰한 범위와 같다.
3. D1과 D10은 닫혔다.
4. D16과 D13은 REQ-AEC-010을 위반하는 재현된 차단 결함이다.
5. 지정 테스트, lint, 포맷, emit·manifest 검사는 통과한다.
6. codex 필수 게이트는 fail을 보고했다.

### Evidence (명령 → exit → 판정에 쓴 출력)

- `git rev-parse HEAD` → `7ea9b205552e41f3381d266fee144fbc95a06f6d`. `git status --short` → 출력 없음(감사 시작 시, 판정 파일 작성 직전 모두).
- `git diff 400caf47ba 7ea9b20555 --stat` → `.moai/reports/t2017/sync-audit-3.md | 147 +++`, `1 file changed`.
- `git merge-base HEAD main` → `9b632c1eb3544cef5f275e502922a375e48d20aa`. `git diff --shortstat main...HEAD` → `81 files changed, 9940 insertions(+), 221 deletions(-)`. `git-strategy.yaml:5` `worktree_base_branch: main`.
- `moai verify audit-plan --project-root <toplevel>` → `config_status: ok`, `model: codex`, codex `required` (`config.gates`, explicit), claude/glm `off`, `enforced_required: ["codex"]`.
- `mcp__moai__codex_audit` (1회) → `verdict: fail`, `audit_receipt: rcpt-919c8ec5ced34525ba7663b0`, `review_base: main (merge base 9b632c1eb3…)`, 지적 5건(위 원문).
- `go test -count=1 ./internal/auditreceipt/` → exit 0, `ok github.com/modu-ai/moai-adk/internal/auditreceipt 30.457s`.
- `go test -count=1 -run 'TestAEC003a_|TestAEC003b_|TestE2' ./internal/hook/` → exit 0, `ok github.com/modu-ai/moai-adk/internal/hook 45.372s`.
- `go test -count=1 -run 'TestAEC|TestClearRejection_|TestPerformCodexAudit_' ./internal/cli/` → exit 0, `ok github.com/modu-ai/moai-adk/internal/cli 280.467s`.
- `make agents-emit-check` → exit 0, `agents-emit-check: specs=12 copies=4 diff=0`.
- `make embed-manifest-check` → exit 0, `ok github.com/modu-ai/moai-adk/internal/template/embedemit 0.335s`.
- `golangci-lint run --timeout=5m ./internal/auditreceipt/... ./internal/hook/... ./internal/cli/...` → exit 0, `0 issues.`. `gofmt -l internal/auditreceipt internal/hook internal/cli` → 출력 없음.
- `go test -count=1 -cover ./internal/auditreceipt/` → exit 0, `coverage: 87.1% of statements`.
- D16 프로브(스크래치 사본에 `zz_auditor4_probe_test.go`): `go -C <scratch> test -count=1 -v -run TestAuditor4Probe ./internal/auditreceipt/` → exit 1.
  - `probe A (codex P1 shape): e2DocsUnchanged=true raw(old)=1 secs(old)=1`
  - `MISMATCH: an E.2 row edit (PASS->FAIL) read as unchanged`
  - 대조군: `probe B (control, no inline line): e2DocsUnchanged=false`
- D13 프로브(같은 사본): `-run 'TestAuditor4Probe_D13'` → exit 1, `probe C (D13 html pre): e2DocsUnchanged=true`, `MISMATCH: D13 shape reads an E.2 edit as unchanged`.
- D10 뮤턴트(사본 `store.go`의 가드를 `if false {`로 바꿈, `grep -c` → 1): `-run 'TestE2DocsUnchanged|Test.*E2'` → exit 1, `--- FAIL: TestE2DocsUnchanged_E2BodyEditIsAlwaysAChange`. 원본 복원.
- D1 뮤턴트(사본 `:1937`의 닫는 조건을 항상 참으로 바꿈. 첫 시도는 미사용 변수로 build failed여서 컴파일되는 형태로 다시 만듦): `go -C <scratch> test -count=1 ./internal/auditreceipt/` → exit 1, `--- FAIL: TestE2Sections_FenceTrackingIsLengthAndCharacterAware`, `--- FAIL: TestE2DocsUnchanged_PositiveControls`. 원본 복원.
- 400caf47ba 읽기: `git show 400caf47ba -- internal/cli/audit_delta_scope_test.go`. 프롬프트가 `codexAdversarialReviewPrompt("")`로 시작하고 `canary_c_delta`를 포함하며 `canary_b_preexisting`은 포함하지 않는지 단언한다. native `review/start`가 아닌지도 확인한다. 해당 테스트는 위 cli 실행(`TestPerformCodexAudit_` 선택)에 포함되어 통과했다.
- 코퍼스: `grep -rlE '^ {0,3}`{3,}[^`]+`' .moai/specs --include=progress.md` → 출력 없음(0개 파일, SPEC 디렉터리 1139개). 첫 정규식은 4-backtick fence 2건을 잘못 잡아 버렸다.
- D2 근거 읽기: `audit_rerun_legs.go:117-157`, `mcp_convergence.go:930-1129`, `:1325-1358`, `mcp_audit_multi.go:157-178`, `internal/config/audit_plan.go:188-232`. `git ls-files .moai/config/sections/workflow.yaml` → 경로 출력됨(추적 파일).

### Baseline-attribution

- 측정 트리: 7ea9b205552e41f3381d266fee144fbc95a06f6d(작업 트리 깨끗, 이번 실행).
- 프로브와 뮤턴트: 같은 커밋을 `git archive -o head.tar HEAD`로 스크래치(`…/scratchpad/a4/src`)에 풀어 실행했다. 추적 트리에는 아무것도 쓰지 않았다.
- Go 테스트, lint, 빌드: 이 트리의 소스를 로컬 `go` 툴체인으로 컴파일해 측정했다.
- `moai verify audit-plan`은 설치본 바이너리, codex 호출은 MCP 서버 빌드 `646955327e`(v3.2.0-rc.42)로 했다. 두 값은 감사 계획과 codex 결과를 읽은 것이고, 코드 판정 근거는 위 Go 측정이다.

### Gaps

- G1. `audit_multi`는 운영자 결정에 따라 호출하지 않았다. 그래서 convergence digest와 orchestrator의 `--result-file` 확인이 없다.
- G2. codex 재호출은 하지 않았다(1회 한도). codex가 스스로 밝힌 대로, codex 쪽 Go 시험은 실행되지 않았다.
- G3. 첫 D16 프로브 실행은 cwd가 워크트리라서 `[no tests to run]`이었다. 이는 판정에 쓰지 않았고, `go -C <scratch>`로 다시 실행한 결과를 썼다. D1 뮤턴트 첫 시도는 build failed였고 판정에 쓰지 않았다.
- G4. D16과 D13은 순수 함수 `e2DocsUnchanged` 수준에서 재현했다. 실제 git 범위와 Stop 훅을 거치는 end-to-end 재현은 하지 않았다.
- G5. D2는 코드 읽기로만 확인했다. 실행 재현은 하지 않았다. `moai verify audit-plan --result-file`의 convergence_check가 이어받은 레그의 gate 불일치를 잡는지는 측정하지 않았다.
- G6. 실행하지 않은 것: 9분짜리 cli 넓은 선택, 각 패키지 전체 스위트, race detector, Windows 빌드와 테스트, auditreceipt 밖 커버리지, `/verify`.
- G7. D3·D4·D7·D8·D9·D11·D12·D14·D15는 다시 측정하지 않았다(3회차 기록만 인용).
- G8. 거부된 명령은 없다. 모든 git 명령은 단일 명령으로 실행했다.

### Residual-risk

- 권장 수리(원문 영역 비교)를 적용해도 §E.2 영역 안의 `## §E.3` 같은 문자열은 코드 블록 안이라도 경계로 읽힌다. 경계 줄 개수 가드로 fail-closed 처리할 수 있지만, 그 가드 자체도 테스트로 고정해야 한다.
- 판정 파일의 `blocking_defect:` id를 감사자 자신이 쓴다는 점(§E B1)은 이번 감사로 바뀌지 않았다.
- 스크래치의 프로브 출력 파일(`…/scratchpad/a4/*.txt`)은 기계 로컬이다. 판정에 쓴 줄은 위 Evidence에 옮겨 적었고, 나머지는 소실을 감수한다.

## 반복 이력

| 회차 | 감사 커밋 | 판정 | 주요 변화 |
|---|---|---|---|
| 1 | 1008b915b2 | FAIL (D1 BLOCKING) | fence boolean 토글이 §E.2를 가릴 수 있었다 |
| 2 | d9665c52c5 | FAIL 0.73 (D10 BLOCKING) | D1 입력은 닫혔지만 빈 목록 비교가 남았다 |
| 3 | 400caf47ba | PASS-WITH-DEBT 0.85, receipts none | Claude 단독. 영수증 가드가 거부했다(codex required) |
| 4 | 7ea9b20555 | FAIL 0.73 (D16, D13 BLOCKING), codex fail `rcpt-919c8ec5ced34525ba7663b0` | 운영자 승인 codex 1회. D1·D10 닫힘 재확인, D16 신규 재현, D13 재분류 |

## 다음 단계 (리더·운영자 판단 사항)

- D16과 D13을 고치는 수리 커밋을 낸 뒤 확인 감사를 받으려면 codex를 다시 호출해야 한다. 운영자 결정 d-20261106T055051Z-db22는 카드당 1회만 허용하므로, 재호출에는 새 운영자 승인이 필요하다.
- 확인 감사의 범위는 7ea9b20555 이후의 델타로 잡을 수 있다(blocking id: D16, D13).

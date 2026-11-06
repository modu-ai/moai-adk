auditor-model: claude-opus-5-5
verdict: FAIL
audited_sha: 985853c683565e6f765ad17608d4526760804333
overall_score: 0.75
functionality_score: 74
security_score: 64
craft_score: 82
consistency_score: 85
must_pass_failed: 2
blocking_count: 1
blocking_defect: D19
required_backend_fail: codex
receipts: rcpt-7911463ec3cd51004d46c029

# Sync 감사 판정 — SPEC-AUDIT-ENGINE-CONVERGENCE-001 (카드 t2017), 5회차 (수리 델타 재감사)

## 요약

판정은 FAIL, 가중 점수 74.75(0.75)다. 차단 결함은 1건(D19)이다.

- **D16, D13은 닫혔다.** 같은 프로브를 두 커밋에 돌렸다. 7ea9b20555에서는 두 형태 모두 `unchanged=true`(fail-open)였고, 985853c683에서는 `unchanged=false`다.
- **D19 (신규, 수리에서 생김, Codex P1과 같은 내용)**: 원문 영역 비교는 `## §E.N` 원문 줄을 그대로 경계로 읽는다. 그래서 §E.2 안의 코드 블록이 `## §E.3 …`을 인용하고, 그 숫자가 뒤따르는 실제 경계 숫자보다 작거나 뒤에 실제 경계가 아예 없으면, §E.2가 인용 줄에서 끝난다. 그 아래 AC 행을 PASS에서 FAIL로 고쳐도 "변경 없음"이 된다. 7ea9b20555의 fence 파서는 이 형태 넷(R1~R4)을 모두 `changed`로 읽었으므로 이번 수리가 만든 회귀다. 작성자가 코드 주석에 잔여 위험으로 적어 둔 바로 그 경우이고, 이번 실행에서 재현했다.
- **D18 (신규, Codex P2, optional)**: fail-closed 방향의 가용성 회귀다. 실제 progress.md 851개(§E.2 포함) 중 73개가 자기 자신과 비교해도 `changed`가 된다. 이 SPEC들은 §E.3/§E.4만 고쳐도 영수증이 항상 stale이 된다. 7ea9b20555에서는 0개였다.
- codex는 `fail`이고 영수증은 `rcpt-7911463ec3cd51004d46c029`다. 이 트리는 codex 게이트가 `required`이므로 PASS 계열 판정은 나올 수 없다.

## 차원별 점수 (default 프로파일, flat 가중)

| 차원 | 점수 | 판정 | 근거 |
|---|---|---|---|
| Functionality (40%, must-pass) | 74/100 | FAIL | 지정 시험은 모두 통과했다(auditreceipt `ok 60.602s`, hook 6개 최상위 `ok 110.118s`). D16·D13과 D1·D10 형태도 닫혀 있다. 그러나 유효한 Markdown 입력에서 REQ-AEC-010("`§E.2` 아래 한 줄만 바뀌어도 stale")이 깨진다(D19, 재현). must-pass 차원에는 부분 점수를 주지 않는다. |
| Security (25%, must-pass) | 64/100 | FAIL | 영수증 신선도 게이트를 우회하는 fail-open 경로가 다시 생겼다(D19). D1·D16과 결과와 전제가 같은 부류다. |
| Craft (20%) | 82/100 | PASS | golangci-lint `0 issues.`, go vet exit 0, auditreceipt 커버리지 87.1%다. 새 규칙(경계 숫자 엄격 증가)을 끈 뮤턴트를 테스트 4개가 잡았다. 잔여 위험을 코드 주석에 정직하게 적었다. |
| Consistency (15%) | 85/100 | PASS | `agents-emit-check: specs=12 copies=4 diff=0`, `embed-manifest-check` ok, Windows 교차 빌드 exit 0이다. 함수 이름과 주석은 파일 관례를 따른다. |

가중 합: 74×0.40 + 64×0.25 + 82×0.20 + 85×0.15 = 74.75. must-pass 두 차원이 기준 미달이라 firewall에 따라 FAIL이다.

## 백엔드별 결과

- **codex** (gate `required`, `mcp__moai__codex_audit`, mode `adversarial`, target `baseBranch`, `delta_base: 7ea9b205552e41f3381d266fee144fbc95a06f6d`, resolved_model `gpt-6.1-sol`, build_commit `646955327e`): verdict **`fail`**, receipt **`rcpt-7911463ec3cd51004d46c029`**, 지적 2건.
  - 범위 형태: 도구 스키마에 `delta_base` 인자가 있어 그것을 썼다. 반환된 `review_base`는 `delta_base 7ea9b205552e41f3381d266fee144fbc95a06f6d (prose-only scope: codex adversarial mode cannot be bounded mechanically)`이다. 즉 adversarial 모드에서는 범위가 프롬프트로만 전달되었다. focus 인자로 수리 델타와 store.go의 e2 영역을 함께 지정했다.
  - 1. [P1] `store.go:1963` "중복되지 않는 인용 경계가 E.2 변경을 숨긴다" → **D19. 동의한다. Go로 재현했다**(codex는 Python 대조만 했다고 스스로 밝혔다).
  - 2. [P2] `store.go:1983` "정상 하위 제목이 있으면 E.3/E.4만 수정해도 영수증이 무효화된다" → **D18. 동의한다. 코퍼스로 규모를 쟀다**(73/851). fail-closed 방향이라 optional로 둔다.
  - 이번이 이 카드의 두 번째이자 마지막 codex 호출이다(운영자 결정 d-20261106T064700Z-de2d). 4회차 fail 영수증 `rcpt-919c8ec5ced34525ba7663b0`는 인용하지 않는다.
- **claude**: 계획에서 gate `off`. 이 감사(Claude, 자체 검토)가 아래 증거를 독립적으로 모았다.
- **glm**: gate `off`, 호출하지 않았다.
- `moai verify audit-plan`은 `cross_model_active: true`를 보고했다. 그러나 운영자 결정에 따라 `audit_multi`는 호출하지 않았다. 그래서 convergence digest(`overall_verdict`, `gate_unmet`, `plan_source`)가 없고, orchestrator가 `--result-file`로 확인할 대상도 없다(Gap G1).

## 결함 목록

| id | 심각도 | 분류 | 위치 태그 | 위치 | 내용 | Required fix / 상태 |
|---|---|---|---|---|---|---|
| D19 | High | **blocking**, repair-born | in-diff | `internal/auditreceipt/store.go` `e2RawRegion` (경계 결정 `:1957-1966`) | §E.2 안의 fenced block이나 HTML 주석이 `## §E.N`을 인용하고, 그 숫자가 뒤따르는 실제 경계보다 작거나 뒤에 실제 경계가 없으면, 인용 줄이 경계가 되고 그 아래 §E.2 수정이 숨겨진다. 재현된 형태: R1(인용 E.3, 실제 E.4만 있음), R2(인용 E.3, §E.2가 마지막 절), R3(인용 E.9, §E.2가 마지막), R4(인용 E.3, 실제 E.3가 한 칸 들여씀), R6(HTML 주석 안 인용 E.3, 실제 E.4만). R1~R4는 7ea9b20555에서 `changed`였다. R6은 두 커밋 모두 fail-open이었다. | 기존 검사를 조인다(REQ-AEC-015 허용 범위). **§E.2 원문 영역이 열린 구성 안에서 끝나면 모호(changed)로 처리한다.** 열린 구성이란 영역 안의 fence 유사 줄(`^ {0,3}(```|~~~)`)이 홀수 개이거나, `<!--`·`<pre` 개수가 `-->`·`</pre` 개수보다 많은 경우다. 이번 실행에서 잰 결과 이 규칙은 R1~R4와 R6을 모두 잡는다. 실제 코퍼스에서 추가로 항상-stale이 되는 문서는 774개 중 6개다. R1~R4, R6을 `TestE2DocsUnchanged_RawRegion_AmbiguityFailsClosed`와 hook 형태 시험에 추가한다. 이 규칙으로도 남는 잔여(짝수 개의 중첩 fence 줄 안에 있는 인용 경계)는 코드 주석과 SPEC §E에 잔여 위험으로 적는다. 이 경로 대신 잔여를 수용하려면 SPEC §E 개정과 운영자 결정이 필요하다. |
| D18 | SHOULD-FIX | optional | in-diff | `store.go` `e2HeadingLike`, 엄격 증가 규칙, 중복 헤딩 규칙 | 실제 progress.md 73/851이 자기 자신과 비교해도 `changed`다(7ea9b20555에서는 0). 원인은 세 가지다. heading-like 56건(예: `## §E.2.1 AC PASS/FAIL 매트릭스`, `### 여전히 **열려 있는** Gap (§E.2 에서 인계)`, `## §E.2b …`, `### §E.2a …`), 경계 비증가 12건(숫자열 `334`, `344`, `3445`), §E.2 헤딩 여러 개 5건. 이 문서들은 §E.3/§E.4만 바뀐 증거 커밋에서도 영수증이 stale이 되므로, 실제 문서에서 REQ-AEC-009의 fresh 조건이 성립하지 않는다. AC-AEC-003a의 fixture는 통과한다. | `#`로 시작하고 `§E.2`를 어디든 포함하는 줄 전부가 아니라, 헤딩 텍스트가 `§E.2`로 시작하는 줄만 heading-like로 본다. 이렇게 하면 3단계 헤딩의 단순 언급이 빠진다. `## §E.2.1`과 `## §E.2b`는 모호로 남겨도 된다. 이전 파서는 이 형태에서 §E.2를 거기서 끊어, 그 아래 수정을 놓쳤다(오히려 fail-open이었다). 비용만 늘어나므로 optional로 둔다. |
| D2–D9, D11, D12, D14, D15, D17 | 3·4회차 기록 | optional | — | 3·4회차 표 참조 | 이번에 재측정하지 않았다 | 열림(이전 기록을 그대로 옮김, 이번 실행의 증거 아님) |
| D16 | (High였음) | 닫힘 | in-diff | `store.go` | 인라인 코드 줄 + `## `를 담은 fence가 §E.2를 끊음 | 닫힘. 프로브에서 7ea9b20555는 `unchanged=true`, HEAD는 `unchanged=false` |
| D13 | (High였음) | 닫힘 | in-diff | `store.go` | `<pre>` 안의 `## `가 §E.2를 끊음 | 닫힘. 프로브에서 7ea9b20555는 `unchanged=true`, HEAD는 `unchanged=false` |
| D1, D10 | (High였음) | 닫힘 유지 | in-diff | `store.go` | 중첩 fence / 빈 목록 비교 | 15개 형태의 표(`d10Shapes`: D1 중첩 fence 3종과 오독 형태 10종 포함)에서 `TestE2DocsUnchanged_E2BodyEditIsAlwaysAChange`가 PASS |

### D19가 blocking인 근거 (잔여로 수용하지 않는 이유)

- **기준이 같다.** 1회차 D1, 4회차 D16·D13은 유효한 Markdown 입력에서 REQ-AEC-010을 깨는 형태라서 blocking이었다. 4회차는 D16 형태가 코퍼스에 0건이었는데도 blocking으로 판정했다. D19도 유효한 Markdown이다(닫힌 fence 안의 인용 줄). 이번 코퍼스에서도 해당 형태는 0건이었다(원문 영역이 fence 안의 경계에서 끝나는 문서 0/774). 앞선 판정과 같은 기준을 적용하면 blocking이다.
- **위협 모델이 이 판정을 바꾸지 않는다.** spec.md §E의 cooperating-process 모델은 "서버가 쓴 파일을 위조하는 프로세스"를 잔여 위험으로 둔다. 그런데 D19는 위조가 아니다. 협조적인 작성자가 정상적으로 쓴 문서에서도 그대로 일어난다. D16도 같은 모델 아래에서 blocking이었다.
- **수리에서 생겼다.** R1~R4는 직전 감사 SHA 7ea9b20555에서 `changed`였고, 이번 diff가 만든 회귀다. 4회차의 blocking(D16, D13)은 repair-born이 아니었다. 그래서 연속 2회 조건은 아직 충족되지 않았고, 반복 수리 중지 규칙은 발동하지 않는다. 다만 D1 → D10 → D16·D13 → D19로 같은 함수의 fail-open이 네 번 이어졌다. 다음 수리는 위의 "열린 구성" 조임처럼 영역을 넓히지 않고 모호를 늘리는 방향으로만 하기를 권한다.

### 레인의 두 해석 판단

- (a) **양쪽 모두 §E.2가 없고 heading 유사 줄도 없으면 unchanged** → 받아들인다. REQ-AEC-009 (b)는 "양쪽 버전에서 계산한 §E.2가 같음"을 요구하는데, 양쪽 모두 없으면 같다. 한쪽에만 있으면 `changed`가 되므로 REQ-AEC-010의 헤딩 삭제와 이름 변경을 잡는다(`TestE2DocsUnchanged_RawRegion_PresenceAndPosition` PASS). 잔여도 하나 있다. `## &sect;E.2`나 `## §E\.2`처럼 정규 패턴에 맞지 않게 쓴 헤딩은 양쪽 모두 "없음"으로 읽혀, 그 아래 수정이 보이지 않는다. 코퍼스에는 이런 헤딩이 0줄이다. era.go 같은 다른 소비자도 이런 표기를 §E.2로 인식하지 않으므로 잔여로 둔다.
- (b) **경계 숫자는 E.3 → E.4 순으로 엄격히 증가해야 하고, 반복이나 역순은 모호** → fail-closed 방향이라 안전하고, 뮤턴트가 테스트 4개로 고정되어 있음을 확인했다. 그러나 충분하지는 않다. 인용 숫자가 뒤의 실제 숫자보다 작거나 뒤에 실제 경계가 없으면 증가 조건을 그대로 통과한다(D19). 그리고 실제 문서 12건이 이 규칙 때문에 항상 stale이다(D18).
- **R9(§E.2 안의 미종결 fence가 실제 E.3를 삼키는 경우)는 blocking이 아니다.** Markdown 기준으로는 원문 `## §E.3` 줄 아래가 §E.2의 코드 내용이다. 하지만 이 문서는 형식이 깨진 문서다. 저장소의 다른 소비자(era.go)도 그 원문 줄을 §E.3로 읽는다. 그리고 수정이 일어난 곳은 원문 기준 §E.3이며, REQ-AEC-009가 허용하는 범위다. 잔여 위험으로 기록한다.

## 바뀐 테스트 단언 (7ea9b20555 대비)

| 테스트 (이전 → 현재) | 변화 | 약해졌나 |
|---|---|---|
| `TestE2Sections_ParsesHeadingOrdinalAndBody` → `TestE2RawRegion_ParsesHeadingOrdinalAndText` | heading과 body 비교를 heading을 포함한 text 비교 하나로 합침 | 아니다(동등) |
| `TestE2Sections_OrdinalAndBoundary` → `TestE2RawRegion_OrdinalAndBoundary` | `e2DocsUnchanged`로 "위에 헤딩 삽입 → changed" 단언을 추가. E.3 재작성은 섹션 동일성 대신 `e2DocsUnchanged == true`로 단언 | 아니다(강해짐) |
| `TestE2Sections_FencedHeadingDoesNotEndTheSection` → `TestE2RawRegion_…` | 편집 판정을 구조체 비교에서 `e2DocsUnchanged`로 바꿈 | 아니다(동등) |
| `TestE2Sections_FenceTrackingIsLengthAndCharacterAware` → `TestE2RawRegion_FenceShapesNeverCutTheSection` | 하위 테스트 이름에서 "does not close it" 같은 fence 의미를 뺐다. 단언은 "헤딩을 찾고 뒤 텍스트를 유지한다"로 같다. 첫 하위 테스트는 섹션 수 1 대신 present/ok를 본다 | fence 의미 자체는 더는 검사하지 않는다. 그러나 운영 코드에서 fence 파서를 지웠으므로 의도된 변화이고, 운영 함수의 계약 기준으로는 약해지지 않았다 |
| `TestE2Sections_DuplicateHeadings` → `TestE2RawRegion_DuplicateHeadings` | "두 섹션을 모두 찾는다"에서 "모호하므로 동일 문서도 changed"로 바뀜 | 아니다(fail-closed로 강해짐, D18 비용의 한 원인) |
| `TestE2DocsUnchanged_E2BodyEditIsAlwaysAChange` | 단언은 그대로, 실패 메시지만 바뀜 | 아니다 |
| `TestE2DocsUnchanged_PositiveControls` 하위 테스트 "E.2 quoted inside a fence only" | 이전과 같이 단언 없이 로그만 남김 | 바뀌지 않음(원래부터 단언 없음) |

삭제된 테스트 함수는 없다. 운영 코드에서 지운 `e2Section`·`fenceRun`은 비교표 테스트용으로만 테스트 파일에 복사되어 있다.

## AC-AEC-003a 양성 대조

- §E.3/§E.4만 고친 수정은 fresh를 유지한다. 근거: `TestE2DocsUnchanged_RawRegion_InsideShapeControls`(6형태 × E.3, E.4) PASS, hook `fresh:` 하위 테스트 6개 PASS, `TestAEC003a_EvidenceOnlyCommits_KeepTheReceipt` PASS, `TestAEC003a_UncommittedStateAlreadyJudged_KeepsTheReceipt` PASS.
- `evidenceOnlyRange`는 바뀌지 않았다. store.go diff의 hunk는 정규식 변수 추가(`e2BoundaryPattern`), `e2Unchanged` 주석, `e2DocsUnchanged`부터 `e2HeadingLike`까지에 한정된다. `evidenceOnlyRange` 본문에는 diff가 없다.
- 예외: 실제 문서 73개에서는 양성 대조가 성립하지 않는다(D18).

## Claim / Evidence / Baseline-attribution / Gaps / Residual-risk

### Claim

1. 감사 대상은 HEAD 985853c683이고 트리는 깨끗하다.
2. D16, D13은 닫혔다. D1, D10은 닫힌 상태를 유지한다.
3. D19: 수리가 만든 fail-open 형태가 있고, 재현했다.
4. D18: 실제 문서 73/851이 항상 stale이 된다.
5. 지정 명령은 모두 exit 0이다.
6. codex 필수 게이트는 fail을 보고했다.

### Evidence (명령 → exit → 판정에 쓴 출력)

- `git rev-parse --show-toplevel` → `/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t2017`. `git rev-parse HEAD` → `985853c683565e6f765ad17608d4526760804333`. `git status --short` → 출력 없음(시작 시와 판정 파일 작성 직전 모두).
- `git show --stat 985853c683` → `5 files changed, 374 insertions(+), 122 deletions(-)`.
- `moai verify audit-plan --project-root <toplevel>` → `config_status: ok`, `model: codex`, codex `required`(config.gates, explicit), claude/glm `off`, `cross_model_active: true`, `enforced_required: ["codex"]`.
- `moai decision read --scope standing` → `d-20261106T064700Z-de2d` 본문 "3.2 카드당 MCP codex_audit 최대 2회 — 첫 감사 1회 + 수리 … 뒤 재감사 1회(수리 델타 한정) … audit_multi 금지, fail 영수증 재사용 금지", `supersedes: d-20261106T061529Z-5d7e`.
- `mcp__moai__codex_audit`(1회, adversarial, baseBranch, delta_base 7ea9b20555…) → `verdict: fail`, `audit_receipt: rcpt-7911463ec3cd51004d46c029`, `resolved_model: gpt-6.1-sol`, 지적 P1 `store.go:1963`, P2 `store.go:1983`.
- 프로브 HEAD(스크래치 사본, `go -C <new> test -count=1 -v -run TestAuditor5 ./internal/auditreceipt/`) → exit 0:
  - `D16 inline-code line then fence holding ## line  unchanged=false ok (changed)`
  - `D13 html pre holding ## line  unchanged=false ok (changed)`
  - `R1 quoted ## §E.3 in fence, real E.4 only (codex P1)  unchanged=true  FAIL-OPEN`
  - `R2 quoted ## §E.3 in fence, E.2 is the last section  unchanged=true  FAIL-OPEN`
  - `R3 quoted ## §E.9 in fence, E.2 last  unchanged=true  FAIL-OPEN`
  - `R4 quoted ## §E.3 in fence, real E.3 indented one space  unchanged=true  FAIL-OPEN`
  - `R6 quoted ## §E.3 in html comment, real E.4 only  unchanged=true  FAIL-OPEN`
  - `R9 unterminated fence in E.2, edit after raw E.3  unchanged=true`
  - `P2 ### §E.2.1 sub-heading, edit confined to E.3  unchanged=false`
  - `corpus files=978 containing §E.2=851 self-compare-changed=73`
- 같은 프로브를 7ea9b20555 사본에서 실행 → exit 0: `D16 … unchanged=true FAIL-OPEN`, `D13 … unchanged=true FAIL-OPEN`, R1·R2·R3·R4는 `unchanged=false`, R6은 `unchanged=true`, R9는 `unchanged=false`, P2는 `unchanged=true`, `corpus … self-compare-changed=0`.
- 진단(HEAD 사본, `-run TestAuditor5Diag`) → exit 0: `heading-like line 56`, `boundaries not strictly rising 12`, `several E.2 headings 5`, `unambiguous with E.2: region ends at a fenced boundary line=0 []; E.2 runs to document end=0`.
- 수리안 비용(HEAD 사본, `-run TestAuditor5FixCost`) → exit 0: 영역 안에 fence/HTML 유사 줄이 하나라도 있으면 모호로 보는 안 → `unambiguous docs with E.2=774; region holds a fence/HTML-like line=321` (비용이 너무 커서 기각). `escaped E.2 heading lines=0`.
- 수리안 비용(HEAD 사본, `-run TestAuditor5FixCostB`) → exit 0: "열린 구성 안에서 끝남" 안 → `unambiguous=774 would-flip=6`. 이 안은 R1, R2, R3, R4, R6에서 모두 `endsOpen=true`, control에서 `endsOpen=false`다.
- 뮤턴트(HEAD 사본 `store.go:1958`의 엄격 증가 조건에 `&& false`를 붙임): `go -C <new> test -count=1 -run TestE2 ./internal/auditreceipt/` → exit 1, `--- FAIL: TestE2DocsUnchanged_RawRegion_AmbiguityFailsClosed`(하위 3개), `--- FAIL: TestE2DocsUnchanged_RawRegion_QuotedBoundaryCannotHideALaterEdit`. 확인 뒤 원상 복구했다.
- `go vet ./internal/auditreceipt/ ./internal/hook/` → exit 0, 출력 없음.
- `go test -count=1 -cover ./internal/auditreceipt/` → exit 0, `ok … 60.602s coverage: 87.1% of statements`.
- `go test -count=1 -run 'TestE2DocsUnchanged_RawRegion_|TestE2RawRegion_|TestE2DocsUnchanged_E2BodyEditIsAlwaysAChange|TestE2DocsUnchanged_PositiveControls' -v ./internal/auditreceipt/` → exit 0, 최상위 12개 `--- PASS`, 하위 48개 PASS.
- `go test -count=1 -list 'TestAEC003a_|TestAEC003b_|TestE2' ./internal/hook/` → exit 0, 6개: `TestAEC003b_E2FenceShapes_RewriteUnderE2MakesTheReceiptStale`, `TestAEC003b_E2FenceShapes_ShapeInsideE2_RewriteBelowItMakesTheReceiptStale`, `TestAEC003a_EvidenceOnlyCommits_KeepTheReceipt`, `TestAEC003b_AnyOtherChange_MakesTheReceiptStale`, `TestAEC003a_UncommittedStateAlreadyJudged_KeepsTheReceipt`, `TestAEC003b_HeadUnchanged_IsJudgedAsBefore_Control`.
- `go test -count=1 -v -run 'TestAEC003a_|TestAEC003b_|TestE2' ./internal/hook/` → exit 0, `ok … 110.118s`, 최상위 6개 PASS, `--- PASS` 줄 57개.
- `GOOS=windows go build ./...` → exit 0, 출력 없음.
- `make agents-emit-check` → exit 0, `agents-emit-check: specs=12 copies=4 diff=0`.
- `make embed-manifest-check` → exit 0, `ok github.com/modu-ai/moai-adk/internal/template/embedemit 0.325s`.
- `golangci-lint run --timeout=5m ./internal/auditreceipt/... ./internal/hook/...` → exit 0, `0 issues.`.
- `git check-ignore -v .moai/reports/t2017/sync-audit-5.md` → `.gitignore:235:.moai/reports/*`.

### Baseline-attribution

- 측정 트리: 985853c683565e6f765ad17608d4526760804333(작업 트리 깨끗, 이번 실행).
- 비교 기준: 7ea9b205552e41f3381d266fee144fbc95a06f6d(4회차 감사 SHA).
- 두 사본은 `git archive -o … HEAD`와 `git archive -o … 7ea9b20555…`로 만들어 새 빈 스크래치 디렉터리(`…/scratchpad/a5.iV5yDm/{new,old}`)에 풀었다. 프로브와 뮤턴트는 사본에서만 실행했다. 추적 트리에는 아무것도 쓰지 않았다.
- 코퍼스: 워크트리의 `.moai/specs/*/progress.md`(HEAD 트리와 같음, 트리 깨끗)를 읽기 전용으로 읽었다.
- Go 시험, lint, 빌드는 이 트리의 소스를 로컬 `go` 툴체인으로 컴파일해 측정했다. 감사 계획은 설치본 `moai`로 읽었고, codex는 MCP 서버 빌드 `646955327e`(v3.2.0-rc.42)로 호출했다. 두 값은 계획과 codex 결과를 읽은 것일 뿐이고, 코드 판정 근거는 위의 Go 측정이다.

### Gaps

- G1. 계획은 `cross_model_active: true`를 보고했지만, 운영자 결정에 따라 `audit_multi`를 호출하지 않았다. convergence digest와 orchestrator의 `--result-file` 확인이 없다.
- G2. codex는 adversarial 모드라 범위가 기계적으로 묶이지 않았다(`prose-only scope`). codex는 Go 재현을 하지 않았다고 스스로 밝혔다. 재호출은 하지 않았다(카드당 2회 소진).
- G3. 거부된 명령이 1건 있다. 수리안 비용 프로브를 heredoc 복합 명령으로 만들려 했으나 워크트리 격리 가드가 "too complex to verify" 사유로 거부했다. Write 도구로 같은 파일을 사본에 만들고 단일 명령으로 다시 실행했으며, 그 결과를 썼다.
- G4. D19와 D18은 순수 함수 `e2DocsUnchanged` 수준과 코퍼스 수준에서 재현했다. 실제 git 범위와 Stop 훅을 거치는 end-to-end 재현은 하지 않았다(hook 형태 시험은 D16·D13·대조군만 다룬다).
- G5. 코퍼스 진단의 "fence 안 경계" 판정은 fence만 추적했다. HTML 주석이나 `<pre>` 안의 인용 경계는 세지 않았다.
- G6. 실행하지 않은 것: 9분짜리 `internal/cli` 넓은 선택(지시대로 미실행), 각 패키지의 전체 스위트(hook은 선택 6개만), race detector, Windows 시험 실행(빌드만), `/verify`.
- G7. D2–D9, D11, D12, D14, D15, D17은 다시 측정하지 않았다(3·4회차 기록을 인용).
- G8. sync-audit-1.md·sync-audit-2.md의 형태는 기존 `d10Shapes` 표(15개)로 확인했다. 그 밖에 이번에 추가로 만든 형태는 위 R1~R9와 P2 계열이다.

### Residual-risk

- 권고한 "열린 구성" 조임을 넣어도 짝수 개의 중첩 fence 줄 안에 있는 인용 경계(예: ````` ```` ````` 안에 ```` ``` ```` 한 쌍과 `## §E.3`)는 다시 숨을 수 있다. Markdown을 모델링하지 않는 비교의 한계다. 코드 주석과 SPEC §E에 잔여로 적어야 한다.
- R9(미종결 fence)와 비정규 헤딩 표기(`&sect;E.2`)는 잔여로 남긴다.
- 판정 파일의 `blocking_defect:` id를 감사자 자신이 쓴다는 점(§E B1)은 바뀌지 않았다.
- 스크래치의 프로브 출력 파일(`…/scratchpad/a5.iV5yDm/*.txt`)은 기계 로컬이다. 판정에 쓴 줄은 위 Evidence에 옮겨 적었고, 나머지는 소실을 감수한다.

## 반복 이력

| 회차 | 감사 커밋 | 판정 | 주요 변화 |
|---|---|---|---|
| 1 | 1008b915b2 | FAIL (D1 BLOCKING) | fence boolean 토글이 §E.2를 가렸다 |
| 2 | d9665c52c5 | FAIL 0.73 (D10 BLOCKING) | D1은 닫혔지만 빈 목록 비교가 남았다 |
| 3 | 400caf47ba | PASS-WITH-DEBT 0.85, receipts none | Claude 단독이라 영수증 가드가 거부했다(codex required) |
| 4 | 7ea9b20555 | FAIL 0.73 (D16, D13 BLOCKING), codex fail `rcpt-919c8ec5ced34525ba7663b0` | D16 신규, D13 재분류 |
| 5 | 985853c683 | FAIL 0.75 (D19 BLOCKING, repair-born), codex fail `rcpt-7911463ec3cd51004d46c029` | 원문 영역 비교로 D16·D13 닫힘. 인용 경계 fail-open(D19)이 회귀로 생김. 가용성 회귀 D18(73/851) |

## 다음 단계 (리더·운영자 판단 사항)

- D19는 위 Required fix(열린 구성 조임과 시험 추가)로 닫거나, SPEC §E 개정과 운영자 결정으로 잔여 위험으로 수용해야 한다. 앞의 방식을 쓰면 확인 감사 범위는 985853c683 이후 델타(blocking id: D19)가 된다.
- 이 카드의 codex 호출 2회가 모두 소진되었다. 확인 감사에 codex 영수증이 필요하면 3번째 호출이 되므로 운영자 승인이 필요하다(d-20261106T064700Z-de2d).
- D18은 optional이다. D19 수리와 함께 heading-like 판정 범위를 좁히면 실제 문서 56건의 비용을 줄일 수 있다.

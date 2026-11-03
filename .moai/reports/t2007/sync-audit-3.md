auditor-model: claude-opus-5-5

# t2007 sync 감사 판정 (3차, 확인 회차)

audited_sha: 16f6e3806f01a86b26204b13775483a5a14f7b0d
verdict: FAIL
score: 0.87

- 카드: t2007 (Class B, SPEC 없음), PR #1957, 브랜치 `WT-plan-plan-md`
- 감사 범위: 전체 범위 허용, 초점은 2차 FAIL(0.85 @85ada173bf, 영수증 `rcpt-b4696a80b4db728d92aa45e7`)의 D1~D3 해소 여부. 대상 커밋은 `16f6e3806f` 한 개(`git diff 85ada173bf 16f6e3806f`, 5개 파일, +57/-10)
- 보류 항목 PLN-02, PLN-03, PLN-04, PLN-07, PLN-09, PLN-11은 3.2 설계 입력이므로 이 카드의 결함으로 세지 않았습니다.
- 감사 계획: `moai verify audit-plan` → `cross_model_active: true`, codex `required`, claude·glm `off`
- 교차 모델: `audit_multi` `overall_verdict: fail` (필수 백엔드 codex fail), 영수증 `rcpt-0ee8ad73e6dc28d018e2a8c0`
- 수렴 확인: `moai verify audit-plan --result-file` → `convergence_check: {"ok": true, "unmet": []}` (게이트는 충족, 즉 codex가 실제로 판정을 냈고 그 판정이 fail)

## 판정 요약

2차 결함 세 건은 고쳐졌습니다. D1(Step 2.3.5의 고정 3회)은 두 사본 모두 tier별 한도 기준으로 바뀌었고, D3(큰따옴표 `status: "ok"`)은 정규식으로 잡히며, D2는 양성 검사가 추가되어 2차의 M15(GEARS 금지 문구)와 "5 iterations exhausted" 변이를 이제 잡습니다. 미러는 바이트 단위로 같고 카탈로그 해시도 맞습니다.

하지만 확인 회차를 통과하지 못했습니다. 이유는 두 가지입니다.

1. 이번 수정이 새 모순을 만들었습니다(D6). 새 Step 2.3.5는 "Tier S(1회 후)에도 똑같이 적용, Phase 12로 자동 진행하지 말 것"이라고 하는데, 같은 파일 160행은 `minimal` 하네스에서 "FAIL은 기록만 하고 Phase 12를 막지 않는다"고 합니다. 이전 문구("세 번 모두 FAIL이면")는 1회짜리 `minimal`에서 발동할 수 없어 충돌이 없었는데, 이제는 `minimal`의 첫 FAIL이 에스컬레이션 조건을 그대로 만족합니다. codex도 독립적으로 같은 결함을 P2로 지적했습니다.
2. 필수 게이트 codex가 fail을 냈습니다(D6 + 양성 검사 우회). 이 감사도 2차 M14 변이("5 iterations are used up")와 본문 고정값 변이("after 4 FAIL iterations")가 여전히 통과함을 재현했습니다.

따라서 FAIL입니다.

## 차원별 점수

| 차원 | 점수 | 판정 | 근거 |
|---|---|---|---|
| Functionality (40%) | 82/100 | FAIL | D1·D3 해소(E2 MA·MC·MD), D2 부분 해소(E2 ME). 새 blocking D6: Step 2.3.5가 `minimal` 비차단 계약(160행)과 충돌 |
| Security (25%) | 95/100 | PASS | 수정 diff 비밀값·주입 탐침 `grep_exit=1`. 운영 Go 코드 변경 없음 |
| Craft (20%) | 82/100 | PASS | `go vet` 출력 없음, `gofmt_exit=0`. 양성 검사는 제목·다른 절의 문구로 충족되어 본문 변이(MB, MF)를 놓침(D2r, 선택) |
| Consistency (15%) | 92/100 | PASS | 지정된 3개 파일 `cmp` exit 0, `TestCatalogHashCoversSkillSubfiles` PASS(재생성된 `moai` 카탈로그 해시가 트리 해시와 일치). Step 2.3.5 문구는 2.3.4와 일치(같은 "tier-resolved ceiling", `remaining_attempts`, `{ceiling}`) |

조화 평균 4/(1/82+1/95+1/82+1/92) ≈ 87.4 → 0.87. 필수 통과 차원인 Functionality가 blocking D6으로 FAIL이고 필수 게이트 codex가 fail이므로 점수와 무관하게 전체 FAIL입니다.

## 2차 결함 처리 현황

| 결함 | 2차 분류 | 3차 결과 | 근거 |
|---|---|---|---|
| D1 Step 2.3.5 고정 3회 | blocking | **해소** | 두 사본 263·265·268행 변경, 금지어 3개 추가, 변이 MD FAIL |
| D2 금지어 목록만 검사 | optional | **부분 해소** (D2r로 이어짐) | ME·MA는 잡음, MB·MF는 통과 |
| D3 큰따옴표 `status: "ok"` | optional | **해소** | `okStatus` 정규식, 변이 MC FAIL |
| D4 schema 90행 미러 차이 | optional, 기존 결함 | 미처리(범위 밖) | 이번 커밋에 없음 |
| D5 plan-auditor.md 고정 3회 | optional, 범위 밖 | 미처리(범위 밖) | `plan-auditor.md:649,702` 두 사본에 그대로 |

## 결함 목록 (structured defect-list)

- **D6** [P2, 신뢰도 높음] [blocking] `.claude/skills/moai/workflows/plan/spec-assembly.md:265` (템플릿 미러 같은 줄) — 새 Step 2.3.5가 하네스 단계를 구분하지 않고 "Tier S (after the single attempt) … Do NOT proceed to Phase 12 automatically"라고 정합니다. 같은 파일 160행은 `minimal`을 "a FAIL verdict is logged but does not block Phase 12"로, 281행과 `harness.yaml` 96행은 `require_must_pass: false`("warn but do not block")로 정합니다. 기준 커밋의 "If all three iterations result in FAIL"은 1회짜리 `minimal`에서 발동할 수 없었으므로 이 충돌은 `16f6e3806f`가 새로 만든 것입니다. codex가 같은 결함을 독립적으로 지적했습니다. 필수 수정: 두 사본 모두 Step 2.3.5 첫 문장에 적용 범위를 명시하십시오(예: "Under the blocking levels (`standard`/`thorough`), escalation fires when …; under `minimal` a final FAIL is logged and Phase 12 proceeds per the harness-level intensity above."). `TestPlanStaleContractPhrases`의 escalation 절 검사에 `standard`/`thorough` 또는 `minimal` 예외 문구가 있는지 양성 검사를 넣고 카탈로그 해시를 다시 생성하십시오.
- **D2r** [P2, 신뢰도 높음] [optional — 다만 필수 게이트 codex가 같은 항목으로 fail을 냈으므로 다음 감사에서도 지적될 가능성이 큼] `internal/template/plan_contract_wording_test.go:222-238` — 양성 검사 `strings.Contains(retry, "tier-resolved ceiling")`은 Step 2.3.4 제목("Retry Loop (tier-resolved ceiling)")과 추적 줄만으로 충족되고, `remaining_attempts`는 파일 전체에서 찾으며, 고정값 정규식은 `[0-9]+ iterations exhausted`와 `iteration \{N\}/[0-9]` 두 형태뿐입니다. 그래서 4번 항목을 "Repeat until PASS or 5 iterations are used up."로 바꾼 변이(MB)와 Step 2.3.5 본문을 "Escalation fires after 4 FAIL iterations."로 바꾼 변이(MF)가 모두 통과합니다. codex는 두 절의 본문을 전부 지우고 제목만 남겨도 통과함을 보고했습니다. 권고: 제목 줄을 제외한 각 절 본문에서 "tier-resolved ceiling"과 `remaining_attempts`를 각각 요구하고, 고정 예산 정규식을 `\b[0-9]+ (FAIL )?iterations?\b`처럼 넓히십시오. 본문 삭제 변이가 반드시 실패해야 합니다.
- **D4** [낮음] [optional, 기존 결함·범위 밖] `.claude/rules/moai/development/spec-frontmatter-schema.md:90` — 로컬에만 `` (`internal/spec/audit.go`) ``가 있는 미러 차이(2차 E3). 이번 커밋에서 바뀌지 않았습니다.
- **D5** [낮음] [optional, 범위 밖] `.claude/agents/moai/plan-auditor.md:649,702` (템플릿 미러 동일) — "Iteration: {N}/3", "all three iterations"가 남아 있습니다(E5 grep). 이 카드가 바꾼 파일이 아니며 PLN-08 설계 정리 때 함께 다루기를 권합니다.
- **N1** [낮음, 신뢰도 중간] [optional, 기존 문구] `spec-assembly.md:285` — `thorough`의 교차 검증 재심이 PASS가 아니면 "trigger one additional iteration"이라고 하는데, 마지막 회차(`remaining_attempts` 0)에서 추가 회차가 한도를 넘는지는 정하지 않습니다. 163행이 "consumed by every reviewer invocation, including cross-validation"이라고 하므로 한도 안에서 해석할 수 있으나 명시는 없습니다. 이번 커밋이 바꾼 줄이 아닙니다.

## Claim

1. D1 해소: Step 2.3.5가 고정 3회 대신 tier별 한도로 발동하고, 보고서 목록이 `-review-{ceiling}.md`까지다 — **확인** (diff, MD FAIL)
2. D2 해소: 양성 검사가 다른 형태의 잘못된 문구를 잡는다 — **부분 확인** (ME·MA FAIL, MB·MF 통과)
3. D3 해소: 큰따옴표 조기 `status: "ok"` 반환을 잡는다 — **확인** (MC FAIL)
4. 미러 일치: spec-assembly.md, advanced-patterns.md, reference.md 로컬과 템플릿 사본이 바이트 단위로 같다 — **확인** (`cmp` exit 0 ×3)
5. 카탈로그 해시: 재생성된 `moai` 스킬 해시가 배포 트리 해시와 같다 — **확인** (`TestCatalogHashCoversSkillSubfiles` PASS)
6. 이번 커밋이 새 결함을 만들지 않았다 — **거짓** (D6)
7. 감사 중 트리가 변하지 않았다 — **확인** (변이 복원마다 `git status --short` 출력 없음, 마지막 HEAD `16f6e3806f…`)

## Evidence

**E0 — 감사 대상 확인**
```
$ git rev-parse HEAD
16f6e3806f01a86b26204b13775483a5a14f7b0d
$ git rev-parse --show-toplevel
/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t2007
$ git status --short
(출력 없음)
$ git diff --stat 85ada173bf 16f6e3806f
 .../skills/moai/workflows/plan/spec-assembly.md    |  6 +--
 internal/template/catalog.yaml                     |  2 +-
 internal/template/plan_contract_wording_test.go    | 45 +++++++++++++++++++++-
 .../template/plan_research_fanout_guard_test.go    |  8 +++-
 .../skills/moai/workflows/plan/spec-assembly.md    |  6 +--
 5 files changed, 57 insertions(+), 10 deletions(-)
```

**E1 — 범위 시험 (대조군, 변이 없음)**
```
$ go test ./internal/template/ -run 'TestPlan|TestCatalog|TestManifestHash' -count=1 -v
--- PASS: TestPlanContractWording (0.00s)
--- PASS: TestPlanSpecAssemblyIssueAndArtifactWording (0.00s)
--- PASS: TestPlanStaleContractPhrases (0.00s)
--- PASS: TestPlanResearchFanoutGuard (0.00s)
--- PASS: TestPlanSignalDocsNameAuditReady (0.00s)
--- PASS: TestCatalogReservedFieldType (0.00s)
--- PASS: TestCatalogNoDuplicateEntries (0.01s)
--- PASS: TestPlanAuditorHasNoStaticSkillPreload (0.00s)
--- PASS: TestCatalogManifestPresent (0.01s)
--- PASS: TestCatalogTierValid (0.01s)
--- PASS: TestCatalogReferencesValid (0.01s)
    catalog_tier_audit_test.go:541: audited 59 directory entries for whole-tree hash coverage
--- PASS: TestCatalogHashCoversSkillSubfiles (0.04s)
--- PASS: TestManifestHashFormat (0.04s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/template	0.549s
```
레인 증거(이 감사가 실행하지 않음): 같은 SHA에서 `internal/template` 전체 패키지 `-timeout 40m` → `ok ... 612.861s` (리더 지시로 인용).

**E2 — 변이 시험 (템플릿 사본 직접 변이, 매 변이 후 `git checkout -- <변이한 파일>` 복원)**

| 변이 | 내용 | 대상 시험 | 결과 |
|---|---|---|---|
| MA | Step 2.3.4 4번을 "Repeat until PASS or 5 iterations exhausted." | Stale·IssueAndArtifact·ContractWording | **FAIL** `plan_contract_wording_test.go:237: spec-assembly.md Steps 2.3.4/2.3.5 carry a fixed budget "5 iterations exhausted"` |
| MB | 4번을 "Repeat until PASS or 5 iterations are used up." (2차 M14) | 같음 | `ok ... 0.503s` — 미검출 (D2r) |
| MC | fanout 검증 블록 앞에 `if (research_md !== undefined) return { status: "ok", ... }` | `TestPlanResearchFanoutGuard` | **FAIL** `plan_research_fanout_guard_test.go:64: status: 'ok' is returned before the synthesis validation` |
| MD | Step 2.3.5 제목을 "Escalation after repeated FAIL", 본문을 "If every iteration results in FAIL, …"로 (tier 문구 제거) | `TestPlanStaleContractPhrases` | **FAIL** `plan_contract_wording_test.go:226: spec-assembly.md Step 2.3.5 does not name the tier-resolved ceiling` |
| ME | advanced-patterns 59행을 "Reject any requirement that chains more than one of Where/While/When" (2차 M15) | Stale·IssueAndArtifact·ContractWording | **FAIL** `plan_contract_wording_test.go:247: … does not carry the GEARS compound wording "GEARS compound clause"` |
| MF | Step 2.3.5 제목 유지, 본문을 "Escalation fires after 4 FAIL iterations. Do NOT proceed …" | 같음 | `ok ... 0.559s` — 미검출 (D2r) |

복원 확인: 각 복원 뒤 `git status --short` 출력 없음, 마지막 `git rev-parse HEAD` → `16f6e3806f01a86b26204b13775483a5a14f7b0d`.

**E3 — 미러 바이트 일치**
```
$ cmp .claude/skills/moai/workflows/plan/spec-assembly.md internal/template/templates/.claude/skills/moai/workflows/plan/spec-assembly.md
spec-assembly_exit=0
$ cmp (advanced-patterns.md 두 사본)
advanced-patterns_exit=0
$ cmp (reference.md 두 사본)
reference_exit=0
```

**E4 — D6 근거 (템플릿 사본, 로컬 사본은 E3로 동일)**
```
spec-assembly.md:160: - `minimal`: lightweight, non-blocking 1-iteration audit (`max_iterations: 1`, `require_must_pass: false`) — a FAIL verdict is logged but does not block Phase 12
spec-assembly.md:161: - `standard`/`thorough`: blocking retry loop using the tier-resolved ceiling
spec-assembly.md:265: The tier-resolved ceiling (S 1 / M 2 / L 3) owns the number of review attempts. Escalation fires when the final permitted iteration for the SPEC's tier ends in FAIL, … This applies identically to Tier S (after the single attempt), Tier M (after the second attempt), and Tier L (after the third attempt). Do NOT proceed to Phase 12 automatically.
spec-assembly.md:281: - `minimal`: plan_audit.enabled: true, max_iterations: 1, require_must_pass: false (lightweight, non-blocking 1-iteration audit …)
harness.yaml:95-96: max_iterations: 1 / require_must_pass: false      # Relaxed for minimal: warn but do not block
기준(59464c5f37) Step 2.3.5: "If all three iterations result in FAIL, do NOT proceed to Phase 12 automatically."
```

**E5 — 남은 고정 3회 문구**
```
$ git grep -n -E 'after 3 FAIL|all three iterations|review-3\.md|iterations exhausted|\{N\}/3' -- .claude internal/template/templates/.claude
.claude/agents/moai/plan-auditor.md:649:Iteration: {N}/3
.claude/agents/moai/plan-auditor.md:702:Stagnation detection: If a defect appears in all three iterations unchanged, …
.claude/rules/moai/core/agent-common-protocol-reference.md:303: … auto-fix iterations exhausted without green CI   (무관: /moai loop)
(템플릿 사본 같은 내용)
```
spec-assembly.md에는 해당 문구가 없습니다.

**E6 — Security / Craft 기계 검사**
```
$ git diff 85ada173bf HEAD -U0 --output=<scratch>/fix.diff ; grep -n -i -E '^\+.*(api[_-]?key|secret|password|token=|BEGIN [A-Z]+ PRIVATE|eval\(|child_process|exec\()' <scratch>/fix.diff
grep_exit=1   (101줄 diff, 일치 없음)
$ go vet ./internal/template/
(출력 없음)
$ gofmt -l internal/template/plan_contract_wording_test.go internal/template/plan_research_fanout_guard_test.go
gofmt_exit=0 (출력 없음)
```

**E7 — 감사 계획, 교차 모델, 수렴 확인**
```
$ moai verify audit-plan --project-root /Users/goos/MoAI/moai-adk-go/.moai/worktrees/t2007
"config_status": "ok", "model": "codex", "cross_model_active": true, "cross_model_required": true,
"enforced_required": ["codex"]; backends: claude off / codex required / glm off; exit=0

mcp__moai__audit_multi(project_root=<위 경로>, target=baseBranch, card_id=t2007,
  focus="plan contract doc wording and lexical guard tests", gates 인자 없음, claude_verdict 없음)
overall_verdict: fail
gate_unmet: (필드 없음)
plan_source: config
per_backend_verdicts: [{backend: codex, gate: required, verdict: fail}]
participant_count: 1
disagreement_flag: null
residual_risk_note: required-backend FAIL: codex
fail_open_backends: []
audit_receipt: rcpt-0ee8ad73e6dc28d018e2a8c0
build_commit: b1806498fe
tree_root: /Users/goos/MoAI/moai-adk-go/.moai/worktrees/t2007

codex 지적 2건(P2):
 (1) spec-assembly.md:265 — 새 에스컬레이션 규칙이 minimal의 비차단 계약과 충돌 (= D6)
 (2) plan_contract_wording_test.go:222 — Step 2.3.4·2.3.5 본문을 지우고 제목만 남겨도 TestPlanStaleContractPhrases 통과 (= D2r)

$ (Write) .moai/state/audit-plan-result.json
{"overall_verdict":"fail","gate_unmet":[],"plan_source":"config","per_backend_verdicts":[{"backend":"codex","gate":"required","verdict":"fail"}]}
$ moai verify audit-plan --project-root /Users/goos/MoAI/moai-adk-go/.moai/worktrees/t2007 --result-file /Users/goos/MoAI/moai-adk-go/.moai/worktrees/t2007/.moai/state/audit-plan-result.json
... "enforced_required": ["codex"],
  "convergence_check": {
    "ok": true,
    "unmet": []
  }
exit=0
```
`ok: true`는 필수 게이트가 실제 판정을 받았다는 뜻이며(미충족 게이트 없음), 그 판정 자체가 fail입니다.

## Baseline-attribution

- 모든 Go 시험, `go vet`, `gofmt`, `cmp`, `git grep`, 변이 시험은 이번 실행에서 워크트리 `/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t2007`, HEAD `16f6e3806f01a86b26204b13775483a5a14f7b0d`에 대해 수행했습니다.
- 변이 시험은 go:embed 대상인 템플릿 사본을 직접 바꿨다가 `git checkout -- <변이한 파일>`로 되돌렸습니다. 각 변이 전 대조군(E1)이 먼저 통과했습니다.
- 도구 출처(§2.2): `moai verify audit-plan`(계획 조회와 `--result-file` 확인)은 설치된 CLI `/Users/goos/go/bin/moai` v3.2.0-rc.39, 빌드 커밋 `dad3603de1`이 판정했습니다. `audit_multi`는 MCP 서버 빌드 `b1806498fe`(v3.2.0-rc.38)가 판정했습니다. `git merge-base --is-ancestor dad3603de1 HEAD` → exit 1, `git merge-base --is-ancestor b1806498fe HEAD` → exit 1: 두 빌드 모두 감사 대상 HEAD의 조상이 아닙니다. 이 트리에서 빌드한 바이너리는 아니며, 두 도구는 감사 계획·교차 모델 수렴만 판정했고 카드 내용은 이 트리의 `go test`가 판정했습니다.

## Gaps

- 지시에 따라 `./internal/template` 전체 패키지, golangci-lint, 커버리지 측정은 실행하지 않았습니다. 전체 패키지 결과(`ok ... 612.861s`)는 레인 증거이며 이 감사가 관측하지 않았습니다. 운영 Go 코드 변경이 없어 커버리지 수치는 의미가 없다고 판단했습니다.
- 워크트리 격리 가드가 복합 명령 네 개를 거부했습니다(`cd … && go test … > 파일`, `for` 루프 `cmp`, `git diff | grep`, `git merge-base … ; …`). 같은 확인을 단일 명령으로 나눠 실행했고 측정 범위는 같습니다. `git diff | grep`은 diff를 scratch 파일로 쓴 뒤 grep했습니다.
- fanout 스크립트의 node 실행은 하지 않았습니다. `16f6e3806f`는 스크립트를 바꾸지 않았으므로(E0 stat) 판단은 Go 가드 변이(MC)에 한정됩니다.
- codex가 보고한 "본문 전부 삭제" 변이는 codex 측 실행이며, 이 감사는 MB·MF로 같은 종류를 재현했습니다.
- 하네스 단계별 실제 오케스트레이터 동작(문서를 읽은 에이전트가 `minimal`에서 실제로 멈추는지)은 관측하지 않았습니다. D6은 문서 계약의 모순으로 판정했습니다.
- 1차 선택 사항 F3·F5·F6은 다시 확인하지 않았습니다.

## Residual-risk

- D6이 남아 있는 동안 `minimal` 하네스의 SPEC은 첫 FAIL에서 Phase 12를 막는 문서와 막지 않는 문서를 동시에 받습니다. 어느 쪽을 따르든 다른 쪽 계약을 어기게 됩니다.
- D2r이 남아 있는 동안 재시도·에스컬레이션 본문이 다른 고정 숫자로 바뀌어도 CI가 잡지 못합니다. 모든 문구 가드는 어휘 검사라 에이전트의 실제 행동을 증명하지 않습니다.
- `harness.yaml` `plan_audit_ceiling_policy.on_final_hit: hold-and-split`과 Step 2.3.5의 AskUserQuestion 3지선다의 관계는 여전히 문서상 정리되지 않았습니다(PLN-08 설계 입력 범위, 2차와 동일).
- residual_risk_note(verbatim): `required-backend FAIL: codex`
- 변이 시험의 원시 출력은 대화 기록과 scratch에만 있으며 이 파일에 옮긴 줄만 근거로 삼습니다.

## 교차 모델 digest

- overall_verdict: fail
- gate_unmet: (필드 없음, digest에는 빈 배열로 기록)
- plan_source: config
- per_backend_verdicts: codex / gate required / verdict fail
- disagreement_flag: null (participant_count 1)
- convergence_check: ok true, unmet [] (위 E7, 이 감사가 직접 실행)

## 반복 이력

- 1차 (`sync-audit-1.md`): FAIL 0.83 @bfb9e6f914 — F1, F2 blocking
- 2차 (`sync-audit-2.md`): FAIL 0.85 @85ada173bf — F1·F2 해소, 새 blocking D1, codex 필수 게이트 fail(D1·D2), 영수증 `rcpt-b4696a80b4db728d92aa45e7`
- 3차 (이 문서, 사전 승인된 단일 확인 회차): FAIL 0.87 @16f6e3806f — D1·D3 해소, D2 부분 해소, 새 blocking D6, codex 필수 게이트 fail(D6·D2r), 영수증 `rcpt-0ee8ad73e6dc28d018e2a8c0`

## 권고

- D6: 두 사본의 Step 2.3.5 첫 문장에 `standard`/`thorough` 적용 범위와 `minimal` 예외를 명시하는 한 문장 수정이면 충분합니다.
- D2r: 양성 검사를 제목 제외 본문 단위로 좁히고 고정 예산 정규식을 넓히면 codex 지적도 함께 해소됩니다. 선택 사항이지만 필수 게이트 통과를 위해 함께 넣는 편이 안전합니다.
- 사전 승인된 확인 회차가 소진되었으므로 다음 감사 회차는 리더·운영자 판단이 필요합니다.
- D4·D5·N1은 선택 사항입니다.

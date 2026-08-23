# t205 — llm.yaml effort 매트릭스 3프로파일 갱신 + manager-lead 행 추가

- 카드: t205 (TARGET VALUES 표가 SSOT)
- 브랜치: `WT-effort-matrix` (base `origin/main` = `28bde4022`)

## 1. Claim

`high` / `medium` / `low` 세 프로파일 전부가 카드의 TARGET VALUES와 정확히 일치하게 해소된다. `manager-lead` 는 세 프로파일 모두에 매핑된 행으로 존재한다. 값은 Go SSOT · 배포 `llm.yaml` · 커플링된 `harness_agents` 셀 · 에이전트 frontmatter `effort:` 네 표면에서 일치한다.

## 2. Evidence

### 2.1 카드 범위가 실제 불변식보다 좁았다 (선행 측정)

카드는 범위를 "`llm.yaml` (template + 로컬) + `make build`"로 적었다. 측정 결과 그 범위만으로는 **CI가 붉어지고, 변경의 절반은 아예 효력이 없다.**

**(a) `llm.yaml` 은 SSOT가 아니다.** `internal/template/profile_matrix.go` 의 `defaultProfileMatrix` 가 SSOT이고(`// DefaultProfileMatrix returns a deep copy of the per-agent Go-code SSOT`), `llm.yaml` 은 그 미러다. 둘의 불일치는 기계적으로 잡힌다:

```
$ go test ./internal/template/ -run EmbeddedLLM
--- FAIL: TestEmbeddedLLMYAMLMatchesMatrix
    high/manager-spec: embedded llm.yaml has {opus high}, DefaultProfileMatrix() has {opus medium}
      — the shipped config would shadow the Go matrix
    high/manager-docs: embedded llm.yaml has {opus low}, DefaultProfileMatrix() has {sonnet low}
    medium/manager-spec: ... / medium/manager-docs: ...
```

**(b) effort 는 frontmatter 로만 전달된다.** 스폰 시 주입되는 것은 `model` 뿐이고, `effort` 에는 Agent 도구 파라미터가 없어 에이전트 파일의 `effort:` 가 유일한 채널이다(`agent-common-protocol.md` § Per-Spawn Model Injection). 즉 frontmatter 를 그대로 두면 **카드가 요구한 effort 변경은 어디에도 반영되지 않는다.** 게다가 LR-12 린트가 medium 컬럼에서 파생되므로(`buildCanonicalEffortMatrix`) 드리프트로 잡힌다.

**(c) `manager-lead` 는 매핑 자체가 없었다.** `profileMatrixAgentOrder` 에 부재 → `ResolveAgentModelEffort` 4단계 fallback 의 `inherit` 센티널로 떨어진다. Tier L 조율자, 즉 다른 모든 에이전트로 팬아웃하는 유일한 행이 세션이 어쩌다 올라탄 모델을 그대로 쓰고 있었다.

**(d) `harness_agents` 셀은 파일 자체가 등식을 선언한다.** "Keep these cells equal to the row they name. The resolver reads a cell here FIRST and only falls through to the row when the cell is absent, so a stale cell silently splits the two paths." 프로파일 행만 바꾸면 그 등식이 깨진다.

그래서 범위를 넓혔다. 넓힌 부분은 전부 **선택이 아니라 귀결**이다.

### 2.2 변경 표면

| 파일 | 변경 |
|---|---|
| `internal/template/profile_matrix.go` | `defaultProfileMatrix` 3컬럼 전면 갱신 + `manager-lead` 행 3개 · `profileMatrixAgentOrder` 에 `manager-lead` · `GroupLead` 신설 + `agentGroupMembership` 등록 · 셀 도출 근거 주석 재작성(구 phase-weighted 서술 폐기) |
| `internal/template/templates/.moai/config/sections/llm.yaml` | 3컬럼 미러 + `manager-lead` · `harness_agents` 커플링 셀 갱신 · Column intent / `max` / profile-invariant 서술 정정 |
| `.claude/agents/moai/{manager-spec,manager-develop,manager-design,manager-lead}.md` + 템플릿 미러 | frontmatter `effort:` 를 medium 컬럼에 맞춤 (spec high→medium, develop high→medium, design medium→high, lead xhigh→high) |
| `internal/template/templates/.codex/agents/moai/*.toml` | `make agents-emit` 재생성 (4개) |
| `internal/config/testdata/shipped_key_inventory.yaml` | `llm.profiles.{high,medium,low}.manager-lead.{model,effort}` 6키 등록 (W/reader) |
| `internal/web/assets/i18n.js` | `agentdesc.manager-lead` ko/ja/zh 3로케일 |
| `internal/cli/profile_setup_translations.go` | 프로파일 선택 라벨 3종 × 4로케일 = 12개 문자열 (아래 §2.6) |
| `internal/cli/wizard/{questions,translations}.go` | init 위저드 티어 설명 3종 × 4로케일 (같은 산문, 테스트 미커버) |
| 테스트 7개 파일 | 아래 §2.4 · §2.6 |
| `internal/template/catalog.yaml` | `make build` 재생성 |

`GroupLead` 를 새로 만든 이유: 기존 7개 그룹 중 `manager-lead` 의 형태(high/high/medium)와 맞는 것이 없다. 그룹은 해소에 쓰이지 않는 표시·오버라이드 게이트지만, 없으면 `AgentGroup()` 이 실패해 웹 설정에서 이 에이전트만 오버라이드가 조용히 무시된다.

### 2.3 3프로파일 실측 (`moai model profile --json`)

카드가 요구한 검증. medium 은 워크트리에서, high/low 는 `profile:` 만 바꾼 임시 프로젝트에서 실행.

```
medium  manager-spec opus/medium · plan-auditor opus/high · sync-auditor opus/high
        manager-develop opus/medium · super-advisor opus/high · manager-design opus/high
        manager-lead opus/high (group=lead) · builder-harness opus/medium · e2e-tester opus/low
        manager-docs sonnet/low · manager-git sonnet/low · Explore sonnet/low

high    manager-spec opus/medium · plan-auditor opus/high · sync-auditor opus/high
        manager-develop opus/medium · super-advisor opus/high · manager-design opus/high
        manager-lead opus/high · builder-harness opus/high · e2e-tester opus/medium
        manager-docs sonnet/low · manager-git sonnet/low · Explore sonnet/low

low     manager-spec opus/medium · plan-auditor opus/medium · sync-auditor opus/medium
        manager-develop opus/medium · super-advisor opus/high · manager-design opus/medium
        manager-lead opus/medium · builder-harness opus/low · e2e-tester sonnet/low
        manager-docs sonnet/low · manager-git sonnet/low · Explore sonnet/low
```

세 컬럼 모두 카드의 TARGET VALUES와 셀 단위로 일치한다. Explore 는 지시대로 불변.

### 2.4 갱신한 테스트와, 그중 **관측력을 잃을 뻔한 것 3개**

값을 옮기면 통과하는 테스트는 값만 고쳤다. 그러나 셋은 새 값에서 **아무것도 관측하지 못하게** 되어 프로브 자체를 바꿔야 했다 — 값만 맞췄으면 초록이면서 무의미해졌을 것들이다:

- `TestResolveAgentModelEffort_LegacyAlias`: `perf_tier: max` → high 컬럼 해소를 확인하는데, 프로브였던 `manager-develop` 이 이제 세 컬럼 모두 `medium` 이다. 어느 컬럼으로 가도 통과한다. 세 컬럼이 아직 다른 `builder-harness` 로 교체.
- `TestG3ReadPathDerivesFromProfileMatrix`: "frontmatter 가 아니라 매트릭스에서 읽는다"를 확인하는데, 심어둔 divergent frontmatter(`medium`)와 `manager-spec` 의 새 high 셀(`medium`)이 같아졌다. `manager-design`(high 셀=`high`)으로 교체.
- `TestResolveAgentModelEffort_StaleGroupKeyedMirror`: 스테일 group-key 셀이 무시되는지 보는데, 심어둔 값 `sonnet/low` 가 `manager-docs` 의 새 셀과 같아졌다. 심는 값을 `opus/max` 로 교체.

나머지: `MatrixAFidelity`·`LowColumn`·`Shape`(33→36셀)·`HarnessAgentModelEffort`(implement max→medium) / `agent_lint_test.go` LR-12 2건 / `agentfm_polish_test.go` 정렬(`manager-docs` 가 sonnet 이 되어 within-opus 비교 대상이 사라짐 → `manager-design` 추가) / `g3` save-default / `golden_test.go` `expectedEffort`.

`TestDefaultProfileMatrix_Monotone`(각 행 high ≥ medium ≥ low)은 수정 없이 통과 — 새 값이 12행 모두 단조다.

### 2.6 CI가 잡은 두 표면 — 내 로컬 검증 범위가 좁았다

첫 push 후 CI `Test (ubuntu-latest)` 가 실패했다. 원인은 코드가 아니라 **내가 고른 검증 범위**다: `./internal/cli/agentlint/...` 는 돌렸지만 `./internal/cli/` 자체를 돌리지 않아, 그 패키지의 두 소비자를 놓쳤다.

```
--- FAIL: TestResolveModelProfileReport_MaxClaude
    model_test.go:30: manager-develop high: got opus/medium, want opus/max
    model_test.go:36: expected 11 agents, got 12
--- FAIL: TestModelPolicyLabels_AgreeWithProfileMatrix  (high/medium/low × en·ko·ja·zh)
    label "High - Opus 5 (max~low) ..." should state the matrix opus effort span (high~medium)
    label ... mentions docs=false but manager-docs on sonnet=true
```

두 번째가 특히 중요하다 — `TestModelPolicyLabels_AgreeWithProfileMatrix` 는 프로파일 선택 UI의 **산문 라벨을 매트릭스에서 파생해 검증한다**. opus effort 폭(`(hi~lo)`), sonnet effort, 그리고 `docs` / `e2e` 가 그 컬럼에서 실제로 sonnet 행인지까지 문자열 포함으로 대조한다. 즉 셀을 바꾸면 사용자가 읽는 문장도 함께 바꿔야 하고, 그 등식은 기계로 강제돼 있다.

새 라벨(4로케일 동일한 기술 문자열):

| 컬럼 | 이전 | 이후 |
|---|---|---|
| high | `Opus 5 (max~low) + Sonnet (low, single-shot rows only)` | `Opus 5 (high~medium) + Sonnet (low, docs/single-shot rows)` |
| medium | `Opus 5 (high~low) + Sonnet (low, single-shot rows only)` | `Opus 5 (high~low) + Sonnet (low, docs/single-shot rows)` |
| low | `Opus 5 (medium~low) + Sonnet (low, docs/e2e/single-shot rows)` | `Opus 5 (high~low) + Sonnet (low, docs/e2e/single-shot rows)` |

`internal/cli/wizard` 의 init 티어 설명도 같은 산문을 들고 있으나 이 테스트가 커버하지 않는다 — 스테일로 남으면 위저드만 `max~low` 라고 말하게 되므로 4로케일 모두 함께 갱신했다.

**교훈(기록용):** "affected packages" 를 서브패키지 단위로 좁게 고른 것이 실패의 직접 원인이다. 매트릭스처럼 fan-in 이 넓은 SSOT 를 건드릴 때는 소비자를 grep 으로 먼저 열거해야 한다.

### 2.5 명령 출력

```
$ go build ./...                                          (무출력)
$ gofmt -l <touched 6 go files>                           (무출력)
$ golangci-lint run ./internal/template/... ./internal/web/... ./internal/cli/agentlint/... ./internal/config/...
0 issues.
$ go test ./internal/template/... ./internal/config/... ./internal/cli/agentlint/... ./internal/web/... ./internal/settings/... -count=1
ok  internal/template            32.2s
ok  internal/template/agentemit   0.8s
ok  internal/config               6.7s
ok  internal/config/atomicfile    1.5s
ok  internal/config/toolpolicy    2.6s
ok  internal/cli/agentlint        2.0s
ok  internal/web                  7.7s
ok  internal/settings             3.2s
ok  internal/settings/agentfm     3.5s
ok  internal/settings/yamlpatch   3.2s

# §2.6 수정 후 추가 실행
$ go test ./internal/cli/ -count=1 -timeout 900s
ok  internal/cli                397.4s
$ go test ./internal/cli/wizard/ ./internal/cli/agentlint/ -count=1
ok  internal/cli/wizard           3.2s
ok  internal/cli/agentlint        0.9s
$ golangci-lint run ./internal/cli/...
0 issues.
```

## 3. Baseline-attribution

- 기준 트리: `WT-effort-matrix`, base `origin/main` = `28bde4022`
- 모든 수치는 이 트리에서 이 라운드에 실행한 명령의 출력. `-count=1` 로 캐시 비활성.
- `moai model profile` 출력은 `make build` 로 재빌드한 `./bin/moai` 로 측정 (설치본 `~/go/bin/moai` 아님).

## 4. Gaps (미검증)

- **로컬 `.moai/config/sections/llm.yaml` 은 이 PR에 없다.** `.gitignore:192` 로 추적되지 않아 커밋될 수 없고, 실물은 primary 체크아웃에만 있다. 이 파일의 `profiles` 블록은 Go 매트릭스를 **가리므로**(위 §2.1(a)의 실패 메시지가 그 하자를 그대로 서술한다), 갱신하지 않으면 운영자 머신에서는 새 값이 적용되지 않는다. 레인이 primary 체크아웃의 공유 경로를 직접 쓰는 것은 별개 위험이라 손대지 않았다 — **리드 처리 필요**(§5 참조).
- 전체 스위트(`go test ./...`)는 로컬 미실행(CLAUDE.local.md §4). 전 패키지 판정은 CI 몫.
- Windows/Linux 매트릭스 미검증 — CI 몫.
- 새 effort 값이 **실제 에이전트 품질/비용에 미치는 영향**은 측정하지 않았다. 값은 운영자 지정 입력이며, 이 카드는 그 값이 네 표면에 일관되게 도달하는지만 확인한다.
- `moai model profile --harness` 표시는 실행하지 않았다. `harness_agents` 셀은 파일 등식으로만 맞췄고 CLI 표시로 재확인하지 않았다.

## 5. Residual-risk

- **로컬 llm.yaml 을 갱신하는 순간, 값이 매트릭스와 어긋나면 조용히 구값이 이긴다.** 설정 미러가 Go 기본값보다 먼저 읽히기 때문이다. 갱신 시 이 PR의 template 파일을 그대로 복사하는 편이 손편집보다 안전하다.
- `max` 는 이제 어느 셀에도 없다. 어휘에는 남아 있고 `TestDefaultProfileMatrix_Shape` 의 허용 집합도 그대로라, 나중에 어떤 행이 `max` 를 다시 집어도 테스트는 막지 않는다 — 의도된 여지다.
- `GroupLead` 신설로 `moai model profile` 의 group 컬럼에 새 값 `lead` 가 등장한다. 이 문자열을 파싱하는 외부 소비자가 있다면 영향을 받는다(리포지토리 내에는 없음).
- frontmatter 를 medium 컬럼에 고정하는 규율은 그대로다: 앞으로 medium 컬럼을 건드리면 **반드시** 해당 에이전트 파일과 `.codex` TOML을 함께 갱신해야 한다. `make agents-emit` 이 TOML 쪽 자동화다.

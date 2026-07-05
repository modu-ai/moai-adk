# research.md — SPEC-WEB-CONSOLE-CONFIG-DIET-001

> 감사 증거 캡처. 5-에이전트 감사 findings를 clean worktree(`preview-wc011`, origin/main
> `97723664c34f5a0aefc5bb3146ae5e510f4f918f`)에서 **manager-spec가 재-grep 검증**한 결과.
> 모든 file:line은 위 트리 기준. verification-claim-integrity §1.1 surface 3 준수 — 각 사강 주장은
> 도구(grep) 출력 관측에 근거하며, 미확인 항목은 명시적으로 표기한다.

## §R.0 검증 방법론

- 도구: `grep -rn ... --include="*.go" | grep -v "_test.go"` (production read만).
- "사강(dead)" 판정: 소비 심볼의 production read가 0건 (round-trip yaml load/default 설정은 소비 아님).
- 렌더 SSOT: `internal/settings/schema.go`(6 수제 섹션 + `AllSections`), `internal/settings/schema_sections.go`
  (`seamSectionFields`, `agentSettingsFields`, `ReadOnlyDisplayFields`, `RawViewBlocks`).

## §R.1 Tier 1 — 편집 가능하나 100% 무효과

### R.1.a research.yaml (12/12 DEAD) — 확인

- `internal/research/` 패키지 부재: `ls internal/research/` → `No such file or directory`.
- `ResearchConfig`는 round-trip 전용: 정의 `internal/config/types.go:593`, 기본값
  `internal/config/defaults.go:129 NewDefaultResearchConfig()`, audit 등록 `audit_registry.go:39`.
- **필드 read 0건**: `grep -rn "\.Research\." internal/ pkg/ --include="*.go" | grep -v "_test.go"` → **빈 출력**.
- 콘솔 렌더: `schema_sections.go:283–294`의 `seamSectionFields()` research 블록 12키 (enabled, active.*,
  dashboard.*, passive.*, safety.*).
- **결론**: N1 진짜 사강. 편집해도 무효과.

### R.1.b db 인터뷰 3키 — 확인

- `schema_sections.go:314–316`: `db.orm`/`db.multi_tenant`/`db.migration_tool` seam 편집 필드.
- 채움 소스 `/moai db init` 폐지(2026-05-16, 감사 근거). 소비자 재-grep 미발견.
- **결론**: N1.

### R.1.c security 스칼라 3키 — 확인

- `schema_sections.go:309–311`: `security.permission.strict_mode`/`security.sandbox.required`/
  `security.sandbox.docker_image` seam 편집 필드.
- config struct 필드 동작 read 0건: `grep -rn "Sandbox.Required\|Sandbox.DockerImage\|Security.Sandbox"
  internal/ cmd/ --include="*.go" | grep -v "_test.go" | grep -v types.go | grep -v defaults.go` → **빈 출력**.
- permission resolver의 `StrictMode`는 별개 심볼이며 `internal/cli/doctor_permission.go:73`에서 하드코딩
  `false`. sandbox의 `DockerImage` read는 `internal/sandbox/docker.go:59`/`profile.go:170`이 **자체
  `SandboxOptions.DockerImage`**를 읽는 것이지 config `SecuritySandbox.DockerImage`가 아님.
- **결론**: N1/N2. config 필드는 동작 미반영.

## §R.2 Tier 2 — 대량 사강

### R.2.a quality.yaml (18/20 DEAD) — 확인

- `core/quality.TrustGate` 존재: `internal/core/quality/trust.go:335`. 그러나 CLI 미배선:
  `grep -rn "NewTrustGate" internal/ --include="*.go" | grep -v "_test.go" | grep -v "core/quality/"` → **빈 출력**.
- `cfg.Quality`→TrustGate 매핑 부재: quality×trust×gate 교차 grep → **빈 출력**.
- 라이브: `development_mode`(schema.go:401, PersistProjectConfig), `test_coverage_target`(schema.go:408).
- **결론**: N2 미배선 게이트.

### R.2.b git-strategy (~35/57) — 확인 + **중요 결(nuance)**

- struct 자체 주석이 명시: `internal/config/types.go:53` — "these keys are currently consumed by skill bodies
  via direct yaml reads, not [by Go code]"; `types.go:71` — "Phase 3 (yaml-direct read), NOT by Go code."
- 콘솔 렌더: `schema_sections.go:131–157 gitStrategyFields()` — 3-mode 프로파일(manual/personal/team) ×
  `gitStrategyProfileLeaves`(15 leaf, schema_sections.go:110–129) 전부 렌더.
- **N3 결**: Go는 안 읽지만 **skill body가 yaml-direct read로 소비 → 효과 있음**. 따라서 "사강"이 아니라
  "중복 노출". 다이어트 각도 = 비선택 mode 프로파일 중복 축소 (삭제 금지).
- 라이브(Go): `mode`, `hooks.pre_push`(`internal/cli/hook_pre_push.go` — git-convention 경유). manager-git
  프롬프트: `automation.auto_*`, `main_branch`.

### R.2.c workflow_agents (14/14 DEAD) — 확인

- 타입 정의: `internal/config/types.go:348 WorkflowAgents map[string]WorkflowAgentEntry`,
  entry `types.go:356`.
- **맵 read 0건**: `grep -rn "\.WorkflowAgents\b" internal/ pkg/ --include="*.go" | grep -v "_test.go"` → **빈 출력**.
- 콘솔 렌더: `schema_sections.go:352–358 agentSettingsFields()` — 7 purpose × {model, effort}.
- `.claude/workflows/*.js`가 config 블록을 읽는 병합점 코드 부재(감사 주장, 재-grep 시 Go측 소비 0 확인).
- **결론**: N1/N2. typed reader test-only.

### R.2.d ralph.yaml (17/19 split-brain) — 확인

- 컴파일 loop 엔진 `internal/ralph/engine.go`가 읽는 cfg 필드: `cfg.AutoConverge`(:62),
  `cfg.HumanReview`(:74), `cfg.LintAsInstruction`/`cfg.WarnAsInstruction`(:239). **콘솔 노출 중첩 키
  (`ast_grep.*`/`loop.*`/`lsp.*`/`loop.completion.*`, schema_sections.go:265–280)는 읽지 않음.**
- `NewDefaultRalphConfig()`(`internal/config/defaults.go:364`)의 하드코딩 기본값을 엔진이 사용.
- 라이브(콘솔 노출 키 중): `lint_as_instruction`, `warn_as_instruction`.
- **결론**: 콘솔 노출 19키 중 중첩 다수가 엔진 미소비 (split-brain).

## §R.3 Tier 3 — 부분 사강

### R.3.a workflow.yaml (12/27 DEAD) — 확인

- `internal/config/workflow_accessors.go` 존재. 그러나 `auto_clear`/`token_budget`/`loop_prevention`/
  `worktree.auto_merge` accessor의 real 호출자 재-grep 미발견 (매치는 전부 무관 심볼: initializer.go의
  template-context `AutoClear`, statusline memory의 별개 `TokenBudget`).
- 콘솔 렌더: `schema_sections.go:225–242` workflow 블록.
- **결론**: N1 dead-code accessor.

### R.3.b harness.yaml (6 DEAD + 표시 불일치) — 확인

- `escalation.*`(test-only EscalationManager)/`mode_defaults.*`(미read)/`auto_detection.enabled`(no-op).
- 콘솔 렌더: `schema_sections.go:250–256`.
- **결론**: N2.

### R.3.c observability.yaml (5/7 DEAD) — 확인

- `internal/observability/` 패키지 부재. config 미러 키 `trace_dir`/`report_dir`/`max_file_size_mb`/
  `retention_days`/`hook_metrics.output_path`의 config-필드 소비자 재-grep 미발견 (매치는 `internal/runtime/
  audit_report.go`의 별개 `ReportDir`, `internal/worktree/divergence_log.go`의 별개 `reportDir`).
- 콘솔 렌더: `schema_sections.go:301–305`.
- 라이브: `enabled`, `hook_metrics.slow_hook_threshold_ms`(:300, :306).
- **결론**: N1 하드코딩 미러.

### R.3.d llm 편집 4키 (DEAD) — 확인

- `performance_tier`/`claude_models.{high,medium,low}` (llmFields, schema_sections.go:162–172).
- 모델 dispatch 소비 재-grep 미발견. `PerformanceTier`는 `internal/config/validation.go:485`의 oneof
  검증에만 등장 (값 유효성 검사이지 dispatch 소비 아님). `ClaudeModels`는 types/defaults/settings 외 read 0.
- Claude 모델은 launch/settings.json에서 해석됨.
- `llm.mode`/`llm.team_mode`는 이미 read-only(`ReadOnlyDisplayFields()`, schema_sections.go:390–391).
- **결론**: N1.

### R.3.e role_profiles model/mode/effort (DEAD) — 확인

- `LoadRoleProfiles`(`internal/cli/team_spawn.go:411`)의 production 호출자 0건:
  `grep -rn "LoadRoleProfiles" internal/ cmd/ | grep -v "_test.go" | grep -v "team_spawn.go:"` → **빈 출력**.
  `SpawnTeammate`류 호출자도 미발견.
- **라이브(중요)**: `role_profiles.*.isolation`은 `internal/cli/workflow_lint.go:70 validateRoleProfiles`가
  소비 (`moai workflow lint`, workflow_lint.go:161). 강등 금지.
- 콘솔 렌더: `schema_sections.go:344–350 agentSettingsFields()` (model/effort/isolation/mode).
- **결론**: model/mode/effort N2 미배선, isolation 라이브.

## §R.4 정합성 결함

- **결함 1 (harness prune)**: 표시 `learning.log_retention_days`=90 (config 기본값) vs 실제 prune 하드코딩.
  `internal/harness/observer.go:147 const defaultRetentionDays = 30`; `observer.go:91
  o.retention.PruneStaleEntries(defaultRetentionDays)`. **확인** — 표시 90 / prune 30 불일치.
- **결함 2 (security @MX stale)**: `internal/config/types.go:429–432` `@MX:ANCHOR`/`@MX:REASON` — "Fan_in >= 3:
  loaded by config/loader.go, consumed by sandbox/launcher.go, displayed by doctor_sandbox.go". 재검증:
  config `SecuritySandbox` 필드의 동작 read 0건(§R.1.c). "consumed by sandbox/launcher.go" 주장은 stale
  (sandbox 패키지는 자체 `SandboxOptions`를 읽음, config struct 아님). **확인** — 앵커 fan_in 과장.
- **결함 3 (db auto_sync 드리프트)**: `schema_sections.go:413 RawViewBlocks`가 `db.auto_sync`를 **nested
  block**으로 렌더. 산문(감사)은 `db.auto_sync: true` **스칼라**로 서술. 재-grep: `types.go`/`defaults.go`에
  `AutoSync` 필드 미발견 → DB struct에 스칼라 필드 없음, nested 취급이 실제. **신뢰도 낮음** — doc-level
  드리프트로 분류. 산문 정정 권장.
- **결함 4 (session_name_pattern)**: `internal/config/types.go:411 SessionNamePattern`, 기본값
  `internal/config/defaults.go:458 "moai-{ProjectName}-{SPEC-ID}"`. config 값 read 소비자 재-grep 미발견 →
  round-trip 전용, 세션 명명은 하드코딩. **확인**.

## §R.5 미확인/불일치 항목 (정직 표기 — verification-claim-integrity)

- **team.auto_selection.min_* "consumed" 주장**: 감사는 "team.auto_selection.min_* 3 keys ARE consumed but
  NOT exposed"로 주장했다. 그러나 본 재-grep은 `WorkflowTeamAutoSelection`/`MinDomainsForTeam`/
  `MinFilesForTeam`을 `internal/config/defaults.go:394–395`(**기본값 설정만**)에서만 발견하고, **production
  read 소비자를 확인하지 못했다**. 접근자 `internal/config/workflow_accessors.go:33 WorkflowTeamAutoSelection`은
  존재하나 호출자 미발견.
  → **함의**: "consumed" 주장은 본 SPEC 재검증으로 확정되지 않음. REQ-CD-050(SHOULD)의 해당 키 노출은
  **소비자 재확인을 전제조건**으로 한다 (미확인 라이브 주장으로 노출 금지). 이 불일치는 다이어트 제거
  범위(M1~M4)에는 영향 없음 — 해당 키는 현재 미노출이므로 어느 쪽이든 제거 대상이 아니다.
- **ralph AutoConverge/HumanReview**: 엔진이 `cfg.AutoConverge`/`cfg.HumanReview`도 읽으나(§R.2.d), 이들은
  콘솔 노출 19키 집합 밖의 별개 RalphConfig 필드다. 따라서 "콘솔 노출 중첩 키는 미소비"라는 결론과 모순
  없음 (감사의 "only lint/warn consumed"는 노출 키 부분집합 기준으로 정확).

## §R.6 요약 카운트 (콘솔 노출 dead 후보)

| Tier | 그룹 | dead 후보 키 | 결 | 검증 |
|------|------|-------------|----|------|
| 1 | research | 12 | N1 | 확인 (빈 read) |
| 1 | db 인터뷰 | 3 | N1 | 확인 |
| 1 | security 스칼라 | 3 | N1/N2 | 확인 |
| 2 | quality | ~18 | N2 | 확인 (TrustGate 미배선) |
| 2 | git-strategy 중복 | ~35 | **N3** | 확인 — **삭제 금지**, 중복 축소 |
| 2 | workflow_agents | 14 | N1/N2 | 확인 (빈 read) |
| 2 | ralph 중첩 | ~17 | N1 | 확인 (엔진 미소비) |
| 3 | workflow | 12 | N1 | 확인 (dead accessor) |
| 3 | harness | 6 | N2 | 확인 |
| 3 | observability | 5 | N1 | 확인 |
| 3 | llm | 4 | N1 | 확인 |
| 3 | role_profiles model/mode/effort | 14 | N2 | 확인 (isolation 라이브) |
| 버그 | 4건 | — | — | 4/4 확인(3은 신뢰도 낮음) |

> 카운트는 감사 제시값 재확인 수준이며, 정확한 필드 수는 구현 시 `allFields()` 열거로 재도출한다
> (feedback: Ground Truth arithmetic — 표 숫자 불신, LIVE 재도출).

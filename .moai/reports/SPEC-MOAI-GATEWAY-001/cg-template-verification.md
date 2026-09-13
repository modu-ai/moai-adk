# CG 배포 template 본문 검증

## Claim

배포 template 38개에서 현재 CG 실행·역할·비용보장·모드 분기 지시를 정리했다. 신규 artifact metadata/dispatch는 만들지 않았다. 부모가 추가 승인한 WT `.claude` source mirror 다섯 파일에도 동일 CG 변경분만 적용했다. primary checkout은 수정하지 않았다. 허용된 신규 시험은 `internal/template/cg_retirement_test.go` 하나다.

`glm-web-tooling.md`의 기존 CG 운영 절차를 명시 retirement/migration SSOT로 교체했다. `moai migrate cg` preview, claude-only 적용의 `--apply --accept-role-change`, 자동 GLM 팀원 배정 제거, private backup, hybrid 실제 capability gate와 사용자 verified 우회 금지를 서술한다. GLM 전체 세션의 MCP 검색·reader·vision routing/등록/입력 규칙은 유지했다. Claude-backed 세션 예외는 현재 backend 기준이며 과거 CG leader로 추론하지 않는다.

CLAUDE §15와 native Agent Teams 실험 허용은 유지하되 CG 비용 60-70% 보장 및 실행 근거를 제거했다. AGENTS 명령표와 worktree/plan/run/factory 안내는 CG를 지원 launcher로 취급하지 않는다. 기존 CG를 다른 명령으로 바꿔도 동일 혼합 역할이라는 주장을 하지 않았다. `harness.mode_defaults.cg`를 제거하고 solo/team 자동 깊이와 always-on plan audit를 보존했다.

sync-auditor·contract negotiation·evaluation의 CG leader-inline 예외를 제거했다. super-advisor의 일반 자문은 유지하지만 독립 audit를 대신하지 못한다고 명시했다. model-policy/run active_mode도 legacy CG로 dispatch하지 않는다. settings-management의 v2.1.119 버전 행은 역사 문맥이므로 보존했다.

## Evidence

각 Go 명령은 같은 invocation에서 `unset ANTHROPIC_BASE_URL ANTHROPIC_API_KEY ANTHROPIC_AUTH_TOKEN Z_AI_API_KEY && GOCACHE=/tmp/gateway-foundation-cache` 뒤에 실행했다. 아래 시험은 실제 `EmbeddedTemplates()`를 `NewRenderer()`에 넣고 렌더한다. 합성 문구만 검사하지 않는다. 필수 preview/apply/role-change 정책, GLM MCP 경로, native team allowance, 실제 렌더한 harness YAML의 cg 부재·solo/team·always-enabled audit, 독립 evaluator 예외 부재를 판정한다.

### RED — 본문 수정 전

`go test ./internal/template -run TestCGEmbeddedRetirement -timeout 30s`

```text
--- FAIL: TestCGEmbeddedRetirementPreservesRoutingAndAudit (0.00s)
    cg_retirement_test.go:27: rendered web doctrine lacks "moai migrate cg"
    cg_retirement_test.go:27: rendered web doctrine lacks "--accept-role-change"
    cg_retirement_test.go:27: rendered web doctrine lacks "--apply"
    cg_retirement_test.go:27: rendered web doctrine lacks "claude-only"
    cg_retirement_test.go:27: rendered web doctrine lacks "claude-glm"
    cg_retirement_test.go:27: rendered web doctrine lacks "verified"
    cg_retirement_test.go:32: rendered retired instruction: Enable CG mode inside tmux
    cg_retirement_test.go:32: rendered retired instruction: moai cg` injects these
    cg_retirement_test.go:37: native team allowance or explicit migration missing
    cg_retirement_test.go:40: CG cost guarantee survived
    cg_retirement_test.go:54: CG still selects harness depth
    cg_retirement_test.go:65: CG inline audit exception survives .claude/skills/moai/workflows/run/task-decomposition.md
    cg_retirement_test.go:65: CG inline audit exception survives .claude/skills/moai/workflows/run/phase-execution.md
    cg_retirement_test.go:65: CG inline audit exception survives .claude/agents/moai/sync-auditor.md
    cg_retirement_test.go:65: CG inline audit exception survives .codex/agents/moai/sync-auditor.toml
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/template	0.477s
FAIL
```

### GREEN — 최종 본문

`go test ./internal/template -run 'TestCGEmbedded|TestRendererRender|TestEmbeddedTemplates_CLAUDEmd|TestDeployerDeploy|TestCodexAgentsDeployFixture' -timeout 60s`

```text
ok  	github.com/modu-ai/moai-adk/internal/template	0.643s
```

`go vet ./internal/template`: exit 0, stdout/stderr 없음.

Python으로 현재 파일과 `git show HEAD:<path>`의 YAML 앞머리, TOML developer_instructions 밖의 문자열, MD/TOML 본문을 비교했다:

```text
modified_template_files= 38 frontmatter_unchanged= 23
manager-lead metadata_unchanged= True body_parity= True
super-advisor metadata_unchanged= True body_parity= True
sync-auditor metadata_unchanged= True body_parity= True
historical v2.1.119 row unchanged: True
```

처음 metadata 확인에 사용하려던 Python tomllib은 환경에 없어 `ModuleNotFoundError`였다. 설치하거나 실패를 성공으로 기록하지 않고, 위의 실제 문자열 경계 비교와 기존 Codex deploy fixture 시험을 사용했다. 새 TOML parser 성공을 주장하지 않는다.

### source/template parity — 최초 FAIL과 승인 후 해결

`go test ./internal/template -run 'TestCGEmbedded|TestRuleTemplateMirrorDrift|TestLateBranchTemplateMirror|TestEmbeddedTemplates_CLAUDEmd|TestRendererRender|TestDeployerDeploy' -timeout 60s` 결과는 다음과 같다. 부모에게 다섯 정확한 쌍을 통지했다. 원본을 복사해 배포 수정 내용을 되돌리거나 사용자 원본을 무단 변경하지 않았다.

```text
--- FAIL: TestRuleTemplateMirrorDrift (0.00s)
    --- FAIL: TestRuleTemplateMirrorDrift/model-policy.md (0.00s)
        rule_template_mirror_test.go:169: RULE_TEMPLATE_MIRROR_DRIFT: source file .claude/rules/moai/development/model-policy.md differs from its mirror at /Users/goos/MoAI/moai-adk-go/.claude/worktrees/moai-proxy-unified/internal/template/templates/.claude/rules/moai/development/model-policy.md (source 33954 bytes, mirror 34183 bytes); run 'cp /Users/goos/MoAI/moai-adk-go/.claude/worktrees/moai-proxy-unified/.claude/rules/moai/development/model-policy.md /Users/goos/MoAI/moai-adk-go/.claude/worktrees/moai-proxy-unified/internal/template/templates/.claude/rules/moai/development/model-policy.md' and stage both files
    --- FAIL: TestRuleTemplateMirrorDrift/spec-workflow.md (0.00s)
        rule_template_mirror_test.go:169: RULE_TEMPLATE_MIRROR_DRIFT: source file .claude/rules/moai/workflow/spec-workflow.md differs from its mirror at /Users/goos/MoAI/moai-adk-go/.claude/worktrees/moai-proxy-unified/internal/template/templates/.claude/rules/moai/workflow/spec-workflow.md (source 40124 bytes, mirror 40234 bytes); run 'cp /Users/goos/MoAI/moai-adk-go/.claude/worktrees/moai-proxy-unified/.claude/rules/moai/workflow/spec-workflow.md /Users/goos/MoAI/moai-adk-go/.claude/worktrees/moai-proxy-unified/internal/template/templates/.claude/rules/moai/workflow/spec-workflow.md' and stage both files
    --- FAIL: TestRuleTemplateMirrorDrift/worktree-integration.md (0.00s)
        rule_template_mirror_test.go:169: RULE_TEMPLATE_MIRROR_DRIFT: source file .claude/rules/moai/workflow/worktree-integration.md differs from its mirror at /Users/goos/MoAI/moai-adk-go/.claude/worktrees/moai-proxy-unified/internal/template/templates/.claude/rules/moai/workflow/worktree-integration.md (source 48555 bytes, mirror 48199 bytes); run 'cp /Users/goos/MoAI/moai-adk-go/.claude/worktrees/moai-proxy-unified/.claude/rules/moai/workflow/worktree-integration.md /Users/goos/MoAI/moai-adk-go/.claude/worktrees/moai-proxy-unified/internal/template/templates/.claude/rules/moai/workflow/worktree-integration.md' and stage both files
    --- FAIL: TestRuleTemplateMirrorDrift/session-handoff-examples.md (0.00s)
        rule_template_mirror_test.go:169: RULE_TEMPLATE_MIRROR_DRIFT: source file .claude/rules/moai/workflow/session-handoff-examples.md differs from its mirror at /Users/goos/MoAI/moai-adk-go/.claude/worktrees/moai-proxy-unified/internal/template/templates/.claude/rules/moai/workflow/session-handoff-examples.md (source 40449 bytes, mirror 40401 bytes); run 'cp /Users/goos/MoAI/moai-adk-go/.claude/worktrees/moai-proxy-unified/.claude/rules/moai/workflow/session-handoff-examples.md /Users/goos/MoAI/moai-adk-go/.claude/worktrees/moai-proxy-unified/internal/template/templates/.claude/rules/moai/workflow/session-handoff-examples.md' and stage both files
--- FAIL: TestLateBranchTemplateMirror (0.00s)
    --- FAIL: TestLateBranchTemplateMirror/spec-assembly.md (0.00s)
        rule_template_mirror_test.go:229: RULE_TEMPLATE_MIRROR_DRIFT: source file .claude/skills/moai/workflows/plan/spec-assembly.md differs from its mirror at /Users/goos/MoAI/moai-adk-go/.claude/worktrees/moai-proxy-unified/internal/template/templates/.claude/skills/moai/workflows/plan/spec-assembly.md (source 32568 bytes, mirror 32622 bytes); run 'cp /Users/goos/MoAI/moai-adk-go/.claude/worktrees/moai-proxy-unified/.claude/skills/moai/workflows/plan/spec-assembly.md /Users/goos/MoAI/moai-adk-go/.claude/worktrees/moai-proxy-unified/internal/template/templates/.claude/skills/moai/workflows/plan/spec-assembly.md' and stage both files
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/template	0.488s
FAIL
```

부모가 아래 다섯 WT source 경로에 한정하여 소유권을 확장했다. 수정 직전 각 source가 HEAD source 및 HEAD template와 동일함을 확인한 뒤, template의 변경 전후 차이를 문맥이 있는 hunk로 적용했다. 전체 파일 복사는 하지 않았다.

```text
.claude/rules/moai/development/model-policy.md targeted_hunks= 3 template_parity= True
.claude/rules/moai/workflow/spec-workflow.md targeted_hunks= 2 template_parity= True
.claude/rules/moai/workflow/worktree-integration.md targeted_hunks= 2 template_parity= True
.claude/rules/moai/workflow/session-handoff-examples.md targeted_hunks= 4 template_parity= True
.claude/skills/moai/workflows/plan/spec-assembly.md targeted_hunks= 3 template_parity= True
```

최종 명령(위와 같은 단일 invocation env scrub):

`go test ./internal/template -run 'TestCGEmbedded|TestRuleTemplateMirrorDrift|TestLateBranchTemplateMirror|TestEmbeddedTemplates_CLAUDEmd|TestRendererRender|TestDeployerDeploy|TestCodexAgentsDeployFixture' -count=1 -timeout 60s`

```text
ok  	github.com/modu-ai/moai-adk/internal/template	0.515s
```

## Baseline-attribution

2026-09-11, WT `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/moai-proxy-unified`, 마지막 HEAD `81c1d58f9`. 같은 WT의 CG runtime 철거와 audited CG-RETIRE 0.1.0을 입력으로 삼았다. 모든 template는 수정할 문맥을 읽은 뒤 파일별 구문을 지정하여 변경했다. README/docs-site는 다른 writer 소유로 손대지 않았다. git mutation/실제 Claude/provider 호출 없음.

변경 파일:

- `internal/template/templates/.claude/agents/moai/manager-lead.md`
- `internal/template/templates/.claude/agents/moai/super-advisor.md`
- `internal/template/templates/.claude/agents/moai/sync-auditor.md`
- `internal/template/templates/.claude/output-styles/moai/moai-learn.md`
- `internal/template/templates/.claude/rules/moai/core/agent-common-protocol.md`
- `internal/template/templates/.claude/rules/moai/core/glm-web-tooling.md`
- `internal/template/templates/.claude/rules/moai/core/moai-constitution.md`
- `internal/template/templates/.claude/rules/moai/core/settings-management.md`
- `internal/template/templates/.claude/rules/moai/development/agent-authoring.md`
- `internal/template/templates/.claude/rules/moai/development/model-policy.md`
- `internal/template/templates/.claude/rules/moai/development/orchestrator-templates.md`
- `internal/template/templates/.claude/rules/moai/workflow/context-window-management.md`
- `internal/template/templates/.claude/rules/moai/workflow/cross-session-messaging-detail.md`
- `internal/template/templates/.claude/rules/moai/workflow/cross-session-messaging.md`
- `internal/template/templates/.claude/rules/moai/workflow/orchestration-mode-selection.md`
- `internal/template/templates/.claude/rules/moai/workflow/session-handoff-examples.md`
- `internal/template/templates/.claude/rules/moai/workflow/spec-workflow.md`
- `internal/template/templates/.claude/rules/moai/workflow/worktree-integration.md`
- `internal/template/templates/.claude/skills/moai-foundation-quality/modules/integration-patterns.md`
- `internal/template/templates/.claude/skills/moai-workflow-worktree/SKILL.md`
- `internal/template/templates/.claude/skills/moai-workflow-worktree/modules/moai-adk-integration.md`
- `internal/template/templates/.claude/skills/moai/SKILL.md`
- `internal/template/templates/.claude/skills/moai/workflows/factory.md`
- `internal/template/templates/.claude/skills/moai/workflows/goal.md`
- `internal/template/templates/.claude/skills/moai/workflows/moai.md`
- `internal/template/templates/.claude/skills/moai/workflows/plan/spec-assembly.md`
- `internal/template/templates/.claude/skills/moai/workflows/run/mode-orchestration.md`
- `internal/template/templates/.claude/skills/moai/workflows/run/phase-execution.md`
- `internal/template/templates/.claude/skills/moai/workflows/run/task-decomposition.md`
- `internal/template/templates/.codex/agents/moai/manager-lead.toml`
- `internal/template/templates/.codex/agents/moai/super-advisor.toml`
- `internal/template/templates/.codex/agents/moai/sync-auditor.toml`
- `internal/template/templates/.moai/config/sections/crosssession.yaml`
- `internal/template/templates/.moai/config/sections/harness.yaml`
- `internal/template/templates/.moai/config/sections/llm.yaml`
- `internal/template/templates/.moai/docs/generic-patterns-guide.md`
- `internal/template/templates/AGENTS.md`
- `internal/template/templates/CLAUDE.md`

## Gaps

- CR006 전체 판정은 아직 하지 않았다. 이번 범위의 source/template parity 다섯 쌍은 승인 후 동기화하고 실제 mirror 시험으로 확인했다. README/docs-site 등 나머지 CR006 범위는 별도 검증 대상이다.
- 실제 embedded 렌더 시험은 핵심 정책·harness·독립 감사 파일에 집중한다. 모든 38개 Markdown을 실제 클라이언트에 배포하여 실행한 시험은 아니다. 기존 deploy 시험은 그 시험 자체의 fixture 범위만 증명한다.
- README/네 언어 docs-site 빌드·링크·브라우저 검증은 별도 writer 범위다. 실제 native TEAMMATE 혼합 provider capability와 production gateway gate는 미검증 상태를 유지한다.
- 저장소 전체 verdict는 통합 브랜치 CI가 소유하며 PENDING이다. commit/push/PR/merge를 하지 않았다.

## Residual-risk / 남긴 CG 참조와 이유

`CG/cg` 문자열 0을 목표로 삼지 않았다. 최종 token 검색의 42개 행은 다음 문맥으로 남긴다. 이 검색 숫자는 의미 검증을 대신하지 않는다.

- core/glm-web-tooling.md: legacy config 차단, `moai migrate cg`, role-change 명령, hybrid gate, 옛 inline 예외 금지. 현재 실행 권고가 아니다.
- CLAUDE.md·AGENTS.md 및 development/model-policy.md: 명시 폐기·이전 안내와 native team 허용의 분리.
- settings-management.md:53 v2.1.119 행: 이전 버전 tools/disallowedTools 행동의 역사 설명. HEAD의 그 행과 동일하다.
- sync-auditor 및 super-advisor MD/TOML, run/task-decomposition·phase-execution: CG가 독립 audit를 우회하지 못한다는 제한.
- plan/spec-assembly·run/mode-orchestration·workflows/moai/factory·session-handoff-examples: legacy CG 발견 시 중단/preview 안내. 다른 provider로 자동 대체하지 않는다.
- worktree-integration·generic-patterns-guide·orchestration-mode-selection·spec-workflow·agent-authoring·manager-lead: native display/팀 가용성과 retired CG 혼합 라우팅을 구별한다.

이번 수정은 독립 코드/문서 감사 판정이 아니다. 일반 GLM·native Teams 관련 기존 제품 사실 전체를 새로 실계정 검증한 것으로 읽으면 안 된다. metadata와 역사 행 보존을 확인했지만 사용자 설치본의 상태는 별개다.

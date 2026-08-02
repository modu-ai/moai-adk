# Local Dev Settings Intent — local settings.json 의도 명문화

> Extracted from `CLAUDE.local.md` §22 for context-budget diet. Canonical reference for local `.claude/settings.json` key intent. The orchestrator reads this on demand instead of carrying it in every session/agent prefix.

## 22. Dev Settings Intent — local settings.json 의도 명문화

Workflow audit 2026-05-16 finding M2 후속. 로컬 `.claude/settings.json`의 몇 가지 키는 template baseline과 의도적으로 다르게 운용되며, 본 섹션은 그 의도를 명문화한다.

### §22.1 defaultMode

- **로컬값**: `"bypassPermissions"` 또는 `"acceptEdits"` (개발자 선호)
- **Template 기본값**: 미지정 (Claude Code 기본 `"default"` 사용)
- **의도**: 메인테이너는 빠른 실험 + bypass 모드 빈번 사용. 사용자 프로젝트는 안전한 prompt-each-time 기본값을 따른다.

### §22.2 enableAllProjectMcpServers

- **로컬값**: `true`
- **Template 기본값**: 미지정 (false 효과)
- **의도**: 메인테이너 머신에는 dev tool 다수 (pencil, chrome-devtools, claude-in-chrome 등)가 등록되어 있어 모두 자동 활성화하는 것이 효율적. 사용자 프로젝트는 `mcp__context7`만 `alwaysLoad`되고 나머지는 ToolSearch preload 경로를 따른다. (Sequential Thinking MCP는 SPEC-V3R6-SEQ-THINKING-RETIRE-001에서 retired — `ultrathink` 키워드로 대체.)

### §22.3 teammateMode

- **로컬값**: `"tmux"` (CG/GLM 모드 — `moai cg` / `moai glm` 진입 시) 또는 `""` (CC 모드 — `moai cc` 진입 시 override 해제). runtime-managed by `moai cg` / `moai glm` / `moai cc` 명령. 코드가 실제로 기록하는 값은 이 둘뿐이다 (`internal/cli/glm.go` `ensureSettingsLocalJSON`/`injectGLMEnvForTeam` 가 `"tmux"`, `internal/cli/launcher.go` `removeGLMEnv` 가 `""`). `"glm"`/`"claude"` 는 코드가 기록하지 않는다.
- **`llm.yaml team_mode` 와 구분 (혼동 금지)**: `.claude/settings.local.json` 의 `teammateMode` 필드(`"tmux"`/`""`)와 `.moai/config/sections/llm.yaml` 의 `team_mode` 필드(`cg`/`glm`/`""`)는 서로 다른 필드다. 전자는 Claude Code teammate 표시 모드(tmux pane vs inline)를 결정하고, 후자는 CG 모드 감지 SSOT (`internal/tmux/cg_detect.go` `IsCGMode` 가 `team_mode == "cg"` 를 읽는다, REQ-CGH-006). 같은 "team/teammate" 어휘를 공유하나 위치·값·용도가 다르다.
- **Template 기본값**: 미지정 (양쪽 필드 모두)
- **의도**: 메인테이너는 CG mode (Claude leader + GLM teammates)로 cost-optimization 검증을 빈번하게 수행. 사용자 프로젝트는 leader 단독으로 시작 후 필요시 `moai cg` 진입.
- **주의**: §2 [HARD] settings.local.json Separation 참조 — teammateMode는 `settings.local.json`에 위치하며 template에 절대 포함 금지.

### §22.4 env.PATH

- **위치 (2026-07-12 이관)**: machine-specific 절대경로 PATH는 이제 `settings.local.json`(gitignored)에만 둔다. `.claude/settings.json`(git-tracked)의 `env`에서 `PATH` 키를 **제거**했다 — git-tracked 파일에 `/Users/goos/...` 절대경로를 두면 fork/clone 사용자에게 깨지기 때문(§2 settings.local.json Separation 적용). settings.json의 `env`에는 machine-independent 키(`CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS`, `ENABLE_TOOL_SEARCH`, `MOAI_CONFIG_SOURCE`)만 남긴다.
- **왜 절대경로인가 (settings.local.json 안에서)**: **Claude Code가 env.PATH 값의 `$HOME`을 expand하지 않음**이 실측 확인됨 (Bash 서브프로세스에 리터럴 `$HOME/go/bin`이 그대로 전달 → `command -v moai` 실패 → moai-lsp MCP가 PATH로 moai 미해석). 따라서 settings.local.json의 PATH는 절대경로를 쓴다 (직전 2026-05-17 F-009/M5의 `$HOME` 권고는 이 실측으로 뒤집힘).
- **Template 기본값**: `settings.json.tmpl`의 PATH 키는 `{{jsonEscape .SmartPATH}}` 로 렌더 (BuildSmartPATH가 issue #467 대응으로 well-known 절대경로 PATH 생성 — `$HOME` 미사용). 사용자 프로젝트는 `moai init`/`moai update`가 절대 SmartPATH로 새로 렌더하므로 fork/clone과 무관. **주의**: `settings_test.go:requiredKeys`가 템플릿 env에 `PATH`/`ENABLE_TOOL_SEARCH`/`CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS` 존재를 강제하므로 템플릿에서 이 키들을 제거하면 CI FAIL — 로컬 settings.json에서만 PATH를 뺀다(템플릿은 그대로).
- **주의 (2026-06-24 정정)**: `StatusLine` command는 Claude Code 내장 토큰 `$CLAUDE_PROJECT_DIR`을 런타임 환경변수로 받는다 (공식 문서 code.claude.com/docs/en/statusline — "runs with the same environment variables as hooks, including `CLAUDE_PROJECT_DIR`"). 종전 "StatusLine은 env var를 expand하지 않는다"는 GitHub Issue #7925를 오독한 것이며, 해당 이슈는 일반 shell 보간/`env` 블록 값에 대한 것이지 내장 토큰이 아니다. 단, **`env.PATH`의 `$HOME` expand 불가는 별개 사실로 여전히 유효**하며 본 절의 핵심 근거는 이쪽으로 한정한다.

### §22.6 outputStyle — 제품 기본값 MoAI-Easy (2026-07-12 채택)

- **템플릿 기본값 (제품 결정)**: `settings.json.tmpl` + 로컬 `.claude/settings.json` 모두 `"outputStyle": "MoAI-Easy"` 고정. 모든 배포 사용자가 기본 MoAI-Easy로 시작한다. 과거 `699e59631`의 "personal preference라 unpin" 결정을 제품 기본값 채택으로 되돌린 것 — CI 가드 `TestSettingsTemplateOutputStyle`도 "MoAI-Easy 필수"로 반전.
- **우선순위 (공식 검증)**: Claude Code outputStyle 해석 순위는 `settings.local.json`(local, 최고) > `settings.json`(project) > `~/.claude/settings.json`(user) > 하드코딩 기본값. 종전 MoAI 내부 문서(settings-management.md)가 **local 스코프를 누락**했었고(2026-07-12 정정), 이 누락이 "outputStyle이 계속 리셋된다"는 오진의 근원이었다.
- **`/config` 저장 위치 (공식)**: `/config` → Output style은 `settings.local.json`에 저장(공식 문서 code.claude.com/docs/en/output-styles). local이 최고 우선순위라 사용자의 `/config` 선택은 항상 템플릿 project 핀을 이긴다 → 제품 기본값 고정이 사용자를 가두지 않는다.
- **적용 시점**: outputStyle은 세션 시작 시 1회만 읽힘 → 변경은 `/clear` 또는 새 세션부터 반영. 이것이 "리셋처럼 보이는" 현상의 실제 원인(버그 아님).
- **코드 보존**: `settings.local.json` writer(`glm.go`/`settings.go`)는 `map[string]any` round-trip으로 outputStyle 등 unknown top-level 키를 보존(SPEC-CLIFIX-CRITICAL-001) → `moai glm`/`cg`/`cc`가 파일을 건드려도 outputStyle 유지.

### §22.5 운영 원칙

- [HARD] 메인테이너 머신에서 위 키들을 변경할 때 template 자동 동기화 금지 (§2 settings.local.json Separation 적용)
- [HARD] 위 키들(defaultMode/enableAllProjectMcpServers/teammateMode/env.PATH/outputStyle)의 의도가 변경되면 본 §22를 즉시 갱신
- [HARD] 사용자 프로젝트에 machine-specific 키(env.PATH 등)가 누락된 것이 정상 — 누락은 결함이 아니라 의도된 격리. 단 outputStyle은 예외: 템플릿에 MoAI-Easy로 고정되므로 배포 사용자에게도 존재한다.

### §22.7 model — 로컬 project settings.json에서 의도적 미탑재 (last-choice 존중, 2026-07-24)

- **로컬값**: 로컬 `.claude/settings.json`에 `model` 키 **없음** (의도적 제거). Template `settings.json.tmpl:398`의 `"model": "sonnet"` 제품 기본값은 **그대로 보존**.
- **왜 제거했나 (근원)**: Claude Code 모델 우선순위는 `local settings.json > project settings.json > user settings > 하드코딩`. `/model`의 "saved as default for new sessions"는 **user 스코프**(`~/.claude/settings.json`)에 저장되는데, project `settings.json`의 `model: sonnet` pin(스코프2)이 이를 항상 덮어 **재시작마다 sonnet 복귀** 트랩이 발생. project pin을 제거하면 user 스코프의 `/model` 마지막 선택(현재 `opus[1m]`)이 이겨서 last-choice가 존중된다. (`outputStyle`은 `/config`가 최상위 local 스코프에 써서 안 가두지만, `model`은 `/model`이 하위 user 스코프에 써서 가두는 비대칭이 근본 원인.)
- **effort는 무관**: user `~/.claude/settings.json`의 `effortLevel: high`를 덮는 pin이 project/local 어디에도 없으므로 effort는 이미 유지됨(순수 `claude` + moai 프로파일 비어있는 `moai cc/glm/cg` 두 경로 공통). "effort도 리셋"은 model 리셋에 묶인 착시.
- **update 경로 주의 (drift)**: `moai update` clean-reinstall은 user의 `.claude/settings.json`을 wholesale 보존하므로 대개 살아남지만(`update_clean_install.go` Step 4.5 `MergeUserFiles`), 3-way 병합 경로는 template sonnet pin을 재도입할 수 있다. 재도입되면 이 절 근거대로 다시 제거. **제품(템플릿) 기본값 변경이 아님** — 배포 사용자는 여전히 sonnet cost-lever 기본값을 받는다.
- [HARD] 이 로컬 미탑재는 **의도된 격리**(§22.5와 동일 원칙) — 감사/동기화 시 "결함"으로 되돌리지 말 것.

### §22.8 web worktree auto-toggles default OFF (EnterWorktree-first policy, 2026-07-28)

- **목적 (intent)**: `internal/config/defaults.go`의 `WorkflowWorktreeConfig` 세 토글 — `AutoCleanup`, `AutoCreate`, `AutoMerge` — 모두 **기본 `false`**. 웹 콘솔의 worktree 자동화는 사용자의 **명시적 opt-in** (웹 토글 ON 또는 `.moai/config/sections/workflow.yaml`의 `workflow.worktree.*` 키 `true` 설정)이 있을 때만 동작한다.
- **Why**: SPEC-WORKTREE-ENTRY-STRATEGY-001 M1 (commit `2fdf77714`)에서 `AutoCleanup: true → false`, `AutoMerge: true → false`로 mutated. 배경 — worktree 자동 정리/자동 병합은 사용자가 의도하지 않은 sprawl을 조장하고, EnterWorktree-first policy (`.claude/rules/moai/workflow/worktree-integration.md` § `EnterWorktree` / `ExitWorktree` Tools)와 충돌한다. 기본 OFF는 "worktree 자동화는 사용자 선택" 원칙을 코드로 정합시킨다.
- **제품(템플릿) 기본값과 동일**: 템플릿 `defaults.go` 또한 세 토글 모두 `false` (CLAUDE.local.md §2 [HARD] Template-First Rule 정합). 배포 사용자에게도 동일한 기본 OFF가 적용된다.
- **`TmuxPreferred: true`는 본 절 범위 밖**: `defaults.go`의 `TmuxPreferred: true`는 SPEC-WORKTREE-ENTRY-STRATEGY-001 OQ-4 결정에 따라 명시적으로 OUT OF SCOPE — 변경 없음 (§22 운영 원칙 §22.5 참조).
- [HARD] `defaults.go`의 세 토글 기본 `false`는 **의도된 정책**. 감사/동기화 시 "결함"으로 되돌리지 말 것. 기본값 토글은 별도 SPEC 통해서만.

### §22.9 branch_guard.enabled — 메인테이너 로컬 opt-in (기본 OFF, 2026-07-30)

- **목적 (intent)**: `internal/config/defaults.go`의 `BranchGuardConfig.Enabled`는 **기본 `false`**. Main-Checkout Branch-State Guard (`internal/hook/branch_guard.go`)는 분산 기본값으로 INERT하게 배포된다 — 단일 개발자 저장소에는 공유 체크아웃 위험이 없으므로 가드가 간섭하지 않는다. 다중 세션 공유 체크아웃을 운영하는 메인테이너만 로컬에서 opt-in한다.
- **Why**: SPEC-WORKTREE-BRANCH-GUARD-OPTIN-001. 종전 가드는 템플릿을 통해 모든 사용자에게 default-on으로 배포되었고, 읽기 전용 명령(`git stash list`, `git merge-base`)까지 과잉 차단했다. v1.2.0에서 (a) 기본 OFF 전환 + (b) 읽기 전용 패턴 정제로 수정. 템플릿 중립성 (§25) 준수를 위해 `internal/template/templates/` 어디에도 `enabled: true` 가 없다.
- **메인테이너 opt-in 경로**: 로컬 config local 티어 (`.moai/config/local/*.yaml`, resolver.go `loadLocalTier`가 읽는 유일한 로컬 오버라이드 티어)에 다음 파일을 둔다:
  ```yaml
  # .moai/config/local/workflow.yaml
  workflow:
    branch_guard:
      enabled: true
  ```
  이 파일은 메인테이너 머신 로컬 전용이다 (아래 BLOCKER 참고).
- **제품(템플릿) 기본값과 동일**: 템플릿 `defaults.go` + `internal/template/templates/.moai/config/sections/workflow.yaml` 모두 `enabled: false` (또는 키 부재 → Go zero-value `false`). §2 [HARD] Template-First Rule 정합.
- **면제 로직 보존**: `MOAI_BRANCH_GUARD_EXEMPT=1` env + `AgentType == "manager-git"` 신원 검사는 변경 없이 enabled 경로에서만 참조된다 (REQ-6 backward compat).
- **gitignore (해결됨)**: `.moai/config/local/` 디렉터리는 이제 `.gitignore`에 등록되었다 (chore 커밋). 따라서 메인테이너 opt-in 파일(`.moai/config/local/workflow.yaml`)을 생성해도 공개 저장소에 커밋되지 않는다. 종전 이 경로가 gitignore에 없어 분산 기본값이 `true`로 뒤집힐 위험이 있었으나 해결됨 (§25 정합).
- [HARD] `defaults.go`의 `BranchGuard.Enabled` 기본 `false`는 **의도된 정책**. 감사/동기화 시 "결함"으로 되돌리지 말 것. 템플릿 기본값 변경은 별도 SPEC 통해서만.

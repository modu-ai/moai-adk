# SPEC-GLM-001: GLM Compatibility Automation

**Status**: Implemented
**Priority**: High
**Created**: 2026-03-31
**Author**: MoAI + GOOS

## Summary

Automate `CLAUDE_CODE_DISABLE_EXPERIMENTAL_BETAS=1` and `DISABLE_PROMPT_CACHING=1` injection/removal across `moai cc/cg/glm` mode transitions. Add `glm-4.5` to the supported model list alongside existing `glm-4.5-air`.

## Background

Z.AI's Anthropic-compatible endpoint does not support Claude Code's experimental beta headers (`anthropic-beta: structured-outputs-*`, `interleaved-thinking-*`, etc.). Setting `CLAUDE_CODE_DISABLE_EXPERIMENTAL_BETAS=1` strips these headers, resolving Error 1214 and 400 errors. This was confirmed by:
- Claude Code Changelog (March 2026): Two bug fixes for DISABLE_BETAS
- GitHub Issue #11960: settings.json loading fix (closed/resolved)
- Community guides: Standard fix for third-party proxy compatibility
- GOOS field testing: GLM-5.1 `/moai:e2e` works with this setting

Currently users must manually add this to settings.local.json. `moai cg/glm/cc` commands should automate this.

## Requirements (EARS Format)

### REQ-1: Auto-inject compatibility env vars in GLM modes
**When** `moai glm` or `moai cg` is executed,
**the system shall** automatically inject `CLAUDE_CODE_DISABLE_EXPERIMENTAL_BETAS=1` and `DISABLE_PROMPT_CACHING=1`.

**Acceptance Criteria:**
- [x] AC-1.1: `setGLMEnv()` sets both vars in process env
- [x] AC-1.2: `injectTmuxSessionEnv()` includes both vars in tmux session env
- [x] AC-1.3: `injectGLMEnvForTeam()` includes both vars in settings.local.json
- [x] AC-1.4: `buildGLMEnvVars()` includes both vars in the env map

### REQ-2: Auto-remove compatibility env vars on CC mode
**When** `moai cc` is executed,
**the system shall** remove `CLAUDE_CODE_DISABLE_EXPERIMENTAL_BETAS` and `DISABLE_PROMPT_CACHING` from settings.local.json and tmux session env.

**Acceptance Criteria:**
- [x] AC-2.1: `removeGLMEnv()` deletes both vars from settings.local.json
- [x] AC-2.2: `clearTmuxSessionEnv()` includes both vars in the clear list

### REQ-3: Add GLM-4.5 to supported model defaults
**When** the default GLM model configuration is initialized,
**the system shall** include `glm-4.5` as an available model option.

**Acceptance Criteria:**
- [x] AC-3.1: `defaults.go` defines `DefaultGLM45` constant for `glm-4.5`
- [x] AC-3.2: Template `llm.yaml` includes `glm-4.5` in available models documentation/comments
- [x] AC-3.3: `glm.go` display message references `glm-4.5` as available

### REQ-4: Tests pass
- [x] AC-4.1: Existing GLM tests updated with DISABLE_BETAS/DISABLE_PROMPT_CACHING assertions
- [x] AC-4.2: `go test ./internal/cli/...` passes
- [x] AC-4.3: `go vet ./...` passes
- [x] AC-4.4: `make build` succeeds

## Files to Modify

| File | Change |
|------|--------|
| `internal/cli/glm.go` | 4 functions: setGLMEnv, injectTmuxSessionEnv, injectGLMEnvForTeam, buildGLMEnvVars |
| `internal/cli/launcher.go` | 1 function: removeGLMEnv |
| `internal/config/defaults.go` | Add DefaultGLM45 constant |
| `internal/cli/glm_team_test.go` | Add DISABLE_BETAS assertions |
| `internal/cli/glm_model_override_test.go` | Add DISABLE_BETAS assertions |
| `internal/cli/glm_new_test.go` | Add DISABLE_BETAS assertions |

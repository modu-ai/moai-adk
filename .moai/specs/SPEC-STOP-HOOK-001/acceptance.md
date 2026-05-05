---
id: SPEC-STOP-HOOK-001
acceptance_version: "0.1.0"
created_at: 2026-04-30
updated_at: 2026-04-30
author: manager-spec
---

# Acceptance Criteria — SPEC-STOP-HOOK-001

## Given-When-Then Scenarios

### Scenario 1: Stop hook 등록 — settings.json에 entry 존재

**Given** a freshly initialized project via `moai init`
**And** the template includes Stop hook entry

**When** the user inspects `.claude/settings.json`

**Then** the file SHALL contain a `Stop` event entry under `hooks`
**And** the entry SHALL reference `handle-stop.sh` shell wrapper
**And** the entry SHALL have `timeout: 60`

---

### Scenario 2: 정상 테스트 통과 → approve

**Given** a Go project with passing test suite (`go test ./...` exits 0)
**And** `MOAI_STOP_HOOK_DISABLED` is unset

**When** orchestrator turn ends and Stop hook fires

**Then** the hook SHALL execute `go test ./...` and observe exit code 0
**And** the hook SHALL output `{"decision":"approve"}` to stdout
**And** the hook SHALL exit with code 0

---

### Scenario 3: 테스트 실패 → block

**Given** a Go project with one intentionally failing test
**And** `MOAI_STOP_HOOK_DISABLED` is unset

**When** orchestrator turn ends and Stop hook fires

**Then** the hook SHALL execute `go test ./...` and observe non-zero exit
**And** the hook SHALL output `{"decision":"block","reason":"<failure summary>"}`
**And** the failure summary SHALL include the first failing test name (truncated to 200 chars)

---

### Scenario 4: stop_hook_active 가드 — 무한 루프 방지

**Given** a project with passing tests
**And** Claude Code injects `stop_hook_active: true` (recursive call detected)

**When** the Stop hook receives the input

**Then** the hook SHALL detect `stop_hook_active = true`
**And** the hook SHALL output `{"decision":"approve","reason":"stop_hook_active guard"}`
**And** the hook SHALL NOT execute any test command
**And** the hook SHALL exit immediately (< 1 second)

---

### Scenario 5: 환경변수 우회

**Given** a Go project with intentionally failing test
**And** environment variable `MOAI_STOP_HOOK_DISABLED=1`

**When** orchestrator turn ends and Stop hook fires

**Then** the hook SHALL skip test execution
**And** the hook SHALL output `{"decision":"approve","reason":"disabled by env"}`
**And** the failing test SHALL NOT block the turn

---

### Scenario 6: 인식 불가 언어 → silent pass

**Given** a project directory with no recognized language marker (no go.mod, package.json, pyproject.toml, etc.)
**And** `MOAI_STOP_HOOK_DISABLED` is unset

**When** orchestrator turn ends and Stop hook fires

**Then** the hook SHALL detect no language marker
**And** the hook SHALL output `{"decision":"approve","reason":"no test command detected"}`
**And** the hook SHALL exit with code 0

---

### Scenario 7: 60초 timeout

**Given** a Go project with intentionally slow test (sleeps > 60 seconds)
**And** `MOAI_STOP_HOOK_DISABLED` is unset

**When** orchestrator turn ends and Stop hook fires

**Then** the hook SHALL execute `go test ./...`
**And** after 60 seconds the hook SHALL kill the test process
**And** the hook SHALL output `{"decision":"block","reason":"test timeout"}`
**And** partial output SHALL be available in stderr or reason field

---

### Scenario 8: 16-language detection — 각 언어 정상 매핑

**Given** 16 sample projects, each with a different project marker (go.mod, package.json, ..., tsconfig.json)

**When** Stop hook detects the language for each project

**Then** for each of the 16 languages, the hook SHALL select the corresponding test command
**And** all 16 detections SHALL succeed (no "no test command detected" message)
**And** the language map SHALL match research.md §2.4 table

---

## Edge Cases

### EC-1: settings.json corrupt
If `.claude/settings.json` is malformed JSON, the hook SHALL not be triggered (Claude Code skip). The hook itself does not validate settings.json.

### EC-2: shell wrapper missing executable bit
If `handle-stop.sh` exists but lacks executable permission, Claude Code surfaces a hook error. `moai init` and `moai update` SHALL set 0755 permission on hook scripts.

### EC-3: project_dir not provided in hook payload
If Claude Code does not pass `project_dir` in the JSON payload, the hook SHALL fall back to `os.Getenv("CLAUDE_PROJECT_DIR")` or current working directory.

### EC-4: Multiple languages detected (e.g., Go monorepo with frontend)
If multiple project markers exist, the hook SHALL prioritize markers in this order: go.mod > pyproject.toml > package.json > Cargo.toml > others. Documented in language detection logic.

### EC-5: Windows shell wrapper incompatibility
On Windows where bash is unavailable, Claude Code uses the alternative command path. `internal/template/templates/.claude/settings.json` SHALL include Windows-compatible command (e.g., `cmd.exe /c "moai hook stop"`).

### EC-6: hook output exceeds Claude Code buffer
If `reason` field is excessively large (>10KB), the hook SHALL truncate to 5KB with "...truncated" suffix.

---

## Quality Gate Criteria

| Gate | Threshold | Evidence |
|------|-----------|----------|
| Stop hook registered in settings.json | template contains entry | grep test |
| stop_hook_active guard | recursive call simulation | unit test PASS |
| Cross-platform | macOS/Linux/Windows CI | 3/3 PASS |
| 16-language detection | all 16 markers detected | unit test 16/16 |
| Privacy: no user content in reason | log scan | 0 violations |
| Auto-deploy via moai update | template + local sync | clean |
| Timeout enforcement | 60s upper bound | unit test PASS |

---

## Definition of Done

- [ ] All 8 Given-When-Then scenarios PASS
- [ ] All 6 edge cases (EC-1 to EC-6) documented and handled
- [ ] All 7 quality gate criteria meet threshold
- [ ] `internal/hook/stop.go` unit tests with >= 90% coverage
- [ ] `cmd/moai/hook.go` integration tests with stop subcommand
- [ ] `.claude/hooks/moai/handle-stop.sh` exists and executable
- [ ] `internal/template/templates/.claude/settings.json` contains Stop hook entry
- [ ] `make build` regenerates embedded.go without diff outside expected scope
- [ ] CHANGELOG.md updated under Unreleased
- [ ] CI runners (ubuntu/macos/windows) PASS
- [ ] plan-auditor PASS
- [ ] privacy review: log scan finds zero user content
- [ ] backward compat: `MOAI_STOP_HOOK_DISABLED=1` opt-out works

End of acceptance.md (SPEC-STOP-HOOK-001).

---
id: SPEC-STOP-HOOK-001
status: draft
version: "0.1.0"
priority: High
labels: [stop-hook, quality-gate, hook-system, test-automation, wave-3, tier-2]
issue_number: null
scope: [.claude/hooks, internal/hook, internal/template/templates/.claude/settings.json, cmd/moai]
blockedBy: []
dependents: []
created_at: 2026-04-30
updated_at: 2026-04-30
author: manager-spec
wave: 3
tier: 2
---

# SPEC-STOP-HOOK-001: Stop Hook 테스트 강제

## HISTORY

- 2026-04-30 v0.1.0: 최초 작성. Wave 3 / Tier 2. Anthropic "Subagents in Claude Code"의 Stop Hook 패턴을 본 프로젝트에 도입하여 orchestrator turn 종료 시 자동 테스트 차단 메커니즘 추가.

---

## 1. Goal (목적)

Claude Code의 Stop hook event를 통해 orchestrator turn 종료 시점에 프로젝트의 테스트 스위트를 자동 실행하여, 실패 시 turn을 차단하고 통과 시 통과시키는 quality enforcement 메커니즘을 구축한다. 16-language 자동 감지 + `stop_hook_active` infinite loop 가드 + JSON contract 출력을 표준 패턴으로 적용.

### 1.1 배경

- Anthropic blog "Subagents in Claude Code": "Stop Hook (Blocks Claude Turn Until Tests Pass): Checks `stop_hook_active` flag to prevent infinite loops. Runs test suite; returns `decision: block, reason: ...` if tests fail."
- 본 프로젝트의 `/moai gate`는 사용자 능동 호출 의존 → orchestrator turn 자동 차단 부재
- 기존 hook 인프라 (`.claude/hooks/moai/handle-*.sh`)에 `handle-stop.sh`만 부재

### 1.2 비목표 (Non-Goals)

- PreCommit hook 단독 구현 (Stop hook이 우선순위 높음)
- pre-commit 컨텍스트가 아닐 때 lint/format까지 강제 (test만 차단)
- 16-language test runner 신규 구현 (기존 `/moai gate` 로직 재사용)
- 사용자 능동 비활성화 옵션 부재 (`MOAI_STOP_HOOK_DISABLED=1` 제공 필수)
- TeammateIdle hook과 통합 (별도 SPEC)

---

## 2. Scope (범위)

### 2.1 In Scope

- `.claude/hooks/moai/handle-stop.sh` 신규 작성 (shell wrapper)
- `cmd/moai/hook.go`에 `moai hook stop` 서브커맨드 추가
- `internal/hook/stop.go` Stop event handler (Go)
- `internal/template/templates/.claude/settings.json`에 `Stop` event entry 추가
- `MOAI_STOP_HOOK_DISABLED=1` 환경변수 우회 경로 제공
- 16-language 테스트 명령 자동 감지 + 실행
- JSON contract 출력 (`{"decision":"block"|"approve","reason":"..."}`)
- `stop_hook_active` flag 가드 (recursive prevention)

### 2.2 Exclusions (What NOT to Build)

- 16-language test runner 신규 구현 (기존 gate 재사용)
- pre-commit hook 자동 설치 (별도 SPEC)
- Stop hook이 lint/format 실행 (test만)
- TeammateIdle hook 통합
- 실시간 dashboard 또는 metric 시각화
- Python 기반 hook (CLAUDE.local.md §7 정책 위반)

---

## 3. Environment (환경)

- 런타임: moai-adk-go (Go 1.23+)
- Claude Code v2.0+ (Stop hook event 지원)
- 영향 디렉터리: `.claude/hooks/moai/`, `internal/hook/`, `cmd/moai/`, `internal/template/templates/.claude/`
- 16-language project markers: go.mod, package.json, pyproject.toml, Cargo.toml, pom.xml, build.gradle, *.csproj, Gemfile, composer.json, mix.exs, CMakeLists.txt, build.sbt, DESCRIPTION, pubspec.yaml, Package.swift, tsconfig.json

---

## 4. Assumptions (가정)

- A1: Claude Code v2.0+ Stop event spec이 안정적으로 유지된다
- A2: `stop_hook_active` flag는 Claude Code가 자동 설정하며 hook은 read-only 접근
- A3: 기존 `/moai gate` 로직이 16-language test runner를 보유 (재사용 가능)
- A4: shell wrapper + Go binary 구조가 macOS/Linux/Windows 모두 호환
- A5: `MOAI_STOP_HOOK_DISABLED=1`은 비상 우회로만 사용, default는 활성화

---

## 5. Requirements (EARS Format)

### 5.1 Ubiquitous Requirements

- **REQ-SH-001**: THE STOP HOOK SHALL be registered in the project `settings.json` under the `Stop` event with timeout 60 seconds.
- **REQ-SH-002**: THE STOP HOOK SHALL output a structured JSON object with fields `decision` (enum "block" or "approve") and `reason` (string) on stdout.
- **REQ-SH-003**: THE STOP HOOK SHALL detect the project's primary language from project markers and select the corresponding test command.
- **REQ-SH-004**: THE STOP HOOK SHALL preserve `stop_hook_active` flag handling per the Anthropic protocol.

### 5.2 Event-Driven Requirements

- **REQ-SH-005**: WHEN orchestrator turn ends AND `MOAI_STOP_HOOK_DISABLED` is unset, THE STOP HOOK SHALL execute the project's test suite using the auto-detected language test command.
- **REQ-SH-006**: WHEN tests fail (non-zero exit code), THE STOP HOOK SHALL return `{"decision":"block","reason":"<test failure summary>"}` to Claude Code.
- **REQ-SH-007**: WHEN tests pass (zero exit code), THE STOP HOOK SHALL return `{"decision":"approve"}` to Claude Code.
- **REQ-SH-008**: WHEN no recognized language marker is present in the project, THE STOP HOOK SHALL return `{"decision":"approve","reason":"no test command detected"}` (silent pass).

### 5.3 State-Driven Requirements

- **REQ-SH-009**: WHILE `stop_hook_active` flag is set to true, THE STOP HOOK SHALL NOT recurse and SHALL return `{"decision":"approve","reason":"stop_hook_active guard"}`.
- **REQ-SH-010**: WHILE the test command is executing, THE STOP HOOK SHALL enforce a 60-second timeout to prevent indefinite blocking.

### 5.4 Conditional (WHERE / IF) Requirements

- **REQ-SH-011**: WHERE `MOAI_STOP_HOOK_DISABLED=1` is set in the environment, THE STOP HOOK SHALL skip test execution and return `{"decision":"approve","reason":"disabled by env"}`.
- **REQ-SH-012**: IF the test command exits with timeout (>= 60 seconds), THEN THE STOP HOOK SHALL return `{"decision":"block","reason":"test timeout"}` and surface the partial output.
- **REQ-SH-013**: WHERE pre-commit context is detected (via `git diff --cached --name-only` non-empty), THE STOP HOOK MAY include staged file count in the reason field for diagnostic purposes.

### 5.5 Unwanted (Negative) Requirements

- **REQ-SH-014**: THE STOP HOOK SHALL NOT modify any project file (read-only operation).
- **REQ-SH-015**: THE STOP HOOK SHALL NOT execute lint or format commands (test execution only).
- **REQ-SH-016**: THE STOP HOOK SHALL NOT block on missing language markers (silent pass per REQ-SH-008).
- **REQ-SH-017**: THE STOP HOOK SHALL NOT recurse if `stop_hook_active` is true (per REQ-SH-009).

---

## 6. Success Criteria (성공 기준)

| Criterion | Measurement | Target |
|-----------|-------------|--------|
| Stop hook executes on turn end | E2E test (intentional code change + turn end) | invocation logged |
| Test failure blocks turn | seed failing test | `decision: block` returned |
| Test pass approves turn | clean test suite | `decision: approve` returned |
| Recursion prevented | `stop_hook_active=true` injection | hook returns approve, no execution |
| Environment opt-out works | `MOAI_STOP_HOOK_DISABLED=1` | hook skips test |
| Cross-platform | macOS/Linux/Windows CI | 3/3 PASS |
| 16-language detection | each language's project marker test | 16/16 detected |

---

## 7. Acceptance References

See `acceptance.md` for Given-When-Then scenarios and Definition of Done.

---

## 8. Constraints

- C1: Python-based hook 금지 (CLAUDE.local.md §7 — Shell scripts only)
- C2: Stop hook은 lint/format 실행 금지 (test만, scope 최소화)
- C3: 모든 변경은 Template-First Rule 준수 (`internal/template/templates/` 동기화)
- C4: shell wrapper는 bash 헤더 + Windows에서는 Go binary 직접 호출 fallback
- C5: timeout default 60s, 사용자가 더 길게 설정 가능 (settings.json `timeout` field)

End of spec.md (SPEC-STOP-HOOK-001 v0.1.0).

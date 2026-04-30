# Research — SPEC-STOP-HOOK-001 (Stop Hook 테스트 강제)

**SPEC**: SPEC-STOP-HOOK-001
**Wave**: 3 / Tier 2 (검증 통과 — 표준 Claude Code Stop hook 활용)
**Created**: 2026-04-30
**Author**: manager-spec

---

## 1. 출처 (Anthropic 공식 자료)

### 1.1 Verbatim 인용

Anthropic blog "Subagents in Claude Code":

> "Stop Hook (Blocks Claude Turn Until Tests Pass): Checks `stop_hook_active` flag to prevent infinite loops. Runs test suite; returns `{\"decision\": \"block\", \"reason\": \"...\"}` if tests fail."

> "Hooks can interrupt agent turns when the project's quality bar is not met. The stop_hook_active flag prevents the hook from triggering itself recursively."

### 1.2 검증 (claude-code-guide 검증 완료)

claude-code-guide 에이전트가 Claude Code 공식 hook event 스펙 기준으로 본 권고를 검증함. 결론:

- **호환성**: ✅ 완전 지원 — `Stop` hook event는 Claude Code v2.0+에서 official spec
- **JSON contract**: `decision: "block" | "approve"`, `reason: <string>`, `stop_hook_active` 필드 모두 표준
- **권고 채택**: ACCEPT — 단, 본 프로젝트의 16-language 정책에 따라 언어 자동 감지 layer 필수

---

## 2. 현재 상태 (As-Is)

### 2.1 기존 hook 인프라 위치

`.claude/hooks/moai/`:

```
handle-session-start.sh
handle-pre-tool.sh
handle-post-tool.sh
handle-user-prompt.sh
handle-subagent-stop.sh
handle-pre-compact.sh
```

**관찰**: `handle-stop.sh`는 부재 또는 빈 stub. 또한 PreCommit/Pre-Compact hook은 있으나 Stop hook은 quality enforcement 역할 미연결.

### 2.2 settings.json 현재 hook 등록 상태

`internal/template/templates/.claude/settings.json` (예상 구조):

```json
{
  "hooks": {
    "SessionStart": [{ "hooks": [...] }],
    "PreToolUse":   [{ "hooks": [...] }],
    "PostToolUse":  [{ "hooks": [...] }],
    "UserPromptSubmit": [{ "hooks": [...] }],
    "SubagentStop": [{ "hooks": [...] }]
  }
}
```

**관찰**: `Stop` event entry 자체가 부재. Stop hook 추가 시 template + local 두 곳 동기화 필수.

### 2.3 /moai gate 명령

현재 `moai gate` (CLI 또는 슬래시) 호출 시에만 lint+format+type-check+test 실행. 즉 **사용자 능동 호출 의존** — orchestrator turn 종료 시 자동 차단 불가.

### 2.4 16-language 자동 감지 인프라

`internal/lang/detect.go` (있을 시) 또는 `pkg/projectlang/`에 project marker 기반 언어 감지 로직 존재.

| Project marker | Language | Test command |
|---------------|----------|--------------|
| `go.mod` | Go | `go test ./...` |
| `package.json` | Node | `npm test` (or yarn/pnpm) |
| `pyproject.toml` / `setup.py` | Python | `pytest` |
| `Cargo.toml` | Rust | `cargo test` |
| `pom.xml` / `build.gradle` | Java/Kotlin | `mvn test` / `./gradlew test` |
| `*.csproj` | C# | `dotnet test` |
| `Gemfile` | Ruby | `bundle exec rspec` |
| `composer.json` | PHP | `vendor/bin/phpunit` |
| `mix.exs` | Elixir | `mix test` |
| `CMakeLists.txt` | C/C++ | `ctest` |
| `build.sbt` | Scala | `sbt test` |
| `DESCRIPTION` (R) | R | `R CMD check` |
| `pubspec.yaml` | Flutter | `flutter test` |
| `Package.swift` | Swift | `swift test` |
| `*.csproj` (.NET) | C# | `dotnet test` |
| `tsconfig.json` | TypeScript | (package.json 우선) |

---

## 3. 격차 분석 (Gap Analysis)

| 영역 | 현재 (As-Is) | 목표 (To-Be) | 격차 |
|------|-------------|-------------|------|
| Stop hook 등록 | 미등록 | settings.json + handle-stop.sh 등록 | 신규 추가 |
| Stop hook handler | 없음 | `moai hook stop` 서브커맨드 | 신규 구현 (Go) |
| stop_hook_active 가드 | 없음 | infinite loop 방지 | 표준 패턴 적용 |
| 언어 자동 감지 → test 명령 | gate 명령 내 존재 (재사용 가능) | hook handler에서 호출 | 재사용 |
| pre-commit 컨텍스트 감지 | 없음 | git index 변경 감지 | 신규 |
| JSON contract 출력 | gate는 stderr | hook은 `{"decision":"block"|"approve","reason":...}` | 신규 형식 |

---

## 4. 코드베이스 분석 (Affected Files)

### 4.1 Primary 수정 대상

| 파일 | 수정 유형 | 변경 사유 |
|------|----------|----------|
| `internal/hook/stop.go` | 신규 | Stop hook handler (Go) |
| `cmd/moai/hook.go` | 수정 | `moai hook stop` 서브커맨드 라우팅 |
| `.claude/hooks/moai/handle-stop.sh` | 신규 | Shell wrapper 호출 `moai hook stop` |
| `.claude/settings.json` | 수정 | `Stop` event entry 추가 |
| `internal/template/templates/.claude/settings.json` | 수정 | Template-First 동기화 |
| `internal/template/templates/.claude/hooks/moai/handle-stop.sh` | 신규 | Template wrapper |

### 4.2 Secondary 영향

- `internal/lang/detect.go` 또는 `pkg/projectlang/`: 재사용
- `internal/gate/runner.go` (있을 시): 테스트 실행 로직 재사용
- `pkg/version/`: hook handler 버전 표시 (디버그용)

### 4.3 Templates (Template-First Rule 준수)

- 모든 `.claude/` 변경은 `internal/template/templates/.claude/`에 동기화
- `make build` 실행 후 `internal/template/embedded.go` 재생성

---

## 5. 위험 및 가정

### 5.1 Risks

| Risk | Likelihood | Impact | Mitigation |
|------|-----------|--------|-----------|
| Stop hook이 매 turn 차단 → DX 저하 | High | High | 빠른 fail-fast (timeout 30s 이내), opt-out flag (`MOAI_STOP_HOOK_DISABLED=1`) 제공 |
| Infinite loop (stop_hook 자기 호출) | Medium | Critical | `stop_hook_active` 표준 가드 strict 검증 |
| 언어 감지 실패 → 차단 | Low | Medium | unknown 언어는 silent pass (fallback approve) |
| 테스트 실행 시간 길어 timeout | Medium | High | `timeout: 60`초 + skip 메커니즘 (`.moai/config/sections/quality.yaml.stop_hook.test_timeout`) |
| pre-commit context 오감지 | Medium | Medium | `git rev-parse --is-inside-work-tree` + `git diff --cached --name-only` 명시 검증 |
| Windows shell 비호환 | Medium | High | shell wrapper는 bash 헤더 + cmd PATH는 Go binary 직접 호출로 fallback |

### 5.2 Assumptions

- A1: `moai gate` 명령이 이미 16-language 테스트 실행 로직을 보유 (재사용 가능)
- A2: Claude Code v2.0+ Stop event spec이 안정적으로 유지됨
- A3: 사용자가 `MOAI_STOP_HOOK_DISABLED=1` 환경변수로 비상 우회 가능
- A4: `stop_hook_active` flag는 Claude Code가 자동 설정하며 hook은 read-only

---

## 6. 측정 계획 (Baseline + Validation)

| Metric | Baseline 측정 방법 | 목표 |
|--------|-------------------|------|
| Stop hook 실행 시간 | go.mod 프로젝트에서 `go test -short ./...` | < 30s |
| 차단 정확도 | 의도적 실패 테스트 시드 5개 | 100% block |
| 통과 정확도 | 정상 프로젝트 5개 | 100% approve |
| Infinite loop 방지 | `stop_hook_active=true` 강제 주입 | 0회 재진입 |
| Cross-platform | macOS/Linux/Windows | 3/3 PASS |

---

## 7. 대안 검토 (Alternatives Considered)

| 대안 | 채택 여부 | 이유 |
|------|----------|------|
| PostToolUse hook에서 매 도구 호출 후 테스트 | ❌ | 토큰/시간 비용 폭증, Anthropic 권고는 Stop event |
| pre-commit hook 단독 (git only) | ❌ | Claude Code의 turn 단위 차단 불가능 |
| SubagentStop hook 활용 | ❌ | subagent 종료 시점이지 orchestrator turn 종료 아님 |
| Python 기반 hook | ❌ | 본 프로젝트 정책 (CLAUDE.local.md §7) — Shell scripts only |
| Go binary 직접 호출 (shell wrapper 없이) | ❌ | Windows 호환성 제약, shell wrapper가 안전 |

---

## 8. 참고 SPEC

- SPEC-HOOK-001 ~ SPEC-HOOK-009: 기존 hook 시스템 기반 SPEC들
- SPEC-GATE-001: `/moai gate` 명령 정의 (재사용 대상)
- SPEC-LSP-QGATE-004: LSP 기반 quality gate (Stop hook 보완 후보)

---

## 9. Open Questions (Plan 단계 해결 대상)

- OQ1: Stop hook이 차단할 때 Claude Code가 자동 retry하는가, 사용자 개입 대기인가? → plan.md에서 결정
- OQ2: pre-commit context 자동 감지 vs 명시 환경변수 (`MOAI_PRE_COMMIT_MODE=1`) 중 우선순위?
- OQ3: `MOAI_STOP_HOOK_DISABLED=1` 외 SPEC-level opt-out (frontmatter) 필요?
- OQ4: 테스트 실패 시 reason 메시지에 실패한 테스트 이름 포함 → 프라이버시/길이 트레이드오프

---

End of research.md (SPEC-STOP-HOOK-001).

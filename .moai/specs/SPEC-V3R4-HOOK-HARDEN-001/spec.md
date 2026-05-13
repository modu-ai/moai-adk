---
id: SPEC-V3R4-HOOK-HARDEN-001
version: "0.1.0"
status: implemented
created_at: 2026-05-13
updated_at: 2026-05-13
author: GOOS행님
priority: High
labels: [hooks, observability, reliability, claude-code, template]
issue_number: null
depends_on: []
related_specs: [SPEC-CC2122-HOOK-001, SPEC-CC2122-HOOK-002]
title: "Hook Wrapper Hardening"
created: 2026-05-13
updated: 2026-05-13
phase: "v3.0.0 - Lifecycle"
module: "hooks"
lifecycle: completed
tags: "legacy"
---

# SPEC-V3R4-HOOK-HARDEN-001: Hook Wrapper Hardening

## HISTORY

| Version | Date | Author | Description |
|---------|------|--------|-------------|
| 0.1.0 | 2026-05-13 | GOOS행님 | Initial draft. Plan-phase artifact. Triggered by diagnostic session (2026-05-13) that surfaced 22 `PreToolUse:Bash` hook cancellations within a 7-day Claude Code session window where direct `moai hook pre-tool` execution measured ~8ms but the wrapper-mediated call hit the 5-second timeout (D1). Five fixes packaged across three waves: visibility (stderr preservation, hardcoded-path removal), mechanism (timeout uplift, stdin pipe shortening), and verification (load-simulated reproduction test). |

## Overview

본 SPEC은 Claude Code의 hook wrapper layer를 **진단 가능하고, 시간 여유 있고, 환경 중립적으로** 강화한다. 진단 세션(2026-05-13)에서 다음 5가지 문제가 evidence-backed로 확인되었다:

| ID | 문제 | 영향 범위 | 심각도 |
|----|------|-----------|--------|
| D1 | `PreToolUse:Bash` hook 22건이 5초 timeout 경계 직후 cancellation (durationMs 5068~5278ms) — 직접 실행은 ~8ms | 사용자 세션 차단 + tool call 중단 | HIGH |
| D2 | 30개 wrapper 모두 `2>/dev/null`로 stderr silence → timeout 발생 시 진단 정보 0건 | 모든 hook 환경 | HIGH |
| D3 | `/Users/goos/go/bin/moai` 사용자 고정 절대 경로가 30개 template wrapper에 박힘 → 다른 사용자 환경에서 fallback chain 의존 | 16개 언어 / 모든 사용자 | MEDIUM |
| D4 | `settings.json` PreToolUse hook timeout이 5초로 설정 — 부하 시 tight | PreToolUse 4개 hook (Bash/Edit/Write/MultiEdit) | MEDIUM |
| D5 | `mktemp` + `cat > temp` + `exec moai hook X < temp` 패턴이 fork chain + temp file IO를 추가 (~수십 ms) | 30 wrapper | MEDIUM |

본 SPEC의 목표는:
1. **Observability 회복** (REQ-1, REQ-2): stderr를 보존하여 timeout 원인 분석 가능
2. **Environment Neutrality 회복** (REQ-3, REQ-9): 하드코딩 사용자 경로 제거, 16-language neutrality 준수
3. **Timeout Margin 확보** (REQ-5): PreToolUse 5→10초로 늘려 부하 여유 확보
4. **Execution Mechanism 단순화** (REQ-4, REQ-7): temp file 우회 없이 direct pipe 사용, fork chain 감소
5. **Reproducibility 보장** (REQ-6): 50KB+ tool_input 시뮬레이션 회귀 테스트 추가

본 SPEC은 hook 자체의 동작(`moai hook X` Go 코드)은 변경하지 않는다. 변경 영향 범위는 (a) 30개 hook wrapper script, (b) 28개 template wrapper script, (c) `settings.json` 1개 + template, (d) 신규 reproduction test 1개로 한정된다.

## Background

본 SPEC의 의사결정 근거 체인:

1. **2026-05-13 진단 세션** — `~/.claude/projects/-Users-goos-MoAI-moai-adk-go/*.jsonl` 최근 7일 분석 결과 `attachment.type == "hook_cancelled"` AND `hookName == "PreToolUse:Bash"` 22건. 모두 durationMs 5068~5278ms 범위 → 5초 timeout 경계 직후. 직접 호출 시 8ms이므로 root cause는 wrapper layer.
2. **`internal/template/templates/.claude/hooks/moai/handle-pre-tool.sh.tmpl`** — 모든 wrapper가 `mktemp` → `head -c 65536 > temp` → `exec moai hook X < temp 2>/dev/null` 패턴 사용. stderr silence + fork chain.
3. **`.claude/hooks/moai/handle-pre-tool.sh`** (local copy) — `/Users/goos/go/bin/moai` 하드코딩 발견. CLAUDE.local.md §14 HARD rule (`$HOME` 사용 필수) + §15 16-language neutrality 위반.
4. **`.claude/settings.json`** — PreToolUse timeout `5`초, matcher `Write|Edit|Bash`. PostToolUse는 `async: true` + 10초로 안전 영역.
5. **CLAUDE.local.md §14** — "Go 코드/template에서 사용자 절대 경로 하드코딩 금지". 이미 명문화된 룰을 wrapper template이 위반한 상태.
6. **`pkg/version/version.go`** 및 `moai version` 검증 — 직접 `moai hook pre-tool < sample.json`을 실행하면 8-15ms 응답. wrapper만 5초+ 소요 → root cause는 stdin handoff 메커니즘.
7. **SPEC-CC2122-HOOK-001 / -002** — 27-event hook system은 이미 활성화. 본 SPEC은 그 기반 위에서 wrapper 신뢰성만 개선.

선례 참조:

- **SPEC-V3R4-CATALOG-001/002 패턴**: D7 lock (특정 파일 미수정) + sentinel-driven test 구조 차용. 본 SPEC의 D-lock: **`moai hook X` Go 핸들러 코드 미수정** — 모든 변경은 wrapper layer / settings.json / template로 한정.
- **CLAUDE.local.md §14 / §15**: 하드코딩 제거 + 16-language neutrality 룰을 wrapper template에 강제 적용.
- **session-handoff.md threshold revision**: 운영 이슈 evidence-backed로 정량 데이터(durationMs 분포) 활용한 fix 패턴 차용.

## EARS Requirements

### 1. Observability — Stderr Preservation (Ubiquitous)

REQ-HOOK-HARDEN-001-001: The system shall preserve stderr output from every hook wrapper script. All wrappers in `.claude/hooks/moai/handle-*.sh` and `internal/template/templates/.claude/hooks/moai/handle-*.sh.tmpl` shall replace the `2>/dev/null` suffix on the `exec moai hook X ...` invocation with `2>>"$MOAI_HOOK_STDERR_LOG"` where `MOAI_HOOK_STDERR_LOG` defaults to `$HOME/.moai/logs/hook-stderr.log` if unset.

REQ-HOOK-HARDEN-001-002: The system shall ensure the stderr log directory exists before redirecting. Each wrapper shall include the line `mkdir -p "$(dirname "$MOAI_HOOK_STDERR_LOG")" 2>/dev/null || true` immediately before the first `exec moai hook X ...` candidate, with the directory creation failure non-blocking (the `|| true` clause).

REQ-HOOK-HARDEN-001-003: When a hook execution duration exceeds 80% of the configured timeout (3 seconds for 5-second timeouts, 8 seconds for 10-second timeouts), the wrapper shall emit a single-line warning `[hook-warn] <hookName> elapsed=<ms>ms threshold=<budget_ms>ms` to stderr (which is now captured per REQ-001). Implementation note: the warning is emitted by `moai hook X` Go code, NOT by the bash wrapper — this SPEC does NOT modify Go handler code, so REQ-003 is deferred to a post-wave follow-up SPEC and marked OPT in §7 Definition of Done.

### 2. Environment Neutrality — Hardcoded Path Removal (Ubiquitous)

REQ-HOOK-HARDEN-001-004: The system shall not embed user-absolute paths (e.g., `/Users/<name>/go/bin/moai`) in any hook wrapper template under `internal/template/templates/.claude/hooks/moai/`. The four-step fallback chain shall consist of: (a) `command -v moai`, (b) `$HOME/go/bin/moai`, (c) `$HOME/.local/bin/moai`, (d) silent exit. The "detected Go bin path" branch using a hardcoded absolute path is REMOVED.

REQ-HOOK-HARDEN-001-005: When `moai init` is run on a fresh machine, the generated wrapper scripts in `.claude/hooks/moai/*.sh` shall contain no occurrence of `/Users/`, `/home/<specific-name>`, or `C:\Users\<specific-name>` literal strings. The rendered output shall only contain environment variable references (`$HOME`, `$CLAUDE_PROJECT_DIR`) or the `{{posixPath .GoBinPath}}` template variable (which expands to either `$HOME/go/bin` or an explicitly user-supplied path at init time).

### 3. Timeout Margin (Event-Driven)

REQ-HOOK-HARDEN-001-006: When the `settings.json` template (`internal/template/templates/.claude/settings.json` or any `.tmpl` variant) is rendered, the `PreToolUse` hook's `timeout` field shall be set to `10` (seconds), up from the previous value of `5`. The existing 5-second timeout for `Notification`, `PostToolUse-failure`, and similar non-blocking events shall remain unchanged.

REQ-HOOK-HARDEN-001-007: When `moai update -t` (template-only sync) is invoked on a project whose `.claude/settings.json` `PreToolUse.hooks.timeout` is `5`, the update command shall NOT silently overwrite the local timeout — instead, a warning shall be emitted naming the file and the recommended new value (`10`). User decides whether to apply. This preserves §2 Protected Directories convention of CLAUDE.local.md (settings.json is rendered, but established projects may have intentional overrides).

### 4. Execution Mechanism — Direct Pipe (Ubiquitous)

REQ-HOOK-HARDEN-001-008: The system shall not use a temporary file (`mktemp`) as an intermediate buffer between Claude Code stdin and `moai hook X` stdin in the wrapper script. Each wrapper shall use direct pipe semantics: Claude Code stdin → wrapper stdin → `exec moai hook X` (which reads stdin directly), eliminating the `cat > "$temp_file"` and trap-cleanup blocks.

REQ-HOOK-HARDEN-001-009: When the `head -c 65536 > "$temp_file"` size-limit pattern is present in a wrapper (currently in `handle-pre-tool.sh.tmpl` only), it shall be replaced by an in-line `head -c 65536` filter that pipes the truncated content directly into `exec moai hook X`. If the 64KB ceiling is exceeded, the wrapper shall emit a `block` decision JSON to stdout and exit 0, matching the existing template's escape hatch (line 12 of the current template).

### 5. Template-Local Sync (Ubiquitous)

REQ-HOOK-HARDEN-001-010: The system shall maintain template-local synchronization. Every file under `.claude/hooks/moai/handle-*.sh` shall have a corresponding `.tmpl` file under `internal/template/templates/.claude/hooks/moai/handle-*.sh.tmpl`. Where the local directory contains files (32) that exceed the template directory (28), the missing 4 wrapper templates shall be added to the template directory during this SPEC OR documented as local-only in `.gitignore` with rationale.

### 6. Reproducibility — Regression Test (Event-Driven)

REQ-HOOK-HARDEN-001-011: When the test `TestHookWrapper_LargeStdin_DoesNotExceedTimeout` is executed (new file `internal/cli/hook_wrapper_load_test.go` or similar), the test shall: (a) generate a 50KB+ JSON payload simulating worst-case `tool_input`, (b) invoke `handle-pre-tool.sh` with the payload on stdin via `exec.Command`, (c) measure end-to-end wall-clock duration, (d) assert the duration is below 1000ms (1 second — the post-hardening budget headroom). Test must be `t.Parallel()`-safe and use `t.TempDir()` per CLAUDE.local.md §6.

REQ-HOOK-HARDEN-001-012: When the test runs in CI (any of ubuntu-latest, macos-latest, windows-latest), it shall account for cold-start overhead by performing one warm-up invocation before measurement. The assertion shall use the second invocation's duration. Cross-platform: on Windows, the test shall skip via `t.Skip()` if `bash` is not available (Git Bash standard on GitHub Actions).

### 7. Log Rotation Boundary (Conditional)

REQ-HOOK-HARDEN-001-013: If the stderr log file (`$HOME/.moai/logs/hook-stderr.log`) exceeds 10MB in size, then on the next wrapper invocation the wrapper shall atomically rename it to `hook-stderr.log.1` (overwriting any previous `.1`) before redirecting new output. The check uses `stat -f%z` (macOS) or `stat -c%s` (Linux) with graceful fallback to skipping rotation if neither is available. No external `logrotate` dependency.

REQ-HOOK-HARDEN-001-014: When the rotated `hook-stderr.log.1` file already exists prior to a new rotation, the wrapper shall overwrite it (single-level rotation only). This avoids accumulating unbounded historical logs while preserving the most recent ~10MB-20MB of error context for diagnosis.

### 8. Backward Compatibility (Unwanted Behavior)

REQ-HOOK-HARDEN-001-015: If a user has manually set `MOAI_HOOK_STDERR_LOG=/dev/null` in their shell environment, then wrappers shall honor that override and produce zero stderr persistence, restoring the pre-hardening silence behavior. This is the opt-out path for users on read-only or memory-only filesystems.

REQ-HOOK-HARDEN-001-016: If a user environment lacks `$HOME` (rare, but possible in some CI / container scenarios), the wrapper shall fall back to writing to `/tmp/moai-hook-stderr.log` AND exit 0 silently if even that fails. No hook invocation shall fail solely because the stderr destination is unavailable.

## Constraints

### Hard Constraints

[HARD] **D-Lock — No Go handler modifications**: This SPEC shall NOT modify any file under `internal/hook/`, `internal/cli/hook*.go`, or `cmd/moai/hook.go`. All changes are confined to: bash wrapper scripts, settings.json (one template + one local file), and new test file. Justification: hook business logic is out of scope; only the wrapper layer is the verified root cause of D1.

[HARD] **D-Lock — No new dependencies**: Wrappers shall remain pure POSIX-ish bash (compatible with `/bin/bash` on macOS 11+ and Ubuntu 22.04+ Git Bash on Windows). No external tools like `logrotate`, `flock`, `jq` are introduced into wrapper hot path. Existing `head`, `stat`, `mkdir`, `mv` suffice.

[HARD] **Template-First Discipline**: Per CLAUDE.local.md §2 [HARD], all wrapper changes must originate in `internal/template/templates/.claude/hooks/moai/*.sh.tmpl`, then propagated to local `.claude/hooks/moai/*.sh` via `make build` + `moai update -t`. Direct local edits without template sync are prohibited.

[HARD] **16-language neutrality**: Per CLAUDE.local.md §15, no language-specific hardcoding may be introduced. The wrapper template applies uniformly to all 16 supported languages.

### Soft Constraints

- Wrapper file size shall remain under 1.5KB per wrapper (current range 750-1465 bytes; rotation logic adds ~200 bytes worst case).
- No introduction of subshell-spawning constructs beyond what is already present (e.g., avoid `$(...)` in hot path where `${var}` suffices).
- Reproducibility test (REQ-011) shall not depend on network or `moai` binary being on `PATH` — uses `exec.Command` with explicit binary lookup.

## Trace

REQ-HOOK-HARDEN-001-001 → AC-HOOK-HARDEN-001-01.a (Wave 1, stderr preservation)
REQ-HOOK-HARDEN-001-002 → AC-HOOK-HARDEN-001-01.a.i (Wave 1, mkdir guard)
REQ-HOOK-HARDEN-001-003 → AC-HOOK-HARDEN-001-03.c (Wave 3, deferred to follow-up, opt-in)
REQ-HOOK-HARDEN-001-004 → AC-HOOK-HARDEN-001-01.b (Wave 1, hardcode removal)
REQ-HOOK-HARDEN-001-005 → AC-HOOK-HARDEN-001-01.b.i (Wave 1, render audit)
REQ-HOOK-HARDEN-001-006 → AC-HOOK-HARDEN-001-02.a (Wave 2, settings.json uplift)
REQ-HOOK-HARDEN-001-007 → AC-HOOK-HARDEN-001-02.a.i (Wave 2, update warn)
REQ-HOOK-HARDEN-001-008 → AC-HOOK-HARDEN-001-02.b (Wave 2, direct pipe)
REQ-HOOK-HARDEN-001-009 → AC-HOOK-HARDEN-001-02.b.i (Wave 2, 64KB limit refactor)
REQ-HOOK-HARDEN-001-010 → AC-HOOK-HARDEN-001-01.c (Wave 1, template-local sync)
REQ-HOOK-HARDEN-001-011 → AC-HOOK-HARDEN-001-03.a (Wave 3, reproduction test)
REQ-HOOK-HARDEN-001-012 → AC-HOOK-HARDEN-001-03.a.i (Wave 3, CI compat)
REQ-HOOK-HARDEN-001-013 → AC-HOOK-HARDEN-001-02.c (Wave 2, log rotation)
REQ-HOOK-HARDEN-001-014 → AC-HOOK-HARDEN-001-02.c.i (Wave 2, rotation idempotency)
REQ-HOOK-HARDEN-001-015 → AC-HOOK-HARDEN-001-03.b (Wave 3, opt-out)
REQ-HOOK-HARDEN-001-016 → AC-HOOK-HARDEN-001-03.b.i (Wave 3, $HOME absence)

### Out of Scope (What NOT to Build)

- See enumerated exclusions below.

1. **Go hook handler logic changes**: `internal/hook/*.go` 와 `internal/cli/hook*.go`의 비즈니스 로직, JSON 파싱, decision 생성 코드는 변경하지 않는다. D1의 root cause는 wrapper layer로 격리되었으므로 Go 코드 수정 없이 해결 가능.
2. **Hook configuration UI / TUI**: settings.json 편집을 위한 대화형 도구나 `moai hook config` 류 CLI 추가는 본 SPEC 범위 밖. 수동 편집 + `moai update` warning(REQ-007)만 제공.
3. **Centralized log aggregation**: `$HOME/.moai/logs/hook-stderr.log`를 외부 시스템(syslog, journald, ELK)으로 전송하는 기능은 도입하지 않는다. 로컬 파일 + 단순 rotation으로 한정.
4. **Performance profiling instrumentation**: pprof, OpenTelemetry, trace export 등 hook 실행 프로파일링은 본 SPEC에서 다루지 않는다 (CLAUDE.local.md §2 [WARN] OTEL test 충돌 회피).
5. **Hook timeout dynamic adjustment**: 시스템 부하나 매트릭 기반 자동 timeout 조정 메커니즘은 도입하지 않는다. settings.json의 정적 값(10초)만 사용.
6. **Cross-OS wrapper variants**: Windows-native PowerShell wrapper나 Cmd.exe wrapper는 추가하지 않는다. Git Bash + bash 호환 wrapper 단일 패스 유지 (Claude Code 공식 권장).
7. **Hook registration UI**: 새로운 hook event 등록/해제 UX는 본 SPEC 범위 밖. 27-event 시스템(SPEC-CC2122-HOOK-001/-002)이 이미 lock-in되어 있음.
8. **`moai hook` Go code refactor**: 본 SPEC은 wrapper layer만 강화. Go handler가 8ms 응답하는 것은 이미 충분히 빠름.
9. **Hot reload of hook scripts**: 현재 Claude Code의 hot-reload 미지원 상태(CLAUDE.local.md Hard Constraints) 유지. wrapper 변경 후 Claude Code 재시작 필요.

## Definition of Done

- [ ] Wave 1 — Visibility: stderr preservation (REQ-001/002), hardcode removal (REQ-004/005), template sync (REQ-010) merged
- [ ] Wave 2 — Mechanism: timeout uplift (REQ-006), update warn (REQ-007), direct pipe (REQ-008/009), log rotation (REQ-013/014) merged
- [ ] Wave 3 — Verification: reproduction test (REQ-011/012), opt-out paths (REQ-015/016) merged
- [ ] (OPT) REQ-003 deferred follow-up SPEC ticketed if needed
- [ ] All 28 hook templates synced under `internal/template/templates/.claude/hooks/moai/`
- [ ] `make build` + `make test` GREEN on macOS + Linux CI lanes (Windows test allowed to skip per REQ-012)
- [ ] CHANGELOG.md Unreleased section updated with one-line summary
- [ ] CLAUDE.local.md §14 / §15 compliance verified via render audit script

## Glossary

- **Hook wrapper**: A bash script in `.claude/hooks/moai/handle-*.sh` that Claude Code invokes for each hook event. Forwards stdin JSON to the `moai hook X` Go subcommand.
- **PreToolUse hook**: Claude Code event fired before any tool invocation (Read, Write, Edit, Bash, etc.). Configured with `matcher` regex in settings.json.
- **Hook cancellation**: `attachment.type == "hook_cancelled"` in session JSONL when Claude Code's per-hook `timeout` elapses before the wrapper exits.
- **Template-First**: CLAUDE.local.md §2 [HARD] discipline — all changes originate in `internal/template/templates/`, then `make build` propagates to embedded FS.
- **D-Lock**: Per-SPEC self-imposed constraint preventing modification of specific files/directories (pattern borrowed from SPEC-V3R4-CATALOG-001 D7 lock).

## References

- Session evidence: `~/.claude/projects/-Users-goos-MoAI-moai-adk-go/*.jsonl` (7-day window through 2026-05-13)
- Hook wrapper template: `internal/template/templates/.claude/hooks/moai/handle-pre-tool.sh.tmpl`
- settings.json template: `internal/template/templates/.claude/settings.json`
- CLAUDE.local.md §2 (Template-First), §6 (Test Isolation), §14 (Hardcode Ban), §15 (16-Language Neutrality)
- Related: SPEC-CC2122-HOOK-001 (27-event system), SPEC-CC2122-HOOK-002 (hook registration)
- Pattern reference: SPEC-V3R4-CATALOG-001 (D7 lock + sentinel-driven audit pattern)

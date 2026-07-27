# SPEC-SEC-GUARDIAN-001 — Acceptance Criteria

## §A Given-When-Then Scenarios

### Scenario 1 — Layer 1 flags a dangerous pattern on Edit (advisory, non-blocking)
- **Given** the guardian is installed and a source file is being edited inside a MoAI session
- **When** an Edit/Write introduces a known-dangerous pattern (e.g. `yaml.load(untrusted)` in Python, or `element.innerHTML = userInput` in JS)
- **Then** the Layer-1 PostToolUse hook forwards the written content to the Go `security-scan` handler, which returns an advisory finding naming the vulnerability class
- **And** the Edit/Write itself completes normally — the guardian NEVER blocks or reverts it (REQ-SG-014)
- **And** the scan runs in-process (regex, no LLM, no per-pattern subprocess) within the 5s hook budget (REQ-SG-013, -015)

### Scenario 2 — Layer 1 is language-neutral across the 16 languages
- **Given** the single-source pattern table organized by vulnerability class
- **When** the same dangerous idiom appears in different languages (a hardcoded secret in Go, Python, TypeScript, Rust, Java, …)
- **Then** the finding fires equally in each language — no language is treated as PRIMARY, and no class is single-language-only unless the idiom is genuinely language-specific (REQ-SG-011)

### Scenario 3 — Layer 2 reviews the turn's diff at Stop (advisory default)
- **Given** the assistant has made several edits during a turn
- **When** the turn finishes (Stop hook fires) and no escalation flag is set
- **Then** the Layer-2 handler runs the Layer-1 pattern engine over the turn's working-tree diff and surfaces any high-severity finding via `systemMessage` (REQ-SG-020, -021, -023)
- **And** no language model is invoked on the default path; the LLM/blocking escalation occurs only when the opt-in env flag is set (REQ-SG-022)

### Scenario 4 — Layer 3 is dormant by default; opt-in cross-file review when enabled
- **Given** the guardian is installed but `MOAI_SECURITY_COMMIT_REVIEW` is NOT set
- **When** a git commit is made in the session
- **Then** the Layer-3 hook self-gates to a silent no-op (exit 0, empty) — no per-commit cost is paid (REQ-SG-032)
- **And** **When** `MOAI_SECURITY_COMMIT_REVIEW` IS set and a commit lands, the Layer-3 review reads the commit's changed files + related files to trace cross-file data flow (IDOR / auth-bypass / cross-file SSRF), surfacing advisory findings (REQ-SG-030, -031, -033)

### Scenario 5 — A block is translated by the orchestrator, never by the hook
- **Given** a guardian layer is running with an opt-in blocking flag set and finds a critical issue
- **When** the hook emits its block signal (structured JSON on the exit-0 stdout channel per the event schema)
- **Then** the orchestrator parses the JSON and runs an `AskUserQuestion` round (accept / override with `--skip-hook` / abort) (REQ-SG-041)
- **And** the hook itself NEVER invokes `AskUserQuestion` or emits a free-form prose question (REQ-SG-040); a `--skip-hook` bypass is appended to `.moai/logs/hook-skip.log` (REQ-SG-043)

### Scenario 6 — Fail-open on missing dependencies
- **Given** the `moai` binary is unresolvable in all three tiers (or `jq`/`git` is absent)
- **When** any guardian hook fires
- **Then** the hook degrades to a silent no-op (exit 0) and never crashes, blocks, or breaks the session (REQ-SG-060)

## §B AC-to-REQ matrix

| AC | Severity | REQ | Verification command (evidence) |
|----|----------|-----|---------------------------------|
| AC-SG-001 | MUST | REQ-SG-010 | `go test -run TestLayer1Scan ./internal/hook/security/` → finding on dangerous fixture |
| AC-SG-002 | MUST | REQ-SG-011 | `go test -run TestPatternsLanguageNeutral ./internal/hook/security/` → same class fires across ≥16 lang fixtures; no PRIMARY |
| AC-SG-003 | MUST | REQ-SG-012 | `go test -run TestPatternClassCoverage ./internal/hook/security/` → all enumerated classes present; pattern count within bounded range 20 ≤ N ≤ 30 (asserted as a range, not a soft "≈25") |
| AC-SG-004 | MUST | REQ-SG-013 | `go test -run TestLayer1NoLLMOrSubprocess ./internal/hook/security/` (asserts the scan execution path imports no `os/exec` / `net/http` and makes no model/tool call) + scoped grep `grep -nE 'exec\.Command\|os/exec\|net/http' internal/hook/security/scan.go internal/hook/security/diff.go \| grep -v "^[^:]*:[0-9]*:[ \t]*//"` → 0 (the dangerous-pattern string literals live in `patterns.go`, which is excluded; `Agent(`/`Skill(` escalation doc-comments are comment-excluded — so the AC no longer self-collides with the package that legitimately names these tokens) |
| AC-SG-005 | MUST | REQ-SG-014 | `go test -run TestLayer1NeverBlocks ./internal/hook/security/` → advisory only; PostToolUse output carries no `decision` |
| AC-SG-006 | SHOULD | REQ-SG-015 | L1 wrapper registered async in settings.json (`grep -A3 security-scan settings.json.tmpl \| grep async`) |
| AC-SG-007 | MUST | REQ-SG-020 | `go test -run TestLayer2TurnDiff ./internal/hook/security/` → engine runs over turn diff |
| AC-SG-008 | SHOULD | REQ-SG-021 | `go test -run TestLayer2Surfaces ./internal/hook/security/` → high-severity finding in `systemMessage` |
| AC-SG-009 | MUST | REQ-SG-022, REQ-SG-023 | `go test -run TestLayer2AdvisoryDefault ./internal/hook/security/` → no LLM/no block without opt-in flag; default path reuses the regex engine (both REQ-SG-022 opt-in gating AND REQ-SG-023 model-free default are exercised) |
| AC-SG-010 | MUST | REQ-SG-030, REQ-SG-031 | `go test -run TestLayer3CrossFile ./internal/hook/security/` → changed-files + related-files read (opt-in); cross-file classes (IDOR / auth-bypass / SSRF) targeted |
| AC-SG-011 | MUST | REQ-SG-032 | `go test -run TestLayer3DormantDefault ./internal/hook/security/` → no-op exit 0 when flag unset |
| AC-SG-012 | SHOULD | REQ-SG-033 | Layer-3 extends the commit-aware gate model via a sibling commit-time Stop hook (L3-A, settled) + orchestrator-mediated escalation — design.md §L3 |
| AC-SG-013 | MUST | REQ-SG-040 | `grep -rn 'AskUserQuestion\|mcp__askuser' .claude/hooks/moai/ internal/hook/security/ \| grep -v _test.go \| grep -v "^[^:]*:[0-9]*:[ \t]*#"` → 0 matches |
| AC-SG-014 | SHOULD | REQ-SG-041 | design.md §C documents the orchestrator-translation path for a block signal |
| AC-SG-015 | MUST | REQ-SG-042 | `go test -run TestAdvisoryFirst ./internal/hook/security/` → every layer advisory unless opt-in flag set |
| AC-SG-016 | SHOULD | REQ-SG-043 | `go test -run TestSkipHookAudit ./internal/hook/security/` → `--skip-hook` appends to hook-skip.log |
| AC-SG-017 | MUST | REQ-SG-050 | `diff` each wrapper local↔template `.sh`/`.sh.tmpl` render + settings.json local↔template → lockstep |
| AC-SG-018 | MUST | REQ-SG-051 | `grep -rn 'SPEC-SEC-GUARDIAN\|2026-07-24' internal/template/templates/.claude/hooks/moai/handle-security-*.sh.tmpl` → 0; `go test -run 'InternalContentLeak\|Neutrality' ./internal/template/...` PASS |
| AC-SG-019 | MUST | REQ-SG-052 | `go build ./...` + `GOOS=windows GOARCH=amd64 go build ./...` exit 0; `go test -cover ./internal/hook/security/...` ≥85%; no OTEL env in parallel tests |
| AC-SG-020 | MUST | REQ-SG-053 | `ls internal/hook/security/patterns.go` exists and is the sole pattern source; `grep -rn 'yaml.load\|innerHTML' .claude/hooks/moai/handle-security-*.sh.tmpl` → 0 (no scattered patterns in wrappers) |
| AC-SG-021 | MUST | REQ-SG-060 | `go test -run TestFailOpen ./internal/hook/security/` + wrapper no-op on missing binary → exit 0 |
| AC-SG-022 | SHOULD | REQ-SG-061 | `go test -run TestHookOutputSchema ./internal/hook/security/` → PostToolUse `additionalContext`, Stop `systemMessage`/decision, no unknown fields |
| AC-SG-023 | MUST | REQ-SG-002 | `go test -run 'Catalog' ./internal/template/...` → agent/skill counts unchanged (no new agent/skill); `find internal/template/templates -name '*.js' \| wc -l` → 0 (the template tree ships zero `.js`; `.claude/workflows/*.js` is user-owned, not template-managed — no placeholder baseline needed) |
| AC-SG-024 | SHOULD | REQ-SG-003 | design.md §A + spec.md §A.3 document the SPEC-1 (deep on-demand) vs SPEC-2 (light always-on) layering |
| AC-SG-025 | MUST | REQ-SG-001 | Guardian is native, not a third-party plugin: `grep -rn 'plugin install\|plugin add\|marketplace\|plugin.json\|\.claude-plugin' internal/hook/security/ .claude/hooks/moai/handle-security-*.sh.tmpl internal/template/templates/.claude/settings.json.tmpl` → 0 (no plugin-marketplace / install-manifest reference); realization is `internal/` Go + template shell hooks + settings.json wiring only, and `go test -run 'Catalog' ./internal/template/...` confirms agent/skill counts unchanged |

## §C Edge cases

- **MultiEdit payload**: Layer 1 must scan every edit in a MultiEdit batch (`tool_input.edits[].new_string`), not just the first.
- **Binary / non-source files**: a Write to a binary or non-source file must not spam findings — scope L1 to text/source content.
- **Large payload**: the PostToolUse payload can be large; the scan must bound its input (cf. the `handle-pre-tool.sh` 1MB stdin cap precedent) so a huge write is not truncated mid-scan silently.
- **Empty / whitespace diff at Stop**: Layer 2 must no-op cleanly when the turn made no code changes.
- **Commit with 0 code-file delta**: Layer 3 (when enabled) must skip a docs-only commit (mirror the sync-gate's `code_delta_pattern` 0-delta skip).
- **`jq` absent**: PostToolUse/Stop wrappers that parse JSON must degrade to no-op (the Go handler owns the parse; the wrapper stays thin, so the parse happens in Go — jq is not required by the guardian wrappers themselves; verify).
- **Recovery turn**: on a recovery-signal turn, Layer 2/3 gates SHOULD defer (documentation-only; hooks don't parse `stopReason`).
- **False positive**: an advisory false positive costs one line, never a broken edit — the benign-fixture test pins the baseline.

## §D Definition of Done

- [ ] All 25 ACs PASS (18 MUST + 7 SHOULD), each with verbatim evidence in the manager-develop §E report.
- [ ] `internal/hook/security/` package: patterns (single-source) + scanner + 3 layer handlers, all with tests (`t.TempDir`, no OTEL env in parallel tests), ≥85% coverage.
- [ ] 3 shell wrappers (`handle-security-{scan,turn,commit}.sh`) authored template-first, byte-lockstep with local; settings.json.tmpl wired (PostToolUse + Stop entries; L3 dormant self-gating) + rendered local.
- [ ] `moai hook security-{scan,turn,commit}` subcommands registered in `internal/cli/hook.go`.
- [ ] Boundary grep clean (no AskUserQuestion in hooks or `internal/hook/security/`).
- [ ] Fail-open verified (missing `moai`/`jq`/`git` → no-op exit 0).
- [ ] Template neutrality clean (no SPEC-ID/date/SHA in shipped wrappers/settings; no PRIMARY language); `internal_content_leak_test.go` PASS.
- [ ] Catalog agent/skill counts unchanged; no shipped `.js`; no new agent/skill.
- [ ] `go build ./...` (host + `GOOS=windows`) exit 0; `go test ./...` PASS; `golangci-lint run` 0 NEW issues.
- [ ] Both design forks settled (Layer-3 surface = L3-A; escalation delivery = orchestrator-mediated `Agent()`) — no open clarifications remain.

## §D.1 Severity gate summary

- **MUST-PASS (18)**: AC-SG-001..005, 007, 009, 010, 011, 013, 015, 017, 018, 019, 020, 021, 023, 025 — the flagship native-realization, layer behavior, boundary, neutrality, fail-open, single-source, and no-overlap invariants.
- **SHOULD-PASS (7)**: AC-SG-006, 008, 012, 014, 016, 022, 024 — async posture, surfacing, L3 surface, translation docs, skip-audit, output-schema, layering docs.
- A MUST-PASS failure blocks completion regardless of SHOULD scores (harmonic-mean gate).

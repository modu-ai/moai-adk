# Acceptance Criteria — SPEC-STEERING-ALIGN-GUARDRAIL-HOOK-001

Each AC is independently mechanically verifiable: a reproducible command plus the
expected output. Following the 5-section evidence discipline of
`verification-claim-integrity.md`, every PASS in §E.2 (run phase) must carry the
verbatim command output, not a summary.

## §A. Given-When-Then scenarios

### Scenario 1 — GLM session gets the reminder (happy path)

- **Given** a session where `ANTHROPIC_BASE_URL` contains `z.ai`
- **When** the SessionStart hook fires
- **Then** the hook's `HookSpecificOutput.AdditionalContext` contains the concise
  GLM routing reminder (the three z.ai MCP replacements + ToolSearch note), and
  the result is allow.

### Scenario 2 — Claude session gets nothing (negative path)

- **Given** a session where `ANTHROPIC_BASE_URL` is absent or does not contain
  `z.ai`
- **When** the SessionStart hook fires
- **Then** the hook injects no GLM routing reminder and returns allow.

### Scenario 3 — cg-leader pane carve-out

- **Given** a `moai cg` leader pane (Claude backend, GLM env stripped, so
  `ANTHROPIC_BASE_URL` does not contain `z.ai`)
- **When** the SessionStart hook fires
- **Then** the hook injects no GLM routing reminder (built-in tools stay
  canonical) — satisfied because the hook reads the **PROCESS** env (`hasGLMEnv` /
  `os.Getenv("ANTHROPIC_BASE_URL")`), which lacks `z.ai` on the leader pane, EVEN
  WHEN the tmux SESSION env still carries `z.ai`. The hook MUST NOT gate on
  `sessionEnvHasGLM` or `IsCGMode()` (both `true` on the leader). See AC-GH-009b.

### Scenario 4 — rule leaves the always-load set

- **Given** `glm-web-tooling.md` carries the new `paths:` frontmatter in both
  trees
- **When** the always-load count command is run
- **Then** the live count is 10 (was 11) and the template count is 9 (was 10).

### Scenario 5 — detection failure is non-blocking

- **Given** the env read errors or returns an unexpected value
- **When** the SessionStart hook fires
- **Then** the hook returns allow and injects nothing — never blocks the session.

## §B. Edge cases

- PROCESS `ANTHROPIC_BASE_URL` set to a non-z.ai custom gateway → treated as
  non-GLM (no injection).
- PROCESS `ANTHROPIC_BASE_URL` containing `z.ai` as part of a longer host (e.g.
  `api.z.ai/api/anthropic`) → GLM (injection) — substring match, consistent with
  the existing `cg_detect.go` `hasGLMEnv` semantics. (D5 note: the code matches
  `z.ai`, a strict superset of the glm-web-tooling.md `api.z.ai` SSOT wording — no
  real GLM session missed; a non-GLM `*z.ai*` gateway falsely triggers an
  advisory-only reminder, an accepted edge.)
- **cg-leader two-detector hazard (D1):** PROCESS env lacks `z.ai` but tmux SESSION
  env carries `z.ai` → MUST be treated as non-GLM (no injection). Gating on
  `sessionEnvHasGLM`/`IsCGMode()` would wrongly inject — see AC-GH-009b.
- Template tree edited but `make build` not run → live/template parity AC
  (AC-GH-006) fails, catching the omission.

## §C. Definition of Done

All ACs in §D PASS; full Go verification batch green; both rule copies
byte-aligned for the frontmatter; template-neutrality clean; before/after
always-load counts captured for both trees.

## §D. AC matrix

| AC | Requirement | Verification command | Expected |
|----|-------------|----------------------|----------|
| AC-GH-001 | REQ-GH-008 (live before) | `cnt=0; for f in $(find .claude/rules/moai -name "*.md" -type f); do sed -n '1,8p' "$f" \| grep -q '^paths:' \|\| cnt=$((cnt+1)); done; echo $cnt` — run BEFORE M4 | `11` |
| AC-GH-002 | REQ-GH-008 (live after) | same command run AFTER M4 | `10` |
| AC-GH-003 | REQ-GH-008 (template before) | `cnt=0; for f in $(find internal/template/templates/.claude/rules/moai -name "*.md" -type f); do sed -n '1,8p' "$f" \| grep -q '^paths:' \|\| cnt=$((cnt+1)); done; echo $cnt` — BEFORE M4 | `10` |
| AC-GH-004 | REQ-GH-008 (template after) | same template command AFTER M4 | `9` |
| AC-GH-005 | REQ-GH-007 (self-glob present, both trees) | `grep -c '^paths: "\*\*/glm-web-tooling.md"' .claude/rules/moai/core/glm-web-tooling.md; grep -c '^paths: "\*\*/glm-web-tooling.md"' internal/template/templates/.claude/rules/moai/core/glm-web-tooling.md` | `1` then `1` (asserts the exact `**/`-prefixed self-glob, D3 — not a bare path, not a template-tree path) |
| AC-GH-005b | REQ-GH-007 (behavioral load-on-touch parity — D4) | `f=.claude/rules/moai/core/glm-web-tooling.md; g=$(grep '^paths:' "$f" \| sed 's/paths: *"//;s/"$//'); base=$(basename "$f"); case "$base" in ${g##**/}) echo LOADS-ON-TOUCH ;; *) echo NO-MATCH ;; esac` — confirms the extracted glob's basename pattern (`glm-web-tooling.md`) matches the rule's own filename, so editing the rule still loads it on a file-touch (parity with RULE-SCOPING AC-SARS-002 — prevents a presence-pass masquerading as a non-matching glob) | `LOADS-ON-TOUCH` |
| AC-GH-006 | REQ-GH-009 (byte parity) | `diff .claude/rules/moai/core/glm-web-tooling.md internal/template/templates/.claude/rules/moai/core/glm-web-tooling.md && echo PARITY-OK` | `PARITY-OK` (no diff output) |
| AC-GH-007 | REQ-GH-001 (single signal, reused) | `grep -rn 'z\.ai' internal/hook/*.go \| grep -v '_test.go'` | shows the GLM-reminder detection reusing the `z.ai` substring check (no second divergent signal literal) |
| AC-GH-008 | REQ-GH-002 / REQ-GH-004 (GLM injection content) | `go test ./internal/hook/ -run 'GLM.*Guardrail\|GuardrailReminder' -v` | PASS — test asserts `AdditionalContext` contains `web_search_prime`, `web_reader`, `zai-mcp-server`, and `ToolSearch` when `z.ai` env is set |
| AC-GH-009 | REQ-GH-003 (no injection off-GLM) | `go test ./internal/hook/ -run 'GLM.*Guardrail.*NonGLM\|GuardrailReminder.*Absent' -v` | PASS — test asserts no GLM reminder in `AdditionalContext` when the PROCESS env lacks `z.ai` |
| AC-GH-009b | REQ-GH-005 / REQ-GH-006 (cg-leader PROCESS-env carve-out — D1 MAJOR) | `go test ./internal/hook/ -run 'GLM.*Guardrail.*Leader\|GuardrailReminder.*CgLeader' -v` | PASS — test asserts NO GLM reminder in `AdditionalContext` when the **PROCESS** env `ANTHROPIC_BASE_URL` lacks `z.ai` EVEN IF a tmux SESSION GLM marker is present (the test sets only the process env clean; gating on `sessionEnvHasGLM`/`IsCGMode` would fail this) |
| AC-GH-009c | REQ-GH-001 (process-env detector, not session-env — D1) | `grep -n 'sessionEnvHasGLM\|IsCGMode' internal/hook/*.go \| grep -v '_test.go'` | (no output) — the hook does NOT call the session-env detector nor `IsCGMode`; it reuses the PROCESS-env check only |
| AC-GH-010 | REQ-GH-012 (non-blocking on error) | `go test ./internal/hook/ -run 'GLM.*Guardrail.*Error\|GuardrailReminder.*AllowOnError' -v` | PASS — test asserts allow result + no injection on env-read error |
| AC-GH-011 | REQ-GH-010 (coverage — NEW code path, D6) | `go test ./internal/hook/... -coverprofile=/tmp/cov.out && go tool cover -func=/tmp/cov.out \| grep -E 'glm.*[Gg]uardrail\|[Gg]uardrail.*[Rr]eminder'` (new detection/injection funcs) AND `go test ./internal/hook/... -cover` (package baseline) | NEW funcs ≥ 90.0%; package coverage NOT below 82.7% baseline (NOT a package-wide 90% target — D6) |
| AC-GH-012 | REQ-GH-010 (vet + lint + full test) | `go test ./... && go vet ./... && golangci-lint run --timeout=2m` | all exit 0, zero issues |
| AC-GH-013 | REQ-GH-011 (template neutrality) | `go test ./internal/template/... -run TestTemplateNeutralityAudit` | PASS |
| AC-GH-014 | REQ-GH-009 (Template-First re-embed) | `make build && git diff --quiet internal/template/embedded.go || echo EMBED-REGENERATED` then confirm embedded copy carries the frontmatter | embedded glm-web-tooling content carries the `paths:` line |

## §D.1 Severity

- MUST-FIX (block close): AC-GH-002, AC-GH-004, AC-GH-005, AC-GH-005b, AC-GH-006,
  AC-GH-008, AC-GH-009, **AC-GH-009b** (D1 carve-out — the correctness gate),
  AC-GH-009c, AC-GH-011, AC-GH-012, AC-GH-013.
- SHOULD-FIX: AC-GH-007 (style of single-signal reuse), AC-GH-010 (error path),
  AC-GH-014 (embed regeneration confirmation).
- Baseline-only (informational): AC-GH-001, AC-GH-003 (before-counts; their value
  is to anchor the delta, not to gate).

## §D.2 Traceability

| Requirement | ACs |
|-------------|-----|
| REQ-GH-001 | AC-GH-007, AC-GH-009c (process-env detector, not session-env) |
| REQ-GH-002 | AC-GH-008 |
| REQ-GH-003 | AC-GH-009 |
| REQ-GH-004 | AC-GH-008 |
| REQ-GH-005 | AC-GH-009b (cg-leader PROCESS-env carve-out) |
| REQ-GH-006 | AC-GH-009b, AC-GH-009c (carve-out via process-env signal, no extra branch) |
| REQ-GH-007 | AC-GH-005, AC-GH-005b (self-glob form + behavioral load-on-touch) |
| REQ-GH-008 | AC-GH-001..004 |
| REQ-GH-009 | AC-GH-006, AC-GH-014 |
| REQ-GH-010 | AC-GH-011, AC-GH-012 |
| REQ-GH-011 | AC-GH-013 |
| REQ-GH-012 | AC-GH-010 |

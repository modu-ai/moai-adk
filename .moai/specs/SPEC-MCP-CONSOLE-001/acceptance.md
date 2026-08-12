# acceptance.md — SPEC-MCP-CONSOLE-001

> Verification layer. Each criterion is `AC-C-NNN`, binary-testable `Given … When … Then …`. GEARS requirements live in `spec.md`. Tier M — 14 criteria, within the 16 ceiling.

## Severity legend

- **MUST** — blocks `completed`.
- **SHOULD** — quality signal; a fail is debt.

## §D. AC Matrix

### M1-M2 — Per-tool surface and the gating seam

**AC-C-001** (MUST) — all 17 tools are represented in the console
- **Given** the console's MCP section,
- **When** it is rendered,
- **Then** a control exists for each of the 17 tools registered by `registerMoaiMCPTools`, and the count matches the registration count exactly.
- Verification: a test asserting the console's tool list length equals the registered tool count, both read from the single shared declaration.

**AC-C-002** (MUST) — a new tool cannot be added without appearing in the console
- **Given** the single shared tool declaration,
- **When** a tool is added to registration without a corresponding console entry,
- **Then** the build or a guard test FAILS.
- Verification: a guard test that derives both lists from the same declaration, so divergence is structurally impossible; or, if two lists are unavoidable, a test asserting set equality that fails on drift.

**AC-C-003** (MUST) — write-capable tools are visually and textually distinguished
- **Given** the rendered console,
- **When** the controls for `goal_arm`, `verify_snapshot`, `codex_task`, `codex_job_cancel` are inspected,
- **Then** each carries a marker distinguishing it from the 13 read-only tools, and the distinction is carried in text (not colour alone) so it survives a monochrome or screen-reader rendering.

**AC-C-004** (MUST) — a disabled tool is actually not registered
- **Given** a project configuration disabling one tool,
- **When** the MCP server starts and a host requests `tools/list`,
- **Then** the disabled tool is **absent** from the response and the remaining tools are present.
- This is the criterion that makes REQ-C-2 falsifiable: it tests the effect, not that a setting persisted. A test asserting only that the config value round-trips does NOT satisfy it.
- Verification: a `tools/list` round-trip against a server constructed with one tool disabled.

**AC-C-005** (MUST) — the setting round-trips through the existing seam
- **Given** the console's MCP section,
- **When** a tool is toggled and saved,
- **Then** the value is written through the same schema-driven form + `yamlpatch` seam the audit-selection and `branch_guard` fields use, and is readable on the next render.
- Verification: the new field names appear in `settings.AllFields()` for the chosen section, mirroring the assertion shape of `internal/web/mcp_audit_surface_test.go:22-40`.

### M3 — codex authentication

**AC-C-006** (MUST) — probe state is displayed, not recomputed
- **Given** codex installed and authenticated,
- **When** the console MCP section renders,
- **Then** it shows `installed`, `binary`, `version`, `auth_provider`, `enable_review_gate`, and `allow_write` as reported by the `codex_setup` probe, and `internal/web` contains no second auth-classification implementation.
- Verification: `grep -rn 'classifyCodexAuth\|codex login status' internal/web/ --include=*.go | grep -v _test.go` → no matches.

**AC-C-007** (MUST) — unauthenticated state surfaces the command, not a login flow
- **Given** codex installed but unauthenticated (`auth_provider` = unknown),
- **When** the console renders,
- **Then** it names the codex login command as the remediation and offers no in-console login action, no OAuth redirect, and no browser launch.
- Verification: no new route in `internal/web/app.go` performs or redirects to an external auth flow.

**AC-C-008** (MUST) — codex absent degrades gracefully
- **Given** codex is not installed,
- **When** the console renders,
- **Then** it shows the not-installed state without erroring, matching the probe's own fail-open behavior (`installed: false`, `auth_provider: unknown`).

**AC-C-009** (MUST) — opt-in toggles write the seam the gates read
- **Given** the review-gate and write-mode toggles,
- **When** each is saved,
- **Then** the value is read back by the existing fail-closed readers (`readCodexReviewGateEnabled`, `readCodexTaskAllowWrite`) — one source of truth, no parallel key.

### M4 — GLM key

**AC-C-010** (MUST) — the GLM credential path is reused unchanged
- **Given** the GLM key surface in the MCP console section,
- **When** `internal/web/glmkey.go` is diffed against base,
- **Then** the parse / validate / disclosure / reveal logic is unchanged, and the field is still absent from `settings.AllFields()`.
- Verification: `go test -run 'TestGLMKeyField_AbsentFromSchema' ./internal/web/...` → `PASS`; `git diff ed70e4354 -- internal/web/glmkey.go` shows no logic change.

**AC-C-011** (MUST) — disclosure stays bounded
- **Given** a stored key longer than four characters,
- **When** the console renders the hint,
- **Then** only a configured-boolean and the final four characters cross into the view model. **And given** a key of four characters or fewer, **then** only the boolean is exposed and no characters are shown.
- Verification: the existing `computeGLMKeyHint` tests still PASS and cover both branches.

### M5 — secret hygiene and i18n

**AC-C-012** (MUST) — no credential in any git-tracked file
- **Given** every file this SPEC touches or writes at runtime,
- **When** they are scanned,
- **Then** no resolved credential appears in a git-tracked file, and any env reference in `.mcp.json` is a `${VAR}` literal.
- Verification: `grep -rEc '(sk-|ghp_|Bearer [A-Za-z0-9])' .mcp.json internal/template/templates/.mcp.json` → `0`; `go test ./internal/template/... -run 'TestMCPNeutrality|TestInternalContentLeak'` → `PASS`.

**AC-C-013** (MUST) — 4-locale coverage complete
- **Given** every user-facing string added by this SPEC,
- **When** the i18n governance suite runs,
- **Then** it PASSES: every new key exists in en, ko, ja, and zh; no orphan key; no non-en value identical to its en value except where the endonym invariants allow.
- Verification: `go test ./internal/web/... -run 'TestI18n' 2>&1 | grep -E 'PASS|FAIL'` → `PASS`.

**AC-C-014** (MUST) — no forked interpreter
- **Given** `internal/web` after the change,
- **When** the no-fork guards run,
- **Then** they PASS: `internal/web` neither defines `activeAuditBackend` nor redefines `ResolveAgentModelEffort`.
- Verification: `go test -run 'TestWebConsole_AuditNoForkedInterpreter|TestWebConsole_ResolveAgentModelEffortSSOTShared' ./internal/web/...` → `PASS`.

## §D.1 Traceability

| REQ | AC | Milestone |
|---|---|---|
| REQ-C-1 | AC-C-001, AC-C-002, AC-C-005 | M2 |
| REQ-C-2 | AC-C-004 | M1 |
| REQ-C-3 | AC-C-003 | M2 |
| REQ-C-4 | AC-C-006, AC-C-008 | M3 |
| REQ-C-5 | AC-C-007 | M3 |
| REQ-C-6 | AC-C-009 | M3 |
| REQ-C-7 | AC-C-010, AC-C-011 | M4 |
| REQ-C-8 | AC-C-012 | M5 |
| REQ-C-9 | AC-C-013 | M5 |
| REQ-C-10 | AC-C-014 | M5 |

## §D.2 Edge cases

- **A tool disabled while an agent's frontmatter grants it (SPEC-B)** — the agent's tool name becomes inert. Accepted as user intent (spec.md R-C-3); AC-C-003's write-capable marking is what makes the consequential cases legible at decision time.
- **A GLM key of exactly four characters** — covered explicitly by AC-C-011's second branch; the naive "last four or the whole key" fallback would disclose it entirely.
- **codex installed, `login status` output unrecognized** — probe fails open to `unknown`; AC-C-007's unauthenticated path applies.
- **All 17 tools disabled** — the server starts with an empty tool list rather than failing. Not an error state; no AC.

## §D.3 Definition of Done

- Both `[NEEDS CLARIFICATION]` markers in `plan.md` §C resolved before Implementation Kickoff Approval.
- All 14 MUST criteria PASS with cited command output.
- `go test ./...` green; `golangci-lint run` shows no NEW findings.
- `.templ` sources and their regenerated `_templ.go` committed together.
- Secret-hygiene and i18n governance suites green.

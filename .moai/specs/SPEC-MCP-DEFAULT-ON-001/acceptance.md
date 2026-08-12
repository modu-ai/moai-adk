# acceptance.md — SPEC-MCP-DEFAULT-ON-001

> Verification layer. Each criterion is `AC-A-NNN`, written as binary-testable `Given … When … Then …`. GEARS requirements live in `spec.md` (`REQ-A-N`); this file does not restate them. Tier M — 16-AC ceiling; 14 criteria below.

## Severity legend

- **MUST** — blocks `completed`.
- **SHOULD** — quality signal; a fail is debt, not a blocker.

## §D. AC Matrix

### M1 — Amendments

**AC-A-001** (MUST) — MOAI-MCP-SERVER-001 amended in place
- **Given** `.moai/specs/SPEC-MOAI-MCP-SERVER-001/`,
- **When** an auditor reads `spec.md` and `acceptance.md`,
- **Then** REQ-MCP-002 states a default-on gate, REQ-MCP-015 describes an opt-out flag, AC-MCP-002 and AC-MCP-006 are marked amended with their original text preserved for comparison, a dated `### Amendments` HISTORY sub-section records cause and scope, `version:` is bumped, `updated:` is `2026-08-12`, and `status:` is still `completed`.
- Verification: `grep -c 'AMENDED' .moai/specs/SPEC-MOAI-MCP-SERVER-001/spec.md` → `≥2`; `grep '^status:' …/spec.md` → `status: completed`.

**AC-A-002** (MUST) — TREND-MCP-001 amended in place, invariants preserved
- **Given** `.moai/specs/SPEC-TREND-MCP-001/`,
- **When** an auditor reads `spec.md` and `acceptance.md`,
- **Then** REQ-TMC-003 states default-on for `moai`, REQ-TMC-001's active count reads one, AC-TMC-001 and AC-TMC-004 are amended, **and REQ-TMC-002's secret-hygiene clause and REQ-TMC-004's `$comment`-free clause are byte-unchanged and still asserted by the amended ACs**.
- Verification: `git diff ed70e4354 -- .moai/specs/SPEC-TREND-MCP-001/spec.md` shows no modification to the REQ-TMC-002 line; the amended AC-TMC-001 verification line still contains `$comment`.

**AC-A-003** (MUST) — reversal rationale is readable from the artifact
- **Given** either amended SPEC,
- **When** an auditor asks "why is a completed MUST AC being inverted?",
- **Then** the answer is present in prose adjacent to the amended criterion — not only in a commit message — naming the cause (the owner's first-class-default directive), what the original clause actually guarded, and which clauses are preserved.

### M2 — Template + repo-root shape

**AC-A-004** (MUST) — template active map is exactly `moai`
- **Given** the rewritten `internal/template/templates/.mcp.json`,
- **When** it is parsed,
- **Then** the active `mcpServers` map has exactly one key, `moai`, whose value is `{"command":"moai","args":["mcp-server"]}` with no `env` block; `$schema` and `staggeredStartup` are present and unchanged.
- Verification: `jq -c '.mcpServers | keys' internal/template/templates/.mcp.json` → `["moai"]`; `jq '.mcpServers.moai | has("env")' …` → `false`.

**AC-A-005** (MUST) — repo-root carries four entries and the divergence is documented
- **Given** the repo-root `.mcp.json`,
- **When** it is parsed,
- **Then** the active map has exactly four keys — `context7`, `chrome-devtools`, `playwright`, `moai` — and the intentional divergence from the template is stated in this SPEC's `spec.md` REQ-A-2 and `plan.md` §A.2.
- Verification: `jq -c '.mcpServers | keys | length' .mcp.json` → `4`; `grep -c 'diverge' .moai/specs/SPEC-MCP-DEFAULT-ON-001/plan.md` → `≥1`.

**AC-A-006** (MUST) — neutrality guard inverted, not weakened
- **Given** `internal/template/mcp_template_neutrality_test.go` after the change,
- **When** `go test -run TestMCPNeutrality ./internal/template/...` runs,
- **Then** it PASSES; `mcpAllowedActiveKeys` is `{moai}`; the count assertion expects 1; **and** the forbidden-token regex set (SPEC-ID, 40-char SHA, short SHA, `/Users/`, `CLAUDE.local.md`, `PR #N`, `$comment`) is byte-unchanged from base.
- Verification: `go test -run TestMCPNeutrality -v ./internal/template/... 2>&1 | grep -E 'PASS|FAIL'` → `PASS`; `git diff ed70e4354 -- internal/template/mcp_template_neutrality_test.go` shows no change inside the `mcpForbiddenTokenRes` block.

**AC-A-007** (MUST) — secret hygiene holds on the new single-entry template
- **Given** the rewritten template `.mcp.json`,
- **When** it is scanned for credential material,
- **Then** no resolved secret appears and every env reference, if any exists, is a `${VAR}` literal. The `moai` entry carries no `env` block at all, so the check passes vacuously — which is the intended shape, not an evasion.
- Verification: `grep -cE '(sk-|ghp_|Bearer [A-Za-z0-9])' internal/template/templates/.mcp.json` → `0`; `go test -run TestInternalContentLeak ./internal/template/...` → `PASS`.

### M3 — Provisioning gate

**AC-A-008** (MUST) — default path provisions the entry
- **Given** a fresh project and a user who did not decline,
- **When** `moai init` runs,
- **Then** the project's `.mcp.json` contains exactly one `moai` entry, written through `mutateClaudeJSONAtomic`, and the provisioning is reported on stdout.
- Verification: the inverted successor of `internal/cli/init_mcp_provision_test.go:44` (`TestProvisionMCPEntryIfOptedIn_OptedIn`), re-pointed at the default path, PASSES.

**AC-A-009** (MUST) — explicit decline is honored and silent
- **Given** a fresh project and a user who explicitly declined,
- **When** `moai init` runs,
- **Then** no `.mcp.json` is created and neither stdout nor stderr emits anything about MCP.
- Verification: the inverted successor of `internal/cli/init_mcp_provision_test.go:27` PASSES with the decline branch.

**AC-A-010** (MUST) — wizard default flipped, polarity kept positive
- **Given** `internal/cli/wizard/questions.go`,
- **When** the MCP provisioning question is read,
- **Then** its `Default` is `"true"`, its `Title` is a positive question (not "Skip …?"), and its ID no longer asserts opt-in semantics.
- Verification: `grep -A6 'MCP' internal/cli/wizard/questions.go | grep 'Default:'` → `Default:     "true"`; `grep -c 'mcp_tools_opt_in' internal/cli/` → `0`.

**AC-A-011** (SHOULD) — provisioning function name matches its behavior
- **Given** `internal/cli/init.go`,
- **When** the provisioning function and its doc comment are read,
- **Then** neither the identifier nor the comment describes an opt-in gate.
- Verification: `grep -c 'provisionMCPEntryIfOptedIn' internal/cli/` → `0`.

### M4 — update merge

**AC-A-012** (MUST) — a user's own MCP entry survives `moai update`
- **Given** a project whose `.mcp.json` carries the shipped `moai` entry plus a user-added entry (e.g. `my-tool`),
- **When** `moai update` runs,
- **Then** both entries are present afterwards — the user's entry is preserved by the 3-way merge, not overwritten by the template deploy.
- Verification: a new test in `internal/cli/` asserting both keys post-update; `grep -c '"\.mcp\.json"' internal/cli/update_template_sync.go` → `≥1`.

**AC-A-013** (MUST) — the false comment is corrected
- **Given** `internal/cli/update_template_sync.go`,
- **When** the `collectMergeableFiles` comment is read,
- **Then** it no longer claims "MoAI no longer ships an MCP template (full MCP removal)" and instead states that `.mcp.json` is shipped and is a merge target.
- Verification: `grep -c 'no longer ships an MCP template' internal/cli/update_template_sync.go` → `0`.

### M5 — documentation

**AC-A-014** (MUST) — every MCP-default surface tells one story
- **Given** the three wizard locale strings, `.claude/rules/moai/core/settings-management.md` + its template mirror, and the template `CLAUDE.md` MCP inventory,
- **When** each is read after the change,
- **Then** none describes MCP provisioning as opt-in-default-off, none states three active third-party entries as the distributed default, none names a tool the distributed user no longer receives by default without saying how to activate it, and no template-side string carries a SPEC ID, REQ token, commit SHA, internal date, `/Users/` path, or `CLAUDE.local.md` reference.
- Verification: `grep -rc 'Opt-in default-off' internal/cli/wizard/translations.go` → `0`; `grep -rEc 'SPEC-[A-Z]|REQ-[A-Z]' internal/cli/wizard/translations.go internal/template/templates/.mcp.json` → `0`; `go test ./internal/template/... -run 'TestTemplateNeutrality|TestInternalContentLeak'` → `PASS`.

## §D.1 Traceability

| REQ | AC | Milestone |
|---|---|---|
| REQ-A-1 | AC-A-004, AC-A-007 | M2 |
| REQ-A-2 | AC-A-005 | M2 |
| REQ-A-3 | AC-A-008, AC-A-009, AC-A-010, AC-A-011 | M3 |
| REQ-A-4 | AC-A-012, AC-A-013 | M4 |
| REQ-A-5 | AC-A-014 | M5 |
| REQ-A-6 | AC-A-001 | M1 |
| REQ-A-7 | AC-A-002 | M1 |
| REQ-A-8 | AC-A-003 | M1 |
| REQ-A-9 | AC-A-006 | M2 |
| REQ-A-10 | AC-A-004, AC-A-014 | M2, M5 |

## §D.2 Edge cases

- **`moai` not on PATH in the host environment** — the runtime reports a failing MCP server at session start. Accepted (spec.md R-A-3); no AC.
- **User already has a `.mcp.json` with their own `moai` key pointing elsewhere** — the 3-way merge (AC-A-012) governs; the idempotent-skip in `provisionMoaiMCPServerEntryAt` prevents a duplicate insert.
- **Template `.mcp.json` with `staggeredStartup` but one entry** — the staggered-start block is preserved verbatim (AC-A-004) even though it has little to stagger; removing it is out of scope.

## §D.3 Definition of Done

- All 13 MUST criteria PASS with cited command output.
- `go test ./...` green; `golangci-lint run` shows no NEW findings.
- `make build` run after every template edit; committed tree and embedded catalog consistent.
- Both amended SPECs still lint clean (`moai spec lint`) and still read `status: completed`.

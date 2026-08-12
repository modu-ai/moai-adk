# acceptance.md — SPEC-MCP-AGENT-WIRING-001

> Verification layer. Each criterion is `AC-B-NNN`, binary-testable `Given … When … Then …`. GEARS requirements live in `spec.md`. Tier M — 11 criteria (iter-2: was 12; the redundant codex-guard AC was dropped after §C was corrected), within the 16 ceiling.

## Severity legend

- **MUST** — blocks `completed`.
- **SHOULD** — quality signal; a fail is debt.

## §D. AC Matrix

### M1 — Grant matrix

**AC-B-001** (MUST) — the seven granted agents carry exactly their matrix row
- **Given** `.claude/agents/moai/` after the change,
- **When** each granted agent's `tools:` line is read,
- **Then** it carries exactly the `mcp__moai__*` entries listed in `plan.md` §B.1 for that agent, and no other `mcp__moai__*` entry.
- Verification: `grep -n '^tools:' .claude/agents/moai/*.md` — `manager-spec` shows `spec_progress, spec_audit, spec_drift`; `manager-develop` shows `verify_snapshot, verify_trend, goal_status`; `manager-docs` shows `spec_progress, spec_audit`; `plan-auditor` shows `audit_multi, spec_audit, spec_drift, codex_audit`; `sync-auditor` shows `audit_multi, verify_trend, audit_cache, glm_audit`; `manager-lead` shows `session_list, goal_status`; `super-advisor` shows `spec_audit, verify_trend`.

**AC-B-002** (MUST) — the five non-granted agents carry no MCP tool
- **Given** `manager-git`, `manager-design`, `e2e-tester`, `builder-harness`,
- **When** each `tools:` line is read,
- **Then** none contains `mcp__moai__`. (`Explore` has no file.)
- Verification: `grep -l 'mcp__moai__' .claude/agents/moai/{manager-git,manager-design,e2e-tester,builder-harness}.md` → no matches.

**AC-B-003** (MUST) — exactly one write-capable grant exists
- **Given** the four write-capable tools (`goal_arm`, `verify_snapshot`, `codex_task`, `codex_job_cancel`),
- **When** the full agent directory is scanned,
- **Then** `verify_snapshot` appears exactly once (on `manager-develop`) and `goal_arm`, `codex_task`, `codex_job_cancel` appear zero times.
- Verification: `grep -c 'mcp__moai__verify_snapshot' .claude/agents/moai/*.md | grep -v ':0'` → one file, count 1; `grep -rc 'mcp__moai__goal_arm\|mcp__moai__codex_task\|mcp__moai__codex_job_cancel' .claude/agents/moai/` → all `0`.

**AC-B-004** (MUST) — `goal_arm` omission from `manager-lead` is a recorded decision
- **Given** the Epic brief proposed `goal_arm` for `manager-lead`,
- **When** an auditor asks why it is absent,
- **Then** `plan.md` §B.2 states the reason (goal arming changes the session termination condition and belongs to the orchestrator downstream of Implementation Kickoff Approval) and names the one-line change if the owner overrides.
- Verification: `grep -c 'goal_arm' .moai/specs/SPEC-MCP-AGENT-WIRING-001/plan.md` → `≥3`.

**AC-B-005** (MUST) — every omission is justified, none is silent
- **Given** `plan.md` §B.3,
- **When** it is read against the retained-agent catalog,
- **Then** each of the five non-granted agents has a named reason, and the count of documented agents (7 granted + 5 non-granted) equals the 12-agent catalog.

**AC-B-006** (MUST) — existing entries preserved, CSV form intact
- **Given** the seven edited files,
- **When** each `tools:` line is diffed against base `ed70e4354`,
- **Then** every pre-existing entry is still present in its original order, the additions are appended, `plan-auditor` and `sync-auditor` still carry `mcp__moai__audit_multi`, and no `tools:` value became a YAML array.
- Verification: `git diff ed70e4354 -- .claude/agents/moai/ | grep '^-tools:'` shows only lines whose additions are supersets; `grep -A1 '^tools:' .claude/agents/moai/*.md | grep -c '^\s*-\s'` → `0`.

**AC-B-007** (MUST) — `manager-lead` depth-2 seal untouched
- **Given** the depth guard,
- **When** `go test -run TestManagerLeadDepth ./internal/template/...` runs,
- **Then** it PASSES, and no agent other than `manager-lead` carries `Agent` in `tools:`.
- Verification: `go test -run TestManagerLeadDepth -v ./internal/template/... 2>&1 | grep -E 'PASS|FAIL'` → `PASS`.

### REQ-B-6 — Model/effort SSOT (verification-shaped; no code change in this SPEC, no milestone deliverable)

**AC-B-008** (MUST) — the existing SSOT guards still pass unchanged
- **Given** `mcp_audit_test.go` (`TestMCPAudit_NoDirectFrontmatterRead`), `mcp_glm_test.go`, `mcp_convergence_test.go`,
- **When** `go test ./internal/cli/...` runs,
- **Then** all three SSOT assertions PASS unchanged, confirming no regression was introduced by this SPEC (which edits agent markdown only and touches no Go file).

**AC-B-009** (SHOULD) — the SSOT scope finding is recorded honestly, including the codex session path
- **Given** the Epic brief framed the SSOT extension as a deliverable and the iter-1 plan had identified the codex session path as "the one real gap",
- **When** `plan.md` §C is read against the base tree,
- **Then** it states that all four model-invoking tools already resolve through `ResolveAgentModelEffort` at base, names each resolution site with a file:line, AND states that the entire codex session path (`openCodexSession`, `openCodexSessionOn`, `threadParams["model"] = me.Model`) lives in `mcp_codex.go` — which `TestMCPAudit_NoDirectFrontmatterRead` at `internal/cli/mcp_audit_test.go:148` already scans — so the path is already structurally guarded and no Go code change is part of this SPEC.
- Verification: `grep -c 'no Go code change' .moai/specs/SPEC-MCP-AGENT-WIRING-001/plan.md` ≥ 1; `grep -c 'mcp_audit_test.go:148' .moai/specs/SPEC-MCP-AGENT-WIRING-001/plan.md` ≥ 1.

### M2 — Template mirror

**AC-B-010** (MUST) — every edited agent file has a matching mirror
- **Given** the seven edited files under `.claude/agents/moai/`,
- **When** the template source is compared,
- **Then** `internal/template/templates/.claude/agents/moai/` carries a corresponding file whose `tools:` line matches, and `make build` has regenerated the embedded catalog in the same change.
- Verification: for each edited file, `diff <(grep '^tools:' .claude/agents/moai/<f>) <(grep '^tools:' internal/template/templates/.claude/agents/moai/<f>)` → empty; `git status --porcelain internal/template/catalog.yaml` shows it staged alongside.

**AC-B-011** (MUST) — mirrored content is neutral
- **Given** the mirrored agent files,
- **When** the neutrality guards run,
- **Then** no SPEC ID, REQ token, commit SHA, internal date, `/Users/` path, or `CLAUDE.local.md` reference appears in any of them.
- Verification: `go test ./internal/template/... -run 'TestTemplateNeutrality|TestInternalContentLeak' 2>&1 | grep -E 'PASS|FAIL'` → `PASS`.

## §D.1 Traceability

| REQ | AC | Milestone |
|---|---|---|
| REQ-B-1 | AC-B-001, AC-B-002, AC-B-005 | M1 |
| REQ-B-2 | AC-B-003, AC-B-004 | M1 |
| REQ-B-3 | AC-B-010, AC-B-011 | M2 |
| REQ-B-4 | AC-B-006 | M1 |
| REQ-B-5 | AC-B-006 | M1 |
| REQ-B-6 | AC-B-008, AC-B-009 | (verification-shaped; no milestone deliverable — SSOT already complete at base) |

## §D.2 Edge cases

- **An agent whose `tools:` line wraps** — none do at base; if one is introduced, the CSV-string form (AC-B-006) still governs.
- **A user project without the `moai` MCP server** — granted tool names are inert, not errors. SPEC-A makes the server present by default; this SPEC does not re-check that.
- **A future 13th retained agent** — the catalog count assertion in AC-B-005 will fail, which is the intended signal to extend the matrix deliberately.

## §D.3 Definition of Done

- All 10 MUST criteria PASS with cited command output.
- `go test ./...` green; `golangci-lint run` shows no NEW findings. (This SPEC edits agent markdown only and adds no Go file; the existing SSOT guards at `mcp_audit_test.go:148`, `mcp_glm_test.go:227`, `mcp_convergence_test.go:518-526` continue to PASS unchanged — AC-B-008.)
- `make build` run; catalog committed with the mirrors.
- Zero write-capable grants beyond the single justified `verify_snapshot`.

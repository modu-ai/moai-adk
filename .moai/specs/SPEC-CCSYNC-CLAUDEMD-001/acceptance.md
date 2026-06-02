# Acceptance Criteria — SPEC-CCSYNC-CLAUDEMD-001

> All acceptance criteria are mechanically checkable. Each AC names a verifiable
> command (grep / version-string check / build / test) with the expected result.
> Run-phase exit requires every Blocking AC to PASS. Verified against HEAD
> `5042e309c`; the run phase re-confirms line numbers before editing.

## A. Definition of Done

- All 13 requirements (REQ-CCSYNC-001..013) satisfied.
- All Blocking ACs (AC-CCSYNC-001..017) PASS.
- `make build` clean; `internal/template/embedded.go` regenerated.
- `go test ./internal/template/...` green (neutrality audit + internal-content leak guard + mirror-drift).
- No file outside the in-scope set modified (docs-site untouched per EXC-1; context-window-management.md / session-handoff.md untouched per EXC-3).

## B. Given-When-Then Scenarios

### Scenario 1 — H1: template no longer ships archived-agent references (REQ-CCSYNC-002/003/004)

- **Given** the stale template CLAUDE.md at HEAD `5042e309c` lists `expert-backend` (§2), `manager-strategy`/`expert-backend`/`expert-frontend`/`manager-quality` (§5), and `manager-quality`/`expert-devops` (§11) as recommended/routed agents,
- **When** the run phase ports the dev-root v14.2.0 §2/§5/§11 content into the template,
- **Then** none of the 12 archived agent names appears as an invocation example, chain phase, or error-routing target in template CLAUDE.md §2/§5/§11 (archived names may still appear ONLY in the line-127 archived-list and migration cross-references, which are correct).

### Scenario 2 — H1: template version caught up (REQ-CCSYNC-001)

- **Given** the template CLAUDE.md is `Version: 14.1.0` and the dev-root is `14.2.0`,
- **When** the run phase bumps the template version line,
- **Then** the template CLAUDE.md `Version:` line reads 14.2.0 or higher.

### Scenario 3 — H1: ported content stays template-neutral (REQ-CCSYNC-005)

- **Given** the dev-root §2/§5/§11 content references an internal SPEC ID and date in its change-note,
- **When** the run phase ports that content into the template and genericizes per §25,
- **Then** the template CLAUDE.md contains no moai-adk-internal SPEC ID, REQ/AC token, audit citation, internal ISO date, or commit SHA, AND `go test ./internal/template/... -run TestTemplateNeutralityAudit` passes.

### Scenario 4 — H2: context-window threshold aligned to SSOT (REQ-CCSYNC-007/008)

- **Given** four locations say "1M = 75%" / "75% of the window" contradicting the canonical 1M = 50% SSOT,
- **When** the run phase replaces the threshold references,
- **Then** no in-scope file contains "1M = 75%" and the settings-management.md copies reference the model-specific threshold (1M=50%, 200K=90%), matching context-window-management.md.

### Scenario 5 — H8: Agent Teams version + spawn-capability reconciled (REQ-CCSYNC-009/010/011)

- **Given** both CLAUDE.md copies say `v2.1.50` for Agent Teams and agent-authoring.md line ~212 claims "teammates CAN spawn other teammates ... v2.1.50+",
- **When** the run phase corrects the version and reconciles the spawn-capability wording,
- **Then** both CLAUDE.md §15 copies say `v2.1.32`, AND agent-authoring.md (both copies) no longer asserts that teammates can spawn other teammates, AND the `v2.1.50+` string is removed/corrected in that sentence, AND the file is internally consistent with its own "subagents cannot spawn other subagents" statement.

### Scenario 6 — Low-priority: dead reference fixed (REQ-CCSYNC-012)

- **Given** both CLAUDE.md copies reference the nonexistent `workflow/progressive-disclosure.md`,
- **When** the run phase corrects or removes the reference,
- **Then** neither CLAUDE.md copy references `progressive-disclosure.md`; if a replacement is used it is `development/skill-authoring.md` (verified to exist).

### Scenario 7 — Mirror parity + build (REQ-CCSYNC-006/013)

- **Given** settings-management.md and agent-authoring.md each have a `templates/` mirror,
- **When** the run phase edits both copies of each in the same commit and runs `make build`,
- **Then** `go test ./internal/template/...` passes (including mirror-drift + internal-content leak), AND `internal/template/embedded.go` is regenerated.

### Scenario 8 (edge) — sibling SPEC sequencing honored (CON-5)

- **Given** `SPEC-CCSYNC-TOOLCAT-001` also targets agent-authoring.md at run-phase,
- **When** the orchestrator enters this SPEC's run phase,
- **Then** the pre-flight check confirms the sibling is not mid-run, this SPEC commits its agent-authoring.md edits first, and the sibling rebases onto this SPEC's commit.

## C. AC Grep / Verification Matrix (Blocking)

All commands run from repo root. Paths: `T = internal/template/templates`,
`R = .claude/rules/moai`.

| AC | REQ | Command | Expected |
|----|-----|---------|----------|
| AC-CCSYNC-001 | 001 | `grep -E '^Version: 14\.[2-9]' internal/template/templates/CLAUDE.md` | ≥1 match (template version ≥ 14.2.0) |
| AC-CCSYNC-002 | 002 | `sed -n '/### Phase 3: Execute/,/### Phase 4/p' internal/template/templates/CLAUDE.md \| grep -cE 'expert-backend\|expert-frontend\|manager-strategy\|manager-quality\|expert-devops'` | `0` |
| AC-CCSYNC-003 | 003 | `sed -n '/### Agent Chain for SPEC Execution/,/### MX Tag Integration/p' internal/template/templates/CLAUDE.md \| grep -cE 'manager-strategy\|expert-backend\|expert-frontend\|expert-security\|expert-devops\|expert-performance\|expert-refactoring\|manager-quality\|manager-brain\|manager-project\|claude-code-guide'` | `0` |
| AC-CCSYNC-004 | 004 | `sed -n '/### Error Recovery/,/### Resumable Agents/p' internal/template/templates/CLAUDE.md \| grep -cE 'manager-quality\|expert-devops\|expert-backend\|expert-frontend\|manager-strategy'` | `0` |
| AC-CCSYNC-005 | 003 | `sed -n '/### Agent Chain for SPEC Execution/,/### MX Tag Integration/p' internal/template/templates/CLAUDE.md \| grep -cE 'manager-spec\|plan-auditor\|manager-develop\|manager-docs\|sync-auditor'` | ≥1 (retained chain present) |
| AC-CCSYNC-006 | 005 | `go test ./internal/template/... -run TestTemplateNeutralityAudit` | PASS (exit 0) |
| AC-CCSYNC-007 | 005 | `grep -cE 'SPEC-V3R6-\|SPEC-AGENCY-\|REQ-[A-Z]+-[0-9]{3}\|AC-[A-Z]+-[0-9]{3}' internal/template/templates/CLAUDE.md` | `0` (no internal tokens in template CLAUDE.md) |
| AC-CCSYNC-008 | 007 | `grep -rh '1M = 75%' CLAUDE.md internal/template/templates/CLAUDE.md \| wc -l` | `0` (summed across both copies) |
| AC-CCSYNC-009 | 007 | `grep -c '1M = 50%' CLAUDE.md; grep -c '1M = 50%' internal/template/templates/CLAUDE.md` | `1` and `1` (one per copy) |
| AC-CCSYNC-010 | 008 | `grep -rhE '75% of the window' .claude/rules/moai/core/settings-management.md internal/template/templates/.claude/rules/moai/core/settings-management.md \| wc -l` | `0` (summed across both copies) |
| AC-CCSYNC-011 | 008 | `grep -rhE '50%\|model-specific' .claude/rules/moai/core/settings-management.md internal/template/templates/.claude/rules/moai/core/settings-management.md \| wc -l` | ≥2 (threshold referenced in both copies) |
| AC-CCSYNC-012 | 009 | `grep -rh 'v2.1.50' CLAUDE.md internal/template/templates/CLAUDE.md \| wc -l` | `0` (summed across both copies) |
| AC-CCSYNC-013 | 009 | `grep -c 'v2.1.32' CLAUDE.md; grep -c 'v2.1.32' internal/template/templates/CLAUDE.md` | `1` and `1` (one §15 line per copy) |
| AC-CCSYNC-014 | 010 | `grep -rhE 'teammates CAN spawn other teammates' .claude/rules/moai/development/agent-authoring.md internal/template/templates/.claude/rules/moai/development/agent-authoring.md \| wc -l` | `0` (summed across both copies) |
| AC-CCSYNC-015 | 012 | `grep -rh 'progressive-disclosure.md' CLAUDE.md internal/template/templates/CLAUDE.md \| wc -l` | `0` (summed across both copies) |
| AC-CCSYNC-016 | 006/013 | `go test ./internal/template/...` | PASS (exit 0 — neutrality + leak + mirror-drift all green) |
| AC-CCSYNC-017 | 011 | `grep -rh 'v2.1.50' .claude/rules/moai/development/agent-authoring.md internal/template/templates/.claude/rules/moai/development/agent-authoring.md \| wc -l` | `0` (co-located version string removed/corrected in both copies) |

Notes:
- AC-CCSYNC-009 / AC-CCSYNC-013 run two single-file `grep -c` invocations (one per copy) and expect `1` and `1` because each string occurs once per CLAUDE.md copy (dev-root + template). A single multi-file `grep -c fileA fileB` is deliberately NOT used: `grep -c` over multiple files prints per-file `fileA:N` / `fileB:N` lines, never a single summed `2`. If the run phase legitimately rewords a sentence such that the literal per-file count differs, the run phase MUST re-derive the expected count and document the deviation; the intent (zero stale values, both copies corrected) is the binding criterion.
- Count-fragility: the "expect 0" ACs (008/010/012/014/015) use the true-sum form `grep -rh '<pattern>' fileA fileB | wc -l` so the expected value is a single summed integer (`0`), avoiding the multi-file `grep -c` per-file-count pitfall.
- AC-CCSYNC-014: line 92 ("subagents cannot spawn other subagents") is the correct statement and remains; only the line-212 "teammates CAN spawn other teammates" claim is removed.

## D. Quality Gate Criteria (must-pass)

- **Functionality**: all 17 Blocking ACs PASS.
- **Consistency**: dev-root and template copies agree (no remaining drift) for the four threshold/version/dead-ref pairs; mirror byte-parity holds.
- **Neutrality**: template carries no internal SPEC IDs / REQ / AC / dates / SHAs (AC-CCSYNC-006/007).
- **Build integrity**: `make build` clean + embedded.go regenerated (AC-CCSYNC-016 + diff-stat evidence).
- **Scope discipline**: `git diff --name-only` shows ONLY the in-scope files (2× CLAUDE.md, 2× settings-management.md, 2× agent-authoring.md, embedded.go) — no docs-site, no context-window-management.md, no session-handoff.md.

## E. Out-of-Scope Verification (negative ACs)

| AC | Check | Expected |
|----|-------|----------|
| AC-CCSYNC-N1 | `git diff --name-only \| grep -c 'docs-site'` | `0` (EXC-1 honored) |
| AC-CCSYNC-N2 | `git diff --name-only \| grep -cE 'context-window-management.md\|session-handoff.md'` | `0` (EXC-3 honored) |

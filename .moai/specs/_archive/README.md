# Archived SPECs

Tombstones for SPECs that have been superseded, made obsolete by code changes,
or were determined to be below minimum viable size.

## Archive Index (2026-04-24)

| SPEC | Reason | Successor / Evidence |
|---|---|---|
| SPEC-AGENCY-001 | Superseded by full absorb plan | `SPEC-AGENCY-ABSORB-001` (merged PR #682) |
| SPEC-DESIGN-CONST-AMEND-001 | Constitution change already applied | `.claude/rules/moai/design/constitution.md` v3.3.0 HISTORY entry (2026-04-20) |
| SPEC-ORCH-001 | Trivial docs change, no SDD artifact value | Changes already in `plan.md` / `run.md` / `sync.md` workflow skills |
| SPEC-INFRA-001 | Below minimum viable SPEC size (585 bytes) | CLAUDE.md cleanup completed; no further infra work tracked here |
| SPEC-THIN-CMDS-001 | Test-enforced — no SPEC needed | `internal/template/commands_audit_test.go` enforces thin-command pattern at build time |
| SPEC-SKILL-001 | Underspecified (892B, commit-note level); feature shipped in commit beba31b7c | Historical acceptance.md preserved here for provenance; future work should produce fresh SPEC |

## Archive Policy

- Archived SPECs MUST NOT be referenced by `depends_on:` in active SPECs
- If a future change reactivates an archived SPEC, restore from this directory and
  document the reason in the new SPEC's HISTORY block
- Audit-reports referencing archived SPECs may remain in `.moai/reports/plan-audit/`
  as historical record

## Audit Source

These archives were determined by `plan-auditor` agent run on 2026-04-24,
results in `.moai/reports/plan-audit/NON-V3R2-triage-audit-2026-04-24.md`.

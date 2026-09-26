# progress — SPEC-ROLE-NAMING-CODE-001

## §E.1 Plan-phase Audit-Ready Signal

- Card: t1256 · branch `WT-role-naming-code` · plan base `e62c3e183`
- Artifacts: spec.md, plan.md, acceptance.md, design.md, research.md (Tier L) + this file
- Evidence: `.moai/reports/t1256/census.md` (reproducible via `census.py`), `.moai/reports/t1256/conflicts.md`
- v0.1.0 (`6fe67c674`): REQ 20 · AC 20.
- v0.2.0 revision (2026-09-26): operator answers relayed by the leader applied — `lane` canonical, legacy spellings rejected (no aliases), persisted values write-new with the run-boundary rule, role-marker guard `lane`-only, lane self-dispatch help text, homonym qualifiers, `manager-lead` kept, t1193 demoted to a recorded dependency, new ordering t1242 → t1245 → t1256 run → t1240 → t1257 (t1193 excluded pending decision), design §3 zh note scoped to the code layer.
  - Rewritten REQ: 001, 003, 004, 005, 007, 009, 010, 011, 012, 013, 014, 016, 017, 018, 019, 020. New REQ: 021, 022, 023. Unchanged: 002, 006, 008, 015.
  - Rewritten AC: 002, 003, 004, 005, 007, 008, 009, 011, 013, 014, 015, 016, 018, 019, 020. New AC: 021, 022, 023. Unchanged: 001, 006, 010, 012, 017.
  - REQ count 23 · AC count 23 (AC-RNC-001..023, contiguous); every REQ maps to ≥1 AC (acceptance.md §C).
- v0.2.0 items to confirm at Kickoff: O0 derived `lead` rejection, O1 env names kept, O2 run-boundary persisted-data rule; O7 open. **Superseded by v0.3.0 below.**
- Plan-audit iteration 1: FAIL 0.77 (`.moai/reports/t1256/plan-audit-iter1.md`, audited at `d0770b9cc`) — MP-7 (O7 open) + blocking D2–D7, D9.
- v0.3.0 revision (2026-09-26):
  - Operator answers given directly in the lane window recorded in plan.md §B as RESOLVED: O7 rename role-sense Go identifiers; O0 `lead`/`lead-<suffix>` rejection confirmed; O1 env var names kept; O2 live legacy records refused with retire-and-relaunch guidance. No clarification marker remains in plan.md or research.md.
  - Rewritten REQ: 001 (D6), 003/005 (case variants, D16), 004 (`<n>` ≥ 1), 009 (detection carve-out), 011 (D18), 012 (D4 — equality assertion is the one coupling), 018 (D2/D5/D15), 019 (D1/D3), 022 (D7 — factory-run scope and membership basis). New REQ: 024 (D9 — legacy-only run retire), 025 (D7/D8 — kanban registry, legacy board role declaration, SessionStart session-record writer). REQ-number placement note added (D12).
  - Rewritten AC: 002, 003, 004, 005, 006, 008, 011, 012, 013 (D13 fixed string), 014, 016 (D5), 018, 019, 022, 023 (D14 fixed population command). New AC: 024, 025. Word-boundary rule stated once at the top of acceptance.md.
  - plan.md: §B all resolved; §C.3 sibling-SPEC re-check with plan-time branch @ SHA (D10); §E remeasure scope adds `./internal/spec` (plus `./internal/homestate`, `./internal/web`); M1 names the SessionStart session-record writer and the retire fallback (D8, D9); M6 records the partial supersession of SPEC-FACTORY-WORKER-NAMING-001 at sync (D11); R9 added.
  - design.md D2/D4/D7/D8 and §2/§4 updated; research.md §3 splits board role declarations from session records (D8).
  - Evidence hygiene: stray nested copy `.moai/reports/t1256/.moai/` removed (D17). No auditor helper scripts (`recount.py`, `sample.py`) found untracked in the worktree.
  - REQ count 25 · AC count 25 (AC-RNC-001..025, contiguous) — both at the Tier L ceiling of 25, neither over; every REQ maps to ≥1 AC (acceptance.md §C).
- Status: ready for plan-audit iteration 2

## §E.2 Run-phase Evidence

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_

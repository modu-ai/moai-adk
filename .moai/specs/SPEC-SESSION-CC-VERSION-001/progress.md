# SPEC-SESSION-CC-VERSION-001 — Progress

## §E.1 Plan-phase Audit-Ready Signal

- Artifacts authored (Tier M set): `spec.md`, `plan.md`, `acceptance.md`, plus this
  `progress.md`.
- Tier: M (justification in `plan.md` §B — ~13 files, 2 packages, no doctrine or docs-site
  sweep). Threshold 0.80.
- SPEC ID regex check executed as Bash, output `PASS`. ID uniqueness confirmed:
  `ls .moai/specs | grep -i "SESSION-CC-VERSION"` returns no match.
- Budget: **10 requirements, 10 acceptance criteria** (ceilings 16 / 16).
- Every absence-based criterion carries a pre-change baseline measured in this tree at
  `30ce3a02d` (branch `WT-session-cc-version`, cut from develop).
- Evidence base: `.moai/reports/t1348/verdict.md` items ② and ④, every code citation
  re-verified in this tree before use.
- `moai spec lint SPEC-SESSION-CC-VERSION-001` → `✓ No findings — all SPEC documents are
  valid`, exit 0 (after two fix rounds: AC→REQ mappings restated in the `maps REQ-…` house
  form, and `-run` selectors anchored `'^Test…$'`; the first measurement's 20 warnings
  included 2 hidden by a `tail`-truncated read — full output captured on re-run).
- Status: `draft`. Awaiting plan-audit.

### Card stage plan (lane-23, card t1465)

This plan phase is the card's first stage. The lane task list carries the remaining stages in
order: plan-audit (independent audit) → plan→run kickoff decision record → run (manager-develop,
M1-M5 of `plan.md` §D) → verification batch (env-scrubbed build + lint + targeted tests) →
sync (manager-docs) → card-review + merge-ready report to the leader. The lane orchestrator
commits these artifacts; nothing is committed or pushed by the plan phase.

### Assumptions made at plan phase

1. **Flag name `--cc-version`** on `moai session list` — a naming choice. The wire contract
   that matters (additive JSON fields under the flag, default path byte-identical) is pinned
   by REQ-SCV-005/006; renaming the flag before run is cheap.
2. **The doctor check is advisory (WARN-only, never `CheckFail`)** — in the `checkFlagSlot`
   pattern. An operator wanting staleness to gate `moai doctor`'s exit status would be a
   separate decision, not this SPEC's.
3. **`unknown` never warns** in the doctor check: staleness cannot be judged from unknown, and
   warning on it would nag npm-style installs with unversioned binary paths.
4. **The relaunch guard refuses the whole loop** rather than stripping the token or honoring it
   once: under `relaunch` a resume token cannot mean what it says (the loop leases new cards),
   so refusing with the safe form in the error text is the honest behavior.
5. **The pure assembler (REQ-SCV-008) changes no one-shot behavior** — the pass-through
   already survives the launcher mechanically (`spec.md` §A.3.1); the assembler names and
   tests what exists, and M4's only behavior change is the relaunch guard.
6. **Versions are never persisted** into the registry file — the live read at query time is
   the whole truth (`spec.md` §D, last exclusion).
7. **Installed-version read is path-parse only** — no `claude --version` exec (REQ-SCV-002);
   installs whose resolved path carries no version segment degrade to `unknown`.
8. **The exact spelling `moai cc -f lane-<n> -- --resume <id>` from the dispatch is refused
   today** (`factoryFlagUsageError`, `factory.go:92`); the SPEC treats the lane-join form as
   the emergency path and has the run-phase observe the refusal verbatim (`plan.md` §F.2).

## §E.2 Run-phase Evidence

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending run-phase>_

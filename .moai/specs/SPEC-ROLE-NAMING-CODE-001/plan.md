---
id: SPEC-ROLE-NAMING-CODE-001
title: "Plan — role naming unification (code + CLI)"
version: "0.1.0"
created: 2026-09-26
---

# Plan — SPEC-ROLE-NAMING-CODE-001

## §A Context

Card t1256 (Tier L, class C), layer A of a two-card rename. Inputs: census `.moai/reports/t1256/census.md`, conflict map `.moai/reports/t1256/conflicts.md`, `research.md`, `design.md`. Development mode per `.moai/config/sections/quality.yaml`; the run phase uses TDD for the new alias behavior (leader alias, `-f lane`, role-value set) and characterization tests before touching the existing label canonicalization.

## §B Open decisions for the Kickoff gate

Each marker must be resolved by the operator before Implementation Kickoff Approval. The SPEC text encodes the recommended option; a different answer amends the named REQ.

- **O0** [NEEDS CLARIFICATION: confirm the reversal of the t1085 worker-canonical contract] — `lane-<n>` was demoted to a legacy spelling on 2026-09-22 and restored-as-legacy on 2026-09-26 (`2a3af0c1e`). This SPEC promotes it back. Recommended: confirm, and record the confirmation in progress.md so the t1193 "unapproved reversal" ground does not reapply.
- **O1** [NEEDS CLARIFICATION: environment variable names — keep vs dual-name window] — Recommended: keep (REQ-RNC-011, design D3). Alternative: new `…_LEADER_…`/`…_LANE…` names written alongside the old for one minor release (design §5); this amends REQ-RNC-011 and forces a change to card t1242's eleven-key lists.
- **O2** [NEEDS CLARIFICATION: persisted role/slot values — write-current vs write-new] — Recommended: write-current / read-both (REQ-RNC-010, design D4). Alternative: write `leader`/`lane` and accept that a current-release binary cannot read post-change leader rows after a downgrade.
- **O3** [NEEDS CLARIFICATION: legacy-spelling lifetime] — Recommended: no removal date in this SPEC; a later card measures usage and removes. Alternative: name a removal release now.
- **O4** [NEEDS CLARIFICATION: amend card t1240 to `-f lane` / `MOAI_FACTORY_ROLE=lane` before it starts] — a queue edit, operator act. Without it, t1240 ships `-f agent` as a new surface that this SPEC immediately demotes.
- **O5** [NEEDS CLARIFICATION: t1245 ordering] — Recommended: t1245 lands with `worker` via its constant; this SPEC flips the constant to `lane` and widens the guard set (REQ-RNC-012). Alternative: t1245 adopts `lane` before its run (re-revision of its AC-AP-008/015/016/017).
- **O6** [NEEDS CLARIFICATION: t1193 disposition] — this SPEC cannot run its CLI milestones while t1193 edits the same six files (REQ-RNC-014). The leader decides land-or-drop for t1193 (and its conflict with t1242, which deletes the `moai codex -f` surface t1193 extends).
- **O7** [NEEDS CLARIFICATION: identifier rename scope] — Recommended: rename role-sense Go identifiers in the touched packages (REQ-RNC-019, design D8). Alternative: strings only, identifiers later.

## §C Pre-flight (run-phase entry)

1. Kickoff approval recorded with answers to O0-O7.
2. REQ-RNC-014 precondition read and recorded: t1242 deletion landed on develop; t1193 merged or closed.
3. Absorb local develop into the card worktree; re-measure the census (`python3 .moai/reports/t1256/census.py`) and record the new counts next to the plan-time counts.
4. Characterization tests pass on the absorbed tree for the packages the milestones touch (`go test ./internal/kanban/... ./internal/cli/... ./internal/hook/... ./internal/factorymsg/...` — scoped, never `./...`).

## §D Constraints

- Schema byte-identical (REQ-RNC-008); no data migration.
- Codex `env_vars` allowlist unchanged (REQ-RNC-011) — also card t1242 REQ-CFR-020.
- No edit under `internal/template/templates/**` (t1257).
- Unrelated senses untouched (REQ-RNC-016); `manager-lead` untouched (REQ-RNC-017).
- Lane-local verification only; the full suite runs in CI on the develop push.

## §E Self-verification

Each milestone closes with: scoped `go test` on the touched packages, `go vet` on them, `golangci-lint run` on them, and the AC rows for that milestone recorded in progress.md §E.2 with command and verbatim output.

## §F Milestones (ordered by change likelihood — highest first)

### M1 — Compatibility contract: aliases and persisted-key split (Priority High)
The decisions most likely to be revised at Kickoff (O1, O2, O5) land here first, as tests before code.
- Split the leader label value (`leader`) from the leader persisted key (`lead`); every reader of the persisted key resolves both.
- Read-time role resolution for persisted values in both vocabularies (REQ-RNC-009, -010); broker delivery for `leader` and all three lane spellings (REQ-RNC-013).
- Leader name namespace shared by `lead`/`leader` (REQ-RNC-007).
- Schema-freeze guard (REQ-RNC-008) and env-name/allowlist freeze guard (REQ-RNC-011).
- ACs: AC-RNC-006, -007, -008, -009, -010, -011.

### M2 — CLI surface: tokens, labels, help, hints (Priority High)
- Canonical prefix swap: `lane-<n>` produced; `worker-<n>`/`agent-<n>` readable with shared numbering (REQ-RNC-004, -005).
- `-f lane` canonical; `-f worker`/`-f agent` aliased with one hint (REQ-RNC-002, -003).
- Usage error, `Use`/help strings for `cc`/`glm` (and `codex` if still present after t1242), collision and bump messages, `moai tokens --role` example, `gtd answer` short text, board write-guard error text (REQ-RNC-001).
- ACs: AC-RNC-001, -002, -003, -004, -005, -016.

### M3 — Notices, locales, dashboard (Priority Medium)
- Factory and kanban SessionStart notices in en/ko/ja/zh per design §3 (REQ-RNC-015).
- Doctor, statusline, and web display labels read `leader`/`lane` while matching persisted keys (REQ-RNC-001).
- ACs: AC-RNC-012, -013.

### M4 — Factory role marker value (Priority Medium, conditional)
- Where t1245 has landed: flip the lane value to `lane`, widen the guard set to {lane, worker, agent}, and prove stamp site and guard share one definition (REQ-RNC-012).
- Where t1245 has not landed: record the value `lane` in progress.md for the leader to hand to t1245; no edit.
- ACs: AC-RNC-014.

### M5 — Mechanical rename and guards (Priority Low)
- Role-sense Go identifiers and comments (REQ-RNC-019), with persisted-key definitions commented as on-disk keys.
- Test fixtures to the new vocabulary, keeping one legacy test per spelling (REQ-RNC-020).
- Vocabulary guard test (REQ-RNC-018) and unrelated-sense diff check (REQ-RNC-016, -017).
- ACs: AC-RNC-015, -017, -018, -019, -020.

## §G Risks

| # | Risk | Mitigation |
|---|---|---|
| R1 | The leader label and the persisted leader key are one constant today; splitting it can make the board write guard or role-declaration reader stop recognizing the leader (sole-writer refusal on every board write) | M1 first, characterization tests on the board guard and role reader before the split; AC-RNC-008 exercises a `lead` role declaration against the new code |
| R2 | Stamp/guard value drift on the role marker → guard silently denies nothing | REQ-RNC-012 one-definition rule; AC-RNC-014 asserts a deny for each of the three values |
| R3 | Merge conflicts with t1193 in six shared files | REQ-RNC-014 halt; conflicts.md ordering |
| R4 | Mid-run binary reinstall: sessions launched as `lead`/`worker-3` meet new hooks | Read-both (REQ-RNC-009) and shared namespaces (REQ-RNC-004, -007); AC-RNC-007, -009 |
| R5 | Guard test allowlist grows into a loophole | Allowlist limited to legacy parse values and hint text (REQ-RNC-018); AC-RNC-018 counts allowlist entries |
| R6 | Heuristic census undercounts multi-line user-facing strings | Pre-flight re-measure plus the guard test, which reads compiled string constants rather than lines |
| R7 | Operator retraining churn (third change in five days) | O0 explicit confirmation; hints name the new form |

## §H Cross-references

- `.moai/reports/t1256/census.md`, `.moai/reports/t1256/conflicts.md`
- SPEC-FACTORY-WORKER-NAMING-001 (the contract this reverses), SPEC-CODEX-FACTORY-RETIRE-001 (t1242), SPEC-AUTONOMY-PRECONDITION-001 (t1245)
- Sibling card t1257 consumes `design.md` §3.

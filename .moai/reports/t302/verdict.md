# t302 verdict — sync-audit-4dim verdict-ownership contradiction resolved

Card: t302 (G2a, lane-9) · Branch: WT-audit-verdict-owner · Base: local develop e85c55fa9 (absorbed via fast-forward from f7cabfc29)

## Claim

The `description` metadata of `.claude/workflows/sync-audit-4dim.js` no longer contradicts the file's own header comment and the caller contract (sync.md § Binding promotion): it now states the A3-promoted ownership — the workflow's harmonic-mean verdict IS the BINDING sync-phase verdict owner on the happy path, and the cold `sync-auditor` subagent is the FALLBACK verdict owner on any trigger.

## Evidence

1. The contradiction, measured (pre-fix): the same file stated ownership two ways —
   - header comment (lines 4-9): "PROMOTED its verdict to BINDING on the happy path ... The cold auditor remains the FALLBACK verdict owner";
   - `description` metadata (line 53, the skill-listing-consumed surface — observed carrying exactly this stale text in this session's own skill listing at session start): "execution vehicle, NOT the binding sync-auditor verdict owner".
2. Caller truth-anchor: `.claude/skills/moai/workflows/sync.md` § Binding promotion (SPEC-AUDIT-SNAPSHOT-001 A3) — happy path → workflow verdict BINDING, cold sync-auditor NOT spawned; fallback triggers (INCOMPLETE / zero_scored / contested finding) → cold sync-auditor spawns as fallback binding-verdict owner. The header agreed; only the description disagreed.
3. Fix (verbatim diff: `.moai/reports/t302/description-change.diff`, both trees — template source first per Template-First, then local mirror; the description line carries no internal tokens so the copies stay in their established sanitize relationship): the stale clause is replaced by "BINDING sync-phase verdict owner on the happy path (PASS, no dim 0, not INCOMPLETE, no contested finding — IsBinding), cold sync-auditor subagent is the fallback verdict owner otherwise".
4. Post-fix ownership agreement across the three in-file/sibling surfaces: header, description, and sync.md now state the same ownership (grep observed: both copies line 53 carry the new clause; "NOT the binding" phrase count in the file = 0).
5. `node --check .claude/workflows/sync-audit-4dim.js` → SYNTAX_OK (edit is a single-line metadata string; workflow still parses).
6. `make build` → EXIT=0 (embedded FS recompiled; catalog.yaml unchanged).

## Baseline-attribution

All measurements this run, this tree (WT-audit-verdict-owner @ e85c55fa9):
- contradiction phrases: `grep -n "NOT the binding\|BINDING" <both copies>` → post-fix shows the new clause and no "NOT the binding" occurrence in either copy.
- diff: `git diff -- <both copies>` → 26 lines (description-change.diff).
- syntax: `node --check` → exit 0.
- build: `make build` → exit 0.

## Gaps (explicitly NOT observed)

- A live re-listing of the skill registry showing the NEW description was not observed (the listing is composed at session start; the correction is proven in the file, not in a fresh listing render).
- The workflow was not executed end-to-end (a full 5-agent run costs a Workflow approval and is not needed to prove a metadata wording change).
- `node --test .claude/workflows/tests/sync-audit-4dim.test.js` was run and FAILED — see finding below; pre-existing at base, unrelated to this card's edit.

## Finding (reported, not fixed — outside t302's scope)

**Pre-existing red, silently disarmed guard**: `node --test .claude/workflows/tests/sync-audit-4dim.test.js` fails with `VERDICT PURE FUNCTIONS START marker not found in workflow source`. Measured pre-existing: `git show e85c55fa9:.claude/workflows/sync-audit-4dim.js | grep -c "VERDICT PURE FUNCTIONS START"` → 0. History: the sentinel pair was ADDED by a098aac98 (SPEC-FOURDIM-PHANTOM-001 IMP-5 phantom-mechanism guard) and REMOVED by 0de8517e5 ("chore: accumulated working-tree sweep"), leaving the guard test permanently red since that sweep. The test is wired to NO gate (no Makefile target, no CI workflow, no Go wrapper references it — grep observed zero hits), so the red has been silent. This is a verification-completeness §1.3 continued-firing defect (the IMP-5 phantom-mechanism guard stopped firing when its subject block was swept). Recommend the lead queue a card: restore the sentinel pair around the current pure-verdict functions and wire the test into a runnable gate, or retire the test deliberately with a record.

## Residual-risk

- The description is advisory text for the orchestrator reading the listing; the BINDING predicate itself is mechanical (`internal/runtime.FourDimVerdict.IsBinding()`) and unaffected by wording. A reader acting on the description alone cannot change gate behavior — the risk this card closes is decision-confusion, not mechanism drift.
- If lane-10's t386 lands first and touches sync-auditor.md's export clause, merge-order interaction with this card is nil (different files); no conflict anticipated on the description line.

# Card t1510 — self-improvement protected zone (verdict record)

## Dispatch record

- Card: t1510 (operator decision D5), class C (plan → run → sync). SPEC: SPEC-SELF-IMPROVE-PROTECTED-ZONE-001.
- Worktree: `.claude/worktrees/t1510`, branch `WT-self-improve-protected-zone`, base `e497f6936` (= local `develop` tip at dispatch; the tree was first created at `2f492df19` with no commits and fast-forwarded to the base).
- **Lease: none** — as with the preceding card on this lane, run under **leader direct dispatch** (no serial-slot lease; no `moai factory complete` expected, per the dispatch).
- Stop rule from the dispatch: keep scope small (declaration + blocking + removal/lowering/deletion proposals routed to a human); plan-audit ceiling reached → report; even on PASS, run starts only after the leader replies.

## Plan phase

- manager-spec authored the SPEC (draft, Tier M): spec.md (16 REQ), plan.md (3 milestones), acceptance.md (13 AC: 5 release-blocking, 8 regression-guard), progress.md, decision-index.md (6 questions), evidence/ (probe, judge, base-tree outputs, latency baseline).
- `moai spec lint --strict SPEC-SELF-IMPROVE-PROTECTED-ZONE-001` → `✓ No findings — all SPEC documents are valid` (re-run by the lane in this tree).
- Plan-audit: pending (record below).

## Plan-audit record (Tier M ceiling = 2, `plan_audit_tier_ceilings.M`)

| Iter | Audited SHA | Verdict | Score | Report |
|---|---|---|---|---|
| 1 | `c71df8f4b` | FAIL | 0.74 | `plan-audit-verdict.md` (D1-D19; D1-D12 blocking) |
| 2 | `24d79bb64` | FAIL | 0.78 (threshold 0.80) | `plan-audit-verdict-iter2.md` (17 of 19 resolved, D19 partial, D15 optional/open; new D20-D28) |

Ceiling reached. Per the dispatch note, no further repair was started; the lane reports instead.

Open blocking defects (iteration 2):
- **D20 (major)** — the shipped manifest lists `**/CLAUDE.local.md` (spec.md:L144); the template-neutrality audit rejects that literal in any scanned `.yaml`, so AC-SIPZ-009 / REQ-SIPZ-003 cannot pass. Fix: move the entry to the dogfood overlay.
- **D21 (major)** — the repair made a shipped manifest with fewer than the seven required categories invalid, yet probe fixtures MS4/MS7/MS8 put a one-category file at the shipped path while AC-SIPZ-004 expects `category=probe_docs` / `category=probe_base`. Fix: give those fixtures all seven categories, or move the rule out of the runtime loader.
- **D23 (major)** — REQ-SIPZ-012 and §C.7 promise human routing for every manifest-listed path, but baseline-matched denials (auditor definitions, gate-policy rules, hooks) keep the legacy reason with no `route=human`. Fix: state the real scope or add the routing fields.
- D22, D24-D26 minor (AC-004 E3 green path names M1/M2 but only M3 can deny it; plan.md:L9/L61 still claim the hook-block round is the route; short `AC-001`-style references make the repo's AC counter read 25 vs 13; REQ-SIPZ-009 "failing file" clause has no assertion). D27/D28 optional.

`plan_audit_ceiling_policy` allows one delta round (`auto_delta_rounds: 1`); it was not taken without the leader's call. Decision owed to the leader: (A) run that delta round (fix D20, D21, D23 plus the touched minors, delta re-audit; leader runs `codex_audit` for a cross-model receipt) or (B) hold-and-split. Run has NOT started.

Gaps: no cross-backend audit (receipts=none); latency not re-measured; Go tests do not exist at plan phase.

# SPEC-LEAD-AUTOPUSH-001 — Implementation Plan

> Card t1346 · Tier M · docs-surface only · 2026-09-29 · manager-spec

## §A Context

Operator directive (2026-09-29): make the lead's develop push discipline durable — a
20-commit batch-close threshold plus a green-conditional — instead of today's ad-hoc
thresholds (operator-stated: 95/26/28/22 across four pushes) and memory-held practice.
Measured landscape (tree `51abf337a`): the threshold config key already exists
(SPEC-MAIN-COMMIT-BAN-001 / card t1337); the operational lead-facing doc
(`.claude/rules/local/gitflow-lane-protocol.md`) carries neither the threshold trigger nor a
green-conditional; the factory surface already has a per-card push gate; no goal template and
no batch-push verb exist.

## §B Known Issues

- The lane protocol §7 currently words batch-close timing as bare lead discretion — the
  exact gap this SPEC closes.
- CI in §4/§8 appears only as the post-push verdict; nothing holds the NEXT push when the
  last push went red.
- `LeadPushThreshold` has zero Go consumers (types.go declaration + defaults.go only) —
  intentional per this SPEC's Out of Scope; recorded here so a future reader does not
  "fix" it as dead config.

## §C Design Decisions (the surface judgment the card demanded)

**D1 — Surface: docs-only (chosen).** Measured against the three candidates:

1. *Docs-only* — the operating rule's consumers are the lead session and the lane sessions;
   both read the local rule files. The threshold half is already config+doctrine (t1337);
   the green-conditional half is pure wording in the doc the lead already consults for push
   duty. Two files, one cross-ref, zero code. CHEAPEST and matches where the rule lived
   when it worked (the lead's head).
2. *Goal-condition* — REJECTED. `/moai goal` is an arm-only continuation condition evaluated
   at turn end to keep a session working toward completion. A batch-close threshold is a
   periodic operational trigger the lead evaluates at dispatch-cycle points; a goal armed on
   "unpushed count ≥ 20" would invert the semantics (the condition HOLDS only when the
   trigger fires — between triggers the goal would block turn-end and spin idle turns).
   The ecosystem also carries a HARD scheduled-runs-never-push invariant
   (goal-directive-detail.md → cadence-bridge). Wiring push into a goal loop would fight
   both.
3. *CLI verb (`moai factory push`)* — REJECTED for this SPEC. The card's own warning holds
   under measurement: the push edge is deliberately kept with the lead's judgment (§7
   "배치를 닫을 시점을 리더가 판단한다" — the discretion is the feature; the threshold makes
   it deterministic, not the verb). The factory surface already carries the per-card
   `decide --gate push` landing gate. A batch verb is new shared-edge code with no measured
   consumer. Deferred: if a future card shows the lead mis-executing the manual sequence
   repeatedly, a verb becomes evidence-backed then.

**D2 — Threshold carrier: config (already landed).** `git_strategy.manual.lead_push_threshold`
(local 20, template 0=disabled) is the single source of the number; docs name the key + the
count command and never restate the value. This answers the card's question (b) as
"config key — settled by t1337; the docs increment is the pointer, not the number".

**D3 — Green-conditional semantics: codify the practiced gate.** Two measured layers:
(i) per-card: the integration-window re-measurement of the merged tree is the pre-push gate —
it is what actually caught the three reds (card evidence ②); (ii) batch-level: when the last
push's CI on `origin/develop` is red, the next push is withheld until repaired. This is
REQ-002/REQ-003 — the actual practice, not an idealized "test the tip locally" rule (§8
already forbids local full-suite runs).

**D4 — Open decision (default: proceed).** Adding the green-conditional cross-ref sentence to
`AGENTS.local.md` §4.1 item 6 is the only edit outside the lane protocol. Default: yes —
item 6 is where the t1337 threshold doctrine lives and leaders read it; a dangling threshold
doctrine with no green-conditional pointer would half-document the rule. Reversible
one-line revert if the operator prefers the lane protocol to stay the sole home.
All design decisions D1-D4 are resolved; no clarification is outstanding.

## §D Constraints

- Local-only targets: `.claude/rules/local/` and `AGENTS.local.md` carry NO template mirror
  (intentional; Template-First does not apply). Both are tracked — normal card-worktree
  commit → develop merge.
- Wording language: the lane protocol is written in Korean (its existing register); edits
  match the file's existing style, not a translation register.
- No Go code, no config value changes, no template changes, no hook changes.

## §E Self-Verification (plan-level)

- RED-now cells for every release-blocking AC measured on tree `51abf337a` before any edit
  (ledger in acceptance.md §D.1).
- Freeze checks re-run directly post-edit (lanes-never-push wording; landing-verification
  wording) — freeze claims get direct re-execution, not inference.
- Positive control for every absence claim: the config key's presence grep (E-3) proves the
  docs' referenced key actually exists, so the docs' zero-hits are absence of the RULE, not
  absence of the REFERENT.

## §F Milestones (decision-reversibility order — semantic wording first, mechanical last)

- **M1 (High)** — `.claude/rules/local/gitflow-lane-protocol.md` §4: add the threshold
  trigger sentence (key + count command named, value delegated to config), the
  green-conditional paragraph (window re-measurement gate + last-push CI hold), and the
  disabled (0) fallback sentence. The semantic core of the SPEC; wording decisions live here.
- **M2 (Medium)** — same file §7: replace the bare-discretion clause in the batch bullet
  with the threshold trigger + green-conditional references (§4 as the detail home; §7
  points, not duplicates).
- **M3 (Low)** — `AGENTS.local.md` §4.1 item 6: append one cross-reference sentence to the
  lane protocol's green-conditional (D4 default), AND delete the word-restatement
  `(초기값 20)` beside the key (audit D2 — the numeric SSOT is the config file; the prose
  carries the key and the count command only).
- **M4 (Low, mechanical)** — verification sweep: re-run RED greps flipped green, re-run the
  two freeze checks, refresh progress.md §E.1 → §E.2 handoff, evidence to
  `.moai/reports/t1346/`.

## §G Anti-Patterns

- Do NOT restate "20" anywhere in the docs — a duplicated number forks the source of truth
  the next time the operator retunes it.
- Do NOT "improve" §5 (충돌), §6, §8-§11 while in the file — scope discipline; the freeze
  ACs exist to catch exactly this.
- Do NOT add a template mirror for the edited local files.
- Do NOT translate the Korean rule file's voice into English or calqued prose.

## §H Cross-References

- `../spec.md` §2 (REQ-001..005), §3 (Out of Scope), §6
- `acceptance.md` §D (AC matrix + RED-now ledger)
- SPEC-LANE-PUSH-BATCH-001, SPEC-MAIN-COMMIT-BAN-001 (related, landed)

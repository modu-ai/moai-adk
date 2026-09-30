# acceptance.md — SPEC-LANE-STALL-WATCHDOG-001

## HISTORY

- 0.1.0 — 2026-09-30 — plan-phase authoring (card t1370, worktree
  `.moai/worktrees/t1370`, branch `WT-lane-stall-watchdog`, HEAD `3dd5adf2f`).

## §A Verification model

- **Two-cell discipline** (`.claude/rules/moai/development/verification-completeness.md`
  §2): every release-blocking criterion below carries a RED-now cell observed
  on the pre-implementation tree and a green-path cell naming the milestone
  that flips it. The RED reason is stated in each cell — the dominant RED root
  for the two new artifacts is their measured absence (file does not exist),
  observed once per file and cited by every criterion that shares it.
- **Document-level tree pin**: `3dd5adf2f` (branch `WT-lane-stall-watchdog`,
  measured 2026-09-30). Binds every criterion carrying no pin of its own.
- **Carrier**: the evidence ledger (§E) holds command + verbatim stdout +
  exit code; table cells cite ledger ids (`verification-completeness.md`
  §2.1 — the ledger is the recommended carrier; table cells mangle shell
  metacharacters).
- **Regression-guard criteria** (AC-LSW-011, AC-LSW-012) carry no RED cell:
  they are invariants that must HOLD at plan close and stay holding through
  sync close. AC-LSW-011 is evaluated PRE-INTEGRATION only (merge-base form;
  post-merge the range empties and the criterion is vacuous — the post-merge
  evidence is the tree-identity argument in the integration window).

## §D AC Matrix

| AC id | REQ | Release-blocking | RED-now (ledger) | Green path |
|---|---|---|---|---|
| AC-LSW-001 | REQ-LSW-007 (doc exists) | YES | E-1 | M1 |
| AC-LSW-002 | REQ-LSW-007 + REQ-LSW-009 (definition + Jev line) | YES | E-2 | M1 |
| AC-LSW-003 | REQ-LSW-002 (skill exists) | YES | E-3 | M2 |
| AC-LSW-004a | REQ-LSW-003 (awaited-actor remedy) | YES | E-4 | M2 |
| AC-LSW-004b | REQ-LSW-004 (blocked-by remedy) | YES | E-4 | M2 |
| AC-LSW-004c | REQ-LSW-005 (shell-error remedy) | YES | E-4 | M2 |
| AC-LSW-004d | REQ-LSW-006 (accidental-stop remedy) | YES | E-4 | M2 |
| AC-LSW-005 | REQ-LSW-001 (three-channel progress definition) | YES | E-5 | M2 |
| AC-LSW-006 | REQ-LSW-004 (t1343 composition + queue-readonly) | YES | E-6 | M2 |
| AC-LSW-007 | REQ-LSW-008 (dispatch explicit-wait) | YES | E-7 | M3 |
| AC-LSW-008 | REQ-LSW-008 + REQ-LSW-002 (dispatch watchdog pointer) | YES | E-8 | M3 |
| AC-LSW-009 | REQ-LSW-008 (lane-protocol explicit wait) | YES | E-9 | M3 |
| AC-LSW-010a | mirror parity — skill | YES | E-10 | M4 |
| AC-LSW-010b | mirror parity — doctrine doc | YES | E-10 | M4 |
| AC-LSW-011 | scope guard — zero Go changes | no (regression-guard) | holds at plan close | guard through close |
| AC-LSW-012 | spec lint clean | no (regression-guard) | holds at plan close | guard through close |

### §D.1 Severity

All YES rows are release-blocking: the card's deliverable IS the doctrine and
its executable carrier; a missing remedy row or a drifted mirror is a shipped
defect of exactly the kind the card exists to close. The two guard rows gate
the same close but observe invariants rather than flips.

### §D.2 Traceability

REQ-LSW-001 → AC-LSW-005 · REQ-LSW-002 → AC-LSW-003, AC-LSW-008 ·
REQ-LSW-003 → AC-LSW-004a · REQ-LSW-004 → AC-LSW-004b, AC-LSW-006 ·
REQ-LSW-005 → AC-LSW-004c · REQ-LSW-006 → AC-LSW-004d · REQ-LSW-007 →
AC-LSW-001, AC-LSW-002 · REQ-LSW-008 → AC-LSW-007, AC-LSW-008, AC-LSW-009 ·
REQ-LSW-009 → AC-LSW-002. All 9 REQs covered; no AC orphans.

### §D.3 Indirect verification

AC-LSW-004a-d and AC-LSW-005/006 verify CONTENT by token presence; the
behavioral truth (a lane actually resuming from evidence) is verified
indirectly by M4's review pass: the skill text walked against spec.md §B.2
line by line, and the awaken prompt exercised once in a dry read-through
recorded in progress.md §E.2. A model-mediated procedure cannot be
unit-tested; the review pass is its verification instrument and is recorded
as evidence, not assumed.

### §D.4 Closure gates

Sync close requires: all 14 YES rows green on the closing tree; AC-LSW-011
still empty; AC-LSW-012 exit 0; `make build` exit 0; template neutrality
scoped test green; progress.md §E.2/§E.3 populated with the §D.3 review-pass
record.

### §D.5 Forward-looking checks

- The awaken prompt's canonical form lives in the unified doc; if the skill
  is renamed or moved, both the doc pointer and the dispatch pointer
  (AC-LSW-008) must move with it — a future rename without those two greps
  going red is a silent liveness loss (verification-completeness §1.3).
- AC-LSW-011's merge-base form decays post-merge by design; the closing
  evidence is recorded pre-integration in progress.md §E.2.

### §D.6 Edge cases the doctrine must survive (review-pass checklist)

- A stall where evidence mtime advances but no commit lands (docs-only
  progress) → NOT a stall (channel b alone satisfies progress).
- A blocked-by whose predecessor was dropped (t1343 §G R-1's wedge) → the
  remedy records the explicit wait and names `todo unrelate` as the operator
  escape; the lane does NOT unrelate by itself (queue mutation).
- A shell error that is determinate (e.g. exit 128 on a missing remote) →
  zero retries, straight to structured blocker with the error text.
- An awaken during an operator-owned gate wait → the re-check reads the gate
  surface; finding no change, it re-records the wait and yields — never
  prompts, never answers.

### §D.7 Definition of Done

Card t1370 closes when: the two artifacts exist locally AND in templates
byte-identical, the two doctrine edits are mirrored (or correctly local-only),
all 14 release-blocking ACs are green on the closing tree, the two guards
hold, spec lint is clean, and the commit chain on `WT-lane-stall-watchdog`
carries the card marker.

## §E Evidence Ledger (RED-now cells, tree pin `3dd5adf2f`, measured 2026-09-30)

- **E-1** — `test -f .claude/rules/moai/workflow/auto-semantics.md` →
  stdout: (empty) · exit 1. RED reason: the unified doc does not exist yet.
  Green: M1 flips to exit 0.
- **E-2** — `grep -c "display-only" .claude/rules/moai/workflow/auto-semantics.md`
  → stdout: `grep: .claude/rules/moai/workflow/auto-semantics.md: No such
  file or directory` class (file-absence; instrument cannot run) · exit 2.
  RED reason: same measured absence as E-1. Green: M1 flips to count ≥ 1,
  exit 0.
- **E-3** — `test -f .claude/skills/moai-lane-watchdog/SKILL.md` →
  stdout: (empty) · exit 1. RED reason: the skill does not exist yet.
  Green: M2 flips to exit 0.
- **E-4** — remedy-label probes on the (absent) skill file. Observed on this
  tree, verbatim:
  `grep -c "blocked-by" .claude/skills/moai-lane-watchdog/SKILL.md` →
  `ugrep: warning: .claude/skills/moai-lane-watchdog/SKILL.md: No such file
  or directory` · exit 2. The remaining three probes (awaited-actor /
  shell-error / accidental-stop tokens, chosen at run-phase authoring from
  spec.md §B.2's labels) share the same observed RED root — the file
  absence measured here — and each flips green in M2 when its label count ≥ 1.
- **E-5** — `grep -cE "integration[- ]window" .claude/skills/moai-lane-watchdog/SKILL.md`
  → file-absence class · exit 2 (same measured absence as E-3). Green: M2
  flips to count ≥ 1, exit 0.
- **E-6** — `grep -c "SPEC-RELATION-PICKUP-FILTER-001" .claude/skills/moai-lane-watchdog/SKILL.md`
  → `ugrep: warning: .claude/skills/moai-lane-watchdog/SKILL.md: No such file
  or directory` · exit 2 (verbatim observed). Green: M2 flips to count ≥ 1,
  exit 0.
- **E-7** — `grep -c "explicit wait" .claude/rules/moai/workflow/kanban-dispatch.md`
  → stdout: `0` · exit 1. RED reason: the file exists (mirror-managed, this
  tree) and carries no explicit-wait posture — a measured zero, not an
  absence. Green: M3 flips to count ≥ 1, exit 0.
- **E-8** — `grep -c "moai-lane-watchdog" .claude/rules/moai/workflow/kanban-dispatch.md`
  → stdout: `0` · exit 1. Green: M3 flips to count ≥ 1, exit 0.
- **E-9** — `grep -c "explicit wait" .claude/rules/local/gitflow-lane-protocol.md`
  → stdout: `0` · exit 1. Green: M3 flips to count ≥ 1, exit 0.
- **E-10** — `diff -r .claude/skills/moai-lane-watchdog internal/template/templates/.claude/skills/moai-lane-watchdog`
  → stdout: `diff: .claude/skills/moai-lane-watchdog: No such file or
  directory` · exit 2 (verbatim observed). AC-LSW-010b's probe (the doctrine
  doc pair) shares the same absence root. Green: M4 flips both to
  exit 0 (byte-identical mirrors).

## §E.2 Quality gates (non-AC, plan-phase completion)

- `go run ./cmd/moai spec lint SPEC-LANE-STALL-WATCHDOG-001` → exit 0, no
  findings (run at plan close; recorded in progress.md §E.1).
- Frontmatter: canonical 12 fields validated against
  `spec-frontmatter-schema.md` § Canonical 12 Required Fields; no snake_case
  aliases; SPEC-ID regex check executed as Bash with verbatim `PASS` output
  (plan-phase, this tree).

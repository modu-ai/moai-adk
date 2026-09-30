# acceptance.md — SPEC-LANE-STALL-WATCHDOG-001

## HISTORY

- 0.1.0 — 2026-09-30 — plan-phase authoring (card t1370, worktree
  `.moai/worktrees/t1370`, branch `WT-lane-stall-watchdog`, HEAD `3dd5adf2f`).
- 0.2.0 — 2026-09-30 — iter-2 revision: AC set re-scoped to the four operator
  directives (ladder, harness neutrality, gate inventory, jev_ask); new
  criteria AC-LSW-013..017; AC-LSW-007/008 consolidated into 007a/b; new RED
  observations pinned at the revision tree.

## §A Verification model

- **Two-cell discipline** (`.claude/rules/moai/development/verification-completeness.md`
  §2): every release-blocking criterion carries a RED-now cell observed on the
  pre-implementation tree and a green-path cell naming the milestone that
  flips it. The RED reason is stated in each cell — the dominant RED root for
  the two new artifacts is their measured ABSENCE (file does not exist),
  observed at `3dd5adf2f` (0.1.0) and re-observed at the revision tree.
- **Document-level tree pins**: `3dd5adf2f` for the 0.1.0 cell set;
  `569202288` for the iter-2 cells (both on branch `WT-lane-stall-watchdog`,
  measured 2026-09-30; the artifacts are the only delta between the pins, so
  the target-file absence is valid at both). Binds every criterion carrying
  no pin of its own.
- **Carrier**: the evidence ledger (§E) holds command + verbatim stdout +
  exit code; table cells cite ledger ids.
- **Regression-guard criteria** (AC-LSW-011, AC-LSW-012) carry no RED cell:
  invariants that hold at plan close and must stay holding through sync close.
  AC-LSW-011 is evaluated PRE-INTEGRATION only (merge-base form; post-merge
  the range empties — the closing evidence is recorded pre-integration in
  progress.md §E.2).
- **Reported-premise criteria**: the codex-side facts in AC-LSW-013/017
  verify DOCUMENT CONTENT (the table and its stated premises), not the
  external Codex tool surface itself — the external facts are lead-reported
  (spec.md §G R-4) and re-checked at M4.

## §D AC Matrix

| AC id | REQ | Release-blocking | RED-now (ledger) | Green path |
|---|---|---|---|---|
| AC-LSW-001a | REQ-LSW-007 (doc exists) | YES | E-1 | M1 |
| AC-LSW-001b | REQ-LSW-013 (amendment list present in spec.md §A.7) | YES | E-1b | iter-2 authoring (flipped at plan close; guard through close) |
| AC-LSW-002 | REQ-LSW-007 + REQ-LSW-009 (definition + Jev/jev_ask line) | YES | E-2 | M1 |
| AC-LSW-003 | REQ-LSW-002 (skill exists) | YES | E-3 | M2 |
| AC-LSW-004a | REQ-LSW-003 (awaited-judgment ladder) | YES | E-4 | M2 |
| AC-LSW-004b | REQ-LSW-004 (blocked-by remedy) | YES | E-4 | M2 |
| AC-LSW-004c | REQ-LSW-005 (shell-error remedy) | YES | E-4 | M2 |
| AC-LSW-004d | REQ-LSW-006 (accidental-stop remedy) | YES | E-4 | M2 |
| AC-LSW-005 | REQ-LSW-001 (three-channel progress definition) | YES | E-5 | M2 |
| AC-LSW-006 | REQ-LSW-004 (t1343 composition + queue-readonly) | YES | E-6 | M2 |
| AC-LSW-007a | REQ-LSW-008 (dispatch explicit-wait posture) | YES | E-7 | M3 |
| AC-LSW-007b | REQ-LSW-002 + REQ-LSW-008 (dispatch watchdog pointer) | YES | E-8 | M3 |
| AC-LSW-009 | REQ-LSW-008 (lane-protocol explicit wait) | YES | E-9 | M3 |
| AC-LSW-010a | mirror parity — skill | YES | E-10 | M4 |
| AC-LSW-010b | mirror parity — doctrine doc | YES | E-10 | M4 |
| AC-LSW-011 | scope guard — zero Go changes | no (regression-guard) | holds at plan close | guard through close |
| AC-LSW-012 | spec lint clean | no (regression-guard) | holds at plan close | guard through close |
| AC-LSW-013a | REQ-LSW-010 (neutrality table in doc — per-runner rows incl. codex) | YES | E-11 | M1 |
| AC-LSW-013b | REQ-LSW-010 (codex ⑤ broker substitute named in doc) | YES | E-11 | M1 |
| AC-LSW-014 | REQ-LSW-008 (gate inventory + keep-set tokens in doc) | YES | E-12 | M1 |
| AC-LSW-015 | REQ-LSW-012 (decision-record format in doc) | YES | E-13 | M1 |
| AC-LSW-016a | REQ-LSW-009 (jev_ask single surface in doc, gate named) | YES | E-14 | M1 |
| AC-LSW-016b | REQ-LSW-009 (jev_ask in skill) | YES | E-14 | M2 |
| AC-LSW-017a | REQ-LSW-011 (codex update_plan view in skill) | YES | E-15 | M2 |
| AC-LSW-017b | REQ-LSW-011 (Claude task-view token in skill) | YES | E-15 | M2 |

### §D.1 Severity

All YES rows are release-blocking: the card's deliverable IS the doctrine and
its executable carrier; a missing ladder step, an un-carried neutrality row,
an uninventoried gate, or a drifted mirror is a shipped defect of exactly the
kind the card exists to close. The two guard rows gate the same close but
observe invariants rather than flips. AC-LSW-001b is flipped by the iter-2
authoring itself and thereafter guarded (its content is the amendment list
M3 applies).

### §D.2 Traceability

REQ-LSW-001 → AC-LSW-005 · REQ-LSW-002 → AC-LSW-003, AC-LSW-007b ·
REQ-LSW-003 → AC-LSW-004a, AC-LSW-013a · REQ-LSW-004 → AC-LSW-004b,
AC-LSW-006 · REQ-LSW-005 → AC-LSW-004c · REQ-LSW-006 → AC-LSW-004d ·
REQ-LSW-007 → AC-LSW-001a, AC-LSW-002 · REQ-LSW-008 → AC-LSW-007a,
AC-LSW-014 · REQ-LSW-009 → AC-LSW-002, AC-LSW-016a, AC-LSW-016b ·
REQ-LSW-010 → AC-LSW-013a, AC-LSW-013b · REQ-LSW-011 → AC-LSW-017a,
AC-LSW-017b · REQ-LSW-012 → AC-LSW-015 · REQ-LSW-013 → AC-LSW-001b.
All 13 REQs covered; no AC orphans.

### §D.3 Indirect verification

AC-LSW-004a-d and AC-LSW-005/006 verify CONTENT by token presence; the
behavioral truth (a lane actually resolving via the ladder) is verified
indirectly by M4's review pass: the skill text walked against spec.md §B.2
step by step, the ladder exercised once in a dry read-through per runner
surface (Claude path and codex path — the codex walk uses the §B.6 table's
substitute rows), recorded in progress.md §E.2. A model-mediated procedure
cannot be unit-tested; the review pass is its verification instrument and is
recorded as evidence, not assumed.

### §D.4 Closure gates

Sync close requires: all YES rows green on the closing tree; AC-LSW-011 still
empty; AC-LSW-012 exit 0; `make build` exit 0; template neutrality scoped
test green; the §A.7 amendment sweep (plan.md E9) shows every row applied;
progress.md §E.2/§E.3 populated with the §D.3 review-pass records.

### §D.5 Forward-looking checks

- The awaken prompt's canonical form lives in the unified doc; if the skill
  is renamed or moved, the doc pointer and the dispatch pointer (AC-LSW-007b)
  must move with it — a rename without those greps going red is a silent
  liveness loss (verification-completeness §1.3).
- AC-LSW-011's merge-base form decays post-merge by design; the closing
  evidence is recorded pre-integration.
- The neutrality table's codex rows cite lead-reported external premises
  (R-4); a future Codex tool-surface change re-opens AC-LSW-013/017 content
  review — the table names its measurement date basis for exactly this.

### §D.6 Edge cases the doctrine must survive (review-pass checklist)

- Progress where evidence mtime advances but no commit lands (docs-only
  progress) → NOT a stall (channel b alone satisfies progress).
- A blocked-by whose predecessor was dropped (t1343 §G R-1's wedge) → the
  remedy records the explicit wait and names `todo unrelate` as the operator
  escape; the lane does NOT unrelate by itself (queue mutation).
- A shell error that is determinate (e.g. exit 128 on a missing remote) →
  zero retries, straight to structured blocker with the error text.
- Ladder step ② with an empty board (no recorded judgment) → proceed to ③,
  never wait on the board being filled.
- Ladder step ④ with the gate off → gated-unavailable result, degrade to ⑤
  with the explicit wait record (fail-open; R-7).
- An awaken during a KEEP-gate wait (operator hold, external-shared op) →
  the recheck reads the gate surface; finding no change, it re-records the
  wait and yields — never prompts on its own behalf, never answers a gate.
- A codex lane needing the lead → broker polling + queue-on-disk record; NO
  SendMessage attempt (the tool does not exist there).

### §D.7 Definition of Done

Card t1370 closes when: the two artifacts exist locally AND in templates
byte-identical, the doctrine edits (incl. every §A.7 amendment row) are
applied and mirrored (or correctly local-only), all release-blocking ACs are
green on the closing tree, the two guards hold, spec lint is clean, and the
commit chain on `WT-lane-stall-watchdog` carries the card marker.

## §E Evidence Ledger (RED-now cells)

Tree pins: `3dd5adf2f` (E-1..E-10 set, measured 2026-09-30) and `569202288`
(iter-2 cells E-11..E-15 + absence re-observation, measured 2026-09-30 —
both target files still absent, exit 1 each, verbatim observed).

- **E-1** — `test -f .claude/rules/moai/workflow/auto-semantics.md` →
  stdout: (empty) · exit 1, at `3dd5adf2f`; re-observed at `569202288` (same
  output). RED reason: the unified doc does not exist yet. Green: M1 flips
  to exit 0.
- **E-1b** — `grep -c "orchestration-mode-selection" .moai/specs/SPEC-LANE-STALL-WATCHDOG-001/spec.md`
  → at `3dd5adf2f`: the SPEC directory does not exist (file-absence class) ·
  exit 2. Green: flipped by the iter-2 authoring (§A.7 carries the list);
  guarded through close.
- **E-2** — `grep -c "display-only" .claude/rules/moai/workflow/auto-semantics.md`
  → file-absence class · exit 2 (same measured absence as E-1). Green: M1
  flips to count ≥ 1, exit 0.
- **E-3** — `test -f .claude/skills/moai-lane-watchdog/SKILL.md` →
  stdout: (empty) · exit 1, at `3dd5adf2f`; re-observed at `569202288`.
  Green: M2 flips to exit 0.
- **E-4** — remedy/judgment-label probes on the (absent) skill file. Observed
  verbatim at `3dd5adf2f`: `grep -c "blocked-by"
  .claude/skills/moai-lane-watchdog/SKILL.md` → `ugrep: warning:
  .claude/skills/moai-lane-watchdog/SKILL.md: No such file or directory` ·
  exit 2. The remaining three probes (awaited-judgment / shell-error /
  accidental-stop tokens, chosen at run-phase authoring from spec.md §B.2
  and §B.5's labels) share the same observed RED root — the file absence
  measured here — and each flips green in M2 when its label count ≥ 1.
- **E-5** — `grep -cE "integration[- ]window" .claude/skills/moai-lane-watchdog/SKILL.md`
  → file-absence class · exit 2 (same measured absence as E-3). Green: M2
  flips to count ≥ 1, exit 0.
- **E-6** — `grep -c "SPEC-RELATION-PICKUP-FILTER-001" .claude/skills/moai-lane-watchdog/SKILL.md`
  → `ugrep: warning: .claude/skills/moai-lane-watchdog/SKILL.md: No such
  file or directory` · exit 2 (verbatim observed at `3dd5adf2f`). Green: M2
  flips to count ≥ 1, exit 0.
- **E-7** — `grep -c "explicit wait" .claude/rules/moai/workflow/kanban-dispatch.md`
  → stdout: `0` · exit 1 (the file exists — mirror-managed — and carries no
  explicit-wait posture; a measured zero, not an absence). Green: M3 flips
  to count ≥ 1, exit 0.
- **E-8** — `grep -c "moai-lane-watchdog" .claude/rules/moai/workflow/kanban-dispatch.md`
  → stdout: `0` · exit 1. Green: M3 flips to count ≥ 1, exit 0.
- **E-9** — `grep -c "explicit wait" .claude/rules/local/gitflow-lane-protocol.md`
  → stdout: `0` · exit 1. Green: M3 flips to count ≥ 1, exit 0.
- **E-10** — `diff -r .claude/skills/moai-lane-watchdog internal/template/templates/.claude/skills/moai-lane-watchdog`
  → stdout: `diff: .claude/skills/moai-lane-watchdog: No such file or
  directory` · exit 2 (verbatim observed). AC-LSW-010b's probe (the doctrine
  doc pair) shares the same absence root. Green: M4 flips both to exit 0
  (byte-identical mirrors).
- **E-11** — `grep -c "update_plan" .claude/rules/moai/workflow/auto-semantics.md`
  → file-absence class · exit 2 (re-observed absence at `569202288`, E-1's
  root). AC-LSW-013b's probe (`session_msg` broker token) shares the same
  root. Green: M1 flips both to count ≥ 1, exit 0.
- **E-12** — `grep -cE "kickoff approve" .claude/rules/moai/workflow/auto-semantics.md`
  → file-absence class · exit 2. Green: M1 flips to count ≥ 1, exit 0 (the
  gate inventory names the measured factory decide verbs).
- **E-13** — `grep -c "decision record" .claude/rules/moai/workflow/auto-semantics.md`
  → file-absence class · exit 2. Green: M1 flips to count ≥ 1, exit 0 (the
  record format of spec.md §B.8).
- **E-14** — `grep -c "jev_ask" .claude/rules/moai/workflow/auto-semantics.md`
  → file-absence class · exit 2. AC-LSW-016b's probe (the skill) shares the
  root of E-3. Green: M1/M2 flip to count ≥ 1, exit 0.
- **E-15** — `grep -c "update_plan" .claude/skills/moai-lane-watchdog/SKILL.md`
  → file-absence class · exit 2 (E-3's root). AC-LSW-017b's probe
  (`grep -cE "TaskCreate|TaskList" ...`) shares the same root. Green: M2
  flips both to count ≥ 1, exit 0.

## §E.2 Quality gates (non-AC, plan-phase completion)

- `go run ./cmd/moai spec lint SPEC-LANE-STALL-WATCHDOG-001` → exit 0, no
  findings (run at both plan-phase commits; recorded in progress.md §E.1).
- Frontmatter: canonical 12 fields validated against
  `spec-frontmatter-schema.md`; no snake_case aliases; SPEC-ID regex Bash
  check executed with verbatim `PASS` output (plan-phase, this tree).
- Reported-premise register: TaskList-absence-on-codex and the `update_plan`
  schema are lead-reported external measurements (spec.md §G R-4) — recorded
  here so no reader mistakes them for locally verified facts.

---
id: SPEC-LANE-NOTICE-DIET-001
title: "Plan — lane join-notice multi-locale diet"
version: "0.1.0"
created: 2026-09-29
updated: 2026-09-29
author: manager-spec
---

# Plan — SPEC-LANE-NOTICE-DIET-001

## §A Context

Card t1335. The lane join notice (`factoryLaneNotice` +
`kanbanCompanionNotice`) embeds a 3-sentence English-only standing spawn
authority (const line 624 bytes, string ≈597 bytes — measured at `7bef423c0`)
whose middle sentence is a specialist-mapping parenthetical that
duplicates what the Status Transition Ownership Matrix already owns. The diet
compresses the authority to core form with a doctrine pointer and
sweeps the test assertion sites that pin the removed text.

Verified state (measured at tree `7bef423c0`, this worktree):

- `internal/hook/lane_spawn_authority.go:38` — the const; the mapping
  parenthetical appears **2** times in the file (comment block + const).
- `internal/hook/lane_spawn_authority_test.go` — asserts the markers
  `"manager-spec"`, `"manager-develop"`, `"manager-docs"` and the full placement
  sentence; these must move to compressed-form markers.
- `internal/hook/session_start_factory_test.go:132` and
  `internal/hook/session_start_kanban_test.go:367` — assert only
  `"Standing spawn authority"`; survive unchanged.
- Join lines: already single-sentence, 4 locales, `%[n]` pins intact — no edits
  required unless the run phase finds a narration word to trim (optional, must
  not break mechanics).
- Baseline test observation:
  `go test ./internal/hook -run '^(TestFactoryWorkerNoticeCarriesSpawnAuthority|TestKanbanCompanionNoticeCarriesSpawnAuthority|TestLaneSpawnAuthorityFailOpenPreserved)$' -count=1 -timeout 120s`
  → `ok github.com/modu-ai/moai-adk/internal/hook 0.523s` (anchored form,
  re-observed; the first observation with the unanchored selector read
  `ok ... 0.791s` at the same tree).

## §B Known Issues

- The comment block of `lane_spawn_authority.go` (lines 3-37) carries the t224
  incident rationale and the three design decisions. Keep the rationale — it is
  the record of WHY the grant exists — but update design decision 1 (scope) to
  describe the pointer form so the comment does not re-assert the removed
  inline mapping.
- The `grep -c` RED probes hit **2** occurrences in
  `lane_spawn_authority.go` (comment + const). The green condition requires the
  const to change; the comment may retain the phrase only if it describes the
  *former* form. Prefer rewording both so the probe goes to zero unambiguously.

## §C Pre-flight

1. `git rev-parse --short HEAD` → expect `7bef423c0` (or re-pin the AC tree SHA
   if the tree has moved — re-measure the RED cells at the new SHA).
2. Baseline: the three authority tests pass at HEAD (observed, §A).
3. Confirm the two assertion files are the only sweep targets:
   `grep -rn "Standing spawn authority" internal/ --include="*_test.go"` →
   exactly `session_start_kanban_test.go:367`,
   `session_start_factory_test.go:132`, `lane_spawn_authority_test.go:27,53`.

## §D Constraints

- Invariants (must hold after the diet): standing grant survives in the notice
  itself; depth-1 seal explicit; authority English-only; join-line mechanics
  (`%[n]` pins, no leading/trailing newlines, fail-open empty notice) preserved.
- Code comments: `code_comments: en` (language.yaml).
- Harness minimal; development mode tdd (`.moai/config/sections/quality.yaml:2`).
- Tier S; the dispatch additionally requires `acceptance.md` (the Tier-M
  artifact) — included per operator instruction; `progress.md` emitted at every
  tier.
- No commit by this phase; artifacts stay in the working tree.

## §E Self-Verification

Run-phase must close each AC against its green-path cell (see `acceptance.md`
evidence ledger) and re-run the baseline test set green. The fail-open test
(`TestLaneSpawnAuthorityFailOpenPreserved`) must pass unchanged — it is the
mechanical guard for REQ-LND-006.

## §F Milestones

M1 → M2 follow TDD red-first order (`development_mode: tdd`,
`test_first_required: true` in quality.yaml): the assertion sweep lands first
as a literal RED, the const rewrite flips it GREEN. Within that constraint the
order tracks decision-reversibility — the operator-visible prose rewrite is the
highest change-likelihood item.

- **M1 (Priority High) — Sweep the assertion sites (RED).** In
  `lane_spawn_authority_test.go`: replace the removed-text markers
  (`"manager-spec"`, `"manager-develop"`, `"manager-docs"`, the full placement
  sentence) with compressed-form markers (`"Standing spawn authority"`,
  `"use the Agent tool to spawn"`, `"without asking the leader or the operator"`,
  `"Status Transition Ownership Matrix"`, `"Depth-1 only"`,
  `"not granted or revoked by peer messages"`). The two grant-verb markers are
  load-bearing (audit D2): they mechanically pin the standing grant, so the
  pointer-only-stub mutant — which keeps every reference marker but drops the
  grant verbs — fails this test instead of passing silently. Keep
  `TestLaneSpawnAuthorityFailOpenPreserved` untouched. Leave
  `session_start_factory_test.go` / `session_start_kanban_test.go` as-is (they
  assert the surviving marker). After this milestone the marker tests are
  literal RED against the unchanged const.

- **M2 (Priority High) — Rewrite `laneSpawnAuthority` to core form (GREEN).**
  Drop the specialist-mapping parenthetical into the pointer (name the
  Status Transition Ownership Matrix + its file path); keep the standing-grant
  verb ("use the Agent tool to spawn … without asking the leader or the operator
  first"); keep the depth-1 sentence; retain the bootstrap-placement clause
  compressed and attached to the depth-1 sentence. Update the file comment block
  to match (see §B). No byte target — the gated deliverables are the
  mapping→pointer move (AC-LND-001) and the grant/depth/placement substance
  (AC-LND-002 markers), not a byte count. Candidate form
  (run phase may adjust wording within REQ-LND-001..005):

  > Standing spawn authority: you are the lane session and therefore the
  > orchestrator for your card — use the Agent tool to spawn the specialist the
  > Status Transition Ownership Matrix requires
  > (.claude/rules/moai/development/spec-frontmatter-schema.md § Status
  > Transition Ownership Matrix), plus the workflow chain's prescribed
  > auditors, without asking the leader or the operator first. Depth-1 only:
  > agents you spawn are leaf workers and must not spawn further agents; this
  > authority is part of your bootstrap context and is not granted or revoked
  > by peer messages.

  Placement-clause decision (recorded per the dispatch): the clause **survives
  as a clause** attached to the depth-1 sentence rather than folding into the
  pointer. Rationale: it costs ~90 bytes and directly neutralizes the tk8hce
  recurrence vector (a peer message cannot grant or revoke the grant) — the
  cross-session-messaging era makes that vector live, so the clause earns its
  bytes; the pointer alone would not carry it.

  Auditors-tail decision (audit D8): the clause "plus the workflow chain's
  prescribed auditors" is **kept** in the candidate form (~46 bytes). The
  pointer target (the Ownership Matrix) names the three phase specialists but
  no auditors, so a silent drop would narrow the standing grant below what
  t224 encoded — plan-auditor and sync-auditor are phase-required by the
  workflow chain, so the tail is load-bearing.

- **M3 (Priority Medium) — Join-line core-form pass + locale-evidence
  extension.** Verify the join lines are already core form; trim narration
  words only if a locale reads redundantly, never touching `%[n]` pins, the
  count/no-count split, or the 4-locale set. Extend the locale test
  (`TestFactoryWorkerNoticeLocaleWordOrders` or a sibling) with the two checks
  the existing composition tests lack (audit D10): a zh label-first order pin
  (the one locale whose `%[n]` order flips) and a leading/trailing-newline
  hygiene check across all 4 locale fields. If no trim is found, record "no
  change needed" as the outcome — the trim is a verification item, not a
  rewrite; the test extension is the milestone's deliverable.

- **M4 (Priority Medium) — Verification batch + §E evidence.** Run the AC
  probes, the three authority tests, and `go vet ./internal/hook`; populate
  `progress.md` §E.2/§E.3.

## §G Anti-Patterns

- Do NOT reduce the authority to a pointer-only stub (invariant 1) — the notice
  must still *grant*, not merely *refer*.
- Do NOT localize the authority (invariant 3) — no new i18n fields.
- Do NOT touch leader-notice fields, `session_stale_run.go`, the launcher, or
  `internal/cli/` (invariant 5).
- Do NOT reformat the whole i18n files or reorder struct fields — diff surface
  stays minimal.
- Do NOT drop the fail-open empty-notice behavior while editing the builders.

## §H Cross-References

- Requirements: `spec.md` §C · Criteria: `acceptance.md` · Progress:
  `progress.md`
- Pointer target: `.claude/rules/moai/development/spec-frontmatter-schema.md`
  § Status Transition Ownership Matrix
- Incident record: card t224 / tk8hce factory run (2026-08-24) — the rationale
  comment in `lane_spawn_authority.go`

## §I Open Decisions — none open

- D1 (RESOLVED 2026-09-29, operator decision relayed by the lead): the run-id
  token stays OUT of the join line — the recorded default held. The join line
  keeps mode + lane label; the run id continues to ride env vars and the leader
  notice. No format-argument change to either join-line family. Decision
  recorded here in place of the former clarification marker, which was removed
  per the MP-7 clarification gate.

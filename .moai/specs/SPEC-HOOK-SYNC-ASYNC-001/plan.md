---
id: SPEC-HOOK-SYNC-ASYNC-001
title: "Implementation plan — hook sync/async cleanup"
version: "1.0.5"
created: 2026-10-15
updated: 2026-10-22
author: manager-spec
tier: M
---

# plan.md — SPEC-HOOK-SYNC-ASYNC-001

## §A Context

Card t1780 (Class C, Template-First, P1, operator directive 2026-10-15). The hook layer carries
~30 real registrations; the original survey decided three async conversions. The v1.0.5
follow-up retains ConfigChange and PostToolUseFailure async and keeps chain-event synchronous
for foreground completion-edge durability. The decided synchronous entry fast path, timeout
doctrine correction and measurement protocol remain unchanged.

Code facts verified at plan phase (all on tree `e04ebd5ca6bc441b380c1072d5a1dff8431459ba`):

| Fact | Source (verified) |
|---|---|
| Receipt store path | `.moai/state/verify/snapshots/<head12>-<digest>.json` — `internal/verify/store.go` `SnapshotPath` (head truncated to 12, `:` → `-`) |
| Receipt key algorithm | `sha256(git status --porcelain=v2 ‖ 0x00 ‖ git diff HEAD ‖ per-untracked-file digests)` — `internal/verify/key.go:40-65` |
| Receipt entry fields | `CheckID "codex-review"`, `Verdict`, `RecordedAt`, `ExitCode`, `ConfigDigest`, `ToolVersion` — `verify.Receipt` / `CheckEntry` |
| Freshness | `CheckReceipt`: 5 bound fields equal + `codexReviewCacheTTL = 24h` (`internal/cli/codex_review_cache.go:29`) |
| Kick command | `<self> verify codex-review --project-root <dir>` → `.moai/logs/codex-review-bg.log` — `codex_review_delay.go` `kickCodexBackgroundReview` |
| Kick marker | `.moai/state/verify/codex-review/<head>-<digest>.kick-<epoch/900>`; exclusive create = dedup; `DefaultCodexReviewGateTimeout = 900s` (`internal/config/defaults.go:819`) |
| Entry semantics | `HandleCodexReviewEntry` NEVER blocks (REQ-RG-001); fresh FAIL → `additionalContext` once per key (delivery record `delivered-<sha>.json`) |
| Scope discriminator | `WT-` branch prefix → card scope; everything else → tree scope, fail-open (`codex_review_scope.go`) |
| Entry-side skip policies | `workflow.codex.review_gate.tree_scope` (unreadable → `review`) and `primary_scope` (unreadable → `skip` — primary-checkout default); both applied at one shared spot `codex_review_delay.go:217` (`treeScopeSkipApplies`) so "the two halves skip together"; a `WT-` card session is never skipped (`codex_review_tree_scope.go:55-135`, REQ-CRO-004) |
| Binary-age policy | `staleBinarySkipApplies` (`codex_review_binary_age.go`) — `binlag.Evaluate` is a pure git-ancestry judge (binary build commit vs tree HEAD, `binlag.go:102-132`), no state cache; leniency contract: only a DECIDABLE behind verdict skips — fresh / divergent / ahead / not-applicable (incl. no commit metadata) review as before; the explicit producer `moai verify codex-review` never calls it; the entry path (:214-216) returns ALLOW on behind |
| Nothing-reviewable detector | `reviewGateScopedChangeDetector` (`codex_review_scope.go:269-277`, tree-scope form): change presence AND not config-only; the docstring's own rule (:268): "The probe can only narrow the detector's answer, never widen it"; the config-only narrowing reads a Go-owned runtime-config prefix list |
| Receipt-absence semantics | `verify.LoadReceipt` (`receipt.go:113-117`): every read failure — no snapshot, unreadable, truncated, unparseable, no entry — is ABSENCE, never an error and never a valid/failing verdict |
| Go SystemMessage emission | `internal/hook/post_tool_failure.go:109` `SystemMessage: message` |
| Template UPS timeouts | `settings.json.tmpl` lines 266/272: `timeout: 5` (both UPS entries) |
| This worktree live UPS | `settings.json` line 266: `timeout: 15` (primary checkout live already raised to 30 by the leader) |
| Doctrine wording to correct | `hooks-system.md:368` (table row "15s \| 30s (reduced max)") and `:374` (ceiling bullet) — live and template byte-identical |
| Async baseline | 11 `"async": true` entries in live settings (grep count, this tree) |

## §B Known Issues

- K1: The three-way timeout drift (primary 30 / this worktree 15·5 / template 5·5) means a naive
  template alignment commit would silently regress the primary's leader-raised values if this
  branch absorbs a stale base. M2 re-observes all three surfaces after the absorb.
- K2: The card's "block on fresh FAIL" phrasing predates the REQ-RG-001 no-block amendment; the
  plan designs the fall-through toward the handler that exists (spec.md §1 note). No scope change.
- K3: The entry wrapper today never kicks ("enforcement and kick-start stay on separate hooks" —
  `codex_review_delay.go` header). The card directs an entry-side shell-detected kick; the
  no-wait property is preserved because the kick is detached, and double-kick safety rides the
  existing shared marker, not a new protocol.

- K4 (v1.0.5): the official async lifecycle cancels running hooks at `claude -p` teardown.
  `chain-event.sh` invokes its append in the foreground without detaching, so async registration
  can lose a valid final-child completion edge. A missing in-turn decision output is not enough
  to make this persistence safe to defer. Both vendor async message fields are delivered on the
  next turn; the failure handler's additionalContext-only choice remains application policy.

## §C Pre-flight

- P1: `go test ./internal/hook ./internal/cli ./internal/template -count=1` green on the absorbed
  base before M3 begins.
- P2: Re-run the live/template inventory and timeout probes on the actual amendment tree.
  The original plan baseline was 11; source inspection at `08b87f504ff8688affe82d468e14b93cccac51ba`
  now finds 14 async command registrations on each surface, including exactly one chain-event.
  Removing only that async property must yield exactly 13 (baseline 11 + the two retained
  conversions). Preserve every other registration and timeout; a moved baseline re-pins the
  ledger per `verification-completeness.md` §4.
- P3: Confirm `.moai/reports/t1780/hook-sync-timings.md` from M1 exists and carries 10 runs per
  synchronous hook before M3 prescribes any further change.

## §D Constraints

- D1 (HARD, from spec.md C1): M3-M7 execute only after M2 passes the t1775 gate. M1 is carved out
  explicitly (read-only on code; writes only `.moai/reports/t1780/`).
- D2: Live + template edits land as commit pairs (Template-First). The run phase MUST NOT commit
  a live edit whose template mirror is absent from the same push.
- D3: The shell fast path's dependency set is `git` + one sha256 tool (`shasum -a 256` on macOS,
  `openssl dgst -sha256` fallback) + `awk` (already used by the gate-OFF path) + the system
  `date +%s` clock. The clock is the only addition in the 2026-10-22 portability amendment:
  GNU and BSD `date` provide Unix seconds, while an awk random seed has no portable time unit.
  Validate the clock as decimal Unix seconds before either receipt-age arithmetic or
  epoch/900 bucket selection. A missing, failing, or invalid clock falls through to Go under
  REQ-HSA-007; it never turns uncertainty into ALLOW or a synthetic receipt. Nothing else is added.
- D4: The shell key computation must reproduce `verify.Key` byte-for-byte: hash
  porcelain-v2 output, one 0x00 byte, `git diff HEAD` output, then each untracked file's content
  digest exactly as `key.go` chains them (read `key.go:40-90` in full before implementing; the
  untracked loop reads file contents into the SAME hasher). A key-parity test (shell script vs
  `go run` harness on a fixture tree) is the gate for M3.
- D5: The fast path handles TREE scope in shell only. Card scope (`WT-` branch), unkeyed state,
  or any probe failure falls through to the Go handler (REQ-HSA-007). This is deliberate: the
  scope resolver's card-class key computation (`cardScopeKeyParts`) is not reproduced in shell.
- D6: No artificial load for measurements (spec.md C4); measurement runs use the tree's own
  sessions/work, and the UserPromptSubmit hard-timeout observation comes from real multi-lane
  operation.

## §E Self-Verification

The run phase (manager-develop) owns §E evidence; this plan pre-declares the checks:

- E1: AC matrix in acceptance.md all GREEN with ledger citations re-measured on the landing tree.
- E2: `go test ./internal/hook ./internal/cli ./internal/template -count=1` output carried in the
  completion report.
- E3: Live/template registration tests assert the two retained async conversions, synchronous
  chain-event, strict async count13, both UPS entries at30 and all other keep-sync invariants.
  A final-child lifecycle regression also proves the foreground completion edge exists before
  hook return/teardown with a writable store; no detached writer or synthetic edge substitutes.
- E4: Subagent-boundary: no cross-tree writes; all edits inside this worktree.
- E5: `gofmt`/`go vet` on touched Go packages.
- E6: The key-parity test green on both macOS (`shasum`) and the `openssl` fallback path.
- E7: The measurement table committed; the RED→GREEN ledger sweep recorded in acceptance.md.

## §F Milestones

Ordering note: milestones are ordered by decision-reversibility inside the post-gate block —
the fast-path design (most likely to change as the key parity and probe behavior meet reality)
leads; mechanical flag flips and doc edits trail. M1 is priority-ordered first because it is
unblocked and its evidence feeds everything downstream.

- **M1 — Measurement protocol (P1, unblocked now; no t1775 dependency).**
  Inventory every synchronous registration from `.claude/settings.json`; measure each 10 times
  under current machine load; record per-run timings + summary (min/median/max) into
  `.moai/reports/t1780/hook-sync-timings.md`. Method: pipe a representative stdin JSON per event
  to the wrapper script with `bash` + `time` (or `perl -MTime::HiRes`), 10 runs each, no
  artificial load; record gate ON/OFF state per measurement.
  **Environment (HARD — audit r1 D3):** measurements run in an ISOLATED environment — a temp
  project dir (fixture `.moai/` tree) plus a temp `HOME` — OR a dry-run allowlist restricted to
  side-effect-free wrappers; the table records, per wrapper, which environment it ran under.
  Rationale: the live-tree SessionEnd handler is stateful and destructive
  (`internal/hook/session_end.go` releases profile leases :72-73, ClearAgentStops :111,
  `os.RemoveAll` on team/task dirs :406-633) — it, and any wrapper reaching a stateful handler,
  MUST be exercised under isolation only, never 10x against the live tree. The zero-side-effect
  wrappers (the pure-shell gate-OFF fast paths) may run live and are marked as such in the table.
  Prescriptions from the table follow
  the fixed order (fast-path → timeout raise → async); only the two retained conversions and
  the codex-review-entry fast path are IN scope — other findings become follow-up cards.
  Acceptance: AC-HSA-013.
- **M2 — t1775 landing gate + absorb (P0 gate).**
  Verify card t1775 landed on local main (its merge commit visible in `git log main`), absorb
  main into this card branch (`git merge main` inside this worktree), resolve any conflict in
  `settings.json.tmpl` / `hooks-system.md` preserving BOTH cards' content, then re-run the §C
  P2 probes to re-pin the baseline. Blocks M3-M7. Acceptance: gate check recorded in the card's
  progress record; probes re-pinned.
- **M3 — codex-review-entry gate-ON shell fast path (P1, highest design reversibility).**
  Extend `.claude/hooks/moai/handle-codex-review-entry.sh` after the existing gate-OFF block:
  (1) eligibility pre-check — a `WT-` branch (card scope) falls through to the binary, and so
  does a session under either entry-side skip policy: the `workflow.codex.review_gate.tree_scope:
  skip` config (read from the same workflow.yaml the gate-OFF awk parse already reads; the
  SHELL fail direction is toward falling through — misreading `skip` as `review` would kick
  where the Go design suppresses, while misreading `review` as `skip` only costs the cold start
  the Go handler then applies correctly) and the primary-checkout `primary_scope: skip` default,
  detected in shell via `git rev-parse --git-common-dir` resolving inside `$CLAUDE_PROJECT_DIR`
  (the session tree IS the primary working tree); any detection ambiguity falls through (D2;
  REQ-HSA-006 Where-clause; mirrors "the two halves skip together", `codex_review_delay.go:
  188/:217`); (2) compute head + tree key in shell (D4); (3) read the snapshot JSON with the
  existing awk tooling (no jq); locate the `codex-review` entry, compare verdict + RecordedAt vs
  the 24h TTL — an unreadable or unparseable snapshot is receipt-ABSENT (the shared
  `verify.LoadReceipt` reading, `receipt.go:113-117`) and takes step (4)'s miss arm;
  (4) cache-MISS states (no receipt / stale receipt) → kick, gated on the two kick-region
  preconditions (`codex_review_delay.go:214-225`), each shell-carried where judgeable:
  (4a) nothing-reviewable — change presence (porcelain/diff non-empty) AND the config-only
  narrowing (all changed paths within the runtime-config prefix list, mirrored from the Go
  constants and pinned by a golden test so the probe only narrows, never widens —
  `codex_review_scope.go:268`); clean tree or config-only-only tree → exit 0 WITHOUT the early
  kick (the Stop gate's Go path owns the decision, REQ-GBN-002 preserved);
  (4b) binary-age — resolve MOAI_BIN, read its version provenance (build commit), then the
  HEAD-match pre-check FIRST (HEAD starting with the bin commit → NOT behind — StatusFresh,
  `binlag.go:139-141` — and the early kick proceeds; the ancestry probe alone would mis-class
  an exactly-current binary, since `merge-base --is-ancestor` succeeds on equality); else
  `git merge-base --is-ancestor <bin-commit> HEAD`: a DECIDABLE ancestor-hit (behind) → exit 0
  WITHOUT the early kick (mirrors the entry path's :214-216 allow); undecidable (no commit
  metadata, non-ancestor) → proceed per binlag's leniency (`codex_review_binary_age.go:33-37`);
  (4c) both preconditions satisfied → kick (marker-guarded,
  `set -o noclobber` exclusive create of the `.kick-<bucket>` marker, bucket = epoch/900) then
  exit 0. Cost note: the (4b) age judgment pays one `version` invocation on miss states before
  the kick (+1 process start vs today on miss states only — the price of the accelerated kick);
  fresh pass/inconclusive remain 0-start and fresh-FAIL remains 1-start (unchanged);
  (5) fresh pass/inconclusive → exit 0 WITHOUT kicking (mirror
  `codex_review_gate.go:234-244` — a fresh passing review is never re-kicked); (6) fresh-FAIL →
  fall through to the binary (the existing line 84 forwarding); (7) every failure → fall through
  (REQ-HSA-007). Mirror the
  script to `internal/template/templates/.claude/hooks/moai/handle-codex-review-entry.sh`.
  Tests: key-parity test (D4) plus a golden test pinning the shell's runtime-config prefix list
  to the Go constants it mirrors (4a); behavioral probe matrix — gate off / no receipt /
  stale receipt / unparseable snapshot (miss arm) / fresh pass (no kick) / fresh inconclusive
  (no kick) / fresh FAIL (fall-through) / clean tree (no kick) / config-only-only tree (no kick) /
  behind binary (no early kick) / HEAD-match binary (StatusFresh — kick proceeds) / binary with
  no commit metadata (kick proceeds per leniency) /
  missing git / missing shasum / `WT-` branch / `tree_scope: skip` / primary-checkout session —
  each asserting
  exit 0 and the expected kick-or-fall-through observation (kicks observed via the shared marker
  path; binary invocations observed via an argv-recording stub distinguishing
  `hook codex-review-entry` from `verify codex-review`); registration tests untouched (timeout
  change is M6). Acceptance: AC-HSA-006, AC-HSA-007,
  AC-HSA-008, AC-HSA-015, AC-HSA-016, AC-HSA-017.
- **M4 — PostToolUseFailure async conversion (P1).** One atomic change set: Go handler emits
  `hookSpecificOutput.additionalContext` (event name `PostToolUseFailure`) instead of
  `SystemMessage` at `post_tool_failure.go:108-110`; `async: true` on the entry in live
  `.claude/settings.json` AND `settings.json.tmpl`; update the handler's Go tests to the new
  output shape; add a template-registration test asserting the flag. The edit set also carries
  the `hooks-system.md` PostToolUseFailure stdout row (systemMessage →
  hookSpecificOutput.additionalContext), live + template mirrors (audit r1 D7), so the doctrine
  does not go stale against the converted handler. The flag and the Go change
  land in the same commit pair to preserve the chosen output contract. The vendor supports
  both message fields on async hooks; claiming that `systemMessage` is unsupported is forbidden.
  General async doctrine must state next-turn delivery of both fields and no user-UI display,
  while this handler continues to emit only additionalContext. Acceptance: AC-HSA-004, AC-HSA-005.
- **M5 — ConfigChange async and synchronous chain-event durability (P2).** Keep ConfigChange
  async on both surfaces and remove only chain-event's async property in both live/template
  settings. Keep its existing timeout, fail-open behavior and foreground append path; no
  detachment is introduced. Registration tests must reject an async chain-event, and a lifecycle
  test with a valid node/writable ledger must schedule final-child shutdown while the append is
  held, then prove the hook does not return until the persisted completion edge is observable.
  An unavailable context/store or append error must remain fail-open without forging an edge.
  Acceptance: AC-HSA-001, AC-HSA-002, AC-HSA-003.
- **M6 — Timeout doctrine correction + alignment (P2).**
  `hooks-system.md` (live + template): rewrite the table row
  (`UserPromptSubmit | 30s | default 30s, no documented per-event maximum (general hook ceiling
  applies) | ...`) and the line-374 bullet (30s is the official default, not a runtime cap;
  headroom above 30s exists where a hook needs it). `settings.json.tmpl`: both UPS entries
  `timeout: 5 → 30`. This tree's live `.claude/settings.json`: UPS entries to 30 (15→30, 5→30);
  primary checkout already at 30 (leader-raised; verify, do not touch). Acceptance: AC-HSA-010,
  AC-HSA-011, AC-HSA-012.
- **M7 — Verification close (P1).** Full AC sweep with ledger re-measure (green cells); keep-sync
  regression sweep — strict async count13 after removing only chain-event from the current14,
  with the two retained conversions and every other keep-sync registration unchanged; E1-E7
  evidence. Include the now-synchronous chain-event in the existing10-run measurement protocol.
  Historical GREEN cells/count14 do not close the changed durability or inventory criteria;
  current registration/lifecycle results, final-head CI, review and audit remain to be observed.
  Acceptance: AC-HSA-009, AC-HSA-014, AC-HSA-015.

## §G Anti-Patterns

- Do NOT convert any keep-sync table registration "while in the file" — scope discipline; the
  table is the boundary.
- Do NOT implement the shell key with a different but "equivalent" hash chaining — D4 requires
  byte parity with `key.go`; a near-miss key reads as receipt-absent and silently disables the
  fast path's kick arm (fail-open but functionally dead).
- Do NOT kick on a fresh pass/inconclusive receipt — the Stop gate never does
  (`codex_review_gate.go:234-244`); an idle passing tree would burn a ~900s review per bucket
  for nothing.
- Do NOT let any probe failure block the prompt — the only acceptable degradation is the current
  behavior (fall through to the binary; exit 0 when it cannot resolve).
- Do NOT raise timeouts without the M1 table row justifying them (spec.md Out of Scope).
- Do NOT edit `.moai/state/verify/` runtime state from the wrapper beyond the kick marker
  create/GC the Stop arm already defines.

## §H Cross-References

- spec.md (requirements REQ-HSA-001..012, constraints, out-of-scope)
- acceptance.md (AC matrix + evidence ledger)
- `.claude/rules/moai/development/verification-completeness.md` §2 (two-cell discipline)
- `.claude/rules/moai/development/hook-independence.md` §7 (authoring checklist)
- `.claude/rules/moai/core/hooks-system.md` (timeout doctrine being corrected)
- internal/verify/key.go, internal/verify/store.go, internal/cli/codex_review_delay.go,
  internal/cli/codex_review_scope.go, internal/hook/post_tool_failure.go

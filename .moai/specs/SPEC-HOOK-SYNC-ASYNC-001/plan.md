---
id: SPEC-HOOK-SYNC-ASYNC-001
title: "Implementation plan — hook sync/async cleanup"
version: "1.0.0"
created: 2026-10-15
author: manager-spec
tier: M
---

# plan.md — SPEC-HOOK-SYNC-ASYNC-001

## §A Context

Card t1780 (Class C, Template-First, P1, operator directive 2026-10-15). The hook layer carries
~30 real registrations; the leader's survey decided three async conversions, one decided-stays-sync
registration that gains a gate-ON shell fast path, a timeout doctrine correction, and a
measurement protocol over every synchronous hook.

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

## §C Pre-flight

- P1: `go test ./internal/hook ./internal/cli ./internal/template -count=1` green on the absorbed
  base before M3 begins.
- P2: Re-run the inventory probe (`grep -c "async.: true" .claude/settings.json` → expect 11) and
  the timeout probes on the absorbed tree; a moved baseline re-pins the ledger (per
  `verification-completeness.md` §4, re-measure on rebase).
- P3: Confirm `.moai/reports/t1780/hook-sync-timings.md` from M1 exists and carries 10 runs per
  synchronous hook before M3 prescribes any further change.

## §D Constraints

- D1 (HARD, from spec.md C1): M3-M7 execute only after M2 passes the t1775 gate. M1 is carved out
  explicitly (read-only on code; writes only `.moai/reports/t1780/`).
- D2: Live + template edits land as commit pairs (Template-First). The run phase MUST NOT commit
  a live edit whose template mirror is absent from the same push.
- D3: The shell fast path's dependency set is `git` + one sha256 tool (`shasum -a 256` on macOS,
  `openssl dgst -sha256` fallback) + `awk` (already used by the gate-OFF path). Nothing else.
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
- E3: Template-registration tests assert: three async flags present in `.tmpl`; both UPS entries
  at 30 in `.tmpl`; keep-sync registrations unchanged.
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
  artificial load; record gate ON/OFF state per measurement. Prescriptions from the table follow
  the fixed order (fast-path → timeout raise → async); only the three decided conversions and
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
  (1) scope pre-check — a `WT-` branch falls through to the binary (D5); (2) compute head +
  tree key in shell (D4); (3) read the snapshot JSON with the existing awk tooling (no jq);
  locate the `codex-review` entry, compare verdict + RecordedAt vs the 24h TTL; (4) ALLOW
  states → kick (marker-guarded, `set -o noclobber` exclusive create of the `.kick-<bucket>`
  marker, bucket = epoch/900) then exit 0; (5) fresh-FAIL → fall through to the binary (the
  existing line 84 forwarding); (6) every failure → fall through (REQ-HSA-007). Mirror the
  script to `internal/template/templates/.claude/hooks/moai/handle-codex-review-entry.sh`.
  Tests: key-parity test (D4); behavioral probe matrix — gate off / no receipt / stale receipt /
  fresh pass / fresh FAIL / missing git / missing shasum / WT- branch — each asserting exit 0
  and the expected fall-through-or-not observation; registration tests untouched (timeout change
  is M6). Acceptance: AC-HSA-006, AC-HSA-007, AC-HSA-008.
- **M4 — PostToolUseFailure async conversion (P1).** One atomic change set: Go handler emits
  `hookSpecificOutput.additionalContext` (event name `PostToolUseFailure`) instead of
  `SystemMessage` at `post_tool_failure.go:108-110`; `async: true` on the entry in live
  `.claude/settings.json` AND `settings.json.tmpl`; update the handler's Go tests to the new
  output shape; add a template-registration test asserting the flag. The flag and the Go change
  land in the same commit pair — flag without the Go change silently drops the message, and vice
  versa on the sync path. Acceptance: AC-HSA-004, AC-HSA-005.
- **M5 — ConfigChange + SubagentStop chain-event async flags (P2, mechanical).** `async: true` on
  the two entries in live settings + template; template-registration tests assert both flags.
  Acceptance: AC-HSA-001, AC-HSA-002, AC-HSA-003 (flag cells).
- **M6 — Timeout doctrine correction + alignment (P2).**
  `hooks-system.md` (live + template): rewrite the table row
  (`UserPromptSubmit | 30s | default 30s, no documented per-event maximum (general hook ceiling
  applies) | ...`) and the line-374 bullet (30s is the official default, not a runtime cap;
  headroom above 30s exists where a hook needs it). `settings.json.tmpl`: both UPS entries
  `timeout: 5 → 30`. This tree's live `.claude/settings.json`: UPS entries to 30 (15→30, 5→30);
  primary checkout already at 30 (leader-raised; verify, do not touch). Acceptance: AC-HSA-010,
  AC-HSA-011, AC-HSA-012.
- **M7 — Verification close (P1).** Full AC sweep with ledger re-measure (green cells); keep-sync
  regression sweep — async count 11 → 14 exactly, no keep-sync registration altered; E1-E7
  evidence; measurement table finalized + committed; acceptance.md ledger closed with GREEN
  observations pinned to the landing tree SHA. Acceptance: AC-HSA-009, AC-HSA-014, AC-HSA-015.

## §G Anti-Patterns

- Do NOT convert any keep-sync table registration "while in the file" — scope discipline; the
  table is the boundary.
- Do NOT implement the shell key with a different but "equivalent" hash chaining — D4 requires
  byte parity with `key.go`; a near-miss key reads as receipt-absent and silently disables the
  fast path's kick arm (fail-open but functionally dead).
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

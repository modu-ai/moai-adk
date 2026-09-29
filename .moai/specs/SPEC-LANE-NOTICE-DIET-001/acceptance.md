---
id: SPEC-LANE-NOTICE-DIET-001
title: "Acceptance — lane join-notice multi-locale diet"
version: "0.1.0"
created: 2026-09-29
updated: 2026-09-29
author: manager-spec
---

# Acceptance — SPEC-LANE-NOTICE-DIET-001

Per `.claude/rules/moai/development/verification-completeness.md` §2, every
release-blocking criterion carries a RED-now cell and a green-path cell as a
pair. Commands are in single-invocation form (no pipes, no chaining); the exit
code is recorded as its own field. Evidence lives in the ledger below; table
cells cite ledger ids.

**Document-level tree pin: `7bef423c0`** (branch `WT-bootstrap-notice-diet`) —
binds every criterion carrying no pin of its own.

## §A Evidence Ledger

```text
EV-1  (RED-now, AC-LND-001)
  cmd:   grep -c "plan-phase artifacts to manager-spec" internal/hook/lane_spawn_authority.go
  stdout: 2
  exit:  0
  tree:  7bef423c0
  why-red: the inline specialist-mapping parenthetical is present in BOTH the
    comment block and the const — the diet's own target text, so red for the
    right reason; M2 (const rewrite) removes both occurrences and the probe
    flips to exit 1.

EV-2  (RED-now, AC-LND-002)
  cmd:   grep -c '"manager-spec"' internal/hook/lane_spawn_authority_test.go
  stdout: 1
  exit:  0
  tree:  7bef423c0
  why-red: the test's marker list asserts a specialist name the diet removes
    from the const — pre-existing assertion text this work itself must change,
    so red for the right reason; M1 (assertion sweep) replaces the marker list
    and the probe flips to exit 1.

EV-3  (baseline-green observation, AC-LND-003)
  cmd:   go test ./internal/hook -run '^(TestFactoryWorkerNoticeCarriesSpawnAuthority|TestKanbanCompanionNoticeCarriesSpawnAuthority|TestLaneSpawnAuthorityFailOpenPreserved)$' -count=1 -timeout 120s
  stdout: ok  	github.com/modu-ai/moai-adk/internal/hook	0.523s
  exit:  0
  tree:  7bef423c0
  note:  anchored selector — sweeps exactly the 3 named tests (no empty-sweep
    ambiguity; the alternation is anchored at every branch).

EV-4  (baseline-green observation, AC-LND-004)
  cmd:   grep -rc laneSpawnAuthority internal/hook/session_start_factory_i18n.go internal/hook/session_start_kanban_i18n.go
  stdout: internal/hook/session_start_kanban_i18n.go:0
          internal/hook/session_start_factory_i18n.go:0
  exit:  1
  tree:  7bef423c0
  note:  exit 1 IS the green signal here — it is grep's no-match exit for an
        absence probe (re-observed 2026-09-29 with the system grep; a 0 would
        mean an unexpected reference). Zero authority references in the i18n
        tables — the English-only invariant already holds.

EV-5  (baseline-green observation, AC-LND-007)
  cmd:   git status --porcelain -- internal cmd pkg
  stdout: (empty)
  exit:  0
  tree:  7bef423c0
  note:  untracked-aware scope probe (audit D3) — unlike `git diff --name-only`,
        `--porcelain` lists modified AND untracked paths, so a leaked new file
        inside internal/ cmd/ pkg/ is visible. At baseline the output is empty
        (the only untracked paths are under .moai/specs/, outside the probe
        scope). The whitelist check: every listed path must be one of
        internal/hook/lane_spawn_authority.go,
        internal/hook/lane_spawn_authority_test.go,
        internal/hook/session_start_factory_i18n.go,
        internal/hook/session_start_kanban_i18n.go.
```

## §D AC Matrix

| AC | Mapping | Statement | RED-now | Green path | Class |
|----|---------|-----------|---------|------------|-------|
| AC-LND-001 | maps REQ-LND-003, REQ-LND-005 | The authority const carries no inline specialist-mapping parenthetical; the mapping reaches the lane only via the Status Transition Ownership Matrix pointer, in a const that still grants the spawn (EV-1 red) | EV-1: exit 0 (2 hits) | M2 (const rewrite) removes both occurrences → EV-1 command exits 1 (no match) | release-blocking |
| AC-LND-002 | maps REQ-LND-001, REQ-LND-002, REQ-LND-005 | The assertion sites assert the compressed-form markers (`"Standing spawn authority"`, `"use the Agent tool to spawn"`, `"without asking the leader or the operator"`, `"Status Transition Ownership Matrix"`, `"Depth-1 only"`, `"not granted or revoked by peer messages"`) and no longer assert the removed specialist-name markers (EV-2 red). The two grant-verb markers mechanically pin the standing grant (REQ-LND-001, audit D2): the pointer-only-stub mutant that drops them now fails the M1-swept test | EV-2: exit 0 (1 hit) | M1 sweeps the marker list → EV-2 command exits 1; M2 rewrites the const → `go test` (EV-3 command) `ok` | release-blocking |
| AC-LND-003 | maps REQ-LND-001, REQ-LND-002, REQ-LND-006 | The notice test set — both carry-authority tests plus the fail-open test — passes on the compressed notice | n/a (baseline green, EV-3) | M2/M4 re-run the EV-3 command → `ok` at the post-diet tree | regression-guard |
| AC-LND-004 | maps REQ-LND-004 | The authority stays English-only: zero `laneSpawnAuthority` references in either i18n table; join lines remain localized in en/ko/ja/zh | n/a (baseline green, EV-4, exit 1 = green) | M4 re-runs EV-4 → counts stay 0:0 with exit 1 | regression-guard |
| AC-LND-005 | maps REQ-LND-007 | Join-line mechanics: the en exact sentence and `%!`-format-artifact absence are pinned TODAY by `TestFactoryWorkerNoticeNamesLabel` / `TestFactoryWorkerNoticeLocaleWordOrders`; the zh label-first `%[n]` order and ja/ko/zh leading/trailing-newline hygiene are NOT pinned by existing tests (audit D10) — M3 extends the locale test with both checks | n/a (baseline green for the pinned checks; the extended checks do not exist yet — they are M3's deliverable, not a pre-work RED) | M3 adds the checks → the extended locale test `ok`; any later zh order flip or newline introduction fails it | regression-guard |
| AC-LND-006 | maps REQ-LND-006 | An unparseable label emits the empty string from both builders | n/a (baseline green — `TestLaneSpawnAuthorityFailOpenPreserved` in EV-3) | M4: the fail-open test passes unchanged | regression-guard |
| AC-LND-007 | maps REQ-LND-008 | Out-of-scope surfaces untouched: the untracked-aware probe `git status --porcelain -- internal cmd pkg` (EV-5) lists only whitelisted paths — `lane_spawn_authority.go`, `lane_spawn_authority_test.go`, and (only if M3 trims) the two i18n files | n/a (regression-guard — no honest pre-work RED exists for a no-leak invariant; baseline observed: EV-5 empty, exit 0) | M4: the probe output stays within the whitelist; any other path (modified or NEW) is a scope violation | regression-guard |

## §D.1 Severity

- Release-blocking: AC-LND-001, AC-LND-002 (both carry a real 4-element RED-now
  cell).
- Regression-guard: AC-LND-003..007 (baseline-green by construction — they
  describe behavior the diet must preserve or invariants that cannot be red
  before the work; per verification-completeness §2.1 their RED cannot be
  reproduced on the pre-work tree, so they are recorded as guards with pinned
  baseline observations, never as passes of this work). AC-LND-007 is a guard,
  not blocking (audit D3): its no-leak claim has no honest pre-work RED, and
  its probe (`git status --porcelain`, EV-5) is untracked-aware, closing the
  `git diff --name-only` blindness to new files.

## §D.2 Traceability

REQ-LND-001 → AC-LND-002/003 · REQ-LND-002 → AC-LND-002/003 ·
REQ-LND-003 → AC-LND-001 · REQ-LND-004 → AC-LND-004 ·
REQ-LND-005 → AC-LND-001/002 · REQ-LND-006 → AC-LND-003/006 ·
REQ-LND-007 → AC-LND-005 · REQ-LND-008 → AC-LND-007.

## §D.3 Edge Cases

- Label parses but `lanes < 1` (incremental `-f lane-<n>` entry): the no-count
  join variant renders, authority still appended — covered by the with-count /
  no-count table in `TestFactoryWorkerNoticeCarriesSpawnAuthority`.
- Both occurrences of the mapping phrase (comment + const) must leave the file;
  a const-only removal leaves EV-1 at 1 hit and fails AC-LND-001.

## §D.4 Quality Gates

- `go vet ./internal/hook` clean.
- The EV-3 test set `ok`.
- Harness: minimal — no coverage threshold applies beyond the package's
  existing suite.

## §D.5 Definition of Done

- All release-blocking ACs (AC-LND-001, AC-LND-002) flipped to their green-path
  state with verbatim output recorded in `progress.md` §E.2.
- All regression-guard ACs re-observed green at the post-diet tree.
- No file outside the AC-LND-007 whitelist modified (verified by the EV-5
  untracked-aware probe).

# Acceptance Criteria — SPEC-TOKEN-EFFICIENCY-001

Canonical AC enumeration (SSOT). Each AC is binary-verifiable. REQ↔AC bijection:
7 REQs (REQ-TEF-001…007) ↔ 7 ACs (AC-TEF-001…007).

> **Renumber note (v0.2.0).** The "~7×" reword item (formerly AC-TEF-005/006/007)
> was carved out to SPEC-DIVECC-ATTRIBUTION-FIX-001. The statusline ACs formerly
> numbered AC-TEF-008/009/010 are renumbered contiguously to AC-TEF-005/006/007.

Legend: **G** = Given, **W** = When, **T** = Then.

---

## P0-1 — always-loaded token-budget guard

### AC-TEF-001 (↔ REQ-TEF-001) — guard measures aggregate and fails over budget

- **G**: the token-budget guard test exists in `internal/config` with a named
  budget constant and an `estimateTokens` function.
- **W**: `go test ./internal/config/ -run <GuardTestName>` runs against a
  synthetic surface whose estimated tokens exceed the budget constant.
- **T**: the test FAILS (non-zero) with a message naming the measured total, the
  budget, and the overflow amount.
- Verify: a fixture/sub-test injecting an over-budget input asserts `t.Errorf`
  fired (e.g. via a table row `{name:"over-budget", want:fail}`).

### AC-TEF-002 (↔ REQ-TEF-002) — always-loaded surface enumerated correctly, PASSES on current tree

- **G**: the current repo tree's always-loaded surface — the set of
  `.claude/rules/moai/**/*.md` files carrying NO `paths:` frontmatter, plus the
  3 fixed surfaces (`CLAUDE.md`, `.claude/output-styles/moai/moai.md`,
  `MEMORY.md` head). The count and byte size are whatever the tree holds at test
  time; the AC does NOT pin them to a snapshot literal.
- **W**: `go test ./internal/config/ -run <GuardTestName>` runs against the real
  repo surface.
- **T**: the guard PASSES; the enumerated surface = `CLAUDE.md` + every
  `.claude/rules/moai/**/*.md` with NO `paths:` frontmatter +
  `.claude/output-styles/moai/moai.md` + `MEMORY.md` head (first 200 lines OR
  25KB, whichever first). The assertion is relative: `(count of no-paths: rule
  files) + 3 fixed surfaces`, and `estimated-token-sum ≤ budget constant` — never
  a literal file count.
- Verify: the enumeration count printed by the test (or asserted in a sub-test)
  equals `(number of no-`paths:` rule files discovered at test time) + 3`; a
  `paths:`-scoped rule (e.g. `languages/go.md`) is NOT in the enumerated set. No
  hardcoded count (e.g. `13`) appears in the assertion.

### AC-TEF-003 (↔ REQ-TEF-003) — estimation method + budget are named/documented

- **G**: the guard source.
- **W**: `grep -n 'estimateTokens\|<BudgetConstName>' internal/config/*.go`.
- **T**: both the estimation function and the budget constant are named symbols
  with a doc comment explaining the method (char/4 approximation) and the budget
  derivation (baseline + headroom) — no bare magic number.

### AC-TEF-004 (↔ REQ-TEF-004) — scoped rules excluded, PASS is silent

- **G**: at least one `.claude/rules/moai/**` file carries a `paths:` restriction
  (e.g. `languages/go.md`, `development/agent-authoring.md`).
- **W**: the guard enumerates the surface.
- **T**: no `paths:`-scoped file is counted; when under budget the guard produces
  no test output (silent PASS).
- Verify: a sub-test asserts a known `paths:`-scoped file is absent from the
  enumerated list.

---

## P0-2 — cache-hit-ratio statusline signal

> (Formerly P0-3; AC IDs renumbered 008/009/010 → 005/006/007 after the "~7×"
> reword carve-out.)

### AC-TEF-005 (↔ REQ-TEF-005) — ratio computed when data present

- **G**: stdin JSON with `context_window.current_usage` non-null and
  `cache_creation_input_tokens > 0` (e.g. `cache_read_input_tokens: 2000`,
  `cache_creation_input_tokens: 5000`).
- **W**: `moai statusline` renders (or the derivation function is unit-tested)
  with the cache segment enabled.
- **T**: the statusline surfaces a cache signal derived from the two fields
  (cache-hit % or ratio per plan.md §D-3).
- Verify: a table-driven test row asserts the expected rendered token/percentage
  for a known input pair.

### AC-TEF-006 (↔ REQ-TEF-006) — graceful degradation (null + zero)

- **G**: two inputs — (i) `context_window.current_usage: null`; (ii)
  `current_usage` present but `cache_creation_input_tokens: 0`.
- **W**: the derivation runs on each.
- **T**: for both, NO cache-ratio value is fabricated and NO divide-by-zero
  occurs; the segment is omitted (or context-% shown alone). The statusline still
  renders successfully.
- Verify: table-driven sub-tests `{name:"null-current-usage"}` and
  `{name:"zero-cache-creation"}` assert the segment is empty and no panic.

### AC-TEF-007 (↔ REQ-TEF-007) — config-toggleable segment

- **G**: the statusline segment configuration.
- **W**: the cache segment is disabled via config.
- **T**: the segment does not render; when enabled it renders — consistent with
  existing segment toggle conventions (parallel to `SegmentEffortThinking`).
- Verify: a test toggles the segment config off/on and asserts absence/presence.

---

## Edge cases (must be covered by tests)

- P0-1: repo path unavailable (test running outside tree) → `t.Skip` gracefully
  (model: `TestAuditParity` skip-on-missing-dir).
- P0-1: a rule file with malformed frontmatter → treated as always-loaded
  (conservative — counts it) rather than silently skipped.
- P0-2: `current_usage` present, both cache fields zero → segment omitted (no
  0/0).
- P0-2: very large token counts → no integer overflow in the percentage math.

## Quality gate (Definition of Done)

- [ ] `go test ./...` PASS (incl. new P0-1 guard + P0-2 statusline tests)
- [ ] `go vet ./...` clean
- [ ] `golangci-lint run` no NEW findings
- [ ] `GOOS=windows GOARCH=amd64 go build ./...` exit 0
- [ ] `moai spec lint --strict` on this SPEC dir: no findings
- [ ] Coverage ≥ 85% on any new Go file (per package threshold)
- [ ] All 7 ACs above PASS with cited evidence

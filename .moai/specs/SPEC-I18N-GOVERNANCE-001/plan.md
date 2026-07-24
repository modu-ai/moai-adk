# SPEC-I18N-GOVERNANCE-001 — Implementation Plan

## §A Context

Add value-level governance to `internal/web/assets/i18n.js`: an explicit allowlist of intentionally-untranslated keys, a detector that consumes it, a structural rule for the endonym family, bidirectional key-coverage checks, and a named owning surface. Tier M, three artifacts, no runtime behavior change, no template mirror.

The whole change is confined to package `web` and is consumed by tests only.

## §B Known Issues / Preconditions

- **B1 — The delegating brief's figures are stale relative to this worktree base.** The brief stated 341 keys per locale with zero asymmetry, 22 identical `ko` values, and materially different `ko` / `ja` / `zh` identical sets. Measurement at the worktree base gives `en` = 340, non-`en` = 350, 25 identical values per locale, and sets differing by exactly one key. Re-measure at run-phase entry before writing any fixture; do not transcribe §A.1 as fact.
- **B2 — The `agentdesc.*` asymmetry is deliberate and already test-enforced.** `TestD3AgentDescIsEnExempt` asserts `en` defines no `agentdesc.*` key. A naive symmetric key-set equality check would break a passing, intentional test. Reverse coverage must be equality modulo the exempt-prefix registry (REQ-I18NGOV-019, REQ-I18NGOV-020).
- **B3 — Case-fold normalization adds exactly one new violation.** `mp.col.effort` is `"Effort"` in `en` and `"effort"` in all three non-`en` locales. Exact comparison misses it; normalized comparison catches it. Its sibling columns (`mp.col.tier`, `mp.col.phase`, `mp.col.model`) are all translated, so the expected ruling is *translate*, not *allowlist* — but the ruling is M1 work, not a pre-decided outcome.
- **B4 — Parser edge cases.** Four keys fall outside `[A-Za-z0-9._-]`: `f.report.format.opt.html+md` and the three `f.model.opt.*[1m]` variants. A key regex assuming that character class silently drops them. Separately, zero values currently contain a quote, backslash, or newline escape — a convenient property, but not one the parser may assume, since a future translation can introduce a quotation mark.
- **B5 — An existing splitter is available.** `localeBlocks()` in `internal/web/webux_followup_test.go` already splits the dictionary into four locale blocks by index. Extend it with a key-value parse step; do not write a second splitter.
- **B6 — The allowlist must not be served.** `staticFS()` exposes everything under `internal/web/assets/`. Placing the allowlist there would publish it at `/static/`. Keep it in `internal/web/` proper.
- **B7 — Concurrent session.** Another session is active on the main checkout. All work stays inside this worktree; the orchestrator owns every git operation.

## §C Pre-flight

```bash
# 1. Baseline: the governed file and its existing suite
wc -l internal/web/assets/i18n.js
go test ./internal/web/... 2>&1 | tail -20

# 2. Re-measure the catalogue (B1) — key counts, coverage both directions,
#    identical sets exact and normalized. Do not reuse spec.md §A.1 verbatim.

# 3. Confirm the exemption contract still holds before touching coverage logic
go test -run 'TestD3AgentDesc' ./internal/web/ -v

# 4. Confirm the splitter to extend
grep -n "func localeBlocks" internal/web/*_test.go
```

## §D Constraints

Inherited verbatim from `spec.md` §C (C1-C5). Additionally:

- Every file created lives under `internal/web/`; nothing under `internal/web/assets/` changes except the `i18n.js` header comment (REQ-I18NGOV-021) and any value corrected by an M1 ruling.
- No production (non-`_test.go`) symbol is added. The allowlist, registry, parser, and detector are all test-binary-only, so the shipped console binary is byte-unaffected apart from the `i18n.js` header comment and corrected values.

## §E Self-Verification

Run-phase reports per acceptance criterion with the verbatim command output, plus `go build ./...`, `GOOS=windows GOARCH=amd64 go build ./...`, and `go test ./internal/web/...`.

## §F Milestones

Ordered by decision-reversibility: the milestones most likely to change under review come first.

### M1 — Governance data model and the rulings (decision-heavy — review focus)

The shape chosen here determines everything downstream, and the two rulings are judgment calls.

Deliverables:

- Allowlist artifact at `internal/web/i18n_untranslated_allowlist_test.go` — a typed slice in package `web`, with the reason taxonomy as a Go enum so an invented reason fails to **compile** rather than merely failing a test (REQ-I18NGOV-001 through REQ-I18NGOV-005).
- The inline governance contract as a doc comment on that artifact: location, entry format, who adds an entry, the reviewer assertion, and the ruling procedure (REQ-I18NGOV-006, REQ-I18NGOV-023).
- The `en`-exempt prefix registry with `agentdesc.` as its sole member plus its justification (REQ-I18NGOV-020).
- Rulings for every member of the re-measured identity set: each becomes either an allowlist entry with a taxonomy reason and justification, or a translation fix in `i18n.js`.

Two rulings need explicit reasoning in the run-phase report:

- **`board.badge.mustfix`** — `"MUST-FIX"`. Ruling: `technical-identifier`, allowlist. The literal is the `Severity` value emitted by `internal/spec/audit.go` and matched by `internal/web/board.go` and `internal/cli/spec_audit.go`; the badge exists so the console reads the same token the CLI and its JSON output emit. Translating it would break that correspondence for a user grepping `moai spec audit --json`.
- **`mp.tier.empty`** — `"(runtime default: medium)"`. Ruling: no taxonomy reason applies; it is English prose whose sibling `mp.tier.default` is translated in all three locales. Fix by translating, not by allowlisting.

Design decision to surface at review: the allowlist is per-key and applies to all non-`en` locales (REQ-I18NGOV-005). The evidence is that once `lang.opt.*` is handled structurally, the three locales' identical sets are byte-identical at 24 keys, so per-locale entries would triple the artifact with no added signal. The residual risk — a global entry masking a genuinely-untranslated copy of that same key in one locale — is accepted and bounded by the orphan check and the entry cap.

### M2 — Parser and detector (interface decisions)

- Extend `localeBlocks()` with a per-block key-value parse producing `map[string]map[string]string` (REQ-I18NGOV-007). Handle the four special-character keys of B4 and tolerate escaped characters in values.
- Implement the detector as a pure function over `(catalogue, allowlist)` returning a violation slice (REQ-I18NGOV-008), with NFC + trim + casefold normalization (REQ-I18NGOV-009) and the `lang.opt.*` exclusion (REQ-I18NGOV-011).
- Implement the orphan check inside the same pure function or a sibling pure function (REQ-I18NGOV-015).

### M3 — Endonym and coverage checks

- The bidirectional endonym invariant: self-equality (REQ-I18NGOV-012) and cross-locale distinctness (REQ-I18NGOV-013), plus the assertion that no `lang.opt.*` key appears in the allowlist (REQ-I18NGOV-014).
- Forward coverage (REQ-I18NGOV-018) and reverse coverage modulo the exempt registry (REQ-I18NGOV-019).

### M4 — Tests, including the negative control (mechanical)

- Real-catalogue tests: detector reports zero violations, zero orphans, entry count within bound, endonym invariants hold, coverage holds both directions.
- Negative control (REQ-I18NGOV-017): a synthetic catalogue containing an inert placeholder key whose non-`en` value equals its `en` value produces a violation with an empty allowlist, and produces none once that key is allowlisted. This is the proof the allowlist is not a rubber stamp; it is why the detector had to be a pure function in M2.
- Wildcard rejection (REQ-I18NGOV-004) and the entry cap (REQ-I18NGOV-016).

### M5 — Header, and verification (mechanical)

- Add the owner line and governance-contract pointer to the `i18n.js` header comment (REQ-I18NGOV-021).
- Full verification batch; confirm the pre-existing suite is untouched.

## §G Anti-Patterns

- Writing a symmetric key-set equality check and then editing `TestD3AgentDescIsEnExempt` to accommodate it. The exemption is the correct design; the check bends to it (B2).
- Making the detector read the shipped files directly, which makes the negative control impossible to write and reduces the anti-rubber-stamp criterion to a tautology.
- Allowlisting `mp.tier.empty` because it appears in the identity set. It is the defect the mechanism exists to find.
- Adding a `lang.opt.*` key to the allowlist instead of relying on the endonym invariant.
- Transcribing `spec.md` §A.1 into a test as a hard-coded expected count. The counts are a plan-phase baseline, not an invariant; asserting them freezes the catalogue against legitimate growth.
- Placing the allowlist under `internal/web/assets/`, which publishes it at `/static/`.

## §H File Impact Map

| File | Change | Milestone |
|---|---|---|
| `internal/web/i18n_untranslated_allowlist_test.go` | new — allowlist, reason enum, exempt-prefix registry, governance contract | M1 |
| `internal/web/assets/i18n.js` | edit — header owner line; translation fixes from M1 rulings | M1, M5 |
| `internal/web/i18n_governance_test.go` | new — parser extension, detector, all checks, negative control | M2, M3, M4 |
| `internal/web/webux_followup_test.go` | read-only reference for `localeBlocks()`; unchanged unless the splitter is refactored in place | M2 |

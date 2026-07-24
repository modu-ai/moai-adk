# SPEC-I18N-GOVERNANCE-001 — Acceptance Criteria

All commands are run from the repository root. Every criterion is binary: the named command exits 0, or the criterion fails.

## §D Acceptance Matrix

| AC | Subject | REQ coverage |
|----|---------|--------------|
| AC-I18NGOV-001 | Untranslated-value detector is green on the real catalogue | 007, 008, 009, 010, 011 |
| AC-I18NGOV-002 | Negative control — a non-allowlisted untranslated key fails | 008, 017 |
| AC-I18NGOV-003 | Allowlist has no orphan entries | 015 |
| AC-I18NGOV-004 | Allowlist shape — exact keys, closed taxonomy, bounded size | 002, 003, 004, 005, 016 |
| AC-I18NGOV-005 | Endonym family is correct by construction | 012, 013, 014 |
| AC-I18NGOV-006 | Forward key coverage: every `en` key in all three locales | 018 |
| AC-I18NGOV-007 | Reverse key coverage modulo the exempt registry | 019, 020 |
| AC-I18NGOV-008 | Parser reads special-character keys | 007 |
| AC-I18NGOV-009 | Governance contract and named owner are present | 001, 006, 021, 023 |
| AC-I18NGOV-010 | Pre-existing suite and the `agentdesc.*` exemption still pass | C1, 022 |
| AC-I18NGOV-011 | No runtime change; cross-platform build clean | C4, C5 |

---

## AC-I18NGOV-001 — Untranslated-value detector is green on the real catalogue

**Given** the shipped `internal/web/assets/i18n.js` and the shipped allowlist,
**when** the detector compares every non-`en` value to its `en` counterpart under NFC + trim + casefold normalization, excluding the `lang.opt.*` family and every allowlisted key,
**then** it reports zero violations.

```bash
go test -run TestI18nUntranslatedValues ./internal/web/ -v
```

Passing requires that every member of the re-measured identity set has received a ruling: it is either an allowlist entry or no longer identical. Reaching green by widening the comparison (dropping case folding, skipping a key prefix) rather than by ruling on each key fails AC-I18NGOV-004 and AC-I18NGOV-002.

## AC-I18NGOV-002 — Negative control: a non-allowlisted untranslated key fails

**Given** a synthetic in-memory catalogue containing an inert placeholder key whose `ko` value is byte-identical to its `en` value,
**when** the detector runs against that catalogue with an allowlist that does **not** contain the key,
**then** it reports at least one violation naming that key and the `ko` locale; **and when** the identical catalogue is passed with an allowlist that **does** contain the key, it reports zero violations.

```bash
go test -run TestI18nUntranslatedDetectorNegativeControl ./internal/web/ -v
```

This is the criterion that proves the allowlist is a gate and not a rubber stamp. It is only satisfiable if the detector is a pure function over its inputs (REQ-I18NGOV-008) — a detector that reads the shipped files cannot be driven by a synthetic catalogue, and a test that only asserts the real catalogue is green would be a tautology. The synthetic fixture uses an inert placeholder key and a short interface word; it contains no secret-shaped or key-shaped literal (C2).

## AC-I18NGOV-003 — Allowlist has no orphan entries

**Given** the shipped allowlist,
**when** each entry is resolved against the shipped catalogue,
**then** every entry names a key that exists in the catalogue **and** whose value is identical to `en` in at least one locale.

```bash
go test -run TestI18nAllowlistNoOrphans ./internal/web/ -v
```

An entry surviving the deletion or translation of its key is a silent blanket exemption; this criterion forces its removal in the same change.

## AC-I18NGOV-004 — Allowlist shape: exact keys, closed taxonomy, bounded size

**Given** the shipped allowlist,
**when** its entries are inspected,
**then** no entry key contains `*`, `?`, or a regular-expression metacharacter; every entry carries a non-empty justification; every entry's reason is one of `technical-identifier`, `proper-noun`, `acronym`; no entry key begins with `lang.opt.`; and the entry count is at most 30.

```bash
go test -run TestI18nAllowlistShape ./internal/web/ -v
```

The reason taxonomy is additionally closed at compile time — the reason field is a Go enum type, so an invented reason fails `go build` before any test runs:

```bash
go vet ./internal/web/
```

## AC-I18NGOV-005 — Endonym family is correct by construction

**Given** the four locale blocks,
**when** the `lang.opt.*` family is checked,
**then** for every locale `L` the value of `lang.opt.<L>` in `L` equals its `en` value, and for every pair where language `X` differs from locale `L` the value of `lang.opt.<X>` in `L` differs from its `en` value.

```bash
go test -run TestI18nEndonymInvariants ./internal/web/ -v
```

The family is covered by this positive bidirectional invariant and by zero allowlist entries, so an English endonym copied into the wrong locale is caught rather than exempted. AC-I18NGOV-004 independently asserts the allowlist carries no `lang.opt.` key.

## AC-I18NGOV-006 — Forward key coverage

**Given** the `en` locale block,
**when** each of its keys is looked up in `ko`, `ja`, and `zh`,
**then** every key is present in all three.

```bash
go test -run TestI18nKeyCoverageForward ./internal/web/ -v
```

## AC-I18NGOV-007 — Reverse key coverage modulo the exempt registry

**Given** the `ko`, `ja`, and `zh` locale blocks,
**when** each key absent from `en` is matched against the `en`-exempt prefix registry,
**then** every such key matches a declared prefix, and every registry member carries a justification naming the surface that supplies the English baseline instead.

```bash
go test -run TestI18nKeyCoverageReverse ./internal/web/ -v
```

The registry's sole initial member is `agentdesc.`. A new unexplained reverse-asymmetry fails this criterion; adding a prefix to the registry is a reviewed act, not a silent one.

## AC-I18NGOV-008 — Parser reads special-character keys

**Given** the shipped catalogue,
**when** the parser produces the per-locale key-value maps,
**then** the maps contain `f.report.format.opt.html+md`, `f.model.opt.opus[1m]`, `f.model.opt.sonnet[1m]`, and `f.model.opt.fable[1m]` with non-empty values in every locale that defines them.

```bash
go test -run TestI18nParserSpecialKeys ./internal/web/ -v
```

A key regex assuming `[A-Za-z0-9._-]` drops these four keys silently, which would shrink the detector's input without failing anything else.

## AC-I18NGOV-009 — Governance contract and named owner are present

**Given** the allowlist artifact and the `i18n.js` header,
**when** they are inspected,
**then** the allowlist artifact carries an inline contract stating its location, entry format, who adds an entry, the reviewer assertion, and the ruling procedure; and the `i18n.js` header names the allowlist artifact path as the owning governance surface.

```bash
go test -run TestI18nGovernanceContractPresent ./internal/web/ -v
```

Supporting evidence, quoted verbatim in the run-phase report:

```bash
grep -n "i18n_untranslated_allowlist" internal/web/assets/i18n.js
```

## AC-I18NGOV-010 — Pre-existing suite and the `agentdesc.*` exemption still pass

**Given** the package test suite as it stood before this SPEC,
**when** it is re-run after the change,
**then** every test passes, including `TestD3AgentDescIsEnExempt` and `TestD3AgentDescKeysInThreeLocales`, neither of which has been edited.

```bash
go test ./internal/web/... 2>&1 | tail -20
go test -run 'TestD3AgentDesc' ./internal/web/ -v
git diff --stat -- internal/web/webux_followup_test.go
```

The third command's expected output is either empty or a diff limited to the `localeBlocks()` parse extension. Any change to the two named test bodies violates C1.

## AC-I18NGOV-011 — No runtime change; cross-platform build clean

**Given** the change set,
**when** the project is built for both host and Windows targets,
**then** both succeed, and the only non-test file modified is `internal/web/assets/i18n.js`.

```bash
go build ./...
GOOS=windows GOARCH=amd64 go build ./...
git diff --name-only | grep -v '_test\.go$'
```

The third command's output is expected to be exactly `internal/web/assets/i18n.js` — the header comment plus any value corrected by an M1 ruling. Any other non-test file indicates production code was added where test-only artifacts were specified (C4).

---

## Definition of Done

- AC-I18NGOV-001 through AC-I18NGOV-011 all pass with verbatim command output quoted in the run-phase report.
- Every member of the re-measured identity set has an explicit ruling — an allowlist entry with a taxonomy reason and justification, or a translation fix — and the two rulings called out in `plan.md` §F M1 (`board.badge.mustfix`, `mp.tier.empty`) are reported with their reasoning.
- `go test ./internal/web/...` is green.
- No secret-shaped or key-shaped literal appears in any fixture (C2).
- The SPEC's own lint is clean: `go run ./cmd/moai spec lint .moai/specs/SPEC-I18N-GOVERNANCE-001/spec.md` reports no findings.

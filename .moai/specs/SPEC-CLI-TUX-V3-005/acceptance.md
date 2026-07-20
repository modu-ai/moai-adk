# acceptance.md — SPEC-CLI-TUX-V3-005

> **Acceptance criteria for AC-TUX3-020 Printer migration.** The headline ratchet AC (AC-TUX3-020) is inherited verbatim from SPEC-CLI-TUX-V3-003; the remaining ACs cover Printer-method usage, behavior preservation, and coverage.

---

## §A. Verification Command (canonical ratchet grep)

```bash
grep -rn 'fmt\.Printf\|fmt\.Println\|fmt\.Print(' internal/cli --include='*.go' | grep -v '_test.go' | wc -l
```

**Pre-flight baseline (measured 2026-07-14)**: **38**

This is the SAME command defined in SPEC-CLI-TUX-V3-003 acceptance.md:33. The baseline value is recorded verbatim here for self-containment; the authoritative pre-flight measurement will be re-run at run-phase entry and recorded in progress.md §E.2.

---

## §B. AC Severity Scale

| Severity | Meaning |
|----------|---------|
| MUST-PASS | Blocks SPEC close. Failure = run-phase cannot complete. |
| SHOULD-PASS | Strong expectation. Failure requires documented debt rationale. |
| NICE-TO-HAVE | Improvement. Failure is non-blocking. |

---

## §C. Traceability

Each AC traces to a REQ in spec.md §B and is verified by an observable command.

---

## §D. AC Matrix

### AC-TUX3-020 — Ratchet succession (inherited, MUST-PASS)

| Field | Value |
|-------|-------|
| **REQ** | REQ-TUX3-020 |
| **Severity** | MUST-PASS |
| **Command** | `grep -rn 'fmt\.Printf\|fmt\.Println\|fmt\.Print(' internal/cli --include='*.go' \| grep -v '_test.go' \| wc -l` |
| **Pass condition** | Result is **strictly less than 38** (the pre-flight baseline) |
| **Evidence** | Verbatim command output quoted in progress.md §E.2 (vci §3 5-section format: Claim + Evidence + Baseline-attribution + Gaps + Residual-risk) |
| **Baseline attribution** | Baseline 38 measured 2026-07-14 by orchestrator investigation. Re-measured at run-phase entry. The baseline is the RAW grep count (includes commented-out lines matched by grep, e.g. `worktree/new.go:433`). |
| **Provenance** | Inherited from SPEC-CLI-TUX-V3-003 acceptance.md:33. Split as debt during SPEC-003 amendment re-close. |

### AC-TUX3-021 — Printer-method usage at migrated sites (MUST-PASS)

| Field | Value |
|-------|-------|
| **REQ** | REQ-TUX3-021 |
| **Severity** | MUST-PASS |
| **Command** | `grep -rn 'fmt\.Printf\|fmt\.Println\|fmt\.Print(' internal/cli/state.go internal/cli/migration.go internal/cli/worktree/tmux_integration.go \| grep -v '_test.go' \| wc -l` |
| **Pass condition** | Result is **0** for all files determined to be in-scope by the M1 architecture gate (state.go, migration.go, tmux_integration.go are the default in-scope set; banner.go and branch_protection.go are in-scope ONLY if M1 approves their migration) |
| **Evidence** | Verbatim grep output per file, showing 0 remaining `fmt.Print*` calls at migrated sites |

### AC-TUX3-022 — Behavior preservation via characterization tests (MUST-PASS)

| Field | Value |
|-------|-------|
| **REQ** | REQ-TUX3-022 |
| **Severity** | MUST-PASS |
| **Command** | (1) `go test -run 'Test.*Characterization\|Test.*Output\|Test.*Print' ./internal/cli/... -v` (all pass) AND (2) `go test -list -run 'Test.*Characterization\|Test.*Output\|Test.*Print' ./internal/cli/... 2>&1 \| grep -c '^Test'` ≥ 1 (guards vacuous pass — a `-run` pattern matching zero tests exits 0). Pin specific test function names after M2-M4 characterization tests are written. |
| **Pass condition** | All characterization tests pass. Each migrated function has at least one characterization test capturing pre-migration stdout/stderr output. Channel re-routing (stdout→stderr for status messages) is documented as an intentional, tested behavior change — NOT a silent regression. |
| **Evidence** | Test output showing all characterization tests green; test source files listing |

### AC-TUX3-023 — Printer package + migrated file coverage (SHOULD-PASS)

| Field | Value |
|-------|-------|
| **REQ** | REQ-TUX3-023 |
| **Severity** | SHOULD-PASS |
| **Command** | `go test -cover ./internal/cli/printer/...` + `go test -cover ./internal/cli/...` |
| **Pass condition** | `internal/cli/printer` coverage ≥ 85.0%. Each migrated file's package coverage ≥ pre-migration baseline (non-regression). |
| **Evidence** | Coverage percentage output per package |

### AC-TUX3-024 — Gap-site disposition recorded (MUST-PASS)

| Field | Value |
|-------|-------|
| **REQ** | REQ-TUX3-024 |
| **Severity** | MUST-PASS |
| **Command** | Per-gap-site anchors (each gap site must have a specific disposition in spec.md §C): `grep -c 'banner.go' .moai/specs/SPEC-CLI-TUX-V3-005/spec.md` ≥ 1 AND `grep -c 'branch_protection' .moai/specs/SPEC-CLI-TUX-V3-005/spec.md` ≥ 1. Replaces the generic keyword grep which any unrelated "excluded" occurrence satisfied. |
| **Pass condition** | Every gap site identified in research.md §D (banner.go, branch_protection.go) has an explicit disposition recorded in spec.md §C (Out of Scope) or plan.md §F M1. No gap site is silently left without a rationale. |
| **Evidence** | spec.md Out-of-Scope section + plan.md M1 decision block |

---

## §D.1 Coverage Detail

| Package | Pre-migration coverage | Target | Non-regression? |
|---------|----------------------|--------|-----------------|
| `internal/cli/printer` | (existing, SPEC-001) | ≥ 85.0% | Maintain or improve |
| `internal/cli` (state.go) | (measure at M2 start) | ≥ pre-migration | Yes |
| `internal/cli` (migration.go) | (measure at M3 start) | ≥ pre-migration | Yes |
| `internal/cli/worktree` (tmux) | (measure at M4 start) | ≥ pre-migration | Yes |

---

## §D.2 Indirect Verification

The ratchet AC (AC-TUX3-020) is a COUNT-based gate — it verifies that the number of `fmt.Print*` calls decreased, but does NOT verify that the migrated calls produce correct output. The companion ACs (AC-TUX3-021 method-usage, AC-TUX3-022 behavior-preservation) provide the qualitative verification:

- AC-TUX3-020 = "did the count go down?" (quantitative)
- AC-TUX3-021 = "are the migrated sites using Printer methods?" (structural)
- AC-TUX3-022 = "does the output still work?" (qualitative)

All three MUST pass for the migration to be considered correct.

---

## §D.3 Closure Gates

| Gate | Condition |
|------|-----------|
| Run-phase exit | AC-TUX3-020, 021, 022, 024 MUST-PASS; AC-TUX3-023 SHOULD-PASS |
| Sync-phase entry | All MUST-PASS ACs verified with verbatim evidence |
| SPEC close | Ratchet count recorded in progress.md §E.2; gap-site dispositions documented |

---

## §D.4 Edge Cases

| Edge case | Expected behavior |
|-----------|-------------------|
| `--json` flag on `moai state show` / `moai migration status` | Output routes through `p.Data()` producing JSON line on stdout in all modes (Data() always marshals to JSON in ModeJSON, prints verbatim in text modes) |
| NO_COLOR environment set | Printer resolves ModeTTY→ModePlain; migrated output carries no ANSI sequences (same as pre-migration `fmt.Print*` which had none) |
| Non-TTY (pipe/CI) | Printer resolves ModePlain; one line per event, no in-place updates |
| Empty result sets ("No blockers found", "No pending migrations") | Routes to `p.Info()` on stderr; exit code unchanged |

---

## §D.5 Forward-Looking Checks

| Check | Purpose |
|-------|---------|
| Ratchet succession is STRICTLY DECREASING | Future SPECs that add `fmt.Print*` calls must migrate an equal or greater number to maintain the ratchet. The baseline for the next SPEC is whatever count this SPEC closes at. |
| Gap-site dispositions are REVIEWABLE | If a future SPEC activates `ttyConfirmer` (branch_protection.go), the Printer may need a Prompt method — the disposition recorded here should be revisited. |

---

## §E. Given-When-Then Scenarios

### Scenario 1: Ratchet verification (AC-TUX3-020)

```
Given the pre-flight baseline of 38 fmt.Print* calls in non-test internal/cli/
When the migration milestones M2-M4 are complete
Then the canonical grep produces a count strictly less than 38
And the verbatim count is recorded in progress.md §E.2
```

### Scenario 2: state.go JSON output preservation (AC-TUX3-022)

```
Given moai state show --json produces a JSON object on stdout
When state.go is migrated to use p.Data(state)
Then the JSON output is preserved byte-for-byte in ModeJSON
And the output channel remains stdout (Data writes to stdout)
And the characterization test capturing pre-migration JSON output passes
```

### Scenario 3: status message channel re-routing (AC-TUX3-022)

```
Given "No checkpoint found" is currently printed to stdout via fmt.Printf
When state.go is migrated to use p.Info(...)
Then the message routes to stderr (Info writes to stderr)
And the channel change is documented as intentional in the characterization test
And the message text is preserved
```

### Scenario 4: tmux session output migration (AC-TUX3-021)

```
Given tmux_integration.go prints 5 session-info lines via fmt.Printf to stdout
When the function is migrated to accept a printer.Printer parameter
Then all 5 lines use p.Info(...) routing to stderr
And zero fmt.Print* calls remain in tmux_integration.go
And the grep for tmux_integration.go returns 0
```

---
id: SPEC-ERRCHECK-CATALOG-HASH-001
title: "errcheck: discard fmt.Fprintf return value in catalog tree hash"
version: "0.1.0"
status: completed
created: 2026-09-06
updated: 2026-09-06
author: manager-spec
priority: P2
phase: "v3.2.0 target"
module: "internal/template"
lifecycle: spec-anchored
tier: M
tags: "lint,errcheck,ci,template"
---

# SPEC-ERRCHECK-CATALOG-HASH-001 — errcheck: discard `fmt.Fprintf` return in catalog tree hash

## HISTORY

| Date | Version | Change |
|------|---------|--------|
| 2026-09-06 | 0.1.0 | Initial draft — plan-phase artifact set authored by manager-spec (card t489, lane-15) |

## §1 Overview

`golangci-lint` (both CI run 34014859880, job Lint, golangci-lint v2.1.6, and a local run with
golangci-lint v2.10.1 against the same committed `.golangci.yml`) reports exactly one finding:

```
internal/template/catalog_tree_hash.go:60:14: Error return value of `fmt.Fprintf` is not checked (errcheck)
```

Local verbatim summary: `1 issues: * errcheck: 1`. No other linter findings exist in the package.

The flagged call at `internal/template/catalog_tree_hash.go:60` writes digest input into
`h := sha256.New()` (line 58). `sha256.New` returns `hash.Hash`, whose `Write` method is documented
in the Go standard library as never returning an error. The unchecked error is therefore
structurally always nil, and the repair is an explicit discard, not error handling.

## §2 Requirements (GEARS)

### REQ-001 — Zero errcheck findings (Ubiquitous)

The `internal/template` package shall produce zero errcheck findings under
`golangci-lint run ./internal/template/... --timeout=2m` with the committed `.golangci.yml` configuration.

### REQ-002 — Explicit discard at the flagged call site (Event-driven)

**When** the catalog tree hash writes digest input into a `hash.Hash` writer via `fmt.Fprintf`,
the code at `internal/template/catalog_tree_hash.go` shall explicitly discard the return value
(`_ = fmt.Fprintf(h, "%s:%s\n", e.rel, e.sum)`) accompanied by a brief English comment citing the
`hash.Hash` never-error contract.

### REQ-003 — No dead error-propagation path (Unwanted)

The implementation shall not propagate, wrap, or branch on the `fmt.Fprintf` error return at the
flagged call site — the error is structurally always nil, so any propagation path is dead code.

### REQ-004 — Baseline-first commit ordering

**When** the fix commit lands, the RED baseline artifact (`.moai/reports/t489/red-baseline.md`,
measured by lane-15 in this run) shall have been committed in a commit that precedes the fix
commit, so the commit graph witnesses the baseline-first ordering (verification-claim-integrity
§2.3).

### REQ-005 — Verification scope is the linter only

The run-phase verification shall be the golangci-lint command alone; no new test shall be added.
Rationale: a `_ =` discard cannot alter program semantics (the discarded value is a compile-time
expression result, not behavior), so the change is behavior-preserving by construction and the
existing package tests remain the behavior evidence.

### REQ-006 — Evidence and traceability

Evidence artifacts shall land under `.moai/reports/t489/`, and every commit on the card branch
shall carry the card id `t489` in its message.

## §3 Acceptance Criteria (summary — full Given-When-Then matrix in acceptance.md)

| AC | Verdict source |
|----|----------------|
| AC-GREEN | errcheck match count 0 on the fixed tree (file-scoped line AND package summary `* errcheck: 0`) |
| AC-MUTANT | errcheck match count 1 with the fix reverted (pre-fix blob restored) |
| AC-ORDERING | baseline commit precedes fix commit in the branch history |
| AC-SCOPE | verification command set = linter only; no new test file added; diff shape = `_ =` discard (no `//nolint` in `git diff -- '*.go'`) |
| AC-EVIDENCE | artifacts under `.moai/reports/t489/`; `t489` in every commit message |

## Out of Scope — Exclusions

### Out of Scope — Error-handling redesign of catalog_tree_hash.go

- No propagation, wrapping, or logging of the `fmt.Fprintf` error at the flagged call site (REQ-003).
- No refactor of the surrounding `CatalogTreeHash` function beyond the single flagged line.

### Out of Scope — Linter configuration changes

- No change to `.golangci.yml`; the file already pins `errcheck.check-blank: false` (lines 31-35),
  which is what makes the `_ =` discard pass, and already documents the intentional
  local-v2.10.1 vs CI-v2.1.6 version skew as safe because the linter set and errcheck settings
  are pinned identically.
- No exclusion directives (`//nolint`, `issues.exclude-*`) to silence the finding.

### Out of Scope — Test authoring

- No new unit test: the fix adds no behavior change (see REQ-005 rationale).

### Out of Scope — Other linter findings

- Only the single measured errcheck finding is in scope; the package currently carries no others
  (measured in this run, both CI and local).

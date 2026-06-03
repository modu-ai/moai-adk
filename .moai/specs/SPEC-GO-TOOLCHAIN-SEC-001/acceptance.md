# Acceptance Criteria — SPEC-GO-TOOLCHAIN-SEC-001

## D. AC Matrix

| AC ID | Requirement | Verification command | Expected result | Severity |
|-------|-------------|----------------------|-----------------|----------|
| AC-GTS-001 | REQ-GTS-001 | `govulncheck ./...` (under go1.26.4) | "No vulnerabilities found" OR 0 affecting findings (19 stdlib → 0) | MUST |
| AC-GTS-002 | REQ-GTS-005 | `go build ./...` ; `go test ./...` (under go1.26.4) | both exit 0 | MUST |
| AC-GTS-003 | REQ-GTS-002, REQ-GTS-003 | inspect `go.mod` `go` directive + `grep -rn 'go-version: *"1.26"' .github/workflows/` | `go` directive ≥ 1.26.4 AND grep returns 0 matches (all 9 migrated to `go-version-file`) | MUST |
| AC-GTS-004 | REQ-GTS-004 | `git diff go.mod` | only the `go` (+ new `toolchain`) directive lines changed; zero `require` block lines modified | MUST |

## D.1 Given-When-Then Scenarios

### Scenario 1 — Toolchain bump clears all stdlib findings (AC-GTS-001, AC-GTS-002)

```
GIVEN  go.mod declares `go 1.26` and govulncheck reports 19 affecting stdlib vulns
WHEN   go.mod is bumped to `go 1.26.4` + `toolchain go1.26.4`, and the effective
       toolchain (verified by `go version`) is go1.26.4 (auto-acquired via GOTOOLCHAIN=auto)
THEN   `govulncheck ./...` reports 0 affecting vulnerabilities
AND    `go build ./...` exits 0
AND    `go test ./...` exits 0
```

### Scenario 2 — CI pin strategy applied consistently, scope held (AC-GTS-003, AC-GTS-004)

```
GIVEN  9 workflow steps hardcode `go-version: "1.26"` and 3 already use `go-version-file: go.mod`
WHEN   the 9 hardcoded steps are migrated to `go-version-file: go.mod`
THEN   `grep -rn 'go-version: *"1.26"' .github/workflows/` returns 0 matches
AND    the 3 pre-existing `go-version-file` workflows are unchanged
AND    `git diff go.mod` shows only the `go`/`toolchain` directive lines changed
AND    no third-party `require` block line is modified
```

## D.2 Edge Cases

- **EC-1 — Effective toolchain still go1.26.0**: If `GOTOOLCHAIN=local` is set in the
  environment, `go build`/`go test`/`govulncheck` run under go1.26.0 and govulncheck
  would still report the 19 findings. The run-phase MUST verify `go version` ≥ go1.26.4
  BEFORE trusting any "0 affecting" result. A "0 affecting" captured under go1.26.0 is a
  false PASS and MUST be rejected.
- **EC-2 — A higher fix version appears**: If govulncheck after the go1.26.4 bump still
  reports a residual finding fixed only at a higher patch (e.g., a finding published after
  this SPEC was authored), bump to that higher patch instead — the SPEC target is "0
  affecting", not literally "1.26.4". Document the actual landing version in progress.md.
- **EC-3 — Go API deprecation under the new toolchain**: If `go build ./...` fails under
  go1.26.4 due to a removed/changed stdlib API, STOP and return a blocker — source-code
  changes are out of scope and would be a separate SPEC.

## D.3 Quality Gate Criteria

- All four MUST-severity ACs PASS.
- No source-code (`.go`) file modified (the diff is limited to `go.mod` + 9 workflow YAML
  files).
- Evidence for AC-GTS-001 and AC-GTS-002 captured under a verified go1.26.4 effective
  toolchain (`go version` output included in progress.md).

## D.4 Definition of Done

- [ ] AC-GTS-001 — `govulncheck ./...` = 0 affecting (evidence under go1.26.4)
- [ ] AC-GTS-002 — `go build ./...` and `go test ./...` pass (under go1.26.4)
- [ ] AC-GTS-003 — `go.mod` `go` ≥ 1.26.4 AND 0 residual `go-version: "1.26"` literals
- [ ] AC-GTS-004 — `git diff go.mod` shows only directive lines; no `require` change
- [ ] CHANGELOG security entry added (sync-phase)
- [ ] frontmatter status transitioned (draft → in-progress → implemented)
- [ ] No PR (Tier S main-direct per CLAUDE.local.md §23); commits pushed to main

## D.5 Closure Gates (4-phase close)

- Plan-phase: this SPEC's 4 artifacts created, status: draft. (manager-spec)
- Run-phase: M1–M3 complete, all 4 MUST ACs PASS, status → in-progress. (manager-develop)
- Sync-phase: CHANGELOG entry, status → implemented. (manager-docs)
- Mx-phase: audit-ready signal + 4-phase close, status → completed.

## D.6 Forward-Looking Checks (out of scope here, noted for traceability)

- The `vitest` Dependabot alerts in `moai-adk-ts/` remain open — a separate follow-up for
  the TypeScript project. NOT closed by this SPEC.
- A recurring govulncheck CI gate is NOT added by this SPEC; if desired later, that is a
  separate SPEC (the user chose the minimal option).

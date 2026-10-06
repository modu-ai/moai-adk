---
id: SPEC-CI-STRUCTURAL-RED-001
title: "Plan — CI structural red unblock"
version: "0.1.0"
created: 2026-10-07
updated: 2026-10-07
---

# Plan — SPEC-CI-STRUCTURAL-RED-001

## A.1 Milestones

- **M1 — windows build guard** (REQ-CIS-001/-002, AC-CIS-001/-002):
  `internal/runtime/audit_counter_fifo_test.go` gains `//go:build !windows`
  and a new `audit_counter_fifo_windows_test.go` (`//go:build windows`)
  carries the same test name with `t.Skip`, so the name stays in the
  swept set on every platform. RED already observed at cad44a751.
- **M2 — codemaps stamp re-anchor** (REQ-CIS-003, AC-CIS-003): rebuild
  the moai binary from this tree, regenerate/stamp codemaps provenance
  at the branch tip (`moai graph stamp codemaps` after any needed
  regeneration), so the stamped commit is reachable from the merge
  preview once merged.

## A.5 PRESERVE

- `.github/workflows/**` — the sg guard stays untouched (measured
  innocent).
- `internal/runtime/audit_counter.go` (non-test) — behavior unchanged.
- Every other codemaps content file — only `provenance.json` is
  re-anchored.

## A.6 Verification plan

- `GOOS=windows go vet ./internal/runtime/` → exit 0 (flips AC-CIS-001).
- `GOOS=windows go test ./internal/runtime/ -run TestCountAuditRounds` →
  the skip line (flips AC-CIS-002).
- `git merge-base --is-ancestor "$(jq -r .commit_sha
  .moai/project/codemaps/provenance.json)" HEAD` → exit 0, stamped SHA =
  branch tip (flips AC-CIS-003).
- darwin regression: `go test ./internal/runtime/` stays green.

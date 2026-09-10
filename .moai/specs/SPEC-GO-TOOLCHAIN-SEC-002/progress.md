# Progress — SPEC-GO-TOOLCHAIN-SEC-002

## Status: draft (plan-phase)

Tier S, Class C (global change). Card t610 (Factory lane-8), branch `WT-go-1266`, base
`d3b7d438d` (local develop at card creation; local develop has since moved to `d1b61005d`).

## §E.1 Plan-phase Audit-Ready Signal

- SPEC ID self-check (executed): `ID="SPEC-GO-TOOLCHAIN-SEC-002"; [[ "$ID" =~ ^SPEC(-[A-Z][A-Z0-9]*)+-[0-9]{3}$ ]] && echo PASS || echo FAIL` → `PASS` (re-run in revision 0.1.1 → `PASS`)
- ID uniqueness: only `SPEC-GO-TOOLCHAIN-SEC-001` existed under `.moai/specs/`, and `ls -d .moai/specs/SPEC-GO-TOOLCHAIN-SEC-002` → `No such file or directory`
- Baseline evidence: `.moai/reports/t610/baseline/` (go.mod state, CI go-version-file refs, doc mentions, toolchain versions, govulncheck ×3 toolchains + exit codes). Command record: `.moai/reports/t610/baseline/commands.md`
- RED-now ledger: acceptance.md § D.0 (E-01..E-08, pinned `d3b7d438d2c9bc041cb3b63ea41f9f1a03e867b1`)
- Decisions awaiting Implementation Kickoff Approval: plan.md § D-DESIGN D1–D4
- D2 delta 1.26.6→1.26.8 (plan.md § D2): no security fixes. No 1.26.8 fix reaches the shipped binary (`CGO_ENABLED=0`; no `debug/elf` import; linux/darwin/windows targets only). go1.26.7 net/http #80927 needs unencrypted HTTP/2, which this repo does not configure (`git grep -n -E 'UnencryptedHTTP2|h2c|x/net/http2' -- '*.go' go.mod` → no output; judged on stdout, the exit value was not independently observable through the tool)
- Plan-audit iteration 1: FAIL (3 blocking, 15 non-blocking), report `.moai/reports/t610/plan-audit.md`. Revision 0.1.1 addresses B1–B3, N1–N13, and N15; N14 is kept as a recorded Gap (plan.md § D3)
- Revision 0.1.1 lint (executed): `moai spec lint .moai/specs/SPEC-GO-TOOLCHAIN-SEC-002` → `✓ No findings — all SPEC documents are valid`, exit 0
- Plan-phase Gaps:
  - the upstream issue diffs for go1.26.7/1.26.8 were not read
  - Go's `http.Server` default for unencrypted HTTP/2 was not measured, so the #80927 impact is a Gap; the `internal/web` tests in AC-GTS2-006 are a net/http regression guard, not coverage of that path
  - how `go-version-file` treats a `toolchain` directive was not verified (plan.md § D3)
  - the version half of AC-GTS2-005's RED is observed only on the installed binary, not on a build from this tree (acceptance.md E-06)
  - whether `make build` modifies tracked generated files was not observed; M3 records it
- Closed in 0.1.1: the baseline govulncheck invocation string is recorded in `commands.md` (acceptance.md E-05)

```yaml
plan_status: audit-ready
plan_complete_at: 2026-09-10
plan_revision: "0.1.2"
tier: S
req_count: 8
ac_count: 8
```

## §E.2 Run-phase Evidence

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_

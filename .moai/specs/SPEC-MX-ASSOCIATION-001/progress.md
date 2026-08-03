# progress.md — SPEC-MX-ASSOCIATION-001

> Plan-phase skeleton. Run/sync evidence is populated by manager-develop (§E.2/§E.3) and manager-docs (§E.4). This agent emits only §E.1.

## §E.1 Plan-phase Audit-Ready Signal

- SPEC ID regex self-check: `SPEC-MX-ASSOCIATION-001` → PASS (executed 2026-08-04).
- Frontmatter: 12 canonical fields present; `status: draft`; `module: internal/mx`; `tier: M`.
- Gap verified in code: scanner.go:35 (`recognizedSubLineKinds["SPEC"]`), scanner.go:110-139 (`errSubLineKind` branch pairs only REASON/UPGRADE), scanner.go:321-323 (sub-line sentinel return), tag.go (no SpecRef field). Cited in spec.md §B.2.
- Coverage baseline measured: 9.7 % (955 / 9,858 tags), 88 `@MX:SPEC` sub-lines contributing 0 associations. Cited in spec.md §B.3.
- Out of Scope section present with 6 `### Out of Scope — <topic>` H3 sub-headings.
- Artifact set (Tier M): spec.md, plan.md, acceptance.md, progress.md.

## §E.2 Run-phase Evidence

### AC PASS/FAIL matrix

| AC | Status | Verification command | Observed output |
|----|--------|---------------------|-----------------|
| AC-MX-ASSOC-001 (sub-line drives assoc, no path/body match) | PASS | `go test -run TestAC001 ./internal/mx/` | `ok internal/mx` — SpecRef captured, SPEC-FIXTURE-001 in associations with neither path nor body matching |
| AC-MX-ASSOC-002 (coverage lift ≥ 10.2 % AND delta ≥ 40) | PASS | `go test -run TestAC002 -v ./internal/mx/` | coverage_on=16.9 % (142/840), coverage_off=9.3 % (78/840), associated_delta=64 |
| AC-MX-ASSOC-003 (unresolved → flag-but-keep, no panic) | PASS | `go test -run TestAC003 ./internal/mx/` | `ok internal/mx` — UnresolvedSpecRef emitted AND SPEC-DOES-NOT-EXIST-001 kept |
| AC-MX-ASSOC-004 (path/body unchanged byte-for-byte) | PASS | `go test -run TestAC004 ./internal/mx/` | `ok internal/mx` — 8 characterization fixtures + additivity assertion pass |
| AC-MX-ASSOC-005 (production validator green) | PASS | `go test -run TestAC005 ./internal/mx/` | `ok internal/mx` — no new error-prefix vocabulary in GetErrors |
| AC-MX-ASSOC-006 (dangling → DanglingSpecRef, no panic) | PASS | `go test -run TestAC006 ./internal/mx/` | `ok internal/mx` — DanglingSpecRef emitted, no spurious SpecRef |
| AC-MX-ASSOC-007 (sub-line + body de-dup) | PASS | `go test -run TestAC007 ./internal/mx/` | `ok internal/mx` — SPEC-DUP-001 exactly once |

### Coverage measurement (AC-002 numbers, recorded for audit)

- coverage_on  = **16.9 %** (142 / 840 tags associated)
- coverage_off = **9.3 %** (78 / 840 tags associated)
- associated_delta = **64** (on − off; ≥ 40 floor met with 60 % headroom)
- known-SPEC set size: 569
- Note: coverage_off 9.3 % is the TRUE off-baseline for the worktree tree (840 tags), which differs from the diagnosis-report 9.7 % / 9,858-tag main-checkout snapshot (different denominator). Per plan.md §G AP-4, the off-baseline guard is asserted self-consistently (coverage_off < 10.2 % AND < coverage_on) rather than against an exact external snapshot.

### Quality gates

- `go test ./internal/mx/` → ok (full suite green)
- `go test -race ./internal/mx/` → ok (race clean)
- `go test -cover ./internal/mx/` → coverage: 90.9 % of statements (≥ 85 % target)
- `golangci-lint run --timeout=2m ./internal/mx/` → 0 issues
- `go build ./...` → exit 0; `GOOS=windows GOARCH=amd64 go build ./...` → exit 0
- Subagent boundary: `grep -rn 'AskUserQuestion' internal/mx | grep -v _test.go | grep -v '// '` → 0 matches

### Invariants

- Additive only: existing path/body association outputs unchanged for tags without `@MX:SPEC` (AC-004 characterization).
- `ExtractSpecIDs` untouched (AP-5 honored) — validation lives at the associator layer.
- `tag.Body` not mutated (AP-1 honored) — SPEC ID lives in the dedicated `SpecRef` field.
- No 3-line proximity cutoff for `@MX:SPEC` (AP-3 honored).

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_status: audit-ready
run_complete_at: 2026-08-04
```

Run-phase independently verified by the orchestrator (8/8 verification matrix PASS): `go build ./...` exit 0; `go vet ./internal/mx/...` exit 0; AC test suite (`go test -run TestAC00 -v ./internal/mx/`) all ACs green; coverage 90.9% on `internal/mx`; `golangci-lint run --timeout=2m ./internal/mx/...` 0 issues; `GOOS=windows GOARCH=amd64 go build ./...` exit 0; full `go test ./...` green modulo the known pre-existing `internal/web` flake (signal:terminated — unrelated to this SPEC's scope).

## §E.4 Sync-phase Audit-Ready Signal

```yaml
sync_status: audit-ready
sync_complete_at: 2026-08-04
sync_commit_sha: "pending-backfill"
```

## §F Phase 4 Mode Selection

- **Input parameters**: tier M; scope ~5 files (`internal/mx`: tag.go, scanner.go, spec_association.go, resolver_query.go, + tests); domain count 1 (Go `internal/mx`); file language Go (100%); concurrency benefit LOW (coding-heavy, single domain); Agent Teams n/a (retired).
- **Mode evaluation**: Mode 1 trivial — no (multi-file additive change); Mode 2 background — no (write work); Mode 3 agent-team — RETIRED; Mode 4 parallel — no (single domain, coding-heavy → Anthropic coding-parallelism caveat); Mode 6 workflow — no (<30 files, not purely mechanical-uniform, core-model touch); Mode 5 sub-agent — **selected**.
- **Decision**: `sub-agent` (Mode 5) — single `manager-develop` (cycle_type=tdd), sequential per milestone.
- **Justification**: Coding-heavy single-domain Go work touching the core `Tag` model; per Anthropic's coding-task parallelism caveat, the sequential sub-agent path is the safe default for coding work. Additive change (new field + scanner arm + associator source) with characterization/regression safety (AC-004/005); no genuine cross-file parallelism benefit.
- **Progression mode**: autonomous — `/moai goal` (ac_converge) armed after Implementation Kickoff Approval (user-approved 2026-08-04); run-phase continues without per-turn prompts until ACs converge or a semantic failure escalates.

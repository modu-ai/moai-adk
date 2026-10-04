# SPEC-SESSION-CC-VERSION-001 — Implementation Plan

## §A Context

Producer: `internal/session` (the version reads, beside the existing `proc_info_*` platform
family). Consumers: `internal/cli` — `moai session list` (`session.go`), `moai doctor`
(`doctor.go` check registry), and the lane launcher (`cc.go` entry path +
`factory_lane_relaunch.go` guard).

Evidence base: `.moai/reports/t1348/verdict.md` (items ② and ④), re-verified in this tree at
`30ce3a02d` — every baseline in `acceptance.md` was measured here, not carried from the
verdict.

Tier **M**, justified in §B.

## §B Tier justification — M

| Signal | Measurement |
|---|---|
| Files | ~13 — `internal/session` 3 platform readers + 1 resolver/parser + 2 test files; `internal/cli` 1 doctor check + edits to `session.go` / `cc.go` / `factory_lane_relaunch.go` + 2 test files |
| Packages touched | 2 (`internal/session`, `internal/cli`) |
| LOC estimate | ~500-700, inside the M band (300-1000) |
| Cross-cutting schema change | none — the registry record is untouched; the JSON addition is flag-gated and additive |
| Doctrine / docs-site sweep | none (the `--clear-policy` docs sweep is card t1482's item ①) |
| Milestones | 5 |

Under the M ceilings (5-15 files); not L — no always-loaded doctrine procedure, no published
docs surface, no constitutional concern. Budget: 10 requirements / 10 acceptance criteria
(ceilings 16 / 16).

## §C Producer / consumer split

`internal/session` owns the reads (it already owns the process-table seam family and the
registry types); `internal/cli` owns every user-facing surface. No third package is created.

**PRESERVE list** (files this SPEC must not modify): `internal/session/registry.go` (the
`Entry` record — REQ-SCV-006 pins its field set), `internal/config/envkeys.go` (read-only use
of existing constants), `.moai/config/**`, `internal/web/**`, and every file outside the two
packages above.

## §D Milestones

Each milestone leaves the tree green and adds no half-state. Ordered by decision
reversibility: the seam shape and package placement (M1) and the guard semantics (M4) are the
decisions expensive to unwind; M5 is the mechanical tail.

**Review attention.** M1 and M4. M1 fixes where the reads live and what their seam looks like
— every later surface consumes it. M4 decides refuse-vs-strip on the relaunch path, and its
refusal text is a user-facing contract.

### M1 — The version reads behind seams (highest reversibility cost)

In `internal/session`, following the `proc_info_*` build-tag pattern:

- `ccversion_darwin.go` — running version from the pid's text mapping (`lsof -a -d txt -p`,
  bounded, parse `(versions/|claude-code/)[0-9.]+`).
- `ccversion_linux.go` — running version from `os.Readlink("/proc/<pid>/exe")` over the same
  path shapes.
- `ccversion_other.go` (`!linux && !darwin`) — unsupported → `unknown`.
- One resolver file (untagged): the pure version-segment parser plus the installed-version
  read (`exec.LookPath("claude")` → `filepath.EvalSymlinks` → parse), both behind package
  seam vars in the `procInfoFunc` style (`session_pid.go:50-57`).
- The view type: running + installed + per-value degradation, `unknown` on every failure
  (REQ-SCV-003); no error ever leaves the reads.

Ships alone: reads with tests, no surface wired.

### M2 — `moai session list --cc-version`

`session.go`: a `--cc-version` flag that enriches each entry with the two reads (JSON fields
additive under the flag; human output names both), degrading entries render `unknown` and are
never omitted (REQ-SCV-005). The no-flag path performs zero probes and keeps today's field set
byte-for-byte (REQ-SCV-006 — the pre-spawn batch contract, `session.go:45-46`).

Ships alone: the visibility the card asked for, end to end.

### M3 — The doctor staleness check

`doctor_ccversion.go` + one registry row in `doctor.go`'s MoAI group (the
`checkTodoGhostInventory` registration pattern, `doctor.go:230`). Per live registry entry: one
probe, running vs installed; behind → `CheckWarn` naming both versions; equal → `CheckOK`;
`unknown` on either side → `CheckOK` with the unknown noted (staleness cannot be judged from
unknown — warning there would nag npm-style installs). Never `CheckFail`; the check is
advisory in the `checkFlagSlot` pattern (REQ-SCV-007).

Ships alone: the leader-facing summary surface.

### M4 — The resume emergency path (guard semantics)

- One pure assembler in `internal/cli`: `func(pass-through args, name, settingsFlag) []string`
  — returns the child argv carrying name, settings, and the pass-through tokens (REQ-SCV-008).
  The one-shot lane-join path routes through it; behavior on that path is unchanged by
  construction (the pass-through already works — §A.3.1).
- Validation beside it: `--resume` without a value refuses before any launch (REQ-SCV-009).
- The guard in `factory_lane_relaunch.go`: on entry, when `claudeArgs` carries `--resume`
  under policy `relaunch`, return the refusal naming the one-shot form
  (`moai cc -l --name lane-<n> -- --resume <session-id>`) before the loop's first iteration —
  zero card sessions started, token neither propagated nor stripped (REQ-SCV-010).

Ships alone: the emergency form becomes explicit, validated, and impossible to misuse through
the relaunch loop.

### M5 — Verification batch and cross-platform build (mechanical tail)

`go build ./...`; `GOOS=windows GOARCH=amd64 go build ./...` (C.2 — the windows path must
compile and degrade to `unknown`); `go vet`; `golangci-lint run` on both packages; targeted
`go test` for the new tests; env-scrubbed invocation form (`unset … && …` in one compound
call). Evidence into `progress.md` §E.2.

## §E Self-Verification (manager-develop deliverables)

Per the 5-section evidence-bearing format, each item names command + verbatim output + tree:

- **E1** — AC PASS/FAIL matrix over `acceptance.md` §A-§C, one row per AC-SCV-001..010.
- **E2** — cross-platform build: both build commands above, exit 0 each.
- **E3** — coverage: `go test -cover ./internal/session/ ./internal/cli/` — new files ≥ 85%.
- **E4** — boundary grep: `grep -rn 'AskUserQuestion' internal/session/ internal/cli/ | grep -v _test | grep -v "// "` → 0 hits (C-HRA-008).
- **E5** — lint: `golangci-lint run ./internal/session/... ./internal/cli/...` — 0 new issues vs the pre-change baseline.
- **E6** — RED evidence (TDD): the fixture-driven tests of AC-SCV-001..003 fail before M1's readers exist (no seam to inject), captured verbatim.
- **E7** — the §F run-phase measurements, with whatever gaps this machine cannot produce (recorded, not inferred).

## §F Open run-phase verifications (measure, do not assume)

1. **The one-shot emergency form, end to end.** `moai cc -l --name <lane> -- --resume <real-session-id>` observed reaching a child argv — via the launcher's debug dump (`debugTiming.debugDump`, `cc.go:292-297`, printed pre-exec) so no interactive session is needed. Expected: name, settings, and `--resume` all present.
2. **The refused spelling, observed.** `moai cc -f lane-3 -- --resume <id>` → `factoryFlagUsageError` text observed verbatim (documents that the t1348 §"제안 절차" spelling is not the emergency form).
3. **The relaunch guard, observed.** `moai cc -l --name <lane> --clear-policy relaunch -- --resume <id>` → the new refusal text, exit non-zero, zero card sessions started.
4. **A live lsof positive control.** `lsof -a -d txt -p <a live claude pid>` on this machine parses to the version that process runs (the platform reader's real-world positive control; the Linux reader stays fixture-tested — a Linux live probe is likely infeasible on this machine and is recorded as a gap if so).

## §G Constraints (DO NOT VIOLATE)

- PRESERVE list of §C — `registry.go`, `envkeys.go`, `.moai/config/**`, `internal/web/**` untouched.
- No `--no-verify`, no `--amend`, no force-push; Conventional Commits with the `🗿 MoAI` trailer.
- B1 cross-platform: the new platform files carry build tags; `GOOS=windows GOARCH=amd64 go build ./...` green before any commit.
- B8 hygiene: no runtime-managed files (`.moai/state/`, `.moai/harness/`), stage by explicit pathspec.
- The reads never shell out to `claude --version` (REQ-SCV-002) and never write the registry (§D).
- Degradation is `unknown` + exit 0 — an unreadable process must never fail `session list` or `doctor` (REQ-SCV-003).

## §H Cross-References

- `.moai/reports/t1348/verdict.md` — the investigation; § "MoAI가 자동화할 수 있는 것" items ② and ④.
- `internal/session/session_pid.go:50-57` — the seam pattern M1 follows.
- `internal/cli/doctor.go:204-246` — the check registry M3 joins.
- `internal/cli/factory.go:92` — the `-f` refusal that decides the emergency form's spelling.
- Related SPECs: `SPEC-FACTORY-SELF-DISPATCH-001` (the relaunch policy's origin), `SPEC-V3R6-MULTI-SESSION-COORD-001` (the session registry's L1 primitive).

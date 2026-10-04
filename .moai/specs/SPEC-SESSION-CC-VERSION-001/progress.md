# SPEC-SESSION-CC-VERSION-001 — Progress

## §E.1 Plan-phase Audit-Ready Signal

- Artifacts authored (Tier M set): `spec.md`, `plan.md`, `acceptance.md`, plus this
  `progress.md`.
- Tier: M (justification in `plan.md` §B — ~13 files, 2 packages, no doctrine or docs-site
  sweep). Threshold 0.80.
- SPEC ID regex check executed as Bash, output `PASS`. ID uniqueness confirmed:
  `ls .moai/specs | grep -i "SESSION-CC-VERSION"` returns no match.
- Budget: **10 requirements, 10 acceptance criteria** (ceilings 16 / 16).
- Every absence-based criterion carries a pre-change baseline measured in this tree at
  `30ce3a02d` (branch `WT-session-cc-version`, cut from develop).
- Evidence base: `.moai/reports/t1348/verdict.md` items ② and ④, every code citation
  re-verified in this tree before use.
- `moai spec lint SPEC-SESSION-CC-VERSION-001` → `✓ No findings — all SPEC documents are
  valid`, exit 0 (after two fix rounds: AC→REQ mappings restated in the `maps REQ-…` house
  form, and `-run` selectors anchored `'^Test…$'`; the first measurement's 20 warnings
  included 2 hidden by a `tail`-truncated read — full output captured on re-run).
- **Plan-audit iter1: FAIL 0.81** (Tier M threshold 0.80 met numerically; FAIL carried by
  blocking defects D1-D4; report `.moai/reports/t1465/plan-audit-iter1.md`). All six deltas
  repaired in place (no requirement rewrites):
  - **D1 (critical)** — the documented emergency form `moai cc -l --name lane-<n>` is itself
    refused (`laneFlagNameError`, `factory.go:391-394`; `operatorSuppliedName`,
    `factory_launch_helpers.go:390-404`). Corrected to the bare lane join
    `moai cc -l -- --resume <session-id>` in spec §A.3, plan M4 + §F.1 + §F.3, AC-SCV-010;
    the refusal re-observed by running this tree's build from /tmp with the env stamps
    scrubbed → `ERROR: -L/--Lane already names the role; drop the --name/-n flag.`, exit 1.
  - **D2 (major)** — the iter-0 lsof baseline did not reproduce (real: 2 comment hits under
    `internal/session/`; 4 exec sites repo-wide, all cwd/port). Restated in spec §A.1 and
    AC-SCV-001 from re-measured greps.
  - **D3 (minor)** — the exit-0 assertion added to AC-SCV-005, where AC-SCV-003 delegates it.
  - **D4 (major)** — AC-SCV-004's selector widened to the three real test names as an anchored
    alternation, with a swept-count-3 + no-`[no tests to run]` requirement.
  - **D5 (adopted)** — both `--resume` spellings (`--resume <v>`, `--resume=<v>`) pinned in
    REQ-SCV-009/010, AC-SCV-009/010, plan M4.
  - **D6 (adopted)** — the lsof txt parse anchored to the claude-binary line in spec §A.2 and
    plan M1.
- Version 0.1.0 → **0.1.1** (HISTORY row added). The REQ layer survives intact per the
  auditor; `updated:` fields refreshed (same-day).
- Status: `draft`. Plan-audit iter1 repaired; **iter2 pending**.

### Card stage plan (lane-23, card t1465)

This plan phase is the card's first stage. The lane task list carries the remaining stages in
order: plan-audit (independent audit) → plan→run kickoff decision record → run (manager-develop,
M1-M5 of `plan.md` §D) → verification batch (env-scrubbed build + lint + targeted tests) →
sync (manager-docs) → card-review + merge-ready report to the leader. The lane orchestrator
commits these artifacts; nothing is committed or pushed by the plan phase.

### Assumptions made at plan phase

1. **Flag name `--cc-version`** on `moai session list` — a naming choice. The wire contract
   that matters (additive JSON fields under the flag, default path byte-identical) is pinned
   by REQ-SCV-005/006; renaming the flag before run is cheap.
2. **The doctor check is advisory (WARN-only, never `CheckFail`)** — in the `checkFlagSlot`
   pattern. An operator wanting staleness to gate `moai doctor`'s exit status would be a
   separate decision, not this SPEC's.
3. **`unknown` never warns** in the doctor check: staleness cannot be judged from unknown, and
   warning on it would nag npm-style installs with unversioned binary paths.
4. **The relaunch guard refuses the whole loop** rather than stripping the token or honoring it
   once: under `relaunch` a resume token cannot mean what it says (the loop leases new cards),
   so refusing with the safe form in the error text is the honest behavior.
5. **The pure assembler (REQ-SCV-008) changes no one-shot behavior** — the pass-through
   already survives the launcher mechanically (`spec.md` §A.3.1); the assembler names and
   tests what exists, and M4's only behavior change is the relaunch guard.
6. **Versions are never persisted** into the registry file — the live read at query time is
   the whole truth (`spec.md` §D, last exclusion).
7. **Installed-version read is path-parse only** — no `claude --version` exec (REQ-SCV-002);
   installs whose resolved path carries no version segment degrade to `unknown`.
8. **The exact spelling `moai cc -f lane-<n> -- --resume <id>` from the dispatch is refused
   today** (`factoryFlagUsageError`, `factory.go:92`); the SPEC treats the lane-join form as
   the emergency path and has the run-phase observe the refusal verbatim (`plan.md` §F.2).

## §E.2 Run-phase Evidence

### Pre-flight baselines (measured before M1, tree `3dc8c9760`)

| # | Command | Observed result |
|---|---|---|
| 1 | `git branch --show-current` + `git rev-parse --short HEAD` | `WT-session-cc-version` @ `3dc8c9760`, clean tree |
| 2 | `go build ./...` | exit 0 |
| 3 | `GOOS=windows GOARCH=amd64 go build ./...` | exit 0 |
| 4 | `golangci-lint run --timeout=2m ./internal/session/... ./internal/cli/...` | `0 issues.` (installed build: v2.1.6 — the CI-pinned version) |
| 5 | `grep -rn "Retired\|TestHarnessRetirement\|superseded" internal/session/ internal/cli/` | hits only in unrelated retirement records (`cg.go` retired launcher, `harness_route.go` supersedence note) — no policy conflict with this SPEC's surface |

### M1 — The version reads behind seams

**E6 RED evidence (TDD)** — the three fixture tests of AC-SCV-001..003, run against
signature-only stubs (no reader logic; bodies returned zero values), before any reader
existed. Command: `go test ./internal/session/ -run '^(TestRunningVersionFromInjectedMapping|TestInstalledVersionFromResolvedPath|TestVersionDegradationRendersUnknown)$' -v`
→ exit **1**; verbatim (abridged to the assertion lines — full log retained):

```
=== RUN   TestRunningVersionFromInjectedMapping
    ccversion_test.go:49: ResolveCCVersions(4242).Running = "", want 2.1.281 (a library mapping's version shape must not satisfy the read)
--- FAIL: TestRunningVersionFromInjectedMapping (0.00s)
=== RUN   TestInstalledVersionFromResolvedPath
=== RUN   TestInstalledVersionFromResolvedPath/versions_shape
    ccversion_test.go:66: installed version = "", want 2.1.288
=== RUN   TestInstalledVersionFromResolvedPath/claude-code_shape
    ccversion_test.go:72: installed version = "", want 2.1.284
=== NAME  TestInstalledVersionFromResolvedPath
    ccversion_test.go:87: versionSegmentFromPath("/Users/dev/.local/share/claude/versions/2.1.281/claude") = "", want "2.1.281"
    ccversion_test.go:87: versionSegmentFromPath("/opt/node/lib/node_modules/@anthropic-ai/claude-code/2.1.284/cli") = "", want "2.1.284"
--- FAIL: TestInstalledVersionFromResolvedPath (0.01s)
=== RUN   TestVersionDegradationRendersUnknown
    ccversion_test.go:133: dead pid running = "", want "unknown"
    ccversion_test.go:141: probe error running = "", want "unknown"
    ccversion_test.go:152: unversioned running = "", want "unknown"
    ccversion_test.go:168: unsupported platform running = "", want "unknown"
--- FAIL: TestVersionDegradationRendersUnknown (0.00s)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/session	0.317s
```

**GREEN** — the same selector after the readers landed → exit 0; verbatim tail:

```
--- PASS: TestRunningVersionFromInjectedMapping (0.00s)
--- PASS: TestInstalledVersionFromResolvedPath (0.01s)
--- PASS: TestVersionDegradationRendersUnknown (0.00s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/session	0.355s
```

**M1 milestone gate** — `go build ./...` → exit 0; `GOOS=windows GOARCH=amd64 go build ./...`
→ exit 0; `go test ./internal/session/` → `ok … 17.618s` exit 0 (whole package, no selector);
`golangci-lint run --timeout=2m ./internal/session/... ./internal/cli/...` → `0 issues.`.

**Gaps (M1)**: none — every AC-SCV-001..003 command was run in this phase on this tree.
**Residual-risk (M1)**: the `ccversion_other.go` runtime path cannot execute on this darwin
machine; its contract is carried by the `unsupported platform read` subtest standing in
through the shared seam, plus the windows cross-build proving the file compiles. The darwin
`lsof` exec site itself is seam-excluded from unit coverage per REQ-SCV-004 (no test spawns a
process); its real-world behavior is measured live in §F.4 below.

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending run-phase>_

## §F Phase 4 Mode Selection

Input parameters: tier M; scope ~13 files across 2 packages (`internal/session`, `internal/cli`); domain count 2 (Go source only); file language mix 100% Go; concurrency benefit LOW (coding-heavy implementation, coupled milestone ordering M1→M5); Agent Teams prereqs not requested (no operator `--team`).

| Mode | Selected | Rationale |
|---|---|---|
| `direct` | not selected | Semantic multi-file feature work, not a typo/single-line fix |
| `serial` | **selected** | Coding-heavy work in 2 coupled packages; single writer in the card worktree; per-Anthropic coding-task parallelism caveat |
| `fanout` | not selected | No independent multi-domain research split; coding tasks favor sequential |
| `sweep` | not selected | ~13 files, semantic new-code work — fails the ≥~30-file mechanical-uniform test |
| `agent-team` | not selected | Explicit-request-only experimental surface; no request |

Decision: `serial`

Justification: the milestones are ordered by decision reversibility (plan §D — M1 seam shape, M4 guard semantics) and each consumes the previous one's types, so parallel spawns would only create integration risk inside one worktree. `serial` keeps one writer per tree and matches the coding-heavy caveat.

Kickoff gate: met in autonomous form — plan-audit iter2 PASS 1.0 (threshold 0.80), artifact hash `49065ad2…` unchanged since the verdict (re-measured this run); decision record: `.moai/reports/t1465/kickoff-decision.md`.

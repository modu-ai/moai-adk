# SPEC-CLIFIX-CONTRACT-001 — Implementation Plan

## §A Context

P1 row of the CLI audit roadmap: contract drift. The `ExitCoder` infrastructure already exists at the main.go boundary — this SPEC is adoption work, not new design (audit §4: "이미 인프라 존재, 채택만 필요").

## §B Known Issues (findings inventory)

| # | File anchor (re-verify before edit) | Defect | Fix direction |
|---|---|---|---|
| 1 | hook.go, hook_pre_push.go, astgrep.go, spec_lint.go, spec_drift.go, migrate_agency.go:590, harness/execute.go:327, agentlint/workflow_lint.go:159 | os.Exit inside RunE/PostRunE (defers skipped, untestable) | return ExitCoder error; map at main.go |
| 2 | github.go:97,103 | --dry-run registered, never read → link-spec writes registry anyway | wire flag into all mutating subcommands |
| 3 | spec_status.go:205-235 | --confirm dead flag; fmt.Scanln hangs non-TTY; git runs in cwd not specs root | gate on --confirm/--yes; `git -C projectRoot` |
| 4 | astgrep.go:107-122 | HasErrors→exit 1 only in text format; json/sarif exit 0 | evaluate HasErrors after format branch |
| 5 | constitution.go:296-319 | exitCodeError{2} never interpreted | map at Execute boundary |
| 6 | spec_lint.go / spec_audit | documented exit 3 (invalid args) / exit 2 (MUST-FIX) not produced | implement via ExitCoder |
| 7 | hook_pre_push.go:180-198 | violations on stdout + exit 2 → Claude Code surfaces stderr only | route detail to stderr |
| 8 | team_spawn_lock_test_unix.go | filename not `_test.go`-suffixed → test never runs; testing pkg in prod binary | rename to team_spawn_lock_unix_test.go |

## §C Pre-flight

1. Read main.go ExitCoder mechanism; enumerate its current adopters to copy the idiom.
2. Confirm SPEC-CLIFIX-CRITICAL-001 landed (hook.go/migrate_agency.go/team_spawn.go bases moved).
3. Inventory help-text exit-code declarations for the changed commands (source for REQ-CONT-001-008 contract tests).
4. Check spec_lint.go note: current RunE itself calls os.Exit(1)/os.Exit(2) — the lint command is in scope of site removal.

## §D Constraints

- Exit codes observed by external callers (CI scripts, git hooks) MUST remain numerically identical where already correct; this SPEC only adds missing codes and reroutes streams.
- No behavior change to what is detected — only verdict communication.
- The renamed lock test must be verified to actually FAIL if the flock logic regresses (guard against a vacuously-green rename).

## §E Self-Verification

- E1: AC matrix PASS/FAIL against acceptance.md.
- E2: `go build ./...` + `go test ./internal/cli/... ./internal/cli/harness/... ./internal/cli/agentlint/... -count=1` verbatim.
- E4: `grep -rn 'os\.Exit' internal/cli --include='*.go' | grep -v _test.go | grep -v main` audit output cited.
- E5: `golangci-lint run` no new findings.

## §F Milestones (priority order)

- M1 — ExitCoder adoption: replace the 8 os.Exit sites with ExitCoder returns; extend main.go mapping if a code is unmapped. Repro-first: a test per command asserting defer execution + exit code via command runner.
- M2 — Exit-code contracts: astgrep format-independent HasErrors; constitution exit 2; spec_lint exit 3; spec_audit exit 2.
- M3 — Flags + streams: github --dry-run wiring; spec_status --confirm/--yes + non-TTY abort + `git -C`; pre-push stderr routing.
- M4 — Test enablement + closure: rename team_spawn_lock_unix_test.go, prove the flock test runs and can fail; contract test suite (REQ-CONT-001-008); §E self-verification.

## §G Anti-Patterns and Risks

- Execution order: P0→P1→P2→P3→P4. Shared-file overlap: hook.go/hook_pre_push.go (with CRITICAL-001 h; with HYGIENE-001 timeout constants), migrate_agency.go (CRITICAL-001 f/g), team_spawn* (CRITICAL-001 b, LINTER-STALE-001 claim validation). This SPEC starts only after CRITICAL-001 is merged.
- Anti-pattern: swallowing errors to avoid os.Exit — the ExitCoder return must carry the original diagnostic.
- Risk: spec_lint exit-code change could affect CI pipelines that treat exit 1 as the only failure signal — document the 0/1/2/3 contract in the command Long text (already declared; implementation aligns to it).
- Risk: pre-push stderr rerouting may change golden outputs in existing tests — update goldens deliberately, never loosen assertions.

## §H Cross-References

- Findings SSOT: `.moai/reports/cli-improvement-audit-20260710.html` §3 clusters 3-5, §4 rows 3-4, §5 P1.
- Depends on: SPEC-CLIFIX-CRITICAL-001. Followed by: SPEC-CLIFIX-CONCURRENCY-001.
- main.go ExitCoder mechanism (existing infrastructure, adoption target).

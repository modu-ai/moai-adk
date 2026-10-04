# SPEC-SESSION-CC-VERSION-001 — Acceptance Criteria

Ten criteria against ten requirements. Every criterion names a command and an expected result;
none is satisfied by reading a file and forming a judgement.

Where a criterion is satisfied by an **absence**, the pre-change baseline is stated, measured
in this tree at `30ce3a02d`, so a zero-hit result is new information rather than a vacuous
pass. A criterion that already passes on the untouched tree is marked as a preservation half
and is never the load-bearing half.

## §A The version reads

**AC-SCV-001** (maps REQ-SCV-001) — Given a fixture pid→mapping table in which pid 4242 maps
to a text mapping carrying `…/claude/versions/2.1.281`, When the session view resolves the
entry, Then it reports `2.1.281`. Command: `go test ./internal/session/ -run '^TestRunningVersionFromInjectedMapping$' -v`
→ `--- PASS`. Baseline (RED-now): `grep -rn "lsof" internal/session/ --include="*.go"` → 0
hits and `internal/session/registry.go:120-129` declares no version field, so no such reader
exists to pass this before M1.

**AC-SCV-002** (maps REQ-SCV-002) — Given a fixture PATH directory whose `claude` entry is a
symlink resolving to `…/versions/2.1.288`, When the installed-version read runs, Then it
reports `2.1.288`; and Given a symlink resolving to `…/claude-code/2.1.284`, Then `2.1.284`.
Table-driven, both path shapes. Command: `go test ./internal/session/ -run '^TestInstalledVersionFromResolvedPath$' -v`
→ `--- PASS`. Baseline: the only installed-version reader in the tree is doctor's
`checkClaudeCode` (`doctor.go:449`), which execs `claude --version` — a different mechanism;
no path-parse reader exists. The version-segment parser is a pure string function and is
asserted directly for both shapes.

**AC-SCV-003** (maps REQ-SCV-003) — Given each of the four degradations — a dead pid (fixture
liveness says not-alive), an injected probe error, a resolved path with no version segment,
and the unsupported-platform build (`ccversion_other.go`) — When the view resolves an entry
under each, Then the value renders `unknown` in every case. Command: `go test ./internal/session/ -run '^TestVersionDegradationRendersUnknown$' -v`
→ `--- PASS`, one subtest per case. The exit-0 half is covered at the surface (AC-SCV-005/006
run the command end to end with degrading fixtures and assert exit 0).

**AC-SCV-004** (maps REQ-SCV-004) — Given the merged tree, When the tests of AC-SCV-001..003
run, Then every one of them supplies its pid→path mappings and PATH resolution through the
package seam, and no test file under `internal/session` shells out. Command:
`go test ./internal/session/ -run '^TestRunningVersion$' -count=1 -v` → `--- PASS`; and
`grep -n "exec.Command" internal/session/ccversion*_test.go` → 0 hits. The seam is a package
var substituted per test, in the `procInfoFunc` pattern (`session_pid.go:50-57`); the fixture
lives in the test, never the process table.

## §B The surfaces

**AC-SCV-005** (maps REQ-SCV-005) — Given a fixture registry holding one entry with injected
reads (running `2.1.281`, installed `2.1.288`), When `moai session list --cc-version --json`
runs, Then that entry's JSON carries both new fields with those values; and When the same
fixture runs without `--json`, Then the human output names both versions for the entry.
Command: `go test ./internal/cli/ -run '^TestSessionListCCVersion$' -v` → `--- PASS`. A
degrading fixture (dead pid) renders `unknown` in both outputs and the entry is still listed.

**AC-SCV-006** (maps REQ-SCV-006) — Given the same fixture and a probe-counting seam, When
`moai session list --json` runs **without** `--cc-version`, Then the probe count is 0, the
command exits 0, and each entry's JSON key set is exactly today's eight
(`session_id`, `spec_id`, `phase`, `started_at`, `last_heartbeat`, `pid`, `host`, `cwd` —
measured at `registry.go:120-129`). Command: `go test ./internal/cli/ -run '^TestSessionListDefaultNoProbe$' -v`
→ `--- PASS`. The zero-probe half passes today by absence (nothing probes); the load-bearing
half is that adding the flag does not change the default path — it is a preservation guard on
the pre-spawn batch contract, and it fails the moment enrichment leaks into the default path.

**AC-SCV-007** (maps REQ-SCV-007) — Given a fixture registry with a live session running
`2.1.281` against installed `2.1.288`, When the doctor staleness check runs, Then its status
is `CheckWarn` and its message names both versions; Given equal versions, Then `CheckOK`;
Given `unknown` on either side, Then `CheckOK` with the unknown noted (never a warn on
unknown). Command: `go test ./internal/cli/ -run '^TestDoctorCCVersionStaleness$' -v` →
`--- PASS`. The check never returns `CheckFail` in any fixture — asserted across all subtests,
so the advisory property (doctor's exit status unchanged) is pinned at the unit layer.

## §C The resume emergency path

**AC-SCV-008** (maps REQ-SCV-008) — Given pass-through args `["--name", "lane-3", "--",
"--resume", "<session-id>"]` and a settings flag `["--settings", "<file>"]`, When the pure
assembler runs, Then the returned argv carries the name pair, the settings pair, and the
resume pair, in a deterministic order; and the function is side-effect-free by construction —
it takes no environment, filesystem, or process parameters and returns `[]string`. Command:
`go test ./internal/cli/ -run '^TestLaneJoinChildArgv$' -v` → `--- PASS`. Baseline (RED-now):
`grep -rn '"--resume"' internal/ --include="*.go" | grep -v _test` → 0 hits — the pass-through
works mechanically today (§A.3.1 of `spec.md`) but exists as no named, tested assembly.

**AC-SCV-009** (maps REQ-SCV-009) — Given args carrying `--resume` with no following value,
When the launcher-side validation runs, Then it refuses with an error naming
`--resume <session-id>`, and the launch seam is never called (asserted via the stubbed launch
function's call count of 0). Command: `go test ./internal/cli/ -run '^TestResumeRequiresValue$' -v`
→ `--- PASS`.

**AC-SCV-010** (maps REQ-SCV-010) — Given the clear policy `relaunch` and child arguments
carrying `--resume <session-id>`, When the supervising loop is started (lease and launch
seams stubbed), Then it returns the refusal error, the error text names the one-shot
lane-join form (`moai cc -l --name lane-<n> -- --resume <session-id>`), and the launch seam
call count is 0 — no card session is started, and the token is neither propagated nor
stripped. Command: `go test ./internal/cli/ -run '^TestRelaunchRefusesResumeToken$' -v` →
`--- PASS`. Baseline (RED-now): `grep -n "resume" internal/cli/factory_lane_relaunch.go` →
0 hits, and the loop hands `claudeArgs` to every iteration's session
(`factory_lane_relaunch.go:115`) — today the token leaks to every card.

## §D Quality gates and Definition of Done

- All ten criteria PASS in the run-phase matrix (`progress.md` §E.2), each with command +
  verbatim output + tree SHA.
- `go build ./...` and `GOOS=windows GOARCH=amd64 go build ./...` exit 0.
- `golangci-lint run ./internal/session/... ./internal/cli/...` — 0 new issues vs the
  pre-change baseline.
- New files in `internal/session` and `internal/cli` ≥ 85% covered (`go test -cover`).
- The §F run-phase measurements of `plan.md` are either measured or recorded as explicit gaps.
- PRESERVE list untouched: `git diff --name-only` over the run shows no file outside
  `internal/session/` and `internal/cli/` (excluding `.moai/specs/`, `.moai/reports/`).

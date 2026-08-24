# SPEC-HOOK-WIRING-DRIFT-001 — Progress

## §E.1 Plan-phase Audit-Ready Signal

```yaml
plan_status: audit-ready
plan_version: 0.3.0
plan_complete_at: 2026-08-24
tier: M
artifacts: [spec.md, plan.md, acceptance.md, progress.md]
requirements: 14
acceptance_criteria: 16
milestones: [M1, M2, M3, M4]
authored_at_head: 950cb4399
amended_at_head: 6331d505c
audit_iterations:
  - iteration: 1
    verdict: FAIL
    score: 0.807
    threshold: 0.85
    must_pass: 7/7
    mutants_constructed: 4
    findings_blocking: 7
    disposition: all 7 blocking fixed, 4 optional fixed, 0 declined
    report: .moai/reports/t216/plan-audit.md
  - iteration: 2
    verdict: PASS
    score: 0.862
    threshold: 0.80
    threshold_note: >-
      iteration 1 scored against 0.85 (Tier L); the Tier M SSOT value is 0.80,
      so v0.1.0's 0.807 was already a PASS. Recorded by the auditor against its
      own prior instance.
    must_pass: 7/7
    mutants_constructed: 3
    mutants_executed: 2
    findings_blocking: 3
    disposition: all 3 blocking fixed, 4 minor/optional fixed, 0 declined
    report: .moai/reports/t216/plan-audit-iter2.md
    terminal: true
authored_in_worktree: .claude/worktrees/t216
branch: WT-hook-wiring-drift
card: t216
source_of_record: .moai/reports/t216/{d1-chain-event,d2-unwired-scripts,d3-mx-cold-start}.md
deferred: [t242, t243, t244, file_changed.go-die-at-exit-twin, deploy-time-template-snapshot]
```

## §E.2 Run-phase Evidence

### M2 — `moai doctor` hook-wiring drift diagnostic

> Continuity note: the delegated implementation run was killed mid-work by a
> spend limit, at the point where it was about to discharge condition 2. The
> implementation and its nine tests were already on disk; the orchestrator
> resumed from there rather than restarting, and performed conditions 2-4, the
> golden regeneration, and this section itself. Nothing was re-run that had
> already landed, and nothing already landed was taken on trust — the tests were
> re-executed here before anything was built on them.

#### Claim

`moai doctor` carries a **Hook Wiring** check that renders the settings template
in memory, diffs hook entries against the project's `.claude/settings.json` in
both directions, and reports the drift **without writing anything**. On this
repository it independently reproduces the drift D-1 measured by hand.

#### Evidence

**1. The check exists, is registered, and is green (9 tests).**

```
$ go test -count=1 -run 'HookWiring' -v ./internal/cli/
--- PASS: TestCheckHookWiringDriftReportsTemplateOnly (0.00s)
--- PASS: TestCheckHookWiringDriftReportsProjectOnly (0.00s)
--- PASS: TestCheckHookWiringDriftCleanFixtureReportsNoDrift (0.00s)
--- PASS: TestCheckHookWiringDriftRendersInjectedTemplate (0.00s)
--- PASS: TestCheckHookWiringDriftMissingSettingsWarns (0.00s)
--- PASS: TestCheckHookWiringDriftCorruptSettingsWarns (0.00s)
--- PASS: TestCheckHookWiringDriftRenderFailureWarns (0.00s)
--- PASS: TestCheckHookWiringDriftWritesNothing (0.00s)
--- PASS: TestHookWiringCheckIsRegistered (5.68s)
ok  	github.com/modu-ai/moai-adk/internal/cli	6.544s

$ go test -count=1 -run 'TestHookEntries|TestParseHookEntries' ./internal/template/
ok  	github.com/modu-ai/moai-adk/internal/template	0.466s
```

**2. End-to-end: the diagnostic reproduces D-1's hand measurement.** This is the
load-bearing observation — the check was written from the reports, and here it
finds the same two drifts independently, through the shipped code path:

```
$ ./bin/moai doctor
│  warn  Hook Wiring  hook wiring drift — template-only (in template, not
│                     registered in project): chain-event.sh,
│                     status-transition-ownership.sh x3; project-only
│                     (registered in project, not in template):
│                     status-transition-ownership.sh
```

`chain-event.sh` (d1 §E1 drift a) and `status-transition-ownership.sh` 3-in-template
vs 1-in-project (d1 §E1 drift b) — both directions, both reported.

**3. AC-HWD-008 — the diagnostic fails open.**

```
$ ./bin/moai doctor >/dev/null 2>&1; echo "rc=$?"
rc=0
```

A drift is a `warn`, never a `fail`, and never a non-zero exit.

**4. [HARD] Condition 2 — the AC-HWD-007 exception is one file, not a directory.**
The named exception is `.moai/state/config-cache.json`, excluded because
`moai doctor` writes it on first run in any project (audit finding N1: the
earlier, wider clause was unsatisfiable no matter how the check was implemented).
The proof that the exclusion did not silently widen to the whole directory is a
mutant that writes a **different** file under the same directory:

```go
// TEMPORARY MUTANT, inserted at the head of checkHookWiringDrift:
_ = os.MkdirAll(filepath.Join(projectRoot, ".moai", "state"), 0o755)
_ = os.WriteFile(filepath.Join(projectRoot, ".moai", "state", "hook-wiring-cache.json"), []byte("{}"), 0o644)
```

```
$ go test -count=1 -run 'TestCheckHookWiringDriftWritesNothing' ./internal/cli/
--- FAIL: TestCheckHookWiringDriftWritesNothing (0.00s)
    doctor_hook_wiring_test.go:234: project tree changed:
        before:
        .claude/settings.json|23678|2026-08-24 17:52:56.355893008 +0000 UTC
        after:
        .claude/settings.json|23678|2026-08-24 17:52:56.355893008 +0000 UTC
        .moai/state/hook-wiring-cache.json|2|2026-08-24 17:52:56.356114596 +0000 UTC
FAIL	github.com/modu-ai/moai-adk/internal/cli	0.900s
```

Mutant reverted; the criterion is green again (`ok … 6.130s`). **The exception is
narrow: any other file under `.moai/state/` still fails the criterion.**

Note the unit-level test is in fact *stricter* than AC-HWD-007 requires. It calls
`checkHookWiringDrift` directly rather than through `bin/moai doctor`, so doctor's
own cache write never occurs, and `treeSnapshot`
(`doctor_hook_wiring_test.go:270-296`) therefore carries **no exclusion at all** —
it snapshots every file under the fixture root. The named exception in the
criterion covers the whole-`doctor`-run path; the unit test needs none.

**5. [HARD] Condition 3 (t241) — every gate was observed failing.** The nine tests
are not nine passing assertions: each constructs an input that must fail and
observes the failure. `ReportsTemplateOnly` and `ReportsProjectOnly` build
fixtures **from the render** (per acceptance.md's [HARD] fixture note — never by
copying the project) with a known entry removed or added, and assert the
diagnostic names it. `WritesNothing` was additionally shown to fail under mutation
(item 4). `MissingSettingsWarns`, `CorruptSettingsWarns`, and
`RenderFailureWarns` each drive a distinct failure path.

**6. [HARD] Condition 4 — the injectable template source survived.** The check's
signature is `checkHookWiringDrift(projectRoot string, tmplFS fs.FS, verbose bool)`
— the template source is a **parameter**, not an internally-read embedded FS
(`doctor_hook_wiring.go:40-45`). Production passes
`hookWiringTemplateSource()`; `TestCheckHookWiringDriftRendersInjectedTemplate`
injects a fixture template through the same seam and asserts the diagnostic's own
returned message moves with it. This is what closes audit iteration 1's mutant
M-2 (a hardcoded expected-entry list that passed the output-only criteria while
never rendering the template).

**7. Static verification.**

```
$ go build ./...                       → exit 0 (no output)
$ go vet ./...                         → exit 0 (no output)
$ GOOS=windows go vet ./...            → exit 0 (no output)
$ gofmt -l <touched files>             → no output
$ golangci-lint run --timeout=5m ./internal/cli/... ./internal/template/...
0 issues.
```

**8. Doctor golden snapshots regenerated.** Adding a check changes doctor's
rendered output, so three golden tests failed by construction:

```
$ grep -E '^--- FAIL' /tmp/t216-cli.log
--- FAIL: TestDoctorGolden_Light
--- FAIL: TestDoctorGolden_Dark
--- FAIL: TestDoctorGolden_NoColor
```

Regenerated with `UPDATE_GOLDEN=1`; the delta is exactly the new row and the
recomputed tallies — no other line moved:

```
$ git diff internal/cli/testdata/doctor-nocolor.golden
+│    warn    Hook Wiring       not checked: .claude/settings.json not found
-│    2 ok, 8 warn, 0 fail
+│    2 ok, 9 warn, 0 fail
-│  [Pass 14]  [Warn 11]  [Fail 0]
+│  [Pass 14]  [Warn 12]  [Fail 0]
```

(The golden fixture has no `.claude/settings.json`, so the check reports
`not checked` there — the fail-open path, exercised incidentally.)

**9. The whole `internal/cli` package is green at the M2 tree.**

```
$ go test -count=1 -timeout 1200s ./internal/cli/
ok  	github.com/modu-ai/moai-adk/internal/cli	1057.604s

$ grep -cE '^--- FAIL' <log>
0
```

Two notes on this measurement, both of which would have made it wrong if left
unstated:

- **1057 s exceeds the 546-866 s band the dispatch cited**, which is why the
  `-timeout 1200s` floor is not optional here — the 600 s figure that
  `CLAUDE.local.md` §6 warns about would have failed a passing tree, and even
  866 s would have. The excess is machine load, not a slower suite: the run
  overlapped a `golangci-lint` pass and a second worktree's activity.
- **The command reported `exit 1` while the package passed.** The wrapper ended
  in `grep -cE '^--- FAIL'`, which exits 1 when it matches nothing, and that
  became the compound command's status. The suite verdict is the `ok …` line and
  the zero count, not the wrapper's exit code — the same shape as the recorded
  hazard where a pipeline's status is read from the wrong end.

#### Baseline-attribution

Measured in this tree (`.claude/worktrees/t216`), branch `WT-hook-wiring-drift`,
against the working tree at base commit `bf9aef6f3` plus the uncommitted M2
change set. `bin/moai` was rebuilt from this tree (`make build`, stamped
`Commit=bf9aef6f3`) — the end-to-end doctor observations in items 2-3 are from
that binary, not from the installed `~/go/bin/moai`.

**Template-First does not apply to M2.** The change set is Go source and test
data only (`internal/cli/`, `internal/template/`); it touches no file under
`.claude/`, `.moai/`, or `.agency/`, so there is no template mirror to keep in
step. M3 is the milestone that carries a template-managed surface.

#### Gaps — what was explicitly NOT observed

1. **`golangci-lint` was scoped to the two touched packages**, not the whole
   repository. A pre-existing issue elsewhere would not appear here.
2. **The mutant in item 4 was reverted, not committed.** The evidence is the
   captured failure output above; the tree carries no trace of it, by design.
3. **Windows runtime is unobserved** — `GOOS=windows go vet` proves compilation
   only.
4. **The end-to-end doctor run was made against this repository**, which is a
   drifting project. The drift-free end-to-end path is covered by
   `CleanFixtureReportsNoDrift` at the unit level, not by a second `bin/moai`
   run against a warmed clean fixture.
5. **M1, M3, M4 are untouched.** This section covers M2 only; the remaining
   milestones carry their own evidence.

#### Residual risk

- **The check renders the template with this platform's context.**
  `hookWiringTemplateSource()` + `NewTemplateContext(WithPlatform(runtime.GOOS))`
  means a project on another OS renders a different expected set. That is correct
  behaviour — the entries are platform-conditional — but it means the drift
  report is not portable between machines, and two developers on different
  platforms can legitimately see different drift for the same repository.
- **A drift the diagnostic reports is not thereby fixed.** M2 deliberately only
  reports; the local drift is M1's subject, and the update-path defect that lets
  the drift arise unnoticed (d1 §E4) stays open with the deploy-time snapshot
  recorded as the structurally correct fix in §G.
- **The check is one more thing doctor runs on every invocation.** It renders a
  template and parses two JSON documents; measured within `TestHookWiringCheckIsRegistered`'s
  5.68 s whole-check-list run, which is dominated by the pre-existing checks, but
  it is not free.

### M4 — MX index build moves to `moai mx query`; the hook-side scan is deleted

> Continuity note: the delegated run was stopped at the shutdown boundary, at the
> point where it had finished implementation and tests and was about to run the
> verification suites. The orchestrator ran the verification, performed the
> end-to-end cold-path observation, and wrote this section. Implementation and
> test authorship are the delegate's; every measurement below is the
> orchestrator's own, re-run against the tree as it stands.

#### Claim

`moai mx query` builds the sidecar index on demand instead of failing with
`SidecarUnavailable`, and the hook-side cold-start scan — which in 153 measured
worktrees had produced its artifact zero times — is deleted along with the
comment that claimed a fire-and-forget goroutine survives process exit.

#### Evidence

**1. The cold-start scan and its plumbing are gone.**

```
$ grep -c 'runMXColdStartScan\|mxScanNeeded' internal/hook/session_start.go
0
```

**2. The false comment is gone.** It claimed the deferred goroutine "continues to
completion in the background (durable side effects still land)" — true for a
long-lived process, false for a CLI that exits on return, and contradicted by the
accurate comment at the scan's own definition:

```
$ grep -n 'durable side effects still land' internal/hook/session_start.go
$ echo rc=$?
rc=1
```

**3. AC-HWD-012 end-to-end, on a genuinely cold tree, with a tree-built binary.**
The recorded gap on this criterion was that its baseline came from the installed
v3.1.2 binary rather than one built here. Re-measured: `make build` (stamped
`Commit=1e75032b3`), index moved aside so the cold path is real, then

```
$ ls .moai/state/mx-index.json
ls: .moai/state/mx-index.json: No such file or directory

$ ./bin/moai mx query --kind DEBT
[
 {
  "kind": "DEBT",
  "file": ".../internal/cli/mcp_server.go",
  "line": 649,
  "body": "single-process in-memory cache only",
  ...
$ echo rc=$?
rc=0

$ ls -la .moai/state/mx-index.json
-rw-r--r--@ 1 goos staff 378793 Aug 25 04:14 .moai/state/mx-index.json
```

The query answered **and** left the index behind. Before M4 this exact invocation
returned `SidecarUnavailable: sidecar index does not exist — run 'moai mx scan'`
and wrote nothing.

**4. Package suites green.**

```
$ go test -count=1 ./internal/hook/
ok  	github.com/modu-ai/moai-adk/internal/hook	233.864s

$ go test -count=1 -run 'TestMXIndex|TestBuildMXIndex|TestMXQuery' ./internal/cli/
ok  	github.com/modu-ai/moai-adk/internal/cli	3.142s

$ go test -count=1 -run 'MX|Mx' ./internal/hook/
ok  	github.com/modu-ai/moai-adk/internal/hook	5.253s
```

**5. Static verification.**

```
$ go build ./...                    → exit 0 (no output)
$ go vet ./...                      → exit 0 (no output)
$ GOOS=windows go vet ./...         → exit 0 (no output)
$ gofmt -l <the five touched files> → no output
$ golangci-lint run --timeout=5m ./internal/cli/... ./internal/hook/...
0 issues.
```

#### Baseline-attribution

Measured in this tree (`.claude/worktrees/t216`), branch `WT-hook-wiring-drift`,
at base commit `1e75032b3` plus the uncommitted M4 change set. `bin/moai` rebuilt
from this tree for item 3.

**Template-First does not apply to M4**, as it did not to M2: the change set is
Go source and tests only (`internal/cli/`, `internal/hook/`), touching nothing
under `.claude/`, `.moai/`, or `.agency/`. M3 is the milestone that carries a
template-managed surface.

#### Gaps — what was explicitly NOT observed

1. **The full `internal/cli` suite was NOT re-run after M4.** M2's run
   (`ok … 1057.604s`, zero FAIL) predates the M4 change set; only the targeted
   MX tests were run here. CI owns the full-suite verdict against the PR head,
   and the branch is unpushed as of that measurement.
2. **The delegate's pre-implementation mutant record and its RED-before-inversion
   capture were not returned** — the run was stopped before it reported. The
   AC-HWD-013(c) inversion is in the tree and green, but the *observed passing
   output before the inversion*, which the lead's condition asked for, exists
   only in the stopped agent's transcript. **This is the one condition of the
   four that is not discharged in the record**, and it is the reason M4 should
   not be called closed without the next session either recovering that capture
   or re-establishing it.
3. **The stale-index path was not exercised.** Item 3 covers the absent-index
   case; a present-but-older-than-7-days index taking the same rebuild path is
   untested here.
4. **`internal/hook/file_changed.go:110-118` was deliberately left alone.** It
   carries the identical die-at-exit defect on the incremental path; it is a §G
   follow-up with no card, and M4 was scoped away from it.

#### Residual risk

- **The cost moved rather than vanished.** A cold `mx query` now pays the
  ~300 ms scan that the hook used to attempt and lose. That is the intended
  trade — the wait is attributable to the process that needs the index — but it
  is a new latency on a command that previously failed fast.
- **The 7-day freshness window is unchanged.** A 6-day-old index still passes as
  fresh and is served to `mx query` as authoritative; the d3 report records
  764 missing tags on a fresh worktree at the last measurement of that window.
  Whatever M4 fixed about cold start, staleness inside the window is untouched.
- **One consumer, one path.** `mx query` is the sole reader; if a second consumer
  appears it will need the same on-demand build, and nothing yet factors that out.

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase — M2 and M4 complete; M3 and M1 not started>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_

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

### M2 + M4 — re-measurement after the develop absorption

> Why this section exists: the M2 and M4 evidence above was measured at base
> commits `bf9aef6f3` and `1e75032b3`, on a tree 1044 commits behind
> `origin/develop`. Those numbers describe a tree that no longer exists. This
> section re-measures every claim on the merge tree and records where the old
> value did NOT reproduce. It supplements the sections above rather than
> replacing them — the originals stay as the record of what was true then.

#### Claim

Every M2 and M4 claim reproduces on the merge tree, with one recorded exception
(M2 evidence item 3's whole-run `rc=0`) whose cause is a check outside this
SPEC. Two failures the original measurements did not surface were found by the
full-package run, and both are repaired here.

#### Evidence

**Merge tree.** `git merge origin/develop` (`18ba3cddb`) into
`WT-hook-wiring-drift` produced merge commit `e8050c135`. Four conflicts, all in
files M2/M4 touched; resolution rationale is in that commit's message.

**Binary attribution.** `make build` on the merge tree stamped
`Commit=e8050c135` — identical to the merge commit — and every end-to-end
observation below used that `./bin/moai`. The installed `~/go/bin/moai` (v3.1.2)
was NOT used; a replacement of the shared binary was announced for later, and
nothing here is measured across that boundary.

| Claim (as recorded above) | Merge-tree measurement | Verdict |
|---|---|---|
| M2-1 nine HookWiring tests green | 9/9 PASS, `ok … 8.295s` | reproduces |
| M2-1b template hook-entry tests | `ok … 0.457s` | reproduces |
| M2-2 e2e reproduces D-1 both directions | `chain-event.sh`, `status-transition-ownership.sh x3` template-only; `status-transition-ownership.sh` project-only | reproduces |
| M2-3 `moai doctor` exits 0 | whole run `rc=1`; Hook Wiring in isolation `rc=0` | **does NOT reproduce — see below** |
| M2-7 build / vet / windows-vet / gofmt / lint | all exit 0, `0 issues.` | reproduces |
| M2-8 golden delta is the new row + recount | same delta, carried in `e8050c135` | reproduces |
| M2-9 full `internal/cli` green | `ok … 407.595s`, 0 FAIL, EXIT=0 (after the repairs below) | reproduces |
| M4-1 `grep -c 'runMXColdStartScan\|mxScanNeeded'` = 0 | `0`; no leftover symbol anywhere in `internal/hook/` | reproduces |
| M4-2 false comment absent (`rc=1`) | `rc=1` | reproduces |
| M4-3 cold `mx query` answers and leaves the index | `rc=0`, index written, **506,662 bytes** (was 378,793 at `1e75032b3` — 1044 commits of additional source) | reproduces |
| M4-4 `internal/hook` green | `rc=0`, `ok … 35.376s` | reproduces |
| M4-4b targeted MX suites | hook `ok … 1.499s`, cli `ok … 0.776s` | reproduces |

**M2 evidence item 3 — the one value that did not reproduce.** The recorded
observation was `./bin/moai doctor >/dev/null 2>&1; echo rc=$?` → `rc=0`. On the
merge tree the same command gives `rc=1`. The cause is attributed
mechanically, not inferred:

```
$ ./bin/moai doctor | grep -n fail
23:│    fail    Harness 5-Layer        L1:FAIL L2:FAIL L3:FAIL L4:FAIL L5:FAIL L6:FAIL
27:│    11 ok, 0 warn, 1 fail          ← MoAI-ADK section (contains Harness 5-Layer)
43:│    9 ok, 3 warn, 0 fail           ← Workspace section (contains Hook Wiring)

$ ./bin/moai doctor --check "Hook Wiring"; echo rc=$?
│    warn    Hook Wiring  hook wiring drift — template-only …
│    0 ok, 1 warn, 0 fail
rc=0
```

Isolated to its own check, the diagnostic exits 0 and reports `warn`. The
whole-run `rc=1` comes entirely from `Harness 5-Layer`, which this SPEC does not
touch. **AC-HWD-008's substance holds** — a drift is a `warn`, never a `fail`.
What no longer holds is the incidental whole-run exit code the original evidence
quoted, because that number was never a property of this check alone.

**Two failures the original measurements missed.** Both were found by the full
`internal/cli` run and neither is merge-introduced.

*F1 — `TestBinaryLag_DoctorCheckNameSetIsUnchanged`.* A guard belonging to the
binary-lag SPEC (REQ-BLV-009) that freezes the doctor check-name set against a
fixed baseline SHA (`22f90b1c7`). M2 registers `hookWiringCheckName`, so it
fires. Attribution was measured, not assumed: develop's own `doctor.go` was
copied into place and the guard re-run against it —

```
$ cp <origin/develop:internal/cli/doctor.go> internal/cli/doctor.go
$ go test -count=1 -run 'TestBinaryLag_DoctorCheckNameSetIsUnchanged' ./internal/cli/
ok  	github.com/modu-ai/moai-adk/internal/cli	1.027s      ← passes on develop
(restored; `git status --porcelain internal/cli/doctor.go` empty)
```

The guard passes on develop and fails only here, so M2 is the first change since
the baseline to add a check name. Its *stated* intent is narrower than its
implementation: it pins that the binary-lag SPEC itself registers no new name,
but measures against a fixed point, so every later SPEC that legitimately adds a
check trips it. Repaired with a documented one-entry allowlist
(`namesAddedAfterBaseline`). **`lagBaselineSHA` was deliberately NOT bumped** —
that would silence every other drift accumulated since the baseline, not just
this entry. The guard was then shown to retain its teeth under two mutations:

```
mutant A (register an unlisted name "Hook Wiring Mutant")  → FAIL, names it
mutant B (rename baseline entry "Git" → "Git Renamed")     → FAIL on both arms
                                                             (added + removed)
both reverted; guard green (`ok … 0.880s`)
```

*F2 — `TestSidecarUnavailable_StderrFormat`.* A second instance of the contract
M4 removed, missed because the targeted selector used during M4 and during the
first pass of this re-measurement (`TestMxQuery|TestMXIndex|TestBuildMXIndex`)
does not match its name. The file is byte-identical on both merge sides
(`git diff 8aa96bfb1 origin/develop -- internal/cli/mx_query_test.go` → empty),
so it was already red at M4's own HEAD. Repointed to the one path that still
errors after AC-HWD-012: a build that cannot succeed.

A sibling sweep was run before the repair rather than after, since missing a
sibling is what produced F2 in the first place:

```
$ grep -rn 'SidecarUnavailable|moai mx scan|사이드카 없을 때' internal/ | grep _test.go
```

The only remaining hits are `internal/mx/resolver_query_test.go`, which tests the
**resolver-level** error — unchanged by M4, and still correct
(`go test ./internal/mx/` → `ok … 5.044s`).

**Both repaired tests were shown non-vacuous, and one was caught vacuous first.**

| Test | Mutation | Result |
|---|---|---|
| `TestMxQueryCmd_AbsentSidecarIsBuiltOnDemand` | restore the old `SidecarUnavailable` error path | RED (restored → GREEN) |
| same, "index left behind" half | disable M4's build block only | **PASS — vacuous** |
| `TestSidecarUnavailable_StderrFormat` (first form) | swallow the build error | **PASS — vacuous** |
| same, after tightening | swallow the build error | RED (restored → GREEN) |

The two vacuous results share one cause and it is recorded rather than tidied
away: develop added a second writer of the same artifact later in the same
command (`graph.MXIndexNeedsRefresh` → `mx.RefreshIndex`, `mx_query.go:118-119`).
Disabling M4's build block still leaves an index behind, and swallowing its error
still yields a `SidecarUnavailable` from the resolver. The stderr test was
tightened to assert the command-layer string `could not build the sidecar index`,
which only M4's branch emits; the "index left behind" assertion in the other test
is left as-is and is **declared non-load-bearing** here rather than presented as
proof.

**Static re-verification after the repairs.**

```
$ go build ./...                → exit 0        $ gofmt -l <edited files>   → no output
$ go vet ./...                  → exit 0        $ GOOS=windows go vet ./... → exit 0
$ golangci-lint run --timeout=5m ./internal/cli/... ./internal/hook/...   → 0 issues.
$ go test -count=1 -timeout 1800s ./internal/cli/  → ok … 407.595s, EXIT=0, 0 FAIL
```

**A note on how the full-suite verdict was read.** The background runner reported
`exit code 0` for the pre-repair run while the suite had in fact failed: the
command ended in `tail`, whose status became the wrapper's. The verdict was taken
from the log's own `ok`/`FAIL` line and `EXIT=` marker instead — the same hazard
M2 evidence item 9 already recorded, hit again from the other direction.

#### Baseline-attribution

Every measurement above was taken in this tree
(`.claude/worktrees/t216`), branch `WT-hook-wiring-drift`, at merge commit
`e8050c135` plus the two test repairs described here, with `bin/moai` built from
that same tree (`Commit=e8050c135`). No figure is carried over from the
`bf9aef6f3` / `1e75032b3` measurements; where a figure differs from the original
it is named as differing above.

#### Gaps — what was explicitly NOT observed

1. **`internal/hook` was not re-run after the two test repairs.** Both repairs
   are in `internal/cli`; the `rc=0 / 35.376s` figure predates them and is
   unaffected by construction, but it was not re-executed.
2. **The edges-refresh landing rate was not measured.** M4 measured the MX
   cold-start scan's artifact in 2 of 153 worktrees. No equivalent measurement
   exists for the edges refresh develop added to the same goroutine; the claim
   that it shares the defect class is a **code reading**, not an observation.
3. **`golangci-lint` remains scoped to the touched packages**, not the repository.
4. **Windows runtime is still unobserved** — `GOOS=windows go vet` proves
   compilation only.
5. **The M4 gap-2 capture is still open.** The delegate's pre-implementation
   mutant record and the AC-HWD-013(c) RED-before-inversion capture were never
   returned. This re-measurement adds mutation evidence for the two tests it
   repaired; it does NOT retroactively supply that capture.
6. **The stale-index path is still unexercised** (M4 gap 3 stands): the cold path
   covers absence, not a present-but-expired index.
7. **M1 and M3 remain not started.**

#### Residual risk

- **Two code paths now build the same index in one command.** M4's on-demand
  build and develop's graph refresh both write `mx-index.json`. Nothing yet
  factors that out, and the redundancy is what makes one assertion above
  unfalsifiable. Reconciling them is outside M4's scope and is not attempted.
- **A durable side effect is dispatched from the deferred goroutine again.** The
  edges refresh occupies the position M4 emptied, for the reason M4 documented.
  Preserved deliberately — it belongs to another SPEC — with the comments
  corrected to describe the tear-down accurately. See §G.
- **The allowlist in the binary-lag guard is a shared-surface edit.** It is one
  entry, documented, and mutation-checked, but it modifies a guard this SPEC does
  not own. If the binary-lag SPEC's owner prefers a different resolution, this is
  the line to revisit.
- **`Harness 5-Layer` fails on this checkout**, which makes `moai doctor`'s
  whole-run exit code 1 for reasons unrelated to this SPEC. Any future criterion
  quoting a whole-run exit code will inherit that.

### M3 — per-script dispositions recorded

#### Claim

`hook-independence.md` and its template twin carry one row per present-but-unwired
wrapper, each with exactly one of the five disposition classes and the class the
plan assigns it. The `team-ac-verify.sh` wording is corrected on **all six**
surfaces to say *not registered* rather than *registered but gated*. The two
counting corrections are recorded with their cause. Nothing was deleted.

#### Evidence

**Template-First order.** All three template files were edited first, `make build`
ran (`rc=0`, `catalog.yaml updated successfully`, binary stamped
`Commit=9a1434912…-dirty`), and only then were the mirrors updated.

**AC-HWD-009 — machine-checked, name by name, on both sides of the pair.** The
criterion is not "the names appear somewhere"; it is one own-row per wrapper,
naming that wrapper and no other of the eleven, carrying exactly one class, and
that class being the expected one. A checker asserts all four properties:
`.moai/reports/t216/m3-evidence/ac009_check.py`, output in `ac009-result.txt`.

```
== .claude/rules/moai/development/hook-independence.md
  ok   chain-event.sh                    line 109  reachable-via-template-settings
  ok   handle-agent-hook.sh              line 110  reachable-via-agent-frontmatter
  ok   handle-session-start-compact.sh   line 111  reachable-via-in-binary-registry
  ok   handle-elicitation.sh             line 112  dead-by-decision
  ok   handle-elicitation-result.sh      line 113  dead-by-decision
  ok   handle-notification.sh            line 114  dead-by-decision
  ok   handle-task-created.sh            line 115  dead-by-decision
  ok   handle-worktree-create.sh         line 116  dead-by-decision
  ok   handle-worktree-remove.sh         line 117  dead-by-decision
  ok   handle-session-start-navigator.sh line 118  open-question
  ok   team-ac-verify.sh                 line 119  open-question
== internal/template/templates/.claude/rules/moai/development/hook-independence.md
  (identical 11 rows, same line numbers)
TOTAL FAILURES: 0
```

The checker was shown to have teeth before its zero was believed:

| Mutation | Result |
|---|---|
| expect `dead-by-decision` for `team-ac-verify.sh` | FAIL, names the class mismatch (both sides) |
| put two classes on the `handle-notification.sh` row | FAIL, names the row and both classes |

Both reverted; the pair is byte-identical again and the checker returns 0.

**AC-HWD-018 — all six surfaces, measured rather than assumed.** For each file,
the count of lines that mention `team-ac-verify` AND contain "dormant":

```
0  .claude/rules/moai/development/hook-independence.md
0  .claude/rules/moai/core/agent-common-protocol.md
0  .claude/rules/moai/core/agent-common-protocol-reference.md
0  internal/template/templates/.claude/rules/moai/development/hook-independence.md
0  internal/template/templates/.claude/rules/moai/core/agent-common-protocol.md
0  internal/template/templates/.claude/rules/moai/core/agent-common-protocol-reference.md
```

**Negative control — the word survived where it was correct.** The [HARD] warning
against a blanket sweep was checked, not just obeyed: `hook-independence.md` still
carries two uses of "dormant", at `:105` (naming the old shared label as the thing
that conflated three states) and `:139` (the `moai hook spec-status` subcommand,
which genuinely is dormant and is not one of the eleven wrappers). Neither was
touched.

**AC-HWD-010 — both corrections, with the cause.** The rule file now states the
`statusLine` cause (`grep -c statusLine` → 1) and the frontmatter registration
(`grep -c handle-agent-hook` → 2), against a pre-impl baseline of 0 for both.

The independent basis was re-derived here rather than copied from the plan, and
the first attempt at it was wrong in a way worth recording. Counting
`len(hooks[event])` gives **23** — that is the number of *matcher groups*, not
entries. The settings shape is `hooks[event] -> [ {matcher, hooks:[…]} ]`, so the
entry count is the sum over groups of `len(group["hooks"])`:

```
$ python3 .moai/reports/t216/m3-evidence/count_entries.py
events        : 20
matcher groups: 23
hook entries  : 33
statusLine present: True

$ grep -c '"type": "command"' .claude/settings.json
34
```

33 entries across 20 events, grep 34, `statusLine` present — the documented
claim and its stated cause both hold. The 23 is recorded because a reader
re-deriving the number the obvious way will land on it and conclude the document
is wrong.

**AC-HWD-011 — nothing deleted, measured against the recorded pre-impl values.**

```
$ git ls-files .claude/hooks/moai/ | wc -l                                  43   (was 43)
$ git ls-files internal/template/templates/.claude/hooks/moai/ | wc -l      47   (was 47)
$ git status --porcelain <both paths>                                       (empty)
$ find .claude/hooks/moai -name '*.sh' -size 0 | wc -l                       0
```

**AC-HWD-016 — neutrality.** On every template-side file M3 touched:

```
hook-independence.md              SPEC-/REQ- 0   card-number 0
agent-common-protocol.md          SPEC-/REQ- 1   ← pre-existing, see below
agent-common-protocol-reference.md SPEC-/REQ- 0
$ go test ./internal/template/...  → ok 25.817s, ok 0.881s (agentemit)
```

The single hit is `agent-common-protocol.md:301`, the string `<SPEC-ID>` inside a
command example (`moai session list --json --filter-spec=<SPEC-ID>`). It is a
placeholder, not an internal identifier, and the count is **1 before and 1 after**
this edit (`git show HEAD:<file> | grep -c` → 1), so M3 introduced nothing.

**AC-HWD-015 — mirror identity, and where it cannot hold.**

```
IDENTICAL  hook-independence.md
IDENTICAL  agent-common-protocol.md
DIFFERS    agent-common-protocol-reference.md   (1 line)
```

The third pair differed **before this milestone touched it** and the difference is
deliberate: the local file's provenance note cites a SPEC ID, which template
neutrality forbids the template twin from carrying. The full residual diff is that
one line and nothing else — M3's own edit landed identically on both sides, which
is what the criterion exists to protect. Recorded as a criterion defect below
rather than worked around silently.

#### Baseline-attribution

Measured in this tree (`.claude/worktrees/t216`), branch `WT-hook-wiring-drift`,
at `9a1434912` plus this milestone's change set — six files, three template
sources and their three mirrors. `bin/moai` rebuilt from this tree between the
template edit and the mirror. Every number above was produced by a command run
here; none is carried from the plan's pre-impl record except where explicitly
compared against it.

#### Gaps — what was explicitly NOT observed

1. **The dispositions were transcribed from the plan, not re-derived.** The plan's
   table came from the D-2 call-graph work; M3 records it. This milestone did not
   independently re-trace the eleven call paths, so a wrong disposition upstream
   stays wrong here.
2. **`reachable-via-*` was not demonstrated by firing anything.** No wrapper was
   observed actually executing through the named path; the classes rest on the
   D-2 reading.
3. **The two `open-question` rows record that nobody decided, which is a claim
   about absence.** It was checked against settings surfaces only.
4. **No runtime test asserts the disposition table.** The AC-HWD-009 checker is a
   verification artifact under `.moai/reports/`, not a committed guard, so the
   table can drift from the wrappers without anything failing.
5. **`agent-common-protocol.md` is always-loaded**; its corrected line was read
   but no measurement confirms the file still fits its size budget.

#### Residual risk

- **A criterion defect, not a measurement problem: AC-HWD-015 is unsatisfiable as
  written.** It requires byte identity for every pair M3 touches, while template
  neutrality requires the reference pair to differ. The two [HARD] rules conflict
  on that one file. The reading applied here — M3's own edit must land on both
  sides, and the only permitted residual delta is a pre-existing neutrality strip
  — preserves what the criterion protects. It is an interpretation, and the plan
  should say so explicitly.
- **A second plan-side inconsistency, resolved toward the plan body.** AC-HWD-018
  instructs leaving `:106` (the worktree wrappers' "dormant … activation
  deferred") alone as accurate, while plan.md §F M3 replaces the block containing
  it and assigns those wrappers `dead-by-decision`. Both cannot hold. The block
  was replaced and the statement sharpened — "deregistered after a recorded
  regression" is stronger and more accurate than "activation deferred", which
  implies a pending activation that the regression record contradicts. Nothing
  accurate was lost, but the criterion's own example is now stale.
- **The table is prose, and prose drifts.** Nothing mechanically ties a row to the
  wrapper's real reachability. The next wrapper added or retired will not update
  this table, and no test will notice — see gap 4.

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase — M2 and M4 complete AND re-measured on the merge tree;
M3 and M1 not started>_

```yaml
run_status: in-progress
milestones_complete: [M2, M3, M4]
milestones_remaining: [M1]
measured_at_head: e8050c135
measured_at_head_note: >-
  merge commit absorbing origin/develop (18ba3cddb) into WT-hook-wiring-drift,
  plus the two test repairs recorded in §E.2 "M2 + M4 — re-measurement after
  the develop absorption". The M2/M4 evidence sections above remain attributed
  to bf9aef6f3 / 1e75032b3; do not read their figures as current.
binary_attribution: bin/moai built from this tree, Commit=e8050c135
suites:
  - internal/cli: ok 407.595s, 0 FAIL, EXIT=0
  - internal/hook: ok 35.376s (pre-repair; repairs are cli-only)
  - internal/mx: ok 5.044s
static: go build / go vet / GOOS=windows go vet exit 0; gofmt clean; golangci-lint 0 issues
open_gaps: [M4-gap-2 mutant capture, M4-gap-3 stale-index path, edges-refresh landing rate unmeasured]
```

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_

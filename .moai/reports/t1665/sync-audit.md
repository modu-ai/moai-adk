auditor-model: claude-opus-5-5

# t1665 sync audit — doctor profile-cap read path (Class B, no SPEC)

verdict: PASS
audited_sha: 44d46d13b30f353f26419140a3b4d68bf29ef9a0

## Evaluation Report
SPEC: none (Class B card t1665)
Overall Verdict: PASS — overall score 91/100 (harmonic over the 4 dimensions below)
Auditor: sync-auditor, independent, iteration 1 (no prior sync audit on this card)
Cross-model: `audit_multi` codex gate=required → `pass`, receipt `rcpt-6b7bda93b8330b081c691a7e`, plan_source=config

### Dimension Scores
| Dimension | Score | Verdict | Evidence |
|-----------|-------|---------|----------|
| Functionality (40%) | 92/100 | PASS | targeted tests green (3 new subtests); RED→GREEN pair present; 2 of 4 overlay mutants killed, 2 survive (F1, F2 — test gaps, not code defects) |
| Security (25%) | 96/100 | PASS | no new input surface beyond a local YAML read already owned by `hygiene`; loader error string printed to the local doctor detail only; no secrets/exec/network |
| Craft (20%) | 88/100 | PASS | `gatherHomeDiskReport` 95.5% / `renderHomeDiskDetail` 97.0% (targeted run); `golangci-lint --new-from-rev` 0 issues; vet/gofmt clean; @MX:NOTE carries the no-fallback contract |
| Consistency (15%) | 90/100 | PASS | same loader + same project-dir resolver the CLI already uses; style matches file; one comment drift in a test file (F3) |

Must-pass firewall (Functionality + Security): both pass independently.

### Findings (structured defect-list)
- F1 [Medium, confidence high] [optional] internal/cli/doctor_disk.go:142 / internal/cli/doctor_disk_test.go:234-298 — the PROJECT-tier wiring is untested at the doctor level. Overlay mutant M1 replacing `hygiene.LoadProfileCapMaxBytes(resolveProjectDir())` with `hygiene.LoadProfileCapMaxBytes("")` passes every doctor test (exit 0). The shipped default (`profile_max_bytes: 8589934592`) lives in the project tier (`internal/template/templates/.moai/config/sections/state.yaml:21`), so a regression to HOME-only reading would silently reopen the doctor/advisory divergence for every project-tier-only user. The production code is correct today (read at line 142); the loader's own project-tier fallback is covered by `TestProfileCapWarnsWithoutDeleting/project-tier_fallback...`. Required fix (if pursued): add a subtest that leaves the HOME tier without the key and writes `profile_max_bytes` into `$CLAUDE_PROJECT_DIR/.moai/config/sections/state.yaml`, asserting the profile above it is flagged.
- F2 [Low, confidence high] [optional] internal/cli/doctor_disk_test.go:262-266 — the "key absent" subtest uses a 6 GiB fixture whose comment calls it "above any plausible compiled cap"; overlay mutant M2 (absent key → compiled 8 GiB fallback, the shipped template value) survives (exit 0). Only a ≤6 GiB fallback (M3, 5 GiB) is killed. Required fix (if pursued): size the fixture above 8 GiB (e.g. 9 GiB sparse) or reword the comment to the cap it actually excludes.
- F3 [Low, confidence high] [optional] internal/cli/clean_home_profile_cap_test.go:6 — comment still names "the compiled DefaultProfileMaxBytes", which no longer exists after this card (comment-only; outside the card's 4 files; scope discipline argues leaving it unless a later card touches the file).
- F4 [Low, confidence medium] [optional] internal/cli/doctor_disk.go:142 — `resolveProjectDir()` falls back to the cwd when `CLAUDE_PROJECT_DIR` is unset, so `moai doctor` run from a subdirectory outside a Claude session misses the project tier while the SessionStart advisory (hook `input.ProjectDir`) reads it. Pre-existing resolver, chosen and recorded in the lane's decision record; not a regression.

No blocking findings.

### Recommendations
- F1 is the one worth a follow-up test: it is the exact shipped-config path the card was opened for.
- F2/F3 are cosmetic; batch with any later touch of those files.

---

## Claim
1. doctor now reads the per-profile cap from `state.profile_max_bytes` through `hygiene.LoadProfileCapMaxBytes`, the same loader the SessionStart advisory uses, with the same no-compiled-fallback contract (absent key → no oversize check; loader error → no check, error named, never fatal).
2. One source of truth for the per-profile cap remains: no non-test Go code references a compiled per-profile cap; `config.DefaultProfileMaxBytes` is removed and nothing references it.
3. The new test proves "changing the setting changes doctor output" for the HOME tier (cap set → exactly the above-cap profile flagged, detail prints the cap; key absent → none; loader error → none + error named), isolates HOME (`HOME`, `MOAI_HOME=""`) and the project tier (`CLAUDE_PROJECT_DIR` = empty temp dir).
4. Scope: 4 code/test files + 2 card records; no behaviour beyond the profile cap changed. doctor stays advisory.
5. Build, vet, gofmt, lint, targeted tests, and hygiene/config packages are green on the audited tree.

## Evidence

Tree and commits:
```
$ git rev-parse HEAD
44d46d13b30f353f26419140a3b4d68bf29ef9a0
$ git log --oneline a4130a809..HEAD
44d46d13b docs(t1665): record cause, run evidence and card review (card t1665)
27ea20974 fix(doctor): read state.profile_max_bytes instead of the compiled profile cap (card t1665)
ec8c2bd58 test(doctor): RED for doctor reading the configured profile cap (card t1665)
$ git diff --stat a4130a809..HEAD
 .moai/reports/t1665/card-review.md | 10 +++++
 .moai/reports/t1665/evidence.md    | 42 ++++++++++++++++++
 internal/cli/doctor_disk.go        | 25 ++++++++++-
 internal/cli/doctor_disk_test.go   | 91 +++++++++++++++++++++++++++++++++++++-
 internal/config/defaults.go        |  1 -
 internal/hygiene/profile_cap.go    |  5 ++-
 6 files changed, 168 insertions(+), 6 deletions(-)
```

Cause confirmed at base (the two lines the card names):
```
$ git show a4130a809:internal/cli/doctor_disk.go | sed -n '155p;261p'
			if size > config.DefaultProfileMaxBytes {
		lines = append(lines, fmt.Sprintf("profiles above %s: %s", formatDiskBytes(config.DefaultProfileMaxBytes), strings.Join(report.OversizeProfiles, ", ")))
```

Fix (read at HEAD, internal/cli/doctor_disk.go:142-146, 173, 279-283):
```go
if capBytes, ok, err := hygiene.LoadProfileCapMaxBytes(resolveProjectDir()); err != nil {
	report.ProfileCapErr = err.Error()
} else if ok {
	report.ProfileCapBytes = capBytes
}
...
if report.ProfileCapBytes > 0 && size > report.ProfileCapBytes {
...
lines = append(lines, fmt.Sprintf("profiles above %s: %s", formatDiskBytes(report.ProfileCapBytes), ...))
if report.ProfileCapErr != "" {
	lines = append(lines, fmt.Sprintf("profile size cap unreadable — oversize check skipped: %s", report.ProfileCapErr))
```
Advisory counterpart (internal/hook/session_start_hygiene.go:108-115): `capBytes, ok, err := hygiene.LoadProfileCapMaxBytes(projectDir)`; `err` → skip; `!ok` → `return // absent key: no warning (REQ-CAP-009); no compiled fallback`. Same loader, same degradation shape. Doctor's loader error only adds a detail line; `checkHomeDisk` status is unaffected (no other consumer of `ProfileCapErr`/`ProfileCapBytes`: grep below).

Single source of truth:
```
$ grep -rn "DefaultProfileMaxBytes" . --include='*.go'
internal/hygiene/profile_cap.go:51:// config.DefaultProfileMaxBytes was removed, card t1665) — a compiled
internal/cli/clean_home_profile_cap_test.go:6:// warning — the compiled DefaultProfileMaxBytes is no longer a candidate
```
Both hits are comments. Non-Go hits (whole tree, excluding .git) are only historical records under `.moai/specs/SPEC-CAPACITY-RECLAIM-001/` (progress.md, acceptance.md, plan.md); no template, YAML, or doc carries the constant.
```
$ grep -n "OversizeProfiles\|ProfileCapErr\|ProfileCapBytes" internal/cli/*.go | grep -v _test
internal/cli/doctor_disk.go:67 / :71 / :74 (fields) / :143 / :145 / :173 / :174 / :278 / :279 / :281 / :282
```

Targeted tests (TMPDIR=GOTMPDIR=/private/tmp/dcap-audit):
```
$ go test -count=1 -v -run 'HomeDisk|ProfileCap|DoctorDisk' ./internal/cli/   → exit 0
--- PASS: TestProfileCapWarnsWithoutDeleting (0.01s)
    --- PASS: .../cap_set:_warning_names_profile_and_key;_no_projects_candidate;_nothing_deleted (0.00s)
    --- PASS: .../key_absent:_no_warning_and_no_projects_candidate (0.00s)
    --- PASS: .../project-tier_fallback:_home_key-absent_state.yaml_falls_through_to_the_project_sections (0.00s)
    --- PASS: .../unreadable_subdirectory:_over-cap_profile_still_warns_with_partial-size_semantics (0.00s)
--- PASS: TestHomeDiskReportDetectsInactiveOversizeAndInsecureProfiles (0.00s)
--- PASS: TestHomeDiskReportOversizeUsesConfiguredProfileCap (0.01s)
    --- PASS: .../configured_cap_flags_exactly_the_profiles_above_it (0.00s)
    --- PASS: .../key_absent_in_every_tier_flags_nothing (0.00s)
    --- PASS: .../loader_error_skips_the_check_and_names_the_error (0.00s)
--- PASS: TestCheckHomeDisk_WarnsWhenCleanableExceedsDefaultThreshold (0.00s)
--- PASS: TestCheckHomeDisk_OKWhenCleanableBelowDefaultThreshold (0.00s)
--- PASS: TestCheckHomeDisk_RegisteredInMoaiADKGroup (3.03s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/cli	3.740s

$ go test -count=1 ./internal/hygiene/ ./internal/config/
ok  	github.com/modu-ai/moai-adk/internal/hygiene	0.439s
ok  	github.com/modu-ai/moai-adk/internal/config	2.417s
exit 0
$ go build ./...                                          → build exit 0
$ go vet ./internal/cli/ ./internal/hygiene/ ./internal/config/   → vet exit 0
$ gofmt -l <the 4 changed .go files>                       → (empty), exit 0
$ golangci-lint run --new-from-rev a4130a809 ./internal/cli/ ./internal/config/ ./internal/hygiene/
0 issues.   (exit 0; golangci-lint v2.1.6)
```

Coverage of the changed functions (targeted run, `-coverprofile` + `go tool cover -func`):
```
internal/cli/doctor_disk.go:130:  gatherHomeDiskReport   95.5%
internal/cli/doctor_disk.go:251:  renderHomeDiskDetail   97.0%
```

Mutation probe via `go test -overlay` (worktree never modified; script and mutant files in the session scratchpad):
```
== m1 (LoadProfileCapMaxBytes("") — drop PROJECT tier)       exit 0  ok  internal/cli 4.026s        → SURVIVES (F1)
== m2 (absent key → 8 << 30 fallback)                          exit 0  ok  internal/cli 3.878s        → SURVIVES (F2)
== m3 (absent key → 5 << 30 fallback)                          exit 1
    --- FAIL: .../key_absent_in_every_tier_flags_nothing
        doctor_disk_test.go:270: oversize profiles = [huge], want none when state.profile_max_bytes is absent
== m4 (loader error swallowed: `_ = err`)                     exit 1
    --- FAIL: .../loader_error_skips_the_check_and_names_the_error
        doctor_disk_test.go:294: detail must name the loader error:
```
The "ignore the config, keep the compiled 5 GiB cap" mutant is the base code itself; the lane's RED commit ec8c2bd58 records its 3 failures (evidence.md run record), consistent with m3.

Audit plan and cross-model review:
```
$ moai verify audit-plan --project-root /Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1665
"config_status": "ok", "model": "codex", backends: claude gate=off, codex gate=required, glm gate=off,
"cross_model_active": true, "cross_model_required": true, "enforced_required": ["codex"]   (exit 0)

mcp__moai__audit_multi(project_root=<this tree>, target=baseBranch, no gates arg)
overall_verdict: pass
gate_unmet: (absent/empty)
plan_source: config
per_backend_verdicts: [{backend: codex, gate: required, verdict: pass}]
audit_receipt: rcpt-6b7bda93b8330b081c691a7e
build_commit: ee981a932 (ancestor of HEAD 44d46d13b — see Gaps)
```
This verdict is not final until the orchestrator's `moai verify audit-plan --result-file` convergence check passes on the digest above.

## Baseline-attribution
- Tree: `/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1665`, branch WT-t1631-d-10, HEAD `44d46d13b30f353f26419140a3b4d68bf29ef9a0`, clean working tree (`git status --short` empty before and after the run). Base = local `main` = merge-base = `a4130a809899b3ecb6e05120c0c2c8423bf7f424`.
- Go test/build/vet/gofmt/lint/coverage: toolchain on PATH, run in this session against this tree; all builds compiled from this tree's sources (Go toolchain, not a moai binary), so §2.2 tool-lag does not apply to them.
- `moai verify audit-plan` and `audit_multi`: installed moai binary / MCP server v3.2.0-rc.36-t1822c, commit `ee981a932` (see Gaps).

## Gaps
- Full `internal/cli` suite, repo-wide lint, and non-darwin platforms not run (scoped per instruction; CI owns them).
- moai tool provenance: the audit-plan reader and the `audit_multi` server run build `ee981a932`, a strict ancestor of the audited HEAD. This card touches no audit/plan code, so the lag cannot change these results, but the provenance is recorded per verification-claim-integrity §2.2.
- The Bash worktree guard refused the first, piped form of the targeted test command; it was re-run as a plain redirect-to-file command (output quoted above) — the measurement itself was not substituted.
- `audit_multi` was called without `card_id` (no second-review.jsonl written), because the dispatch did not state the card runs under contract autonomy.
- The codex review is static (its card-review sibling states it ran no tests); its pass is corroborating, not a substitute for the measurements above.
- Lane-reported implementer runs (`go test ./internal/config/...`, etc.) were not relied on; each claim above was re-measured here.

## Residual-risk
- F1: a future change dropping the project-dir argument would pass CI and silently restore the shipped-config divergence (doctor 0 flags vs advisory at 8 GiB).
- F4: doctor and advisory can still resolve different project roots when `moai doctor` runs outside a Claude session from a subdirectory.
- Sparse-file fixtures assume a filesystem that reports logical size from `Info().Size()`; on such filesystems the 6 GiB fixture costs no disk, elsewhere it could be slow (not observed here).

## Iteration history
- Iteration 1 (this audit): PASS, 0 blocking, 4 optional findings.

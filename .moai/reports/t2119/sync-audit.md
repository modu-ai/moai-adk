auditor-model: claude-opus-5-5
verdict: PASS
score: 0.91
audited_sha: b695b89b93c3881355d6cafef95cc6948c4434c3
audit_receipt: rcpt-8133aa42141ea9d1f940b19d

# t2119 sync-audit — #1969 Windows race residuals, first slice

## Iteration 2 (re-audit) — PASS 0.91 @b695b89b93

Scope: the repair delta `703a5b4035..b695b89b93` (test file +7/-4, progress.md +5, this report),
judged together with the unchanged production code at HEAD. `durable.go` has no change in the delta
(`git diff --name-only 703a5b4035..HEAD -- internal/agenticfactory/claim/durable.go | wc -l` → `0`).

### Dimension scores (iteration 2)

| Dimension | Score | Verdict | Evidence |
|---|---|---|---|
| Functionality (40%) | 92/100 | PASS | F1 resolved: the existing-directory leg now asks `dirSyncFailureIsFatal(runtime.GOOS, err)`; targeted tests `ok 0.233s`; mutants MA/MB/MC killed |
| Security (25%) | 95/100 | PASS | Delta is test-only plus records; the production tolerance is unchanged (`goos=="windows" && errors.Is(err, fs.ErrPermission)`) |
| Craft (20%) | 85/100 | PASS | `GOOS=windows go vet` exit 0, `gofmt -l` empty; F2 (call-site GOOS pin) still optional and open |
| Consistency (15%) | 90/100 | PASS | The test uses the same decision function as the call site `durable.go:72`; comment updated to explain why |

Weighted: 0.4·92 + 0.25·95 + 0.2·85 + 0.15·90 = 91.05 → **0.91**. Must-pass Functionality and Security both pass.

### Findings (iteration 2)

- **F1** [High] [blocking] [in-diff] — **resolved.** `durable_dirsync_t2119_test.go:44` now reads
  `if err := y.syncDir(t.TempDir()); dirSyncFailureIsFatal(runtime.GOOS, err) {`. On Windows the
  real call yields `*PathError{Op:"sync", Err: ERROR_ACCESS_DENIED}`, `errors.Is(…, fs.ErrPermission)`
  is true (stdlib `Errno.Is`, iteration-1 evidence), so the decision returns false and the leg passes;
  a nil result also passes. The missing-directory leg (line 47) is unchanged and still requires a
  non-nil error: on Windows `os.Open` of a non-existent child returns `ERROR_FILE_NOT_FOUND`, which is
  non-nil, so the leg passes there and still fails a success-on-missing mutant (MA). Other assertions
  in the file: `TestDirSyncFailureIsFatal` is a pure table over explicit `goos` strings and has no
  platform dependence. Repair-born check: no blocking finding inside the repair delta.
- **F2** [Low] [optional] [in-diff, unchanged] — the `runtime.GOOS` argument at `durable.go:72` is
  still not pinned by a test (iteration-1 mutant M6). The lane recorded it as deliberately left
  (progress.md). Discretionary; no verdict effect.
- **F3** [Low] [optional] — **resolved.** progress.md now records the round, the F1 cause and the repair.
- **N1** [Info] [optional] [in-diff] — the new test comment says Windows "answers 'Access is denied'",
  which rests on the CI error text and stdlib source, not a Windows run by this auditor (see Gaps).

### Claim / Evidence / Baseline-attribution / Gaps / Residual-risk (iteration 2)

**Claim.** (1) F1 is resolved: the existing-directory leg cannot fail on Windows for the tolerated
refusal, and the missing-directory leg still fails a success. (2) The test still kills a
missing-directory success, a non-Windows tolerance, and a hard error on an existing directory.
(3) Vet for Windows and gofmt are clean. (4) Codex finds no new defect in the delta.

**Evidence.**
- `git rev-parse HEAD` → `b695b89b93c3881355d6cafef95cc6948c4434c3`
- `go test ./internal/agenticfactory/claim -run 'TestDirSyncFailureIsFatal|TestSyncDirDefaultPath' -count=1 -v`:
  ```
  --- PASS: TestDirSyncFailureIsFatal (0.00s)
  --- PASS: TestSyncDirDefaultPath (0.01s)
  PASS
  ok  	github.com/modu-ai/moai-adk/internal/agenticfactory/claim	0.233s
  ```
- `GOOS=windows go vet ./internal/agenticfactory/claim` → `vet_exit=0`; `gofmt -l internal/agenticfactory/claim` → no output, `gofmt_exit=0`.
- Mutation via `go test -overlay` (replacing `durable.go`; run filter = the two new tests):
  - MA — `syncDir` returns nil when `os.Open` fails (missing directory reports success):
    ```
    --- FAIL: TestSyncDirDefaultPath (0.02s)
        durable_dirsync_t2119_test.go:48: syncDir on a missing directory reported success
    ```
    killed.
  - MB — `dirSyncFailureIsFatal` tolerates `ErrPermission` on every platform (linux tolerance):
    ```
    --- FAIL: TestDirSyncFailureIsFatal (0.00s)
        durable_dirsync_t2119_test.go:32: linux access denied is fatal: dirSyncFailureIsFatal("linux", sync C:\work\docs\checkpoint: permission denied) = false, want true
        durable_dirsync_t2119_test.go:32: darwin access denied is fatal: dirSyncFailureIsFatal("darwin", sync C:\work\docs\checkpoint: permission denied) = false, want true
    ```
    killed.
  - MC — existing directory's fsync returns a non-permission I/O error (checks the repaired leg still bites):
    ```
    --- FAIL: TestSyncDirDefaultPath (0.01s)
        durable_dirsync_t2119_test.go:45: syncDir on an existing directory: input/output error
    ```
    killed.
- Codex (`mcp__moai__codex_audit`, native, target `baseBranch`, `delta_base` 703a5b4035 →
  `"review_base":"delta_base (merge base 703a5b40352cc9c4fa0d52e5ca4cca640da2f215)"`), verbatim:
  `"verdict":"pass"`, `"findings":[]`, summary `지정한 기준 커밋 이후 변경에서 새 결함은 찾지 못했네. 정적 검토만 했으며 테스트·빌드·재현 명령은 실행하지 않았네.`,
  `"resolved_model":"gpt-6.1-sol"`, `"audit_receipt":"rcpt-8133aa42141ea9d1f940b19d"`. One call, as instructed.
- `moai verify audit-plan --project-root <tree>` → `config_status: ok`, model `codex`, codex `required`, claude/glm `off`, `enforced_required: ["codex"]`.

**Baseline-attribution.** All commands ran in this run against tree t2119 at HEAD b695b89b93, on
darwin/arm64, go1.27.3. The moai binary serving the audit plan and codex_audit was built from
04d8c27a7 (`build_lag` reported by the tool: an ancestor of HEAD). The delta changes no moai audit
code, so the lag does not affect these two results. Tests, vet, gofmt and mutants used the Go
toolchain directly on this tree.

**Gaps.**
- No Windows host: the Windows behaviour of both legs rests on stdlib source reasoning (iteration 1:
  `syscall_windows.go` `Errno.Is`, `ERROR_ACCESS_DENIED = 5`) and the CI error text, not on an
  executed Windows run. Cross-compiled vet covers type-checking only.
- The full claim package suite was not re-run in this iteration (production code unchanged since the
  iteration-1 run `ok 149.610s`); only the two affected tests and three mutants ran.
- Codex reported a static review only (no tests run).
- `audit_multi` convergence check not run (forbidden by the lead's spawn instructions).
- Mutant overlay files live in session scratch; their verbatim outcomes are quoted above, the files themselves are not exported.

**Residual-risk.**
- If a Windows runner returns a non-`ErrPermission` errno for the directory fsync (for example a
  sharing violation), both the production write and the existing-directory leg fail — consistently,
  which is the intended signal, but the CI symptom would remain.
- F2: a regression hard-coding the GOOS argument at the call site would not be caught.
- On Windows, directory-entry durability after a rename rests on NTFS metadata journaling (accepted by design, as in iteration 1).

---

## Iteration 1 (historical) — FAIL 0.74 @703a5b4035

Overall verdict (iteration 1): **FAIL** (must-pass Functionality fails on one blocking, in-diff defect). Score 0.74.
Scope: commit 703a5b4035 only (durable.go, durable_dirsync_t2119_test.go, progress.md).
Receipt: rcpt-47cdae1a479a0b784de770af.

## Summary

The production change is correct. `dirSyncFailureIsFatal` tolerates exactly the Windows
`Access is denied` directory fsync and keeps every other case fatal. The mutation checks show the
decision function is well pinned. But the new test `TestSyncDirDefaultPath` calls the raw
`syncer{}.syncDir` on an existing directory and requires `nil`. On Windows that is the very call
that produces `sync C:\...: Access is denied`. So the slice adds a new deterministic Windows
failure on the platform it is meant to repair. Codex found the same defect on its own (P1).

## Dimension scores

| Dimension | Score | Verdict | Evidence |
|---|---|---|---|
| Functionality (40%) | 60/100 | FAIL | Production fix correct (stdlib trace below); new test fails on Windows (F1, blocking, in-diff) |
| Security (25%) | 95/100 | PASS | Durability relaxation is limited to `goos=="windows" && errors.Is(err, fs.ErrPermission)`, applied after a successful rename into the same directory; no input or secret surface |
| Craft (20%) | 80/100 | PASS | Mutants M1-M5 killed, M6 survives (F2, optional); full claim package `ok 149.610s`; vet and gofmt clean |
| Consistency (15%) | 90/100 | PASS | Same posture as `ledger/snapshot.go:55-58`, narrower (error kind checked as well) |

## Findings

- **F1** [High] [blocking] [in-diff] `internal/agenticfactory/claim/durable_dirsync_t2119_test.go:41`: `TestSyncDirDefaultPath` asserts `syncer{}.syncDir(t.TempDir()) == nil`. On Windows, `os.Open(dir)` succeeds (syscall opens directories with `FILE_FLAG_BACKUP_SEMANTICS`, `syscall_windows.go:418-446`), and then `d.Sync()` returns `ERROR_ACCESS_DENIED`. That is exactly the CI error the card fixes (`sync C:\…\docs\checkpoint: Access is denied.`, op `sync`). The test file has no build constraint, so it runs in the Windows race matrix and fails `t.Fatalf`. Independent reproduction: Codex P1 on the same line. **Required fix:** on the existing-directory leg, assert `!dirSyncFailureIsFatal(runtime.GOOS, err)` (or skip that one leg with a reason when `runtime.GOOS == "windows"`). Keep the missing-directory leg as is: `ERROR_FILE_NOT_FOUND` is still an error, and it is not `ErrPermission`.
- **F2** [Low] [optional] [in-diff] `internal/agenticfactory/claim/durable.go:72`: no test pins the `runtime.GOOS` argument at the call site. Mutant M6 (`dirSyncFailureIsFatal("windows", err)`) survives, because the existing wiring test injects `errors.New(...)` and not `fs.ErrPermission`. An optional extra case would inject `&fs.PathError{Err: fs.ErrPermission}` through `s.sync.dir` and expect the write to fail on non-Windows hosts. This is discretionary and does not affect the verdict.
- **F3** [Low] [optional] [in-diff] `.moai/reports/t2119/progress.md` Residual-risk names only the "different errno" risk. It misses the risk that the new test itself fails on Windows (F1). The record should be corrected together with the F1 fix.

Outside the diff (no verdict effect): no other directory-fsync site exists in `internal/agenticfactory`. `ledger/snapshot.go:56` is already tolerant on Windows, and `durable.go` is the only other site. The non-test `.Sync()` calls elsewhere in `internal/` (homestate/admission.go:99, cli/migrate_home_state.go:343 and others) sync files, not directories.

## Card-scope judgment (one slice plus a split proposal)

This is acceptable. The card text allows "필요하면 분할 제안" (propose a split where needed). The slice takes the largest deterministic cause from the code (326 failures), and the six proposed follow-ups each map to a row of the CI analysis. The t2118 SQLite row is correctly left alone. Issuing the follow-up cards is the leader's decision.

## Claim / Evidence / Baseline-attribution / Gaps / Residual-risk

### Claim
1. The tolerance is correct for Windows `ERROR_ACCESS_DENIED`, whether it comes from the open or from the sync step.
2. The 326 failures all route through `claim.writeAtomic`, and no other agenticfactory dir-fsync site is missed.
3. The new unit tests pass on darwin and kill the decision-function mutants.
4. The new test file fails on Windows (F1).

### Evidence
- `git rev-parse HEAD` → `703a5b40352cc9c4fa0d52e5ca4cca640da2f215`
- Go stdlib `syscall/syscall_windows.go:189-194` (go1.27.3):
  ```
  func (e Errno) Is(target error) bool {
  	switch target {
  	case oserror.ErrPermission:
  		return e == ERROR_ACCESS_DENIED ||
  ```
  `types_windows.go:13 ERROR_ACCESS_DENIED Errno = 5`. `os/file_posix.go:167-168` wraps the error as `f.wrapErr("sync", e)` → `*PathError`, which `errors.Is` unwraps. In `syncDir`, an `os.Open` failure (op `open`) flows through the same `err` and gets the same tolerance.
- `grep -rn -e '\.Sync()' -e 'syncDir' -e 'syncPath' -e 'SyncDir' internal/agenticfactory | grep -v _test.go`:
  ```
  internal/agenticfactory/ledger/snapshot.go:12:var syncPath = func(path string) error {
  internal/agenticfactory/ledger/snapshot.go:17:	if err := f.Sync(); err != nil {
  internal/agenticfactory/ledger/snapshot.go:49:	if err := syncPath(name); err != nil {
  internal/agenticfactory/ledger/snapshot.go:56:	if err := syncPath(dir); err != nil && runtime.GOOS != "windows" {
  internal/agenticfactory/claim/durable.go:24:	return f.Sync()
  internal/agenticfactory/claim/durable.go:27:func (y syncer) syncDir(path string) error {
  internal/agenticfactory/claim/durable.go:35:	err = d.Sync()
  internal/agenticfactory/claim/durable.go:72:	if err := s.sync.syncDir(dir); dirSyncFailureIsFatal(runtime.GOOS, err) {
  ```
  `writeAtomic(` callers: claim/presence.go:79, claim/docs.go:111, claim/claim.go:491.
- `go test ./internal/agenticfactory/claim -run 'TestDirSyncFailureIsFatal|TestSyncDirDefaultPath' -count=1 -v`:
  ```
  --- PASS: TestDirSyncFailureIsFatal (0.00s)
  --- PASS: TestSyncDirDefaultPath (0.01s)
  ok  	github.com/modu-ai/moai-adk/internal/agenticfactory/claim	0.235s
  ```
- `go test ./internal/agenticfactory/claim -count=1` → `ok  github.com/modu-ai/moai-adk/internal/agenticfactory/claim	149.610s` (exit 0)
- `GOOS=windows go vet ./internal/agenticfactory/claim` → no output, exit 0. `gofmt -l internal/agenticfactory/claim` → no output, exit 0.
- Mutation (`go test -overlay`, run filter = the 2 new tests + TestPutDocFailsWhenTheDirectorySyncFails + TestEveryRecordWriteSyncsFileThenDirectory):
  - M1 drop the windows condition → `FAIL` `linux access denied is fatal: … = false, want true` (killed)
  - M2 tolerance on linux/darwin as well → `FAIL` same lines (killed)
  - M3 drop the error-kind check → `FAIL` `windows other failures are still fatal: … = false, want true` (killed)
  - M4 never fatal → `FAIL` incl. `durable_test.go:67: a failed directory sync must fail the write: <nil>` (killed)
  - M5 wiring ignores the dir error → `FAIL` `durable_test.go:67` (killed)
  - M6 wiring hardcodes `"windows"` → `ok … 0.514s` (survived — F2)
- Codex (`mcp__moai__codex_audit`, native, target baseBranch, delta_base 04d8c27a7e → review_base `delta_base (merge base 04d8c27a7ed77afc09217f8623519806d48df3c4)`), verbatim: `"verdict":"fail"`, finding `[P1] 새 테스트에서도 Windows의 디렉터리 sync 거부를 허용하라 — internal/agenticfactory/claim/durable_dirsync_t2119_test.go:41`, `"audit_receipt":"rcpt-47cdae1a479a0b784de770af"`, `resolved_model":"gpt-6.1-sol"`, transport `codex app-server` (the MCP tool's own transport; no direct app-server route was used by this auditor).
- `moai verify audit-plan --project-root <tree>` → `config_status: ok`, model `codex`, codex gate `required`, claude/glm `off`, `enforced_required: ["codex"]`.

### Baseline-attribution
All commands ran in this run against tree t2119 at HEAD 703a5b4035, on darwin/arm64 with go1.27.3. The moai binary that served the audit plan and codex_audit was built from 04d8c27a7, the direct parent of HEAD (build_lag reported by the tool). The audited commit changes no moai audit-tool code, so the lag does not affect these two results. The tests, vet, gofmt and mutants used the toolchain directly on this tree.

### Gaps
- No Windows host. F1 rests on stdlib source reasoning plus the CI error text (op `sync`, Access is denied) and Codex's independent static finding. It was not executed on Windows.
- Codex reported that it ran no tests (static review).
- The `audit_multi` convergence check was not run (forbidden by the lead's spawn instructions). Codex was called once, directly.
- Whether the 26 epoch-0 refusals clear after this fix was not checked (progress item 5 leaves it to CI).
- The mutant overlay files live in session scratch, not in this report. Their verbatim outcomes are quoted above.

### Residual-risk
- Even after F1 is fixed, the Windows CI could surface a non-`ErrPermission` errno (for example, a sharing violation on some runners). The progress record already names this.
- On Windows, directory-entry durability after a rename now rests on NTFS metadata journaling and not on an explicit fsync. This is accepted by design, matching the ledger.

## Iteration history
| Iteration | Audited SHA | Verdict | Score | Codex receipt | Notes |
|---|---|---|---|---|---|
| 1 | 703a5b40352cc9c4fa0d52e5ca4cca640da2f215 | FAIL | 0.74 | rcpt-47cdae1a479a0b784de770af | blocking F1 (test requires nil from real dir fsync, fails on Windows); F2/F3 optional |
| 2 | b695b89b93c3881355d6cafef95cc6948c4434c3 | PASS | 0.91 | rcpt-8133aa42141ea9d1f940b19d | F1 and F3 resolved; F2 optional, open; Codex delta review pass, no findings |

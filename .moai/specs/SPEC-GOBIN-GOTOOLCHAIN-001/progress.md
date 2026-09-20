# Progress — SPEC-GOBIN-GOTOOLCHAIN-001

Card: t969. Tier S.

## §E.1 Plan-phase Audit-Ready Signal

- Artifacts: `spec.md`, `plan.md`, `acceptance.md` (Tier S set), plus this record.
- SPEC ID regex self-check: executed as Bash, output `PASS`.
- Plan authored at HEAD `b1ec8602c`, branch `WT-gotoolchain-local`, `develop...HEAD` = `0 0`.
- Every path cited in the artifacts was verified present at that HEAD before citation.
- No implementation file was touched during the plan phase.
- `plan_status: audit-ready`

## §E.2 Run-phase Evidence

Card t969. Branch `WT-gotoolchain-local`. Every figure below was measured in this run against
the tree named in its row. Figures attributed to card t964 are cited as prior measurements and
were NOT re-measured (R-4).

### Environmental precondition for RED (measured this run, tree `31c38fd0e`)

The defect needs the PATH `go` to be older than the module's `go` directive (plan.md R-1).

| Command | Observed output |
|---|---|
| `/opt/homebrew/bin/go version` (run from `/tmp`, outside the module) | `go version go1.26.0 darwin/arm64` |
| `head -3 go.mod` | `go 1.26.8` |
| `env \| grep -E '^(GOPATH\|GOMODCACHE)='` | (no output — both unset, so the child derives its module cache from `HOME`) |

Precondition holds: PATH `go` 1.26.0 < module directive 1.26.8. RED is producible on this
machine.

### AC-GGT-004 — RED-now cell (four elements)

- **Tree SHA**: `a0d853cc0` (M1 commit — test only, `resolver.go` still unfixed; `git status --short` empty at measurement time)
- **Command** (built at the same tree, then run directly — the precompiled form, R-1):

  ```
  go test ./internal/runtime/gobin -c -o <scratch>/gobin.test
  cd internal/runtime/gobin && <scratch>/gobin.test -test.v -test.run TestDetect_DoesNotDownloadToolchain
  ```

- **Verbatim stdout**:

  ```
  === RUN   TestDetect_DoesNotDownloadToolchain
      toolchain_test.go:50: gobin.Detect downloaded a Go toolchain into the test HOME: /var/folders/kt/nq2q81cn4gx3y41r7x47ggmr0000gn/T/TestDetect_DoesNotDownloadToolchain1615513091/001/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.26.8.darwin-arm64 holds 12 entries (glob /var/folders/kt/nq2q81cn4gx3y41r7x47ggmr0000gn/T/TestDetect_DoesNotDownloadToolchain1615513091/001/go/pkg/mod/golang.org/toolchain@*); the go env subprocesses must run with GOTOOLCHAIN=local
      testing.go:1464: TempDir RemoveAll cleanup: unlinkat /var/folders/kt/nq2q81cn4gx3y41r7x47ggmr0000gn/T/TestDetect_DoesNotDownloadToolchain1615513091/001/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.26.8.darwin-arm64/LICENSE: permission denied
  --- FAIL: TestDetect_DoesNotDownloadToolchain (7.14s)
  FAIL
  ```

- **Exit code**: `1`
- **R-2 (the failure is the assertion's own)**: the first failure line is `toolchain_test.go:50`,
  the `t.Errorf` of the AC-GGT-003 assertion. The build step exited 0 (`build_exit=0`), so this
  is not a compile failure. The `testing.go:1464` line is the downstream `t.TempDir` cleanup
  failure — the same read-only module cache card t964 recorded — not the catching assertion.
- Machine load at the RED measurement: `load averages: 13.23 14.49 12.59`,
  `CPU usage: 52.56% user, 16.29% sys, 31.13% idle`.

### AC-GGT-005 — GREEN pair (tree `35483be56`, M2 fix applied, working tree clean)

Binary rebuilt from the fixed tree before the run (`build_exit=0`).

- **Precompiled form** — `cd internal/runtime/gobin && <scratch>/gobin.test -test.v -test.run TestDetect_DoesNotDownloadToolchain`:

  ```
  === RUN   TestDetect_DoesNotDownloadToolchain
  --- PASS: TestDetect_DoesNotDownloadToolchain (0.03s)
  PASS
  ```

  Exit code `0`.

- **`go test` form** — `go test ./internal/runtime/gobin -count=1 -v`:

  ```
  === RUN   TestDetect_GOBINFirst
  --- PASS: TestDetect_GOBINFirst (0.01s)
  === RUN   TestDetect_GOPATHSecond
      resolver_test.go:43: GOPATH/bin 우선순위 검증: result=/custom/gopath/bin
  --- PASS: TestDetect_GOPATHSecond (0.01s)
  === RUN   TestDetect_HomeFallback
  --- PASS: TestDetect_HomeFallback (0.01s)
  === RUN   TestDetect_LastResort
      resolver_test.go:82: Last resort: 빈 문자열이 아닌 값 반환=/Users/goos/go/bin
  --- PASS: TestDetect_LastResort (0.01s)
  === RUN   TestDetect_DoesNotDownloadToolchain
  --- PASS: TestDetect_DoesNotDownloadToolchain (0.01s)
  PASS
  ok  	github.com/modu-ai/moai-adk/internal/runtime/gobin	0.349s
  ```

  Exit code `0`. The four pre-existing fallback-chain tests are unchanged and still pass, which
  is the behavioral evidence for REQ-GGT-004.

Card t964's control pair (precompiled `--- FAIL (8.07s)` / `go test` PASS — prior measurement,
cited not re-measured) is therefore turned into PASS / PASS.

Machine load at the GREEN measurements: `load averages: 17.01 15.32 13.02`,
`CPU usage: 27.78% user, 17.38% sys, 54.83% idle`.

### AC-GGT-001 / AC-GGT-002 — source-level (tree `35483be56`)

| Command | Observed output |
|---|---|
| `grep -c 'GOTOOLCHAIN=local' internal/runtime/gobin/resolver.go` | `2` |
| `grep -c 'os.Environ()' internal/runtime/gobin/resolver.go` | `2` |
| `git diff --stat` (M2, staged change) | `internal/runtime/gobin/resolver.go \| 3 +++` / `1 file changed, 3 insertions(+)` |

Both call sites read `cmd.Env = append(os.Environ(), "GOTOOLCHAIN=local")` — the parent
environment plus the single override, not a bare literal slice (Mutant A rejected). The pre-fix
count was `0`, measured at plan phase. The 3 insertions are the two `cmd.Env` lines plus the
`os` import.

### AC-GGT-006 — vet and lint (tree `35483be56`)

| Command | Verbatim output | Exit code |
|---|---|---|
| `go vet ./internal/runtime/gobin/...` | (no output) | `0` |
| `golangci-lint run ./internal/runtime/gobin/...` | `0 issues.` | `0` |
| `gofmt -l internal/runtime/gobin/` | (no lines listed — judged on the OUTPUT, not the exit code) | — |

### Supplementary (not an AC)

| Command | Exit code |
|---|---|
| `go build ./internal/runtime/gobin/...` | `0` |
| `GOOS=windows GOARCH=amd64 go build ./internal/runtime/gobin/...` | `0` |

### Residue cleanup (this run's own RED run)

The RED run left a read-only temp tree that nothing removes:

- Path: `/var/folders/kt/nq2q81cn4gx3y41r7x47ggmr0000gn/T/TestDetect_DoesNotDownloadToolchain1615513091`
- Size: `231M` (`du -sk` = `236508` KB)
- Removed with `chmod -R u+w <dir>` then `rm -rf <dir>`; `rm_exit=0`; a follow-up `ls -d` on the
  glob reports `no matches found`.
- The 27 `TestRunInit_*` / `TestInit*` / `TestRunWizardFn*` directories belonging to another card
  were NOT touched; a post-run count confirms `27` still present.

### Gaps

- The AC-GGT-003 assertion was exercised on darwin/arm64 only. Windows and Linux behavior of the
  same code path was not observed in this run; only a cross-compile build was.
- One Bash invocation combining staging, a heredoc and a commit was REFUSED by the
  worktree-isolation guard ("in a form too complex to verify"). It was re-issued as plain
  separate commands; nothing was substituted for a measurement, and no verification below rests
  on the refused form.

### Residual risk

- The fix pins `GOTOOLCHAIN=local` for two `go env` queries only. Should a future change make
  either call site depend on a module-specific toolchain, `local` would become the wrong value —
  the queries `GOBIN` and `GOPATH` are toolchain-independent, which is what makes `local` safe
  here and would not make it safe on an arbitrary `go` invocation.
- Every other `exec.Command` site in the repository retains the same `GOTOOLCHAIN` exposure.
  Out of scope by `spec.md` §4, deliberately not surveyed in this run.
- R2 (removing the subprocess entirely) remains an unapproved follow-up candidate; this run did
  not evaluate it.

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-09-20
run_commit_sha: 35483be56
m1_commit_sha: a0d853cc0
run_status: audit-ready
ac_pass_count: 6
ac_fail_count: 0
red_observed_via: precompiled-binary
red_tree_sha: a0d853cc0
green_tree_sha: 35483be56
preserve_list_post_run_count: 0
new_warnings_or_lints_introduced: 0
cross_platform_build:
  darwin_arm64: pass
  windows_amd64: pass
total_run_phase_files: 2
m1_to_mN_commit_strategy: "M1 test-only commit a0d853cc0 precedes M2 fix commit 35483be56; the commit graph witnesses the RED-before-GREEN ordering required by AC-GGT-004"
push_state: "not pushed - lane does not push; lead batch-pushes the integration branch"
```

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_

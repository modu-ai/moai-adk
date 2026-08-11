# SPEC-CONFIG-ATOMIC-WRITE-001 — Acceptance Criteria

> Given-When-Then scenarios for REQ-CAW-001..007. The verification layer for `spec.md` §D.
> Every AC MUST be runnable under `go test ./...` (not only package-local — per
> feedback-full-test-suite-verification). Every AC test uses `t.TempDir()`.

## §D AC Matrix

| AC | REQ | Severity | Traceability |
|----|-----|----------|--------------|
| AC-CAW-001 | REQ-CAW-001 | MUST | Atomicity — no partial write on crash-simulated rename failure |
| AC-CAW-002 | REQ-CAW-002 | MUST | Mode preservation round-trip (0600, 0644, 0750) |
| AC-CAW-003 | REQ-CAW-003 | MUST | Destination-not-present default mode (`defs.FilePerm`) |
| AC-CAW-004 | REQ-CAW-004 / REQ-CAW-004a | MUST | Call-site remediation — all enumerated writers route through the helper |
| AC-CAW-005 | REQ-CAW-005 | MUST | No numeric mode literals in helper or changed call sites |
| AC-CAW-006 | REQ-CAW-006 | MUST | Temp-file cleanup on both success and error paths |
| AC-CAW-007 | REQ-CAW-007 | MUST | Regression guard — bare `os.WriteFile` to config paths is rejected |

## §D.1 Severity

All seven AC are MUST-pass (MUST-FIX on failure). There are no SHOULD or NICE-TO-HAVE AC.

## §D.2 Given-When-Then Scenarios

### AC-CAW-001 — Atomicity (no partial write)

**Given** a destination file `dir/target.yaml` containing `prior: content`, and the atomic-write
helper invoked with `data = []byte("new: content\n")`,
**When** the rename step is allowed to complete normally,
**Then** `dir/target.yaml` reads exactly `new: content\n` and the prior content is gone.

**Given** a destination file `dir/target.yaml` containing `prior: content`, and a failing-write
scenario (e.g. helper invoked with a reader that errors mid-write, or rename forced to fail),
**When** the write fails before rename,
**Then** `dir/target.yaml` still reads exactly `prior: content` (unchanged) AND no temp file
matching `.moai-config-*.tmp` remains in `dir/`.

### AC-CAW-002 — Mode-preservation round-trip (table-driven)

**Given** a destination file created at mode `M` (for each `M` in `{0o600, 0o644, 0o750}`) via
`os.WriteFile(dir/target.yaml, []byte("old"), M)` inside a `t.TempDir()`,
**When** the atomic-write helper writes new content to `dir/target.yaml`,
**Then** `os.Stat(dir/target.yaml).Mode().Perm() == M` for every row in the table — the mode is
preserved, NOT narrowed to 0600, NOT widened to 0644.

The table MUST include at minimum:

| Input mode | Asserted result mode |
|------------|----------------------|
| `0o600` | `0o600` (preserved, not widened) |
| `0o644` | `0o644` (preserved, not narrowed to 0600 — the live Defect 1) |
| `0o750` | `0o750` (preserved — explicit non-default mode survives) |

### AC-CAW-003 — Destination-not-present default mode

**Given** a `t.TempDir()` where `dir/target.yaml` does NOT exist,
**When** the atomic-write helper is invoked with `defaultMode = defs.FilePerm` (or with no
default, in which case the helper applies `defs.FilePerm`),
**Then** `os.Stat(dir/target.yaml).Mode().Perm() == defs.FilePerm` (`0o644`) — the helper does NOT
create the file at `os.CreateTemp`'s 0600 default.

**Given** a `t.TempDir()` where `dir/secret.yaml` does NOT exist,
**When** the helper is invoked with `defaultMode = 0o600` (caller-declared secret file),
**Then** `os.Stat(dir/secret.yaml).Mode().Perm() == 0o600` — the caller-supplied default wins.

### AC-CAW-004 — Call-site remediation

**Given** HEAD after M2 + M3 land,
**When** `grep -n 'os\.WriteFile\|atomicWrite\|atomicfile\.Write'` is run against the enumerated call
sites (`internal/config/manager.go:420-438`, `internal/cli/harness.go:390`,
`internal/cli/update_cleanup.go:243,258,446`, `internal/cli/update_preserve_inventory.go:350`,
`internal/cli/update_wizard.go:182,222,277,309,346`, `internal/cli/update.go:1004`,
`internal/config/toolpolicy/tier_render.go:146`, `internal/config/toolpolicy/codegen.go:244`),
**Then** every enumerated site routes through the shared helper `atomicfile.Write` (no remaining
bare `os.WriteFile` to a config/settings persistence path on those lines), AND `go test ./...` is
green.

### AC-CAW-005 — No numeric mode literals in changed code

**Given** the diff between the M1-M3 changes and HEAD `ed70e4354`,
**When** the added/modified lines are scanned for numeric mode literals
(`grep -nE '0o[0-7]{3,4}|0[0-7]{3,4}'` on the changed files),
**Then** the only literal allowed inside the helper source is the one inside `internal/defs/perms.go`
(the named-constant SSOT); every other reference is by name (`defs.FilePerm`, or a `defaultMode`
parameter, or a value read via `os.Stat`). Existing call sites that already used `defs.FilePerm`
must not regress to a literal.

### AC-CAW-006 — Temp-file cleanup on success AND error

**Given** a `t.TempDir()` and a count of files matching `.moai-config-*.tmp` BEFORE the call (zero),
**When** the helper completes successfully,
**Then** the count of `.moai-config-*.tmp` files AFTER the call is still zero (the temp was
renamed, not left behind).

**Given** a `t.TempDir()` and a forced error after temp creation but before rename (e.g. inject a
write failure via a seam, or point the helper at an unwritable sub-step),
**When** the helper returns an error,
**Then** the count of `.moai-config-*.tmp` files AFTER the failed call is zero — the
`defer os.Remove(tmpName)` cleaned up the orphan.

### AC-CAW-007 — Regression guard (mutation-verified)

**Given** the guard test from M4,
**When** `go test ./internal/config/... ./internal/cli/... -run <GuardTestName>` runs against the
clean tree,
**Then** the guard PASSES (no violator found).

**Given** the guard test and a throwaway edit that reintroduces a bare
`os.WriteFile(".moai/config/sections/x.yaml", ...)` in a non-exempted file under
`internal/config/` or `internal/cli/`,
**When** the guard test runs,
**Then** the guard FAILS, naming the offending file and line. (Capture this RED output for E8;
then revert the throwaway edit so the guard is green again on the real tree.)

Exemptions enumerated by the guard MUST include at minimum:
- the `diskBackupWriteFile` seam declaration (`internal/cli/update_disk_backup.go:39`);
- probe/lock/log writes that do not target `.moai/config/**` or settings-persistence paths;
- the helper's own internal `os.CreateTemp` call (it writes a temp, not a config path);
- `internal/cli/update_recovery_manifest.go:78` — target path is a backup-directory artifact
  (`backupDir/<recoveryManifestFileName>`), predicate: "target path is under a backup directory,
  not `.moai/config/**` or a settings path";
- `internal/cli/update_namespace_protect.go:242` — target path is a backup-directory atomicity
  marker (`backupDir/.complete`), predicate: "target path is under a backup directory, not a
  config or settings path".

The guard resolves each `os.WriteFile` target path per the REQ-CAW-007 target-path resolution
algorithm (string-literal resolution → variable-to-assignment intra-package trace → match against
config-set `.moai/config/**` and settings-set `settings.json`/`settings.json.tmpl`). A path that
is neither config-set nor settings-set is exempt by predicate, not by human judgment.

**No migration AC.** This SPEC carries NO acceptance criterion for the mode-widening migration
(parent REQ-CTP-025/026). That migration belongs to follow-up sibling SPEC
`SPEC-CONFIG-MODE-MIGRATE-001`. The absence of a migration AC here is by design, not a gap: this
SPEC's scope is the write-path invariant only.

## §D.3 Indirect verification

- `go test ./...` (FULL suite, not package-local) is green — cross-package regressions caught.
- `GOOS=windows GOARCH=amd64 go build ./...` succeeds — cross-platform build intact.
- `golangci-lint run --timeout=2m` reports NEW=0 against the pre-flight baseline.

## §D.4 Closure gates

- All 7 MUST AC PASS.
- Coverage ≥ 85% on `internal/config` and on the touched `internal/cli` files.
- No `[NEEDS CLARIFICATION]` markers remain in `plan.md`.
- The guard (AC-CAW-007) is mutation-verified — both the green-on-clean and red-on-violation
  directions are demonstrated.

## §D.5 Forward-looking checks

- The helper API is documented (godoc) so sibling-slice-(c) SPECs can adopt it without redesign.
- The exemption list in the guard is explicit and commented, so a future contributor adding a
  legitimate non-config `os.WriteFile` knows how to exempt it without disabling the guard.

## §D.6 Definition of Done

This SPEC is DONE when:
- every MUST AC PASSes under `go test ./...`;
- the guard is mutation-verified;
- the helper is the single choke point for all enumerated config/settings writes;
- the parent `SPEC-CONFIG-TIER-PERSIST-001` `## §L Split Branches` section correctly records
  REQ-CTP-021/022/023/024/027 as satisfied by this SPEC's REQ-CAW-001..007, and REQ-CTP-025/026 as
  deferred to `SPEC-CONFIG-MODE-MIGRATE-001`.

# SPEC-PROGRESS-RECORD-IO-001 — Implementation Plan

## §A Context

- Base: **a2a184ad3** (= origin/main tip), branch `WT-progress-record-io`, card t1598.
- The F14 defect lives in one module surface: the darwin metadata seeder
  (`internal/runtime/progress_metadata_darwin.go`) shells out (`chmod -N` + `cp -p`) to seed
  mode/ACL/xattr, opening a name-swap→symlink→victim-overwrite window that the post-checks can
  only detect, not prevent.
- The fix must be Go-native and fd-anchored; the acceptance target is the held three-axis family
  executed with `PATH=""`. Everything else about the replace (fd holding, swap semantics, close
  hygiene) is landed on the base and is preserved, not redesigned.
- The darwin route is UNMEASURED — the plan's first milestone exists to measure it before any
  implementation decision is made.

## §B Known Issues

| ID | Status on base | Disposition in this SPEC |
|----|----------------|--------------------------|
| F14 darwin seeder exec window | OPEN (P1) | The core work item (M1 probe → M2 fix) |
| F15 mismatch-path removal omission | LANDED (`swapped` flag) | Verification-only (regression guard must survive) |
| F16 fd close hygiene | LANDED (defer close + close-before-rename) | Verification-only |
| Residual fd-verify→rename microsecond window | Structural (rename(2) is not fd-anchored) | Re-documented post-fix at the reduced harm class (REQ-DOC-009) |
| F9/F10/6b/6c/6d GOOS-tagged CI accumulation | Not yet observed | M4: accumulate on the card PR's CI runs |
| Ruling (i) follow-up note | Stale pre-fix comment | Re-documented (REQ-DOC-009) |

## §C Pre-flight

1. Base pinned: `git rev-parse --short HEAD` → `a2a184ad3`; branch `WT-progress-record-io`.
2. Held family present: `.moai/state/verify/t1598-prework/held-audit_ceiling_axes_test.go`.
3. Dependency present: `golang.org/x/sys` v0.48.0 in `go.mod` (the held family imports
   `golang.org/x/sys/unix`).
4. Darwin build environment available (the authoring platform is darwin; the seeder work is
   darwin-tagged).
5. t1560 evidence sources reachable: sync-audit-8 and the gate ledger (read-only citations).

## §D Constraints

- TDD red-first: the held family IS the red. No implementation may land before the family is
  observed RED in-package on the pre-fix tree.
- `-race -count=2` for the append family; affected package only (`go test -timeout 30m
  ./internal/runtime/`); NO local full-suite run (lane discipline).
- No cgo anywhere (`tech.md` no-CGo clause, `CGO_ENABLED=0`). The cgo ACL-API fork leg is closed.
- All darwin changes behind `//go:build darwin` so linux/windows CI stays green.
- No mode-only fallback; no re-scoping of the held family without an operator decision
  (decision-index Q2); no queue/factory state writes.

## §E Self-Verification

Verification is the acceptance layer: every AC in `acceptance.md` names its command and its
expected observable; CI-only criteria name the CI surface instead of a local command. The
run-phase record lands in `progress.md` §E.2 with verbatim outputs and tree SHAs.

## §F Milestones

Ordered by decision-reversibility: the route decision (M1) is the highest-change-likelihood
decision and comes first; mechanical steps trail.

### M1 — Darwin fd-xattr route probe (cycle_type: tdd, Priority High)

Evidence-first: measure before choosing. RED state at entry: the route is unmeasured (no probe
artifact exists).

1. **Positive control**: seed an ACL on a scratch file via
   `chmod +a "group:_guest deny read" <scratch>`; confirm visibility via `ls -le <scratch>`.
2. **Raw xattr surface**: enumerate the seeded file's xattrs
   (`xattr -l <scratch>` plus a raw listxattr sweep) to measure the exact name the kernel stores
   ACLs under (historically a kauth_filesec-class name; APFS may expose none). **A null result is
   a decisive measurement, not a probe failure** — it kills the pure-Go xattr route.
3. **Blob layout**: where an ACL xattr surface exists, capture the raw bytes of the seeded blob
   (header + entry array layout) — the analogue of `minimalAclBlob` on linux.
4. **Non-root fd-set writability**: attempt the fd-based write through `golang.org/x/sys/unix` fd
   xattr APIs. First measure which fd-xattr wrappers v0.48.0 actually exposes on darwin
   (`go doc golang.org/x/sys/unix` sweep); then attempt the write as non-root on the held
   descriptor's inode and verify via `ls -le`.
5. **Decision fork (explicit, evidence-recorded)**:
   - (i) pure-Go fd route measured writable non-root → adopt in M2;
   - (ii) not writable → STOP; escalate decision-index Q2 (default: keep the current seeder and
     re-document the residual; M2 re-scopes to documentation only). The cgo leg is excluded on
     policy — it is not a fork branch.
6. **Artifact**: `.moai/state/verify/t1598/probe-darwin-fd-xattr.md` — commands, verbatim
   outputs, tree SHA; the decisive lines are carried into `progress.md` §E.2.

Gate: the fork branch is selected WITH recorded evidence, or Q2 is escalated. No implementation
starts on an unmeasured route.

### M2 — RED then GREEN: Go-native darwin seeder (cycle_type: tdd, Priority High)

- **RED**: promote the held family into `internal/runtime` verbatim (darwin build tag). Run
  `go test -run TestAppendProgressRecordPreservesAllMetadataAxes ./internal/runtime/` on the
  PRE-fix tree; capture verbatim FAIL + exit code + tree SHA into `progress.md` §E.2. RED is
  red-by-construction (empty `PATH` defeats `exec.Command` lookup of `chmod`/`cp`); the
  observation confirms it for the right reason.
- **GREEN**: implement the seeder per M1's measured route — fd-anchored, Go-native, no exec.
  Preserve the strip-inherited-ACL-first ordering (the Go analogue of `chmod -N`, or the
  blob-overwrite posture mirroring the linux seeder), the abort/fail-closed contract (no
  mode-only fallback), and the `swapped`-flag semantics untouched.
- **Refactor**: remove the exec path and the pre/post name-based re-checks it required (the
  window they guarded is eliminated by construction); update the ruling (i) comment to the
  post-fix harm class.

Gate: the promoted family passes with `-race -count=2`; `grep -c "exec.Command"
internal/runtime/progress_metadata_darwin.go` returns 0.

### M3 — Regression sweep + disposition (cycle_type: tdd, Priority Medium)

1. `go test -race -count=2 -run '^TestAppendProgressRecord' ./internal/runtime/` → ok (whole
   append family: per-axis, umask, swap-keep, hardlink, symlink — the F15/F16 guarantees ride
   here).
2. `go test -timeout 30m ./internal/runtime/` → ok (affected package only).
3. Ruling (i) re-documentation: post-fix harm class + kauth_filesec follow-up disposition line;
   pre-fix exec-exception comment superseded (REQ-DOC-009).
4. Probe artifact + decisive lines present in `progress.md` §E.2 (AC-PRI-001).

### M4 — CI decisive-run accumulation (cycle_type: autofix, Priority Medium)

- The linux/windows GOOS-tagged seeder families take their decisive verdict from the card PR's
  CI runs (F9/F10/6b/6c/6d accumulation). Record run URLs + per-family verdicts in
  `progress.md` §E.2. No local darwin run is cited as their evidence (REQ-CI-008).
- No red of its own: this milestone watches the CI surface and records observed results; a red
  CI family on the changed surfaces loops back into M2/M3 repair rounds.

## §G Anti-Patterns

- No mode-only fallback (the two F8 faces: dropped ACL, silently rewritten write-restricted
  file).
- No explicit `chmod` after seeding on darwin — a chmod DELETES the file's ACL; the mode rides
  the seeding.
- Do not "simplify" by removing the `swapped`-flag removal semantics or the close-before-rename —
  F15/F16 are landed guarantees with promoted tests.
- Do not cite a local darwin run as evidence for GOOS-tagged linux/windows families.
- Do not widen the route work into a general kauth_filesec library for arbitrary files (Out of
  Scope; decision-index Q4).
- Do not re-open the measured-rejected kernel surfaces (`clonefileat` / `SYS_COPYFILE` /
  `setattrlist`) without new evidence.

## §H Cross-References

- Card t1598; t1560 sync-audit-8 + gate ledger (evidence sources, worktree t1560).
- Held family: `.moai/state/verify/t1598-prework/held-audit_ceiling_axes_test.go`.
- Linux precedent: `internal/runtime/progress_metadata_acl.go`; fd-unified replace:
  `internal/runtime/audit_ceiling.go:640-726`; fd/name match: `progress_fd_unix.go` /
  `progress_fd_other.go`.
- `.moai/project/tech.md` no-CGo clause; `decision-index.md` Q1-Q4 (route evidence, fallback,
  CI surface, disposition).

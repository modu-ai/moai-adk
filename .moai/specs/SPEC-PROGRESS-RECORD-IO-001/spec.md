---
id: SPEC-PROGRESS-RECORD-IO-001
title: "Progress record I/O semantics — Go-native darwin metadata seeding (F14 exec-window closure)"
version: "0.1.0"
status: draft
created: 2026-10-08
updated: 2026-10-08
author: lane-5 (MoAI factory)
priority: P1
phase: "v3.2.0"
module: "internal/runtime"
lifecycle: spec-anchored
tags: "progress-io, darwin, metadata, acl, xattr, fd-hygiene, security"
card: t1598
tier: M
---

# SPEC-PROGRESS-RECORD-IO-001 — progress record I/O semantics (darwin F14)

## HISTORY

- **2026-10-08** — authored at plan phase from card t1598 (leader-issued 2026-10-08) against
  absorbed base **a2a184ad3** (= origin/main tip; t1560 landed via squash PR #1808, commit
  7caaaf91f). Scope corrected by measurement against that base: **F15** (mismatch-path removal
  omission — `swapped` flag, `internal/runtime/audit_ceiling.go:674-696`) and **F16** (fd close
  hygiene — defer closes the held descriptor, `:676-684`; close-before-rename `:716-720`) are
  ALREADY LANDED and carry the promoted regression test
  `TestAppendProgressRecordSwapKeepsForeignFile` (`internal/runtime/audit_ceiling_replace_test.go:220`)
  — both are verification-only items in this SPEC. The open defect is **F14** (darwin path-based
  seeder exec window, P1). The held three-axis family
  (`.moai/state/verify/t1598-prework/held-audit_ceiling_axes_test.go`) is the RED acceptance
  target.

---

## §A Background (measured)

- The progress-record append (`appendProgressRecord`, `internal/runtime/audit_ceiling.go:588`)
  writes through an atomic same-directory replace. The temp descriptor is HELD end-to-end and
  every name-based step is verified against the held fd; the rename is the only step that creates
  the final name.
- **F14 (OPEN, P1)**: the darwin metadata seeder
  (`internal/runtime/progress_metadata_darwin.go`) seeds mode, ACL entries, and extended
  attributes by shelling out — `chmod -N tmpPath` + `cp -p original tmpPath` (`exec.Command`
  present twice) — guarded by pre/post Lstat-symlink + fd-match checks. Between the pre-check and
  `cp`'s internal open, a directory writer can swap the temp NAME for a symlink to a victim file;
  `cp -p` follows it and overwrites the victim with the caller's privileges. The post-checks
  abort the replace, but the damage is already done (t1560 sync-audit-8 F14; adversarial swap
  reproduction observed: error returned, external content replaced).
- **General contract** (card verbatim): 「호출부는 os.WriteFile의 플랫폼 가시 시맨틱(umask·ACL·xattr·경로·fd
  위생)을 보존한다」 — callers preserve `os.WriteFile`'s platform-visible semantics (umask, ACL,
  xattr, path, fd hygiene), completed on top of fd unification. The darwin seeder is the last
  surface that does not satisfy this contract Go-natively.
- **Rejected-by-measurement kernel surfaces** (t1560 round-4 closure): `clonefileat` copies
  xattrs but not ACLs; the `SYS_COPYFILE` trap returns EINVAL; the ACL bit is excluded from
  `ATTR_CMN_SETMASK`. These measurements stand; this SPEC does not re-open them.
- **Linux precedent (Go-native ACL)**: `internal/runtime/progress_metadata_acl.go` constructs a
  minimal POSIX ACL blob and writes it as the `system.posix_acl_access` xattr. The darwin
  analogue (kauth_filesec-class route) is UNMEASURED: the exact xattr name the kernel stores ACLs
  under, the blob layout, and non-root fd-set writability via `golang.org/x/sys/unix` fd xattr
  APIs must be measured by an M1 probe before the route decision.
- **Policy constraint**: `.moai/project/tech.md` — "No CGo dependencies: The binary must compile
  without CGo for maximum cross-platform portability (`CGO_ENABLED=0`)". A cgo
  `acl_get_fd`/`acl_set_fd` fallback is therefore excluded on policy, leaving the pure-Go fd
  xattr route as the only implementable branch.
- **Held family = the RED**: `.moai/state/verify/t1598-prework/held-audit_ceiling_axes_test.go`
  (darwin-only) asserts ONE family over THREE metadata axes on the replaced file — umask axis
  (mode 0600 under umask 0077), ACL axis (explicit `group:_guest deny read` survives), xattr axis
  (`user.t1560-axis=seeded` survives) — executed with `t.Setenv("PATH", "")`. The current
  exec-based seeder cannot pass it; the promoted per-axis tests
  (`audit_ceiling_acl_test.go`: PreservesACL / AclExactlyOriginal / PreservesXattr;
  `audit_ceiling_umask_test.go`: NewFileModeAppliesUmask / ExistingFileModePreserved) do NOT
  strip PATH and do NOT combine the three axes — the held family's coverage is unique.

## §B Requirements (GEARS)

- **REQ-PRI-001 (metadata preservation)**: The darwin progress-record metadata seeder shall
  preserve the original file's platform-visible metadata — permission mode, ACL entries, and
  extended attributes — onto the replaced file, completing the `os.WriteFile`-visible semantics
  contract on top of fd unification.
- **REQ-PRI-002 (no external process)**: While seeding metadata on darwin, the seeder shall not
  spawn any external process; every seeding operation shall be executable with `PATH` stripped to
  the empty string.
- **REQ-PRI-003 (fail closed)**: When a Go-native seeding step fails, the darwin seeder shall
  abort the replace, remove only the temp entry it still owns, leave the original untouched, and
  leave any swapped-in foreign entry untouched — there shall be no mode-only fallback.
- **REQ-PRI-004 (fd anchoring)**: The darwin seeder shall perform its metadata reads and writes
  against the inode the held temp descriptor references, so a directory-entry name swap cannot
  redirect the seeding onto a different inode.
- **REQ-PRI-005 (route selection)**: Where the pure-Go fd xattr route is measured writable as
  non-root by the M1 probe, the implementation shall use it; the cgo ACL-API route is excluded by
  the project's no-CGo constitution clause (`tech.md`, "No CGo dependencies").
- **REQ-PRI-006 (probe-gated fallback)**: When the M1 probe measures no writable pure-Go route,
  the route decision shall escalate to the operator per decision-index Q2 (default: keep the
  current seeder and re-document the residual; the held family shall not be silently re-scoped).
- **REQ-PRI-007 (held-family promotion)**: The held three-axis family (umask / ACL / xattr under
  `PATH=""`) shall be promoted into the package test suite as a darwin-tagged regression guard,
  observed RED before the fix, and shall pass with `-race -count=2` after it.
- **REQ-CI-008 (CI decisive verdict)**: The linux/windows GOOS-tagged seeder test families shall
  take their decisive verdict from the `release-pr-multi-os.yml` 3-OS leg (a `release/*`→`main`
  PR, or a `workflow_dispatch` run on the card branch); a local darwin run shall not be cited as
  their evidence.
- **REQ-DOC-009 (disposition re-documentation)**: When the F14 fix lands, the seeder's ruling (i)
  note shall be re-documented to the post-fix harm class — the residual fd-verify→rename
  microsecond window is structural (rename(2) is not fd-anchored) and the harm class drops from
  victim-overwrite to own-append-fails / foreign-temp-entry-replaced — and the pre-fix
  exec-exception comment shall be superseded, recording the kauth_filesec follow-up disposition.

## Out of Scope

### Out of Scope — linux/windows seeder behavior
- The linux and windows seeders are already Go-native and landed; this SPEC changes no
  linux/windows behavior. Their GOOS-tagged tests appear only as CI evidence obligations
  (REQ-CI-008), not as work items.

### Out of Scope — generalizing the darwin metadata route beyond progress-record I/O
- A general kauth_filesec reimplementation for arbitrary files is a follow-up candidate
  (decision-index Q4); this SPEC pins only the progress-record seeder and re-documents the
  disposition.

### Out of Scope — private-directory staging redesign
- Direction (b) from the t1560 audit (private-dir pre-staging, attacker-unwritable) still shells
  out and therefore violates the held family's `PATH=""` constraint; it is viable only if the
  family were re-scoped with justification — an operator decision (decision-index Q2), not a
  design branch this SPEC implements on its own authority.

### Out of Scope — queue and factory state
- No queue, factory, or card-state mutation; stage transitions belong to the lane.

## §C Acceptance

Acceptance criteria with commands and observables: `acceptance.md` (Tier M). Every criterion
names its evidence surface — local command or CI run — and no criterion is satisfied by
inspection alone.

## §D Constraints

- Methodology TDD (red-first; the held family IS the red). `-race -count=2` for the append
  family; affected package only (`-timeout 30m`) per local Go test rules — no local full-suite
  run.
- No cgo (`tech.md` no-CGo clause). Darwin changes stay behind build tags so linux/windows CI
  remains green.
- No mode-only fallback (F8 posture stands). The swap-semantics contract (F15, `swapped` flag)
  and fd close hygiene (F16) are landed guarantees — preserved, not redesigned.

## §E References

- Card t1598 (leader-issued 2026-10-08); t1560 sync-audit-8 + gate ledger
  (`.moai/worktrees/t1560/.moai/reports/t1560/sync-audit-8.md` — this path lives in the t1560
  card worktree of the PRIMARY checkout and does not resolve inside this tree) — evidence
  sources.
- Held family: `.moai/state/verify/t1598-prework/held-audit_ceiling_axes_test.go`.
- Linux ACL precedent: `internal/runtime/progress_metadata_acl.go`; fd-unified replace:
  `internal/runtime/audit_ceiling.go:640-726`.
- `.moai/project/tech.md` no-CGo clause; `decision-index.md` (this SPEC).

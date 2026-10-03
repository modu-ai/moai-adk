---
id: SPEC-FACTORY-MANAGED-TUI-001
title: "plan.md — implementation plan"
version: "0.1.0"
created: 2026-10-03
updated: 2026-10-03
author: GOOS (manager-spec)
tier: M
---

# plan.md — implementation plan

## §A. Context and tier

- **Tree**: branch `WT-managed-codex-tui-attach` @ `2b9e4a4d0`, worktree `.claude/worktrees/t1408`, clean at plan start. Card t1408; every commit carries the card id.
- **Artifacts**: `spec.md` (REQ 14), `acceptance.md` (AC 15), `design.md` (D-1..D-10), `plan.md`, `progress.md`.
- **Tier M.** Estimate (assumption-based, not measured): production ~350 lines (new `managed_codex_tui.go` ~220, edits ~130 in the two existing files and two config files), tests ~700. About 11 files affected (5-15 is Tier M). The LOC total sits at the Tier M ceiling; no constitutional change and no research.md need (the codebase research is the premise ledger in `spec.md` §A.1 and `design.md`). If the plan auditor reads it as Tier L, add research.md; nothing else changes.
- **PRESERVE**: `internal/factorymsg/store.go`; both parent SPEC directories; `factoryMoAIMCPApprovalArgs` and the App Server command line; the HARDEN-001 answer table; `codex_launcher.go` (card t1440's); `launch_exec_*.go`; the Claude/GLM stream owners.
- **Ownership boundaries (leader constraint)**: edits only in `managed_codex_factory.go`, `managed_factory_session.go`, new `managed_codex_tui.go` / `managed_codex_tui_test.go`, `internal/config/defaults.go`, `internal/config/envkeys.go`, the operator doc (one minimal separate bullet edit), CHANGELOG. Do not create `managed_operator_input.go` or `managed_card_child_test.go`.

## §B. Pre-flight (re-run before the run phase)

```bash
git branch --show-current && git rev-parse --short HEAD
go build ./...
GOOS=windows GOARCH=amd64 go build ./...
codex --version
codex resume --help
```

Measured in this plan run (tree `2b9e4a4d0`): `codex --version` → `codex-cli 0.160.0`; `codex resume --help` lists `--remote` and `--remote-auth-token-env` (0.040 s); the evidence ledger of `acceptance.md` §1.2 holds the grep baselines; managed top-level test names by `grep -h '^func Test' internal/cli/*_test.go` filtered on `Managed|managed`: 79 (a proxy for `go test -list`; the run phase re-measures with `-list`).

## §C. Milestones (ordered by how likely the decision is to change, after the fixed M1)

- **M1 — RED baseline (fixed first by REQ-MT-014).** Write the fake codex helper (`managed_codex_tui_test.go`: App Server role, TUI role, shared append log) and every reproduction test of AC-MT-001..011; observe each RED on the unchanged production code; land the tests and `red-baseline.md` in one commit before any production change. Adopts each AC's RED-now cell.
- **M2 — Interaction semantics (highest change likelihood).** The attach and the operator-visible flow: capability probe and preconditions, fallback notices, opt-out, terminal ownership (no stdin reader), log diversion to the session file, the optional capability interface on the driver, constants. Files: `managed_codex_tui.go`, the pre-`Start` decision in `runManagedFactoryCodex`, the one `cmd.Stderr` line in `Start`, the driver's optional-interface block and the `Factory inbox:` line, `defaults.go`, `envkeys.go`. Gate: AC-MT-001..005, 010.
- **M3 — Shared-thread concurrency.** Busy set and stale ceiling, reader no longer feeds the event channel with frames of turns the owner did not start, owned-turn set and request scoping. Files: `managed_codex_factory.go` reader and window helpers. Gate: AC-MT-006, 007; run with `-race`.
- **M4 — Lifecycle.** TUI wait goroutine, exit status via `exitCodeError`, server-death monitor, `stopTUI` as the first step of `Close`, TUI-exit release of a blocked `DeliverTurn`. Gate: AC-MT-008, 009.
- **M5 — Integration and regression.** Loopback round trip (AC-MT-011), the managed regression set, `GOOS=windows` cross build, zero-`syscall` grep, `gofmt`/`vet`/lint. Gate: AC-MT-010, 012.
- **M6 — Disclosure and hand-off.** Operator document (one minimal separate bullet edit in "알려진 한계", plus the opt-out, log path, `/exit` `/quit` and probe notes), CHANGELOG (B12 discipline), the manual check procedure file for AC-MT-015, sync hand-off. Gate: AC-MT-013, 014.

Priorities: M1 High (fixed), M2 High, M3 High, M4 High, M5 Medium, M6 Medium. Commit subjects `feat(SPEC-FACTORY-MANAGED-TUI-001): M<n> … (card t1408)`.

## §D. Risks

1. **P6/P7 are unobserved**, so M3's busy tracking and request scoping rest on a fake server's behavior. Mitigation: the design holds under either routing (D-5), degrades without deadlock (D-7), and AC-MT-015 is the observation.
2. **Merge conflict with t1459 and t1410** at three lines (struct field, `cmd.Stderr` line, `Close` first step). Mitigation: D-9; the TUI code lives in one file.
3. **Shipping default-on inside the managed gate without a real-TUI observation.** Mitigation: the opt-out (open question 2).
4. **Terminal corruption** from a stray writer. Mitigation: AC-MT-005 and the two grep baselines.
5. **Test fixture portability.** The re-exec shim is POSIX-only like the existing managed fixtures; Windows is proved by cross build.

## §E. Anti-patterns

- Closing the loop with a fake that mirrors the implementation's assumptions about routing; the fake must support both routings (AC-MT-007 has both).
- Adding `syscall.` anywhere in `managed_*` including tests.
- Calling `Close` from a second goroutine (unsafe until t1459).
- Reading the TUI exit status when the owner asked it to stop.
- Editing `codex_launcher.go` (card t1440's) or creating t1440's two new file names.

## §F. Self-verification deliverables (run phase)

E1 AC matrix with command and verbatim output for each of AC-MT-001..014 (015 is manual, labeled); E2 `go build ./...` and the `GOOS=windows` cross build exit codes; E3 coverage of the new file; E4 the zero-`syscall` and subagent-boundary greps; E5 lint new-versus-baseline; E6 commit SHAs and the local merge SHA; E7 blockers; E8 verbatim RED output before GREEN.

## §G. Cross-references

`spec.md`, `acceptance.md`, `design.md` D-1..D-10; SPEC-FACTORY-MANAGED-HARDEN-001 (landed, `861f416ae`); t1459 draft SPEC-FACTORY-MANAGED-SIGNAL-001 on `WT-managed-codex-session-f5` @ `2e41b007c` (read-only reference, not in this tree).

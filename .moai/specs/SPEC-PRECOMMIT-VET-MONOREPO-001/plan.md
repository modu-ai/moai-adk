# SPEC-PRECOMMIT-VET-MONOREPO-001 — Implementation Plan

> Tier M. The approach is settled (adopt reference patch b6f478b1a's logic, re-authored in place — no OPEN design decision); the risk concentrates in the twin-edit discipline (one bug-shaped hazard: editing one surface without the other) and in capturing honest RED evidence for a defect that cannot be reproduced against this repository's own layout. Milestones follow the card dispatch: M1 twin edit, M2 contrast tests, M3 verification sweep. Initial SPEC status: draft (spec.md frontmatter).

## §A Context

- **Tree**: worktree `.claude/worktrees/t237`, branch `WT-precommit-vet-module`, plan-phase baseline `b9298de32`.
- **SPEC artifacts**: `.moai/specs/SPEC-PRECOMMIT-VET-MONOREPO-001/{spec,plan,acceptance,progress}.md`
- **Module**: `internal/cli` (the constant) + `internal/template` (the twin file).
- **The two code surfaces (the twin pair, both named in every relevant AC)**:
  1. `internal/cli/hook_install_precommit.go` — Go constant `preCommitHookContent`, vet block at lines 78-103 of the constant.
  2. `internal/template/templates/.git_hooks/pre-commit` — the deployed template, vet block at lines 35-59.
  `TestPreCommitTemplateMatchesConstant` (`internal/cli/hook_install_precommit_test.go:38`) enforces byte identity; editing one without the other is RED by construction.
- **Test harness already in place** (all reused, none modified): `gitInitRepo` (test :373), `stageFile` (:386), `runPreCommitHook` (:402), the shim-PATH technique of `TestPreCommitHook_ToolchainAbsent` (:531), and the guard tests `TestPreCommitHook_GofmtBlocks` (:428), `TestPreCommitHook_SkipBypass` (:446), `TestPreCommitHook_NoStagedGo` (:469), `TestPreCommitHook_GoVetBlocks` (:504).
- **Reference patch**: `t312-precommit-vet` @ `b6f478b1a` — read via `git show`; judged sound (module-root walk, MODROOTS grouping, per-module subshell vet, build-tags-before-cd, module-root naming in the failure message). Adopted by re-authoring, not cherry-picking.

## §B Known Issues (relevant subset)

- **B4/B6 (frontmatter + lint)**: spec.md carries the canonical 12 fields; Out of Scope uses `### §E.N` h3 sub-sections (a bare `## Out of Scope` h2 triggers `MissingExclusions`).
- **B5 (CI 3-tier)**: `internal/cli` is a critical package (90%+ coverage target). The fix adds shell logic inside a Go string constant — Go coverage of the constant itself is unchanged in nature; the new coverage lives in the contrast tests. Measure the lint baseline at M1 pre-flight and report the delta.
- **B10 (PRESERVE)**: `.moai/state/`, `.moai/config/`, `.claude/hooks/moai/*.sh` wrappers are runtime-managed — untouched. The `.sh`/`.sh.tmpl` hook-wrapper pair discipline does NOT apply here: `.git_hooks/pre-commit` is a plain template file with no `.tmpl` sibling (verified: the template twin has no `.tmpl` variant; it is deployed verbatim by the installer's byte-identity contract).
- **Twin-edit hazard (this SPEC's bug-shaped risk)**: any edit to the constant that is not mirrored byte-for-byte into the template file fails `TestPreCommitTemplateMatchesConstant`. Mitigation: author the shell block once, paste into BOTH surfaces, run the identity test before anything else.
- **Windows**: the contrast tests skip on Windows (`runtime.GOOS == "windows"` → `t.Skip`, POSIX-only shim construction), matching `TestPreCommitHook_ToolchainAbsent`'s precedent. CI's windows leg sees them as skips, not failures.

## §C Pre-flight (run before M1)

```bash
git branch --show-current && git rev-parse --short HEAD
go build ./... && GOOS=windows GOARCH=amd64 go build ./...
go test -count=1 ./internal/cli/ -run 'TestPreCommit'   # baseline: all green pre-change
golangci-lint run --timeout=2m ./internal/cli/... 2>&1 | tail -5
```

Record baseline outputs in progress.md §E.2 before the first change commit.

## §D Constraints

- **PRESERVE**: gofmt block, bypass, toolchain-absent skip, no-staged-Go no-op, repo-root vet behavior (`TestPreCommitHook_GoVetBlocks` unchanged), heavy-gate block, installer machinery — all byte-unchanged outside the vet block (REQ-PVM-009).
- **TWIN DISCIPLINE [HARD]**: both surfaces edited in the SAME commit; `TestPreCommitTemplateMatchesConstant` green on every run-phase commit (REQ-PVM-006).
- **FORBIDDEN**: blind `git cherry-pick b6f478b1a` (the patch carries card-t312-specific provenance in comments; re-author with generalized comments); quoting the `$MODROOTS`/`$PKGS` expansions (out of scope — §E.1 of spec.md records why); touching `installPreCommitHookOptional`, the provenance/backup machinery, or any `.claude/hooks/moai/` wrapper.
- **Timeout floor**: every `go test ./internal/cli/...` invocation runs with a timeout of at least 600s and its exit code observed directly (unpiped — no `| tee`/`| head` between the test binary and the rc capture). The internal/cli package budget requires the 600s floor; a piped rc has masked failures before.
- **No card id in commits authored here**: plan-phase artifacts are committed by the run-phase owner; nothing this phase writes carries the card id.

## §E Self-Verification

E1 AC matrix (12 rows, acceptance.md) · E2 cross-platform build (`GOOS=windows GOARCH=amd64 go build ./...`) · E3 coverage for `internal/cli` vs the 90% critical target · E4 n/a (no subagent-boundary surface touched) · E5 lint delta vs the M1 pre-flight baseline · E6 branch HEAD + commit list (no push obligation — develop integration is the lane/lead's window, not this SPEC's run phase) · E7 blockers · E8 RED evidence — both contrast tests' failing command + verbatim RED output + exit code + fixture description + tree SHA, captured against the PRE-EDIT constant pair BEFORE the twin edit lands (the adoption gate per verification-completeness §2; see M2's RED-first clause for the exact mechanism).

## §F Milestones

### M1 — Twin edit: constant + template, byte-identical (Priority High)

Opening act (RED-first, before any edit lands): author the two contrast tests of M2 in the working tree and run them against the UNCHANGED constant pair, capturing verbatim RED output + exit codes + tree SHA into progress.md §E.2. The two REDs differ in kind and both must be recorded with their reason:
- `TestPreCommitHook_SubmodulePassesClean` RED via exit code (got 1, expected 0 — the defect itself: a clean file blocked).
- `TestPreCommitHook_SubmoduleVetBlocks` RED via the stderr assertion (exit 1 occurs on old too, for the wrong reason — module-resolution failure; the missing `module root: submod` line is the discriminator).

Then re-author the vet block after the reference patch's logic, in BOTH surfaces in one commit: `_moai_module_root()` upward walk, MODROOTS grouping, per-module `( cd "$_mr" && go vet $BT_TAGS $PKGS )` subshell, build-tags read hoisted before any `cd`, failure message + hint naming the module root. Comments generalized (no card-t312 provenance). Immediately verify `TestPreCommitTemplateMatchesConstant` green, then re-run both contrast tests → GREEN (M2 records the two-cell flip).

### M2 — Contrast tests completed + two-cell adoption record (Priority High)

Finalize the test code per the reference patch: `moduleShimPath` symlinks git/grep/sort/dirname/sed/awk/tr/go/gofmt but NOT moai (deterministic heavy-gate skip); `writeSubmoduleFixture` creates the repo whose only go.mod lives in `submod/`; `vetCleanGo` / `vetBadGo` constants (gofmt-clean in both cases; the bad one carries a Printf verb/arg mismatch); Windows skips. Assertions: clean fixture → exit 0 and no `FAILED` in stderr; bad fixture → exit 1 and `module root: submod` in stderr (the vacuous-green guard — without the naming assertion the bad-fixture test passes on the OLD constant for the wrong reason). Deliverable: the two-cell adoption record in progress.md §E.2 — RED cells (from M1's opening act) + GREEN cells (command, verbatim output, exit code, tree SHA of the post-edit run).

### M3 — Verification sweep (Priority Medium; mechanical)

The §E self-verification batch as one turn's parallel read-only commands, findings landing in progress.md §E.3: full `go test -count=1 ./internal/cli/...` (timeout ≥ 600s, rc unpiped), the twin-identity and guard-test targeted runs, `GOOS=windows GOARCH=amd64 go build ./...`, coverage for `internal/cli`, lint delta vs the M1 baseline, and the build-tags-before-cd ordering check (AC-PVM-005). The closing report carries the release-coordination line for the lead (AC-PVM-012): t230's landing precondition is satisfied; the remaining "one release must pass after t230's landing" is deployment-time, owned by release card t204.

## §G Anti-Patterns (this SPEC specifically)

- Do NOT run `go vet` once from the repo root with `-C <module-root>` style flags instead of `cd`-ing into the module — the reference patch's subshell `cd` is the adopted shape; flag-based variants change toolchain-version assumptions for no gain.
- Do NOT let the per-module loop's directory change leak past its invocation (REQ-PVM-005) — the heavy gate at the script tail must still run from the repo root.
- Do NOT move the build-tags read inside the per-module loop or after the first `cd` — `.moai/config/build-tags` is a repo-root-relative path (REQ-PVM-004).
- Do NOT "improve" the word-splitting situation by quoting `$MODROOTS`/`$PKGS` — out of scope (spec.md §E.1); an untested quoting change inside a byte-identical twin pair is exactly how regressions land.
- Do NOT capture the RED cells after the twin edit has landed — a RED observed on an already-fixed tree is either synthetic or absent, and the adoption gate (verification-completeness §2) rejects it.

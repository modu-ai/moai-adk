---
id: SPEC-PRECOMMIT-VET-MONOREPO-001
title: "pre-commit go vet — run from each staged file's module root so monorepos whose only go.mod lives in a subdirectory are not blanket-blocked"
version: "0.1.0"
status: completed
created: 2026-09-04
updated: 2026-09-04
author: manager-spec
priority: P1
phase: "v3.2.0 target"
module: internal/cli
lifecycle: spec-anchored
era: V3R6
tier: M
tags: "cli, pre-commit, go-vet, monorepo, git-hooks, template-twin"
related_specs: [SPEC-PRECOMMIT-GATE-SCOPE-001, SPEC-PRETOOL-GATE-MOVE-001, SPEC-UPDATE-HOOK-DELIVERY-001]
---

# SPEC-PRECOMMIT-VET-MONOREPO-001

## §A Problem / Motivation

The pre-commit hook's fast-subset `go vet` step collects staged `.go` files into package paths of the form `./$(dirname "$f")` and runs ONE `go vet` invocation from the hook's working directory — the repository root. In a monorepo whose only `go.mod` lives in a subdirectory (e.g. `apps/id/`), `go vet ./apps/id` executed from the root fails with `go: cannot find main module`, so **every commit that stages a `.go` file is blocked regardless of code quality**, and the hook's failure message misreports the module-layout error as a vet finding.

### Root cause (verified on tree b9298de32, branch WT-precommit-vet-module)

The defect is a single-cwd assumption inside the hook's vet block, present identically on both surfaces of the twin pair:

1. **Package paths are computed relative to the repo root** — `internal/cli/hook_install_precommit.go:80-85` (the `PKGS` assignment inside `preCommitHookContent`): `printf './%s\n' "$(dirname "$f")"`. Template twin: `internal/template/templates/.git_hooks/pre-commit:37-42`, identical.
2. **One `go vet` runs from the hook's own cwd** — `internal/cli/hook_install_precommit.go:96`: `go vet $BT_TAGS $PKGS` with no directory change, so the invocation inherits the repo-root cwd. Template twin: `.git_hooks/pre-commit:53`. Where the repo root holds no `go.mod`, the invocation cannot resolve a module and exits non-zero for reasons unrelated to the staged code.
3. **The failure message attributes the block to vet findings** — `internal/cli/hook_install_precommit.go:97-99` prints `FAILED: go vet reported issues in the staged packages`, which is false in the monorepo case (the actual error is module resolution).

### Local reproduction is impossible against this repository

This repository's `go.mod` sits at the repo root, so the defect cannot be reproduced by committing to this tree. The RED evidence therefore comes from **contrast tests** that build a fixture module tree inside `t.TempDir()` — a repo whose ONLY `go.mod` lives in `submod/` — and run the installed hook through the existing `runPreCommitHook` harness (`internal/cli/hook_install_precommit_test.go:402`). The two tests are RED against the current constant and GREEN after the fix (see §C.3).

### Reference patch — read and judged

Branch `t312-precommit-vet` @ `b6f478b1a` (read via `git show b6f478b1a`) implements the fix: a `_moai_module_root()` helper walks upward from each staged file's directory to the nearest `go.mod`; staged files are grouped by module root (`MODROOTS`); each module root gets its own `( cd "$_mr" && go vet $BT_TAGS $PKGS )` subshell run with module-relative package paths; the build-tags read moved BEFORE any `cd` so `.moai/config/build-tags` is still read from the repo root; the failure message names the module root and the hint includes the `cd`. Function-local variables are isolated by command substitution, so no global-state leak crosses blocks. **Lane judgment: adopt the patch's logic; do not blindly cherry-pick — re-author in place under this SPEC, with the twin-edit discipline of §C.2.**

## §B History

| Date | Version | Change |
|------|---------|--------|
| 2026-09-04 | 0.1.0 | Initial plan-phase artifacts (Tier M, card t237). Root cause verified against b9298de32. Reference patch b6f478b1a read and judged; adoption-by-re-authoring recorded in §A. |

## §C Requirements (GEARS)

### §C.1 Core fix requirements

**REQ-PVM-001** — **When** the pre-commit hook's fast subset runs with staged `.go` files present and `go` on PATH, the system shall determine each staged file's nearest enclosing module root (walking upward from the file's directory to the nearest directory containing `go.mod`) and shall execute `go vet` once per distinct module root, from within that module root, over the module-relative package paths of the staged files it owns.

**REQ-PVM-002** — **When** a staged `.go` file has no findable module root (no `go.mod` at or above its directory up to the repo root), the system shall treat the repo root (`.`) as that file's module root. The worst case is therefore the previous behavior — never less vet coverage than before the fix.

**REQ-PVM-003** — **When** a per-module `go vet` invocation fails, the hook's failure message shall name the module root the vet ran from (form: `module root: <path>`), and the remediation hint shall carry the reproduction shape `cd <module-root> && go vet <packages>` so the user can reproduce the failure in their own shell.

**REQ-PVM-004** — The optional build-tags read (`.moai/config/build-tags`, first non-comment non-blank line) shall execute from the repo root BEFORE any directory change into a module root, so the tags file remains resolvable in the monorepo layout and the resolved `BT_TAGS` value is shared by every per-module invocation.

**REQ-PVM-005** — Directory changes performed for per-module vet execution shall be isolated inside a subshell scoped to the invocation, so the hook's subsequent blocks (the opt-in heavy gate at the tail of the script) still execute from the repository root.

### §C.2 Twin-edit discipline

**REQ-PVM-006** — The Go constant `preCommitHookContent` (`internal/cli/hook_install_precommit.go`) and the template file `internal/template/templates/.git_hooks/pre-commit` SHALL remain byte-identical after the change. Both surfaces are edited in lockstep in the SAME commit; `TestPreCommitTemplateMatchesConstant` (`internal/cli/hook_install_precommit_test.go:38`) enforces the identity mechanically, and editing one surface without the other is RED by construction. Every requirement in §C.1 names both surfaces as its delivery vehicle.

### §C.3 Contrast-test requirements (the RED mechanism)

**REQ-PVM-007** — **When** the installed hook runs against a fixture repository whose only `go.mod` lives in `submod/` and the staged `.go` file under `submod/` is gofmt-clean and vet-clean, the hook shall exit 0 (clean pass). Against the current constant this scenario exits 1 — this RED is the fixture-based reproduction of the reported defect.

**REQ-PVM-008** — **When** the same fixture stages a gofmt-clean `.go` file carrying a real vet diagnostic (Printf verb/arg mismatch), the hook shall exit 1 AND the stderr shall contain `module root: submod`. This is the vacuous-green guard: it proves REQ-PVM-007's pass comes from vet actually running inside the module, not from the check being skipped. Note for the RED cell: the OLD constant also exits 1 on this fixture (for the wrong reason — module-resolution failure misreported as a vet finding), so the exit-code assertion alone is green-on-old; the stderr module-root naming is the discriminating assertion that makes this test a true contrast.

### §C.4 Preserved behaviors (characterization guards)

**REQ-PVM-009** — The hook's adjacent behaviors shall be byte-for-byte behaviorally unchanged by this fix: the gofmt format check (block preceding vet), the `SKIP_MOAI_PRECOMMIT=1` bypass, the toolchain-absent skip (neither `go` nor `gofmt` on PATH → exit 0 — the highest-blast-radius guarantee for the 16-language template audience), the no-staged-Go no-op, and the repo-root vet block for repositories whose `go.mod` IS at the root (`TestPreCommitHook_GoVetBlocks`, `internal/cli/hook_install_precommit_test.go:504`, must keep passing unchanged).

## §D Success Criteria

- All 9 requirements pass; the twin pair is byte-identical on every commit of the run phase (REQ-PVM-006 held continuously, not just at the end).
- Both contrast tests exist and are adopted per the two-cell discipline: verbatim RED output + exit code + fixture + tree SHA captured against the PRE-EDIT constant before the twin edit lands, then GREEN after (verification-completeness §2).
- The 5 preserved-behavior guard tests (§C.4) stay green throughout.
- `go test ./internal/cli/...` exit 0 on the run-phase tree (timeout floor 600s per the internal/cli budget; exit code observed unpiped).

## §E Out of Scope

### §E.1 Out of Scope — space-in-path word-splitting (documented limitation)

- The per-module grouping inherits the pre-existing unquoted `$MODROOTS` / `$PKGS` word-splitting, which breaks module or package paths containing spaces — pre-existing behavior (the old single-invocation `$PKGS` had the same property), not a regression introduced by this fix. Recorded as a documented limitation; this SPEC does not fix it. A future quoting pass would touch the same twin pair and deserves its own card.

### §E.2 Out of Scope — release gating

- The constraint "at least one release must pass after t230's landing before this ships" is a DEPLOYMENT-TIME concern owned by the release card (t204); it does not block implementation or the run/sync phases of this SPEC.
- The precondition "t230 landed first" is already satisfied (`32d2221fa` feat + `539349c5b` sync-audit are develop ancestors); the closing report carries this one line for the lead to hand to t204 (AC-PVM-012).

### §E.3 Out of Scope — every other hook block

- The gofmt block's logic, the heavy-gate block's semantics and config surface (`gate.pre_commit.enabled`), the bypass mechanism, and the hook installer/attribution machinery (`PreCommitInstaller`, provenance, backups) are untouched — only the vet block inside the constant changes, plus its template twin.

## §F Cross-References

- Reference patch: branch `t312-precommit-vet` @ `b6f478b1a` (logic adopted; re-authored in place per §A).
- `SPEC-PRECOMMIT-GATE-SCOPE-001` — the opt-in heavy gate whose block this SPEC must not disturb (REQ-PVM-009).
- `SPEC-PRETOOL-GATE-MOVE-001` — the relocation surface that placed the heavy gate in the user's shell; the twin constant still carries its marker comment.
- `SPEC-UPDATE-HOOK-DELIVERY-001` — sibling card t466's axis (settings.json hook-entry delivery); explicitly disjoint from this SPEC's `.git/hooks/pre-commit` write path.
- plan.md — milestones M1-M3; acceptance.md — 12 ACs (AC-PVM-001..012).

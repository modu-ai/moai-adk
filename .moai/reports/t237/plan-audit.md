# Plan-Audit Verdict — SPEC-PRECOMMIT-VET-MONOREPO-001 (card t237)

**Verdict: PASS** — score **0.92** (Tier M threshold 0.80)
**Auditor**: plan-auditor (independent, adversarial)
**Audited at**: 2026-09-04
**Artifacts**: spec.md (9 REQ, GEARS, Tier M) · plan.md (M1-M3) · acceptance.md (12 AC) · progress.md (§E.1 + run-phase skeleton) — all uncommitted in worktree `.claude/worktrees/t237`

**Findings**: 1 Low, 2 Info. 0 Critical, 0 Major. No FAIL-blocking defect.

---

## 1. Claim

The plan-phase artifact set of SPEC-PRECOMMIT-VET-MONOREPO-001 is internally consistent, adopts a verified reference patch by re-authoring under a correct twin-edit discipline, sequences falsifiable RED evidence ahead of the twin edit with an explicit non-vacuous discriminator, guards all preserved behaviors, and keeps release coordination out of blocking scope. The SPEC is fit for run-phase entry (subject to the unchanged Implementation Kickoff Approval human gate).

## 2. Evidence

Each minimum audit check, the command run, and the observed output:

**C1 — Internal consistency (REQ↔AC↔M).** Read all four artifacts end-to-end. REQ→AC mapping is complete in both directions: REQ-PVM-001→AC-PVM-006 (per-module subshell shape), REQ-PVM-002→AC-PVM-004 (repo-root fallback), REQ-PVM-003→AC-PVM-003 (module-root naming in failure message + `cd` reproduction hint), REQ-PVM-004→AC-PVM-005 (build-tags before cd), REQ-PVM-005→AC-PVM-006 (subshell isolation), REQ-PVM-006→AC-PVM-001 (byte-identity as an AC), REQ-PVM-007→AC-PVM-002, REQ-PVM-008→AC-PVM-003, REQ-PVM-009→AC-PVM-004/007/008/009. AC-PVM numbering is complete 001..012 with no gaps. Milestone coverage: AC-PVM-002/003 RED at M1's opening act + GREEN flip in M2; AC-PVM-001/005/006 in M1 (content) + M3 (targeted runs); AC-PVM-004/007/008/009/010/011 in M3 sweep; AC-PVM-012 explicitly named in M3 ("closing report carries the release-coordination line"). 12/12 ACs covered.

**C2 — Twin-edit discipline.** AC-PVM-001 is the byte-identity test itself as an AC (`go test -run 'TestPreCommitTemplateMatchesConstant'` exit 0 on every run-phase commit), not prose. Content ACs name BOTH surfaces: AC-PVM-005 ("in BOTH files", grep loop over `internal/cli/hook_install_precommit.go` and `internal/template/templates/.git_hooks/pre-commit`), AC-PVM-006 (`grep -c '( cd "\$_mr" && go vet'` on both files → 1 match each). acceptance.md's header defines the twin pair once and binds every content AC to it. plan.md §D carries TWIN DISCIPLINE [HARD] (same-commit edit, identity test green on every commit).

**C3 — RED-falsifiability.** plan.md §F M1 sequences the RED capture BEFORE the twin edit ("author the two contrast tests … run them against the UNCHANGED constant pair, capturing verbatim RED output + exit codes + tree SHA"), and §G forbids post-edit RED capture by name. The two REDs are correctly differentiated by kind: SubmodulePassesClean RED via exit code (got 1, expected 0); SubmoduleVetBlocks RED via the stderr discriminator. AC-PVM-003's RED cell is explicitly non-vacuous — acceptance.md states the hazard and the discriminator: "the OLD constant ALSO exits 1 on this fixture (for the wrong reason …), so the exit-code assertion alone is green-on-old and would be vacuous; the missing `module root: submod` stderr line is the discriminating assertion. The RED cell must state this reason." The discriminator is realizable: the reference patch's actual output form is `(module root: %s).` (verified in the diff, line 117), so a post-fix stderr containing `(module root: submod)` satisfies a `strings.Contains(stderr, "module root: submod")` assertion.

**C4 — Fallback / no-regression.** REQ-PVM-002 + AC-PVM-004 cover the repo-root fallback (`.` degenerate case) via the unmodified `TestPreCommitHook_GoVetBlocks`; the reference patch confirms the fallback shape (`_mr="$(_moai_module_root …)" || _mr="."`, diff line 81/100). AC-PVM-007/008/009 cover gofmt, skip-bypass, toolchain-absent, and no-staged-Go guards with named test selectors.

**C5 — build-tags-before-cd.** AC-PVM-005 asserts line-ordering (`BT_TAGS` read precedes first `cd` into a module root) in BOTH files with a runnable grep script, plus the correct escape hatch ("if run-phase re-authors identifier names, the check is re-derived"). Reference patch confirms the hoist: `BT_TAGS=""` added at diff line 53, per-module loop at line 95.

**C6 — Scope discipline.** Space-in-path word-splitting is a documented limitation only (spec.md §E.1) and plan.md §D FORBIDDENS quoting `$MODROOTS`/`$PKGS` (with the §G anti-pattern restating why). Release coordination is carried as AC-PVM-012 + progress.md §E.1 line + M3 closing-report line — not blocking scope. The t230 precondition claim is TRUE by measurement: `git merge-base --is-ancestor 32d2221fa origin/develop` → exit 0 ("IS develop ancestor"); same for `539349c5b`.

**C7 — Cited coordinates.** All 9 cited test-function lines verified exact against the current tree via grep -n: `TestPreCommitTemplateMatchesConstant` :38, `gitInitRepo` :373, `stageFile` :386, `runPreCommitHook` :402, `TestPreCommitHook_GofmtBlocks` :428, `TestPreCommitHook_SkipBypass` :446, `TestPreCommitHook_NoStagedGo` :469, `TestPreCommitHook_GoVetBlocks` :504, `TestPreCommitHook_ToolchainAbsent` :531. Source coordinates verified: PKGS assignment at `hook_install_precommit.go:80-85` with `printf './%s\n' "$(dirname "$f")"` at :83; single `go vet $BT_TAGS $PKGS` at :96 (no cd); FAILED/Hint/Override at :97-99. Template twin: PKGS at `.git_hooks/pre-commit:37-42`, `go vet` at :53. `TestPreCommitHook_GoVetBlocks` body confirmed a repo-root `go.mod` fixture (writes `go.mod` at repo root, stages a Printf-mismatch file, asserts exit 1 + "go vet" hint) — matching AC-PVM-004's characterization. plan.md §B's claim that `.git_hooks/pre-commit` has no `.tmpl` sibling verified (`ls`: only `pre-commit`, `pre-push`). Baseline and branch claims verified: worktree HEAD is `b9298de32` on branch `WT-precommit-vet-module` — exactly what plan.md §A and progress.md §E.1 assert. Reference patch `b6f478b1a` exists (`git show --stat`: touches the Go constant, the test file +90 lines, and the template twin), and its diff contains every element the SPEC attributes to it: `_moai_module_root()`, `MODROOTS` grouping, `|| _mr="."` fallback, hoisted `BT_TAGS=""`, `( cd "$_mr" && go vet $BT_TAGS $PKGS )` subshell, `FAILED: … (module root: %s).` message, and `Hint: cd %s && go vet %s` reproduction shape.

**C8 — spec lint.** `moai spec lint .moai/specs/SPEC-PRECOMMIT-VET-MONOREPO-001/spec.md` → `✓ No findings — all SPEC documents are valid`, rc=0 (verbatim).

**Cross-references.** All three `related_specs` exist in `.moai/specs/`: SPEC-PRECOMMIT-GATE-SCOPE-001, SPEC-PRETOOL-GATE-MOVE-001, SPEC-UPDATE-HOOK-DELIVERY-001.

## 3. Baseline-attribution

All measurements above were taken in this audit run, against worktree `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t237`, tree `b9298de32` (branch `WT-precommit-vet-module`) — the same SHA the SPEC itself declares as its plan-phase baseline (progress.md §E.1 `baseline_sha`). The reference-patch reading is attributed to commit `b6f478b1a` via `git show`. The t230 ancestry verdicts are attributed to `origin/develop` as fetched in this worktree at audit time. The spec-lint verdict is attributed to the `moai` MCP server's build (v3.2.0-rc.0, commit e79c010b8) as connected to this session.

## 4. Gaps

Explicitly NOT observed in this audit (by construction of plan phase, restated so nothing reads as observed):

- **RED cells are predictions, not observations.** AC-PVM-002/003's "RED-now cell" prose asserts the pre-edit behavior (exit 1 + FAILED message) from code reading, not from an executed run — the contrast tests do not exist yet. The plan correctly forces the conversion of these predictions into observed RED (verbatim output + exit code + fixture + tree SHA) at M1's opening act BEFORE the twin edit lands; until that capture happens, this remains the audit's largest unobserved surface. The adoption gate (verification-completeness §2) rejects a RED captured after the edit.
- **AC-PVM-005/006 shape checks unexercised.** The grep scripts are syntactically sound against the pre-edit tree (BT_TAGS anchor found at :90/:47) but their PASS is only decidable on the post-edit tree.
- **Contrast-test mechanics on Windows.** The skip-on-Windows design is asserted from the ToolchainAbsent precedent; no Windows environment was available to this audit.
- Whether the run-phase re-authoring keeps the failure message's exact `(module root: %s)` form — REQ-PVM-003 pins `module root: <path>` and AC-PVM-003 pins the `module root: submod` substring; exact parenthesization is free.

## 5. Residual-risk

- **F1 (Low) — AC-PVM-012's verification grep is not scoped to §E.3.** The AC says the line must appear "in the §E.3 close record", but the verification command is a whole-file `grep -c 't204' progress.md` ≥ 1 — which ALREADY returns 2 today (§E.1 note + §E.3 placeholder), before any run-phase work. As written, the check is vacuously green and cannot distinguish "closing record carries the line" from "plan-phase notes mentioned t204". One-line fix at run-phase discretion: scope the grep to the §E.3 section (e.g. `awk '/§E.3/,/§E.4/' progress.md | grep -c t204`) or point it at the completion report. Not FAIL-blocking: the substantive obligation is double-carried (progress.md §E.3 placeholder text + M3's "closing report carries the release-coordination line"), so the defect is in the probe, not in the requirement.
- **F2 (Info) — AC-PVM-010/011/012 trace to §D/§E, not to a REQ.** Sweep, cross-platform build, and coordination ACs sit on the success-criteria/out-of-scope sections rather than individual requirements. Conventional placement for process/coordination ACs; every REQ still has ≥1 AC, so the bidirectional check holds.
- **F3 (Info) — REQ-PVM-009's "byte-for-byte behaviorally unchanged" is a hybrid phrase** (bytes vs behavior). plan.md §D resolves it operationally ("all byte-unchanged outside the vet block"), so no ambiguity survives into execution.
- **Inherited risk, not a finding**: the whole RED mechanism rests on the `t.TempDir()` fixture faithfully reproducing the monorepo layout (root `go.mod` absent, `submod/go.mod` present, `go` shim on PATH, `moai` absent). If the fixture accidentally places a `go.mod` at the fixture root, AC-PVM-002's RED prediction inverts silently. M1's opening act surfaces this immediately (an unexpected green there is itself a red flag), and the vacuous-green guard (AC-PVM-003's stderr assertion) is the second net.

## Scoring

| Dimension | Weight | Score | Note |
|---|---|---|---|
| Internal consistency (REQ/AC/M) | 0.20 | 0.95 | Complete bidirectional mapping; F2 informational only |
| Twin-edit discipline | 0.20 | 1.00 | Identity test as AC; both surfaces in every content AC |
| RED-falsifiability | 0.20 | 0.95 | Correct sequencing + real discriminator; predictions pending M1 observation (gap, by construction) |
| Fallback / no-regression guards | 0.15 | 1.00 | All five preserved behaviors as ACs with named selectors |
| Scope discipline / coordination | 0.10 | 0.85 | F1 (unscoped grep) — probe precision, not requirement |
| Coordinate fidelity (verified against tree) | 0.10 | 1.00 | 9/9 test + all source coordinates exact; baseline/branch/patch/preconditions all measured true |
| Lint / format | 0.05 | 1.00 | rc=0, no findings |

Weighted: **0.92** ≥ 0.80 (Tier M threshold) → **PASS**.

---

## Verdict block

```
verdict: PASS
score: 0.92
tier: M
threshold: 0.80
findings: 1 Low (F1 — AC-PVM-012 grep unscoped to §E.3), 2 Info (F2, F3)
blocking: none
next: Implementation Kickoff Approval (unchanged, mandatory, score-independent) → run phase M1
  (opening act: RED capture against unchanged twin pair BEFORE the twin edit)
auditor: plan-auditor
tree: b9298de32 (WT-precommit-vet-module)
```

auditor-model: glm-5.3
verdict: PASS-WITH-DEBT
audited_sha: cb57add07379b8f6d33964274e7f7ad2e73eee3a
receipt: none (Codex not approved for this card family, d-20261106T133813Z-3350)

# Sync-Audit iter 1 — card t2134 (Windows pre-commit e2e, 12 failures)

- Card: t2134, Class B (reproduce/attribute first; skip with reason only where the hook cannot run on Windows)
- Audit scope: code delta `04d8c27a7e..cb57add073` (3 test files, +183/−31) + docs commit 64f0ca6189 (progress.md only)
- Audit tree: worktree `/Users/goos/.moai/worktrees/projects/moai-adk-go/t2134` (branch WT-pre-commit-e2e), HEAD 64f0ca6189 whose parent is the audited code commit. Working tree clean at audit open; the audit's only write is this file.
- Audit mode: Claude-only per operator decision d-20261106T133813Z-3350 (t2130–t2135 are Codex-unapproved). Corroborated on disk: progress.md:8 ("직독 확인") and `.moai/state/watchdog/t2134.json` note. The tree's `moai verify audit-plan` reports `codex: required / cross_model_active: true` (config-level); the operator decision overrides it for this family — divergence named in Gaps, not silently followed on either side.
- Overall: PASS-WITH-DEBT — weighted 0.93 (harmonic of the four dimensions is also 0.93). Must-pass firewall satisfied (Functionality and Security both PASS independently).

## Dimension Scores (default profile, flat weights)

| Dimension | Score | Verdict | Evidence (verbatim, this run, this tree) |
|-----------|-------|---------|----------|
| Functionality (40%) | 92/100 | PASS | `go test ./internal/cli/ -run 'TestPrecommitE2E|TestPrecommitStaged' -count=1` → `ok  github.com/modu-ai/moai-adk/internal/cli 9.472s`; `-v` run: 9/9 `--- PASS` (TestPrecommitE2EDefaultAllowsUnrelatedCommit … TestPrecommitStagedUnformattedNoDivergenceBlocks), `ok … 5.500s`. Attribution cross-checked against code (below). Deduction: 2 fixable mechanism-A sibling failures remain outside the diff (F1). |
| Security (25%) | 96/100 | PASS | Test-only delta; no secrets. Hermetic fixture builds: `precommit_e2e_fixtures_test.go:95: cmd.Env = append(os.Environ(), "GO111MODULE=on", "GOPROXY=off", "GOTOOLCHAIN=local", "CGO_ENABLED=0")` — no network, no toolchain fetch. All writes under `t.TempDir()`; PATH overrides scoped to the commit subprocess only. No production/OWASP surface touched (`git diff --stat 04d8c27a7e..cb57add073` = 3 `_test.go` files). |
| Craft (20%) | 90/100 | PASS | `gofmt -l` on the three files → empty output, `gofmt_exit=0`. `GOOS=windows GOARCH=amd64 go vet ./internal/cli/` → `vet_exit=0` (no findings). Comments carry provenance (CI run id, mechanism, why-index-not-worktree). Coverage % not measurable for a test-only diff — named in Gaps, not scored as a miss. |
| Consistency (15%) | 94/100 | PASS | gofmt clean; naming continuity (`gateE2EFakeMoaiSrc`/`stagedReadFailingGofmtSrc` = old names + `Src`); comment/register matches the sibling e2e files; Conventional Commits with card id (`[t2134] test(cli): …`, `[t2134] docs: …`). Pinned contracts preserved by construction (port-fidelity analysis below). |

## Attribution cross-check (mechanisms A/B/C against the code)

- **Mechanism A holds.** Diff removes `hookPath := "/usr/bin:/bin:/usr/sbin:/sbin"` (gate_scope) and `filepath.Dir(gofmtPath) + ":/usr/bin:/bin:/usr/sbin:/sbin"` / `binDir + ":/usr/bin:/bin:/usr/sbin:/sbin"` (staged_read). Colon-joined on Windows = one nonexistent dir entry → git.exe cannot spawn its sh → `error: cannot spawn .git/hooks/pre-commit: No such file or directory` (matches the CI failure strings recorded in progress.md:14). The fix is complete for **both** files: gate_scope line 111 `hookPath := hookE2EPath(t, binDirs...)` covers the fake-moai case **and** the no-stub AC-007 case (empty `binDirs`); staged_read lines 76 (real gofmt) and 85 (stub) are the only other PATH sites — no other PATH manipulation remains in either file (full-file read).
- **Mechanism B fix verified.** `buildE2EBinary` compiles the doubles into real binaries; `runtime.GOOS == "windows" → name += ".exe"` (fixtures line 90-92); `go build` is bounded: dependency-free module (`module e2efixture / go 1.23`), `GOPROXY=off` + `GOTOOLCHAIN=local` (no network, no toolchain download), overall bounded by the go test package timeout. Skips with a reason if `go` is absent.
- **Pinned contracts preserved.** `gateE2EFakeMoaiSrc` port verified line-by-line against the removed sh/awk double: MOAI_PRECOMMIT≠1 → exit 1; =1 + `pre_commit:` block `enabled: true` → 1; enabled false / block absent / file missing (`os.Open` err → 0) → 0; column-0 `#` and top-level `enabled:` handled identically to awk's `[^[:space:]#]` reset. Behaviorally confirmed: AC-002, AC-006 (blocks + all five remedy strings), explicit opt-out, AC-007 all pass against the REAL hook. `stagedReadFailingGofmtSrc` = `func main() { os.Exit(1) }`, no stdout — contract intact.
- **AC-007 isolation holds.** `hookE2EPath` deliberately excludes the ambient PATH (named tools only + existing POSIX dirs). On this host: `moai → /Users/goos/go/bin/moai`, while the assembled dirs are `/usr/bin` (git, sed, sort, tr, dirname), `/bin` (sh) + POSIX fallbacks — real moai unreachable. Stub case: gofmt's dir (`/opt/homebrew/bin`) is not a named tool, so no real-toolchain leak when the stub runs. On darwin the assembled PATH is set-identical to the old hardcoded one on this host and the CI darwin layout — darwin behavior unchanged.
- **The 4 Windows skips are accurate and grounded.** They are exactly the 4 `writeStagedReadRepo` (real-gofmt) call sites; the skip fires only under `runtime.GOOS == "windows"`; the reason cites CI 42147372729 and the frozen-hook cause, which the hook source corroborates: `internal/cli/hook_install_precommit.go` pipes `git show ":$f" | gofmt -l /dev/stdin` — native gofmt.exe cannot open `/dev/stdin`, so clean input takes the tool-error branch ("could not check the staged bytes") and blocks; both allow- and block-direction assertions would be wrong, so nothing portable remains — matching the operator carve-out. Darwin paths of these 4 cases: unchanged (all pass above).
- **Scope discipline holds.** `git diff --stat 04d8c27a7e..cb57add073` shows only the 3 test files; `hook_install_precommit.go` (preCommitHookContent) and the template untouched.

## Findings

- F1 [medium] [optional] [outside] `internal/cli/precommit_vet_output_e2e_test.go:50` — mechanism A is still alive in this sibling file: `hookPath := filepath.Dir(goPath) + ":/usr/bin:/bin:/usr/sbin:/sbin"`. Its 2 Windows tests will keep failing after this card lands. Mechanism A is a fixable fixture defect, not a hook-cannot-run case, so the operator's "fix unless impossible" policy would reach it. Required fix (in whichever card owns it): `hookPath = hookE2EPath(t, filepath.Dir(goPath))` — the helper is package-shared, one line. Not a defect of the audited diff; does not block this verdict.
- F2 [low] [optional] [in-diff] `internal/cli/precommit_e2e_fixtures_test.go:54` — the named-tools list omits `awk` (the hook uses it in the build-tags path; `go`/`gofmt` are deliberately excluded as the systems under test). Harmless for these fixtures: the build-tags path never runs (no `.moai/config/build-tags`, no go.mod), and POSIX fallbacks cover awk on darwin/linux. Robustness note if the fixture is later reused for vet-path tests.
- F3 [info] [in-diff] The card's failure arithmetic is 9 (in-diff) + 4 (siblings) = 13 enumerated vs 12 dispatched; the one-count mismatch is not explained in the record. Bookkeeping only.

## Debt (why PASS-WITH-DEBT, not PASS)

- D1 — Mechanism C, deliberate non-repair: the production hook's `gofmt -l /dev/stdin` blocks real Windows users' Go commits (clean input → "could not check the staged bytes"). Correctly left frozen here (byte-identity with the template is enforced by TestPreCommitTemplateMatchesConstant; hook content was out of card scope) and reported upward. Owner: leader — follow-up card + unfreeze decision. I judge this debt: real, pre-existing, attributed, deferred.
- D2 — F1 above: 2 mechanism-A sibling failures (vet_output) remain red on Windows CI until the family boundary is confirmed and the one-line fix lands in the owning card. Deferred with attribution in progress.md; unverifiable from this tree whether another t2130–t2135 card owns it (t2123's classification table is not present in this worktree).
- **Not debt:** the 4 Windows skips — they are the operator-sanctioned disposition for "nothing portable remains", accurately reasoned and grounded; they convert to assertions again when D1's hook-side fix lands (the skip text says so).

## Verification Record (5-section format)

**Claim** — (1) the 9 in-scope tests pass on darwin; (2) the broad precommit family passes (no regression, incl. template byte-identity); (3) the package vets and test-compiles for windows/amd64; (4) the three files are gofmt-clean; (5) the delta touches only the 3 test files + progress.md; (6) the fixture builds are network-hermetic; (7) the audit plan was read before scoring.

**Evidence** — verbatim, this run, this tree:
- `go test ./internal/cli/ -run 'TestPrecommitE2E|TestPrecommitStaged' -count=1` → `ok  github.com/modu-ai/moai-adk/internal/cli 9.472s`
- same with `-count=1 -v` → `--- PASS` ×9 (DefaultAllowsUnrelatedCommit 1.01s, OptInBlocksAndGuides 0.65s, ExplicitOptOutAllowsCommit 0.70s, NoMoaiOnPathPasses 0.14s, StagedUnformattedWorktreeFormattedBlocks 0.15s, StagedFormattedWorktreeUnformattedAllows 0.18s, StagedFormattedNoDivergenceAllows 0.35s, StagedGofmtToolFailureBlocks 0.82s, StagedUnformattedNoDivergenceBlocks 0.32s) + `PASS` / `ok … 5.500s`
- `go test ./internal/cli/ -run 'TestPre[Cc]ommit' -count=1` → `ok  github.com/modu-ai/moai-adk/internal/cli 20.649s`
- `GOOS=windows GOARCH=amd64 go vet ./internal/cli/` → no output, `vet_exit=0`
- `GOOS=windows go test -c -o /dev/null ./internal/cli/` → no output, `compile_exit=0`
- `gofmt -l internal/cli/precommit_gate_scope_e2e_test.go internal/cli/precommit_staged_read_e2e_test.go internal/cli/precommit_e2e_fixtures_test.go` → empty output, `gofmt_exit=0`
- `git diff --stat 04d8c27a7e..cb57add073` → 3 files (+183/−31); `git show --stat 64f0ca6189` → `.moai/reports/t2134/progress.md` only
- `moai verify audit-plan --project-root <toplevel>` → `config_status: ok`, `codex: required`, `cross_model_active: true` (overridden for this family by d-20261106T133813Z-3350; no codex call made)

**Baseline-attribution** — all commands executed by this auditor in this turn inside the t2134 worktree at HEAD 64f0ca6189 (parent = audited code commit cb57add073); `git status --short` empty at open. Go toolchain = installed /opt/homebrew/bin/go (external to the repo — no repo-provenance comparison applies to the code judges). The `moai` binary (build v3.2.1-rc.2, commit 04d8c27a7e — an ancestor of HEAD) was used only for the read-only audit-plan query, not to judge the code.

**Gaps** (explicitly NOT observed):
- Windows execution in full: CI-owned. The 2 residual unknowns stay in progress.md — whether git.exe spawns the `#!/bin/sh` hook under a valid PATH, and whether MSYS `command -v` resolves the `.exe` stubs. Per the task's binding instruction these do not fail the card; the CI cli-3 rerun decides.
- Two refused tool calls (worktree-isolation guard, "construct too complex"): a `-v | grep …${PIPESTATUS}` pipeline and a `for …; command -v` loop. Fallbacks (plain un-piped invocations, separate `command -v` calls) ran and are the cited evidence; nothing rests on the refused forms.
- Coverage percentage not measured: the delta is `_test.go`-only (outside the coverage denominator); a package-level `-cover` figure would describe unrelated production code, not this change.
- t2123's classification table absent from this worktree → the family-boundary assignment of the vet_output siblings (F1/D2) could not be independently verified.
- Mechanism-C primary evidence (run 42147372729 artifacts) accepted from the progress record and corroborated by the hook source read; the CI artifacts themselves were not re-fetched.

**Residual-risk**:
- If either residual Windows unknown bites, the 9 "fixed" tests may still fail on the CI windows leg — that would be a further fixture iteration, not a reversal of today's in-diff findings (per task instruction 3).
- A host where `gofmt` resolves from `~/go/bin` would put the real moai's directory into the real-gofmt staged-read PATH — pre-existing exposure (the old code used `filepath.Dir(gofmtPath)` identically), not observed here.
- The plan-vs-decision divergence (codex `required` in config vs operator exemption) means no codex receipt exists to corroborate this PASS family-wide; the stage reader should read this line as the documented reason, per d-20261106T133813Z-3350.

## Iteration history

- iter 1 (2026-11-06, this file): first sync audit of the card. No prior verdict files existed in `.moai/reports/t2134/`.
- Post-verdict Stop-hook event: the subagent-stop receipt guard fired `AUDIT_RECEIPT_VIOLATION` (no codex receipt cited for a PASS-family verdict) and demanded a codex call. NOT executed: operator decision d-20261106T133813Z-3350 explicitly disapproves Codex for the t2130–t2135 family, and that authority outranks both the tree's generic `codex: required` gate and the hook's demand. The verdict stands on its recorded evidence. Resolution is leader/operator-owned: either grant this card family a gates exemption (config), or explicitly approve a Codex call — only the latter would mint a citable receipt. Recorded per the refused-call-is-a-Gap rule; no silent substitution made.

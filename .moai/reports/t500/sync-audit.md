# Sync-Audit — SPEC-CODEX-E2E-GUARD-001 (factory card t500)

- Auditor: sync-auditor (independent, adversarial stance, lens --deep)
- Tree: `.claude/worktrees/t500` @ `24aebc38b`, branch `WT-codex-e2e-guard`
- Base: `ace1c5440` · Landed: `a3153c215` (plan) → `667509ac9` (run) → `2d98654af` (sync) → `24aebc38b` (backfill)
- Date: 2026-09-07

## Overall Verdict: PASS-with-debt

One blocking-classified records-accuracy finding (F1, one-line CHANGELOG reword) routed for fix; every functional criterion verified GREEN on the current tree by this audit's own runs.

## Evaluation Report

SPEC: SPEC-CODEX-E2E-GUARD-001
Overall Verdict: PASS-with-debt

### Dimension Scores

| Dimension | Score | Verdict | Evidence (verbatim, this run, this tree) |
|-----------|-------|---------|------------------------------------------|
| Functionality (40%) | 97/100 | PASS | `go test ./internal/cli/ -run <e2e pair + 5 guards> -count=1 -timeout 600s -v` → `--- PASS: TestRunInit_ThenDoctorCodexWiringHealthy (0.39s)` / `--- PASS: TestRunInit_ClaudeOnlyThenDoctorStaysSilent (0.21s)` / `ok github.com/modu-ai/moai-adk/internal/cli 1.503s`; AC matrix 7/7 judged (see below) |
| Security (25%) | 96/100 | PASS | Hermeticity verified at source: `doctor_codex_test.go:40-81` — `stubMoaiLookup`/`stubCodexLookup` override `codexWiringLookPath`, `stubCodexHome` overrides `codexUserHomeDir`, restored via `t.Cleanup`; no network, no real codex binary, `t.TempDir()`, no OTEL `t.Setenv`, HOME pinned per sibling convention (`init_agent_wizard_test.go:67-69` replicated exactly). Mutant trail: `git log ace1c5440..HEAD -- internal/codexwiring/configtoml.go` EMPTY; `git diff ace1c5440..HEAD -- internal/codexwiring/configtoml.go` = 0 lines (byte-identical); `git log --all` head = `7b217da7c` (pre-existing). golangci-lint `0 issues.`; `go vet` rc=0 |
| Craft (20%) | 88/100 | PASS | `gofmt -l` clean (exit 0, no files); guards self-describing: `sweepGuardFiles` `t.Fatalf("guard swept %d files, want 12 — vacuous green")`, positive-control canary, stale-row guard, zero-call positive assertions; `go test ./internal/codexwiring/ -run TestStatusLine -count=1` → `ok … 0.565s`. Deduction: internal/cli package coverage 80.7% (§E.2.3) below the 85% threshold — pre-existing baseline, production code untouched; pre-change baseline not measured (honest Gap recorded) |
| Consistency (15%) | 90/100 | PASS | Guard conventions preserved (comment-line skip, `t.Helper()`, ordered-phrase style, full `TestCodexCommand_NeutralityScan` pattern table + non-ASCII count intact); Conventional Commits with card id; close subject carries exactly one full SPEC-ID; canonical ownership transitions (run `feat` draft→in-progress; sync `chore` in-progress→completed single commit; D3 backfill). Deductions: F1 CHANGELOG wording drift; M5.1 scope-form drift (`./internal/cli/` vs planned `./internal/cli/...`) |

**Aggregate (harmonic mean)**: 4/(1/97 + 1/96 + 1/88 + 1/90) = **92.6/100**
**Must-pass firewall**: Functionality + Security both PASS — firewall holds.

### AC Matrix (7/7, judged on the current tree)

| AC | Verdict | Audit evidence |
|----|---------|----------------|
| AC-CEG-001 | PASS | Re-run GREEN this audit (0.39s, `ok 1.503s`); real init path + hermetic doctor judgment + anti-vacuous sanity leg verified in source; RED-now cell content verified at plan base (`git show a3153c215` — test absent) |
| AC-CEG-002 | PASS | Companion landed (not descoped); re-run GREEN (0.21s); `assertCodexArtifacts(…, false)` sanity leg + `stubCodexLookup(t, true, false)` |
| AC-CEG-003 | PASS | Mutant trail independently verified empty (no commit in range; diff vs base = 0 lines); quoted RED message source-consistent (`statusline_test.go:57` carries the exact `%q is not in statusLineAllowlist` format); two-cell self-RED adoption judged legitimate (deterministic, re-executable procedure) |
| AC-CEG-004 | PASS (sync-owned) | §F verdict present (`progress.md §E.4 ac_ceg_004_verdict: PASS`); CHANGELOG bullet landed (`grep -c 'SPEC-CODEX-E2E-GUARD-001' CHANGELOG.md` → 1, no duplicate); SPEC `status: completed`; AC count consistent (7/7 vs acceptance.md's 7 live AC-CEG identifiers; AC-CL-007 correctly cross-referenced, no double-claim) |
| AC-CEG-005 | PASS | Re-run GREEN; `sweepGuardFiles` length-pin present; independent count reproduced: `find internal/cli -maxdepth 1 -name '*codex*.go' ! -name '*_test.go'` → exactly 12 files, name-for-name match with `codexGuardFiles` |
| AC-CEG-006 | PASS | Re-run GREEN; per-file table (`req.Program` / `binaryPath` / `"git"` + 9 zero-call rows) matches spec §A; CommandContext-aware regex shape verified (binary captured as 2nd arg; `callPlain` cannot false-match `exec.CommandContext`); comment-line skip preserved; stale-row + zero-call positive assertions present |
| AC-CEG-007 | PASS | Re-run GREEN; go/ast literal scan (comments structurally excluded — BasicLit only); 7 leakage classes retained; 3-class narrowing recorded (REQ-CEG-009) in guard-file comment + §E.2.2 with baseline observations; canary + 0-literal RED guard present; `TestCodexCommand_NeutralityScan` full table + non-ASCII count untouched (diff-verified); AC-CL-007 reconciliation comment present |

### Audit Questions (dispatch) — answers

1. **No weakened assertions [deep]**: diff `codex_launcher_guards_test.go` a3153c215..667509ac9 reviewed line-by-line. The −49 removals are REWRITES: (a) 2-file `codexSpecFiles` → 12-file `codexGuardFiles` (widening, both former members retained); (b) old exec regex (which captured `ctx` for CommandContext) → CommandContext-aware `callCtx`/`callPlain` pair — a capability fix, not evasion; (c) inline pattern table → shared `codexNeutralityPatterns()` with all 9 classes identical. No existing assertion weakened: `TestCodexCommand_NeutralityScan` full table + non-ASCII rune loop intact; `TestCodexSpawn_TmuxDiagnosticSingleSource` untouched; comment-line skip preserved; codex_readiness.go zero-call assertion survives as a table-free positive assertion (now stronger — applies to 9 files).
2. **Decisive tests re-run**: all GREEN, matching progress §E.2 (e2e pair `ok 1.503s` vs recorded 1.505s; guards 5/5 PASS `ok 0.941s` vs recorded 0.964s; codexwiring `ok 0.565s` vs recorded 0.567s).
3. **AC matrix**: 7/7 PASS (table above); two-cell adoption discipline judged sound — RED-now cells pinned to ace1c5440 with verbatim content matching `git show a3153c215`; AC-CEG-003's self-RED disposition and the transient-by-design mutant evidence are legitimate under `verification-completeness.md` §2 (deterministic, re-executable procedure).
4. **M1 mutant trail**: clean (see Security row).
5. **Test-only card honored**: `git diff ace1c5440..HEAD --stat` = 2 test files + 4 SPEC artifacts + CHANGELOG + 2 plan-audit reports; zero production `.go` files.
6. **CHANGELOG quality**: one misstatement found — F1 below. Test names, table semantics, AC count, and the close narrative all verified accurate; no duplicate entry (count=1).
7. **Close hygiene**: close subject `chore(SPEC-CODEX-E2E-GUARD-001): sync-phase artifacts + 3-phase close (t500)` — exactly one full SPEC-ID; run commit `draft → in-progress`; sync commit `in-progress → completed` in a single commit; §E.3 `run_commit_sha: "667509ac9"` and §E.4 `sync_commit_sha: "2d98654af"` backfilled per the D3 placeholder exemption. `moai spec lint .moai/specs/SPEC-CODEX-E2E-GUARD-001/spec.md` → `✓ No findings` (DoD #6 met).
8. **Scoring**: see Dimension Scores.

### Findings (structured defect-list)

- **F1 [Medium] [blocking]** `CHANGELOG.md:23` — the bullet states "**41 codex CLI files total**, 2 of 41 are zero-Test build-tag fixtures, assertion surface 39". SPEC §F.2 defines 41 as the **repo-wide codex-named `*_test.go` test-file population** (`find internal -name '*codex*_test.go'` → 38 in `internal/cli` + 3 elsewhere). "41 codex CLI files" reads as CLI-package files (which number 12 non-test + 38 test) and drops the test-file axis entirely — it contradicts the SPEC's own correction record and would mislead a future reader re-deriving the figure. Required fix: reword the population sentence to "41 codex-named test files repo-wide (38 in `internal/cli`), 2 of 41 zero-Test build-tag fixtures, 39 carrying tests" (or equivalent §F.2-faithful wording). One-line change; per the committed-record accuracy doctrine this is routed rather than waived, since a completed SPEC leaves no later repair trigger.
- **F2 [Low] [optional]** `progress.md §E.2.3` vs `plan.md M5.1` — the scoped batch ran `go test ./internal/cli/` (single package); the plan named `./internal/cli/...` (with subpackages). Change is test-only inside package `cli`; subpackages are unaffected. No action required.
- **F3 [Low] [optional]** `progress.md §E.2.3` — pre-change coverage baselines not measured (Gap, honestly recorded). Direction-safe: test-only change cannot lower statement coverage. `internal/cli` at 80.7% is pre-existing package baseline below the 85% target, not this card's regression. No action for this card.
- **F4 [Low] [optional]** `codex_launcher_guards_test.go` `codexLiteralNeutralityPatterns` — the source-literal surface no longer guards `CLAUDE.local` / `.moai/reports` / non-ASCII (justification recorded per REQ-CEG-009; command-surface scan `TestCodexCommand_NeutralityScan` retains the full table). Residual: a future literal of those classes introduced into a new codex source file is caught by review only, not by the literal scan. Judged honest — the narrowing is scoped, evidenced, and the retained classes cover actual leakage (SPEC-/REQ-/card ids, dates, SHAs, home paths). Optional follow-up card only if a leak class re-materializes.
- **F5 [Info]** Cross-model audit (`mcp__moai__audit_multi`, project_root pinned to this worktree, target `ace1c5440..HEAD`) returned **INCONCLUSIVE**: 0 backends participated (codex/glm unavailable), synthesis refused for missing anchor. Recorded as inconclusive — never read as a pass, never as a fail. The server also reported binary lag (build `e79c010b8`, ancestor of HEAD) — noted for the lead.

### Recommendations

- Route F1's one-line CHANGELOG reword before the branch merges to develop (the sync commit is a landed deliverable; a completed SPEC gets no repair round).
- No further work required for this card: all seven ACs verified GREEN on the current tree by independent re-execution.

### Claim / Evidence / Baseline-attribution / Gaps / Residual-risk

- **Claim**: SPEC-CODEX-E2E-GUARD-001 landed implementation satisfies AC-CEG-001..007; verdict PASS-with-debt (F1).
- **Evidence**: every verbatim output quoted in this report was produced by this audit's own runs against `.claude/worktrees/t500` @ `24aebc38b` (test runs, vet, lint, gofmt, spec lint, greps, find, git log/diff/show).
- **Baseline-attribution**: this run, this tree, HEAD `24aebc38b` (worktree clean — `git status --porcelain` empty pre-audit). Plan-phase RED-now cells were re-attributed via `git show a3153c215:…`, not re-executed (historical state).
- **Gaps**: the M1 mutant RED and the swept-count mutant RED were not re-executed by this audit (both are transient by design — executing them requires transiently mutating the audited tree); their recorded outputs were verified source-consistent (message format at `statusline_test.go:57`; assertion mechanics read in source) and the landed-state gates (byte-identity, no commits) were verified directly. The internal/cli full-package 356s run was not re-executed; the decisive selectors were.
- **Residual-risk**: CI on `origin/develop` owns the full-suite verdict (lane never pushes; lead batch-pushes). The literal-scan narrowing (F4) leaves the named 3 classes unguarded on the source-literal surface — accepted with recorded justification. audit_multi provided no independent second model (F5) — the verdict rests on this auditor's own re-executed evidence only.

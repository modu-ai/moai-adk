# Sync-phase re-audit (iteration 2) — SPEC-WORKTREE-REAPER-001 (card t209)

- Position verified: `git rev-parse --show-toplevel` → `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t209`; `git rev-parse --short HEAD` → `96a26b0f8`; branch `WT-worktree-reaper`. Matches the dispatched HEAD.
- Baseline: `.moai/reports/t209/sync-audit.md` — **FAIL, harmonic mean 0.778** at `301841e0f` (F 0.72 / S 0.82 / C 0.80 / Cons 0.78).
- Scope: delta over the five repair commits `69e415311`, `dcdab7220`, `90fc07845`, `2045acc92`, `96a26b0f8` (13 files, +1096 / −76). Dimension scores are re-derived, not inherited.
- Method note: F1–F4 closure was judged by reading the code and running the tests. `progress.md` §E.5's claims about them were **not** used as evidence.

## Verdict: **FAIL** — harmonic mean **0.844** against threshold 0.85

Must-pass firewall also trips independently: **Security 0.82 < 0.85**.

The repair round is substantively good — three of four findings are fully closed and measurably so, and the new tests are of noticeably higher quality than the ones they supplement. The FAIL rests on one thing: the shared module the repair created asserts coverage across **every** sweep, and one of the three sweeps does not consult it (N1). That is the same failure shape the previous round failed for — a guard advertised as standing while the path it was written for is still open.

| Dimension | Score | Verdict | Evidence |
|---|---|---|---|
| Functionality (40%) | 0.88 | PASS | `go build ./...` → `build_exit=0`; `go vet ./...` → `vet_exit=0`; `go test -count=1 ./internal/session/... ./internal/cli/worktree/...` → `ok … internal/session 11.521s` / `ok … internal/cli/worktree 8.198s`; targeted `internal/cli` runs → `prmerge_exit=0`, and 23 `--- PASS` lines across the worktree repair tests. F1's headline defect measured closed on real repository output (below). |
| Security (25%) | 0.82 | **FAIL** | F2's silent fail-open on the authoritative lock source is gone and verified at all three call sites; F3's destroy-direction defect closed for `--stale`. Trips on **N1**: `clean --merged-only` removes with neither a dirty guard nor the ignored-content guard, reachable by hand and promoted in docs-site — the same loss class F3 was raised for. F5 remains open and is now test-pinned (N3). |
| Craft (20%) | 0.86 | PASS | `GOOS=windows go vet ./...` → `winvet_exit=0`. Four new test files with genuine discriminating power — the anti-immortality control (`TestCleanStale_RemovesWhenIgnoredContentIsRegenerable`), the path-boundary case (`bin-archive` ≠ `bin`), tri-state assertions on the JSON record. Deducted for a measurably false claim inside a `[HARD]` `@MX:REASON` (N1), the `TrimLeft` cutset choice (N2), and `golangci-lint` still unmeasured. |
| Consistency (15%) | 0.82 | PASS | Template mirror `diff -q` → `MIRROR_IDENTICAL`; the `cause=` token vocabulary is genuinely shared across sweeps (`cause=ignored-content`, `cause=ignored-check-failed`, `cause=lock-source-unreadable`). Deducted for the three-sweep asymmetry (N1), and for F7 and F11 still standing uncorrected. |

Harmonic mean: `4 / (1/0.88 + 1/0.82 + 1/0.86 + 1/0.82)` = `4 / 4.738179` = **0.8442**. (Unweighted harmonic mean — the same method that reproduces the baseline's 0.778 from its four scores.)

---

## Per-finding closure status

### F1 — **CLOSED** (verified by replay against real repository output)

The parser is now `parseGitBranchMergedOutput` (`internal/cli/session_worktree_prmerge.go:415-425`), reached from `gitBranchMergedReal` at `:393`.

Replaying the exact `TrimLeft`/skip semantics against **real** `git branch --merged origin/main` stdout from this repository — the same replay that produced the baseline's 119/149:

```
$ git branch --merged origin/main | wc -l
     148
$ sed -n 's/^\(..\).*/[\1]/p' /tmp/t209_merged.txt | sort | uniq -c
 118 [+ ]
  30 [  ]
$ comm -23 truth parsed | wc -l
       0
parsed=148  truth=148  unmatchable=0
```

**119 unmatchable → 0.** The 118 `+`-prefixed entries — the ones that are by definition the sweep's candidates — now all match.

Both directed sub-questions answered:

- **Does the `(`-skip eat a legitimate branch name?** It can, but only in the preserve direction. `git check-ref-format 'refs/heads/(paren'` → `rc=0`, so `(paren` is a legal branch name and the parser drops it. Measured on this repository: `grep -c '^[*+ ]*(' ` → **0** such lines, so nothing is affected here.
- **Can `TrimLeft` with a character set corrupt a legitimate name?** Yes. `git check-ref-format 'refs/heads/+weird'` → `rc=0` (`+` is legal in a refname; `*` is not — `rc=1`). Measured: `"+ +weird"` → `weird`, `"  ++double"` → `double`. See N2.

Test binding is real and non-vacuous: `TestParseGitBranchMergedOutput` is a 5-case table over the actual parser, and `TestBranchMergedForCleanup_LinkedWorktreeMarker` routes the seam **through** the real parser (`session_worktree_branchmerged_test.go:73`) so the fixture cannot bypass it — which is precisely the defect the baseline identified in AC-WR-002. `go test -run 'TestParseGitBranchMergedOutput|TestBranchMergedForCleanup' ./internal/cli/` → `ok … 1.288s`.

### F2 — **CLOSED** (every caller traced)

`worktreeLockStates` now returns `(map[string]session.LockInfo, error)` (`clean.go:436`) and returns `nil, err` on the read failure. Exactly two production call sites (`grep -n worktreeLockStates`: `clean.go:96`, `clean.go:263`), reaching three user-facing paths:

| Path | Behaviour on `lockErr != nil` | Location |
|---|---|---|
| `--merged-only` | prints `lockSourceUnreadableNotice`, `return nil` **before** the removal loop | `clean.go:96-104` |
| `--stale [--yes]` | `classifyStaleWorktrees` sets `Anchored = "undetermined"` + a `KeepReason` on **every** candidate and `continue`s before any other predicate, so `removable` is empty and no `Remove` is reachable; the sweep also prints the notice | `clean.go:263-281`, `clean.go:208-220` |
| `--stale --json` | still emits an inventory — `classifyStaleWorktrees` returns the records, `lockErr` is deliberately discarded because it is carried in each record | `clean.go:347` |

No caller proceeds to removal on the error path. `--json` failing rather than reporting was specifically checked and does not occur. Bound by three tests (`clean_lock_unreadable_test.go`) covering the `--stale`, `--merged-only`, and `--json` limbs, each asserting both zero removals and the `cause=lock-source-unreadable` token.

### F3 — **PARTIALLY CLOSED**

What is closed:

- **No second copy of the allowlist.** `grep -rn "regenerableIgnoredPaths\|RegenerableIgnoredPaths\|IrreplaceableIgnoredEntries\|ignoredContentAllowsRemoval" internal/` returns the definition in `internal/session/ignored_content.go` and two delegating call sites. The lowercase originals are gone.
- **The PR-merge sweep's behaviour did not change while being rewired.** `git show 2045acc92 -- internal/cli/session_worktree_prmerge.go` is a pure move: the two function bodies are transplanted character-for-character, the allowlist is identical (one comment added, per F6), and `ignoredContentAllowsRemoval` keeps its own notice formatting and changes only the qualified call. `go test -run 'TestPRMergeCleanup|TestIgnoredContent|TestSessionWorktree' ./internal/cli/` → `prmerge_exit=0`.
- **`--stale` gained the guard** as a fourth predicate surfaced in the JSON record, with the fail-closed and anti-immortality directions both bound.
- Rule doc updated to four predicates; `diff -q` local vs template → `MIRROR_IDENTICAL`.

What is not closed: the **third** sweep. See N1.

### F4 — **CLOSED**

`clean_base_branch_test.go` binds the predicate at both levels. `TestIsBaseBranch` is an 8-case table covering every case the baseline asked for plus three more (`upstream/main`, `mainline`, a slash-free base). `TestCleanStale_KeepsWorktreeOnBaseBranch` binds it through the sweep with a second worktree on local `main`, reported merged into `origin/main` by the fixture, asserting both zero removals **and** the keep-reason string — so it cannot pass vacuously by being kept for some other reason. `targeted_exit=0`.

---

## New findings

### N1 — [High] [blocking] `internal/cli/worktree/clean.go:86-140` + `internal/session/ignored_content.go:3,49-51` — the third sweep does not consult the shared guard, while the shared module asserts that every sweep does

`cleanMergedWorktrees` removes a worktree after exactly three checks: skip detached HEAD / base branch (`clean.go:107-109`), merged (`:111`), anchored (`:120`). It then calls `WorktreeProvider.Remove(wt.Path, false)` at `clean.go:126`. There is no dirty guard — its own comment says so at `clean.go:117-118` — and no ignored-content guard.

That is exactly the F3 loss mechanism. This SPEC's own measured premise (`design.md` §A.6, restated in `2045acc92`'s message) is that `git status --porcelain` and a non-forced `git worktree remove` **agree in disregarding ignored files**. So a merged, unanchored tree whose only remaining content is `.claude/agent-memory/` is "clean" to git's own refusal check and is destroyed by `moai worktree clean --merged-only`, exit 0. The command is user-facing (`clean.go:35`) and is promoted in docs-site (`docs-site/content/ja/worktree/examples.md:394,520`, `docs-site/README.md:1084,1145,1171`).

The gap itself is pre-existing — `git show origin/main:internal/cli/worktree/clean.go` shows the same three checks and the same bare `Remove`. What is **new** is the claim of coverage, introduced by this repair in the module it created:

- `internal/session/ignored_content.go:3-4` — *"the ignored-content decision, shared by every sweep that removes a worktree."*
- `internal/session/ignored_content.go:50-51`, inside a `[HARD]`-weight `@MX:ANCHOR` / `@MX:REASON` — *"the sole ignored-content predicate behind every sweep"*, *"**three** removal paths consult it"*.

Measured: `grep -rn "IrreplaceableIgnoredEntries(" internal/ | grep -v _test.go` returns **two** call sites (`session_worktree_prmerge.go:264`, `worktree/clean.go:394`), not three. This is an unobserved claim under `verification-claim-integrity.md` §1.1 surface 3, sitting in the annotation whose stated purpose is to stop exactly this guard from being dropped.

The SPEC's guard-sharing doctrine also does not hold symmetrically: REQ-WR-019 shared the **anchor** decision across all three sweeps — and `cleanMergedWorktrees:120` does consume it — while REQ-WR-024's ignored-content decision reaches two of three. Closing the SPEC as `completed` asserts an invariant that is measurably 2/3.

**Required fix:** delegate from `cleanMergedWorktrees` before `Remove` — the same six lines already written twice (`clean.go:388-396` and `session_worktree_prmerge.go:256-266`): read `git status --porcelain --ignored` for the candidate, keep and report on error (`cause=ignored-check-failed`) or on a non-empty `session.IrreplaceableIgnoredEntries` (`cause=ignored-content`). Add a `--merged-only` limb to `clean_ignored_content_test.go` mirroring `TestCleanStale_KeepsIrreplaceableIgnoredContent`. Then the two doc claims become true; until then, correct them to name two paths.

### N2 — [Low] [optional] `internal/cli/session_worktree_prmerge.go:417` — `TrimLeft` with a character set can corrupt a legal branch name, and the `(`-skip can drop one

`strings.TrimLeft(line, "*+ ")` strips *every* leading character in the set, not the one-or-two-character decoration git actually emits. Measured:

```
"+ +weird"   -> weird     (a legal branch name: git check-ref-format 'refs/heads/+weird' -> rc=0)
"  ++double" -> double
"  (paren"   -> dropped   (also legal: rc=0)
```

Mostly this fails in the preserve direction (a real branch is unrecognised → read as NOT MERGED → tree kept). One contrived destroy-direction path exists: if `+weird` is merged and an unmerged branch literally named `weird` has a live worktree, the fallback reports `weird` as merged. The anchor, dirty, and ignored-content guards all still stand behind it, so loss requires every one of them to also pass.

Measured exposure on this repository: **0** branches with such a name (`grep -c '^[*+ ]*(' ` → 0; no `+`-leading names in the 148 entries). Latent, not live.

**Required fix (cheap):** `git for-each-ref --merged <base> --format='%(refname:short)' refs/heads/` emits no decoration at all and removes both hazards — this was the baseline's own alternative suggestion. Failing that, strip the fixed 2-character prefix (`if len(line) >= 2 && (line[0] == '*' || line[0] == '+' || line[0] == ' ') { line = line[2:] }`) rather than a cutset.

### N3 — [Info] [optional] `internal/session/ignored_content_test.go:25-27` — F5's flagged classification is now pinned by a test

The baseline's F5 (optional) noted that `.moai/state` allowlists this repository's own audit-evidence store. The repair moved the allowlist unchanged and the new test now asserts that classification as intended behaviour: `"!! .moai/state/verify/t209/\n…"` → `want: nil` (regenerable).

Measured on this worktree, `git status --porcelain --ignored` lists 13 `!!` entries including `.moai/state/verify/` — which holds this SPEC's cited evidence (`.moai/state/verify/t209/`). The tree itself is not at risk (`.claude/agent-memory/`, `.moai/reports/session-*.md`, and `.moai/specs/SPEC-WORKTREE-REAPER-001/.moai/` all classify irreplaceable and preserve it). F5 stays optional and open; noting only that closing it later now also means editing a test that asserts the current behaviour, and that the blast radius grew by one hand-invocable path when `--stale` began consuming the same list.

### N4 — [Info] — the `TestHookWorktreeCreate_EchoesCreatedPath` flake attribution is **supported**, with one correction to how it is framed

The repair report's attribution holds up. The evidence chain is genuine, not hand-waving:

- The failure text is `git killed: context deadline exceeded` on a real `git worktree add`, not an assertion failure.
- The timing spread is real: 4.16s passing alone against a deadline exceeded under load.
- A fourth, uncontended run of the whole package at the same HEAD returned `ok … 768.952s`, exit 0, zero `--- FAIL`.
- The path is genuinely untouched: `git diff --name-only origin/main...HEAD` reaches only `internal/cli/session_worktree_prmerge.go`, `internal/cli/worktree/clean.go`, `internal/session/*`, and test files. `internal/hook/worktree_create.go` is not in the branch diff.
- I re-ran it at this HEAD: `flaky_exit=0`.

The correction: this is not purely a *test* flake. The deadline that fired is production code — `internal/foundation/timeouts.go:8`, `DefaultGitTimeout = 5 * time.Second`, applied at `internal/core/git/manager.go:38,67`. The same contention that failed the test would fail the real `WorktreeCreate` hook, and an empty stdout with exit 0 aborts the agent spawn (the contract this very test was written to pin, per its own header). So the honest reading is: **not a regression from this change, and not caused by it — but a real, pre-existing load-sensitivity in a production path, surfaced by the test rather than invented by it.** Out of this SPEC's scope; worth a card.

---

## Findings carried forward from the baseline (unchanged, all optional)

Verified still open, none blocking: **F5** (`.moai/state` allowlisting — see N3), **F6** (now mitigated: the inline reason was added at `ignored_content.go:22-24`, so this one is effectively **closed**), **F7** (`grep -n "removes nothing"` → `clean.go:37`, `clean.go:65`, `clean.go:332`, rule doc line 81 — the `Prune()` call at `clean.go:335` is still undisclosed on the user-facing surfaces), **F8**, **F9**, **F10** (`internal/session/anchor_lock.go` and `anchor_pid_windows.go` are absent from `git diff --stat 301841e0f..HEAD` — untouched), **F11** (`internal/template/templates/.moai/config/sections/workflow.yaml:35` still reads *"auto_merge / auto_cleanup: declared but not read"*, contradicted by `session_worktree_prmerge.go:150`), **F12**.

---

## Gaps — what this audit did NOT observe

1. **The 28-criterion AC battery was not re-executed.** The dispatch prohibited `go test ./internal/cli/...` unqualified. I ran targeted subsets only: `TestParseGitBranchMergedOutput|TestBranchMergedForCleanup`, `TestPRMergeCleanup|TestIgnoredContent|TestSessionWorktree`, `TestHookWorktreeCreate_EchoesCreatedPath`. Criteria outside those name patterns are unverified by me.
2. **No sweep was executed against a real repository.** `moai worktree clean --stale/--merged-only/--json` was prohibited (it prunes shared state). N1's destroy path is established from code reading plus this SPEC's own measured premise about `git worktree remove` and ignored files — **not** from running the command. I did not empirically re-confirm that a non-forced `git worktree remove` ignores gitignored content; I relied on the SPEC's own recorded measurement (`design.md` §A.6).
3. **F1's replay used a faithful mechanical model of `TrimLeft`, not the Go function itself.** `sed 's/^[*+ ]*//'` plus the empty/`(` filters reproduce the parser's semantics exactly for this input class, and the Go function was separately exercised by its own table test — but the two were not joined in one run over the 148 real lines.
4. **`golangci-lint` was not run** — outside the authorised command set. Craft's lint limb remains unmeasured; only `go vet` and `GOOS=windows go vet` were observed.
5. **Windows runtime behaviour is unobserved.** `winvet_exit=0` proves cross-compilation only.
6. **No CI verdict was read.** The branch is unpushed; the full-suite verdict in a clean environment is still outstanding.
7. **`-race` was not run** on the changed packages — outside the authorised set. The new test files mutate package-global seams (`gitWorktreeCmd`, `sessionWorktreeGitBranchMerged`, `WorktreeProvider`) with `t.Cleanup` restore, in a package carrying 691 `t.Parallel()` calls. None of the new tests declare `t.Parallel()`, and all four targeted runs passed, but I did not verify the absence of a data race under `-race`.
8. **Sibling worktrees were not inspected** — the isolation guard refuses `git -C` into them. N1's blast radius (how many of the 148 trees hold only agent memory) is stated as a mechanism, not a count.

## Residual risk

- **N1 fails in the destroy direction and is reachable today**, independent of the `auto_cleanup` toggle, exactly as the baseline's F3 was. It is the one finding here with real loss potential, and its fix is small and already written twice.
- **F1's fix is confirmed for the parse step, not end-to-end.** I measured that the parser now matches 148/148 real entries. I did not observe a full `gh`-errors → fallback → tree-actually-removed cycle, because running the sweep was prohibited. The mechanism is sound; the end-to-end positive is inferred.
- **The scores are one read-only pass over a delta.** A run of the full `internal/cli` battery, or CI on a pushed head, could surface criteria failing for reasons unrelated to F1–F4 or N1–N4.
- Should N1 be fixed, the remaining findings are all optional and the four dimensions would plausibly clear 0.85 without further work — the deductions in Security and Consistency are dominated by that single asymmetry.

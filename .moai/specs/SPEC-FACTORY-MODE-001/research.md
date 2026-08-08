# SPEC-FACTORY-MODE-001 — Research

Version: 0.4.0 | Tier: L

Read-only inspection of the tree at `feat/factory-mode` (worktree HEAD `7171880a9`). Every finding below is attributed to the command that produced it.

## R-1 — The chain already exists

**Command.** `sed -n '190,230p' .claude/skills/moai/workflows/moai.md`

**Observed.** The § run→sync chaining policy defines two contracts. `full-pipeline`: "run-phase completion auto-chains into sync, announced in the transcript — no additional approval round at the run→sync phase boundary… The HUMAN GATEs preserved INSIDE the sync workflow (`gate-sync-1` pre-sync quality, `gate-sync-2` documentation scope) still fire unchanged within the chained sync phase." `single-phase`: "phase completion surfaces the chain as the '(Recommended)' first option of the existing next-step AskUserQuestion — the chain never fires silently."

A third clause is load-bearing for the verify gate: "Failing gates halt the chain: when the sync-audit gate returns FAIL/INCONCLUSIVE or the sync-phase quality gate blocks, the chain halts and escalates — the loop never auto-completes past a failing gate." The verify gate is the same shape applied one boundary earlier.

**Consequence.** Factory must extend, not replace. Encoded as REQ-FM-007 and design §2.

## R-2 — Sync overlap is Phase 8 only

**Command.** `sed -n '40,90p' .claude/skills/moai/workflows/sync.md`; then `grep -n "Phase 8" -A25 .claude/skills/moai/workflows/sync/quality-gates-quality.md`

**Observed.** The Phase Routing Table lists Phase 7 Quality Check, Phase 8 Security Scan, Phase 9 MX Tag Validation, Phase 10 Coverage Analysis. Only Phase 8 is a security scan.

Two further properties matter and were not in the brief:

1. **Phase 8 is already conditional.** It runs only when changed files match security-sensitive patterns (auth, database, API endpoint, input handling, secret-bearing config); otherwise it logs "Security scan skipped: no security-sensitive files changed." So the dedup contract adds a *second* skip condition to a phase that already has one — the two must compose, not conflict.
2. **Phase 8 carries a separate always-on dependency-manifest audit** that runs "regardless of whether manifest files changed in this SPEC", auditing `go.mod`, `package.json`, and eight other manifests. This is explicitly described as "a SEPARATE, automatic mechanism distinct from the agent-invoked security analysis".

**Consequence.** The dedup skip must apply only to the agent-invoked security analysis, **not** to the dependency-manifest audit — a deep scan of source code is not a substitute for a transitive-dependency check. This distinction is recorded here and constrains M3's edit; it is not something a run-phase agent should have to rediscover.

## R-3 — `revision.json` has no Go producer

**Command.** `grep -rn "revision.json" --include=*.go internal/ cmd/` → zero matches. `grep -rn "security-deepscan\|revision.json" --include="*.md" .claude/skills/moai/workflows/` → exactly three hits, all in `review.md` (lines 360, 364, 390).

**Observed schema** (`review.md:393`):

```json
{"scanned_commit":"<sha>","effort_tier":"<low|medium|high|xhigh|max>","working_tree_included":true,"scope":"repo|branch|commit|staged|file","generated_at":"<injected-timestamp>"}
```

**Also observed.** Every results directory "ships its own `.gitignore` (ignoring its entire contents) so a stray `git add` never sweeps a scan into a commit." This means the artifact is deliberately untracked — a factory session cannot assume it survives a `/clear` plus a fresh checkout, only that it survives within the session's working tree. That reinforces the fail-safe default.

**Consequence.** R1 resolution: reader in Go, writer left to doctrine, predicate defaults to RUN. REQ-FM-014 through REQ-FM-018.

## R-4 — `--repo` contradiction, enumerated

**Command.** `grep -n "repo" .claude/skills/moai/workflows/review.md`

**Observed** — four relevant sites:

| Line | Section | Says |
|---|---|---|
| 48 | flag definition list | with `--deep`, scopes the deep scan to the whole tree |
| 221 | `--lean Mode` § Scope | honored **only** in `--lean` mode |
| 283 | `--deep Mode` preamble | `--deep` "reuses the same scope-selection machinery (`--repo` / `--staged` / `--branch B` / `--commit SHA`)" |
| 293 | `--deep Mode` job menu | "Scan the whole repository \| `--repo` \| The entire working tree" |

Three sites versus one. The three include the flag's own definition and two statements inside the section that defines `--deep` behavior. The one is a sentence in the `--lean` section's scope subsection, written for a mode that at the time was the only one with a repo/diff split.

**Consequence.** Lines 48/283/293 authoritative; line 221 corrected. REQ-FM-019, REQ-FM-020, AC-FM-021.

## R-5 — The block-cap inject is launch-time-only

**Command.** `cat internal/cli/launcher_blockcap_infinite.go`

**Observed.** `injectStopHookBlockCapForGoal` returns `base` unchanged unless `goal.LoadGoal(projectRoot, sessionID)` yields a goal with `Status == goal.StatusArmed` **and** `Ceiling.MaxTurns == 0`. The file's own comment frames the feature as making "arm and walk away" work — i.e. arm first, launch second.

Cross-checked against `run.md` § Run-phase Autonomy: the `ac_converge` goal is armed by the orchestrator *after* Implementation Kickoff Approval, which is mid-session.

**Consequence.** A mid-session-armed goal is invisible to the launch-time inject. Factory injects unconditionally. Recorded as R3 in `plan.md` §B and REQ-FM-023; this contradicts the brief's assumption that factory could simply "reuse this path" as-is, and is the single most consequential finding of this research pass.

Also observed, and relevant to the goal preset: `run.md` warns that a prose "stop after N turns" / "Max 20 turns" clause "is NOT parsed — `parseCondition` matches only a trailing `exits <N>`", so `ac_converge` actually runs at the default 30-turn ceiling. The `factory_chain` preset must bound itself with `--max-turns`, never with prose.

## R-6 — Manual flag parsing on the launchers

**Commands.** `sed -n '40,110p' internal/cli/cc.go`; `sed -n '795,880p' internal/cli/launcher.go`; `sed -n '45,72p' internal/cli/spawn.go`; `grep -n "DisableFlagParsing\|parseProfileFlag\|stripSpawnFlag\|normalizeWorktreeFlag\|unifiedLaunch" internal/cli/glm.go internal/cli/cg.go`

**Observed.** `cc.go`, `cg.go`, and `glm.go` all set `DisableFlagParsing: true`. `runCC` sequences: manual `--help`/`-h` scan → `stripSpawnFlag` → `parseProfileFlag` → `resolveWorktreeL2Path` → `normalizeWorktreeFlag` → `unifiedLaunch(profile, "claude", args)`. `runCG` is identical with mode `claude_glm`; `runGLM` adds manual subcommand routing and calls `unifiedLaunch(profile, "glm", args)`.

All three parsers break at `--` and append the remainder verbatim. `normalizeWorktreeFlag` has a "consume the next token unless it starts with `-`" branch, which is why ordering matters for the new parser (design §5).

`internal/cli/cc_test.go` swaps `unifiedLaunchFunc` to capture the forwarded args — the seam the new tests reuse.

**Consequence.** REQ-FM-002, REQ-FM-003, design §5.

## R-7 — `-f` availability

**Command.** `grep -rn '"-f"' internal/cli/` (narrowed to launcher files)

**Observed.** `-f` appears only in `internal/cli/doctor_config.go` and `internal/cli/state.go` as a `--format` shorthand. Neither is `cc`, `glm`, or `cg`.

**Consequence.** `-f` is free on the three launchers. REQ-FM-006, AC-FM-006 (which uses the pre-change grep as its positive control).

## R-8 — State directory precedent

**Command.** `ls .moai/state/` (primary checkout)

**Observed.** `goal/`, `handoff/`, `lsel/`, `verify/` — per-subsystem directories, session- or id-keyed JSON inside. Note: the *worktree's* `.moai/state/` contains only `.gitkeep`, because state is gitignored; the precedent is therefore a live-tree observation, not a tracked-file one.

**Consequence.** `.moai/state/factory/<session>.json` fits. REQ-FM-024.

## R-9 — Template mirror surface

**Command.** `ls internal/template/templates/.claude/skills/moai/workflows/` and its `sync/` subdirectory

**Observed.** `moai.md`, `run.md`, `sync.md`, `goal.md`, and the `sync/` sub-skills (`quality-gates-quality.md`, `quality-gates-context.md`, `doc-execution.md`, `delivery.md`) are all mirrored. So every file M1-M3 touches has a template counterpart that must move with it.

**Consequence.** REQ-FM-025, milestone M6.

## R-10 — `/moai security` stays retired

**Command.** `sed -n '285,300p' .claude/skills/moai/workflows/review.md`

**Observed, verbatim.** "it lives entirely under `/moai review --deep`; there is no separate top-level security subcommand (the former `/moai security` entry was retired and is not revived — the deep scan is a mode of `/moai review`, and no `security` / `audit` / `sec` alias is added)."

**Consequence.** The exclusion in `spec.md` §C is not a new prohibition; it restates an existing one, which is why "add a `/moai verify` subcommand" was rejected on two independent grounds (design §8).

## R-11 — the inject's call site, and why the flag cannot reach it (v0.2.0)

**Commands.** `grep -rn "injectStopHookBlockCapForGoal" internal/ cmd/`; `sed -n '775,795p' internal/cli/launcher.go`; `grep -n "^func " internal/cli/launcher.go`; `cat internal/cli/cc.go`.

**Observed.** The function has exactly TWO non-test references: its own definition at `internal/cli/launcher_blockcap_infinite.go:38`, and one production call at `internal/cli/launcher.go:790`. That call sits inside `launchClaudeDefault` (function span 617-791), on the line immediately after `launchEnv` is assigned from `os.Environ()` — `buildEnvForGLMLaunch(effectiveEffort, os.Environ())` at `:783` for the GLM branch, `buildEnvForLaunch(effectiveEffort, os.Environ())` at `:786` for the Claude branch.

The reachable chain from the command entry point is five hops: `runCC` (`cc.go`) → `unifiedLaunch` (`launcher.go:54`) → `unifiedLaunchFunc` = `unifiedLaunchDefault` (`:127`) → `launchClaude` (`:609`) → `launchClaudeFunc` = `launchClaudeDefault` (`:617`) → `:790`. The signature is `injectStopHookBlockCapForGoal(ctx context.Context, base []string, projectRoot, sessionID string) []string`; `unifiedLaunch(profileName, modeOverride string, extraArgs []string)` carries no factory-shaped parameter, and `extraArgs` is post-strip.

**Consequence.** The v0.1.0 claim that Factory "injects the cap unconditionally at launch" named no transport. Two mechanisms were possible and neither had been chosen. The environment route is viable precisely because of the `os.Environ()` seam observed above; the parameter route would touch five signatures plus two test seams. Encoded as REQ-FM-023 (environment route) with REQ-FM-024's variables becoming load-bearing as a stated consequence.

## R-12 — `revision.json` is not proof of completion; `findings.jsonl` is (v0.2.0)

**Command.** `sed -n '352,400p' .claude/skills/moai/workflows/review.md`

**Observed.** A deep-scan results directory contains `.gitignore`, `report.md`, `findings.jsonl`, `revision.json`, and — only under `--patch` — `patches/`. `findings.jsonl` is "machine-readable, exactly one finding per line (JSONL)". `revision.json` is described as "the revision stamp" and carries exactly `{scanned_commit, effort_tier, working_tree_included, scope, generated_at}` — no completion field, no status field, no rung field.

**Consequence.** A scan that aborted after stamping `revision.json` satisfies every conjunct of the v0.1.0 predicate. `findings.jsonl` is the discriminating artifact: a clean scan writes it with zero lines, an aborted scan does not write it. Added to the predicate as a completeness conjunct (REQ-FM-015) with a zero-line acceptance case asserted explicitly (AC-FM-019b), so a genuinely clean scan is not mistaken for an aborted one.

**Also observed.** The degradation ladder names three rungs — PRIMARY, FALLBACK, DEGRADED — and states that only DEGRADED reduces rigor ("single-pass, no per-finding 3-voter panel") and that "the report MUST label this rung as rigor-reduced". The rung lives in the report prose, not in `revision.json`. A predicate reading only `revision.json` therefore cannot exclude a DEGRADED scan, which is why the rung is carried on the factory state record (`verify_rung`) instead.

## R-13 — the `--repo` contradiction, quoted verbatim (v0.2.0)

**Command.** `grep -n 'only in' .claude/skills/moai/workflows/review.md`

**Observed, verbatim (line 221)** — reproduced inside a fenced block so the backtick placement is unambiguous:

```
- Repo-scope (with --repo): sweep the WHOLE tree. This is the "sweep everything" variant. The `--repo` flag is honored only in --lean mode.
```

**Consequence.** The trailing clause carries backticks around `--repo` but **not** around `--lean`. The v0.1.0 draft quoted it as "only in `--lean` mode" with backticks, which would have made the AC-FM-021 positive-control grep match zero lines and pass vacuously. The corrected pattern is `honored only in --lean mode`.

## R-14 — the template tree already carries 8 `internal/...` references (v0.3.0)

**Command.** `grep -rn 'internal/factory\|internal/cli' internal/template/templates/.claude/`

**Observed.** 8 matches, in 5 files, none of which this SPEC touches: `agents/moai/plan-auditor.md:210`, `rules/moai/core/askuser-protocol-reference.md:149`, `rules/moai/core/agent-hooks.md:69`, `rules/moai/workflow/worktree-integration.md:307,369,380`, `rules/moai/workflow/worktree-state-guard.md:2,15`.

**Command.** the same pattern against the mirrored counterparts of the acceptance §A.2 set.

**Observed.** 0 matches in each of `skills/moai/workflows/{moai,run,review}.md`, `skills/moai/workflows/sync/quality-gates-quality.md`, and `rules/moai/workflow/goal-directive.md`; `skills/moai/workflows/factory.md` does not yet exist.

**Consequence.** The v0.2.0 AC-FM-025 ran the tree-wide form and expected no output, so it failed on every run regardless of this SPEC's cleanliness — and the obvious remedy (editing five unrelated shipped files) is a scope violation with its own mirror-parity blast radius. REQ-FM-025 was already correctly scoped to the files this SPEC mirrors; the criterion is bounded to `$MIRRORS` to match, with the measured 0 recorded as the baseline. The 8 pre-existing references are listed as out of scope in `spec.md` §C.

## R-15 — gate-token baselines across the doctrine set (v0.3.0)

**Command (v0.4.0 form — quoted array, widened token set).** The v0.3.0 form used an unquoted scalar `DOCS`, which `zsh` does not word-split: it ran the loop once against the whole six-path string and printed a single `0`. Corrected together with the `acceptance.md` §A.2 and `plan.md` §C copies:

```bash
for f in "${DOCS[@]}"; do
  if [ ! -f "$f" ]; then printf '%s ABSENT\n' "$f"; continue; fi
  printf '%s ' "$f"
  grep -o 'HUMAN GATE\|gate-sync-1\|gate-sync-2\|Implementation Kickoff Approval\|AskUserQuestion' "$f" | wc -l
done
```

**Observed (four-token set, v0.3.0 — re-measured at v0.4.0 and unchanged).** `moai.md` 12 (`gate-sync-1` ×1, `gate-sync-2` ×1, `HUMAN GATE` ×3, `Implementation Kickoff Approval` ×7); `run.md` 23 (`HUMAN GATE` ×1, `Implementation Kickoff Approval` ×22); `review.md` 0; `sync/quality-gates-quality.md` 0; `goal-directive.md` 12 (`Implementation Kickoff Approval` ×12); `factory.md` absent.

**Observed (five-token set, v0.4.0 — the pattern AC-FM-012 now uses).** `moai.md` 23; `run.md` 33; `review.md` 3; `factory.md` absent; `sync/quality-gates-quality.md` 2; `goal-directive.md` 18. (A seventh file, `run/mode-orchestration.md`, joined the §A.2 doctrine set at v0.7.0 when the verify exit gate relocated there; its pre-relocation baseline measures **1** under this same pattern and **0** under the four-token pattern above. See `acceptance.md` §A.2.1.) The deltas are the `AskUserQuestion` occurrences. The token was added because the four-token set was gate-shaped only by accident: a fifth gate introduced as plain prose, or as an `AskUserQuestion` round carrying no gate label, changed no count and passed.

**Consequence.** The v0.2.0 AC-FM-012 asked a human to classify roughly 47 context-free fragments and could not have distinguished a newly added `HUMAN GATE` token from the pre-existing ones — the exact failure the criterion existed to catch. Converted to a per-file baseline-delta assertion with a planned `N` per file, and the baseline capture promoted to `plan.md` §C item 5 because M1 destroys it. Note also that `gate-sync-1` already appears once in `moai.md`, so AC-FM-007's second grep (`>= 1`) passed before any edit; it now asserts a delta from the measured baseline of 1.

**Also observed.** `sync.md` § HUMAN GATE Map lists `gate-sync-1` as a human gate, so a factory chain fires four gates, not three. REQ-FM-012's count label is corrected; the substance — this SPEC adds exactly one — is unchanged.

## R-16 — `goal-directive.md` states a launch-time-only condition, and its mirror is byte-identical (v0.3.0)

**Command.** `sed -n '15p' .claude/rules/moai/workflow/goal-directive.md` and `cmp .claude/rules/moai/workflow/goal-directive.md internal/template/templates/.claude/rules/moai/workflow/goal-directive.md`

**Observed.** The § Raising the block cap for an infinite goal paragraph states: "The `moai cc` / `moai cg` launchers do this automatically: when an armed `--max-turns 0` goal exists for the resolving session **at launch time**, the launcher injects `CLAUDE_CODE_STOP_HOOK_BLOCK_CAP=200` into the session env (best-effort, fail-open)." `cmp` reports the live file and its template mirror byte-identical. `grep -c 'Factory Mode'` returns 0 in both.

**Consequence.** REQ-FM-023 adds a second, unconditional trigger to the same function, so after implementation that sentence is false on both trees — and it is precisely the sentence a user consults when a session unexpectedly runs 200 blocks. In v0.2.0 the file appeared only as a `plan.md` §I cross-reference (`grep -c 'goal-directive' acceptance.md` = 0), in no requirement and no milestone. Added as REQ-FM-028, AC-FM-026, acceptance §A.2, and M1's edit list. The byte-identical mirror means the amendment must be applied to both trees in the same milestone, and `cmp` remains the check.

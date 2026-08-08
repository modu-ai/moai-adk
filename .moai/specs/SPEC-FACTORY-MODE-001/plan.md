# SPEC-FACTORY-MODE-001 — Implementation Plan

Version: 0.8.0 | Tier: L | Status: draft | 26 requirements · 36 acceptance-criterion leaves (over the Tier L ceiling of 25 — declared, see `spec.md` §D)

Milestones below are ordered by **decision-reversibility**: the contract and gate semantics a reviewer is most likely to want changed come first, the data-model decision (the state-record schema) follows, and the mechanical launcher plumbing comes last. Each milestone declares its hard dependencies explicitly — reversibility governs the ordering, but a dependency overrides it where the two conflict (v0.2.0: the state-record schema moved ahead of the dedup gate that reads it).

## §A Context — verified ground truth

Every claim in this section was established by direct inspection of the tree at `feat/factory-mode`. File:line references are stable content anchors, not guesses.

| Fact | Source |
|---|---|
| `full-pipeline` contract: run-phase completion auto-chains into sync with no extra approval at that boundary; `gate-sync-1` / `gate-sync-2` still fire inside the chained sync | `.claude/skills/moai/workflows/moai.md` § run→sync chaining policy |
| `single-phase` contract for an explicit `/moai run`: the sync chain is surfaced as the "(Recommended)" first option, never fired silently | `.claude/skills/moai/workflows/run.md` § Chaining |
| Sync overlap with verify is Phase 8 (Security Scan) only; Phases 7 / 9 / 10 are Quality / MX / Coverage and do not duplicate a deep scan | `.claude/skills/moai/workflows/sync.md` § Phase Routing Table |
| Phase 8 is itself conditional — it runs only when changed files match security-sensitive patterns — and its Step 0.55.1 agent analysis is distinct from an always-on dependency-manifest audit | `.claude/skills/moai/workflows/sync/quality-gates-quality.md` § Phase 8 |
| The dependency-manifest audit runs "regardless of whether manifest files changed in this SPEC" and is "a SEPARATE, automatic mechanism distinct from the agent-invoked security analysis", covering ten manifests | same file, § Phase 8 dependency manifest audit |
| Human gates: `gate-sync-1` (Pre-Sync Quality), `gate-sync-2` (Documentation Scope) | `sync.md` § HUMAN GATE Map |
| Implementation Kickoff Approval is score-independent and mandatory; `ac_converge` is armed only after it | `run.md` § Run-phase Autonomy |
| A trailing "stop after N turns" clause is **not parsed**; only `--max-turns N` binds | `run.md` § Run-phase Autonomy, actual-turn-ceiling note |
| `CLAUDE_CODE_STOP_HOOK_BLOCK_CAP=200` inject exists, but fires **only when an armed `--max-turns 0` goal already exists at launch time** | `internal/cli/launcher_blockcap_infinite.go` — `injectStopHookBlockCapForGoal` |
| That inject has exactly ONE production caller: `internal/cli/launcher.go:790`, inside `launchClaudeDefault` | repo-wide grep for `injectStopHookBlockCapForGoal` (2 non-test hits: the definition and that call) |
| `launchEnv` at that call site is built from `os.Environ()` immediately above, at `launcher.go:783` (GLM branch) / `:786` (Claude branch) | `internal/cli/launcher.go` |
| The launch chain is `runCC → unifiedLaunch → unifiedLaunchFunc(=unifiedLaunchDefault) → launchClaude → launchClaudeFunc(=launchClaudeDefault)` — five hops, none carrying a factory parameter | `internal/cli/cc.go`, `internal/cli/launcher.go:54/127/609/617` |
| The raised-cap constant is `DefaultRaisedStopHookBlockCap = 200`; the env key constant is `config.EnvClaudeCodeStopHookBlockCap` | same file; `internal/config/envkeys.go` |
| `moai cc` and `moai cg` both set `DisableFlagParsing: true` and parse args manually via `stripSpawnFlag` → `parseProfileFlag` → `resolveWorktreeL2Path` → `normalizeWorktreeFlag` → `unifiedLaunch` | `internal/cli/cc.go`, `internal/cli/cg.go` |
| `moai glm` follows the same manual sequence plus manual subcommand routing | `internal/cli/glm.go` |
| Every manual parser stops at the `--` pass-through marker and forwards the remainder verbatim | `launcher.go` `parseProfileFlag` / `normalizeWorktreeFlag`; `spawn.go` `stripSpawnFlag` |
| `-f` is bound only by `doctor config dump` and `state` as a `--format` shorthand — different commands, no collision on cc / glm / cg | `internal/cli/doctor_config.go`, `internal/cli/state.go` |
| `.moai/state/` already carries per-subsystem dirs (`goal/`, `handoff/`, `lsel/`, `verify/`), so `factory/` fits the convention | primary checkout `.moai/state/` |
| A deep-scan results directory contains `.gitignore`, `report.md`, `findings.jsonl`, `revision.json`, and (with `--patch`) `patches/` | `review.md` § Results directory |
| `revision.json` schema is `{scanned_commit, effort_tier, working_tree_included, scope, generated_at}` — it carries **no** completion or status field | `review.md` § revision.json |
| `revision.json` has **zero Go writers** (repo-wide grep over `internal/` and `cmd/` returns only doc lines) | `review.md` lines 360 / 364 / 390 |
| The DEGRADED rung is "single-pass, no per-finding 3-voter panel" and self-labels as rigor-reduced; the rung is NOT recorded in `revision.json` | `review.md` § degradation ladder |
| Workflow skills are mirrored under `internal/template/templates/.claude/skills/moai/workflows/`, including the `sync/` sub-skill directory | template tree listing |

## §B Known issues resolved by this plan

### R1 — `revision.json` has no producer, and is not proof of completion

**Finding.** The artifact is specified at `review.md:364` / `:390` with the schema `{scanned_commit, effort_tier, working_tree_included, scope, generated_at}` under `.moai/reports/security-deepscan-<timestamp>/`. A repo-wide grep for the literal `revision.json` across `internal/` and `cmd/` returns nothing: it is produced, if at all, by workflow agents writing files.

**Finding (new in v0.2.0).** The schema carries no completion or status field. An aborted scan that stamped `revision.json` is therefore indistinguishable — from that file alone — from a completed one. `findings.jsonl` is the completion evidence: a clean scan writes it with zero lines, while an aborted scan characteristically never writes it.

**Resolution (settled by the user, encoded here).** Keep the dedup contract in v1, with an explicit fail-safe fallback. The predicate defaults to RUN, never to SKIP:

```
revision_match :=
      results directory exists at record.deepscan_dir
  AND findings.jsonl exists AND every line parses as JSON
  AND revision.json exists AND readable AND parses as JSON
  AND revision.scanned_commit == headSHA      (git rev-parse HEAD at sync entry)
  AND revision.scope == "repo"
  AND ( NOT treeDirty  OR  revision.working_tree_included == true )
```

Every conjunct that fails yields FALSE, and FALSE means Phase 8 Step 0.55.1 runs. Absence is FALSE. Malformed JSON is FALSE. A commit mismatch is FALSE. A missing `findings.jsonl` is FALSE.

**Input derivation (new in v0.2.0).** Both runtime inputs are derived, never judged:

- `headSHA` — stdout of `git rev-parse HEAD` at sync entry.
- `treeDirty` — `git status --porcelain` at sync entry over tracked and untracked paths, excluding `.moai/state/`, `.moai/reports/`, `.moai/cache/`, `.moai/logs/` (the `worktree-state-guard.md` § Divergence Threshold exclusion set), and excluding the results directory itself. Dirty when at least one line survives the exclusion.

**The dirty-tree sub-decision.** `working_tree_included: false` combined with a dirty tree at sync entry **invalidates the match**. Rationale: a scan that excluded the working tree did not examine the very edits sync is about to document. Treating that as a match would mean a scan is skipped precisely where the unscanned surface is largest. When the tree is clean, `working_tree_included` is irrelevant and does not participate in the predicate.

**Ownership split.** The predicate is a Go **reader** under `internal/factory` — pure, deterministic, unit-testable against fixtures in `t.TempDir()`. The **writer** stays where it is: the `--deep` workflow agents. This SPEC adds no Go writer and makes no claim that one exists.

**Consumption (new in v0.2.0).** A reader nothing calls is dead code, and every unit test over a pure function passes regardless. The sync Phase 8 dedup gate doctrine therefore names the predicate call and both derived inputs explicitly (REQ-FM-017, AC-FM-020d), so the gate is falsifiable at the doctrine layer as well as the unit layer.

### R2 — `--repo` is self-contradictory in `review.md`

**Finding.** Four sites speak to `--repo`:

| Line | Statement | Section |
|---|---|---|
| 48 | "With `--lean`, sweeps the WHOLE tree… **With `--deep`, scopes the deep scan to the whole tree.** Ignored without `--lean` or `--deep`." | flag definition list |
| 221 | verbatim: `The \`--repo\` flag is honored only in --lean mode.` (note: **no** backticks around `--lean`) | inside `--lean Mode` § Scope |
| 283 | `--deep` "reuses the same scope-selection machinery (`--repo` / `--staged` / `--branch B` / `--commit SHA`)" | `--deep Mode` preamble |
| 293 | job-menu table row: "Scan the whole repository \| `--repo` \| The entire working tree" | `--deep Mode` job menu |

**Resolution.** Lines 48, 283, and 293 are authoritative — three independent sites, two inside the section that defines `--deep` behavior. Line 221 is a single section-local sentence written when `--lean` was the only mode with a repo/diff split, and it was not revisited when `--deep` adopted the same machinery. It is stale, and this SPEC corrects it (REQ-FM-019) rather than routing around it.

The v0.1.0 draft mis-quoted line 221 as containing backticks around `--lean`, which would have made the AC-FM-021 positive-control grep return zero. The verbatim form above was re-read from the live file and is what the grep must match.

The settled verify call `/moai review --security --deep --repo` therefore stands unchanged.

### R3 — the block-cap inject cannot see a mid-session goal, and cannot see a factory flag either

**Finding (part 1).** `injectStopHookBlockCapForGoal` raises the cap only when `goal.LoadGoal(projectRoot, sessionID)` returns an **armed** goal with `Ceiling.MaxTurns == 0` **at launch time**. The `factory_chain` goal is armed mid-session, after Implementation Kickoff Approval. The existing conditional path would therefore never fire for a factory session, and the chain would die at the runtime default of 8 consecutive Stop-hook blocks.

**Finding (part 2, new in v0.2.0).** The function has exactly ONE production caller — `internal/cli/launcher.go:790`, inside `launchClaudeDefault` — reached through five hops from `runCC`, none of which carries a factory parameter. Its signature is `(ctx context.Context, base []string, projectRoot, sessionID string)`. `unifiedLaunch(profile, mode, args)` receives `args` from which REQ-FM-001 has already stripped the `--factory` token, so the flag is simply not in scope at the call site. The v0.1.0 draft asserted the inject would fire for a factory launch without naming how the signal would arrive.

**Resolution — the environment route.** `runCC` / `runGLM` set `MOAI_FACTORY` (and `MOAI_FACTORY_SPEC`) in the process environment before calling `unifiedLaunch`; `injectStopHookBlockCapForGoal` reads `os.Getenv(config.EnvMoaiFactory)` and takes its unconditional branch when set. This works without any signature change because `launchEnv` is built from `os.Environ()` at `launcher.go:783` / `:786`, immediately above the call — so the variable is already in the slice the inject mutates, and it also reaches the child process, which the orchestrator needs anyway.

**Rejected alternative — thread a parameter.** Would require changing `unifiedLaunch`, the `unifiedLaunchFunc` variable type, `launchClaude`, and the `launchClaudeFunc` variable type, plus both test seams (`cc_test.go`, `cg_test.go`), for a signal that the environment already carries. Rejected on blast radius.

**Consequence, stated explicitly.** REQ-FM-024's environment variables become load-bearing for REQ-FM-023. Removing or renaming them silently disables the cap raise, and the failure is quiet — the session simply stops after 8 blocks with no error. AC-FM-023c exists to catch exactly that.

**Blast radius, accepted.** The raised cap is session-wide, not scoped to the `factory_chain` goal. A user who declines at Implementation Kickoff Approval, or arms an unrelated goal later in the same session, inherits a 200-block ceiling. This is not a gate bypass — all four human gates are orchestrator-issued `AskUserQuestion` rounds, not Stop-hook blocks — but it is documented in `factory.md` so an operator can reason about an unexpectedly long unrelated loop. Scoping the cap per-goal would need a runtime mechanism Claude Code does not expose; it is listed as out of scope.

### R4 — the verify gate failed open (new in v0.2.0)

**Finding.** The v0.1.0 requirements partitioned verify outcomes into three cases — confirmed CRITICAL/HIGH, MEDIUM/LOW-or-none, and DEGRADED — with no case for "verify produced no result". An errored invocation, an aborted pipeline, or a results directory that was never written all read as "no confirmed findings", and the MEDIUM/LOW-or-none rule routed them to sync. The gate that exists to keep unscanned changes out of sync would have passed them through.

**Resolution.** A **3-case severity partition** — **S1** readable + confirmed CRITICAL/HIGH (REQ-FM-009); **S2** readable + MEDIUM/LOW-or-none (REQ-FM-011); **S3** no readable result (REQ-FM-026) — crossed with an **orthogonal rung attribute** (`PRIMARY` / `FALLBACK` / `DEGRADED`, REQ-FM-013) carried by every S1 and S2 result; S3 produced no result and carries no rung. Precedence is stated explicitly: the severity case governs routing and the rung never changes it; the rung governs Phase 8 suppression and the severity case never relaxes it. The no-result case HALTs, does not count against the two-re-entry ceiling, and emits the 5-section verdict. The MEDIUM/LOW branch is guarded on the word *readable*, so it cannot absorb the no-result case by default. The v0.2.0 "four disjoint, jointly exhaustive outcomes" framing is **withdrawn** — see §G AP-15.

### R5 — the dedup skip was unqualified in two directions (new in v0.2.0)

**Finding (a).** REQ-FM-014 said "sync Phase 8 shall be skipped", but Phase 8 contains two mechanisms: the agent-invoked security analysis (Step 0.55.1) and a dependency-manifest audit that runs unconditionally across ten manifests to catch transitive-vulnerability drift unrelated to the current SPEC. A source-code deep scan does not substitute for a dependency audit. `research.md` R-2 recorded this constraint; it never reached the requirement layer.

**Finding (b).** `revision.json` carries `effort_tier`, and a DEGRADED-rung run (single-pass, no 3-voter panel) satisfies every conjunct of the predicate. Such a run would both lose its own adversarial verification AND suppress the independent Phase 8 analysis — the rigor-reduced scan silently replacing the rigorous one. REQ-FM-013 required only that the label be *surfaced*; nothing forbade proceeding.

**Resolution (a).** The skip is scoped to Step 0.55.1 by name. The manifest audit runs unconditionally under the `factory` contract, and AC-FM-020b asserts the pre-existing always-on sentence is still present and unmodified.

**Resolution (b).** The rung is not machine-readable from `revision.json` — the schema has `effort_tier` (an effort level, not a rung) and no rigor field, and extending an artifact schema this SPEC does not own is out of scope. Instead the orchestrator records the transcript-surfaced rung into `verify_rung` on the factory state record (a record this SPEC does own), and REQ-FM-014 excludes `DEGRADED` from suppression. `effort_tier` was deliberately NOT added to the predicate: an effort floor is not equivalent to a rung, and conflating them would produce a predicate that reads correct and is not.

## §C Pre-flight

1. Confirm the worktree is on `feat/factory-mode` and synced with `origin/main`.
2. `go test ./internal/cli/... ./internal/config/...` green before the first edit (baseline).
3. `moai spec lint .moai/specs/SPEC-FACTORY-MODE-001/spec.md` reports zero errors on the draft.
4. Capture the pre-change positive-control greps that AC-FM-007 / AC-FM-008 / AC-FM-020 / AC-FM-021 / AC-FM-023b / AC-FM-025 / AC-FM-026 depend on. Each must be recorded **before** the corresponding milestone edits the file, because a post-change run cannot reconstruct a pre-change baseline.
5. Capture the **AC-FM-012 per-file gate-token baseline** across the §A.2 doctrine set. This is the criterion for REQ-FM-012, and a baseline-delta check is only as good as the baseline; running it after M1 has edited `moai.md` and `run.md` destroys the evidence permanently.
6. Measure the **remaining headroom of every capped file a milestone will add lines to** (added v0.7.0 — the fifth shape of **AP-16** in §G). `internal/skills/workflow_split_test.go` caps entry routers at 200 LOC and sub-skills at 600; `wc -l` each target and subtract. A milestone that plans to add lines to a file already at its ceiling is unachievable as written, and the failure surfaces only at CI, after the criteria have already been authored against the wrong file. `run.md` stood at exactly 200/200 at `7171880a9`.

`DOCS` and `MIRRORS` MUST be arrays with quoted expansions. The project shell is `zsh`, which does not word-split an unquoted scalar: the scalar form runs the loop once against the whole six-path string and turns `MIRRORS` into a single non-existent path, so the mirror grep errors to stderr and emits no stdout — vacuously satisfying an absence check. See `acceptance.md` §A.2 for the full statement of the failure mode.

```bash
DOCS=(
  .claude/skills/moai/workflows/moai.md
  .claude/skills/moai/workflows/run.md
  .claude/skills/moai/workflows/run/mode-orchestration.md
  .claude/skills/moai/workflows/review.md
  .claude/skills/moai/workflows/factory.md
  .claude/skills/moai/workflows/sync/quality-gates-quality.md
  .claude/rules/moai/workflow/goal-directive.md
)

for f in "${DOCS[@]}"; do
  if [ ! -f "$f" ]; then printf '%s ABSENT\n' "$f"; continue; fi
  printf '%s ' "$f"
  grep -o 'HUMAN GATE\|gate-sync-1\|gate-sync-2\|Implementation Kickoff Approval\|AskUserQuestion' "$f" | wc -l
done

# also, in the same pre-flight pass:
grep -c 'gate-sync-1' .claude/skills/moai/workflows/moai.md          # AC-FM-007 delta baseline
grep -c 'Factory Mode' .claude/rules/moai/workflow/goal-directive.md  # AC-FM-026 baseline
MIRRORS=()
for f in "${DOCS[@]}"; do MIRRORS+=("internal/template/templates/$f"); done
for m in "${MIRRORS[@]}"; do [ -f "$m" ] || printf 'MIRROR-ABSENT %s\n' "$m"; done
grep -n 'internal/factory\|internal/cli' "${MIRRORS[@]}"              # AC-FM-025 bounded baseline
```

Baselines measured at authoring time with the widened five-token pattern, for comparison against the pre-flight run: `moai.md` 23, `run.md` 33, `run/mode-orchestration.md` 1 (measured pre-relocation at `7171880a9`; the file joined the §A.2 set at v0.7.0 — see `acceptance.md` §A.2.1), `review.md` 3, `factory.md` absent, `sync/quality-gates-quality.md` 2, `goal-directive.md` 18. (The former four-token figures — 12 / 23 / 0 / absent / 0 / 12 — were re-measured and confirmed unchanged; `AskUserQuestion` is what the widened pattern adds. See `acceptance.md` AC-FM-012 for why the token set widened.) Also: `gate-sync-1` in `moai.md` = 1; `Factory Mode` in `goal-directive.md` = 0; the bounded mirror grep returns 0 matches over the five mirrors that exist, with `factory.md`'s mirror reported `MIRROR-ABSENT` pre-change. A divergence at implementation time means the tree moved — use the freshly measured value, not the recorded one.

## §D Constraints

- **Template-First, same-milestone.** Every `.claude/` edit is mirrored under `internal/template/templates/.claude/` **within the milestone that made it**, followed by `make build`. Mirrored content carries no SPEC ID, requirement token, acceptance token, internal date, commit SHA, or internal Go package path. Deferring mirrors to a later milestone would push intermediate commits whose `.claude/` changes have no mirror, and the mirror-parity and neutrality CI guards fire on push — so a deferred mirror is a red build on every intermediate commit, not merely untidy. M6 verifies the mirrors; it does not create them.
- **No new runtime.** No new hook, evaluator, subcommand, or daemon.
- **Manual-parser discipline.** `--factory` / `-f` handling obeys the `--` boundary exactly as the three existing parsers do.
- **Test isolation.** `t.TempDir()` for every fixture; no `t.Setenv` on OTEL variables; the `MOAI_FACTORY` `t.Setenv` tests are non-parallel; launcher tests live beside the code.
- **Coverage.** `internal/factory` at 85% or above (absolute floor, already met at 90.9%). `internal/cli` is a **non-regression against its measured pre-SPEC baseline** — floor `76.3%`, the lower of two independent observations (76.4% / 76.3%) at commit `7171880a9` — *plus* 100% per-function coverage on every function this SPEC introduces or extends. The v0.4.0 line read "90% for `internal/cli`", a figure never measured against the package's real starting point and ~13.6pp above it, so it was unsatisfiable from this SPEC's own scope. That is **AP-16** below; see `acceptance.md` AC-FM-025 § Coverage clause for the runnable form, the jitter rationale, and the two pre-existing `internal/cli` defects recorded there (a `--- FAIL` and a hang), both of which perturb the statement set a partial run executes — which is why the coverage figures are measured on a package-scoped run, never inferred from a full-suite run.
- **Gate preservation.** Four human gates fire across a factory chain (Implementation Kickoff Approval, the verify CRITICAL/HIGH decision added by this SPEC, `gate-sync-1`, `gate-sync-2`); no fifth, no bypass.
- **Fail-closed direction.** Every ambiguity in the verify gate and the dedup predicate resolves toward *run the check*, never toward *skip it*.

## §E Self-verification

- `go test ./...` is clean **except for two scope-bounded pre-existing defects**, each named individually: defect 1 — `internal/cli` / `TestRunHarnessObserveStop_ProposeChainAutoRuns` (a `--- FAIL`, root cause unidentified); defect 2 — real network I/O in `internal/statusline`'s usage collector (`fetchUsageFromOAuthAPI`, `usage.go:572`), which HANGS every test reaching `defaultBuilder.Build` and so kills both `internal/statusline` and `internal/cli` on a timeout panic. Judged by the **hang-aware, package-level** 3-step procedure in `acceptance.md` AC-FM-025 § Scope bound (v0.8.0): STEP 1 gates on `grep -E '^FAIL[[:space:]]'` package lines (a `--- FAIL` grep cannot see a hang), STEP 2 confirms defect identity inside each reported package by isolated re-run, STEP 3 resolves the timeout confound before any verdict is recorded. A third failing package, a `--- FAIL` naming any other test, or a timeout whose dump lacks the `fetchUsageFromOAuthAPI` frame is a FAIL. Both exclusions rest on provenance (both reproduce at the pre-SPEC baseline `7171880a9`; `git diff --name-only 7171880a9..HEAD -- internal/` touches neither code path), not on a root-cause diagnosis. A third defect, `internal/skills` / `TestEntryRouterLOCCeiling`, WAS ours and was repaired in `e9aa2c363` — it is deliberately absent from the exclusion list so a future regression of it still fails.
- `golangci-lint run` clean.
- The three-part coverage clause of `acceptance.md` AC-FM-025 holds: `internal/cli` at or above its measured 76.3% baseline floor, `internal/factory` at or above 85%, and each of the seven functions this SPEC introduces or extends at 100.0%.
- `make build` succeeds and the resulting diff includes the regenerated embedded catalog.
- Template-neutrality CI guard passes on the mirrored files, including the internal-package-path clause.
- The clarification-marker sweep below returns no matches. The search token is assembled at runtime from two fragments so that neither this plan nor the acceptance criteria can satisfy their own check — a literal pattern spelled out in prose matches itself, which is a self-trip, not a green:

```bash
T='[NEEDS CLA'; T="${T}RIFICATION:"
grep -rnF "$T" .moai/specs/SPEC-FACTORY-MODE-001/
# expected: no output (exit 1)
```

## §F Milestones

### M1 — Pipeline contract and the verify exit gate (doctrine)

*Depends on: nothing. Highest reversibility — this defines the user-visible flow and the gate placement.*

- Add a `factory` contract clause to `moai.md` § run→sync chaining policy, stated as an **extension of `full-pipeline`**: inherits the run→sync auto-chain and the `gate-sync-1` / `gate-sync-2` preservation clause; adds a plan-phase chain head and the verify exit gate. Explicitly state that it defines no second chaining mechanism.
- Add the verify exit gate to the run-phase doctrine: position (after AC convergence, before the auto-chain fires), the invocation `/moai review --security --deep --repo`, and the **3-case severity partition crossed with an orthogonal rung attribute** — **S1** readable + confirmed CRITICAL/HIGH → re-entry scoped to the changed surface (the one human gate this SPEC adds); **S2** readable + MEDIUM/LOW-or-none → proceed with inherited evidence, at every rung; **S3** no readable result → HALT (does not count against the ceiling). Then, separately, the rung attribute (`PRIMARY` / `FALLBACK` / `DEGRADED`) recorded on every S1/S2 result, with `DEGRADED` forcing Phase 8 suppression OFF and surfacing in the transcript and sync report. State the precedence explicitly: the severity case governs routing and the rung never changes it; the rung governs suppression and the severity case never relaxes it. Do **not** present the rung as a fourth peer outcome and do **not** call the four "disjoint" — that was the v0.2.0 defect. Include the 2-re-entry ceiling and the 5-section halt verdict.
  - **Landing file (corrected v0.7.0).** As authored, M1 placed this block inline in `run.md`. `run.md` is an entry router capped at 200 LOC by `internal/skills/workflow_split_test.go` `TestEntryRouterLOCCeiling`, and its pre-SPEC baseline was exactly 200 — so the milestone as written was structurally unachievable in that file. The block's landing file is `.claude/skills/moai/workflows/run/mode-orchestration.md` (the existing run-phase orchestration sub-skill); `run.md` carries only an in-place extension of its existing Phase Routing Table description cell, net-zero lines. See `acceptance.md` §A.2.1 for the relocation record and §G **AP-16** for the authoring defect this exposed.
- Correct `review.md` line 221 per REQ-FM-019, matching the verbatim pre-change text recorded in §B R2.
- Amend `.claude/rules/moai/workflow/goal-directive.md` § Raising the block cap for an infinite goal per REQ-FM-028: the launch-time-only clause becomes a two-trigger statement (armed `--max-turns 0` goal at launch time **or** Factory Mode). Apply the identical text to the template mirror — the two files are byte-identical, and `cmp` must still report so afterward. This is the sentence a user reads when a session unexpectedly runs 200 blocks; REQ-FM-023 falsifies it, so it cannot be left to a follow-up.
- Author `.claude/skills/moai/workflows/factory.md`: the contract, the four stages, the human gates (one added by this SPEC, three inherited — four fire in total), the exclusion of `moai cg`, the state-record shape (including `verify_rung` and its allow-list semantics), and the session-wide block-cap blast radius.
- Mirror all six edited/new `.claude/` files into `internal/template/templates/.claude/`, stripped of internal tokens; `make build`. (Five as originally planned, plus `run/mode-orchestration.md` — the verify gate's corrected landing file per v0.7.0.)

### M2 — The `factory_chain` goal preset

*Depends on: M1 (writes into `factory.md`).*

- Author the preset condition in `factory.md`, written entirely as **model conditions** (each predicate references a line the orchestrator surfaces in the transcript), mirroring the `ac_converge` authoring style.
- Encode the arming rules: armed only after Implementation Kickoff Approval; armed alongside the work, never in place of it; bound by the settled flags `--max-turns 0 --max-duration 14400`, never by a prose "stop after N turns" clause, which is not parsed. State the accepted risk: an unattended run may consume up to four hours of tokens.
- Encode the semantic-failure escalation and graceful-degradation clauses by reference to the canonical rules rather than restating them.
- Mirror `factory.md`; `make build`.

### M3 — Signal propagation and the state record

*Depends on: nothing in this SPEC (data-model decision, promoted ahead of M4 in v0.2.0 because M4's dedup gate reads `verify_rung` from this schema).*

- Add `EnvMoaiFactory = "MOAI_FACTORY"` and `EnvMoaiFactorySpec = "MOAI_FACTORY_SPEC"` to `internal/config/envkeys.go`.
- Implement `internal/factory/record.go`: the `.moai/state/factory/<session>.json` schema (`session_id`, `spec_id`, `backend`, `entered_at`, `deepscan_dir`, `verify_rung`, `verify_reentries`), plus write and read helpers. Best-effort and fail-open — a record write failure never blocks a launch.
- Tests: round-trip write/read, unwritable state directory (fail-open), and a `verify_rung` round-trip covering `PRIMARY` / `FALLBACK` / `DEGRADED`.

### M4 — Dedup contract and the `revision.json` reader

*Depends on: M1 (`factory.md` cross-reference) and M3 (`verify_rung` field the exclusion reads).*

- Add the Phase 8 dedup gate to `sync/quality-gates-quality.md`: the predicate, the fail-safe default, the REQ-FM-018 disclosure requirement, the Step-0.55.1-only skip scope with the manifest audit explicitly unaffected, and the DEGRADED-rung exclusion. The procedure names the predicate call and both derived inputs (`git rev-parse HEAD`, `git status --porcelain`) so the gate is falsifiable at the doctrine layer.
- Implement `internal/factory/revision.go`: `type Revision struct{...}`, `LoadRevision(path) (*Revision, error)`, and `Matches(rev *Revision, headSHA string, treeDirty bool) bool` as a pure predicate, plus the `findings.jsonl` completeness check.
- Tests: absent directory, absent `revision.json`, unreadable file, malformed JSON, commit mismatch, `scope != "repo"`, dirty tree with `working_tree_included: false`, absent `findings.jsonl`, zero-line `findings.jsonl` (must PASS), a `findings.jsonl` line that does not parse, `verify_rung: "DEGRADED"` composed decision, and the single happy path. Every negative case asserts FALSE.
- Mirror `sync/quality-gates-quality.md` with the predicate stated **behaviorally** — what it evaluates and which direction it fails — and no `internal/factory` path, which is meaningless in a distributed user project; `make build`.

### M5 — Launcher entry

*Depends on: M3 (`config.EnvMoaiFactory`, `internal/factory/record.go`).*

- Add `parseFactoryFlag(args) (spec string, enabled bool, rest []string)` in a new `internal/cli/factory.go`, mirroring the `--` boundary discipline of `stripSpawnFlag`.
- Wire it into `runCC` and `runGLM` ahead of `unifiedLaunch`; set `MOAI_FACTORY` / `MOAI_FACTORY_SPEC` in the process environment there **with a `defer`ed restore of the prior value and prior presence** (`os.LookupEnv` before, `os.Setenv` or `os.Unsetenv` in the defer — restoring absence, not `""`), so the mutation does not leak across the rest of the process or across tests in the `internal/cli` binary; write the state record; emit `FACTORY_MODE_UNSUPPORTED_BACKEND` from `runCG`.
- Extend `injectStopHookBlockCapForGoal` with an unconditional factory branch gated on `os.Getenv(config.EnvMoaiFactory)`, ahead of the existing goal read, reusing `DefaultRaisedStopHookBlockCap` and `config.EnvClaudeCodeStopHookBlockCap`. The pre-existing goal-conditional branch is preserved verbatim for non-factory launches.
- Tests in `cc_test.go`, `glm_test.go`, `cg_test.go`, and `launcher_blockcap_infinite_test.go`, using the existing `unifiedLaunchFunc` seam; plus the `os.Environ()`-derived child-environment assertion of AC-FM-023c.

### M6 — Full verification

*Depends on: M1-M5. Verification only — mirrors were created in the milestone that made each edit (§D).*

- Confirm every `.claude/` file touched by M1-M4 has a mirrored counterpart, and that `make build` output (the regenerated embedded catalog) is committed.
- Run the full suite, the linter, the coverage check, and the template-neutrality guard including the internal-package-path clause.
- Run every grep-anchored acceptance criterion and record the observed counts against the pre-flight baselines.

## §G Anti-patterns

- **AP-1 — Inventing a second chain.** Adding a bespoke factory chaining mechanism instead of extending `full-pipeline`. The chain already exists; the delta is the verify gate and the plan head.
- **AP-2 — Verify as a pre-sync stage.** Placing verify after the run→sync boundary. It is the run exit gate; a CRITICAL finding must never have reached sync in the first place.
- **AP-3 — Relying on the goal-conditional cap inject.** The existing branch reads goal state at launch; a mid-session goal is invisible to it (R3).
- **AP-4 — Treating an absent artifact as a match.** The predicate defaults to RUN. Any implementation whose failure mode is "skip" is wrong by construction.
- **AP-5 — Papering over R2.** Substituting a different scope-flag combination to dodge the contradiction, rather than resolving which line is authoritative.
- **AP-6 — Cobra flag binding on cc / glm.** Both set `DisableFlagParsing: true`; a `cmd.Flags()` registration would be silently inert.
- **AP-7 — Mirroring verbatim.** Copying a `.claude/` file into the template tree with its SPEC ID, requirement tokens, or `internal/...` package paths intact trips the neutrality guard.
- **AP-8 — Treating "no result" as "no findings" (new).** Any branch that reaches sync without having observed a readable verify result. The MEDIUM/LOW rule is guarded on *readable* precisely so this cannot happen by default.
- **AP-9 — Skipping the whole of Phase 8 (new).** The skip is scoped to Step 0.55.1. Skipping the dependency-manifest audit removes the only check for transitive-vulnerability drift unrelated to this SPEC.
- **AP-10 — Letting a rigor-reduced scan suppress the rigorous one (new).** A DEGRADED-rung verify never suppresses Phase 8. Adding `effort_tier` to the predicate instead is not a fix — an effort level is not a rung.
- **AP-11 — Testing the predicate without testing its caller (new).** `Matches` is a pure function; every unit test over it passes whether or not sync Phase 8 ever calls it. The consumption criterion (AC-FM-020d) is the one that fails on dead code.
- **AP-12 — Threading a factory parameter through the launch chain (new).** Four signatures plus two test seams, for a signal the environment already carries to the same call site.
- **AP-13 — Expressing the rung exclusion as a deny-list (v0.3.0).** `verify_rung != "DEGRADED"` permits suppression on an absent or empty field, which a best-effort record makes reachable. The exclusion is an allow-list: suppression requires a *recorded* `PRIMARY` or `FALLBACK`. Any predicate whose behavior on a missing field is "suppress" is wrong by construction, in the same way AP-4 is.
- **AP-14 — Leaking the factory env mutation past the call (v0.3.0).** An `os.Setenv` without a `defer`ed restore turns AC-FM-022a's negative control into a test-order lottery and lets a later spawn inherit factory semantics silently. Restore prior presence, not an empty string, and restore on the error path too.
- **AP-15 — Presenting the rung as a fourth peer outcome (v0.3.0).** Severity and rung are orthogonal axes; listing them as four alternatives and asserting disjointness produces two binding requirements with no precedence between them, which is how v0.2.0 shipped a requirement that was unsatisfiable exactly where it mattered.
- **AP-16 — Writing an acceptance criterion whose baseline was never measured (v0.3.0; third shape added v0.5.0, fourth added v0.6.0, fifth added v0.7.0, sixth added v0.8.0).** Six shapes have now been found in this SPEC: an absence grep run against a tree with pre-existing matches (AC-FM-025's mirror clause), a presence grep whose pattern already matches (AC-FM-007), — found during run-phase, after M5 landed — **a coverage floor set above the package's actual starting point** (AC-FM-025's coverage conjunct: 90% asserted on a package measured at 76.3-76.4%), and **a green-suite assertion made against a repository that already has a red test** (AC-FM-025's `go test ./...` exits-0 conjunct, with `TestRunHarnessObserveStop_ProposeChainAutoRuns` red at the pre-SPEC baseline), and **a requirement whose feasibility against an existing mechanical guard was never measured** (M1's "add the verify exit gate to `run.md`", written against a file that `internal/skills/workflow_split_test.go` `TestEntryRouterLOCCeiling` caps at 200 LOC and that already stood at exactly 200, leaving zero headroom for the 34-line block; the content relocated to `run/mode-orchestration.md`, and every criterion greping `run.md` for it had to be re-pointed), and **a decision procedure whose detection mechanism was never tested against the failure modes it must catch** (AC-FM-025's v0.6.0 full-suite bound: `go test ./... | grep '^--- FAIL'`, written from the one failure shape that had been observed — a `--- FAIL` line — against a tool that also produces timeout panics, build failures, and vet failures, none of which emit a `--- FAIL` line at all; a hang was therefore invisible to it, and a second pre-existing defect went undetected until the full suite was actually run). Measure in §C pre-flight, assert a delta against the measured value, and never let "fixing" a criterion mean editing unrelated shipped files. The third shape is the one that survives plan-audit most easily, because a coverage number reads as a policy target rather than as a claim about this tree; the fourth survives for the same reason and one worse — "all tests pass" reads as so obviously correct that nobody checks whether it is true of the starting tree. The fifth is the widest: the unmeasured baseline is not a number in a criterion at all but the *design premise of a milestone*, so plan-audit reads a plausible sentence about where content goes and has nothing to compare it against — a file's remaining headroom under a CI guard is as much a measurable baseline as a coverage figure, and belongs in §C pre-flight whenever a milestone adds lines to a capped file. The sixth is the subtlest, because the criterion it produces *looks* rigorous: it is mechanical, it is falsifiable, and it names its expected output exactly. What was never measured is the **detection mechanism's own coverage** — the set of failure shapes the tool can emit, versus the subset the pattern can see. A grep written from the single failure that happened to be in front of the author will silently pass every failure that prints differently, and the criterion's apparent precision is what stops anyone from checking. Enumerate the tool's output shapes before choosing the pattern, and prefer the coarser signal that every shape produces (`go test`'s per-package `FAIL` line) over the finer one that only some shapes produce (`--- FAIL`), then narrow to identity in a second step. Four corollaries learned while correcting them: when a measured baseline is noisy, anchor the delta to the **lowest** observation — a floor set at the highest observation fails on jitter and is unfalsifiable in the opposite direction; and when bounding around a pre-existing defect, name **each individual defect** — by test name where it is a failing test, by call site where it is a hang reached from many tests — never a failure count or a blanket "known failures" clause, or the bound becomes a hiding place for the next regression; and when a relocation moves content between files, re-point the criteria by **substituting the path only** — a re-point that also "tidies" a pattern silently changes what is asserted, and the relocation record must state that the address changed rather than the content disappearing; and a defect the SPEC itself caused is **repaired, never excluded** — listing a repaired defect among the exclusions preserves a hole through which its own regression would later pass unnoticed, and the repaired-vs-predated distinction is what makes the whole bound defensible rather than convenient.

## §H Resolved decisions

No unresolved clarification markers remain. The three open questions carried by v0.1.0 are closed as follows.

- **Factory goal wall-clock bound — RESOLVED by user decision (v0.2.0).** The `factory_chain` goal arms with `--max-turns 0 --max-duration 14400`: infinite turns, four-hour wall clock. The block cap stays at 200 via the existing inject path, which is self-consistent because `injectStopHookBlockCapForGoal` already keys on exactly the `--max-turns 0` pattern. Termination is by chain completion, the four-hour bound, the stagnation guard, or a human-gate refusal. Accepted risk: an unattended run may consume up to four hours of tokens. Encoded as REQ-FM-027 with the flags as concrete values — `run.md` records that a prose "stop after N turns" clause is not parsed, so only the flags bind.
- **Chain head without a SPEC id — CLOSED as already answered.** REQ-FM-005 `shall`-states it: the launcher accepts the optional SPEC identifier and treats its absence as "the chain begins at plan-phase from the operator's first prompt". The v0.1.0 marker re-opened a question its own requirement had already decided.
- **`revision.json` writer ownership in v1 — CLOSED as already decided.** REQ-FM-018 scopes the Go deliverable to a reader, and §C Exclusions lists "Writing `revision.json` or `findings.jsonl` from Go" as out of scope. No Go writer ships in v1; the producer stays with the `--deep` workflow agents.

## §I Cross-references

- `.claude/skills/moai/workflows/moai.md` § run→sync chaining policy — the `full-pipeline` contract being extended
- `.claude/skills/moai/workflows/run.md` § Run-phase Autonomy — Kickoff ordering, `ac_converge`, the unparsed-turn-clause warning
- `.claude/skills/moai/workflows/sync/quality-gates-quality.md` § Phase 8 — the dedup target and the always-on manifest audit
- `.claude/skills/moai/workflows/review.md` — `--deep`, `--repo`, the results directory, `revision.json`, `findings.jsonl`, the degradation ladder
- `.claude/rules/moai/workflow/goal-directive.md` — arm-only semantics, the block-cap raise. **Also an edit target** (REQ-FM-028, M1): its § Raising the block cap for an infinite goal states the inject fires only for an armed `--max-turns 0` goal "at launch time", which REQ-FM-023 falsifies. Its template mirror is byte-identical and receives the same amendment.
- `.claude/rules/moai/workflow/worktree-state-guard.md` § Divergence Threshold — the porcelain exclusion set reused by the `treeDirty` derivation
- `internal/cli/launcher_blockcap_infinite.go` and `internal/cli/launcher.go:783/786/790` — the inject path being extended and the `os.Environ()` seam it reads

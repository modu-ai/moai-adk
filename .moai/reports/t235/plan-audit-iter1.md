# SPEC Review Report: SPEC-GATE-THREE-AXES-001

Iteration: 1/2 (Tier M ceiling)
Verdict: **PASS-WITH-DEBT**
Overall Score: **0.85** (Tier M PASS threshold 0.80) — skip-eligible

Audited at pinned commit `10e252834`, branch `WT-gate-three-axes`, worktree `.claude/worktrees/t235`.
`git status --short` was empty and `git status --short -- .moai/specs/SPEC-GATE-THREE-AXES-001/` was empty: the audited content is the committed content, not a drifted tree.
`git diff --stat 294b4b6ab HEAD` touches only `.moai/reports/t235/` and `.moai/specs/SPEC-GATE-THREE-AXES-001/` (5 files, 578 insertions, 0 deletions) — no source file differs between the SPEC's measurement baseline and the audited commit, so every `294b4b6ab` citation was verifiable at HEAD.

Reasoning context ignored per M1 Context Isolation. No author reasoning was supplied; the audit reads the four SPEC artifacts, the premise report, and the cited source.

---

## Must-Pass Results

- **[PASS] MP-1 REQ number consistency** — `REQ-GTA-001` … `REQ-GTA-016`, sixteen distinct ids, no gap, no duplicate, uniform three-digit padding. Measured: `grep -o 'REQ-GTA-[0-9]*' spec.md | grep -o '[0-9]*$' | sort -u` → `001 002 003 004 005 006 007 008 009 010 011 012 013 014 015 016`; `grep -c '^\*\*REQ-GTA-'` → `16`.
- **[PASS] MP-2 GEARS format compliance** — judged against the **requirement layer** (`REQ-XXX` in `spec.md` §B). All sixteen match a GEARS pattern, each with its pattern named inline: Ubiquitous (001 `spec.md:67`, 005 `:79`), event-driven (002 `:70`, 008 `:90`, 013 `:107`), state-driven (003 `:73`), capability-gate (004 `:76`, 014 `:110`), unwanted/`shall not` (006 `:82`, 007 `:85`, 010 `:96`, 011 `:99`, 015 `:113`, 016 `:116`), and two compound clauses (009 `:93` event-driven + capability-gate; 012 `:104` state-driven + conjunct). Zero informal requirements, zero Given-When-Then entries in the requirement layer. The Given-When-Then entries in `acceptance.md` are `AC-GTA-XXX` verification-layer criteria and are graded under Group 4, not here.
- **[PASS] MP-3 YAML frontmatter validity** — all 12 canonical fields present with correct types (`spec.md:2-15`): `id title version status created updated author priority phase module lifecycle tags`, plus permitted extras `tier: M` and `era: V3R6`. `version: "0.1.0"` is a quoted semver string; `created`/`updated` are ISO dates; `priority: P1` is enum-valid; `lifecycle: spec-anchored` is enum-valid; `tags` is a comma-separated string. No rejected snake_case alias (`created_at` / `updated_at` / `labels` / `spec_id`). Independently corroborated: `moai spec lint .moai/specs/SPEC-GATE-THREE-AXES-001/spec.md` → `✓ No findings — all SPEC documents are valid`.
- **[N/A] MP-4 language neutrality** — single-language SPEC, scoped to this repository's own Go source (`internal/hook/quality`, `internal/cli`). N/A auto-passes. Noted in passing: `plan.md:110` correctly binds the one shipped artifact (`gate.yaml`) to Template-First ordering and a neutrality audit, and forbids SPEC ID / REQ token / internal date in template content.
- **[PASS] MP-5 D7 cross-SPEC reconciliation** — no BLOCKING finding. `grep -Eo 'SPEC-([A-Z][A-Z0-9]+-)+[0-9]+'` across all four artifacts returns only `SPEC-GATE-THREE-AXES-001` itself. The other identifiers in the body (`t218`, `t233` / issue #1631, `t235` / issue #1639) are card and issue ids, not SPEC ids, so no retired / superseded / archived SPEC is referenced without reconciliation.
- **[PASS] MP-6 D8 cross-platform discipline** — no BLOCKING finding. `syscall` appears 3× in `plan.md` and 0× in `spec.md` / `acceptance.md`. All three occurrences carry the discipline rather than violating it: `plan.md:37` states the rule ("no naked syscall in the shared body") and prescribes a build-tagged pair following `internal/spec/lock_{unix,windows}.go`; the other two (`:50`, `:54`) cite existing code in the **rejected** substrate. The one platform primitive the SPEC introduces (`Setpgid`) is governed by an explicit cross-platform exemption clause — `plan.md:37` (platform split), `:39` (Windows applies no process-group primitive and reports that descendants may survive), and `plan.md:89` (M2 step 1 adds the build-tagged pair). This satisfies D8-2's second admissible form.
- **[PASS] MP-7 clarification gate** — `grep -rn '\[NEEDS CLARIFICATION' .moai/specs/SPEC-GATE-THREE-AXES-001/` → rc=1, zero matches across all four artifacts (`research.md` does not exist at Tier M).

---

## Category Scores

| Dimension | Score | Rubric Band | Evidence |
|-----------|-------|-------------|----------|
| Clarity | 0.85 | between 0.75 and 1.0 | Every requirement carries its GEARS pattern inline and reads unambiguously. The two ambiguities found are in the verification layer, not the requirement layer, and one has a genuine scope contradiction against `plan.md` (D1). Deducted for D1 and D2. |
| Completeness | 0.90 | between 0.75 and 1.0 | HISTORY `spec.md:20-24`; WHY §A `:26-60`; WHAT §B `:62-116`; HOW `plan.md`; REQUIREMENTS 16; ACCEPTANCE CRITERIA 16; four `### Out of Scope — <topic>` H3 sub-headings each with `-` bullets (`spec.md:124,129,136,142`); traceability table `:148-152`; Definition of Done `acceptance.md:186-196`; risk register `plan.md:133-140`. Deducted only for the AC-GTA-003 coverage gap (D1). |
| Testability | 0.85 | between 0.75 and 1.0 | The strongest dimension and the site of both blocking defects. Every AC is binary-testable, carries an explicit **Mutant** line and a named **RED**, and the file opens with two standing rules (`acceptance.md:7-8`) that are exactly the right ones. Deducted for D1 and D2. |
| Traceability | 0.80 | between 0.75 and 1.0 | Every REQ has at least one AC and no AC is orphaned (reconstructed below), but the mapping is stated at range granularity only and is non-1:1 on two of three axes (D4). |

Aggregate: (0.85 + 0.90 + 0.85 + 0.80) / 4 = **0.85**.

### REQ → AC coverage, reconstructed and verified

The SPEC states the mapping only as ranges. Reconstructed per-REQ, coverage is complete:

| REQ | Covered by | Note |
|-----|-----------|------|
| 001 | AC-001 | |
| 002 | AC-002 | |
| 003 | AC-003 | **partial — see D1** |
| 004 | AC-004 | fixture underspecified — see D2 |
| 005 | AC-005 | |
| 006 | AC-006 | |
| 007 | AC-007 | |
| 008 | AC-008 | |
| 009 | AC-009 (Unix half) + AC-008's Windows half | compound REQ split across two AC halves |
| 010 | AC-010 first half (two regression tests + diff constraint) | merged AC |
| 011 | AC-010 second half (byte-identity + cwd) | merged AC |
| 012 | AC-011 + AC-012 | compound REQ split across two ACs |
| 013 | AC-013 | |
| 014 | AC-014 | |
| 015 | AC-015 + AC-013's upper bound | "never fail" + "never block unbounded" |
| 016 | AC-016 | |

No uncovered REQ. No orphaned AC.

---

## Findings the audit was asked to weigh

**Citation accuracy — every cited span re-read at the tree, all resolve.** Including the four the author reports having corrected (`progress.md:13`), which I re-verified independently rather than trusting: `runStep`'s success branch is `err := cmd.Run(); if err == nil { return true, "" }` at `gate.go:1020-1022` (exact); the captured-output buffers are `var stdout, stderr bytes.Buffer` at `:1016-1018` (exact); `exec.CommandContext(stepCtx, name, args...)` at `:1006` (exact); the dropped test-step value is `if ok, out := g.executeStep(ctx, testStep, g.config.TestTimeout); !ok {` at `:397` (exact — `out` is never read on the success branch). Also confirmed exact: `Run` at `:322`, `parentBinds` at `:996-1002`, `resolveNodeTestStep` at `:676`, `executeStep` at `:776`, `appendReason` at `gate_typecheck.go:150`, the two regression tests at `gate_timeout_attribution_test.go:44` and `:67`, `mapConfigGateToQuality` at `cli/gate.go:136`, `quality.GateConfig` at `gate.go:20`, `config.GateConfig` at `types.go:764-782`, `gate.yaml:16-22`, `go.mod:3` (`go 1.26.4`) and `go.mod:29` (`golang.org/x/sys v0.47.0`). Every negative grep the SPEC asserts reproduces: `Setpgid|Killpg|killpg|SysProcAttr` in `internal/hook/quality/` → 0; `WaitDelay` in `internal/` → 0; `_unix|_windows` files in `internal/hook/quality/` → 0; lock acquisition in `internal/cli/gate.go` → 0.

**The axis-1 named hazard is defended.** The card's hazard — output is emitted but its content is not the execution result — is killed at four independent points, not one. AC-GTA-005 (`acceptance.md:58`) explicitly names and rejects the non-emptiness formulation: "An assertion of the form 'stderr is non-empty' would pass this mutant and is explicitly rejected as a formulation of this criterion." AC-GTA-001 requires a *difference* between two runs that executed different work with every other line identical, which no config-rendered summary can produce. AC-GTA-002's two-sided bound (`≥ D` **and** `< D` against one shared budget) admits no constant and no budget echo. AC-GTA-007 targets the residue hazard that the SPEC's own chosen design (`plan.md:17`, records seeded from the toolchain before execution) introduces. I checked each axis-1 AC against exactly this mutant; only AC-GTA-003 has a gap, recorded as D1.

**The two-mechanism separation holds (plan.md §A.2).** Tested directly against the question asked. A WaitDelay-only implementation **does** satisfy AC-GTA-008 in full: `Wait` returns once the delay closes the pipes, so the T+G bound is met. AC-GTA-009 catches it — the recorded descendant PID is still alive when probed, because nothing signalled the process group. The two criteria are therefore not redundant, and `acceptance.md:103` states this reasoning correctly. The separation is sound as written.

**AC-GTA-010 does establish t218 preservation, and M2 is implementable without violating it.** This needed checking because `plan.md:91` proposes to "extend the timeout reason", which would break the regression tests if they asserted string equality. They do not: both use `strings.Contains` (`gate_timeout_attribution_test.go:58`, `:62`, `:78`). Added text around the existing reason keeps both green, so `plan.md:41`'s constraint ("any new text must be added around the existing reason, not in place of it") is satisfiable and AC-GTA-010's no-modified-line diff constraint is not in conflict with M2. The criterion's own strength is adequate: the `git diff` constraint kills the edit-the-test mutant mechanically, and the byte-identity plus working-directory assertion kills the detach-from-`cmd.Dir` / reorder / truncate mutant. Both directions of the attribution semantics are covered, since the two preserved tests assert the parent-binds and step-binds cases respectively.

**The substrate rejection premise is correct, and so is the substitute.** `internal/lockfile`'s Unix path is `syscall.Flock(int(f.Fd()), syscall.LOCK_EX)` with no `LOCK_NB` (`lockfile_unix.go:23-25`) — blocking, as claimed. Its Windows path is a `sync.Mutex` map keyed by path, documented in-file as offering no cross-process protection (`lockfile_windows.go:11-26`). The `internal/spec/lock.go` alternative is `unix.LOCK_EX|unix.LOCK_NB` returning `ErrSpecCloseLockHeld` immediately (`lock_unix.go:36-38`, exact) and a genuine cross-process `O_CREATE|O_EXCL` on Windows (`lock_windows.go:80-102`). The premise correction at `spec.md:58` also verifies: `grep -rn "internal/lockfile" internal/` (excluding tests and the package itself) returns exactly the three importers named — `internal/cli/settings.go:12`, `internal/cli/glm_tools.go:32`, `internal/cli/taskledger/taskledger.go:16` — and `board_lock.go`'s only occurrence is the comment at `:10` stating non-use, verbatim as quoted.

**The Tier M trim did not drop an obligation.** Both compound clauses remain independently testable: REQ-GTA-009's two branches map to AC-GTA-009 (Unix termination) and AC-GTA-008's Windows half (reported survival) — separately verifiable, on separate platforms. REQ-GTA-012's two conjuncts map to AC-GTA-011 (non-overlap, asserted on execution timestamps) and AC-GTA-012 (notice names PID P) — separately verifiable. Of the merged AC pairs, AC-GTA-010 is the only genuine merge, carrying REQ-GTA-010 and REQ-GTA-011; both failure modes survive the merge with their own assertion and their own named mutant (Mutant A for attribution, Mutant B for happy-path side effects). The cost is granularity, not coverage: a single AC-GTA-010 verdict cannot say which of the two REQs regressed. Recorded as part of D4.

**Scope holds against the lint axis.** M1 step 2 (`plan.md:77`) does touch `executeStep`, whose five early-return skip paths I confirmed at `gate.go:778-815` (exactly five: `DisabledSteps`, optional-binary `LookPath`, `configFiles`, `changedExts`, `sourceExts`). But it changes what that frame *reports*, not how lint steps are selected, resolved, or executed, so the exclusion at `spec.md:126` is respected semantically. The collision risk with t233 is not denied but registered as a sequencing risk (`plan.md:140`), which is the right treatment for an undispatched card.

---

## Defects Found

**D1.** REQ-GTA-003 scope contradiction — `spec.md:73` vs `plan.md:77` vs `acceptance.md:33-40` — Severity: **major** — Class: **blocking** — REQ-GTA-003 reads "distinguishing a configuration-disabled step from one skipped because its tool, its config file, or its source files were absent". Grammatically that is a two-way distinction: config-disabled versus an absence group. `plan.md` M1 step 2 reads it as a five-way one — "Each of `executeStep`'s five early-return skip paths (`gate.go:777-816`) must report *which* skip it took, per REQ-GTA-003" — and that five-path set includes the `changedExts` staged-file path (`gate.go:792-801`), which REQ-GTA-003 never names. AC-GTA-003 verifies only the two-way reading: its two fixtures are tool-absent-from-PATH and turned-off-via-`disabled_steps`. Either reading leaves a hole. Under the plan's reading, three skip paths (config-file absent, `changedExts` no-match, `sourceExts` no-match) ship with no acceptance coverage, and an implementation collapsing them into one generic reason passes AC-GTA-003 while violating REQ-GTA-003 — a surviving mutant on the SPEC's load-bearing property. Under the REQ's literal reading, the plan implements a distinction nothing asked for, against the simplicity ladder. **Required fix**: settle the scope. Either narrow `plan.md` M1 step 2 to the two-way distinction, or widen REQ-GTA-003 to enumerate each skip path it binds (adding `changedExts` explicitly) and extend AC-GTA-003's Given to a fixture per named path.

**D2.** AC-GTA-004's second fixture does not determine its expected value — `acceptance.md:44-46` — Severity: **major** — Class: **blocking** — The second fixture is specified only as "a second Node fixture that declares only `scripts.test`", with expected value "the unresolved step form". `resolveNodeTestStep` (`gate.go:676-699`) has three branches, not two: after the `test:run` branch, `if flag := nodeNonWatchFlag(scripts["test"]); flag != ""` resolves the step to a *second* form (`name: nodeTestStepName + " " + flag`, args `npm test -- --passWithNoTests <flag>`). `nodeNonWatchFlag` (`gate.go:729-744`) returns non-empty when the script is watch-prone and mentions vitest (`--run`) or jest (`--ci`) — and a bare `vitest` counts as watch-prone by `nodeScriptWatchProne`'s documented rule (`gate.go:746-752`). So a fixture written as `"test": "vitest"` — an entirely natural choice — yields a resolved form, and the criterion fails against a *correct* implementation. The mutant analysis itself survives (Mutant B is still killed), but the criterion's expected value is underdetermined by its stated Given. **Required fix**: name the `scripts.test` content in the Given, choosing a non-watch-prone script (e.g. `"test": "echo ok"`) so the unresolved form is the determined outcome; or add a third fixture and expected value for the `nodeNonWatchFlag` branch.

**D3.** `disabled_steps` polarity is inverted and unnamed — `acceptance.md:16` (AC-GTA-001 Given) and `:35` (AC-GTA-003 Given) — Severity: **minor** — Class: **blocking** — Both criteria turn a step off through `gate.disabled_steps`. That knob's polarity is inverted by documented convention: `internal/config/types.go:775-779` states "the runner's inverted convention: an entry whose value is FALSE skips that step (issue #667 Fix 3)", `internal/cli/gate.go:150-152` maps it through verbatim for that reason, and the runner reads it as `if disabled, ok := g.config.DisabledSteps[step.name]; ok && !disabled { return true, "" }` (`gate.go:778-780`). The SPEC never names this. A fixture written with the intuitive polarity (`{test: true}`) leaves the step *running* in both of AC-GTA-001's fixtures, so the required difference between the two summaries is absent and the criterion fails against a correct implementation — for a reason that has nothing to do with the code under test. Cheap to prevent, expensive to diagnose in run-phase. **Required fix**: state the polarity in both Given clauses, e.g. "disabled via `gate.disabled_steps: {test: false}` (the runner's inverted convention — a FALSE value skips the step)".

**D4.** REQ-to-AC mapping is range-level only — `spec.md:148-152` — Severity: **minor** — Class: **optional** — The traceability table maps ranges (`REQ-GTA-001 … 007` → `AC-GTA-001 … 007`), and two of three axes are non-1:1: axis 2 is 4 REQs → 3 ACs, axis 3 is 5 REQs → 6 ACs. No AC names its REQ. Coverage is in fact complete — I reconstructed it above and every REQ has at least one AC with no orphaned AC — but the SPEC does not state which AC discharges which REQ, so a run-phase reader must re-derive it, and an AC-GTA-010 failure cannot be attributed to REQ-GTA-010 versus REQ-GTA-011 without that derivation. **Required fix (optional)**: add a `Verifies: REQ-GTA-NNN` line to each AC, or expand §E's table to per-AC granularity.

**D5.** `plan.md` §A.3 holder-identity row is overstated — `plan.md:52` — Severity: **minor** — Class: **optional** — The comparison table's final row attributes "PID recorded in the artifact" to the `internal/spec/lock.go` pattern. Only the Windows implementation writes one (`lock_windows.go:86`, `fmt.Fprintf(file, "pid=%d\n", os.Getpid())`); the Unix path opens with `O_CREAT|O_RDWR` and flocks, recording no identity (`lock_unix.go:31-40`). The obligation REQ-GTA-012 and REQ-GTA-014 rest on is nonetheless met by the chosen precedent — `internal/kanban/board_lock.go`'s `BoardLockOwner{PID, CreatedAt}` (`board_lock.go:50-53`), which `plan.md:56` names correctly. So the decision is sound and only the table row is imprecise. **Required fix (optional)**: qualify the row as "Windows impl only; `board_lock.go` layers it on both".

**D6.** AC-GTA-008's RED is an intentional hang with no stated isolation — `acceptance.md:95` — Severity: **minor** — Class: **optional** — The RED is "this test does not return within T + G — it blocks until the test binary's own `-timeout` fires". Observing it therefore costs a full test-binary timeout stall. The existing sleeper narrows with `-test.run=^TestHelperSleep$` and `-test.timeout=60s` (`gate_timeout_attribution_test.go:28-31`), so the pattern is safe when reused, and `plan.md:96` shows the author is alert to the re-execution hazard. But the SPEC does not say the RED is observed in isolation rather than inside a package run. **Required fix (optional)**: state that AC-GTA-008's RED is observed with a narrowing `-test.run`, not as part of `go test ./internal/hook/quality/...`.

---

## Regression Check

Not applicable — iteration 1.

---

## Recommendation

**PASS-WITH-DEBT.** All seven must-pass criteria pass, and the aggregate 0.85 clears the Tier M threshold of 0.80. The SPEC is unusually well-grounded: every citation resolves, every negative grep reproduces, the premise corrections are real corrections, the substrate rejection is verified against the source rather than asserted, and the acceptance criteria carry mutant analysis that is genuinely adversarial rather than decorative — AC-GTA-005 pre-emptively rejects the exact weak formulation the card warned about, and AC-GTA-007 targets a hazard the SPEC's own design choice introduces.

The debt is three blocking-class defects in `acceptance.md`, all cheap to discharge and none requiring re-planning:

1. **Settle REQ-GTA-003's scope (D1).** This is the one surviving mutant on the SPEC's load-bearing property and should be closed before M1 starts. Decide whether the requirement binds a two-way or a per-path distinction, then align `plan.md` M1 step 2 and AC-GTA-003 to that decision. If per-path, note that `changedExts` is a fifth path the requirement text does not currently name.
2. **Determine AC-GTA-004's second fixture (D2).** Name the `scripts.test` content so the expected value follows from the Given; a non-watch-prone script is the simplest choice.
3. **State the `disabled_steps` polarity (D3).** One clause in each of AC-GTA-001 and AC-GTA-003.

D4, D5, and D6 are optional and left to the orchestrator's discretion.

Because the three blocking defects are confined to criterion wording and none touches the requirement set, the milestones, or the substrate decisions, a re-audit — if run — should be scoped to the AC-GTA-003 / AC-GTA-004 / AC-GTA-001 delta rather than repeating this audit.

---

## Residual risk

The cross-model second opinion did not materialize and this verdict rests on the single-auditor anchor. `mcp__moai__audit_multi` was invoked; the GLM backend returned `inconclusive` (fail-open, "z.ai response carried no content"), and the codex backend audited a different target entirely — its findings cite `SPEC-CC297-001`, `internal/statusline/renderer.go`, and `.moai/astgrep-rules/go/*.yml`, which are uncommitted changes in the *main* checkout, not this worktree's pinned commit. Those findings were discarded as off-target rather than folded in; none of them bears on `SPEC-GATE-THREE-AXES-001`. The convergence engine reported `overall_verdict: PASS-WITH-DEBT` with `disagreement_flag: false`, but that agreement is an artifact of the anchor being the only on-target verdict and carries no independent weight.

Two further risks survive the evidence gathered here. AC-GTA-008's Windows half is unverifiable at plan time and remains a gap until the CI Windows matrix run is observed — `acceptance.md:94` states this correctly and forbids citing `GOOS=windows go vet` as behavioural evidence, which is the right constraint. And AC-GTA-009's liveness probe is keyed on a PID; PID reuse inside the grace window would produce a false negative. The window is short enough that this is negligible in practice, and it is noted rather than raised as a defect.

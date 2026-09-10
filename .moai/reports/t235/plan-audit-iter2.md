# SPEC Review Report: SPEC-GATE-THREE-AXES-001

Iteration: 2/2 (Tier M ceiling)
Verdict: **PASS-WITH-DEBT**
Overall Score: **0.92** — delta **+0.07** from iter-1's 0.85 (monotonic; no score regression, so no STOP escalation)

Audited at pinned commit `f8186b172`, branch `WT-gate-three-axes`, worktree `.claude/worktrees/t235`. `git status --short` empty — audited content is committed content.

**Scope**: delta re-audit, not a repeat of iter-1. Two commits since the iter-1 target `10e252834`: `16d2ad0ed` (defect discharge, 4 files +116/−22) and `f8186b172` (the iter-1 report, which closes the dangling reference `progress.md` was citing while the file was untracked). The audit covers the `git diff 10e252834 f8186b172` surface — AC-GTA-001 / AC-GTA-003 / AC-GTA-004, REQ-GTA-003, `spec.md` §E, `plan.md` §A.3 and M1 steps 2-3, `progress.md` — plus the source claims those changes newly depend on. Findings carried unchanged from iter-1 are not restated; see `plan-audit-iter1.md`.

Reasoning context ignored per M1 Context Isolation.

---

## Must-Pass Results — no regression, 7/7

- **[PASS] MP-1 REQ number consistency** — re-measured at `f8186b172`: `001 002 003 004 005 006 007 008 009 010 011 012 013 014 015 016`, `grep -c '^\*\*REQ-GTA-'` → `16`. No gap, no duplicate, padding uniform.
- **[PASS] MP-2 GEARS format compliance** — only REQ-GTA-003 changed. Its widened form (`spec.md:73`) remains state-driven: "**While** a step is not executed, the quality gate **shall** record in the summary the observed reason it was not executed, distinguishing each of the five paths…". The `(a)`–`(e)` enumeration is a qualifier on the response clause, not a pattern change. The other fifteen are untouched by the diff.
- **[PASS] MP-3 YAML frontmatter validity** — frontmatter is outside the diff surface (`spec.md` changes are confined to §B REQ-GTA-003 and §E). Re-confirmed independently: `moai spec lint` → `✓ No findings — all SPEC documents are valid`.
- **[N/A] MP-4 language neutrality** — unchanged; single-language SPEC.
- **[PASS] MP-5 D7 cross-SPEC reconciliation** — the diff introduces no new `SPEC-*` token; the only SPEC-ID present across all four artifacts is still `SPEC-GATE-THREE-AXES-001`.
- **[PASS] MP-6 D8 cross-platform discipline** — no `syscall` occurrence changed; the §A.2 platform-split clause and the M2 build-tagged-pair prescription are outside the diff surface.
- **[PASS] MP-7 clarification gate** — `grep -rn 'NEEDS CLARIFICATION' .moai/specs/SPEC-GATE-THREE-AXES-001/` → rc=1, zero matches.

Independently re-run at this commit rather than carried forward: `moai spec audit --filter-spec SPEC-GATE-THREE-AXES-001` → `Grandfathered: 0 / Modern-era clean: 1 / Drift findings: 0`. This corroborates `progress.md`'s post-fix verification line rather than trusting it.

---

## Category Scores

| Dimension | iter-1 | iter-2 | Δ | Evidence |
|-----------|--------|--------|---|----------|
| Clarity | 0.85 | **0.90** | +0.05 | D1's widening removed a structural REQ-vs-plan contradiction; the polarity warning and the fixture (d) caveat remove two fixture ambiguities. Offset by the new D7, which is a one-clause wording divergence rather than a structural one. |
| Completeness | 0.90 | **0.95** | +0.05 | The iter-1 coverage gap is closed: AC-GTA-003 now carries a fixture per skip path, including the `changedExts` path REQ-GTA-003 never previously named. Per-REQ traceability table added. |
| Testability | 0.85 | **0.88** | +0.03 | Mutant B (two-way collapse) is genuinely killed and the kill is mechanically checkable; Mutant C added and killed; all three AC-GTA-004 fixture→branch mappings verified correct against source; RED isolation clause added. Held down by D7 — two of three AC-GTA-004 expected values do not match the command actually executed. |
| Traceability | 0.80 | **0.95** | +0.15 | Per-requirement table in `spec.md` §E plus a `Verifies:` line on every AC — bidirectional. `grep -c '^\*\*Verifies\*\*:'` → `16`, one per criterion. AC-GTA-010's two halves are explicitly declared separable so a failure attributes to REQ-GTA-010 or REQ-GTA-011. |

Aggregate: (0.90 + 0.95 + 0.88 + 0.95) / 4 = **0.92** against the Tier M PASS threshold of 0.80.

---

## Regression Check — iter-1 defects

| iter-1 defect | Status | Evidence |
|---------------|--------|----------|
| D1 — REQ-GTA-003 scope contradiction | **RESOLVED** | REQ-GTA-003 (`spec.md:73`) now enumerates five paths (a)–(e), including the `changedExts` path the earlier text never named. AC-GTA-003 rewritten from two fixtures to a five-row table. `plan.md` M1 step 2 now cites five discrete spans instead of a range. All five spans verified exact — see below. |
| D2 — AC-GTA-004 fixture underdetermined | **RESOLVED** (fix introduced D7) | Each fixture now names its exact `package.json` content and target branch. All three fixture→branch mappings verified correct against source. The underdetermination is gone; a separate expected-value divergence is raised as D7. |
| D3 — `disabled_steps` polarity unnamed | **RESOLVED** | Polarity stated in both Givens with a load-bearing warning. AC-GTA-001's fixture is now literally `disabled_steps: {"go test": false}` versus the key omitted. All three cited sources verified exact: `gate.go:778-780` is `if disabled, ok := g.config.DisabledSteps[step.name]; ok && !disabled { return true, "" }`; `types.go:775-779` documents "an entry whose value is FALSE skips that step (issue #667 Fix 3)"; `cli/gate.go:150-152` maps through verbatim. |
| D4 — range-level REQ→AC mapping | **RESOLVED** | `spec.md` §E is per-requirement; every AC carries `Verifies:`. Counts measured: 16 REQs / 16 ACs / 16 `Verifies:` lines. |
| D5 — overstated holder-identity row | **RESOLVED** | `plan.md:52` now reads "Windows impl only (`lock_windows.go:86` writes `pid=%d`); the Unix path opens `O_CREAT|O_RDWR` and flocks, recording nothing (`lock_unix.go:31-40`). `board_lock.go` layers identity on **both** platforms via `BoardLockOwner{PID, CreatedAt}` (`board_lock.go:50-53`)" — which is exactly what I measured at iter-1. |
| D6 — AC-GTA-008 RED unbounded | **RESOLVED** | Mandatory RED-isolation clause added: narrowing `-test.run` plus explicit `-test.timeout`, never inside a package run, citing the existing pattern at `gate_timeout_attribution_test.go:28-31`. |

No stagnation: every defect moved. No iter-1 defect persists unchanged.

---

## Verification of the iter-2 changes

**The widened AC-GTA-003 does kill Mutant B.** Verified by construction rather than accepted as asserted. Mutant B is "reporting `disabled` for path (a) and one shared `absent` reason for paths (b) through (e)" — it emits **two** distinct reason texts. The Then clause requires "five mutually distinct reason texts for the skipped step", which is a pairwise-inequality check over five strings. Two ≠ five, so the mutant fails the criterion. This is the mutant the earlier two-fixture formulation admitted in full, and it is now excluded mechanically, not by judgement.

**The five reason texts are discriminable as specified, not merely asserted.** The criterion carries its weight in two layers. The mechanical layer — five mutually distinct strings — is decidable without interpretation. The semantic layer ("each reason names its own observation") is pinned by three concrete non-claims rather than left to taste: the config-disabled case must not claim the tool was missing, the absent-binary case must not claim a config file was missing, and the no-staged-match case must be distinguishable from the no-source case. A tester can decide all three. The distinctness requirement is what makes the criterion binary; the named non-claims stop distinctness from being satisfied by five arbitrary labels.

**All five skip-path citations are exact.** Measured line by line at `f8186b172`:

| Path | Cited | Actual construct |
|---|---|---|
| (a) config-disabled | `gate.go:778-780` | `if disabled, ok := …DisabledSteps[step.name]; ok && !disabled { return true, "" }` ✓ |
| (b) optional + LookPath | `:782-786` | `if step.optional { if _, err := exec.LookPath(step.binary); err != nil { return true, "" } }` ✓ |
| (c) configFiles absent | `:787-789` | `if len(step.configFiles) > 0 && !g.anyConfigFileExists(…) { return true, "" }` ✓ |
| (d) changedExts no match | `:793-801` | `if len(step.changedExts) > 0 { … }` ✓ |
| (e) sourceExts no source | `:806-816` | `if len(step.sourceExts) > 0 { … }` ✓ |

**Author finding 1 — the `changedExts` nil-staged caveat — VERIFIED.** `gate.go:796-800` reads `staged := g.cachedStagedFiles(ctx, dir)` / `// If staged is nil, cannot determine — run step conservatively` / `if staged != nil && !hasStagedExt(staged, step.changedExts) { return true, "" }`. The skip is gated on `staged != nil`, so a bare `t.TempDir()` outside a git repository executes the step instead of skipping it, and fixture (d) would test nothing. The AC's requirement of an initialized repository with a staged non-matching file is correct and load-bearing. The in-source comment at `:792` states the same thing independently.

**Author finding 2 — bare `vitest` is watch-prone — VERIFIED.** `nodeScriptWatchProne` spans exactly `gate.go:752-771` as cited. Its logic: empty → false; any `--watch` / `--watchAll` token → true; `fields[0] != "vitest"` → false; a `run` token after the first → false; otherwise true. So `"test": "vitest"` returns true, `nodeNonWatchFlag` then matches the `vitest` token and returns `--run` (`gate.go:733-737`, exact), and fixture B lands in tier (ii). This is what makes the D2 fixture choice load-bearing, as claimed.

**All three AC-GTA-004 fixture→branch mappings are correct.** Fixture A (`"test:run": "vitest run"`) triggers `strings.TrimSpace(scripts["test:run"]) != ""` at `gate.go:684` → tier (i). Fixture B (`"test": "vitest"`, no `test:run`) falls through to `nodeNonWatchFlag` at `:691` → tier (ii). Fixture C (`"test": "echo ok"`) is not watch-prone (`fields[0] == "echo"`), so `nodeNonWatchFlag` returns `""` and the step passes through at `:698` → tier (iii). Each fixture lands where the table says.

**All five skip paths are reachable without touching the toolchain table.** Checked because the per-language toolchain table is excluded by `spec.md` §D, so a fixture requiring a new step property would have been a scope violation. Measured occurrences in the toolchain definitions: `optional:` 25, `configFiles:` 11, `changedExts:` 5, `sourceExts:` 5. Every path has existing steps that exercise it, so the five fixtures are constructible from the table as it stands.

**The budget claim's substantive half holds.** The arithmetic (16/16/16) I re-measured and it matches. The load-bearing half — that no obligation was lost — also holds, and by a stronger route than "nothing was removed": REQ-GTA-003 was **widened** from a two-way distinction to a five-way one, and AC-GTA-003 from two fixtures to five, so the obligation set strictly grew. AC-GTA-004 went from two fixtures to three, likewise additive. No REQ and no AC was deleted, merged, or weakened to make room; the count held because the additions landed inside existing numbered items. The Tier M ceiling still binds exactly.

---

## Defects Found

**D7.** AC-GTA-004's expected values are step names, but REQ-GTA-004 requires the command actually executed — `acceptance.md` AC-GTA-004 table vs `spec.md:76` — Severity: **major** — Class: **blocking** — REQ-GTA-004 is unchanged by this iteration and reads "the quality gate shall name in the summary **the command that was actually executed** rather than the step as configured". The new table's expected values are the resolved `gateStep.name` field, which for two of three fixtures is not the executed command:

| Fixture | Expected in the AC | Command actually executed | Match |
|---|---|---|---|
| A — tier (i) | `npm run test:run` | `npm run test:run` (`binary: "npm"`, `args: ["run","test:run"]`, `gate.go:686-688`) | ✓ coincides |
| B — tier (ii) | `npm test --run` | `npm test -- --passWithNoTests --run` (`args: append([]string{"test","--","--passWithNoTests"}, flag)`, `gate.go:693-695`) | ✗ diverges |
| C — tier (iii) | `npm test` | `npm test -- --passWithNoTests` (the unchanged toolchain step at `gate.go:166`) | ✗ diverges |

`runStep` executes `step.binary` with `step.args...` (`gate.go:818`), so the argv is what the table's third column shows; `name` is a display label that happens to equal the argv only for tier (i). An implementer who reads REQ-GTA-004 literally and reports the executed argv **fails AC-GTA-004 on fixtures B and C**; one who reports `name` passes the AC while under-reporting what ran, silently dropping `-- --passWithNoTests`. This is the same class of defect as iter-1's D2 — an expected value that does not follow from the requirement — relocated into the fix rather than eliminated by it, and it will stop run-phase for a decision the SPEC should already have made. **Required fix** (one clause, either direction): narrow REQ-GTA-004 to "shall name the resolved step form" so the name field is the specified vehicle, or change the AC's expected values for B and C to the full argv. The first is the likelier intent — a summary line wants to be readable — but that is the author's call, not the auditor's.

**D8.** Minor citation drift in the new tables — `acceptance.md` AC-GTA-003 Given header and AC-GTA-004 table — Severity: **minor** — Class: **optional** — Four spans are off by one at a boundary, though each still lands on the construct it names: AC-GTA-003's header says the five paths live at `gate.go:778-815` while its own row (e) cites `:806-816` (the inner `return true, ""` is at 814, its closing brace at 815, the outer at 816); AC-GTA-004 cites tier (i) as `:683-689` where the `if` opens at 684, tier (ii) as `:690-696` where the `if` opens at 691, and the pass-through as `:697` where `return step` is at 698 and 697 is the preceding brace. Nothing resolves to the wrong construct. **Required fix (optional)**: align the four boundaries.

**D9.** AC-GTA-003 does not state the sequential-guard ordering its fixtures depend on — `acceptance.md` AC-GTA-003 fixture table — Severity: **minor** — Class: **optional** — `executeStep`'s five paths are ordered guards, not independent branches: (a) at 778, (b) at 782, (c) at 787, (d) at 793, (e) at 806, first match wins. A fixture targeting path (c) must therefore use a step that survives (a) and (b) — one that is either not `optional`, or `optional` with its binary present — and fixtures (d) and (e) must clear every earlier guard in turn. Since 25 steps declare `optional:` and 11 declare `configFiles:`, the overlap is real and a fixture can silently exercise an earlier path than intended, which is the same failure shape the author correctly documented for fixture (d) and for the `disabled_steps` polarity. The omission is a consistency gap against the standard this file otherwise sets. **Required fix (optional)**: add one sentence noting that each fixture's step must clear the preceding guards.

---

## Recommendation

**PASS-WITH-DEBT.** All seven must-pass criteria hold with no regression, the aggregate rose monotonically from 0.85 to 0.92, and every iter-1 defect is resolved — verified individually against source rather than accepted from the discharge table. The two source findings the author surfaced while building the fixtures both check out, the fixture→branch mappings are all correct, the five skip paths are reachable without touching the excluded toolchain table, and the widened AC-GTA-003 kills the mutant that iter-1 found surviving. The D1 ruling — widen the requirement rather than narrow the plan — also closed a gap the requirement text had independently of the plan, since `changedExts` was never named in REQ-GTA-003 at all.

One blocking defect remains, and it is a single clause:

1. **Reconcile REQ-GTA-004 with AC-GTA-004's expected values (D7)** before Implementation Kickoff Approval. Decide whether the summary names the resolved step form or the executed argv, then make the requirement and the three expected values agree. Fixture A is unaffected either way.

D8 and D9 are optional and left to the orchestrator's discretion.

This is iteration 2 of the Tier M ceiling of 2, so no further audit iteration is available under the retry contract. Given that D7 is a wording reconciliation with no design uncertainty and no structural consequence, discharging it does not warrant an iteration-3 exception: apply the clause, and let the run-phase AC verification confirm it against the implementation.

---

## Residual risk

**This verdict rests on a single auditor.** Per the lead's instruction the cross-model second opinion was not re-attempted; the iter-1 attempt is a known tooling defect being tracked separately (codex audited the main checkout's uncommitted changes rather than this worktree, and `audit_multi` reported `disagreement_flag: false` regardless, which reads as corroboration where none existed). No backend independently confirmed any finding in this report.

The scope narrowing is itself a risk boundary worth naming: this audit covers the `10e252834 → f8186b172` delta and the source claims it newly depends on. Findings outside that surface were not re-derived, so iter-1's residual risks stand unchanged — AC-GTA-008's Windows half remains unverifiable until the CI Windows matrix run is observed, and AC-GTA-009's PID-keyed liveness probe carries a negligible-but-nonzero PID-reuse false-negative window.

One further note on D7's blast radius: it is confined to AC-GTA-004 and reachable only on the Node toolchain path. The Go path this repository's own gate exercises resolves no substitute at all, so the divergence will not surface in this project's self-hosted runs — which is precisely why it is worth fixing at plan time rather than discovering from a Node user's summary output.

# Plan-Phase Audit — SPEC-DOCTOR-STAT-SEAM-001 (card t563)

- Auditor: plan-auditor (independent)
- Iteration: 1/1 (Tier S ceiling)
- Tree: worktree `.claude/worktrees/t563`, branch `WT-doctor-stat-shim`, HEAD `45b590c24`, base `ef10a2524` (verified ancestor of HEAD; origin/develop tip has since advanced to `9dddac882`)
- Artifacts read: `spec.md` + `plan.md` + `acceptance.md` + `progress.md` (Tier S input contract; acceptance.md present per the orchestrator's explicit deliverable list, deviation documented in progress.md §E.1)
- Reasoning context from the SPEC author: none supplied. No M1 isolation note required.

## Verdict: PASS-WITH-CONDITIONS

Overall Score: **0.94** (Tier S PASS threshold 0.75 — exceeded)

Conditions (all doc-level corrections; resolve before run-phase entry):
1. **D1 (major)** — plan.md §B "Everything the swap changes in site B is currently unpinned" is factually false; rewrite the premise and re-scope M1's framing.
2. **D2 (minor)** — plan.md §C.1 pre-flight command expectation is already false on this branch.
3. **D3 (minor)** — REQ-005 / plan M1 "no seam injection is available" is false for the userHomeDirFn seam and contradicts M1's own unresolvable-home fixture.
4. **D4 (minor)** — M1 exit selector `'StaleSkill|SkillMirror'` matches no existing test name; pin the new-test naming convention or the exit gate can sweep zero.

D5/D6 are optional and do not gate run-phase entry.

---

## Must-Pass Results

- **[PASS] MP-1 REQ number consistency**: REQ-001..REQ-007 sequential, zero gaps, zero duplicates, uniform 3-digit padding (spec.md:52-111).
- **[PASS] MP-2 EARS/GEARS format compliance** (judged against the requirement layer, spec.md §2 — not the AC layer): REQ-001 Ubiquitous ("shall invoke … shall not remain", spec.md:52-56); REQ-002 Event ("When the mirror-check loop follows … the doctor shall stat", :58-63); REQ-003 Event (:65-73); REQ-004 Unwanted ("shall not change", :75-80); REQ-005 Event ("When the seam-swap commit lands … shall already exist", :82-91); REQ-006 Where ("Where a test overrides `osStatFn` … shall be observable", :93-101); REQ-007 Where (:103-111). All seven match GEARS patterns; no informal modality.
- **[PASS] MP-3 YAML frontmatter validity**: all 12 canonical fields present with canonical names and correct types (spec.md:1-15) — `id` matches `^SPEC-[A-Z][A-Z0-9]+-[0-9]{3}$`, quoted semver `version: "0.1.0"`, `status: draft`, ISO dates, `priority: P2`, `phase: "v3.2.0"` (release target, not a stage name), `lifecycle: spec-anchored`, CSV string `tags`. No rejected snake_case aliases. `moai spec lint .moai/specs/SPEC-DOCTOR-STAT-SEAM-001/spec.md` → `✓ No findings — all SPEC documents are valid` (this run).
- **[N/A] MP-4 Section 22 language neutrality**: single-language (Go) SPEC — auto-pass per the MP-4 scoping rule.
- **[PASS] MP-5 D7 cross-SPEC reconciliation**: mechanical extraction `SPEC-([A-Z][A-Z0-9]+-)+[0-9]+` over all three artifacts yields only the self-ID (verified by grep, this run). The t540 reference is written as "SPEC-CODEX-SKILLS-PRUNE-SERIES (t540)" — not an ID-regex token — and its not-landed status is explicitly declared with a motivation-only discipline (spec.md:130-136, §3.3). No D7 BLOCKING finding. See D5 (optional) for the dangling-reference note.
- **[PASS] MP-6 D8 cross-platform discipline**: `grep -c 'syscall'` across spec.md/plan.md/acceptance.md → 0/0/0 (this run). D8-4 auto-PASS.
- **[PASS] MP-7 clarification gate**: `grep -rn 'NEEDS CLARIFICATION' plan.md` → no matches (exit 1, this run); `research.md` does not exist (Tier S) — N/A per the MP-4 precedent.

## Category Scores (rubric-anchored)

| Dimension | Score | Rubric Band | Evidence |
|-----------|-------|-------------|----------|
| Clarity | 0.75 | 0.75 — minor ambiguity in one requirement, resolvable consistently | REQ layer is crisp (exact symbols + line pins throughout). One defect: REQ-005's "(no seam injection is available before the swap)" (spec.md:87-88) reads as all-seams but means STAT-seam-only — see D3. AC-SEAM-002's "git stash-free check" phrasing (acceptance.md:22-23) is loose but decidable. |
| Completeness | 1.0 | 1.0 — all sections + frontmatter complete | HISTORY (spec.md:19-23), §1 Problem, §2 Requirements, §3 Scope, §4 Out of Scope with three `### Out of Scope — <topic>` H3 headings each carrying `-` bullets (:138-153), §5 cross-refs. Frontmatter 12/12. 7 REQ / 8 AC within the Tier S ceiling (8/8). |
| Testability | 1.0 | 1.0 — every AC binary-testable, no weasel words | All 8 ACs decidable by a named command or file read (grep assertions, `git log`, `git show`, evidence inspection). Weasel-word grep (`appropriate\|adequate\|reasonable\|proper`) over spec.md + acceptance.md → 0 matches (this run). AC-SEAM-005 pins the identity subject explicitly (finding set + each finding's text + detail strings). |
| Traceability | 1.0 | 1.0 — bidirectional, no orphans | REQ-001→AC-001/006 · REQ-002→AC-003 · REQ-003→AC-004 · REQ-004→AC-005/006 · REQ-005→AC-002 · REQ-006→AC-003/004 · REQ-007→AC-007/008. Every AC header cites its REQ (acceptance.md:8-71); every REQ covered; 7/7 and 8/8. |

---

## Ground-Truth Verification (evidence-bearing format)

**Claim**: The SPEC's §1 measured facts and all cited code anchors hold on this tree (HEAD `45b590c24`).

**Evidence** (verbatim outputs, this run, this tree):
- `grep -n 'os\.Stat(' internal/cli/doctor_codex.go` → exactly two lines: `459:  _, serr := os.Stat(entryPath)` and `857:  _, serr := os.Stat(statPath)`.
- `grep -n 'os\.Lstat(' internal/cli/doctor_codex.go` → `441: info, lerr := os.Lstat(entryPath)`.
- `grep -c 'osStatFn' internal/cli/doctor_codex.go` → `0`.
- `grep -n 'osStatFn' internal/cli/{update_preserve_inventory,codex_skills_prune,codex_skills_disable}.go` → definition `update_preserve_inventory.go:59` (`var osStatFn = os.Stat`) with doc comment at `:44` and `@MX:WARN` at `:53`; consumers `codex_skills_prune.go:96`, `codex_skills_disable.go:149` and `:168`, `update_preserve_inventory.go:439`. Matches spec.md §5 verbatim.
- `codexStaleSkillFinding` defined `doctor_codex.go:815`, called `:319`; zero references in any `*_test.go` (grep over `internal/cli/*.go`, this run) — the "ZERO direct tests" claim is literally true.
- Site A three-arm handling matches REQ-002 exactly (read :441-465); site B bucket structure matches REQ-003 exactly (read :815-880: per-`Enabled`-state missing arms, `indeterminate`, `relativeCount`/`oddlyFormed` `continue` before the stat, directory-resolves comment).
- Override patterns confirmed at `codex_skills_prune_test.go:55-57` and `update_preserve_partial_test.go:84-86` (save → replace → `t.Cleanup` restore).
- `moai spec lint .moai/specs/SPEC-DOCTOR-STAT-SEAM-001/spec.md` → `✓ No findings`.

**Baseline-attribution**: all greps/reads executed in this run against `WT-doctor-stat-shim` @ `45b590c24` (worktree-isolated session). `git merge-base --is-ancestor ef10a2524 HEAD` → ancestor confirmed; `git rev-parse --short origin/develop` → `9dddac882` (tip advanced past the SPEC's pinned `ef10a2524` — the SPEC's dual attribution "measured against origin/develop blob ef10a2524, re-confirmed in this worktree" remains sound because the re-confirmation on this tree is what carries the claim).

**Gaps**: not measured — the M1/M2 test suites (not yet written); t540 artifacts (absent from the tree entirely — see D5); `golangci-lint`/`go vet` baselines (run-phase concern, not plan-phase).

**Residual-risk**: line numbers may shift if `origin/develop` advances before run-phase entry; the SPEC's anchors were re-confirmed at `45b590c24` but AC-SEAM-001's `:441` survival assertion should be re-checked at M2 time on the then-current tree.

---

## Defects Found

**D1. Existing site-B coverage mischaracterized as "unpinned" — plan.md:21-23 (§B), spec.md:43-48 (§1), plan.md:66-78 (M1) — Severity: major — Class: blocking.**
plan.md §B asserts "Everything the swap changes in site B is currently unpinned." False. `codexStaleSkillFinding` is exercised indirectly and comprehensively through `checkCodexWiring` by the existing suite (all real fixtures, `t.TempDir()`): `TestCodexSkillPath_{HomeRelativeExistingNotMissing:896, HomeRelativeMissingStillCounted:917, RelativeNotMissingDistinctClassification:942, RelativeOnlyNonDestructiveFinding:970, BackslashAndOtherUserHomeNotMissing:993, AbsoluteExistingAndMissing:1020, RealMissingWithSymlinkLoopIndeterminate:~1051, ExpansionUsesUserHomeSeam:1090}` plus `TestCheckCodexWiring_{StaleHomeSkillsReported:247, EmptyPathEntryNotCountedMissing:311, UnspecifiedEnabledReportedSeparately:363, NonBooleanEnabledCountedSeparatelyInStaleSplit:399, IndeterminateStatNotMissing:498, DirectoryPathNotMissing:535}` in `doctor_codex_test.go`. Every M1 bullet bucket — resolves (file AND directory), per-`Enabled`-state missing, home-relative expanded, relative, oddly-formed, indeterminate (via a real symlink loop), empty path — already has an existing indirect fixture. The mirror-check half is likewise covered (`TestCheckCodexWiring_Mirror*` family, :1259-1508). What is genuinely unpinned is narrower and is exactly what M2 adds: (a) the stat **argument** (no test observes which path was stat'ed) and (b) portable stat-failure injection (site B currently needs a symlink-loop workaround). "Zero direct tests" (spec.md:43) is literally true — no `*_test.go` calls the function by name — but the §1 sentence "the t540 arm-A red had to be scoped down to the prune site for exactly this reason" conflates the two: the t540 scoping reason was the **missing seam** (no injection), not the absence of direct tests. Required fix: rewrite plan §B bullet 1 and the spec §1 consequence sentence to acknowledge the existing indirect suite; reframe M1's value as *struct-level direct pinning* (finer-grained than the existing rendered-Message/Detail assertions — rendered strings can coincide while the `codexFinding` struct differs) rather than filling a void; and note that AC-SEAM-005's identity proof is **strengthened** by the existing suite, since the pre-flight `-run 'Codex'` baseline already includes it. Run-phase consequence if unfixed: M1 is scoped as new-coverage work and silently duplicates ~7 existing fixtures.

**D2. Pre-flight base expectation already false on this branch — plan.md:31 (§C.1) — Severity: minor — Class: blocking.**
§C.1 says "Confirm worktree state before first commit: `git rev-parse --short HEAD` → base `ef10a2524`". On the current graph HEAD is `45b590c24` (the plan-phase artifact commit already landed on the branch) and origin/develop has advanced to `9dddac882`. Executed as written, the pre-flight fails and misleads the implementer into diagnosing a wrong worktree. Required fix: reword to an ancestry check — `git merge-base --is-ancestor ef10a2524 HEAD` (exit 0) plus `git branch --show-current` = `WT-doctor-stat-shim` — and record that the measured facts were re-confirmed at `45b590c24`.

**D3. "No seam injection is available" is false for the userHomeDirFn seam and contradicts M1's own fixture list — spec.md:87-88 (REQ-005), plan.md:73 (M1) — Severity: minor — Class: blocking.**
REQ-005's parenthetical "(no seam injection is available before the swap)" and M1's "no seam injection exists yet (that is the point of M1)" are literally false as general claims: the `codexUserHomeDir`/`userHomeDirFn` seam exists (`doctor_codex.go:682-695`, used by `expandCodexHomeRelativePath:687`) and is already overridden by tests (`stubCodexHome`; `TestCodexSkillPath_ExpansionUsesUserHomeSeam` doc comment names it). M1's own bullet "unresolvable-home → indeterminate" is drivable **only** by overriding that seam — with real fixtures alone it is unreachable, because a failed home resolution also fails `codexUserSkillConfig`'s `resolveCodexHomeDir` (`:714-717`) and the whole finding silently skips (`ok=false`) before any entry loop runs. Intended meaning is clearly STAT-seam-only. Required fix: scope both sentences to the stat seam explicitly ("no **stat-seam** injection…"), and state that M1 MAY override `userHomeDirFn` (serial, same save/replace/`t.Cleanup` pattern) to drive the unresolvable-home arm.

**D4. M1 exit selector can sweep zero — plan.md:78 (§F M1 Exit) — Severity: minor — Class: blocking.**
The M1 exit gate is `go test ./internal/cli/ -run 'StaleSkill|SkillMirror' -count=1 -timeout 1800s`. Verified: no existing test name in `internal/cli` contains `StaleSkill`, and `SkillMirror` appears in exactly one name (`TestResolveCodexSkillMirrorPathPublishesLiteralMirrorPath`, `codex_skills_disable_test.go:248`). The selector's non-vacuity therefore depends entirely on the not-yet-written M1 test names, which the plan does not pin — an empty-sweep green (`ok … 0.00s`) would read as M1 exit success (verification-completeness rule §1.1, empty-swept-set hazard). Required fix: pin the M1 test-name convention in plan M1 (e.g., `TestCodexStaleSkillFinding_*` and `TestInspectSkillMirror_*`) so the selector cannot match zero, and require the exit evidence to include the swept-test count, not just the exit code.

**D5. t540 reference dangles — spec.md:132, 162 — Severity: minor — Class: optional.**
"SPEC-CODEX-SKILLS-PRUNE-SERIES (t540)" has no directory under `.moai/specs/` (only `SPEC-CI-FLAKE-SERIES-001` matches "SERIES") and no t540 artifacts exist anywhere in the tree; the cited "§G gap 5" is unverifiable today. The SPEC handles this honestly — §3.3 declares t540 not-landed, restricts the reference to motivation and §5, and Out of Scope forbids any deliverable claim — so this is not a D7 BLOCKING (the name also carries no `-NNN` suffix, so the mechanical SPEC-ID extraction finds nothing). Optional fix: when t540's SPEC materializes under its final ID, update the §5 cross-reference; until then the motivation-only discipline is correctly encoded and no commit message or comment claims gap closure (verified: none present).

**D6. Artifact-set/tier mismatch, documented — spec.md:14 (`tier: S`) vs the presence of acceptance.md — Severity: minor — Class: optional.**
Tier S convention is 2 artifacts (spec.md + plan.md, AC inline in spec.md §3); this SPEC carries acceptance.md with AC-SEAM-001..008 and has no inline AC section. The deviation is explicit and orchestrator-ordered (progress.md §E.1: "acceptance.md included per the orchestrator's explicit deliverable list"), lint reports nothing, and the AC layer is arguably stronger separated. No action required; recorded so the next tier judgment treats this as a settled precedent rather than an oversight.

---

## Regression Check

N/A — iteration 1.

## Recommendation

All four blocking defects are prose-level corrections in plan.md and spec.md; none changes the SPEC's direction, its AC set, or its feasibility. The change itself (two-line seam swap + characterization-first ordering + output-identity proof) is well-conceived: the GEARS layer is clean, the AC set is mechanically judgeable, the ordering witness (AC-SEAM-002) is correctly specified (characterization commit strictly before the seam commit, both carrying `t563`, with M1's passing run as evidence), and the grep AC (AC-SEAM-001) provably distinguishes `os.Stat` from `os.Lstat` (case + prefix; verified zero comment mentions of `os.Stat` in the file) and is paired with the Lstat-survival assertion so the two assertions together prove discrimination. The identity AC (AC-SEAM-005) pins the comparison subject (finding set + text + detail strings) and the mechanism (M1 suite unmodified + pre-flight baseline on shared fixtures) — concrete, not taste. Every card-t563 hard rule has a covering AC: scope-settled (§3.1 + AC-SEAM-001), observability-not-behavior (AC-SEAM-005), characterization-first with commit-graph witness (AC-SEAM-002), scoped verification (AC-SEAM-007), t540 motivation-only (§3.3 + §4 + plan §G).

Route D1-D4 back to manager-spec as a single annotation-cycle revision; re-audit is scoped to the D1-D4 delta only (plan.md §B/§C.1/§F-M1, spec.md §1/REQ-005).

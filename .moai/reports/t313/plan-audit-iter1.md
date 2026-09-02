# SPEC Review Report: SPEC-WORKTREE-BASEREF-001

Card: t313 · Iteration: 1/2 (Tier M ceiling) · Auditor: plan-auditor
Tree measured: `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t313`, branch `WT-worktree-baseref`, HEAD `48eb945df` (re-read by this auditor, not inherited).
Artifacts read: `spec.md` (0.2.0), `plan.md`, `acceptance.md`, `progress.md` — the full Tier M input set.

Reasoning context ignored per M1 Context Isolation. The orchestrator's dispatch supplied operator decisions D1-D4 and five pre-verified facts; the decisions were treated as settled inputs and the five facts were re-measured rather than assumed (results in § Coverage).

**Verdict: FAIL**
**Aggregate score: 0.78 (harmonic mean) / 0.80 (arithmetic mean) — Tier M PASS threshold 0.80**

The must-pass firewall is fully clear (7/7). The FAIL is carried entirely by three blocking defects that put the harmonic aggregate below the Tier M threshold: one requirement with no acceptance criterion, one measured vacuous-pass mode affecting eight MUST criteria, and one criterion that verifies less than the requirement it is mapped to. All three are cheap to repair and none requires re-litigating an operator decision. This is an actionable FAIL, not an obstructive one — an iteration 2 scoped to D1/D2/D3 should clear it.

---

## Must-Pass Results

- **[PASS] MP-1 REQ number consistency** — `grep -oh 'REQ-WBR-[0-9]\{3\}' *.md | sort -u` returns exactly `REQ-WBR-001 … REQ-WBR-016`, 16 distinct ids, sequential, no gaps, no duplicates, uniform 3-digit zero-padding. Matches the orchestrator's stated count of 16; no discrepancy.
- **[PASS] MP-2 GEARS format compliance** — all 16 entries in `spec.md` §B carry an explicit pattern label and match it. Ubiquitous: 001 (`spec.md:B.1`), 002, 014, 015. Event-driven: 004, 007, 009, 010, 012, 016. State-driven: 005, 006, 011. Unwanted: 003 (`"The shipped template default shall not name develop, main, or any other repository-specific branch"` — canonical `shall not` form). Where/capability-gate: 013 (`"Where the web console renders the git-worktree panel…"`). REQ-WBR-009 is a compound `When … shall …; and while … shall …`, which is GEARS-compound and PASS-equivalent. Judgement made against the **requirement layer** (`spec.md` §B) only; the Given/When/Then entries in `acceptance.md` are verification-layer `AC-XXX` and were graded under Group 4, never under this criterion.
- **[PASS] MP-3 YAML frontmatter validity** — all 12 canonical fields present with correct types: `id`, `title`, `version: "0.2.0"` (quoted semver), `status: draft`, `created: 2026-08-27`, `updated: 2026-08-27`, `author`, `priority: P1`, `phase`, `module`, `lifecycle: spec-anchored`, `tags` (comma-separated string). No rejected snake_case alias (`created_at` / `updated_at` / `labels` / `spec_id`) appears. Two extra keys (`tier: M`, `related_specs`) are additive, not schema violations.
- **[PASS] MP-4 language neutrality** — the SPEC is scoped to this repository's own Go codebase plus one shipped config key. It names no programming-language-specific tooling, and the one shipped artifact it touches (`git-strategy.yaml.tmpl`) is a language-agnostic config file. REQ-WBR-003 + AC-WBR-002 additionally assert the shipped default names no branch. See A7 for the one latent hazard.
- **[PASS] MP-5 D7 cross-SPEC reconciliation** — two external ids referenced: `SPEC-SYNC-STRATEGY-KEY-001` (`.moai/specs/…/spec.md` → `status: in-progress`) and `SPEC-WORKTREE-BRANCH-GUARD-001` (→ `status: completed`). Neither is `retired` / `superseded` / `archived`; both directories exist. No BLOCKING finding.
- **[PASS] MP-6 D8 cross-platform discipline** — `grep -c 'syscall'` returns 0 across all four artifacts. D8 auto-PASS; no cross-platform build-tag concern.
- **[PASS] MP-7 clarification gate** — `grep -rn '\[NEEDS CLARIFICATION' .moai/specs/SPEC-WORKTREE-BASEREF-001/` → rc=1, no match. `progress.md` independently records `blocker: none` with D2 CLOSED by the operator ruling. No open clarification marker.

---

## Category Scores (rubric-anchored)

| Dimension | Score | Rubric band | Evidence |
|-----------|-------|-------------|----------|
| Clarity | 0.90 | 1.0 → 0.75 (upper) | Every requirement has one reading. Decisions and their rejected alternatives are recorded with the reason preserved (`plan.md` §A D2.1 explicitly writes the anti-improvement note). Deduction: AC-WBR-012's `Then` clause is narrower than the REQ it maps to (D3). |
| Completeness | 0.95 | 1.0 | All required sections present. Frontmatter 12/12. Six `### Out of Scope — <topic>` H3 sub-headings each carrying a specific `-` bullet (`spec.md` §C). `plan.md` §B enumerates seven gaps (G1-G7) explicitly rather than asserting completeness, and §E carries five risks. Deduction: `plan.md` §D's write list omits any file for the AC-WBR-014 round-trip test. |
| Testability | 0.70 | between 0.50 and 0.75 | Every AC names a runnable command. But eight MUST criteria are satisfiable while executing zero tests (D2, measured), one criterion carries a judgement step (D9), and one pins a string no requirement fixes (D8). |
| Traceability | 0.65 | below 0.75, above 0.50 | Both 0.75-band exceptions occur simultaneously: REQ-WBR-004 has no AC (D1) **and** AC-WBR-014 references no REQ at all (D6). 15 of 16 REQs covered; 1 of 15 ACs orphaned. |

Harmonic mean = 4 / (1/0.90 + 1/0.95 + 1/0.70 + 1/0.65) = **0.78**. Below the Tier M threshold of 0.80. Arithmetic mean 0.80 sits exactly on it; per the skeptical-evaluation stance (`agent-common-protocol.md` § Skeptical Evaluation Stance) the harmonic mean is the binding aggregate, and a score that clears only on the more forgiving average and only by rounding is not a PASS this auditor can evidence.

---

## Defects Found

### Blocking

**D1 — REQ-WBR-004 has no acceptance criterion.** `acceptance.md:14-28` — Severity: **major** — Class: **blocking**.
The `§D AC Matrix` REQ column covers 001, 002, 003, 005, 006, 007, 008, 009, 010, 011, 012, 013, 014, 015, 016. REQ-WBR-004 — *"When a session starts, the SessionStart handler shall read the configured value and compare it against the branch currently named by `refs/remotes/origin/HEAD`"* — appears nowhere. It is not implicitly covered: AC-WBR-003/004/005/015 each pin an **outcome** for a given config state, but none asserts that the read-and-compare **happens at session start** rather than at some other moment or not at all. An implementation that only ever ran the comparison inside `moai doctor` would pass every existing AC while failing REQ-WBR-004 outright.
**If it ships as written:** the SPEC's central firing-point requirement — the one that makes this a SessionStart consumer at all — is unverifiable, and run-phase has no criterion that fails when the errgroup task is never registered.
**Required fix:** either add an AC asserting that `Handle` invokes the alignment path exactly once per session-start invocation (a seam call-count of 1 against the read seam, distinct from the write seam), or extend AC-WBR-003's Given/Then to bind the read as well as the write and re-map it to `005, 004`.

**D2 — Eight MUST criteria pass vacuously when their `-run` regex matches nothing.** `acceptance.md` AC-WBR-003, -004, -005, -006, -010, -011, -012, -015 — Severity: **major** — Class: **blocking**.
Every one of these criteria is discharged solely by a `go test … -run '<regex>'` invocation. Measured in this tree at HEAD `48eb945df`:

```
$ go test ./internal/settings/... -run 'WorktreeBaseBranch.*Type' -count=1
rc=0
ok  github.com/modu-ai/moai-adk/internal/settings          0.367s [no tests to run]
ok  github.com/modu-ai/moai-adk/internal/settings/agentfm  1.088s [no tests to run]
ok  github.com/modu-ai/moai-adk/internal/settings/yamlpatch 0.772s [no tests to run]
```

That is AC-WBR-011's own literal command, run against the current tree, exiting **0**. A run-phase actor capturing exit codes into `.moai/state/verify/t313/` would record eight green MUST criteria having executed zero assertions. The failure is silent by construction: nothing in the AC text, and nothing in `§D.3 Definition of Done`, requires a non-zero executed-test count.
**If it ships as written:** the SPEC's entire MUST evidence chain can be satisfied by an implementation that adds the config key and writes no tests at all — the exact unobserved-verification-claim shape `verification-claim-integrity.md` §1 forbids, arriving through the acceptance criteria rather than around them.
**Required fix:** add one binding sentence to `§D.3 Definition of Done` (one edit covers all eight): *"Every `-run` invocation must report at least one executed test; a `[no tests to run]` line FAILS the criterion it was run for."* Optionally pin it mechanically by requiring `-v` plus a `grep -c '^=== RUN'` floor.

**D3 — AC-WBR-012 verifies one third of REQ-WBR-015.** `acceptance.md:170-183` vs `spec.md` §B.5 — Severity: **major** — Class: **blocking**.
REQ-WBR-015 requires a regression guard *"modelled on `internal/web/dead_config_guard_test.go`"* pinning **three** properties: the key is present in `settings.AllFields()`, present in the rendered console HTML, **and** reaches a consumer. AC-WBR-012's `Then` asserts only the third (*"the git seam receives the value"*), and its command targets `./internal/hook/... ./internal/cli/...` — neither of which is where the guard lives. The first two properties are asserted by AC-WBR-010, but AC-WBR-010 is mapped to REQ-WBR-013 (the web-surface requirement), not to 015, and AC-WBR-010 asserts them as *presence*, not as a *guard that fails on later removal*. So no criterion verifies that the regression guard exists.
**If it ships as written:** REQ-WBR-015 — the anti-dead-key requirement, which exists precisely because this repository has already shipped dead config keys (`dead_config_guard_test.go` is the scar) — is discharged by the ordinary consumer test that AC-WBR-005 and AC-WBR-007 already demand. The guard need never be written.
**Required fix:** restate AC-WBR-012's `Then` as the three-part conjunction REQ-WBR-015 states, and add `./internal/web/...` to its command. A2's note applies: the criterion must assert the guard *test's* existence and failure mode, not merely that the key is reachable.

**D4 — R1's "one repository, one key" premise is false; the concurrency residual risk is under-stated.** `plan.md` §E R1, final sentence — Severity: **major** — Class: **blocking**.
R1 dismisses the two-sessions-disagreeing case with: *"two sessions configured to different values would fight… That is a misconfiguration, not a race — one repository, one key."* Measured in this tree:

```
$ git ls-files --error-unmatch .moai/config/sections/git-strategy.yaml
.moai/config/sections/git-strategy.yaml      # rc=0 — the file is TRACKED
```

A tracked file has one working-tree copy **per worktree**, and its content is branch-dependent. `refs/remotes/origin/HEAD` has one copy per **repository**. So the premise is inverted: there are as many values of the key as there are active card worktrees (eight concurrent lanes at audit time, per the dispatch), all writing one shared handle. Two lanes on branches carrying different values are not misconfigured — each is correctly configured for its own branch — and they will alternately rewrite `origin/HEAD` on every session start, each seeing a "difference" the other just created. The write-only-on-difference mitigation does **not** confine this to a single transition; it confines it to one write per divergence, and divergence is continuously re-created.
**If it ships as written:** the run-phase implementer reads R1 and concludes the concurrency window is bounded and one-shot, when under the repository's own eight-lane operating mode it is neither. `AGENTS.md` §2 and `main-checkout-branch-guard.md` both treat repository-global mutation under concurrency as a first-class hazard; this SPEC introduces an automatic one and then reasons about it from a false premise.
**Required fix:** correct R1's premise (the setting is per-working-tree, the handle is per-repository), and state the real steady-state invariant: silence holds only while every active worktree's checked-out `git-strategy.yaml` carries the same value. Then either accept the residual risk explicitly, or add a requirement narrowing consumer 1 — the cheapest narrowing is to fire the alignment only from the primary checkout (detectable via `git rev-parse --git-dir` vs `--git-common-dir`, the same discriminant `sessionWorktreeInGitWorktree` already uses at `internal/cli/session_worktree.go:56-57`), which leaves consumer 2 unaffected and removes the multi-lane write contention entirely.

### Debt (can ride)

**D5 — AC-WBR-002 and plan M1's comment guidance are in latent contradiction.** `acceptance.md:38-46` vs `plan.md` §C M1 — Severity: **minor** — Class: **optional**.
AC-WBR-002's second grep pipes the key's line into `grep -e develop -e main` and requires exit 1. M1 instructs the implementer to add the key *"with a comment naming its effect and its neutral default"*, and REQ-WBR-014 establishes the house style of naming `main` and `develop` as the two common values in prose. A perfectly reasonable template comment — `# branch that card worktrees are cut from (e.g. main, develop); empty = no action` — satisfies M1 and the spirit of 014 while **failing** AC-WBR-002. Fix: state in AC-WBR-002 that the prohibition binds the **value**, not the comment, and narrow the grep to the value side of the colon; or state in M1 that the template comment must not name any branch.

**D6 — AC-WBR-014 references no REQ (orphaned criterion).** `acceptance.md:27` — Severity: **minor** — Class: **optional**.
Its REQ column reads `R4 / G3` — a risk id and a gap id, not a requirement. It is the only AC in the set with no requirement behind it, which is consistent with it being a run-phase precondition rather than a criterion for this SPEC's own scope. Fix: either promote the round-trip preservation to a REQ (it is a real correctness property of the write path this SPEC introduces), or move it out of the AC matrix into `§D.3 Definition of Done` where the other preconditions live.

**D7 — Citation drift at REQ-WBR-008.** `spec.md` §B.2 REQ-WBR-008 — Severity: **minor** — Class: **optional**.
The requirement cites `internal/hook/session_start.go:168-171` for the best-effort contract. Measured: the migration step's best-effort comment sits at `:165-168`, and the `"Handle never returns a non-nil error from these steps"` contract line is at `:176`. `plan.md` §D3 cites `:176-181` correctly, so the spec and the plan disagree by ~8 lines about the same anchor. Fix: align REQ-WBR-008's citation to `:176-181`.

**D8 — AC-WBR-009's manual command pins a check name no requirement fixes.** `acceptance.md:150` — Severity: **minor** — Class: **optional**.
`moai doctor --check 'Worktree Base Branch'` is runnable — the flag exists (`internal/cli/doctor.go:58`) and filters by **exact** name equality (`internal/cli/doctor.go:232`: `if filterCheck != "" && c.name != filterCheck`). But REQ-WBR-012 never fixes the `DiagnosticCheck.Name` string, so an implementer naming it `Worktree Base` or `Base Branch` satisfies the requirement and fails the criterion's manual step. Fix: name the string in REQ-WBR-012.

**D9 — AC-WBR-013 carries a judgement step and a vacuous loop.** `acceptance.md:186-200` — Severity: **minor** — Class: **optional**.
Two of its three commands are not binary. `git status --short  # expect no local-only .claude/.moai file lacking a template counterpart` requires a human to decide which entries qualify. And the `.sh`/`.sh.tmpl` drift loop is followed by *"If this SPEC touches no hook wrapper, the loop reports pre-existing drift only"* — which this SPEC does not (no hook wrapper appears in `plan.md` §D), making the loop's output uninterpretable as a pass or fail signal for this change. Fix: scope both to the SPEC's own diff (`git diff --name-only` against the base), so the criterion is decidable from this change alone.

**D10 — `grep -n -A0` is a no-op flag.** `acceptance.md:41` — Severity: **minor** — Class: **optional**. Cosmetic; `-A0` adds nothing to `-n`. Drop it.

---

## A1-A7 Findings (dispatch-directed)

**A1 — Testability of every AC.** Every one of the 15 criteria names a runnable command and an outcome; none uses "inspect the code" or "confirm the behaviour". No weasel words ("appropriate", "reasonable", "adequate") appear anywhere in `acceptance.md`. The three singled out for scrutiny:

- **AC-WBR-015 (unresolvable value)** — the strongest criterion in the set. Four conjunct assertions (seam invoked 0 times; exactly one stderr line containing the offending value; nil error; `origin/HEAD` unchanged), plus two implementation-shape pins (the predicate treats only `rc == 0` as resolvable; `git show-ref --verify`, not the BranchGuard-refused porcelain form). It also states the "still resolves to the ref it named before the run" post-condition, which is what actually catches a partial write. No defect.
- **AC-WBR-014 (typed round trip)** — the Given/When/Then is precise and the failure disposition is stated ("escalated as a blocker rather than absorbed"). Two weaknesses: it traces to no REQ (D6), and `plan.md` §D lists no file for the test it demands, so the write list is incomplete against it.
- **AC-WBR-012 (anti-dead-key guard)** — it is **not** satisfiable by a grep; the AC says so explicitly and the command runs tests. But it is satisfiable by the ordinary consumer tests other criteria already require, because it asserts one of REQ-WBR-015's three properties (D3). The grep-prohibition it was written to enforce holds; the guard-existence obligation does not.

Overriding all three: D2's vacuous-pass mode, which is a property of the `-run`-regex discharge style rather than of any single criterion.

**A2 — G7 re-measured, and the plan is correct.** In this worktree at HEAD `48eb945df`:

```
$ git show-ref --verify refs/remotes/origin/develop        ; echo rc=$?    → rc=0
$ git show-ref --verify refs/remotes/origin/nonexistent-xyz              → fatal: 'refs/remotes/origin/nonexistent-xyz' - not a valid ref
                                                            ; echo rc=$?  → rc=128
```

`plan.md` §B G7 and AC-WBR-015's inline measurement both state 128, both are right, and the REQ-WBR-009 predicate's `rc == 0` test is therefore correctly specified. A predicate written against `rc == 1` would misclassify every missing ref. No defect. (Also re-measured: `git symbolic-ref refs/remotes/origin/HEAD` → `refs/remotes/origin/develop`, confirming §A.2.)

**A3 — Backward compatibility, argv level: consumer 2 holds; consumer 1 holds.**
Measured at `internal/cli/session_worktree.go:217-218`:

```go
func gitWorktreeAddReal(destDir, branch string) (string, error) {
	cmd := exec.Command("git", "worktree", "add", "-b", branch, destDir)
```

AC-WBR-008 asserts the recorded argv is exactly `["git","worktree","add","-b",<branch>,<dest>]` and adds *"an extra trailing empty-string operand FAILS"* — which is precisely the failure mode a naive `append(args, base)` would produce. That is argv-level, not resolved-commit-level, and it matches the measured invocation byte for byte. REQ-WBR-011 states the same at requirement level. **No defect** — this is the criterion set's strongest backward-compat guarantee.
For consumer 1, REQ-WBR-005 + AC-WBR-003 bind the unset path to "no process spawned, nothing on stderr, `origin/HEAD` unchanged", which is the equivalent no-op guarantee for a path that has no argv of its own. Note one asymmetry worth recording though not a defect: AC-WBR-003 asserts *spawn count 0* against a fake seam, so it verifies the guard, not the absence of any other git invocation the helper might add.

**A4 — Blast radius: the mitigation is real but the residual risk is materially UNDER-stated.** See D4. In detail: the write-only-on-difference rule (REQ-WBR-006) genuinely does what R1 claims of it in a **single-working-tree** repository — steady state involves no writes, and the window opens once per configuration change. R1's error is the premise it applies that rule under. `.moai/config/sections/git-strategy.yaml` is tracked (measured, `git ls-files --error-unmatch` rc=0), so each of the eight active card worktrees carries its own working-tree copy whose content follows its own branch, while `refs/remotes/origin/HEAD` is single and repository-global. Under that topology "difference" is not a one-shot event: lane A's write creates the difference lane B's next session start will observe and reverse. R1's closing dismissal — *"one repository, one key"* — is the one sentence in the plan that is factually wrong, and it is load-bearing, because it is what converts a live concurrency hazard into an accepted "misconfiguration". A secondary under-statement: any external actor (`git remote set-head -a`, a manual reset, a fresh clone) also re-creates the difference, so even a single-lane repository sees repeated writes rather than one. The recommended narrowing (fire consumer 1 only from the primary checkout) is stated in D4's fix; it costs nothing the SPEC currently claims, since the handle is repository-global and one writer suffices.

**A5 — Consumer parity: the plan binds it, the requirements do NOT.** `plan.md` §A D4 is explicit and well-argued: *"'Unresolvable' is decided by the SAME predicate consumer 1 uses (REQ-WBR-009 …), so the two consumers cannot disagree… Implement the predicate once and call it from both; a second, divergent resolvability rule is the defect this note exists to prevent."* §C M2 repeats it (*"a single exported helper consumed by both M2 and M3"*). But the requirement layer does not carry it. REQ-WBR-009 scopes itself to *"the alignment step"* — consumer 1 only, in both its `When` and its `while` clause. REQ-WBR-010 and REQ-WBR-011 say only *"unresolvable as a git ref"*, naming no predicate and cross-referencing REQ-WBR-009 nowhere. And no AC binds them: AC-WBR-008 asserts consumer 2's argv outcome for an unresolvable value without constraining **how** unresolvability was decided, while AC-WBR-015's two implementation-shape pins (`rc == 0` only; `show-ref --verify`, not the porcelain form) are written against consumer 1 alone.
So a run-phase implementation that resolves consumer 2 with a second, divergent rule — `git rev-parse --verify`, a `git branch --list` scrape, or a local-branch check — passes every criterion. This is exactly the outcome D4's note exists to prevent, escaping through the gap between the plan and the requirements. **Severity: this is the finding most likely to actually bite in run-phase**, because the plan's prose reads as though the obligation is already discharged. Recorded here rather than in the blocking list only because the fix is one clause: extend REQ-WBR-011 to name REQ-WBR-009's predicate as the sole resolvability authority for both consumers, and extend AC-WBR-008's `Then` to assert that consumer 2's unresolvable determination came from that shared helper (a call-count or a shared-seam assertion, not a behavioural equivalence).

**A6 — Scope discipline: clean.** `git status --porcelain` in this worktree returns exactly one entry, `?? .moai/specs/SPEC-WORKTREE-BASEREF-001/` — no tab_schema.json copy touched, no source file modified at plan time. The `plan.md` §D write list contains 16 entries, all inside the declared surface (one config key, two consumers, doctor item, web field, guards, docs/template parity); `internal/web/assets/i18n.js` is the only entry not named in the dispatch's scope summary and it is a direct consequence of the web field, not creep. The shared schema type set is explicitly preserved: D2.1 rules that *"the `FieldType` set at `internal/settings/schema.go:105-113` stays exactly as it is — no eighth type is added"*, which I verified is the live shape (`TypeSelect` :106 … `TypeBool` :112, block :105-113, `TypeText` at :109 exactly as REQ-WBR-014 cites). No TUI change is proposed. `spec.md` §C carries four explicit non-repair exclusions (`automation.auto_branch` spelling, the `moai update` revert hazard, the `ModeProfile` gaps, retroactive worktree repair) — each an adjacent defect the SPEC observed and declined, which is the correct disposition. **No defect.**

**A7 — Template neutrality: asserted at both layers, with one latent hazard.** REQ-WBR-003 (Unwanted) forbids the shipped default from naming any repository-specific branch, and AC-WBR-002 tests it with a grep over `internal/template/templates/.moai/config/sections/git-strategy.yaml.tmpl`. Both layers are present, satisfying the dispatch's requirement. Verified the current template state: the file (measured, lines 1-9) carries `mode` / `provider` / `github_username` / `gitlab.instance_url` as Go-template placeholders and does **not** yet carry the new key — consistent with pre-implementation, and consistent with G5's observation that no plain-file mirror exists. `plan.md` §C M1 correctly says the local file MAY carry `develop` while the template MUST NOT, and D2.1's rejection of the `TypeSelect` alternative is argued on exactly the right ground (`internal/settings/` is template-managed and reaches every downstream project; a `main`/`develop` option set would be unusable for a `trunk` user). The one hazard is D5: AC-WBR-002's grep binds the whole line, not the value, so a comment written in the house style would fail it.

---

## Coverage Statement

What this audit actually checked, so an empty finding is distinguishable from an unlooked-at surface.

**Measured in this tree (commands run, output observed):**
- HEAD, branch, and toplevel re-read: `48eb945df`, `WT-worktree-baseref`, worktree confirmed.
- REQ/AC distinct-id counts: 16 / 15 — matches the dispatch, no discrepancy.
- `git show-ref --verify` exit codes for an existing and a missing ref (A2): 0 / 128.
- `git symbolic-ref refs/remotes/origin/HEAD` → `refs/remotes/origin/develop`.
- Vacuous `-run` exit code (D2): AC-WBR-011's own command, rc=0, `[no tests to run]` ×3 packages.
- `git ls-files --error-unmatch` on the config file (D4): tracked, rc=0.
- `git status --porcelain` (A6): one untracked entry, the SPEC directory.
- D7 referenced-SPEC statuses: `in-progress`, `completed`.
- D8 `syscall` occurrence count: 0.
- MP-7 `[NEEDS CLARIFICATION` scan: rc=1.

**Citations verified against source (read, line numbers confirmed):**
`internal/settings/schema.go:105-113` (FieldType block; `TypeText` at :109 — the SPEC's line numbers are right, the dispatch's `106-112` counts content lines only); `internal/cli/uikit/types.go:11-18` (three-value CheckStatus enum); `internal/cli/doctor.go:30-35` (DiagnosticCheck shape), `:58` (`--check` flag), `:220` (`Worktree State` registration), `:232` (exact-match filter), `:876-891` (`checkWorktreeState` model); `internal/cli/session_worktree.go:51-53` (seam), `:181-196` (`materializeSessionWorktree`), `:217-219` (`gitWorktreeAddReal` argv); `internal/hook/session_start.go:66` (`Handle`), `:94` (stderr precedent), `:120-175` (four-task errgroup), `:176` (best-effort contract — see D7); `internal/settings/schema_sections.go:160-178` (`gitStrategyFields`); `internal/settings/sectionapply.go:171-206` (`applyGitStrategyKey` — confirmed a root-level key requires a case in the first switch, exactly as M5 states, since the fallthrough `strings.Cut` path errors on a dotless key); `internal/web/schemaform.go:226-230` (git-worktree panel consuming `SectionFields(SectionGitStrategy)`); `internal/config/types.go:106-132` (`ModeProfile`, confirmed lacking DevelopBranch / ReleaseBranchPrefix / RCVersionFormat), `:143-167` (`GitStrategyConfig`); `.moai/config/sections/git-strategy.yaml:1-6` and `:15-17` (the three unmodelled `manual.*` keys, present as G3 states); `internal/template/templates/.claude/rules/moai/workflow/worktree-integration.md:171` and its local mirror at the same line (the `"accepts only \"fresh\" or \"head\", not arbitrary refs"` sentence, verbatim in both); `internal/web/dead_config_guard_test.go` (schema half, render half, and the `workflow.worktree.*` neighbour list — all within ±2 lines of the cited ranges).

Every citation checked resolved to what the artifact claimed, with the single exception recorded as D7.

**NOT checked (gaps in this audit):**
- `moai doctor` was not executed (the SPEC's own G4 records the same gap; the `--check` flag and its exact-match filter were read, not run).
- No Go test was run other than the deliberate zero-match probe for D2; no build, vet, or lint was performed — plan-phase audit, no implementation exists.
- `EnterWorktree`'s actual read of `origin/HEAD` was not verified from source (G2 states the same; Claude Code is not in this repository). The SPEC's premise that consumer 1 has any effect rests on inference, which G2 discloses honestly and REQ-WBR-012's doctor item is the stated fallback for.
- The reflog line quoted in §A.2 (`branch: Created from origin/develop`) was not re-measured; G1 already marks it inherited.
- The rendered attribute order of a `TypeText` control was not measured (G6 discloses this, and AC-WBR-011's two-branch condition plus its run-phase collapse instruction is the correct disposition — no defect).

**Residual risk in this verdict:** the four blocking defects are all defects of specification, not of design. The design itself — one stored key, a consumer that derives the metadata mutation from it, a shared resolvability predicate, and a fail-open contract — survived scrutiny; D1-D3 are gaps between what the requirements say and what the criteria verify, and D4 is a wrong premise in a risk analysis whose mitigation is otherwise sound. I did not find a reason to doubt the approach. A reader should not read this FAIL as a judgement on the plan's substance.

---

## Recommendation

Iteration 2 should be scoped to the four blocking defects; the six debt items may ride or be folded in opportunistically.

1. **D1** — cover REQ-WBR-004. Add an AC asserting the read-and-compare fires once per session start (read-seam call count 1), or extend AC-WBR-003 to bind the read and re-map it.
2. **D2** — add one sentence to `acceptance.md` §D.3: a `-run` invocation reporting `[no tests to run]` FAILS the criterion it discharged. This single edit repairs all eight affected criteria.
3. **D3** — restate AC-WBR-012's `Then` as REQ-WBR-015's three-part conjunction and add `./internal/web/...` to its command, so the regression guard's existence is verified rather than assumed.
4. **D4** — correct R1's premise (per-working-tree config vs per-repository handle; the tracked-file measurement is in this report), state the real steady-state invariant, and either accept the eight-lane residual explicitly or narrow consumer 1 to the primary checkout.
5. **A5 (fold into D-level)** — bind REQ-WBR-011 to REQ-WBR-009's predicate as the sole resolvability authority for both consumers, and extend AC-WBR-008 to assert consumer 2 used the shared helper. Without this, `plan.md` §A D4's parity note is unenforceable.

Nothing in this list re-opens an operator decision. D1-D4 and A5 are all additions or corrections within the settled D1-D4 rulings; the `TypeText` ruling, the both-consumers scope, the SessionStart+doctor surfacing, and the t316 boundary are all faithfully implemented by the artifacts as written and were not questioned.

No stagnation assessment applies (iteration 1). Tier M ceiling is 2 iterations: one revision cycle remains before escalation.

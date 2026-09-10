# SPEC Review Report: SPEC-IGNORED-EVIDENCE-CITATION-001

Iteration: 1/2 (Tier M ceiling, `harness.plan_audit_tier_ceilings`)
Verdict: **FAIL**
Overall Score: **0.74** (harmonic mean) — Tier M PASS threshold **0.80**

Reasoning context ignored per M1 Context Isolation. The dispatch's framing was treated as a
hypothesis and re-measured; where it was wrong, that is recorded in the final section.

Audit tree: `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t381`, branch `WT-ignored-evidence-cite`,
HEAD `3f03d9c36` (`= origin/develop`). Every measurement below was taken in this tree, in this run.

---

## Must-Pass Results

- **[PASS] MP-1 REQ number consistency** — `REQ-IEC-001` … `REQ-IEC-009` at spec.md:130, 138, 146,
  152, 159, 167, 176, 184, 190. Sequential, no gap, no duplicate, uniform 3-digit zero-padding.
- **[PASS] MP-2 GEARS format compliance** — judged against the **requirement layer** (`REQ-IEC-*` in
  `spec.md`), never against an AC. All nine match a GEARS pattern: Unwanted `shall not`
  (001 "The repository **shall not** carry", 006 "The repair **shall not** modify", 009 "The repair
  **shall not** modify any file owned by"), event-driven `When` (002 spec.md:140, 007 spec.md:178),
  `Where` capability-gate (003 spec.md:148, 004 spec.md:154), `Where`+`When` compound (005
  spec.md:161-164), ubiquitous (008 "This SPEC **shall** carry", spec.md:186 — `<subject>` as artifact
  is permitted). The `Given/When/Then` entries in `acceptance.md` are the **verification layer** and
  are graded under Group 4 only.
- **[PASS] MP-3 YAML frontmatter validity** — spec.md:1-16 carries all 12 canonical fields with
  correct types: `id`, `title`, `version: "0.1.0"` (quoted semver), `status: draft`,
  `created: 2026-08-31`, `updated: 2026-08-31`, `author`, `priority: P2`, `phase: "v3.1.4 target"`,
  `module`, `lifecycle: spec-anchored`, `tags` (comma-separated string). No rejected snake_case alias
  (`created_at` / `updated_at` / `labels` / `spec_id`) present. `phase` is a release-target label, not
  one of the prohibited lifecycle tokens. Optional `related_specs` and `tier: M` are permitted.
- **[N/A] MP-4 Section 22 language neutrality** — single-language scope. The SPEC touches Go source
  and two report artifacts in one repository; it makes no multi-language tooling claim and names no
  language-specific tool as primary. Auto-passes per the MP-4 N/A rule.
- **[PASS] MP-5 D7 cross-SPEC reconciliation** — verb executed.
  `grep -Eoh 'SPEC-([A-Z][A-Z0-9]+-)+[0-9]+' spec.md | sort -u` → three ids.
  `SPEC-HIERARCHICAL-TEAM-001` → `grep '^status:'` = `completed` (not retired/superseded/archived; no
  reconciliation owed). `SPEC-EVIDENCE-CITATION-CANON-001` → `ls -d .moai/specs/SPEC-EVIDENCE-CITATION-CANON-001`
  = *No such file or directory* ⇒ D7-5 SHOULD, **not** BLOCKING: the SPEC discloses the absence
  explicitly (spec.md:67-73 "That branch is **UNMERGED** — absent from `origin/develop`") and I
  confirmed it — `git cat-file -e origin/develop:.moai/specs/SPEC-EVIDENCE-CITATION-CANON-001/spec.md`
  → `fatal: path … does not exist in 'origin/develop'`, rc=128. Self is the third id. No BLOCKING
  finding.
- **[PASS] MP-6 D8 cross-platform discipline** — verb executed. `grep -c syscall` = 0 on spec.md,
  plan.md, acceptance.md. D8-4 auto-PASS.
- **[PASS] MP-7 clarification gate** — `grep -rn 'NEEDS CLARIFICATION' .moai/specs/SPEC-IGNORED-EVIDENCE-CITATION-001/`
  → rc=1, no match. `plan.md` present; no `research.md` (correct at Tier M). The §A.4 REQ-ECC-004
  boundary question is carried as a **named open question routed to the lead** with a stated
  conservative default (plan.md:86-88), which is the correct handling — it is not an unresolved
  `[NEEDS CLARIFICATION]` marker.

**All seven must-pass criteria pass.** The FAIL is driven by the rubric scores and the blocking
defect list below, not by the firewall.

---

## Category Scores (rubric-anchored)

| Dimension | Score | Rubric Band | Evidence |
|---|---|---|---|
| Clarity | 0.85 | 0.75 | Per-file treatments are individually argued (spec.md:232-270), not swept; the anti-pattern list (plan.md:135-145) names the sweep explicitly. One genuine internal contradiction: acceptance.md:12-13 asserts "No AC rests on reading prose for a judgment" while AC-IEC-001 (acceptance.md:57-60, 71-72) decides by a human reading a `grep -C2` window. |
| Completeness | 0.95 | 1.0 | HISTORY (spec.md:20-24), Context (§A), Requirements (§B), Scope (§C), Constraints (§D), Cross-references (§E) all present. `grep -c '^### Out of Scope'` = **5**, each with `-` bullets (spec.md:349, 363, 369, 374, 380) — satisfies `OutOfScopeRule`. Probe boundary disclosed in full: `grep -c "\| B1 \|"` … `B6` = 1 each. |
| Testability | 0.55 | 0.50 | 7 of 12 MUST commands do not execute in the worktree the card runs in (D1). 1 of 12 decides by prose (D2). 3 of 12 are green at arrival and unfalsifiable by any run-phase act (D3). Only AC-IEC-002, 006, 007 both run here and carry a decision. |
| Traceability | 0.75 | 0.75 | All nine REQ-IEC ids have a covering AC (001→001, 002→002, 003→003, 004→004, 005→005, 006→006, 007→010, 008→008, 009→007). But 3 of 12 ACs cite a document **section** rather than a REQ id (acceptance.md:44, 46, 47), and AC-IEC-012 is the sole coverage of one of the five in-scope files (D6). |

Harmonic mean of {0.85, 0.95, 0.55, 0.75} = **0.7434 → 0.74**. Tier M threshold 0.80 (per
`spec-workflow.md` § SPEC Complexity Tier). **0.74 < 0.80 ⇒ FAIL.**

---

## Defects Found

**D1. Seven of twelve MUST acceptance commands cannot execute in the worktree this card runs in.**
`acceptance.md`:L62-68 (AC-IEC-001), L96-98 (AC-IEC-003), L113-117 (AC-IEC-004), L131-135
(AC-IEC-005), L206-207 (AC-IEC-008), L222-225 (AC-IEC-009), L271-272 (AC-IEC-012).
Each uses a `for … done` loop and/or `$(…)` command substitution. Measured in this
worktree-isolated session, this tree, this run — both forms are refused before execution:

```
$ for f in internal/cli/mcp_glm.go internal/cli/audit_pin_live_test.go; do echo "### $f"; done
This session is isolated in the worktree /Users/goos/MoAI/moai-adk-go/.claude/worktrees/t381,
but this command is too complex to verify that it stays inside the worktree. Refusing to run it

$ printf 'oos=%s\n' "$(grep -c '^### Out of Scope' .../spec.md)"
… Refusing to run it
```

The refusal is not git-specific — the loop above contains no git. It is the same worktree-isolation
guard the SPEC itself cites at spec.md:72-73 and progress.md:24 as refusing cross-tree `git -C`, so
the card already knows this guard is live in its own execution environment and did not carry the
knowledge into its verification layer. `acceptance.md`:L12-13 states "Every AC below is decided by a
**named command whose output decides it**" — for these seven the named command does not run, so
under `verification-completeness.md` §1.2(c) their failure is unreachable: nothing executes, so
nothing can go red. Severity: **critical**. Class: **blocking**.
*Required fix*: rewrite all seven as plain single invocations — one `grep -c <pattern> <file>` per
line, no loop, no `$(…)`, no `for`. Do **not** move the batch into a script file: the guard cannot
read inside a script, which bypasses every check for that payload (`kanban-dispatch.md` § The
env-isolated verification form). Re-verify each rewritten command actually runs in this worktree
before adopting it.

**D2. AC-IEC-001 decides by prose judgment, contradicting the verification contract it sits under.**
`acceptance.md`:L57-60 defines PASS as "every occurrence sits within two lines of an explicit
non-resolution marker"; L61-68 emits a `grep -n -C2` window and computes nothing; L71-72 asks a
reader to judge adjacency. `acceptance.md`:L12-13 asserts the opposite as the contract for the whole
file. This is the criterion covering REQ-IEC-001 — the defect the card exists to repair — so the
card's central gate is the one that is not mechanical. Severity: **major**. Class: **blocking**.
*Required fix*: make adjacency computable, e.g. compare the count of `.moai/state/` occurrences
against the count of marker-bearing `-C2` windows and require equality, each as its own plain
invocation; or amend `acceptance.md` §A to state honestly that one criterion is a reviewed read.

**D3. Three MUST ACs are green at arrival, green at exit, and no run-phase action can falsify them.**
`acceptance.md`:L198-211 (AC-IEC-008), L213-228 (AC-IEC-009), L230-245 (AC-IEC-010). All three
assert properties of plan-phase artifacts that already hold. Measured now, before any repair:
`probe=1`; `B1..B6 = 1` each; `grep -c '^### Out of Scope' spec.md` = **5**; `grep -c 'class C4'` =
**1**; `grep -c 're-produced'` = **1**; `grep -c 'tee .moai/state' acceptance.md` = **0**;
`git check-ignore -v .moai/reports/t381/verify` → rc=1 (not ignored). None of milestones M1-M5
(plan.md:78-129) touches `spec.md` or `acceptance.md`, so no mutation reachable from this card's
own work flips any of them — the always-green shape of `verification-completeness.md` §1.2(a), and
they can never acquire a RED-now cell (§2). Severity: **minor**. Class: **optional**.
*Required fix (optional)*: demote from MUST to a recorded structural check, or re-scope AC-IEC-010
to assert the *evidence directory as actually written at M5* rather than the AC text.

**D4. The cited `status:` of SPEC-EVIDENCE-CITATION-CANON-001 is false at audit time.**
`spec.md`:L67 states "v0.3.0, `status: draft`, `tier: M`"; repeated at L403. Measured now at
`/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t375/.moai/specs/SPEC-EVIDENCE-CITATION-CANON-001/spec.md`:
`version: "0.3.0"` (L4, correct), **`status: in-progress`** (L5, cited as `draft`), `tier: M` (L14,
correct). §D:L393-394 anticipates exactly this ("t375's canon is a moving target … re-read, not
assumed") but attaches the discipline only to the *wording* of the requirements, not to the status
attribute, so the stale value stands unqualified in two places. On a card whose whole subject is
citation integrity, a cited attribute that is false as stated is in-domain. Severity: **minor**.
Class: **blocking**.
*Required fix*: drop the `status:` attribute from the citation (it is load-bearing for nothing this
card decides), or re-read it at M1 and pin the read with a date as §A.4 already does for the
requirement text.

**D5. The do-not-touch list contains a path that does not exist in this tree.**
`spec.md`:L302-303 lists "the machine-emitted `.codex/agents/moai/manager-lead.toml`" among the
t375-owned files this card must not modify, and `acceptance.md`:L186 diffs that same path.
Measured: `git ls-files '.codex/agents/moai/manager-lead.toml'` → empty;
`ls .codex/agents/moai/` → *No such file or directory*. The real emitted artifact is
`internal/template/templates/.codex/agents/moai/manager-lead.toml` — confirmed by the 25-line census
probe, which places all seven C4 lines (54, 78, 136, 139, 140, 143, 150) there and nowhere else.
`git diff --stat 3f03d9c36 -- '.codex/agents/moai/manager-lead.toml'` exits 0 with empty output, so
the AC does not break — it simply asserts nothing for that entry. §C.5's parenthetical "(Verified
absent in this tree.)" attaches only to the guard test file, so this entry was never checked. A
collision-safety list carrying one unverified entry is the same unobserved-premise defect the card
repairs. Severity: **minor**. Class: **blocking**.
*Required fix*: delete the top-level `.codex/…` entry from `spec.md` §C.5 and `acceptance.md`
AC-IEC-007, or re-point it to the `internal/template/templates/.codex/…` path already listed.

**D6. Three MUST ACs have no requirement backing, and one of them is the sole gate on a fifth of the
card's scope.** `acceptance.md`:L44 (AC-IEC-009 → "§C.8"), L46 (AC-IEC-011 → "§D
no-behavior-change"), L47 (AC-IEC-012 → "§C.3 stale-coordinate repair"). All three cite a document
section rather than a `REQ-IEC-*` id. The consequential one is AC-IEC-012: the repair of
`.moai/reports/t299/verify-sync/e2e-lint-4paths.extract.txt` — one of the five in-scope files, and
the whole of milestone M4 (plan.md:119-124) — is mandated by `spec.md` §C.3 prose and by no
requirement. REQ-IEC-001 does not reach it: I measured `grep -ci 'gitignore'` = 1 on that file, and
its header already declares non-resolution (`# … That path is gitignored (.gitignore:284) and
vanishes with the worktree, so this extract is the citable carrier`), so AC-IEC-001 passes for that
file today, repaired or not. Severity: **major**. Class: **blocking**.
*Required fix*: add a requirement covering stale evidence coordinates (a `REQ-IEC-010` in the
Unwanted form — a tracked citation shall not name a line number without naming the tree it was
resolved in), and re-point AC-IEC-012 at it. For AC-IEC-009 and AC-IEC-011, either back them with
requirements or reclassify them as constraint checks below the MUST line.

**D7. The `467` figure is stated without its derivation.** `spec.md`:L330 gives blind spot B2 as
"467 lines". `census.md` §1 measures `git grep -n '\.moai/state' -- . ':!*.md'` = **492**; the 467 is
492 − 25 and the subtraction appears in neither document. Every other figure in the SPEC names its
command (the 25-line probe, the 531, the 506 shown as `531 − 25`). Severity: **minor**. Class:
**optional**. *Required fix*: write it as `492 − 25 = 467` with the command, matching the B1 row.

**D8. The lane-8 convergence claim is asserted as observed fact with no source named.**
`spec.md`:L120-124 — "Lane-8 reached the same repair split … on a different file
(`manager-lead.md:150`) and without coordination". I verified the *target* is real:
`sed -n '150p' .claude/agents/moai/manager-lead.md` carries "The path MUST resolve at audit time …",
the exact analogue of the `audit_pin_live_test.go:32` clause. But what lane-8 *decided* is a claim
about another card's internal reasoning, and the SPEC names no way it was learned — while the same
section correctly marks the relayed `b64043481` SHA as unverified. Severity: **minor**. Class:
**optional**. *Required fix*: mark it relayed, as the SHA already is; the argument it supports (the
discriminator is real, not one author's framing) survives being labelled second-hand.

**D9. The extract.txt brace-glob exemption is argued against REQ-ECC-004 alone, not against the
004+005 pair.** `spec.md`:L259-261 leaves the brace-glob standing because "under REQ-ECC-004 it names
the sources of an export, not the evidence for a verdict". Read verbatim at
`.claude/worktrees/t375/.moai/specs/SPEC-EVIDENCE-CITATION-CANON-001/spec.md`:168 and the note
following it, REQ-ECC-005 and 004 are stated to "form one ceiling" precisely because a broadly-written
citation satisfied both while exporting everything. The exemption may well be correct — the extract
does carry per-source sha256 and the decision-bearing lines inline — but the SPEC does not engage the
clause written to close that shape. Severity: **minor**. Class: **optional**. *Required fix*: add one
sentence engaging REQ-ECC-005, or re-confirm at M1 once t375's wording settles.

---

## What I verified and found sound (recorded so it is not re-litigated)

These were checked adversarially and hold. They are cited here because a re-audit should not spend
budget on them again.

- **Every in-scope coordinate resolves.** The 25-line probe `git grep -n '\.moai/state/verify' -- . ':!*.md'`
  returns exactly 25 in this tree, and places all five C3 citations at the lines §C.1 claims:
  `mcp_glm.go:110`, `audit_pin_live_test.go:32`, `evidence_writer_zeroexec_test.go:10`,
  `e2e-lint-4paths.extract.txt:3`, `template-skill-improvement-plan-20260710.html:684`.
- **`.gitignore:298`** — `git check-ignore -v .moai/state/verify/t225` →
  `.gitignore:298:.moai/state/`; `ls .moai/state/verify/` → *No such file or directory*. The SPEC
  states the coordinate **and** names the tree it was measured in (spec.md:33-35), which is the
  hygiene D7/D8 above ask for elsewhere.
- **All six cited t375 line numbers are exact**: REQ-ECC-001 at :164, 002 at :165, 003 at :166,
  004 at :167, 006 at :169, 007 at :175. Verified by `grep -n` on the t375 worktree file.
- **"Outside the guard, inside the canon" is SOUND.** REQ-ECC-007 (:175) reads, verbatim, "저장소는
  doctrine 표면 하위 `.md`에서 `.moai/state/verify` 인용을 검출하는 가드를 가져야 한다" — the guard is
  scoped on **both** axes (`.md` **and** doctrine surface), exactly as spec.md:85-90 states, and all
  five in-scope citations are `.go`/`.txt`/`.html` under `internal/` or `.moai/reports/`. REQ-ECC-001
  and REQ-ECC-003 carry no doctrine-surface qualifier — confirmed by reading them. The card's reason
  for existing is verified, not assumed.
- **Carve-out arithmetic is right.** `git grep -c '\.moai/state/verify'` over the six carve-out files
  sums to exactly **12** (store.go 1, schema.go 1, store_test.go 3, events.go 1,
  evaluate_snapshot_test.go 3, ignored_content_test.go 1, two `.html` 1 each), and
  `git diff --stat 3f03d9c36 --` over them is empty.
- **AC-IEC-006 and AC-IEC-007 are well-built measured negatives — do not "fix" them.** Each is
  mutation-detectable by its own command: editing a carve-out makes the diff non-empty; *deleting*
  a carve-out file is caught by the paired positive control (`total=12`); creating t375's guard file
  is caught by the `ls` half of AC-IEC-007. These two are the counter-example to D3, and the
  distinction matters: a measured negative is vacuous only when nothing this card could do would
  flip it.
- **SPEC-HIERARCHICAL-TEAM-001 citations are accurate** — `spec.md:103`, `:161`, `:227` and
  `acceptance.md:9` each carry the `.moai/state/verify/<session>/` instruction, as spec.md:376-378
  claims, and the SPEC's `status:` is `completed`.
- **`e2e-lint-4paths.extract.txt` is 75 lines** as stated, and its header does declare its own
  non-resolution.

---

## Regression Check

N/A — iteration 1.

---

## Recommendation

FAIL. Four blocking defects (D1, D2, D5, D6) and one blocking-minor (D4) must be fixed before the
verdict is revisited. The five optional findings (D3, D7, D8, D9 and the optional half of D3) are
surfaced for the orchestrator's discretion and **must not** be routed into revision wholesale —
this SPEC is already unusually disciplined and over-correcting it would add speculative structure.

Fix in this order:

1. **D1 first — it is the load-bearing one.** Rewrite AC-IEC-001, 003, 004, 005, 008, 009, 012 as
   plain single invocations with no `for` and no `$(…)`, then *run each one in this worktree* and
   paste the observed output into `acceptance.md` alongside the command. Until that is done, the
   card's verification layer is a set of commands nobody has executed where they must execute.
2. **D2** — make AC-IEC-001's adjacency test computable, or amend `acceptance.md` §A:L12-13 to stop
   claiming what the file does not do.
3. **D6** — add the missing requirement for the stale-coordinate repair and re-point AC-IEC-012 at
   it; decide whether AC-IEC-009/011 stay MUST.
4. **D5** — remove or re-point the phantom `.codex/agents/moai/manager-lead.toml` entry in
   `spec.md` §C.5 and `acceptance.md` AC-IEC-007.
5. **D4** — drop or re-date the `status: draft` attribute at `spec.md`:67 and :403.

The must-pass firewall is clean and the scope discipline is sound; the failure is concentrated in
the verification layer. A revision addressing D1, D2, D5, D6 should clear the Tier M threshold
comfortably — Testability at 0.55 is what pulls the harmonic mean under 0.80, and it is the only
dimension below its band.

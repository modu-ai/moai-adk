# SPEC Review Report: SPEC-TODO-LANDING-STATE-001 (half A)

Iteration: 2/2 (Tier M ceiling — `harness.yaml` `plan_audit_tier_ceilings.M: 2`; there is no iteration 3 for this tier)
Verdict: **PASS**
Overall Score: **0.85** (Tier M threshold 0.80) — up from 0.74

Reasoning context ignored per M1 Context Isolation. Audited from `spec.md`, `plan.md`,
`acceptance.md` (Tier M input contract) plus the source tree, at worktree
`/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t331`, branch `WT-card-landing-state`,
HEAD `e1d480eba`, working tree clean (`git status --short` → empty).

**PASS is not a clean bill.** Three blocking defects are enumerated in §Defects Found. Because
the Tier M iteration ceiling is exhausted, they route as a **scoped delta fix before Implementation
Kickoff Approval**, not as another audit round. D1 in particular is a narrower recurrence of
iteration 1's D5 class, closed by the author on one mutant shape and reopened here on a second.

---

## Scope-split honesty — the first-class question

**The split was performed honestly.** Four independent checks, all measured:

1. **The destination exists and carries the moved defects.** `moai todo` → card `t359`, state
   `queued`, whose text explicitly carries iteration 1's D1 (the `REQ-TODO-013` misattribution, with
   the correct reading — `SPEC-KANBAN-TODO-CLI-001/spec.md:59`) and D2 (the Table 2 / REQ-1.10
   conflict) forward as preconditions. The moved half did not evaporate into a promise.
2. **§B.2's pointer relocates the question rather than asserting the misattributed freeze.** It
   reads "*what `REQ-TODO-013` permits, how `SPEC-TODO-ANALYSIS-001` read it* … belong to t359 and
   are deliberately not argued here" — question-form, and it names **both** sides of the disputed
   reading. §D's per-item-field bullet cites `backlog_store.go:62-63` as *naming* the five-field
   contract (the code comment there does say "frozen"), then explicitly refuses to rule:
   "Whether that contract is extensible at all is t359's question." I verified the canonical text
   independently — `REQ-TODO-013` says "changing it only **additively**", and
   `SPEC-TODO-ANALYSIS-001/spec.md:51` (completed) already ruled the freeze reading wrong. Neither
   surviving passage asserts the freeze. This is a correct relocation.
3. **Nothing in the narrowed artifacts depends on the moved half.** All 11 requirements were read
   against the source. `REQ-TLS-007` (queue `state` on the `todo pr` row) is the only one that could
   have needed storage; it does not — the value is already loaded at `internal/cli/todo_pr.go:167-169`
   (`text` map built from `rec.Items`), and `internal/kanban/backlog_store.go:53-60` already carries
   the three-value enum. `acceptance.md` §E gate 5 ("nothing under `internal/kanban/backlog_*.go` is
   modified") is consistent with the whole requirement set, not in tension with it.
4. **The REQ-1.10 conflict is genuinely gone from this document.** `spec.md:305-310` now states the
   opposite of iteration 1's Table 2: "**What S2 does NOT gain: the delivering commit's name**",
   citing `SPEC-KANBAN-QUEUE-PR-SYNC-001` REQ-1.10 (verified verbatim at that SPEC's `spec.md:251-255`)
   and its two enforcement sites (`prlink_landed.go:62-67`, `prlink.go:42-43` — both re-opened and
   confirmed).

**One consequence of the split is not honest, and is D2 below**: the C4-C7 row collapse is correct on
the S2 column (all four genuinely answer `landed` through one mechanism after the change) and its
deferral of the run-vs-sync distinction is stated plainly. But the collapse also depends on the S3
column reading "*retired as a decision input*" in every row of Table 2 — and nothing in this SPEC
retires it.

---

## Must-Pass Results

- **[PASS] MP-1 REQ number consistency** — `REQ-TLS-001` … `REQ-TLS-011`, sequential, no gap, no
  duplicate, uniform 3-digit padding. Definitions at `spec.md:456,461,466,469,475,478,485,490,493,497,501`.
- **[PASS] MP-2 GEARS format compliance** — judged against the **requirement layer** (`spec.md` §C)
  only; the `AC-TLS-*` entries in `acceptance.md` are the verification layer and were graded under
  Group 4, not here. All 11 match a GEARS pattern: Where-compound (`REQ-TLS-001`), ubiquitous
  (`002`, `007`, `009`, `010`, `011`), ubiquitous+unwanted (`003`, `005`), event-driven
  (`REQ-TLS-004`, "**When** the resolved landed ref does not resolve"), state-driven
  (`REQ-TLS-006`, "**While** `--require-landed` is set"), unwanted (`REQ-TLS-008`, "shall not
  change … shall not close, archive, drop").
- **[PASS] MP-3 YAML frontmatter validity** — all 12 canonical fields present with correct types
  (`spec.md:1-19`): `id`, `title`, `version: "0.2.0"` (quoted semver), `status: draft`,
  `created`/`updated` ISO `2026-08-28`, `author`, `priority: P1`, `phase`, `module`,
  `lifecycle: spec-anchored`, `tags` (comma-separated string). No rejected snake_case alias
  (`created_at` / `updated_at` / `labels` / `spec_id`) present. `tier: M` additionally declared.
- **[N/A] MP-4 language neutrality** — single-language (Go) SPEC scoped to `internal/kanban`,
  `internal/cli`, `internal/config`. Auto-passes. Note: `plan.md` §B.7 correctly carries the
  template-neutrality CI guard obligation for the `REQ-TLS-010` mirror edit; the workflow file exists
  (`.github/workflows/template-neutrality-check.yaml`, 3689 bytes).
- **[PASS] MP-5 D7 cross-SPEC reconciliation** — five SPEC ids referenced. Statuses read directly:
  `SPEC-KANBAN-QUEUE-PR-SYNC-001` `in-progress`, `SPEC-TODO-DESTRUCTIVE-GUARD-001` `completed`,
  `SPEC-TODO-ANALYSIS-001` `completed`, `SPEC-WORKTREE-BASEREF-001` `completed`,
  `SPEC-KANBAN-TODO-CLI-001` `in-progress`. None is `retired` / `superseded` / `archived`. No
  BLOCKING finding.
- **[PASS] MP-6 D8 cross-platform discipline** — `grep -c 'syscall' spec.md` → `0`. Auto-PASS per D8-4.
- **[PASS] MP-7 clarification gate** — `grep -rn '\[NEEDS CLARIFICATION' .` across the SPEC
  directory → rc=1, zero matches. Iteration 1's three markers are gone: two ruled in `plan.md` §C,
  one retired with the B half. Grounding of the two rulings judged separately (D5 below).

---

## Category Scores (rubric-anchored)

| Dimension | Score | Rubric Band | Evidence |
|---|---|---|---|
| Clarity | 0.90 | 0.75-1.0 | Every ref-dependent figure travels with its command, both ref SHAs, and an observation instant (`spec.md:82-98`); Table 1/2 carry their own row and column legends (`:242-256`); the C4-C7 collapse is argued rather than asserted (`:291-303`). One residual ambiguity, D2: Table 2's S3 cell "*retired as a decision input*" (`:282`) does not say by what. |
| Completeness | 0.85 | 0.75-1.0 | HISTORY (`:23-28`), WHY (§A), WHAT (§A.6/§B), REQUIREMENTS (§C), Out of Scope (§D — five `### Out of Scope — <topic>` H3 sub-headings, each with specific `-` bullets), Traceability (§E). Tier M artifact set complete (spec/plan/acceptance). Docked for D2: the state model promises an S3 change no requirement delivers. |
| Testability | 0.70 | 0.50-0.75 | Eight of ten ACs carry a RED I independently re-measured or re-derived from source. But **two ACs carry a clause that cannot fail**: AC-TLS-008 for `REQ-TLS-009`'s "no cache" (D1, proved with a live mutant) and AC-TLS-009 for `REQ-TLS-011` (D3, already satisfied by `todo.md:59-67`). Rubric band 0.50-0.75: more than one AC requires judgment to evaluate as a release gate. |
| Traceability | 0.95 | 0.75-1.0 | `spec.md` §E (`:583-595`) verified in both directions against `acceptance.md`'s "Verifies" column: 11 REQs each with ≥1 AC, 10 ACs each naming an existing REQ, zero orphans. Iteration 1's orphaned AC closed by `REQ-TLS-007` (`:485-486`). Docked 0.05 because `REQ-TLS-009`'s four-clause prohibition is only partly reachable from its sole AC. |

Aggregate = (0.90 + 0.85 + 0.70 + 0.95) / 4 = **0.85**.

---

## Independent verification performed

### Provenance (the pin the whole document rests on)

`git diff --name-only 3de2f85a2 HEAD` → returns exactly five paths, all this SPEC's own artifacts
plus `.moai/reports/t331/plan-audit.md`. **No cited source file changed between the pinned tree and
HEAD**, so every `@ 3de2f85a2` pin resolves at HEAD. The claim at `spec.md:30-36` is true.

### Exhaustive citation sweep (iteration 1 D7 / D18)

Every `file:line` citation in all three artifacts was extracted mechanically and re-opened.
**All resolve, and all carry the content claimed.** The load-bearing ones:

| Citation | Verified content |
|---|---|
| `prlink_landed.go:28` | `const LandedRef = "origin/main"` |
| `prlink_landed.go:26-27`, `:52` | the two stale doc comments named in `plan.md` M1 |
| `prlink_landed.go:44`, `:78` | the `git log` ref operand and the error string |
| `prlink_landed.go:62-67`, `:76-79` | REQ-1.10 enforcement; the `(false, err)` map on rc≠0 |
| `prlink.go:18-21`, `:30-48`, `:31`, `:42`, `:50-51` | purity ruling; four kinds; two stale comments; the "closed set" comment belongs to `PRLinkConfidence`, exactly as `plan.md` R6 says |
| `prlink.go:175-185` / `:179-181` | the landed leg: `landed.Landed()` at `:178`, `if isLanded` at `:182` |
| `todo.go:20`, `:357`, `:399`, `:417-431`, `:426-428` | subagent boundary; three text sites; **both `return nil` paths (`:423`, `:430`) write only to `cmd.ErrOrStderr()`** — the stdout-identity claim is established by the code, as `acceptance.md` A.4 says |
| `todo_pr.go:1-16`, `:8-13`, `:57-65`, `:75`, `:87`, `:135`, `:141-144`, `:147-151`, `:167-169`, `:175-176` | read-only property; gh budget; runner seam; two text sites; five columns `CardID, Kind, PRs, Confidence, text` — **none is the queue state**, exactly as AC-TLS-010 claims |
| `todo_undone_test.go:277-278`, `:287-303`, `:297`, `:306-335`, `:326-331` | `:297` is `strings.HasPrefix(stdout, "done t1")` — `plan.md` §C.1's "a suffix does not disturb a prefix" is mechanically correct; `TestTodoDone_NoLandingQueryWithoutTheFlag` discards stdout (`_, _, err :=`) and indeed never inspects it |
| `todo_pr_test.go:142-199`, `:177`, `:193` | `TestTodoPR_QueueDirUnchanged`, four sub-cases, digest + engine mtime + lock mtime; the two error strings quoted verbatim in `acceptance.md` §D.1 are at exactly those lines |
| `backlog_sqlite.go:113`, `:95-97` | the `CHECK (state IN ('queued','picked','dropped'))` and the ALTER-CHECK rationale |
| `backlog_store.go:53-60`, `:62-63` | the three-value enum; "The five fields are the frozen per-item contract (REQ-TODO-013)" |
| `config/types.go:164-174`, `loader_worktree_base.go:28-35` | the neutral-empty ruling; the resolver |
| `hook/worktree_base_branch.go:70-82`, `:156`; `session_worktree.go:215`; `doctor_worktree_base.go:40-43` | the resolvability authority and **all three existing consumers** M1 claims |
| `todo.md:51`, `:53-57`, `:59-67` | four outcomes; the [HARD] operator-only rule (quoted **verbatim** in `spec.md` §B.4 — I diffed the quote); the permissive-policy note with its stated reason |
| `git-strategy.yaml:5`, `:16`, `merge_method: squash` (`:22`) | `worktree_base_branch: develop`, `develop_branch: develop`, squash confirmed |
| `SPEC-KANBAN-QUEUE-PR-SYNC-001/spec.md:251-255` | REQ-1.10 verbatim |

The mirror-parity claim in AC-TLS-009 also holds: `diff` of the two `todo.md` copies → rc=0,
**241 lines each**, exactly as stated.

**Iteration 1's D7/D18 class is closed.** This is the first iteration in which I found no
misaddressed citation.

### Ref-dependent figures re-measured, not re-cited (iteration 1 D8)

`git fetch origin main develop`, then, observed **2026-08-28T15:58Z**:

```
origin/main    = 48239c7dc7428c8751a04f6321887c2d36123884
origin/develop = 48d8ef4bee768645d1a14a53eb5e4ba85170d447
git rev-list --count --left-right origin/main...origin/develop  →  0   375
git merge-base --is-ancestor origin/main origin/develop          →  rc=0
```

The right column has moved again — `329 → 334 → 349 → **375**` across four measurements — which is
precisely the hazard `spec.md:38-43` warns about and the reason D8's remediation was correct. The
load-bearing claim (**left column zero**, `main` a strict ancestor of `develop`) holds.

Per-card and per-input probes, re-run:

| Probe | Result | SPEC's claim |
|---|---|---|
| `t293` on `origin/main` / `origin/develop` | `0` / `9` | `0` / `9` ✓ |
| `t338` on `origin/develop` | `0` | `0` ✓ (AC-TLS-010's input) |
| `merge-base --is-ancestor 7fc161b36 origin/develop` | rc=`1` | not an ancestor ✓ |
| `merge-base --is-ancestor 294b4b6ab origin/develop` | rc=`0` | landed as the squash ✓ |
| `git log origin/no-such-ref --grep='\bt331\b'` | rc=`128`, `fatal: ambiguous argument` | ✓ (AC-TLS-004's input) |

### Blast radius re-measured

`grep -rn 'LandedRef' --include='*.go' internal/` → **12 lines**, matching the stated total. The
**seven production sites** are correct and complete: `todo_pr.go:75,:87`; `todo.go:357,:399,:428`;
`prlink_landed.go:44,:78`. Iteration 1's D10 ("six") is genuinely corrected. The line-class
breakdown, however, does not sum — see D4.

`plan.md` R6's measurements are correct in full: `PRLink` outside tests and `internal/kanban/prlink*`
appears in **exactly one file** (`internal/cli/todo_pr.go`, lines 135, 141, 176, 181, 182);
`internal/web` holds **108** Go files and **zero** `PRLink` references (rc=1). Iteration 1's D15 is
closed with real numbers.

### AC-TLS-008 — my own mutant, of a shape the author did not test

Baseline confirmed GREEN first:
`go test ./internal/cli/ -count=1 -run TestTodoPR_QueueDirUnchanged -v` → `--- PASS` with all four
sub-cases PASS. `shasum internal/cli/todo_pr.go` → `a80ca6bdf8cd61f278310befdc400e547aa00d04`,
corroborating the author's stated pre/post SHA-1 prefix `a80ca6bd`.

**Mutant A — a landing cache written outside the queue state directory.** `REQ-TLS-009` forbids
"no field, no finding, **no cache**, no lock". `queueDirDigest` (`todo_pr_test.go:110-137`) walks
only `kanban.StateDirForRoot(root)` = `<root>/.moai/state/<queue-dir>`. I inserted an unconditional
call in `runTodoPR` writing `<root>/.moai/cache/landing-sweep.cache` on every invocation.

Result: **`TestTodoPR_QueueDirUnchanged` stayed fully GREEN, 4/4 sub-cases PASS.**

Liveness proved rather than assumed — a probe test read the file back:

```
zz_audit_probe_test.go:21: MUTANT LIVE - wrote .../001/.moai/cache/landing-sweep.cache with contents:
    t1=no-link
    t2=no-link
--- PASS: TestAuditProbe_MutantCacheIsWritten (0.10s)
```

**Mutant B — the evasion the author named but did not test** (a write restored to the same bytes
within one invocation): write-all-dropped then restore, both through `store.Mutate`.

Result: **caught**, 4/4 sub-cases FAIL — `queue directory changed across the invocation`,
`backlog.json mtime moved`, and `backlog.lock mtime moved: … — the read verb took the lock`. SQLite
does not reproduce the file byte-for-byte, and the lock mtime moves regardless. The author's stated
residual worry is closed by measurement; credit where due.

**Tree restored.** `cp` back, probe test removed, `shasum` → `a80ca6bd…` (identical),
`git status --short` → empty, and `TestTodoPR_QueueDirUnchanged` → `ok` again.

**Judgment: the criterion's threat model is conveniently drawn, not honestly bounded.** It is drawn
at the *queue directory*, while `REQ-TLS-009`'s prohibition is drawn at *the verb*. The gap is not
theoretical — it swallowed a live cache write on my second attempt, against a non-goal the SPEC
itself calls "the most plausible slip for the next person reading this design" (`spec.md:398-401`).

### Declared-gap honesty

`acceptance.md` §C: release-blocking = AC-TLS-001, 003, 004, 005, 006, 008, 009, 010 (**8**);
regression-guard = AC-TLS-002, 007 (**2**). "Seven from measurement, and AC-TLS-008 from a planted
mutant" = 8 — the arithmetic is correct.

The **regression-guard classification is honest**. Both AC-TLS-002 (empty ⇒ `origin/main`) and
AC-TLS-007 (`unknown` ⇒ archive, exit 0) are preservation criteria: they have no *pre-change* RED
because nothing about them is broken today, yet both **can** fail after the change if M1 or M3
breaks the preserved behaviour. Declaring them regression-guards rather than dressing them with an
invented failure is exactly right, and `acceptance.md:7-10` states the rule it is applying.

All seven measurement REDs are genuine and were independently confirmed above. The gap statement
is honest **except** on AC-TLS-009, where the RED cell establishes failure for the `REQ-TLS-010`
content but not for the `REQ-TLS-011` clause — see D3.

### Factual correction carried

Confirmed and **not repeated**: `.moai/reports/` is not gitignored in this repository. Iteration 1's
closing note was wrong on that point.

---

## Defects Found (structured defect-list)

**D1.** `acceptance.md:75` (AC-TLS-008) + `spec.md:493` (REQ-TLS-009) — the sole criterion for
`REQ-TLS-009` cannot fail on that requirement's "no cache" clause. The criterion is scoped to the
**queue directory** (`queueDirDigest` walks only `kanban.StateDirForRoot(root)`,
`internal/cli/todo_pr_test.go:110-137`), while the requirement prohibits writes by **the verb**. I
planted a live mutant writing `<root>/.moai/cache/landing-sweep.cache` on every `todo pr`
invocation; the test stayed GREEN 4/4 and a read-back probe confirmed the file was written. This is
a narrower recurrence of iteration 1's D5 class, which the author's own single-shape mutant probe
did not reach. — Severity: **major** — Class: **blocking** — Required fix: widen AC-TLS-008's
assertion beyond the queue directory so it can detect a write anywhere the verb could reach (for
example, digest the whole project fixture root, or add a sub-case asserting no file is created
outside the queue directory across the invocation), and re-establish its RED by planting a
write **outside** `StateDirForRoot` — not only inside it. Alternatively narrow `REQ-TLS-009` to the
queue record and move "no cache / no lock" to a separately-verified requirement; do not leave the
requirement broad and the criterion narrow.

**D2.** `spec.md:280-289` (Table 2, S3 column, every row) — Table 2 asserts that hand ancestry is
"*retired as a decision input*" after this SPEC, but **no requirement delivers it and no criterion
verifies it**. Measured: `grep -i 'ancestr'` over `spec.md` §C (lines 447-505) → rc=1, zero matches;
over `plan.md` §D milestones (lines 99-162) → rc=1; over `acceptance.md` → one match, at line 55, in
prose about the input's provenance, not in any AC. `REQ-TLS-010`'s doctrine list is
(resolved ref, `unknown` outcome, stdout verdict token, [HARD] rule) and `REQ-TLS-011` states the
remaining limit — neither retires ancestry. This matters beyond bookkeeping: the C4-C7 collapse
(`spec.md:291-296`) is justified by the row legend's "separately observable" rule, and in Table 1
those rows were separated partly *by their S3 cells* (C6 `NOT-ANCESTOR` vs C7 `ANCESTOR`). If S3 is
not actually retired, the collapsed row is not observationally single. `plan.md:180` calls Table 2
"the state table this plan implements", so an unimplemented column is a live inconsistency.
— Severity: **major** — Class: **blocking** — Required fix: either add a requirement that retires
hand ancestry as a decision input (most naturally as a clause on `REQ-TLS-010`, with AC-TLS-009
extended to assert the doctrine says so) and cite it from Table 2's S3 header, or change the S3
cells to state what this SPEC actually leaves them as and re-argue the C4-C7 collapse on the S2
column alone.

**D3.** `acceptance.md:76` (AC-TLS-009 RED cell) + `spec.md:501-503` (REQ-TLS-011) — AC-TLS-009 is
the sole criterion for both `REQ-TLS-010` and `REQ-TLS-011`, and its RED establishes failure only
for the `REQ-TLS-010` content ("Both copies @ `3de2f85a2` describe four outcomes and a two-valued
landed check"). `REQ-TLS-011`'s substance is **already present in the doctrine today**:
`.claude/skills/moai/workflows/todo.md:59-67` reads "It asks whether ANY commit on the landed ref
names the card — not whether the card's LAST step landed — so a card whose run commit shipped reads
as landed while its sync commit is still sitting unpushed", which is `REQ-TLS-011` almost word for
word. As written the requirement is a no-op and the criterion cannot go red on it. This is the same
class as D1 and as iteration 1's D5, and it contradicts `acceptance.md:7-10`'s own rule that a
criterion without an observed RED is classified as a regression-guard — AC-TLS-009 is classified
release-blocking at `acceptance.md:83`. — Severity: **major** — Class: **blocking** — Required fix:
sharpen `REQ-TLS-011` to the delta that is actually missing (the limit is stated today only under
`--require-landed`; extend it to the `todo pr` `landed` outcome, and to the new `unknown` outcome),
then state that delta in AC-TLS-009's RED cell with the `todo.md:59-67` quote shown as the *existing*
coverage so the gap is visible. If no delta survives that exercise, fold `REQ-TLS-011` into
`REQ-TLS-010` rather than keeping an unfailable requirement.

**D4.** `spec.md:350-352` (§B.1 M3) — the stated line-class breakdown of the `LandedRef` grep does
not sum to the total it reports. "returns 12 lines: 1 doc comment, 1 declaration, 7 production uses,
and 2 test lines" = **11**. The unaccounted twelfth is a substring false positive at
`internal/cli/todo_undone_test.go:266` — the function name `TestTodoDone_RequireLandedRefuses…`
contains `LandedRef` inside "Require**LandedRef**uses". The load-bearing figure (seven production
sites) is correct and I verified it; only the enumeration is wrong. Notable because this is the same
sentence iteration 1's D10 corrected, and it is presented as an exhaustive account of the grep's
output. — Severity: **minor** — Class: **blocking** — Required fix: add the twelfth line to the
breakdown as a name-collision false positive (it is worth naming: it is the reason a future reader
re-running the grep gets a count that seems one too high).

**D5.** `plan.md:75-83` (§C ruling 1) — the stdout token ruling is presented as a derivation from
`spec.md` §A.6 Table 2 ("the SPEC had in fact decided it"), but Table 2 decides only the **token**
(`landing=landed`); it does not decide the **placement**. The ruled form is `done <id>
landing=<verdict>` — a suffix on the existing line — and the only ground given for suffix-vs-second-
line is "One line per act; a second line would give an operator script two records for one event",
which is an unsourced design preference, not a derivation. `REQ-TLS-005` ("exactly one
landing-verdict token on stdout") is satisfied by either form, so nothing else constrains it. The
reasoning is sound; the framing overclaims its source. **Ruling 2 (exit code 0 on `unknown`) is by
contrast genuinely sourced** — `REQ-TLS-006` ("shall archive … shall not refuse"), `spec.md` §D's
exclusion of policy reversal, and `internal/cli/todo.go:20`'s "exit 0/1" two-code contract together
decide it; only the sub-claim "every unattended caller treats non-zero as refusal" is an unverified
universal, and it is not load-bearing. — Severity: **minor** — Class: **optional** — Required fix:
in §C ruling 1, separate the two decisions — cite Table 2 for the token spelling, and label the
suffix placement as a plan-phase design ruling with its stated reason, rather than as something the
SPEC had already decided.

**D6.** `spec.md:222` (§A.5 half-A/half-B table) — the half-A row describes the deliverable as
"the queue `state` becomes visible beside the link outcome", and Table 2's C2 row (`spec.md:283`)
renders the same idea as "`no-link` **+ `picked` is visible in the same row**". Both are correct,
but neither says which of the five existing columns the state joins or replaces, and `plan.md` M4
(`:144-151`) adds a sixth column without saying so explicitly. `todo pr`'s row is a tab-separated
machine surface consumed by scripts; a column count change is a contract change of the same kind as
R6's `unknown` kind, which the plan *does* flag as an external-consumer risk. — Severity: **minor**
— Class: **optional** — Required fix: state in `plan.md` M4 that the row goes from five columns to
six and where the new column sits, and add the column-count change to R6's "name it in the
sync-phase notes" mitigation.

---

## Regression Check (defects from iteration 1)

| # | Iteration-1 defect | Status | Evidence |
|---|---|---|---|
| D1 | REQ-TODO-013 freeze misattributed | **RESOLVED (by honest move)** | §B.2 is question-form, names both readings; the correction travelled into card `t359`'s text |
| D2 | Table 2 C4/C5 reversed REQ-1.10 | **RESOLVED** | `spec.md:305-310` now explicitly preserves REQ-1.10 and cites both enforcement sites |
| D3 | "the observed commit SHA" singular | **RESOLVED (moved)** | `spec.md:521-522` excludes it and names the many-to-one problem (t322 matches 24) |
| D4 | 26 requirements vs Tier M ceiling 16 | **RESOLVED** | 11 REQ / 10 AC against 16/16 (`spec-workflow.md` REQ/AC budget table), counted mechanically |
| D5 | AC that cannot fail; mutant satisfies it | **PARTIAL** | Rebuilt behaviourally and mutant-probed — real progress. But my mutant of a second shape survives it (D1 above), and the same class recurs at AC-TLS-009 (D3 above) |
| D6 | orphaned AC-TLS-016 | **RESOLVED** | `REQ-TLS-007` added; §E verified in both directions, zero orphans |
| D7 | four citations at wrong addresses | **RESOLVED** | Exhaustive sweep: every citation in all three artifacts re-opened; all resolve with the claimed content |
| D8 | ref figures pinned to a tree SHA | **RESOLVED** | Re-expressed as re-runnable commands with both ref SHAs + instant; I re-measured and the count had moved again (349→375), vindicating the fix |
| D9 | three `[NEEDS CLARIFICATION]` markers | **RESOLVED** | Zero markers (rc=1). Grounding of the two rulings judged in D5 above — one sound, one over-framed |
| D10 | "six production sites" vs seven | **RESOLVED** | Corrected to seven; I re-measured seven. New minor arithmetic issue in the same sentence — D4 above |
| D11 | one-column argument answered nothing | **RESOLVED (moved)** | Left with the B half |
| D12 | rows cell-identical once S3 retired | **PARTIAL** | Rows collapsed C4-C7, which addresses the symptom. The root — S3's retirement being unbacked — persists and is now D2 above, in a sharper form |
| D13 | missing `queued`+landed state | **RESOLVED** | C11 added to both tables (`spec.md:271`, `:289`) with a reachability argument |
| D14 | AC-TLS-006 cited a test that does not assert it | **RESOLVED** | The RED cell now states plainly that `TestTodoDone_NoLandingQueryWithoutTheFlag` "never inspects stdout"; verified in source — it discards stdout |
| D15 | R6 measurable, measures zero | **RESOLVED** | Measured; I independently reproduced all three figures (one file, 108 web files, zero) |
| D16 | doc comments not carried in M1 | **RESOLVED** | `plan.md:117-119` names all four; all four addresses verified |
| D17 | template-neutrality guard not carried | **RESOLVED** | `plan.md` §B.7; workflow file exists |
| D18 | `backlog_store.go:42-46` off by one | **RESOLVED** | Now `:62-63`, verified exact |

**No stagnation.** No defect survived all iterations unchanged; the two PARTIALs both moved
materially and their residue is stated in a sharper, more specific form than before.

---

## Score movement, justified

0.74 → **0.85** (+0.11). What earned each change:

- **Testability 0.55 → 0.70 (+0.15).** Iteration 1 had one criterion that could not fail at all and
  was defended by a name-based sweep a mutant walked through. That criterion is now behavioural,
  its RED was established by an actually-planted mutant, and eight of ten REDs are independently
  reproducible — I reproduced every one. The dimension does not reach 0.75 because two criteria
  still carry an unfailable clause (D1, D3), one of which I demonstrated live.
- **Traceability 0.70 → 0.95 (+0.25).** The orphan is closed by `REQ-TLS-007`, and the §E table now
  verifies in both directions against `acceptance.md` with zero orphans and zero uncovered
  requirements.
- **Completeness 0.80 → 0.85 (+0.05).** The requirement set now fits its declared tier honestly
  (11/16, 10/16) rather than by relabelling, and the Tier M artifact set is complete. Held below
  0.90 by D2.
- **Clarity 0.90 → 0.90 (0.00).** Iteration 1's clarity was already the document's strongest
  dimension and the remediation neither improved nor harmed it; the citation corrections were
  accuracy, not clarity. The one residual ambiguity (D2's S3 cell) predates this iteration.

Movement is monotonic: no dimension regressed.

---

## Recommendation

**PASS at 0.85**, above the Tier M threshold of 0.80, with all seven must-pass criteria satisfied
and no unresolved D7 or D8 BLOCKING finding.

This iteration is materially different in kind from iteration 1. There, the document was two SPECs
wearing one frontmatter, its central storage argument rested on a misattributed requirement, and
three clarification markers were open at audit time. Here the requirement set is coherent, the split
is honest and its destination real, every citation resolves, and the ref-dependent figures are
expressed so that a later reader is forced to re-measure them — which I did, and they had already
moved again.

The Tier M iteration ceiling is now exhausted (2/2). The three blocking defects therefore route as a
**scoped delta fix before Implementation Kickoff Approval**, not as a third audit round:

1. **D1 first** — widen AC-TLS-008 past the queue directory, or narrow `REQ-TLS-009` to match it,
   and re-establish the RED with a write placed **outside** `StateDirForRoot`. This is the guard on
   the SPEC's own most-plausible slip, and I have shown it porous.
2. **D2** — decide whether this SPEC retires hand ancestry. If yes, give it a requirement and a
   criterion; if no, rewrite Table 2's S3 column and re-argue the C4-C7 collapse on S2 alone.
3. **D3** — sharpen `REQ-TLS-011` to the delta the doctrine does not already carry, or fold it into
   `REQ-TLS-010`.
4. **D4** is a one-line correction and should ride along.
5. **D5 and D6** are optional; leave them to orchestrator discretion.

None of the three requires redesign — all are verification-layer and state-model edits inside
`acceptance.md` and `spec.md` §A.6/§C. Confirmation should be scoped to those five defects.

**On the scope split, explicitly: it was performed honestly.** The moved half went to a card that
exists, carries the moved defects forward as named preconditions, and is correctly declared
downstream of this one. §B.2 relocates the `REQ-TODO-013` question without asserting the
misattributed freeze — I checked the canonical text at `SPEC-KANBAN-TODO-CLI-001/spec.md:59` and
`SPEC-TODO-ANALYSIS-001/spec.md:51` and neither surviving passage takes the wrong side. Nothing in
the narrowed artifacts depends on the moved half. The one place the split left load-bearing
reasoning behind is Table 2's S3 column, and that is D2.

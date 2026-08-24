# SPEC Review Report: t210 (two SPECs)

Iteration: **2/2** (Tier M ceiling — final iteration)

> **Read this verdict as AUDIT-TIME, pinned to `22c3df39e`.** Both PASS scores,
> and the four surviving blocking findings N1-N4, describe the artifacts as they
> stood at that commit. **N1-N4 were repaired afterwards, outside this audit**,
> in `985343fad` — this report was NOT re-run against the repaired artifacts and
> its scores were not recomputed. The post-audit disposition of each finding is
> recorded in each SPEC's `progress.md` §E.1, which is the current-state surface;
> this file is the record of what was judged, not of what is now true. Nothing
> below is edited to match the repairs, because a verdict rewritten to agree with
> later work stops being evidence of what was actually observed.

| SPEC | Tier | Threshold | Score | Verdict |
|---|---|---|---|---|
| `SPEC-KANBAN-QUEUE-PR-SYNC-001` | M | 0.80 | **0.801** | **PASS (marginal)** |
| `SPEC-KANBAN-PR-CARD-TRACEABILITY-001` | S | 0.75 | **0.775** | **PASS (marginal)** |

Reasoning context ignored per M1 Context Isolation. The dispatch's framing of what
the author fixed was not treated as evidence; every disposition below rests on the
pinned artifact files plus commands re-run in this worktree. Where the dispatch
implied a defect (REQ-1.10 "a wish", AC-011 "vacuous"), I tested the implication
and, on both, **found against it** — reasons in §3.

**Both margins are thin — 0.001 and 0.025 — which is inside the resolution of my
own scoring. Do not read either PASS as comfortable.** The verdict rests on the
seven must-pass criteria (all pass on both SPECs) and on eighteen iteration-1
findings being genuinely closed (§2). What the margins do not excuse is four
blocking findings that survive (N1-N4). Each is a one-to-five-line edit; §8 says
exactly what remains.

---

## 1. Pin verification

Verified before the audit and again immediately before writing this verdict.

```
$ git status --short
                                   (empty — no tracked drift, both times)
$ git rev-parse HEAD
22c3df39e5d2c514111720f494fbe5ac72d0f833
$ git ls-tree -r HEAD .moai/specs/SPEC-KANBAN-QUEUE-PR-SYNC-001/ \
                      .moai/specs/SPEC-KANBAN-PR-CARD-TRACEABILITY-001/
06a2fc85f4a91e7962c437a01c37a17425ee215f  PR-CARD-TRACEABILITY-001/plan.md
b8874230cdfcb6472c28eceddcfc9c32e11f3ed6  PR-CARD-TRACEABILITY-001/progress.md
776bc491d19c2a6410f0d136053b93943c0d820f  PR-CARD-TRACEABILITY-001/spec.md
4a81edacc7fe0c5de9a25012d6e40d3beb2f52a6  QUEUE-PR-SYNC-001/acceptance.md
cea4c1075b9e98599ab2846064668e146ef5c8a6  QUEUE-PR-SYNC-001/plan.md
5c8acfd18ca4bf71c6ee1d49ee4c87f90f0d8a7f  QUEUE-PR-SYNC-001/progress.md
7ab347efd05e78fa44f8085c4ee51410caf18216  QUEUE-PR-SYNC-001/spec.md
```

All seven blobs match the dispatch's expected set, both times. The
content-addressed freeze held; iteration 1's failure mode did not recur.

---

## 2. Iteration-1 finding disposition (D1-D18)

Each verified independently. "The author says it is fixed" was not accepted as
evidence for any row.

| # | Disposition | Evidence |
|---|---|---|
| **D1** frontmatter `tags` | **closed** | `spec.md`:L13 `tags: "kanban, todo, github, observability, read-only"` (quoted string). `moai spec lint <spec.md> --json` → `[]` EXIT=0 on **both** SPECs — the file now parses, so the empty array is meaningful rather than suppressed. |
| **D2** scope-motivation | **closed**, via option (c) | §A:L39-58 now records both failure shapes and the one-lane cost, and states why a merged-PR extension does not reach t199. §C.2 + REQ-1.9 adopt the landed carrier. §F:L311-320 narrows the exclusion to *merged-PR attribution* and says so. This is the fix that actually closes the incident, not the cheap one. |
| **D3** veto vs re-decide | **closed** | Doctrine `spec.md`:L62-73 carries the prescribed wording verbatim — *"until the operator, so informed, confirms or withdraws it"* — plus the reasoning against line 29, plus the explicit note that no mechanical check can tell the readings apart. |
| **D4** refuted fixtures | **closed**, re-verified live | I re-ran the fixture command today. `t188` → body of #1601 only, no title. `t201` → bodies of #1611 and #1612, no title. `t200` → title of #1612. All three assignments in AC-001/002/003 hold against current data. |
| **D5** `--pr` MAY | **closed** | REQ-2.5:L272-273 *"No `--pr` flag on `moai todo list` is in scope"*; the option moved to §H (non-normative); AC-009:L175-177 records the whole-claim consequence. `grep -icE '\bmay\b\|\bshould\b'` over the 16 requirement lines → **0**. |
| **D6** precedent omission | **closed** | §B.1:L81-91 cites the third [HARD] clause at line 31 verbatim and names it as REQ-2.1's wording source. |
| **D7** AC-004 coverage | **closed** | AC-004 is now a recursive digest over `.moai/state/kanban/` with a path-set assertion (no sidecar, no lock), extended over the fail-open, ambiguous, and landed paths. |
| **D8** REQ numbering | **open** (optional), now **aggravated** | Still `REQ-1.1`…`REQ-2.6`. The sibling authored in the same pass uses canonical `REQ-001`…`REQ-004`, so one card now ships two incompatible id schemes. `moai spec lint` tolerates both. See N9. |
| **D9** `Where`→`While` | **closed** | REQ-1.2/1.3/1.4 all open with `**While**`. |
| **D10** non-normative trailers | **closed** | REQ-2.1 *"and shall write no field, no `findings[]` entry, and no timestamp"*; REQ-2.4 *"and `moai todo list` shall remain lock-free, network-free…"* — both promoted to `shall`, neither dropped. |
| **D11** AC-005 disjunction | **closed** | AC-005:L113-114 — *"renders **empty** for every card (not omitted — the column is present and blank)"*. |
| **D12** grep tail untested | **closed in tooling; partially closed in doctrine** | AC-013 is split into a mechanical half and a labelled reviewer-judgement half. In the doctrine SPEC AC-001 and AC-002 are correctly split, **but AC-003's `**Mechanical.**` block still claims the four-carrier check with no command covering it** → N4. |
| **D13** AC-006 count | **closed** | The "15 tokens" figure is gone; AC-006:L138-142 states only the load-bearing property. |
| **D14** tier budget | **closed structurally, zero headroom** | 19 → 16 (measured) + a Tier S sibling at 4/8. But the split lost a requirement (N3) and the tooling SPEC sits at exactly 16/16, so the missing template-mirror requirement (N1) **cannot be added without breaching the ceiling**. The pressure D14 named is absorbed exactly, not relieved. |
| **D15** mirror split | **closed** | AC-013 covers both `todo.md` copies; doctrine AC-004 covers both `kanban-dispatch.md` copies. Division is correct — tooling keeps `todo.md`, doctrine takes `kanban-dispatch.md`. |
| **D16** REQ-1.5 over-general | **closed — genuinely, not relabelled** | See §3.1. |
| **D17** `-E` boundary | **closed** | REQ-1.9 mandates `--perl-regexp` and names `-E` as forbidden with the reason. AC-011 carries three controls. `plan.md` §B:L32-35 and §G:L129-130 both name it. |
| **D18** REQ-1.8 confirmed | **carried** | AC-008 unchanged and still sound. |

**Regressions: none.** No previously-closed finding reopened.

---

## 3. The new material, attacked

### 3.1 Does the Q1/Q2 split dissolve D16, or relabel it?

**It dissolves it.** Three independent pieces of evidence, not one:

1. REQ-1.5's prohibition is scoped in the requirement text itself — *"when
   attributing an open pull request to a delivering card (question Q1)"* — not in
   a footnote that an implementer can skip.
2. REQ-1.9 **affirmatively mandates** the commit carrier for Q2. A relabelling
   would narrow the ban and leave Q2 unspecified; this creates the counterpart
   obligation.
3. `plan.md` §G:L133-134 names *"Extending REQ-1.5's commit-message ban to the
   landed check"* as an anti-pattern by name. The run phase is warned in the
   direction the defect actually pointed.

**Does any requirement silently apply Q1's reasoning to Q2?** I checked all 16.
It does not. The one place the two touch is REQ-1.9's guard (*"While no open pull
request carries the card id token in its title or body"*), which is a
**sequencing** rule, not Q1 reasoning leaking. It has its own consequence — N7.

### 3.2 Is REQ-1.10's boolean-only constraint enforceable, or a wish?

**Enforceable, by construction. I find against the dispatch's framing here.**

The prohibition is not a behavioural request that an implementation might quietly
ignore — it is a constraint on the **shape of the returned type**, and AC-012
asserts exactly that: *"the returned record contains no commit SHA, no commit
subject, and no field naming a delivering commit."* A record type with no such
field cannot report a delivering commit; there is nothing to report it with.
`plan.md` M2 reinforces it at the implementation layer (*"returning a boolean and
nothing else"*).

Compare a genuinely unenforceable prohibition — "the resolver shall not mislead
the operator" — which no test can observe. This one is observable by inspecting
the type. The dispatch's worry ("an unenforceable prohibition reads as a
guarantee") is the right worry applied to the wrong requirement.

One real weakness survives, and it is smaller: AC-012's second clause — *"the
public resolver API exposes no accessor that would return one"* — has no stated
verb, and unlike AC-013 and the doctrine SPEC's ACs it is **not labelled a
judgement**. It is mechanizable in Go (reflect over the exported fields), but the
SPEC neither says so nor admits it is a judgement. → **N8**.

### 3.3 Can AC-011's `-E` tripwire pass vacuously?

**No.** I traced the logic rather than the intent, and the trap the dispatch
describes is closed by the positive control, not by the tripwire.

The criterion has three parts. Part A asserts the query string carries
`--perl-regexp`. Part B pins the control's own assumption (`-E` returns empty for
`t199`) — so if a future git made `\b` work under `-E`, the assumption breaks
*loudly* rather than silently. Part C fails the test when the implementation's
result matches the `-E` result on the positive control.

Three broken implementations, each caught:

| Broken implementation | What catches it |
|---|---|
| uses `-E` | result is empty; matches the `-E` result on the positive control → Part C fails. Also: the positive control itself requires `t199` → `landed` with a **non-empty** commit set. |
| returns `landed` unconditionally | the negative control (`t205` → `no-link`, empty set) fails. |
| returns `no-link` unconditionally | the positive control fails. |

The dispatch's concern — "an assertion comparing against an empty result may
itself be satisfiable by a broken implementation" — would hold if Part C were the
*only* assertion. It is not: the positive control independently demands a
non-empty result, so the comparison is a second line of defence rather than the
first. I re-ran the baseline and it reproduces:

```
$ git log origin/main --perl-regexp --grep='\bt199\b' --oneline | wc -l   → 3
$ git log origin/main -E           --grep='\bt199\b' --oneline | wc -l   → 0
$ git log origin/main --perl-regexp --grep='\bt205\b' --oneline | wc -l   → 0
```

AC-011 is the best-constructed criterion in either SPEC. One ambiguity attaches
to it → **N6**.

### 3.4 The Tier S SPEC's four inline ACs

**The grep-shaped halves do real work.** I confirmed all three zero baselines
myself today, in this tree:

```
$ grep -c 'pre-dispatch PR cross-check'                .../kanban-dispatch.md   → 0
$ grep -c 'confirms or withdraws'                      .../kanban-dispatch.md   → 0
$ grep -c 'PR title MUST carry the delivering card id' .../kanban-dispatch.md   → 0
```

Zero on the pre-implementation tree is the project's admissibility bar, and all
three clear it. These are not the "grep a word already in the prose" pattern.
AC-001's second grep (`\[HARD\].*pre-dispatch PR cross-check`) is a genuine
addition — it makes the marker observable rather than asserted — and `plan.md`
§B:L27-29 warns that it is order-sensitive, which is the kind of self-aware
detail that keeps a correct clause from failing its own criterion.

**Are the judgement halves honestly labelled?** Two of three, yes — AC-001 and
AC-002 both carry *"Reviewer judgement (recorded as such)"* and AC-001 adds *"No
command verifies this, and this SPEC does not claim one does."* That is exactly
what D12 asked for.

**AC-003 is the exception and it is a D12-shape relapse.** Its `**Mechanical.**`
block reads *"After M1 that grep returns at least 1, **and the same section names
all four traceability carriers**"* — and the code block underneath carries one
grep, which covers the first clause only. The four-carrier claim is a judgement
sitting under a Mechanical heading. It is *also* stated correctly in the judgement
half below, so no honesty is lost in substance — but the label is wrong, in the
one SPEC that exists because mislabelled judgements were a blocking finding.
→ **N4**.

### 3.5 The 16/16 count and the REQ-2.6 consolidation

**The consolidation is genuine, not a staple.** I expected to find otherwise.

REQ-2.6 joins two clauses with a semicolon — render outcome+confidence in both
forms; the JSON form carries these five fields. The test for a staple is whether
the second clause is an independent obligation or an elaboration of the first.
**The document's own precedent settles it**: REQ-1.1 already enumerates five
fields of a returned record inside one requirement, and nobody counts that as five
requirements. REQ-2.6 does the same thing for the render surface. Applying a
stricter standard to REQ-2.6 than to REQ-1.1 would be inventing a rule to
manufacture a finding.

What I will not soften: **the budget fits exactly, with zero headroom, and that
is load-bearing.** The tooling SPEC is at 16/16 requirements. N1 identifies a
missing normative requirement (the template mirror). It cannot be added without
tiering up or splitting again. So D14's pressure was absorbed to the last slot
rather than relieved, and the run phase inherits a document that cannot take one
more requirement.

**Tier S is genuinely Tier S.** Two files changed, zero Go code, 4 requirements
against a ceiling of 8, 4 criteria against 8, artifact set spec.md + plan.md with
ACs inline — matches the S row of `spec-workflow.md` § SPEC Complexity Tier on
every axis.

### 3.6 Did the split drop anything?

**Yes — one requirement, and it left an orphaned criterion behind.** Mapping the
old REQ-3 block (blob `313a5c31a`) against the new SPEC:

| old | new | status |
|---|---|---|
| REQ-3.1 (read + report before dispatch) | REQ-001 **+** REQ-002 | preserved and split, improved per D3 |
| REQ-3.2 (PR title carries card id) | REQ-003 | preserved, scoped to card-delivering PRs |
| REQ-3.3 (non-contradiction note) | REQ-004 | preserved, extended to name four carriers |
| **REQ-3.4 (template mirror obligation)** | **— none —** | **dropped from the requirement layer** |

The obligation survives as doctrine AC-004, `plan.md` §D, and milestone M3 — so
nothing is unimplemented. But AC-004 now references no requirement, and the Tier S
SPEC had **four unused requirement slots**, so no budget pressure explains the
drop. → **N3**.

Everything else divides correctly: the `todo.md` mirror stayed with the tooling
SPEC (AC-013), the `kanban-dispatch.md` mirror went with the doctrine (AC-004),
and cross-references resolve in both directions (4 refs each way in the spec
bodies; both siblings `status: draft`, neither retired or superseded).

### 3.7 The self-declared gaps

**Pinned fixtures vs. a PR set that has grown to 12.** I judge the pinning
**sound engineering, not rationalized staleness**, and the anti-pattern label is
correct. The reason is that v0.1.1's fixtures were refuted by *misdescription*,
not by drift: AC-002 asserted `inferred` for a token that already appeared in two
bodies at the time it was written. Pinning does not reproduce that failure — it
prevents a *different* one (non-determinism). And I confirmed the current
fixtures survive the drift anyway: `t188`, `t201`, `t200` all still resolve as the
ACs assert, against today's 12-PR set. What *will* drift is §C.1's carrier
scorecard (7/11, 11/11), and that is motivation rather than test input.

**Merged-PR attribution precision unscored.** Disclosed in §F and §H, with the
narrowing argued rather than asserted. Honest.

**Reviewer-judgement halves with no command.** Disclosed in `progress.md` §Gaps
and labelled in the criteria themselves — except AC-003 (N4) and AC-012 (N8).

---

## 4. Must-Pass Results

### `SPEC-KANBAN-QUEUE-PR-SYNC-001`

- **[PASS] MP-1** — 16 leaf requirements, `REQ-1.1`…`REQ-1.10`, `REQ-2.1`…`REQ-2.6`, extracted in document order: sequential, no gaps, no duplicates. The hierarchical scheme is non-canonical (D8/N9) but MP-1's substance — sequence and uniqueness — holds.
- **[PASS] MP-2 (requirement layer)** — all 16 checked individually. Ubiquitous: REQ-1.1, 1.7, 1.8, 2.2, 2.4, 2.5, 2.6. State-driven `While`: REQ-1.2, 1.3, 1.4, 1.9. Event-driven `When`: REQ-2.3. Unwanted `shall not`: REQ-1.5, 1.6, 1.10. Compound: REQ-2.1. `grep -icE '\bmay\b|\bshould\b'` over the requirement lines → **0**. Judged against `spec.md` REQ entries only; `acceptance.md`'s Given-When-Then is the verification layer and was graded under §5 Testability, not here.
- **[PASS] MP-3** — all 12 canonical fields present with correct types (`version`/`title`/`phase`/`module`/`tags` quoted strings; `created`/`updated` ISO dates; `priority: P1`; `lifecycle: spec-anchored`), plus `tier: M`. No rejected snake_case alias. `moai spec lint --json` → `[]` EXIT=0.
- **[N/A] MP-4** — single-language SPEC (Go tooling in this repo's own CLI). Auto-passes.
- **[PASS] MP-5 (D7)** — the only SPEC references are the two siblings; both exist, both `status: draft`. No retired/superseded reference. No BLOCKING finding.
- **[PASS] MP-6 (D8)** — `grep -c syscall` → 0 on spec.md and plan.md. Auto-pass.
- **[PASS] MP-7** — `grep -rn '\[NEEDS CLARIFICATION'` over both SPEC directories → no match (exit 1).

### `SPEC-KANBAN-PR-CARD-TRACEABILITY-001`

- **[PASS] MP-1** — `REQ-001`…`REQ-004`, canonical flat zero-padded form, sequential, no duplicates.
- **[PASS] MP-2** — REQ-001 event-driven `When`; REQ-002 state-driven `While`; REQ-003 ubiquitous with a non-system subject (GEARS permits any noun); REQ-004 ubiquitous.
- **[PASS] MP-3** — 12 fields, correct types, `tier: S`. `moai spec lint --json` → `[]` EXIT=0.
- **[N/A] MP-4** — doctrine SPEC, no language-specific tooling.
- **[PASS] MP-5 / MP-6 / MP-7** — as above; no syscall, no clarification marker, sibling reference resolves and is `draft`.

---

## 5. Category Scores

### `SPEC-KANBAN-QUEUE-PR-SYNC-001` — harmonic mean **0.801** (threshold 0.80)

| Dimension | Score | Band | Evidence |
|---|---|---|---|
| Clarity | 0.85 | 0.75 (upper) | §C's two-question table (L119-122) and §C.2's measured both-directions block make the hardest design call legible. One genuine ambiguity survives (N6). N7 is a scope gap, not an ambiguity — REQ-1.9's guard is explicit. |
| Completeness | 0.85 | 0.75-1.0 | Every required section present and substantive; five `### Out of Scope —` H3 sub-headings each with specific bullets (L296-339); frontmatter complete. Deductions: the template-mirror obligation has no normative requirement (N1/N3), and all four NFRs are AC-uncovered (`grep -n 'NFR-' acceptance.md` → **no match**) (N2). |
| Testability | 0.80 | 0.75 (upper) | AC-004 is a recursive directory digest, not a "no error" check. AC-011 carries three controls. AC-012 is a structural type assertion. Weasel-word scan (`appropriate\|adequate\|reasonable\|proper\|as needed\|if necessary\|where possible`) → only the substring "property" in prose; **zero real hits**. Deductions: N2 (NFR-1's one-call bound — the sole basis for REQ-2.5 — is observed by nothing), N8. |
| Traceability | 0.72 | 0.75 (lower) | D.2 covers all 16 requirements — the primary property is fully intact. But AC-013 maps to *"mirror parity (plan.md M4)"*, not to any REQ, and is **absent from the D.2 table** (`awk '/D.2 Traceability/,/D.3/' \| grep -c 'AC-013'` → **0**) while L295-296 asserts *"every criterion exercises at least one requirement"* — a false claim about the document's own contents (N1). Scored below the 0.75 anchor for the false claim, well above 0.50 since 1 of 13 criteria is affected. |

`4 / (1/0.85 + 1/0.85 + 1/0.80 + 1/0.72) = 4 / 4.9918 = 0.8013`

### `SPEC-KANBAN-PR-CARD-TRACEABILITY-001` — harmonic mean **0.775** (threshold 0.75)

| Dimension | Score | Band | Evidence |
|---|---|---|---|
| Clarity | 0.90 | 1.0 (lower) | All four requirements single-interpretation. REQ-002's L66-73 note is the strongest passage in either document: it names the failure mode, argues against line 29, and states outright that *"No mechanical check can tell the two readings apart after the fact, so the wording is the only control."* |
| Completeness | 0.85 | 0.75-1.0 | HISTORY, Context, Requirements, rationale, inline ACs, four `### Out of Scope —` H3s with bullets, cross-references — all present. Deductions: the mirror requirement dropped in the split (N3); no traceability matrix. |
| Testability | 0.75 | 0.75 | Three grep criteria with independently-confirmed zero baselines (all three re-measured at 0 by me today). Deductions: AC-004 — the parity criterion — states *"every clause added by AC-001, AC-002, and AC-003 is present in the mirror"* with **no command**, in a SPEC whose own `plan.md` §B:L30-34 warns the mirror is deliberately not a verbatim copy. The three literal phrases contain no SPEC ID, path, date, or SHA, so they are neutrality-safe and could simply be grepped against the mirror; the criterion does not say so. Plus N4. |
| Traceability | 0.65 | 0.50-0.75 | REQ-001→AC-001, REQ-002→AC-002, REQ-003+REQ-004→AC-003 all hold in substance. But there is no traceability matrix; AC-001 cites no REQ id; and **AC-004 references no requirement at all** (N3). Two of four criteria have no explicit REQ citation. |

`4 / (1/0.90 + 1/0.85 + 1/0.75 + 1/0.65) = 4 / 5.1594 = 0.7753`

---

## 6. Defects Found

**N1. AC-ORPHAN-FALSE-CLAIM — `acceptance.md`:L43 and L295-296 — AC-013 references no requirement, and the file asserts the opposite about itself — Severity: major — Class: blocking — Required fix:** the AC matrix row reads `| AC-013 | mirror parity (plan.md M4) | …`, and AC-013 is absent from the D.2 table (verified: 0 occurrences); yet L295-296 states *"every criterion exercises at least one requirement."* That is an unobserved claim about the document's own contents. Two routes, and the choice is not free: (a) add a normative template-mirror requirement — **but the SPEC is at exactly 16/16, so this breaches the Tier M ceiling and forces a tier-up**; or (b) restate L295-296 honestly, add an AC-013 row to D.2 mapping it to the Template-First constraint in `plan.md` §D, and record that the mirror obligation is carried by a plan constraint rather than a requirement. (b) is the cheap route and is what I recommend for run-phase entry; (a) is the structurally correct one and belongs in a follow-up.

**N2. NFR-UNCOVERED — `spec.md`:L283-292 (§E) and `acceptance.md` (all) — none of the four NFRs has any acceptance criterion; NFR-1 is the sole justification for the SPEC's headline design ruling — Severity: major — Class: blocking — Required fix:** `grep -n 'NFR-' acceptance.md` returns **no match**. NFR-1 bounds the verb at *one* `gh pr list` with no per-card call, and REQ-2.5's dedicated-verb ruling rests entirely on the 0.878s figure that bound protects. An implementation issuing one `gh` query per card satisfies every one of the 13 criteria while costing ~0.9s × queue length. The fix is nearly free and reuses machinery AC-009 already requires: AC-009 injects a command executor that *records every subprocess invocation* — extend it (or add AC-014) to assert that `moai todo pr` spawns exactly one `gh` process regardless of queue length. NFR-2 (landed check does no network) rides the same assertion. NFR-3 and NFR-4 need no criterion.

**N3. SPLIT-DROPPED-REQUIREMENT — `SPEC-KANBAN-PR-CARD-TRACEABILITY-001/spec.md` §B and §D AC-004 — old REQ-3.4 (the template-mirror obligation) did not survive the split into the requirement layer, leaving AC-004 orphaned — Severity: major — Class: blocking — Required fix:** verified against the iteration-1 blob `313a5c31a`:L211-213, which carried *"REQ-3.4 — The doctrine change shall be mirrored into `internal/template/templates/…/kanban-dispatch.md` per the Template-First rule, subject to the template neutrality catalogue."* No successor exists in §B. Unlike N1 there is **no budget obstacle** — the Tier S SPEC is at 4 requirements against a ceiling of 8. Restore it as REQ-005 and map AC-004 to it. Two lines.

**N4. MECHANICAL-OVERCLAIM — `SPEC-KANBAN-PR-CARD-TRACEABILITY-001/spec.md`:L175-179 (AC-003) — a judgement sits inside the `**Mechanical.**` block, in the SPEC that exists partly because mislabelled judgements were a blocking finding — Severity: minor — Class: blocking — Required fix:** the block reads *"After M1 that grep returns at least 1, **and the same section names all four traceability carriers**"* and then shows a single grep covering only the first clause. The four-carrier claim is already stated correctly in the reviewer-judgement half below, so no substance is lost — only the label is wrong. Delete the eight words from the Mechanical block, or add a grep that observes the four carriers (e.g. a count over the carrier names). One line either way. Flagged blocking because it is a direct relapse of D12 inside D12's own remedy.

**N5. FIXTURE-BLOCK-ELIDED — `acceptance.md`:L10-21 — a command's output is presented as a live measurement but is an unmarked excerpt — Severity: minor — Class: optional — Required fix:** the header says the fixtures *"were re-measured live in this worktree with"* the stated `gh … -q '.[] | …'` command. That jq emits **one line per open PR unconditionally** — I re-ran it today and got 12 lines for a 12-PR set. §C.1 states the measurement covered **11** open PRs, so the block should carry 11 lines; it carries 5. At least four token-bearing PRs from that set are absent (#1605 `t194`, #1606 `t187`, #1613 `t202`, #1617 `t205` — all present today with tokens). **No AC fixture is refuted** — I re-verified t188, t201, and t200 against current data and all three hold — and the 5-PR set is sufficient for every criterion, so this is a labelling defect, not a correctness one. Under this project's verbatim-output rule it is still a defect: add the omitted lines, or mark the block an excerpt and state the full set size.

**N6. AC-011/AC-012-OBSERVABILITY-TENSION — `acceptance.md`:L201-207 vs L230-232 — two criteria impose opposite observability demands on the same value, and neither says how to satisfy both — Severity: minor — Class: optional — Required fix:** AC-011 requires that *"the underlying query returns a non-empty commit set"* be observed; AC-012 requires that *"the public resolver API exposes no accessor that would return one."* Both are correct and both are satisfiable — the test observes the git-query helper one layer below the resolver's public surface — but the SPEC never says that, and `plan.md` M2's *"returning a boolean and nothing else"* reads as closing the only door AC-011 needs. Add one sentence to AC-011 naming the layer its control observes. An implementer who resolves this the other way will either weaken AC-011 to a boolean check (losing the anti-vacuous property) or add the accessor AC-012 forbids.

**N7. LANDED-MASKED-BY-OPEN-PR — `spec.md`:L227-234 (REQ-1.9) with L219-222 (REQ-1.7), and §F — a card that both carries an open PR and is already landed reports `linked` only, and nothing discloses it — Severity: minor — Class: optional — Required fix:** REQ-1.9 runs the landed query only *"While no open pull request carries the card id token in its title or body"*, and REQ-1.7 requires exactly one of four kinds. So an operator checking a card with a stale open PR whose work already merged via a batch is told `linked` and never told it landed — a near-neighbour of the t199 shape the SPEC exists to catch. This may well be the right design (one question, one answer), but it is a scope boundary and §F does not name it. Either add a bullet to §F stating that the landed check is not run when an open PR is found, or state in REQ-1.7 that the kinds are priority-ordered and why.

**N8. AC-012-UNLABELLED-JUDGEMENT — `acceptance.md`:L232 — the API-surface half carries no verb and, unlike its siblings, is not recorded as a judgement — Severity: minor — Class: optional — Required fix:** *"the public resolver API exposes no accessor that would return one"* is mechanizable (reflection over the exported record's fields) but the criterion neither names that mechanism nor labels the clause a reviewer judgement. AC-013 and the doctrine SPEC's AC-001/AC-002 both handle this correctly; AC-012 should match one of those two patterns.

**N9. TWO-ID-SCHEMES — `SPEC-KANBAN-QUEUE-PR-SYNC-001/spec.md` §D vs `SPEC-KANBAN-PR-CARD-TRACEABILITY-001/spec.md` §B — one card ships two incompatible requirement-id schemes — Severity: minor — Class: optional — Required fix:** iteration-1 D8 flagged `REQ-1.1`-style ids as cosmetic when only one SPEC existed. The sibling authored in the same pass uses canonical `REQ-001`. Grep-based tooling keying on `REQ-[0-9]{3}` matches one and not the other, and a reader moving between siblings meets two conventions. `moai spec lint` returns `[]` on both, so nothing mechanical breaks. Renumbering the tooling SPEC touches 16 requirement headings, 16 traceability rows, and every `plan.md` / `acceptance.md` back-reference — not free, and deferrable.

---

## 7. Regression Check (iteration 1 → 2)

Score moved **0.60 → 0.801** on the tooling SPEC. No regression, so the LEAN
STOP-on-regression clause does not fire. Of the eighteen iteration-1 findings,
sixteen are closed, one (D12) is closed in the tooling SPEC and partially closed
in the doctrine SPEC, and one (D8) stays open as an optional cosmetic item. No
closed finding reopened. **No stagnation:** no defect appears unchanged across
both iterations except D8, which was explicitly graded optional at iteration 1 and
is graded optional again — that is a deliberate deferral, not an absence of
progress.

---

## 8. Recommendation

**Both SPECs are fit to enter run-phase, and I would not describe either margin as
comfortable.**

> Audit-time, per the header note: the four items listed below as remaining were
> repaired in `985343fad` after this verdict was written. The list is preserved
> as issued rather than struck through — what it records is what the audit found.

The eighteen iteration-1 findings were addressed substantively rather than
cosmetically. Three repairs are better than the fixes I prescribed: the D2/D16
landed-carrier adoption closes the incident where my own first-pass suggestion was
refuted; the D3 wording carries the control clause *and* the argument for why
wording is the only control available; and AC-011's three-way control set is more
than the positive control D17 asked for. The Q1/Q2 split is a real design
distinction, argued from measurement in both directions, not a relabelling.

What remains, in the order I would fix it:

1. **N4** — delete eight words from doctrine AC-003's Mechanical block. One line.
2. **N3** — restore the template-mirror requirement as REQ-005 in the Tier S SPEC and map AC-004 to it. Two lines, four spare slots available.
3. **N2** — extend AC-009's recording executor to assert exactly one `gh` spawn on `moai todo pr`. One assertion in a harness the SPEC already requires. **This is the one whose absence can send the run phase wrong**: without it, a per-card-query implementation passes every criterion while defeating the design ruling those criteria exist to protect.
4. **N1** — restate `acceptance.md`:L295-296 honestly and add AC-013 to the D.2 table. Two lines. The structurally correct fix (a normative mirror requirement) is blocked by the 16/16 ceiling and belongs in a follow-up, not here.

All four are cheap; none requires re-measurement, re-argument, or a design change.
N5-N9 are optional and can ride into run-phase as recorded debt.

**The one thing I would not let the run phase inherit silently is the zero budget
headroom.** The tooling SPEC sits at exactly 16 requirements against a ceiling of
16, and N1's structurally-correct fix cannot be made without breaching it. D14's
pressure was absorbed to the last slot rather than relieved. If the run phase
discovers it needs one more requirement — and N2 and N7 are both candidates for
one — the answer is a tier-up or a further split, not a relaxed budget.

Iteration ceiling reached (Tier M = 2). No iteration 3 is available under this
contract.

# SPEC Review Report: SPEC-AGENTS-MD-CANON-001

Iteration: 1/3
Verdict: **FAIL**
Overall Score: **0.69** (harmonic mean; Tier L PASS threshold 0.85)

Audited tree: `.claude/worktrees/t82`, frozen at commit `ac6b209b8`
("feat(spec-agents-md-canon-001): rebuild the SPEC on the measured probe"). Working tree clean
apart from this report. All six artifacts re-read from the frozen state; every finding below is
cited against it.

Reasoning context ignored per M1 Context Isolation. Every figure was independently re-measured in
this worktree; no number is carried from the SPEC's own account, and no framing is carried from the
dispatch — the two findings handed to me were re-derived from the artifacts before being accepted,
and one of them is recorded here as stronger than it was handed over.

**Read-window note.** An earlier read of this SPEC caught the tree mid-rewrite (v0.1.0 artifacts at
02:08-02:09). Those findings are discarded, not carried forward. Line counts and heading line
numbers in the frozen set match the v0.2.0 content read at 02:12-02:14 exactly (spec 352, plan 147,
acceptance 117, design 186, research 56, progress 44), and the final `#### → ###` pass on the Out of
Scope headings left line numbers unchanged — so the citations below resolve against `ac6b209b8`.

**Score-regression note.** This verdict scores lower than a preliminary read of the same SPEC (0.76).
That is not a SPEC regression and does not trigger the STOP clause: it is iteration 1, the SPEC has
never been revised in response to a verdict, and the movement is entirely mine — a defect found late
(D2) on a broader evidence base. Recorded so the difference is not misread as deterioration.

---

## Must-Pass Results

- **[PASS] MP-1 REQ number consistency** — `REQ-AMC-001` … `REQ-AMC-017`, sequential, no gaps, no
  duplicates, uniform 3-digit padding (17 matches of `^\*\*REQ-AMC`). AC side `AC-AMC-001`…`021`,
  contiguous, 21 matches. Both under the Tier L ceiling of 25.
- **[PASS] MP-2 GEARS format compliance** — judged against the **requirement layer** (`spec.md` §C
  `REQ-AMC-*`), never against an AC. All 17 match a GEARS pattern: Ubiquitous (001, 004, 008, 009,
  011, 013, 015, 017), Unwanted (002, 005, 010, 016), Event-driven (003, 007, 012, 014), Where (006).
  `acceptance.md`'s Given-When-Then entries are the correct verification-layer form and are graded
  under Group 4. Two label/form slips recorded as optional findings (D10, D11).
- **[PASS] MP-3 YAML frontmatter validity** — all 12 canonical fields present with correct types
  (`spec.md:1-15`), plus optional `tier: L`. No rejected snake_case alias. `phase: "v3.2.0 target"`
  is a release label, not a lifecycle stage. Mechanically confirmed:
  `moai spec lint .moai/specs/SPEC-AGENTS-MD-CANON-001/spec.md` → `✓ No findings`, exit 0. The `id`
  matches the enforcing regex `^SPEC(-[A-Z][A-Z0-9]*)+-\d{3}$` (`internal/spec/lint.go:715`); the
  schema doc's single-segment regex is the stale surface, not the SPEC.
- **[N/A] MP-4 language neutrality** — the SPEC governs harness rule files and names no
  language-specific tooling. Auto-passes.
- **[PASS] MP-5 D7 cross-SPEC reconciliation** — one external reference,
  `SPEC-ALWAYS-LOADED-DIET-001`; that directory exists with `status: completed`. Not in
  {retired, superseded, archived}. No BLOCKING finding.
- **[PASS] MP-6 D8 cross-platform discipline** — `grep -rn 'syscall'` over the SPEC directory → no
  match. D8-4 auto-PASS.
- **[PASS] MP-7 clarification gate** — `grep -rn '\[NEEDS CLARIFICATION' plan.md research.md` → no
  match.

No must-pass failure. The FAIL is the aggregate falling below the Tier L threshold, concentrated in
Testability.

---

## Category Scores

| Dimension | Score | Rubric Band | Evidence |
|-----------|-------|-------------|----------|
| Clarity | 0.75 | 0.75 | "The integration branch" (`spec.md:186`, `plan.md:110`, `design.md:162`, `acceptance.md:75`) is load-bearing and defined nowhere (D3). `AC-AMC-009`/`AC-AMC-010` do not say which copy of `AGENTS.md` they bind once the M6 mirror exists (D2). |
| Completeness | 0.80 | 0.75 | All sections present; five `### Out of Scope — …` sub-headings (`spec.md:309/315/320/326/331`), each with `-` bullets, now at the canonical h3 level. Tier L 5-artifact set complete. Docked for the undisclosed measurement proxy (D5), a design rationale contradicted by measurement (D6), and the unaddressed nested-`CLAUDE.md` asymmetry (D7). |
| Testability | 0.55 | 0.50 | Five criteria cannot decide what they claim to: `AC-AMC-017` and `AC-AMC-001` require output a passing run does not emit (D1), `AC-AMC-010` fails on a correct tree (D2), `AC-AMC-002` names an unreproducible fixture (D4), `AC-AMC-012` names no mechanism (D9). `AC-AMC-007` remains a genuine negative-path criterion. |
| Traceability | 0.70 | 0.75→0.50 | `REQ-AMC-006` has no AC at all (D8); `REQ-AMC-005`'s sole criterion (`AC-AMC-010`) is broken, so a second requirement is effectively uncovered. No orphaned AC. |

Harmonic mean of (0.75, 0.80, 0.55, 0.70) = **0.6861**. Tier L threshold 0.85 → FAIL.

---

## Correction to my preliminary read

My earlier read listed `AC-AMC-010` under what the SPEC gets right — "`REQ-AMC-005` forbids nested
documents, `AC-AMC-010` enforces it mechanically". **That was wrong.** The criterion is falsified by
the SPEC's own M6 milestone, and is nondeterministic besides. It is now D2, the second critical
finding. The dispatch flagged the contradiction; re-derivation shows it holds and has a second
failure mode the dispatch did not name.

---

## What the SPEC gets right

Every numeric claim in the artifact set reproduces exactly in this worktree:

| Claim | Cited | Re-measured | Match |
|---|---:|---:|---|
| always-loaded rules, 14 files | 202,621 B | 202,621 B | yes |
| `CLAUDE.md` | 20,523 B | 20,523 B | yes |
| output style | 61,706 B | 61,706 B | yes |
| `[HARD]` lines, rules | 30,353 B / 93 | 30,353 B / 93 | yes |
| `[HARD]` lines, `CLAUDE.md` | 2,190 B / 4 | 2,190 B / 4 | yes |
| contract subtotal | 32,543 B / 97 | 32,543 B / 97 | yes |
| output-style `[HARD]` | 11,898 B / 75 | 11,898 B | yes |
| imperative union | 43,638 B | 40,501 + 3,137 = 43,638 B | yes |
| Claude-only upper bound, 6 files | 14,360 B / 38 | 14,360 B / 38 | yes |
| ceiling derivation | 24,576 = 75 % of 32,768; reserve 8,192; optimistic 32,543 − 14,360 = 18,183; pessimistic trim 24.5 % | 24,576/32,768 = 0.75; 32,768 − 24,576 = 8,192; 18,183; (32,543−24,576)/32,543 = 24.48 % | yes |
| output-style §8 share | 46,765 B, lines 193-713, 75.8 % | 46,765 B; file 755 lines; §9 starts 714; 46,765/61,706 = 75.79 % | yes |
| R3 residue | ~179,500 B | 223,144 − 43,638 = 179,506 B | yes |
| repo-root `MEMORY.md` | absent | absent | yes |
| guard symbols | `alwaysLoadedSurface()`, `AlwaysLoadedTokenBudget = 76000`, "별도 카드" note | `internal/config/token_budget_guard.go:107, 32, 29-31` | yes |
| guard test green | passes on this tree | `go test ./internal/config/ -run 'Budget\|AlwaysLoaded'` → ok | yes |

The 14,360 B figure the dispatch asked me to check is reproducible from the command `design.md:36-42`
prints verbatim, and the derivation that rests on it is arithmetically sound in both directions.

On the entry premises: `codex-probe.md` genuinely discharges P1-P3 by observation
(`codex debug prompt-input`, zero model calls) — a byte ruler locates the 32,768 B cut, a
three-level fixture establishes the git-root→CWD merge scope, and stderr 0 / exit 0 establishes the
silence. The SPEC carries no open premise gate: `grep -rn 't91'` over the directory returns nothing,
and `plan.md:37` states "No premise gate remains". P4 is correctly absent as a blocker. The
milestone set is M1-M6 with the nested-map milestone deleted, and `design.md:171` lists the card's
nested proposal under *rejected* alternatives with the measured reason — Option A is decided, not
offered.

---

## Defects Found

**D1 — `AC-AMC-017` is unsatisfiable as written; the ratchet has no working enforcement criterion.**
`acceptance.md:75-78` — "Then it exits 0 with `AlwaysLoadedTokenBudget` at or below 75,000, and the
achieved token figure is **quoted from that run's output**". `TestAlwaysLoadedTokenBudget`
(`internal/config/token_budget_guard_test.go:62-69`) emits the token total **only via `t.Errorf` on
failure**; a passing run prints `ok github.com/modu-ai/moai-adk/internal/config` and nothing else —
confirmed by running it. The criterion therefore demands a passing run and a failing run at once.
`AC-AMC-001` (`acceptance.md:8-10`, "its output is recorded as the pre-diet baseline") fails the
same way.
Compounding it, nothing checks the **derivation**. `REQ-AMC-013` says the constant is "derived from
the achieved post-diet measurement"; `plan.md:109` says "achieved figure plus a stated headroom
ratio". No criterion compares the two, so a run that lands the surface at 60,000 tokens and sets the
constant to exactly 75,000 satisfies `AC-AMC-017` verbatim while ratcheting nothing. This is the
vacuity the dispatch predicted, one level below where it was expected — not in the choice of tree,
but in the criterion's own mechanics.
Severity: critical — Class: blocking — Required fix: (a) add
`t.Logf("always-loaded surface = %d tokens (budget %d, headroom %d)", total, AlwaysLoadedTokenBudget, AlwaysLoadedTokenBudget-total)`
to `TestAlwaysLoadedTokenBudget` ahead of the over-budget check, so a passing run emits the figure
under `-v`; (b) restate `AC-AMC-001` and `AC-AMC-017` against
`go test -v ./internal/config/ -run 'Budget|AlwaysLoaded'` and quote that logged line; (c) add a
criterion asserting `AlwaysLoadedTokenBudget == achieved + stated_headroom` within a stated
tolerance, so the constant cannot sit at the ceiling independent of the achieved figure.

**D2 — `AC-AMC-010` fails on a correct tree: it contradicts the SPEC's own M6, and is
nondeterministic under worktrees.**
`acceptance.md:48-49` — "When `find . -name AGENTS.md -not -path './.git/*'` runs, Then exactly one
result is returned". Two independent failure modes, both measured against the file this repo already
has in `AGENTS.md`'s position:

1. **The template mirror.** `REQ-AMC-015` (`spec.md:192-193`) and M6 (`plan.md:115`) require every
   root-level shipped file to be mirrored into `internal/template/templates/`. That mirror exists
   today for the analogue: `internal/template/templates/CLAUDE.md`, 20,523 B, byte-identical to
   `./CLAUDE.md`. The moment M6 mirrors `AGENTS.md`, the count is 2 — on a tree that is correct.
   **`AC-AMC-010` and `REQ-AMC-015` cannot both be satisfied.**
2. **L1 worktrees.** Card worktrees are created at `.claude/worktrees/<name>` *inside the primary
   checkout* and are full checkouts — this audit is running from
   `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t82`, whose own root carries `CLAUDE.md`.
   `find .` from the primary checkout descends into every live lane, so the count varies with how
   many lanes happen to exist when the criterion runs. A criterion whose result depends on unrelated
   concurrent work cannot decide anything.

Measured directly: `find . -name 'CLAUDE.md' -not -path './.git/*' | wc -l` → **7** in this
worktree alone (repo root, five nested under `internal/`, one template mirror), with no
`.claude/worktrees/` present here to inflate it further.
Severity: critical — Class: blocking — Required fix: replace the global filename count with a
criterion that names its scope and asserts the property `REQ-AMC-005` actually cares about — that
**no `AGENTS.md` exists in the live tree outside the repository root**. Something of the shape
`git ls-files '*AGENTS.md' ':!internal/template/templates/'` (tracked files only, so worktrees and
build output cannot contribute, mirror explicitly excluded) returning exactly `AGENTS.md`. While
editing, state in `AC-AMC-009` which copy the 24,576 B ceiling binds — the live root file, the
template mirror, or both — since the mirror is what users receive.

**D3 — "the integration branch" is load-bearing and undefined.**
`spec.md:186`, `plan.md:110`, `design.md:162`, `acceptance.md:75`. `REQ-AMC-014` exists precisely to
stop the ratchet being measured on a convenient tree, and the SPEC states the hazard correctly (this
worktree measures ≈71,212, already under 75,000; the state that forced the raise measured 75,282).
But no artifact says which branch qualifies or how a tester decides, so a run-phase actor can
declare any branch "the integration branch" — including this one. The distinguishing property is not
the branch's name but that it carries the merged state of the sibling cards.
Severity: major — Class: blocking — Required fix: name the discriminator in `REQ-AMC-014` and
`AC-AMC-017` — the release/batch branch the card merges into — plus the recording commands
(`git rev-parse --abbrev-ref HEAD`, `git rev-list --count main..HEAD`) so the measured tree is
identified in the evidence rather than asserted.

**D4 — `AC-AMC-002` cites a fixture that will not exist at run-phase.**
`acceptance.md:12-15` requires re-running "the three commands in `.moai/reports/t82/codex-probe.md`
§ 검증 재현". That block (`codex-probe.md:101-106`) opens with
`fixture=<scratchpad>/codexprobe` and records only the *invocations* — the fixture's construction
(git init, three `AGENTS.md` levels, the 110-byte-interval ruler to 40,040 B) is nowhere recorded.
The path resolves to a session-local scratchpad that does not survive the session. A third party
cannot execute this criterion. The same gap applies to `codex-probe-p4.md`, whose four-way
differential depends on a `CODEX_HOME` fixture that is likewise unrecorded.
Severity: major — Class: blocking — Required fix: record the fixture construction — literal commands
or a small generator script committed under `.moai/reports/t82/` — and have `AC-AMC-002` cite that
rather than a scratchpad path.

**D5 — the `[HARD]` line grep is presented as a measurement of the contract, with no proxy
disclosure.**
`spec.md:89-93` calls 32,543 B the "verbatim Codex-relevant contract"; `design.md:13` calls it
"Contract. Always-loaded. Non-negotiable."; `progress.md:14-15` records the commands. The commands
are honest and reproduce exactly — but they count *lines bearing the marker*, which is not the same
quantity as *the contract*, and no artifact says so. Both error directions are present:
- **Undercount, unbounded.** A `[HARD]` lead line whose obligation continues into a list, table, or
  fenced block contributes only its first line. `kanban-dispatch.md:69` ("A dispatch is a
  fixed-field address block, not prose. The fields:") is followed by the field block at lines 71-83,
  uncounted; `:98` continues into a numbered list and code block at 100-107, uncounted. Measured:
  **16 of the 93 rule lines end in `:`** — structurally incomplete sentences whose bodies fall
  outside the total.
- **Overcount.** **15 of 93** carry the marker non-clause-initially — prose mentions and inline
  references. `kanban-dispatch.md:7` ("The stub keeps every [HARD] rule and pointer") is a
  navigation note counted as contract.
This matters because `spec.md` §D.1 derives `REQ-AMC-004`'s 24,576 B ceiling from the figure and
`AC-AMC-006` extrapolates M1's ratio against it. `design.md:181-183` raises a *different* concern
(clauses carrying no marker at all); the proxy's own error bars are raised nowhere. The design
survives the number moving — M1's stop condition halts and renegotiates rather than overrunning,
which is the right shape — so this is a disclosure and method defect, not a broken design.
`.moai/reports/t82/pending-spec-edits.md` item 4 already anticipates this and independently
reproduces the 15/93 count; it is retained as a finding so the fix lands with the verdict rather
than at the author's discretion.
Severity: major — Class: blocking — Required fix: state in `spec.md` §A.4 and `design.md` §1 that
the figure is a line-level proxy, citing both error directions with the counts above; make M1's
classification (`plan.md:55-57`, `AC-AMC-003`) operate on **clause blocks** — marker line plus its
continuation to the next clause or heading — rather than grep lines, and re-derive the ceiling
against the clause-block figure once M1 produces it.

**D6 — the rejection of the cap-raise lever rests on a premise that measurement contradicts.**
`design.md:86-88` — "A third lever — raising `project_doc_max_bytes` — is deliberately **not** used.
It is a **per-user config setting**, so it cannot help distributed users" — and `design.md:172`,
"Per-user config; cannot ship."
`.moai/reports/t82/codex-probe-p4.md` (committed `9330bd321`) measures otherwise: a project-scope
`<repo>/.codex/config.toml` **does** take effect, and beats the user value, when the project is
registered `trust_level = "trusted"` in the user config — established by a four-way differential
that toggles only the trust entry (rows 3 → 4, effective cap 4,096 → 8,192).
The **decision** is unaffected and in fact strengthened: the same report's §설계 귀결 #3 concludes
that cap-raising must not substitute for the diet, because a distributed user's *first* session is
untrusted, the effective cap is then 32,768, and misapplication is **silent** (stderr 0 bytes on all
four runs). So the SPEC reaches the right conclusion from a stated reason that is false, and the
true reason is the stronger one.
`pending-spec-edits.md` item 2 already schedules the constructive half (moving P4 to discharged and
adding a `[HARD]` prohibition); this finding covers only the factual contradiction, which must not
survive the edit.
Severity: major — Class: blocking — Required fix: correct `design.md:86-88` and `design.md:172` to
state the measured position — project scope works only under trust registration, the untrusted
first session is the binding case at 32,768 B, and non-application is silent — citing
`codex-probe-p4.md`.

**D7 — the nested-instruction-file asymmetry is unaddressed.**
This repository already runs five nested per-directory instruction files on the Claude side —
`internal/{cli,config,hook,spec,template}/CLAUDE.md` — so "one instruction file per repository" is
not the shape this codebase has. `REQ-AMC-005` (`spec.md:148-150`) and `design.md` §3 justify the
single-root design purely from the codex measurement and never mention the prior art;
`grep -rn 'nested CLAUDE\|per-directory'` over the SPEC directory returns nothing. A reader will ask
why `AGENTS.md` must be singular when `CLAUDE.md` is not, and the SPEC does not answer.
The measured answer already exists in the artifacts and needs only to be stated: nested `AGENTS.md`
share one 32,768 B budget consumed root-first and are not loaded at repo-root CWD
(`codex-probe.md` §2), whereas nested `CLAUDE.md` are loaded by path relevance under no such cap.
Severity: minor — Class: blocking — Required fix: add the asymmetry to `REQ-AMC-005`'s justification
or `design.md` §3, naming the five existing nested `CLAUDE.md` files as the prior art and the two
measured properties that distinguish the cases.

**D8 — `REQ-AMC-006` has no acceptance criterion.**
`spec.md:152-154` binds a *future* SPEC ("that SPEC shall state the evidence explicitly and
re-derive the root ceiling"), so nothing in this SPEC's Definition of Done can verify it.
Severity: minor — Class: blocking — Required fix: move it into `§D` as a recorded design note, where
an unverifiable forward constraint belongs, or add a criterion checking that *this* SPEC's body
records the revival condition — not the future SPEC's compliance with it.

**D9 — `AC-AMC-012` names no deciding mechanism.**
`acceptance.md:57-58` — "When a Claude Code session starts, Then the contract text is present in
context exactly once". Duplicate injection is the right hazard (`design.md:132-133`), but "present
in context" has no command, file, or observable behind it, so PASS/FAIL rests on judgment.
Severity: minor — Class: blocking — Required fix: bind it to something observable — that the clause
set appears exactly once across `CLAUDE.md` and the imported `AGENTS.md`, checked by a
duplicate-line scan over the resolved import set, with the command named.

**D10 — `REQ-AMC-004` is labelled `(Ubiquitous)` but written in the GEARS Unwanted form.**
`spec.md:144` — "shall not exceed 24,576 B". The sentence matches a GEARS pattern, so MP-2 passes;
the label contradicts the form it labels.
Severity: minor — Class: optional — Required fix: relabel `(Unwanted)`, or rephrase positively.

**D11 — `REQ-AMC-006` leads with `MAY` in normative text.**
`spec.md:152-154` — "a nested `AGENTS.md` … **MAY** be proposed". Permissive modality is what RQ-5
excludes; the binding half ("that SPEC **shall** state the evidence") is correct.
Severity: minor — Class: optional — Required fix: recast as a constraint on the proposing SPEC, or
fold into `§D` per D8.

**D12 — two requirements carry their rationale inside the clause.**
`spec.md:148-150` (`REQ-AMC-005`) and `spec.md:163-164` (`REQ-AMC-009`, "Because truncation is
measured silent, …"). This is the inline-justification pattern `design.md:81-84` identifies as the
compression target — the SPEC exhibits it while proposing to remove it elsewhere.
Severity: minor — Class: optional — Required fix: none required for PASS.

---

## Regression Check

Not applicable — iteration 1. The SPEC has not yet been revised in response to a verdict; the
v0.1.0 → v0.2.0 rebuild was author-initiated before any verdict existed.

Recorded for iteration 2, because it is the failure mode most likely to recur: the v0.1.0 set
asserted P1-P3 unmeasured and carried a live nested-`AGENTS.md` design while `codex-probe.md` —
written minutes earlier in the same worktree — had measured all three and overturned the nested leg.
The rebuild fixed it unprompted. D6 is the same shape at smaller scale: `codex-probe-p4.md` landed
after `design.md`, and `design.md`'s rationale still states the pre-measurement position. A
measurement landing after the artifact that depends on it needs a sweep of the dependents, not just
of the section that names it.

---

## Recommendation

FAIL. No must-pass criterion failed, the design decisions are sound, and every asserted figure
reproduces — including the 14,360 B measurement the ceiling now rests on. The gap is concentrated
in the acceptance criteria: **five of twenty-one cannot decide what they claim**, and two of those
five carry the SPEC's two headline properties (the ratchet, the single-document rule). A SPEC whose
enforcement criteria cannot be executed will report success without having verified anything —
precisely the failure `spec.md` §D.4's measurement-provenance constraint exists to prevent.

Blocking fixes, in the order they should be made:

1. **D2** — replace `AC-AMC-010`'s global filename count with a tracked-files check that excludes
   the template mirror. As written the SPEC cannot satisfy M6 and `AC-AMC-010` together, so this is
   the fix that unblocks the milestone set.
2. **D1** — add the `t.Logf`, restate `AC-AMC-001` / `AC-AMC-017` against `go test -v`, and add the
   derivation criterion. Without it the SPEC's headline goal is unverifiable.
3. **D3** — define "integration branch" with a discriminator and a recording command.
4. **D6** — correct the cap-raise rationale to the measured position (lands with
   `pending-spec-edits.md` item 2).
5. **D4** — record the probe fixtures' construction so `AC-AMC-002` is executable by someone else.
6. **D5** — disclose the line-grep proxy with both error directions; move M1's classification to
   clause blocks (lands with `pending-spec-edits.md` item 4).
7. **D7** — state the nested-`CLAUDE.md` asymmetry in `REQ-AMC-005`'s justification.
8. **D9** — give `AC-AMC-012` an observable. **D8** — cover `REQ-AMC-006` or move it to `§D`.

Optional, surfaced and left to the orchestrator (routing these into a revision is **not** required
to reach PASS): D10, D11, D12.

Expected effect: D1, D2, and D9 lift Testability out of the 0.50 band; D2 and D8 restore
Traceability; D3 lifts Clarity; D5, D6, and D7 lift Completeness. The re-audit is scoped to this
enumerated delta plus a regression pass over it, not a from-scratch review.

Two parked edits (`pending-spec-edits.md` items 1 and 3 — the output-style §8 lineage correction
and the M4 trust-notice obligation) were checked for contradiction against the frozen artifacts and
none was found; they are not defects and are not counted in the score.

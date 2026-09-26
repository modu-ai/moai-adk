# SPEC Review Report: SPEC-LOCAL-INSTRUCTIONS-MIGRATE-001

Card: t1259 · Iteration: 1/3 · Tier: L (threshold 0.85)
Tree: `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t1259` · Branch `WT-local-instructions` · HEAD `5ba87003f`
Date: 2026-09-26

**Verdict: PASS-WITH-DEBT**
**Overall Score: 0.85** (harmonic mean of the four dimensions below, at the Tier L threshold)

Reasoning context ignored per M1 Context Isolation. The dispatch's list of plan-phase decisions was
read as a list of claims to test; each is judged below against this tree's text and this tree's
mechanical measurements, not against the author's account of them.

Input contract: Tier L → all five artifacts read (`spec.md`, `plan.md`, `acceptance.md`,
`design.md`, `research.md`), plus `progress.md` §E.1.

---

## Must-Pass Results

- **[PASS] MP-1 REQ number consistency.** Nine ids, zero-padded, no duplicates:

  ```
  $ grep -n -o '^- \*\*REQ-IFU-[0-9]\{3\}' spec.md
  75:- **REQ-IFU-007   77:- **REQ-IFU-008   82:- **REQ-IFU-009   85:- **REQ-IFU-010
  93:- **REQ-IFU-011   95:- **REQ-IFU-012  100:- **REQ-IFU-020  106:- **REQ-IFU-021
  112:- **REQ-IFU-022
  ```

  The gaps (`001~006`, `013~019`, `023~025`) are the carve's footprint and are declared as such at
  `spec.md:70-73`; they resolve to the parent SPEC, which exists in this tree. A gap that is
  documented, attributable, and required to keep prior audit citations resolving is not an id
  defect.

- **[PASS] MP-2 GEARS format compliance — judged against the `REQ-XXX` requirement layer only.**
  Every one of the nine matches a GEARS pattern: `Where` (`007`, `010a`, `010b`), Ubiquitous
  `shall` (`008`, `009`, `012`, `020`, `021`, `022`), Unwanted `shall not` (`010` head, `011`).
  `REQ-IFU-010`'s two lettered clauses are sub-clauses of one requirement, not requirements, and
  each carries its own `Where … shall …` (`spec.md:85-92`). The verification layer
  (`AC-IFU-*`, Given-When-Then) was graded under Group 4, never here.

- **[PASS] MP-3 YAML frontmatter validity.** All 12 canonical fields present, correct types, no
  rejected snake_case alias. `version: "0.2.0"` quoted; `created`/`updated` ISO; `priority: P1`;
  `lifecycle: spec-anchored`; `tags` comma-separated string. `tier: L` is an additive 13th field,
  not a substitution (`spec.md:1-14`).

- **[N/A] MP-4 language neutrality.** Single-language SPEC — the code half is Go (`internal/cli`),
  the docs half is locale-scoped (ko/en/ja/zh), which is the *locale* axis, not the 16
  programming-language axis. No language-specific tool is named as primary.

- **[PASS] MP-5 D7 cross-SPEC reconciliation.** One referenced SPEC:

  ```
  $ grep -o 'SPEC-[A-Z][A-Z0-9-]*-[0-9]\{3\}' *.md | sed 's/.*://' | sort -u
  SPEC-INSTRUCTION-FILES-UNIFY-001
  SPEC-LOCAL-INSTRUCTIONS-MIGRATE-001
  $ grep '^status:' .moai/specs/SPEC-INSTRUCTION-FILES-UNIFY-001/spec.md
  status: draft
  ```

  `draft` ∉ {retired, superseded, archived}. No BLOCKING finding.

- **[PASS] MP-6 D8 cross-platform discipline.** `grep -c 'syscall' spec.md` → `0`. Auto-PASS per
  D8-4.

- **[PASS] MP-7 clarification gate.**

  ```
  $ grep -rn 'NEEDS CLARIFICATION' .moai/specs/SPEC-LOCAL-INSTRUCTIONS-MIGRATE-001/
  (no output; grep exit=1)
  ```

---

## Category Scores

| Dimension | Score | Band | Evidence |
|---|---|---|---|
| Clarity | 0.90 | 0.75–1.0 | Every decision states its alternative and why it lost (`design.md` §A, §B). One soft spot: `REQ-IFU-012` carries its condition as a trailing `when` clause rather than a leading `When`. |
| Completeness | 0.95 | 1.0-adjacent | HISTORY, §A Context, §B Goal, §C Requirements, §D Out of Scope (four `### Out of Scope — <topic>` H3s at `spec.md:151,161,169,176`, each with specific bullets), acceptance matrix, Tier L's `design.md` + `research.md` both present and substantive. |
| Testability | 0.65 | 0.50–0.75 | Four of ten criteria carry a demonstrated defect — D1 (AC-IFU-023 provably fails on current state), D2 (AC-IFU-029 passes today, before any work), D3 (AC-IFU-031's evidence source does not exist in the governing regime), D4 (AC-IFU-024 names a command that cannot show what it asserts). |
| Traceability | 1.00 | 1.0 | The SPEC's own §D.2 verification command run in this tree: `diff <(grep -o 'REQ-IFU-[0-9]\{3\}' acceptance.md \| sort -u) <(grep -o '^- \*\*REQ-IFU-[0-9]\{3\}' spec.md \| grep -o 'REQ-IFU-[0-9]\{3\}' \| sort -u)` → **empty output, exit 0**. Ten criteria, ten requirement clauses, one-to-one. |

Harmonic mean: `4 / (1/0.90 + 1/0.95 + 1/0.65 + 1/1.00)` = `4 / 4.7022` = **0.851**.

---

## What the plan phase claimed, and what reproduced

Every figure the plan phase changed was re-measured in this tree rather than accepted.

| Claim | Command | Output | Verdict |
|---|---|---|---|
| `CLAUDE.local.md` is 44,740 chars, not the carved 44,381 | `git show origin/develop:CLAUDE.local.md \| wc -m` | `44740` | **reproduces** |
| both develop copies agree | `git show develop:CLAUDE.local.md \| wc -m` | `44740` | **reproduces** |
| the carved glob matched the empty set | `ls docs-site/content/ko/claude-md-guide.md` | `No such file or directory` | **reproduces** |
| all six pages now resolve, 24 files | per-path `[ -f ]` loop over 6 paths × 4 locales | `24 OK`, zero MISS | **reproduces** |
| `memory` is the context-memory page, not the CLI ref | `grep -c 'CLAUDE.local.md\|CLAUDE.md'` on both ko copies | `0` / `28` | **reproduces** |
| Q4: producer is pure, sole caller holds the stderr | `grep -rn 'codexLocalDeveloperInstructionArgs' internal/cli/` | one def (`:123`), one call (`:825`) | **reproduces** |
| Q4: that caller already writes operator diagnostics | `grep -n 'ErrOrStderr' internal/cli/codex_launcher.go` | install hint `:804`, worktree err `:842`, both inside `runCodexLaunch` (`:801`) | **reproduces** |
| `DualFileMatrix` exists and asserts the literal preamble | read `codex_local_instructions_test.go:175-213` | builds `want += "<!-- source: " + name + " -->\n"` | **reproduces** |
| `FallbackAdvisory` does not exist | `grep -rn 'func TestCodexLocalInstructions' internal/cli/` | 19 symbols, none is `_FallbackAdvisory` | **reproduces** |

Nine of nine. The plan phase's own measurements are sound, and the two figures it corrected were
corrected correctly. The defects below are not in what it measured — they are in what it did not
measure.

---

## Defects Found

**D1. AC-IFU-023 locale-parity clause — provably fails against current state.** —
`acceptance.md:214-219` — Severity: **major** — Class: **blocking**

The criterion asserts "the per-page `^## ` section count is **equal** across the four locales".
Measured in this tree:

```
$ for P in <the six paths>; do printf "%-45s" "$P"; \
    for L in ko en ja zh; do printf "%s " "$(grep -c '^## ' docs-site/content/$L/$P)"; done; echo; done
advanced/claude-md-guide.md                  10 18 18 18
advanced/codex-dual-harness.md               6 6 6 6
advanced/harness-learning.md                 6 6 6 6
claude-code/context-memory/memory.md         7 7 7 7
getting-started/quickstart.md                14 10 10 10
cli-reference/update.md                      7 7 7 7
```

Two of the six pages are already unequal, in opposite directions, by 8 and by 4 sections. Neither
`REQ-IFU-020` nor any Out of Scope clause asks for a locale restructure of `claude-md-guide.md` or
`quickstart.md`, and `spec.md` §D explicitly scopes the docs work to *describing the three-file
structure*. As written the criterion therefore either (a) silently absorbs an unscoped
twelve-section repair into M4, or (b) fails for an implementation that does exactly what
`REQ-IFU-020` asks.

This is the **same criterion and the same class** the plan phase just repaired on its other half
(the vacuous glob). The glob half was measured; the parity half was not.

**Required fix:** either narrow the parity assertion to the *sections the change adds* (e.g. every
locale of a page gains the same new `^## ` heading), or state the pre-existing inequality as a
measured baseline the criterion is asserted against, or move the parity repair into scope
explicitly with its own requirement. Do not leave an equality assertion over a set that is
currently unequal.

---

**D2. AC-IFU-029 is discharged by pre-existing state — it passes today, before any work.** —
`acceptance.md:66-74` — Severity: **major** — Class: **blocking**

```
$ unset MOAI_KANBAN … && go test ./internal/cli/ -run '^TestCodexLocalInstructions_DualFileMatrix$' -v
--- PASS: TestCodexLocalInstructions_DualFileMatrix (0.01s)
    --- PASS: … /body/body (0.00s)
PASS
ok  github.com/modu-ai/moai-adk/internal/cli  0.817s
```

The criterion's entire verification is this command, and this command passes at `5ba87003f` with
none of the SPEC's work done. `REQ-IFU-008` is already satisfied by
`codex_launcher.go:138` (`fmt.Fprintf(&payload, "<!-- source: %s -->\n", name)`).

This is the exact class `acceptance.md`'s own [HARD] head block forbids: *"ask not 'does this pass
when the work is done' but 'can this pass when it is not'."* It can. The anchoring repair fixed the
sibling-match face of vacuity and left the already-satisfied face untouched — the criterion is
well-anchored and still non-discriminating.

The honest reading is that `AC-IFU-029` is a **regression guard**, not a verification, and
`REQ-IFU-008` is a **preservation requirement**, not new behaviour. Nothing in the SPEC says so, so
a run phase reading §D.3 ("all ten criteria pass") gets a free pass on one of ten.

**Required fix:** declare `REQ-IFU-008` as already-satisfied-and-guarded in `spec.md` §C.1 and mark
`AC-IFU-029` a regression guard with its passing baseline recorded (`5ba87003f`, the output above),
so a *later* failure is meaningful; or extend the criterion to assert the preamble under the new
fallback-advisory path, which is the behaviour this SPEC actually adds.

---

**D3. AC-IFU-031 requires evidence from a PR head that this card will never produce.** —
`acceptance.md:234-237`, `acceptance.md:308` (§D.3 DoD), `plan.md:140` — Severity: **major** —
Class: **blocking**

The criterion: *"Given the whole change on its PR head, When CI completes … The verdict is read
from the PR head's own run — a local pass, or a run against an earlier head, does not discharge
it."* §D.3 lists it as a Definition-of-Done item.

`plan.md:82` (§C) states the governing regime in the same artifact set: *"The lane does not push.
Integration is a lead-granted window; push is the lead's batch."* Under
`.claude/rules/local/gitflow-lane-protocol.md` and `CLAUDE.local.md` §4.1 the card branch never
opens a PR; only a `release/vX.Y.Z` branch PRs to `main`, and that head carries many cards, so it
cannot attribute a verdict to *this* change. The artifacts therefore contain a Definition-of-Done
item the lane is structurally unable to discharge.

`progress.md:191-194` discloses the Route A/B tension and defers it to the lead's dispatch — which
is the right disclosure, and it is why this is a defect in the *binding* artifacts rather than a
process gap. The disclosure lives in `progress.md`; the unsatisfiable obligation lives in
`acceptance.md` §D.3, and a run phase reads the latter.

**Required fix:** re-site the criterion's evidence on what this regime actually produces — the CI
run on the `develop` head that carries the lane's merge commit, named by that merge SHA, which is
exactly how every other card in this repository closes. Keep the criterion's real content (clean
environment, not the lane's machine; both `go test ./internal/cli/...` and the docs-site build);
change only where the head comes from.

---

**D4. AC-IFU-024 names a command that cannot show what the criterion asserts.** —
`acceptance.md:207-210` — Severity: **minor** — Class: **blocking**

The criterion requires that `grep -c 'CLAUDE.local.md' AGENTS.local.md` *"reports only
historical-reference occurrences, each within a sentence marking it as a retired filename."*
`grep -c` emits a single integer. It cannot report which occurrences they are, and it cannot show
the sentence around any of them — so the assertion's second half is not decidable from the named
command's output. As written the criterion resolves to a per-occurrence human judgment with a
command attached that does not feed it.

**Required fix:** name `grep -n` (or `grep -n -C1`), so the occurrences and their surrounding lines
are in the output the verdict is read from, and state the passing condition per line rather than in
aggregate.

---

**D5. `AC-IFU-030` fixes the advisory to stdout without a measurement behind the choice.** —
`acceptance.md:138-141` — Severity: **minor** — Class: **optional**

The criterion reads *"each command's stdout"*. For `doctor` this matches the code
(`internal/cli/doctor.go:74` `out := cmd.OutOrStdout()`, `:80` printer primary writer) and for
`update` the primary writer is likewise stdout (`internal/cli/update.go:153`) — so the choice is
probably right. But `design.md` §B establishes *stderr* as the operator-diagnostic surface for the
launcher advisory (`AC-IFU-011`), and the SPEC nowhere records why the same message goes to a
different stream on the other two commands. A run phase implementing "operator diagnostics go to
stderr" consistently would fail a correct implementation.

**Required fix (optional):** one sentence recording the measurement above and the deliberate
asymmetry — launcher advisory on stderr because the payload is stdout-adjacent, `update`/`doctor`
advisory on stdout because that is those commands' report stream.

---

**D6. `REQ-IFU-012` carries its condition as a trailing clause.** — `spec.md:95-96` — Severity:
**minor** — Class: **optional**

*"`moai doctor` shall report the same advisory as `moai update` when both local instruction files
exist or when only `CLAUDE.local.md` exists."* GEARS event-driven form leads with the trigger. The
meaning is unambiguous, which is why this is optional rather than an MP-2 failure; the eight
siblings all lead with theirs.

---

## Judgments the author explicitly referred to this audit

**The `REQ-IFU-010a`/`010b` sub-clause shape — accepted, and it is the right call.** Two clauses
under one id, rather than two ids. The contradiction it resolves is real and reproduces on the
current text: the requirement head commands no-coexistence and `AC-IFU-014` commands refusal, and
refusal leaves both files present. The data-integrity argument in `design.md` §A is correct and
correctly asymmetric — a tool that picks between two user-authored files fails quietly, and the
failure surfaces when instructions stop taking effect rather than when the command runs. Keeping
one id preserves the cross-references the carve exists to preserve, and the §D.2 table records the
two clauses on separate rows so the mapping is visible rather than inferred. The alternative (two
ids) buys nothing this shape lacks and costs exactly what the carve was built to avoid.

`design.md` §A's claim that no new criterion is needed was checked, not accepted: `AC-IFU-013`
fixtures one-file-present (`010a`), `AC-IFU-014` fixtures both-present (`010b`). Correct. The
sha256 addition to `AC-IFU-014` closes a real hole — "without modifying either file" was previously
unasserted and a refusal that wrote before exiting non-zero would have passed.

**Excluding `AC-IFU-031` from the §D.2 coverage table — accepted.** A whole-change criterion citing
the full requirement set would, if counted, let any per-requirement gap hide behind it. Excluding
it is the conservative direction and the table says so in its own text. Note that D3 above is about
the criterion's *evidence source*, not about this exclusion; the exclusion is sound either way.

**Withdrawing the fixed reduction floor — accepted, and it strengthens the criterion rather than
weakening it.** The floor was a constant derived from a quantity that moves: it moved 359
characters between the carve and this plan phase, and `research.md` §B is right that an arithmetic
repair to a stale constant makes it look authoritative. The bound that decides the requirement is
`after <= 39,999`, which does not drift, and the recording obligation (both values, each with its
command, before measured at the milestone) is what stops the criterion being discharged against an
already-compliant copy. The withdrawn floor added no failure mode the cap does not already catch —
a migration landing at 40,358 fails the cap regardless. No weakening.

---

## Gaps — what this audit did NOT observe

- `moai spec lint` did not complete within its budget (backgrounded at 120s, no output produced for
  this SPEC before the report was written). Its domain — REQ collection, out-of-scope rule,
  frontmatter — was covered by the direct greps above, but the lint's own verdict on this file is
  **unobserved**, not passed.
- The AC-counter figure (`live=10 excluded=3`) was supplied by the orchestrator and is consistent
  with the 10 `**AC-IFU-***` headers counted here, but the counter itself was not re-run.
- `SPEC-INSTRUCTION-FILES-UNIFY-001` M2's landing state, and t1175's — both are plan.md §B
  dependencies and both are the lead's read at dispatch, as `research.md` §E states.
- Whether `internal/cli` as a whole is green at this HEAD. Only the one named test was run.

## Residual risk

The four blocking defects are all in the acceptance layer and all cheap to repair; none touches the
SPEC's design, whose two decisions (`REQ-IFU-010` split, advisory at the caller) verified cleanly
against the source. The risk that remains after repair is D1's: if the parity clause is narrowed
rather than scoped, M4 could still land 24 files whose locale structure diverges further, and
nothing in this SPEC would catch it.

## Recommendation

**PASS-WITH-DEBT at 0.85.** The must-pass firewall is clean on all seven criteria, traceability is
exact, and every measurement the plan phase newly asserted reproduces. The debt is four acceptance
criteria, in this order:

1. **D1 (AC-IFU-023)** — before M4 is dispatched. It is the one defect that can silently expand a
   milestone's scope by twelve sections.
2. **D3 (AC-IFU-031)** — before close, and ideally now: it is a Definition-of-Done item the lane
   cannot discharge as written, and discovering that at close is the expensive moment.
3. **D2 (AC-IFU-029)** — before M2. One declaration sentence plus a recorded baseline.
4. **D4 (AC-IFU-024)** — before M3. One flag.

D5 and D6 are optional and left to the orchestrator; neither justifies a revision on its own.

🗿 MoAI

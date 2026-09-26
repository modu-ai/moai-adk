# SPEC Review Report: SPEC-LOCAL-INSTRUCTIONS-MIGRATE-001

Card: t1259 · Iteration: 2/3 · Tier: L (threshold 0.85)
Tree: `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t1259` · Branch `WT-local-instructions`
Subject: `bac73d358` (clean tree) · Prior iteration: `b8fb023e8`
Date: 2026-09-26

**Verdict: PASS-WITH-DEBT**
**Overall Score: 0.92** (harmonic mean; Tier L threshold 0.85)
**Score movement: 0.85 → 0.92.** No regression, so no STOP escalation.

Reasoning context ignored per M1 Context Isolation. The dispatch's account of each repair was read
as a claim to test. Per the iter-2 contract this audit is scoped to the enumerated iter-1 defect
delta plus a regression check over it, and to any new defect the repairs introduced.

---

## Regression Check — the six iter-1 defects

| # | Defect | Disposition |
|---|---|---|
| D1 | `AC-IFU-023` parity clause failed on measured state | **REPAIRED** |
| D2 | `AC-IFU-029` passed before any work | **REPAIRED** (introduces N1) |
| D3 | `AC-IFU-031` cited a PR head the lane never produces | **PARTIALLY REPAIRED** |
| D4 | `AC-IFU-024` named `grep -c` for a per-sentence assertion | **REPAIRED** |
| D5 | stdout-vs-stderr asymmetry unrecorded (optional) | **REPAIRED** |
| D6 | `REQ-IFU-012` trailing `when` (optional) | **REPAIRED** |

No iter-1 defect is unchanged; no stagnation.

---

### D1 — REPAIRED. The baseline is byte-accurate and the equal-delta clause is mechanically decidable.

The recorded table was re-measured independently in this tree, not compared to the author's
transcription:

```
$ for P in <the six paths>; do printf "%-42s" "$P"; \
    for L in ko en ja zh; do printf "%s " "$(grep -c '^## ' docs-site/content/$L/$P)"; done; echo; done
advanced/claude-md-guide.md               10 18 18 18
advanced/codex-dual-harness.md            6 6 6 6
advanced/harness-learning.md              6 6 6 6
claude-code/context-memory/memory.md      7 7 7 7
getting-started/quickstart.md             14 10 10 10
cli-reference/update.md                   7 7 7 7
```

All 24 figures match the table at `acceptance.md` §D.4 exactly. The 24 paths still resolve
(`24 OK`, zero MISS). `REQ-IFU-020` is unwidened — confirmed against the diff, `spec.md:100-105`
is untouched by `bac73d358`.

**Mechanical evaluability — the first question referred to this audit.** The binding clause reads
*"each page's four counts differ from the recorded pre-change baseline by the same delta"*. That
evaluates as `delta_L = after_L − before_L` for `L ∈ {ko,en,ja,zh}`, with the pass condition
`delta_ko = delta_en = delta_ja = delta_zh`, per page. No judgment call, no weasel term, and the
absolute values never enter. It is decidable.

**Baseline staleness — the second question, and the answer is that it cannot bite.** The table is
not the comparison operand: the criterion says the baseline *"is re-measured at M4 against that
milestone's own base, not read from this table"*. So a sibling card editing a page between now and
M4 moves both operands together and the delta is unaffected. Worked through the three ways the
table could go stale:

| Sibling card lands | Table numbers | Shape | Tripwire | Delta clause |
|---|---|---|---|---|
| +2 sections to `ko/claude-md-guide.md` only | stale | still 2 unequal / 4 equal | silent | unaffected — base re-measured |
| brings `claude-md-guide.md` to parity | stale | 1 unequal / 5 equal | **fires** | unaffected |
| makes `update.md` unequal | stale | 3 unequal / 3 equal | **fires** | unaffected |

The one case the tripwire stays silent on is the one where silence is correct — the shape it pins
is intact and the delta clause never reads the stale number. The table is documentation with a
tripwire attached, and that is exactly what it is described as. No residual.

**On the choice between the two repairs I offered:** the author took equal-delta over narrowing and
the reasoning is right, including on my own terms. Narrowing to the touched sections would have
left M4 free to land 24 files diverging further — my iter-1 residual-risk note — and equal-delta
catches precisely that. Declining to widen `REQ-IFU-020` is also correct: whether ko lacks eight
sections or en/ja/zh carry eight too many is a question this SPEC has no basis to answer, and
inventing an answer inside a docs-description milestone is how unscoped restructures get smuggled
in.

---

### D2 — REPAIRED. The criterion is red today. One narrow defect is introduced (N1 below).

```
$ unset MOAI_KANBAN MOAI_KANBAN_ID MOAI_KANBAN_LABEL MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED \
    && go test ./internal/cli/ \
       -run '^TestCodexLocalInstructions_DualFileMatrix$|^TestCodexLocalInstructions_FallbackAdvisory$' -v
--- PASS: TestCodexLocalInstructions_DualFileMatrix (0.00s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/cli	0.869s
```

The criterion requires the output to contain **both** `--- PASS: …_DualFileMatrix ` and
`--- PASS: …_FallbackAdvisory `. Only the first is present, so the criterion **fails at
`bac73d358`** with none of the work done. It can now discriminate, which is what D2 asked for.

Applying both repairs rather than one was the better call, and not merely the more thorough one:
the declaration alone would have left a labelled criterion that still cannot fail, and the
extension alone would have left `REQ-IFU-008` misrepresented as new behaviour. The pairing's stated
rationale — the advisory is emitted from the same launch path that builds the payload, so the
change most likely to break the preamble is this SPEC's own — is verified against the source: the
producer at `codex_launcher.go:123` and its sole caller at `:825` are one path.

**On the two-severities question referred to this audit: do not split.** Both halves assert the
same property (the literal preamble) on the same requirement. Splitting would put two criteria on
`REQ-IFU-008` and break the one-to-one the §D.2 table achieved at v0.2.0 — paying a traceability
regression to separate a guard from its own extension. The real cost of the pairing is N1, which is
a one-line fix and does not need a split to resolve.

---

### D3 — PARTIALLY REPAIRED. The head is re-sited correctly and verified; the content clause names a check the named run does not perform.

**The head half is repaired, and the evidence source exists.** The criterion now reads *"the CI run
for the `develop` head carrying this lane's merge SHA"*, and that run is real:

```
$ grep -n -A6 '^on:' .github/workflows/ci.yml
16:on:
17-  push:
18-    branches: [main, develop]
```

`ci.yml` runs on every push to `develop`, and its `test` job at `:229` runs
`go test -json -coverprofile=coverage.out -covermode=atomic ./...`, which covers
`./internal/cli/...`. Identifying the run by head SHA also keeps `origin/develop` out of
moving-ref territory — `moai spec lint` raises no `MovingRefUnpinned` on this file (below).

**The content half is not.** The criterion also asserts *"the run includes both
`go test ./internal/cli/...` **and the docs-site build**"*. No workflow in this repository builds
the docs-site:

```
$ grep -rn 'hugo' .github/workflows/
(no output)
$ grep -rln 'docs-site' .github/workflows/
.github/workflows/docs-i18n-check.yml
.github/workflows/test-install.yml
$ grep -n 'hugo\|build' .github/workflows/docs-i18n-check.yml
(no output)
$ grep -n '^  [a-z0-9_-]*:$' .github/workflows/ci.yml
44:  detect   116:  test   275:  test-race   340:  test-skip-marker
392:  test-integration   446:  lint   493:  build   563:  constitution-check   624:  test-browser
```

`docs-i18n-check.yml` is a content-drift check, not a build — and its `push` trigger is
path-filtered to `docs-site/content/**`, so it runs on the develop head only when that head touches
docs content. `ci.yml`'s `build` job builds the Go binary. The docs-site is deployed by Vercel,
which is a different system with its own head and its own verdict surface.

So the repair moved the head onto something real and left the payload clause pointing at something
that does not exist in it. **This is the same class as the original D3** — a criterion whose
evidence source is absent from where it points — one clause to the left of where it was found. It
is discovered here rather than at close, which is the whole value of the criterion, but it is not
closed.

**Required fix.** Name the checks the develop run actually produces, and site the docs verdict where
it actually lives. Concretely: `ci.yml`'s required checks plus `docs-i18n-check` (noting its path
filter, so M4's own commit is what triggers it), and either drop "the docs-site build" or re-site it
on the Vercel deployment for that head, named as such. Do not leave a Definition-of-Done clause
citing a build no run performs.

---

### D4 — REPAIRED.

`acceptance.md` now reads: *"When `grep -n -C1 'CLAUDE.local.md' AGENTS.local.md` is run, Then
**every numbered occurrence in that output** sits within a sentence marking the name as retired or
historical. The verdict is read per line, not in aggregate: one unmarked occurrence fails the
criterion however many marked ones surround it."* The named command emits the occurrences with
their surrounding lines, so the assertion is now decidable from the output the verdict is read
from, and the aggregate escape is closed explicitly.

### D5 — REPAIRED, and the scoping judgment referred to this audit holds.

The premise is recorded with its measurement (`doctor.go:74` + the printer at `:80`,
`update.go:153`), which I confirmed independently at iter-1. The rule stated — *each command's own
report stream* — is the correct generalization rather than "diagnostics go to stderr", and it is
what makes the launcher's stderr and `update`/`doctor`'s stdout one rule instead of two.

**The claim that `design.md` §B needs no change is correct.** §B scopes itself in its own opening
line — *"Whether the Codex launcher's fallback branch has a diagnostic surface"* (`design.md:66`) —
and its only normative sentence attributes the stderr assertion to `AC-IFU-011` alone
(`design.md:99`). It never generalizes to `update` or `doctor`. Nothing there contradicts
`AC-IFU-030`.

### D6 — REPAIRED.

`spec.md:105-106`: *"When both local instruction files exist, or when only `CLAUDE.local.md`
exists, `moai doctor` shall report the same advisory as `moai update`."* GEARS event-driven, trigger
leading.

---

## Gaps from iter-1 — both closed, both re-verified here rather than accepted

**`moai spec lint`.** The author is right about the cause: my 120-second budget went on a
full-corpus scan. Scoped, it returns immediately. Run here at `bac73d358`:

```
$ unset MOAI_KANBAN … && ~/go/bin/moai spec lint SPEC-LOCAL-INSTRUCTIONS-MIGRATE-001
✓ No findings — all SPEC documents are valid
exit=0
```

My iter-1 full-corpus run also completed afterwards (exit 0, 0 errors / 4845 warnings corpus-wide)
and produced **no finding against this SPEC id** — so the file was clean before the repairs as well
as after. Notably `MovingRefUnpinned`, which the corpus run does raise elsewhere, does not fire on
this file's `origin/develop` citations, because both pin the measurement (`AC-IFU-007` by measuring
at the milestone, `AC-IFU-031` by naming the head SHA).

**AC counter and traceability.** Re-derived here:

```
$ grep -c '^\*\*AC-IFU-[0-9]\{3\}\*\*' acceptance.md   → 10
$ grep -c '^- \*\*REQ-IFU-[0-9]\{3\}' spec.md          → 9   (10 clauses; REQ-IFU-010 carries two)
$ diff <(grep -o 'REQ-IFU-[0-9]\{3\}' acceptance.md | sort -u) \
       <(grep -o '^- \*\*REQ-IFU-[0-9]\{3\}' spec.md | grep -o 'REQ-IFU-[0-9]\{3\}' | sort -u)
(empty; exit 0)
```

Consistent with the `10 / 10` the author reports. The repairs added no criterion and moved no id.

---

## Must-Pass Results (re-checked at `bac73d358`)

- **[PASS] MP-1** — nine ids, no duplicates, consistent padding; documented carve gaps unchanged.
- **[PASS] MP-2** — all nine GEARS-conformant; `REQ-IFU-012` improved to leading `When` (D6).
- **[PASS] MP-3** — 12 canonical fields intact; only `version` moved, `"0.2.0"` → `"0.2.1"`, still a quoted semver string.
- **[N/A] MP-4** — single-language SPEC.
- **[PASS] MP-5** — sole cross-SPEC reference `SPEC-INSTRUCTION-FILES-UNIFY-001` is `status: draft`; no BLOCKING.
- **[PASS] MP-6** — `grep -c 'syscall' spec.md` → `0`.
- **[PASS] MP-7** — no `[NEEDS CLARIFICATION]` markers in `plan.md` or `research.md`.

## Category Scores

| Dimension | Score | Band | Movement | Evidence |
|---|---|---|---|---|
| Clarity | 0.95 | 0.75–1.0 | 0.90 → 0.95 | D6 closed; every repair states the alternative it rejected and why. |
| Completeness | 0.95 | 0.75–1.0 | unchanged | All sections and both Tier L artifacts intact; no section lost to the repairs. |
| Testability | 0.80 | 0.75–1.0 | 0.65 → 0.80 | Three of four blocking defects fully closed and mechanically re-verified; D3's content clause open; N1 introduced. |
| Traceability | 1.00 | 1.0 | unchanged | §D.2 diff empty at this HEAD; 10 criteria, 10 requirement clauses, one-to-one preserved through the repairs. |

Harmonic mean: `4 / (1/0.95 + 1/0.95 + 1/0.80 + 1/1.00)` = `4 / 4.3553` = **0.918**.

---

## Defects Found

**N1. The anti-vacuity positive control in `AC-IFU-029` is neutered by its own alternation.** —
`acceptance.md:70-79` — Severity: **minor** — Class: **blocking** — *new, introduced by the D2
repair*

The criterion retains the file's [HARD] guard *"must not contain `no tests to run`"*, but the
alternation `'^…_DualFileMatrix$|^…_FallbackAdvisory$'` makes that guard permanently silent: one
branch always matches, so `go test` never prints the string and never exits non-zero, however
absent the other test is. Verbatim, at this HEAD:

```
--- PASS: TestCodexLocalInstructions_DualFileMatrix (0.00s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/cli	0.869s
```

Exit 0, `PASS`, no `no tests to run` — with the required test entirely absent. The criterion still
fails correctly, but only because a reader applies the prose *"must contain **both**"*. The
mechanical control the head block prescribes has been reduced to prose in exactly the criterion
that was just repaired for vacuity.

**Required fix:** run the two symbols as two commands with separate exit codes and separate
`no tests to run` reads — which is `plan.md` §C's own *"two mirrors, two commands"* discipline
applied to tests. One extra fenced line; no split of the criterion needed.

---

**D3-residual. `AC-IFU-031` asserts a docs-site build that no run in the named CI performs.** —
`acceptance.md:334-339` — Severity: **major** — Class: **blocking** — *carried from D3, one clause
to the left*

Evidence in the D3 section above: `grep -rn 'hugo' .github/workflows/` returns nothing,
`docs-i18n-check.yml` carries no build step and is path-filtered, and `ci.yml`'s `build` job builds
the Go binary. The head is now real; the check named on it is not.

**Required fix:** as stated in the D3 section.

---

## Gaps — what this audit did NOT observe

- Whether `ci.yml`'s required-check set (as configured in branch protection) matches what the
  criterion means by "every required check". I read the workflow file, not the protection rules.
- Whether Vercel builds the docs-site on a `develop` push. I established only that GitHub Actions
  does not.
- `.moai/reports/t1259/blocking-dependencies.md` (`a9e5f9d5a`) was treated as input, not judged. Its
  content bears on dispatch, not on whether this plan phase is sound, and nothing in it changes a
  verdict above.
- The full `internal/cli` package state. Only the one alternation command was run.
- Parent-SPEC M2 and t1175 landing states — unchanged from iter-1, still the lead's read.

## Residual risk

Both open items are in one criterion each and neither touches the design, which has now verified
cleanly twice. The risk that survives this iteration is narrow: if D3-residual is closed by simply
deleting the docs-site clause rather than re-siting it, M4 ships 24 locale files with no
integration-level check of any kind behind them — the Go suite says nothing about docs content, and
`docs-i18n-check`'s path filter means it fires only if M4's own commit touches
`docs-site/content/**` (it will, but that coupling should be stated rather than assumed).

## Recommendation

**PASS-WITH-DEBT at 0.92**, up from 0.85. Five of six iter-1 defects are fully closed and each was
re-verified mechanically rather than accepted; the sixth moved from unsatisfiable to
partially-grounded. The must-pass firewall is clean, traceability survived the repairs intact, and
the two judgment calls referred to this audit (equal-delta over narrowing; `design.md` §B unchanged)
were both decided correctly.

Two items remain, both small:

1. **D3-residual (`AC-IFU-031` docs-site build)** — before close, and preferably now. It is the
   second time this criterion's evidence has pointed somewhere that does not exist, which is the
   signal to ground the whole clause against the workflow files once rather than patch it again.
2. **N1 (`AC-IFU-029` positive control)** — before M2. One line.

Neither justifies a third iteration on its own. If both are repaired, iter-3 should be scoped to
those two lines rather than re-run.

🗿 MoAI

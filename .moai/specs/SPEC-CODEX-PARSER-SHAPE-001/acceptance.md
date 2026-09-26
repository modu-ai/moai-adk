# SPEC-CODEX-PARSER-SHAPE-001 — acceptance criteria

> Verification layer. Each criterion is a binary-testable Given-When-Then
> scenario. Measured statements trace to `.moai/reports/t1053/verdict.md` by
> section; no figure appears here that is not in that file.
>
> v0.2.0 (card t1203): criteria for the #1718 real case (§C.1, AC-CPS-011..014)
> and the REQ-CPS-010 decision (AC-CPS-015) trace to
> `.moai/reports/t1203/verdict.md` (tree `df526c9a9`), cited as `t1203 §N`. The
> same rule binds it.
>
> v0.2.1 (card t1203): repair of plan-audit iter-1 defects D1–D14
> (`.moai/reports/t1203/plan-audit.md`). The criterion count is unchanged (15).
> RED-now cells added or re-pinned in v0.2.1 were measured on tree `77b01ef4b`;
> at measurement time the working tree differed from that commit only in
> `spec.md` and `plan.md` of this SPEC, which none of those commands reads.
>
> v0.2.2 (card t1203): repair of plan-audit iter-2 defects
> (`.moai/reports/t1203/plan-audit-iter2.md`). AC-CPS-011's structural commands
> move out of the table into a verbatim evidence ledger, S1 gains a FAIL-statement
> check and a location-link check, and the fidelity check gets a named test
> selector with a non-empty-sweep condition. The criterion count is unchanged
> (15). The mutant table in AC-CPS-011 cites one further local record,
> `.moai/reports/t1203/repro/mutant-checks.log`, and no figure in it appears
> anywhere else.

## §A Gating criteria — satisfied BEFORE run-phase entry

These sit ahead of the Implementation Kickoff Approval gate (REQ-CPS-001). The
SPEC does not enter the run phase while either is unsatisfied.

**Both are SATISFIED.** The codex usage limit reset and the live call ran on
2026-09-21; the record carrying all four AC-CPS-002 items is
`.moai/reports/t1053/live-convention-20260921.md`. Result: **same-shape** — the
live body carries `- [P1] <message> — <path>:<line>` bullets and the parser
structured three of them with severities P1/P2/P2, read off that call's
`findings` array. Tree `a5c3f5dc6`, codex-cli 0.155.1 (the same version as the
blocked day, so version is not a variable). Two tolerated sub-shape differences —
a line RANGE whose end is discarded, and an absolute path — are recorded in
spec.md §A.4. The criteria below are retained as written because they govern any
re-measurement.

### AC-CPS-001 — live-convention comparison

**Given** the codex account usage limit has reset (2026-09-21 04:21) and live
codex calls succeed again,
**When** a codex review is invoked live through the moai MCP path against a real
target,
**Then** the returned review body is recorded verbatim, and the record states
explicitly whether that body's finding shape is one the current recognizers
accept (same-shape) or is not (different-shape).

[HARD] **What cannot satisfy this criterion.** It is NOT satisfied by:

- the eighteen offline measurements of verdict.md §E3 — those measure parser
  behaviour on bodies the measurer wrote, not codex output;
- the fixture `issue1632ReviewBody` or any other fixture in the tree;
- any test compiled from this tree;
- reading `internal/cli/mcp_codex.go`;
- an inference from the codex CLI version.

The criterion is satisfied only by an observation of **codex's own live output**,
taken after the reset. A record that cites a substitute measurement in place of
that observation leaves the criterion OPEN.

[HARD] **This criterion does not close verdict.md §6's first Gap.** That Gap
records that the convention was unmeasured on the measurement day; that remains
true permanently. AC-CPS-001 records a new measurement at a new point in time.

### AC-CPS-002 — the comparison is recorded, not summarized

**Given** the AC-CPS-001 observation has been taken,
**When** its result is written into the SPEC's evidence record,
**Then** the record carries all four of: the verbatim body, the exact invocation
that produced it, the tree it was measured against, and the codex CLI version —
so that a later reader can re-derive the same-shape / different-shape verdict
rather than taking it on trust.

A one-line "same as before" with no body recorded leaves AC-CPS-002 OPEN.

## §B Selection criterion

### AC-CPS-003 — candidates presented, not pre-selected

**Given** the plan-phase artifacts are complete,
**When** the Implementation Kickoff Approval gate is reached,
**Then** candidates (a), (b), (c), and (d) are all present with their measured
coverage, no one of them has been adopted as the implementation plan, and where
the plan orders them, it states that the ordering criterion is measured
failure-shape coverage (verdict.md §E3, §A2, §A5; t1203 §1) and nothing else,
names its unit (the failure shape, not the body instance), and applies the same
test — the candidate's remedy executed on the shape — to all four candidates, so
that the order follows from the stated counts without further judgement.

A plan that orders the candidates without naming the unit, or that counts a
shape for one candidate by a weaker test than for another, fails this criterion.

### AC-CPS-015 — the REQ-CPS-010 question is answered by the operator and recorded

**Given** the plan-phase artifacts present the §C.1 question — whether the #1718
adversarial outcome, `inconclusive` with an empty findings list for a body whose
prose states FAIL, is acceptable — with its two answers (keep REQ-CPS-010 as
written, or revise it) and with what each candidate changes on the adversarial
path,
**When** the Implementation Kickoff Approval gate is passed,
**Then** `progress.md` carries exactly one line of the form
`- REQ-CPS-010 decision: <keep|revise> — reason: <the operator's stated reason> — source: <where the Kickoff answer is recorded>`,
and the commit that introduced that line is a strict ancestor of the SPEC's first
run-phase commit — the commit that writes `status: in-progress` into `spec.md`.

A plan-side or agent-side answer — a line written without the operator's answer
behind it, or a run-phase commit that precedes the line — fails this criterion.
Neither answer excludes a candidate (spec.md §C.1).

**Mechanical checks** (paths relative to the repository root; `P` =
`.moai/specs/SPEC-CODEX-PARSER-SHAPE-001/progress.md`, `S` =
`.moai/specs/SPEC-CODEX-PARSER-SHAPE-001/spec.md`):

1. Format — `grep -cE '^- REQ-CPS-010 decision: (keep|revise) — reason: .+ — source: .+$' P`
   prints `1`, exit `0`.
2. Introducing commit — `git log --format=%H -S'- REQ-CPS-010 decision:' -- P`;
   the last (oldest) line of its output is `D`.
3. First run-phase commit — `git log --format=%H -G'^status: in-progress$' -- S`;
   the last (oldest) line of its output is `R`.
4. Order — `git merge-base --is-ancestor D R` exits `0`, and `D` ≠ `R`.

**What no check decides.** Whether the line reflects the operator's own answer
is not mechanically decidable: no command separates a line typed on the
operator's answer from one an agent wrote without it. The `source` field makes
the claim reviewable — a reader can open the named record and compare — not
decidable. The run-phase record states this rather than reporting the format
check as proof of authorship.

- **RED-now** (tree `77b01ef4b`; `progress.md` unmodified from that commit at
  measurement time).
  - Check 1: stdout `0`; exit `1`.
  - Check 2: stdout empty; exit `0`. The empty output is absence, not a blind
    probe: the same form with `-S'2026-09-26 · v0.2.0 amendment'` on the same
    path printed `686b75ebb`, exit `0`.
  - Red for the stated reason: no answer has been given; §E.1 records the
    question as pending. The v0.2.1 commit adds no decision line, so the RED
    holds on it too.
- **Green path.** M1 (Kickoff) — the operator's answer is recorded; check 1 then
  prints `1` with exit `0`, and checks 2–4 are run at the first run-phase commit
  and their outputs recorded in `progress.md` §E.2.

## §C Behavioural criteria — conditional on the selected candidate

Only the criteria matching the operator's selection apply; the others are
recorded as not-applicable with the selection as the reason.

### AC-CPS-004 — candidate (b): native disambiguation before downgrade

**Given** candidate (b) is selected,
**When** the native path receives a body carrying no recognized signal,
**Then** the implementation distinguishes "codex found nothing to block on" (the
V9 shape) from "the shape was not recognized" (the V2–V6 shapes) before any
downgrade from `pass`, and a V9-shaped body still yields `pass`.

A change that downgrades both cases identically fails this criterion — that is
the byte-level indistinguishability recorded in verdict.md §E4, not a
verification detail.

### AC-CPS-005 — candidate (c): verdict/findings contradiction is reported, and the unmet-gate state is NOT

**Given** candidate (c) is selected,
**When** a synthesized review output would carry a blocking verdict together
with an empty findings list **and an empty `GateUnmet`** — the V8 shape,
measured as `fail` / `findings=0` in both modes (verdict.md §E3),
**Then** that state is reported as self-contradictory rather than emitted as a
clean review, and a consumer reading the output can tell the content was lost.

**Control case (mandatory, REQ-CPS-006a).**
**Given** candidate (c) is selected,
**When** a review output carries `Verdict == "fail"` with an empty findings list
**and a non-empty `GateUnmet`** — the state `applyGateUnmet`
(`internal/cli/mcp_codex.go`) produces when `workflow.audit.gates.codex` is
`required` and the audit returned a fail-open `inconclusive` with
`Findings: []Finding{}`,
**Then** that state is **NOT** reported as self-contradictory.

The control is mandatory rather than advisory because the bare two-term
predicate (`verdict == fail && len(findings) == 0`) matches this legitimate
state exactly: an unmet required gate is a correctly functioning gate, and
flagging it as a parser contradiction would convert a true signal into a false
defect report. Without a test pinning it, an implementation can regress into
that behaviour and every V8 test still passes.

### AC-CPS-006 — candidate (a): widening parses more without changing what parsed

**Given** candidate (a) is selected,
**When** a review body carries findings in a shape the widened recognizers
accept,
**Then** the parsed findings match the fixture's **exact expected count AND
content** — severity, message, file, and line for every finding — **and** the
findings parsed from the V1 shape (`- [P1] …`, measured as `fail` / 2 findings
in both modes — verdict.md §E3) are unchanged in count and content.

**Partial-drift case (mandatory).**
**Given** candidate (a) is selected and a widened fixture carrying N findings,
**When** only some of those findings are rewritten into a shape the widened
recognizers do NOT accept,
**Then** the exact-count assertion fails — it does not pass on the surviving
subset.

[HARD] **This is §A.5's lesson applied to a new criterion, and it was originally
written the wrong way.** The first draft of this AC required only that "that
finding is parsed" — singular, no count. A two-finding body that parses one and
silently drops the other satisfied it, which is precisely the partial-drift
failure spec.md §A.5 records: a non-empty-style assertion passes while content is
lost, and only an exact count fires. The V1 unchanged-count clause does not cover
it, because V1 is the OLD shape and the drift happens in the NEW one. Every AC
this SPEC adds for a parsed-findings shape states an exact count; an author
tempted to write "at least one is parsed" is re-creating the defect this SPEC
exists to close.

## §C.1 #1718 real-case criteria — conditional on a candidate being claimed for #1718

Only the criteria matching the operator's selection apply; the others are
recorded as not-applicable with the selection as the reason. None of them may be
satisfied by the raw #1718 bodies: those carry another project's absolute paths
and content and are never committed (REQ-CPS-014).

**The shared target observation (evidence ledger E-1718).** AC-CPS-012 and
AC-CPS-013 cite this entry as their RED-now cell until AC-CPS-011 closes;
AC-CPS-011 cites it as the output its reductions must reproduce.

```
id:       E-1718
command:  go test -count=1 -v -run '^TestT1203Probe$' ./internal/cli/
tree:     df526c9a9
exit:     0
stdout (verbatim and complete — t1203 repro/probe.log):
=== RUN   TestT1203Probe
PROBE body1 turn/start verdict=inconclusive findings=0 note="" findingsJSON=[]
PROBE body1 review/start verdict=pass findings=0 note="" findingsJSON=[]
PROBE body2 turn/start verdict=inconclusive findings=0 note="" findingsJSON=[]
PROBE body2 review/start verdict=pass findings=0 note="" findingsJSON=[]
PROBE ctrlA turn/start verdict=fail findings=1 note="" findingsJSON=[{"severity":"P1","title":"broken check — a/b.go:3","body":"broken check — a/b.go:3","file":"a/b.go","line":3,"confidence":0,"recommendation":""}]
PROBE ctrlA review/start verdict=fail findings=1 note="" findingsJSON=[{"severity":"P1","title":"broken check — a/b.go:3","body":"broken check — a/b.go:3","file":"a/b.go","line":3,"confidence":0,"recommendation":""}]
PROBE ctrlB turn/start verdict=fail findings=0 note="" findingsJSON=[]
PROBE ctrlB review/start verdict=fail findings=0 note="" findingsJSON=[]
--- PASS: TestT1203Probe (0.00s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/cli	0.756s
```

[HARD] **Disposition of E-1718.** The probe that produced it was a temporary test
file deleted after measurement, reading bodies from a session scratch path
(t1203 §1). It therefore cannot be re-executed on the current tree as written,
so per verification-completeness §2.1 no criterion that rests on it as its RED
is release-blocking, and none is recorded as a pass on the strength of E-1718.

**Guard classification, before and after AC-CPS-011 closes.** AC-CPS-011 exists
to end the undecidable state; its closure is the transition point.

| Criterion | Before AC-CPS-011 closes | After AC-CPS-011 closes |
|---|---|---|
| AC-CPS-011 | **release-blocking** where #1718 is claimed — its own RED-now (fixtures absent) is a single-invocation command re-executable on the current tree | closed |
| AC-CPS-012 | regression-guard — RED-now is E-1718, not re-executable | **release-blocking** — its RED-now becomes AC-CPS-011's recorded closure observation (command, verbatim stdout, exit code, tree SHA) on the pre-change tree |
| AC-CPS-013 | regression-guard — same reason | **release-blocking** — same replacement |
| AC-CPS-014 | regression-guard — RED-now is the population observation, not re-executable | **unchanged: regression-guard.** Its RED cannot be re-executed and AC-CPS-011's fixtures do not replace it, because (d) acts on live codex output, not on fixtures. It is recorded with its live observation, never used as a release gate, and never recorded as a pass on the population figures |

### AC-CPS-011 — the sanitized reductions reproduce the measured raw output

**Given** a candidate is claimed to address #1718,
**When** five fixture files are committed under `internal/cli/testdata/codex-1718/`
— the address is fixed here so every check below has a target; moving it
requires amending this criterion —
- `S1.txt` — sanitized reduction of body1: a greeting opens the verdict line, the
  verdict label is localized (`판정`), findings are a markdown table with bold
  severity words;
- `S2.txt` — sanitized reduction of body2: a greeting is followed by a bold
  `verdict: fail`, findings are bold severity-word bullets carrying a
  `[path:line](<…>)` link;
- `S2p.txt` — S2′, the ctrlB analogue: `S2.txt` with the greeting prefix of line 1
  (the text up to and including the first `, `) removed and nothing else changed;
- `N1.txt`, `N2.txt` — negative fixtures for AC-CPS-012: prose that mentions a
  verdict and uses `fail` or `pass` as an ordinary word on the same line without
  stating a verdict — in N1 the mention opens the line, in N2 it follows a
  greeting mid-line,

**Then** all of the following hold, each checked by the single-invocation
command given and recorded verbatim (command, stdout, exit code) with the tree
SHA:

1. **Fidelity.** On the parser **before** any candidate change, the fixture test
   synthesizes each file on turn/start and review/start: S1 and S2 each yield
   `inconclusive` with `findings=0` on turn/start and `pass` with `findings=0` on
   review/start, and S2′ yields `fail` with `findings=0` on both paths — the
   outputs E-1718 recorded for body1, body2, and ctrlB. N1 and N2 are
   synthesized and their outputs recorded, with no expected value: they are the
   pre-change baseline AC-CPS-012 compares against. Command, run from the
   repository root:

   ```
   go test -count=1 -v -run '^TestCodex1718Fixtures$' ./internal/cli/
   ```

   Pass condition: exit `0`; stdout carries the line
   `--- PASS: TestCodex1718Fixtures`; stdout does **not** carry
   `[no tests to run]`; and stdout carries ten synthesis lines, one per fixture
   file per path (five files × turn/start and review/start), each naming the
   file, the path, the verdict, and the findings count. A run missing any of
   these is an empty or partial sweep and does not close this check — a selector
   that matches no test also exits `0` and prints `ok`.
2. **Structural properties** — so that a reduction with the right outputs but
   none of the #1718 shape cannot pass. The table states each property and its
   pass condition; the commands themselves are recorded only in the evidence
   ledger below it, verbatim, one command per line, run from the repository
   root. A reader copies the ledger line, never a table cell. In the ledger every
   command is exactly what the shell receives: inside the single-quoted regular
   expressions, `\|` is a literal pipe character and a bare `|` is alternation.

   | # | Property | Files | Ledger | Pass condition |
   |---|---|---|---|---|
   | P1 | a greeting precedes the verdict on line 1 | S1, S2 | P1-S1, P1-S2 | stdout's first line begins `1:`; exit `0` |
   | P2 | no verdict label opens any line | S1, S2 | P2-S1, P2-S2 | stdout `0`; exit `1` |
   | P3 | the label is localized, not English | S1 | P3a, P3b | P3a ≥ `1`, exit `0`; P3b `0`, exit `1` |
   | P4 | findings are table rows with a bold severity word | S1 | P4 | equals S1's declared finding count, and that count is ≥ `2` |
   | P5 | findings are bold severity-word bullets with a `[path:line](<…>)` link | S2 | P5 | equals S2's declared finding count, and that count is ≥ `1` |
   | P6 | no bracketed-severity bullet anywhere | all five | P6-S1 … P6-N2 | stdout `0`; exit `1` for each |
   | P7 | S2′ differs from S2 only on line 1 | S2, S2p | P7 | the only hunk header is `1c1`; exit `1`; and the fixture test asserts line 1 of S2p equals line 1 of S2 with its greeting prefix removed |
   | P8 | negatives mention a verdict and an ordinary `fail`/`pass` on one line, state none | N1, N2 | P8a-N1, P8a-N2, P8b-N1, P8b-N2 | each P8a ≥ `1`, exit `0`; each P8b `0`, exit `1` |
   | P9 | placement of the mention | N1 / N2 | P9-N1, P9-N2 | each ≥ `1`, exit `0` |
   | P10 | the S1 prose states FAIL under the localized label | S1 | P10 | ≥ `1`; exit `0` |
   | P11 | every S1 finding row carries a `[path:line](…)` location link | S1 | P11 | equals S1's declared finding count — the same count P4 must equal |

   Evidence ledger (AC-CPS-011 check 2):

   ```
   # P1-S1
   grep -nE '^[^[:space:]*#|>-][^,]*, .*(판정|[Vv]erdict)' internal/cli/testdata/codex-1718/S1.txt
   # P1-S2
   grep -nE '^[^[:space:]*#|>-][^,]*, .*(판정|[Vv]erdict)' internal/cli/testdata/codex-1718/S2.txt
   # P2-S1
   grep -cE '^[[:space:]]*(\*\*)?([Vv]erdict|판정)' internal/cli/testdata/codex-1718/S1.txt
   # P2-S2
   grep -cE '^[[:space:]]*(\*\*)?([Vv]erdict|판정)' internal/cli/testdata/codex-1718/S2.txt
   # P3a
   grep -c '판정' internal/cli/testdata/codex-1718/S1.txt
   # P3b
   grep -ci 'verdict' internal/cli/testdata/codex-1718/S1.txt
   # P4
   grep -cE '^\|[^|]*\*\*(Critical|High|Medium|Low)\*\*[^|]*\|' internal/cli/testdata/codex-1718/S1.txt
   # P5
   grep -cE '^- \*\*(Critical|High|Medium|Low) · \[[^]]+:[0-9]+\]\(<[^>]+>\)' internal/cli/testdata/codex-1718/S2.txt
   # P6-S1
   grep -cE '^- \[P[0-9]\]' internal/cli/testdata/codex-1718/S1.txt
   # P6-S2
   grep -cE '^- \[P[0-9]\]' internal/cli/testdata/codex-1718/S2.txt
   # P6-S2p
   grep -cE '^- \[P[0-9]\]' internal/cli/testdata/codex-1718/S2p.txt
   # P6-N1
   grep -cE '^- \[P[0-9]\]' internal/cli/testdata/codex-1718/N1.txt
   # P6-N2
   grep -cE '^- \[P[0-9]\]' internal/cli/testdata/codex-1718/N2.txt
   # P7
   diff internal/cli/testdata/codex-1718/S2.txt internal/cli/testdata/codex-1718/S2p.txt
   # P8a-N1
   grep -cE '[Vv]erdict[^:]* (fail|pass)' internal/cli/testdata/codex-1718/N1.txt
   # P8a-N2
   grep -cE '[Vv]erdict[^:]* (fail|pass)' internal/cli/testdata/codex-1718/N2.txt
   # P8b-N1
   grep -cE '[Vv]erdict[[:space:]]*[:：]|판정' internal/cli/testdata/codex-1718/N1.txt
   # P8b-N2
   grep -cE '[Vv]erdict[[:space:]]*[:：]|판정' internal/cli/testdata/codex-1718/N2.txt
   # P9-N1
   grep -cE '^[Vv]erdict' internal/cli/testdata/codex-1718/N1.txt
   # P9-N2
   grep -cE '^[^[:space:]][^,]*, .*[Vv]erdict' internal/cli/testdata/codex-1718/N2.txt
   # P10
   grep -cE '판정[^|]*\*\*FAIL\*\*' internal/cli/testdata/codex-1718/S1.txt
   # P11
   grep -cE '^\|[^|]*\*\*(Critical|High|Medium|Low)\*\*[^|]*\|[^|]*\[[^]]+:[0-9]+\]\(' internal/cli/testdata/codex-1718/S1.txt
   ```

   A declared finding count is the count the fixture test asserts for that file;
   P4, P5, and P11 tie the structure to that assertion, so a one-line prose
   fixture fails them. In P1 the pipe sits inside a bracket expression, where it
   is an ordinary character either way; in P4 and P11 the escaped `\|` anchors a
   literal table pipe — written bare, `^|` and the trailing `|` become empty
   alternatives and the pattern matches every line (plan-audit iter-2 N4-P4).

   **Which check kills which mutant.** Each mutant below satisfies the checks
   not named in its row; the named checks are what reject it. Executed on
   2026-09-26 against sanitized scratch copies outside the tree (tree
   `51a41e187`), grep checks only — the parser fidelity check was **not**
   executed on these copies; command lines, stdout, and exit codes are recorded
   in t1203 `repro/mutant-checks.log` (gitignored, as the rest of `repro/`).
   Declared finding count taken as `3` for the faithful copy and `2` for M2.

   | Mutant | Shape | P4 | P10 | P11 | Killed by |
   |---|---|---|---|---|---|
   | faithful S1 (sanitized body1; control) | greeting + `판정은 **FAIL**` + three linked bold-severity rows | `3` / exit 0 | `1` / exit 0 | `3` / exit 0 | none — passes, as it must |
   | M1 | two-line prose; no table, no FAIL, no link | `0` / exit 1 | `0` / exit 1 | `0` / exit 1 | P4, P10, P11 |
   | M2 | two bold-severity table rows; no FAIL statement, no location link | `2` / exit 0 | `0` / exit 1 | `0` / exit 1 | P10, P11 (P4 alone passes it) |
   | one-line BAD | one sentence stating FAIL; no table | `0` / exit 1 | `1` / exit 0 | `0` / exit 1 | P4, P11 |

   The same log records the v0.2.1 transcription of P4 (bare pipes) on the same
   copies: `12`, `2`, `6`, `1` — it counts lines, not rows, which is the N4-P4
   defect this ledger removes.
3. **Sanitization** — the fixtures carry no absolute user path and no content of
   the originating project. Command:
   `grep -rniE '(/Users/|/home/|/private/|/var/folders/|[A-Za-z]:\\Users|cowork|SKILL\.md|test-cases\.yaml|oai-mem-citation|구스|오뽜|영실)' internal/cli/testdata/codex-1718/`
   — stdout empty; exit `1`. The forbidden list is: absolute user-home and
   temporary path prefixes (`/Users/`, `/home/`, `/private/`, `/var/folders/`,
   a drive-letter `\Users`); the originating project's name (`cowork`); the
   originating files the #1718 findings cite (`SKILL.md`, `test-cases.yaml`); the
   codex memory-citation tag the raw bodies carry (`oai-mem-citation`); and the
   persona tokens quoted in t1203 §2. The check must also be observed firing:
   the same command run against a scratch copy of the directory with one
   forbidden token inserted prints the matching line, exit `0` — recorded
   alongside, so an empty result is shown to be absence and not a blind pattern.

A reduction that does not reproduce its raw body's measured output is not a
reduction of that shape, and no candidate claim may rest on it.

- **RED-now.** Command `ls internal/cli/testdata/codex-1718`; stdout empty
  (stderr `ls: internal/cli/testdata/codex-1718: No such file or directory`);
  exit `1`; tree `77b01ef4b`. Red for the stated reason: no fixture has been
  committed, so none of checks 1–3 has a target. This observation is of this
  criterion, not of the target values it must reproduce (those are E-1718).
- **Green path.** M2 first step, before any recognizer or prompt change: the five
  fixtures and the fixture test are committed; checks 1–3 are run on that commit
  and recorded; this AC closes when check 1 matches E-1718 row for row and
  checks 2–3 meet their pass conditions. That recorded run becomes the RED-now of
  AC-CPS-012 and AC-CPS-013 (§C.1 guard classification).

### AC-CPS-012 — candidate (a) on the #1718 shapes: exact verdict, count, and content

**Given** candidate (a) is selected for the #1718 shapes and AC-CPS-011 holds,
**When** S1, S2, and S2′ are synthesized on turn/start and on review/start,
**Then** each yields verdict `fail` and a findings list whose **exact count and
content** — severity, message, file, and line for every finding — match the
fixture's declared expectation; the V1 and V9 shapes are unchanged in count and
content (AC-CPS-006); and a prose sentence that mentions a verdict without
stating one is still not read as a verdict (the narrowness contract in
`codexStatedVerdict`'s comment).

**Negative case (mandatory).** The negative fixtures N1 (mention at line head)
and N2 (mention after a greeting, mid-line) of AC-CPS-011 — each carrying the word
`verdict` and an ordinary `fail` or `pass` on one line, with no verdict stated,
as AC-CPS-011's P8 and P9 check — yield, after the (a) change, exactly the
verdict and findings count they yielded before it (AC-CPS-011 check 1). A
widening that reads either as a stated verdict fails this criterion. The two
placements are both required because an anchor relaxed to admit a greeting
prefix is exactly what would newly accept N2.

**Partial-drift case (mandatory).** When one finding of S1 or S2 is rewritten
into a shape the widened recognizers do not accept, the exact-count assertion
fails; it does not pass on the surviving subset (§A.5, AC-CPS-006).

- **RED-now.** Until AC-CPS-011 closes: E-1718 — the raw counterparts of S1/S2
  yield `inconclusive`/0 and `pass`/0, that of S2′ yields `fail`/0 — red against
  a `fail`/N expectation because neither the greeting-prefixed verdict nor the
  table / bold-severity findings are recognized (t1203 §2); regression-guard.
  From AC-CPS-011's closure: that closure observation on the pre-change tree
  (§C.1 guard classification); release-blocking.
- **Green path.** M2 under (a): the three recognizer changes of spec.md §C (#1718
  table) turn the fixture test green; each is measured, none inherited from the
  numbered-list result (plan.md §F M1).

### AC-CPS-013 — candidate (c) on the #1718 shapes: what it catches, and what it does not

**Given** candidate (c) is selected and AC-CPS-011 holds,
**When** S1, S2, and S2′ are synthesized with an empty `GateUnmet`,
**Then** S2′ is reported as self-contradictory on both paths (its output is
`fail` with `findings=0`); S1 and S2 are **not** reported as self-contradictory,
because no blocking verdict survives on them; and the run-phase record states that
(c) alone leaves the raw #1718 shapes at `inconclusive`/0 (turn/start) and
`pass`/0 (review/start).

The second half is load-bearing: it keeps (c)'s ctrlB coverage from being
reported as #1718 coverage (plan.md §G anti-pattern 6).

- **RED-now.** Until AC-CPS-011 closes: E-1718 — ctrlB yields `fail`/0 on both
  paths with nothing marking it contradictory — red because (c) does not exist
  yet; regression-guard. From AC-CPS-011's closure: the S2′ row of that closure
  observation (§C.1 guard classification); release-blocking.
- **Green path.** M2 under (c): the contradiction report appears on S2′ and only
  there; the AC-CPS-005 control case still holds.

### AC-CPS-014 — candidate (d): the pinned format is observed in live output

**Given** candidate (d) is selected and the AC-CPS-015 decision line exists,
**When** a live adversarial codex review is invoked through the moai MCP path
against a target known to produce findings, under project instructions that
previously produced an unrecognized shape,
**Then** the returned body is recorded verbatim together with the invocation, the
tree, and the codex CLI version, and its synthesized output carries a verdict
other than the unrecognized-body fall-through and a findings list whose exact
count matches the findings stated in that body.

**Where the record lives.** The verbatim body goes to
`.moai/reports/<run-card-id>/ac-cps-014-live-body.md` — a gitignored path (the
repository ignores `.moai/reports/*`), because a body produced under another
project's instructions carries that project's content (spec.md §E). The committed
record — `progress.md` §E.2 — carries the invocation, the tree SHA, the codex CLI
version, the sha256 of the verbatim body file, the synthesized verdict and
findings count, and the path above; it never carries the body text.

[HARD] **What cannot satisfy this criterion** — the same exclusions as
AC-CPS-001: a fixture, any test compiled from the tree, reading the prompt text,
or an inference from the codex CLI version. (d) acts on what codex emits, so only
codex's own live output can show it worked.

- **RED-now.** The population observation (t1203 §3): of 142 moai-cowork
  codex-gate final bodies, turn/start gave `inconclusive`/0 for 138, and no run
  produced a parsed finding (0 of 284). Recorded in t1203 `repro/pop.log`
  (tree `df526c9a9`, exit 0); same disposition as E-1718 — the probe is not in
  the tree — so this criterion is regression-guard before and after AC-CPS-011
  closes (§C.1 guard classification). Which path each session took, and whether
  each body is byte-identical to the `reviewText` the parser received, are Gaps
  (spec.md §A.6).
- **Green path.** M2 under (d), then one live call recorded as above.

## §D Preserved-behaviour criteria — apply whichever candidate is selected

### AC-CPS-007 — exact-count assertions stay exact

**Given** the run phase has completed,
**When** `internal/cli/codex_findings_parse_test.go` is inspected,
**Then** every findings-count assertion that was an exact equality before the
change is still an exact equality — none has been relaxed to an inequality such
as `>= 1`.

Basis: under partial shape drift (one bullet of four changed), the exact-count
assertions are what fire, while a "non-empty" style control passes because three
findings remain (verdict.md §4, §A3). The exact count is the measured defence;
loosening it erases the defence silently.

### AC-CPS-008 — a genuinely clean native review is not made loud

**Given** any candidate has been implemented,
**When** the native path receives a body carrying no findings because codex found
nothing to block on,
**Then** the emitted verdict is `pass` — not `inconclusive`.

### AC-CPS-009 — the adversarial path is unchanged

**Given** any candidate has been implemented,
**When** the adversarial path receives a body matching no recognized signal,
**Then** it returns `inconclusive`, as it did before the change (verdict.md §E3)
— the handling of a body that no recognizer accepts is unchanged. (A candidate
may change which bodies are recognized; that is not a change to this handling —
spec.md §C.1.)

**Scope (v0.2.0; re-stated v0.2.1).** This criterion applies as written when the
operator's AC-CPS-015 answer is `keep`. If the answer is `revise`, this criterion
is re-authored at that decision together with the revised requirement; it is not
pre-written here, and it is not silently dropped. Neither answer excludes a
candidate.

### AC-CPS-010 — the closed axis stays closed

**Given** the run phase has completed,
**When** the diff is inspected,
**Then** no change touches the codex synthesis `next_steps` field or any
behaviour card t1052 closed.

## §D.1 Traceability (REQ → AC)

| REQ | AC |
|---|---|
| REQ-CPS-001 (no run-phase entry while unmeasured) | AC-CPS-001 |
| REQ-CPS-002 (comparison recorded verbatim) | AC-CPS-002 |
| REQ-CPS-003 (candidates presented, not chosen) | AC-CPS-003 |
| REQ-CPS-004 (ordering criterion is coverage) | AC-CPS-003 |
| REQ-CPS-005 (candidate (b) disambiguation) | AC-CPS-004 |
| REQ-CPS-006 (candidate (c) contradiction, `GateUnmet == ""` scoped) | AC-CPS-005 |
| REQ-CPS-006a (unmet gate is NOT a contradiction) | AC-CPS-005 control case |
| REQ-CPS-007 (candidate (a) widening) | AC-CPS-006 |
| REQ-CPS-008 (exact counts stay exact) | AC-CPS-007 |
| REQ-CPS-009 (clean native review stays `pass`) | AC-CPS-008 |
| REQ-CPS-010 (adversarial unchanged) | AC-CPS-009 |
| REQ-CPS-011 (`next_steps` closed) | AC-CPS-010 |
| REQ-CPS-012 (candidate (d) format pin, live-observed) | AC-CPS-014 |
| REQ-CPS-013 (the §C.1 question — is the #1718 outcome acceptable — is the operator's) | AC-CPS-015 |
| REQ-CPS-014 (#1718 claims rest on sanitized, fidelity-checked reductions) | AC-CPS-011 (fidelity, structural properties, sanitization check), AC-CPS-012, AC-CPS-013 |

## §E Edge cases

- **Live call succeeds but returns a body with no findings at all.**
  [HARD] **AC-CPS-001 stays OPEN**, and the observation is retried against a
  target known to produce findings. A body carrying no finding shape does not
  establish that the finding shape is unchanged — it establishes nothing about
  the finding shape at all, which is the exact state the criterion exists to
  detect. The record states that no finding shape was observable and names that
  as the limit of what the observation establishes; it is NOT recorded as
  "same shape", and — the load-bearing half — **it does not close the
  criterion.** Allowing the gate to close here would let the SPEC enter the run
  phase with the convention still unmeasured.
- **Live call is still blocked after the reset.** AC-CPS-001 stays OPEN and the
  SPEC does not enter the run phase. The blockage is recorded with the same
  two-path control used on the measurement day — through moai and through a
  direct `codex exec` call — so that "codex account" and "moai path" are
  distinguishable (verdict.md §E1).
- **The observed live shape is outside the nine measured variants.** The record
  says so. verdict.md §6 already states that the nine variants were enumerated by
  the measurer rather than derived from codex's behaviour; an unenumerated shape
  is expected, not anomalous.
- **A #1718 reduction reproduces only one of the two paths.** AC-CPS-011 stays
  OPEN for that reduction. Its raw body was measured on both paths (t1203 §1),
  and a half-faithful reduction cannot stand in for it.
- **A reduction reproduces both outputs but fails a structural property check.**
  AC-CPS-011 stays OPEN. Matching outputs is necessary, not sufficient: the
  checks exist because a one-line prose body with no recognized signal falls
  through to body1's outputs as well — deduced from the spec.md §A.2
  fall-through, not measured (plan-audit D4).
- **The sanitization command prints nothing and its firing control was not run.**
  AC-CPS-011 stays OPEN: an empty result without the control is unmeasured, not
  clean.
- **The population is re-measured and the distribution differs.** Record the new
  figures with their tree; do not overwrite the t1203 figures, which remain the
  observation of their day.

## §F Quality gates

- Package tests for `internal/cli` pass, compiled from the tree — not from the
  installed binary (`f67d2193f`, older than this tree; verdict.md §6).
- `go vet ./internal/cli/...` clean.
- The run-phase evidence record names the command it ran and the output it
  observed for every criterion it marks satisfied.

## §G Definition of Done

- [ ] AC-CPS-001 and AC-CPS-002 satisfied and recorded, BEFORE Kickoff approval
- [ ] AC-CPS-003 satisfied — candidates presented, not pre-selected
- [ ] AC-CPS-015 satisfied — the REQ-CPS-010 answer recorded, and its
      introducing commit an ancestor of the first run-phase commit (checks 1–4);
      authorship recorded as reviewable, not as mechanically proven
- [ ] Selected candidate's criterion (AC-CPS-004 / 005 / 006) satisfied
- [ ] Where #1718 is claimed: AC-CPS-011 satisfied (release-blocking), and then
      the matching AC-CPS-012 / 013 satisfied as release-blocking criteria on the
      RED-now AC-CPS-011's closure supplied (§C.1 guard classification)
- [ ] Where (d) is selected: AC-CPS-014's live observation recorded at the
      paths it names — as a regression-guard, not a release gate, and not
      recorded as a pass on the population figures
- [ ] The unselected criteria recorded as not-applicable with the selection as
      the reason
- [ ] AC-CPS-007 through AC-CPS-010 satisfied
- [ ] Every claim in the run-phase evidence traces to a command run in this tree
- [ ] Windows path behaviour and live `audit_multi` appear, if at all, as
      explicitly-unmeasured notes — never as claims

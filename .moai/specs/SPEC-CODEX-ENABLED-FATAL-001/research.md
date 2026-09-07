# SPEC-CODEX-ENABLED-FATAL-001 — Research

Derived from three evidence files already committed to this tree. Nothing here is re-derived; where
a figure appears, it is carried with its attribution to the run that measured it.

- `.moai/reports/t508/codex-enabled-lab.md` — the six-row acceptance lab
- `.moai/reports/t508/red-baseline.md` — the RED baseline and the structural read
- `.moai/reports/t508/develop-t506-impact.md` — the t506 landing, measured against develop

**Tree pin.** Structural claims below were re-read from source in this tree at commit
**`069795602`** — the HEAD after the branch absorbed develop. The pre-absorb draft's citations had
drifted; §9 records the restamp. The codex-behaviour claims were NOT re-measured here — they are
carried from the lab, which measured them in its own run.

---

## 1. What codex actually requires

Measured by the lab: codex-cli 0.153.4 at `/Users/goos/.local/bin/codex`, isolated `CODEX_HOME`
under `/tmp/t508lab`, probe `codex mcp list`.

The probe choice is load-bearing and is recorded because it is easy to get wrong: `codex --version`
does NOT load the config and returns rc=0 on every fixture. A version probe would have measured
nothing and reported agreement.

`enabled` must be a **bare TOML boolean**. Four distinct shapes are fatal, and codex distinguishes
two error classes:

- absent key → ``missing field `enabled` in `skills.config` ``
- non-boolean value → ``invalid type: <type> …, expected a boolean in `skills.config.enabled` ``

`path`, by contrast, is OPTIONAL (lab row 6, rc=0 with no `path` line). This asymmetry is the scope
boundary of the whole card: without row 6 the natural generalisation would be "entry keys are
required", which is false.

**The induction, named.** Three non-boolean shapes were probed: integer, double-quoted string,
single-quoted string. REQ-CEF-004 generalises to "not a bare TOML boolean". The uniform error text
is the ground for that generalisation, and it remains an induction — a float, an array, an inline
table, and a bareword such as `yes` were not probed (acceptance.md §D.4).

## 2. Where moai's reading diverges

`internal/codexwiring/skills.go` models `enabled` as a tri-state (`SkillEnabledUnspecified` /
`True` / `False`, line 34). Two divergence classes result, and they are not equally bad:

**Silent (rows 2, 3).** Absent key and integer value both land on `Unspecified`. The doc comment at
`skills.go:34` says this outright, and the sentence is the defect in miniature:

> `SkillEnabledUnspecified` is an entry declaring no `enabled` key, **or one whose value this parser
> does not recognise**. It asserts nothing.

The emphasised clause is why the integer case is exactly as silent as the absent case: one state
carries both "nothing was said" and "something was said that codex rejects". REQ-CEF-001 exists to
separate them. (The previous draft quoted this sentence with the clause elided behind an ellipsis,
which removed the part that makes the case.)

**Actively wrong (rows 4, 5).** Quoted values are read at face value by `skillEnabledKeyRe`
(`skills.go:131`; three alternation branches — bare, `"..."`, `'...'`). `enabled = "true"` therefore
reads `SkillEnabledTrue` — a confident, healthy reading of a config codex refuses to load.

The second class is worse in a way that matters for prioritisation: a silent reading leaves the user
without information, an actively-wrong reading gives them false information they may act on.

## 3. The leniency that has to be reversed, and why it was there

The quoted-value acceptance is not an oversight. It is commented, and the comment argues for it.
Quoted in FULL from `skills.go:123-127` — the previous draft dropped the final sentence with no
ellipsis, and that sentence is one of the comment's two grounds:

> The enabled matcher accepts a quoted value as well as a bare one. A quoted `"true"` is a TOML
> string rather than a boolean, so it is arguably malformed; reading it as false, however, silently
> DEMOTES a live registration to stale bookkeeping, which is the more damaging misreading. **The
> declared intent is unambiguous, so it is taken at face value and reported as declared.**

It is also test-pinned in two places: `skills_test.go:69` `TestParseSkillEntriesEnabledQuoted` (a
five-row table), and `doctor_codex_test.go:302`, which is a **live assertion**, not merely the
fixture comment at `:297` the first draft cited.

**Two grounds, two answers.** The first draft answered only the first and described the result as
"a decision whose stated ground turned out to be false" — singular. There are two:

- **Ground 1 — demotion of a LIVE registration.** *Falsified by measurement.* Lab row 4 establishes
  there is no live registration on that input: codex exits 1 before loading anything, so the
  demotion cannot occur. The real choice is between silence and a fatal report.
- **Ground 2 — the declared intent is unambiguous, so take it at face value.** *True, and beside the
  point.* **Codex never reads the intent.** It reads the bytes, finds a string where a boolean is
  required, and refuses to start. Reporting a registration codex will not load is
  confident-and-wrong regardless of what the author meant. Face-value reporting is right only where
  the intent is what the consuming tool acts on; here it is not.

This is worth stating at length in the SPEC (§B.1 there) because the reversal will otherwise read as
someone who did not notice the comment. The distinguishing fact is that one ground was measured
false and the other does not reach the consuming tool — not that a later author preferred a
different trade-off. REQ-CEF-013 requires both answers to land in the code.

## 4. The severity axis does not exist yet

Re-read in this session at `069795602`:

- `internal/cli/doctor_codex.go:114` — `codexFinding` is `{summary, detail}`. No severity field.
- `internal/cli/doctor_codex.go:283` — `check.Status = uikit.CheckWarn` for any non-empty `problems`
  slice. Every problem, whatever it is, lands on warn.
- `internal/cli/uikit/types.go:17` — `CheckFail` exists as a status value.
- `internal/cli/doctor.go:142` `doctorExitStatus` — `failCount == 0` returns nil; otherwise exit 1,
  with the comment recording the intent: a run that printed `Fail N` must not exit 0, and warn-only
  runs stay 0 deliberately.

So the fix cannot select a severity — there is nothing to select from. `codexFinding` must gain a
severity, and the status computation must become a fold over per-finding severities rather than a
constant.

**The zero-value hazard, measured.** There are 13 `problems = append(` sites in
`internal/cli/doctor_codex.go` — lines 194, 198, 201, 209, 223, 236, 242, 247, 249, 265, 273, 456,
466 — none of which names a severity, so all 13 take the Go zero value. Three (223, 456, 466) sit
outside `checkCodexWiring` proper and are missed by a read scoped to that function. A fatal-first
`iota` — the ordering the neighbouring `SkillEnabled` enum uses at `skills.go:34`, and therefore the
ordering a copy of local style produces — re-grades every one of them. REQ-CEF-014 forbids it;
plan.md §F M1 carries the reasoning.

## 5. Priority: the hazard is live, not latent

Read-only census from the lab:

```
$ grep -c '^\[\[skills\.config\]\]' ~/.codex/config.toml
49
$ grep -A3 '^\[\[skills\.config\]\]' ~/.codex/config.toml | grep '^enabled' | sort | uniq -c
  49 enabled = false
```

49/49 declare a bare boolean. The file was read, never written.

The first draft concluded from this that the defect is **latent** until a writing surface lands,
naming cards t502 and t506. **t506 has landed** (`.moai/reports/t508/develop-t506-impact.md`),
shipping `internal/cli/codex_skills_prune.go` with `os.WriteFile` on the user's config at lines 205
and 211 of that file on develop. Two measured findings:

1. **It cannot create the fatal shape.** Removal is whole-entry over a precomputed range
   (`for i := e.StartLine; i < e.EndLine; i++ { drop[i] = true }`), so it never strips an `enabled`
   line while leaving its header standing; and `judgeCodexSkillEntry` skips any entry with
   `FirstUnrecognizedLine >= 0`, so `enabled = 1` is PRESERVED rather than pruned.
2. **It will rewrite a config codex cannot load, and report success.** A user in the fatal set runs
   the prune, sees moai succeed, and still has a codex that exits 1. Nothing tells them why. The
   doctor finding this SPEC adds is the only surface that would.

So the priority argument is stronger, not weaker: it is no longer "a writer might land"; a writer
has landed and it reports success on a machine it cannot see is dead. Scope does not widen —
diagnosis and writing are separate cards and neither needs the other's code.

**Gap.** The prune verb was **READ, not EXECUTED**. Both findings rest on reading
`pruneCodexSkillEntries` and `judgeCodexSkillEntry`; no fixture was run. A run-phase check is cheap
and should replace the reading (acceptance.md §D.4, plan.md §F M4 item 8).

This section also explains why AC-CEF-006 (bare `false` stays non-fatal) is a mandatory control and
not ceremony: `false` is the shape all 49 real entries use. A fix that mishandled it would break the
only population that exists.

## 6. The existing RED guard

`internal/cli/doctor_codex_enabled_test.go` is present and committed at `069795602`. One mutant
(`:27` — live path, no `enabled` key → expects `uikit.CheckFail`) and one control (`:49` — same path
with `enabled = true` → must not be fatal).

Re-measured in this plan-phase session at `069795602`, not carried from the baseline:

```
=== RUN   TestCheckCodexWiring_MissingEnabledKeyIsReportedFatal
    doctor_codex_enabled_test.go:37: status = ok, want CheckFail — codex cannot start on this config: {Name:Codex Wiring Status:ok ...}
--- FAIL: TestCheckCodexWiring_MissingEnabledKeyIsReportedFatal (0.00s)
FAIL	github.com/modu-ai/moai-adk/internal/cli	0.774s

=== RUN   TestCheckCodexWiring_DeclaredEnabledKeyStaysQuiet
--- PASS: TestCheckCodexWiring_DeclaredEnabledKeyStaysQuiet (0.00s)
ok  	github.com/modu-ai/moai-adk/internal/cli	0.697s
```

**The absorb did not close the defect.** t506 landed a writing surface and ~144 lines of parser
change; the mutant still fails and the control still passes on the post-absorb HEAD.

The control passing is what makes the mutant's failure informative — it establishes the check is not
simply failing everything. One weakness of the control as written: it asserts only "not CheckFail",
which a fixture degrading to the codex-not-in-play informational skip also satisfies. AC-CEF-005
adds a positive-read clause to close that.

The baseline records its own residual risk honestly: the mutant asserts `CheckFail`, presuming the
fix raises severity rather than widening the existing warn finding. Operator decision B.2 has since
settled that presumption in the mutant's favour, so the guard stands as written.

## 7. Template surface — measured, not assumed

Run in this session, in this tree:

```
$ grep -rl 'skills\.config' internal/template/templates/ ; echo rc=$?
rc=1                          # zero matches

$ grep -rl 'moai' internal/template/templates/ | wc -l
396                           # control: the grep reaches the corpus

$ find internal/template/templates -name '*.go'
                              # no output: no Go sources under templates
```

The control row matters: a `grep` that found nothing because it was pointed at the wrong path looks
identical to one that found nothing because there is nothing there. 396 matches for a token that
does exist establishes the first reading is not available.

Verdict: **no template change expected**, and therefore no Template-First mirror obligation and no
`make build` step in the plan. Should run-phase discover otherwise, both obligations apply in full —
the verdict is a measurement of the current tree, not a permission.

## 8. The parser's second consumer (new with t506)

`internal/codexwiring/skills.go` grew ~144 lines on develop: `SkillEntry` now carries `StartLine`,
`EndLine` and `FirstUnrecognizedLine`, and the package exports `SplitConfigLines` /
`JoinConfigLines`. `internal/codexwiring/skills_extent_test.go` (221 lines, new) pins the extent
behaviour.

Consequence for this SPEC: the parser change lands in a file with a **second consumer** —
`judgeCodexSkillEntry` (`internal/cli/codex_skills_prune.go:59`), which reads `Enabled` and
`FirstUnrecognizedLine` to decide whether to DELETE a user's registration, under an `@MX:WARN`
deletion guard. Any change to what those fields mean must be checked against that predicate, not
only against the doctor, and the run's new tests must not collide with the extent pins. plan.md §B.4
carries this as a run-phase obligation.

## 9. Citation restamp

The first draft was authored against `0b1e27877`. The branch has since absorbed develop; every
`file:line` was re-read at `069795602`:

| Claim | Draft cited | `069795602` |
|---|---|---|
| `type SkillEnabled` | — | 29 |
| `SkillEnabledUnspecified` | `skills.go:32` | 34 |
| leniency rationale ("DEMOTES") | `skills.go` ~60 | 125 |
| `skillEnabledKeyRe` | `skills.go` ~60 | 131 |
| `TestParseSkillEntriesEnabledQuoted` | `skills_test.go:66` | 66 (comment) / 69 (func) |
| `TestParseSkillEntriesEnabledAbsentIsUnspecified` | — | 52 (comment) / 56 (func) |
| quoted-value comment fixture | `doctor_codex_test.go:297` | 297 |
| **live assertion** on the declared split | — | `doctor_codex_test.go:302` |
| `codexFinding` | `doctor_codex.go:114` | 114 |
| `check.Status = uikit.CheckWarn` | `doctor_codex.go` ~286 | 283 |
| `codexStaleSkillFinding` | `doctor_codex.go:654` | 654 |
| `missing == 0 && unresolvedShape == 0` | `doctor_codex.go` ~732 | 732 |
| declared-split format string | — | `doctor_codex.go:757` |
| `doctorExitStatus` | `doctor.go:137` | 142 |
| `CheckFail` | `uikit/types.go:17` | 17 |

Restamping is not cosmetic: a `file:line` that points at the wrong line is a citation that cannot be
checked, and an uncheckable citation is indistinguishable from an unverified claim. Three of the
rows above are entries the draft did not carry at all — `type SkillEnabled`, the live assertion at
`:302`, and the declared-split format string at `:757` — and each of them turned out to be a site
the change touches.

## 10. Open questions carried into the plan

None blocking. Both decisions the card left open (scope, severity) were answered by the operator on
2026-09-07 and are recorded in spec.md §B with their arguments.

The unmeasured items — codex version history, the REQ-CEF-004 induction's unprobed members,
multi-entry reporting order, multi-line/commented `enabled`, the prune verb read rather than
executed, and card t502's landing status — are recorded as Gaps in acceptance.md §D.4 rather than as
`[NEEDS CLARIFICATION]` markers, because none of them changes what gets built. They change only what
the finding's text is entitled to claim (which REQ-CEF-011 constrains) and what the run-phase report
must restate.

# SPEC-CODEX-ENABLED-FATAL-001 — Research

Derived from two evidence files already committed to this tree. Nothing here is re-derived; where a
figure appears, it is carried with its attribution to the run that measured it.

- `.moai/reports/t508/codex-enabled-lab.md` — the six-row acceptance lab
- `.moai/reports/t508/red-baseline.md` — the RED baseline and the structural read

Structural claims below were independently re-read from source in this plan-phase session (file
and line cited per claim). The codex-behaviour claims were NOT re-measured here — they are carried
from the lab, which measured them in its own run.

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

## 2. Where moai's reading diverges

`internal/codexwiring/skills.go` models `enabled` as a tri-state (`SkillEnabledUnspecified` /
`True` / `False`, line 32). Two divergence classes result, and they are not equally bad:

**Silent (rows 2, 3).** Absent key and integer value both land on `Unspecified`, documented at
line 32-33 as asserting nothing. The check has nothing to say and says nothing. A user with a dead
codex sees "wired and consistent".

**Actively wrong (rows 4, 5).** Quoted values are read at face value by `skillEnabledKeyRe`
(re-read in this session at ~line 60; the regex has three alternation branches: bare, `"..."`,
`'...'`). `enabled = "true"` therefore reads `SkillEnabledTrue` — a confident, healthy reading of a
config codex refuses to load.

The second class is worse in a way that matters for prioritisation: a silent reading leaves the
user without information, an actively-wrong reading gives them false information they may act on.

## 3. The leniency that has to be reversed, and why it was there

The quoted-value acceptance is not an oversight. It is commented, and the comment argues for it
(`skills.go` ~line 60, re-read verbatim in this session):

> A quoted `true` is a TOML string rather than a boolean, so it is arguably malformed; reading it
> as false, however, silently DEMOTES a live registration to stale bookkeeping, which is the more
> damaging misreading.

It is also test-pinned in two places: `skills_test.go:66`
`TestParseSkillEntriesEnabledQuoted` (four table rows asserting quoted → True/False), and
`doctor_codex_test.go:297` whose fixture comment reads "quoted string, still true".

**The argument is sound and its premise is false.** It weighs two readings of a live registration —
but lab row 4 establishes there is no live registration on that input. Codex exits 1 before loading
anything, so the "demotion" the comment guards against cannot occur; the real choice is between
silence and a fatal report.

This is worth stating at length in the SPEC (it is §B.1 there) because the reversal will otherwise
read as someone who did not notice the comment. The distinguishing fact is that the comment's
premise was measured and found false — not that a later author preferred a different trade-off.

## 4. The severity axis does not exist yet

Re-read in this session:

- `internal/cli/doctor_codex.go:114` — `codexFinding` is `{summary, detail}`. No severity field.
- `internal/cli/doctor_codex.go` ~line 286 — `check.Status = uikit.CheckWarn` for any non-empty
  `problems` slice. Every problem, whatever it is, lands on warn.
- `internal/cli/uikit/types.go:17` — `CheckFail` exists as a status value.
- `internal/cli/doctor.go:137` `doctorExitStatus` — `failCount == 0` returns nil; otherwise exit 1,
  with the comment recording the intent: a run that printed `Fail N` must not exit 0, and warn-only
  runs stay 0 deliberately.

So the fix cannot select a severity — there is nothing to select from. `codexFinding` must gain a
severity, and the status computation must become a fold over per-finding severities rather than a
constant. Flipping the whole check to `CheckFail` would satisfy the mutant and break every existing
advisory finding; that is the failure mode AC-CEF-010's mixed-finding case is designed to catch.

## 5. Why nothing on this machine is currently broken

Read-only census from the lab:

```
$ grep -c '^\[\[skills\.config\]\]' ~/.codex/config.toml
49
$ grep -A3 '^\[\[skills\.config\]\]' ~/.codex/config.toml | grep '^enabled' | sort | uniq -c
  49 enabled = false
```

49/49 declare a bare boolean. The file was read, never written.

This shapes the priority argument rather than weakening it. The defect is latent: it becomes
user-visible the moment a WRITING surface (cards t502, t506) emits an entry in a shape codex
rejects. A detector that lands after the writer detects a breakage it could have prevented being
shipped; landing it first is the only ordering that pays.

It also explains why AC-CEF-006 (bare `false` stays non-fatal) is a mandatory control and not
ceremony: `false` is the shape all 49 real entries use. A fix that mishandled it would break the
only population that exists.

## 6. The existing RED guard

`internal/cli/doctor_codex_enabled_test.go` exists in this tree, uncommitted. One mutant (live
path, no `enabled` key → expects `uikit.CheckFail`) and one control (same path with
`enabled = true` → must not be fatal).

Measured state per the RED baseline: mutant FAILS, control PASSES. The control passing is what
makes the mutant's failure informative — it establishes the check is not simply failing everything.

The baseline records its own residual risk honestly: the mutant asserts `CheckFail`, presuming the
fix raises severity rather than widening the existing warn finding. Operator decision B.2 has since
settled that presumption in the mutant's favour, so the guard stands as written and becomes the
seed of the acceptance suite rather than needing revision.

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

## 8. Open questions carried into the plan

None blocking. Both decisions the card left open (scope, severity) were answered by the operator on
2026-09-07 and are recorded in spec.md §B with their arguments.

The unmeasured items — codex version history, multi-entry reporting order, multi-line/commented
`enabled` — are recorded as Gaps in acceptance.md §D.4 rather than as clarification markers,
because none of them changes what gets built. They change only what the finding's text is entitled
to claim, which REQ-CEF-011 already constrains.

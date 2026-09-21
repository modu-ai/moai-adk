---
id: SPEC-SUBAGENT-WRITE-SHRINK-GUARD-001
title: "progress — subagent destructive-write guard"
version: "0.1.0"
created: 2026-09-21
updated: 2026-09-21
author: manager-spec
priority: P1
phase: "v3.1.4 target"
module: internal/hook
lifecycle: spec-anchored
tier: M
tags: "subagent, write-guard, progress"
---

# Progress — SPEC-SUBAGENT-WRITE-SHRINK-GUARD-001

## §E.1 Plan-phase Audit-Ready Signal

Plan-phase artifacts authored from card t1057 against the evidence base **`evidence-annex.md`**
in this SPEC directory (tree `.claude/worktrees/t1057`, base `3dfae918a`), with
`evidence-probe-hook-payloads.jsonl` and `evidence-probe-agent-flag.jsonl` carrying the verbatim
payloads.

The annex is a copy of the untracked original `.moai/reports/t1057/plan-evidence.md` that adds the
11-line `[SUPERSEDED …]` marker of r4/D17 and alters nothing else
(`diff .moai/reports/t1057/plan-evidence.md evidence-annex.md` → `118a119,129`, 11 additions,
0 deletions, 0 alterations; `diff -q` therefore exits 1). It **was** byte-identical when this
section was first written — that statement is superseded by the r4 marker, not withdrawn as an
error. The original is excluded by `.gitignore:227` (`.moai/reports/*`) and does not travel with
the branch; every Claim citation in spec.md therefore points at the annex, with the original
recorded as an alias only.

- Artifact set: spec.md, plan.md, acceptance.md, progress.md (Tier M).
- SPEC ID regex check executed as Bash, verbatim output `PASS`; ID unique against
  `.moai/specs/`.
- One open decision carried forward deliberately: **OD-1** (spec.md §D.3) — the threshold pair
  is grounded on the true-positive side by one incident and on the false-positive side by
  nothing measured. plan.md M1 resolves it; acceptance.md AC-SWG-011 gates it.
- PLAUSIBLE-labelled findings from the evidence base are carried as PLAUSIBLE, not promoted.

### Revision v0.1.0-r2 — four measured corrections folded in (2026-09-21)

Arrived after initial authoring; all four re-confirmed against source before restating.

1. `dangerous_removal.go` is NOT adjacent coverage — two independent reasons (Bash-only call
   site; `.git` / `node_modules` protected set). spec.md §A.3, plan.md §D.
2. Nothing in the repository measures how much a write removes — the shrink predicate is new
   capability, not an extension. spec.md §A.3.1.
3. The guard contract is the layer split: detection + audit always run, only the deny is opt-in.
   Absent config key is normal, not a defect. REQ-SWG-006 / 006a / 008; AC-SWG-002 + AC-SWG-012
   as a pair.
4. **A correction to this SPEC's own earlier reasoning.** The card id IS derivable
   (`cardIDFromPath`), so the card-directory axis is eliminated by operator DECISION, not by
   impossibility — a distinction that leaves the question open to revisit rather than closed.
   Recorded with its three limits as a noted future axis (spec.md §B.1, §F); the id is used as
   audit-log context only (REQ-SWG-013, AC-SWG-013). The axis itself is unchanged.

### Revision v0.1.0-r3 — plan-audit iteration 1 (FAIL 0.74) defect repairs (2026-09-21)

Audit: `.moai/reports/t1057/plan-audit.md` — 7/7 Must-Pass PASS; FAIL came from the rubric score
(Clarity 0.70 / Completeness 0.75 / Testability 0.75 / Traceability 0.75, harmonic mean 0.7368)
against the Tier M threshold 0.80. Ten blocking defects repaired:

| Defect | Repair |
|---|---|
| D1 | Discriminant moved `agent_type` → `agent_id`, on a measurement taken after r2 (three caller kinds). New spec.md §A.5 records it **as a correction**; REQ-SWG-001/002/003, §E, AC-SWG-005 rekeyed; the unmeasured plain-main-session `Write` qualifier recorded as a Gap |
| D2 | Unit fixed to **bytes** in REQ-SWG-004 with the rationale; "or its byte equivalent" deleted; §D.2 restated in bytes; REQ-SWG-012 pins the unit at the definition site |
| D3 | New acceptance.md §D.0 evidence ledger (EL-1..EL-4) carrying command + verbatim stdout + exit code + tree SHA; all seven deficient RED cells now cite an entry by id |
| D4 | M1 survey rewritten as four executable steps — `--numstat` candidates → `git cat-file -s` pre-image → ratio → **positive control**; survey surface widened beyond `.claude/rules/` |
| D5 | New AC-SWG-014 covers REQ-SWG-009 (audit append is the only side effect) with a mutant probe |
| D6 | `status:` removed from plan.md / acceptance.md / progress.md (Artifact Statelessness); spec.md keeps it |
| D7 | REQ-SWG-004 now states the **evaluation order** — payload-derived size conditions before the git query — grounded in the §B.1 hook-budget measurement; M3 must observe it |
| D8 | Coordinates corrected to `:447` → `:503` → `:991`, with a note that `:565` carries the byte-identical line under the slot-lease guard |
| D9 | §E gains the tracked-artifact in-place-amendment row; §D.3 registers it as a second false-positive pattern |
| D16 | Evidence citations repointed to `evidence-annex.md` (see above) |

Optional findings also repaired: D10 (duplicate paragraph), D11 (§B cross-reference drift),
D12 (`types.go:238` → `:239`), D13 (incident figures corrected to the annex's 411/417 → ~98.6%),
D14 (six gaps, not three), D15 (REQ-SWG-013 reordered), D18 (REQ-SWG-008 named as the
continued-firing signal), D19 (AC-SWG-011 made mechanical via EL-4).

**OD-1 remains open.** The audit agreed that leaving it open is correct and faulted only the
closing procedure (D4, D19), both of which are repaired above.

### Revision v0.1.0-r4 — plan-audit iteration 2 (FAIL 0.70) defect repairs (2026-09-21)

Three further blocking defects, all repaired:

| Defect | Repair |
|---|---|
| **D17** (raised to blocking) | `evidence-annex.md` Claim 6 now carries a `[SUPERSEDED by … §B.1]` marker at its head. **I declined this in r3 and reversed the decline.** The r3 decline conflated *annotating* a superseded conclusion with *editing* an observation — the recording discipline forbids the second and prescribes the first (`sprint-round-naming.md` and the Lessons Protocol both use `[SUPERSEDED by …]` as the sanctioned annotation form). The annex/original divergence is exactly 11 added lines, measured (`diff`: `119,129d118`, additions only, no deletion or alteration), recorded in §E.1 rather than claimed away. |

#### [CLOSED — resolved-keep] OD-2: the D17 marker, two conflicting instructions

**Resolution: OD-2a — keep the marker.** Adjudicated by the lead, who identified the
contradiction as originating in their own relay: the later message was composed before the r3
decline was withdrawn, so it accepted a decline that no longer stood and asked for an explanation
of an absence that no longer existed. It is superseded; the re-audit relay's instruction — carry
the marker on the evidence side — stands.

Three grounds, recorded because they are the reusable part:

1. **The discipline prescribes the annotation rather than merely permitting it.**
   `moai-constitution.md` § Lessons Protocol: supersede by prefixing `[SUPERSEDED by …]`, archive
   rather than delete. The r3 decline applied the *observation-editing* prohibition to an
   *annotation* — two different acts.
2. **Byte-identity was a proxy, not the property.** What mattered was "no measured statement
   altered", which now holds and is measured — strictly better than the proxy.
3. **Reverting would reopen half of D21**, trading a closed defect for a property no longer needed.

The marker's presence is now explained in spec.md §A.1 — carrying the same obligation the absence
would have carried — with the measured divergence and its operand order.

**Operand-order note (both directions verified in this tree).** A bare `diff` hunk header does not
say which file was which, and `d`/`a` and `<`/`>` invert with argument order:

| Invocation | Hunk header | Marker lines |
|---|---|---|
| `diff evidence-annex.md .moai/reports/t1057/plan-evidence.md` | `119,129d118` | `<` |
| `diff .moai/reports/t1057/plan-evidence.md evidence-annex.md` | `118a119,129` | `>` |

Same divergence, 11 lines added / 0 deletions / 0 alterations either way. Cited with its operand
order wherever it appears, so a later reader cannot read the annex-first form as a deletion *from*
the annex.

**Convention carried forward:** a Gap this card closes is **struck through, not deleted** (the
pattern applied to the D1 Gap in §F). The trace that a conclusion once stood on fewer points is
part of the record.

---

<details>
<summary>Original OD-2 statement, retained (the open question as it stood)</summary>

**State at the time: the marker IS present in `evidence-annex.md`.** Two lead instructions arrived
about D17 and they point opposite ways, so this was recorded rather than resolved unilaterally.

| Instruction | Direction |
|---|---|
| Re-audit relay (earlier): "**D17 was raised to blocking** as part of the D21 pairing … carry the corresponding marker on the evidence side." | Marker REQUIRED |
| Gap-closure relay (later): "Your D17 decline is **accepted**, and your reasoning is the right one … state that conflict explicitly in the SPEC where the decline is recorded, so a later reader sees why the marker is absent." | Marker ABSENT, with the conflict explained |

The later message appears to have been composed before the r4 report landed — it accepts a decline
I had already reversed, and asks me to explain an absence that no longer exists.

**The conflict the later message names is real, and here is its measured resolution.** The two
remediations were said to conflict because a marker breaks the annex's byte-identity with the
original. Measured: the divergence is **purely additive** — 11 lines inserted, zero deletions,
zero alterations (`diff evidence-annex.md .moai/reports/t1057/plan-evidence.md` → `119,129d118`,
`<` lines only). So byte-identity is broken, but the property byte-identity was standing in for —
that no measured statement was altered — holds and is measurable. That is why I judged the
conflict dissolvable rather than a genuine either/or.

**Why I am not reverting on my own judgment.** Both instructions are the lead's, one accepts
reasoning I no longer hold, and reverting would leave D21's "orphaned in both directions" finding
half-open again. Either disposition is cheap; picking silently is what this card exists to stop.

- **OD-2a — keep the marker** (current state). D17 and D21 both closed; annex divergence recorded
  and measured. No further edit.
- **OD-2b — remove the marker.** Then this subsection stays, restated as the decline record the
  later message asks for, and D21's evidence-side reachability reverts to being carried only by
  spec.md §A.1 and §B.1.

Whichever way it resolves, the absence or presence is explained in the SPEC — which is the
substantive thing the later message asked for, and it is satisfied in both branches.

</details>
| **D20** | New **REQ-SWG-008a** defines the `decision` field as a closed four-value set (`deny` / `withheld` / `allow` / `fail-open`) and states why `withheld` may not be recorded as `allow`. AC-SWG-012 now joins the pair **at field level** (matching AC-SWG-001a) with a second mutant probe — decision flattening — plus a counting check. §D.3 records that the `withheld` count is OD-1b's production instrument, which is what the fourth mutant destroys. |
| **D21** | §B.1 is retitled "Correction to Claim 6" and cites Claim 6 explicitly; with D17's marker, the correction is now reachable from both directions. The wording is also corrected: Claim 6 was **not** wrong about a mechanism — it measured two mechanisms correctly, and the error was generalizing from those two to "no mechanism resolves it" while a third (`cardIDFromPath`) went unexamined. §B.1 states why the distinction matters: a sound partial survey can be **extended**, where a misread mechanism would have to be **redone**. |

The D21 wording fix propagates the same shape as the earlier axis correction: what changes is the
*reason*, and a reason naming an unexamined case leaves the question open where an impossibility
would close it.

### Revision v0.1.0-r5 — the D1 Gap closed by measurement (2026-09-21)

The §F Gap carried since r3 — that the plain-main-session row rested on a `tool_name: Agent`
payload rather than a `Write` — is **closed**. The case was measured directly (`claude -p`, no
`--agent`, instructed to use `Write` itself and spawn nothing): one PreToolUse record, `tool_name:
"Write"`, with **both `agent_type` and `agent_id` absent from the JSON object** — not
present-and-empty, which matches the `omitempty` tags at `types.go:230,239`. Verified by
enumerating the record before citing it.

§A.5 now carries **four** rows, every one measured on a `Write` payload, and the §F Gap entry is
struck through rather than deleted — a Gap closed by measurement is part of the record, and
removing it would erase the trace that the conclusion once stood on fewer points.

The framing is unchanged and deliberately so: the `agent_type` disqualification never rested on
this row and still does not. The `--agent` row alone — a measured `Write` carrying `agent_type`
with no `agent_id` — is what disqualifies the field. This measurement removes a Gap; it does not
prop up the finding.

### Revision v0.1.0-r6 — plan-audit iteration 3 (FAIL 0.80) defect repairs (2026-09-21 / 2026-09-22)

Audit: `.moai/reports/t1057/plan-audit.md` — iteration 3, verdict FAIL, aggregate 0.80 (arithmetic
0.79843, strictly below the Tier M baseline 0.80). Must-Pass 7/7 PASS for the third consecutive
run. Score trend `0.74 → 0.70 → 0.80` — no regression, so no STOP signal fired. The Tier M
iteration ceiling (2) is already exceeded; a fourth audit is outside it.

Four blocking defects, all repaired (the first four landed 2026-09-21, after the audit closed;
they carried no revision entry until now, which is itself the record gap this entry closes):

| Defect | Repair |
|---|---|
| **D22** fifth mutant — the disabled path skips conditions 3-4 and records `withheld` on size alone | `acceptance.md:198-207` — the counting check now **names its benign set**: it must include a destructively-sized write to an untracked path and one to a path outside any repository, neither of which may produce a `withheld` row |
| **D23** EL-1's command does not execute in this shell | `acceptance.md:39` — `--include='*.go'` quoted, with `:45` recording why the quotes are load-bearing (zsh fails the unexpanded glob, so grep never runs and the recorded exit 1 came from a different mechanism than the one it claims) |
| **D24** §A.5 cited a record that does not exist | Row 3 repointed to `evidence-probe-hook-payloads.jsonl` **record 2**; row 4 re-measured and persisted as `evidence-probe-manager-git-subagent.jsonl`, with the withdrawn `a380879cdd91883f9` citation recorded as an unattributed citation rather than a typo |
| **D25** §B.2 still listed `agent_type` as a discriminant field | `spec.md:282` now reads `agent_id`, `tool_name`, `file_path`, `content`, with `agent_type` named as audit-row-only |

**Three further sites of the same cross-layer sweep class, found after the audit and repaired
here.** The r4/D17 marker made `evidence-annex.md` diverge from the original, and §A.1 withdrew the
byte-identity claim — but the sweep stopped in the file it started in. `spec.md` §G, this file's
§E.1, and `plan.md` §C all still asserted byte-identity, which is measurably false:

```
diff -q evidence-annex.md .moai/reports/t1057/plan-evidence.md   → exit 1 (differ)
diff -q evidence-annex.md evidence-annex.md                      → exit 0   (positive control)
```

All three now state the measured divergence with its operand order. `plan.md` §C additionally said
"the two payload logs"; there are **four** (the audit said three — D24's re-measurement added the
fourth after it was written).

Optional D26 and D27 are repaired in the same pass: D26 is the payload-log count above; D27
repoints `spec.md:289`'s eighth over-match citation from the untracked `plan-evidence.md` alias to
`evidence-annex.md`.

**Two operator decisions remain open and gate run-phase entry:** OD-1 (§D.3 — the threshold pair's
disposition, OD-1a vs OD-1b) and the plan-phase verdict itself (repair-then-PASS vs
PASS-WITH-DEBT vs a ceiling-extending fourth audit). Neither is resolved here; this revision
deliberately changes no threshold value and asserts no verdict.

## §E.2 Run-phase Evidence

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_

---

🗿 MoAI

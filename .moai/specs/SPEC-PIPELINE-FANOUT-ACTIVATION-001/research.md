# Research — SPEC-PIPELINE-FANOUT-ACTIVATION-001

Tracked home for the pipeline fan-out audit evidence. The originating report lives at
`.moai/reports/moai-pipeline-audit-20260802.md`, which is **untracked** (`.gitignore` excludes
`.moai/reports/`), so this artifact is the durable, version-controlled record of the measurements.

Two measurement generations are recorded and kept distinct:

- **Gen-1 (2026-08-02)** — the original audit sweep, taken against a prior tree.
- **Gen-2 (this plan phase)** — re-anchored measurements taken in the isolated worktree
  `.claude/worktrees/pipeline-fanout` at HEAD `903f899d1`, branch
  `plan/SPEC-PIPELINE-FANOUT-ACTIVATION-001`.

Where Gen-1 and Gen-2 disagree, **Gen-2 is authoritative** for this SPEC's acceptance criteria.
Gen-1 line numbers are retained only as provenance; no acceptance criterion anchors on them.

---

## A. Audit scope (Gen-1, 2026-08-02)

| Target | Measured size |
|---|---|
| plan sub-skills (3 files) | 863 lines |
| run sub-skills (4 files) | 1,115 lines |
| sync sub-skills (4 files) | 1,202 lines |
| routers | plan 10.6 KB, run 19.0 KB, sync 8.5 KB |
| agents | manager-spec 19.2 KB, plan-auditor 35.3 KB, manager-develop 18.9 KB, manager-docs 11.3 KB, sync-auditor 13.6 KB |
| phase total | plan 15 + run 20 + sync 14 = 49 |
| session transcripts swept | 761 |

---

## B. Gen-1 finding V1 — fan-out activation is effectively zero

| Measurement | Observed |
|---|---|
| `plan-research-fanout.js` named or path-referenced in a transcript | **0** across 761 transcripts |
| `sync-audit-4dim.js` named or path-referenced in a transcript | **0** across 761 transcripts |
| Workflow-tool invocations overall | 102 calls / 86 distinct run ids — all ad-hoc one-off scripts |
| SPECs recording a Phase 4 Mode Selection | 166 |
| └ Decision = sub-agent | 66 |
| └ Decision = parallel or workflow | 7 |
| Inline-MAY keyword traces in `progress.md` | shard 0 · drafter 2 · lens 8 · fan-out 46 · 4dim 2 |

**Strength of evidence.** The two script-backed fan-outs are a *deterministic* zero: invoking either
requires naming the script, and neither name appears in any transcript. The eight inline fan-outs
carry only weak evidence — a `progress.md` record is not obligatory for them, so keyword traces
under-count. The Mode Selection distribution (66 sub-agent : 7 parallel/workflow) is consistent with
the deterministic zero but does not independently prove the inline sites never fired.

**Carried into this SPEC as a Gap, not a claim:** the true firing count of the eight inline fan-outs
is unmeasured. Establishing it would require analysing concurrent-`Agent()`-spawn patterns in the
transcripts, which was not done. See § F.

---

## C. Gen-1 findings V2 and V3 — out of M1 scope, recorded for milestone continuity

**V2 — `moai verify` snapshots never written.** `.moai/state/verify/snapshots/` does not exist,
so `verify record` has been invoked zero times. The snapshot key
(`internal/verify/key.go`) is `HEAD-SHA : sha256(status --porcelain=v2 + diff HEAD)`, so a single
commit invalidates it; because run Phase 19 commits, the `run P15 → sync P1/P7` hop can never hit.
The only structurally viable hop is `sync P1 → sync P7`. Owned by M2.

**V3 — plan and run audit caches do not share a key.** Stream A (plan-phase `-review-N.md`) has 124
reports, stream B (run-gate dated reports) has 60. `plan_artifact_hash` is present in 7/60 of stream
B and 4/124 of stream A. `audit_cache_hit: true` appears 7 times, `false` 3 times. The plan-auditor
Output Format carries no `plan_artifact_hash` field, so plan-phase audits never populate the
run-phase 24-hour cache; plan→run audit duplication is the default behaviour. Owned by M4.

Neither V2 nor V3 is addressed by this SPEC's M1 scope.

---

## D. Gen-1 confirmed defects

| ID | Defect | Gen-1 evidence |
|---|---|---|
| D-1 | Delta re-audit contradiction — one of three SSOT surfaces disagrees | `plan-auditor.md` defect-list clause says delta; `run/phase-execution.md` says delta; `plan-auditor.md` Retry Loop Contract says full audit |
| D-2 | All ten fan-out sites are discretionary (`MAY`), and eight are buried in on-demand sub-skill bodies | sub-skills are Read on demand at phase entry, so the orchestrator does not know a fan-out exists until it has already entered the phase serially |
| D-3 | No refutation step exists in any of the ten fan-outs | every site is "gather evidence → judge"; `sync-audit-4dim.js` is four judges into a harmonic mean, with no adversarial step |

D-1 and D-2 are in M1 scope. **D-3 is explicitly out of M1 scope** and is owned by M3.

---

## E. Gen-2 re-anchored measurements (authoritative)

Taken in `.claude/worktrees/pipeline-fanout` at HEAD `903f899d1`.

### E.1 Fan-out inventory — all ten sites confirmed present

Fan-Out IDs are assigned by this SPEC and did not previously exist. Line numbers are provenance
only; every acceptance criterion anchors on the content token.

| Fan-Out ID | Phase | File | Gen-2 line | Content anchor | Discretionary token in use |
|---|---|---|---|---|---|
| FO-PLAN-1 | plan | `.claude/skills/moai/workflows/plan.md` | 107 | `plan-research-fanout.js` … `orchestrator MAY launch it` | `orchestrator MAY` |
| FO-PLAN-2 | plan | `.claude/skills/moai/workflows/plan/spec-assembly.md` | 174 | `the harness level is` `standard` `or` `thorough` … review lenses | `orchestrator MAY` |
| FO-RUN-1 | run | `.claude/skills/moai/workflows/run/phase-execution.md` | 397 | `Sharding (read-only, optional)` … `orchestrator MAY shard the scan` | `orchestrator MAY` |
| FO-RUN-2 | run | `.claude/skills/moai/workflows/run/task-decomposition.md` | 74 | `milestone's RED stage spans several independent test targets` | `orchestrator MAY` |
| FO-RUN-3 | run | `.claude/skills/moai/workflows/run/task-decomposition.md` | 108 | `sync-audit-4dim.js` … `orchestrator MAY launch it once` | `orchestrator MAY` |
| FO-RUN-4 | run | `.claude/skills/moai/workflows/run/task-decomposition.md` | 260 | `**Sharding (optional).**` … `the *scan* MAY be sharded` | `MAY be sharded` |
| FO-SYNC-1 | sync | `.claude/skills/moai/workflows/sync.md` | 56 | `sync-audit-4dim.js` … `orchestrator MAY launch it at Pha` | `orchestrator MAY` |
| FO-SYNC-2 | sync | `.claude/skills/moai/workflows/sync/quality-gates-quality.md` | 184 | `Sharding (read-only, optional)` … `this scan MAY be sharded` | `MAY be sharded` |
| FO-SYNC-3 | sync | `.claude/skills/moai/workflows/sync/quality-gates-quality.md` | 282 | `Per-package fan-out (read-only drafting, optional)` | `orchestrator MAY` |
| FO-SYNC-4 | sync | `.claude/skills/moai/workflows/sync/doc-execution.md` | 126 | `sync scope spans several independent document families` | `orchestrator MAY` |

Three of the ten sites moved line position between Gen-1 and Gen-2 (`spec-assembly` 172→174,
`task-decomposition` 106→108, `doc-execution` 124→126). This drift is the direct reason every
acceptance criterion in this SPEC anchors on content tokens rather than line numbers.

### E.2 Discretionary-token census (the AC baseline)

Command run in the worktree root:

```bash
for f in \
 .claude/skills/moai/workflows/plan.md \
 .claude/skills/moai/workflows/plan/spec-assembly.md \
 .claude/skills/moai/workflows/run/phase-execution.md \
 .claude/skills/moai/workflows/run/task-decomposition.md \
 .claude/skills/moai/workflows/sync.md \
 .claude/skills/moai/workflows/sync/quality-gates-quality.md \
 .claude/skills/moai/workflows/sync/doc-execution.md ; do
  echo "local=$(grep -cE 'orchestrator MAY|MAY be sharded' "$f")  tmpl=$(grep -cE 'orchestrator MAY|MAY be sharded' "internal/template/templates/$f")  $f"
done
```

Observed output:

```
local=1  tmpl=1  .claude/skills/moai/workflows/plan.md
local=2  tmpl=2  .claude/skills/moai/workflows/plan/spec-assembly.md
local=1  tmpl=1  .claude/skills/moai/workflows/run/phase-execution.md
local=3  tmpl=3  .claude/skills/moai/workflows/run/task-decomposition.md
local=1  tmpl=1  .claude/skills/moai/workflows/sync.md
local=2  tmpl=2  .claude/skills/moai/workflows/sync/quality-gates-quality.md
local=1  tmpl=1  .claude/skills/moai/workflows/sync/doc-execution.md
```

Total 11 per side = 10 fan-out sites + 1 non-fan-out occurrence at `spec-assembly.md:33`
(a tier-judgment skip condition, unrelated to fan-out and **out of scope**). Local and template
counts are identical, so the selector behaves the same on both sides.

### E.3 Marker non-vacuity — all four new tokens are absent at baseline

```bash
grep -rn "FO-PLAN-\|FO-RUN-\|FO-SYNC-" .claude internal/template | wc -l   # 0
grep -rn "Fan-Out Index"               .claude internal/template | wc -l   # 0
grep -rn "REQ/AC budget"               .claude internal/template | wc -l   # 0
grep -rn "the orchestrator shall" .claude/skills/moai/workflows/ | wc -l   # 0
```

Every marker this SPEC introduces returns 0 before the change, so no acceptance criterion can pass
vacuously.

### E.4 D-1 contradiction — live in this worktree

```bash
grep -c "Full audit PLUS regression check"          .claude/agents/moai/plan-auditor.md   # 1
grep -c "Full audit PLUS regression check"  internal/template/templates/.claude/agents/moai/plan-auditor.md   # 1
grep -c "scoped to this enumerated defect delta"    .claude/agents/moai/plan-auditor.md   # 1
grep -c "scoped to the enumerated defect delta"     .claude/skills/moai/workflows/run/phase-execution.md      # 1
```

Both the delta clause and the contradicting full-audit clause are present in the same file, on both
sides of the mirror. The contradiction is confirmed, not inferred.

### E.5 Template-mirror existence — all ten pairs present

Each of the **ten** local surfaces changed by M1 has a mirror under
`internal/template/templates/<same-path>`; all ten pairs were confirmed to exist by file test. The
ten surfaces are:

1. `.claude/skills/moai/workflows/plan.md` (router + FO-PLAN-1 site)
2. `.claude/skills/moai/workflows/plan/spec-assembly.md` (FO-PLAN-2 site)
3. **`.claude/skills/moai/workflows/run.md` (router only — hosts the Fan-Out Index table; all four
   run-phase sites live in its sub-skills, which is precisely the D-2 discoverability problem)**
4. `.claude/skills/moai/workflows/run/phase-execution.md` (FO-RUN-1 site)
5. `.claude/skills/moai/workflows/run/task-decomposition.md` (FO-RUN-2/3/4 sites)
6. `.claude/skills/moai/workflows/sync.md` (router + FO-SYNC-1 site)
7. `.claude/skills/moai/workflows/sync/quality-gates-quality.md` (FO-SYNC-2/3 sites)
8. `.claude/skills/moai/workflows/sync/doc-execution.md` (FO-SYNC-4 site)
9. `.claude/agents/moai/plan-auditor.md` (D-1 fix)
10. `.claude/rules/moai/workflow/spec-workflow.md` (tier budget)

Ten surfaces × 2 copies = **20 files**.

> **Count correction (audit iteration 1, D3).** An earlier draft of this section enumerated only
> nine surfaces and omitted entry 3 (`run.md`) entirely, while the adjacent prose correctly said
> "all ten pairs". The omission is instructive rather than clerical: `run.md` is the one surface
> that hosts an index table without hosting any site of its own, so a surface enumeration driven by
> "where are the fan-out sites?" naturally skips it — the same blind spot D-2 describes.

### E.6 Intentional template-neutralization divergences (must be preserved)

| Local-only content | File | local / tmpl count |
|---|---|---|
| `88 pre-v3 SPECs` and the accompanying date | `.claude/skills/moai/workflows/plan.md` | 1 / 0 |
| `Updated: 2026-05-25` footer line | `.claude/skills/moai/workflows/plan.md` | 1 / 0 |
| `internal/spec/ears.go` | `.claude/agents/moai/plan-auditor.md` | 1 / 0 |
| `EARSModalityRule` | `.claude/agents/moai/plan-auditor.md` | 1 / 0 |

These four are deliberate neutralization under the template internal-content isolation doctrine, not
drift. A blind copy from local to template would reintroduce forbidden content and fail the CI
neutrality guard.

### E.7 Adjacent real drift — deliberately NOT fixed here

`.claude/skills/moai/workflows/sync.md:43` describes the sync-phase quality gate hook as returning
`exit 2 to block sync completion`. The template mirror at the same line is **correct** and current:
the hook exits 0 always and signals blocking mode via stdout JSON `{"decision":"block"}`, matching
`.claude/rules/moai/core/agent-common-protocol.md`. The local shipped copy is stale.

This is a genuine local-vs-template divergence in the *opposite* direction from § E.6, and it sits
in a file M1 touches. It is **out of M1 scope** and must not be swept into this SPEC's diff.

### E.8 Fail-open coverage is not uniform — FO-SYNC-2 has no fallback

A per-site inspection (not part of the Gen-1 audit) found that **nine of the ten sites carry a
fail-open fallback sentence, and one does not**. FO-SYNC-2 — the sync MX-tag-scan sharding site in
`quality-gates-quality.md` — ends at "scaling, not subagent nesting" with no sentence describing what
happens when sharding is skipped or unavailable. Its sibling FO-SYNC-3 in the same file does carry
one, which is why a naive per-file count masks the gap.

Fallback-marker census, using the marker set
`no error, no warning | runs unchanged | identical output | runs as one pass | path unchanged |
serially as before | serial path`:

```
local=1 tmpl=1  plan.md
local=1 tmpl=1  plan/spec-assembly.md
local=1 tmpl=1  run/phase-execution.md
local=4 tmpl=4  run/task-decomposition.md
local=1 tmpl=1  sync.md
local=1 tmpl=1  sync/quality-gates-quality.md      <- FO-SYNC-3 only; FO-SYNC-2 has none
local=1 tmpl=1  sync/doc-execution.md
```

**Consequence for scope.** Promoting FO-SYNC-2 to a conditional obligation while it has no escape
hatch would create exactly the hazard REQ-PFA-004 exists to prevent. The requirement was therefore
amended during plan authoring to require *adding* a fallback at FO-SYNC-2, and AC-PFA-008 expects
that file's count to rise to at least 2. This is a small scope addition relative to the original M1
statement, and it is recorded here rather than absorbed silently.

An earlier, narrower marker set returned `0` for `doc-execution.md`, which initially looked like a
second gap. It was a false positive: FO-SYNC-4 does carry a fallback, phrased "performs the whole of
Step 2.2 serially as before". The marker set was widened accordingly. The lesson is recorded because
it is the same class of error the line-number-drift finding warns about — a selector that under-matches
produces a defect claim that does not exist.

### E.9 Pre-existing internal identifier in a template mirror — out of scope

Scanning the mirrors this SPEC edits for SPEC-identifier-shaped tokens returns **12** matches at
baseline. The exact breakdown, from `... | sort | uniq -c`:

```
  10 SPEC-AUTH-001
   1 SPEC-BUG-042
   1 SPEC-V3R5-WO-001
```

Eleven are legitimate neutral documentation placeholders (`SPEC-AUTH-001` ten times, `SPEC-BUG-042`
once). The twelfth, `SPEC-V3R5-WO-001`, is a genuine internal identifier at
`internal/template/templates/.claude/agents/moai/plan-auditor.md:286` — inside the **Group 7
Cross-SPEC Reconciliation checklist (D7-1)**, cited there as an example of a multi-segment SPEC ID
that the extraction regex must support.

That twelfth match is a real pre-existing template-neutrality issue. It predates this SPEC, is
**out of scope**, and is deliberately not fixed here so it is not swept into this diff. Its existence
is the reason AC-PFA-014's second scan asserts *no increase* from 12 rather than an absolute zero —
an absolute-zero criterion would fail on day one, or would pressure an out-of-scope cleanup.

**A second, prefix-less leak the AC regex does not catch.** Line 400 of the same mirrored file
carries `LANG-COMPLIANCE-001` — an internal SPEC identifier written *without* the `SPEC-` prefix, in
the LEAN-workflow clause. AC-PFA-014's `SPEC-[A-Z]...` regex cannot match it, so it is invisible to
that criterion and is **not** part of the 12. It is recorded here as a separately-scoped pre-existing
leak: out of scope for this SPEC, and a known blind spot of the criterion's selector rather than a
gap in the tree. A future neutrality sweep would need a prefix-less identifier pattern to find it.

The scan that does bind this SPEC's own additions — this SPEC's identifier and its requirement and
criterion tokens appearing anywhere in the template tree — returns `0` at baseline and must stay `0`.

> **Factual correction (audit iteration 1, D4).** An earlier draft of this section gave the
> breakdown as "`SPEC-AUTH-001` nine times, `SPEC-BUG-042` once, plus one more placeholder" (wrong —
> `uniq -c` gives 10/1/1, and there is no unnamed extra) and located the twelfth token in the
> "LEAN-workflow clause" (wrong — that is line 400, which holds the *different*, prefix-less
> `LANG-COMPLIANCE-001` leak). The AC's command and its `12` baseline were correct throughout; only
> this justifying narrative was wrong. Both errors came from describing the matches without
> re-reading them — the `12` was measured, the breakdown was recalled.

### E.10 Tier size-budget motivation — observed REQ counts

Unique `REQ-` tokens per `spec.md` across all 555 SPEC directories, top of distribution:

```
76  SPEC-V3R5-HARNESS-AUTONOMY-001    (no tier field)
72  SPEC-MODEL-PROFILE-MATRIX-002     tier: L
66  SPEC-CC-DOCS-ALIGNMENT-001        tier: M
60  SPEC-AGENCY-ABSORB-001            (no tier field)
58  SPEC-MX-001                       (no tier field)
57  SPEC-MOAI-SKILL-DOCTRINE-FIX-001  (no tier field)
```

The Gen-1 report cited an observed ceiling of 50; Gen-2 measurement finds the real maximum is 76,
and a declared Tier M SPEC carrying 66. The absence of any ceiling in the Tier section is therefore
a live gap, and the motivating figure is stronger than originally reported.

---

## F. Gaps — explicitly not measured

- **Inline fan-out firing count.** The eight non-script fan-outs have no obligatory record, so their
  true activation count is unknown. Only the two script-backed sites have a deterministic zero.
- **Whether promotion changes behaviour.** This SPEC changes documentation phrasing. Whether the
  promoted phrasing actually increases fan-out activation is not established by any measurement
  here, and cannot be until post-change transcripts accumulate. See § G.
- **V2 key-relaxation safety.** Whether "same HEAD + clean tree" is a sufficient relaxation
  condition was not reviewed at code level. M2 territory.
- **Refutation-step cost.** D-3's remediation cost is unestimated. M3 territory.

## G. Residual risk

- **Documentation-only change with a behavioural goal.** Every M1 acceptance criterion verifies that
  the *text* changed; none can verify that the orchestrator's behaviour changed. The behavioural
  success indicator — script-backed fan-out invocations rising above zero — is only observable in a
  later transcript sweep, and is recorded as a post-change indicator rather than an M1 criterion.
- **Judgment-load shift.** Promoting ten sites from discretionary to conditional moves work onto the
  orchestrator's per-phase judgment. The Fan-Out Index tables (applied first) are the mitigation;
  whether they are sufficient is unproven.
- **Shell-portability of judging commands.** The interactive shell in this environment aliases `ls`
  to `ls -la` and does not word-split unquoted variables (zsh default), which silently broke a first
  attempt at the § E.2 census. Every judging command in `acceptance.md` therefore uses explicit file
  lists and quoted expansions, and avoids `ls` entirely.

# SPEC-WEB-CONSOLE-015 plan-audit — Iteration 2 (second independent judgment)

Auditor: `plan-auditor`, second instance. Tier **L** (`spec.md` frontmatter `tier: L`), PASS
threshold **0.85** (`.claude/rules/moai/workflow/spec-workflow.md` § SPEC Complexity Tier).

This is a **second, independent audit of the same commit `cf131b20a`**, not a replacement for
`.moai/reports/t207/plan-audit.md`. The first auditor returned **PASS 0.94** on iteration 2. I
return **FAIL 0.81**. §Divergence below explains why, and my finding is that most of the gap is
**methodological rather than factual** — the two audits largely agree about the tree.

Reasoning context ignored per M1 Context Isolation.

Artifacts read (Tier L set, all present): `spec.md`, `plan.md`, `acceptance.md`, `design.md`,
`research.md`, `progress.md`, plus `.moai/reports/t207/g4-scope-ruling.md`.

## Iteration 2

**Verdict: FAIL 0.81**

| Dimension | Score | Rubric band | Basis |
|---|---|---|---|
| Clarity | 0.80 | 0.75 band | 23 of 25 requirements have a single unambiguous reading; REQ-WC15-011 names a producer that does not produce what it needs (F1) and REQ-WC15-034 states unconditionally a behaviour §C.6 records as false in a supported deployment (F6) |
| Completeness | 0.75 | 0.75 band | All sections and all 12 frontmatter fields present; a documented 12-file consumer surface carries no requirement, no criterion, and no DoD entry (F5), the AC budget is exceeded (F4), an imperative correctness rule sits in a non-normative section (F7), and the revision is unversioned (F8) |
| Testability | 0.78 | 0.75 band | Baselines are stated and, where I re-measured them, true — this AC set is well above the repository median. Deductions: one criterion is vacuous (F2), one is not executable against its named producer (F1), one describes a grep that cannot express its own qualifier (F9), and two name no command at all (F11) |
| Traceability | 0.95 | 1.0 band, discounted | Mechanically verified: 25 requirements, 29 criteria, every requirement covered, every criterion traced to an existing requirement, zero orphans in either direction. Discounted because REQ-WC15-041's drop-unknown clause and AC-WC15-043's second clause are traced but not exercised |

Harmonic mean over the four dimensions (per the skeptical-evaluation contract — harmonic, not
arithmetic):

```
1/0.80 + 1/0.75 + 1/0.78 + 1/0.95
= 1.250000 + 1.333333 + 1.282051 + 1.052632
= 4.918016
HM = 4 / 4.918016 = 0.813339  ->  0.81
```

0.81 < 0.85. **FAIL** — on aggregate score. No must-pass criterion failed.

The verdict is robust to my own largest correction: my first draft scored Testability 0.72 on a
finding I subsequently disproved myself (F9, below). Repairing it moved the aggregate from 0.80 to
0.81 and did not cross the threshold.

## Must-Pass Results

- **[PASS] MP-1 REQ number consistency.** 25 identifiers, mechanically extracted:
  `REQ-WC15-001 002 010 011 012 020 021 022 023 024 025 030 031 032 033 034 040 041 042 043 044
  045 046 050 051`. Zero duplicates, consistent 3-digit zero-padding, no gap within any band.
  The banding is by section (`B.0`→00x, `B.1`→01x/02x, `B.2`→03x, `B.3`→04x, `B.4`→05x). PASS
  because the banding is systematic and no requirement is lost to an accidental gap, which is the
  failure mode MP-1 exists to catch. The first auditor reached the same conclusion by a stronger
  route — it verified the convention against sibling SPECs `SPEC-WEB-CONSOLE-013/014`. I did not
  perform that check; I accept its result as corroboration, and note it makes the ruling firmer
  than my own reasoning alone did.
- **[PASS] MP-2 GEARS format compliance.** Judged against the **requirement layer** (`spec.md` §B
  `REQ-XXX` entries) only. All 25 match a GEARS pattern: ubiquitous (`001`, `002`, `010`, `020`,
  `021`, `022`, `025`, `030`, `031`, `032`, `040`, `042`, `043`, `044`, `046`, `050`), event-driven
  `When` (`011`, `012`, `023`, `033`, `034`, `041`, `051`), state-driven `While` (`045`), `Where`
  (`024`). `acceptance.md`'s Given-When-Then entries are the correct verification-layer format and
  were graded under Group 4, never here. One pattern misuse recorded as F13.
- **[PASS] MP-3 YAML frontmatter validity.** All 12 canonical fields present with correct types
  (`spec.md:2-13`): `id`, `title`, `version: "0.1.0"` (quoted semver), `status: draft`,
  `created: 2026-08-24`, `updated: 2026-08-24`, `author`, `priority: P2`, `phase`,
  `module: internal/web`, `lifecycle: spec-anchored`, `tags` (comma-separated string). No rejected
  snake_case alias (`created_at` / `updated_at` / `labels` / `spec_id`). Extras `era`, `tier`,
  `related_specs` are additive and well-formed. (The stale `version` value is a hygiene defect
  graded under Completeness as F8, not an MP-3 failure — the field is present and correctly typed.)
- **[N/A] MP-4 language neutrality.** Scoped to one Go repository's internals plus two doctrine
  markdown mirror pairs; names no per-language tooling. N/A auto-passes.
- **[PASS] MP-5 D7 cross-SPEC reconciliation.** Four referenced SPECs, all resolvable, none in
  `{retired, superseded, archived}`: `SPEC-KANBAN-TODO-CLI-001: in-progress`,
  `SPEC-WEB-CONSOLE-REDESIGN-001: completed`, `SPEC-HANDOFF-THRESHOLD-001: completed`,
  `SPEC-FACTORY-WORKER-FANOUT-001: implemented`. No BLOCKING finding.
- **[PASS] MP-6 D8 cross-platform discipline.** `grep -c 'syscall' spec.md` returns `0`. Auto-pass.
- **[PASS] MP-7 clarification gate.** `grep -rn 'NEEDS CLARIFICATION' .moai/specs/SPEC-WEB-CONSOLE-015/`
  returns rc=1, zero matches, across the whole directory including `plan.md` and `research.md`.
  **This was measured directly on `cf131b20a` before I saw any prior report.** The FAIL recorded at
  line 22 of the first report belongs to its iteration-1 block (`f11dd5cb3`) and is not the current
  state; the two audits agree that iteration 2 passes MP-7.

## Findings

Numbered `F1…F13` in this report's own sequence. They are **not** the first report's `D1-D16` and
carry no correspondence to them.

### F1 — REQ-WC15-011's named producer does not produce model and effort, and four of the eight call sites have no producer at all

`spec.md:148-150`, `plan.md:82-84`, `acceptance.md:35-38` — Severity: **major** — Class: **blocking**
— *Not raised by the first audit in either iteration.*

REQ-WC15-011 requires the launcher to record "the model and the effort resolved for the session at
launch". `plan.md` M2 prescribes the producer: *"Model and effort resolve from the existing
`internal/config/profile.go` `ModelEffort` / `EffectiveProfile` surface at launch."* AC-WC15-011
asserts equality against *"the values `EffectiveProfile` resolved for that session"*.

Measured, that surface does not exist in the shape the requirement needs:

```
internal/config/profile.go:73   type ModelEffort struct { Model string; Effort string }
internal/config/profile.go:92   func (l LLMConfig) EffectiveProfile() string
```

`EffectiveProfile` returns a **profile name** (`high` / `medium` / `low`), not a `{model, effort}`
pair. `ModelEffort` at `:73` is a bare type declaration with no resolver in that file; the per-agent
resolution lives in a different package (`internal/template/profile_matrix.go:346,407`). AC-WC15-011
is therefore not executable as written — there is nothing named `EffectiveProfile` that resolves two
values to compare against.

Second, and the more consequential half: `ModelEffort`'s own doc comment scopes its vocabulary to
Claude — *"Model is a Claude Code short alias (opus/sonnet/fable/inherit)"* (`profile.go:70-72`).
Four of the eight call sites M2 must widen are GLM launches (`internal/cli/glm.go:224,237,250,264`,
each passing `kanban.BackendGLM`), where the model actually in play is a z.ai model; the tree
already carries a separate resolution for that (`internal/statusline/metrics.go:192-232`
`resolveGLMModelName`). Writing a Claude alias into a GLM-backed session's record would put wrong
data into a field whose entire justification is honesty about what was recorded. Nothing in
`spec.md`, `plan.md`, or `design.md` addresses GLM model resolution, and AC-WC15-011 exercises only
`moai cc`.

This is the SPEC's own §A.2 claim carried forward without re-checking: §A.2 says *"The producers of
model and effort exist (`internal/config/profile.go:73-76` `ModelEffort`; `internal/settings/agentfm`
`AgentInfo`), they are simply not threaded into `recordKanbanSession`"*. The first clause is true of
`AgentInfo` and false of the `profile.go` surface as a per-session resolver, and the plan built M2
on the false half.

**Required fix:** name the producer that actually resolves `{model, effort}` per session and per
backend; state the GLM branch explicitly in REQ-WC15-011 or M2; rewrite AC-WC15-011 against the real
resolver and add a `BackendGLM` Given.

### F2 — AC-WC15-012 observes nothing: it already passes on the pre-implementation tree

`acceptance.md:40-42` (and `spec.md:151-153`) — Severity: **major** — Class: **blocking**
— *Not raised by the first audit in either iteration.*

AC-WC15-012: *"Given a `kanban.Record` whose model field is empty, When the role view model is built
and rendered, Then the model cell carries the 'not recorded' marker and the rendered fragment
contains no empty `<td>` for that column."*

Both halves are already true at `cf131b20a`:

```
internal/web/screens.templ:165-175
    if r.Model != "" { <span class="mono">{ r.Model }</span> } else { @missing() }
    ... identical shape for r.Effort
internal/web/widgets.templ:122-124
    templ missing() {
        <span class="missing" title="not recorded anywhere yet (kanban.Record extension required)"
              aria-label="not recorded">—</span>
    }
```

The "not recorded" marker is implemented, wired, and rendered today, and `viewmodel_ops.go:253`
hardcodes `Model: ""`, so the AC's Given is the *current* state of every row. The second clause is
worse than redundant: this view is `div`/`span`-based and emits no `<td>` at all, so "contains no
empty `<td>`" is trivially true under every possible implementation.

This is the exact failure shape the caller asked me to hunt, and the document demonstrates elsewhere
that its authors know it: AC-021, AC-022, AC-024, AC-035, and AC-040 each state a measured baseline
precisely so a passing result is new information. AC-WC15-012 states none, and it is the one that
needed it. REQ-WC15-012 is correspondingly a no-op — the tree already satisfies it.

**Required fix:** drop REQ/AC-WC15-012 as already-satisfied, or sharpen the requirement to the
distinction the new fields actually introduce (an absent field versus a recorded-empty one) and give
the criterion a measured baseline and an assertion against markup this view emits.

### F3 — the M4 resolution/adoption split fixes a write and introduces a read divergence

`spec.md:184-192`, `plan.md:127-139`, `design.md` §3 — Severity: **major** — Class: **blocking**
— *Second-order consequence of the first audit's iter1 D5 fix; not examined when that fix was accepted.*

REQ-WC15-031 requires the console's exported entry point to "perform no filesystem mutation on any
branch", with adoption reachable only from the `moai todo` path. `design.md` §3 fixes the shape: a
pure resolver for `internal/web`, an adopt-then-resolve entry point for `internal/cli/todo.go`. That
correctly closes the write hazard the first audit raised.

What it opens is not addressed anywhere. Measured (`internal/cli/todo.go:89-139`):
`fallbackTodoQueueRoot` returns `~/.moai/todo/<key>` **and calls `adoptLocalTodoQueue`**, which moves
a pre-existing project-local `backlog.json` into that root. A *pure* resolver returns the same root
**without** adopting. So in a non-git launch context where the operator's cards still sit at the
project-local path:

- the console resolves to the fallback root, finds no file there, and renders an **empty queue**;
- `moai todo`, on the same tree, adopts first and reports **N cards**.

Which one the operator sees depends on which ran first, and the console keeps showing empty until
`moai todo` is run. That is the failure the SPEC opens with — *"30 queued cards on the primary,
'queue is empty' from a linked worktree"* (`spec.md:96-99`) — reappearing not as two implementations
but as one resolution with two behaviours.

AC-WC15-031c asserts only that the disk is untouched. It never asserts **what the console renders**
in that state, so a run can satisfy every criterion and still ship the divergence.

**Required fix:** state the intended behaviour. Either the pure resolver returns the project-local
root when the fallback root holds no queue (read-through, still no mutation), or the SPEC records the
divergence as an accepted limitation the way §C.6 records the stale-event one. Either way, extend
AC-WC15-031c to assert the rendered count under its own preconditions.

### F4 — the acceptance-criterion count exceeds the Tier L ceiling

`acceptance.md` (whole document) — Severity: **major** — Class: **blocking**
— *Also found by the first audit (its N1), classed optional there. See §Divergence.*

`.claude/rules/moai/workflow/spec-workflow.md:146-152` fixes a per-tier budget: Tier L carries a
requirement ceiling of 25 **and** an acceptance-criterion ceiling of 25, applied independently, and
states that exceeding either "is a signal to tier up or to split the SPEC, **not to relax the
budget**".

Measured: **25 requirements** (exactly at ceiling) and **29 acceptance criteria** — four over.
Collapsing lettered suffixes into their parents (`010a/b` → one, `031a/b/c` → one) still gives 26.

The first auditor found the same overage and proposed recording it as a governed deviation, on the
ground that the four extra criteria were each added under plan-audit direction and that the ceiling
exists to stop *author-driven* over-formalization. That reasoning is sound about intent, and I would
accept it from the operator. I cannot accept it from the SPEC, because "record the deviation" **is**
relaxing the budget, which is the one response the [HARD] clause names and forbids — the actor
granting the exception would be the actor bound by it.

The compounding problem is the requirement count sitting exactly at 25: F5 and F7 identify surfaces
that need requirements and there is no budget to add them.

**Required fix:** either an explicit operator-approved deviation (which closes this in one line and
is a legitimate outcome), or a split. The natural seam is the one `plan.md` §C already considered:
M3 carries its own requirements (020-025), its own criteria (020-025, 052), and its own consumer
set. Lifting it restores headroom on both budgets and creates room for F5 and F7.

### F5 — a documented 12-file consumer surface has no requirement, no criterion, and no DoD entry

`spec.md:273`, `plan.md:99` (M3 step 7), `design.md` §2 step 7 — Severity: **major** — Class: **blocking**
— *Not raised by the first audit; its iter1 D6 covered the doctrine-file enumeration, which iteration 2 did close.*

`spec.md` §C.3 lists the docs-site as a consumer of the path being moved: *"`content/{en,ko,ja,zh}/
advanced/statusline.md`, `advanced/token-budget.md`, `cli-reference/tokens.md` reference the path —
12 files, four locales"*. Verified exactly:

```
$ grep -rln "context-usage.json" docs-site/content | wc -l
12
```

The plan sweeps them (M3 step 7) and `design.md` §2 step 7 repeats it. But no requirement covers
them: REQ-WC15-024 names only the two doctrine mirror pairs, and AC-WC15-024's grep is scoped to
`.claude internal/template/templates` — `docs-site/` is outside it. The Definition of Done
(`acceptance.md:234-243`) lists the mirror-pair diffs and omits the docs-site entirely.

A run can therefore satisfy every requirement, pass every criterion, clear the DoD, and leave twelve
published pages documenting a path that no longer exists — in a repository carrying a [HARD]
four-locale documentation synchronisation obligation.

The gap is structurally identical to the one the first audit caught at iteration 1 (`-detail.md`
unenumerated ⇒ survives stale). That one was closed by adding the files to §C.3 **and** widening
REQ-024's scope **and** widening AC-024's grep. Here only the first of the three was done.

**Required fix:** extend REQ-WC15-024 to name the docs-site surface (budget permitting after F4),
widen AC-WC15-024's grep scope to include `docs-site/content`, and add the sweep to the DoD.

### F6 — REQ-WC15-034 states unconditionally a behaviour the SPEC records as false in a supported deployment

`spec.md:197-203` against `spec.md:328-333` — Severity: **moderate** — Class: **blocking**
— *Adjacent to the first audit's "Lead-Approved Deviation 2", which accepted the scope decision. See §Divergence.*

REQ-WC15-034: *"**When** the existing `kanban` live-refresh event fires, the `/todo` route's section
shall be re-fetched through the existing refresh path."* No qualification.

§C.6's residual-limitation paragraph then records that this is false when the console is served from
a linked worktree: `Hub.Watch` watches the *served* root while REQ-WC15-031 resolves the backlog to
the *primary checkout*, so a primary-checkout queue change fires no event, and the 30s fallback poll
does not engage because SSE is healthy.

Worktree launch is not a fringe case here — §A.4 makes it load-bearing: *"`moai web` can be launched
from inside a worktree, so it must use the same resolution"* is the entire justification for
REQ-WC15-031. The SPEC therefore declares a deployment supported, builds one requirement on it, and
states a second requirement that is known-false in it, without qualification.

I want to be precise about what I am and am not disputing. The lead ruled that widening the watched
paths is out of scope, and the first auditor accepted that ruling; **I accept it too** — the
reasoning is sound and the limitation is staleness, not corruption. My finding is narrower and
survives that ruling untouched: the *requirement text* does not carry the condition its own §C.6
establishes. A requirement that is false in a supported configuration should say so in the
requirement, not only in a constraints section three pages away.

I independently verified the mechanism and the SPEC's analysis is correct on its own terms:
`watchMap["kanban"]` is `.moai/state/kanban` (`internal/web/events.go:30`), the backlog file is
inside it, and `eventFor` (`events.go:168-183`) resolves by longest watch path so `kanban` beats
`session` for that directory. The defect is the wording, not the analysis.

**Required fix:** qualify REQ-WC15-034 ("**While** the console is served from the checkout that holds
the resolved backlog…"), or cite the §C.6 limitation from the requirement body.

### F7 — an imperative correctness rule sits in a non-normative section with no requirement and no criterion

`acceptance.md:225-232` (§F Edge cases) — Severity: **moderate** — Class: **optional**

§F carries three edge cases, one stated as a hard rule rather than a note: *"Two lanes whose registry
entries carry the same PID (a stale entry not yet pruned) — the join **must not** silently attribute
one session's record to both lanes."*

That is a correctness constraint on REQ-WC15-043's join and the hazard is reachable:
`LoadFactoryRegistry` is fail-open (`internal/kanban/factory_slots.go:55`) and
`PruneFactoryDeadClaims` (`:84`) is a separate call the console is not required to make. Nothing in
§B requires the behaviour and no criterion observes it.

The document itself demonstrates the correct handling: AC-WC15-052 was *promoted* out of §F for
exactly this reason (*"Promoted from a §F edge-case note because … a write boundary taking outside
input earns an assertion rather than prose"*). The same argument applies to a join that can
mis-attribute a session; it was not applied.

Classed optional only because the requirement budget is full (F4); it becomes blocking once a split
frees headroom.

**Required fix:** promote the duplicate-PID rule to a requirement and a criterion, or state
explicitly that the SPEC accepts mis-attribution on a stale registry.

### F8 — the revision is unversioned

`spec.md:5` (`version: "0.1.0"`), `spec.md:23-25` (HISTORY) — Severity: **minor** — Class: **optional**
— *Independently found by the first audit (its N2). I missed it on my first pass and adopt it here.*

All four authored artifacts were substantively rewritten between `f11dd5cb3` and `449736deb`, and a
fifth (`design.md`) was created, while `version` remains `"0.1.0"` and HISTORY carries its single
original row. The frontmatter field `updated: 2026-08-24` is unchanged too, though it happens to
remain accurate because both iterations landed the same day.

**Required fix:** bump to `0.2.0` and add one HISTORY row naming the iteration-2 revision.

### F9 — AC-WC15-002 describes a grep that cannot express its own qualifier

`acceptance.md:15-19` — Severity: **minor** — Class: **optional**

*This finding is a correction of my own. My first draft graded it critical, asserting the stated
baseline was false. I then tested that assertion against the tree and disproved it. The corrected
finding is materially smaller, and the correction is recorded here rather than silently dropped
because the first version would have driven a wrong fix.*

The criterion reads: *"When `internal/web` is grepped for `Mutate(`, `WriteBestEffort(`,
`acquireLock`, `SaveFactoryRegistry`, and `os.WriteFile` **against any `.moai/state` path**, Then
every hit is zero. Baseline: the same grep on the pre-change tree also returns zero."*

A literal grep for those five tokens does **not** return zero today:

```
internal/web/server.go:9          // yaml.Marshal/os.WriteFile in the web layer — ...
internal/web/projectconfig.go:158 // YAML을 직접 marshal/os.WriteFile 하는 것은 금지된 안티패턴 ...
internal/web/projectconfig.go:221 // ... No direct yaml.Marshal/os.WriteFile.
```

But the qualifier "against any `.moai/state` path" is a semantic filter, and under it the baseline is
**correct** — verified:

```
$ grep -rn "moai/state" internal/web/ | grep -v _test.go | grep -i "writefile\|mkdir\|rename\|create"
rc=1   (no matches)
```

There is no non-test write to a `.moai/state` path in `internal/web`. So the criterion is true about
the property it intends and false about the command it names — no grep expresses "against any
`.moai/state` path", which makes the check human-interpreted rather than mechanical. The
inconsistency is visible against its own sibling: AC-WC15-022 explicitly pins its scope ("*including
`_test.go` files*") and names the exact command; AC-WC15-002 pins nothing.

**Required fix:** name a command whose plain output is the assertion — for example an `ast-grep`
pattern for write calls whose path argument derives from the state dir, or a grep restricted to
non-test files with the state-path co-occurrence expressed explicitly.

### F10 — AC-WC15-025's removal-half baseline is incomplete, and two cited line numbers point at comments

`acceptance.md:96-100`, `spec.md:169-170` — Severity: **minor** — Class: **optional**

The criterion states: *"Baseline: on the pre-change tree the same grep returns `tokens.go:30` (the
constant) and `tokens.go:79` (the struct)"*. Measured, `grep -rn 'context-usage' internal/cli/`
returns four lines:

```
internal/cli/tokens.go:30        const tokensContextSnapshotFilename = "context-usage.json"
internal/cli/tokens.go:79        // tokensContextSnapshot is the subset of the statusline context-usage.json
internal/cli/tokens.go:426       "... embed the context-usage snapshot when present, and append one JSON "
internal/cli/tokens_test.go:284  os.WriteFile(filepath.Join(stateDir, "context-usage.json"), ...)
```

`tokens.go:426` is a command help string the change does not remove, so the criterion's literal first
sentence survives only via its "specifically" qualifier. Separately `:79` is the struct's **doc
comment**; the declaration is at `:81`, and `spec.md:169-170` repeats the same off-by-two.

Worth noting for the delta record: the first audit's iteration-2 table marks this criterion RESOLVED
citing "stated baseline tokens.go:30/:79" — it accepted the baseline as written rather than
re-running that particular grep. The adjacent AC-022 baseline it *did* re-run verbatim, and that one
is exactly right.

**Required fix:** restate the baseline as measured (four hits, two out of scope and why); correct
`:79` → `:81` in both documents.

### F11 — two criteria name no executable check

`acceptance.md:125-127` (AC-WC15-031b), `acceptance.md:187-190` (AC-WC15-043) — Severity: **minor** — Class: **optional**
— *AC-043 also flagged by the first audit as its sole iteration-2 Testability deduction. Convergent.*

AC-WC15-031b: *"When `internal/cli/todo.go` **is read**, Then its queue-root resolution delegates to
the exported `internal/kanban` function and declares no git-common-dir resolution of its own."* "Is
read" is a human judgement. A mechanical form exists and the document uses it elsewhere: grep
`internal/cli/todo.go` for `gitcore.ResolveGitDirs` and assert zero.

AC-WC15-043's second clause: *"and When the merged tree **is grepped**, Then no new file is created
under `.moai/state/` by this join."* Grepped for what? A grep cannot observe file creation; the
property wants a directory-listing assertion around the join call.

**Required fix:** give both a command and an expected result.

### F12 — a hard-coded English note banner that the change falsifies appears in no enumeration table

`internal/web/screens.templ:192` — Severity: **minor** — Class: **optional**

```
@noteBanner("info", "Stage is estimated from heartbeat. Model, effort and context usage are not
recorded yet, so they are left blank — they fill in once kanban.Record is extended.", "")
```

M1-M3 make that sentence false, so the banner must change. It appears in neither §C.3 nor §C.6 — the
two tables the SPEC presents as exhaustive enumerations of where the change lands — nor in any
requirement or criterion. Its third argument is the empty i18n key, so it is also an untranslated
user-visible string that REQ-WC15-050's four-locale rule would pull into scope once touched.

**Required fix:** add the banner to §C.3's table and name it in M5.

### F13 — minor form defects

Severity: **minor** — Class: **optional**

- **REQ-WC15-024 uses `Where` for a temporal condition** (`spec.md:163`). GEARS reframes `Where` as a
  capability gate / feature flag / static configuration; "once the path is adopted" is `When`.
- **REQ-WC15-025 prescribes implementation rather than behaviour** (`spec.md:169-175`): two symbols,
  three line numbers, a target file, and a four-line rationale paragraph inside a requirement body.
  It belongs in `plan.md` M3 / `design.md` §2 step 3, where it is already stated more clearly;
  REQ-WC15-022 already carries the behavioural form. Milder instances: REQ-030 (route string),
  REQ-031 (source file, package relocation), REQ-040 (Go type), REQ-050 (`i18n.js` and its maps) —
  several of these are defensible as deliberate constraints, since the SPEC's thesis is that the
  implementation detail *is* the risk.
- **REQ-WC15-042's override environment-variable name is unspecified.** Run-phase detail; the repo's
  hardcoding rule will require the constant in `internal/config/envkeys.go`. Flagged only so the
  implementer does not inline a string. *(Independently found by the first audit as its N4; adopted.)*
- **`g4-scope-ruling.md:19` cites `events.go:29`; the entry is at `:30`** (`spec.md` §C.6 has it
  right). *(Independently found by the first audit as its N3; adopted.)*

## Divergence from the first audit

Two independent audits of `cf131b20a` returned **PASS 0.94** and **FAIL 0.81**. The gap is worth
more than either verdict alone, so this section is the analysis rather than a footnote.

### The two audits agree about the tree

On every fact both of us measured, we agree: 25 requirements; 29 criteria; all 12 frontmatter fields
valid; four related SPECs at non-terminal status; zero `syscall`; zero clarification markers at
`cf131b20a`; AC-022's pinned grep returning exactly four hits in the four listed files; AC-021's
exported-reader baseline of zero; `rail()` carrying five `navRow`s; the `eventFor` longest-match tie
hazard; the `watchMap`/`EVENTS` six-entry symmetry. We also independently reached the same reading of
AC-043's second clause. There is no factual dispute of any consequence between the two reports.

### Most of the gap is the scope of the audit, not a disagreement about quality

The first audit scored iteration 2 as a **delta re-audit**, which the retry-loop contract explicitly
authorizes: *"On iteration 2+: the re-audit is scoped to the enumerated defect delta from the previous
iteration's report, plus a regression check."* Its Clarity 1.0, Completeness 1.0, and Traceability 1.0
are best read as *"the iteration-1 defects are thoroughly closed"* — and on that question it is right.
I verified the D1-D14 closures it claims and did not find one overstated. The iteration-2 work is
genuinely good: the resolved-decision records in `plan.md` §G, the measured §C.6 event-vocabulary
argument, the AC-031c fallback-branch construction, and the corrected AC-022 baseline are all above
the repository median.

I was dispatched to audit **from scratch on the artifacts' own merits**, so my scores answer a
different question: *"how good is this SPEC in absolute terms?"* A delta audit cannot surface a defect
that was never in the delta, and that is precisely where my three new blocking findings live — F1
(REQ-011's producer), F2 (AC-012 vacuous), F3 (the second-order consequence of the D5 fix). None of
the three was in the iteration-1 defect list, so a contract-compliant delta audit would not have
looked at them. F5 (docs-site) is a near-miss of the same kind: iteration 1 caught the doctrine-file
enumeration gap, iteration 2 closed it properly, and the structurally identical docs-site gap sitting
in the same table went unexamined because it was not in the delta.

**This is not a criticism of the first audit's method.** It is the known cost of delta scoping, and it
is the reason a second full pass returns value on the same commit.

### Three genuine disagreements

1. **The AC ceiling (my F4 vs its N1).** Both found 29 > 25. It classed the overage optional and
   proposed recording a governed deviation; I class it blocking because the [HARD] clause names
   "record the deviation" as the response it forbids, and the SPEC cannot grant itself the exception.
   The practical distance is small: an operator-approved deviation closes it in one line and I would
   accept that.
2. **AC-WC15-021's mechanical pin (its iter1 D9 → my F13-adjacent reasoning).** The first audit asked
   for the `^func Read.*ContextUsage` count pin and marked it resolved. I initially graded that pin a
   defect — it measures a naming convention REQ-021 never states, so `LoadContextUsage` would satisfy
   the requirement and fail the criterion. On reflection I do not carry it as a finding: a
   name-shaped pin is better than the unpinned claim it replaced, the first audit asked for it
   explicitly, and the residual risk is that a future rename silently breaks a criterion — real but
   small. Recorded here rather than in the findings list so the disagreement is visible without
   inflating the defect count. *(This is the second place in this report where I withdrew a finding
   after testing it; the other is F9.)*
3. **REQ-WC15-034's wording (my F6 vs its Deviation-2 acceptance).** We agree on the substance — the
   watch-widening exclusion is sound and I accept the lead's ruling. The first audit's note that "no
   AC falsely claims live liveness for the worktree case" is correct about the criteria. My finding is
   about the requirement text, which the deviation review did not examine.

### Methodological note

The first report's iteration-2 aggregate of **0.94** is the arithmetic mean of (1.0, 1.0, 0.75, 1.0).
The harmonic mean of the same four scores is `4 / (1 + 1 + 1.333333 + 1) = 0.923`. Both clear 0.85, so
the verdict is unaffected — but the skeptical-evaluation contract specifies the harmonic mean
precisely because it refuses to let three perfect dimensions mask one weak one, and the difference
grows as scores spread. Flagged for the method, not for this verdict.

### What I would tell the orchestrator

Do not read this as "the first audit was wrong". Read it as: the delta audit correctly certified that
iteration 2 closed iteration 1's defects, and a full pass on the same commit found four blocking
defects that were never in that delta. Both statements are true simultaneously. The actionable output
is F1-F5, not the 0.81.

## Evidence

### Claim
The verdict, the seven must-pass results, and every finding above rest on commands run in this
worktree during this audit.

### Evidence (command + verbatim output)

```
$ git rev-parse --short HEAD
cf131b20a

$ git status --short
(empty)

$ git diff --stat 28bde4022..HEAD
 .moai/reports/t207/g4-scope-ruling.md          |  58 ++++
 .moai/specs/SPEC-WEB-CONSOLE-015/acceptance.md | 243 ++++++++++++++++
 .moai/specs/SPEC-WEB-CONSOLE-015/design.md     | 207 ++++++++++++++
 .moai/specs/SPEC-WEB-CONSOLE-015/plan.md       | 293 +++++++++++++++++++
 .moai/specs/SPEC-WEB-CONSOLE-015/progress.md   |  24 ++
 .moai/specs/SPEC-WEB-CONSOLE-015/research.md   | 153 ++++++++++
 .moai/specs/SPEC-WEB-CONSOLE-015/spec.md       | 375 +++++++++++++++++++++++++
 7 files changed, 1353 insertions(+)
```

Counts (F4, traceability):

```
$ grep -c '^- \*\*REQ-WC15-' spec.md
25
$ grep -c '^\*\*AC-WC15-' acceptance.md
29
```

F1 — the named producer:

```
$ grep -n "func EffectiveProfile\|type ModelEffort" internal/config/profile.go
73:type ModelEffort struct {

$ grep -rn "EffectiveProfile" internal | grep -v _test.go
internal/config/profile.go:92:func (l LLMConfig) EffectiveProfile() string {
internal/web/schemaform.go:378:	activeProfile := cfg.LLM.EffectiveProfile()
internal/web/agentfm.go:457:		baseProfile = llm.EffectiveProfile()
internal/template/profile_matrix.go:346:	profile := cfg.EffectiveProfile()
internal/template/profile_matrix.go:407:	profile := cfg.EffectiveProfile()
internal/cli/model.go:91:		Profile: llm.EffectiveProfile(),
(plus 5 comment lines)

$ sed -n '70,76p' internal/config/profile.go
// ModelEffort carries a {model, effort} assignment for one agent or one
// profile group cell. Model is a Claude Code short alias (opus/sonnet/fable/
// inherit); Effort is a reasoning effort level (low/medium/high/xhigh/max).
type ModelEffort struct {

$ grep -n "recordKanbanSession" internal/cli/glm.go
224:		recordKanbanSession(entry.Spec, kanban.BackendGLM, kanban.RoleLead)
237:		recordKanbanSession(entry.Spec, kanban.BackendGLM, kanban.RoleLane)
250:			recordKanbanSession(entry.Spec, kanban.BackendGLM, kanban.RoleLead)
264:			recordKanbanSession(entry.Spec, kanban.BackendGLM, companionRole(finalLabel))
```

F2 — the marker already exists:

```
$ sed -n '164,175p' internal/web/screens.templ
        <span data-i18n="kanban.model">Model</span>
        if r.Model != "" {
                <span class="mono">{ r.Model }</span>
        } else {
                @missing()
        }
        <span>effort</span>
        if r.Effort != "" {
                <span class="mono">{ r.Effort }</span>
        } else {
                @missing()
        }

$ sed -n '121,124p' internal/web/widgets.templ
// missing — 그럴듯한 값을 채우지 않는다. 왜 없는지는 title 에 남긴다.
templ missing() {
        <span class="missing" title="not recorded anywhere yet (kanban.Record extension required)" aria-label="not recorded">—</span>
}

$ grep -n '3단계' internal/web/viewmodel_ops.go
253:			Model:          "", // 3단계: Record 에 모델 스냅샷이 추가되면 채운다
254:			Effort:         "", // 3단계
255:			ContextPct:     -1, // 3단계: context-usage/<session-id>.json 분리 후
```

F3 — the adoption branch:

```
$ grep -n "os.MkdirAll\|os.Rename\|os.WriteFile\|func adoptLocalTodoQueue\|func fallbackTodoQueueRoot\|func resolveTodoQueueRoot" internal/cli/todo.go
66:func resolveTodoQueueRoot() string {
89:func fallbackTodoQueueRoot(base string) string {
115:func adoptLocalTodoQueue(base, fallbackRoot string) {
124:	if err := os.MkdirAll(fallbackRoot, 0o755); err != nil {
128:	if err := os.Rename(local, target); err == nil {
139:	_ = os.WriteFile(target, data, 0o600)

$ grep -n "func BacklogPathForRoot" -A2 internal/kanban/backlog_store.go
249:func BacklogPathForRoot(root string) string {
250-	return filepath.Join(root, ".moai", "state", "kanban", "backlog.json")
```

F5 — the docs-site surface:

```
$ grep -rln "context-usage.json" docs-site/content | wc -l
12
```

F6 / §C.6 mechanism (the SPEC's analysis verified correct):

```
$ sed -n '24,32p' internal/web/events.go
var watchMap = map[string][]string{
	"spec":    {".moai/specs"},
	"session": {".moai/state"},
	"goal":    {".moai/state/goal"},
	"verify":  {".moai/state/verify"},
	"kanban":  {".moai/state/kanban"},
	"config":  {".moai/config/sections"},
}
(eventFor at :168-183 selects by longest watch path; header comment :6-9 records the
 map-iteration-order bug the SPEC cites)
```

F9 — the correction that reversed my own strongest finding:

```
$ grep -rn "Mutate(\|WriteBestEffort(\|acquireLock\|SaveFactoryRegistry\|os.WriteFile" internal/web/ | grep -v _test.go
internal/web/server.go:9:// yaml.Marshal/os.WriteFile in the web layer — and no parallel validation rule
internal/web/projectconfig.go:158:// YAML을 직접 marshal/os.WriteFile 하는 것은 금지된 안티패턴(REQ-WC3-008). ...
internal/web/projectconfig.go:221:// every other section's content unchanged. No direct yaml.Marshal/os.WriteFile.
   -> the literal grep is non-zero (three comment hits)

$ grep -rn "moai/state" internal/web/ | grep -v _test.go | grep -i "writefile\|mkdir\|rename\|create"
rc=1
   -> but no write to a .moai/state path exists, so the criterion's INTENDED
      property, and its stated baseline, are correct. Finding downgraded
      critical -> minor accordingly.
```

F10 — the `internal/cli` grep:

```
$ grep -rn 'context-usage' internal/cli/
internal/cli/tokens_test.go:284: ... "context-usage.json" ...
internal/cli/tokens.go:30:const tokensContextSnapshotFilename = "context-usage.json"
internal/cli/tokens.go:79:// tokensContextSnapshot is the subset of the statusline context-usage.json
internal/cli/tokens.go:426: "... embed the context-usage snapshot when present, ..."

$ grep -n "type tokensContextSnapshot" internal/cli/tokens.go
81:type tokensContextSnapshot struct {
```

Baselines I verified **correct** (a clean result is evidence too):

```
$ grep -rn '"raw_pct"' internal/                       # AC-022: 4 hits / 4 files, 2 outside statusline
internal/statusline/context_usage.go:63
internal/statusline/context_usage_test.go:150
internal/cli/tokens.go:86
internal/cli/tokens_test.go:283

$ grep -rln "state/context-usage.json" .claude internal/template/templates   # AC-024: exactly 4
.claude/rules/moai/workflow/context-window-management-detail.md
.claude/rules/moai/workflow/context-window-management.md
internal/template/templates/.claude/rules/moai/workflow/context-window-management-detail.md
internal/template/templates/.claude/rules/moai/workflow/context-window-management.md

$ grep -rn "^func Read.*ContextUsage" internal/statusline/ | wc -l           # AC-021 baseline: 0
0
$ grep -n "func readContextUsage" internal/statusline/context_usage.go
186:func readContextUsage(path string) (*contextUsageRecord, error) {

$ grep -n "navRow" internal/web/shell.templ                                  # AC-035 baseline: five rows
130,131,132,133,134  (overview, kanban, specs, monitor, settings)

$ grep -n 'case "' internal/web/icons.templ                                  # no todo case
70: case "overview":  72: case "kanban":  74: case "specs":  76: case "monitor":  78: case "settings":

$ grep -rn 'NEEDS CLARIFICATION' .moai/specs/SPEC-WEB-CONSOLE-015/ | wc -l   # MP-7
0    (rc=1, no matches)
$ grep -c 'syscall' .moai/specs/SPEC-WEB-CONSOLE-015/spec.md                 # MP-6
0
$ grep -m1 '^status:' <the four related SPECs>                               # MP-5
SPEC-KANBAN-TODO-CLI-001: in-progress
SPEC-WEB-CONSOLE-REDESIGN-001: completed
SPEC-HANDOFF-THRESHOLD-001: completed
SPEC-FACTORY-WORKER-FANOUT-001: implemented
```

Line-number spot-check of the SPEC's own citations — all confirmed accurate at `cf131b20a`:
`viewmodel_ops.go:46` (`ChainRoles`), `:253-255`, `record.go:45` (`@MX:ANCHOR`), `:121` (`WithRole`),
`role.go:42` (`RoleLane`), `kanban.go:472` (`recordKanbanSession`), `cc.go:161,175,192,208`,
`glm.go:224,237,250,264`, `context_usage.go:56,63,128,134,186,203,216,236,255`,
`registry.go:86-95` (carrying `PID`), `factory_slots.go:37,47,55,84,99`,
`todo.go:66,89,115,124,128,139`, `screens.go:23`, `i18n.js` locale maps at `27/646/1267/1888` with
`nav.*` at `30/649/1270/1891`. `internal/statusline/{builder,context_usage}_test.go` carry 10
literal-path hits (the SPEC claims "at least five" — true). `internal/web/i18n_governance_test.go`
and `i18n_untranslated_allowlist_test.go` both exist.
`.moai/reports/webredesign/moai-web-menu-spec.md` exists.

### Baseline-attribution
Every command above was run **in this run**, in this tree
(`/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t207`), on branch `WT-web-live-todo` at HEAD
`cf131b20a`, working tree clean. The `git diff --stat 28bde4022..HEAD` output establishes that no Go
source changed between the SPEC's own stated measurement base (`28bde4022`, `research.md:3`) and this
HEAD, so the SPEC's baselines and mine are attributable to the same code. No figure was carried over
from another tree, package, or point in time.

**Contamination disclosure (the caller's explicit question).** I chose to **read the first auditor's
report, and only after completing my own audit end to end** — full artifact read, all verification
commands, all findings written, scores computed, and a complete report already written to disk at
`.moai/reports/t207/plan-audit.md` before the path change. The order is verifiable from that prior
file, which predates my reading of the first report and contains F1-F13 in substance. What changed
afterwards: (a) I adopted two findings I had genuinely missed, crediting them (F8 version hygiene,
and two items in F13); (b) I **withdrew one finding and downgraded another** after testing claims the
first report's existence prompted me to re-check — F9's downgrade from critical to minor came from my
own follow-up measurement, not from the other report, which never examined AC-002; (c) I added the
§Divergence section. No score moved because the first report said 0.94; the one score that moved
(Testability 0.72 → 0.78) moved because I disproved my own finding, and it moved the aggregate the
wrong way for agreement — 0.80 → 0.81, still FAIL.

`mcp__moai__spec_audit` was not invoked; every check is a direct filesystem or git read from this
worktree, with `project_root` moot because no MCP tool was called.

### Gaps
Explicitly **not** observed:

1. **No code was compiled or executed.** `go build`, `go vet`, `go test` were not run. F1's claim that
   `EffectiveProfile` cannot supply two values rests on the declared signature, not a failed build.
2. **The docs-site content was not read**, only counted by grep (12 files). Whether all twelve need
   substantive edits or only some was not determined.
3. **The two screenshots attached to card t207 were not opened**, consistent with the SPEC's own §8
   and §D. I did not verify they contain nothing requirement-bearing.
4. **`.moai/reports/webredesign/moai-web-menu-spec.md` was confirmed to exist but not read.** The
   SPEC's citations to its §4.6 / §6.1 are unverified by me; the first audit reports verifying them
   (its D15) and I have not independently confirmed that.
5. **No cross-model audit backend was invoked** (`audit_multi` / `codex_audit` / `glm_audit`). The
   first audit attempted this and recorded both backends as off-target no-ops, so neither audit
   carries a genuine second opinion from another model — the second opinion here is a second Claude
   auditor, which shares a model-class blind spot with the first.
6. **I did not verify the first report's D1-D14 closure claims item by item.** I read its delta table
   and spot-checked the closures that intersected my own findings (D2/AC-022, D5/M4 split, D6/§C.3,
   D7/design.md, D9/AC-021, D10/AC-052). Those six are genuinely closed. The remaining eight I did
   not audit.
7. **F1's producer question was answered from `internal/config` only.** I did not trace whether some
   other package already resolves a per-session `{model, effort}` pair valid for both backends; F1
   asserts the *named* surface does not, not that no such surface exists.
8. **The first audit's sibling-SPEC convention check (MP-1) was not reproduced.** I accept its
   `SPEC-WEB-CONSOLE-013/014` finding as corroboration without having run it.

### Residual-risk

- **Two Claude auditors are not two independent opinions in the strongest sense.** We share a
  model-class prior, and both cross-model backends failed off-target. Where we agree, that agreement
  is weaker evidence than it looks.
- **The score sits 0.04 under threshold and three dimension scores are judgements within a band.**
  Clarity 0.85 and Testability 0.82 together would reach 0.85. The verdict is not fragile to any
  single re-grade — I tested that by re-scoring after the F9 correction — but the blocking-finding
  list, not the decimal, is the actionable output. If the orchestrator disposes of F1-F5, the score
  question becomes moot.
- **F3 is an inference about behaviour that does not yet exist.** I read the current
  `adoptLocalTodoQueue` and reasoned about what a pure resolver returns. The SPEC's *silence* is
  certain; the concrete divergence is the most likely reading of that silence, not the only one.
- **F1's GLM half rests on a doc comment.** `ModelEffort`'s "Claude Code short alias" scoping is
  authoritative as documentation, but I did not trace every write into that field to confirm no GLM
  value ever lands there.
- **F4's disposition is a governance question I cannot settle.** Whether an operator-approved AC-budget
  deviation is acceptable is the operator's call; I graded the violation, not the remedy.
- **I have already been wrong once in this audit, in the direction of over-calling severity** (F9,
  critical → minor). The remaining findings were each re-checked against the tree, but that error rate
  is the honest prior to apply to my severity labels.
- **Line-number citations decay.** Every citation was accurate at `cf131b20a`; any commit touching
  `internal/` invalidates them.

## Recommendation

**FAIL at 0.81** against the Tier L threshold of 0.85, as a second independent judgment. The first
audit's **PASS 0.94** stands as a valid delta-scoped verdict; this report does not retract it and the
orchestrator now holds two verdicts on one commit.

No score regression: iteration 1 scored 0.75 and this pass scores 0.81, so the STOP-on-regression
clause does not fire under either auditor's numbers. This is iteration 2 of a maximum 3, so one
revision cycle remains.

Priority order — the first four are the ones a delta audit could not have found, and they are why
this second pass was worth running:

1. **F1** — name the producer that actually resolves `{model, effort}` per session and per backend;
   state the GLM branch; rewrite AC-WC15-011 against it. This is the only finding that changes M2's
   implementation rather than its description.
2. **F2** — drop or sharpen REQ/AC-WC15-012. The criterion passes today on the untouched tree.
3. **F3** — decide what the console renders on the non-git fallback branch and assert it.
4. **F5** — bring the 12 docs-site files inside REQ-WC15-024, AC-WC15-024's grep scope, and the DoD.
5. **F4** — obtain an explicit operator deviation for the 29 > 25 AC count, or split. Note that F5
   and F7 need requirement budget and the requirement count is exactly at 25, so a split is the
   response that solves three findings at once. `plan.md` §C already contains the M3 carve-out
   analysis; it was closed on release-lifecycle grounds (G-3 hard cut), which is a different question
   from budget, so reopening it on budget grounds is not a reversal of G-3.
6. **F6** — qualify REQ-WC15-034 with the condition §C.6 already establishes. The lead's scope ruling
   is accepted and unaffected; this is a wording fix.
7. **F8, F10** — one-line hygiene: bump `version` and add a HISTORY row; restate AC-025's baseline and
   fix `:79` → `:81`.

F7 and F9/F11/F12/F13 are optional — surfaced for the orchestrator's discretion, not required for a
re-audit. Routing all of them into a revision would add speculative requirements to a SPEC already at
its requirement ceiling, which is the over-engineering the simplicity mandate forbids.

A confirming re-audit should be scoped to F1-F6 plus a regression check, and should be run by
whichever auditor did not produce the change — the delta-scoping lesson in §Divergence applies to
this report exactly as much as it applied to the first one.

# Progress — SPEC-SVG-QUALITY-ABSORB-001

Branch `WT-svg-quality`, worktree `.claude/worktrees/t165`, base `a6fe13232`.

## §E.2 Run-phase Evidence

### Claim

All six milestones are complete. M1-M4 and M6 landed as text, checker, fixtures,
mirrors, and attribution. M5 produced one sample pair (`journey`) and put the
carve-out to the operator per SPEC §5; they declined it, so the exception list is
empty and `SKILL.md` § Step 0's routing table is unchanged. Every AC is closed —
AC-BUDGET with a stated estimate rather than a measurement (see Gaps).

### Evidence

Commands run in this worktree, with what they printed.

**AC-REQ-3b — the checker fails in the right direction (both directions run):**

```
$ node .../check-svg.mjs .../fixtures/a11y-present.svg
0 errors, 0 warnings                                              exit=0

$ node .../check-svg.mjs .../fixtures/a11y-missing.svg
...:3:1  error  SVG060  root <svg> has no role; ...
...:3:1  error  SVG062  root <svg> has no <title>; ...
...:3:1  error  SVG063  root <svg> has no <desc>; ...
...:3:1  error  SVG061  root <svg> has no aria-labelledby ...
4 errors, 0 warnings                                              exit=1
```

Two further directions checked ad hoc: a bare `title` / `desc` id pair produced
two `SVG064` errors (exit 1); an `aria-hidden="true"` root produced 0 errors
(exit 0), confirming the decorative exemption.

**AC-REQ-1 — six connector rules, each carrying its number:**

```
$ grep -nE '^\*\*C[1-6] ' references/authoring.md
159:**C1 — ... rounded right angle at `r = 8`.**
166:**C2 — A label's mask clears its own stroke by 6–10 units.**
180:**C3 — No two connectors share a path; separation is ≥ 12 units.**
193:**C4 — Connectors on a shared edge fan to distinct attach points.**
208:**C5 — A connector does not pass behind a box that is neither its source nor
     its destination — except where that box is geometrically unavoidable.**
226:**C6 — A label mask may not overlap a node painted after it.**
```

C4 carries `offset(k) = L * k / (N + 1)` with the `>= 12` floor; C5 carries the
transit exception verbatim in substance (dashed `4,3`, label at the visible end,
no arrowhead on the intervening edge).

**AC-REQ-2 — every archetype maps to a ceiling:** the four archetypes declared
in `archetypes.md` (`## A1`..`## A4`) each appear in the ceiling table, and the
table introduces none that does not exist.

**AC-REQ-4 — 14 entries:** `grep -cE '^\| [0-9]+ \|'` over the slop table
printed `14`.

**AC-REQ-5 — four dials, values and defaults:** format `svg+png`, size
`doc-inline`, detail `balanced`, audience `mixed`.

**AC-REQ-6a / 6b:** 11 semantic roles each carry a light and a dark value; the
inversion is stated once as a rule (section 3.0) rather than as a second
maintained table. The one-accent discipline survived the restructure —
`authoring.md:342` still reads "One diagram carries one focal element (two at
the absolute most)", and the `accent` row of the roles table repeats the cap.

**AC-REQ-8a — mirrors and build:**

```
$ make build
catalog.yaml updated successfully (12899 bytes)
go build -ldflags ... -o bin/moai ./cmd/moai                       exit=0

$ diff -r .claude/skills/moai-domain-svg-infographic \
          internal/template/templates/.claude/skills/moai-domain-svg-infographic
                                                                   exit=0

$ MOAI_TEMPLATE_LEAK_STRICT=1 go test ./internal/template/...
ok  github.com/modu-ai/moai-adk/internal/template  21.010s
```

**AC-REQ-8b — attribution:** present in `SKILL.md` § Attribution and pointed at
from `authoring.md` and `archetypes.md`; mirrored.

**AC-REQ-9 — no view-time fetch:**

```
$ grep -rnE 'fonts\.googleapis\.com|fonts\.gstatic\.com|@import|url\(\s*['"]?https?:' \
    <skill> <mirror>
                                                          (no output) exit=1
```

**AC-REQ-7a — the one sample pair, both renders committed:**

```
$ mmdc -i journey.mmd -o journey-mermaid.png -s 2 -b white
Generating single mermaid chart          (Google Chrome 151.0.7922.170)

$ node scripts/render.mjs journey-absorbed.svg --out journey-absorbed.png
png IHDR   2400x1140
verified   dimensions match the target                             exit=0

$ node scripts/check-svg.mjs journey-absorbed.svg
0 errors, 2 warnings                                               exit=0
```

The two `SVG030` warnings are on the legend text, where the heuristic measures a
label against its 10-unit colour swatch (the nearest preceding `<rect>`) rather
than a container. Triaged against the render as noise, not reflowed.

### Baseline-attribution

Every command above was run in this worktree against `WT-svg-quality`, base
commit `a6fe13232`, working tree as recorded by `git status --porcelain` at the
time of the run: four modified skill files, four modified mirrors, modified
`internal/template/catalog.yaml`, plus the untracked `scripts/fixtures/` pair
(and its mirror) and `.moai/reports/t165/samples/`.

### Gaps

(AC-REQ-7 is no longer a gap — see § REQ-7 decision below.)

- **AC-BUDGET is at its boundary, not comfortably inside it.** `SKILL.md` is
  19,094 bytes / 3,080 words after the additions, against a declared
  `level2_tokens: 5000`. No tokenizer was available in this environment, so the
  token count is an estimate (~4,800-5,400 by word-and-symbol count), not a
  measurement. Recorded as an estimate rather than reported as a pass.
- **The 14 slop entries were counted, not validated.** Each is phrased as
  something observable in a render; whether the set is the right fourteen rests
  on the surveyed source.
- **Per-archetype ceilings are inherited, not re-derived** against this skill's
  canvas widths. A1/A2/A4 come from the source's per-type budgets; A3 has no
  counterpart there and was derived from this file's own preset. Stated in
  `archetypes.md` where the table sits.
- **Full-suite verdict is CI's.** Only `./internal/template/...` was run locally,
  per the repo's local-test discipline.

### REQ-7 decision — the exception list stays empty

**Claim.** The mermaid bypass exception list is empty and `SKILL.md` § Step 0's
routing table is unchanged. AC-REQ-7a, 7b, and 7c are closed on that basis.

**Evidence.** One pair was produced (`journey`), rendered and committed under
`.moai/reports/t165/samples/`. The comparison and its observations are recorded
in that directory's `README.md`. `timeline` was not rendered and therefore could
not be listed whatever its merits; `quadrant` and `ER-schema` were dropped on
scope, because neither maps onto any of the four archetypes and adding one is
the type-catalogue path the SPEC rejects (§2 Out of Scope, plan D1).

The judgement itself is the operator's, per SPEC §5 ("The samples are judged by
a human"). It was put to them with the pair's observations and the durable cost
of a carve-out stated, and they chose to leave the list empty. The grounds given
were the ones the SPEC itself names: AC-REQ-7b states that an empty list passes,
the skill's own routing rule is that mermaid wins a tie, and one sample is thin
evidence for a permanent carve-out.

**Per-AC closure.**

- **AC-REQ-7a** — every listed type has both renders committed. Vacuously true
  on an empty list; the one pair that exists is committed regardless.
- **AC-REQ-7b** — no type appears without a committed pair. Satisfied: the list
  is empty, which the AC states explicitly is a pass.
- **AC-REQ-7c** — every entry is image-path-scoped and none touches
  locale-synced text. Vacuously true on an empty list. The docs-site 4-locale
  path is verified unchanged by the run's own diff, which touches only the skill,
  its template mirror, `internal/template/catalog.yaml`, and this SPEC's
  artifacts.

**Gap this leaves.** The `journey` comparison is now recorded evidence with no
decision resting on it. Should a later card revisit the carve-out, the pair is
there to start from rather than to re-derive — and a second candidate
(`timeline`, which maps onto A2 the same way) is the obvious next pair.

### Residual-risk

- `check-svg.mjs` asserts the accessible contract's *structure*. A `<desc>`
  reading "diagram" satisfies every check and tells a screen-reader user
  nothing; absence is mechanical, vacuity is not.
- The `journey` comparison is one sample of one journey. A journey with Latin
  labels and no satisfaction track could narrow the gap the pair shows.
- A carve-out, once taken, is durable: every future diagram of that type becomes
  an image with an image's maintenance cost and leaves the locale-sync path.

### Observed out of scope

- `SPEC §4 Evidence` and `plan.md §H` cited
  `.moai/reports/diagram-design-absorption/…`, a lead-side uncommitted research
  directory that does not exist on this branch. The artifacts are at
  `.moai/reports/t165/upstream/`, which `plan.md §C` already named correctly.
  **Repaired in sync-phase** at the lead's direction — the citations now name the
  committed copies (`survey.md`, `absorption-verdict.md`, `UPSTREAM-LICENSE`),
  and `grep -rn 'diagram-design-absorption'` over `spec.md` and `plan.md` returns
  no match. The change is a citation repair only: no requirement, criterion, or
  scope statement was touched.
- Running a command from a drifted working directory caused the statusline hook
  to write `.moai/state/{config-cache,context-usage}.json` **relative to that
  directory** rather than to the project root, creating a stray `.moai/` tree
  inside the skill folder. Removed; the underlying path-resolution behaviour is
  unaddressed and belongs to another card.

---

## §E.3 Run-phase Audit-Ready Signal

```
run_status: audit-ready
run_complete_at: 2026-08-22
run_commits: 04c7f8474, 120c89e79
ac_pass: 13 of 14 (AC-BUDGET recorded as an estimate, not a measurement — §E.2 Gaps)
```

## §E.4 Sync-phase Audit-Ready Signal

```
sync_status: audit-ready
sync_complete_at: 2026-08-22
sync_commit_sha: 68d63318c
```

The placeholder is the documented self-reference workaround: a commit cannot name
its own hash, so the sync commit writes `pending-backfill-sync` and a follow-up
commit replaces it with the real SHA (spec-frontmatter-schema.md § SHA
placeholder backfill exemption).

**Sync-phase deliverables.**

- `CHANGELOG.md` `[Unreleased]` — Summary / Fixed / Added / Changed, naming the
  accessibility defect as a defect rather than as an addition, and recording that
  the routing table is unchanged and the exception list empty.
- Evidence-citation repair in `spec.md` §4 and `plan.md` §H (`3d725d2ab`), at the
  lead's direction. Citation-only; no requirement or criterion touched.
- Status transition `in-progress → implemented → completed` on this sync commit,
  per the 3-phase close.

**Sync-phase gaps.**

- **No README or docs-site change — checked, not assumed.**
  `grep -rn 'svg-infographic' README*.md docs-site/content` returns 12 matches:
  8 in the four READMEs and 4 in the per-locale `advanced/skill-guide.md`. Every
  one is the skill's **name in a catalogue list** or a one-line description of
  what it produces ("editable SVG technical infographics, CJK font"). None
  describes its connector geometry, its palette, or its accessibility contract,
  and this work changed neither the skill's name nor what it produces — so all 12
  remain accurate and none required a sync edit.

  (An earlier draft of this section claimed the grep returned no match. It
  returns 12; the conclusion holds but the evidence behind it did not, and the
  claim is corrected here rather than left standing.)
- **Pre-existing docs drift, left alone.** The Korean `skill-guide.md` row for
  this skill omits the "architecture, flow, comparison" clause the en / ja / zh
  rows carry. It predates this work and is out of its scope; recorded so it is
  not lost, not repaired here.
- **`SPEC §4 Evidence` still names `/tmp/diagram-design`.** It is now labelled
  volatile and explicitly not a precondition, but the clone itself is outside the
  repository and will not survive a reboot. The three committed files under
  `.moai/reports/t165/upstream/` are what the citations resolve to.

## §F Phase 4 Mode Selection

```
Decision: direct
```

**Input parameters.** Tier M; scope 8 skill/mirror files plus SPEC artifacts;
domains 2 (skill documentation + one Node script); file mix markdown-dominant with
one `.mjs`; concurrency benefit LOW (single-skill authoring with no independent
sub-tasks).

**Mode evaluation.** `direct` selected — the orchestrator executed throughout and
spawned no `Agent()` at all. `serial` not selected: it is the default fallback,
but a delegation would have had to re-read the 27 KB survey and both skill files
the orchestrator had already loaded, so the spawn would have paid for context it
already held. `fanout` not selected: the work is authoring-heavy rather than
research-heavy, and the milestones share one file set, so parallel writers would
race. `sweep` not selected: 8 files with no uniform mechanical transform.

**Recorded retroactively, and stated as such.** The logging contract asks for this
section before the first run-phase `Agent()` spawn. No spawn occurred, so there was
no such boundary to write it at; this record is written at sync rather than
back-dated to pretend otherwise.

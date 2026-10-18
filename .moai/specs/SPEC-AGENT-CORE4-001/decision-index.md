# decision-index.md — SPEC-AGENT-CORE4-001

Authored at plan phase per `interview.decision_gate: on`
(`.moai/config/sections/interview.yaml:6`). One row per decision surfaced while assembling the SPEC
that the operator did not settle. Labels: `DECIDED` / `POLICY-COVERED` / `EVIDENCE-NEEDED` /
`FOUNDER`. The design documents behind the card (`.moai/reports/moai-agents-skills-redesign-20261010.html`,
`...master-upgrade-plan-20261010.html`) and the card text are not in the admissible authority
register (untracked or not a committed artifact), so no row cites them as authority; rows they
informed are `FOUNDER` with a SPEC-embedded default that stays while the row is open and is
kickoff-amendable (an operator answer supersedes it and the affected REQ/AC set is re-audited
before run entry). No row carries a recommendation. `Default:` lines are applications of the
published Default rule, not recommendations. `implementation-level` rows with a `Default:` carry
`DEFAULT-APPLIED`; every other FOUNDER row is open and goes to the operator.

### Q1: Is this one Tier L SPEC, or three (core-4 render and guards / retirement / hiding)?

Label: FOUNDER
Class: implementation-level
Authority anchor: (none — the Tier table is a rule file outside the admissible register, and it
leaves the split to judgment)
Why unresolved: Scope is 977 reference files, nine agents, two renders, two guards and a settings
generator. The Tier L ceiling (25 REQ / 25 AC) holds at 20 REQ / 24 AC, but the plan-audit ceiling
has hit oversized SPECs on other cards. Milestones M0-M4 / M5+M7+M8 / M6 are separable seams.
Default: one SPEC, M0-M8 (rule: undo is a single revert of this SPEC's own commits)
Alternate: three SPECs split at those seams
Operator verdict: DEFAULT-APPLIED 2026-10-18T04:52:36Z card t1649 plan-author (manager-spec)

### Q2: The card says "TOML 12/12"; the retire list and the new roster give nine. Which is meant?

Label: FOUNDER
Class: product-level
Authority anchor: (none)
Why unresolved: 12 definitions − 3 retired − 5 absorbed + 5 new = 9 (acceptance §3). 12 is the
pre-change count and is 11/12 in the project mirror today. If the three carried agents
(e2e-tester, manager-design, super-advisor) leave for the seed catalog, the count is 6. The SPEC
measures "N/N fresh" with N = 9; the operator may restate the target.
SPEC-embedded default: N = 9 (AC-AC4-004, 005, 007).
Operator verdict:

### Q3: Are the five absorbed names retired with a refusal, or kept as aliases for a transition period?

Label: FOUNDER
Class: product-level
Authority anchor: (none)
Why unresolved: Retiring `manager-spec`, `manager-develop`, `manager-docs`, `plan-auditor`,
`sync-auditor` removes names that appear in 977 files, 4-locale docs and the commit-trailer
vocabulary. A refusal that names the replacement costs one retry for a running lane; an alias
keeps eight names alive in the roster, registry and docs. The SPEC designs for refusal
(design §6) because the plan-pair criterion needs the new agents to be the ones the workflows
spawn.
SPEC-embedded default: refuse with replacement; trailers keep legacy values readable forever.
Operator verdict:

### Q4: Which three copies become generated?

Label: FOUNDER
Class: implementation-level
Authority anchor: (none — `AGENTS.local.md` §2.0 names C1/C2/C3 but is outside the register, and the
redesign's third copy is `~/.claude/agents`)
Why unresolved: The repository's own definition of "the three copies" (project Markdown, template
Markdown, template Codex) and the redesign's (project, template, user-scope) differ. User-scope
files live outside the repository.
Default: C1, C2, C3 plus the project Codex mirror C3p; user-scope reported stale, not touched
(rule: preserves current behavior)
Alternate: add the user-scope directory to the generated set
Operator verdict: DEFAULT-APPLIED 2026-10-18T04:52:36Z card t1649 plan-author (manager-spec)

### Q5: Which kilobyte, and is memory counted at its cap?

Label: FOUNDER
Class: implementation-level
Authority anchor: (none)
Why unresolved: 196,367 bytes is the measured plan-pair baseline and equals the redesign's "196KB"
only at 1 KB = 1,000 bytes. Counting machine-local MEMORY.md at actual size makes the check
unreproducible on another clone.
Default: 1 KB = 1,000 bytes; memory counted at its cap (rule: preserves current behavior — the
unit already used by the baseline)
Alternate: 1,024-byte KB, or live memory sizes
Operator verdict: DEFAULT-APPLIED 2026-10-18T04:52:36Z card t1649 plan-author (manager-spec)

### Q6: Is the auditor verdict-only guard on by default in the distributed template?

Label: FOUNDER
Class: product-level
Authority anchor: (none)
Why unresolved: The guard denies an auditor's writes outside `.moai/reports/`. Shipping it on
changes a template default; existing branch and stop guards ship opt-in. The card lists the guard
as a deliverable, which the SPEC reads as on-by-default.
SPEC-embedded default: on, fail-open on uncertainty, no kill switch.
Operator verdict:

### Q7: Is the 8,000-byte agent MEMORY.md cap enforced at write time?

Label: FOUNDER
Class: product-level
Authority anchor: (none)
Why unresolved: A write-time refusal blocks an agent that tries to grow its index past the cap
(remedy: shorten entries, then archive topic files; nothing is deleted). A doctor-only report would
not block. Existing oversize files on the author's machine (39,509 / 38,426 / 22,133 / 12,768
bytes) are machine-local and are listed by the doctor, not migrated by a commit.
SPEC-embedded default: write-time refusal plus doctor report.
Operator verdict:

### Q8: Which list-hiding candidates are tried, and in what order?

Label: FOUNDER
Class: product-level
Authority anchor: (none)
Why unresolved: The documentation says `skillOverrides` keys are skill names (no wildcard) and that
a skill Claude cannot invoke cannot be preloaded; it does not say whether `off` blocks preload.
The card names `off` and, for Codex, `skills.config`. The SPEC tries H1 `off` (explicit names),
H2 `user-invocable: false` + `name-only`, H3 inline-at-render with nothing published, in that
order, the first passing a recorded probe.
SPEC-embedded default: ladder H1 → H2 → H3.
Operator verdict:

### Q9: If no live runtime can run the preload probe in the run phase, which mode ships?

Label: EVIDENCE-NEEDED
Why unresolved: The probe needs a live Claude Code subagent spawn that a plan phase cannot make and
a run lane may not be able to make. Without the measurement, H1 and H2 are unverified.
Operator verdict:

### Q10: Do docs-site and README files that name retired agents stay stale until the documentation rewrite cards?

Label: FOUNDER
Class: product-level
Authority anchor: (none)
Why unresolved: 170 docs-site files and 4 README files name retired agents. The roster guard
excludes `docs-site/` by declaration. Editing them needs the four-locale same-change obligation,
which belongs to the documentation rewrite cards.
SPEC-embedded default: allowlisted with the owning card as the expiry; public docs stay stale in
the interim.
Operator verdict:

### Q11: Do the new generic names (`planner`, `implementer`, `documenter`, `auditor`, `worker`) ship unprefixed?

Label: FOUNDER
Class: product-level
Authority anchor: (none)
Why unresolved: Claude Code agent names come from frontmatter, not from the directory, so a user
project that already defines an agent named `worker` or `planner` collides with the distributed
one. The current names are prefixed (`manager-*`) or specific (`plan-auditor`).
SPEC-embedded default: names as in the card.
Operator verdict:

### Q12: Codex hiding — publish nothing and inline, or emit `[[skills.config]]` entries?

Label: FOUNDER
Class: implementation-level
Authority anchor: (none)
Why unresolved: `.agents/skills` mirrors command shims only, so no `moai-agent-*` path exists on
Codex; an entry for an unpublished path is a ghost the existing prune removes, and the manifest
records per-agent skill grants as unobservable.
Default: nothing published, body inlined, no entries (rule: preserves current behavior)
Alternate: emit entries for paths a user-scope install might add
Operator verdict: DEFAULT-APPLIED 2026-10-18T04:52:36Z card t1649 plan-author (manager-spec)

### Q13: What exactly does the retired-name refusal do?

Label: FOUNDER
Class: implementation-level
Authority anchor: (none)
Why unresolved: Refusal versus rewrite versus alias is Q3; this row is the mechanism. A deny that
names the replacement never rewrites the spawn and honors a same-named user-owned definition
outside the `.claude/agents/moai/` namespace.
Default: deny with replacement text, no rewrite, user-owned same-name definition honored (rule:
smaller user-visible surface — no retired agent stays reachable; undo is a single revert)
Alternate: deny only when no definition file of that name exists anywhere
Operator verdict: DEFAULT-APPLIED 2026-10-18T04:52:36Z card t1649 plan-author (manager-spec)

### Q14: How does a spawn tell the auditor which phase it is auditing?

Label: FOUNDER
Class: implementation-level
Authority anchor: (none)
Why unresolved: The merged auditor needs a phase argument. The Agent tool has no phase parameter,
so the carrier is the prompt text.
Default: first prompt line `phase: plan|sync`; absent → record `unknown` and return a blocker
report (rule: smaller user-visible surface — no new flag or agent name)
Alternate: two agent names
Operator verdict: DEFAULT-APPLIED 2026-10-18T04:52:36Z card t1649 plan-author (manager-spec)

### Q15: How does `auditor` coexist with `plan-auditor` in the roster guard until cutover?

Label: FOUNDER
Class: implementation-level
Authority anchor: (none)
Why unresolved: The guard test `TestNoAgentNameIsASubstringOfAnother` fails while `auditor` and
`plan-auditor` are both in the roster. Relaxing a test is not permitted; the alternatives are
deferred emission of `auditor` to the M8 swap commit, or retiring the two old auditors before the
new one exists (no auditor at all for a window).
Default: deferred emission (`hold_emission`), `auditor` measured in memory until M8 (rule:
preserves current behavior — the guard and both old auditors stay intact until the swap)
Alternate: retire first
Operator verdict: DEFAULT-APPLIED 2026-10-18T04:52:36Z card t1649 plan-author (manager-spec)

### Q16: Which development methodology does the run phase use?

Label: POLICY-COVERED
Authority anchor: `.moai/config/sections/quality.yaml` `constitution.development_mode` (value `tdd`, :4)
Why unresolved: Nothing is unresolved about the choice; the row records that the plan's RED-first
ordering and characterization-first M1 follow the setting, and that a change of the key changes the
run phase.
Operator verdict:

---
id: SPEC-SUBAGENT-WRITE-SHRINK-GUARD-001
title: "refuse a subagent's Write that drastically shrinks an existing tracked file"
version: "0.1.0"
status: in-progress
created: 2026-09-21
updated: 2026-09-22
author: manager-spec
priority: P1
phase: "v3.1.4 target"
module: internal/hook
lifecycle: spec-anchored
tier: M
tags: "subagent, write-guard, destructive-write, pretooluse, scope-discipline, t1039"
---

# SPEC-SUBAGENT-WRITE-SHRINK-GUARD-001 — a subagent's destructive overwrite has no mechanical refusal

## HISTORY

- 2026-09-21 · v0.1.0 · manager-spec · Initial authoring from card t1057. Every measured
  statement in this SPEC is attributed to `.moai/reports/t1057/plan-evidence.md` by Claim
  number. No measurement is re-derived here, and no figure appears here that is not recorded
  there. Where that file labels a finding PLAUSIBLE rather than CONFIRMED, this SPEC carries
  the same label.

---

## §0 Governing principle [HARD]

> **A subagent may overwrite a file it was asked to change. It may not, unnoticed, replace an
> existing tracked file with a fraction of what that file held.**

The subject of this SPEC is not where a subagent writes. It is that one particular write —
the blind full-file overwrite that destroys most of an existing tracked file's content — is
currently refused by nothing, and that its only signal after the fact was a rule's line number
moving.

---

## §A Background

### A.1 Evidence base

The evidence base for this SPEC is **`evidence-annex.md`, in this SPEC directory**, measured in
the `.claude/worktrees/t1057` tree at base `3dfae918a`. Claim references in this document
(`Claim 1`, `Claim 5`, …) point into that file, alongside four verbatim payload logs:
`evidence-probe-hook-payloads.jsonl`, `evidence-probe-agent-flag.jsonl`,
`evidence-probe-plain-main-write.jsonl` and `evidence-probe-manager-git-subagent.jsonl`.

The citation target is the in-directory copy deliberately. The original measurement lives at
`.moai/reports/t1057/plan-evidence.md`, which `.gitignore:227` (`.moai/reports/*`) excludes from
tracking — so it does not travel with the branch, and a reader in any other tree would find every
Claim citation resolving to nothing. `evidence-annex.md` was copied from it byte-for-byte; the
original path is an alias, never the primary citation. The copy now differs from the original by
exactly one additive annotation, measured and recorded below.

**One correction to carry when reading the annex: Claim 6.** Its conclusion — that card scope is
"not derivable by a hook" and that the card-directory axis "requires new plumbing" — is
superseded by §B.1 of this document. **Claim 6 carries a `[SUPERSEDED by … §B.1]` marker at its
head**, so a reader who opens only the annex learns this without reading back to here.

**Why the marker is present, stated so its presence is not left to be re-derived.** Annotating a
superseded conclusion is what this repository's recording discipline *prescribes* —
`moai-constitution.md` § Lessons Protocol: supersede by prefixing `[SUPERSEDED by …]`, archive
rather than delete. What the discipline forbids is a different act: editing a measured observation
to match a later fact. No observation here is edited. An earlier revision of this SPEC declined the
marker by treating the prohibition as if it governed the annotation; that decline was withdrawn.

**The measured divergence, with operand order.** The marker is the one and only difference between
`evidence-annex.md` and the original `.moai/reports/t1057/plan-evidence.md`:

| Invocation | Hunk header | Marker lines |
|---|---|---|
| `diff evidence-annex.md .moai/reports/t1057/plan-evidence.md` (annex → original) | `119,129d118` | `<` |
| `diff .moai/reports/t1057/plan-evidence.md evidence-annex.md` (original → annex) | `118a119,129` | `>` |

Both describe the same divergence: **11 lines added, 0 deletions, 0 alterations.** The operand
order is stated because a bare hunk header does not say which file was which — `d` versus `a` and
`<` versus `>` invert with the argument order, and reading one as a deletion from the annex would
be exactly backwards.

Byte-identity is therefore broken, deliberately. It was only ever a cheap proxy for "no measured
statement was altered", and that property now holds *and is measured* — a stronger guarantee than
the proxy it replaces.

### A.2 The incident (Claim 5)

A subagent, working outside the scope it was given, overwrote `.gitignore` — 417 lines at the
time — leaving a stub. The deletion entered no commit on any reachable ref: it was a
working-tree deletion caught before commit, and `git log --numstat` swept across reachable
history for any commit removing more than 50 lines from that file produced no output.

Two facts about the incident bound what a guard can be built from:

- **The loss path was inside the worktree root.** A root-boundary guard would not have fired.
- **There was no signal.** The only thing that surfaced was a rule citation's line number
  shifting `:227` → `:3`. Nothing refused, warned, or logged.

`main` carrying 319 lines of this file and `develop` 417 is **not** two damage figures for one
incident. `main` lags `develop` by 15 commits on that path and the gap is additions on
`develop` (Claim 5). It is a restoration-baseline hazard — restoring from `main` silently drops
98 lines — not a measure of the loss.

### A.3 What exists today, and what it does not cover (Claim 8)

Every constraint on subagent write scope in this repository is prose:
`moai-constitution.md` §5 Maintain Scope Discipline, and `kanban-dispatch-detail.md`'s per-card
fan-out clause. The one mechanical guard on the neighbouring axis is the branch guard, whose
axis is git — and a `Write` to `.gitignore` is not a git operation, so it is not on that axis.
`worktree-integration.md` governs WHERE a subagent writes and never WHICH files; its model is
collision avoidance, not scope confinement.

**`dangerous_removal.go` is not adjacent coverage, for two independent reasons** (measured):

1. **Tool gating.** The chain is `pre_tool.go:447` (`if input.ToolName == "Bash" &&
   len(input.ToolInput) > 0`) → `:503` (`h.checkBashCommand(input.ToolInput)`) → `:991`
   (`dangerousRemovalTarget(command)`, inside `checkBashCommand`, declared at `:978`). A `Write`
   payload never reaches it at all.

   > An earlier revision of this SPEC cited `:565` as the gate. That line carries the
   > **byte-identical** text `if input.ToolName == "Bash" && len(input.ToolInput) > 0` but opens
   > the **slot-lease** guard, which does not reach `:991`. The wrong coordinate survived a
   > re-confirmation against source, because confirming a quoted line does not confirm its
   > coordinate — verify the enclosing `^func` as well, per this repository's recorded lesson.
2. **Target set.** `protectedBasenames` (`dangerous_removal.go:178-180`) is exactly `.git` and
   `node_modules`; every other decision it makes is by path depth. Even `rm .gitignore` passes
   it.

Either reason alone disqualifies it. Nothing in this SPEC builds on it, extends it, or should be
read as describing it as partial coverage of the same hazard.

The closest mechanical precedent is the harness FROZEN zone guard
(`internal/hook/pre_tool.go`, Claim 7), which already denies `Write`/`Edit` by
`(agent identity, file_path)`. It is a precedent for the SHAPE of this guard. It is not a
precedent for its correctness: Claim 7 records two defects in it, both carried here as out of
scope (§F).

### A.3.1 The shrink check is new, not an extension [measured]

A sweep of `internal/`, `.claude/hooks/` and `scripts/` for
`mass.?deletion|deleted.?lines|diff.?size|lines_removed|deletion_guard` returned one hit, and
that hit was a false positive on an unrelated field name in `internal/statusline/types.go`.

**Nothing in this repository measures how much content a write removes.** The shrink-ratio
predicate specified below is genuinely new capability. No requirement here may be described —
in a run-phase report, a commit message, or a later SPEC — as extending, hardening, or
generalizing an existing check.

### A.4 The discriminant inputs already arrive (Claims 1-3)

Nothing needs new plumbing. A PreToolUse hook fires for a tool-spawned subagent's `Write`, and
the payload carries, measured:

| Field | Observed |
|---|---|
| `tool_name` | `Write` |
| `agent_id` | **the subagent discriminant** — see §A.5 |
| `agent_type` | an agent NAME, set on two different caller kinds — see §A.5 |
| `tool_input.file_path` | an ABSOLUTE path |

The Go side already parses both fields — `internal/hook/types.go:230` (`AgentType`) and `:239`
(`AgentID`; `:238` is the section comment above it). Both sit under section comments naming
other events, which is what made them read as unavailable on PreToolUse; the struct is flat, so
the tags bind regardless of the comment.

### A.5 Correction — `agent_type` does not identify a subagent; `agent_id` does [measured]

An earlier revision of this SPEC keyed the subagent discriminant on `agent_type` being
non-empty. **That was wrong, and the correction rests on a measurement taken after that
revision**, across three caller kinds:

All four rows are measured on `tool_name: "Write"` payloads specifically:

| Caller | `agent_type` | `agent_id` | Record |
|---|---|---|---|
| main session, plain (`claude -p`, no `--agent`, spawning nothing) | absent | absent | `evidence-probe-plain-main-write.jsonl` record 1 |
| main session launched `claude --agent manager-git` | `"manager-git"` | **absent** | `evidence-probe-agent-flag.jsonl` record 1 |
| tool-spawned subagent (`subagent_type: general-purpose`) | `"general-purpose"` | **present** (`a23362b9057cd44df`) | `evidence-probe-hook-payloads.jsonl` record 2 |
| tool-spawned subagent (`subagent_type: manager-git`) | `"manager-git"` | **present** (`ade96efaa38194545`) | `evidence-probe-manager-git-subagent.jsonl` record 2 |

In the plain-main-session record both fields are **absent from the JSON object**, not present-and-
empty — the `omitempty` shape `types.go:230,239` declares.

**Row 4 was re-measured, and the earlier citation is withdrawn.** An earlier revision cited
`a380879cdd91883f9` against `evidence-probe-hook-payloads.jsonl`. That id appears in no payload
file in this directory: **the run that produced it was never persisted**, so the value reached this
SPEC through a report rather than a record. That is an unattributed citation, not a typo — the
distinction matters, because a typo would be corrected against a file that exists, and here no such
file ever did. The case was re-run and persisted as `evidence-probe-manager-git-subagent.jsonl`;
the new id differs, which is itself the evidence that the original run was not the one now on disk.

`agent_type` carries an agent NAME and is set on BOTH a tool-spawned subagent and a
`claude --agent <name>` main session — the two are indistinguishable on that field. This matches
the field's own documented meaning (`types.go:230`, `// Custom agent name if --agent flag used`)
and the `--agent` launch form that `main-checkout-branch-guard.md:96-97` describes as real in
this repository. `agent_id` separates all three cases.

Keying on `agent_type` would therefore have denied a `--agent` main session's writes — a session
the user is directly driving, which §B.3 puts outside this guard's subject. Treated as
**measured**, not PLAUSIBLE.

**A qualification that has since been closed by measurement.** An earlier revision of this section
carried the plain-main-session row from a `tool_name: Agent` payload (the spawn call) rather than a
`Write`, and recorded "plain main session, on a `Write`, carries neither field" as a Gap in §F
rather than asserting it. That case has now been measured directly — `claude -p`, no `--agent`,
instructed to use `Write` itself and spawn nothing — and the Gap is closed; the row above cites
the payload.

The conclusion is unchanged by the closure, and was never resting on this row: the `--agent` row
alone — a measured `Write` carrying `agent_type` with no `agent_id` — is what disqualifies
`agent_type` as the discriminant. The new measurement removes a Gap; it does not prop up the
finding.

---

## §B The chosen axis, and the two that were eliminated

The guard keys on the **destructiveness of the write**, not on path scope. Both scope
definitions were considered and eliminated by measurement. They are recorded here so a later
reader does not reopen them.

| Candidate axis | Disposition |
|---|---|
| **Worktree root** — refuse a write outside the session's worktree | **Eliminated.** The t1039 loss path was INSIDE the worktree root (§A.2), so a root-boundary guard would not have fired on the very incident that motivates the card. |
| **Card directory** — refuse a write outside `.moai/specs/<card>` or equivalent | **Eliminated by operator decision, not by unavailability** — see §B.1, which corrects the conclusion of `evidence-annex.md` Claim 6. |
| **Destructiveness of the write** — RETAINED | The discriminant inputs all arrive already (§A.4), and the axis fires precisely on the shape the incident took. |

### B.1 Correction to Claim 6 — the card id IS derivable, and the axis is still not chosen

**Claim 6** of `evidence-annex.md` concluded that card scope is "not derivable by a hook in the
failing case", and that "card-directory scope requires new plumbing before it can be an
enforcement axis". An earlier revision of this SPEC carried that conclusion forward.

**What was wrong was the generalization, not the measurement.** Claim 6 examined two mechanisms
— the `MOAI_KANBAN_CARD` environment variable (`internal/config/envkeys.go:250`, unset by the
launcher at `internal/cli/kanban.go:484`) and the on-disk card record (written only for
`Source == "startup"` sessions, `internal/hook/session_start_record.go:62,74`). **Both findings
are correct and still stand.** The error was the step from "these two do not resolve it" to
"no mechanism resolves it": a **third** mechanism exists that Claim 6 never examined.

That distinction is load-bearing for the next reader. A sound partial survey with an over-general
conclusion can be **extended** — add the mechanism it missed and the conclusion updates. A
misread mechanism would have to be **redone**, because its findings could not be trusted. Claim
6's two findings can be trusted; only its closing generalization is withdrawn.

The same shape as the axis correction below: what changes is the *reason*, and a reason that
names a decision or an unexamined case leaves the question open, where an impossibility closes
it.

`internal/hook/session_start_record.go:166` `cardIDFromPath(dir)` walks up to the first ancestor
whose parent is named `worktrees` and returns that directory's basename —
`.claude/worktrees/t1057` → `t1057`. It is deliberately pure path arithmetic: the comment at
`:154-157` records that resolving the same directory via `git rev-parse --show-toplevel` pushed
SessionStart's synchronous return from under 500ms to 650-890ms per kanban session, against a
5s hook budget. This guard lives in `internal/hook`, the same package, so the function is
directly callable.

Three limits bound what the derivation actually delivers, all recorded at the source:

1. **One gated call site.** It is package-private and called only by `writeKanbanSessionRecord`
   (`:96`), which is gated to `Source == "startup"` kanban/factory sessions (`:62,74`). The
   function itself is not gated; its current consumer is.
2. **It does not verify the card exists** (`:163-165`): "a card worktree whose directory name is
   not a real card id still records that name."
3. **The `worktrees`-parent containment test is load-bearing** (`:159-165`). Without it the
   derivation always yields something inside any checkout — a session in a primary checkout
   named `moai-adk-go` would file `moai-adk-go` as its card identifier.

**The axis is unchanged: destructiveness, which stands on its own merits.** What the correction
changes is the *reason* the card-directory axis is not chosen — a decision, not an
impossibility. The two are not interchangeable: an impossibility closes a question, a decision
leaves it open to revisit. The derivation is recorded as a noted future axis in §F so the next
person finds this measurement instead of re-deriving it.

What the correction does buy this SPEC: the card id is useful as **audit-log context** — naming
which card an offending write came from — without gating anything on it. REQ-SWG-008 takes it
on those terms, and REQ-SWG-013 states the boundary.

### B.2 Structured fields, never text scanning [HARD]

The guard reads `agent_id`, `tool_name`, `file_path`, and `content` as **parsed JSON fields** (`agent_type`
is read for the audit row only — REQ-SWG-001, §A.5). It MUST NOT pattern-match over command text,
and no requirement below may be satisfied by a text scan.

This is not a stylistic preference. The repository carries seven recorded instances of a
text-scanning guard over-matching — a document's prose read as a git invocation — and the
session that produced this SPEC's evidence hit an eighth: a compound shell command containing
no git invocation of any kind was refused as "too complex to verify" (`evidence-annex.md`, Side
observation). Choosing structured fields sidesteps that entire failure class. A later
"simplification" of this guard into a text scan re-opens it.

### B.3 Write, not Edit

The guard scopes to the `Write` tool alone. `Edit` requires the caller to supply the exact
existing text it replaces, so a large deletion through `Edit` cannot be a blind overwrite — the
caller demonstrably held the content. `Write` carries no such demonstration: it replaces the
whole file from whatever the caller believes is there.

---

## §C Requirements (GEARS)

- **REQ-SWG-001** (ubiquitous) — The subagent destructive-write guard shall decide solely from
the parsed PreToolUse payload fields `agent_id`, `tool_name`, `tool_input.file_path`, and
`tool_input.content`, and shall not match any pattern against command text. `agent_type` MAY be
read for the audit row (REQ-SWG-008); it shall not be a discriminant input (§A.5).

- **REQ-SWG-002** (event-driven) — When a PreToolUse payload arrives with `tool_name` equal to
`Write` and a **non-empty `agent_id`**, the guard shall evaluate the write against the
destructive-overwrite predicate defined in REQ-SWG-004.

- **REQ-SWG-003** (unwanted) — The guard shall not evaluate a payload whose **`agent_id` is
empty**, whatever `agent_type` holds. That set is exactly the two main-session forms measured in
§A.5 — plain, and `claude --agent <name>` — and a main-session write is outside this SPEC's
subject: the main session is the party the user is talking to, and its writes are already visible
to the user in the turn that makes them.

- **REQ-SWG-004** (ubiquitous) — All sizes in this SPEC are **bytes**. The existing file's size is
its byte length on disk; the incoming size is the byte length of `tool_input.content`. Lines are
never the unit: `content` arrives as bytes, so a line count would require a second pass whose
result depends on line length, and two implementers choosing differently would change the set of
writes the guard fires on.

The destructive-overwrite predicate shall hold when all four conditions hold together, and shall
be **evaluated in this order**, stopping at the first condition that fails:

1. the incoming `content` byte length is at or below the **shrink ratio** (`SWG-T2`) of the
   existing file's byte length; and
2. the existing file's byte length is at or above the **size floor** (`SWG-T1`); and
3. the target resolves to a path inside the repository being worked in; and
4. the target is **tracked by git at HEAD** in that repository.

The order is load-bearing, not stylistic. Condition 4 costs a git subprocess on a path that fires
far more often than SessionStart — and §B.1 cites this repository's own measurement that one
`git rev-parse` per kanban session moved SessionStart's synchronous return from under 500ms to
650-890ms against a 5s budget. Putting the two payload-derived size conditions first means the git
query is reached only by writes that are already destructively shaped. Writing the conditions in
their natural narrative order (repository → tracked → floor → ratio) would put the most expensive
check first, which is why the order is stated rather than left to the implementer.

No performance claim is made here beyond the ordering; a measurement against the hook budget is
M3's deliverable.

- **REQ-SWG-005** (event-driven) — When the destructive-overwrite predicate holds, the guard
shall emit a PreToolUse deny whose reason begins with the sentinel
`SUBAGENT_DESTRUCTIVE_WRITE_VIOLATION:` and names the target path, the existing size, and the
incoming size.

- **REQ-SWG-006** (state-driven) — While `workflow.subagent_write_guard.enabled` is false, the
guard shall emit no deny, **and shall continue to run its detection and audit-log append**. Only
the refusal is gated. This is the established contract in this codebase, stated three times in
`internal/config/defaults.go` — `settings_drift_gate` (`:995-997`), `agent_model_guard`
(`:1012-1014`), `agent_stop_guard` (`:1019-1021`): the recording / observation layer runs
unconditionally, and the deny layer alone is opt-in. The distributed default is `false`, and
template neutrality binds — no `enabled: true` for this key anywhere under
`internal/template/templates/`.

- **REQ-SWG-006a** (ubiquitous) — An absent key is the normal state and shall not be read as a
defect. This repository's `.moai/config/sections/workflow.yaml` carries only three guard keys
today — `branch_guard: true` (`:164-165`), `drift_cache_fill: true` (`:173-174`),
`agent_stop_guard: true` (`:181-182`). The new key is absent from that file until someone opts
in, and takes the `defaults.go` value (`false`) meanwhile.

- **REQ-SWG-007** (ubiquitous) — On any uncertainty the guard shall fail OPEN: not a git
repository, git unavailable or exiting non-zero, an unreadable existing file, an unparseable
payload, a path that cannot be resolved, or an absent `content` field shall each allow the
write. A deny requires positive evidence on every one of REQ-SWG-004's four conditions.

- **REQ-SWG-008** (event-driven) — When the guard reaches a decision — deny or fail-open allow,
and **whether or not the deny layer is enabled** (REQ-SWG-006) — it shall append one structured
row to `.moai/logs/subagent-write-guard.log` naming the decision, the agent id, the agent type,
the target path, the two byte sizes, the derived card id where one resolves (REQ-SWG-013), and
the fail-open reason where one applies.

- **REQ-SWG-008a** (ubiquitous) — Each audit row shall carry a `decision` field valued from a
closed set, and the four values shall be mutually distinguishable by a reader of the log alone:

| `decision` | Meaning |
|---|---|
| `deny` | the predicate held, the deny layer was enabled, the write was refused |
| `withheld` | the predicate held, **the deny layer was disabled**, the write proceeded |
| `allow` | the predicate did not hold; the write proceeded |
| `fail-open` | the guard could not decide (REQ-SWG-007); the write proceeded |

**`withheld` is not a variant of `allow`, and recording it as one defeats the recording layer's
purpose.** A maintainer running with the deny layer off needs to count how often the guard *would*
have fired — that count is the production instrument for OD-1b (§D.3), the only way to measure the
false-positive rate against real traffic rather than against committed history. A row that says
`allow` where the predicate actually held destroys that count while leaving the log superficially
complete: the rows are all present, and the number that matters cannot be recovered from them.

A guard that writes a row on the disabled path but marks it `allow` therefore satisfies "a row is
written" and still violates this SPEC. AC-SWG-012 joins the pair on this field, not on the row
count.

This unconditional append is also the guard's **continued-firing signal**
(`verification-completeness.md` §1.3): because rows accumulate whether or not the deny layer is
enabled, a log that stops gaining rows while subagents are still writing means the guard has
stopped running — not that nothing destructive happened. A guard whose non-execution is
indistinguishable from its success has no liveness answer; this is that answer, and it is named
here rather than left implicit.

- **REQ-SWG-009** (state-driven) — While the guard is enabled and evaluating, it shall not
perform any write, network call, or repository mutation of its own; its only side effect is the
audit-log append of REQ-SWG-008.

- **REQ-SWG-010** (unwanted) — The guard shall not deny a write to a path that is untracked at
HEAD, including a file the calling subagent created earlier in the same run. An untracked file
has no committed content to destroy, and refusing it would block the common legitimate case of
a SPEC-authoring subagent iterating on artifacts it authored.

- **REQ-SWG-011** (ubiquitous) — The guard's deny reason shall name a route that actually works
for the denied caller: the write proceeds through `Edit` (which requires the caller to hold the
existing text), or through the main session. It shall not direct the caller to an exemption the
caller cannot reach — the failure mode already recorded for the branch guard's `manager-git`
exemption remediation text.

- **REQ-SWG-012** (ubiquitous) — The threshold pair (`SWG-T1`, `SWG-T2`) shall be defined in one
place in the Go source, in **bytes** (REQ-SWG-004), alongside a comment recording the unit, that
the values are calibrated against a single measured incident, and the named result of the M1
false-positive survey.

- **REQ-SWG-013** (ubiquitous) — The guard shall resolve the card id via `cardIDFromPath` (§B.1)
for **audit-log context only**, and shall gate no decision on it. A card id that does not resolve
shall be recorded as empty and shall change no outcome — the derivation does not verify that the
card exists, so treating it as an authorization input would build a decision on an unverified
name.

---

## §D Thresholds — a proposed pair and an open decision

### D.1 What the incident grounds

The measured incident is one data point. The annex (Claim 5) records a **411-line deletion** from
a **417-line** file — so 6 lines remained, a **~98.6%** reduction. An earlier revision of this
section wrote "roughly a 10-line stub" and "97.6%"; neither figure appears in the evidence base,
and both are corrected here to the annex's numbers. Nothing downstream moves — the proposed
thresholds sit an order of magnitude away either way — but a figure with no source is exactly
what this SPEC forbids everywhere else.

It establishes that the hazardous shape exists and what it looked like once. **It does not
establish a distribution**, and any threshold between roughly 5% and 90% would have caught it.
The incident therefore grounds the existence of the predicate, not the value of its constants.

The reduction is stated in lines because that is how the annex measured it. The guard's own unit
is bytes (REQ-SWG-004); the two are not convertible, which is why §D.2 states the proposed
constants in bytes rather than converting this figure.

### D.2 The proposed starting values

Both constants are in **bytes** (REQ-SWG-004).

| Constant | Proposed | Why this side of the range |
|---|---|---|
| `SWG-T1` — size floor | existing file ≥ **2000 bytes** | Below this, a full rewrite is a normal authoring act and the content is cheap to reconstruct. A floor is what keeps the guard away from small config files, stubs, and fixtures where whole-file replacement is routine. The value is the byte order-of-magnitude of a ~50-line source file; it is a proposal on the same footing as the ratio, NOT a conversion of a line count (no fixed conversion exists). |
| `SWG-T2` — shrink ratio | incoming ≤ **25%** of existing | The incident sits at ~1.4% by line (6 of 417), an order of magnitude inside this. The margin is deliberate: a value tight against the incident would catch that incident and nothing shaped slightly differently. |

### D.3 [RESOLVED 2026-09-22 — OD-1a] OD-1 — the false-positive side is unmeasured

**This is an open decision for the operator, deliberately not resolved here.**

The values above are grounded on the true-positive side by exactly one incident and on the
false-positive side by **nothing measured**. The question that decides them is: *how often does
a legitimate subagent `Write` shrink a tracked file of ≥`SWG-T1` bytes to ≤`SWG-T2` of its size?*
That number has not been measured, and this SPEC does not assert it.

Two **real** legitimate patterns of exactly that shape exist in this repository, and the survey
must count both:

1. **The stub-plus-lazy-companion refactor** — a long always-loaded rule file deliberately
   replaced by a short stub while its body moves to a companion file. `kanban-dispatch.md` and
   `main-checkout-branch-guard.md` both carry that structure today. A subagent performing that
   refactor through `Write` would be denied at any ratio near 25%.

   The pattern is **wider than rule files**. `CLAUDE.local.md` § References records that
   "Sections §18-27 were consolidated into external `.moai/docs/` files" — the same shape applied
   to an instruction document. It recurs in agent and skill bodies too. A survey scoped to
   `.claude/rules/` alone would undercount the false-positive population.

2. **The in-place SPEC amendment** — `spec-frontmatter-schema.md`'s documented
   `completed → in-progress (amendment)` transition, in which manager-spec rewrites a **tracked**
   `spec.md`. An amendment that relocates sections into a companion file has the same shape as
   (1), on a file this SPEC's own §E previously assumed was untracked. See the §E row.

Two dispositions are available and the operator picks:

- **OD-1a** — ship the proposed pair and accept the stub-refactor false positive, whose escape
  is REQ-SWG-011's route (`Edit`, or the main session). The guard is opt-in and default-OFF, so
  the cost lands only on a maintainer who turned it on.
- **OD-1b** — hold the values until a false-positive survey runs (plan.md M1), then fix them
  against what it measures.

**OD-1b has a second, better instrument available after the guard ships.** M1's survey measures
committed history, which the motivating incident never reached (plan.md M1 Note). Once the guard
is landed and running with the deny layer OFF, the `decision: withheld` rows of REQ-SWG-008a count
how often the proposed thresholds *would* have fired against real subagent traffic — a direct
measurement of the false-positive rate on live behaviour rather than a proxy drawn from history.
That is why `withheld` must be distinguishable from `allow` in the log: if it is not, this
instrument does not exist.

With OD-1 answered as OD-1a (2026-09-22), the values in §D.2 ship as the guard's starting
values; they remain a **proposal** rather than a measurement, and no run-phase artifact may
cite them as measured.

Operator decision round 4 (2026-09-22, AskUserQuestion, pull mode) selected **OD-1a**: the
§D.2 proposed pair ships as the guard's starting values; the stub-refactor false positive is
accepted, with escape via REQ-SWG-011 (`Edit`, or the main session); the pair stays unmeasured
on the false-positive side, and the post-landing `withheld`-log instrument (REQ-SWG-008a rows)
remains available to a later calibration card.

### D.4 What the guard does not catch, deliberately

A scope-violating write that is **not destructive** passes. A subagent that appends to a file
outside its scope, rewrites it at similar length, or creates a new file anywhere in the tree is
not refused by this guard. That is a deliberate boundary of this card, not an oversight: the
scope axis was eliminated by measurement (§B), and a guard that fires on the destructive shape
is the part that can be built from inputs that actually arrive.

---

## §E The false-positive surface

Named here so the run phase tests against it rather than discovering it.

| Case | Guard behaviour | Why |
|---|---|---|
| SPEC-authoring subagent writing a new artifact under `.moai/specs/` — **before the plan-phase commit lands** | passes | untracked at HEAD (REQ-SWG-010) |
| The same subagent rewriting an artifact it created earlier in the same pre-commit run | passes | still untracked at HEAD |
| **SPEC-authoring subagent rewriting a TRACKED artifact — the in-place amendment path** | **denied** (when the rewrite is destructively shaped) | Once the plan-phase commit lands, `spec.md` / `plan.md` / `acceptance.md` ARE tracked, and `spec-frontmatter-schema.md`'s `completed → in-progress (amendment)` transition is a documented flow that rewrites them. A section-relocating amendment has the §D.3 shape. OD-1 owns the disposition; REQ-SWG-011 is the escape. The two rows above cover only the pre-commit window — this row is what they do not cover |
| A subagent growing or lightly editing a tracked file | passes | fails REQ-SWG-004 condition 1 (ratio) |
| A subagent rewriting a short tracked file (< `SWG-T1`) | passes | fails condition 2 (floor) |
| **Stub-plus-companion refactor of a long tracked file** (rule, instruction, agent, or skill body) | **denied** | the other real false positive; OD-1 owns its disposition, REQ-SWG-011 its escape |
| A subagent regenerating an emitted artifact through `make` | passes | not a `Write` tool call |
| Main-session write of any shape — plain or `claude --agent <name>` | passes | REQ-SWG-003 (`agent_id` empty; §A.5) |

An adjacent note, not a requirement: Claude Code already refuses a `Write` to an existing file
the caller has not read. The incident happened anyway, which is the measured reason that
protection is insufficient on its own.

---

## §F Exclusions

### Out of Scope — owned by a separate card the operator has already ordered issued

- **The BranchGuard `manager-git` exemption reachability question.** `evidence-annex.md` Claim 4
  records that the doctrine at `main-checkout-branch-guard.md:96-97` declares the exemption
  unreachable from a tool-spawned subagent, while `agent_type` measurably arrives on such a
  subagent's calls and `branch_guard.go:527` returns on exactly that value. That chain is
  labelled **PLAUSIBLE, not CONFIRMED** — the end-to-end run was not measured. It is a separate
  card. This SPEC neither repairs it, nor asserts it, nor depends on its outcome.

  §A.5 adds one measured fact that bears on that card without settling it: a `claude --agent
  manager-git` main session also carries `agent_type: "manager-git"`. That widens the set of
  callers the exemption's identity test admits; whether the exemption then actually suppresses a
  deny end-to-end remains unmeasured, and remains that card's question.

### Out of Scope — unmeasured, therefore not claimed and not acted on

- **`Bash`-mediated mutation** (`sed -i`, `rm`, `>` redirect, `truncate`). Whether those calls
  carry `agent_type` / `agent_id` is Gap 1 — unmeasured. This guard covers the `Write` tool
  only, and no requirement here may be read as covering the Bash path.
- **Whether a PreToolUse deny is visible to the denied subagent.** Gap 2 — unmeasured. The
  guard's deny is specified as a deny; what the subagent then sees is not claimed.
- **Whether a worktree-anchored subagent's `cwd` is its worktree.** Gap 3 — unmeasured. This is
  why REQ-SWG-004 condition 1 resolves against the repository rather than against a `cwd`-derived
  boundary.
- **The t1039 first-hand incident account.** `.moai/reports/t1039/` is absent from this tree and
  the path is gitignored, so its absence is absence-in-this-tree, not absence-of-the-incident
  (Gap 6). Everything this SPEC says about the incident comes from the measurements in Claim 5.
- ~~**A plain main-session `Write` payload.**~~ **[CLOSED — measured]** This Gap was carried while
  the plain-main-session row rested on a `tool_name: Agent` payload. The case has since been
  measured directly on a `Write` (`evidence-probe-plain-main-write.jsonl`): both fields absent.
  Retained here, struck through, rather than deleted — a Gap that was closed by measurement is
  part of the record, and its removal would leave no trace that the conclusion once stood on
  fewer points than it does now. §A.5 carries the measured row.

### Out of Scope — deliberately left unchanged

- **The harness FROZEN zone guard's two recorded defects** (Claim 7): Defect A, the relative
  `frozenZonePrefixes` that a measured-absolute `file_path` cannot prefix-match — itself labelled
  PLAUSIBLE, not confirmed end-to-end; and Defect B, `SentinelHarnessFrozenConfig` declared at
  `pre_tool.go:66-68` and present in no prefix entry. This SPEC adds a sibling guard and repairs
  neither. Sharing the new guard's path-resolution helper back into that guard is a later card's
  decision, not this one's.
- **Path-scope enforcement in any form**, including the two axes eliminated in §B.

### Out of Scope — a noted future axis, with its measurement recorded

- **Card-directory scope enforcement.** Not built here, and recorded rather than simply excluded
  so the next person finds the measurement instead of re-deriving it.

  The card id **is** derivable inside `internal/hook` today, by pure path arithmetic, via
  `cardIDFromPath` (§B.1) — the earlier reading that a hook cannot resolve it was wrong.
  Anyone reopening this axis inherits three limits, all recorded at
  `internal/hook/session_start_record.go`: the function is package-private with one call site
  gated to `Source == "startup"` kanban/factory sessions (`:96`, `:62,74`); it does not verify the
  card exists (`:163-165`); and its `worktrees`-parent containment test is load-bearing, because
  without it a primary-checkout session files the checkout's own directory name as its card
  (`:159-165`). The comment at `:154-157` records why it is path arithmetic rather than a
  `git rev-parse` call: the subprocess pushed SessionStart's synchronous return from under 500ms
  to 650-890ms per kanban session against a 5s budget.

  What this SPEC does with that: uses the id as audit-log context only (REQ-SWG-013), gating
  nothing on it. What it does not do: enforce scope. That remains the operator's decision to
  revisit, and the destructiveness axis stands on its own merits either way.
- **The `Edit` tool**, per §B.3.
- **Prose doctrine.** No rule file under `.claude/rules/` is edited by this SPEC; the mechanical
  layer is added beside the existing prose, which stays as written.

---

## §G Cross-references

- **`evidence-annex.md`** (this directory) — the evidence base (Claims 1-8, Gaps 1-6), with
  `evidence-probe-hook-payloads.jsonl`, `evidence-probe-agent-flag.jsonl`,
  `evidence-probe-plain-main-write.jsonl` and `evidence-probe-manager-git-subagent.jsonl`
  carrying the verbatim payloads. It diverges from the untracked original
  `.moai/reports/t1057/plan-evidence.md` by the 11-line `[SUPERSEDED …]` marker and nothing else
  (`diff .moai/reports/t1057/plan-evidence.md evidence-annex.md` → `118a119,129`, 11 additions,
  0 deletions, 0 alterations). The original is an alias only (§A.1).
- `internal/hook/pre_tool.go` — the PreToolUse dispatch order and the FROZEN-zone precedent
  (`:579` the Write/Edit branch, `:583` the FROZEN check the new guard is placed after).
- `internal/hook/types.go:206-326` — `HookInput`; `AgentType` declared at `:230`, `AgentID` at
  `:239`.
- `internal/hook/dangerous_removal.go` — **not** adjacent coverage; see §A.3 for the two
  measured reasons (Bash-only: `pre_tool.go:447` `ToolName == "Bash"` branch → `:503`
  `checkBashCommand` → `:991` `dangerousRemovalTarget`; protected basenames `.git` /
  `node_modules` only, so even `rm .gitignore` passes). Cited here so a later reader does not
  re-derive it as a candidate.
- `internal/hook/session_start_record.go:166` `cardIDFromPath` — the card-id derivation, used
  here for audit context only (REQ-SWG-013, §B.1, §F).
- `.claude/rules/moai/core/verification-claim-integrity.md` — why the PLAUSIBLE labels above are
  carried rather than promoted.
- `.claude/rules/moai/development/verification-completeness.md` §2 — the two-cell adoption
  discipline the acceptance criteria follow.

---

🗿 MoAI

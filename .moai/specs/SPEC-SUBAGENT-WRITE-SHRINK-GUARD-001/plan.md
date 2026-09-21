---
id: SPEC-SUBAGENT-WRITE-SHRINK-GUARD-001
title: "implementation plan — subagent destructive-write guard"
version: "0.1.0"
created: 2026-09-21
updated: 2026-09-21
author: manager-spec
priority: P1
phase: "v3.1.4 target"
module: internal/hook
lifecycle: spec-anchored
tier: M
tags: "subagent, write-guard, pretooluse, plan"
---

# Plan — SPEC-SUBAGENT-WRITE-SHRINK-GUARD-001

Milestones are ordered by **decision-reversibility**: the decisions most likely to change come
first, the mechanical work last. M1 and M2 are where review attention belongs.

---

## §A Context

The guard is a new sibling in an established family. Five guards already sit in the PreToolUse
dispatch (`internal/hook/pre_tool.go`): the branch guard, the integration-lock guard, the
slot-lease guard, the harness FROZEN-zone guard, and the agent-model guard.

The family contract is stated three times in `internal/config/defaults.go` —
`settings_drift_gate` (`:995-997`), `agent_model_guard` (`:1012-1014`), `agent_stop_guard`
(`:1019-1021`) — and this plan follows it exactly: **detection, observation and the audit-log
append run unconditionally; only the deny layer is gated** on a `workflow.*` key shipped
default-`false`, with template neutrality (no `enabled: true` under
`internal/template/templates/`), a sentinel-prefixed deny reason, and fail-open on every
uncertainty.

One thing here is genuinely new rather than an extension: **nothing in this repository measures
how much content a write removes** (spec.md §A.3.1, measured). The shrink predicate is new
capability; the wrapper around it is the family shape.

---

## §B Known issues carried in

- The threshold pair is a **proposal**, not a measurement (spec.md §D.3, OD-1). M1 exists to
  resolve it.
- The FROZEN-zone precedent carries two recorded defects (spec.md §F). The new guard must not
  inherit Defect A — the relative-prefix / absolute-path mismatch — which is why path resolution
  is its own milestone (M3) rather than a line inside the predicate.
- **Six** measurement gaps are recorded in spec.md §F, and none may be asserted in a run-phase
  report. Three of them bind the run-phase implementation directly — Bash-mediated mutation
  (Gap 1), deny visibility to the subagent (Gap 2), and worktree `cwd` (Gap 3). The other three
  bind what may be *claimed* rather than what is built: the BranchGuard exemption end-to-end
  (Gap 4), the FROZEN Defect A end-to-end (Gap 5), and the t1039 first-hand account (Gap 6), plus
  the plain-main-session `Write` payload recorded alongside them. Read §F for all of them; this
  list narrows to the implementation-binding subset, it does not replace the count.
- The discriminant is `agent_id`, **not** `agent_type` — spec.md §A.5 records the measurement and
  the correction. An implementation keying on `agent_type` would deny a `claude --agent <name>`
  main session's writes.

---

## §C Pre-flight

1. Read `evidence-annex.md` in this SPEC directory in full, with the four payload logs beside it.
   The PLAUSIBLE labels bind. (The untracked original `.moai/reports/t1057/plan-evidence.md` is an
   alias — it does not travel with the branch, so cite the annex. The annex adds the 11-line
   `[SUPERSEDED …]` marker and alters nothing else; see spec.md §A.1.)
2. Confirm the guard family's current shape by reading the branch-guard and agent-stop-guard
   call sites in `internal/hook/pre_tool.go` and their config entries in
   `internal/config/types.go` / `defaults.go`.
3. Confirm `internal/hook/types.go` `HookInput` already carries `AgentType` and `AgentID` — no
   struct change should be needed. A needed struct change is a signal the design drifted.

---

## §D Constraints

- **[HARD] Structured fields only.** No regex over command text anywhere in this guard
  (spec.md REQ-SWG-001, §B.2).
- **[HARD] Deny layer default OFF; detection and audit always on.** `internal/config/defaults.go`
  is the SSOT and `internal/template/templates/` carries no `enabled: true` for this key
  (template neutrality). Gating the detection or the audit append on the flag is a contract
  violation, not a simplification — the family's three existing statements all put the flag on
  the refusal alone.
- **Do not build on `dangerous_removal.go`.** Measured: Bash-only call site, and a protected set
  of `.git` / `node_modules` (spec.md §A.3). It is not partial coverage of this hazard and must
  not be described as such in any run-phase report.
- **[HARD] Fail-open.** A deny requires positive evidence on all four predicate conditions.
- **[HARD] No writes from the guard** beyond the audit-log append.
- Scope discipline: the files listed in §F are the whole change surface. Repairing the
  FROZEN-zone guard, editing rule files, or touching `.gitignore` is out of scope.

---

## §E Self-verification

Each milestone states what it runs and what output makes it done. Per
`verification-completeness.md` §1.1, a green whose swept set is empty asserts nothing — every
test invocation below reports its swept count before its verdict is read.

---

## §F Milestones

### M1 — resolve OD-1: measure the false-positive side [highest reversibility cost]

The threshold pair is the decision most likely to be wrong and most expensive to change after
adoption, so it is measured first.

**The survey needs a pre-image size, and `--numstat` does not carry one.** `git log --numstat`
reports per-commit added/deleted line *deltas*. The predicate needs the size the file had
*before* the commit, which must be measured separately per candidate. A survey specified as
"`--numstat` gives the population" cannot produce the quantity it exists to produce; the four
steps below are the executable form.

1. **Candidates** — enumerate commits with a large deletion on a tracked file across the working
   surfaces. `--numstat` is the right tool for this step and only this step: it narrows the
   population cheaply.

   ```
   git log --numstat --format='%H' -- .claude/ .moai/ internal/ pkg/ cmd/ *.md
   ```

   Keep rows whose deleted count is large and whose added count is small. This is a *candidate
   filter*, deliberately loose — the ratio is not decidable here.

2. **Pre-image** — for each candidate `(commit, path)`, measure the size the file had in the
   parent commit. Bytes, matching REQ-SWG-004's unit:

   ```
   git cat-file -s "<commit>^:<path>"
   ```

   and the post-image with `git cat-file -s "<commit>:<path>"`. A path absent in the parent (a
   new file) drops out here — it has no pre-image and cannot be a shrink.

3. **Ratio** — keep `(commit, path)` where `post <= SWG-T2 * pre` and `pre >= SWG-T1`. This is
   the step that produces the population OD-1 asks about.

4. **Positive control** — before reading any count, confirm the pipeline fires at all: run
   steps 1-3 against a `(commit, path)` known to carry a large deletion, and observe it survive
   to step 3. **An empty result at step 3 with no positive control is "not measured", not
   "zero"** — the same discipline §G already applies to test runs, applied to the survey.

Then classify each surviving hit — destructive accident, deliberate stub-plus-companion refactor,
or in-place SPEC amendment (spec.md §D.3 names all three) — and report the counts to the
operator with the candidate pair, taking the OD-1a / OD-1b answer.

**Survey surface:** not `.claude/rules/` alone. spec.md §D.3 records that the stub-plus-companion
shape also appears in instruction documents (`CLAUDE.local.md` § References), agent bodies, and
skill bodies. A survey scoped to rule files undercounts the false-positive population, which
biases OD-1 toward shipping.

**Done when:** all four steps' commands and their verbatim output are recorded, the positive
control is shown firing, the counts are stated per class, and the operator has answered OD-1.
The thresholds are not written into Go source before this answer lands.

**Note:** this survey measures *committed* history, and the motivating incident never reached a
commit (spec.md §A.2). It therefore measures the false-positive side, which is what OD-1 needs;
it is not evidence about how often the true positive occurs, and must not be reported as such.

### M2 — the predicate and its two directions [design decision]

Implement `checkSubagentDestructiveWrite` in a new file `internal/hook/subagent_write_guard.go`:
payload-field extraction, the four-condition predicate (REQ-SWG-004), the deny reason with its
sentinel and a reachable remediation route (REQ-SWG-005, REQ-SWG-011), and the threshold pair in
one place with the calibration comment (REQ-SWG-012).

Tests land with the code, and they land **bidirectionally** — a mutation-detecting deny case and
a no-mutation pass case — per acceptance.md AC-SWG-001a / AC-SWG-001b. One direction alone
cannot distinguish a working guard from one that denies everything.

**Done when:** `go test ./internal/hook/... -run SubagentWriteGuard -v` shows both directions
passing with a non-empty swept set, and the same tests fail on a mutant that inverts the
predicate's ratio comparison.

### M3 — path resolution, not inherited from the precedent

Resolve the absolute `file_path` against the repository, and answer tracked-at-HEAD. This is
its own milestone because the precedent's Defect A is exactly a path-resolution mistake and
inheriting it silently is the expected failure.

- Resolve the repository from the payload's `cwd`, falling back the way the existing hooks do.
- Answer tracked-at-HEAD with git, and treat any non-zero exit as fail-open (REQ-SWG-007).
- **Honour REQ-SWG-004's evaluation order**: the two payload-derived size conditions are decided
  before the repository is resolved, so the git subprocess is reached only by writes already shown
  to be destructively shaped.

**Done when:** a test proves an absolute `file_path` is matched (the shape the precedent misses),
an untracked path passes (REQ-SWG-010), and **a test observes that a payload failing the ratio
condition reaches no git invocation at all** — the order is the deliverable, not a comment about
it. Recording a measurement of the guard's added latency against the 5s hook budget is
recommended; making an unmeasured performance claim is not.

### M4 — config key and wiring [mechanical]

- `WorkflowConfig.SubagentWriteGuard SubagentWriteGuardConfig \`yaml:"subagent_write_guard"\``
  in `internal/config/types.go`, with the family's comment shape.
- Default `false` in `internal/config/defaults.go`.
- Bump the config cache version, the way `AgentStopGuard` / `SlotLease` each did — an
  old-binary cache serving a stale struct is a recorded hazard in `internal/config/cache.go`.
- Call site in `internal/hook/pre_tool.go`, inside the existing
  `ToolName == "Write" || ToolName == "Edit"` branch but gated to `Write` alone, placed **after**
  the FROZEN-zone check so it cannot displace an established deny.
- Local dogfood opt-in in `.moai/config/sections/workflow.yaml` (`enabled: true`), with the
  distributed template left at the default.

**Done when:** `go test ./internal/config/... ./internal/hook/...` passes, and two tests prove
the disabled path separately: no deny is emitted (AC-SWG-002) AND the audit row is still written
(AC-SWG-012). Testing only the first would pass a mutant that disables the whole guard.

### M5 — audit log, including the card-id context [mechanical]

`.moai/logs/subagent-write-guard.log` append on every decision — deny, fail-open, and
deny-layer-disabled alike (REQ-SWG-008) — matching the branch guard's audit-log shape. Each row
carries the card id resolved via `cardIDFromPath` (REQ-SWG-013), empty when it does not resolve.

**Done when:** a test observes a row on the deny path, a row carrying the fail-open reason on an
uncertainty path, a row on the deny-layer-disabled path, and a row whose card-id field is
populated from a worktree-shaped path and empty from a primary-checkout-shaped one.

---

## §G Anti-patterns

- **Text-scanning the command.** The eighth over-match instance was hit while gathering this
  card's own evidence (spec.md §B.2). A regex here re-opens the class.
- **Calling the thresholds measured.** Until OD-1 lands they are a proposal; a run-phase report
  citing them as measured is an unobserved claim.
- **Widening to `Edit` "while in the file".** §B.3 states the reason `Edit` is excluded; widening
  is a scope decision, not a convenience.
- **A remediation string pointing at an unreachable exemption.** The branch guard's reason text
  already had to be corrected for exactly this (REQ-SWG-011).
- **Gating the audit append on the config flag.** Inverts the family contract: the flag is a kill
  switch for the DENY, never for the record.
- **Gating a decision on the derived card id.** `cardIDFromPath` does not verify the card exists
  (spec.md §B.1 limit 2), so the id is audit context and nothing more (REQ-SWG-013).
- **Describing the shrink check as extending an existing guard.** Measured: no such check exists
  (spec.md §A.3.1).
- **Reading an empty test run as a pass.** A `-run` selector matching zero tests exits 0 and
  prints `ok`.

---

## §H Cross-references

- spec.md §A.5 (the `agent_id` discriminant), §B (axis elimination), §D (thresholds + OD-1),
  §E (false-positive surface), §F (scope + the six gaps).
- acceptance.md — the bidirectional mutation pair and the RED-now cells.
- `internal/hook/pre_tool.go` — dispatch order and the precedent's shape.
- `internal/config/cache.go` — the cache-version bump obligation.

---

🗿 MoAI

---
id: SPEC-SUBAGENT-WRITE-SHRINK-GUARD-001
title: "acceptance criteria — subagent destructive-write guard"
version: "0.1.0"
created: 2026-09-21
updated: 2026-09-21
author: manager-spec
priority: P1
phase: "v3.1.4 target"
module: internal/hook
lifecycle: spec-anchored
tier: M
tags: "subagent, write-guard, acceptance, bidirectional-mutation"
---

# Acceptance — SPEC-SUBAGENT-WRITE-SHRINK-GUARD-001

Every criterion below is Given-When-Then and binary-testable. Per
`verification-completeness.md` §2, each carries a RED-now cell (the criterion observed red on
the pre-implementation tree, with the reason it is red) and a green-path cell naming the
milestone that flips it.

Tree for every RED-now observation: `.claude/worktrees/t1057` at base `3dfae918a`. This
document-level pin binds every criterion carrying no pin of its own
(`verification-completeness.md` §2.1).

---

## §D.0 Evidence ledger

`verification-completeness.md` §2.1 requires every release-blocking criterion's RED-now cell to
carry four elements — command, verbatim stdout, exit code, tree SHA — and names the fenced
evidence-ledger entry as the recommended carrier, because a table cell mangles shell
metacharacters. Criteria below cite these entries by id rather than restating them.

### EL-1 — the guard symbol does not exist

```
command: grep -rn SUBAGENT_DESTRUCTIVE_WRITE_VIOLATION internal/ --include='*.go'
stdout:  (empty)
exit:    1
tree:    3dfae918a
```

The quotes around `*.go` are load-bearing. Unquoted, zsh glob-expands `--include=*.go` against the
cwd, finds no match, and **refuses to run `grep` at all** (`zsh:1: no matches found`) — while still
exiting 1. The recorded exit and the observed exit then coincide through different mechanisms: one
is "grep searched and found nothing", the other is "grep never ran". The second would stay exit 1
after M2 lands, so the red→green transition this ledger exists to demonstrate could never be
observed. Both forms were run in this tree; only the quoted one executes `grep`.

Cited by: AC-SWG-001a, AC-SWG-001b, AC-SWG-002, AC-SWG-003, AC-SWG-004, AC-SWG-005, AC-SWG-006,
AC-SWG-012, AC-SWG-013, AC-SWG-014.

This is one RED command shared by ten criteria, and that is the honest shape: the whole guard is
absent, so every criterion about its behaviour is red for the same reason. What each criterion
must still carry separately is its **green path** and its **mutant probe** — those are what
distinguish the criteria from one another, and a shared RED does not merge them.

### EL-2 — the guard source file does not exist

```
command: ls internal/hook/subagent_write_guard.go
stdout:  (empty — the path names no file)
exit:    1
tree:    3dfae918a
```

Cited by: AC-SWG-007.

### EL-3 — no audit log exists

```
command: ls .moai/logs/subagent-write-guard.log
stdout:  (empty — the path names no file)
exit:    1
tree:    3dfae918a
```

Cited by: AC-SWG-008.

### EL-4 — OD-1 carries no recorded answer

```
command: ls .moai/specs/SPEC-SUBAGENT-WRITE-SHRINK-GUARD-001/od1-survey.md
stdout:  (empty — the path names no file)
exit:    1
tree:    3dfae918a
```

Cited by: AC-SWG-011. The file is the M1 survey record; its absence is the mechanical form of
"OD-1 is open" (see AC-SWG-011 for why this replaces a document-reading check).

---

## §D AC matrix

| AC | Subject | Severity | Milestone |
|---|---|---|---|
| AC-SWG-001a | destructive write IS refused (mutation detected) | release-blocking | M2 |
| AC-SWG-001b | normal write STILL passes (no-mutation success) | release-blocking | M2 |
| AC-SWG-002 | disabled config emits no deny | release-blocking | M4 |
| AC-SWG-012 | disabled config writes the audit row marked `withheld` (not `allow`) | release-blocking | M4 |
| AC-SWG-013 | card id is audit context, gates nothing | release-blocking | M5 |
| AC-SWG-003 | untracked target passes | release-blocking | M3 |
| AC-SWG-004 | absolute path is matched | release-blocking | M3 |
| AC-SWG-005 | BOTH main-session forms pass (plain + `--agent`) | release-blocking | M2 |
| AC-SWG-006 | fail-open on uncertainty | release-blocking | M2 |
| AC-SWG-007 | no text scanning in the guard source | release-blocking | M2 |
| AC-SWG-008 | audit row on every decision path, all four `decision` values distinguishable | regression-guard | M5 |
| AC-SWG-009 | size floor holds | regression-guard | M2 |
| AC-SWG-010 | remediation route is reachable | regression-guard | M2 |
| AC-SWG-011 | OD-1 answered before thresholds are fixed | release-blocking | M1 |
| AC-SWG-014 | audit append is the guard's only side effect | release-blocking | M2/M5 |

15 matrix rows across 14 logical criteria — `AC-SWG-001a` / `AC-SWG-001b` are the paired
sub-criteria of one AC, per the sub-ID convention. Both counts sit within the Tier M ceiling of
16 (`spec-workflow.md` § SPEC Complexity Tier).

---

## §D.1 The bidirectional mutation pair [HARD]

These two are authored and reviewed as ONE unit. Either alone cannot distinguish a working
guard from a broken one: AC-SWG-001a alone passes for a guard that denies every write, and
AC-SWG-001b alone passes for a guard that denies nothing.

### AC-SWG-001a — a destructive out-of-scope write is refused

**Given** the guard is enabled, and a tracked file of 20000 bytes exists in the repository,
**When** a PreToolUse payload arrives with a non-empty `agent_id`, `tool_name: "Write"`,
`tool_input.file_path` the absolute path of that file, and `tool_input.content` of 300 bytes
(1.5% — inside `SWG-T2` and above `SWG-T1`),
**Then** the handler returns a deny whose reason begins with
`SUBAGENT_DESTRUCTIVE_WRITE_VIOLATION:` and names the target path and both byte sizes.

- **RED now:** EL-1. Red because the guard is absent, not because it is wrong — the correct
  reason, and the one M2 flips.
- **Green path:** M2. Passing output is a `go test ./internal/hook/... -run
  SubagentWriteGuard/destructive -v` run reporting a non-empty swept set and `--- PASS`.
- **Mutant probe:** invert the ratio comparison in the predicate (`<=` → `>`). This criterion
  must go red. A criterion that survives that mutation is too shallow to adopt.

### AC-SWG-001b — a normal in-scope write still passes

**Given** the guard is enabled, and the same tracked 20000-byte file exists,
**When** the identical payload arrives except that `tool_input.content` is 20600 bytes,
**Then** the handler returns no deny, and the decision path reached is the allow fall-through.

- **RED now:** EL-1 — the same observation as its twin, and red for the same stated reason
  (the guard is absent). Note what EL-1 does and does not establish here: with no guard in the
  tree there is no "allow the guard produced" to distinguish from the tree's existing allow, so
  this criterion only becomes *discriminating* once M2 lands. It is adopted as one half of the
  pair, never on its own.
- **Green path:** M2. Passing output is `-run SubagentWriteGuard/normal -v` reporting a
  non-empty swept set and `--- PASS`.
- **Mutant probe:** make the predicate return true unconditionally. This criterion must go red
  while AC-SWG-001a stays green — that divergence is exactly what the pair buys, and observing
  it is the completion act for both.

---

## §D.2 Remaining criteria

### AC-SWG-002 / AC-SWG-012 — the layer split, as a pair

These two are the family contract expressed as criteria, and like §D.1 they are read together:
AC-SWG-002 alone passes for a guard that is entirely inert when disabled, which is precisely the
contract violation (`defaults.go` puts the flag on the refusal, never on the record).

**AC-SWG-002 — the disabled path emits no deny.**
**Given** `workflow.subagent_write_guard.enabled` is false (the distributed default, and the
state of this repository's `workflow.yaml`, which carries no such key),
**When** the AC-SWG-001a payload arrives,
**Then** no deny is emitted.

**AC-SWG-012 — the disabled path still writes the audit row, marked `withheld`.**
**Given** the same disabled state,
**When** the same payload arrives and the predicate would have held,
**Then** one row is appended to `.moai/logs/subagent-write-guard.log` whose `decision` field
reads exactly **`withheld`** — not `allow` (REQ-SWG-006, REQ-SWG-008, REQ-SWG-008a).

**The join is on the field, not on the row count.** An earlier form of this criterion asserted
only that a row appears, which a fourth mutant satisfies while violating the requirement: a guard
that writes a row on the disabled path and marks it `allow` passes both AC-SWG-002 (no deny) and a
count-based AC-SWG-012 (a row exists). That mutant destroys the `withheld` count, which is OD-1b's
production instrument (spec.md §D.3) — the rows all look present and the number that matters is
unrecoverable. AC-SWG-001a joins its pair at field level for the same reason; this criterion now
matches it.

- **RED now (both):** EL-1. **Green path:** M4.
- **Mutant probe 1 (whole-guard gating):** make the config flag gate the entire guard rather than
  the deny alone. AC-SWG-002 stays green; AC-SWG-012 must go red.
- **Mutant probe 2 (the fourth mutant — decision flattening):** on the disabled path, write the
  row with `decision: allow` instead of `withheld`. AC-SWG-002 stays green; AC-SWG-012 must go
  red. A criterion that survives this mutation is asserting row presence, not row content.
- **Counting check (not a mutant):** with the deny layer disabled, a run of N destructively-shaped
  payloads and M benign ones must yield exactly N rows matching `decision: withheld`. **The benign
  set M must include payloads that fail ONLY condition 3 or ONLY condition 4** — a size-shaped
  write to an untracked path, and one to a path outside the repository — because a benign set built
  solely from size-condition failures lets a fifth mutant pass: an implementation whose disabled
  path evaluates conditions 1-2 and skips the git query records `withheld` on size alone, satisfies
  AC-SWG-002 and AC-SWG-012, and inflates the count with writes that were never destructive. That
  mutant is motivated by this SPEC's own REQ-SWG-004 ordering rationale (the git query is the
  expensive condition), which is exactly why the benign set must reach it.
- **Mutant probe 3 (condition skipping):** on the disabled path, evaluate only conditions 1-2 and
  record `withheld`. With the benign set defined above, this criterion must go red.

### AC-SWG-013 — the card id is context, not a gate

**Given** the guard is enabled,
**When** it evaluates a payload whose `cwd` sits inside `.claude/worktrees/<card>`, and
separately one where the card id does not resolve,
**Then** the audit row carries the derived id in the first case and an empty field in the second,
**and the deny/allow outcome is identical in both** (REQ-SWG-013).

- **RED now:** EL-1. **Green path:** M5.
- **Mutant probe:** make an unresolved card id suppress the deny. This criterion must go red —
  `cardIDFromPath` does not verify the card exists (spec.md §B.1 limit 2), so any decision resting
  on it rests on an unverified name.

### AC-SWG-003 — an untracked target passes

**Given** the guard is enabled and the target path is NOT tracked at HEAD,
**When** a payload arrives that satisfies every other predicate condition,
**Then** no deny is emitted.

This is the criterion that keeps a SPEC-authoring subagent working on artifacts it created
earlier in its own run (REQ-SWG-010, spec.md §E).

- **RED now:** EL-1. **Green path:** M3.

### AC-SWG-004 — an absolute `file_path` is matched

**Given** the guard is enabled,
**When** `tool_input.file_path` arrives as an absolute path — the form measured in Claim 1 —
**Then** the guard resolves it against the repository and reaches the predicate, rather than
failing to match.

This is the precedent's Defect A written as a criterion so the new guard cannot inherit it.

- **RED now:** EL-1. **Green path:** M3.
- **Mutant probe:** replace the resolution with a relative-prefix `strings.HasPrefix` — the
  precedent's shape. This criterion must go red.

### AC-SWG-005 — both main-session forms pass

**Given** the guard is enabled,
**When** a payload arrives with an **empty `agent_id`** and otherwise satisfies the predicate —
tested in both measured main-session forms: `agent_type` also empty (plain), **and** `agent_type`
carrying an agent name such as `"manager-git"` (the `claude --agent` launch),
**Then** no deny is emitted in either case (REQ-SWG-003).

Both forms are required. The `--agent` form is the one that a guard keyed on `agent_type` would
wrongly deny, and it is measured (spec.md §A.5, `evidence-probe-agent-flag.jsonl`).

- **RED now:** EL-1. **Green path:** M2.
- **Mutant probe:** key the discriminant on `agent_type != ""` instead of `agent_id != ""`. The
  `--agent` case must go red while the plain case stays green — that divergence is what makes
  this criterion detect the D1 defect rather than merely describe it.

### AC-SWG-006 — fail-open on uncertainty

**Given** the guard is enabled,
**When** the payload is unparseable, or `content` is absent, or the target is not in a git
repository, or the git query exits non-zero,
**Then** no deny is emitted and the write is allowed.

- **RED now:** EL-1. **Green path:** M2.
- **Mutant probe:** make any one uncertainty path deny. This criterion must go red.

### AC-SWG-007 — no text scanning

**Given** the landed guard source,
**When** it is read,
**Then** it contains no regular expression or substring match over a command string, and its
only inputs are the parsed payload fields named in REQ-SWG-001.

- **RED now:** EL-2.
- **Green path:** M2. Verified by reading the landed file, not by a grep for the absence of
  regexes — an absence grep over a file that does not exist returns 0 for the wrong reason.

### AC-SWG-008 — an audit row on every decision path

**Given** the guard, enabled or disabled,
**When** it denies, when it allows, when it fails open, and when the deny layer is disabled,
**Then** `.moai/logs/subagent-write-guard.log` gains one structured row per decision, each
carrying the correct `decision` value from REQ-SWG-008a's closed set — all four values reachable
and mutually distinguishable — with the fail-open row also naming its reason. (The `withheld`
value specifically is AC-SWG-012's subject.)

- **RED now:** EL-1 (the guard) and EL-3 (the log path). **Green path:** M5.

### AC-SWG-009 — the size floor holds

**Given** the guard is enabled and the existing tracked file is below `SWG-T1`,
**When** it is overwritten with a fraction of its content,
**Then** no deny is emitted.

- **RED now:** EL-1. **Green path:** M2. The concrete floor value is OD-1's (AC-SWG-011).

### AC-SWG-010 — the remediation route is reachable

**Given** a deny reason emitted by this guard,
**When** it is read by the denied caller,
**Then** the route it names is one the caller can actually take (`Edit`, or the main session)
and it names no exemption unreachable from a tool-spawned subagent (REQ-SWG-011).

- **RED now:** EL-1. **Green path:** M2. Verified by reading the emitted string.

### AC-SWG-011 — OD-1 is answered before the thresholds are fixed

**Given** the landed guard source defining `SWG-T1` and `SWG-T2`,
**When** a test reads the M1 survey record at
`.moai/specs/SPEC-SUBAGENT-WRITE-SHRINK-GUARD-001/od1-survey.md`,
**Then** that file exists, carries a non-empty `od1_answer:` field valued `OD-1a` or `OD-1b`, and
the REQ-SWG-012 calibration comment in the Go source contains the survey's recorded counts
verbatim. Absence of the file, an unanswered field, or a comment whose counts do not match the
record fails the test.

This is a **mechanical** gate, replacing an earlier form that only asked a reader to confirm the
survey had been recorded. The distinction is the one `verification-completeness.md` §2.1 draws:
a criterion whose RED cannot be re-executed loses release-blocking eligibility. EL-4 re-executes
— the file's absence is observable today and observable after M1 — so this criterion keeps its
release-blocking classification honestly rather than by assertion.

- **RED now:** EL-4 (the survey record does not exist) plus EL-1 (no Go source defines the
  constants yet). Red for the right reason: OD-1 is genuinely open, and spec.md §D.3 states the
  false-positive side is unmeasured.
- **Green path:** M1. Passing output is the four survey steps' commands with their verbatim
  output, the positive control firing, the per-class counts, and `od1_answer:` recorded.
- **Mutant probe:** write the constants into Go source with the survey record absent. This
  criterion must go red. A criterion that stays green under that mutation is the document-check
  form this one replaced.

### AC-SWG-014 — the guard's only side effect is the audit append

**Given** the guard is enabled,
**When** it evaluates any payload to completion — deny, allow, or fail-open,
**Then** the only file written is `.moai/logs/subagent-write-guard.log`, no network call is
made, and the repository state is unchanged (`git status --porcelain` identical before and
after) (REQ-SWG-009).

This criterion exists because REQ-SWG-004 condition 4 *requires* a git invocation, which makes
"what side effect is permitted" a boundary that has to be tested rather than assumed. A guard on
the PreToolUse path that mutates the repository or calls out to the network is a real hazard, and
the requirement that forbids it had no criterion until now.

- **RED now:** EL-1. **Green path:** M2 (the predicate) confirmed at M5 (the audit path).
- **Mutant probe:** add a temporary-file write to the guard. This criterion must go red.

---

## §D.3 Definition of Done

- All release-blocking criteria green, each with a non-empty swept count recorded.
- The bidirectional pair (AC-SWG-001a / AC-SWG-001b) observed diverging under both mutants of
  §D.1 — the deny-everything mutant and the deny-nothing mutant.
- The layer-split pair (AC-SWG-002 / AC-SWG-012) observed diverging under BOTH its mutants —
  whole-guard gating AND decision flattening (`withheld` written as `allow`).
- The `withheld` counting check run with the deny layer disabled: N destructive payloads yield
  exactly N `decision: withheld` rows.
- AC-SWG-005's two forms observed diverging under the `agent_type`-discriminant mutant.
- OD-1 answered, `od1-survey.md` recorded with `od1_answer:` populated, and the answer noted in
  progress.md §E.2.
- `go vet ./internal/hook/... ./internal/config/...` and `golangci-lint run` clean on the
  changed packages.
- No file outside the change surface in plan.md §F modified; `git status --porcelain` shows
  nothing unexpected.
- Every PLAUSIBLE label from spec.md §F still carried as PLAUSIBLE in the run-phase report.

---

🗿 MoAI

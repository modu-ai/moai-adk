---
id: SPEC-KANBAN-BOOTSTRAP-001
title: "Research — Kanban session topology, bootstrap, and dispatch"
version: "0.3.0"
status: draft
created: 2026-08-10
updated: 2026-08-11
author: manager-spec
priority: High
phase: "v3.1.0 target"
module: internal/kanban
lifecycle: spec-anchored
tags: "kanban, research, measurement, substrate, guard, baseline, env-access, sole-writer"
tier: L
dependencies: [SPEC-KANBAN-RENAME-001, SPEC-KANBAN-BOARD-001, SPEC-KANBAN-WORKTREE-001]
related_specs: [SPEC-KANBAN-MULTISESSION-001, SPEC-FACTORY-MODE-001]
---

## §A. What was measured, and where

All measurements below were taken in the worktree `/Users/goos/.moai/worktrees/kanban` at plan-phase authoring time. Each is recorded with the command that produced it, because a claim without its command is not attributable and cannot be re-checked later.

Three of these measurements changed a position the predecessor SPEC held. Those three are §B, §C, and §D. Two further measurements were added at v0.2.0 in response to the plan audit: §H, which shows an acceptance criterion was scanning a direction the governed code does not use, and §I, which identifies the three predecessor requirements one requirement here absorbs.

Two more were added at v0.3.0, both in response to the same audit's blocking findings. §D gains a third gap at the seam it already documents — runtime role resolution, used by four requirements across three SPECs and defined by none — and §J measures the launcher surface the backend decision of `spec.md` §A.8 had been resting on unmeasured. The launcher measurement contradicts part of what the audit assumed, and the contradiction is recorded rather than smoothed: the per-session half of the premise **holds**.

## §B. The canonical CLI guard the predecessor cited does not exist

`SPEC-KANBAN-MULTISESSION-001` §A.7 and §D.3 named `TestNew_NoAskUserQuestion` in `internal/cli/worktree/new_test.go` as *the* canonical static guard, and instructed the implementer to extend it.

```
grep -rn 'TestNew_NoAskUserQuestion' --include='*.go' .
```

Six hits, **all comments**, in six different test files:

- `internal/cli/update_hooks_guidance_test.go:171`
- `internal/cli/init_update_notice_test.go:202`
- `internal/cli/web_test.go:79`
- `internal/cli/preference/cmd_test.go:14`
- `internal/cli/harness/propose_boundary_test.go:4`
- `internal/harness/cluster/subagent_boundary_test.go:12`

Zero `func` definitions. And the file it is attributed to does not exist:

```
ls internal/cli/worktree/
# clean.go clean_stale_test.go done.go done_test.go guard.go guard_test.go
# mock_extensions_test.go recover.go remove.go render.go root.go root_test.go
# shared.go subcommands_test.go sync.go .gitkeep
```

There is no `new*.go` file in that directory at all.

**What this means.** The guard *pattern* is real and heavily used — 29 `func Test…NoAskUserQuestion` definitions exist across the tree — but the specific instance everyone cites is a ghost. Each of the six comments above was written by an author copying the pattern and citing its source; the source has since been removed or was never there, and the citation propagated anyway. The predecessor's criterion, written from those comments, would have been unsatisfiable: an implementer told to extend `TestNew_NoAskUserQuestion` finds nothing to extend and either invents something or reports a false blocker.

**Live definitions verified present**, any of which can serve as the anchor:

```
grep -rn '^func Test.*NoAskUserQuestion' --include='*_test.go' . | wc -l   # 29
```

Among them `TestMCP_NoAskUserQuestion` (`internal/cli/mcp_boundary_test.go:17`), `TestWeb_NoAskUserQuestion`, `TestPropose_NoAskUserQuestion` (`internal/cli/harness/propose_boundary_test.go:31`), `TestDoctor_NoAskUserQuestion`, `TestGoal_NoAskUserQuestion`, `TestCluster_NoAskUserQuestion`, `TestNoAskUserQuestionInSubagents`, `TestDecayScanCmd_NoAskUserQuestion`, `TestNewMxScanCmd_NoAskUserQuestion`.

`TestMCP_NoAskUserQuestion` was read in full and is a straightforward `os.ReadFile` plus two `strings.Contains` assertions — a shape trivially reproducible for a new source file. This SPEC anchors on it, and re-verifies the anchor at M0 rather than trusting this record, because trusting an authoring-time citation is precisely what produced the dead name.

**A second stale citation, out of band.** `internal/cli/CLAUDE.md` § Conventions repeats the dead name twice (line 11, line 31) and additionally attributes `TestPropose_NoAskUserQuestion` to `internal/cli/harness/route_test.go`, whereas its real home is `propose_boundary_test.go`. Since every new guard in this repository is written by copying that citation, an uncorrected doc manufactures a new dead reference per copy. Recorded in `plan.md` §B.2; repairing it is out of scope here because this SPEC did not create it.

## §C. The backward-compatibility baseline has a recording window

The predecessor's §D.1 required the no-config path to be compared against a recorded baseline of parse result, environment mutation, and exec argument vector. It never said when the recording happens.

The entry switch surface measured here:

```
grep -rn 'parseFactoryFlag' internal/cli/cc.go internal/cli/glm.go
# internal/cli/cc.go:96   if specID, factoryEnabled, factoryArgs := parseFactoryFlag(filteredArgs); factoryEnabled {
# internal/cli/glm.go:169 if specID, factoryEnabled, factoryArgs := parseFactoryFlag(filteredArgs); factoryEnabled {
```

```
grep -n 'MOAI_FACTORY' internal/config/envkeys.go
# 142: EnvMoaiFactory     = "MOAI_FACTORY"
# 147: EnvMoaiFactorySpec = "MOAI_FACTORY_SPEC"
```

Every one of those tokens — the flag the parser recognizes, both environment-variable values, and the function name itself — is renamed by `SPEC-KANBAN-RENAME-001`. A baseline recorded before that rename encodes all of them.

**The failure that produces.** The comparison then reports a difference on every run, and the difference is the rename. It is a true observation and a useless one: the guard fires on the one change everybody already agreed to, and it keeps firing until somebody re-records the baseline. At that moment the honest response (investigate, then re-record) and the lazy response (just re-record) look identical from outside, and the guard has taught its own readers that re-recording is the normal reaction to a red result. A guard that cries wolf on a sanctioned change is worse than no guard, because it consumes the attention a real regression would need.

**Resolution.** Record against the post-rename tree — after the rename lands on the base branch, before this SPEC touches the entry switch. That window is narrow and sequencing-sensitive, which is why it is milestone M1 rather than a line inside a later one, and why the criterion checks the baseline's **provenance** (post-rename identifiers present, retired identifiers absent) in addition to its content.

## §D. Board state has one origin, and the discriminant behaves

The predecessor proposed a `column:` field in each SPEC's frontmatter as the board's source of truth. That is rejected; the board sibling owns a single-origin store under the primary checkout's `.moai/state/kanban/`, and `REQ-KB-007` forbids the board from writing SPEC frontmatter at all.

The resolution rule was measured from inside this worktree:

```
git rev-parse --git-common-dir
# /Users/goos/MoAI/moai-adk-go/.git
```

The parent of that path is the primary checkout root, resolved correctly from a worktree — the same discriminant the repository's branch guard already uses to distinguish a primary checkout from a worktree. So the rule is not novel machinery; it is an existing discriminant applied to a new path.

**Why it matters to this SPEC specifically.** This SPEC prints guidance and sends dispatches, and both are places where a careless instruction would tell a session to consult "the board" without saying which copy. A worker resolving board state relative to its own working tree finds a different file, or no file, and the board forks per worktree with no error emitted. The obligation here is negative — emit nothing that contradicts the rule — which is why it appears in `plan.md` §D as constraint C8 rather than as a requirement.

**A second rule over the same store, added at v0.2.0.** The board sibling reached v0.2.0 with `REQ-KB-017`: the `lead` is the sole writer of board state and every other session reads it. That rule had been bundled in the predecessor with the `column:` frontmatter mechanism §D rejects, and the split deleted both — the sibling's own research records `grep -rc 'sole writer\|single writer'` returning **0** across all fourteen files of the three SPECs it produced. What let the loss hide is partly this SPEC's doing: its §C said it decides who is *told* about a card, which is not a claim about who writes one, and the sibling's §C said it named no actor. Two disclaimers, no owner. This SPEC now defers by name (`plan.md` §D constraint C8a) rather than by silence; the rule itself is not restated here and is not this SPEC's to restate.

**A third gap at the same seam, found at v0.3.0 — and it is the sole-writer failure repeating.** `REQ-KB-017` names its writer as "the session occupying the `lead` role" and explicitly defers election to `REQ-KS-004`; `REQ-KW-007` refuses a disposal "**when** no session occupying the `lead` role is resolvable" and `REQ-KW-011` performs a release "by the session occupying the `lead` role", both deferring the same way. Read together with this SPEC's `REQ-KS-019` ("the session whose **declared role** owns the card's column") that is four consumers of a runtime role lookup. Measured across the family:

```
grep -rn 'occupying the `lead` role' .moai/specs/SPEC-KANBAN-WORKTREE-001/spec.md
# 211: (prose) Who performs that write is not this SPEC's …
# 372: REQ-KW-007
# 386: REQ-KW-011

grep -rn 'occupying the `lead` role' .moai/specs/SPEC-KANBAN-BOARD-001/spec.md
# 173: (prose) One writer. Exactly one role — the session …
# 229: REQ-KB-017
```

Five occurrences across the two siblings — three of them in requirement text — and not one of them accompanied by a definition of how the occupancy is observed. The phrase is used and never grounded. `REQ-KS-004` — the requirement every one of them points at — defines the role *set* and elects `lead`; it says nothing about which running session occupies one. `REQ-KB-004`, the only record that could have carried it, stores a **session identifier**, not a role.

The mechanism is identical to the sole-writer loss above: each document points at a neighbour, the neighbour points back or points at a third, and the contract exists in no document at all. What differs is only that this one was caught before implementation rather than after a split. Resolution: `REQ-KS-006` is widened from addressability to addressability plus role declaration, including the clause the sibling gates need — resolvable from a session that is not the `lead` (`spec.md` §A.13).

## §E. The messaging substrate

Five properties were measured in a prior session against the live runtime and are carried into `spec.md` §A.3. They are **re-measured at run-phase** (REQ-KS-003), because a property measured in a prior session is evidence about that session's runtime rather than about this one — and the substrate is a moving target in a way the filesystem is not.

The two constraints that follow are not measurements but consequences, and neither is negotiable:

- **Sockets cannot launch sessions.** No design in which the lead spawns its workers is available. Bootstrap is manual by necessity.
- **A message is not consent.** The receiving runtime is told the text came from another session, not from the user. This is the direct source of the three prohibitions in `design.md` §D.5.

## §F. The configuration surface

```
ls .moai/config/sections/
# archive constitution context delegation design feedback gate git-convention
# git-strategy handoff harness interview language llm lsp mcp-matrix mx
# observability project quality ralph report research security state statusline
# sunset ...
```

A per-concern sectioned YAML layout, which the topology configuration joins as one more section rather than as a new mechanism. Two things follow: the `moai init` question flow is the established way such a section gets written, and there is no precedent for expressing a section's contents as launcher arguments — which is the existing shape REQ-KS-008's prohibition preserves rather than invents.

## §G. Alternatives considered and rejected

### G.1 A lead that spawns its workers — rejected (not possible)

Not a preference. No mechanism exists; sockets carry messages between running sessions and cannot start one.

### G.2 Proceeding with a partial team at quorum expiry — rejected

The generous-looking option. Produces a board that stops moving with no error emitted, because the lead's correct refusal to dispatch across role boundaries is indistinguishable from idleness. Aborting converts a silent late failure into a loud immediate one at the cost of one relaunch.

### G.3 An interactive backend prompt — rejected (prohibited)

The natural implementation of "let the operator choose" is a selection prompt, and CLI code may not prompt. Replaced by printing both commands where the backend is unconfigured — a choice offered without a question asked.

### G.4 Topology as entry-switch arguments — rejected

Would put role counts and backends on the launcher command line. Rejected against the sectioned-configuration precedent of §F, and because a topology expressed as arguments has to be retyped identically on every start.

### G.5 Checking the pointer-only rule by grepping for requirement tokens — rejected

Passes on a payload that paraphrases, which is the case worth catching. Replaced by a type with no free-text field, making a body unrepresentable rather than discouraged.

### G.6 Admitting the mixed backend — rejected (preserved boundary)

Contradicts the one-session/one-backend premise the state record encodes, and the messaging layer does not replace the process model it assumes. The rejection and its sentinel are preserved verbatim; this is deliberate avoidance of a destructive change, not an unfilled gap.

### G.7 Deriving the session count from the WIP limit — rejected

Collapses two knobs into one and makes a WIP change silently ineffective until the session count is also raised. Held apart by `REQ-KB-010` on the board side and REQ-KS-005 here.

### G.8a Scanning only `os.Getenv("` for the env-constant rule — rejected (vacuous)

The v0.1.0 criterion for REQ-KS-009. Rejected on the measurement of §H: the governed surface performs five environment accesses and none of them is a `os.Getenv(` call, so the scan cannot fire against it even when the violation is deliberate. Replaced by an enumeration over all five access forms with a positive control in each direction. Narrowing it back to the read form is named as an anti-pattern in `plan.md` §G (AP-9) because the read-only version is the shorter, more natural thing to write.

### G.8b Comparing against an unrecorded "before" — rejected

Three criteria asked for something to be "identical to before" while nothing obliged anyone to capture what before was. Rejected as unfalsifiable rather than merely weak: at verification time the original is gone, so the comparand is a recollection and every implementation passes. Replaced by REQ-KS-011's existing shape — a durable artifact plus a provenance check — generalized to REQ-KS-022, REQ-KS-023 and REQ-KS-024.

### G.8c Stating the lead's fixed backend without an override attempt — rejected

"Not operator-selectable" as a bare assertion is satisfied identically by an implementation that never mentions the lead's backend and by one that silently honors an override. Replaced by a per-channel refusal (`spec.md` §A.12). The alternative of *also* claiming that no hand-launched non-Claude lead can exist was considered and rejected: the entry switch prints and does not supervise, so no observation available here decides it, and a criterion asserting it would be asserting an unmakeable measurement.

### G.8d Deriving a session's role from its launch label — rejected (does not invert)

The cheapest possible answer to the §D role-resolution gap: the label already exists, is stable, and is what the lead addresses. Rejected on this SPEC's own text rather than on taste. `REQ-KS-014` requires every emitted command's label to differ from every other's, while `AC-KS-013` has each unconfigured worker role emit two commands — one per supported backend — and `REQ-KS-005` permits two `run` sessions. One role therefore corresponds to two-or-more possible labels, and which exists is the operator's choice at launch. Making the derivation work would require collapsing the distinct-label rule or removing the backend choice, and both properties are held for independent reasons. Replaced by a declaration carried alongside the label (`spec.md` §A.13, REQ-KS-006), with the derivation named as an anti-pattern at `plan.md` §G (AP-14) because it is the shorter thing to write.

### G.8e A role declaration readable only by the lead — rejected

The form an implementation reaches naturally, since the lead is the only consumer visible from inside this SPEC. Rejected on the sibling measurement of §D: `REQ-KW-007` and `REQ-KW-011` both resolve the `lead` occupant from a session that is not the lead, so a lead-private declaration satisfies every dispatch-routing test here and breaks two gates over there — a failure that surfaces in another SPEC's criteria, which is the kind least likely to be caught by the SPEC that caused it. REQ-KS-006 therefore carries the resolvable-from-any-session clause explicitly rather than leaving it implied by the consumers.

### G.8f Fixing the declaration's carrier in this SPEC — rejected

Considered, and rejected for lack of a measurement to decide it. The launch command, the session registry, and the peer-discovery output can each carry a role declaration, and nothing measured here favours one: the registry already resolves a session identifier to a process identity (`REQ-KW-009` consumes it that way), the discovery tool already returns a peer list the lead reads, and the launch command already carries `--name`. Fixing one would be fixing a run-phase decision on preference, so REQ-KS-006 fixes the contract — a declaration exists, is distinct from the label, resolves from any session — and leaves the carrier open. The cost is that the SPEC cannot be read to learn where the role lives; that cost is recorded in `plan.md` §E.

### G.8 Re-anchoring the criterion on `TestPropose_NoAskUserQuestion` as cited — rejected

It exists, but not where `internal/cli/CLAUDE.md` says it does (`propose_boundary_test.go`, not `route_test.go`). Citing it from the doc would reproduce the §B failure in miniature. The anchor is taken from a verified `func` definition, and re-verified before use.

## §H. The entry switch accesses the environment by writing, not by reading

`AC-KS-009` at v0.1.0 verified REQ-KS-009 by scanning authored call sites for an inline `os.Getenv("` literal. The surface it governs is `internal/cli/factory.go`, the existing entry-switch environment mutation this SPEC extends and renames. Measured here:

```
grep -nE 'os\.(Setenv|Getenv|LookupEnv|Unsetenv)\(' internal/cli/factory.go
#  93:    _ = os.Setenv(config.EnvMoaiFactory, "1")
#  95:        _ = os.Setenv(config.EnvMoaiFactorySpec, specID)
# 107:    prev, had := os.LookupEnv(key)
# 110:            _ = os.Setenv(key, prev)
# 113:        _ = os.Unsetenv(key)
```

```
grep -c 'os.Getenv(' internal/cli/factory.go
# 0
```

**What this establishes.** Five environment accesses, three of them writes, one a presence-preserving lookup, one an unset — and **zero** `os.Getenv(` calls. The read-only scan was therefore not a weak check of this surface but a **vacuous** one: it examines a construct the code does not contain, so it returns clean regardless of what the code does, and a violation authored as `os.Setenv("MOAI_KANBAN", …)` passes it untouched. The absence of a positive control is what let that stand — a control would have shown immediately that the scan could not fire.

Two smaller facts fall out of the same reading and are worth recording. The file's own comment at line 84 explains that a variable's *presence* is the signal downstream readers key on, and that `os.Getenv` cannot observe presence while `os.LookupEnv` can — so `os.LookupEnv` is a read the naive scan would have missed even on a surface that did read. And every one of the five sites already passes a `config.Env…` constant rather than a literal (`internal/config/envkeys.go:142`, `:147`), so the precedent this SPEC extends already satisfies REQ-KS-009; what was missing was a check capable of noticing if the new code stopped doing so.

**Resolution.** `AC-KS-009` enumerates all five access forms in both directions and asserts each key argument resolves to a constant, with a positive control inserted once as a write and once as a read. Reads stay in scope because REQ-KS-009 binds every call site — the read-only form failed by being incomplete, not by being wrong about reads.

## §I. What `REQ-KS-024` absorbed

The audit found `REQ-KS-024` carrying three predecessor requirements with no disclosure. Read from the preserved predecessor at `.moai/state/kanban-source/SPEC-KANBAN-MULTISESSION-001/spec.md`:

- `REQ-KM-037` — template source before the local counterpart, `make build`, commit the regenerated `internal/template/catalog.yaml`.
- `REQ-KM-038` — a mirrored pair preserves its measured relationship; a sanitized pair becoming byte-identical is a failure, not a convergence.
- `REQ-KM-039` — no SPEC identifier, REQ or AC token, internal date, or commit SHA under `internal/template/templates/`.

`REQ-KM-040` (full suite, not an affected-packages subset) is carried one-to-one as `REQ-KS-025`, and `REQ-KM-001` one-to-one as `REQ-KS-001`; neither is a compression and neither needs disclosure.

**Why it matters that it was undisclosed.** An auditor reconciling 25 requirements against 59 cannot see that three arrived at one slot, so the compression is invisible exactly where an omission would be. And the sibling `SPEC-KANBAN-BOARD-001` v0.2.0 records what a bundled requirement costs when nobody can see the bundle: its predecessor's `REQ-KM-044` carried a rejected storage mechanism and a surviving ownership rule in one requirement, the split rejected the first and deleted both, and the loss went unnoticed until two independent auditors found it. Disclosure is cheap; the failure it prevents is not. The bundling is kept at the requirement layer — all three bind one milestone and one file set — and undone at the criterion layer, where the seven observations that were reported by a single verdict are now decided by four (`spec.md` §D.7).

## §J. The launchers are per-session backend selectors **and** project-global mutations

The backend design of `spec.md` §A.8 — a fixed Claude-backed `lead`, worker backends chosen by which printed command the operator runs — rests on a premise that had never been measured: that `moai cc` and `moai glm` select a backend **per session**, so `plan` can run under one and `run` under the other in the same project. The v0.2.0 audit named this as the SPEC's central unverified premise, and observed that the SPEC re-measures five messaging-substrate properties (REQ-KS-003) while measuring nothing about the launcher surface underneath the decision.

Measured here, the premise is **half right**, and both halves matter.

### J.1 The per-session half holds — through the process environment, not through a file

```
grep -n 'func setGLMEnv' internal/cli/glm.go
# 229:func setGLMEnv(glmConfig *GLMConfigFromYAML, apiKey string) {
```

`setGLMEnv` writes the backend selection into the **process** environment (`os.Setenv` of the auth token, the base URL, and the four model slots, at `glm.go:230-236`). The launcher then replaces the process:

```
grep -n 'syscall.Exec' internal/cli/launcher.go
```

`launchClaudeDefault`'s own doc comment (`launcher.go:613-616`) records that it "replaces the current process with claude via `syscall.Exec`", which inherits that environment. The backend therefore travels **with the launched session** and not through any shared file — and the GLM path says so deliberately, in a comment at `glm.go:255-260` explaining that `settings.local.json` injection was removed precisely so the GLM environment would not leak into later `claude` invocations.

So an interleaved launch does give each worker its own backend. The mechanism `spec.md` §A.8 depends on is sound, and this is the half that would have invalidated the design had it failed.

### J.2 The project-global half was assumed away, and the two launchers undo each other

```
grep -nE 'clearTmuxSessionEnv\(\)|removeGLMEnv\(settingsPath\)|resetTeamModeForCC\(root\)|cleanupMoaiWorktrees\(root\)|persistTeamMode\(root, "glm"\)' internal/cli/launcher.go
# 213:    if err := clearTmuxSessionEnv(); err != nil {
# 218:    if err := removeGLMEnv(settingsPath); err != nil {
# 222:    teamModeMsg := resetTeamModeForCC(root)
# 227:    worktreeMsg := cleanupMoaiWorktrees(root)
# 264:    if err := persistTeamMode(root, "glm"); err != nil {
```

Lines 213-227 are inside `applyCCMode` (`launcher.go:212`); line 264 is inside `applyGLMMode` (`:237`). Reading each:

| Launcher | Project-global effect | Evidence |
|---|---|---|
| `moai glm` | writes `team_mode: glm` into `.moai/config/sections/llm.yaml` | `persistTeamMode` (`glm.go:601-616`) resolves the sections directory and saves the LLM section |
| `moai glm` | inside tmux, writes the **shared** tmux session environment | `injectTmuxSessionEnv`, guarded by `tmux.NewDetector().InTmuxSession()` |
| `moai cc` | clears that shared tmux session environment | `clearTmuxSessionEnv` (`:213`) |
| `moai cc` | strips the GLM variables from `.claude/settings.local.json` | `removeGLMEnv` on `filepath.Join(root, defs.ClaudeDir, defs.SettingsLocalJSON)` (`:218`) |
| `moai cc` | disables `team_mode`, printing what it cleared | `resetTeamModeForCC` (`:429-448`) returns `"Team mode disabled (was: %s)"` |
| `moai cc` | removes `worker-`-prefixed worktrees under two bases | `cleanupMoaiWorktrees` (`:481-`), enumerating `git worktree list --porcelain` |

The bootstrap prescribes running these commands up to five times in one project (one `lead` plus four workers, `REQ-KS-013`), and `spec.md` §A.5 prescribes re-running the entry switch after a quorum abort. Interleaving is therefore the normal case, and each `moai cc` in the sequence undoes what a prior `moai glm` persisted.

### J.3 What follows, and what does not

**It is tolerable, and the reason is specific rather than optimistic**: none of the mutated project-global state determines the backend of a session that has already been exec'd (§J.1). A worker launched under GLM stays on GLM when a later `moai cc` resets `team_mode`. The claim being made is that the interleaving does not break the design — not that it has no effects.

Two effects it does have are recorded rather than dismissed:

- **The residual `team_mode` is not a record of the team's composition.** After a mixed launch it holds whatever the last launcher wrote, which is a fact about launch order. Anything wanting composition reads the role declarations of `spec.md` §A.13. Named as an anti-pattern at `plan.md` §G (AP-16).
- **`worker-` is a reserved worktree-name prefix.** `moai cc` removes worktrees carrying it, so a per-card worktree named that way would be destroyed by the ordinary act of launching a Claude-backed worker — silently, since the removal reports through a launcher message nobody is reading at that moment. The naming belongs to `SPEC-KANBAN-WORKTREE-001` `REQ-KW-003`; the collision is recorded at `plan.md` §B.5 so that sibling does not walk into it.

**Resolution.** `REQ-KS-003`'s re-measurement list is extended to cover both halves, `spec.md` §A.8 states the consequence rather than assuming per-session isolation, and `plan.md` §C gains checks 8 and 9 so the measurement is re-run at M0 rather than trusted from this record — the same discipline §B applies to the guard anchor, and for the same reason.

## §K. Out of Scope

### Out of Scope — this document

- Measurements of the board store's internals or the worktree lifecycle. Those belong to the sibling SPECs' research, and duplicating them here would create a second record to drift.
- Re-measuring the rename's mapping. `SPEC-KANBAN-RENAME-001` owns it; this SPEC consumes the result and halts if it is absent.
- Any measurement taken outside this worktree at this HEAD. Every claim above is attributable to a command run here; a figure recalled from another session would not be.
- Measuring which carrier a role declaration should use, or benchmarking the three candidates of §G.8f. No such measurement was taken and none is claimed; the carrier is a run-phase decision.
- Measuring `moai cg`. It is rejected (§G.6) and §J measures only the two launchers whose commands the bootstrap guidance emits.

# design.md — SPEC-FACTORY-BOOTSTRAP-001

> Design alternatives considered and rejected. This document records the **decisions not taken** so a future reader (or auditor) can see what was on the table and why it lost. The decisions taken are in `spec.md` (requirements) and `plan.md` (milestones).

---

## §1. The dispatch shape — how to evaluate `-f` and `--name` together

### Alternative A — flat selection on the combination (CHOSEN)

```
1. (specID, factoryEnabled, rest) := parseFactoryFlag(filteredArgs)
2. (label, isCompanion) := parseCompanionLabel(rest)
3. switch {
   case factoryEnabled && isCompanion:  enterFactoryCompanionMode(label)
   case factoryEnabled && !isCompanion: enterFactoryMode(specID)
   default:                             // no-op — covers both !factoryEnabled && isCompanion
                                       // (BREAKING from 94025ce0a: was companion entry under prior art)
                                       // and !factoryEnabled && !isCompanion
   }
```

Both flags parsed unconditionally; the branch is selected on the combination. The two `!factoryEnabled` rows of the §A.2 truth table collapse into one default because both are no-ops — `isCompanion` is consulted only when `-f` is present. This is REQ-FB-001 / REQ-FB-002.

### Alternative B — keep `else if`, add a "both" check up front (REJECTED)

```
if factoryEnabled && hasCompanionShapeName {
    enterFactoryCompanionMode(...)
} else if factoryEnabled {
    enterFactoryMode(...)
} else {
    // no-op (both companion-shape --name alone and non-companion --name alone)
}
```

Rejected because under this SPEC's semantics (`--name` alone is a no-op regardless of shape — see spec.md §A.2.1), the `else if isCompanion` branch that would have called `enterFactoryCompanionMode(...)` under `94025ce0a` is dead code: `isCompanion` is only load-bearing when `-f` is also present. Alternative B therefore either (a) parses `--name` even when `-f` is absent (wasted work — the result is no-op regardless of `--name` shape), or (b) skips parsing `--name` when `-f` is absent, which makes the dispatch shape conditional on `-f` and obscures the four-row truth table's symmetry. Alternative A's flat switch — parse both flags unconditionally, select branch on the combination — is the cleaner expression of the truth table and satisfies REQ-FB-002's "evaluate both flags together" rule without special-casing the `-f`-absent path.

### Alternative C — make `parseFactoryFlag` role-aware (REJECTED)

Have `parseFactoryFlag` itself inspect `--name` and return a `(role, specID, rest)` triple. Rejected because it couples the flag parser (whose contract is "find and remove `-f` / `--factory`") to the companion-label parser (whose contract is "recognize the `<role>-<run-id>` shape"). The two parsers have independent invariants and test suites today; coupling them churns both for no semantic gain.

### Alternative D — leave dispatch as-is, change the companion env var instead (REJECTED)

Keep the `else if`, but make a companion launched with `-f --name X` set a third env var (say `MOAI_FACTORY_COMPANION`) that the dispatch already recognizes. Rejected because it does not solve the problem: the `else if` still routes the `-f`-carrying companion to the lead branch, so the lead env (`MOAI_FACTORY`) would be set anyway; distinguishing after the fact requires the dispatch to re-check inside the lead branch, which is the same complexity as Alternative A with worse readability.

---

## §2. The `crossSessionInbound` injection surface

### Alternative A — transient `--settings <file>` under `os.TempDir()` (CHOSEN)

Write `{"crossSessionInbound": "accept"}` to `os.TempDir()/moai-factory-<pid>-<rand>.json`, pass `--settings <that>` to `claude` / `glm`. The `--settings` flag's documented merge semantics take the strictest tier across project / local / file, so the injected file's `accept` wins regardless of the operator's local config. Cleanup via `defer os.Remove(...)` in the existing restore-on-exit path.

### Alternative B — write `crossSessionInbound: accept` into `.claude/settings.local.json` (REJECTED)

Rejected on three grounds:

1. **Ineffective.** `settings.local.json` is the *lower-priority* tier; the stricter project settings would override it. The accept/hold/refuse ladder cannot be satisfied from a persistent write at all (that is the measured finding — `crossSessionInbound` is in 0 settings files today because no project-file write can guarantee `accept` in the presence of a stricter tier).
2. **Intrusive.** It mutates an operator-managed file. The operator's `settings.local.json` carries per-machine values (tmux pane IDs, API tokens, absolute paths — CLAUDE.local.md §2); appending a runtime field to it pollutes the diff and risks being committed.
3. **Race-prone.** Two concurrent launches would race on the same file. The transient-file approach gives each launch its own file by PID.

### Alternative C — pass `--setting` (sic, Claude Code runtime flag) inline as JSON (REJECTED)

`claude -p` accepts `--settings '{"model": "opus"}'` as an inline JSON argument (this is the form that appears in `moai-foundation-cc/reference/*.md`). Rejected for the launcher path because (a) the launcher invokes `claude` in TUI mode, not `-p`, and the inline-JSON form is documented for headless; (b) inline JSON on the command line is visible to every process on the host via `ps`, and although `crossSessionInbound: accept` is not a secret, the pattern sets a bad precedent for future fields that might be; (c) the file form composes cleanly with the operator's own `--settings <file>` (Alternative A's path: if the operator passed one, moai's injection is suppressed — REQ-FB-007).

### Alternative D — env var `CLAUDE_SETTINGS_PATH` or similar (REJECTED)

Rejected because Claude Code does not document an env-var equivalent of `--settings`; the contract is the flag. Inventing an env var would assert behavior the runtime does not provide.

---

## §3. The notice-content ordering and the SPEC-line conditionality

### Alternative A — fixed order, SPEC line conditional on env (CHOSEN)

Lead notice order: (a) run id → (b) four companion lines → (c) leader socket → (d) inbound-automation → (e) SPEC id **iff `MOAI_FACTORY_SPEC` set**. The conditionality sits on the SPEC line alone.

### Alternative B — always print the SPEC line, with `(none)` when unset (REJECTED)

Rejected because printing `SPEC: (none)` teaches the operator to ignore the line. The whole point of conditioning on the env var is that the SPEC identifier is meaningful only when the operator explicitly targeted one; an empty placeholder degrades the signal.

### Alternative C — make the SPEC line carry a default (`SPEC-FACTORY-MODE-001`) when unset (REJECTED)

Rejected because `SPEC-FACTORY-MODE-001` is the predecessor that defines the chain seed; defaulting to it would mis-attribute a run that the operator did not tie to a SPEC, and would collide with this SPEC's own identifier in operator logs.

### Alternative D — emit the SPEC line at the top, not the bottom (REJECTED)

Rejected because the run id is the load-bearing identifier for companionship (it is what appears in the `--name <role>-<run-id>` form); the SPEC identifier is operator-facing bookkeeping and belongs after the launch instructions.

---

## §4. The companion notice — role or no role

### Alternative A — join-only, role-less (CHOSEN)

`"Factory Mode: joined run <id>."` Period.

### Alternative B — keep the role (prior-art behavior) (REJECTED)

Rejected on the collision-with-sibling ground (spec.md §A.6). `SPEC-KANBAN-BOOTSTRAP-001` `REQ-KS-006` owns the role-declaration carrier; this SPEC's notice is unconditional and would be a second source of role truth.

### Alternative C — print the role only when a topology config is absent (REJECTED)

Rejected because the topology config does not exist yet (the sibling is deferred). Conditionalizing on a config that does not exist is a no-op today and commits the SPEC to a behavior change the moment the sibling lands; better to ship the role-less form now and let the sibling's eventual declaration be the sole role truth.

### Alternative D — print the role from the launch-time label, defer to the sibling's declaration when it lands (REJECTED)

This is the "two sources, eventually reconciled" position. Rejected because the launch-time label and the runtime declaration are not guaranteed to agree — the sibling's `REQ-KS-006` widens from "stable label" to "addressability plus role declaration resolvable from a session that is not the lead", which is a runtime property a launch-time notice cannot know. Two divergent sources is worse than one source deferred.

---

## §5. The docs-site section — `multi-llm/` vs `advanced/` vs new section

### Alternative A — `multi-llm/` (CHOSEN)

`multi-llm/` carries backend-mixing sessions (`moai cg`, GLM panes, tmux Agent Teams). A factory run is a backend-mixing multi-session pattern. The lead notice itself says "Substitute 'moai glm' for 'moai cc' on any companion", confirming the backend-mix framing.

### Alternative B — `advanced/` (REJECTED)

`advanced/` carries single-session advanced patterns (deep worktree flows, autonomy loops, advanced configuration). A multi-session factory run is structurally different (multiple terminals, inter-session messaging) and would sit awkwardly among single-session topics.

### Alternative C — new section `factory/` or `multi-session/` (REJECTED)

Rejected because it requires (a) a new `_meta.yaml` section entry, (b) a new icon in `layouts/partials/menu.html:28-44` SVG switch, and (c) a new `main.yaml` top-level entry with 4-locale `name` map — all for one page. The page fits cleanly under `multi-llm/`; a new section is unjustified YAGNI until at least 3 multi-session pages exist.

### Alternative D — `workflow-commands/` next to `moai-plan` / `moai-run` (REJECTED)

Rejected because `workflow-commands/` is the reference for the `/moai:*` slash commands; Factory Mode is a launch-time entry switch (`moai cc -f`), not a slash command. Placing it there would mis-categorize the surface.

---

## §6. The CLI help shape — `Use`/`Long` vs flag registration

### Alternative A — documentation in `Use` / `Long`, no `cmd.Flags()` (CHOSEN)

Both commands set `DisableFlagParsing: true`; the flag registration would be silently inert. The documentation lives in the prose that `moai cc --help` prints.

### Alternative B — register via `cmd.Flags()` anyway, for `--help` consistency (REJECTED)

Rejected because a registered-but-inert flag misleads readers: it implies cobra parses the flag, which it does not. The flag is forwarded to `claude`, and the parse site is the launcher's own `parseFactoryFlag`. Registering it would be a documentation anti-pattern (plan.md §G AP-3).

### Alternative C — split into a subcommand (`moai cc factory`, `moai cc companion`) (REJECTED)

Rejected because it breaks the prior-art UX: the operator runs `moai cc -f` as a single command, and `94025ce0a`'s tests and notice text assume the flag form. A subcommand split is a larger UX change than this SPEC's scope allows, and the sibling (`SPEC-KANBAN-BOOTSTRAP-001`) will revisit the entry-switch shape when it lands.

---

## §7. AC003 preserve — why the block-cap wiring is untouched

The prior-art `injectStopHookBlockCapForGoal` reads **`EnvMoaiFactory != "" || EnvMoaiFactoryLabel != ""`** unconditionally (line ~50), ahead of the goal read. This is the property that gives both the lead (via `MOAI_FACTORY`) and the companion (via `MOAI_FACTORY_LABEL`) the raised cap without consulting goal state at launch. The `-f` redefinition changes **which env var a companion sets** but not **whether one of the two is set** — so the OR-branch continues to fire for companions. Touching the block-cap wiring would risk the two AC003 tests (REQ-FB-018) for no semantic gain. The decision is to leave it alone.

---

## §8. Sibling boundary — what stays unilateral

This SPEC states the boundary with `SPEC-KANBAN-BOOTSTRAP-001` from its own side only (spec.md §C). The sibling's own §C is authoritative for its side; this SPEC does not edit it. The boundary is asymmetric because:

1. The sibling is deferred (draft status); editing its files now would lock it into a boundary it has not yet negotiated from its own side.
2. The forward reference ("when the sibling lands, its topology-config-gated notice supersedes this SPEC's unconditional notice") is a contract **this SPEC offers**, not one it imposes; the sibling is free to accept or renegotiate when it activates.
3. The `verification-claim-integrity.md` §1.1 surface 4 binds the premise: this SPEC does not claim the sibling's behavior; it claims only what this SPEC does when the sibling lands.

---

## §9. Design alternatives not enumerated (out of scope)

The following are deliberately not enumerated here because they are out of scope (spec.md §C), not because they were not considered:

- **Role-declaration carrier** — sibling (`REQ-KS-006`) territory.
- **Quorum wait / dispatch protocol** — sibling (`REQ-KS-007`, `REQ-KS-012`, `REQ-KS-018`, `REQ-KS-019`) territory.
- **Topology config format** — sibling territory.
- **Board model** — `SPEC-KANBAN-BOARD-001` territory.
- **Worker spawn automation / liveness** — `SPEC-KANBAN-WORKTREE-001` territory.

A reader who reaches for one of these should consult the sibling named, not this design doc.

# t360 — GLM effort keyed to the slot the session actually occupies

Card: t360 · Branch: `WT-glm-effort-slot` · Base: local `develop` e50964ad3

**Partial delivery. The wire is repaired; the two display sites are deliberately NOT.** The reason
is in § The display half, held, and it is a decision the lead has to make, not one this card can
settle from inside the tree.

## Claim

A per-tier GLM effort a user stores is no longer discarded on the wire when the session sits on a
slot whose flash-ness differs from the high slot's.

The effort was already resolved per slot (`resolveGLMMainSessionEffort` → `glmSlotEffortForModel`).
The model that governs the collapse was not: all three consumption sites passed
`llm.GLM.Models.High`. `ResolveGLMReasoningForModel` pins the result to `max` whenever the model it
is handed is flash, so a session on a non-flash slot with a stored `low` wired as `max` because the
HIGH slot happened to be flash. One axis was wrong, not two.

## Evidence

**RED, planted by the lane rather than accepted from the implementing agent.** The lane re-keyed
the wire back to the high slot and re-ran (`lane-verify-red.log`):

```
--- FAIL: TestBuildEnvForGLMLaunch_KeysCollapseToTheSessionSlot
    --- FAIL: .../fable_session_on_a_non-flash_slot_keeps_its_stored_low
        glm_slot_model_wire_test.go:53: ANTHROPIC_REASONING_EFFORT = "max", want "low" —
        the fable slot runs glm-5.3 (not flash), so a stored low must survive; keying the
        collapse to the high slot (glm-5.3-flash) discards it
```

Restored, same test unchanged (`lane-verify-green.log`): `ok ... 0.997s`.

**One mapping, shared by call and not by comment.** The alias→slot literals live in exactly one
switch, `template.GLMSlotForModel`. Both halves call it — `GLMSlotEffortForModel` (effort) and
`GLMSlotModelOrHigh` (model) — and `internal/cli`'s `glmSlotEffortForModel` is now a one-line
delegation, so its doc comment's claim that this is the ONE alias/slot pairing in the tree is true
by call graph rather than by assertion. `TestGLMSlotHalvesAgree` fixes that, and the implementing
agent verified the guard bites by pinning the effort half's switch to the high slot — eight lines
of failure across the four aliases and five no-slot inputs (`mutant-halves-agree.log`).

The two halves deliberately differ on a no-slot input, and the guard pins that difference too:
effort returns `""` (the caller keeps its prefs chain), model falls back to the high slot
(pre-repair behaviour preserved).

## The display half, held

`internal/cli/model.go:117` and `internal/web/agentfm.go:314` still key on `Models.High`. The
implementing agent repaired them to key on the agent's own resolved alias, with a coherent
rationale: one rule — "the alias in play → its slot" — applied by the wire to the session alias and
by a display row to that row's agent alias.

The lane reverted both, because the rationale rests on a premise the repository contradicts:

```
internal/web/fieldsets.templ:539
  "Under a GLM backend every sub-agent inherits the session model (llm.glm.models),
   so effort is the only per-agent axis ... The runtime delivers one session-wide
   reasoning value; these states record per-agent intent."
```

If a sub-agent inherits the SESSION model, its own alias is not what it runs on, and keying a chip
to the agent's alias reports a slot the agent never occupies. If instead each spawn carries its own
alias and routes through that slot's `ANTHROPIC_DEFAULT_*_MODEL`, the agent's alias is exactly
right and the shipped sentence above is the stale claim.

Both readings are internally consistent and they select opposite repairs:

| premise | correct display key | old High-slot keying is right when | agent-alias keying is right when |
|---|---|---|---|
| sub-agents inherit the session model | the session alias (not available to either surface) | the session runs opus | never |
| each spawn routes on its own alias | the agent's alias | every agent is opus | always |

Which premise holds is a fact about how the orchestrator spawns sub-agents under a GLM backend —
runtime behaviour outside this repository — so it cannot be settled by reading this tree, and the
lane did not settle it by reasoning. Shipping a keying whose premise is contradicted by a shipped
sentence would put a third statement into a disagreement that already has two.

The wire has no such ambiguity: the session model is the session model, `resolveMainSessionModel`
resolves it from prefs, and it is demonstrably not always opus.

**Escalated to the lead as a decision item.** No SPEC or CHANGELOG sentence was touched (t442 owns
that surface).

## Baseline-attribution

Measured by the lane in this run, in this worktree, at `e50964ad3` + working changes, AFTER the
display half was reverted:

    go build ./...                                → rc 0
    go vet ./internal/cli/ ./internal/template/ ./internal/web/   → rc 0
    go test ./internal/cli/ -count=1              → ok  545.360s
    go test ./internal/web/ -count=1              → ok    3.506s
    go test ./internal/template/ -count=1         → FAIL 30.943s  (inherited, below)
    go test ./internal/template/ -run TestGLMSlot → 4/4 PASS

**The template failure is inherited from the base, not produced here.** `TestManifestHashFormat`
reports `CATALOG_HASH_UNSTABLE` for `sync-auditor` (stored `f1b4487f…`, computed `545d03d9…`,
source `.claude/agents/moai/sync-auditor.md`). Attribution is measured, not argued: this branch's
diff contains none of that test's inputs —

    git status --short .claude/agents/moai/sync-auditor.md internal/template/catalog.yaml \
                       internal/template/catalog_tier_audit_test.go internal/template/catalog_tree_hash.go
      → empty
    git diff --name-only | grep -E '\.claude/|catalog'
      → no match

and the hashing code is untouched, so the computation is identical to the base's. **This means
`develop`'s tip is currently red on `internal/template`** — reported separately to the lead; it
looks like an agent-file edit landed without the catalog hash being regenerated.

## Gaps

- **CI has not run.** No push, per lane discipline. Full suite and the darwin/windows matrix are
  unobserved.
- **No live `moai glm` launch was exercised.** The wire repair is verified at
  `buildEnvForGLMLaunch` — the function that composes `ANTHROPIC_REASONING_EFFORT` — not by
  observing z.ai receive it. What a real launch does end to end is unobserved here.
- **No controlled side-by-side observation of wire vs display.** The implementing agent stated this
  explicitly and the lane did not close it. It is moot for this commit, since the display half is
  not being shipped, but it stays open for whoever ships it.
- Test expectations: none moved. One web fixture premise changed, and that change is reverted along
  with the rest of the display half.

## Residual-risk

- `GLMSlotModelOrHigh` falls back to the high slot for an alias that claims no slot (`""`,
  `inherit`, a raw GLM id, `opusplan`). That preserves pre-repair behaviour exactly, which is what
  makes the change safe — and it also means a future routing alias that ought to own a slot would
  be silently absorbed into the high-slot fallback rather than failing loudly. `TestGLMSlotHalvesAgree`
  pins the current five no-slot inputs, so adding a slot is a visible edit, but a NEW alias arriving
  without a test is not caught.
- The repair makes the wire correct for the slot; it does not change the delivery granularity. The
  reasoning value remains session-global, which is the premise the held display half turns on.

# Design — SPEC-MODEL-MATRIX-CONFIG-001

Design decisions for the config retirement and the two effort channels. Ordered by decision-reversibility: the channel mechanics that are expensive to revise come first; the config removals last.

Inherited from `SPEC-MODEL-PROFILE-MATRIX-002/design.md` §B.1 (partial), §C, and the config-relevant rejected alternatives. The matrix data model (§A) stayed with `SPEC-MODEL-MATRIX-CORE-001`; the wizard design (§D) went to `SPEC-MODEL-MATRIX-SURFACES-001`; the documentation strategy and guard realignment (§E, §F) to `SPEC-MODEL-MATRIX-DOCS-001`.

---

## §A Effort actualization — the two channels

### §A.1 Why two channels exist

The matrix carries `{model, effort}` per cell, but the two values reach the runtime through different mechanisms depending on which orchestration primitive spawns the agent:

```
                      model                     effort
Agent tool      →     runtime arg               frontmatter file       ← Channel A
Workflow agent()→     opts.model                opts.effort            ← Channel B
```

This asymmetry is the whole reason M3 exists. `SPEC-MODEL-PROFILE-MATRIX-001` assumed a single channel and concluded effort was uninjectable; the conclusion was correct for the `Agent` tool and wrong for `Workflow`.

The two channels are **independent, not redundant**. Channel B does not read frontmatter, so under a dynamic workflow the effort is correct even if Channel A never ran; and Channel A is the only path for an `Agent`-tool spawn, which has no `effort` parameter at all. Neither substitutes for the other.

### §A.2 Channel A — frontmatter rewrite

Function shape (name illustrative; the milestone owns the final naming):

```
ApplyAgentEffort(projectRoot string, profile string, overrides map[string]ModelEffort) error
```

Behaviour:

1. Enumerate `.md` files under the **deployed** `.claude/agents/moai/`.
2. For each file, derive the agent name from the file stem.
3. Skip when the agent has an `agent_overrides[agent].effort` entry (REQ-MPME-015).
4. Skip when the agent is not in the matrix (defensive; a user-added file under the MoAI namespace).
5. Write the matrix cell's effort via the existing frontmatter patcher, passing the model argument as "leave unchanged".
6. Never touch `internal/template/templates/`.

**Point 5 is the load-bearing constraint.** The existing `Patch(path, model, effort string, deleteEffort bool)` takes both keys; M3 must invoke it in a mode that leaves `model:` alone. If the existing signature cannot express "leave model unchanged", that capability is added — it is **not** acceptable to pass the current value back in, because a read-modify-write would resurrect the concrete-model pin the moment the read is wrong. The failure is silent in the happy path, which is exactly what makes it dangerous.

**Point 1 rather than "iterate matrix keys"** is also load-bearing: `Explore` is mapped but has no file, so a key-driven iteration erroring on a missing file breaks on every run (REQ-MPME-017).

**Failure posture**: a rewrite failure on one agent should not abort the whole init/update. Log and continue; the profile is a preference, not a correctness requirement, and a half-applied profile is strictly better than a failed `moai update`.

### §A.3 Channel B — Workflow lookup route

REQ-MPME-019 needs a route by which a workflow script obtains `{model, effort}` for an agent under the active profile. `moai model profile --json` already computes exactly this, including the GLM overlay. The design is therefore **expose, don't build**: document the JSON as the lookup contract and, if needed, add an agent-name filter flag so a script does not have to parse the full 11-entry array.

What the workflow script then does:

```
agent(prompt, { agentType: 'manager-develop', model: <cell.model>, effort: <cell.effort> })
```

**Shape dependency**: `SPEC-MODEL-MATRIX-CORE-001` decides whether the `Group` key is dropped from this JSON. Elevating the JSON to a consumed contract before that decision lands would ship a contract and then break it.

**Range note**: the Workflow effort set includes `max`, while the matrix uses at most `xhigh`. No matrix cell can therefore produce an out-of-range `opts.effort` value.

### §A.4 Ordering hazard on `moai update`

```
WRONG:  apply-effort → deploy-templates      (deploy overwrites the effort lines)
RIGHT:  deploy-templates → apply-effort      (REQ-MPME-014)
```

`internal/cli/update.go`'s `--profile` flag path already calls `ApplyProfile` after the deploy step, so the correct ordering exists at that seam and the effort application rides along with it. The other three seams (`init.go`, the `update.go` wizard path, `web/agentfm.go`) must each be checked individually — the `init` seam has no prior deploy to race, and the web seam has no deploy at all.

Getting this wrong does not fail loudly: the user's profile silently reverts on each update, which presents as intermittent rather than broken.

### §A.5 Template baseline re-set — a prerequisite, not a side effect

REQ-MPME-013 requires the template tree to carry the Medium-profile efforts, and REQ-MPME-022's "no-op at medium in this repo" holds only once that baseline is in place. Several of the ten current template values differ from the incoming Medium column.

Required ordering:

```
1. matrix lands                                      (SPEC-MODEL-MATRIX-CORE-001, M1)
2. template agent frontmatter effort values re-set   (SPEC-MODEL-MATRIX-CORE-001, M1 step 7)
3. make build
4. effort application wired                          (this SPEC, M3)
5. only now is a medium-profile run a no-op
```

Skipping step 2 produces a repo where every `moai update` rewrites several agent files back and forth — visible churn that would look like a bug. This is why the dependency on the predecessor is a hard `depends_on` rather than a soft ordering preference.

---

## §B Config retirement

### §B.1 Removed public surface

| Symbol | Disposition |
|---|---|
| `RouteModelFor(specTier, phase, perfTier)` | deleted |
| `ModelRoutingProfiles` type + config field | deleted |
| `validRoutingModels` map | deleted |
| `llm.profiles` config block (both copies) | deleted |
| `model_routing_profiles` config block (both copies) | deleted |

The group symbols (`GroupSpecAuditors` … , `agentGroupMembership`, `AgentGroup`) are also deleted, but by `SPEC-MODEL-MATRIX-CORE-001` — they belong to the matrix, not to the config axis.

### §B.2 Why deletion rather than migration

The 36-cell axis has no production consumer. Migrating dead code into the new matrix would be pure cost, and would give the new matrix a second shape to satisfy. Its tests are deleted rather than retargeted for the same reason: a test whose subject no longer exists has nothing to assert, and retargeting it would invent a new assertion under the guise of preserving an old one.

### §B.3 The `profiles:` mirror and the migration branch

Per-agent overrides already provide the user-editability the mirror was justified by, at one cell of granularity instead of one group. Removing the mirror leaves the Go constant and its fidelity test — two copies of the literal rather than four.

The open question is whether the non-representable branch of REQ-MPME-007 can ever fire. A group-keyed cell expands to 1-3 agent overrides mechanically, so representability is arguably always true under the old six-group shape. If it is always true, the warning branch is dead code shipped as a safety net. The alternative design — unconditional migration plus a summary notice, with no warning branch — is simpler and has no unreachable path.

Deciding this is a run-phase clarification (`research.md` §F), not a plan-phase assumption, because the answer depends on whether any non-group-shaped `profiles:` block can exist in a user project at all.

### §B.4 Load tolerance is a compatibility requirement

REQ-MPME-008 makes the loader treat a legacy `profiles:` or `model_routing_profiles` block as inert rather than failing. This is not politeness: a hard load failure would break every existing project the moment it upgraded, before the user had any chance to migrate. Tolerance is the mechanism that makes the removal safe to ship at all.

---

## §C Rejected alternatives

| Alternative | Why rejected |
|---|---|
| Migrate the 36-cell axis to the new matrix | It has no production consumer; migrating dead code is pure cost (§B.2). |
| Keep `llm.yaml profiles:` as an editable mirror | The mirror is the third and fourth copies of a literal that must stay in lock-step. Per-agent overrides already provide the user-editability the mirror was justified by, at 1 cell of granularity instead of 1 group. |
| Write both `model:` and `effort:` in the rewrite | This is precisely what got the predecessor retired. |
| Read-modify-write frontmatter (preserving model by echoing it back) | A wrong read silently re-pins the model. Passing "leave unchanged" is the only form that cannot regress (§A.2). |
| Rewrite the template tree too | Would make every user's `moai update` diff carry the maintainer's profile choice. |
| Build a second lookup for Channel B | `moai model profile --json` already computes it, GLM overlay included; a second implementation would be a third copy of resolver logic (§A.3). |
| Reject a legacy config block on load | Breaks existing projects on upgrade with no migration window (§B.4). |
| Retarget the orphaned `RouteModelFor` tests instead of deleting them | A test whose subject no longer exists cannot be retargeted without inventing a new assertion (§B.2). |
| Whole-block sync the two `llm.yaml` copies | Would import the local-only `glm.models` alias keys, `haiku` included, into the template tree. |

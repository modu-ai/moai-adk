# Design — SPEC-MODEL-PROFILE-MATRIX-002

Design decisions for the agent-direct profile matrix and effort actualization. Ordered by decision-reversibility: the data-model and contract changes that are expensive to revise come first; mechanical surfaces come last.

---

## §A Data model — the direct matrix

### §A.1 Shape change

Today:

```
profile → group → {model, effort}      (3 × 6 = 18 cells)
agent → group                          (10 entries, Explore absent)
```

After M1:

```
profile → agent → {model, effort}      (3 × 11 = 33 cells)
```

The second map disappears entirely. The indirection hop `agent → group → cell` collapses to `agent → cell`.

### §A.2 Why direct rather than a reduced group set

An alternative was considered: keep grouping but re-partition into the two surviving pairs plus seven singletons. Rejected — a 9-group partition of 11 agents is not an abstraction, it is a rename with extra machinery. The grouping layer earned its keep when 6 groups covered 10 agents (a 1.67:1 compression); at 9:11 (1.22:1) it costs more than it saves, and every future per-agent cell assignment would re-split it again.

### §A.3 Structural invariants (mechanically assertable)

| Invariant | Assertion |
|---|---|
| Cell count | `3 profiles × 11 agents = 33`, no profile column missing an agent |
| Model closed set | every cell model ∈ `{fable, opus, sonnet}` |
| Effort closed set | every cell effort ∈ `{low, medium, high, xhigh}` |
| No haiku | no cell model is `haiku` |
| No in-matrix inherit | no cell model is `inherit` |
| Profile-invariant trio | `manager-docs`, `manager-git`, `Explore` have identical cells in all three columns |
| Display-order agreement | the agent display order and the matrix key set are the same 11 names |

The last invariant is new and worth asserting: today `profileMatrixAgentOrder` has 11 entries while `agentGroupMembership` has 10, and that asymmetry is exactly what made `Explore` an accidental `inherit`. With one map, the asymmetry cannot recur — but a test should still pin that the display order and the matrix agree, because they remain two literals.

### §A.4 Unmapped-agent path

`inherit` leaves the matrix but survives as the **absence** signal:

```
lookup(profile, agent):
    if agent ∈ agent_overrides         → override cell, injectable = true
    if agent ∈ matrix[profile]         → matrix cell,   injectable = true
    if agent ∈ matrix[medium]          → medium cell,   injectable = true   (unknown-profile fallback)
    otherwise                          → {inherit, ""}, injectable = false
```

The boolean's name should change with its meaning. `hasGroup` is now a lie — there are no groups. `injectable` (or `mapped`) states what the caller actually branches on: whether to pass a `model` argument at spawn time. Per constraint C-7 the `inherit` value must never be passed to the `Agent` tool, whose enum does not contain it.

### §A.5 Explore's asymmetry

`Explore` is mapped in the matrix but has no file on disk. It therefore participates in exactly one of the two effort channels:

| Channel | `Explore` |
|---|---|
| `model` runtime arg (Agent tool) | applies — `sonnet` is injected |
| frontmatter `effort:` rewrite | **does not apply** — no file |
| `opts.effort` (Workflow tool) | applies — `medium` is injected |

M3's effort application must therefore iterate **files present on disk** rather than **matrix keys**, and skip silently on absence (REQ-MPM2-047). Iterating matrix keys and erroring on a missing file would break on `Explore` every run.

---

## §B Contract changes rippling out of §A

### §B.1 Removed public surface

| Symbol | Disposition |
|---|---|
| `GroupSpecAuditors` … `GroupGit` | deleted |
| `agentGroupMembership` | deleted |
| `AgentGroup(agent) (string, bool)` | deleted |
| `RouteModelFor(specTier, phase, perfTier)` | deleted |
| `ModelRoutingProfiles` type + config field | deleted |

`ProfileMatrixAgents()`, `DefaultProfileMatrix()`, `ResolveAgentModelEffort()`, and `ApplyProfile()` survive. `DefaultProfileMatrix()`'s return type changes shape (outer key stays profile; inner key becomes agent name instead of group key) — the Go type is identical (`map[string]map[string]config.ModelEffort`), so this is a **silent semantic change with no compiler signal**. Its doc comment must state the inner key is now an agent name, and `DefaultProfileMatrix()` currently has zero production call sites, which bounds the blast radius to tests.

### §B.2 `moai model profile` report shape

`modelProfileEntry` carries `Agent`, `Group`, `Model`, `Effort`, `GLMModel`, `GLMReasoning`. The `Group` field loses its source.

Two options, recorded as an open question in `research.md` §F:

- **Drop the field** — cleanest; changes the `--json` shape. No consumer was found, but `--json` is a documented public surface and REQ-MPM2-049 proposes making it the Workflow-path lookup route, which argues for shaping it deliberately now rather than twice.
- **Retain as `"-"`** — preserves shape; ships a permanently meaningless key.

Design preference: **drop it**, and do so in M1 rather than deferring, because M3 (REQ-MPM2-049) elevates this JSON to a consumed contract. Shipping a contract with a vestigial key and then removing it later is two breaking changes instead of one.

### §B.3 Test contract amendments

The two structurally-driven amendments (as opposed to value-drift amendments):

1. **Split the inherit test.** `Explore` asserts `{sonnet, medium}, injectable=true`; an arbitrary name such as `some-user-agent` asserts `{inherit, ""}, injectable=false`. Keeping them in one loop is what allowed `Explore` to be silently unmapped.
2. **Retarget the override-precedence test.** Its second half asserts "a sibling *in the same group* is unaffected". With groups gone, the meaningful assertion is "an override on agent A does not perturb agent B", which is a strictly weaker but still correct property. Pick a B whose cell differs from A's so the assertion cannot pass vacuously.

`TestResolveAgentModelEffort_ConfigProfilesOverrideDefault` is deleted, not amended — it asserts the `llm.profiles` precedence step that M2 removes.

---

## §C Effort actualization — the two channels

### §C.1 Why two channels exist

The matrix carries `{model, effort}` per cell, but the two values reach the runtime through different mechanisms depending on which orchestration primitive spawns the agent:

```
                      model                     effort
Agent tool      →     runtime arg               frontmatter file       ← M3 rewrite
Workflow agent()→     opts.model                opts.effort            ← M3 lookup route
```

This asymmetry is the whole reason M3 exists. SPEC-001 assumed a single channel and concluded effort was uninjectable; the conclusion was correct for the `Agent` tool and wrong for `Workflow`.

### §C.2 Channel A — frontmatter rewrite

Function shape (name illustrative; the milestone owns the final naming):

```
ApplyAgentEffort(projectRoot string, profile string, overrides map[string]ModelEffort) error
```

Behavior:

1. Enumerate `.md` files under the **deployed** `.claude/agents/moai/`.
2. For each file, derive the agent name from the file stem.
3. Skip when the agent has an `agent_overrides[agent].effort` entry (REQ-MPM2-045).
4. Skip when the agent is not in the matrix (defensive; a user-added file under the MoAI namespace).
5. Write the matrix cell's effort via the existing frontmatter patcher, passing the model argument as "leave unchanged".
6. Never touch `internal/template/templates/`.

Point 5 is the load-bearing constraint. The existing `Patch(path, model, effort string, deleteEffort bool)` takes both keys; M3 must invoke it in a mode that leaves `model:` alone. If the existing signature cannot express "leave model unchanged", that capability is added — it is not acceptable to pass the current value back in, because a read-modify-write would resurrect the concrete-model pin the moment the read is wrong.

**Failure posture**: a rewrite failure on one agent should not abort the whole init/update. Log and continue; the profile is a preference, not a correctness requirement, and a half-applied profile is strictly better than a failed `moai update`.

### §C.3 Channel B — Workflow lookup route

REQ-MPM2-049 needs a route by which a JS workflow script obtains `{model, effort}` for an agent under the active profile. `moai model profile --json` already computes exactly this, including the GLM overlay. The design is therefore **expose, don't build**: document the JSON as the lookup contract and, if needed, add an agent-name filter flag so a script does not have to parse the full 11-entry array.

What the workflow script then does:

```
agent(prompt, { agentType: 'manager-develop', model: <cell.model>, effort: <cell.effort> })
```

Note this path does **not** read frontmatter — so under Mode 6 the effort is correct even if Channel A never ran. The two channels are independent, not redundant: neither substitutes for the other.

### §C.4 Ordering hazard on `moai update`

```
WRONG:  apply-effort → deploy-templates      (deploy overwrites the effort lines)
RIGHT:  deploy-templates → apply-effort      (REQ-MPM2-044)
```

At `internal/cli/update.go:469` `ApplyProfile` is already called after the deploy step, so the correct ordering exists at that seam and the effort application rides along with it. The other three seams (`init.go:608`, `update.go:1495`, `web/agentfm.go:92`) must each be checked individually — the `init` seam has no prior deploy to race, and the web seam has no deploy at all.

### §C.5 Template baseline re-set — a prerequisite, not a side effect

REQ-MPM2-043 requires the template tree to carry the Medium-profile efforts. Per `research.md` §B.6, four of the ten current template values differ from the incoming Medium column. So the template baseline must be **re-set as part of M1/M3**, and only after that does REQ-MPM2-112's "no-op in this repo at the default profile" hold.

Required ordering within the milestone:

```
1. matrix lands (M1)
2. template agent frontmatter effort values re-set to the new Medium column
3. make build
4. effort application wired (M3)
5. only now is a medium-profile run a no-op
```

Skipping step 2 produces a repo where every `moai update` rewrites four agent files back and forth — visible churn that would look like a bug.

---

## §D Wizard and the value-vocabulary seam

The wizard's `model_policy` values are `high`/`medium`/`low`; the profile vocabulary is `max`/`medium`/`low`; `NormalizeToTier` bridges them (`high`→`max`).

Two designs:

| | Change values to max/medium/low | Keep high/medium/low |
|---|---|---|
| Normalizer hop | removed | retained |
| User-visible vocabulary | matches `llm.profile`, `moai model profile`, docs | mismatched (a user picking "High" gets `profile: max`) |
| Risk | any stored/scripted answer of `high` must still normalize | none |

Design preference: **change the values**, keep `NormalizeToTier` as a tolerant reader for legacy `high`. The mismatch is a real user-facing confusion — the whole point of M4 is to stop the question from lying about what it does — and the normalizer already handles the backward-compatible direction.

Localization: the question needs `ko`/`ja`/`zh` entries for title, description, and each option's label and description. The existing `translations.go` `ko` block omits `model_policy` entirely, so today the question renders English in every locale. The new entries must be added for all three locales in one change (REQ-MPM2-063/064).

### §D.1 Option copy — what the descriptions must and must not say

Must say: which subscription tier the profile targets.
Must not say: that a higher profile is stronger or produces better results — §A.2 of `spec.md` establishes that ordering is false.

The honest framing is access-based: the Max profile uses models available on the higher subscription tiers; the Low profile restricts to models available on the entry tier. Cost and quality do **not** move monotonically with the profile name, and the option text must not imply they do.

---

## §E Documentation strategy (M6)

### §E.1 The naming-inversion disclosure

This is the single most important documentation output and should not be buried. It needs one canonical statement, written once and referenced from the other surfaces rather than paraphrased four times per locale:

> `max` / `medium` / `low` name the **subscription tier** whose models the profile draws on — not a performance grade. Under the v1.1 leaderboard the Max profile is both cheaper per task and higher-scoring than the Medium profile.

Placement: `advanced/profile-matrix.md` is the natural home (it is the page that renders the matrix). README and `multi-llm/model-policy.md` link to it.

### §E.2 The Max-Opus rationale

Separate from §E.1 and easy to conflate with it. The statement:

> The Max profile assigns Opus to `manager-develop` and `builder-harness` deliberately, because those agents' failures are expensive to recover from. This is a quality-first choice, not a benchmark optimum — the leaderboard would favour Fable for both.

Without this, a reader who has just absorbed §E.1 will read the Max column's Opus cells as an error.

### §E.3 Contradiction resolution order

`multi-llm/model-policy.md` and `advanced/profile-matrix.md` currently contradict each other. Rather than editing both toward each other, designate one as authoritative:

- `advanced/profile-matrix.md` — **authoritative**: renders the 33-cell matrix, the naming disclosure, the Max-Opus rationale.
- `multi-llm/model-policy.md` — **narrative**: explains how to pick a profile for your plan, links to the matrix page, carries no per-agent table of its own.

Deleting the per-agent table from the narrative page is what makes the contradiction structurally impossible to recur, and it is what lets the zh copy's Haiku column (REQ-MPM2-084) be removed rather than corrected.

### §E.4 The rules-file self-contradiction

`.claude/rules/moai/development/model-policy.md` carries both the stale "all workers Sonnet 5 fixed" tier table and the modern resolver section. The stale table also names `model_routing_profiles` as "the 3-tier config SSOT" — a block M2 deletes. The table is therefore not merely outdated, it will reference a nonexistent artifact. Delete the stale section; keep and extend the resolver section. Byte-parity twin must be edited in the same commit (C-3).

### §E.5 4-locale execution

Route through the `oss-docs` harness (R-4). Its canonical-locale chain and same-PR 4-locale obligation are the existing mechanism for exactly this failure mode; hand-editing 4 locales × 6 surfaces is where parity breaks.

---

## §F Guard realignment (M7)

The haiku-residual rule's surface list is edited in both directions in one change:

| Action | Surface | Reason |
|---|---|---|
| add | web-console agentfm model option set | REQ-MPM2-070 removes the value; the guard prevents its return |
| add | v4manifest tier-suggestion table | REQ-MPM2-071 removes the value; same |
| remove | `model_routing_profiles` in workflow.yaml | block deleted by M2 |
| remove | `validRoutingModels` in model_routing.go | file's routing map deleted by M2 |

Surfaces 1 (agent frontmatter) and 2 (`claude_models` in llm.yaml) are unchanged, as are all four exemptions.

A guard that scans for a deleted artifact is not merely dead — it is actively misleading, because a future reader will infer the artifact still exists. That is why REQ-MPM2-102 couples the removals to the additions rather than deferring them.

---

## §G Rejected alternatives

| Alternative | Why rejected |
|---|---|
| Rename profiles to match the inverted ordering | Breaks `llm.profile` values, the CLI flag, wizard values, config files in the field, and every doc surface simultaneously. Disclosure achieves the same reader outcome at a fraction of the blast radius. |
| Keep groups, re-partition to 9 | 9:11 compression is not an abstraction (§A.2). |
| Migrate the 36-cell axis to the new matrix | It has no production consumer; migrating dead code is pure cost. |
| Keep `llm.yaml profiles:` as an editable mirror | The mirror is the third and fourth copies of a literal that must stay in lock-step. Per-agent overrides already provide the user-editability the mirror was justified by, at 1 cell of granularity instead of 1 group. |
| Write both `model:` and `effort:` in the rewrite | This is precisely what got the predecessor retired (`research.md` §C). |
| Rewrite the template tree too | Would make every user's `moai update` diff carry the maintainer's profile choice. |
| Read-modify-write frontmatter (preserving model by echoing it back) | A wrong read silently re-pins the model. Passing "leave unchanged" is the only form that cannot regress (§C.2). |
| Defer the naming-inversion disclosure to a follow-up | Shipping a matrix whose column names assert a false ordering, with no disclosure, is the failure this SPEC exists to prevent. |

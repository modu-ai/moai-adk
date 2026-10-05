# Design — SPEC-MODEL-MATRIX-CORE-001

Design decisions for the agent-direct profile matrix. Ordered by decision-reversibility: the data-model change that is expensive to revise comes first.

Inherited from `SPEC-MODEL-PROFILE-MATRIX-002/design.md` §A, §B, and the matrix-relevant rejected alternatives. The effort-actualization design (§C of the original) travelled to `SPEC-MODEL-MATRIX-CONFIG-001`; the wizard design (§D) to `SPEC-MODEL-MATRIX-SURFACES-001`; the documentation strategy and guard realignment (§E, §F) to `SPEC-MODEL-MATRIX-DOCS-001`.

---

## §A Data model — the direct matrix

### §A.1 Shape change

Before:

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

The last invariant is the new one and is worth asserting: before M1, `profileMatrixAgentOrder` had 11 entries while `agentGroupMembership` had 10, and that asymmetry is exactly what made `Explore` an accidental `inherit`. With one map the asymmetry cannot recur — but a test should still pin that the display order and the matrix agree, because they remain two literals.

**Ownership note**: the property test asserting all seven invariants is authored by `SPEC-MODEL-MATRIX-DOCS-001` (M7), which owns the guard realignment. This SPEC owns the invariants themselves and the per-invariant behavioural assertions in `acceptance.md`; the single consolidated property test lands downstream.

### §A.4 Unmapped-agent path

`inherit` leaves the matrix but survives as the **absence** signal:

```
lookup(profile, agent):
    if agent ∈ agent_overrides         → override cell, injectable = true
    if agent ∈ matrix[profile]         → matrix cell,   injectable = true
    if agent ∈ matrix[medium]          → medium cell,   injectable = true   (unknown-profile fallback)
    otherwise                          → {inherit, ""}, injectable = false
```

The boolean's name should change with its meaning. `hasGroup` is now a lie — there are no groups. `injectable` (or `mapped`) states what the caller actually branches on: whether to pass a `model` argument at spawn time. Per constraint C-4 the `inherit` value must never be passed to the `Agent` tool, whose enum does not contain it.

### §A.5 Explore's asymmetry

`Explore` is mapped in the matrix but has no file on disk. It therefore participates in exactly one of the two effort channels:

| Channel | `Explore` |
|---|---|
| `model` runtime arg (Agent tool) | applies — `sonnet` is injected |
| frontmatter `effort:` rewrite | **does not apply** — no file |
| `opts.effort` (Workflow tool) | applies — `medium` is injected |

The consequence for the effort application — iterate files present on disk rather than matrix keys — is a `SPEC-MODEL-MATRIX-CONFIG-001` obligation. It is recorded here because the asymmetry originates in this SPEC's data model, and CONFIG inherits it as a constraint rather than discovering it.

---

## §B Contract changes rippling out of §A

### §B.1 Removed public surface

| Symbol | Disposition |
|---|---|
| `GroupSpecAuditors` … `GroupGit` | deleted |
| `agentGroupMembership` | deleted |
| `AgentGroup(agent) (string, bool)` | deleted |

`ProfileMatrixAgents()`, `DefaultProfileMatrix()`, `ResolveAgentModelEffort()`, and `ApplyProfile()` survive. `DefaultProfileMatrix()`'s return type changes shape (outer key stays profile; inner key becomes agent name instead of group key) — the Go type is identical (`map[string]map[string]config.ModelEffort`), so this is a **silent semantic change with no compiler signal**. Its doc comment must state the inner key is now an agent name, and `DefaultProfileMatrix()` has zero production call sites, which bounds the blast radius to tests.

`RouteModelFor` and `ModelRoutingProfiles` are also deleted, but by `SPEC-MODEL-MATRIX-CONFIG-001` (M2) — they belong to the 36-cell axis, not to the matrix.

### §B.2 `moai model profile` report shape

`modelProfileEntry` carries `Agent`, `Group`, `Model`, `Effort`, `GLMModel`, `GLMReasoning`. The `Group` field loses its source.

Two options, recorded as an open question in `research.md` §F:

- **Drop the field** — cleanest; changes the `--json` shape. No consumer was found, but `--json` is a documented public surface and `SPEC-MODEL-MATRIX-CONFIG-001` proposes making it the Workflow-path lookup route, which argues for shaping it deliberately now rather than twice.
- **Retain as `"-"`** — preserves shape; ships a permanently meaningless key.

Design preference: **drop it**, and do so in this SPEC rather than deferring, because CONFIG elevates this JSON to a consumed contract. Shipping a contract with a vestigial key and then removing it later is two breaking changes instead of one.

### §B.3 Test contract amendments

The two structurally-driven amendments (as opposed to value-drift amendments):

1. **Split the inherit test.** `Explore` asserts `{sonnet, medium}, injectable=true`; an arbitrary name such as `some-user-agent` asserts `{inherit, ""}, injectable=false`. Keeping them in one loop is what allowed `Explore` to be silently unmapped.
2. **Retarget the override-precedence test.** Its second half asserts "a sibling *in the same group* is unaffected". With groups gone, the meaningful assertion is "an override on agent A does not perturb agent B", which is a strictly weaker but still correct property. Pick a B whose cell differs from A's so the assertion cannot pass vacuously.

`TestResolveAgentModelEffort_ConfigProfilesOverrideDefault` is deleted rather than amended — but by `SPEC-MODEL-MATRIX-CONFIG-001`, since it asserts the `llm.profiles` precedence step M2 removes.

---

## §C The S0 record as a design artifact

S0 is not a code change; its output is a record. Its design constraint is that the record must be **reconstructible by a later reader**, which is why REQ-MPMC-003 pins the effort level of each row alongside the four metrics, and why REQ-MPMC-002 pins the estimate form (point value vs interval) rather than only a number.

The token-column semantics requirement exists for the same reason: `README.md` ships `$/solved` and `Tokens/solved`, which are **derived** metrics. A record capturing raw output tokens without saying so leaves a downstream author unable to tell whether a derived column can be reconstructed from it or must be re-read.

**Placement note**: S0 sits in this SPEC rather than travelling with the milestone it blocks, because placing it with M6 would push that successor to 27 requirements and 26 criteria — over the Tier L ceiling the split exists to satisfy. The blocking relation survives structurally as `SPEC-MODEL-MATRIX-DOCS-001`'s `depends_on` edge on this SPEC, and DOCS states in its own body that its blocking precondition lives in a predecessor and is undischarged.

---

## §D Rejected alternatives

| Alternative | Why rejected |
|---|---|
| Keep groups, re-partition to 9 | 9:11 compression is not an abstraction (§A.2). |
| Rename profiles to match the inverted ordering | Breaks `llm.profile` values, the CLI flag, wizard values, config files in the field, and every doc surface simultaneously. Disclosure achieves the same reader outcome at a fraction of the blast radius. |
| Make `Explore` an ordinary unmapped agent | Preserves the bug: the display order already carries `Explore`, so the two literals disagree and the disagreement is invisible. |
| Retain the `Group` JSON key as `"-"` | Ships a permanently meaningless key into a contract a downstream SPEC is about to elevate to consumed status (§B.2). |
| Treat the plan-phase leaderboard probe as discharging S0 | The probe returned rows that corroborate the user reading for one model and contradict it for two, most likely because the two sources read different per-effort rows. A lead, not evidence. |
| Move S0 into the DOCS successor so it travels with the milestone it blocks | Would make that successor 27 REQ / 26 AC — over the Tier L ceiling this split exists to satisfy (§C). |

# Research Note — SPEC-CC2178-MODEL-POLICY-REPAIR-001 M1

> Run-phase M1 research deliverable. Covers REQ-MPR-013/014/015 (`[1m]` re-verification),
> REQ-MPR-012 (effort-map deferral rationale), REQ-MPR-016/017 (task-triage decision),
> and AC-MPR-003 task 4 (Default-model JSON key confirmation).

## §1. `[1m]` Re-Verification Verdict (REQ-MPR-013, AC-MPR-011)

### Verdict: **STILL-ACTIVE (conservative)**

The `[1m]` context-entitlement inheritance constraint that forbids per-agent
`model:` pins is treated as **still active** at CC 2.1.178. Per EX-01, per-agent
model pinning remains out-of-scope for this SPEC. The Default-model routing
lever (`availableModels` / `enforceAvailableModels`) is the only confirmed-safe
cost lever.

### Evidence (fetched 2026-06-16 via GitHub REST API + canonical CC CHANGELOG)

**Issue state (GitHub API `api.github.com/repos/anthropics/claude-code/issues/<N>`):**

| Issue | Title | State | Closed_at | Updated_at | Labels |
|-------|-------|-------|-----------|------------|--------|
| #45847 | skill with `model:` frontmatter fails from Opus 4.6 `[1m]` parent | **closed** | 2026-04-13 | 2026-05-07 | `duplicate` |
| #51060 | Subagent `model: opus` fails with "1M context requires extra usage" even when Extra Usage enabled | **closed** | 2026-05-26 | 2026-05-26 | `bug, area:model, area:agents, stale` |
| #36670 | Team teammates don't inherit `[1m]` context variant from leader | **open** | — | 2026-06-02 | `bug, has repro, platform:macos, area:agents, stale` |

**CC CHANGELOG entries touching the `[1m]` / model-pin class (2.1.170-2.1.178):**

- **2.1.172**: "Fixed sessions using 1M context without usage credits getting permanently stuck — the session now automatically compacts back under the standard context limit." → Recovery fix for the stuck-session *symptom*, NOT a fix for the spawn-time *entitlement-inheritance failure*.
- **2.1.172**: "Fixed model IDs getting a doubled 1M-context suffix (e.g. `[1M][1m]`) when `ANTHROPIC_DEFAULT_OPUS_MODEL` already includes one." → Suffix-normalization fix; does not change the spawn-time entitlement path.
- **2.1.172**: "Fixed `availableModels` restrictions not being applied to subagent model overrides, the agent dispatch model picker, and the advisor model." → Tightens the allowlist surface; orthogonal to `[1m]` inheritance.
- **2.1.173**: "Fixed Fable 5 model names with a `[1m]` suffix not being normalized." → Fable-5-specific normalization.
- **2.1.174**: "Fixed background sessions inheriting another session's `ANTHROPIC_*` provider env (gateway URL, custom headers, `/model` aliases) from the shell that started the background daemon." → Background-session env-inheritance fix; distinct from subagent model-pin entitlement.
- **2.1.175**: "Added `enforceAvailableModels` managed setting — when enabled, the `availableModels` allowlist also constrains the Default model (a Default that would resolve to a disallowed model now falls back to the first allowed model), and user or project settings can no longer widen a managed `availableModels` list." → The `[1m]`-safe cost lever this SPEC wires.
- **2.1.176**: "Fixed `availableModels` enforcement: alias model picks can no longer be redirected to a blocked model via `ANTHROPIC_DEFAULT_*_MODEL` environment variables, and `/fast` now refuses to toggle when it would switch to a model outside the allowlist." → Tightens allowlist; orthogonal to `[1m]` inheritance.

### Verdict rationale

The evidence is **mixed but the conservative-default applies**:

1. #51060 (the direct "subagent `model: opus` spawn fails" issue) was closed on 2026-05-26. However it carries the `stale` label, and **no CC CHANGELOG entry between 2.1.170 and 2.1.178 explicitly fixes the subagent-spawn entitlement-inheritance failure**. The 2.1.172 "stuck session" fix addresses the *symptom* (the session no longer hangs), not the *root cause* (a subagent declaring an explicit `model:` still does not inherit the parent's `[1m]` entitlement). Closure-with-`stale` is ambiguous: it may mean "no longer reproducible on the current plan" or "wontfix for the reporter's plan tier". Without an explicit "fixed" resolution, the conservative reading is that the spawn-time entitlement path is unchanged.

2. #36670 (Team-mode `[1m]` inheritance) is **STILL OPEN** at CC 2.1.178 (updated 2026-06-02). This confirms the Team-mode path is definitively unfixed.

3. The MoAI `model-policy.md` Inherit-by-Default doctrine is built around the *spawn-time entitlement mismatch* class, which spans both single-spawn (#45847, #51060) and Team-mode (#36670) paths. Because the Team-mode path (#36670) is confirmed-open and no changelog entry resolves the single-spawn path's root cause, the constraint must be treated as **still-active** for the purpose of this SPEC.

4. Per acceptance.md EC-01, when the re-verification cannot definitively confirm the constraint is relaxed, the default is "still-active" (conservative — preserves EX-01). This verdict applies that default.

### Implication for per-agent pinning (REQ-MPR-014)

EX-01 remains in force: per-agent `model:` pins are forbidden regardless of the re-verification outcome. Even if the single-spawn path were relaxed, the Team-mode path (#36670) is unfixed, so any per-agent pin would still break Team spawns. Default-model routing via `availableModels` + `enforceAvailableModels` is the only confirmed-safe lever.

### Relaxation conditions for a follow-up SPEC (REQ-MPR-014 / REQ-MPR-015)

A follow-up SPEC (`SPEC-CC2178-PER-AGENT-PIN-RELAXATION-001`, conditional) MAY re-enable per-agent pinning when ALL of the following are confirmed:

- (a) Issue #36670 (Team-mode `[1m]` inheritance) is **closed with an explicit "fixed" resolution** AND a CC CHANGELOG entry confirms Team teammates inherit the leader's `[1m]` entitlement when a teammate declares an explicit `model:`.
- (b) The `availableModels` / `enforceAvailableModels` settings fields are confirmed to allow per-agent `model:` overrides that escape the allowlist WITHOUT breaking `[1m]` inheritance at CC 2.1.178+. (Today, 2.1.172 made `availableModels` apply to subagent overrides — this may CONFLICT with per-agent pins rather than enable them; the follow-up SPEC must verify.)
- (c) A caller-migration path exists for any agent whose pinned model differs from Default (e.g., `manager-spec` at Opus vs. Default Sonnet).

Until ALL three hold, per-agent pinning stays out-of-scope.

## §2. Default-Model JSON Key Confirmation (AC-MPR-003 task 4, REQ-MPR-003)

### Confirmed key: `model` (top-level), constrained by `availableModels` + `enforceAvailableModels`

CC 2.1.175 changelog verbatim: _"Added `enforceAvailableModels` managed setting — when enabled, the `availableModels` allowlist also constrains the Default model (a Default that would resolve to a disallowed model now falls back to the first allowed model)"_.

The Default model in a Claude Code `settings.json` is the top-level `"model"` field. A naive `grep '"model"'` matches multiple lines in the MoAI template (hook `"type": "command"` lines do not contain `"model"`, but the template already pins `"effortLevel"` and other fields — however there is no existing top-level `"model"` key in the committed template, verified by grep on 2026-06-16: `grep -c '"model"' settings.json.tmpl` on the top-level key returns 0). Therefore the AC-MPR-003 verification uses a structured JSON parse to isolate the top-level `"model"` value.

### M4 wiring (authoritative)

The template `internal/template/templates/.claude/settings.json.tmpl` gains three sibling keys at the top level (alongside `effortLevel`, `outputStyle`, etc.):

```json
"model": "sonnet",
"availableModels": ["sonnet", "opus", "haiku"],
"enforceAvailableModels": true
```

Semantics:
- `"model": "sonnet"` — Default model explicitly set to Sonnet (the cost-routing thesis: route the busy-agent cost through Sonnet, not Opus).
- `"availableModels": ["sonnet", "opus", "haiku"]` — allowlist of the 3 CC model aliases (language-agnostic; no Go/internal IDs).
- `"enforceAvailableModels": true` — the allowlist constrains Default resolution; a per-agent `model:` pin that escapes the allowlist is rejected by the runtime (2.1.172 + 2.1.175 + 2.1.176 enforcement chain).

### AC-MPR-003 verification command (pinned by M1 task 4)

```bash
# Render the template to a temp project and inspect the top-level "model" key.
moai init /tmp/mpr-verify-001 --force 2>/dev/null
python3 -c "import json; d=json.load(open('/tmp/mpr-verify-001/.claude/settings.json')); print('model=', d.get('model')); print('availableModels=', d.get('availableModels')); print('enforceAvailableModels=', d.get('enforceAvailableModels'))"
# Expected:
#   model= sonnet
#   availableModels= ['sonnet', 'opus', 'haiku']
#   enforceAvailableModels= True
```

The Python JSON-parse approach is used instead of a raw `grep '"model"'` because post-2.1.175 the settings file carries multiple model-named keys (`model`, `availableModels`) and the template also contains `ANTHROPIC_*` references in hook commands. A structured JSON parse isolates the top-level `"model"` value unambiguously.

If `moai init` is unavailable in the verification environment, the fallback is a direct template render test in `internal/template/settings_test.go` (the M4 RED test) that asserts the rendered JSON has `"model": "sonnet"` at the top level.

## §3. Effort-Map Deferral Rationale (REQ-MPR-012, AC-MPR-010 part d)

### Decision: PRUNE + RECONCILE only; full retirement DEFERRED

`ApplyEffortPolicy` (`internal/template/model_policy.go:134-180`) has **2 production callers**, verified by grep at plan-phase (2026-06-16):

1. `internal/core/project/initializer.go:181` — the `moai init` deployment path.
2. `internal/cli/update.go:2661` — the `moai update` deployment path.

Both callers invoke `ApplyEffortPolicy` to **inject** `effort:` into freshly deployed agent files that lack it (`if effortLineRegex.Match(content) { continue }` — only injects when absent). The hand-authored `effort:` values in current agent files exist *because* the map injected them on a prior `moai init` / `moai update`.

### Regression risk if retired without migration

Retiring `agentEffortMap` + `ApplyEffortPolicy` without migrating the 2 callers means: a fresh `moai init` deployment injects NO `effort:` for reasoning agents (`manager-spec`, `plan-auditor`, `sync-auditor`, `manager-develop`, `builder-harness`). The runtime default (`xhigh` on Opus 4.7+) would apply — but only for agents whose runtime reads a missing `effort:` as "use default". Any agent that behaves differently with an explicit `effort:` vs. its absence silently regresses.

### Follow-up SPEC candidate

`SPEC-CC2178-EFFORT-MAP-RETIREMENT-001` (deferred) — scope:
1. Migrate `initializer.go:181` to an alternative effort-injection path (e.g., hard-code the 5-entry map inline, or move effort injection to a build-time template step).
2. Migrate `update.go:2661` likewise.
3. Remove `agentEffortMap`, `GetAgentEffort`, `ApplyEffortPolicy`, `effortLineRegex`, `frontmatterOpenPrefix`, `insertEffortInFrontmatter`.
4. Update `internal/template/CLAUDE.md` to reflect the removal.

This SPEC (SPEC-CC2178-MODEL-POLICY-REPAIR-001) takes only the SAFE subset: prune the 3 archived phantom keys (`manager-strategy`, `expert-security`, `expert-refactoring`), reconcile map↔file divergence (`plan-auditor`/`sync-auditor` `high` → `xhigh`), and add the missing retained agents (`manager-develop` `xhigh`, `builder-harness` `high`).

## §4. Task-Triage Decision (REQ-MPR-016/017, AC-MPR-012)

### Decision: **DEFERRED**

The per-task triage signal (failure-cost × visual-verifiability, per the research doc's 3-axis model) is **deferred** to a follow-up SPEC.

### Rationale

1. This SPEC's scope is already substantial: 3-axis cost routing alignment (model × effort × cycle_type) + phantom-map cleanup + `[1m]` re-verification + `ResolveCycleType` new symbol + `availableModels` lever + doctrine update. Adding a 4th axis (per-task triage) would push the SPEC past Tier M into Tier L territory without a commensurate increase in deliverable coherence.

2. The triage signal requires a concrete integration point in the harness router (`internal/harness/router/router.go`) that does not yet exist — the router resolves a `Level` from SPEC frontmatter `harness_level` but has no per-task hook. Designing the triage signal means designing the integration point, which is a separate architectural decision.

3. The research doc's triage model (Sonnet+v2 ≈ Opus on diagnostics, +5-7% gap on deep reasoning) is directionally useful but the coding-agent metrics are UNVALIDATED (per the research doc's own caveat). Wiring a triage signal on unvalidated metrics would codify an assumption.

### Follow-up SPEC candidate

`SPEC-CC2178-TASK-TRIAGE-001` (deferred) — scope:
1. Validate the failure-cost × visual-verifiability dimensions against actual MoAI run-phase task distributions.
2. Define the mapping function {opus/full-tdd, sonnet/ddd-lite} concretely.
3. Design the harness router integration point (per-task hook or per-milestone hook).
4. Wire the triage signal into `manager-develop` delegation prompts.

## §5. Research completeness checklist

- [x] #45847 state fetched (GitHub API) → closed/duplicate
- [x] #51060 state fetched (GitHub API) → closed/stale
- [x] #36670 state fetched (GitHub API) → open
- [x] CC 2.1.170-2.1.178 CHANGELOG fetched verbatim (raw.githubusercontent.com)
- [x] `[1m]` verdict recorded (still-active, conservative per EC-01)
- [x] Default-model JSON key confirmed (`model` top-level + `availableModels` + `enforceAvailableModels`)
- [x] AC-MPR-003 verification command pinned (Python JSON-parse of rendered settings)
- [x] Effort-map deferral rationale recorded (2 callers, regression risk, follow-up SPEC name)
- [x] Task-triage decision recorded (deferred, rationale, follow-up SPEC name)

---

Authored: 2026-06-16 (M1, run-phase)
Authored-By-Agent: manager-develop

---
name: moai-workflow-harness
description: >
  V3R4 Self-Evolving Harness lifecycle workflow. Owns all lifecycle verbs
  (status, apply, rollback, disable) entirely within the moai skill body
  using file-system operations — no Go binary subcommand is invoked. Tier-4
  evolution applications are gated by an orchestrator-issued AskUserQuestion
  approval round per the 5-Layer Safety architecture.
user-invocable: false
metadata:
  version: "2.0.0"
  category: "workflow"
  status: "active"
  updated: "2026-05-14"
  tags: "harness, learning, observer, tier-4, safety, evolution, apply, rollback, v3r4, cli-retired"

# MoAI Extension: Progressive Disclosure
progressive_disclosure:
  enabled: true
  level1_tokens: 100
  level2_tokens: 5000

# MoAI Extension: Triggers
triggers:
  keywords: ["harness", "harness status", "harness apply", "harness rollback", "harness disable", "tier 4", "harness proposal", "harness evolve"]
  agents: ["moai-harness-learner"]
  phases: ["harness"]
---

<!-- @MX:NOTE: [AUTO] V3R4 contract — workflow body owns the lifecycle entirely via file-system operations. No `moai harness` CLI subcommand is invoked. The Go file `internal/cli/harness.go` is retained only as a deprecation marker awaiting downstream physical removal per SPEC-V3R4-HARNESS-001 (BC-V3R4-HARNESS-001-CLI-RETIREMENT). -->
<!-- @MX:REASON: [AUTO] V3R3-era workflow body shelled out to `moai harness <verb>` cobra subcommand. V3R4 retires that CLI verb path (REQ-HRN-FND-001, REQ-HRN-FND-002) and consolidates all lifecycle execution into the slash command + workflow body surface (REQ-HRN-FND-003). -->

# Workflow: harness — V3R4 Self-Evolving Harness Lifecycle

Purpose: Surface the harness learning subsystem (observer, 4-tier proposal ladder, 5-layer safety pipeline) to the user. This workflow IS the implementation — every verb (`status`, `apply`, `rollback`, `disable`) is executed by file-system reads and writes inside this workflow body. No Go binary subcommand is invoked. The Go CLI factory in `internal/cli/harness.go` remains in the tree as a deprecation marker only; it is never registered into the cobra command tree (REQ-HRN-FND-001, REQ-HRN-FND-002).

## Authoritative Sources

- SPEC (active): `SPEC-V3R4-HARNESS-001` (foundation, supersedes the three V3R3 harness SPECs)
- Constitution: `.claude/rules/moai/design/constitution.md` §5 (5-Layer Safety) and §2 (Frozen/Evolvable zones)
- AskUserQuestion contract: `.claude/rules/moai/core/askuser-protocol.md` (canonical reference)
- Agent boundary: `.claude/rules/moai/core/agent-common-protocol.md` § User Interaction Boundary
- Config: `.moai/config/sections/harness.yaml` (`learning.enabled` field)
- State roots:
  - `.moai/harness/usage-log.jsonl` (PostToolUse observation append-only log)
  - `.moai/harness/proposals/` (pending Tier-4 proposals)
  - `.moai/harness/learning-history/snapshots/<ISO-DATE>/` (pre-modification snapshots)
  - `.moai/harness/learning-history/applied/` (applied proposals, used for rate-limit window)
  - `.moai/harness/learning-history/tier-promotions.jsonl` (tier classifier transitions)
  - `.moai/harness/learning-history/frozen-guard-violations.jsonl` (L1 audit log)

## Related Rules and References

- AskUserQuestion ToolSearch preload procedure: `.claude/rules/moai/core/askuser-protocol.md` § ToolSearch Preload Procedure
- Free-form circumvention prohibition: `.claude/rules/moai/core/askuser-protocol.md` § Free-form Circumvention Prohibition
- Orchestrator-subagent boundary: `.claude/rules/moai/core/agent-common-protocol.md` § User Interaction Boundary
- Lessons (auto-memory): `lessons.md` workflow + harness categories when present

---

## Tier-4 Application Gate (Orchestrator-Only AskUserQuestion)

[HARD] Tier-4 application of any harness evolution proposal MUST be gated by an orchestrator-issued `AskUserQuestion` round before any file modification occurs (REQ-HRN-FND-004, REQ-HRN-FND-015). The workflow body itself is executed in the orchestrator's main context; `AskUserQuestion` is invoked here as the orchestrator's tool. Subagents reachable from this workflow MUST NOT call `AskUserQuestion`; if a subagent needs user input it returns a structured blocker report and the orchestrator re-runs the round (canonical reference: `.claude/rules/moai/core/askuser-protocol.md`).

### Canonical Four-Option Pattern

Before invoking `AskUserQuestion`, preload the schema:

```
ToolSearch(query: "select:AskUserQuestion")
```

Then invoke with at least three options (Apply / Defer / Reject — additional options permitted), the first option marked `(권장)` or `(Recommended)`:

```
AskUserQuestion({
  questions: [{
    question: "Harness Tier-4 evolution proposal ready to apply (id: <proposal_id>).\n\nTarget: <target_path>\nField: <field_key>\nProposed value: <new_value>\nObservation count: <observation_count>\n\nApply this change?",
    header: "Tier-4 Apply",
    options: [
      { label: "Apply (권장)", description: "Run the 5-Layer Safety pipeline. On pass, create a snapshot and write the change. Snapshot lives under .moai/harness/learning-history/snapshots/<ISO-DATE>/." },
      { label: "Modify", description: "Open the proposal for manual edit before re-presenting. Proposal file is preserved." },
      { label: "Defer", description: "Re-record the proposal for later approval. The 7-day rate-limit window starts only when an Apply is accepted." },
      { label: "Reject", description: "Move the proposal to .moai/harness/learning-history/rejected/. No file is modified." }
    ]
  }]
})
```

### Rate-Limit Enforcement (REQ-HRN-FND-012)

Before opening the AskUserQuestion round, count entries in `.moai/harness/learning-history/applied/` whose `applied_at` falls within the last 7 days. If the count is ≥ 1, the new candidate MUST be deferred and recorded with a `deferred_at` timestamp; do NOT invoke `AskUserQuestion`. The rate-limit floor of 1 application per 7-day window is invariant per REQ-HRN-FND-018 — any future adaptive expansion may raise the limit but never lower it.

---

## Input

`$ARGUMENTS` — the first word is the verb (`status`, `apply`, `rollback`, `disable`, or empty for help). Remaining arguments are passed to the verb dispatcher.

## Verb Routing

[HARD] Extract the FIRST WORD from `$ARGUMENTS`. Match against the verb table below. If empty or unrecognized, show the help block (Phase 0).

| Verb | Workflow-body operations | Tools used | Purpose |
|------|---------------------------|------------|---------|
| `status` | Read JSONL state files, aggregate counts | Read, Bash (jq) | Inspect tier distribution, rate-limit window, pending proposal count, Tier-4 reach rate |
| `apply` | Layer 1 (FrozenGuard path-prefix check) → Layer 4 (rate-limit window check) → AskUserQuestion → snapshot creation → file write → audit log append | Read, Write, Edit, Bash, AskUserQuestion | Surface next Tier-4 proposal via AskUserQuestion, apply on approval through 5-layer pipeline |
| `rollback <YYYY-MM-DD>` | Read snapshot `manifest.json`, copy files back | Read, Write, Bash | Restore byte-identical pre-modification state from snapshot |
| `disable` | AskUserQuestion confirm → Edit harness.yaml `learning.enabled: false` | AskUserQuestion, Edit | Turn off observer; observer becomes a no-op (REQ-HRN-FND-009) |

---

## Phase 0: Help (default when no verb)

If `$ARGUMENTS` is empty or matches `help` / `--help` / `-h`, render the verb table above plus a one-line summary in the user's `conversation_language`. Then stop. Do not perform any read or write.

---

## Phase 1: Pre-Execution Sanity Check

Before executing any verb, verify:

1. Project root is detected (`.moai/config/config.yaml` exists). If absent, abort with guidance to run `moai init` first.
2. Harness learning subsystem state: read `.moai/config/sections/harness.yaml` `learning.enabled` field.
   - If `enabled: false` and verb is anything other than `status`, surface a warning that the subsystem is disabled via AskUserQuestion (continue / abort). The first option is `Continue (권장)` only when the verb is `rollback` (rollback should remain functional even with learning disabled); for `apply`, first option is `Abort (권장)`.
3. For `rollback`: the date argument is mandatory. If missing, abort with usage hint `/moai:harness rollback <YYYY-MM-DD>`.

---

## Phase 2: Verb Dispatch

### 2.1 status (file-IO only — no binary)

Operations:

1. Read `.moai/harness/usage-log.jsonl` (line-count via `wc -l` or progressive Read).
2. Read `.moai/harness/learning-history/tier-promotions.jsonl` (group by tier: observation / heuristic / rule / auto_update).
3. Read `.moai/harness/learning-history/applied/` directory listing for Tier-4 application count within the last 7 days (REQ-HRN-FND-016 telemetry).
4. Compute Tier-4 reach rate = (unique patterns with `tier=auto_update` in tier-promotions.jsonl) / (unique patterns in usage-log.jsonl) × 100. Patterns are keyed by `event_type + subject + context_hash` (REQ-HRN-FND-010).
5. Read `.moai/harness/proposals/` directory listing for pending proposal count.
6. Render the result as a Markdown table in the user's `conversation_language`:

```
Harness Learning Status (V3R4)
  Enabled:                  <bool from harness.yaml learning.enabled>
  Log entries:              <line count of usage-log.jsonl>
  Unique patterns:          <distinct event_type+subject+context_hash>

Tier Distribution:
  observation (T1):         <int>
  heuristic   (T2):         <int>
  rule        (T3):         <int>
  auto_update (T4):         <int> patterns (pending user approval)

Telemetry (REQ-HRN-FND-016):
  Weekly Tier-4 applications (rolling 7-day): <int>
  Tier-4 reach rate:                          <pct>%

Pending Proposals: <int>
  [PENDING] <proposal_id>: <skill_path> (<field>)
  ...
```

No user prompt. Stop after rendering.

### 2.2 apply (5-Layer Safety pipeline, file-IO only)

Operations:

1. **Layer 4 (Rate Limiter) pre-screen**: Scan `.moai/harness/learning-history/applied/` for any entry with `applied_at` within the last 7 days. If count ≥ 1, defer:
   - Append `{ "deferred_at": <ISO-8601>, "reason": "rate-limit window active", "proposal_id": <id> }` to the candidate proposal's metadata.
   - Render "Tier-4 rate-limit window active — proposal <id> deferred" and STOP. Do NOT invoke AskUserQuestion (REQ-HRN-FND-012).
2. **Load next pending proposal**: Read `.moai/harness/proposals/` directory; pick the oldest pending entry (`.json` payload). If none, render "No Tier-4 proposals awaiting approval" and stop.
3. **Layer 1 (Frozen Guard) pre-screen**: Read the proposal's `target_path`. Match against the FROZEN prefix list:
   - `.claude/agents/moai/`
   - `.claude/skills/moai-`
   - `.claude/rules/moai/`
   - `.moai/project/brand/`
   If any prefix matches, append a JSONL entry to `.moai/harness/learning-history/frozen-guard-violations.jsonl` with at minimum: ISO-8601 timestamp, the attempted target path, the proposal id (as calling subject), and a rejection rationale (REQ-HRN-FND-006, REQ-HRN-FND-014). Then move the proposal to `.moai/harness/learning-history/rejected/` and stop. Do NOT raise an error to the user; the rejection is silent except for the audit log.
4. **Layer 3 (Contradiction Detector) pre-screen**: Out of scope for the V3R4 foundation SPEC. Downstream `SPEC-V3R4-HARNESS-005` introduces principle-based scoring; this workflow body documents the contract assertion (REQ-HRN-FND-017) and treats Layer 3 as a no-op pass-through for the foundation release.
5. **Tier-4 Application Gate**: `ToolSearch(query: "select:AskUserQuestion")` → `AskUserQuestion` with the canonical four-option pattern from the section above. The first option `Apply (권장)` MUST carry the `(권장)` / `(Recommended)` suffix per `.claude/rules/moai/core/askuser-protocol.md` § Option Description Standards.
6. **On `Apply` selection**:
   - **Layer 2 (Canary Check)**: Out of scope for the V3R4 foundation SPEC. Downstream `SPEC-V3R4-HARNESS-006` introduces multi-objective scoring with score-drop check; the workflow body treats Layer 2 as a no-op pass-through for the foundation release. The constitution §5 L2 layer remains documented as the binding contract.
   - **Create snapshot** (REQ-HRN-FND-007): Create directory `.moai/harness/learning-history/snapshots/<ISO-DATE>/` (ISO-8601 timestamp). For each file the proposal will touch, Read the current contents, compute a content hash, and Write a byte-identical copy into the snapshot directory. Write `manifest.json` recording absolute target paths and content hashes. The snapshot MUST be complete before any modification of the target file.
   - **Apply the change**: Read the proposal's `new_value`, perform the field replacement on `target_path` (typically a frontmatter `description` or `triggers` edit on a `my-harness-*` skill body). Use the `Edit` tool with exact `old_string` / `new_string` match.
   - **Move the proposal**: Move the proposal file from `.moai/harness/proposals/` to `.moai/harness/learning-history/applied/`. Append an `applied_at` ISO-8601 timestamp and the snapshot directory path to the JSON payload.
   - **Render outcome**: Confirm `Applied. Snapshot: <path>. Run /moai:harness status to verify the new tier distribution.`.
7. **On `Modify` selection**: Render guidance to open the proposal file at `.moai/harness/proposals/<id>.json` in the user's editor. Do NOT modify state. Re-presenting requires the next `/moai:harness apply` invocation after edit.
8. **On `Defer` selection**: Append `deferred_at` timestamp to the proposal metadata. Stop.
9. **On `Reject` selection**: Move the proposal to `.moai/harness/learning-history/rejected/`. Stop.

Edge case (EDGE-001): If the snapshot directory cannot be created (disk-full, permission), abort BEFORE any modification. Inform the user via the orchestrator's response. No partial state is left on disk.

Edge case (EDGE-006): Concurrent `/moai:harness apply` invocations are out of scope for the foundation SPEC. The orchestrator's AskUserQuestion contract is sequential by design.

### 2.3 rollback

Operations:

1. Validate that the date argument matches `YYYY-MM-DD` or full ISO-8601 (`YYYY-MM-DDThh:mm:ssZ`). If invalid format, abort with usage hint.
2. Locate snapshot directory: `.moai/harness/learning-history/snapshots/<date>/`. If the directory does not exist, render a diagnostic message naming the missing snapshot and STOP. Do NOT modify any file (REQ-HRN-FND-008 explicit unwanted behavior on missing snapshot).
3. Read `manifest.json` from the snapshot directory.
4. For each entry in the manifest: Read the snapshot copy, Write to the absolute target path. Verify the post-write content hash matches the snapshot hash (byte-identical restoration).
5. Render confirmation message naming the snapshot date and the restored file count.

Safety: rollback does NOT delete the post-application state — it overlays the snapshot. The current state is preserved at `.moai/harness/learning-history/snapshots/<rollback-timestamp>/pre-rollback/` before the rollback overwrite.

### 2.4 disable

Operations:

1. `ToolSearch(query: "select:AskUserQuestion")` → `AskUserQuestion` to confirm intent (this is the only place in `disable` that prompts the user directly, because the consequence is global subsystem deactivation):
   - Question (Korean conversation_language example): `Harness 학습 서브시스템을 비활성화하시겠습니까? 옵서버가 이벤트 수집을 중단하고 신규 제안이 생성되지 않습니다.`
   - Options:
     - `Disable (권장)` — Set `learning.enabled: false`. Snapshots and proposals are preserved.
     - `Keep enabled` — Stop without changes.
     - `Disable temporarily (1h)` — Not supported in the V3R4 foundation SPEC; render guidance to re-enable manually via config edit and stop.
2. On `Disable` selection: Use `Edit` tool on `.moai/config/sections/harness.yaml` to set `learning.enabled: false`. Comments and key ordering MUST be preserved (YAML round-trip discipline).
3. On `Keep enabled`: stop without changes.
4. On `Disable temporarily`: render guidance and stop.

Observer no-op contract (REQ-HRN-FND-009): When `learning.enabled: false`, the PostToolUse observer hook MUST be a complete no-op — it does not read, write, or append to `.moai/harness/usage-log.jsonl`. Existing log entries MUST NOT be deleted.

---

## Phase 3: Post-Execution Summary

After any successful verb execution, render a one-paragraph summary in the user's `conversation_language` covering:

1. What was done (verb + key result).
2. Where the artifact lives (`.moai/harness/...` path).
3. Suggested next step (e.g., after `apply`: "Run `/moai:harness status` to verify the new tier distribution.").

---

## Error Handling

| Symptom | Likely cause | Recovery |
|---------|-------------|----------|
| `harness.yaml not found` | Project not initialized or harness disabled at scaffold | Run `moai init` first; harness subsystem is created on first run |
| `Error: learning.enabled: false` on `apply` | Subsystem disabled by user or admin | Re-enable manually via editing `.moai/config/sections/harness.yaml` (no `enable` verb yet — deferred to a downstream SPEC) |
| `Error: no snapshot matching <date>` on `rollback` | Snapshot directory absent for requested date | Run `/moai:harness status` to list available snapshot dates under `.moai/harness/learning-history/snapshots/` |
| AskUserQuestion schema not loaded | Deferred tool preload missed | The workflow body explicitly preloads via `ToolSearch(query: "select:AskUserQuestion")` before each `apply` or `disable` invocation |
| Frozen-guard violation log permission failure | Disk-full or permission denied on `.moai/harness/learning-history/` | Per EDGE-003, the L1 Frozen Guard MUST still block the modification; log emission is best-effort audit |

---

## Cross-references

- SPEC (active): `.moai/specs/SPEC-V3R4-HARNESS-001/spec.md` (REQ-HRN-FND-001 ~ REQ-HRN-FND-018)
- SPEC (superseded by V3R4-001; preserved as historical reference):
  - `.moai/specs/SPEC-V3R3-HARNESS-LEARNING-001/spec.md` (REQ-HL-001 ~ REQ-HL-012 — 4-tier ladder preserved)
  - `.moai/specs/SPEC-V3R3-HARNESS-001/spec.md` (Meta-harness skill + generated artifacts)
  - `.moai/specs/SPEC-V3R3-PROJECT-HARNESS-001/spec.md` (16Q socratic interview + 5-Layer wiring)
- Skill: `.claude/skills/moai-harness-learner/SKILL.md` (Tier-4 surfacing companion — text-annotated only per SPEC-V3R4-HARNESS-001 §10 exclusion #10)
- Skill: `.claude/skills/moai-meta-harness/SKILL.md` (project-specific harness generation — text-annotated only)
- README: `.moai/harness/README.md` (subsystem overview)
- Attribution: `.claude/rules/moai/NOTICE.md` (Apache-2.0 attribution to revfactory/harness)
- Constitution: `.claude/rules/moai/design/constitution.md` §5 (5-Layer Safety — preserved verbatim per REQ-HRN-FND-005)

<!-- Verifies REQ-HRN-FND-001/002: no Bash invocation of `moai harness` anywhere in this workflow body. -->
<!-- Verifies REQ-HRN-FND-003: all four verbs (status/apply/rollback/disable) implemented in this body. -->
<!-- Verifies REQ-HRN-FND-004/015: Tier-4 application is gated by orchestrator-issued AskUserQuestion. -->
<!-- Verifies REQ-HRN-FND-005: 5-Layer Safety preserved (L2/L3 documented as deferred no-op to downstream SPECs; L1/L4/L5 active). -->
<!-- Verifies REQ-HRN-FND-006/014: FROZEN zone path-prefix block + audit log emission. -->
<!-- Verifies REQ-HRN-FND-007/008: pre-modification snapshot + rollback to byte-identical state. -->
<!-- Verifies REQ-HRN-FND-009: observer no-op contract documented (enforced by hook code path, Wave C). -->
<!-- Verifies REQ-HRN-FND-010/011: PostToolUse baseline schema + 4-tier ladder preserved. -->
<!-- Verifies REQ-HRN-FND-012/018: Tier-4 rate-limit floor of 1 per 7-day window. -->
<!-- Verifies REQ-HRN-FND-016: telemetry exposure (weekly Tier-4 apply count + reach rate) via `status` verb. -->

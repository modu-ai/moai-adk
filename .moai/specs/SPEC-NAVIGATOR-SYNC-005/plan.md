# plan.md — SPEC-NAVIGATOR-SYNC-005 (BAS Epic M3 — Falconer Fix)

> Implementation plan for the M3 Fix layer. Tier L — the most complex BAS milestone.
> Orders milestones/sections by **decision-reversibility** (lead with highest-change-likelihood
> decisions: the split-architecture + approval-surface UX), deferring mechanical steps to the bottom.

## §A. Context

### §A.1 What M3 is

M3 is the **Fix** element of the Falconer 3-element loop (detect → route → fix). M1 (Detect, `status: completed`) surfaces *that* a change touched bound graph rows; M2 (Route, `status: completed`) surfaces *which* code path owns the follow-up + *what* action closes it. M3 closes the loop: it takes the M2 work-items + the M1 changed-paths + the M0 graph context, **incrementally regenerates only the stale doc subtrees** via an AI-drafted rewrite (the CodeWiki `--compare-to <commit>` pattern), and presents the draft for 1-click human approval. On approval, the corrected subtrees land atomically; on rejection, nothing changes. The doc map stays alive.

### §A.2 Precondition satisfied

`depends_on: [SPEC-NAVIGATOR-SYNC-002, SPEC-NAVIGATOR-SYNC-004]` — both M1 and M2 are MERGED. The design report's core safety point — *"Detect 없는 Fix는 잘못된 곳을 고친다"* (§6 line 374) — is satisfied. M3 may proceed.

### §A.3 The design-report authority

- §3.3(C) line 274: the Fix box ("AI-drafted incremental rewrite, CodeWiki --compare-to 패턴, changed subtrees만")
- §5 Q2 line 327: *"CodeWiki의 --update / --compare-to <commit> + metadata.json 증분 모델이 '갱신'의 검증된 기법 — Fix 요소가 이 패턴을 가져온다."*
- §5 Q3 line 337: *"Tier 1·3 Blueprint/Symbol — detect→route→fix 루프. CodeWiki --compare-to로 변경된 하위 트리만 재생성."*
- §6 M3 line 371-375: *"CodeWiki --compare-to <commit> 패턴 — changed subtrees만 재생성, AI가 초안 작성, 1-click 승인. M1·M2가 검증된 뒤에만."*
- §7 line 399: M3 = Tier L ("검증 필요")
- §9 line 443: Fix 자동화율 ≥ 50%

### §A.4 PRESERVE targets (do NOT modify — REQ-NS5-005/006)

- `internal/navigator/sync/` (M0 graph-join layer)
- `internal/navigator/detect/` + `internal/hook/navigator_detect*.go` + `internal/hook/post_tool.go` (M1 detect)
- `internal/navigator/route/` + `internal/cli/navigator_route.go` (M2 route)
- `internal/navigator/tiers/` (M4 tiers)
- `internal/mx/` (scanner + SpecAssociator)
- `internal/navigator/astx/` (tree-sitter symbol extraction — 001/003)
- `.claude/skills/moai-workflow-project/scripts/navigator-{audit,regen,enrich}.sh` (the 3 predecessor chains)

## §B. The Tier-L Architecture Decision (agent-vs-engine split)

> **This is the highest-change-likelihood decision in the SPEC — review it first.**
> Defended in full in `design.md §A`; summarized here for the plan-level review.

### §B.1 The decision

M3 is a **split-architecture system** with three separable layers:

| Layer | Lives in | Does what | Calls an LLM? |
|---|---|---|---|
| **(1) Diff-scope + draft-request engine** | `internal/navigator/fix/` (Go) | Diffs current tree to baseline commit; identifies stale subtrees from M2 work-items + M1 detect + M0 graph; emits `request.json` manifest | NO — pure deterministic Go |
| **(2) AI draft** | orchestrator → `manager-develop` delegation | Reads `request.json` + live doc surfaces; produces rewritten doc subtrees as a draft at the staging surface | YES — rides Claude Code's `Agent()` runtime |
| **(3) Approval + apply** | orchestrator → `AskUserQuestion` gate; apply is Go (`internal/navigator/fix/apply.go`) | Presents the draft for 1-click approval; on approval, atomic-renames the approved subtrees to the live doc surfaces | NO (approval is human; apply is deterministic Go) |

### §B.2 Why the split (the two load-bearing reasons)

1. **moai-adk has no Go-embedded LLM client in the navigator subsystem.** The `moai` binary is a Go CLI that does deterministic work (hooks, scans, subcommands). The LLM is reached ONLY through Claude Code's `Agent()` subagent runtime, which a Go CLI binary cannot invoke at process-runtime. CodeWiki is a standalone tool with its own LLM client; moai-adk is a harness ON TOP of Claude Code. The AI-draft step therefore MUST ride the orchestrator's `Agent()` spawn, not a Go-embedded inference path. (design.md §A.1)

2. **CLAUDE.md mandates the orchestrator delegates writes to manager-develop.** M3's draft IS a write (a doc rewrite). If the Go engine produced the draft text itself, it would bypass the orchestrator's write-delegation discipline. The split keeps the write authority where CLAUDE.md put it: the orchestrator delegates the draft to manager-develop; the Go engine never writes doc content. (design.md §A.2)

### §B.3 What this means for run-phase

The `navigator-fix` CLI invocation does NOT produce a draft on its own. It produces a `request.json` (layer 1) and signals "draft-request ready." The orchestrator then spawns `manager-develop` with the `request.json` path injected (layer 2), receives the draft, and runs the `AskUserQuestion` approval gate (layer 3). If `navigator-fix` runs OUTSIDE a Claude Code session (e.g. a bare `moai navigator-fix` in a shell), layer 2 cannot fire (no `Agent()` runtime) — the CLI exits 0 with a "draft-request produced, run inside `/moai project` to generate the AI draft" message (REQ-NS5-009 fail-open: no-LLM-runtime is a degraded mode, not an error).

### §B.4 Alternatives rejected (design.md §E carries the full evaluation)

- **(alt-A) Go-embedded LLM client** — rejected: duplicates model-config, bypasses write-delegation, adds secret-management surface (REQ-NS5-007 forbids).
- **(alt-B) Full regenerate-and-replace on approval** — rejected: discards human edits to untouched sections, expensive, non-incremental (the 001/003/M0 contract; M3 is the incremental complement, not a replacement — REQ-NS5-003/006).
- **(alt-C) Auto-apply without approval** — rejected: collapses M3's safety value proposition (REQ-NS5-008 forbids; the approval gate IS the feature).

## §C. Technical Approach

### §C.1 On-demand CLI trigger (no PostToolUse — the cadence decision)

**Decision**: M3 registers as a Hidden cobra subcommand `navigator-fix`, invoked from `/moai project`. NO PostToolUse hook branch.

**Why no PostToolUse** (defended): a per-edit draft would (a) usually find no work-items to consume (M2 is on-demand — work-items don't exist until the engineer runs `navigator-route`), and (b) flood the engineer with one draft per keystroke — the cadence is wrong for a human-gated, batched operation. M1 already owns the real-time advisory surface (`systemMessage`); M3 owns the coarser-cadence, human-gated fix surface. The engineer invokes `navigator-fix` when they have accumulated drift they are ready to close — not on every keystroke. This mirrors M2's on-demand decision (REQ-NS4-001, plan.md §C.1).

### §C.2 Diff-scope engine (layer 1 — deterministic Go)

**Input**: M2 `work-items.json` (`owner_path` + `action` + `source_kind`) + M1 detect JSONL (`changed_path` set) + M0 `nav-graph.json` (graph edges) + the baseline commit.

**Baseline-commit resolution** (priority order):
1. Explicit `--compare-to <commit>` flag (user override).
2. `nav-graph.json` `provenance.extract_commit_sha` (the last known-consistent doc-map state — M0's own provenance; the default).
3. `HEAD~1` (degenerate fallback for a fresh checkout with no nav-graph; logged as degraded).

**Diff computation**: `git diff --name-only <baseline>..HEAD` → the set of paths changed since the baseline. Intersect with (M1 `changed_path` ∪ M2 `owner_path`) → the "touched" set. For each touched path, traverse the M0 graph to find the bound doc subtrees (nodes + edges whose `source_path` matches) → the "stale subtree" set.

**Output**: the `diff_scope[]` array — deduplicated, sorted `(doc_surface, subtree_id, stale_reason, work_item_ref)` triples. This is the manifest the AI draft consumes.

### §C.3 Draft-request artifact (layer 1 output — the handoff)

**Path**: `.moai/project/navigator/fix-drafts/<draft-id>/request.json` where `<draft-id>` = SHA-256 of the sorted diff-scope + baseline SHA (deterministic, NOT wall-clock).

**Schema**:
```json
{
  "provenance": {
    "fix_commit_sha": "<git rev-parse HEAD>",
    "baseline_commit_sha": "<resolved baseline>",
    "captured_at": "<git log -1 --format=%cI of fix_commit_sha>"
  },
  "diff_scope": [
    { "doc_surface": "capability-map.md", "subtree_id": "...", "stale_reason": "...", "work_item_ref": {...} }
  ],
  "work_item_refs": [ { "source_kind": "...", "owner_path": "...", "action": "..." } ],
  "draft_instructions": { "per_subtree": [ { "subtree_id": "...", "strategy": "..." } ] }
}
```

**Properties**: idempotent (same HEAD + baseline + inputs → byte-identical `request.json`), provenance-attributed (no wall-clock), fail-open on empty diff-scope (`diff_scope: []` + exit 0 + "doc map consistent" message).

### §C.4 AI draft delegation (layer 2 — orchestrator → manager-develop)

**Trigger**: the `navigator-fix` CLI signals the orchestrator (via exit-0 + a stdout JSON line `{"draft_request_path": "...", "status": "ready"}`) that a draft-request is ready.

**Orchestrator action**: spawns `manager-develop` (cycle_type=`tdd` or `ddd` per `quality.yaml`; domain context: navigator-fix) with the spawn prompt injecting:
- the `request.json` path (the diff-scope + per-subtree strategy hints)
- the live doc surface paths (READ targets)
- the staging draft path (WRITE target — `.moai/project/navigator/fix-drafts/<draft-id>/draft/`)
- the constraint: "produce ONLY the patched subtrees named in `diff_scope[]`; do NOT touch subtrees outside the scope; write each patched subtree as a separate file + a unified-diff preview"

**Output**: the draft at `.moai/project/navigator/fix-drafts/<draft-id>/draft/` — one patched file per stale subtree + a `draft.json` manifest + `*.patch` unified-diff previews.

**Non-overlap**: manager-develop writes ONLY to the staging surface (REQ-NS5-008); it does NOT touch the live doc surfaces (those are the apply step's target, post-approval).

### §C.5 Approval surface (layer 3 — AskUserQuestion gate)

**The orchestrator presents the draft** via `AskUserQuestion` (preload via `ToolSearch` first, per askuser-protocol.md § ToolSearch Preload Procedure) with 4 options:
- **(a) approve + apply** (권장) — apply ALL patched subtrees atomically.
- **(b) approve selected** — apply a subset (engineer selects which).
- **(c) edit then apply** — open the draft for manual editing first.
- **(d) reject** — discard the draft.

**Preview**: the `AskUserQuestion` `preview` field renders the unified-diff previews (the `*.patch` files) so the engineer can see exactly what changes before deciding (askuser-protocol.md § Preview Field Standards — the side-by-side TUI layout).

### §C.6 Apply-on-approval (layer 3 — Go, post-approval)

**Trigger**: the orchestrator receives the approval decision (option a/b/c) and delegates the apply to `internal/navigator/fix/apply.go` (or the orchestrator calls `moai navigator-fix --apply <draft-id> --options <a|b|c>` if option b names a subset).

**Apply semantics**:
- **approval_token write (REQ-NS5-008 sub-clause c4)**: on approval, the orchestrator writes `approval.json` at `fix-drafts/<draft-id>/approval.json` carrying draft-id + approval option + approver + approval timestamp (git-committer-date) + token value (deterministic hash of draft-id + option + request.json provenance, NOT a random nonce). The `--apply` CLI entry point validates this token before applying — a direct shell invocation without a valid token is refused (exit non-zero, no live-doc mutation). This makes the approval a real artifact, not just "someone typed a command".
- **scope-conformance validation (REQ-NS5-013)**: before the gate preview AND before the apply, the orchestrator (or `apply.go`) validates each draft subtree ID ∈ `diff_scope[]`. An over-produced subtree (drafted by manager-develop outside the stale set) is excluded from the gate preview + the apply + logged as a warning. This closes the gap between the spawn-prompt instruction ("produce ONLY diff_scope[] subtrees") and the mechanical enforcement — the draft OUTPUT is validated, not just the Go engine source.
- atomic-rename each approved (and scope-conformant) subtree to its target live doc surface (`.tmp` + `os.Rename`, carrying forward M0's `atomicWrite`). **Apply idempotence (DBT-2)**: the `applied.json` ledger records each applied subtree ID; a resume after a crash-mid-set skips already-applied IDs.
- record `applied.json` ledger at the draft-id staging dir: approver, approval timestamp (git-committer-date), applied subtree IDs, resulting live-doc SHA.
- touch ONLY the approved subtrees; unapproved subtrees in the same live doc are NOT modified (the incremental-not-regenerate-and-replace contract — REQ-NS5-003/006).

## §D. File Touch List + Template-First Map

### §D.1 New files (run-phase creation)

| Path | Purpose | Layer |
|---|---|---|
| `internal/navigator/fix/scope.go` | diff-scope engine (baseline resolution + git-diff + graph traversal) | 1 (Go) |
| `internal/navigator/fix/request.go` | draft-request manifest emission (`request.json`) | 1 (Go) |
| `internal/navigator/fix/apply.go` | apply-on-approval (atomic-rename of approved subtrees) | 3 (Go) |
| `internal/navigator/fix/types.go` | shared types (DiffScopeEntry, DraftRequest, AppliedLedger) | 1 (Go) |
| `internal/navigator/fix/nonoverlap_test.go` | non-overlap grep guard (carries forward M0/M1/M2 pattern) | test |
| `internal/navigator/fix/scope_test.go` | diff-scope unit tests | test |
| `internal/navigator/fix/apply_test.go` | apply-on-approval tests | test |
| `internal/navigator/fix/testdata/fix-corpus/` | fixture corpus (work-items + detect + nav-graph + live docs) | test fixtures |
| `internal/cli/navigator_fix.go` | Hidden cobra subcommand registration | CLI |
| `internal/cli/navigator_fix_test.go` | Hidden-subcommand assertion (mirrors `navigator_tiers_test.go:65-70`) | test |

### §D.2 Runtime artifacts (generated, gitignored)

| Path | Purpose |
|---|---|
| `.moai/project/navigator/fix-drafts/<draft-id>/request.json` | draft-request manifest (layer 1 output) |
| `.moai/project/navigator/fix-drafts/<draft-id>/draft/` | AI-drafted patched subtrees (layer 2 output) |
| `.moai/project/navigator/fix-drafts/<draft-id>/draft/draft.json` | draft manifest (subtree → target mapping) |
| `.moai/project/navigator/fix-drafts/<draft-id>/draft/*.patch` | unified-diff previews (approval-gate preview) |
| `.moai/project/navigator/fix-drafts/<draft-id>/applied.json` | apply ledger (approver + timestamp + applied IDs) |

### §D.3 Template-first distribution

M3 ships NO distributed config under `.claude/` (pure CLI + runtime Go — the expected case, like M2). The AI-draft step rides the orchestrator's `Agent()` runtime, not a distributed config key. REQ-NS5-012 reduces to: "no template path in the diff, no catalog regen required" — documented in the PR body. If a `navigator-fix` config key is later needed in `settings.json`, the template source carries it first per CLAUDE.local.md §2.

## §E. Asset-Reuse Map (design report §4)

| Reused asset | Location | M3 role |
|---|---|---|
| M0 `atomicWrite` pattern | `internal/navigator/sync/write.go` | apply-on-approval atomic-rename (REQ-NS5-008 sub-assertion c1) |
| M0 `Provenance` model | `internal/navigator/sync/types.go` | `request.json` + `applied.json` provenance blocks (no wall-clock) |
| M0 graph edges | `nav-graph.json` `edges[].source_path` + `line_number` | diff-scope subtree identification (the bound rows whose source paths changed) |
| M1 detect `changed_path` | `.moai/state/navigator-detect/*.jsonl` | diff-scope seed (the real-time touched-path set) |
| M2 work-item `action` field | `work-items.json` `work_items[].action` | per-subtree fix-strategy hint (the "what" the AI draft should do) |
| M2 work-item `owner_path` | `work-items.json` `work_items[].owner_path` | diff-scope fan-in dimension (WHERE the stale doc is) |
| `navigator-route` Hidden subcommand | `internal/cli/navigator_route.go` | sibling pattern for `navigator-fix` (REQ-NS5-011) |
| M0/M1/M2 `nonoverlap_test.go` pattern | `internal/navigator/{sync,detect}/nonoverlap_test.go` | `internal/navigator/fix/nonoverlap_test.go` (REQ-NS5-006 AC) |

## §F. Milestones (ordered by decision-reversibility — highest-change-likelihood first)

> Per the SPEC Builder protocol, milestones lead with the decisions most likely to change
> (data-model + UX decisions), deferring mechanical/refactoring steps to the bottom.

### M3.1 — Split-architecture scaffold + diff-scope contract (the Tier-L decision lands here)

- Create `internal/navigator/fix/types.go` (DiffScopeEntry, DraftRequest, AppliedLedger types).
- Create `internal/navigator/fix/scope.go` (baseline-commit resolution + `git diff --name-only` + M0 graph traversal → stale subtree set).
- The `request.json` schema (§C.3) is the data-model decision — review it here first.
- AC: AC-NS5-003 (diff-scope identifies stale subtrees), AC-NS5-004 (draft-request schema + idempotence).

### M3.2 — Draft-request emission + Hidden CLI subcommand

- Create `internal/navigator/fix/request.go` (emit `request.json` at staging surface, idempotent + provenance-attributed).
- Create `internal/cli/navigator_fix.go` (Hidden cobra subcommand, sibling of `navigator-route`).
- AC: AC-NS5-001a (CLI produces draft-request), AC-NS5-001b (no PostToolUse), AC-NS5-011 (Hidden subcommand).

### M3.3 — AI draft delegation wiring (orchestrator → manager-develop)

- The orchestrator-side wiring: `navigator-fix` CLI signals "draft-request ready" via stdout JSON; the orchestrator spawns `manager-develop` with `request.json` injected.
- This milestone is the agent-vs-engine boundary in action — review whether the handoff contract (stdout JSON → orchestrator → Agent() spawn) is correct.
- AC: AC-NS5-007 (zero LLM-client imports in Go engine — the split is mechanically enforced).

### M3.4 — Approval surface (the UX decision — review second)

- The orchestrator's `AskUserQuestion` gate: 4 options (approve + apply / approve selected / edit then apply / reject), `preview` field rendering the `*.patch` unified diffs.
- This is the user-facing UX — the highest-change-likelihood surface after the architecture split.
- AC: AC-NS5-008b (approval surface: AskUserQuestion 4-option gate + preview), AC-NS5-013 (draft scope-conformance: out-of-scope subtree excluded from gate preview + apply — aligned with acceptance.md §B traceability row REQ-NS5-013 → M3.4).

### M3.5 — Apply-on-approval + non-overlap enforcement

- Create `internal/navigator/fix/apply.go` (atomic-rename approved subtrees + `applied.json` ledger).
- Create `internal/navigator/fix/nonoverlap_test.go` (grep guard — carries forward M0/M1/M2 pattern).
- AC: AC-NS5-005a/005b (consumer-only), AC-NS5-006 (non-overlap), AC-NS5-008a (no live-doc mutation without approval), AC-NS5-008c (apply atomic-rename + ledger), AC-NS5-008d (approval_token refusal — `--apply` without a valid token exits non-zero + no live-doc mutation; aligned with acceptance.md §B traceability row REQ-NS5-008 → M3.5).
- **Exit-code scope (REQ-NS5-009 vs REQ-NS5-008 c4)**: fail-open (REQ-NS5-009, AC-NS5-009) is a *degraded* path → **exit 0** + advisory message (no-LLM-runtime, absent inputs, empty diff-scope are NOT errors); token-refusal (REQ-NS5-008 c4, AC-NS5-008d) is a *hard guard* → **exit non-zero** (a shell invocation without a valid approval token is refused, not degraded). The two MUST NOT share an exit-code contract — fail-open degrades and continues; token-refusal refuses and stops.

### M3.6 — Fail-open + coverage measurement (the ≥50% automation-rate gate)

- Fail-open paths across all error modes (work-items absent / detect absent / nav-graph absent / baseline unresolvable / unparseable JSON / schema-invalid / empty diff-scope / no-LLM-runtime).
- The fixture corpus at `internal/navigator/fix/testdata/fix-corpus/` + the `TestFixAutomationRate` test emitting the ratio.
- AC: AC-NS5-009 (fail-open across error modes — table-driven), AC-NS5-010 (≥50% automation rate, mechanically measured), AC-NS5-002 (four read-only inputs), AC-NS5-012 (template-first).

## §G. Coverage Plan (the ≥50% automation-rate measurement)

**Fixture corpus** at `internal/navigator/fix/testdata/fix-corpus/`:
- 10 stale-subtree scenarios spanning the 3 `source_kind`s (audit-missing / audit-orphan / detect) × the 3 `action` strategies (regenerate row / re-link symbol / draft SPEC stub).
- Each scenario carries: a fixture `work-items.json` (the M2 output), a fixture detect JSONL (the M1 output), a fixture `nav-graph.json` (the M0 output), fixture live doc surfaces, and a fixture baseline commit.
- A simulated approval loop records, per scenario, whether the draft was approved unmodified (option a) or needed edit/selection (option b/c) or was rejected (option d).

**Dual-arithmetic** (the floor survives the worst case):
- Happy path: 6 of 10 scenarios approved unmodified → automation rate = 60% ≥ 50% ✓.
- Worst case (4 scenarios need edit): automation rate = 60% still ≥ 50% ✓ (the 4 edit-cases are in the denominator, not the numerator).

**Command**: `go test ./internal/navigator/fix/ -run TestFixAutomationRate -v` — emits the observed automation rate; FAILS (non-zero exit) if `< 50.0`.

## §H. Risks + Anti-Patterns

- **AP-NS5-001 — Go engine grows an LLM call**: a future "optimization" adds an LLM client to `internal/navigator/fix/` to "skip the orchestrator round-trip." Forbidden by REQ-NS5-007; the CI grep guard (zero LLM-client imports) catches it.
- **AP-NS5-002 — Apply without approval**: a `--yes` flag that skips the AskUserQuestion gate. Forbidden by REQ-NS5-008; collapses M3 into regenerate-and-replace with no safety.
- **AP-NS5-003 — Full regen fallback**: when the diff-scope is large (> N subtrees), falling back to regenerating the whole doc artifact. Out of scope (§F); M3 is incremental only. A future threshold-based fallback is a separate SPEC.
- **AP-NS5-004 — Draft touches unapproved subtrees**: the AI draft rewrites a subtree not in the `diff_scope[]`. The non-overlap AC (AC-NS5-006) + the manager-develop spawn-prompt constraint ("produce ONLY the patched subtrees named in `diff_scope[]`") prevent this.
- **AP-NS5-005 — Baseline-commit drift**: the `nav-graph.json` provenance SHA is stale (the graph was regenerated long ago), making the diff-scope too broad. Mitigation: the `--compare-to` flag lets the engineer override; the default degrades gracefully (broader scope = more drafts, not an error).
- **AP-NS5-006 — Approval-gate fatigue**: if M3 drafts too often, the engineer approves without reading (the gate stops working). Mitigation: M3 is on-demand (the engineer invokes it when ready), NOT real-time; the cadence is self-paced.

## §I. Cross-References

- `design.md` §A — the split-architecture decision (full defense).
- `design.md` §B — the approval-surface UX design.
- `research.md` — the CodeWiki `--compare-to` pattern investigation + mapping.
- `acceptance.md` §D — the AC matrix traceable to each REQ.
- `.moai/reports/navigator-redesign-bas-20260805.html` §3.3(C), §4, §5 Q2/Q3, §6 M3, §7, §9 — the design-report authority.
- `.claude/rules/moai/core/askuser-protocol.md` § ToolSearch Preload Procedure + § Preview Field Standards — the approval-gate mechanics.
- `.claude/rules/moai/development/manager-develop-prompt-template.md` § Section A-E — the AI-draft delegation prompt template (Tier L: full 5-section template required).

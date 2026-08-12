# design.md — SPEC-NAVIGATOR-SYNC-005 (BAS Epic M3 — Falconer Fix)

> Tier-L architecture decision document. Defends the three highest-change-likelihood
> decisions in the SPEC: (A) the agent-vs-engine split, (B) the approval-surface UX,
> (C) the diff-scope / baseline-commit selection. Each decision names the alternatives
> rejected + the load-bearing reason the chosen design wins.

## §A. The Split-Architecture Decision (agent-vs-engine)

### §A.1 The question M3 must answer

The design report (§3.3(C) line 274, §6 line 371-375) names the Fix element as *"AI-drafted incremental rewrite, CodeWiki --compare-to 패턴, changed subtrees만 재생성, AI가 초안 작성, 1-click 승인."* The question: **where does the "AI가 초안 작성" (AI writes the draft) step live in moai-adk's architecture?**

CodeWiki is a standalone tool (FSoft-AI4Code, ACL 2026) with its own embedded LLM client — it diffs against a commit, calls its LLM, and produces the draft in one process. moai-adk is NOT a standalone tool; it is a harness ON TOP of Claude Code. The `moai` binary is a Go CLI that does deterministic work (hooks, scans, subcommands), and the LLM is reached ONLY through Claude Code's `Agent()` subagent runtime. A Go CLI binary cannot invoke `Agent()` at process-runtime.

### §A.2 The two load-bearing reasons for the split

**Reason 1 — no Go-embedded LLM client exists in the navigator subsystem.**

The navigator subsystem (`internal/navigator/{sync,detect,route,tiers}/`) is pure deterministic Go. It has no LLM client — no OpenAI/Anthropic SDK, no model-config surface, no API-key management. The LLM is configured in `.moai/config/sections/llm.yaml` + Claude Code's runtime, and reached through `Agent()` spawns the orchestrator issues. Embedding an LLM client in `internal/navigator/fix/` would:

- duplicate the model-config surface that `llm.yaml` + Claude Code already own (two sources of truth for "which model");
- introduce a new secret-management surface (API keys in the Go binary's env — a new attack surface the navigator subsystem does not currently have);
- bypass Claude Code's permission/effort/thinking model (the `Agent()` runtime applies effort levels, Adaptive Thinking, permission modes — none of which a Go-embedded call would get).

The design report itself grounds this: §4 asset-reuse table (line 287-308) lists the reused assets as deterministic Go (astx, enrich parser, atomic-write, provenance model) — NOT an LLM client. The Fix element reuses the deterministic assets and adds the LLM call at the orchestrator layer, where the LLM already lives.

**Reason 2 — CLAUDE.md mandates the orchestrator delegates writes to manager-develop.**

CLAUDE.md §1 Core Identity: *"MoAI is the strategic orchestrator... All tasks must be delegated to specialized agents."* §2 Request Processing Pipeline: the orchestrator composes the execution plan and delegates. The manager-develop agent owns run-phase implementation (write authority). M3's draft IS a write — it produces a doc rewrite. If the Go engine produced the draft text itself, it would bypass the orchestrator's write-delegation discipline, violating the architectural invariant the entire harness rests on.

The split keeps the write authority where CLAUDE.md put it: the orchestrator delegates the draft to `manager-develop`; the Go engine never writes doc content. The Go engine's job is the deterministic scope (diff-scope + draft-request manifest); the orchestrator's job is the AI draft (via `Agent()` → `manager-develop`); the human's job is the approval (via `AskUserQuestion`).

### §A.3 The three separable layers

| Layer | Owner | Input | Output | LLM? |
|---|---|---|---|---|
| **(1) Diff-scope + draft-request** | Go engine (`internal/navigator/fix/`) | M2 work-items + M1 detect + M0 graph + baseline commit | `request.json` manifest at staging surface | NO |
| **(2) AI draft** | orchestrator → `manager-develop` | `request.json` + live doc surfaces | patched subtrees at staging surface (`draft/`) | YES (via `Agent()`) |
| **(3) Approval + apply** | orchestrator → `AskUserQuestion`; apply is Go | the draft + approval decision | approved subtrees atomic-renamed to live docs | NO (human + Go) |

The layers are **separable and independently testable**:
- Layer 1 is testable without an LLM (fixture inputs → expected `request.json`).
- Layer 2 is testable against a fixture `request.json` (the manager-develop delegation reads the fixture + produces a draft; the draft is checked against expected content).
- Layer 3 is testable at the orchestrator-integration level (the apply step reads a fixture draft + a simulated approval → expected live-doc mutation).

### §A.4 The handoff contract (layer 1 → layer 2)

The `navigator-fix` CLI (layer 1) does NOT produce the draft. It produces `request.json` and signals the orchestrator via a stdout JSON line:

```json
{"draft_request_path": ".moai/project/navigator/fix-drafts/<id>/request.json", "status": "ready", "draft_id": "<id>"}
```

The orchestrator consumes this signal, spawns `manager-develop` with the `request.json` path + live-doc paths + staging draft path injected into the spawn prompt, and receives the draft. The handoff is a file-based contract (the `request.json` schema in plan.md §C.3) — NOT an in-process function call. This is deliberate: the Go engine and the orchestrator run in different process contexts (the Go CLI vs. the Claude Code session), and a file-based handoff is the only contract that crosses that boundary without coupling.

### §A.5 The no-LLM-runtime degraded mode

If `navigator-fix` runs OUTSIDE a Claude Code session (e.g. `moai navigator-fix` in a bare shell, or in a CI script), layer 2 cannot fire — there is no `Agent()` runtime. The CLI exits 0 (NOT an error) with the stdout message *"draft-request produced; run inside /moai project to generate the AI draft."* Layer 1 is complete; layer 2 is deferred until the engineer opens a Claude Code session and the orchestrator picks up the `request.json`. This is REQ-NS5-009 fail-open case 009h — a degraded mode, not a failure.

## §B. The Approval-Surface UX Design

### §B.1 Why a gate (not auto-apply)

M3's entire safety value proposition is that it asks before applying. The design report §8 risk grid (line 430-433, "다이어그램≠문서") names the hazard: a generated doc that is never re-updated degrades into a snapshot. M3's renewal infrastructure closes that — BUT only if the renewal is trustworthy. An auto-apply that ships a wrong draft (the LLM hallucinated a capability, mis-linked a symbol, invented a SPEC) is worse than no draft: it corrupts the doc map with confidence. The human approval gate is the trust boundary that makes M3 safe to use.

The design report's *"1-click 승인"* (§6 line 374) is the UX target: the approval should be one click, not a multi-step review. But "one click" presupposes the engineer can SEE what they are approving — which means the draft must be presented with a preview (the unified diff), not blind.

### §B.2 The 4-option gate

The `AskUserQuestion` gate offers exactly 4 options (per askuser-protocol.md § Option Description Standards — each option's `description` states the immediate result + irreversibility):

| Option | Label | What it does | When the engineer picks it |
|---|---|---|---|
| (a) | approve + apply (권장) | atomic-rename ALL patched subtrees to live docs | the draft is correct as-is |
| (b) | approve selected | apply a subset (engineer names which subtrees) | part of the draft is correct, part needs rework |
| (c) | edit then apply | open the draft for manual editing, then apply | the draft is close but needs a small fix |
| (d) | reject | discard the draft (staging dir remains for audit, no apply) | the draft is wrong; re-run later |

The `(권장)` label is on option (a) — the statistically-majority case (the ≥50% automation-rate target IS the rate at which (a) is chosen). Per askuser-protocol.md § Recommendation Placement Principles, the recommendation is grounded in the observed majority, and option (a)'s `description` states the precondition ("Recommended when the draft preview looks correct").

### §B.3 The preview field (unified diffs)

Per askuser-protocol.md § Preview Field Standards, the `preview` field renders a monospace block beside the option list (side-by-side TUI layout). M3 renders the `*.patch` unified-diff previews — one per stale subtree — so the engineer sees exactly what changes before deciding. The preview is truncated at ~12 lines (the preview pane does not scroll); longer diffs are summarized with a "... (N more lines in the full patch)" pointer to the staging dir.

**Single-select only**: the `preview` field is silently dropped when `multiSelect: true` (askuser-protocol.md § Preview Field Standards HARD rule). Option (b) "approve selected" is a single-select follow-up gate (a second `AskUserQuestion` listing the subtrees), NOT a multi-select on the main gate — so the main gate's preview renders correctly.

### §B.4 The orchestrator-only boundary

Per askuser-protocol.md § Channel Monopoly + agent-common-protocol.md § User Interaction Boundary, the `AskUserQuestion` gate is invoked by the **orchestrator only**. The Go engine (`internal/navigator/fix/`) produces the `*.patch` previews and exits; it does NOT invoke `AskUserQuestion` (the subagent-boundary grep AC-NS5-007a + the indirect-verification subagent-boundary check enforce this). The orchestrator reads the draft + previews, constructs the `AskUserQuestion` call (preloaded via `ToolSearch`), and routes the approval decision to the apply step.

## §C. The Diff-Scope / Baseline-Commit Selection

### §C.1 Why a baseline commit (the --compare-to pattern)

CodeWiki's `--compare-to <commit>` (design report §5 Q2 line 327) is the verified technique: diff the current tree against a baseline commit, regenerate ONLY the changed subtrees. The baseline is "the last time the doc map was known-consistent." Without a baseline, M3 would have no definition of "stale" — every subtree would be a candidate, collapsing M3 into full regen.

### §C.2 Baseline resolution priority

1. **Explicit `--compare-to <commit>` flag** — the user override. The engineer knows their context (e.g. "regen everything since the last release tag") and can name the baseline.
2. **`nav-graph.json` `provenance.extract_commit_sha`** — the DEFAULT. This is M0's own provenance: the commit at which the graph was last regenerated. By definition, the doc map was consistent with the tree at that commit (the graph was built from it). Any change since then is a drift candidate. This is the most principled default because it is self-attributing (M0 already records it) and semantically correct (it IS the last-known-consistent state).
3. **`HEAD~1`** — the degenerate fallback. Used when `nav-graph.json` is absent (M0 not yet run) AND no `--compare-to` flag. One commit of history — a poor baseline (most of the tree is "changed"), but it prevents a hard failure on a fresh checkout. Logged as degraded.

### §C.3 The three-set union (what counts as "stale")

The diff-scope formula is stated once as the single source of truth in spec.md REQ-NS5-003 (this section quotes it; it does not re-derive it):

> `diff_scope = (git_diff_paths ∪ M1_changed_paths ∪ M2_owner_paths) ∩ graph_bound_paths`

**UNION semantics** — the three input sets are OR'd, not AND'd:

- **git_diff_paths** (baseline-to-HEAD `git diff --name-only`) — paths that changed in the committed tree since the baseline. A graph-bound path here seeds a subtree EVEN IF M1/M2 did not surface it (the change is committed but not yet detected/routed — the common case for a baseline far behind HEAD).
- **M1_changed_paths** (the detect `changed_path` set) — paths the real-time detect flagged. A graph-bound path here seeds a subtree EVEN IF git-diff did not catch it (e.g. an uncommitted in-session edit M1 caught but the engineer has not committed yet — the real-time-catch case the auditor's recommendation preserves).
- **M2_owner_paths** (the route `owner_path` set) — paths the route layer bound to a work-item. A graph-bound path here seeds a subtree (the owner-binding dimension supplements the other two).

The `∩ graph_bound_paths` filter is the ONE exclusion: a path that is NOT graph-bound (no M0 graph edge carries it as a `source_path`) does NOT seed a subtree regardless of which input set it is in — there is no doc row to fix. This is distinct from "git-diff but not in M1/M2" (which DOES seed under UNION); the exclusion is "not graph-bound" (which does NOT seed), not "not in M1/M2". The prior draft conflated these two and produced an intersection reading (`git-diff ∩ (M1 ∪ M2)`) that contradicted the UNION formula above; this is corrected to match REQ-NS5-003 verbatim.

### §C.4 The incremental contract (NOT full regen)

The diff-scope yields ONLY the stale subtrees. The AI draft rewrites ONLY those subtrees. The apply patches ONLY those subtrees. The 8 unchanged subtrees (in a fixture where 2 of 10 are stale) are NEVER touched. This is the structural difference between M3 (incremental) and the 001/003/M0 chains (regenerate-and-replace). AC-NS5-003 sub-assertion (a) encodes this: a fixture with 2 stale subtrees yields exactly 2 patched drafts, NOT 10.

## §D. The Draft Staging Surface Schema

### §D.1 Directory layout

```
.moai/project/navigator/fix-drafts/
  └── <draft-id>/                      # deterministic hash of sorted diff-scope + baseline SHA
      ├── request.json                  # layer 1 output (the handoff manifest)
      ├── draft/                        # layer 2 output (the AI-drafted subtrees)
      │   ├── draft.json                #   manifest: subtree → target live-doc mapping
      │   ├── capability-map.md.patch   #   unified-diff preview for approval gate
      │   ├── audit-report.json.patch   #   one .patch per stale subtree
      │   └── <subtree-files>           #   the patched subtree content
      └── applied.json                  # layer 3 output (written only AFTER approval)
```

### §D.2 Why a staging surface (not direct apply)

1. **Auditability**: the staging dir preserves the draft + the approval ledger. A rejected draft is recoverable (the engineer can re-inspect why they rejected it). An auto-applied draft leaves no trace.
2. **Atomicity**: the apply step atomic-renames ALL approved subtrees in one pass. A crash mid-apply leaves the staging dir intact (the apply is retryable).
3. **Preview source**: the `*.patch` files are the approval gate's preview content. Generating them at the staging surface (not in-memory) lets the `AskUserQuestion` preview field render them without the orchestrator holding the full draft in context.

### §D.3 Idempotence + provenance

- `<draft-id>` = SHA-256 of the sorted `diff_scope[]` + `baseline_commit_sha`. Deterministic — two runs on the same inputs produce the same id. NOT wall-clock.
- `request.json` `provenance`: `fix_commit_sha` + `baseline_commit_sha` + `captured_at` (git-committer-date). No `time.Now()`.
- `applied.json` `provenance`: approver + approval timestamp (git-committer-date) + applied subtree IDs + resulting live-doc SHA. No wall-clock.

This carries forward M0's REQ-NS-009 (atomic-rename + idempotence + provenance, no wall-clock) and M2's REQ-NS4-008. Two runs on the same HEAD + same baseline + same inputs produce byte-identical `request.json` (AC-NS5-004b).

## §E. Alternatives Considered + Rejected

### §E.1 (alt-A) Go-embedded LLM client

**Proposal**: add an OpenAI/Anthropic SDK to `internal/navigator/fix/` so the CLI produces the draft itself, no orchestrator round-trip.

**Rejected** (REQ-NS5-007 forbids). Reasons (§A.2): duplicates model-config, adds secret-management surface, bypasses Claude Code's permission/effort/thinking model. The CI grep guard (AC-NS5-007a) mechanically enforces zero LLM-client imports.

### §E.2 (alt-B) Full regenerate-and-replace on approval

**Proposal**: when the diff-scope is non-empty, regenerate the WHOLE target doc artifact (not just the stale subtrees), same as the 001/003/M0 chains.

**Rejected** (REQ-NS5-003/006 forbid). Reasons: discards human edits to untouched sections; expensive (full regen vs. subtree patch); non-incremental. M3 is the INCREMENTAL complement to the regenerate-and-replace chains, not a replacement. A future threshold-based fallback (> N subtrees stale → full regen) is a separate SPEC candidate, NOT an M3 deliverable (§F Out of Scope in spec.md).

### §E.3 (alt-C) Auto-apply without approval

**Proposal**: a `--yes` flag that skips the AskUserQuestion gate and applies the draft directly.

**Rejected** (REQ-NS5-008 forbids). Reasons: collapses M3's safety value proposition; an auto-applied wrong draft corrupts the doc map with confidence. The approval gate IS the feature that distinguishes M3 from regenerate-and-replace. A future `--yes` mode for trusted low-risk drafts is a follow-up SPEC that MUST defend why the safety boundary is droppable (§F Out of Scope).

### §E.4 (alt-D) PostToolUse real-time drafting

**Proposal**: register a PostToolUse hook that drafts a fix on each edit.

**Rejected** (REQ-NS5-001 forbids). Reasons (plan.md §C.1): a per-edit draft would (a) usually find no work-items (M2 is on-demand), and (b) flood the engineer with one draft per keystroke. M3's cadence is a deliberate, batched, human-gated operation. M1 owns the real-time advisory surface; M3 owns the coarser-cadence fix surface.

### §E.5 (alt-E) In-process function call (Go engine calls the orchestrator)

**Proposal**: the Go engine calls an orchestrator function directly (in-process) to spawn the manager-develop delegation, no stdout handoff.

**Rejected**. Reason: the Go engine and the orchestrator run in different process contexts (the `moai` CLI binary vs. the Claude Code session). An in-process call is not possible across that boundary. The file-based handoff (`request.json` + stdout JSON signal) is the only contract that crosses the boundary without coupling the Go engine to Claude Code's runtime internals.

## §F. Cross-References

- `.moai/reports/navigator-redesign-bas-20260805.html` §3.3(C) line 274, §4 line 287-308, §5 Q2 line 327, §5 Q3 line 333-341, §6 line 371-375, §8 line 430-433, §9 line 443 — the design-report authority.
- `research.md` — the CodeWiki `--compare-to` pattern investigation + the mapping onto moai-adk's architecture.
- `spec.md` §C REQ-NS5-007 (the split-architecture REQ) + REQ-NS5-008 (the approval-gate REQ) — the canonical requirements this design defends.
- `plan.md` §B (the Tier-L architecture decision summary) + §C (the technical approach) — the implementation-level view.
- `.claude/rules/moai/core/askuser-protocol.md` § Channel Monopoly + § Preview Field Standards + § Option Description Standards — the approval-gate mechanics.
- `.claude/rules/moai/development/manager-develop-prompt-template.md` § Section A-E — the AI-draft delegation prompt template (the orchestrator uses this to spawn manager-develop for layer 2).
- CLAUDE.md §1 (Core Identity — the orchestrator delegates writes) + §2 (Request Processing Pipeline) — the architectural invariant that grounds Reason 2 (§A.2).

# acceptance.md — SPEC-NAVIGATOR-SYNC-005 (BAS Epic M3 — Falconer Fix)

> Acceptance criteria for the M3 Fix layer. Each AC is a binary-testable Given-When-Then,
> traceable to a spec.md §C REQ. The automation-rate AC (AC-NS5-010) names the exact command + ratio
> formula (verification-claim-integrity §1.1 surface 3 + §2).

## §A. Severity Legend

- **MUST** — failure blocks merge (M3 cannot ship without this passing).
- **SHOULD** — failure is a debt item, mergeable with a recorded `[NEEDS CLARIFICATION]` or follow-up.
- **MAY** — quality bar, not a gate.

## §B. Traceability Matrix

| REQ | AC(s) | Severity | Milestone |
|---|---|---|---|
| REQ-NS5-001 (on-demand CLI, no PostToolUse) | AC-NS5-001a, AC-NS5-001b | MUST | M3.2 |
| REQ-NS5-002 (four read-only inputs) | AC-NS5-002 (sub-assertions a/b/c/d) | MUST | M3.6 |
| REQ-NS5-003 (diff-scope, stale subtrees) | AC-NS5-003 (sub-assertions a/b/c) | MUST | M3.1 |
| REQ-NS5-004 (draft-request artifact) | AC-NS5-004a, AC-NS5-004b | MUST | M3.1 |
| REQ-NS5-005 (consumer-only) | AC-NS5-005a, AC-NS5-005b | MUST | M3.5 |
| REQ-NS5-006 (non-overlap) | AC-NS5-006 | MUST | M3.5 |
| REQ-NS5-007 (AI draft via orchestrator, no Go LLM) | AC-NS5-007a, AC-NS5-007b | MUST | M3.3 |
| REQ-NS5-008 (draft-not-auto-apply, approval gate + token) | AC-NS5-008a, AC-NS5-008b, AC-NS5-008c, AC-NS5-008d | MUST | M3.4/M3.5 |
| REQ-NS5-009 (fail-open on every error mode) | AC-NS5-009 (table-driven: 8 modes) | MUST | M3.6 |
| REQ-NS5-010 (≥50% automation rate, measured) | AC-NS5-010 | MUST | M3.6 |
| REQ-NS5-011 (Hidden subcommand) | AC-NS5-011 | MUST | M3.2 |
| REQ-NS5-012 (template-first) | AC-NS5-012 | MUST | M3.6 |
| REQ-NS5-013 (draft scope-conformance validation) | AC-NS5-013 | MUST | M3.4 |

## §C. Definition of Done

- Every MUST AC in §D shows PASS with a cited command + observed output (no narrative "passes").
- `go test ./internal/navigator/fix/ -count=1` exits 0.
- `go test ./internal/navigator/fix/ -run TestFixAutomationRate -v` prints an observed automation rate ≥ 50.0.
- `go test ./internal/navigator/fix/ -run TestFixNonOverlap -count=1` exits 0.
- `go test ./internal/navigator/fix/ -run TestFixNoLLMClient -count=1` exits 0 (the split-architecture grep guard — zero LLM-client imports).
- `go test ./internal/cli/ -run TestNavigatorFix -count=1` exits 0 (Hidden-subcommand assertion).
- `moai spec lint --strict .moai/specs/SPEC-NAVIGATOR-SYNC-005/` reports 0 errors (the repo-wide baseline of pre-existing findings is NOT attributed to this SPEC).
- An idempotence test (two consecutive `navigator-fix` runs on the same inputs + same baseline) produces byte-identical `request.json`.
- An incremental test (a fixture where only 2 of 10 subtrees are stale) yields exactly 2 patched drafts, NOT 10.
- No live doc surface (`capability-map.md` / `audit-report.{md,json}` / `capability-symbols.{md,json}` / `nav-graph.json`) is mutated by the `navigator-fix` CLI run alone — mutation occurs ONLY after the orchestrator's apply-on-approval path fires (AC-NS5-008a).
- `go test ./internal/navigator/fix/ -run TestApplyTokenRefusal -count=1` exits 0 (the approval_token refusal — `--apply` without a valid token refuses + no live-doc mutation; AC-NS5-008d).
- `go test ./internal/navigator/fix/ -run TestDraftScopeConformance -count=1` exits 0 (the scope-conformance validation — an out-of-scope draft subtree is excluded from gate + apply; AC-NS5-013).
- `internal/navigator/{sync,detect,route,tiers}/`, `internal/hook/navigator_detect*.go`, `internal/mx/`, and the three predecessor chain scripts are byte-unchanged by the M3 run-phase diff (AC-NS5-005a).

## §D. Acceptance Criteria (Given-When-Then)

### AC-NS5-001a — on-demand CLI produces draft-request (REQ-NS5-001, MUST, M3.2)

**Given** a fixture corpus at `internal/navigator/fix/testdata/fix-corpus/` containing a non-empty `work-items.json` (M2) + a detect-state directory with ≥1 JSONL row (M1) + a `nav-graph.json` (M0) + fixture live doc surfaces + a fixture baseline commit,
**When** the `navigator-fix` subcommand is invoked with the corpus as the project root,
**Then** the Fix layer produces `.moai/project/navigator/fix-drafts/<draft-id>/request.json` with a non-empty `diff_scope[]` array, the process exits 0, the stdout emits a `{"draft_request_path": "...", "status": "ready"}` JSON line (the orchestrator handoff signal), and no user-facing error is surfaced.

### AC-NS5-001b — no PostToolUse real-time path (REQ-NS5-001, MUST, M3.2)

**Given** the M3 run-phase diff,
**When** the integration checks run,
**Then** BOTH sub-assertions hold:

- **(a) no PostToolUse hook branch**: `grep -rn 'navigator-fix\|navigator_fix\|runNavigatorFix' internal/hook/` returns zero matches — the Fix layer registers NO PostToolUse branch, NO `handle-navigator-fix.sh` wrapper exists (`ls .claude/hooks/moai/handle-navigator-fix.sh 2>/dev/null` shows no file), and the only entry point is the `navigator-fix` Hidden cobra subcommand.
- **(b) no forked hook chain**: `grep -rn 'navigator-fix' internal/cli/` shows the subcommand registration is a Hidden cobra command (mirroring `navigator-route`/`navigator-tiers`), NOT a hook handler.

### AC-NS5-002 — four read-only inputs consumed (REQ-NS5-002, MUST, M3.6)

**Given** a fixture corpus with all four inputs present (work-items.json + detect JSONL + nav-graph.json + live doc surfaces),
**When** the Fix layer runs,
**Then** ALL FOUR sub-assertions hold:

- **(a) M2 work-items consumed**: the `request.json` `work_item_refs[]` array contains entries matching the fixture's `work_items[]` (`source_kind` + `owner_path` + `action`), and the `draft_instructions.per_subtree[]` strategies are derived from the `action` field (e.g. an `action` of "create a SPEC" yields a "draft SPEC stub" strategy).
- **(b) M1 detect consumed across sessions**: the `diff_scope[]` includes subtrees seeded by `changed_path` values from ALL `*.jsonl` files in the detect-state fixture directory (deduplicated — a path edited in two session files seeds one subtree, latest `changed_at` wins).
- **(c) M0 nav-graph consumed for subtree identification**: at least one `diff_scope[]` entry whose stale path is a graph edge's `source_path` resolves to the bound doc subtree (node + owning section) via graph traversal, NOT just a file-path match.
- **(d) live doc surfaces read (not written)**: the `request.json` `draft_instructions` reference the live doc surface paths as READ targets; the Fix layer's only writes are the staging surface (`request.json`) + the log — verified by `grep -rn 'os.WriteFile\|os.Rename' internal/navigator/fix/*.go` showing writes ONLY under `.moai/project/navigator/fix-drafts/` and `.moai/logs/`.

### AC-NS5-003 — diff-scope identifies stale subtrees, incremental not full-regen (REQ-NS5-003, MUST, M3.1)

**Given** a fixture corpus where the baseline-to-HEAD `git diff --name-only` touches 2 of 10 bound doc subtrees (the other 8 are unchanged since the baseline),
**When** the Fix layer computes the diff-scope,
**Then** ALL THREE sub-assertions hold:

- **(a) only stale subtrees identified**: the `diff_scope[]` array contains exactly 2 entries (the 2 stale subtrees), NOT 10. The 8 unchanged subtrees do NOT appear.
- **(b) baseline resolution priority**: when no `--compare-to` flag is passed, the baseline is the `nav-graph.json` `provenance.extract_commit_sha` (the default); when the flag IS passed, the flag value wins; when `nav-graph.json` is absent, the baseline degrades to `HEAD~1` (logged).
- **(c) UNION semantics (the SSOT formula from REQ-NS5-003)**: `diff_scope = (git_diff_paths ∪ M1_changed_paths ∪ M2_owner_paths) ∩ graph_bound_paths`. The three input sets are OR'd (UNION), then intersected with graph-bound paths. A graph-bound path in git-diff alone (not in M1/M2) DOES seed a subtree; a graph-bound path in M1 alone (not in git-diff) DOES seed; a path NOT graph-bound does NOT seed regardless of which set it is in. The test fixture MUST include one case per input set contributing independently (a git-diff-only graph-bound path, an M1-only graph-bound path, an M2-only graph-bound path) and assert each seeds exactly one subtree — validating UNION, not intersection.

### AC-NS5-004a — draft-request schema + provenance, no wall-clock (REQ-NS5-004, MUST, M3.1)

**Given** a successful Fix layer run (non-empty diff-scope),
**When** `request.json` is inspected,
**Then** the artifact carries: a `provenance` block with `fix_commit_sha` (matches `git rev-parse HEAD`) + `baseline_commit_sha` (the resolved baseline) + `captured_at` (matches `git log -1 --format=%cI` of `fix_commit_sha`); a `diff_scope[]` array; a `work_item_refs[]` array; a `draft_instructions` block. There is NO `time.Now()`-derived field anywhere. `grep -E '"(generated_at|created_at|timestamp)":' request.json` returns zero matches (the canonical field is `captured_at`, sourced from git).

### AC-NS5-004b — idempotence: byte-identical re-run (REQ-NS5-004, MUST, M3.1)

**Given** the same input set (unchanged work-items + detect state + nav-graph) at the same HEAD + same baseline,
**When** `navigator-fix` is invoked twice in succession,
**Then** the two `request.json` outputs are byte-identical (`cmp -s <run1> <run2>` exits 0), AND the `<draft-id>` (deterministic hash) is the same across both runs. (The atomic-rename is a no-op write when content is unchanged.)

### AC-NS5-005a — consumer-only: M0/M1/M2/M4 + mx byte-unchanged (REQ-NS5-005, MUST, M3.5)

**Given** the M3 run-phase diff is applied,
**When** `git diff --name-only origin/main...HEAD` is run,
**Then** the output does NOT contain any path under `internal/navigator/sync/`, `internal/navigator/detect/`, `internal/navigator/route/`, `internal/navigator/tiers/`, `internal/hook/navigator_detect*.go`, or `internal/mx/`. Command: `git diff --name-only origin/main...HEAD | grep -E '^internal/(navigator/(sync|detect|route|tiers)|hook/navigator_detect|mx)/' ; [ $? -eq 1 ]` (grep exit 1 = no matches = PASS).

### AC-NS5-005b — consumer-only: read via os.ReadFile, write only to staging (REQ-NS5-005, MUST, M3.5)

**Given** the M3 source files,
**When** `grep -rn 'os.WriteFile\|os.Rename' internal/navigator/fix/*.go` is run (excluding `apply.go` which is post-approval),
**Then** no write/rename call targets `internal/navigator/{sync,detect,route,tiers}/`, `internal/hook/`, `internal/mx/`, or any live doc surface path. The only writes are under `.moai/project/navigator/fix-drafts/` and `.moai/logs/`. AND the Fix layer reads work-items + detect + nav-graph + live docs via `os.ReadFile` (not by executing any predecessor chain script).

### AC-NS5-006 — non-overlap with predecessors + M0/M1/M2/M4 (REQ-NS5-006, MUST, M3.5)

**Given** the M3 production source files (`internal/navigator/fix/*.go` excluding `*_test.go`, excluding `apply.go`'s post-approval writes) + `internal/cli/navigator_fix.go`,
**When** `grep -rn 'capability-map\.md\|audit-report\.\(md\|json\)\|capability-symbols\.\(md\|json\)\|nav-graph\.json\|tiers\.json\|work-items\.\(md\|json\)\|blueprint/\|decisions/' internal/navigator/fix/ internal/cli/navigator_fix.go` is run,
**Then** the ONLY matches for `work-items.json`, `nav-graph.json`, the detect JSONL path, `capability-map.md`, `audit-report.{md,json}`, and `capability-symbols.{md,json}` are in a READ context (`os.ReadFile`, `os.Open`). There are ZERO matches for `tiers.json`, `blueprint/`, `decisions/`, or `work-items.md` as write targets. A new `internal/navigator/fix/nonoverlap_test.go` encodes this assertion and carries forward the pattern from M0's + M1's + M2's nonoverlap tests. (The `apply.go` post-approval writes to live doc surfaces are governed by AC-NS5-008c, NOT this AC — `apply.go` is excluded from this grep because its writes ARE the approved-subtree patches, which is the intended exception.)

### AC-NS5-007a — split-architecture: zero LLM-client imports in Go engine (REQ-NS5-007, MUST, M3.3)

**Given** the M3 Go engine source files (`internal/navigator/fix/*.go` excluding `*_test.go`),
**When** `grep -rn 'openai\|anthropic\|langchain\|mcp__\|claude\.ai\|api\.openai\|generativeai' internal/navigator/fix/*.go` is run,
**Then** the output is empty — the Go engine contains ZERO LLM-client imports. The AI draft is produced by the orchestrator-spawned `manager-develop` delegation, NOT by the Go engine. A CI guard test `internal/navigator/fix/llm_client_guard_test.go` (mirroring the C-HRA-008 subagent-boundary pattern) encodes this grep assertion.

### AC-NS5-007b — split-architecture: stdout handoff signal (REQ-NS5-007, MUST, M3.3)

**Given** a successful `navigator-fix` CLI run producing a non-empty `request.json`,
**When** the CLI stdout is inspected,
**Then** it emits exactly one JSON line `{"draft_request_path": "<path>", "status": "ready", "draft_id": "<id>"}` — the handoff signal the orchestrator consumes to spawn the `manager-develop` AI-draft delegation. The CLI does NOT produce the draft itself (no draft files exist at `fix-drafts/<id>/draft/` after the CLI run — those are produced by the orchestrator's subsequent delegation).

### AC-NS5-008a — draft-not-auto-apply: no live-doc mutation without approval (REQ-NS5-008, MUST, M3.5)

**Given** a fixture live doc surface (e.g. `capability-map.md`) with a known SHA before the `navigator-fix` CLI run,
**When** `navigator-fix` runs and produces a `request.json` (but the orchestrator has NOT yet run the AI-draft delegation + approval gate),
**Then** the live doc surface SHA is UNCHANGED after the CLI run (`git diff --stat <live-doc-path>` shows no change). The only artifacts produced are under `fix-drafts/<id>/` (the staging surface). Mutation of the live doc occurs ONLY after the orchestrator's apply-on-approval path fires (AC-NS5-008c).

### AC-NS5-008b — approval surface: AskUserQuestion 4-option gate + preview (REQ-NS5-008, MUST, M3.4)

**Given** the orchestrator has received the AI-drafted subtrees at `fix-drafts/<id>/draft/` (layer 2 complete),
**When** the orchestrator presents the approval gate,
**Then** the gate is an `AskUserQuestion` invocation (preloaded via `ToolSearch(query: "select:AskUserQuestion")` per askuser-protocol.md) offering exactly 4 options: (a) approve + apply (권장), (b) approve selected, (c) edit then apply, (d) reject — each with a `description` field stating the immediate result + irreversibility. AND the `preview` field renders the `*.patch` unified-diff previews so the engineer sees exactly what changes before deciding. The gate is NOT free-form prose (per askuser-protocol.md § Channel Monopoly). (This AC is verified at the orchestrator-integration level — the `internal/navigator/fix/` Go code produces the `*.patch` previews; the orchestrator renders the `AskUserQuestion` gate consuming them.)

### AC-NS5-008c — apply-on-approval: atomic-rename + applied.json ledger (REQ-NS5-008, MUST, M3.5)

**Given** the orchestrator received an approval decision (option a — approve + apply),
**When** the apply step runs (`internal/navigator/fix/apply.go` or `moai navigator-fix --apply <draft-id>`),
**Then** the approved subtrees are atomic-renamed (`.tmp` + `os.Rename`) to their target live doc surfaces, an `applied.json` ledger is written at `fix-drafts/<id>/applied.json` recording the approver + approval timestamp (git-committer-date, NOT wall-clock) + applied subtree IDs + resulting live-doc SHA, and ONLY the approved subtrees are touched (unapproved subtrees in the same live doc are NOT modified — the incremental contract).

### AC-NS5-008d — approval_token: --apply without token refuses (REQ-NS5-008 sub-clause c4, MUST, M3.5)

**Given** a draft-id with NO `approval.json` token at its staging dir (no prior gate approval — simulating a bare-shell invocation),
**When** `moai navigator-fix --apply <draft-id>` is invoked,
**Then** the CLI refuses: exits non-zero, emits a message naming the missing/invalid `approval.json` token, and does NOT mutate any live doc surface (`git diff --stat <live-doc-paths>` shows no change). AND given the SAME draft-id WITH a valid `approval.json` token (the gate wrote it on approval — draft-id + approval option + request.json provenance match), the same invocation proceeds to apply normally (AC-NS5-008c fires). The test: `go test ./internal/navigator/fix/ -run TestApplyTokenRefusal -v` asserts BOTH branches (no-token → refuse; valid-token → apply).

### AC-NS5-013 — draft scope-conformance: out-of-scope subtree excluded (REQ-NS5-013, MUST, M3.4)

**Given** a fixture draft at `fix-drafts/<id>/draft/` containing 3 patched subtrees — 2 whose IDs ARE in the `request.json` `diff_scope[]` (in-scope) + 1 whose ID is NOT in `diff_scope[]` (out-of-scope, simulating manager-develop over-production),
**When** the orchestrator prepares the approval gate (and/or `apply.go` runs on an approval),
**Then** the gate preview + the apply step process ONLY the 2 in-scope subtrees; the 1 out-of-scope subtree is excluded from BOTH the gate preview AND the apply. A warning naming the excluded subtree ID + the `diff_scope[]` it is not in is logged to `.moai/logs/navigator-sync.log`. The live doc surface is NOT mutated by the out-of-scope subtree. The test: `go test ./internal/navigator/fix/ -run TestDraftScopeConformance -v` asserts the exclusion (2 applied, 1 excluded + warned).

### AC-NS5-009 — fail-open across 8 error modes (REQ-NS5-009, MUST, M3.6)

**Given** any of the 8 failure modes in the table below,
**When** `navigator-fix` is invoked,
**Then** for every row the Fix layer produces the expected exit-0 behavior and the expected empty-or-partial output. In ALL 8 cases: exit 0, no user-facing error, no cascade into sibling `/moai project` steps.

| case | trigger | expected exit-0 | expected empty-or-partial output |
|---|---|---|---|
| 009a work-items absent | `work-items.json` does NOT exist (M2 not run) | exit 0, degrade (diff-scope from M1 detect + git-diff only, no `action` hints) | `request.json` with degraded `work_item_refs: []` + a log line |
| 009b detect absent/empty | `.moai/state/navigator-detect/` absent or no `*.jsonl` | exit 0, degrade (diff-scope from M2 owner_paths + git-diff only) | `request.json` with diff-scope seeded by M2 only; one log line |
| 009c nav-graph absent | `nav-graph.json` does NOT exist | exit 0, degrade (subtree identification falls back to file-path → doc-section heuristic, baseline falls back to `HEAD~1`) | `request.json` with degraded subtree resolution; one log line |
| 009d baseline unresolvable | no `--compare-to`, no `nav-graph.json` provenance, AND `HEAD~1` fails (shallow clone) | exit 0, write NO staging artifact | no `request.json` written; one summary log line "baseline unresolvable, skipping" |
| 009e unparseable JSON | one input (work-items OR nav-graph OR a detect JSONL line) is invalid JSON | exit 0, skip the malformed input/line, continue with well-formed inputs | partial `request.json` (well-formed inputs only); one log line per malformed input |
| 009f schema-invalid | `work-items.json` valid JSON but missing `work_items[]` (or `nav-graph.json` missing `nodes[]`) | exit 0, degrade the affected source | partial `request.json`; one log line |
| 009g empty diff-scope | baseline-to-HEAD diff + M1/M2 sets yield ZERO stale subtrees (doc map already consistent) | exit 0, write `request.json` with `diff_scope: []` + stdout message "0 stale subtrees, doc map consistent" | `request.json` with empty `diff_scope`; one log line (this is the SUCCESS case, not an error) |
| 009h no-LLM-runtime | `navigator-fix` ran outside a Claude Code session (bare shell, no `Agent()` runtime) | exit 0, produce `request.json` (layer 1 complete), stdout message "draft-request produced; run inside /moai project to generate the AI draft" | `request.json` produced; NO draft files (layer 2 cannot fire without orchestrator); one log line |

### AC-NS5-010 — ≥50% Fix automation rate, mechanically measured (REQ-NS5-010, MUST, M3.6)

**Given** the fixture corpus at `internal/navigator/fix/testdata/fix-corpus/` (10 stale-subtree scenarios: 3 audit-missing × 3 audit-orphan × 4 detect, spanning the 3 `action` strategies — regenerate row / re-link symbol / draft SPEC stub; each scenario carries a fixture work-items.json + detect JSONL + nav-graph.json + live docs + baseline commit) and a simulated approval loop recording per-scenario whether the draft was approved unmodified (option a) or needed edit/selection (option b/c) or was rejected (option d),

**Dual-arithmetic (the floor survives the worst case)**:
- Happy path: 6 of 10 scenarios approved unmodified → automation rate = 6/10 = **60.0%** ≥ 50% ✓.
- Worst case (4 scenarios need edit): automation rate = 6/10 = **60.0%** ≥ 50% ✓ (the 4 edit-cases are in the denominator, not the numerator).

**When** the command `go test ./internal/navigator/fix/ -run TestFixAutomationRate -v` is run,
**Then** the test emits an observed automation rate (the ratio `(drafts approved unmodified) / (total drafts produced)`, where "unmodified" = option (a) WITHOUT a prior edit) and the observed rate is `>= 50.0`. The test FAILS (non-zero exit) if the observed rate is `< 50.0` and prints the observed value on failure. The percentage printed on the PASS line is the Evidence; the ≥50% assertion is the Claim.

**Attribution (verification-claim-integrity §2)**: the automation rate is attributable to the exact `go test … -v` invocation + its stdout; it is NOT a carried-over estimate, NOT a sampled figure, NOT a narrative "high automation". A subsequent run on the same fixture corpus produces the same rate (deterministic).

### AC-NS5-011 — Hidden cobra subcommand (REQ-NS5-011, MUST, M3.2)

**Given** the M3 run-phase diff,
**When** the CLI integration checks run,
**Then** BOTH sub-assertions hold:

- **(a) navigator-fix is Hidden**: `go test ./internal/cli/ -run TestNavigatorFixHidden -v` asserts the `navigator-fix` cobra command has `Hidden: true`, mirroring `internal/cli/navigator_tiers_test.go:65-70`.
- **(b) no top-level moai subcommand**: `grep -rn 'navigator-fix' internal/cli/root.go` returns zero matches — `navigator-fix` is registered as a subcommand under the project step sequence, NOT as a top-level `moai navigator fix` user-facing command.

### AC-NS5-012 — template-first (REQ-NS5-012, MUST, M3.6)

**Given** the M3 run-phase diff touches any template-managed file under `.claude/`,
**When** `git diff --name-only origin/main...HEAD | grep '^internal/template/templates/'` is run,
**Then** every template-managed path modified locally ALSO appears in `internal/template/templates/`. AND if any template file changed, `make build` regenerated `catalog.yaml` and the regenerated catalog is committed in the same PR. (If M3 ships NO distributed surface — the expected case, pure CLI + runtime Go + orchestrator-mediated AI draft — this AC reduces to: no template path in the diff, no catalog regen required — document this in the PR body.)

## §E. Indirect Verification (cross-cutting)

- **LSP / lint gate**: `go vet ./internal/navigator/fix/... ./internal/cli/...` and `golangci-lint run ./internal/navigator/fix/... ./internal/cli/...` exit 0.
- **Race detector**: `go test -race ./internal/navigator/fix/` exits 0 (the apply-on-approval atomic-rename + the multi-session JSONL read are the race-relevant surfaces).
- **Subagent boundary** (C-HRA-008 family): `grep -rn 'AskUserQuestion\|mcp__askuser' internal/navigator/fix/*.go internal/cli/navigator_fix.go` returns zero matches — the Fix layer runs in CLI/subagent context and never prompts (the approval gate is orchestrator-side, NOT in the Go engine).

## §F. Edge Cases (quality bar, not merge gates)

- **Empty work-items (M2 found no drift)**: `work-items.json` with `work_items: []` → the diff-scope falls back to M1 detect + git-diff only; `work_item_refs: []` in `request.json`. Not an error.
- **Baseline far behind HEAD (large diff)**: the diff-scope is large → many stale subtrees → many drafts. The engineer may approve selected (option b) to apply incrementally. Not an error; the on-demand cadence lets the engineer batch.
- **Draft touches a subtree NOT in diff-scope**: the manager-develop delegation is constrained by the spawn prompt ("produce ONLY the patched subtrees named in `diff_scope[]`"). If it violates this, the non-overlap AC (AC-NS5-006) catches the extra write. The orchestrator's trust-but-verify checks the draft against the `diff_scope[]` before presenting the approval gate.
- **Same subtree stale via both audit and detect**: the diff-scope deduplicates by `(doc_surface, subtree_id)`; the `work_item_ref` carries both source_kinds. One draft, one approval, one apply.
- **Approval gate rejected**: the draft is discarded; the staging dir `fix-drafts/<id>/` remains (for audit) but NO apply fires. The live doc is untouched. A rejected draft is NOT an error.
- **Very large diff-scope (> 50 subtrees)**: M3 does NOT fall back to full regen (out of scope, §F in spec.md). The engineer sees 50+ drafts in the approval gate and may approve selected (option b). A future threshold-based full-regen fallback is a separate SPEC.

## §G. Forward-Looking Checks (advisory, not gates)

- **M5 readiness**: M3's incremental regen handles EXISTING doc subtrees that drifted. M5 (brownfield reverse-extraction) handles code that has NO doc yet. M3 + M5 together cover both directions. M3's `request.json` schema + the apply-on-approval pattern are reusable for M5 (M5 would produce `request.json` entries with `stale_reason: "no-existing-doc"` rather than `"drifted"`).
- **PostToolUse real-time path future**: if a debounce/cadence policy is later desired (e.g. "draft only on session end", or "draft only when ≥N drifts accumulate"), a follow-up SPEC may add a SessionEnd hook that calls the same `fix.ComputeScope` engine. The engine is pure (layer 1) so the real-time path would be a thin wrapper. Explicitly deferred — the on-demand CLI is the M3 deliverable.
- **Auto-apply mode future**: a future `--yes` flag that skips the AskUserQuestion gate for trusted low-risk drafts (e.g. symbol re-link where the rename is mechanical). Explicitly out of scope for M3 (§F in spec.md); a follow-up SPEC MUST defend why the safety boundary is droppable.
- **External-tracker integration future**: a future SPEC MAY consume `applied.json` read-only to close corresponding M2 work-items (mark them resolved). The `applied.json` ledger is designed to be tracker-friendly (subtree IDs → work-item refs).

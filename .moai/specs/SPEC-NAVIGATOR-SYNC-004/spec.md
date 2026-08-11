---
id: SPEC-NAVIGATOR-SYNC-004
title: "Navigator Sync (BAS M2) — Falconer Route: audit advisory → actionable work item promotion with code-path-bound owner"
version: "0.1.0"
status: completed
created: 2026-08-11
updated: 2026-08-12
author: manager-spec
priority: P1
phase: "v3.3 target"
module: navigator-sync
lifecycle: spec-anchored
tier: M
era: V3R6
tags: "navigator, sync, falconer, route, work-item, owner-binding, bas-epic"
related_specs: [SPEC-NAVIGATOR-SYNC-001, SPEC-NAVIGATOR-SYNC-002, SPEC-NAVIGATOR-SYNC-003, SPEC-PROJECT-NAVIGATOR-001, SPEC-PROJECT-NAVIGATOR-002, SPEC-PROJECT-NAVIGATOR-003]
depends_on: [SPEC-NAVIGATOR-SYNC-001]
---

# SPEC-NAVIGATOR-SYNC-004 — Navigator Sync (BAS Epic M2) — Falconer Route

## HISTORY

- 2026-08-11 (initial draft) — BAS (Blueprint-Anchored Synchronization) Epic M2, plan-phase. M0 (`SPEC-NAVIGATOR-SYNC-001`, PR #1375 squash `f4bd58acbd`, MERGED 2026-08-05, `status: completed`) delivered the SSOT binding-token trio (`@NAV:DEC-<id>` / `@NAV:SYM:<symbol>` / `@MX:SPEC:<id>` bridged not absorbed per REQ-NS-005) and the graph-join schema layer at `internal/navigator/sync/` producing `.moai/project/navigator/nav-graph.json` (nodes: decision/spec/symbol; edges: dec-edge/spec-edge/sym-edge, each carrying `source_path` + `line_number`). M1 (`SPEC-NAVIGATOR-SYNC-002`, PR #1379 squash `304907b6d`, MERGED 2026-08-06, `status: completed`) delivered the Falconer **Detect** element (PostToolUse changed-path → affected-graph-rows mapping at `internal/navigator/detect/` + `internal/hook/navigator_detect*.go`, writing `.moai/state/navigator-detect/<session-id>.jsonl`). M4 (`SPEC-NAVIGATOR-SYNC-003`, PR #1384 run `23abbd206` + #1385 sync `7e9648650`, MERGED 2026-08-06/07, `status: completed`) delivered the 4-tier addressable map (`tiers.json` overlay + blueprint/ADR/symbol enrichments at `internal/navigator/tiers/`). M2 is the **Route** element — the second of the Falconer 3-element loop (detect → route → fix) from the design report `.moai/reports/navigator-redesign-bas-20260805.html` §3.3(C) line 268-284, §4 asset-reuse table line 287-308, §6 M2 milestone card line 365-369, §7 slice table line 398, §9 success metrics line 442 (Route accuracy ≥ 70%). M2 promotes audit missing/orphan findings + M1 detect findings into **actionable work items**, with the **owner bound to a code path** (not a volunteer) — falconer's living-documentation principle: *"tell the engineer who touched the code which doc sections are now their responsibility"* (design report §3.3 line 274, §4 line 300). REQ-NS2-003 (M1) explicitly states *"The Detect layer SHALL NOT promote findings to an actionable work item — Route (M2) owns work-item promotion"*; this SPEC occupies that reserved seat. Bridge-not-absorb principle (M0 REQ-NS-005, M1 REQ-NS2-005) carries forward as REQ-NS4-005: M2 consumes audit-report.json + M1 detect state + M0 nav-graph.json READ-ONLY, never mutates producers.

## §A. User Story

**As a** developer or agent staring at a drifted Navigator graph — a SPEC with no matching design feature, a design feature with no SPEC, or a code edit that just touched bound graph rows —
**I want** each drift finding promoted into a concrete work item that names exactly which code path or design doc is responsible for closing the drift,
**so that** I know who needs to act (bound to code, not to a person who may have moved on) and what action closes the gap — without re-reading the whole audit or replaying every detection.

Today the 002 audit surfaces `missing`/`orphan` findings as advisory text, and M1 Detect surfaces per-edit affected rows as an ephemeral `systemMessage` + a JSONL line. Neither is **actionable**: neither names an owner, neither persists as a work item, neither tells the engineer "this code path is now your doc responsibility." M2 closes that gap — the **route** element of the 3-element living-document loop. M2 alone does not fix drift (M3 owns Fix); it makes drift **addressable** by binding each finding to the code path that owns the follow-up.

**Falconer framing**: "담당자를 자원봉사자가 아니라 코드 경로에 결합한다" (design report §6 M2 line 368). Owner = code path, never a person. A person moves on; a code path stays.

## §B. Scope

### In scope

- An on-demand Route layer that reads `.moai/project/navigator/audit-report.json` (002) + `.moai/state/navigator-detect/*.jsonl` (M1) + `.moai/project/navigator/nav-graph.json` (M0) and promotes findings into actionable work items.
- **Owner = code path binding**: each work item's owner resolves to a code path or a design-doc path — three resolution paths (orphan → `implementation_path`; missing → `@NAV:SYM` symbol's owning package via the M0 graph, else the design-doc `source.file`; detect → `changed_path`).
- Independent output artifact: `.moai/project/navigator/work-items.{md,json}` — human-readable `.md` + machine-readable `.json`, atomic-rename, fail-open, idempotent, provenance-attributed per `verification-claim-integrity.md` §2.
- Trigger surface: an on-demand Hidden cobra subcommand `navigator-route` (sibling of `navigator-sync` / `navigator-enrich` / `navigator-tiers`), invoked from the existing `/moai project` surface. **No PostToolUse real-time path in M2** — the decision is defended in plan.md §C.1.
- Fail-open semantics: every error mode (audit absent / detect state absent / nav-graph absent / parse error) degrades to an empty work-item set + exit 0 + one log line.
- Reuse of M0's `atomicWrite` pattern + `Provenance` model + the audit-report.json schema fields (carrying forward 002's `implementation_path` for orphan owner resolution).
- Template-first distribution: any new distributed config keys ship in `internal/template/templates/` first.
- ≥70% Route accuracy success metric, mechanically measurable (design report §9 line 442).

### Excluded from M2 (full list in §F)

M3 Fix (AI-drafted incremental regen), PostToolUse real-time work-item promotion, modifying 002's audit chain / M0's sync layer / M1's detect layer / M4's tiers layer / the three predecessor Navigator chains, assigning owners to persons, creating GitHub issues / Linear tickets / SPEC drafts (M2 emits a local artifact only — no external-tracker writes). The canonical exclusions live in §F with `### Out of Scope — <topic>` sub-headings.

## §C. Functional Requirements (GEARS)

### §C.1 Trigger + input consumption

#### REQ-NS4-001 (Event-driven — on-demand CLI invocation)

**When** the `navigator-route` Hidden cobra subcommand is invoked (sibling of `navigator-sync` / `navigator-enrich` / `navigator-tiers`, wired into the existing `/moai project` surface the same way M0's join step is wired), the Route layer SHALL consume the latest available inputs and produce `.moai/project/navigator/work-items.{md,json}`. The Route layer SHALL NOT register a PostToolUse hook branch, SHALL NOT create a `handle-navigator-route.sh` wrapper, and SHALL NOT spawn a parallel real-time path — M2 is on-demand only (decision defended in plan.md §C.1).

#### REQ-NS4-002 (Ubiquitous — three read-only inputs, consumed latest-available)

The Route layer SHALL consume exactly three read-only inputs, each tolerated as absent via REQ-NS4-009 fail-open:
- **(a) audit-report.json** (002, `.moai/project/navigator/audit-report.json`) — the `missing[]`, `orphan[]`, and `matched[]` arrays; each `orphan` entry carries `implementation_path` (the code-path owner anchor); each `missing` entry carries `source.file` + `source.heading_path` (the design-doc owner anchor).
- **(b) M1 detect impact record** (`.moai/state/navigator-detect/<session-id>.jsonl`) — the accumulated per-detection rows `{changed_path, changed_at, affected_nodes[], affected_edges[]}` across all sessions present on disk. M2 reads ALL `*.jsonl` files in the detect state directory (not just one session), deduplicating by `changed_path` (the latest `changed_at` wins).
- **(c) M0 nav-graph.json** (`.moai/project/navigator/nav-graph.json`) — the graph context for owner resolution: when a `missing` entry's design doc references an `@NAV:SYM:<symbol>` token, M2 resolves the symbol's owning package via the graph's symbol nodes + their `source_path` edges. M0 is also the fallback context for cross-referencing a `changed_path` to its affected SPEC / decision nodes.

### §C.2 Work-item promotion + owner binding (the core transform)

#### REQ-NS4-003 (Ubiquitous — work-item promotion)

The Route layer SHALL promote each consumed finding into a work item. A work item SHALL carry five fields: `source_kind` ∈ `{audit-missing, audit-orphan, detect}`, `source_entry` (the original audit/detect entry verbatim, preserving its provenance), `owner_path` (an absolute code path or design-doc path, resolved per REQ-NS4-004), `action` (a one-line directive naming the closing action — e.g. "link this SPEC to a design feature or document its design rationale", "create a SPEC for this unfulfilled design feature or link existing code", "verify the affected doc rows still hold after this edit"), and `confidence` ∈ `{high, medium, low}` (high = owner resolved from a direct path field; medium = owner resolved via graph traversal; low = owner fell back to a design-doc path with no symbol binding). The Route layer SHALL deduplicate work items by `(source_kind, owner_path, source_entry.identifier)` so the same finding promoted across runs produces a stable work-item set.

#### REQ-NS4-004 (Ubiquitous — owner = code path binding, three resolution paths)

Each work item's `owner_path` SHALL resolve to an absolute code path or a design-doc path — NEVER a person, a team name, or a git-blame author. The Route layer SHALL resolve the owner via exactly three paths, one per `source_kind`:

- **audit-orphan → `implementation_path`**: the audit-report.json `orphan[]` entry already carries `implementation_path` (002 SPEC-PROJECT-NAVIGATOR-002, emitted at `.claude/skills/moai-workflow-project/scripts/navigator-audit.sh:575`). `owner_path` = that path, normalized to absolute. The action: the SPEC at this code path has no matching design feature — either connect it to a design decision (`@NAV:DEC`) or document why the code exists without a design-feature anchor.
- **audit-missing → symbol's package via graph, else design-doc `source.file`**: the audit-report.json `missing[]` entry carries `source.file` + `source.heading_path`. M2 first checks whether the design doc at `source.file` references an `@NAV:SYM:<symbol>` token near `source.heading_path` (by scanning the M0 graph for a `sym-edge` whose `source_path` is the design doc and whose target symbol is within the heading's line range). If a symbol is found, `owner_path` = the symbol's owning package (resolved from the symbol node's declaration `source_path` in the M0 graph). If no symbol is found, `owner_path` = the design-doc `source.file` (the doc author's responsibility to create a SPEC or link existing code). The action: create a SPEC for this design feature, or link the existing code that already implements it.
- **detect → `changed_path`**: the M1 detect impact record carries `changed_path`. `owner_path` = `changed_path` verbatim (the engineer who touched this code owns the follow-up: verify the affected doc rows still hold). The action: review the `affected_nodes` + `affected_edges` surfaced by M1 and confirm the bound design decisions / SPECs / symbols are still accurate.

**The owner is always a path**, never a person. A code path survives author turnover; a person does not. This is the falconer binding (design report §6 M2 line 368).

### §C.3 Output artifact

#### REQ-NS4-005 (Ubiquitous — consumer-only on M0/M1/M4 + mx; bridge-not-absorb carries forward)

The Route layer SHALL treat `internal/navigator/sync/` (M0), `internal/navigator/detect/` + `internal/hook/navigator_detect*.go` (M1), `internal/navigator/tiers/` (M4), `internal/mx/`, and the three predecessor Navigator chains as READ-ONLY inputs. The Route layer SHALL NOT modify any producer surface. This carries forward the M0 bridge-not-absorb principle (REQ-NS-005, REQ-NS-012) and the M1 consumer-only principle (REQ-NS2-005) to the M2 read surface: M2 consumes the graph + the detect state + the audit report, never mutates producers.

#### REQ-NS4-006 (Ubiquitous — non-overlap with the 3 predecessors + M0/M1/M4)

The Route layer SHALL NOT write to `.moai/project/navigator/capability-map.md`, `audit-report.{md,json}`, `capability-symbols.{md,json}`, `nav-graph.json`, `tiers.json`, `.moai/project/blueprint/*`, `.moai/decisions/*`, `.moai/project/navigator/symbols/*`, or any `.moai/specs/` SPEC surface. The Route layer SHALL NOT modify the three predecessor Navigator chains' regen / enrich / audit surfaces, M0's sync layer, M1's detect layer, or M4's tiers layer — carries forward REQ-PN-016 (001), REQ-NA-011 (002), REQ-NT-018/019 (003), M0's REQ-NS-013/015/016, M1's REQ-NS2-008, and M4's REQ-NS3-016/017/018.

#### REQ-NS4-007 (Ubiquitous — independent output artifact, work-items.{md,json})

The Route layer's write surface SHALL be limited to exactly two paths: `.moai/project/navigator/work-items.md` (human-readable) and `.moai/project/navigator/work-items.json` (machine-readable), plus their `.tmp` transients during atomic write. The `.json` artifact SHALL carry a `provenance` block (REQ-NS4-008) + a `work_items[]` array where each element is the 5-field work item of REQ-NS4-003. The `.md` artifact SHALL render the same work-item set grouped by `source_kind` (audit-missing / audit-orphan / detect), each row naming the owner path + the action directive. The two files are produced from the same in-memory work-item set in a single pass (no drift between them).

#### REQ-NS4-008 (Ubiquitous — atomic-rename + idempotence + provenance, no wall-clock)

The Route layer SHALL write each output via the atomic-rename pattern (write to `<path>.tmp`, then `os.Rename`) — carrying forward M0's `atomicWrite` (`internal/navigator/sync/write.go`) and 003's `navigator_enrich.go:128` pattern. The `work-items.json` artifact SHALL carry a `provenance` block with `route_commit_sha` (`git rev-parse HEAD`) and `captured_at` (committer date of that SHA, `git log -1 --format=%cI`), using NO wall-clock timestamp, so two runs on the same HEAD with the same inputs produce byte-identical output (idempotence). This carries forward M0's REQ-NS-009, M4's REQ-NS3-019, and 003's `enrich.go:21` `Provenance` contract.

### §C.4 Fail-open + verification

#### REQ-NS4-009 (Event-driven — fail-open on every error mode)

**When** the Route layer encounters any failure — `audit-report.json` absent (002 not yet run), `nav-graph.json` absent (M0 not yet run), the detect state directory absent or empty (M1 not yet run or no edits this session), any input unparseable as JSON, any input schema-invalid, an owner-resolution error (e.g. a symbol node references a non-existent path), or a per-run timeout exceeded — the Route layer SHALL degrade silently to "no work items surfaced": write an empty `work_items: []` set (or write NO output if ALL inputs are absent), return exit code 0, surface NO user-facing error, and append a single diagnostic line to `.moai/logs/navigator-sync.log`. The Route layer SHALL NEVER abort the calling `/moai project` step, SHALL NEVER emit exit 2, and SHALL NEVER cascade a failure into sibling project steps. A partial input set (e.g. audit present + detect absent) SHALL produce work items from the available inputs only, not fail the whole run.

#### REQ-NS4-010 (Capability gate — ≥70% Route accuracy, mechanically measured)

**Where** the ≥70% Route accuracy success metric is evaluated (design report §9 line 442: "missing/orphan 중 actionable work item으로 전환되는 비율"), the accuracy percentage SHALL be produced by the mechanical measurement procedure named in acceptance.md §D.AC-NS4-010 (a fixture corpus of audit-report + detect-state inputs + the Route CLI smoke command emitting the ratio of input findings that yield a work item with a non-empty `owner_path`). The accuracy percentage SHALL NOT be a narrative claim, a sampled estimate, or a carried-over number — every accuracy assertion MUST be attributable to a run command + observed output (per `verification-claim-integrity.md` §1.1 surface 3 + §2 attribution). "Actionable" is defined as: the work item's `owner_path` is non-empty AND its `confidence` is `high` or `medium` (a `low`-confidence owner — a design-doc fallback with no symbol binding — counts toward the denominator but not the numerator, honestly reflecting that the Route layer could not bind a code path).

#### REQ-NS4-011 (Capability gate — Hidden cobra subcommand, sibling pattern)

**Where** the Route layer plugs into the CLI surface, it SHALL register as a Hidden cobra subcommand named `navigator-route` — mirroring the `navigator-sync` Hidden subcommand (M0, `internal/cli/navigator_sync.go:22-32`), the `navigator-enrich` Hidden subcommand (003, `internal/cli/navigator_enrich.go:31`), and the `navigator-tiers` Hidden subcommand (M4, `internal/cli/navigator_tiers.go:35-41`). The Route layer SHALL NOT introduce a new top-level `moai` subcommand, SHALL NOT create a new wrapper script, and SHALL be invoked from the existing `/moai project` surface as a sibling step (the same way M0/003/M4 are wired). A CI guard test SHALL assert `navigator-route` is Hidden (mirroring `internal/cli/navigator_tiers_test.go:65-70` which asserts the same for `navigator-tiers`).

#### REQ-NS4-012 (Capability gate — template-first distribution)

**Where** the Route layer ships any new distributed surface — a new `settings.json` config key, a new hook block, or any artifact under `.claude/` — the template source at `internal/template/templates/.claude/` SHALL carry the change first, `make build` SHALL regenerate the embedded catalog (`catalog.yaml`), and the local copy SHALL mirror — per CLAUDE.local.md §2 [HARD] Template-First Rule + §25 Template Internal-Content Isolation. The run-phase touch list (template paths + mirror paths) is named in plan.md §D so compliance is planned, not discovered late. If the Route layer ships NO distributed surface (env-var-only gate or pure CLI), this REQ reduces to "no template path in the diff, no catalog regen required" — documented in the PR body.

## §D. Constraints (Non-Functional)

- **Performance budget**: the Route layer runs inside the existing `/moai project` timeout envelope. The full promote-and-write cycle SHALL complete in < 500ms p99 for an input set of up to 200 audit findings + 500 detect rows (the current repo's audit is O(tens); the budget headroom is for fan-out). The Route layer SHALL honor the existing context-cancellation discipline (no hardcoded sleeps; `context.WithTimeout`).
- **Read-only invariant**: the Route layer's only write surfaces are `work-items.{md,json}` (REQ-NS4-007) + `.moai/logs/navigator-sync.log` (advisory diagnostic, REQ-NS4-009). No other file mutation. No external-tracker writes (no GitHub issue, no Linear ticket, no SPEC draft — M2 emits a local artifact only).
- **Idempotence / determinism**: for the same input set (audit-report.json + detect state + nav-graph.json at the same HEAD) + the same Route layer code, the Route layer SHALL produce the same `work-items.{md,json}` byte-for-byte (sorted work-item array, no wall-clock, provenance-attributed per REQ-NS4-008). Re-running `navigator-route` on unchanged inputs is a no-op write (the `.tmp` + rename produces a byte-identical file).
- **Owner = path, never person**: REQ-NS4-004 is the load-binding contract. The Route layer SHALL NOT call `git blame`, SHALL NOT resolve a `CODEOWNERS` file, and SHALL NOT emit a person/team/email in the `owner_path` field. The owner is a code path or a design-doc path, full stop.
- **Provenance alignment** (`verification-claim-integrity.md`): every accuracy / promotion-rate claim in an AC is attributable to a mechanical command + observed output. The `work-items.json` artifact is the Evidence; the promotion claim is the Claim.
- **Non-overlap**: REQ-NS4-005 / REQ-NS4-006 / REQ-NS4-007 are inviolable.
- **Language neutrality**: the Route layer's owner-resolution is language-neutral (operates on paths + graph nodes, not source ASTs); it inherits the language coverage of the M0 graph + the 002 audit without re-detecting language markers.
- **Fail-open**: REQ-NS4-009 binds every error mode to exit 0 + empty/log, never abort.

## §E. Verification Surface

- **Coverage AC**: acceptance.md §D.AC-NS4-010 names the exact fixture-corpus directory, the Route CLI smoke command, and the ratio formula. The percentage is the observed output of that command, not a narrative.
- **Fail-open ACs**: one AC per error mode (audit absent / detect state absent / nav-graph absent / unparseable JSON / schema-invalid / owner-resolution error / timeout) — each Given-When-Then, each asserting exit 0 + empty work-item set (or no output) + a log line.
- **Owner-binding AC**: a test asserting every emitted `owner_path` resolves to a path (never a person/team/email), across all three `source_kind` resolution paths.
- **Non-overlap AC**: a grep-based test asserting the Route source files do not name forbidden write surfaces (carries forward M0's `nonoverlap_test.go` + M1's `navigator_detect_nonoverlap_test.go` pattern).
- **Idempotence AC**: two consecutive `navigator-route` runs on the same inputs produce byte-identical `work-items.json` (provenance included — same HEAD, same `route_commit_sha`, same `captured_at`).
- **Hidden-subcommand AC**: a test asserting `navigator-route` is Hidden, mirroring `navigator_tiers_test.go:65-70`.

## §F. Out of Scope

### Out of Scope — M3 Fix (AI-drafted incremental regen)

- Auto-drafting an incremental regeneration of `capability-map.md` / `audit-report.{md,json}` / `nav-graph.json` / `tiers.json` to absorb a work item. Owned by M3 Fix (separate SPEC, design report §6 M3 line 371-375).
- Invoking any LLM to draft a fix, diff, or patch for a work item. M2 produces the work item (the "what + who"); M3 produces the fix (the "how").
- A 1-click approval UI for AI-drafted fixes. That is M3's `--compare-to` approval surface.

### Out of Scope — PostToolUse real-time work-item promotion

- Registering a PostToolUse hook branch that promotes each detection into a work item in real time. M2 is on-demand CLI only (REQ-NS4-001, decision defended in plan.md §C.1). A real-time path would (a) usually find no audit to promote (audit is on-demand), and (b) flood the engineer with one work item per keystroke — the cadence is wrong for a persisted, owner-bound artifact. M1 already owns the real-time advisory surface (`systemMessage`); M2 owns the coarser-cadence persisted work-item surface. A future real-time path is deferred to acceptance.md §G Forward-Looking Checks.

### Out of Scope — external-tracker writes

- Creating a GitHub issue, a Linear ticket, a `.moai/specs/` SPEC draft, or a TODO file from a work item. M2 emits `work-items.{md,json}` — a local, regenerable artifact — and does NOT write to any external tracker. Tracker integration (if ever desired) is a follow-up SPEC that would consume `work-items.json` read-only, the same way M2 consumes audit-report.json read-only.

### Out of Scope — owner = person / team / CODEOWNERS

- Resolving the owner to a person via `git blame`, a team via `CODEOWNERS`, or an email via any means. The owner is a code path or a design-doc path, never a person (REQ-NS4-004, falconer binding). A person-binding feature is explicitly out of scope — it would couple the work item to author turnover, which is the anti-pattern the falconer principle exists to prevent.

### Out of Scope — M0/M1/M4 producer modification

- Modifying `internal/navigator/sync/` (M0), `internal/navigator/detect/` or `internal/hook/navigator_detect*.go` or `internal/hook/post_tool.go` (M1), `internal/navigator/tiers/` (M4), `internal/mx/`, or the three predecessor chains (`internal/navigator/astx/`, `navigator-audit.sh`, `navigator-regen.sh`, `navigator-enrich.sh`). REQ-NS4-005 / REQ-NS4-006 forbid this.

### Out of Scope — M5 brownfield reverse-extraction

- Reverse-extracting design docs from existing code (M5, design report §6 line 383-387). M2 routes existing findings; it does not reverse-extract new ones.

## §G. Related SPECs (non-blocking)

- **SPEC-NAVIGATOR-SYNC-001** (M0, `status: completed`, PR #1375 `f4bd58acbd`) — produces `nav-graph.json`, the graph context M2 consumes for owner resolution. `depends_on` target (the hard dependency — M2 cannot resolve symbol→package owners without the graph).
- **SPEC-NAVIGATOR-SYNC-002** (M1, `status: completed`, PR #1379 `304907b6d`) — produces the detect impact record at `.moai/state/navigator-detect/`; REQ-NS2-003 explicitly reserves work-item promotion for M2. Referenced via `related_specs` (not `depends_on`): M2 degrades gracefully when detect state is absent (fail-open, REQ-NS4-009), so M1 is a soft input, not a hard dependency.
- **SPEC-NAVIGATOR-SYNC-003** (M4, `status: completed`, PRs #1384/#1385) — produces `tiers.json`; M2 does not consume tiers (M2's owner resolution uses M0's graph + 002's audit, not M4's tier overlays), but M4's non-overlap contract (REQ-NS3-016/017/018) is carried forward as REQ-NS4-006.
- **SPEC-PROJECT-NAVIGATOR-001** (regen, completed) — produces `capability-map.md`. M2 does not consume it directly (audit-report.json is the roll-up), but the non-overlap boundary REQ-PN-016 carries forward.
- **SPEC-PROJECT-NAVIGATOR-002** (audit, completed) — produces `audit-report.{md,json}`, the PRIMARY input M2 consumes. The `implementation_path` field (002's orphan-owner anchor) is the load-bearing input for REQ-NS4-004's orphan-resolution path. Non-overlap REQ-NA-011 carries forward as REQ-NS4-006.
- **SPEC-PROJECT-NAVIGATOR-003** (enrich, completed) — produces `capability-symbols.json`. M2 does not consume it directly, but the non-overlap boundary REQ-NT-018/019 carries forward.

## §H. Cross-References

- `.moai/reports/navigator-redesign-bas-20260805.html` — design report. §3.3(C) Falconer 3-element loop (line 268-284, M2 = Route line 274), §4 asset reuse (line 287-308, `navigator-audit.sh` token-normalization matching line 300), §6 M2 milestone card (line 365-369), §7 slice table (line 398, M2 = Tier M, deps M0 parallel with M1), §9 success metrics (line 442, Route accuracy ≥ 70%).
- `.claude/rules/moai/workflow/nav-tokens.md` — the binding-token trio author surface (M0); the `@NAV:SYM:<symbol>` token is what M2's missing→symbol owner resolution consults via the graph.
- `internal/navigator/sync/{schema,scan,join,write,mx_bridge}.go` — M0 graph-join layer (consumed read-only).
- `internal/navigator/detect/traverse.go` — M1 reverse-traversal engine (consumed read-only via its JSONL output).
- `.claude/skills/moai-workflow-project/scripts/navigator-audit.sh:553-591` — the audit-report.json emitter whose schema (`missing[]` with `source.file`, `orphan[]` with `implementation_path`) M2 consumes.
- `internal/cli/navigator_sync.go:22-32` + `internal/cli/navigator_tiers.go:35-41` — the Hidden-subcommand sibling pattern M2 mirrors (REQ-NS4-011).
- `.claude/rules/moai/core/verification-claim-integrity.md` §1.1 surface 3 + §2 — accuracy-claim attribution (REQ-NS4-010).
- CLAUDE.local.md §2 [HARD] Template-First Rule + §25 Template Internal-Content Isolation (REQ-NS4-012).

---
id: SPEC-CC2219-UPSTREAM-ALIGN-001
title: "Claude Code 2.1.208..2.1.219 upstream alignment umbrella (GD-1/2/4/5/6/7/8/9)"
version: "0.1.0"
status: draft
created: 2026-07-25
updated: 2026-07-25
author: manager-spec
priority: P1
phase: "v3.1.x target"
module: ".claude/rules + .claude/agents + internal/template + internal/web + docs-site"
lifecycle: spec-anchored
tier: L
tags: "upstream-align, claude-code, doctrine, nesting, permission-mode, opus-5, hooks, template-mirror"
related_specs: [SPEC-SUBAGENT-NESTING-DOCTRINE-001, SPEC-MODEL-PROFILE-MATRIX-001]
---

# SPEC-CC2219-UPSTREAM-ALIGN-001 — CC 2.1.219 Upstream Alignment Umbrella

## HISTORY

| Version | Date | Author | Change |
|---|---|---|---|
| 0.1.0 | 2026-07-25 | manager-spec | Initial draft — umbrella SPEC per research report §7 Option A (GD-3 excluded as landed hotfix) |
| 0.1.1 | 2026-07-25 | manager-spec | plan-audit iter-1 (PASS 0.86) D1-D3 + S1-S4 applied: tier frontmatter, AC-X-005, AC-X-003 re-anchor to origin/main, research.md pointer, AC label/path fixes |

## §A Context

Claude Code upstream delta `2.1.207 (exclusive) .. 2.1.219` (11 released versions, ~335 items) produced **9 genuine-drift clusters** against live MoAI doctrine, agent definitions, templates, hooks, and Go code. Evidence SSOT: `.moai/research/cc-update-2.1.207-to-2.1.219.md` (§4 per-cluster detail, §6 verification statement, §9 file inventory).

This umbrella SPEC covers clusters **GD-1, GD-2, GD-4, GD-5, GD-6, GD-7, GD-8, GD-9**, organized as children **A** (GD-1+GD-2 nesting/permission doctrine), **C** (GD-4 Opus 5 migration incl. Go code), **D** (GD-5/6/7 doc sync), **E** (GD-8/9 doc sync), **F** (docs-site + README propagation).

### §A.1 Fresh empirical evidence — GD-1 three-way conflict RESOLVED (2026-07-25 probe)

The research report's Gap 1/2/3 (§6) recorded an unresolved three-way conflict (changelog: default depth 3; official sub-agents doc: default off; MoAI doctrine: default 0=off). A same-day probe on the installed 2.1.219 binary resolved it:

- A `sync-auditor` subagent (which carries `Agent` in `tools`) ran with `CLAUDE_CODE_MAX_SUBAGENT_SPAWN_DEPTH` verifiably **UNSET** (`depth env: [unset]`) and **successfully spawned** an `Explore` child, which returned "NESTING-OK".
- **Conclusion**: on 2.1.219, subagent nesting is **ENABLED BY DEFAULT** (changelog confirmed; official doc lags). MoAI's "env-default-off double guarantee" (CLAUDE.md §4 Watch note; SPEC-SUBAGENT-NESTING-DOCTRINE-001 REQ-SND-015/AC-SND-012 encoding) is **stale**. The ONLY remaining flat-hierarchy guard for an `Agent`-carrying subagent is **omitting the `Agent` tool**.
- **Probe caveats (bounded evidence)**: single trial; depth-1 nesting only (the depth-3 ceiling was not probed); env-var propagation from main session vs subagent was not separately verified. Doctrine rewrites MUST carry these caveats where they assert the ceiling value.

### §A.2 GD-1 × GD-2 interaction (safety core)

Per report §4 GD-2: the Task tool spawn-time `mode` parameter is **deprecated/ignored** since 2.1.213; subagents inherit the parent's permission mode, and a parent in `bypassPermissions`/`acceptEdits` takes precedence and cannot be overridden. Combined with §A.1: `sync-auditor` nests by default in every shipped project, and its children — documented as read-only via `mode: "plan"` — are **not actually constrained**. Tool restriction (`Explore`, or `tools:` without Write/Edit) is the only robust read-only mechanism. This interaction drives the child-A requirements.

## §B Requirements (GEARS)

### §B.1 Child A — GD-1: Subagent nesting doctrine re-correction

- **REQ-GD1-001**: When the nesting doctrine surfaces are rewritten, the doctrine shall state that on Claude Code v2.1.219+ subagent nesting is **enabled by default** (changelog: depth 3), that `CLAUDE_CODE_MAX_SUBAGENT_SPAWN_DEPTH=1` disables nesting, and that the former "defaults to 0 = off" claim applied only to v2.1.217–2.1.218.
- **REQ-GD1-002**: The doctrine shall state that for an `Agent`-carrying subagent, **omission of the `Agent` tool is the sole remaining flat-hierarchy guarantee** — the "double guarantee" framing shall be removed from every surface that carries it.
- **REQ-GD1-003**: The rewrite shall update the following live doctrine files and their template mirrors (report §4 GD-1, line numbers indicative — anchor by quoted stale text), keeping live/template byte-parity where the mirror is byte-identical class:
  1. `CLAUDE.md` (§4 Watch note + §14 depth-cap framing)
  2. `.claude/rules/moai/development/agent-authoring.md` (three stale sites: "default off since v2.1.217", "defaults to off", "double guarantee")
  3. `.claude/rules/moai/development/agent-patterns.md` ("as of v2.1.217 the runtime default is off")
  4. `.claude/rules/moai/workflow/orchestration-mode-selection.md` ("the runtime default is now off … (default `0`)")
  5. `.claude/agents/moai/sync-auditor.md` ("shipped distribution leaves … unset, so the default behavior is flat")
  6. `internal/template/templates/CLAUDE.md` + the `internal/template/templates/.claude/` mirrors of files 2-5.
- **REQ-GD1-004**: Where the doctrine asserts the depth-3 ceiling or env-propagation behavior, it shall annotate that the empirical probe covered a single depth-1 trial only (§A.1 caveats), so ceiling claims are changelog-sourced, not observed.
- **REQ-GD1-005**: The `sync-auditor` read-only nesting pilot shall be **re-evaluated**: While the parent session may run `bypassPermissions`/`acceptEdits` (GD-2), the pilot's safety rationale ("children are read-only via `mode: plan`") no longer holds; the SPEC's run phase shall either (a) retire `Agent` from `sync-auditor.md` `tools` (restoring flat by omission), or (b) retain the pilot with a rewritten safety rationale based exclusively on tool-restricted children (`Explore` only, never `general-purpose` with write tools). The chosen option is a plan-phase-flagged decision requiring orchestrator/user confirmation before M1 execution (see plan.md §F M1).
- **REQ-GD1-006**: The doctrine shall record that this SPEC partially supersedes the nesting encoding of SPEC-SUBAGENT-NESTING-DOCTRINE-001 (REQ-SND-015/AC-SND-012 "shipped default stays flat" premise), via a cross-reference note — the prior SPEC's frontmatter is NOT rewritten by this SPEC.

### §B.2 Child A — GD-2: Deprecated spawn-time `mode` parameter

- **REQ-GD2-001**: The doctrine shall state that the Task/Agent spawn-time `mode` parameter is deprecated and **ignored** since v2.1.213, that subagents inherit the parent session's permission mode, and that a parent `bypassPermissions`/`acceptEdits` mode takes precedence over any child `permissionMode` frontmatter.
- **REQ-GD2-002**: Every doctrine surface asserting read-only enforcement via `mode: "plan"` (report §4 GD-2 table: CLAUDE.md L64, `sync-auditor.md` L127, `worktree-integration.md` L179+L250, `agent-authoring.md` L163/L232/L235, + template mirrors) shall be rewritten to ground read-only guarantees in **tool restriction** (`Explore`, or a `tools:` list omitting Write/Edit/NotebookEdit) instead.
- **REQ-GD2-003**: Where an example spawn snippet carries `mode: "plan"`, the snippet shall be updated to the tool-restriction pattern; the doctrine shall not retain `mode: "plan"` as a recommended mechanism on any live or template surface (historical SPEC/report artifacts under `.moai/specs/` and `.moai/reports/` are exempt).
- **REQ-GD2-004**: The doctrine shall not claim GD-2 runtime behavior was empirically observed — it rests on the 2.1.213 changelog + official-doc frontmatter table (report §6 Gap 4); the rewrite shall cite that provenance.

### §B.3 Child C — GD-4: Opus 5 default-Opus migration (includes Go code)

- **REQ-GD4-001**: When the model policy is updated, `internal/template/model_policy.go` shall resolve the `opus` alias to `claude-opus-5`, add a deprecated-canonical row migrating `claude-opus-4-8`, and re-verify `opus[1m]` picker semantics against Opus 5's native 1M context.
- **REQ-GD4-002**: **@MX:ANCHOR hazard**: `ModelAliasTable` is an @MX:ANCHORed symbol with fan_in ≥ 3 (`launcher.go` `expandModelString`, `profile_setup.go` `normalizeModel`, `settings/schema.go` `modelOptions`). The run phase shall treat this edit as invariant-contract work: all three consumers re-verified, `make build` + `go test ./internal/template/... ./internal/cli/...` green, template mirror-parity + neutrality guards re-run.
- **REQ-GD4-003**: `internal/web/root_templ.go` appbar default (`claude-opus-4-8` literal) and `appbar_context_test.go` shall be updated to the Opus 5 id; templ regeneration follows the go.mod-pinned templ version.
- **REQ-GD4-004**: The doctrine naming sweep shall update the stale "opus = Opus 4.8" alias claim (`model-policy.md` L19) and the Opus-4.7/4.8 naming across the report-§4-GD-4 surface list (`moai-constitution.md` heading, CLAUDE.md §12, `context-window-management.md` threshold table gains an Opus 5 (1M) row, `quality.yaml.tmpl` effort notes, `harness.yaml` local-only, + the ~10-file naming sweep), preserving the still-true effort facts (default `high`; `xhigh`/`max` available on Opus 5/Sonnet 5/Opus 4.8/4.7; Opus 5 carries effort choice across sessions — no hold).
- **REQ-GD4-005**: The `context-window-management.md` table's bare "Opus/Fable (256K)" row shall be disambiguated so "Opus" no longer ambiguously covers Opus 5 (1M).

### §B.4 Child D — GD-5/GD-6/GD-7: native-invocation + workflow-size + hook-event doc sync

- **REQ-GD5-001**: `native-invocation-model.md` matrix rows for `/code-review` and `/deep-research` (and the `/verify` reference) shall be re-marked from auto-invocable PROGRAMMATIC to manual-invocation-only per 2.1.215/2.1.218, and the Axis A worked recommendation (L67/L87) shall be annotated accordingly; CLAUDE.md L211 + `dynamic-workflows.md` L78 `/deep-research` prose shall note manual-only invocation.
- **REQ-GD6-001**: `dynamic-workflows.md` L61 shall be updated: size enum gains `unrestricted`, default is explicitly `medium` (<15 agents), and the new `workflowSizeGuideline` settings key exists (hiding the `/config` row when set); `settings-management.md` settings-key table gains a `workflowSizeGuideline` row.
- **REQ-GD7-001**: `hooks-system.md` (live + template) event catalog and exit-code table shall gain the `DirectoryAdded` event (fires after `/add-dir` or SDK `register_repo_root`), annotated that the official hooks doc lags (report §4 GD-7) and that MoAI wires no handler for it.

### §B.5 Child E — GD-8/GD-9: fork/subtask + `context: fork` skill doc sync

- **REQ-GD8-001**: `agent-authoring.md` § Fork Subagents (L100, + template mirror) shall be updated: the in-session fork is invoked with `/subtask`; `/fork` now creates a background session copy; the upstream 2.1.212-vs-2.1.213 version inconsistency is recorded as noted-not-resolved.
- **REQ-GD9-001**: `skill-authoring.md` L28 and `moai-foundation-cc/reference/claude-code-skills-official.md` L79/L106 (+ mirrors) shall document that `context: fork` skills run in the background by default as of 2.1.218, with per-skill `background: false` opt-out; annotated that MoAI ships no `context: fork` skill (doc-only impact).

### §B.6 Child F — docs-site 4-locale + README propagation

- **REQ-F-001**: After children A/C/D/E land, docs-site content (`docs-site/content/{ko,en,ja,zh}/**`) and `README.md`/`README.ko.md` shall be grep-swept for the same stale claims (nesting default-off, `mode: "plan"` read-only, Opus 4.8 as default Opus, auto-invoked `/deep-research`) and updated with 4-locale same-PR parity per `.moai/docs/docs-site-i18n-rules.md`.
- **REQ-F-002**: Where a docs-site surface has no counterpart drift (grep returns 0), the milestone shall record the 0-match evidence rather than editing speculatively.

### §B.7 Cross-cutting

- **REQ-X-001**: Every `.claude/` edit shall carry its paired `internal/template/templates/` mirror edit (Template-First, CLAUDE.local.md §2), except surfaces documented as intentionally local-only/neutrality-divergent; `make build` regenerates the embedded FS after mirror edits.
- **REQ-X-002**: Template mirror edits shall pass the §25 neutrality guards (`internal_content_leak_test.go`, `template-neutrality-check.yaml`) — no SPEC IDs, internal dates, or commit SHAs introduced into template mirrors.
- **REQ-X-003**: The run phase shall not weaken any approval gate: the nesting/permission rewrite changes documentation of runtime behavior, not MoAI's own concurrency safeguard (one write-capable agent at a time), which is preserved verbatim.

## §C Scope

### In scope
- Clusters GD-1, GD-2 (child A), GD-4 (child C), GD-5/6/7 (child D), GD-8/9 (child E), docs-site/README propagation (child F).
- Doctrine rule files, agent definitions, CLAUDE.md, template mirrors, `internal/template/model_policy.go`, `internal/web` appbar surfaces, `quality.yaml.tmpl`, docs-site 4-locale content.

### Out of scope

The literal exclusions below bound this SPEC — out of scope items are NOT built here.

### Out of Scope — GD-3 SessionStart fork matcher (child B)
- GD-3 (`fork` source missing from the SessionStart matcher) already landed as a separate hotfix — **PR #1146** (matcher fork fix). This SPEC records the exclusion and makes no GD-3 edit; acceptance verification only confirms non-regression by observation, not re-implementation.

### Out of Scope — Tier 2/3 items
- All ~24 Tier 2 watch items and ~293 Tier 3 items from report §3 (e.g. T2-a AskUserQuestion neutral wording annotation, T2-b worktree baseline bump, T2-k `subagentStatusLine` reasoning-effort enhancement) — candidates for future SPECs, not edited here.

### Out of Scope — deeper empirical probing
- Depth-3 ceiling probing, env-propagation matrix testing, and GD-2 runtime spawn observation — doctrine carries provenance caveats (REQ-GD1-004, REQ-GD2-004) instead of new probes.

### Out of Scope — SPEC-SUBAGENT-NESTING-DOCTRINE-001 artifact rewrite
- The prior SPEC's own artifacts are not amended; supersession is recorded by cross-reference note only (REQ-GD1-006).

## §D Acceptance Criteria

See `acceptance.md` — per-cluster AC groups AC-GD1-*, AC-GD2-*, AC-GD4-*, AC-D-*, AC-E-*, AC-F-*, AC-X-*.

## §H Cross-References

- Evidence SSOT: `.moai/research/cc-update-2.1.207-to-2.1.219.md`
- Superseded encoding: `.moai/specs/SPEC-SUBAGENT-NESTING-DOCTRINE-001/` (PR #1133)
- GD-3 hotfix: PR #1146
- Template-First + neutrality: CLAUDE.local.md §2/§25

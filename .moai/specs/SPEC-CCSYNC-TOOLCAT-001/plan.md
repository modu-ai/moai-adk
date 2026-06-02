# Plan — SPEC-CCSYNC-TOOLCAT-001

> Implementation plan for the Claude Code tool-catalog sync (remove `MultiEdit`,
> migrate `TodoWrite` → Task* family). Plan-phase artifact authored by
> manager-spec. Run-phase execution owned by manager-develop.

## A. Context

This SPEC remediates two verified tool-catalog drift findings (H3 retired
`MultiEdit`, H4 default-disabled `TodoWrite`) plus a low-priority bundle
(foundation-cc authoring-kit scrub + a NEW Go CI guard test). Baseline HEAD
`5042e309c`. The 7 retained agents at this baseline are manager-spec,
manager-develop, manager-docs, manager-git, plan-auditor, **sync-auditor**,
builder-harness.

The work is content-edit + one NEW test file + build-regeneration. No Go logic
changes to shipped binaries (the only Go file is the new test). The primary risks
are (1) mirror byte-parity drift across the many edited agent files, (2) a
shared-line edit collision with the sibling SPEC-CCSYNC-CLAUDEMD-001 over
`agent-authoring.md` line ~212, and (3) over-broad scrubbing of foundation-cc
(verbatim official-docs reproductions must stay untouched).

## B. Known Issues / Pre-existing State (verified at HEAD 5042e309c)

| Item | Location | Current state |
|------|----------|---------------|
| MultiEdit in tools | manager-spec.md:13, manager-develop.md:21 (+ mirrors) | declared in `tools:` |
| MultiEdit in matchers | manager-develop.md:39,45 (+ mirror) | `matcher: "Write\|Edit\|MultiEdit"` |
| MultiEdit body instr | manager-spec.md:130 (+ mirror) | "[HARD] Use MultiEdit for simultaneous 3-file creation" |
| MultiEdit incidental | manager-spec.md:140,153,207 (+ mirror) | "Write/MultiEdit" mentions in SPEC ID self-check |
| MultiEdit settings | settings.json.tmpl:81 (matcher), :383 (permissions array) | declared |
| TodoWrite in tools | manager-spec:13, manager-develop:21, manager-docs:13, manager-git:12, builder-harness:11 (+ mirrors) | declared (5 agents) |
| TodoWrite read-only | plan-auditor, sync-auditor | NOT declared — correct, out of scope |
| TodoWrite recommendation | agent-authoring.md:212 (+ mirror) | "… Skill, TodoWrite (…)" |
| Shared-line claim | agent-authoring.md:212 (+ mirror) | co-located "teammates CAN spawn other teammates … v2.1.50+" — sibling SPEC owns this |
| foundation-cc authoring examples | skill-formatting-guide.md:144; sub-agent-examples.md:28,196,437,934; sub-agent-formatting-guide.md:154,390,848,892; sub-agent-integration-patterns.md:768 (+ mirrors) | MoAI-authored examples teaching MultiEdit/TodoWrite |
| foundation-cc official docs | claude-code-iam-official.md:24,89; claude-code-sub-agents-official.md:334; claude-code-plugins-official.md:190 (+ mirrors) | verbatim official-docs — OUT OF SCOPE (EXC-3) |
| CI guard | internal/template/ | no tool-catalog guard test exists |

Note: line numbers were re-verified at HEAD `5042e309c` during plan authoring.
The run phase MUST Grep/Read each file to re-confirm exact lines before editing
(a parallel session may shift lines by a few).

## C. Pre-flight Checklist (run-phase entry)

- [ ] `git fetch origin main && git rev-list --count --left-right origin/main...HEAD` — confirm no parallel-session divergence (per agent-common-protocol §Pre-Spawn Sync Check).
- [ ] **Confirm SPEC-CCSYNC-CLAUDEMD-001 (sibling sharing `agent-authoring.md` line ~212) is committed/merged and NOT mid-run.** This SPEC runs SECOND and rebases onto the sibling's agent-authoring.md commit (CON-5 / §H). If the sibling is mid-run, STOP and return a blocker report.
- [ ] Re-Grep all `MultiEdit` and `TodoWrite` locations in §B to confirm current line numbers (especially agent-authoring.md:212 which the sibling may have shifted).
- [ ] Confirm `make build` works in a clean tree before starting (baseline green).
- [ ] Confirm `go test ./internal/template/...` is green at baseline (so the NEW guard test's RED→GREEN is attributable to this SPEC).

## D. Constraints (binding)

- **Template-First** (CLAUDE.local.md §2): edit `internal/template/templates/` source AND the dev-root `.claude/**` copy, then `make build`. Never hand-edit `embedded.go`.
- **Mirror parity** (`internal/template/rule_template_mirror_test.go` + embedded mirror guards): every agent file, agent-authoring.md, and each foundation-cc reference has a `templates/` mirror — edit both copies in the same commit.
- **Template Internal-Content Isolation** (CLAUDE.local.md §25): no internal SPEC IDs / REQ / AC / audit citations / dates / SHAs in any template-written text. The NEW guard test's code comment documenting the allowlist source MUST stay generic (cite "the current Claude Code tool catalog", not this SPEC ID).
- **Scrub scope discipline** (EXC-3): edit ONLY the MoAI-authored foundation-cc authoring-kit examples; do NOT touch the verbatim official-docs reproductions.
- **Hybrid Trunk Tier M** (CLAUDE.local.md §23.9): main-direct, no PR, unless user escalates with `--pr`. No `manager-git` routing by default.
- **No `--no-verify` / `--amend` / force-push** (manager-develop-prompt-template §B9).

## E. Self-Verification (run-phase exit, read-only batch)

The run phase must execute the acceptance.md AC greps as a single parallel Bash
batch (per verification-batch-pattern.md), plus:

```bash
go test ./internal/template/...                                  # mirror-drift + neutrality + leak + NEW guard
go test ./internal/template/... -run TestToolCatalog            # NEW guard test isolated (exact name set in M4)
make build && git diff --stat internal/template/embedded.go     # embedded regenerated
go build ./... && GOOS=windows GOARCH=amd64 go build ./...       # cross-platform (new test file compiles both)
```

## F. Milestones (priority-ordered, no time estimates)

### M1 — Finding H3: MultiEdit removal (Priority High)

Remove every `MultiEdit` reference except verbatim official-docs (EXC-3).

- M1.1: Re-Grep all `MultiEdit` locations (§B) to confirm current lines.
- M1.2: Remove `MultiEdit` from `tools:` frontmatter of manager-spec (line ~13) and manager-develop (line ~21) — source + template mirror.
- M1.3: Replace the manager-spec body "[HARD] Use MultiEdit for simultaneous 3-file creation (60% faster than sequential)" instruction (line ~130) with "[HARD] Make parallel `Edit`/`Write` calls in a single turn for simultaneous multi-file creation" — source + template mirror.
- M1.4: Correct the manager-spec body incidental "Write/MultiEdit" mentions (lines ~140, ~153, ~207) to "Write" or "Write/Edit" — source + template mirror.
- M1.5: Drop `MultiEdit` from manager-develop hook matchers (lines ~39, ~45 → `"Write|Edit"`) — source + template mirror.
- M1.6: Drop `MultiEdit` from settings.json.tmpl matcher (line ~81 → `"Write|Edit"`) and remove the `"MultiEdit"` permissions-array entry (line ~383).

Covers REQ-CCSYNC-T-001, REQ-CCSYNC-T-002, REQ-CCSYNC-T-003, REQ-CCSYNC-T-004.

### M2 — Finding H4: TodoWrite → Task* migration (Priority High)

Replace `TodoWrite` with the Task* family in the 5 implementation-capable agents.

- M2.1: For each of manager-spec (line ~13), manager-develop (line ~21), manager-docs (line ~13), manager-git (line ~12), builder-harness (line ~11): replace `TodoWrite` with `TaskCreate, TaskUpdate, TaskList, TaskGet` in the `tools:` frontmatter — source + template mirror.
- M2.2: Leave `plan-auditor` and `sync-auditor` untouched (EXC-1 — they declare no task-tracking tool).
- M2.3: Update agent-authoring.md line ~212 recommendation: `… Skill, TodoWrite` → `… Skill, TaskCreate, TaskUpdate, TaskList, TaskGet` — source + template mirror. **Edit ONLY the TodoWrite→Task* portion; leave the teammates-spawn/version claim for the sibling (EXC-4 / CON-5).**

Covers REQ-CCSYNC-T-005, REQ-CCSYNC-T-006, REQ-CCSYNC-T-007.

### M3 — Low-priority: foundation-cc authoring-kit scrub (Priority Medium)

- M3.1: Re-Grep `MultiEdit`/`TodoWrite` in the MoAI-authored authoring-kit examples (B.3.1 category 1) to confirm current lines.
- M3.2: In `skill-formatting-guide.md`, `sub-agent-examples.md`, `sub-agent-formatting-guide.md`, `sub-agent-integration-patterns.md` (source + template mirror): remove `MultiEdit` from example `tools:`/`allowed-tools:` lines; replace bare `TodoWrite` with the Task* family (or remove) in those examples; remove the "MultiEdit tool: Use with caution …" guidance bullet (sub-agent-formatting-guide.md ~line 848) or rephrase to a current tool.
- M3.3: Do NOT edit `claude-code-iam-official.md`, `claude-code-sub-agents-official.md`, `claude-code-plugins-official.md` (EXC-3 — verbatim official-docs).

Covers REQ-CCSYNC-T-008.

### M4 — NEW CI guard test (Priority High)

- M4.1: Add `internal/template/tool_catalog_audit_test.go` (TDD: write it to FAIL first against the pre-M1/M2 state, confirming it catches the drift, then GREEN after M1-M3).
- M4.2: The test walks the embedded agent files; parses each `tools:` frontmatter line; SHOULD also walk skill `allowed-tools:` lines. It declares an allowlist of current Claude Code tool-catalog names in a code comment citing the catalog source (generic — no SPEC ID per CON-3 §25).
- M4.3: Assertions: (a) hard-fail on any `MultiEdit` token in any agent/skill `tools:`/`allowed-tools:` line; (b) hard-fail on any retained-agent `tools:` line containing `TodoWrite` (default-disabled — migration target is Task*). Suggested test names: `TestToolCatalogNoMultiEdit`, `TestToolCatalogNoTodoWrite` (final names set during implementation; acceptance.md AC uses the `TestToolCatalog` prefix grep).
- M4.4: Run the test in isolation to confirm RED→GREEN attribution.

Covers REQ-CCSYNC-T-009.

### M5 — Build regeneration + full verify (Priority High, must follow M1-M4)

- M5.1: `make build` to regenerate `internal/template/embedded.go`.
- M5.2: Run the acceptance.md AC grep batch + `go test ./internal/template/...` (mirror-drift + neutrality + leak + NEW guard) + `go build ./...` + `GOOS=windows GOARCH=amd64 go build ./...` in a single parallel Bash batch.
- M5.3: Confirm `git diff --stat internal/template/embedded.go` shows the regeneration (or `git diff --exit-code internal/template/embedded.go` is non-zero before commit, zero after `make build`).

Covers REQ-CCSYNC-T-010, REQ-CCSYNC-T-011, REQ-CCSYNC-T-012, and the cross-cutting verification of all prior milestones.

## G. Anti-Patterns to Avoid (run-phase)

- **AP-1**: Editing `embedded.go` directly instead of regenerating via `make build`.
- **AP-2**: Editing only one side of a mirror pair (agent file, agent-authoring.md, or a foundation-cc reference) — breaks the mirror-drift guard.
- **AP-3**: Scrubbing the verbatim official-docs reproductions (EXC-3) — those document the historical catalog and must stay byte-for-byte.
- **AP-4**: Editing the agent-authoring.md teammates-spawn / v2.1.50 claim on line 212 (EXC-4 — sibling SPEC owns it). Touch ONLY the TodoWrite→Task* portion.
- **AP-5**: Running concurrently with SPEC-CCSYNC-CLAUDEMD-001 on agent-authoring.md (CON-5 — this SPEC runs SECOND; sibling first).
- **AP-6**: Adding the Task* family to `plan-auditor` / `sync-auditor` (EXC-1 — read-only, no task tracking).
- **AP-7**: Embedding this SPEC's ID / REQ tokens / date in the NEW test's allowlist comment (CON-3 §25 — keep it generic; the test lives under `internal/template/` whose neutrality leak guard scans it).
- **AP-8**: Over-broad tool-surface redesign (EXC-6) — only drop `MultiEdit`, only swap `TodoWrite`→Task*; do not reorder or add other tools.

## H. Shared-File Sequencing vs SPEC-CCSYNC-CLAUDEMD-001 (CON-5)

`agent-authoring.md` line ~212 (dev-root + template mirror) is a run-phase target
of BOTH SPECs, on the SAME line:

```
Manager agents: Read, Write, Edit, Grep, Glob, Bash, Skill, TodoWrite (NOTE: Agent
tool is NOT included by default for regular subagents. However, Agent Teams
teammates CAN spawn other teammates using Agent() with the team_name parameter,
v2.1.50+)
```

- **Sibling SPEC-CCSYNC-CLAUDEMD-001 (M3.2)** reconciles the parenthetical claim:
  removes "teammates CAN spawn other teammates" + corrects `v2.1.50+`.
- **THIS SPEC (M2.3)** replaces the leading `… Skill, TodoWrite` with
  `… Skill, TaskCreate, TaskUpdate, TaskList, TaskGet`.

Sequencing rule: **SPEC-CCSYNC-CLAUDEMD-001 runs FIRST; this SPEC (CCSYNC-TOOLCAT-001)
rebases onto its agent-authoring.md commit** before making its own edit. This
mirrors the sibling plan.md §H ("this SPEC runs FIRST … the sibling
SPEC-CCSYNC-TOOLCAT-001 rebases"). Running them concurrently would cause an
intra-batch edit collision on the same line in the same file (source + mirror).

The orchestrator MUST confirm at this SPEC's pre-flight check (§C) that the
sibling is committed/merged (not mid-run). When this SPEC's run phase begins, it
re-Greps line ~212 (the sibling may have rewritten the parenthetical and shifted
the line) before editing the `TodoWrite` portion.

## I. Tier & Routing

- Tier M (standard). Estimated scope: agent files (manager-spec, manager-develop,
  manager-docs, manager-git, builder-harness) ×2 copies + settings.json.tmpl +
  agent-authoring.md ×2 + ~4 foundation-cc reference files ×2 + 1 NEW Go test +
  embedded.go regeneration. Content-edit + one test file.
- Hybrid Trunk Tier M → main-direct, no PR (CLAUDE.local.md §23.9), unless user
  escalates with `--pr` (then manager-git PR routing per §23.9).

## J. Cross-References

- spec.md — requirements + exclusions.
- acceptance.md — Given-When-Then + AC grep matrix.
- Sibling: `.moai/specs/SPEC-CCSYNC-CLAUDEMD-001/plan.md` §H (sequencing source-of-truth).
- CLAUDE.local.md §2 (Template-First), §23.9 (Tier routing), §25 (internal-content isolation).
- CLAUDE.md §3 / §14 — Task*-family mandate this SPEC aligns the frontmatter to.
- `internal/template/rule_template_mirror_test.go` + `internal/template/agent_frontmatter_audit_test.go` — mirror + frontmatter invariants the new guard sits beside.

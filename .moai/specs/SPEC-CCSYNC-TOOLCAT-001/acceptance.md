# Acceptance Criteria — SPEC-CCSYNC-TOOLCAT-001

> All acceptance criteria are mechanically checkable. Each AC names a verifiable
> command (grep / build / test) with the expected result. Run-phase exit requires
> every Blocking AC to PASS. Verified against HEAD `5042e309c`; the run phase
> re-confirms line numbers before editing.

## A. Definition of Done

- All 12 requirements (REQ-CCSYNC-T-001..012) satisfied.
- All Blocking ACs (AC-CCSYNC-T-001..017) PASS.
- `make build` clean; `internal/template/embedded.go` regenerated.
- `go test ./internal/template/...` green (mirror-drift + neutrality + internal-content leak + the NEW tool-catalog guard).
- No file outside the in-scope set modified (verbatim official-docs reproductions untouched per EXC-3; read-only evaluators untouched per EXC-1; the teammates-spawn/version claim untouched per EXC-4 — owned by the sibling).
- Sibling SPEC-CCSYNC-CLAUDEMD-001 confirmed committed/merged before this SPEC's agent-authoring.md edit (CON-5).

## B. Given-When-Then Scenarios

### Scenario 1 — H3: MultiEdit removed from agent tool lists (REQ-CCSYNC-T-001)

- **Given** manager-spec and manager-develop declare `MultiEdit` in `tools:` (source + mirror) at HEAD `5042e309c`,
- **When** the run phase removes `MultiEdit` from those `tools:` lines,
- **Then** no agent `tools:` frontmatter line in `.claude/agents/` or `internal/template/templates/.claude/agents/` contains `MultiEdit`.

### Scenario 2 — H3: MultiEdit body instruction replaced (REQ-CCSYNC-T-002/T-004)

- **Given** the manager-spec body says "[HARD] Use MultiEdit for simultaneous 3-file creation" and references "Write/MultiEdit" in its SPEC ID self-check,
- **When** the run phase rewrites the instruction to parallel `Edit`/`Write` calls and corrects the incidental mentions,
- **Then** the manager-spec body (source + mirror) contains no "MultiEdit" string and instructs the agent to make parallel `Edit`/`Write` calls in a single turn.

### Scenario 3 — H3: MultiEdit removed from matchers + settings (REQ-CCSYNC-T-003)

- **Given** manager-develop hook matchers and settings.json.tmpl reference `MultiEdit`,
- **When** the run phase drops `MultiEdit` from the matchers, the settings.json.tmpl matcher, and the permissions array,
- **Then** no `.md` agent file and no `settings.json.tmpl` (source + mirror) contains `MultiEdit`.

### Scenario 4 — H4: TodoWrite migrated to Task* in 5 agents (REQ-CCSYNC-T-005)

- **Given** manager-spec, manager-develop, manager-docs, manager-git, builder-harness declare `TodoWrite` in `tools:` (source + mirror),
- **When** the run phase replaces `TodoWrite` with `TaskCreate, TaskUpdate, TaskList, TaskGet`,
- **Then** none of the 5 agents' `tools:` lines contains `TodoWrite`, and each contains all four Task* tools (source + mirror).

### Scenario 5 — H4: read-only evaluators untouched (REQ-CCSYNC-T-006)

- **Given** plan-auditor and sync-auditor declare neither `TodoWrite` nor the Task* family,
- **When** the run phase migrates the 5 implementation agents,
- **Then** plan-auditor and sync-auditor `tools:` lines (source + mirror) still declare neither `TodoWrite` nor any Task* tool.

### Scenario 6 — H4: agent-authoring.md recommendation updated (REQ-CCSYNC-T-007), sibling-line preserved (EXC-4/CON-5)

- **Given** agent-authoring.md line ~212 recommends `TodoWrite` AND co-locates the teammates-spawn/v2.1.50 claim,
- **When** the sibling SPEC-CCSYNC-CLAUDEMD-001 has FIRST reconciled the teammates-spawn claim, and this SPEC THEN edits only the `TodoWrite` → Task* portion,
- **Then** agent-authoring.md (source + mirror) recommends the Task* family, does not recommend `TodoWrite`, and the teammates-spawn claim is whatever the sibling left (this SPEC did not touch it).

### Scenario 7 — Low-priority: foundation-cc authoring examples scrubbed, official docs preserved (REQ-CCSYNC-T-008 / EXC-3)

- **Given** the MoAI-authored authoring-kit examples teach `MultiEdit`/`TodoWrite` while the verbatim official-docs reproductions also mention them,
- **When** the run phase scrubs ONLY the authoring-kit examples,
- **Then** the authoring-kit example files contain no `MultiEdit` and no bare `TodoWrite`, AND the three official-docs files (`claude-code-iam-official.md`, `claude-code-sub-agents-official.md`, `claude-code-plugins-official.md`) are unchanged (still contain their verbatim mentions).

### Scenario 8 — NEW CI guard test (REQ-CCSYNC-T-009)

- **Given** there is no tool-catalog guard test at HEAD `5042e309c`,
- **When** the run phase adds `internal/template/tool_catalog_audit_test.go` (RED before M1/M2, GREEN after),
- **Then** `go test ./internal/template/... -run TestToolCatalog` passes, the test rejects any `MultiEdit` entry and any retained-agent `TodoWrite` entry, and its allowlist source is documented in a generic code comment.

### Scenario 9 — Mirror parity + build (REQ-CCSYNC-T-010/T-011)

- **Given** every edited agent file, agent-authoring.md, and foundation-cc example has a `templates/` mirror,
- **When** the run phase edits both copies of each in the same commit and runs `make build`,
- **Then** `go test ./internal/template/...` passes (mirror-drift + leak + neutrality + new guard) AND `internal/template/embedded.go` is regenerated.

### Scenario 10 — Template neutrality (REQ-CCSYNC-T-012)

- **Given** the new test file and any reworded template text are written under `internal/template/templates/` / `internal/template/`,
- **When** the run phase authors them generically (no internal SPEC IDs / REQ / dates / SHAs),
- **Then** `go test ./internal/template/... -run TestTemplateNeutralityAudit` and the internal-content leak guard pass.

### Scenario 11 (edge) — sibling sequencing honored (CON-5)

- **Given** SPEC-CCSYNC-CLAUDEMD-001 also targets agent-authoring.md line ~212,
- **When** the orchestrator enters this SPEC's run phase,
- **Then** the pre-flight check confirms the sibling is committed/merged (not mid-run), and this SPEC rebases onto the sibling's agent-authoring.md commit before its own edit.

## C. AC Grep / Verification Matrix (Blocking)

All commands run from repo root. Paths: `A = .claude/agents/moai`,
`TA = internal/template/templates/.claude/agents/moai`,
`S = internal/template/templates/.claude/settings.json.tmpl`,
`R = .claude/rules/moai/development/agent-authoring.md`,
`TR = internal/template/templates/.claude/rules/moai/development/agent-authoring.md`,
`FC = .claude/skills/moai-foundation-cc`,
`TFC = internal/template/templates/.claude/skills/moai-foundation-cc`.

| AC | REQ | Command | Expected |
|----|-----|---------|----------|
| AC-CCSYNC-T-001 | T-001 | `grep -rhE '^tools:' .claude/agents/moai/ internal/template/templates/.claude/agents/moai/ \| grep -c 'MultiEdit'` | `0` (no agent tools line has MultiEdit) |
| AC-CCSYNC-T-002 | T-002/T-004 | `grep -c 'MultiEdit' .claude/agents/moai/manager-spec.md internal/template/templates/.claude/agents/moai/manager-spec.md` | `0` (manager-spec body fully scrubbed, both copies) |
| AC-CCSYNC-T-003 | T-002 | `grep -ci 'parallel.*\(Edit\|Write\)\|Edit/Write\|Write/Edit' .claude/agents/moai/manager-spec.md` | ≥1 (parallel Edit/Write instruction present) |
| AC-CCSYNC-T-004 | T-003 | `grep -c 'MultiEdit' .claude/agents/moai/manager-develop.md internal/template/templates/.claude/agents/moai/manager-develop.md` | `0` (frontmatter + matchers scrubbed, both copies) |
| AC-CCSYNC-T-005 | T-003 | `grep -c 'MultiEdit' internal/template/templates/.claude/settings.json.tmpl` | `0` (matcher + permissions array scrubbed) |
| AC-CCSYNC-T-006 | T-005 | `for f in manager-spec manager-develop manager-docs manager-git builder-harness; do grep '^tools:' .claude/agents/moai/$f.md \| grep -c 'TodoWrite'; done \| awk '{s+=$1} END{print s}'` | `0` (no TodoWrite in any of the 5 source agents) |
| AC-CCSYNC-T-007 | T-005 | `for f in manager-spec manager-develop manager-docs manager-git builder-harness; do grep '^tools:' internal/template/templates/.claude/agents/moai/$f.md \| grep -c 'TodoWrite'; done \| awk '{s+=$1} END{print s}'` | `0` (no TodoWrite in any of the 5 mirror agents) |
| AC-CCSYNC-T-008 | T-005 | `for f in manager-spec manager-develop manager-docs manager-git builder-harness; do grep -c 'TaskCreate.*TaskUpdate.*TaskList.*TaskGet' .claude/agents/moai/$f.md; done \| awk '{s+=$1} END{print s}'` | `5` (all 4 Task* tools present in each of the 5 source agents) |
| AC-CCSYNC-T-009 | T-006 | `grep '^tools:' .claude/agents/moai/plan-auditor.md .claude/agents/moai/sync-auditor.md \| grep -cE 'TodoWrite\|TaskCreate\|TaskUpdate\|TaskList\|TaskGet'` | `0` (read-only evaluators unchanged) |
| AC-CCSYNC-T-010 | T-007 | `grep -n 'TodoWrite' .claude/rules/moai/development/agent-authoring.md internal/template/templates/.claude/rules/moai/development/agent-authoring.md \| grep -c 'Manager agents'` | `0` (manager-agent recommendation no longer names TodoWrite) |
| AC-CCSYNC-T-011 | T-007 | `grep -c 'TaskCreate' .claude/rules/moai/development/agent-authoring.md internal/template/templates/.claude/rules/moai/development/agent-authoring.md` | `2` (Task* family recommended in both copies) |
| AC-CCSYNC-T-012 | T-008 | `grep -rc 'MultiEdit' .claude/skills/moai-foundation-cc/reference/skill-formatting-guide.md .claude/skills/moai-foundation-cc/reference/sub-agents/sub-agent-examples.md .claude/skills/moai-foundation-cc/reference/sub-agents/sub-agent-formatting-guide.md .claude/skills/moai-foundation-cc/reference/sub-agents/sub-agent-integration-patterns.md \| awk -F: '{s+=$2} END{print s}'` | `0` (authoring-kit examples scrubbed, source) |
| AC-CCSYNC-T-013 | T-008 | `grep -c 'MultiEdit' .claude/skills/moai-foundation-cc/reference/claude-code-iam-official.md .claude/skills/moai-foundation-cc/reference/claude-code-sub-agents-official.md \| awk -F: '{s+=$2} END{print s}'` | ≥3 (verbatim official-docs PRESERVED — EXC-3) |
| AC-CCSYNC-T-014 | T-009 | `go test ./internal/template/... -run TestToolCatalog` | PASS (exit 0 — new guard test present + green) |
| AC-CCSYNC-T-015 | T-012 | `go test ./internal/template/... -run TestTemplateNeutralityAudit` | PASS (exit 0) |
| AC-CCSYNC-T-016 | T-010/T-011 | `go test ./internal/template/...` | PASS (exit 0 — mirror-drift + leak + neutrality + new guard all green) |
| AC-CCSYNC-T-017 | T-010 | `make build && git diff --quiet internal/template/embedded.go; echo $?` | `0` (after `make build`, embedded.go matches the regenerated source — no residual diff) |

Notes:
- AC-CCSYNC-T-008 / AC-CCSYNC-T-011 expect exact counts (`5`, `2`). If the run phase legitimately formats the Task* family differently (e.g., reorders), it MUST re-derive the expected count and document the deviation; the binding intent is: all four Task* tools present per migrated agent, zero `TodoWrite`.
- AC-CCSYNC-T-013 is a POSITIVE preservation check — it asserts the verbatim official-docs reproductions STILL contain `MultiEdit` (proving EXC-3 was honored, not over-scrubbed). The exact count may vary; ≥3 confirms the official-docs were not gutted.
- AC-CCSYNC-T-014 uses the `TestToolCatalog` test-name prefix; the run phase sets the final function name(s) (e.g., `TestToolCatalogNoMultiEdit`, `TestToolCatalogNoTodoWrite`) and the grep prefix must match.
- AC-CCSYNC-T-010 keys on the "Manager agents" recommendation line specifically; incidental `TodoWrite` mentions elsewhere in agent-authoring.md (if any historical example) are not the target — the binding criterion is the manager-agent tool-list recommendation.

## D. Quality Gate Criteria (must-pass)

- **Functionality**: all 17 Blocking ACs PASS.
- **Catalog correctness**: zero `MultiEdit` in agent `tools:`/matchers/settings; zero `TodoWrite` in the 5 migrated agents; Task* family present in all 5 (AC-CCSYNC-T-001..008).
- **Scope discipline**: read-only evaluators unchanged (AC-CCSYNC-T-009); verbatim official-docs preserved (AC-CCSYNC-T-013); the teammates-spawn/version claim left to the sibling (EXC-4 — verified by `git diff` not touching that clause).
- **Consistency / mirror parity**: dev-root and template copies agree (mirror-drift green); `git diff --name-only` shows matched source+mirror pairs (AC-CCSYNC-T-016).
- **Neutrality**: template + new test carry no internal SPEC IDs / REQ / AC / dates / SHAs (AC-CCSYNC-T-015 + leak guard inside AC-CCSYNC-T-016).
- **Regression guard**: the NEW tool-catalog test passes and would fail if `MultiEdit` or a bare `TodoWrite` were reintroduced (AC-CCSYNC-T-014).
- **Build integrity**: `make build` clean + embedded.go regenerated (AC-CCSYNC-T-017).

## E. Out-of-Scope Verification (negative ACs)

| AC | Check | Expected |
|----|-------|----------|
| AC-CCSYNC-T-N1 | `git diff --name-only \| grep -cE 'plan-auditor.md\|sync-auditor.md'` | `0` (EXC-1 honored — read-only evaluators untouched) |
| AC-CCSYNC-T-N2 | `git diff --name-only \| grep -cE 'claude-code-iam-official.md\|claude-code-sub-agents-official.md\|claude-code-plugins-official.md'` | `0` (EXC-3 honored — verbatim official-docs untouched) |
| AC-CCSYNC-T-N3 | `git diff .claude/rules/moai/development/agent-authoring.md \| grep -c 'teammates CAN spawn other teammates'` | `0` in the ADDED lines (EXC-4 — this SPEC did not author/alter that clause; the sibling owns it) |
| AC-CCSYNC-T-N4 | `git diff --name-only \| grep -c 'CLAUDE.md'` | `0` (EXC-5 honored — CLAUDE.md not touched) |

# Acceptance Criteria — SPEC-V3R6-RULES-CATALOG-SCRUB-001

All grep commands run from the project root `/Users/goos/MoAI/moai-adk-go`. Each per-defect AC is satisfied when the archived name is **absent at the live-reference site** AND the canonical replacement is present. The two PRESERVE reference files and role_profile contexts are explicitly excluded from "absence" expectations.

## §A Definition of Done

- [ ] All HARD ACs (AC-RCS-001..009, 016..019, 021) PASS
- [ ] AC-RCS-010 (D14 SHOULD) PASS or has a documented blocker re-scoping decision
- [ ] `make build` succeeded AND `go build ./...` exits 0 (template tree embedded via `//go:embed all:templates`; there is NO `embedded.go` regen step)
- [ ] `go test ./internal/template/... -run TestTemplateNeutralityAudit` PASS
- [ ] No archived agent name survives as a live spawn/example/hook-action anywhere except the 4 PRESERVE files + role_profile contexts
- [ ] status: draft → (run-phase sets in-progress)

## §B Given-When-Then Scenarios

### Scenario 1 — P0 CI-protocol no longer spawns an archived agent
- **Given** `ci-autofix-protocol.md` and `ci-watch-protocol.md` previously named `manager-quality` as a live diagnostic subagent,
- **When** the scrub (M1) replaces those references with the Stop hook `sync-phase-quality-gate.sh` / `Agent(general-purpose)` diagnostic,
- **Then** `grep -rn 'manager-quality' .claude/rules/moai/workflow/ci-autofix-protocol.md .claude/rules/moai/workflow/ci-watch-protocol.md` returns no live-spawn line, and the same is true in the template mirror.

### Scenario 2 — agent-authoring ships the 8-agent catalog
- **Given** `agent-authoring.md` §Agent Categories listed the OLD 17-agent catalog (Manager Agents (7) + Expert Agents (6)),
- **When** the scrub (M2) rewrites it to 7 retained + `Explore`,
- **Then** the section names no archived agent and is consistent with the file's own 8-agent ceiling section.

### Scenario 3 — PRESERVE files are untouched
- **Given** `archived-agent-rejection.md`, `NOTICE.md`, `agent-patterns.md`, `spec-frontmatter-schema.md` intentionally enumerate archived names,
- **When** the full scrub completes,
- **Then** those four files still contain their archived-agent enumerations verbatim (git diff shows no change to their archived-name lines).

### Scenario 4 — role_profiles survive the scrub
- **Given** `researcher`/`reviewer`/`analyst` appear as `role_profile` values,
- **When** the scrub completes,
- **Then** every `role_profile` token is unchanged (no role_profile was scrubbed mistaking it for an archived agent).

### Scenario 5 — design-pipeline expert-frontend is no longer silently load-bearing
- **Given** `design/constitution.md` names `expert-frontend` as a live pipeline agent in 10 places,
- **When** the carve-out note (M7) is added,
- **Then** the file cross-references the migration table so the archived name denotes a documented ROLE resolving to `Agent(general-purpose)`+frontend-whitelist, not a live spawn target.

## §C Per-defect Grep-Verifiable AC

> Each AC checks BOTH `.claude/rules/moai/...` and its `internal/template/templates/.claude/rules/moai/...` mirror (D-mirror parity is AC-RCS-018).

### AC-RCS-001 (D4) — ci-autofix-protocol manager-quality removed
```bash
grep -n 'manager-quality' .claude/rules/moai/workflow/ci-autofix-protocol.md
# Expected: no line presenting manager-quality as a live subagent/spawn target.
# Replacement present: grep -n 'sync-phase-quality-gate.sh\|general-purpose' .claude/rules/moai/workflow/ci-autofix-protocol.md  → ≥1
```

### AC-RCS-002 (D5) — ci-watch-protocol manager-quality removed
```bash
grep -n 'manager-quality' .claude/rules/moai/workflow/ci-watch-protocol.md
# Expected: no live Wave-3 handoff to manager-quality.
```

### AC-RCS-003 (D10) — agent-authoring 8-agent catalog
```bash
grep -nE 'expert-(backend|frontend|security|devops|performance|refactoring)|manager-(strategy|quality|brain|project)' .claude/rules/moai/development/agent-authoring.md
# Expected: no archived name in §Agent Categories.
grep -c 'Explore' .claude/rules/moai/development/agent-authoring.md  # ≥1 (built-in present)
```

### AC-RCS-004 (D1) — agent-hooks archived rows removed/marked
```bash
grep -nE 'expert-backend|expert-frontend|expert-devops|manager-quality' .claude/rules/moai/core/agent-hooks.md
# Expected: no archived row presented as a live hook action (rows removed OR table marked "archived — non-functional").
```

### AC-RCS-005 (D2) — agent-common-protocol example scrub
Note: the live L401 text is `` `manager-quality` in diagnostic `` — a backtick
sits BETWEEN the tokens, so `grep 'manager-quality in diagnostic'` returns
exit 1 against the UNEDITED file (vacuous before-state). The grep MUST match
the actual live text (the backtick-and-space form) to correctly detect the
read-only-exemption mention that needs scrubbing.

```bash
# (a) Spawn-example archived glob (L~345) removed.
grep -n 'expert-\*' .claude/rules/moai/core/agent-common-protocol.md   # Expected: 0

# (b) Read-only-exemption mention (L~401) scrubbed → Explore. Match the live
#     backtick form: `manager-quality` followed by " in diagnostic".
grep -nE '`manager-quality`[^`]*in diagnostic' \
  .claude/rules/moai/core/agent-common-protocol.md  # Expected: 0 after fix
grep -n 'Explore' .claude/rules/moai/core/agent-common-protocol.md  # ≥1 (replacement present)
# L37 prose "phantom manager-quality / expert-security" MAY remain (documentary
# — it describes the consolidation, not a live spawn/exemption).
```

### AC-RCS-006 (D6) — agent-teams-pattern dead path removed
```bash
grep -n 'manager-strategy' .claude/rules/moai/workflow/agent-teams-pattern.md
# Expected: 0 (dead frontmatter paths: token removed).
```

### AC-RCS-007 (D7) — spec-workflow phantom team agents rewritten
```bash
grep -nE 'team-validator|team-tester|backend-dev|frontend-dev' .claude/rules/moai/workflow/spec-workflow.md
# Expected: 0; replaced by role_profiles (implementer/tester/reviewer) spawned general-purpose.
```

### AC-RCS-008 (D8) — worktree-integration recount + archived names removed
```bash
grep -nE 'expert-backend|expert-frontend|expert-refactoring' .claude/rules/moai/workflow/worktree-integration.md
# Expected: 0 archived names in the isolation-requiring list (L167) and the count comment (L253).
# role_profile lines (researcher/analyst/reviewer as role_profiles) remain — NOT scrubbed.
```

### AC-RCS-009 (D9) — worktree-state-guard escalation target
```bash
grep -n 'claude-code-guide' .claude/rules/moai/workflow/worktree-state-guard.md
# Expected: 0 MoAI-custom claude-code-guide; replaced with Explore.
grep -n 'Explore' .claude/rules/moai/workflow/worktree-state-guard.md  # ≥1
```

### AC-RCS-010 (D11) — orchestrator-templates valid subagent_type
```bash
grep -nE 'subagent_type: *"(analyzer|designer|implementer|reviewer)"' .claude/rules/moai/development/orchestrator-templates.md
# Expected: 0 invalid subagent_type; rewritten to "general-purpose" + role profile.
```

### AC-RCS-011 (D12) — model-policy tiers scrubbed
The bare English words "quality"/"strategy" survive in non-archived contexts
(e.g. the "Maximum quality" description cell at model-policy.md:97). Therefore
the AC MUST assert the archived agent-NAME is gone from the tier ROW, NOT that
the word never appears. Scope the grep to archived-agent tokens as whole
comma-delimited or backtick-delimited table cells, not bare words.

```bash
# (a) The Model Policy Tiers table MUST NOT list archived agents as agent-LIST
#     cell members. Archived tokens in the Opus/Sonnet/Haiku columns are bounded
#     by agent-list delimiters: a "| " or ", " on the LEFT and a "," or " |" on
#     the RIGHT. The "Maximum quality" DESCRIPTION cell is NOT matched, because
#     its "quality" is preceded by "Maximum " (not "| " or ", "). VERIFIED
#     satisfiable: this pattern returns exit 0 on the pre-fix table (3 rows) and
#     exit 1 (zero matches) on the post-fix table even with "Maximum quality"
#     retained.
grep -nE '(\| |, )(manager-)?(strategy|quality|researcher)(,| \|)' \
  .claude/rules/moai/development/model-policy.md
# Expected after fix: 0 matches (exit 1). The "Maximum quality" prose survives
# without matching. The table either lists only the 8-agent catalog OR carries a
# note that tiers are role-profile-based.

# (b) Positive confirmation the table was actually updated (not just emptied):
grep -nE 'role-profile|role profile|manager-spec|plan-auditor|sync-auditor|builder-harness' \
  .claude/rules/moai/development/model-policy.md  # ≥1
```

### AC-RCS-012 (D13) — language boilerplate consistency (16-file parity)
```bash
for f in cpp csharp elixir flutter javascript r ruby typescript go python rust java kotlin scala php swift; do
  c=$(grep -c 'manager-quality' .claude/rules/moai/languages/$f.md 2>/dev/null)
  echo "$f.md: $c"
done
# Expected: ALL 16 report 0 (the 8 that had the boilerplate line now match the 8 that omit it).
```

### AC-RCS-013 (D14) — design/constitution carve-out present
```bash
grep -n 'archived-agent-rejection' .claude/rules/moai/design/constitution.md
# Expected: ≥1 (carve-out note cross-referencing the migration table for pipeline expert-frontend).
```

### AC-RCS-014 (D3) — zone-registry CONST cross-reference
```bash
grep -n 'archived-agent-rejection\|general-purpose' .claude/rules/moai/core/zone-registry.md
# Expected: ≥1 near the expert-frontend CONST entry (carve-out OR retained rename).
```

### AC-RCS-015 (D15) — skill-authoring frontmatter example scrub
```bash
grep -n 'expert-backend' .claude/rules/moai/development/skill-authoring.md
# Expected: 0; agents: ["expert-backend"] examples replaced with a retained agent.
```

## §D Global Invariant AC

### AC-RCS-016 — PRESERVE files intact (REQ-RCS-002)
```bash
# These MUST still contain archived-agent enumerations (unchanged):
grep -lE 'manager-strategy|expert-backend|manager-quality' \
  .claude/rules/moai/workflow/archived-agent-rejection.md \
  .claude/rules/moai/NOTICE.md \
  .claude/rules/moai/development/agent-patterns.md \
  .claude/rules/moai/development/spec-frontmatter-schema.md
# Expected: all 4 files listed (enumerations preserved). git diff on these 4 files: no archived-name lines changed.
```

### AC-RCS-017 — role_profiles unchanged (REQ-RCS-003)
```bash
# role_profile tokens are valid; confirm they survive.
grep -rn 'role_profile' .claude/rules/moai/workflow/worktree-integration.md .claude/rules/moai/core/zone-registry.md
# Expected: researcher/analyst/reviewer/implementer/tester role_profile lines present & unchanged.
```

### AC-RCS-018 — template mirror parity (REQ-RCS-008) — per-file diff by enrollment split
[ENROLLMENT-CRITICAL] `TestRuleTemplateMirrorDrift` enrolls ONLY a fixed
allowlist (hooks-system + `spec-workflow.md` + session-handoff + 2 evaluator
profiles). Among THIS SPEC's 22 targets, only `spec-workflow.md` is enrolled.
For the other ~17 byte-identical-but-NOT-enrolled files, "the Go test PASSed" is
**vacuous** — a missed mirror edit passes silently because the test never
inspects those files. Therefore the parity AC MUST use EXPLICIT per-file
assertions, split by enrollment + §25-sanitization status into 3 groups.

Group classification of this SPEC's 22 targets (deterministically verified by `diff -q` against HEAD):
- **deployed↔template divergent at HEAD (Group 3 — NEVER `diff -q`)**: `agent-common-protocol.md` (genuinely §25-sanitized, 17 tokens) AND `zone-registry.md` (§25-divergent — its deployed copy carries a `CONST-V3R6-001` block, count 2, that the template copy has §25-stripped to count 0). PLUS two files whose §25-divergence has historically existed but is byte-identical at HEAD today: `ci-watch-protocol.md`, `agent-teams-pattern.md` (currently byte-identical; historically §25-leak-bearing). Group 3 is safe for all four — for the byte-identical pair the change-confinement diff simply shows zero extra hunks.
- **enrolled byte-identical (covered by the Go test)**: `spec-workflow.md`.
- **non-enrolled byte-identical (need explicit `diff -q`)**: the remaining **17** files (agent-hooks, ci-autofix-protocol, worktree-integration, worktree-state-guard, agent-authoring, orchestrator-templates, model-policy, skill-authoring, design/constitution, + 8 language files). NOTE: `zone-registry.md` was REMOVED from this group — a Group-1 `diff -q` FALSE-FAILS on it TODAY (before any scrub) because of the pre-existing §25 `CONST-V3R6-001` divergence, and a run-phase implementer obeying "PARITY-OK for all" could be driven into a §25 neutrality violation. It is now Group 3.

```bash
# Group 1 — non-enrolled byte-identical files: EXPLICIT per-file diff -q.
#   (These are NOT §25-divergent, so deployed and template MUST be byte-identical.)
#   NOTE: core/zone-registry.md is INTENTIONALLY ABSENT here — it is §25-divergent
#   (CONST-V3R6-001 deployed-only block) and lives in Group 3, NOT here. A diff -q
#   on zone-registry FALSE-FAILS today before any scrub.
NONENROLLED="core/agent-hooks.md \
  workflow/ci-autofix-protocol.md workflow/worktree-integration.md \
  workflow/worktree-state-guard.md development/agent-authoring.md \
  development/orchestrator-templates.md development/model-policy.md \
  development/skill-authoring.md design/constitution.md \
  languages/cpp.md languages/csharp.md languages/elixir.md languages/flutter.md \
  languages/javascript.md languages/r.md languages/ruby.md languages/typescript.md"
for f in $NONENROLLED; do
  diff -q ".claude/rules/moai/$f" "internal/template/templates/.claude/rules/moai/$f" \
    && echo "PARITY-OK: $f" || echo "PARITY-FAIL: $f"
done
# Expected: PARITY-OK for all 17 — explicit diff catches a missed mirror edit
# that the Go test would silently pass (non-enrolled = vacuous).

# Group 2 — enrolled byte-identical file: the Go test is authoritative here.
go test ./internal/template/... -run TestRuleTemplateMirrorDrift
# Expected: PASS — spec-workflow.md (enrolled) is byte-identical deployed↔template.

# Group 3 — §25-divergent (or historically-§25) files: do NOT use diff -q for parity.
#   These four are EXCLUDED from byte-parity: change-confinement diff instead.
#   - agent-common-protocol.md: genuinely §25-sanitized at HEAD (17 tokens divergent).
#   - zone-registry.md: §25-divergent at HEAD (deployed-only CONST-V3R6-001 block).
#   - ci-watch-protocol.md, agent-teams-pattern.md: currently byte-identical at HEAD
#     (historically §25-leak-bearing); their change-confinement diff shows zero
#     extra hunks today, which is the expected SAFE state.
#   Verify (a) no internal-content leak, AND (b) ONLY the archived-name edit
#   changed — no NEW divergence introduced beyond any pre-existing §25 delta.
go test ./internal/template/... -run TestTemplateNoInternalContentLeak
# Expected: PASS — template mirrors carry no internal-content leak.
for f in core/agent-common-protocol.md core/zone-registry.md workflow/ci-watch-protocol.md workflow/agent-teams-pattern.md; do
  echo "=== change-confinement: $f (deployed vs template diff should show ONLY the archived-name edit + any pre-existing §25 delta) ==="
  diff ".claude/rules/moai/$f" "internal/template/templates/.claude/rules/moai/$f"
done
# Expected: the diff hunks correspond ONLY to (i) the archived-name scrub this
# SPEC applied to both copies, and (ii) any pre-existing §25 sanitization delta
# (the CONST-V3R6-001 block for zone-registry; the 17-token sanitization for
# agent-common-protocol; ZERO extra hunks for the byte-identical pair) — NO new
# unexplained divergence. (Manual review of the diff hunks.)
```

Note: if a Group-1 file later becomes enrolled in `workflowOptMirroredPaths` (or
a sibling allowlist) by another SPEC, it MAY move to Group 2. Do NOT add a
§25-sanitized (Group-3) file to the byte-parity allowlist — it would fail by
design. The explicit Group-1 `diff -q` is the load-bearing assertion; the Go
test alone is insufficient for non-enrolled files.

### AC-RCS-019 — template neutrality (REQ-RCS-009)
```bash
go test ./internal/template/... -run TestTemplateNeutralityAudit
# Expected: PASS — no SPEC ID/date/SHA/REQ-token/memory-path leaked into template content.
```

### AC-RCS-020 — make build + go build succeed with embedded template edits
```bash
# The template tree is embedded at compile time via //go:embed all:templates
# (internal/template/embed.go:28). There is NO embedded.go file to regenerate.
# make build = templ-generate + gen-catalog-hashes.go --all + go build (Makefile:23-25).
make build
go build ./...
echo "exit=$?"
# Expected: make build succeeds AND go build ./... exits 0 — the edited template
# source under internal/template/templates/.claude/rules/moai/** is picked up by
# the go:embed directive directly at compile time.
```

### AC-RCS-021 — catalog-wide residual scan (REQ-RCS-001)
```bash
# After scrub: archived names survive ONLY in the 4 PRESERVE files + role_profile contexts.
grep -rlE 'manager-strategy|manager-quality|manager-brain|manager-project|expert-backend|expert-frontend|expert-security|expert-devops|expert-performance|expert-refactoring' .claude/rules/ \
  | grep -vE 'archived-agent-rejection.md|NOTICE.md|agent-patterns.md|spec-frontmatter-schema.md'
# Expected: each remaining file (if any) contains ONLY role_profile mentions or a documented carve-out cross-reference — manually confirm each line is a role_profile or a migration-table cross-reference, never a live spawn/example/hook-action.
```

## §E Severity & Closure Gates

| AC | Severity | Gate |
|----|----------|------|
| AC-RCS-001, 002, 003 | MUST-FIX (P0) | Block run-phase completion |
| AC-RCS-004..012, 015 | MUST-FIX | Block sync |
| AC-RCS-013, 014 (D14/D3) | SHOULD-FIX | Carve-out OR documented blocker re-scope |
| AC-RCS-016, 017 | MUST-FIX (regression guard) | Over-scrub prevention |
| AC-RCS-018, 019, 020 | MUST-FIX | Template-First + neutrality |
| AC-RCS-021 | MUST-FIX | Catalog-wide closure |

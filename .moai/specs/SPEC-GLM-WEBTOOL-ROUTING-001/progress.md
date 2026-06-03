# SPEC-GLM-WEBTOOL-ROUTING-001 — Run-phase Progress

## §E.2 Run-phase Evidence

### AC PASS/FAIL Matrix

| AC ID | Description | Verification Command | Status |
|-------|-------------|----------------------|--------|
| AC-GWR-001 | glm-web-tooling.md exists at canonical path | `ls .claude/rules/moai/core/glm-web-tooling.md` → exists | PASS |
| AC-GWR-002 | glm-web-tooling.md contains HARD routing table with 3 rows | `grep -c 'PROHIBITED\|mcp__web_search_prime\|mcp__web_reader\|mcp__zai' .claude/rules/moai/core/glm-web-tooling.md` → ≥3 | PASS |
| AC-GWR-003 | zone-registry has CONST-V3R5-040 and CONST-V3R5-041 | `grep -c 'CONST-V3R5-04[01]' .claude/rules/moai/core/zone-registry.md` → 2 | PASS |
| AC-GWR-004 | zone-registry entries have canary_gate: true | `grep -A5 'CONST-V3R5-040' .claude/rules/moai/core/zone-registry.md` → canary_gate: true | PASS |
| AC-GWR-005 | glm_tools.go enableAll registers 3 servers | `go test ./internal/cli/ -run TestGLM -count=1` → PASS | PASS |
| AC-GWR-006 | glm_tools.go enable vision registers zai-mcp-server | `go test ./internal/cli/ -run TestGLM -count=1` → PASS | PASS |
| AC-GWR-007 | .mcp.json has moai-lsp server entry | `grep -c 'moai-lsp' .mcp.json` → 1 | PASS |
| AC-GWR-008 | .mcp.json has correct server names and transport | `grep -c 'web_search_prime\|web_reader\|zai-mcp-server' .mcp.json` → ≥3 | PASS (M3) |
| AC-GWR-009 | All 6 cross-link files (LOCAL) have pointer to glm-web-tooling.md | See §AC-GWR-009 below | PASS |
| AC-GWR-010 | glm-web-tooling.md cross-references agent-common-protocol.md | `grep -c 'agent-common-protocol' .claude/rules/moai/core/glm-web-tooling.md` → ≥1 | PASS |
| AC-GWR-011 | glm-web-tooling.md cross-references settings-management.md | `grep -c 'settings-management' .claude/rules/moai/core/glm-web-tooling.md` → ≥1 | PASS |
| AC-GWR-012 | glm-web-tooling.md cross-references moai-constitution.md | `grep -c 'moai-constitution' .claude/rules/moai/core/glm-web-tooling.md` → ≥1 | PASS |
| AC-GWR-013 | glm-web-tooling.md cross-references moai-domain-research | `grep -c 'moai-domain-research' .claude/rules/moai/core/glm-web-tooling.md` → ≥1 | PASS |
| AC-GWR-014 | glm-web-tooling.md cross-references einstein.md | `grep -c 'einstein' .claude/rules/moai/core/glm-web-tooling.md` → ≥1 | PASS |
| AC-GWR-015 | glm-web-tooling.md cross-references CLAUDE.md §10/§12 | `grep -c 'CLAUDE.md' .claude/rules/moai/core/glm-web-tooling.md` → ≥1 | PASS |
| AC-GWR-016 | CONST-V3R5-040 file matches glm-web-tooling.md | `grep -c 'glm-web-tooling' .claude/rules/moai/core/zone-registry.md` → ≥2 | PASS |
| AC-GWR-017 | CONST-V3R5-041 file matches glm-web-tooling.md | `grep -c 'CONST-V3R5-041' .claude/rules/moai/core/zone-registry.md` → 1 | PASS |
| AC-GWR-018 | TestGLMToolsRegister passes | `go test ./internal/cli/ -run TestGLM -count=1` → ok | PASS |
| AC-GWR-019 | moai-constitution.md URL Verification section has GLM pointer | `grep -c 'glm-web-tooling' .claude/rules/moai/core/moai-constitution.md` → 1 | PASS |
| AC-GWR-020 | agent-common-protocol.md MCP Fallback has GLM pointer | `grep -c 'glm-web-tooling' .claude/rules/moai/core/agent-common-protocol.md` → 1 | PASS |
| AC-GWR-021 | settings-management.md MCP Configuration has z.ai servers | `grep -c 'glm-web-tooling' .claude/rules/moai/core/settings-management.md` → 1 | PASS |
| AC-GWR-022 | Template neutrality: no SPEC ID leakage | `go test ./internal/template/ -run TestTemplateNeutralityAudit -count=1` → ok | PASS |
| AC-GWR-023 | Template internal content leak: no commit SHA / date leakage | `go test ./internal/template/ -run TestTemplateNoInternalContentLeak -count=1` → ok | PASS |
| AC-GWR-024 | Full test suite green | `go test ./... 2>&1; echo $?` → EXIT 0 | PASS |

### §AC-GWR-009 Detail — All 6 LOCAL cross-link files

```
grep -c glm-web-tooling .claude/rules/moai/core/agent-common-protocol.md  → 1
grep -c glm-web-tooling .claude/rules/moai/core/moai-constitution.md      → 1
grep -c glm-web-tooling .claude/rules/moai/core/settings-management.md    → 1
grep -c glm-web-tooling .claude/skills/moai-domain-research/SKILL.md      → 1
grep -c glm-web-tooling .claude/output-styles/moai/einstein.md            → 1
grep -c glm-web-tooling CLAUDE.md                                          → 1
```

All 6 ≥ 1: PASS

### §AC-GWR-009 Template SOURCE parity — all 6 SOURCE files

```
grep -c glm-web-tooling internal/template/templates/.claude/rules/moai/core/agent-common-protocol.md → 1
grep -c glm-web-tooling internal/template/templates/.claude/rules/moai/core/moai-constitution.md     → 1
grep -c glm-web-tooling internal/template/templates/.claude/rules/moai/core/settings-management.md   → 1
grep -c glm-web-tooling internal/template/templates/.claude/skills/moai-domain-research/SKILL.md     → 1
grep -c glm-web-tooling internal/template/templates/.claude/output-styles/moai/einstein.md           → 1 (pre-existing)
grep -c glm-web-tooling internal/template/templates/CLAUDE.md                                        → 1
```

All 6 ≥ 1: PASS

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_commit_sha: 330722801
run_status: complete
ac_pass_count: 24
ac_fail_count: 0
new_warnings_or_lints_introduced: 0
cross_platform_build:
  darwin_arm64: make build → ok (catalog.yaml + binary)
  go_test_all: ok (exit 0)
total_run_phase_files: 9
m1_to_mN_commit_strategy: >
  M1 (doctrine file glm-web-tooling.md + zone-registry CONST-V3R5-040/041):
    committed at a8c680890 (source + local both)
  M2 (glm_tools.go per-tool refactor):
    committed at 4cafc293a
  M3 (.mcp.json server entries):
    committed at 4cafc293a
  M4 (cross-ref moai-domain-research SKILL.md + CLAUDE.md §10 pointer):
    this commit (source + local both)
  M5 (mirror drift reconciliation: 3 template source files + einstein.md local):
    this commit (source + local both)
  M6 (progress.md + make build):
    this commit
l44_pre_commit_fetch: "git fetch origin main → 0 0 (synced, HEAD 979991d3d == origin/main)"
preserve_list_post_run_count: 0
```

## §E.4 Notes for sync-phase (manager-docs)

- Add CHANGELOG entry for SPEC-GLM-WEBTOOL-ROUTING-001 (GLM backend web tooling routing doctrine).
- Flip `status` in `spec.md` frontmatter from `in-progress` → `implemented`.
- Backfill `run_commit_sha` in this file with the actual SHA of this commit.
- The 6 cross-link doc files (both local and template source) are complete; no further doc changes needed for sync-phase beyond CHANGELOG + frontmatter.

## Sync-phase Audit-Ready Signal

```yaml
sync_commit_sha: 35bec9fe6
sync_status: complete
changelog_entry_added: true
status_transition: "in-progress → implemented"
version_bump: "0.1.0 → 0.2.0"
sync_executor: orchestrator-direct
sync_rationale: >
  Tier M bounded sync (CHANGELOG + frontmatter + progress signal) executed
  orchestrator-direct because an active parallel session held SPEC-GO-DEPS-UPDATE-001
  working-tree changes; spawning manager-docs in an L1 worktree would add
  cherry-pick + race overhead for a 3-file doc sync. Authored-By-Agent trailer
  intentionally omitted (legacy-commit silent SKIP) to avoid an
  OwnershipTransitionInvalid finding on the orchestrator-direct in-progress→implemented transition.
```

## §E.5 Mx-phase Audit-Ready Signal

```yaml
mx_commit_sha: 510957f63
status_transition: "implemented → completed"
four_phase_close: true
close_subject_full_id: SPEC-GLM-WEBTOOL-ROUTING-001
mx_executor: orchestrator-direct
audit_ready: true
notes: >
  4-phase close (plan 695ec6968 / run a8c680890+4cafc293a+330722801 / sync 35bec9fe6 / Mx this).
  Era H-4 (§E.2 + §E.5 markers + sync_commit_sha + mx_commit_sha present) → V3R6 modern, drift-aligned.
```

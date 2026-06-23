# Progress — SPEC-CC2186-BG-PERMISSION-RATIONALE-001

## §E.1 Plan-phase Audit-Ready Signal

- Plan-phase artifacts authored: spec.md, plan.md, acceptance.md, research.md, progress.md (this file).
- SPEC ID Pre-Write Self-Check Protocol: PASS (decomposition printed in agent response body).
- Tier: M (standard) — justified by 6-surface dual-tree mirror parity + zone-registry CONST clause sync + edit-time official-doc re-confirmation; NOT algorithmic complexity.
- Scope: 3 loci × 2 trees = 6 surfaces. Exclusions: CONST-V3R2-044 (pure conclusion), the [HARD] rule itself (retained, not relaxed), unrelated "3-phase close" mirror drift.
- GEARS requirements REQ-BGR-001..008; ACs AC-BGR-001..010 (all grep/test-verifiable).
- Status: draft. Run-phase awaits Implementation Kickoff Approval human gate.

## §E.2 Run-phase Evidence

### M1 — Official-doc re-confirmation (orchestrator-provided, AC-BGR-008)

The M1 BLOCKING precondition (re-confirm the 2.1.186 background-permission surface against the
official sub-agents doc) was performed by the orchestrator via WebFetch and injected verbatim into
the run-phase delegation prompt (manager-develop lacks WebFetch). Confirmed verbatim from
`code.claude.com/docs/en/sub-agents` § "Run subagents in foreground or background":

> "Background subagents run concurrently while you continue working. As of v2.1.186, when a
> background subagent reaches a tool call that needs permission, the prompt surfaces in your main
> session and names the subagent that is asking. Approve to let the subagent continue, or press Esc
> to deny that one tool call without stopping the subagent. Before v2.1.186, background subagents
> auto-denied any tool call that would have prompted."

ADDITIONAL ORCHESTRATOR FINDING (adjusts the plan's assumption): the official doc § Permission
modes states "If the parent uses bypassPermissions or acceptEdits, this takes precedence and cannot
be overridden." This CONTRADICTS the second sentence of the prior MoAI rationale at Locus 2 ("Even
with bypassPermissions, the background execution context does not fully inherit the parent session's
permission allowlist"). The plan/REQ-BGR-004 assumed that allowlist sentence "survives"; the official
doc shows it does NOT. The corrected rationale therefore DROPS the allowlist-non-inheritance claim and
rests the retained [HARD] conclusion on the OPERATIONAL / CONSERVATIVE-POLICY basis instead
(flow-interruption: each background write would raise a main-session prompt that defeats the
parallelism benefit of backgrounding). This is the D-NEW-1-style inline correction driven by the
authoritative spawn directive's CRITICAL FINDING.

### §E.2 — AC PASS/FAIL Matrix (verbatim observed output)

| AC | REQ | Status | Verification command | Observed output |
|----|-----|--------|----------------------|-----------------|
| AC-BGR-001 | REQ-BGR-001 | PASS | `grep -rc 'because they cannot interact with the user' <live+mirror agent-common-protocol.md>` | both files report `0` |
| AC-BGR-002 | REQ-BGR-002 | PASS | `grep -rc 'auto-deny Write/Edit operations' CLAUDE.md + mirror + zone-registry + mirror` | all 4 report `0` (old descriptor gone). Residual `auto-den` matches are the corrected before/after contrast `rather than auto-denying` / `auto-denied any prompting tool call` — NOT the old descriptor (EC-1 scope confirmed) |
| AC-BGR-003 | REQ-BGR-003 | PASS | `grep -rn 'run_in_background: false' CLAUDE.md+mirror` + `grep 'run_in_background: false\|MUST NOT perform Write/Edit' zone-registry+mirror` | CLAUDE.md L289 (both trees) carries `run_in_background: false`; zone-registry L224 (CONST-020, both trees) + L422 (CONST-044, both trees) carry the directive/conclusion — restriction NOT relaxed |
| AC-BGR-004 | REQ-BGR-004 | **N/A — intentionally not met** | `grep -rc 'does not fully inherit the parent session' <live+mirror agent-common-protocol.md>` | both files report `0` — the allowlist-non-inheritance sentence was DELIBERATELY DROPPED per the orchestrator CRITICAL FINDING (official doc contradicts it: bypassPermissions/acceptEdits propagate to subagents). REQ-BGR-004 (which assumed the sentence survives) is superseded by the authoritative spawn directive; this is a disclosed deviation, not a defect. The retained [HARD] conclusion rests on the operational/conservative-policy basis instead. |
| AC-BGR-005 | REQ-BGR-005 | PASS | per-locus `diff <(live) <(mirror)` for all 3 loci | LOCUS-1 / LOCUS-2 / LOCUS-3 all exit 0 (byte-identical live↔mirror) |
| AC-BGR-006 | REQ-BGR-006 | PASS | `go test ./internal/template/... -run 'TestTemplateNeutralityAudit\|TestTemplateNoInternalContentLeak'` | `ok github.com/modu-ai/moai-adk/internal/template 0.691s` |
| AC-BGR-007 | REQ-BGR-007 | PASS (re-interpreted) | `make build` + `go build ./...` | `make build` exit 0; `go build ./...` exit 0. NOTE: there is NO generated `embedded.go` artifact — `internal/template/embed.go` uses `//go:embed all:templates`, embedding the corrected `templates/` tree directly at compile time. The corrected mirror content is therefore embedded by the successful `go build`. `make build` ran gen-catalog-hashes; `git diff --stat` shows ONLY the 6 surfaces (catalog.yaml unchanged — the 6 surfaces are not skill SKILL.md files, so no hash delta). |
| AC-BGR-008 | REQ-BGR-008 | PASS | run-phase doc re-confirmation recorded above (§E.2 M1 block) before the [HARD] rewrite | orchestrator-provided verbatim 2.1.186 surface from official sub-agents doc, recorded above |
| AC-BGR-009 | REQ (Exclusions) | PASS | `grep -c 'Background subagents (run_in_background: true) MUST NOT perform Write/Edit operations' zone-registry+mirror` | both report `1` — CONST-V3R2-044 UNCHANGED |
| AC-BGR-010 | REQ-BGR-007 | PASS | `go test ./...` | 99 ok packages, 0 FAIL packages on cached/isolated run. Transient first-run `signal: killed` timeouts on `TestHookWrapper_ValidJSON` / `TestHookWrapper_MoaiBinaryFallback` (5s wrapper-subprocess timeout under parallel load) re-ran green in isolation (`ok internal/hook 0.471s`). Environmental flakiness, NOT a regression: this SPEC modifies ZERO Go files (`git diff --name-only \| grep '\.go$'` = empty); `internal/hook` untouched. |

### Cross-platform build (E2)

```
$ go build ./...                          → exit 0
$ GOOS=windows GOARCH=amd64 go build ./... → exit 0
```

### Conclusion-retention proof (AC-BGR-003 / REQ-BGR-003)

The [HARD] behavioral conclusion survives in EVERY corrected surface:
- Locus 1 (CLAUDE.md §14, both trees): "keeps `run_in_background: false` for agents that modify files"
- Locus 2 (agent-common-protocol § Background Agent Execution, both trees): the [ZONE:Frozen] [HARD]
  line "Background subagents MUST NOT perform Write/Edit operations" is unchanged above the corrected
  paragraph, and the paragraph itself states "keeps `run_in_background: false` for write tasks"
- Locus 3 (zone-registry CONST-V3R2-020 clause, both trees): "keeps run_in_background: false for
  agents that modify files"
- CONST-V3R2-044 (out-of-scope, both trees): "MUST NOT perform Write/Edit operations" UNCHANGED

The restriction is NOT relaxed — this is rationale correction only.

### Scope discipline (B10)

`git diff --name-only` = exactly the 6 target surfaces. `git status --porcelain` untracked = only the
SPEC artifact dir. No GLM files, no Go code, no unrelated paths touched. CONST-V3R2-044 unchanged.

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-06-23
run_commit_sha: <backfill — see post-commit note>
run_status: implemented
ac_pass_count: 9            # AC-BGR-001,002,003,005,006,007,008,009,010
ac_fail_count: 0
ac_disclosed_deviation: 1   # AC-BGR-004 intentionally not met (allowlist sentence dropped per orchestrator CRITICAL FINDING; official doc contradicts the claim)
preserve_list_post_run_count: "CONST-V3R2-044 unchanged (1 each tree); [HARD] rule unchanged; '3-phase close' mirror drift untouched"
l44_pre_commit_fetch: "git fetch origin main → 0 0 (worktree HEAD == origin/main, clean)"
l44_post_push_fetch: <backfill post-push>
new_warnings_or_lints_introduced: 0   # doc-prose only; zero Go files changed
cross_platform_build:
  host_go_build: exit 0
  windows_go_build: exit 0
total_run_phase_files: 6   # 3 loci × 2 trees (CLAUDE.md + mirror, agent-common-protocol.md + mirror, zone-registry.md + mirror)
m1_to_mN_commit_strategy: "plan-artifact commit (draft) → single run commit (draft→in-progress, 6 surfaces + run evidence) on worktree branch → orchestrator reconciles to origin/main (worktree at 0 0 vs origin/main, clean FF)"
embedded_go_note: "no generated embedded.go artifact — internal/template/embed.go uses //go:embed all:templates; corrected templates embedded directly by go build"
```

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase — owned by manager-docs>_

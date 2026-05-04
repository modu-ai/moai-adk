# SPEC-V3R3-HYBRID-001 Progress

- plan_complete_at: 2026-05-03T20:20:23Z
- plan_status: audit-ready

## Artifacts

- spec.md — v0.1.0 (frontmatter v0.2.0 정합화 완료, 9 required fields; 18 REQs / 18 ACs; BC-V3R3-HYBRID-001 declared)
- research.md — Phase 0.5 deep research (33 file:line citations; codebase grounded scan of `internal/cli/{cg,glm,launcher,update}.go` + `internal/config/types.go` + 6 documentation surfaces; 4-provider Anthropic-compat endpoint verification table)
- plan.md — Phase 1B implementation plan (5 milestones M1-M5; 23 file:line anchors; mx_plan with 10 tags / 7 files; REQ↔AC traceability matrix 18/18)
- acceptance.md — Given/When/Then for 18 ACs (happy path + edge cases per AC; integration test scaffold names declared for `TestProviderAllowlist`, `TestCGCommand*`, `TestHybrid*`, `TestMigrateCGTeamMode*`)
- tasks.md — M1-M5 milestone breakdown (36 tasks T-HYBRID-01..36; owner roles aligned with TDD methodology — 19 expert-backend / 4 manager-tdd / 13 manager-docs)
- spec-compact.md — Auto-extracted REQ + AC + Files-to-modify + Exclusions reference

## Branch

- Branch: feature/SPEC-V3R3-HYBRID-001
- Mode: solo, no worktree (per user directive)
- Working directory: /Users/goos/MoAI/moai-adk-go
- Base: origin/main HEAD aa55780ce (Wave 6 plan PRs all merged: WF-004 #765, WF-003 #767, session-handoff #763)

## Key Plan Decisions

- BC scope: clean break (no deprecation alias period). `moai cg` 호출 시 즉시 `MOAI_CG_REMOVED` BC 에러 + actionable migration suggestion. `cgCmd`는 binary stub로 보존 (REQ-HYBRID-014) — cobra "unknown command" fallback 차단.
- 4-LLM allow-list: GLM (Z.AI), Kimi K2 (Moonshot AI), DeepSeek V4, Qwen3-Coder (Alibaba DashScope). 닫힌 집합 (REQ-HYBRID-002, REQ-HYBRID-017). 5번째 provider 추가는 atomic-reversal SPEC 필요.
- Endpoint verification status (research.md §3): GLM `https://api.z.ai/api/anthropic` (verified), DeepSeek `https://api.deepseek.com/anthropic` (verified), Kimi `https://api.moonshot.ai/anthropic` (pending — `--proxy` fallback), Qwen3 `https://dashscope.aliyuncs.com/compatible-mode/anthropic` (pending — `--proxy` fallback).
- TDD methodology: per `.moai/config/sections/quality.yaml development_mode: tdd`. M1 RED gate (4 test files) → M2/M3/M4 GREEN gates (provider abstraction → hybrid command + cg BC stub → injection refactor + launcher routing + migration) → M5 REFACTOR + Trackable (documentation substitution + CHANGELOG + 10 MX tags).
- Auto-migration: `moai update` cleanup phase가 `team_mode: "cg"` config를 자동으로 `team_mode: "hybrid"` + `provider: "glm"`으로 rewrite (REQ-HYBRID-011). Atomic write infrastructure는 SPEC-V3R3-UPDATE-CLEANUP-001 (PR #764 merged) 차용.
- Provider preservation: `moai glm` (단일 LLM all-GLM) + `moai cc` (Claude-only) 동작 무변경 (REQ-HYBRID-006). `moai hybrid`는 multi-LLM 슈퍼셋이지 대체 아님.
- Sentinel strings (NEW): `UNKNOWN_PROVIDER` (REQ-HYBRID-002), `MOAI_CG_REMOVED` (REQ-HYBRID-003, 010, 014), `HYBRID_REQUIRES_TMUX` (REQ-HYBRID-008), `HYBRID_MISSING_API_KEY` (REQ-HYBRID-009), `PROVIDER_ALLOWLIST_VIOLATION` (REQ-HYBRID-017), `HYBRID_AUTH_FAILED` (REQ-HYBRID-018).
- Provider-namespaced backup keys: `MOAI_BACKUP_AUTH_TOKEN_<UPPER>` (e.g., `MOAI_BACKUP_AUTH_TOKEN_GLM`). 단일 슬롯 패턴 deprecate — provider 전환 시 데이터 손실 방지 (research.md §5 risk → MX:WARN tag).

## Frontmatter Migration Verification (spec.md v0.1.0)

- Required fields present (9/9): `id`, `version`, `status`, `created_at`, `updated_at`, `author`, `priority`, `labels`, `issue_number` ✅
- Rejected aliases absent (0): `created`, `updated`, `spec_id`, `title:` H1-alias ✅
- `version` quoted string: `"0.1.0"` ✅
- `priority` enum: `P1` (bare uppercase, no descriptor) ✅
- `labels` YAML array: `[hybrid, multi-llm, glm, kimi, deepseek, qwen, breaking-change, v3r3]` ✅
- `created_at` / `updated_at` ISO date: `2026-05-04` / `2026-05-04` ✅
- `issue_number: null` ✅
- Optional BC fields: `breaking: true` + `bc_id: [BC-V3R3-HYBRID-001]` + `lifecycle: spec-anchored` ✅

## Codebase Scan Summary (research.md grounded)

### `moai cg` command surface (impacted files)

1. `internal/cli/cg.go` (lines 1-58) — current `cgCmd` definition, M3 REPLACE target.
2. `internal/cli/launcher.go:57` — `case "claude_glm":` mode dispatch, M4 generalize target.
3. `internal/cli/launcher.go:185` — `persistTeamMode(root, "cg")`, M4 update target.
4. `internal/cli/glm.go:148-149` — stderr message "moai cg" mention, M4 substitution.
5. `internal/cli/glm.go:266` — `persistTeamMode(root, "cg")`, M4 update target.

### `moai cg` documentation surface (substitution targets)

1. `.claude/skills/moai/team/glm.md` — 5 mentions (lines 41, 50, 103, 127, 143)
2. `.claude/skills/moai/team/run.md` — 5 mentions (lines 73, 88, 92, 103, 104)
3. `.claude/skills/moai/workflows/run.md:904` — 1 mention (`active_mode: cc | glm | cg`)
4. `.claude/rules/moai/development/model-policy.md:44` — 1 mention (Activation line)
5. `CLAUDE.md §15` (~line 488-510) — `### CG Mode` heading + body
6. `CLAUDE.local.md §16 line 129` — runtime-managed config list

### Provider abstraction layer (NEW additions)

- `internal/llm/provider.go` — Provider interface (~6 methods)
- `internal/llm/providers/{glm,kimi,deepseek,qwen}.go` — 4 provider metadata files
- `internal/llm/providers/registry.go` — closed allow-list Registry (MX:ANCHOR)
- `internal/llm/inject.go` — provider-agnostic env-injection helpers (refactor of `glm.go:356-581`)

### Schema extensions

- `internal/config/types.go:52-100` `LLMConfig` — add `Provider string` + `Providers map[string]ProviderYAML` fields
- `.moai/config/sections/llm.yaml` — add top-level `provider: ""` + `providers:` section (4 entries)

## Next Phase

- Phase 0.5 Plan Audit Gate (plan-auditor) at `/moai run SPEC-V3R3-HYBRID-001` entry — see `.claude/rules/moai/workflow/spec-workflow.md:172-204`.
- Implementation Methodology: TDD (per `.moai/config/sections/quality.yaml`).
- Run-phase command: `/moai run SPEC-V3R3-HYBRID-001` (executed from `/Users/goos/MoAI/moai-adk-go` on branch `feature/SPEC-V3R3-HYBRID-001`).
- Post-implementation: `/moai sync SPEC-V3R3-HYBRID-001` for documentation sync (docs-site 4-locale per CLAUDE.local.md §17) + PR creation.

## Plan-Audit-Ready Checklist Summary

All 18 criteria PASS per plan.md §8:

- C1: Frontmatter v0.2.0 (9 required fields) ✅
- C2: HISTORY v0.1.0 entry ✅
- C3: 18 EARS REQs across 5 categories (Ubiquitous 6, Event-Driven 5, State-Driven 3, Optional 2, Complex 2) ✅
- C4: 18 ACs with 100% REQ mapping (18/18 REQ → AC traceability matrix in plan.md §1.4) ✅
- C5: BC scope clarity (clean break, `breaking: true`, `bc_id: [BC-V3R3-HYBRID-001]`) ✅
- C6: File:line anchors ≥10 (research.md: 33, plan.md: 23) ✅
- C7: Exclusions section present (spec.md §1.2 Non-Goals + §2.2 Out of Scope; spec-compact.md §Exclusions with 13 entries) ✅
- C8: TDD methodology declared ✅
- C9: mx_plan section (10 tags / 7 files; 3 ANCHOR + 3 NOTE + 3 WARN + 1 TODO) ✅
- C10: Risk table with file-anchored mitigations (spec.md §8: 10 risks; plan.md §5: 14 risks) ✅
- C11: Solo mode path discipline (4 HARD rules, no worktree per user directive) ✅
- C12: No implementation code in plan documents ✅
- C13: Acceptance.md G/W/T format with edge cases (18 ACs covered) ✅
- C14: tasks.md owner roles aligned with TDD (19 expert-backend / 4 manager-tdd / 13 manager-docs) ✅
- C15: Cross-SPEC consistency (SPEC-GLM-MCP-001 PR #769 OPEN dependency declared; SPEC-V3R3-UPDATE-CLEANUP-001 PR #764 merged infrastructure reused; SPEC-V3R2-WF-005 PR #768 merged neutrality pattern applied) ✅
- C16: BC migration completeness (spec.md §10: BC ID, migration table, CHANGELOG wording, docs touch points, removal timeline) ✅
- C17: 4-LLM allow-list documented (REQ-HYBRID-002 + REQ-HYBRID-017; AC-HYBRID-15 + AC-HYBRID-18 5번째 provider 차단 검증) ✅
- C18: Endpoint verification status documented (research.md §3 + §3.5 verification table; GLM/DeepSeek verified, Kimi/Qwen3 pending — `--proxy` fallback per REQ-HYBRID-015) ✅

## Open Items for plan-auditor Review

- Confirm `internal/llm/` 패키지 트리가 기존 codebase에 존재하지 않음 (verified absent at baseline; M2 신규 패키지). Final check 권장.
- Validate Kimi/Qwen3 Anthropic-compat endpoint candidate URLs (research.md §3.2, §3.4 pending) — provider 공식 docs 재확인 시점에 base_url 갱신 가능. `--proxy` fallback (REQ-HYBRID-015)이 stable escape hatch.
- Confirm `internal/cli/cg.go` BC stub은 cobra root에서 등록 보존되어야 함 (REQ-HYBRID-014). 단순 파일 삭제 / `cgCmd` un-registration 시 cobra "unknown command" fallback으로 actionable 메시지 손실. M3 T-HYBRID-19 sequence 검증.
- Verify `MOAI_BACKUP_AUTH_TOKEN_<UPPER>` namespaced backup pattern은 v2.x 단일 슬롯 방식과 backward-incompatible — 기존 `MOAI_BACKUP_AUTH_TOKEN` 사용자 데이터 마이그레이션 / cleanup path 명시 필요 시 plan §5 risk row 보강.
- Validate `migrateCGTeamMode` idempotency contract — 두 번째 실행이 stderr notice 미발행 (silent no-op) 여부 — M1 T-HYBRID-05 `TestMigrateCGTeamModeIdempotent` 케이스 검증.
- Confirm SPEC-GLM-MCP-001 (PR #769 OPEN) implementation 완료 시점이 `moai hybrid glm tools` UX 활성화 trigger — 본 SPEC run-phase에서 stub만 제공, 실제 attach 로직은 SPEC-GLM-MCP-001 후속 PR에서 활성화.

---

End of progress.md.

# docs-site 업데이트 계획 — v2.14.0 → HEAD

> **목적**: v2.14.0 릴리스 이후 main HEAD까지 누적된 사용자 가시 변화를 `adk.mo.ai.kr` (docs-site)에 반영하기 위한 분석 + 단계별 실행 계획.
>
> **산출물 유형**: research doc (SPEC artifact 아님). 본 문서 합의 후 후속 작업은 별도 SPEC(s)로 분기 가능.
>
> **분석 깊이**: User-facing 변화만 (feat/fix 중 사용자 동작·CLI·workflow·설정에 영향이 있는 항목).
>
> **4-locale 정책**: §17.3에 따라 ko 우선 + en/ja/zh 동일 PR 동기화.

---

## 1. 범위 + 사실관계

### 1.1 기준선

| 항목 | 값 |
|------|-----|
| 직전 릴리스 태그 | `v2.14.0` |
| 비교 대상 | `main` HEAD (current: `03a2552a2` SECURITY-CRIT-001 머지 직후) |
| 전체 commit 수 | 313 (merge 포함) |
| Non-merge feat/fix/hotfix | 133 |
| 분석 대상 user-facing commit | ~80 (feat/fix만, internal lint/test/CI 제외 후) |
| 머지된 SPEC lifecycle 수 | 50+ (SPEC-V3R2~V3R5 + SPC/MIG/RT/ORC/HRN/CON/STATUS/CATALOG/WORKFLOW/HARNESS/CI/DESIGN/GLM/MX/KARPATHY/TUI/UPDATE/AGENCY/CLAUDE-REFRESH/CONSTITUTION-DUAL/CORE-SLIM/HARNESS-AUTONOMY/WORKFLOW-OPT/WORKFLOW-LEAN/LATE-BRANCH/STATUSLINE-V2145/SECURITY-CRIT) |

### 1.2 docs-site 현재 구조 (ko canonical)

13개 top-level section / ~58개 콘텐츠 파일:

| Section | 파일 수 | 주요 페이지 |
|---------|---------|-------------|
| `advanced/` | 13 | agent-guide, builder-agents, catalog-system, claude-md-guide, harness-profiles, hooks-guide, hooks-reference, mcp-servers, pencil-guide, settings-json, skill-guide, statusline, stitch-guide |
| `contributing/` | 0 (`_index.md`만) | (미완성 — boundary doc 신설 후보) |
| `core-concepts/` | 7 | constitution, ddd, harness-engineering, moai-memory, spec-based-dev, trust-5, what-is-moai-adk |
| `db/` | 4 | getting-started, migration-patterns, project-db-directory, schema-sync |
| `design/` | 5 | claude-design-handoff, code-based-path, gan-loop, getting-started, migration-guide |
| `getting-started/` | 9 | cli, faq, init-wizard, installation, introduction, profile, quickstart, update, windows-guide |
| `guides/` | 2 | ci-autonomy, multi-llm-ci |
| `multi-llm/` | 2 | cg-mode, model-policy |
| `quality-commands/` | 4 | moai-codemaps, moai-coverage, moai-e2e, moai-review |
| `utility-commands/` | 7 | moai, moai-clean, moai-feedback, moai-fix, moai-github, moai-loop, moai-mx |
| `workflow-commands/` | 7 | moai-brain, moai-db, moai-design, moai-plan, moai-project, moai-run, moai-sync |
| `worktree/` | 3 | examples, faq, guide |

4-locale: `docs-site/content/{ko,en,ja,zh}/` 모두 동일 트리.

### 1.3 분석에서 제외한 항목

- Internal-only refactor / lint / CI guard / template-mirror parity 작업 (사용자 체감 0)
- `chore(spec)`, `test(...)`, `docs(spec)` (SPEC 산출물만 영향)
- pre-existing baseline cleanup (e.g., SPEC-V3R5-LINT-CLEAN-001 — debt 해소만, 사용자 동작 무변경)
- SPEC artifact lifecycle metadata-only update (`status: draft → completed`)

---

## 2. SPEC 그룹 분류 (10개 그룹)

각 SPEC을 사용자 가시 도메인 기준으로 분류. 그룹 내 SPEC들은 docs-site 동일 페이지(들)에 수렴.

### Group A — Harness Autonomy (자체 진화 메커니즘 신설)

| SPEC | 효과 |
|------|------|
| **SPEC-V3R5-HARNESS-AUTONOMY-001** (W3) | 4-Tier evolution + 5-Layer safety + 18 sentinels (8 FROZEN + 10 LEARNING) + 10 CLI verbs (route/validate/status/apply/rollback/disable/mute/mute-list/unmute/verify) |
| SPEC-V3R3-HARNESS-LEARNING-001 | Self-Learning Dynamic Harness (5-wave) |
| SPEC-V3R3-PROJECT-HARNESS-001 | 16Q interview + 5-Layer 통합 |
| SPEC-V3R2-HRN-001 | Harness Routing + harness.yaml Go Loader + `moai harness route/validate` CLI |
| SPEC-V3R2-HRN-002 | Evaluator Memory Scope amendment |
| SPEC-V3R2-HRN-003 | Profile loader + EvaluatorConfig + agent body augment |
| SPEC-V3R4-HARNESS-001 | V3R4 self-evolving harness foundation + CLI retirement + lifecycle consolidation |
| SPEC-V3R4-HARNESS-002 | Multi-Event Observer Expansion (Stop/SubagentStop/UserPromptSubmit) |
| SPEC-V3R4-HARNESS-NAMESPACE-001 | harness namespace governance |
| moai-meta-harness 신설 | revfactory/harness Apache 2.0 흡수 (7-Phase 메타-시스템) |

→ **docs 영향**: `core-concepts/harness-engineering.md` (대폭 재작성), `advanced/harness-profiles.md` (신규 CLI verbs + 18 sentinels 추가), `workflow-commands/moai-project.md` (16Q interview)

### Group B — Workflow Discipline (Tier S/M/L + Late-Branch + 8-Layer)

| SPEC | 효과 |
|------|------|
| **SPEC-V3R5-WORKFLOW-LEAN-001** | Tier S/M/L SPEC complexity + artifact 2/3/5 차등 + Section A-E optional (Tier S) + plan-auditor threshold 0.75/0.80/0.85 + `tier:` optional frontmatter |
| **SPEC-V3R5-WORKFLOW-OPT-001** | 8-Layer Workflow Optimization (manager-develop prompt 5-section 표준화 + B1-B8 known issues + verification batch pattern + Agent Teams 5+1+1 composition) |
| **SPEC-V3R5-LATE-BRANCH-001** | Late-branch workflow (main 위 commits 누적 → PR 시점 `git switch -c` → squash merge → reset) + no auto GitHub Issue (default opt-in flip) |
| SPEC-V3R4-WORKFLOW-SPLIT-001 (Wave 1~4) | run/sync/project/plan skill 4 monolithic (~4284 LOC) → 4 thin entry router (≤200 LOC) + 13 phase-scoped sub-skill (≤500 LOC). 토큰 부담 -76% |
| SPEC-V3R2-WF-003 | Multi-Mode Router (`--mode` flag for run/loop/design) |
| SPEC-V3R2-WF-004 | Agentless 분류 매트릭스 + 5+4 subcommand 분류 (utility는 fixed pipeline) |
| SPEC-V3R2-WF-005 | Language rules vs skills boundary 명문화 |

→ **docs 영향**: `core-concepts/spec-based-dev.md` (Tier S/M/L 신설 + Section A-E template + EARS hierarchical AC), `workflow-commands/moai-plan.md` (frontmatter tier 필드 + 점진적 artifact 발행), `workflow-commands/moai-run.md` (`--mode` flag + Late-branch), `workflow-commands/moai-sync.md` (Late-branch PR 절차), `utility-commands/moai.md` (Agentless classification), `contributing/_index.md` (language rules vs skills boundary 새 페이지 후보)

### Group C — Agents & Skills (Catalog 변화)

| SPEC | 효과 |
|------|------|
| **SPEC-V3R4-CATALOG-001** | 3-tier catalog manifest + typed loader |
| **SPEC-V3R4-CATALOG-002** | slim init via SlimFS tier filter |
| SPEC-V3R5-CORE-SLIM-001 | LR-08 rule refinement + expert agent foundation symmetry |
| SPEC-V3R5-CORE-SLIM-B-001 | Category B dead-weight skill retire (1,432 LOC, 4 skills) + 2 language rule cross-ref + catalog.yaml 정리 |
| SPEC-V3R2-ORC-001 | Agent roster 22 → 17 consolidation |
| **SPEC-V3R2-ORC-002** | `moai agent lint` 8-rule engine (LR-01~LR-08) + `--strict`/`--format=json`/`--path` flags + CI guard |
| SPEC-V3R2-ORC-003 | Effort-Level Calibration Matrix scaffolding (LR-03 promotion path) |
| SPEC-V3R2-ORC-004 | Worktree MUST → SHOULD harmonization (SPEC-WORKTREE-DOCS-001 후속에서 SHOULD-tier로 격하) |
| SPEC-V3R2-ORC-005 | Dynamic Team Generation Formalization (general-purpose + role profiles) |
| SPEC-V3R3-RETIRED-AGENT-001 + RETIRED-DDD-001 | manager-ddd retired stub 표준화 + manager-cycle 템플릿 정합화 |
| moai-meta-harness 신설 | revfactory/harness Apache 2.0 흡수 |
| 16 정적 skills 제거 (BC-V3R3-007) | template 슬림화 |
| archive 마이그레이터 + `restore-skill` subcommand | moai update M4 |
| doctor namespace 분리 (`moai-*` / `my-harness-*`) | template diff 가독성 |

→ **docs 영향**: `advanced/agent-guide.md` (agent roster 17 + `moai agent lint` 사용법), `advanced/skill-guide.md` (Skill catalog + 16 정적 skill 제거 + `moai-meta-harness` 신규), `advanced/catalog-system.md` (3-tier manifest + slim init), `advanced/builder-agents.md` (builder-platform/harness 활용)

### Group D — Constitution & Rules

| SPEC | 효과 |
|------|------|
| **SPEC-V3R5-CONSTITUTION-DUAL-001** (W1) | Constitution Dual-Zone (Frozen/Evolvable) + zone_class 4-enum + `moai constitution validate` CLI + 111 [HARD] ZONE markers across 15 source files + 9 sentinel keys |
| SPEC-V3R5-CLAUDE-REFRESH-001 (W0) | Architecture Truth + Bundle A Settings (CLAUDE.md §5 rewrite + AskUserQuestion N=29→9 SSOT compression + Footer v14.0.0→v14.2.0) |
| SPEC-KARPATHY-001 | Karpathy Coding Principles 통합 (4 principles + anti-pattern catalog 흡수) |
| SPEC-V3R2-CON-001 + CON-002 + EXT-001 restore | Constitution + Memory + Extension restore |

→ **docs 영향**: `core-concepts/constitution.md` (Dual-Zone 신설 + `moai constitution validate` 추가), `core-concepts/what-is-moai-adk.md` (Architecture Truth 반영), `core-concepts/moai-memory.md` (auto-memory + Lessons Protocol 갱신 가능성), `core-concepts/trust-5.md` (Karpathy 4 principles 흡수 표시)

### Group E — Security (P0 정정)

| SPEC | 효과 |
|------|------|
| **SPEC-V3R5-SECURITY-CRIT-001** (PR #1032 머지 완료, `03a2552a2`) | 3건 P0 보안 결함 정정 — (M1) settings.local.json `0o600` hardening (CWE-732/552), (M2) tmux sensitive env source-file injection (CWE-214), (M3) mandatory checksum verification with retry (CWE-345), (M4) cross-cutting verification + frontmatter |

→ **docs 영향**: `advanced/settings-json.md` (settings.local.json 0o600 권고 추가), `getting-started/update.md` (checksum 검증 동작 명시), **신규**: `advanced/security-notes.md` 또는 `getting-started/security.md` (CWE-732/214/345 사용자 가시 변화 요약 + 권장 점검)

### Group F — Settings / Hooks / Permission / Sandbox

| SPEC | 효과 |
|------|------|
| **SPEC-V3R2-RT-002** | 8-tier Permission Stack + Bubble Mode + `moai doctor permission` 확장 (`--all-tiers`/`--mode`/`--fork`/`--format`) + 5-enum PermissionMode strict + `security.yaml` 신규 키 (strict_mode, pre_allowlist, session_rules) |
| **SPEC-V3R2-RT-003** | Sandbox Execution Layer — Bubblewrap (Linux) + Seatbelt (macOS) + Docker (CI) + `moai doctor sandbox` CLI + `Sandbox` 4-값 열거형 + agent_lint LR-33 rule + `security.yaml` sandbox 키 (BC-V3R2-003 breaking) |
| SPEC-V3R2-RT-004 | Typed Session State + Phase Checkpoint |
| SPEC-V3R2-RT-005 | Multi-layer settings resolution (`.moai/config/sections/*.yaml`) |
| **SPEC-V3R2-RT-006** | Hook Handler 27-Event Coverage + observability opt-in + `moai doctor hook` CLI + tmux pane leak fix (P-H02) + 4 hook event retire (Notification/Elicitation/ElicitationResult/TaskCreated) — BC-V3R2-018 |
| SPEC-V3R2-RT-007 | Hardcoded Path Fix + Versioned Migration |
| SPEC-V3R2-MIG-002 | Hook Registration Cleanup + EventSetup retirement + 3-way sync invariant |
| SPEC-V3R2-MIG-003 | Config Loader Completeness — 4 new section loaders + CI guards |
| SPEC-V3R4-HOOK-HARDEN-001 | Hook 강화 Wave 1 (Visibility F1+F2) |
| SPEC-CC2122-HOOK-002 | yaml config + agents disallowedTools 정적 linter |
| SPEC-V3R3-UPDATE-CLEANUP-001 | `moai update` 멱등 배포 + 폐기 경로 정리 |

→ **docs 영향**: `advanced/settings-json.md` (대폭 — 8-tier permission + sandbox + 27-event coverage + security.yaml schema), `advanced/hooks-guide.md` (4 event retire + observability opt-in), `advanced/hooks-reference.md` (27-event table 갱신 + duration_ms + updatedToolOutput), `getting-started/cli.md` (`moai doctor permission/sandbox/hook` 3 신규 subcommand), `getting-started/update.md` (멱등 배포 + 폐기 경로)

### Group G — Multi-LLM + MCP + Claude Code 호환성

| SPEC | 효과 |
|------|------|
| SPEC-GLM-MCP-001 | Z.AI MCP 서버 통합 (`zai-mcp-server enable/disable`) |
| SPEC-CI-MULTI-LLM-001 | Multi-LLM CI 통합 + Wizard + Auth 수정 |
| Claude Code v2.1.119 MCP alwaysLoad | pre-load support |
| Claude Code v2.1.119-121 hook | duration_ms + updatedToolOutput |
| Claude Code v2.1.122 | effort + thinking indicator |
| **SPEC-V3R5-STATUSLINE-V2145-001** | Statusline v2.1.145 alignment (disappearing fix + PR segment + 4-locale docs) |
| **SPEC-V3R3-CLI-TUI-001** M1~M7 | TUI 패키지 골격 + 6 TUI primitives + 74 snapshots + banner DDD migration + version/doctor/status/update DDD migration + huh wizard ◆/◇ Theme + Stepper + 5-command batch + auto-detect + NO_COLOR + i18n + golden suite |

→ **docs 영향**: `advanced/mcp-servers.md` (alwaysLoad + Z.AI MCP), `advanced/statusline.md` (v2.1.145 PR segment + effort/thinking indicator), `advanced/hooks-reference.md` (duration_ms + updatedToolOutput), `multi-llm/cg-mode.md` + `multi-llm/model-policy.md` (Z.AI 통합), `guides/multi-llm-ci.md` (CI 통합 + Wizard 흐름), `getting-started/init-wizard.md` (huh wizard 도입 + TUI 갱신), `getting-started/cli.md` (TUI DDD migration 영향)

### Group H — CI Autonomy + Branch Origin Protocol

| SPEC | 효과 |
|------|------|
| **SPEC-V3R3-CI-AUTONOMY-001** Wave 1~7 | (W1) pre-push hook + branch protection / (W2) CI watch loop + skill (`moai pr watch`) / (W3) auto-fix loop (max 3 iterations) / (W4) auxiliary workflow cleanup / (W5) worktree state guard / (W6) i18n validator (`scripts/docs-i18n-check.sh`) / (W7) Branch Origin Decision Protocol (P0, final — `moai worktree new --base` flag) |
| SPEC-V3R4-CI-FASTTRACK-001 | 1-developer 3-tier CI philosophy + paths-filter docs-only fast-path + 5 review workflow 제거 + lefthook + preflight + release-pr-multi-os.yml (3-OS full matrix on release/*) |
| SPEC-V3R4-CI-INFRA-FIX-001 | 3-defect bundle fix (SIGPIPE + 403 + fetch-depth: 0) |
| SPEC-WORKTREE-DOCS-001 | Worktree workflow rules SHOULD-tier로 격하 + L1/L2/L3 terminology 표준화 (4-row Terminology Glossary) |

→ **docs 영향**: `guides/ci-autonomy.md` (대폭 — 7-Wave CI watch + auto-fix + i18n validator + BODP), `worktree/guide.md` (L1/L2/L3 terminology + state guard + `moai worktree new --base`), `worktree/examples.md` (BODP 예시), `worktree/faq.md` (autonomous policy 2026-05-17)

### Group I — SPEC Tooling (Lint / Status / @MX / Brain)

| SPEC | 효과 |
|------|------|
| **SPEC-V3R2-SPC-001** | EARS hierarchical acceptance criteria framework (flat `AC-XXX-NN` → tree `AC-XXX-NN.a/.b` with inherited Given) + `MaxDepth=3` + `moai spec view --shape-trace` CLI |
| **SPEC-V3R2-SPC-003** | `moai spec lint` CLI (Wave 5) |
| SPEC-V3R2-SPC-004 | `@MX` anchor resolver query API + `moai mx query` CLI |
| SPEC-V3R4-SPECLINT-DEBT-001 | spec-lint debt 일괄 해소 (66 ERROR + 141 WARNING → 0/0) |
| SPEC-V3R4-SPECLINT-DEBT-002 | plan.md 12-field canonical 정렬 + SSOT 신설 |
| SPEC-V3R4-STATUS-LIFECYCLE-001 Wave 1~3 | 7-Layer Defense (Policy + Lint / Hook + Transitions / Automation + Visibility) |
| SPEC-STATUS-AUTO-001 | SPEC status auto-update system |
| SPEC-V3R4-STATUS-DRIFT-FOLLOWUP-001/002 | 18+77건 status drift 일괄 해소 + terminal-state exemption |
| SPEC-V3R4-LINT-SPECID-GREP-FIX-001 | walker SPEC-ID word-boundary filter |
| SPEC-V3R4-LINT-STATUS-CHORE-SKIP-001 | drift.go chore(spec) walker filter |
| **SPEC-V3R3-BRAIN-001** | `/moai brain` 7-phase ideation workflow 구현 |
| SPEC-V3R3-DESIGN-PIPELINE-001 | Hybrid Design Pipeline (6-wave) |
| SPEC-V3R3-DESIGN-FOLDER-FIX-001 | reserved collision update path warning 격하 |

→ **docs 영향**: `workflow-commands/moai-plan.md` (EARS hierarchical AC + `moai spec view --shape-trace` + 12-field canonical schema), `workflow-commands/moai-sync.md` (auto-update + 7-Layer Defense), `workflow-commands/moai-brain.md` (7-phase ideation 구현 완료 + AskUserQuestion 흐름), `workflow-commands/moai-design.md` + `design/*` (6-wave hybrid pipeline + reserved collision warning), `utility-commands/moai-mx.md` (`moai mx query` 신규)

### Group J — Statusline / Theme / Wizard / Banner

| SPEC | 효과 |
|------|------|
| **SPEC-V3R5-STATUSLINE-V2145-001** | Statusline v2.1.145 alignment (disappearing fix + PR segment + 4-locale docs) |
| Statusline 개선 (`9373e558f`) | statusline + deps + worktree --tmux + release/githug |
| Statusline stdin fallback (`25dce1c63`) | cwd guard + model name fallback chain |
| Statusline GLM context size (`716c54a25`) | MOAI_STATUSLINE_CONTEXT_SIZE GLM 모델별 자동 주입 |
| Banner DDD migration (`f359a0fb2`) | terra cotta + 보라 hex 제거 |
| Theme adaptive detection (`6671d4185`) | 어두운 터미널 가독성 |

→ **docs 영향**: `advanced/statusline.md` (이미 5/20 갱신됨 — PR segment + effort + thinking indicator 검증 필요)

### Group K — Miscellaneous (사용자 가시 작은 변화)

| Commit | 효과 |
|--------|------|
| `feat(workflow): AskUserQuestion Enforcement Protocol §19` | CLAUDE.local.md §19 (deferred tool preload + pre-response self-check + anti-patterns) — **dev-only** 가능성 검토 후 사용자 문서 반영 |
| `feat: statusline + worktree --tmux + release/githug` | worktree --tmux 옵션 + githug 개선 |
| `feat(dev-tooling): Makefile dev-sync target + MCP servers default` | dev-only |
| `feat(brain): 7-phase ideation` | moai-brain.md 갱신 대상 |

---

## 3. docs-site 영향 매트릭스

핵심 변경 페이지를 우선순위(P0/P1/P2) + 작업 유형(NEW/MAJOR/MINOR/VERIFY)별로 정리.

### 3.1 P0 — Critical (사용자 가시 동작·보안·CLI 변화)

| 페이지 | 유형 | 영향 SPEC | 변경 요약 |
|--------|------|-----------|-----------|
| `advanced/settings-json.md` | MAJOR | RT-002, RT-003, RT-005, RT-006, MIG-002, MIG-003, SECURITY-CRIT-001 | 8-tier permission stack + sandbox 4-enum + 27-event coverage + security.yaml schema + settings.local.json 0o600 hardening |
| `advanced/hooks-guide.md` | MAJOR | RT-006, MIG-002, HOOK-HARDEN-001, HARNESS-002 | 27-event coverage + 4 event retire (observability opt-in) + Multi-Event Observer + Hook Handler 강화 |
| `advanced/hooks-reference.md` | MAJOR | RT-006, Claude Code v2.1.119-121 | 27-event 표 갱신 + duration_ms + updatedToolOutput |
| `advanced/agent-guide.md` | MAJOR | ORC-001, ORC-002, ORC-003, ORC-004, ORC-005, CORE-SLIM-001, AGENT-RETIRED-001 | Agent roster 22→17 + `moai agent lint` 8 rules + Dynamic Team Generation + retired stub |
| `advanced/skill-guide.md` | MAJOR | CORE-SLIM-B-001, BC-V3R3-007 (16 정적 skill 제거), moai-meta-harness 신설 | 4 dead-weight skill retire + 16 static skill 제거 + meta-harness 흡수 |
| `advanced/catalog-system.md` | MAJOR | CATALOG-001, CATALOG-002, CORE-SLIM-B-001 | 3-tier catalog manifest + slim init via SlimFS + dead-weight 정리 |
| `advanced/harness-profiles.md` | MAJOR | HARNESS-AUTONOMY-001, HRN-001/002/003, HARNESS-LEARNING-001, HARNESS-NAMESPACE-001 | 4-Tier evolution + 5-Layer safety + 18 sentinels + 10 CLI verbs |
| `advanced/mcp-servers.md` | MINOR | GLM-MCP-001, Claude Code v2.1.119 alwaysLoad | Z.AI MCP + alwaysLoad pre-load |
| `core-concepts/spec-based-dev.md` | MAJOR | WORKFLOW-LEAN-001, WORKFLOW-OPT-001, SPC-001, WORKFLOW-SPLIT-001 | Tier S/M/L + Section A-E + EARS hierarchical AC + sub-skill split |
| `core-concepts/constitution.md` | MAJOR | CONSTITUTION-DUAL-001, KARPATHY-001 | Dual-Zone + `moai constitution validate` + Karpathy 4 principles |
| `core-concepts/harness-engineering.md` | MAJOR | HARNESS-AUTONOMY-001, HARNESS-LEARNING-001, PROJECT-HARNESS-001, moai-meta-harness | 4-Tier + 5-Layer + 16Q interview + meta-harness |
| `core-concepts/what-is-moai-adk.md` | MINOR | CLAUDE-REFRESH-001 (W0) | Architecture Truth 갱신 |
| `workflow-commands/moai-plan.md` | MAJOR | LATE-BRANCH-001, WORKFLOW-LEAN-001, SPC-001/003, BRAIN-001, STATUS-LIFECYCLE-001 | Late-branch + tier 필드 + hierarchical AC + `moai spec view --shape-trace` + no auto GH Issue (default opt-in) |
| `workflow-commands/moai-run.md` | MAJOR | WF-003, LATE-BRANCH-001 | `--mode` flag + Late-branch commit accumulation |
| `workflow-commands/moai-sync.md` | MAJOR | LATE-BRANCH-001, STATUS-LIFECYCLE-001, STATUS-AUTO-001 | Late-branch PR phase + auto status update + 7-Layer Defense |
| `workflow-commands/moai-project.md` | MAJOR | PROJECT-HARNESS-001, WORKFLOW-SPLIT-001 Wave 3 | 16Q interview + 5-Layer 통합 + 4 sub-skill 분할 |
| `workflow-commands/moai-brain.md` | MAJOR | BRAIN-001 | 7-phase ideation 구현 (구현 완료 — 기존 문서 placeholder→실제 흐름) |
| `workflow-commands/moai-design.md` | MAJOR | DESIGN-PIPELINE-001 | Hybrid 6-wave pipeline |
| `getting-started/cli.md` | MAJOR | TUI-001 M1~M7, RT-002, RT-003, RT-006, SPC-003, SPC-004, HRN-001, ORC-002, CONSTITUTION-DUAL-001 | 신규/변경 CLI verbs: `moai constitution validate`, `moai harness route/validate`, `moai agent lint`, `moai spec lint`, `moai spec view --shape-trace`, `moai mx query`, `moai doctor permission/sandbox/hook`, `moai pr watch`, `moai worktree new --base` |
| `getting-started/init-wizard.md` | MAJOR | TUI-001 M4/M5/M6 | huh wizard ◆/◇ Theme + Stepper + 5-command batch |
| `getting-started/update.md` | MAJOR | UPDATE-CLEANUP-001, RT-007, SECURITY-CRIT-001 M3 | 멱등 배포 + 폐기 경로 + checksum verification |

### 3.2 P1 — High (워크플로우 가이드, 일관성)

| 페이지 | 유형 | 영향 SPEC | 변경 요약 |
|--------|------|-----------|-----------|
| `advanced/statusline.md` | VERIFY | STATUSLINE-V2145-001 (2026-05-20 갱신됨) | PR segment + effort + thinking indicator — 4-locale parity 검증만 |
| `advanced/claude-md-guide.md` | MINOR | CLAUDE-REFRESH-001, AskUserQuestion §19 | CLAUDE.md §5 rewrite 반영 (architecture truth) |
| `advanced/mcp-servers.md` | MINOR | GLM-MCP-001 | Z.AI MCP server 등록 절차 |
| `core-concepts/trust-5.md` | MINOR | KARPATHY-001 | Karpathy 4 principles + anti-pattern 흡수 표시 |
| `core-concepts/moai-memory.md` | MINOR | CLAUDE-REFRESH-001 | Lessons Protocol / auto-memory 사용 갱신 |
| `core-concepts/ddd.md` | MINOR | RT-004 (Typed Session State) | RED-GREEN-REFACTOR 사이클 표현 갱신 가능성 |
| `guides/ci-autonomy.md` | MAJOR | CI-AUTONOMY-001 Wave 1~7, CI-FASTTRACK-001, CI-INFRA-FIX-001 | 7-Wave CI watch + auto-fix + i18n validator + BODP + 3-tier CI philosophy |
| `guides/multi-llm-ci.md` | MAJOR | CI-MULTI-LLM-001 | Multi-LLM CI + Wizard + Auth |
| `multi-llm/cg-mode.md` | MAJOR | GLM-MCP-001, CG mode | Claude leader + GLM teammates (tmux 세션 기반) |
| `multi-llm/model-policy.md` | MINOR | 16 moai/ agents sonnet/opus → inherit (`440a54690` `66e0c8e69`) | model 정책 정리 |
| `worktree/guide.md` | MAJOR | CI-AUTONOMY-001 Wave 5/7, WORKTREE-DOCS-001 | L1/L2/L3 terminology + state guard + `moai worktree new --base` + autonomous policy |
| `worktree/examples.md` | MINOR | BODP, CI-AUTONOMY-001 W7 | `--base` flag 예시 |
| `worktree/faq.md` | MINOR | WORKTREE-DOCS-001 | autonomous policy 2026-05-17 |
| `utility-commands/moai.md` | MAJOR | WF-004 | Agentless fixed-pipeline classification + 5+4 subcommand 매트릭스 |
| `utility-commands/moai-loop.md` | MINOR | WF-003 | `--mode` flag |
| `utility-commands/moai-mx.md` | MAJOR | SPC-004, MX TAG v2 | `moai mx query` CLI + HookSpecificOutput.MxTags |
| `utility-commands/moai-github.md` | MAJOR | LATE-BRANCH-001, CI-AUTONOMY-001 W2/W3 | PR watch loop + auto-fix + no auto Issue |
| `quality-commands/moai-review.md` | MINOR | (evaluator-active 통합) | review pipeline 갱신 |
| `design/*` | MAJOR | DESIGN-PIPELINE-001, DESIGN-FOLDER-FIX-001 | 6-wave hybrid pipeline + reserved collision warning |

### 3.3 P2 — Medium (신규 페이지 / 보조 가이드)

| 페이지 | 유형 | 영향 SPEC | 변경 요약 |
|--------|------|-----------|-----------|
| **NEW**: `advanced/security-notes.md` | NEW | SECURITY-CRIT-001 | CWE-732/214/345 사용자 가시 동작 변화 요약 + 권장 점검 절차 |
| **NEW**: `contributing/language-vs-skill.md` | NEW | WF-005 | Language rules vs skills boundary 명문화 (16개 언어 중립성 + skill scope) |
| **NEW**: `advanced/agent-lint.md` | NEW | ORC-002 | `moai agent lint` 8-rule engine 사용법 + pre-commit YAML snippet |
| **NEW**: `advanced/spec-lint.md` | NEW | SPC-003, SPECLINT-DEBT-001/002 | `moai spec lint` CLI 사용법 + 12-field canonical schema |
| **NEW**: `advanced/constitution-validate.md` | NEW | CONSTITUTION-DUAL-001 | `moai constitution validate` + zone-registry + 9 sentinel keys |
| **NEW**: `advanced/harness-cli.md` | NEW | HRN-001, HARNESS-AUTONOMY-001 | 10 CLI verbs 사용법 (route/validate/status/apply/rollback/disable/mute/mute-list/unmute/verify) |

### 3.4 P3 — Low (Optional, dev-only 검토)

| 페이지 | 결정 필요 |
|--------|----------|
| Karpathy quickref / anti-patterns | dev-only (`.claude/rules/`)인지 사용자 문서 대상인지 결정 필요. 추천: 사용자 노출 (`core-concepts/trust-5.md`에 sectioning) |
| AskUserQuestion §19 Enforcement Protocol | dev-only (CLAUDE.local.md §19). 사용자 문서 노출 불필요 — 검토 결과 dev-only 유지 권장 |
| L2/L3 worktree autonomous policy 2026-05-17 | `worktree/faq.md` opt-in 정책 명시 |

---

## 4. SECURITY-CRIT-001 (P0 보안) 섹션 설계

### 4.1 영향 받은 사용자 가시 동작

| CWE | 변경 |
|-----|------|
| **CWE-732/552** (Insecure Permission / Sensitive Info Insertion) | `settings.local.json` 신규 생성/업데이트 시 `0o600` 권한 강제. moai 명령(예: `moai glm`, `moai cc`, `moai cg`)이 token/credential 쓰기 전후 권한 검증 |
| **CWE-214** (Sensitive Info in Process Arguments) | tmux 패널 생성 시 GLM/Z.AI 인증 환경변수를 argv로 노출하지 않고 source-file을 통해 주입 (`~/.moai/.env.*` 경로) |
| **CWE-345** (Insufficient Verification of Authenticity) | `moai update` 시 checksum 강제 검증 + 1회 retry. 검증 실패 시 작업 중단 |

### 4.2 docs-site 반영 방식 (권장안 A)

**옵션 A (권장)**: 신규 페이지 `advanced/security-notes.md` 생성 + 기존 페이지 cross-reference.

- `advanced/security-notes.md` 구조:
  1. **Why this matters**: v2.20.0-rc1 release-blocker 3건 해소 배경
  2. **CWE-732/552**: settings.local.json 권한 모델 + 사용자 점검 절차 (`ls -l ~/.claude/settings.local.json` 권장)
  3. **CWE-214**: tmux 세션 환경변수 전달 방식 변화 + `~/.moai/.env.*` 파일 권장 권한 (`0o600`)
  4. **CWE-345**: `moai update` checksum 검증 동작 + 실패 시 복구 방법
  5. **점검 체크리스트** (사용자 액션 아이템)
  6. **CHANGELOG 링크** (v2.20.0-rc1)

- 기존 페이지 cross-reference:
  - `advanced/settings-json.md`에 settings.local.json 0o600 권고 추가
  - `getting-started/update.md`에 checksum 검증 절차 추가
  - `multi-llm/cg-mode.md`에 tmux env source-file 방식 명시

**옵션 B (대안)**: 신규 페이지 없이 기존 페이지에 분산 — 사용자가 보안 변화를 한눈에 찾기 어려움. 권장하지 않음.

### 4.3 우선순위

- P0 release-blocker 해소 SPEC이므로 **이번 docs-site 업데이트의 첫 번째 작업**으로 권장.
- 4-locale 동시 작성 의무 (§17.3) — ko canonical 우선 작성 후 en/ja/zh 동일 PR에 포함.

---

## 5. 4-locale 동기화 체크리스트

§17.3 정책 준수. 각 페이지마다 다음을 모두 확인:

```
[ ] ko 본문 작성 완료
[ ] en 본문 작성 완료 (ko → en 1:1)
[ ] ja 본문 작성 완료 (en → ja)
[ ] zh 본문 작성 완료 (en → zh)
[ ] 4개 locale frontmatter (title/description/weight) 대응
[ ] Mermaid 다이어그램: TD only, 노드 라벨 번역 (구문/방향 미변경)
[ ] Anchor 링크: 번역된 heading slug 가리킴
[ ] 금지 URL 재도입 없음 (docs.moai-ai.dev / adk.moai.com / adk.moai.kr)
[ ] 본문 이모지 유입 없음
[ ] CHANGELOG.md 반영
[ ] 버전 스냅샷 필요 여부 판단 (§17.4 — Patch만이면 NO)
[ ] `scripts/docs-i18n-check.sh` 통과
```

### 5.1 대규모 콘텐츠 예외 적용 여부

§17.3 예외 규칙(5,000 단어 이상은 ko 머지 후 en 48h / zh+ja 72h)을 적용할 후보:
- `advanced/settings-json.md` (현재 32KB → 갱신 후 추정 +5K~+10K word)
- `core-concepts/harness-engineering.md` (현재 6KB → 갱신 후 추정 +8K~+12K word 가능)

→ **결정**: 이번 작업은 ko-first + en/ja/zh 동일 PR 정책을 유지하되, 페이지별 word count 측정 후 5,000 단어 초과 시 `translation_status: pending` 처리 후 후속 PR 분리.

---

## 6. 단계별 실행 계획

### Phase 0 — 사전 합의 (현재 단계)

- [x] research doc 작성 (본 문서)
- [ ] 사용자 검토 + 우선순위 확정
- [ ] 후속 작업 분기 방식 결정 (단일 SPEC vs 다중 SPEC)

### Phase 1 — Security Foundation (가장 시급)

- 페이지: `advanced/security-notes.md` (신규) + `advanced/settings-json.md` (0o600 추가) + `getting-started/update.md` (checksum) + `multi-llm/cg-mode.md` (tmux env)
- 4-locale 동시
- 추정 분량: ko 3K~5K word
- 의존성 없음

### Phase 2 — Harness Autonomy Documentation (큰 변화)

- 페이지: `core-concepts/harness-engineering.md` + `advanced/harness-profiles.md` + (신규) `advanced/harness-cli.md`
- HARNESS-AUTONOMY-001 (W3) + HRN-001/002/003 + HARNESS-LEARNING-001 + PROJECT-HARNESS-001 + moai-meta-harness
- 4-locale 동시
- 추정 분량: ko 8K~12K word (대규모 — §17.3 예외 검토)
- 의존성: Phase 1 무관

### Phase 3 — Workflow Discipline (Tier S/M/L + Late-Branch)

- 페이지: `core-concepts/spec-based-dev.md` + `workflow-commands/moai-plan.md` + `workflow-commands/moai-run.md` + `workflow-commands/moai-sync.md` + (신규) `advanced/spec-lint.md`
- WORKFLOW-LEAN-001 + WORKFLOW-OPT-001 + LATE-BRANCH-001 + WORKFLOW-SPLIT-001 + SPC-001/003
- 4-locale 동시
- 추정 분량: ko 5K~8K word
- 의존성: Phase 2 (harness 일부 cross-ref)

### Phase 4 — Agents/Skills/Catalog (`moai agent lint` 등)

- 페이지: `advanced/agent-guide.md` + `advanced/skill-guide.md` + `advanced/catalog-system.md` + (신규) `advanced/agent-lint.md`
- CATALOG-001/002 + ORC-001~005 + CORE-SLIM-001/B-001 + moai-meta-harness
- 4-locale 동시
- 추정 분량: ko 5K~8K word
- 의존성: Phase 3 (sub-skill split + Tier 설정 cross-ref)

### Phase 5 — Constitution + Trust + Memory

- 페이지: `core-concepts/constitution.md` + (신규) `advanced/constitution-validate.md` + `core-concepts/trust-5.md` + `core-concepts/moai-memory.md` + `core-concepts/what-is-moai-adk.md`
- CONSTITUTION-DUAL-001 + CLAUDE-REFRESH-001 + KARPATHY-001
- 4-locale 동시
- 추정 분량: ko 4K~6K word
- 의존성: Phase 3 (Tier 설정 cross-ref), Phase 4 (Agent 동작 변화 cross-ref)

### Phase 6 — Settings / Hooks / Permission / Sandbox

- 페이지: `advanced/settings-json.md` (대폭) + `advanced/hooks-guide.md` + `advanced/hooks-reference.md` + `getting-started/cli.md` (`moai doctor permission/sandbox/hook` 3 신규)
- RT-002/003/004/005/006/007 + MIG-002/003 + HOOK-HARDEN-001 + CC2122-HOOK-002
- 4-locale 동시
- 추정 분량: ko 10K~15K word (대규모 — §17.3 예외 적용 후보)
- 의존성: Phase 1 (security cross-ref), Phase 5 (CLAUDE.md cross-ref)

### Phase 7 — Multi-LLM + MCP + Statusline + CC 호환성

- 페이지: `advanced/mcp-servers.md` + `advanced/statusline.md` (VERIFY) + `multi-llm/cg-mode.md` + `multi-llm/model-policy.md` + `guides/multi-llm-ci.md`
- GLM-MCP-001 + CI-MULTI-LLM-001 + Claude Code v2.1.119~145 호환성 + STATUSLINE-V2145-001
- 4-locale 동시
- 추정 분량: ko 4K~6K word
- 의존성: Phase 1 (보안 cross-ref)

### Phase 8 — CI Autonomy + Worktree + BODP

- 페이지: `guides/ci-autonomy.md` (대폭) + `worktree/guide.md` + `worktree/examples.md` + `worktree/faq.md`
- CI-AUTONOMY-001 Wave 1~7 + CI-FASTTRACK-001 + WORKTREE-DOCS-001
- 4-locale 동시
- 추정 분량: ko 6K~10K word
- 의존성: Phase 3 (workflow command cross-ref)

### Phase 9 — SPEC Tooling (Brain / Design / MX / Status)

- 페이지: `workflow-commands/moai-brain.md` + `workflow-commands/moai-design.md` + `design/*` (5 페이지) + `utility-commands/moai-mx.md` + `quality-commands/moai-review.md`
- BRAIN-001 + DESIGN-PIPELINE-001 + SPC-004 (MX) + STATUS-LIFECYCLE-001 + STATUS-AUTO-001
- 4-locale 동시
- 추정 분량: ko 6K~10K word
- 의존성: Phase 3 + Phase 4

### Phase 10 — Init/CLI/Wizard/Update

- 페이지: `getting-started/cli.md` + `getting-started/init-wizard.md` + `getting-started/update.md` + `getting-started/quickstart.md`
- TUI-001 M1~M7 + UPDATE-CLEANUP-001 + huh wizard
- 4-locale 동시
- 추정 분량: ko 5K~8K word
- 의존성: Phase 6 (cli verbs cross-ref)

### Phase 11 — Contributing 신설 (Language vs Skill boundary)

- 페이지: (신규) `contributing/language-vs-skill.md` + `contributing/_index.md` 확장
- WF-005
- 4-locale 동시
- 추정 분량: ko 2K~3K word
- 의존성: Phase 4

### Phase 12 — Final Pass

- `scripts/docs-i18n-check.sh` 전체 실행 → 4-locale 정합성 확인
- CHANGELOG.md `[Unreleased]` 섹션 정리
- `hugo.yaml` baseURL/og 확인
- 버전 스냅샷 결정 (만약 v3.0.0 major 출시 직전이면 `content/{locale}/v2/` 동결 복사 — §17.4)
- Vercel 빌드 cost 확인 (Elastic 머신 유지 — §20)

---

## 7. 후속 SPEC 후보 (계획 합의 후 분기 가능)

### 옵션 1: 단일 통합 SPEC (Tier L)

- SPEC-V3R5-DOCS-SITE-UPDATE-001 (가칭, Tier L)
- 11 Phase 단일 lifecycle로 처리
- 산출물: spec/plan/acceptance/design/research 5종
- 추정 wall-time: 매우 큼 (수십 시간 분량 문서 작성)
- 장점: 일관성, 4-locale 동기 단일 PR
- 단점: PR 규모 과대, 검토 부담, 부분 롤백 곤란

### 옵션 2: Phase별 분할 SPEC (Tier S 또는 M 각각)

- SPEC-V3R5-DOCS-SECURITY-001 (Tier S, Phase 1)
- SPEC-V3R5-DOCS-HARNESS-001 (Tier M, Phase 2)
- SPEC-V3R5-DOCS-WORKFLOW-001 (Tier M, Phase 3)
- SPEC-V3R5-DOCS-AGENTS-001 (Tier M, Phase 4)
- SPEC-V3R5-DOCS-CONSTITUTION-001 (Tier S, Phase 5)
- SPEC-V3R5-DOCS-SETTINGS-001 (Tier M, Phase 6)
- SPEC-V3R5-DOCS-MULTI-LLM-001 (Tier S, Phase 7)
- SPEC-V3R5-DOCS-CI-001 (Tier M, Phase 8)
- SPEC-V3R5-DOCS-SPEC-TOOLING-001 (Tier M, Phase 9)
- SPEC-V3R5-DOCS-GETTING-STARTED-001 (Tier S, Phase 10)
- SPEC-V3R5-DOCS-CONTRIBUTING-001 (Tier S, Phase 11)
- 장점: 점진적 머지, 부분 롤백 가능, 사용자 피드백 빠른 반영
- 단점: 11 PR 운영 부담, cross-ref 일관성 검증 부담

### 옵션 3: 2-단계 그룹화 (권장)

- **Wave A** (P0 즉시 필요):
  - SPEC-V3R5-DOCS-SECURITY-001 (Phase 1, Tier S)
  - SPEC-V3R5-DOCS-CLI-REFERENCE-001 (Phase 10 일부 — cli.md만, Tier S)
- **Wave B** (P0 주요 변경):
  - SPEC-V3R5-DOCS-HARNESS-WORKFLOW-001 (Phase 2 + 3, Tier M)
  - SPEC-V3R5-DOCS-SETTINGS-HOOKS-001 (Phase 6, Tier M)
- **Wave C** (P1 보조):
  - SPEC-V3R5-DOCS-AGENTS-SKILLS-001 (Phase 4 + 5, Tier M)
  - SPEC-V3R5-DOCS-CI-WORKTREE-001 (Phase 8, Tier M)
- **Wave D** (P1 후속):
  - SPEC-V3R5-DOCS-MULTI-LLM-001 (Phase 7, Tier S)
  - SPEC-V3R5-DOCS-SPEC-TOOLING-001 (Phase 9, Tier M)
- **Wave E** (P2 신설/마무리):
  - SPEC-V3R5-DOCS-GETTING-STARTED-001 (Phase 10 잔여 + Phase 11, Tier S)
  - SPEC-V3R5-DOCS-FINAL-PASS-001 (Phase 12, Tier S)

총 10 SPEC, Wave 단위 PR. 추천.

---

## 8. 결정 사항 (2026-05-20 사용자 합의 완료)

사용자 답변으로 다음 5건 모두 확정됨:

1. **분기 옵션** (§7): **Wave A~E 그룹화 10 SPEC** 채택. 단계적 머지 + 부분 롤백 가능. PR 운영 부담 < 단일 Tier L SPEC < Phase별 11 SPEC.
2. **dev-only vs 사용자 문서 노출**: (사용자 별도 결정 없음 — 본 research doc 권장 적용 가능)
   - Karpathy 4 principles → `core-concepts/trust-5.md` 노출 권장
   - AskUserQuestion Enforcement Protocol §19 → dev-only 유지 권장 (CLAUDE.local.md 내부 운영 규칙)
3. **신규 페이지 6건 (§3.3) 신설**: **6개 전부 신설** 확정.
   - `advanced/security-notes.md` (Wave A)
   - `advanced/agent-lint.md` (Wave C)
   - `advanced/spec-lint.md` (Wave D)
   - `advanced/constitution-validate.md` (Wave C)
   - `advanced/harness-cli.md` (Wave B)
   - `contributing/language-vs-skill.md` (Wave E)
4. **버전 스냅샷 (§17.4)**: **모든 Wave 완성 후 v2.20.0-rc1 release + content/{locale}/v2/ 동결**. release 타이밍은 docs 완성도 우선.
5. **Late-branch 적용**: 본 research doc은 **main 위 직진 commit** (Late-Branch-001 정책). 각 후속 SPEC은 별도 세션 + **`--worktree` 모드 (L3)**로 진행 — 병렬 작업 격리 + 메인 세션 컨텍스트 보호.

### 8.1 Wave 진행 순서 (확정)

```
Wave A (P0 즉시)        → Wave B (P0 주요)         → Wave C (P1 보조)
[Security + CLI Ref]   [Harness/Workflow + Settings]   [Agents + Constitution]
   ↓
Wave D (P1 후속)        → Wave E (P2 마무리)
[Multi-LLM + SPEC Tools]   [Getting Started + Contributing + Final Pass]
   ↓
v2.20.0-rc1 release tag → content/{locale}/v2/ 동결 (§17.4 minor snapshot)
```

### 8.2 각 Wave SPEC 식별자 (확정)

| Wave | SPEC | Tier | 포함 Phase | 신규 페이지 |
|------|------|------|------------|-------------|
| **A** | SPEC-V3R5-DOCS-SECURITY-001 | S | Phase 1 (4 페이지) | security-notes.md |
| **A** | SPEC-V3R5-DOCS-CLI-REFERENCE-001 | S | Phase 10 일부 (cli.md만) | — |
| **B** | SPEC-V3R5-DOCS-HARNESS-WORKFLOW-001 | M | Phase 2 + 3 (8 페이지) | harness-cli.md |
| **B** | SPEC-V3R5-DOCS-SETTINGS-HOOKS-001 | M | Phase 6 (4 페이지) | — |
| **C** | SPEC-V3R5-DOCS-AGENTS-SKILLS-001 | M | Phase 4 + 5 (7 페이지) | agent-lint.md, constitution-validate.md |
| **C** | SPEC-V3R5-DOCS-CI-WORKTREE-001 | M | Phase 8 (4 페이지) | — |
| **D** | SPEC-V3R5-DOCS-MULTI-LLM-001 | S | Phase 7 (5 페이지) | — |
| **D** | SPEC-V3R5-DOCS-SPEC-TOOLING-001 | M | Phase 9 (8 페이지) | spec-lint.md |
| **E** | SPEC-V3R5-DOCS-GETTING-STARTED-001 | S | Phase 10 잔여 + Phase 11 (5 페이지) | language-vs-skill.md |
| **E** | SPEC-V3R5-DOCS-FINAL-PASS-001 | S | Phase 12 | — |

### 8.3 세션 운영 정책

- 본 research doc commit은 **이번 세션** (main 직진)
- 각 Wave SPEC plan/run/sync는 **별도 세션** (`--worktree` 모드, L3 isolation)
- 세션 시작 시 본 research doc 경로를 paste-ready resume message에 포함:
  - `.moai/research/docs-site-v2.14-to-HEAD-update-plan-2026-05-20.md`
- 4-locale 동기화는 각 Wave SPEC 내부에서 처리 (단일 PR)

---

## 9. Risks

| Risk | Likelihood | Impact | Mitigation |
|------|------------|--------|-----------|
| R1: 4-locale ja/zh 번역 품질 격차 | High | Medium | en → ja/zh 동기 작성 + glossary 준수 |
| R2: Mermaid 다이어그램 LR/TD 일관성 위반 | Medium | Low | §17.2 [HARD] TD only — `scripts/docs-i18n-check.sh`로 자동 탐지 |
| R3: 금지 URL 재도입 | Low | High | CI check + grep guard |
| R4: 페이지 word count 5K 초과 → §17.3 예외 발생 | High (Phase 2/6) | Medium | `translation_status: pending` 표시 + 48h/72h 후속 PR |
| R5: 신규 6 페이지 분류 (advanced vs core-concepts vs reference) 불일치 | Medium | Low | 본 research doc §3.3 매트릭스 합의 후 고정 |
| R6: cross-reference 파괴 (heading slug 번역 불일치) | High | Medium | 번역 시 heading slug 일치 자동 확인 + manual review |
| R7: SPEC commit이 docs-site 갱신 없이 머지된 상태 (이미 누적된 잡음) | Already realized | Medium | 본 작업이 그 정리 |
| R8: v2.20.0-rc1 release 타이밍 vs docs 완성도 mismatch | Medium | High | Wave A (Security + CLI Reference)만 release 전 완성, 나머지는 release 직후 |
| R9: Hugo build 시간 폭증 (Vercel cost) | Low | Medium | §20 Elastic 머신 유지 확인 |

---

## 10. Acceptance Criteria (본 research doc 합의 — 완료)

다음 모두 사용자 확정 완료 (2026-05-20):

- [x] §7 옵션 합의 (**Wave A~E 그룹화 10 SPEC** 채택)
- [x] §8.3 신규 6 페이지 신설 여부 (**6개 전부 신설** 확정)
- [x] §8.4 버전 스냅샷 타이밍 (**모든 Wave 완성 후 v2.20.0-rc1 release**)
- [x] §8.5 Late-branch 적용 확인 (main 직진 commit + 후속 SPEC `--worktree` 별도 세션)
- [x] 본 research doc commit + push → 후속 Wave SPEC plan은 별도 세션에서 진입

---

## 11. 다음 액션 (확정)

**채택**: 옵션 Q — 본 research doc commit + 후속 Wave 단위 SPEC을 별도 세션에서 진행. 단, 각 Wave SPEC plan은 **`--worktree` (L3) 모드**로 격리.

### 11.1 다음 세션 paste-ready resume (Wave A 진입용)

```text
ultrathink. SPEC-V3R5-DOCS-SECURITY-001 plan-phase 진입.
applied lessons: project_docs_site_update_plan, lessons #9 wave-split, lessons #14 worktree single-vs-multi-session.

전제 검증:
0) git rev-parse --show-toplevel → ~/.moai/worktrees/moai-adk-go/SPEC-V3R5-DOCS-SECURITY-001 (★ critical L3 worktree)
1) git log --oneline -3 → research doc commit 확인 (.moai/research/docs-site-v2.14-to-HEAD-update-plan-2026-05-20.md)
2) ls .moai/research/docs-site-v2.14-to-HEAD-update-plan-2026-05-20.md → 존재 확인
3) git branch --show-current → plan/SPEC-V3R5-DOCS-SECURITY-001 또는 worktree 자동 branch

실행: /moai plan SPEC-V3R5-DOCS-SECURITY-001 --worktree

머지 후: Wave A 잔여 → SPEC-V3R5-DOCS-CLI-REFERENCE-001 plan
```

### 11.2 Wave 진행 시 공통 prompt 구조

각 Wave SPEC plan-phase 시 manager-spec 위임 prompt에 다음 cross-reference 포함 필수:

- Research doc 경로: `.moai/research/docs-site-v2.14-to-HEAD-update-plan-2026-05-20.md`
- 본 doc §3.x 매트릭스 참조 (해당 Wave에 포함된 페이지 enumeration)
- 4-locale 정책: `.moai/docs/docs-site-i18n-rules.md` §17.3
- 금지 URL 가드: `.moai/docs/docs-site-i18n-rules.md` §17.1

---

## 12. 부록: v2.14.0..HEAD User-Facing Commit Index

전체 ~80 commit (feat/fix 기준). 그룹별 매핑은 §2 참조. 본 부록은 trace 용도.

(생략 — `git log v2.14.0..HEAD --format='%h %s' --no-merges | grep -E '^[a-f0-9]+ (feat|fix|hotfix|breaking|release)'`로 재현 가능)

핵심 milestone commit:

| Commit | SPEC / 변화 |
|--------|-------------|
| `03a2552a2` | SECURITY-CRIT-001 4 modules — P0 보안 3건 정정 |
| `0d7debf19` | CORE-SLIM-B-001 — Category B dead-weight retire |
| `c0eb30da6` | WORKFLOW-LEAN-001 — Tier S/M/L + Section A-E |
| `664cd6eae` | LATE-BRANCH-001 — Late-branch workflow + no auto GH Issue |
| `fb3d1e22b` | STATUSLINE-V2145-001 — Statusline v2.1.145 |
| `bae98ce19` | HARNESS-AUTONOMY-001 W3 run — 4-Tier + 5-Layer + 18 sentinels + 10 CLI verbs |
| `81d42a1ae` | CONSTITUTION-DUAL-001 W1 — Dual-Zone + validate CLI |
| `fc31b30b4` | CLAUDE-REFRESH-001 W0 — Architecture Truth |
| `f66a5764b` | MIG-003 — Config Loader Completeness |
| `e6074ad36` | RT-003 — Sandbox Execution Layer |
| `477e30611` | RT-002 — 8-tier Permission Stack |
| `a165706e6` | RT-006 — Hook Handler 27-Event Coverage |
| `8c53ce860` | ORC-002 — `moai agent lint` |
| `985~986~880` | WORKFLOW-SPLIT-001 Wave 1~4 — sub-skill split |
| `c45cf4814` | BRAIN-001 — `/moai brain` 7-phase ideation |
| `3da4e33c1` | DESIGN-PIPELINE-001 — Hybrid 6-wave |
| `68f023289` | HARNESS-LEARNING-001 — Self-Learning Dynamic Harness |
| `9666c03fe` | GLM-MCP-001 — Z.AI MCP server |
| `dca57b14d~085efe76f` | TUI-001 M1~M7 — huh wizard + DDD migration |
| `a715fcbaa` | UPDATE-CLEANUP-001 — 멱등 배포 |
| `7e098ee37` | KARPATHY-001 — 4 principles |
| `b47101779` | CON-001 + EXT-001 restore |
| `0caf60d41` | v2.16.0 release — Pattern Cookbook + V3R2 restore + Phase A |
| `48176482a` | AskUserQuestion Enforcement Protocol §19 신설 |

---

**작성일**: 2026-05-20
**작성자**: MoAI orchestrator (사용자 합의 4-답안: research doc / user-facing / 보안 포함 / 4-locale 동일 PR)
**상태**: APPROVED (사용자 결정 5건 완료, §8 참조)
**다음 세션**: Wave A SPEC-V3R5-DOCS-SECURITY-001 plan-phase 진입 (`--worktree` 모드)

# research.md — SPEC-CC2178-DOCS-ALIGN-001

> **Falsifiable evidence layer** for CC 2.1.169→2.1.178 Tier 1 docs-only 정합화. 각 Tier 1 항목을 CC 공식 문서에 대조하여 (a) 기능 존재, (b) 도입 버전, (c) MoAI 현재 wiring 상태, (d) 문서화 gap을 검증한다. 이 파일은 plan-auditor의 독립 검증 근거다.

## §1 — Evidence Sources (CC official docs)

본 research.md가 인용하는 CC 공식 문서 surface (falsifiable — plan-auditor가 URL을 독립 접근하여 검증 가능):

| Feature category | Canonical CC docs URL |
|------------------|----------------------|
| Permissions / settings | `https://docs.claude.com/en/docs/claude-code/settings` + `https://docs.claude.com/en/docs/claude-code/iam` (permissions) |
| Hooks | `https://docs.claude.com/en/docs/claude-code/hooks` + `https://docs.claude.com/en/docs/claude-code/hooks-guide` |
| Subagents | `https://docs.claude.com/en/docs/claude-code/sub-agents` |
| Skills | `https://docs.claude.com/en/docs/claude-code/skills` |
| Slash commands / `/cd` | `https://docs.claude.com/en/docs/claude-code/slash-commands` |
| CC changelog (버전 검증) | `https://github.com/anthropics/claude-code/blob/main/CHANGELOG.md` (raw: `https://raw.githubusercontent.com/anthropics/claude-code/main/CHANGELOG.md`) |

**Source research doc**: `.moai/research/cc-update-2.1.163-to-2.1.178.md` (2026-06-16 analysis pass, verbatim changelog fetch).

## §2 — Per-Item Evidence Matrix

각 Tier 1 docs-only 항목에 대해: CC 도입 버전(changelog 검증) + CC 공식 문서 위치 + MoAI 현재 상태(grep 검증) + 문서화 gap.

### Item 1 — `Tool(param:value)` permission-rule syntax with `*` wildcard

| Attribute | Value | Evidence |
|-----------|-------|----------|
| **CC version** | 2.1.178 | `.moai/research/cc-update-2.1.163-to-2.1.178.md` L33: "`Tool(param:value)` permission-rule syntax with `*` wildcard — 2.1.178" |
| **CC docs location** | `docs.claude.com/en/docs/claude-code/iam` (permissions reference) + `settings` 페이지 | Tier 1 테이블 L33 매핑 |
| **MoAI current state** | `Tool(param` → rules 파일 0 매치 (2026-06-16 grep); `Tool(specifier)` 기존 형식은 `settings-json.md` L362에 존재 | 본 SPEC plan-phase §B 검증 |
| **Doc gap** | param-scoped wildcard 형식 미문서화 | REQ-DA-001 |
| **MoAI wiring** | MoAI는 현재 param-scoped 규칙을 emit하지 않음 — "옵션으로 존재" framing | REQ-DA-001 accuracy note |

**Falsification**: plan-auditor는 `docs.claude.com/en/docs/claude-code/iam`에서 `Tool(param:value)` 형식을 검색하여 존재를 확인. changelog `2.1.178` 항목에서 해당 bullet을 확인.

---

### Item 2 — Nested `.claude/` closest-wins on name collision (agent/workflow/output-style)

| Attribute | Value | Evidence |
|-----------|-------|----------|
| **CC version** | 2.1.178 | research doc L35: "Nested `.claude/` closest-wins on name collision (agent/workflow/output-style) — 2.1.178" |
| **CC docs location** | `docs.claude.com/en/docs/claude-code/sub-agents` (agent precedence) + `skills` (skill discovery) | Tier 1 테이블 L34-35 |
| **MoAI current state** | `closest-wins` → rules 파일 0 매치 | grep 검증 |
| **Doc gap** | closest-directory-wins 선행 규칙 미문서화 | REQ-DA-002 |
| **MoAI wiring** | N/A (CC 런타임 동작; MoAI는 문서화만) | — |

**Falsification**: plan-auditor는 CC subagents/skills 문서에서 "closest" 또는 "nested" 선행 규칙을 검색.

---

### Item 3 — Nested `.claude/skills` loading precedence

| Attribute | Value | Evidence |
|-----------|-------|----------|
| **CC version** | 2.1.178 | research doc L34: "Nested `.claude/skills` directories now load when working on files there — 2.1.178" |
| **CC docs location** | `docs.claude.com/en/docs/claude-code/skills` (skill scope/discovery 섹션) | Tier 1 테이블 L34 |
| **MoAI current state** | rules 파일 관련 노트 0매치; `skill-authoring.md`에 discovery 절은 있으나 nested 로딩 명시 없음 | §B 검증 |
| **Doc gap** | nested skills 디렉터리 로딩 시맨틱 미문서화 | REQ-DA-003 |

**Note**: Item 2와 Item 3는 같은 CC 2.1.178 릴리스의 상관 기능(nested `.claude/` 동작)이나 선행 규칙(closest-wins)과 로딩 시맨틱(load-on-work)은 구분됨 — 별도 REQ.

---

### Item 4 — `disableBundledSkills` setting + env var

| Attribute | Value | Evidence |
|-----------|-------|----------|
| **CC version** | 2.1.169 | research doc L39: "`disableBundledSkills` setting + env var (hide bundled skills/workflows) — 2.1.169" |
| **CC docs location** | `docs.claude.com/en/docs/claude-code/settings` (settings reference) + `skills` (bundled skills 비활성화) | Tier 1 테이블 L39 |
| **MoAI current state** | `disableBundledSkills` → rules 파일 0 매치; docs-site에는 `commands.md`/`goal.md`/`interactive-mode.md`에 비슷한 용어가 있으나 설정 키로는 미문서화 | grep 검증 |
| **Doc gap** | 설정 키 + 환경변수로서의 `disableBundledSkills` 미문서화 | REQ-DA-004 |
| **MoAI wiring** | MoAI는 이 설정을 emit하지 않음 — "옵션으로 존재" framing | REQ-DA-004 accuracy note |

**Falsification**: CC settings 문서에서 `disableBundledSkills` 키를 검색하여 존재 + 기본값 확인.

---

### Item 5 — `--safe-mode` flag

| Attribute | Value | Evidence |
|-----------|-------|----------|
| **CC version** | 2.1.169 | research doc L40: "`--safe-mode` flag (start with all customizations disabled) — 2.1.169" |
| **CC docs location** | `docs.claude.com/en/docs/claude-code/settings` (또는 troubleshooting) | Tier 1 테이블 L40 |
| **MoAI current state** | `safe-mode` → rules 파일 0 매치 | grep 검증 |
| **Doc gap** | 디버깅 플래그 미문서화 | REQ-DA-004 (settings-management.md + settings-json.md에 `disableBundledSkills`와 함께 노트) |
| **MoAI wiring** | N/A (CC CLI 플래그; MoAI는 문서화만) | — |

**Note**: `--safe-mode`는 REQ-DA-004에 `disableBundledSkills`와 함께 포함되었다 (둘 다 CC 2.1.169 governance/debug 토글이며 같은 rules 절에 자연스럽게 배치). 별도 REQ로 분리하면 9번째 REQ가 되나, 항목 수를 8로 유지하기 위해 통합.

---

### Item 6 — `post-session` lifecycle hook

| Attribute | Value | Evidence |
|-----------|-------|----------|
| **CC version** | 2.1.169 | research doc L41: "`post-session` lifecycle hook (self-hosted runner, after session ends) — 2.1.169" |
| **CC docs location** | `docs.claude.com/en/docs/claude-code/hooks` (hook events 목록) + `hooks-guide` | Tier 1 테이블 L41 |
| **MoAI current state** | `post-session` → rules 파일 0 매치; `hooks-system.md`에 SessionStart/Stop/PreToolUse/PostToolUse/SubagentStart/TaskCompleted 등은 있으나 post-session 부재 | §B 검증 |
| **Doc gap** | post-session 라이프사이클 이벤트 미문서화 | REQ-DA-005 |
| **MoAI wiring** | **MoAI는 현재 post-session 훅을 wiring하지 않음** — accuracy-over-completeness 핵심 케이스. "존재와 옵션"으로만 문서화. | REQ-DA-005 accuracy note (C5) |

**Falsification**: CC hooks 문서에서 `post-session` 이벤트를 검색하여 존재 + 발화 시점(세션 종료 후) 확인.

---

### Item 7 — auto-mode pre-launch classifier

| Attribute | Value | Evidence |
|-----------|-------|----------|
| **CC version** | 2.1.178 | research doc L36: "Improved auto mode: subagent spawns evaluated by classifier before launch — 2.1.178" |
| **CC docs location** | CC permission modes / auto-mode 문서 (정확 URL은 run-phase가 WebFetch로 확인) | Tier 1 테이블 L36 |
| **MoAI current state** | `pre-launch classifier` → rules 파일 0 매치 | grep 검증 |
| **Doc gap** | auto-mode의 classifier 게이트 미문서화; `/goal` unattended 루프와의 상보성 미연결 | REQ-DA-007 |
| **MoAI wiring** | MoAI auto-mode는 `orchestration-mode-selection.md`에 문서화되어 있으나 classifier 세부는 부재 | — |

---

### Item 8 — subagent `disallowedTools` MCP enforcement

| Attribute | Value | Evidence |
|-----------|-------|----------|
| **CC version** | 2.1.178 | research doc L42: "MCP server-level specs in subagent `disallowedTools` now honored (was silently ignored) — 2.1.178" |
| **CC docs location** | `docs.claude.com/en/docs/claude-code/sub-agents` (disallowedTools semantics) | Tier 1 테이블 L42 |
| **MoAI current state** | `disallowedTools`는 `agent-authoring.md`에 존재하나 MCP 강제 semantics 미문서화 | §B 검증 |
| **Doc gap** | CC 2.1.178 behavior fix (silent-ignore → enforce) 미반영 | REQ-DA-006 |
| **MoAI wiring** | run-phase가 MoAI agent가 이전 silent-ignore 동작에 의존하지 않았음을 grep 확인 (예상: 의존 없음) | REQ-DA-006 |

**Falsification**: CC subagents 문서에서 `disallowedTools` + MCP server-level 강제를 검색.

---

### Item 9 — `/cd` command (prompt-cache-preserving cwd move)

| Attribute | Value | Evidence |
|-----------|-------|----------|
| **CC version** | 2.1.169 | research doc L48 (Tier 2): "`/cd` command — move session cwd without breaking prompt cache — 2.1.169"; C/Q/S-4 섹션 L125-134 상세 |
| **CC docs location** | `docs.claude.com/en/docs/claude-code/slash-commands` (또는 sessions) | research doc L48 |
| **MoAI current state** | `/cd ` → rules 파일 0 매치; `session-handoff.md` Block 0는 new-terminal 경로만 문서화 | §B 검증 |
| **Doc gap** | `/cd` cache-preserving 대안 미문서화; Block 0 new-terminal cold-start만 있음 | REQ-DA-008 |
| **MoAI wiring** | MoAI prompt caching은 `internal/runtime/cache_control.go` (SPEC-V3R6-PROMPT-CACHE-001)에 구현되어 있으나, `/cd`는 CC 런타임 기능 — MoAI는 Block 0 노트로 보완 | REQ-DA-008 |

**Cost/speed relevance** (research doc C/Q/S-4, L125-134): `/cd`는 worktree 재개 시 캐시된 시스템 프롬프트를 warm 상태로 유지하는 반면, 새 터미널(Block 0)은 cold-start. worktree-anchored resume의 speed + cost win. `[1m]` constraint와 직교(모델 pinning 무관).

**Falsification**: CC slash-commands 문서에서 `/cd` 명령을 검색하여 캐시 보존 시맨틱 확인.

---

## §3 — MoAI Current-State Audit (grep evidence, 2026-06-16)

본 research.md의 "MoAI current state" 열은 2026-06-16 plan-phase grep 검증 결과다 (검증 범위: `.claude/rules/moai/` 하위 규칙 파일만):

```
Tool(param       → 0 rules files (전부 신규 문서화 대상)
closest-wins     → 0 rules files
disableBundledSkills → 0 rules files
safe-mode        → 0 rules files
post-session     → 0 rules files
/cd              → 0 rules files
pre-launch classifier → 0 rules files
```

> **검증 범위 명시 (D1 정정)**: 위 카운트는 `.claude/rules/moai/` 하위로 한정한다. `/cd`는 rules 파일에는 0매치이나, `.claude/skills/` 하위 skill 파일 3개(`moai-foundation-core/modules/agents-reference.md`, `moai-workflow-spec/references/examples.md`, `moai-foundation-quality/SKILL.md`)에 부분적으로 등장한다 (이들은 rules가 아닌 skills 영역이므로 본 audit 범위 밖). REQ-DA-008의 대상은 `session-handoff.md`(rules) + docs-site이므로, rules 범위 0매치가 유효한 "미문서화" 신호다. `.claude/worktrees/` 하위 매치는 stale worktree 사본이므로 무시.

**docs-site 기존 부분 문서화** (확장 필요):
- `docs-site/content/en/advanced/settings-json.md` L362: `Tool(specifier)` 기존 형식 존재 → `Tool(param:value)` wildcard는 신규 추가
- `docs-site/content/*/claude-code/foundations/commands.md` 등: `disableBundledSkills` 유사 용어가 일부 locale에 존재하나 설정 키로는 미문서화 → REQ-DA-004에서 정식 문서화

## §4 — Verdict Summary

| Item | CC version | Doc gap severity | MoAI wiring | REQ |
|------|-----------|------------------|-------------|-----|
| 1. `Tool(param:value)` wildcard | 2.1.178 | MUST (high) | emit 안 함 (옵션 노트) | REQ-DA-001 |
| 2. closest-wins precedence | 2.1.178 | MUST (high) | N/A (CC 동작) | REQ-DA-002 |
| 3. nested skills loading | 2.1.178 | SHOULD (med) | N/A | REQ-DA-003 |
| 4. `disableBundledSkills` + `--safe-mode` | 2.1.169 | SHOULD (med) | emit 안 함 (옵션 노트) | REQ-DA-004 |
| 5. `post-session` hook | 2.1.169 | SHOULD (med) | **wiring 안 함** (accuracy 핵심) | REQ-DA-005 |
| 6. `disallowedTools` MCP enforce | 2.1.178 | SHOULD (low) | 의존 없음 (예상) | REQ-DA-006 |
| 7. auto-mode classifier | 2.1.178 | SHOULD (low) | N/A | REQ-DA-007 |
| 8. `/cd` cache-preserving | 2.1.169 | MUST (med) | CC 기능 (Block 0 보완) | REQ-DA-008 |

**총 8 REQ, 9 Tier 1 docs-only 항목** (Item 5 `--safe-mode`는 Item 4에 통합).

## §5 — Items Considered but Deferred (Out of Scope)

본 SPEC의 exclusion(spec.md §E)과 research doc P1/P2/P3/P6 항목은 sibling/별도 SPEC으로 위임됨:

| Deferred item | Reason | Target SPEC |
|---------------|--------|-------------|
| `availableModels` / `enforceAvailableModels` allowlist 문서화 | CC 2.1.175/2.1.176 비용 레버 — cost/quality 영역 | Sibling MODEL-POLICY-REPAIR-001 P2 |
| `[1m]` model-policy constraint 재검증 | CC 2.1.172-2.1.174 fix가 실패 표면 좁혔는지 검증 | Sibling MODEL-POLICY-REPAIR-001 P3 |
| Fable 5 tier 채택 | deliberate cost-strategy 결정 (반응형 채택 금지) | 별도 전략 SPEC (P6 deferred) |
| Go 코드 변경 (전부) | 본 SPEC은 ZERO Go code | 해당 없음 (exclusion) |

## §6 — CC Version Verification (changelog cross-check)

research doc `.moai/research/cc-update-2.1.163-to-2.1.178.md` Executive Summary에 따르면, CC 2.1.164..2.1.178 창의 실제 릴리스 버전:

```
2.1.164*, 2.1.169, 2.1.170, 2.1.172, 2.1.173, 2.1.174, 2.1.175, 2.1.176, 2.1.178
(never-released gaps: 2.1.171, 2.1.177; non-itemized: 2.1.165-2.1.168)
```

본 SPEC의 Tier 1 항목 도입 버전(2.1.169, 2.1.178)은 전부 실제 릴리스된 버전에 해당 — changelog verbatim fetch로 검증됨(research doc L9-16).

**Falsification**: plan-auditor는 `https://raw.githubusercontent.com/anthropics/claude-code/main/CHANGELOG.md`에서 2.1.169 / 2.1.178 항목을 직접 확인하여 각 feature bullet의 존재를 검증.

---
spec_id: SPEC-V3R4-CC2X-ADOPT-001
title: Claude Code v2.1.0+ Adoption — Master Research (Umbrella)
status: research-only
phase: research
created: 2026-05-12
author: MoAI orchestrator + GOOS
related_prs: []
child_specs:
  - SPEC-V3R4-CC2X-HOOKS-001       # Tier 1 — Hook system overhaul
  - SPEC-V3R4-CC2X-AGENT-FM-001    # Tier 1 — Agent frontmatter expansion
  - SPEC-V3R4-CC2X-SKILL-001       # Tier 1 — Skill modernization
  - SPEC-V3R4-CC2X-MCP-001         # Tier 1 — MCP integration enhancement
  - SPEC-V3R4-CC2X-WORKTREE-001    # Tier 1 — Worktree native alignment
  - SPEC-V3R4-CC2X-SESSION-001     # Tier 1 — Session continuity
  - SPEC-V3R4-CC2X-PERMS-001       # Tier 1 — Permission model evolution
  - SPEC-V3R4-CC2X-SDK-001         # Tier 2 — SDK/headless performance
  - SPEC-V3R4-CC2X-PLUGIN-001      # Tier 2 — Plugin maturity
  - SPEC-V3R4-CC2X-STATUS-001      # Tier 2 — Statusline expansion
  - SPEC-V3R4-CC2X-TUI-001         # Tier 2 — TUI modernization
  - SPEC-V3R4-CC2X-LOOP-001        # Tier 2 — /loop + Cron scheduling
  - SPEC-V3R4-CC2X-OTEL-001        # Tier 2 — Telemetry expansion
  - SPEC-V3R4-CC2X-PLAN-MODE-001   # Tier 2 — Plan mode integration
  - SPEC-V3R4-CC2X-PLATFORM-001    # Tier 3 — Platform extensions (PowerShell/Voice/Remote)
  - SPEC-V3R4-CC2X-ENTERPRISE-001  # Tier 3 — Enterprise features
  - SPEC-V3R4-CC2X-MISC-UX-001     # Tier 3 — Misc UX
analysis_scope: Claude Code v2.1.0 (2025-10) ~ v2.1.139 (2026-05-12), 1500+ changelog entries
---

# SPEC-V3R4-CC2X-ADOPT-001 — Claude Code v2.1.0+ Adoption Master Research

## 0. Purpose

이 문서는 Claude Code v2.1.0 ~ v2.1.139 사이 1,500+ 변경사항을 전수 분석하여 moai-adk-go에 도입 가치 있는 기능을 분류·우선순위화한 **마스터 research.md**입니다. 본 문서 자체는 plan/spec/acceptance를 포함하지 않으며, 17개 child SPEC의 공통 참조 자료로 사용됩니다.

**Constitution 준수**:
- §1 HARD: 본 문서는 user-facing report로 Markdown만 사용
- §16 Self-check: SPEC 분해는 manager-spec 위임 예정 (research는 직접 작성 가능)
- §18.0 Doctrine: plan-in-main (PR #822) — research도 main에 직접 커밋

## 1. Executive Summary

**분석 결과**: 47개 도입 후보 식별

| Tier | 항목 수 | 카테고리 수 | child SPEC 매핑 |
|------|---------|-------------|------------------|
| Tier 1 (즉시 도입) | 17 | 7 | HOOKS, AGENT-FM, SKILL, MCP, WORKTREE, SESSION, PERMS |
| Tier 2 (중기 도입) | 18 | 7 | SDK, PLUGIN, STATUS, TUI, LOOP, OTEL, PLAN-MODE |
| Tier 3 (장기 검토) | 12 | 3 | PLATFORM, ENTERPRISE, MISC-UX |
| **총합** | **47** | **17** | **17 child SPECs** |

**Sprint 권고**:
- Sprint 12: Tier 1 (7 SPECs)
- Sprint 13: Tier 2 (7 SPECs)
- Sprint 14: Tier 3 (3 SPECs)

---

## 2. Tier 1 — 즉시 도입 권고 (17건, 7 SPECs)

### 2.1 SPEC-V3R4-CC2X-HOOKS-001 — Hook System Overhaul (9 items)

**Rationale**: moai-adk-go의 `internal/hook/` 패키지는 다수 핸들러를 보유하나 v2.1.49+ 신규 hook event 미적용. 도입 시 worktree 자동화·session-handoff 자동화·FROZEN zone 보호 등 핵심 워크플로우 자동화 가능.

| # | Hook Event | 버전 | moai 통합 포인트 |
|---|-----------|------|------------------|
| 1 | `WorktreeCreate` / `WorktreeRemove` | v2.1.49 | EnterWorktree/ExitWorktree → `.moai/specs/<SPEC>/` 자동 scaffold + BODP audit trail + lessons #12 (수동 isolation 위배) 자동 방지 |
| 2 | `PostCompact` | v2.1.105 | Auto-compact 시점에 progress.md 자동 영속화 → session-handoff.md `paste-ready resume` 자동 생성 |
| 3 | `ConfigChange` | v2.1.49 | FROZEN zone (constitution.md, askuser-protocol.md) 런타임 변경 감사/차단 |
| 4 | `InstructionsLoaded` | v2.1.69 | CLAUDE.md/`.claude/rules/*.md` 로드 시 최근 lessons 5건 + 활성 SPEC progress.md 자동 주입 |
| 5 | `SubagentStart` | v2.0.43 | subagent 시작 시 worktree base ref 검증, agent_id audit, `--team` 환경변수 주입 |
| 6 | `PermissionDenied` | v2.1.109 | 자동 모드 거부 시 retry 가능 안내 (`{retry: true}`) |
| 7 | Hook `if` filtering (permission-rule syntax) | v2.1.85 | spawn overhead -70% (`if: "Edit(.moai/specs/**)"` 매칭만 실행) |
| 8 | HTTP hooks (`type: "http"`) | v2.1.63 | SessionEnd → GitHub webhook → PR 상태 sync (§17 docs-i18n 워크플로우 자동화) |
| 9 | `args` exec form + `continueOnBlock` | v2.1.139 | shell 인터폴레이션 회피 (보안) + PostToolUse 피드백 |

**의존성**: `internal/hook/dispatcher.go` 라우팅 확장, `.claude/settings.json` settings.json.tmpl 신규 event 등록.

**예상 LOC**: ~1,500-2,000 (handler 9개 + dispatcher 확장 + 테스트 + 템플릿 sync)

---

### 2.2 SPEC-V3R4-CC2X-AGENT-FM-001 — Agent Frontmatter Expansion (9 items)

**Rationale**: 17개 agent .md 파일이 기본 frontmatter만 사용. v2.0.43+ 신규 필드 도입으로 declarative 정의 가능 → 코드 변경 없이 행동 정의.

| # | Frontmatter Field | 버전 | moai 활용 |
|---|-------------------|------|-----------|
| 1 | `mcpServers: [context7, sequential-thinking]` | v2.1.117 | general-purpose subagent spawn 시 MCP 자동 주입 |
| 2 | `hooks: {PreToolUse, Stop}` | v2.1.0 | agent 별 lifecycle hook (worktree 정리, lessons 기록) |
| 3 | `memory: {scope, filename}` | v2.1.33 | agent별 persistent memory ↔ `memory/lessons.md` 연동 |
| 4 | `effort: high\|xhigh\|max` | v2.1.78 | SPEC-V3R3-ORC-003 17-agent effort matrix frontmatter화 |
| 5 | `permissionMode: acceptEdits\|plan\|bypassPermissions` | v2.0.43 | worktree-isolation rule (CLAUDE.md §14) 명문화 |
| 6 | `disallowedTools: ["WebFetch", "Bash(rm:*)"]` | v2.0.30 | background-agent write 제한 declarative화 |
| 7 | `isolation: "worktree"` | v2.1.49 | moai 강제 규칙 native frontmatter 표현 |
| 8 | `background: true` | v2.1.49 | read-only researcher/analyst agent |
| 9 | `--agents` JSON 동적 spawn | v2.0.0 | workflow.yaml role_profiles → `--agents` JSON 변환 (SPEC-V3R3-RT-001) |

**의존성**: 17개 agent .md 일괄 업데이트 → **builder-harness 위임**.

**예상 LOC**: ~500-800 (agent frontmatter 일괄 수정 + 검증 도구)

---

### 2.3 SPEC-V3R4-CC2X-SKILL-001 — Skill Modernization (8 items)

**Rationale**: 17개 skill .md 파일이 기본 frontmatter만 사용. v2.1.0+ 신규 필드로 context isolation·effort tuning·user-only invocation 등 정밀 제어 가능.

| # | Field/Feature | 버전 | moai 활용 |
|---|--------------|------|-----------|
| 1 | `context: fork` | v2.1.0 | RED→GREEN→REFACTOR 각 단계 fork된 context로 → 메인 컨텍스트 오염 방지 |
| 2 | `agent: <name>` | v2.1.0 | skill을 특정 agent로 invoke (e.g., moai-workflow-tdd → manager-develop) |
| 3 | Skill hot-reload (native) | v2.1.0 | `internal/template/templates/.claude/skills/` 수정 시 재시작 불필요 → 개발 사이클 단축 |
| 4 | `${CLAUDE_SKILL_DIR}` variable | v2.1.69 | moai-domain-* skill이 자기 디렉토리 asset 참조 |
| 5 | `effort: max` | v2.1.80 | First Principles 분석 skill (moai-foundation-thinking) 자동 부스트 |
| 6 | `keep-coding-instructions: true` | v2.0.37 | output style 변경 시에도 핵심 지시 유지 |
| 7 | `disable-model-invocation: true` | v2.1.110 | /moai feedback 같은 사용자 전용 skill 모델 자동 호출 방지 |
| 8 | `skillOverrides: {visibility: ...}` | v2.1.129 | 컨텍스트 절약 + 사용자 제어 강화 |

**의존성**: 17개 skill .md 일괄 업데이트 → **builder-harness 위임**.

**예상 LOC**: ~400-600

---

### 2.4 SPEC-V3R4-CC2X-MCP-001 — MCP Integration Enhancement (8 items)

| # | Feature | 버전 | moai 활용 |
|---|---------|------|-----------|
| 1 | `alwaysLoad: true` MCP server | v2.1.121 | Sequential Thinking + Context7 ToolSearch 우회, 즉시 사용 |
| 2 | MCP Elicitation | v2.1.76 | Pencil MCP에서 design preset 선택 등 mid-task 입력 |
| 3 | `outputSchema` / `structuredContent` | v2.0.21 | design pipeline JSON contract 표준화 |
| 4 | `headersHelper` 동적 인증 | v2.0.53, v2.1.105 | SessionStart hook이 MCP 서버에 동적 헤더 주입 |
| 5 | `_meta["anthropic/maxResultSizeChars"]` | v2.1.91 | moai-domain-database schema dump 통과 (≤500K) |
| 6 | `oauth.authServerMetadataUrl` | v2.1.9 | enterprise IdP (ADFS) MCP 연결 |
| 7 | `list_changed` 알림 | v2.1.0 | 동적 tool 등록 (reconnect 없이) |
| 8 | `MCP_CONNECTION_NONBLOCKING=true` | v2.1.109 | `-p` 모드 startup 가속 (CI) |

**의존성**: `.claude/settings.json.tmpl` mcpServers 섹션 확장.

**예상 LOC**: ~300-500

---

### 2.5 SPEC-V3R4-CC2X-WORKTREE-001 — Worktree Native Alignment (5 items)

**Rationale**: moai-adk-go worktree 시스템은 성숙하지만 native 기능과 정렬 필요. v2.1.49 native `--worktree`, v2.1.72 `EnterWorktree`/`ExitWorktree` 도구 통합.

| # | Feature | 버전 | moai 통합 |
|---|---------|------|-----------|
| 1 | Native `--worktree`/`-w` flag | v2.1.49 | `moai worktree new` ↔ native `--worktree` 통합 가이드, WorktreeCreate hook이 moai 디렉토리 자동 생성 |
| 2 | `EnterWorktree`/`ExitWorktree` 도구 | v2.1.72 | BODP §18.12 3-signal evaluation → native 도구 매핑, audit trail 통합 |
| 3 | `worktree.sparsePaths` | v2.1.76 | moai-adk-go monorepo SPEC 단위 격리 (`internal/cli/`, `.moai/specs/`) |
| 4 | `worktree.baseRef: fresh\|head` | v2.1.133 | lessons #13 (--team + SPEC worktree base 불일치) 자동 방지 |
| 5 | Stale worktree 자동 정리 | v2.1.76 | `feedback_pr_merge_branch_cleanup.md` 자동화 |

**예상 LOC**: ~600-800 (worktree 매니저 native 통합)

---

### 2.6 SPEC-V3R4-CC2X-SESSION-001 — Session Continuity Overhaul (6 items)

| # | Feature | 버전 | moai 통합 |
|---|---------|------|-----------|
| 1 | Auto-memory + moai lessons sync | v2.1.32~v2.1.97 | Claude auto-memory ↔ `memory/lessons.md` 양방향 sync 또는 SSOT 결정 |
| 2 | `/recap` 명령어 | v2.1.108 | session-handoff `paste-ready resume`과 결합, 세션 시작 시 자동 recap |
| 3 | `--from-pr` flag | v2.1.27 | PR-driven workflow 직접 매핑 (`claude --from-pr 856`) |
| 4 | Session auto-link to PR | v2.1.27, v2.1.81 | `feedback_pr_merge_branch_cleanup.md` 자동화 |
| 5 | `--name` / `/rename` | v2.1.62, v2.1.76 | SPEC 단위 세션 명명 ("SPEC-V3R4-CATALOG-001 Wave 2") |
| 6 | `/branch` for forking | v2.1.78 | manager-strategy alternate-strategy 탐색 |

**예상 LOC**: ~700-900

---

### 2.7 SPEC-V3R4-CC2X-PERMS-001 — Permission Model Evolution (6 items)

| # | Feature | 버전 | moai 통합 |
|---|---------|------|-----------|
| 1 | Auto Mode 기본화 | v2.1.111 | routine read-only 명령 자동 승인 (`Bash(go *)`, `Bash(make *)`) |
| 2 | `PermissionRequest` hook | v2.0.45 | production env 보호 프로그래매틱 적용 |
| 3 | Wildcard rules (`Bash(npm *)`) | v2.1.0 | 16-language CI 단순화 |
| 4 | `sandbox.failIfUnavailable: true` | v2.1.83 | enterprise 환경 안전 모드 강제 |
| 5 | `sandbox.network.deniedDomains` | v2.1.113 | test 환경 외부 호출 차단 |
| 6 | `CLAUDE_CODE_SUBPROCESS_ENV_SCRUB` + PID namespace | v2.1.83, v2.1.98 | Bash tool/hooks/MCP credential 격리 |

**예상 LOC**: ~400-600 (주로 settings.json 정책 + 테스트)

---

## 3. Tier 2 — 중기 도입 (18건, 7 SPECs)

### 3.1 SPEC-V3R4-CC2X-SDK-001 — SDK/Headless Performance (5 items)

| # | Feature | 버전 | moai 활용 |
|---|---------|------|-----------|
| 1 | `--bare` mode | v2.1.81 | hooks/LSP/plugin sync 스킵 → CI ~14% 가속 |
| 2 | `--max-budget-usd` | v2.0.28 | CG mode cost-optimization 정량화 |
| 3 | Custom session ID (`--session-id moai-sprint-12`) | v2.0.73 | SPEC 단위 deterministic 세션 |
| 4 | `--system-prompt-file` | v1.0.54 | specialized agent 프롬프트 외부 파일화 |
| 5 | SDK MCP `type: 'sdk'` | v2.1.69 | 내부 in-process MCP (MX validator) |

### 3.2 SPEC-V3R4-CC2X-PLUGIN-001 — Plugin Maturity (10 items)

| # | Feature | 버전 | moai 활용 |
|---|---------|------|-----------|
| 1 | `manifest.userConfig` | v2.1.83 | install-time setup wizard (GitHub token 등) |
| 2 | `themes/` 디렉토리 | v2.1.118 | moai 브랜드 컬러 (orange #FF6B35) 배포 |
| 3 | `monitors/` 디렉토리 | v2.1.105 | harness learning 백그라운드 monitor |
| 4 | `claude plugin tag` | v2.1.118 | SPEC-V3R2-ORC-* 릴리스 자동화 |
| 5 | `claude plugin prune` | v2.1.121 | stale dependency 정리 |
| 6 | `claude plugin details` | v2.1.139 | 컴포넌트 인벤토리 + 토큰 비용 |
| 7 | `claude plugin validate` 강화 | v2.1.78, v2.1.139 | moai 템플릿 CI에서 frontmatter/hook 검증 |
| 8 | `--plugin-url <zip>` | v2.1.129 | plugin archive 배포 |
| 9 | Plugin 의존성 해결 | v2.1.105 | conflicting/range-conflict 처리 |
| 10 | `$schema` in manifest | v2.1.139 | IDE 자동완성 |

### 3.3 SPEC-V3R4-CC2X-STATUS-001 — Statusline Expansion (7 fields)

`status_line.sh.tmpl`에 추가:
- `context_window.used_percentage` (v2.1.6) — session-handoff threshold (1M=75%, 200K=90%) 시각화
- `worktree` (name/path/branch) (v2.1.97) — moai worktree 정보
- `rate_limits` (v2.1.80) — usage limit
- `current_usage` (v2.0.70) — cost
- `git_worktree` (v2.1.97) — linked worktree 감지
- `refreshInterval` (v2.1.97) — 주기적 갱신
- `effort.level` / `thinking.enabled` (v2.1.119) — 현재 모드

### 3.4 SPEC-V3R4-CC2X-TUI-001 — TUI Modernization (5 items)

- Fullscreen mode (`/tui fullscreen`, v2.1.110)
- `/focus` mode (v2.1.110)
- `/agents` Running tab (v2.1.98)
- `claude agents` CLI (v2.1.50, v2.1.139)
- `/usage` 통합 + `/insights` (v2.1.101, v2.1.118)

### 3.5 SPEC-V3R4-CC2X-LOOP-001 — /loop + Cron (2 items)

- `/loop` 명령어 (v2.1.71) — 기존 `moai loop` skill alias
- Cron 도구 (CronCreate/CronList/CronDelete, v2.1.71) — nightly sweep, weekly audit

### 3.6 SPEC-V3R4-CC2X-OTEL-001 — Telemetry Expansion (5 items)

- `agent_id` / `parent_agent_id` (v2.1.117)
- `effort` 속성 (v2.1.117)
- `skill_activated` event (v2.1.126)
- `pull_request.count` metric (v2.1.129)
- W3C `TRACEPARENT` 전파 (v2.1.97, v2.1.98)

### 3.7 SPEC-V3R4-CC2X-PLAN-MODE-001 — Plan Mode Integration (5 items)

- `/plan` 명령어 (v2.1.0)
- Plan files (`plansDirectory: .moai/plans/`, v2.0.28, v2.1.9)
- Plan mode compaction preservation (v2.1.68)
- `/ultrareview` (v2.1.111)
- `/ultraplan` (v2.1.101)

---

## 4. Tier 3 — 장기 검토 (12건, 3 SPECs)

### 4.1 SPEC-V3R4-CC2X-PLATFORM-001 — Platform Extensions

- PowerShell tool (v2.1.84) — Windows 지원 확대
- Voice mode (v2.1.111, 한국어 STT v2.1.69) — 접근성
- Remote Control / Bridge (v2.1.78) — mobile/web 제어
- `/copy` 코드블록 picker (v2.1.59)

### 4.2 SPEC-V3R4-CC2X-ENTERPRISE-001 — Enterprise Features

- `forceRemoteSettingsRefresh` (v2.1.92)
- `managed-settings.d/` drop-in (v2.1.83)
- `feedbackSurveyRate` (v2.1.83)
- `disableAllHooks` (v1.0.68)

### 4.3 SPEC-V3R4-CC2X-MISC-UX-001 — Misc UX

- `themes/` 디렉토리 (v2.1.118)
- `keybindings.json` (v2.1.18)
- Settings 핫리로드 (v1.0.90)
- `--channels` (v2.1.80)

---

## 5. 도입 시 주의사항

### 5.1 Cross-cutting Concerns

1. **Hook 충돌**: 기존 moai hook handler ↔ 신규 native hook event 우선순위 정의 필요
2. **Frontmatter 마이그레이션**: 17 agent + 17 skill 일괄 업데이트 → builder-harness 위임 권장
3. **Settings 호환성**: `.moai/config/*.yaml` ↔ `.claude/settings.json` 양방향 변환
4. **버전 매트릭스**: moai-adk-go 최저 지원 Claude Code 버전 정책 (현재 v2.1.50 → 신규 도입 시 v2.1.119+ 권고)
5. **Plugin manifest 마이그레이션**: `experimental.themes`/`experimental.monitors` 네임스페이스 (v2.1.129)
6. **Auto-memory 충돌**: Claude auto-memory ↔ moai `memory/` SSOT 결정 필수

### 5.2 FROZEN Zone 영향

본 SPEC 시리즈는 **EVOLVABLE zone**에만 작용. FROZEN zone (constitution.md §2, askuser-protocol.md §1-7) 수정 없음.

`ConfigChange` hook (Tier 1.1 #3)은 FROZEN zone **보호 강화**가 목적 — 기존 FROZEN zone 신성성을 runtime으로 확장.

### 5.3 백워드 호환성

- 기존 SPEC-CC2122-* (v2.1.119-122) 시리즈와 중복 항목 없음 — SPEC-CC2122는 좁은 범위, SPEC-V3R4-CC2X-*는 광역
- 기존 `internal/hook/` retired stub (PR #858)과 정합 유지
- `manager-tdd`/`manager-ddd` retired stub (SPEC-V3R3-RETIRED-DDD-001) 정합 유지

---

## 6. 예상 효과 (정량)

| 메트릭 | 현재 | Sprint 12-14 완료 후 | 개선 |
|--------|------|----------------------|------|
| Worktree 작업 자동화율 | ~60% | ~95% | +35%p |
| Hook spawn overhead | 매 tool 호출 | `if` 매칭만 | -70% |
| Session-handoff 누락 | 수동 | 자동 (PostCompact) | -100% |
| `-p` cold-start | ~3s | ~0.5s (`--bare`) | -83% |
| FROZEN zone runtime 보호 | doc-only | enforced | 0→1 |
| Agent 정의 declarative율 | ~40% | ~95% | +55%p |

---

## 7. Sprint 권고 로드맵

### Sprint 12 (즉시 진입)
- **목표**: Tier 1 7 SPECs plan 작성
- **순서**: HOOKS → AGENT-FM → SKILL → MCP → WORKTREE → SESSION → PERMS
- **선행 의존**: SPEC-V3R4-CATALOG-001 (현재 plan merged, run 진행 중) 완료 후 시작
- **doctrine**: plan-in-main (PR #822) — plan PR은 main에 squash merge

### Sprint 13
- **목표**: Tier 2 7 SPECs plan + Tier 1 일부 run
- **순서**: SDK → PLUGIN → STATUS → TUI → LOOP → OTEL → PLAN-MODE
- **병렬**: Tier 1 HOOKS/AGENT-FM/SKILL run phase (TDD M1-M5)

### Sprint 14
- **목표**: Tier 3 3 SPECs plan + Tier 2 run
- **순서**: PLATFORM → ENTERPRISE → MISC-UX

### Sprint 15+
- 잔여 run/sync + v3.0.0-rc 준비

---

## 8. Open Questions (OQ) — child SPEC plan 단계에서 해결

- **OQ-1 [HOOKS]**: 기존 `internal/hook/` 핸들러와 native hook event의 우선순위 (parallel vs replace)
- **OQ-2 [AGENT-FM]**: agent .md 일괄 수정 방법 (builder-harness migration script vs 수동)
- **OQ-3 [SKILL]**: `context: fork` 적용 범위 (전체 17 skill vs 선택)
- **OQ-4 [MCP]**: Sequential Thinking + Context7만 `alwaysLoad` vs 모든 MCP
- **OQ-5 [WORKTREE]**: `moai worktree` CLI 명령어 deprecate 여부 (native --worktree로 완전 이주)
- **OQ-6 [SESSION]**: auto-memory ↔ moai memory SSOT 결정
- **OQ-7 [PERMS]**: auto mode 기본 활성화 여부 (CLAUDE.md §7 Rule 1 Approach-First와 상충 검토)

---

## 9. References

- Claude Code 공식 changelog v2.1.0 ~ v2.1.139 (1,500+ 항목 전수 분석)
- CLAUDE.md §1 HARD Rules
- CLAUDE.md §14 Parallel Execution Safeguards (worktree isolation)
- CLAUDE.md §15 Agent Teams
- CLAUDE.md §16 Context Search Protocol
- CLAUDE.local.md §18.0 5-axis git framework
- CLAUDE.local.md §18.12 BODP
- CLAUDE.local.md §19 AskUserQuestion Enforcement
- .claude/rules/moai/core/zone-registry.md (Frozen zone reference)
- .claude/rules/moai/workflow/session-handoff.md
- .claude/rules/moai/workflow/context-window-management.md
- memory/lessons.md (#1~#15)
- memory/feedback_*.md (12개 feedback memories)
- 기존 SPEC-CC2122-* 시리즈 (v2.1.119-122 좁은 범위 선행 작업)

---

**Status**: research-only (plan/spec/acceptance 미작성)
**Next Action**: `/moai plan SPEC-V3R4-CC2X-HOOKS-001` (Tier 1 첫 child SPEC 진입)
**Owner**: GOOS + MoAI orchestrator
**Last Updated**: 2026-05-12

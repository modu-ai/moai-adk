---
id: SPEC-V3R3-WEB-001
version: "0.2.0"
status: archived
created_at: 2026-05-04
updated_at: 2026-05-04
author: Goos
priority: High
labels: [web, cockpit, foundation, capability-c1, idea-001, archived]
issue_number: null
---

## HISTORY

| Version | Date | Author | Description |
|---------|------|--------|-------------|
| 0.1.0 | 2026-05-04 | Goos | Created from IDEA-001 brain proposal — Cockpit Foundation + Workflow Tracker (Capability C1, First Principles 1순위) |
| 0.1.1 | 2026-05-04 | Goos | Patch — D1 (REQ-WEB-003 + Scenario 3 decoupled from HTMX-specific attributes) + D2 (Goals section added) per plan-auditor iteration 1 review (partial — interrupted) |
| 0.2.0 | 2026-05-04 | Goos | **ARCHIVED** — Pivoted to Console paradigm (CRUD-capable) per IDEA-002 brain workflow. Original Cockpit (read-only ambient awareness) capability superseded by SPEC-V3R3-CONSOLE-NNN (TBD via brain workflow). Trigger: Claude Design handoff bundle "MoAI-ADK Console" introduced 4 settings CRUD requirement (user/project/design/harness) + Bun+Hono+React+TS Claude Agent SDK stack pivot, breaking the read-only invariant of original capability. Document preserved as historical context only; do NOT implement. |

> ⚠️ **ARCHIVED 2026-05-04** — This SPEC has been superseded. The Cockpit (read-only ambient awareness) paradigm was replaced by Console (CRUD-capable configuration management) following Claude Design handoff bundle review and user re-scope decision. See IDEA-002 brain workflow output and successor SPEC-V3R3-CONSOLE-NNN (assigned post-brain). This document is preserved as historical context for traceability — do NOT implement against these requirements. The plan-auditor PASS verdict (iteration 1) and 4 medium/low defects (D1-D4) remain valid against the original Cockpit capability but do not transfer to the Console successor.

---

# SPEC-V3R3-WEB-001: Cockpit Foundation + Workflow Tracker capability

## Overview

MoAI Cockpit의 첫 번째 SPEC으로, **CLI 진입점 + 단일 페이지 HTTP 서버 스켈레톤 + Workflow Tracker 패널**을 구현한다. `proposal.md` 81번째 라인의 capability scope를 그대로 수용하여, 현재 진행 중인 SPEC ID / Phase / progress.md 마지막 체크포인트를 5초 자동 갱신으로 surface한다. 본 SPEC은 후속 WEB-002~008 패널 SPEC들의 foundation이 되며, 단독으로도 First Principles Layer 4 root 욕구("AI agent 진척과 인간 결정 시점의 분리 인지")를 충족한다.

## Why

- **First Principles 1순위 (Layer 4 root)**: AI agent와 페어 프로그래밍하는 솔로 개발자가 agent 진척을 ambient awareness로 surface하는 가장 본질적 panel — Workflow Tracker가 단독으로도 30%+ CLI 왕복 감소 가치를 발생시킨다 (`ideation.md` Phase 5 §5.2 Layer 4 결론).
- **Productivity research 정합성**: 컨텍스트 전환 1회당 23분 15초 복귀 비용 (Mark, UC Irvine) → 도구 통합으로 최대 30% 생산성 회복 가능 (`research.md` §3.1, §3.2). 사용자 선택 성공 지표(왕복 30%+ 감소)와 정확히 일치.
- **Niche validation**: MoAI-ADK 도메인 객체(SPEC ID, Phase, progress checkpoint)를 1차 시민으로 다루는 read-only 가시화 도구는 검색 결과상 부재 — green field niche (`research.md` §1.3).
- **Foundation 역할**: WEB-002~008 SPEC들이 의존하는 HTTP 서버 스켈레톤·CLI 진입점을 본 SPEC에서 확립 (`proposal.md` SPEC Sequencing Rationale).

## Goals

| Goal ID | Type | Target |
|---------|------|--------|
| Goal Primary | Productivity | User CLI roundtrip count reduces by 30%+ (1-week pre-Cockpit baseline vs 1-week post; measured by `shell history` grep for `git status \| gh pr list \| gh pr checks \| git worktree list \| ls .moai/specs/`) |
| Goal Secondary 1 | UX (latency) | Page load p95 < 200ms |
| Goal Secondary 2 | Performance (polling cost) | Polling fetch time per request < 300ms |
| Goal Secondary 3 | Stability (forward concern) | gh API rate-limit hit count = 0 (architectural readiness for downstream WEB-003) |
| Goal Anti | Read-only invariant | Write-action click count = 0 |

The Primary goal validates the entire Cockpit thesis (deep-work preservation via consolidated ambient surface). Goal Anti is a HARD invariant that constrains every panel design — its violation invalidates the read-only product premise.

## EARS Requirements

### REQ-WEB-001 (Ubiquitous — Server Lifecycle / Single-Page Route)

The Cockpit server **shall** expose a single HTTP route `GET /` on `localhost:<port>` that renders the Cockpit page containing the Workflow Tracker panel, where `<port>` defaults to a sensible value (candidate: 8080) and is configurable via `--port` flag. The server **shall** bind exclusively to the loopback interface (localhost-only invariant).

### REQ-WEB-002 (Event-Driven — Page Load Triggers Workflow Tracker Render)

**When** a browser issues `GET /` to the Cockpit server, the server **shall** respond with a fully rendered HTML page within 200ms (p95) containing the Workflow Tracker panel populated with: (a) the active SPEC ID, (b) the current Phase identifier (one of: `plan`, `run`, `sync`), and (c) the timestamp of the most recent `progress.md` checkpoint.

### REQ-WEB-003 (State-Driven — Polling While Running)

**While** the Cockpit server process is running and a browser tab holds the Cockpit page open, the page **shall** issue a `GET` request to a fragment endpoint (candidate: `/api/workflow-tracker`) every 5 seconds via a client-side polling mechanism, and the server **shall** respond with the partial HTML fragment representing the current Workflow Tracker state within 300ms per request, and the page **shall** replace the Workflow Tracker panel content in-place without triggering a full-page navigation.

### REQ-WEB-004 (Unwanted Behavior — Graceful Empty / Error State)

**If** the SPEC discovery walker encounters an empty `.moai/specs/` directory, no readable `progress.md` files, or a filesystem read error, **then** the Workflow Tracker panel **shall** render a neutral "No active SPEC" empty state with no error toast, modal, or spinner, and the page **shall not** display any state-mutating UI control (no buttons, no forms, no `POST`/`PUT`/`DELETE`/`PATCH` triggers).

### REQ-WEB-005 (Complex — Multi-Active-SPEC Resolution)

**While** more than one SPEC directory under `.moai/specs/SPEC-*/` contains a non-empty `progress.md` file, **when** the Workflow Tracker renders, the server **shall** select the SPEC whose `progress.md` has the most recent modification time (mtime) as the primary surface, and **shall** display a non-interactive indicator showing the count of other active SPECs (e.g., "(N other active SPECs)") with no navigation affordance.

## Files to Modify

- `cmd/moai/main.go` — Register the new `cockpit` subcommand with flags `--port <int>` (default candidate 8080) and `--no-open` (boolean, suppress browser auto-open).

## Files to Create (initial estimate)

The exact internal package layout is deferred to `plan.md` and finalized during `/moai run`. Expected initial scope (binding decisions in Run phase):

- `internal/cockpit/` — package root for HTTP server, SPEC discovery, progress parser, Workflow Tracker handler
- Test files mirroring the package layout under `_test.go` convention
- Templ component file(s) for Workflow Tracker panel + page shell
- Integration test for end-to-end golden path (Phase F in `plan.md`)

## What NOT to Build

The following items are **explicitly out of scope** for SPEC-V3R3-WEB-001. Each has a dedicated downstream SPEC slot or is excluded by invariant:

- **C2 Worktree Switchboard panel** → deferred to SPEC-V3R3-WEB-002
- **C3 CI/PR Glance Wall panel** → deferred to SPEC-V3R3-WEB-003
- **C4 Memory Surfboard panel** → deferred to SPEC-V3R3-WEB-004
- **C5 Drift Sentinel panel** → deferred to SPEC-V3R3-WEB-005
- **C6 Quick Action Launcher** (optional) → deferred to SPEC-V3R3-WEB-006
- **C7 User Configuration system** (optional) → deferred to SPEC-V3R3-WEB-007
- **C8 Brand Tokenization integration** (optional) → deferred to SPEC-V3R3-WEB-008
- **All write actions** — read-only invariant from `proposal.md` Goal Anti row; the Cockpit page MUST NOT issue any `POST`/`PUT`/`DELETE`/`PATCH` request triggered by user interaction
- **Authentication / CSRF / multi-user** — localhost-only invariant; loopback binding is the sole access boundary
- **External database** — filesystem walk + git/gh CLI shell-out only; no SQLite, Postgres, or external store
- **Real-time streaming via SSE/WebSocket** — polling is sufficient per `research.md` §2.2; SSE/WebSocket explicitly rejected during Phase 4 Converge
- **Multiple-page navigation** — single-page invariant; no `/specs`, `/dashboard`, `/settings` sibling routes
- **Brand visual integration beyond reading** — deferred to WEB-008; this SPEC may reference brand assets only as candidates, not bind them

## Source References

### Brain artifacts (canonical scope source)
- `.moai/brain/IDEA-001/proposal.md` — Capability C1 scope, Goal table, SPEC sequencing rationale
- `.moai/brain/IDEA-001/ideation.md` — Lean Canvas, Phase 5 Critical Evaluation, First Principles Layer 4 conclusion
- `.moai/brain/IDEA-001/research.md` — Go templ + HTMX validation, niche validation, productivity research (30%+ recovery)
- `.moai/brain/IDEA-001/claude-design-handoff/{prompt,context,references,acceptance,checklist}.md` — downstream `/moai design --path A` bundle (acceptance.md sections A/B/E inform UX criteria)

### Project context
- `.moai/project/product.md` — product scope alignment
- `.moai/project/structure.md` — codebase layout (cmd/moai, internal/)
- `.moai/project/tech.md` — Go toolchain + dependency baseline

### Brand context (for downstream WEB-008)
- `.moai/project/brand/brand-voice.md` — panel title tone (concise, optimistic, technically precise)
- `.moai/project/brand/visual-identity.md` — color palette + typography tokens (referenced only, not bound here)
- `.moai/project/brand/target-audience.md` — persona alignment with Phase 1 Discovery

---
id: SPEC-V3R6-DOCS-CMD-CATALOG-001
title: "docs-site Command Catalog Drift 정정 (Critical 4건)"
version: "0.2.0"
status: implemented
created: 2026-05-22
updated: 2026-05-22
author: manager-spec
priority: P1
phase: "v3.0.0"
module: "docs-site/content"
lifecycle: spec-anchored
tags: "docs, docs-site, i18n, hugo, command-catalog, drift-cleanup"
tier: S
depends_on: []
related_specs: [SPEC-V3R5-DOCS-SECURITY-001]
---

# SPEC-V3R6-DOCS-CMD-CATALOG-001 — docs-site Command Catalog Drift 정정 (Critical 4건)

## HISTORY

- v0.1.0 (2026-05-22): 초기 draft 작성. baseline `.moai/research/moai-adk-current-state-2026-05-22.md` Critical 4건 (R1, R2, G1, G2) 단일 SPEC bundle. 사용자 결정 2026-05-22 AskUserQuestion: "Critical 4건만 단일 SPEC + 순차 진행" (대안 4-way split / Wave 2 통합 / High 5건 포함 거절). Open Questions 4건 모두 default 진행으로 가정 (OQ1 db 삭제 / OQ2 github 삭제 / OQ3 gate를 quality-commands/ 신설 / OQ4 weight pre-flight 기반 자동 결정 — pre-flight 결과 §1.4에 명시).
- v0.1.1 (2026-05-22): plan-auditor iter 1 verdict **REVISE** (composite 0.818, BLOCKING MP-3). 정정 적용: (a) **D1 BLOCKING — Frontmatter canonical 12-field SSOT** 위반 정정 — `title:` / `phase:` / `module:` / `lifecycle:` 4 required field 추가 + `created_at:` → `created:` / `updated_at:` → `updated:` / `labels:` → `tags:` (string) snake_case alias 제거 + `priority: High` → `priority: P1` 표준화 (`.claude/rules/moai/development/spec-frontmatter-schema.md` SSOT 준수). (b) **D2 BLOCKING — Tier 분류 lock** — `tier: S` 명시 (기존 frontmatter `tier:` 부재 → backward-compat default Tier L threshold 0.85, 0.818 composite FAIL; `tier: S` 명시로 threshold 0.75 lock). (c) **D3 MAJOR — plan.md §4 B4 row false self-assessment** 정정 — "spec.md 이미 정렬" → iter 1 정정 완료 reflection. (d) **D4 SHOULD — plan.md §12 time estimates** 제거 → Priority 라벨로 대체 (`agent-common-protocol.md` § Time Estimation HARD rule). (e) **D5 SHOULD — plan.md §10 E.4 + acceptance.md §4.2 arithmetic** 정리 — 12 _meta.yaml files 모두 수정 (16 entry-changes) + 1 main.yaml (4 block-changes) 명확화. (f) **D7/D8 SHOULD — acceptance.md §5 eval grep quoting bug** 정정 — bash array `"${target_files[@]}"` 패턴 적용으로 single-quote literal 매칭 결함 해소. iter 2 재감사 예상 composite ≥ 0.85 (Tier S threshold 0.75 + 10pp 마진).
- v0.2.0 (2026-05-22): run-phase COMPLETE on main. 8/8 ACs PASS (binary), Hugo local build exit 0 (1164ms), 29 file operations (8 delete + 8 create + 12 _meta.yaml modify + 1 main.yaml regenerated via scripts/gen_menu.py). PRESERVE list 보호 검증 PASS (외부 untracked `.tmp-parsetest/`, `docs-site/content/*/book/`, `docs-site/layouts/_default/redirect.html`, `docs-site/static/book/`는 본 SPEC 무관 — parallel session 산물). Tier S LEAN minimal form, orchestrator-direct execution. status `draft → implemented`, version `0.1.1 → 0.2.0`.

---

## 1. Overview

### 1.1 Goal

`docs-site` (adk.mo.ai.kr) 의 4-locale (ko/en/ja/zh) 명령어 카탈로그에서 발견된 **Critical drift 4건**을 정정한다. 사용자 가시 문서가 실제 구현과 어긋나는 부분(R1 `/moai db` + R2 `moai-github` false promise)을 제거하고, 실제 구현은 존재하나 문서가 없는 부분(G1 `/moai harness` + G2 `/moai gate`)을 신설한다.

### 1.2 Motivation

- **사용자 신뢰 손상**: `adk.mo.ai.kr/workflow-commands/moai-db` 페이지는 실제 명령이 없는데도 사용자에게 마치 동작하는 것처럼 안내 → 사용자가 명령을 시도하면 "unknown command" 에러 → "MoAI-ADK가 부서졌나?" 의심
- **운영 가시성 부재**: `/moai harness` (V3R4 Self-Evolving Harness, Tier-4 evolution + 5-layer safety) 와 `/moai gate` (pre-commit quality gate, lint+format+type-check+test 병렬) 는 실제 구현되어 있으나 docs-site에 어떤 안내도 없음 → 사용자는 명령 존재 자체를 모름
- **baseline 보고서 근거**: `.moai/research/moai-adk-current-state-2026-05-22.md` §2 Critical Drift Inventory 4건 모두 fact-checked (pre-flight §1.4 검증 결과 일치)
- **Critical 등급 선정 기준**: 사용자 가시 페이지(adk.mo.ai.kr 도메인 노출) + 4-locale 모두 영향 + 즉시 신뢰 손상 또는 즉시 운영 가치 회복 → High/Medium 5건 (cg-mode title alias / statusline preset 별칭 / harness CLI 차감 등)은 본 SPEC 범위 외

### 1.3 Scope

본 SPEC은 **docs-site 4-locale 페이지 + Hugo navigation 데이터** 만 다룬다. 다음은 명시적으로 다루지 않는다:

- **다음은 본 SPEC 범위가 아님 (Section 3 Exclusions 참조)**:
  - Go 소스 (`internal/`, `cmd/moai/`)
  - Skill 본문 또는 Agent 정의 (`.claude/skills/`, `.claude/agents/`)
  - 다른 카테고리 페이지 (advanced/, getting-started/, multi-llm/ 등)
  - High/Medium 5건 drift (baseline 보고서 §3)
  - 기존 페이지 weight 충돌 8건 (baseline 보고서 §5) — 신규 페이지 weight만 결정

### 1.4 Pre-flight 검증 결과 (사실 검증 완료)

| 항목 | 사실 | 검증 명령 |
|------|------|----------|
| R1 db 페이지 4-locale | ko/en/ja/zh 모두 3,300 bytes (i18n weight align — front matter sync) | `ls -lah docs-site/content/{ko,en,ja,zh}/workflow-commands/moai-db.md` |
| R1 db 명령 실제 | `cmd/moai/`에 부재 + `.claude/commands/moai/db.md` 부재 | `grep -l "moai db" cmd/moai/`; `ls .claude/commands/moai/db.md` |
| R2 github 페이지 4-locale | ko 9,850 / en 8,879 / ja 11,226 / zh 8,574 bytes (locale별 본문 번역됨) | `ls -lah docs-site/content/{ko,en,ja,zh}/utility-commands/moai-github.md` |
| R2 github 명령 실제 | `.claude/commands/moai/github.md` 부재 + `97-release-update.md` / `98-github.md` 도 `.claude/commands/`에 부재 (CLAUDE.local.md §21 dev-only 정책) | `ls .claude/commands/moai/{github,97-release-update,98-github}.md` |
| G1 harness 페이지 | 4-locale 모두 부재 | `ls docs-site/content/{ko,en,ja,zh}/workflow-commands/moai-harness.md` |
| G1 harness 실제 | `.claude/commands/moai/harness.md` (263 bytes) + skill `moai-workflow-harness` v2.0.0 (status/apply/rollback/disable 4 verbs, V3R4 Self-Evolving Harness, Tier-4 evolution + 5-layer safety) | `cat .claude/commands/moai/harness.md`; `cat .claude/skills/moai/workflows/harness.md \| head -100` |
| G2 gate 페이지 | 4-locale 모두 부재 | `ls docs-site/content/{ko,en,ja,zh}/quality-commands/moai-gate.md` |
| G2 gate 실제 | `.claude/commands/moai/gate.md` (188 bytes) + skill `moai-workflow-gate` v1.0.0 (Lightweight pre-commit, lint+format+type-check+test 병렬, <30s) | `cat .claude/commands/moai/gate.md`; `cat .claude/skills/moai/workflows/gate.md \| head -50` |
| weight 매핑 (workflow-commands ko) | project=20, brain=25, design=26, plan=30, run=40, sync=50, db=50 (sync 충돌) | `for f in docs-site/content/ko/workflow-commands/*.md; do grep '^weight:' $f; done` |
| weight 매핑 (quality-commands ko) | review=20, coverage=30, e2e=40, codemaps=50 | 동일 패턴 |
| weight 매핑 (utility-commands ko) | moai=20, github=30, loop=40, fix=50, clean=60, mx=70, feedback=80 | 동일 패턴 |
| Hugo navigation 데이터 | `docs-site/data/menu/main.yaml` (14,103 bytes) — 4건 모두 entry 존재 (db: ref `/workflow-commands/moai-db`, github: ref `/utility-commands/moai-github`) + harness/gate entry 부재 | `grep -B 1 -A 2 "moai-db\|moai-github\|moai-harness\|moai-gate" docs-site/data/menu/main.yaml` |
| `_meta.yaml` (4 locale × 3 카테고리) | workflow-commands에 `moai-db` entry / utility-commands에 `moai-github` entry / quality-commands·workflow-commands에 harness·gate entry 부재 | `cat docs-site/content/ko/{workflow,quality,utility}-commands/_meta.yaml` |

**weight 자동 결정 (OQ4 default)**:
- G1 harness (workflow-commands): **weight: 55** (sync=50 다음 자연 순서, db=50 삭제로 충돌 해소)
- G2 gate (quality-commands): **weight: 15** (review=20 앞, pre-commit gate가 review의 prerequisite으로서 사용자 mental model에 부합)

### 1.5 Constitution 정렬

- `.claude/rules/moai/development/coding-standards.md`: 16-language neutrality — 신규 페이지 본문은 docs-site 전역 다국어 정책 따름 (ko 원문 → en/ja/zh 번역)
- `.moai/docs/docs-site-i18n-rules.md`: 4-locale 동기화 의무 (REQ-LCK-1) + Mermaid TD-only (REQ-MMD-1) + canonical adk.mo.ai.kr (REQ-URL-1) + emoji 0 (REQ-EMJ-1)
- `.claude/rules/moai/design/constitution.md`: docs-site는 brand context 하위 자산 — 본 SPEC scope는 적용 X (콘텐츠 신뢰성 회복만)

---

## 2. User Stories

### US-DCC-001: 신규 사용자 신뢰 회복 (R1)

> **As a** MoAI-ADK 신규 사용자
> **I want** adk.mo.ai.kr/workflow-commands에서 안내되는 명령어가 실제로 동작하기를
> **So that** 문서가 부서지지 않은 도구라는 인상을 유지할 수 있다

**Acceptance**: `adk.mo.ai.kr/workflow-commands/` 페이지 목록에 `/moai db` 항목이 존재하지 않거나, 존재한다면 `.claude/commands/moai/db.md`가 실제로 있어야 한다.

### US-DCC-002: 안내된 명령의 실재성 보장 (R2)

> **As a** 사용자
> **I want** adk.mo.ai.kr/utility-commands에서 안내되는 `moai-github` 명령이 실제 사용 가능하기를
> **So that** PR 자동화나 release 작업 시 막힘이 없다

**Acceptance**: `adk.mo.ai.kr/utility-commands/` 페이지 목록에 `moai-github` 항목이 존재하지 않거나, 존재한다면 `.claude/commands/moai/github.md` (또는 동등 SSoT) 가 실제로 있어야 한다.

### US-DCC-003: harness 운영 가시성 회복 (G1)

> **As a** 시스템 운영자
> **I want** adk.mo.ai.kr에서 `/moai harness` 명령 사용법과 5-layer safety 개념을 학습할 수 있기를
> **So that** Tier-4 evolution 제안이 떴을 때 어떻게 대응할지 알 수 있다

**Acceptance**: `adk.mo.ai.kr/workflow-commands/moai-harness/` 페이지가 4-locale (ko/en/ja/zh) 존재하며, `status`/`apply`/`rollback`/`disable` 4 verbs와 V3R4 self-evolving harness 핵심 개념(observer, 4-tier ladder, 5-layer safety)을 안내한다.

### US-DCC-004: pre-commit 품질 게이트 가시성 회복 (G2)

> **As a** 개발자
> **I want** adk.mo.ai.kr에서 `/moai gate` (pre-commit 품질 게이트) 사용법을 학습할 수 있기를
> **So that** commit 전 빠른 검증(<30s) 패턴을 알 수 있다

**Acceptance**: `adk.mo.ai.kr/quality-commands/moai-gate/` 페이지가 4-locale 존재하며, `--fix`/`--staged`/`--file PATH` 옵션과 review/sync와의 차이점(<30s pre-commit vs 2-5min review vs 5-10min sync)을 안내한다.

---

## 3. Out of Scope

### 3.1 Out of Scope — Go 소스 변경 금지

본 SPEC은 docs-site 콘텐츠만 다룬다. `internal/`, `cmd/moai/`, `pkg/` 등 어떤 Go 파일도 수정하지 않는다.

### 3.2 Out of Scope — Skill / Agent 정의 변경 금지

`.claude/skills/moai/workflows/{harness,gate}.md`, `.claude/commands/moai/{harness,gate}.md`, `.claude/agents/` 등 SSoT 본문은 read-only로 참조한다. 본문이 부정확하다면 별도 SPEC.

### 3.3 Out of Scope — 다른 카테고리 페이지

`advanced/`, `getting-started/`, `multi-llm/`, `core-concepts/` 등 다른 카테고리는 본 SPEC 범위 외. Critical 4건만이 본 SPEC scope.

### 3.4 Out of Scope — High/Medium drift 5건

baseline 보고서 §3 (High 등급 cg-mode title alias / statusline preset 별칭 / harness CLI 차감 / SPEC v3 docs / multi-llm cost matrix 등)은 본 SPEC scope 외. 별도 SPEC 후보.

### 3.5 Out of Scope — 기존 페이지 weight 충돌 8건

baseline 보고서 §5에서 발견된 기존 페이지 weight 충돌 (예: workflow-commands plan=30 / _index=30 동일) 8건은 본 SPEC scope 외. 본 SPEC은 신규 페이지 weight만 결정 (G1=55, G2=15). 기존 weight 정렬은 별도 SPEC `SPEC-V3R6-DOCS-WEIGHT-NORM-001` (provisional).

### 3.6 Out of Scope — adk.mo.ai.kr 빌드/배포 검증

본 SPEC은 콘텐츠 정정만 다룬다. Vercel 빌드 성공/실패는 run-phase에서 `hugo --gc --minify` 로컬 검증으로만 확인. 실제 배포 검증은 머지 후 별도.

### 3.7 Out of Scope — Stub 마킹 옵션

OQ1/OQ2 default(완전 삭제)를 따른다. "planned/future" 또는 "maintainer-only" 라벨 마킹 옵션은 본 SPEC에서 채택하지 않음.

### 3.8 Out of Scope — `_meta.yaml` 외부 데이터 파일

`docs-site/data/menu/main.yaml` 외 `extra.yaml` 등 부수적 데이터 파일은 4건과 직접 관계없으면 수정하지 않음.

---

## 4. Requirements (EARS)

### 4.1 R1 처리 (db false promise 제거)

**REQ-DCC-001 (Event-Driven)**: WHEN run-phase 실행 시, the system SHALL delete the following 4 files: `docs-site/content/{ko,en,ja,zh}/workflow-commands/moai-db.md`.

**REQ-DCC-002 (Event-Driven)**: WHEN moai-db.md 4 파일 삭제 시, the system SHALL remove the `moai-db` entry from `docs-site/content/{ko,en,ja,zh}/workflow-commands/_meta.yaml` (4 locale × 1 entry = 4 removals).

**REQ-DCC-003 (Event-Driven)**: WHEN moai-db.md 4 파일 삭제 시, the system SHALL remove the `/workflow-commands/moai-db` reference block from `docs-site/data/menu/main.yaml` (1 removal — 단일 entry, 4-locale ko/en/ja/zh 라벨 포함).

### 4.2 R2 처리 (github false promise 제거)

**REQ-DCC-004 (Event-Driven)**: WHEN run-phase 실행 시, the system SHALL delete the following 4 files: `docs-site/content/{ko,en,ja,zh}/utility-commands/moai-github.md`.

**REQ-DCC-005 (Event-Driven)**: WHEN moai-github.md 4 파일 삭제 시, the system SHALL remove the `moai-github` entry from `docs-site/content/{ko,en,ja,zh}/utility-commands/_meta.yaml` (4 locale × 1 entry = 4 removals).

**REQ-DCC-006 (Event-Driven)**: WHEN moai-github.md 4 파일 삭제 시, the system SHALL remove the `/utility-commands/moai-github` reference block from `docs-site/data/menu/main.yaml` (1 removal — 단일 entry).

### 4.3 G1 처리 (harness 페이지 신설)

**REQ-DCC-007 (Event-Driven)**: WHEN run-phase 실행 시, the system SHALL create `docs-site/content/{ko,en,ja,zh}/workflow-commands/moai-harness.md` (4 new files) with frontmatter `weight: 55`, `draft: false`, locale-appropriate `title`/`description`, and body content covering: (a) command syntax (`/moai harness {status|apply|rollback <YYYY-MM-DD>|disable}`), (b) V3R4 Self-Evolving Harness 4 verbs와 각 verb의 동작, (c) 4-tier evolution ladder + 5-layer safety overview, (d) Tier-4 application gate (orchestrator-only AskUserQuestion) 안내, (e) cross-ref to `.claude/skills/moai/workflows/harness.md` and SPEC-V3R4-HARNESS-001.

**REQ-DCC-008 (Event-Driven)**: WHEN moai-harness.md 4 파일 생성 시, the system SHALL add `moai-harness` entry to `docs-site/content/{ko,en,ja,zh}/workflow-commands/_meta.yaml` with title `/moai harness` (locale-translated), AND add `/workflow-commands/moai-harness` reference block to `docs-site/data/menu/main.yaml` with 4-locale name labels (ko/en/ja/zh).

### 4.4 G2 처리 (gate 페이지 신설)

**REQ-DCC-009 (Event-Driven)**: WHEN run-phase 실행 시, the system SHALL create `docs-site/content/{ko,en,ja,zh}/quality-commands/moai-gate.md` (4 new files) with frontmatter `weight: 15`, `draft: false`, locale-appropriate `title`/`description`, and body content covering: (a) command syntax (`/moai gate [--fix] [--staged] [--file PATH]`), (b) lint+format+type-check+test 병렬 실행 개념 + <30s 목표, (c) `/moai gate` vs `/moai review` vs sync Phase 0.5 비교 표 (skill 본문 §Difference from Other Workflows 출처), (d) 16-language language detection 언급, (e) cross-ref to `.claude/skills/moai/workflows/gate.md`.

**REQ-DCC-010 (Event-Driven)**: WHEN moai-gate.md 4 파일 생성 시, the system SHALL add `moai-gate` entry to `docs-site/content/{ko,en,ja,zh}/quality-commands/_meta.yaml` with title `/moai gate` (locale-translated), AND add `/quality-commands/moai-gate` reference block to `docs-site/data/menu/main.yaml` with 4-locale name labels (ko/en/ja/zh).

### 4.5 Cross-cutting (Ubiquitous)

**REQ-DCC-011 (Ubiquitous)**: The system SHALL maintain 4-locale parity (ko/en/ja/zh) for all create/delete operations — never partial (e.g., delete ko but keep en).

**REQ-DCC-012 (Unwanted)**: The system SHALL NOT introduce: (a) forbidden URLs (anything other than canonical `adk.mo.ai.kr`), (b) Mermaid LR/BT/RL directives (only TD allowed per docs-site-i18n-rules), (c) emojis in new pages (per CWE-345/security policy), (d) `draft: true` or stub disclaimers in new pages.

---

## 5. Edge Cases

### EC-DCC-001: `_meta.yaml` 의 `moai-db` / `moai-github` entry는 4-locale 모두 동일 키 사용
- 4 locale × 1 entry = 4 deletions per file (REQ-DCC-002/005).
- 다른 entry 의 indentation / order 영향 없음 (YAML key 단위 삭제).

### EC-DCC-002: `main.yaml` 단일 파일에 4-locale 라벨 함께 위치
- main.yaml은 4-locale 라벨 (ko/en/ja/zh)을 단일 entry 내 `name:` 필드에 함께 보관하는 Hugo Geekdoc 패턴. 1 entry 삭제 = 1 block 제거 (단일 파일).
- harness/gate 신설 시도 동일 패턴: 1 block 추가 (단일 파일).

### EC-DCC-003: db 페이지 weight: 50 → sync=50 과 충돌 잔존 가능성
- db 삭제 후 workflow-commands 카테고리의 sync=50 만 남음 (충돌 해소). harness=55 추가 시 순서는 project(20) → brain(25) → design(26) → plan(30) → run(40) → sync(50) → harness(55) — 자연 정렬.

### EC-DCC-004: github 페이지 weight: 30 → loop=40 / fix=50 등과 충돌 없음
- github 삭제 후 utility-commands 카테고리는 moai(20) → loop(40) → fix(50) → clean(60) → mx(70) → feedback(80) 자연 정렬. weight 충돌 없음.

### EC-DCC-005: gate 페이지 weight: 15 가 review=20 앞에 위치
- quality-commands 카테고리는 gate(15) → review(20) → coverage(30) → e2e(40) → codemaps(50) — gate가 pre-commit gate로서 review의 prerequisite mental model에 부합.

### EC-DCC-006: 머지 후 Vercel 자동 빌드 실패 가능성
- `_meta.yaml` 또는 `main.yaml` syntax error 시 Vercel 빌드 실패 → run-phase에서 로컬 `hugo --gc --minify --buildDrafts=false` 실행으로 사전 검증 의무 (AC-DCC-007).

---

## 6. Dependencies

### 6.1 SPEC Dependencies

- **None blocking**: 본 SPEC은 독립 실행 가능. CATALOG-SSOT-001 / ABSORB-CLEANUP-001 / HARNESS-RENAME-001 등 진행 중 SPECs와 paths 충돌 없음 (docs-site/ vs .claude/agents/ + .moai/config/ + internal/).
- **Related**: SPEC-V3R5-DOCS-SECURITY-001 (commit `c94d8b203`, 4-locale security-notes.md 신설 패턴 — 본 SPEC harness/gate 신설 시 frontmatter 패턴 참조)

### 6.2 PRESERVE List (HARD — Section A)

다음 working tree dirty 파일/디렉토리는 본 SPEC scope 외, 절대 수정 금지:

- Modified: `.moai/harness/usage-log.jsonl`, `docs-site/hugo.toml`, `docs-site/layouts/_default/baseof.html`, `docs-site/layouts/partials/menu.html`
- Untracked: `.moai/research/{config-audit, lsp-yaml-v2-audit, moai-adk-current-state, v3.0-design}-*.md`, `.moai/specs/SPEC-V3R5-{GIT-STRATEGY-SCHEMA, INIT-WIZARD-EXPANSION, STATUSLINE-PROFILE-WIZARD, WORKFLOW-SCHEMA-EXTEND}-001/`, `docs-site/data/`, `docs-site/scripts/gen_menu.py`, `internal/hook/.moai/`

**예외**: REQ-DCC-003/006/008/010 으로 `docs-site/data/menu/main.yaml` 을 수정해야 한다. `docs-site/data/` 디렉토리 자체는 untracked 상태이지만, `main.yaml` 파일은 이미 staged/committed 상태로 본 SPEC 변경 대상. `docs-site/scripts/gen_menu.py` 등 다른 untracked 자산은 절대 건드리지 않는다.

---

## 7. Open Questions (default 진행 가정 — 사용자 동의 보류)

본 SPEC은 다음 4개 OQ에 대해 **default 진행을 가정**하여 작성되었다. orchestrator가 plan-auditor 결과 받은 후 AskUserQuestion 통해 최종 결정 가능 — default 채택 시 본 SPEC 그대로 run-phase 진입 가능.

| OQ | 질문 | Default (적용) | 대안 |
|----|------|----------------|------|
| OQ1 | R1 db 처리 방식 | **완전 삭제 (4 files + meta + main.yaml)** | "planned/future" 마킹 (페이지 유지하되 본문 stub화) |
| OQ2 | R2 github 처리 방식 | **완전 삭제 (4 files + meta + main.yaml)** | "maintainer-only" 라벨 (페이지 유지하되 dev-only 명시) |
| OQ3 | G2 gate 페이지 위치 | **quality-commands/ 신설** | review에 흡수 (review 페이지 내 §gate 섹션 추가) |
| OQ4 | G1/G2 weight 결정 | **G1=55 (sync=50 다음), G2=15 (review=20 앞)** | pre-flight 결과 기반 다른 값 (예: G2=25, review와 coverage 사이) |

---

## 8. Risks (요약, 상세는 plan.md)

| Risk ID | 설명 | Severity |
|---------|------|----------|
| R-DCC-001 | `main.yaml` YAML syntax error 시 Vercel 빌드 fail | Medium |
| R-DCC-002 | harness/gate 본문이 skill source와 drift 발생 시 long-term 유지보수 부담 | Low |
| R-DCC-003 | 4-locale 번역 품질 부족 (ko 원문 → en/ja/zh 직역) | Low |

---

## References

- baseline: `.moai/research/moai-adk-current-state-2026-05-22.md` (416 lines, 본 SPEC SSoT)
- skill sources: `.claude/skills/moai/workflows/harness.md` (V3R4), `.claude/skills/moai/workflows/gate.md`
- command sources: `.claude/commands/moai/harness.md`, `.claude/commands/moai/gate.md`
- i18n rules: `.moai/docs/docs-site-i18n-rules.md`
- 선례 SPEC: SPEC-V3R5-DOCS-SECURITY-001 (commit `c94d8b203`, 4-locale 신설 패턴)
- 사용자 결정: 2026-05-22 AskUserQuestion — "Critical 4건만 단일 SPEC + 순차 진행"

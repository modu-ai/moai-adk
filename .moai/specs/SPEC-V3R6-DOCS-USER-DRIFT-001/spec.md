---
id: SPEC-V3R6-DOCS-USER-DRIFT-001
title: "docs-site 사용자 가시 잔여 drift 정리 (F8 sync CI doctrine + 4-locale 회귀 차단)"
version: "0.1.0"
status: draft
created: 2026-05-22
updated: 2026-05-22
author: manager-spec
priority: High
phase: "v3.0.0"
module: "docs-site/content"
lifecycle: spec-anchored
tags: "docs, docs-site, i18n, hugo, sync-doctrine, ci-watch, drift-cleanup, wave-0"
tier: S
depends_on: [SPEC-V3R6-DOCS-CMD-CATALOG-001]
related_specs: [SPEC-V3R5-DOCS-SECURITY-001, SPEC-V3R6-DOCS-CMD-CATALOG-001]
---

# SPEC-V3R6-DOCS-USER-DRIFT-001 — docs-site 사용자 가시 잔여 drift 정리

## HISTORY

- v0.1.0 (2026-05-22): 초기 draft. Wave 0 두 번째 SPEC (Wave 0 SPEC #2 — `.moai/research/v3.0-design-2026-05-22.md` §Wave 0 라인 351-356). baseline 보고서 `.moai/research/moai-adk-current-state-2026-05-22.md` §8 사용자 가시 표면 결함 F1-F8 중 F1/F2/F3/F4는 선행 PR #1045 (SPEC-V3R6-DOCS-CMD-CATALOG-001, OPEN MERGEABLE)이 완전 해소. 본 SPEC은 **F8 (sync workflow CI watch/autofix doctrine 사용자 미가시)** 만 처리하는 좁은 scope. F5/F7 (weight 충돌 + design migration-guide ambiguity)은 별도 SPEC 후보로 명시 deferral. F6 (WorktreeCreate hook 회귀) 은 commit `a3239d3de`로 정리 완료.

---

## 1. Overview

### 1.1 Goal

`docs-site` (adk.mo.ai.kr) `workflow-commands/moai-sync.md` 4-locale (ko/en/ja/zh) 페이지에 **PR 생성 후 CI 모니터링 doctrine** (Wave 1 ci-watch + Wave 2 ci-autofix 2단계 패턴) 을 사용자 가시 형태로 surface 한다. 현재 사용자는 `/moai sync` 가 PR을 만든 직후 무슨 일이 일어나는지 (CI 폴링, 자동 fix 시도) 모르며, 막연한 대기 상태에서 혼란을 겪는다.

### 1.2 Motivation

- **사용자 가시성 부재**: `.claude/rules/moai/workflow/ci-watch-protocol.md` (Wave 1, 6,520 bytes) + `.claude/rules/moai/workflow/ci-autofix-protocol.md` (Wave 2, 6,124 bytes) 가 모두 존재하지만 **rules 레이어 internal 문서**일 뿐, `adk.mo.ai.kr` 사용자 가시 페이지에서는 일언반구도 없음. `docs-site/content/{ko,en,ja,zh}/workflow-commands/moai-sync.md` 4 파일 모두 `ci-watch` / `ci-autofix` / "CI 모니터" / "auto-fix" 키워드 0건 (pre-flight grep 검증 완료).
- **사용자 혼란**: PR 생성 후 사용자는 (a) CI 진행 중 — 무한 대기? (b) auto-fix 가 자동 commit 을 누르는 중? (c) AskUserQuestion 이 뜨면 무엇을 답해야 하나? — 어떤 안내도 없음.
- **선례 SPEC**: SPEC-V3R6-DOCS-CMD-CATALOG-001 (PR #1045 OPEN) 은 명령어 카탈로그 drift 4건을 정리했으나, sync workflow **다음 단계** doctrine 은 다루지 않음. 본 SPEC은 PR #1045 의 후속 작업으로서 sync 페이지 본문에 새 H2 섹션을 삽입.

### 1.3 Scope (이번 SPEC이 다루는 것)

- `docs-site/content/{ko,en,ja,zh}/workflow-commands/moai-sync.md` 4 파일에 **새 H2 섹션 `## PR 머지 후 CI 모니터링`** (locale-translated) 을 `## 품질 게이트` 직전에 삽입.
- 새 섹션은 ci-watch 2단계 패턴 (Wave 1 폴링 + Wave 2 자동 fix 루프 최대 3회 + 4회차 escalation AskUserQuestion) 을 사용자 언어로 설명. rules/skill 본문은 read-only 참조만 (수정 없음).
- 4-locale 동시 업데이트 (ko 원문 → en/ja/zh 번역) — `.moai/docs/docs-site-i18n-rules.md` §17.3 HARD 의무.
- Hugo 로컬 빌드 통과 검증 (`hugo --gc --minify` exit 0, 4-locale 페이지 200 OK).

### 1.4 Pre-flight 검증 결과 (사실 검증 완료)

| 항목 | 사실 | 검증 명령 |
|------|------|----------|
| F1 db 페이지 | PR #1045 (OPEN MERGEABLE) 가 `workflow-commands/moai-db.md` × 4-locale 삭제 처리. 본 SPEC 범위 외. | `gh pr diff 1045 --name-only` |
| F2 github 페이지 | PR #1045 가 `utility-commands/moai-github.md` × 4-locale 삭제 처리. 본 SPEC 범위 외. | `gh pr diff 1045 --name-only` |
| F3 harness 페이지 | PR #1045 가 `workflow-commands/moai-harness.md` × 4-locale 신설 처리. 본 SPEC 범위 외. | `gh pr diff 1045 --name-only` |
| F4 gate 페이지 | PR #1045 가 `quality-commands/moai-gate.md` × 4-locale 신설 처리. 본 SPEC 범위 외. | `gh pr diff 1045 --name-only` |
| F8 sync 페이지 CI doctrine | 4-locale 모두 `ci-watch` / `ci-autofix` / "auto-fix" / "CI 모니터" 키워드 0건 | `grep -l 'ci-watch\|ci-autofix\|auto-fix\|CI 모니터' docs-site/content/{ko,en,ja,zh}/workflow-commands/moai-sync.md` (출력 empty) |
| F8 baseline rules 자료 | `ci-watch-protocol.md` 6,520 bytes (30s 폴링 + 30분 timeout + `.github/required-checks.yml` SSoT) + `ci-autofix-protocol.md` 6,124 bytes (최대 3 iter + iter≥4 AskUserQuestion blocking + `.env`/secrets 보호) 모두 존재. | `ls .claude/rules/moai/workflow/ci-{watch,autofix}-protocol.md` |
| 현재 sync 페이지 라인 수 baseline | ko 570 / en 511 / ja 511 / zh 512 (총 2,104 라인) | `wc -l docs-site/content/{ko,en,ja,zh}/workflow-commands/moai-sync.md` |
| 현재 sync 페이지 H2 구조 (ko 기준) | 개요 / 사용법 / 지원 모드 / 지원 플래그 / 실행 과정 / 단계별 상세 / 생성되는 문서 / [Unreleased] / Git 자동화 / 품질 게이트 / 실전 예시 / 자주 묻는 질문 / v2.9.0 신규 기능 / 관련 문서 (총 14 H2) | `grep '^## ' docs-site/content/ko/workflow-commands/moai-sync.md` |
| 삽입 위치 | `## 품질 게이트` 직전 (Git 자동화 → CI 모니터링 → 품질 게이트 자연 흐름) | 위 H2 구조 기준 |
| PR #1045 파일 충돌 사전 검증 | PR #1045 file 목록과 본 SPEC 수정 대상 (`*/workflow-commands/moai-sync.md` × 4) 교차점 **0 건** | `gh pr diff 1045 --name-only \| grep moai-sync` (empty) |

### 1.5 Constitution 정렬

- `.moai/docs/docs-site-i18n-rules.md` §17.1 — canonical URL `adk.mo.ai.kr` 만 사용 (forbidden URL 0건 의무).
- `.moai/docs/docs-site-i18n-rules.md` §17.2 — Mermaid TD-only (본 SPEC 새 섹션은 Mermaid 다이어그램 0개 → 위반 가능성 0).
- `.moai/docs/docs-site-i18n-rules.md` §17.3 — 4-locale 동시 업데이트 HARD (ko 원문 + en/ja/zh 번역 동일 PR 내).
- `.moai/docs/dev-only-commands-isolation.md` §21 — dev-only `97-*`/`98-*`/`99-*` 식별자 본 SPEC scope 외 (sync 페이지에 dev-only 참조 도입 금지).
- `.claude/rules/moai/development/coding-standards.md` — 16-language neutrality (본 SPEC 영향 X — docs-site 다국어 정책).

---

## 2. User Stories

### US-DUD-001: PR 생성 후 무엇이 일어나는지 알 권리

> **As a** MoAI-ADK 사용자
> **I want** `adk.mo.ai.kr/workflow-commands/moai-sync` 페이지에서 `/moai sync` 가 PR 을 생성한 직후 어떤 자동화가 진행되는지를 명확히 학습할 수 있기를
> **So that** PR 페이지에서 5분간 댓글이 안 달려도 "MoAI가 죽었나?" 하고 의심하지 않는다

**Acceptance**: 4-locale 페이지에 `## PR 머지 후 CI 모니터링` (locale-translated) H2 섹션이 존재. 본문에 (a) 30초 폴링 + 30분 timeout, (b) auto-fix loop 최대 3 iter, (c) iter≥4 시 AskUserQuestion blocking 안내 모두 포함.

### US-DUD-002: auto-fix 의 안전 경계 이해

> **As a** 사용자
> **I want** auto-fix 가 `.env` / credentials / `scripts/ci-watch/run.sh` 를 절대 건드리지 않음을 사전에 보장받기를
> **So that** 자동화에 commit 권한을 주는 두려움이 줄어든다

**Acceptance**: 새 섹션에 "auto-fix가 절대 수정하지 않는 파일" subsection 또는 callout 존재 — `.env`, `.env.*`, credentials, scripts/ci-watch/run.sh, .github/required-checks.yml 명시.

### US-DUD-003: semantic failure vs syntax failure 의 차이 이해

> **As a** 사용자
> **I want** auto-fix 가 자동으로 처리하는 결함 (lint/format/test syntax error) vs 사람이 결정해야 하는 결함 (data race / deadlock / panic / test assertion failure) 의 차이를 알기를
> **So that** AskUserQuestion 이 뜰 때 무엇을 답해야 하는지 사전에 mental model 을 준비할 수 있다

**Acceptance**: 새 섹션에 "자동 처리 vs 사람 결정 필요" 분류 표 또는 명확한 prose 설명 존재.

---

## 3. Scope

### 3.1 In Scope (본 SPEC 범위 — 4-locale 동시)

- `docs-site/content/ko/workflow-commands/moai-sync.md` — 새 H2 `## PR 머지 후 CI 모니터링` 삽입 (`## 품질 게이트` 직전).
- `docs-site/content/en/workflow-commands/moai-sync.md` — 새 H2 `## CI monitoring after PR creation` 삽입 (`## Quality gates` 직전, locale-equivalent heading).
- `docs-site/content/ja/workflow-commands/moai-sync.md` — 새 H2 `## PR 作成後の CI モニタリング` 삽입 (`## 品質ゲート` 직전).
- `docs-site/content/zh/workflow-commands/moai-sync.md` — 새 H2 `## PR 创建后的 CI 监控` 삽입 (`## 质量门` 직전).

### 3.2 Out of Scope (별도 SPEC 후보)

#### 3.2.1 Out of Scope — F1-F4 (PR #1045 책임)

baseline 보고서 §8 F1 (`/moai db` 페이지) / F2 (`moai-github` 페이지) / F3 (`/moai harness` 페이지) / F4 (`/moai gate` 페이지) 은 선행 PR #1045 (`SPEC-V3R6-DOCS-CMD-CATALOG-001`, OPEN MERGEABLE) 가 완전 처리. 본 SPEC은 PR #1045 와 file overlap 0건 (pre-flight 검증 §1.4 참조).

#### 3.2.2 Out of Scope — F5 (weight 충돌 8건)

baseline §8 F5 docs-site weight 충돌 8건 (예: `workflow-commands/_index=30` vs `moai-plan=30` 충돌) 은 별도 SPEC 후보 `SPEC-V3R6-DOCS-WEIGHT-FIX-001` (가칭, provisional) 에서 처리. 본 SPEC 은 신규 섹션 삽입만 다루며 기존 페이지 weight 정렬은 건드리지 않는다.

#### 3.2.3 Out of Scope — F7 (design/migration-guide ambiguity)

baseline §8 F7 `design/migration-guide.md` "무엇으로부터" 불명 항목은 별도 SPEC 후보 또는 Wave 5 `SPEC-V3R6-DOCS-V3-MIGRATION-001` 흡수 가능. 본 SPEC scope 외.

#### 3.2.4 Out of Scope — F6 (WorktreeCreate hook 회귀)

baseline §8 F6 은 commit `a3239d3de` (2026-05-22) 에서 settings.json 등록 해제 + 4-locale hooks-reference 정정 + 공식 컨트랙트 문서화로 정리 완료. 메모리 `project_v3r6_worktree_hook_root_cause_fix_2026_05_22` 참조.

#### 3.2.5 Out of Scope — Go 소스 / Skill 본문 / rules 본문 변경

`.claude/rules/moai/workflow/ci-{watch,autofix}-protocol.md`, `.claude/skills/moai/workflows/sync.md`, `scripts/ci-watch/run.sh`, `internal/` Go 코드 모두 read-only 참조. 본문이 부정확하다면 별도 SPEC.

#### 3.2.6 Out of Scope — adk.mo.ai.kr Vercel 빌드 검증

본 SPEC 은 로컬 `hugo --gc --minify` exit 0 + 4-locale public/ 페이지 생성으로 검증. 실 배포 Vercel preview URL 검증은 머지 후 별도.

#### 3.2.7 Out of Scope — sync 페이지의 기존 본문 수정

기존 `## 개요` / `## 사용법` / `## 지원 모드` 등 14개 H2 섹션의 본문은 본 SPEC 이 수정하지 않음. 새 H2 한 개 삽입만이 본 SPEC scope. 기존 본문의 사실 오류가 발견되면 별도 SPEC.

#### 3.2.8 Out of Scope — `_meta.yaml` / `main.yaml` 변경

본 SPEC은 페이지 본문 추가만 다루며 navigation 데이터 (Hugo `data/menu/main.yaml`, `_meta.yaml`) 는 건드리지 않음 — 기존 `moai-sync.md` 페이지의 entry 가 이미 존재하므로 추가 navigation 변경 불필요.

---

## 4. Requirements (EARS)

### 4.1 새 섹션 삽입 (REQ-DUD-001..003)

**REQ-DUD-001 (Event-Driven)**: WHEN run-phase 실행 시, the system SHALL insert exactly one new H2 section into `docs-site/content/{ko,en,ja,zh}/workflow-commands/moai-sync.md` (4 files), with the heading text translated per locale (`PR 머지 후 CI 모니터링` / `CI monitoring after PR creation` / `PR 作成後の CI モニタリング` / `PR 创建后的 CI 监控`), positioned immediately before the existing `## 품질 게이트` (or locale equivalent) H2 section.

**REQ-DUD-002 (Event-Driven)**: WHEN inserting the new H2 section, the system SHALL include body content covering all three doctrine pillars: (a) Wave 1 polling — 30초 minimum poll interval + 30분 hard timeout per `ci-watch-protocol.md`, (b) Wave 2 auto-fix loop — 최대 3 iterations per PR push + new commits only (force-push and amend prohibited) per `ci-autofix-protocol.md`, (c) escalation — iteration ≥ 4 triggers blocking AskUserQuestion to user with no auto-resume timeout.

**REQ-DUD-003 (Event-Driven)**: WHEN inserting the new H2 section, the system SHALL include a "protected files" / "auto-fix 가 절대 건드리지 않는 파일" subsection (h3 or callout) listing at minimum: `.env`, `.env.*`, credentials 파일, `scripts/ci-watch/run.sh`, `.github/required-checks.yml`.

### 4.2 4-locale parity (REQ-DUD-004..005)

**REQ-DUD-004 (Ubiquitous)**: The system SHALL maintain 4-locale parity for the new section across all 4 files (ko/en/ja/zh) — heading translation only, body structure (h3 subsection count + table rows + callout count) MUST be identical across locales.

**REQ-DUD-005 (Unwanted)**: The system SHALL NOT introduce: (a) forbidden URLs (anything other than `adk.mo.ai.kr`, `github.com/modu-ai/moai-adk`), (b) Mermaid `flowchart LR` / `graph LR` / `BT` / `RL` directives (only `TD` / `TB` allowed per docs-site-i18n-rules §17.2), (c) emojis anywhere in the new section, (d) `draft: true` frontmatter changes, (e) references to dev-only `97-*` / `98-*` / `99-*` commands per dev-only-commands-isolation.md.

### 4.3 회귀 차단 + 빌드 검증 (REQ-DUD-006..007)

**REQ-DUD-006 (Event-Driven)**: WHEN run-phase 실행 시, the system SHALL NOT modify any file modified by PR #1045 (SPEC-V3R6-DOCS-CMD-CATALOG-001) — verified by `gh pr diff 1045 --name-only` returning 0 overlap with this SPEC's modified file set. This applies regardless of PR #1045 merge state (OPEN or MERGED).

**REQ-DUD-007 (Event-Driven)**: WHEN run-phase 완료 직전, the system SHALL execute `cd docs-site && hugo --gc --minify` with exit code 0 and verify that `docs-site/public/{ko,en,ja,zh}/workflow-commands/moai-sync/index.html` all 4 files exist and contain the new section heading text (locale-translated).

---

## 5. Edge Cases

### EC-DUD-001: PR #1045 가 본 SPEC run-phase 보다 먼저 머지될 때

- 시나리오: 본 SPEC plan PR 머지 → 본 SPEC run-phase 진입 직전 PR #1045 가 머지됨.
- 영향: 본 SPEC 의 PRESERVE list (§6.2) 에서 PR #1045 가 만든 `moai-harness.md` × 4-locale 및 `moai-gate.md` × 4-locale 는 이미 main 에 존재. 본 SPEC 은 이들을 read-only 로 가정하고 건드리지 않음 (REQ-DUD-006).
- 검증: run-phase 시작 시 `git log origin/main --oneline | grep DOCS-CMD-CATALOG` 으로 PR #1045 머지 여부 확인 + 본 SPEC 수정 대상 (`moai-sync.md` × 4) 의 PR #1045 직전 라인 수와 동일한지 확인.

### EC-DUD-002: PR #1045 가 본 SPEC run-phase 보다 늦게 머지될 때

- 시나리오: 본 SPEC 머지 (PR #X) → PR #1045 가 추후 머지.
- 영향: 본 SPEC 은 `moai-db.md` / `moai-github.md` / `moai-harness.md` / `moai-gate.md` 4 파일은 건드리지 않으므로 PR #1045 merge conflict 0건 (pre-flight 검증 §1.4 file overlap 0).
- 검증: `gh pr diff 1045 --name-only` + 본 SPEC 변경 파일 set 교집합 empty.

### EC-DUD-003: 4-locale 본문 길이 비대칭 허용 범위

- 시나리오: ko 원문 35 라인 → ja 번역 시 일본어 특성상 30-45 라인 가능 (한자/카타카나 표기 길이 차).
- 허용 범위: 새 섹션 라인 수의 locale별 편차 ±20% 이하 (ko 35 라인 기준 28-42 라인). h3 subsection 개수 + 표 행 수는 정확히 동일.
- 검증: AC-DUD-005 에서 명시.

### EC-DUD-004: 기존 sync 페이지에 `## 품질 게이트` H2 가 없는 locale 존재 시

- 시나리오: en/ja/zh 중 일부 locale 이 한국어와 다른 H2 순서 (예: ko 는 14 H2, en 은 12 H2 — 누락 가능).
- 처리: pre-flight 으로 각 locale 의 H2 list 추출 후 fallback target heading 결정 (`## 실전 예시` 또는 `## 관련 문서` 직전). 단, 4-locale 모두 동일 fallback 사용 (parity 유지).
- 검증: run-phase pre-flight 의무 — 4-locale H2 list 추출 + insertion point 결정 보고서.

### EC-DUD-005: Hugo 빌드가 인접 페이지 weight 충돌로 fail 할 가능성

- 시나리오: PR #1045 가 `moai-harness.md weight: 55` 를 추가했고, 기존 `_meta.yaml` 의 다른 entry weight 와 충돌 가능.
- 처리: 본 SPEC 은 weight 변경 없음 (REQ-DUD-005 (d) — `draft:` / `weight:` 등 frontmatter 변경 금지). Hugo 빌드 fail 시 PR #1045 책임이며 본 SPEC blocker 보고.
- 검증: AC-DUD-007 (Hugo 빌드 exit 0).

---

## 6. Dependencies

### 6.1 SPEC Dependencies (depends_on)

- **SPEC-V3R6-DOCS-CMD-CATALOG-001** (PR #1045 OPEN MERGEABLE): blocking 아님, file overlap 0건. 단 본 SPEC의 EC-DUD-001/002 시나리오에서 머지 순서에 따라 보존 verification 변동 — 본 SPEC run-phase 진입 직전 PR #1045 머지 상태 확인 의무.
- **선례 SPEC**: SPEC-V3R5-DOCS-SECURITY-001 (commit `c94d8b203`, 2026-05-20) — 4-locale security-notes.md 신설 패턴 참조 (frontmatter 패턴 + i18n 동시 업데이트 + Hugo 빌드 검증 흐름).

### 6.2 PRESERVE List (HARD — 본 SPEC scope 외, 절대 수정 금지)

#### 6.2.1 Modified (dirty) — 본 SPEC 범위 외

- `.moai/harness/usage-log.jsonl` (runtime-managed per CLAUDE.local.md §2)
- `docs-site/hugo.toml`, `docs-site/layouts/_default/baseof.html`, `docs-site/layouts/partials/menu.html` (별도 워크스트림)

#### 6.2.2 Untracked — 본 SPEC 범위 외

- `.moai/research/moai-adk-current-state-2026-05-22.md`, `.moai/research/v3.0-design-2026-05-22.md` (baseline 보고서, read-only 참조)
- `docs-site/content/{ko,en,ja,zh}/book/` (별도 워크스트림 — book 섹션 untracked)
- `docs-site/data/menu/extra.yaml`, `docs-site/scripts/gen_menu.py`, `docs-site/static/book/`, `docs-site/layouts/_default/redirect.html` (별도 워크스트림)
- `internal/hook/.moai/` (CLAUDE.local.md §2 working tree hygiene)

#### 6.2.3 PR #1045 영역 (file overlap 0건, 본 SPEC 절대 수정 금지)

- `docs-site/content/{ko,en,ja,zh}/workflow-commands/moai-db.md` (PR #1045 deletes)
- `docs-site/content/{ko,en,ja,zh}/utility-commands/moai-github.md` (PR #1045 deletes)
- `docs-site/content/{ko,en,ja,zh}/workflow-commands/moai-harness.md` (PR #1045 creates)
- `docs-site/content/{ko,en,ja,zh}/quality-commands/moai-gate.md` (PR #1045 creates)
- `docs-site/content/{ko,en,ja,zh}/{workflow,quality,utility}-commands/_meta.yaml` (PR #1045 modifies)
- `docs-site/data/menu/main.yaml` (PR #1045 modifies)

---

## 7. Risks

자세한 mitigation 은 plan.md §Risks 참조.

| Risk ID | 설명 | Severity | Mitigation 요약 |
|---------|------|----------|------------------|
| R-DUD-001 | 4-locale 동시 업데이트 중 일부 locale 누락 → §17.3 HARD 위반 | High | run-phase Self-Verification SV-6 (4-locale wc -l parity) + pre-commit 자동 차단 |
| R-DUD-002 | Hugo 빌드 fail (frontmatter syntax error / shortcode 미지원) | Medium | run-phase 종료 직전 `hugo --gc --minify` exit 0 의무 (REQ-DUD-007) |
| R-DUD-003 | PR #1045 와 file 충돌 (예상 0건 → 실제 충돌 발생) | Low | pre-flight `gh pr diff 1045 --name-only` cross-check + AC-DUD-002 |
| R-DUD-004 | 새 섹션 본문이 `ci-{watch,autofix}-protocol.md` source 와 drift | Medium | rules 본문 직접 인용 + cross-ref 링크 명시 (REQ-DUD-002) |
| R-DUD-005 | en/ja/zh 번역 품질 부족 (ko 직역 의존) | Low | locale별 native phrase 차이는 EC-DUD-003 허용 범위 내 |

---

## References

- baseline 보고서: `.moai/research/moai-adk-current-state-2026-05-22.md` §8 (사용자 가시 표면 결함 F1-F8)
- v3.0 환골탈태 설계: `.moai/research/v3.0-design-2026-05-22.md` §Wave 0 (라인 351-356)
- Rules SSoT: `.claude/rules/moai/workflow/ci-watch-protocol.md` + `.claude/rules/moai/workflow/ci-autofix-protocol.md`
- i18n doctrine: `.moai/docs/docs-site-i18n-rules.md`
- dev-only isolation: `.moai/docs/dev-only-commands-isolation.md`
- Frontmatter SSOT: `.claude/rules/moai/development/spec-frontmatter-schema.md` (12 canonical fields)
- 선례 SPEC: SPEC-V3R6-DOCS-CMD-CATALOG-001 (PR #1045 OPEN), SPEC-V3R5-DOCS-SECURITY-001 (commit `c94d8b203`)

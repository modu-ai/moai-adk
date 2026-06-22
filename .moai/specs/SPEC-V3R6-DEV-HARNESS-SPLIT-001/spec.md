---
id: SPEC-V3R6-DEV-HARNESS-SPLIT-001
title: "devkit 단일 진입을 3개 독립 harness 커맨드로 분리"
version: "0.1.0"
status: in-progress
created: 2026-06-22
updated: 2026-06-22
author: manager-spec
priority: P2
phase: "v3.0.0"
module: ".claude/commands/harness"
lifecycle: spec-anchored
tier: S
tags: "harness, dev-only, split"
depends_on: [SPEC-V3R6-DEV-HARNESS-CONSOLIDATION-001]
---

# SPEC-V3R6-DEV-HARNESS-SPLIT-001 — devkit 단일 진입을 3개 독립 harness 커맨드로 분리

## HISTORY

| Version | Date | Author | Change |
|---------|------|--------|--------|
| 0.1.0 | 2026-06-22 | manager-spec | 최초 draft — devkit 단일 진입(`/harness:devkit`)을 3개 독립 harness(`/harness:release-update`, `/harness:github`, `/harness:release`)로 분리. SPEC-V3R6-DEV-HARNESS-CONSOLIDATION-001의 *통합-진입 결정만* supersede (CONSOLIDATION-001 자체는 completed 유지). |

## §A. Context (배경)

직전 SPEC-V3R6-DEV-HARNESS-CONSOLIDATION-001 (status: `completed`, origin/main `9ef450f74`)은 3개의 dev-only 메인테이너 커맨드(`97-release-update` / `98-github` / `99-release`)를 단일 `/harness:devkit` 진입점 + sub-dispatch로 통합했다. 통합으로 전달된 자산:

- 단일 진입 thin command `.claude/commands/harness/devkit.md` (첫 인자로 capability dispatch)
- 단일 SSOT manifest `.claude/commands/harness/manifest.json` (3 specialist)
- Runner `.claude/workflows/harness-devkit-run.js` (비-상호작용 research fan-out 전용)
- 3개 specialist body `.claude/agents/harness/harness-devkit-{release-update,github,release}-specialist.md` (구 multi-phase workflow 본문을 structural fidelity로 이식)
- CI guard `internal/template/devkit_namespace_test.go` (`TestDevkitNamespaceNoLeak`)

사용자는 이제 통합-진입 결정을 **번복**한다: 3개 capability를 **3개의 독립 harness 커맨드**로 분리하여, 사용자가 `/harness:devkit release-update`가 아니라 `/harness:release-update`처럼 **직접** 호출하도록 한다.

본 SPEC은 CONSOLIDATION-001의 **통합-진입 *결정*만** supersede한다 — SPEC 전체를 supersede하지 않는다. CONSOLIDATION-001은 `completed` 상태를 유지한다 (그것이 전달한 specialist body들 + 97/98/99 삭제는 본 SPEC이 재사용한다). 본 SPEC은 `depends_on: [SPEC-V3R6-DEV-HARNESS-CONSOLIDATION-001]`로 그 의존을 기록한다.

### §A.1 Supersession-of-decision 관계 (명시)

- CONSOLIDATION-001은 **superseded로 표시하지 않는다.** 그 frontmatter `status`는 `completed` 그대로 둔다.
- 본 SPEC이 번복하는 것은 CONSOLIDATION-001의 *하나의 설계 결정* — "3 capability를 단일 `/harness:devkit` 진입으로 통합한다" — 뿐이다.
- 재사용하는 것: 3개 specialist body의 이식된 workflow 본문, 97/98/99 번호 커맨드 삭제, dev-only 격리 doctrine, embedded-tree-absence CI guard 패턴.

## §B. Goal (목표)

3 capability를 **3개 독립 harness 커맨드**로 분리한다 — 완전 분리 + 이름 변경.

- 진입: `/harness:release-update`, `/harness:github`, `/harness:release` (직접 호출)
- specialist 이름 변경: `harness-devkit-{X}-specialist.md` → `harness-{X}-specialist.md`
- 통합 자산 제거: `.claude/commands/harness/devkit.md` + `.claude/commands/harness/manifest.json`
- CI guard + doctrine를 단일 `harness-devkit` namespace에서 3개 분리 namespace로 갱신

### §B.1 사용자 지정 per-command surface (그대로 채택)

| Harness | 진입 + 인자 surface |
|---------|---------------------|
| release-update | `/harness:release-update [--since vX.Y.Z \| --dry]` |
| github | `/harness:github issues\|pr [...]` |
| release | `/harness:release [VERSION] [--hotfix]` |

### §B.2 핵심 설계 결정 — Runner 비대칭성 (Runner asymmetry)

3개 capability 중 **`release-update`만** 진짜 비-상호작용 research fan-out(CC release-notes version-delta sweep)을 갖는다. `github`과 `release`는 순수 human-gated/interactive specialist이며 fan-out이 **없다**. 따라서 정직한(honest) v4 설계는 대칭을 강제하지 않고 비대칭을 명시한다:

- **release-update harness**: thin command + `.claude/commands/harness/release-update/manifest.json` + `.claude/workflows/harness-release-update-run.js` Runner + `harness-release-update-specialist`.
- **github + release harness**: thin command → specialist 직접 라우팅 (Runner 없음, manifest 없음 — 모델에 fan-out할 대상 자체가 없음). thin command는 `Use the harness-<name>-specialist subagent` 형태로 specialist에 라우팅.

이 비대칭은 단순성(Enforce Simplicity) 원칙에 부합한다: 없는 fan-out을 위해 Runner/manifest를 만들지 않는다. github/release에 최소 per-harness manifest를 두는 것이 일관성 측면에서 가치가 있다고 판단되면 정당화할 수 있으나, 기본값은 두 순수-specialist harness에 대해 **no-Runner / no-manifest** (단순성 우선)이다.

## §C. Requirements (EARS/GEARS)

### REQ-DHS-001 — 3개 독립 thin command 진입

The harness command system **shall** provide three independent thin command entries at `.claude/commands/harness/{release-update,github,release}.md`, each routing directly to its specialist (and, for release-update, to its Runner), with the §B.1 per-command flag/subcommand surface declared in the body and `argument-hint`.

- **When** a maintainer invokes `/harness:release-update [--since vX.Y.Z | --dry]`, the release-update thin command **shall** route to the release-update Runner / `harness-release-update-specialist`.
- **When** a maintainer invokes `/harness:github issues|pr [...]`, the github thin command **shall** route to `harness-github-specialist` directly (no Runner).
- **When** a maintainer invokes `/harness:release [VERSION] [--hotfix]`, the release thin command **shall** route to `harness-release-specialist` directly (no Runner).

각 thin command는 coding-standards.md § Thin Command Pattern에 따라 20 LOC 미만 본문 + `description` / `argument-hint` / `allowed-tools` frontmatter를 갖는다.

### REQ-DHS-002 — specialist 이름 변경 (body verbatim 보존)

The harness specialist files **shall** be renamed `harness-devkit-{release-update,github,release}-specialist.md` → `harness-{release-update,github,release}-specialist.md`, preserving the ported workflow body verbatim except for:

- frontmatter `name` 갱신 (`harness-devkit-X-specialist` → `harness-X-specialist`),
- 자기-참조(self-reference) 갱신 (진입 커맨드 `/harness:devkit X` → `/harness:X`; manifest role 참조; Runner 파일명),
- Migration Provenance 섹션에 본 SPEC(SPEC-V3R6-DEV-HARNESS-SPLIT-001) 인용 추가.

`[DEV-ONLY]` 배너는 유지된다.

### REQ-DHS-003 — 통합 진입 + Runner 재범위화

The harness command system **shall** remove the unified `devkit.md` entry and the unified `manifest.json`, and **shall** re-scope the Runner to release-update only.

- 통합 `.claude/commands/harness/devkit.md` 제거.
- 통합 `.claude/commands/harness/manifest.json` 제거.
- Runner는 처음부터 release-update의 fan-out만 모델링했으므로, `.claude/workflows/harness-devkit-run.js` → `.claude/workflows/harness-release-update-run.js`로 이름 변경하고 release-update 전용으로 재범위화한다 (manifest 경로 참조 갱신 포함). github/release는 Runner를 갖지 않는다.

### REQ-DHS-004 — CI guard 재지향 + stale `devkit` 토큰 제거 (RED → GREEN)

The CI guard **shall** be re-pointed from the single `harness-devkit` namespace to the three separate namespaces, asserting that NONE of the three split harness artifact shapes leak into the embedded template tree while STILL asserting `.claude/commands/harness/` and `.claude/workflows/harness-*` absence. The guard's own `devkit` identifiers (filename, function, sentinel) **shall** be renamed — this is MANDATORY, not optional, because the split's core purpose is zero stale `devkit` tokens and the guard currently carries them in its filename/function/sentinel.

- [MANDATORY] 테스트 파일 이름 변경: `internal/template/devkit_namespace_test.go` → `internal/template/split_namespace_test.go` (`git mv`로 이력 보존).
- [MANDATORY] 테스트 함수 이름 변경: `TestDevkitNamespaceNoLeak` → `TestSplitHarnessNamespaceNoLeak`.
- [MANDATORY] sentinel 이름 변경: `DEVKIT_NAMESPACE_LEAK` (×4 occurrences) → `SPLIT_HARNESS_NAMESPACE_LEAK`.
- 단언 갱신: embedded tree에 `harness-release-update*`, `harness-github*`, `harness-release*` agent/workflow 부재 + `.claude/commands/harness/` 경로 부재를 단언 (3개 분리 namespace 보호 — 기존 path-based 보호 유지).
- RED→GREEN: 누출 심으면 FAIL, 제거하면 PASS.
- 갱신 후 `grep -rn 'devkit' internal/template/*namespace_test.go`가 EMPTY여야 한다 (stale 파일명/함수/sentinel 토큰 잔존 0).

### REQ-DHS-005 — dev-only 누출 invariant 보존

The embedded template tree **shall** contain none of: `.claude/commands/harness/`, `harness-{release-update,github,release}*` agent files, `.claude/workflows/harness-*` files.

`find internal/template/templates` 기준 위 셋 모두 비어있어야 한다.

### REQ-DHS-006 — doctrine 갱신 (5개 surface)

The dev-only doctrine surfaces **shall** be updated, replacing every `harness:devkit` / `harness-devkit` reference **AND every standalone `devkit` reference in section/file titles and headers** with the three separate harness names, adding a migration note crediting this SPEC, and removing now-stale unified-devkit references.

대상 5개 surface:
1. `.moai/docs/dev-only-commands-isolation.md`
2. `CLAUDE.local.md` §2 (Local-Only Files) + §21 stub
3. `.claude/rules/moai/development/skill-authoring.md` (Deprecated Skill Slots table)
4. `.claude/skills/moai-foundation-core/modules/INDEX.md`
5. `.claude/skills/moai/references/reference.md`

[MANDATORY] standalone `devkit` (the `harness:devkit`/`harness-devkit` 패턴에 매칭되지 않는 단독 단어) 가 남는 위치 명시 갱신:
- `.moai/docs/dev-only-commands-isolation.md` 파일 title (L1 `# Dev-Only Commands Isolation (devkit Harness)`) → 3개 분리 harness를 반영하도록 reword (예: `# Dev-Only Commands Isolation (Split Harnesses)`).
- `CLAUDE.local.md` §21 header (`## 21. Dev-Only Commands Isolation (devkit Harness)`) → 동일 reword.

[MANDATORY] `.moai/docs/dev-only-commands-isolation.md`의 단일 devkit 배포-금지 표 행을 **3개 행으로 확장** (split harness 별 1개 행: release-update / github / release — 각 thin command + specialist + (release-update만) Runner/manifest 경로). 통합 devkit.md/manifest/devkit-run.js 행은 제거.

### REQ-DHS-007 — 메모리 갱신 보류 (sync-phase)

The memory index update **shall** be deferred to sync-phase. plan.md에 sync-phase 메모리 갱신을 명기하되, plan-phase에서는 메모리를 작성하지 않는다.

## §D. Constraints (제약)

- **Enforce Simplicity (Tier S)**: 정확히 3개 harness, 불필요한 추상화 없음. github/release는 정당화 없이는 Runner 없음.
- **Scope discipline**: devkit→3-split 재구성 + 그 CI guard + doctrine만. Layer B `agents/harness/{cli-template,hook-ci,quality,workflow}` 또는 사용자-대상 template은 건드리지 않음. CONSOLIDATION-001의 닫힌(closed) 산출물 재수정 금지.
- **Template-First / neutrality**: `internal/template/templates/` 하위에 `commands/harness/`, `harness-{release-update,github,release}*`, `workflows/harness-*` 누출 금지.
- **Frontmatter**: 12 required fields + `tier: S`; `module: ".claude/commands/harness"`; `priority: P2`; `phase: "v3.0.0"`; `tags: "harness, dev-only, split"`; `depends_on: [SPEC-V3R6-DEV-HARNESS-CONSOLIDATION-001]`.

## §E. Exclusions (범위 밖)

본 섹션은 본 SPEC이 **build하지 않는** 것을 명시한다.

### Out of Scope — Layer B harness specialists

- `.claude/agents/harness/{cli-template,hook-ci,quality,workflow}` 등 무관한 Layer B specialist는 건드리지 않는다 (이름 변경 / 삭제 / 수정 모두 금지).
- Layer B harness의 manifest / Runner / 커맨드는 본 SPEC 범위 밖이다.

### Out of Scope — CONSOLIDATION-001 closed artifacts

- SPEC-V3R6-DEV-HARNESS-CONSOLIDATION-001의 닫힌 산출물(이미 merged된 specialist 본문 외 결과물, frontmatter)을 재수정하지 않는다.
- CONSOLIDATION-001을 `superseded`로 표시하지 않는다 — `completed` 유지.
- 97/98/99 번호 커맨드를 부활시키지 않는다 (이미 삭제됨).

### Out of Scope — user-facing template assets

- 사용자-대상 `internal/template/templates/**` 자산은 수정하지 않는다 (dev-only 격리 invariant는 유지만 한다).
- harness 자산의 사용자 프로젝트 배포(template화)는 금지이며 본 SPEC이 도입하지 않는다.

### Out of Scope — github/release Runner & manifest (default)

- 기본 설계상 github / release harness는 Runner와 manifest를 갖지 않는다 (fan-out 부재 → 단순성 우선). 이를 도입하는 것은 범위 밖이며, 일관성 정당화가 plan/run에서 명시될 때만 예외적으로 검토한다.

### Out of Scope — harness 실행 / 동작 변경

- 본 SPEC은 plan-phase 산출물 authoring이다. 실제 harness/command/Runner 파일 생성·수정은 run-phase에서 manager-develop이 수행한다.
- 3개 capability의 **워크플로우 로직(동작)** 자체는 변경하지 않는다 — body는 verbatim 보존이고, 변경은 진입 라우팅 / 이름 / Runner 범위뿐이다.

### Out of Scope — DIVECC ROADMAP dangling ref

- `.moai/specs/SPEC-DIVECC-HOOK-FAILURE-MODE-AUDIT-001/ROADMAP.md` (L37, L88, L91)가 `/harness:devkit list` (split 이후 존재하지 않을 미래의 통합 inventory surface)를 참조한다. 이는 **KNOWN out-of-scope dangling ref**이며 본 SPEC이 sweep하지 않는다 — 해당 DIVECC SPEC의 author가 자체적으로 reconcile해야 한다 (DIVECC ROADMAP은 별도 SPEC의 미래 설계 문서). 따라서 "no stale harness-devkit reference" anti-pattern은 본 SPEC의 6개 doctrine surface로 honestly bounded되며, DIVECC ROADMAP은 그 경계 밖이다.

## §F. References (참조)

- `.moai/specs/SPEC-V3R6-DEV-HARNESS-CONSOLIDATION-001/` — 직전 통합 SPEC (depends_on)
- `.claude/skills/moai/workflows/harness-builder.md` — v4 artifact shapes (command / manifest / Runner / specialist)
- `.moai/docs/harness-namespace-doctrine.md` (CLAUDE.local.md §24) — harness namespace user-owned 정책
- `.moai/docs/dev-only-commands-isolation.md` — dev-only 격리 doctrine
- `.claude/rules/moai/development/coding-standards.md` § Thin Command Pattern
- `.claude/rules/moai/development/agent-authoring.md` § Agent Directory Convention (`.claude/agents/harness/` user-owned)
- `internal/template/devkit_namespace_test.go` — 재지향 대상 CI guard

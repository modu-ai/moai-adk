---
id: SPEC-V3R6-DEV-HARNESS-CONSOLIDATION-001
title: "dev-only 메인테이너 3-커맨드(97/98/99)를 단일 devkit 하네스로 통합"
version: "1.0.0"
status: completed
created: 2026-06-22
updated: 2026-06-22
author: manager-spec
priority: P2
phase: "v3.0.0"
module: ".claude/commands/harness"
lifecycle: spec-anchored
tags: "harness, dev-only, consolidation"
tier: S
---

# SPEC-V3R6-DEV-HARNESS-CONSOLIDATION-001 — dev-only 메인테이너 도구 단일 하네스 통합

## HISTORY

| Version | Date | Author | Change |
|---------|------|--------|--------|
| 0.1.0 | 2026-06-22 | manager-spec | plan-phase 최초 작성 (draft). 97/98/99 dev-only 커맨드를 단일 `devkit` 하네스로 통합하는 Tier S envelope. v4 Builder가 run-phase 내부에서 실행. |
| 0.2.0 | 2026-06-22 | manager-spec | plan-auditor iter-2 defect 반영. D1(BLOCKING): REQ-DHC-005/§B re-anchor — dev-only 보호 메커니즘을 `embedded_namespace_test.go` 패턴의 embedded-tree-absence 단언으로 정정(`dev_only_skill_test.go`는 skills-only walker라 commands/agents/workflows artifact를 보호 못 함). D2: `tier: S` frontmatter 추가. D6: REQ-DHC-006 scope에 CLAUDE.local.md §2 추가. D3/D4/D5/D7/D8: 부수 정정. |

## §A. Context (배경)

moai-adk-go 메인테이너는 세 개의 dev-only 슬래시 커맨드를 사용한다:

- `97-release-update` — Claude Code 업스트림 변경 추적기 (`.claude/agents/local/release-update-specialist.md` body)
- `98-github` — GitHub issue/PR 워크플로우 (`.claude/agents/local/github-specialist.md` body)
- `99-release` — production release 워크플로우 (`.claude/skills/moai/workflows/release.md` body)

이 세 커맨드는 모두 메인테이너 전용이며 사용자 프로젝트에 절대 배포되지 않는다 (격리 doctrine: `.moai/docs/dev-only-commands-isolation.md`). 번호 prefix(97/98/99)는 격리 가독성을 위한 임시 관례였으나, 세 개의 분리된 진입점은 (a) 메인테이너에게 세 개의 별도 UI 항목으로 노출되고, (b) 격리 보호(CI guard)가 세 개의 개별 grep 패턴에 분산되며, (c) `harness-v4` 자체-도그푸딩(self-harness) 기회를 놓친다.

이 SPEC은 세 dev-only 커맨드를 **단일 통합 dev-maintainer 하네스(`devkit`)**로 통합한다. 통합은 `harness-v4` Builder(ANALYZE/PLAN/GENERATE/ACTIVATE)를 통해 수행되며, 이 Tier S SPEC은 그 Builder 실행을 감싸는 추적 envelope이다.

### §A.1 사용자 확정 결정 (재논쟁 금지 — 채택된 접근)

다음은 Socratic 라운드에서 사용자가 이미 확정한 결정이다. SPEC은 이를 채택된 접근으로 인코딩한다:

- **구조**: 단일 통합 dev-maintainer 하네스 (`/harness:devkit` 단일 진입점, 3개 capability). 세 개의 분리 하네스가 아님.
- **빌드 방법**: 전체 `harness-v4` Builder (ANALYZE/PLAN/GENERATE/ACTIVATE → manifest.json + Runner `.js` + specialist 서브에이전트 + 동반 skill + thin command).
- **추적**: 이 Tier S SPEC이 추적 envelope이며, v4 Builder는 그 run-phase 내부에서 실행된다.

### §A.2 하네스 이름 결정: `devkit` (정당화)

도메인 "moai-adk-go dev-only maintainer tooling"으로부터 하네스 `<name>`을 도출한다. 후보: `maintainer`(10자), `devkit`(6자), `maintain`(8자).

채택: **`devkit`**.

- 간결(6자 ≤ 32 한계), lowercase, 단일 토큰 → DNS-safe, 하이픈 충돌 없음.
- `/harness:devkit` 진입 커맨드로 자연스럽게 읽힘.
- "developer/maintainer 도구 키트"를 신호 → 3-capability 번들(release-update + github + release)과 정확히 일치.
- 기존 `.claude/agents/harness/` Layer B specialist(`cli-template-specialist`, `hook-ci-specialist`, `quality-specialist`, `workflow-specialist`)와 충돌 불가 — 신규 dev-maintainer specialist는 `harness-devkit-*-specialist.md`로 prefix되므로 이름 공간이 완전히 분리됨.

## §B. Verified Ground-Truth (조사 완료 — 재조사 금지)

- 커맨드 존재: `.claude/commands/97-release-update.md`, `98-github.md`, `99-release.md` (전부 thin wrapper).
- 워크플로우 body 이미 마이그레이션됨: SPEC-V3R6-LOCAL-NAMESPACE-CONSOLIDATION-001(status: completed, 2026-05-25)이 97/98 body를 `.claude/agents/local/{release-update,github}-specialist.md`로 이동. 99-release body는 `.claude/skills/moai/workflows/release.md`에 존재.
- v4 하네스 아직 없음: `.claude/commands/harness/` 부재, `.claude/workflows/harness-*.js` Runner 부재.
- `.claude/agents/harness/`는 무관한 set(cli-template/hook-ci/quality/workflow specialist — moai-adk-go 코드베이스 Layer B Pipeline 팀)이 이미 점유 중. 신규 dev-maintainer specialist는 이들과 이름 충돌 금지.
- 템플릿 누출 검사 현재 clean (97/98/99/release-update/github/release artifact가 `internal/template/templates/` 아래에 없음).
- dev-only 격리 doctrine: `.moai/docs/dev-only-commands-isolation.md`. "배포 금지 파일 목록" 표 + `find internal/template/templates -name "9{7,8,9}-*"` grep HARD 체크리스트(이 체크리스트는 doc-level 수동 검증이며 CI 자동 gate가 아님) + LOCAL-NAMESPACE-CONSOLIDATION 마이그레이션 노트 보유. CLAUDE.local.md §21은 이 파일을 가리키는 stub.
- **[D1 정정] dev-only artifact의 실제 CI 보호 메커니즘은 "embedded-tree 부재(absence-from-embedded-tree)"이다.** 다음 테스트가 `EmbeddedTemplates()` 트리를 walk/ReadDir하여 누출을 검출한다 (`internal/template/embedded_namespace_test.go` `TestTemplateAgentsStructure`: `.claude/agents/`에 `{moai}` subdir만 허용 → `agents/harness/` 등장 시 `HARNESS_NAMESPACE_LEAK` FAIL; `commands_audit_test.go`: `.claude/commands` 전수 walk; `namespace_protection_audit_test.go`). **97/98/99 + `agents/local/*`는 애초에 `dev_only_skill_test.go`로 보호된 적이 없다** — `dev_only_skill_test.go`의 `TestDevOnlySkillLeak`은 `.claude/skills/`만 walk하고 하드코딩된 skill-name map에 base name을 매칭하는 **skills-only walker**이므로, commands/agents/workflows artifact(`commands/harness/`, `agents/harness/`, `workflows/`)는 그 테스트의 보호 범위 밖이다.
- 따라서 REQ-DHC-005의 CI guard 확장은 `dev_only_skill_test.go` 확장이 아니라, `embedded_namespace_test.go`를 모델로 한 **신규 embedded-tree-absence 단언**으로 앵커링한다 (devkit artifact가 `internal/template/templates/`에 누출되지 않음을 단언; RED는 embedded 트리 하위에 `harness-devkit` 경로를 심어 walker가 검출하도록 함).
- CI guard 워크플로우: `.github/workflows/template-neutrality-check.yaml` (neutrality 패턴은 `TestTemplateNeutralityAudit` 키, dev-only 누출과는 별개 패턴 set).

## §C. Requirements (GEARS notation)

### REQ-DHC-001 — 단일 devkit 하네스 생성

The harness builder **shall** v4 Builder(ANALYZE/PLAN/GENERATE/ACTIVATE)를 통해 단일 통합 dev-maintainer 하네스 `devkit`을 생성한다. 진입 커맨드 `/harness:devkit`은 `.claude/commands/harness/devkit.md`에 위치하고, SSOT manifest는 `.claude/commands/harness/manifest.json`(또는 `.claude/harness/devkit/manifest.json`)에 위치하며 하네스별 고정 위치로 entry command에 기록된다. 하네스는 정확히 3개 capability를 가진다: release-update(CC tracker), github(issue/PR), release(production).

### REQ-DHC-002 — specialist 콘텐츠 소스 재사용 (구조적 충실도 포팅)

The GENERATE phase **shall** 이미 마이그레이션된 `.claude/agents/local/release-update-specialist.md`, `.claude/agents/local/github-specialist.md` body와 `.claude/skills/moai/workflows/release.md` body를 specialist 콘텐츠 소스로 재사용한다. 9-phase/multi-phase 워크플로우 로직을 처음부터 다시 작성하지 않고, 구조적 충실도(structural fidelity)를 유지하며 하네스 specialist 파일 `.claude/agents/harness/harness-devkit-*-specialist.md`로 포팅한다.

### REQ-DHC-003 — 번호 커맨드 삭제 + 구 소스 정리

**When** devkit 하네스가 live 상태가 되면, the run-phase **shall** 세 번호 커맨드 `.claude/commands/97-release-update.md`, `98-github.md`, `99-release.md`를 삭제한다. 구 `agents/local/*-specialist.md` + `skills/moai/workflows/release.md`의 처분은 **삭제 권장**(dual-source drift 방지) — 포팅 완료 후 삭제하며, 그 결정은 plan.md에 문서화한다.

### REQ-DHC-004 — 신규 하네스 namespace 하에서 dev-only 격리 보존

The devkit 하네스 artifacts (`.claude/commands/harness/devkit*`, `.claude/workflows/harness-devkit-run.js`, `.claude/agents/harness/harness-devkit-*`, `.claude/skills/harness-devkit-*/`) **shall not** `internal/template/templates/`로 누출된다. `harness-namespace-doctrine.md` §24의 user-owned 보호(`moai update`가 보존)를 따른다.

### REQ-DHC-005 — CI guard 확장 (devkit namespace embedded-tree 부재 단언)

The CI guard **shall** dev-only artifact 누출 검사가 신규 `harness-devkit` namespace를 커버하도록 확장된다. [D1 정정] 확장 대상은 `dev_only_skill_test.go`(skills-only walker라 commands/agents/workflows artifact를 보호하지 못함)가 **아니라**, `internal/template/embedded_namespace_test.go`를 모델로 한 **신규 embedded-tree-absence 단언**이다. The new test **shall** `EmbeddedTemplates()` 트리(또는 `internal/template/templates/`)에 다음이 모두 부재함을 단언한다: (a) `.claude/commands/harness/` 경로 부재, (b) `harness-devkit*` agent 파일 부재, (c) `.claude/workflows/harness-devkit-*` 파일 부재. RED 케이스는 walker가 실제 검출할 수 있는 누출 형태(embedded 트리 하위에 심은 `harness-devkit` 경로)를 주입하여 FAIL을 확인하고, 제거 후 PASS(GREEN)를 확인한다. 본 단언은 `TestTemplateAgentsStructure`의 `{moai}`-only allowlist 패턴(`agents/harness/` 등장 시 `HARNESS_NAMESPACE_LEAK` FAIL)과 정합하며, 그 allowlist가 이미 `agents/harness/` 부재를 보호하므로 신규 단언은 commands/workflows 차원 + `harness-devkit` 이름 차원을 보완한다.

### REQ-DHC-006 — doctrine 업데이트 (3개 surface)

The sync-phase **shall** 다음 3개 surface를 업데이트한다:

1. `.moai/docs/dev-only-commands-isolation.md` — "배포 금지 파일 목록" 표 + 검증 체크리스트를 harness-namespaced dev-only 도구(`commands/harness/devkit*`, `workflows/harness-devkit-run.js`, `agents/harness/harness-devkit-*`)를 반영하도록 교체/확장하고, 이 SPEC을 credit하는 마이그레이션 노트를 추가하며, SPEC-V3R6-LOCAL-NAMESPACE-CONSOLIDATION-001 노트와 정합한다. doctrine은 신규 하네스 artifacts + 번호-커맨드 제거를 명명한다.
2. CLAUDE.local.md §21 stub — pointer 텍스트 조정(필요 시).
3. **[D6 추가] CLAUDE.local.md §2 "Local-Only Files (Never in Templates)"** — §2 목록은 `97-release-update.md`, `98-*.md`, `.claude/skills/moai/workflows/release-update.md`를 dev-only 항목으로 독립적으로 명명한다. M5가 97/98/99 + 구 body를 삭제한 후 이 목록은 stale해진다. sync-phase는 §2 목록에서 삭제된 항목을 제거하고 신규 `commands/harness/devkit*` + `workflows/harness-devkit-run.js`를 추가한다. (대안: 명시적 carve-out + rationale.) **주의 — false positive**: `git-workflow-doctrine.md`의 `release.md` 언급은 GitHub-Actions `release.yml`/Release Drafter를 가리키는 false positive이므로 **건드리지 않는다**.

### REQ-DHC-007 — Runner / human-gate 정합 (핵심 설계 제약)

The devkit Runner (`harness-devkit-run.js`, dynamic-workflow 스크립트) **shall** 비-상호작용(non-interactive) fan-out 부분만 모델링한다. **Where** capability가 사람-게이트/상호작용이 필요한 경우(release-update의 WebSearch+WebFetch 리서치 sweep + 사용자 승인 + PR; github의 Agent Teams + gh CLI 상호작용; release의 production human gate), the orchestrator **shall** workflow를 launch하기 **전에** 모든 AskUserQuestion 게이트를 보유하며, 상호작용/사람-게이트 작업은 specialist 서브에이전트로 위임한다. Runner는 사용자에게 프롬프트하지 않는다(비대칭 boundary per `.claude/rules/moai/core/agent-common-protocol.md` §User Interaction Boundary).

## §D. Constraints (제약)

- Enforce Simplicity: Tier S. SPEC을 최소로 유지 — 추가 추상화 금지. 하네스는 정확히 3 capability를 가지며 더 만들지 않는다.
- Scope discipline: 97/98/99 dev 도구 + 그 doctrine/CI/memory만 건드린다. 무관한 `.claude/agents/harness/` Layer B specialist 또는 사용자-facing 템플릿은 절대 건드리지 않는다.
- Frontmatter: schema SSOT의 12 required 필드 정확히. `module: ".claude/commands/harness"`; `priority: P2`; `phase: "v3.0.0"`; `tags`에 "harness, dev-only, consolidation" 포함.
- 하네스 artifacts는 user-owned namespace(`harness-*` / `.claude/agents/harness/` / `.claude/commands/harness/` / `.claude/workflows/harness-*.js`)에 위치 — `moai update`가 보존한다.

## §E. Self-Verification / Audit-Ready Signal (plan-phase)

이 SPEC의 plan-phase audit-ready 신호는 `progress.md` §E.1에 기록된다. run/sync-phase 증거는 §E.2~§E.4(placeholder)에 후속 채워진다.

## §J. Exclusions (이 SPEC이 만들지 않는 것)

이 섹션은 scope 경계를 명시한다. 각 제외 항목은 `### Out of Scope — <topic>` H3로 표현된다.

### Out of Scope — Go 코드 변경

- 이 SPEC은 `internal/` 또는 `pkg/`의 Go 구현 변경을 포함하지 않는다 (CI guard 테스트 `dev_only_skill_test.go` 확장은 예외이며, 이는 dev-only 누출 보호 테스트의 패턴 추가일 뿐 기능 변경이 아니다).
- `internal/cli/update.go` 등의 `moai update` namespace 보호 enforcement는 이미 SPEC-V3R6-HARNESS-NAMESPACE-V2-001(§24.5 RESOLVED)에서 `harness-*`로 전환 완료되어 있으므로 이 SPEC에서 다시 건드리지 않는다.

### Out of Scope — 사용자 배포 템플릿

- `internal/template/templates/**` 아래 어떤 파일도 추가/수정하지 않는다. devkit 하네스는 dev-only이며 user-owned namespace에 위치하므로 템플릿 트리에 절대 등장하지 않는다 (REQ-DHC-004).
- 16-language neutrality, 사용자-facing skill/command/agent는 이 SPEC의 scope 밖이다.

### Out of Scope — capability 확장

- 하네스는 정확히 3 capability(release-update/github/release)만 가진다. swarm status/done/kill-all, CI auto-fix 통합, 추가 메인테이너 워크플로우 같은 새 capability는 만들지 않는다 (Enforce Simplicity).
- 9-phase/multi-phase 워크플로우 로직의 재설계 또는 기능 향상은 scope 밖 — 기존 body를 구조적 충실도로 포팅만 한다 (REQ-DHC-002).

### Out of Scope — memory 직접 작성

- memory 업데이트는 sync-phase deliverable이며 manager-docs가 수행한다. manager-spec(이 SPEC 작성자)은 memory를 직접 작성하지 않는다 (plan.md §F에 sync task로 노트만 남긴다).

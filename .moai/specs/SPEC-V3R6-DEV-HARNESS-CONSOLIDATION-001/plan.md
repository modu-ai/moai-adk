# SPEC-V3R6-DEV-HARNESS-CONSOLIDATION-001 — Implementation Plan

> Tier S (minimal). 이 plan은 run-phase에서 `harness-v4` Builder가 실행하는 작업의 milestone 순서와 핵심 설계 결정을 정의한다. 시간 추정은 사용하지 않는다(priority 기반).

## §A. Context

run-phase는 `harness-v4` Builder(orchestrator-direct, `.claude/skills/moai/workflows/harness-builder.md`)를 실행하여 단일 `devkit` 하네스를 생성하고, 그 결과 97/98/99 번호 커맨드를 삭제한 뒤 dev-only 격리(CI guard + doctrine)를 신규 namespace로 이전한다. 이 Tier S SPEC은 그 Builder 실행을 감싸는 추적 envelope이다.

핵심 사실(spec.md §B에서 확정):
- 97/98 워크플로우 body는 이미 `.claude/agents/local/{release-update,github}-specialist.md`로 마이그레이션됨 — specialist 콘텐츠 소스.
- 99 워크플로우 body는 `.claude/skills/moai/workflows/release.md`에 존재 — specialist 콘텐츠 소스.
- `.claude/agents/harness/`는 무관한 Layer B set(cli-template/hook-ci/quality/workflow)이 점유 — 신규 specialist는 `harness-devkit-*` prefix로 충돌 회피.
- **[D1 정정] dev-only artifact의 실제 embedded-tree 보호 메커니즘은 `internal/template/embedded_namespace_test.go`(`TestTemplateAgentsStructure` — `.claude/agents/`에 `{moai}`-only allowlist, `agents/harness/` 등장 시 `HARNESS_NAMESPACE_LEAK` FAIL) + `commands_audit_test.go`(`.claude/commands` 전수 walk) + `namespace_protection_audit_test.go`이다.** `dev_only_skill_test.go`는 `.claude/skills/`만 walk하는 skills-only walker이므로 commands/agents/workflows artifact를 보호하지 못한다 — M6 확장은 이 테스트가 아니라 embedded-tree-absence 단언으로 앵커링한다.

## §B. Known Issues / Load-Bearing Design Risk

### B.1 Runner / human-gate 정합 (핵심 — 손쉽게 넘어가지 말 것)

`harness-v4` Builder의 GENERATE는 dynamic-workflow Runner(`harness-devkit-run.js`)를 emit한다. **Dynamic workflow는 mid-run에 AskUserQuestion을 호출할 수 없다**(비대칭 boundary, `.claude/rules/moai/core/agent-common-protocol.md` §User Interaction Boundary; `.claude/rules/moai/workflow/dynamic-workflows.md`). 그러나 3개 capability 전부가 사람-게이트/상호작용이다:

- **release-update**: WebSearch+WebFetch 리서치 sweep + 사용자 승인 + PR 생성.
- **github**: Agent Teams + gh CLI + issue/PR 상호작용.
- **release**: production release 사람 게이트.

**채택된 정합 설계 (명시적 결정)**:

1. **Runner는 비-상호작용 fan-out 부분만 모델링한다.** 가장 명확한 fan-out 후보는 release-update의 CC-release-notes 리서치 sweep(여러 버전 델타를 병렬로 분석 → 집계). 이것은 read-only fan-out이므로 dynamic-workflow에 적합하다.
2. **모든 상호작용/사람-게이트 작업은 specialist 서브에이전트로 위임한다.** specialist 콘텐츠 소스는 이미 마이그레이션된 `agents/local/*-specialist.md` body + `release.md` body. 사용자 승인, PR 생성, gh CLI 상호작용, production release 게이트는 전부 specialist(또는 orchestrator-direct)가 보유.
3. **orchestrator는 어떤 workflow를 launch하기 전에 모든 AskUserQuestion 게이트를 보유한다.** Runner가 사용자에게 프롬프트할 수 있는 척하지 않는다.

이 제약은 명시적 설계 결정(spec.md REQ-DHC-007)이며, acceptance.md에 검증 가능한 AC(AC-DHC-007a/b)로 인코딩된다.

### B.2 namespace 충돌 회피

`.claude/agents/harness/`에 이미 4개 Layer B specialist가 존재한다. devkit specialist는 `harness-devkit-*-specialist.md`로 명명하여 충돌을 회피한다. PLAN phase는 GENERATE 전에 이 prefix를 검증해야 한다.

## §C. Pre-flight Checks (run-phase 진입 전)

- [ ] `.claude/commands/harness/devkit.md` 부재 확인 (이름 충돌 없음)
- [ ] `.claude/agents/harness/harness-devkit-*` 부재 확인 (Layer B set와 충돌 없음)
- [ ] `.claude/agents/local/{release-update,github}-specialist.md` + `.claude/skills/moai/workflows/release.md` 존재 확인 (specialist 콘텐츠 소스)
- [ ] `internal/template/embedded_namespace_test.go` + `commands_audit_test.go` 현재 GREEN 확인 (M6 embedded-tree-absence 단언 확장 baseline — `dev_only_skill_test.go` 아님, D1)

## §D. Constraints

- Tier S — 최소. 3 capability 고정, 추가 추상화 금지.
- Scope: 97/98/99 dev 도구 + doctrine/CI/memory만. Layer B specialist, 사용자 템플릿 불가침.
- 모든 하네스 artifact는 user-owned namespace(`moai update` 보존).

## §E. Self-Verification (plan-phase audit-ready)

plan-phase audit-ready 신호는 `progress.md` §E.1에 기록. run-phase 진입 시 manager-develop가 §E.2/§E.3 채움, sync-phase에서 manager-docs가 §E.4 채움.

## §F. Milestones (priority 순서, 시간 추정 없음)

v4 Builder의 4 phase(ANALYZE/PLAN/GENERATE/ACTIVATE)에 매핑한 run-phase milestone:

### M1 — ANALYZE (Priority High)

orchestrator parallel `Agent(Explore, effort:low)` fan-out (read-only, main tree):
- 3개 dev-only 커맨드 + 그 body 소스(`agents/local/*-specialist.md`, `release.md`) 추출 → domain profile.
- 기존 `.claude/agents/harness/` set 점검(충돌 회피용 이름 인벤토리).
- dev-only 격리 doctrine + CI guard 테스트 점검.
- 산출: dev-maintainer domain profile + task-pattern 인벤토리(release-update=Fan-out/Fan-in 리서치 sweep; github=Supervisor+Expert Pool; release=Pipeline).

load-bearing minimum: 도메인이 이미 잘 알려졌으므로 ANALYZE를 단일 Explore로 축소 가능 — 축소 시 rationale 기록.

### M2 — PLAN + 승인 게이트 (Priority High)

orchestrator가 단일 `Agent(opus, xhigh)` spawn:
- 6-pattern catalog에서 패턴 선택(예: Pipeline 전체 흐름 + Fan-out/Fan-in의 release-update 리서치 stage).
- specialist roster 정의: `harness-devkit-release-update-specialist`, `harness-devkit-github-specialist`, `harness-devkit-release-specialist` (3 capability ↔ 3 specialist).
- 각 specialist의 primitive 매핑(sub-agent / dynamic-workflow / `/goal` / adversarial-fan-out) + isolation + effort + model.
- B.1 정합 설계 인코딩: Runner는 비-상호작용 fan-out만; 사람-게이트는 specialist 위임.
- draft manifest emit(8 top-level 필드).
- **orchestrator가 PLAN→GENERATE 경계에서 AskUserQuestion 승인 게이트 실행**(first-class — orchestrator가 boundary 보유).

### M3 — GENERATE (Priority High)

orchestrator fan-out으로 5 artifact type emit:
- thin-wrapper 진입 커맨드 `.claude/commands/harness/devkit.md` (<20 LOC, Runner 참조).
- Runner Workflow `.claude/workflows/harness-devkit-run.js` (manifest 읽기, primitive별 dispatch, determinism: Date.now()/Math.random() 호출 금지).
- specialist 서브에이전트 `.claude/agents/harness/harness-devkit-{release-update,github,release}-specialist.md` (REQ-DHC-002: 기존 body를 구조적 충실도로 포팅).
- 동반 skill `.claude/skills/harness-devkit-*/SKILL.md` (필요 시).
- `manifest.json` (SSOT, 8 필드).

### M4 — ACTIVATE (Priority Medium)

orchestrator-direct dry-run + `/goal` 자율 수렴. A/B는 load-bearing minimum 원칙으로 단순 task 범위면 skip(rationale 기록). PASS → `/harness:devkit` 노출.

### M5 — 번호 커맨드 삭제 + 구 소스 처분 (Priority High)

REQ-DHC-003:
- `.claude/commands/97-release-update.md`, `98-github.md`, `99-release.md` 삭제.
- **구 소스 처분 결정: 삭제 권장**. `.claude/agents/local/release-update-specialist.md`, `.claude/agents/local/github-specialist.md`, `.claude/skills/moai/workflows/release.md`는 devkit specialist로 포팅 완료 후 삭제하여 dual-source drift를 방지한다. (보존 시 두 소스가 갈라질 위험 — Single source of truth 원칙.)
- 정당화: 포팅이 구조적 충실도를 유지하므로 구 body는 중복이 된다. dual-source는 향후 한쪽만 수정되어 silent drift를 일으킨다.

### M6 — CI guard 확장 (embedded-tree-absence 단언, Priority High)

REQ-DHC-005 [D1 re-anchored]:
- **`internal/template/embedded_namespace_test.go`를 모델로 한 신규 embedded-tree-absence 단언**을 추가한다 (`dev_only_skill_test.go` 확장이 아님 — 그 테스트는 `.claude/skills/`만 walk하는 skills-only walker라 commands/agents/workflows artifact를 보호 못 함).
- 신규 단언은 `EmbeddedTemplates()`(또는 `internal/template/templates/`)에 다음이 모두 부재함을 단언한다: (a) `.claude/commands/harness/` 경로 부재, (b) `harness-devkit*` agent 파일 부재, (c) `.claude/workflows/harness-devkit-*` 파일 부재. `fs.ReadDir`/`fs.WalkDir`(Go-native) 사용, 외부 grep/shell 없음.
- RED/GREEN: RED는 walker가 검출 가능한 누출 형태(embedded 트리 하위에 심은 `harness-devkit` 경로 — `make build`가 컴파일해 넣는 형태)를 주입해 FAIL을 확인하고, 제거 후 PASS(GREEN). (`.claude/skills/` 하위 dir 주입이 아님 — skills walker가 아니므로.)
- `TestTemplateAgentsStructure`의 `{moai}`-only allowlist가 이미 `agents/harness/` 부재를 보호하므로, 신규 단언은 commands/workflows 차원 + `harness-devkit` 이름 차원을 보완한다. 97/98/99 grep 패턴은 doc-level 수동 체크리스트였을 뿐 CI 자동 gate가 아니었으므로, 삭제 후 신규 namespace의 CI 보호는 이 embedded-tree-absence 단언이 제공한다.

## §G. Anti-Patterns (피해야 할 것)

- Runner가 사용자에게 프롬프트할 수 있는 척하기 (B.1 위반).
- Layer B specialist(`cli-template`/`hook-ci`/`quality`/`workflow`) 이름 재사용/덮어쓰기.
- 9-phase 워크플로우 로직을 처음부터 재작성(포팅이어야 함 — REQ-DHC-002).
- 구 소스를 삭제하지 않고 dual-source 유지(M5 권장과 반대 — drift 위험).
- 사용자 템플릿(`internal/template/templates/`)에 devkit artifact 누출(REQ-DHC-004).
- 4번째 capability 추가(Tier S Enforce Simplicity 위반).

## §H. Cross-References

- v4 Builder 진입: `.claude/skills/moai/workflows/harness-build-entry.md`
- v4 Builder phase 로직: `.claude/skills/moai/workflows/harness-builder.md`
- namespace doctrine: `.moai/docs/harness-namespace-doctrine.md` (CLAUDE.local.md §24)
- dev-only 격리 doctrine: `.moai/docs/dev-only-commands-isolation.md` (CLAUDE.local.md §21 stub)
- dynamic-workflow 비대칭 boundary: `.claude/rules/moai/core/agent-common-protocol.md` §User Interaction Boundary
- embedded-tree-absence 보호 테스트(M6 모델): `internal/template/embedded_namespace_test.go` + `commands_audit_test.go` + `namespace_protection_audit_test.go`
- (참고만) skills-only walker: `internal/template/dev_only_skill_test.go` — devkit artifact는 commands/agents/workflows이므로 이 테스트의 보호 범위 밖(D1)

## §F-sync. Sync-phase task 노트 (manager-docs 수행)

REQ-DHC-006 — 3개 doctrine surface 업데이트:
- (1) `.moai/docs/dev-only-commands-isolation.md` 업데이트 (배포 금지 표 + 체크리스트를 harness-namespaced 도구로 교체/확장, 이 SPEC credit 마이그레이션 노트, LOCAL-NAMESPACE-CONSOLIDATION 노트와 정합).
- (2) CLAUDE.local.md §21 stub pointer 텍스트 조정(필요 시).
- (3) **[D6] CLAUDE.local.md §2 "Local-Only Files (Never in Templates)" 업데이트** — §2는 `97-release-update.md`/`98-*.md`/`.claude/skills/moai/workflows/release-update.md`를 독립적으로 명명. M5 삭제 후 stale → 삭제된 항목 제거 + 신규 `commands/harness/devkit*` + `workflows/harness-devkit-run.js` 추가. **false positive 주의**: `git-workflow-doctrine.md`의 `release.md`(= GitHub-Actions release.yml/Release Drafter)는 건드리지 않는다.

기타 sync 작업:
- memory 업데이트 (manager-docs가 수행 — manager-spec은 직접 작성 안 함).
- CHANGELOG/frontmatter status transition.

# Harness Namespace Separation Doctrine — extracted from CLAUDE.local.md §24

> Maintainer-local doctrine extracted from CLAUDE.local.md to cut session-launch context (CLAUDE.local.md loads in full at every launch). The matching CLAUDE.local.md section now carries a short stub pointing here. This file is NOT loaded at launch — read it when the topic applies. Subsection numbering is preserved so existing cross-references still resolve.

## 24. Harness Namespace 분리 정책

[HARD] Skills + Agents의 namespace는 **"범용 배포"** vs **"사용자 생성"** 으로 명확히 분리한다.

### §24.1 Skills Namespace

| Prefix | 범위 | Source of Truth | `moai update` 영향 |
|--------|------|-----------------|---------------------|
| `moai-foundation-*` / `moai-workflow-*` / `moai-domain-*` / `moai-ref-*` / `moai-meta-*` | 범용 배포 — moai-adk 패키지에 포함, 모든 사용자 프로젝트에 deploy | template | sync (overwrite local) |
| `moai-harness-*` | **하네스 빌더 (builder/lifecycle)** — moai-adk 패키지가 제공하는 generator/learner. 현재 `moai-meta-harness` + `moai-harness-learner`만 해당 | template | sync |
| **`hns-*`** | **사용자 생성 (canonical)** — v4 harness Builder가 사용자 프로젝트 도메인에 맞춰 generate. Builder는 `hns-` prefix만 emit | user project | **NOT synced (보호)** |
| **`harness-*`** / **`my-harness-*`** | **사용자 생성 (legacy 세대 — recognition-based backward compat)** — 기존 프로젝트의 구세대 artifact. 신규 generate 금지, 기존 자산은 계속 인식·보존 | user project | **NOT synced (보호)** |

[HARD] 사용자 프로젝트별 도메인 specialist skill은 **`hns-*` prefix만** 신규 emit. `harness-*` / `my-harness-*` 는 legacy 세대로 인식·보존만 하며(tri-generation recognition), `moai-harness-*` 또는 다른 `moai-*` prefix로 emit하면 contract 위반.

### §24.2 Agents Directory

| Path | 범위 | Source of Truth | `moai update` 영향 |
|------|------|-----------------|---------------------|
| `.claude/agents/moai/` | manager-spec, manager-develop, manager-docs, manager-git, plan-auditor, sync-auditor, builder-harness (retained 7, FLAT layout per v.2.x baseline) | template | sync |
| **`.claude/agents/harness/`** | **사용자 생성 domain specialist agents** — `moai-meta-harness`가 `/moai project` Phase 5+ 인터뷰 후 generate | user project | **NOT synced (보호)** |

[HARD] `internal/template/templates/.claude/agents/harness/` 디렉토리는 **존재 자체가 금지**. template에는 `{core,expert,meta}/` 만 mirror. `harness/` directory 등장 시 cleanup chore + 본 §24 cross-reference.

### §24.3 운영 원칙

- [HARD] `moai-harness-*` prefix로 사용자 프로젝트별 skill generate 금지 — v4 Builder는 `hns-*` prefix만 emit
- [HARD] template (`internal/template/templates/`)에 `hns-*` / `harness-*` skill 또는 `.claude/agents/harness/*-specialist.md` 누출 금지 (CI guard가 dual name set 검사)
- [HARD] `moai update`의 namespace 보호 contract: `hns-*` skill(canonical) + `harness-*` / `my-harness-*` skill(legacy) + `.claude/agents/harness/` 디렉토리는 sync 대상 제외 (user-owned, tri-generation recognition)
- [HARD] `moai-meta-harness` skill 본체는 `moai-*` namespace (generator/builder이므로 범용 배포 대상)
- 선례: chore commit `4f1135684` (2026-05-23) — moai-adk-go 도메인 specialist 4 agent + `moai-harness-cli-template` / `moai-harness-patterns` 2 skill 잘못된 누출을 제거하면서 본 정책 명문화. 정정 전 SPEC-V3R6-HARNESS-RENAME-001 (PR #1043, 2026-05-22)의 my-harness → moai-harness 통합은 본 namespace 분리 정책 도입으로 부분 supersede됨
- 후속: 2026-05-26 prefix doctrine을 `my-harness-*` → `harness-*` 으로 migration (이번 Phase 1 doctrine-only 변경). Go code enforcement (`internal/cli/update.go`, `internal/cli/update_archive.go`, `internal/cli/update_preserve_inventory.go`, `internal/harness/prefix_conflict.go`, test fixtures)는 여전히 `my-harness-*` enforce 상태이며 별개 SPEC (가칭 `SPEC-V3R6-HARNESS-NAMESPACE-V2-001`, Tier M, 39 Go files + 30+ tests scope)에서 catch-up 예정. **그 때까지 `harness-*` prefix actual generation 금지** — 새 prefix는 protection 안 받으므로 `moai update` 시 삭제 위험.

### §24.4 `moai update` 동작 Contract

[HARD] `moai update`는 `.claude/skills/` + `.claude/agents/` + `.claude/commands/` + `.claude/workflows/` 에 대해 다음 동작을 수행한다:

| Namespace / Path | 동작 | 백업 정책 |
|------------------|------|-----------|
| `.claude/skills/moai-*` (incl. `moai-harness-*`, `moai-meta-*`, `moai-foundation-*`, `moai-workflow-*`, `moai-domain-*`, `moai-ref-*`) | **삭제 후 신규 설치** (overwrite) | 백업 불필요 — template-managed, 사용자 수정 시 손실됨 |
| **`.claude/skills/hns-*`** | **절대 삭제 금지 + 절대 modify 금지** (SPEC-HNS-PREFIX-RENAME-001 M2 — canonical) | **백업 + 보존** (user-owned, Go enforcement 작동) |
| **`.claude/skills/harness-*`** / **`.claude/skills/my-harness-*`** | **절대 삭제 금지 + 절대 modify 금지** (legacy 세대) | **백업 + 보존** (user-owned, tri-generation recognition) |
| `.claude/agents/moai/` | 삭제 후 신규 설치 (overwrite) | 백업 불필요 — template-managed (FLAT layout per v.2.x baseline) |
| **`.claude/agents/harness/`** | **절대 삭제 금지 + 절대 modify 금지** (디렉토리 단위 보호 — 파일 prefix 세대 무관) | **백업 + 보존** (user-owned) |
| **`.claude/commands/harness/`** | **절대 삭제 금지 + 절대 modify 금지** (SPEC-V3R6-HARNESS-V4-001 M1) | **백업 + 보존** (user-owned — `/harness:<name>` command files) |
| **`.claude/workflows/hns-*.js`** | **절대 삭제 금지 + 절대 modify 금지** (SPEC-HNS-PREFIX-RENAME-001 M2 — canonical Runner) | **백업 + 보존** (user-owned — harness Runner Workflow scripts) |
| **`.claude/workflows/harness-*.js`** | **절대 삭제 금지 + 절대 modify 금지** (SPEC-V3R6-HARNESS-V4-001 M1 — legacy Runner) | **백업 + 보존** (user-owned — harness Runner Workflow scripts) |
| 기타 사용자 직접 추가 자산 (`.claude/agents/<custom>.md`, `.claude/skills/<custom>/` 단 prefix가 `moai-` 시작 아닌 것) | 보존 | 백업 + 보존 |
| `.moai/harness/` (main.md, interview-results.md, extensions) | 절대 삭제 금지 | 백업 + 보존 (user-owned) |

[HARD] `moai update` 실행 전 user-owned 자산은 **반드시 백업** — 갑작스러운 process kill 등 비정상 종료 시에도 손실 위험 0이어야 한다. 백업 위치: `.moai/backups/update-{ISO-DATE}/` 권장.

[HARD] 이 contract는 다음 SSOT와 일관성을 유지:
- `.claude/skills/moai-meta-harness/SKILL.md` § Namespace Separation 의 Storage Roots 표
- `.claude/rules/moai/development/skill-authoring.md` § Skills Namespace Policy
- `.claude/rules/moai/development/agent-authoring.md` § Agent Directory Convention

[HARD] Go 구현 (`internal/cli/update.go`, `internal/cli/update_archive.go`)이 본 contract를 정확히 준수하는지는 SPEC-V3R6-UPDATE-NAMESPACE-PROTECT-001 (별도 작성 예정)에서 검증한다. 현재 본 contract는 정책 명문화이며, 코드 구현 검증은 후속 작업.

### §24.5 Phase 2 Drift — RESOLVED (2026-06-18, SPEC-V3R6-HARNESS-NAMESPACE-V2-001)

`harness-*` user-owned namespace의 Go enforcement + test fixture + `moai-meta-harness` generator emission + CI sentinel + `my-harness-*` legacy backward-compat(dual-recognition deprecation window)가 모두 atomic하게 `harness-*`로 전환 완료되었다. 이전 Phase 2 drift entry-condition 노트는 해소되어 간략화됨. `harness-*` vs `moai-harness-*` substring 분리는 `strings.HasPrefix` exact match로 정확히 유지된다.

### §24.6 hns- Prefix Rename — RESOLVED (2026-07-13, SPEC-HNS-PREFIX-RENAME-001)

사용자 생성 artifact prefix가 `harness-<name>` → **lowercase `hns-<name>`** 으로 rename 완료. 배경: `harness-` prefix가 6개 비-artifact "harness" namespace(`moai-harness-*` 템플릿 skill, `harness-spec.yaml`, `.moai/harness/` state, `harness.yaml` config, `/harness:` command namespace, `moai harness` CLI verb)와 시각적으로 충돌하여 audit/sweep 분류 오류가 반복됨. `hns-`는 case-sensitive 기준 production tree 내 충돌 0인 grep-unambiguous prefix.

- **Tri-generation recognition**: Go 인식 로직(update.go / frozen_guard.go / prefix_conflict.go / v4lifecycle.go / doctor.go / doctor_skills.go / doctor_harness.go)은 `hns-*`(canonical) + `harness-*` + `my-harness-*`(legacy 세대)를 모두 user-owned로 인식·보존한다. 기존 사용자 프로젝트의 legacy artifact는 절대 삭제되지 않는다 (recognition-based backward compat — migration tooling 없음, 의도된 선택).
- **Builder emission**: v4 Builder GENERATE contract는 `hns-<name>-*` 6종 artifact만 emit (Runner `hns-<name>-run.js`, specialist `hns-<name>-*-specialist.md`, skill `hns-<name>-*/`, verify skill `hns-<name>-verify`, manifest `runner_workflow`, thin-command reference).
- **디렉토리 불변**: `.claude/agents/harness/`, `.claude/commands/harness/`, `/harness:` namespace는 이름 유지 — 파일 prefix만 변경.
- **CI guard dual name set**: `split_namespace_test.go`(sentinel `SPLIT_HARNESS_NAMESPACE_LEAK`)는 `harness-{release-update,github,release}*` + `hns-{release-update,github,release}*` 양쪽 부재를 검사; `namespace_protection_audit_test.go`는 `hns-` leak도 거부.
- **Legacy 세대 sunset**: `harness-*` / `my-harness-*` 인식 창 종료는 별도 후속 chore SPEC 소관.


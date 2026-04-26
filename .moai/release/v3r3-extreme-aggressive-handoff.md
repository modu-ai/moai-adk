# V3R3 Extreme Aggressive — Handoff Document for Next Session(s)

**작성일**: 2026-04-26
**현재 main HEAD**: 3ed706a5b
**현재 버전**: v2.13.2 (배포) → v2.15 ~ v2.17 진행 예정
**작성 기준**: 사용자 6 결정 + claude-code-guide 공식 조사 + revfactory/harness 분석

---

## 0. 핵심 설계 결정 (final)

### 0.1 Namespace 분리 (사용자 명시)

```
STATIC (기존 그대로, MoAI maintainer가 추적):
├── .claude/agents/moai/
│   ├── manager-*.md         (8개)
│   ├── expert-*.md          (8개)
│   ├── builder-*.md         (3개)
│   └── evaluator-*.md       (2개)
└── .claude/skills/
    ├── moai-foundation-*    (4개)
    ├── moai-workflow-*      (8개)
    ├── moai-ref-*           (5개)
    ├── moai-design-*        (3개, pencil-integration 제거)
    ├── moai-domain-{copywriting,brand-design}  (2개 FROZEN)
    └── moai-meta-harness    (1개 NEW)
                             = 22 base skills + 1 meta = 23

DYNAMIC (per-project, harness 생성, 사용자 namespace):
├── .claude/agents/my-harness/         ← 신규 namespace
│   ├── ios-architect.md               (예: 모바일 프로젝트)
│   ├── swiftui-engineer.md
│   └── healthkit-integrator.md
└── .claude/skills/
    ├── my-harness-ios-patterns/        ← 신규 prefix
    ├── my-harness-swift-testing/
    └── my-harness-healthkit-permissions/
```

**핵심 원칙**:
- Static `moai-*` prefix vs Dynamic `my-harness-*` prefix → 충돌 방지
- 사용자가 직접 customization 했음을 시각적으로 인지
- Git diff에서 명확히 구분 (moai update가 my-harness/* 건드리지 않음)

### 0.2 통합 장치 설계 (Static workflow ↔ Dynamic harness) — **5-Layer (사용자 통찰 반영)**

**핵심 원칙**: moai-managed workflow files는 절대 수정 안 함. template에 정적 import line만 포함되어 배포. user customization은 `.moai/harness/` user area에만 존재.

| 영역 | 위치 | moai update | 책임 |
|---|---|---|---|
| **MOAI-managed** | .claude/skills/moai/workflows/, .claude/agents/moai/, .claude/rules/moai/ | 갱신 | maintainer |
| **USER-managed** | .moai/harness/, .claude/agents/my-harness/, .claude/skills/my-harness-*/, CLAUDE.md, workflow.yaml | 보존 | user + harness |

기존 `/moai plan`, `/moai run`, `/moai sync` 등이 dynamic harness를 자동 발견 + 활용하도록 5-layer 장치:

#### Layer 1: Skill 자동 트리거 (frontmatter)

harness 생성 skill의 frontmatter에 `triggers.agents` 자동 inject:

```yaml
---
name: my-harness-ios-patterns
description: "iOS Swift native development patterns auto-generated for THIS project."
metadata:
  generated_by: "moai-meta-harness"
  generated_at: "2026-04-26T14:00:00Z"
  parent_spec: "SPEC-PROJ-INIT-001"

triggers:
  keywords: ["ios", "swift", "swiftui", "xcode"]
  agents: ["manager-tdd", "manager-ddd", "manager-spec", "expert-frontend"]
  phases: ["plan", "run", "sync"]
  paths: "**/*.swift,**/Package.swift,**/*.xcodeproj"
---
```

→ manager-tdd가 .swift 파일 작업 시 my-harness-ios-patterns가 **자동 활성**

#### Layer 2: workflow.yaml 명시 등록

`/moai project` 실행 시 harness가 `workflow.yaml`에 customization 등록:

```yaml
# .moai/config/sections/workflow.yaml (harness가 갱신)
workflow:
  team:
    enabled: true
    role_profiles: [...]

  # NEW — harness customization registry
  harness:
    enabled: true
    generated_at: "2026-04-26T14:00:00Z"
    domain: "ios-mobile"
    spec_id: "SPEC-PROJ-INIT-001"
    custom_agents:
      - name: "ios-architect"
        path: ".claude/agents/my-harness/ios-architect.md"
        invoke_in: ["plan", "run"]
      - name: "swiftui-engineer"
        path: ".claude/agents/my-harness/swiftui-engineer.md"
        invoke_in: ["run"]
    custom_skills:
      - name: "my-harness-ios-patterns"
        path: ".claude/skills/my-harness-ios-patterns/SKILL.md"
        triggers_in: ["plan", "run", "sync"]
    chaining_rules:
      - phase: "run"
        before_specialist: "ios-architect"   # manager-tdd가 expert-* 호출 전 ios-architect 먼저
        after_specialist: "swiftui-engineer"
```

→ 기존 manager-tdd가 Phase 2 시작 시 workflow.yaml.harness.chaining_rules를 read하여 custom agent를 명시적 chain에 삽입

#### Layer 3: CLAUDE.md @import marker

`/moai project` 완료 시 CLAUDE.md에 harness 섹션 자동 inject:

```markdown
# CLAUDE.md (마지막 섹션 추가)

## Project-Specific Configuration (Harness-Generated)
<!-- moai:harness-start id="SPEC-PROJ-INIT-001" generated="2026-04-26" -->
**Domain**: iOS Mobile (Swift + SwiftUI)
**Harness level**: thorough
**Updated**: 2026-04-26

See @.moai/config/sections/workflow.yaml for team roles + harness chaining
See @.claude/rules/project/ios-swift-patterns.md for Swift conventions
See @.claude/agents/my-harness/ios-architect.md
See @.claude/agents/my-harness/swiftui-engineer.md
See @.claude/skills/my-harness-ios-patterns/SKILL.md
<!-- moai:harness-end -->
```

→ 새 세션 시작 시 CLAUDE.md 로드 → @import 자동 follow → harness customization이 context에 포함

#### Layer 4 (재설계): Workflow Static Import Line ★

**moai-managed workflow files는 수정하지 않음**. Template에 처음부터 한 줄 정적 import 포함.

```markdown
# template/.../workflows/plan.md (MoAI maintainer가 한번 작성, 모든 사용자 동일)
...
## Phase N: Final
...

---

## Custom Harness Extension (Optional)

@.moai/harness/plan-extension.md

*(이 파일은 `/moai project --harness`로 생성됩니다. 파일이 없으면 자동으로 skip됩니다.)*
```

동일 패턴: run.md, sync.md, design.md (각 workflow 1줄씩 추가).

→ moai update 시 import line 그대로 유지. 사용자 `.moai/harness/*` 보존.
→ 파일 없으면 Claude Code @import 자동 skip (graceful)

#### Layer 5 (NEW): `.moai/harness/` Directory Convention ★

사용자 customization 전용 디렉터리. moai update가 절대 건드리지 않음.

```
.moai/harness/                            ← user area
├── main.md                               ← CLAUDE.md @import 진입점
├── plan-extension.md                     ← plan.md 워크플로우 확장
├── run-extension.md                      ← run.md 워크플로우 확장
├── sync-extension.md                     ← sync.md 워크플로우 확장
├── design-extension.md (optional)        ← design.md 확장
├── chaining-rules.yaml                   ← agent chaining 규칙
├── interview-results.md                  ← 소크라테스 인터뷰 결과 영구 기록
└── README.md                             ← 사용자 가시성
```

**파일별 역할** (예: iOS 프로젝트):
- `main.md` — 프로젝트 전체 customization entry (CLAUDE.md에서 @import)
- `plan-extension.md` — `/moai plan` 시 manager-spec이 ios-architect로 사전검증 등 명시
- `run-extension.md` — `/moai run` 시 expert-* chaining 순서 (ios-architect → expert-frontend → swiftui-engineer → xcode-tester)
- `chaining-rules.yaml` — agent chain 정의 (machine-readable)
- `interview-results.md` — Q1-Q16 답변 영구 기록 (재인터뷰 시 참조)

### 0.3 활성화 메커니즘 표

| Layer | 활성화 트리거 | moai update 영향 | 사용자 경험 |
|---|---|---|---|
| **L1** Skill triggers | paths/keywords 매칭 | ✅ 보존 (사용자 영역) | 인지 불필요 (자동) |
| **L2** workflow.yaml | harness 섹션 등록 | ✅ 보존 (사용자 영역) | manager-tdd가 chain 적용 |
| **L3** CLAUDE.md @import | `<!-- moai:harness-* -->` markers | ✅ 보존 (사용자 영역) | 가시적 |
| **L4** Workflow import line | template static import | ✅ 갱신 (한 줄 그대로) | 자동 |
| **L5** `.moai/harness/` | user directory | ✅ 보존 (사용자 영역) | 사용자가 직접 편집 가능 |

**핵심**: L4의 import line은 template에서 갱신되지만 한 줄이므로 conflict 없음. L5의 실제 customization은 사용자 영역.

**5-layer 모두 적용** → robust + safe + visible + extensible + graceful fallback

---

## 1. 다음 세션 진행 계획 (Option A — Phased)

### Phase A — v2.15.0 마무리 (현재 진행)

**목표**: 작성된 SPECs 구현 + 사용자 6 결정 적용 + v2.15 배포

#### Step 1: SPEC-V3R3-ARCH-007 구현 (Token Circuit Breaker)
- `.moai/config/sections/runtime.yaml` 신설
- `internal/runtime/budget.go` (NEW Go runtime)
- `internal/template/templates/.moai/config/sections/runtime.yaml` 미러
- SessionStart hook에서 runtime.yaml 로드
- 75/90% threshold + per-agent budget + circuit breaker

#### Step 2: SPEC-V3R3-COV-001 구현 (Mobile Seed)
- `.claude/agents/moai/expert-mobile.md` (NEW)
- `.claude/skills/moai-domain-mobile/SKILL.md` (NEW, harness seed)
- 단, 4 mobile strategy skills은 **harness 동적 생성용 reference**로만 (정적 skill로 만들지 않음)

#### Step 3: SPEC-V3R3-CMD-CLEANUP-001 신규 작성 + 구현
- `.claude/commands/moai/gate.md` 신규
- `.claude/skills/moai/workflows/security.md` 보존 (review/sync 호출용)
- **review.md, sync.md 보안 검수 phase를 security.md 수준으로 강화**:
  - review.md Phase 4: dependency vuln scan + secrets git history + data isolation 추가
  - sync.md Phase 0.55: 변경 파일 + dependency files 모두 audit
- `.claude/skills/moai/workflows/context.md` **삭제**
- `.claude/commands/moai/` 의 context 관련 routing 제거
- SKILL.md routing 표 업데이트 (context 제거, gate 추가, security routing 변경)

#### Step 4: v2.15 release prep + PR + tag
- CHANGELOG.md
- system.yaml v2.13.2 → v2.15.0
- README.md
- RELEASE-NOTES-v2.15.0.md (`.moai/release/v2.15.0-draft.md` 기반 확정)
- feature branch + PR + tag v2.15.0

### Phase B — v2.16.0 (PATTERNS-001)

**목표**: revfactory/harness 6 reference docs 흡수 + Pattern Cookbook

#### Step 1: SPEC-V3R3-PATTERNS-001 신규 작성

산출 파일:
- `.claude/rules/moai/development/agent-patterns.md` (6 architectural patterns)
- `.claude/rules/moai/quality/boundary-verification.md` (QA 경계면 교차 검증, 7 실제 버그 사례)
- `.claude/rules/moai/development/skill-ab-testing.md` (With-skill vs Baseline 방법론)
- `.claude/rules/moai/workflow/team-pattern-cookbook.md` (5 팀 예시)
- `.claude/rules/moai/development/orchestrator-templates.md` (Team / Sub / Hybrid 3 templates)
- `.claude/rules/moai/development/skill-writing-craft.md` (description, body, schema)

License credit: Apache 2.0 (revfactory/harness)

#### Step 2: 검증
- 6개 rule 파일 frontmatter (paths/triggers)
- agent들이 자동 reference 가능한지 (paths matching)
- existing manager-quality, expert-testing이 자동 참조하는지

#### Step 3: v2.16 release
- CHANGELOG.md
- system.yaml v2.16.0
- PR + tag

### Phase C — v2.17.0 (Extreme Aggressive 핵심)

**목표**: meta-harness + Vibe Design + Project Harness Activation

#### Step 1: SPEC-V3R3-HARNESS-001 신규 작성 + 구현

산출:
- `.claude/skills/moai-meta-harness/SKILL.md` (revfactory/harness 흡수 + MoAI 통합)
- 6 references (agent-patterns 등은 v2.16에서 이미 .claude/rules/로 이동했으므로 reference 통합)
- **16 정적 skills 제거** (BC-V3R3-007):
  - domain-backend, domain-frontend, domain-database, domain-db-docs, domain-mobile (5)
  - framework-electron (1)
  - library-shadcn, library-mermaid, library-nextra (3)
  - tool-ast-grep (1)
  - platform-auth, platform-deployment, platform-chrome-extension (3)
  - workflow-research, workflow-pencil-integration, formats-data (3)
- moai update 마이그레이터: 16 skills 자동 archive + meta-harness 자동 설치

⚠️ **주의**: 16 정적 skills 제거는 BREAKING. 사용자에게 명시 + grace period 권장.

#### Step 2: SPEC-V3R3-DESIGN-PIPELINE-001 신규 작성 + 구현

산출:
- `.claude/skills/moai-workflow-design-import/SKILL.md` 확장 (Path A + B1 + B2 통합)
- `.claude/skills/moai-design-system/SKILL.md` 확장 (DTCG 2025.10 token spec)
- W3C DTCG token validator (Go util)
- `/moai design` 워크플로우 분기:
  - Path A: Claude Design handoff bundle
  - Path B1: Figma → meta-harness 동적 생성 (figma-extractor skill)
  - Path B2: Pencil → meta-harness 동적 생성 (pencil-mcp skill)

`moai-workflow-pencil-integration` 제거 (HARNESS-001 16개에 포함).

#### Step 3: SPEC-V3R3-PROJECT-HARNESS-001 신규 작성 + 구현

산출:
- `.claude/skills/moai/workflows/project.md` 확장 (Phase 5+: 소크라테스 인터뷰 + harness 분기)
- AskUserQuestion 인터뷰 6 질문 (도메인/팀크기/방법론/디자인툴/배포/customization)
- meta-harness 호출 → `.claude/agents/my-harness/`, `.claude/skills/my-harness-*/` 생성
- **통합 장치 4-layer 구현**:
  - Layer 1: harness 생성 skills의 frontmatter triggers 자동 inject
  - Layer 2: workflow.yaml.harness 섹션 갱신
  - Layer 3: CLAUDE.md `<!-- moai:harness-start -->` ~ `<!-- moai:harness-end -->` marker inject
  - Layer 4: workflows/{plan,run,sync}.md에 Phase 0.95 hook 추가

검증:
- 새 세션 시뮬레이션 (CLAUDE.md 로드 → harness 인식 → 자동 활성)
- moai doctor (구성 일관성)
- diff -rq template/local

#### Step 4: v2.17 release prep + PR + tag

- BREAKING change 안내 (BC-V3R3-007: 16 skills 제거)
- migration guide (`moai update` 자동 + manual fallback)
- CHANGELOG.md (대규모 entry)
- RELEASE-NOTES-v2.17.0.md
- PR + tag

---

## 2. 다음 세션 Resume Message (paste-ready) — LEARNING-001 SPEC 작성

> **STATUS UPDATE (2026-04-26 세션 종료)**: Phase B Step 1+2 완료.
> - PATTERNS-001 구현 완료, PR #715 (release/v2.16.0 branch, 5 commits) 생성됨
> - 보너스: PostToolUse async + freeze diagnosis + V3R2 inherited 11 lint cleanup
> - 다음 세션 목표: PR admin merge 확인 → tag v2.16.0 → LEARNING-001 SPEC 작성

```
ultrathink. V3R3 Phase B Step 3 진행. SPEC-V3R3-HARNESS-LEARNING-001 작성.

진행 위치:
- main HEAD: 같음 (b47101779) — 새 SPEC 작성은 main에서 새 feat/* 브랜치
- release/v2.16.0 branch에 5 commits, PR #715 OPEN, 사용자 검토/머지 대기
- 완료: PATTERNS-001 4파일 SPEC + 6 rule files + NOTICE + Template-First sync
- 완료: PostToolUse 60s→10s + async:true 적용 (template + local)
- 완료: freeze 진단 가이드 (settings-management.md 4-step checklist)
- 완료: V3R2 restore inherited 11 lint cleanup (audit_cache.go + audit_report.go)
- 미작성 (이 세션 목표): SPEC-V3R3-HARNESS-LEARNING-001
- 미작성 (다음): SPEC-V3R3-HARNESS-001, DESIGN-PIPELINE-001, PROJECT-HARNESS-001

handoff document: .moai/release/v3r3-extreme-aggressive-handoff.md
                  §4.3 LEARNING-001 요구사항 (4-tier 학습 + 5-layer safety + /moai harness)

다음 단계 (이 세션):
1. PR #715 status 확인: gh pr view 715
   • CI all green이면 사용자 admin merge 대기 (squash 금지, merge commit 사용 — release branch)
   • merge 후: git tag v2.16.0 && git push origin v2.16.0 → GoReleaser 자동 release
2. SPEC-V3R3-HARNESS-LEARNING-001 작성 (manager-spec dispatch)
   • 산출: .moai/specs/SPEC-V3R3-HARNESS-LEARNING-001/{spec,plan,acceptance,tasks}.md
   • 핵심 요구사항 (handoff §4.3):
     - 4-tier 학습 (1x observation / 3x heuristic / 5x rule / 10x auto-update)
     - 5-layer safety (Frozen Guard / Canary / Contradiction / Rate Limit / Human Oversight)
     - .claude/agents/my-harness/ + .claude/skills/my-harness-*/ 자동 진화
     - .moai/harness/ 사용자 영역만 자동 업데이트 (moai-managed 절대 불가침)
     - .moai/config/sections/harness.yaml learning section + /moai harness CLI subcommand
   • 의존성: HARNESS-001 (meta-harness), PROJECT-HARNESS-001 (소크라테스 인터뷰)
   • design.yaml safety architecture (Section 5) 활용
3. (시간 여유 시) HARNESS-001 + DESIGN-PIPELINE-001 + PROJECT-HARNESS-001 SPEC 작성

applied lessons:
- feedback_large_spec_wave_split.md (각 dispatch ~1.5KB)
- context-window-management.md (75% 임계 모니터)

긴 세션 예상 — 75% 도달 시 progress.md 저장 + /clear + 새 세션 resume.
완료 후: v2.17 phase 4-SPEC 구현 (HARNESS → DESIGN → PROJECT → LEARNING 순)
```

---

## 3. 다음 세션 단계별 Dispatch 지침

### 3.1 CMD-CLEANUP-001 SPEC 작성 (manager-spec)

```
SPEC-V3R3-CMD-CLEANUP-001 작성. 4파일 (spec.md/plan.md/acceptance.md/tasks.md).

Scope:
1. .claude/commands/moai/gate.md 신규 (thin wrapper, Skill("moai") 패턴)
2. .claude/skills/moai/workflows/security.md 보존 (review/sync 호출용 유지)
3. review.md Phase 강화: dependency vuln scan + secrets git history + data isolation 추가
4. sync.md Phase 0.55 강화: 변경 파일 + dependency files 모두 audit
5. .claude/skills/moai/workflows/context.md 삭제
6. SKILL.md routing 표 갱신: context 제거, gate 추가, security routing 통합

frontmatter:
- id: SPEC-V3R3-CMD-CLEANUP-001
- title: Commands Cleanup — gate 추가, security 흡수, context 제거
- priority: P0
- breaking: false
- bc_id: []
- lifecycle: spec-anchored

EARS REQ ~6개, AC ~5개, AC-REQ traceability 100%.
```

### 3.2 ARCH-007 구현 (manager-tdd, Go runtime + yaml)

```
SPEC-V3R3-ARCH-007 구현 (TDD). main 브랜치 기준.

Read SPEC artifacts (.moai/specs/SPEC-V3R3-ARCH-007/).

Scope:
1. .moai/config/sections/runtime.yaml 신설 (75/90% threshold + per-agent budget)
2. internal/template/templates/.moai/config/sections/runtime.yaml 미러
3. internal/runtime/budget.go (NEW): per-agent token tracker, stall detection
4. internal/runtime/budget_test.go (TDD RED → GREEN)
5. SessionStart hook에서 runtime.yaml 로드 (internal/hook/session_start.go)
6. 75% 도달 시 progress.md 자동 저장 + resume message 자동 생성

HARD rules: Template-First, t.TempDir() for tests, conventional commit (한국어 body), make build && go test.
```

### 3.3 COV-001 구현 (manager-tdd, mobile seed)

```
SPEC-V3R3-COV-001 구현 (TDD).

Scope (간소화 — harness 시대에 맞춰):
1. .claude/agents/moai/expert-mobile.md 신규 (4-strategy router only)
2. .claude/skills/moai-domain-mobile/SKILL.md 신규 (harness seed reference)
3. 4 mobile sub-skills (ios/android/RN/flutter)는 ❌ 만들지 않음 — harness 위임
4. moai-domain-mobile은 "harness가 어떤 mobile skill을 생성할지 가이드"

Template + local sync, conventional commit.
```

### 3.4 CMD-CLEANUP-001 구현 (manager-tdd)

```
SPEC-V3R3-CMD-CLEANUP-001 구현 (TDD).

Scope (4 commits 권장 분할):

Commit 1: gate command 추가
- .claude/commands/moai/gate.md 신규 (thin wrapper)
- internal/template/templates/.claude/commands/moai/gate.md 미러
- SKILL.md routing 표에 gate 추가 (already routed in spec)

Commit 2: review.md 보안 강화
- .claude/skills/moai/workflows/review.md
- Phase 4 (security perspective) 강화: dependency vuln scan + secrets git history + data isolation
- expert-security 호출 시 security.md 수준 깊이 명시

Commit 3: sync.md 보안 강화
- .claude/skills/moai/workflows/sync.md
- Phase 0.55 강화: 변경 파일 + dependency files (go.mod/package.json/...) audit

Commit 4: context 제거
- .claude/skills/moai/workflows/context.md 삭제
- SKILL.md routing 표에서 context 제거
- internal/template/commands_audit_test.go 갱신

Template + local sync, conventional commit, 한국어 body.
```

### 3.5 v2.15 Release Prep (manager-docs)

```
v2.15.0 release prep.

Scope:
1. CHANGELOG.md 갱신 — V3R3 Phase A entries (DEF-007/001/ARCH-003/ARCH-007/COV-001/CMD-CLEANUP-001)
2. .moai/config/sections/system.yaml: v2.13.2 → v2.15.0
3. internal/template/templates/.moai/config/config.yaml: 동일
4. README.md / README.ko.md: Version 줄 갱신
5. .moai/release/RELEASE-NOTES-v2.15.0.md 확정 (.moai/release/v2.15.0-draft.md 기반)
6. SPEC 상태 갱신:
   - SPEC-V3R3-DEF-007 → completed
   - SPEC-V3R3-DEF-001 → completed
   - SPEC-V3R3-ARCH-003 → completed
   - SPEC-V3R3-ARCH-007 → completed
   - SPEC-V3R3-COV-001 → completed
   - SPEC-V3R3-CMD-CLEANUP-001 → completed

Single commit: docs(release): v2.15.0 release prep — V3R3 Phase A 마무리.
```

### 3.6 PR + tag v2.15.0 (manager-git)

```
v2.15.0 release.

Scope:
1. feature branch: feature/v3r3-phase-a-v2.15
2. push -u origin feature/v3r3-phase-a-v2.15
3. gh pr create:
   - Base: main
   - Title: feat(v3R3): Phase A — Foundation Hardening + v2.15.0 release
   - Body: V3R3 Phase A 6 SPECs + BREAKING change 안내 + test plan
4. PR 머지 후 (수동 검토 후):
   - git tag v2.15.0
   - git push origin v2.15.0
   - gh release create v2.15.0 --notes-file .moai/release/RELEASE-NOTES-v2.15.0.md

DO NOT auto-merge — 사용자 검토 후 manual merge.
```

---

## 4. v2.16 / v2.17 진행 지침 (다다음 세션)

### 4.1 v2.16 PATTERNS-001

새 세션에서 manager-spec 호출:

```
SPEC-V3R3-PATTERNS-001 작성.

revfactory/harness Apache 2.0 기반 6 reference docs 흡수:
- .claude/rules/moai/development/agent-patterns.md (6 architectural patterns)
- .claude/rules/moai/quality/boundary-verification.md
- .claude/rules/moai/development/skill-ab-testing.md
- .claude/rules/moai/workflow/team-pattern-cookbook.md
- .claude/rules/moai/development/orchestrator-templates.md
- .claude/rules/moai/development/skill-writing-craft.md

소스: /tmp/harness-analysis/harness/skills/harness/references/* (이미 cloned)
또는 다시 clone: git clone https://github.com/revfactory/harness /tmp/harness-analysis/harness

License credit: Apache 2.0
attribution: each rule file 상단에 "# Source: revfactory/harness Apache 2.0"

이후 manager-tdd로 구현 + v2.16 release.
```

### 4.2 v2.17 HARNESS-001 + DESIGN-PIPELINE-001 + PROJECT-HARNESS-001

3개 SPEC 동시 작성 (manager-spec 1회 dispatch):

```
3 SPECs 동시 작성 (V3R3 Iteration 2-3 핵심):

1. SPEC-V3R3-HARNESS-001
   - moai-meta-harness skill 신설 (revfactory/harness 7-Phase workflow + MoAI 통합)
   - 16 정적 skills 제거 (BC-V3R3-007)
   - moai update 마이그레이터 확장
   - Static namespace: .claude/agents/moai/, .claude/skills/moai-*/
   - Dynamic namespace: .claude/agents/my-harness/, .claude/skills/my-harness-*/

2. SPEC-V3R3-DESIGN-PIPELINE-001
   - DTCG 2025.10 token spec 적용
   - Path A (Claude Design) + B1 (Figma) + B2 (Pencil) 통합
   - moai-workflow-design-import 확장 (3-path entry)
   - moai-design-system 확장 (DTCG validator)
   - moai-workflow-pencil-integration 제거 (HARNESS-001 16개에 포함)

3. SPEC-V3R3-PROJECT-HARNESS-001
   - /moai project Phase 5+: 소크라테스 인터뷰 (★ 16 질문, 4 라운드)
     Round 1: 도메인/기술/규모/팀 (Q1-Q4)
     Round 2: 방법론/디자인툴/UI복잡도/디자인시스템 (Q5-Q8)
     Round 3: 보안/성능/배포/외부통합 (Q9-Q12)
     Round 4: customization 범위/특수제약/우선순위/최종확인 (Q13-Q16)
   - meta-harness 호출 + ★ 5-layer 통합 장치 구현 (사용자 통찰 반영):
     Layer 1: harness skill frontmatter triggers 자동 inject (manager-* 자동 활성)
     Layer 2: workflow.yaml.harness 섹션 갱신 (chaining_rules) — 사용자 영역
     Layer 3: CLAUDE.md @import marker (<!-- moai:harness-start --> ~ <!-- moai:harness-end -->)
     Layer 4: ★ Workflow static import line — moai-managed workflow 수정 없음
              template/.../workflows/{plan,run,sync,design}.md에 한 줄씩 정적 추가:
                @.moai/harness/{phase}-extension.md
              moai update가 갱신해도 import line 그대로, 사용자 .moai/harness/* 보존
              파일 없으면 Claude Code 자동 skip (graceful)
     Layer 5: ★ .moai/harness/ user directory convention (NEW)
              main.md, plan-extension.md, run-extension.md, sync-extension.md,
              chaining-rules.yaml, interview-results.md, README.md
   - 새 세션 자동 활성화 검증
   - moai update 안전성 검증 (사용자 .moai/harness/* + .claude/agents/my-harness/* 보존)

3 SPECs 모두 .moai/specs/SPEC-V3R3-XXX/ 4파일 (spec/plan/acceptance/tasks).
이후 manager-tdd로 순차 구현 (HARNESS → DESIGN → PROJECT 순) + v2.17 release.
```

### 4.3 SPEC-V3R3-HARNESS-LEARNING-001 (사용자 추가 요구사항 — 2026-04-26)

**요구사항 출처**: 사용자 요청 ("사용자가 설치한 하네스는 사용자가 사용을 하면 할수록 학습된 내용을 지속적으로 자동으로 업데이트가 되도록")

**핵심**: 사용자 프로젝트에 설치된 dynamic harness (`.claude/agents/my-harness/`, `.claude/skills/my-harness-*/`, `.moai/harness/`)가 사용자 활동을 학습하여 자동으로 진화.

#### Scope

1. **사용자 활동 모니터링**:
   - `/moai` 명령 사용 빈도 (subcommand 별)
   - 자주 사용하는 SPEC ID 패턴
   - commit message 트렌드
   - agent 호출 통계 (어떤 agent가 어떤 작업에 반복 호출되는지)
   - 사용자 명시 feedback (`/moai feedback` 입력)

2. **학습 알고리즘 (3-tier)**:
   - **Tier 1 (Observation)**: 1x 발생 — 로그만 기록
   - **Tier 2 (Heuristic)**: 3x 발생 — harness skill body의 description 보강
   - **Tier 3 (Rule)**: 5x 발생 — harness skill의 triggers/keywords 자동 inject
   - **Tier 4 (Auto-update)**: 10x 발생 — 사용자 승인 후 chaining_rules.yaml 갱신

3. **학습 결과 자동 반영 대상**:
   - `.moai/harness/main.md` (project customization 진입점)
   - `.moai/harness/{plan,run,sync,design}-extension.md` (workflow 확장)
   - `.moai/harness/chaining-rules.yaml` (agent chain 규칙)
   - `.claude/skills/my-harness-*/SKILL.md` body
   - `.claude/agents/my-harness/*.md` body

4. **5-layer 안전 메커니즘 (design constitution §5 활용)**:
   - **L1 Frozen Guard**: moai-managed (`.claude/agents/moai/`, `.claude/skills/moai-*/`)는 절대 수정 금지
   - **L2 Canary Check**: 학습 변경 적용 전 shadow eval (이전 3 프로젝트에 영향 시뮬레이션)
   - **L3 Contradiction Detector**: 신규 학습이 기존 사용자 customization과 충돌 시 alert
   - **L4 Rate Limiter**: 주당 최대 3회 자동 업데이트, 24h cooldown
   - **L5 Human Oversight**: Tier 4 자동 업데이트 시 AskUserQuestion으로 승인 받기

#### 산출 (예정)

- `moai-harness-learner` skill (학습 + 자동 업데이트 코디네이터)
- `internal/harness/learner.go` (Go 백엔드, 활동 로그 분석 + 학습 추출)
- `internal/harness/observer.go` (PostToolUse hook 기반 활동 수집)
- `.moai/harness/usage-log.jsonl` (활동 로그 schema)
- `.moai/harness/learning-history/` (학습 변경 이력 + rollback)
- `/moai harness` 신규 subcommand:
  - `/moai harness status` — 현재 학습 상태 표시
  - `/moai harness apply` — pending 학습 변경 수동 적용
  - `/moai harness rollback <date>` — 특정 시점으로 복원
  - `/moai harness disable` — 자동 학습 일시 중단
- `.moai/config/sections/harness.yaml`:
  ```yaml
  harness:
    learning:
      enabled: true
      auto_apply: false  # 기본 false (사용자 승인 모드)
      tier_thresholds: { observation: 1, heuristic: 3, rule: 5, auto_update: 10 }
      rate_limit: { per_week: 3, cooldown_hours: 24 }
      log_retention_days: 90
  ```

#### 통합 위치

- v2.17 phase에 포함 (HARNESS-001 + DESIGN-PIPELINE-001 + PROJECT-HARNESS-001 + **LEARNING-001** = 4 SPECs)
- 의존성:
  - HARNESS-001 (meta-harness) 선행 필수
  - PROJECT-HARNESS-001 (소크라테스 인터뷰) 선행 — 사용자 baseline 확보
  - design.yaml safety architecture (Section 5) 활용

#### Risks & Mitigations

| Risk | Mitigation |
|---|---|
| 자동 변경이 사용자 의도와 다름 | Tier 4 자동 적용 전 AskUserQuestion 의무, default `auto_apply: false` |
| 학습 결과 발산 (drift) | Rate Limiter + 주간 회고 보고 (`/moai harness status`) |
| 활동 로그 privacy 우려 | 로컬 보존만, 외부 전송 없음, 90일 retention |
| moai-managed 영역 침범 | Frozen Guard L1 강제 — 절대 변경 불가 |

---

## 5. 핵심 검증 시나리오 (v2.17 완료 후)

### 5.1 Static Core 검증
```bash
ls .claude/skills/moai-* | wc -l    # 22 + 1 meta = 23
ls .claude/agents/moai/ | wc -l      # 22 (변동 없음)
ls .claude/commands/moai/ | wc -l    # 16 (gate 추가, security/context 제거)
```

### 5.2 Dynamic Harness 시나리오 (iOS 프로젝트 예시)
```bash
# 새 빈 프로젝트
$ moai init my-ios-app && cd my-ios-app
$ /moai project

# 소크라테스 인터뷰 (AskUserQuestion 6 라운드)
# Q1: 도메인? → Mobile (iOS)
# Q2: 팀? → Solo
# Q3: 방법론? → TDD
# Q4: 디자인? → Figma
# Q5: 배포? → App Store
# Q6: customization? → 표준

# meta-harness 자동 호출
# → .claude/agents/my-harness/{ios-architect,swiftui-engineer,xcode-tester}.md 생성
# → .claude/skills/my-harness-{ios-patterns,swiftui-best-practices,xctest-fixtures}/ 생성
# → .moai/config/sections/workflow.yaml.harness 섹션 갱신
# → CLAUDE.md에 @import marker inject

# 새 세션 시작
$ claude
# CLAUDE.md 로드 → @import follow → harness 활성

$ /moai plan "user authentication with FaceID"
# manager-spec이 ios-architect를 자동 chain
# my-harness-ios-patterns가 paths 매칭으로 활성
# SPEC 작성 시 iOS-specific patterns 적용

$ /moai run SPEC-AUTH-001
# manager-tdd가 workflow.yaml.harness.chaining_rules read
# Phase 2: ios-architect → expert-frontend → swiftui-engineer 순서 chain
# my-harness-* skills 자동 활성
```

### 5.3 Static Maintenance 검증
```bash
# moai update 시 my-harness/* 보존
$ moai update
# .claude/agents/moai/* 갱신 ✅
# .claude/skills/moai-*/* 갱신 ✅
# .claude/agents/my-harness/* 보존 ✅ (사용자 customization)
# .claude/skills/my-harness-*/* 보존 ✅
```

---

## 6. Risks & Mitigations

| Risk | Mitigation |
|---|---|
| 16 skills 제거가 사용자에게 큰 BC | grace window (1 minor release), migration guide, moai update 자동 처리 |
| harness 동적 생성 결과 품질 편차 | revfactory/harness +60% A/B 데이터 + evaluator-active 검증 |
| 4-layer 통합 장치 복잡도 | 단계 도입 (Layer 1 → 2 → 3 → 4 순차), 각 layer 독립 테스트 |
| CLAUDE.md @import 5-hop 한계 | rule 파일을 평탄화 (depth 2 이내), markers로 관리 |
| paths frontmatter 버그 (Read-only) | globs 대안 사용, doctor 진단 추가 |
| Phase 0.95 hook 미구현 시 fallback | "no harness configured" 메시지 후 default workflow 진행 |
| Token 폭발 (소크라테스 인터뷰) | runtime.yaml.budget으로 예산 강제 |

---

## 7. 핸드오프 체크리스트

다음 세션 시작 시 확인:

- [ ] `git log --oneline -5` 으로 main HEAD = `3ed706a5b` 확인
- [ ] `.moai/release/v3r3-extreme-aggressive-handoff.md` (이 파일) 읽기
- [ ] `.moai/release/v2.15.0-draft.md` 읽기 (release notes draft)
- [ ] `.moai/specs/SPEC-V3R3-{ARCH-007,COV-001}` 작성됨 확인
- [ ] V3R3-CMD-CLEANUP-001 SPEC 미작성 확인 (이 세션에서 신규)
- [ ] PR #709 상태 확인 (`gh pr view 709`) — v2.15 PR과 별도

진행 순서 (다음 세션):
1. CMD-CLEANUP-001 SPEC 작성 → 검증
2. ARCH-007 구현 → make build → test
3. COV-001 구현 → make build → test
4. CMD-CLEANUP-001 구현 → make build → test
5. v2.15 release prep
6. PR + tag

각 단계마다 75% 컨텍스트 모니터링. 도달 시 progress.md 저장 + 새 세션.

---

## 8. 핵심 결정 요약 (1줄 정리)

```
moai-adk = workflow framework (22 base skills) + meta-harness (∞ generated)
   - Static: .claude/agents/moai/, .claude/skills/moai-*/, .claude/skills/moai/workflows/*.md
   - Dynamic: .claude/agents/my-harness/, .claude/skills/my-harness-*/, .moai/harness/
   - ★ moai workflow files 절대 수정 안 함 (template static import line 한 줄만 추가)
   - 통합 장치: 5-layer
     L1: skill frontmatter triggers (자동 활성)
     L2: workflow.yaml harness registry
     L3: CLAUDE.md @import markers
     L4: workflow static import line (moai-managed, 한 줄)
     L5: .moai/harness/ user directory (사용자 영역, moai update 보존)
   - Vibe Design: DTCG 2025.10 단일 토큰 표준
   - claude-code-guide: 공식 Claude Code 메타 위임 (foundation-cc는 보존)
   - /moai project: 소크라테스 인터뷰 16 질문 (4 라운드) + harness 자동 분기
   - /moai security: 흡수 (review/sync 강화) — command 추가 안 함
   - /moai context: 제거 — typed memory(EXT-001)만 사용
   - /moai gate: 추가 (V2.15)
   - ★ Self-Learning Harness (LEARNING-001, v2.17): 사용자 활동 학습 → harness 자동 진화
     - 4-tier (1x observation / 3x heuristic / 5x rule / 10x auto-update)
     - 5-layer safety (Frozen Guard / Canary / Contradiction / Rate Limit / Human Oversight)
     - moai-managed 영역 절대 불가침, my-harness/ 영역만 자동 업데이트
```

---

**Status**: Ready for next session
**Last Updated**: 2026-04-26 (§4.3 LEARNING-001 추가)
**Author**: MoAI Orchestrator (Claude Opus 4.7)
**Reviewed**: User decision integrated (6 decisions + self-learning 요구사항)

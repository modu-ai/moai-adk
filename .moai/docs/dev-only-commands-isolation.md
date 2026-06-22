# Dev-Only Commands Isolation (devkit Harness)

> Externalized verbatim from CLAUDE.local.md §21 on 2026-05-20 (v2.20.0-rc1 release-readiness consolidation). Original section authored 2026-04~05.
> **2026-06-22 — SPEC-V3R6-DEV-HARNESS-CONSOLIDATION-001**: the three numeric dev-only commands (`97-release-update` / `98-github` / `99-release`) were consolidated into the single `devkit` dev-maintainer harness. The numeric commands and their old workflow/agent bodies were removed; the doctrine below tracks the new harness-namespaced artifacts.

---


### [HARD] devkit 하네스 + 관련 산출물은 로컬 moai-adk 개발 전용

`devkit` 하네스 (`/harness:devkit` 진입 + 3 capability) 와 그 산출물은 모두 moai-adk-go 메인테이너 전용 도구다. 패키지 사용자 프로젝트에는 **절대 배포되어서는 안 된다**. `internal/template/templates/` 어디에도 흔적이 남으면 안 된다.

devkit 하네스는 3개 capability 를 단일 진입점으로 통합한다:
- `release-update` — 외부 시스템 추적/동기화 (CC upstream tracker; 구 `97-release-update`)
- `github` — GitHub 등 외부 플랫폼 워크플로우 (issue/PR; 구 `98-github`)
- `release` — production release 워크플로우 (Enhanced GitHub Flow; 구 `99-release`)

> **Retired numeric prefixes (historical)**: 구 `97-*` / `98-*` / `99-*` 번호 prefix 슬래시 커맨드 관례는 SPEC-V3R6-DEV-HARNESS-CONSOLIDATION-001 에서 폐지되었다. 세 번호 커맨드는 단일 `/harness:devkit` 진입점으로 통합되었고, dev-only 격리 보호는 번호 grep 에서 harness-namespace embedded-tree-absence 단언으로 이전되었다.

### 배포 금지 파일 목록

| 파일 경로 | 목적 | 격리 이유 |
|---------|------|----------|
| `.claude/commands/harness/devkit.md` | Entry slash command — `/harness:devkit` thin wrapper (3 capability dispatch) | 사용자 프로젝트에는 CC 추적 / `gh` / release 권한 미보장 |
| `.claude/commands/harness/manifest.json` | devkit 하네스 SSOT manifest (3 specialist) | 메인테이너 전용 하네스 설정, 사용자 프로젝트 무관 |
| `.claude/workflows/harness-devkit-run.js` | Runner (비-상호작용 리서치 fan-out 전용) | 메인테이너 전용 dynamic-workflow 스크립트 |
| `.claude/agents/harness/harness-devkit-release-update-specialist.md` | Specialist body — CC upstream tracker 9-phase workflow | 메인테이너 전용 하네스 specialist, 사용자 프로젝트 무관 |
| `.claude/agents/harness/harness-devkit-github-specialist.md` | Specialist body — GitHub issue/PR workflow | 메인테이너 전용 하네스 specialist, 사용자 프로젝트 무관 |
| `.claude/agents/harness/harness-devkit-release-specialist.md` | Specialist body — production release (Enhanced GitHub Flow) | 동일 — dev maintainer 전용, project-local git workflow doctrine |
| `.moai/state/last-cc-version.json` | 마지막 분석 버전 + history | 사용자 프로젝트별 상태 추적 불요 |
| `.moai/research/cc-update-*.md` | 분석 보고서 + update plan | dev 산출물, 사용자 사용 안 함 |

> **Migration note (2026-06-22 — SPEC-V3R6-DEV-HARNESS-CONSOLIDATION-001)**: The three numeric dev-only commands (`.claude/commands/97-release-update.md` / `98-github.md` / `99-release.md`) and their workflow/agent bodies (`.claude/agents/local/release-update-specialist.md`, `.claude/agents/local/github-specialist.md`, `.claude/skills/moai/workflows/release.md`) were **deleted** in M5 of this SPEC and consolidated into the single `devkit` dev-maintainer harness. The three multi-phase workflow bodies migrated with structural fidelity into `.claude/agents/harness/harness-devkit-{release-update,github,release}-specialist.md`; the three numeric entry points collapsed into one `/harness:devkit` thin command that dispatches by sub-command. The Runner (`.claude/workflows/harness-devkit-run.js`) models ONLY the non-interactive research fan-out; all human-gated work (user approval, PR creation, gh CLI, production-release gate) is held by the specialists and the orchestrator. The harness lives in the user-owned namespace (`moai update` preserves it) and is dev-only. This note supersedes the earlier 2026-05-25 SPEC-V3R6-LOCAL-NAMESPACE-CONSOLIDATION-001 migration (which had moved 97/98 bodies into `.claude/agents/local/`) — that intermediate `local/` namespace was removed in this consolidation.

### [HARD] 검증 체크리스트 (dev-only 도구 변경 시 매번)

- [ ] `find internal/template/templates -path "*commands/harness*"` 결과 비어있음
- [ ] `find internal/template/templates -path "*harness-devkit*"` 결과 비어있음
- [ ] `find internal/template/templates -path "*agents/harness*"` 결과 비어있음 (harness namespace 는 user-owned — 사용자 프로젝트에 template 으로 배포 금지)
- [ ] `find internal/template/templates -name "last-cc-version.json"` 결과 비어있음
- [ ] `find internal/template/templates -name "cc-update-*.md"` 결과 비어있음
- [ ] `find internal/template/templates -name "97-*" -o -name "98-*" -o -name "99-*"` 결과 비어있음 (구 번호 커맨드 잔재 없음 — 폐지됨)
- [ ] `find internal/template/templates -path "*/agents/local/*"` 결과 비어있음 (구 local namespace 제거됨)
- [ ] `go test ./internal/template/... -run TestDevkitNamespaceNoLeak` PASS (embedded-tree-absence 단언; `internal/template/devkit_namespace_test.go` — `commands/harness/` + `harness-devkit*` + `workflows/harness-devkit-*` 부재 검출, sentinel `DEVKIT_NAMESPACE_LEAK`)
- [ ] `moai init test-project` 후 위 모두 사용자 프로젝트에 복사되지 않음 확인

> **CI guard (자동)**: dev-only 격리 보호의 1차 메커니즘은 더 이상 doc-level 수동 grep 체크리스트가 아니라 `internal/template/devkit_namespace_test.go` (`TestDevkitNamespaceNoLeak`) 의 embedded-tree-absence 단언이다. `make build` 가 누출된 harness-devkit 경로를 `internal/template/templates/` 하위에서 임베드하면 `go test ./internal/template/...` 가 FAIL 한다. `embedded_namespace_test.go` (`TestTemplateAgentsStructure`) 의 `{moai}`-only allowlist 가 이미 `.claude/agents/harness/` 부재를 보호하므로, 신규 단언은 commands/workflows 차원 + `harness-devkit` 이름 차원을 보완한다.

만약 위 파일 중 하나가 `internal/template/templates/`에 발견되면 즉시 `git rm` + `make build` 재실행 + commit.

### 워크플로우 본문 자체에도 dev-only 배너 필수

- `.claude/agents/harness/harness-devkit-release-update-specialist.md` 최상단에 `> **[DEV-ONLY]**` 경고 banner 유지
- `.claude/agents/harness/harness-devkit-github-specialist.md` 최상단에 `> **[DEV-ONLY]**` 경고 banner 유지
- `.claude/agents/harness/harness-devkit-release-specialist.md` 최상단에 `> **[DEV-ONLY]**` 경고 banner 유지
- `.claude/commands/harness/devkit.md` frontmatter `description` + body 에 `"(dev-only)" / "NOT distributed to user projects"` 문구 유지
- `.claude/workflows/harness-devkit-run.js` 상단 주석에 `[DEV-ONLY]` 명시 유지

### 위반 시 영향

사용자가 `moai init my-project` 실행 시:

- `commands/harness/devkit.md` 배포 → `/harness:devkit` 슬래시 커맨드가 사용자 UI에 노출. 사용자에게 CC 추적 / `gh` / release 권한 부재 → 실행 시 오류 + 혼란
- `harness-devkit-*-specialist.md` 배포 → 사용자 repo에 무관한 dev-only 워크플로우(CC tracker / PR/issue / production release) 가 specialist 로 등록되어 의도치 않은 호출 가능 → 혼란 + production release 위험
- `harness-devkit-run.js` 배포 → 사용자 `.claude/workflows/` 에 메인테이너 전용 Runner 가 섞임
- `last-cc-version.json` 배포 → 사용자 프로젝트가 moai-adk 자체의 CC 추적 상태를 들고다님 (의미 없음)
- `cc-update-*.md` 배포 → 사용자 `.moai/research/`에 메인테이너 보고서 섞임

### 관련 정책

- §2 File Synchronization "Local-Only Files (Never in Templates)" 등록 — 본 §21은 그 카테고리의 명시적 expansion (devkit 하네스 항목 포함)
- §15 템플릿 언어 중립성 의 16-language equivalence 원칙과 별개의 dev-only 격리 룰
- §17 docs-site 4-locale sync 규칙은 패키지 배포의 일부 (dev-only 아님) — 헷갈리지 말 것
- §24 Harness Namespace 분리 정책 — `harness-*` / `.claude/agents/harness/` / `.claude/commands/harness/` / `.claude/workflows/harness-*.js` 는 user-owned (`moai update` 보존); devkit 하네스가 이 user-owned namespace 에 위치

### 신규 dev-only capability 추가 시

향후 비슷한 메인테이너 전용 capability 를 추가할 때는:

1. devkit 하네스의 `manifest.json` `specialists` 배열에 항목 추가 (또는 신규 하네스가 적절하면 별도 하네스 생성) + 본 "배포 금지 파일 목록" 표에 행 추가
2. §2 Local-Only Files 목록에 등록 (이미 `commands/harness/devkit*` / `workflows/harness-devkit-run.js` 패턴은 등록되어 있음 — 패턴 외 추가 파일만 명시 추가)
3. specialist body 최상단 `[DEV-ONLY]` banner 추가
4. entry command frontmatter `description`에 `"(dev-only). NOT distributed to user projects."` 문구 유지
5. `TestDevkitNamespaceNoLeak` (또는 신규 하네스용 embedded-tree-absence 단언) 가 신규 artifact 이름/경로를 커버하는지 확인

---


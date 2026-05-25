# Dev-Only Commands Isolation (97/98/99 Series)

> Externalized verbatim from CLAUDE.local.md §21 on 2026-05-20 (v2.20.0-rc1 release-readiness consolidation). Original section authored 2026-04~05.

---


### [HARD] 97/98/99 prefix 슬래시 커맨드 + 관련 산출물은 로컬 moai-adk 개발 전용

`97-*`, `98-*`, `99-*` prefix 슬래시 커맨드는 모두 moai-adk-go 메인테이너 전용 도구다. 패키지 사용자 프로젝트에는 **절대 배포되어서는 안 된다**. `internal/template/templates/` 어디에도 흔적이 남으면 안 된다.

번호 prefix 의미 (예약):
- `97-*` — 외부 시스템 추적/동기화 (예: CC upstream tracker)
- `98-*` — GitHub 등 외부 플랫폼 워크플로우
- `99-*` — 내부 감사/cleanup/실험 (예약)

### 배포 금지 파일 목록

| 파일 경로 | 목적 | 격리 이유 |
|---------|------|----------|
| `.claude/commands/97-release-update.md` | Entry slash command — CC upstream tracker | 사용자 프로젝트에는 CC 추적 권한 부재 |
| `.claude/commands/98-github.md` | Entry slash command — GitHub workflow with Agent Teams | 사용자 프로젝트에 `gh` 권한/repo 컨텍스트 미보장 |
| `.claude/agents/local/release-update-specialist.md` | Agent body — CC upstream tracker 9-phase workflow (97 entry target) | 메인테이너 전용 hand-authored agent, 사용자 프로젝트 무관 |
| `.claude/agents/local/github-specialist.md` | Agent body — GitHub issue/PR workflow (98 entry target) | 메인테이너 전용 hand-authored agent, 사용자 프로젝트 무관 |
| `.claude/skills/moai/workflows/release.md` | Production release 워크플로우 본문 (99 entry) | 동일 — dev maintainer 전용, Enhanced GitHub Flow (project-local git workflow doctrine) |
| `.claude/commands/99-release.md` | Entry slash command — production release | 사용자 프로젝트에는 release 권한/repo 컨텍스트 미보장 |
| `.moai/state/last-cc-version.json` | 마지막 분석 버전 + history | 사용자 프로젝트별 상태 추적 불요 |
| `.moai/research/cc-update-*.md` | 분석 보고서 + update plan | dev 산출물, 사용자 사용 안 함 |

> **Migration note (2026-05-25 — SPEC-V3R6-LOCAL-NAMESPACE-CONSOLIDATION-001)**: The `.claude/skills/moai/workflows/release-update.md` and `.claude/skills/moai/workflows/github.md` skill bodies were removed in M3 of the local-namespace consolidation SPEC. Their 9-phase / multi-phase workflow contents migrated verbatim (with structural fidelity preserved) into the two new `.claude/agents/local/*-specialist.md` agent files listed above. The thin command wrappers (`97-release-update.md` + `98-github.md`) now delegate to the local agents via `Use the <name>-specialist subagent` routing instead of `Skill("moai/workflows/<name>")` invocation. The migration retained the Thin Command Pattern (body ≤ 20 LOC) on both wrappers.

### [HARD] 검증 체크리스트 (dev-only 커맨드 변경 시 매번)

- [ ] `find internal/template/templates -name "97-*"` 결과 비어있음
- [ ] `find internal/template/templates -name "98-*"` 결과 비어있음
- [ ] `find internal/template/templates -name "99-*"` 결과 비어있음
- [ ] `find internal/template/templates -name "release-update.md"` 결과 비어있음
- [ ] `find internal/template/templates -name "github.md" -path "*/workflows/*"` 결과 비어있음
- [ ] `find internal/template/templates -name "release.md" -path "*/workflows/*"` 결과 비어있음
- [ ] `find internal/template/templates -name "last-cc-version.json"` 결과 비어있음
- [ ] `find internal/template/templates -name "cc-update-*.md"` 결과 비어있음
- [ ] `find internal/template/templates -path "*/agents/local/*"` 결과 비어있음 (HARD — SPEC-V3R6-LOCAL-NAMESPACE-CONSOLIDATION-001 REQ-LNC-012)
- [ ] `find internal/template/templates -name "release-update-specialist.md"` 결과 비어있음
- [ ] `find internal/template/templates -name "github-specialist.md"` 결과 비어있음
- [ ] `moai init test-project` 후 위 모두 사용자 프로젝트에 복사되지 않음 확인

만약 위 파일 중 하나가 `internal/template/templates/`에 발견되면 즉시 `git rm` + `make build` 재실행 + commit.

### 워크플로우 본문 자체에도 dev-only 배너 필수

- `.claude/skills/moai/workflows/release-update.md` 최상단에 `> **[DEV-ONLY]**` 경고 banner 유지
- `.claude/skills/moai/workflows/github.md` 최상단에 `> **[DEV-ONLY]**` 경고 banner 유지 (없으면 추가)
- `.claude/skills/moai/workflows/release.md` frontmatter `description`에 `"(dev-only) MoAI-ADK production release skill"` 명시 유지 (body banner 또는 frontmatter 명시 중 하나로 충분)
- `.claude/commands/97-release-update.md` frontmatter `description`에 `"NOT distributed to user projects"` 문구 유지
- `.claude/commands/98-github.md` frontmatter `description`에 `"(dev-only). NOT distributed to user projects."` 문구 유지
- `.claude/commands/99-release.md` frontmatter `description`에 `"NOT distributed to user projects (dev-only)"` 문구 유지

### 위반 시 영향

사용자가 `moai init my-project` 실행 시:

- `97-release-update.md` 배포 → `/97-release-update` 슬래시 커맨드가 사용자 UI에 노출. 사용자에게 권한 부재 → 실행 시 오류 + 혼란
- `98-github.md` 배포 → `/98-github` 슬래시 커맨드가 사용자 UI에 노출. 사용자 repo에 무관한 dev-only PR/issue 관리 워크플로우 호출 가능 → 혼란
- `workflows/release-update.md` 배포 → MoAI 스킬 intent router에 등록되어 "release-update" 키워드 자동 매칭 → 의도치 않은 routing
- `workflows/github.md` 배포 → MoAI 스킬 intent router에 "github" 키워드 자동 매칭 → 사용자가 의도하지 않은 PR 관리 시도
- `99-release.md` 배포 → `/99-release` 슬래시 커맨드가 사용자 UI에 노출. 사용자 repo에 release 권한/PR merge 권한 부재 → 실행 시 오류 + 혼란
- `workflows/release.md` 배포 → MoAI 스킬 intent router에 "release" 키워드 자동 매칭 → 사용자가 의도하지 않은 production release 시도 + 위험
- `last-cc-version.json` 배포 → 사용자 프로젝트가 moai-adk 자체의 CC 추적 상태를 들고다님 (의미 없음)
- `cc-update-*.md` 배포 → 사용자 `.moai/research/`에 메인테이너 보고서 섞임

### 관련 정책

- §2 File Synchronization "Local-Only Files (Never in Templates)" 등록 — 본 §21은 그 카테고리의 명시적 expansion
- §15 템플릿 언어 중립성 의 16-language equivalence 원칙과 별개의 dev-only 격리 룰
- §17 docs-site 4-locale sync 규칙은 패키지 배포의 일부 (dev-only 아님) — 헷갈리지 말 것

### 신규 dev-only 워크플로우 추가 시

향후 비슷한 메인테이너 전용 워크플로우 (예: `/99-internal-audit`, `/99-cleanup-script`)를 추가할 때는:

1. 본 §21 "배포 금지 파일 목록" 표에 행 추가 (entry command + workflow body 양쪽)
2. §2 Local-Only Files 목록에 등록 (이미 `97-*.md`/`98-*.md`/`99-*.md` 패턴은 등록되어 있음 — 패턴 외 추가 파일만 명시 추가)
3. workflow body 최상단 `[DEV-ONLY]` banner 추가
4. entry command frontmatter `description`에 `"(dev-only). NOT distributed to user projects."` 문구 유지
5. 본 §21 검증 체크리스트에 신규 파일명 grep 항목 추가

---


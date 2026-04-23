---
id: SPEC-V3-SKL-002
title: "Skill Drift Detection & Resolution (template vs local parity)"
version: "0.1.0"
status: draft
created: 2026-04-22
updated: 2026-04-22
author: GOOS
priority: P1 High
phase: "v3.0.0 — Phase 5 Internal Cleanup"
module: "internal/cli/doctor/, internal/template/sync/"
dependencies:
  - SPEC-V3-MIG-001
  - SPEC-V3-CLN-001
related_gap:
  - gm#184
related_theme: "Theme 9 — Internal Cleanup & Template Drift Resolution"
breaking: false
bc_id: []
lifecycle: spec-anchored
tags: "skill, drift, template, sync, moai-doctor, v3"
---

# SPEC-V3-SKL-002: Skill Drift Detection & Resolution (template vs local parity)

## HISTORY

| Version | Date       | Author | Description                                              |
|---------|------------|--------|----------------------------------------------------------|
| 0.1.0   | 2026-04-22 | GOOS   | Initial draft (Wave 4 SPEC writer)                       |

---

## 1. Goal (목적)

template (50 skills)과 local (47 skills) 간의 skill drift를 공식 감지·해소 메커니즘으로 정형화한다. 현재 3개 스킬(`moai-domain-db-docs`, `moai-workflow-design-context`, `moai-workflow-pencil-integration`)이 template에만 존재하고 local에 배포되지 않은 상태다(findings-wave1-moai-current.md §7.2; §15.1; gap matrix #184). v3.0에서는 `moai doctor` 계통 명령이 이 드리프트를 감지하고 `moai migrate v2-to-v3` 의 migration step M04(SPEC-V3-MIG-002 기준)가 자동 해소하도록 한다.

### 1.1 배경

findings-wave1-moai-current.md §7.1–7.2 (인용):

> - `.claude/skills/` (local): 47 skill dirs (including `moai/` entry skill).
> - `internal/template/templates/.claude/skills/` (template): 50 skill dirs.
>
> Template-only (3 additions vs local):
> - `moai-domain-db-docs/` (template only)
> - `moai-workflow-design-context/` (template only)
> - `moai-workflow-pencil-integration/` (template only)
>
> These are forward-looking skills not yet copied to local.

이 드리프트는 사용자에게 기능 누락으로 이어지며, `moai update` 명령에서 silent하게 발생했다. v3.0은 다음 세 층위에서 해소한다:

1. **Detection** — `moai doctor skill --drift` 명령 신설, drift report 출력
2. **Resolution** — `moai migrate v2-to-v3` M04 migration이 template에서 local로 3개 스킬을 byte-identical copy
3. **Prevention** — `moai update` 이후에도 drift 재발 감지, warning 출력

### 1.2 Non-Goals

- 로컬에만 존재하는 skill의 template으로의 자동 promotion (policy: 사용자 custom skill은 local-only 유지)
- Skill content 비교(checksum 외 의미론적 diff; 개별 skill 버전 관리)
- Skill removal migration(drift가 "template이 skill 제거"인 경우; v3.1+)
- User-facing GUI (moai-adk는 CLI only)
- 3rd-party marketplace skill drift (SPEC-V3-PLG-001 범위)
- Template 내 skill 자체의 일관성 검증(CI 별도)
- 기존 47개 스킬 변경(새 3개만 추가)

---

## 2. Scope (범위)

### 2.1 In Scope

- `moai doctor skill --drift` 하위 명령 신설 (`internal/cli/doctor/skill_drift.go`)
- Drift 판정 알고리즘: template vs local `.claude/skills/` 디렉터리 비교 (set difference)
- Drift report 출력 포맷: template-only, local-only, both 세 카테고리
- SPEC-V3-MIG-002의 M04 migration에 대한 스킵/실행 조건 정의
- 3개 누락 skill 명시적 목록 (`moai-domain-db-docs`, `moai-workflow-design-context`, `moai-workflow-pencil-integration`) + 각 skill의 배포 경로
- `moai update` 실행 후 post-hook: drift 감지 → warning 출력
- Drift resolution 후 `.moai/state/last-skill-sync.json` 갱신 (migration idempotency)

### 2.2 Out of Scope (Exclusions — What NOT to Build)

- Skill 내용 semantic diff (checksum-based binary-identity check only)
- Skill removal workflow (template이 skill을 제거하는 경우의 migration)
- User's local-only custom skill에 대한 promotion/demotion
- Skill version field(SPEC-V3-SKL-001 Non-Goals와 동일)
- Skill marketplace integration (SPEC-V3-PLG-001)
- Interactive UI for conflict resolution (CLI non-interactive only)
- Cross-project skill sync (각 프로젝트 독립)
- Skill-level telemetry 수집

---

## 3. Environment (환경)

- 런타임: moai-adk-go v3.0.0+ (Go 1.23+)
- 의존: SPEC-V3-MIG-001(migration framework), SPEC-V3-CLN-001(template drift resolution migrations)
- 영향 디렉터리: `.claude/skills/`, `internal/template/templates/.claude/skills/`, `internal/cli/doctor/`, `internal/template/sync/`
- 현 drift 상태 (v2.12.0 기준):
  - template: 50개 skill 디렉터리
  - local: 47개 skill 디렉터리
  - template-only 3: `moai-domain-db-docs`, `moai-workflow-design-context`, `moai-workflow-pencil-integration`
- OS 동등성: macOS / Linux / Windows
- 참조: findings-wave1-moai-current.md §7.1, §7.2, §15.1; gap matrix #184

---

## 4. Assumptions (가정)

- Template은 `go:embed` 기반이며 빌드 시점에 모든 50 skill이 포함된다 (`internal/template/embedded.go` 자동 생성).
- Local의 47개 skill은 사용자가 v2.12.0 이전 버전에서 초기화했으며 이후 3개 신규 skill이 template에 추가되었다.
- Template이 변경되지 않는 한 drift set은 재현 가능하고 결정적이다.
- 3개 누락 skill의 추가는 사용자 시스템에 breaking change를 유발하지 않는다 (opt-in trigger가 부재하면 로드되지 않음).
- `moai update`는 v3.0 이후 drift 감지를 post-step으로 수행할 수 있다.
- Migration framework(SPEC-V3-MIG-001)의 M04 step이 skill drift 전담이다.

---

## 5. Requirements (EARS 요구사항)

### 5.1 Ubiquitous (항시)

**REQ-SKL-002-001 (Ubiquitous) — drift 감지 알고리즘**
The drift detector **shall** compute three sets: `template_only = template - local`, `local_only = local - template`, `both = template ∩ local` based on skill directory names under `.claude/skills/` only.

**REQ-SKL-002-002 (Ubiquitous) — identity metric**
The detector **shall** use skill **directory name** as the identity key. No skill content comparison is performed in v3.0.

**REQ-SKL-002-003 (Ubiquitous) — 3 known missing skills**
The detector **shall** recognize the following 3 skill names as expected template-only drift in v2.12.x and emit them in the drift report with a `known_in_v2_12` annotation:
- `moai-domain-db-docs`
- `moai-workflow-design-context`
- `moai-workflow-pencil-integration`

**REQ-SKL-002-004 (Ubiquitous) — resolution via M04**
The v3.0 migration framework **shall** include a migration step `M04-skill-drift` (defined in SPEC-V3-MIG-002) that copies the 3 known missing skills from template to local byte-identically.

### 5.2 Event-Driven (이벤트 기반)

**REQ-SKL-002-005 (Event-Driven) — moai doctor 실행**
**When** the user runs `moai doctor skill --drift`, the command **shall** output a structured report in YAML format with sections `template_only`, `local_only`, `both`, and final summary line `Drift: {N} skill(s) need attention`.

**REQ-SKL-002-006 (Event-Driven) — moai update post-hook**
**When** `moai update` completes its template sync phase, a post-hook **shall** invoke the drift detector and, if `template_only` is non-empty AND the skills are not in the known-acceptable list, emit a stderr warning with the skill names and remediation steps.

**REQ-SKL-002-007 (Event-Driven) — M04 execution**
**When** `M04-skill-drift` migration is applied (via `moai migrate v2-to-v3` or `moai migrate --only M04`), the migration **shall**:
1. For each skill in the known missing list, copy `internal/template/templates/.claude/skills/{skill}/` recursively to `.claude/skills/{skill}/` preserving file permissions per platform rules (POSIX: chmod; Windows: no-op per CLAUDE.local.md lessons.md #7)
2. Record the sync timestamp and skill list in `.moai/state/last-skill-sync.json`
3. Emit success summary with count of skills added

### 5.3 State-Driven (상태 기반)

**REQ-SKL-002-008 (State-Driven) — dry-run**
**While** `--dry-run` flag is present on `moai migrate v2-to-v3 --only M04`, the migration **shall** output the would-be copy list without writing any files.

**REQ-SKL-002-009 (State-Driven) — idempotency**
**While** `.moai/state/last-skill-sync.json` records a successful sync for a given skill AND the target file already exists with matching mtime, the migration **shall** skip re-copying that skill.

### 5.4 Optional (선택)

**REQ-SKL-002-010 (Optional) — exit code**
**Where** the user invokes `moai doctor skill --drift` with `--strict` flag AND `template_only` is non-empty, the command **shall** exit with code 2 (signals CI to fail).

**REQ-SKL-002-011 (Optional) — JSON output**
**Where** the user passes `--json` flag, the drift report **shall** be emitted as JSON for scripting.

### 5.5 Unwanted Behavior

**REQ-SKL-002-012 (Unwanted Behavior) — partial copy failure**
**If** the M04 migration fails partway (e.g., disk full, permission denied), **then** the migration **shall** roll back any partially-copied skills (delete partially-written directories) and return error `SKL_M04_PARTIAL_FAILURE` with the failed skill name.

**REQ-SKL-002-013 (Unwanted Behavior) — unexpected local-only skill**
**If** `local_only` is non-empty for skills beginning with `moai-` prefix (which are reserved for the canonical 50-skill catalog), **then** the drift report **shall** flag these as "potential removed skill" and suggest filing a bug.

**REQ-SKL-002-014 (Unwanted Behavior) — M04 after manual copy**
**If** the user has manually copied the 3 missing skills before running M04 AND the copied content byte-diffs from template, **then** M04 **shall** NOT overwrite; instead emit warning `SKL_M04_LOCAL_MODIFIED` listing the divergent skill and recommend manual resolution.

---

## 6. Acceptance Criteria (수용 기준 요약)

- **AC-SKL-002-01**: `moai doctor skill --drift` 실행 시 3개 template-only skill 목록 출력 + `known_in_v2_12: true` 표시
- **AC-SKL-002-02**: `moai migrate v2-to-v3 --only M04 --dry-run` 실행 시 copy 예정 파일 목록만 출력, 파일 시스템 변경 없음
- **AC-SKL-002-03**: `moai migrate v2-to-v3 --only M04` (실행) 시 3개 skill이 local에 byte-identical 복사됨 (SHA256 검증)
- **AC-SKL-002-04**: M04 2회 연속 실행 시 두 번째 실행은 idempotent (`last-skill-sync.json` match → skip)
- **AC-SKL-002-05**: M04 실행 중 중간에 권한 오류 발생 → `SKL_M04_PARTIAL_FAILURE` + 부분 복사 롤백
- **AC-SKL-002-06**: 사용자가 수동으로 한 skill을 변경한 상태에서 M04 실행 → `SKL_M04_LOCAL_MODIFIED` warning, 덮어쓰지 않음
- **AC-SKL-002-07**: `moai update` 후 drift 발견 시 stderr warning 출력
- **AC-SKL-002-08**: `moai doctor skill --drift --strict` + drift 존재 → exit code 2
- **AC-SKL-002-09**: `moai doctor skill --drift --json` → 구조화된 JSON 출력 (jq 파싱 가능)
- **AC-SKL-002-10**: `go test ./internal/cli/doctor/... ./internal/template/sync/...` 전체 통과

---

## 7. Constraints (제약)

- [HARD] 3개 known missing skill 목록은 본 SPEC의 REQ-SKL-002-003에 고정 (v3.0 한정). v3.1 이후의 drift는 별도 SPEC 또는 본 SPEC의 후속 마이너 버전에서 처리.
- [HARD] Skill content의 byte-identical 복사만 수행. 변환·개명·콘텐츠 가공 금지.
- [HARD] User's local-only skill은 건드리지 않는다 (REQ-SKL-002-013은 warning only).
- [HARD] Platform 권한 처리는 CLAUDE.local.md lessons.md #7 규칙(POSIX chmod / Windows no-op).
- [HARD] Template 우선 원칙(CLAUDE.local.md §2) 유지: `moai-` prefix 스킬은 template이 원천.
- [HARD] 하드코딩 방지: 3개 known skill 목록도 `internal/template/sync/known_missing.go`에 const로 정의.
- [HARD] 16개 언어 중립성: skill drift 감지는 언어 비편중.

---

## 8. Risks & Mitigations (리스크 및 완화)

| 리스크                                                             | 영향   | 완화                                                                                    |
|--------------------------------------------------------------------|--------|-----------------------------------------------------------------------------------------|
| User가 수동 편집한 skill이 M04로 덮여짐                            | High   | REQ-SKL-002-014 `SKL_M04_LOCAL_MODIFIED`; 수동 편집 감지 후 warn-and-skip                |
| M04 partial failure로 skill 디렉터리 orphan                        | High   | REQ-SKL-002-012 rollback; 트랜잭션 로그                                                 |
| 3개 skill이 향후 변경되어 v3.x 시점 drift 목록이 틀림              | Medium | known list는 v3.0 전용; v3.1+ drift는 별도 미니 SPEC으로 갱신                          |
| Windows 권한 보존 시 ACL 차이로 test 실패                          | Medium | CLAUDE.local.md lessons.md #7 준수; Windows CI에서 no-op 검증                           |
| `moai doctor skill --drift` 출력이 stdout을 차지해 paging 필요    | Low    | 일반 50 skill 규모에서는 paging 불필요; future --limit flag로 제어                      |
| Known-missing skill name typo (예: `moai-domain-dbdocs`)           | Low    | REQ-SKL-002-003 list + unit test로 정확성 검증                                          |
| Local-only skill false flag (사용자 의도)                          | Low    | REQ-SKL-002-013은 warning만 (non-blocking)                                              |

---

## 9. Dependencies (의존성)

### 9.1 Blocked by

- **SPEC-V3-MIG-001** (Versioned migration framework): M04 migration step의 runtime이 선행.
- **SPEC-V3-CLN-001** (sibling; Template drift resolution): M04의 구체 구현 코드는 SPEC-V3-CLN-001이 담당; 본 SPEC은 detection + trigger 계층.

### 9.2 Blocks

- (없음)

### 9.3 Related

- **SPEC-V3-MIG-002** (Initial migration set M01–M05, sibling writer SCH/MIG domain): M04-skill-drift 구체 명세는 거기서 작성 — 본 SPEC은 요구사항 공급.
- **SPEC-V3-SKL-001** (Skill frontmatter v2): 본 SPEC은 프론트매터 변경과 무관 (내용 복사만).
- **SPEC-V3-CLI-001** (sibling writer BSC domain; `moai doctor` profiler + bare mode): `moai doctor skill --drift` 하위 명령이 여기에 통합.

---

## 10. Traceability (추적성)

- 총 REQ 개수: 14 (Ubiquitous 4, Event-Driven 3, State-Driven 2, Optional 2, Unwanted Behavior 3)
- 예상 AC 개수: 10
- 관련 Wave 1 근거:
  - findings-wave1-moai-current.md §7.1 (Local vs template parity, 47 vs 50 skill dirs)
  - findings-wave1-moai-current.md §7.2 (Template-only 3 skills enumerated)
  - findings-wave1-moai-current.md §15.1 (Self-identified: skill drift)
  - gap matrix #184 (Medium severity, XS effort, No breaking)
  - master-v3 §2.3 item 2 (Technical debt inventory: template/local skill drift)
  - master-v3 §3.9 T1-CLN-01 (Template drift resolution via migration M04)
- 구현 시 각 소스 파일에 `@SPEC:SPEC-V3-SKL-002:REQ-SKL-002-NNN` 주석 부착
- 코드 구현 예상 경로:
  - `internal/cli/doctor/skill_drift.go` (REQ-SKL-002-005, 010, 011)
  - `internal/template/sync/drift.go` (REQ-SKL-002-001, 002)
  - `internal/template/sync/known_missing.go` (REQ-SKL-002-003)
  - `internal/template/sync/m04_apply.go` (REQ-SKL-002-004, 007, 012, 014)
  - `internal/template/sync/m04_idempotency.go` (REQ-SKL-002-008, 009)
  - `internal/cli/update/post_hook.go` (REQ-SKL-002-006)
  - `internal/cli/doctor/skill_drift_test.go` (AC-SKL-002-01, 08, 09)
  - `internal/template/sync/m04_test.go` (AC-SKL-002-02..06)

---

End of SPEC.

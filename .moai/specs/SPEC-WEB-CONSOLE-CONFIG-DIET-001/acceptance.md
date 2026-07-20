# acceptance.md — SPEC-WEB-CONSOLE-CONFIG-DIET-001

> 검증 SSOT. 각 AC-CD-0xx는 spec.md의 REQ-CD-0xx와 1:1 대응하며, grep(필드 부재) 또는 go-test(스키마
> 카운트/멤버십) assertion으로 기계 검증 가능하다. "은닉" 처리는 필드가 `allFields()` 렌더 집합에서
> 빠짐을 뜻하고, "read-only 강등"은 편집 `FieldDef`에서 빠지고 `ReadOnlyDisplayFields()`로 이동함을 뜻한다.
> (§F.5 결정에 따라 그룹마다 은닉/read-only 중 하나가 선택되므로, 각 AC는 "편집 필드 집합에 없음"을 공통
> 최소 조건으로 하고, read-only 대상은 추가로 `ReadOnlyDisplayFields()` 등장을 확인한다.)

## §A. 검증 헬퍼 규약

- **편집 필드 부재 검증**: `settings.allFields()`(또는 export된 동등 접근자)가 반환한 `FieldDef.Name`
  집합에 대상 키가 없어야 한다. go-test로 검증:
  ```go
  names := fieldNameSet(allFields())   // map[string]bool
  assert.False(t, names["research.enabled"])
  ```
  또는 소스 부재 grep(구현 형태에 따라):
  ```bash
  # research 블록이 seamSectionFields에서 제거되었는가
  grep -n 'SectionResearch' internal/settings/schema_sections.go   # expect: 정의 잔재 없음(또는 read-only만)
  ```
- **섹션 부재 검증**: `AllSections()`에 대상 SectionID가 없어야 한다(섹션 통째 제거 시).
- **라이브 유지 검증**: 보존 대상 키는 여전히 편집 필드 집합에 있어야 한다(회귀 방지).

## §B. Definition of Done

- [ ] M1~M4 각 AC PASS (grep/go-test 출력 관측).
- [ ] 라이브 유지 키 회귀 없음 (AC-CD-090).
- [ ] `go test ./internal/settings/... ./internal/web/... ./internal/cli/... -count=1` green.
- [ ] `golangci-lint run` zero errors.
- [ ] 두 표면(templ 콘솔·TUI 위저드) 빌드 성공.
- [ ] 정합성 결함 4건 수정 확인 (AC-CD-040..043).

## §C. Given-When-Then 시나리오 (대표 2+)

### 시나리오 1 — Tier 1 무효과 필드가 콘솔에서 사라진다

- **Given**: SPEC-WEB-CONSOLE-011 M2 상태에서 research 12키가 편집 가능하게 렌더됨.
- **When**: M1 다이어트 적용 후 `settings.allFields()`를 열거.
- **Then**: `research.*` 이름을 가진 편집 `FieldDef`가 0개다.
  ```bash
  go test ./internal/settings/ -run TestNoDeadResearchFields -count=1   # PASS
  ```

### 시나리오 2 — skill-body 소비 키(git-strategy)는 삭제되지 않는다

- **Given**: git-strategy 키는 skill body가 yaml-direct read로 소비(N3).
- **When**: M2 다이어트(중복 3-mode 축소) 적용.
- **Then**: git-strategy 라이브 키(`git_strategy.mode`, `git_strategy.hooks.pre_push`,
  `git_strategy.<mode>.automation.auto_*`)는 여전히 존재하고, 비선택 mode 프로파일만 중복 노출에서 빠진다.
  ```go
  assert.True(t, names["git_strategy.mode"])                 // 유지
  assert.True(t, names["git_strategy.team.main_branch"])     // 선택 mode 유지 (예)
  ```

### 시나리오 3 — 정합성 결함: harness prune 값 일치

- **Given**: `learning.log_retention_days`는 90으로 표시되나 prune은 하드코딩 30.
- **When**: REQ-CD-040 수정(권장: 표시/문서를 30으로 정정, 또는 config 배선).
- **Then**: 표시 값과 `defaultRetentionDays` 또는 config-읽은 값이 일치한다 (불일치 0).

## §D. AC 매트릭스 (REQ 1:1)

### D.1 상위

- **AC-CD-001**: 무효과(N1)로 판정된 그룹의 키가 편집 필드 집합에 0개.
  검증: `go test ./internal/settings/ -run TestDietNoDeadEditableFields`.
- **AC-CD-002**: templ/TUI 표면 코드에 신규 하드코딩 필드 필터가 추가되지 않음.
  ```bash
  git diff --stat -- internal/web/ internal/cli/profile_setup*.go | grep -iE 'filter|exclude' || true
  # 스키마 SSOT 외 필터 로직 부재를 리뷰로 확인 (mechanical: 신규 필터 함수 grep)
  grep -rn 'skipDeadField\|isDeadKey\|excludeKey' internal/web/ internal/cli/ --include="*.go"  # expect: empty
  ```
- **AC-CD-003**: git-strategy N3 키가 삭제되지 않음 (AC-CD-021과 결합).

### D.2 Tier 1 (M1)

- **AC-CD-010** (research 12): `AllSections()`에 `SectionResearch` 없음 AND `allFields()`에 `research.*`
  편집 필드 0개.
  ```bash
  grep -n 'SectionResearch,' internal/settings/schema.go   # expect: AllSections()/SchemaSectionIDs()에서 제거
  go test ./internal/settings/ -run TestNoDeadResearchFields -count=1
  ```
- **AC-CD-011** (db 인터뷰 3): `allFields()`에 `db.orm`/`db.multi_tenant`/`db.migration_tool` 편집 필드 없음
  (은닉) 또는 `ReadOnlyDisplayFields()`로 이동(read-only).
  ```go
  for _, k := range []string{"db.orm","db.multi_tenant","db.migration_tool"} {
    assert.False(t, editableNames[k])
  }
  ```
- **AC-CD-012** (security 스칼라 3): `security.permission.strict_mode`/`security.sandbox.required`/
  `security.sandbox.docker_image`가 편집 필드에서 부재. read-only 선택 시 `ReadOnlyDisplayFields()`에 등장 +
  no-effect 주석 i18n 키 존재.

### D.3 Tier 2 (M2)

- **AC-CD-020** (quality 무효과): `qualityExtraFields()` 유래 무효과 키가 편집 집합에서 부재/강등. **라이브
  유지**: `development_mode`, `quality.test_coverage_target` 여전히 편집 가능.
  ```go
  assert.True(t, editableNames["development_mode"])
  assert.True(t, editableNames["quality.test_coverage_target"])
  ```
- **AC-CD-021** (git-strategy 3-mode): 비선택 mode 프로파일 leaf의 중복 편집 노출이 축소됨. 라이브 유지:
  `git_strategy.mode`, `git_strategy.hooks.pre_push`, `git_strategy.<선택mode>.automation.auto_*`,
  `git_strategy.<선택mode>.main_branch`. **키 삭제 0** (grep으로 gitStrategyProfileLeaves 정의 잔존 확인).
- **AC-CD-022** (workflow_agents 14): `allFields()`에 `workflow.workflow_agents.*.model`/`.effort` 편집 필드
  0개.
  ```bash
  go test ./internal/settings/ -run TestNoWorkflowAgentEditableFields -count=1
  ```
- **AC-CD-023** (ralph 중첩): `ralph.ast_grep.*`/`ralph.loop.*`/`ralph.lsp.*`/`ralph.loop.completion.*`가
  편집 집합에서 부재/강등. **라이브 유지**: `ralph.lint_as_instruction`, `ralph.warn_as_instruction`.
  ```go
  assert.True(t, editableNames["ralph.lint_as_instruction"])
  assert.False(t, editableNames["ralph.loop.max_iterations"])
  ```

### D.4 Tier 3 (M3)

- **AC-CD-030** (workflow 12): `workflow.auto_clear.*`(4)/`workflow.token_budget.*`(3)/
  `workflow.loop_prevention.*`(2)/`workflow.worktree.auto_merge`(1)가 편집 집합에서 부재/강등.
- **AC-CD-031** (harness 6): `harness.escalation.*`/`harness.mode_defaults.*`/`harness.auto_detection.enabled`
  가 편집 집합에서 부재/강등.
- **AC-CD-032** (observability 5): `observability.trace_dir`/`report_dir`/`max_file_size_mb`/`retention_days`/
  `hook_metrics.output_path`가 편집 집합에서 부재/강등. **라이브 유지**: `observability.enabled`,
  `observability.hook_metrics.slow_hook_threshold_ms`.
- **AC-CD-033** (llm 4): `llm.performance_tier`/`llm.claude_models.{high,medium,low}`가 편집 집합에서
  부재/강등.
- **AC-CD-034** (role_profiles model/mode/effort): 7 profiles의 `role_profiles.*.model`/`.mode`/`.effort`가
  편집 집합에서 부재/강등. **[HARD] 라이브 유지**: `role_profiles.*.isolation`.
  ```go
  assert.True(t, editableNames["workflow.team.role_profiles.implementer.isolation"])  // 유지
  assert.False(t, editableNames["workflow.team.role_profiles.implementer.model"])     // 강등
  ```
  회귀 재확인:
  ```bash
  go test ./internal/cli/ -run TestWorkflowLint -count=1   # role_profiles.isolation 검증 여전히 green
  ```

### D.5 정합성 결함 (M4)

- **AC-CD-040** (harness prune): 표시 값과 실제 prune 값이 일치. 권장(옵션 B) 채택 시:
  ```bash
  grep -n 'defaultRetentionDays = 30' internal/harness/observer.go   # 하드코딩 유지 시
  # 표시/문서(config 기본값 or 콘솔 라벨)가 30을 반영하는지 확인 (불일치 0)
  ```
  옵션 A(배선) 채택 시: `PruneStaleEntries`가 `learning.log_retention_days` config를 인자로 받는지 test.
- **AC-CD-041** (security @MX stale): `internal/config/types.go`의 `SecuritySandbox` `@MX:REASON`이 실측과
  일치 (fan_in 과장 제거 또는 앵커 강등).
  ```bash
  grep -n 'consumed by sandbox/launcher.go' internal/config/types.go   # expect: 정정되어 부재 또는 실측 반영
  ```
- **AC-CD-042** (db auto_sync): `db.auto_sync` 서술이 실제 nested yaml 구조와 일치 (산문 스칼라 서술 정정).
  신뢰도 낮음 — doc-level 확인.
- **AC-CD-043** (session_name_pattern): `workflow.worktree.session_name_pattern`이 편집 불가(read-only)로
  강등되거나 세션 명명이 config를 읽음. 권장(옵션 B): 편집 필드 집합에서 부재.
  ```go
  assert.False(t, editableNames["workflow.worktree.session_name_pattern"])
  ```

### D.6 선택 (M5, SHOULD)

- **AC-CD-050** (과소 노출 라이브 키): `team.auto_selection.min_*`/`llm.glm.base_url`은 **production read
  소비자가 재확인된 경우에만** 편집 필드로 노출. 재확인 실패 시 노출하지 않음 (verification-claim-integrity —
  미확인 라이브 주장 금지). 이 AC는 SHOULD이며 미충족이 DoD를 차단하지 않는다.

### D.7 회귀 방지 (전 마일스톤 공통)

- **AC-CD-090** (라이브 키 회귀 없음): 아래 라이브 키가 다이어트 후에도 편집 필드 집합에 존재:
  `user_name`, `conversation_lang`, `model`, `effort_level`, `permission_mode`, statusline 16 세그먼트,
  `development_mode`, `quality.test_coverage_target`, git-convention 5키
  (`git_convention`, `git_convention.auto_detection.*`, `git_convention.validation.enforce_on_push`),
  `git_strategy.mode`, `ralph.lint_as_instruction`, `observability.enabled`,
  `role_profiles.*.isolation`.
  ```bash
  go test ./internal/settings/ -run TestLiveFieldsPreserved -count=1   # PASS
  ```
- **AC-CD-091** (전체 스위트): `go test ./... -count=1` green (본 SPEC 무관 병렬 실패 제외).

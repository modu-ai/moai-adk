# Plan — SPEC-TODO-ENABLE-FLAG-001

> 구현 계획. 경로는 워크트리 상대. 순서는 **되돌리기 어려움 순** — 데이터 모델과 사용자 대면 흐름을 앞에, 기계적 편집을 뒤에.

## §A Context

카드 t170의 **todo 축**을 구현한다. 피드백 축은 `SPEC-FEEDBACK-AUTO-SUBMIT-001`이 갖는다. 두 SPEC은 파일 9종을 공유하며 병합 규율은 `spec.md` §E.1의 [HARD]에 있다.

run-phase 모듈: `internal/config` · `internal/hook` · `internal/statusline` · `internal/cli/wizard` · `internal/cli` · `internal/core/project` · `internal/settings` · `internal/web` · 스킬 2사본 · 템플릿.

Tier M이므로 Route A(main-direct)가 기본이나, 형제 SPEC과 같은 배치에 실린다면 배치 레인의 통합 규율을 따른다.

## §B Known Issues (착수 전 알고 있어야 할 함정)

1. **카드가 지목한 배선 선례가 사장 코드다.** `internal/cli/init_autonomy_wizard.go`의 `applyAutonomyTierFromWizard`는 프로덕션 호출자가 없다(테스트 제외). 따라 쓰면 질문이 물어지고 `WizardResult`에 저장된 뒤 **버려진다**. 이웃 둘도 같은 상태다 — `applyWorkflowBranchGuardFlags`(그래서 `--branch-guard` / `--worktree-auto-*` 플래그 4개가 적용되지 않는다), `writeWorkflowAuditYAML`. 살아 있는 경로는 M4에 명시했고, AC-T-008이 그 방어다.
2. **평범한 `bool`로는 요구를 표현할 수 없다.** 부재(=활성)와 명시적 false를 구별해야 한다.
3. **마법사 질문 추가는 번역 테스트를 깨뜨린다.** `TestWizardQuestionTranslationCompleteness`(`translations_completeness_test.go:89`)는 그렇게 설계돼 있다 — 고칠 테스트가 아니라 따를 관례다.
4. **`DefaultQuestions`에 넣으면 개수 고정 테스트가 깨진다** — 5개(`questions_test.go:101`), 12개(`:190-210`).
5. **statusline 억제 경로가 둘이다** — 기존 `statusline.yaml`의 `backlog: false`와 신규 플래그. OR 판정이어야 하며 한쪽이 다른 쪽을 덮어쓰면 안 된다(`spec.md` §E.3).
6. **형제 SPEC과 파일 9종을 공유한다** — `spec.md` §E.1.
7. **억제 범위가 부분적이다** — 상시 로드 룰과 스킬 목록은 범위 밖. 완료 보고에서 "전부 껐다"고 쓰지 않는다.

## §C Pre-Flight (run-phase 진입 전 확인)

```bash
go test ./internal/config/... ./internal/cli/wizard/... ./internal/statusline/...   # 초록 baseline
grep -rn "workflow.todo\|TodoEnabled" --include='*.go' internal/ | grep -v _test    # 0건이어야 함
grep -n 'backlog' .moai/config/sections/statusline.yaml                             # 부재 확인(키-부재-활성 전제)
```

형제 SPEC이 이미 착지했는지 확인 — 착지했다면 M4 이후 마법사 테스트를 병합 트리에서 돌린다(AC-T-011).

## §D Constraints (Hard)

- `*bool` 필수. 판독 헬퍼는 `readMCPToolEnablement` 형태.
- 상시 로드 룰·스킬 목록·슬래시 스텁은 건드리지 않는다.
- 충족 불가능한 문구의 AC를 쓰지 않는다(`spec.md` §A.1 [HARD]).
- 형제 SPEC 공유 파일은 **다른 항목만 추가**한다.
- 로컬 검증은 패키지 스코프로만. `go test ./...` 금지(CLAUDE.local.md §4).
- 건드린 모든 패키지에 `GOOS=windows go vet`.
- 신규 테스트는 `t.TempDir()`.

## §E Self-Verification (설계 결정)

### 결정 T1 — 키의 집은 `workflow.yaml`, 전용 섹션 금지

`workflow`는 seam 쓰기와 웹 렌더가 동시에 되는 4개 섹션 중 하나이고, 이미 `auto_clear.enabled`·`branch_guard.enabled` 라는 형제 `*.enabled` 게이트를 담고 있다. 전용 `todo.yaml`은 `sectionRootKeys` / `sectionRoutes` / `SectionID` 상수 / `SchemaSectionIDs()` / `consoleTabs()` / `schemaSectionMetas()` 여섯 곳 등록 + i18n을 요구한다 — bool 하나에 과한 비용이다.

### 결정 T2 — `*bool` + 기본 ON 판독

`readMCPToolEnablement`(`internal/cli/mcp_server.go:411-437`)의 형태를 그대로 쓴다: 맵을 `true`로 seed한 뒤 `Enabled *bool`을 가진 익명 구조체로 언마셜, 포인터가 "부재"와 "명시적 false"를 가른다. 그 파일의 `@MX:NOTE`(`:410`)가 fail-OPEN 선택을 의도적으로 기록해 뒀으므로 태도까지 계승한다.

### 결정 T3 — CLI 등록은 유지 (REQ-3)

플래그가 끄는 것은 **안내**이지 기능이 아니다. 명령을 숨기면 foreman 스킬의 `allowed-tools`와 큐 파일은 남는데 진입점만 사라져 진단이 어려워진다. 카드가 결정하지 않은 사항이므로 REQ로 명문화해 검증 가능하게 만든다.

**기각한 대안**: 플래그 false 시 `newTodoCmd()` 등록 생략. 기존 사용자가 큐를 가진 채 명령을 잃는다.

### 결정 T4 — statusline은 OR 판정

`isSegmentEnabled(SegmentBacklog)`와 신규 플래그를 AND로 묶어 "둘 다 켜져야 표시"로 만든다(= 어느 쪽이든 끄면 꺼진다). 데이터 수집 단계에서 차단하지 않고 렌더 판정에 합류시키는 이유: 수집 단계 차단은 `Backlog.Available == false` 경로와 섞여 진단이 어려워진다.

## §F Milestones (되돌리기 어려움 순)

### M1 — 설정 데이터 모델 (형태가 바뀔 확률 최대)

**파일**:
- `internal/config/types.go` — `WorkflowConfig`에 `Todo WorkflowTodoConfig \`yaml:"todo"\``(`Worktree` 이웃, `:365` 부근) + `WorkflowTodoConfig{ Enabled *bool \`yaml:"enabled"\` }`.
- `internal/config/defaults.go` — `NewDefaultWorkflowConfig()`에 `Todo`(nil = 활성).
- 접근자 — `func (c *Config) TodoEnabled() bool`(nil → true).

**Exit**: `go test ./internal/config/...` 초록. AC-T-001.

### M2 — 런타임 표면 억제 2종 (관측 가능한 동작 변경)

**파일**:
- `internal/hook/session_start_kanban.go:180` — 백로그 요약 추가 앞에 `TodoEnabled()` 가드.
- `internal/statusline/renderer.go:188` — 결정 T4의 OR 판정.
- 대응 테스트: 억제 케이스 + **대조 케이스**(플래그 부재 시 표시됨).

**Exit**: `go test ./internal/hook/... ./internal/statusline/...` 초록. AC-T-002, AC-T-003.

### M3 — 스킬 라우팅 + CLI 등록 유지 확인

**파일**:
- `.claude/skills/moai/SKILL.md` + `internal/template/templates/.claude/skills/moai/SKILL.md` — 라우팅 단계에서 플래그가 `false`면 `workflows/todo.md`로 라우팅하지 않는다는 조건 명시. **스킬 목록 메타데이터(`:6`, `:81`, `:105`)는 손대지 않는다** — 범위 밖(§B).
- `internal/cli/todo.go:512` — **변경 없음**. REQ-3의 결정을 테스트로 고정한다.

**Exit**: 스킬 두 사본에 조건 문장 존재(grep). `go test ./internal/cli/ -run 'TestTodoCommandRegisteredRegardlessOfFlag'` 초록. AC-T-004, AC-T-005.

### M4 — 마법사 질문 1개 (살아 있는 경로만)

`init_autonomy_wizard.go`를 따르지 않는다(§B-1).

**파일**:
- `internal/cli/wizard/questions.go` — `Page3Questions`의 "Quality & Workflow" 그룹(`:365-374` `worktree_auto_create` 이웃)에 `todo_enabled`(`Default: "true"`).
- `internal/cli/wizard/types.go` — `WizardResult` 필드 1개.
- `internal/cli/wizard/wizard.go:459-468` — `saveBoolAnswer` case 1개.
- `internal/cli/wizard/translations.go` — ko/ja/zh 3블록에 title+description.
- `internal/cli/init.go:185-224` — `applyWizardPage3ToOpts` 대입 1줄.
- `internal/core/project/initializer_expansion.go:30` — `yamlpatch.PatchFile`로 `workflow.yaml`에 기록(주석 보존).
- 테스트: `wizard/worktree_test.go`(`:8`, `:29`, `:47`) 3종 세트 복제.

**Exit**: `go test ./internal/cli/wizard/... ./internal/cli/... ./internal/core/project/...` 초록. AC-T-006~008.

### M5 — 웹 스키마 + i18n (기계적)

**파일**:
- `internal/settings/schema_sections.go` — `s(SectionWorkflow, "workflow", TypeBool, "workflow", "todo", "enabled")` (`:334` 이웃).
- `internal/web/assets/i18n.js` — 4로케일 × (title, desc).
- 필요 시 `internal/web/schemaform.go:156` 이웃에 탭 배치 술어 1줄(기본은 workflow 탭).

**Exit**: `go test ./internal/settings/... ./internal/web/...` 초록. AC-T-009.

### M6 — 템플릿 결정 + 인벤토리 + 빌드

**기본안: 템플릿에 블록을 싣지 않는다**(`branch_guard` 선례, `schema_sections.go:330-333` 주석이 같은 상황을 문서화). 싣기로 하면:
- `internal/template/templates/.moai/config/sections/workflow.yaml` — `todo: enabled: true` + 중립 주석.
- `internal/config/testdata/shipped_key_inventory.yaml` — 항목 등록(미등록 시 `TestShippedConfigKeysHaveReaders` 실패).

어느 쪽이든 `make build`.

**Exit**: `go test ./internal/config/... ./internal/template/...` 초록, `make build` exit 0. AC-T-010.

### M7 — 검증 스윕 + 커밋

```bash
go test ./internal/config/... ./internal/hook/... ./internal/statusline/... ./internal/cli/... ./internal/cli/wizard/... ./internal/core/project/... ./internal/settings/... ./internal/web/... ./internal/template/...
GOOS=windows go vet ./internal/config/... ./internal/hook/... ./internal/statusline/... ./internal/cli/... ./internal/cli/wizard/... ./internal/core/project/... ./internal/settings/...
golangci-lint run --timeout=2m
make build
```

형제 SPEC이 이미 착지했다면 마법사 개수·번역 테스트를 병합 트리에서 다시 돌린다(AC-T-011).

커밋: 마일스톤 단위 Conventional Commits, footer에 SPEC ID.

## §G Anti-Patterns (피할 것)

- **AP-1**: `init_autonomy_wizard.go` 등 사장 writer 3건을 배선 선례로 따르기.
- **AP-2**: `workflow.todo.enabled`를 평범한 `bool`로 선언.
- **AP-3**: `todo.yaml` 섹션 신설 — bool 하나에 등록 지점 6곳.
- **AP-4**: 신규 질문을 `DefaultQuestions`에 넣기.
- **AP-5**: 영어 단독 번역 후 완전성 테스트 실패를 "낡은 테스트"로 읽기.
- **AP-6**: 상시 로드 룰(`kanban-dispatch.md`)이나 스킬 목록 메타데이터를 건드려 억제 시도 — 범위 밖이며 [HARD] 조항을 훼손한다.
- **AP-7**: "사용 안 함이면 todo 안내가 전혀 뜨지 않는다"는 AC를 쓰기 — 충족 불가능.
- **AP-8**: 완료 보고에서 "todo 안내를 전부 껐다"고 서술하기.
- **AP-9**: statusline 억제를 데이터 수집 단계에서 차단해 `Available == false` 경로와 섞기.
- **AP-10**: 플래그 false 시 CLI 명령 등록을 생략 — REQ-3 위반.
- **AP-11**: 형제 SPEC 공유 파일에서 기존 항목 재배치·재서식.
- **AP-12**: 로컬에서 `go test ./...` 실행.

## §H Cross-References

- SPEC: `spec.md` · 수용: `acceptance.md`
- 형제 SPEC: `.moai/specs/SPEC-FEEDBACK-AUTO-SUBMIT-001/`
- 근거 렌즈: `.moai/reports/t170/lens-web-todo.md`, `lens-init.md`
- CLAUDE.local.md §2(Template-First), §2.1(중립성), §4(검증 규율), §6(테스트 격리)

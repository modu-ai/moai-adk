# Tasks: SPEC-V3R2-EXT-001 — Typed Memory Taxonomy

- SPEC: SPEC-V3R2-EXT-001
- Plan ref: `.moai/specs/SPEC-V3R2-EXT-001/plan.md`
- Methodology: TDD (RED → GREEN → REFACTOR)
- Parallelism: T1.* 제외 직렬화 금지. T2, T3, T4는 internal/hook 동일 파일 편집이 있어 주의 필요.

---

## 1. Task List

### T0. Precheck (block everything else)

- Owner role: analyst
- Files: read-only
- Scope:
  - [ ] **OPEN-3 strategy confirm**: `ls internal/hook/memo/` 결과가 plan.md §3.1에 명시된 6개 파일(`priority.go`, `priority_test.go`, `reader.go`, `reader_test.go`, `writer.go`, `writer_test.go`)과 일치하는지 재확인. 불일치 시 plan.md §3.1 Consequences 갱신 후 진행.
  - [ ] **Strategy B (adapted) 적용 확인**: 신규 코드는 `internal/hook/memo/taxonomy/` 서브패키지에 배치한다는 설계 재확인 (plan.md §3.1 참조). 기존 `memo` 패키지 파일은 비수정.
  - [ ] **OPEN-2 static-keyword v1 confirm**: plan.md §8 OPEN-2 RESOLVED 섹션에 명시된 category→keyword 매핑 테이블을 T4 구현의 기준으로 사용. LLM 분류는 v2 범위 외.
  - [ ] **200-line 재확인 (spec.md §4)**: Claude Code 공식 문서에서 MEMORY.md/context-loader의 truncation 동작을 확인. 문서화된 값이 200이 아닐 경우 `.moai/config/sections/workflow.yaml` `memory.index_line_cap` 기본값과 spec.md §4를 동기화.
  - [ ] `make build && make install` 바이너리 싱크
  - [ ] `go test ./internal/hook/... -race` 현재 green 확인
- DoD:
  - OPEN-3 strategy 재확인 완료 (서브패키지 경로 확정: `internal/hook/memo/taxonomy/`)
  - OPEN-2 detection method 확정 (static-keyword v1)
  - 200-line 수치 공식 문서 근거 기록 또는 대체 값 합의
  - 빌드/테스트 baseline green

### T1. Rule documentation (Phase 1)

- Owner role: implementer
- Files owned:
  - `.claude/rules/moai/workflow/moai-memory.md`
  - `internal/template/templates/.claude/rules/moai/workflow/moai-memory.md`
- Scope:
  - [ ] 4-type 가이드 섹션 (user / feedback / project / reference) 추가
  - [ ] `MEMORY.md` 200줄 cap + truncation 설명 (REQ-003)
  - [ ] staleness caveat 해석 가이드 (REQ-006 결과물 해석)
  - [ ] 타입별 body structure 예시 2–3개 (REQ-010, 011)
  - [ ] 제외 카테고리 목록 (REQ-015)
- Dependencies: T0
- DoD:
  - REQ-005, 009, 010, 011, 012, 015 의 문서화 요건 충족
  - live 파일과 template 파일 diff 없음 (단, template-only placeholder 제외)
  - `make build` 후 embedded 파일 변경 반영
  - 이모지/XML 태그 없음 (coding-standards 준수)

### T2. Memo taxonomy package — taxonomy core (Phase 2)

- Owner role: tester + implementer (TDD)
- Package: `internal/hook/memo/taxonomy` (신규 서브패키지 — OPEN-3 RESOLVED, plan.md §3.1)
- Files owned:
  - `internal/hook/memo/taxonomy/taxonomy.go`
  - `internal/hook/memo/taxonomy/taxonomy_test.go`
  - `internal/hook/memo/taxonomy/fixtures/*.md` (신규 11개: valid 4종 + missing_type/name/description + unknown_type + overflow_index + excluded_claude_md + feedback_missing_why)
- Scope:
  - [ ] RED: `TestValidateType_*`, `TestParseFile_*`, `TestParseFile_MissingName`, `TestParseFile_MissingDescription`, `TestParseFile_MissingBoth` (테이블 드리븐)
  - [ ] GREEN: `MemoryType` const, `ValidateType`, `ParseFile` (Frontmatter name/description 키 추출 포함)
  - [ ] REFACTOR: frontmatter 파서 util 추출 검토
- Dependencies: T0
- Parallel: T1과 병렬 가능 (파일 겹침 없음, 서브패키지 신설이라 기존 `memo` 패키지 무영향)
- DoD:
  - REQ-001, 002, 004 구현 (AC-01, AC-01b, AC-02, AC-09)
  - `go test ./internal/hook/memo/taxonomy/... -race` green
  - 기존 `internal/hook/memo/` 회귀 없음 (`go test ./internal/hook/memo/... -race`)
  - 90% coverage 달성
  - 모든 테스트 `t.TempDir()` 사용, `t.Setenv("HOME", ...)` 없음

### T3. Memo taxonomy package — staleness (Phase 3 prep)

- Owner role: tester + implementer
- Package: `internal/hook/memo/taxonomy`
- Files owned:
  - `internal/hook/memo/taxonomy/staleness.go`
  - `internal/hook/memo/taxonomy/staleness_test.go`
- Scope:
  - [ ] RED: `TestDetectStale_Boundary` (23h/24h/25h), `TestAggregateWarning_Counts` (0/1/9/10/11)
  - [ ] GREEN: `DetectStale(dir, hours, now)`, `AggregateWarning([]StaleReport)`
  - [ ] REFACTOR: wrap 포맷 상수 추출 (하드코딩 방지)
- Dependencies: T2
- DoD:
  - REQ-006, 017 (aggregation) 구현
  - `now time.Time` 주입형 서명 (flake 방지)
  - 커버리지 90%+

### T4. Memo taxonomy package — audit (Phase 4 prep)

- Owner role: tester + implementer
- Package: `internal/hook/memo/taxonomy`
- Files owned:
  - `internal/hook/memo/taxonomy/audit.go`
  - `internal/hook/memo/taxonomy/audit_test.go`
- Scope:
  - [ ] RED: `TestAuditFile_MissingType`, `TestAuditFile_MissingFrontmatterKeys` (AC-01b), `TestAuditFile_FeedbackBodyMarkers` (AC-04a), `TestAuditFile_ExcludedCategory`, `TestAuditIndex_Overflow`, `TestAuditDuplicates`
  - [ ] GREEN: `AuditFile`, `AuditIndex`, `AuditDuplicates` (+ 신규 `AuditFeedbackBody` 또는 `AuditFile` 내부 dispatch)
  - [ ] **OPEN-2 static-keyword v1 구현**: plan.md §8 OPEN-2 RESOLVED 카테고리→키워드 매핑 테이블을 코드 상수로 구현 (`internal/config/defaults.go` 또는 `taxonomy` 전역 상수). LLM 분류 경로는 추가하지 않음.
  - [ ] REFACTOR: AuditCode 상수, 판정 키워드 목록을 단일 원천화
- Dependencies: T2 (ParseFile 재사용)
- Parallel: T3과 병렬 가능 (별도 파일)
- DoD:
  - REQ-007, 008, 015, 016 구현 (AC-03, AC-04a, AC-10, AC-11)
  - OPEN-2 (제외 카테고리 판정) **static-keyword v1 확정** — PR 설명에 카테고리 매핑 테이블 인용
  - False-positive 수집 로그 경로 명시 (Phase 5 모니터링 포인트)
  - 커버리지 90%+

### T5. SessionStart hook integration (Phase 3 live)

- Owner role: implementer
- Files owned:
  - `internal/hook/session_start.go`
  - `internal/hook/session_start_test.go`
  - `internal/hook/session_start_extra_coverage_test.go`
- Scope:
  - [ ] 메모리 로드 경로에서 `taxonomy.DetectStale` 호출 (import: `…/internal/hook/memo/taxonomy`)
  - [ ] aggregation threshold(10) 초과 시 단일 warning 주입
  - [ ] 각 stale 파일 content를 `<system-reminder>` wrap으로 교체 후 agent context 전달
  - [ ] env flag `MOAI_MEMORY_AUDIT=0`로 비활성화 경로 (rollback 대비)
- Dependencies: T3
- DoD:
  - REQ-006, 017 통합 동작
  - 기존 SessionStart 테스트 모두 green (회귀 없음)
  - 새 테스트: stale wrap 포함, aggregation 경로
  - `make build && make install && moai version` 커밋 hash 확인

### T6. PostToolUse hook integration (Phase 4 live)

- Owner role: implementer
- Files owned:
  - `internal/hook/post_tool.go`
  - `internal/hook/post_tool_test.go`
- Scope:
  - [ ] Write/Edit tool event가 memory 파일 경로인 경우에만 audit 트리거
  - [ ] 각 AuditFinding을 stderr non-blocking warning으로 출력
  - [ ] env flag `MOAI_MEMORY_AUDIT=0` 존중
- Dependencies: T4
- Parallel: T5와 파일 겹침 없어 병렬 가능. 단, T3/T4의 패키지 API 안정화 이후 착수.
- DoD:
  - REQ-007, 008, 015, 016 훅 경로 동작
  - 기존 post_tool 테스트 회귀 없음
  - exit code 0 유지 (non-blocking)

### T7. Config schema + defaults (Phase 5)

- Owner role: implementer
- Files owned:
  - `.moai/config/sections/workflow.yaml`
  - `internal/template/templates/.moai/config/sections/workflow.yaml`
  - `internal/config/defaults.go` (또는 해당 loader 구조체)
- Scope:
  - [ ] `memory.staleness_hours: 24`, `memory.index_line_cap: 200`, `memory.stale_aggregate_threshold: 10` 키 등록
  - [ ] loader 구조체 + 기본값
  - [ ] 기본값 단일 원천(하드코딩 금지 — CLAUDE.local.md §14)
- Dependencies: T5, T6 (API 확정 후 설정 연결)
- DoD:
  - `go test ./internal/config/...` green
  - 템플릿/live 두 yaml 완전 동일
  - OPEN-1 키 이름 확정 반영

### T8. Verification sweep (Phase 6)

- Owner role: reviewer
- Files: read-only + 테스트 실행만
- Scope:
  - [ ] `go test ./... -race -count=1`
  - [ ] `go vet ./...`
  - [ ] `golangci-lint run`
  - [ ] `make build && make install`
  - [ ] manual smoke: `moai hook session-start` stdin fixture로 aggregated 경로 확인
  - [ ] `audit_sweep_patterns.md` Pattern A — YAML frontmatter CSV/array 규칙 위반 없는지 grep
- Dependencies: T1–T7 전부
- DoD:
  - 모든 명령 zero error
  - 커버리지 리포트 `internal/hook/memo/taxonomy` 90%+ (기존 `internal/hook/memo` session memo 패키지는 회귀 없음만 확인)
  - TRUST 5 리뷰 체크리스트 통과 (acceptance.md DoD 참조)

---

## 2. Dependency Graph

```
T0
├──► T1 (docs)        ──────────────────────┐
├──► T2 (taxonomy)                          │
│     ├──► T3 (staleness) ──► T5 (session_start)
│     └──► T4 (audit)     ──► T6 (post_tool)
└──► T7 (config defaults) depends on T5+T6 API
                                           │
T1..T7 all ──► T8 (verification)
```

- Edges는 API/파일 의존만 명시. 런타임 의존(훅 flag 등)은 T8에서 검증.

---

## 3. Parallelizable Groups

| Group | Tasks | 조건 |
|-------|-------|------|
| G1 | T1, T2 | T0 완료 후 병렬 시작 가능 (파일 겹침 없음) |
| G2 | T3, T4 | T2 완료 후 병렬 가능 |
| G3 | T5, T6 | T3/T4 각각 완료 후 병렬 가능 (다른 훅 파일) |
| G4 | T7 | T5+T6 API 안정화 후 진행 |
| G5 | T8 | 전 Task 완료 후 단독 실행 |

실질 임계 경로: T0 → T2 → T3 → T5 → T7 → T8 (6 단계).

---

## 4. File Ownership Map (team 모드용)

| Path pattern | Owner |
|--------------|-------|
| `.claude/rules/moai/workflow/moai-memory.md` | T1 |
| `internal/template/templates/.claude/rules/moai/workflow/moai-memory.md` | T1 |
| `internal/hook/memo/taxonomy/taxonomy*.go`, `internal/hook/memo/taxonomy/fixtures/*.md` | T2 |
| `internal/hook/memo/taxonomy/staleness*.go` | T3 |
| `internal/hook/memo/taxonomy/audit*.go` | T4 |
| `internal/hook/session_start*.go` | T5 |
| `internal/hook/post_tool*.go` | T6 |
| `.moai/config/sections/workflow.yaml`, template mirror, `internal/config/defaults.go` | T7 |
| NEVER (비수정) | `internal/hook/memo/{priority,reader,writer}*.go` — 기존 session memo 패키지 |

team 모드에서 T5/T6 분리 병렬 시 worktree isolation 필수 (.claude/rules/moai/workflow/worktree-integration.md §Implementation Agents).

---

## 5. Global DoD (모든 Task 공통)

- [ ] Conventional Commits 포맷 (`feat`/`fix`/`docs`/`test`/`chore`)
- [ ] 이모지 없음, XML 태그 user-facing 없음 (coding-standards)
- [ ] Go code, godoc, comment: English (language.yaml `code_comments: ko`이나 Go 규약 우선 — 본 프로젝트 CLAUDE.local.md §3 "All code, comments, godoc in English")
- [ ] 커밋 메시지: 한국어 (language.yaml `git_commit_messages: ko`)
- [ ] 테스트 격리: `t.TempDir()`, OTEL env/`t.Setenv("HOME", ...)` 금지
- [ ] TRUST 5 5개 항목 녹색
- [ ] `make build && make install` 수행 후 바이너리 stale 검증 완료 (MEMORY.md Hard Constraint)

---

End of Tasks.

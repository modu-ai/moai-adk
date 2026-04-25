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
  - [ ] `ls internal/hook/memo/` 구조 확인 (plan §Open Questions OPEN-3)
  - [ ] `make build && make install` 바이너리 싱크
  - [ ] `go test ./internal/hook/... -race` 현재 green 확인
- DoD: OPEN-3 답 확정 + 빌드/테스트 baseline green.

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

### T2. Memo package — taxonomy core (Phase 2)

- Owner role: tester + implementer (TDD)
- Files owned:
  - `internal/hook/memo/taxonomy.go`
  - `internal/hook/memo/taxonomy_test.go`
  - `internal/hook/memo/fixtures/*.md` (신규 8개)
- Scope:
  - [ ] RED: `TestValidateType_*`, `TestParseFile_*` (테이블 드리븐)
  - [ ] GREEN: `MemoryType` const, `ValidateType`, `ParseFile`
  - [ ] REFACTOR: frontmatter 파서 util 추출 검토
- Dependencies: T0
- Parallel: T1과 병렬 가능 (파일 겹침 없음)
- DoD:
  - REQ-001, 002, 004 구현
  - `go test ./internal/hook/memo/... -race` green
  - 90% coverage 달성
  - 모든 테스트 `t.TempDir()` 사용, `t.Setenv("HOME", ...)` 없음

### T3. Memo package — staleness (Phase 3 prep)

- Owner role: tester + implementer
- Files owned:
  - `internal/hook/memo/staleness.go`
  - `internal/hook/memo/staleness_test.go`
- Scope:
  - [ ] RED: `TestDetectStale_Boundary` (23h/24h/25h), `TestAggregateWarning_Counts` (0/1/9/10/11)
  - [ ] GREEN: `DetectStale(dir, hours, now)`, `AggregateWarning([]StaleReport)`
  - [ ] REFACTOR: wrap 포맷 상수 추출 (하드코딩 방지)
- Dependencies: T2
- DoD:
  - REQ-006, 017 (aggregation) 구현
  - `now time.Time` 주입형 서명 (flake 방지)
  - 커버리지 90%+

### T4. Memo package — audit (Phase 4 prep)

- Owner role: tester + implementer
- Files owned:
  - `internal/hook/memo/audit.go`
  - `internal/hook/memo/audit_test.go`
- Scope:
  - [ ] RED: `TestAuditFile_MissingType`, `TestAuditFile_ExcludedCategory`, `TestAuditIndex_Overflow`, `TestAuditDuplicates`
  - [ ] GREEN: `AuditFile`, `AuditIndex`, `AuditDuplicates`
  - [ ] REFACTOR: AuditCode 상수, 판정 키워드 목록을 `internal/config/defaults.go` 또는 패키지 전역 상수로 단일화
- Dependencies: T2 (ParseFile 재사용)
- Parallel: T3과 병렬 가능 (별도 파일)
- DoD:
  - REQ-007, 008, 015, 016 구현
  - OPEN-2(제외 카테고리 판정) PR 설명에 명시
  - 커버리지 90%+

### T5. SessionStart hook integration (Phase 3 live)

- Owner role: implementer
- Files owned:
  - `internal/hook/session_start.go`
  - `internal/hook/session_start_test.go`
  - `internal/hook/session_start_extra_coverage_test.go`
- Scope:
  - [ ] 메모리 로드 경로에서 `memo.DetectStale` 호출
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
  - 커버리지 리포트 `internal/hook/memo` 90%+
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
| `internal/hook/memo/taxonomy*.go`, `fixtures/*.md` | T2 |
| `internal/hook/memo/staleness*.go` | T3 |
| `internal/hook/memo/audit*.go` | T4 |
| `internal/hook/session_start*.go` | T5 |
| `internal/hook/post_tool*.go` | T6 |
| `.moai/config/sections/workflow.yaml`, template mirror, `internal/config/defaults.go` | T7 |

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

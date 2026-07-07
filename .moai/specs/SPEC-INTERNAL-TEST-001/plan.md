---
id: SPEC-INTERNAL-TEST-001
version: "0.1.0"
status: in-progress
created: 2026-07-08
updated: 2026-07-08
---

# SPEC-INTERNAL-TEST-001 — 구현 계획 (plan)

## §A Context

`go test ./internal/...` 베이스라인 복구. 5개 검증된 발견사항(F1-F5, spec.md §A)을 6개 GEARS 요구사항(REQ-TEST-001..006)으로 변환 완료. Tier S — 수정 표면은 테스트 파일 + doc 2건으로 한정되며 production 코드 무접촉(C-1..C-3).

## §B Known Issues / 사전 확인 사항

1. **메모리-현재 상태 충돌**: `project_ac_css_001_rescan_debt.md`(2026-07-08 재실측)는 F1(`TestRunHookEvent_ReadInputError`)·F2(agentlint doc-tests)를 PASS로 기록했으나, 본 SPEC 저작 시점 HEAD의 read-only 재검증에서 두 결함 모두 구조적으로 존재함을 확인(coverage_test.go:75-78 nil deref 구조, agent_lint_test.go:1077-1083 stale 목록). rescan 이후 doc/template 편집(예: 2026-07-07 agent-authoring.md 갱신)으로 재유입된 것으로 추정 — **현재 관측 우선** (verification-claim-integrity §1.1).
2. **F3 oracle 정밀성**: `drift_doctrine_test.go:47-48` 정규식은 `(combined|abbreviated)[^\n]*scope`로 **같은 줄** 배치를 요구하고, prohibition token과의 거리는 `(?s).{0,400}?` (개행 포함 400자, 양방향). 문구 삽입 시 이 두 조건을 모두 만족해야 한다.
3. **F2 테스트의 추가 단언**: `TestAuthoringDocHasEffortMatrix`는 `expectedAgents` 외에도 `## Effort-Level Calibration Matrix` 헤딩과 effort 값(`xhigh`/`high`/`medium`) 존재를 단언한다(agent_lint_test.go:1072, 1097-1102). doc 정합 시 헤딩·effort 값 보존 필수.
4. **병렬 세션 위생**: working tree에 무관 변경(system.yaml 등) 존재. run-phase 커밋은 specific-path로 한정할 것 (공유 체크아웃 동시 커밋 레이스 교훈).

## §C 기술 접근 (Technical Approach)

- **F1**: 테스트를 fail-open 계약으로 갱신 — `err == nil` 단언 + default `HookOutput`이 기록되었는지 확인(가능하면 stdout capture 또는 `writeHookOutput` 경로 검증). production `hook.go` 무접촉.
- **F2**: 양방향 정합 — (a) `expectedAgents`를 현행 retained 카탈로그(7 MoAI-custom: manager-spec/manager-develop/manager-docs/manager-git/plan-auditor/sync-auditor/builder-harness; Explore는 MoAI 파일 없음 — matrix 실제 내용 기준으로 결정)로 교체 + archived 명칭 전부 제거, (b) `agent-authoring.md` effort matrix가 그 목록을 실제로 포함하는지 확인·보정. doc 편집 발생 시 **Template-First**: 템플릿 원본 → `make build` → local mirror, 양 트리 diff 무결 확인.
- **F3**: `lifecycle-sync-gate.md`의 "Status Transition Ownership Matrix Cross-Reference" 절에 amendment 블록 추가 — `spec-frontmatter-schema.md` § Close-subject full-ID mandate의 canonical prohibition prose를 미러(같은 줄 `combined/abbreviated ... scope` + prohibition verb ≤400자). **템플릿 미러 생성 금지**(dev-only 파일).
- **F4**: 기존 table 구조를 살려 17개 skip body 실체화. 검증 대상: crash-recovery(`detectInFlightState`/`cleanupInFlightState`), idempotency(동일 마이그레이션 재실행 무해성), atomic version-file update, log append. 전부 `t.TempDir()` 격리. 커버리지 ≥ 85%.
- **F5**: `canary_test.go` + `contradiction_test.go` 신규 — table-driven으로 exported 진입점(Evaluate, Scan) 중심 + 내부 helper 경로 도달. 패키지 커버리지 ≥ 85%.

## §F Milestones (priority-based — 시간 추정 없음)

| # | Priority | 대상 | 내용 | 완료 판정 |
|---|----------|------|------|-----------|
| M1 | P0 | F1 (REQ-TEST-002) | `coverage_test.go` fail-open 계약 갱신 — **최우선: internal/cli 커버리지 측정 unblock** | AC-TEST-002a/b |
| M2 | P2 | F3 (REQ-TEST-004) | `lifecycle-sync-gate.md` amendment 삽입 (doc-only, no-mirror) — quick win | AC-TEST-004a/b |
| M3 | P2 | F2 (REQ-TEST-003) | `expectedAgents` 정합 + `agent-authoring.md` matrix 보정 (**Template-First 필수**) | AC-TEST-003a/b/c |
| M4 | P1 | F4 (REQ-TEST-005) | migration 17개 skip → 실 테스트, coverage ≥ 85% | AC-TEST-005a/b |
| M5 | P2 | F5 (REQ-TEST-006) | constitution canary/contradiction 테스트 신규, coverage ≥ 85% | AC-TEST-006a/b |
| M6 | — | headline (REQ-TEST-001) | `go test ./internal/... -count=1` 전체 검증 + 증거 기록 | AC-TEST-001 |

순서 근거: M1이 측정 자체를 unblock하므로 선행 필수. M2/M3는 결정론적 doc 게이트(소규모)로 조기 처리. M4/M5는 테스트 저작 비중이 큰 작업. M6은 최종 게이트.

## §G Risks / Anti-Patterns

| 위험 | 완화 |
|------|------|
| M3에서 local `.claude/`만 편집 → Template-First 위반, parity CI FAIL | 템플릿 원본 선편집 → `make build` → mirror; AC-TEST-003c diff 게이트 |
| M2에서 템플릿 미러 생성 → neutrality(§25) 위반 (파일에 내부 SPEC ID 포함) | C-2 shall-not + AC-TEST-004b (템플릿 부재 확인) |
| M6 headline이 statusline pre-existing flake로 비결정적 FAIL | acceptance.md §D.3 재실행 프로토콜 — 해당 단일 테스트 한정 1회 재실행 판정 |
| M4/M5 중 production 결함 발견 | C-3: silent fix 금지, blocker report 반환 |
| M4 테스트가 실제 파일시스템/훅 경로 접촉 | C-4: `t.TempDir()` 전면 격리 |
| constitution 41.3% → 85% 도달 미달 | Scan/Evaluate 진입점 table 케이스 확장으로 helper 경로 커버; 도달 불가 시 실측 수치 + 사유를 blocker report로 |
| 병렬 세션 커밋 레이스 | specific-path commit + pre-spawn `git fetch` divergence 확인 |

## §H Cross-References

- spec.md §B (REQ-TEST-001..006), §C (C-1..C-5), Out of Scope
- acceptance.md §D (AC-TEST-001..006b), §D.3 (flake 재실행 프로토콜)
- CLAUDE.local.md §2 (Template-First / Local-Only), §6 (Test Isolation / Coverage)
- `.claude/rules/moai/development/spec-frontmatter-schema.md` § Close-subject full-ID mandate (F3 미러 원문)

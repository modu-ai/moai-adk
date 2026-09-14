---
id: SPEC-INIT-UPDATE-CONSISTENCY-001
title: "implementation plan — init/update 정합성·문서 정리"
created: 2026-09-13
---

# plan.md — SPEC-INIT-UPDATE-CONSISTENCY-001

## §A Context

카드 t588 — init/update 전수 조사(2026-09-09) C6 축. 워크트리 `.claude/worktrees/t588`(HEAD `fac132d38`)에서 전 축 재측정 완료: 수정 6축(F8·F9·F12·F13-요약·F14-parity·F15-안내), 기록 2축(F13-분산 본체, F17), 소멸 3건(F16, F14-탭수, F14-model_policy). 상세 근거는 spec.md §1 처분 요약 + §3 소멸 판정.

## §B Known Issues

- 감사 보고서의 file:line 앵커 일부가 이 트리에서 이동했다: F8 writer 는 `initializer_expansion.go:52-70`(구 :154-181 아님), F9 는 `defaults.go:882`(구 :793 — 그 자리는 이제 git-strategy ModeProfile), F12 소비자는 `update_tux.go:165`(t694 재작업), F14 handlers 는 `internal/web/handlers.go`. run-phase 는 본 plan의 앵커를 쓴다.
- `update_namespace_protect.go` 패키지 문서가 F13 분산을 "No consolidation" 으로 기록 — 분산 본체는 재오픈하지 않는다.

## §C Pre-flight

- `git rev-parse --short HEAD` 로 fac132d38 이상(develop 동기) 확인 — t694 표시 경로 위에 작업함.
- `make build` 선행 필요 여부: project.yaml.tmpl 수정 시 필수. `agents-emit-check`/`commands-emit-check` 영향 없음(템플릿 yaml).
- 계측 baseline: `go test ./internal/cli/... ./internal/config/... ./internal/core/project/... ./internal/web/...` 시작 상태 기록.

## §D Constraints

- 표시 렌더(update_tux.go t694 영역) 무접촉 — REQ-ICU-004 의 renderUpdateOutcome 시그니처 확장만 허용.
- `.sh`/`.sh.tmpl` 쌍 삭제 금지 — 계수 dedupe 만.
- 질문 집합 무변경 (Out of Scope).
- 개발 모드: quality.yaml `constitution.development_mode` 따름(기본 tdd) — 각 축 RED→GREEN.

## §E Self-Verification

- E1: 축별 AC PASS/FAIL 행렬 — acceptance.md §D 대조.
- E2: `GOOS=windows go build ./...` 크로스 빌드(config/template 수정 동반).
- E3: 영향 패키지 커버리지 — baseline 대비 감소 없음.
- E4: 유령 잔존 grep — `project.mode`, `ProjectMode`, `--project-mode` 리더 0건 확인. 플래그 등록 부재도 단정한다 — `grep -rc 'Flags().String("project-mode"' internal/cli/init.go` → 0 (등록 잔존은 리더 grep 이 못 잡는 반쪽 제거 형태를 기계적으로 차단).
- E5: `golangci-lint run` 영향 패키지 clean.

## §F Milestones

변경-가능성 역순 정렬: 데이터 기본값(의미 반전 위험) → 사용자 표면 제거 → 계수·표시 → 테스트·기록.

### M1 (High) — F9 execution_mode 기본값 정렬 [verification-gated, 먼저 판정 확정]

- RED: `TestExecutionModeDefaultMatchesTemplate` 신설 — `NewDefaultWorkflowConfig().ExecutionMode` == 템플릿 `workflow.yaml` 파싱값(현실: "team" vs "auto" 로 실패).
- GREEN: `internal/config/defaults.go:882` `"team"` → `"auto"` + 주석에 정렬 근거(spec REQ-ICU-002 3근거) 기입.
- 영향 확인: `execution_modes_test.go:75` `TestExecutionModeDefaultIsInSet` 은 membership 검사라 무수정 통과 확인.
- 완료 판정: M1 테스트 GREEN + `go test ./internal/config/...` 통과.

### M2 (High) — F8 project.mode 유령 제거

- `internal/core/project/initializer_expansion.go`: `writeProjectModeYAML` 삭제 + `WritePhase1Configs` 호출부 정리.
- `internal/core/project/initializer.go:50`: `ProjectMode` 필드 삭제.
- `internal/cli/init.go`: `--project-mode` 플래그 등록(:91)·할당(:581)·검증 블록(:369-375) 삭제.
- `internal/template/templates/.moai/config/sections/project.yaml.tmpl`: `mode:` 키+주석 삭제 → `make build`.
- 관련 테스트(expansion_test.go 등) 갱신 — "테스트가 false만 재어 미탐지"(F1 전례) 반복 금지: 제거 대상 키의 부재를 assert 하는 방향으로 재작성.
- 완료 판정: E4 grep 0건 + 프로젝트 패키지 테스트 통과.

### M3 (Medium) — F12 페어 이중 계수 제거

- RED: synthetic 템플릿 목록에 `x.sh`+`x.sh.tmpl` 페어 포함 → `managedRedeployed` 2 회관측 실패 테스트.
- GREEN: `internal/cli/update_template_sync.go` 계수 루프에 stripped-target 기집재 skip 추가.
- 완료 판정: 신규 테스트 GREEN + 기존 t40 카운트 테스트 정합.

### M4 (Medium) — F13 요약 완전 표기 + F15 수동 복구 안내 [같은 표시 영역, 함께]

- `internal/cli/update_tux.go` `updateOutcomeDetail`/`renderUpdateOutcome` 확장 — 생성된 백업 뿌리 전부 표기(REQ-ICU-004) + 비-섹션 config 미병합 복원 안내 행(REQ-ICU-006, 백업 경로 포함).
- 호출부(update_template_sync.go)에서 뿌리 3종 존재 여부 수집해 전달.
- 렌더 테스트: 다중 뿌리 존재 시 전부 표기 / 단일 뿌리 시 기존 출력과 동일(regression 가드).
- 완료 판정: 렌더 테스트 GREEN — t694 기존 표시 계약 테스트(`update_noise.go` 등) 무정통과 확인.

### M5 (Medium) — F14 파싱-렌더 parity 가드

- census 산출: `AllFields()` editable ↔ 렌더 표면 대응표(테스트 케이스 형태).
- `internal/web` parity 테스트 신설 — editable FieldDef 전건이 (a) `schemaSectionMetas()`/전용 컴포넌트에 렌더되거나 (b) exempt 목록에 명시.
- census 에 걸린 필드 처리: 렌더 홈 부여 또는 exempt 선언(근거 주석 필수).
- 완료 판정: parity 테스트 GREEN — `tab_layout_test.go` 계약 무정통과.

### M6 (Low) — F17 기록 내구화 + 마무리

- `internal/config/manager.go` `Save` godoc 에 6섹션 범위 명시 + `@MX:DEBT`(+`@MX:CEILING`/`@MX:UPGRADE`) 부착.
- E1-E5 셀프 검증 배치.

## §G Anti-Patterns

- 표시 렌더 영역 "while I'm here" 리팩터 금지(M4 시그니처 확장이 필요 최소선).
- 계수 dedupe 를 ListTemplates 측에서 고치는 것 금지 — ListTemplates 의 stripped-target 목록 자체는 deployer 계약이며 다른 소비자가 있다. 계수 원천만.
- `project.mode` 제거를 "주석만 남기고 키 유지"로 반쪽 처리 금지 — 유령의 정의가 반쯤 남는다.

## §H Cross-References

- 감사 원문: `.moai/reports/init-tui-audit-20260909.md` (primary 체크아웃 로컬 전용)
- 선행/인접: SPEC-INIT-QUIET-WIZARD-001(t583 baseline), SPEC-CLI-TUX-INIT-UPDATE-001, SPEC-V3R6-UPDATE-NAMESPACE-PROTECT-001(F13 기록), SPEC-GITSTRATEGY-SAVE-ISOLATION-001(F17 선례), SPEC-WEB-CONSOLE-REDESIGN-001(F14 스키마 구조)

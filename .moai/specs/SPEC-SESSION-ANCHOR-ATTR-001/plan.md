# plan.md — SPEC-SESSION-ANCHOR-ATTR-001

## §A Context

카드 t1339(및 동일 근원의 t1337)의 통합 수리. 관측된 사건은 모두 거부형이지만, 클래스는 (a) t741이 실측한 무음 변이(경로 없는 명령이 오트리에서 조용히 실행), (b) 처분 가드(`LiveAnchoredSessions`)로 흘러가는 잘못된 레지스트리 cwd, (c) lane-parallel 표준 운용이 곧 재현 조건이라는 점에서 Medium-High다. 런타임 앵커 자체는 저장소 밖이므로, 본 SPEC은 **귀속 가능화(W1) → 재배치 감사·타당성 플래그(W2) → 트레이스 스위치(W3) → 규율 문서(W4) → 심각도 기록(W5)**의 다섯 워크스트림으로 부분 수리한다.

결정 가변성 순(검토 우선순위): W2의 advisory/opt-in-blocking 경계와 감사 행 스키마(§F M2)가 가장 바뀔 가능성이 높은 결정이고, W3 트레이스 행 스키마가 그 다음이다. W4 문서 갱신은 기계적이라 마지막에 둔다.

## §B Known Issues

- [NEEDS CLARIFICATION 없음 — 조사 4종(종합+3 렌즈)이 전부 소진됨. 미증명 3항목(런타임 앵커 keying, t1339 방어 층 귀속, 간헐성 역학)은 단정 대신 W3 계측의 측정 대상으로 명시적으로 이관했다.]
- live 병리 잔존: active-sessions.json의 pid 41489 삼중 session_id, 폐기 t1337 트리 지시 stale 항목 3건 — 본 SPEC은 이들을 수리하는 저장소 정리(sweep)가 아니라 **앞으로의 판독을 귀속 가능하게 하는 계측**을 만든다. stale 항목 정리는 t1369 sweep 카드 소관.

## §C Pre-flight

- [ ] `go build ./...` GREEN (작업 분기 기준)
- [ ] `go test ./internal/hook/... ./internal/session/...` — 기존 GREEN 기록(변경 전 baseline 캡처)
- [ ] `worktree_guard_refusal_test.go` 살생 변이 2건 유지 확인
- [ ] 템플릿 소스 존재 확인: `internal/template/templates/.claude/rules/moai/workflow/worktree-integration-ops.md` (W4는 템플릿 소스 선수정 + `make build` 후 로컬 동기)

## §D Constraints

1. **R1 착지 게이트**: `LiveAnchoredSessions` 단일 소스화 금지, 처분 가드 시맨틱 불변 (REQ-SAA-011).
2. **Seam A 분리 유지**: git-context는 `input.CWD`에서, 감사 로그 디렉터리는 `CLAUDE_PROJECT_DIR` 앵커 (REQ-SAA-012).
3. **상류 문구 핀**: `worktreeGuardAnchor` 불변 (REQ-SAA-010).
4. **훅 사본 제약(t1064)**: 계측기는 실제 실행되는 곳(Go 훅 층 / primary 훅 경로)에만 — worktree-로컬 훅 사본 금지 (REQ-SAA-013).
5. **MOAI_HOME 변이 대상 전용**: `RefuseMutationFromNonCanonicalTree` 불변 (REQ-SAA-013).
6. **fail-open 도크트린**: W2 플래그는 advisory 기본, blocking은 opt-in config(`workflow.anchor_relocation_guard.enabled`, defaults.go 패턴) — 기본 동작 변화 없음.
7. **재배치 두-패스 후보 순서 불변** (REQ-RAR-002/003/004).
8. **env 상수는 envkeys.go 선언** (하드코딩 금지 패턴).
9. **Template-First**: W4 문서는 템플릿 소스를 먼저 고치고 `make build` → 로컬 동기.

## §E Self-Verification

- E1: AC 매트릭스 판정 (acceptance.md) — 각 AC별 실행 증거 인용
- E2: `go build ./...` + `go vet ./internal/...` 출력
- E3: `go test ./internal/hook/... ./internal/session/... ./internal/config/...` 출력 — baseline 대비 GREEN
- E4: `grep -rn "MOAI_ANCHOR_TRACE" internal/ --include="*.go"` — envkeys.go 상수 선언 외 하드코딩 0건
- E5: `golangci-lint run ./internal/hook/... ./internal/session/...` — CI 판 버전
- E6: W1/W2/W3 로그 행의 실측 샘플(테스트 출력) — session_id/cwd 필드 존재 관측
- E7: 템플릿 중립성: `go test ./internal/template/... -run '^TestTemplateNeutralityAudit$'`

## §F Milestones

- **M1 — W1 거부 귀속 가능화 (Priority High)**
  `[MODIFY]` internal/hook/post_tool_failure.go, internal/hook/failure_observer.go
  `WorktreeGuardRefusal` 행에 session_id + 해상 cwd + 거부 인용 트리 경로 추가. 미해상 시 명시적 `unknown` 마커(행 유지). `worktree_guard_refusal_test.go`에 행-필드 단정 추가. 데이터 모델 변경(행 스키마)이므로 최우선 — W2/W3가 같은 로그 표면을 공유.
- **M2 — W2 재배치 감사 + 소유 타당성 플래그 (Priority High)**
  `[MODIFY]` internal/session/anchor.go (RelocateSession 감사 행), internal/hook/cwd_changed_relocate.go (플래그 판정), `[MODIFY]` internal/config/defaults.go (`workflow.anchor_relocation_guard.enabled` 기본 false), envkeys.go 필요 시
  모든 RelocateSession에 감사 행(session, from, to, trigger hook, timestamp). 목표 트리 lock 소유를 REQ-SAA-004 케이스표(spec.md §C.2)로 판정 — lock reason card-id/pid 토큰 파싱, self/other-live/unreadable/registry-only 4케이스. other-live·registry-only는 advisory 플래그 + 진행. opt-in blocking은 플래그 시 refusal 기록. 두-패스 순서·fail-open 불변.
- **M3 — W3 앵커 트레이스 스위치 (Priority High)**
  `[NEW]` 트레이스 경로 (anchor.go / cwd_changed_relocate.go / branch_guard.go의 의사결정점), `[MODIFY]` internal/config/envkeys.go (`MOAI_ANCHOR_TRACE` 상수), branch_guard.go의 앵커 읽기 지점 계측
  게이트 시 세션 귀속 가능 상세 로그(session_id+pid+cwd+timestamp/행). 미게이트 시 출력 0·오버헤드 env 조회 1회. 다음 발생의 재현/측정 계측기 — t1064 클래스의 이 도메인 종결.
- **M4 — W4 회피 규율 문서 (Priority Medium)**
  `[MODIFY]` internal/template/templates/.claude/rules/moai/workflow/worktree-integration-ops.md → `make build` → 로컬 동기
  복구 절차(ExitWorktree keep + 재-EnterWorktree), 런타임 경계 선언, 무음 경로-없는-명령 위험(t741), 정정된 t1337/t1339 기록(재시작 없이 지속). 세션-종료 메모리 갱신은 run-phase 후속으로 주석만.
- **M5 — W5 심각도 기록 확인 + 종합 검증 (Priority Medium)**
  spec.md §A.3 심각도 논거와 plan 반영 최종 일치 확인, §E Self-Verification 전 배치 실행, plan-audit (--deep) 통과. `[EXISTING]` internal/session/anchor_lock.go — 동작 불변 확인만.

## §G Anti-Patterns

- 로그 행 스키마를 확정하지 않은 채 세 워크스트림을 병렬 구현 — M1 스키마가 단일 표면이다.
- "수정이니 blocking을 기본으로" — fail-open 도크트린 위반. opt-in 패턴(branch_guard/agent_stop_guard 선례)을 따른다.
- worktree-로컬 훅 사본에 계측 추가 — t1064 측정상 구조적으로 죽은 코드다.
- 런타임 앵커 내부를 단정하는 문장을 문서에 심기 — 미증명은 unknown으로 기록한다.
- stale 레지스트리 항목을 본 SPEC에서 "청소" — sweep은 t1369 소관. 경계 침범 금지.

## §H Cross-References

- spec.md §A (측정 확립 사실 8건 + 심각도 논거), acceptance.md (AC-001..AC-010), research.md (조사 종합 축어 + 렌즈 경로)
- SPEC-SESSION-REGISTRY-READ-ANCHOR-001 · SPEC-WORKTREE-BRANCH-GUARD-DISCRIM-001 · SPEC-WORKTREE-REAPER-001 · SPEC-WORKTREE-BRANCH-GUARD-001

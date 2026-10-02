---
id: SPEC-FACTORY-MANAGED-SESSION-001
title: "plan.md — 구현 계획"
version: "0.1.0"
created: 2026-10-01
updated: 2026-10-01
author: GOOS (manager-spec)
tier: L
---

# plan.md — 구현 계획

## §A. Context

- **브랜치/트리**: `WT-crosshost-rebuild` @ `f22e2d7ac`, 워크트리 `.claude/worktrees/t1375`(fresh from develop `f22e2d7ac`).
- **카드**: t1375 — 절차 plan → plan-audit `--deep` → run → sync, 레인 연속. **plan-audit `--deep` 통과가 run 착수 조건이다.**
- **아티팩트**: `.moai/specs/SPEC-FACTORY-MANAGED-SESSION-001/{spec,plan,acceptance,design,research,progress}.md`.
- **설계 참조**: PR 브랜치 `pr-1722`(head `52381570e`) — 순수 신규 분만 참조(R1, R3, R4). 병합·체리픽 금지(t1365).
- **PRESERVE 대상**: `internal/cli/factory.go`(어휘·조인 판별), `internal/factorymsg/store.go`(무수정 원칙), `internal/codexwiring/configtoml.go`의 `mcpApprovalMode="writes"` 값, `launch_exec_posix.go`/`launch_exec_windows.go`(기존 exec 경로), `internal/hook/session_start_factory*.go`(lane 방향 안내), `internal/kanban/*`.
- **EXTEND 대상**: `internal/cli/`(신규 관리 파일 + cc/glm/codex_launcher 진입 분기), `internal/codexwiring/`(승인 인수 스코핑 보조), `internal/cli/doctor_codex.go`(경고).

## §B. Known Issues (Tier L 필수 — 관련 범주만)

- **B1 크로스플랫폼**: 신규 파일 zero-syscall(AC-MS-014). 자식 소유 모델이라 build tag 분할 불필요가 설계 전제(D-4) — `syscall` 참조가 새로 생기면 설계 위반.
- **B2 cross-SPEC 정합**: SPEC-FACTORY-RECORD-001 v0.3.0(T29b/T29c)·SPEC-FACTORY-LANE-JOIN-SOCKET-001(-l/--leader)·SPEC-FACTORY-LANE-AUTONOMY-001(디스패치)과 표면 인접 — 관리 계층은 이들 소유 표면을 수정하지 않는다(design.md D-1 테이블).
- **B3 서브에이전트 경계**: `internal/cli` 신규 코드에 AskUserQuestion 호출 금지 — grep 0 유지.
- **B6 헤딩**: spec.md §F는 `### Out of Scope — <topic>` H3 형식 준수(작성 완료).
- **B8 트리 위생**: 커밋은 명시 pathspec 스테이징. `.moai/state/`·`.moai/harness/` 무수정.
- **B9 커밋**: 레인 연속 — 커밋+로컬 develop 병합은 레인이 수행, push는 리더 일괄(gitflow-lane-protocol §4). `Authored-By-Agent: manager-spec`(plan)/`manager-develop`(run) 트레일러.

## §C. Pre-flight (run 착수 전 재실행)

```bash
git branch --show-current && git rev-parse --short HEAD   # WT-crosshost-rebuild @ f22e2d7ac 계열 확인
go build ./...
GOOS=windows GOARCH=amd64 go build ./...
go test ./internal/factorymsg -count=1                     # 브로커 기준선 녹색
go test ./internal/cli -run '^(TestFactory|TestParseLauncherEntry)$' -count=1   # 어휘 기준선 녹색
golangci-lint run --timeout=2m ./internal/cli/... 2>&1 | tail -5            # 신규 대비 baseline
```

## §D. Constraints

- `store.go` 무수정 — API 부족 시 blocker(REQ-MS-008).
- PR의 어휘 변경 파일 43종 재도입 금지(R13 기각 목록).
- 병합 자동화·Decider·T29 핸드오버 발행 금지(REQ-MS-014).
- `--no-verify` 금지, 강제 push 금지, 어휘 회귀 금지.
- 문서·안내 한국어, 코드 주석·커밋 영어, 커밋 트레일러 `🗿 MoAI`.

## §E. Self-Verification (run 종료 시)

E1 AC 18행 PASS/FAIL 매트릭스(명령+verbatim 출력) / E2 `go build ./...` + `GOOS=windows` 크로스빌드 exit 코드 / E3 관리 파일 커버리지(`go test -cover ./internal/cli/...` 85%+) / E4 서브에이전트 경계 grep 0 / E5 lint 신규-vs-baseline 분리 보고 / E6 커밋 SHA 목록+push 상태(레인은 로컬 병합 SHA 보고) / E7 blocker 유무 / E8 TDD RED 실패 출력(구현 전 verbatim).

## §F. Milestones (되돌릴 수 있는 결정 순 — 변화 가능성 높은 것 우선)

- **M1 — 관리 세션 코어 인터페이스 + Claude/GLM 소유자** (데이터 모델·인터페이스 — 최고 변화 가능성): `managed_factory_session.go`를 현 API 위로 재작성 — stream-json 강제, 직렬 큐, launch-pending→bind/rollback, PeerByOwner claim 루프, 메타데이터 프롬프트(R4). 순수 신규, 의존성 0. TDD: 큐 직렬화·플래그 거부·프롬프트 조립 RED 먼저.
- **M2 — 관리 Codex App Server 세션 + websocket 의존**: `managed_codex_factory.go` 재작성 — gorilla/websocket v1.5.3 추가(D-2), 토큰/readyz/스레드 바인딩/턴 주입(R3). AC-MS-001..006. live 게이트 테스트 골격(AC-MS-016).
- **M3 — 런처 배선**: cc/glm/codex_launcher의 factory 진입에 관리 경로 분기(R8), Claude-only 송신 거부 유지(AC-MS-011), `--` 순서 회귀 테스트(R2). 일반 실행 경로 무변경.
- **M4 — MCP 승인 스코핑 + doctor**: 소유 프로세스 한정 approve 인수(R11), 프로젝트 `writes` 불변(AC-MS-012), 구식 전역 승인 경고(AC-MS-013).
- **M5 — loopback 통합 + 회귀·문서**: re-exec 가짜 세션 왕복(AC-MS-015), `internal/cli`·`internal/kanban`·`internal/hook`·`internal/template` 회귀 스위트, 관리 세션 사용 안내 문서(한국어) — rules 문서 축 최소 수정.
- **M6 — 감사·마무리**: E1..E8 셀프검증, CHANGELOG 사전 점검(B12), sync 핸드오프.

각 마일스톤 = 커밋 1개 이상, `feat(SPEC-FACTORY-MANAGED-SESSION-001): M<n> ...` + 카드 id 트레이스. 우선순위: M1·M2 High, M3·M4 High(배선 없으면 데모 불가), M5 Medium, M6 Medium.

## §G. Anti-Patterns

- PR 코드 복붙 후 어휘만 치환 — API·구조를 develop 기준으로 재판정(R3·R4와 R6 대조)한다.
- receipt를 완료 판정으로 사용(REQ-MS-015 위반).
- 프로젝트 전역 승인 완화(D-5 기각).
- `store.go` 침묵 패치 — blocker 없이.
- 관리 계층 안에 컨트롤러 기계 유입(AC-MS-010 위반).

## §H. Cross-References

- `design.md` D-1..D-7 / `research.md` R1..R14 / `acceptance.md` AC-MS-001..018.
- `.moai/reports/t1365/verdict.md` — 경로 판정. `pr-1722` 참조물 보존.

# t43 — BranchGuard deny 메시지가 도달 불가능한 해결책을 안내하던 결함 수정

카드: t43 (칸반 tjv7iy 배치) · 워크트리: `.claude/worktrees/t43` (branch `WT-t43`, base `release/v3.1.1` @ 36a12cf82 병합)

## 1. 결함 확정 (전제 검증)

- `internal/hook/branch_guard.go:290` (release/v3.1.1 기준)의 deny 문구:
  ```go
  reason = fmt.Sprintf("%s: %s in primary checkout (use a worktree or invoke via manager-git)", ...)
  ```
- 동일 파일 `:24-41` 주석은 두 면제 축(AgentType 신원 / `MOAI_BRANCH_GUARD_EXEMPT` 센티넬) 모두
  tool-spawned subagent에서 도달 불가임을 이미 명시 — 즉 "invoke via manager-git"은 서브에이전트 위임 경로로는
  반드시 동일 거부로 되돌아오는 안내였음 (2개 세션에서 턴 낭비 관측, 카드 원문).

## 2. 변경 내용 (전/후)

**후 (deny 메시지)** — `internal/hook/branch_guard.go`:
```go
reason = fmt.Sprintf("%s: %s in primary checkout (use a worktree; the manager-git identity and %s exemptions fire only for main-thread launches, not for tool-spawned subagents)",
    branchGuardViolationPrefix, suffix, branchGuardExemptEnv)
```
- 워크트리 경로를 1차 안내(카드 조치 b) + manager-git/센티넬 면제를 "메인 스레드 실행 한정"으로 한정(조치 a).
- 환경변수명은 기존 상수 `branchGuardExemptEnv`에서 삽입(하드코딩 방지 §14).
- `checkBranchState` doc 주석에 안내문 근거 추가.

**테스트** — `internal/hook/branch_guard_test.go` 신규 `TestBranchGuard_DenyReasonRemediationContract`:
- "use a worktree" 포함 ✓ / "invoke via manager-git" 부재 ✓ / 면제 도달성 한정 문구 포함 ✓

**룰 문서 (조치 c)** — `.claude/rules/moai/workflow/main-checkout-branch-guard.md` + 템플릿 미러:
- 카드가 요청한 "면제 절 tool-spawned subagent 한정"은 양쪽 모두 **이미 존재**(:136 / 미러 :135, grep 확인) —
  실제 갭은 Go 메시지뿐이었음. 양쪽에 deny 안내문 정렬 문단(v1.3.1)을 동일 텍스트로 추가.
- 미러 추가 라인 중립성: SPEC-/REQ-/날짜 토큰 0건 (grep 확인). 로컬↔미러 delta는 기존 중립화 헝크만 유지.

## 3. 검증 (관측된 출력)

- `unset MOAI_KANBAN … && go test -count=1 -run 'TestBranchGuard|TestIsExemptAgent|TestIsPrimaryCheckout' -v ./internal/hook/`
  → `--- PASS` 23건 (신규 테스트 포함), `ok github.com/modu-ai/moai-adk/internal/hook 6.874s`
- `golangci-lint run ./internal/hook/...` → `0 issues.`
- `make build` → exit 0 (catalog.yaml 재생성 + 바이너리 재컴파일)
- 전체 패키지 런(`go test ./internal/hook/... ./internal/template/...`): `internal/template` ok,
  `internal/hook`는 **기존 결함 1건만 FAIL** — `TestPreTool_AstGrepSkipReasonSurfaces`.

### 기존 결함 재현 (t43 변경 무관 증명)

t43 변경을 임시 revert(클린 베이스 36a12cf82)한 상태에서 동일 명령 재실행:
```
--- FAIL: TestPreTool_AstGrepSkipReasonSurfaces (0.00s)
    pre_tool_astgrep_reason_test.go:78: SystemMessage is empty: the ast-grep skip reason was dropped somewhere in the three-frame chain
```
→ release/v3.1.1 베이스 자체의 실패. t43 이전 커밋(아마 t62 cwd_fallback 로그 분리 또는 인접 병합)에서 유입 추정,
본 카드 범위 외 — 리드 회신에 명시함.

## 4. 커밋 / 통합

- 커밋: `fix(hook): stop branch-guard deny reason from suggesting the unreachable manager-git route` (아래 기입)
- 자가 통합: release-v311 워크트리 진입 → `git merge --no-ff WT-t43` → `git push origin release/v3.1.1`
  (머지 SHA는 리드 보고 메시지에 기입)

# t236 재현 증거 — MOAI_PROJECT_DIR 워크트리 전환 후 스테일 (이슈 #1640 잔여)

측정 트리: `65196a5a7` (origin/develop tip, 워크트리 `.claude/worktrees/t236`, 브랜치 `WT-project-dir-stale`)
측정 시각: 2026-09-02 (lane 세션, source_session_id=018df390-f2ab-46f9-9a20-0ecb2c44a0ae)
러닝 mcp-server: build v3.1.2 commit `64bba61aa` (2026-08-31 빌드 — project_root 착지 이후 빌드)

## Claim

이슈 #1640의 잔여 결함 2축이 모두 현재 develop에서 실재하며, 이슈의 미검증 항목 (ii)(라이브
파라미터 부착)은 해소됐고 (i)(PostToolUse 실발화)는 본 세션에서 직접 검증 불가(신규 배선이
세션 시작 스냅샷에 반영되지 않음)로 잔여 위험 문서화 대상이다.

## Evidence (명령 + 출력 원문)

### L1 — EnterWorktree 후 세션 레지스트리 cwd 미재배치 (잔여 A의 라이브 증상)

본 세션은 primary에서 시작해 `EnterWorktree(t236)`를 수행했다(트랜스크립트 기록). 이후:

```
$ python3 -c "... (레지스트리에서 session_id 018df390-f2ab-46f9-9a20-0ecb2c44a0ae 엔트리 판독)"
{
  "session_id": "018df390-f2ab-46f9-9a20-0ecb2c44a0ae",
  "spec_id": "(none)",
  "phase": "(none)",
  "started_at": "2026-09-02T06:01:14.415979Z",
  "last_heartbeat": "2026-09-02T06:01:14.415979Z",
  "pid": 12451,
  "host": "goos.local",
  "cwd": "/Users/goos/MoAI/moai-adk-go"
}
```

`cwd`가 launch-time 값(primary)에 고정 — 워크트리 경로로 재배치되지 않음. t74의
`relocateSessionCwd`(`internal/hook/cwd_changed_relocate.go`)는 CwdChanged 훅에서만 호출되는데
그 이벤트가 발화하지 않았음을 의미.

배선 무결성 통제변수:
```
$ ls -la .claude/hooks/moai/handle-cwd-changed.sh
-rwxr-xr-x 1 goos staff 1562 Sep  2 15:05 .claude/hooks/moai/handle-cwd-changed.sh
$ ls .moai/logs/hook-missing.log
ls: .moai/logs/hook-missing.log: No such file or directory
```
래퍼 존재 + 미싱 폴백 로그 부재 → 배선 자체는 정상. "이벤트 미발화"가 지배 설명.

### L2 — mcp-server 폴백의 조용한 오독 + 파라미터 라이브 부착 (잔여 B + 이슈 미검증 (ii))

probe SPEC을 워크트리에만 생성(`.moai/specs/SPEC-T236-PROBE-001/spec.md`, 미커밋 스크래치) 후
러닝 서버에 필터 조회 2건:

```
mcp__moai__spec_audit { "filter_spec": "SPEC-T236-PROBE-001" }
→ {"audited_at":"2026-09-02T06:19:55.607422Z","total_specs":0,...,"drift_findings":[]}

mcp__moai__spec_audit { "filter_spec": "SPEC-T236-PROBE-001",
                        "project_root": ".../.claude/worktrees/t236" }
→ {"audited_at":"2026-09-02T06:19:56.32202Z","total_specs":1,...,
   "drift_findings":[{"spec_id":"SPEC-T236-PROBE-001",...,"severity":"INFO",...}]}
```

- no-param: **세션이 현재 앉아 있는 트리에 존재하는 probe가 0으로 조용히 누락** — 오류·경고·
  어느 트리를 읽었는지 표시 전무. 폴백이 spawn 시점 상태(CLAUDE_PROJECT_DIR/server cwd)에
  얼어 있는 결함의 종단간 라이브 재현.
- param: probe 발견 — **파라미터 부착 라이브 동작 확인** → 이슈 미검증 항목 (ii) 해소.
- 응답에 `_root` 프리비넌스 없음 → 폴백 가시화 갭도 라이브 확인.

### L3 — verify_snapshot/verify_trend에 project_root 파라미터 부재 (잔여 B)

```
$ grep -n "project_root" internal/cli/mcp_server.go
(0 matches — :192-198 verify_snapshot, :201-206 verify_trend 등록에 projectRootOption() 없음)
$ sed -n '689,690p;715,720p' internal/cli/mcp_server.go
func handleVerifySnapshot(...): root := resolveProjectDir()   # :690
func handleVerifyTrend(...):    verify.Load(resolveProjectDir(), key)  # :720
```
세션 툴 스키마에서도 verify_snapshot/verify_trend에 project_root 인자 없음.

### L4 — EnterWorktree/ExitWorktree를 커버하는 PostToolUse matcher 부재 (잔여 A 배선)

```
$ python3 -c "(로컬 settings.json PostToolUse 판독)"
matcher: Write|Edit|MultiEdit   (handle-post-tool.sh + status-transition-ownership.sh×3)
matcher: None
```
`internal/hook/post_tool_worktree.go` 파일 자체가 develop에 없음(`ls internal/hook/` 실측),
`post_tool.go`에 EnterWorktree/ExitWorktree 분기 없음.

### L5 — 패치 초안 심볼이 develop에 전무 (패치 미착지 재확인)

```
$ grep -rn "resolveMCPRoot|resolveProjectDirWithSource|stampProjectDirEnv|handleWorktreeMove" internal/ --include='*.go' | grep -v _test.go
(0 matches)
```

### L6 — MOAI_PROJECT_DIR의 코드 소비자 전무 (이슈 서술 정정 소관)

```
$ grep -rn "MOAI_PROJECT_DIR" internal/ pkg/ cmd/ --include='*.go' | grep -v _test.go
internal/hook/cwd_changed.go:66  (주석)
internal/hook/cwd_changed.go:68  (export 생성 — 유일 생산자)
```
소비자 0. 이슈 원문의 "mcp-server project_root 폴백이 MOAI_PROJECT_DIR을 읽는다"는 서술은
현재 코드와 불일치 — 폴백은 `CLAUDE_PROJECT_DIR` env → server cwd를 음
(`internal/cli/session.go:264-272 resolveProjectDir` 실측). 결함의 실체는 변수명이 아니라
**spawn-frozen 폴백**이며, MOAI_PROJECT_DIR 스탬프는 시설 정합성·사용자 스크립트용.

## Baseline-attribution

모두 본 세션, 워크트리 HEAD `65196a5a7`에서 직접 측정. L2는 러닝 서버 빌드 `64bba61aa`
(2026-08-31) 상에서의 관측 — project_root 착지(`2026-08-23`) 이후 빌드로 파라미터 경로 유효.

## Gaps

- 이슈 미검증 항목 (i): PostToolUse가 EnterWorktree에서 실제 발화하는지 — 훅 설정은 세션 시작
  스냅샷이므로 본 세션에서 신규 matcher를 실험 불가. 착지 후 **신규 세션에서의 첫 systemMessage
  관측**으로 검증 예정. CC 문서상 matcher는 도구명 정규식이며 EnterWorktree는 일반 도구로
  스키마·permissions.allow에 등장 — 문서 근거는 구현 시 보강.
- CwdChanged 미발화의 런타임 내부 사유(CC가 이벤트를 억제하는지)는 관측 불가 — 본 증거는
  "발화하지 않아 재배치가 없었다"는 증상 수준까지가 범위.
- L1의 대안 설명(이벤트는 발화했으나 relocate가 실패)은 완전 배제 못함 — 해당 경로는
  slog.Warn만 남기며 그것은 훅 stderr로 유실. 다만 레지스트리가 존재하고 엔트리를 포함함을
  확인했으므로 실패 경로의 확률질량은 낮음.

## Residual-risk

- PostToolUse matcher가 신규 세션에서도 EnterWorktree에 발화하지 않으면 잔여 A의 시스템 메시지
  경로는 여전 no-op — 착지 후 신규 세션 관측이 닫을 항목.
- 러닝 mcp-server가 구빌드(`64bba61aa`)인 점: 착지 후 검증은 바이너리 재빌드+서버 재접속이 전제.

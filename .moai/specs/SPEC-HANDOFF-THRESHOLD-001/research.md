# Research — SPEC-HANDOFF-THRESHOLD-001 (Handoff-v2 M4/4)

> plan-phase codebase 실측 근거. 모든 라인 번호/사실은 worktree HEAD 기준 직접 확인(grep/Read). C-axis blocker 4건의 근거를 이 문서가 SSOT로 보유.

---

## §A — autoCompactThreshold 위치 (C-axis blocker 1 해소)

**질문(task)**: "autoCompactThreshold가 어디 사는지 grep; statusline이 못 읽으면 open question." → **읽을 수 있음. open question 아님.**

실측 (`internal/statusline/memory.go`):
```
16: const defaultAutoCompactPct = 85          // 기본 auto-compact 임계 %
39: func getAutoCompactThreshold() int {       // env override 우선, 아니면 85
40:   if override := os.Getenv(config.EnvClaudeAutoCompactPct); override != "" { ... }
45:   return defaultAutoCompactPct
```
- `getAutoCompactThreshold()`는 `renderer.go`와 **동일 패키지**(`package statusline`) → `shouldShowHandoffGuide`/`handoffGuideStage`가 **직접 호출** 가능. 신규 import·wiring 불필요.
- env override: `CLAUDE_AUTOCOMPACT_PCT_OVERRIDE`(`internal/config/envkeys.go:89` `EnvClaudeAutoCompactPct`). 하드 상한 공식이 이 함수를 통하면 override도 자동 반영.
- `memory.go:212` `effectiveBudget = contextSize * threshold / 100` — CW 바는 auto-compact 스케일된 budget으로 100% 도달. 단 `shouldShowHandoffGuide`는 **raw** `ContextWindowSize`·`TokensUsed`(raw %)를 사용(§B) → 하드 상한도 raw % 도메인.

**결론**: 하드 상한 = `min(config.HandoffHardCeilingCapPct, getAutoCompactThreshold() + config.HandoffHardCeilingMarginPct)`. reachability 한계(auto-compact 85%가 raw에서 먼저 발화)는 REQ-THRESHOLD-005로 명시(design.md §B.3).

---

## §B — shouldShowHandoffGuide 현재 상태 (D1 기저)

실측 (`internal/statusline/renderer.go:578-592`):
```go
func shouldShowHandoffGuide(data *StatusData) bool {
    if data == nil { return false }
    cwSize := data.Memory.ContextWindowSize
    if cwSize <= 0 { return false }
    used := data.Memory.TokensUsed
    rawPct := float64(used) * 100.0 / float64(cwSize)
    if cwSize >= 500_000 { return rawPct >= 50.0 }
    return rawPct >= 90.0
}
```
- suffix 렌더: `renderBarsInline`(316행) `if shouldShowHandoffGuide(data) { bar += " (⚠️/clear)" }`(324-326행). **무조건**(config 무관, 순수 usage 함수) — REQ-006 불변식의 기저.
- 인라인 리터럴 `500_000`, `50.0`, `90.0` → REQ-004로 config 상수화.
- M1 테스트: `internal/statusline/stdinfields_test.go` `TestShouldShowHandoffGuide_*`(L31-116). `shouldShowHandoffGuide` bool 시그니처 유지 시 무손상 → REQ-001 wrapper 전략.
- **renderer.go는 현재 config를 import하지 않음**(grep 확인). REQ-004 상수 참조를 위해 `internal/config` import 추가 필요(memory.go가 이미 하는 패턴).

---

## §C — 호출부 배치 (C-axis blocker: session_id + Memory 동시 스코프)

실측 (`internal/statusline/builder.go`):
```
138: func (b *defaultBuilder) Build(ctx, r) (string, error) {
142:   input := b.parseStdin(r)          // input.SessionID 여기 스코프 (types.go:59)
145:   data := b.collectAll(ctx, input)  // data.Memory.ContextWindowSize/TokensUsed
148:   result := b.renderer.Render(data, mode)
150:   return result, nil
```
- `collectAll`(187행)은 `CollectMemory(input)`로 `data.Memory`만 채우고 **`input.SessionID`를 StatusData에 캡처하지 않음**. `StatusData`(types.go:219-250)에 **SessionID 필드 없음** — 확인.
- 따라서 "naive call-site"(collector 내부 또는 renderer)는 session_id 미보유 → task가 지적한 misplacement. **해소: 쓰기는 `Build`에서 `collectAll` 직후**(`input.SessionID` + `data.Memory` 동시 스코프)에 `writeContextUsage(...)` 호출. `b.homeDir` 보유하나 projectDir는 stdin에서 유도(§E).
- 대안(StatusData에 SessionID 필드 추가 + collector 쓰기)은 스키마 확장 + collectAll 병렬성 고려 필요 → 불필요. Build 배치가 최소 변경.

---

## §D — atomic write 선례 (D3 재사용)

실측 (`internal/statusline/model_cache.go` `WriteModelCache`):
```
45: func WriteModelCache(homeDir, modelName string) error {
53:   os.MkdirAll(stateDir, 0o755)      // silent-ignore
64:   os.WriteFile(tempPath, ...)        // temp
70:   os.Rename(tempPath, cachePath)     // atomic; 실패 시 temp cleanup + nil
```
- 패턴: MkdirAll → temp write → atomic rename → **모든 실패 silent(nil 반환, EC-SF-003)**. `WriteContextUsage`는 이 패턴 그대로 재사용(JSON marshal만 추가). **신규 write 메커니즘 도입 금지**(task 명시).
- M3의 `internal/hook/handoff/` `atomicWriteFile`도 동일 계열이나, D3는 statusline 패키지 내 작업이므로 **model_cache.go 선례**가 더 가깝고 패키지-로컬.

---

## §E — 기존 소비 경로 + 파일 위치 (D2/D4)

- **HandoffConfig**: `internal/config/types.go:601` `HandoffConfig{Mode string; Guide bool}`. default `internal/config/defaults.go:104` `DefaultHandoffMode="manual"`, `NewDefaultHandoffConfig()`(192행) `Guide:false`. `DefaultHandoffStaleTTL`(113행, var — time.Duration).
- **Guide 소비 지점(M3)**: `internal/hook/handoff_inject.go:146` `handoffConfig()` → `return mode, c.Handoff.Guide`(159행). notice-only 셀 stderr 힌트 게이트(REQ-AUTORESUME-010). **M4는 이 경로 무변경** — Guide advisory는 D4 독트린 서술로만 확장.
- **Mode 소비(M3)**: `handoffInjectHandler` auto-resume 게이트(`source==clear ∧ mode==auto`). M4 무접촉.
- **projectDir 유도**: `input.CWD`(types.go:61 `json:"cwd"`) 또는 `input.Workspace.CurrentDir`. 폴백 `os.Getwd()`. context-usage.json은 orchestrator가 project cwd에서 읽으므로 **project-relative** `<projectDir>/.moai/state/context-usage.json`. (M3 handoff pending.json도 `<projectDir>/.moai/state/handoff/`로 project-relative — 정합.)
- **Detection 독트린 + 256K template drift (D1, plan-auditor iter-1)**: `.claude/rules/moai/workflow/context-window-management.md:69` `## Detection Heuristics`(4-신호 휴리스틱). `context-usage` 언급 0(신규).
  - LIVE `grep -c '256,000' .claude/rules/moai/workflow/context-window-management.md` = **1**(28행 `Opus/Fable (256K) | 256,000 tokens | 90% | ~230,000 tokens`, M1 추가).
  - template mirror `grep -c '256,000' internal/template/templates/.claude/rules/moai/workflow/context-window-management.md` = **0**(부재 — M1이 mirror 미동기화 drift). template Targets 표는 1M/200K/200K 3행뿐.
  - → D1 정정: D4가 template mirror에 256K 행 ADD(parity), Detection 절은 BOTH files section-level 편집, full-sync 금지(LIVE 256K 삭제 회귀 방지). REQ-016/017 + AC-016/017.
- **template mirror 존재**: `internal/template/templates/.claude/rules/moai/workflow/context-window-management.md` — grep 확인. **task 전제("template 밖")는 오기**. D4는 Template-First(§2): template 편집 → `make build` → live sync. template 사본은 §25 중립성(SPEC-ID/REQ 토큰 금지) 준수.

---

## §F — 실측 안 된 / 구현 시 확인 사항 (Gaps)

- `input.Workspace.CurrentDir` 정확 필드명(vs `input.CWD`) — 둘 다 존재하나 우선순위는 구현 시 확정(design.md §E 권장: Workspace.CurrentDir → CWD → os.Getwd).
- write-if-changed throttle 시 `captured_at` plateau 캐비어트(usage 정체 시 mtime 미갱신) — reader freshness 창을 관대하게(session-scoped) → 잔여 위험(design.md §D.4).
- reader(orchestrator/독트린)의 state-file 파싱은 Go 코드가 아닌 **독트린 서술** — 런타임 자동 파서 없음(D4는 문서). 향후 Go 파서는 별도 SPEC.
- **`writer_pid` session-stability (D2, plan-auditor iter-1)**: statusline은 Claude Code가 render마다 `.moai/status_line.sh` 래퍼 경유로 spawn하는 fresh 프로세스 → `os.Getpid()`는 render-ephemeral, `os.Getppid()`(래퍼 shell)도 render마다 새로 spawn되어 완전 session-stable 아님. 완전 session-stable 토큰(예: stdin `transcript_path` 파생 — 세션당 유일·render 간 불변)은 실측 미확인, 후속 Go reader 구현 판단. 따라서 M4 writer_pid는 (1) throttle 제외, (2) Go 헬퍼(`isFreshForSession`, AC-018)만 명시 `curWriterID` 공급받아 기계 guard, (3) 독트린-only reader는 single-session 가정 하 freshness 검사 — concurrent empty-id는 residual(design §D.4b).
- pre-existing baseline: `internal/cli TestRunHookEvent_ReadInputError` nil-deref panic(M3 §E.3 기록) — 본 SPEC 범위 밖(statusline 무관).

---

## §G — Cross-References

- design.md §A~F (결정), spec.md §C(REQ), acceptance.md §D(AC)
- `internal/statusline/{renderer,memory,builder,model_cache,types}.go`
- `internal/config/{types,defaults,envkeys}.go`
- `internal/hook/handoff_inject.go`
- `.claude/rules/moai/workflow/context-window-management.md` (+ template mirror)
- M1 `SPEC-HANDOFF-CTXGUIDE-001`, M3 `SPEC-HANDOFF-AUTORESUME-001`

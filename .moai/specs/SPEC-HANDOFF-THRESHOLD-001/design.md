# Design — SPEC-HANDOFF-THRESHOLD-001 (Handoff-v2 M4/4)

> 4개 C-axis blocker를 구체적 결정으로 해소한다. 모든 결정은 research.md 실측에 근거하며, 2개 LOCKED 사용자 결정(기존 HandoffConfig 필드만 소비 / 밴드 경계 하드코딩 defaults.go 상수)을 준수한다.

---

## §A — 아키텍처 개관

M4 = 3개 표면(statusline · config 상수 · 독트린) 편집 + M3 무접촉:

```
[statusline 렌더 경로]
  builder.Build(input)                                  ← D3 호출부
    ├─ data := collectAll(input)                        (data.Memory)
    ├─ writeContextUsage(projDir, input.SessionID, data.Memory, stage)  ← D3 (무조건, atomic, throttle)
    └─ renderer.Render(data)
          └─ renderBarsInline → handoffGuideStage(data) ← D1 열거형 게이트
               ├─ none → suffix 없음
               ├─ soft → " (⚠️/clear)"   (M1 불변)
               └─ hard → " (🛑/clear!)"   (2단계, min(95, ac+margin))

[config]
  defaults.go: Handoff{Soft,Hard,Cutoff} 명명 상수       ← D1/§14
  HandoffConfig{Mode,Guide}                             ← M3 landing, statusline 미소비 (D2 불변식)

[독트린]
  context-window-management.md § Detection Heuristics    ← D4 state-file-first
    (template mirror + live, Template-First §2)
```

핵심 불변식: **statusline suffix + state-file write는 HandoffConfig 무관(무조건)** → M1 무회귀(D2/REQ-006).

---

## §B — D1: 2단계 statusline suffix (blocker 1: 하드 상한)

### B.1 stage 열거형 + wrapper (REQ-001/002)

```go
type handoffStage int
const (
    handoffStageNone handoffStage = iota
    handoffStageSoft
    handoffStageHard
)

// handoffGuideStage: none < soft < hard by usage. hard 먼저 판정.
func handoffGuideStage(data *StatusData) handoffStage {
    if data == nil { return handoffStageNone }
    cwSize := data.Memory.ContextWindowSize
    if cwSize <= 0 { return handoffStageNone }
    rawPct := float64(data.Memory.TokensUsed) * 100.0 / float64(cwSize)
    soft := softThresholdPct(cwSize)   // 밴드
    hard := hardCeilingPct(cwSize)     // §B.3
    switch {
    case rawPct >= hard: return handoffStageHard
    case rawPct >= soft: return handoffStageSoft
    default:             return handoffStageNone
    }
}

// M1 backward-compat: 기존 bool 시그니처 유지 → TestShouldShowHandoffGuide_* 무손상.
func shouldShowHandoffGuide(data *StatusData) bool {
    return handoffGuideStage(data) != handoffStageNone
}
```

`renderBarsInline`(현 324행) 교체:
```go
switch handoffGuideStage(data) {
case handoffStageHard: bar += " (🛑/clear!)"
case handoffStageSoft: bar += " (⚠️/clear)"   // M1 문자열 verbatim
}
```

### B.2 밴드 soft 임계 (REQ-004, config 상수)

```go
func softThresholdPct(cwSize int) float64 {
    if cwSize >= config.HandoffLargeWindowCutoff { // 500_000
        return config.HandoffSoftLargePct          // 50
    }
    return config.HandoffSoftStandardPct           // 90
}
```
- M1 로직(≥500K→50 / <500K→90) verbatim 승계, 리터럴만 config 상수로.
- `internal/config/defaults.go` 신규 상수(§14):
  ```go
  const (
      HandoffSoftLargePct        = 50   // ≥500K 밴드 soft
      HandoffSoftStandardPct     = 90   // <500K 밴드 soft
      HandoffLargeWindowCutoff   = 500_000
      HandoffHardCeilingCapPct   = 95   // 하드 상한 절대 cap
      HandoffHardCeilingMarginPct = 10  // auto-compact 위 margin (§B.3)
  )
  ```
  타입: int 상수(pct는 int, 비교 시 float64 캐스팅). `DefaultHandoffStaleTTL`처럼 var 필요 없음(모두 정수 const).

### B.3 하드 상한 공식 (blocker 1 핵심, REQ-003)

```go
func hardCeilingPct(cwSize int) float64 {
    ac := getAutoCompactThreshold()  // statusline/memory.go, 동일 패키지 (research §A)
    ceil := ac + config.HandoffHardCeilingMarginPct
    if ceil > config.HandoffHardCeilingCapPct {
        ceil = config.HandoffHardCeilingCapPct
    }
    soft := softThresholdPct(cwSize)
    if float64(ceil) < soft {           // clamp: degenerate override → hard=soft
        return soft                      // stage-2가 soft 위로 올라가지 못하면 soft로 흡수
    }
    return float64(ceil)
}
```

**결정 근거 (blocker 1)**:
- 기본값(ac=85, margin=10): `min(95, 95) = 95`. → soft 90 < hard 95 (5-pt gap). 1M 밴드: soft 50 < hard 95.
- **clamp의 목적**: env override로 ac가 낮아지면(예 ac=60 → 60+10=70) `<500K` 밴드 soft(90)보다 낮아진다. 이때 hard=soft(90)로 흡수 → stage-2가 soft 임계와 동일 지점에서 `🛑`로 발화(inversion 방지). 즉 "degenerate override → 오직 hard만 표시(soft 창 소멸)". 안전한 저하.
- **reachability 한계 (REQ-005, 명시 의무)**: raw 도메인에서 auto-compact가 ~85%에 먼저 발화하므로, hard(95%)는 실사용에서 auto-compact에 자주 선점된다(드물게 발화). 이는 auto-compact-aware 공식의 **의도된 tradeoff**이며, 독트린 + CHANGELOG에 "stage-2 always fires 주장 금지" 문구로 명시. M1의 soft(90%)도 동일 도메인 한계를 갖지만 실사용에서 동작하므로(M1 shipped) hard는 그보다 marginal하게 덜 도달. 대안(`ac - margin`, auto-compact 직전 경고)은 task 명시 공식(`+margin`)과 배치되어 채택 안 함 — task 공식 준수 + 한계 명시.

### B.4 emoji/문자열
- soft: `" (⚠️/clear)"` (M1 verbatim, 회귀 금지).
- hard: `" (🛑/clear!)"` (2단계 구별 마커, `(⚠️/clear)` 괄호 스타일 정합). Go 소스 내 emoji 리터럴은 statusline 관례(⚠️ 이미 사용) 허용.

---

## §C — D2: HandoffConfig 소비 reconciliation (blocker: M1 무회귀)

### C.1 불변식 (REQ-006/007) — statusline은 config 미소비

- statusline `renderBarsInline`/`handoffGuideStage`/`writeContextUsage`는 `HandoffConfig`를 **읽지 않는다**. 순수 usage 함수 + 무조건 write. → default(Mode=manual, Guide=false)에서도 soft suffix + write 그대로 발생.
- **회귀 위험 정의**: 만약 suffix를 `guide==true`에 게이팅하면 default(false)에서 suffix 소멸 → M1 shipped 동작 손실. **금지**(spec §B.2 Out of Scope). AC-006이 이 불변식을 잠근다.
- 설계상 보장: statusline 패키지가 config를 suffix 목적으로 import하지 않음(밴드 상수는 config에서 읽되, 그것은 HandoffConfig가 아닌 defaults.go 상수 — config struct 필드 아님).

### C.2 Guide/Mode 경계 화해 (REQ-008)

| 필드 | 소비자 | M4 변화 | statusline 게이팅? |
|------|--------|---------|--------------------|
| `Mode` (auto/manual) | M3 `handoffInjectHandler` (SessionStart auto-resume) | 무변경 | ❌ (statusline 무관) |
| `Guide` (bool) | M3 notice-only stderr 힌트 + D4 독트린 advisory | 독트린이 state-file advisory를 Guide-gated로 서술 | ❌ (statusline 무관) |

- **Guide advisory는 신규 Go 훅 코드 아님**(spec §B.2). D4 독트린이 "Guide==true면 orchestrator가 state-file 기반 advisory를 표면화할 수 있다"고 **서술**. M3 handler는 무변경. 이로써 M4 코드 표면 최소화 + M1 불변식 보호.
- **왜 이 화해인가**: task D2 "config 소비는 advisory/persistence 경로에 적용, unconditional statusline render 아님". statusline을 config에서 완전히 격리(순수 함수)하는 것이 불변식을 **구성적으로** 보장하는 가장 안전한 설계. positive 소비(Guide advisory)는 독트린 레이어로 이동해 statusline 오염 0.

---

## §D — D3: context-usage.json 영속화

### D.1 스키마 (REQ-010)

```json
{
  "schema_version": 1,
  "session_id": "<uuid|empty>",
  "writer_pid": 48213,
  "captured_at": "2026-07-06T20:30:00.000000000+09:00",
  "context_window_size": 256000,
  "tokens_used": 230400,
  "raw_pct": 90.0,
  "stage": "hard",
  "band": "standard"
}
```
- `stage` ∈ {none, soft, hard}, `band` ∈ {large, standard} (cwSize≥500K → large).
- `session_id`: stdin `input.SessionID` 스냅샷(빈 문자열 허용, §D.3).
- `writer_pid`: 쓰기 프로세스 식별자(D2 hole 정정, §D.4). concurrent same-checkout empty-`session_id` 세션을 구별하는 discriminator. reader-supplied expected writer identity와 비교(§D.4).

### D.2 호출부 + writer (REQ-009/011, blocker: 스코프)

`builder.Build`에서 `collectAll` 직후:
```go
data := b.collectAll(ctx, input)
writeContextUsage(resolveProjectDir(input), input.SessionID, os.Getpid(), data.Memory, handoffGuideStage(data))
result := b.renderer.Render(data, mode)
```
- `resolveProjectDir(input)`: `input.Workspace.CurrentDir` → `input.CWD` → `os.Getwd()`(research §E). project-relative `<projDir>/.moai/state/context-usage.json`.
- `writer_pid` = `os.Getpid()`(쓰기 프로세스 PID, D2 discriminator §D.4a). **session-stability caveat**: statusline은 render마다 fresh 프로세스(status_line.sh 래퍼 경유)이므로 PID는 render-ephemeral일 수 있다. 이 때문에 (1) `writer_pid`는 throttle 비교에서 제외(§D.3), (2) 독트린-only reader는 single-session 가정 하에 freshness만 검사(§D.4b), (3) Go 헬퍼(`isFreshForSession`, AC-018)만 `curWriterID`를 명시 공급받아 기계적 guard 수행. 완전 session-stable 토큰(예: `transcript_path` 파생)은 후속 Go reader의 구현 판단(research §F Gap).
- `writeContextUsage`: model_cache.go `WriteModelCache` 패턴(MkdirAll 0o755 + temp write + atomic rename + silent-fail). JSON marshal 추가. best-effort — 어떤 실패도 statusline 방해 없음(REQ-009).
- `data.Memory.Available==false`면 write skip(원신호 부재).

### D.3 write-if-changed throttle (REQ-012, blocker)

```go
// 읽기 → semantic payload 비교 → 불변 시 skip.
existing := readContextUsage(path)          // best-effort, 실패 시 nil
next := buildRecord(...)
if existing != nil && sameSemanticPayload(existing, next) { return } // skip write
writeAtomic(path, next)                       // captured_at 갱신
```
- `sameSemanticPayload`: `session_id` == ∧ `stage` == ∧ `context_window_size` == ∧ `int(raw_pct)` ==. `captured_at` **및 `writer_pid`는 비교 제외**(둘 다 변경 트리거 아님 — 특히 `writer_pid`는 render-ephemeral이므로 포함 시 매 render write churn 유발 → throttle 무력화).
- 목적: statusline은 매우 빈번히 렌더 → payload 정체 시 디스크 write 억제.
- **plateau 캐비어트**: usage 정체 시 write skip → `captured_at` 미갱신 → reader freshness 창이 짧으면 stale 오판. 완화: reader freshness 창을 관대하게(session-scoped, seconds 아님) — §D.4 잔여 위험.

### D.4 session_id guard + fallback-UUID (REQ-013/014, blocker 2건)

**guard semantics (REQ-013)**:
- writer: 현재 `session_id`를 last-writer-wins로 스탬프. 동일 checkout에 2세션 write 시 최근 세션이 파일 소유.
- reader(독트린): 파일 `session_id` != 현재 세션 → **stale** 판정 → 휴리스틱 폴백. cross-session false-resume 방지(다른 세션 컨텍스트로 재개 금지).

**fallback-UUID dead-path (REQ-014, blocker)**:
- 문제: `input.SessionID`가 빈 문자열/fallback sentinel이면, session_id-일치 guard가 항상 실패 → primary path 사망 → **single-session 공통 경우까지 heuristics-always**.
- 해소: writer는 session_id 부재여도 **여전히 write**(빈 문자열 스탬프). reader는 **양측 session_id가 모두 부재**면 session_id 일치 대신 **`captured_at` freshness + `writer_pid` discriminator**로 유효성 판정(§D.4a). 이로써 UUID 미노출 환경(공통 single-session)에서 primary path 생존.

### D.4a `writer_pid` discriminator — concurrent empty-id hole 정정 (REQ-018, plan-auditor iter-1 D2)

**hole 정의 (plan-auditor iter-1 D2 SHOULD-FIX)**: REQ-013 session_id guard는 UUID 세션의 cross-session false-resume를 막지만, REQ-014의 empty-id fallback(`empty|empty → freshness→valid`)이 **정확히 그 hole을 재개방**한다 — UUID 없는 2개 concurrent same-checkout 세션이 하나의 `context-usage.json`을 공유하면, `captured_at` freshness만으로는 둘을 구별 불가 → session B가 session A의 신선 스냅샷을 자신의 컨텍스트로 오독할 수 있다.

**채택 mitigation — (a) process 식별자 discriminator** (plan-auditor 권고 (a), 대안 (b) 창 축소 / (c) 보증 범위 명시 대비 선택):
- writer는 `writer_pid`(쓰기 프로세스 식별자)를 record에 스탬프한다. concurrent 세션들의 writer는 서로 다른 process 정체를 가지므로 구별 가능하다.
- freshness 헬퍼 `isFreshForSession(rec, curSession, curWriterID)`: empty-`session_id` record에 대해 `captured_at` 신선 **AND** `rec.writer_pid == curWriterID`일 때만 own-session-valid. 불일치 시 stale → 휴리스틱 폴백. 이로써 freshness alone이 cross-read를 허용하던 hole이 닫힌다(Go 헬퍼 레벨 기계적 보장, AC-018 단위 테스트).
- **왜 (a)인가**: consume이 doctrine-only이고 trigger가 "2+ concurrent same-checkout empty-id 세션"이라는 드문 조합이므로, 창 축소(b)는 statusline이 빈번히 write하는 특성상 무력하고(항상 신선), 보증-범위-명시(c)는 single-session 이점을 유지하되 hole을 닫지 못한다. `writer_pid` 필드 추가는 값싸고(1 필드) Go 헬퍼에서 기계적으로 검증 가능하다.

**판정 매트릭스 (reader — Go 헬퍼는 curWriterID 공급, 독트린 reader는 §D.4b 잔여)**:

  | 파일 session_id | 현재 session_id | writer_pid 검사 | 판정 |
  |-----------------|-----------------|-----------------|------|
  | UUID X | UUID X | (불필요) | ✅ valid (session_id 일치) |
  | UUID X | UUID Y | (불필요) | stale → 휴리스틱 |
  | empty | empty | `rec.writer_pid == curWriterID` ∧ fresh | ✅ valid (single-session / 동일 writer) |
  | empty | empty | `rec.writer_pid != curWriterID` | **stale → 휴리스틱 (concurrent empty-id 정정 — cross-read 차단)** |
  | UUID X | empty (또는 반대) | (불필요) | 보수적 stale → 휴리스틱 |

### D.4b 잔여 위험 — 독트린-only reader의 writer_pid 미비교

- Go 헬퍼(`isFreshForSession`)는 `curWriterID`를 공급받아 기계적으로 concurrent empty-id를 차단한다(AC-018). 그러나 **현행 reader는 doctrine-only**(orchestrator가 JSON을 Read tool로 읽음, `curWriterID` 핸들 없음). 따라서 doctrine-layer에서는 concurrent same-checkout UUID-less 세션이 여전히 freshness guard의 기계적 보증 밖이다 — 독트린은 이 경우 보수적 휴리스틱 폴백을 지시한다. `writer_pid` 필드는 후속 Go reader(D3 reader 파서, 후속 SPEC)가 이 hole을 기계적으로 닫을 수 있게 하는 forward-compat 기반이다. 이는 acceptance §D.4 / 본 §D.4b에 residual risk로 명시된다.

---

## §E — D4: Detection Heuristics 독트린 재작성

### E.1 state-file-first (REQ-015)

현 § Detection Heuristics(4-신호 휴리스틱)를 다음 우선순위로 재작성:
1. **`.moai/state/context-usage.json` 우선 읽기** — `raw_pct` + `stage`가 권위 신호. session_id guard(§D.4 매트릭스)로 유효성 판정.
2. **폴백**(파일 부재/stale/파싱실패): 기존 4-신호 휴리스틱(누적 출력 바이트 / system-reminder 볼륨 / large tool 결과 수 / Agent() 호출 수).
- state-file 스키마(§D.1) + guard 매트릭스(§D.4)를 독트린에 명시 → orchestrator가 읽는 법을 안다.
- Guide advisory(§C.2): Guide==true면 orchestrator가 state-file 기반 advisory 표면화 가능(독트린 서술).

### E.2 Template-First + template drift parity (REQ-016/017, blocker: mirror + D1 drift)

- **template mirror 존재**(research §E): `internal/template/templates/.claude/rules/moai/workflow/context-window-management.md`. → **task 전제("template 밖") 정정**.
- **D1 drift 실측 (plan-auditor iter-1 D1 SHOULD-FIX)**: LIVE는 256K 행 보유(`grep -c '256,000'`=1), **template mirror는 부재**(`grep -c '256,000'`=0). M1이 mirror를 동기화하지 않은 drift. 따라서 "256K already present"는 LIVE에만 참, template에는 거짓.
  - **hazard (a)**: AC-016/017의 `grep -c '256,000'`==1이 template에서 0≠1로 실패.
  - **hazard (b)**: run-phase에서 full-file template→live sync(`moai update`)를 하면 LIVE의 256K 행이 **삭제**되어 M1 회귀.
- **section-level 편집만 (BOTH files)** [HARD]: template + live 각각에서 **§ Detection Heuristics 절만** 편집한다. **full-file template→live overwrite / `moai update` full sync 금지** (hazard (b) 방지). Targets 표는 별도 처리(아래).
- **템플릿 256K 행 ADD (parity 회복, REQ-017)**: D4는 template mirror § Context Window Targets에 누락된 `Opus/Fable (256K) | 256,000 tokens | 90% | ~230,000 tokens` 행을 **추가**한다 → 두 사본 모두 `grep -c '256,000'`==1로 수렴(drift 제거). LIVE Targets 표에는 중복 행 추가 금지(이미 1개 존재).
- 절차: (a) template 편집 — Detection Heuristics 절 section-level 재작성 + Targets 표에 256K 행 add(§25 중립 — SPEC-ID/REQ 토큰/내부 날짜/commit SHA 금지) → (b) `make build` → (c) live `.claude/rules/...`도 Detection Heuristics 절만 section-level 편집(Targets 표 무접촉 — 256K 이미 존재).

---

## §F — 실패 모드 (best-effort)

- context-usage.json write 실패(권한/디스크) → silent, statusline 정상 렌더(REQ-009).
- state file 파싱 실패(reader) → 휴리스틱 폴백(독트린).
- getAutoCompactThreshold env override 이상값 → 함수 자체가 1..100 검증(memory.go:41), 벗어나면 85 폴백. hard 공식 안전.
- projectDir 유도 실패(빈 CWD + Getwd 실패) → write skip(경로 없음), 렌더 정상.

---

## §G — Cross-References

- research.md §A~G(실측), spec.md §C(REQ), acceptance.md §D(AC)
- `internal/statusline/{renderer,memory,builder,model_cache}.go`
- `internal/config/defaults.go`(신규 상수), `internal/config/types.go`(HandoffConfig)
- `.claude/rules/moai/workflow/context-window-management.md`(+ template mirror)
- `.claude/rules/moai/core/verification-claim-integrity.md`(REQ-005 "always fires 주장 금지" 근거)
- CLAUDE.local.md §2(Template-First)/§14(하드코딩)/§25(중립성)

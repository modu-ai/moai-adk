# Design — SPEC-HANDOFF-AUTORESUME-001 (Handoff-v2 M3/4, auto-resume)

> §B 재설계 blocker를 구체적 결정으로 해소한다. 모든 결정은 research.md 실측에 근거하며, 3개 확정 사용자 결정(mode default=manual, directive degrade-to-guidance, M1/M2/M3 split)을 준수한다.

---

## §A — 아키텍처 개관

역방향 핸드오프 = 3개 소형 컴포넌트, 기존 인프라 무접촉:

```
[세션 A: orchestrator/user]
    moai handoff save --body <resume>   ← M2 CLI
        └─ writes .moai/state/handoff/pending.json (atomic)      [신규 경로]

        ── /clear (or new session) ──

[세션 B: SessionStart hook, source=clear]
    handoffInjectHandler.Handle()        ← M3 hook (3rd SessionStart handler)
        ├─ config.Handoff.Mode == "auto"?  (default manual → no-op)   [M1 config]
        ├─ input.Source == "clear"?        (non-clear → notice-only, never consume)
        ├─ read pending.json → degrade directives → inject AdditionalContext (i18n header)
        └─ atomic rename pending.json → handoff/consumed/<ts>-<nonce>.json   [claim]
                (loser sees rename failure — errno-agnostic — proceeds normally)
```

기존 `session-handoff/pending.md` → SessionEnd → memory 흐름은 **완전 분리, 무접촉**.

---

## §B — 경로/포맷 조정 verdict (blocker 1 해소)

| 흐름 | 경로 | 포맷 | 소비자 | audit trail | 소유 SPEC |
|------|------|------|--------|-------------|-----------|
| 기존 memory 영속화 | `.moai/state/session-handoff/pending.md` | Markdown | SessionEnd `PersistIfPending` | `<memoryDir>/project_*.md` (삭제 안 함) | SESSION-HANDOFF-AUTO-001 |
| 신규 auto-resume | `.moai/state/handoff/pending.json` | JSON | SessionStart `handoffInjectHandler` | `.moai/state/handoff/consumed/<ts>-<nonce>.json` (삭제 안 함) | 본 SPEC |

**verdict: 별도 경로 + 별도 포맷.** 근거(research.md §B.2): 공유 시 SessionEnd가 pending을 먼저 소비·삭제하여 SessionStart 소비 불가. 두 흐름은 서로 다른 lifecycle 지점(종료 vs 시작), 서로 다른 목적(다수 세션 discoverability vs 즉시 다음-세션 주입)이므로 분리가 자연스럽다.

**상호작용 계약**: `moai handoff save`/`clear`와 `handoffInjectHandler`는 `session-handoff/` 트리를 **읽지도 쓰지도 않는다** (REQ-AUTORESUME-008). 반대로 `PersistIfPending`은 `handoff/` 트리를 모른다. 정적 격리 → race 도달 불가.

### B.1 pending.json 스키마

```json
{
  "schema_version": 1,
  "spec_id": "SPEC-HANDOFF-AUTORESUME-001",
  "phase": "run",
  "saved_at": "2026-07-05T20:30:00.000000000+09:00",
  "saved_by_session": "<uuid|empty>",
  "conversation_language": "ko",
  "directives": {
    "ultrathink": true,
    "ultracode": false,
    "goal": ""
  },
  "body": "<verbatim 6-block paste-ready resume, cut-line markers 포함>"
}
```

- `body`: session-handoff.md 6-block 형식 (cut-line markers `✂──── ... ────✂` 포함, verbatim). 주입 시 그대로 additionalContext에 실린다.
- `directives`: 모드-변경 지시자 메타. **주입 시 활성화되지 않고 "복원 안내"로만 렌더** (§E degrade-to-guidance).
- `conversation_language`: save 시점 언어 스냅샷. 주입 header i18n에 사용(핸들러가 config 재조회도 가능하나 스냅샷 우선).
- `saved_by_session`: attribution용. 부재(environment-fallback) 시 빈 문자열 → nonce fallback 트리거(§C.4).

---

## §C — SessionStart 소비 로직 (blocker 2 해소)

### C.1 신규 핸들러 vs 기존 핸들러 확장

**결정: 신규 `handoffInjectHandler` 등록** (research.md §G.1 (b)). deps.go에서 `sessionStartHandler` → `autoUpdateHandler` 다음, **3번째로 등록**. 근거:
- registry accumulate-all(research.md §C)이 3개 핸들러 additionalContext 공존 보장
- 관심사 분리 + 테스트 격리 (`handoff_inject_test.go` 독립)
- 마지막 등록 → auto-resume 안내가 additionalContext 최후미 append (사용자 가독)

핸들러는 `ConfigProvider`를 주입받아 `cfg.Handoff.Mode`를 읽는다 (sessionStartHandler와 동일 패턴, `NewHandoffInjectHandler(cfg ConfigProvider)`).

### C.2 4-source × mode branch table (blocker 핵심)

`input.Source ∈ {startup, resume, clear, compact}` × `cfg.Handoff.Mode ∈ {manual, auto}`:

| Source | mode=manual | mode=auto |
|--------|-------------|-----------|
| `startup` | no-op (pending 보존) | **notice-only** — pending 존재 시 "auto-resume 대기 중, /clear로 진입" stderr 힌트, **소비 안 함** |
| `resume` | no-op | **notice-only** — resume는 동일 세션 재개, 소비 안 함 (pending은 clear용) |
| `clear` | no-op (pending 보존) | **INJECT + CONSUME** — 유일한 소비 경로 |
| `compact` | no-op | **notice-only** — compact은 컨텍스트 압축 복구(post_compact 담당), 소비 안 함 |

**불변식**: `INJECT+CONSUME`은 `source==clear ∧ mode==auto` 단 하나의 셀에서만 발생. 나머지 7개 셀은 pending.json을 **보존**한다(rename 안 함). notice-only 셀은 `guide` 옵션이 true일 때만 stderr 힌트(비침투, best-effort). manual 4개 셀은 순수 no-op.

**stale 우선순위 (N1 — REQ-019 ⟩ REQ-010)**: auto-mode에서 pending이 **stale**(TTL 초과)이면, auto-mode stale-cleanup(REQ-019)이 notice-only 힌트(REQ-010)보다 **우선**한다. 즉 notice-only 셀(startup/resume/compact) + `guide==true` + stale인 경우, 핸들러는 stale pending을 조용히 제거하고 힌트를 **생략**한다(재개할 live 컨텍스트가 없으므로). 이 우선순위는 `source`·`Guide` 값과 무관하게 적용된다. (§F 실패 모드의 TTL 게이팅과 정합.)

근거: `resume`/`compact`을 소비하면 사용자가 의도한 "clear 후 다음 세션 진입"이 아닌 시점에 컨텍스트가 소진된다. clear만이 명시적 세션 경계 신호.

### C.3 소비 = atomic rename (claim 방식)

소비는 **삭제가 아닌 rename**:
```
os.Rename(handoff/pending.json, handoff/consumed/<ts>-<nonce>.json)
```
- `consumed/`는 audit trail — 핸들러는 memory `.md`도, consumed/도 **삭제하지 않는다**.
- rename은 atomic claim: 2개 세션이 동시에 clear로 진입해도 **먼저 rename에 성공한 세션만** 승자. 패자는 `os.Rename`이 **실패(errno 무관)** → 조용히 정상 진행 (주입 없이, best-effort). POSIX에서는 통상 ENOENT(source 이미 소진)지만, Windows `MoveFileEx`는 동시 rename 실패를 ENOENT로 매핑하지 않을 수 있으므로 핸들러는 **`os.IsNotExist`에 특정하지 않고 "rename err != nil ⇒ 주입 생략"** 으로 판정한다 (D6 cross-platform fail-open, REQ-013).

### C.4 NULL session_id nonce fallback (blocker 명시 항목)

consumed 파일명 `<ts>-<nonce>.json`:
- `ts` = **소비 시각의 정수 타임스탬프 `time.Now().UnixNano()`** (RFC3339 `saved_at` 문자열이 아님 — 파일명은 순수 정수여야 AC-014 정규식 `^\d+-...`가 성립). 예: `1751712600000000000` (단조 증가, 파일명 안전).
- `nonce`:
  1. `saved_by_session` 비어있지 않으면 → 앞 8자 (session8)
  2. 비어있으면(environment-fallback) → `crypto/rand` 8-hex
  3. `crypto/rand` 실패(극히 드묾) → `UnixNano()` 하위 32비트 hex (deterministic fallback)

**결정성 vs 충돌안전성 (설계 불변식 — runtime AC 아님, D7)**: consumed 파일명의 cross-session 충돌은 **atomic-rename-as-claim에 의해 도달 불가능**하다. 논증: `os.Rename(pending.json, consumed/X)`에서 source(`pending.json`)는 유일한 claim 대상이다. 두 세션이 동시에 진입해도 첫 rename이 source를 소진하므로 둘째 rename은 실패(§C.3) → 둘째 세션은 consumed 파일을 애초에 생성하지 않는다. 따라서 서로 다른 두 세션이 동일 파일명을 쓰는 상황 자체가 발생하지 않는다 — nonce 알고리즘의 충돌 확률과 무관하게 성립. nonce는 (a) 세션-내 유일성(단일 세션이 여러 pending을 순차 소비할 때)과 (b) 사람이 읽는 audit 식별자 목적일 뿐이다. session8이 있으면 attribution 유지, 없으면 crypto/rand 8-hex가 유일성 보장. 이 불변식은 설계 논증이므로 AC-014는 파일명 **shape**(`^\d+-[0-9a-f]{8}\.json$`)만 binary 검증하고, 충돌-불가 논증은 본 prose가 SSOT다.

예시:
- session 있음: `consumed/1751712600000000000-a1b2c3d4.json` (nonce=session 앞 8자)
- session 부재: `consumed/1751712600000000000-9f3e0c72.json` (nonce=crypto/rand)

---

## §D — 주입 콘텐츠 렌더 (blocker: directive degrade)

### D.1 verification-claim-integrity 준수 (확정 결정 2)

주입되는 additionalContext는 **ultrathink/xhigh/ultracode/goal이 활성화되었다고 주장하지 않는다**. 훅은 effort/model을 바꿀 수 없다(verification-claim-integrity §1.1 — no unobserved-verification-claim). 모드-변경 지시자는 "복원하려면 이 줄을 입력" **안내 텍스트로 격하**된다.

### D.2 렌더 예시 (conversation_language=ko)

```
[MoAI 자동 재개 — 이전 세션 핸드오프]
아래는 이전 세션이 저장한 재개 컨텍스트입니다. 이 주입은 컨텍스트 전달일 뿐이며,
xhigh/ultrathink를 활성화하지 않습니다. 필요 시 아래 안내 줄을 직접 입력하세요.

복원 안내(수동 입력):
  • ultrathink   ← 확장 추론 활성화하려면 이 줄을 입력
  • /goal <조건> ← 자율 continuation 필요 시 (본 핸드오프에 goal 조건 있음: 없음)

── 저장된 재개 메시지(참고) ──
<pending.json.body verbatim, cut-line markers 포함>
```

- header/안내 텍스트: `conversation_language`(pending.json.conversation_language 또는 language.yaml 재조회, i18n). ko/en/ja/zh는 명시 문자열, 그 외 ISO-639는 en fallback.
- `directives.ultrathink==false`면 ultrathink 안내 줄 생략. `ultracode`/`goal`도 조건부.
- body는 verbatim — degrade는 header/안내에만 적용, body 원문(사용자가 붙여넣을 수 있는 원본)은 변형 없음. body 내부의 `ultrathink.` opener는 "이미 활성"이 아니라 "사용자가 붙여넣을 원본의 일부"로 문맥상 명확(header가 비활성임을 선언).

### D.3 i18n 문자열 소스

핸들러는 `.moai/config/sections/language.yaml`의 `conversation_language`를 읽거나(session_start.go의 cfg 경로 재사용), pending.json 스냅샷을 우선한다. 미지원 locale은 en fallback (session-handoff.md Localization 규칙 동일).

### D.4 64 KiB 절단 상호작용

3개 SessionStart 핸들러 additionalContext 합산이 64 KiB 초과 시 `ValidateHookResponse`(dual_parse.go:92-99)가 절단하고 SystemMessage에 notice. auto-resume body는 diet 제약(paste-ready 예산) 준수 → 통상 수 KB. 절단 도달 시 body가 잘려도 훅은 계속 진행(best-effort). AC로 "합산 64 KiB 근처에서 절단이 우아하게 발생"을 잠근다(선택).

---

## §E — HandoffConfig (M1, config landing)

### E.1 struct (types.go)

```go
// HandoffConfig — auto-resume 핸드오프 설정 (SPEC-HANDOFF-AUTORESUME-001).
type HandoffConfig struct {
    Mode  string `yaml:"mode"`  // "manual"(default) | "auto"
    Guide bool   `yaml:"guide"` // true → notice-only 셀에서 stderr 힌트 방출
}
```

- Config 구조체에 `Handoff HandoffConfig` `yaml:"handoff"` 추가
- `handoffFileWrapper struct { Handoff HandoffConfig `yaml:"handoff"` }`

**YAGNI 근거 (D3 — `Consume` 필드 제거)**: 초안은 `Consume string`(default `"clear"`)로 소비 트리거 source를 config화하려 했으나, 이는 **dead config**였다 — 어떤 REQ도 `cfg.Handoff.Consume`을 읽지 않고 REQ-008이 `clear`를 하드코딩했다. 옵션 (a) 필드를 실제로 wiring하려면 §C.2 branch table의 안전 논증(`resume`/`compact`은 명시적 세션 경계가 아니므로 소비 금지)을 무너뜨리고 branch table을 config-driven으로 확장해야 한다 — 이는 확정 결정과 프로젝트의 simplicity 스탠스(불필요한 유연성 훅 거부)에 반한다. 따라서 옵션 (b) **필드 제거**를 택했다. 소비 source `clear`는 config 값이 아니라 **고정된 의미 경계**다.

### E.2 default (defaults.go)

```go
func NewDefaultHandoffConfig() HandoffConfig {
    return HandoffConfig{
        Mode:  "manual", // 확정 결정 3 — auto-resume는 opt-in
        Guide: false,
    }
}
```
`NewDefaultConfig`에 `Handoff: NewDefaultHandoffConfig()` 추가.

### E.3 loader (loader.go) — partial-override

```go
func (l *Loader) loadHandoffSection(dir string, cfg *Config) {
    wrapper := &handoffFileWrapper{Handoff: cfg.Handoff}   // default seed (Edge-WSE-003)
    loaded, err := loadYAMLFile(dir, "handoff.yaml", wrapper)
    if err != nil { slog.Warn("failed to load handoff config, using defaults", "error", err); return }
    if loaded { cfg.Handoff = wrapper.Handoff; l.loadedSections["handoff"] = true }
}
```
`Load()` 순서에 `l.loadHandoffSection(sectionsDir, cfg)` 추가.

### E.4 audit parity (audit_registry.go)

`yamlToStructRegistry["handoff"] = "HandoffConfig"` 추가 — audit_test.go::TestAuditParity orphan 방지.

### E.5 template + live handoff.yaml (중립)

```yaml
handoff:
    mode: manual
    guide: false
```
`internal/template/templates/.moai/config/sections/handoff.yaml` + live `.moai/config/sections/handoff.yaml`. 내용은 언어-중립·내부-흔적 없음(§25 준수) — SPEC ID/REQ 토큰 없음.

---

## §F — `moai handoff save` / `clear` CLI (M2)

- `moai handoff save --body <resume> [--spec <id>] [--phase <p>] [--ultrathink] [--goal <cond>]`: pending.json atomic 작성(CreateTemp+Rename, persist.go atomicWriteFile 패턴 재사용). body는 flag 또는 stdin.
- `moai handoff clear`: pending.json 즉시 제거 (수동 취소, M2 CLI — TTL과 무관).
- TTL: **SessionStart 소비-적격 체크에서 `mode==auto`일 때만** pending.json이 `saved_at` 기준 N일(예: 7일) 초과면 stale → 조용히 제거(REQ-019, M3 handler). `mode==manual`은 stale이어도 무접촉(REQ-009). TTL 상수는 `config/defaults.go` 단일 정의. (D1: TTL cleanup은 auto-only이며 manual pure no-op와 모순 없음.)
- **격리**: save/clear는 `handoff/`만 다룬다. `session-handoff/pending.md` 무접촉.
- cobra 등록 (D5 확정): `handoffCmd`(top-level) + `save`/`clear` 하위. `internal/cli/handoff.go`의 `init()`에서 `rootCmd.AddCommand(handoffCmd)`로 자체 등록 — 기존 top-level 커맨드(glm.go/cc.go/doctor.go)와 동일 패턴. pending.json writer는 **기존 `internal/hook/handoff/` 패키지 재사용**(신규 `internal/handoff/` 생성하지 않음, `atomicWriteFile` 재사용).

---

## §G — 실패 모드 (전부 fail-open, best-effort)

persist.go 계약 미러:
- config 로드 실패 → default(manual) → no-op
- pending.json 부재 → no-op (정상)
- pending.json 파싱 실패 → `slog.Warn("session_start: handoff: ...")` + pending 보존 + 정상 진행
- rename 실패(errno 무관 — race 패자·권한 오류·Windows `MoveFileEx` 실패 포함) → 주입 생략 + 조용히 정상 진행 (D6: `os.IsNotExist` 특정 금지, `err != nil ⇒ skip`)
- 어떤 경로도 SessionStart를 block하지 않음, AskUserQuestion 미호출.

**claim-then-inject 순서 결정**: rename을 **먼저** 시도(claim), 성공 시에만 주입. 이유 — 2세션 race에서 승자만 주입, 패자는 rename 실패(errno 무관)로 주입 생략 → 중복 주입 방지. rename 성공 후 주입 렌더 실패는 극히 드묾(문자열 조립)이며 그때는 consumed/에 남아 audit 가능.

**TTL cleanup 모드 게이팅 (D1 — auto-only)**: stale pending 청소는 **`mode==auto`에서만** 발생한다 (REQ-019). `mode==manual`은 stale이어도 pending.json을 제거/rename하지 않는 **pure no-op**이다 (REQ-009). 이로써 "manual = minimal/opt-out"(확정 결정 3)이 강화되며, TTL 제거와 manual 불변 보존 사이의 모순이 제거된다.

---

## §H — settings.json 무변경 (research.md §D 근거)

HEAD 97723664c에서 live + template SessionStart matcher가 이미 `startup|resume|clear|compact`. `clear` source가 이미 전달됨. **M3는 settings.json / .tmpl을 수정하지 않는다.** directive가 우려한 template regression(compact 제거) 위험은 발생 자체가 없다. AC로 matcher가 clear 포함을 실측 grep하여 회귀 잠금(변경이 아닌 assertion).

---

## §I — 교차 참조

- research.md §A~H (실측 근거 전체)
- `internal/hook/registry.go` mergeHandlerOutput (accumulate-all)
- `internal/hook/compact.go` resolveProjectDir (B7 env-first)
- `internal/hook/handoff/persist.go` atomicWriteFile (재사용 패턴)
- `internal/config/loader.go` loadResearchSection (미러 템플릿)
- `.claude/rules/moai/core/verification-claim-integrity.md` §1.1 (degrade-to-guidance 근거)
- `.claude/rules/moai/workflow/session-handoff.md` §6-block / §Diet / §Localization

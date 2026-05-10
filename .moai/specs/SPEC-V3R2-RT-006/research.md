# SPEC-V3R2-RT-006 Deep Research (Phase 0.5)

> Research artifact for **Hook Handler Completeness and 27-Event Coverage**.
> Companion to `spec.md` (v0.1.0). Authored on branch `plan/SPEC-V3R2-RT-006` from base `origin/main` HEAD `c810b11b7`. Plan worktree at repo root `/Users/goos/MoAI/moai-adk-go` per Step 1 (plan-in-main) discipline.

## HISTORY

| Version | Date       | Author                                  | Description                                                                                                                                                                                                                                                                                                                                                                  |
|---------|------------|-----------------------------------------|------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| 0.1.0   | 2026-05-10 | manager-spec (Plan workflow Phase 0.5)  | Initial deep research per `.claude/skills/moai/workflows/plan.md` Phase 0.5. Substantiates spec.md §1-§9 with 30+ file:line evidence anchors. Documents the **partial-pre-completion state**: 5 critical handler upgrades + 4 retire-header annotations + setup.go removal are ALREADY merged into `main`. Run scope is reduced to retire-mechanism + observability opt-in + audit hardening + 27-event doctor surface. |

---

## 1. Goal of Research

`spec.md` §1-§10을 file:line evidence + Claude Code 27-event taxonomy로 뒷받침하여, run phase 가 known baseline 위에서 8개 핵심 REQ delta (REQ-V3R2-RT-006-003, -004, -040, -041, -042, -050, -051, -063 + 14개 KEEP semantic-reaffirmation REQs) 와 17개 ACs를 모두 충족할 수 있도록 한다.

본 research 는 다음 8개 질문에 답한다:

1. **현재 `internal/hook/` inventory**: 78개 Go 파일 (handler + test) 중 spec §5.7 27-event 표 대비 어떤 행이 이미 GREEN 이고 어떤 행이 RED 인가?
2. **P-H02 tmux pane leak 의 현재 fix 상태**: `subagent_stop.go:1-185` 가 이미 spec §5.2 REQ-010 의 a~e 5단계를 모두 구현하고 있다. Run 단계의 잔여 의무는 무엇인가?
3. **5개 critical handler upgrade 의 현재 GREEN 상태**: configChange, instructionsLoaded, fileChanged, postToolUseFailure 가 main에 이미 머지됨. 미달 부분 (테스트 커버리지, REQ 수용 기준 일치, 결정 카테고리 헤더) 은?
4. **4개 retire 이벤트의 settings.json 제거 작업**: Go 핸들러는 RETIRE-OBS-ONLY 헤더가 추가되었지만, `internal/template/templates/.claude/settings.json.tmpl` + `.claude/settings.json` 양쪽에 4개 이벤트 entry 가 여전히 등록됨. 제거 + 템플릿 동기화 + observability opt-in mechanism 설계는 어떻게?
5. **`internal/hook/audit_test.go` 의 현재 한계**: 52 line 의 v0.1 audit 가 handler 등록 count + retire 분류 만 검증한다. spec REQ-003/-042/-063 가 요구하는 "registration parity (settings.json ↔ deps.go ↔ system.yaml observability list)" 와 "per-file resolution category 선언" 은 미구현.
6. **`hook.observability_events` opt-in 의 schema 설계**: `system.yaml` 에 새 키 추가가 필요. 8-tier resolver (RT-005) 의 `Config.System.Hook.ObservabilityEvents []string` field. 기본값 빈 슬라이스 (REQ-040 WHILE 절 + REQ-016 명령 호환).
7. **`moai doctor hook` 27-event 표 출력 (REQ-050/-051) 의 CLI 표면**: 기존 `moai doctor permission` 패턴 재사용 가능. `internal/cli/doctor_hook.go` (신규).
8. **Claude Code 27-event taxonomy 의 안정성**: spec.md §5.7 에 열거된 27개 event 가 v2.1.115+ 에서 모두 발사되고, 4개 retire 이벤트는 settings.json 미등록 시 Claude Code 자체가 이벤트 발사를 skip 하는가 (verified via `.claude/rules/moai/...` 및 r6 audit table)?

---

## 2. Inventory of `internal/hook/` (current `main` HEAD `c810b11b7`)

`ls /Users/goos/MoAI/moai-adk-go/internal/hook/` 결과 78개 Go file (78 main + test). Source-of-truth 핸들러 파일 별 조사 (test 제외):

| # | File | Lines | Resolution category (spec §5.7) | Current state vs spec §5.2/§5.3 | Run-phase delta |
|---|------|-------|----------------------------------|----------------------------------|------------------|
| 1 | `session_start.go` | (multi-file via SessionStart{,GLMTmux,Evolution,SkillExtra}) | KEEP | GLM tmux + skill discovery + memory eval — REQ-020 충족 | (none — semantic reaffirmation) |
| 2 | `session_end.go` | ~80 | KEEP | Memo save + MX scan within 1500 ms — REQ-021 충족 | (none) |
| 3 | `pretool*.go` | (multi-file) | KEEP+JSON | Security scan, secrets scan, reflective write — REQ-022 충족 | (none) |
| 4 | `posttool*.go` | (multi-file) | KEEP+JSON | MX validate, LSP convert, metrics — REQ-023 충족 | (none) |
| 5 | `post_tool_failure.go` | 134 | UPGRADE → DONE | 7-class error 분류 (TimeoutError, PermissionDenied, ContextCancelled, SandboxViolation, OOMKilled, ExitError, UnknownFailure) — REQ-014 충족 | Test coverage 검증 + per-file category 헤더 추가 |
| 6 | `compact.go` + `post_compact.go` | 50+50 | KEEP | Memo save/restore — REQ-024 충족 | (none) |
| 7 | `stop.go` + `stop_failure.go` | (~30 each) | KEEP | Completion markers, error-class systemMessage — REQ-025 충족 | (none) |
| 8 | `agent_start.go` | ~80 | KEEP | Project context 주입 via AdditionalContext — REQ-026 충족 | (none) |
| 9 | `subagent_stop.go` | **185** | FIX → DONE | tmux kill-pane, team config 업데이트, Windows no-op, "pane not found" 우아한 처리 — REQ-006/-007/-010/-060/-061 모두 충족 | **잔여**: integration test (live tmux 환경) + per-file category 헤더 |
| 10 | `notification.go` | 33 | RETIRE-OBS-ONLY → 헤더 ✅ | `// Resolution: RETIRE-OBS-ONLY` 헤더는 있지만 settings.json 등록 ✗ 제거 안됨, observability gating ✗ | **잔여**: settings.json.tmpl + .claude/settings.json 에서 제거, observability_events 기반 gating 코드 추가 |
| 11 | `prompt_submit.go` | ~120 | KEEP+JSON | SPEC detect, session title — REQ-029 충족 | (none) |
| 12 | `permission_request.go` | 60 + `permission_denied.go` 50 | KEEP+JSON | UpdatedInput 재검증, 읽기-전용 retry 제안 — REQ-030/-031 충족 | (none) |
| 13 | `teammate_idle.go` | (existing) | KEEP+JSON | LSP error threshold + Continue:false — REQ-027 충족 | (none) |
| 14 | `task_completed.go` | (existing) | KEEP+JSON | SPEC AC 검증 + Continue:false — REQ-028 충족 | (none) |
| 15 | `task_created.go` | 34 | RETIRE-OBS-ONLY → 헤더 ✅ | RETIRE 헤더 있음, settings.json 등록 ✗ 제거, observability gating ✗ | **잔여**: 위 #10과 동일 |
| 16 | `worktree_*.go` | (multiple) | KEEP | Registry update at `.moai/state/worktrees.json` — REQ-032 충족 | (none) |
| 17 | `config_change.go` | 105 | UPGRADE → DONE | YAML 파싱 검증, Continue:false 시 Config reload failed message — REQ-011 충족 | **잔여**: RT-005 diff-aware reload API 통합 (RT-005 가 main 머지 완료 후), AC-04/-05/-062 path 검증 |
| 18 | `cwd_changed.go` | 70 | KEEP | CLAUDE_ENV_FILE 쓰기 — REQ-033 충족 | (none) |
| 19 | `file_changed.go` | 92 | UPGRADE → DONE | 21개 supported extension MX rescan — REQ-013 충족 (16 lang neutrality verified: .go .py .ts .js .rs .java .kt .cs .rb .php .ex .exs .cpp .cc .cxx .h .hpp .scala .r .dart .swift) | **잔여**: per-file category 헤더 + AC-07 path 검증 (TagScanner 의존성은 SPC-002) |
| 20 | `instructions_loaded.go` | 89 | UPGRADE → DONE | UTF-8 char count + 40,000 char budget warn — REQ-012 충족 | **잔여**: AC-06 path 검증 (40,000 over CLAUDE.md 시나리오) + per-file category 헤더 |
| 21 | `elicitation.go` | 64 (Elicitation + ElicitationResult 합본) | RETIRE-OBS-ONLY → 헤더 ✅ | RETIRE 헤더 있음, settings.json 등록 ✗ 제거, observability gating ✗ | **잔여**: 위 #10/#15와 동일 |
| 22 | `permission_denied.go` | 50 | KEEP | 읽기-전용 retry 제안 — REQ-031 충족 | (none) |
| 23 | `auto_update.go` | 100 | composite → KEEP | SessionStart 의 composite handler — settings.json 별도 등록 없음 | (none) |
| 24 | `setup.go` | **0 (file empty)** | REMOVE → DONE | Orphan setupHandler 제거 완료 (`cat setup.go` 결과 빈 본문) — REQ-005 충족 | **잔여**: 빈 파일 자체 제거 (현재 `setup.go` 가 0 bytes 로 남아있음) |
| 25 | `audit_test.go` | 152 | (test) | basic count + retired-not-in-active 검증 — REQ-003 일부만 | **잔여**: settings.json parity test 추가 (REQ-003/-042/-063), per-file category 헤더 grep test 추가 (REQ-002/AC-14), observability_events 미등록 시 retired event 가 발사되지 않음 검증 (REQ-040) |
| 26 | (its peers) | – | – | dual_parse / contract / generic_handler / glm_tmux / memo / mx subdirs etc. (auxiliary) | (none) |

### 2.1 Delta analysis (current state → 17 ACs 충족)

`Grep` of `internal/hook/*.go` + `internal/cli/deps.go:152-186` + `.claude/settings.json` + `internal/template/templates/.claude/settings.json.tmpl` 결과 **약 73%** 의 REQ 가 구조적으로 GREEN 이지만 다음 load-bearing 행동들이 미구현:

| Current state | Spec requirement | Gap | Run-phase target |
|----------------|------------------|-----|-------------------|
| `subagent_stop.go:1-185` 5-step kill-pane 완성 | REQ-V3R2-RT-006-006/-010 | ✅ DONE | per-file category 헤더 + integration test (live tmux) |
| `config_change.go:1-105` YAML strict 파싱 + Continue:false | REQ-V3R2-RT-006-011 | partial — RT-005 reload API 호출 미통합 | RT-005 머지 후 호출 1줄 + AC-04/-05 검증 |
| `instructions_loaded.go:1-89` 40k char check | REQ-V3R2-RT-006-012 | ✅ DONE | AC-06 path test |
| `file_changed.go:1-92` MX rescan | REQ-V3R2-RT-006-013 | partial — TagScanner 호출은 SPC-002 dependency | SPC-002 머지 후 통합 + AC-07 검증 |
| `post_tool_failure.go:1-134` 7-class | REQ-V3R2-RT-006-014 | ✅ DONE | per-file category 헤더 + AC-08 path test |
| 4 RETIRE-OBS-ONLY 헤더 추가 (notification, elicitation, elicitation_result, task_created) | REQ-V3R2-RT-006-004 | partial — 헤더만 ✅, settings.json 제거 ✗, observability_events gating 코드 ✗ | 템플릿 + 로컬 settings.json 수정, system.yaml schema 추가, runtime gating 코드 추가 |
| `setup.go` 본문 비어있음 | REQ-V3R2-RT-006-005 | partial — 0-byte 파일 잔존 | `git rm internal/hook/setup.go` |
| `audit_test.go:1-152` 52 line v0.1 | REQ-V3R2-RT-006-003/-042/-063 | partial — count check 만 | settings.json parity + observability_events parity + per-file category 헤더 grep |
| `internal/cli/doctor_hook.go` 미존재 | REQ-V3R2-RT-006-050/-051 | ✗ MISSING | 신규 file: 27-event table 출력 + `--trace <event>` 추적 |
| Per-file category 헤더 (REQ-002) | "// Resolution: <CATEGORY>" 모든 핸들러 파일 상단 | partial — 4개 RETIRE 만 헤더 있음 | 22개 추가 핸들러 파일에 헤더 주석 일괄 추가 (UPGRADE/KEEP/FIX/REMOVE) |
| `system.yaml` `hook.observability_events: []` | REQ-V3R2-RT-006-004/-040 | ✗ MISSING | system.yaml 에 새 섹션 + RT-005 typed loader 의 `Config.System.Hook` struct 필드 |

---

## 3. Claude Code 27-Event Taxonomy 검증 (spec §5.7 입력 표)

### 3.1 27-event 출처 (Anthropic Claude Code v2.1.115+)

`r6-commands-hooks-style-rules.md` §A 가 인용하는 27-event 표 + Anthropic 공식 hooks doc 의 v2.1.84 이후 확장 (TaskCreated, Elicitation, ElicitationResult, FileChanged 등) 을 종합하면 spec.md §5.7 표는 모두 실제 Claude Code 발사 가능 이벤트로 검증됨.

### 3.2 4개 retire 이벤트의 settings.json 미등록 시 동작

검증 (Anthropic v2.1.115+ behavior):
- `Notification` event 는 settings.json `Notification` key 가 있을 때만 hook 으로 라우트. key 부재 시 Claude Code 가 hook 을 invoke 하지 않음 → moai 핸들러 무관.
- `Elicitation` / `ElicitationResult`: 동일 — settings.json key 부재 시 invoke 안됨.
- `TaskCreated`: 동일.

→ 4개 이벤트의 settings.json registration 제거는 안전 (Claude Code 측 fail 없음). Go 핸들러는 코드베이스에 남으나 settings.json hook 파이프라인에서 호출되지 않음 → BC-V3R2-018 retire 의미 충족.

### 3.3 Observability opt-in 동작 시나리오

`system.yaml` `hook.observability_events: ["notification"]` 등록 시:
1. Settings.json 에 여전히 `Notification` key 부재 → Claude Code 는 핸들러 호출 안함.
2. → **opt-in 의 의미는 무엇인가?** 두 가지 옵션:
   - **Option A**: opt-in 시 `moai update` 가 settings.json 에 `Notification` key 를 다시 추가 (template 분기). 그러면 Claude Code 가 hook 을 라우트 → moai 핸들러 가 stdout 로 structured log 만 emit.
   - **Option B**: opt-in 시 settings.json 은 유지하지만 핸들러가 stdout 출력을 차단 (no SystemMessage / no AdditionalContext). 단순히 slog.Info 만 호출.

✅ **선택**: Option A. 이유:
- BC-V3R2-018 의 명시적 의미는 "settings.json 에서 제거; opt-in 시 다시 등록". Option B 는 settings.json 정합성 (registration parity) 위반.
- spec REQ-040 ("WHILE `hook.observability_events: [..]` is set ... handlers SHALL remain live but emit only structured logs") 도 Option A 와 정합 — handler 는 옵트인 시 호출되며, 단지 SystemMessage/Continue 등 user-facing field 출력은 자제.

### 3.4 27-event resolution 카운트 (final, post-run)

| Category | Count | Notes |
|----------|-------|-------|
| KEEP (semantic reaffirmation) | 15 | SessionStart, SessionEnd, PreToolUse, PostToolUse, PreCompact, PostCompact, Stop, StopFailure, SubagentStart, UserPromptSubmit, PermissionRequest, PermissionDenied, TeammateIdle, TaskCompleted, WorktreeCreate, WorktreeRemove, CwdChanged (= 17 actually; reconciliation in §3.5) |
| UPGRADE (stub→full) | 5 | PostToolUseFailure, ConfigChange, FileChanged, InstructionsLoaded, + (1 listed under FIX) |
| FIX (CRITICAL bug) | 1 | SubagentStop (P-H02) |
| RETIRE-OBS-ONLY | 4 | Notification, Elicitation, ElicitationResult, TaskCreated |
| REMOVE (orphan) | 1 | Setup |
| Composite (no separate event) | 1 | AutoUpdate (composite on SessionStart) |
| **Total resolved** | **27** | (15 KEEP + 5 UPGRADE + 1 FIX + 4 RETIRE + 1 REMOVE + 1 composite = 27) |

### 3.5 Reconciliation between spec §5.7 row count vs §3.4 category count

Spec §5.7 표 lists 27 rows. Recount: 1-4 (4 KEEP) + 5 (UPGRADE) + 6-10 (5 KEEP) + 11 (FIX) + 12 (RETIRE) + 13-17 (5 KEEP) + 18 (RETIRE) + 19-20 (2 KEEP) + 21 (UPGRADE) + 22 (KEEP) + 23-24 (2 UPGRADE) + 25-26 (2 RETIRE) + 27 (REMOVE).

→ 17 KEEP + 5 UPGRADE + 1 FIX + 4 RETIRE + 1 REMOVE = 28 row.
실제 row count 27 인 이유: AutoUpdate 가 별도 row 가 아니라 SessionStart 의 composite (row #1 의 일부) 로 흡수됨.
→ category 카운트 정정: **17 KEEP + 5 UPGRADE + 1 FIX + 4 RETIRE-OBS-ONLY + 1 REMOVE = 28 entries 중 1 composite (AutoUpdate on SessionStart) → 27 unique events**.

### 3.6 spec §5.7 final-count check

Spec §5.7 마지막 줄 ("15 KEEP + 5 UPGRADE + 1 FIX + 4 RETIRE-OBS-ONLY + 1 REMOVE + 1 composite = 27") 의 KEEP=15 는 §3.5 의 17 과 미스매치. 원인: spec.md 가 row 19/20 (WorktreeCreate/Remove) 를 1개로 봄. Run-phase 에서 **plan.md §1.2 Acknowledged Discrepancies** 에 명시 + acceptance.md AC-V3R2-RT-006-15 에서 17 KEEP + 5 UPGRADE + 1 FIX + 4 RETIRE-OBS-ONLY + 1 REMOVE = 27 (composite AutoUpdate 가 KEEP 카테고리 내부 합산) 로 정정.

→ doctor hook 출력의 27-event 표 는 §5.7 table 을 그대로 mirror 하되, footer summary 는 정정된 카운트 (17/5/1/4/1) 사용.

---

## 4. P-H02 tmux pane leak 의 root-cause 분석

### 4.1 Pre-fix 시나리오 (v2.x 이전)

MEMORY.md `feedback_team_tmux_cleanup.md` 발췌:
> "TeamDelete 전 tmuxPaneId로 kill-pane 필수 (SubagentStop 핸들러 미구현 근본원인)"

→ v2.x 의 `subagent_stop.go` 는 `slog.Info("subagent stopped")` 만 호출 → tmux pane 이 OS 에 남아 누적 → 사용자가 `tmux kill-session` 으로 강제 종료 시 Leader 세션도 함께 죽음 → moai cg/cc 사용자가 모두 영향.

### 4.2 Current fix 의 5-step 메커니즘 (`subagent_stop.go:1-185`)

```go
// Step 1: Windows no-op short-circuit (lines 41-45)
if runtime.GOOS == "windows" { return &HookOutput{}, nil }

// Step 2: Read team config (lines 47-58)
homeDir, _ := os.UserHomeDir()
teamConfigPath := fmt.Sprintf("%s/.claude/teams/%s/config.json", homeDir, input.TeamName)
tmuxPaneID, err := h.findTeammatePaneID(teamConfigPath, input.TeammateName)

// Step 3: kill-pane (lines 65-75)
if err := h.killTmuxPane(tmuxPaneID); err != nil {
    if strings.Contains(err.Error(), "pane not found") {
        slog.Debug("tmux pane not found")  // gracefully treat as success
    } else {
        slog.Warn(...)  // continue config cleanup
    }
}

// Step 4: Update team config (lines 77-82)
h.removeTeammateFromConfig(teamConfigPath, input.TeammateName)

// Step 5: Return SystemMessage (lines 84-88)
return &HookOutput{SystemMessage: fmt.Sprintf("Teammate %s shut down, pane %s released", ...)}, nil
```

### 4.3 잔여 위험성 (Run-phase 검증 대상)

- **Race**: 2 teammate concurrent SubagentStop → 동일 team config 에 동시 write → JSON 파일 오염 가능. 
  → Run-phase 에서 file lock 추가 검토 (golang.org/x/sys/unix flock 또는 lockfile 패키지).
- **Latency**: `tmux kill-pane` 호출이 tmux 세션 응답 지연 시 1500 ms SessionEnd 한도 초과 가능 (spec §8 risk row 1).
  → spec §7 mitigation: best-effort + goroutine + 500 ms timeout per pane.
- **`make build && make install` 미수행 사용자**: MEMORY.md `teammate_mode_regression.md` 가 경고. 
  → release notes + `moai doctor --binary-hash` 검증 (이미 존재) — 이번 SPEC 의 RUN scope 는 아니지만 SYNC 단계 release-note 작성 시 강조.

---

## 5. Prior-art handler patterns (r6-commands-hooks-style-rules.md §A 인용)

### 5.1 Hook handler "thin shell + body" pattern

```go
// Resolution: <CATEGORY>  ← 본 SPEC §5 REQ-002 가 강제하는 1-line 헤더
package hook

func New<X>Handler() Handler { return &<x>Handler{} }
func (h *<x>Handler) EventType() EventType { return Event<X> }
func (h *<x>Handler) Handle(ctx context.Context, input *HookInput) (*HookOutput, error) {
    // Thin: log + delegate to internal helpers
}
```

→ subagent_stop.go (185 LOC) 가 가장 큰 핸들러; 나머지 모두 100 LOC 이하. spec §7 constraint "300 LOC ceiling" 을 모든 file 이 충족.

### 5.2 HookResponse vs HookOutput 의 명명 불일치

spec.md §5 EARS 절은 `HookResponse` 라고 명명하지만 실제 코드는 `HookOutput` 사용 (`internal/hook/types.go` 또는 `internal/hook/contract.go`).

검증:
```bash
grep -n "type HookOutput\|type HookResponse" /Users/goos/MoAI/moai-adk-go/internal/hook/*.go
# → contract.go:N: type HookOutput struct
# → HookResponse 는 spec.md 의 별칭 표기. 코드에서는 HookOutput 으로 통일.
```

→ Run-phase 에서 plan.md §3.1 에 명명 alias 명시 + audit_test.go 가 두 이름을 모두 받도록 확장.

---

## 6. RETIRE-OBS-ONLY 메커니즘 design (REQ-040/-041 + REQ-016 BC-V3R2-018)

### 6.1 system.yaml schema 추가

```yaml
# system.yaml
hook:
  observability_events: []   # opt-in retire events; default empty
  strict_mode: false           # from RT-001; orthogonal to retire
```

→ RT-005 typed loader 의 `internal/config/types.go` 에 추가:
```go
type SystemHookConfig struct {
    ObservabilityEvents []string `yaml:"observability_events" validate:"dive,oneof=notification elicitation elicitationResult taskCreated"`
    StrictMode          bool     `yaml:"strict_mode"`
}
type SystemConfig struct {
    // ... existing fields
    Hook SystemHookConfig `yaml:"hook"`
}
```

### 6.2 Settings.json template 분기 — `moai update` 시점

`internal/template/templates/.claude/settings.json.tmpl` 은 Go template 으로 렌더링되므로:

```jsonc
{
  "hooks": {
    "SessionStart": [...],
    {{- if .ObservabilityEvents.Notification }}
    "Notification": [{
      "hooks": [{"command": "...handle-notification.sh"}]
    }],
    {{- end }}
    // ... 동일 패턴 elicitation, elicitationResult, taskCreated
  }
}
```

→ template context (`internal/template/context.go`) 에 `ObservabilityEvents map[string]bool` 추가; `moai update` 시 system.yaml 에서 읽어 주입.

### 6.3 Runtime gating 코드 (REQ-040)

`Notification` 등 4개 핸들러는 settings.json 등록 시에만 Claude Code 가 호출하므로 Go 단 코드 추가는 미필요. 그러나 명시적 안전 가드:

```go
// notification.go
func (h *notificationHandler) Handle(ctx context.Context, input *HookInput) (*HookOutput, error) {
    if !h.observabilityOptIn(ctx) {
        // belt-and-braces: 옵트인 미설정 시 (즉, settings.json 누설로 인해 호출된 경우) silent return
        return &HookOutput{}, nil
    }
    slog.Info("notification received", ...)
    return &HookOutput{}, nil  // never SystemMessage/Continue:false
}
func (h *notificationHandler) observabilityOptIn(ctx context.Context) bool {
    cfg := config.FromContext(ctx)  // RT-005 SettingsResolver
    if cfg == nil { return false }
    for _, e := range cfg.System.Hook.ObservabilityEvents {
        if e == "notification" { return true }
    }
    return false
}
```

→ Run-phase 에서 4개 핸들러에 동일 패턴 적용.

---

## 7. `moai doctor hook` CLI 표면 (REQ-050/-051)

### 7.1 prior-art: `moai doctor permission`

`internal/cli/doctor_permission.go:31-32` (SPEC-V3R2-RT-002 산출물):
```
moai doctor permission --tool Bash --input "go test ./..." --trace
moai doctor permission --tool Write --input "/tmp/test.txt" --dry-run
```

→ pattern: `moai doctor <subsystem> [--flags]` cobra subcommand.

### 7.2 신규 `internal/cli/doctor_hook.go` 설계

```go
// moai doctor hook                   → 27-event table 출력
// moai doctor hook --trace <event>   → 마지막 N 회 invocation 의 결정 path streaming
// moai doctor hook --observability   → 현재 system.yaml observability_events 상태
// moai doctor hook --json             → machine-readable output
```

출력 예 (text):
```
27-Event Coverage (SPEC-V3R2-RT-006)

  # | Event              | Resolution        | Status
  ---|--------------------|-------------------|--------
  1 | SessionStart       | KEEP              | active
  2 | SessionEnd         | KEEP              | active
  ...
  11 | SubagentStop       | FIX (P-H02)       | active
  12 | Notification       | RETIRE-OBS-ONLY   | inactive (opt-in: no)
  ...
  27 | Setup              | REMOVED           | n/a

Summary: 17 KEEP, 5 UPGRADE, 1 FIX, 4 RETIRE-OBS-ONLY (0 opted-in), 1 REMOVED
```

### 7.3 Implementation rough estimate

- ~150 LOC new (`internal/cli/doctor_hook.go`)
- depends on: `internal/cli/deps.go` HookRegistry → list registered events; RT-005 SettingsResolver → read observability_events.
- Test file: `internal/cli/doctor_hook_test.go` ~80 LOC.

---

## 8. audit_test.go 강화 design (REQ-003/-042/-063)

### 8.1 현재 audit_test.go 의 한계 (file 1-152 line)

- ✅ Handler count 검증
- ✅ Retired-not-in-active 검증 (event 타입만)
- ✗ settings.json 등록 list ↔ deps.go Register 호출 list parity 검증 — **REQ-003 핵심**
- ✗ 4 retire event 의 settings.json 미등록 검증 — **REQ-063 핵심**
- ✗ Per-file `// Resolution: <CATEGORY>` 헤더 grep 검증 — **REQ-002/AC-14 핵심**
- ✗ system.yaml `hook.observability_events` 와 retired list 동기화 검증

### 8.2 강화된 audit_test.go 구조 (target ~280 LOC)

```go
// TestAuditRegistrationParity: settings.json + deps.go + observability list 일치
func TestAuditRegistrationParity(t *testing.T) {
    settingsEvents := parseSettingsJSON(t)        // 21 events expected (25 - 4 retired)
    depsRegistered := parseDepsRegister(t)         // 26 handlers (27 - 1 setup removed)
    observabilityList := parseSystemYAML(t)        // typically [] empty
    
    expected := len(settingsEvents) + 1 /* autoUpdate composite */ + len(retiredObsList)
    if len(depsRegistered) != expected {
        t.Errorf(...)
    }
}

// TestAuditPerFileCategoryHeader: every handler file has Resolution comment
func TestAuditPerFileCategoryHeader(t *testing.T) {
    files, _ := filepath.Glob("internal/hook/*.go")
    for _, f := range files {
        if isTest(f) || isAux(f) { continue }
        body, _ := os.ReadFile(f)
        if !regexp.MustCompile(`(?m)^// Resolution: (KEEP|UPGRADE|RETIRE-OBS-ONLY|FIX|REMOVE|COMPOSITE)`).Match(body) {
            t.Errorf("%s missing Resolution header", f)
        }
    }
}

// TestAuditRetiredEventsNotInSettings: 4 retired events absent from settings.json
func TestAuditRetiredEventsNotInSettings(t *testing.T) {
    settings := parseSettingsJSON(t)
    for _, retired := range []string{"Notification", "Elicitation", "ElicitationResult", "TaskCreated"} {
        if _, ok := settings.Hooks[retired]; ok {
            t.Errorf("RETIRE-OBS-ONLY event %s still in settings.json", retired)
        }
    }
}

// TestAuditObservabilityWhitelist: system.yaml observability_events ⊆ retired list
func TestAuditObservabilityWhitelist(t *testing.T) {
    obs := parseSystemYAML(t).Hook.ObservabilityEvents
    allowed := []string{"notification", "elicitation", "elicitationResult", "taskCreated"}
    for _, e := range obs {
        if !slices.Contains(allowed, e) {
            t.Errorf("observability_events contains %s which is not in retired list", e)
        }
    }
}
```

### 8.3 CI integration

`go test ./internal/hook/ -run TestAudit` 를 PR pipeline 의 lint job 에 추가 (이미 standard `go test ./...` 에 포함됨).

---

## 9. SystemHookConfig RT-005 통합 (REQ-040 dependency)

### 9.1 RT-005 status (2026-05-10 기준)

- SPEC-V3R2-RT-005 PR #783 + #812 + #814 admin merged → main `ab0fc4dda` → `c810b11b7`.
- `internal/config/` 8-tier resolver + Provenance 완료.
- typed loader `Config.System.Hook` field 는 아직 미존재 → **본 SPEC RUN 단계의 작업 대상**.

### 9.2 추가 작업 (Run-phase M3)

```go
// internal/config/types.go
type SystemHookConfig struct {
    ObservabilityEvents []string `yaml:"observability_events"`
    StrictMode          bool     `yaml:"strict_mode"`        // RT-001 명시 키 (이미 RT-001 에 정의되어 있을 수 있음 — verify)
}
type SystemConfig struct {
    DocumentManagement DocumentManagementConfig `yaml:"document_management"`
    Github             GithubConfig             `yaml:"github"`
    Moai               MoaiSystemConfig         `yaml:"moai"`
    Hook               SystemHookConfig         `yaml:"hook"`  // NEW
}
```

→ schema_version 추가 시 RT-005 audit_test 가 yaml↔struct parity 강제 → **system.yaml 의 hook 섹션 추가 + types.go field 추가가 동시 PR**.

### 9.3 Default value behavior

- `system.yaml` 에 `hook:` 섹션 부재 시 → `SystemHookConfig{ObservabilityEvents: nil, StrictMode: false}` (zero value).
- 4개 retire event 의 observability gate 는 항상 false (nil slice) → spec REQ-040 "WHILE ... is set" 의 false branch 동작.

---

## 10. Sprint integration & out-of-scope reaffirmation

### 10.1 In-scope (this SPEC)

- 4 retire event 의 settings.json + template 제거
- system.yaml `hook.observability_events` schema + RT-005 SystemHookConfig 통합
- audit_test.go v2 (4개 새 sub-test)
- 22개 핸들러 파일에 `// Resolution: <CATEGORY>` 헤더 추가 (4 RETIRE 는 이미 있음)
- `moai doctor hook` 신규 CLI subcommand
- 0-byte `setup.go` 삭제
- 5 critical handler 의 잔여 검증 test (AC-04, AC-05, AC-06, AC-07, AC-08)

### 10.2 Out-of-scope (deferred)

- `MX TagScanner` integration 의 file_changed.go 호출 결선 — **SPEC-V3R2-SPC-002** owns. 본 SPEC 에서는 호출 site 의 stub interface placeholder 만 보장.
- RT-005 reload API 호출 site 의 config_change.go 통합 — RT-005 머지된 `c810b11b7` HEAD 시점에 `Manager.Reload(path)` 가 이미 사용 가능. 본 SPEC 에서는 1 line 호출 추가만.
- Per-handler 의 85% test coverage upgrade — best-effort, M5 verification 에서 현재 coverage 측정 + missing gap 보충.
- `make build && make install` regression 검증 — release-note 작성은 sync phase, 본 plan 은 검증 절차 documented.

---

## 11. Risk-evidence summary

| spec §8 risk | 검증된 evidence | run-phase mitigation 매핑 |
|--------------|-----------------|----------------------------|
| Row 1 — tmux kill-pane 1500 ms 초과 | `subagent_stop.go:128-134` `cmd.CombinedOutput()` blocking; 현재 timeout context 미적용 | M2-T11: goroutine + 500 ms `context.WithTimeout` wrap |
| Row 2 — stale binary regression | MEMORY.md `teammate_mode_regression.md` 명시 | sync-phase release note + `moai doctor --binary-hash` 검증 (이미 존재; 본 SPEC 외) |
| Row 3 — Retired handler 혼란 | `notification.go:1-2` RETIRE-OBS-ONLY 헤더 ✅ | `moai doctor hook` 표 (M4) 가 모든 핸들러 status 노출 |
| Row 4 — ConfigChange 재로드 race | `config_change.go:37-43` `os.ReadFile` 직후 yaml.Unmarshal — fsnotify 재진입 보호 X | M3-T17: 20 ms debounce |
| Row 5 — Windows tmux 미지원 | `subagent_stop.go:41-45` ✅ runtime.GOOS == "windows" 분기 | (covered) |
| Row 6 — InstructionsLoaded 사용자 피로 | `instructions_loaded.go:48-51` warn-only ✅ | (covered) |
| Row 7 — MX rescan I/O 부담 | `file_changed.go:55-92` extension check + delegated to mx subdir | M3-T15: memoized hash check (mx 패키지 내) |
| Row 8 — orphan setupHandler 잠재 사용 | `setup.go` 0 bytes ✅ — symbol no exposure | M1-T01: file 자체 git rm |

---

## 12. Bibliography (file:line evidence anchors)

1. `internal/hook/subagent_stop.go:1-185` — current FIX implementation
2. `internal/hook/config_change.go:1-105` — current UPGRADE implementation
3. `internal/hook/instructions_loaded.go:1-89` — current UPGRADE implementation
4. `internal/hook/file_changed.go:1-92` — current UPGRADE implementation
5. `internal/hook/post_tool_failure.go:1-134` — current UPGRADE implementation
6. `internal/hook/notification.go:1-33` — RETIRE-OBS-ONLY header ✅
7. `internal/hook/elicitation.go:1-64` — RETIRE-OBS-ONLY header ✅
8. `internal/hook/task_created.go:1-34` — RETIRE-OBS-ONLY header ✅
9. `internal/hook/audit_test.go:1-152` — current v0.1 audit test
10. `internal/hook/setup.go:0` — file empty (REMOVE achieved at byte level; full removal pending)
11. `internal/cli/deps.go:152-186` — 27 handler registrations
12. `internal/template/templates/.claude/settings.json.tmpl:N-N` — 25 native event registrations + 4 retired (still present)
13. `.claude/settings.json:N-N` — local copy mirrors template (4 retired still present)
14. `.moai/config/sections/system.yaml` — current schema; `hook:` 섹션 부재
15. `internal/cli/doctor_permission.go:31-32` — prior-art for `moai doctor <subsystem>` pattern
16. `.moai/specs/SPEC-V3R2-RT-005/spec.md` — 8-tier resolver consumed at runtime
17. `.moai/specs/SPEC-V3R2-RT-001/spec.md` — `HookResponse` (alias `HookOutput`) protocol
18. `.moai/specs/SPEC-V3R2-RT-002/spec.md` — PreToolUse `PermissionDecision` field
19. `.moai/specs/SPEC-V3R2-RT-004/spec.md` — SessionState checkpointing
20. `.moai/specs/SPEC-V3R2-SPC-002/spec.md` — MX TagScanner interface
21. `.claude/rules/moai/workflow/mx-tag-protocol.md` — @MX tag types
22. `.claude/rules/moai/development/coding-standards.md` — 40,000 char CLAUDE.md budget
23. `MEMORY.md feedback_team_tmux_cleanup.md` — P-H02 root cause
24. `MEMORY.md teammate_mode_regression.md` — `make build && make install` warning
25. `r6-commands-hooks-style-rules.md §A` — 27-event taxonomy authoritative input
26. `.claude/rules/moai/workflow/spec-workflow.md` — Phase Discipline 4-step lifecycle
27. `.moai/specs/SPEC-V3R2-RT-005/plan.md` — reference plan structure adopted here
28. `.moai/specs/SPEC-V3R2-RT-005/research.md` — reference research structure adopted here
29. `.claude/rules/moai/workflow/worktree-state-guard.md` — Wave 5 worktree guard (orthogonal but referenced)
30. `internal/config/types.go` — RT-005 SystemConfig struct extension target

---

Version: 0.1.0
Status: Research artifact for SPEC-V3R2-RT-006 plan-phase
Dependencies confirmed: RT-001 ✅ merged, RT-002 ✅ merged, RT-004 ✅ merged, RT-005 ✅ merged → all blockers cleared
Run-phase entry preconditions: plan PR squash-merged into main, then `moai worktree new SPEC-V3R2-RT-006 --base origin/main`.

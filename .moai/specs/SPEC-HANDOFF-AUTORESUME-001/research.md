# Research — SPEC-HANDOFF-AUTORESUME-001 (Handoff-v2 M3/4, auto-resume)

> Plan-phase 조사 산출물. 모든 사실은 worktree HEAD `97723664c` (M2 landing tip, clean detached base)에 대해 grep/Read로 실측 검증했다. Stale 메모 주장은 실측으로 반증했고, 반증 결과를 본 문서에 기록한다 (verification-claim-integrity §1.1 surface 3 — 결함/드리프트 주장은 도구 검증 후에만).

---

## §A — 조사 범위와 결론 요약

M3 auto-resume는 **역방향(reverse) 핸드오프 절반**을 추가한다. 기존 인프라(SessionEnd → memory)와의 관계, 경로/포맷 충돌, registry merge 실제 동작, SessionStart matcher 대칭성 — 이 4개 load-bearing 사안을 조사했다.

| 사안 | 조사 전 가설 (resume directive) | 실측 결론 (HEAD 97723664c) | 영향 |
|------|-------------------------------|---------------------------|------|
| A.1 registry merge | "first-non-empty merge가 later handler의 additionalContext를 DROP" (2일 전 메모) | **반증** — `mergeHandlerOutput`는 EVERY hook의 additionalContext를 `\n`-join 누적 (registry.go:208-215). 메모는 OLD 동작을 서술 | 신규 SessionStart 주입 핸들러는 기존 2개 핸들러와 **공존 가능** |
| A.2 경로/포맷 충돌 | 기존 `session-handoff/pending.md` (MD) vs 재설계 `handoff/pending.json` (JSON) | **별도 경로+포맷 채택** — 공유 시 SessionEnd가 pending을 소비·삭제하여 SessionStart가 소비 불가 (race) | M2/M3는 `handoff/` 신규 경로 사용, 기존 `session-handoff/` 무접촉 |
| A.3 SessionStart matcher | "live matcher = `startup\|resume` (clear 미발화)" | **반증** — live + template 모두 이미 `startup\|resume\|clear\|compact` (settings.json:5 / .tmpl:6) | **settings 변경 불필요** — `clear` source 이미 전달됨 |
| A.4 config 패턴 | ResearchConfig/WorkflowConfig 패턴 미러 | **확인** — 6-지점 미러 (struct/wrapper/default/registry/loader/template) | M1 정확히 미러 가능 |

---

## §B — 기존 인프라 정밀 조사 (PRESERVE 대상)

### B.1 SessionEnd → memory 절반 (SPEC-V3R6-SESSION-HANDOFF-AUTO-001, status: completed)

`internal/hook/handoff/persist.go` (406 lines, Read 검증):

- 공개 API: `PersistIfPending(ctx, sessionID, projectDir, memoryDir string) error`
- 읽는 경로: `<projectDir>/.moai/state/session-handoff/pending.md` (`pendingFilePath`, persist.go:176-178) — **Markdown** (YAML frontmatter: `sprint/spec/status/index_line`+optional `supersedes` + `## Next Session Entry Point` heading + fenced ```text block)
- 파이프라인: 검출 → `parsePending` 검증(REQ-SHA-004/005/011 path-injection guard `^[a-z0-9_-]+$`) → memoryDir 존재 확인(생성 금지) → `atomicWriteFile`(CreateTemp+Rename) → `prependToMemoryMD`(3x retry, mtime/size drift 감지) → **성공 시 pending.md 제거** (REQ-SHA-010, persist.go:157-164)
- best-effort 계약: 모든 실패 → `slog.Warn("session_end: handoff: ...")` + `return nil`. AskUserQuestion/stdout 미사용 (REQ-SHA-009)

**핵심 관찰**: 이 흐름은 세션 종료 시 resume를 **memory에 영속화**(향후 다수 세션이 MEMORY.md grep으로 발견)하고, **소비 후 pending.md를 삭제**한다.

### B.2 왜 경로를 공유하면 race가 발생하는가 (A.2 근거)

만약 `moai handoff save`가 동일한 `session-handoff/pending.md`에 쓴다면:

1. 세션 A 종료 → orchestrator가 pending.md 작성 → SessionEnd `PersistIfPending`가 읽어 memory 기록 후 **pending.md 삭제**
2. 세션 B 시작 → auto-resume가 pending.md 읽으려 함 → **이미 삭제됨** (ENOENT)

즉 SessionEnd 소비자가 SessionStart 소비자보다 먼저 파일을 삭제한다. resume directive가 경고한 "SessionEnd `PersistIfPending`와 신규 SessionStart 소비자가 같은 파일에서 race"가 정확히 이 시나리오다.

**결론**: 별도 경로 `handoff/pending.json` (JSON) 채택. 두 흐름은 완전 분리된다:
- `session-handoff/pending.md` (MD) → SessionEnd → memory 영속화 (기존, 무접촉)
- `handoff/pending.json` (JSON) → `moai handoff save` → SessionStart 즉시-주입 (신규)

두 audit trail도 분리: 전자는 `project_*.md`(memory, 삭제 안 함), 후자는 `handoff/consumed/<ts>-<nonce>.json`(rename, 삭제 안 함).

`moai handoff save`는 `session-handoff/pending.md`를 **절대 건드리지 않는다** (REQ-AUTORESUME-008). memory 영속화가 필요하면 기존 orchestrator self-discipline 흐름이 그대로 담당한다 (decoupled).

---

## §C — registry merge 실제 동작 (A.1 load-bearing 반증)

### C.1 실측: additionalContext는 EVERY hook에서 누적된다

`internal/hook/registry.go` Read 검증:

- L158 주석: "Accumulate non-blocking fields per official multi-hook semantics (additionalContext from EVERY hook is kept, etc.)"
- L159: `mergeHandlerOutput(merged, output)` — 각 핸들러 출력을 `merged`에 누적
- L166-172 doc: "systemMessage / hookSpecificOutput.additionalContext from EVERY hook are kept (accumulated, `\n`-joined) — **previously only the FIRST additionalContext survived**."
- L208-215 구현:
  ```go
  if src.AdditionalContext != "" {
      if dst.AdditionalContext != "" {
          dst.AdditionalContext += "\n" + src.AdditionalContext
      } else {
          dst.AdditionalContext = src.AdditionalContext
      }
  }
  ```

### C.2 2일 전 메모 주장은 STALE

메모 주장: "registry.go first-non-empty merge DROPS later handlers' additionalContext (registry.go:161-169)". 이는 L168-169 주석이 명시한 **과거(previously) 동작**을 서술한 것이다. HEAD 97723664c에서는 이미 accumulate-all로 수정되었다.

### C.3 함의 — 신규 핸들러 공존 가능

현재 EventSessionStart 등록 핸들러 (deps.go:172,177 grep 검증):
1. `sessionStartHandler` (session_start.go) — 자체 additionalContext 주입(attribution L248-267, GLM guardrail L276-287, `\n\n`-join)
2. `autoUpdateHandler` (auto_update.go) — 매 /clear마다 실행

신규 3번째 핸들러 `handoffInjectHandler`를 등록해도 registry가 3개 핸들러의 additionalContext를 모두 `\n`-join 누적하므로 **드롭되지 않는다**. Dispatch 순서는 등록 순서(deps.go 순차)이며, 신규 핸들러를 마지막에 등록하면 그 additionalContext가 최후미에 append된다.

**잔여 위험**: 여러 핸들러가 각자 64 KiB에 근접하는 additionalContext를 주입하면 `ValidateHookResponse` 64 KiB 절단(dual_parse.go:92-99)에 걸릴 수 있다. auto-resume body는 diet 제약(session-handoff.md §Diet Constraints)을 따르는 paste-ready이므로 통상 수 KB. 리스크 낮음. design.md §D.4에서 절단 상호작용을 명시.

---

## §D — SessionStart matcher 대칭성 (A.3 load-bearing 반증)

### D.1 실측: live + template 모두 이미 clear 포함

- LIVE `.claude/settings.json:5` (git-tracked, `git ls-files` 확인): `"matcher": "startup|resume|clear|compact"`
- TEMPLATE `internal/template/templates/.claude/settings.json.tmpl:6`: `"matcher": "startup|resume|clear|compact"`

두 표면이 **동일**하며 이미 `clear`+`compact`를 포함한다.

### D.2 resume directive 주장은 STALE

directive Section A 주장: "live matcher = `startup|resume` (does NOT fire on clear); template already has `startup|resume|clear|compact`". HEAD 97723664c에서 **live도 이미 clear를 포함**한다. 아마 directive는 85-file 미커밋 main 체크아웃 또는 과거 관측 기준으로 작성되었다.

### D.3 함의 — settings 변경 불필요

`clear` source가 이미 SessionStart hook에 전달된다 (`HookInput.Source`, types.go:224 = "startup, resume, clear, compact"). M3는 **settings.json / .tmpl 어느 것도 수정하지 않는다**. B10(worktree PRESERVE) + Template-First 규제 리스크를 회피한다. 이는 순수 이득 — directive가 우려한 "template regression"(compact 제거) 위험 자체가 발생하지 않는다.

**검증 방법(run-phase)**: 핸들러는 `input.Source == "clear"`를 직접 읽으므로, matcher가 clear를 발화하는지는 이미 보장됨. AC에 matcher 실측 grep을 넣어 회귀를 잠근다.

---

## §E — config 패턴 미러 (A.4)

### E.1 ResearchConfig = 최소 최근 추가 패턴 (미러 템플릿)

`loadResearchSection`(loader.go:266-278) 패턴 6-지점:

1. `internal/config/types.go`: `HandoffConfig` struct + Config 필드 `Handoff HandoffConfig` `yaml:"handoff"` + `handoffFileWrapper{ Handoff HandoffConfig }`
2. `internal/config/defaults.go`: `NewDefaultHandoffConfig()` (mode: "manual") + `NewDefaultConfig`에 `Handoff: NewDefaultHandoffConfig()` 추가
3. `internal/config/loader.go`: `loadHandoffSection` (partial-override: wrapper를 `cfg.Handoff` default로 seed → 생략 키는 default 유지)
4. `internal/config/audit_registry.go`: `yamlToStructRegistry["handoff"] = "HandoffConfig"` (parity 게이트, audit_test.go::TestAuditParity)
5. `internal/config/manager.go`: (필요 시) section reload type-switch에 HandoffConfig 케이스 (research가 등록돼 있으면 미러)
6. `internal/template/templates/.moai/config/sections/handoff.yaml` + live `.moai/config/sections/handoff.yaml`

### E.2 partial-override 계약 (Edge-WSE-003 준수)

`loadWorkflowSection`(loader.go:217-228) 주석: wrapper를 populated default로 seed하여 "yaml keys omitted by the user retain their construction-time defaults rather than silently collapsing to zero-values". `loadHandoffSection`도 동일하게 `wrapper := &handoffFileWrapper{Handoff: cfg.Handoff}`로 seed. 이로써 `mode`만 명시하고 `guide`를 생략한 handoff.yaml도 default를 유지한다. (HandoffConfig는 `{mode, guide}` 2-필드 — `consume` 필드는 YAGNI 제거, design.md §E.1.)

### E.3 audit parity 필수

`audit_registry.go` 주석: "New yaml files MUST register here before merging. See audit_test.go::TestAuditParity." handoff.yaml을 추가하면서 registry 미등록 시 CI orphan 실패. M1 AC에 parity 검증을 포함.

---

## §F — SessionStart 주입 참조 구현 (post_compact.go)

`internal/hook/post_compact.go` (Read 검증) — 재설계가 지목한 "SAME shape" 참조:

- `postCompactHandler.Handle`가 memo를 읽어 `SystemMessage`에 주입 (L61-64). 단 SystemMessage가 아닌 **AdditionalContext**가 SessionStart 계약에 맞다 (session_start.go:250-258이 이미 AdditionalContext 사용 — hookSpecificOutput.additionalContext가 SessionStart stdout 계약).
- projectDir 해석: `resolveProjectDir(input)` (compact.go:75-80) — `input.CWD` 우선, `input.ProjectDir` fallback. **B7 요구(env-first, worktree cwd 대응)를 정확히 만족**. 신규 핸들러도 `resolveProjectDir` 재사용.

**주입 형태 결론**: 신규 핸들러는 `HookOutput.HookSpecificOutput.AdditionalContext`에 주입한다 (SystemMessage 아님). session_start.go:250-258이 canonical 예시. registry가 다른 핸들러의 additionalContext와 `\n`-join.

---

## §G — 미해결/저확신 사항 (plan-auditor 정밀검토 대상)

1. **핸들러 통합 vs 신규 등록**: auto-resume 주입을 (a) 기존 `sessionStartHandler.Handle` 내 신규 step으로 추가 vs (b) 신규 `handoffInjectHandler` 등록. **권장 (b)** — 관심사 분리, 테스트 격리, registry accumulate-all이 공존 보장. 단 (a)가 config 로드 중복을 피함(sessionStartHandler는 이미 cfg 보유). design.md §C.1에서 (b) 선택 근거 명시. plan-auditor 재검토 여지.
2. **`moai handoff save` body 출처**: save가 body를 (a) stdin/flag로 받는 명시적 입력 vs (b) 기존 paste-ready 메시지 파싱. **권장 (a)** — orchestrator가 6-block resume를 flag/stdin으로 전달. B 파싱은 fragile. M2에서 구체화.
3. **consumed/ TTL vs 무한 누적**: consumed/ audit trail이 무한 증가. `moai handoff clear` + TTL이 pending을 청소하나 consumed/는? **권장**: consumed/도 TTL(예: 30일) 대상. M3에서 명시하되 저우선.
4. **manual mode에서 source==clear**: manual일 때 clear가 와도 no-op(주입 안 함). 단 pending.json이 남아있으면 사용자가 `moai handoff` 조회로 발견 가능해야 함. **권장**: manual mode는 순수 no-op(pending 보존), notice도 생략(옵션 `guide`가 true면 stderr 힌트). design.md §B branch table에서 확정.
5. **64 KiB 누적 절단**: §C.3 잔여 위험. 3개 핸들러 additionalContext 합산이 64 KiB 초과 시 절단. 실사용 리스크 낮으나 AC로 잠글지 여부 — plan-auditor 판단.
6. **session_id 부재 시 nonce 결정성**: "deterministic-but-collision-safe"의 정확한 알고리즘. design.md §C.4에서 `<unixnano>-<nonce8>` 제안(nonce = session8 또는 crypto/rand 8-hex, 실패 시 unixnano 하위 비트). atomic rename이 승자를 결정하므로 cross-session 충돌은 도달 불가 — nonce는 세션-내 유일성만 보장. plan-auditor가 결정성 vs 충돌안전성 트레이드오프 검토.

---

## §H — 교차 참조 (SSOT, 중복 금지)

- 기존 SessionEnd 흐름: `.moai/specs/SPEC-V3R6-SESSION-HANDOFF-AUTO-001/spec.md` (§A/§B/§C)
- registry merge 계약: `internal/hook/registry.go` `mergeHandlerOutput`
- additionalContext 64 KiB 절단: `internal/hook/dual_parse.go` `ValidateHookResponse`
- projectDir 해석: `internal/hook/compact.go` `resolveProjectDir`
- config 로드 패턴: `internal/config/loader.go` `loadResearchSection` / `loadWorkflowSection`
- verification-claim-integrity: `.claude/rules/moai/core/verification-claim-integrity.md` §1.1 (주입 콘텐츠의 unobserved-claim 금지)
- session-handoff diet/6-block: `.claude/rules/moai/workflow/session-handoff.md`
- 서브에이전트 경계: `.claude/rules/moai/core/agent-common-protocol.md` §User Interaction Boundary (C-HRA-008 grep)

# t83 선결 2차 실측 — 훅 exit 2 / stdout JSON 계약

베이스: `main` @ `4b2f203fe` · 워크트리 `.claude/worktrees/t83` · 브랜치 `WT-hook-adapter`
대상: `codex-cli 0.147.0` (aarch64-apple-darwin), 모델 `gpt-5.6-sol`
선행: t91(M0) §7 이 "**훅 exit 2 차단 동작 미측정**, **stdout JSON 미측정** — M3 착수 전 2차 실측 필요"로 남긴 Gap

**결론 먼저: 세 계약 모두 Claude Code 와 동일하게 동작한다. 그리고 카드가 상정한 "출력 방언 차이"는
존재하지 않는다** — Codex 0.147.0 은 Claude Code 의 훅 출력 스키마를 통째로 구현하고 있다.

---

## 0. 측정 격리 + 원상복구

- 전 측정을 `CODEX_HOME=<probe-root>/home` 으로 분리. 사용자의 실제 `~/.codex/config.toml`·`hooks.json` 은 **무변경**
  (`~/.codex/hooks.json` mtime `Aug 13 07:21:46` — 측정 전후 동일 확인)
- 인증은 **복사하지 않고 심볼릭 링크**로 빌려 썼다(`home/auth.json -> ~/.codex/auth.json`). 토큰 사본이
  디스크에 남지 않는다. 측정 종료 후 링크 제거 확인(`auth symlink removed`), 원본 `~/.codex/auth.json` 무결 확인
- 토큰 값은 이 보고서 어디에도 출력하지 않았다(존재 여부만 확인)
- 원본 덤프·JSONL 은 `probe/` 에 보존하되 절대경로를 `<repo>`/`<probe-root>` 로 마스킹

## 1. 측정 결과 3건 — 전부 확인

| # | 계약 | 측정 | 결과 |
|---|---|---|---|
| 1 | PreToolUse **exit 2** 차단 | 훅이 stderr 에 사유를 쓰고 exit 2 | **차단됨.** `command_execution` 아이템 자체가 생성되지 않음(`T83RANOK` 0건). stderr 텍스트가 모델에 사유로 전달: *"The command was blocked by a workspace hook: `T83-BLOCKED-BY-EXIT2 …`"* |
| 2 | **stdout JSON** 거부 | exit 0 + `{"hookSpecificOutput":{"hookEventName":"PreToolUse","permissionDecision":"deny","permissionDecisionReason":"T83-DENY-VIA-JSON"}}` | **차단됨.** 명령 미실행(`T83RANOK` 0건). 모델 응답: *"The command was blocked by a pre-tool hook (`T83-DENY-VIA-JSON`)"* — 사유가 그대로 전달 |
| 3 | **additionalContext** 주입 | UserPromptSubmit 훅이 exit 0 + `{"hookSpecificOutput":{…,"additionalContext":"…codeword is ZARQUON."}}` | **주입됨.** 프롬프트 어디에도 없는 `ZARQUON` 을 모델이 답했다(2건) |

증거: `probe/run-exit2.jsonl`, `probe/run-jsondeny-addctx.jsonl` (JSONL 이벤트 스트림 원본).

## 2. **카드 전제 정정 — "출력 방언 차이"는 없다**

카드는 흡수할 차이 3가지 중 하나로 *"출력 방언(decision 대 permissionDecision/continue)"* 을 꼽았다.
**이 대비는 두 하네스 사이의 방언 차이가 아니라, 두 하네스가 공유하는 한 스키마 안의 공존 메커니즘이다.**

바이너리 문자열에서 추출한 Codex 0.147.0 의 훅 출력 계약 전체:

```
공통 wire         : continue  stopReason  suppressOutput  systemMessage  decision  reason  hookSpecificOutput
hookSpecificOutput: hookEventName  permissionDecision  permissionDecisionReason  additionalContext  updatedMCPToolOutput
PreToolUseDecisionWire            : approve | block
PreToolUsePermissionDecisionWire  : allow | deny | ask
PermissionRequestBehaviorWire     : behavior  updatedInput  updatedPermissions  interrupt
struct 목록: {PreToolUse,PostToolUse,SessionStart,UserPromptSubmit,Stop,SubagentStop,SubagentStart,
              PreCompact,PostCompact,PermissionRequest}CommandOutputWire + …HookSpecificOutputWire
```

이벤트별 exit 2 계약도 개별 문자열로 존재한다 — 즉 이벤트마다 stderr 의 의미가 다르다:

| 이벤트 | exit 2 시 stderr 의 의미 (바이너리 원문) |
|---|---|
| PreToolUse | `…exited with code 2 but did not write a blocking reason to stderr` |
| PermissionRequest | `…did not write a **denial** reason to stderr` |
| UserPromptSubmit | `…did not write a blocking reason to stderr` |
| Stop / SubagentStop | `…did not write a **continuation prompt** to stderr` |
| Stop / SubagentStop | `hook returned decision:block without a non-empty reason` |

**M3 에 대한 함의**: 출력 방언 변환 계층은 **필요 없을 가능성이 높다**. moai 훅이 지금 내보내는
`{"decision":"block"}` / `{"continue":false,"stopReason":…}` / `systemMessage` / `additionalContext` 는
Codex 가 **같은 이름으로 받는다**. M3 의 실제 잔여 범위는 ① 이벤트 이름 정규화 ② 설정 파일 형태(§3)
③ 이벤트별 능력 차이(§4) 로 좁아진다.

주의: 이는 **선언의 확인**이지 전 항목 동작 확인이 아니다. 실제로 돌려 본 것은 위 §1 의 3건뿐이다.
`continue`/`stopReason`/`suppressOutput`/`systemMessage`/`updatedMCPToolOutput` 은 **바이너리에 존재하나
동작 미측정** (§5).

## 3. **설정 파일 형태 — M0 §6 과 어긋나는 실측 2건 (M4 직결)**

측정 중 훅이 **네 번 연속 발화하지 않았다**. 원인을 좁혀 얻은 결과가 M4(배선 생성기)에 직결된다.

### 3.1 `matcher` 와 `version` 을 넣으면 조용히 죽는다

동작하는 실제 `~/.codex/hooks.json`(orca 가 설치한 것)의 형태는:

```json
{ "hooks": { "PreToolUse": [ { "hooks": [ { "type": "command", "command": "…", "timeout": 10 } ] } ] } }
```

- **`"version": 1` 없음**, **`"matcher"` 없음**, 타임아웃 키는 `timeout`(바이너리 내부 struct 는 `timeoutSec`)
- 이벤트 키는 **PascalCase**(`PreToolUse`) — 바이너리에 camelCase 집합(`preToolUse` …)도 있으나 설정 키는 PascalCase 가 동작
- `"matcher": "*"` 를 넣은 구성은 **경고 없이 미발화**했다. 바이너리에 `invalid matcher ` 문자열이 있고
  `*` 는 단독으로 유효한 정규식이 아니므로, matcher 가 정규식으로 해석돼 항목째 버려진 것으로 보인다(추정 — §5)

**M0 §1 의 "미지의 이벤트 이름은 조용히 무시된다"가 이름에 국한되지 않는다**: 스키마 위반 항목도 조용히
무시된다. M4 생성기의 화이트리스트 검증은 **이벤트 이름뿐 아니라 필드 집합까지** 덮어야 한다.

### 3.2 `<proj>/.codex/hooks.json` 은 발화하지 않았다 — M0 §6 과 배치

깨끗한 A/B 였다. **완전히 동일한 내용**의 파일을 위치만 바꿔 실행:

| 위치 | 결과 |
|---|---|
| `<proj>/.codex/hooks.json` (run A4) | **미발화** — 훅 stdin 덤프 0건, 명령은 정상 실행됨 |
| `$CODEX_HOME/hooks.json` (run A5, 같은 파일을 `cp`) | **발화** — 덤프 생성 + 차단 성공 |

앞선 3회(run A/A2/A3, 형태 변형 3종)도 전부 프로젝트 레벨 단독이었고 전부 미발화다.
반면 홈 레벨은 **첫 시도에 발화**했다.

M0 §6 은 *"글로벌 + `<proj>/.codex/hooks.json` 의 SessionStart 훅이 둘 다 발화 — 확인"* 이라고 적었다.
이번 측정과 **어긋난다**. 가능한 설명: 이벤트별 차이(SessionStart 만 프로젝트 레벨 지원), 신뢰(trust)
등록 경로 차이, 또는 M0 의 프로젝트 경로가 달랐을 가능성. **어느 쪽인지 규명하지 못했다(§5).**
`--dangerously-bypass-hook-trust` 경고는 두 경우 모두 2회 떴으므로 "발견은 됐으나 실행되지 않았다"에 가깝다.

M4 가 `.codex/hooks.json` 을 생성하도록 설계돼 있다면 **이 지점이 선결 확인 대상**이다 — 생성물이
조용히 아무것도 하지 않을 수 있다.

## 4. 페이로드 — t91 골든과 필드 집합 완전 일치

이번에 관측한 PreToolUse stdin 의 키 집합:

```
cwd  hook_event_name  model  permission_mode  session_id  tool_input  tool_name  tool_use_id
transcript_path  turn_id
```

`.moai/reports/t91/hook-payloads/PreToolUse.json` 의 키 집합과 **완전 동일**(정렬 비교).
`tool_name` 이 `"Bash"` 로 정규화돼 오는 것도 재확인. `permission_mode` = `bypassPermissions`.

→ M0 의 "`internal/hook` 파싱 계층 재사용 가능" 판단이 독립 측정으로 한 번 더 뒷받침된다.
골든 기준값은 t91 덤프 8건을 그대로 쓰면 된다(이번 관측이 그 중 2건을 재현했다).

## 5. 미검증 (Gaps — 단정하지 않은 것)

- **동작을 확인한 출력 키는 `permissionDecision`/`permissionDecisionReason`/`additionalContext` 3개뿐.**
  `continue`·`stopReason`·`suppressOutput`·`systemMessage`·`updatedMCPToolOutput`·`decision:"block"` 은
  바이너리에 **존재를 확인**했을 뿐 **돌려 보지 않았다**. moai 훅이 실제로 쓰는 키가 여기 포함되므로
  (`team-ac-verify.sh` 의 `continue`/`stopReason`, sync 게이트의 `systemMessage`) **M3 착수 전 3차 실측 대상**이다.
- **`matcher` 가 정규식이라는 것은 추론이다.** `invalid matcher` 문자열 + `*` 미발화로부터의 추론이며,
  유효한 정규식(예: `Bash`)을 넣어 발화하는지 재보지 않았다.
- **§3.2 의 M0 불일치 원인 미규명.** 이벤트별 차이인지 trust 경로 차이인지 가르지 못했다.
- exit 2 를 PreToolUse 외 이벤트(Stop/UserPromptSubmit/PermissionRequest)에서 재보지 않았다.
  바이너리 문자열상 의미가 이벤트마다 다르므로(§2 표) 동일 동작을 가정하면 안 된다.
- `additionalContextLimit`(~2,500 토큰), `"this event cannot emit additionalContext"` 가 **어느 이벤트**에
  걸리는지 미측정.
- 서브에이전트 경로는 이번 범위 밖 — M0 §2 의 `SubagentStop` 매핑 폐기 판정을 그대로 승계한다.

## 6. M3 설계에 대한 함의 (SPEC 초안 전 정리)

1. **출력 방언 변환 계층은 범위에서 빼는 쪽이 유력하다** — 같은 스키마다. 단 §5 의 3차 실측이 이를 확정해야 한다.
2. **남는 것은 이벤트 이름 정규화 + 설정 파일 생성 형태 + 이벤트별 능력 차이** 세 가지다.
3. **설정 형태 검증이 M3/M4 의 실질 위험**이다 — 스키마 위반이 조용히 무시되므로, 어댑터가 만든 배선이
   "설치됐는데 아무 일도 안 하는" 상태가 될 수 있고 그것을 알려 주는 신호가 없다.
4. `internal/hook` 로직 불변 유지는 페이로드 동형성(§4)으로 뒷받침된다.

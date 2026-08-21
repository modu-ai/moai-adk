# t83 3차 실측 — moai 실사용 키 / Stop exit 2 / 발화 경로 / matcher

베이스: `main` @ `4b2f203fe` · 브랜치 `WT-hook-adapter` · 대상 `codex-cli 0.147.0`, 모델 `gpt-5.6-sol`
지시: 리드 판정 (a) — 4항목 확장 실측 후 SPEC 작성

**결론 먼저: 방언 계층은 "제외"가 아니라 "부분 매핑"이 정답이다.** moai 가 실제로 쓰는 키 중
**`continue`/`stopReason`/`systemMessage` 세 개가 선언돼 있으나 동작하지 않는다.** 나머지는 네이티브 동작한다.

---

## 1. moai 실사용 키 — 3개 동작, 3개 무동작

| 키 | 이벤트 | 결과 | 관측 |
|---|---|---|---|
| `hookSpecificOutput.permissionDecision:"deny"` + `…Reason` | PreToolUse | **동작** | 명령 미실행, 사유가 모델에 전달 (2차) |
| `hookSpecificOutput.additionalContext` | UserPromptSubmit | **동작** | 프롬프트에 없는 코드워드를 모델이 답함 (2차) |
| `decision:"block"` + `reason` | Stop | **동작** | 턴이 이어지고 모델이 `reason` 의 지시를 수행(`KIWI` 발화) |
| **`continue:false`** | PostToolUse | **무동작** | 훅은 발화했으나 턴이 그대로 완주. 모델이 최종 요약까지 정상 출력 |
| **`continue:false`** | PreToolUse | **무동작** | 명령이 그대로 실행됨(`T83RANOK` 정상 출력) |
| **`stopReason`** / **`systemMessage`** | Pre/PostToolUse | **무동작** | JSONL·stderr 어디에도 문자열이 나타나지 않음(각 0건) |

증거: `probe/run-stop-decisionblock.jsonl`, `probe/run-continuefalse-inert-posttool.jsonl`,
`probe/run-continuefalse-inert-pretool.jsonl`.

### 이것이 SPEC 범위에 갖는 의미 (리드 지시 1의 판정)

리드 지시: *"전부 네이티브 동작 확인 시에만 방언 계층 제외를 확정하고, 하나라도 어긋나면 그 키만 남겨 명시."*
→ **어긋났다. 방언 계층을 통째로 제외할 수 없다.** 남겨야 하는 것은 정확히 이 셋이다:

| moai 현재 사용처 | 내보내는 키 | Codex 에서 | 필요한 매핑 |
|---|---|---|---|
| `team-ac-verify.sh` (TaskCompleted 거부) | `{"continue":false,"stopReason":…}` | **무시됨** | `Stop` 계열이면 `decision:"block"` + `reason` 으로 치환 |
| sync 게이트 advisory | `systemMessage` | **무시됨** | 전달 경로 없음 → `additionalContext`(해당 이벤트 지원 시) 또는 포기 결정 필요 |
| pre-tool 위험 패턴 가드 | `permissionDecision:"deny"` | 동작 | 매핑 불필요 |
| stop-goal 차단 | `decision:"block"` | 동작 | 매핑 불필요 |

`decision:"block"` 이 Stop 에서 동작하므로, `continue:false` 계열은 **의미 손실 없이 치환 가능**하다 —
이것이 부분 매핑을 얇게 유지할 수 있는 근거다.

## 2. Stop exit 2 — 동작하며, PreToolUse 와 stderr 의미가 다르다

Stop 훅이 stderr 에 문자열을 쓰고 exit 2 → **턴이 이어졌고 모델이 그 지시를 수행**(`MANGO` 발화).

**이벤트별 stderr 의미 차이가 실측으로 확인된다** — 바이너리 문자열의 서술과 일치한다:

| 이벤트 | exit 2 시 stderr 의 역할 | 관측된 표면 |
|---|---|---|
| PreToolUse | **차단 사유** | 모델에게 사유로 **표시**됨 (*"blocked by a workspace hook: …"*) |
| Stop | **continuation prompt** | 모델에게 **지시로 주입**되고 문자열 자체는 스트림에 나타나지 않음(0건) |

→ 어댑터가 stderr 를 이벤트 구분 없이 다루면 안 된다. Stop 계열의 stderr 는 사용자에게 보이는 사유가 아니라
모델에 대한 프롬프트다.

Stop 페이로드 키(관측): `cwd, hook_event_name, last_assistant_message, model, permission_mode,
session_id, stop_hook_active, transcript_path, turn_id` — `stop_hook_active: true` 확인.
t91 골든 `Stop.json` 과 동일 구성.

## 3. **발화 경로 — 프로젝트 레벨은 어떤 형태로도 발화하지 않는다 (M4 블로커)**

리드 지시 3(근본 원인 규명)에 대한 답. 가설을 하나씩 제거했다.

| # | 가설 | 검증 | 결과 |
|---|---|---|---|
| H1 | 설정 형태(`matcher`/`version`)가 원인 | 동작하는 실제 `~/.codex/hooks.json` 과 동일 형태로 프로젝트 레벨 배치 | **기각** — 여전히 미발화 |
| H2 | 프로젝트가 **신뢰(trust)** 되지 않아서 | 격리 `config.toml` 에 `[projects."<probe>/proj"] trust_level = "trusted"` 추가 후 재실행 | **기각** — 여전히 미발화 |
| H3 | 경로가 `.codex/hooks/hooks.json` 이어야 | 해당 경로에 배치(바이너리에 `hooks/hooks.json` 문자열 존재) | **기각** — 미발화 |
| H4 | 홈 레벨만 발화 | **동일 파일**을 `$CODEX_HOME/hooks.json` 으로 복사 | **성립** — 즉시 발화 |

**판정: codex-cli 0.147.0 에서 프로젝트 레벨 훅 발견은 관측되지 않는다.**
시도한 경로 2종 × 설정 형태 4종 × 신뢰 등록 유/무 전부 미발화, 홈 레벨은 첫 시도에 발화.

**M0 §6 과의 모순**: M0 는 *"글로벌 + `<proj>/.codex/hooks.json` 의 SessionStart 훅이 둘 다 발화 — 확인"*
이라고 적었다. **이 세션에서는 재현되지 않는다.**

한 가지 혼동 가능성을 지목해 둔다: 훅이 **한 개만** 등록된 상태에서도
`--dangerously-bypass-hook-trust` 경고 아이템이 **매 실행 정확히 2회** 나타난다(모든 run 에서 일관). M0 가
이 2회를 "두 레이어가 각각 발화"로 읽었을 가능성이 있다 — **추정이며 M0 의 실제 절차를 확인하지 못했다.**

> **[M4 블로커 후보]** t88(M4) 배선 생성기가 `.codex/hooks.json` 을 프로젝트에 생성하도록 설계돼 있다면,
> 생성물이 **설치는 되고 아무 일도 하지 않는** 상태가 된다. 그리고 그것을 알려 주는 신호가 없다.
> t88 착수 전 (1) 프로젝트 레벨 지원 여부 확정 (2) 미지원이면 생성 대상 경로를 `$CODEX_HOME/hooks.json`
> 으로 변경하거나 사용자 안내 경로를 재설계, 둘 중 하나가 선결이다. M0 §6 재실측 권고.

## 4. **matcher 가설 — 내 2차 보고의 추정을 철회한다**

2차 보고서 §3.1 에서 나는 *"`matcher:"*"` 가 유효한 정규식이 아니라 항목째 버려진 것으로 보인다"* 고
**추정**으로 적었다. **그 추정은 틀렸다.**

실측(홈 레벨, 한 파일에 두 항목):

| matcher | 발화 |
|---|---|
| `"Bash"` (유효 정규식) | **발화** |
| `"*"` | **발화** |

**진짜 원인은 `"version": 1` 이다.** 단일 변수 A/B:

| run | 파일 | 결과 |
|---|---|---|
| F | `version` 없음, PreToolUse + Stop | **둘 다 발화** |
| E | 같은 파일에 **`"version": 1` 만 추가** | **둘 다 미발화** |

→ **최상위 `version` 키 하나가 파일 전체를 무력화한다.**

**정정 (감사 지적 D5, 수용).** 초안은 여기에 "경고도 오류도 없다"고 적었다. **틀렸다.**
같은 run 의 JSONL **4행**에 오류 아이템이 있다:

```
failed to parse hooks config <probe-root>/home/hooks.json:
  unknown field `version`, expected `description` or `hooks` at line 2 column 11
```

파일·필드·행·열까지 나온다. 내가 훅 발화 여부와 마커 grep 만 보고 **error 아이템을 읽지 않아서**
무음으로 단정했다. 정확한 진술: **프로세스는 exit 0 이고 대화형 경고도 없으나, `--json` 스트림에는
기계 판독 가능한 오류가 실린다.** 덤으로 이 메시지가 **허용 최상위 키 집합(`description`, `hooks`)**
까지 알려 준다 — 내가 추측으로 세우려던 화이트리스트의 실측 근거다.

**진짜 무음은 §3 의 프로젝트 레벨 미발화 쪽이다** — `run-projectlevel-nofire.jsonl` 에는 오류
아이템이 0건이다. "아무도 알려 주지 않는다"는 서술은 그쪽에만 붙어야 한다.

증거: `probe/run-versionkey-kills-file.jsonl` (4행에 파싱 오류 아이템).

**M4 화이트리스트 검증이 필드 집합까지 덮어야 한다는 근거는 유지된다** — 다만 근거는 matcher 가 아니라
version 이다. 그리고 이 실패 양식이 M0 §1("미지의 이벤트 이름은 조용히 무시된다")보다 넓다는 점이
확인됐다: **알 수 없는 최상위 키는 그 항목이 아니라 파일 전체를 죽인다.**

## 5. 미검증 (Gaps)

- **`continue`/`stopReason`/`systemMessage` 를 Stop·SessionStart·UserPromptSubmit 에서는 재보지 않았다.**
  Pre/PostToolUse 두 이벤트에서 무동작을 확인했을 뿐이다. 다른 이벤트에서 동작할 가능성을 배제하지 못한다.
- `updatedMCPToolOutput`, `PermissionRequest` 계열(`behavior`/`updatedInput`/`updatedPermissions`/`interrupt`),
  `PreCompact`/`PostCompact` 미측정.
- `suppressOutput` 미측정.
- §3 의 M0 모순은 **원인을 규명하지 못하고 재현 불가로만 판정**했다. 2회 경고 혼동설은 추정이다.
- `additionalContextLimit`(~2,500 토큰)과 `"this event cannot emit additionalContext"` 가 걸리는 이벤트 미확정.
- 측정은 전부 `--dangerously-bypass-hook-trust --dangerously-bypass-approvals-and-sandbox` 아래에서 이뤄졌다.
  일반 승인 경로에서의 동작은 재보지 않았다.
- `permission_mode` 는 항상 `bypassPermissions` 로만 관측됐다(위 플래그 때문). 다른 모드에서의 페이로드 미관측.

## 6. 측정 위생

- `CODEX_HOME` 격리 유지. 실제 `~/.codex/hooks.json` mtime `Aug 13 07:21:46` — 전 측정 전후 동일
- 실제 `~/.codex/config.toml` 에 probe 경로가 들어가지 않았음을 확인(`grep -c t83probe` → 0,
  `grep -c worktrees/t83` → 0). 이 파일의 mtime 은 `Aug 22 02:22:02` 으로 **첫 codex 실행(02:35)보다 앞서며**
  이 측정에 기인하지 않는다
- 인증은 **복사 대신 심볼릭 링크**로 대여, 측정 종료 후 제거 확인. 토큰 값은 출력하지 않음
- 보존 증거의 절대경로는 `<repo>`/`<probe-root>` 로 마스킹

## 7. SPEC 범위에 대한 정리 (초안 전 확정 사항)

1. **부분 방언 매핑이 필요하다** — `continue:false`+`stopReason` → `decision:"block"`+`reason` 치환,
   `systemMessage` 는 전달 경로 부재로 **정책 결정 필요**(포기 또는 additionalContext 대체).
2. **이벤트별 stderr 의미 분기**가 어댑터에 들어가야 한다(PreToolUse=사유 표시 / Stop=continuation prompt).
3. **설정 생성 형태는 엄격하다** — 최상위 `version` 금지, `matcher` 는 허용. 알 수 없는 키가 파일 전체를
   죽이므로 생성기는 필드 화이트리스트를 가져야 한다(t88 소관이나 근거는 여기서 확정). 허용 최상위
   키는 오류 메시지가 명시한 `description`·`hooks` 둘뿐이다. 단 이 실패는 무음이 아니라
   `--json` 스트림에 오류로 실린다(§4 정정).
4. **프로젝트 레벨 배선은 현재 빌드에서 동작하지 않는다** — t83 어댑터의 배치 대상과 t88 생성 경로 모두에
   영향. **M4 블로커 후보로 등록.**
5. 페이로드 동형성은 재확인됐다(Stop 포함) → `internal/hook` 로직 불변 유지 근거 유효.

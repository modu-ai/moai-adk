# t452 M2 — 에이전트 `skills` 키 측정 결과

카드: t452 · SPEC-CODEX-SKILL-LOADER-001 · 브랜치 `WT-codex-skill-wiring`
BASELINE_SHA: `c529b2e4aaf5148aee7e6c67649bf392837bbb06`
codex 버전: `codex-cli 0.152.1` — 회차 안에서 관측(`m2-runs/A-version.txt`)
격리: `/tmp/t452-m2/proj` + `CODEX_HOME=/tmp/t452-m2/codexhome`(신규, `auth.json`만 복사)

## 판정 — 세 성질 모두 **확인되지 않음**

plan.md M2 가 요구한 세 관측 각각의 판정:

| # | 성질 | 판정 | 근거 |
|---|---|---|---|
| 1 | 존재 — `skills` 키를 단 에이전트가 등록되는가 | **확인되지 않음** | `skills` 키 **없는** baseline 에이전트조차 어떤 표면에도 나타나지 않는다 |
| 2 | 값 형태 — 이름인가 경로인가 | **확인되지 않음** | 1 이 미확인이므로 두 형태를 가를 관측이 성립하지 않는다 |
| 3 | 오값 거동 — 조용히 무시인가 파일 전체 폐기인가 | **확인되지 않음** | 이 매니페스트가 "보이는 오류를 낸다"고 기록한 필드의 오값조차 침묵했다 |

따라서 REQ-CSL-005 에 따라 이 필드는 방출하지 않으며, M3 은 REQ-CSL-007 경로
(`documented-drop`)로 간다.

**이것은 "재는 데 실패했다"가 아니라 "이 코덱스 버전의 비대화형 경로에는 잴 표면이
없다"는 관측이다.** 아래 대조군이 그 구분을 만든다.

## 관측 1 — 세션 자기 나열: baseline 에이전트가 없다

설치: `.codex/agents/moai/a-baseline.toml` 한 개(= `skills` 키 없음, 통제군).

```
$ cd /tmp/t452-m2/proj
$ CODEX_HOME=/tmp/t452-m2/codexhome codex --version
codex-cli 0.152.1
$ CODEX_HOME=/tmp/t452-m2/codexhome timeout 300 codex exec \
    --cd /tmp/t452-m2/proj --skip-git-repo-check -s read-only --json \
    -o /tmp/t452-m2/ev/A-last.txt \
    "List the names of every agent role (delegate/subagent) available to you in this session, one per line, exactly as registered. If none are available, reply with exactly: NONE" \
    > A.jsonl 2> A.err
$ echo "exit=$?"
exit=0
```

최종 메시지 전문(`m2-runs/A-last.txt`):

```
NONE
```

stderr 전문(`m2-runs/A.err`): `Reading additional input from stdin...` 한 줄.
이벤트 스트림(`m2-runs/A.jsonl`)에는 `thread.started` · `turn.started` ·
`item.completed`(=최종 메시지 `NONE`) · `turn.completed` 넷뿐이다.

## 관측 2 — `codex debug prompt-input`: 스킬은 보이고 에이전트 역할은 없다

`codex debug prompt-input` 은 "모델이 실제로 보는 입력 목록"을 JSON 으로 렌더한다.
**모델 자기보고가 아니라 런타임이 조립한 것을 그대로 찍는 표면이므로**, 관측 1 의
자기보고보다 강한 증거다.

```
$ CODEX_HOME=/tmp/t452-m2/codexhome codex debug prompt-input > A-prompt-input.json
$ echo "exit=$?"
exit=0
$ grep -c "probe-a-baseline" A-prompt-input.json
0
```

같은 파일이 스킬 뿌리 표를 **담고 있다**(`m2-runs/A-prompt-input.json`, 발췌):

```
### Skill roots
- `r0` = `/Users/goos/.agents/skills`
- `r1` = `/private/tmp/t452-m2/codexhome/skills/.system`
- `r2` = `/private/tmp/t452-m2/proj/.agents/skills`
...
- moai-probe-m2skill: Probe marker for the M2 agent skills-key measurement. ... (file: r2/moai-probe-m2skill/SKILL.md)
```

**이 줄은 M0 의 분기 A 를 독립적으로 재확인한다** — 프로젝트 `.agents/skills` 가 뿌리
`r2` 로 세션에 실려 있고, 심은 표식이 그 뿌리에 귀속되어 나타난다. 즉 계측기는 이
파일에서 **초록을 낼 수 있다**. 같은 파일에 에이전트 역할이 없는 것은 계측기가 죽어서가
아니다.

## 관측 3 — `spawn_agent` 도구 스키마에 역할 선택 인자가 없다

```
$ CODEX_HOME=/tmp/t452-m2/codexhome timeout 400 codex exec --cd /tmp/t452-m2/proj \
    --skip-git-repo-check -s read-only --json -o A3-last.txt \
    "Do not call any tool. Print the complete JSON schema of the collaboration spawn_agent tool exactly as it appears in your tool definitions, including every parameter name, its type, its description, and any enumerated allowed values. Print nothing else."
$ echo "exit=$?"
exit=0
```

출력 전문(`m2-runs/A3-last.txt`)의 `properties` 키 집합: `fork_turns`, `message`,
`model`, `reasoning_effort`, `task_name`. `required`: `message`, `task_name`.
**역할·role·agent 정의를 고르는 인자가 하나도 없다.**

선행 시도(`m2-runs/A2-last.txt`)에서 역할 이름으로 스폰을 요청했을 때 돌아온 런타임
오류는 `agent_name must use only lowercase letters, digits, and underscores` 였다 —
이름 검증이지 역할 조회가 아니다.

**한계 — 이 관측은 모델의 자기보고다.** A2 의 오류가 `agent_name` 을 말하고 A3 의
스키마는 `task_name` 을 말하므로 라벨이 어긋난다. 그래서 이 관측 하나로 판정하지 않고,
아래 대조군으로 확정했다.

## 대조군 — 계측기가 살아 있는지, 두 방향

### 양성 대조 (계측기가 초록을 낼 수 있는가)

관측 2 의 스킬 뿌리 표. 같은 명령·같은 파일에서 프로젝트 스킬이 이름으로 나타난다 →
**낼 수 있다.**

### 음성 대조 X — 이 매니페스트가 "보이는 오류"를 기록한 필드의 오값

`agents-codex.yaml` 의 `sandbox_mode` 절은 선행 회차(0.147.0)에서 오값이
`Ignoring malformed agent role definition` 을 눈에 보이게 내고 에이전트 파일 전체가
버려짐을 기록하고 있다. 그 필드에 오값을 넣은 `x-badsandbox.toml` 한 개만 설치하고
실행했다:

```
$ CODEX_HOME=/tmp/t452-m2/codexhome timeout 300 codex exec --cd /tmp/t452-m2/proj \
    --skip-git-repo-check -s read-only --json -o X-last.txt "Reply with exactly: PING"
$ echo "exit=$?"
exit=0
```

`m2-runs/X-last.txt` → `PING`. `m2-runs/X.err` → `Reading additional input from stdin...`
한 줄. **오류 없음.** `RUST_LOG=codex_agent_roles=trace` 를 얹은 재실행
(`m2-runs/X2.err`)도 역할 로더 이벤트를 하나도 내지 않았고, 격리 홈에 로그 디렉터리도
생기지 않았다.

### 음성 대조 Y — 파싱 자체가 불가능한 TOML

침묵이 "읽고 조용히 무시"인지 "아예 안 읽음"인지 가르는 대조군이다. 파서는 종결되지
않은 문자열을 조용히 받아들일 수 없다.

`y-brokentoml.toml`(`name = "probe-y-brokentoml` 미종결 + `description = ''' unterminated`)을
**두 위치에 차례로** 두고 각각 실행했다:

| 위치 | exit | 최종 메시지 | stderr |
|---|---|---|---|
| `.codex/agents/moai/y-brokentoml.toml` | 0 (`m2-runs/Y-exit.txt`) | `PING` (`Y-last.txt`) | 한 줄, 오류 없음 (`Y.err`) |
| `.codex/agents/y-brokentoml.toml` (flat) | 0 (`m2-runs/Yflat-exit.txt`) | `PING` (`Yflat-last.txt`) | 한 줄, 오류 없음 (`Yflat.err`) |

**결론: `codex exec` 0.152.1 은 프로젝트 `.codex/agents/` 를 이 경로에서 읽지 않는다.**
서브디렉터리 배치 탓이 아니다 — flat 위치도 같았다. 따라서 `skills` 키의 값 집합은
"확인되지 않음"보다 강하게 **관측 불가**다.

### 부수 표면 — `codex doctor`

```
$ CODEX_HOME=/tmp/t452-m2/codexhome codex doctor --json > A-doctor.json
$ grep -c "probe-a-baseline" A-doctor.json
0
```

`m2-runs/A-doctor.json` 에 에이전트 역할 점검 항목은 없다. M0 이 스킬 뿌리에 대해
관측한 것과 같은 결론이다.

## Gaps — 관측하지 **않은** 것

- **대화형 TUI 를 시험하지 않았다.** `Ignoring malformed agent role definition` 문자열은
  바이너리에 존재한다(`strings -a` 로 확인). 즉 그 기구는 어딘가에 살아 있고, 이 회차는
  그것이 **비대화형 경로에 나타나지 않는다**는 것만 관측했다. 대화형 세션에서는 다르게
  보일 수 있다.
- **`strings` 프로브를 판정 근거로 쓰지 않았다.** 바이너리에 `struct AgentRoleToml with 3
  elements`(필드 `description` · `config_file` · `nickname_candidates`)와, 그 옆에
  serde 의 `unknown field ..., expected ...` 경로가 보인다. 이는 **단서**이지 증거가
  아니다 — 문자열의 부재도 존재도 런타임 거동을 말하지 않는다(AC-CSL-001 이 M0 에
  대해 정한 것과 같은 이유). 방출 판단은 위의 런타임 관측만으로 내렸다.
- **에이전트 역할이 `~/.codex/config.toml` 의 `[agents.*]` 표를 경유해야 등록되는지**
  시험하지 않았다. 개발 저장소의 사용자 계층을 건드리지 않는다는 리드 제약과, 격리
  홈에 `config.toml` 을 만드는 것이 배포 사용자의 조건과 다르다는 점 때문이다. 이
  가능성이 열려 있다는 사실 자체가 `documented-drop` 의 이유를 강화한다.
- **관측 3 은 모델 자기보고다**(위 한계 참조). 확정은 대조군 Y 가 했다.

## Residual risk

- 관측은 **이 codex 버전·이 비대화형 경로**에 한정된다. 대화형 TUI 나 app-server 경로에
  역할이 살아 있다면, `.codex/agents/` 미러 자체는 여전히 쓸모가 있고 `skills` 키의
  의미론은 그때 다시 재야 한다. `documented-drop` 의 rationale 이 재프로브 조건을 적어
  두는 이유다.
- 대조군 Y 는 "읽지 않는다"를 exec 경로에 대해 보인다. 코덱스가 역할 파일을 **지연
  적재**해 스폰 시점에만 읽는 구현이라면, 스폰 자체가 불가능한 이 세션에서는 같은
  침묵이 나온다 — 두 설명 모두 "이 경로에서 `skills` 를 잴 수 없다"로 수렴하므로
  판정은 바뀌지 않지만, 원인 귀속은 열려 있다.

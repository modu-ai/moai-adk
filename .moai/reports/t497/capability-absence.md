# t497 M1 — `tool_classes` 전수 능력 부재 파생 기록

card: t497 · tree: `.claude/worktrees/t497` · branch `WT-codex-neutrality`
측정 HEAD: `76263a02e` (이 트리, 이 실행) · base `ace1c5440`

SPEC-CODEX-BODY-NEUTRALITY-001 M1 산출 1. `SPEC-CODEX-SKILL-NEUTRAL-001` REQ-CSN-003 이
못박은 파생 기준 — **결속표의 행 집합 = 이 표를 읽는 하네스에 존재하지 않는 모든
`tool_classes` 능력** — 을 `tool_classes` 값 집합 **11개 전수**에 적용한 기록이다.
부분 집합만 판정한 기록은 파생을 증명하지 못한다(REQ-CBN-016).

## 모집단 — 명령으로 뽑았다 (손 열거 아님)

```
$ sed -n '/^tool_classes:/,/^$/p' internal/template/agentemit/agents-codex.yaml \
    | grep -oE ': [a-z-]+$' | sed 's/^: //' | sort -u
cross-session-messaging
design-sync
file-read
file-write
moai-mcp
question-channel
shell
skill-loader
subagent-spawn
task-list
web

$ ... | wc -l
11
```

## 판별식 — 낱말이 아니라 verdict

`agents-codex.yaml` `classes:` 의 해당 클래스 rationale 을 판별식으로 삼는다.
**rationale 이 코덱스 쪽 대응물의 존재를 서술하면 `present`, 대응물이 없다고 서술하면
`absent`.** `disposition: documented-drop` 이라는 사실만으로는 행이 생기지 않는다 —
필드 부재와 능력 부재는 다른 것이고, 결속표는 **능력 부재**를 채운다.

**[HARD] 트랩 1건 — `moai-mcp` 는 `present` 다.** 그 rationale 은 `unavailable` 이라는
낱말을 싣지만, 부재한 것은 **한 서버 안의 도구별 필터링**이지 MCP 능력 자체가 아니다.
같은 rationale 이 `[mcp_servers.moai]` 테이블로 서버 수준 부여가 성립한다고 적고
disposition 은 `emit-field` 다. 낱말 `unavailable` 로 키를 잡으면 부재가 3 이 아니라
**4** 로 나와 아래 대조가 깨진다.

## 전수 판정

| class | rationale 인용 (agents-codex.yaml `classes:`) | verdict |
|---|---|---|
| file-read | "Codex exposes built-in file read; no agent-TOML field exists or is needed." — 읽기 능력 존재, 필드만 불필요 | present |
| file-write | "The Read-vs-Write tool distinction is not mechanically preserved on Codex: its sandbox is workspace-level, not tool-level." — 쓰기 능력 존재, 구분만 미보존 | present |
| shell | "Shell executes built-in on Codex." disposition `consequence` — 능력 존재 | present |
| web | "No per-agent web field exists; web access is a global Codex feature/config outside agent TOML." — 능력 존재, 부여 축만 다름 | present |
| task-list | "The Claude task-tool family has **no known Codex equivalent**; body prose degrades gracefully." | absent |
| skill-loader | "Session skill loading itself **IS confirmed** on this version (the roots table lists the project .agents/skills), so this drop is about the per-agent grant, not about skills reaching a session." | present |
| subagent-spawn | "Codex delegation **exists** (internal collaboration\* tools) but a per-agent spawn grant is not expressible in agent TOML." | present |
| design-sync | "DesignSync is a Claude-specific MCP-backed tool with **no Codex equivalent**." | absent |
| cross-session-messaging | "The Codex **counterpart rides the moai MCP broker** (session_msg_register/list/send/poll) under the existing server-level moai-mcp grant." | present |
| question-channel | 실측 — "told to invoke the question tool, codex reported it **unavailable** and substituted a prose question; its own request_user_input tool errors 'unavailable in Default mode'." | absent |
| moai-mcp | "Agents carrying any mcp__moai__\* token declare the server-level grant as a TOML table [mcp_servers.moai] … the table form registers cleanly." disposition `emit-field`. 부재한 것은 서버 안 도구별 필터링뿐 | present |

## 파생 결과 — 부재 3건, 실린 표 3행, 값이 맞물린다

부재 3건: `task-list` · `design-sync` · `question-channel`.
이 셋은 **현재 `AGENTS.md` 결속표에 실린 바로 그 3행**이다.

```
$ grep -cE '\|[[:space:]](absent|present)[[:space:]]\|[[:space:]]*$' .moai/reports/t497/capability-absence.md
11
$ grep -cE '\|[[:space:]]absent[[:space:]]\|[[:space:]]*$' .moai/reports/t497/capability-absence.md
3
$ sed -n '/^\*\*Capability bindings/,/^---$/p' AGENTS.md | grep -c '^| [a-z]'
3
```

**행 0개 추가가 이 카드의 예상 경로이며 정당한 결과다.** 이 기록이 묻는 것은 행이
늘었는가가 아니라 파생이 지켜졌는가이고, 부재 건수 == 데이터 행 수가 그 판정이다.

`skill-loader` 와 `subagent-spawn` 은 둘 다 **능력 존재** 쪽이므로 행을 얻지 못한다.
따라서 `manager-lead` 의 자기 스폰 지시 줄과 41개 `invoke Skill(` 줄은 **존재하지 않는
결속행을 가리키지 않고**, 각각 본문에 능력 이름 + 대체 행동을 직접 적거나(전자)
`AGENTS.md` 결속표 문단의 덮개 1문장으로 덮인다(후자).

## 산출 2 — REQ-CSN-003 문면 정정

`SPEC-CODEX-SKILL-NEUTRAL-001` REQ-CSN-003 의 「현재 측정값 4행」은 스테일하다.
위 전수 대조가 부재 3건에서 멈추고 실물 표도 3행이므로, 스테일한 쪽은 문면이다.
파생 기준 문장(「존재하지 않는 모든 능력이 정확히 한 행을 얻는다」)의 **의미는 바꾸지
않고** 실측 수치만 3행으로 갈아쓰고, 그 SPEC HISTORY 에 Amendments 1행을 남긴다.

같은 파일의 「4행 = 373 B」는 후보 표 `.moai/reports/t196/csn003-table-4row.txt` 의
**크기 측정 기록**이므로 손대지 않는다 — 그 파일은 실제로 4행이다.

## Gaps

- 위 판정은 rationale **문면** 판정이다. 각 능력의 코덱스 실제 거동을 이 트리에서 새로
  프로브하지 않았다. `question-channel` 만 매니페스트에 실측 기록이 있다.
- `classes:` 에는 `tool_classes` 값이 아닌 항목(`effort` · `model` ·
  `frontmatter-hooks` · `memory-color-permission` · `explore-builtin`)도 있으나, 파생
  기준의 주어가 `tool_classes` 능력이므로 이 표의 모집단이 아니다.

## Residual-risk

rationale 이 낡았을 수 있다 — 측정 버전이 클래스마다 다르다(`skill-loader` 0.152.1,
`question-channel` 0.150.1). 이 잔여 위험은 행을 **더하지 않는** 방향이므로 보수적이다.
새 프로브가 어떤 능력의 부재를 보이면 그때 행이 하나 늘고, 이 기록은 그 판정의 기준선이 된다.

# (b) 와이어 모양 판정 — 설치 바이너리 대상

측정 시점: 2026-09-02 · 워크트리 `.claude/worktrees/t401` · 브랜치 `WT-analysis-pull` · HEAD `6352897a5`

## 왜 별도 측정이 필요했나

구현자의 e2e 증거는 **워크트리 빌드**(`./bin/moai`)에 **최소 형태 payload** 를 먹인 것이었다.
남은 위험은 두 갈래였고 섞으면 판정이 안 된다:

- **(a)** 실행 중 Claude Code 세션이 새 PreToolUse matcher 를 집어 드는가
- **(b)** 런타임이 보내는 payload 의 실제 모양이 탐지기가 읽는 필드와 맞는가

(b) 가 어긋나면 모든 live 행이 조용히 `label_present:false` 로 떨어지고, **첫 창이 "규약이 잘
지켜진 상태"와 구분되지 않는다.** 이 문서는 (b) 만 판정한다.

## 측정

설치 바이너리 확인:

```
$ ~/go/bin/moai version
 v3.1.2   v3.1.2-1308-g65196a5a7   built 2026-09-02T05:53:08Z
exit=0
```

`65196a5a7` = 리드가 push 한 develop tip(= 제 M0 병합 커밋). 관측자가 실려 있다.

입력 payload(`e2e/payload-installed.json`) — 구현자 것보다 실제에 가깝게 `cwd` ·
`hook_event_name` · `transcript_path` · `permission_mode` · `multiSelect` · 한글 레이블 ·
`description` 를 포함:

```json
{"session_id":"wirecheck-1","cwd":"/tmp","hook_event_name":"PreToolUse","transcript_path":"/dev/null","permission_mode":"bypassPermissions","tool_name":"AskUserQuestion","tool_input":{"questions":[{"question":"어느 쪽으로 진행할까요?","header":"범위","multiSelect":false,"options":[{"label":"판단 우선 모드 축 신설 (권장)","description":"기존 절을 일반화합니다."},{"label":"배너 필드만 pull","description":"Frozen 미접촉."}]}]}}
```

실행 + 결과:

```
$ CLAUDE_PROJECT_DIR=<scratch> ~/go/bin/moai hook pre-tool < e2e/payload-installed.json
{"hookSpecificOutput":{"hookEventName":"PreToolUse","permissionDecision":"allow"}}
exit=0

$ cat <scratch>/.moai/logs/askuser-observations.jsonl
{"timestamp":"2026-09-02T06:03:02Z","session_id":"wirecheck-1","mode":"push","label_present":true,"option_count":2,"question_count":1,"payload_parsed":true}
```

## 판정

| 축 | 결과 |
|---|---|
| 설치 바이너리에 관측자 존재 | ✔ 행이 생성됨 |
| deny 없음 | ✔ `permissionDecision: "allow"`, exit 0 |
| 한글 `(권장)` 레이블 탐지 | ✔ `label_present:true` |
| 파싱 성공 표시 | ✔ `payload_parsed:true` — 미파싱을 진짜 음성으로 오독할 여지 없음 |
| 개수 | ✔ `option_count:2` · `question_count:1` |
| `question_type` 생략 | ✔ 직렬화에 부재 |
| 모드 해석 | ✔ `push` (설정 키 부재 시 기본값) |

**(b) 초록.** 다만 이것은 *제가 만든* payload 에 대한 판정이다. 런타임이 실제로 보내는 payload
와의 일치는 live 호출 1건으로만 확정된다 — 그 대조가 [HARD] 로 남아 있는 이유다.

## 잔여

(a) 는 미판정. `/hooks` 로 PreToolUse 항목이 3개인지 4개인지 보면 갈리는데 슬래시 커맨드라
오케스트레이터가 실행할 수 없다. 운영자 답 대기 중이며, **그 전까지 "집어 든다"고 가정하지 않는다.**

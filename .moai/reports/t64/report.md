# t64 — GLM(glm-5.3) 컨텍스트 창 1M 표시 실측

> 카드 원문 (backlog t64, picked): "glm 5.3의 경우 이렇게 나오는데 모두 1M 가 아니라 200k 인거 같다 확인을 해서 이 문제를 해결하자."
>
> 처리 범위(디스패치 tjv7iy 라운드4): t65 잔무 — `CLAUDE_CODE_MAX_CONTEXT_TOKENS` 선언 후 실제 표시가 1M인지 실측 + MoAI statusline 교정 경로(glmContextWindows) 확인. 측정 카드(코드 변경 0).

## 판정 요약

**원 카드 주장("200K로 보임")은 배포 바이너리 v3.1.0에서 재현되지 않는다 — 4개 표면이 전부 1,000,000을 보고한다.** t65(PR #1574, 머지 225a51e24)의 `CLAUDE_CODE_MAX_CONTEXT_TOKENS` 선언이 `moai glm` 실행 경로에서 실제로 작동한다. "해결"은 이미 머지·릴리스된 상태이며 본 카드는 그 실측 확인이다.

| 표면 | 관측값 | 증거 |
|---|---|---|
| ① CC 원시 statusline 페이로드 (교정 전) | `context_window.context_window_size: 1000000` | `raw/statusline-stdin.jsonl` |
| ② MoAI statusline 스냅샷 | `context_window_size: 1000000`, `raw_pct: 0`, `band: large` | `raw/context-usage.json` |
| ③ 세션 TUI `/context` | **"Auto-compact window: 1m tokens"** | tmux capture (본문 인용) |
| ④ 세션 결과 JSON | `modelUsage["glm-5.3"].contextWindow: 1000000` | t77 `raw/run-a.txt`, `raw/run-b.txt` |
| (배경) live 프로세스 env | `CLAUDE_CODE_MAX_CONTEXT_TOKENS=1000000`, `CLAUDE_CODE_AUTO_COMPACT_WINDOW=1000000` | `ps eww` (t77 리포트 §증거 2) |

## 배경 — 200K 오보고의 원인과 수정

- Claude Code는 `claude-` 접두사가 없는 커스텀 모델 id(예: glm-5.3)에 대해 창을 200K로 가정하며, 이 때문에 텔레메트리 표시와 `CLAUDE_CODE_AUTO_COMPACT_WINDOW` 상한이 모두 200K에 묶였다 (이슈 #653, `internal/cli/glm.go:318-331` 주석).
- t65 수정: `setGLMEnv`이 `glmMaxContextTokens(모델)`로 해석된 창(glm-5.3 → 1,000,000, `internal/statusline/memory.go:27-36` glmContextWindows 테이블)을 `CLAUDE_CODE_MAX_CONTEXT_TOKENS`로 선언 — 표시와 오토컴팩트 상한을 함께 해제.
- MoAI statusline은 자체 교정 경로(우선순위 2: glmContextWindows 서브스트링 매칭, 우선순위 1: `MOAI_STATUSLINE_CONTEXT_SIZE` 수동 오버라이드)를 유지한다. 본 실측에서는 원시 페이로드 자체가 이미 1M라 교정이 "뒤집는" 것이 아니라 "일치하는" 상태로 관측됨 — 두 경로가 독립적으로 같은 값(1M)을 내는 교차확인.

## 환경

- t77과 동일 실험장(`/tmp/t77t64-lab`, llm.yaml 전 슬롯 glm-5.3, `raw/llm-yaml-used.yaml` 참조), 설치 바이너리 **v3.1.0 (d6b80a01c)** — t65 수정 포함을 ancestry로 확인, claude v2.1.233.
- 측정 세션: tmux 대화형 `moai glm` 세션(신뢰 대화상자·MCP 승인 수락 후 안정 상태) + headless 런 A/B.

## 절차 (재현)

```bash
# statusline 원시 페이로드 캡처 — status_line.sh를 tee 래퍼로 교체 후 세션 기동
mv .moai/status_line.sh .moai/status_line.real.sh
# 래퍼: tee -a <증거경로> | .moai/status_line.real.sh
tmux new-session -d -s t64glm -c /tmp/t77t64-lab
tmux send-keys -t t64glm -l '~/go/bin/moai glm' && tmux send-keys -t t64glm Enter
# (신뢰/MCP 대화상자 수락) → tmux send-keys -l '/context' + Enter → tmux capture-pane -p

# 확인 포인트
cat /tmp/t77t64-lab/.moai/state/context-usage.json     # 표면 ②
tail -1 <tee로 캡처한 statusline-stdin.jsonl>            # 표면 ①
```

## 증거 발췌 (원시 그대로)

표면 ① — CC 원시 페이로드(statusline 교정 전, 세션 a5bfa9c6):

```json
{"model":{"id":"glm-5.3","display_name":"glm-5.3"},
 "context_window":{"total_input_tokens":0,"total_output_tokens":0,
                   "context_window_size":1000000,...},
 "version":"2.1.233","exceeds_200k_tokens":false,...}
```

표면 ③ — TUI `/context` 패널 상단 (tmux capture):

```
     Auto-compact window: 1m tokens

     MCP tools · /mcp
     └ 19 tools · 0 tokens
     ...
```

statusline 렌더(`🔋 CW: ░░░░░░░░░░ 0%` 게이지, `🤖 glm-5.3 │ 🗿 v3.1.0`)도 정상 동작 — 0%는 신선 세션(토큰 미소비)의 올바른 표시이며, 창 분모가 1M이므로 초기 프리픽스(~97K)로도 10% 미만을 유지한다는 계산과 일관된다.

표면 ④ — 런 A/B JSON: `"modelUsage":{"glm-5.3":{...,"contextWindow":1000000,"maxOutputTokens":32000,"canonicalModel":"glm-5.3","provider":"firstParty"}}`.

## Gaps (미검증)

- **수정 전 상태 재현 생략**: pre-t65 바이너리에서 "200K 표시"를 직접 재현하지 않았다(구 바이너리 체크아웃·빌드 비용 대비 이력이 명확 — t65 카드·PR #1574 기록으로 대체). 원 주장의 재현 불능은 "현재 바이너리에서 4표면 전부 1M"로 간접 입증.
- glm-5.3 이외 모델(glm-5.2 / glm-4.7 등 창 크기가 다른 티어)은 미측정 — glmContextWindows 테이블상 값만 확인(소스).
- 세션 내부에 보이는 "잔여 15,000,000 tokens" 예산 표시(런 A/B 세션 자가 보고)는 원인 미규명 — 세션 스스로 "신뢰성 제한적"이라 단서를 붙였고 창 크기 필드(1M)와는 별개다. 후속 카드 후보.
- 오토컴팩트가 실제 1M 스케일로 발동하는지(대용량 세션에서의 임계값)는 미측정 — 본 카드 범위(표시 실측) 밖.

## 잔여 위험

- CC 버전 업그레이드로 statusline 페이로드 스키마·창 보고 동작이 바뀔 수 있음 — statusline 교정 경로(glmContextWindows)가 이중 안전망으로 유지되므로 영향은 제한적.
- `CLAUDE_CODE_MAX_CONTEXT_TOKENS`는 z.ai가 아닌 **클라이언트(CC) 측 선언**이다 — 실제 모델 서버의 하드 한도와 어긋나면 창이 1M로 표시돼도 서버가 거부할 수 있다(본 실측 범위에서는 97K+ 토큰 정상 처리 관측).

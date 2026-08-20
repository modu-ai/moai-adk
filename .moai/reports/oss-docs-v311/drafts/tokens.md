---
title: moai tokens
weight: 17
draft: false
new: true
added_in: "v3.1.1"
---

{{< new-badge v3.1.1 >}}

Claude Code 세션의 토큰 사용량을 **풀(pool)과 출처(origin)별로 나눠** 기록하는 회계 도구입니다. "이번 카드에 토큰을 얼마나 썼는가"를 세션 전체 합계 하나로만 보면, 어느 백엔드가 얼마를 썼는지·메인 대화와 서브에이전트 사이드체인이 각각 얼마를 썼는지가 묻힙니다. 이 명령은 그 둘을 갈라서 남깁니다.

{{< callout type="info" >}}
**한 줄 요약**: `moai tokens record`가 세션 트랜스크립트를 읽어 풀별(glm/claude/other)·출처별(메인 대화/서브에이전트 사이드체인) 사용량을 집계하고, 컨텍스트 사용 스냅샷을 덧붙여 `.moai/state/token-accounting.jsonl`에 한 줄로 append합니다.
{{< /callout >}}

## 사용법

```bash
# 열려 있는/직전 세션 트랜스크립트를 지정해 기록
$ moai tokens record --transcript <경로> --card t12 --role run

# 세션 id로 지정
$ moai tokens record --session <세션-ID> --card t12

# 레코드를 JSON으로 출력 (파일 기록과 함께)
$ moai tokens record --transcript <경로> --json
```

| 플래그 | 설명 |
|--------|------|
| `--transcript <경로>` | 집계할 Claude Code 트랜스크립트 파일 |
| `--session <id>` | 세션 식별자로 트랜스크립트를 지정 |
| `--card <카드>` | 이 사용량을 묶을 칸반 카드(예: `t12`) |
| `--role <역할>` | 세션의 역할(예: `run`, `sync`, `worker-3`) |
| `--json` | 표준 출력으로도 레코드를 JSON으로 내보냅니다 |

## 기록의 생김새

레코드는 `.moai/state/token-accounting.jsonl`에 **append-only**로 쌓입니다 — 세션·카드가 끝날 때 한 줄씩 남기는 장부입니다. 각 줄에는:

- **풀별 사용량** — `glm` / `claude` / `other` 로 나뉜 합계. 어느 백엔드가 청구서를 만들었는지가 풀에서 바로 보입니다.
- **출처별 사용량** — 메인 대화와 서브에이전트 사이드체인. 워커를 여럿 띄운 런에서 "정작 구현은 워커가 다 썼는가"를 가려 줍니다.
- **컨텍스트 스냅샷** — 기록 시점에 컨텍스트 사용 상태(`.moai/state/context-usage.json`)가 있으면 그 값이 함께 들어갑니다.

## 언제 기록하나

설계상 카드 또는 세션이 **닫히는 시점**의 기록입니다. 칸반·팩토리 런에서는 카드 하나가 끝날 때마다, 단일 세션에서는 큰 작업이 끝날 때 남기면 카드별 비용 비교가 성립합니다. 이 명령 자체는 토큰을 소비하지 않습니다 — 이미 발생한 사용량을 트랜스크립트에서 다시 세는 회계입니다.

## 관련 문서

- [토크노믹스 개요](/ko/advanced/tokenomics-overview) — 왜 배정이 단가보다 중요한가
- [상태 표시줄](/ko/advanced/statusline) — 세션이 진행 중일 때 사용량을 보는 자리
- [칸반 모드](/ko/advanced/kanban-mode) — 카드·레인 단위로 비용을 묶는 런의 형태

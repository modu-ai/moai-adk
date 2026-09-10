---
title: "Codex 듀얼 하네스 — AGENTS.md·에이전트 이중 게시·훅 어댑터"
weight: 31
draft: false
added_in: "v3.1.3"
description: "codex-cli가 Claude Code와 나란히 MoAI-ADK를 쓰게 하는 공통 표면과 하네스별 개인 지침."
---

MoAI-ADK는 Claude Code를 1차 하네스(에이전트를 실제로 구동하는 실행 환경)로 삼지만, v3.1.3부터 **codex-cli에서도 같은 계약을 읽을 수 있는 이중 표면**을 갖추었습니다. 공통 규칙과 에이전트 정의는 Codex가 찾는 위치와 형식으로 함께 게시하고, 개인 지침은 하네스별로 분리합니다. 이 문서는 각 표면이 어떤 문제를 푸는지 설명합니다.

## 루트 AGENTS.md — 하네스 공통 standing contract

저장소 루트의 `AGENTS.md`는 Claude Code 전용이 아니라 **어떤 에이전트 하네스가 턴을 구동하든 묶는 standing contract**(항시 계약)입니다. 하나의 파일로 존재하는 이유는 codex의 읽기 방식에 있습니다: codex는 프로젝트 지시문을 바이트 상한 안에서 읽는데, 넘치는 뒷부분을 **경고 없이, 종료 코드 0으로 조용히 버립니다**. 상한을 넘은 계약은 마치 온전한 것처럼 보고됩니다. 그래서 파일 하나가 상한 안에 들어가는 것 자체가 요구사항이고, 빌드 가드(build guard, 빌드 시마다 이 파일이 상한 이내인지 검사하는 장치)가 그것을 지킵니다.

공간을 만들기 위해 항상 로드되던 문서 11개는 8개의 지연 로드 컴패니언(lazy companion, 필요할 때만 읽는 상세 문서)을 가리키는 스텁(stub, 짧은 요약)으로 내려갔습니다. **옮겨진 것은 의무가 아니라 그 의무를 설명하는 산문**입니다 — `AGENTS.md`가 하네스 공통 계약의 기준이고, `.claude/rules/moai/**`와 `CLAUDE.md`는 Claude 전용 메커니즘을 확장합니다.

{{< callout type="info" >}}
개인 `~/.codex/AGENTS.md`는 같은 병합 체인에서 이 파일 **앞에** 소비되어, 프로젝트 계약이 실을 수 있는 폭을 좁힙니다. 넘침은 뒤에서부터 조용히 버려지므로, 이 파일의 조항은 가장 중요한 것부터 앞에 배치돼 있습니다.
{{< /callout >}}

## 에이전트 이중 게시 — 11개의 TOML

유지되는 11개 에이전트가 두 형태로 게시됩니다. Claude Code용 `.claude/agents/moai/*.md`(원본)와 codex가 읽는 `.codex/agents/moai/*.toml`(파생본)입니다. TOML은 손으로 쓰지 않습니다 — `internal/template/agentemit`이 마크다운 원본에서 **결정적으로(deterministically, 같은 입력에 언제나 같은 출력)** 생성하며, 생성된 파일 머리글은 "regenerate, do not edit"(다시 생성하라, 직접 고치지 마라)이라고 못 박혀 있습니다.

원본과 파생본이 어긋나는 일을 세 겹의 가드가 막습니다: 골든 파일 비교(golden, 기대 출력과의 대조), 임베드 검증(바이너리에 심긴 템플릿과의 대조), 배포 검증(사용자 저장소에 깔린 결과와의 대조). 마크다운을 고치면 TOML이 따라오고, TOML만 고치면 가드가 붙잡습니다.

## `.agents/skills` — 스킬 미러

codex-cli는 Claude Code의 `.claude/skills/`를 읽지 않으므로, 스킬을 `.agents/skills` 아래에 **미러**(거울 복사본)로 배포합니다. 미러 목록은 손으로 관리하지 않고 배포 실행 시점의 실제 스킬 집합에서 도출되므로, 스킬이 늘고 줄어도 목록이 어긋나지 않습니다. 이 디렉터리는 **사용자 저장소 밖**을 향한 배치 산출물 취급이라 git에 기록되지 않으며, 심볼릭 링크를 우선하되 만들 수 없는 환경에서는 복사로 대체 배포됩니다(`moai init`·`moai update`의 완료 요약이 그 사실을 알립니다 — 자세한 것은 [moai update](/ko/cli-reference/update/) 문서 참조).

## 하네스별 개인 지침

`AGENTS.local.md`는 Codex 전용입니다. 로컬 `moai codex` 런처가 프로젝트 루트에서 이 파일을 읽고, 내용을 그대로 Codex 세션의 `developer_instructions` 덮어쓰기로 전달합니다. 공통 파일에서 `@`로 가져오지 않습니다. `CLAUDE.local.md`·`.claude/settings.local.json`·Claude 자동 `MEMORY.md`는 Claude 전용으로 남습니다. Codex Web 세션은 로컬 런처를 거치지 않으므로 이 주입을 받지 않습니다.

## `internal/codexadapter` — 훅 어댑터 라이브러리

두 하네스의 훅 표면은 거의 같지만 완전히 같지는 않습니다. 실측(codex-cli 0.153.4 기준)에서 갈린 지점은 세 가지: 하네스가 넘기는 **이벤트 이름**, codex가 선언은 하지만 실제로는 반응하지 않는 **출력 키 세 개**(`systemMessage`·`continue`·`stopReason`), 그리고 **PreToolUse 결정 계약**입니다 — codex 파서는 `updatedInput` 없는 `permissionDecision:allow`와 `permissionDecision:ask`를 거부합니다. `internal/codexadapter`는 디스패처 **앞에** 앉는 얇은 번역층이고(`internal/hook`은 건드리지 않음) 거부되는 결정 형태(allow·ask·defer)는 무의견 `{}`로 열화해 codex 자체 승인 흐름에 맡기며, 각 열화는 discard 싱크로 공지되고 사유 없는 deny에는 기본 사유가 채워집니다.

### 12-이벤트 표

| Codex 이벤트 | MoAI 디스패처 인자 | 이 마일스톤에서 적응? |
|---|---|---|
| PreToolUse | `pre-tool` | 예 |
| PostToolUse | `post-tool` | 예 |
| SessionStart | `session-start` | 예 |
| SessionEnd | `session-end` | 예 |
| Stop | `stop` | 예 |
| UserPromptSubmit | `user-prompt-submit` | 예 |
| PreCompact | `compact` | 아니오 — 비대화형 실행에서 압축이 한 번도 촉발되지 않음 |
| PostCompact | `post-compact` | 아니오 — 비대화형 실행에서 압축이 한 번도 촉발되지 않음 |
| PermissionRequest | `permission-request` | 아니오 — 비대화형 실행에서 승인 요청이 발화되지 않음 |
| SubagentStart | `subagent-start` | 예 |
| SubagentStop | `subagent-stop` | 예 |
| Interrupt | — (대응물 없음) | 아니오 — SIGINT에서 발화; 적응에는 새 디스패처 서브커맨드 필요(후속 카드) |

열두 개 이벤트를 다룹니다: 열한 개에는 디스패처 대응물이 존재하고, 공식 문서화된 12번째 이벤트 `Interrupt`는 대응물이 생길 때까지 별도의 no-counterpart 메시지로 인식 후 거부됩니다. codex-cli 0.153.4 실측에서 `SubagentStart`와 `SubagentStop`은 **발화가 확인됐고**(SubagentStop이 발화하지 않는다는 이전 0.147.0 관측을 뒤집은 결과) 지금은 적응됐습니다 — `RenderHooks`가 사용자 `.codex/hooks.json`에 `moai hook subagent-start --harness codex`와 `moai hook subagent-stop --harness codex` 줄을 심습니다. compact·permission 계열은 추정이 아니라 정직한 근거로 보류됐습니다: 비대화형 `codex exec` 실행은 압축에 도달하지 못했고(측정된 입력 상한 1,048,576자 대비 최선 264,808 입력 토큰) 승인 요청도 끌어내지 못했습니다 — "발화하지 않는다"가 아니라 **촉발 미달성**으로 기록됩니다.

미적응 이벤트는 묵살되지 않고 **거부**됩니다. 미확인 이벤트(오타)와 인식되었으나 다루지 않는 이벤트(범위 결정 또는 `Interrupt`처럼 대응물 부재)가 서로 다른 오류로 구분되므로, 운영자는 실수와 결정을 가려낼 수 있습니다. 설정 검증기는 알 수 없는 키 위반을 첫 번째에서 멈추지 않고 **전부 수집해** 한꺼번에 보여줍니다.

### 지금 호출부는 어디인가

`RenderHooks`가 적응된 이벤트 명령 여덟 개를 사용자 `.codex/hooks.json`에 기록합니다. `moai init --llm codex|both`가 이 배선을 만들고, 기존 프로젝트에서는 `moai tool enable codex`로 추가하거나 갱신합니다.

## 다음 단계

- [다중 모델 감사 수렴](/ko/advanced/multi-model-audit/) — codex 백엔드가 감사에 참여하는 지금의 경로
- [moai update](/ko/cli-reference/update/) — 스킬 미러의 symlink·복사 배포와 그 통지
- [에이전트 가이드](/ko/advanced/agent-guide/) — 이중 게시되는 11개 에이전트의 역할

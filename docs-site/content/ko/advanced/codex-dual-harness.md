---
title: "Codex 듀얼 하네스 — AGENTS.md·에이전트 이중 게시·훅 어댑터"
weight: 31
draft: false
added_in: "v3.1.3"
description: "codex-cli가 MoAI-ADK를 읽을 수 있게 하는 네 가지 산출물 — 루트 AGENTS.md standing contract, 에이전트 TOML 이중 게시, .agents/skills 스킬 미러, internal/codexadapter 훅 어댑터 라이브러리."
---

MoAI-ADK는 Claude Code를 1차 하네스(에이전트를 실제로 구동하는 실행 환경)로 삼지만, v3.1.3부터 **codex-cli에서도 같은 계약을 읽을 수 있는 이중 표면**을 갖추었습니다. 무엇 하나도 Claude Code 쪽 동작을 바꾸지 않습니다 — 이미 있던 규칙과 에이전트 정의를 codex가 찾는 위치와 형식으로 한 번 더 게시했을 뿐입니다. 이 문서는 그 네 가지 산출물이 무엇이고, 각각 어떤 문제를 푸는지를 다룹니다.

## 루트 AGENTS.md — 하네스 공통 standing contract

저장소 루트의 `AGENTS.md`는 Claude Code 전용이 아니라 **어떤 에이전트 하네스가 턴을 구동하든 묶는 standing contract**(항시 계약)입니다. 하나의 파일로 존재하는 이유는 codex의 읽기 방식에 있습니다: codex는 프로젝트 지시문을 바이트 상한 안에서 읽는데, 넘치는 뒷부분을 **경고 없이, 종료 코드 0으로 조용히 버립니다**. 상한을 넘은 계약은 마치 온전한 것처럼 보고됩니다. 그래서 파일 하나가 상한 안에 들어가는 것 자체가 요구사항이고, 빌드 가드(build guard, 빌드 시마다 이 파일이 상한 이내인지 검사하는 장치)가 그것을 지킵니다.

공간을 만들기 위해 항상 로드되던 문서 11개는 8개의 지연 로드 컴패니언(lazy companion, 필요할 때만 읽는 상세 문서)을 가리키는 스텁(stub, 짧은 요약)으로 내려갔습니다. **옮겨진 것은 의무가 아니라 그 의무를 설명하는 산문**입니다 — 원천은 여전히 `.claude/rules/moai/**`와 `CLAUDE.md`이고, `AGENTS.md`는 그것을 하네스 중립 형태로 운반합니다.

{{< callout type="info" >}}
개인 `~/.codex/AGENTS.md`는 같은 병합 체인에서 이 파일 **앞에** 소비되어, 프로젝트 계약이 실을 수 있는 폭을 좁힙니다. 넘침은 뒤에서부터 조용히 버려지므로, 이 파일의 조항은 가장 중요한 것부터 앞에 배치돼 있습니다.
{{< /callout >}}

## 에이전트 이중 게시 — 11개의 TOML

유지되는 11개 에이전트가 두 형태로 게시됩니다. Claude Code용 `.claude/agents/moai/*.md`(원본)와 codex가 읽는 `.codex/agents/moai/*.toml`(파생본)입니다. TOML은 손으로 쓰지 않습니다 — `internal/template/agentemit`이 마크다운 원본에서 **결정적으로(deterministically, 같은 입력에 언제나 같은 출력)** 생성하며, 생성된 파일 머리글은 "regenerate, do not edit"(다시 생성하라, 직접 고치지 마라)이라고 못 박혀 있습니다.

원본과 파생본이 어긋나는 일을 세 겹의 가드가 막습니다: 골든 파일 비교(golden, 기대 출력과의 대조), 임베드 검증(바이너리에 심긴 템플릿과의 대조), 배포 검증(사용자 저장소에 깔린 결과와의 대조). 마크다운을 고치면 TOML이 따라오고, TOML만 고치면 가드가 붙잡습니다.

## `.agents/skills` — 스킬 미러

codex-cli는 Claude Code의 `.claude/skills/`를 읽지 않으므로, 스킬을 `.agents/skills` 아래에 **미러**(거울 복사본)로 배포합니다. 미러 목록은 손으로 관리하지 않고 배포 실행 시점의 실제 스킬 집합에서 도출되므로, 스킬이 늘고 줄어도 목록이 어긋나지 않습니다. 이 디렉터리는 **사용자 저장소 밖**을 향한 배치 산출물 취급이라 git에 기록되지 않으며, 심볼릭 링크를 우선하되 만들 수 없는 환경에서는 복사로 대체 배포됩니다(`moai init`·`moai update`의 완료 요약이 그 사실을 알립니다 — 자세한 것은 [moai update](/ko/cli-reference/update/) 문서 참조).

## `internal/codexadapter` — 훅 어댑터 라이브러리

두 하네스의 훅 표면은 거의 같지만 완전히 같지는 않습니다. 실측(codex-cli 0.147.0 기준)에서 갈린 지점은 정확히 두 가지: 하네스가 넘기는 **이벤트 이름**, 그리고 codex가 선언은 하지만 실제로는 반응하지 않는 **출력 키 세 개**(`systemMessage`·`continue`·`stopReason`)입니다. 나머지는 전부 동일하게 측정됐으므로, `internal/codexadapter`는 디스패처 **앞에** 앉는 얇은 번역층이고 `internal/hook`은 건드리지 않습니다.

### 11-이벤트 표

| Codex 이벤트 | MoAI 디스패처 인자 | 이 마일스톤에서 적응? |
|---|---|---|
| PreToolUse | `pre-tool` | 예 |
| PostToolUse | `post-tool` | 예 |
| SessionStart | `session-start` | 예 |
| SessionEnd | `session-end` | 예 |
| Stop | `stop` | 예 |
| UserPromptSubmit | `user-prompt-submit` | 예 |
| PreCompact | `compact` | 아니오 — 측정 없음 |
| PostCompact | `post-compact` | 아니오 — 측정 없음 |
| PermissionRequest | `permission-request` | 아니오 — 측정 없음 |
| SubagentStart | `subagent-start` | 아니오 — 측정 없음 |
| SubagentStop | `subagent-stop` | 아니오 — 실측에서 발화되지 않음 |

11개 전부에 디스패처 대응물이 존재합니다. 적응에서 제외하는 것은 측정 커버리지에 관한 범위 결정이지 대응물의 부재가 아닙니다. SubagentStop이 특별한 이유: 실측에서 **한 번도 발화되지 않았습니다** — codex에서 위임은 도구 이름이 "collaboration"으로 시작하는 PostToolUse로 드러나므로, 이를 연결하면 절대 흐르지 않는 경로에 선을 얹는 셈입니다.

미적응 이벤트는 묵살되지 않고 **거부**됩니다. 미확인 이벤트(오타)와 인식되었으나 이번엔 다루지 않는 이벤트(범위 결정)가 서로 다른 오류로 구분되므로, 운영자는 실수와 결정을 가려낼 수 있습니다. 설정 검증기는 알 수 없는 키 위반을 첫 번째에서 멈추지 않고 **전부 수집해** 한꺼번에 보여줍니다.

### 아직 호출부가 없다

{{< icon warning warn >}} 이 패키지는 라이브러리로 출하된 상태이고 **아무것도 이것을 호출하지 않습니다**. `--agent` 설정 생성기(생성된 codex 설정 파일에 어댑터를 배선하는 뒤따르는 카드)가 연결을 완성합니다. 지금 시점에서 이 문서는 배선이 될 자리의 지도이지, 켜진 스위치의 사용 설명서가 아닙니다.

## 다음 단계

- [다중 모델 감사 수렴](/ko/advanced/multi-model-audit/) — codex 백엔드가 감사에 참여하는 지금의 경로
- [moai update](/ko/cli-reference/update/) — 스킬 미러의 symlink·복사 배포와 그 통지
- [에이전트 가이드](/ko/advanced/agent-guide/) — 이중 게시되는 11개 에이전트의 역할

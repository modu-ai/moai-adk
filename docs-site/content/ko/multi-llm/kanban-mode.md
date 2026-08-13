---
title: 칸반 모드
weight: 30
draft: false
---

{{< callout type="info" >}}
칸반 모드의 전체 개요와 Origin-Trail Chain 설계 방향은 [칸반 모드](/ko/advanced/kanban-mode)를 보세요. 이 페이지는 다중 세션(리드 + 컴패니언) 운용 절차를 다룹니다.
{{< /callout >}}

## 칸반 모드란?

칸반 모드는 하나의 **리드** 세션이 `plan -> run -> verify -> sync` 체인을
주도하고, 네 개의 **컴패니언** 세션이 같은 런에 합류해 작업을 병렬로
분산합니다. 런의 모든 세션(리드와 컴패니언 모두)은 상향된 Stop-hook
블록 캡을 받아 세션 중반에 설정한 골이 기본 연속 블록 한계를 넘어
계속 실행됩니다.

리드가 체인을 시드하고, 컴패니언은 그렇지 않습니다. 각 컴패니언은
칸반 멤버십 플래그(`-k`)와 역할 라벨(`--name <role>-<run-id>`)을
함께 전달하므로 디스패처가 올바르게 분류하고 SessionStart 훅이
멤버십을 안내합니다.

## 진입 스위치

### 리드 진입

```bash
moai cc -k                     # Claude 백엔드 리드
moai cc -k SPEC-AUTH-001       # SPEC에 묶인 리드
moai glm -k                    # GLM 백엔드 리드
```

리드 세션은:

{{< icon check-circle ok >}} `MOAI_KANBAN` + `MOAI_KANBAN_ID` 설정 (체인 시드).
{{< icon check-circle ok >}} SessionStart에서 런 id와 네 개의 컴패니언 실행 명령을 출력.
{{< icon x-circle danger >}} `MOAI_KANBAN_LABEL`은 설정하지 않음 (컴패니언 신호).

### 컴패니언 진입

```bash
moai cc -k --name plan-abc123    # plan 컴패니언
moai cc -k --name run-abc123     # run 컴패니언
moai cc -k --name review-abc123  # review 컴패니언
moai cc -k --name sync-abc123    # sync 컴패니언
moai glm -k --name run-abc123    # GLM 백엔드에서 동일
```

`<run-id>`는 리드가 시작 시 안내한 식별자이며, 네 가지 역할은
`plan`, `run`, `review`, `sync` 입니다.

컴패니언 세션은:

{{< icon check-circle ok >}} `MOAI_KANBAN_LABEL` 설정 (멤버십 + 역할 라벨).
{{< icon check-circle ok >}} 리드와 동일한 상향된 Stop-hook 블록 캡.
{{< icon x-circle danger >}} `MOAI_KANBAN`은 설정하지 않음 — 체인을 시드하지 않습니다.

### 무처리 (변경 없는 세션)

```bash
moai cc --name mysession         # -k 없음, 칸반 멤버십 없음
moai cc --name run-abc123        # 컴패니언 형태이나 -k 없음 → 무처리
```

`-k`가 없으면 `--name` 형태와 무관하게 디스패처는 아무 작업도 하지
않습니다. `--name` 플래그는 그대로 Claude에 전달됩니다.

## 다중 세션 부트스트랩 흐름

```
터미널 1 (리드)            터미널 2-5 (컴패니언)
─────────────────          ────────────────────────
moai cc -k                 moai cc -k --name plan-<run-id>
                           moai cc -k --name run-<run-id>
                           moai cc -k --name review-<run-id>
                           moai cc -k --name sync-<run-id>
```

부트스트랩은 수동입니다: 세션은 다른 세션을 실행할 수 없습니다. 리드
SessionStart 알림이 복사할 네 개의 명령을 정확히 출력합니다. 각
컴패니언을 GLM 백엔드로 실행하려면 `moai cc`를 `moai glm`으로 바꾸면
됩니다.

## 교차 세션 메시징

세션 간 통신은 Claude Code의 교차 세션 메시징(`ListAgents` /
`SendMessage`)을 사용합니다. `crossSessionInbound` 설정 필드가 인바운드
메시지를 수락할지, 보류할지, 거부할지를 제어합니다.

칸반 모드는 인바운드 메시지를 자동 수락합니다: 실행기가
`{"crossSessionInbound": "accept"}`를 담은 임시 설정 파일을 작성하고
`--settings`로 백엔드에 전달합니다. 파일은 세션 전용이며(종료 시
정리) 영구 설정을 변경하지 않습니다.

### 운영자 제공 `--settings`

명령줄에 `--settings <file>`을 전달하면 실행기는 자체 설정 파일을
주입하지 않습니다. 파일에 다음 내용이 있는지 확인하세요:

```json
{
  "crossSessionInbound": "accept"
}
```

리드 SessionStart 알림은 실행기가 주입하지 않았을 때 확인을
상기시키는 안내를 출력합니다.

## SessionStart 알림

리드 알림은 런 id, 네 개의 컴패니언 실행 명령, 리더 소켓 경로,
인바운드 자동화 상태를 안내합니다. 컴패니언 알림은 합류를 확인하는
역할 없는 한 줄입니다. 두 알림 모두 프롬프트하지 않으며 정보 제공용
stdout입니다.

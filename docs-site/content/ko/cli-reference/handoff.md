---
title: moai handoff 핸드오프 레코드
weight: 68
draft: false
---

`moai handoff` 는 auto-resume 핸드오프 pending 레코드를 관리합니다. 세션 경계(`/clear`)를 넘어 작업을 이어 가려고 준비해 둔 paste-ready resume 본문을 저장하거나 지웁니다. `handoff.mode: auto` 로 설정해 두면 저장된 레코드가 다음 세션이 시작될 때 자동으로 주입됩니다.

SPEC 하나가 여러 세션에 걸쳐 진행될 때, 이전 세션의 진행 상황이 없으면 다음 세션이 처음부터 다시 맥락을 모아야 하기 때문에 토큰과 시간 낭비가 큽니다. 그래서 이 커맨드는 오케스트레이터가 emits 한 6-블록 resume 본문을 매개로 삼아, 이전 SPEC 단계의 전제·검증·실행 명령을 다음 세션으로 건네 줍니다. 따라서 관리자 에이전트가 긴 에픽을 연속으로 끌고 가는 흐름에서 끊김 없는 복귀 지점을 만들어 줍니다.

## 하위 명령어

| 명령어 | 설명 |
|--------|------|
| `moai handoff save` | paste-ready resume 본문을 pending 레코드로 저장 |
| `moai handoff clear` | pending 레코드 제거 |

공통 플래그로 `--project-dir <path>` (프로젝트 루트, 기본: 현재 디렉터리)를 받습니다.

## moai handoff save

```bash
moai handoff save --stdin --spec SPEC-AUTH-001 --phase run < resume.txt
```

| 플래그 | 설명 |
|--------|------|
| `--body <text>` | resume 본문 (verbatim 6-block paste-ready) |
| `--stdin` | `--body` 대신 stdin에서 본문 읽기 |
| `--spec <id>` | 이 핸드오프가 재개하는 SPEC id |
| `--phase <plan\|run\|sync>` | 단계 |
| `--session <uuid>` | saved_by_session uuid (귀속) |
| `--lang <lang>` | conversation_language 스냅샷 |
| `--ultrathink` | ultrathink 지시 기록 (복원 안내용) |
| `--ultracode` | ultracode 지시 기록 (복원 안내용) |
| `--goal <condition>` | `/moai goal` 조건 기록 (복원 안내용) |

## moai handoff clear

```bash
moai handoff clear
```

pending 핸드오프 레코드를 제거합니다.

## Fail-open 보장

`moai` CLI가 PATH에 없거나 `moai handoff save` 가 non-zero로 종료되어도 오케스트레이터의 paste-ready 출력은 그대로 나옵니다. 저장에 실패해도 핸드오프 출력이 멈추는 일은 없고, 수동 paste 경로는 저장 없이도 온전히 동작합니다. 저장은 보조 단계일 뿐 필수 조건이 아닙니다.

## 관련 문서

- [자율 연속 루프](/ko/advanced/autonomous-loops)
- [moai goal](/ko/cli-reference/goal)
- [CLI 개요](/ko/getting-started/cli)

---
title: moai handoff 핸드오프 레코드
weight: 68
draft: false
---

`moai handoff` 는 auto-resume 핸드오프 pending 레코드를 관리합니다. 세션 경계(`/clear`)를 넘어 작업을 이어가기 위한 paste-ready resume 본문을 저장하거나 지웁니다. 저장된 레코드는 `handoff.mode: auto` 설정 시 다음 세션 시작에 자동 주입됩니다.

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
| `--goal <condition>` | `/goal` 조건 기록 (복원 안내용) |

## moai handoff clear

```bash
moai handoff clear
```

pending 핸드오프 레코드를 제거합니다.

## Fail-open 보장

`moai` CLI가 PATH에 없거나 `moai handoff save` 가 non-zero로 종료되어도 오케스트레이터의 paste-ready 출력은 변경 없이 유지됩니다. 저장 실패는 핸드오프 방출을 절대 막지 않으며, 수동 paste 경로는 저장 없이도 완전히 동작합니다 — 저장은 부가적 영속화 단계일 뿐 게이트가 아닙니다.

## 관련 문서

- [자율 연속 루프](/ko/advanced/autonomous-loops)
- [moai goal](/ko/cli-reference/goal)
- [CLI 개요](/ko/getting-started/cli)

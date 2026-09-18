---
title: /moai gtd
weight: 29
draft: false
new: true
---

# /moai gtd

`/moai gtd`와 `moai gtd`는 작업을 수집하고 실행 가능성을 판단한 뒤 기존 개발 대기열로 연결하는 정식 GTD 표면입니다. 기존 `todo`의 SQLite DB, 카드 ID, 순서, `queued`·`picked`·`dropped` 상태와 보관·복원 의미는 바뀌지 않습니다. `todo`는 같은 명령 트리를 쓰는 호환 이름으로 남습니다.

```bash
moai gtd add "인증 오류 경로 정리"
moai gtd list
moai gtd next t1 --spec SPEC-AUTH-001
moai gtd done t1 --expect "인증"
```

기존 `moai todo ...` 호출도 같은 결과를 냅니다. 전체 대기열 동사와 플래그는 [todo 호환 명령 참고서](/ko/utility-commands/moai-todo)를 확인하세요.

GTD 전용 다섯 동사는 실제 SQLite 상태를 이어서 사용합니다.

```bash
moai gtd capture "인증 오류 경로 정리" --event inbox-42 --source user --sensitivity private
moai gtd clarify <gtd-id> --disposition action --outcome "병합 완료" --evidence "테스트와 병합 SHA" --authority "queue,commit" --trusted
moai gtd organize <gtd-id> --class action --context computer --depends-on <gtd-id>
moai gtd reflect --rebuild-projection --json
moai gtd engage <gtd-id> --approve --fresh --dependencies-ready --resources --pick --dispatch --lane lane-10 --run-id mission-42
```

`capture`에는 재시도 중복을 막는 `--event`가 필수입니다. `engage`의 배차는 `--pick`, `--lane`, `--run-id`가 모두 있어야 하며 승인·최신성·의존성·자원 확인을 하나라도 생략하면 효과를 만들지 않습니다. 작업 receipt는 SQLite에 먼저 준비되고, 중간 종료 뒤에는 같은 작업 ID의 실제 상태를 다시 읽어 중복 실행을 막습니다.

GTD v2 내보내기·가져오기는 항목과 관계뿐 아니라 봉인 계약, 임무, 이벤트, 작업 receipt까지 보존하며 명시적 opt-in과 무결성 검사를 요구합니다. 카드 archive/reopen은 같은 GTD 정체성을 유지하고, `reflect`는 보관 완료·재개방·취소와 source revision 변경을 다시 읽어 오래된 근거와 막힌 후속 작업을 표시합니다.

## 다섯 절차

GTD는 개발 보드의 열이 아니라 **일을 정리하고 선택하는 절차**입니다.

| 절차 | 판단과 저장 결과 |
|---|---|
| Capture | 출처·민감도·중복 식별자를 기록하며, 아직 개발 카드를 만들지 않습니다. |
| Clarify | 원하는 결과, 완료 근거, 권한과 자료 신뢰도를 확인합니다. 하나라도 불분명하면 발행을 보류합니다. |
| Organize | 프로젝트·행동·참고·보류 성격, 실행 맥락, 검토 시점과 관계를 정리합니다. 순환 의존은 거절합니다. |
| Reflect | 근거 변경, 보관·재개방·취소와 예약 검토 이벤트를 다시 살핍니다. 취소는 선행 작업의 완료로 보지 않습니다. |
| Engage | 승인 범위, 최신 근거, 의존성, 레인 소유권과 자원 한도를 모두 통과한 항목만 추천하거나 대기열에 연결합니다. |

개발은 계속 `backlog → plan → run → sync → done` 순서로 진행합니다. Capture부터 Engage까지가 이 열들을 대신하지 않습니다.

## 관계와 비공개 그래프

`depends_on`만 실행을 막는 관계입니다. `part_of`, `supported_by`, `related_to`는 참고 관계이며, 기존 `contains`, `absorbs`, `replaces`, `conflicts`의 의미도 유지됩니다. GTD 그래프는 대기열 DB 옆의 비공개 파생물이며 저장소, 로그, 텔레메트리, 기본 내보내기와 기본 백업에 포함되지 않습니다. 원본 revision과 메타데이터가 다르면 오래된 그래프로 판정합니다.

## 자율 운영과 안전 경계

LLM과 `mission-governor`는 구조화된 제안을 만들 뿐 직접 파일, Git, 대기열이나 배차 상태를 바꾸지 않습니다. 일반 코드가 사용자가 승인해 봉인한 목표·범위·허용 행위·완료 근거·자원 한도·중단 조건과 최신 snapshot을 검사한 뒤에만 작업을 준비합니다. 범위 확대, 오래된 근거, 승인되지 않은 행위는 자동 승인하지 않고 `blocked`로 남깁니다.

현재 구현은 정책·복구·배차·명시 경로 커밋·local develop `--no-ff` 병합을 수행하는 owner adapter를 제공합니다. 커밋은 현재 HEAD에 대응하는 저장소 내부 `0600` 테스트 receipt가 있어야 하며, 로컬 병합은 manager-git 역할, `WT-*` 브랜치, 기준 SHA, `.git` 아래 `0600` lease를 다시 확인합니다. 백업·복원·내보내기는 명시적 opt-in에서만 GTD 확장 정보를 포함하고, private projection은 SQLite revision에서 다시 만듭니다. 실제 공급자가 시작·재연결·교체·자격 증명·프로세스 식별 능력을 모두 제공한다고 확인되기 전에는 세션 종료 뒤에도 계속 실행되는 durable 모드나 원격 push·PR·병합 완료를 보장하지 않습니다.

관련 문서: [`/moai goal --auto`](/ko/utility-commands/moai-goal#auto-임무-모드) · [Kanban Mode](/ko/advanced/kanban-mode)

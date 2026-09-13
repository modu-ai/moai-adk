---
card: t647
branch: WT-todo-audit
source_session_id: 01a08e64-c9d3-76f3-a589-5d5d893e1b62
baseline: ee99507fbe3b4a22c6a0a74815723d222dfdc04d
---

# Todo 전수 조사 개선 작업 배정 기록

사용자 요청: todo 기능 전수 조사 → 개선 보고 → 개선 카드 발행 → 구현 완료.

구현 작업 공간은 런처가 생성한 `.claude/worktrees/todo-audit`이며, 원래 공유 체크아웃의 작업물은 수정하지 않는다. 보고 후 발행한 카드 `t647`로 T01–T22를 추적한다.

| 담당 | 소유 범위 | 결과 근거 |
|---|---|---|
| todo_cli_audit | internal/cli/todo*.go 및 관련 테스트 | cli.md |
| todo_storage_audit | internal/kanban/backlog*, todo_root, state_dir, internal/homestate 경로 | storage.md |
| todo_surfaces_audit | 웹 화면·감시·번역, 4개 언어 Todo 및 웹 콘솔 문서, Todo 스킬·규칙과 배포 사본 | surfaces.md |
| 영실 | 통합 검토, 보고서, 최종 검증·커밋·카드 상태 확인 | todo-audit-20260911/verdict.md |

분담자는 다른 담당자의 변경을 되돌리지 않으며 커밋·푸시·실사용 큐 변경은 수행하지 않는다. 저장 계층은 웹 담당자의 읽기 전용 교차 검토를 추가로 받았다. 교차 검토에서 확인한 이관 대상 존재 분기의 쓰기 차단 누락은 저장 담당자에게 전달했다.

Git 설정은 manual/local이며 자동 push와 PR을 비활성화하고 있다. 이번 완료 판정은 로컬 구현·관련 회귀 검증·구현 커밋까지이며, 원격 반영·설치 바이너리 교체·실사용 홈 이관은 별도 상태로 표시한다.

기준 트리에는 과거의 다른 작업에 속한 `t647/verdict.md`가 이미 있다(`691f489eb`). 현재 큐에서 새로 발행된 `t647`의 원문은 이번 Todo 감사 작업으로 확인했다. 과거 보고서를 덮어쓰지 않으며, 이번 판정은 날짜별 하위 디렉터리에 기록한다. 카드 ID만으로 과거 작업과 이번 작업이 같다고 간주하지 않는다.

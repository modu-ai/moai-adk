# t648 작업 인계

- card: t648
- branch: WT-todo-unified
- worktree: /Users/goos/MoAI/moai-adk-go/.claude/worktrees/todo-unified
- source_session_id: 01a08e64-c9d3-76f3-a589-5d5d893e1b62
- observed_session_current: 5c437366-d8cb-45cf-9261-a850f7127958
- remote_default_base: b155c95f9942206d820ad3b7b4fc1cb132f41bb4
- integrated_baseline: 8ab93ee20d3adb4750552067f29fcf96977fb724
- predecessor: WT-todo-audit / 2bae4f1cc1892af61632b0c672056bdff7198932
- status: 준비 완료, 전문 에이전트 생성 제한으로 신규 구현 미착수
- report: reports/todo-unified-progress-20260911.md

## 확정된 운영자 지시

카드와 실행의 정확성에 필요한 데이터는 Todo DB로 통일하고, heartbeat·상세 로그는 별도로 둘 수 있다. 완료 확정·복구 → 의미 있는 관계 Graph → 시각화 순서로 모두 진행한다. 앞선 제안의 구현 범위와 착수를 승인받았다. 카드 발행과 pick은 이 지시에 근거한다.

## 재개 순서

1. 새 세션에서 기존 작업 공간에 재진입한다. 새 카드를 중복 발행하지 않는다.
2. `moai session current`, 해당 worktree의 branch/HEAD/status, `moai todo list --json`의 t648을 다시 확인한다. 현재 기록과 다른 상태를 덮어쓰지 않는다.
3. 이 dispatch와 Markdown 보고서를 읽는다. 선행 통합을 다시 실행하지 않는다.
4. manager-spec으로 SPEC-TODO-UNIFIED-001을 작성하고 독립 plan-auditor 검토를 수행한다. 아직 이 SPEC은 생성되지 않았다. 승인된 범위는 유지하고 실제로 새 결정이 필요한 항목만 운영자에게 확인한다. 적용되는 plan→run 게이트를 따른다.
5. manager-develop의 TDD 구현 → 범위별 검증 → manager-docs 동기화 → sync-auditor 검토 순서를 따른다. Graph와 웹은 완료 확정 기능 이후에 진행한다.
6. 현재 카드의 승인된 완료 조건을 모두 검증한 뒤에만 완료 처리하고 결과를 다시 읽는다.

## 금지·보존 경계

- primary checkout의 기존 변경은 다른 작업이다. 수정·stash·reset·branch 전환을 하지 않는다.
- 운영 홈 DB 이전, 설치 바이너리 교체, push/PR/merge/deploy는 이번에 실행하지 않았다. 별도 범위 확인 없이 수행하지 않는다.
- 현재 worktree는 미전달 작업을 보유하므로 삭제하지 않는다.
- 실행 환경의 세 차례 `agent thread limit reached`는 애플리케이션 결함이 아니다. 구현이나 검증 성공으로 바꾸어 기록하지 않는다.

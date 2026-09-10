# t541 internal/cli 슬롯 기록

트리: 워크트리 `.claude/worktrees/t541`, 브랜치 `WT-auditor-order-conflict`, HEAD `542f7dbc2` + 미커밋 `internal/cli/plan_audit_order_conflict_test.go`.
승인: 리드 1회 슬롯(internal/cli 스코프). 전체 스위트는 돌리지 않았다.

## 사전 확인

- `pgrep -fl 'internal/cli|cli\.test'` → 출력 없음, exit 1 (다른 internal/cli 컴파일·테스트 프로세스 0)
- 대조 `pgrep -f claude | wc -l` → `52` (pgrep 이 살아 있는 프로세스를 본다는 증거)
- 원시 프로세스 목록은 커밋하지 않는다(다른 세션 argv 포함 가능).
- 뮤턴트 전 해시: `pre-mutant-sha.txt`

## 실행 결과

선택자 `-run 'TestPlanAuditOrder_' -v`. 테스트 함수는 15개다(슬롯 요청문의 "14"는 계수 착오 — 파일에 15개가 있다).

| 단계 | 파일 | exit | PASS 줄 | FAIL 테스트 |
|---|---|---|---|---|
| 초록 | `green.txt` / `green.exit` | 0 | 15 | 없음 |
| vet | `vet.txt` / `vet.exit` | 0 | — | 출력 없음 |
| m1 하위 제목 수준 검사 제거(템플릿) | `m1.txt` / `m1.exit` | 1 | 13 | SubheadingKeepsBinding, VerbIdenticalAcrossCopies |
| m2 after 방향 분기 삭제(템플릿) | `m2.txt` / `m2.exit` | 1 | 13 | AfterConflictIsReported, VerbIdenticalAcrossCopies |
| m3 `Milestone` 제목 형식 제거(템플릿) | `m3.txt` / `m3.exit` | 1 | 13 | MilestoneHeadingFormIsRead, VerbIdenticalAcrossCopies |
| m4 절 제목 AC 대체 제거(템플릿) | `m4.txt` / `m4.exit` | 1 | 13 | SectionCriterionIsTheSubject, VerbIdenticalAcrossCopies |
| m5 "never forces FAIL by itself" 줄 삭제(로컬) | `m5.txt` / `m5.exit` | 1 | 14 | ClausesPresent |
| 복원 후 초록 | `final-green.txt` / `final-green.exit` | 0 | 15 | 없음 |

각 뮤턴트는 적용 직후 `/usr/bin/grep -c` 로 대상 사본에서 0, 반대 사본에서 1 을 확인한 뒤 실행했다.

## 복원

- 백업은 세션 스크래치패드의 `cp` 사본. 뮤턴트마다 복원 후 `cmp` exit: m1 0, m2 0, m3 0, m4 0, m5 0.
- 최종 `shasum -a 256 -c pre-mutant-sha.txt` → 두 사본 OK, exit 0.
- 최종 `git status --short` → 추적 파일 수정 0 (미추적: 증거 디렉터리와 새 테스트 파일뿐).
- 뮤턴트는 커밋하지 않았다.

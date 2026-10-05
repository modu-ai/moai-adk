# progress.md — SPEC-FACTORY-COMPLETION-RECOVERY-001

## §A 진행 기록

- **2026-10-06** — phase=plan 시작. lane-1, 트리 `.moai/worktrees/t1538`, 브랜치 `WT-t1538-factory-recovery`, base = local develop `a158b4b5f`.
- **임대 거절·영수증 경로**: `moai factory next --card t1538` → `refused serial-slot`(설치 빌드에 t1513 수리 `de388878e` 미반영). t1522식 영수증 경로(리더 배차=권한)로 진행, 크론 재시도 유지. 본 SPEC의 M1 receipt 게이트가 이 경로의 도구적 뒷받침이다.
- **앵커 검증 완료**: 리포트(main@ec13872f3) 앵커 14건을 본 트리 기준 전수 재확인 — 13건 일치, 1건 부분 이동(todo_analysis.go:74는 기록 분기, 임계 상수는 internal/factory/backlog_analysis.go:33). 상세는 spec.md §B.2. RED-now probe 8건(P1-P8)은 acceptance.md §0에 verbatim 기록.
- **자체 해결 스코프 질문**: (1) 통합 경로 표기 — 미션 문구 "PR to main per GitHub Flow"를 저장소 현행 상태(카드 PR은 develop 대상, PR #1748 선례, gitflow 전환 경과)에 맞춰 "통합 브랜치(develop)로 PR"로 기록; (2) "receipt" 명칭 충돌 — todo.go:818/:873의 스토어 발급 receipt와 구별되는 이름(LeaderApproval 계열)을 plan §D에 명시.

## §E.1 Plan-phase Audit-Ready Signal

plan_status: pending-audit
plan_complete_at: (plan-audit 통과 시 기록)
비고: 산출물 3종(spec/plan/acceptance) + 본 파일. Tier M, REQ 16건, AC 16건. plan-auditor 판정 대기 중.

## §E.2 Run-phase Evidence

(run-phase에서 작성 — manager-develop 소관. 첫 M1 커밋 시 spec.md frontmatter가 draft → in-progress로 전이된다.)

## §E.3 Run-phase Audit-Ready Signal

(run-phase에서 작성 — manager-develop 소관. E1-E7 자체검증 매트릭스는 plan.md §E 형식.)

## §E.4 Sync-phase Audit-Ready Signal

(sync-phase에서 작성 — manager-docs 소관. sync_commit_sha는 sync 커밋에서 pending-backfill로 기록 후 후속 커밋에서 백필.)

# t413 verdict — SPEC-GRAPH-REPORT-001 (graph report toolchain)

- 카드: t413 · 브랜치: `WT-graph-report` · HEAD: `03f5e97ce` (origin/develop `b6231290d` 기준 12커밋 선행, 미푸시)
- 상태: plan → run(4/4) → sync 3단계 마감 완료 · spec.md `status: completed` · 트리 클린

## Claim (주장)

SPEC-GRAPH-REPORT-001의 4개 산출물(shortest_path MCP 도구, `moai graph report`, edges 수축 가드, 지연 SessionStart edges refresh)이 구현·검증·문서 동기화까지 마쳤고 3단계 마감 커밋이 브랜치에 착륙했다.

## Evidence (증거 — 명령/산출물은 본 워크트리 실측)

- **plan**: 감사 3반복 — iter-1 FAIL 0.625(D1-D12) → iter-2 PWD 0.9375(D17-D21) → iter-3 PWD 0.875 MP 7/7(정지 신호는 점수 기준 변경 기인, 기재) + fold-in D22-D29 → `plan_status: audit-ready`. 리포트: `.moai/reports/plan-audit/SPEC-GRAPH-REPORT-001-review-{1,2,3}.md`
- **run** (§E.2 E1-E8 귀속 표): M4 `258013d0a`(지연 refresh DI seam) · M1 `31566c117`(graph_shortest_path, 카탈로그 29번째) · M2 `2cea86ec2`(architecture report, SCC·디렉터리 프록시) · M3 `cfe86675c`(수축 가드 — 쿼리 경로 `graph_refresh_cli.go:78` + 빌드 경로 `graph.go:348` 배선 실측). 오케스트레이터 독립 재검증: 각 마일스톤 스코프 diff 실측, 테스트 재실행(hook ok 33.5s / cli 신규 3건 PASS / graph M1·M2·M3 셀렉터 ok), windows 빌드 OK, lint `0 issues.`, `internal/hook` M2·M3 단계 diff 0행.
- **sync**: `00a911182`(CHANGELOG + status completed + §E.4) → 백필 `03f5e97ce`(sync_commit_sha + 도구 수 29 정정 — `audit_multi`가 `auditMultiToolName` 상수 등록이라 리터럴 grep 사각지대였음을 특정, RegistrationMatchesCatalog PASS).
- **SPEC lint**: `✓ No findings` (마지막 실측, 백필 후).

## Baseline-attribution (귀속)

모든 수치는 본 세션·본 워크트리(`.claude/worktrees/t413`)의 관측. 흡수 기준: `58fbc3b5e`(1차) → `b6231290d`(2차, t412 착지 포함 — 6916ee5a7 조상 판정으로 게이트 개방 확인).

## Gaps (미검증)

- 전체 스위트·CI 매트릭스는 미실행(레인-로컬 규율) — develop push 후 원격 CI가 전체 판정.
- nocgo 런타임 거동은 빌드·단위 테스트 수준만(CI 몫).
- windows 실행 환경 검증 없음(빌드 수준만).

## Residual-risk (잔여 위험)

- M2 fan-in 카운트가 원시 Source 기준(패키지가 import+호출 동시 사용 시 2로 계산) — REQ 문자 그대로이며 §E.2에 공개.
- deferred refresh는 프로세스 수명 축의 best-effort(세션 조기 종료 시 도중 중단 가능) — SPEC 공개 한계, 쿼리시 refresh가 liveness 보장.
- M3 매니저가 HEAD-baseline 확인을 위해 /tmp에 raw `git worktree add`를 일회 사용(읽기 전용·즉시 삭제·브랜치 무영향) — 런처 규율 위반으로 자진 보고됨, 재발 시 주의.

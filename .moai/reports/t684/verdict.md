# Card t684 Verdict — 착지 후 카드 자동 완료 (auto-done on land)

- Date: 2026-09-13 · Lane: lane-4 (Factory) · Branch: `WT-auto-done-on-land` (unpushed)
- SPEC: `SPEC-TODO-LAND-AUTO-DONE-001` (draft → completed, 3-phase close, v0.4.0)
- Base at authoring: origin/develop `74d872aaf`. Audit-time note: origin/develop moved to `5e0f71175` mid-audit — absorb merge + re-measure in the merge tree owed at the integration window.

## Claim

`moai todo auto-done` — 리드가 develop push 확인 직후 실행하는 착지 스캔 명령 — 이 구현됐고, 17/17 AC 대응 테스트가 이 트리에서 PASS로 재관측됐으며, sync-auditor 독립 감사 91.9/100 PASS. plan-auditor 진입 판정 PASS (0.9375, iter3).

## Evidence

| Phase | Result | Evidence |
|---|---|---|
| plan | PASS 0.9375 (0.75 → 0.875 → 0.9375, 결함 D1-D7·E1-E3 수리) | 감사 판정 3회, 최종 진입 판정 = run-gate 재확인 E3 스코프 |
| run | AC 17/17 PASS, 커밋 `a5c0c0eca`·`286a6713a`·`0ece9ef72` | `go test ./internal/kanban/ -count=1` ok 172.8s; `go test ./internal/cli/ -run 'TestTodoAutoDone' -count=1` ok 30.5s (21/21); 전체 cli 패키지 1건 FAIL(TestAuditLagUsesBinlagSeam — 신규 merge-base 도달성 호출) → 기준선 선언 후 `-run 'TestAuditLagUsesBinlagSeam\|TestTodoAutoDone'` ok 재확인; `go build ./...` 0; windows 빌드 0; vet clean |
| sync | 커밋 `61ef78596`·`0673b3ac0` | `make build` exit 0 (catalog 재생성); 4로케일 `grep -c auto-done` = 2/2/2/2; todo.md 본사본↔미러 diff identical; sync 커밋 후 `git status --short` clean |
| sync-audit | **PASS 91.9/100** (기능 95/보안 95/품질 88/일관성 90, 조화평균) | 독립 재실행: kanban `-run 'AutoDone\|Landed'` 35 PASS/0 FAIL/0 skip; cli `-run 'TestTodoAutoDone'` 21/0/0; gofmt·vet·lint(수정파일 0건) clean; 결정 코어(`AutoDoneDecide`/`AutoDoneDistinctTexts`) 커버리지 100%, CLI 동사 경로 ≥89.7% |

- 상세 감사 보고서: `.moai/reports/t684/sync-audit.md` (같은 브랜치에 커밋)
- 신뢰 가능한 diff 범위: `74d872aaf..HEAD` (22 files, +3094/−12) — 감사 시점 기준
- 핵심 위험 4종 독립 재검증: (a) 재발행 충돌 fail-closed + 기록-SHA 구제 (b) 말뭉치 재고정 = 정확히 t68만(쉼표형 실측) (c) identity-test 기준선 = 자기 코드 선언(제거 시 FAIL) (d) 닫기 배타성 = `ArchiveCard` 비테스트 호출처 전체 2곳(done, auto-done)

## Baseline-attribution

모든 수치는 이번 레인 세션에서 이 워크트리(`.claude/worktrees/t684`, base `74d872aaf`)에서 실행한 명령의 관측값. CI 판정은 리드의 develop push 이후 원격 몫(§4.1 — 레인 push 없음).

## Gaps

- 전체 스위트/CI 미실행(로드 규율) — develop push 시 CI가 전체 판정.
- `moai todo auto-done` 실큐 라이브 실행 미수행(닫기는 리드의 행위 — 테스트 증거로 대체).
- 성공 경로 fetch 엔드투엔드(F7), hugo 사이트 빌드 미실행.
- AC-AD-004/005 픽스처 전제: 스토어 불변식(`todo_identity.go:293-313`)이 live+archived 동일 id를 거부 → 충돌 close는 결정 계층+`--dry-run`으로 검증(감사 판정: AC 의도 미검증 잔여 없음).

## Residual-risk

- 부정 가드는 주제 스트림 한정(§D.1 기록 — SPEC-TODO-LANDING-ATTRIBUTION-001 소관).
- M1 cross-store 재발행은 `DistinctTexts=1`로 표시(단일 레코드 카운터 — progress.md §E.3 문구 과대 기재, F1 참조).
- 커밋 5개 중 2개가 `(t684)` 트레일러 누락(F4).
- README 4로케일에 `auto-done`/`landed` 미반영(공개된 후속 — 별도 카드 권장).

## 후속 권장

1. README 4-locale todo 동사 목록 갱신(후속 카드)
2. F1–F7 선택 결함(sync-audit.md 참조) — 필요시 별도 카드
3. 리드 절차 문서(`.claude/skills/moai/workflows/todo.md` + 미러) — 병합 후 리드 측 첫 운용 시 `--dry-run` 선행 권장

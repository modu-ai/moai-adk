# Card t658 Verdict — 홈 SQLite 단일 원천 확정(진술 수리 + 재발 가드)

- Date: 2026-09-13 · Lane: lane-4 (Factory) · Branch: `WT-home-queue-canonical` (unpushed)
- SPEC: `SPEC-TODO-QUEUE-HOME-CANON-001` (draft → completed, 3-phase close, v1.2.0)
- Base: b66789479 (리드 지명과 정확 일치) · 계획 감사 기록: `.moai/reports/t658/plan-audit.md` (첫 커밋 조건 이행)

## Claim

큐 경로 해석은 이미 홈 SQLite 단일 봉인으로 수렴돼 있었고(카드 전제 일부 만료 — statusline 독자 기수렴), 잔여 결함인 foreman 스킬 큐 감시의 죽은 경로를 수리하고 재발 방지 가드 2종을 세웠으며, sync-auditor 독립 감사 98.7/100 PASS.

## Evidence

| Phase | Result | Evidence |
|---|---|---|
| plan | 진입 PASS 0.95 (3회 반복: FAIL→MP-7 토큰 수리→N1/N2 delta→PASS, 상한 연장은 감사관 권고+조정자 승인) | `.moai/reports/t658/plan-audit.md` |
| run | AC 6/6 PASS, 커밋 `d26091f5b`·`c5ef1abf1` | 가드(a) RED(결함 문서)→GREEN→뮤테이션 RED→GREEN, 초기 GREEN에서 실제 조각 버그(`tr -c` 개행) 포착 후 수리; 가드(b) 1218파일 스캔 예외 정확히 2곳, 주입 위반 RED 관측; kanban 패키지 220.0s ok; make build 0; 중립성 grep 전부 0건 |
| sync | 커밋 `da3e72ea5`·`a2590d81a` | CHANGELOG Fixed 1항목, frontmatter 전환 정합, 템플릿 무변경 실측 |
| sync-audit | **PASS 98.7/100** (기능 100/보안 100/품질 95/일관성 100) | 6 AC 전건 재관측 + 변이 제어 2건 독립 RED 재현 + **프래그먼트를 실제 git 저장소·링크드 워크트리에서 실행해 StateDirForRoot와 3층 일치 확인** + catalog 해시 dry-run 재계산 일치 + make build 후 0-drift |

- 감사 보고서: `.moai/reports/t658/sync-audit.md` (커밋 예정)
- 재측정 범위: `b66789479..HEAD` (foreman SKILL 2벌·catalog·kanban 테스트 4종·SPEC artifacts·CHANGELOG)

## Baseline-attribution

모든 수치는 이번 레인 세션에서 이 워크트리에서 실행한 명령의 관측값. 전체 스위트는 CI 몫(§4 규율). push는 리드 일괄.

## Gaps

- POSIX-sh 가드 테스트는 Windows에서 건너뜀(판정면 = darwin/linux CI)
- `--separate-git-dir` 체크아웃과 비ASCII basename 저장소는 조각 근사치 미측정(F1/F2 — 기록됨)
- CHANGELOG의 RED/RED/GREEN 서술은 run 기록 인용(sync 트리 재실행 아님)

## Residual-risk

- F1: 비ASCII basename 저장소에서 감시 디렉터리 분기 가능성(byte `tr` vs rune 정규화) — 기록된 협소 재발 경로
- t706(임시 루트 3종 앵커)이 선행 착지 예정 — 겹침 없음 실측(변경 파일 목록·homestate 참조 grep 0건), 흡수 시 충돌 예상 없음

## 후속

- M5(t657 SPEC, 프로젝트 스토어 정리) — 별도 승인 대기
- t835(병합 검증기 결함+t538 참조+t204/t718) — 리드 발행됨

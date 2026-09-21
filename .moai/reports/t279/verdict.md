# Lane Verdict — card t279 / CR #1665 스레드 처분 라운드 (lane-3)

작성: 2026-08-27 · 트리 `.claude/worktrees/t279` · 브랜치 `WT-t250-followup` @ `0daa30355`

## Claim (주장)

1. 카드 원문 과업(연기 판정표·채택군 실행·#1648 스레드 정리)은 **이전 세션에서 이미 완료**돼 있었음을 직접 관측으로 확인했다 — 대표표본 10건 코드·SPEC·GitHub 양측 검증(#1648 스레드 77건 전원 resolved).
2. 본 라운드의 실질 잔무는 **PR #1665 위 신규 미해결 스레드 13건**이었고, manager-spec(t279-docfix)이 3커밋으로 처분했다: 12 FIXED(+1건에 FOLLOWUP 동반) · 1 BLOCKER-summary(#7 AC-GF-022).
3. 통합 준비 완료 — 증거 포함 전부 커밋 상태(heads: `6b3d8e9d2` → `811818bde` → `0daa30355`).

## Evidence (증거 — 2026-08-27, lane 세션 af9f2ca2 이번 실행)

**코드 배치(문서 수정 전 트리 `b9dd6a9e4`, 코드 불변이므로 유효 귀속 유지):**

| 항목 | 명령 | 결과 |
|------|------|------|
| 빌드 | `go build ./...` | exit 0 |
| 스코프 테스트 | `go test -count=1 -p 4 ./internal/{graph,cli,hook/quality,mx,config,navigator/astx}/` | 6패키지 전부 `ok` (graph 6.6s / cli 183.6s / hook.quality 7.4s / mx 1.9s / config 1.9s / astx 1.1s) |

원문 로그: 워크트리 내 `.moai/state/verify/af9f2ca2/`(head.txt·test.log)

**문서 처분 독립 검증(수정 후):**

| 확인 | 방법 | 결과 |
|------|------|------|
| 코드 무변경 | `git diff --name-only b9dd6a9e4..HEAD` 비-md/yaml 필터 | 공집합 |
| 개인정보 적출 | verify-docs·SPEC-002 plan.md `grep goos\|/Users/` (비-$HOME 행) | 0건 |
| AC-GF-020 강화 | acceptance.md:130 문구 직접 열람 | non-exported 선언 load-bearing 요구 확인 |
| 인벤토리 포인터 | shipped_key_inventory.yaml:394 | graph_refresh_cli.go:77 (graphRefreshBudgetMS) 로 교정 확인 |
| 파이프 이스케이프 | verify-graph.md:24 원문 열람 | `\|\|` 인라인코드 리터럴 처리 확인 |
| SPEC lint | `moai spec lint` ×2 (내 실행) | 둘 다 `✓ No findings — all SPEC documents are valid` |

## Baseline-attribution (귀속)

- 모든 명령: 이번 실행, `WT-t250-followup` 트리, 환경스크럽(MOAI_ 접두 7종 unset) 단일 복합 호출.
- 코드 테스트 베이스라인 = `b9dd6a9e4`; 이후 변경은 md/yaml 전용임을 위 diff로 측정했으므로 결과가 현재 HEAD(`0daa30355`) 코드에 유효.

## Gaps (미검증)

- 전체 스위트 — gitflow 프로토콜 §8 에 따라 origin/develop CI 소관.
- **AC-GF-022(#3865025070) 종결 판정 — 연산자 결정 보류 중**(아래 Residual-risk), 기계 검증 불가 항목(소급 baseline 날짜 부재).
- #1665 스레드 reply+close 는 develop push 후 수행 예정(FIXED 12건만, #3865025070은 의사결정까지 오픈 유지).
- main 최신 스탬프 건강도 실측 안 함 — t291/t292 소관.

## Residual-risk (잔여 위험)

- **#7 [Major] SPEC-001 status=completed vs AC-GF-022 MUST 미달**: §D.1 자체 닫기규칙("All MUST ACs PASS")과 충돌 — 소급 불가능한 측정 특성상 ① 기록된 편차로 수용(t289 선례 pass-with-debt 계열) / ② completed 유지+예외 조항 명문화 / ③ 재오픈 중 연산자 선택 필요. 현행: progress.md §E.4 편차 기록만 추가, status 무변경.
- FOLLOWUP 후보 1건 (#11): closer가 sync_commit_sha를 SHA 형식 없이 수용 — 코드 변경 수반이라 본 카드 제외.
- 병합 발산 `25 13`(origin/develop 기준) — develop 쪽 진행분과의 텍스트 충돌 가능성. 레인이 해결하고 의미 충돌 시 blocker.

## Disposition 요약 (13건)

FIXED: 25031 25021 25038 · 25045 25050(stays-MUST 사유 기록) 25074 25081 25092 25100 25108(+FOLLOWUP) 25115(GFR-016 폐쇄노트) 25121 │ BLOCKER-summary: 25070(연산자 결정) │ 상세: 회신 원문 트리아지 기록 참조(브랜치 3커밋)

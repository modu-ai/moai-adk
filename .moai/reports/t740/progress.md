# t740 progress — run → sync 기록

카드: t740 (Class B) · 브랜치: `WT-spec-lint-gate` · 배차: 리드, 2026-09-14

## Run phase

원인 확립 경과 (명령 + 출력 요지, 전체 인용은 verdict.md §Evidence):

1. 세션 상태 확인 — primary 체크아웃(main)에서 출발, ExitWorktree 불요 확인.
2. `EnterWorktree(t740)` → `git merge develop` (fast-forward, c044a911f → b1bd81b23) → `git branch -m WT-spec-lint-gate`.
3. CI 재측정:
   - `gh run list --workflow "SPEC Lint" --branch develop --limit 8` → 재기준선 후 success 5건.
   - `gh run view 34088825415 --log-failed` (09-07 재현 run) → `0 error(s), 4344 warning(s)` → exit 1, 단계 "Run SPEC lint".
   - `git show d4162b368:.github/workflows/spec-lint.yml` → 당시 명령 `go run ./cmd/moai spec lint --strict` — **경고 승격 플래그 확인, 원인 (b) 확정**.
   - `gh api .../actions/runs/34775055880` → a404132e7 run = completed/success (gh run list 의 빈 conclusion 필드를 failure 로 읽은 1차 오독 정정, `gh run view` 스테일 뷰 2차 오독 정정 — REST API 가 권위).
4. 로컬 재현/검증 (b1bd81b23):
   - `go run ./cmd/moai spec lint --baseline .moai/spec-lint-baseline.json` → **rc=0**, `0 error(s), 3147 warning(s)`, `baseline: OK`. 전문: `lint-baseline-b1bd81b23.txt`.
5. 부수 항목: SpecsDirMissingSpecFile 0건 (9fb52f746, t525 M4 가 폐쇄).

## 결론

- 수리 불요 — 카드 명제의 수리는 t525 가 09-08~09-11 착지 완료. 현재 게이트는 신호 정상(초록).
- 코드 변경 0건. verdict.md 가 카드의 유일한 산출물(증거 동봉).

## Sync phase

- 변경 파일: `.moai/reports/t740/verdict.md`, `.moai/reports/t740/progress.md`, `.moai/reports/t740/lint-baseline-b1bd81b23.txt` — 문서 3건, Go 코드 0건.
- 재측정 범위: Go 코드 변경 없음 → 영향 패키지 테스트 불요(측정 대상 자체가 lint 실행이며 그 출력이 증거).
- plain(error-only) 경로 rc: **rc=0**, `0 error(s), 3147 warning(s)` (2026-09-14 백그라운드 측정 완료) — 두 게이트 경로 모두 초록.

## 남은 절차

- 커밋 → 리드 완료 보고 + 통합 창 요청 (push·CI 직접 요청 금지 준수, t700 워크트리 원격 착지 전까지 유지는 리드 소관 그대로).

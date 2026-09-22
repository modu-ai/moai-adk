# t527 verdict — 페이지 전체 첫-등장 앵커 스윕 (SPEC-WEB-ANCHOR-SCOPE-001)

- card: t527 (lane-12, Factory) · spec: SPEC-WEB-ANCHOR-SCOPE-001 · branch: `WT-anchor-scope-sweep`
- base: `bf779ecf2` (= develop 팁, codex 미러 존재 — `1aaf4951f` 조상 포함 실측)
- run 커밋: `014861ee2` (M1) · `15c1f739c` (plan 산출물) · `1e2e22729` (M2 RED 증거) · `cebc79cdd` (M3 종결, `run_commit_sha`) · `ccef6f06b` (SHA 백필)
- sync: 3-phase close는 sync 커밋에 실림 (주제 `docs(SPEC-WEB-ANCHOR-SCOPE-001): sync-phase artifacts + 3-phase close (t527)` 형태) + D3 백필 커밋 1개

## Claim (주장)

1. 원시 상위집합(이 트리 `bf779ecf2` 직접 측정): `internal/web/*_test.go`의 `strings.Index(` = **60사이트 / 19파일**. 카드의 34/15는 판별식 이전의 잠정 상한이었고, 리드 재측정치(55/18 @ `0b1e27877`)·lane-1 측정치(60/19)와의 차이는 트리 차이로 문서화됨(research.md §1).
2. 판별식(research.md §4, 내구 문서화): **(c)참 ⇔ 노출 ∧ 복제 ∧ 순서** — ① 수신자가 전체 GET /settings 렌더링(미러는 탭 8/14 위치에서만 등장) ② needle이 미러의 텍스트 인벤토리에 속함 ③ 미러가 의도 대상보다 앞서 렌더링. 구조적 안전: 미러는 `name=`·`<form`·`data-panel=`·`<script>`를 배출하지 않음.
3. 분류표(research.md §5, 60행 전수): **대기 (c)참 = 0** · t509가 이미 수리한 역사적 (c)참 = 1 (`mcp_console_test.go:115`) · 노출-(c)거짓 = 36 · 미노출 = 23.
4. 판별식은 뮤테이션으로 반증 시도 후 생존(AC-WAS-005): t509 수리부를 전체-본문 앵커로 되돌리자 `TestMCPConsoleWriteCapableTextDistinction`가 **RED**(exit 1) — `codex_task`·`codex_job_cancel`의 배지가 없는 미러 행에서 판정됨(복제∧순서 예측 그대로), 비미러링 도구(`goal_arm`·`verify_snapshot`)는 초록 유지(복제 조건의 선택성 확인). 복원 후 `git diff --stat -- internal/` 빈 출력, 패키지 테스트 `ok`.
5. 운영자 게이트(2026-09-07) 확정에 따라 **수리 없는 종결**(REQ-WAS-006): 코드 변경 0건 착지. plan-audit FAIL(0.97)은 미해결 마커(D1) 단일 사유였고 게이트에서 해소 → 델타 **PASS 0.99**.

## Evidence (증거)

- M1 통화: `grep -o 'strings\.Index(' internal/web/*_test.go | wc -l` → `60`, `grep -l … | wc -l` → `19` — §5 표 60행과 file:line 1:1, 드리프트 0 (progress.md §E.2).
- M2 RED 원문: `.moai/reports/t527/m2-mutation-red.txt` (명령·트리·타임스탬프·exit 1·원문 출력).
- 미러 존재 하 회귀: `go build ./internal/web/` exit 0 + `go test -count=1 ./internal/web/` → `ok 5.912s` (동일 트리).
- Lint: `golangci-lint run ./internal/web/... --timeout=2m` → `0 issues.`
- AC 행렬 7/7 PASS (progress.md §E.3 YAML).

## Baseline-attribution (귀속)

위 모든 수치는 이번 실행·이 트리(`bf779ecf2`)에서 직접 측정. 감사(plan-audit.md)는 독립 재측정으로 60↔60 대응·최고위험군 스윕(미러링 needle + 소유 패널 탭 8 뒤 + 노출 수신자 = 정확히 1건 = t509 행)·구조적 비배출 grep 0을 확인.

## Gaps (미검증)

- 로컬 판정은 레인 범위(`./internal/web/`) — 전체 스위트 판정은 리드 일괄 push가 일으키는 `origin/develop` CI 소관.
- go build/vet 전체(`./...`)는 plan/run 양상에서 불필요하여 미실행(코드 변경 0건; 범위 이유 기록됨).

## Residual-risk (잔여 위험)

- 미래 codex 필드 추가가 복제 폭을 조용히 넓힐 수 있음(런타임 가드 없음 — 판별식이 조건족에 키되어 재측정 훅으로만 감지).
- 미래 탭 재정렬이 순서 조건을 뒤집을 수 있음(탭 순서는 자체 시험 `wantTabOrder`가 지킴 — 코드가 아닌 검토 경로).
- 계열 교훈(「판정 대상이 존재하되 다른 것이다」 세 번째 변형)은 `feedback_confident_verdicts_over_empty_sets.md`에 종결 시 추가 예정 — 카드 [HARD] 의무.

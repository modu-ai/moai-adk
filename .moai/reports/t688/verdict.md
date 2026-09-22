# t688 판정서 — SPEC-GRAPH-STAMP-ANCESTRY-001

- card: `t688` · lane-9 (Claude 세션 `aa752bc3-a199-4c10-8ab9-4ecb5a3f912f`)
- branch: `WT-graph-stamp-freshness` · 인계 HEAD `30d505ac4` → 최종 HEAD `860685326` (verdict 커밋 포함)
- 미푸시: 15 (origin/develop 대비 `126 15` — origin 이 그 사이 126 앞으로 이동) · push 없음 · PR 없음

## Claim

plan 재개 → plan-audit PASS → Kickoff 승인 → run(M1~M5) → sync → 3단계 close → sync-audit PASS 까지 카드가 워크트리 안에서 끝났다. 두 결함 분리(스탬프 비조상 = exit 2 미측정 / 값 초과 = exit 1 stale)가 보존됐고 warning-only·자동 재스탬프는 들어가지 않았다.

## Evidence (레인이 직접 관측한 것)

| 항목 | 명령 | 관측 |
|---|---|---|
| plan-audit | plan-auditor 보고서 | PASS 0.974, blocking 0 |
| 빌드·vet | `go build ./internal/graph/... ./internal/cli/...` · `go vet` 동일 범위 | exit 0 / exit 0 (HEAD a810ee75a) |
| SPEC 검사 | `moai spec lint … --strict` · `moai spec audit --filter-spec …` | 0 findings · drift 0 (HEAD 2f984c4b4) |
| 3단계 close | `moai spec close … --backfill-only` | commit 2f984c4b4, status completed, sync_commit_sha 5ea0bac5f |
| sync-audit | sync-auditor 보고서 | PASS 0.914 (F 0.96 / S 0.92 / C 0.90 / Cs 0.88), blocking 0 |

에이전트가 관측하고 레인이 귀속만 한 것: AC 12/12 PASS · RED 원장 4본 GREEN · M4 두 독립 관측(`merge-base --is-ancestor 7097e6e21 HEAD` exit 0, `graph check` codemaps value=2 < 40 fresh) — 원문은 `progress.md §E.2` 와 `.moai/reports/sync-audit/…sync-audit.md`.

## 커밋 계보 (30d505ac4 이후)

7927c0825 plan 보강 0.2.0 · e247612b3 증거 기록 · 8275c82a5 plan D1 · b6e84fe5c M1 · c613b7c6b M2 · eced8eb38 M4 · dfe04c9b8 / a810ee75a M5 · 89237b4ce plan §D.1 정합 · 5ea0bac5f sync · 2f984c4b4 close · 4508b490b CHANGELOG 정정

## Gaps

- `go test ./...` 미실행(규율) · windows 크로스빌드 미실행 · workflow 가드는 합성 이력 대상 실행뿐(실제 GitHub 이벤트 미검증) — CI 몫.
- golangci-lint 37건은 카드 무관 선재(변경 파일과 겹침 0) — 병합 베이스에서 재측정하지 않음.
- 게이트웨이 400/502 하위 원인은 이 카드에서 진단하지 않음.

## Residual-risk (통합 창에서 리드가 볼 것)

1. **통합 후 codemaps 재stale**: origin/develop 이 분기점보다 112 앞이고 described-worthy 파일 58개가 움직여, develop 흡수 시 `described-source-diff` 가 40을 다시 넘는다(조상성은 유지). 병합 트리에서 재측정 후 흡수 tip 으로 재스탬프하면 value 2 로 복귀 — 본문은 이미 재생성됐으므로 bare restamp 가 아니다.
2. **판정 순서 결정**: plan §D.1 의 문자 그대로(본문 부재 최우선)는 기존 테스트 `TestGraphCheckCmd_NotComparableExitsTwo` 를 깨서, spec §D 1행(C1 동결)에 따라 해석→본문→조상성 순서로 구현·문서화했다.
3. docs-site `cli-reference/graph.md` 는 push 경로의 HEAD 조상성 판정을 아직 적지 않음(거짓은 아님, 후속 카드 후보).

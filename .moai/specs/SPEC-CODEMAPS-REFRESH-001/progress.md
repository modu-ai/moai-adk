---
id: SPEC-CODEMAPS-REFRESH-001
title: "progress.md"
version: "0.1.1"
created: 2026-09-02
updated: 2026-09-02
author: manager-spec
phase: "v3.2.0 target"
module: ".moai/project/codemaps"
tier: M
---

# progress.md — SPEC-CODEMAPS-REFRESH-001

## §E.1 Plan-phase Audit-Ready Signal

- Tier: M (근거: 다중 마일스톤 구조 + 정확성 검증의 독립 AC 열거 필요 + 스탬프 고아화 회피 절차 규율. LOC는 소규모이나 검증 깊이와 실패모드 면에서 S 상한 초과 → plan-audit skip 임계 0.80 적용)
- Artifact set: spec.md / plan.md / acceptance.md (Tier M 3종) + tracking용 progress.md
- GEARS 준수: REQ-CMR-001~008 (Ubiquitous / While / When / Where / Unwanted 패턴 사용, IF/THEN 없음)
- 기준선 실측(2026-09-02, 워크트리 t432 @ ad272be20): `moai graph check` codemaps value=60 threshold=40 verdict=stale (contribution 13 described-worthy files vs first parent f3e11e113) · mx-index/edges absent(신규 worktree 예상 상태) · `go list ./...` 137 · 팬텀 6 디렉터리 부재 · `graph stamp --commit` 플래그 존재(`internal/cli/graph_stamp.go:68`)
- Out of Scope: t304 소관 팬텀 수정 / 임계값 재보정 / Go 코드·신규 툴링 — `### Out of Scope` H3 3건으로 명시
- NEEDS CLARIFICATION 마커: 0건 (카드 의도·범위·기준선 모두 실측 확정)
- plan-audit iter-1: FAIL 0.81 (D1 BLOCKING) → 수리 완료(D1~D6 전부 적용, 2026-09-02) — iter-2 delta 재감사 대기

## §E.2 Run-phase Evidence

### AC-by-AC 판정 행렬 (M1→M4, 2026-09-02, 트리 a87e8ec2c / worktree t432)

| AC | 상태 | 검증 명령 | 실제 출력 |
|----|------|----------|----------|
| AC-CMR-001 재생성 완전성 | PASS | `ls .moai/project/codemaps/` | 7개 항목 (6문서 + provenance.json) — overview·modules·dependencies·entry-points·data-flow·docs-truth 전부 재생성 (data-flow는 3개소 정정, 나머지 5문서 전면 갱신) |
| AC-CMR-002 경로 실존 표 | PASS | `grep -ohE '\b(internal\|pkg\|cmd)/...' *.md \| sort -u` + `test -e` 전수 | 유니크 경로 100 (93 EXISTS + 7 ABSENT) — absent 7건 전부 분류 (t304 known-6 = 6, 부정 인용 각주 bodp = 1). 증거: `.moai/reports/t432/codemaps-accuracy-verification.md` §1 |
| AC-CMR-003 패키지 구조 대조 | PASS | `go list ./... \| wc -l` = 137 + 양방향 comm 대조 + 45토큰 하위-목록 grep | 진짜 누락 0, 유령 = known-6 전부(t304), scripts 3개는 described_roots 밖. 증거: 동 파일 §2 |
| AC-CMR-004 식별자 hit/miss | PASS | 식별자별 `grep -q ... && echo HIT \|\| echo MISS` (27항목) + docs-truth §1 12행 양방향 | HIT 26 / 진짜 MISS 1 (`ListActive` — session API가 `Query`/`QueryActiveWork`로 진화, 기록만·본문 무수정), 카탈로그 12/12 일치. 증거: 동 파일 §3 |
| AC-CMR-005 스탬프 도달성 | PASS | `moai graph stamp codemaps --commit ad272be20abff9e4f3b1b363fce3e48dac4c5132` | `OK: stamped .../provenance.json` + `git merge-base --is-ancestor <sha> origin/develop && echo REACHABLE` → `REACHABLE rc=0`. 새 sha = merge-base (worktree HEAD a87e8ec2c와 다름 — merge-surviving) |
| AC-CMR-006 게이트 종결 | PASS | `moai graph check` (아래 verbatim) | codemaps value=0 threshold=40 verdict=fresh; mx-index/edges verdict=absent (신규 worktree 예상 상태 — stale 아님) |
| AC-CMR-007 범위 위생 | PASS | plan §E4 2명령 (아래 verbatim) | 양쪽 무출력 — 변경이 codemaps/reports/t432/state/specs/SPEC-CODEMAPS-REFRESH-001에 한정 |
| AC-CMR-008 증거 독립성 | PASS | 증거 파일 5섹션 + 게이트 출력 병존 | `.moai/reports/t432/codemaps-accuracy-verification.md` (§1-§5) + 아래 게이트 verbatim — 정확성 증거와 게이트 판정이 별도 근거로 병존 |

**AC-CMR-006 게이트 verbatim 출력** (2026-09-02, 재스탬프 직후):

```
codemaps  metric=described-source-diff value=0 threshold=40 verdict=fresh
mx-index  metric=inventory-content-diff value=0 threshold=1 verdict=absent  (mx-index absent (untracked runtime artifact — fresh worktree state))
edges     metric=source-fingerprint-mismatch value=0 threshold=0 verdict=absent  (edges.jsonl absent (untracked derived artifact — fresh worktree state))
graph check: layer mx-index verdict=absent value=0 threshold=1 — mx-index absent (untracked runtime artifact — fresh worktree state)
graph check: layer edges verdict=absent value=0 threshold=0 — edges.jsonl absent (untracked derived artifact — fresh worktree state)
```

**AC-CMR-007 범위 위생 verbatim** (plan §E4 — tracked 변경 필터 + untracked 필터, 양쪽 무출력):

```
$ git diff --name-only ad272be20 | grep -v -E '^\.moai/(project/codemaps|reports/t432|state|specs/SPEC-CODEMAPS-REFRESH-001)/'
(무출력)
$ git status --porcelain | grep '^??' | cut -c4- | grep -v -E '^\.moai/(project/codemaps|reports/t432|state|specs/SPEC-CODEMAPS-REFRESH-001)/'
(무출력)
```

### M1 재생성 요약

- 6문서 전부 재생성 (한국어 문서 컨벤션 보존): overview·modules·dependencies·entry-points 전면 갱신, data-flow 3개소 정정(Linter 규칙 수 2개소 + exit 코드 계약 1개소), docs-truth 전면 갱신(§1 카탈로그 12개, §4.1 CLI 실측, §4.2 16 commands, §5 GLM 5.3-flash).
- 주요 실측 반영: non-test 1074 / test 1714 / ~229.4k LOC, `go list` 137 (기준선 불변), root 등록 60건, 팬-인 측정 방법 명시 교체(`go list -f '{{range .Imports}}'` 기반 — config 27, defs 17 등), GLM 기본 모델 `glm-5.3-flash`, 에이전트 12 retained (manager-lead 포함), `/moai` 명령 16개(todo 포함).
- known-6 서술은 전부 원문 이월 (t304 소관 — 수정 없음). `internal/factory` 섹션 포함.
- 신규 패키지 반영: navigator(6서브)/kanban/epic/sessionmsg/glmcred/paths/execerr/feedback/timing/report/planhtml/mirrornotice 등. `internal/bodp`는 트리 부재 확정(#1278 제거)으로 서술하지 않고 제거 기록 각주만.

### M2 검증 — new-findings 발췌

전체 10건은 증거 파일 §5 참조. 주요: `ListActive` 식별자 미적중(수정 대상 아님·기록만), `moai graph check` exit 계약 실측(0=all-fresh / 1=stale·absent / 2=system — graph_check.go:21-25, data-flow.md의 종전 "0/1/2=fresh/stale/absent" 표기 정정), 문서 간 규칙 수 불일치(18 vs 13+3)의 실측 19로 해소.

### M3 스탬프

- `moai graph stamp codemaps --commit ad272be20abff9e4f3b1b363fce3e48dac4c5132` → OK. provenance: commit_sha = merge-base (ad272be20abff9e4f3b1b363fce3e48dac4c5132), tree_root = 본 워크트리, generated_at = 2026-09-02T05:31:10Z.
- 도달성: `git merge-base --is-ancestor <sha> origin/develop` → rc=0 (REACHABLE). bare-HEAD 스탬프 아님.

## §E.3 Run-phase Audit-Ready Signal

run_complete_at: 2026-09-02
run_commit_sha: "4548b947b"
run_status: complete
ac_pass_count: 8
ac_fail_count: 0
ac_pass_with_debt_count: 0
preserve_list_post_run_count: 0
l44_pre_commit_fetch: not-performed (worktree 격리 레인 — 리드 배차 하 run, push 없음)
l44_post_push_fetch: not-applicable (본 레인 push 없음 — develop 병합 창은 리드 소관)
new_warnings_or_lints_introduced: 0 (Go 소스 변경 0 — 문서 전용 SPEC)
cross_platform_build.linux: not-applicable (Go 소스 변경 없음)
cross_platform_build.darwin: not-applicable (동일)
cross_platform_build.windows: not-applicable (동일)
total_run_phase_files: 9 (codemaps 7 + reports 1 + spec.md frontmatter) + progress.md 1
m1_to_mN_commit_strategy: single-run-commit (M1-M4 통합) + progress.md 후속 커밋 — 문서 전용 Tier M

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_

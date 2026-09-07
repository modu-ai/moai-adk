# progress.md — SPEC-SPECLINT-GATE-SIGNAL-001

## §E.1 Plan-phase Audit-Ready Signal

plan_status: audit-ready
plan_complete_at: 2026-09-07
tier: M
artifacts: spec.md, plan.md, acceptance.md, progress.md
baseline_tree: dd1439502 (origin/develop tip at plan time)
baseline_measurement: 0 error(s), 4368 warning(s), rc=0 without --strict (.moai/reports/t525/spec-lint-baseline-dd1439502.txt)
kickoff_gate: 기제 모양 (i)-(iv) 선택은 운영자 결정 — plan.md §F M2 / spec.md §2.2

참고: plan-auditor 판정은 이 신호 이후 오케스트레이터가 수행한다.

## §E.2 Run-phase Evidence

### M1 — 판정 (2026-09-08, card t525)

측정 규율: 모든 수치는 **트리 빌드 `go run ./cmd/moai`**(installed 바이너리 e79c010b8 미사용 —
1069 커밋 지연, 리드 지시 2026-09-08). 카운팅/부재 판정은 `/usr/bin/grep`. attribution triple
(명령 / 관측 출력 / 트리 SHA) — 상세 원문 `.moai/reports/t525/{m1-demographics.md,verdict.md,census-b6efc874f.json}`.

| # | Claim | Command | Observed output | Tree SHA |
|---|---|---|---|---|
| M1-1 | 인구통계 재도출 (AC-SLGS-001) | `go run ./cmd/moai spec lint --json > .moai/reports/t525/census-b6efc874f.json` | rc=0, `jq length` → **4378** findings (warning 4378, error 0; advisory=true 4376 / advisory=false **2**) | `b6efc874f` |
| M1-2 | 비-strict 경로 초록 | `go run ./cmd/moai spec lint` | `0 error(s), 4378 warning(s)`, RC_NONSTRICT=0 | `b6efc874f` |
| M1-3 | (b) 구조 입증 (AC-SLGS-002 rc=1 쪽) | `go run ./cmd/moai spec lint --strict` | `0 error(s), 4378 warning(s)`, RC_STRICT=**1** | `b6efc874f` |
| M1-4 | 비-advisory 재고 = M4 쌍 | `jq -r '[.[] \| select(.severity=="warning" and (.advisory // false) == false)] \| .[] \| "\(.code)\t\(.file)"' census-b6efc874f.json` | `SpecsDirMissingSpecFile` × 2 (SPEC-V3R4-CC2X-ADOPT-001/002) | `b6efc874f` |
| M1-5 | 수치 동결 없음 (AC-SLGS-004) | `/usr/bin/grep -rnE '4[,.]?3(44|68)' .moai/specs/SPEC-SPECLINT-GATE-SIGNAL-001 internal/spec internal/cli` | 12 히트 — 전부 SPEC 문서 4건의 출처 명시 인용; internal/spec·internal/cli **0** 히트 | `b6efc874f` |
| M1-6 | t518 미착지 (REQ-SLGS-011 전제) | `git merge-base --is-ancestor 6cfcfef00 origin/develop` | 비-조상 (T518_NOT_LANDED) — 흡수 병합 전후 2회 재측정 | `b6efc874f` / `2cee65571` |

M1 측정 전 흡수: `origin/develop` `a849d99d2` → `19cf21408` 병합 (창 67 커밋, `internal/spec/` 변경 0).

**M1 판정 요약** (원문 verdict.md):

- **(a)** 성립·지연 — advisory 재고 4,376건(12 규칙)은 실재하는 부채; t518 착지까지 상환 금지 (REQ-SLGS-011).
- **(b)** 구조 입증 — error 0 + 비-advisory 경고만으로 rc=1. 단, 전제 교정: **비-advisory 재고는 2건**(M4 쌍)이며 "수천 건" 전제는 역사적 표본이 advisory 비-표시 표였던 데서 온 추론으로 기각. era demotion(`dd644b5e0`, 2026-07-21)이 이미 대량 재고를 흡수한 상태. M4가 쌍을 "왜"와 함께 닫으면 bare `--strict`는 0 비-advisory → 초록(신호 회복). M2 기제(iii)의 잔여 가치: 규칙별 델타 가시성 + t518 인구 이동 흡수(M3 게이트 재기준). — kickoff 승인 범위 변경 아님, 전제 교정 보고.
- **(c)** 소관 외 — advisory 경계 자체는 t518 소관; 측정된 t518 스코프 = advisory 표시 4,376건.
- **AC-SLGS-002 mutation 절반**: M2 관측 창 소관 — M1에서 미관측, pending으로 기록 (통과 아님).

크로스플랫폼 빌드: `go build ./...` exit 0, `GOOS=windows GOARCH=amd64 go build ./...` exit 0 (M1 커밋 전, tree `b6efc874f`).

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_

## §F Phase 4 Mode Selection

**Input parameters** (orchestrator, 2026-09-08):

- tier: M
- scope (files): ~8 (internal/spec/lint.go, internal/cli/spec_lint.go, .github/workflows/spec-lint.yml, baseline file, tests, docs)
- domain count: 2 (Go linter/CLI policy surface, CI wiring)
- file language mix: Go + YAML + JSON baseline
- concurrency benefit: LOW — coding-heavy, sequential dependency M1 verdict → M2 mechanism → M3 wiring
- Agent Teams prereqs: not requested

**Mode evaluation table**:

| Mode | Selected | Rationale |
|------|----------|-----------|
| direct | no | Multi-file Go + CI change, not a typo fix |
| serial | **yes** | Coding-heavy Tier M; M1→M4 ordered dependencies; single writer per milestone |
| fanout | no | Not research-heavy; coding-task parallelism caveat |
| sweep | no | <30 files, semantic (not mechanical-uniform) change |

Decision: serial
Justification: Anthropic coding-task parallelism caveat applies — the mechanism (M2) depends on the M1 demographic verdict, and CI wiring (M3) depends on the mechanism; nothing downstream of M1 is parallelizable. One manager-develop spawn per milestone, sequential.

Kickoff outcomes (operator decisions, 2026-09-08, AskUserQuestion round):

- kickoff: **approved** (plan-audit PASS 0.875 prerequisite met)
- mechanism (spec.md §2.2 reservation): **(iii) per-rule baseline ratchet** — CI gates via `--baseline <checked-in file>`; `--strict` retained for non-CI use
- progression mode: **autonomous** — ac_converge goal armed by the orchestrator after this gate; no per-milestone operator prompts; evidence reported at close

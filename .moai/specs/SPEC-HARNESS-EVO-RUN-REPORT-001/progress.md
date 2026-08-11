# SPEC-HARNESS-EVO-RUN-REPORT-001 — Progress

SPEC: 하네스 실행→학습 배선 (Epic Harness-Evolution 2/4) · Tier M · development_mode: tdd

---

## §E.1 Plan-phase Audit-Ready Signal

```yaml
plan_status: audit-ready-refresh
plan_complete_at: 2026-08-12   # v0.2.0 REFRESH (직전 2026-07-03 v0.1.0)
tier: M
artifact_set: [spec.md, plan.md, acceptance.md, progress.md]   # Tier M = 3-file + progress skeleton
req_count: 10        # REQ-HRR-001 ~ REQ-HRR-010 (불변)
ac_count: 10         # AC-HRR-001 ~ AC-HRR-010 (불변)
milestone_count: 5   # M1 ~ M5 (불변)
development_mode: tdd
depends_on: SPEC-HARNESS-EVO-PIPE-REPAIR-001   # completed a661da107
sibling_specs_added:
  - SPEC-HARNESS-LEARNING-EVO-001             # L1 instrumentation (completed) — 축 ② 재대조 참조
  - SPEC-HARNESS-LEARNING-EVO-002             # L2 analyzer (completed) — harness_run producer sibling 패턴 정본
  - SPEC-LSEL-LOCAL-EVOLUTION-001             # LSEL batch cluster 루프 (completed) — 별개 경로 sibling
exclusions_forward_links:
  - SPEC-HARNESS-EVO-CONFIDENCE-MEASURE-001    # learner.go confidence 실측화 (별도 후속, ID 미확정)
  - SPEC-HARNESS-EVO-WRITE-SURFACE-001         # write-surface 개방 + 헌법 amendment (SPEC-3)
  - SPEC-HARNESS-EVO-REQ-ARTIFACT-001          # 요구사항 아티팩트 스키마 + 레거시 retire (SPEC-4)
refresh_axes:
  - axis1: "hns- prefix rename (SPEC-HNS-PREFIX-RENAME-001 completed) — §B/§C/§D-D5/§E/§F-M4 file:line 앵커 전부 재측정"
  - axis2: "§F Milestones LEARNING-EVO 001/002 분석기 계약 재대조 — findings→proposal 경로 proposalgen reserved-namespace 제3 producer(harness_run:) 채택 (delegation-map sibling)"
  - axis3: "frontmatter priority P1→P2, phase v3.0.0→v3.1.0, updated 2026-08-12, status draft 유지, version 0.1.0→0.2.0"
plan_auditor_verdict: pending-re-audit   # v0.1.0 = PASS-WITH-DEBT 0.87 (iter-1, Tier M 임계 0.80); v0.2.0 REFRESH는 재감사 대기
```

### Plan-phase 요약 (v0.2.0 REFRESH)

- scope **불변**: 4 배선 항목(manifest `learning` 블록 / Runner `findings` 계약 / specialist improvement-findings 방출 단계 / 오케스트레이터 post-run push), 10 REQ / 10 AC / 5 milestone 유지
- status **draft 유지** (본 갱신은 plan-phase REFRESH, run-phase 진입 아님)
- 축 ①: SPEC-HNS-PREFIX-RENAME-001(completed) `harness-*` → `hns-*` 개명에 맞춰 모든 §B 앵커 재실측 — Runner `harness-release-update-run.js:82-94` → `hns-release-update-run.js:89-101`, specialist `harness-*-specialist.md` → `hns-*-specialist.md` (Phase 8 at `:170`), harness.md `:34`/`:106` → `:37`/`:113`. 디렉터리 `.claude/agents/harness/`·`.claude/commands/harness/`와 `moai-harness-learner` skill, harness.md workflow doc은 content-token 보존
- 축 ②: §F M2/M4 findings→proposal 경로를 proposalgen reserved-namespace 제3 producer(`harness_run:`)로 채택 — LEARNING-EVO 002 `delegationmap.BuildCandidates`(`internal/harness/delegationmap/proposal.go:37`) sibling 패턴 계승, `ProposalCandidate.Evidence` map seam 사용. moai-harness-learner skill schema gap + LSEL sibling 구분을 known-interaction으로 기록(본 SPEC 처리 대상 아님)
- 축 ③: frontmatter priority P1→P2(enhancement, regression 아님), phase v3.0.0→v3.1.0(40일 slip; era.go H-5 tie-breaker load-bearing — lifecycle token 금지 준수), updated 2026-08-12, version 0.2.0
- Template-First 3-클래스 유지: Go 코드(live만) / user-owned `hns-*` specialist(live만) / dev-only `hns-*-run.js` exemplar(live만) / template-managed doctrine(mirror+make build)

_다음 단계: v0.2.0 REFRESH에 대한 plan-auditor 독립 재감사 게이트 (Phase 0.5) → PASS 시 Implementation Kickoff Approval → run-phase. 직전 v0.1.0 verdict(PASS-WITH-DEBT 0.87)은 v0.2.0 본문 변경(line 앵커 + §F/§H 재대조)으로 cache invalidate — 재감사 필수._

---

## §E.2 Run-phase Evidence

_<pending run-phase — manager-develop 소관>_

---

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase — manager-develop 소관>_

---

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase — manager-docs 소관>_

---

## §F Phase 0.95 Mode Selection

**Input parameters** (orchestrator, plan→run boundary):
- tier: M · scope: ~5-15 files · domain count: 4 (Go v4manifest/doctor · JS Runner · agent MD specialist · doctrine/rule + template mirror)
- file language mix: Go + JavaScript + Markdown (mixed, coding-heavy)
- concurrency benefit: LOW (coding-heavy — Anthropic coding-task parallelism caveat)
- Agent Teams prereqs: not met (harness level ≠ thorough / team.enabled unverified / env unset)

**Mode evaluation**:
| Mode | Selected | Rationale |
|------|----------|-----------|
| 1 trivial | no | 다중 파일 + 시맨틱 변경 (typo 아님) |
| 2 background | no | Write/Edit 수반 (read-only 아님) |
| 3 agent-team | no | Agent Teams capability-gate 미충족 |
| 4 parallel | no | coding-heavy → §B.2 tie-breaker Mode 5 우선 |
| 5 sub-agent | **selected** | coding-heavy 단일 SPEC 5-milestone 순차 구현 (기본 fallback) |
| 6 workflow | no | 기계적 대량(≥~30 파일) 변환 아님; 시맨틱 신규 코드 |

**Decision: sub-agent** (sequential manager-develop, cycle_type=tdd, M1..M5)

**Justification**: 4개 도메인에 걸치지만 코딩 중심(Go 스키마/게이트 + JS 계약 + 문서)이라 Anthropic coding-task parallelism caveat에 따라 병렬화 이득이 낮다. §B.2 tie-breaker(coding-heavy + multi-domain → Mode 5)로 순차 sub-agent를 선택. Implementation Kickoff Approval은 explicit-gate 분기로 사용자 AskUserQuestion 승인 완료(Path A, score-independent) — Phase 0.95는 그 downstream.

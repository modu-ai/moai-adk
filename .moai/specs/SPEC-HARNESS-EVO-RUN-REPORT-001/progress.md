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

### M1 — manifest `learning` 블록 스키마 (REQ-HRR-001, REQ-HRR-002) [TDD]

**상태**: PASS — GREEN 달성 (2026-08-12, worktree HEAD `8d9612655` 위 M1 커밋 대기)

**구현 범위 (AC-HRR-001, AC-HRR-002, AC-HRR-010 legacy 케이스)**:

- `internal/harness/v4manifest/types.go` — `Manifest.Learning *LearningBlock` 옵션 필드 추가 (포인터 → 부재 시 nil = legacy 하위 호환, REQ-HRR-001/010); `LearningBlock` 구조체 정의 (`Enabled bool` / `Tier string` / `ConfidenceFloor float64` / `MaxFindingsPerRun int`). `Schedule *Schedule`에 이은 두 번째 optional 필드 (D1 plan-auditor MINOR 실측 반영 — Schedule 존재가 Learning 추가를 막지 않음)
- `internal/harness/v4manifest/schema.go` — `LearningTier{Observation,Heuristic,Rule,AutoUpdate}` 상수 + `validLearningTiers` 맵 추가. 어휘는 `harness.Tier.String()` SSOT(`internal/harness/types.go:245-258`)에서 파생 (REQ-HRR-002). 별도 병렬 어휘 정의 금지 (AP-1). SSOT 파생은 test-only import로 기계 강제 (`TestLearningTierVocabularyMatchesHarnessSSOT`).
- `internal/harness/v4manifest/validate.go` — `Validate`에 learning.tier 어휘 검증 분기 추가. nil `Learning` = legacy harness 정상 (REQ-HRR-010). 부분 필드(zero-value)는 schema 수준에서 허용 (EC-1 정책: defaults는 M2 findings→proposal mapping에서 적용). `confidence_floor`/`max_findings_per_run` 범위 검증은 doctor(M3, REQ-HRR-005) 소관 — M1 Validate는 tier 어휘만 강제.
- `internal/harness/v4manifest/learning_test.go` — 6 테스트 (RED→GREEN):
  - `TestManifestLearningBlockParsing` (AC-HRR-001) — learning 블록 JSON → 4필드 파싱
  - `TestManifestLearningBlockLegacyNil` (AC-HRR-001/010) — legacy 8-필드 JSON → `Learning == nil` + Validate 통과
  - `TestLearningTierVocabulary` (AC-HRR-002) — 4 유효 tier PASS + 병렬 어휘(`recommendation`/`approval_required`/`auto`/`RULE`) 거부 + empty(unset) 허용
  - `TestLearningTierVocabularyMatchesHarnessSSOT` (AC-HRR-002 기계 강제) — `harness.Tier.String()` SSOT와 `validLearningTiers` 집합 동일성 단언 (test-only import, parallel vocabulary drift 방지)
  - `TestValidate_LearningAbsentRegression` — nil `Learning` baseline 회귀 없음
  - `TestValidate_LearningBlockValid` — fully-populated learning 블록 happy path

**EC-1 (partial block) 정책 채택**: 부분 필드(`enabled`만, 혹은 `tier`만 등)는 schema 수준에서 유효. zero-value 필드는 defaults가 M2 findings→proposal mapping에서 적용 (예상 기본값: tier=observation, confidence_floor=0.70, max_findings_per_run=합리적 기본). M1 `Validate`는 non-empty `tier`가 SSOT 어휘 밖일 때만 거부. 이 정책은 acceptance.md §D.2 EC-1과 plan.md §F M1에 명시됨.

**@MX tag 변경**: `validLearningTiers` 맵에 `@MX:ANCHOR` + `@MX:REASON`(fan_in ≥ 3 candidate) + `@MX:SPEC`(SPEC-HARNESS-EVO-RUN-REPORT-001 M1 / REQ-HRR-002) 추가.

**TDD 증거 (E8 — verbatim RED 출력, GREEN 이전 캡처)**:
```
# go test -run "TestManifestLearningBlock|TestLearningTierVocabulary|TestValidate_Learning" ./internal/harness/v4manifest/
# github.com/modu-ai/moai-adk/internal/harness/v4manifest [github.com/modu-ai/moai-adk/internal/harness/v4manifest.test]
internal/harness/v4manifest/learning_test.go:20:4: m.Learning undefined (type Manifest has no field or method Learning)
internal/harness/v4manifest/learning_test.go:20:16: undefined: LearningBlock
internal/harness/v4manifest/learning_test.go:22:23: undefined: LearningTierAutoUpdate
internal/harness/v4manifest/learning_test.go:65:7: m.Learning undefined (type Manifest has no field or method Learning)
...
internal/harness/v4manifest/learning_test.go:77:7: too many errors
FAIL	github.com/modu-ai/moai-adk/internal/harness/v4manifest [build failed]
```
GREEN 후 6 테스트 전부 PASS.

**자체 검증 (이 run, 이 tree, HEAD `8d9612655`)**:

| 항목 | 결과 | 명령 | 관측 출력 |
|------|------|------|-----------|
| E1 AC 매트릭스 | AC-HRR-001/002/010 PASS | `go test -run "TestManifestLearningBlock\|TestLearningTierVocabulary\|TestValidate_Learning" -v ./internal/harness/v4manifest/` | 6 PASS (유효 tier 4 + 무효 4 + legacy nil + parsing + SSOT 동일성 + 회귀) |
| E2 cross-platform build | PASS | `go build ./...` + `GOOS=windows GOARCH=amd64 go build ./...` | 둘 다 exit 0 (syscall 무관 — 순수 struct/validate) |
| E3 coverage | 100.0% (≥85%) | `go test -cover ./internal/harness/v4manifest/...` | `coverage: 100.0% of statements` |
| E4 subagent boundary | clean | `grep -rn 'AskUserQuestion' internal/harness/v4manifest/ \| grep -v _test.go` | exit 1 (0 matches) |
| E5 lint | NEW 0 | `golangci-lint run --timeout=2m` | `0 issues` (baseline 동일) |
| E6 HEAD + push | 커밋 대기 (아래 커밋에서 기록) | — | — |
| E7 blocker | 없음 | — | — |
| E8 RED 출력 | verbatim 캡처 (위) | GREEN 이전 컴파일 실패 | `FAIL [build failed]` |
| 추가: v4manifest full suite | PASS | `go test ./internal/harness/v4manifest/...` | `ok ... 0.332s` |
| 추가: cli/harness 회귀 (PIPE-REPAIR doctor) | PASS | `go test ./internal/cli/harness/...` | `ok ... 4.911s` |
| 추가: race (touched + sibling) | PASS | `go test -race ./internal/harness/v4manifest/... ./internal/cli/harness/...` | 둘 다 ok |

**Gaps**: 없음 (M1 범위 전량 검증 완료). `full suite` (`go test ./...`)는 백그라운드 실행 — cascade 회귀는 v4manifest + cli/harness 타겟 suite로 이미 커버됨 (M1은 v4manifest schema에 국한, 타 패키지 import 변경 없음).

**Residual-risk**: (i) EC-1 부분 블록 기본값 적용 시점이 M2로 연기됐으므로, 부분 블록을 파싱은 하지만 downstream 소비자가 아직 없음 — M2에서 defaults 헬퍼 추가 시 재검증 필요; (ii) `harness.Tier.String()` SSOT 어휘가 future SPEC으로 확장될 경우 `validLearningTiers` + 상수 동기화가 `TestLearningTierVocabularyMatchesHarnessSSOT`에 의해 강제되므로 자동 감지됨.

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

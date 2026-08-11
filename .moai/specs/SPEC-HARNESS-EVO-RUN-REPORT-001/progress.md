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

**Residual-risk**: (i) EC-1 부분 블록 기본값 적용 시점이 M2로 연기됐으므로, 부분 블록을 파싱은 하지만 downstream 소비자가 아직 없음 — M2에서 defaults 헬퍼 추가 시 재검증 필요; (ii) `harness.Tier.String()` SSOT 어휘가 future SPEC으로 확장될 경우 `validLearningTiers` + 상수 동언성이 `TestLearningTierVocabularyMatchesHarnessSSOT`에 의해 강제되므로 자동 감지됨.

---

### M2 — Runner return-schema `findings` 계약 + harness_run producer 매핑 (REQ-HRR-003, 004, 007) [TDD]

**상태**: PASS — GREEN 달성 (2026-08-12, worktree HEAD `870cdd72a` 위 M2 커밋 대기)

**구현 범위 — M2-a (Runner findings 표준 계약, REQ-HRR-003/004)**:

- `.claude/workflows/hns-release-update-run.js` (dev-only exemplar, §21 격리) — return 객체에 `findings: []` 표준 필드 추가. 본 Runner는 읽기 전용 research sweep 모델이므로 실행 시점 개선 신호가 없고, 빈 배열이 정직한 신호 (REQ-HRR-003: 부재 ≠ 무신호 — 필드 생략 금지, 빈 배열로 구분). 주석에 REQ-HRR-003/004 + `harness_run:` producer routing 명시. AC-HRR-010(b) legacy Runner 호환 (빈 findings = legacy 취급).
- **v4 Builder GENERATE Runner 템플릿 계약**: `internal/template/templates/`에 Runner 템플릿 **부재** 실측 (`find -path '*harness*run.js' -o -path '*workflows*run.js'` → empty). 따라서 template mirror 불필요 — exemplar 수정만으로 충분 (plan.md §F M2 + §D-D5 소급 범위 확정). forward-note: Builder가 Runner 템플릿을 향후 생성하면 그 시점에 template-managed 표면으로 §25 중립성과 함께 반영.

**구현 범위 — M2-b (`harness_run:` 제3 producer, path A — REQ-HRR-007)**:

- `internal/harness/harnessrun/types.go` (신규 패키지) — `Finding{Surface, Kind, Summary, Confidence, SuggestedTier}` 5-필드 구조체 (REQ-HRR-003 표준 shape); `PatternNamespace = "harness_run"` (reserved namespace, `delegationmap` sibling); kind 어휘 상수 `{KindDrift, KindGap, KindFriction, KindDefect}`; `ConservativeConfidenceFloor = 0.70` (floor-aligned default, REQ-HRR-004 — **learner.go defaultConfidence(1.0) 재사용 금지**, 본 패키지에서 자체 정의).
- `internal/harness/harnessrun/proposal.go` — `BuildHarnessRunCandidates(findings []Finding) []proposalgen.ProposalCandidate`. `delegationmap.BuildCandidates`(`internal/harness/delegationmap/proposal.go:37`) sibling 패턴 계승: (i) reserved namespace `harness_run:` 사용, (ii) `ProposalCandidate` direct construction (`MapPromotions` 경유 금지 — AP-11), (iii) `Evidence map[string]any` seam에 `{surface, kind, summary, confidence, suggested_tier, approval_gate}` 6-field 전달. pattern-key = `harness_run:<sha256(surface)[:8]>:<kind>` (plan.md §D-D2). `ObservationCount=1` (harness-run 단발 관측). `SourceTs` zero time (dynamic-workflow determinism — clock read 금지, caller가 run 후 stamp).
- `internal/harness/harnessrun/proposal_test.go` — 8 테스트 (RED→GREEN):
  - `TestBuildHarnessRunCandidates_MapsFields` (AC-HRR-003/004 매핑) — pattern_key namespace + kind suffix, confidence verbatim (0.75), tier verbatim, ObservationCount=1, DraftID PROPOSAL- prefix, Evidence 6-key
  - `TestBuildHarnessRunCandidates_EmptyInput` (REQ-HRR-003 no-signal) — nil/empty → non-nil empty slice (field-present/no-signal 구분)
  - `TestBuildHarnessRunCandidates_DeterministicIdempotent` (sibling of `TestAnalyze_DeterministicIdempotent`) — 동일 findings 2회 호출 → byte-identical candidates (pattern_key, draft_id, SourceTs 동일)
  - `TestBuildHarnessRunCandidates_DistinctSurfacesDistinctKeys` — surface별 pattern_key 충돌 방지 (sha256 discriminator)
  - `TestBuildHarnessRunCandidates_DistinctKindsDistinctKeys` — 동일 surface + 다른 kind → 별도 key
  - `TestPatternNamespace_IsolatedFromMapperSSOT` (residual-risk i 폐쇄, E9) — `harness_run`이 `PatternBearingEventTypes()` 비멤버임 단언 (`delegationmap/proposal_test.go:54-56` 동등 메커니즘, AP-11 기계 강제)
  - `TestPatternKey_RejectedByExistingMapper` — 모든 emitted `harness_run:` key를 maximally-actionable promotion으로 real `MapPromotions`에 feed → 0 candidates (격리 실증)
  - `TestNoLearnerDefaultConfidenceReference` (REQ-HRR-004 기계 강제) — `proposal.go`/`types.go` AST walk → `defaultConfidence` code identifier 부재 단언 (prose mention은 허용, `go/ast` + `go/parser`로 comment와 code 식별)

**설계 결정 (격리 vs 재사용 균형, plan §I gap (ii) 폐쇄)**: `harnessrun/` sibling 패키지 신설 채택. 이유: (i) namespace 격리가 `harness_run:` reserved prefix + `PatternBearingEventTypes` SSOT 비추가로 기계적으로 강제되므로, `proposalgen/` 확장이나 `delegationmap/` 내부보다 sibling이 격리 경계를 명확히 함; (ii) `proposalgen.ProposalCandidate`/`WriteProposals` 재사용은 유지 (코드 중복 방지) — sibling이 import 경로로 재사용. 이 균형은 §E 보고에 명시됨 (plan §I gap (ii) 확정).

**@MX tag**: 본 패키지는 신규 + 작은 범위라 `@MX:ANCHOR`(fan_in ≥ 3) 대상 아직 아님. `BuildHarnessRunCandidates`가 향후 3+ caller 확보 시 ANCHOR 추가 예정.

**TDD 증거 (E8 — verbatim RED 출력, GREEN 이전 캡처)**:
```
# go test ./internal/harness/harnessrun/...
# github.com/modu-ai/moai-adk/internal/harness/harnessrun [github.com/modu-ai/moai-adk/internal/harness/harnessrun.test]
internal/harness/harnessrun/proposal_test.go:26:16: undefined: Finding
internal/harness/harnessrun/proposal_test.go:29:19: undefined: KindFriction
internal/harness/harnessrun/proposal_test.go:36:16: undefined: BuildHarnessRunCandidates
internal/harness/harnessrun/proposal_test.go:44:38: undefined: PatternNamespace
internal/harness/harnessrun/proposal_test.go:45:81: undefined: PatternNamespace
internal/harness/harnessrun/proposal_test.go:47:42: undefined: KindFriction
internal/harness/harnessrun/proposal_test.go:48:82: undefined: KindFriction
internal/harness/harnessrun/proposal_test.go:94:25: undefined: Finding
internal/harness/harnessrun/proposal_test.go:95:10: undefined: BuildHarnessRunCandidates
internal/harness/harnessrun/proposal_test.go:113:16: undefined: Finding
internal/harness/harnessrun/proposal_test.go:113:16: too many errors
FAIL	github.com/modu-ai/moai-adk/internal/harness/harnessrun [build failed]
```
GREEN 후 8 테스트 전부 PASS (`ok ... 0.498s`).

**자체 검증 (이 run, 이 tree, HEAD `870cdd72a` + M2 변경 미커밋)**:

| 항목 | 결과 | 명령 | 관측 출력 |
|------|------|------|-----------|
| E1 AC 매트릭스 | AC-HRR-003 PASS (findings 5-field + 빈 배열), AC-HRR-004 PASS (confidence 출처 분리, defaultConfidence 미참조 — AST guard), AC-HRR-010(b) PASS (legacy 빈 findings Runner 호환) | `go test -v ./internal/harness/harnessrun/...` | 8 PASS |
| E2 cross-platform build | PASS | `go build ./...` + `GOOS=windows GOARCH=amd64 go build ./...` | 둘 다 exit 0 (syscall 무관 — 순수 변환 헬퍼) |
| E3 coverage | 100.0% (≥85%) | `go test -cover ./internal/harness/harnessrun/...` | `coverage: 100.0% of statements` |
| E4 subagent boundary | clean | `grep -rn 'AskUserQuestion' internal/harness/harnessrun/` | exit 1 (0 matches); Runner는 comment 줄만 (AC-HRR-008) |
| E5 lint | NEW 0 | `golangci-lint run --timeout=3m ./internal/harness/harnessrun/...` | `0 issues.` |
| E6 HEAD + push | 커밋 대기 (아래 커밋에서 기록) | — | — |
| E7 blocker | 없음 | — | — |
| E8 RED 출력 | verbatim 캡처 (위) | GREEN 이전 컴파일 실패 | `FAIL [build failed]` |
| E9 namespace 격리 live guard | PASS | `go test -run 'TestPatternNamespace_IsolatedFromMapperSSOT\|TestPatternKey_RejectedByExistingMapper' -v ./internal/harness/harnessrun/` | 2 PASS (SSOT 비멤버 + real mapper 0 candidates) |
| E10 Template-First | PASS (template 미건드림) | `grep -rn 'HARNESS-EVO-RUN-REPORT\|REQ-HRR' internal/template/templates/` + `find internal/template/templates -path '*hns*run.js' -o -path '*agents/harness*'` | 둘 다 0 matches (exemplar dev-only 유지, template 중립성 보존) |
| 추가: AP-11 SSOT 미확장 | PASS | `grep -n 'harness_run' internal/harness/types.go` | 0 matches (`harness_run` not in PatternBearingEventTypes) |
| 추가: AP-2 confidence 출처 | PASS | `grep -rn 'defaultConfidence' internal/harness/harnessrun/*.go \| grep -v '// '` | 0 matches (code reference 부재; prose만) |
| 추가: sibling package 회귀 | PASS | `go test ./internal/harness/v4manifest/... ./internal/harness/delegationmap/... ./internal/harness/proposalgen/...` | 3 ok (M1 + delegationmap + proposalgen 무영향) |
| 추가: race | PASS | `go test -race ./internal/harness/harnessrun/...` | `ok ... 1.487s` |

**Gaps**: (i) v4 Builder GENERATE Runner 템플릿이 현재 `internal/template/templates/`에 부재하여 template 계약 반영 불가 — 향후 Builder가 Runner 템플릿을 생성하면 그 SPEC에서 `findings` 계약을 template-managed 표면으로 반영 (forward-note); (ii) `full suite` (`go test ./...`) 실행 시 `internal/template`의 `TestTemplateNoInternalContentLeak` + `TestRuleDateProvenance` 2건 FAIL 관측 — 이는 `zone-registry.md` (본 SPEC 미접근)의 pre-existing failure로, clean M1 baseline `870cdd72a` (stash 후 실측)에서 동일 실패 재현됨 → 본 M2 scope 외 (B10 scope discipline).

**Residual-risk**: (i) `BuildHarnessRunCandidates`의 단일 caller가 현재 테스트뿐 — orchestrator post-run push 배선(REQ-HRR-007)은 M4 doctrine 표면 소관으로, 실 caller는 M4에서 연결됨. producer 계약 자체는 M2에서 확정 + 테스트 강제; (ii) 3 producer(tier-ladder/delegation-map/harness-run)가 동일 Tier-4 게이트 rate_limit를 공유할 때 producer간 경합 시맨틱(EC-6)은 M4 doctrine 표면에서 명시 예정.

### M3 — doctor learning 축 (REQ-HRR-005) [TDD]

**상태**: PASS — GREEN 달성 (2026-08-12, worktree HEAD `cd88275fe` 위 M3 커밋 대기)

**Claim**: `checkHarness`에 learning 축 2-검사(confidence_floor 범위 + enabled/findings 선언 정합)를 추가했다. tier 어휘 검사는 M1 `v4manifest.Validate`가 schema 수준에서 이미 소관(axis="manifest" ERROR) — doctor에서 재검하지 않는다(double-reporting 회피, M1 `LearningBlock.Tier` godoc + AP-1 정합). learning 블록 부재는 무보과(ERROR 아님, REQ-HRR-010/AP-6).

**Evidence** (이 run, 이 tree, HEAD `cd88275fe` + M3 변경 미커밋):

| AC | Status | Verification Command | Actual Output |
|----|--------|---------------------|---------------|
| AC-HRR-005 (tier 오타 → ERROR) | PASS | `go test -run TestDoctor_LearningAxis_TierTypo ./internal/cli/harness/` | `--- PASS: TestDoctor_LearningAxis_TierTypo` (schema Validate 경로, axis="manifest") |
| AC-HRR-005 (floor 범위밖 → ERROR) | PASS | `go test -run TestDoctor_LearningAxis_ConfidenceFloorOutOfRange ./internal/cli/harness/` | `--- PASS` (doctor learning 축, axis="learning") |
| AC-HRR-005 (enabled+findings 미선언 → ERROR) | PASS | `go test -run TestDoctor_LearningAxis_EnabledButNoFindingsDeclaration ./internal/cli/harness/` | `--- PASS` (doctor learning 축, axis="learning") |
| AC-HRR-005 (learning 부재 → exit 0) | PASS | `go test -run TestDoctor_LearningAxis_AbsentNotError ./internal/cli/harness/` | `--- PASS` (ErrorCount=0, 무보과) |
| AC-HRR-010(c) (legacy 회귀) | PASS | `go test -run TestDoctor_ValidHarness_Passes ./internal/cli/harness/` | `--- PASS` (기존 4축 무회귀, learning 블록 없는 manifest 정상) |

**RED verbatim (GREEN 이전 캡처)**:
```
$ go test -run TestDoctor_LearningAxis ./internal/cli/harness/
--- FAIL: TestDoctor_LearningAxis_TierTypo (0.01s)
    doctor_learning_test.go:90: expected a learning-axis ERROR for invalid tier; findings=[{Harness:badtier Axis:manifest Severity:ERROR ...}]
--- FAIL: TestDoctor_LearningAxis_EnabledButNoFindingsDeclaration (0.01s)
    doctor_learning_test.go:134: expected a learning-axis ERROR for enabled+no-findings-declaration; findings=[]
    doctor_learning_test.go:137: error_count = 0, want >= 1 (enabled harness must declare findings)
--- FAIL: TestDoctor_LearningAxis_ConfidenceFloorOutOfRange (0.01s)
    doctor_learning_test.go:111: expected a learning-axis ERROR for confidence_floor > 1; findings=[]
    doctor_learning_test.go:114: error_count = 0, want >= 1 (out-of-range floor must fail the gate)
FAIL    github.com/modu-ai/moai-adk/internal/cli/harness    0.432s
```

**자체 검증 (이 run, 이 tree, HEAD `cd88275fe` + M3 변경 미커밋)**:

- (a) command + (b) observed output + (c) baseline-attribution:
  - `go test ./internal/cli/harness/...` → `ok github.com/modu-ai/moai-adk/internal/cli/harness 4.874s` (전량 PASS, 4축 + learning 축 + dormancy + hns 회귀)
  - `go test -race ./internal/cli/harness/...` → `ok ... 6.292s` (race clean)
  - `go build ./...` → exit 0; `GOOS=windows GOARCH=amd64 go build ./...` → exit 0 (cross-platform)
  - `go test -coverprofile=/tmp/cover_m3.out ./internal/cli/harness/...` → touched file doctor.go: `checkLearningAxis 100.0%`, `checkHarness 90.6%`, `Doctor 92.3%`, `NewHarnessDoctorCmd 84.0%` (모든 M3 touched function ≥ 84%; `checkLearningAxis` 100%)
  - `golangci-lint run --timeout=2m ./internal/cli/harness/...` → `0 issues.`
  - subagent boundary: `grep -rn 'AskUserQuestion' internal/cli/harness/ | grep -v _test` → 0 실제 호출 (comment/prose만, `TestDoctor_NoAskUserQuestion` static guard PASS)
- **설계 결정**: `max_findings_per_run` 범위 검사는 M3에서 추가하지 않았다 — M1 `LearningBlock.MaxFindingsPerRun` godoc가 명시적으로 "truncation policy is applied downstream in the M2 findings→proposal mapping (EC-2)"로 위임하므로, doctor에서 중복 검사는 scope 침범. AC-HRR-005가 요구하는 3-검사(tier/floor/findings)는 tier(schema 경로) + floor(doctor) + findings(doctor)로 모두 충족.

**@MX tag 변경**: `runnerFindingsKeyRE` 변수에 `@MX:ANCHOR` + `@MX:REASON`(fan_in ≥ 2: `checkLearningAxis` + `doctor_learning_test.go`; AP-3 JS AST 금지 — regex가 정규 heuristic) 추가.

**Gaps**: (i) `max_findings_per_run` 음수/0 범위 검사가 doctor에 없음 — 의도적 (M2 EC-2 downstream truncation 소관, 위 설계 결정). 향후 M2 mapping 구현 시 음수 처리가 드러나면 그 SPEC에서 재검; (ii) `full suite` (`go test ./...`)는 M2 evidence와 동일하게 `internal/template` 2건 pre-existing FAIL(zone-registry.md, 본 SPEC 미접근) — 본 M3 scope 외 (B10 scope discipline).

**Residual-risk**: (i) `findings` 키 감지가 정규식 heuristic(`\bfindings\s*:\s*\[`)이므로, Runner가 `findings`를 비-배열 원시 값으로 선언하는 변형은 감지 못함 — 단 REQ-HRR-003 계약이 배열을 규정하므로 계약 위반은 별도 검증 단계에서 잡힘; (ii) learning 축이 runner file이 읽힌 후에만 동작하므로, runner file 부재(ERROR) 시 learning 결함이 동시에 존재해도 manifest/runner ERROR가 우선 보고됨 — 이는 의도적 (더 근본적인 결함 우선, noise 회피).

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

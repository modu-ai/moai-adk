# SPEC-HARNESS-EVO-RUN-REPORT-001 — Implementation Plan

Tier: **M** (5 milestones) · development_mode: **tdd** (quality.yaml 실측: `constitution.development_mode: tdd`) · status: draft · version 0.2.0 (2026-08-12 plan-phase REFRESH)

> 본 plan.md는 2026-08-12 REFRESH (축 ① hns- 개명 앵커 재측정 + 축 ② §F LEARNING-EVO 001/002 분석기 계약 재대조)를 반영한다. scope는 불변 (4 배선 항목), status는 draft 유지 (plan-phase 갱신, run-phase 진입 아님).

---

## §A 컨텍스트

### §A.1 Epic Harness-Evolution 개요 (4 SPECs, 순서 1→2→3→4)

정본 로드맵은 `SPEC-HARNESS-EVO-PIPE-REPAIR-001/plan.md §A.1`(완결, origin a661da107). 재현:

| # | SPEC ID | 범위 | 상태 |
|---|---------|------|------|
| 1 | SPEC-HARNESS-EVO-PIPE-REPAIR-001 | 파이프 수리(어휘/스키마/자동화) + v4 스모크 게이트 + 디스패처/문서 드리프트 수리 | **completed** (a661da107) |
| 2 | **SPEC-HARNESS-EVO-RUN-REPORT-001 (본 SPEC)** | 실행→학습 배선: manifest `learning` 블록, Runner return-schema `findings`, specialist 필수 improvement-findings 최종 단계, post-run findings 수집 → proposalgen reserved-namespace producer(`harness_run:`) → 즉시 AskUserQuestion push(현행 pull-only apply 대체). **learner.go confidence 실측화는 §E 제외 — 별도 후속 SPEC** | draft (본 SPEC, v0.2.0 REFRESH) |
| 3 | SPEC-HARNESS-EVO-WRITE-SURFACE-001 | frozen_guard allowlist 단계적 확장, LOOP-CLOSURE C1 헌법 제약 티어별 표면 자율 amendment, harness-namespace-doctrine evolution-write 소유권 | 미작성 |
| 4 | SPEC-HARNESS-EVO-REQ-ARTIFACT-001 | manifest source_request 구조화 요구사항 스키마, 레거시 5-layer marker retire, Builder 개선 | 미작성 |

**사용자 결정 기록 (본 세션, AFK — 권장 옵션 채택)**: 본 SPEC의 범위를 **4개 배선 항목으로 한정**하고 learner.go confidence 실측화는 별도 후속 SPEC으로 분리(연기). 아래 SPEC-3/4의 "수정 가능(revisable)" 결정들(티어별 표면 자율 목표 상태, 단계적 write-surface 개방, Epic+4-split, cleanup 4건)은 PIPE-REPAIR plan.md §A.1에 기록된 대로 SPEC-3/4 소관이며 본 SPEC 범위 밖이다 — 컨텍스트로만 인지.

### §A.2 이 SPEC이 닫는 링크

PIPE-REPAIR가 관측→분류→제안(수동/passive) 파이프를 복구했다. 본 SPEC은 **하네스 실행 그 자체에서 발견된 개선 신호를 학습 서브시스템으로 전달**하는 능동(active) 배선을 놓는다 — 4개 배선 지점(manifest learning 선언 / Runner findings 반환 / specialist findings 방출 / proposalgen reserved-namespace producer → 오케스트레이터 post-run push). spec.md §A.2 참조.

### §A.3 findings→proposal producer 구도 (축 ② 재대조 — 본 SPEC의 위치)

본 SPEC plan-phase REFRESH (2026-08-12) 시점에 도달한 Tier-4 오케스트레이터 승인 게이트에 닿는 **3-producer 구도** (+1 sibling 루프):

| # | Producer | source | pattern-key namespace | Finding/Candidate shape | 소유 SPEC |
|---|----------|--------|----------------------|-------------------------|----------|
| 1 | **tier-ladder producer** | `.moai/harness/learning-history/tier-promotions.jsonl` (usage-log 관측) | `proposalgen.MapPromotions` 정규식 (PatternBearingEventTypes SSOT) | `Promotion` → `ProposalCandidate` (fixed 6-field) | SPEC-V3R6-HARNESS-PROPOSAL-GEN-001 (completed) |
| 2 | **delegation-map producer** | `internal/harness/delegationmap/Analyze` (routing-ledger 행 집계) | `delegation_map:` reserved (proposalgen 정규식이 의도적 기각, `proposal.go:21`) | `Finding{Kind,Subcommand,Agent,ObservationCount,SupportRatio,QualifyingRows,UnattributedShare}` → `BuildCandidates` → `ProposalCandidate` (`Evidence` map에 7-field) | SPEC-HARNESS-LEARNING-EVO-002 (completed) |
| 3 | **harness-run findings producer (본 SPEC)** | Runner/specialist 실행 중 발견된 harness artifact friction | `harness_run:` reserved (본 SPEC이 신설, sibling 패턴 계승) | `{surface, kind, summary, confidence, suggested_tier}` → `ProposalCandidate` (`Evidence` map에 4-field) | SPEC-HARNESS-EVO-RUN-REPORT-001 (본 SPEC, draft) |

**핵심 설계 결정 (축 ② 채택 — option A)**: 본 SPEC의 harness-run findings는 **proposalgen reserved-namespace 제3 producer**로 routing된다. 이유: (i) LEARNING-EVO 002 `delegationmap.BuildCandidates`(`internal/harness/delegationmap/proposal.go:37`)가 이미 확립한 sibling 패턴(reserved namespace + `ProposalCandidate` direct construction + `Evidence map[string]any` seam at `internal/harness/proposalgen/types.go:71`)의 계승; (ii) 3 producer가 동일 Tier-4 오케스트레이터 AskUserQuestion 게이트와 `harness.yaml` rate_limit SSOT를 공유하는 구조적 일관성; (iii) 직전 draft의 (B) 직접-push 기울임이 초래하는 3개 병렬 push path(orchestrator가 producer별 특수 코드를 요구) 회피.

**sibling 루프 (LSEL, 별개 경로)**: SPEC-LSEL-LOCAL-EVOLUTION-001 LSEL은 `.moai/lessons-inbox.jsonl` cluster → human-approved `decision.json` → `diff.patch`로 6개 evolvable surface(`.claude/lsel/frozen-allowlist.json` per `.claude/agents/harness/** + hns-* skills` 포함)를 개선하는 **batch cluster** 루프이다. 본 SPEC의 harness-run findings는 **live push**(실행 종료 즉시 Tier-4 게이트)이며 경로가 다르다. LSEL applier(`internal/harness/applier.go:22` `enableTriggerInjectionWrites = false`)는 frozen 유지 — 본 SPEC이 LSEL APPLY path를 변경하지 않는다. 두 루프의 surface 교집합은 run-phase 관찰 대상이나 본 SPEC이 통합하지 않는다.

**moai-harness-learner skill schema gap (known-interaction)**: `.claude/skills/moai-harness-learner/SKILL.md` payload schema는 flat 8-field(`proposal_id/target_path/field_key/current_value/new_value/observation_count/confidence/recommended_action`)이며 tier-ladder producer의 skill-text apply 전용으로 설계되었다(재실측). 본 SPEC의 harness-run findings는 이 schema를 사용하지 않고 `ProposalCandidate` → Tier-4 게이트 경로로 routing되므로, learner skill schema 확장은 **본 SPEC 범위 밖**(known-interaction, forward-link 후속 SPEC). run-phase M4에서 schema gap이 실측 장애로 드러나면 blocker report.

---

## §B Known Issues (검증 앵커 요약 — 2026-08-12 REFRESH 재실측)

| # | 결함 | 앵커 (사용 명령 — 2026-08-12 재실측) | REQ |
|---|------|------------------|-----|
| B1 | manifest 스키마에 learning 선언 슬롯 부재 | `v4manifest/types.go:18-48` Manifest 8필드 (전량 Read 재실측); `release-update/manifest.json` jq keys → 8키 no learning | REQ-HRR-001/002 |
| B2 | Runner return-schema 표준 findings 계약 부재 | `hns-release-update-run.js:89-101` return `{manifest,capability,sweep_target_count,impact_tables,note}`, `module.exports` at `:103` (Read 재실측; 직전 `harness-release-update-run.js:82-94`에서 개명+드리프트 보정) | REQ-HRR-003/004 |
| B3 | specialist improvement-findings 방출 단계 부재 | `hns-release-update-specialist.md:170` Phase 8=Completion (요약만); 3개 hns-* specialist (ls 재실측: `hns-{release-update,github,release}-specialist.md`; 디렉터리 `.claude/agents/harness/` 보존) | REQ-HRR-006 |
| B4 | post-run apply pull-only (push 부재) | `harness.md:37` "every verb (`status`, `apply`, `rollback`, `disable`)"; `:113` apply 행 = "Surface next Tier-4 proposal via AskUserQuestion" (grep 재실측; 직전 `:34`/`:106`에서 `:37`/`:113`로 드리프트. harness.md 자체는 content-token 보존) | REQ-HRR-007/008 |
| B5 | doctor learning 축 부재 | `internal/cli/harness/doctor.go` 4축(command/manifest/Runner/agent)+ERROR/INFO, learning 축 없음 (Read 재실측, PIPE-REPAIR M3 소관) | REQ-HRR-005 |
| B6 | moai-harness-learner skill schema gap (축 ② known-interaction, 신규) | `.claude/skills/moai-harness-learner/SKILL.md` flat 8-field payload schema — tier-ladder skill-text apply 전용; delegation-map/harness-run findings 미소화 (Read 재실측) | (본 SPEC 처리 대상 아님 — forward-link 후속 SPEC. REQ-HRR-007 경로 설계로 회피) |

**DEFERRED(§E) 실측**: `learner.go:96` `const defaultConfidence = 1.0` (재실측); tier-promotions.jsonl 16 레코드 전량 `confidence: 1` (`grep -o '"confidence":[0-9.]*' | sort | uniq -c`). 본 SPEC은 이를 수리하지 않음.

---

## §C Pre-flight

착수 전 manager-develop 실행 (agent-common-protocol § Pre-Spawn Sync Check + PIPE-REPAIR 전제 재확인 + 축 ② sibling 계약 재확인):

1. `git fetch origin main` + `git rev-list --count --left-right origin/main...HEAD` divergence 확인
2. `go build ./...` + `GOOS=windows GOARCH=amd64 go build ./...` 크로스 플랫폼 그린 베이스라인
3. `go test ./internal/harness/... ./internal/cli/harness/...` 그린 베이스라인 (PIPE-REPAIR doctor 테스트 + LEARNING-EVO 001/002 delegationmap 테스트 포함 통과 확인)
4. **PIPE-REPAIR 전제 재고정**: `types.go`에서 `Tier.String()` 최종 어휘 재-Read (`{observation, heuristic, rule, auto_update}` 확정 확인 — REQ-HRR-002 파생 근거). `proposalgen/mapper.go` `actionableTiers`가 `{rule, auto_update}`로 정렬되었는지 확인 (본 SPEC의 `learning.tier` 어휘가 여기서 파생)
5. **축 ② sibling 계약 재확인 (신규)**: `internal/harness/delegationmap/proposal.go:37` `BuildCandidates` → `internal/harness/proposalgen/types.go:33` `ProposalCandidate` 변환 계약 재-Read; reserved namespace 패턴(`delegation_map:` at `proposal.go:21`)과 `Evidence map[string]any` seam(`types.go:71`) 확인. 본 SPEC `harness_run:` namespace가 이 패턴의 sibling임을 run-phase M2에서 재확인
6. content-token 앵커 재고정: `Manifest`(struct at `types.go:18`), `MANIFEST_PATH`/return object(Runner `hns-release-update-run.js:89-103`), `Phase 8`/`Completion`(specialist `hns-release-update-specialist.md:170`), `apply`/`Surface next`(harness.md `:37`/`:113`), `checkHarness`/`SeverityInfo`(doctor.go)
7. baseline 재측정: `wc -l .moai/harness/learning-history/tier-promotions.jsonl`, confidence 분포 grep (progress.md §E.2 기록)

---

## §D Constraints + 설계 결정 (방향 권고 — run-phase에서 검증 후 확정)

### D1 — manifest `learning` 블록 shape (최소성 원칙, Enforce Simplicity)

`v4manifest.Manifest`에 옵션 필드 `Learning *LearningBlock`(포인터 → 부재 시 nil = 하위 호환). 최소 4-필드 shape:

```jsonc
"learning": {
  "enabled": true,               // bool  — 이 하네스가 개선 신호를 방출하는가
  "tier": "auto_update",         // enum  — findings 승격 tier (Tier.String() SSOT: rule|auto_update; observation|heuristic = pre-actionable 유효값)
  "confidence_floor": 0.70,      // number — 제안 후보 자격 최소 confidence (confidenceThreshold=0.70 정합)
  "max_findings_per_run": 5      // int   — 실행당 findings 상한 (노이즈 방지)
}
```

- **어휘 파생 방향**: `learning.tier` 유효값 = `Tier.String()` 어휘에서 파생 (REQ-HRR-002). 별도 어휘 정의 금지 (PIPE-REPAIR B1 재도입 방지)
- **하위 호환**: nil `Learning` = legacy 하네스, 정상 (REQ-HRR-001/010). 기존 8-필드 manifest 거부 금지
- **필드 최소성 재검증 의무**: run-phase M1에서 proposalgen 계약과 대조하여 필드 부족/과잉 확정 (§F 미검증)

### D2 — Runner return-schema `findings` 계약 (표준 shape) + harness_run producer 매핑

Runner 반환 객체에 `findings` 배열 추가. 각 원소 최소 shape:

```jsonc
"findings": [
  {
    "surface": ".claude/commands/harness/<name>.md",  // string — 개선 적용 대상
    "kind": "friction",                                // enum   — drift|gap|friction|defect
    "summary": "release notes 수집이 매 실행 수동 재입력 요구",  // string — 한 줄 서술
    "confidence": 0.75,                                // number — 실행 시점 측정치 (learner.go 하드코딩 재사용 금지, REQ-HRR-004)
    "suggested_tier": "auto_update"                    // enum   — rule|auto_update
  }
]
// 신호 없음 = "findings": []  (필드 생략 금지 — 부재/무신호 구분, REQ-HRR-003)
```

- **confidence 출처 분리(REQ-HRR-004)**: `findings[].confidence`는 실행 시점 값 — `learner.go` `defaultConfidence`(1.0) 재사용 금지. 근거 없으면 보수 기본값 0.70(=confidence_floor 경계) + 추정 표기. §E 제외된 learner 실측화와 **별개** (verification-claim-integrity §1: 미측정을 측정으로 위장 금지)
- **소급 범위(§F 미검증)**: exemplar Runner(release-update, 1개 — `hns-release-update-run.js`로 개명) 실제 수정 = dev-only 로컬 전용(§21). v4 Builder GENERATE가 생성하는 Runner **템플릿** 계약 = template-managed. run-phase에서 Builder 계약과 대조 확정
- **축 ② — harness_run reserved-namespace producer 매핑 (신규)**: 본 SPEC의 findings shape `{surface, kind, summary, confidence, suggested_tier}`는 `ProposalCandidate`(`internal/harness/proposalgen/types.go:33`)로 다음과 같이 매핑된다 — `confidence` → `ProposalCandidate.Confidence`, `suggested_tier` → `ProposalCandidate.Tier`, `surface`/`kind`/`summary` → `ProposalCandidate.Evidence` map. pattern-key = `harness_run:<surface-sha256[:8]>:<kind>` (reserved namespace, `delegationmap.BuildCandidates`의 `delegation_map:` sibling). `ObservationCount`는 harness-run 단발 관측이므로 1 (run-phase에서 반복 관측 누적 여부 재확정). `BuildHarnessRunCandidates` 헬퍼(또는 동등한 변환)를 `delegationmap.BuildCandidates` 패턴에서 분리하여 본 SPEC 소관 패키지에 신설하거나, proposalgen 직접 호출 — run-phase M2에서 확정.

### D3 — doctor learning 축 (정규식 heuristic, JS AST 금지)

`doctor.go` `checkHarness`에 learning 축 추가:
- `learning.tier` ∈ `Tier.String()` 어휘 → 아니면 ERROR
- `learning.confidence_floor` ∈ `[0,1]` → 아니면 ERROR
- `learning.enabled: true`인데 Runner가 `findings` 반환 미선언(정규식 grep `findings` 키) → ERROR
- `learning` 블록 부재 → INFO note만 (ERROR 아님, REQ-HRR-005/010)
- **AP: JS AST 파서 도입 금지** (PIPE-REPAIR AP-2 계승 — 정규식 heuristic 충분; 게이트는 계약 유효성 확인이 목적)

### D4 — post-run push 배선 지점 (doctrine/rule 표현계 + harness_run producer 백엔드 — 축 ② 설계 재대조)

REQ-HRR-007/008을 doctrine 표면으로 배선. **축 ② REFRESH로 방향 확정**: findings→proposal 변환은 proposalgen reserved-namespace 제3 producer(`harness_run:`)가 담당(D2), doctrine/rule 표면은 이 producer의 출력을 오케스트레이터 Tier-4 게이트로 수렴하는 **표현계**(render surface)로 기능한다. doctrine 표면 후보 (run-phase 확정):
- (a) `.claude/skills/moai/workflows/harness.md`(`:113` apply 행 근처)에 post-run push 절 추가 — 실행 종료 시 Runner/specialist findings 수집 → `harness_run:` producer 구동 → `ProposalCandidate` → 오케스트레이터 AskUserQuestion (기존 pull `apply` verb와 공존; pull은 유지, push를 추가하여 pull-only → push-first) **(권고 방향)**
- (b) 별도 rule 파일(`.claude/rules/moai/workflow/harness-post-run-push.md`) — SSOT 신설
- (c) output-style 배너 표면 — 실행 종료 배너에 findings 요약
- **AskUserQuestion 경계(REQ-HRR-008)**: specialist/Runner는 AskUserQuestion 호출 금지 — findings 방출 → 오케스트레이터 수집 → 오케스트레이터 AskUserQuestion. rate-limit(`harness.yaml` `rate_limit` SSOT: max_per_week/cooldown_hours, harness.md `:94`에 명시된 값) 준수
- **빈 findings**: AskUserQuestion 미발화 (결정 피로 방지, REQ-HRR-007)
- **tier-ladder/delegation-map producer와의 공존**: 본 SPEC post-run push는 `harness_run:` namespace proposal에 한정. tier-ladder promotion proposal과 delegation-map amendment proposal은 각자의 기존 경로(`moai harness apply` / `delegationmap.BuildCandidates`)를 유지. 3 producer가 동일 Tier-4 게이트에 도달하되 pattern-key namespace로 출처가 구분됨.

### D5 — Template-First + §25 중립성 (표면별 3-클래스 — 축 ① 갱신)

| 표면 | 클래스 | 처리 |
|------|--------|------|
| `v4manifest/types.go`, `doctor.go` (+테스트) | Go 코드 (template 무관) | live만 수정 |
| `.claude/agents/harness/hns-*-specialist.md` | **user-owned** (§24; 디렉터리 `.claude/agents/harness/` 보존, 파일명만 `harness-*` → `hns-*` 개명) | live만 수정 — template 반입 금지(§21/§24). Runner 보유 여부는 §F 재검증 |
| `.claude/workflows/hns-*-run.js` (exemplar) | **dev-only** (§21; `harness-*-run.js` → `hns-*-run.js` 개명) | live만 수정 — template 반입 절대 금지 |
| v4 Builder GENERATE Runner **템플릿** 계약 (있으면) | template-managed | live + `internal/template/templates/` 동시 + `make build` |
| harness.md / rule / output-style (post-run push doctrine) | template-managed (mirror 실측 필요) | live + template mirror + `make build`; **§25 중립성**: SPEC ID/REQ-HRR/감사 인용 금지, generic prose |

- **§25 중립성 self-check**: `grep -rn "HARNESS-EVO-RUN-REPORT\|REQ-HRR" internal/template/templates/` → 0 matches
- **격리 self-check**: `find internal/template/templates -path '*harness*run.js' -o -path '*agents/harness*'` → empty (dev-only/user-owned 미유출; 파일명이 `hns-*`로 개명됐어도 template 유출 금지는 동일)

### D6 — TDD + 하위 호환

- development_mode=tdd — M1(manifest 스키마)/M2(Runner 계약)/M3(doctor 축)은 RED(fixture 재현) → GREEN → REFACTOR
- 하위 호환 필수: nil `Learning`, findings 없는 Runner, improvement-findings 단계 없는 legacy specialist 모두 정상 동작 (REQ-HRR-010) — RED 단계에서 legacy fixture 통과 테스트 먼저

---

## §E Self-Verification (run-phase 검증 배치 명세 — 축 ① 경로 정정 반영)

run-phase 완료 시 단일 턴 병렬 배치 실행:

1. `go test ./...` (full suite)
2. `go test -coverprofile=cover.out ./internal/harness/v4manifest/... ./internal/cli/harness/...` — touched pkg ≥85%
3. `go test -race ./internal/harness/... ./internal/cli/...`
4. `GOOS=windows GOARCH=amd64 go build ./...` (크로스 플랫폼)
5. `golangci-lint run --timeout=2m`
6. `go test -run TestSplitHarnessNamespaceNoLeak ./internal/template/` (dev-only 격리)
7. 중립성 grep: `grep -rn "HARNESS-EVO-RUN-REPORT\|REQ-HRR" internal/template/templates/` → 0 matches
8. 격리 grep: `find internal/template/templates -path '*hns*run.js' -o -path '*agents/harness*'` → empty (축 ①: 파일명이 `hns-*`로 개명됐어도 template 유출 금지는 동일; legacy `*harness*run.js` 패턴도 보조 확인)
9. `moai harness doctor` 스모크 — learning 축 포함, exemplar(learning 블록/findings 배선 후) 0 ERROR-severity findings
10. subagent boundary: `grep -rn 'AskUserQuestion' .claude/agents/harness/` (디렉터리 보존, `hns-*` specialist 포함) `| grep -v '^[^:]*:[0-9]*:[ \t]*<!--'` → specialist가 AskUserQuestion 직접 호출 없음 (blocker report 패턴만)
11. **축 ② sibling 계약 검증 (신규)**: `grep -n "harness_run:" internal/harness/` → reserved namespace producer 구현 위치 확인; `ProposalCandidate.Evidence` map에 `surface`/`kind`/`summary` keys 전달 확인

---

## §F Milestones

### M1 — manifest `learning` 블록 스키마 (REQ-HRR-001, 002) [TDD]

- RED: `learning` 블록 있는 manifest fixture → 현행 `v4manifest` 파싱이 필드 무시/거부 재현; legacy 8-필드 fixture 통과 테스트(하위 호환)
- GREEN: `Manifest`에 `Learning *LearningBlock` 옵션 필드; `LearningBlock{Enabled, Tier, ConfidenceFloor, MaxFindingsPerRun}`; tier 유효값 = `Tier.String()` 파생
- REFACTOR: nil-safe 접근; 스키마 doc 주석
- 산출: `internal/harness/v4manifest/types.go`(+`validate.go` 확장), `*_test.go`

### M2 — Runner return-schema `findings` 계약 + harness_run producer 매핑 (REQ-HRR-003, 004) [TDD 계약 + 문서 + 변환 헬퍼]

**축 ② 재대조 반영**: 본 milestone은 단순 Runner return-shape 추가에 그치지 않고, findings → `ProposalCandidate` 변환을 담당하는 **harness_run reserved-namespace producer**의 계약을 확정한다.

- RED: findings shape 계약 테스트 — (a) exemplar Runner findings 반환 스모크, (b) `harness_run:` reserved namespace → `ProposalCandidate` 매핑 단위 테스트(`confidence`/`tier`/`evidence` 필드 매핑 단언)
- GREEN:
  - Runner return object에 `findings: []` 표준 필드 추가; confidence 출처 분리(learner.go 상수 재사용 금지) 명문화
  - `BuildHarnessRunCandidates` 헬퍼(또는 동등 변환) 신설 — `delegationmap.BuildCandidates`(`internal/harness/delegationmap/proposal.go:37`)의 sibling 패턴 계승, reserved namespace `harness_run:` 사용, `ProposalCandidate.Evidence` map에 `{surface, kind, summary}` 전달
  - `proposalgen` 정규식이 `harness_run:` key를 기각하는지 확인(`delegation_map:`과 동일 격리 — `PatternBearingEventTypes` SSOT 비추가로 정규식이 reject)
- v4 Builder GENERATE Runner 템플릿 계약(있으면) 반영 — §F 재검증 후 template-managed 여부 확정
- **rate-limit 정합**: `ProposalCandidate.Tier`가 actionable(`{rule, auto_update}`)일 때 `harness.yaml` `rate_limit` SSOT(max_per_week/cooldown_hours)가 Tier-4 게이트에서 준수됨을 확인(REQ-HRR-008)
- 산출: (exemplar 수정=dev-only) `hns-release-update-run.js`; (producer 헬퍼=live Go) `internal/harness/` 내 본 SPEC 소관 패키지(위치 run-phase 확정 — `harnessrun/` 신설 또는 `proposalgen/` 확장); (Builder 계약=template) 해당 시 template + `make build`

### M3 — doctor learning 축 (REQ-HRR-005) [TDD]

- RED: learning 블록 결함 fixture(tier 오타 / confidence_floor 범위 밖 / enabled:true인데 Runner findings 미선언) → doctor ERROR 반환 테스트; learning 블록 없는 하네스 → INFO만(ERROR 아님) 테스트
- GREEN: `checkHarness`에 learning 축 3-검사 추가 (정규식 heuristic, JS AST 금지)
- 산출: `internal/cli/harness/doctor.go`(+`doctor_test.go`)

### M4 — specialist improvement-findings 단계 + post-run push doctrine (REQ-HRR-006, 007, 008) [문서 + doctrine 표현계]

**축 ② 재대조 반영**: post-run push doctrine은 `harness_run:` producer(백엔드)의 출력을 오케스트레이터 Tier-4 게이트로 수렴하는 표현계로 기능한다.

- specialist(release-update — Runner 보유; github/release는 §F 재검증 후 결정): `hns-release-update-specialist.md` Phase 8(`:170`) 이후/직전 필수 improvement-findings 방출 단계 추가; findings shape = REQ-HRR-003 정합; AskUserQuestion 직접 호출 금지(blocker/structured output만)
- post-run push doctrine(§D-D4 방향 확정): `harness.md:113` apply 행 근처에 post-run push 절 추가(권고) — 실행 종료 findings 수집 → `harness_run:` producer 구동 → `ProposalCandidate` → 오케스트레이터 AskUserQuestion; pull `apply` verb 공존; rate-limit SSOT 준수(`harness.yaml` + harness.md `:94`); 빈 findings 미발화; tier-ladder/delegation-map producer와의 namespace 공존 명시
- template-managed 표면은 live + mirror + `make build`; §25 중립성 self-check
- **moai-harness-learner skill schema gap (known-interaction)**: learner skill의 8-field flat schema는 본 SPEC이 확장하지 않는다(§A.3). run-phase M4에서 schema gap이 실측 장애로 드러나면 blocker report → 별도 후속 SPEC.
- 산출: `.claude/agents/harness/hns-*-specialist.md`(user-owned live), post-run push doctrine 표면(`.claude/skills/moai/workflows/harness.md` template mirror)

### M5 — 통합 검증 배치 + 마감

- §E 검증 배치 전체 실행(단일 턴 병렬)
- progress.md §E.2/§E.3 증거 기록 (manager-develop 소관)
- 커밋 분할: M1-M3 Go(테스트 포함) / M2 producer 헬퍼 / M4 specialist(dev)+doctrine(template) — pathspec 제한 커밋

---

## §G Anti-Patterns

- **AP-1**: `learning.tier`에 별도 어휘(`recommendation`/`approval_required`) 정의 — PIPE-REPAIR B1이 제거한 어휘 불일치 재도입 (REQ-HRR-002 위반)
- **AP-2**: `findings[].confidence`에 `learner.go` `defaultConfidence`(1.0) 재사용 — 미측정을 측정으로 위장 (REQ-HRR-004 + verification-claim-integrity §1 위반)
- **AP-3**: doctor learning 축에 JS AST 파서 도입 — 정규식 heuristic 충분 (Enforce Simplicity, PIPE-REPAIR AP-2 계승)
- **AP-4**: specialist가 improvement-findings 단계에서 AskUserQuestion 직접 호출 — subagent boundary 위반 (REQ-HRR-008)
- **AP-5**: exemplar Runner(`hns-*-run.js`) findings 수정을 template 미러링 — dev-only 격리 위반 (§21, CLAUDE.local.md)
- **AP-6**: `learning` 블록 부재를 doctor ERROR로 계상 — learning은 옵션, 하위 호환 파괴 (REQ-HRR-005/010 위반)
- **AP-7**: learner.go confidence 하드코딩을 본 SPEC에서 실측화 — §E 명시 제외, SPEC 범위 침범
- **AP-8**: post-run push를 pull-only apply 제거로 오해 — pull `apply` verb는 유지, push를 **추가**하여 pull-only → push-first (write-surface 정책은 SPEC-3 소관, 본 SPEC은 표면화 시점만 변경)
- **AP-9**: "게이트/배선 통과" 주장을 실행 출력 없이 보고 — verification-claim-integrity §1 위반
- **AP-10 (축 ② 신규)**: harness-run findings를 `moai-harness-learner` skill의 8-field flat schema로 욱여넣기 — schema mismatch, tier-ladder apply 전용 schema를 harness-run findings에 강제 적용. 본 SPEC은 `ProposalCandidate` → Tier-4 게이트 경로를 사용 (§A.3 known-interaction)
- **AP-11 (축 ② 신규)**: `harness_run:` reserved namespace를 proposalgen 정규식 SSOT(`PatternBearingEventTypes`)에 추가 — `delegation_map:`과 동일하게, 정규식이 의도적 기각하도록 두어야 (tier-ladder producer의 관측 taxonomy를 묵음 expand 금지)
- **AP-12 (축 ② 신규)**: 본 SPEC post-run push를 LSEL APPLY path와 통합 시도 — LSEL은 batch cluster + human-approved decision.json 루프(별개 경로), 본 SPEC은 live push(Tier-4 게이트). 경로 혼합 금지

---

## §H Cross-References

### Epic 내 sibling / 선행 SPEC

- `.moai/specs/SPEC-HARNESS-EVO-PIPE-REPAIR-001/` — Epic 1/4 (완결, a661da107); 어휘/스키마 SSOT 정렬 + doctor 스모크 게이트 도입 (본 SPEC의 전제)
- `.moai/specs/SPEC-HARNESS-EVO-WRITE-SURFACE-001` (미작성) — write-surface 개방 + 헌법 amendment (§E forward-link)
- `.moai/specs/SPEC-HARNESS-EVO-REQ-ARTIFACT-001` (미작성) — 요구사항 아티팩트 스키마 + 레거시 retire (§E forward-link)

### 축 ② 재대조로 추가된 sibling SPEC (LEARNING-EVO 001/002 + LSEL)

- `.moai/specs/SPEC-HARNESS-LEARNING-EVO-001/` — L1 instrumentation (completed); 하네스 학습 계측 기반. 본 SPEC의 harness-run findings는 L1 관측망과 독립적(실행 시점 Runner/specialist 방출)이나, 학습 서브시스템 sibling으로 컨텍스트 인지.
- `.moai/specs/SPEC-HARNESS-LEARNING-EVO-002/` — L2 analyzer (completed); `internal/harness/delegationmap/` producer 계약의 정본. **본 SPEC의 `harness_run:` reserved-namespace producer는 이 SPEC이 확립한 sibling 패턴(`BuildCandidates` + reserved namespace + `ProposalCandidate.Evidence` seam)을 계승한다** — 본 SPEC findings→proposal 변환 계약(§D-D2, §F M2)의 설계 근거. 단, 본 SPEC findings는 routing aggregation(delegation-map)이 아닌 **harness-run-time artifact friction** 임을 명시 — 두 producer는 동일 Tier-4 게이트에 도달하되 pattern-key namespace(`delegation_map:` vs `harness_run:`)와 evidence shape로 출처가 구분됨.
- `.moai/specs/SPEC-LSEL-LOCAL-EVOLUTION-001/` — LSEL batch cluster 루프 (completed); `.moai/lessons-inbox.jsonl` → human-approved `decision.json` → `diff.patch`로 6개 evolvable surface 개선. **본 SPEC의 live push 경로와 다른 sibling 루프** — 통합하지 않는다(§A.3 sibling 구분 노트, §G AP-12). LSEL applier(`internal/harness/applier.go:22`)는 frozen 유지.

### proposalgen / harness v4 / loop-closure 계약

- `.moai/specs/SPEC-V3R6-HARNESS-PROPOSAL-GEN-001/` — `proposalgen.MapPromotions` + `ProposalCandidate` + Tier-4 게이트 계약 (completed). **본 SPEC의 `harness_run:` producer는 `MapPromotions`를 경유하지 않고 `ProposalCandidate`를 직접 construct** 한다(`delegationmap.BuildCandidates` sibling 패턴). `Evidence map[string]any` seam이 producer-specific 필드 전달.
- `.moai/specs/SPEC-V3R6-HARNESS-V4-001/` — v4 Builder 4-phase + manifest Runner (manifest 스키마 원 소유; `learning` 필드는 본 SPEC이 옵션 확장)
- `.moai/specs/SPEC-HARNESS-LOOP-CLOSURE-001/` — C1 헌법 제약(auto_apply:false); write-surface 정책 개정은 SPEC-3 소관 (본 SPEC 무관)

### 코드 SSOT

- `internal/harness/v4manifest/types.go:18-48` — Manifest 스키마 SSOT (8-필드, `learning` 부재 — 본 SPEC이 옵션 확장)
- `internal/harness/proposalgen/types.go:33-72` — `ProposalCandidate` (fixed 6-field + `Evidence map[string]any` seam at `:71`); 본 SPEC `harness_run:` producer의 변환 target
- `internal/harness/delegationmap/proposal.go:37` — `BuildCandidates` (sibling 패턴 정본, reserved namespace `delegation_map:` at `:21`)
- `internal/cli/harness/doctor.go` — 스모크 게이트 (PIPE-REPAIR M3)
- `internal/harness/learner.go:96` — `defaultConfidence` (§E 제외, 별도 후속 SPEC)
- `internal/harness/applier.go:22` — LSEL frozen applier `enableTriggerInjectionWrites = false` (LSEL sibling 루프, 본 SPEC 미변경)

### doctrine / 규칙

- `.claude/skills/moai/workflows/harness.md` — harness lifecycle verb surface (apply 행 at `:113`, rate_limit at `:94`); post-run push doctrine 배선 대상 표현계. content-token 보존(hns- 개명 대상 아님)
- `.claude/skills/moai-harness-learner/SKILL.md` — tier-ladder producer apply payload (8-field flat schema, 본 SPEC known-interaction — schema 확장 범위 밖)
- `.claude/rules/moai/core/askuser-protocol.md` — AskUserQuestion 오케스트레이터-단독 경계 (REQ-HRR-008)
- `.claude/rules/moai/core/verification-claim-integrity.md` §1 — 미측정 confidence 위장 금지 (REQ-HRR-004); §2 baseline 귀속 (본 REFRESH 앵커 재실측 근거)
- CLAUDE.local.md §2 Template-First / §21 dev-only isolation / §24 harness namespace / §25 neutrality

---

## §I REFRESH 근거 (2026-08-12, verification-claim-integrity §3)

- **Claim (축 ①)**: 본 plan-phase REFRESH는 SPEC-HNS-PREFIX-RENAME-001(completed) `harness-*` → `hns-*` 개명에 맞춰 §B/§C/§D-D5/§E/§F-M4 의 file:line 앵커를 전부 재측정했다.
- **Evidence**: `ls .claude/agents/harness/` → `hns-{release-update,github,release}-specialist.md` (개명 확인); `ls .claude/workflows/` → `hns-release-update-run.js` (개명 확인); `grep -n 'return {' .claude/workflows/hns-release-update-run.js` → `:89` (직전 `:82`에서 drift); `grep -n 'Phase 8' hns-release-update-specialist.md` → `:170`; `grep -n 'apply' .claude/skills/moai/workflows/harness.md` → verb table `:113` (직전 `:106`에서 drift), 본문 `:37`; `grep -n 'Manifest struct' internal/harness/v4manifest/types.go` → `:18` (8-필드 유지, `learning` 부재 재확인).
- **Claim (축 ②)**: 본 REFRESH는 §F M2/M4의 findings→proposal 경로를 proposalgen reserved-namespace 제3 producer(`harness_run:`)로 채택했다.
- **Evidence**: `internal/harness/delegationmap/proposal.go:37` `BuildCandidates` + `:21` `patternNamespace = "delegation_map"` (sibling 패턴 실측); `internal/harness/proposalgen/types.go:33` `ProposalCandidate` + `:71` `Evidence map[string]any` (확장 seam 실측); `.claude/skills/moai-harness-learner/SKILL.md` flat 8-field payload schema (tier-ladder apply 전용, schema gap 실측); `.claude/lsel/frozen-allowlist.json` evolvable_surfaces (`.claude/agents/harness/** + hns-* skills` 포함, LSEL sibling 루프 실측); `internal/harness/applier.go:22` `enableTriggerInjectionWrites = false` (frozen 실측).
- **Baseline-attribution**: 본 재실측은 2026-08-12 본 worktree HEAD `ed70e4354` (origin/main과 divergence 0 0) 기준이다.
- **Gaps**: (i) template mirror 필요 표면 판정 보류 — run-phase M4에서 post-run push doctrine 표면이 template-managed인지(`.claude/skills/moai/workflows/harness.md` mirror) 최종 확정; (ii) `harness_run:` producer 헬퍼의 Go 패키지 위치(`harnessrun/` 신설 vs `proposalgen/` 확장)는 run-phase M2에서 확정; (iii) LEARNING-EVO 001/002 spec/acceptance 본문은 본 REFRESH에서 요약만 참조(직접 Read는 sibling 계약 코드 영역에 한함) — plan-auditor 독립 감사 시 해당 SPEC 본문 교차검증 권고.
- **Residual-risk**: (i) `harness_run:` reserved namespace가 proposalgen 정규식 SSOT와 충돌 없이 기각되는지는 run-phase M2에서 실측 확인 필요(sibling 패턴 가정); (ii) 3 producer가 동일 Tier-4 게이트 rate_limit를 공유할 때 producer간 rate_limit 경합(tier-ladder가 한도를 소진하면 harness-run이 밀림) 시맨틱은 run-phase M4 doctrine 표면에서 명시 필요.

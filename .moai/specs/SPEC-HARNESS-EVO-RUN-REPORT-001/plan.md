# SPEC-HARNESS-EVO-RUN-REPORT-001 — Implementation Plan

Tier: **M** (5 milestones) · development_mode: **tdd** (quality.yaml 실측: `constitution.development_mode: tdd`) · status: draft

---

## §A 컨텍스트

### §A.1 Epic Harness-Evolution 개요 (4 SPECs, 순서 1→2→3→4)

정본 로드맵은 `SPEC-HARNESS-EVO-PIPE-REPAIR-001/plan.md §A.1`(완결, origin a661da107). 재현:

| # | SPEC ID | 범위 | 상태 |
|---|---------|------|------|
| 1 | SPEC-HARNESS-EVO-PIPE-REPAIR-001 | 파이프 수리(어휘/스키마/자동화) + v4 스모크 게이트 + 디스패처/문서 드리프트 수리 | **completed** (a661da107) |
| 2 | **SPEC-HARNESS-EVO-RUN-REPORT-001 (본 SPEC)** | 실행→학습 배선: manifest `learning` 블록, Runner return-schema `findings`, specialist 필수 improvement-findings 최종 단계, post-run findings 수집 → 즉시 AskUserQuestion push(현행 pull-only apply 대체). **learner.go confidence 실측화는 §E 제외 — 별도 후속 SPEC** | draft (본 SPEC) |
| 3 | SPEC-HARNESS-EVO-WRITE-SURFACE-001 | frozen_guard allowlist 단계적 확장, LOOP-CLOSURE C1 헌법 제약 티어별 표면 자율 amendment, harness-namespace-doctrine evolution-write 소유권 | 미작성 |
| 4 | SPEC-HARNESS-EVO-REQ-ARTIFACT-001 | manifest source_request 구조화 요구사항 스키마, 레거시 5-layer marker retire, Builder 개선 | 미작성 |

**사용자 결정 기록 (본 세션, AFK — 권장 옵션 채택)**: 본 SPEC의 범위를 **4개 배선 항목으로 한정**하고 learner.go confidence 실측화는 별도 후속 SPEC으로 분리(연기). 아래 SPEC-3/4의 "수정 가능(revisable)" 결정들(티어별 표면 자율 목표 상태, 단계적 write-surface 개방, Epic+4-split, cleanup 4건)은 PIPE-REPAIR plan.md §A.1에 기록된 대로 SPEC-3/4 소관이며 본 SPEC 범위 밖이다 — 컨텍스트로만 인지.

### §A.2 이 SPEC이 닫는 링크

PIPE-REPAIR가 관측→분류→제안(수동/passive) 파이프를 복구했다. 본 SPEC은 **하네스 실행 그 자체에서 발견된 개선 신호를 학습 서브시스템으로 전달**하는 능동(active) 배선을 놓는다 — 4개 배선 지점(manifest learning 선언 / Runner findings 반환 / specialist findings 방출 / 오케스트레이터 post-run push). spec.md §A.2 참조.

---

## §B Known Issues (검증 앵커 요약 — 2026-07-03 실측)

| # | 결함 | 앵커 (사용 명령) | REQ |
|---|------|------------------|-----|
| B1 | manifest 스키마에 learning 선언 슬롯 부재 | `v4manifest/types.go` Manifest 8필드 (전량 Read); `release-update/manifest.json` jq keys → 8키 no learning | REQ-HRR-001/002 |
| B2 | Runner return-schema 표준 findings 계약 부재 | `harness-release-update-run.js:82-94` return `{manifest,capability,sweep_target_count,impact_tables,note}` (Read) | REQ-HRR-003/004 |
| B3 | specialist improvement-findings 방출 단계 부재 | `harness-release-update-specialist.md` Phase 8=Completion (요약만); 3개 harness-* specialist (ls) | REQ-HRR-006 |
| B4 | post-run apply pull-only (push 부재) | `harness.md:34,:106` apply="Surface next Tier-4 proposal" (grep) | REQ-HRR-007/008 |
| B5 | doctor learning 축 부재 | `doctor.go` 4축(command/manifest/Runner/agent)+ERROR/INFO, learning 축 없음 (Read) | REQ-HRR-005 |

**DEFERRED(§E) 실측**: `learner.go:96` `defaultConfidence = 1.0`; tier-promotions.jsonl 16 레코드 전량 `confidence: 1` (`grep -o '"confidence":[0-9.]*' | sort | uniq -c`). 본 SPEC은 이를 수리하지 않음.

---

## §C Pre-flight

착수 전 manager-develop 실행 (agent-common-protocol § Pre-Spawn Sync Check + PIPE-REPAIR 전제 재확인):

1. `git fetch origin main` + `git rev-list --count --left-right origin/main...HEAD` divergence 확인
2. `go build ./...` + `GOOS=windows GOARCH=amd64 go build ./...` 크로스 플랫폼 그린 베이스라인
3. `go test ./internal/harness/... ./internal/cli/harness/...` 그린 베이스라인 (PIPE-REPAIR doctor 테스트 포함 통과 확인)
4. **PIPE-REPAIR 전제 재고정**: `types.go`에서 `Tier.String()` 최종 어휘 재-Read (`{observation, heuristic, rule, auto_update}` 확정 확인 — REQ-HRR-002 파생 근거). `proposalgen/mapper.go` `actionableTiers`가 `{rule, auto_update}`로 정렬되었는지 확인 (본 SPEC의 `learning.tier` 어휘가 여기서 파생)
5. content-token 앵커 재고정: `Manifest`(struct), `MANIFEST_PATH`/return object(Runner), `Phase 8`/`Completion`(specialist), `apply`/`Surface next`(harness.md), `checkHarness`/`SeverityInfo`(doctor.go)
6. baseline 재측정: `wc -l .moai/harness/learning-history/tier-promotions.jsonl`, confidence 분포 grep (progress.md §E.2 기록)

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

### D2 — Runner return-schema `findings` 계약 (표준 shape)

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
- **소급 범위(§F 미검증)**: exemplar Runner(release-update, 1개) 실제 수정 = dev-only 로컬 전용(§21). v4 Builder GENERATE가 생성하는 Runner **템플릿** 계약 = template-managed. run-phase에서 Builder 계약과 대조 확정

### D3 — doctor learning 축 (정규식 heuristic, JS AST 금지)

`doctor.go` `checkHarness`에 learning 축 추가:
- `learning.tier` ∈ `Tier.String()` 어휘 → 아니면 ERROR
- `learning.confidence_floor` ∈ `[0,1]` → 아니면 ERROR
- `learning.enabled: true`인데 Runner가 `findings` 반환 미선언(정규식 grep `findings` 키) → ERROR
- `learning` 블록 부재 → INFO note만 (ERROR 아님, REQ-HRR-005/010)
- **AP: JS AST 파서 도입 금지** (PIPE-REPAIR AP-2 계승 — 정규식 heuristic 충분; 게이트는 계약 유효성 확인이 목적)

### D4 — post-run push 배선 지점 (doctrine/rule 표면 — Go 코드 필수 아님)

REQ-HRR-007/008을 doctrine 표면으로 배선. 방향 후보 (run-phase 확정):
- (a) `.claude/skills/moai/workflows/harness.md`에 post-run push 절 추가 — 실행 종료 시 findings 수집 → 오케스트레이터 AskUserQuestion (기존 pull `apply` verb와 공존; pull은 유지, push를 추가하여 pull-only → push-first) **(권고 방향)**
- (b) 별도 rule 파일(`.claude/rules/moai/workflow/harness-post-run-push.md`) — SSOT 신설
- (c) output-style 배너 표면 — 실행 종료 배너에 findings 요약
- **AskUserQuestion 경계(REQ-HRR-008)**: specialist/Runner는 AskUserQuestion 호출 금지 — findings 방출 → 오케스트레이터 수집 → 오케스트레이터 AskUserQuestion. rate-limit(`harness.yaml` `rate_limit` SSOT: max_per_week/cooldown_hours) 준수
- **빈 findings**: AskUserQuestion 미발화 (결정 피로 방지, REQ-HRR-007)

### D5 — Template-First + §25 중립성 (표면별 3-클래스)

| 표면 | 클래스 | 처리 |
|------|--------|------|
| `v4manifest/types.go`, `doctor.go` (+테스트) | Go 코드 (template 무관) | live만 수정 |
| `.claude/agents/harness/*-specialist.md` | **user-owned** (§24) | live만 수정 — template 반입 금지(§21/§24). Runner 보유 여부는 §F 재검증 |
| `.claude/workflows/harness-*-run.js` (exemplar) | **dev-only** (§21) | live만 수정 — template 반입 절대 금지 |
| v4 Builder GENERATE Runner **템플릿** 계약 (있으면) | template-managed | live + `internal/template/templates/` 동시 + `make build` |
| harness.md / rule / output-style (post-run push doctrine) | template-managed (mirror 실측 필요) | live + template mirror + `make build`; **§25 중립성**: SPEC ID/REQ-HRR/감사 인용 금지, generic prose |

- **§25 중립성 self-check**: `grep -rn "HARNESS-EVO-RUN-REPORT\|REQ-HRR" internal/template/templates/` → 0 matches
- **격리 self-check**: `find internal/template/templates -path '*harness*run.js' -o -path '*agents/harness*'` → empty (dev-only/user-owned 미유출)

### D6 — TDD + 하위 호환

- development_mode=tdd — M1(manifest 스키마)/M2(Runner 계약)/M3(doctor 축)은 RED(fixture 재현) → GREEN → REFACTOR
- 하위 호환 필수: nil `Learning`, findings 없는 Runner, improvement-findings 단계 없는 legacy specialist 모두 정상 동작 (REQ-HRR-010) — RED 단계에서 legacy fixture 통과 테스트 먼저

---

## §E Self-Verification (run-phase 검증 배치 명세)

run-phase 완료 시 단일 턴 병렬 배치 실행:

1. `go test ./...` (full suite)
2. `go test -coverprofile=cover.out ./internal/harness/v4manifest/... ./internal/cli/harness/...` — touched pkg ≥85%
3. `go test -race ./internal/harness/... ./internal/cli/...`
4. `GOOS=windows GOARCH=amd64 go build ./...` (크로스 플랫폼)
5. `golangci-lint run --timeout=2m`
6. `go test -run TestSplitHarnessNamespaceNoLeak ./internal/template/` (dev-only 격리)
7. 중립성 grep: `grep -rn "HARNESS-EVO-RUN-REPORT\|REQ-HRR" internal/template/templates/` → 0 matches
8. 격리 grep: `find internal/template/templates -path '*harness*run.js' -o -path '*agents/harness*'` → empty
9. `moai harness doctor` 스모크 — learning 축 포함, exemplar(learning 블록/findings 배선 후) 0 ERROR-severity findings
10. subagent boundary: `grep -rn 'AskUserQuestion' .claude/agents/harness/ | grep -v '^[^:]*:[0-9]*:[ \t]*<!--'` → specialist가 AskUserQuestion 직접 호출 없음 (blocker report 패턴만)

---

## §F Milestones

### M1 — manifest `learning` 블록 스키마 (REQ-HRR-001, 002) [TDD]

- RED: `learning` 블록 있는 manifest fixture → 현행 `v4manifest` 파싱이 필드 무시/거부 재현; legacy 8-필드 fixture 통과 테스트(하위 호환)
- GREEN: `Manifest`에 `Learning *LearningBlock` 옵션 필드; `LearningBlock{Enabled, Tier, ConfidenceFloor, MaxFindingsPerRun}`; tier 유효값 = `Tier.String()` 파생
- REFACTOR: nil-safe 접근; 스키마 doc 주석
- 산출: `internal/harness/v4manifest/types.go`(+`validate.go` 확장), `*_test.go`

### M2 — Runner return-schema `findings` 계약 (REQ-HRR-003, 004) [TDD 계약 + 문서]

- RED: findings shape 계약 테스트(Go 측 계약 검증기가 있으면) 또는 exemplar Runner findings 반환 스모크
- GREEN: Runner return object에 `findings: []` 표준 필드; confidence 출처 분리(learner.go 상수 재사용 금지) 명문화
- v4 Builder GENERATE Runner 템플릿 계약(있으면) 반영 — §F 재검증 후 template-managed 여부 확정
- 산출: (exemplar 수정=dev-only) `harness-release-update-run.js`; (Builder 계약=template) 해당 시 template + `make build`

### M3 — doctor learning 축 (REQ-HRR-005) [TDD]

- RED: learning 블록 결함 fixture(tier 오타 / confidence_floor 범위 밖 / enabled:true인데 Runner findings 미선언) → doctor ERROR 반환 테스트; learning 블록 없는 하네스 → INFO만(ERROR 아님) 테스트
- GREEN: `checkHarness`에 learning 축 3-검사 추가 (정규식 heuristic, JS AST 금지)
- 산출: `internal/cli/harness/doctor.go`(+`doctor_test.go`)

### M4 — specialist improvement-findings 단계 + post-run push doctrine (REQ-HRR-006, 007, 008) [문서]

- specialist(release-update — Runner 보유; github/release는 §F 재검증 후 결정): Phase 8 이후/직전 필수 improvement-findings 방출 단계 추가; findings shape = REQ-HRR-003 정합; AskUserQuestion 직접 호출 금지(blocker/structured output만)
- post-run push doctrine(§D-D4 방향 확정): harness.md 또는 rule 표면 — 실행 종료 findings 수집 → 오케스트레이터 AskUserQuestion; pull `apply` verb 공존; rate-limit SSOT 준수; 빈 findings 미발화
- template-managed 표면은 live + mirror + `make build`; §25 중립성 self-check
- 산출: `.claude/agents/harness/*-specialist.md`(user-owned live), post-run push doctrine 표면(template mirror)

### M5 — 통합 검증 배치 + 마감

- §E 검증 배치 전체 실행(단일 턴 병렬)
- progress.md §E.2/§E.3 증거 기록 (manager-develop 소관)
- 커밋 분할: M1-M3 Go(테스트 포함) / M4 specialist(dev)+doctrine(template) — pathspec 제한 커밋

---

## §G Anti-Patterns

- **AP-1**: `learning.tier`에 별도 어휘(`recommendation`/`approval_required`) 정의 — PIPE-REPAIR B1이 제거한 어휘 불일치 재도입 (REQ-HRR-002 위반)
- **AP-2**: `findings[].confidence`에 `learner.go` `defaultConfidence`(1.0) 재사용 — 미측정을 측정으로 위장 (REQ-HRR-004 + verification-claim-integrity §1 위반)
- **AP-3**: doctor learning 축에 JS AST 파서 도입 — 정규식 heuristic 충분 (Enforce Simplicity, PIPE-REPAIR AP-2 계승)
- **AP-4**: specialist가 improvement-findings 단계에서 AskUserQuestion 직접 호출 — subagent boundary 위반 (REQ-HRR-008)
- **AP-5**: exemplar Runner findings 수정을 template 미러링 — dev-only 격리 위반 (§21, CLAUDE.local.md)
- **AP-6**: `learning` 블록 부재를 doctor ERROR로 계상 — learning은 옵션, 하위 호환 파괴 (REQ-HRR-005/010 위반)
- **AP-7**: learner.go confidence 하드코딩을 본 SPEC에서 실측화 — §E 명시 제외, SPEC 범위 침범
- **AP-8**: post-run push를 pull-only apply 제거로 오해 — pull `apply` verb는 유지, push를 **추가**하여 pull-only → push-first (write-surface 정책은 SPEC-3 소관, 본 SPEC은 표면화 시점만 변경)
- **AP-9**: "게이트/배선 통과" 주장을 실행 출력 없이 보고 — verification-claim-integrity §1 위반

---

## §H Cross-References

- `.moai/specs/SPEC-HARNESS-EVO-PIPE-REPAIR-001/` — Epic 1/4 (완결, a661da107); 어휘/스키마 SSOT 정렬 + doctor 스모크 게이트 도입 (본 SPEC의 전제)
- `.moai/specs/SPEC-V3R6-HARNESS-V4-001/` — v4 Builder 4-phase + manifest Runner (manifest 스키마 원 소유; `learning` 필드는 본 SPEC이 옵션 확장)
- `.moai/specs/SPEC-V3R6-HARNESS-PROPOSAL-GEN-001/` — mapper/proposalgen 계약 (findings→proposal 변환 정합 대상)
- `.moai/specs/SPEC-HARNESS-LOOP-CLOSURE-001/` — C1 헌법 제약(auto_apply:false); write-surface 정책 개정은 SPEC-3 소관 (본 SPEC 무관)
- `.moai/specs/SPEC-HARNESS-EVO-WRITE-SURFACE-001` (미작성) — write-surface 개방 + 헌법 amendment (§E forward-link)
- `.moai/specs/SPEC-HARNESS-EVO-REQ-ARTIFACT-001` (미작성) — 요구사항 아티팩트 스키마 + 레거시 retire (§E forward-link)
- `internal/harness/v4manifest/types.go` — Manifest 스키마 SSOT
- `internal/cli/harness/doctor.go` — 스모크 게이트 (PIPE-REPAIR M3)
- `internal/harness/learner.go:96` — `defaultConfidence` (§E 제외, 별도 후속 SPEC)
- `.claude/rules/moai/core/askuser-protocol.md` — AskUserQuestion 오케스트레이터-단독 경계 (REQ-HRR-008)
- `.claude/rules/moai/core/verification-claim-integrity.md` §1 — 미측정 confidence 위장 금지 (REQ-HRR-004)
- CLAUDE.local.md §2 Template-First / §21 dev-only isolation / §24 harness namespace / §25 neutrality

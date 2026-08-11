---
id: SPEC-HARNESS-EVO-RUN-REPORT-001
title: "하네스 실행→학습 배선 (manifest learning / Runner findings / specialist emission / post-run push) — Epic Harness-Evolution 2/4"
version: "0.2.0"
status: draft
created: 2026-07-03
updated: 2026-08-12
author: manager-spec
priority: P2
phase: "v3.1.0"
module: "internal/harness/v4manifest"
lifecycle: spec-anchored
tier: M
tags: "harness, learning-loop, findings, manifest, runner, specialist, post-run, template-first"
related_specs: [SPEC-HARNESS-EVO-PIPE-REPAIR-001, SPEC-V3R6-HARNESS-V4-001, SPEC-V3R6-HARNESS-PROPOSAL-GEN-001, SPEC-HARNESS-LOOP-CLOSURE-001, SPEC-HARNESS-LEARNING-EVO-001, SPEC-HARNESS-LEARNING-EVO-002, SPEC-LSEL-LOCAL-EVOLUTION-001]
depends_on: [SPEC-HARNESS-EVO-PIPE-REPAIR-001]
---

# SPEC-HARNESS-EVO-RUN-REPORT-001: 하네스 실행→학습 배선

## HISTORY

| 버전 | 날짜 | 변경 내용 | 작성자 |
|------|------|-----------|--------|
| 0.1.0 | 2026-07-03 | 최초 작성 — Epic Harness-Evolution 2/4 (실행→학습 배선: manifest `learning` 블록 / Runner `findings` / specialist emission phase / post-run push). learner.go confidence 실측화는 §E로 명시 제외(별도 후속 SPEC) | manager-spec |
| 0.2.0 | 2026-08-12 | plan-phase REFRESH (scope 불변, 4 배선 항목 유지, status draft 유지). 축 ① `hns-` 개명(SPEC-HNS-PREFIX-RENAME-001 completed)에 맞춘 file:line 앵커 전부 재측정 — Runner `harness-release-update-run.js`→`hns-release-update-run.js`, specialist `harness-*-specialist.md`→`hns-*-specialist.md`, harness.md 앵커 재측정. 축 ② §F Milestones 를 LEARNING-EVO 001/002(completed) 분석기 계약과 재대조 — findings→proposal 경로를 proposalgen reserved-namespace 제3 producer(`harness_run:`)로 채택(delegation-map sibling); moai-harness-learner skill schema gap + LSEL sibling 구분 노트 추가. 축 ③ frontmatter priority P1→P2, phase v3.0.0→v3.1.0 (40일 slip, era.go H-5 tie-breaker load-bearing), updated 2026-08-12 | manager-spec |

---

## §A 배경

### §A.1 Epic 위치

본 SPEC은 Epic Harness-Evolution(4 SPECs)의 **2번**이다. Epic 전체 로드맵은 `SPEC-HARNESS-EVO-PIPE-REPAIR-001/plan.md §A.1` (완결됨, origin a661da107)이 정본이며, 본 SPEC의 범위는 그 표의 2번 행에서 파생된다:

> **SPEC-HARNESS-EVO-RUN-REPORT-001** — 실행→학습 배선: manifest `learning` 블록, Runner return-schema `findings`, specialist 필수 "improvement findings" 최종 단계, post-run findings 수집 → 즉시 AskUserQuestion push(현행 pull-only apply 대체).

### §A.2 이 SPEC이 닫는 끊어진 링크 (R4 재귀 자기개선 루프)

사용자 원 요청 R4는 "실행 중 발견된 개선 필요가 재귀적 자기개선 루프로 하네스를 갱신"이다. 8-표면 분석은 R4가 6개 끊어진 링크로 구조적 0 커버리지임을 확인했다. Epic 1/4(PIPE-REPAIR)는 그중 **파이프 계층 3개**를 수리했다: classify→propose 어휘 정렬(B1), pattern_key 스키마 정렬(B2), classify 자동 배선(B3). 즉 **관측→분류→제안 생성** 경로는 복구되었다.

그러나 그 경로는 **수동적(passive)** 학습이다 — Stop hook이 usage-log를 관측하고 tier promotion을 파생할 뿐, **하네스 실행(command→Runner→specialist end-to-end) 그 자체에서 발견된 개선 신호**는 학습 서브시스템으로 전달되지 않는다. 하네스 실행은 학습에 대해 벙어리다.

본 SPEC은 이 **실행→학습 배선**을 담당한다. 4개 배선 지점:

1. **manifest `learning` 블록** — 하네스가 "나는 개선 신호를 방출한다"를 선언하는 스키마 슬롯 (`v4manifest.Manifest`에 옵션 필드 추가).
2. **Runner return-schema `findings` 필드** — Runner가 실행 후 표준화된 개선 신호 배열을 반환.
3. **specialist 필수 improvement-findings 최종 단계** — specialist 워크플로우가 종료 직전 구조화된 개선 신호를 방출하는 필수 단계 추가.
4. **오케스트레이터 post-run 단계 표준화** — 현행 pull-only apply(사용자가 `/moai:harness apply`로 당겨오는 모델)를 **push 모델**(실행 종료 시 findings 수집 → 즉시 오케스트레이터 AskUserQuestion 제시)로 대체.

### §A.3 관측 베이스라인 (2026-08-12 plan-phase REFRESH 재실측 — verification-claim-integrity §2)

> 축 ① 갱신: 직전 draft(2026-07-03) 이후 SPEC-HNS-PREFIX-RENAME-001(completed)이 `harness-*` → `hns-*` 개명을 착지시켰다. 개명 규칙 — content-token 앵커(`Manifest`, `defaultConfidence`, `apply`/`Surface next`)는 유효, file:line 앵커는 drift하므로 전부 재측정. case-sensitive 매칭(`hns-` 소문자 ≠ `REQ-HNS-*`/`AC-HNS-*` 대문자). 디렉터리명 `.claude/agents/harness/`·`.claude/commands/harness/`, `/harness:` namespace, `harness.yaml`, `moai-harness-learner` skill, `.claude/skills/moai/workflows/harness.md` workflow doc은 **보존**(content-token).

- `.moai/harness/learning-history/tier-promotions.jsonl`: **16 promotion 레코드** (`wc -l` 실측), 전량 `confidence: 1` (`grep -o '"confidence":[0-9.]*' | sort | uniq -c` → `16 "confidence":1`)
- `internal/harness/learner.go:96`: `const defaultConfidence = 1.0` — 하드코딩 (재실측 확인, §E 제외 유지)
- `internal/harness/v4manifest/types.go:18-48`: `Manifest` 8-필드 구조체(`Name/Domain/SourceRequest/Patterns/Specialists/SprintContract/EntryCommand/RunnerWorkflow`), `learning` 필드 **부재** (재실측 확인)
- `.claude/commands/harness/release-update/manifest.json`: 8-키(domain/entry_command/name/patterns/runner_workflow/source_request/specialists/sprint_contract), `learning` 키 **부재** (재실측)
- `.claude/workflows/hns-release-update-run.js:89-101`: return `{manifest, capability, sweep_target_count, impact_tables, note}` (`module.exports` at `:103`) — 표준 `findings` 계약 **부재** (재실측; 직전 `harness-release-update-run.js:82-94`에서 개명+드리프트 보정)
- `.claude/agents/harness/hns-release-update-specialist.md:170`: Phase 8 = "Completion" (Phase 0 Load State `:51` ~ Phase 8 `:170`, 9-단계 human-gated 워크플로우) — learning emission 단계 **부재** (재실측; `harness-release-update-specialist.md` → `hns-` 개명)
- `.claude/skills/moai/workflows/harness.md:113`: verb table의 `apply` 행 = "Surface next Tier-4 proposal via AskUserQuestion, apply on approval through 5-layer pipeline" — **pull-only** 모델 (재실측; 직전 `:106`에서 `:113`로 드리프트. `:37` = "every verb (`status`, `apply`, `rollback`, `disable`)" 본문)
- **축 ② 추가 베이스라인**: `internal/harness/delegationmap/proposal.go:37` `BuildCandidates` → `proposalgen.ProposalCandidate` 변환은 reserved namespace `delegation_map:`(`proposal.go:21`) 사용; `ProposalCandidate.Evidence map[string]any`(`proposalgen/types.go:71`)가 producer-specific 확장 seam. 본 SPEC findings producer는 이 seam 의 sibling(`harness_run:` namespace)으로 설계 재대조됨 — §F/§H 참조.

즉 실행 경로(command→Runner→specialist)와 학습 경로(observe→classify→propose) 사이에 **배선이 존재하지 않는다**. 본 SPEC이 그 배선을 놓는다.

---

## §B 문제 정의 (검증 앵커 — 2026-08-12 plan-phase REFRESH 재실측)

> 아래 앵커는 본 plan-phase REFRESH에서 파일을 직접 Read/Grep/jq하여 재실측했다 (verification-claim-integrity §2 baseline 귀속 — 2026-07-03 draft 이후 hns- 개명 + 40일 drift 보정). 사용한 명령은 각 항목에 명시. 라인 번호는 참고용이며 run-phase에서는 content-token 앵커로 재고정할 것.

### B1. manifest 스키마에 learning 선언 슬롯 부재

- `internal/harness/v4manifest/types.go` — `Manifest` 구조체(8 필드: `Name/Domain/SourceRequest/Patterns/Specialists/SprintContract/EntryCommand/RunnerWorkflow`). learning 관련 필드 없음 (전량 Read 실측)
- `.claude/commands/harness/release-update/manifest.json` — `jq 'keys'` → 8 키, `learning` 없음
- 결과: 하네스는 "개선 신호를 방출하는가"를 선언할 수단이 없다. Runner/specialist가 findings를 만들어도 스키마 차원에서 그것이 기대되는지/어느 tier로 승격되는지 알 방법이 없다.

### B2. Runner return-schema에 표준 findings 계약 부재

- `.claude/workflows/hns-release-update-run.js:89-101` — return 객체 = `{manifest, capability, sweep_target_count, impact_tables, note}` (전량 Read 재실측). `findings` 키 없음. 동 파일 `module.exports = { run, selectResearchSweepTargets, MANIFEST_PATH }` at `:103`
- 반환 계약이 하네스마다 임의(ad-hoc); 표준 shape 없음 (개명 전 `harness-release-update-run.js:82-94`에서 `hns-`+드리프트 보정)
- 결과: 오케스트레이터가 실행 후 Runner 반환에서 개선 신호를 **일관된 방식으로 추출할 수 없다**. 각 Runner가 서로 다른 필드명/구조로 반환하면 post-run 수집이 하네스별 특수 코드를 요구한다.

### B3. specialist 워크플로우에 improvement-findings 방출 단계 부재

- `.claude/agents/harness/hns-release-update-specialist.md:170` — "### Phase 8 — Completion" (Phase 0 Load State at `:51` ~ Phase 8 at `:170`, 9-단계 human-gated). Phase 8은 요약 출력만 (grep + Read 재실측; `harness-release-update-specialist.md` → `hns-` 개명)
- `.claude/agents/harness/hns-github-specialist.md`, `hns-release-specialist.md` — 동일 계열의 human-gated 워크플로우 (ls 재실측: `.claude/agents/harness/` 디렉터리 보존, 파일명만 `harness-*` → `hns-*` 개명. 비-harness specialist 4개 + hns-oss-docs-* 4개는 본 SPEC 범위 외)
- 결과: specialist가 실행 중 발견한 개선 신호(예: 반복 수동 단계, 문서-코드 드리프트, 마찰 지점)를 구조화하여 학습 서브시스템에 넘기는 계약이 없다. 발견은 사람 머릿속에만 남고 소실된다.

### B4. post-run apply가 pull-only (push 부재)

- `.claude/skills/moai/workflows/harness.md:37` — 본문 "every verb (`status`, `apply`, `rollback`, `disable`)"; `:113` — verb table의 `apply` 행 = "Surface next Tier-4 proposal via AskUserQuestion, apply on approval through 5-layer pipeline" (grep 재실측; 직전 `:34`/`:106`에서 `:37`/`:113`로 드리프트. harness.md 자체는 content-token으로 보존, 개명 대상 아님)
- 즉 개선 제안은 사용자가 **명시적으로 `/moai:harness apply`를 호출해야만** 표면화된다 (pull). 하네스 실행 직후 발견된 findings가 자동으로 사용자 앞에 놓이지 않는다
- 결과: 실행→발견→제안의 시간적 근접성이 끊긴다. 사용자는 개선 신호가 쌓였는지 알려면 별도로 status/apply를 pull해야 한다. R4의 "재귀적 자기개선 루프"는 실행 종료 시점에 신호를 push하지 않으면 닫히지 않는다

### B5. (참고) doctor 스모크 게이트의 learning 축 부재

- `internal/cli/harness/doctor.go` — PIPE-REPAIR M3이 추가. 4축 cross-ref(command/manifest/Runner/agent) + ERROR/INFO severity 정책 (전량 Read 실측). learning 블록/findings 계약 검증 축 없음
- 결과: 새 `learning` 블록과 `findings` 계약이 스키마 유효한지 doctor가 검증하지 않으면, 배선이 조용히 깨져도(예: learning.tier 오타, findings shape 불일치) 탐지되지 않는다. 본 SPEC은 doctor 커버리지 확장을 in-scope로 포함한다 (REQ-HRR-005)

---

## §C GEARS 요구사항

> 요구사항 ID 접두: **REQ-HRR** (Harness-Run-Report). GEARS 표기 사용. `<subject>`는 일반화된 명사(엔진/스키마/Runner/specialist/오케스트레이터).

### 스키마 배선 (Go)

**REQ-HRR-001 (manifest `learning` 블록 스키마)** — The `v4manifest.Manifest` 스키마는 하네스가 개선 신호 방출을 선언하는 옵션 `learning` 블록을 수용해야 한다(shall). **Where** manifest에 `learning` 블록이 존재하면, the schema shall 최소 다음 필드를 인식해야 한다: `enabled`(bool — 개선 신호 방출 여부), `tier`(enum — findings가 승격될 제안 tier, `Tier.String()` 어휘 SSOT에서 파생: `{rule, auto_update}`), `confidence_floor`(number — 제안 후보 자격 최소 confidence), `max_findings_per_run`(integer — 실행당 findings 상한). **Where** `learning` 블록이 부재하면, the schema shall 하위 호환으로 이를 "learning 미방출 하네스"로 해석해야 하며 기존 8-필드 manifest를 거부해서는 안 된다(shall not). (블록 shape 상세: plan.md §D-D1)

**REQ-HRR-002 (learning tier 어휘 SSOT 정합)** — The `learning.tier` 필드의 유효값 집합은 classifier `Tier.String()` 어휘(`internal/harness/types.go`, PIPE-REPAIR가 정렬한 SSOT)에서 파생되어야 한다(shall). The schema shall not `learning.tier`에 대해 별도 병렬 어휘(예: `recommendation`/`approval_required`)를 정의해서는 안 된다 — PIPE-REPAIR B1이 제거한 어휘 불일치를 재도입하지 않기 위함이다. **When** `learning.tier`가 `{observation, heuristic}`(pre-actionable)로 설정되면, the schema validation shall 이를 유효하되 non-actionable로 처리해야 한다.

### 실행 반환 배선 (Runner)

**REQ-HRR-003 (Runner return-schema `findings` 계약)** — The 동적 워크플로우 Runner(`.claude/workflows/hns-<name>-run.js` — `harness-<name>-run.js`에서 SPEC-HNS-PREFIX-RENAME-001로 개명)의 반환 객체는 표준화된 `findings` 필드(배열)를 포함해야 한다(shall). **When** Runner가 실행을 완료하면, the Runner shall 각 개선 신호를 최소 다음 shape로 반환해야 한다: `surface`(string — 개선이 적용될 파일/아티팩트), `kind`(enum — `{drift, gap, friction, defect}`), `summary`(string — 한 줄 사람 판독 서술), `confidence`(number — Runner의 신호 confidence; learner.go 하드코딩과 무관한 실행 시점 측정치), `suggested_tier`(enum — `{rule, auto_update}`). **Where** 실행 중 개선 신호가 없으면, the Runner shall `findings: []`(빈 배열)를 반환해야 하며 필드를 생략해서는 안 된다(shall not) — 오케스트레이터가 필드 부재와 "신호 없음"을 구분할 수 있도록. (계약 상세 + findings→proposal 경로: plan.md §D-D2/§D-D4)

**REQ-HRR-004 (findings confidence의 출처 분리)** — The Runner가 방출하는 `findings[].confidence`는 실행 시점에 측정/추정된 값이어야 하며(shall), `learner.go`의 `defaultConfidence`(하드코딩 1.0) 상수를 재사용해서는 안 된다(shall not). **Where** Runner가 실행 시점에 confidence를 산출할 근거가 없으면, the Runner shall 명시적 보수 기본값(예: 0.70 = `confidence_floor` 경계)을 방출하고 그 값이 추정임을 findings 스키마 상 구분 가능하게 해야 한다 — 이는 §E에서 명시 제외된 learner.go 실측화와 **혼동되어서는 안 되는** 별개 계약이다(verification-claim-integrity §1: 미측정 confidence를 측정치로 위장 금지).

### 스모크 게이트 배선 (doctor)

**REQ-HRR-005 (doctor learning 축 검증)** — The `moai harness doctor` 스모크 게이트는 `learning` 블록과 `findings` 계약에 대한 참조 무결성 검사 축을 포함해야 한다(shall). **When** doctor가 `learning` 블록을 가진 하네스를 검사하면, the checker shall: (1) `learning.tier`가 `Tier.String()` 어휘 유효값인지, (2) `learning.confidence_floor`가 `[0,1]` 범위인지, (3) `learning.enabled: true`인 하네스의 Runner가 `findings` 반환 계약을 선언하는지(정규식 heuristic — JS AST 파싱 도입 금지, PIPE-REPAIR AP-2 계승)를 검증해야 한다. **Where** 하네스에 `learning` 블록이 없으면, the checker shall 이를 ERROR로 계상해서는 안 되며(shall not) INFO note로만 보고해야 한다 — learning은 옵션 기능이다. exit code는 ERROR-severity finding ≥ 1일 때만 비-0이다.

### 실행 방출 배선 (specialist)

**REQ-HRR-006 (specialist improvement-findings 필수 최종 단계)** — The harness specialist 에이전트(`.claude/agents/harness/hns-<name>-specialist.md` — `harness-<name>-specialist.md`에서 개명; 디렉터리 `.claude/agents/harness/`는 보존)의 워크플로우는 종료 직전 구조화된 improvement-findings를 방출하는 필수 최종 단계를 포함해야 한다(shall). **When** specialist가 모든 human-gated 작업 단계를 완료하면, the specialist shall REQ-HRR-003의 `findings` shape와 정합하는 개선 신호 목록(없으면 빈 목록 명시)을 방출해야 한다. The specialist shall not 이 단계에서 사용자에게 직접 질문해서는 안 되며(subagent boundary — CLAUDE.md §8), findings는 오케스트레이터가 소비할 구조화 출력으로만 반환해야 한다.

### 오케스트레이터 배선 (post-run push)

**REQ-HRR-007 (post-run findings 수집 → push)** — **When** 하네스 실행(command→Runner/specialist)이 종료되면, the orchestrator post-run 단계는 Runner/specialist가 방출한 `findings`를 수집하여 즉시 사용자에게 AskUserQuestion으로 제시해야 한다(shall) — 현행 pull-only apply 모델(사용자가 `/moai:harness apply`를 명시 호출해야 표면화)을 대체한다. **Where** 수집된 findings가 비어 있으면, the orchestrator shall AskUserQuestion을 발화하지 않고 조용히 진행해야 한다(불필요한 결정 피로 방지). **findings→proposal 경로(축 ② 설계 재대조, plan.md §D-D4)**: 본 SPEC의 harness-run findings는 `proposalgen` reserved-namespace 제3 producer(`harness_run:` namespace)로 routing된다 — LEARNING-EVO 002 `delegationmap.BuildCandidates`(`internal/harness/delegationmap/proposal.go:37`, `delegation_map:` namespace at `:21`)가 확립한 sibling 패턴을 계승하여 `ProposalCandidate`(`internal/harness/proposalgen/types.go:33`, `Evidence map[string]any` seam at `:71`)로 변환 → 기존 Tier-4 orchestrator AskUserQuestion 게이트로 unified disposition. 이 경로는 tier-ladder producer(`proposalgen.MapPromotions`)와 delegation-map producer와 **동일** Tier-4 게이트를 공유하며, moai-harness-learner skill(`.claude/skills/moai-harness-learner/SKILL.md`)의 flat 8-field apply payload schema(tier-ladder skill-text edit 전용)와는 **별개**이다 — 본 SPEC findings가 learner skill schema를 확장하지 않는다(known-interaction, §H 참조). (본 REQ의 doctrine/rule 표면 소관 — plan.md §D-D4; Go 코드 필수 아님, Builder/Runner/specialist 계약 수정이 본 SPEC 소관)

**REQ-HRR-008 (AskUserQuestion 오케스트레이터 단독 경계 보존)** — The post-run push 배선은 AskUserQuestion 오케스트레이터-단독 경계를 위반해서는 안 된다(shall not). The specialist/Runner shall not AskUserQuestion을 호출하거나 사용자에게 직접 질문해서는 안 되며(agent-common-protocol § User Interaction Boundary), findings 방출 → 오케스트레이터 수집 → 오케스트레이터 AskUserQuestion의 비대칭 경계를 준수해야 한다. **Where** push 대상 findings가 `learning.tier` actionable(`{rule, auto_update}`) 항목을 포함하면, the orchestrator AskUserQuestion shall 기존 Tier-4 application rate-limit(`harness.yaml` `rate_limit`: `max_per_week`, `cooldown_hours` — SSOT)을 준수해야 한다.

### 횡단 제약

**REQ-HRR-009 (Template-First + 중립성)** — **Where** 본 SPEC이 수정하는 표면이 template-managed이면, the run-phase shall template 대응 표면을 함께 갱신하고 `make build`를 실행해야 한다. The dev-only harness artifacts(`hns-*-run.js`, `.claude/agents/harness/hns-*-specialist.md`, `.claude/commands/harness/*/manifest.json`)는 §21/§24/§25 격리에 의해 template로 유출되어서는 안 된다(shall not) — 이들은 사용자 소유(user-owned) 또는 dev-only 로컬 전용이다 (파일명은 SPEC-HNS-PREFIX-RENAME-001로 `harness-*` → `hns-*` 개명됐으나 디렉터리 `.claude/agents/harness/`·`.claude/commands/harness/`와 namespace `/harness:`는 content-token으로 보존). **When** run-phase가 doctrine/rule/output-style 표면(template-managed)을 수정하면, the template content shall not 본 SPEC의 ID / REQ-HRR 토큰 / 감사 인용을 포함해서는 안 된다(§25 neutrality). (템플릿 mirror 대상 표면 판정은 plan.md §D-D5)

**REQ-HRR-010 (하위 호환 — 기존 하네스 무영향)** — The `learning` 블록과 `findings` 계약은 옵션이므로, **Where** 기존 하네스(learning 블록 없는 8-필드 manifest, findings 없는 Runner, improvement-findings 단계 없는 specialist)가 존재하면, the schema/Runner/doctor shall 이들을 유효한 legacy 하네스로 처리하고 정상 동작해야 한다(shall). The doctor shall not learning 블록 부재를 ERROR로 계상해서는 안 된다(REQ-HRR-005 재확인).

---

## §D 수용 기준

AC 매트릭스는 `acceptance.md` §D 참조 (AC-HRR-001 ~ AC-HRR-010 + REQ↔AC 추적표).

---

## §E 제외 범위 (Exclusions)

본 SPEC은 Epic Harness-Evolution의 2/4이며, 아래 항목은 명시적으로 범위 밖이다.

### Out of Scope — learner.go confidence 실측화 (별도 후속 SPEC)

- `learner.go`의 `defaultConfidence = 1.0`(하드코딩) 상수를 outcome-event 기반 실측 confidence로 대체하는 작업 (실데이터 전 16 레코드가 `confidence: 1`임을 실측 확인)
- outcome-event 인프라(하네스 실행 결과의 성공/실패/부분성 신호를 confidence로 환산하는 새 EventType/신호 파이프라인)는 **아직 존재하지 않는 신호**를 요구하는 별개 관심사이다
- 본 SPEC은 findings의 `confidence`를 실행 시점 측정치로 방출하는 계약(REQ-HRR-004)만 정의하며, learner 서브시스템의 promotion confidence 산출은 건드리지 않는다
- → 별도 후속 SPEC (예: `SPEC-HARNESS-EVO-CONFIDENCE-MEASURE-001` — 본 SPEC이 ID를 확정하지 않으며, Epic 후속으로 별도 착수)

### Out of Scope — write-surface 개방 + 헌법 제약 개정 (SPEC-3)

- `frozen_guard.go` `allowedPrefixes` 확장(`.claude/commands/harness/`, `.claude/workflows/harness-*.js`, specialists) — 단계적 write-surface 개방(M1 manifest-first → M2 full)
- snapshot + rollback + regression-gate 의무화
- LOOP-CLOSURE C1 헌법 제약(`auto_apply: false` per-item 승인)의 티어별 표면 자율(tiered per-surface autonomy)로의 supersede/amendment
- `.moai/docs/harness-namespace-doctrine.md`의 evolution-write 소유권 조항 추가
- → `SPEC-HARNESS-EVO-WRITE-SURFACE-001`

> 주의: 본 SPEC의 REQ-HRR-007 post-run push는 findings를 **제시(surface)**할 뿐이며, 승인 후 실제 파일 쓰기는 여전히 기존 5-layer apply pipeline(FrozenGuard + rate-limit + AskUserQuestion + snapshot)을 경유한다. write-surface 자체의 개방/자율화는 SPEC-3 소관이다. 본 SPEC은 pull을 push로 바꿀 뿐, 쓰기 권한 정책을 바꾸지 않는다.

### Out of Scope — 요구사항 아티팩트 스키마 + 레거시 retire (SPEC-4)

- manifest `source_request`(단일 raw string)의 구조화 요구사항 스키마(domain/goal/constraints/scope + AC + Discovery 응답 기록) 승격 + drift 감지 시 재-Discovery
- 레거시 5-layer marker 경로 retire, Builder 개선
- `harness-delivery-strategy.md` Model B rejection supersede 선언
- → `SPEC-HARNESS-EVO-REQ-ARTIFACT-001`

### Out of Scope — 학습 루프 알고리즘 개선 일반

- tier 승급 임계값(`{1,3,5,10}`) 조정, 새 EventType 추가, 관측 커버리지 확대, proposal 클러스터링/dedup 개선 등 "배선"을 넘는 알고리즘 개선은 본 SPEC 범위 밖
- findings→proposal 변환의 정교화(중복 findings 병합, 우선순위 랭킹)는 배선이 놓인 후의 별도 개선 관심사이다

---

## §F 가정 및 미검증 항목 (verification-claim-integrity)

- **`learning` 블록 필드 집합의 최소성**: 본 plan-phase에서 `{enabled, tier, confidence_floor, max_findings_per_run}` 4-필드를 최소 shape로 제안했으나, run-phase에서 실제 findings→proposal 변환 경로를 배선하며 필드 부족/과잉이 드러날 수 있다. run-phase M1에서 proposalgen `MapPromotions` 계약과 대조하여 필드 집합 확정 의무.
- **Runner return-schema 표준화의 소급 적용 범위**(축 ① 갱신): 현존 Runner는 `hns-release-update-run.js` 1개뿐(ls 재실측 — `harness-release-update-run.js`에서 SPEC-HNS-PREFIX-RENAME-001로 개명). `findings` 계약을 이 exemplar Runner에 적용하는 것이 dev-only 로컬 전용 수정인지(§21 격리), 아니면 v4 Builder가 생성하는 Runner 템플릿에도 반영되어야 하는지는 run-phase에서 Builder GENERATE 단계 계약과 대조 후 확정. (본 plan-phase 방향: Builder 생성 계약 = template-managed 반영, exemplar Runner 실제 수정 = dev-only)
- **post-run push의 배선 지점**(축 ② 재대조로 방향 확정): REQ-HRR-007의 findings→proposal 경로를 **proposalgen reserved-namespace 제3 producer(`harness_run:`)**로 채택했다(§D-D4). 이는 LEARNING-EVO 002 `delegationmap.BuildCandidates`(`internal/harness/delegationmap/proposal.go:37`, `delegation_map:` namespace)가 확립한 sibling 패턴의 계승이며, `ProposalCandidate.Evidence map[string]any`(`internal/harness/proposalgen/types.go:71`) seam으로 producer-specific 필드(`surface`/`kind`/`summary`)를 전달한다. doctrine/rule 표면(harness.md post-run 절 추가)은 이 producer의 출력을 오케스트레이터 Tier-4 게이트로 수렴하는 표현계로 유지된다. 직전 draft의 (B) 직접-push 기울임에서 (A) reserved-namespace producer로 정합 — 이유: 3 producer(tier-ladder/delegation-map/harness-run)가 동일 Tier-4 게이트와 rate-limit SSOT를 공유하는 구조적 일관성.
- **specialist emission 단계의 3개 specialist 일괄 적용**(축 ① 갱신): hns-{release-update, github, release}-specialist 3개(개명 후 파일명) 모두에 improvement-findings 단계를 추가하는지, 아니면 Runner 보유 하네스(release-update)만인지는 run-phase에서 각 specialist의 Runner 유무(release-update만 Runner 보유 — manifest 재실측)와 대조 후 확정.
- **라인 번호 드리프트**: 본 문서의 모든 `파일:라인` 앵커는 2026-08-12 REFRESH 시점 재실측이며, run-phase에서는 content-token(식별자/문자열 리터럴) 기준으로 재고정한다.
- **moai-harness-learner skill schema gap (축 ② known-interaction)**: `.claude/skills/moai-harness-learner/SKILL.md` payload schema는 flat 8-field(`proposal_id/target_path/field_key/current_value/new_value/observation_count/confidence/recommended_action`)이며 tier-ladder producer의 skill-text apply 전용으로 설계되었다(재실측). 본 SPEC의 harness-run findings는 이 schema를 사용하지 않고 `ProposalCandidate` → Tier-4 게이트 경로로 routing되므로, learner skill schema 확장은 **본 SPEC 범위 밖**(known-interaction 표기, forward-link 후속 SPEC). run-phase M4에서 schema gap이 실측 장애로 드러나면 blocker report → 별도 후속 SPEC.
- **LSEL sibling 구분 (축 ② known-interaction)**: LSEL(SPEC-LSEL-LOCAL-EVOLUTION-001)은 `.moai/lessons-inbox.jsonl` cluster → human-approved `decision.json` → `diff.patch`로 6개 evolvable surface(其中包括 `.claude/agents/harness/** + hns-* skills` per `.claude/lsel/frozen-allowlist.json`)를 개선하는 별개 루프이다. 본 SPEC의 harness-run findings는 **live push**(실행 종료 즉시 Tier-4 게이트)이며 LSEL의 batch cluster 모델과 **경로가 다르다**. LSEL applier(`internal/harness/applier.go:22` `enableTriggerInjectionWrites = false`)는 frozen 상태 유지 — 본 SPEC이 LSEL APPLY path를 변경하지 않는다. 두 루프의 surface 교집합(`.claude/agents/harness/hns-*`)은 run-phase 관찰 대상이나 본 SPEC이 통합하지 않는다.
- **PIPE-REPAIR 완결 전제**: 본 SPEC은 PIPE-REPAIR가 어휘/스키마를 정렬했음을 전제로 `Tier.String()` SSOT에서 `learning.tier` 어휘를 파생한다(REQ-HRR-002). PIPE-REPAIR origin a661da107 completed는 Epic 로드맵/git log로 확인했으나, `Tier.String()`의 최종 어휘가 `{observation, heuristic, rule, auto_update}`로 확정되었는지는 run-phase M1에서 `types.go`를 재-Read하여 재고정 의무.

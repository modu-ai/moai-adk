---
id: SPEC-AUDIT-GATE-INTEGRITY-001
title: "감사 게이트 무결성 P0 4결함 수정 — D7/D8 BLOCKING 배선, sync-auditor 채점모델 정합, 보고서 파일명 split-brain 해소, SPEC ID Bash 검증 전환"
version: "0.1.1"
status: draft
created: 2026-07-09
updated: 2026-07-09
author: GOOS행님
priority: P0
phase: "v3.1.x"
module: ".claude/agents/moai + .claude/rules/moai/workflow + internal/template/templates"
lifecycle: spec-anchored
tier: M
tags: "audit-gate, plan-auditor, sync-auditor, manager-spec, verification-claim-integrity, doc-only, template-mirror"
---

# SPEC-AUDIT-GATE-INTEGRITY-001 — 감사 게이트 무결성 4결함 수정

## HISTORY

| Version | Date | Author | Description |
|---------|------|--------|-------------|
| 0.1.0 | 2026-07-09 | GOOS행님 (via manager-spec) | 초안 — 3-agent 병렬 감사(2026-07-09) P0 4건 반영 |
| 0.1.1 | 2026-07-09 | GOOS행님 (via manager-spec) | plan-audit iter-1 결함 수정 (D1-D11): 언어중립 재설계(D3/MP-4), ERE escaped-pipe AC 수정(D1/D2), REQ-AGI-009 hash 메커니즘 교정(D4), `tier: M` 추가(D5), stale lint.go 라인 갱신(D6), mirror 다중-토큰 AC 강화(D7), 다절 REQ 토큰 추가(D8), 취약 AC 강건화(D9), placeholder 제거(D10), 어휘/GEARS 스타일 정합(D11) |

## §A Context

### A.1 문제 정의

2026-07-09 3-agent 병렬 감사(agent-definitions 감사 / workflow-doctrine 감사 / SDD 웹 리서치)에서 SPEC 감사 게이트의 무결성을 훼손하는 P0 결함 4건이 발견되었다. 공통 병리는 **"게이트가 실제로 게이트하지 않는다"** — BLOCKING 판정이 판정 계산에 배선되지 않거나, 검증 주장이 실측 없이 성립하거나, 동일 디렉터리에 두 파일명 규약이 문서 상호 참조 없이 공존한다.

### A.2 SDD 근거 (외부 표준 정합)

- **GitHub Spec Kit `/speckit.analyze` 패턴**: cross-artifact consistency 게이트는 advisory가 아니라 실제로 block해야 한다. BLOCKING finding이 aggregate score에 흡수되어 소멸하는 구조는 게이트 무결성 위반이다.
- **Kiro 3-artifact 모델 + EARS 인수 기준**: requirements / design / tasks 3-아티팩트와 기계 검증 가능한 EARS 인수 기준이 SDD 표준. 본 SPEC의 모든 AC는 명령 + 기대 출력으로 기계 검증된다.
- **`verification-claim-integrity.md` §1.1 surface 1-2**: 실측하지 않은 검증 주장(unobserved verification claim) 금지. R2의 "OWASP Top 10 compliance"(명령 부재)와 R4의 "mentally apply → PASS 보고"가 이 원칙의 직접 위반이다.

### A.3 결함 증거 (실측 anchor, 2026-07-09 기준)

| # | 결함 | 증거 위치 (실측) |
|---|------|------------------|
| R1 | plan-auditor D7/D8 BLOCKING이 Verdict에 미배선 | `.claude/agents/moai/plan-auditor.md` L127-137 (M5 = MP-1..MP-4만), L297-335 (D7 BLOCKING 방출), L337-370 (D8 BLOCKING 방출), L382-386 (Must-Pass Results 표에 MP-1..MP-4 행만) |
| R2 | sync-auditor 이중 채점모델 미정합 + 기계 검증 경로 부재 | `.claude/agents/moai/sync-auditor.md` L37-46 (flat 가중치 모델), L48-68 (flat 출력 형식만), L103-144 (HRN-003 계층 모델, worked example 부재), L42 (Security "OWASP Top 10 compliance" — 검증 명령 0건), frontmatter skills에 OWASP/testing 참조 스킬 부재 |
| R3 | plan-audit 보고서 파일명 split-brain | `.claude/rules/moai/workflow/spec-workflow.md` § Report Persistence (`<SPEC-ID>-<YYYY-MM-DD>.md`, `internal/runtime/audit_report.go` 구현 일치) vs `.claude/agents/moai/plan-auditor.md` L374·L471-472 + `.claude/skills/moai/workflows/plan/spec-assembly.md` L182·L196-211 (`{SPEC-ID}-review-{N}.md`) — 동일 디렉터리, 상호 참조 0건 |
| R4 | manager-spec SPEC ID 자가 점검이 mental regex | `.claude/agents/moai/manager-spec.md` L157-184 (L162 "Apply the canonical regex mentally" → 실측 없는 PASS 자기 보고) |

### A.4 수정 방침

전면 doc-only(agent 정의 + rule/doctrine 편집). Go 코드 동작 무변경. 모든 편집은 live → template mirror(Template Content Neutrality strip 적용) → `make build` 순서.

## §B Requirements (GEARS)

### R1 — plan-auditor D7/D8 BLOCKING의 Must-Pass/Verdict 배선

채택 설계: **MP-5/MP-6 must-pass 행 추가 방식** (grep-stable 검증 가능. "normative clause만 추가" 대안은 기각 — Output Format 표와 M5 정의가 분리된 채 남아 동일 drift가 재발한다).

#### REQ-AGI-001
The plan-auditor agent definition (`.claude/agents/moai/plan-auditor.md` § M5 Must-Pass Firewall) shall define **(MP-5)** "no unresolved D7 BLOCKING finding" and **(MP-6)** "no unresolved D8 BLOCKING finding" as the 5th and 6th must-pass criteria, with the same non-compensable semantics as MP-1..MP-4 (ANY single must-pass failure = overall FAIL).

#### REQ-AGI-002
**When** a BLOCKING finding is emitted by Group 7 (D7 Cross-SPEC Reconciliation) or Group 8 (D8 Cross-Platform Discipline) during a plan-phase audit, the plan-auditor shall treat the finding as **must-pass-equivalent**: force `Verdict: FAIL` regardless of aggregate score, and fold the finding into `## Defects Found` at severity=critical. (어휘 정합: D7-4/D8-3은 미해소(unresolved) 위반만 BLOCKING으로 방출하므로 본 조항의 "emitted"와 REQ-AGI-001의 "unresolved"는 동일 상태를 지칭한다 — plan-auditor.md 편집 본문에서는 "emitted (unresolved)" 병기로 통일한다.)

#### REQ-AGI-003
The plan-auditor Output Format `## Must-Pass Results` table shall carry MP-5 and MP-6 result rows (each row citing the D7/D8 verification evidence or "no BLOCKING finding"), so a BLOCKING D7/D8 finding can never be silently absorbed into the aggregate score.

### R2 — sync-auditor 채점모델 정합 + 차원별 기계 검증 경로

#### REQ-AGI-004
The sync-auditor agent definition shall state the scoring-model selection rule normatively: the flat weighted-percentage model (Functionality 40% / Security 25% / Craft 20% / Consistency 15%) is the **default**; **Where** `harness.yaml` sets `evaluator_mode: hierarchical`, the HRN-003 hierarchical model applies instead; and the two models shall be related explicitly — the hierarchical model is a **sub-criteria refinement** of the same 4 canonical dimensions (관계 서술은 리터럴 토큰 `sub-criteria refinement`를 포함; canonical anchors 0.25/0.50/0.75/1.00, min/mean aggregation).

#### REQ-AGI-005
**Where** `evaluator_mode: hierarchical` is active, the sync-auditor shall render the hierarchical report format; the agent definition shall include a worked hierarchical-mode output example (heading literal: `### Hierarchical-Mode Output Example`) showing sub-criterion rows, canonical anchor scores, per-dimension aggregation, and the must-pass firewall verdict — so invocations under either mode produce consistent, comparable reports.

#### REQ-AGI-006
**While** scoring any of the 4 evaluation dimensions, the sync-auditor shall execute at least 1 dimension-specific mechanical verification command and cite its verbatim output as the Evidence cell (per `verification-claim-integrity.md` §1.1 surface 2 + §3.2). The agent definition shall specify the dimension-command table **language-neutrally** via **project-language auto-detection** (편집 본문에 리터럴 토큰 `project-language auto-detection` 포함): 프로젝트 언어를 자동 감지해 해당 언어 toolchain 명령을 실행하고, 미설치 도구는 우아하게 건너뛴다 — quality gate의 언어별 자동감지 규약(CLAUDE.md §7 Language-Specific Guidelines: Go `go vet`→`golangci-lint`→`go test` / Node.js `eslint`→`npm test` / Python `ruff`→`pytest` / Rust `cargo clippy`→`cargo test`)과 동일 패턴. 설계 출처 인용은 본 spec.md에 한정하며, sync-auditor.md 편집 본문은 자체 완결 표로 작성한다(외부 섹션 번호 의존 금지). 차원 매핑과 언어 예시(4개 언어 동등 열거 — 어떤 언어도 PRIMARY로 승격 금지, Go는 동등 예시 중 하나):

- Functionality: 프로젝트 언어 test runner 결과를 SPEC AC matrix와 대조 (예: Go `go test ./...` / Python `pytest` / Node.js `npm test` / Rust `cargo test`)
- Security: grep 기반 OWASP checklist probe(입력 검증/시크릿/인젝션 표면) + 의존성 manifest 감사 (언어 불문)
- Craft: 프로젝트 언어 coverage + linter (예: Go `go test -cover`+`golangci-lint run` / Python `pytest --cov`+`ruff` / Node.js coverage+`eslint` / Rust `cargo clippy`)
- Consistency: 프로젝트 언어 lint/format 결과 + naming convention grep (grep은 언어 불문)

이 언어-인지형 설계로 live 파일과 template mirror의 R2 편집 내용은 **동일하면서 중립적**이다 — 언어 편향 신규 유입 없음, live/template 분기 불필요.

#### REQ-AGI-007
The sync-auditor frontmatter `skills:` list shall include `moai-ref-owasp-checklist` and `moai-ref-testing-pyramid` (both verified to exist under `.claude/skills/`), so the Security/Craft dimension probes are grounded in preloaded reference material instead of unstated knowledge.

### R3 — plan-audit 보고서 파일명 split-brain 해소 (doc-only)

채택 설계: 두 스트림을 **의도된 별개 스트림으로 명문화**한다. plan-phase 적대적 리뷰 스트림 `{SPEC-ID}-review-{N}.md`(iteration 기반, spec-assembly/plan-auditor 소비)와 run-entry 게이트 스트림 `<SPEC-ID>-<YYYY-MM-DD>.md`(날짜 기반, `internal/runtime/audit_report.go` 구현)는 병존이 정당하다 — 결함은 병존이 아니라 상호 참조 0건과 skip-eligibility 소비 대상 미지정이다.

#### REQ-AGI-008
`.claude/rules/moai/workflow/spec-workflow.md` § Report Persistence and `.claude/agents/moai/plan-auditor.md` § Output Format shall each document BOTH report streams as deliberately distinct — using the literal tokens **"plan-phase review stream"** (`{SPEC-ID}-review-{N}.md`) and **"run-gate stream"** (`<SPEC-ID>-<YYYY-MM-DD>.md`) — with a mutual cross-reference to the other document.

#### REQ-AGI-009
**When** the Phase 0.5 Plan Audit Gate evaluates skip-eligibility ("most recent plan-auditor verdict" + artifact-hash check), the doctrine (`spec-workflow.md`) shall designate normatively — 실제 Go 메커니즘과 일치하게: (a) the **plan-phase review stream's final-iteration verdict** (리터럴 토큰 `final-iteration verdict`) is the input the run-gate consults; (b) skip-eligibility의 artifact-hash 검사는 **plan-artifact hash**(리터럴 토큰 `plan-artifact hash`)를 재계산·대조한다 — `internal/runtime/audit_cache.go` `ComputeHash`가 specDir의 plan artifacts(spec.md/plan.md/acceptance.md/tasks.md)를 whitespace-normalized SHA-256으로 해싱하며 cache key = (specID, planArtifactHash); (c) the **run-gate stream's date-file**은 verdict **기록 표면(record surface)**일 뿐 hash 대상이 아니다. This requirement is doc-only — `internal/runtime/audit_report.go` / `audit_gate.go` / `audit_cache.go` 동작은 무변경.

### R4 — manager-spec SPEC ID 자가 점검: mental → executed Bash

#### REQ-AGI-010
**When** the manager-spec agent prepares to `Write` or `Edit` a new SPEC document containing a SPEC ID in its YAML frontmatter, the agent shall execute the Bash one-liner `[[ "$ID" =~ ^SPEC(-[A-Z][A-Z0-9]*)+-[0-9]{3}$ ]] && echo PASS || echo FAIL` (Bash ERE에서 `\d`는 미지원이므로 `[0-9]` 사용 — `internal/spec/lint.go` `specIDPattern`(실측 L649, 2026-07-09)의 `\d{3}`과 의미 동일) and cite the command's **verbatim output** as the self-check evidence; the agent definition shall prohibit mental-only regex application as the sole basis of a PASS claim.

#### REQ-AGI-011
The manager-spec agent definition shall retain the 4-step decomposition guidance (decompose / apply regex / print `decomposition: ... → PASS|FAIL` trace / halt-or-proceed), the canonical regex literal, and exactly 1 worked example. The manager-spec agent definition shall compress the ~30 lines of historical-incident narrative (L32 chain context, 5-incident enumeration) out of the agent body — history moves to memory/lessons, not the always-loaded agent definition. M4 편집 시 동일 섹션의 stale 라인 인용(`lint.go:573`)도 `specIDPattern`(L649 실측) 기준으로 함께 교정한다.

### Cross-cutting — Template Mirror 동기화

#### REQ-AGI-012
**When** any live file touched by REQ-AGI-001..011 has a template mirror under `internal/template/templates/` (실측 확인: `plan-auditor.md`, `sync-auditor.md`, `manager-spec.md` — `internal/template/templates/.claude/agents/moai/`; `spec-workflow.md` — `internal/template/templates/.claude/rules/moai/workflow/`), the same edit shall be applied to the mirror with the Template Content Neutrality strip (내부 SPEC ID, REQ 토큰, 감사 인용, 내부 날짜/SHA 제거 — CI guard `template-neutrality-check.yaml`), followed by `make build` exiting 0.

## §C Constraints

- **doc-only**: `internal/runtime/audit_report.go`, `internal/spec/lint.go`, `internal/spec/era.go` 등 Go 코드 동작 무변경. `make build`는 임베드 재컴파일 목적일 뿐 동작 변경이 아니다.
- **AC 기계 검증 의무**: 모든 AC는 정확한 명령 + 기대 출력(카운트/exit code)을 명기 (`verification-claim-integrity.md` §3.2).
- **Template Content Neutrality**: mirror 편집 시 `SPEC-AUDIT-GATE-INTEGRITY` 토큰·REQ-AGI 토큰·내부 날짜가 template 트리에 유입 금지.
- **grep-stable 토큰 설계**: 수정으로 도입되는 문구는 AC grep이 안정적으로 잡을 리터럴 토큰(`(MP-5)`, `(MP-6)`, `must-pass-equivalent`, `severity=critical`, `### Hierarchical-Mode Output Example`, `sub-criteria refinement`, `project-language auto-detection`, `plan-phase review stream`, `run-gate stream`, `final-iteration verdict`, `plan-artifact hash`, `=~ ^SPEC`)을 포함해야 한다.
- **언어 중립(템플릿 배포 자산)**: sync-auditor 차원-명령 표는 project-language auto-detection 형식으로 작성하며 특정 언어를 PRIMARY로 승격하지 않는다 (CLAUDE.local.md §15 16-언어 동등 [HARD]). Go 리터럴은 4-언어 동등 예시의 하나로만 등장한다. 이로써 R2 편집은 live/template 동일 내용으로 미러링된다.
- **frontmatter 12 canonical fields** + GEARS compound-clause 형식 준수 (plan-auditor MP-2/MP-3 통과 요건).

## Out of Scope (제외 범위)

### Out of Scope — Go 코드 동작 변경
- `internal/runtime/audit_report.go`의 날짜 기반 파일명 로직 + `audit_gate.go`/`audit_cache.go`의 plan-artifact hash 로직 변경 금지 (R3는 문서 층에서 두 스트림과 hash 메커니즘을 명문화할 뿐)
- `internal/spec/lint.go` SPEC ID regex, `FrontmatterSchemaRule`, `internal/spec/era.go` 파서 무변경
- sync-auditor HRN-003 계층 채점의 Go 구현(evaluator profile loader 등) 무변경

### Out of Scope — 감사 P1/P2 발견 항목
- 2026-07-09 3-agent 감사에서 P0로 분류되지 않은 모든 발견 항목은 본 SPEC 범위 밖 (후속 SPEC 소관)

### Out of Scope — spec-assembly.md 스트림 개편
- `.claude/skills/moai/workflows/plan/spec-assembly.md`는 이미 plan-phase review 스트림(`review-{N}`)을 일관되게 사용하므로 필수 편집 대상이 아니다. 상호 참조 1줄 추가는 run-phase 재량(conditional)이며 AC에 미포함

### Out of Scope — 신규 lint rule / hook 추가
- MP-5/MP-6 준수를 기계 강제하는 신규 Go lint rule 또는 hook은 본 SPEC에서 만들지 않는다 (문서-층 배선이 deliverable)

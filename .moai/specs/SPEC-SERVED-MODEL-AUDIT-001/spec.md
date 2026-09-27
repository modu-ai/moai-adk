---
id: SPEC-SERVED-MODEL-AUDIT-001
title: "서브에이전트 서빙 모델 관측 — 선언 모델만 보는 감사 로그의 사각 해소와 감사관 판정 채택 거부"
version: "0.1.0"
status: draft
created: 2026-09-27
updated: 2026-09-27
author: manager-spec
priority: P1
phase: "v3.2.0 target"
module: "internal/hook, internal/auditreceipt, internal/config, internal/cli, .claude/agents/moai"
lifecycle: spec-anchored
tags: "hook, subagent-stop, served-model, agent-model-audit, auditor, adoption-gate, opt-in-gate, doctor"
tier: M
era: V3R6
related_specs: [SPEC-AGENT-MODEL-ENFORCE-001, SPEC-CODEX-AUDIT-GATE-AXES-001, SPEC-WORKTREE-STATE-ROOT-001]
---

# SPEC: 서브에이전트 서빙 모델 관측과 감사관 판정 채택 거부

## HISTORY

| Version | Date | Author | Description |
|---------|------|--------|-------------|
| 0.1.0 | 2026-09-27 | manager-spec | 최초 draft. 착수 판정서(`.moai/reports/t1282/verdict.md`)의 재현 3세션과 plan 단계 추가 실측(§A.2)에 근거. 리드 결정 (c) — 관측 지점 2곳(SubagentStop + 사후 스캔), 기본은 기록·경고, 감사관 한정 opt-in 채택 거부 — 를 요구사항으로 옮김. |

---

## §A 배경과 동기

### §A.1 결함

PreToolUse 에이전트 모델 감사(`internal/hook/agent_model_guard.go:237` `checkAgentModel`)는 spawn 페이로드의 **선언 모델**과 프로필이 해석한 **해석 모델**만 비교해 `.moai/logs/agent-model-audit.jsonl`에 한 행을 남긴다(`appendAgentModelAudit`, 같은 파일 173행; 레코드 스키마 160-167행 — `declared_model`, `resolved_model`, `verdict`). 그 서브에이전트에 **실제로 응답한 모델**은 어디에서도 관측되지 않는다. 서빙 모델은 응답이 도착한 뒤에야 생기므로 PreToolUse 시점에는 원리상 알 수 없다.

그 결과 감사 로그는 선언과 해석이 같기만 하면 `ok`를 기록한다. 착수 판정서가 재현한 사례:

- 세션 `d46e0166`, manager-develop — 선언 `opus`, 감사 판정 `ok`, 트랜스크립트의 assistant 응답 85건 전부 `glm-5.3-flash`
- 세션 `6754629d`(Explore, 45건)와 `0729f3f3`(manager-develop, 79건) — 같은 모양

이 결함은 감사관 판정의 신뢰성으로 번진다. 선언은 opus 였으나 실제로는 GLM 이 쓴 plan-audit·sync-audit 판정이 아무 신호 없이 채택될 수 있다.

### §A.2 plan 단계 실측 (이 트리, 이 세션)

| # | 관측 | 명령 요지 | 결과 |
|---|------|-----------|------|
| M1 | 최근 3일 서브에이전트 트랜스크립트 수 | `find <projects> -path '*/subagents/agent-*.jsonl' -mtime -3` | 203개 |
| M2 | 그 트랜스크립트의 assistant 행 `.message.model` 분포 | `jq -r 'select(.type=="assistant")\|.message.model'` 집계 | `claude-opus-5-5` 12208, `claude-opus-5` 3751, `glm-5.3-flash` 3671, `claude-sonnet-5` 1255, `glm-5.3` 395, `<synthetic>` 10 |
| M3 | 감사관 실행별 선언(meta.json `model`) 대 서빙 집합 | meta.json `agentType` ∈ {plan-auditor, sync-auditor} | 선언 opus → 서빙 `glm-5.3-flash` 인 감사관 실행 **10건**(plan-auditor 2, sync-auditor 8) |
| M4 | 서브에이전트 입력 필드 | `internal/hook/types.go:239-241` | `agent_id`, `agent_transcript_path`, `last_assistant_message` 필드가 이미 파싱됨 |

M2 의 `<synthetic>` 는 모델 응답이 아닌 런타임 합성 행이므로 서빙 모델 판정에서 제외해야 한다(REQ-SMA-001). M3 은 판정서가 언급한 3건(t1237·t1239·t1099)보다 넓은 범위에서 같은 모양이 반복됨을 보인다 — 개별 카드와의 대조는 하지 않았다(§A.4).

### §A.3 기존 배선 (재사용 대상)

| 위치 | 역할 | 본 SPEC 과의 관계 |
|------|------|-------------------|
| `internal/hook/agent_model_guard.go:54` `agentModelAuditFileName` | 감사 로그 파일명 | 서빙 관측 행도 같은 파일에 적재 |
| `internal/hook/agent_model_guard.go:113` `resolveAgentModel` | 프로필 해석기의 유일한 진입점 | SubagentStop 에서도 같은 해석기로 해석 모델을 얻음 |
| `internal/hook/subagent_stop.go:38` `Handle` | SubagentStop 진입점. `checkAuditorStop` 을 먼저 평가하고 `mergeAuditorStopGuard`(53행)로 병합 | 서빙 관측을 이 진입점에 추가 |
| `internal/hook/audit_receipt_guard.go:116` `checkAuditorStop` | 감사관 최종 메시지의 판정 줄을 읽는 기존 판정 판독기 | 감사관 판정 채택 거부의 선례·연결점 |
| `internal/hook/audit_receipt_guard.go:199` `persistAuditRejection` / `internal/auditreceipt/store.go:123` `Rejection` | 채택 거부 기록(서브에이전트보다 오래 남음) | 서빙 모델 거부도 같은 저장소에 기록 |
| `internal/hook/audit_receipt_guard.go:241` `checkAuditReceiptSpawn`, 호출부 `internal/hook/pre_tool.go:724` | 거부 기록이 남아 있는 동안 phase 진입 spawn(manager-develop / manager-docs / manager-git) 차단 | 서빙 모델 거부가 읽히는 경로 |
| `internal/config/types.go:481` `AgentModelGuard` / `:795` `AgentModelGuardConfig` / `internal/config/defaults.go:1094` | opt-in 가드 설정 관례 | 새 키가 따르는 형태 |
| `internal/cli/doctor.go:201` `moaiChecks` | doctor 점검 등록부 | 사후 스캔 점검 등록 위치 |
| `internal/cli/binary_lag_test.go:198` `namesAddedAfterBaseline` | doctor 점검 이름 허용 목록 | 새 점검 이름 등록 대상 |

### §A.4 확인하지 않은 것 (plan 단계 Gap)

- SubagentStop 실제 페이로드에 `agent_transcript_path` 가 채워져 오는지는 실측하지 않았다. Claude Code 공식 훅 문서는 이 필드를 SubagentStop 입력으로 명시하지만, 이 머신의 훅 로그에는 포획 기록이 없다. 그래서 REQ-SMA-007 은 주 경로와 파생 경로를 함께 요구한다.
- SubagentStop 발화 시점에 트랜스크립트 마지막 assistant 행이 이미 디스크에 쓰였는지는 관측하지 않았다. 판정은 행 집합 기준이며 한 행의 누락은 판정을 바꾸지 않는 경우가 대부분이지만, 행이 0건이면 `unknown` 이 된다(REQ-SMA-004).
- 판정서가 든 세 감사 판정(t1237·t1239·t1099)은 개별 대조하지 않았다.

---

## §B 용어

- **서빙 모델(served model)**: 서브에이전트 트랜스크립트의 `type=="assistant"` 행이 담은 `.message.model` 값. 실제로 응답을 생성한 모델의 유일한 관측 소스다.
- **서빙 집합**: 한 서브에이전트 트랜스크립트에서 비어 있지 않고 `<synthetic>` 이 아닌 서빙 모델 값의 서로 다른 집합.
- **기대 모델**: 선언 모델이 있으면 선언 모델, 없으면 프로필 해석 모델.
- **서빙 판정**: `ok` / `served_drift` / `unknown` / `unmapped` 네 값.
- **채택 거부**: 이미 끝난 감사관 실행의 판정을 파이프라인이 근거로 받아들이지 않는 것. 실행을 되돌리거나 다시 돌리거나 보고서를 고치는 것이 아니다.
- **게이트 감사관**: `plan-auditor`, `sync-auditor`.

---

## §C 요구사항 (GEARS)

### §C.1 서빙 모델 관측

- **REQ-SMA-001** (Ubiquitous) — The served-model observer shall derive the served set of a subagent solely from the `.message.model` values of `type=="assistant"` rows in that subagent's own transcript, excluding empty values, absent fields, and the `<synthetic>` marker.
- **REQ-SMA-002** (Event-driven) — **When** a subagent stops, the SubagentStop hook shall append exactly one served-observation row to `.moai/logs/agent-model-audit.jsonl` carrying the session id, the agent id, the agent type, the declared model, the resolved model, the served set, the served verdict, and a source marker that distinguishes it from the PreToolUse row.
- **REQ-SMA-003** (Ubiquitous) — The served-model observer shall classify each observation as `ok` when every member of the served set matches the expected model, `served_drift` when any member does not match it, `unmapped` when neither a declared nor a resolved model exists to compare against, and `unknown` under the conditions of REQ-SMA-004; a served model shall match an expected model when the two are equal ignoring case, or when the expected model is a bare family alias and the served model is a model identifier of that family.
- **REQ-SMA-004** (Event-driven) — **When** the subagent transcript is absent or unreadable, or the transcript contains no assistant row with a non-empty, non-`<synthetic>` model value, the served-model observer shall record the verdict `unknown` and shall not record `ok`.
- **REQ-SMA-005** (Event-driven) — **When** a served verdict is `served_drift` or `unknown`, the SubagentStop hook shall emit a non-blocking warning naming the agent type, the expected model, and the served set.
- **REQ-SMA-006** (Unwanted) — The served-model observation shall not return a block decision to the subagent, shall not fail or delay the hook on an observation error, and shall not modify or remove any existing row of the agent-model audit log.
- **REQ-SMA-007** (Ubiquitous) — The served-model observer shall locate the subagent transcript from the hook input's `agent_transcript_path`, falling back to `subagents/agent-<agent_id>.jsonl` under the directory named by the hook input's `transcript_path` with its `.jsonl` suffix removed, and shall read the declared model from the sibling `agent-<agent_id>.meta.json`.
- **REQ-SMA-008** (Event-driven) — **When** reading the transcript exceeds the observer's bounded read budget, the served-model observer shall stop reading and record the verdict `unknown`.

### §C.2 게이트 감사관 판정 채택 거부 (opt-in)

- **REQ-SMA-009** (Ubiquitous) — The configuration shall expose an opt-in key `workflow.served_model_gate.enabled` whose engine default is `false`, whose distributed template value is `false`, and whose value in this repository's local configuration is `true`.
- **REQ-SMA-010** (State-driven) — **While** `workflow.served_model_gate.enabled` is `false` or absent, the SubagentStop hook shall record and warn per REQ-SMA-002 and REQ-SMA-005 and shall neither persist an adoption refusal nor cause any spawn to be denied.
- **REQ-SMA-011** (Compound) — **Where** `workflow.served_model_gate.enabled` is `true` in the tree the auditor ran in, **when** a gate auditor stops with a served verdict of `served_drift` or `unknown`, the SubagentStop hook shall persist an adoption-refusal record for that auditor role and SPEC in that tree's audit store, with a cause that names the served-model condition.
- **REQ-SMA-012** (Ubiquitous) — The adoption refusal shall refuse to adopt the auditor's verdict and shall not undo, re-run, interrupt, or rewrite the completed audit run or its report.
- **REQ-SMA-013** (State-driven) — **While** a served-model adoption refusal of a tree is outstanding, the PreToolUse hook shall deny phase-entry spawns (`manager-develop`, `manager-docs`, `manager-git`) from that tree with a reason that names the refused auditor role, the SPEC, and the served-model cause.
- **REQ-SMA-014** (Event-driven) — **When** a later run of the same gate auditor role stops in the same tree with a served verdict of `ok`, the SubagentStop hook shall clear that role's served-model refusals in that tree; a receipt-proven PASS alone shall not clear a served-model refusal, and a served verdict of `ok` alone shall not clear a receipt refusal.
- **REQ-SMA-015** (Unwanted) — The served-model gate shall not persist an adoption refusal for, nor deny a spawn on account of, any agent other than the two gate auditors.

### §C.3 사후 스캔

- **REQ-SMA-016** (Event-driven) — **When** `moai doctor` runs, a read-only diagnostic shall scan the subagent transcripts of the current project under every Claude configuration base the auto-memory resolver already considers, classify each per REQ-SMA-003 and REQ-SMA-004, and report the count per verdict together with every gate-auditor run whose verdict is `served_drift` or `unknown`.
- **REQ-SMA-017** (Unwanted) — The post-hoc diagnostic shall not write to the audit log, the audit store, or any transcript, and shall not change doctor's exit status on account of a `served_drift` or `unknown` finding.

### §C.4 감사관 자기 보고

- **REQ-SMA-018** (Ubiquitous) — The plan-auditor and sync-auditor agent definitions shall require the first line of every audit report to be `auditor-model: <served model>`, in the local copies, the distributed template copies, and the generated Codex copies.
- **REQ-SMA-019** (Ubiquitous) — The served-model observer shall record the auditor's self-reported `auditor-model` value alongside the served set and shall never let that self-reported value override or replace the transcript observation in the served verdict.

---

## §D 비기능 요구사항

- **fail-open**: 관측 경로의 모든 실패(경로 미해석, 파일 부재, JSON 파싱 오류, 로그 쓰기 실패)는 조용히 계속한다 — REQ-SMA-006. 단 판정 값은 `unknown` 으로 남아 성공으로 위장되지 않는다 — REQ-SMA-004.
- **시간 상한**: SubagentStop 훅 래퍼 timeout 은 5초다(`internal/template/templates/.claude/settings.json.tmpl:217`). 관측은 제한된 읽기 예산 안에서 끝나야 하며 예산 초과는 `unknown` 이다 — REQ-SMA-008. 실측 트랜스크립트 최대 크기는 3.4MB 급이다(판정서 세션 `d46e0166`).
- **템플릿 중립성**: `internal/template/templates/**` 에 들어가는 문구에는 SPEC ID·카드 ID·날짜·내부 경로를 넣지 않는다.
- **프로그래밍 언어 중립성**: 감사관 정의에 추가되는 문구는 특정 프로그래밍 언어를 전제하지 않는다.

---

## §E 제외 범위 (Exclusions)

이 절은 이번 SPEC 에서 만들지 않는 것을 적는다.

### Out of Scope — 강제 수단

- 서빙 모델이 기대와 다를 때 서브에이전트를 중단하거나 다시 실행하는 것 — 채택 거부만 한다.
- 비감사관 에이전트(manager-develop, Explore 등)의 결과 채택 거부 — 경고와 기록만 한다.
- PreToolUse 시점에 서빙 모델을 예측하거나 spawn 을 막는 것 — 원리상 관측 불가.
- `auditor-model` 첫 줄의 부재나 자기 보고 불일치를 거부 사유로 삼는 것 — 기록만 한다.

### Out of Scope — 데이터 보정과 백필

- 사후 스캔이 과거 서브에이전트 실행에 대해 감사 로그 행을 소급 기록하는 것.
- 이미 채택된 과거 감사 판정(t1237·t1239·t1099 포함)을 재판정하거나 무효 표시하는 것.
- 기존 PreToolUse 감사 행의 스키마 변경이나 재작성.

### Out of Scope — 백엔드 정책

- GLM 백엔드를 의도적으로 쓰는 사용자의 경고 억제 정책 — 기본값 OFF 로 거부만 막고, 경고 문구 조정은 후속 카드 몫이다.
- effort 관측 — Agent 도구에 effort 인자가 없다(기존 `agent_model_guard.go` 28-29행과 같은 이유).
- 모델 라우팅 자체의 수정(프로필 매트릭스, `moai glm` 환경 변수 배선).

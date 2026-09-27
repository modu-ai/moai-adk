---
id: SPEC-SERVED-MODEL-AUDIT-001
title: "서브에이전트 서빙 모델 관측 — 선언 모델만 보는 감사 로그의 사각 해소와 감사관 판정 채택 거부"
version: "0.2.0"
status: in-progress
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
| 0.2.0 | 2026-09-27 | manager-spec | plan-audit iter1(FAIL 0.75) 반영. D3: Tier M 예산에 맞춰 REQ 19→16, AC 25→16 으로 통합(리드 결정 동작은 하나도 빼지 않음, 형제 부정 사례는 표 기반 AC 로 묶음). D2: 서빙 게이트 전용 스코프 + 영수증 가드 결합 금지 REQ-SMA-013. D4: 해석 모델 출처를 PreToolUse 와 같은 설정 제공자로 고정(REQ-SMA-003). D1/D11: 워크트리 slug 열거 규칙(REQ-SMA-014)과 §A.2 실측 원문. D5: 자기 보고를 보고서 파일과 최종 메시지 양쪽에 고정(REQ-SMA-015/016). D6: 센티널·병합 사유(REQ-SMA-010). D9: 빈 스캔은 ok 가 아님(REQ-SMA-014). D10: Kind 부재 레코드는 영수증 종류(REQ-SMA-011). 규칙 문서 편집(M7)은 sync 단계로 넘김(§E). |

---

## §A 배경과 동기

### §A.1 결함

PreToolUse 에이전트 모델 감사(`internal/hook/agent_model_guard.go:237` `checkAgentModel`)는 spawn 페이로드의 **선언 모델**과 프로필이 해석한 **해석 모델**만 비교해 `.moai/logs/agent-model-audit.jsonl`에 한 행을 남긴다(`appendAgentModelAudit`, 같은 파일 173행; 레코드 스키마 160-167행 — `declared_model`, `resolved_model`, `verdict`). 그 서브에이전트에 **실제로 응답한 모델**은 어디에서도 관측되지 않는다. 서빙 모델은 응답이 도착한 뒤에야 생기므로 PreToolUse 시점에는 원리상 알 수 없다.

그 결과 감사 로그는 선언과 해석이 같기만 하면 `ok`를 기록한다. 착수 판정서가 재현한 사례:

- 세션 `d46e0166`, manager-develop — 선언 `opus`, 감사 판정 `ok`, 트랜스크립트의 assistant 응답 85건 전부 `glm-5.3-flash`
- 세션 `6754629d`(Explore, 45건)와 `0729f3f3`(manager-develop, 79건) — 같은 모양

이 결함은 감사관 판정의 신뢰성으로 번진다. 선언은 opus 였으나 실제로는 GLM 이 쓴 plan-audit·sync-audit 판정이 아무 신호 없이 채택될 수 있다.

### §A.2 plan 단계 실측 (이 머신, 최근 3일)

측정 스크립트(원문 그대로, `$HOME` 아래 CLAUDE_CONFIG_DIR 프로필 기준):

```bash
#!/bin/bash
# served-model measurement over primary slug vs worktree slugs, last 3 days
P="$HOME/.moai/claude-profiles/moai-adk/projects"
PRI="-Users-goos-MoAI-moai-adk-go"
echo "M1a primary-slug transcripts: $(find "$P/$PRI" -path '*/subagents/agent-*.jsonl' -mtime -3 | wc -l | tr -d ' ')"
echo "M1b worktree-slug transcripts: $(find "$P" -maxdepth 4 -path "*/$PRI--claude-worktrees-*/*/subagents/agent-*.jsonl" -mtime -3 | wc -l | tr -d ' ')"
echo "M1c worktree slug dirs: $(ls -d "$P/$PRI--claude-worktrees-"* 2>/dev/null | wc -l | tr -d ' ')"
count() { # $1 label, $2 find root glob
  while IFS= read -r f; do
    m="${f%.jsonl}.meta.json"
    t=$(jq -r '.agentType // ""' "$m" 2>/dev/null)
    case "$t" in plan-auditor|sync-auditor) ;; *) continue;; esac
    d=$(jq -r '.model // "-"' "$m" 2>/dev/null)
    s=$(jq -r 'select(.type=="assistant")|.message.model // empty' "$f" 2>/dev/null | grep -v '^<synthetic>$' | sort -u | paste -sd, -)
    echo "$1 $t declared=$d served=$s"
  done
}
find "$P/$PRI" -path '*/subagents/agent-*.jsonl' -mtime -3 | count primary | sort | uniq -c
find "$P" -maxdepth 4 -path "*/$PRI--claude-worktrees-*/*/subagents/agent-*.jsonl" -mtime -3 | count worktree | sort | uniq -c
```

`<synthetic>` 집계(같은 날, primary slug 최근 3일 파일 목록 `files.txt` 대상):

```bash
xargs -I{} jq -r 'select(.type=="assistant")|.message.model // "<ABSENT>"' {} < files.txt 2>/dev/null | sort | uniq -c | sort -rn | head -12
```

```
12208 claude-opus-5-5
3751 claude-opus-5
3671 glm-5.3-flash
1255 claude-sonnet-5
 395 glm-5.3
  10 <synthetic>
```

위 측정 스크립트의 출력 원문:

```
M1a primary-slug transcripts: 203
M1b worktree-slug transcripts: 200
M1c worktree slug dirs: 363
   2 primary plan-auditor declared=- served=claude-opus-5-5
   1 primary plan-auditor declared=- served=glm-5.3
   5 primary plan-auditor declared=opus served=claude-opus-5
  19 primary plan-auditor declared=opus served=claude-opus-5-5
   2 primary plan-auditor declared=opus served=glm-5.3-flash
   1 primary plan-auditor declared=sonnet served=claude-sonnet-5
  12 primary sync-auditor declared=- served=claude-opus-5-5
   2 primary sync-auditor declared=- served=glm-5.3
  28 primary sync-auditor declared=opus served=claude-opus-5-5
   8 primary sync-auditor declared=opus served=glm-5.3-flash
  23 worktree plan-auditor declared=- served=claude-opus-5-5
   2 worktree plan-auditor declared=opus served=claude-opus-5
  26 worktree plan-auditor declared=opus served=claude-opus-5-5
   3 worktree plan-auditor declared=sonnet served=claude-sonnet-5
   7 worktree sync-auditor declared=- served=claude-opus-5-5
  16 worktree sync-auditor declared=opus served=claude-opus-5-5
   2 worktree sync-auditor declared=opus served=glm-5.3-flash
   1 worktree sync-auditor declared=sonnet served=claude-sonnet-5
```

읽는 법:

- 트랜스크립트는 **git 루트가 아니라 세션이 시작된 경로의 slug** 아래 쌓인다. 워크트리에서 띄운 세션은 `<primary slug>--claude-worktrees-<name>` 아래에 기록된다. 최근 3일 기준 primary slug 203개, 워크트리 slug 200개로 절반가량이 워크트리 쪽이다. 사후 스캔 범위는 REQ-SMA-014 로 정한다.
- 선언 opus · 서빙 GLM 인 감사관 실행: primary slug 10건(plan 2, sync 8), 워크트리 slug 2건(sync 2). 선언 없음(`-`) · 서빙 `glm-5.3` 3건(primary)은 해석 모델이 opus 이므로 역시 drift 후보다.
- 위 `<synthetic>` 집계에서 assistant 행 `.message.model` 에 `<synthetic>` 이 10행 있었다(primary slug 203개 전체, 에이전트 유형 무관). 런타임 합성 행이므로 판정에서 제외한다(REQ-SMA-001).
- 서브에이전트 입력 필드 `agent_id`, `agent_transcript_path`, `last_assistant_message` 는 이미 파싱된다(`internal/hook/types.go:239-241`).

### §A.3 기존 배선 (재사용 대상)

| 위치 | 역할 | 본 SPEC 과의 관계 |
|------|------|-------------------|
| `internal/hook/agent_model_guard.go:54` `agentModelAuditFileName` | 감사 로그 파일명 | 서빙 관측 행도 같은 파일에 적재 |
| `internal/hook/agent_model_guard.go:113` `resolveAgentModel` / `:222` `llmConfig` | 프로필 해석기의 유일한 진입점과 그 설정 출처 | SubagentStop 도 같은 설정 제공자로 해석(REQ-SMA-003) |
| `internal/cli/deps.go:255` `NewPreToolHandlerWithScanner(deps.Config, …)` / `:273` `NewSubagentStartHandlerWithConfig(deps.Config)` / `:283` `NewSubagentStopHandler()` | 훅 핸들러 등록. PreToolUse·SubagentStart 는 설정 제공자를 받고 SubagentStop 은 받지 않음 | SubagentStop 에 같은 `deps.Config` 주입 |
| `internal/hook/subagent_stop.go:38` `Handle` | SubagentStop 진입점. `checkAuditorStop` 을 먼저 평가하고 `mergeAuditorStopGuard`(53행)로 병합 | 서빙 관측을 이 진입점에 추가 |
| `internal/hook/audit_receipt_guard.go:72` `auditReceiptScope` | 영수증 가드 3개 진입점(`:98`, `:120`, `:246`)이 공유하는 스코프 술어 | **넓히지 않는다** — 서빙 게이트는 별도 술어(REQ-SMA-013) |
| `internal/hook/audit_receipt_guard.go:116` `checkAuditorStop`, `:199` `persistAuditRejection` / `internal/auditreceipt/store.go:123` `Rejection` | 감사관 판정 판독과 채택 거부 기록 | 서빙 거부도 같은 저장소에 kind 를 달아 기록 |
| `internal/hook/audit_receipt_guard.go:241` `checkAuditReceiptSpawn`, 호출부 `internal/hook/pre_tool.go:724` | 거부 기록이 남아 있는 동안 phase 진입 spawn 차단 | 서빙 거부가 읽히는 경로 |
| `internal/config/types.go:481` `AgentModelGuard` / `:795` / `internal/config/defaults.go:1094` | opt-in 가드 설정 관례 | 새 키가 따르는 형태 |
| `internal/escalation/roots.go:21` `MemorySlug`, `:35` `memoryRoots` | 경로 → slug 규칙과 설정 기준 디렉터리 집합 | 사후 스캔의 slug·기준 계산 |
| `internal/cli/doctor.go:201` `moaiChecks` / `internal/cli/binary_lag_test.go:198` `namesAddedAfterBaseline` | doctor 등록부와 점검 이름 허용 목록 | 새 점검 등록과 허용 목록 등록 |

### §A.4 확인하지 않은 것 (plan 단계 Gap)

- SubagentStop 실제 페이로드에 `agent_transcript_path` 가 채워져 오는지는 실측하지 않았다. Claude Code 공식 훅 문서는 이 필드를 SubagentStop 입력으로 명시하지만 이 머신의 훅 로그에는 포획 기록이 없다. REQ-SMA-001 이 파생 경로를 함께 요구한다.
- SubagentStop 발화 시점에 트랜스크립트 마지막 assistant 행이 이미 디스크에 쓰였는지는 관측하지 않았다.
- 판정서가 든 세 감사 판정(t1237·t1239·t1099)은 개별 대조하지 않았다.
- L2 워크트리(`~/.moai/worktrees/…`)에서 띄운 세션의 slug 는 이번 측정 범위(primary slug 와 `--claude-worktrees-` 접두 slug)에 들어 있지 않다.

---

## §B 용어

- **서빙 모델(served model)**: 서브에이전트 트랜스크립트의 `type=="assistant"` 행이 담은 `.message.model` 값.
- **서빙 집합**: 한 트랜스크립트에서 비어 있지 않고 `<synthetic>` 이 아닌 서빙 모델 값의 서로 다른 집합.
- **기대 모델**: 선언 모델이 있으면 선언 모델, 없으면 해석 모델.
- **서빙 판정**: `ok` / `served_drift` / `unknown` / `unmapped`.
- **채택 거부**: 이미 끝난 감사관 실행의 판정을 파이프라인이 근거로 받아들이지 않는 것. 실행을 되돌리거나 다시 돌리거나 보고서를 고치는 것이 아니다.
- **거부 종류(kind)**: 채택 거부 기록의 출처 — `receipt`(기존 감사 영수증 가드) 또는 `served`(본 SPEC).
- **게이트 감사관**: `plan-auditor`, `sync-auditor`.

---

## §C 요구사항 (GEARS)

### §C.1 서빙 모델 관측

- **REQ-SMA-001** (Ubiquitous) — The served-model observer shall derive the served set of a subagent solely from the `.message.model` values of `type=="assistant"` rows in that subagent's own transcript, excluding empty values, absent fields, and the `<synthetic>` marker; it shall locate that transcript from the hook input's `agent_transcript_path`, falling back to `subagents/agent-<agent_id>.jsonl` under the directory named by the hook input's `transcript_path` with its `.jsonl` suffix removed, and shall read the declared model from the sibling `agent-<agent_id>.meta.json`.
- **REQ-SMA-002** (Event-driven) — **When** a subagent stops, the SubagentStop hook shall append exactly one served-observation row to `.moai/logs/agent-model-audit.jsonl` carrying the session id, the agent id, the agent type, the declared model, the resolved model, the served set, the served verdict, the self-reported model, and a source marker that distinguishes it from the PreToolUse row.
- **REQ-SMA-003** (Ubiquitous) — The served-model observer shall compute the resolved model through the same profile resolver and the same configuration provider the PreToolUse agent-model guard uses (active profile and per-agent overrides included), and shall classify each observation as `ok` when every member of the served set matches the expected model, `served_drift` when any member does not, and `unmapped` when the agent has neither a declared model nor a profile mapping; a served model shall match an expected model when the two are equal ignoring case, or when the expected model is a bare family alias and the served model is a model identifier of that family.
- **REQ-SMA-004** (Event-driven) — **When** the transcript is absent or unreadable, the transcript holds no assistant row with a non-empty, non-`<synthetic>` model value, reading the transcript exceeds the observer's bounded read budget, or no declared model exists and the configuration provider yields no configuration, the served-model observer shall record the verdict `unknown` and shall not record `ok`.
- **REQ-SMA-005** (Event-driven) — **When** a served verdict is `served_drift` or `unknown`, the SubagentStop hook shall emit a non-blocking warning naming the agent type, the expected model, and the served set.
- **REQ-SMA-006** (Unwanted) — The served-model observation shall not return a block decision to the subagent, shall not fail or delay the hook on an observation error, and shall not modify or remove any existing row of the agent-model audit log.

### §C.2 게이트 감사관 판정 채택 거부 (opt-in)

- **REQ-SMA-007** (Ubiquitous) — The configuration shall expose an opt-in key `workflow.served_model_gate.enabled` whose engine default is `false`, whose distributed template value is `false`, and whose value in this repository's local configuration is `true`.
- **REQ-SMA-008** (State-driven) — **While** `workflow.served_model_gate.enabled` is `false` or absent, the SubagentStop hook shall record and warn per REQ-SMA-002 and REQ-SMA-005 and shall neither persist a served-kind adoption refusal nor cause any spawn to be denied on a served-model ground.
- **REQ-SMA-009** (Compound) — **Where** `workflow.served_model_gate.enabled` is `true` in the tree the auditor ran in, **when** a gate auditor stops with a served verdict of `served_drift` or `unknown`, the SubagentStop hook shall persist a served-kind adoption-refusal record for that auditor role and SPEC in that tree's audit store; the refusal shall refuse to adopt the auditor's verdict and shall not undo, re-run, interrupt, or rewrite the completed audit run or its report.
- **REQ-SMA-010** (State-driven) — **While** a served-kind refusal of a tree is outstanding and the served gate of that tree is enabled, the PreToolUse hook shall deny phase-entry spawns (`manager-develop`, `manager-docs`, `manager-git`) from that tree with a reason that begins with the `SERVED_MODEL_VIOLATION` sentinel and names the refused auditor role, the SPEC, and the served-model cause; when receipt-kind and served-kind refusals are outstanding together, the reason shall carry both the `AUDIT_RECEIPT_VIOLATION` and the `SERVED_MODEL_VIOLATION` sentinels, each followed by its own outstanding records.
- **REQ-SMA-011** (Event-driven) — **When** a later run of the same gate auditor role stops in the same tree with a served verdict of `ok`, the SubagentStop hook shall clear only the served-kind refusals of that role in that tree; a receipt-proven PASS shall clear only receipt-kind refusals; a refusal record that carries no kind shall be read and cleared as a receipt-kind record; and a served-kind and a receipt-kind refusal for the same role and SPEC shall be stored so that neither overwrites the other.
- **REQ-SMA-012** (Unwanted) — The served-model gate shall not persist an adoption refusal for, nor deny a spawn on account of, any agent other than the two gate auditors.
- **REQ-SMA-013** (Unwanted) — The served-model gate shall not enable, block on, record a start marker for, or persist any audit-receipt check in a tree that does not declare the codex audit gate required; enabling `workflow.served_model_gate.enabled` shall leave the scope predicate of the audit-receipt guard unchanged.

### §C.3 사후 스캔

- **REQ-SMA-014** (Event-driven) — **When** `moai doctor` runs, a read-only diagnostic shall scan subagent transcripts under every Claude configuration base the auto-memory resolver considers, within every project directory whose name equals the slug of the primary checkout, begins with that slug followed by `--claude-worktrees-`, or equals the slug of a path listed by `git worktree list`; it shall classify each per REQ-SMA-003 and REQ-SMA-004, report the number of transcripts swept, the bases and slugs searched, the count per verdict, and every gate-auditor run whose verdict is `served_drift` or `unknown`; a sweep that finds no transcript shall report an informational status stating a swept count of zero and never an `ok` status; and the diagnostic shall not write to the audit log, the audit store, or any transcript, nor change doctor's exit status on account of any served-model finding.

### §C.4 감사관 자기 보고

- **REQ-SMA-015** (Ubiquitous) — The plan-auditor and sync-auditor agent definitions shall require the line `auditor-model: <served model>` as the first line of both the audit report file and the final message the auditor returns to its caller, in the local copies, the distributed template copies, and the generated Codex copies.
- **REQ-SMA-016** (Ubiquitous) — The served-model observer shall read the self-reported model from the first line of the hook input's `last_assistant_message`, record it alongside the served set, and never let the self-reported value override or replace the transcript observation in the served verdict.

---

## §D 비기능 요구사항

- **fail-open**: 관측 경로의 모든 실패는 조용히 계속한다 — REQ-SMA-006. 단 판정 값은 `unknown` 으로 남아 성공으로 위장되지 않는다 — REQ-SMA-004.
- **시간 상한**: SubagentStop 훅 래퍼 timeout 은 5초다(`internal/template/templates/.claude/settings.json.tmpl:217`). 관측은 제한된 읽기 예산 안에서 끝나야 하며 예산 초과는 `unknown` 이다 — REQ-SMA-004. 실측 트랜스크립트 최대 크기는 3.4MB 급이다(판정서 세션 `d46e0166`).
- **템플릿 중립성**: `internal/template/templates/**` 에 들어가는 문구에는 SPEC ID·카드 ID·날짜·내부 경로를 넣지 않는다.
- **프로그래밍 언어 중립성**: 감사관 정의에 추가되는 문구는 특정 프로그래밍 언어를 전제하지 않는다.
- **Tier 판정**: REQ 16 · AC 16 으로 Tier M 예산 안이다. 변경 파일 수는 Tier M 대역(5-15)을 넘는 18개 안팎이지만, 그중 C3 2개는 기계 방출물이고 C1/C2 4개와 템플릿·로컬 workflow.yaml 2개는 같은 한 줄 결정의 미러다. 판단이 들어가는 독립 편집 단위는 hook·auditreceipt·config·cli 4개 패키지로 Tier M 범위라서 Tier L 로 올리지 않았다.

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
- 이미 삭제된 L2 워크트리처럼 현재 `git worktree list` 에도 없고 `--claude-worktrees-` 접두도 아닌 과거 slug 의 트랜스크립트 — 사후 스캔이 보지 못한다.

### Out of Scope — 백엔드 정책과 문서

- GLM 백엔드를 의도적으로 쓰는 사용자의 경고 억제 정책 — 기본값 OFF 로 거부만 막고, 경고 문구 조정은 후속 카드 몫이다.
- effort 관측 — Agent 도구에 effort 인자가 없다(`agent_model_guard.go` 28-29행과 같은 이유).
- 모델 라우팅 자체의 수정(프로필 매트릭스, `moai glm` 환경 변수 배선).
- `.claude/rules/**` 의 감사 로그 설명 문장 갱신 — sync 단계(manager-docs) 몫이며 이 SPEC 의 인수 기준에 들지 않는다.

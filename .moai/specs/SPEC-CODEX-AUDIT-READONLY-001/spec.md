---
id: SPEC-CODEX-AUDIT-READONLY-001
title: "Codex read-only roles launched as top-level read-only processes — audit launcher, parent-written verdict file, inherited AC-DHR-012/023"
version: "0.1.0"
status: draft
created: 2026-09-24
updated: 2026-09-24
author: manager-spec
priority: P1
phase: "v3.2.0 target"
module: "internal/cli, internal/template/agentemit, internal/template/templates"
lifecycle: spec-anchored
tags: "codex, sandbox, read-only, audit, plan-auditor, sync-auditor, security, live"
tier: M
card: t1143
depends_on:
  - SPEC-DUAL-HARNESS-RECOVERY-001
related_specs:
  - SPEC-CODEX-LOCALMD-001  # owns the single developer_instructions producer and the 126,976-byte argument ceiling reused here
  - SPEC-CODEX-LAUNCHER-001
  - SPEC-FACTORY-LANE-WORKTREE-HANDOFF-001  # transfer-mechanics precedent (0.5.10 / 0.5.11)
---

# SPEC-CODEX-AUDIT-READONLY-001

## HISTORY

| Version | Date | Change |
|---|---|---|
| 0.1.0 | 2026-09-24 | 카드 t1143 plan 초안. SPEC-DUAL-HARNESS-RECOVERY-001 0.3.1이 이관한 세 항목(REQ-DHR-015 런타임 조항, AC-DHR-012, AC-DHR-023)을 원문 그대로 이어받고(§B), 설계 경로 (i) — Codex에서 read-only 계약 역할을 `spawn_agent`가 아니라 최상위 `codex exec -s read-only` 프로세스로 띄우고 부모가 반환문으로 판정 파일을 쓴다 — 를 요구사항으로 적었다. 실행 경로(shell 대 MCP)는 M1 측정으로 정한다(REQ-CAR-011). |

## §A 배경과 목적

카드 t1100 M8에서 codex-cli 0.156.1로 측정한 결과, `spawn_agent`로 띄운 하위 에이전트는 부모 세션의 sandbox를 물려받았고 역할 TOML의 `sandbox_mode`는 적용되지 않았다. 역할 TOML의 `developer_instructions`와 `model_reasoning_effort`는 적용되었다. 그래서 `read-only`로 방출된 `plan-auditor`, `sync-auditor`는 부모가 `workspace-write`일 때 쓰기에 성공했다.

근거(primary checkout에 반출된 t1100 증거, 이 plan에서 읽기 전용으로 읽음):

- `.moai/reports/t1100/ac012-evidence.json`: `codex_version` `0.156.1`, `invocations` 14, `aborted` false, 두 쓰기 시도 모두 `denied: false`, `probe_exists: true`.
- `.moai/reports/t1100/ac023-evidence.json`: 두 감사 역할 모두 `write_denied: false`, 반환문 `... write=allowed`. 해시는 일치했지만 감사자가 판정 파일을 직접 썼다.
- `.moai/reports/t1100/m8-sbx/summary.json`(sha256 `ee41c9e8…b47cc9`): run1은 `-s` 없이 config `sandbox_mode = "workspace-write"`인 부모가 `plan-auditor`를 spawn했고, 두 세션의 `sandbox_policy.type`이 모두 `workspace-write`였으며 `probe-plan-auditor.txt`와 `probe-parent.txt`가 생겼다. run2는 부모를 `codex exec -s read-only`로 띄웠고(`run2-argv.txt`), 두 세션이 모두 `read-only`였으며 실행 뒤 루트에는 `.codex`만 남았다. 부모 자신의 탐침 쓰기도 막혔다.

run2가 이 SPEC의 전제다. 최상위 `codex exec -s read-only` 세션은 모델이 낸 shell 쓰기를 막는다. 이 SPEC은 Codex 경로의 read-only 계약 역할을 그 형태로 띄우는 실행기(audit launcher)와, 감사자가 반환한 원문을 부모가 판정 파일로 쓰는 경로를 요구한다. Claude 쪽 정의와 감사 흐름은 바꾸지 않는다.

## §B 이어받은 항목 (SPEC-DUAL-HARNESS-RECOVERY-001 0.3.1에서 이관)

출처: `.moai/specs/SPEC-DUAL-HARNESS-RECOVERY-001/spec.md` 131-133행, 이관 커밋 `de5faa77a`, SPEC 버전 0.3.1. 이관 방식은 SPEC-FACTORY-LANE-WORKTREE-HANDOFF-001 0.5.10·0.5.11(카드 t1082 → t1145) 선례를 따른다. 아래 표시 사이의 줄은 출처와 한 글자도 다르지 않다(AC-CAR-009가 바이트 비교로 판정한다). 이 SPEC이 이어받는 것은 이 요구사항의 **런타임 조항**, 즉 "When a Codex audit role returns, the parent lane orchestrator SHALL write the audit verdict or report file with exactly the returned text." 부분과, 감사 역할이 런타임에 read-only로 실행된다는 전제다. 방출 쪽 계약(역할 계약 `sandbox: read-only`, 방출된 `sandbox_mode = "read-only"`, Codex 전용 반환 지시)은 출처 SPEC의 AC-DHR-013으로 이미 검증되었으며 이 SPEC은 다시 검증하지 않는다.

AC-DHR-012, AC-DHR-023의 원문은 `acceptance.md` §B에 같은 방식으로 옮겼다. AC-DHR-012가 검증하는 REQ-DHR-014는 이관되지 않았고 출처 SPEC에 남는다. 이 SPEC은 AC-DHR-012의 판정식만 이어받는다.

<!-- inherited:begin REQ-DHR-015 -->
### REQ-DHR-015 — Codex audit roles are read-only and return their verdict text

Where `plan-auditor` or `sync-auditor` is emitted as a Codex role, the role's contract SHALL state `sandbox: read-only` and the emitter SHALL emit `sandbox_mode = "read-only"` for it together with a Codex-only instruction that the role returns its complete verdict or report text instead of writing a file. When a Codex audit role returns, the parent lane orchestrator SHALL write the audit verdict or report file with exactly the returned text. This exception to the contract "the auditor writes its own verdict file" SHALL apply only to the Codex path; the Claude agent definitions, their emitted Claude copies, and the Claude audit workflow SHALL remain unchanged.
<!-- inherited:end REQ-DHR-015 -->

이 SPEC 안에서 위 런타임 조항의 "parent lane orchestrator"는 REQ-CAR-004의 audit launcher가 부모를 대신해 수행한다. launcher는 부모 lane 세션이 부르는 도구이며, 판정 파일에 쓰는 바이트는 감사 프로세스가 반환한 원문뿐이다.

## §C 요구사항 (GEARS)

### REQ-CAR-001 — Read-only-contract roles run as top-level read-only processes

Where the Codex harness drives the lane, when the parent lane orchestrator needs the output of a role whose permission contract states `sandbox: read-only`, the audit launcher shall start that role as exactly one top-level `codex exec` process whose sandbox is set to `read-only` by command-line flag, whose approval policy is `never`, and whose working root is the caller's worktree root, and shall not start that role through `spawn_agent`.

### REQ-CAR-002 — Launchable roles are derived from the permission contract

The audit launcher shall derive the set of launchable roles from the emitted Codex permission contract as the roles whose contract sandbox is `read-only`. When a role outside that set, or a role with no emitted role file, is requested, the audit launcher shall exit non-zero with a diagnostic naming the role and shall start no process.

### REQ-CAR-003 — Role instructions and effort reach the top-level process

The audit launcher shall deliver to the top-level process the role's `developer_instructions` text and `model_reasoning_effort` value exactly as they appear in the emitted role file. When the delivered instruction argument would exceed the launcher's existing argument ceiling, the audit launcher shall fail closed before starting any process and shall report the measured byte length and the ceiling.

### REQ-CAR-004 — The parent writes the verdict file with the returned text

When the top-level audit process exits with status zero and its final agent message is non-empty, the audit launcher shall write that message byte-for-byte to the destination the caller named, by a write that leaves either the previous file or the complete new file and never a partial file.

### REQ-CAR-005 — The destination comes from the caller, never from the model

The audit launcher shall take the verdict destination only from its caller's argument and shall resolve it, after symlink resolution, inside the caller's worktree root. The audit launcher shall not take a destination path from the audit process's output. When the resolved destination lies outside the worktree root, the audit launcher shall exit non-zero without writing.

### REQ-CAR-006 — A failed audit writes nothing

When the top-level audit process exits non-zero, exceeds its time bound, or returns an empty final message, the audit launcher shall not create or modify the destination file, shall exit non-zero, and shall name the role and the failure reason in its diagnostic.

### REQ-CAR-007 — Side channels outside the Codex sandbox are closed or declared

The audit launcher shall start the audit process with no MCP server enabled. The launch record shall state that the read-only guarantee covers the model-generated commands and edits governed by the Codex sandbox, and shall list the writers that sandbox does not govern — Codex's own session files under `CODEX_HOME` and project hook commands — as `UNSUPPORTED`. The audit launcher shall not pass any sandbox-bypass or approval-bypass option to the audit process.

### REQ-CAR-008 — The instruction surface names the launcher, not spawn_agent

The deployed `AGENTS.md` `audit-verdict-file` row and the Codex-only addenda of the read-only-contract roles shall instruct the Codex parent to start these roles through the audit launcher and not through `spawn_agent`, and shall state that the verdict or report file is written by the launcher from the returned text.

### REQ-CAR-009 — The Claude side is unchanged and generated files change only by regeneration

The Claude agent definitions (`.claude/agents/moai/*.md` in the project and in the template mirror) and the Claude plan and sync workflows shall not change. The generated Codex role files shall change only through the emitter's regeneration step.

### REQ-CAR-010 — LIVE invocations are budgeted and never inflated into PASS

The LIVE verification of this SPEC shall count one invocation per `codex exec` process and shall not start an invocation that would exceed the ceiling of its evidence item. When a ceiling would be exceeded, the verification shall record `ABORTED`, stop that item's remaining steps, and fail. An unexecuted LIVE item shall be recorded `NOT_RUN`. Neither `ABORTED` nor `NOT_RUN` shall be counted as PASS.

### REQ-CAR-011 — The launch route is the one measured to work

The audit launcher shall be reachable from a running Codex lane session by the route that the M1 measurement recorded as able to start a read-only top-level process that reaches the model. Where the shell route is recorded as unable, the audit launcher shall be exposed through the moai MCP server and shall take the worktree root as an explicit input. The instruction surface shall name only the route that was measured to work.

## §D 추적

| 요구사항 | 인수 기준 |
|---|---|
| § REQ-DHR-015 (이어받음, 런타임 조항) | AC-DHR-023 (이어받음), AC-CAR-003, AC-CAR-010, AC-CAR-009 (원문 보존) |
| § REQ-DHR-014 (출처 SPEC에 남음) | AC-DHR-012 (이어받음, 판정식만) |
| § REQ-CAR-001 | AC-CAR-001, AC-CAR-010, AC-CAR-011, AC-DHR-012 |
| § REQ-CAR-002 | AC-CAR-002 |
| § REQ-CAR-003 | AC-CAR-001, AC-CAR-006, AC-CAR-010 |
| § REQ-CAR-004 | AC-CAR-003, AC-DHR-023 |
| § REQ-CAR-005 | AC-CAR-005 |
| § REQ-CAR-006 | AC-CAR-004 |
| § REQ-CAR-007 | AC-CAR-001, AC-CAR-010 |
| § REQ-CAR-008 | AC-CAR-007 |
| § REQ-CAR-009 | AC-CAR-008 |
| § REQ-CAR-010 | AC-CAR-010, AC-CAR-011, AC-DHR-012 |
| § REQ-CAR-011 | AC-CAR-007, AC-CAR-010 |

## §E 범위 밖

### Out of Scope — `spawn_agent` 경로의 sandbox 강제

- Codex가 하위 에이전트에 역할 TOML의 `sandbox_mode`를 적용하게 만드는 일. 이것은 Codex 호스트의 동작이며 이 SPEC은 우회 경로만 만든다.
- 감사 역할 TOML을 방출 목록에서 빼는 일. AC-DHR-012의 판정식이 12개 역할 파일 집합을 요구하므로 역할 파일은 그대로 방출한다.
- `spawn_agent`로 감사 역할을 띄우는 것을 기계적으로 막는 가드. 이 SPEC은 지시면(REQ-CAR-008)만 바꾼다.

### Out of Scope — Claude 쪽 감사 흐름

- Claude의 `plan-auditor`, `sync-auditor` 정의와 "감사자가 자기 판정 파일을 쓴다" 계약. Claude 쪽은 바꾸지 않는다(REQ-CAR-009).

### Out of Scope — sandbox 밖 쓰기 주체의 봉쇄

- Codex가 `CODEX_HOME` 아래에 쓰는 세션 기록과 프로젝트 hook 명령의 쓰기를 막는 일. 이 SPEC은 그것들을 `UNSUPPORTED`로 선언만 한다(REQ-CAR-007).
- `write-path-scope`(경로 단위 쓰기 제한). Codex sandbox는 이것을 표현하지 못하며, 이 SPEC은 감사자를 read-only로 두고 쓰기를 부모에게 넘기는 방식으로 우회한다.

### Out of Scope — REQ-DHR-013·014의 나머지

- REQ-DHR-013의 `sandbox` 축 매핑(`UNSUPPORTED`/`measured`, 커밋 `17bfceaa8`)을 되돌리는 일. 하위 에이전트 경로의 사실은 그대로다.
- REQ-DHR-014 본문. AC-DHR-012 판정식만 이어받는다.

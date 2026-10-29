---
id: SPEC-AGENTIC-FACTORY-001
title: "AgenticFactory 실행 계약 — GitHub Issue 원장 기반 소유권·전이·안전 게이트"
version: "0.2.0"
status: draft
created: 2026-10-29
updated: 2026-10-29
author: GOOS
priority: P0
phase: "v3.3.0 target"
module: "internal/agenticfactory"
lifecycle: spec-anchored
tags: "agentic-factory, github-issues, ledger, ownership, lease, policy, contract"
tier: L
issue_number: 1948
related_specs: [SPEC-ASIDE-BROWSER-001, SPEC-RESOURCE-SLOT-LEASE-001, SPEC-FACTORY-LEASE-RENEW-001]
---

# SPEC-AGENTIC-FACTORY-001 — AgenticFactory 실행 계약

## HISTORY

| 버전 | 날짜 | 작성 | 내용 |
|---|---|---|---|
| 0.1.0 | 2026-10-29 | manager-spec | 최초 작성 (draft). 규범 원천은 GitHub Issue #1948 하나뿐이다. 조정자 추가 요구(단일 Issue 시나리오 7종·6개 결정 사항, 세션 구조·명칭·CLI 재정의, 명칭 미확정 표시·lane-22 대응표 정렬), 입력 보고서 세 건(lane-22 갭 매트릭스·이관 대응표, lane-21 브라우저 계약), 상위 설계 권고(Agentic Factory / ControlService / Coordinator / Worker, `moai factory start·status·attach`, TTL 대 epoch 교차 점검)를 같은 판에 PROPOSAL로 반영했다. |
| 0.2.0 | 2026-10-29 | manager-spec | plan-audit iteration-1 FAIL 0.69(`.moai/reports/SPEC-AGENTIC-FACTORY-001/plan-audit-iter1.md`)의 결함 D1~D18을 앞으로 고치는 판. **REQ ID 재사용(상한 25 유지)**: 옛 REQ-AF-002(차단·보류)는 REQ-AF-001에 합치고 REQ-AF-002는 `done` 수용 조건(원격 반영 + 로컬 동기화 + 배포 증거 또는 기록된 N/A)으로, 옛 REQ-AF-005(비밀 가림)는 REQ-AF-004에 합치고 REQ-AF-005는 관찰 모니터 표시 항목으로 다시 썼다. REQ-AF-015는 환경별 단계 실제 실행, REQ-AF-016은 3선택 선택 모델(선택·저장·다시 읽기), REQ-AF-014는 거울 쓰기 실패 규칙 하나로 정리(D2), REQ-AF-010은 활성 판정 정본을 `homestate.ClassifyOwnerWith`로 정함(D12, §F.2), REQ-AF-025는 단계별 호환 기준(D10). AC는 25개 그대로이며 시나리오 하네스 계약(D5), 증거 원장의 "대상 부재 참고 관측" 전환(D1), AC-AF-019·024 기준점(D1·D9), AC-AF-018 여섯 증거 제거 사례(D6), AC-AF-002 미전달 dispatch 사례(D7), AC-AF-020 운영 기록 점검(D8), 마일스톤 재배치(D11)를 반영했다. `factory.go:284-286`·`approve` `Use:` 문자열 사실 정정(D17), progress §E.1 기록(D18). 명칭·명령은 계속 미확정이다. |

## §A 범위와 규범 원천

### A.1 규범 원천

이 SPEC의 유일한 규범 원천은 GitHub Issue #1948 "[P0] AgenticFactory 우선 완성: GitHub Issues 기반 실행·관찰·안전 게이트 및 단계적 카드 이관"이다(`updatedAt` 2026-10-29T03:12:53Z, 상태 OPEN, 댓글 0건 — `gh issue view 1948` 결과를 저장한 사본으로 확인). 그 밖의 설계 문서는 없다. 조정자가 넘긴 로컬 보고서 세 건(커밋되지 않음)과 상위 설계 권고는 규범 원천이 아니라 설계 입력이며, 살아남아야 할 내용은 이 SPEC 본문에 옮겨 적었다. 이 SPEC은 Issue 본문을 계약·정책·설정 초안으로 옮긴 것이며, Issue에 없는 내용을 지어내지 않는다. Issue 본문이 직접 정하지 않고 이 SPEC이 정한 결정은 §F에 "결정"으로 따로 표시했고, 확인할 수 없는 사항은 §G 미결 질문과 §H 확인 공백으로 넘겼다.

### A.2 이 SPEC이 다루는 것 (계약 범위)

이 SPEC은 **계약 + 정책 + 설정 초안**이다. 코드는 쓰지 않는다. 다루는 것은 다음과 같다.

1. 업무 상태 모델과 전이표(누가 어떤 전이를 쓸 수 있는가)
2. 명령 전달의 네 단계 상태(전송/수신/실행/완료)와 각각의 증거 원천
3. GitHub Issue에 싣는 원장 필드와 필드별 쓰기 권한
4. 소유권·단일 writer·중복 dispatch 방지·중단 복구 계약
5. 공급자·로컬 환경·브라우저 정책
6. 새 설정 키 초안(기존 설정 파일은 고치지 않음)
7. 병합·완료 수용 규칙
8. 기존 결함 다섯 건(#1938, #1926, #1924, #1847, #1813)의 대체 동작과 재현 시험 연결
9. 세션 구조·역할 명칭·명령 체계의 **미확정(provisional)** 초안과 호환·단계 폐기 계약

이 문서에 나오는 역할 이름(`ControlService`, `Coordinator`, `Worker`), 실행 단위 이름(`run`, `attempt`, `slot`), 명령 이름(`moai factory start` 등)은 모두 **미확정(provisional)**이다. 요구사항은 이름이 아니라 역할의 권한 경계와 명령의 동작 계약을 묶으며, 이름이 확정되면 문서의 이름만 바뀐다(OQ-2).

### A.3 "구현 완료"의 정의 경계

Issue #1948은 "상위 착수 원장이지 구현 완료나 실행 시작의 증명이 아니다"라고 적는다. 이 SPEC도 같다. 이 SPEC이 PASS를 받아도 AgenticFactory가 구현된 것이 아니며, Phase 2(남은 카드의 Issue 이관)를 수행한 것도 아니다. 이관 자체는 어떤 경우에도 구현 완료로 세지 않는다(REQ-AF-002, §I).

## §B 배경 — 이 트리에서 읽은 사실

아래는 모두 이 트리(HEAD `5c3f0ebf27584d2c45deb2d00f9d69a4eea932c4`)에서 직접 읽거나 실행한 결과다. 상세 근거와 명령은 `research.md`에 있다.

- 카드 큐는 SQLite다: `internal/factory/backlog_sqlite.go`가 `modernc.org/sqlite`를 쓴다.
- 자원 슬롯 임대는 `internal/factory/slot_lease.go`(SPEC-RESOURCE-SLOT-LEASE-001)다. 자원 이름 단위의 JSON 기록이며, 선언 기한이 지난 임대는 **소유 프로세스가 살아 있어도 `--force` 없이 인계된다**(파일 머리 주석, REQ-RSL-006).
- 카드 임대 기간은 `internal/homestate/factory_lease.go`가 `workflow.factory_lease.duration` / `renew_interval`로 읽고, 범위를 벗어난 값은 팩토리 개방을 거절하며 값을 깎아 맞추지 않는다.
- `internal/factory/factory_card.go`는 **없다**. 카드 명령은 `internal/cli/factory_card.go`에 있다(`next`, `stage`, `renew`, `complete`, `assign`, `status`, `decide` 등).
- `internal/github`에는 Issue 파싱(`issue_parser.go`)·종료(`issue_closer.go`)·SPEC 연결(`spec_linker.go`)·PR 병합/리뷰 도우미만 있다. **Issue를 원장으로 쓰는 코드는 없다.**
- `internal/agenticfactory/`는 없다(`go test ./internal/agenticfactory/...` → `lstat ... no such file or directory`, exit 1).
- 이 프로젝트의 로컬 설정 `.moai/config/sections/workflow.yaml`에는 `audit.model: codex`, `audit.gates.codex: required`, `codex.review_gate.enabled: true`, `codex.task.allow_write: true`가 켜져 있다. 반면 코드 기본값(`internal/config/defaults.go`)과 배포 템플릿은 `review_gate.enabled: false`, `allow_write: false`다. 즉 **이 저장소에서 Codex는 로컬 설정으로 이미 켜져 있다.**
- `.moai/config/sections/llm.yaml`에는 `team_mode: glm`, `glm` 블록, `defaults.sessions.codex`(leader/lane 모델)가 있다.
- SPEC-ASIDE-BROWSER-001(completed) REQ-ASB-009는 Aside를 "자동 감지하지 않고, Phase 0.5 선택지에 넣지 않고, 권하지도 않는다"고 정한다. 이것이 이 SPEC이 말하는 "제안 금지 조항"이다.
- 이 트리 어디에도 `egolite` / `ego-lite` / `ego-browser`라는 문자열이 없다.
- `main` 보호 설정(읽기 전용 `gh api` 조회): 필수 검사 `Lint`, `Build (linux/amd64)`, `Analyze (Go) (go)`, `enforce_admins: true`, `allow_force_pushes: false`, `allow_deletions: false`, **`required_linear_history: false`**. 저장소는 squash와 merge commit 병합을 모두 허용하고 rebase 병합은 막는다. `origin/main`에는 merge commit이 있다(`fc17f49b10 Merge pull request #1862`).

아래 활성 판정 사실은 0.2.0에서 트리 `e14b5bbac2eb2499cbaa3ae9434595001c91cb39`(0.1.0 SPEC 커밋, 코드는 `5c3f0ebf27`과 같음)에서 직접 읽었다.

- 프로세스 활성 판정 함수가 둘 있다. ① `internal/factory/factory_alive_unix.go:22-28` `FactoryProcessAlive(pid)` — `pid <= 0`이면 `false`, 아니면 `kill(pid, 0)`이 성공하거나 `EPERM`이면 `true`. PID만 보고 프로세스 시작 시각은 보지 않는다. Windows 짝은 `factory_alive_windows.go:28`. 호출처는 `internal/factory/slot_lease.go:183`(`Stale()`), `internal/factory/integration_lock.go:232`, `internal/cli/factory_card.go:401`(`factoryWorkerPidAlive`), `internal/cli/gate_lock.go:242`, `internal/cli/integration_wait.go:286`, `internal/cli/codex_review_gate.go:363`. ② `internal/homestate/factory_run_retire.go:152-168` `ClassifyOwnerWith(probe, pid, processStart)` — `live` / `dead` / `indeterminate` 셋을 돌려준다. `pid < 1`이거나 기록된 시작 시각이 비어 있으면 `indeterminate`, 프로세스가 살아 있고 시작 시각이 같으면 `live`, 살아 있지만 시작 시각이 다르면(PID 재사용) `dead`, 프로세스가 없으면 `dead`, 나머지는 `indeterminate`다. 기본 구현 `DefaultOwnerClassifier`(:139-141)는 `ProbeProcessIdentity`(`internal/homestate/profile_lease.go:210-220`)를 쓰고, 이것은 PID 상태를 먼저 본 뒤 플랫폼별 시작 시각(Linux 등 `ps -o lstart=`, macOS `sysctl kern.proc.pid`, Windows `GetProcessTimes`)을 읽으며, 시작 시각을 못 읽으면 `indeterminate`다.
- `moai cc -l` 도움말은 합류 시 리더를 "pid + process-start"로 검증한다고 적는다(`internal/cli/cc.go:69-70`). 즉 pid + 시작 시각 방식을 구현한 함수는 `homestate` 쪽에 있고(그 도움말의 검증이 어느 함수를 부르는지는 읽지 않음), `FactoryProcessAlive`는 PID만 본다. 0.1.0 §F.2가 둘을 같은 방식으로 적은 것은 틀렸다(D12, 0.2.0에서 고침).

## §C 요구사항 (GEARS)

요구사항은 25개로 Tier L 상한(25)과 같다. 패턴 열은 GEARS 형식을 가리키며, 두 형식 이상을 한 요구사항에 묶은 경우는 `Compound`로 적었다. 0.2.0에서 REQ-AF-002와 REQ-AF-005는 ID를 재사용했다(HISTORY 참조). 옛 내용은 각각 REQ-AF-001과 REQ-AF-004에 그대로 들어 있다.

### C.1 상태 모델과 전달 상태

| ID | 패턴 | 요구사항 |
|---|---|---|
| REQ-AF-001 | Compound (Event-driven + State-driven) | When an actor requests an Issue state transition, the state engine SHALL apply it only if the edge appears in the transition table of `design.md` §2.2, its guard evidence is present, and the requesting role is the edge's listed writer; the engine SHALL refuse every other request, including any request that skips a stage, and SHALL leave the prior state unchanged. While an Issue is in `blocked` or `on-hold`, the state engine SHALL record exactly one reason from the `block_reason` set of `design.md` §2.1a together with its trigger evidence and normal resume condition, SHALL NOT report a `blocked` Issue as successful, SHALL keep the stage the Issue left (`return_stage`), and SHALL return the Issue only to that stage. |
| REQ-AF-002 | Event-driven | When the state engine is asked to record `done`, it SHALL record it only if all three completion evidences are present for the merged HEAD: (a) the merge is reflected on the remote `main`, confirmed by a direct remote query; (b) a local-sync record shows the local integration surface was brought to that remote state without discarding any uncommitted change (the dirty path list before and after is recorded and the before-list is a subset of the after-list); (c) a deploy record for that HEAD, or a recorded `deploy: N/A` with its reason where the project defines no deploy step. Missing any of the three, the engine SHALL keep the Issue in `landing` or move it to `blocked` with the reason, and SHALL NOT close the Issue. Migration of remaining cards (Phase 2) SHALL NOT be recorded as `done` or as implementation completion. |
| REQ-AF-003 | Compound (Ubiquitous + Event-driven) | The delivery tracker SHALL record `sent`, `received`, `executed`, and `completed` as four separate states per command, each set only from its own evidence source in `design.md` §3; a successful send result or an idle notice SHALL NOT set `received`, `executed`, or `completed`. When an acknowledgement is missing, the tracker SHALL re-query the recipient state before any re-send and SHALL NOT re-send blindly. |

### C.2 원장

| ID | 패턴 | 요구사항 |
|---|---|---|
| REQ-AF-004 | Compound (Ubiquitous + Unwanted) | The Issue ledger SHALL carry the fields of `design.md` §4 (task name, acceptance criteria, priority, dependencies, owner, execution state, HEAD, check links, review / audit / handoff evidence, next action) with their declared types; the ledger writer SHALL refuse a write to a field by a role that is not that field's declared owner, and observation reads SHALL NOT require or confer execution-control authority. The ledger, the delivery records, and every log written by AgenticFactory SHALL NOT contain a secret (token, API key, credential, cookie); when a value matching a secret pattern reaches a write path, the writer SHALL redact it and record the redaction. |
| REQ-AF-005 | Ubiquitous | The observation monitor SHALL show, for each project, each Issue, and each owner, the items of `design.md` §8.1 — execution state, last-checked time, HEAD SHA, required-check results, progress (current stage and attempt), block reason with resume condition, and next action — each item read from its authoritative source (claim store, Issue ledger, or direct remote query) and marked with the time it was read; the monitor SHALL separate observation from control so that its read path holds no write capability and its output carries no control action, and running it SHALL leave the store sequence number and every tree key unchanged. |

### C.3 소유권과 복구

| ID | 패턴 | 요구사항 |
|---|---|---|
| REQ-AF-006 | Ubiquitous | The claim store SHALL admit exactly one accountable owner and exactly one writer per Issue: of any set of concurrent claims for the same Issue exactly one SHALL succeed, and every other claimant's write SHALL be refused. |
| REQ-AF-007 | Event-driven | When a dispatch names an Issue that already has a live attempt or an undelivered dispatch, the dispatcher SHALL refuse the new dispatch and SHALL record the existing one it collided with. |
| REQ-AF-008 | Event-driven | When ownership verification fails or cannot be completed (claim absent, holder mismatch, store unreadable), the writer SHALL refuse the write and SHALL NOT fall back to an unverified write. |
| REQ-AF-009 | Event-driven | When a session disconnects, a lease expires, a run's ControlService is gone, a stale HEAD result arrives, or a step fails partway, the recovery path SHALL first query the actual execution state (process liveness, worktree HEAD and dirty state, delivery state, remote state) and SHALL then resume, re-issue, or block; it SHALL NOT force-kill a process and SHALL NOT start a duplicate attempt. |
| REQ-AF-010 | Compound (State-driven + Ubiquitous) | While an attempt's lease is expired and the liveness classification of its holder (`§F.2`: pid plus the process-start fingerprint recorded at claim time) is `live` or `indeterminate`, the claim store SHALL refuse a takeover and SHALL mark the attempt `lapsed`; a new attempt SHALL be issued only when the classification is `dead` (including a live PID whose start fingerprint differs from the recorded one) or the holder records a safe stop, and the prior attempt's worktree, branch, SHA, and dirty state SHALL be preserved and referenced by the new attempt. Every attempt SHALL carry an assignment epoch, a takeover SHALL increment it, and the claim store SHALL refuse any write that carries an epoch older than the current one. |
| REQ-AF-011 | Compound (Ubiquitous + Event-driven) | The local claim store SHALL be the authoritative store for effect state (claim, lease, attempt, delivery); the Issue ledger comment SHALL be its mirror carrying the store sequence number, and when the two disagree the reconciler SHALL rewrite the mirror from the store. When the store is missing or unreadable, the reconciler SHALL block and SHALL NOT rebuild a claim from the mirror. Human-owned ledger fields SHALL be authoritative on the Issue. |

### C.4 정책

| ID | 패턴 | 요구사항 |
|---|---|---|
| REQ-AF-012 | Unwanted | AgenticFactory implementation work and local-environment verification SHALL NOT invoke a GLM-managed backend. |
| REQ-AF-013 | Capability gate | Where the operator has not explicitly set the Codex opt-in key (`design.md` §6), AgenticFactory SHALL make zero Codex calls — no automatic call, no default enablement, no fallback call. |
| REQ-AF-014 | Event-driven | When a provider or tool returns an authentication failure, a permission denial, or HTTP 429 to an operation, the caller SHALL NOT bypass it and SHALL NOT retry it in a loop; it SHALL preserve the current state, record a block (`auth-failed`, `permission-denied`, or `rate-limited`) with its normal resume condition on the operation that failed, and SHALL NOT acquire a new key, paid capacity, or additional spend without separate operator approval. Where the failed operation is only the Issue-ledger mirror write, the block SHALL be recorded on the mirror (`mirror_state: blocked(<reason>)`, `design.md` §4) and SHALL NOT change the Issue's execution state, because the claim store is authoritative (REQ-AF-011). |
| REQ-AF-015 | Compound (Ubiquitous + Event-driven) | For each local environment in the matrix of `design.md` §5.1, the environment matrix SHALL record the result of actually executing each step — install check, start, receive a task, run a test, safe hold and resume, collect the result — with the command, its exit code, and an output excerpt per step, and SHALL record the environment as `PASS` only when every step was executed and succeeded. When an environment is not installed, is denied, is unconfigured, or lacks capacity, the matrix SHALL record it as `BLOCKED(<cause>)` with the step at which it stopped and the observed evidence, and SHALL NOT record a step that was not executed as a success. |
| REQ-AF-016 | Compound (State-driven + Event-driven) | The browser selection SHALL offer exactly three choices (`egolite`, `aside`, `chrome`); when the operator selects one, the selection SHALL be saved as `workflow.agentic_factory.browser.selected` and SHALL read back unchanged after the configuration is re-read, and no code path SHALL pick a browser on the operator's behalf. While no browser is selected, the browser adapter SHALL use `chrome` without saving `chrome` as a selection; when a browser is explicitly selected, the adapter SHALL invoke only that browser, SHALL preserve the existing selection, and, when the selected browser is not installed or its connection fails, SHALL block with a resume condition and SHALL NOT fall back to another browser. |
| REQ-AF-017 | Event-driven | When a value in the `workflow.agentic_factory.*` draft section is out of range or of the wrong type, the configuration loader SHALL refuse it and name the key, the allowed range, and the file; it SHALL NOT clamp the value. |

### C.5 병합·완료

| ID | 패턴 | 요구사항 |
|---|---|---|
| REQ-AF-018 | Ubiquitous | The merge gate SHALL admit a merge only when, for the latest HEAD, every required check actually ran and passed, a same-HEAD code review and security review are present, the required independent audit is a debt-free PASS, the official QA skill (§F.5) verdict for that HEAD is PASS, and both the writer's and the coordinator's ready-handoffs are present; a skipped, missing, failed, cancelled, stale, or PASS-WITH-DEBT result, or an exhausted audit budget, SHALL be recorded as `blocked` and SHALL NOT count as a pass. |
| REQ-AF-019 | Unwanted | AgenticFactory SHALL NOT delete a required check, weaken branch protection, change the repository's current linear-history setting, force-push `main`, or merge through an administrator bypass. |
| REQ-AF-020 | Ubiquitous | The evidence store SHALL keep SHA-bound evidence (audit, review, QA verdicts) outside the code candidate's commit history so that recording evidence does not change the code HEAD it describes, and every stored evidence item SHALL name the full HEAD SHA it judged and be re-verified against the candidate HEAD before the merge gate reads it. |
| REQ-AF-021 | Event-driven | When a merge outcome is unknown, the merge step SHALL query the remote state first and SHALL NOT re-merge; when the remote reflection of the merge on `main` is confirmed, the state engine SHALL move the Issue to `done` only under the completion evidence of REQ-AF-002, SHALL close the Issue only after `done` is recorded, and SHALL admit the next dispatch for the freed slot only after the Issue is closed. |
| REQ-AF-022 | Ubiquitous | The Codex default migration SHALL proceed in the order of `plan.md` §F.3 — introduce the explicit opt-in gate, verify zero calls when unset, then change defaults — and no existing Codex feature SHALL be removed before its replacement is verified; inspection, independent audit, and handoff obligations SHALL remain in force at every stage. |

### C.6 세션 구조·명령·호환 (명칭은 모두 미확정)

| ID | 패턴 | 요구사항 |
|---|---|---|
| REQ-AF-023 | Ubiquitous | The deterministic control role (provisional name `ControlService`) SHALL be the only role that writes claim, lease, liveness, reconciliation, and resume-decision state; the LLM coordination role (provisional `Coordinator`) SHALL NOT write that state and SHALL act on it only through the ControlService's commands; a Worker session SHALL write only within its own attempt. |
| REQ-AF-024 | Event-driven | When an operator or coordinator reconnects, lists, attaches to, resumes, holds, or stops a run, the command set of `design.md` §8 SHALL query the actual execution first, SHALL require an explicit run or attempt id when more than one candidate matches, SHALL NOT start a duplicate attempt, SHALL NOT force-kill, SHALL NOT widen the caller's permissions on resume, and an observe-only command SHALL NOT mutate any state. |
| REQ-AF-025 | Ubiquitous | The existing entry commands (`moai cc -f`, `moai cc -l`, and their `glm` / `codex` counterparts as they exist) SHALL remain as compatibility aliases through the deprecation stages of `design.md` §9.4 with this per-stage behaviour: at S1 their help text, argument parsing, refusal messages, stdout, stderr, and exit codes SHALL be byte-identical to the baseline golden files captured before S1 (`design.md` §9.6); at S2 and S3 they SHALL stay byte-identical to that baseline except for exactly one added stderr deprecation line naming the replacement; at S4 a command SHALL be removed only when the S4 conditions of §9.4 hold for it. At every stage, running sessions SHALL NOT be restarted by the migration, and existing session IDs, permissions, lane labels (addresses), owners, and handles SHALL stay byte-identical; the compatibility table SHALL stay aligned with lane-22's migration correspondence table (S1–S4, rows C1–C11). |

## §D 수용 기준 요약

수용 기준은 `acceptance.md`에 25개(AC-AF-001 ~ AC-AF-025)로 있다. 시나리오 일곱 개(AC-AF-001 ~ AC-AF-007)는 **하나의 실제 Issue에서 처음부터 끝까지** 하나의 순서 있는 시험 묶음(`TestScenarioSuite`)으로 측정한다(acceptance.md §A.1 하네스 계약). 이 측정은 실제 GitHub Issue에 쓰는 외부 공유 동작이므로 운영자 승인 대상이며, 어떤 Issue를 쓸지는 미결 질문 OQ-1이다. 보호된 `main`으로의 실제 병합과 Issue 종료(시나리오 6·7)는 운영자 keep-set 승인 대상이며, 승인이 없으면 그 두 시나리오는 `BLOCKED(operator-ack-missing)`로 기록되고 PASS가 아니다.

모든 AC는 run 단계가 각 AC의 이름 붙은 시험에서 `EXPECTED_RED`를 기록하기 전까지 **채택 전**이며 release-blocking이 아니다(acceptance.md §B). 예외적으로 AC-AF-019는 regression-guard로 분류한다.

## §E 추적표 (REQ → AC)

| REQ | AC |
|---|---|
| REQ-AF-001 | AC-AF-008, AC-AF-025 |
| REQ-AF-002 | AC-AF-007 |
| REQ-AF-003 | AC-AF-003 |
| REQ-AF-004 | AC-AF-012 |
| REQ-AF-005 | AC-AF-022 |
| REQ-AF-006 | AC-AF-001, AC-AF-025 |
| REQ-AF-007 | AC-AF-002 |
| REQ-AF-008 | AC-AF-009 |
| REQ-AF-009 | AC-AF-004, AC-AF-025 |
| REQ-AF-010 | AC-AF-004, AC-AF-010 |
| REQ-AF-011 | AC-AF-011 |
| REQ-AF-012 | AC-AF-014 |
| REQ-AF-013 | AC-AF-013 |
| REQ-AF-014 | AC-AF-015 |
| REQ-AF-015 | AC-AF-014 |
| REQ-AF-016 | AC-AF-016 |
| REQ-AF-017 | AC-AF-017 |
| REQ-AF-018 | AC-AF-005, AC-AF-018 |
| REQ-AF-019 | AC-AF-019 |
| REQ-AF-020 | AC-AF-018 |
| REQ-AF-021 | AC-AF-006, AC-AF-007 |
| REQ-AF-022 | AC-AF-020 |
| REQ-AF-023 | AC-AF-021, AC-AF-025 |
| REQ-AF-024 | AC-AF-022, AC-AF-023 |
| REQ-AF-025 | AC-AF-024 |

## §F 이 SPEC이 내린 결정

### F.1 claim은 slot lease인가, 별도 기록인가 — 결정: 별도 기록

- **결정**: Issue claim은 `internal/factory/slot_lease.go`의 slot lease가 **아니다**. Issue 하나당 하나뿐인 별도 claim 기록을 새로 둔다(제안 경로 `internal/agenticfactory/claim/`). 상호 배제에 쓰는 잠금 기반(`acquireStateLockImpl` 계열: Unix flock, Windows atomic-create)만 재사용 후보로 둔다.
- **근거(코드에서 읽음)**: ① slot lease는 선언 기한이 지나면 소유 프로세스가 살아 있어도 `--force` 없이 인계한다(REQ-RSL-006). slot lease의 설계 근거는 "잘못된 '비어 있음'의 비용이 무거운 실행 두 개가 겹치는 것뿐"이라는 것인데, Issue의 writer가 겹치면 같은 브랜치에 writer가 둘 생긴다. 비용 전제가 다르다. ② slot lease 키는 자원 이름(`^[a-z0-9-]{1,64}$`)이고 Issue·run·attempt 식별자가 없다. ③ 현행 카드 임대는 factory.db의 run별 행이라서, #1938에서처럼 한 카드가 run 셋에 걸쳐 행 셋을 갖는 일이 생긴다.
- **#1938에 대한 귀결**: claim 기록은 run과 무관하게 Issue 하나에 행 하나다. 죽은 조정 세션이 남긴 만료 attempt는 같은 claim 아래 새 attempt로 재발급되며, 옛 attempt는 `lapsed`/`superseded`로 닫힌다. 행이 늘어나는 우회 경로(`assign`을 다른 run에 넣기)는 필요 없어진다.
- **#1813에 대한 귀결**: claim 기록은 작업 트리의 정규화된 절대 경로와 브랜치를 attempt 속성으로 갖는다. 진입 대상은 이름이 아니라 claim이 가리키는 트리로 정해지므로, 같은 이름의 다른 트리가 조용히 생기는 일이 claim 검사에서 거절된다.

### F.2 "기한은 지났지만 소유자는 살아 있음" — 결정: 인계 금지, `lapsed` 표시

**활성 판정 정본(결정)**: claim·인계 규칙의 활성 판정은 `homestate.ClassifyOwnerWith`(`internal/homestate/factory_run_retire.go:152-168`)를 정본으로 삼는다. 기본 구현은 `DefaultOwnerClassifier`(:139-141) → `ProbeProcessIdentity`(`internal/homestate/profile_lease.go:210-220`)다.

| 비교 항목 | `homestate.ClassifyOwnerWith` | `factory.FactoryProcessAlive` |
|---|---|---|
| 위치 | `internal/homestate/factory_run_retire.go:152-168` | `internal/factory/factory_alive_unix.go:22-28`, `factory_alive_windows.go:28` |
| 입력 | pid + claim 때 기록한 프로세스 시작 시각(fingerprint) | pid만 |
| 결과 | `live` / `dead` / `indeterminate` 셋 | `true` / `false` 둘 |
| PID 재사용 | 시작 시각이 다르면 `dead`로 구분한다 | 구분하지 못한다(재사용된 PID를 살아 있다고 봄) |
| 알 수 없음 | `pid < 1`, 시작 시각 미기록, 시작 시각 읽기 실패 → `indeterminate` | `pid <= 0` → `false`(죽었다고 봄). Unix에서 `EPERM`은 `true` |
| 플랫폼 | 시작 시각: Linux 등 `ps -o lstart=`, macOS `sysctl kern.proc.pid`, Windows `GetProcessTimes` | Unix `kill(pid,0)`, Windows `OpenProcess` |
| 현행 용도 | factory run 퇴역 판단(죽음이 확정될 때만 퇴역, 함수 주석). 같은 pid + 시작 시각 방식이 `cc -l` 합류 시 리더 검증 도움말에도 적혀 있다(`internal/cli/cc.go:69-70`). 그 검증이 이 함수를 부르는지는 확인하지 않았다 | 자원 slot lease `Stale()`(`slot_lease.go:183`), 통합 잠금(`integration_lock.go:232`), 카드 임대의 worker 생존 확인(`factory_card.go:401`), `gate_lock.go:242`, `integration_wait.go:286`, `codex_review_gate.go:363` |

- 참고: 자원 slot lease의 `Stale()`은 `FactoryProcessAlive`를 부르기 전에 `pid <= 0`을 따로 "살아 있음"으로 막는다(`internal/factory/slot_lease.go:175-183`). 즉 보수적 해석은 호출하는 쪽이 덧붙인 것이지 `FactoryProcessAlive`의 성질이 아니다.
- **근거**: claim 인계에서 틀린 "죽었음" 판정은 같은 Issue에 writer 둘을 만든다. `ClassifyOwnerWith`는 PID 재사용을 잡아내고, 모르는 경우를 `indeterminate`로 따로 돌려준다. `FactoryProcessAlive`는 두 가지 모두 하지 못한다. 특히 `pid <= 0`을 "죽음"으로 읽으므로 claim 인계에 쓰면 PID를 기록하지 못한 소유자를 바로 인계 대상으로 만든다.
- **인계 규칙(갱신)**: `dead`일 때만 인계할 수 있다. `live`와 `indeterminate`는 둘 다 인계를 막고 attempt를 `lapsed`로 표시한다. `indeterminate`에서 벗어나는 길은 소유자의 안전 정지 기록이나 운영자 결정 증거뿐이다.
- **조건(이 결정의 전제)**: claim 때 ControlService가 소유 프로세스의 pid와 시작 시각을 함께 기록해야 한다. 기록하지 못한 attempt는 언제나 `indeterminate`가 되어 자동 인계되지 않는다.
- **`FactoryProcessAlive`의 자리**: 기존 호출처(자원 slot lease, 통합 잠금, 카드 임대 등)는 그대로 둔다. 이 SPEC은 그 함수를 claim 판정에 쓰지 않을 뿐 고치지 않는다(§I).
- **확인하지 못한 것(OQ-17)**: Linux의 `ps -o lstart=`는 초 단위라 같은 초 안의 PID 재사용은 구분하지 못할 수 있다(macOS는 마이크로초 단위). Windows 시작 시각 비교의 세부 의미는 끝까지 읽지 않았다. 이 둘이 확인되기 전까지 정본 결정은 위 조건 아래에서만 유효하며, 같은 초 재사용 위험은 epoch fencing(REQ-AF-010)으로 늦은 쓰기를 막는 것으로 보완한다.

규칙:

- 살아 있거나(`live`) 판정할 수 없는(`indeterminate`) 소유자의 attempt는 기한이 지나도 인계하지 않는다. attempt를 `lapsed`로 표시하고 조정자에게 알린다.
- 인계할 수 있는 경우는 둘뿐이다. (a) 활성 판정이 `dead`일 때(시작 시각이 다른 PID 재사용 포함), (b) 소유자가 안전 정지(현재 SHA, dirty 상태, 실행 중이던 검사, 재개 지점)를 기록했을 때. 인계를 실행하는 주체는 결정적 제어 역할(ControlService)뿐이며, 조정자(LLM)는 요청만 한다.
- 보존 대상: 옛 attempt의 작업 트리, 브랜치, HEAD SHA, dirty 경로 목록, 마지막 delivery 상태. 새 attempt는 이 기록을 참조로 갖고, 옛 트리를 삭제하거나 초기화하지 않는다. 같은 run에서 살아 있는 attempt는 언제나 하나다.

### F.3 효과 상태의 저장 위치 — 결정: 로컬 claim 저장소가 권위, Issue는 거울

- **권위 저장소**: claim·lease·attempt·delivery 같은 효과 상태는 로컬 claim 저장소가 권위다. GitHub 댓글로는 원자적 compare-and-set을 할 수 없어 단일 writer를 보장할 수 없기 때문이다.
- **거울**: Issue에는 고정된 원장 댓글 하나를 두고, 저장소 순번(sequence)과 함께 효과 상태를 비춘다. 관찰자는 이것만 읽어도 된다.
- **사람 소유 필드**(업무명, 수용 기준, 우선순위, 의존성)는 Issue 쪽이 권위다. 저장소는 다시 읽어 올 뿐 덮어쓰지 않는다.
- **원격 사실**(병합 여부, 원격 `main` 반영, 검사 결과)은 어느 저장소에도 권위가 없고, 매번 GitHub에 직접 물어본다.
- **불일치 조정**: 효과 상태가 다르면 저장소를 따라 거울을 다시 쓴다. 저장소가 없거나 읽을 수 없으면 `blocked`로 두고, 거울에서 claim을 재구성하지 않는다(거울에서 소유권을 만들어 내면 단일 writer가 깨진다).

### F.4 Codex 명시 설정 키 — 결정(키 이름은 미확정): `workflow.agentic_factory.codex.enabled`

- 형식: 불리언, **기본값은 미설정**(키 없음). 미설정과 `false`는 모두 "꺼짐"이며 호출은 0건이다. `true`만 켜짐이다. 불리언이 아닌 값은 거절한다(깎아 맞추지 않음).
- 범위: AgenticFactory 경로만 이 키를 읽는다. 기존 `workflow.codex.*`, `workflow.audit.model: codex`, `workflow.audit.gates.codex: required`는 그대로 두며, 이 SPEC 단계에서 끄지 않는다. 기존 기본값의 전환은 `plan.md` §F.3의 순서를 따른다.
- 이 키의 실제 삽입은 계약 감사 통과와 lane-15 t1984(Codex 철수) 및 lane-14/16 체크포인트 ACK 이후로 미룬다. 현재 설정·Codex 파일에는 쓰기 배정이 없다.

### F.5 "공식 QA 스킬"의 정의

- **관측 사실**: 이 트리에는 "공식 QA 스킬"을 자칭하는 스킬이 없다. `.claude/skills`에 `moai-e2e`·`moai-gate` 디렉터리는 없고, 가까운 것은 `/moai gate` 워크플로(`.claude/skills/moai/workflows/gate.md`), `/moai e2e` 워크플로(`.claude/skills/moai/workflows/e2e.md`), `moai-workflow-testing` 스킬, `verify` 스킬(커밋 직전 품질 게이트 위임)이다. Codex 쪽 거울로는 `.agents/skills/moai-gate`, `.agents/skills/moai-e2e`가 있다.
- **정의(결정)**: 이 프로그램에서 "공식 QA 스킬"은 다음 최소 계약을 모두 만족하는, 배포 카탈로그에 등록된 단일 진입점을 말한다. 대상은 `/moai gate`(같은 HEAD의 lint·format·type·test)로 정하고, 변경이 웹·브라우저 표면을 건드릴 때만 `/moai e2e`를 추가 단계로 붙인다.
  1. 판정 대상 HEAD의 전체 SHA를 입력으로 받고 결과에 그대로 적는다.
  2. 결과는 `PASS` / `FAIL` / `BLOCKED` 셋 중 하나이며, 실행하지 않은 검사는 `BLOCKED`다.
  3. 아무것도 훑지 않은 실행(빈 선택, `[no tests to run]`)은 PASS가 아니다.
  4. 각 검사의 명령·종료 코드·결과 분류(`tdd-result-contract.md`의 `PASS` / `REGRESSION_FAILURE` / `TOOL_FAILURE`)를 남긴다.
  5. 비밀을 출력하지 않고, 읽기 전용 상태 모드에서는 파일을 바꾸지 않는다.
- 운영자가 다른 대상을 염두에 두었을 수 있으므로 OQ-5로 확인한다.

### F.6 "선형 이력" 문구 — 결정: 현재 설정 보존으로 해석

- **관측**: `main`의 `required_linear_history`는 **꺼져 있고**, 저장소는 merge commit 병합을 허용하며, `origin/main`에는 merge commit이 있다. `git-strategy.yaml`은 `mode: manual`, `manual.workflow: git-flow`, `merge_method: squash`다. 조정자 메시지가 말한 "agentic-trunk"는 `git-strategy.yaml`에는 없고, 로컬 규칙 `.claude/rules/local/repo-local-pr-policy.md`·`gitflow-lane-protocol.md`에 있다. 그 규칙에 따르면 `main`은 PR로만 전진하고, release PR(`release/main-batch-YYYYMMDD` 포함)은 merge commit 방식이며, 카드 PR은 리더가 auto-merge 요청 없이 병합한다. 같은 규칙은 "live 값은 cutover 전까지 임시 `github-flow`"라고 적지만, 측정한 live 값은 `git-flow`다. 이 차이는 OQ-6에 함께 넣는다.
- **문구(결정)**: "기존 branch protection과 선형 이력 정책을 유지한다"는 다음 뜻으로 읽는다. 측정한 `main` 보호 설정(필수 검사 3개, `enforce_admins`, 강제 푸시·삭제 금지, 현재 꺼져 있는 `required_linear_history`)을 바꾸지 않고, `main`의 이력을 다시 쓰지 않으며, 병합 방식은 설정된 `merge_method`를 따른다. AgenticFactory가 선형 이력을 새로 켜거나 그것에 기대지 않는다(REQ-AF-019).
- **남은 불일치**: Issue 작성자가 선형 이력이 이미 켜져 있다고 여겼을 수 있다. 이를 켤지는 OQ-6이다.

### F.7 병합 게이트 결정 D1~D5 (lane-22 갭 매트릭스에서 옮겨 이 SPEC이 정함)

lane-22 보고서(`a-merge-gate-gap-matrix.md`)는 D1~D5를 "결정 필요 항목"으로만 적고 정하지 않았다. 보고서는 커밋되지 않으므로 이 SPEC이 내용을 옮기고 결정한다. 결정은 모두 이 SPEC 범위의 결정이며, 보호 규칙·워크플로·설정을 실제로 바꾸는 결정은 하나도 없다.

**D1 — 선형 이력**
- 결정: 보호 규칙을 바꾸지 않는다. "선형 이력 정책 유지"는 F.6의 해석(현재 설정 보존, `main` 이력 재작성 금지)을 따른다. AgenticFactory는 `required_linear_history`를 켜지도 기대지도 않는다.
- 근거: `required_linear_history: false`(이 세션의 `gh api` 측정, lane-22 측정과 같음). 저장소는 merge commit을 허용하고 `origin/main`에 merge commit이 있다. 로컬 규칙상 release PR은 merge commit 방식이다.
- 출처: research.md R6, lane-22 G6, `.claude/rules/local/repo-local-pr-policy.md`.
- 결과: REQ-AF-019, AC-AF-019. 켤지 여부는 운영자 승인 대상이며 OQ-6이다.
- 보고서와의 차이: lane-22는 squash를 "운영 관행"으로 읽었다(`git-strategy.yaml` `merge_method: squash`). 이 세션은 `.github/workflows/auto-merge.yml:627-632`에서 "All other branches use --squash" 주석과 `MERGE_FLAG="--squash"`를 읽었다. 즉 squash는 일부 브랜치에만 해당하고 release PR은 merge commit이다. 둘은 모순이 아니라 범위 차이지만, "squash만 쓴다"로 일반화하면 틀린다.

**D2 — 병합 게이트의 Codex 의존**
- 결정: 현행 `auto-merge.yml`의 Codex 리뷰 읽기는 지금 제거하지도 바꾸지도 않는다. AgenticFactory 병합 게이트(REQ-AF-018)의 "같은 HEAD 코드·보안 리뷰"는 공급자 중립으로 정의하며, Codex가 미설정이면 Codex가 아닌 리뷰 경로가 그 조건을 채워야 한다. 그 경로가 검증되기 전까지 Codex 미설정 프로젝트의 병합은 `blocked(not-configured)`다. 현행 게이트 변경은 plan.md §F.3의 3단계(기본값 변경)에서만 다룬다.
- 근거: `.github/workflows/auto-merge.yml:29,38`이 Codex 앱 계정(user id 199175422)의 PR 코멘트를 기동 조건으로 쓰고, `:325`·`:407`이 `scripts/ci/codex-review-gate.sh`를 부른다(이 세션에서 grep으로 확인). Issue는 "Codex는 사용자가 명시 설정한 경우에만"이라고 적는다.
- 출처: lane-22 G3, 이 세션의 grep.
- 결과: REQ-AF-013, REQ-AF-018, REQ-AF-022, AC-AF-013, AC-AF-020. 대체 리뷰 경로의 실체는 정하지 않았다(lane-22 보고서는 "19 담당"으로 적음) — OQ-9와 함께 본다.

**D3 — PASS-WITH-DEBT 착지 선례**
- 결정: 무부채 PASS 기준은 AgenticFactory 경로로 처리되는 Issue에만 적용한다. 이미 착지한 기존 카드는 다시 열지 않는다. Phase 2에서 이관되는 카드가 AgenticFactory 경로로 다시 실행되면 그때 무부채 PASS를 새로 받아야 한다.
- 근거: Issue는 PASS-WITH-DEBT를 합격으로 쓰지 않는다. lane-22는 `a8372b21f5`(#1936)가 커밋 메시지상 "iteration-3 sync audit PASS-WITH-DEBT" 뒤에 착지했다고 적었다(이 세션은 그 커밋을 읽지 않음).
- 출처: lane-22 G4.
- 결과: REQ-AF-018, AC-AF-018, `block_reason`의 `debt-present`.

**D4 — 판정 파일의 HEAD 순환**
- 결정: SHA 결합 증거는 코드 후보 브랜치 밖에 둔다(제안 위치: Issue별 증거 전용 ref). 증거 항목은 판정한 HEAD 전체 SHA와 내용 digest를 싣고, 병합 게이트는 읽기 직전에 후보 HEAD와 대조한다. 이전 HEAD의 결과로는 새 HEAD를 통과시킬 수 없다.
- 근거: lane-22 G7 — 판정 파일이 PR 브랜치 안에 커밋되고, 새 HEAD에 다시 묶는 PR #1939가 있다(PR 제목 근거).
- 출처: lane-22 G7·D4, `design.md` §11.
- 결과: REQ-AF-020, AC-AF-018. 원격에 올릴 수 있는지는 OQ-14.

**D5 — 필수 검사 집합**
- 결정: AgenticFactory가 말하는 "필수 검사"는 `main` 보호 규칙에 등록된 집합(현재 `Lint`, `Build (linux/amd64)`, `Analyze (Go) (go)`)이다. 보호 규칙에 검사를 더하는 일은 약화는 아니지만 비용이 있어 운영자·리더 판단이며, 이 SPEC은 더하지 않는다. 필수가 아닌 검사의 SKIPPED는 병합 게이트의 통과 조건에 들어가지 않는다는 한계를 그대로 기록한다.
- 근거: 이 세션의 보호 규칙 측정(research.md R6). lane-22 G2: 현행 `auto-merge.yml`은 필수 검사의 `skipping`·`fail`·`cancel`을 보류로 처리하지만 필수가 아닌 검사의 SKIPPED(#1943에서 12건 관측)는 보지 않는다. `strict: false`라 base에 뒤처진 head도 보호 규칙을 통과한다(G1).
- 출처: lane-22 G1·G2·D5.
- 결과: REQ-AF-018, AC-AF-018. 집합을 넓힐지는 OQ-15.

### F.8 카드 TTL/renew 대 지속 배정 + epoch — 판정 요약

상세 비교는 `design.md` §14다.

- 판정: 기존 사용자 결정(카드 TTL 15m, renew 5m)과는 원칙상 충돌하지 않는다. TTL을 활성 힌트로, 배정 epoch를 권위로 읽으면 된다. 자원 slot lease의 "기한이 지나면 살아 있어도 인계" 규칙과는 충돌하므로 claim을 slot lease로 만들지 않는다(F.1).
- 조정: TTL·renew 값과 범위 검사는 그대로 둔다. 인계는 ControlService만, 활성 판정 `dead`(§F.2 정본) 또는 안전 정지 기록 뒤에만, epoch+1로 한다. 오래된 epoch의 쓰기는 거절한다(REQ-AF-010).
- 남은 결정: 기한이 지난 뒤 살아 있는 소유자의 renew를 계속 `refused-expired`로 거절할지는 기존 동작을 바꾸는 문제라 사용자에게 묻는다(OQ-11). 답이 오기 전까지는 현행 거절을 유지한다.

### F.9 현행 진입 명령 확인 결과

`design.md` §9.1에 파일:줄과 함께 적었다. 요약: `-f`는 leader 진입(`internal/cli/cc.go:58`), `-l`은 다음 빈 `lane-<n>` 합류(`cc.go:65-66`), `-f`와 `-l` 동시 입력 거절(`internal/cli/factory.go:271-272`), `--leader`와 `--factory-run` 동시 입력 거절(`factory.go:280-282`), `worker-<n>`·`agent-<n>` 레거시 거절(`factory.go:332-347`, `:259-263`, `internal/factory/bootstrap.go:237-242`). `--leader`는 `-l` 없이 주면 거절된다(`factory.go:284-286`: `if !p.LaneRole { return p, errors.New(leaderNeedsLaneEntry) }`, 문구는 `:102`). 같은 `--leader`나 같은 `--factory-run`을 두 번 주는 경우의 거절은 읽은 범위(`factory.go:153-189`)에서 찾지 못했다 — 이 코드는 값을 다시 대입한다. 상위 설계 권고의 "중복 `--leader`/`--factory-run` 거절" 서술과 다르므로 확인 공백으로 남긴다.

## §G 미결 질문

| ID | 질문 | 막는 것 |
|---|---|---|
| OQ-1 | 시나리오 AC 일곱 개를 처음부터 끝까지 측정할 실제 Issue는 무엇인가? 시나리오 7의 "다음 dispatch"를 실제로 보이려면 대기 중인 두 번째 실제 Issue가 하나 더 필요하다. 실제 Issue에 댓글을 쓰는 것은 외부 공유 동작이라 운영자 승인이 필요하고, 보호된 `main`으로의 실제 병합과 Issue 종료(시나리오 6·7)는 운영자 keep-set 승인 대상이다. | AC-AF-001~007 측정 |
| OQ-2 | 상위 설계 권고는 도착해 PROPOSAL로 반영했다(`design.md` §13). 리더·사용자가 명칭(Agentic Factory / ControlService / Coordinator / Worker)과 정식 명령(`start`·`status`·`attach`)을 확정하는가? 또 권고가 정하지 않은 두 가지 — `run`의 범위(`design.md` §7.3), 새 실행 자리 `slot`과 기존 자원 `slot` lease의 이름 충돌 — 는 무엇으로 정하는가? | REQ-AF-023~025 확정 |
| OQ-3 | lane-22 대응표는 이 트리에는 없고 `agentic-factory-integration` 작업 트리의 미커밋 보고서로 읽었다. 그 내용(S1~S4, C1~C11, PR 착지 순서 후보)을 `design.md` §9.3~§9.5에 옮겨 맞췄다. 보고서가 바뀌면 누가 다시 맞추는가? 또 C1~C11의 "새 명령" 칸은 lane-22와 이 SPEC 모두 미정이다. | S2 이후 단계 진입 |
| OQ-4 | 조정자가 말한 "여섯 범위(six-scope) 소유 경계"의 정식 목록을 이 트리에서 찾지 못했다. 이 SPEC은 `design.md` §10의 여섯 범위를 제안으로 둔다. | 코드 경계 확정 |
| OQ-5 | "공식 QA 스킬"을 `/moai gate`(+조건부 `/moai e2e`)로 정한 결정이 운영자 의도와 맞는가? | REQ-AF-018 판정 원천 |
| OQ-6 | `main`에 `required_linear_history`를 켤 의도가 있는가? 켜면 merge commit을 쓰는 release PR 착지 방식과 충돌한다. 또 `git-strategy.yaml`의 live 값(`git-flow`)과 로컬 규칙이 말하는 임시 값(`github-flow`)·목표 값(`agentic-trunk`)이 서로 다르다. 어느 쪽이 기준인가? | REQ-AF-019 문구 |
| OQ-7 | egolite·aside 호출 계약은 lane-21 보고서로 받았다(`design.md` §5.2). 그러나 연결 실패·앱 미실행 시 오류 문구와 종료 코드, aside 로그인 상태 감지 명령은 아무도 측정하지 않았다. 차단 사유 `connection-failed`·`not-logged-in`을 가를 근거를 누가 측정하는가? | REQ-AF-016, `block_reason` |
| OQ-8 | 기존 리더·writer의 안전 중단 ACK가 아직 확인되지 않았다(Issue 본문). 단일 accountable owner와 첫 구현 조각은 누가 확정하는가? | run 진입 |
| OQ-9 | PR #1937(t1980, "shipped codex session defaults를 terra/xhigh로 바꿈")은 Codex 기본값을 건드린다. REQ-AF-022의 순서와 충돌하는지 판정이 필요하다. | Codex 전환 순서 |
| OQ-10 | GLM 이외 "지원 로컬 환경"의 목록(예: `moai cc`, `moai codex`)을 무엇으로 확정하는가? Codex 환경은 명시 설정이 없으면 매트릭스에서 `BLOCKED(미설정)`로 남는다. | AC-AF-014 |
| OQ-11 | 기한이 지난 뒤 살아 있는 소유자의 `renew`를 현행처럼 `refused-expired`로 거절할 것인가, 같은 epoch의 살아 있는 소유자에게는 활성 힌트를 되살리도록 허용할 것인가? 후자는 기존 사용자 결정(카드 TTL/renew)의 동작을 바꾼다(`design.md` §14.3). | REQ-AF-010 세부 |
| OQ-12 | 제안 정식 명령 `moai factory status`는 기존 명령(`internal/cli/factory_card.go:4175`)과 이름이 같다. 기존 출력을 그대로 두고 필드만 더할 것인가, 다른 이름을 쓸 것인가? | REQ-AF-024·025 |
| OQ-13 | 브라우저: `CI=true`일 때 3선택 모델에서 aside를 어떻게 다루는가(SPEC-ASIDE-BROWSER-001 REQ-ASB-011)? aside `exec`의 권한 기본값(Guard 유지, `full-access` 금지)과 `repl` 사용 범위를 그대로 이어 가는가? | REQ-AF-016 |
| OQ-14 | 증거 전용 ref(`refs/agentic-factory/evidence/<issue>`)를 원격에 올려도 되는가? 외부 공유 동작이다. 안 된다면 증거를 어디에 둘 것인가(Issue 댓글 첨부, 검사 결과 등)? | REQ-AF-020 |
| OQ-15 | 필수 검사 집합을 보호 규칙의 세 검사에서 넓힐 것인가? 넓히면 실행 비용·불안정 비용이 따른다(결정 D5). | REQ-AF-018 범위 |
| OQ-16 | 설정 키 `workflow.agentic_factory.slots`의 상한은 얼마인가? 하한은 1로 정했지만 상한의 근거(기기 용량, 리뷰 처리량)가 이 트리에 없다. 정해지기 전까지 이 키는 "1 이상, 상한 미정"이며 AC-AF-017은 하한 위반과 형식 위반만 시험한다. | REQ-AF-017 범위 |
| OQ-17 | 활성 판정 정본(§F.2)의 한계: Linux `ps -o lstart=`는 초 단위라 같은 초 안의 PID 재사용을 구분하지 못할 수 있고, Windows `GetProcessTimes` 비교의 세부 의미는 읽지 않았다. 이를 측정해 정본 결정을 확정하거나, 더 정밀한 시작 시각 원천을 요구할 것인가? | REQ-AF-010 확정 |
| OQ-18 | 브라우저 3선택을 사람이 보는 설정 화면(웹 UI)에 그릴 것인가? 그 화면은 열린 PR #1937·#1944가 고치는 `internal/settings` 영역이라 이 SPEC은 설정 키 저장·다시 읽기 계약만 정하고 화면 표시는 범위 밖으로 둔다(§I). | REQ-AF-016 표시 범위 |

## §H 확인 공백 (관측하지 못한 것)

- 카드 t1965(브라우저 설정, HOLD): 이 트리에 `.moai/reports/t1965`가 없고, SPEC-FETCH-WEB-001이 "t1965 linkage"로 언급할 뿐이다. 내용을 읽지 못해 설계 입력으로만 기록한다.
- PR #1937 / #1944 / #1943 / #1916 / #1921은 `gh issue view`로 제목·상태만 읽었다. 변경 파일 목록은 읽지 않았다. 충돌 범위(`internal/settings`, `internal/config`, factory, launcher)는 조정자 지시와 제목에 근거한다.
- `moai codex -l` / `moai glm -f` / `moai glm -l`의 구현은 읽지 않았다. `moai cc -f` / `-l` / `--leader`는 `internal/cli/cc.go` 도움말에서 읽었고, `moai codex -l`은 `.moai/config/sections/llm.yaml` 주석에서만 확인했다.
- Aside 공식 개발자 문서의 호출 계약은 읽지 않았다(외부 문서, 이번에는 URL을 열지 않음).
- #1926의 두 번째 댓글은 앞부분(약 1,200자)만 읽었다. 해시 불일치 가설이 "측정됐다"는 결론 줄은 읽었다.
- #1847·#1813의 Codex 연결 수리 커밋(`e8c42a231c`, `fa562e3110`)이 원격에 있는지는 확인하지 않았다.

- 상위 설계 권고가 말한 "중복 `--leader` / `--factory-run` 거절"을 이 트리에서 확인하지 못했다(`internal/cli/factory.go:153-189`는 반복 시 값을 다시 대입). `--leader`가 `-l`을 요구하는 검사는 0.2.0에서 `factory.go:284-286`을 읽어 확인했다.
- 도구 출처(VCI §2.2): spec lint는 이 트리에서 빌드한 바이너리로 돌렸다. 빌드 명령은 `go build -ldflags "-X github.com/modu-ai/moai-adk/pkg/version.Version=v3.2.0-tree -X github.com/modu-ai/moai-adk/pkg/version.Commit=e14b5bbac2" -o <scratchpad>/moai-tree2 ./cmd/moai`이고, `version` 출력은 `v3.2.0-tree e14b5bbac built unknown`이다. 커밋 표시는 빌드 시점 HEAD(`e14b5bbac2`)를 손으로 넣은 값이며 자동 기록이 아니다. 빌드 날짜는 넣지 않았고, 빌드 당시 트리에는 커밋되지 않은 SPEC 문서 변경이 있었다(코드 변경은 없음). `Makefile`의 정식 빌드(`make build`)는 쓰지 않았다. 설치된 `moai`(b1806498fe)는 HEAD의 조상이 아니어서(`git merge-base --is-ancestor b1806498fe HEAD` → exit 1) 판정에 쓰지 않았다.
- 활성 판정의 Windows 구현(`factory_alive_windows.go`, `process_fingerprint_windows.go`, `pid_state_windows.go`)은 함수 머리와 분기만 읽었고 실행해 보지 않았다(OQ-17).
- lane-22·lane-21 보고서(`/Users/goos/MoAI/moai-adk-go/.moai/worktrees/agentic-factory-integration/.moai/reports/agentic-factory/a-merge-gate-gap-matrix.md`, `c-factory-migration-map.md`, `/Users/goos/MoAI/moai-adk-go/.moai/worktrees/agentic-factory-21/.moai/reports/agentic-factory/21-browser-contracts.md`)는 읽었지만, 그 안의 측정(브라우저 도구 실행, PR 상태, `auto-merge.yml` 186-316행 해석, `a8372b21f5` 커밋 메시지, #1943 검사 집계, worktree 등록부)은 이 세션이 다시 실행하지 않았다. 이 SPEC은 그것을 보고서 출처로만 인용한다.
- 보고서와 이 트리의 차이: (a) lane-22는 squash를 운영 관행으로 읽었으나 `auto-merge.yml:627-632`는 일부 브랜치만 squash다(D1). (b) lane-22는 Codex 리뷰 게이트를 `auto-merge.yml:392-420`으로 적었고 이 세션은 `:325`와 `:407` 두 곳에서 `codex-review-gate.sh` 호출을 찾았다(범위 차이). (c) lane-22 C8의 `moai factory approve`는 설치 바이너리 `--help` 근거였고, 이 트리에서는 `internal/cli/leader_approval_gate.go:286`에 `Use: "approve <card-id> --run <run-id>"`가 있음을 확인했다(명령 존재는 일치. 0.1.0이 적은 `Use: "approve"`는 문자열을 줄여 적은 것이라 0.2.0에서 정정).
- `auto-merge.yml`은 grep으로 일부 줄만 봤다. `scripts/ci/codex-review-gate.sh` 본문은 읽지 않았다.

## §I 범위 밖

### Out of Scope — 남은 카드의 Issue 이관 (Phase 2)

- 남은 카드·열린 PR·이슈의 대사, 중복 확인, 재우선순위, 기존 owner 인계는 이 SPEC 뒤의 하위 단계다.
- 이관 자체는 구현 완료로 세지 않는다.
- 하위 이슈를 만들거나 기존 이슈를 대량 발행·종료하지 않는다.

### Out of Scope — 기존 설정·코드 수정

- `.moai/config/**`, `internal/config`, `internal/settings`, 템플릿 트리(`internal/template/templates`)는 이번에 고치지 않는다. 설정 키는 초안일 뿐이며, 실제 삽입은 계약 감사 통과와 해당 레인 ACK 이후다.
- `internal/cli/factory_card.go`, `internal/factory/slot_lease.go`, `internal/cli/launcher.go`는 첫 writer의 범위 밖이다.
- `internal/factory/factory_alive_*.go`(`FactoryProcessAlive`)와 그 기존 호출처는 고치지 않는다. claim 판정에 쓰지 않을 뿐이다(§F.2).
- SPEC-ASIDE-BROWSER-001은 고치지 않는다. 그 "제안 금지 조항"(REQ-ASB-009)은 대체 대상으로만 기록하며, 실제 개정은 대체 동작이 검증된 뒤 별도로 한다.

### Out of Scope — 기존 결함의 종료와 기존 작업의 처분

- #1938 / #1926 / #1924 / #1847 / #1813을 자동으로 닫지 않는다. 대체 동작과 재현 시험만 연결하고 처분은 나중에 따로 한다.
- 기존 PR을 닫거나, 브랜치를 덮어쓰거나, 카드를 일괄 done 처리하지 않는다. 기존 병합 보류도 유지한다.
- 강제 프로세스 종료, 터미널 종료, 작업 트리 삭제, 미커밋 변경 폐기는 하지 않는다.

### Out of Scope — 설정 화면(웹 UI) 표시

- 브라우저 3선택을 사람이 보는 설정 화면에 그리는 일은 다루지 않는다. 이 SPEC은 선택 저장·다시 읽기·자동 선택 금지 계약(REQ-AF-016)만 정한다. 화면은 열린 PR #1937·#1944의 `internal/settings` 영역과 겹치므로 그 작업이 끝난 뒤 따로 다룬다(OQ-18).

### Out of Scope — 공급자·비용

- GLM managed 경로, Codex 자동 활성화, 새 키·유료 용량·추가 지출, Cloud 개발 재개, 앱 설치는 다루지 않는다.

### Out of Scope — 충돌 중인 열린 작업

- PR #1937·#1944(`internal/settings`, `internal/config`), #1943·#1916·#1921(factory, launcher), 카드 t1965(브라우저 설정, HOLD)의 파일은 건드리지 않는다. 각 공유 접점은 레인 하나만 쓴다(`design.md` §10).

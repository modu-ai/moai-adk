---
id: SPEC-DUAL-HARNESS-RECOVERY-001
title: "Dual-harness recovery — Codex wiring unwire/rollback, Codex worktree and kanban parity, role permission contract, exactly-once dispatch results, mixed factory card flow"
version: "0.1.0"
status: draft
created: 2026-09-23
updated: 2026-09-23
author: manager-spec
priority: P1
phase: "v3.2.0 target"
module: "internal/codexwiring, internal/factorymsg, internal/cli, internal/template/agentemit"
lifecycle: spec-anchored
tags: "dual-harness, codex, claude, codexwiring, rollback, worktree, kanban, agentemit, factorymsg, factory"
tier: L
card: t1100
depends_on:
  - SPEC-FACTORY-MIXED-HOOK-001
  - SPEC-CODEX-WIRING-001
related_specs:
  - SPEC-FACTORY-LANE-WORKTREE-HANDOFF-001
  - SPEC-CODEX-DUAL-AGENTS-001
  - SPEC-CODEX-LAUNCH-VERB-001
  - SPEC-CODEX-LAUNCHER-001
  - SPEC-CODEX-PARTIAL-WIRING-001
  - SPEC-UPDATE-ADD-CODEX-001
---

# SPEC-DUAL-HARNESS-RECOVERY-001

## HISTORY

| Version | Date | Change |
|---|---|---|
| 0.1.0 | 2026-09-23 | 카드 t1100 plan 초안. 이중 하네스 설계(§19)의 AC-MIG-01, AC-WT-01, AC-AGENT-01, AC-MSG-01, AC-FACT-01 다섯 기준만 다룬다. |

## §A 배경과 목적

`reports/moai-dual-harness-full-design-20260922.md`(설계 제안, 구현 기록 아님)는 Claude Code와 Codex CLI의 동등 지원을 14개 인수 기준으로 정의한다. 이 SPEC은 그중 상태·복구(M4)와 실행·협업(M3)에 속하는 다섯 기준만 구현 가능한 요구사항으로 옮긴다.

이 트리(`d87e9af2e`)에서 코드를 읽고 확인한 출발점은 다음과 같다. 자세한 근거와 줄 번호는 `research.md`에 있다.

- `internal/codexwiring`은 설치만 한다. 제거 경로가 없고, 신뢰 사이드카(`.moai/state/codex-wiring.json`)는 전체 파일 해시 두 개만 담는다. 원자적 쓰기는 temp+rename뿐이며 저널, 잠금, rename 직전 해시 재확인, 기록 후 대조가 없다.
- claude 프로필 배포는 `.codex/`를 숨기지만(`harness_fs.go` `hideCodex`), 이미 있는 `.codex/*`를 치우지 않는다. `both → claude` 전환 후 배선 파일과 에이전트 TOML이 고아로 남는다.
- `moai codex -w`는 새 트리를 만들거나 기존 트리에 들어가지만, 동시 writer 검사나 소유 표식을 남기지 않는다. `moai codex -k`는 없다(`-f`는 있다).
- Codex 역할 TOML은 작업 공간 단위 `sandbox_mode`만 표현한다. 역할별 shell·MCP 도구 제한은 표현할 수 없다.
- `internal/factorymsg`는 전달을 at-least-once로 보장하고 멱등 키로 송신 중복을 막는다. 멱등 범위가 송신자 세션 UUID라서 송신자가 재시작하면 같은 키도 새 메시지가 된다. 결과를 한 번만 반영하는 작업 레코드는 없다.
- 혼합 팩토리 4조합 테스트는 `MOAI_FACTORY_LIVE=1` 없이는 모두 `SKIP`이다. 이때도 패키지 결과는 `ok`로 출력된다(이번 실행에서 관측).

## §B 범위

다섯 설계 기준을 다음 REQ 묶음으로 옮긴다.

| 설계 기준 | 이 SPEC의 REQ | 이 SPEC의 AC |
|---|---|---|
| AC-MIG-01 | REQ-DHR-001 ~ REQ-DHR-007 | AC-DHR-001 ~ AC-DHR-005 |
| AC-WT-01 | REQ-DHR-008 ~ REQ-DHR-012 | AC-DHR-006 ~ AC-DHR-009 |
| AC-AGENT-01 | REQ-DHR-013 ~ REQ-DHR-015 | AC-DHR-010 ~ AC-DHR-013 |
| AC-MSG-01 | REQ-DHR-016 ~ REQ-DHR-021 | AC-DHR-014 ~ AC-DHR-016 |
| AC-FACT-01 | REQ-DHR-022 ~ REQ-DHR-024 | AC-DHR-017 ~ AC-DHR-019 |

### 이전 결정과의 관계

- SPEC-FACTORY-MIXED-HOOK-001 REQ-FMH-006은 브로커가 at-least-once 전달을 유지하고 exactly-once 실행을 주장하지 않는다고 정했다. 이 SPEC은 그 결정을 뒤집지 않는다. "한 번만 반영"은 메시지 전달이 아니라 작업 레코드(dispatch record)에 결과를 적용하는 단계에서 보장한다.
- SPEC-CODEX-WIRING-001 REQ-CW-005는 기존 `[mcp_servers.moai]` 테이블과 `status_line` 키를 사용자 소유로 본다. 이 SPEC은 MoAI가 직접 만든 부분만 MoAI 소유로 기록하고, 설치 전에 있던 부분은 계속 사용자 소유로 둔다.
- SPEC-CODEX-WIRING-001 REQ-CW-012는 배선 생성기가 `.codex/agents/**`를 건드리지 못하게 한다. 에이전트 TOML의 정리는 배선 생성기가 아니라 템플릿 manifest 경로(REQ-DHR-007)가 맡는다.
- 설계 ADR-02(기존 상태 저장 API 재사용)에 따라 소유 기록은 `internal/manifest`, 작업 레코드는 기존 factorymsg SQLite 브로커를 확장한다. 새 저장소나 데몬은 만들지 않는다.

## §C 요구사항 (GEARS)

### REQ-DHR-001 — Codex wiring ownership provenance

The Codex wiring generator SHALL record, for every file and every file part it writes, a provenance entry through `internal/manifest` naming the path, the owned part (whole file, a hook handler identified by its `moai hook ` command, a TOML table, or a TOML key), whether MoAI created that part or found it pre-existing, and the hash of the bytes MoAI wrote. A part that existed before MoAI first wrote the file SHALL be recorded as user-owned and SHALL never be promoted to MoAI-owned by a later run.

### REQ-DHR-002 — Journaled per-file atomic write with readback

When the Codex wiring generator writes a file, it SHALL stage the full content in a temporary file in the target directory, append a journal entry naming the path, the pre-write hash, and the intended post-write hash, acquire the existing project update lock, re-read the target and compare its hash with the pre-write hash immediately before rename, rename, then read the target back and compare its hash with the intended post-write hash before marking the journal entry complete. The generator SHALL NOT claim that a multi-file wiring pass is atomic as a whole.

### REQ-DHR-003 — Refusal on concurrent or user modification

When the target hash re-read immediately before rename differs from the pre-write hash, the Codex wiring generator SHALL NOT rename, SHALL leave the target byte-identical, SHALL record the journal entry as `conflict`, and SHALL report the path and both hashes as a conflict with a non-zero outcome for that file.

### REQ-DHR-004 — Interrupted-install recovery

When `moai tool enable codex`, `moai update`, the Codex unwire command, or `moai doctor` finds an incomplete wiring journal entry, the recovery step SHALL classify each entry by the target's current hash as `completed` (equals the post-write hash), `not-applied` (equals the pre-write hash), or `diverged` (equals neither), SHALL finalize `completed` entries, SHALL discard the staged file of `not-applied` entries without touching the target, and SHALL leave `diverged` targets untouched and reported. Recovery SHALL NOT overwrite a file whose current hash is neither the pre-write nor the post-write hash.

### REQ-DHR-005 — Owned-part-only unwire

When the operator runs the Codex unwire command, the unwire step SHALL remove only parts whose provenance is MoAI-created AND whose current bytes hash to the recorded hash. Where a file also contains user-owned parts, the unwire step SHALL rewrite the file with only the MoAI-owned parts removed through the REQ-DHR-002 write path, preserving every other byte. The unwire step SHALL delete a whole file only when the entire file is MoAI-created and its hash matches the record.

### REQ-DHR-006 — No deletion on hash alone

The Codex unwire and recovery steps SHALL NOT delete or modify a file or part on hash evidence alone. When a part lacks a MoAI-created provenance record, is recorded user-owned, has a hash mismatch, or resolves through a symbolic link to a location outside the project root, the step SHALL leave it untouched and SHALL report its path and the reason (`no-provenance`, `user-owned`, `modified`, or `symlink-boundary`).

### REQ-DHR-007 — Harness profile transition preservation

When the configured harness profile transitions (`claude → both`, `gpt → both`, `both → claude`, `both → gpt`, `gpt → claude`, or an older binary's deployment is refreshed by a newer one), the deployment SHALL preserve every user-owned file and part. When the target profile no longer deploys a path that the manifest records as template-managed (for example `.codex/agents/moai/*.toml` after `both → claude`), the deployment SHALL classify it with the existing deprecated-path classification, SHALL remove only pristine entries after backup, and SHALL keep and report user-modified or unverified entries. When the Codex wiring files become orphaned by such a transition, the deployment SHALL report them and SHALL NOT remove them unless the Codex unwire step runs.

### REQ-DHR-008 — Codex worktree anchor

When `moai codex -w` launches a Codex session into a new or existing card worktree, the launcher SHALL place a git worktree lock on that tree whose reason carries `pid <n>` of the process that becomes the Codex session, before the Codex process starts, so the existing lock-aware anchor decision recognizes the tree as anchored for the session's lifetime. When the tree already carries a lock whose recorded pid is confirmed dead, the launcher SHALL replace it; it SHALL NOT replace a lock whose pid is live or whose liveness is undetermined.

### REQ-DHR-009 — Concurrent writer rejection

When `moai codex -w` or `moai cc -w` targets an existing worktree that the existing anchor decision reports as anchored by a live or undetermined owner other than the caller, the launcher SHALL refuse to launch with a non-zero exit and a diagnostic naming the anchor source and holder, and SHALL NOT modify the tree, its lock, or its branch.

### REQ-DHR-010 — Un-integrated tree deletion rejection

The worktree disposal paths (`moai worktree clean --stale`, `moai worktree done`, and session-exit cleanup) SHALL refuse to remove a Codex-created worktree that holds commits not integrated into its base branch, holds uncommitted changes, or is anchored by a live or undetermined owner, with the same classification and diagnostics they apply to Claude-created trees.

### REQ-DHR-011 — Codex worktree creation base verification

When `moai codex -w` creates a new worktree, the launcher SHALL resolve the base through the existing base resolution, SHALL verify after creation that the new tree's HEAD equals the resolved base commit, and when they differ SHALL refuse to launch and leave the tree for inspection without deleting it. The launcher SHALL NOT change the base resolution policy.

### REQ-DHR-012 — Codex kanban entry parity

Where the operator passes `-k` or `--kanban` to `moai codex`, the launcher SHALL accept the same entry shapes as `moai cc -k` for the roles this SPEC admits, SHALL publish the same kanban launch facts with backend `codex`, SHALL claim or resolve the session name with the same rules, and SHALL NOT forward the kanban tokens to the Codex child process. When an unsupported shape is passed, the launcher SHALL exit non-zero with a usage diagnostic rather than degrade to a plain launch.

### REQ-DHR-013 — Role permission contract

The agent emitter SHALL carry a permission contract for each of the 12 Codex roles naming the role's required write scope, shell use, MCP servers, subagent spawning, and web access, and SHALL map every axis to exactly one of `enforced` (with the Codex field that enforces it) or `UNSUPPORTED` (with the host-expressivity reason). When a role's contract requires a restriction that is neither enforced nor declared `UNSUPPORTED`, the emitter SHALL fail. The emitter SHALL NOT emit a sandbox value broader than the role's contract requires.

### REQ-DHR-014 — Unsupported axes are never reported as passing

The Codex role verification SHALL report every `UNSUPPORTED` axis as `UNSUPPORTED` and SHALL NOT count it as PASS. When runtime role loading and read-only enforcement are verified against a real Codex binary, that verification SHALL be recorded as a separate evidence item naming the Codex version; an unexecuted runtime verification SHALL be reported as `NOT_RUN`.

### REQ-DHR-015 — Auditor write-scope boundary

Where `plan-auditor` or `sync-auditor` runs on Codex with `workspace-write`, the audit workflow SHALL capture a working-tree snapshot before the auditor starts and verify it after the auditor returns against an allowlist of the auditor's declared report and verdict paths. When any path outside the allowlist changed, the verification SHALL exit non-zero, the audit verdict SHALL be rejected, and the changed paths SHALL be reported.

### REQ-DHR-016 — Dispatch record

The factory broker SHALL store one dispatch record per run and dispatch identifier carrying the dispatch ID, card ID, target lane slot, attempt number, the assignee generation used as the fencing token, the idempotency key, the lifecycle state, and a result reference. The lifecycle states SHALL be `assigned`, `delivered`, `started`, `result_recorded`, `integrated`, and `abandoned`, recorded as distinct states. A message arrival or receipt SHALL NOT move a dispatch past `delivered`.

### REQ-DHR-017 — Lane-stable idempotency scope

The factory broker SHALL scope message and result idempotency to the project key, run ID, sender lane slot, and idempotency key, independent of the sender's session UUID and generation, so that a retry after a sender restart deduplicates to the original record. When the same scope carries a different recipient lane, kind, task reference, correlation ID, or payload, the broker SHALL reject it without mutation.

### REQ-DHR-018 — Exactly-once result application under fencing

When a worker reports a result for a dispatch, the factory broker SHALL apply it in one transaction that verifies the reporter's lane and current generation equal the dispatch's assignee lane and generation and that the dispatch is in `started`. When the same result is applied again through duplicate delivery, lease redelivery after a lost receipt, or a retry after restart, the broker SHALL leave the dispatch unchanged and return a `duplicate` outcome. When the reporter's generation or attempt is stale, the broker SHALL reject it with a stale outcome and leave the dispatch unchanged. The broker message layer SHALL remain at-least-once.

### REQ-DHR-019 — Result persisted before receipt

When a worker processes a message that carries a result, the result SHALL be recorded in the dispatch record before the message receipt is acknowledged, so that a crash between recording and receipt leads to redelivery that REQ-DHR-018 resolves as `duplicate`.

### REQ-DHR-020 — Superseded-generation messages

When a lane's generation advances while messages addressed to its previous generation are pending or claimed, the factory broker SHALL NOT let the superseded generation claim, read, dispose, or acknowledge them and SHALL report them as `superseded` in the broker status rather than as pending. The broker SHALL re-issue dispatch authority only through a new dispatch attempt (REQ-DHR-021) and SHALL NOT release a superseded message body on its own.

### REQ-DHR-021 — Reassignment fences the previous attempt

When the lead reassigns a dispatch, the factory broker SHALL require that the previous assignee's owner is confirmed not live or that the lead explicitly revokes it, SHALL increment the attempt, SHALL record the new assignee lane and generation, and SHALL thereafter treat any result from the previous attempt as stale under REQ-DHR-018.

### REQ-DHR-022 — Deterministic mixed-backend card flow

The factory test suite SHALL drive each of the four backend combinations (Claude lead to Claude worker, Codex to Codex, Claude to Codex, Codex to Claude) through `assigned → delivered → started → result_recorded → integrated` against the real broker store without a model, including a worker interruption after `started`, a reassignment to a new attempt, and a late result from the interrupted attempt, and SHALL assert that exactly one result is applied per dispatch.

### REQ-DHR-023 — LIVE mixed factory card flow evidence

When LIVE verification runs, each of the four backend combinations SHALL complete the card flow of REQ-DHR-022 in real separate CLI and model contexts inside an isolated temporary repository and `MOAI_HOME`, within a declared per-case model invocation budget and timeout, with cleanup of every spawned process registered before the first spawn. Each case SHALL write its own evidence file. LIVE evidence SHALL be recorded separately from the deterministic evidence and SHALL NOT be claimed to run in CI unless a CI workflow that sets the LIVE gate is observed.

### REQ-DHR-024 — SKIP and NOT_RUN are not PASS

Every verification command in this SPEC SHALL count `pass` events for the exact named tests and SHALL fail when any named test reports `skip` or `fail`, when the expected pass count is not met, or when output contains `NOT_RUN`. A package-level `ok` line SHALL NOT be accepted as evidence.

## §D 요구사항 ↔ 인수 기준 추적

| Requirement anchor | Acceptance criteria |
|---|---|
| § REQ-DHR-001 | AC-DHR-003, AC-DHR-004 |
| § REQ-DHR-002 | AC-DHR-001 |
| § REQ-DHR-003 | AC-DHR-001 |
| § REQ-DHR-004 | AC-DHR-002 |
| § REQ-DHR-005 | AC-DHR-003 |
| § REQ-DHR-006 | AC-DHR-004 |
| § REQ-DHR-007 | AC-DHR-005 |
| § REQ-DHR-008 | AC-DHR-006 |
| § REQ-DHR-009 | AC-DHR-007 |
| § REQ-DHR-010 | AC-DHR-008 |
| § REQ-DHR-011 | AC-DHR-006 |
| § REQ-DHR-012 | AC-DHR-009 |
| § REQ-DHR-013 | AC-DHR-010 |
| § REQ-DHR-014 | AC-DHR-011, AC-DHR-012 |
| § REQ-DHR-015 | AC-DHR-013 |
| § REQ-DHR-016 | AC-DHR-014 |
| § REQ-DHR-017 | AC-DHR-014 |
| § REQ-DHR-018 | AC-DHR-014, AC-DHR-015 |
| § REQ-DHR-019 | AC-DHR-014 |
| § REQ-DHR-020 | AC-DHR-015 |
| § REQ-DHR-021 | AC-DHR-016 |
| § REQ-DHR-022 | AC-DHR-017 |
| § REQ-DHR-023 | AC-DHR-018 |
| § REQ-DHR-024 | AC-DHR-019 |

## §E t1082 경계 (SPEC-FACTORY-LANE-WORKTREE-HANDOFF-001)

t1082는 draft이며 run이 시작되지 않았다. 열린 결정 M1(launch-pending 중 handoff)과 M2(두 rebind 경로)는 t1082의 것이고, 이 SPEC은 그 둘을 결정하지 않는다.

### 겹침 표

| t1100 | t1082 | 공유 파일 | 소유 |
|---|---|---|---|
| REQ-DHR-016 dispatch record | REQ-FLH-009 dispatch 보존·BOUND 후 방출 / AC-FLH-006 | `internal/factorymsg/store.go` | 레코드 스키마와 상태 전이: t1100. BOUND 전 body 보류 gate: t1082 |
| REQ-DHR-017 멱등 범위 | REQ-FLH-009 "idempotency를 handoff generation에 결합" / AC-FLH-008 | `store.go` (`UNIQUE(sender_session, idem_key)`) | 범위 정의: t1100 한 곳. t1082는 새 범위를 만들지 않는다 |
| REQ-DHR-018 결과 적용 fencing | REQ-FLH-010 tombstone·STALE_* / AC-FLH-007 | `store.go` `verifyPeer`, peers 테이블 | generation 비교 규칙: t1100. tombstone·redirect metadata·재시작 후 생존: t1082 |
| REQ-DHR-020 superseded 메시지 | REQ-FLH-009 SWITCH_PENDING 중 metadata 보존 | `store.go` Claim/Status | superseded 표시와 claim 금지: t1100. handoff 중 보류 dispatch를 새 generation으로 방출하는 경로: t1082 |
| REQ-DHR-021 재할당 | REQ-FLH-011 crash·재시작·abandoned 복구 / AC-FLH-009 | `store.go` | 일반 재할당(새 attempt): t1100. handoff 전이의 resume/finalize/ABANDONED: t1082 |
| REQ-DHR-022 / 023 4조합 | REQ-FLH-014 LIVE Codex↔Codex, Claude↔Codex / AC-FLH-012, 013 | `internal/cli/factory_live_test.go` | 일반 카드 흐름 왕복: t1100. `/cd`·`thread/fork` relocation LIVE: t1082 |
| REQ-DHR-008 ~ 011 codex -w | REQ-FLH-004 develop pin·L1 생성 / AC-FLH-002, 016 | `internal/cli/codex_launcher.go`, `session_worktree.go` `gitWorktreeAddReal` | codex -w anchor·base 대조·동시 writer 거부: t1100. lane handoff용 develop pin과 `BASE_DRIFT`: t1082 |
| (없음) | REQ-FLH-016 ~ 018 launch-pending·rebind·UserPromptSubmit | `store.go` `BindLaunchPending`, `RegisterPeer` | 전부 t1082. t1100은 이 경로를 바꾸지 않는다 |

### 경계 규칙

- generation 필드는 기존 `peers.generation` 하나다. 이 SPEC은 fencing용 두 번째 필드를 만들지 않는다. dispatch record의 assignee generation은 그 값의 복사본이다.
- 멱등 범위는 REQ-DHR-017의 하나다. t1082 design.md §8의 "t1074 idempotency key에 handoff generation을 결합"을 키에 generation을 섞는 뜻으로 읽으면 AC-FLH-008(BOUND 전후 같은 키 → 실행 한 번)과 어긋난다. 이 SPEC은 그 결합을 "BOUND gate + generation fencing"으로 표현하자고 제안하며, 최종 문구는 t1082가 정한다.
- t1082는 이 SPEC의 계약을 소비하는 쪽이다. handoff 상태기계(RESERVED, WT_READY, SWITCH_PENDING, BOUND)는 t1100의 lifecycle 위에 얹히는 gate이며 두 번째 상태기계가 아니다.

## §F 범위 밖

### Out of Scope — M1 정책·템플릿 완결성

- AC-POL-01, AC-TPL-01/02: 중립 정책 배포, 참조 완결성, 재생성 결정론.

### Out of Scope — M2 훅·승인·목표

- AC-HOOK-01/02, AC-GOAL-01: Stop 체인 동등성, compact·permission·interrupt 이벤트, goal 지속.
- 훅 신뢰 상태 `installed-untrusted` / `operational` 도입. AC-MIG-01은 사용자 데이터 보존을 판정하며, 훅이 실제로 실행되는지는 이 SPEC이 증명하지 않는다. 기존 신뢰 사이드카의 divergence 신호는 그대로 둔다.

### Out of Scope — M5 인증·MCP

- AC-WF-01(17개 명령 인증), AC-MCP-01, AC-OBS-01.

### Out of Scope — 다른 실행 프로파일

- Desktop·Web·원격 실행 프로파일, macOS 이외 운영체제 인증. Windows 경로는 기존 테스트를 깨지 않는 수준으로만 유지한다.

### Out of Scope — t1082 소관

- lane worktree relocation(`/cd`, `thread/fork`), handoff tombstone과 redirect metadata, launch-pending handoff(M1), 두 rebind 경로(M2).

### Out of Scope — dispatch record의 나머지 설계 필드

- `scope` / `authority_ref`, `worktree_ref` / `expected_head`, 결과 외 `evidence_refs`. AC-MSG-01·AC-FACT-01 판정에 필요하지 않다. `worktree_ref` / `expected_head`는 t1082가 필요할 때 추가한다.

### Out of Scope — 새 저장소·전송 수단

- 새 브로커, 데몬, 별도 DB, 파일 기반 두 번째 소유 저장소. 레거시 `internal/sessionmsg` 의미 변경.

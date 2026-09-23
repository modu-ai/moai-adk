---
id: SPEC-DUAL-HARNESS-RECOVERY-001
title: "Dual-harness recovery — Codex wiring unwire/rollback, Codex worktree and kanban parity, role permission contract, exactly-once dispatch results, mixed factory card flow"
version: "0.3.1"
status: completed
created: 2026-09-23
updated: 2026-09-24
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
  - SPEC-FACTORY-LANE-WORKTREE-HANDOFF-001  # unmerged; status in-progress on the t1082 branch (v0.5.2, 745ae0e6d); exists only on the t1082 branch/worktree, not in this tree
  - SPEC-CODEX-DUAL-AGENTS-001
  - SPEC-CODEX-LAUNCH-VERB-001
  - SPEC-CODEX-LAUNCHER-001
  - SPEC-CODEX-PARTIAL-WIRING-001
  - SPEC-UPDATE-ADD-CODEX-001
  - SPEC-WORKTREE-DONE-TIER-001  # completed; its L1 refusal in `moai worktree done` is kept unchanged (REQ-DHR-010)
---

# SPEC-DUAL-HARNESS-RECOVERY-001

## HISTORY

| Version | Date | Change |
|---|---|---|
| 0.1.0 | 2026-09-23 | 카드 t1100 plan 초안. 이중 하네스 설계(§19)의 AC-MIG-01, AC-WT-01, AC-AGENT-01, AC-MSG-01, AC-FACT-01 다섯 기준만 다룬다. |
| 0.2.0 | 2026-09-23 | plan-audit iter-1(FAIL 0.76) 결함 D1~D21 수리. 리드 결정 1~4·D6 반영(출처와 한계는 `plan.md` §B). 멱등 범위를 조건부 REQ로 바꾸고 재현 측정 REQ-DHR-025를 추가. 배선 잠금을 codexwiring 소유 잠금으로 교체. 감사 역할의 Codex 예외를 REQ-DHR-015로 한정. |
| 0.3.0 | 2026-09-23 | plan-audit iter-2(FAIL 0.81) 결함 ND1~ND14 수리. `done`의 L1 거부를 기존 계약으로 인정하고 삭제 보호를 실제 삭제 경로에 둠(REQ-DHR-010). ND2는 리드 조정 결정을 §E에 기록. LIVE·측정 증거를 파일 채널로 옮김(1 KiB 분할 회피). 프로필 전환 전제를 코드 판독에 맞게 고치고 Claude 쪽 관리 뿌리 삭제는 범위 밖 발견으로 둠. REQ·AC 수는 그대로(25·23). |
| 0.3.1 | 2026-09-24 | 리드 결정에 따라 sync-audit(`.moai/reports/t1100/sync-audit.md`, 감사 대상 HEAD `dfaba6381`) F1을 반영했다. 감사는 AC-DHR-012가 FAIL, AC-DHR-023이 미충족, REQ-DHR-015의 런타임 조항(read-only 감사자가 반환하고 부모가 판정 파일을 쓴다)이 미충족인데도 status가 `completed`이고 정식 이관 기록이 없다고 지적했다. 원인은 측정으로 확인했다. codex-cli 0.156.1에서 `spawn_agent`로 띄운 하위 에이전트는 부모 세션의 sandbox를 물려받고, 역할 TOML의 `sandbox_mode`는 적용되지 않는다(`design.md` §C.1 정정 문단). 세 항목은 후속 카드 t1143(최상위 read-only 실행 경로)으로 정식 이관한다. 이관 방식은 t1082 선례(SPEC-FACTORY-LANE-WORKTREE-HANDOFF-001 0.5.10)를 따른다. 요구사항 SHALL 문구, AC 본문, 판정식(jq), 기대값은 한 글자도 바꾸지 않고 제자리에 두며, t1143이 그대로 이어받는다. acceptance.md의 두 AC 주석, 매핑표, §C 집계표, §D 완료 정의와 이 문서의 REQ-DHR-015 한계 주석, §D 추적표, § Out of Scope — AC-DHR-012·023과 REQ-DHR-015 런타임 조항에 이관 표시를 더했다. status는 되돌리지 않는다. 나머지 범위에 대해서는 `completed`가 유지되며, 이 SPEC은 이관한 세 항목을 충족한다고 주장하지 않는다. |

## §A 배경과 목적

`reports/moai-dual-harness-full-design-20260922.md`(설계 제안, 구현 기록 아님)는 Claude Code와 Codex CLI의 동등 지원을 14개 인수 기준으로 정의한다. 이 SPEC은 그중 상태·복구(M4)와 실행·협업(M3)에 속하는 다섯 기준만 구현 가능한 요구사항으로 옮긴다.

이 트리에서 코드를 읽고 확인한 출발점은 다음과 같다. 자세한 근거와 줄 번호는 `research.md`에 있다.

- `internal/codexwiring`은 설치만 한다. 제거 경로가 없고, 신뢰 사이드카(`.moai/state/codex-wiring.json`)는 전체 파일 해시 두 개만 담는다. 원자적 쓰기는 temp+rename뿐이며 저널, 잠금, rename 직전 해시 재확인, 기록 후 대조가 없다.
- `hooks.json`은 매번 문서 전체를 다시 직렬화한다(`hooks.go:122`). 따라서 이 파일에 대한 보존 보장은 바이트가 아니라 구조 단위로만 할 수 있다. `config.toml`은 텍스트 덧붙이기·줄 삽입으로만 바뀐다. MoAI가 만든 영역 밖에서 바뀌는 바이트는 하나뿐이다. 원본이 줄바꿈으로 끝나지 않으면 덧붙이기(`appendSection`, `configtoml.go:225-227`)와 줄 삽입(`EnsureStatusLine` 분기 ii, `:130-141`)이 파일 끝에 줄바꿈 한 바이트를 더한다. 이 바이트는 MoAI 영역에 들지 않고 unwire가 되돌리지 않는다(`design.md` §A.1 `region`).
- claude 프로필 배포는 `.codex/`를 숨기지만(`harness_fs.go` `hideCodex`), 이미 있는 `.codex/*`를 치우지 않는다. `both → claude` 전환 후 배선 파일과 에이전트 TOML이 고아로 남는다.
- `moai codex -w`는 새 트리를 만들거나 기존 트리에 들어가지만, 동시 writer 검사나 소유 표식을 남기지 않는다. `moai codex -k`는 없다(`-f`는 있다).
- Codex 역할 TOML은 작업 공간 단위 `sandbox_mode`와 역할별 MCP 서버 부여(`[mcp_servers.<name>]` 테이블)를 표현한다. 한 MCP 서버 안의 도구 단위 제한, 경로 단위 쓰기 제한은 표현하지 못한다.
- `internal/factorymsg`는 전달을 at-least-once로 보장하고 멱등 키로 송신 중복을 막는다. 멱등 범위가 송신자 세션 UUID라서, 송신자가 재시작하면 같은 키도 새 메시지 행이 될 수 있다(코드 판독, 미측정). 결과를 한 번만 반영하는 작업 레코드는 없다.
- 혼합 팩토리 4조합 테스트는 `MOAI_FACTORY_LIVE=1` 없이는 모두 `SKIP`이다. 이때도 패키지 결과는 `ok`로 출력된다(plan 단계에서 관측).

## §B 범위

다섯 설계 기준을 다음 REQ 묶음으로 옮긴다.

| 설계 기준 | 이 SPEC의 REQ | 이 SPEC의 AC |
|---|---|---|
| AC-MIG-01 | REQ-DHR-001 ~ REQ-DHR-007 | AC-DHR-001 ~ AC-DHR-005, AC-DHR-021, AC-DHR-022 |
| AC-WT-01 | REQ-DHR-008 ~ REQ-DHR-012 | AC-DHR-006 ~ AC-DHR-009 |
| AC-AGENT-01 | REQ-DHR-013 ~ REQ-DHR-015 | AC-DHR-010 ~ AC-DHR-013, AC-DHR-023 |
| AC-MSG-01 | REQ-DHR-016 ~ REQ-DHR-021, REQ-DHR-025 | AC-DHR-014 ~ AC-DHR-016, AC-DHR-020 |
| AC-FACT-01 | REQ-DHR-022 ~ REQ-DHR-024 | AC-DHR-017 ~ AC-DHR-019 |

### 이전 결정과의 관계

- SPEC-FACTORY-MIXED-HOOK-001 REQ-FMH-006은 브로커가 at-least-once 전달을 유지하고 exactly-once 실행을 주장하지 않는다고 정했다. 이 SPEC은 그 결정을 뒤집지 않는다. "한 번만 반영"은 메시지 전달이 아니라 작업 레코드(dispatch record)에 결과를 적용하는 단계에서 보장한다.
- SPEC-CODEX-WIRING-001 REQ-CW-005는 기존 `[mcp_servers.moai]` 테이블과 `status_line` 키를 사용자 소유로 보고 다시 쓰지 않는다. 같은 생성기가 `hooks.json`에서는 `moai hook ` 접두사나 `.codex/hooks/moai/` 네임스페이스 명령을 가진 handler를 이미 있더라도 MoAI 소유로 보고 매번 교체한다(`hooks.go:138-141`, `:167`). 이 SPEC의 소유 분류(REQ-DHR-001)는 이 두 기존 동작을 그대로 따른다. handler는 명령 식별자로 MoAI 소유, config 부분은 MoAI가 이번 기록 아래서 직접 만든 것만 MoAI 소유다.
- SPEC-CODEX-WIRING-001 REQ-CW-012는 배선 생성기가 `.codex/agents/**`를 건드리지 못하게 한다. 이 SPEC도 에이전트 TOML을 지우지 않는다. 프로필 전환으로 배포되지 않게 된 템플릿 파일은 보고만 한다(REQ-DHR-007).
- 설계 ADR-02(기존 상태 저장 API 재사용)에 따라 소유 기록은 `internal/manifest`, 작업 레코드는 기존 factorymsg SQLite 브로커를 확장한다. 새 저장소나 데몬은 만들지 않는다. 배선 잠금은 사이드카와 같은 `.moai/state/` 아래의 파일 하나다.

## §C 요구사항 (GEARS)

### REQ-DHR-001 — Codex wiring ownership provenance

The Codex wiring generator SHALL record, for every file and every file part it writes, a provenance entry through `internal/manifest` naming the path, the part kind (`whole-file`, `hook-handler`, `json-key`, `toml-table`, or `toml-key`), the part key, the origin (`created`, `preexisting`, or `unknown`), and the hash of the bytes MoAI wrote. A hook handler whose command starts with `moai hook ` or lies inside the `.codex/hooks/moai/` namespace SHALL be recorded as `created`, because the generator replaces such handlers on every run. The top-level `description` key of `hooks.json` SHALL be recorded as a `json-key` part whose origin is `created` only when MoAI added it. A `[tui]` table that MoAI appended as a whole SHALL be recorded as a `toml-table` part, and a `status_line` assignment MoAI inserted into a user `[tui]` table SHALL be recorded as a `toml-key` part. When a part other than a MoAI-identified handler already exists and the project carries no earlier MoAI wiring evidence (no manifest part record, no trust sidecar, no wiring journal), the generator SHALL record it as `preexisting`. When such a part exists and earlier wiring evidence exists but no part record does, the generator SHALL record it as `unknown`. A `preexisting` or `unknown` part SHALL never be promoted to `created` by a later run.

### REQ-DHR-002 — Journaled per-file write under the wiring lock

When the Codex wiring generator, the unwire step, or the recovery step changes a wiring file, it SHALL hold a wiring lock owned by the wiring package, distinct from the `moai update` lock and acquirable whether or not the caller already holds the update lock, and while holding it SHALL append a journal entry naming the path, the operation, the pre-change hash, the intended post-change hash, the intended provenance record, and the temporary file name; stage the full content in that temporary file in the target directory; re-read the target and compare its hash with the pre-change hash immediately before rename; rename; read the target back and compare its hash with the intended post-change hash; apply the intended provenance record to the manifest; and only then mark the journal entry complete. Where both locks are held, the wiring lock SHALL be acquired after the update lock. When the wiring lock is held by a live or undetermined other owner, the step SHALL change no wiring file and SHALL return a distinct lock-held outcome that its caller reports. The generator SHALL NOT claim that a multi-file wiring pass is atomic as a whole.

### REQ-DHR-003 — Refusal on concurrent or user modification

When the target hash re-read immediately before rename differs from the pre-change hash, the Codex wiring step SHALL NOT rename, SHALL leave the target byte-identical, SHALL remove its temporary file, SHALL record the journal entry as `conflict`, and SHALL report the path and both hashes as a conflict with a non-zero outcome for that file.

### REQ-DHR-004 — Interrupted-change recovery and its entry points

When `moai tool enable codex`, `moai tool disable codex`, or the update-path wiring refresh starts while an incomplete wiring journal entry exists, the recovery step SHALL run under the wiring lock before any new write and SHALL classify each entry by the target's current state as `completed` (equals the intended post-change state), `not-applied` (equals the pre-change state), or `diverged` (equals neither); SHALL apply the entry's recorded provenance and mark `completed` entries complete; SHALL discard the temporary file of `not-applied` entries without touching the target; SHALL leave `diverged` targets untouched and reported while deleting the temporary file each `diverged` entry references; and SHALL delete a `.codexwiring-*` temporary file in a wiring target directory that no journal entry references. Recovery SHALL NOT overwrite a file whose current state is neither the pre-change nor the intended post-change state. `moai doctor` SHALL report incomplete journal entries and unreferenced temporary files with the command that recovers them, and SHALL NOT write, rename, or delete any file.

### REQ-DHR-005 — Owned-part-only unwire through `moai tool disable codex`

When the operator runs `moai tool disable codex`, the unwire step SHALL remove only parts whose origin is `created` and whose current content hashes to the recorded hash, through the REQ-DHR-002 write path. For `config.toml` it SHALL produce exactly the pre-unwire bytes with each removed part's recorded byte region, including the separator MoAI inserted with it, cut out, and SHALL leave every other byte unchanged. For `hooks.json`, which the generator re-serializes, it SHALL preserve every non-MoAI top-level key value, every user hook entry's matcher, and every user handler as parsed JSON values, SHALL preserve the array order of the entries under each event and of the handlers within each entry, SHALL NOT compare object key order (the generator re-serializes with sorted keys), and SHALL NOT claim byte preservation. It SHALL delete a whole file only when the file's provenance is `whole-file` with origin `created` and the file hash matches the record; otherwise it SHALL fall back to part-level removal.

### REQ-DHR-006 — No write through unverified ownership or symbolic links

The Codex wiring write, unwire, and recovery steps SHALL NOT delete or modify a file or part on hash evidence alone. When a part lacks a provenance record, is recorded `preexisting` or `unknown`, has a hash mismatch, or when the target path is itself a symbolic link or any directory between the project root and the target resolves through a symbolic link to a location outside the project root, the step SHALL leave it untouched and SHALL report its path and the reason (`no-provenance`, `user-owned`, `unknown-origin`, `modified`, or `symlink-boundary`). The check SHALL use the link itself (`Lstat`), not the link's destination.

### REQ-DHR-007 — Harness profile transitions report and preserve

When the configured harness profile transitions (`claude → both`, `gpt → both`, `both → claude`, `both → gpt`, `gpt → claude`, or an older binary's deployment is refreshed by a newer one), the deployment SHALL preserve every user-owned file and part under `.codex/` and every Codex wiring file and part. When the Codex wiring files become orphaned by such a transition, `moai update` SHALL report each orphaned file and the command that removes it (`moai tool disable codex`) and SHALL NOT remove or rewrite it. When the target profile no longer deploys a `.codex/` template path that the manifest records (for example `.codex/agents/moai/*.toml` after `both → claude`), the deployment SHALL report it and SHALL NOT delete it. This requirement SHALL NOT be read as governing the Claude-side managed roots that the existing update cleanup stage removes before deployment on every profile; that behaviour is unchanged by this SPEC (§F).

### REQ-DHR-008 — Codex worktree anchor

When `moai codex -w` launches a Codex session into a new or existing card worktree, the launcher SHALL place a git worktree lock on that tree whose reason carries `pid <n>` before the Codex process starts, so the existing lock-aware anchor decision recognizes the tree as anchored for the session's lifetime. On POSIX, where the launcher replaces itself with the Codex process, `<n>` SHALL be the launcher's own pid. On Windows, where the launcher starts the Codex child and waits for it, `<n>` SHALL be the waiting launcher's pid. When the tree already carries a lock whose recorded pid is confirmed dead, the launcher SHALL replace it only while holding a per-tree replacement guard that it creates exclusively, SHALL re-read the lock reason under that guard, and SHALL refuse the launch with a non-zero exit when the reason changed, when the guard cannot be created, or when `git worktree lock` fails; it SHALL NOT force a lock and SHALL NOT replace a lock whose pid is live or whose liveness is undetermined.

### REQ-DHR-009 — Concurrent writer rejection

When `moai codex -w` or `moai cc -w` targets an existing worktree that the existing anchor decision reports as anchored by a live or undetermined owner other than the caller, the launcher SHALL refuse to launch with a non-zero exit and a diagnostic naming the anchor source and holder, and SHALL NOT modify the tree, its lock, or its branch.

### REQ-DHR-010 — Disposal protection for Codex-created trees on the paths that can delete them

`moai codex -w <name>` creates its tree under `<project root>/.claude/worktrees/`, the L1 tier. `moai worktree done` SHALL keep refusing such a tree with `L1_SESSION_WORKTREE` whether or not `--force` is given, exactly as it refuses a Claude-created L1 tree under SPEC-WORKTREE-DONE-TIER-001, and this SPEC SHALL NOT change that refusal. `moai worktree clean --stale`, `moai worktree remove`, and the PR-merge cleanup, which are the paths that can delete an L1 tree, SHALL decide anchoring through the shared lock-aware anchor decision, so that a git worktree lock placed under REQ-DHR-008 whose pid is live or undetermined counts as an anchor. `moai worktree clean --stale` and the PR-merge cleanup SHALL keep a Codex-created tree that is anchored, holds uncommitted changes, or holds commits that their existing no-work-lost checks treat as not integrated, SHALL name the reason, and SHALL give the same result as for a Claude-created tree in the same state. `moai worktree remove` SHALL refuse an anchored Codex-created tree unless `--force` is given, SHALL name the anchor source, and SHALL keep its existing explicit-removal semantics for integration state.

### REQ-DHR-011 — Codex worktree creation base verification

When `moai codex -w` creates a new worktree, the launcher SHALL resolve the base through the existing base resolution, SHALL verify after creation that the new tree's HEAD equals the resolved base commit, and when they differ SHALL refuse to launch and leave the tree for inspection without deleting it. The launcher SHALL NOT change the base resolution policy.

### REQ-DHR-012 — Codex kanban entry parity

Where the operator passes `-k` or `--kanban` to `moai codex`, the launcher SHALL accept the same lead and companion entry shapes as `moai cc -k`, SHALL publish the same kanban launch facts with backend `codex`, SHALL claim or resolve the session name with the same rules, and SHALL NOT forward the kanban tokens to the Codex child process. When an unsupported shape is passed, the launcher SHALL exit non-zero with a usage diagnostic rather than degrade to a plain launch.

### REQ-DHR-013 — Role permission contract

The agent emitter SHALL carry a permission contract for each of the 12 Codex roles over the axes `sandbox`, `write-path-scope`, `shell`, `mcp-server`, `mcp-tool`, `subagent`, and `web`, and SHALL map every axis on which the contract requires a restriction to exactly one of `enforced` (naming the Codex field that enforces it and a `measured` or `documented` basis) or `UNSUPPORTED` (naming the host-expressivity reason and a `measured`, `documented`, or `unmeasured` basis). An axis with an `unmeasured` basis SHALL NOT be mapped `enforced`. An axis that carries more than one restriction kind (for `mcp-server`: granting a server and denying a server) SHALL be mapped once per restriction kind, each mapping being exactly one of the two values. The basis of an `enforced` mapping SHALL state what was observed about the field (for example that Codex accepts the value); a claim that the field blocks the restricted action at runtime SHALL rest only on the runtime evidence of REQ-DHR-014. When a role's contract requires a restriction that is neither enforced nor declared `UNSUPPORTED`, the emitter SHALL fail. The emitter SHALL NOT emit a `sandbox_mode` broader than the role's contract states.

### REQ-DHR-014 — Unsupported axes are never reported as passing

The Codex role verification SHALL report every `UNSUPPORTED` axis as `UNSUPPORTED` and SHALL NOT count it as PASS. When runtime role loading and read-only enforcement are verified against a real Codex binary, that verification SHALL be recorded as a separate evidence item naming the Codex version, the invocation count, and positive evidence of each claim; an unexecuted runtime verification SHALL be reported as `NOT_RUN`, and one stopped by its invocation budget SHALL be reported as `ABORTED`.

### REQ-DHR-015 — Codex audit roles are read-only and return their verdict text

Where `plan-auditor` or `sync-auditor` is emitted as a Codex role, the role's contract SHALL state `sandbox: read-only` and the emitter SHALL emit `sandbox_mode = "read-only"` for it together with a Codex-only instruction that the role returns its complete verdict or report text instead of writing a file. When a Codex audit role returns, the parent lane orchestrator SHALL write the audit verdict or report file with exactly the returned text. This exception to the contract "the auditor writes its own verdict file" SHALL apply only to the Codex path; the Claude agent definitions, their emitted Claude copies, and the Claude audit workflow SHALL remain unchanged.

> **알려진 한계 (측정, codex-cli 0.156.1 — REQ-DHR-013·015 공통).** `spawn_agent`로 띄운 하위 에이전트는 부모 세션의 sandbox를 물려받으며, 역할 TOML의 `sandbox_mode`는 그 sandbox를 좁히지도 넓히지도 못한다. `plan-auditor`와 `sync-auditor`는 `read-only` TOML로 방출되었는데도 `workspace-write`로 실행되어 쓰기에 성공했다(`design.md` §C.1 정정 문단, AC-DHR-012). 따라서 REQ-DHR-013의 `sandbox` 축 `enforced` 매핑과 REQ-DHR-015의 read-only 감사자는 이 경로에서 런타임에 충족되지 않는다. 두 요구사항의 SHALL 문구는 그대로 유지하며, 그 충족(최상위 `codex exec -s read-only` 실행 경로)은 후속 카드 t1143의 몫이다.
>
> **정식 이관 (0.3.1, sync-audit F1 리드 결정).** REQ-DHR-015의 런타임 조항 — 감사 역할이 read-only로 실행되어 판정문을 반환하고 부모 lane 오케스트레이터가 그 원문으로 판정 파일을 쓴다는 부분 — 은 카드 t1143으로 정식 이관한다. 이 SPEC 범위에서는 충족되지 않았고, 이 SPEC은 충족을 주장하지 않는다. 이 SPEC 안에서 REQ-DHR-015는 방출 쪽 계약(AC-DHR-013)으로만 검증된다. 위 SHALL 문구는 바꾸지 않으며 t1143이 그대로 이어받는다. 함께 이관하는 AC는 AC-DHR-012와 AC-DHR-023이다(§ Out of Scope — AC-DHR-012·023과 REQ-DHR-015 런타임 조항).

### REQ-DHR-016 — Dispatch record

The factory broker SHALL store one dispatch record per run and dispatch identifier carrying the dispatch ID, card ID, assignee lane slot, attempt number, the assignee generation used as the fencing token, the result digest, the lifecycle state, and a result reference. The lifecycle states SHALL be `assigned`, `delivered`, `started`, `result_recorded`, `integrated`, and `abandoned`, recorded as distinct states. The idempotency keys of the assignment message and the result message SHALL be scoped to the dispatch ID and attempt, so a reassignment to another lane or attempt uses a new key. A message arrival or receipt SHALL NOT move a dispatch past `delivered`.

### REQ-DHR-017 — Idempotency scope decided by measurement

Where the measurement of REQ-DHR-025 recorded `reproduced`, the factory broker SHALL scope message idempotency to the project key, run ID, sender lane slot, and idempotency key, independent of the sender's session UUID and generation, SHALL migrate existing rows without loss, and SHALL return the original message for a same-scope retry after a sender restart. Where that measurement recorded `not-reproduced`, the broker SHALL keep the existing `UNIQUE(sender_session, idem_key)` scope and SHALL NOT migrate the message schema. In both branches, when the same scope carries a different recipient (a different recipient session UUID or recipient generation), kind, task reference, correlation ID, or payload, the broker SHALL reject it without mutation.

### REQ-DHR-018 — Exactly-once result application under fencing

When the receiver of a result message applies it to a dispatch record, the factory broker SHALL decide in one transaction and in this order: (1) no record for the dispatch ID → `unknown`; (2) the reported attempt differs from the record's current attempt → `stale`; (3) the reporter's lane differs from the assignee lane → `stale`; (4) the reporter's generation is lower than the lane's current generation → `stale`; (5) the record already holds a result for this attempt → `duplicate` when the result digest is equal, `collision` when it differs; (6) the reporter's generation differs from the assignee generation → `stale`; (7) the record is not in `started` → `invalid-state`; (8) otherwise record the result, set `result_recorded`, and return `accepted`. Every outcome other than `accepted` SHALL leave the record unchanged. The broker message layer SHALL remain at-least-once.

### REQ-DHR-019 — Result persisted before receipt

When the receiver of a message that carries a result processes it, the receiver SHALL apply the result to the dispatch record under REQ-DHR-018 before acknowledging the message receipt, so that a crash between the application and the receipt leads to a redelivery that REQ-DHR-018 resolves as `duplicate`.

### REQ-DHR-020 — Superseded-generation messages

When a lane's generation advances while messages addressed to its previous generation are pending or claimed, the factory broker SHALL NOT let the superseded generation claim, read, dispose, or acknowledge them and SHALL report them as `superseded` in the broker status rather than as pending. The broker SHALL NOT itself release a superseded message body to the new generation. Dispatch authority SHALL move to a new generation only through a new attempt (REQ-DHR-021) or through a same-attempt regrant that updates the record's assignee generation in the same transaction as the caller's own state change; this SPEC provides the regrant operation and leaves its admission conditions to the SPEC that calls it.

### REQ-DHR-021 — Reassignment fences the previous attempt

When the lead reassigns a dispatch, the factory broker SHALL require that the previous assignee's owner is confirmed not live or that the lead explicitly revokes it, SHALL increment the attempt, SHALL record the new assignee lane and generation, and SHALL thereafter treat any result from the previous attempt as `stale` under REQ-DHR-018.

### REQ-DHR-022 — Deterministic mixed-backend card flow

The factory test suite SHALL drive each of the four backend combinations (Claude lead to Claude worker, Codex to Codex, Claude to Codex, Codex to Claude) through `assigned → delivered → started → result_recorded → integrated` against the real broker store without a model, including a worker interruption after `started`, a reassignment to a new attempt, and a late result from the interrupted attempt, and SHALL assert that exactly one result is applied per dispatch.

### REQ-DHR-023 — LIVE mixed factory card flow evidence

When LIVE verification runs, each of the four backend combinations SHALL complete the card flow of REQ-DHR-022 in real separate CLI and model contexts inside an isolated temporary repository and `MOAI_HOME`, within at most 8 model invocations and 900 seconds per combination, with cleanup of every spawned process registered before the first spawn. When a combination would exceed its invocation budget (a 9th model invocation would be needed) or exceeds its 900-second time budget, the case SHALL stop before that invocation, clean up, and report `ABORTED`; a combination that finishes within its budget, including one that uses exactly 8 invocations, SHALL NOT be reported as `ABORTED`. Each case SHALL emit its own evidence record. LIVE evidence SHALL be recorded separately from the deterministic evidence and SHALL NOT be claimed to run in CI unless a CI workflow that sets the LIVE gate is observed.

### REQ-DHR-024 — SKIP and NOT_RUN are not PASS

Every verification command in this SPEC SHALL count `pass` events for the exact named tests and SHALL fail when any named test reports `skip` or `fail`, when the expected pass count is not met, or when output contains `NOT_RUN` or `ABORTED`. A package-level `ok` line SHALL NOT be accepted as evidence.

### REQ-DHR-025 — Idempotency-scope reproduction measurement

When the run phase starts, before any change to the factorymsg schema, the reproduction test SHALL register a sender lane, send a message with an idempotency key, re-register the same lane slot with a new session UUID so that its generation advances, send an identical message with the same idempotency key, and record `reproduced` when a second message row exists that the recipient can claim, or `not-reproduced` when exactly one such row exists and the second send returned that row's message ID, together with the unique constraint text the store was opened with, the two sender session UUIDs and generations, the row count, the number of distinct message IDs the recipient claimed, the second send's result, and the claim result of a control message sent with a fresh key after the re-registration. When the first send left no row, the re-registration did not yield a different session UUID with a higher generation, the second send returned an error, or the control message could not be claimed, the test SHALL record `NOT_RUN` instead of either outcome, and neither branch of REQ-DHR-017 SHALL be selected from that run.

## §D 요구사항 ↔ 인수 기준 추적

| Requirement anchor | Acceptance criteria |
|---|---|
| § REQ-DHR-001 | AC-DHR-003, AC-DHR-004 |
| § REQ-DHR-002 | AC-DHR-001, AC-DHR-022 |
| § REQ-DHR-003 | AC-DHR-001 |
| § REQ-DHR-004 | AC-DHR-002, AC-DHR-021 |
| § REQ-DHR-005 | AC-DHR-003 |
| § REQ-DHR-006 | AC-DHR-001, AC-DHR-004 |
| § REQ-DHR-007 | AC-DHR-005 |
| § REQ-DHR-008 | AC-DHR-006 |
| § REQ-DHR-009 | AC-DHR-007 |
| § REQ-DHR-010 | AC-DHR-008 |
| § REQ-DHR-011 | AC-DHR-006 |
| § REQ-DHR-012 | AC-DHR-009 |
| § REQ-DHR-013 | AC-DHR-010 |
| § REQ-DHR-014 | AC-DHR-011, AC-DHR-012 (`FAIL → t1143 이관`) |
| § REQ-DHR-015 | AC-DHR-013, AC-DHR-023 (`미충족 → t1143 이관`) — 런타임 조항은 t1143까지 미충족 |
| § REQ-DHR-016 | AC-DHR-014 |
| § REQ-DHR-017 | AC-DHR-014, AC-DHR-015 |
| § REQ-DHR-018 | AC-DHR-014, AC-DHR-015, AC-DHR-016 |
| § REQ-DHR-019 | AC-DHR-014 |
| § REQ-DHR-020 | AC-DHR-015 |
| § REQ-DHR-021 | AC-DHR-016 |
| § REQ-DHR-022 | AC-DHR-017 |
| § REQ-DHR-023 | AC-DHR-018 |
| § REQ-DHR-024 | AC-DHR-019 |
| § REQ-DHR-025 | AC-DHR-020 |

이 표와 `acceptance.md` §A 매핑 표가 REQ·AC 연속성의 근거다. `moai spec lint`는 REQ 번호의 빈칸과, 존재하지 않는 REQ를 가리키는 AC·표 행을 보고하지 않는다(plan-audit iter-2 변이 m1·m2 측정). 그래서 두 표는 손으로 대조해 일치시킨다.

## §E t1082 경계 (SPEC-FACTORY-LANE-WORKTREE-HANDOFF-001)

t1082의 SPEC은 이 트리에 없다. `.claude/worktrees/t1082/.moai/specs/SPEC-FACTORY-LANE-WORKTREE-HANDOFF-001/`에만 있는 미병합 SPEC이며(이 개정 시점 t1082 브랜치에서 status in-progress, v0.5.2 `745ae0e6d`), 아래 인용은 이번 개정 시점에 그 경로를 읽기 전용으로 읽은 것이다. t1082의 열린 결정 M1(launch-pending 중 handoff)과 M2(두 rebind 경로)는 t1082의 것이고, 이 SPEC은 그 둘을 결정하지 않는다.

### 겹침 표

| t1100 | t1082 | 공유 파일 | 소유 |
|---|---|---|---|
| REQ-DHR-016 dispatch record | REQ-FLH-009 dispatch 보존·BOUND 후 방출 / AC-FLH-006 | `internal/factorymsg/store.go` | 레코드 스키마와 상태 전이: t1100. BOUND 전 body 보류 gate: t1082 |
| REQ-DHR-017, 025 멱등 범위 | REQ-FLH-009 "current t1074 schema's idempotency key unchanged" / AC-FLH-008(같은 attempt 재부여 뒤 재전송은 아래 경계 규칙의 리드 조정) | `store.go` (`UNIQUE(sender_session,idem_key)`) | 범위 결정: t1100(측정 결과에 따른 조건부). 이관 시 t1082 영향은 리드가 조정 |
| REQ-DHR-018 결과 적용 fencing | REQ-FLH-010 tombstone·STALE_* / AC-FLH-007 | `store.go` `verifyPeer`, peers 테이블 | 결과 적용 판정 순서: t1100. tombstone·redirect metadata·재시작 후 생존: t1082 |
| REQ-DHR-020 superseded 메시지 | REQ-FLH-009 SWITCH_PENDING 중 metadata 보존, BOUND 후 방출 | `store.go` Claim/Status | superseded 표시·claim 금지·같은 attempt 재부여 연산: t1100. 재부여를 호출할 조건(BOUND): t1082 |
| REQ-DHR-021 재할당 | REQ-FLH-011 crash·재시작·abandoned 복구 / AC-FLH-009 | `store.go` | 일반 재할당(새 attempt): t1100. handoff 전이의 resume/finalize/ABANDONED: t1082 |
| REQ-DHR-022 / 023 4조합 | REQ-FLH-014 LIVE Codex↔Codex, Claude↔Codex / AC-FLH-012, 013 | `internal/cli/factory_live_test.go` | 일반 카드 흐름 왕복: t1100. `/cd`·`thread/fork` relocation LIVE: t1082 |
| REQ-DHR-008 ~ 011 codex -w | REQ-FLH-004 develop pin·L1 생성 / AC-FLH-002, 016 | `internal/cli/codex_launcher.go`, `session_worktree.go` `gitWorktreeAddReal` | codex -w anchor·base 대조·동시 writer 거부: t1100. lane handoff용 develop pin과 `BASE_DRIFT`: t1082 |
| (없음) | REQ-FLH-016 ~ 018 launch-pending·rebind·UserPromptSubmit | `store.go` `BindLaunchPending`, `RegisterPeer` | 전부 t1082. t1100은 이 경로를 바꾸지 않는다 |

### 경계 규칙

- generation 필드는 기존 `peers.generation` 하나다. 이 SPEC은 fencing용 두 번째 필드를 만들지 않는다. dispatch record의 assignee generation은 그 값의 복사본이다.
- REQ-DHR-018의 판정 (4)는 "이전 generation의 재시도는 stale"이라는 t1082 AC-FLH-008의 성질(t1082 acceptance.md의 AC-FLH-008 본문)과 같은 방향이다. 현재 generation의 재시도는 (5)에서 `duplicate`가 된다.
- 대화형 `/cd` handoff는 같은 프로세스를 유지하므로(t1082 REQ-FLH-006) REQ-DHR-021의 "비생존 확인"이 성립하지 않는다. 그래서 REQ-DHR-020은 새 attempt 외에 같은 attempt 재부여 연산을 제공한다. t1082가 BOUND 방출을 이 연산으로 표현할지, 명시 철회 후 새 attempt로 표현할지는 t1082가 정한다.
- 멱등 범위: t1082 spec.md REQ-FLH-009(이 개정 시점 본문)는 "detecting duplicates by the current t1074 schema's idempotency key unchanged"라고 적고, t1082 design.md(같은 시점, 204행)는 "키의 기준은 t1100이 소유하며 이 SPEC은 현행 스키마를 따른다"고 적는다. REQ-DHR-025 측정이 `reproduced`여서 범위를 옮기면 이 문구와 맞춰야 한다. 이 SPEC은 t1082의 문구를 정하지 않으며, 맞출 필요가 생기면 리드가 t1082와 조정한다(리드에게 전달할 제안: REQ-FLH-009의 "unchanged"를 "t1100이 정한 범위"로 바꾸는 것).
- 같은 attempt 재부여 뒤 같은 키 재전송(plan-audit iter-2 ND2): 할당 키는 `dispatch:<dispatch_id>:<attempt>`이고 재부여는 attempt를 바꾸지 않는다. 현행 메시지 층은 같은 송신 세션이 같은 키로 수신 세션이나 수신 generation이 다른 요청을 보내면 `idempotency key collision with different request`로 거부한다(`store.go:624-626`). 이 개정 시점의 t1082 design.md 204-205행과 AC-FLH-008은 BOUND 전후 같은 키와 현재 generation 재전송의 `duplicate` 처분을 기대했다. **리드 조정 결정(2026-09-23)**: t1082가 자기 계약을 좁힌다. BOUND 뒤 재전송은 새 멱등 키를 쓰고, 같은 키의 `duplicate`는 같은 generation 안에서만 요구된다. 이렇게 좁힌 계약은 현행 `UNIQUE(sender_session, idem_key)` 동작과 충돌하지 않는다. 이 결정은 리드의 조정이며 이 SPEC이 정한 것이 아니다. t1082 문서의 반영은 t1082 소관이다. 이 SPEC은 이 경우를 위해 멱등 범위를 바꾸지 않으며, 이 경우에 한한 lane slot 기준 범위는 후속 후보로만 적는다(§F).
- This is separate from operator decision 3's conditional migration branch, which remains in force. 곧 ND2 조정은 운영자 결정 3의 조건부 이관 분기(REQ-DHR-017, REQ-DHR-025)를 바꾸지 않으며, 그 분기는 그대로 유효하다.
- t1082는 이 SPEC의 계약을 소비하는 쪽이다. handoff 상태기계(RESERVED, WT_READY, SWITCH_PENDING, BOUND)는 t1100의 lifecycle 위에 얹히는 gate로 제안되며, 두 번째 상태기계로 만들지 않는 것은 t1082의 수락을 전제로 한다.

### 리드 조정 항목

이 SPEC이 정하지 않고 리드에게 넘기는 사항이다.

1. ① 해소 — t1082 `745ae0e6d` (v0.5.2, t1082 브랜치에서 status in-progress): REQ-FLH-009(t1082 spec.md:103)와 AC-FLH-008(t1082 acceptance.md:152-158)이 같은 recipient generation 안의 계약(BOUND 뒤 재전송은 새 키, 같은 키 `duplicate`는 같은 recipient generation 안)을 담는다. 칸반 리드가 해당 커밋에서 확인했다. 새 키도 dispatch ID와 attempt를 담아야 REQ-DHR-016의 키 범위가 지켜진다. 새 키에 붙일 구분자는 t1082가 정한다.
2. 분기 A(`reproduced`)가 선택되면 t1082의 "unchanged" 문구 조정(위 멱등 범위 규칙). 운영자 결정 3에 따른 기존 항목이며 ND2와 별개다.

## §F 범위 밖

### Out of Scope — M1 정책·템플릿 완결성

- AC-POL-01, AC-TPL-01/02: 중립 정책 배포, 참조 완결성, 재생성 결정론.
- `AGENTS.md.tmpl`의 worktree-entry 행("resolves an existing tree and never creates one")과 실제 `moai codex -w` 생성 동작의 불일치. 이번 조사에서 관측했지만 이 카드에서 고치지 않는다(`research.md` §C).

### Out of Scope — M2 훅·승인·목표

- AC-HOOK-01/02, AC-GOAL-01: Stop 체인 동등성, compact·permission·interrupt 이벤트, goal 지속.
- 훅 신뢰 상태 `installed-untrusted` / `operational` 도입. AC-MIG-01은 사용자 데이터 보존을 판정하며, 훅이 실제로 실행되는지는 이 SPEC이 증명하지 않는다. 기존 신뢰 사이드카의 divergence 신호는 그대로 둔다.

### Out of Scope — M5 인증·MCP

- AC-WF-01(17개 명령 인증), AC-MCP-01, AC-OBS-01.
- MCP 서버 안의 도구 단위 제한 구현. 이 SPEC은 표현 불가를 `UNSUPPORTED`로 기록만 한다.

### Out of Scope — 다른 실행 프로파일

- Desktop·Web·원격 실행 프로파일, macOS 이외 운영체제 인증. Windows 동작은 REQ-DHR-008에 적었지만 이 SPEC은 Windows 실측을 요구하지 않는다. 로컬 검증은 darwin이며 Windows 결과는 CI 매트릭스에서 관측될 때만 인용한다.

### Out of Scope — 템플릿 파일 삭제와 구버전 config 부분 제거

- 프로필 전환 때 배포되지 않게 된 템플릿 관리 파일의 삭제. 보고만 한다.
- provenance 기록 이전에 설치된 `[mcp_servers.moai]`·`status_line`의 자동 제거. `unknown`으로 기록되어 unwire가 남기고 보고한다. 운영자가 손으로 지운다.

### Out of Scope — Claude 쪽 관리 뿌리의 프로필 무관 삭제 (plan-audit iter-2 ND6 발견)

- update의 관리 경로 정리 단계는 프로필과 무관하게 `CleanMoaiManagedPaths`를 호출하고(`update_template_sync.go:405-424`), 이 함수는 `.claude/settings.json`, `.claude/{commands,agents,hooks}/moai`, `.claude/skills/moai*`, `.claude/rules/moai`, `.claude/output-styles/moai`를 지운다(`deploy.go:56-85`). gpt 프로필 배포는 `.claude/**`를 숨기므로(`internal/template/harness_fs.go:112`) 코드상으로는 `both → gpt` 전환 뒤 이 경로들이 다시 배포되지 않는다. 코드 판독이며 측정하지 않았다(가설).
- 이 카드의 범위는 codexwiring이므로 여기서 고치지 않는다. 후속 카드 후보: "gpt 프로필 update에서 Claude 쪽 관리 뿌리가 지워지는지 재현 측정하고, 확인되면 프로필 인식 정리로 바꾼다". 근거와 줄 번호는 `design.md` §A.6.

### Out of Scope — 같은 attempt 재부여 뒤 재전송의 lane slot 기준 멱등 범위

- ND2의 경우(같은 attempt 재부여 뒤 같은 키 재전송)를 위한 lane slot 기준 멱등 범위는 후속 후보로만 둔다(§E 리드 조정 결정). 운영자 결정 3의 조건부 이관(REQ-DHR-017 분기 A)과는 별개다.

### Out of Scope — t1082 소관

- lane worktree relocation(`/cd`, `thread/fork`), handoff tombstone과 redirect metadata, launch-pending handoff(M1), 두 rebind 경로(M2), 같은 attempt 재부여를 부를 조건.

### Out of Scope — dispatch record의 나머지 설계 필드

- `scope` / `authority_ref`, `worktree_ref` / `expected_head`, 결과 외 `evidence_refs`. AC-MSG-01·AC-FACT-01 판정에 필요하지 않다. `worktree_ref` / `expected_head`는 t1082가 필요할 때 추가한다.

### Out of Scope — AC-DHR-012·023과 REQ-DHR-015 런타임 조항 (카드 t1143으로 이관, 0.3.1)

- 리드 결정(sync-audit `.moai/reports/t1100/sync-audit.md` F1)에 따라 AC-DHR-012(FAIL), AC-DHR-023(미충족), REQ-DHR-015의 런타임 조항을 이 SPEC의 충족 범위에서 빼고 카드 t1143(최상위 read-only 실행 경로)에 넘겼다. 이 SPEC은 이 세 항목을 충족한다고 주장하지 않는다.
- 원인(측정, codex-cli 0.156.1): `spawn_agent`로 띄운 하위 에이전트가 부모 세션의 sandbox를 물려받아, 역할 TOML의 `sandbox_mode = "read-only"`가 적용되지 않았다. 근거는 `design.md` §C.1 정정 문단과 `acceptance.md`의 두 AC 이관 주석이다.
- 세 항목의 원문(요구사항 SHALL 문구, AC 본문의 Given/When/Then, 판정식 jq, 기대값)은 옮기지 않고 제자리에 둔다. t1143이 바꾸지 않고 이어받는다.

### Out of Scope — 새 저장소·전송 수단

- 새 브로커, 데몬, 별도 DB, 파일 기반 두 번째 소유 저장소. 레거시 `internal/sessionmsg` 의미 변경.

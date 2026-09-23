---
id: SPEC-DUAL-HARNESS-RECOVERY-001
document: design
created: 2026-09-23
updated: 2026-09-23
author: manager-spec
card: t1100
---

# Design — SPEC-DUAL-HARNESS-RECOVERY-001

이 문서는 run 단계가 따를 데이터 모델과 결정 표를 적는다. 함수 이름과 파일 배치는 run 단계에서 정하며 여기 적힌 이름은 제안이다.

## §A 배선 소유 기록과 저널 (REQ-DHR-001 ~ 007)

### A.1 소유 기록 — `internal/manifest` 확장 (제안)

현재 `FileEntry`는 파일 단위(`Provenance`, `TemplateHash`, `DeployedHash`, `CurrentHash`)다. 배선 파일은 한 파일에 사용자 부분과 MoAI 부분이 섞이므로 부분 단위 기록이 필요하다.

| 필드 | 뜻 |
|---|---|
| `provenance` | 기존 enum에 배선 생성물용 값 하나를 더한다(제안: `generated_managed`). 템플릿 관리 파일과 구분한다 |
| `parts[]` | 파일 안의 소유 단위 목록 |
| `parts[].kind` | `whole-file` / `hook-handler` / `toml-table` / `toml-key` |
| `parts[].key` | handler는 이벤트 키 + `moai hook ` 명령, 테이블은 `mcp_servers.moai`, 키는 `tui.status_line` |
| `parts[].origin` | `created`(MoAI가 만듦) / `preexisting`(설치 전부터 있음 → 사용자 소유) |
| `parts[].hash` | MoAI가 쓴 부분 바이트의 해시(`preexisting`은 비움) |

기존 신뢰 사이드카(`.moai/state/codex-wiring.json`)는 doctor의 divergence 신호로 계속 쓴다. 제거 판단의 근거는 사이드카가 아니라 이 기록이다.

`origin`은 처음 기록된 뒤 `preexisting → created`로 바뀌지 않는다(REQ-DHR-001). 이미 배선된 프로젝트에 이 기록이 없으면(마이그레이션 전 설치) 모든 부분은 provenance 없음으로 취급되고, unwire는 아무것도 지우지 않고 `no-provenance`로 보고한다. 이 보수적 기본값 때문에 구버전 설치를 한 번은 다시 배선해야 제거가 가능하다.

### A.2 쓰기 순서

```text
render → stage(temp in target dir) → journal append {path, pre_hash, post_hash, state: staged}
      → acquire update lock → re-read target, hash == pre_hash ?
            no  → journal state=conflict, remove temp, report (REQ-DHR-003)
            yes → rename → read back, hash == post_hash ?
                    no  → journal state=diverged, report
                    yes → journal state=complete, manifest parts update
      → release lock
```

잠금은 기존 `update_cleanup.go`의 update lock을 재사용한다. 저널은 사이드카와 같은 디렉터리에 두고 같은 패키지가 쓴다. 새 저장소 범주를 만들지 않는다.

### A.3 복구 분류 (REQ-DHR-004)

| 저널 상태 | 대상의 현재 해시 | 분류 | 행동 |
|---|---|---|---|
| staged / renamed | post_hash | completed | 저널 완료 처리, manifest 반영 |
| staged | pre_hash | not-applied | 임시 파일 삭제, 대상 무변경 |
| 무엇이든 | 둘 다 아님 | diverged | 대상 무변경, 보고 |
| conflict | 무엇이든 | conflict | 보고만 |

### A.4 unwire 결정표 (REQ-DHR-005, 006)

| 조건 | 행동 | 보고 사유 |
|---|---|---|
| 부분 `origin=created`, 현재 해시 = 기록 해시 | 그 부분만 제거(A.2 경로로 재작성) | — |
| 파일 전체가 `created`, 해시 일치 | 파일 삭제 | — |
| provenance 기록 없음 | 무변경 | `no-provenance` |
| `origin=preexisting` | 무변경 | `user-owned` |
| 해시 불일치 | 무변경 | `modified` |
| 경로 또는 상위가 프로젝트 밖을 가리키는 symlink | 무변경(Lstat 기준, 링크를 따라가지 않음) | `symlink-boundary` |
| 파싱 불가 파일 | 무변경 | `unparseable` |

### A.5 프로필 전환 (REQ-DHR-007)

- 템플릿 관리 파일(`.codex/agents/moai/*.toml` 등)은 배선 생성기가 아니라 템플릿 배포 경로가 다룬다(REQ-CW-012 유지). 대상 프로필이 배포하지 않는 경로는 기존 `classifyDeprecatedFile`(`PristineDeprecated` / `UserModifiedDeprecated` / `UnverifiedDeprecated`)과 `backupDeprecatedPaths`로 처리한다.
- 배선 파일(`hooks.json`, `config.toml`, 사이드카)은 전환만으로 지우지 않는다. 고아로 보고하고, 제거는 unwire가 한다. 자동 unwire 여부는 `plan.md` [NEEDS CLARIFICATION: unwire-trigger].

## §B Codex worktree 소유와 동시 writer (REQ-DHR-008 ~ 012)

### B.1 anchor lock

기존 `internal/session/anchor_lock.go`는 git worktree lock을 권위 있는 anchor 출처로 쓰고, 사유 문자열의 `pid <n>`으로 생존을 판정한다. Claude Code는 EnterWorktree 때 이 lock을 쓰지만 Codex는 쓰지 않는다. 그래서 `moai codex -w`가 대신 쓴다.

- 사유 형식(제안): `moai codex session <tree-name> (pid <pid> start <process-start>)`. `parseLockPID`가 읽는 `pid ` 토큰을 포함한다.
- direct launch는 `syscall.Exec`로 moai 프로세스가 codex로 바뀌므로 pid가 그대로 이어진다. exec 전에 `os.Getpid()`로 lock을 건다. spawn launch는 tmux pane 프로세스의 pid를 쓴다(기존 `defaultCodexSpawnPaneIdentity`).
- exec 이후에는 해제할 주체가 없다. 종료 후 남은 lock은 pid가 죽은 것으로 확인되어 anchor가 아니게 되고, 다음 launch가 교체한다(REQ-DHR-008).

### B.2 동시 writer 판정

`AnchorDecision(tree, lock, now)`를 launch 전에 호출한다. `Anchored=true`이고 그 보유자가 호출자 자신이 아니면 거부한다. 판정은 fail-closed다(판독 불가 사유, 생존 미확정 모두 anchored). `moai cc -w`의 기존 트리 진입에도 같은 판정을 쓴다.

### B.3 삭제 보호

`internal/cli/worktree/clean.go`의 분류(`classifyStaleWorktrees`, `worktreeHasLocalChanges`, `protectedWorktreePaths`, lock 상태)와 `done.go`의 L1 tier guard는 `.claude/worktrees/<name>` + `WT-<name>` 트리라면 누가 만들었든 적용된다. 새 보호 코드를 만들기보다, Codex가 만든 트리가 이 분류에 실제로 걸리는지 테스트로 증명한다(AC-DHR-008). 걸리지 않는 경로가 발견되면 그 분류기를 고친다.

### B.4 kanban 진입

`moai cc -k`의 파서(`parseLauncherEntry`), 이름 규칙(`appendLeadName`, `resolveCompanionName`), launch facts(`exportKanbanLaunchFacts`)를 codex 경로에서 재사용한다. codex는 `-f` 토큰을 이미 같은 방식(동사 조회 전에 떼어 냄)으로 처리한다(`codex_factory.go`). 인정할 역할 범위는 `plan.md` [NEEDS CLARIFICATION: codex-kanban-roles].

## §C 역할 권한 계약 (REQ-DHR-013 ~ 015)

### C.1 호스트 표현력

| 축 | Codex 역할 TOML로 강제 가능? | 판정 |
|---|---|---|
| 파일 쓰기(작업 공간 전체) | `sandbox_mode` 3값 | `enforced` |
| 쓰기 도구 구분(Write/Edit 대 shell 쓰기) | 불가 | `UNSUPPORTED` |
| shell 사용 금지 | 불가(sandbox는 쓰기 범위만 정함) | `UNSUPPORTED` |
| 역할별 MCP 도구 제한 | 불가(서버 등록은 세션 단위) | `UNSUPPORTED` |
| 역할별 하위 에이전트 생성 허가 | 불가 | `UNSUPPORTED` |
| 역할별 웹 접근 허가 | 불가(네트워크 차단 여부는 미측정) | `UNSUPPORTED` |

`read-only` 역할에서 shell 명령의 쓰기가 실제로 막히는지는 이번에 측정하지 않았다. AC-DHR-012의 LIVE 항목이 측정한다. 그 전까지 "read-only가 쓰기를 막는다"는 가설이다.

### C.2 현재 12개 역할 (이 트리의 생성 결과)

| 역할 | 현재 `sandbox_mode` | Claude 쪽 쓰기 도구 | 비고 |
|---|---|---|---|
| mission-governor | read-only | 없음(Read, Grep, Glob, Skill) | shell도 없음 → shell 금지는 `UNSUPPORTED` |
| super-advisor | read-only | 없음(Bash 있음) | MCP `codex_task`는 프로젝트 opt-in(`workflow.codex.task.allow_write`) 시 쓰기 가능한 작업을 만든다 → MCP 경유 쓰기는 `UNSUPPORTED` |
| plan-auditor | workspace-write | Write, Edit | 보고서 작성 |
| sync-auditor | workspace-write | 없음(Bash로 판정 파일 작성) | C.3 |
| manager-spec, manager-develop, manager-docs, manager-git, manager-design, manager-lead, builder-harness, e2e-tester | workspace-write | Write, Edit | manager-lead만 Agent 보유 |

Codex의 12개 역할은 MoAI 11개 + `mission-governor`다. CLAUDE.md의 12개(MoAI 11개 + 내장 `Explore`)와 구성이 다르다.

### C.3 감사 역할의 `workspace-write` 판정

**판정: 유지. 단, 사후 쓰기 범위 검증을 기계적 경계로 추가하고 위반 시 판정을 기각한다(REQ-DHR-015).**

근거:

1. 설계 §10은 "호스트가 표현하지 못하면 더 넓은 권한을 조용히 주지 말고, 차단하거나 검증된 worker 경계를 쓰라"고 한다. 여기서 비교 기준은 Claude 쪽 실효 권한이다. `plan-auditor`는 Claude에서도 Write/Edit를 가진다. `sync-auditor`는 Write/Edit가 없지만 Bash를 가지며, Claude의 Bash는 파일을 쓸 수 있다. 따라서 Codex `workspace-write`는 두 역할 모두에 Claude보다 넓은 쓰기 능력을 주지 않는다. 좁아지지 않는 것은 "쓰기 도구 구분" 한 축이며, 이것은 `UNSUPPORTED`로 기록한다.
2. `read-only`로 좁히면 두 역할은 판정·보고 파일을 쓰지 못한다. 그러면 호출자가 대신 쓰도록 두 하네스의 공통 에이전트 계약을 바꿔야 한다. 영향 범위가 이 카드보다 크다(기각한 대안).
3. 차단하면 Codex에서 감사 기능이 사라지고, 설계 §03의 "같은 필수 산출물" 계약을 깨뜨린다(기각한 대안).
4. 사후 검증은 기존 `moai worktree snapshot` / `verify`(감사 전후 작업 트리 비교)를 허용 목록과 함께 쓰는 방식이다. 새 메커니즘이 아니다.

남는 위험: Codex에서 이 검증을 호출하는 것은 워크플로 지시다. 호스트가 호출을 보장하지 않는다. 호출 강제는 M2(훅) 범위이며 여기서 증명하지 않는다. 또 검증은 막는 장치가 아니라 사후에 잡는 장치다. 쓰기는 이미 일어났을 수 있고, 판정 기각과 변경 경로 보고까지만 한다.

## §D dispatch record (REQ-DHR-016 ~ 021)

### D.1 테이블(제안)

`dispatches` — 기존 factorymsg SQLite 브로커 안의 새 테이블.

| 열 | 뜻 |
|---|---|
| `project_key`, `run_id` | 기존 브로커 범위 |
| `dispatch_id` | run 안에서 유일 |
| `card_id` | 카드 식별 |
| `lane_slot` | 대상 lane(안정 주소) |
| `attempt` | 1부터. 재할당마다 +1 |
| `assignee_generation` | 할당 시점의 `peers.generation` 복사본. fencing token |
| `idem_key` | 할당·결과 요청의 멱등 키 |
| `state` | `assigned` / `delivered` / `started` / `result_recorded` / `integrated` / `abandoned` |
| `result_ref` | 결과 참조(증거 경로 등). 한 번만 설정 |
| `updated_at` | 기록 시각 |

### D.2 전이

```text
assigned --(assignee의 receipt)--> delivered --(assignee의 start, fenced)--> started
started --(결과 적용, fenced + idempotent)--> result_recorded --(lead)--> integrated
assigned|delivered|started --(재할당: 이전 소유자 비생존 확인 또는 명시 철회)--> assigned(attempt+1)
어느 비종결 상태 --(lead 포기)--> abandoned
```

- fenced: 호출자의 (lane, generation, attempt)가 레코드의 값과 같아야 한다. 다르면 stale, 무변경.
- idempotent: 같은 결과가 다시 오면 `duplicate`, 무변경.
- 메시지 receipt는 `delivered`까지만 올린다. 그 이상은 명시적 호출만 올린다(REQ-DHR-016).

### D.3 멱등 범위 변경 (REQ-DHR-017)

현재 `messages`는 `UNIQUE(sender_session, idem_key)`다. 송신자가 재시작하면 세션 UUID가 바뀌어 같은 키가 새 메시지가 된다. 범위를 `(project_key, run_id, sender_slot, idem_key)`로 옮긴다.

- 마이그레이션: `sender_slot` 열 추가, 기존 행은 `peers`로 매핑 가능한 경우 그 slot, 불가능하면 `legacy:<sender_session>`으로 채운다. `SchemaVersion`을 올린다.
- 이 변경은 t1074가 착지시킨 계약을 바꾼다. 채택 여부는 `plan.md` [NEEDS CLARIFICATION: idempotency-scope].

### D.4 superseded (REQ-DHR-020)

`Claim`은 이미 `recipient_generation`으로 거른다. 수신 lane이 새 generation으로 재등록되면 이전 generation 앞 메시지는 누구도 claim하지 못한 채 남는다(코드 판독; 측정은 AC-DHR-015). 이 SPEC은 그 메시지를 status에 `superseded`로 따로 세고, 권한 재발급은 dispatch record의 새 attempt로만 한다. body를 새 generation으로 옮기는 경로는 만들지 않는다. t1082의 BOUND 후 방출이 그 경로를 따로 정한다.

## §E t1082 경계 요약

`spec.md` §E가 정본이다. 요점만 적는다.

- generation 필드는 하나(`peers.generation`), 멱등 범위도 하나(D.3).
- t1082의 handoff 상태기계는 D.2 위에 얹히는 gate이며, t1100은 handoff 상태를 읽지도 쓰지도 않는다.
- t1082의 M1(launch-pending 중 handoff), M2(두 rebind 경로)는 이 설계가 결정하지 않는다.

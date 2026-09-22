# t1082 — t1074 착지 후 SPEC 전제 재검증 (2026-09-23)

card: t1082 · SPEC-FACTORY-LANE-WORKTREE-HANDOFF-001 · 브랜치 WT-factory-lane-worktree-handoff

## 흡수

- 흡수 대상: 로컬 develop `08113ff0f` (t1074 착지 `861510fb6` 포함)
- 흡수 전 HEAD: `a63964217` (미병합 t1074 `8c5d9be99` 위)
- 흡수 병합 커밋: `a2d8afe84`, 트리 `82002a109cb20c5ce7c5b1eeb94a64fef92049a3`
- 충돌 1건: `internal/mcp/catalog_test.go` — 주석 2줄만 겹침(`wantCatalogSize = 36` 은 양쪽 동일). develop 문구 채택.
- 흡수 후 `git diff --stat 08113ff0f HEAD`: t1082 자체 파일 9개(SPEC 6 + 보고서 3), 코드 0.

## 기준점 이동량

`git diff --stat 8c5d9be99 f6e33c534 -- . ':!.moai/reports'` → 33 files, +3908/-263.
SPEC 이 쓰인 기준점 이후 t1074 가 착지 전에 추가한 것: `internal/factorymsg/store.go` +128(launch-pending 3 함수), `internal/cli/factory_launch_pending.go`, `internal/hook/factory_messages.go`, `internal/hook/user_prompt_submit.go`, `internal/cli/spawn.go`.
`git grep -c BindLaunchPending 8c5d9be99 -- internal/` → 0건, `08113ff0f` → store.go 6건 외.

## 전제 대조

| SPEC 전제 (출처) | 착지 코드 | 판정 |
|---|---|---|
| `Peer` 에 stable `Slot` + `SessionUUID`/`Generation` (research.md:15) | `store.go:49,81-85` | 일치 |
| `RegisterPeer` 는 살아 있는 lane owner 를 덮지 않음 (research.md:16) | `store.go:305` — "factory logical lane has a live owner" 거절 | 일치 |
| `ResolveLane` 존재 (plan.md:17) | `store.go:496` | 일치하나 동작 추가 — 아래 M1 |
| generation-bound message/receipt (plan.md:17) | Claim/Receipt SQL 이 `recipient_generation=?` 로 묶임 (`store.go:663~`, `749~`) | 일치 |
| `homestate.CanonicalProjectRoot` (research.md:17) | `internal/homestate/paths.go:40` | 일치 |
| `moai worktree new` → 공유 materializer (research.md:18) | `worktree/new.go` → `WorktreeCreator`, `root.go:137` 에서 `materializeSessionWorktree` 주입 (t1070 `1063e5aa0`) | 일치 |
| `worktree_branch_flag.go` 기존 브랜치 경로 (research.md:19) | 파일 존재 | 일치 |
| `mcp_codex.go` 에 `thread/start`·`thread/resume` (research.md:20) | `mcp_codex.go:62,69` 존재, `thread/fork`·`turn/steer` 없음 | 일치 (fork 는 SPEC 이 신규 확장으로 명시) |
| `session_start_record.go`, `cwd_changed_relocate.go` (research.md:21-22) | 둘 다 존재 | 일치 |
| t1074 canonical run selection (REQ-FLH-015) | `internal/cli/factory.go:222 enterSelectedFactoryRun` | 존재. plan.md:41 은 homestate 쪽에 두는 듯 서술 — 위치 표기만 부정확 |
| launcher provisional endpoint → 첫 SessionStart rebind (REQ-FLH-015) | `RegisterLaunchPending`/`BindLaunchPending`/`RollbackLaunchPending` (`store.go:~380-470`) | 존재. 단 SPEC 이 이 API 와의 상호작용을 규정하지 않음 — M1·M2 |

## 불일치 목록

- **M1 (차단 후보)** — `ResolveLane` 은 lane 끝점이 launch-pending 이면 `ErrEndpointLaunchPending` 을 돌려준다(`store.go:505-507`; `Peer`/`PeerByOwner`/송신 경로도 같은 오류, `store.go:485,521,564`). SPEC 은 handoff 가 기존 lane 주소를 `ResolveLane` 으로 재사용한다고 전제하지만, "lane 이 아직 launch-pending 인 상태에서 handoff 요청" 경우를 다루는 REQ·AC 가 없다. `grep -ci 'launch.?pending|provisional'` → spec.md 1(REQ-FLH-015 한 줄), design/plan/acceptance/research 0.
- **M2 (설계 공백)** — handoff 가 새 worktree 에서 새 세션을 띄우면 그 세션도 t1074 launcher 경로로 launch-pending 행을 만든 뒤 `BindLaunchPending` 으로 바인딩된다. SPEC 의 RESERVED→BOUND atomic rebind(reservation nonce)와 `BindLaunchPending`(정확한 owner identity 일치 요구, `store.go:423,444`)이 같은 slot 을 두고 두 개의 rebind 경로가 된다. 어느 쪽이 먼저·어떤 순서로 쓰는지 design.md 에 없다.
- **M3 (경미)** — plan.md:41 이 run resolver 를 homestate 인접으로 서술하지만 실제는 `internal/cli/factory.go:222`. 구현 착수 시 위치만 바로잡으면 된다.

M1·M2 는 SPEC 이 t1074 착지 전(`8c5d9be99`)에 쓰여 launch-pending 이 코드에 없던 시점의 산물로 보인다(기준점 grep 0건). 이 판단은 추론이며, SPEC 저자 의도는 확인하지 않았다.

## Gaps

- 테스트·빌드는 돌리지 않았다(병합 트리 코드 변경 0, 이번 요청 범위는 전제 대조).
- `BindLaunchPending` 과 handoff rebind 의 실제 경합은 코드로 재현하지 않았다 — M2 는 코드 읽기 기반 가설.
- push 없음.

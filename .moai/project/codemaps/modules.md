# 패키지 모듈 상세

**현재 갱신 — t1187, `origin/develop` `a8a9b9376` (2026-09-25).** 앵커
`bd71c59e4` 뒤 비테스트 Go 소스 41개의 끝점 변경을 대조했다. 새 경계는
`internal/cli/codex_audit_launch.go`와 `codex_audit_mcp.go`의 읽기 전용 Codex 감사,
`internal/homestate/factory_run_retire.go`의 부팅 시각 증거를 포함한 런 은퇴,
`internal/spec/lint_tier_artifacts.go`의 Tier 산출물 표 기반 린트다. 아래 패키지·클러스터
파일 수는 이 트리에서 `find ... -name '*.go' -not -name '*_test.go'`로 다시 셌다.

> `/moai codemaps`로 생성된 패키지 목록입니다. 존재 여부는 작업 트리만을 근거로 판정했고,
> 이전 codemaps 문서를 존재의 근거로 쓰지 않았습니다.
> **Go** 버전은 재측정 트리의 `go.mod`에서 직접 읽었습니다(`go 1.26.8`).

**모듈**: `github.com/modu-ai/moai-adk` · **Go**: 1.26.8
**최초 측정 트리**: worktree `.claude/worktrees/t592`, 브랜치 `WT-home-state-rollout`, HEAD `e7bd89ee3`, 2026-09-10
**재측정 트리**: worktree `.claude/worktrees/t869`, 브랜치 `WT-codemaps-refresh`, HEAD `a851b205c`, 2026-09-18 — 모든 표의 비테스트 파일 수, `internal/cli` 클러스터 표, fan-in 칸, 신규·누락 패키지 행(`internal/mission` · `internal/codextools` · `internal/gitenv` · `cmd/t657-merge`), § 네거티브 스페이스의 목록과 파일 크기. 책임 칸의 서술형 판단 중 이번 변경과 무관한 것은 앞 판을 이어받았습니다.
**정정 재측정**: worktree `.claude/worktrees/t872`, 브랜치 `WT-codemaps-citations`, HEAD `9a8cc4277`, 2026-09-18 — cross-cutting 표에서 삭제된 패키지 행 하나를 빼고, § 프로덕션 코드 없이 테스트만 있는 자리를 다시 셌습니다. 다른 표의 파일 수는 같은 명령으로 재확인해 변동이 없었습니다.
**정기 재측정**: worktree `.claude/worktrees/t999`, 브랜치 `WT-codemaps-remediation`, HEAD `56c64891a`, 2026-09-20 — 파일 수가 움직인 여섯 행(`internal/cli` 315→318, `internal/hook` 137→140, `internal/harness` 82→87, `internal/config` 56→57, `internal/spec` 36→41, `internal/statusline` 21→22)과 `internal/cli` 클러스터 표의 루트 파일 수(241→244), 그리고 신규 패키지 3개의 행 — `internal/auditreceipt`는 data/persistence 표에, `internal/harness/rosterguard` · `internal/harness/cellguard`는 § 네거티브 스페이스에 들어갔습니다. 변동이 없어 손대지 않은 행도 같은 명령으로 확인했습니다(`internal/kanban` 57 · `internal/navigator` 53 · `internal/web` 31 · `internal/session` 22 · `internal/homestate` 14 · `internal/mission` 15 · `internal/mcp` 1). 책임 칸의 서술형 판단은 이번에 건드린 행을 빼고 앞 판을 이어받았습니다.

**부분 재측정**: worktree `.claude/worktrees/t1083`, 브랜치 `WT-jev-guard-green`, sync-phase HEAD `dd19e6b90`, 2026-09-22 — sync-phase 부분 갱신(card t1083, SPEC-JEV-GUARD-001). Consumer B(스킬 제안 앵커 파일 + 그 테스트 — 파일명은 저장소 이력 참조)를 철수했다 — 측정 게이트 미실행 상태에서 배송된 게이트-언런 컨슈머로, consumer-guard 계약(소비자는 측정 이후에 배송) 위반이 확정됐다. `internal/cli` 총 326·루트 249로 재측정(철수 -1과 흡수된 develop 커밋의 `mcp_jev.go` 등 +분이 겹쳐 이전 판 수치와 선형으로 대응하지 않는다 — `find`/`grep`으로 이 트리에서 직접 센 값이다). 남은 `jev*` 루트 파일 넷은 `doctor_jev.go`·`init_jev_wizard.go`·`mcp_jev.go`(게이트된 `jev_ask` MCP 도구)·`todo_jev_finding.go`(Consumer C)다. `provenance.json`은 `codemaps-gen` 재생성 전용 스탬프라 손대지 않았다.

**부분 재측정(이력)**: worktree `.claude/worktrees/t1066`, 브랜치 `WT-jev-consumers`, run-phase HEAD `8c0e5dc9b`, 2026-09-22 — sync-phase 부분 갱신(card t1066). `internal/cli` 총 318→325, 루트 244→249. 루트 +5 중 이 카드 몫은 2개(`todo_jev_finding.go` — Consumer C 게이트 미실행 admission 훅, Consumer B 게이트 미실행 앵커 — 당시 파일, 이후 card t1083이 철수)이고, 나머지 3개(`doctor_jev.go`·`init_jev_wizard.go` — t1020 CORE-001 sync, `integration_codemaps_card.go` — t1018)는 흡수된 develop 커밋으로 들어와 각 카드의 sync가 codemap을 갱신하지 않아 누적된 몫이다. 하위 패키지 +2(`update/backup/file_snapshot.go`·`mcp_snapshot.go`)도 흡수 몫이다. `provenance.json`은 `codemaps-gen` 재생성 전용 스탬프라 손대지 않았다 — 다음 전체 재생성이 다시 찍는다.

**정기 재측정**: worktree `.claude/worktrees/t1069`, 브랜치 `WT-graph-restamp`, HEAD `0314801c2`, 2026-09-22 — 움직인 행 여섯(`internal/hook` 140→141, `internal/web` 31→32, `internal/homestate` 14→15, 신규 `internal/jev` 1 · `internal/jevcred` 1 · `internal/jevmeasure` 2)과 cross-cutting 표의 fan-in 칸 다섯(`internal/defs` 11→12, `internal/paths` 11→12, `internal/atomicfile` 10→11, `internal/lockfile` 1→2, `internal/stateanchor` 2→3), 그리고 §네거티브 스페이스의 큰 파일 표와 여기서 갱신한 행의 책임 칸. 변동이 없어 손대지 않은 행도 같은 명령으로 확인했습니다(`internal/cli` 325·루트 249, `internal/kanban` 57, `internal/navigator` 53, `internal/session` 22, `internal/mission` 15, `internal/harness` 87, `internal/config` 57, `internal/spec` 41, `internal/statusline` 22, `internal/mcp` 1).

**부분 재측정**: worktree `.claude/worktrees/t1092`, 브랜치 `WT-codemaps-restamp`, base `08113ff0f`, 2026-09-23 — 카드 t1092, 앵커 `598e8f748`(card t1069) 이후 착지분(주로 card t1071·t1057·t1083·t1077 등 factory/jev/가드 계열 흡수)을 본문에 반영. `internal/cli` 총 325→334, 루트 249→257, `codex*` 클러스터 12→18(신규 `codex_direct_posix.go`/`codex_direct_windows.go` — factory 실행과 대화형 Codex 세션 사이에 프로세스 identity를 보존하는 launch 래퍼, `codex_local_file*.go` 4개 — symlink race를 배제하는 플랫폼별 로컬 지시문 오픈), `mcp*` 클러스터 20→22(신규 `mcp_factory_msg.go`·`mcp_jev.go`), `internal/cli/worktree` 하위 신규 `new.go`(`moai worktree new <name>` — L1 워크트리를 기존 materializer로 생성만 하고 진입은 하지 않는 명령). `internal/hook` 141→143(신규 `subagent_write_guard.go` — SPEC-SUBAGENT-WRITE-SHRINK-GUARD-001, 서브에이전트가 추적 파일을 큰 폭으로 줄여 쓰는 PreToolUse를 거부; `factory_messages.go` — SessionStart/UserPromptSubmit에서 factory peer를 등록). `internal/homestate` 15→16(신규 `process_fingerprint_darwin.go` — `unix.SysctlKinfoProc`로 PID 재사용을 가르는 프로세스 시작시각 지문). **신규 패키지** `internal/factorymsg`(비테스트 1, `store.go`) — SQLite 기반 factory 전용 런-스코프 메시지 브로커, 레거시 `internal/sessionmsg`를 읽거나 이관하지 않는다고 패키지 주석이 명시. moai MCP 도구 총수 30→36(신규 `factory_msg_{send,list,body,receipt,status}` 5개 + `jev_ask` 1개, `internal/mcp/catalog.go` 카운트로 확인). 삭제된 Consumer B 스킬 제안 앵커 파일(파일명은 저장소 이력 참조)은 card t1083의 Consumer B 철수 몫으로 앞 판(t1083)에 이미 반영돼 있다. `provenance.json`은 `codemaps-gen` 재생성 전용 스탬프라 손대지 않았다.

**부분 재측정**: worktree `.claude/worktrees/t1132`, 브랜치 `WT-codemaps-refresh`, base `40bb5bb08`, 2026-09-23 — 카드 t1132, 앵커 `40bb5bb08`(card t1092) 이후 착지분(card t1104·t1111·t1122·t1126 의 `internal/spec` 정합성 계열, factory lane→worker 개명, `internal/jev` 벤더 실제 스키마 정합, Opus 5.5 모델 id 개명)을 본문에 반영. `internal/spec` 41→42(신규 `lint_req_bare.go` — 마크다운 마커 없는 REQ 정의 줄 수집; 아래 행 참조). `internal/kanban`·`internal/cli`는 파일 수 변동 없이(57·334 그대로) factory `-f agent`/`lane-<n>` 어휘가 `-f worker`/`worker-<n>`으로 개명됐다(레거시 스펠링은 읽기 전용으로 남아 별칭 처리). `internal/jev`(1)는 벤더가 실제로 문서화한 스키마(`questions`가 id로 키잉된 객체, `choice`/`score` 질문의 `criteria`, noul 은 확률값)에 맞춰 재작성됐다 — 파일 수는 그대로다. `internal/template`(31)은 `ModelIDOpus5`→`ModelIDOpus55`(`claude-opus-5-5`)로 개명하고 구 id 를 deprecated 목록에 얹었으며, `skill_mirror.go`에 배포 중 자기 미러 심볼릭 링크를 해제하는 `releaseOwnMirrorLink`가 더해졌다(파일 수는 그대로). `provenance.json`은 `codemaps-gen` 재생성 전용 스탬프라 손대지 않았다.

**부분 재측정**: worktree `.claude/worktrees/t1151`, 브랜치 `WT-codemaps-refresh2`, base `60017eb83`, 2026-09-24 — 카드 t1151, 앵커 `ee4e6d22f`(card t1132) 이후 착지분(주로 card t1100 SPEC-DUAL-HARNESS-RECOVERY-001·card t1082 SPEC-FACTORY-LANE-WORKTREE-HANDOFF-001·card t1107 SPEC-FACTORY-RUN-RETIRE-001)을 본문에 반영. `internal/cli` 총 334→341, 루트 257→264. **신규 명령** `moai tool disable codex`(`tool.go`) — MoAI가 실제로 기록한 부분만(자기 훅 핸들러·설명, `[mcp_servers.moai]`/`[tui]` 표나 자기가 넣은 `status_line` 줄, 만든 그대로 바뀌지 않은 배선 파일 전체) 골라 제거하고, 증명 못 하는 부분은 손대지 않은 채 사유와 함께 나열한다(`internal/codexwiring/unwire.go`). `codex*` 클러스터 18→19(신규 `codex_kanban.go` — `moai codex -k`, `moai cc -k`와 같은 파서·이름 레지스트리를 그대로 공유하는 Kanban Mode 진입), `update*` 클러스터 24→25(신규 `update_identity.go` — update 렌더가 `project.yaml`/`user.yaml`을 프로젝트가 이미 가진 이름으로 렌더해 이름 없는 재배포를 막는다, card t1139), `factory*` 하위 클러스터 2→8(신규 `factory_lane_handoff.go`·`_bind.go`·`_recover.go`·`_switch.go`·`factory_run_owner.go` — 안정된 lane을 카드 전용 L1 워크트리로 옮기는 M1(예약·로컬 develop pin·브로커 예약)→M2(SWITCH_PENDING 릴로케이션, 대화형은 `/cd` 안내만, headless는 `thread/fork`)→M3(원자적 rebind, BOUND)→재시작 복구(M4, 저널·파일시스템·Git 사실을 다시 읽어 resume/idempotent finalize/NACK/ABANDONED 중 하나만 고른다) 수명주기). `internal/codexwiring` 5→11(신규 `journal.go`·`lock.go`·`ownership.go`·`recover.go`·`unwire.go`·`write.go` — 아래 행 재작성 참조). `internal/factorymsg` 1→8(신규 `dispatch.go`·`handoff.go`·`handoff_abandon.go`·`handoff_bind.go`·`handoff_relocation.go`·`factory_run_retire.go`·`schema_migrate.go` — 아래 행 재작성 참조). `internal/homestate` 16→17(신규 `factory_run_retire.go` — 런 소유자가 살았는지 판별하는 `OwnerClassification`, 닫힌 집합이 아니라 `OwnerDead` 양성에만 은퇴를 허용). `internal/config` 57→58(신규 `loader_identity.go` — `project.name`/`user.name` 단일 키 판독기, update 렌더 경로가 managed cleanup 전에 값을 읽으려고 씀). `internal/hook` 143→144(신규 `factory_handoff_bind.go` — 헤드리스 rebind의 훅 쪽 절반). `internal/session`(22→23, 신규 `anchor_lock_holder.go` — 어느 세션이 잠금을 쥐고 있는지 판별하는 읽기 전용 accessor. `moai worktree remove`가 이를 소비해, 등록된 앵커 세션이 없어도 git worktree lock 자체가 앵커라면 git의 자체 에러 대신 이름 붙은 사유로 거절한다). `internal/harness/rosterguard/registry.go`는 파일 수 변동 없이 model-policy 프로필-매트릭스 행 두 개의 스테일 선언 세 개를 카드 t1141이 일괄 제거했다(문서 쪽이 옳아졌다 — 13행×3열=39셀로 정정). `internal/web`(32, 변동 없음)의 WAL 워처가 kqueue 생성-이벤트 경합(디렉터리 스캔이 파일 생성을 보고하기 전에 SQLite가 첫 프레임을 이미 쓰는 경우) 재확인 프로브(`walProbe`)를 얻었다(card t1136). `provenance.json`은 `codemaps-gen` 재생성 전용 스탬프라 손대지 않았다.

파일 수는 전부 `find <dir> -name '*.go' -not -name '*_test.go' | wc -l`로 센 **비테스트 파일**이며
하위 패키지를 포함합니다.

---

## presentation

| 패키지 | 비테스트 | 책임 | 주요 하위 패키지 |
|---|---|---|---|
| `cmd/moai` | 1 | 배포 바이너리 유일 진입점. `cli.Execute()` 호출 후 `cli.ResolveExitCode`로 종료 코드만 매핑 | — |
| `cmd/t657-merge` | 1 | **배포되지 않는 일회성 큐 병합 도구**(카드 t657). 파일 머리 주석이 사용자 verb가 아님을 밝히고, `internal/kanban` 저장소 API를 재사용하며 실제 저장소를 명시적 절대 경로 플래그로만 받는다 | — |
| `internal/cli` | 344 | 아래 클러스터 표 참조 | `update`(+`plan`/`deploy`/`merge`/`backup`/`report`), `harness`, `worktree`, `agentlint`, `preference`, `wizard`, `uikit`, `printer`, `specid`, `taskledger`, `pr`, `ptycaptest`, `jev` |
| `internal/hook` | 144 | Claude Code 26종 훅 이벤트의 핸들러 레지스트리와 개별 핸들러. `registry.Dispatch`가 이벤트별 체인을 돌려 `HookOutput`을 만든다. **이 판에서 `factory_handoff_bind.go`가 더했다** — SPEC-FACTORY-LANE-WORKTREE-HANDOFF-001 M3 원자적 rebind의 훅 쪽 절반(대화형 lane의 다음 턴 증거를 통한 bind; 헤드리스 절반은 `internal/cli/factory_lane_handoff_bind.go`). 이전 판에서 `subagent_write_guard.go`(SPEC-SUBAGENT-WRITE-SHRINK-GUARD-001, PreToolUse: 서브에이전트의 `Write`가 추적 중인 기존 파일을 바이트 기준 큰 폭으로 축소하면 거부. 판별은 경로 범위가 아니라 파괴성 자체 — 2000바이트 이상(SWG-T1)이던 파일을 25% 이하(SWG-T2)로 줄이는 전체 덮어쓰기. 거부 계층만 `workflow.subagent_write_guard.enabled`에 게이트되고 감지·감사는 무조건 돈다)와 `factory_messages.go`(SessionStart·UserPromptSubmit에서 `internal/factorymsg` peer를 등록하는 훅 쪽 배선)가 더했다. 그 앞 판에서 `internal/hook/session_heartbeat.go`가 더했다 — UserPromptSubmit마다 세션 레지스트리의 `last_heartbeat`를 갱신하는 seam(등록된 세션에만, 모든 실패는 침묵. 실측 근거는 § `entry-points.md` 훅 절) | `quality`, `security`, `mx`(+`complexity`), `memo`(+`taxonomy`), `handoff`, `perf`, `trace`, `testutil` |
| `internal/web` | 32 | 루프백 전용 브라우저 콘솔. `a-h/templ` 컴파일 뷰(`*_templ.go`) + htmx + SSE(fsnotify)로 프로파일·설정·todo 큐를 편집하고, **codex 탭 하나는 편집이 아니라 읽기 전용 미러**다(§ codex 미러 탭). 이 판에서 저장 실패 관측성(handleSave의 9개 persistence seam이 실패를 2xx 재렌더와 stderr 한 줄로 내보낸다 — § `data-flow.md` G)과 워크플로 설정의 Jev 패널·자격증명 필드(`internal/web/jevkey.go` — `internal/jevcred`를 통해서만 읽고 쓴다)가 더했다 | `assets` |
| `internal/statusline` | 22 | Claude Code statusLine 렌더러. git·github·model·backlog·goal·usage 세그먼트 조립. 렌더가 읽고 쓰는 상태의 **앵커는 세션의 현재 디렉터리가 아니라** `internal/stateanchor` seam이 정한 프로젝트 루트이며, 그 어댑터가 `internal/statusline/state_anchor.go`다 | — |
| `internal/tui` | 19 | 터미널 UI 디자인 시스템 — 박스·필·테이블·테마(Catppuccin)·i18n 메시지(`//go:embed messages/*.yaml`) | `golden`, `internal` |
| `internal/mcp` | 1 | self-hosted MCP 도구 카탈로그(도구명 + write 여부) 단일 선언 — 현재 39개(`codex_role_audit{,_status,_result}` 포함) | — |

### `internal/cli` 기능 클러스터

`codex*` 클러스터에는 이번 판의 `codex_audit_launch.go`와
`codex_audit_mcp.go`가 들어갔다. 셸과 MCP가 같은 `runCodexAudit` 코어를
쓰며, `codex exec -s read-only`를 최상위 프로세스로 실행한다. `-c`로
MCP 서버를 끄고 역할 파일의 지시문 크기·워크트리·목적지를 제한하며,
성공한 반환문을 런처가 `.moai/reports/`에 기록한다. MCP job 상태는
서버 프로세스의 메모리에만 있으므로 서버 재시작 후 같은 job ID 조회는
지원하지 않는다. 파일 수 19→21은 이 두 신규 파일에서 발생했다.

루트 267개 비테스트 파일(`find internal/cli -maxdepth 1 -name '*.go' -not -name '*_test.go'`)을
파일명 접두어로 묶은 것입니다. 파일 수는 같은 명령에 `-name '<접두어>*.go'`를 붙여 셌습니다.

| 클러스터 | 파일 | 담당 |
|---|---|---|
| `update*` | 25 | 템플릿 재배포 — 계획/분류/네임스페이스 보호, 3-way 머지, 백업·롤백, 클린 인스톨, dry-run. 단계 로직은 `cli/update/{plan,deploy,merge,backup,report}` 하위로 분해돼 있다. **재배포가 일어나지 않는 경로에도 복구 하나가 붙는다** — `internal/cli/update_mirror_heal.go`는 버전 일치 update가 Deploy 앞에서 조기 반환하는 자리 옆에서 `.agents/skills` 미러를 복구하며, 존재 게이트는 프로젝트의 기록된 배포 버전이다. 백업 쪽에 `.mcp.json`이 settings.json과 같은 스냅샷 처리를 얻었다(`backup/mcp_snapshot.go` — 배포가 실제로 쓴 렌더를 staging→promote로 기억해, 다음 update의 3-way 머지가 템플릿이 바꾼 값을 사용자가 못 건 키에서 보이게 한다. 생명주기 기계는 `file_snapshot.go`로 추출돼 두 파일이 공유한다). **이 판에서 `update_identity.go`가 더했다**(card t1139) — update가 `project.yaml`/`user.yaml`을 렌더할 때 프로젝트가 이미 가진 `project.name`/`user.name`을 읽어 실어, 바뀌지 않은 이름이 빈 값으로 렌더되지 않게 한다(managed cleanup이 `.moai/config`를 지우기 전에 값을 미리 읽음, `internal/config/loader_identity.go` 소비) |
| `doctor*` | 18 | 진단 — config, disk, harness, hook wiring, mcp version, permission, sandbox, skills, worktree base, agentemit embed, codex, jev(게이트·credential·도달성을 읽기 전용으로 확인 — 판정 요청을 보내지 않는다; 도달성은 TCP 접속·종료만 한다), 그리고 이 판에서 더해진 git-strategy workflow 판정(`doctor_git_strategy_workflow.go` — 허용 4값 판정·상시 브랜치·통합 대상을 읽기 전용으로 보고). binary-lag 판정은 이 판에서 비교 ref가 바이너리보다 오래된 경우(`StatusAhead`)를 OK 대신 WARN으로 보고한다 — 그 비교는 바이너리의 신선도에 대해 아무 말도 하지 못한다 |
| `mcp*` | 22 | 세 갈래. `mcp_server.go`(51KB)는 stdio JSON-RPC 서버, `mcp.go`/`mcp_codex.go`(97KB — CLI 최대 파일)/`mcp_glm.go`/`mcp_convergence.go`는 codex·GLM 위임과 다중 모델 감사 수렴, 그리고 `mcp_claude*.go` 5개(`_runner` · `_protocol` · `_process_unix` · `_process_windows` 포함)는 `claude` CLI를 서브프로세스로 띄우는 읽기 전용 `claude_audit` 도구다. 서브프로세스 환경에서 `CLAUDE_CODE_*`·`CLAUDECODE`를 지우고 출력 상한을 둔다. `audit_multi` 수렴도 `mcp_convergence.go`에서 같은 함수를 부른다. **이 판에서 둘이 더했다** — `mcp_factory_msg.go`(`internal/factorymsg` 위의 5개 도구 `factory_msg_{send,list,body,receipt,status}` — 발신자는 현재 세션·프로세스로 귀속되고 본문은 `factory_msg_body`로만 신뢰되지 않는 데이터로 읽힌다)와 `mcp_jev.go`(게이트된 `jev_ask` 1개 — `workflow.jev.enabled` 기본 꺼짐이면 요청 자체를 조립하지 않는다) |
| `todo*` | 15 | 백로그 큐 CLI. 파일 헤더가 스스로를 `kanban.BacklogStore`에 대한 얇은 cobra 배선이라고 밝힌다. 이 판에서 Jev near-duplicate admission 훅(`todo_jev_finding.go` — 카드 admission 경로에서만 불리는 게이트 미실행 Consumer C 훅)이 더했다 |
| `gtd*` | 2 | **이 판에서 새로 생긴 클러스터.** `gtd.go`의 `NewGTDCommand()`는 `newTodoCmd()`를 감싸 `Use`만 `gtd`로 바꾸고 `capture`·`clarify`·`organize`·`reflect`·`engage`를 더한다 — 같은 SQLite 큐 위의 두 번째 이름이지 별도 저장소가 아니다. `gtd_answer.go`의 `answer`는 게이트에서 멈춘 카드에 대한 운영자 답을 파일로 남기며 큐 항목을 바꾸지 않는다 |
| `codex*` | 21 | 외부 에이전트 백엔드 런처, 잡 제어, 준비 상태 점검, 리뷰 게이트, 그리고 `moai codex -f` 진입(`codex_factory.go` — 코덱스 verb 조회 전에 factory 플래그를 떼어낸다). **여기에 사용자 HOME 계층에 대한 스킬 노출 제어 두 개가 함께 산다** — `internal/cli/codex_skills_disable.go`는 `~/.codex/config.toml`에 `enabled = false`를 실은 `[[skills.config]]` 항목을 발행하고, `internal/cli/codex_skills_prune.go`는 가리키는 파일이 사라진 유령 등록을 제거한다(부재를 증명할 수 있는 것만 지우는 allowlist 형 판정, 기본 dry-run). `codex_direct_{posix,windows}.go`(factory 런치와 대화형 Codex 세션 사이에서 프로세스 identity를 하나로 보존하는 `exec.Cmd` 래퍼 — 훅 서브프로세스가 sandbox 래퍼가 아니라 스탬프된 owner를 직접 쓸 수 있게 한다)와 `codex_local_file{,_unix,_windows,_unsupported}.go`(로컬 지시문 파일을 심볼릭 링크 재대상·FIFO 블로킹 같은 경합 없이 여는 플랫폼별 오픈 — unix는 `O_NOFOLLOW`·`O_NONBLOCK`, windows는 `CreateFile`의 reparse-point 플래그, 그 외 플랫폼은 미지원을 명시적으로 에러 반환). **이 판에서 `codex_kanban.go`가 더했다** — `moai codex -k`: `moai cc -k`와 같은 파서(`parseKanbanFlag`·`parseCompanionLabel`·`parseLeadLabel`)와 이름 레지스트리(`resolveCompanionName`·`appendLeadName`)를 그대로 공유하는 Kanban Mode 진입으로, 두 문이 모양의 의미에서 갈릴 수 없게 한다 |
| `migrate*` | 11 | 프로파일·에이전시·스킬 복원과 HOME SQLite 상태의 점검·이전·복구·롤백 verb. `migrate_home_state.go`는 기본 dry-run이며 실제 쓰기는 `--apply --verified-live` 이중 승인과 두 번의 zero-active census를 요구한다 |
| `factory*` / `handoff*` / `profile*` | 14 (8 + 1 + 5) | Factory 인계 v2 lease·만료 재점유·token CAS, 레거시 claim의 명시적 `recover-resume`, 전역 프로필 lease의 provisional→transfer→enrich→release 수명주기. **`factory*`가 이 판에서 2→8로 늘었다**(SPEC-FACTORY-LANE-WORKTREE-HANDOFF-001) — `factory_lane_handoff.go`(M1: 안정된 lane을 카드 전용 L1 워크트리로 옮기기 전의 fail-closed admission, 로컬 develop pin, 내구 브로커 예약, WT_READY까지의 정확한 provenance 검증), `factory_lane_handoff_switch.go`(M2: WT_READY→SWITCH_PENDING 릴로케이션 어댑터 둘 — 대화형은 사용자가 실행할 `/cd` 안내만 내보내고, 헤드리스는 공식 app-server thread/fork·thread/start로 옮기고 그 증거와 컨트롤러 자신의 대상 재판독을 함께 기록), `factory_lane_handoff_bind.go`(M3 헤드리스 절반: M2가 남긴 결과와 대상 재판독만으로 BOUND로 원자적 rebind, 훅 이벤트·모델 턴 없이), `factory_lane_handoff_recover.go`(M4: 재시작한 컨트롤러가 브로커·파일시스템·Git 사실을 다시 읽어 resume/idempotent finalize/NACK/ABANDONED 중 정확히 하나만 고르는 재시작 복구 — 워크트리를 절대 지우지 않고 증명 못 하는 사실을 추측하지 않는다), `factory_handoff_recover.go`(`abandon-lane`이 브로커의 종결 트랜잭션에 건네는 소스-소유자 생존 프로브), `factory_run_owner.go`(`stampFactoryRunOwner` — 세션 identity가 생긴 뒤, 자기 프로세스를 대체하지 않은 모든 런치 문이 부르는 소유자 재스탬프 seam, REQ-002b) |
| `spec*` | 8 | SPEC 문서 lifecycle CLI (view/close/audit/drift) |
| `hook*` | 7 | 훅 디스패처 진입점(`hook.go`, 61KB)과 pre-commit/pre-push 설치 |
| `harness*` | 7 | harness route/validate/ledger/mute/delegation/clusters |
| `init*` | 9 | 프로젝트 초기화(`init.go`) — 템플릿 배포 + settings 생성 + MCP 프로비저닝. **이 판에서 `init.go`가 `InitOptions.AfterTemplateDeploy` 훅을 쓰게 됐다**(card t1139) — `internal/core/project/initializer.go`의 `InitOptions`에 더해진 콜백으로, 배포 직후·섹션 패치(report format, Page-3 위자드 답, workflow 토글) **전에** 순수 템플릿 렌더를 다음 update의 3-way 머지 BASE로 기록한다(패치 뒤에 스냅샷을 찍으면 위자드 답이 BASE가 돼, 다음 update가 그 답을 템플릿 기본값으로 되돌려 버린다). 이전 판에서 init wizard의 Jev 답을 `moai web` settings 화면과 같은 공유 persistence seam 으로 돌리는 bridge(`init_jev_wizard.go`)가 더했다 |
| `glm*` | 5 | GLM 백엔드 런처·잡 제어 |
| `navigator*` | 5 | BAS 파이프라인 CLI 단계 (enrich/sync/tiers/route/fix) |
| `graph*` / `gate*` / `web*` / `mx*` | 4 각 | 신선도·인용 게이트, 품질 게이트, 콘솔 기동, MX 태그 스캔 |
| `session*` | 4 | 세션 레지스트리 조회·메시징 CLI |
| `integration*` | 3 | 병합 창(acquire/status/release)과 **그 선행 조건인 설정 드리프트 단정**. `integration_settings_drift.go`가 `acquire`의 precondition 이자 독립 verb `moai integration preflight`이며, 창을 잡지 않고도 같은 질문을 물을 수 있게 두 표면을 함께 둔다. 이 판에서 codemaps 부채의 상시 발화원(`integration_codemaps_card.go` — `release` 시점에 부채 문턱을 넘으면 큐에 카드를 쌓는다, 발화만 하고 고르지는 않는다)이 더했다 |
| `jev*` | 4 | 게이트 뒤의 표시 전용 소비자 넷: `doctor_jev.go`(읽기 전용 Jev check), `init_jev_wizard.go`(init 질문), `mcp_jev.go`(게이트된 `jev_ask` MCP 도구 — `workflow.jev.enabled` 기본 꺼짐), `todo_jev_finding.go`(Consumer C admission 훅). Consumer B 앵커(당시 Hidden CLI 스킬 제안 명령을 포함)는 측정 게이트 미실행 위반으로 SPEC-JEV-GUARD-001이 철수했다 — 재편입은 측정 게이트(`SPEC-JEV-OPTIN-MEASURE-001`) 통과 이후다 |
| `kanban*` / `goal*` | 2 각 | 보드 CLI, goal 조건 arm/status/clear |
| `skills*` | 1 | `moai skills` 명령 트리(`internal/cli/skills.go`). 스킬 노출을 **계층별** 관심사로 두고 계층을 verb 가 아니라 플래그로 명명하며, `--codex`를 필수로 만들어 사용자 HOME 쓰기를 호출 시점 opt-in으로 고정한다 |
| 나머지 | 69 | `launcher.go`(54KB, cc/glm 런처 — `cg`는 `root.go`의 `trivialCommands`에 은퇴 토큰으로만 남았다), `slot.go`(자원 슬롯 임대 `moai slot`), `deps.go`(합성 루트), `root.go`, `profile*` 등과 플랫폼 분기(`*_windows.go` / `*_unix.go`) |

---

## business / domain

| 패키지 | 비테스트 | 책임 | 주요 하위 패키지 |
|---|---|---|---|
| `internal/harness` | 87 | GAN 루프 harness — Socratic 인터뷰 버퍼, 계층적 수락 스코어링, 패턴 학습·티어 분류, FROZEN 가드, lineage 매니페스트, 회귀 게이트. **이 판에서 문서 드리프트 가드 두 개(`rosterguard` · `cellguard`)가 더해졌는데, 둘은 런타임 경로가 없어 이 줄의 책임 서술에 들어가지 않는다** — 아래 § 네거티브 스페이스가 소유한다 | `curator`, `cluster`, `proposalgen`, `router`, `routing`, `safety`, `seeds`, `throttle`, `tier`, `capture`, `delegationmap`, `v4manifest`, `harnessrun`, `rosterguard`, `cellguard` |
| `internal/navigator` | 53 | BAS(Blueprint-Anchored Synchronization) 파이프라인. 루트에 Go 파일이 없고 전부 단계별 하위 패키지 | `astx`(tree-sitter 16개 언어), `detect`, `sync`, `tiers`, `route`, `fix` |
| `internal/kanban` | 57 | 백로그 큐의 상태 레코드·컬럼·역할 모델, SQLite 저장 엔진, 보드 락, PR 링크, 정합성 조정. **이 판에서 GTD 계층 10개 파일이 같은 `backlog.db` 위에 올라왔다** — `gtd_capture/clarify/organize/reflect/engage.go`가 다섯 단계를, `gtd_relation.go`가 항목 관계를, `gtd_operation.go`가 준비 후 실행하는 operation을, `gtd_persistence.go`가 export/import/backup/restore를, `backlog_gtd_schema.go`가 스키마 마이그레이션을 맡는다. 비테스트 소비자는 `internal/cli/gtd.go`, `internal/cli/goal.go`, `internal/graph/gtd_private.go`다. finding 출처도 이 판에서 셋이 됐다 — model-produced 답을 위한 세 번째 상수(`jev`)가 기존 둘(`mechanical` · `agent`) 옆에 더해졌는데, 재사용이 아니라 셋째인 이유는 둘이 **누가 관찰했는가**로 공간을 나누고 모델 답은 어느 쪽도 아니기 때문이다(§ `data-flow.md` L). 저장소 해석은 `internal/homestate`의 프로젝트 키 경로를 사용하고 Factory 런타임 진입은 migration admission gate를 통과한다. **여기에 워킹 트리 검사 하나가 더 있다** — `settings_drift.go`가 병합 전 tracked `.claude/settings.json`의 워킹 사본 드리프트를 단정하고 사본을 보존하며 원장에 남긴다(`--no-optional-locks` 강제 — 평범한 status가 인덱스 쓰기 락을 잡아 병합 직전 경합을 스스로 만들기 때문). **이 판에서 factory 워커 어휘가 `lane-<n>`/`agent-<n>`에서 `worker-<n>`으로 정식 개명됐다** — `factoryLaneRole` 상수 자체가 `"worker"`를 담고(레코드에 쓰이는 role 키는 하위호환을 위해 여전히 `"lane"`), 두 레거시 스펠링(`lane-<n>`·`agent-<n>`)은 `SplitFactoryLaneLabel`/`IsLegacyFactoryLabel`/`CanonicalFactoryLabel`로 계속 읽히되 이 패키지가 새로 만드는 라벨은 전부 `worker-<n>`이다. 클레임 경로도 `ClaimFactoryWorkerName`(구) 대신 `ClaimFactoryWorker`(신, `auto`/레거시 충돌 보고를 갖는 `FactoryClaim`/`FactoryLegacyCollisionError` 반환)로 갈렸다 | — |
| `internal/spec` | 43 | SPEC 문서 파싱/린트/감사, era 분류, per-SPEC 파일 락, atomic close 오케스트레이터. 이 판에서 close·audit의 §E.4 leg가 표지 존재가 아니라 **본문 내용**을 읽게 바뀌었다 — 플랜 페이즈가 §E.N 헤더를 자리표시자와 함께 먼저 심으므로, 한 줄 강조문뿐인 본문은 비어 있는 것으로 판정한다(판별은 철자가 아니라 구조로 한다). **이 판에서 셋이 더했다** — ① `NormalizeStatusValue`(`audit.go`)가 frontmatter `status:` 값의 앞뒤 공백과 YAML 따옴표 한 겹을 벗기는 공유 정규화 함수로 신설돼, 이 패키지와 `internal/kanban`의 모든 status 판독기(`checkV3R6Drift`·`loadSpecCloseState`·`parseStatusDiffLine`·`parseStatusFromYAML`·`kanban.parseFrontmatterStatus`)가 이 함수 하나로 수렴했다(fan_in 5, `@MX:ANCHOR`). ② `isValidInPlaceAmendment`(`audit.go`)가 `completed → in-progress (amendment)` 정당한 상태(`amendment_of` 선언 + 본문 Amendments 기록 + 그 기록이 이전 §E.4 `sync_commit_sha`를 인용)를 판별해, 이 조건을 만족하면 `SyncStatusDrift` finding 을 내지 않는다. ③ 신규 파일 `lint_req_bare.go`가 마크다운 마커(리스트 불릿·표 행·헤딩) 없이 줄 맨 앞에서 바로 시작하는 REQ 정의(`**REQ-X-001** — …`)를 네 번째 수집 소스(`REQSourceBare`)로 잡는다 — 기존 세 소스는 각자의 마커 문자로 서로 겹치지 않는데, 이 소스는 마커가 없는 대신 앵커를 **줄 맨 앞 칸**으로 고정해(들여쓰기·리스트 불릿 오인 방지) 같은 배타성을 지킨다. 구분자(`—`/`:`) 없이 ID(와 분류 괄호)만 담은 굵은 헤더 줄 다음 줄에 서술이 오는 **두 줄 형태**도 같은 소스로 잡는다(card t1120) — 헤더는 여는 `**`가 필수이고(줄바꿈된 산문 문단의 오인 방지), 다음 줄이 리스트·표·헤딩·인용·코드 펜스가 아닌 평문 문단일 때만 서술로 채택하며, 아니면 추측하지 않고 건너뛴다 | — |
| `internal/mission` | 15 | **이 판에서 새로 생긴 패키지.** 자율 미션의 권한 계층 — 봉인된 계약(`contract.go`)과 미션 상태(`auto_state.go`)를 저장하고, 거버넌스·완료 receipt를 읽어(`governance_receipt.go` · `completion_receipt.go`) 결정을 정책(`policy.go`)에 대조한 뒤, 증거 적재→정책 검증→실행→readback의 감독 루프(`supervisor.go`)를 돈다. git·전달 소유자(`git_owner.go` · `delivery_owner.go`)는 결과를 상태 재판독으로 확인한다. 비테스트 import는 `internal/atomicfile`·`internal/goal`이고, 비테스트 소비자는 `internal/cli/goal.go` 하나다 | — |
| `internal/template` | 31 | `//go:embed all:templates` + `catalog.yaml`. 배포기, 렌더러, settings 생성, 스킬 미러, 카탈로그 트리 해시, 모델 정책·프로파일 매트릭스. **배포 뒤편에 두 개의 기계 방출기와 두 개의 미러 보호·복구 seam이 붙어 있다**(§ 템플릿 방출·미러 계열). 임베드 트리는 588개 파일이다. **이 판에서 모델 정책의 `opus` alias 타깃이 `ModelIDOpus5`(`claude-opus-5`)에서 `ModelIDOpus55`(`claude-opus-5-5`, Claude Code v2.1.280+ 필요)로 갈렸다** — 구 id 는 `ModelDeprecatedCanonicalIDs`에 얹혀 여전히 `opus` alias 로 역정규화된다. **미러 보호 seam이 셋째를 얻었다** — `skill_mirror.go`의 `releaseOwnMirrorLink`가 배포기가 `.agents/skills/<skill>/` 아래로 실제 파일을 쓰기 직전, 이전 배포가 같은 경로에 남긴 자기 미러 심볼릭 링크(대상이 `MirrorLinkTarget(skill)`과 정확히 일치하는 것만)를 지운다 — 그러지 않으면 managed clean 이 링크 타깃을 지운 뒤 `MkdirAll`이 "file exists"로 실패한다 | `agentemit`, `commandemit`, `scripts` |
| `internal/core` | 23 | 응집 없는 우산 패키지 (§ `overview.md` 참조) | `git`, `project`, `quality` |
| `internal/mx` | 16 | `@MX:` 코드 주석 태그 스캐너·리졸버 (16개 언어) | — |
| `internal/graph` | 16 | 코드베이스 엣지 리스트를 git-diffable JSONL로 영속화하고 fan-in·최단경로·인용 검증·아키텍처 리포트를 생성. `gtd_private.go`는 GTD 항목의 비공개 그래프 투영을 만들고 권한을 검사한다(`internal/kanban` import). **freshness 게이트도 여기 있다** — codemaps 층은 값을 재기 전에 비교 가능성부터 판정한다(§ `data-flow.md` I) | `symbol` |
| `internal/constitution` | 18 | 규칙 트리의 FROZEN/EVOLVABLE 존 모델과 개정 절차 | — |
| `internal/migration` | 8 | 버전 간 마이그레이션 스텝 레지스트리 | `migrations` |
| `internal/feedback` | 7 | 피드백 리포트 스크러빙(민감정보 제거)과 재시도 큐 | — |
| `internal/goal` | 6 | goal 엔진 — 세션별 조건 선언형 완료 조건 | — |
| `internal/loop` | 6 | Ralph 피드백 루프 상태 기계 | — |
| `internal/verify` | 6 | 공유 진단 스냅샷 계약 | — |
| `internal/merge` | 6 | 템플릿 3-way 머지 엔진 | — |
| `internal/permission` | 6 | 8-tier 권한 스택 (`mvdan.cc/sh`로 셸 명령 파싱) | — |
| `internal/evolution` | 5 | Reflective Learning write phase | — |
| `internal/epic` | 5 | 디스크 기반 epic 진행률 산출 | — |
| `internal/codexwiring` | 11 | Codex 측 배선 파일 생성·갱신. **이 판에서 5→11로 늘며 SPEC-DUAL-HARNESS-RECOVERY-001의 저널드 쓰기·소유권 인식 unwire를 얻었다.** `write.go`가 한 배선 변경의 중단 지점 4개(P1 저널됨→P2 스테이징됨→P3 재확인됨→완료)를 정의하고, `journal.go`(`.moai/state/codex-wiring-journal.json`)가 임시 파일이 생기기 **전에** 모든 변경을 append해 중단된 변경이 저널만으로 항상 복구 가능하게 한다. `lock.go`(`.moai/state/codex-wiring.lock`)는 이 패키지 전용 락으로, `moai update` 락과 순서가 고정돼(wiring 락 다음에 update 락은 가능해도 역순은 쓰지 않음) 교착을 배제한다. `recover.go`가 중단된 변경을 4개 `RecoveryClass`(completed/not-applied/diverged/orphan-temp) 중 하나로 분류한다. `ownership.go`가 파트 키(`description`·`mcp_servers.moai`·`tui`·`tui.status_line`)별로 MoAI가 실제로 쓴 부분만 추적한다. `unwire.go`가 `moai tool disable codex`의 엔진 — 소유가 증명되고 그 이후 안 바뀐 부분만 제거하고, 증명 못 하는 부분은 사유와 함께 그대로 남긴다 | — |
| `internal/guardliveness` | 4 | 가드 발화 생존성 표면 | — |
| `internal/workflow` | 4 | worktree 전반 워크플로 오케스트레이션 | — |
| `internal/foundation` | 4 | TRUST 등 방법론 타입 정의 | `trust`(빈 디렉터리) |
| `internal/profile` | 3 | 사용자 프로파일·선호 동기화 | — |
| `internal/ciwatch` | 3 | CI watch 루프 분류기 | — |
| `internal/jevmeasure` | 2 | **이 판에서 새로 생긴 패키지.** Jev 소비자를 위한 측정 장치 — 한국어 원문·영역 번역 두 언어 팔(arm)로 라벨 붙은 표본을 돌리고, 상수 응답 baseline과의 대조로 소비자가 존재해도 되는지를 판정하는 보고서를 낸다. 이 패키지는 접점에 연락하지 않는다 — `Answerer`를 주입받고, 살아 있는 구현은 `internal/jev` 클라이언트다. 비테스트 소비자 0 — 아래 §네거티브 스페이스 | — |
| `internal/ralph` | 1 | Ralph 결정 엔진 (`engine.go` 단일 파일) | — |

### 템플릿 방출·미러 계열 — 앵커 이후 자란 하위 계층

`internal/template`의 책임 칸 한 줄로는 담기지 않는 네 단위가 하위에 있습니다. 넷 다
"배포기·렌더러"와 다른 축의 일을 합니다.

| 단위 | 비테스트 | 책임 |
|---|---|---|
| `internal/template/agentemit` | 6 | 보존된 에이전트 정의(`.md`)와 임베드 매니페스트(`agents-codex.yaml`)의 쌍을 **중립 원본**으로 삼아 `.codex/agents/` TOML을 결정적으로 이중 발행한다. `.md`의 발행은 항등(identity)이라 재렌더·재정렬이 없고, Codex 쪽은 (`.md` × 매니페스트)의 결정적 변환이다. **fail-closed** — 알 수 없는 tool 토큰·미매핑 effort·유효하지 않은 sandbox 값이면 어느 파일의 어느 토큰인지 지목하며 실패하고 부분 산출물을 남기지 않는다(codex-cli가 알 수 없는 설정을 조용히 무시하므로 생성기 쪽이 자기 출력을 검증해야 한다). **이 판에서 `permission.go`가 더했다** — 역할 권한 계약: 발행되는 모든 Codex 역할이 고정된 축 집합 위의 계약을 진다. 역할의 Claude 도구 목록과 contract sandbox에서 파생된 요구 제약은 매니페스트 축 표의 정확히 한 행에 매핑돼야 하고, 그 행은 `enforced`(이 생성기가 쓰는 Codex 필드)이거나 `UNSUPPORTED`(호스트가 표현 못 함)여야 한다 — 매핑되는 행이 없는 요구 제약은 발행을 실패시켜, 제약이 조용히 빠지지 않고 `UNSUPPORTED` 행이 통과로 잘못 세어지지 않는다 |
| `internal/template/commandemit` | 3 | `/moai` 명령 소스를 codex 스킬 아티팩트(`.agents/skills/moai-<command>/SKILL.md`)로 발행한다. 명령 소스는 읽기 전용으로 소비하며 **본문은 바이트 동일 verbatim** — 본문에 남은 Claude 전용 도구 참조는 여기서 고치지 않고 경계 플래그로만 기록한다(그 수리는 명령 본문 계층 소관). fail-closed: 프론트매터 구분자 누락, 설명 누락, 무조건 분기 없는 로케일 조건부 설명, 기존 정본 스킬 디렉터리와 충돌하는 파생 이름 |
| `internal/template/published_skills.go` | (파일) | 위 발행 스킬 경로에 대한 **배포측 보호**. 발행 스킬은 보통의 템플릿 파일처럼 배포되지만 경로 네임스페이스가 스킬 미러가 쓰는 `.agents/skills` 루트와 겹치고, update 모드(forceUpdate)는 다른 곳에서 provenance 검사를 건너뛴다. 이 검사가 그 경로들에 한해 init 모드의 provenance 동작을 살려 사용자 소유 파일이 update를 살아남게 하고, 건너뜀을 침묵이 아니라 보고로 남긴다 |
| `internal/template/skill_mirror_repair.go` | (파일) | `.agents/skills`의 두 생산자(심볼릭 링크 미러, 발행 SKILL.md) 결과를 **Deploy 없이** 복구하는 패키지 수준 패스. DeployerOption이 아닌 형상을 의도적으로 골랐다 — 옵션이었다면 배포 경로에서도 살아나 수리 기능의 부작용으로 배포 동작이 바뀐다. 항목별 의미는 미러 생산자의 것을 재사용하므로 생산자와 갈라질 수 없다 |

두 방출기는 **비테스트 코드에서 아무도 import 하지 않습니다**(`internal/template/agentemit`,
`internal/template/commandemit` 둘 다 패키지 단위 fan-in 0). 소비자는 빌드 타깃
(`make agents-emit` / `make commands-emit`)과 골든 테스트이며, 이는 고아가 아니라
**빌드타임 도구**라는 뜻입니다 — 아래 §네거티브 스페이스의 "호출자 0"과 구별해야 합니다.

### codex 미러 탭 — `internal/web`의 편집하지 않는 표면

`internal/web/codexmirror.go`는 codex 탭의 **행 모델**이며, Audit·MCP 탭에 사는 codex 설정의
**읽기 전용 미러**입니다. 어떤 필드도 옮기지 않습니다 — 미러된 필드는 각자의 소유 탭에서
그대로 선언·렌더·편집되고, 이 파일은 미러가 무엇을 보여주고 어디를 가리키는지만 정합니다.
행은 `settings.AllFields()`와 공유 MCP 도구 카탈로그에 대한 **술어로 파생**되며 손으로
열거하지 않습니다(손 열거는 codex 필드가 하나 늘어나는 순간 조용히 어긋납니다).

렌더 쪽 패널은 `a-h/templ`이 짝 `.templ` 소스에서 생성한 산물입니다.
이 패널은 `name` 속성을 가진 폼 요소를 하나도 내지 않으며, 그 금지는 숨은 bool 동반자
`<name>__present`에도 그대로 걸립니다 — 모든 패널이 한 폼 안에 살고 탭은 표시 전환일 뿐이라
비활성 패널도 함께 제출되기 때문입니다.

---

## data / persistence

| 패키지 | 비테스트 | 책임 | 주요 하위 패키지 |
|---|---|---|---|
| `internal/config` | 59 | 프로젝트 설정의 SSOT. 섹션별 YAML 로딩·캐시·검증(`go-playground/validator`), `envkeys.go`의 환경변수 상수 카탈로그, 기본값. **트리 최대 fan-in (22)** 이며 `types.go`(82KB)·`defaults.go`(59KB)가 트리에서 가장 큰 손 저작 파일 축에 든다. **이 판에서 `loader_identity.go`가 더했다**(card t1139) — `project.name`/`user.name` 단일 키 판독기, `LoadGitMode`를 본떴다. 소비자는 update 렌더 경로다: update가 Loader 생명주기 밖에서 템플릿 컨텍스트를 만들고, template-sync 경로에서 managed cleanup이 `.moai/config`를 지우기 전에 값을 읽어야 하기 때문이다. 이전 판에서 두 판정 파일이 더해졌다 — `harness.go`(`llm.harness` 허용 값의 닫힌 집합, 기본 claude)와 `loader_workflow_disposition.go`(git-strategy workflow 허용 4값, 3방향 disposition, flow별 통합 대상 표). workflow에 `jev.enabled`가 더해지며 설정 캐시 스키마 버전도 5로 올랐다 | `atomicfile`, `toolpolicy` |
| `internal/session` | 23 | 세션 레지스트리·체크포인트·페이즈·앵커·태스크 원장. PID 조회를 OS별 파일로 분기. **이 판에서 `anchor_lock_holder.go`가 더했다** — 잠금이 있는지뿐 아니라 **누가** 그 잠금을 쥐고 있는지 알아야 하는 호출자를 위한 읽기 전용 accessor 모음(자기 락을 쓰는 런처가 자기 pid를 알아보거나, 락 보유자가 확실히 죽었을 때만 대체하려는 경우). 아무것도 새로 판정하지 않고 `AnchorDecision`이 쓰는 것과 같은 `parseLockPID`/`lockAnchorVerdict` 쌍에 위임해, `AnchorDecision`의 세 호출자와 이 런처가 잠금의 의미에서 갈릴 수 없게 한다. `internal/cli/worktree/remove.go`가 이를 소비 — 등록된 앵커 세션이 없어도 git worktree lock 자체가 앵커라면(`moai codex -w` 세션은 그 외 아무 데도 등록하지 않는다) git 자신의 에러 대신 앵커 출처를 이름 붙여 거절한다. 이전 판에서 레지스트리 경로 해석이 앵커를 얻었다 — `RegistryPathFor`가 `internal/stateanchor` seam으로 리포지터리의 primary checkout을 가리켜 워크트리마다 레지스트리가 갈라지지 않는다(`DefaultRegistryPath`는 프로젝트 상대라, 전에는 워크트리에서 쓴 등록이 primary의 파일에 보이지 않았다). PID 스탬프는 물려받은 `MOAI_SESSION_PID`를 무조건 존중하지 않고 조상 사슬 검사를 통과할 때만 받는다 | — |
| `internal/settings` | 10 | `moai web` 콘솔과 `moai profile setup` TUI 두 표면이 공유하는 설정 스키마. 이 판에서 Jev opt-in(`workflow.jev.enabled`)이 같은 `ApplySchemaEdits` seam 위의 네이밍 진입(`jev.go`의 `SetJevEnabled` — `moai init` 위자드가 쓴다)을 얻었다. 두 번째 쓰기 경로가 아니라 같은 경로의 두 번째 이름이다 | `agentfm`, `yamlpatch` |
| `internal/sessionmsg` | 7 | 단일 머신 세션 간 메시징 브로커 (envelope 스키마) | — |
| `internal/chain` | 4 | **워크트리 세션 origin-trail 체인** — `.moai/state/chain/events.jsonl`에 spawn 경계·`session_id` 백필·완료 엣지를 append-only JSONL 계보 트리로 적는다. 쓰기는 매번 `O_APPEND`로 열어 커널이 동시 append를 직렬화하게 두며, 읽고-고치고-쓰는 주기가 없다(전체 파일을 올려 변형하지 않는다). 깨진 줄은 스트림을 중단시키지 않고 건너뛴다. 목적은 depth-N 워크트리에 `/clear` 이후 재진입한 사람이 grep·스크롤백 고고학 없이 origin·완료·재개 지점을 바로 복원하는 것이다 | — |
| `internal/homestate` | 21 | HOME 상태의 경로·SQLite 스키마·동시성 계약. 프로젝트별 `todo/backlog.db`, `factory/factory.db`, 전역 `run/profile-leases.db`, migration marker·admission lock, PID 지문과 runtime census를 소유한다. Unix `flock`과 Windows `LockFileEx`를 같은 계약으로 제공한다. **이 판에서 `factory_run_retire.go`가 더했다**(SPEC-FACTORY-RUN-RETIRE-001) — "이 런의 소유자가 아직 살아 있는가?"의 답인 `OwnerClassification`(`live`/그 외). 집합은 의도적으로 열려 있다 — 은퇴는 양성 `OwnerDead`에만 걸리게 게이트돼 있어서(`retirable`), 나중에 추가되는 값은 기본으로 은퇴를 거절한다(폴스루가 아니다). 이전 판에서 비정준 트리 게이트(`internal/homestate/noncanonical_tree_guard.go`)가 admission lock 획득과 marker 설치 두 변이 진입점 앞에 배선됐다 — 고립된 HOME을 가진 워크트리에서 호출하면 실제로는 canonical 루트의 살아 있는 상태를 건드리게 되므로, 락 파일을 만들기 **전에** 거절한다. **이 판에서 darwin 전용 프로세스 지문(`process_fingerprint_darwin.go`)이 더했다** — `golang.org/x/sys/unix.SysctlKinfoProc("kern.proc.pid", pid)`로 프로세스 시작시각을 읽어 PID 재사용을 가른다(`unix`/`windows` 빌드 태그 형제와 나란히 플랫폼 분기 완성) | — |
| `internal/factorymsg` | 8 | SQLite(`modernc.org/sqlite`) 기반 factory 전용 런-스코프 메시지 브로커. 패키지 주석이 레거시 `internal/sessionmsg`를 읽거나 이관하지 않는다고 명시한다. `store.go`의 `Peer`(project/run/backend/role/slot/session/generation/PID/process-start)로 발신·수신자를 식별하고, `Send`/`Claim`/`ReadBody`/`RecordDisposition`/`Receipt`로 클레임 기반 at-least-once 전달을 구현한다. `RegisterLaunchPending`/`BindLaunchPending`/`RollbackLaunchPending`은 프로세스만 살아 있고 세션 UUID가 아직 없는 factory launch 창을 `launch-pending:` 접두 provisional 키로 담아, SessionStart 훅이 실제 세션 UUID로 치환한다(owner-preserving upsert). `internal/cli/mcp_factory_msg.go`(5개 MCP 도구)와 `internal/hook/factory_messages.go`가 비테스트 소비자다. **이 판에서 1→8로 늘며 lane worktree handoff의 브로커 절반을 얻었다**(SPEC-FACTORY-LANE-WORKTREE-HANDOFF-001) — `dispatch.go`(전달 기록·펜스드 결과 적용·재배정·regrant), `handoff.go`(핸드오프 상태 기계 본체), `handoff_abandon.go`(소스-소유자 생존 프로브를 건네받는 종결 트랜잭션), `handoff_bind.go`(M3 rebind 사유 상수, STALE_* 리다이렉트 메타데이터 포함), `handoff_relocation.go`(`HeadlessRelocation` — 공식 app-server 결과와 컨트롤러의 대상 재판독을 M3로 실어 나르는 증거 타입), `schema_migrate.go`(메시지 테이블을 스키마 버전 2 — 발신자 세션 UUID·세대와 무관하게 `(project_key, run_id, sender_slot, idem_key)`로 idempotency를 스코프하는 lane-scope 마이그레이션). `factory_run_retire.go`는 별개로 SPEC-FACTORY-RUN-RETIRE-001의 `LeadPeerIdentity`(런의 등록된 role='lead' peer가 실어 나르는 프로세스 identity, REQ-006 fallback 소스) | — |
| `internal/guardstate` | 4 | 가드 생존성의 상태 모델·매니페스트 | — |
| `internal/manifest` | 3 | 파일 provenance 추적과 변경 감지. **이 판에서 `types.go`가 더했다**(card t1100) — 다섯 번째 provenance 값 `GeneratedManaged`(템플릿 배포자가 아니라 MoAI 생성기가 쓰는 파일 하나에 사용자 소유 부분과 MoAI 소유 부분이 공존할 수 있음을 표시, 소유권은 파일 전체가 아니라 `FileEntry.Parts` 단위로 결정)와 그 파트 스키마(`PartKind` — whole-file/hook-handler/json-key/toml-table/toml-key, `PartOrigin` — created/preexisting)가 더해졌다. `internal/codexwiring/ownership.go`가 소비한다 | — |
| `internal/tokenusage` | 3 | Claude Code 트랜스크립트 JSONL을 파싱해 토큰 사용량을 귀속·기록. **호출자 0 — 아래 §네거티브 스페이스** | — |
| **`internal/auditreceipt`** | **2** | **이 판에서 새로 생겼다.** `.moai/state/audit-receipts/` 아래 세 종류의 런타임 기록을 소유한다 — 감사 도구 호출 1건당 영수증, 감사자 서브에이전트 1건당 시작 마커, 거부된 PASS 1건당 거부 기록. 존재 이유를 패키지 주석이 직접 적는다: **PASS 판정은 에이전트가 쓴 텍스트이고, 텍스트는 도구가 실제로 불렸음을 보일 수 없다** — 그것을 기록할 수 있는 것은 런타임뿐이다. 쓰기는 임시 파일 + rename 원자 교체이고, 기록 1건이 파일 1개다(JSONL 아님). 트리 루트 판정은 `git rev-parse --show-toplevel`(2초 타임아웃) + 심볼릭 링크 해석이며 `CLAUDE_PROJECT_DIR`를 **의도적으로 무시**한다 — 워크트리 세션에서 그 변수는 primary 체크아웃을 가리키기 때문이다 | — |

### `internal/homestate` — 프로젝트 로컬 파일과 HOME DB 사이의 안전 경계

정본 경로는 `~/.moai/db/<project-key>/todo/backlog.db`와
`~/.moai/db/<project-key>/factory/factory.db`이며, 프로필 점유는 프로젝트와 무관한
`~/.moai/run/profile-leases.db`에 둡니다. `project-key`는 정규화한 프로젝트 루트에서
결정되므로 여러 워크트리가 같은 프로젝트 DB를 공유합니다.

이 패키지는 경로만 계산하지 않습니다. migration marker가 있는 동안 SessionStart, MCP 서버,
Factory 런타임 진입을 동일한 admission lock 아래에서 거절하고, 실제 이전은 두 번의 런타임
census가 모두 0일 때만 허용합니다. Factory 인계는 v2 lease와 token CAS로 만료 재점유와 ABA를
막고, 주입 뒤 crash는 at-least-once 경계로 남깁니다. 이 설명은 구현된 계약이며, 운영 DB에
`--apply`가 실행됐다는 뜻은 아닙니다.

### `internal/settings/yamlpatch` — 보존 쓰기 경로

`internal/settings`의 책임은 "두 표면이 공유하는 **스키마**"지만, `yamlpatch`가 지는 것은
스키마가 아니라 **쓰기 방식**입니다. `ConfigManager.Save()`의 typed struct 재직렬화는
YAML 주석 전량과 미모델링 키(예: `workflow.yaml`의 `team.patterns`, role-profile의 `effort`)를
파괴합니다. `yamlpatch`는 `gopkg.in/yaml.v3` 노드 트리를 수술해 대상 스칼라만 upsert 하고
나머지 문서 구조를 보존하며, **`Save()` 경로가 없는 8개 섹션**(workflow, harness, ralph,
research, feedback, observability, security, db)의 **유일한 쓰기 경로**입니다.

한계도 요건의 일부입니다 — yaml.v3 Encoder는 재직렬화 시 일부 포매팅(빈 줄, 긴 스칼라
줄바꿈)을 정규화할 수 있어 byte-stability는 보증이 아니라 **검증 대상**이고, 섹션별 골든
round-trip 테스트가 그 범위를 고정합니다. 노드 삭제는 지원하지 않습니다.

---

## infrastructure / platform

| 패키지 | 비테스트 | 책임 | 주요 하위 패키지 |
|---|---|---|---|
| `internal/lsp` | 35 | LSP 클라이언트 스택. `charmbracelet/x/powernap` 전송 위에 gopls 브릿지, 서브프로세스 수명 관리, TTL 진단 캐시, 다중 서버 집계 | `core`, `gopls`, `transport`, `subprocess`, `cache`, `config`, `aggregator`, `hook` |
| `internal/astgrep` | 13 | ast-grep(`sg`) CLI 래핑 기반 AST 분석·룰 시드 | — |
| `internal/github` | 10 | `gh` CLI 기반 PR/이슈 오퍼레이션 | `workflow` |
| `internal/runtime` | 11 | 토큰 서킷 브레이커, 감사 캐시/게이트/리포트, 클록 | `gobin` |
| `internal/sandbox` | 8 | 명령 실행을 감싸는 임시 샌드박스 실행 계층 | — |
| `internal/git` | 8 | `core/git` 위에 얹은 상위 유틸리티. **루트 패키지 자체의 비테스트 import 는 0이고, 최상위 집계 fan-in 1은 전부 하위 `convention`이 받은 것**이다 — 아래 §네거티브 스페이스 | `convention` |
| `internal/update` | 7 | 바이너리 셀프 업데이트 (체커 + 오케스트레이터) | — |
| `internal/shell` | 6 | 셸 탐지와 환경 구성 | — |
| `internal/telemetry` | 5 | 텔레메트리 수집·전송 | — |
| `internal/codexadapter` | 5 | Codex 훅 표면 ↔ MoAI 훅 스키마 번역 | — |
| `internal/tmux` | 4 | 병렬 SPEC용 tmux 세션 관리 | — |
| `internal/worktree` | 4 | 워킹 트리 상태 가드 프리미티브 | — |
| `internal/resilience` | 3 | 외부 서비스 연동용 서킷 브레이커 | — |
| `internal/glmcred` | 1 | GLM API 자격증명 단일 구현 | — |
| `internal/jev` | 1 | TypeSafe System One 호출 경로의 유일 구현 — 요청 조립, HTTP 전송, 응답 해석, 사용량 산출. 표준 라이브러리만 import 하므로(순환에 끼지 못하고 `internal/cli`·`internal/web` 양쪽이 쓸 수 있다) 모델 id·엔드포인트는 컴파일 상수 핀이다. 불가능한 답은 에러가 아니라 **값**이다(`no-credential` 등 Availability — 에러 반환이 퍼지면 어딘가에서 비종료가 되기 때문)이고, 비활성 게이트는 호출자 위가 아니라 **패키지 안쪽**에 있다(끈 상태에서는 요청 자체를 조립하지 않는다). 표시 전용 — 파일·큐·git에 쓰는 의존성이 하나도 없다. **이 판에서 와이어 스키마가 벤더가 실제 문서화한 형태로 다시 쓰였다** — `questions`는 리스트가 아니라 호출자 id로 키잉된 객체(값은 `type`/`instructions`/`criteria`)이고, `noul`은 불리언이 아니라 확률값(`Answer.Probability`가 곧 참 확률이며 참/거짓 컷은 호출자 몫), choice/score 질문은 `Levels`/`Probabilities`/`Legend`까지 왕복한다(`toWire`/`fromWire`가 변환을 맡는다). `UserAgent` 상수도 이 판에서 더해졌다 — 벤더 엣지가 일부 기본 User-Agent(예: Python urllib 기본값)를 403으로 거부한 관측 때문에 클라이언트가 자기 이름을 명시한다 | — |
| `internal/jevcred` | 1 | **이 판에서 새로 생긴 패키지.** `~/.moai/.env.typesafe` 자격증명의 쓰기·읽기 단일 구현 — `internal/glmcred`를 의도적으로 모델 삼았다(쓰기 시 chmod 조임, 네 글자 미만 공개 하한까지). 스키마 `AllFields()`에 **일부러 없는** 필드라 어떤 스키마 순회 루프도 그 값을 읽거나 렌더하지 못한다. 최상위 fan-in 2(`internal/cli`, `internal/web`) | — |
| `internal/gitenv` | 1 | 자식 프로세스가 **어느 리포지터리에** 작용할지를 정하는 git 환경변수(`GIT_DIR`, 커밋 경로의 `GIT_INDEX_FILE`)를 지운다. 훅이 내보낸 이 변수들은 작업 디렉터리보다 우선하므로 `cmd.Dir`만으로는 격리가 되지 않는다(GH #1691). 한 패키지의 수리가 형제에게 닿지 않았던 결함을 막으려고 독립 패키지로 둔 것을 패키지 주석이 밝힌다. 최상위 fan-in 2(`internal/cli`, `internal/hook`) | — |
| `internal/binlag` | 1 | 설치된 바이너리 지연 판정. 이 판에서 판정이 하나 더 갈라졌다 — 바이너리 커밋이 비교 ref의 **엄격 자손**인 경우(`StatusAhead`)는 무관 계열에서 떼어져 자기 문장을 얻었다. 스테일이 아니라 비교가 무의미한 경우이며, notice는 재빌드가 아니라 비교 대상을 가리킨다 | — |
| `internal/mirrornotice` | 1 | 스킬 미러 결과를 사용자 알림으로 전환 | — |
| `internal/report` | 1 | 루트에 Go 파일 없음 — 하위 `planhtml`만 존재 | `planhtml` |

---

## cross-cutting

| 패키지 | 비테스트 | fan-in | 책임 |
|---|---|---|---|
| `internal/defs` | 5 | 12 | 디렉터리명·파일명 등 프로젝트 전역 상수 |
| `internal/atomicfile` | 5 | 11 | 크로스 플랫폼 원자적 파일 교체 (unix/windows 분기) |
| `pkg/models` | 4 | 8 | 공유 데이터 모델. 외부 공개 2개 패키지 중 하나 |
| `pkg/version` | 2 | 5 | 빌드타임 버전 정보 (ldflags 주입) |
| `internal/lockfile` | 2 | 2 | 크로스 플랫폼 advisory 파일 락 |
| `internal/paths` | 1 | 12 | `~/.moai` 디렉터리 해석의 단일 지점 |
| `internal/execerr` | 1 | 7 | 서브프로세스 종료 실패를 안전하게 출력 가능한 형태로 유지 |
| `internal/stateanchor` | 1 | 3 | **상태 앵커 seam** — 아래 상세 |
| `internal/measure` | 1 | 2 | 의존성 없는 순수 leaf — 프로젝트 헬스 지표 |
| `internal/timing` | 1 | 0 | 테스트용 보정된 지연 상한 (비테스트 fan-in 0) |
| `internal/codextools` | 2 | 0 | 네이티브·지연 디스패처 도구 레지스트리를 인증된 대화 하나에 묶는다(패키지 주석: 도구나 RPC를 실행하지 않는다). `github.com/santhosh-tekuri/jsonschema/v6`를 직접 쓰는 유일한 패키지다. **비테스트 fan-in 0 — 아래 §네거티브 스페이스** |
| `internal/skills` | 0 | 0 | 프로덕션 코드 없음 |

### `internal/stateanchor` — 상태를 어디에 쓸지 정하는 단일 seam

`.moai/state/`를 읽고 쓰는 **모든** 표면이 앵커(프로젝트 루트)를 여기서 받습니다. 이 seam
이전에는 각 표면이 세션이 서 있던 자리에서 제 앵커를 유도했고, statusline의 텔레메트리
쓰기가 `workspace.current_dir`에 앵커돼 **cd 한 디렉터리마다 `.moai` 디렉터리가 하나씩
남았습니다**(GH #1694). 지금 이 seam을 통과하는 소비자는 statusline 텔레메트리 쓰기,
보드 루트와 그 landed·github-counts 소비자, goal 상태 읽기, CLI 설정 캐시 사슬입니다.

우선순위 사슬은 요건으로 **고정**돼 있습니다(단계 삽입·재정렬은 요건 변경입니다):

1. stdin `workspace.project_dir`
2. `worktree.original_cwd`
3. 세션 디렉터리에 대한 리포지터리 해석 — `core/git`의 공통 디렉터리 해석을 재사용하며,
   그 부모는 리포지터리의 모든 체크아웃과 워크트리에 대해 하나다

셋 다 실패하면 앵커는 빈 문자열이고, 호출자는 상태 쓰기·읽기를 건너뛴 채 렌더를 정상
완료합니다(프로젝트가 없으면 상태도 없다). **세션의 현재 디렉터리는 앵커가 아니며**
리포지터리 walk-up의 입력으로만 쓰입니다. 표시 이름 유도는 별개 관심사로
statusline의 `extractProjectDirectory`에 남아 있습니다. statusline 쪽 어댑터가
`internal/statusline/state_anchor.go`입니다.

---

## 네거티브 스페이스

결정적 도구가 만들 수 없는 관찰입니다. 이 절이 이 문서에서 가장 값이 나가는 부분입니다.

### 테스트가 없는 패키지 — 4개, 전부 main

`go list -f '{{.ImportPath}} {{len .TestGoFiles}} {{len .XTestGoFiles}}' ./...` 기준입니다.

| 패키지 | 판단 |
|---|---|
| `cmd/moai` | 정당. 20줄 위임 로직이고 `internal/cli`에 통합 테스트가 있다 |
| `cmd/t657-merge` | 일회성 큐 병합 도구(카드 t657). 로직은 테스트가 있는 `internal/kanban` 저장소 API를 재사용한다 |
| `internal/template/scripts` | 정당. 빌드타임 생성 도구 main |
| `scripts/convert-nextra-to-hextra` | 일회성 문서 변환 스크립트 |

테스트 비율이 1.67:1이므로 **테스트 부족은 이 코드베이스의 약점이 아닙니다.**

### 프로덕션 코드 없이 테스트만 있는 자리 — 2개

- **`internal/skills`** — `workflow_split_test.go` 하나뿐, 비테스트 파일 0개.
- **`internal/tui/golden`** — `doc.go`와 `index_test.go`뿐.

둘 다 "테스트가 다른 곳(템플릿 트리, 골든 파일)을 검증하는데 담을 자리가 없어 만들어진 빈
패키지"로 보입니다. 필요한 것은 패키지가 아니라 테스트 파일을 둘 자리입니다.

> **이 자리에 있던 세 번째 항목이 사라진 경위.** 앞 판은 `internal/orchestration`을 같은 계열로
> 적었지만, 그 디렉터리와 유일한 파일(`naming_manifest_contract_test.go`)은 커밋 `ae3075280`에서
> 삭제됐습니다. 그 테스트가 고정하던 매니페스트 JSON은 어느 커밋에도 존재한 적이 없고
> (`git log --all` 출력 없음), 준비 대상이던 코드 경로는 이미 철회된 상태였습니다. 즉 "테스트만
> 있는 빈 패키지"가 아니라 **한 번도 생성되지 않은 산출물을 검증하던 테스트**였습니다.
> 삭제는 테스트를 지워 초록을 만들지 않는다는 규칙에 대한 운영자 승인 예외로, 그 파일 하나에만
> 적용됐습니다.

### 비테스트 코드에서 아무도 import 하지 않는 패키지

`.Imports`(테스트 import 제외) 기준입니다. **넷으로 갈립니다** — 빌드타임 도구, 테스트 시점
가드, 의도된 고아, 그리고 새로 생긴 것.

네 갈래를 나누는 것은 fan-in 0이라는 수치가 아니라 **그 0이 정상인 이유**입니다. 빌드타임
도구는 `make` 타깃이 부르고, 테스트 시점 가드는 `go test`가 부르며, 고아는 아무도 부르지
않습니다. 세 경우 모두 `.Imports` 집계에서는 구별되지 않으므로, 이 표의 「상태」 칸이
판별식입니다.

| 패키지 | 상태 |
|---|---|
| `internal/template/agentemit` · `internal/template/commandemit` | **고아가 아니다.** 소비자가 `make agents-emit` / `make commands-emit` 빌드 타깃과 골든 테스트다. 방출기는 빌드타임 도구이므로 런타임 fan-in 0이 정상 상태다 |
| **`internal/git`** (루트 패키지) | **이 판에서 새로 잡혔다.** `core/git` 위의 상위 유틸리티 8 파일인데 루트 패키지를 import 하는 비테스트 코드가 0이다. 실제로 import 되는 것은 하위 `internal/git/convention` 하나뿐이며(`internal/cli` → `internal/git/convention`), 최상위 집계 fan-in 1은 그것이다. 소비자가 `core/git`로 직접 내려가면서 중간 계층만 남은 모양으로 읽힌다 — 확인이 필요한 관찰이며, 이 문서가 답을 주지는 않는다 |
| **`internal/tokenusage`** | 완전 고아. 유일한 언급이 `internal/spec/audit.go`의 주석인데, 상수를 공유할 수 있지만 파서를 self-contained로 두려고 로컬 상수를 쓴다는 내용이다 — **공유 의도가 있었으나 거부된 뒤 아무도 쓰지 않게 된** 패키지다. `moai tokens` 서브커맨드조차 이것을 쓰지 않는다 |
| `internal/github/workflow` | GitHub Actions 워크플로 검증기. import 하는 코드가 없다 |
| `internal/harness/harnessrun` · `seeds` · `throttle` | harness 하위인데 형제 패키지 어느 것도 참조하지 않는다 |
| **`internal/harness/rosterguard`(4 파일) · `internal/harness/cellguard`(1 파일)** | **이 판에서 새로 잡혔고, 위 고아들과 종류가 다르다.** 둘 다 **런타임 경로가 아예 없는 테스트 시점 가드**다 — CLI·훅·MCP 어디에도 배선돼 있지 않고, 자기 테스트가 살아 있는 저장소 트리를 직접 읽어 드리프트를 찾으면 `go test`를 빨갛게 만드는 것이 유일한 발화 경로다. `rosterguard`는 에이전트 로스터 목록이 사이트마다 어긋나는 것을, `cellguard`는 docs-site의 profile-matrix 셀 표가 `internal/template/profile_matrix.go`와 어긋나는 것을 본다. **fan-in 0이 결함이 아니라 설계**이며, 그 점에서 `agentemit`·`commandemit`과 같은 부류이고 `tokenusage`와는 반대다 |
| `internal/migration/migrations` | `internal/cli/migration_m3_test.go`가 명시한다 — `internal/cli`가 이 패키지를 import 하지 않으므로 m001/m002는 `Register()`를 호출하지 않는다. blank import로 등록되는 패턴인데 그 blank import가 어디에도 없다. 테스트가 이 사실을 *기술*할 뿐 *거부*하지 않는 것이 문제다 |
| **`internal/jevmeasure`** | **이 판에서 새로 잡혔다.** 측정 장치인데 소비자가 0이다 — 자기 테스트조차 이 패키지를 import 하는 밖의 코드가 없다. 이것이 설계다: 측정 게이트가 아직 실행되지 않았고(게이트 미실행 상태), 소비자는 게이트가 통과해야 존재 허가를 받는다. `rosterguard`·`cellguard`의 「테스트 시점 가드」와 다른 이유의 0이다 — 이쪽은 **아직 시간이 안 온** 0이다 |
| **`internal/codextools`** | **이 판에서 새로 잡혔다.** 도구 레지스트리를 대화에 묶는 2 파일 패키지인데 비테스트 소비자가 0이다. 트리에서 `jsonschema/v6`를 쓰는 유일한 자리이기도 해서, `go.mod`는 그 모듈을 `// indirect`로 적고 있다(§ `dependencies.md` 이례적인 것 7). 이전 측정 트리 이후 삭제된 gateway 계열 패키지들과 같은 시기의 산물로 보이지만, 그 인과는 이 문서가 확인하지 않았다 |
| `internal/cli/taskledger` · `internal/cli/ptycaptest` · `internal/lsp/aggregator` · `internal/hook/testutil` · `internal/timing` · `internal/tui/golden` | 테스트 전용 소비자만 갖는 leaf. `ptycaptest`는 PTY 렌더 캡처 테스트 드라이버다. `timing`은 이름이 그 의도를 말한다 |

### 빈 디렉터리 — 3개

`.gitkeep`만 있고 Go 코드가 없습니다.

- `internal/core/integration/`
- `internal/core/migration/` — 실제 마이그레이션은 `internal/migration`에 있습니다
- `internal/foundation/trust/` — TRUST 5 구현은 `internal/core/quality/trust.go`에 있습니다

뒤의 둘이 특히 위험합니다. **실제 구현이 다른 곳에 있는데 자리표시 디렉터리가 그 이름을
선점**하고 있어서, 다음 사람이 여기에 코드를 넣으면 두 벌이 생깁니다.

### 중복된 능력

- **`internal/atomicfile`(18개 파일에서 사용) vs `internal/config/atomicfile`(10개 파일에서 사용).**
  둘 다 살아 있고 둘 다 원자적 파일 쓰기를 합니다. 전자는 `Replace` + unix/windows 분기 + 읽기
  헬퍼, 후자는 `Write` + guard입니다. **한 프로세스 안에서 서로 다른 fsync/rename 전략의 두
  구현이 같은 디렉터리를 건드릴 수 있습니다.**
- `internal/merge`(3-way 머지 엔진) vs `internal/cli/update/merge`(업데이트 오케스트레이션 머지) —
  층위가 다르지만 grep 시 매번 함께 걸립니다. `internal/report/planhtml` vs
  `internal/cli/update/report`도 같은 모양입니다.
- **설정 드리프트 판정이 두 곳에 있습니다** — `internal/kanban/settings_drift.go`(도메인 절반)와
  `internal/cli/integration_settings_drift.go`(CLI 절반). 이것은 중복이 아니라 의도된 분할이며,
  두 표면(`acquire` precondition / `preflight` 독립 verb) 중 어느 쪽도 뺄 수 없다는 것이 그
  파일들의 주석이 적어 둔 설계입니다.

### 경계가 잘못 그어진 것

- **`internal/cli/preference`와 `internal/cli/specid`는 CLI 표면이 아닙니다.** `preference`는
  AskUserQuestion 결정 메모리 레이어, `specid`는 SPEC-ID sanitizer leaf입니다. 둘 다
  `internal/cli` 밖에서 import 당하고, 그것이 `dependencies.md`의 집계 상호 참조 두 쌍을
  만듭니다. 최상위로 승격하면 두 쌍이 사라집니다.
- **`internal/hook/session_start.go`가 67KB**입니다. 옆에
  `session_start_compact.go` · `_factory.go` · `_kanban.go` · `_guard_liveness.go` ·
  `_binary_lag.go` 등이 이미 따로 있는데도 그렇습니다. 세션 시작은 이미 자기 패키지가 되기에
  충분한 크기입니다.
- **`internal/cli/mcp_codex.go`가 97KB**로 CLI 최대 파일입니다. `internal/codexadapter`와
  `internal/codexwiring`이 이미 있는데도 로직 대부분이 CLI 파일에 남아 있습니다.
- **`internal/goal` / `loop` / `ralph`** — 6 / 6 / 1 파일이며 셋이 한 루프 서브시스템입니다.
  SPEC이 셋이었다는 것 외에 경계가 셋인 근거가 보이지 않습니다.
  **`guardliveness`(4) / `guardstate`(4)** 도 8개 파일을 둘로 나눌 분량이 아닙니다.
  **`internal/stateanchor`(1)** 도 같은 계열이지만 이쪽은 정당화가 있습니다 — 여러 표면이
  공유해야 하는 단일 결정 규칙이라 어느 소비자 밑에도 둘 수 없습니다.

### 트리에서 가장 큰 비테스트 파일은 손으로 쓴 것이 아닙니다

```
$ find internal cmd pkg -name '*.go' -not -name '*_test.go' -exec ls -l {} + | sort -k5 -rn | head -6
182575 internal/web/fieldsets_templ.go     (생성, 178KB)
126699 internal/web/screens_templ.go       (생성, 124KB)
 99355 internal/cli/mcp_codex.go           (손 저작 — CLI 최대, 97KB)
 83999 internal/config/types.go            (손 저작, 82KB)
 68985 internal/hook/quality/gate.go       (손 저작, 67KB)
 68523 internal/hook/session_start.go      (손 저작, 67KB)
```

앞선 판은 `internal/hook/session_start.go`(당시 67KB)를 "트리 최대 비테스트 Go 파일"이라고
적었습니다. **지금은 사실이 아닙니다** — 상위 둘이 `a-h/templ` 생성 산물이고, `session_start.go`는
67KB로 6위입니다. 크기 순위를 읽을 때는 생성 파일과 손 저작 파일을 갈라 세어야 합니다.

### 폐기 표식이 코드로 남은 것

- `internal/cli/root.go` — `newHarnessCmd()`가 은퇴 마커로서 컴파일 가능 상태로만 남아 있고,
  트리에 등록되지 않습니다. `TestHarnessFactoryStillCompiles`가 이를 고정합니다.
  죽은 코드를 테스트가 살려두는 구조입니다.
- `internal/hook/retired_events.go` — 은퇴한 훅 이벤트 목록. 소비자가 테스트 스위트와
  (등록되지 않는) `migration/migrations` m002뿐입니다.

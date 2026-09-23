---
id: SPEC-DUAL-HARNESS-RECOVERY-001
document: research
created: 2026-09-23
updated: 2026-09-23
author: manager-spec
card: t1100
---

# Research — SPEC-DUAL-HARNESS-RECOVERY-001

## §A 출처

- 읽기 전용 조사 세 편: `.moai/reports/t1100/research-m4-codexwiring.md`(A), `research-m3-launcher-roles.md`(B), `research-msg-factory-t1082.md`(C). 모두 트리 `d87e9af2e` 기준.
- 이 문서의 모든 줄 번호는 같은 트리에서 이번에 다시 읽어 확인했다. 조사 문서의 주장 중 확인되지 않았거나 틀린 것은 §C에 따로 적는다.
- t1082 SPEC(`.claude/worktrees/t1082/.moai/specs/SPEC-FACTORY-LANE-WORKTREE-HANDOFF-001/`)은 읽기만 했다.

## §B 확인된 사실

### B.1 codexwiring (A 확인)

| 사실 | 근거 |
|---|---|
| 쓰기 대상: `.codex/hooks.json`, `.codex/config.toml`, 사이드카 `.moai/state/codex-wiring.json` | `internal/codexwiring/wire.go` `wireProject` (59~) |
| MoAI handler 판정은 명령 접두사 `moai hook ` | `codexwiring.go:48` `moaiHandlerPrefix` |
| 사이드카는 `HooksSHA256`, `ConfigSHA256` 두 값뿐 | `wire.go:18-21` `sidecarDoc` |
| `writeAtomic`은 CreateTemp → Write → Close → Rename. 저널·잠금·rename 전 재확인·기록 후 대조 없음 | `wire.go:224-245` |
| 제거·롤백·저널 코드 없음 | `grep -rniE 'unwire\|uninstall\|rollback\|journal' internal/codexwiring/` 결과 0줄 |
| 호출자는 init(`init.go` `wireCodexUnlessClaude`), update(`update_codex_wiring.go:23`), `moai tool enable codex`(`update_codex_wiring.go:46`), init offer(`codex_init.go:136`) | grep 결과 |
| `moai tool`에는 `enable` 하위 명령만 있고 `disable`은 없음 | `internal/cli/tool.go:19-26` |
| manifest는 파일 단위 provenance(`template_managed` / `user_modified` / `user_created` / `deprecated`) | `internal/manifest/types.go:12-29` |
| 기존 update 경로에 잠금·폐기 분류·백업 도구가 있음 | `internal/cli/update_cleanup.go:56` `acquireUpdateLock`, `:203` `backupDeprecatedPaths`, `:343` `classifyDeprecatedFile` |
| 이전 결정: 기존 `[mcp_servers.moai]`·`status_line`은 사용자 소유 | SPEC-CODEX-WIRING-001 REQ-CW-005 |
| 이전 결정: 배선 생성기는 `.codex/agents/**`를 건드리지 않음 | SPEC-CODEX-WIRING-001 REQ-CW-012 |
| (iter-2 추가) `hooks.json`은 문서 전체를 `json.MarshalIndent`로 다시 렌더한다. 섞인 entry는 `marshalEntry`로 재구성된다 | `hooks.go:122`, `:198-212` |
| (iter-2 추가) 최상위 `description`이 없으면 MoAI가 추가한다 | `hooks.go:115-117` |
| (iter-2 추가) `.codex/hooks/moai/` 네임스페이스 명령도 MoAI 소유로 보고 교체한다 | `hooks.go:138-141`, `:167` |
| (iter-2 추가) `[tui]`가 없으면 테이블 전체를 덧붙인다. 덧붙일 때 `appendSection`이 구분 빈 줄을 넣는다 | `configtoml.go:145`, `:221-229` |
| (iter-2 추가) `acquireUpdateLock`은 package cli 비공개, O_EXCL. `runUpdate`는 이 잠금을 쥔 채 배선 갱신을 부른다. cli가 codexwiring을 import한다 | `update_cleanup.go:56-88`, `update.go:275-279`, `:506`, `update_codex_wiring.go:14` |
| (iter-2 추가) codexwiring의 moai 내부 import는 `codexadapter`, `hook`뿐이다 | `grep -h '"github.com/modu-ai' internal/codexwiring/*.go` |
| (iter-2 추가) manifest 파싱 실패 시 `.corrupt`로 옮기고 새 manifest로 초기화한다 | `manifest.go:81-88` |
| (iter-2 추가) `moai doctor --fix`는 제안만 한다. 관련 검사 주석도 "never writes files" | `doctor.go:59`, `:1079` |

### B.2 launcher·worktree (B 확인)

| 사실 | 근거 |
|---|---|
| codex 동사 표 `{"", cli, status, app}` | `codex_launcher.go:87-92` `codexVerbRouting` |
| `-w` 새 트리: `.claude/worktrees/<name>`, 브랜치 `WT-<name>`, 이름 없으면 `codex-<session>` | `codex_launcher.go:398-429`, `session_worktree.go:41,47` |
| base: `git_strategy.worktree_base_branch` → 원격 기본 브랜치 | `codex_launcher.go:388-396` `codexWorktreeBase` |
| 생성 뒤 HEAD와 base 대조 없음 | 같은 함수 |
| direct launch는 `syscall.Exec`로 pid 유지 | `codex_direct_posix.go` `defaultCodexDirectLaunch` |
| lock·anchor·동시 writer 검사 없음 | `codex_launcher.go` 전체에 `AnchorDecision`·`worktree lock` 호출 없음 |
| 권위 있는 anchor 출처는 git worktree lock, 사유의 `pid <n>`으로 생존 판정, fail-closed | `internal/session/anchor_lock.go:1-15, 107-147` |
| `AnchorDecision` 사용처는 `clean.go`, `session_worktree_prmerge.go`뿐 | grep 결과 |
| `moai cc -w`도 진입 전 동시 writer 검사가 없음(Claude Code의 EnterWorktree가 lock을 씀) | `session_worktree.go` 150-260 |
| codex `-f`(factory) 진입은 이미 있음 | `codex_factory.go`(카드 t865) |
| codex `-k`는 없음. cc 쪽 파서 `parseKanbanFlag`(`kanban.go:86`), `parseLauncherEntry`(`factory.go:176`) | grep 결과 |
| (iter-2 추가) `moai worktree done`·`remove`의 anchor 가드는 레지스트리 기반 `LiveAnchoredSessions`이며 git worktree lock을 보지 않는다. lock-aware `AnchorDecision`은 `clean.go:136, 352`, `session_worktree_prmerge.go:217`에서만 쓴다 | `done.go:86, 284`, `remove.go:51` |
| (iter-2 추가) `remove`는 `git worktree remove [--force]`를 부른다(`--force`는 한 번) | `internal/core/git/worktree.go:103-115` |
| (iter-2 추가) 세션 종료 정리는 dirty·unpushed 판정 후 제거하며 anchor 판정은 없다. `auto_cleanup`이 꺼져 있으면 아무것도 하지 않는다 | `session_worktree.go:635-689` |
| (iter-2 추가) Windows direct launch는 자식을 시작하고 `cmd.Wait()`로 기다린다 | `codex_direct_windows.go:14-29` |
| (iter-2 추가) lock 사유의 pid 생존 판정은 `sessionProcessLiveness` seam이며 POSIX·Windows 구현이 있다 | `anchor_lock.go:59, 128-147`, `anchor_pid_unix.go`, `anchor_pid_windows.go` |

### B.3 agentemit (B 확인)

| 사실 | 근거 |
|---|---|
| `sandbox_mode` 기본 `workspace-write`, `mission-governor`·`super-advisor`만 `read-only` | `agents-codex.yaml:57-73`, 생성된 12개 TOML의 `sandbox_mode` 줄 |
| 12개 Codex 역할 = MoAI 11개 + `mission-governor` | `internal/template/templates/.codex/agents/moai/` 목록 |
| `sync-auditor` Claude 도구에 Write/Edit 없음, Bash 있음 | `sync-auditor.md` `tools:` |
| `plan-auditor` Claude 도구에 Write/Edit 있음 | `plan-auditor.md` `tools:` |
| `codex_task`의 쓰기는 프로젝트 opt-in(`workflow.codex.task.allow_write`)으로만 허용 | `codex_task.go:203-221` |
| (iter-2 추가) 역할별 MCP 서버 부여는 역할 TOML의 `[mcp_servers.moai]` 테이블로 표현되며 codex-cli 0.147.0에서 등록이 실측됐다(배열형은 거부). 한 서버 안의 도구 단위 필터는 표현 불가로 기록돼 있다 | `agents-codex.yaml:196-207` |
| (iter-2 추가) 생성된 12개 TOML 중 7개에 `[mcp_servers.moai]`가 있다: manager-develop, manager-lead, manager-docs, plan-auditor, super-advisor, manager-spec, sync-auditor | `grep -l '^\[mcp_servers.moai\]' internal/template/templates/.codex/agents/moai/*.toml` |
| (iter-2 추가) 현재 생성기 근거는 "sync-auditor must write its verdict file, so making it read-only would violate the current export mandate"라고 적는다. D6 결정 (a)로 이 근거를 고쳐야 한다 | `agents-codex.yaml:262-264` |
| (iter-2 추가) 부여하지 않은 역할이 프로젝트 `config.toml`의 전역 `[mcp_servers.moai]`를 물려받는지는 측정 기록이 없다 | 기록 부재(측정하지 않음) |

### B.4 factorymsg (C 확인)

| 사실 | 근거 |
|---|---|
| Kind: `dispatch_notice`, `status_request`, `status_report`, `blocker`, `receipt` | `store.go:26-30` |
| 멱등 범위 `UNIQUE(sender_session, idem_key)` | `store.go:282`, 재조회 `:624` |
| 같은 키·다른 본문 거부 | `store.go:625` |
| `Claim`은 `recipient_session`과 `recipient_generation`으로 거름. 재등록 시 기존 메시지 재주소 없음 | `store.go:672, 692`, `UPDATE ... recipient_generation` 없음 |
| 새 세션이 같은 slot에 등록하면 generation +1 | `store.go:337-366` `RegisterPeer` |
| claim token 불일치 receipt 거부, lease 만료 재전달 | `store.go:749-763`, `store_test.go:197-265` |
| 이전 결정: 브로커는 at-least-once, exactly-once 실행을 주장하지 않음 | SPEC-FACTORY-MIXED-HOOK-001 spec.md 77행 REQ-FMH-006 |

### B.5 factory live 테스트 (C 확인)

| 사실 | 근거 |
|---|---|
| 4조합 테스트는 `MOAI_FACTORY_LIVE=1`이 없으면 SKIP | `factory_live_test.go:72-80` |
| `MOAI_FACTORY_LIVE_CASE`가 다르면 SKIP이 아니라 `Fatalf` → 조합마다 따로 실행해야 함 | 같은 곳 |
| 현재 케이스는 메시지 왕복만 본다(카드 흐름·중단·재할당·늦은 응답 없음) | `factory_live_test.go:55-68` |
| `.github/workflows/`에 `MOAI_FACTORY_LIVE` 사용 없음 | `grep -rn MOAI_FACTORY_LIVE .github/workflows/` 결과 0줄 |
| t1074는 왕복 LIVE를 실행해 PASS 증거를 남겼음(카드 흐름은 아님) | SPEC-FACTORY-MIXED-HOOK-001 progress.md AC-FMH-010/011 행 |

## §C 조사 문서가 틀렸거나 빠뜨린 것

| 조사 | 주장 | 실제 | 영향 |
|---|---|---|---|
| A | `harness_fs.go:44-50`에 `.codex/**` 역방향 규칙이 없다 | 이 트리에는 `hideCodex`가 있어 claude 프로필 배포가 `.codex/`를 숨긴다(`harness_fs.go` `isHidden`, 커밋 `8925682d2`에서 추가). 다만 숨김은 배포에서 빼는 것이지 기존 파일 제거가 아니다 | 고아 문제는 그대로. REQ-DHR-007 근거 수정 |
| A | 기록 후 다시 읽지 않는다 | 사이드카를 쓰기 위해 `hooks.json`을 다시 읽는다(`wire.go` sidecar 분기). 그러나 렌더 결과와 비교하지 않는다 | "대조 없음"이 정확한 표현 |
| A | `config.toml`은 create-if-absent | 테이블·키 단위다. 테이블이 없으면 사용자 파일에 덧붙인다 → 한 파일에 사용자·MoAI 부분이 섞인다 | 부분 단위 소유 기록이 필요한 이유(REQ-DHR-001) |
| A | 재사용 후보로 manifest만 제시 | `update_cleanup.go`의 잠금·폐기 분류·백업도 재사용 가능 | 새 코드 감소 |
| B | 줄 번호 `:96`, `:298-330` | `:87-92`, `:356` | 인용 수정 |
| B | integration lock 위치 `internal/cli/integration.go:33-64` | 저장 구현은 `internal/kanban/integration_lock.go`, CLI는 `internal/cli/integration.go` | 인용 수정 |
| B | anchor 판정 재사용 언급 없음 | `internal/session/anchor_lock.go`가 핵심 재사용 대상(동시 writer 거부를 새 코드 없이 구현) | REQ-DHR-008/009 설계 근거 |
| B | codex factory 진입 언급 없음 | `codex_factory.go`에 `-f` 진입 존재 | `-k` 파서 재사용 설계의 선례 |
| B | read-only는 파일 쓰기만 막고 Bash 부작용은 거르지 않는다 | Bash는 Codex sandbox 안에서 돈다. read-only가 shell의 쓰기를 막는지는 측정되지 않은 가설이다. MCP 호출은 sandbox 밖(MCP 서버 프로세스)에서 돈다 | AC-DHR-012가 측정. MCP 경유는 `UNSUPPORTED` |
| C | exactly-once를 factorymsg 계약으로 본다 | t1074 REQ-FMH-006이 exactly-once 실행을 주장하지 않는다고 명시 | "한 번 반영"은 dispatch record 적용 단계로 한정(REQ-DHR-018) |
| C | 멱등 범위 문제 언급 없음 | 범위가 송신자 세션 UUID라 송신자 재시작 뒤 재전송이 중복됨 | REQ-DHR-017 |
| C | 재시작 시 메시지 처리 언급 없음 | 수신 generation이 바뀌면 이전 generation 앞 메시지가 claim 불가 상태로 남음(코드 판독, 미측정) | REQ-DHR-020, AC-DHR-015가 측정 |
| C | 경계 제안: "한 generation, 한 멱등 범위" | 제안 자체는 맞다. 초안 시점의 t1082 design.md §8은 "idempotency key에 handoff generation을 결합"으로 읽혔다. iter-2 시점에 t1082를 다시 읽으니 design.md 204행이 "Handoff generation은 key의 일부가 아니며… Key의 기준은 t1100이 소유하며 이 SPEC은 현행 스키마를 따른다"로 바뀌어 있고, spec.md REQ-FLH-009 본문은 "detecting duplicates by the current t1074 schema's idempotency key unchanged"다 | spec.md §E 경계 규칙에 현재 문구를 인용. 분기 A로 범위를 옮기면 "unchanged"와 맞춰야 하며 리드가 조정 |
| (iter-2) 초안 자체 | AGENTS.md 템플릿의 worktree-entry 행이 "`moai codex -w <worktree>` — resolves an existing tree and never creates one"이라고 적는다 | `resolveOrCreateCodexWorktreeDir`는 이름이 가리키는 트리가 없으면 만든다(`codex_launcher.go:398-429`) | 불일치 관측. 이 카드의 범위 밖(spec.md §F) |
| (iter-2) 초안 자체 | design.md 초안 C.1 "역할별 MCP 도구 제한 — 서버 등록은 세션 단위" | 역할별 서버 부여는 표현되고 실측돼 있다(위 B.3). 표현 못 하는 것은 서버 안의 도구 단위 제한과, 부여하지 않은 역할에 대한 거부(미측정)다 | design.md C.1을 서버 단위·도구 단위로 나눠 고침 |
| (iter-2) 초안 자체 | B.3 삭제 보호 "분류는 누가 만들었든 적용된다" | `done`·`remove`는 레지스트리만 보므로 lock만 가진 Codex 트리를 anchor로 보지 않는다 | REQ-DHR-010에 lock 판정 요구 추가 |
| (iter-3) iter-2 개정 자체 | REQ-DHR-010이 `moai worktree done`도 미통합 Codex 트리만 거부하고 병합된 트리는 지운다고 기대 | `done`은 `<mainRoot>/.claude/worktrees/` 아래 트리를 `--force`와 무관하게 `L1_SESSION_WORKTREE`로 거부한다(`done.go:76-81`, `:277`, SPEC-WORKTREE-DONE-TIER-001 completed). `moai codex -w <name>`은 바로 그 경로에 트리를 만든다(`codex_launcher.go:411`, `session_worktree.go:47`). 절대 경로 값은 기존 L2 트리에 들어갈 뿐이다(`codex_launcher.go:356-378`) | REQ-DHR-010을 L1 거부 유지 + 실제 삭제 경로(clean --stale, remove, PR-merge 정리)의 보호로 고침 |
| (iter-3) iter-2 개정 자체 | 세션 종료 정리를 Codex 트리 삭제 경로로 봄 | 세션 종료 정리는 `init.go:504`, `web.go:115`에서만 부른다. Codex launch는 부르지 않는다. 반대로 PR-merge 정리는 `WT-*` 브랜치 트리 전부를 대상으로 하므로(`session_worktree_prmerge.go:170-173`) Codex 트리도 지울 수 있고, lock-aware `AnchorDecision`(`:217`)과 `LockRefusesRemoval`을 이미 쓴다 | 세션 종료 정리를 REQ-DHR-010에서 빼고 PR-merge 정리를 넣음 |
| (iter-3) iter-2 개정 자체 | 같은 attempt 재부여 뒤 같은 키 재전송의 처리 미정 | 현행 메시지 층은 같은 송신 세션·같은 키에 수신 세션이나 generation이 다르면 거부한다(`store.go:624-626`). t1082 design.md 204-205행·AC-FLH-008(iter-2 시점)은 `duplicate`를 기대했다 | 리드 조정 결정(2026-09-23)으로 t1082가 계약을 좁힘(재전송은 새 키). spec.md §E에 기록 |
| (iter-3) iter-2 개정 자체 | LIVE 증거를 표준 출력 한 줄로 받음 | `go test -json`은 1024바이트를 넘는 출력 줄을 여러 이벤트로 쪼갠다(`test2json.go` `outBuffer = 1024`). 이번 실행의 scratch 모듈에서 4552바이트 증거 줄이 5개 `output` 이벤트로 쪼개지고, 옛 판정식이 `jq` exit 5로 끝남을 재현했다 | 증거를 파일로 옮기고 짧은 해시 태그 줄만 표준 출력에 찍음(acceptance.md §A) |
| (iter-3) iter-2 개정 자체 | 프로필 차이로 지우는 장치는 정적 `DeprecatedPaths`뿐 | update의 관리 경로 정리 단계가 프로필과 무관하게 `CleanMoaiManagedPaths`를 부르고(`update_template_sync.go:405-424`), 이 함수가 `.claude/` 관리 뿌리 7곳을 지운다(`deploy.go:56-85`). gpt 배포자는 `.claude/**`를 숨긴다(`internal/template/harness_fs.go:112`). 코드 판독이며 측정하지 않았다 | design.md A.6 전제 정정. REQ-DHR-007·AC-DHR-005를 `.codex/` 쪽으로 한정. Claude 쪽 삭제는 범위 밖 발견(spec.md §F) |

## §D 이번 plan 단계 측정

```text
$ go test ./internal/codexwiring ./internal/factorymsg ./internal/template/agentemit -count=1
ok  	github.com/modu-ai/moai-adk/internal/codexwiring	0.608s
ok  	github.com/modu-ai/moai-adk/internal/factorymsg	3.044s
ok  	github.com/modu-ai/moai-adk/internal/template/agentemit	0.384s

$ go test ./internal/cli -run '^TestFactoryLive(CodexCodex|CodexClaude|ClaudeCodex|ClaudeClaudeCompletionSeparation)$' -count=1 -v -timeout=600s  (필터한 줄)
--- SKIP: TestFactoryLiveCodexCodex (0.00s)
--- SKIP: TestFactoryLiveCodexClaude (0.00s)
--- SKIP: TestFactoryLiveClaudeCodex (0.00s)
--- SKIP: TestFactoryLiveClaudeClaudeCompletionSeparation (0.00s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/cli	0.826s
```

두 번째 결과는 네 테스트가 모두 SKIP인데 패키지가 `PASS`/`ok`로 끝나는 모습이다. REQ-DHR-024와 AC-DHR-019가 이 경로를 막는다.

## §E 측정하지 않은 것

- Codex `read-only` 역할에서 shell 쓰기·네트워크가 실제로 막히는지.
- 수신자 재시작 뒤 이전 generation 앞 메시지가 실제로 남는지(코드 판독만 함).
- Claude Code의 EnterWorktree가 이미 lock된 트리에서 어떻게 동작하는지.
- 송신자 재시작 뒤 같은 멱등 키 재전송이 두 번째 메시지 행을 만드는지(코드 판독만 함 → AC-DHR-020이 run 첫 단계에서 측정).
- git이 lock된 worktree를 `worktree remove --force` 한 번으로 지우지 않는다는 동작(git 문서상 동작, 이 트리에서 측정 안 함 → AC-DHR-008).
- Codex 세션 기록(rollout)에 하위 에이전트 반환문이 남는지(→ AC-DHR-023, 남지 않으면 `NOT_RUN`).
- 부여하지 않은 역할이 전역 MCP 서버를 물려받는지.
- Windows에서의 anchor lock 동작.
- `internal/cli` 전체 테스트, CI 결과. 이번 단계에서는 실행하지 않았다.

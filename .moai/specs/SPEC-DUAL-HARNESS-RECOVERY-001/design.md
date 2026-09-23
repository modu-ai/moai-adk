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
| `parts[].kind` | `whole-file` / `hook-handler` / `json-key` / `toml-table` / `toml-key` |
| `parts[].key` | handler는 이벤트 키 + 명령, json-key는 `description`, toml-table은 `mcp_servers.moai` 또는 `tui`, toml-key는 `tui.status_line` |
| `parts[].origin` | `created` / `preexisting` / `unknown` |
| `parts[].hash` | MoAI가 쓴 부분의 해시. `preexisting`·`unknown`은 비움 |
| `parts[].region` | config.toml 한정: MoAI가 넣은 바이트 영역(앞에 붙인 구분 빈 줄 포함). unwire가 이 영역을 내용으로 찾아 잘라 낸다. 원본이 줄바꿈으로 끝나지 않을 때 덧붙이기·줄 삽입이 파일 끝에 더하는 줄바꿈 한 바이트(`configtoml.go:225-227`, `:130-141`)는 영역 밖이며 unwire가 되돌리지 않는다. REQ-DHR-005의 바이트 기준점은 unwire 직전 파일이므로 이 바이트가 있어도 기준은 성립한다 |

부분 종류와 코드 근거:

| 종류 | 생기는 곳 | 근거 |
|---|---|---|
| `hook-handler` | `moai hook ` 접두사 또는 `.codex/hooks/moai/` 네임스페이스 명령. 생성기가 매번 제거 후 재추가 | `hooks.go:138-141`, `:167` |
| `json-key` `description` | 최상위 `description`이 없을 때만 MoAI가 추가 | `hooks.go:115-117` |
| `toml-table` `mcp_servers.moai` | 테이블이 없을 때 파일 끝에 덧붙임 | `configtoml.go:90-94`, `appendSection` `:221-229` |
| `toml-table` `tui` | `[tui]`가 아예 없을 때 테이블 전체를 덧붙임(분기 i) | `configtoml.go:145` |
| `toml-key` `tui.status_line` | 사용자 `[tui]`에 키만 없을 때 헤더 다음 줄에 삽입(분기 ii) | `configtoml.go` `EnsureStatusLine` 분기 (ii) |
| `whole-file` | `hooks.json` 또는 `config.toml`이 없던 상태에서 MoAI가 만든 경우 | `wire.go` `wireProject` |

origin 결정 규칙(REQ-DHR-001):

| 상황 | origin |
|---|---|
| MoAI 식별 handler(접두사·네임스페이스) | `created` — 기존 생성기가 이미 MoAI 소유로 교체한다 |
| 이번 기록 아래서 MoAI가 새로 쓴 부분 | `created` |
| 부분이 이미 있고, 프로젝트에 이전 MoAI 배선 증거(manifest 부분 기록, 사이드카, 저널)가 없음 | `preexisting` |
| 부분이 이미 있고, 이전 배선 증거는 있으나 부분 기록이 없음(구버전 설치, manifest 손상 후 초기화 `manifest.go:81-88`) | `unknown` |

`preexisting`과 `unknown`은 이후에 `created`로 바뀌지 않는다. 이 규칙의 결과를 정직하게 적으면 다음과 같다.

- handler는 구버전 설치여도 다음 배선 한 번이면 `created`로 기록되어 제거할 수 있다.
- 구버전 설치의 `[mcp_servers.moai]`·`status_line`·`description`은 `unknown`이 되고 unwire가 지우지 않는다. 다시 배선해도 바뀌지 않는다. 운영자가 손으로 지워야 한다(doctor가 안내).
- manifest가 손상되어 초기화되면 이미 `created`였던 config 부분도 `unknown`이 된다. 이 경우 제거 가능성을 잃지만, 사용자 부분을 MoAI 소유로 잘못 올리는 일은 없다.

기존 신뢰 사이드카(`.moai/state/codex-wiring.json`)는 doctor의 divergence 신호로 계속 쓰고, 위 판정의 "이전 배선 증거" 중 하나로만 쓴다. 제거 판단의 근거는 부분 기록이다.

### A.2 잠금

배선 잠금은 codexwiring 패키지가 소유한다(제안 경로 `.moai/state/codex-wiring.lock`, O_EXCL 생성, pid 기록, 죽은 pid 잠금은 정리). 기존 update 잠금(`internal/cli/update_cleanup.go:56` `acquireUpdateLock`)은 재사용하지 않는다. 이유:

- `acquireUpdateLock`은 package cli의 비공개 함수이고, cli가 codexwiring을 import하므로(`update_codex_wiring.go:14`) codexwiring에서 부를 수 없다.
- O_EXCL이라 재진입이 안 된다. `runUpdate`는 이 잠금을 잡은 채(`update.go:275-279`) 배선 갱신을 부른다(`update.go:506`). 같은 잠금을 다시 잡으면 update 중에는 항상 실패한다.

순서 규칙: 두 잠금을 함께 잡을 때는 update 잠금 → 배선 잠금 순서만 허용한다. 배선 잠금을 잡은 쪽이 update 잠금을 기다리는 경로는 만들지 않는다. 그래서 교착이 없다. 배선 잠금을 살아 있는 다른 owner가 쥐고 있으면 단계는 파일을 바꾸지 않고 lock-held 결과를 돌려준다. update 경로는 기존 경고 줄(`update_codex_wiring.go:23-27`)로 "Codex wiring not refreshed: lock held"를 출력한다.

### A.3 쓰기 순서

```text
acquire wiring lock (update 잠금 보유 여부와 무관)
  → recover incomplete journal entries (A.4)
  → render, compute pre_hash / post_hash / intended provenance
  → symlink check (Lstat: target 자체, root~target 사이 디렉터리) → 실패 시 symlink-boundary, 무변경
  → journal append {op: write|delete, path, pre_hash, post_hash, provenance, temp}      [P1]
  → write temp (op=write)                                                                  [P2]
  → re-read target, hash == pre_hash ?                                                    [P3]
        no  → remove temp, journal state=conflict, report (REQ-DHR-003)
        yes → rename temp→target (op=delete: remove target)                              [P4]
              → read back, state == post ?                                                 [P5]
                    no  → journal state=diverged, report
                    yes → manifest ← journal.provenance                                    [P6]
                          → journal state=complete
release wiring lock
```

manifest 반영이 저널 완료보다 앞이다. 그래서 둘 사이에서 멈춰도 저널이 미완료로 남고, 복구가 저널에 적힌 provenance를 다시 적용한다(멱등). 저널 추가가 임시 파일보다 앞이라, 저널에 없는 임시 파일은 이전 바이너리나 저널 이전 단계의 잔재뿐이다.

### A.4 복구 분류 (REQ-DHR-004)

중단 지점(각 지점을 AC-DHR-002가 시험한다):

| 지점 | 상태 | 복구 결과 |
|---|---|---|
| P0 | 저널 없이 `.codexwiring-*` 임시 파일만 있음(구버전 `writeAtomic` 잔재 포함) | 잠금 아래서 참조 없는 임시 파일 삭제. 대상 무변경 |
| P1 | 저널 추가 후, 임시 파일 작성 전·중 | not-applied: 임시 파일 있으면 삭제 |
| P2 | 임시 파일 작성 후, 재확인 전 | not-applied: 임시 파일 삭제 |
| P3 | 재확인 후, rename 전 | not-applied(대상 = pre) 또는 diverged(사용자가 그 사이 수정). 어느 쪽이든 임시 파일 삭제 |
| P4 | rename 후, 대조 전 | completed(대상 = post) → provenance 적용 후 완료 |
| P5 | 대조 후, manifest 반영 전 | completed → provenance 적용 후 완료 |
| P6 | manifest 반영 후, 완료 표시 전 | completed → provenance 재적용(같은 값) 후 완료 |
| D1 | unwire 삭제: 저널 추가 후, 삭제 전 | not-applied(대상 = pre) → 저널 폐기 |
| D2 | unwire 삭제: 삭제 후, manifest 반영 전 | completed(대상 없음) → 부분 기록 제거 후 완료 |

| 저널 상태 | 대상의 현재 상태 | 분류 | 행동 |
|---|---|---|---|
| staged / renamed | post(삭제면 부재) | completed | provenance 적용, 완료 처리 |
| staged | pre | not-applied | 임시 파일 삭제, 대상 무변경 |
| 무엇이든 | 둘 다 아님 | diverged | 대상 무변경, 그 항목이 참조하는 임시 파일 삭제, 보고 |
| conflict | 무엇이든 | conflict | 보고만 |

진입점: `moai tool enable codex`, `moai tool disable codex`, update 경로의 배선 갱신. `moai doctor`는 미완료 저널과 참조 없는 임시 파일을 보고하고 복구 명령을 안내할 뿐 쓰지 않는다(`doctor.go:59`의 `--fix`도 제안만 한다).

### A.5 unwire 결정표 (REQ-DHR-005, 006)

| 조건 | 행동 | 보고 사유 |
|---|---|---|
| 부분 `origin=created`, 현재 해시 = 기록 해시 | 그 부분만 제거(A.3 경로로 재작성) | — |
| 파일이 `whole-file`·`created`, 파일 해시 일치 | 파일 삭제(op=delete) | — |
| 파일이 `whole-file`·`created`인데 해시 불일치 | 부분 단위 제거로 내려감 | 제거 못 한 부분은 각 사유 |
| provenance 기록 없음 | 무변경 | `no-provenance` |
| `origin=preexisting` | 무변경 | `user-owned` |
| `origin=unknown` | 무변경 | `unknown-origin` |
| 해시 불일치 | 무변경 | `modified` |
| 대상이 symlink, 또는 root~대상 사이 디렉터리가 프로젝트 밖으로 해석되는 symlink | 무변경(Lstat 기준) | `symlink-boundary` |
| 파싱 불가 파일 | 무변경 | `unparseable` |

보존 기준(형식별):

| 파일 | 보장 | 기준점 | 시험 방법 |
|---|---|---|---|
| `config.toml` | 바이트 보존: 결과 = unwire 직전 바이트에서 제거한 부분의 기록 영역만 잘라 낸 것 | unwire 직전 파일 | 기대 바이트를 테스트가 직접 만들어 `bytes.Equal` |
| `hooks.json` | 구조 보존: MoAI 외 최상위 키 값, 사용자 entry의 matcher, 사용자 handler를 JSON 값으로 비교해 같음. 순서 보존은 배열 원소(이벤트별 entry 목록, entry별 handler 목록)에만 적용한다. 객체 키 순서는 렌더러가 정렬하므로(`hooks.go:120-122`) 비교하지 않는다. 바이트 보존은 주장하지 않음(`hooks.go:122` `MarshalIndent`, `:198-212` `marshalEntry` 재직렬화) | unwire 직전 파일을 파싱한 구조 | 파싱한 구조 비교 |

설치 쪽도 같은 symlink 검사를 거친다(A.3). `writeAtomic`의 rename(`wire.go:241`)은 경로 위의 symlink를 일반 파일로 바꾸므로, 검사 없이 쓰면 사용자 링크가 사라지거나 프로젝트 밖에 쓴다.

### A.6 프로필 전환 (REQ-DHR-007)

- 배선 파일(`hooks.json`, `config.toml`, 사이드카)은 전환만으로 지우지도 고치지도 않는다. `moai update`는 고아로 보고하고 `moai tool disable codex`를 안내한다(리드 결정 1).
- 대상 프로필이 배포하지 않는 `.codex/` 템플릿 관리 경로(`.codex/agents/moai/*.toml` 등)도 보고만 한다.
- 전제 정정(plan-audit iter-2 ND6, 코드 판독): 프로필 차이로 파일을 지우는 장치는 정적 `defs.DeprecatedPaths`(`update_cleanup.go:127-131`)만이 아니다. update의 `cleanManagedPathsStage`는 프로필과 무관하게 `deploy.CleanMoaiManagedPaths`를 부르고(`update_template_sync.go:405-424`), 이 함수는 `ManagedCleanTargets`의 `.claude/settings.json`, `.claude/{commands,agents,hooks}/moai`, `.claude/skills/moai*`, `.claude/rules/moai`, `.claude/output-styles/moai`를 지운다(`deploy.go:56-85`). 템플릿에 없는 파일은 지우기 전에 `.moai-backups/<시각>/pre-clean/` 아래로 복사한다(같은 파일 함수 주석, 카드 t111). gpt 프로필 배포자는 `.claude/**`를 숨긴다(`internal/template/harness_fs.go:112` `isHidden`, `update_template_sync.go:60-67` 프로필별 배포자 선택).
- 따라서 코드상으로는 `both → gpt`(및 `claude → gpt`) 전환의 update가 위 `.claude/` 관리 뿌리를 지운 뒤 다시 배포하지 않는다. 이는 코드 판독이며 측정하지 않았다(가설). 이 카드의 범위(codexwiring)가 아니므로 고치지 않고 `spec.md` §F에 범위 밖 발견과 후속 카드 후보로 적는다. REQ-DHR-007과 AC-DHR-005는 `.codex/` 쪽과 배선 파일만 판정하며 `.claude/` 관리 뿌리의 상태를 주장하지 않는다.

## §B Codex worktree 소유와 동시 writer (REQ-DHR-008 ~ 012)

### B.1 anchor lock

기존 `internal/session/anchor_lock.go`는 git worktree lock을 권위 있는 anchor 출처로 쓰고, 사유 문자열의 `pid <n>`으로 생존을 판정한다(`lockAnchorVerdict`, `parseLockPID`). 생존 판정은 `sessionProcessLiveness` seam을 쓰며 POSIX·Windows 구현이 있다(`anchor_pid_unix.go`, `anchor_pid_windows.go`). Claude Code는 EnterWorktree 때 이 lock을 쓰지만 Codex는 쓰지 않는다. 그래서 `moai codex -w`가 대신 쓴다.

- 사유 형식(제안): `moai codex session <tree-name> (pid <pid> start <process-start>)`.
- POSIX: direct launch는 `syscall.Exec`로 moai 프로세스가 codex로 바뀌므로 pid가 그대로 이어진다(`codex_direct_posix.go`, `//go:build !windows`). exec 전에 `os.Getpid()`로 lock을 건다.
- Windows: `codex_direct_windows.go:14-29`는 자식을 시작하고 `cmd.Wait()`로 기다린다. 부모가 자식 생존 동안 살아 있으므로 부모 pid로 lock을 건다. 부모만 강제 종료되면 자식이 살아 있어도 lock이 죽은 것으로 보이는 잔여 위험이 있다(`plan.md` §G).
- spawn launch는 tmux pane 프로세스의 pid를 쓴다(기존 `defaultCodexSpawnPaneIdentity`).
- exec 이후에는 해제할 주체가 없다. 종료 후 남은 lock은 pid가 죽은 것으로 확인되어 anchor가 아니게 되고, 다음 launch가 교체한다.

교체 경합: 두 launcher가 같은 죽은 lock을 보고 동시에 교체하면, 늦은 쪽의 `unlock`이 먼저 쪽의 새 lock을 지울 수 있다. 그래서 교체는 트리별 교체 가드(제안: `<git-dir>/worktrees/<name>/moai-anchor-replace`를 O_EXCL로 생성, 끝나면 삭제) 안에서만 한다. 가드 안에서 lock 사유를 다시 읽어 처음 본 값과 같을 때만 unlock → lock을 하고, lock 후 사유를 읽어 자기 pid인지 확인한다. 가드 생성 실패, 사유 변경, `git worktree lock` 실패는 모두 launch 거부다. `--force`나 재시도로 넘어가지 않는다.

### B.2 동시 writer 판정

`AnchorDecision(tree, lock, now)`를 launch 전에 호출한다. `Anchored=true`이고 그 보유자가 호출자 자신이 아니면 거부한다. 판정은 fail-closed다(판독 불가 사유, 생존 미확정 모두 anchored). `moai cc -w`의 기존 트리 진입에도 같은 판정을 쓰며, 이 사전 판정은 lock을 읽기만 한다.

### B.3 삭제 보호

`moai codex -w <name>`은 이름을 받으면 `<project root>/.claude/worktrees/<name>`에 트리를 만든다(`codex_launcher.go:411` `filepath.Join(projectRoot, sessionWorktreeSubdir, value)`, `session_worktree.go:47` `".claude" + sep + "worktrees"`). 절대 경로 값은 이미 있는 L2 트리에 들어갈 뿐 만들지 않는다(`codex_launcher.go:356-378` `resolveCodexWorktreeDir`). 그래서 Codex가 만든 트리는 모두 L1이다.

| 경로 | L1 Codex 트리를 지울 수 있나 | 현재 판정 | 이 SPEC의 변화 |
|---|---|---|---|
| `moai worktree done` | 아니오. `<mainRoot>/.claude/worktrees/` 아래 트리는 `--force`와 무관하게 `L1_SESSION_WORKTREE`로 거부(`done.go:76-81`, `:277`, SPEC-WORKTREE-DONE-TIER-001 completed) | tier 거부가 anchor 가드보다 앞 | 없음. 기존 계약을 뒤집지 않는다. Claude L1 트리와 결과가 같은지 시험만 |
| `moai worktree clean --stale --yes` | 예(L1 제외 규칙 없음, 저장소 뿌리와 현재 트리만 보호 `clean.go:549-560`) | lock-aware `AnchorDecision`(`clean.go:136, 352`), 미커밋 변경, base 미병합, 무시된 내용 | 없음. Codex 트리가 실제로 걸리는지 시험만. 빈틈이 나오면 고침 |
| `moai worktree remove` | 예(명시 제거) | 레지스트리 기반 `LiveAnchoredSessions`(`remove.go:51`)만. lock만 가진 Codex 트리는 moai 가드를 통과하고 git이 판정 | lock-aware `AnchorDecision`으로 anchor 판정. `--force` 없으면 anchor 출처를 적고 거부. 통합 상태 판정은 추가하지 않음 |
| PR-merge 정리(`auto_cleanup` 켜짐, `moai session register`·`list` 시) | 예(`WT-*` 브랜치 트리 전부, `session_worktree_prmerge.go:170-173`) | 미커밋 변경, unpushed, lock-aware `AnchorDecision`(`:217`), lock 존재 사전 거부 `LockRefusesRemoval` | 없음. Codex 트리 상태에 같은 판정이 나오는지 시험만 |
| 세션 종료 정리 | 해당 없음. `moai cc`·`moai web` 경로만 부른다(`init.go:504`, `web.go:115`). Codex launch는 부르지 않는다 | dirty·unpushed | 없음. 이 SPEC의 판정 대상이 아니다 |

git이 lock된 트리를 `worktree remove`(또는 `--force` 한 번)로 지우지 않는 것은 git 문서상의 동작이며 이번에 측정하지 않았다. `remove`에 moai 쪽 lock-aware 판정을 넣는 목적은 git 오류가 아니라 anchor 출처를 적은 거부를 내는 것이다. AC-DHR-008이 측정한다.

### B.4 kanban 진입

`moai cc -k`의 파서(`parseLauncherEntry`), 이름 규칙(`appendLeadName`, `resolveCompanionName`), launch facts(`exportKanbanLaunchFacts`)를 codex 경로에서 재사용한다. codex는 `-f` 토큰을 이미 같은 방식(동사 조회 전에 떼어 냄)으로 처리한다(`codex_factory.go`). lead와 companion을 모두 인정한다(리드 결정 2). Codex lead에는 Claude의 `SendMessage` 알림 경로가 없어 알림은 큐와 브로커에만 의존한다.

## §C 역할 권한 계약 (REQ-DHR-013 ~ 015)

### C.1 호스트 표현력

REQ-DHR-013의 축 목록과 같다. 각 축은 근거를 갖는다.

| 축 | Codex 역할 TOML로 강제 가능? | 매핑 | 근거 |
|---|---|---|---|
| `sandbox` | `sandbox_mode` 3값(`read-only`, `workspace-write`, `danger-full-access`) | `enforced` | 필드 수용: measured(`agents-codex.yaml:57-73`, codex-cli 0.147.0 P-01 — 허용값 목록과 잘못된 값이면 역할 파일 전체가 버려짐). 쓰기 강제: AC-DHR-012 전까지 미측정(아래 문단) |
| `write-path-scope` | 불가. 3값 중 경로 단위가 없음 | `UNSUPPORTED` | measured: 위와 같은 허용값 집합 |
| `shell` | 역할 TOML에서 shell 사용을 끄는 필드를 찾지 못함 | `UNSUPPORTED` | unmeasured |
| `mcp-server` | 역할별 `[mcp_servers.<name>]` 테이블로 서버를 부여할 수 있음. 부여하지 않은 역할이 프로젝트 `config.toml`의 전역 등록을 물려받는지는 측정되지 않음 | 제한 종류별로 하나씩(REQ-DHR-013). 부여: `enforced`. 거부: `UNSUPPORTED` | 부여 measured: `agents-codex.yaml:196-207`(0.147.0, 배열형 거부·테이블형 등록). 생성된 12개 중 7개 TOML에 테이블 있음(`grep -l '^\[mcp_servers.moai\]'`). 거부 unmeasured |
| `mcp-tool` | 불가. 한 서버 안의 도구 단위 필터 없음 | `UNSUPPORTED` | documented: `agents-codex.yaml:206-207` "Per-tool filtering inside one MCP server is unavailable — documented drop" |
| `subagent` | 역할별 하위 에이전트 허가 필드를 찾지 못함 | `UNSUPPORTED` | unmeasured |
| `web` | 역할별 웹 허가 필드 없음. 웹 접근은 전역 설정 | `UNSUPPORTED` | documented: `agents-codex.yaml` `per-agent-web-grants` 항목 |

`sandbox` 축의 `enforced`는 "Codex가 이 필드와 값을 받아들인다"는 관측에 근거한다. `read-only` 역할에서 shell 명령의 쓰기가 실제로 막히는지는 측정하지 않았다. AC-DHR-012의 LIVE 항목이 측정한다. 그 전까지 "read-only가 쓰기를 막는다"는 가설이며, REQ-DHR-013에 따라 런타임 차단 주장은 REQ-DHR-014의 LIVE 증거에만 기댄다.

`UNSUPPORTED` 보고 집합은 고정 목록이 아니라 계약에서 계산한다. 어떤 역할의 계약이 어떤 축에서 제한을 요구하고, 그 축이 위 표에서 `UNSUPPORTED`이면 (역할, 축)이 보고 집합에 들어간다(AC-DHR-011).

### C.2 12개 역할 (이 트리의 생성 결과와 계획된 변경)

| 역할 | 현재 `sandbox_mode` | 계획 | Claude 쪽 쓰기 도구 |
|---|---|---|---|
| mission-governor | read-only | 유지 | 없음(Read, Grep, Glob, Skill) |
| super-advisor | read-only | 유지 | 없음(Bash 있음). MCP `codex_task`는 프로젝트 opt-in(`workflow.codex.task.allow_write`) 시 쓰기 가능한 작업을 만든다 |
| plan-auditor | workspace-write | **read-only**(REQ-DHR-015) | Write, Edit |
| sync-auditor | workspace-write | **read-only**(REQ-DHR-015) | 없음(Bash로 판정 파일 작성) |
| manager-spec, manager-develop, manager-docs, manager-git, manager-design, manager-lead, builder-harness, e2e-tester | workspace-write | 유지 | Write, Edit. manager-lead만 Agent 보유 |

Codex의 12개 역할은 MoAI 11개 + `mission-governor`다. CLAUDE.md의 12개(MoAI 11개 + 내장 `Explore`)와 구성이 다르다.

### C.3 감사 역할의 Codex 예외 (REQ-DHR-015)

설계 §10 원문(`reports/moai-dual-harness-full-design-20260922.md:177`): "역할별 권한을 호스트가 표현하지 못하면 더 넓은 권한을 조용히 부여하지 않는다. 외부 worker의 sandbox로 강제할 수 있는지 먼저 검증하고 불가능하면 해당 역할을 차단한다."

감사 역할에 필요한 쓰기는 보고·판정 경로뿐인데 Codex sandbox는 경로 단위 제한(`write-path-scope`)을 표현하지 못한다. 그래서 `workspace-write`는 필요보다 넓다. 리드 결정 (a)(출처와 한계는 `plan.md` §B)에 따라 Codex에서 두 역할을 `read-only`로 두고, 판정·보고 파일은 부모 lane 오케스트레이터가 감사자의 반환문 그대로 기록한다. 쓰기 범위를 sandbox가 강제하므로 설계 §10의 "차단" 분기로 가지 않는다.

변경 범위:

| 대상 | 변경 | 하네스 |
|---|---|---|
| `internal/template/agentemit/agents-codex.yaml` `sandbox_mode.role_values` | `plan-auditor: read-only`, `sync-auditor: read-only` 추가 | Codex만 |
| 같은 파일 `read-vs-write-distinction` 근거(262-264행 "sync-auditor must write its verdict file…") | 새 계약에 맞게 고쳐 씀 | Codex만 |
| 생성된 `plan-auditor.toml`, `sync-auditor.toml` | `sandbox_mode = "read-only"` + Codex 전용 지시문(반환문으로 결과를 돌려주고 파일을 쓰지 않음) | Codex만, `make agents-emit`으로 생성 |
| 부모 오케스트레이터 지시 | Codex가 읽는 표면(템플릿 `AGENTS.md.tmpl`의 capability 표 또는 Codex 전용 발행물)에 "Codex 감사 역할이 반환하면 반환문 그대로 판정 파일을 기록한다" | Codex만 |
| `.claude/agents/moai/plan-auditor.md`, `sync-auditor.md`(C1·C2), Claude 감사 워크플로 | 변경 없음 | Claude 경로 불변 |

남는 위험: 부모가 반환문을 바꿔 적을 수 있다. 이를 기계로 확인하는 방법은 AC-DHR-023(LIVE, Codex 세션 기록과 판정 파일 대조)뿐이며, 세션 기록에 하위 에이전트 반환문이 남는지는 측정되지 않았다. 남지 않으면 AC-DHR-023은 `NOT_RUN`이다.

## §D dispatch record (REQ-DHR-016 ~ 021, 025)

### D.1 테이블(제안)

`dispatches` — 기존 factorymsg SQLite 브로커 안의 새 테이블.

| 열 | 뜻 |
|---|---|
| `project_key`, `run_id` | 기존 브로커 범위 |
| `dispatch_id` | run 안에서 유일 |
| `card_id` | 카드 식별 |
| `lane_slot` | 현재 assignee lane(안정 주소) |
| `attempt` | 1부터. 재할당마다 +1 |
| `assignee_generation` | 할당(또는 재부여) 시점의 `peers.generation` 복사본. fencing token |
| `state` | `assigned` / `delivered` / `started` / `result_recorded` / `integrated` / `abandoned` |
| `result_attempt`, `result_digest`, `result_ref` | 기록된 결과의 attempt, 본문 해시, 참조. 한 번만 설정 |
| `updated_at` | 기록 시각 |

메시지 멱등 키: 할당 `dispatch:<dispatch_id>:<attempt>`, 결과 `result:<dispatch_id>:<attempt>`. 재할당은 attempt가 바뀌므로 다른 lane으로 보내도 키 충돌(다른 수신자 → 거부)에 걸리지 않는다. 같은 attempt 재부여는 attempt를 바꾸지 않으므로, 재부여 뒤 같은 키로 새 generation에게 할당을 다시 보내면 현행 메시지 층이 수신 세션·generation 불일치로 거부한다(`store.go:624-626`). 리드 조정 결정(2026-09-23)에 따라 그 재전송은 새 키를 쓰며(t1082 소관), 이 SPEC은 이 경우를 위해 멱등 범위를 바꾸지 않는다(`spec.md` §E). 운영자 결정 3의 조건부 이관(D.4)과는 별개다.

### D.2 전이

```text
assigned --(assignee의 receipt)--> delivered --(assignee의 start, fenced)--> started
started --(결과 적용, D.3 표)--> result_recorded --(lead)--> integrated
assigned|delivered|started --(재할당: 이전 소유자 비생존 확인 또는 명시 철회)--> assigned(attempt+1)
assigned|delivered|started --(같은 attempt 재부여: 호출자 SPEC이 조건 판정)--> 같은 상태(assignee_generation 갱신)
어느 비종결 상태 --(lead 포기)--> abandoned
```

메시지 receipt는 `delivered`까지만 올린다. 그 이상은 명시적 호출만 올린다(REQ-DHR-016).

### D.3 결과 적용 판정 순서 (REQ-DHR-018)

| 순서 | 검사 | 실패 시 결과 |
|---|---|---|
| 1 | dispatch 레코드 존재 | `unknown` |
| 2 | 보고 attempt = 현재 attempt | `stale` |
| 3 | 보고 lane = assignee lane | `stale` |
| 4 | 보고 generation ≥ lane의 현재 `peers.generation` | `stale` |
| 5 | 이 attempt의 결과가 이미 있음 → digest 같음 `duplicate`, 다름 `collision` | (해당 시 종료) |
| 6 | 보고 generation = assignee generation | `stale` |
| 7 | state = `started` | `invalid-state` |
| 8 | 결과 기록, `result_recorded` | `accepted` |

시나리오 확인:

| 시나리오 | 결과 |
|---|---|
| 같은 결과 메시지가 두 번 전달 | 첫 번째 `accepted`, 두 번째 5에서 `duplicate` |
| 수신자가 적용 후 receipt 전에 죽고 lease 만료로 재전달 | 5에서 `duplicate` |
| worker가 결과를 보낸 뒤 재시작(generation +1)하고 같은 결과를 다시 보냄 | 4 통과(현재 generation), 5에서 `duplicate` |
| 위 상황에서 이전 generation 메시지가 뒤늦게 다시 전달 | 4에서 `stale`(t1082 AC-FLH-008과 같은 방향) |
| worker가 결과 전에 재시작하고 새 generation으로 첫 결과를 보냄 | 6에서 `stale`. lead가 재할당(REQ-DHR-021)하거나 호출자 SPEC이 재부여 |
| 재할당(다른 lane, attempt 2) 뒤 attempt 1의 늦은 결과 | 2에서 `stale` |
| 브로커 재시작 직후 재전달 | 저장된 레코드 기준으로 5에서 `duplicate` |

결과 적용이 attempt 단위로 한 번이므로, 메시지 층의 멱등 범위(D.4)가 어느 쪽이든 결과는 한 번만 반영된다. 멱등 범위가 바꾸는 것은 수신자가 같은 결과를 두 번 받아 두 번 판정하느냐뿐이다.

### D.4 멱등 범위 — 측정 후 결정 (REQ-DHR-017, 025)

현재 `messages`는 `UNIQUE(sender_session, idem_key)`(`store.go:282`)이고, 충돌 재조회도 `sender_session`으로 한다(`store.go:624`). 코드만 보면 송신자 재시작 뒤 같은 키가 새 행이 될 것으로 보이지만 측정되지 않았다.

run의 첫 마일스톤이 이를 측정한다(REQ-DHR-025, AC-DHR-020). 재현 기준은 "두 번째 메시지 행이 생기고 수신자가 서로 다른 메시지 ID 두 개를 claim한다"이다. 바뀌지 않은 트리에는 결과 적용 단계가 없으므로, 수신자에게 두 번 전달되는 것이 결과를 두 번 반영할 수 있는 유일한 경로다.

| 측정 결과 | 행동 |
|---|---|
| `reproduced` | 범위를 `(project_key, run_id, sender_slot, idem_key)`로 옮긴다. `sender_slot` 열 추가, 기존 행은 `peers`로 매핑 가능하면 그 slot, 불가능하면 `legacy:<sender_session>`. `SchemaVersion` 증가. t1082 영향은 리드가 조정 |
| `not-reproduced` | 스키마를 바꾸지 않는다. 측정 기록을 근거로 남긴다 |

### D.5 superseded (REQ-DHR-020)

`Claim`은 이미 `recipient_generation`으로 거른다(`store.go` `Claim`). 수신 lane이 새 generation으로 재등록되면 이전 generation 앞 메시지는 누구도 claim하지 못한 채 남는다(코드 판독; 측정은 AC-DHR-015). 이 SPEC은 그 메시지를 status에 `superseded`로 따로 세고, body를 새 generation으로 옮기지 않는다. 권한 이동은 새 attempt 또는 같은 attempt 재부여 연산으로만 한다. 재부여를 언제 부를지는 호출자 SPEC(t1082)이 정한다.

## §E t1082 경계 요약

`spec.md` §E가 정본이다. 요점만 적는다.

- generation 필드는 하나(`peers.generation`), 결과 적용 판정 순서도 하나(D.3).
- 멱등 범위는 측정 결과에 따라 t1100이 정한다(D.4). 옮기게 되면 리드가 t1082와 조정한다.
- 같은 attempt 재부여 뒤 재전송(ND2)은 리드 조정 결정으로 t1082가 새 키를 쓴다. 이 SPEC은 이 경우 멱등 범위를 바꾸지 않는다. 결정 3의 조건부 분기와 별개다.
- t1082의 handoff 상태기계는 D.2 위에 얹히는 gate로 제안되며, t1100은 handoff 상태를 읽지도 쓰지도 않는다.
- t1082의 M1(launch-pending 중 handoff), M2(두 rebind 경로)는 이 설계가 결정하지 않는다.

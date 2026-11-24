---
id: SPEC-20261121-go-path-removal
title: "moai drops the Go path: inventory refresh, Go-link removal in update, name-switch retirement"
status: approved
revision: 6
created: 2026-11-21
approval:
  by: leader
  at: 2026-11-24
  basis: "리더 결정 노트 8461 (2026-11-24T10:22:48Z, 사용자 위임 2026-11-23T13:57Z, 리더 판단, 사용자 직접 선택 아님; 노트 8458 의 D-12 와 앞서 노트 8409 의 D-11, 노트 8341·8346 의 D-10, 노트 8336 의 D-1~D-9): 판 6(D-10·D-11·D-12 고침) 승인, head 329204cb2d; 레인은 그 노트를 인용만 해요"
  digest: "sha256:254e2cdb26a748ff181c766bd7924f231023d09067b7524377bca7d3062f06e6"
---

# moai drops the Go path: inventory refresh, Go-link removal in update, name-switch retirement

## Intent

사용자 지시(2026-11-21, 리더 전달): "go하고도 연결 모두 끊어라", "go path 완전 제거해라", "모든 moai-adk-go 버전을 사용중단하고 moai update 하면 기존의 go와 연결 고리 제거후 업데이트 하고 moai-adk-go 관련 내용은 제거한다". 제품이 더 이상 어떤 경로로도 Go 빌드를 실행하거나 설치하지 않게 하는 카드다.

기계 상태(2026-11-21 기준): `moai` 는 `~/.local/bin/moai` → `~/.moai/bin/moai` → `~/.moai/versions/4.2647.15/bin/moai` 로 바로 Rust 빌드로 연결된다. 이름 전환 shim(`~/.moai-cli/path/moai`, v2)과 보관 사본 네 개는 PATH 에서 내려온 고아 파일이 됐고, `~/go/bin` 에는 Go moai 바이너리가 없다(사용자가 휴지통으로 치움, 다른 Go 도구는 남음).

전체 계획은 세 조각이다: (1) 읽기 전용 목록 조사 — 이 revision 이 넘기는 것. (2) `moai update` 가 모든 사용자의 Go 연결고리를 제거한다(옛 런처 스크립트와 보관 사본, 그 이름을 부르는 PATH 줄, 프로젝트·홈 설정의 Go 세대 훅 줄 — 미리보기·적용·되돌리기 기록과 함께, 사용자 파일을 조용히 지우지 않고, Go 바이너리 자체는 사용자가 지우도록 목록으로 보여 준다). (3) 제품에서 이름 전환 스크립트와 그 update 단계와 Go 라우팅 낱말 목록을 제거하고 문서·스킬의 Go 시절 글을 정리한다. go-state·migrate 도구는 update 이식이 필요로 하는 동안만 남는다. Go 가드를 내리기 전에 보안 가드 이식 카드가 먼저 착지해야 하고, SPEC 은 영향을 받는 가드를 적어야 한다. 조각 2·3 읊의 세부 AC 는 각 조각 착수 때 이 SPEC 의 다음 revision 에서 갱신한다. HOME 경로 카드(폴더 이동과 설정 줄을 소유)와 겹치는 항목은 그 카드와 조율한다.

기준 목록: SPEC-20261119-rust-only-moai 의 `inventory.md`(가지 spec/20261119-rust-only-moai, head 903061f05233a42fa8c014d3e0819b70871fe19b). 이 revision 의 목록은 그 기준을 2026-11-21 기계 상태와 현 tip 에 맞춰 갱신한다.

조각 순서 결정(2026-11-21, 카드 결정 노트 3건 — 사용자 "진행해라" 승인): 목록 조각 바로 다음에, `moai update` 가 사용자 저장소의 Go 세대 git 훅을 찾아(곁의 해시 파일과 머리말 글로 인식) 미리보기에 보여 주고, 적용 때 되돌리기 기록과 함께 은퇴시키고, 그 자리에 Rust 훅을 설치한다. MoAI 표시가 없는 사용자 작성 훅은 건드리지 않고 "kept" 으로 나열한다. 이 슬라이스가 사용자가 처음 만나는 부분이다(Go 바이너리가 사라진 뒤 실패하는 커밋). 제품 작업: 어떤 프로젝트에도 이 git 훅을 설치·미리보기·제거하는 moai 명령 하나와, 그 명령을 은퇴 직후에 쓰는 update 단계. 트리 안의 `install/commit-msg` 는 옛 런처 경로를 고정하고 있어 plain `moai` 를 부르게 고친다. 모든 훅은 100ms 안쪽이고, 시험은 Rust 로 임시 저장소에서: Go 훅 은퇴 후 교체, 사용자 훅 바이트 그대로 보존, 섞인 폴더, 두 번째 실행은 변화 없음, 되돌리기 복원. pre-push 가 Rust 검사를 돌릴지(예: 밀린 범위의 trailer-check)는 그 조각의 revision 이 이 SPEC 에서 정한다. 리더가 이 기계의 공유 훅 폴더에서 Go 세대 훅을 이미 은퇴시키고 작은 Rust 훅 둘(commit-msg, pre-commit)을 넣었다 — 목록의 기계 상태에 기록했다.

### 판 6 (2026-11-24): 조각 2·3 에 남은 것

2026-11-24 에 `bd7ee6a41f`(그 커밋 시점의 통합 브랜치 `NeXT`)에서 읽었어요. 읽기만 했고 코드 지도는 `unavailable` 이라서 `rg` 로 찾아 파일을 읽었어요. 근거 행과 읽은 명령은 Decisions 의 Evidence 표에 있어요.

- 착지한 것: 목록 조각(판 1~4)과 판 5 의 git 훅 조각 — 머지 `be62c06a39`(head `d1492e6e41`). 아래 REQ-1~REQ-12 와 AC-1~AC-15 는 그 일의 기록이고 그대로 남아요.
- 다른 SPEC 아래 트리에 이미 있는 것(각각 미리보기·확인 토큰·되돌리기가 있어요): `go-state-*`(프로젝트의 Go `.moai/`), `home-retire-*`·`home-install-*`(`~/.moai` 의 Go 항목), `home-archive-*`(옛 홈 폴더 `<home>/.moai-cli` 를 사용자 확인 뒤 이름 바꾸기 한 번으로 통째로 보관 — SPEC-20261123-home-state-retire), `go-hooks-*`(Go 세대 프로젝트 훅과 그 settings 항목), `go-settings-*`(`.claude/settings.json` 의 Go 세대 키), `worktree-move-*`, `bin-names-*`. 그래서 처음 계획의 조각 2 는 라우팅 스크립트(`<state home>/path/moai`)와 `moai mcp-server` 를 부르는 `.mcp.json` 파일(`inventory.md` 의 MCP 등록 행이 조각 2 의 미리보기에 적어 둔 것) 두 항목을 빼고 끝났어요. 그 이름을 부르는 `PATH` 줄은 제품이 일부러 손대지 않고 사람 몫으로 두어요.
- 조각 3 중 아직 트리에 있는 것: v2 라우팅 스크립트 `install/moai-switch.sh` — `src/update/switch.rs` 가 바이너리에 넣고 `moai update --confirm-switch` 가 써요(Go 전용 낱말을 `moai.old` 로 보내요). `tools/claude-statusline.sh` 의 Go 폴백. `moai.old` 가 있으면 그것을 시작하는 차이 시험 `tests/diff_go.rs`.
- 일부러 두는 것: 셸 파일은 어느 모듈도 읽거나 고치지 않아요(`home-archive-plan` 이 옛 폴더 이름을 담은 `PATH` 항목을 보여 주고 안내만 해요). `install/README.md` 는 Go 판이 설치되던 곳을 비교표로 적은 것이고 이 제품이 실행하는 경로가 아니에요.

판 6 은 세 조각이에요. (L) `moai update` 의 switch 부분이 라우팅 스크립트를 끝내요: `<state home>/path/moai` 의 v1 또는 v2 스크립트를 `moai.before-<UTC>` 로 보관하고 옆의 런처 `moai-cli` 를 가리키는 상대 링크로 바꿔요. 그 부분 자신의 토큰과 되돌리기는 그대로예요. (T) 트리가 Go 경로를 더 싣지 않아요: 바이너리에 라우팅 스크립트가 없고, 래퍼에서 Go 폴백이 빠지고, 차이 시험이 나가요. (E) Rust 시험 하나가 임시 HOME 에서 새 셸이 받는 것을 보여 줘요: Rust 버전 줄, 번들 동사, 훅 호출 한 번과 상태 줄 호출 한 번, 그리고 `go` 나 `moai.old` 는 시작되지 않아요.

실제 기계는 사용자 몫이에요(HOLD `guarded`: 실제 HOME, `PATH` 줄, `<home>/.moai-cli`). 이 판은 제품과 그 임시 HOME 시험을 써요. 사람이 기계에서 실행하는 일은 Plan 의 "User steps on the real machine" 에 적었고 여기의 요구는 아니에요. 카드 기준 2 는 AC-16~AC-22 가, 카드 기준 3 은 AC-19 와 사용자의 확인(D-9)이 채워요.

## Interview

The questions revision 6 put to the leader on the card (note 8333) and the leader's answers (note 8336, 2026-11-24T07:45:22Z, a leader judgment under the user's delegation of 2026-11-23T13:57Z and not the user's own pick):

- Q1: Does the switch part end the routing script by a link, or is the part removed? -> A: by a link, for v1 and v2, with the saved copy, its own token and the revert (D-1)
- Q2: Does the binary keep embedding the v2 script? -> A: no; the byte copies stay in the preview folder of the v2 SPEC (D-2)
- Q3: Do AC-1 and AC-5 keep their revision 5 verify commands? -> A: no; AC-1 reads the preserved shim copy and AC-5 names this revision's paths against `NeXT` (D-3)
- Q4: Is the probe of `install` that starts a `moai` a machine holds left as it is? -> A: yes; it is named as the one place a Go build can be started, with `--version` or `statusline --help` only (D-4)
- Q5: Are `.mcp.json` entries that run `moai mcp-server` previewed in this revision? -> A: no; the user steps name the search and a read-only preview is a later SPEC (D-5)
- Q6: Is revision 6 split into SPECs linked by `parent`? -> A: no; one SPEC with 22 requirements, and the checker's size warning is accepted as a decision (D-6)
- Q7: Are `tests/diff_go.rs` and its fixture kept? -> A: no; both are deleted and the approved-differences list stays readable in history (D-7)
- Q8: Does the statusline wrapper keep its Go fallback? -> A: no; it loses the fallback only, and the Python wrapper test is adjusted as text and is never run or cited as evidence (D-8)
- Q9: What closes card criterion 3? -> A: the temporary-HOME run of AC-19; the user's look at a fresh shell on the real machine is a user step, recorded on the card after the landing (D-9)
- Q10: Does the fresh-shell test run `bundle build` and `bundle verify` with exit 0 on its scrubbed `PATH`? -> A: no; it shows them answering with this build's envelope, and the verify line of AC-19 runs the two bundle test targets that show their working (D-10)
- Q11: Does AC-22 ask for exactly one start of a `moai` on `PATH` per `install-inspect` run? -> A: no; every start carries `--version` alone and no other argument, and one run asks twice today (D-11)
- Q12: Does the verify command of AC-14 run as revision 5 wrote it? -> A: no; `cargo test` takes one test name before `--`, so the two filters of AC-14 follow `--`; what the criterion asks does not change (D-12)

## Non-goals

- 이식·퇴역·shim 제거의 실행: 이 revision 은 읽기만 한다(조각 2·3 의 몫).
- Go 도구체인, `~/go/bin` 의 다른 도구(buf, actionlint 등), 다른 Go 프로젝트, 열린 세션, 작업 트리는 건드리지 않는다.
- 실제 HOME 에 쓰기: 읽기만 한다.
- Go 하위 명령 실행: `--help` 없이는 실행하지 않는다(이 기계엔 Go moai 가 없어 실행 자체가 불가능하다).

조각 본체(구현, 리더 결정 노트 20 기준): 피처 모듈은 `src/git_hooks/`(신규 폴더 — Go 세대 훅 인식·미리보기·적용·되돌리기), 훅 텍스트는 `install/git-hooks/`(pre-commit·pre-push), 시험은 `tests/git_hooks_cli.rs`(임시 저장소, `#[path]` 포함으로 배선 없이 단독 검증 — quality_gate 조각 선례), 그리고 `install/commit-msg` 는 plain `moai` 를 부르게 고친다. `src/update.rs` 와 `src/lib.rs` 는 update를 고치는 레인이, `src/main.rs` 는 명령 연결 레인이 쥐고 있어 pair request 노트를 남겼고(각 홀더 카드의 노트 5800·5801), update 레이어 연결과 `git-hooks-plan|apply|revert` 명령어 연결은 홀더와 조율해 붙인다. pre-push 는 리더가 받은 기본안으로 정한다: 밀린 범위의 trailer-check(카드 번호 스캔 포함), 100ms 안쪽, moai 가 PATH 에 없으면 안내문과 함께 통과. 12개 저장소의 막힌 pre-commit 을 레인이 손대는 일은 없다(리더가 사용자에게 보고).

연결 revision(3, 리더 지시 "같은 조각을 끝까지"): 이 조각이 이름 붙이고 편집하는 파일은 다음과 같다. (1) `src/lib.rs` — 정렬된 모듈 목록에 `pub mod git_hooks;` 한 줄(update 레인 ACK 노트 5803 승인). (2) `src/git_hooks/update.rs`(신규, 내 모듈 안의 하위 모듈) — update 흐름의 Layer(전용 토큰 --confirm-git-hooks, 미리보기 JSON 부분, 적용·되돌리기 위임). (3) `src/update.rs` — switch 레이어 옆의 새 한 줄들: UserLayerPlan 필드, plan 호출, 생성자 줄, asks 토큰 경로와 적용 순서 목록(홀더 범위 안이라 정확한 줄을 노트 5820 로 요청, ACK 5822 로 반영), 그 밖의 summary·JSON·적용 발동·되돌리기는 홀더 범위 밖에 새로 넣는다. (4) `src/ops/git_hooks_plan.rs`·`git_hooks_apply.rs`·`git_hooks_revert.rs`(신규) — build.rs 가 src/ops/*.rs 를 자동 수집해 명령이 생기므로 main.rs 편집이 필요 없어졌고(main.rs 홀더 ACK 노트 5812 는 적용되지 않음), 각 명령의 문서 조각 `docs/design-review/appendix/operations/git_hooks_<word>.md`(신규 3개)가 전수 검사의 요구다. (5) `tests/update_git_hooks_cli.rs`(신규) — 임시 저장소에서 명령과 update 레이어를 검증. help.rs 행은 update 레인 몫으로 같은 노트에 선택 요청 — 빠지면 이 revision 이 기록한다.

판 6 (2026-11-24) 가 두고 오는 것:
- 실제 HOME 과 그 `PATH` 줄과 `<home>/.moai-cli` 전체: `home-archive-plan`·`home-archive-apply` 와 rc 파일은 사용자의 단계예요.
- 프로젝트의 `.mcp.json` 파일(D-5)과 `install` 의 버전 조회(D-4).
- `src/shim_check.rs`(shim 에 대한 읽기 전용 보고 — 그런 shim 이 기계에 있는 동안 남아요), `install/install.sh`(링크를 놓고 평범한 파일은 경고만 하고 그대로 둬요), `install/README.md`.
- `tests/switch_script_v2.rs` 와 `SPEC-20261119-switch-script-v2` 의 `preview/` 폴더: 그 SPEC 의 시험은 그 SPEC 의 미리보기 파일을 돌려요(제품 파일이 아니에요). 앞선 두 SPEC(`SPEC-20261119-switch-script-v2`, `SPEC-20261119-update-switch-layer`)은 기록으로 남고 그 글은 여기서 고치지 않아요.
- Go 세대가 한 일을 적은 주석과 `docs/` 기록(역사), 래퍼 `tools/claude-statusline.sh` 의 조회 순서.

## Risks

- 기계 상태는 하루 안에도 움직였다(11-19 목록 이후 shim 고아화, Go 바이너리 퇴장, 자산 번들 착지, 공유 훅 교체). 목록의 수는 읽은 시점의 것이고, 조각 2·3 실행 전에 관련 행을 다시 세야 한다.
- 판정을 갈린 호출자 수에 의존한다. 게이트 호출 12곳처럼 실시간으로 막힌 경로는 목록이 아니라 수리(슬라이스)가 닫는다 — 목록만으로 커밋 막힘이 풀리지 않는다.
- 폴더·파일 제거는 사용자 파일을 다룬다(shim 보관 사본, Go 세대 훅). 조용한 삭제 없이 미리보기·토큰·되돌리기 기록이 없는 조각은 착지하면 안 된다.
- 알려진 한계(안전한 쪽으로 틀린다): (a) 이 기계에서 찾은 Go 세대 변형 중 다음은 알아보지 못해 사용자 것으로 둔다 — 빠른 부분집합 머리글의 pre-commit(복사 사본들), 명령 줄이 없는 pre-push(여러 저장소), 손으로 깐 Rust pre-commit(출하본과 해시가 다름, 약 14곳). 못 알아본 훅은 지워지지 않고 `kept` 로 보고된다. (b) pre-push 는 새 브랜치·새 태그를 훑지 않는다(범위가 없어서). 원격 끝이 로컬에 없는 강제 푸시도 범위를 못 만들어 안내만 하고 통과시킨다. sha256 저장소의 새 브랜치도 같은 이유로 안내와 함께 통과한다. (c) 읽을 수 있는 사용자 편집이 은퇴한 이름 위에 있으면 그 이름만 되돌리지 않고 `blocked` 로 보고하며 보관본을 폴더에 남긴다.

판 6 (2026-11-24):
- 링크는 상대 링크예요(한 폴더에서 `moai` 가 `moai-cli` 를 가려요). 두 이름이 한 빌드로 답하고, `path/` 로 부른 `moai` 는 런처 링크가 실패할 때만 실패해요. 설치 스크립트의 "자기가 만든 링크가 아니다" 경고가 둘 다 덮어요.
- Rust 표식이 없어 옛 폴더를 그대로 쓰는 HOME 에서는 링크 단계와 `home-archive-apply` 가 둘 다 `<home>/.moai-cli/path` 에 닿아요. 보관 적용은 그 폴더 아래 경로를 쓰는 프로세스가 있는 동안 거절하고, 보관 사본들은 폴더와 함께 움직여요.
- 링크를 계획하기 전에 런처를 `--version` 으로 한 번 시작해 Rust 인지 확인해요. 이 조회는 `install` 의 기존 조회예요(환경을 비우고 시간 한도가 있어요).
- `tests/diff_go.rs` 는 `moai.old` 없는 기계에서 건너뛰기만 하는 시험 여덟 개와 함께 나가요. 승인 차이 목록은 역사(`5b85682932`)에서 읽을 수 있어요(D-7).
- `tests/claude_statusline_wrapper_test.py` 는 깨지지 않게 고쳐요. Python 모듈이라 증거가 아니에요.
- `src/main.rs`·`src/help.rs`·`src/update.rs`·`tools/repo_scans.json`·`tests/repo_scans_table.rs` 는 공유 파일이고 오늘 다른 카드가 쥐고 있어요. 줄 임대로 쓰고 순서는 리더 배차예요.
- 이 SPEC 이 요구 22 개로 커져 `moai spec check` 가 쪼개라고 경고해요(D-6).
- 임시 HOME 실행은 제품 쪽만 보여요. 기계(rc 파일, 다른 도구, `PATH` 에 무엇이 있는지)는 사용자의 확인으로만 보여요(D-9).

## Requirements

- REQ-1: When the inventory is read, it shall list every word on the Go line of the name-switch shim v2 and one row per routed spec session and bundle action. · failure: none
- REQ-2: When a row is read, it shall name one disposition rust-has port or retire and carry evidence that makes the disposition checkable today. · failure: none
- REQ-3: When the inventory is read, it shall reflect the machine state of 2026-11-21 and mark what changed since the 2026-11-19 baseline. · failure: none
- REQ-4: When the inventory is read, it shall list every place the product can still execute or install a Go path (installer, update switch step, launcher, hooks, statusline, commit-msg, templates/bundle-packs, tests reading the real HOME) with its disposition. · failure: none
- REQ-5: While the inventory is being made, the work shall change no file outside this SPEC folder and write nothing in the real HOME. · failure: none
- REQ-6: When the feature module scans a repository's hooks folder, it shall recognise each hook as Go-generation by its hash file or header text and name every hook without MoAI marks as kept. · failure: none
- REQ-7: When the feature applies with its token, it shall retire the recognised Go hooks with a revert record and install the Rust hooks in their place without touching a hook the user wrote. · failure: none
- REQ-8: When the slice installs hooks, it shall provide a pre-push hook that runs trailer-check on the pushed range under 100 ms and lets the push through with a message when moai is absent. · failure: none
- REQ-9: When the update layer lands through the pair, moai update shall run this feature's retire and install step with its own confirm token and a revert record. · failure: none
- REQ-10: When the git-hook verbs are built, the standalone commands git-hooks-plan git-hooks-apply and git-hooks-revert shall preview apply and revert the feature for any project, each with its appendix fragment. · failure: none
- REQ-11: When the update screen shows the git-hook part, it shall name the part and its --confirm-git-hooks command and, once the step ran, the retire folder with the revert command or, when the step did not finish, that it did not finish. · failure: none
- REQ-12: When the revert meets a retired name that holds anything but the bytes this feature installed, it shall leave that entry untouched and report it, whether or not the entry can be read. · failure: none
- REQ-13: When the switch preview finds the v1 or the v2 routing script at `<state home>/path/moai` by its sha256 and the launcher `path/moai-cli` beside it answers `--version` with the Rust signature, the preview shall offer the action `link` with its own confirm token and name the saved copy `moai.before-<UTC>` · failure: none — a preview writes nothing
- REQ-14: When the apply is given the preview's token, the apply shall keep the old script as `moai.before-<UTC>` with its mode and put a relative link `moai` to `moai-cli` in its place by one rename · failure: partial — the old script is saved first and a failed rename leaves it in place
- REQ-15 (REQ-14): IF the token differs from the preview's or the launcher beside the script is absent or does not answer with the Rust signature, THEN the apply shall write nothing and name what is missing
- REQ-16: When the revert is given the stamp of a saved copy that is byte-equal to the v1 or the v2 script and `moai` is still the link this step made, the revert shall put the copy back by one rename
- REQ-17 (REQ-16): IF the saved copy is not a known script or `moai` is not the link this step made, THEN the revert shall change nothing and say why
- REQ-18: When the link is started with a word only the Go build had and `go/bin/moai.old` exists in the HOME, the call shall reach the Rust build without starting `moai.old` · failure: none — a link has no path to the Go build
- REQ-19: The binary and the tree shall carry no routing script and no word list of the Go build and no test or wrapper that starts `moai.old`
- REQ-20: The help of `--confirm-switch` and `--revert-switch` and the update screens shall describe the link and the saved copy and name no v1-to-v2 replacement
- REQ-21: The probe of `install` shall start a `moai` candidate only with `--version` or with `statusline --help` and a cleared environment
- REQ-22: A test in a temporary HOME that reaches `moai` only through the links an install lays and has a planted `go` and `moai.old` first on `PATH` shall show the Rust version line and answer for the bundle verbs and exit 0 for one hook call and one statusline call and no start of the planted programs

## Acceptance

- AC-1 (REQ-1): every word on the shim's Go line has a row in `inventory.md`.
  verify: `for w in $(awk '/^  af\|agent/{sub(/\)$/,"");sub(/^  /,"");n=split($0,a,"|");for(i=1;i<=n;i++)print a[i]}' .moai/specs/SPEC-20261119-switch-script-v2/preview/moai.v2); do grep -q "^| $w " .moai/specs/SPEC-20261121-go-path-removal/inventory.md || { echo "missing $w"; exit 1; }; done` (the v2 shim is read from the preserved copy of the v2 SPEC: the product file `install/moai-switch.sh` leaves the tree in this revision, D-3)
- AC-2 (REQ-2): every row of the word and action tables names one of the three dispositions.
  verify: `awk '/^## Words/{f=1;next} /^## Execution paths/{f=0} f&&/^\| [a-z]/&&!/^\| (word|action|---)/{n++; if ($0 ~ /\| (rust-has|port|retire) /) d++} END{print "rows="n" with-disposition="d; exit (n==d && n>0 ? 0 : 1)}' .moai/specs/SPEC-20261121-go-path-removal/inventory.md`
- AC-3 (REQ-3): the inventory has a machine-state section naming the resolution chain, the off-PATH shim with its saved copies, and the emptied Go bin folder of the HOME.
  verify: `grep -c -e 'versions/4.2647.15' -e 'moai.before-' -e 'no Go moai binary' .moai/specs/SPEC-20261121-go-path-removal/inventory.md`
- AC-4 (REQ-4): the execution-path table has rows for installer, update, launcher, hooks, statusline, commit-msg, bundle-packs and the tests.
  verify: `for p in installer update launcher hooks statusline commit-msg bundle-packs tests; do awk '/^## Execution paths/{f=1;next} f&&/^\| '"$p"'/{found=1} END{exit !found}' .moai/specs/SPEC-20261121-go-path-removal/inventory.md || { echo "missing $p"; exit 1; }; done`
- AC-5 (REQ-5): the change touches only the paths this SPEC names.
  verify: `git diff --name-only $(git merge-base HEAD NeXT) HEAD | grep -v -e '^\.moai/specs/SPEC-20261121-go-path-removal/' -e '^src/update/switch\.rs$' -e '^src/update\.rs$' -e '^src/update_tux\.rs$' -e '^src/main\.rs$' -e '^src/help\.rs$' -e '^install/moai-switch\.sh$' -e '^tools/claude-statusline\.sh$' -e '^docs/2383-statusline-switch-design\.md$' -e '^tests/claude_statusline_wrapper_test\.py$' -e '^tests/update_switch_cli\.rs$' -e '^tests/update_switch_layer\.rs$' -e '^tests/go_path_retired_cli\.rs$' -e '^tests/diff_go\.rs$' -e '^tests/fixtures/diff_go/' -e '^tests/fixtures/state_folder_literals/allow\.tsv$' -e '^tests/repo_scans_table\.rs$' -e '^tools/repo_scans\.json$' -e '^tests/areas\.toml$' | wc -l` prints 0 (revision 6: the base is the integration branch and the list names this revision's paths; the path list of the landed slice is in the Result).
- AC-6 (REQ-6): the recognition cases pass: Go-generation by hash file, by header text, Rust hook, and a user hook kept.
  verify: `cargo test --offline --locked --test git_hooks_cli recognition`
- AC-7 (REQ-7): the lifecycle cases pass: retire-and-replace, a user hook kept byte for byte, a mixed folder, a second run that changes nothing, and a revert that restores.
  verify: `cargo test --offline --locked --test git_hooks_cli lifecycle`
- AC-8 (REQ-8): the hook-text cases pass: pre-push checks the pushed range, passes a new branch, a deleted ref, a forced push over a tip the repository does not hold and a sha256 repository's new branch with a note that names the pushed ref, and passes when moai is absent; the managed set is pre-commit and pre-push only, so the product ships no commit-msg hook.
  verify: `cargo test --offline --locked --test git_hooks_cli hook_text`
- AC-9 (REQ-10): the verb cases pass: git-hooks-plan previews a fixture repository, git-hooks-apply applies only with the preview's token, and git-hooks-revert restores, all in temporary repositories.
  verify: `cargo test --offline --locked --test update_git_hooks_cli verbs`
- AC-10 (REQ-9): the update-layer cases pass: the update preview carries the git_hooks part with its token, a confirmed run retires and installs, and the revert restores.
  verify: `cargo test --offline --locked --test update_git_hooks_cli layer`
- AC-11 (REQ-9): a run with only the git-hook token (no digest) applies the hook step, and a run with both applies it too.
  verify: `cargo test --offline --locked --test update_git_hooks_cli update_applies`
- AC-12 (REQ-11): the update screens name the git-hook part with its confirm command, after the step the retire folder and the revert command, and a step that did not finish as not finished.
  verify: `cargo test --offline --locked --test update_tux_git_hooks`
- AC-13 (REQ-12): the revert keeps an entry it cannot read and a readable user text at a retired name, loses no byte, and restores the saved modes.
  verify: `cargo test --offline --locked --test git_hooks_cli lifecycle_revert`
- AC-14 (REQ-6, REQ-7): the header window is ten lines (line 10 is inside, line 11 is outside), a stale hash file for an absent hook is a leftover and not a kept hook, saved copies keep their modes, a name that appears during the apply is kept and named with the revert command, and a stranded commit-msg is named in the preview.
  verify: `cargo test --offline --locked --test git_hooks_cli -- recognition lifecycle` (the two filters follow `--`, the form `cargo test` takes; revision 5 named them without `--` and cargo refused the second, D-12)
- AC-15 (REQ-5): no shipped text or source comment still describes the retired commit-msg install or the replaced-hook flow.
  verify: `cargo test --offline --locked --test git_hooks_cli the_spec_and_shipped_texts_carry_no_stale_claims`
- AC-16 (REQ-13, REQ-14, REQ-15): in a temporary HOME with v1 as `path/moai` and a `path/moai-cli` whose `--version` line is the Rust signature the preview offers `link` with a token and writes nothing; the confirmed run leaves `moai.before-<stamp>` byte-equal to v1 with its mode and `moai` a relative link to `moai-cli`; a wrong token, an absent launcher and a launcher whose line is not the Rust signature write nothing and name the reason
  verify: `cargo nextest run --locked --offline --test update_switch_layer --test update_switch_cli`
- AC-17 (REQ-13): v2 as `path/moai` gets the same offer and the same link; a file that is neither script is left with its reason; a link and an absent file give no part; `go/bin/moai` is only reported
  verify: `cargo nextest run --locked --offline --test update_switch_layer --test update_switch_cli`
- AC-18 (REQ-16, REQ-17): the revert restores the saved copy byte for byte with its mode; a copy that is not a known script, a `moai` that is not the link this step made and an unknown stamp change nothing and say why
  verify: `cargo nextest run --locked --offline --test update_switch_cli`
- AC-19 (REQ-18, REQ-22): with a `PATH` of the folder holding the planted `go` and `moai.old`, then `path/` holding two links to this build, then `/usr/bin:/bin`, and a `go/bin/moai.old` in the temporary HOME that also records a start, `moai --version` prints the Rust line; `moai bundle add`, `moai bundle remove` and `moai bundle list`, one `moai hook` call with host input on stdin and one `moai statusline` call with a statusline event on stdin exit 0; `moai bundle build` and `moai bundle verify` answer with this build's envelope; after all of them the planted programs recorded no start
  verify: `cargo nextest run --locked --offline --test go_path_retired_cli --test bundle_build_cli --test bundle_verify_cli`
- AC-20 (REQ-19): `install/moai-switch.sh`, `tests/diff_go.rs` and `tests/fixtures/diff_go/` are absent, no line of `src/` embeds a routing script and `tools/claude-statusline.sh` has no `moai.old`
  verify: `test ! -e install/moai-switch.sh && test ! -e tests/diff_go.rs && test ! -e tests/fixtures/diff_go && test -z "$(grep -rnE 'include_bytes!\([^)]*moai-switch' src)" && test -z "$(grep -n 'moai\.old' tools/claude-statusline.sh)"`
- AC-21 (REQ-20): the two help entries and the update screens name the link and the saved copy and say no v2; the help tests pass
  verify: `cargo nextest run --locked --offline --test help_text_names_moai --test update_switch_cli --test cli_text_citations`
- AC-22 (REQ-21): a fake `moai` first on `PATH` that records every start with its arguments is started by `install-inspect` with `--version` alone and with no other argument, and the test that lets only `--version` reach a Go `moai` first on `PATH` stays green
  verify: `cargo nextest run --locked --offline --test go_path_retired_cli --test cli_name_moai`

## Scope

- write: `.moai/specs/SPEC-20261121-go-path-removal/**`
- write: `src/update/switch.rs`, `src/update.rs`, `src/update_tux.rs`, `src/main.rs`, `src/help.rs`, `install/moai-switch.sh`, `tools/claude-statusline.sh`, `docs/2383-statusline-switch-design.md`, `tests/claude_statusline_wrapper_test.py`, `tests/update_switch_cli.rs`, `tests/update_switch_layer.rs`, `tests/go_path_retired_cli.rs`, `tests/diff_go.rs`, `tests/fixtures/diff_go/approved-differences.json`, `tests/fixtures/state_folder_literals/allow.tsv`, `tests/repo_scans_table.rs`, `tools/repo_scans.json`, `tests/areas.toml`
Notes on the two `write` lines above (prose, not a list item): the second line is revision 6 after its approval. Until the leader's approval note the card's reservation is the SPEC folder alone (its take of 2026-11-24); the leader widens it when this revision is approved. Limits: `src/update.rs` and `src/update_tux.rs` the switch lines only; `src/main.rs` and `src/help.rs` the two flag texts only; `tools/repo_scans.json` and `tests/repo_scans_table.rs` the rows this change needs; `tests/areas.toml` the row the new-test-file precheck prints; `docs/2383-statusline-switch-design.md` the script block only. `src/main.rs`, `src/help.rs`, `tools/repo_scans.json` and `tests/repo_scans_table.rs` are shared files; the lane takes their lines by `card line-lease`. Revision 5 changed `src/git_hooks/**`, `tests/git_hooks_cli.rs`, `install/commit-msg`, `install/git-hooks/**`, `src/ops/git_hooks_plan.rs`, `src/ops/git_hooks_apply.rs`, `src/ops/git_hooks_revert.rs`, `docs/design-review/appendix/operations/git_hooks_plan.md`, `docs/design-review/appendix/operations/git_hooks_apply.md`, `docs/design-review/appendix/operations/git_hooks_revert.md`, `tests/update_git_hooks_cli.rs`, `tests/update_tux_git_hooks.rs`, `tests/fixtures/help/**` and `src/lib.rs`, with the pair-approved lines of `src/update.rs` and `src/main.rs` (the holders' ACK notes on their cards); revision 6 leaves them alone.
- read-only: the rest of `src/`, `tests/`, `install/`, `docs/`, `skills/`, `tools/`, `templates/`, `.moai/specs/SPEC-20261119-switch-script-v2/**`, `tests/switch_script_v2.rs`, the installed Rust binary (its help text and its probes in a temporary HOME), and this machine's state
- forbidden: the 12 repositories' hook folders on this machine, any write in the real HOME, any run of a Go subcommand, the `.mcp.json` files of any project, the sibling Go repository, a card number in a SPEC id, branch or commit

## Plan

1. Read the baseline inventory and the current tip sources: `install/moai-switch.sh`, `src/update/switch.rs`, `src/update.rs`, `src/update_tux.rs`, `src/shim_check.rs`, `tools/claude-statusline.sh`, `src/go_hooks.rs`, `src/user_assets.rs`, `src/ops/bundle.rs`, `install/commit-msg`, `tests/diff_go.rs`, the installer and launcher sources.
2. Probe the installed Rust binary with every routed word in a temporary HOME and record what it answers today.
3. Re-verify the machine-side counts the baseline recorded (project hook folders, git pre-commit callers, MCP registrations, shell rc lines) and record what changed.
4. Write `inventory.md` with the machine state, the word table, the action table, the execution-path table and the findings, and keep every claim on a command run for this file.

### Revision 6

1. The leader widens the card's reservation to the second `write` line of Scope. `tests/go_path_retired_cli.rs` goes through the new-test-file precheck first, and the rows it prints go into `tests/areas.toml` and the scan tables.
2. RED: the link cases in `tests/update_switch_layer.rs` and `tests/update_switch_cli.rs` (v1 and v2, the token, the launcher shown to be Rust, the saved copy with its mode, the relative link, the three refusals, the revert and its three refusals); `tests/go_path_retired_cli.rs` (AC-19 with the planted programs, AC-22).
3. GREEN, slice L (REQ-13 to REQ-18): `src/update/switch.rs` offers `link` and writes the link, with no embedded script; the switch lines of `src/update.rs` and `src/update_tux.rs`; the two help entries, rewritten; the card citation they carry today goes with the old text.
4. GREEN, slice T (REQ-19, REQ-20): `install/moai-switch.sh` is deleted; `tools/claude-statusline.sh` loses its Go fallback and the script block of the design page and the Python test follow it; `tests/diff_go.rs` and its fixture are deleted; `allow.tsv`, `tests/repo_scans_table.rs` and `tools/repo_scans.json` drop or move their rows.
5. Slice E (REQ-21, REQ-22): `tests/go_path_retired_cli.rs` is green on the head.
6. The dangerous path runs once in a throwaway HOME with the built binary: v1, v2, an edited script, a link already there, a launcher that is not the Rust build, the revert, and a stop after the saved copy (`MOAI_TEST_FAULT=switch-after-save`) and before the rename (`switch-rename`). The table goes into the READY note.
7. One `moai check affected` at READY, unwrapped, with the `guard-mutant:` lines. SYNC writes the Result lines of AC-16 to AC-22 and runs AC-1 and AC-5 again at the head.

### User steps on the real machine (not part of this SPEC's requirements; HOLD `guarded`)

The product changes of revision 6 need none of these. Each is the user's own step and runs when the user picks it; the leader records its outcome on the card.

1. `moai home-archive-plan` writes nothing: it shows the earlier folder and the `PATH` entries that name it. After the user confirms, `moai home-archive-apply` with that preview's digest and a permission reference moves `<home>/.moai-cli` whole to `<home>/.moai/archive/moai-cli-<UTC>/`; the old shim and its saved copies go with it, nothing is deleted, and `home-archive-revert` moves it back. The help of the command says who runs it on a real HOME.
2. `moai update` is a preview. When it reports a routing script at `<state home>/path/moai`, `moai update --confirm-switch <token>` replaces it by the link (revision 6) and `moai update --revert-switch <stamp>` undoes that.
3. The `PATH` lines that name `<home>/.moai-cli/path` are removed by hand from the rc files; the `~/go/bin` lines stay for the other Go tools.
4. `grep -l 'moai mcp-server' */.mcp.json` over the projects; an entry it finds is removed by hand (D-5).
5. A fresh shell: `which -a moai` names only links to the Rust build; `moai --version` prints `moai <version> (rust)`; `moai bundle build`, `verify`, `add`, `remove` and `list` answer in a scratch project; a hook and the statusline answer; nothing named `moai.old` is left in `~/go/bin`.

## Decisions

The decisions D-1 to D-9 were the lane's proposals for revision 6 and count from the leader's note of 2026-11-24T07:45:22Z. D-10 is the leader's decision on the amendment of REQ-22 and AC-19 (its note on the card, 2026-11-24T07:58:02Z), approved with that text by its digest. D-11 is the lane's amendment of AC-22, found while its test was written, approved by the leader with that text by its digest. D-12 is the lane's correction of the verify command of AC-14, found when every verify command was run as written, approved by the leader with that text by its digest.

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | The switch part ends the routing script by a link: v1 and v2 are both kept as `moai.before-<UTC>` and replaced by a relative link `moai` to `moai-cli`, with the part's own token and the revert | criterion 2 of the card asks for the shim to become a Rust link with its own token and a working revert; the installer leaves a plain file at `path/moai` with a warning (`install/install.sh:539-541`), so nothing else ends it | the part reports and replaces nothing (drop REQ-13 to REQ-18 and AC-16 to AC-18) |
| D-2 | The binary stops embedding the v2 script; its byte copies stay in the `preview/` folder of the v2 SPEC | v2 starts `moai.old` for the Go-only words (`install/moai-switch.sh:31-32,41-42`) and names the earlier folder as its Rust default (`:15`) | put the file and the `include_bytes!` back |
| D-3 | REQ-1 and AC-1 read the Go line from the preserved copy `preview/moai.v2`; AC-5 names the paths of this revision against `NeXT` | `install/moai-switch.sh` leaves the tree (D-2); the copy has the same sha256; the earlier base `v2026.11.0` no longer names this branch's base | one line each |
| D-4 | The probe of `install` stays as it is and is named as the one place a Go build can be started: a `moai` that a machine holds is started with `--version` or `statusline --help` in a cleared environment and no subcommand runs | it is how Rust is told from Go for `install-inspect`, the hook set and the status-line gate (`src/install.rs:509-524`; callers under Evidence) | stop probing `$GOBIN` and `go/bin` and report them as `not_probed` |
| D-5 | `.mcp.json` entries that run `moai mcp-server` are not previewed or changed here; the user steps name the search | a project settings file is guarded and lies outside the state folder; nothing lists them at this head (no reader or writer in `src/`); a preview would be a new command | a later SPEC adds a read-only preview |
| D-6 | The remaining work stays in this SPEC as revision 6 with 22 requirements; the `moai spec check` warning to split is accepted | the dispatch names revision 6 of this SPEC and the user wants fewer SPECs | a child SPEC with `parent:` |
| D-7 | `tests/diff_go.rs` and its fixture are deleted | the Go oracle is gone: the eight tests skip on every machine without `moai.old` and would start it where it exists (`tests/diff_go.rs:1-8,23-53`); the approved-differences list stays readable at `5b85682932` | keep both files |
| D-8 | The statusline wrapper keeps its lookup and loses the Go fallback only | the wrapper is the source of the script block of `docs/2383-statusline-switch-design.md`; the fallback (`tools/claude-statusline.sh:25-29`) is the last text in the tree that starts `moai.old` | leave the fallback (the binary does not ship the wrapper) |
| D-9 | Card criterion 3 is met by the temporary-HOME run of AC-19; the look at a fresh shell on the real machine is the user's step and is recorded on the card when the user runs it | the dispatch of 2026-11-24: the real HOME is the user's (HOLD `guarded`); the product side is provable in a temporary HOME | criterion 3 stays open until the user's look |
| D-10 | Through the links the fresh-shell test runs the verbs that need no project (`bundle add`, `bundle remove`, `bundle list`), one hook call and one statusline call to exit 0; `bundle build` and `bundle verify` are shown answering with this build's envelope, and their working is shown by the two bundle test targets that the verify line of AC-19 runs | `bundle verify` runs the check list through `cargo`, `python3` and `moai-cli` on `PATH` (`tests/bundle_verify_cli.rs:1-6,155-183`), so the test's `PATH` would start a real build; `bundle build` needs the ledger world of `tests/bundle_build_cli.rs`; copying both harnesses into the new test file duplicates about 250 lines | a copy of both harnesses in `tests/go_path_retired_cli.rs` |
| D-11 | AC-22 asks that every start of a `moai` on `PATH` carries `--version` alone, not that there is exactly one start | one `install-inspect` run starts the first `moai` on `PATH` twice today, each with `--version` only (`src/install.rs:771` for its candidates and `:914` for its PATH resolution); making it one start is a change of `src/install.rs`, which is outside the Scope, and REQ-21 asks only for the argument | a later SPEC makes `install-inspect` ask once |
| D-12 | The verify command of AC-14 passes its two test filters after `--` | `cargo test` takes one test name before `--` and refused the second (`error: unexpected argument 'lifecycle' found`, exit 1, run on this head); revision 5 wrote both without `--`, so the command never ran as written and the AC-14 line of that revision cited the `check affected` run instead | none needed |

### Evidence

Read on 2026-11-24 at `bd7ee6a41f`, read-only. A row of this table is a reading, not a run of the product.

| Item | Source read | Finding |
| --- | --- | --- |
| The switch part | `src/update/switch.rs:24-48,86-113,155-196,304-446`; `src/update.rs:1711-1713,1807,2192-2194,2240,2275,2398,2499-2539` | v1 gives `replace` and v2 is written; v2, a link and an absent file give no part; an unknown file is `left`; the revert puts a copy byte-equal to v1 back over a file that is still v2 |
| The v2 script | `install/moai-switch.sh:15-16,31-32,37,41-42`; `src/update/switch.rs:29`; `shasum -a 256` of the script and of `.moai/specs/SPEC-20261119-switch-script-v2/preview/moai.v2` | the Go-only words go to `$HOME/go/bin/moai.old`; the script is embedded by `include_bytes!`; the preserved copy has the same sha256 `36dd242b…26e4` |
| The installer | `install/install.sh:522-550,552-567` | `path/moai` and `path/moai-cli` are links into `versions/<v>/bin`; a plain file at `path/moai` is left with a warning; no rc file is edited; a Go `moai` first on `PATH` is not changed |
| The launcher shown to be Rust | `src/go_settings_clean.rs:380-410`; `src/install.rs:636-660` | a launcher counts as Rust when its `--version` first line is `moai` or `moai-cli`, a version and ` (rust)` |
| The version line | `src/lib.rs:215-231`; `src/main.rs:39-46`; `tests/cli_name_moai.rs:173-217` | both binary names print `moai <version> (rust)` |
| Process starts in `src/` | `rg -n -o 'Command::new\([^)]*\)' src` (137 single-line matches in 77 files) and the definitions of the variable ones read | literal names: `git` (72), `mkfifo`, `kill`, `true`, `sleep`, `ps`, `df`, `cp`, `stat`, `sh`, `lsof`, `head`, `gh`, `echo`, `claude`, `cargo` and test-only probes; the clap builders of help tests (`moai`, `ops`, `language-setup`) start no process; variables: `rustc` (`src/verify.rs:389-391`), `orca` (`src/lane_send.rs:228`), `git` through `MOAI_GIT_BIN`, this executable and a command the caller names; no line names `go` or `moai.old`; a `moai` that a machine holds is started only by the probe of `install` (`src/install.rs:540-613`; callers `src/install.rs:771,882`, `src/bin_names.rs:195`, `src/go_settings_clean.rs:405`, `src/status_line.rs:136`) |
| The whole earlier folder | `src/home_state_retire.rs:20,163`; `SPEC-20261123-home-state-retire` REQ-2, REQ-6 and its Result; `moai home-archive-apply --help` of the installed build | archived by one rename after the user's confirmation; the `PATH` entries that name it are listed with guidance; no shell file is read or edited |
| The other Go-era families | `--help` of the installed build `moai 4.2648.23` (local build 23, built 2026-11-23, not this head) for `go-state-*`, `home-state-*` and `worktree-move-*`; `src/go_hooks.rs:1-22`; `src/go_settings_clean.rs:1-23` | each family is present with a preview, a digest and a permission reference; the Go `.moai/` goes to `.moai-backup` and the Go `~/.moai` to `~/.moai-backup` |
| `.mcp.json` | `rg -n -i 'mcp-server\|\.mcp\.json' src` | read only: `src/shim_check.rs` (a report on a shim) and `src/project_detect.rs:634-639`; no writer |
| Shell files | `rg -n -i 'zshrc\|zshenv\|bash_profile\|rc file\|shell file' src` | no module reads or edits one (`src/bin_names.rs:10`, `src/home_state_retire.rs:20`, `src/ops/bin_names_plan.rs:12`) |
| Tests that pin the switch part | `rg -c 'confirm-switch\|revert-switch\|confirm_switch\|revert_switch\|moai-switch\|moai\.before' tests -g '*.rs'` | `tests/update_switch_cli.rs` 27 lines, `tests/update_switch_layer.rs` 12, `tests/switch_script_v2.rs` 2, `tests/help_text_names_moai.rs` 2, `tests/repo_scans_table.rs` 1 (`:1082`) |
| Scans and pins on this file | `tools/repo_scans.json:1562-1626`; `tests/git_hooks_cli.rs:661-688` | the scan `reads-moai-state` runs the switch tests and `git_hooks_cli` when this file or `install/moai-switch.sh` changes; `the_spec_and_shipped_texts_carry_no_stale_claims` reads this file for `-e '^src/update_tux\.rs$'` |
| Differential tests | `tests/diff_go.rs:1-8,23-53`; `tests/fixtures/state_folder_literals/allow.tsv:77` | eight tests; the Go build is looked up at `MOAI_GO_BIN`, else `$HOME/go/bin/moai.old`; absent, they print why and pass without comparing |
| The statusline wrapper | `tools/claude-statusline.sh:17-29`; `docs/2383-statusline-switch-design.md:155`; `rg -n 'claude-statusline' src build.rs install skills templates .github` | the fallback at `:25-29` starts `moai.old`; nothing embeds or installs the wrapper at this head; its one reader is the Python test |
| Bundle verbs and hooks | `src/ops/bundle.rs:24-37`; `src/main.rs:2606-2625,2630-2646`; `tests/bundle_build_cli.rs`, `tests/bundle_verify_cli.rs`, `tests/bundle_user_assets_cli.rs`; `tests/cli_name_moai.rs:269-413` | the five verbs have tests that run the binary by its path; an unknown hook event exits 0 with empty stdout; one test lets only `--version` reach a Go `moai` first on `PATH`; no test looks at the process table |
| The install README | `install/README.md:36-50` | a comparison whose Go column says where the Go edition installed; not a path the Rust installers run |

## Result

Filled in by the implementation lane after the re-audit repair (second repair round); the audit checks it again. Branch `spec/20261121-go-path-removal`, repair commit `9dc2d50ac676292151a93ef7a6ed5634021dfe8d` on the earlier work-in-progress commit, checked against base `65b0ed5efbf07326a1323420c7d48c8cb7a57a3d`.

- AC-1: met — the verify loop over the shim's Go line printed no `missing` word.
- AC-2: met — the verify awk printed `rows=77 with-disposition=77`, exit 0.
- AC-3: met — the verify grep counted 7 matching lines of `inventory.md`.
- AC-4: met — the verify loop printed no `missing` path.
- AC-5: met — the verify pipeline printed `0` (the path list now names the update screen file, its test file and the pair-approved files).
- AC-6: met — `recognition_*` tests in `tests/git_hooks_cli.rs` pass inside `check affected`, among them `recognition_the_header_on_line_10_is_inside_the_window` and `recognition_a_go_header_on_line_11_marks_nothing`.
- AC-7: met — `lifecycle_*` tests in `tests/git_hooks_cli.rs` pass inside `check affected`.
- AC-8: met — `hook_text_*` tests in `tests/git_hooks_cli.rs` pass; the new ones are `hook_text_pre_push_names_the_pushed_ref_when_it_passes_without_a_range`, `hook_text_pre_push_passes_a_forced_push_over_a_tip_the_repository_does_not_hold` and `hook_text_pre_push_passes_a_sha256_new_branch_with_a_note`.
- AC-9: met — `verbs_plan_previews_and_apply_needs_the_token` and `verbs_revert_restores_the_retired_hooks` pass.
- AC-10: met — the `layer_*` tests in `tests/update_git_hooks_cli.rs` pass.
- AC-11: met — `update_applies_the_hook_step_with_the_token_alone` and `update_applies_the_hook_step_with_both_confirm_flags` pass.
- AC-12: met — the six tests of `tests/update_tux_git_hooks.rs` pass; mutants of the applied screen (the step dropped, a failed step not making the run partial) are killed by them.
- AC-13: met — `lifecycle_revert_keeps_an_unreadable_entry_at_a_retired_name`, `lifecycle_revert_keeps_a_symlink_at_a_retired_name`, `lifecycle_revert_keeps_a_readable_user_edit_at_a_retired_name_and_reports_it` and `lifecycle_saved_copies_and_the_restore_keep_their_modes` pass.
- AC-14: met — the two window tests above, `lifecycle_a_stale_hash_file_for_an_absent_hook_is_a_leftover_not_a_kept_hook`, `lifecycle_a_name_that_fills_during_the_apply_is_kept_and_the_error_names_the_undo`, `lifecycle_revert_restores_a_stranded_commit_msg_only_to_an_absent_name` and `lifecycle_the_preview_names_a_stranded_product_commit_msg` pass.
- AC-15: met — `the_spec_and_shipped_texts_carry_no_stale_claims` passes.

### Checks

`check affected` (the branch's own build, `--max 2`, `CLAUDE_CODE_SESSION_ID` unset): `check-affected base=65b0ed5efbf0 head=9dc2d50ac676 targets=63 py=8 scans=listed:43 diff_check=0 fmt=0 ops_rustfmt=0 clippy=0 test_lib=0 test_bins=0 test_targets=0 build_cli=0 python=0 result=PASS`. Clippy over all targets with warnings denied and rustfmt were also run on their own and were clean.

Mutants on the committed files, each applied, run against its whole target and restored (`git status` clean after): header window widened (killed by the line-11 test), revert overwriting a readable user text (killed by the readable-edit test), a token-only update run not applying (killed by the token-alone test), the appeared-after-preview refusal removed (killed by the fill-during-apply test), saved copies always 0755 (killed by the modes test), an unreadable entry not stopping the revert (killed by the unreadable and symlink tests), the applied screen dropping the step and a failed step not making the run partial (killed by the screen tests).

### Not run

- The hooks in a real repository of this machine or the real HOME: not run (forbidden by the Scope).
- A second pass of the same targets on the same tree: not repeated.
- A merge with the tip as it is now: not made (the leader merges once at landing); the `check affected` PASS is cited for it.
- The known limits listed under Risks (Go variants not recognised, no scan of new refs) have no test beyond the pass-through cases above.

### Revision 6

Status: revision 6, with the amendments of D-10, D-11 and D-12, is approved by the leader (its notes on the card of 2026-11-24T07:45:22Z, 08:08:17Z, 09:33:25Z and 10:22:48Z; leader judgments under the user's delegation of 2026-11-23T13:57Z and not the user's own pick) and built. Branch `spec/go-path-removal-20261121-r2`. It started from `NeXT` at `bd7ee6a41f` and merged the tip `974267e03d` before its last commits, as the leader asked, without a conflict. Every run below was made once in this worktree with `RUSTC_WRAPPER=sccache` in a heavy slot; the logs are files of the worktree's ignored `target/go-path-removal/` and are named here.

RED, commit `e412760925` (the tests against the unchanged code): `cargo nextest run --offline --locked --no-fail-fast --test update_switch_layer --test update_switch_cli` ran 27 tests, 9 passed and 18 failed, exit 100 (`red-run.log`); the 9 that passed pin what the change keeps. GREEN, commit `dfc49a0b55`: the same two targets with `repo_scans_table`, `help_text_names_moai` and `cli_text_citations` ran 64 tests, 64 passed, exit 0 (`green-l2.log`). The first green run (`green-l.log`, 26 of 27) showed one expectation of a test wrong — the product asks the launcher for its version in the preview and once more in the apply — and the test was corrected; the product did not change for it. The fresh-shell test of commit `68ac20a67d` ran 2 tests, 2 passed (`green-e2.log`); its first run (`green-e.log`, 1 of 2) showed the count of AC-22 wrong (D-11). clippy with warnings denied on the changed targets and tests: exit 0 (`clippy-l.log`, `clippy-e.log`, `recheck.log`). `cargo fmt --all -- --check`: exit 0. `tools/smoke/run.sh update` on the debug binary built from this worktree: `smoke: update pass steps=4 time=3.9s` (`smoke-update.log`).

The verify commands as this revision writes them were run as they are (`ac.sh`, `ac-verify.log`, `recheck.log`). AC-1 found a row for each of the 60 words of the Go line, AC-2 printed `rows=77 with-disposition=77`, AC-3 counted 7 lines, AC-4 and AC-20 exited 0, and AC-5 printed 0 paths outside its list. AC-6 passed 11, AC-7 14, AC-8 6, AC-9 2, AC-10 4, AC-11 2, AC-12 6, AC-13 7 and AC-15 1, each exit 0. AC-14 as revision 5 wrote it exited 1 because `cargo test` refused its second filter (D-12); in the corrected form it passed 25. The same pass found the state-folder literal count of the two switch test files one over the baseline for the second preserved copy they read: the two rows of the baseline were raised (commit `356f4b82db`) and the CLI test was brought back to its old count (commit `021f683430`); the table scans then ran with it, 70 tests, 70 passed (`re-tables2.out`).

The dangerous path ran for real with the debug binary built from this worktree, each case in its own throwaway HOME under `target/go-path-removal/` with a bare environment and git configuration pinned to nothing (`dangerous-path.log`, 2026-11-24T09:28:31Z). v1 and v2 beside a Rust launcher are offered `link`, applied with exit 0 into `link -> moai-cli` with one kept copy, and reverted with exit 0 to the script. An edited script, no launcher and a launcher answering with a Go line are offered `left` or `held` with no token, and a token forced on them is refused with exit 1 while the script and the folder stay as they were. A stop after the kept copy and a stop before the rename both end with exit 1, the script whole and one kept copy; the next confirmed run finishes with two kept copies and reverts. A plain file put in place of the link by hand makes the revert refuse with exit 1 and leaves that file and the kept copy.

Revision 6 changes the verify commands of AC-1, AC-5 and AC-14 (D-3, D-12). The path list AC-5 had when the git-hook slice was checked named, beside this SPEC folder: `src/git_hooks/`, `tests/git_hooks_cli.rs`, `install/git-hooks/`, `install/commit-msg`, `src/ops/git_hooks_*`, `docs/design-review/appendix/operations/git_hooks_*`, `tests/update_git_hooks_cli.rs`, `tests/update_switch_layer.rs`, `tests/update_tux_git_hooks.rs`, `tools/repo_scans.json`, `tests/fixtures/help/`, `src/help.rs`, `src/main.rs`, `src/lib.rs`, `src/update.rs` and `src/update_tux.rs`, against the base `v2026.11.0`.

- AC-16: met — 27 run, 27 passed (`update_switch_layer` and `update_switch_cli`).
- AC-17: met — the same run, 27 passed.
- AC-18: met — 11 run, 11 passed (`update_switch_cli`).
- AC-19: met — 36 run, 36 passed (`go_path_retired_cli`, `bundle_build_cli`, `bundle_verify_cli`).
- AC-20: met — the verify command exits 0.
- AC-21: met — 18 run, 18 passed.
- AC-22: met — 7 run, 7 passed (`go_path_retired_cli`, `cli_name_moai`).

Not run for revision 6: the real HOME, its `PATH` lines and the whole earlier folder (the user steps under Plan); a fresh shell on the real machine (the user's look, D-9); Windows and Linux (the step makes a unix symbolic link, and on another platform the apply refuses before its first write, which was not executed there); a HOME whose state folder is the new name (the tests build the earlier layout only, and the step takes its folder from the HOME resolver, a line this change did not touch); the Python wrapper test (edited as text and not run, D-8).

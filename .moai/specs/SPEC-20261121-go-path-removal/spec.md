---
id: SPEC-20261121-go-path-removal
title: "moai drops the Go path: inventory refresh, Go-link removal in update, name-switch retirement"
status: draft
revision: 5
created: 2026-11-21
---

# moai drops the Go path: inventory refresh, Go-link removal in update, name-switch retirement

## Intent

사용자 지시(2026-11-21, 리더 전달): "go하고도 연결 모두 끊어라", "go path 완전 제거해라", "모든 moai-adk-go 버전을 사용중단하고 moai update 하면 기존의 go와 연결 고리 제거후 업데이트 하고 moai-adk-go 관련 내용은 제거한다". 제품이 더 이상 어떤 경로로도 Go 빌드를 실행하거나 설치하지 않게 하는 카드다.

기계 상태(2026-11-21 기준): `moai` 는 `~/.local/bin/moai` → `~/.moai/bin/moai` → `~/.moai/versions/4.2647.15/bin/moai` 로 바로 Rust 빌드로 연결된다. 이름 전환 shim(`~/.moai-cli/path/moai`, v2)과 보관 사본 네 개는 PATH 에서 내려온 고아 파일이 됐고, `~/go/bin` 에는 Go moai 바이너리가 없다(사용자가 휴지통으로 치움, 다른 Go 도구는 남음).

전체 계획은 세 조각이다: (1) 읽기 전용 목록 조사 — 이 revision 이 넘기는 것. (2) `moai update` 가 모든 사용자의 Go 연결고리를 제거한다(옛 런처 스크립트와 보관 사본, 그 이름을 부르는 PATH 줄, 프로젝트·홈 설정의 Go 세대 훅 줄 — 미리보기·적용·되돌리기 기록과 함께, 사용자 파일을 조용히 지우지 않고, Go 바이너리 자체는 사용자가 지우도록 목록으로 보여 준다). (3) 제품에서 이름 전환 스크립트와 그 update 단계와 Go 라우팅 낱말 목록을 제거하고 문서·스킬의 Go 시절 글을 정리한다. go-state·migrate 도구는 update 이식이 필요로 하는 동안만 남는다. Go 가드를 내리기 전에 보안 가드 이식 카드가 먼저 착지해야 하고, SPEC 은 영향을 받는 가드를 적어야 한다. 조각 2·3 읊의 세부 AC 는 각 조각 착수 때 이 SPEC 의 다음 revision 에서 갱신한다. HOME 경로 카드(폴더 이동과 설정 줄을 소유)와 겹치는 항목은 그 카드와 조율한다.

기준 목록: SPEC-20261119-rust-only-moai 의 `inventory.md`(가지 spec/20261119-rust-only-moai, head 903061f05233a42fa8c014d3e0819b70871fe19b). 이 revision 의 목록은 그 기준을 2026-11-21 기계 상태와 현 tip 에 맞춰 갱신한다.

조각 순서 결정(2026-11-21, 카드 결정 노트 3건 — 사용자 "진행해라" 승인): 목록 조각 바로 다음에, `moai update` 가 사용자 저장소의 Go 세대 git 훅을 찾아(곁의 해시 파일과 머리말 글로 인식) 미리보기에 보여 주고, 적용 때 되돌리기 기록과 함께 은퇴시키고, 그 자리에 Rust 훅을 설치한다. MoAI 표시가 없는 사용자 작성 훅은 건드리지 않고 "kept" 으로 나열한다. 이 슬라이스가 사용자가 처음 만나는 부분이다(Go 바이너리가 사라진 뒤 실패하는 커밋). 제품 작업: 어떤 프로젝트에도 이 git 훅을 설치·미리보기·제거하는 moai 명령 하나와, 그 명령을 은퇴 직후에 쓰는 update 단계. 트리 안의 `install/commit-msg` 는 옛 런처 경로를 고정하고 있어 plain `moai` 를 부르게 고친다. 모든 훅은 100ms 안쪽이고, 시험은 Rust 로 임시 저장소에서: Go 훅 은퇴 후 교체, 사용자 훅 바이트 그대로 보존, 섞인 폴더, 두 번째 실행은 변화 없음, 되돌리기 복원. pre-push 가 Rust 검사를 돌릴지(예: 밀린 범위의 trailer-check)는 그 조각의 revision 이 이 SPEC 에서 정한다. 리더가 이 기계의 공유 훅 폴더에서 Go 세대 훅을 이미 은퇴시키고 작은 Rust 훅 둘(commit-msg, pre-commit)을 넣었다 — 목록의 기계 상태에 기록했다.

## Non-goals

- 이식·퇴역·shim 제거의 실행: 이 revision 은 읽기만 한다(조각 2·3 의 몫).
- Go 도구체인, `~/go/bin` 의 다른 도구(buf, actionlint 등), 다른 Go 프로젝트, 열린 세션, 작업 트리는 건드리지 않는다.
- 실제 HOME 에 쓰기: 읽기만 한다.
- Go 하위 명령 실행: `--help` 없이는 실행하지 않는다(이 기계엔 Go moai 가 없어 실행 자체가 불가능하다).

조각 본체(구현, 리더 결정 노트 20 기준): 피처 모듈은 `src/git_hooks/`(신규 폴더 — Go 세대 훅 인식·미리보기·적용·되돌리기), 훅 텍스트는 `install/git-hooks/`(pre-commit·pre-push), 시험은 `tests/git_hooks_cli.rs`(임시 저장소, `#[path]` 포함으로 배선 없이 단독 검증 — quality_gate 조각 선례), 그리고 `install/commit-msg` 는 plain `moai` 를 부르게 고친다. `src/update.rs` 와 `src/lib.rs` 는 update를 고치는 레인이, `src/main.rs` 는 명령 연결 레인이 쥐고 있어 pair request 노트를 남겼고(각 홀더 카드의 노트 5800·5801), update 레이어 연결과 `git-hooks-plan|apply|revert` 명령어 연결은 홀더와 조율해 붙인다. pre-push 는 리더가 받은 기본안으로 정한다: 밀린 범위의 trailer-check(카드 번호 스캔 포함), 100ms 안쪽, moai 가 PATH 에 없으면 안내문과 함께 통과. 12개 저장소의 막힌 pre-commit 을 레인이 손대는 일은 없다(리더가 사용자에게 보고).

연결 revision(3, 리더 지시 "같은 조각을 끝까지"): 이 조각이 이름 붙이고 편집하는 파일은 다음과 같다. (1) `src/lib.rs` — 정렬된 모듈 목록에 `pub mod git_hooks;` 한 줄(update 레인 ACK 노트 5803 승인). (2) `src/git_hooks/update.rs`(신규, 내 모듈 안의 하위 모듈) — update 흐름의 Layer(전용 토큰 --confirm-git-hooks, 미리보기 JSON 부분, 적용·되돌리기 위임). (3) `src/update.rs` — switch 레이어 옆의 새 한 줄들: UserLayerPlan 필드, plan 호출, 생성자 줄, asks 토큰 경로와 적용 순서 목록(홀더 범위 안이라 정확한 줄을 노트 5820 로 요청, ACK 5822 로 반영), 그 밖의 summary·JSON·적용 발동·되돌리기는 홀더 범위 밖에 새로 넣는다. (4) `src/ops/git_hooks_plan.rs`·`git_hooks_apply.rs`·`git_hooks_revert.rs`(신규) — build.rs 가 src/ops/*.rs 를 자동 수집해 명령이 생기므로 main.rs 편집이 필요 없어졌고(main.rs 홀더 ACK 노트 5812 는 적용되지 않음), 각 명령의 문서 조각 `docs/design-review/appendix/operations/git_hooks_<word>.md`(신규 3개)가 전수 검사의 요구다. (5) `tests/update_git_hooks_cli.rs`(신규) — 임시 저장소에서 명령과 update 레이어를 검증. help.rs 행은 update 레인 몫으로 같은 노트에 선택 요청 — 빠지면 이 revision 이 기록한다.

## Risks

- 기계 상태는 하루 안에도 움직였다(11-19 목록 이후 shim 고아화, Go 바이너리 퇴장, 자산 번들 착지, 공유 훅 교체). 목록의 수는 읽은 시점의 것이고, 조각 2·3 실행 전에 관련 행을 다시 세야 한다.
- 판정을 갈린 호출자 수에 의존한다. 게이트 호출 12곳처럼 실시간으로 막힌 경로는 목록이 아니라 수리(슬라이스)가 닫는다 — 목록만으로 커밋 막힘이 풀리지 않는다.
- 폴더·파일 제거는 사용자 파일을 다룬다(shim 보관 사본, Go 세대 훅). 조용한 삭제 없이 미리보기·토큰·되돌리기 기록이 없는 조각은 착지하면 안 된다.
- 알려진 한계(안전한 쪽으로 틀린다): (a) 이 기계에서 찾은 Go 세대 변형 중 다음은 알아보지 못해 사용자 것으로 둔다 — 빠른 부분집합 머리글의 pre-commit(복사 사본들), 명령 줄이 없는 pre-push(여러 저장소), 손으로 깐 Rust pre-commit(출하본과 해시가 다름, 약 14곳). 못 알아본 훅은 지워지지 않고 `kept` 로 보고된다. (b) pre-push 는 새 브랜치·새 태그를 훑지 않는다(범위가 없어서). 원격 끝이 로컬에 없는 강제 푸시도 범위를 못 만들어 안내만 하고 통과시킨다. sha256 저장소의 새 브랜치도 같은 이유로 안내와 함께 통과한다. (c) 읽을 수 있는 사용자 편집이 은퇴한 이름 위에 있으면 그 이름만 되돌리지 않고 `blocked` 로 보고하며 보관본을 폴더에 남긴다.

## Requirements

- REQ-1: When the inventory is read, it shall list every word on the Go line of the name-switch shim in install/moai-switch.sh and one row per routed spec session and bundle action. · failure: none
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

## Acceptance

- AC-1 (REQ-1): every word on the shim's Go line has a row in `inventory.md`.
  verify: `for w in $(awk '/^  af\|agent/{sub(/\)$/,"");sub(/^  /,"");n=split($0,a,"|");for(i=1;i<=n;i++)print a[i]}' install/moai-switch.sh); do grep -q "^| $w " .moai/specs/SPEC-20261121-go-path-removal/inventory.md || { echo "missing $w"; exit 1; }; done`
- AC-2 (REQ-2): every row of the word and action tables names one of the three dispositions.
  verify: `awk '/^## Words/{f=1;next} /^## Execution paths/{f=0} f&&/^\| [a-z]/&&!/^\| (word|action|---)/{n++; if ($0 ~ /\| (rust-has|port|retire) /) d++} END{print "rows="n" with-disposition="d; exit (n==d && n>0 ? 0 : 1)}' .moai/specs/SPEC-20261121-go-path-removal/inventory.md`
- AC-3 (REQ-3): the inventory has a machine-state section naming the resolution chain, the off-PATH shim with its saved copies, and the emptied Go bin folder of the HOME.
  verify: `grep -c -e 'versions/4.2647.15' -e 'moai.before-' -e 'no Go moai binary' .moai/specs/SPEC-20261121-go-path-removal/inventory.md`
- AC-4 (REQ-4): the execution-path table has rows for installer, update, launcher, hooks, statusline, commit-msg, bundle-packs and the tests.
  verify: `for p in installer update launcher hooks statusline commit-msg bundle-packs tests; do awk '/^## Execution paths/{f=1;next} f&&/^\| '"$p"'/{found=1} END{exit !found}' .moai/specs/SPEC-20261121-go-path-removal/inventory.md || { echo "missing $p"; exit 1; }; done`
- AC-5 (REQ-5): the change touches only the paths this SPEC names.
  verify: `git diff --name-only $(git merge-base HEAD v2026.11.0) HEAD | grep -v -e '^\.moai/specs/SPEC-20261121-go-path-removal/' -e '^src/git_hooks/' -e '^tests/git_hooks_cli\.rs$' -e '^install/git-hooks/' -e '^install/commit-msg$' -e '^src/ops/git_hooks_' -e '^docs/design-review/appendix/operations/git_hooks_' -e '^tests/update_git_hooks_cli\.rs$' -e '^tests/update_switch_layer\.rs$' -e '^tests/update_tux_git_hooks\.rs$' -e '^tools/repo_scans\.json$' -e '^tests/fixtures/help/' -e '^src/help\.rs$' -e '^src/main\.rs$' -e '^src/lib\.rs$' -e '^src/update\.rs$' -e '^src/update_tux\.rs$' | wc -l` prints 0.
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
  verify: `cargo test --offline --locked --test git_hooks_cli recognition lifecycle`
- AC-15 (REQ-5): no shipped text or source comment still describes the retired commit-msg install or the replaced-hook flow.
  verify: `cargo test --offline --locked --test git_hooks_cli the_spec_and_shipped_texts_carry_no_stale_claims`

## Scope

- write: `.moai/specs/SPEC-20261121-go-path-removal/**`, `src/git_hooks/**`, `tests/git_hooks_cli.rs`, `install/commit-msg`, `install/git-hooks/**`, `src/ops/git_hooks_plan.rs`, `src/ops/git_hooks_apply.rs`, `src/ops/git_hooks_revert.rs`, `docs/design-review/appendix/operations/git_hooks_plan.md`, `docs/design-review/appendix/operations/git_hooks_apply.md`, `docs/design-review/appendix/operations/git_hooks_revert.md`, `tests/update_git_hooks_cli.rs`, `tests/update_switch_layer.rs`, `tools/repo_scans.json`, `tests/fixtures/help/**`, `src/help.rs`, `src/main.rs`, `src/update_tux.rs`, `tests/update_tux_git_hooks.rs`, and the pair-approved lines of `src/lib.rs` and `src/update.rs` (the update lane's ACK note 5803 + the named lines of note 5820, and the main.rs and help.rs holders' ACKs on their cards)
- read-only: the rest of `src/`, `tests/`, `install/`, `docs/`, `skills/`, `tools/`, `templates/`, the installed Rust binary (read-only probes, temporary HOME), and this machine's state
- forbidden: the 12 repositories' hook folders on this machine, any write in the real HOME, any run of a Go subcommand

## Plan

1. Read the baseline inventory and the current tip sources: `install/moai-switch.sh`, `src/update/switch.rs`, `src/update.rs`, `src/update_tux.rs`, `src/shim_check.rs`, `tools/claude-statusline.sh`, `src/go_hooks.rs`, `src/user_assets.rs`, `src/ops/bundle.rs`, `install/commit-msg`, `tests/diff_go.rs`, the installer and launcher sources.
2. Probe the installed Rust binary with every routed word in a temporary HOME and record what it answers today.
3. Re-verify the machine-side counts the baseline recorded (project hook folders, git pre-commit callers, MCP registrations, shell rc lines) and record what changed.
4. Write `inventory.md` with the machine state, the word table, the action table, the execution-path table and the findings, and keep every claim on a command run for this file.

---
id: SPEC-20261121-rm-guard
title: "delete_guard: refuse recursive deletes of the root, the home, the project root or an empty-variable path"
status: draft
revision: 1
created: 2026-11-21
---

# delete_guard — 재앙급 재귀 삭제를 3겹으로

## Intent

한 레인의 감사 서브에이전트가 `rm -rf $M`($M=$S/$1, 비어 있으면 `/`로 풀림)을 만들었고 Claude
Code 자체 프롬프트("Dangerous rm operation on possibly-empty variable path")만이 이를 잡았다 —
moai 의 훅 가드는 잡지 못했다. 사용자 지시(2026-11-20T22:5xZ): `rm -rf /` 같은 삭제는 Claude
Code·Codex 설정과 moai-adk 자체 가드, 세 겹으로 막혀야 하고 프로젝트 루트 전체 삭제도 막혀야
한다. 기존 가드(`src/hook_guards.rs` 의 `is_dangerous_rm_target`)는 literal 형태만 알아본다 —
일반 변수 경로, 프로젝트 루트와 그 부모, `find -delete` 같은 다른 재귀 삭제가 빠져 있다.

## Non-goals

- `src/hook_guards.rs` 의 연결 편집과 `src/lib.rs` 의 모듈 선언 — C-234(M-Flow forward guard,
  같은 파일 보유)가 READY 된 뒤 별도 짧은 조각으로(리더 결정, 카드 노트 4657 답). 이 SPEC 은
  판정 모듈·update 제안·시험만 착지시킨다.
- 실제 HOME 설정에 deny 규칙을 곧바로 적용 — 미리보기와 사용자 선택을 거치는 guarded 단계로,
  이 SPEC 은 설치 제안 코드와 시험만 만든다.
- `git clean` 정책 변경 — 루트에서 `-x` 없는 `git clean` 은 그대로 허용(카드 결정 사항: 루트에
  `-x` 는 선택 없이는 금지지만 훅 연결 조각에서 다룬다).
- Claude Code·Codex 이외의 호스트(Omp 등) — 후속.

## Requirements

- REQ-1: The pure function `delete_guard::check` shall classify a recursive-delete target argument (`rm` with a recursive flag: `-r`, `-R`, `-rf`, `-fr`, `-Rf`, `-rR`, `--recursive` in any spelling, and `rm -r` without `-f`) as refused when it resolves to the filesystem root (`/`, `/*`, trailing slashes, `/.`), the home folder (`~`, `~/`, `$HOME`, `${HOME}`, each also with `/*`), a bare wildcard, `.` or `..`, a top-level directory such as `/usr`, or a `.git` directory · failure: none — a non-recursive `rm -f` or any argument that is not one of these shapes returns no refusal.
- REQ-2: The same function shall refuse a recursive-delete target built from a shell variable that may be empty or unset — `$VAR`, `${VAR}`, a quoted form `"$VAR"`, either followed by `/` or `/*` — because an empty expansion deletes the folder the command runs in · failure: none — a variable guarded by `:?` (`${VAR:?}`) or expanded with a non-empty literal default (`${VAR:-x}`) is allowed, and a bare word that merely contains `$` in the middle is not a variable path.
- REQ-3: When the caller passes the project root (the value of `git rev-parse --show-toplevel` in the folder the command runs in, compared as an absolute path after normalization), the function shall refuse a recursive-delete target equal to the project root, equal to its parent, or equal under trailing-slash, `/.`, duplicate-slash or quoted-then-split forms, and it shall still allow a named child of the project root such as `build` or `target/debug` · failure: none — with no root passed the root-based checks are skipped and the other checks still run.
- REQ-4: The function shall classify other recursive-delete shapes by the same target rules: `find <target> -delete` (and `-exec rm -rf {} +`) where the target is refused, `find` depth-ordered deletes of the refused targets, `xargs`-piped `rm -rf` reading the refused targets, and `sudo`-wrapped forms of any refused command · failure: none — a `find` that deletes a named child, or an `xargs` pipeline whose arguments carry no refused target, returns no refusal.
- REQ-5: The update flow shall offer to write Claude Code `permissions.deny` entries for the dangerous `rm` shapes into the host settings, showing the entries and the quoted official permission-rule syntax in the preview and applying only on the user's pick · failure: none — without the user's pick nothing is written, and a refusal of the settings write leaves the settings file byte-identical.
- REQ-6: The update flow shall name the Codex equivalent: when the official Codex docs express these shapes it offers the same preview-and-pick, otherwise it records in the preview that Codex cannot express them and the hook stays the defense · failure: none — the offer must not cite a syntax the docs do not define.
- REQ-7: Safe deletes keep working: a named subfolder (`rm -rf build`), a literal scratch path (`rm -rf /tmp/x`), a non-recursive `rm -f file`, and an `echo` or `grep` that merely mentions the dangerous text are all allowed · failure: none — refusing one of these is a defect.

## Acceptance

- AC-1 (REQ-1..4, REQ-7): 단위 시험 — 위험 형태마다 거절 이유, `${VAR:?}`·`:-` 가드·하위 폴더·리터럴 스크래치·비재귀·언급형 허용, 경계(/tmp 그 자체 vs 자식, 프로젝트 루트 vs 그 하위, 부모), 이유 문장이 무엇인지 말함. verify: `cargo test --locked --offline --test rm_guard_cli`
- AC-2 (REQ-5, REQ-6): update 제안 시험(임시 HOME·프로젝트) — 미리보기에 deny 항목과 공식 문서 문법 인용이 보이고, 선택 뒤에만 settings 가 바뀌며, 거절하면 바이트 동일; Codex 문장이 문서 근거를 말함. verify: `cargo test --locked --offline --test rm_guard_cli`
- AC-3: 이웃 검사. verify: `cargo test --locked --offline --test hooks_cli`; `cargo fmt --check`; `cargo clippy --locked --offline --all-targets -- -D warnings`

## Scope

- write: `src/delete_guard.rs`(신설 — 판정 로직), `src/update_hooks.rs`(deny 설치 제안), `tests/rm_guard_cli.rs`(신설), `.moai/specs/SPEC-20261121-rm-guard/**`
- read-only: `src/hook_guards.rs`(연결 지점 확인), `src/hosts/codex.rs`·`src/hosts/codex_config.rs`(Codex 설정 모양), `tests/hooks_cli.rs`(이웃), Claude Code·Codex 공식 문서(웹)
- forbidden: `src/hook_guards.rs` 편집(C-234 보유 — 연결은 뒤 조각), `src/lib.rs` 편집, 실제 HOME 설정 변경, 카드 번호를 SPEC id·가지·커밋에 사용

## Plan

1. RED: `tests/rm_guard_cli.rs` — `#[path]` include 로 `src/delete_guard.rs` 를 읽어 판정 표(REQ-1..4, REQ-7)와 경계를 검증; lib.rs 선언이 없어도 컴파일.
2. GREEN: `src/delete_guard.rs` — recursive 플래그 파서 + 대상 분류(REQ-1..4) + 프로젝트 루트 인자 처리(REQ-3).
3. update_hooks.rs — Claude Code deny 설치 제안(REQ-5)과 Codex 안내(REQ-6), 문서 인용; 시험 AC-2.
4. 이웃 검사 + READY 폭 검사; 연결 조각(lib.rs 선언·hook_guards.rs 위임)은 C-234 READY 뒤.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | 판정 로직은 신설 `src/delete_guard.rs` — hook_guards.rs 연결은 C-234 READY 뒤 짧은 조각 | 같은 파일을 두 p0 가 쥐는 경합 회피(리더 결정, 노트 4657 답) | 모듈을 hook_guards.rs 로 통합 |
| D-2 | 시험은 `#[path = "../src/delete_guard.rs"]` include 로 lib.rs 없이 판정만 검증; 연결 착지 뒤 훅 CLI 시험을 추가 | lib.rs 선언이 연결 조각에 속함; include 는 시험을 RED→GREEN 로 돌릴 수 있게 함 | lib.rs 선언 뒤 CLI 시험으로 대체 |
| D-3 | 프로젝트 루트는 판정 함수의 인자 — 순수 함수 유지, 훅 호출부가 toplevel 을 넘김 | 단위 시험이 경계를 직접 재게 함; 판정이 git 을 실행하지 않음 | 판정 함수가 git 실행 |
| D-4 | deny 규칙은 보조 방어 — 공식 문법이 재귀 삭제를 완전히 못 표현하면 그 한계를 인용과 함께 기록하고 훅이 주 방어 | Claude Code deny 는 도구·패턴 규칙; rm 의 셸 풀이 전체를 패턴으로 못 잡을 수 있음 | 문서 확인 결과에 따라 문장 수정 |
| D-5 | `${VAR:?}`·`${VAR:-lit}` 가드 형태는 허용 | 셸 표준 — 비어 있으면 명령을 중단시키는 형태 | 전부 거절 |
| D-6 | 기존 literal 목록과의 단일화는 연결 조각에서 — rm_guard 가 delete_guard::check 로 위임 | 중복 판정은 시험 두 벌을 만들어 흐트러짐 | 위임 없이 병존 |

## Risks

- deny 규칙은 도구·패턴 수준이라 셸 풀이 결과까지 못 잡을 수 있음 — 훅(주) + deny(보조) + Claude Code 자체 프롬프트(있으면)의 겹 방어로 기록.
- 프로젝트 루트 판정은 훅 cwd 의 toplevel 에 의존 — cwd 가 프로젝트 밖이면 root-based 검사는 건너뛰고 다른 검사는 계속(REQ-3 failure 문장).
- `find -delete`·`xargs` 파이프는 셸 파서 의존 — 연결 조각에서 hook_guards.rs 의 shell_commands 파서와 같은 입력을 쓰는지 시험으로 확인.

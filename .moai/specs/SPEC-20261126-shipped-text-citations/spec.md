---
id: SPEC-20261126-shipped-text-citations
title: "배포 파일에는 카드 번호, Issue 번호, 결정 인용이 없어요"
status: draft
revision: 1
created: 2026-11-26
---

# 배포 파일에는 카드 번호, Issue 번호, 결정 인용이 없어요

## Intent

`skills/`, `agents/`, `templates/`, `install/` 의 파일은 사용자에게 그대로 가요. 사용자는 거기 적힌 moai-adk 의 Issue 번호, 카드 번호, "user decision" 인용, 저장소 이름을 따라갈 수 없어요. 첫 릴리스 전에 지우고, Rust 검사가 다시 들어오는 것을 막아요. 설치 문서는 설치 스크립트가 묻는 스킬 설치 질문과 `--yes` 를 설명해요.

## Non-goals

- 배포되지 않는 개발자 글: `src/`, `tests/`, `tools/`, `docs/` 의 주석과 문서.
- `templates/bundle-packs/`: 옮겨 온 팩 내용이라 검사에서 빼요(리더 결정, 원장 노트 5089).
- `skills/moai/references/source-manifest.json` 의 `origin.reason` 문장(base 에서 29줄): 검사의 PENDING 목록에 둬요(D-2).
- `skills/moai/workflows/audit.md` 의 한국어 표지(`## 독립 review` 등): `src/self_audit.rs` 와 `src/pr_link.rs` 가 글자 그대로 읽는 기계 표지예요(D-3).
- SPEC id: 검사가 잡지 않고 지우지 않아요(D-4).
- `LANE-*` 표지, `templates/github/labels.yml` 의 "a user decision is pending" 문장, 설치 언어 메뉴의 `한국어`, 표시 이름 "프로젝트 맞춤 하네스": 기능 문장이에요.
- `install/moai-switch.sh` 는 고치지 않아요(바이트 해시로 알아보는 파일).

## Requirements

- REQ-1: The shipped files outside the exempt pack folder and the pending list shall carry no card number, no Issue or pull-request number, no user or leader decision citation and no name of a moai-adk development script or repository.
- REQ-2: The shipped-text scan shall leave machine markers, SPEC ids, CSS colour codes and the one card-id format line of `skills/moai/references/audit.md` unflagged.
- REQ-3: The shipped-text scan shall fail when an exception it names (the exempt folder, a format line or a pending file) is no longer the named one or is no longer needed.
- REQ-4: A cleaned sentence shall keep everything it tells the reader, losing only the citation or the repository name.
- REQ-5: `install/README.md` shall state the question both installers ask before they install the MoAI skill and agents, and when they ask it.
- REQ-6: `install/README.md` shall list `--yes` (`-yes`) in its options table with what it does when nobody can be asked.
- REQ-7: The commit that changes a shipped sentence a test pins shall carry the new wording in that test.
- REQ-8: The source manifest shall carry the digest of every changed file of the skill package.
- REQ-9: The repository scan table shall run the shipped-text scan for a change under any of the four shipped folders.

## Acceptance

- AC-1 (REQ-1, REQ-2): 보기 줄에서 카드 번호, Issue·PR 번호(여러 표기), 결정 인용, 개발 도구 이름은 잡고, 표지·형식·색 코드·SPEC id 는 안 잡아요 · verify: `cargo test --offline --locked --test skill_text_citations scan_flags_each_kind_and_leaves_markers_alone`
- AC-2 (REQ-1, REQ-4): 배포 파일 검사가 통과해요(base 에서는 6개 파일로 실패) · verify: `cargo test --offline --locked --test skill_text_citations scan_shipped_files_outside_the_pending_list_are_clean`
- AC-3 (REQ-3): 예외가 하나씩 이름 붙어 있고 아직 필요해요 · verify: `cargo test --offline --locked --test skill_text_citations scan_exceptions_stay_named_and_needed`
- AC-4 (REQ-5, REQ-6): 설치 문서가 질문 문장과 옵션 표의 `--yes` 줄을 가져요(base 에서는 실패) · verify: `cargo test --offline --locked --test skill_text_citations installer_document_names_the_skill_question_and_the_yes_option`
- AC-5 (REQ-4, REQ-7, REQ-8): 바뀐 글을 읽는 시험이 통과해요 · verify: `cargo test --offline --locked --test update_route --test product_cli --test docs_installer_paths --test card --test launcher_retire_role_flags --test next_flow_skill_scan --test agent_defs_package --test flow_name_scan --test skill_generalization_scan`
- AC-6 (REQ-9): 검사가 표에 올라 있고 표의 시험이 통과해요 · verify: `cargo test --offline --locked --test repo_scans_table --test repo_scans_cli`

둘째 수락 기준의 실행 파일 쪽(첫 실행 권한 참조 문구에 내부 번호 없음)은 base 에서 이미 맞아요: `src/launcher_init.rs:26`, 검사 `tests/cli_text_citations.rs`(이 파일은 PENDING 에 없어요).

## Scope

- write: `.moai/specs/SPEC-20261126-shipped-text-citations/spec.md`
- write: `tests/skill_text_citations.rs`
- write: `skills/moai/workflows/update.md`
- write: `skills/moai/references/adapters/aside.md`
- write: `install/install.sh` (주석만)
- write: `install/install.ps1` (주석만)
- write: `install/README.md`
- write: `templates/github/manifest.json` (`license`, `contract` 문장)
- write: `tests/update_route.rs` (고정 문장 1곳)
- write: `docs/cli-text-rationale.md` (지운 인용 기록)
- write: `skills/moai-html-report/references/artifact-contract.md`, `skills/moai-html-report/references/fonts.md` (날짜 붙은 결정 인용 한 줄씩; 감사 FAIL 노트 10196 뒤 리더 지시로 더함)
- shared (card line-lease): `tools/repo_scans.json`, `skills/moai/references/source-manifest.json` (바뀐 파일의 sha256 줄만)
- read-only: `src/launcher_init.rs`, `src/self_audit.rs`, `src/pr_link.rs`, `install/moai-switch.sh`
- forbidden: `src/**`, `tools/*.py`

## Plan

1. RED: 새 검사 `tests/skill_text_citations.rs`(감사 PASS 였던 옛 검사의 규칙을 옮기고 지금 트리에 맞춤).
2. GREEN: 6개 파일에서 인용만 지우고, 설치 문서에 스킬 설치 질문과 `--yes` 를 쓰고, 고정 시험·원장 sha256·검사 표를 맞춰요.
3. 변형 확인(새 시험마다 하나), clippy, 영수증, submit.

## Risks

- 막는 검사가 새로 생겨요. base 에서 막히는 것: `install/README.md` 6줄, `install/install.ps1` 8줄, `install/install.sh` 7줄, `skills/moai/references/adapters/aside.md` 1줄, `skills/moai/workflows/update.md` 7줄, `templates/github/manifest.json` 10줄(6개 파일), PENDING 의 원장 29줄. 찾아서 고친 오탐: `skills/moai-html-report/` 9개 파일의 CSS·SVG 색 코드, `templates/github/ko/ISSUE_TEMPLATE/moai-investigation.yml` 의 "사용자 결정이 났다" 1줄.
- 앞으로 스킬 글이나 원장이 아닌 배포 파일에 번호를 쓰는 카드는 이 검사에 막혀요. 정말 형식 예시면 `FORMAT_LINES` 에 한 줄을 더해야 해요.
- 색 코드 예외(`color:`, `background:` 뒤)는 그 자리의 인용을 놓칠 수 있어요.
- 원장(`source-manifest.json`)은 다른 카드도 바꿔요. 양쪽이 바꾼 머지는 리더가 해요(노트 9788). 레인은 그런 짝에 `git merge`·`merge-tree` 를 돌리지 않아요.
- PowerShell 설치 스크립트는 `pwsh` 가 없으면 실행 확인이 `not_run` 이에요. 고치는 것은 주석뿐이에요.
- 위험 경로 실행: 해당 없음(글과 시험만 바꾸고, 삭제·프로세스 시작·권한 규칙을 만드는 코드가 없어요).

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | 옛 가지 3개는 머지하지 않고 검사 파일의 규칙만 옮겨요 | 옛 head(`b508a4e2cc` 등)는 물러난 흐름의 글 위에 있고 base 와 충돌해요. 원장 머지 드라이버는 python3 을 불러요 | 옛 가지는 그대로 있어요 |
| D-2 | 원장(`source-manifest.json`)의 인용 줄 30개(`origin.reason` 29줄과 `status` 1줄)는 PENDING 에 두고 후속 카드가 지워요 | 리더의 위임 결정(원장 노트 9921 (b), 10144). 공용 파일이라 문장을 많이 고치면 머지 충돌이 커요 | PENDING 항목을 지우고 문장을 고치면 돼요 |
| D-3 | `workflows/audit.md` 의 한국어 표지는 남겨요 | 검사기와 PR 본문 생성기가 그 글자를 읽어요. 글만 바꾸면 안내한 절을 명령이 거절해요 | 검사기가 두 표기를 읽게 하는 카드로 |
| D-4 | SPEC id 는 대상이 아니에요 | 카드 본문의 대상 목록에 없고, 감사 PASS 였던 검사도 잡지 않았고, `tests/flow_name_scan.rs` 가 하나를 고정해요 | 검사에 종류 하나를 더하면 돼요 |
| D-5 | `LANE-*` 는 잡지 않아요 | 실행 파일이 읽는 프로토콜 표지예요(리더 결정, 노트 3711) | 검사에 낱말을 더하면 돼요 |
| D-6 | `templates/bundle-packs/` 는 검사에서 빼요 | 리더 결정, 노트 5089 | `EXEMPT_PREFIXES` 를 비우면 돼요 |
| D-7 | 지운 인용은 `docs/cli-text-rationale.md` 끝에 적어요 | 새 문서를 검사 표가 이름하면 `tools/ci_scope.py`(Python)를 고쳐야 해요 | 문서 부분을 옮기면 돼요 |
| D-8 | `references/audit.md` 의 판정 보기 줄(`AUDIT` 로 시작하고 카드 번호 형식을 보여 주는 줄)은 남겨요 | 카드 번호 형식의 예시이고 `tests/gate_collect.rs` 가 그 줄을 읽어요 | `FORMAT_LINES` 를 비우고 줄을 바꾸면 돼요 |

## Result

- 감사 FAIL(head b46308adbb, 원장 노트 10196)의 수리: `moai-html-report` 의 두 파일에 날짜 붙은 결정 인용이 남아 있었고 검사가 놓쳤어요. 검사가 `구절 (날짜·id)` 꼴과 복수형을 잡게 하고 예시 줄을 더한 뒤 두 줄의 인용을 지웠어요(커밋 29f1ddb0f3). 새 예시 7줄은 고치기 전 규칙에서 실패했고, 괄호 처리 줄이나 복수형 처리 줄을 지운 뮤턴트는 각각 exit 101 이었어요.
- AC-1 met: `scan_flags_each_kind_and_leaves_markers_alone` — 29f1ddb0f3 에서 통과.
- AC-2 met: `scan_shipped_files_outside_the_pending_list_are_clean` — 새 규칙만 있고 글을 고치기 전에는 두 파일로 실패, 29f1ddb0f3 에서 통과. 예외는 PENDING 의 원장 30줄(D-2).
- AC-3 met: `scan_exceptions_stay_named_and_needed` — 29f1ddb0f3 에서 통과.
- AC-4 met: `installer_document_names_the_skill_question_and_the_yes_option` — 29f1ddb0f3 에서 통과.
- AC-5 met: head f4369b3cba 에서 exit 0. 그 뒤에 바뀐 파일을 읽는 시험은 READY 의 영수증이 다시 골라 돌려요.
- AC-6 met: head b46308adbb 에서 exit 0(58 passed). 수리는 검사 표를 바꾸지 않았어요.
- 남은 것(감사가 막지 않는 지적으로 적은 것, 이 수리에서 고치지 않음): 날짜 없는 "by user decision" 두 곳(`skills/moai-html-report/SKILL.md`, `references/craft-fundamentals.md`), 3~5자리 밖의 번호와 `GH-123`·`이슈 123` 표기, PENDING 이 파일 단위인 점, `install/install.ps1` 의 `CLAUDE_CONFIG_DIR`.
- not_run: `pwsh` 와 Windows 설치기, Python 시험.

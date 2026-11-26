---
id: SPEC-20261122-secret-file-deny-rules
title: "Key files found in a project are offered as Read deny rules in the root project settings, with the .claudeignore entries carried over, only on the user's pick"
status: draft
revision: 3
created: 2026-11-22
parent: SPEC-20261121-flow-skeleton
---

# Key files found in a project are offered as Read deny rules in the root project settings, with the .claudeignore entries carried over, only on the user's pick

## Interview

- Q1: 설정 문서 조사 보고서의 항목들을 어떻게 다루나요? → A: 모든 항목을 카드로 만들어요(2026-11-22 사용자 선택, 이 작업의 카드 본문에 기록). 비밀 파일 거절 규칙은 그중 한 항목이고, 가드 파일이라 미리보기 뒤 사용자의 선택으로만 써요.
- Q2: 루트 설정 파일이 git 에 무시되면 어떻게 하나요? → A: 쓰기를 거부하고, 메시지는 고치는 법(`.claude/settings.json` 추적)만 안내해요(리더 결정 카드 노트 6132, 2026-11-22. 사용자의 직접 선택은 아니에요).
- Q3: 이 작업이 쓰는 설정은 어디까지인가요? → A: 사용자 설정 폴더는 쓰지 않고 프로젝트 설정과 템플릿만이에요(리더 지시, 2026-11-26).
- Q4: Decisions 의 D-1, D-2, D-4, D-10 은? → A: 모두 기본값 A 로 개정 3 이 승인됐어요(리더의 승인 기록, 카드 노트 10189, 2026-11-26).

## Intent

비밀 파일을 읽지 못하게 하는 `Read` 거절 규칙을, 프로젝트에서 실제로 찾은 파일과 `.claudeignore` 항목에서 만들어 미리보기로 보여 주고, 사용자가 고른 것만 저장소 루트의 프로젝트 설정에 더해요. 첫 조각은 SPEC 만 쓰고, 테스트는 Rust 만이에요.

2026-11-26 에 직접 받은 원문(`https://code.claude.com/docs/en/<이름>`):

| 문서 | 받은 시각 | sha256 | 읽은 절 |
| --- | --- | --- | --- |
| permissions.md | 07:55:33Z | 687a80f8ae55138a03648fbd54fa1c109291656a512a8ed28c1c3f2825460615 | Read and Edit(범위·경로 꼴·깊이·`!`·심볼릭 링크), Project allow rules and workspace trust 부터 What runs before you trust a folder 까지 |
| settings.md | 07:55:33Z | 7587a3251756cd151948718b70fff20fd26e3a4b545858d172d0941d57ab8f10 | Keep personal settings out of a repository, Where Claude Code keeps the local file, Settings precedence, Lists merge |
| settings-reference.md | 07:56:00Z | f87b28f014d323f1d99655f4dd8f329693e8295eec0f1cecf8af0bcc534a5249 | `permissions.deny` |
| worktrees.md | 07:56:00Z | 1b3eac588abbe62de80f6f6d4275c80de00d989883007cddcf9cdb773b373f20 | Copy gitignored files into worktrees, What worktrees share with the main checkout |
| large-codebases.md | 07:55:33Z | 672c530f0ba7e26782daf13f7e51d0f7aa065da7351cf341ec29f49e8c9438d6 | Block reads of generated and vendored code, 거절 규칙 두 번째 사본 문단 |
| common-workflows.md | 07:55:33Z | 9d44630362418bf644178eacf71479537b4646aba63d56f29940e02c7a47e2e6 | `@` 참조 파일의 훅 문단 |
| agent-sdk/secure-deployment.md | 07:55:33Z | 1d1ad468499aa0fe108ba85c10f0be7697958d9287cb77d67d439f2ca880fa02 | 마운트 전에 빼야 할 자격 증명 파일 표 |

문서의 사실:

1. "If your project has a `.claudeignore` file, it has no effect, so move its entries into `Read` deny rules" (permissions.md, Read and Edit).
2. `@` 로 참조한 파일은 `PreToolUse` 훅을 거치지 않고, `Read` 거절 규칙에 맞는 파일은 실리지 않아요(common-workflows.md).
3. 거절은 어느 범위에서 와도 허용을 이겨요. `deny`·`ask` 는 작업공간 신뢰를 기다리지 않아요. `--setting-sources user` 나 프로젝트 설정을 뺀 SDK 실행은 프로젝트 설정을 읽지 않아요(permissions.md).
4. 공유 프로젝트 설정은 세션의 주 작업 디렉터리에서 읽어요. 워크트리는 자기 체크아웃이라, 규칙이 워크트리에 닿으려면 루트의 설정 파일이 커밋돼 있어야 해요. 로컬 설정(`.claude/settings.local.json`)은 git 저장소에서는 메인 체크아웃 루트의 것을 써요(Windows 등 예외 있음, settings.md; large-codebases.md "Yourself only").
5. `Read` 거절은 같은 경로의 Edit·Write(새 파일 만들기 포함)도 막고 NotebookEdit 은 덮지 않아요. Bash 에서는 Claude Code 가 알아보는 파일 명령과 리디렉션 대상에만 닿고, 파일을 직접 여는 하위 프로세스에는 닿지 않아요(permissions.md).
6. `Read(<이름>)` 은 `**/<이름>` 과 같아 현재 폴더 아래 모든 깊이에 맞아요. 앞의 `/` 는 설정 출처(프로젝트 설정이면 주 작업 디렉터리)에 붙어요. 폴더 안 내용만 막는 문서의 꼴은 `./**/<폴더>/**/*` 예요(permissions.md, large-codebases.md).
7. `!` 로 시작하는 거절 항목은 같은 설정 파일에서 자기 앞의 `path`·`./path` 규칙을 파내요. 뒤의 규칙은 파내지 못해요(permissions.md, Special characters and `!` negation).
8. 자격 증명 파일의 모양: `.env`, `.env.local`, `.npmrc`, `.pypirc`, `*-service-account.json`, `*.pem`, `*.key` (secure-deployment.md 의 표. `~/` 경로 줄은 이 SPEC 밖이에요).

받은 팁(b3048a821a)의 코드에서 확인한 것:

- 프로젝트 설정 쓰기는 `claude-settings-plan`/`-apply`(`src/claude.rs`) 하나예요. `permissions` 는 통째로 하나의 키예요: 사용자가 이미 가진 `permissions` 는 충돌로 계획되고, 확인 토큰으로 키 전체를 바꿔야 해요. 그렇게 바꾸면 영수증이 그 키 전체를 moai 소유로 적고, 뒤의 요청이 `permissions` 를 빼고 `--keep-owned` 도 없으면 그 키의 삭제가 충돌 없이 계획돼요(`plan_settings`).
- 같은 계획의 답은 `settings_ignored`(`ignored`·`not_ignored`·`unknown`)와 `worktree_sessions` 를 이미 내요.
- `SPEC-20261122-settings-write-check` 개정 4 는 별도 쓰기 함수를 만들지 않고 이 쓰기 길 안에 관찰 검사를 넣어요. 아직 착지 전이에요.
- `src/project_scale.rs` 는 루트 설정의 `permissions.deny` 에 `Read(./<폴더>/**/*)` 줄을 제안하는 라이브러리이고 CLI 에 연결돼 있지 않아요.
- 템플릿과 `src` 는 `.claudeignore` 를 만들거나 읽지 않아요.
- 이 저장소의 `.gitignore` 는 `.claude/` 를 무시해요(357행 `/.claude/`, 398행 `.claude/`). `.claudeignore` 는 메인 체크아웃에만 있는 추적 안 된 파일이에요.

## Non-goals

- 운영체제 수준의 차단(샌드박스). `Read` 거절은 파일을 직접 여는 하위 프로세스에 닿지 않아요.
- NotebookEdit 을 막는 `Edit` 규칙.
- 사용자 범위·관리 설정, `~/`·`//` 경로 규칙.
- 훅 검사, 큰 저장소의 생성·벤더 폴더 규칙(`SPEC-20261122-project-scale-settings` 의 몫), 키 파일 진단.
- `.gitignore` 의 `.claude/` 무시를 좁히는 변경.
- `claude-settings-plan`/`-apply` 의 다른 키 동작 변경, 로컬 설정 파일에 거절 규칙 쓰기(D-2 의 B).
- 실제 Claude Code 를 돌려 규칙의 맞춤을 관찰하는 일.

## Requirements

- REQ-1: When the secret deny preview runs on a project, the planner shall list every `Read` deny rule it proposes with the files that rule matched and the settings file it would write, writing nothing. · failure: none — no settings file, receipt or backup changes.
- REQ-2: The planner shall build its key-file proposals from one list of credential file shapes (`.env`, `.env.*`, `*.pem`, `*.key`, `*-service-account.json`, `.npmrc`, `.pypirc`), proposing a shape only when a file in the project matches it. · failure: none — a shape that matched nothing proposes no rule.
- REQ-3: When the project root holds a `.claudeignore` file, the planner shall list each entry that is neither blank nor a comment as one `Read` deny rule in the converted form of D-5, to be picked one by one, with the warning that the rule also blocks editing, the present files it matches (tracked ones named) and a recommendation only where the entry matches a credential file shape. · failure: none — an entry with no matching file is still listed.
- REQ-4: The planner shall match files by name only, without opening, hashing or printing any matched file. · failure: none — a matched file the process cannot read is listed like any other.
- REQ-5: The planner's walk shall visit ignored files, skip `.git` and every folder below the root that holds its own `.git` entry, follow no symlink and stop at a fixed entry bound while saying that its file list is partial. · failure: none — the preview names the bound it stopped at and the folders it skipped as nested checkouts.
- REQ-6: The resulting `permissions.deny` list shall hold every key-file rule after the last `!` entry, with each `!` entry that preceded and could carve a key-file rule named in the preview before the pick. · failure: none — this holds for the list as written, entries that were already in the file included, on every apply.
- REQ-7: The apply shall add only the picked rules to `permissions.deny` of the root `.claude/settings.json`, keeping every other settings key and permission entry with its value, keeping the original file in the settings backup folder and taking no ownership of `permissions` that MoAI did not already hold. · failure: none — a stale preview digest or a file that is not a JSON object writes nothing.
- REQ-8: While git ignores the root `.claude/settings.json`, the apply shall write nothing, saying that the rules would not reach worktree sessions and naming the fix of tracking that file. · failure: none — a file git does not ignore is written and the answer says the rules reach worktrees once committed; an `unknown` answer from git is stated as unknown.
- REQ-9: Where the user makes the separate pick to retire `.claudeignore`, the apply shall move the file into the settings backup folder, never deleting it. · failure: none — without that pick the file stays and the preview says it has no effect either way.
- REQ-10: The preview shall propose only `Read(...)` rule forms and state that deny rules apply without the workspace trust step, headless runs included, except in a run that does not load project settings (`--setting-sources user`, or an SDK session without them). · failure: none — no `Write`, `Glob` or `Grep` path rule is ever proposed.
- REQ-11: The project-setup route shall offer the preview as one pick, while the update route only names it. · failure: none — neither route applies a rule without the user's pick.

## Acceptance

- AC-1 (REQ-1, REQ-2): the preview of a temporary project lists the rules for the key files present, with their matched files, and changes no file.
  verify: `cargo test --locked --offline --test secret_deny_cli preview_lists_rules_and_writes_nothing`
- AC-2 (REQ-3): a fixture holding this repository's `.claudeignore` entries (the table under Plan) lists each one as its converted rule with the edit warning and marks, recommending only the credential-shaped ones.
  verify: `cargo test --locked --offline --test secret_deny_cli claudeignore_entries_are_picked_one_by_one`
- AC-3 (REQ-4): a matched file with mode 000 and one holding a planted secret value are listed, and neither the value nor an open error appears.
  verify: `cargo test --locked --offline --test secret_deny_cli matched_file_is_never_opened`
- AC-4 (REQ-5): ignored files are found; a nested checkout, `.git` and a symlinked folder are not entered; past the bound the preview says the list is partial.
  verify: `cargo test --locked --offline --test secret_deny_cli the_walk_is_bounded_and_stays_in_the_project`
- AC-5 (REQ-6): with a user `!` entry already in the file and a carried-over negation applied in a second run, every key-file rule ends after the last `!` entry and the preview names the carve-out.
  verify: `cargo test --locked --offline --test secret_deny_cli key_rules_stay_after_every_negation`
- AC-6 (REQ-7): after an apply on a file with user `permissions` and other keys, every earlier value is kept, the original is in the backup folder, and a following `claude-settings-plan` without `permissions` plans no removal of it.
  verify: `cargo test --locked --offline --test secret_deny_cli apply_adds_rules_without_taking_ownership`
- AC-7 (REQ-8): with the root settings file ignored the apply writes nothing and names the fix; with an untracked file git does not ignore it writes and says to commit it.
  verify: `cargo test --locked --offline --test secret_deny_cli ignored_root_settings_refuses_the_write`
- AC-8 (REQ-9): `.claudeignore` stays without the second pick, and with it the file is in the backup folder with the same bytes.
  verify: `cargo test --locked --offline --test secret_deny_cli retiring_claudeignore_needs_the_second_pick`
- AC-9 (REQ-10): every proposed rule starts with `Read(` and the preview carries the trust, headless and setting-sources sentence.
  verify: `cargo test --locked --offline --test secret_deny_cli preview_states_rule_forms_and_trust_behavior`
- AC-10 (REQ-11): the project-setup route text names the preview as one pick and the update route text names it without an apply step.
  verify: `cargo test --locked --offline --test secret_deny_cli the_routes_offer_the_preview_and_apply_only_on_a_pick`
- AC-11: the READY set of the code slices passes.
  verify: `moai check affected --card <base>..<head>` — `summary.txt` 의 `result=PASS`

## Scope

- write (지금): `.moai/specs/SPEC-20261122-secret-file-deny-rules/**`
- write (코드 조각, 차례가 오면 `card scope` 로 넓혀요): `src/secret_deny.rs`(새 모듈), `tests/secret_deny_cli.rs`, `tests/fixtures/secret_deny/`(이 저장소 `.claudeignore` 사본과 기대 규칙), `src/ops/secret_deny_plan.rs`·`src/ops/secret_deny_apply.rs`(명령 `moai secret-deny-plan`·`moai secret-deny-apply`; `src/main.rs` 는 고치지 않아요)와 `docs/design-review/appendix/operations/` 의 두 명령 문서, `src/worktree_include.rs`(이름 맞춤 함수를 같이 써요), `src/claude.rs`(`SPEC-20261122-settings-write-check` 착지 뒤), `skills/moai/workflows/project.md`, `skills/moai/workflows/update.md`
- write (공유 파일은 줄 임대): `src/lib.rs`, `src/help.rs`, `tests/fixtures/help/*.txt`
- read-only: `.moai/specs/SPEC-20261122-settings-write-check/**`, `src/project_scale.rs`, `.gitignore`
- forbidden: 실제 사용자 설정이나 메인 체크아웃의 `.claudeignore` 를 테스트·명령이 쓰는 일, 맞은 키 파일을 열거나 해시하는 일, 실제 `claude` 실행, 기존 설정 쓰기의 거절을 약하게 하는 변경, Python

## Plan

1. 이 조각: 이 SPEC(개정 3). 커밋하고 SPEC-READY 를 보내요.
2. 코드 조각 A: 플래너와 `secret-deny-plan`, 빨간 테스트 먼저(REQ-1~5, REQ-10; AC-1~4, AC-9).
3. 코드 조각 B: `secret-deny-apply`, 순서 맞춤, 퇴역(REQ-6~9; AC-5~8). `src/claude.rs` 를 만지므로 `SPEC-20261122-settings-write-check` 착지 뒤에 해요. 지우기·덮어쓰기·옮기기를 하므로 위험 경로 확인은 임시 프로젝트의 Rust 통합 테스트로 해요.
4. 코드 조각 C: 두 라우트 글(REQ-11, AC-10). `skills/moai/**` 가 바뀌므로 착지 때 사용자 사본 맞추기 규칙이 적용돼요.

### 이 저장소의 `.claudeignore` 를 옮긴 계획

메인 체크아웃의 추적 안 된 파일을 2026-11-26 에 읽은 항목 34개예요. 비밀값은 없어요. `권장` 표시가 없는 줄은 기본으로 고르지 않아요.

| 항목 | 제안 규칙 | 권장 |
| --- | --- | --- |
| `.env` | `Read(.env)` | 권장 |
| `.env.local` | `Read(.env.local)` | 권장 |
| `.env.*.local` | `Read(.env.*.local)` | 권장 |
| `.env.production` | `Read(.env.production)` | 권장 |
| `.env.test` | `Read(.env.test)` | 권장 |
| `node_modules/` | `Read(**/node_modules/**/*)` | |
| `vendor/` | `Read(**/vendor/**/*)` | |
| `.venv/` | `Read(**/.venv/**/*)` | |
| `__pycache__/` | `Read(**/__pycache__/**/*)` | |
| `target/` | `Read(**/target/**/*)` | |
| `dist/` | `Read(**/dist/**/*)` | |
| `build/` | `Read(**/build/**/*)` | |
| `out/` | `Read(**/out/**/*)` | |
| `.next/` | `Read(**/.next/**/*)` | |
| `.git/` | `Read(**/.git/**/*)` | |
| `.vscode/` | `Read(**/.vscode/**/*)` | |
| `.idea/` | `Read(**/.idea/**/*)` | |
| `.cache/` | `Read(**/.cache/**/*)` | |
| `tmp/` | `Read(**/tmp/**/*)` | |
| `temp/` | `Read(**/temp/**/*)` | |
| `logs/` | `Read(**/logs/**/*)` | |
| `.moai/cache/` | `Read(.moai/cache/**/*)` | |
| `.moai/archive/` | `Read(.moai/archive/**/*)` | |
| `docs-site/public/` | `Read(docs-site/public/**/*)` | |
| `.claude/worktrees/` | `Read(.claude/worktrees/**/*)` | |
| `*.pyc` | `Read(*.pyc)` | |
| `.gitignore` | `Read(.gitignore)` | |
| `.gitattributes` | `Read(.gitattributes)` | |
| `*.swp` | `Read(*.swp)` | |
| `*.swo` | `Read(*.swo)` | |
| `*~` | `Read(*~)` | |
| `.DS_Store` | `Read(.DS_Store)` | |
| `*.tmp` | `Read(*.tmp)` | |
| `*.log` | `Read(*.log)` | |

## Risks

- 규칙이 실제 Claude Code 에서 어떻게 맞는지는 관찰하지 않았어요. 테스트는 쓰인 글자를 봐요. 폴더 꼴 변환(D-5)은 문서의 예에 기대요.
- 폴더·로그 항목을 고르면 읽기뿐 아니라 쓰기와 알아보는 Bash 파일 명령도 막혀요. 예를 들어 `target/`, `*.log` 를 고르면 빌드 기록을 읽거나 그 아래에 적는 흐름이 멈춰요. 그래서 자격 증명 모양만 권장해요.
- 모양 목록이 좁으면 비밀이 새고, `.env.*` 는 `.env.example` 같은 본보기 파일도 막아요. 미리보기가 맞은 파일을 보여 줘요.
- 키 파일 규칙을 `!` 항목 뒤로 두면 사용자가 일부러 만든 예외가 효력을 잃을 수 있어요. 고르기 전에 이름을 보여 줘요.
- 파일을 직접 여는 하위 프로세스와 NotebookEdit 은 막지 못해요.
- 걷기가 상한에서 멈추면 못 찾은 파일이 있을 수 있어요. 미리보기가 말해요.
- 이 저장소는 루트 설정이 무시돼서 지금은 적용할 수 없어요.
- Windows 는 not_run 이에요.

## Decisions

추천이 맨 앞이고 기본값이에요.

- D-1 (규칙을 쓰는 길): (A, 기본·권장) 기존 설정 쓰기(`src/claude.rs`) 안에 `permissions.deny` 에 항목만 더하는 길을 넣어요: 원본 백업, 계획 digest 확인, 원자 교체, 영수증의 파일 지문 갱신, 소유 표시는 그대로. (B) 지금 코드 그대로, 기존 `permissions` 에 새 규칙을 합친 값을 `claude-settings-apply` 의 확인 교체로 써요 — 새 쓰기 코드가 없지만 moai 가 키 전체를 소유하게 되어, 뒤의 요청이 그 키를 빼면 사용자의 규칙 삭제가 충돌 없이 계획돼요. (C) 설정 쓰기의 목록 병합 후속 일이 생길 때까지 적용 조각을 미뤄요.
- D-2 (규칙을 두는 곳): (A, 기본) git 에 추적되는 루트 `.claude/settings.json`(리더 결정 노트 6132; 팀과 워크트리에 같은 규칙). (B) 루트가 무시되는 저장소에는 메인 체크아웃의 `.claude/settings.local.json` 을 두 번째 선택으로 내요 — 문서상 워크트리에도 닿지만 개인 설정이고 Windows 는 달라요. 후속.
- D-3 (쓰기 검사): 별도 검사나 쓰기 함수를 만들지 않아요. D-1 의 길은 `SPEC-20261122-settings-write-check` 가 쓰기 길에 넣는 관찰 검사를 같이 지나가요(개정 2 의 "검사된 writer 로만" 을 지금 설계에 맞춰 바꿨어요).
- D-4 (`.claudeignore` 항목의 기본 선택): (A, 기본·권장) 자격 증명 모양에 맞는 항목만 권장하고 나머지는 고르지 않은 채로 내요. (B) 모두 권장 — 로그·빌드 폴더 읽기와 쓰기가 막혀요. (C) 아무것도 권장하지 않아요.
- D-5 (변환 꼴): `/` 로 끝나는 항목은 폴더 안 내용을 막는 문서의 꼴로 바꿔요: 중간에 `/` 가 없으면 `Read(**/<폴더>/**/*)`, 있으면 `Read(<경로>/**/*)`, 앞의 `/` 는 그대로 둬요. 다른 항목과 `!` 항목은 글자 그대로 `Read(<항목>)`·`Read(!<항목>)` 이에요. 개정 2 의 "끝 `/` 유지" 는 문서에 없는 꼴이라 바꿨어요.
- D-6 (순서): 키 파일 규칙은 목록의 마지막 `!` 항목 뒤로 옮겨 써요. 뒤로 옮기는 것은 보호를 늘리기만 해요. 다른 항목의 서로 순서는 그대로예요.
- D-7 (짝 맞춤): 이름으로만 해요(리뷰 노트 6130).
- D-8 (걷기 상한): `.git` 과 중첩 체크아웃을 건너뛰고 링크를 따라가지 않으며 항목 100,000개에서 멈춰요. `project_detect` 의 걷기(깊이 3, 5,000개)는 비밀 파일을 찾기에 얕아서 쓰지 않아요.
- D-9 (퇴역): 지우지 않고 백업 폴더로 옮겨요.
- D-10 (템플릿): (A, 기본) 템플릿은 거절 규칙을 싣지 않고, 규칙은 미리보기의 선택으로만 프로젝트에 생겨요(2026-11-22 사용자 결정 "프로젝트의 `.claude` 에는 기본으로 아무것도 두지 않아요"). (B) 새 프로젝트 템플릿에 자격 증명 규칙을 기본으로 실어요 — 그 사용자 결정을 바꾸는 일이라 사용자만 고를 수 있어요.
- D-11 (테스트): 임시 프로젝트와 임시 설정 폴더의 Rust 테스트만이에요.

## Result

조각 A(미리보기, `moai secret-deny-plan`)의 결과예요. 코드 head `b4b985747e`, base `54d2491a5b`. 조각 B(적용)와 C(라우트 글)는 아직이에요. 확인 명령은 2026-11-26 에 적힌 그대로 실행했어요.

- AC-1: met — 종료 0, `preview_lists_rules_and_writes_nothing` 1 통과.
- AC-2: met — 종료 0, `claudeignore_entries_are_picked_one_by_one` 1 통과. 픽스처는 메인 체크아웃의 `.claudeignore` 와 바이트까지 같아요(`cmp`).
- AC-3: met — 종료 0, `matched_file_is_never_opened` 1 통과.
- AC-4: met — 종료 0, `the_walk_is_bounded_and_stays_in_the_project` 1 통과.
- AC-5: not_run — 조각 B.
- AC-6: not_run — 조각 B.
- AC-7: not_run — 조각 B.
- AC-8: not_run — 조각 B.
- AC-9: met — 종료 0, `preview_states_rule_forms_and_trust_behavior` 1 통과.
- AC-10: not_run — 조각 C.
- AC-11: not_run — 조각 A 의 영수증은 이 커밋 뒤에 받아 카드의 READY 노트에 적어요.

SPEC 에 없던 안전장치 두 가지를 테스트와 함께 넣었어요: 링크로 바뀐 `.claudeignore` 는 읽지 않고(`a_linked_ignore_list_is_not_read`), 추적 파일을 물을 때 저장소가 설정한 `core.fsmonitor` 프로그램은 돌리지 않아요(`the_preview_runs_no_program_the_repository_configures`). 뮤턴트 9개는 모두 해당 테스트가 종료 101 로 잡았어요.

한계: 이름 맞춤은 대소문자를 구분해요(규칙이 대소문자를 어떻게 보는지는 문서에서 확인하지 못했어요). 맞은 파일은 줄이지 않고 모두 내므로 큰 저장소에서는 답이 커져요. 실제 Claude Code 의 맞춤, Windows, Linux 는 not_run 이에요.

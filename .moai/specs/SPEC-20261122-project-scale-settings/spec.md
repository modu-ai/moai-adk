---
id: SPEC-20261122-project-scale-settings
title: "The large-codebase setup writes only Claude Code's own files and keys, each from a documented line, and the settings writer checks the new values first"
status: draft
revision: 1
created: 2026-11-22
parent: SPEC-20261122-project-scale
---

# The large-codebase setup writes only Claude Code's own files and keys, each from a documented line, and the settings writer checks the new values first

## Intent

부모 SPEC(`SPEC-20261122-project-scale`)이 규모를 읽고 고르고 미리보기·적용하는 틀을 정해요. 이 SPEC 은 사용자가 `monorepo` 나 `large_tree` 를 골랐을 때 무엇을 쓰는지를 정해요: 패키지별 `CLAUDE.md`, 로컬 설정의 `claudeMdExcludes`, 루트 설정의 `Read` 거절 규칙, 언어 서버 플러그인의 `enabledPlugins`, `worktree.sparsePaths` 와 `symlinkDirectories`, 패키지 설정의 `additionalDirectories`, 패키지별 스킬. 모두 공식 문서 `large-codebases` 와 그 링크 페이지의 줄에서 와요(아래 Evidence). 미리보기의 줄마다 그 문서 줄이 붙어요.

지금의 쓰기 규칙(`src/claude.rs`)은 `worktree` 를 `baseRef` 하나로, `permissions` 를 `allow`·`deny`·`ask`·`defaultMode` 로 제한하고 `additionalDirectories` 를 거절해요. `claudeMdExcludes`, `enabledPlugins`, `sparsePaths`, `symlinkDirectories` 는 관리 키가 아니에요. 이 SPEC 이 그 값 검사를 더해요. 안전하게 쓰는 길(백업, 키별 병합, 다시 읽기, 원자적 바꿈)과 비밀 파일 거절 규칙 항목은 `SPEC-20261122-settings-write-check` 와 `SPEC-20261122-secret-file-deny-rules` 의 몫이고 이 SPEC 은 다시 만들지 않아요.

## Non-goals

- 언어 서버 바이너리와 플러그인을 설치하지 않아요. `enabledPlugins` 키와 설치 명령 글까지예요.
- 패키지 `CLAUDE.md` 의 규칙을 지어내지 않아요: 감지한 사실만 넣고 규칙은 디렉터리 주인이 써요.
- `plansDirectory` 를 정하거나 계획 파일을 만들지 않아요.
- 사용자 설정 폴더, 관리형 설정, 저장소 설정에 쓰지 않아요.

## Requirements

- REQ-1: When the scale is `monorepo`, the plan shall propose one `CLAUDE.md` for each package folder the user picks, holding a header line and the package's detected facts only · failure: none — no convention is invented and a package that already has one is a conflict kept as it is.
- REQ-2: The plan shall write no rule file without a `paths` field and no instruction file over 200 lines · failure: none — an unscoped rule loads in every session and the documentation targets under 200 lines.
- REQ-3: When the user picks packages they never work in, the plan shall list one `claudeMdExcludes` glob for each in `.claude/settings.local.json` · failure: none — the file is created only on this pick and a git ignore for it is asked as its own pick.
- REQ-4: When the user picks generated or vendored folders that the repository tracks, the plan shall list a `Read` deny rule for each in `permissions.deny` of the root `.claude/settings.json` · failure: none — folders that `.gitignore` already lists are not offered.
- REQ-5: When the user picks a code intelligence plugin for a detected language, the plan shall list its `enabledPlugins` key in the project settings and print the install command as text · failure: none — the plan installs nothing and names the language server binary the language needs.
- REQ-6: When the user picks sparse worktrees, the plan shall list `worktree.sparsePaths` holding `.claude` and the picked folders and `worktree.symlinkDirectories` for the dependency folders found at the root · failure: none — the plan also offers the one `.claude/worktrees/` line in `.git/info/exclude` (the common git dir's in a linked worktree) as its own pick.
- REQ-7: When the user starts sessions from a package folder, the plan shall list a self-contained `.claude/settings.json` in that package whose `permissions.additionalDirectories` name only sibling folders that exist inside the repository · failure: none — the root settings file carries the deny rules again because a worktree session reads the root file.
- REQ-8: When the user picks per-package skills, the plan shall list one `SKILL.md` per package with a `description` of at most 250 characters that begins with words a request would contain · failure: none — the body holds the package's detected facts only.
- REQ-9: The Claude settings writer shall accept `claudeMdExcludes` as an array of one-line glob strings, `enabledPlugins` as an object of `name@marketplace` keys with `true` values, `worktree.sparsePaths` and `worktree.symlinkDirectories` as arrays of repository-relative folder paths and `permissions.additionalDirectories` as relative paths without a drive or an absolute root · failure: none — each value is checked before any write.
- REQ-10 (REQ-9): IF a value holds `..` that leaves the repository, a control character, or any other sub-key than these, THEN the writer shall refuse it as it refuses an unknown key today.

## Acceptance

- AC-1 (REQ-1, REQ-3, REQ-4, REQ-5, REQ-6): 모노레포 고정물의 미리보기가 패키지별 `CLAUDE.md`, `claudeMdExcludes`, 루트 설정의 `Read` 거절 규칙, `.claude` 를 포함한 `sparsePaths`, 언어 플러그인을 모두 올리고 줄마다 문서 줄(URL 과 인용 한 줄)이 달려요 · verify: `cargo test --locked --offline --test project_scale_plan_lib a_monorepo_preview_lists_every_native_setting_with_its_doc_line`
- AC-2 (REQ-2, REQ-8): 계획 어디에도 `paths` 없는 규칙이 없고 200 줄을 넘는 지침이 없고 스킬 설명이 250 자 이하예요 · verify: `cargo test --locked --offline --test project_scale_plan_lib the_plan_never_writes_an_unscoped_rule_or_a_long_file`
- AC-3 (REQ-7, REQ-9, REQ-10): 패키지 설정이 자기 완결이고 형제 폴더만 가리켜요; `..` 로 저장소를 벗어나는 값, 제어 문자, 모르는 하위 키는 쓰기 전에 거절돼요 · verify: `cargo test --locked --offline --test claude_settings_scale_keys the_new_keys_are_checked_before_any_write`

## Scope

- write: `.moai/specs/SPEC-20261122-project-scale-settings/**`(이 조각), 뒤 조각에서 카드 범위에 더해요: `src/claude.rs` 의 값 검사(새 줄만), 새 모듈 `src/project_scale.rs` 의 설정·파일 행, 문서 인용 데이터 파일, 시험 파일
- read-only: `SPEC-20261122-project-scale`, `SPEC-20261122-settings-write-check`, `SPEC-20261122-secret-file-deny-rules`, `SPEC-20261122-worktree-place`
- forbidden: 사용자의 Claude 설정 폴더에 쓰기, 저장소 설정 쓰기, 플러그인 설치 실행, 카드 번호를 SPEC id·가지·커밋·PR 에 쓰기, 실제 HOME

## Plan

1. 조각 1(이번): 이 SPEC 과 아래 Evidence(문서 인용).
2. 조각 B: `src/claude.rs` 의 값 검사와 관리 키 목록(REQ-9, REQ-10, AC-3). 기존 줄은 바꾸지 않고 새 줄과 새 시험만 더해요. 그 파일은 여러 카드가 만지는 뜨거운 파일이에요.
3. 조각 C: 설정·파일 행의 계획과 미리보기 줄마다의 문서 줄(REQ-1 에서 REQ-8, AC-1, AC-2). 문서 인용 데이터 파일이 읽은 날짜와 Claude Code 버전을 가져요.
4. 조각 E(선택): 패키지별 스킬(REQ-8). 감지가 패키지 매니페스트의 바이트를 읽지 않아서 본문이 짧아요.

### Slice C2 first items

- 패키지 설정이 거절 규칙을 다시 실어요(REQ-7, D-2): `project_scale_plan_lib a_monorepo_preview_lists_every_native_setting_with_its_doc_line`
- 토큰 모양의 `enabledPlugins` 키와 관리 밖 키는 거절되고 메시지에 값이 나오지 않아요(REQ-9): `claude_settings_scale_keys a_token_shaped_plugin_key_is_refused_and_hidden`
- 설정 계획 자체가 저장소를 벗어나는 `additionalDirectories` 를 거절하고 `.git` 찾기는 저장소 루트에서 멈춰요(REQ-10): `claude_settings_scale_keys the_writer_refuses_an_additional_directory_outside_the_repository`

### Slice C3 follow-ups (C2 감사의 막지 않는 두 항목)

- 거절 메시지는 제어 문자뿐 아니라 보이지 않는 서식·구분 문자(폭 없는 공백, 소프트 하이픈, 방향 표시·뒤집기, 줄·문단 구분자, BOM, 태그 문자)도 지운 뒤 자격 증명 검사를 하고, 지운 글만 보여 줘요(REQ-9): `claude_settings_scale_keys a_token_split_by_an_invisible_format_or_separator_character_is_never_echoed`
- 고른 플러그인 중 이미 같은 값인 것은 그대로 두고 나머지를 더해요. 다른 값으로 이미 있는 플러그인만 충돌이고, 충돌 글은 그 플러그인만 이름 대요(REQ-5): `project_scale_plan_lib a_picked_plugin_already_set_the_same_way_is_kept_and_the_rest_are_added`

### Slice C4 follow-ups (C3 감사의 막지 않는 다섯 항목)

- N1: 모르는 `permissions` 하위 키의 거절 글이 키를 `shown_value` 로 보여 줘요. 보이지 않는 문자로 쪼갠 토큰 모양 키는 통째로 나오지 않아요. 같은 검사 묶음(`check_env`, `check_worktree`, `check_scale_list`, `check_claude_md_excludes`, `check_enabled_plugins`, `check_model`, `refuse_unlisted`)에서 사용자 키나 값을 그대로 찍는 다른 거절 글은 없었어요(REQ-9): `claude_settings_scale_keys an_unknown_permissions_key_shaped_like_a_split_token_is_never_echoed`, `claude::tests::an_unknown_permissions_key_is_shown_without_a_split_token`
- N2: 값을 받아들이는 검사도 보이지 않는 문자를 지운 글을 함께 봐요(D-6). U+200B 로 쪼갠 토큰은 `claudeMdExcludes` 항목, `model`, `worktree` 목록, `permissions` 규칙과 `additionalDirectories`, `outputStyle`, `statusLine`, `env` 값으로 거절되고 아무것도 계획되지 않아요(REQ-9, REQ-10): `claude_settings_scale_keys a_value_holding_a_token_split_by_an_invisible_character_is_refused_and_nothing_is_planned`
- N3: 감사가 든 거의 보이지 않는 문자를 갈래마다 정했어요(D-7). 지우는 갈래마다 한 점씩, 기존 19 갈래와 제어 문자, 새 다섯 갈래를 봐요(REQ-9): `claude_settings_scale_keys a_token_split_by_any_invisible_class_is_never_echoed_or_accepted`
- N4: `credential_like` 가 맞추는 모양마다 쪼갠 토큰 한 점씩, 모두 36 점이에요: 개인 키 머리 1, 토큰 접두어 19(`secret_kind` 의 접두어 표), JWT 1, `Authorization` 의 `Bearer`·`Basic` 2, 비밀번호가 든 URL 1, `credential_like` 자신의 `://`·`@` 규칙 1, 대입 키 11(REQ-9): `claude_settings_scale_keys every_credential_pattern_split_by_an_invisible_character_is_refused_and_hidden`
- N5: 보이지 않는 문자를 지운 거절 글은 `(invisible characters removed)` 를 붙여서, 지운 이름이 허용된 이름처럼 읽히지 않아요(D-8, 보여 주기만): `claude_settings_scale_keys a_refusal_that_removed_invisible_characters_says_so`

## Risks

- 거절 규칙은 보안 경계가 아니에요(문서: 하위 프로세스가 직접 여는 파일은 안 막아요). 미리보기가 그 한계를 한 줄로 말해요.
- `permissions.additionalDirectories` 는 사용자가 작업 공간 신뢰를 수락해야 효력이 있어요(문서). 미리보기가 말해요.
- `enabledPlugins` 는 설치하지 않아요. 각자 설치 명령을 돌려야 로드돼요(문서). 그 명령은 마켓플레이스(GitHub) 접속이 필요하고 사용자가 고른 뒤에만 이 세션이 돌려요.
- `.claude/settings.local.json` 은 Claude Code 가 직접 쓸 때만 전역 git 제외 목록에 올라가요(문서). 손으로 만들면 무시되지 않아서, `.gitignore` 한 줄을 별도 선택으로 물어요.
- `sparsePaths` 가 켜지면 git 이 공유 `.git/config` 에 `extensions.worktreeConfig` 를 켜요(문서). 미리보기가 말해요.
- 문서 줄은 버전마다 달라질 수 있어서 인용 데이터 파일에 읽은 날짜와 버전을 둬요.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | `paths:` 없는 규칙과 200 줄 넘는 지침은 쓰지 않아요. | 문서(규칙 로드 시점, 200 줄 목표)와 슬라이스 0 측정. | 카드 열림 |
| D-2 | 거절 규칙은 `./**/<폴더>/**/*` 꼴이고 루트 설정에 써요(워크트리 세션이 받도록). 패키지에서 시작하기를 고르면 REQ-7 그대로 그 패키지의 `.claude/settings.json` 도 자기 완결이어서, 그 패키지 안의 고른 폴더 규칙을 패키지 기준 경로로 다시 실어요(리더 결정 2026-11-22, 되돌릴 수 있어요; 예전 기록 "거절 규칙은 루트 파일에만" 을 바꿔요). | 문서(`Put it together`, 거절 규칙 두 번째 사본; 패키지 설정은 부모에서 상속되지 않아요). | 카드 열림 |
| D-3 | 플러그인은 `enabledPlugins` 키와 설치 명령 글까지만 다뤄요. 설치는 사용자가 고른 뒤 이 세션이 그 명령을 돌려요. | 문서: 프로젝트 설정의 키는 설치하지 않아요. 설치는 GitHub 접속이 필요해요. | 카드 열림 |
| D-4 | 계획 파일은 건드리지 않아요. | 문서: 계획은 압축 뒤 디스크에서 다시 들어와요. | 카드 열림 |
| D-5 | 쓰는 파일의 자리는 문서의 자리 그대로예요: 패키지 `CLAUDE.md` 는 패키지 폴더, 제외·개인 항목은 `settings.local.json`, 팀 항목은 `settings.json`. | `large-codebases` 각 절의 "committed or local" 안내. | 카드 열림 |
| D-6 | 받아들이는 검사는 보이지 않는 문자가 든 값을 통째로 거절하지 않고, `credential_like` 가 글을 그대로 한 번, 보이지 않는 문자를 지우고 한 번 봐요. 쪼갠 토큰만 거절돼요. 이 함수를 쓰는 모든 곳(`scale_value_ok`, `holds_credential`, `env_entry_refusal`, `shown_value`, `project_scale` 의 경로 검사, 설정 병합)이 같이 바뀌어요. | 진짜 값에 보이지 않는 문자가 쓰여요: 이모지 연결(U+200D), 변형 선택자(U+FE0F), 깃발 태그 문자는 폴더 이름과 명령에 나올 수 있어요. 통째로 거절하면 그런 값이 막혀요. | `credential_like` 의 두 번째 판단을 빼면 예전대로예요. |
| D-7 | 거의 보이지 않는 문자는 갈래마다 정해요. 넣어요(`invisible_char`): 점자 빈칸 U+2800, 할당 안 된 특수 U+FFF0–FFF8, 음표 빈 머리 U+1D159, 거란 소자 채움 U+16FE4, 이집트 상형 서식 제어 U+13430–1343F — 빈칸이나 아무것도 아닌 것으로 찍히고 설정 값(경로, glob, 플러그인 이름, 모델 이름, 명령)에 쓸 일이 없어요. 넣지 않아요: 머리카락 공백 U+200A 를 포함한 공백 문자(U+00A0, U+2000–200A, U+202F, U+205F, U+3000) — 토큰을 쪼개면 틈이 눈에 보여 읽는 사람이 두 낱말로 보고, 지우면 진짜 규칙(`Bash(git status)`)의 낱말이 붙어서 보여요. 결합 문자(U+0301 같은) — 앞 글자에 붙어 찍혀서 숨기지 않고, 분해형(NFD) 파일 이름에 쓰여요. 닮은 글자(키릴 `а`, 전각 글자) — 숨기는 문자가 아니고 다른 글자예요. 바꿔 맞추려면 혼동 문자 표가 필요해서 남겨요. | C3 감사 N3. 결합 문자나 닮은 글자로 바꾼 토큰은 검사에 걸리지 않아요. 그 값은 진짜 토큰과 바이트가 달라요. 남은 위험으로 적어요. | 새 갈래를 `invisible_char` 에서 빼면 예전대로예요. |
| D-8 | 보이지 않는 문자를 지운 값은 거절 글에서 `(invisible characters removed)` 를 붙여 보여 줘요. 보여 주기만 바뀌어요. | C3 감사 N5: `model` 뒤에 U+200B 를 붙인 키가 거절 글에서 허용된 `model` 로 읽혔어요. | `shown_value` 의 덧붙임을 빼면 예전대로예요. |

### Evidence

문서와 블로그는 2026-11-22 에 읽었어요(원문 인용, 읽은 시각은 UTC 06:32 에서 06:33). 조각 C 의 인용 데이터 파일이 이 표를 줄 단위 출처로 써요.

| 설정 | 인용 | 출처 |
| --- | --- | --- |
| 하위 `CLAUDE.md` 로드 | "Instead of loading them at launch, Claude Code loads each one once Claude reads, writes, or edits another file in that subdirectory." | `memory.md`, How CLAUDE.md files load |
| `paths:` 규칙 | "A path-scoped rule loads when Claude uses the Read, Write, or Edit tool on a matching file." / "Rules without `paths` frontmatter are loaded at launch with the same priority as `.claude/CLAUDE.md`." | `memory.md`, Path-specific rules, Set up rules |
| 지침 크기 | "target under 200 lines per CLAUDE.md file" / on-demand 파일은 40,000 자(200K 창)까지만 받아요 | `memory.md`, Write effective instructions, Claude ignores the end of my subdirectory CLAUDE.md |
| `claudeMdExcludes` | "You can configure `claudeMdExcludes` at any settings layer: user, project, local, or managed policy. Arrays merge across layers." / "Patterns are matched against absolute file paths using glob syntax." | `memory.md`, Exclude specific CLAUDE.md files |
| 로컬 설정의 git 무시 | "Claude Code keeps it out of git when it creates the file; if you create it by hand, add it to `.gitignore` yourself" | `settings.md`, Settings files and who they affect |
| `Read` 거절 규칙 | "add `Read` deny rules in `permissions.deny` to block Claude from opening those files." / 거절 규칙은 하위 프로세스가 직접 여는 파일은 막지 않아요 | `large-codebases.md`, Block reads of generated and vendored code; `permissions.md`, What Read and Edit rules cover |
| 설정은 부모에서 상속되지 않음 | "Project settings in `.claude/settings.json` aren't inherited from parent directories the way CLAUDE.md files are." | `large-codebases.md`, Choose where to start Claude |
| 워크트리 안의 설정 | "Project settings inside the worktree therefore load from the worktree root's `.claude/settings.json`" | `large-codebases.md`, Check out only the directories you need |
| `sparsePaths` | "Include `.claude` in the list so that each worktree also has the repository's `.claude/settings.json` and `.claude/rules/`." | `settings-reference.md`, `worktree.sparsePaths` |
| `symlinkDirectories` | "Symlink directories from the main repository into each worktree so you don't duplicate large directories on disk." | `settings-reference.md`, `worktree.symlinkDirectories` |
| 워크트리 자리와 무시 | "By default, the worktree is created under `.claude/worktrees/<name>/` ... Add `.claude/worktrees/` to your `.gitignore`" | `worktrees.md`, Start Claude in a worktree |
| `additionalDirectories` | "Directories listed in `permissions.additionalDirectories` in a settings file grant file access only and don't load any of the configuration below." | `permissions.md`, Additional directories grant file access, not configuration |
| 언어 서버 플러그인 | 13 개 언어의 플러그인과 바이너리 표(예: `typescript-lsp` 와 `typescript-language-server`, `rust-analyzer-lsp` 와 `rust-analyzer`); "It doesn't include the language server." | `plugins/code-intelligence.md`, Install a code intelligence plugin |
| `enabledPlugins` | "Project settings: plugins turned on for everyone in the repository. This doesn't install them, so each user installs them." | `settings-reference.md`, `enabledPlugins` |
| 스킬 위치와 설명 한도 | "Skills in a `.claude/skills/` directory below where you started don't load at startup. They load the first time Claude reads or edits a file in that subdirectory" / 설명과 `when_to_use` 합쳐 1,536 자가 한도 | `skills.md`, Load skills in monorepos and subdirectories, Skill descriptions are cut short |
| 계획 파일 | "The plan Claude wrote in plan mode / Re-injected from disk" | `context-window.md`, What survives compaction |
| 블로그 | "The root file should be pointers and critical gotchas only; everything else drifts into noise." / "Initializing in subdirectories, not at the repo root." / 숫자 기준은 없어요 | 블로그 "How Claude Code works in large codebases", Making the codebase navigable at scale |

## Result

(status draft — written when the slice lands)

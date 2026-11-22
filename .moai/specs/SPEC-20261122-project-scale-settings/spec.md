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
- REQ-6: When the user picks sparse worktrees, the plan shall list `worktree.sparsePaths` holding `.claude` and the picked folders and `worktree.symlinkDirectories` for the dependency folders found at the root · failure: none — the plan also offers the one `.gitignore` line for `.claude/worktrees/` as its own pick.
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
| D-2 | 거절 규칙은 `./**/<폴더>/**/*` 꼴이고 루트 설정에 써요(패키지 설정에서 시작한 세션만 받는 것을 막고 워크트리 세션이 받도록). | 문서(`Put it together`, 거절 규칙 두 번째 사본). | 카드 열림 |
| D-3 | 플러그인은 `enabledPlugins` 키와 설치 명령 글까지만 다뤄요. 설치는 사용자가 고른 뒤 이 세션이 그 명령을 돌려요. | 문서: 프로젝트 설정의 키는 설치하지 않아요. 설치는 GitHub 접속이 필요해요. | 카드 열림 |
| D-4 | 계획 파일은 건드리지 않아요. | 문서: 계획은 압축 뒤 디스크에서 다시 들어와요. | 카드 열림 |
| D-5 | 쓰는 파일의 자리는 문서의 자리 그대로예요: 패키지 `CLAUDE.md` 는 패키지 폴더, 제외·개인 항목은 `settings.local.json`, 팀 항목은 `settings.json`. | `large-codebases` 각 절의 "committed or local" 안내. | 카드 열림 |

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

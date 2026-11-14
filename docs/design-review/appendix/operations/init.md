# `init`

operation `project.init` · 코드 `src/ops/init.rs`, `src/launcher_init.rs`(`init_here`·`init_new`), `src/home_workdir.rs`(`project_dir`) · Issue #2835

`init [--new NAME] [--host claude|codex|glm] [--dry-run]` 는 프로젝트를 초기화한다.

- `--new` 없이: 지금 폴더를 런처 첫 실행(#2481)과 똑같이 초기화한다 — 필요하면 `git init`, 표식이 붙은 프로젝트 설정, `AGENTS.md`. 이미 있는 파일은 건드리지 않는다. HOME 자체와 파일 시스템 루트는 거부한다.
- `--new NAME`: `<HOME>/.moai/projects/NAME/` 을 만들고(HOME 은 `$HOME`, Windows 는 `%USERPROFILE%`, #2811 D1) 그 안을 같은 경로로 초기화한다. `projects/` 는 처음 쓸 때 만든다. 쓰기 전에 거부하는 것: 안전한 경로 한 칸이 아닌 이름(`[A-Za-z0-9._-]`, `-`·`.` 로 시작·끝 금지 — 고쳐 쓰지 않고 거부), `~/.moai` 나 `~/.moai/projects` 가 심볼릭 링크(그 대상에는 쓰지 않음), 이미 있는 이름(덮어쓰지 않음, 만들기는 배타적).
- 결과: `project`(폴더), `created`(이번 실행이 만들었는지), `note`(초기화 한 줄), `git_pending`, `writes`. `--dry-run` 은 아무것도 쓰지 않고 `would …` 한 줄을 돌려준다.
- 기존 프로젝트는 옮기지 않는다(#2835 D2).

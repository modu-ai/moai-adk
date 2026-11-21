# `git-hook-apply`

operation `git-hook-apply` · 코드 `src/ops/git_hook_apply.rs`, `src/git_hooks.rs` · SPEC-20261120-commit-message-card-numbers (수리 조각)

`git-hook-apply --dir <프로젝트>`(기본 `.`) 는 이 저장소가 함께 실어 나르는 commit-msg 훅(`install/commit-msg` 와 같은 바이트, 컴파일에 포함)을 프로젝트의 git 훅 폴더 — `git rev-parse --git-path hooks` 가 이름하는 폴더라 `core.hooksPath` 설정도 따른다 — 에 `commit-msg` 이름으로 쓴다(모드 0755). 훅이 커밋 메시지의 카드 번호 꼴(`C-<n>`, `c-<n>`)을 기계로 거절하게 하는 설치 명령이다: `moai project-update` 의 git-hook 스텝이 이 명령을 apply 로 이름하고, 사람이 수동으로 설치할 때도 같은 파일이 `install/commit-msg` 로 함께 실려 나른다.

판정은 먼저 읽는다: 훅 폴더에 `commit-msg` 가 없으면 `install`, 같은 바이트가 이미 있으면 `unchanged`, 다른 훅 파일(사용자 자신의 훅)이 있으면 `needs_decision` — **사용자의 훅은 절대 바뀌지 않는다**. apply 는 판정이 `install` 일 때만 쓰고, plan 과 apply 사이에 훅이 나타나면(다른 세션·사람이) 거부해 덮지 않는다. git 은 한 번만 돌고(`rev-parse --git-path hooks`, scrubbed 환경) 나머지 읽기·쓰기는 로컬 파일뿐이다. 저장소가 아니면 `git-hook apply refused` 와 exit 1, 사용법 오류는 exit 2.

훅 스크립트의 경계: 설치된 moai 가 없거나 구버전이라 `--message-file` 을 모르는 usage 오류(exit 2)일 때는 커밋을 막지 않고 경고 한 줄과 함께 통과한다 — 제거 환경·구버전 설치가 커밋을 막는 사고는 이 검사의 목적 밖이다(스캔이 카드 번호를 찾은 exit 1 만이 커밋을 거절한다).

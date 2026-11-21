# `git-hooks-plan`

operation `git_hooks.plan` · 코드 `src/ops/git_hooks_plan.rs`, `src/git_hooks/mod.rs` · SPEC-20261121-go-path-removal

`git-hooks-plan [PROJECT]` 은 저장소의 MoAI 관리 git 훅(pre-commit·pre-push)을 바꾸면 어떻게 되는지 보여 준다. 쓰지 않는다. Go 세대 증거가 있는 훅은 `go_hooks` 로 은퇴 대상 — 증거는 둘 중 하나: 곁의 해시 파일 `.moai-<이름>.sha256` 이 훅의 sha256 과 일치하거나, 첫 줄 안에 정확한 Go 머리말과 본문의 Go 명령줄이 함께 있는 경우. 그 외의 훅은 전부 사용자의 것으로 `kept` 에 사유와 함께 나열되고 바이트 그대로 남는다(편집된 Go 훅도 사용자 것). 바뀔 때는 비어 있는 관리 이름에 Rust 훅 텍스트가 설치되고 `confirm_token` 이 함께 나온다. 워크트리의 `.git` 파일도 읽어 실제 git dir 의 hooks 폴더를 본다.

# `git-hooks-plan`

operation `git_hooks.plan` · 코드 `src/ops/git_hooks_plan.rs`, `src/git_hooks/mod.rs` · SPEC-20261121-go-path-removal

`git-hooks-plan [PROJECT]` 은 저장소의 MoAI 관리 git 훅(pre-commit·pre-push·reference-transaction)을 바꾸면 어떻게 되는지 보여 준다. 쓰지 않는다. 세 번째 훅 `reference-transaction` 은 통합 브랜치 보호 훅이라(SPEC-20261125-integration-guard-install) 설치 텍스트가 `install/git-hooks/reference-transaction` 이고 `guard::hook_text()` 와 같은 바이트다. 이 훅만은 프로필이 통합 브랜치를 이름 대거나 보호 기록(`moai/integration-branch`)이 이미 있는 저장소에서만 계획에 오르고, 그런 저장소에서는 지울 Go 훅이 없어도 빈 이름에 혼자 설치된다. 그 밖의 저장소는 이 이름을 건드리지 않고 계획이 예전과 같다. 사용자가 둔 훅이나 설치 뒤 고친 훅은 `kept` 로 남고 그 이름만 막는다. 설치되지 않은 저장소에서 세션을 열면 세션 시작이 한 줄(`MoAI: integration branch <이름> is not guarded: <빠진 것> (moai git-hooks-plan)`)로 알리고, 그 줄이 이 명령을 가리킨다. Go 세대 증거가 있는 훅은 `go_hooks` 로 은퇴 대상 — 증거는 둘 중 하나: 곁의 해시 파일 `.moai-<이름>.sha256` 이 훅의 sha256 과 일치하거나, 첫 줄 안에 정확한 Go 머리말과 본문의 Go 명령줄이 함께 있는 경우. 그 외의 훅은 전부 사용자의 것으로 `kept` 에 사유와 함께 나열되고 바이트 그대로 남는다(편집된 Go 훅도 사용자 것). 바뀔 때는 비어 있는 관리 이름에 Rust 훅 텍스트가 설치되고 `confirm_token` 이 함께 나온다. 워크트리의 `.git` 파일도 읽어 실제 git dir 의 hooks 폴더를 본다.

## `moai update` 의 가드 훅 제안

`moai update` 는 보호 기록이 있거나 프로필이 통합 브랜치를 이름 대는 저장소에서 같은 계획으로 훅 층을 보이고, 빠진 `reference-transaction` 훅을 `git_hooks.install` 에 올려 `--confirm-git-hooks` 로 설치한다(SPEC-20261126-guard-record-repoint). 단 기록이 프로필의 통합 브랜치와 달라 `stale` 이면 update 는 이 훅을 withheld(보류)한다. 보류는 주 워크트리와 링크 워크트리 모두에서 같고(링크 워크트리는 그 체크아웃의 프로필과 비교한다), 안내는 주 워크트리에서만 낸다. 낡은 기록에 훅을 깔면 방금 놓으려는 옛 브랜치를 잠그기 때문이다. 이 명령(`git-hooks-plan`)은 그대로 훅을 계획에 올린다. 그 대신 update 미리보기는 기록, 프로필 브랜치, `moai card guard-repoint` 를 이름 대는 읽기 전용 안내 한 줄을 달고(어떤 digest 에도 들어가지 않고, 링크 워크트리이거나 읽지 못하면 없다), 기록을 맞추면 영수증이 공통 git 폴더의 `moai/record-repoint/` 에 남는다.

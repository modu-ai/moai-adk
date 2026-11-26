# `git-hooks-apply`

operation `git_hooks.apply` · 코드 `src/ops/git_hooks_apply.rs`, `src/git_hooks/mod.rs` · SPEC-20261121-go-path-removal

`git-hooks-apply --confirm-git-hooks TOKEN [PROJECT]` 은 확인한 `git-hooks-plan` 그대로 바꾼다. 순서: 모든 조건을 첫 쓰기 전에 검사하고, 은퇴 분과 해시 파일을 `.git/hooks/moai-retired-<stamp>/` 에 원래 모드 그대로 바이트 검증 보관한 뒤 원본을 치우고, 비어 있는 관리 이름(pre-commit·pre-push, 그리고 보호가 걸릴 수 있는 저장소에서는 reference-transaction)에 설치 텍스트(실행 파일 안에 품은 바이트)를 0755 로 새로 쓴다. 훅을 쓴 뒤 보호 기록 `moai/integration-branch`(공통 git 폴더)가 없으면 프로필의 통합 브랜치로 한 번 써 넣고, 답의 `guard.record` 가 상태를 말한다: `written`(이번 호출이 썼다), `present`(이미 있어 다시 쓰지 않았다), `not_written`(주 워크트리가 아니거나 프로필에 통합 브랜치가 없어 못 썼고, `guard.why` 가 까닭을 적는다). 기록은 훅 다음에 쓴다. 사용자 훅은 건드리지 않고 `kept` 로 나열하며, 미리보기 뒤 새로 생긴 이름이 있으면 아무것도 쓰지 않는다. 토큰이 미리보기와 다르면 아무것도 쓰지 않는다.

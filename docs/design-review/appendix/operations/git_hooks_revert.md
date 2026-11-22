# `git-hooks-revert`

operation `git_hooks.revert` · 코드 `src/ops/git_hooks_revert.rs`, `src/git_hooks/mod.rs` · SPEC-20261121-go-path-removal

`git-hooks-revert STAMP [PROJECT]` 는 적용 때 보관한 `moai-retired-<stamp>/` 를 원래대로 돌려놓는다. 현재 훅이 이 기능이 설치한 바이트와 같거나 없을 때만 되돌리고, 보관함의 알 수 없는 파일은 거절한다. 복원 후 보관 폴더는 비어 사라진다.

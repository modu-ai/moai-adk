# `worktree-install-plan`

operation `worktree.install.plan` · 코드 `src/ops/worktree_install_plan.rs` · #3010 P6a

`worktree-install-plan PROJECT --scope NAME` 은 프로필 `scopes` 에 적힌 scope 하나가 고른 폴더만 설치하는 명령 한 줄을 **보여 주기만** 한다. 명령을 실행하지 않고 아무것도 쓰지 않는다(패키지 설치는 네트워크로 받는 일이라 외부 전송이다). 프로젝트 루트의 잠금 파일(`package-lock.json`·`pnpm-lock.yaml`·`yarn.lock`·`bun.lock`/`bun.lockb`·`Cargo.lock`·`go.sum`)로 매니저를 고르고, 잠금 파일이 없거나 둘 이상이면 `undecidable` 이다. yarn 1 과 go 는 부분 설치 방법이 없어 `full_only`(전체 설치)로 말한다. 매니저 문서에 없는 것은 `unverified` 에 적는다.

# `worktree-install-plan`

operation `worktree.install.plan` · 코드 `src/ops/worktree_install_plan.rs` · #3010 P6a

`worktree-install-plan PROJECT --scope NAME` 은 프로필 `scopes` 에 적힌 scope 하나가 고른 폴더만 설치하는 명령 한 줄을 **보여 주기만** 한다. 명령을 실행하지 않고 아무것도 쓰지 않는다(패키지 설치는 네트워크로 받는 일이라 외부 전송이다). 프로젝트 루트의 잠금 파일(`package-lock.json`·`pnpm-lock.yaml`·`yarn.lock`·`bun.lock`/`bun.lockb`·`Cargo.lock`·`go.sum`)로 매니저를 고르고, 잠금 파일이 없거나 둘 이상이면 `undecidable` 이다. yarn 1 과 go 는 부분 설치 방법이 없어 `full_only`(전체 설치)로 말한다. 매니저 문서에 없는 것은 `unverified` 에 적는다.

폴더를 고르는 모양은 매니저마다 다르다. npm 은 `-w <폴더>` 로 폴더를 그대로 받고, pnpm 과 Bun 은 `--filter './<폴더>'` 처럼 `./` 로 시작해야 폴더 선택자이며 `./` 가 없으면 패키지 이름 선택자로 읽는다(pnpm.io/filtering, bun.sh/docs/pm/filter, 2026-11-15 읽음, #3083 감사 B1). yarn 2+ 와 cargo 는 폴더가 아니라 `package.json`/`Cargo.toml` 의 이름으로 고르므로 이름이 없는 폴더가 있으면 `needs_package_names` 이다. 어떤 매니저도 실행해 보지 않았으므로 명령이 고르는 폴더가 정말 그 폴더뿐인지는 `unverified` 에 남는다. scope 폴더의 중간 경로가 링크이거나 `..` 를 담으면 `refused` 이다. 모든 답(거절 포함)에 `ran: false` 와 "preview only" 문구가 들어 있다.

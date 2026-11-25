# `build-number`

operation `build.number` · 코드 `src/ops/build_number.rs`, `src/build_number.rs` · SPEC-20261125-install-build-number

설치할 빌드의 번호(`MOAI_LOCAL_BUILD`)를 사람이 정하지 않고 moai 가 내줘요. `next` 는 다음 번호를 기록한 뒤 출력하고, `show` 는 마지막 번호와 다음 번호를 읽기만 해요.

`build-number next|show [--project DIR] [--home DIR]`

- 기록은 HOME 의 프로젝트 상태 폴더의 `build-number.log` 예요(프로젝트 등록이 만든 폴더, 이 명령은 폴더를 만들지 않아요). 한 줄이 번호 하나예요: `<번호>`, UTC 시각, 세션 id(없으면 `-`)를 탭으로 이어요. 덧붙이기만 하고 0600 으로 만들어요.
- `next` 의 번호는 기록의 마지막 번호와, 같은 head 의 설치 폴더(`versions/<head>.<yearweek>.<n>[-dev]`) 가운데 가장 큰 번호 가운데 큰 쪽보다 1 커요. 그래서 처음 부르면 지금 설치본 위에서 이어져요. 다른 head 의 폴더와 모양이 다른 이름은 무시해요.
- 배타 잠금 아래에서 기록을 읽고 줄을 쓰고 `fsync` 한 뒤에만 번호를 출력해요. 겹쳐 부른 둘은 다른 번호를 받아요. 잠금은 10초까지 기다린 뒤 거부해요. 번호는 쓰였든 아니든 다시 내주지 않아요.
- 답은 `build`, `env`(`MOAI_LOCAL_BUILD=<번호>`), `version`(`{head}.{yearweek}.{번호}-dev`), `recorded`, `previous`, `installed_floor` 예요. `show` 는 `last`(없으면 null), `next`, `installed_floor`, `recorded` 예요.
- 기록이 번호 줄이 아니거나 번호가 올라가지 않거나 끝 줄에 줄바꿈이 없으면 파일 이름을 대며 거부하고 기록을 그대로 둬요. `HEADVER` 파일이 없거나 프로젝트 상태 폴더가 없거나 링크이면 거부하고 아무것도 쓰지 않아요.
- 설치 자체는 하지 않아요. 검사만 하는 빌드는 번호를 쓰지 않아요: `check affected` 는 자식 프로세스에서 `MOAI_LOCAL_BUILD` 와 `MOAI_BUILD_NUMBER` 를 지워요.

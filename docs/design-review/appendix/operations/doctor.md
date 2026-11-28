# `doctor`

operation `doctor` · 코드 `src/ops/doctor.rs`, `src/moai_doctor.rs` · SPEC `SPEC-20261127-doctor-moai-layer`

`doctor [--project PATH] [--auth-status]` 는 Claude Code 설정에서 moai 층만 읽기 전용으로 진단한다. 답은 JSON 하나이고 `writes` 가 false, `checks`(각각 `id`, `state` ok·warn·fail·unknown·not_applicable, `summary`, `evidence`, `next`)와 `pointers`(Claude Code 자체 점검의 이름, 각각 `verified`·`unverified`)가 들어 있다.

첫 점검: `config_folder`(`host_home` 로 Claude 설정 폴더를 풀고, 비었거나 상대 경로인 `CLAUDE_CONFIG_DIR` 는 변수 이름만 말하며 fail), `claude_version`(`claude --version` 을 5초 상한으로 한 번 돌려 최소 버전 표의 `verified` 행마다 ok·below), `login`(변수 일곱 개의 설정 여부, `.credentials.json` 존재만, 문서의 우선순위가 고르는 자격 증명; `--auth-status` 가 있을 때만 `claude auth status` 를 10초 상한으로 한 번 돌려 `authMethod`·`configDirectory`·`loginExpiresAt` 과 종료 코드만 남김), `api_key`(키 변수나 `apiKeyHelper` 가 있으면 warn, 이름만).

비밀 값·키 접두사·길이·`apiKeyHelper` 출력·`.credentials.json` 내용은 어디에도 나오지 않는다. 파일과 레저를 쓰지 않고 네트워크를 열지 않으며 `claude --version`(과 요청 시 `claude auth status`) 밖의 프로세스를 시작하지 않는다.

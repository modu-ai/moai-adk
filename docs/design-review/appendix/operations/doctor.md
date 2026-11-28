# `doctor`

operation `doctor` · 코드 `src/ops/doctor.rs`, `src/moai_doctor.rs`, `src/moai_doctor/locks.rs`, `src/moai_doctor/capabilities.rs` · SPEC `SPEC-20261127-doctor-moai-layer`

`doctor [--project PATH] [--auth-status] [--table] [--managed-dir PATH]` 는 Claude Code 설정에서 moai 층만 읽기 전용으로 진단한다. 답은 JSON 하나이고 `writes` 가 false, `checks`(각각 `id`, `state` ok·warn·fail·unknown·not_applicable, `summary`, `evidence`, `next`)와 `pointers`(Claude Code 자체 점검의 이름, 각각 `verified`·`unverified`)가 들어 있다.

점검: `config_folder`(`host_home` 로 Claude 설정 폴더를 풀고, 비었거나 상대 경로인 `CLAUDE_CONFIG_DIR` 는 변수 이름만 말하며 fail), `claude_version`(`claude --version` 을 5초 상한으로 한 번 돌려 최소 버전 표의 `verified` 행마다 ok·below), `login`(변수 일곱 개의 설정 여부, `.credentials.json` 존재만, 문서의 우선순위가 고르는 자격 증명; `--auth-status` 가 있을 때만 `claude auth status` 를 10초 상한으로 한 번 돌려 `authMethod`·`configDirectory`·`loginExpiresAt` 과 종료 코드만 남김, 시간 초과면 프로세스 그룹째 멈춘다), `api_key`(키 변수나 `apiKeyHelper` 가 있으면 warn, 이름만).

`managed_locks`: 관리 설정 파일(`managed-settings.json`, `managed-settings.d/*.json` 알파벳 순)과 사용자·프로젝트·로컬 설정을 파일로만 읽어 `allowManagedHooksOnly`·`strictPluginOnlyCustomization`·`allowManagedModsOnly`·`disableAllHooks` 를 문서의 범위 규칙대로 effective·ignored·unset·unknown 으로 답한다(effective 면 warn). 읽을 수 없는 파일은 이름과 함께 unknown, 파일이 아닌 소스(`policyHelper`, 서버 관리 설정, 세션의 `--settings`, macOS 관리 환경설정 도메인, Windows 레지스트리)는 `not_read` 로 이름만 말한다. `--managed-dir` 은 관리 설정 폴더를 바꾸는 읽기 전용 옵션이다(기본은 플랫폼 폴더, 테스트는 임시 폴더를 준다).

`cross_session_inbound`: `crossSessionInbound` 의 유효 값을 문서의 우선순위(관리 파일, 사용자 설정; 프로젝트·로컬은 더 엄격하거나 신뢰 값이 없을 때만)로 구해 값과 파일을 말한다. unset·hold·refuse 는 `-l`·`-a` 레인이 권한 우회로 돌아 메시지를 붙잡으므로 warn, accept 는 ok.

`host_capabilities`: `ANTHROPIC_BASE_URL` 의 호스트가 `api.anthropic.com` 이 아니면 gateway 열, 아니면 cli 열의 능력 칸(기능 여덟 개)을 보인다. URL 은 출력하지 않는다. `--table` 이면 최상위 `capability_table` 에 다섯 열(cli·gateway·desktop·vscode·cloud)과 여덟 행 전체가 나오고 칸마다 `value`(yes·no·partial·unverified)·`docs`·`line`·`read_at` 이 있다. 문서가 말하지 않았거나 추론뿐인 칸은 unverified 이고 yes 로 나오지 않는다. 그런 열한 칸은 문서 페이지 일곱(`cross-session-messaging`, `output-styles`, `statusline`, `artifacts`, `desktop`, `vs-code`, `llm-gateway-connect`)을 2026-11-28 에 다시 읽고도 닫히지 않았다: 그 칸의 `read_at` 은 2026-11-28 이고 `line` 은 페이지가 대신 말하는 것이다(칸을 말하는 줄이 없다는 사실 포함). 나머지 칸의 `read_at` 은 2026-11-27 이다.

비밀 값·키 접두사·길이·`apiKeyHelper` 출력·`.credentials.json` 내용은 어디에도 나오지 않는다. 파일과 레저를 쓰지 않고 네트워크를 열지 않으며 `claude --version`(과 요청 시 `claude auth status`) 밖의 프로세스를 시작하지 않는다.

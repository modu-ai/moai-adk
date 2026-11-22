# `project-check`

operation `project.check` · 코드 `src/ops/project_check.rs`, `src/project_check.rs` · SPEC-20261122-project-ci-design (REQ-5)

`project-check PROJECT [--timeout-secs N]` 는 프로젝트 프로필에 선언한 검사 명령을 선언한 순서대로 프로젝트 폴더에서 실행한다. 생성된 CI 워크플로의 `check` 작업이 실행하는 목록과 같은 목록이고, 두 목록이 같다는 것은 `tests/project_ci_compare_lib.rs` 가 증명한다.

- 단계마다 `id`, `command`, `exit`(시간 초과나 신호로 끝나면 `null`), `timed_out`, `seconds`, `output_tail`(끝 4096바이트). 앞 단계가 실패해도 나머지는 모두 실행한다. 한도(`--timeout-secs`, 기본 1800)를 넘긴 명령은 프로세스 그룹째 멈추고 실패로 센다.
- 답의 `summary` 한 줄은 `project-check <id>=<exit> … result=PASS|FAIL`. `check affected` 처럼 봉투는 `ok` 이고 결과는 이 줄에 있다.
- 선언한 검사가 없는 프로젝트는 아무것도 실행하지 않고 평이한 설명으로 거절한다. 저장소 자체의 `check affected`·`tools/local_check.sh` 와는 별개다.
- 명령은 git 의 저장소 환경 변수를 뺀 환경에서 `sh -c` 로 실행한다. 이 연산은 자기 파일을 쓰지 않는다.

# `check-ready`

operation `check-ready` · 코드 `src/ops/check_ready.rs` · SPEC-20261120-shipped-workflows-commands (조각 2)

`check-ready --project <폴더>`(기본 `.`) 는 프로젝트가 **선언한** 준비 전 검사를 읽고 맞는지 보는 읽기 전용 명령이다. 검사 대상은 프로젝트가 선언한 것뿐이다: `.moai/specs/` 의 SPEC 문서 전부(폴더가 없으면 `specs: absent` 로 보고하고 검사 없이 지나간다)와 프로필(`project.yaml`)이 `test_paths` 로 선언한 경로 전부(프로젝트 안에 실제로 없는 선언 경로는 문제 하나다; 형식이 잘못된 선언은 `read_profile` 이 이미 거부하므로 `profile: invalid` 로 문제가 된다). git 범위를 읽지 않고 시험도 돌리지 않는다 — 변경 범위에서 검사를 고르는 것은 `check`( `moai check affected` , `src/check_affected.rs` )의 몫이므로 두 명령은 겹치지 않는다. SPEC 파서·검사는 `moai_cli::spec_doc`(`spec check` 와 같은 결과), 프로필 읽기와 선언 검증은 `moai_cli::read_profile` 을 쓴다.

답: `action`(`check-ready`), `project`, `profile`(`present` / `absent` / `invalid`), `declared_paths`(선언 경로 수), `missing_paths`(없는 선언 경로 수), `specs`(`present` / `absent`), `spec_folders`(SPEC 폴더 수), `spec_errors`·`spec_warnings`, `nothing_declared`(프로필과 SPEC 폴더가 둘 다 없음), `result: "PASS"`, `writes: false`. 문제가 하나라도 있으면 envelope 대신 stderr 에 `check-ready refused: <n> problem(s): <문제들>` 한 줄과 exit 1 이다. 문제 문구: `declared test path missing: <선언 경로>`, `profile: invalid profile (<이유>)`, `<SPECS>/<폴더>/spec.md<줄> <코드>`(spec check 의 error finding 과 같은 이름). 프로필과 SPEC 폴더가 둘 다 없으면 `nothing_declared: true` 로 성공(exit 0)한다 — 빈 프로젝트는 선언한 것이 없어서 검사할 것도 없다.

쓰기 0: 읽기뿐이고 프로젝트 나무에 아무것도 남기지 않는다. `<project>` 폴더가 없거나 SPEC 폴더를 못 읽으면 `check-ready refused` 와 exit 1, 사용법 오류는 exit 2 다.

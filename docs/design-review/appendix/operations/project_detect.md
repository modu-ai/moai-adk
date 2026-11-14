# `project-detect`

operation `project.detect` · 코드 `src/ops/project_detect.rs` · #2496

`project-detect PROJECT` 는 `/moai project` 경로를 위해 프로젝트 사실(새 프로젝트인지, manifest, 언어, 언어 서버, 설정된 MCP 서버 이름, 데이터베이스)을 찾는다. 읽기만, `detected` 가 채워진 프로필을 `project-plan` 에 넘긴다. `src/main.rs` 에서 그대로 옮김(#2834 R4a).

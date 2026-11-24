# `tidy-plan`

operation `tidy.plan` · 코드 `src/ops/tidy_plan.rs`, `src/tidy/plan.rs` · SPEC-20261124-home-tidy-slice2 (부모 SPEC-20261122-home-tidy)

`tidy-plan HOME` 는 `HOME/.moai` 에서 무엇을 보관하거나 지울 수 있는지 미리 보여 준다. 쓰지 않고 명령을 시작하지 않는다. 인자와 상한은 `tidy-inventory` 와 같다.

- `items`: 항목마다 `path`(`.moai` 기준), `kind`, `action`(`archive`·`delete`·`keep`), `rule`, `reason`, `bytes`, `newest`. 사실을 확인할 수 없으면 `keep` 이고 이유는 `unknown` 이다.
- `delete` 는 옛 도구(`retired_tool`)의 빈 폴더 하나뿐이다(`retired_tool_empty`). 내용이 있는 옛 도구 항목은 `archive`(`retired_tool_archive`)다. 빌드 산출물과 깨진 등록의 삭제는 조각 4 가 증명 규칙과 함께 더한다.
- 종류가 임계값 하나라도 넘으면 `reports` 는 나이 규칙(`archive_after_days`, 기본 90일 초과)에 닿는 파일을 오래된 순으로 `archive`(`reports_age`)하고 나머지는 `keep`(`not_old_enough`)한다. 모든 임계값 아래면 아무것도 내지 않고 `keep`(`under_threshold`)한다.
- `worktrees`·`build`·`evidence`·`receipts` 의 항목은 이 조각에서 `keep` 이다. 임계값을 넘으면 `proof_pending`, 아니면 `under_threshold`. git 증명(병합·랜딩, 조각 4)이 있어야 후보가 된다.
- `logs` 는 쓰는 쪽을 몰라서(부모 D-10) 폴더 하나가 `keep`(`unknown`)이다. 다른 제품, 알 수 없는 항목, 보호 항목은 나열만 하고 `keep` 이다.
- `digest`: `archive`·`delete` 항목의 `path·action·rule·bytes·newest` 줄의 sha256. 변하지 않은 폴더는 같은 값이고 후보가 바뀌면 달라진다. 조각 3 이 같은 값을 확인 토큰으로 쓴다.
- `memory_index`: `tidy-inventory` 와 같은 줄. 어느 파일이든 80% 를 넘으면 `memory_tidy` 가 붙는다: 세 묶음(`rules_to_instructions`·`retired_items`·`merges`), 사용자가 Claude 세션에서 돌릴 문구(`run_in_claude_session`), 그리고 참고 줄 `purge_reference`(`claude purge --all`, `never_run: true`). moai 는 `MEMORY.md` 를 고치지 않고 `claude purge` 를 실행하지 않는다.

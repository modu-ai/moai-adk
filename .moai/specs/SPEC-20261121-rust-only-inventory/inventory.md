# 파이썬 파일별 판정 목록 (조각 1)

이 문서는 `SPEC-20261121-rust-only-inventory` 의 본문이 가리키는 표예요. 추적 중인 파이썬 파일 84개(`git ls-files -- '*.py'`, 26386줄)를 한 줄씩 판정해요. 이 조각은 아무것도 지우지 않고 검사 단계도 바꾸지 않아요.

## 읽는 법

- 판정은 셋 중 하나예요. **port** = Rust 시험이나 `moai` 명령으로 옮겨요(대체 칸에 이름). **covered** = Rust 쪽이 이미 같은 사실을 단정해요(대체 칸에 이름, 빠진 사례가 있으면 그 사실). **drop** = 대체 없이 지워요(대체 칸에 잃는 보호). 이름이 `(이름은 조각에서 정함)` 인 명령은 제안이에요.
- 부르는 곳의 `U` = 모든 `*_test.py` 는 `tools/local_check.sh` 의 `unittest` 단계(`python3 -m unittest discover -s tests -p '*_test.py'`)와 CI 의 rest 잡이 돌려요. `CA` = `check affected` 가 `tools/repo_scans.json` 의 규칙으로 그 모듈 하나만 골라 돌려요(규칙 8개, 아래 목록).
- 크기와 시험 수는 `ast` 로 센 값이고, 부르는 곳은 추적 파일 전체에서 그 파일 이름을 찾은 결과를 읽어 정리했어요. 보호 내용은 각 파일의 머리말과 시험 이름을 읽은 것이라, 조각 2·3 이 옮길 때 시험 본문을 다시 읽어 사례별로 대조해요.

합계: port 65, covered 2, drop 17 (총 84).

`check affected` 가 파이썬 시험을 고르는 규칙 8개(`tools/repo_scans.json`): `tracked-skill-references`, `ci-scope-docs-paths`, `github-templates`, `github-assets`, `design-review-fixes`, `github-fixtures-private-shapes`, `repo-scans-table`, `source-manifest-tool-output`. 조각 2 는 이 규칙의 검증 명령을 `cargo test` 로 바꾸고 `src/check_affected.rs` 의 `python` 단계(`python3 -m unittest tests.<이름>`)를 걷어요.

## 1. 도구 22개와 SPEC 안의 check.py 2개

| 파일 | 크기 | 보호하거나 하는 일 | 부르는 곳 | 판정 | 대체(Rust 시험 또는 `moai` 명령)·이유 |
|---|---|---|---|---|---|
| `tools/atlas_baseline.py` | 243줄 | C10a 고정 입력 기준선과 임계값 판정 | Rust 시험 `tests/search_atlas.rs` 가 `--judge` 를 실행, 시험 1개 | port | 판정 로직을 `tests/search_atlas.rs`(또는 tests/common)의 Rust 함수로 |
| `tools/atlas_three_arm.py` | 236줄 | C10b: 어휘·+Map·+Graph 세 경로를 고정 입력에 실행 | Rust 시험 `tests/atlas_three_arm.rs` 가 두 번 실행, 시험 1개 | port | `tests/atlas_three_arm.rs` 안으로 이관 |
| `tools/audit_eval.py` | 251줄 | 감사 에이전트가 실제로 중요했던 버그를 잡는지 채점 | .moai/evals/audit/README.md, 시험 1개 | drop | 일회성 측정 도구예요. 잃는 보호: 평가 세트를 다시 돌려 채점하는 방법. 필요하면 나중에 `moai` 명령으로 새로 만들어요(결정 D2) |
| `tools/build_folder.py` | 456줄 | 워크트리마다 cargo 빌드 폴더를 `~/.moai/build/` 에 하나씩 두고 소유 기록을 남기는 도구(`env`·`clean`) | tools/local_check.sh, tools/self_check.sh, 스킬 build-folder.md(에이전트에게 실행하라고 지시), Rust 시험 `tests/lane_env_cli.rs`(두 구현을 같은 저장소에 돌려 비교) | port | `env` 는 `moai lane-env`(src/lane_env.rs)가 이미 같은 이름·기록을 만들어요(covered). `clean` 은 Rust 대응을 찾지 못했어요(조각 3 에서 재확인). 없으면 `moai lane-env` 하위 동작으로 옮겨야 해요 |
| `tools/candidate_summary.py` | 135줄 | 후보 검증 한 건의 요약과 판정(건너뛴 대상은 실패) | CI(candidate.yml), github_assets 시험이 파일 이름을 단정 | port | `moai candidate-summary`(CI 단계라 결정 D3) |
| `tools/check_ac_commands.py` | 175줄 | SPEC 에 적은 `cargo test --test X …` 검증 명령이 있는 시험 대상·필터를 가리키는지 점검 | Rust 시험 `tests/ac_command_lint.rs` 가 실행(python3 -I), .config/nextest.toml·repo_scans.json 이 이름을 적음 | port | `tests/ac_command_lint.rs` 안의 Rust 함수로 옮겨요(그 시험이 이미 같은 종류의 Rust 점검을 가져요). 별도 명령은 필요 없어요 |
| `tools/check_guard.py` | 248줄 | PR 이 자기를 판정하는 검사를 약화하지 못하게 막는 CI 점검 | CI(check-guard.yml: 베이스 커밋의 이 파일 사본을 실행), src/gate.rs, tests/protected_paths_one_list.rs, tools/protected_checks.json·review_checklist.json | port | `moai check-guard`. 베이스 커밋의 사본을 돌리는 신뢰 모델을 Rust 로 어떻게 지킬지 결정 D6 |
| `tools/check_support_matrix.py` | 149줄 | 지원 행렬에서 supported 행은 통과한 영수증을 인용해야 한다는 점검 | 시험 2개(support_matrix_test, host_live_harness_test), 문서 7곳이 언급 | port | 지원 행렬 문서가 남는 동안은 문서 점검이라 `tests/docs_scans.rs` 의 Rust 시험으로(문서가 사라지면 같이 drop) |
| `tools/ci_scope.py` | 220줄 | 변경 파일 집합이 문서만인지 전체인지, macOS 대상이 필요한지 분류 | CI(ci.yml 3곳), tools/local_check.sh, tools/protected_checks.json·repo_scans.json·review_checklist.json 의 문구, src/gate.rs 의 보호 경로 목록(Python 파일 이름을 직접 적음), Rust 시험 6곳(경로 이름 단정) | port | `moai ci-scope`. CI 단계에서 Python 을 걷는 일이라 조각 2 에 넣을지 결정 D3 |
| `tools/ci_superseded.py` | 138줄 | 병합 뒤 CI 실행이 실패인지 새 병합에 밀려 대체된 것인지 가림 | CI(ci.yml 1곳), github_assets 시험 | port | `moai ci-superseded`(CI 단계라 결정 D3) |
| `tools/flow_metrics.py` | 155줄 | PR·이슈 흐름 기준선을 `gh` 호출로 읽기 전용 측정 | docs/design-review/flow-control-2026-11-09.md, 시험 1개 | drop | 2026-11-09 기준선을 재현하던 일회성 도구예요. 잃는 보호: 그 측정의 재현(결정 D2) |
| `tools/heavy_check_slot.py` | 569줄 | 무거운 검사 동시 실행 수 제한(슬롯 폴더 잠금) | 스킬 verify.md 는 이미 `moai heavy-slot-acquire` 를 가리켜요. Python 을 실행하는 곳은 Rust 시험 `tests/heavy_slot_cli.rs`(같은 사례를 양쪽에 돌리는 동등성 시험 2곳)와 `src/heavy_slot.rs:1318` 의 시험(파일 내용을 컴파일 때 `include_str!` 로 담음: 파일을 지우면 빌드가 깨져요) | covered | Rust 명령 `moai heavy-slot-acquire`/`heavy-slot-status`(src/heavy_slot.rs)가 이미 같은 폴더·잠금을 써요. 동등성 시험 2곳을 Rust 단독 단정으로 바꾼 뒤 Python 을 지워요 |
| `tools/home_entry_count.py` | 151줄 | 검증 실행 전후 `<HOME>/.moai` 항목 수가 같은지 측정 | docs 2곳, `tests/home_isolation_scan.rs` 는 주석으로만 언급, 시험 1개 | drop | 측정 수동 도구예요. Rust 시험은 시험마다 임시 HOME 을 써서 실제 HOME 오염을 따로 막아요(`tests/home_isolation_scan.rs`). 잃는 보호: 실제 `~/.moai` 항목 수 전후 비교(결정 D2) |
| `tools/hook_latency.py` | 146줄 | `moai-cli hook-dispatch` 호출당 지연 P50/P95 측정 | 시험 1개(백분위 계산만) | drop | 일회성 측정이에요. 잃는 보호: 지연 수치 재측정(결정 D2) |
| `tools/host_live_harness.py` | 336줄 | Claude Code·Codex 실제 세션에서 지침 로딩·스킬 노출·훅 발화를 격리 프로젝트로 측정 | docs 2곳, 시험 1개(오프라인 검사만) | drop | 실제 호스트가 있어야 의미가 있는 옛 시기 도구예요. Codex 지원 제거와 함께 반은 이미 무의미해요. 잃는 보호: 실제 세션 측정 재현(결정 D2) |
| `tools/legacy_verdicts.py` | 253줄 | `.moai-cli/manifests/legacy-asset-map.json` 의 행별 판정을 규칙에서 다시 만듦 | 그 JSON, 스킬 source-manifest 의 항목, 시험 1개 | drop | Go→Rust 이관 시기의 옛 도구예요. 잃는 보호: 맵이 규칙 재생성 결과와 같다는 단정(결정 D2) |
| `tools/merge_driver_manifest.py` | 345줄 | source-manifest.json 의 sha256 줄을 병합 때 다시 계산하는 git 병합 드라이버 | .gitattributes, 각 클론의 `merge.*.driver` 설정(`--install` 로 등록), tools/source_manifest.py, 시험 2개 | port | `moai` 하위 명령. 클론마다 등록된 드라이버 명령이 `python3` 로 되어 있어서 이관 뒤 클론이 다시 등록해야 해요(조각 3 의 위험) |
| `tools/mutate.py` | 1034줄 | 격리된 HEAD 복제본에서 한 줄 변형을 돌려 시험이 잡는지 보는 변형 실행기 | 스킬·감사 절차(docs/mutation-runner.md), 감사 에이전트 번들 글, 실행기 시험 7개 | port | `moai mutate`(이름은 조각에서 정함). 조각 3 의 첫 후보 |
| `tools/sandbox_write_probe.py` | 343줄 | 샌드박스가 원장의 파일 연산을 쓸 수 있는지 탐침 | docs/design-review/appendix/ledger_sandbox_probe.md, 시험 1개 | drop | 일회성 탐침이에요. 잃는 보호: 샌드박스별 쓰기 가능 표 재생성(결정 D2) |
| `tools/smoke/smoke.py` | 615줄 | 빈 HOME·임시 git 프로젝트에서 빌드한 바이너리를 돌리는 공용 smoke 하네스(레인·감사자 공용) | tools/smoke/run.sh → CI check 잡의 smoke 단계(ci.yml), 감사 절차 | port | `moai smoke`. run.sh 는 셸이라 남기되 호출만 바꿔요 |
| `tools/source_manifest.py` | 375줄 | `skills/moai/references/source-manifest.json` 을 패키지 파일에서 다시 만들거나(`write`) 맞는지 확인(`check`) | tools/self_check.sh(항목 manifest), 스킬 verify.md·self-check.md(에이전트에게 실행하라고 지시), repo_scans.json 규칙 `source-manifest-tool-output`, Rust 시험 3곳이 그 문구를 단정(`tests/agent_defs_package.rs`, `tests/cli_text_citations.rs`, `tests/codex_adapter.rs`) | port | `moai` 하위 명령(예: `skill-manifest write/check`). 병합 드라이버와 한 쌍으로 조각을 나눌지는 결정 D1 |
| `tools/statusline_golden.py` | 437줄 | Claude 상태줄 세 줄 계약을 새 HOME·프로젝트마다 실행해 골든과 비교 | Rust 시험 `tests/statusline_golden.rs`(그대로 실행), `tests/statusline_guard_false_positive.rs`, `tests/statusline_tag_order.rs`(명령 문구 단정), 시험 1개 | port | 골든 비교를 Rust 하네스(tests/common)로. 현재 Rust 시험은 이 Python 을 그대로 부르기만 해요 |
| `.moai/specs/SPEC-20261120-guidance-sweep-core/check.py` | 118줄 | 그 SPEC 의 AC 검증(`python3 -I …/check.py`, `--sentences`) | 그 SPEC 의 verify 줄 2곳 | port | AC 가 Rust 시험을 가리키게 SPEC 의 verify 를 바꾸고 파일을 지워요(SPEC 은 draft 라 수정 가능) |
| `.moai/specs/SPEC-20261120-statusline-docs-items-2-3/check.py` | 43줄 | 문서가 이름 붙인 시험 파일·픽스처·골든 사례·시험 함수가 실제로 있는지 점검(AC-4) | 그 SPEC 의 verify 줄 | port | Rust 시험(docs 이름 존재 점검)으로 옮기고 SPEC 의 verify 를 바꿔요 |

## 2. 시험 60개

| 파일 | 크기 | 보호하거나 하는 일 | 부르는 곳 | 판정 | 대체(Rust 시험 또는 `moai` 명령)·이유 |
|---|---|---|---|---|---|
| `tests/agents_local_build_folder_test.py` | 65줄, 시험 5 | 저장소의 개인 지침이 빌드 폴더를 워크트리마다 하나로 두고 check set 이 CARGO_TARGET_DIR 을 정하지 않는다(5) | U | port | `tests/docs_scans.rs`(local_check.sh 문구 단정) |
| `tests/atlas_baseline_test.py` | 268줄, 시험 10 | C10a 코퍼스·기준선 일관성(10) | U | port | `tests/search_atlas.rs` 쪽 Rust 시험(도구 이관과 같은 조각) |
| `tests/atlas_three_arm_selection_test.py` | 149줄, 시험 9 | 그래프 근거가 고정 오라클과 별개로 답 선택을 바꾼다(9) | U, 그리고 Rust 시험 `tests/atlas_three_arm.rs` 가 이 파일을 python3 로 실행 | port | `tests/atlas_three_arm.rs` 안으로 이관 |
| `tests/audit_eval_test.py` | 391줄, 시험 14 | audit_eval 도구와 평가 세트(14) | U | drop | 도구와 함께(결정 D2) |
| `tests/build_folder_test.py` | 667줄, 시험 42 | 빌드 폴더 도구의 이름·소유 기록·정리 동작(42) | U | port | `tests/lane_env_cli.rs`(8개)가 `env` 일부만 덮어요. 나머지와 `clean` 은 Rust 로 옮겨요(같은 조각) |
| `tests/candidate_summary_test.py` | 182줄, 시험 12 | 후보 요약 판정(12) | U | port | `moai candidate-summary` 의 Rust 시험 |
| `tests/cargo_metadata_test.py` | 60줄, 시험 5 | Cargo.toml 패키지 메타데이터가 저장소 사실과 맞다(5) | U, `tests/product_cli.rs` 가 같이 바뀌어야 한다고 주석 | port | `tests/product_cli.rs` 나 `tests/docs_scans.rs` |
| `tests/check_guard_test.py` | 326줄, 시험 24 | check_guard 오프라인 시험(24) | U | port | `moai check-guard` 의 Rust 시험(같은 조각) |
| `tests/check_pr_reaudit_test.py` | 452줄, 시험 48 | 명시적 재감사 CLI 완전성 점검(48) | U | port | `tests/self_audit_cli.rs` |
| `tests/check_pr_self_audit_test.py` | 330줄, 시험 39 | PR 본문 검사(`moai-cli pr-self-audit-check`)의 네 줄 측정·형식 규칙(39) | U(바이너리 필요: CI 가 먼저 빌드) | port | `tests/self_audit_cli.rs`(12개)를 늘려요 |
| `tests/check_pr_zero_selected_test.py` | 321줄, 시험 32 | `0 passed` 를 근거로 쓴 경우와 필터가 둘 이상의 대상을 고르는 경우(32) | U | port | `tests/self_audit_cli.rs` |
| `tests/ci_install_ripgrep_test.py` | 377줄, 시험 20 | Linux 설치 스크립트를 격리된 명령·아카이브로 실행하는 시험(20, PyYAML 사용) | U | port | `tests/ci_install_ripgrep_cli.rs`(셸 스크립트는 남기고 Rust 가 가짜 명령으로 실행) |
| `tests/ci_scope_test.py` | 903줄, 시험 47 | ci_scope 분류와 macOS 판정(47, docs 경로 규칙 포함) | U, CA(규칙 `ci-scope-docs-paths`) | port | `moai ci-scope` 의 Rust 시험(같은 조각) |
| `tests/ci_smoke_step_test.py` | 323줄, 시험 16 | ci.yml 의 check 잡이 check set 뒤에 smoke 하네스를 돌린다는 워크플로 문구 점검 | U | port | `tests/ci_workflow_scan.rs`(워크플로 텍스트 점검) |
| `tests/ci_superseded_test.py` | 249줄, 시험 13 | 대체된 실행 판정(13, 고정 JSON 픽스처 사용) | U | port | `moai ci-superseded` 의 Rust 시험 |
| `tests/claude_statusline_wrapper_test.py` | 416줄, 시험 38 | PATH 에 `moai` 이름이 없어도 Claude Code 상태줄 래퍼가 moai-cli 에 닿는다(38) | U | port | `tests/statusline_wrapper_cli.rs`(기존 `tests/statusline_*.rs` 와 같은 하네스) |
| `tests/dev_profile_sha2_test.py` | 33줄, 시험 2 | dev 프로파일 sha2 최적화 고정(2) | U | port | Cargo.toml 을 읽는 Rust 시험 2개 |
| `tests/dist_include_test.py` | 60줄, 시험 4 | 릴리스 아카이브가 스킬 패키지와 기본 에이전트를 바이너리 곁에 담는다(4) | U | port | `tests/release_suite.rs` |
| `tests/e2e_host_claude_verdict_test.py` | 175줄, 시험 6 | tests/e2e/host-claude.sh 의 AC-08C-3·6 판정(가짜 `claude` 로 구동)(6) | U | port | `tests/e2e_host_claude_verdict.rs`(`tests/e2e_host_fixture.rs` 가 이미 Rust 픽스처) |
| `tests/e2e_host_codex_verdict_test.py` | 167줄, 시험 9 | tests/e2e/host-codex.sh 의 AC-08C-4 판정(9) | U | drop | Codex 지원 제거와 함께 의미를 잃어요. 잃는 보호: Codex 호스트 판정 점검(결정 D2·D7) |
| `tests/e2e_receipt_not_run_test.py` | 131줄, 시험 2 | 호스트 스크립트의 인자 조립·영수증 형식·스킬 사전 복사·다이제스트 기록(2) | U | port | `tests/e2e_host_claude_verdict.rs`(Claude 반만, Codex 반은 제거와 함께 drop) |
| `tests/flow_metrics_test.py` | 152줄, 시험 7 | flow_metrics 픽스처(7) | U | drop | 도구와 함께(결정 D2) |
| `tests/github_assets_test.py` | 949줄, 시험 46 | GitHub 이슈 양식·라벨 분류·워크플로 파일 구조 점검(46, PyYAML 사용) | U, CA(규칙 `github-assets`), ci.yml·candidate-target.yml 이 PyYAML 설치 단계를 둠 | port | `tests/github_assets_scan.rs`. YAML 파싱 방식은 결정 D4 |
| `tests/github_inventory_collect.py` | 435줄 | RC1 00-B 용 GitHub 설정 목록·중복 지도 읽기 전용 수집기(시험 폴더에 있는 도구) | 시험 1개, 문서 3곳 | drop | 일회성 수집기예요. 잃는 보호: 설정 목록 재수집(결정 D2) |
| `tests/github_inventory_test.py` | 192줄, 시험 17 | GitHub 설정 목록 수집기 오프라인 점검과 픽스처에 토큰·/Users/·이메일 모양이 없다는 단정(17) | U, CA(규칙 `github-fixtures-private-shapes`) | port | 비밀 모양 단정만 `tests/github_assets_scan.rs` 로 옮기고 수집기 시험은 수집기와 함께 drop |
| `tests/github_templates_test.py` | 247줄, 시험 12 | Rust GitHub 템플릿의 정적 점검(12, PyYAML 사용) | U, CA(규칙 `github-templates`) | port | `tests/github_assets_scan.rs`(템플릿 코드는 src/github_templates.rs 가 이미 Rust) |
| `tests/heavy_check_slot_test.py` | 646줄, 시험 25 | 실제 프로세스와 임시 슬롯 폴더로 슬롯 제한 검증(25) | U | covered | `tests/heavy_slot_cli.rs`(같은 사례를 Rust 명령에 돌림). 25개 중 Rust 에 같은 사례가 없는 것을 조각 3 에서 사례별로 대조해 빠진 것만 옮겨요 |
| `tests/home_entry_count_test.py` | 261줄, 시험 23 | home_entry_count(23) | U | drop | 도구와 함께(결정 D2) |
| `tests/hook_latency_test.py` | 66줄, 시험 6 | 백분위 수학·빌드 프로파일 표기·사용법 종료 코드(6) | U | drop | 도구와 함께(결정 D2) |
| `tests/host_live_harness_test.py` | 219줄, 시험 8 | host_live_harness 오프라인 점검(8) | U | drop | 도구와 함께(결정 D2). 지원 행렬 점검은 support_matrix_test 쪽에서 이관 |
| `tests/host_matrix_test.py` | 158줄, 시험 5 | docs/host-support-matrix.md 의 호스트 행 오프라인 점검(5) | U | port | `tests/docs_scans.rs`. Codex 행은 Codex 지원 제거 때 같이 정리돼요 |
| `tests/install_script_test.py` | 1489줄, 시험 64 | install/install.sh 와 install.ps1 시험(64) | U | port | `tests/install_script_cli.rs`(sh 는 실행, ps1 은 이 환경에서 못 돌려서 정적 점검만) |
| `tests/legacy_verdicts_test.py` | 54줄, 시험 5 | 옛 자산 맵이 규칙 재생성과 같다(5) | U | drop | 도구와 함께(결정 D2) |
| `tests/local_check_test.py` | 1117줄, 시험 54 | tools/local_check.sh 의 단계 목록 고정(54) | U | port | `tests/local_check_script.rs`. 조각 2 가 Python 단계를 걷으면서 같이 바뀌어요 |
| `tests/mutate_compiler_test.py` | 264줄, 시험 20 | 실제 오프라인 Cargo 로 확인하는 증거(합성 협력자와 구분) | U | port | `moai mutate` 시험(같은 조각) |
| `tests/mutate_deadline_test.py` | 301줄, 시험 11 | 핸드셰이크 기반 마감 계약(벽시계 가정 없음) | U | port | `moai mutate` 시험(같은 조각) |
| `tests/mutate_gitlink_test.py` | 286줄, 시험 10 | gitlink 가 깨끗한지 계약 | U | port | `moai mutate` 시험(같은 조각) |
| `tests/mutate_native_tool_test.py` | 244줄, 시험 6 | 실행 파일 별칭 해석 경계 | U | port | `moai mutate` 시험(같은 조각) |
| `tests/mutate_source_safety_test.py` | 235줄, 시험 13 | 소스 경계 회귀(src 바깥을 건드리지 않음) | U | port | `moai mutate` 시험(같은 조각) |
| `tests/mutate_test.py` | 705줄, 시험 21 | 변형 실행기 CLI 의 동작 계약(표준 라이브러리만) | U | port | `moai mutate` 의 Rust 통합 시험(도구와 같은 조각) |
| `tests/mutate_timeout_pin_test.py` | 156줄, 시험 7 | mutate_test 의 하위 마감이 부하와 무관하다는 고정 시험 | U | port | `moai mutate` 시험(같은 조각). 마감 방식이 바뀌면 의미가 달라질 수 있어 Rust 쪽 설계에 맞춰 다시 써요 |
| `tests/omp_host_survey_test.py` | 158줄, 시험 9 | 호스트별 기능 조사표가 코드와 맞는지(9) | U | port | `tests/docs_scans.rs` |
| `tests/oss_docs_test.py` | 268줄, 시험 18 | 공개 문서(README 등)의 오프라인 점검(18) | U | port | `tests/docs_scans.rs` |
| `tests/platform_targets_test.py` | 81줄, 시험 5 | 릴리스(dist) 대상이 지원 플랫폼 셋(5) | U | port | `tests/docs_scans.rs` 나 `tests/release_suite.rs` |
| `tests/prop_manifest_test.py` | 141줄, 시험 2 | 매니페스트 속성 시험 2개(무작위 입력) | U | port | Rust 속성 시험(같은 조각, 새 크레이트 없이 시드 고정 난수로) |
| `tests/release_rehearsal_test.py` | 64줄, 시험 5 | 릴리스 리허설 문서의 체크리스트 행마다 결과가 있다(5) | U | port | `tests/docs_scans.rs` |
| `tests/release_workflow_test.py` | 413줄, 시험 29 | 릴리스 워크플로가 HeadVer 를 매기고 이름·검사를 두며 게시하지 않는다(29, PyYAML 사용) | U | port | `tests/ci_workflow_scan.rs`. YAML 방식은 결정 D4 |
| `tests/repo_scans_test.py` | 170줄, 시험 6 | repo_scans.json 표 점검 6개(표 자체와 저장소 시험 파일 훑기; 바이너리 사례는 이미 Rust 로 이관됨) | U, CA(규칙 `repo-scans-table`) | port | `tests/repo_scans_table.rs`. `tests/repo_scans_cli.rs` 는 바이너리 사례만 가져요 |
| `tests/sandbox_write_probe_test.py` | 418줄, 시험 31 | sandbox_write_probe(31) | U | drop | 도구와 함께(결정 D2) |
| `tests/self_audit_run.py` | 40줄 | 위 세 시험이 쓰는 `moai-cli pr-self-audit-check` 실행 도우미(시험 파일이 아님) | 위 세 시험, ci.yml 주석 | drop | 위 세 시험이 Rust 로 가면 쓸 곳이 없어요(같은 슬라이스에서 같이 지움) |
| `tests/self_check_test.py` | 898줄, 시험 66 | tools/self_check.sh 의 기계 항목(66) | U | port | `tests/self_check_script.rs` |
| `tests/setup_approval_fixture_test.py` | 152줄, 시험 6 | SETUP 승인 분리 픽스처(6) | U | port | `tests/skill_text_fixtures.rs` |
| `tests/skill_references_test.py` | 122줄, 시험 5 | 추적 파일의 `skills/...` 참조가 실제 파일을 가리킨다(5) | U, CA(규칙 `tracked-skill-references`), Rust 시험 4곳이 이 시험의 읽기 규칙을 피하려고 주석으로 이름을 적음 | port | `tests/skill_references_scan.rs`(규칙의 verify 를 cargo 로) |
| `tests/skill_web_pages_test.py` | 71줄, 시험 4 | `## Web pages` 절이 워크플로마다 한 번(4) | U | port | `tests/docs_scans.rs` |
| `tests/smoke_harness_test.py` | 632줄, 시험 30 | smoke 하네스가 빈 HOME·임시 프로젝트에서 바이너리를 돌리는 방식 | U | port | `moai smoke` 의 Rust 통합 시험(같은 조각) |
| `tests/source_manifest_test.py` | 861줄, 시험 45 | source_manifest 도구와 병합 드라이버(임시 git 저장소만) | U, CA(규칙 `source-manifest-tool-output`) | port | 새 `moai` 하위 명령의 Rust 통합 시험(같은 조각). 이 시험이 check affected 에서 뽑히므로 규칙의 verify 도 같이 바꿔요 |
| `tests/space_applied_record_test.py` | 105줄, 시험 7 | docs/design-review/applied.md 표와 fixes 폴더 대응(7) | U, CA(규칙 `design-review-fixes`) | port | `tests/docs_scans.rs`(규칙의 verify 를 cargo 로) |
| `tests/statusline_fixtures_test.py` | 144줄, 시험 11 | 상태줄 하네스와 샌드박스 동작(11) | U | port | Rust 골든 하네스의 시험(`tests/statusline_golden.rs` 쪽) |
| `tests/support_matrix_test.py` | 106줄, 시험 5 | 지원 행렬 점검 픽스처와 문서에 비밀 모양·절대 홈 경로가 없다는 단정(5) | U | port | `tests/docs_scans.rs` |
| `tests/sync_drift_fixture_test.py` | 234줄, 시험 9 | SYNC 드리프트 감지 픽스처 4종(9) | U | port | `tests/skill_text_fixtures.rs` |

## 3. 파이썬을 실행하거나 파이썬 파일을 읽는 Rust 쪽

`git grep -n -E '\.py\b|python3?\b|unittest|_test\.py'` 를 `tests/*.rs`, `src/**/*.rs`, `build.rs` 에 돌린 94줄을 읽고 나눴어요. 파이썬 파일을 지우는 조각이 이 줄들을 먼저 바꿔야 해요.

### 3-1. 파이썬을 실제로 실행해요 (지우기 전에 Rust 로 바꿔야 하는 시험)

| Rust 파일 | 실행하는 것 | 해야 할 일 |
|---|---|---|
| `tests/ac_command_lint.rs:119` | `tools/check_ac_commands.py` | 점검을 Rust 함수로 옮겨요 |
| `tests/atlas_three_arm.rs:9`, `:111` | `tools/atlas_three_arm.py`, `tests/atlas_three_arm_selection_test.py` | 둘 다 Rust 로 이관 |
| `tests/search_atlas.rs:151` | `tools/atlas_baseline.py --judge` | 판정을 Rust 로 이관 |
| `tests/statusline_golden.rs:25` | `tools/statusline_golden.py --check` | 골든 비교를 Rust 하네스로 이관 |
| `tests/heavy_slot_cli.rs:1290`, `:1404` | `tools/heavy_check_slot.py`(Rust 명령과 같은 사례를 돌려 비교) | 비교를 Rust 단독 단정으로 |
| `tests/lane_env_cli.rs:402` | `python3 -I tools/build_folder.py env`(Rust 명령과 비교) | 비교를 Rust 단독 단정으로 |
| `tests/home_state_plan.rs:195` | 저장소 파일이 아니라 `fcntl.flock` 을 잡는 한 줄짜리 파이썬 | `libc::flock` 을 잡는 자식 프로세스나 스레드로 바꿔요(`libc` 는 이미 의존성) |

### 3-2. 파이썬 파일의 이름이나 내용을 단정하거나 컴파일에 담아요

| Rust 파일 | 무엇을 | 해야 할 일 |
|---|---|---|
| `src/heavy_slot.rs:1318` | 시험이 `include_str!("../tools/heavy_check_slot.py")` 로 파일 내용을 컴파일에 담아요 | 파일을 지우면 빌드가 깨져요. 지우기 전에 이 항목을 빼야 해요 |
| `tests/heavy_slot_cli.rs:1274` | 위 도구의 경로를 만들어요 | 동등성 시험과 함께 정리 |
| `src/gate.rs:270-271` | 보호 경로 목록에 `tools/ci_scope.py`·`tools/check_guard.py` | 명령으로 옮기면 목록이 바뀌고 `tests/protected_paths_one_list.rs` 와 `tools/protected_checks.json`·`tools/review_checklist.json` 이 같이 바뀌어요 |
| `tests/agent_defs_package.rs:980`, `tests/cli_text_citations.rs:64`, `tests/mode_workflow_steps.rs:629` | 스킬 글에 `tools/source_manifest.py …`, `python3 -m unittest tests.ci_scope_test` 문구가 있다는 단정 | 해당 도구·시험을 옮기는 조각이 스킬 글과 이 단정을 같이 바꿔요 |
| `tests/repo_scans_cli.rs`(7곳), `tests/check_affected_cli.rs`(`tests.update_tux_tty_test` 등), `src/check_affected.rs:330` | `python3 -m unittest tests.<이름>` 단계가 계획에 들어간다는 단정과 그 구현 | 조각 2 가 구현과 이 단정을 같이 바꿔요 |
| `tests/claude_python_adapter_removed.rs` | 이미 지운 파이썬 시험 6개가 없다는 단정과 `python3 tests/claude_preview_test.py` 문구 | 그대로 둬도 돼요(없다는 단정이라 파일이 더 사라져도 참) |
| `tests/skill_references_test.py` 때문에 피한 이름: `tests/project_skills_generation.rs:438`, `tests/update_cleanup_cli.rs:243`, `tests/update_user_layer_cli.rs:507` | 파이썬 시험의 읽기 규칙에 걸리지 않게 주석에 이름을 풀어 적음 | 시험이 Rust 로 옮겨가면 주석을 다시 읽어요 |
| `tests/self_audit_cli.rs:4`, `tests/product_cli.rs:4148`, `tests/*terminal_cli.rs` 머리말 | 옮겨온 파이썬 모듈 이름을 설명으로 적음 | 문구만 정리 |

### 3-3. 손대지 않아요 (제품 기능이라 사용자의 파이썬 프로젝트를 다뤄요)

`src/bundle.rs`, `src/bundle/verify.rs`, `src/ops/bundle.rs`, `src/verify.rs`, `src/tdd.rs`, `src/project_detect.rs`, `tests/bundle_*`, `tests/green_test_lock.rs`, `tests/run_verify.rs`, `tests/boundary_repro.rs`, `tests/spec_links_fixture.rs`, `tests/rule_history_guard.rs`, `tests/board_bundle_link_cli.rs`, `tests/project_detect_cli.rs`: 사용자 프로젝트의 `*_test.py`·`unittest` 출력을 읽고 분류하는 제품 동작이에요. 이 저장소의 파이썬 시험과 상관이 없으니 이 작업의 대상이 아니에요.

## 4. 파이썬 파일이 아닌데 파이썬을 쓰는 곳

추적 중인 `.sh`·`.ps1`·`.yml` 35개(`docs`·`.moai`·`references` 아래 제외)에서 `python`·`.py` 를 찾은 75줄을 읽은 결과예요. 조각 2 는 CI 와 check set 에서 파이썬을 걷는 일이라 이 목록도 함께 맡아요.

| 위치 | 하는 일 | 대체 방향 |
|---|---|---|
| `tools/local_check.sh` 의 `unittest` 단계, `python-slots` 폴더 | 파이썬 시험 전체 실행과 전용 슬롯 | 단계와 슬롯 코드를 걷어요(시험이 모두 Rust 가 된 뒤) |
| `.github/workflows/ci.yml:191-196`, `candidate-target.yml:93-97`, `moai-label-sync.yml:38-42` | `setup-python` 과 `PyYAML==6.0.2` 설치 | 파이썬 시험이 사라지면 필요 없어요. `moai-label-sync.yml` 의 인라인 파이썬은 별도 확인(`tests/labels_sync.rs` 가 Rust 쪽에 있어요) |
| `.github/workflows/ci.yml:108-111,183-185,495-497` | `tools/ci_scope.py` 실행과 그 JSON 을 `python3 -c` 로 읽기 | `moai ci-scope` 가 같은 JSON 을 주면 `jq` 나 `moai` 로 읽어요 |
| `.github/workflows/ci.yml:431`, `candidate.yml:124`, `check-guard.yml:81` | `ci_superseded.py`, `candidate_summary.py`, `check_guard.py` 실행 | 해당 `moai` 명령 |
| `.github/workflows/audit-gate.yml:81-82`, `release.yml:107` | `python3 -c` 로 JSON 한 줄 읽기, 연주차 계산 | `moai` 출력이나 셸 `date` 로 |
| `.github/workflows/ci.yml:325`, `candidate-target.yml:200` | 요약 줄에 `python --version` 출력 | 그 줄을 뺀다 |
| `tools/self_check.sh:70-265,342-355` | `--full-python` 선택지, 바뀐 파일 이름에서 `tests/<이름>_test.py` 를 짐작해 `python3 -m unittest` 로 돌리는 `python` 항목들, 파이썬 슬롯 | 단계와 선택 로직을 걷거나 Rust 시험 선택으로 바꿔요 |
| `tools/self_check.sh:351,372` | `python3 -I tools/source_manifest.py check` | 새 `moai` 하위 명령 |
| `tools/smoke/run.sh:13` | `exec python3 -I smoke.py`(CI check 잡이 이 스크립트를 불러요) | `moai smoke` 로 |
| `tools/local-rc1-isolation.sh:37-59` | `python3 -I -c` 로 JSON 필드 8번 읽기 | `moai` 의 JSON 출력을 읽는 방식으로 |
| `scripts/benchmark.sh:42,69,126` | 인라인 파이썬 3곳(측정 집계) | 이 스크립트를 유지할지 결정 D5 |

## 5. 여섯 워크플로 도구를 옮기는 순서 제안

기준은 틀렸을 때 퍼지는 범위, 크기, Rust 쪽에 이미 있는 양이에요. 조각 3 은 한 번에 하나씩 해요.

| 순서 | 도구 | 이유 |
|---|---|---|
| 1 | `mutate.py` | 1034줄에 시험 88개로 가장 커요. 감사의 변형 증거가 이 도구에서 나와서 조용히 약해지면 감사가 약해져요. 동등성 기간을 두고 같은 변형을 양쪽에 돌려 비교해요 |
| 2 | `source_manifest.py` 와 `merge_driver_manifest.py` | 스킬을 고치는 모든 레인이 매번 써요. 병합 드라이버는 클론마다 `python3` 명령으로 등록돼 있어서 이관 뒤 각 클론이 다시 등록해야 해요. 둘이 서로를 불러서 한 쌍으로 갈지 둘로 가를지는 결정 D1 |
| 3 | `smoke/smoke.py` | CI check 잡과 감사자가 써요. 615줄에 시험 30개. 바이너리를 부르는 도구라 `moai smoke` 로 옮기기 자연스러워요 |
| 4 | `build_folder.py` | `env` 는 `moai lane-env` 가 이미 해요. `clean` 만 새로 필요하고 호출처(`local_check.sh`, `self_check.sh`, 스킬 글)를 바꿔요 |
| 5 | `heavy_check_slot.py` | Rust 명령이 이미 있어요. 동등성 시험 2곳과 `include_str!` 한 줄을 정리하면 지울 수 있어요 |

값싼 것부터(5→1) 하면 지울 파일이 빨리 줄어요. 이 카드는 위험이 큰 것부터라고 했으므로 위 순서를 제안하고, 뒤집을지는 결정 D1 에 둬요.

## 6. 결정이 필요한 것 (리더에게 묻는 목록)

- **D1** 조각 3 의 순서와 단위: 위 표대로(위험 큰 것부터)인지, 소스 매니페스트와 병합 드라이버를 한 조각으로 할지.
- **D2** drop 17개를 그대로 지워도 되는지(잃는 보호는 각 줄에 적었어요). 일회성 도구 7개와 그 시험 7개(`audit_eval`, `flow_metrics`, `hook_latency`, `home_entry_count`, `sandbox_write_probe`, `host_live_harness`, `legacy_verdicts`), 수집기 `tests/github_inventory_collect.py`, 실행 도우미 `tests/self_audit_run.py`, Codex 판정 시험 `tests/e2e_host_codex_verdict_test.py`. 필요하면 `moai` 명령으로 새로 만든다는 단서로 지울지.
- **D3** `ci_scope`·`check_guard`·`candidate_summary`·`ci_superseded` 네 개는 CI 단계가 부르는 도구라 "검사 단계가 파이썬을 안 돌린다"(카드 기준 2)를 이루려면 조각 2 에 들어가야 해요. 카드의 조각 3 은 여섯 도구만 말하니, 이 네 개를 조각 2 에 넣거나 따로 조각을 하나 두는 결정.
- **D4** 워크플로·이슈 양식 YAML 점검: 지금 파이썬은 PyYAML 로 파싱해요. Rust 에는 YAML 크레이트가 없어서(`Cargo.lock`, 의존성 최소 원칙) 줄 단위 텍스트 점검으로 옮길지, 크레이트 하나를 들일지.
- **D5** `scripts/benchmark.sh`(인라인 파이썬 3곳)와 `tools/local-rc1-isolation.sh`(8곳)를 유지할지, 고칠지, 지울지.
- **D6** `check_guard` 는 CI 에서 베이스 커밋의 사본을 돌려서 PR 이 자기 점검을 약화하지 못하게 해요. Rust 명령으로 바꾸면 베이스 커밋을 빌드해야 하거나 설치된 릴리스 바이너리를 믿어야 해서 이 신뢰 모델을 정해야 해요.
- **D7** Codex 지원 제거와 겹치는 파이썬 파일(`host_live_harness`, `e2e_host_codex_verdict_test`, `host_matrix_test` 의 Codex 행)을 제거 작업이 먼저 정리하길 기다릴지, 이 작업이 먼저 지울지.

## 7. 지우는 조각이 지켜야 할 제약

- `src/heavy_slot.rs:1318` 의 `include_str!` 때문에 `tools/heavy_check_slot.py` 는 코드 변경 없이 지우면 빌드가 깨져요.
- `tools/ci_scope.py`·`tools/check_guard.py` 등은 `src/gate.rs:270-271` 의 보호 경로예요. 지우거나 바꾸는 PR 은 check-guard 점검 대상이라 그 절차(결정 기록 또는 승인)를 따라야 해요.
- 파이썬 시험을 지우면 `tools/repo_scans.json` 의 규칙 8개와 `tools/local_check.sh` 의 단계 목록 시험(`tests/local_check_test.py`)이 함께 바뀌어야 해요(둘이 같은 변경 안에서 맞아야 해요).

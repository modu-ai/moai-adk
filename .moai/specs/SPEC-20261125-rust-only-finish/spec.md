---
id: SPEC-20261125-rust-only-finish
title: "Python tools and tests removed, slice A1: the drop verdicts"
status: draft
revision: 1
created: 2026-11-25
branch: spec/rust-only-finish-six-20261125
---

# Python tools and tests removed, slice A1: the drop verdicts

## Intent

사용자가 2026-11-24T22:15:56Z 에 "tools나 python 코드 모두 제거하고 rust로 변경해라" 라고 지시했어요(리더 대화, 카드 노트로 기록). 조각 1 의 판정 목록(`.moai/specs/SPEC-20261121-rust-only-inventory/inventory.md`)에서 `drop` 17개를 지우라는 리더 결정 D2 가 있고, 이 조각 A1 은 그중 다른 파일에 묶이지 않은 15개를 지우고 그 파일을 이름으로 부르던 곳을 정리해요. 이 SPEC 은 카드 하나의 조각 여러 개가 같이 쓰는 SPEC 이라 조각마다 Result 에 한 절씩 더해요.

## Non-goals

- `tools/heavy_check_slot.py` 와 `tests/heavy_check_slot_test.py`(covered 2개): Rust 시험에 같은 사례가 없는 것을 옮겨야 지울 수 있어서 조각 A2 가 맡아요.
- `tests/github_inventory_collect.py` 와 `tests/github_inventory_test.py`: 시험 파일 전체가 수집기를 불러 쓰고, 비밀 모양 단정을 받을 Rust 시험이 아직 없어서 GitHub 설정 시험을 옮기는 조각 C 에서 같이 지워요.
- `tests/self_audit_run.py`: 아직 남은 시험 셋(`check_pr_reaudit_test.py`, `check_pr_self_audit_test.py`, `check_pr_zero_selected_test.py`)이 불러 써요. 그 셋을 Rust 로 옮기는 조각 C 에서 같이 지워요.
- `tools/ci_scope.py` 의 분류표: 이 파일은 `src/gate.rs` 의 보호 경로예요. 지운 파일 이름이 표에 남아도 아무 일도 하지 않아서 이 조각은 건드리지 않고 CI 단계를 옮기는 조각이 정리해요.
- 검사 단계(`tools/local_check.sh`, `check affected`, CI)에서 Python 을 걷는 일: 조각 E 예요.

## Requirements

- REQ-1: The repository shall track none of the fifteen Python files of the drop verdicts that no other tracked file depends on, which are listed under Scope.
- REQ-2: The Rust side of the repository shall name none of those fifteen files as a file that exists.
- REQ-3: A document that told a reader to run one of those tools shall say that the tool was removed.
- REQ-4: The Rust tests that `moai check affected --card` selects for the range shall pass.

## Acceptance

- AC-1 (REQ-1): `git ls-files` over the fifteen paths prints nothing · verify: `git ls-files tools/audit_eval.py tools/flow_metrics.py tools/home_entry_count.py tools/hook_latency.py tools/host_live_harness.py tools/legacy_verdicts.py tools/sandbox_write_probe.py tests/audit_eval_test.py tests/e2e_host_codex_verdict_test.py tests/flow_metrics_test.py tests/home_entry_count_test.py tests/hook_latency_test.py tests/host_live_harness_test.py tests/legacy_verdicts_test.py tests/sandbox_write_probe_test.py`
- AC-2 (REQ-2): the allow list of the state folder scan holds no row of a removed file and the scan passes · verify: `cargo test --test state_folder_literal_scan`
- AC-3 (REQ-2): the home isolation scan, the table of repo scans and the check-affected tables still pass without the removed files · verify: `cargo test --test home_isolation_scan --test repo_scans_table --test check_affected_tables_cli`
- AC-4 (REQ-3): the four documents and the eval README carry the removal sentence, and no tracked file outside the records tells a reader to run a removed tool · verify: `git grep -n -E "python3? (-I )?tools/(audit_eval|flow_metrics|home_entry_count|hook_latency|host_live_harness|legacy_verdicts|sandbox_write_probe)" -- . ':!.moai/specs' ':!.moai-cli'`
- AC-5 (REQ-4): the card receipt of the final head is PASS · verify: `moai check affected --card <base>..<head>`

## Scope

- write: `tools/audit_eval.py`, `tools/flow_metrics.py`, `tools/home_entry_count.py`, `tools/hook_latency.py`, `tools/host_live_harness.py`, `tools/legacy_verdicts.py`, `tools/sandbox_write_probe.py` (all deleted)
- write: `tests/audit_eval_test.py`, `tests/e2e_host_codex_verdict_test.py`, `tests/flow_metrics_test.py`, `tests/home_entry_count_test.py`, `tests/hook_latency_test.py`, `tests/host_live_harness_test.py`, `tests/legacy_verdicts_test.py`, `tests/sandbox_write_probe_test.py` (all deleted)
- write: `tests/fixtures/state_folder_literals/allow.tsv`, `tests/home_isolation_scan.rs`, `.gitignore`, `.moai/evals/audit/README.md`
- write: `docs/design-review/appendix/ledger_sandbox_probe.md`, `docs/design-review/flow-control-2026-11-09.md`, `docs/host-live-session-checklist.md`, `docs/rust-product-path.md`
- write: `skills/moai/references/source-manifest.json` (one status sentence, no hash)
- write: `.moai/specs/SPEC-20261125-rust-only-finish/**`
- read-only: `tools/ci_scope.py`, `src/gate.rs`, `.moai/specs/SPEC-20261121-rust-only-inventory/inventory.md`
- forbidden: `.moai-cli/**`, `tools/check_guard.py`, `tools/repo_scans.json`

## Plan

1. 지울 15개를 `git rm` 으로 지워요.
2. 그 이름을 적은 Rust 쪽(허용 목록 세 줄, 머리말 문구)을 고쳐요.
3. 문서 네 곳과 평가 README 의 "그 도구를 돌려라" 부분에 도구가 제거됐다는 문장을 달아요. 예전 기록(SPEC, 측정 문서의 본문)은 고쳐 쓰지 않아요.
4. 영수증은 `moai check affected --card <base>..<head>` 한 번이에요.

## Risks

- 지운 시험이 지키던 보호가 사라져요. 판정 목록의 잃는 보호 칸이 근거이고 이 SPEC 의 Decisions 에 옮겨 적었어요(되돌리기: 이 커밋을 되돌려요).
- 보호 경로 `tools/ci_scope.py` 의 분류표에 지운 파일 이름이 남아요. 분류만 하고 파일을 읽지 않아서 동작이 바뀌지 않아요. 표를 정리하는 조각이 따로 있어요.
- Rust 시험이 지운 파일 이름을 읽는 곳을 놓쳤을 수 있어요. 전체 참조 검색(`git grep`)과 영수증의 선택 시험으로 확인해요.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | drop 17개 중 15개를 이 조각에서 지우고, `github_inventory_collect.py` 와 `self_audit_run.py` 는 그것을 쓰는 시험을 옮기는 조각 C 로 미뤄요 | 지우면 아직 Rust 로 안 옮긴 시험 넷이 깨져요(조용한 보호 손실 금지 규칙). 리더 결정 D2 의 "그대로 지워요" 는 시험을 옮긴 뒤에 지우는 것으로 읽었어요 | 두 파일을 지우는 조각 C 에서 바로잡아요 |
| D-2 | 지워진 도구를 부르는 문서는 고쳐 쓰지 않고 "제거됐다"는 문장을 달아요 | 측정·설계 기록이라 본문을 바꾸면 기록이 거짓이 돼요. 읽는 사람이 실행해 보지 않게 하면 충분해요 | 문장을 지워요 |
| D-3 | `tools/ci_scope.py` 의 분류표는 이 조각이 건드리지 않아요 | 보호 경로라 게이트 변경(tier 2)이 돼요. 분류표의 옛 이름은 동작에 영향이 없어요 | 조각 2(CI 단계 이관)가 표를 새로 써요 |
| D-4 | 사용자의 "모두 제거" 지시와 리더 결정 D2 를 이 조각의 승인 근거로 써요 | 이 작업을 담은 카드의 리더 결정 노트(2026-11-24T22:15:56Z)에 지시 원문과 D1~D7 이 있어요 | 해당 없음 |

## Result

초안이에요. 구현 뒤 이 자리에 조각 A1 의 변경한 파일, 실행한 검사, AC 별 met 와 남은 불일치를 적어요.

### 조각: 문서를 검사하던 Python 시험 3개 (2026-11-26)

승인 근거는 카드의 DELEGATED-APPROVAL 노트(리더, approval_digest `sha256:ec482d0e…d4b9`)예요. 앞의 Requirements·Acceptance·Scope 는 조각 A1 의 것이고, 이 조각은 그 Scope 의 forbidden 에 적힌 `tools/repo_scans.json` 을 바꿔요(지운 시험이 그 표의 규칙 파일이라서예요. 승인 노트도 이 줄 변경을 tier 2 감사 대상으로 적었어요).

| 지운 파일 | 지키던 것 | 대신 지키는 Rust 시험·스캔 |
| --- | --- | --- |
| `tests/skill_references_test.py` | PR 템플릿이 있는 워크플로를 가리켜요 | `skill_references_scan::the_pr_template_names_a_workflow_that_exists` |
| 같은 파일 | 추적 파일의 현재 스킬 파일 참조가 모두 있어요 | `skill_references_scan::every_current_skill_file_reference_resolves`, 스캔 `tracked-skill-references`(검사를 이 Rust 시험으로 바꿈, 조건은 그대로) 와 새 시험을 올린 읽기 스캔(`reads-skills`, `reads-docs`, `reads-shipped-files`) |
| 같은 파일 | 배포 파일이 있는 워크플로만 이름으로 불러요 | `skill_references_scan::shipped_files_name_only_existing_workflows` |
| 같은 파일 | 감지 언어마다 품질 안내 한 줄과 파일 하나, 안내 파일의 길이와 머리말 | `skill_references_scan::every_detected_language_has_a_guide_row`, `every_named_guide_exists_and_every_guide_is_named`, `guides_are_short_and_point_back_to_quality` |
| `tests/skill_web_pages_test.py` | 웹 페이지 규칙 12문장, 두 사본이 같음, 패키지 파일을 부르지 않음 | `skill_text_fixtures::both_web_pages_copies_keep_every_rule_sentence`, `the_two_web_pages_copies_say_the_same`, `neither_web_pages_copy_names_a_package_file` |
| `tests/sync_drift_fixture_test.py` | sync 3단계의 갈래 이름·근거·제안 행동, 대응표의 이름 | `skill_text_fixtures::step_3_names_its_first_three_kinds`, `step_3_documents_the_four_further_kinds_with_ground_and_action`, `the_crosswalk_names_every_further_kind` |

옮기지 않은 것과 달라진 것:

- `sync_drift_fixture_test.py` 의 `classify()` 사례 5개는 버렸어요. 시험 파일 안에서만 정의한 함수를 시험해서 제품 코드를 지키지 않아요. 그 사례가 문서의 제안 행동과 맞춰 보던 것은 3단계 시험의 갈래별 행동 문구 확인이 받아요.
- 같은 파일의 "일곱 갈래" 문장과 3단계의 부록 경로 확인은 옮기지 않았어요. 그 두 문장은 앞선 글 줄이기에서 이미 본문에서 빠졌어요.
- 웹 페이지의 RUN 쪽 사본은 `references/internal/web-pages.md` 로 옮겨져 있어서 그 파일을 읽어요.
- `lua` 는 감지 언어인데 품질 안내가 없어요. `NO_GUIDE_YET` 한 줄로 예외를 두고, 안내가 생기면 시험이 실패해서 목록에서 빼게 했어요.
- 다른 Rust 시험이 임시 프로젝트에 쓰는 경로 8개가 "깨진 참조"로 잡혀요. 지운 시험의 규칙에도 예외가 없었어요. (파일, 경로) 짝 8개만 `FIXTURE_PATHS` 로 허용하고, 짝이 사라지면 시험이 실패해요.

남은 Python: 이 카드의 `check_pr_self_audit_test.py`, `self_check_test.py`, `smoke_harness_test.py`(smoke.py 를 옮긴 뒤)와 기준 2·3 의 나머지예요.

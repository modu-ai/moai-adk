---
id: SPEC-20261117-edit-diagnostics-followups
title: "edit-diagnostics S1 후속: 꺼진 상태 비용 문구, 프로젝트당 잠금 파일 하나, 잘못된 설정 표가 다른 hook 규칙을 끄지 않음"
status: draft
revision: 1
created: 2026-11-17
parent: SPEC-20261117-edit-diagnostics
---

## Intent

edit-diagnostics S1 감사(PR #3179, PASS)의 비차단 지적 세 가지를 고쳐요. (a) 부모 SPEC Result의 꺼진 상태 비용이 한 번의 측정값
5.88 ms로만 적혀 있어 감사 측정(9–21 ms)과 달라요. (b) 잠금 파일이 HOME 상태 폴더 `edit-diagnostics/`에 남아요. (c) 잘못된
`[hooks.edit_diagnostics]` 표가 같은 설정 파일의 다른 hook 규칙을 끄지 않는다는 시험이 없어요(이전 N5).

## Requirements

- REQ-1: The off-path cost shall be stated as the budget and the measured range.
- REQ-2: The edit-diagnostics duty shall keep one lock file per project and reuse it on every run.
- REQ-3: IF `[hooks.edit_diagnostics]` is malformed, THEN tool-preflight shall keep enforcing the default guards and the other rules of the same config file.
  · failure: none

## Acceptance

- AC-1 (REQ-1): 부모 SPEC AC-6 Result가 "50 ms 안(측정 debug 9–21 ms; …)"으로 적혀요.
  verify: `grep -c '50 ms 안(측정 debug 9–21 ms;' .moai/specs/SPEC-20261117-edit-diagnostics/spec.md` → `1`
- AC-2 (REQ-2): 같은 프로젝트에서 여러 번, 여러 파일을 검사해도 잠금 파일은 하나이고, 다른 프로젝트는 하나를 더해요.
  verify: `cargo test --test edit_diagnostics_cli one_lock_file_per_project_is_reused`
- AC-3 (REQ-3): `mode = "block"`인 표와 `[hooks.outside_project_writes]`가 함께 있는 설정에서 tool-preflight가 `.git/` 쓰기와
  프로젝트 밖 쓰기를 거부하고, edit-diagnostics는 꺼져요.
  verify: `cargo test --test edit_diagnostics_cli a_malformed_table_leaves_the_other_guards_on`

## Scope

- write: `tests/edit_diagnostics_cli.rs`, `src/edit_diagnostics.rs`(문서 주석),
  `.moai/specs/SPEC-20261117-edit-diagnostics/spec.md`(AC-6 Result 문구, D21), `.moai/specs/SPEC-20261117-edit-diagnostics-followups/spec.md`
- forbidden: 실제 HOME, `.github/workflows/audit-gate.yml`

## Plan

1. 문구 고치기.
2. 잠금: 프로젝트당 파일 하나를 다시 쓰는 지금 방식을 고정하고 문서로 적기(D1).
3. 시험 2개, 검사, mutation 2개.

## Decisions

- D1 잠금 파일은 지우지 않고 프로젝트당 하나를 다시 써요(파일 이름 = 프로젝트 경로 digest 앞 16자, 빈 파일). 끝난 뒤 지우면,
  다른 실행이 지우기 전의 파일을 열어 둔 채 잠그고 세 번째 실행이 새 파일을 만들어 잠가 둘이 함께 도는 경쟁이 생겨요. 이를 막으려면
  잠근 뒤 경로와 파일이 같은지 다시 확인해야 해서 더 복잡해요. 그래서 수는 사용자가 검사한 프로젝트 수로 묶이고(실행 수가 아님),
  각 파일은 0바이트예요.
- D2 이 가지는 #3179가 아직 병합되지 않아 그 head `4c2401120b` 위에 쌓았어요(리더 지시).

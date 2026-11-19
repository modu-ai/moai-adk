---
id: SPEC-20261119-language-reach-applied-table
title: "PROJECT shows an applied-settings table so nothing is silently dropped (C-87 criterion 4)"
status: draft
revision: 1
created: 2026-11-19
---

# PROJECT applied-settings table (C-87 criterion 4)

## Intent

카드 C-87 기준 4: "PROJECT shows which settings were applied where (a short applied-settings table) so the user can see nothing was silently dropped". 사용자 불만(2026-11-17, `SPEC-20261117-language-reach`): `/moai project` 에서 언어와 각종 설정을 골랐는데 전혀 적용이 안 된 것 같다. 앞 조각(`SPEC-20261117-language-reach`)이 언어 줄(AGENTS.md 관리 구간, 시작·재개·압축 hook 문장, 핵심 규칙 7)을 끝냈고 이 표만 "후속"으로 남겼어요(그 SPEC Non-goals 22행, Result 102~103행).

2026-11-19 tip `4a45b73eea` 에서 읽은 것: `skills/moai/workflows/project.md` 는 단계 4 에서 `project-apply`, 단계 5 에서 관리 구간 갱신, 단계 8 에서 마감 커밋을 안내하지만(카드 C-59), 사용자가 고른 설정이 어디에 적용됐는지 한 번에 보여 주는 단계는 없어요. `Completion Check` 도 표를 요구하지 않아요(`grep -i 'applied\\|dropped' project.md` 가 이 뜻의 문장을 못 찾아요).

카드 기준 1 은 `SPEC-20261117-language-reach` AC-1(met)로 충족돼 있어요. 증거만 카드에 옮기면 돼요(이 SPEC 의 일이 아니에요).

## Non-goals

- 표를 만드는 새 CLI 명령이나 `project-apply` 결과의 새 필드. 이 경로는 사람이 읽는 안내 문장이에요.
- 설정 값의 변경, 기록, 되돌리기. 표는 읽기 전용 보고예요.
- 카드 C-59 의 마감 커밋·프로파일 추적 문장, 다른 단계의 문장.
- MCP 편집(리더 결정, C-59 기준 2).

## Requirements

- REQ-1: When the PROJECT route reaches its final report, it shall show one table with a row for every setting the interview or the profile set. · failure: none
- REQ-2: When a row is shown, it shall name the setting, its value, the file or key it was written to and one state of `applied`, `default`, `HOLD` or `not applied: <reason>`. · failure: none
- REQ-3 (REQ-2): IF a setting the user chose has no place it was written to, THEN the table shall show it as `not applied: <reason>` instead of leaving it out.
- REQ-4: When the Completion Check of the PROJECT route is read, it shall name the table, so a final report without it does not pass. · failure: none
- REQ-5: When `project.md` changes in a commit, the same commit shall carry the manifest line that `python3 -I tools/source_manifest.py write` produces for it. · failure: none
- REQ-6: When the added sentences are in `project.md`, `tests/workflow_size_budget.rs` and `product_cli skill_package` shall still pass. · failure: none

## Acceptance

- AC-1 (REQ-1, REQ-2, REQ-3): `project.md` 에 최종 보고 앞의 표 문장이 있고 행 열(설정·값·쓴 곳·상태), 네 가지 상태 낱말(`applied`·`default`·`HOLD`·`not applied: <reason>`), "빠뜨리지 않는다" 규칙이 이름 붙은 한 파일에서 읽혀요(문구 핀).
  verify: `cargo test --locked --offline --test project_applied_table the_final_report_has_an_applied_settings_table`
- AC-2 (REQ-4): `Completion Check` 가 표를 이름으로 불러요.
  verify: `cargo test --locked --offline --test project_applied_table the_completion_check_names_the_table`
- AC-3 (REQ-5, REQ-6): manifest 줄이 새 `project.md` 와 같고 크기·패키지 검사가 통과해요.
  verify: `python3 -m unittest tests.source_manifest_test`
  verify: `cargo test --locked --offline --test workflow_size_budget`
  verify: `cargo test --locked --offline --test product_cli skill_package`
- AC-4: 이웃 검사(대상 한정; 전량은 PR CI run 이 인용).
  verify: `cargo fmt --check`
  verify: `cargo clippy --locked --offline --all-targets -- -D warnings`
  verify: `python3 -m unittest tests.skill_references_test`

## Scope

- write: `.moai/specs/SPEC-20261119-language-reach-applied-table/**`, `skills/moai/workflows/project.md`, `skills/moai/references/source-manifest.json`, `tests/project_applied_table.rs`
- read-only: `skills/moai/references/internal/**`, `src/ops/project*.rs`, `tests/workflow_size_budget.rs`
- forbidden: 실제 HOME, 네트워크, 다른 단계의 문장 편집, 새 명령·새 JSON 필드

## Plan

1. (지금) 이 SPEC — tip 에서 읽은 사실만, 문구 변경 0.
2. RED: `tests/project_applied_table.rs`(`include_str!` 로 `project.md` 한 파일만 읽음) 가 표 문장과 Completion Check 문장을 찾지 못해 실패하는 것을 확인.
3. GREEN: `project.md` 단계 8 바로 앞(최종 보고 전)에 한 문장 단락, `Completion Check` 에 한 문장 → 같은 커밋에 `source_manifest.py write` 가 쓴 manifest 줄. 막힘: `source-manifest.json` 은 카드 C-112 가 쥐고 있어요(C-118 의 겹침 허용이 병합되면 풀려요).
4. 변이 ≤3(커밋된 파일에서 하나씩): 상태 낱말 하나 삭제, "not applied" 규칙 삭제, Completion Check 문장 삭제 — 시험이 잡아야 해요.
5. fmt, clippy `--all-targets`, 리더의 검사 묶음(등재 검사 시험 전부, `skill_references_test`, `--bins`), trailer-check, tip 과 merge-tree, push, tier 0/1 감사 후 READY.

## Risks

- 표 낱말이 실제 `project-apply` 결과의 이름(`profile`, `agents`, `go-state` …)과 어긋날 수 있어요. 문장은 값의 출처를 "쓴 파일 또는 키"로만 말해서 이름을 고정하지 않아요.
- `project.md` 가 이미 커서(32 KB대) 패키지 파일 한계 65536 바이트와 크기 예산이 있어요. 두 문장만 더해요.
- C-59 가 만든 마감 커밋 단계 8 과 문장이 겹쳐 보일 수 있어요. 표는 그 앞에서 설정을, 마감 커밋은 그 뒤에서 변경 경로를 다뤄요.

## Result

(조각 1 은 계획과 SPEC 만이에요. 조각 2~3 뒤에 채워요.)

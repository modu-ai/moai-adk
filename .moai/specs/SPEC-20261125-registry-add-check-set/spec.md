---
id: SPEC-20261125-registry-add-check-set
title: "스캔 표에 줄만 더한 변경은 영수증을 넓은 세트로 보내지 않아요: 카드 세트에 추가된 대상과 표 테스트만 더해요"
status: draft
revision: 1
created: 2026-11-25
---

# 스캔 표에 줄만 더한 변경은 영수증을 넓은 세트로 보내지 않아요: 카드 세트에 추가된 대상과 표 테스트만 더해요

## Interview

- Q1: 무엇이 영수증을 느리게 하나요? -> A: `moai check affected --card` 가 변경 경로 `tools/repo_scans.json` 을 "audit floor" 넓은 세트 경로로 봐서 전체 세트(183~247 대상)로 떨어져요. `src/check_affected.rs` 의 `wide_class` 가 `gate::TIER2_PREFIXES` 에 든 경로를 모두 넓은 경로로 세는데, 그 목록에 이 파일이 들어 있기 때문이에요. 새 테스트 파일마다 이 파일에 한 줄을 넣어야 하므로 11-25 의 영수증 넷이 이 한 줄 때문에 넓은 세트로 돌았어요(카드 본문: 11-24 이후 착지 114건 중 37건이 이 파일을 건드렸어요).
- Q2: 사용자가 정한 것은요? -> A: 2026-11-25, 리더의 선택 질문 "위험 등급과 승인의 범위를 좁힐까요?" 에 고른 답 "둘 다 (Recommended)" 의 (1)번: 레지스트리 "줄 추가"를 tier 2 바닥과 넓은 세트에서 빼고, 줄 삭제·스캔 삭제·`always` 해제·규칙 완화는 지금처럼 tier 2 로 둬요. 판단은 diff 를 읽어 코드가 하고, 읽을 수 없으면 tier 2 예요.
- Q3: 바닥(tier) 쪽은 이미 되어 있나요? -> A: 네. `gate::registry_only_grew` 와 `gate::tier_of_changes` 가 NeXT 에 있고(`ac9fccef9c`), `moai bundle check`, 착지의 `floor_tier`, `audit-gate` 가 그 규칙을 읽어요. 시험은 `tests/gate_registry_tier.rs` 에 있어요. 그래서 이 카드의 기준 1~3 은 이미 지켜져요. 남은 틈은 `check affected` 의 넓은 세트(기준 4)와 `audit.md` 바닥 목록 글(기준 5)이에요. 같은 바닥 규칙을 다루는 `SPEC-20261124-registry-line-add-tier`(draft, revision 5)는 바꾸지 않아요.
- Q4: `check affected` 는 변경의 글을 읽나요? -> A: 경로만 읽어요. `choose`·`select_card` 는 변경 경로 목록과 head 의 표 글(`Facts::scan_list`)만 받아요. "더하기만 했는가"를 알려면 base 쪽 표 글이 있어야 해요. 새로 읽는 곳은 이것 하나예요.
- Q5: 넓은 세트에서 빼면 무엇이 안 돌 수 있나요? -> A: 표에 더해진 줄이 만드는 영향은 둘이에요. 새 `--test <이름>` 대상이 돌아야 하고(추가된 대상), 표를 읽는 테스트(표 테스트)가 새 줄을 받아들여야 해요. 둘만 카드 세트에 더하면 넓은 세트가 잡던 것을 잃지 않아요. 표를 고치는 다른 방식은 지금처럼 넓은 세트예요.

## Intent

레지스트리 `tools/repo_scans.json` 에 대상이나 규칙 줄을 더하는 변경은 흐름에서 가장 흔한 한 줄 변경이에요. 감사 바닥은 이미 이 변경을 tier 1 로 읽는데(`registry_only_grew`), 영수증 쪽 `check affected --card` 는 아직 경로 이름만 보고 전체 세트로 보내요. 이 SPEC 은 같은 판정 함수를 `check affected` 가 쓰게 해서, 더하기만 한 변경은 넓은 세트를 강제하지 않고 추가된 대상과 표 테스트만 카드 세트에 더하게 해요. 줄 삭제, 스캔 삭제, `always` 해제, 규칙 완화, 읽을 수 없는 글은 지금처럼 넓은 세트와 tier 2 예요. 감사 바닥 코드와 착지·`audit-gate` 판정은 건드리지 않아요. 읽는 사람: 사용자(승인), 레인, 감사자, 착지 담당.

## Non-goals

- 감사 바닥(`gate::TIER2_PREFIXES`, `tier_of_changes`, `registry_only_grew`) 코드와 `moai bundle check`, 착지 `floor_tier`, `audit-gate` 의 판정. 그대로예요(REQ-7). 이 SPEC 은 그 함수를 읽기만 해요.
- 레지스트리 밖의 경로가 넓은 세트가 되는 규칙(`Cargo.toml`, 훅, 원장 저장소 등). 그대로예요(REQ-4).
- 새 스캔이 좋은 스캔인지 판단하는 일. 더하기는 일반 코드 변경처럼 감사해요.
- 맨 `check affected`(전체 세트)와 넓은 세트의 구성. 전체 세트는 그대로 모든 스캔 대상을 돌려요.
- `docs/lane-rules.md`, `CLAUDE.md` 의 레지스트리 설명. 그 파일을 쥔 카드가 고쳐요.

## Requirements

<!-- The change reads one more text (the registry at the range's base) and removes one reason for the full set; the judgment of "only grew" is the existing function, never a second copy. Everything that is not growth keeps today's answer. -->
- REQ-1: When `check affected --card` picks the set for a range whose changed paths include `tools/repo_scans.json` and `gate::registry_only_grew` answers that its change from the range's base only grows it, the card pick shall not count that path as an audit floor path. · failure: none — the pick only decides whether the path forces the full set.
- REQ-2: When the card pick does not count the registry path under REQ-1, it shall add the cargo targets that the new registry text lists and the base text did not, and the test targets whose source names `tools/repo_scans.json`, to the card set. · failure: none — a target the registry names that no test file defines is the existing `table` step's finding.
- REQ-3: IF the change of `tools/repo_scans.json` is not growth, or the base text or the head text cannot be read, or the range has no base, THEN the card pick shall keep the path as an audit floor path and answer the full set. · failure: none — the full set is the safe answer.
- REQ-4: While another changed path of the range is a wide path, the card pick shall answer the full set with that path as its reason. · failure: none — a path rule has no failure path.
- REQ-5: When the card pick judges whether a registry change only grows it, it shall call `gate::registry_only_grew` and no other rule. · failure: none — a pure function.
- REQ-6: When `check affected --list` and a run pick the set for one range, they shall give the same selection. · failure: none — both call one function.
- REQ-7: When a change of any path is judged for the audit floor, `moai bundle check`, the landing check and `audit-gate` shall answer as before this change. · failure: none — this SPEC changes none of them.
- REQ-8: When the audit reference lists the paths that set the tier 2 floor, it shall say that `tools/repo_scans.json` sets it unless the change only grows the registry, and that `check affected --card` reads the same rule for the full set. · failure: none — a document line.

## Acceptance

- AC-1 (REQ-1, REQ-2): 임시 저장소에서 base 이후 head 가 `tools/repo_scans.json` 에 규칙 파일·`--test` 대상·새 스캔을 더하기만 하면 `check affected --card --list` 의 `selection` 이 `card` 이고 넓은 이유가 없으며, 선택에 더해진 대상과 표를 읽는 테스트 대상이 들어 있어요 · verify: `cargo nextest run --locked --offline --test check_affected_registry_add an_added_line_keeps_the_card_set_and_adds_its_targets`
- AC-2 (REQ-3): 줄 삭제, 스캔 삭제, `always` 해제, 규칙 완화(규칙 파일 빼기), 텍스트 필드 고침이 각각 `selection` 이 `full` 이고 이유가 `audit floor` 예요 · verify: `cargo nextest run --locked --offline --test check_affected_registry_add an_edit_or_a_removal_still_gives_the_full_set`
- AC-3 (REQ-3): base 쪽 표가 없거나 JSON 이 아니고, head 쪽 표가 JSON 이 아니고, base 를 읽을 수 없으면 `full` 이에요 · verify: `cargo nextest run --locked --offline --test check_affected_registry_add an_unreadable_registry_gives_the_full_set`
- AC-4 (REQ-4): 같은 범위에서 레지스트리에 줄만 더하고 `Cargo.toml` 도 바꾸면 `full` 이고 이유는 `Cargo.toml` 쪽(`dependencies`)이에요 · verify: `cargo nextest run --locked --offline --test check_affected_registry_add another_wide_path_still_gives_the_full_set`
- AC-5 (REQ-5, REQ-6): `--list` 와 실제 실행의 `selection` 이 같은 범위에서 같아요; `check_affected.rs` 에 `registry_only_grew` 호출이 하나이고 판정 규칙의 사본이 없어요 · verify: `cargo nextest run --locked --offline --test check_affected_registry_add the_list_and_a_run_pick_the_same_set`
  manual: `grep -c "registry_only_grew" src/check_affected.rs` prints 1 or 2 (a call and its doc).
- AC-6 (REQ-7): `registry_only_grew` 와 바닥 판정 시험이 그대로 통과해요 · verify: `cargo nextest run --locked --offline --test gate_registry_tier`
- AC-7 (REQ-4, REQ-7): 옛 `check affected --card` 시험이 그대로 통과해요(`choose` 의 옛 서명을 쓰는 호출은 base 글이 없다는 뜻이라 지금처럼 `full`) · verify: `cargo nextest run --locked --offline --test check_affected_card_select --test check_affected_scans_cli --test check_affected_tables_cli`
- AC-8 (REQ-8): `skills/moai/references/audit.md` 의 바닥 목록 글이 레지스트리의 더하기 예외와 `check affected --card` 의 같은 규칙을 영어로 말해요 · verify: `grep -c "registry_only_grew" skills/moai/references/audit.md` prints 1 or more.

## Scope

- write: `.moai/specs/SPEC-20261125-registry-add-check-set/**`, `src/check_affected.rs`, `tests/check_affected_registry_add.rs`(새 파일), `skills/moai/references/audit.md`
- read-only: `src/gate.rs`, `src/bundle.rs`, `src/card/land/check.rs`, `src/audit_gate.rs`, `tools/repo_scans.json`, `tests/gate_registry_tier.rs`
- forbidden: 감사 바닥 목록과 판정 코드, `tools/repo_scans.json` 의 변경(이 카드는 레지스트리에 줄을 더하지 않아요; 새 테스트 파일은 임시 저장소만 읽어요), `CLAUDE.md`, `docs/lane-rules.md`, 이 저장소의 `.moai/project.yaml`, 실제 HOME
- 주의: `src/check_affected.rs` 는 다른 카드(영역 세트와 모듈 필터)가 쥐고 있어요. 그 카드가 끝나거나 반납한 뒤 구현해요.

## Plan

1. SPEC 초안을 커밋하고 digest 를 리더에게 올려요. 승인은 기록하지 않아요. 승인과 `src/check_affected.rs` 를 쥔 카드의 반납이 있기 전에는 구현 코드를 쓰지 않아요(RED 테스트까지만).
2. RED: `tests/check_affected_registry_add.rs` 에 AC-1~AC-5 를 써요(임시 git 저장소, 진짜 `moai check affected --card --list` 호출). 지금 코드는 `audit floor` 로 `full` 을 답하므로 AC-1 이 올바른 이유로 실패해요.
3. GREEN: `Facts` 에 base 쪽 표 글(`scan_list_base`, `git show <base>:tools/repo_scans.json`, 일반 파일일 때만)을 읽어 두고, `select_card` 가 `wide_class` 로 넓은 경로를 모으기 전에 레지스트리 경로가 `gate::registry_only_grew(path, base, head)` 이면 그 경로를 빼며 추가된 대상(`scan_entries` 의 새 대상 − 옛 대상)과 `tools/repo_scans.json` 을 글에 담은 테스트 대상을 카드 세트에 더해요. `choose` 의 서명은 그대로 두고(옛 호출은 base 글 없음 → `full`) `choose_with_base` 를 더해 `list` 와 실행이 쓰게 해요.
4. 확인: 새 테스트와 이웃 테스트, `cargo fmt --all`, 새 가드 테스트마다 가드 뮤턴트 한 줄(`registry_only_grew` 호출을 항상 참/거짓으로, 추가 대상 더하기 삭제; 셸 `sed` 정확히 한 번, `git checkout --` 로 복원, Python 없음), `moai check affected --card` 를 READY 에서 한 번. 위험 경로 실행: 임시 저장소에서 더하기만 한 범위와 삭제한 범위를 `--card` 로 실행해 `selection` 을 읽어요.
5. SYNC: `audit.md` 바닥 목록 글을 영어로 고치고 Result 에 AC 별 met·not_run 을 적어요. tier 2(감사 하한과 영수증 선택이 바뀌어요)이므로 이 변경을 쓰지 않은 세션이 전체 감사로 맡아요.

## Risks

- 넓은 세트가 잡던 것을 잃을 수 있어요: 표에 더해진 줄이 만드는 영향은 새 대상과 표 테스트라고 보고 그 둘만 더해요(D-2). 표를 고치는 다른 방식과 읽을 수 없는 글은 넓은 세트로 남아요. 착지의 넓은 세트(배치 검사)가 한 번 더 잡아요.
- base 쪽 표 글을 읽는 git 호출이 실패하면 넓은 세트로 돌아가요. 지금보다 나빠지지 않아요.
- 판정을 두 군데에 두면 어긋나요: 같은 함수 `registry_only_grew` 를 부르고 사본을 두지 않아요(REQ-5).
- 다른 카드가 쥔 `src/check_affected.rs` 와 겹쳐요: 그 카드가 끝난 tip 위에서 구현해서 머지 충돌을 피해요.

## Decisions

승인 전 제안이에요. 사용자나 위임받은 리더가 고르면 revision 을 올려 `approval` 을 기록해요(레인은 승인을 기록하지 않아요). 기본값은 권장안이에요.

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | "더하기만 했는가"의 기준 글은 범위의 base(`facts.base`)에서 읽어요(기본). 대안: 통합 tip 의 글과 비교 | 바닥 판정도 merge base 와 head 의 글을 비교해요(`bundle check`, 착지, `audit-gate`). 같은 두 글이면 세 판정과 영수증이 같은 답을 내요 | 읽는 개정 한 곳을 바꿔요 |
| D-2 | 더하기만 한 변경이면 추가된 대상과 `tools/repo_scans.json` 을 글에 담은 테스트 대상(표 테스트)만 카드 세트에 더해요(기본). 대안: 추가된 대상만 더해요 | 표 테스트는 새 줄을 받아들이는지 보는 시험이라 빼면 표 형식 오류를 놓쳐요. 대안은 더 빠르지만 그 시험을 잃어요 | 표 테스트 더하기를 빼요 |
| D-3 | 읽을 수 없는 글, base 없음, 더하기 아닌 변경은 지금처럼 `full`·`audit floor` 예요(기본). 대안: 읽을 수 없는 글은 카드 세트로 | 관문을 약하게 하지 않는 선이에요. 사용자 결정도 "읽을 수 없으면 tier 2" 예요 | 해당 없음(약화가 되므로 되돌리지 않아요) |
| D-4 | `choose` 의 서명을 바꾸지 않고 `choose_with_base` 를 더해요(기본). 대안: `choose` 서명에 base 글 매개변수를 더한다 | `choose` 를 직접 부르는 옛 시험(다른 카드가 쥔 파일)이 깨지지 않아요. 옛 호출은 base 글이 없으니 `full` 로 안전해요 | 두 함수를 합쳐요 |

## Result

- AC-1 met: `check_affected_registry_add::an_added_line_keeps_the_card_set_and_adds_its_targets` (규칙 파일·`--test`·새 스캔을 더한 범위가 `selection: card`, `card_wide: null`; 선택에 새 대상 `scan_beta_three`, `scan_gamma` 와 표 테스트 `repo_scans_table` 이 있고 옛 대상과 `card_other_*` 는 없어요).
- AC-2 met: `an_edit_or_a_removal_still_gives_the_full_set` (스캔 삭제, 점검 줄 삭제·교체, `always` 해제, 규칙 파일 삭제, 범위 좁힘, 텍스트 필드 고침 일곱 경우가 모두 `full` 과 `audit floor`, 경로에 `tools/repo_scans.json`).
- AC-3 met: `an_unreadable_registry_gives_the_full_set` (head 글이 JSON 아님, base 글이 JSON 아님, base 에 표 파일이 없음 셋이 모두 `full`).
- AC-4 met: `another_wide_path_still_gives_the_full_set` (더하기와 `Cargo.toml` 변경이 `full`, 이유 `dependencies`, 넓은 경로 목록은 `["Cargo.toml"]` 이라 레지스트리를 담지 않아요).
- AC-5 met: `the_list_and_a_run_pick_the_same_set` (`--list` 와 가짜 `cargo` 실행이 더하기에서 `card`(영수증 `selection=card`), 편집에서 `full`(영수증에 `selection=` 없음)). 수동: `grep -c "registry_only_grew" src/check_affected.rs` 는 2 예요(문서 한 줄과 호출 한 곳); 판정 규칙의 사본은 없어요.
- AC-6 met: `gate_registry_tier` 전체가 통과해요(바닥 코드는 바꾸지 않았어요).
- AC-7 met: `check_affected_card_select`, `check_affected_scans_cli`, `check_affected_tables_cli` 가 그대로 통과해요. `choose` 의 옛 서명은 레지스트리 글이 없다는 뜻이라 레지스트리를 바꾸는 범위에 `full` 을 답해요; 그 서명을 직접 부르는 시험은 없어서 이 줄은 코드로만 읽어요(`choose` → `choose_with_registry(…, None, …)`).
- AC-8 met: `audit.md` 바닥 목록 문단 끝에 영어 문장 두 개를 더했어요(레지스트리는 더하기만이면 바닥을 세우지 않고 `gate::registry_only_grew` 가 규칙이며, `check affected --card` 가 같은 규칙을 읽어 더하기는 카드 세트, 그 밖은 전체 세트). `grep -c "registry_only_grew" skills/moai/references/audit.md` 가 1 이에요. `source-manifest.json` 의 `references/audit.md` sha256 도 다시 계산했어요(`f278092f…d7b9`). `audit_text_facts`, `audit_scratch_rule`, `core_rules_2876`, `agent_defs_package`, `next_flow_skill_scan`, `bundle_check_cli` 를 포함한 139개가 통과했어요.
- 범위 목록에 없던 파일: `tests/check_affected_card_areas.rs`. 영역 세트 SPEC 의 한 시험(`narrowed_floor_paths_take_their_area_and_the_scan_table_pins_its_tests`)이 "베이스에 표가 없다가 생긴 표"를 카드 세트로 고정했어요. 그 변경은 이 SPEC 의 REQ-3 으로 `full` 이 되므로, 픽스처 베이스에 표를 두고 변경을 "표에 스캔 한 개 더하기"로 바꿨어요. 예전 경우(표가 새로 생김)가 `full` 로 가는 것은 `an_unreadable_registry_gives_the_full_set` 의 셋째 경우가 지켜요.
- 읽는 곳: 옛 글은 `facts.base` 가 아니라 `git merge-base <base> <head>` 에서 읽어요. `changed` 가 `base...head` 로 재어지는 쪽과 같고, base 가 head 의 조상이면 둘은 같아요(D-1 의 뜻은 그대로예요).
- 확인하지 않은 것: 병합 base 가 head 의 조상이 아닌 범위에서의 실제 실행(시험은 조상인 범위만 써요).

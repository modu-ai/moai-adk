---
id: SPEC-20261125-land-closes-audited-merge
title: "카드 착지가 감사한 head 를 tip 위에 머지한 커밋도 닫아요"
status: draft
revision: 1
created: 2026-11-25
---

# 카드 착지가 감사한 head 를 tip 위에 머지한 커밋도 닫아요

## Interview

- Q1: 무엇이 깨졌나요? -> A: "tip 이 움직인 뒤 착지 담당이 감사한 head H 를 tip 위에 머지한 M 을 `card land` 하면 `no_pass`(\"a verdict names one exact head\")로 거절돼요. 오늘 시범 카드 다섯 장이 모두 `--keep-open` 착지나 손으로 옮긴 뒤 `card finish --head H` 두 단계로 돌아갔어요."(리더 배차 2026-11-24T23:25:41Z, 사용자 지시 아래의 흐름 재설계)
- Q2: 어떻게 되길 바라나요? -> A: "M 의 둘째 부모가 PASS 받은 H 이고 첫째 부모가 그때의 tip 이면 그 PASS 가 M 의 닫는 착지를 덮어요. 조건은 엄격하게: M 이 정확히 부모 둘, 부모 2 가 PASS head, 머지가 충돌 없이 났거나(M 의 트리가 `git merge-tree` 결과와 같음) — 충돌을 풀었으면 지금처럼 거절하고 그 차이를 감사자에게 보이는 길을 안내해요. 게이트를 약하게 만들지 않는 것이 조건이라 tier 2 예요."(같은 배차)

## Intent

착지 담당이 감사받은 head 를 새 tip 위에 머지해서 착지하는 일은 흐름의 보통 길이에요(tip 은 늘 움직여요). 그런데 지금 `card land` 는 머지 커밋 M 에 판정이 없다는 이유로 닫는 착지를 거절해서, 매번 두 단계(`--keep-open` 착지 뒤 `card finish`)로 돌아가요. 묶음 착지(`land-batch`)는 이미 같은 일을 하는데, 그 길은 안에서 `audited` head 를 넘겨 판정을 H 에서 읽어요. 이 SPEC 은 단독 `card land` 가 M 에서 같은 근거를 스스로 찾게 해서, 깨끗한 머지 한 건은 한 번의 호출로 닫고 착지하게 해요. 판정이 덮는 범위는 늘리지 않아요: H 의 PASS 가 모든 검사를 지금과 똑같이 통과해야 하고, M 이 H 와 tip 의 깨끗한 머지 말고 다른 것을 담고 있으면 거절해요. 읽는 사람: 리더(착지 담당), 감사자.

## Non-goals

- 충돌을 푼 머지나 손으로 고친 머지를 H 의 PASS 로 덮지 않아요. 그 M 은 감사자가 M 자체를 보고 판정해야 해요(REQ-4).
- 판정 규칙(독립성, tier 바닥, 영수증, 늦은 FAIL)과 저장소 스캔 규칙은 바꾸지 않아요. 판정을 읽는 head 만 H 로 바뀌어요.
- `card land-batch` 의 `audited` 경로, 체크아웃된 브랜치 이동(`worktree_follow`), 작업 수 상한은 이 SPEC 이 아니에요.
- 부모가 둘이 아닌 머지(문어발 머지)와 빨리감기(부모가 하나)는 다루지 않아요. 빨리감기 head 는 감사한 head 자신이에요.
- 새 옵션을 만들지 않아요(Decisions D-1).

## Requirements

<!-- This SPEC adds no new blocking check; it lets one more shape of head reach the verdict check that already blocks, with every refusal of that check kept. -->
- REQ-1: When `card land` is given a merge commit with exactly two parents whose second parent is named by a verdict event of the card, that no verdict event of the card names, and whose tree is the tree `git merge-tree --write-tree` makes of its two parents, the landing shall read the verdicts of that second parent in place of the merge commit's own. · failure: none — a head that does not meet all four conditions is judged on its own verdicts as before (REQ-5).
- REQ-2: When the landing relies on the verdict of the second parent, it shall apply to that verdict every check it applies to a verdict on a landing head: the reviewer is none of the authors of the range from the tip to the second parent and not the lander and held no reservation of the card, the tier is the higher of the stated tier and the floor of the changed paths, a receipt covers tier 1 and 2, no newer FAIL of the same reviewer stands, and the repository scans the merge tree needs are covered by a receipt of the second parent. · failure: none — every check that fails refuses before the first write.
- REQ-3: When a landing that relies on the verdict of the second parent moves the integration branch, the landing shall require the first parent of the merge commit to be the integration tip it read. · failure: none — a first parent that is not the tip refuses before the first write.
- REQ-4 (REQ-1): IF a merge commit has exactly two parents and its second parent is named by a verdict event of the card but its tree is not the clean merge of its parents, THEN `card land` shall refuse, writing nothing, with a message that begins `no_pass:`, names the paths in which the commit differs from the clean merge (or says that the merge of the parents conflicts) and says that the auditor must record a verdict on the merge commit itself. · failure: none — the refusal is the safe behavior.
- REQ-5 (REQ-1): IF a merge commit has not exactly two parents, or no verdict event of the card names its second parent, THEN `card land` shall refuse as `no_pass` as it does today and write nothing. · failure: none — nothing is relied on.
- REQ-6: When a verdict event of the card names the landing head itself, `card land` shall judge the verdicts of that head and not read its second parent. · failure: none — an own verdict, a FAIL included, decides.
- REQ-7: When the landing relies on the verdict of the second parent, the answer shall name that head as `audited_head` and carry it as `PASS head <12 hex>` in the evidence line of each criterion the landing met. · failure: none — the answer is derived from the head the check read.

## Acceptance

- AC-1 (REQ-1, REQ-7): 첫째 부모가 tip 이고 둘째 부모가 PASS 받은 H 인, 충돌 없는 머지 M 을 `card land --by leader` 하면 한 번의 호출로 브랜치가 M 으로 움직이고 카드가 done 이며 열린 기준이 없고, 답의 `audited_head` 가 H 이고 기준 근거가 `verdict event <n> PASS head <H 의 12자리>` 예요 · verify: `cargo test --locked --offline --test card_land_audited_merge_cli a_clean_merge_of_the_audited_head_closes_the_card_in_one_call`
- AC-2 (REQ-2): H 의 PASS 가 막히는 모든 경우 — 감사자가 H 의 작성자임, 감사자가 착지하는 세션임, tier 1 인데 영수증 없음, 새 FAIL 이 섰음, 필요한 스캔을 덮는 영수증이 없음 — 에서 M 의 착지가 같은 코드로 거절되고 브랜치와 원장이 그대로예요 · verify: `cargo test --locked --offline --test card_land_audited_merge_cli every_check_of_the_audited_verdict_still_refuses_the_merge`
- AC-3 (REQ-3): 첫째 부모가 tip 이 아닌 M(tip 이 M 을 만든 뒤 또 움직였음)은 착지가 거절되고 브랜치와 원장이 그대로예요 · verify: `cargo test --locked --offline --test card_land_audited_merge_cli a_merge_whose_first_parent_is_not_the_tip_is_refused`
- AC-4 (REQ-4): 충돌을 풀어 만든 M 과 머지 뒤 손으로 한 줄을 고친 M 이 각각 `no_pass:` 로 시작하는 문구로 거절되고, 문구가 M 이 깨끗한 머지와 달라지는 경로(충돌이면 충돌했다는 말)와 "M 자체에 `card verdict` 를 남겨요" 안내를 담고, 브랜치와 원장이 그대로예요 · verify: `cargo test --locked --offline --test card_land_audited_merge_cli a_resolved_or_edited_merge_is_refused_with_the_paths_the_auditor_must_read`
- AC-5 (REQ-5): 부모가 셋인 머지와, 둘째 부모에 판정이 없는 머지가 지금처럼 `no_pass` 로 거절되고 아무것도 쓰이지 않아요 · verify: `cargo test --locked --offline --test card_land_audited_merge_cli a_merge_without_two_parents_or_without_a_verdict_on_the_second_is_refused_as_before`
- AC-6 (REQ-6): M 자신에게 FAIL 판정이 있으면 H 의 PASS 가 있어도 거절되고, M 자신에게 PASS 가 있으면 H 를 읽지 않고 그 PASS 로 착지해요 · verify: `cargo test --locked --offline --test card_land_audited_merge_cli a_verdict_on_the_merge_itself_decides_and_the_second_parent_is_not_read`
- AC-7: 이웃 테스트가 그대로 통과해요(단독 착지, 묶음 착지, 체크아웃된 브랜치 이동) · verify: `cargo test --locked --offline --test card_land_cli` and `cargo test --locked --offline --test card_land_batch_cli` and `cargo test --locked --offline --test card_land_checked_out_cli`

## Scope

- write: `.moai/specs/SPEC-20261125-land-closes-audited-merge/**`, `src/card/land.rs`, `src/card/land/check.rs`, `tests/card/card_land_audited_merge_cli.rs`
- shared (line lease, 도움말 문장만): `src/ops/card.rs`(`Land` 의 문서 주석), 그 도움말을 담은 골든 파일이 있으면 그 파일
- read-only: `src/card/land/batch.rs`(묶음 착지가 `audited` 를 넘기는 방식의 본보기), `src/card/land/scans.rs`, `src/ledger/workflow.rs`
- forbidden: 판정 규칙의 약화(독립성·tier·영수증·FAIL), `card land-batch` 의 동작, `CLAUDE.md`, `skills/**`, 이 저장소의 `.moai/project.yaml`, 실제 HOME

## Plan

1. SPEC 초안을 커밋하고 승인을 `LANE-BLOCKED kind=decision` 으로 올려요. 레인은 승인을 기록하지 않아요. 승인 전에는 코드를 쓰지 않아요.
2. RED: `tests/card/card_land_audited_merge_cli.rs` 에 AC-1 ~ AC-6 을 써요. 임시 저장소와 임시 HOME 에서 진짜 `git merge` 로 M 을 만들고 `card verdict`, `card land` 를 그대로 불러요. 지금 코드는 `no_pass` 로 거절하는 것이 옳은 이유의 실패예요(AC-1).
3. GREEN: `land_inner` 에서, 요청에 `audited` 가 없을 때 M 의 부모와 판정 이벤트를 읽어 `audited_refusal` 과 같은 조건으로 둘째 부모를 `verdict_head` 로 골라요. 조건 확인은 새 함수 하나로 두고, 깨끗한 머지 판별(`is_clean_join`)은 그대로 써요. 깨끗하지 않을 때 차이 경로를 읽는 도우미를 더하고 `no_pass_reason` 에 안내를 더해요. 판정 검사와 스캔 검사는 한 줄도 바꾸지 않아요.
4. 확인: 새 테스트와 AC-7 의 이웃 테스트, `cargo fmt --all`, 새 가드 테스트마다 가드 뮤턴트 한 줄(Python 을 쓰지 않아요: 셸과 awk 로 정확히 한 번 치환하고 `git checkout --` 로 되돌려요), 위험 경로는 임시 저장소에서 진짜 착지를 돌린 AC-1·AC-3·AC-4 가 해당해요, `moai check affected --card` 를 READY 에서 한 번.
5. SYNC: Result 에 AC 별 met·not_run 을 적어요. tier 2 이므로 감사는 이 변경을 쓰지 않은 세션이 전체 감사로 맡아요.

## Risks

- 게이트 약화: 가장 중요한 위험이에요. 막는 것은 네 군데예요. 판정은 H 의 PASS 가 지금과 같은 검사(REQ-2)를 모두 지나야 하고, M 은 H 와 tip 의 깨끗한 머지 말고 다른 것을 담을 수 없고(REQ-1, REQ-4), 첫째 부모가 읽은 tip 이어야 하고(REQ-3), 판정이 덮는 head 가 답에 이름으로 남아요(REQ-7). 깨끗한 머지가 의미로는 깨질 수 있다는 한계는 묶음 착지의 `audited` 경로와 같아요: 그 길은 합친 트리에 넓은 검사를 돌려 보완하고, 이 길은 필요한 저장소 스캔을 H 의 영수증이 덮어야 한다는 지금의 규칙(REQ-2)이 보완해요. 이 SPEC 이 넓은 검사를 더하지는 않아요.
- 판정을 읽는 head 의 바뀜: 닫는 근거(`verdict event n PASS head …`)가 착지 head M 이 아니라 H 를 가리켜요. 읽는 쪽(`late_fail`, 늦은 FAIL 확인)은 같은 `verdict_head` 를 쓰므로 H 의 늦은 FAIL 을 봐요. M 이 이미 브랜치에 있는 호출(손으로 옮긴 뒤 닫기)에도 같은 규칙이 서고, 그때 첫째 부모 규칙은 브랜치를 움직이는 착지에만 적용돼요(지금 `audited` 와 같아요).
- git 의 결정성: `git merge-tree --write-tree` 가 읽히지 않거나 실패하면 깨끗한 머지가 아니에요(거절). 같은 git 으로 만든 머지의 트리와 한 번 더 계산한 결과가 다르면 거절이에요(안전한 쪽).
- 오래된 판정의 base: 감사자가 본 base 와 지금 tip 이 달라도 H 의 PASS 를 읽어요. 지금의 묶음 착지와 같고, 그 사이 tip 이 바꾼 것은 H 가 바꾼 것이 아니므로 판정 범위(tip..H)에 들어오지 않아요.

## Decisions

승인 전 제안이에요. 사용자나 위임받은 리더가 고르면 revision 을 올려 `approval` 을 기록해요(레인은 승인을 기록하지 않아요).

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | 새 옵션 없이 `card land` 가 M 의 부모와 판정 이벤트에서 H 를 스스로 찾아요 | 리더 배차가 "그 PASS 가 M 의 닫는 착지를 덮어요"라고 했어요. 옵션(`--audited H`)은 착지 담당이 임의의 head 를 대는 길을 열고, 지금 `audited` 는 안에서만 쓰여요. 조건이 닫혀 있어(REQ-1) 옵션 없이도 H 는 하나로 정해져요 | 찾는 함수 한 곳을 지우면 옛 거절로 돌아가요 |
| D-2 | M 자신에게 판정이 있으면 그것이 이겨요(REQ-6) | M 에 FAIL 이 섰는데 부모의 PASS 로 덮으면 감사자의 반대를 지워요 | 한 조건을 지워요 |
| D-3 | 깨끗하지 않은 머지의 거절은 `no_pass:` 머리를 그대로 두고 안내만 더해요 | 지금 이 접두를 읽는 도구와 노트가 있어요 | 안내 문장을 지워요 |
| D-4 | 찾는 조건에 둘째 부모의 판정이 PASS 인지는 넣지 않고 "판정 이벤트가 있다"만 봐요 | PASS·FAIL 판단은 판정 검사 한 곳에만 두어야 규칙이 둘로 갈라지지 않아요(FAIL 은 `verdict_fail` 로 거절돼요) | 조건을 하나 더해요 |
| D-5 | 부모가 둘이 아닌 머지와 빨리감기는 다루지 않아요 | 이번에 터진 모양은 부모 둘의 머지예요. 문어발 머지에서 "둘째"는 뜻이 없어요 | 뒤 SPEC 으로 넓혀요 |

## Result

(구현 뒤에 적어요.)

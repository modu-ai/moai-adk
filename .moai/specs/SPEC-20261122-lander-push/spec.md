---
id: SPEC-20261122-lander-push
title: "the lander copies the integration branch to the remote after every landing, fast-forward only, and a lost copy is retried by the next run"
status: draft
revision: 1
created: 2026-11-22
parent: SPEC-20261122-remote-copy-slim
---

# the lander copies the integration branch to the remote after every landing, fast-forward only, and a lost copy is retried by the next run

## Intent

흐름 재설계 조각 2 의 첫 코드 조각. 착지된 부모 SPEC-20261122-remote-copy-slim 의 REQ-1 을 코드로 만들고, 리더가 2026-11-22 에 조각 A 의 승인 기준으로 받은 동료 제안을 그대로 승인 줄로 삼는다: 임시 로컬 bare 저장소 하나를 원격으로 두고 (a) 마지막 착지는 원격에 닿지 못한 채 끝나고 프로세스가 끝나도, 새 실행 하나가 다른 착지를 기다리지 않고 그 커밋을 fast-forward 로 민다 · (b) 원격이 이미 같은 커밋이나 그 자손을 가지고 있고 답만 잃었으면 성공으로 정산한다 · (c) 역사가 갈라졌으면 원격과 로컬은 그대로 두고 이유를 보류로 기록한다 · (d) 로컬 완료는 한 번만 기록된다.

측정(2026-11-22, 이 개발기, git 2.54.0): `git push --porcelain` 은 git-push(1) 의 OUTPUT 절 그대로 답한다 — 성공은 빈칸(앞깔림)·`*`(새 분기)·`=`(이미 같음) 플래그와 exit 0, 거절은 `! refs/heads/… [rejected] (non-fast-forward)` 과 exit 1, 원격 도달 불가는 porcelain 상태 줄 없이 `fatal:` 과 exit 128. PUSH RULES 절은 분기로의 push 가 강제 없이는 fast-forward 뿐임을 정한다. 이 조각의 전제 사실: 착지 명령은 통합 팁에 있고 참조 가드는 아직 착지 전이라 이 조각은 그것 없이 돌아간다 — push 는 로컬 `refs/heads` 를 움직이지 않으므로 가드와 만나지 않는다. 워크플로 파일은 건드리지 않는다(다음 조각, 사용자 선택 전 병합 금지).

## Non-goals

- 슬림 원격 워크플로와 그 lint 시험, 모든 워크플로 파일 변경: 조각 B, 병합 전 사용자의 선택.
- 원격 실행 reader 와 수리 카드: 조각 C.
- 릴리스 분기 철자의 정리와 통합 브랜치 이름의 프로필 한 곳 규칙: 조각 D(부모 REQ-7·REQ-8).
- 통합 브랜치의 실제 이름 값(사용자 결정, 2026-11-22)의 적용: 이름 한 값을 바꾸는 일은 첫 본격 복사와 함께 리더가 사용자 지시로 한다. 이 조각의 코드는 프로필에서 이름을 읽으므로 어떤 값에서도 돌아간다.
- push 주위의 시간 제한과 원격 워크플로 실행 결과 읽기: 이후 조각.
- 뒤처진 실제 원격 분기의 첫 복사: 부모 SPEC 이 사용자 결정으로 남긴 일. 이 조각은 실제 원격을 향해 아무것도 push 하지 않는다.

## Requirements

- REQ-1: When a landing has closed its transaction, the lander shall copy the integration branch to its remote with one fast-forward-only push and report the copy state in the landing answer. · failure: none — a refused, failed or unreachable copy leaves the landing, every table and every reference unchanged, is reported in the landing answer, and never fails the landing command.
- REQ-2: When the copy runs, the copy shall classify the remote position first and push only when the remote ref is behind the local tip or absent, never with a force flag. · failure: none — a remote that already holds the same commit or a descendant is settled as success with no push, and a remote that shares no ancestry with the local tip is recorded as a hold with nothing pushed.
- REQ-3: When the push answer is read, the copy shall read git's documented push output for the one ref to tell a rejection from a remote failure from an unreachable remote. · failure: none — a rejection re-reads the remote once and settles or holds, a remote failure records a pending copy whose outcome the next run settles, and an unreachable remote records a pending copy.
- REQ-4: When a copy has not finished, the next landing command and the on-demand command shall push the integration branch again without waiting for another landing. · failure: none — no resident loop runs, no new card is made, and a new process pushes the recorded commit fast-forward.
- REQ-5: When a copy ends in any state, the copy shall write one record of its state in the common git folder by a temporary file and a rename, beside the one landing record the close already wrote. · failure: none — retries write no second record and no second landing event.
- REQ-6: When the on-demand command runs, the command shall answer the copy state as an action of the card command and write no ledger row. · failure: none — a hold is an answer with a record, not an error exit and not a card.

## Acceptance

- AC-1 (REQ-1, REQ-2): a landing whose remote is behind pushes the branch fast-forward (the remote branch equals the tip and the answer says done and pushed), and a landing whose remote shares no ancestry completes, closes the card, moves the branch, reports the refused push in the landing answer, and leaves the remote as it was.
  verify: `cargo test --locked --offline --test card_remote_copy_cli card_land_pushes_fast_forward_only_and_reports_a_refused_push` — red first
- AC-2 (REQ-3, REQ-4): the last landing completes with the remote unreachable, the process ends, and a new on-demand run pushes that commit fast-forward without another landing.
  verify: `cargo test --locked --offline --test card_remote_copy_cli an_offline_landing_completes_and_a_new_run_pushes_the_commit` — red first
- AC-3 (REQ-2, REQ-3): when the remote already holds the same commit or a descendant and only the answer was lost, the copy is settled as success with no push.
  verify: `cargo test --locked --offline --test card_remote_copy_cli a_lost_answer_is_settled_as_success` — red first
- AC-4 (REQ-2): when the histories diverged, remote and local stay as they are and the reason is recorded as a hold.
  verify: `cargo test --locked --offline --test card_remote_copy_cli a_diverged_remote_is_recorded_as_a_hold_and_nothing_moves` — red first
- AC-5 (REQ-5): the local completion is recorded once — after the offline landing and its retry the ledger holds one landing event and one close for the card, and the copy state is one record.
  verify: `cargo test --locked --offline --test card_remote_copy_cli the_local_completion_is_recorded_once` — red first
- AC-6 (REQ-2, REQ-6): a remote without the branch takes the copy as a creation that destroys nothing, and a profile with no integration branch answers pending with the reason and no push.
  verify: `cargo test --locked --offline --test card_remote_copy_cli an_absent_remote_branch_is_created_and_a_profile_without_a_branch_is_reported` — red first
- AC-7: the wider READY set passes.
  verify: `cargo test --locked --offline --bins` and `--lib`

## Scope

- write: `.moai/specs/SPEC-20261122-lander-push/**`, `src/card/remote_copy.rs`, `tests/card_remote_copy_cli.rs`
- integration lines this slice starts beside (added when those files are free, by pair approval or after the holder's release, and the card's write scope is widened first): `src/card/mod.rs` (one module line), `src/ops/card.rs` (the copy tail of the land action and the clap lines of `card remote-copy`), `docs/design-review/appendix/operations/card.md` (the lines of the new action)
- read-only: `src/card/landed.rs` (the integration-branch, tip and containment readers the copy calls), `src/card/land.rs`, `.moai/specs/SPEC-20261122-remote-copy-slim/**`
- forbidden: any push to the real remote integration branch (tests use one temporary local bare repository); a force flag anywhere in the refspec or the command; a resident retry loop; a repair card from this slice; workflow file edits

## Plan

1. This SPEC, committed on the branch, with the probe of git's documented push output named in Intent.
2. Tests first (RED): `tests/card_remote_copy_cli.rs` on a temporary home and a temporary repository whose remote is one local bare repository — AC-1 to AC-6. Red is watched before the module and the integration lines exist.
3. `src/card/remote_copy.rs`: the copy operation — the branch and tip readers it calls from `landed`, git's own push-remote resolution for the branch, the `ls-remote` classification (equal, behind, ahead, absent, no shared ancestry), the porcelain push reader, the one record in the common git folder, the entry both callers call. Every failure folds into the answer, so no caller can fail on it.
4. The integration lines of Scope step two when those files are free, then AC-1 to AC-6 green.
5. Light checks (fmt, clippy, the bins and lib tests), `spec check` of this SPEC with a binary built on this branch, one READY note with head and base.

## Risks

- 복사는 착지 명령 안에서 돌아가므로 네트워크가 끊긴 채 push 가 매달리면 착지 명령도 매달린다: 이 조각은 자체 시간 제한을 두지 않는다(git 전송의 시간 제한이 적용된다). 시간 제한은 이후 조각이 측정 뒤 정한다.
- 뒤처진 실제 원격 분기의 첫 복사는 사용자 결정이고 이 조각이 아니라 리더가 실행한다: 코드는 fast-forward 만 하고, 갈라진 실제 원격은 강제가 아니라 보류로 기록된다.
- 기록은 공용 git 폴더에 있어 모든 작업트리가 공유한다: 두 착지가 동시에 쓰면 임시 파일과 이름 바꾸기로 마지막 쓴 것이 이긴다. 상태는 매 실행마다 다시 유도되므로 잃은 기록은 다음 복사가 치유한다.
- 통합 브랜치의 원격이 하나도 없는 저장소에서는 매 착지가 복사를 pending 으로 보고한다(기존 착지 시험들의 저장소가 그렇다): 답이 늘어도 착지는 그대로 성공이다.
- 참조 가드가 아직 착지 전이므로 이 조각은 그것 없이 검사된다: 가드가 착지되면 push 는 로컬 `refs/heads` 를 움직이지 않아 가드와 여전히 만나지 않는다.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | 복사는 명령층의 연산이지 착지 트랜잭션의 일부가 아니다: land 행동이 `land()` 의 답 뒤에 복사를 돌리고 그 상태를 답에 합친다. 그래서 이동과 종결의 트랜잭션은 손대지 않고 복사 실패가 착지를 거부할 수 없다. | 부모 REQ-1 의 failure 절, `src/card/land.rs` 모듈 주석의 순서 | 카드 열림 |
| D-2 | 복사 상태는 저장이 아니라 유도다: 매 실행이 원격 참조를 로컬 팁과 먼저 비교(같음·뒤짐·앞섬·없음·조상 아님)하고 뒤었거나 없을 때만 민다. | 동료 제안 (a)·(b), 리더 기록 2026-11-22 | 카드 열림 |
| D-3 | 거절과 원격 실패의 구분은 git 이 문서화한 porcelain 상태 줄로 한다(`! … [rejected] (non-fast-forward)` 인지 `remote failure` 인지; 도달 불가는 상태 줄이 없다). 측정: git 2.54.0, 2026-11-22. | git-push(1) OUTPUT 절과 이 개발기의 검사 | 카드 열림 |
| D-4 | 재시도 상태는 공용 git 폴더의 기록 하나 `moai/remote-copy`(branch·remote·tip·state·why, 임시 파일과 이름 바꾸기)이고, 재시도를 시작하는 것은 다음 착지 명령과 `card remote-copy` 뿐이다. | 동료 제안의 "SPEC names where the retry state is kept"; 참조 가드의 공용 폴더 선례(저장소당 기록 하나) | 카드 열림 |
| D-5 | 필요 시 명령은 `card remote-copy` 다(새 최상위 낱말 없음). 원장 행은 쓰지 않고 복사 상태를 답하며, 보류는 기록을 남긴 답이지 오류 종료나 카드가 아니다. | flow-skeleton D-1, 동료 제안 (c)·"no new card" | 카드 열림 |
| D-6 | 원격은 git 자신의 push 원격 규칙을 따른다(`branch.<name>.pushRemote`, `remote.pushDefault`, `branch.<name>.remote`, `origin` 순서)이고 refspec 은 언제나 `refs/heads/<branch>:refs/heads/<branch>` 에 강제 플래그가 없다. | git-push(1), fast-forward 전용 규칙 | 카드 열림 |

## Result

(status draft — written when the slice lands)

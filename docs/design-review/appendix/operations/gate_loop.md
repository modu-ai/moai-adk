# `gate-loop`

operation `gate.loop` · 코드 `src/ops/gate_loop.rs`, `src/gate_loop.rs` · Issue #2674 S4, #2860

`gate-loop --once [--dry-run] [--fetch] [--permission-reference REF] [--repo OWNER/REPO] [--base v2026.11.0] [--dir .] [--tip …] [--review-confirmed …]` 는 리더 병합 루프 한 바퀴다.

동작은 이래요.

- 통합 가지로 가는 열린 PR 을 **번호가 낮은 것부터**(#2860 규칙 5, 쉘판 `sort_by(.number)`) 하나씩 `gate-merge` 에 돌린다. 게이트를 통과한 것만 병합한다. GitHub 목록은 최신부터 오지만 여기서 다시 정렬한다.
- **초안도 판정한다**(#2860 규칙 4, 리더 결정 6870150891). 게이트를 통과한 초안은 `gate-merge` 가 status 를 올린 뒤, 병합 바로 전에 GraphQL `markPullRequestReadyForReview` 로 ready 로 바꾼다(REST 에는 이 동작이 없다). ready 가 실패하면 `rule: "READY"` 로 멈추고 병합하지 않는다.
- 멈춘 이유에 따라 head 마다 한 번 행동한다.
  - G6 "코드 파일 겹침": update-branch(`expected_head_sha` 고정)와 감사자 안내를 남긴다.
  - G2(판정 쓴 레인이 허용된 예약 보유자가 아님): 감사자 안내를 남긴다.
- "head 마다 한 번"은 PR 댓글 안의 숨은 표지 `<!-- moai:gate-loop <종류> head=<sha> -->` 로 판단하고, 로컬 상태 파일은 쓰지 않는다.
- head 가 로컬에 없는 PR 은 건너뛴다(`skipped`). 반복은 호출하는 쪽(리더, 도구의 스케줄러)이 한다.
- `--fetch`(#2906 S3, 기본은 꺼짐)는 열린 PR 목록을 읽은 뒤 `git fetch origin` 을 한 번 돌린다. refspec 은 `+refs/heads/<base>:refs/remotes/origin/<base>` 와 열린 PR 마다 `+refs/pull/<n>/head:refs/remotes/origin/pull/<n>` 이다.
  - 이름 붙은 원격 ref 로만 받고 `FETCH_HEAD` 는 쓰지 않는다(`--no-write-fetch-head`).
  - 운영자의 git 설정(인증, URL 바꿔 쓰기)은 쉘판처럼 그대로 쓴다.
  - fetch 가 실패하면 출력의 `fetch: {"ok": false, "error": …}` 로 알리고, 바퀴는 로컬에 있는 것으로 계속한다. 성공하면 `fetch: {"ok": true, "refspecs": […]}` 이다.
  - 쓰기는 `--dir` 저장소의 원격 ref 뿐이다. `--dry-run` 과 함께 써도 이 ref 는 바뀐다.
- tip 이 빨강이면(G8, #2777 — `gate-check` 와 같은 판정) 수리 PR 이 아닌 PR 은 `stopped`·`rule: "G8"` 로 적고, 병합·update-branch·안내 댓글을 하나도 쓰지 않는다.
- 위 C2 기계적 delta(#2833)는 루프 안에서도 쓰인다. PASS 받은 head 에 tip 만 합친 PR 은 새 판정 없이 `AUTO-DELTA` 기록 뒤 병합된다.

`--dry-run`(#2860 D2)은 읽기만 하고 아무것도 쓰지 않으며, 권한 참조가 필요 없다. PR 마다 `outcome` 이 하나 나온다.

- `would_merge`, `would_merge_auto_delta`
- `would_update_branch` 또는 `already_updated`
- `would_ask_auditor` 또는 `already_asked`
- `stopped`(그 밖의 게이트 중단, `rule`·`reason` 포함)
- `skipped`(읽기 실패)

실제 바퀴가 첫 쓰기 직전에 하는 원격 base·head 재확인은 하지 않는다. 그래서 `would_merge` 도 실제 바퀴에서는 그 확인을 한 번 더 거친다. 리더가 쉘 루프에서 이 명령으로 바꾸기 전에 같은 PR 목록에서 결정을 견주는 데 쓴다.

병합 결과를 다시 읽지 못한 PR(`merged: "uncertain"`)은 `outcome: "uncertain"` 으로만 적고 다른 쓰기를 하지 않는다. status 만 감사 App 토큰으로 올리고, 읽기·병합·ready·update-branch·안내 댓글은 운영자(gh 자체 인증)로 한다(#2735 감사 F3).

게이트 규칙 쪽(#2860, `gate-check`·`gate-merge` 공통)도 쉘판에 맞췄다.

- G2: 댓글 **첫 줄**의 `moai-claim-handoff: comment=<id> to=<lane> by=leader` 가 허용된 예약을 그 레인으로 옮긴다. 허용된 예약을 가리키고 그 뒤에 쓴 것만 인정하며, 마지막 인계가 이긴다.
- G1: 판정 base 가 7–39자 짧은 SHA 여도 collect 가 `rev-parse --verify` 로 커밋 하나로 풀면(`verdict_base_full`) 인정한다. head 는 40자 그대로다.
- G6: `src/lib.rs` 겹침은 PR 쪽 변경이 `mod x;`/`pub mod x;` 추가 줄뿐이고 병합 tree 에 모두 남으면(`lib_mod_only_kept`) 허용한다.
- G8(tip 빨강 멈춤)은 #2554 D8(a) 로 판정한다(#2860 범위 추가, 근거 #2856 감사 N10; 리더 도구 `green_tip.py` 와 같은 규칙).
  - 모음 job `ci` 는 보지 않고, 남은 job 을 `gate_collect::d8a_judgment` 로 판정한다. 이 함수는 #2856 `lane-next` 와 같은 함수다.
  - 진짜 job 실패는 빨강이다.
  - 다음 병합이 취소한 macOS job 은, tip 쪽 더 새 커밋에서 같은 job 이 성공했을 때만 통과로 본다. 그렇지 않거나 macOS 가 아닌 job 이 취소됐으면 그 커밋은 아직 판정하지 않고 더 오래된 커밋을 읽는다.
  - job 이 모음 `ci` 하나뿐인 저장소(사용자 프로젝트)는 예전처럼 그 check 로 판정한다.
- 쉘판의 마감 인계 자동 게시(`auto_handoff.py`)는 옮기지 않았다. 제품 설계가 "자동 인계 없음"(`src/claim.rs`, `src/handoff.rs`)이라서다. counts 갱신은 상태줄 렌더가 TTL 로 한다.

## 쉘 루프 한 바퀴와 제품의 대응 (#2906 S3)

리더의 `auto_loop.sh` 가 한 바퀴에 하는 일과 제품에서 그 일을 맡는 곳이다.

| 쉘 루프 단계 | 제품 | 비고 |
| --- | --- | --- |
| `auto_handoff.py --post`(마감 인계 자동 게시) | 옮기지 않음 | 제품 설계가 "자동 인계 없음"이다(`src/claim.rs`, `src/handoff.rs`). 인계는 사람이나 리더가 쓴다. |
| `refresh_counts.sh`(상태줄 counts 갱신) | 옮기지 않음 | 상태줄 렌더가 TTL 로 스스로 갱신하므로 루프가 할 일이 없다. |
| `gh pr list … --jq`(CI SUCCESS 이고 지금 head 에 PASS 가 있는 PR 만 고름) | `gate-loop` 이 열린 PR 을 모두 게이트에 돌림 | 미리 거르지 않는다. 지금 head 에 PASS 가 없으면 게이트가 멈춰 `stopped` 가 된다. #2906 S5b 부터 CI 는 막지 않는다: CI 가 SUCCESS 가 아니어도 지금 head 의 `local-check:` pass 줄이 있으면 병합 대상이고, 그 줄이 없을 때만 G5 에서 `stopped` 가 된다(`tests/gate_loop.rs` `shell_prefilter_cases_stop_in_a_dry_run`). |
| `git fetch -q origin <base> refs/pull/$N/head`(PR 마다) | `gate-loop --fetch` | 한 번의 fetch 에 명시 refspec 을 모두 담는다. `FETCH_HEAD` 에 기대지 않는다(`fetch_brings_open_heads_before_the_pass`). |
| `gate_merge.sh $n` | `gate-merge`(루프 안에서 PR 마다) | 같은 게이트 규칙 G1–G9. |
| "STOP overlap in non-doc file" → update-branch + 안내 | G6 → update-branch(`expected_head_sha`) + 안내, head 마다 한 번 | 쉘의 `auto-updated.txt` 대신 PR 댓글 속 숨은 표지. |
| "STOP the last verdict … audit claim" → 감사자 안내 | G2 → 감사자 안내, head 마다 한 번 | 쉘의 `auto-owner-asked.txt` 대신 숨은 표지. |
| `sleep 90` | 스케줄러 간격(#2906 S2) | launchd `StartInterval 90`(90초), cron `*/2`(2분, cron 은 분 단위가 가장 작다). `gate-loop` 은 한 바퀴만 돈다. |

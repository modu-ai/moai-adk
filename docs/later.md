# Later list

One line per item, newest last: `- <date UTC> | <where it was seen> | <what> | <the change at the source or the follow-up>`.

The lines after the dated ones came from the leader's carry-over of closed cards and keep their original wording. A non-blocking finding of an audit, a failed remote run and the second sight of one problem class each add one line (`docs/leader-rules.md`, `docs/lane-rules.md`). The session that sees it writes the line; the leader turns lines into cards while the work cap allows and deletes a line that is done or does not reproduce. A line holds no secret, adds no criterion to a card and opens no card by itself.

- 2026-11-25 | the completion flow rules | the `spec/` branch hook still refuses code writes on a branch with no SPEC file, while the rules ask for a SPEC at tier 2 only | relax the hook for branches of changes below tier 2 (code change)
- 2026-11-25 | the completion flow rules | `moai card land` does not yet move the integration branch and the primary checkout's work tree together when that checkout is clean (user decision of 2026-11-24T21:52:39Z, item 3) | the landing code change
- 2026-11-24 | audit of a spec-digest card | the SPEC `content_digest` hashes neither Supersedes nor the Decisions table, so a table changed after approval needs no new approval | hash the table or warn when it changes (SPEC tool)
- 2026-11-25 | bundle reads REPAIRED notes | `moai card inbox` does not count a `REPAIRED` note as `ready_for_integration` (`inbox_family_follows_marker_of_line` pins it) | decide whether the inbox counts `REPAIRED`
- 2026-11-25 | bundle reads REPAIRED notes | `tests/skill_references_test.py` (Python) already fails since the MANIFESTS table left | rewrite it as a Rust test or delete it
- 2026-11-25 | workflow audit report, section 4 | the installed `work.md` in the home skill folder lacks a line the source `work.md` has, and nothing reports whether the source hash, the home install and a new session's load agree | a read-only `moai` check that prints the three facts apart (source hash, home install hash, load by a new session), normal rendering differences allowed
- C-469 #4: 후속(감사 N1): 라운드의 lease_sweep 답이 via 와 silent_secs 를 그대로 보여 주고, 문서와 도움말이 24시간 grace 와 2시간 하트비트 두 길을 말해요
- C-160 #4: 감사 후속: src/update/switch.rs:224 의 go/bin 안내 문구를 '링크' 로 고치고, 남은 임시 링크(.moai.switch-*)와 앞선 보관 사본을 다음 apply 나 revert 에서 정리해요 — update_switch_layer 테스트
- C-255 #2: Follow-ups from the C-159 audit: the no-pick test in tests/project_harness_pick.rs asserts the three conflicts AC-4 names (CLAUDE.md, REVIEW.md, .gitignore) with a fixture that makes all three; the src/generation.rs:1264-1268 comment says what a never-generated project records under the harness-only pick; research.md:28 gives the source URL; tests/generation_request_limits.rs:39 drops the retired codex host
- C-255 #3: 감사 후속: src/generation.rs record_generation 주석(REQ-9, AC-7)을 쓰고 project.md:69 의 'the generation apply' 를 표시 이름으로 바꿔요 — project_harness_label 테스트
- C-461 #3: 감사 후속: 실행 중인 폴더의 다른 이름 영수증 id 도 자기 영수증 후보로 읽기(REQ-6, N1), 다른 이름 id 와 journal 로 옮긴 폴더를 함께 다루는 테스트로 뮤턴트 &own[..1] 잡기(N2)
- C-464 #4: 감사 후속: REQ-12 를 실제 동작(거절된 take 도 sweep 은 남음)대로 고치고, grace_secs <= 0 경계 양쪽 테스트(N-1), dry-run 읽기 전용 열기 또는 문구(N-2), 쓰이지 않는 grace_secs/grace_hours 정리(N-3), CODEX_THREAD_ID 제거(N-6)
- C-466 #5: 후속(감사 N3·N5): every_stage_draws_its_item_of_the_table 가 ⏸️ 와 🙋 에 서로 다른 수를 써요. SPEC-20261124-card-flow-icons 범위 78줄 문구가 범위 하나로 읽혀요
- C-477 #4: 후속: card inbox 가 첫 줄을 200자에서 잘라도 kind=decision 을 막힘으로 잘못 읽지 않아요 — 테스트
- C-477 #5: 후속(감사 minor 1): 답이 4096바이트 한도 몇 바이트 안까지 차는 테스트로 경계를 고정해요(ENVELOPE_BYTES 뮤턴트가 죽어야 해요)
- 상태줄: 브랜치가 원격 통합 브랜치보다 몇 커밋 앞서고 뒤처졌는지 다시 표시(사용자 2026-11-24, upstream 이 없으면 릴리스 기준 브랜치와 비교)
- 클라우드 세션 지원: 저장소 .claude/ 에 스킬·훅 커밋 + 시작 스크립트로 moai 설치(공개 릴리스 뒤에 판단)
- 카드 scope 명령이 공유 파일과 겹치는 글롭을 거절할 때 겹친 파일만 빼고 넣는 선택지
- 승인 기록을 쓰는 일이 손작업이라 `spec check --all` 을 빠뜨리면 넓은 검사가 실패해요(2026-11-25, interview_missing): 승인 블록을 쓰고 digest·Interview·전체 검사를 한 번에 하는 `moai spec approve-record` 명령
- check affected 의 TMPDIR 이름이 길거나 세션 id·`agent-` 글자를 담으면 테스트가 깨져요(2026-11-25 세 번, sccache SUN_LEN): TMPDIR 을 받을 때 길이와 글자를 검사해 거절하고, `nextest_with_log` 가 RUSTC_WRAPPER·SCCACHE_* 를 지워요
- 착지 머지에서 tools/repo_scans.json 의 긴 `--test` 한 줄과 source-manifest.json 해시가 카드마다 충돌해요(2026-11-25 네 번): 테스트 이름을 한 줄에 하나씩 쓰는 배열로 바꾸고 해시는 `moai` 명령이 다시 계산해요
- agents/audit.md 와 workflows/run.md 에 옛 docs-only 줄("the lander reads the change and lands it") 사본이 남았어요(C-522 감사): 세 곳 복사를 한 곳 가리키기로 줄여요
- 카드 쓰기 범위나 SPEC 의 `## Scope` 와 바뀐 경로를 착지·감사 때 견주는 코드가 없고, `card check` 는 레인만 부르며, `card scope` 는 호출자 확인 없이 예약을 넓혀요(2026-11-25, SPEC approval-digest review of the write scope; `src/card/work.rs` check, `src/ops/card.rs` scope, `src/card/submit.rs` 의 outside-scope 줄): 착지 때 헤드의 변경 경로를 예약 범위와 견주거나 `card scope` 에 호출자 확인을 두어요(카드 도구 코드 변경)

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
- `agents-local-merge` 가 사용자가 거절한 옛 백업 `AGENTS.local.md` 를 update 때마다 다시 제안하고, 같은 때 `agents-picks` 는 원본이 없다며 skip 이라 둘이 어긋나요(2026-11-25 `/moai update`): 거절을 기억하는 C-518 이 받아요
- 카드 쓰기 범위나 SPEC 의 `## Scope` 와 바뀐 경로를 착지·감사 때 견주는 코드가 없고, `card check` 는 레인만 부르며, `card scope` 는 호출자 확인 없이 예약을 넓혀요(2026-11-25, SPEC approval-digest review of the write scope; `src/card/work.rs` check, `src/ops/card.rs` scope, `src/card/submit.rs` 의 outside-scope 줄): 착지 때 헤드의 변경 경로를 예약 범위와 견주거나 `card scope` 에 호출자 확인을 두어요(카드 도구 코드 변경)
- 실행·동기화 글을 줄이면서 Python 시험 여섯 개가 원격 CI 에서 깨져요(2026-11-25, 로컬 `check affected` 는 Python 시험을 고르지 않아요): `check_pr_self_audit_test.py`(run.md 점검 규칙 표·CI 줄), `self_check_test.py`(self-check.md 의 스크립트·`--full-python` 줄), `smoke_harness_test.py`, `skill_web_pages_test.py`(run.md 웹 페이지 절), `sync_drift_fixture_test.py`(sync.md 3단계의 "일곱 갈래"·부록 경로), `skill_references_test.py`(옮겨진 절 이름): 시험은 C-317 이 Rust 로 옮기고 지워요
- 작업 수 상한이 세션 등록부의 레인만 세는데 `workflows/work.md` 에 등록 단계가 없어 레인 1개로 세었어요(2026-11-25, C-524 take 거절, 사용자 결정 "레인마다 등록"): 레인 글의 첫 단계에 `moai session-register <프로젝트 경로> --name <이름> --mode agent` 와 갱신·해제를 넣고, 등록 항목의 열쇠를 이름이 아니라 세션 id 로 해서 이름이 같은 다른 세션을 덮어쓰지 않게 해요(mo.ai.kr 덧붙임 7 J)
- `card land-batch` 가 SPEC 없는 카드를 `no_spec` 으로 떨궈요(src/card/land/batch.rs, C-524 승인 D-2): SPEC 없는 tier 0·1 카드도 묶음 착지가 받게 해요
- `card submit` 이 예약에 적힌 브랜치 이름과 실제 브랜치가 다르면 "HEAD 가 예약 브랜치의 tip 이 아니다"로 거절해요(2026-11-25 C-520): 예약 브랜치를 실제 브랜치로 맞추는 명령이나 take 때 이름 검사
- `moai audit-delta` 가 부모 순서를 [pass_head, tip] 로 요구해서 tip 이 첫째 부모인 착지 머지를 `parent_mismatch` 로 거절해요(2026-11-25 C-508·C-520 묶음 판정): 두 순서를 모두 받아요
- `card submit` 이 예약에 상대 경로로 적힌 워크트리(`--worktree .claude/worktrees/<이름>`)를 풀지 못해 거절해요(2026-11-25 C-509·C-518·C-524 세 레인, 워크트리 안에서는 "not a folder", 주 체크아웃에서는 trailers 관문의 `git -C` 이중 해석): `card take` 가 `--worktree` 를 프로젝트 기준 절대 경로로 저장하거나 submit 이 예약 경로를 프로젝트 기준으로 풀어요. 그때까지 레인은 READY 노트 → `card review` 손 순서를 써요
- SPEC 브랜치 쓰기 훅이 SPEC id 와 브랜치 이름이 다른 워크트리(예 spec/home-guard-flake-tmpdir-20261124 와 SPEC-20261124-home-guard-flake)에서 `target/<이름>/` 의 감사 노트 초안까지 막아요(2026-11-25 C-516 감사, agent-7): 훅이 `target/` 아래 스크래치는 SPEC 연결과 상관없이 허락해요
- 감사·읽기 명령의 `cd` 가 호스트 세션의 작업 폴더를 다른 레인 워크트리 안으로 옮겨요(2026-11-26 C-538 감사 nora 두 번, 앞서 ivan 한 번): 다른 워크트리로 가는 `cd` 를 거절하거나 경고하는 가드
- src 안의 테스트가 옛 상태 폴더 철자(`.moai-cli`)를 직접 써서 `project_state_dir_is_spelled_once` 스캔에 걸려요(2026-11-26 C-175 조각 1b 첫 영수증, 앞 조각에도 한 번): 테스트 도우미가 상태 폴더 이름을 주고 테스트 작성 지침에 한 줄
- 부하에 따라 시간이 흔들리는 테스트가 착지 검사를 막아요(2026-11-25 lane_screen C-538, 2026-11-26 outbox ac5 C-539): 고정 임대·고정 대기 대신 장부 시각을 옮기는 도우미를 쓰고, 새 테스트의 고정 대기를 찾는 스캔
- `SPEC-20261125-check-affected-rerun` 이 착지·닫힘(2026-11-25)까지 Result 절에 AC-1..AC-8 줄이 없어 `moai bundle check` 가 `result_missing` 을 냈는데 `card land` 는 막지 않았어요. 원인을 고칠 곳: `card submit` 이나 `card land` 가 승인 SPEC 의 Result AC 줄을 bundle check 와 같은 규칙으로 읽어 빠지면 거절(그 SPEC 의 Result 는 다음 SYNC 카드가 채워요).

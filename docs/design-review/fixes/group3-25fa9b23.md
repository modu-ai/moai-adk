# group3-25fa9b23 수정안 — MoAI 릴리스와 Desktop 상세 설계

- 대상: `references/chatgpt-space/page_25fa9b23c690819180a1351d6da22e23.md` (읽기 전용 입력)
- 검토 근거: `review-group3-install-host-plan-run-release.md` §3·§8(X4·X9·X11·X12)·§9.2
- 고정 근거 tip: **86762b8e1** (처음 작성 기준은 604792666). E3 교체문은 통합 base **8c5732615**, 04-D 상태는 2026-11-10 독립감사 `6759943258`·리더 회신으로 보강했다. 기존 고정 인용과 현재 보강 근거는 구분한다.
- 우선순위: S2(§5·§10 결정됨/구현 상태)

## 수정 항목

| # | 위치 | 현재 | 교체 | 근거 (tip 86762b8e1 재확인) |
| --- | --- | --- | --- | --- |
| 1 (S2, X4) | §5(tag·version 절 — "v2026.11.0-rc.1 tag는 후보이며 …")과 §10(packaging 절 — "cargo-dist는 packaging 후보지만 …") | 결정 이전 상태("후보")로만 서술 — 결정됨·구현 상태의 구분이 없다 | §5·§10 각 항목을 "결정 / 구현 상태" 두 열로 다시 쓴다: "tag `moai-cli-v2026.11.0-rc.1` — **결정됨**(L9) / 미태그, tag ruleset `moai-cli-v*`는 사용자 승인 1회 카드(L107, #2197 진행 중). packaging — **결정됨**: cargo-dist(L7) / **도입됨**: `dist-workspace.toml`(cargo-dist 0.34.0, `dist build`로 archive별 `.sha256`·통합 `sha256.sum` 산출 실측, E1 #2193). 다만 게시 workflow는 없고(`ci = []`) target은 `aarch64-apple-darwin` 하나다. provenance — **결정됨**: attestation + sha256만, notarization 없음(L8·L107) / **미구현**: 운영자 승인 항목만 문서화(`docs/release-attestation.md`, E4 #2196). self-replace — **결정됨**(L12·L107) / **구현됨**: `self-replace = "=1.5.1"` 의존성과 `src/self_update.rs`가 있으며 E3 #2195 / PR #2320은 base에 병합됐다. Windows `windows-build`는 시험 target까지 컴파일하지만 시험 실행·실행 중 교체·host 로딩은 **not_run**이다. OS 범위 — **결정됨**(L107): macOS arm64·x86_64 + Linux x86_64·arm64, Windows는 빌드+self-replace 구현만이다. rc-ci는 이 4 target에서 check set을 돈다(E2 #2194)" | 고정 결정 근거: `86762b8e1:docs/decisions-2026-11-08.md` L7·L8·L9·L12·L107, `86762b8e1:docs/release-checklist.md:30`, `dist-workspace.toml`(`ci = []`, 단일 target 유지). 현재 E3 근거: `8c5732615:Cargo.toml:19`, `8c5732615:src/self_update.rs`, `8c5732615:docs/support-matrix.md:17·47`; 독립감사 `6759943258`의 #2195 완료·PR #2320 병합 확인. `.github/workflows/rc-ci.yml`(4 target matrix), 검토 §8 X4·§9.2 |
| 2 (X9) | §격리 시험 명령 예시(L114) — "moai-cli setup --dry-run --json, moai-cli project init --apply-plan <id> --json, moai-cli maintenance update --dry-run --json, … doctor --json, capabilities --json" | 존재하지 않는 내부 subcommand 체계 | 실제 subcommand로 교체: "`moai-cli project-plan <project> --json`(stdin profile) → `moai-cli project-apply <project> --plan-digest <d> --permission-reference <r>`, `moai-cli update-plan <home> <stage>` → `moai-cli update-apply <home> <stage> --plan-digest <d> --permission-reference <r>`, `moai-cli claude-doctor`·`moai-cli codex-doctor`, `moai-cli install-inspect --json`. **`capabilities`는 RC1에 없다.** 이후 검증된 명령 전환 단계에서만 실행 이름을 moai로 바꾼다는 문장은 유지(X12와 정렬)" | `docs/design-review/appendix/interfaces.md` § "RC1 OPERATIONS 전수"(70개 — project-plan/apply·update-plan/apply·claude-doctor·codex-doctor·install-inspect 존재, capabilities 부재; `src/main.rs:1866` `operations_doc_matches_code`가 코드와 대조), 검토 §8 X9 |
| 3 | §3 crate 구조 절 | 다중 crate 구조로 읽히는 서술 | "RC1은 **단일 crate**(`moai-cli` 1개 — Cargo.toml에 [package] 1건)이다. host adapter·forge·claims 등 모듈 분리는 분리 후보이며 RC1에서 수행하지 않는다" | `Cargo.toml`(tip 관측 — [package] 1건, `dist-workspace.toml`의 workspace 멤버도 `cargo:.` 하나), 검토 §3.6.3 |
| 4 | §"Host 전환" 단락 — "P17에 따라 성공 host는 유지하고 실패 host의 안전한 이번 변경만 복구한다." | "P17"이 정의 없이 사용됨 | P17 정의를 넣는다: "P17(설계 01-page D72/P17): UPDATE 실패 시 **성공 host는 유지**하고 실패 host의 **이번 변경만** 복구한다(사용자 편집·이전 버전으로의 전면 롤백 아님). RC1 실물: 실패 host는 `pending` receipt + known-good 사본을 남기고(04-B #2091), `recovery-plan`/`recovery-apply`가 되돌린다(04-C #2000). 모든 중단 지점의 복구는 rc.2 E6 시험(`tests/release_suite.rs`)이 고정한다" | `docs/rust-product-path.md` UPDATE·Recovery 절(tip — "hosts that already committed are never rolled back", pending+known-good), `src/update.rs`(failed host pending 유지), `tests/release_suite.rs`(E6 #2198 중단 지점 k=0~3), 검토 §3.6.4 |
| 5 (X12) | §9 전환 절 머리(L735 전후 "설치 탐지: …" 단락들) | 전환 단계가 RC1 내 계획처럼 읽힘 | 절 머리에 배너: "**RC1 비범위 — 이 절 전체(설치 탐지·분리 시험·Host 전환·성공 gate·명령 전환)는 04-D의 계획이다. #2016은 2026-11-11에 재설계 범위(합성 HOME 한정 백업·선택 이전·복구와 CLI 연결, PR #2449)로 닫혔고, 실제 HOME 이전과 Go `moai update` 브리지는 #2459(HOLD, 사용자 결정 대기)로 분리됐다. 어느 쪽도 기존 Go 설치에 대한 자동 쓰기·이동·삭제·전환 승인이 아니다. RC1에서 실기·설치·삭제·전환은 수행하지 않았다(본문 말미 명시 유지).**" | 과거 범위 근거: `86762b8e1:docs/decisions-2026-11-08.md` L94 ②·L90. 현재 Issue 상태(2026-11-12 재확인): #2016 CLOSED 2026-11-11T09:58Z(리더 종료 댓글 6792799989, 대조표 6792572670, 병합 `313c7f186f`), #2459 OPEN·`needs:decision`. #2016의 실제 작업 범위·owner·검증은 해당 Issue와 최신 사용자 결정으로 별도 확인, 검토 §3.6.5·§8 X12 |
| 6 | §말미 시험 ID(L751) — "필수 시험 ID는 RL-GENERATION-ISOLATION, RL-HOST-EXPLICIT-SWITCH, RL-COMMAND-CUTOVER, RL-CUTOVER-RESTORE다. 실기·설치·삭제·전환은 아직 수행하지 않았다." | RC1 대응 시험·not_run 여부 미기재 | 각 ID에 RC1 상태 추가: "RL-GENERATION-ISOLATION — **not_run**(RC1 receipt 없음; 부분 관측만 존재: T1 fixture가 가짜 Go 상태 digest 불변을 단언 — AC-08C-1 receipt 내부, AC-08C-7 sandbox가 실제 HOME 쓰기 차단). RL-HOST-EXPLICIT-SWITCH·RL-COMMAND-CUTOVER·RL-CUTOVER-RESTORE — **not_run, RC1 범위 밖**(04-D #2016은 합성 HOME 범위로 닫혔고, 실제 HOME 이전·Go 브리지는 #2459 HOLD다. 합성 HOME 이전 시험 `tests/migrate_cli.rs`·`tests/migration_fixture.rs`도 receipt가 아니다). Go updater 격리는 단위 시험으로만 있다: `tests/install_cli.rs`(04-A 특성화)와 `tests/release_suite.rs`(E6 — 평문 `v*` tag가 Go RC 채널에 걸리는 위험, 정식판이 `/releases/latest`를 가져가는 위험 재현). 이는 receipt가 아니다" | `tests/install_cli.rs`·`tests/release_suite.rs`(tip 존재), `docs/e2e-receipts/receipts.jsonl`(RL-* receipt 없음), `docs/support-matrix.md` "RELEASE 시험군" 절; #2016 상태는 2026-11-12 재확인(CLOSED 2026-11-11, 리더 종료 6792799989, 분리 #2459 HOLD), 검토 §3.6.6 |
| 7 (X11) | L737 — "새 Rust RC의 실행 파일·업데이트 경로·상태는 기존 Go 설치와 분리한다. …" 및 L739 — "테스트 완료 후 기존 Go 설치를 제거하고 …" | [83f] L18·L2611과 같은 문단 | **한 교체문 · 세 위치 적용** — 홈·프로젝트 root 분리 문단(L737)은 group3-83f162ef.md X11 절의 공통 교체문을 동일 적용, Go 전환 문장(L739)은 같은 문서 표 5번의 X12 배너와 83f 표 8번 교체문을 동일 적용 | 검토 §8 X11·X12 |

## 시험 ID 행 RC1 상태 (AC-2)

| ID | RC1 상태 | 근거 |
| --- | --- | --- |
| RL-GENERATION-ISOLATION | not_run (부분 관측: AC-08C-1·7 receipt 내부, E6 단위 시험) | receipts.jsonl에 RL-* 없음 |
| RL-HOST-EXPLICIT-SWITCH · RL-COMMAND-CUTOVER · RL-CUTOVER-RESTORE | not_run, RC1 범위 밖 | decisions L94 ②·L90(04-D 후속) |
| DT-* (Desktop) | not_run | Desktop은 RC1 후속 |

## 비고

- 가설 단독 근거 0, 개인 절대 경로·비밀 0건(AC-3).
- 기록 정리 대상(사용자 선택 아님): decisions L107(01:1x)이 OS matrix를 정했는데 L132의 "P12에서 아직 HOLD인 것" 목록에는 아직 "OS/arch matrix"가 남아 있다. 이 수정안은 L107을 따른다.

## 갱신 기록 (#2207, tip 86762b8e1)

- 1번: "cargo-dist 미구현(release-checklist:29)"은 낡은 사실이었다. E1 #2193이 `dist-workspace.toml`과 `sha256.sum` 산출을 도입했다(게시 workflow는 아직 없음). "P12·release matrix HOLD"는 L107로 결정됐다. attestation(E4)·self-replace(E3)·tag ruleset(E5)의 현재 상태를 넣었다.
- 2번: 근거를 `src/main.rs:613-640`(옛 OPERATIONS 배열, 47개)에서 `docs/design-review/appendix/interfaces.md` 전수(70개)로 바꿨다.
- 4번·6번: rc.2 E6 시험(`tests/release_suite.rs`)을 단위 시험 근거로 더했다.
- 5번·시험 ID 표: 근거 "decisions:89-91"은 잘못된 줄이었다. L94 ②·L90으로 고쳤다.

## 독립감사 후 보강 (#2207, base 8c5732615)

- E3의 구현·병합과 Windows 실기 `not_run`을 분리하고, 04-D의 과거 종료 기록과 #2016의 OPEN 재개 상태를 구분했다. 기존 Go 자산 변경 승인이나 Space 적용은 수행하지 않았다.

## 현재 tip 재확인 (#2207, 2026-11-12, tip `325cc2bf03`)

- 인용 대조(`86762b8e1` → `325cc2bf03`, 같은 줄 범위의 글자 비교): 5건 중 그대로 1, 줄만 이동 1, 내용 바뀜 2, 모호 1, 고정 tip에 없던 파일 0(뒤 base `8c5732615` 근거). 인용은 각자 적힌 고정 tip 기준이라 그대로 두고, 교체문에 옮겨지는 사실만 다시 확인했다.
- 5번 교체문·근거: #2016을 "OPEN 재개"에서 "2026-11-11 재설계 범위로 닫힘(PR #2449), 실제 HOME 이전·Go 브리지는 #2459 HOLD"로 고쳤다.
- 6번 교체문·근거: 같은 #2016 상태와, 합성 HOME 이전 시험(`tests/migrate_cli.rs`·`tests/migration_fixture.rs`)도 receipt가 아니라는 점을 적었다.
- 교체문에 들어간 다음 사실은 tip에서 다시 확인했고 그대로다: claim lease 기본 4시간·heartbeat 없음 (`src/claim.rs` `DEFAULT_LEASE`), `Budget`은 `max_attempts` 하나·기본값 없음(`src/run.rs`), E2E receipt 9건(6 passed + 3 not_run, `docs/e2e-receipts/receipts.jsonl`), GitLab은 capability 선언만 (`src/forge/gitlab.rs`, 호출 경로 없음), hook 생성의 host 로딩 `not_run`(`docs/rust-product-path.md` `hooks.apply` 행), `capabilities` 명령 없음, Codex config CLI 없음(#2237 PR #2534 미병합), #2197·#2014 OPEN. OPERATIONS 전수는 지금 99개다(고정 근거 `86762b8e1`에서는 70개). 70을 적은 근거 열은 그 고정 tip 기준이며, 교체문에는 개수가 없다.

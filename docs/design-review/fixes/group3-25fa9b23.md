# group3-25fa9b23 수정안 — MoAI 릴리스와 Desktop 상세 설계

- 대상: `references/chatgpt-space/page_25fa9b23c690819180a1351d6da22e23.md` (읽기 전용 입력)
- 검토 근거: `review-group3-install-host-plan-run-release.md` §3·§8(X4·X9·X11·X12)·§9.2
- 적용 tip: **604792666** (2026-11-09 재확인)
- 우선순위: S2(§5·§10 결정됨/구현 상태)

## 수정 항목

| # | 위치 | 현재 | 교체 | 근거 (tip 604792666 재확인) |
| --- | --- | --- | --- | --- |
| 1 (S2, X4) | §5(tag·version 절 — "v2026.11.0-rc.1 tag는 후보이며 …")과 §10(packaging 절 — "cargo-dist는 packaging 후보지만 …") | 결정 이전 상태("후보")로만 서술 — 결정됨·미구현의 구분이 없다 | §5·§10 각 항목을 두 열로 다시 쓴다: "tag `moai-cli-v2026.11.0-rc.1` — **결정됨**(decisions:9)/미태그(P12·release matrix HOLD). packaging — **결정됨**: cargo-dist 결정 A(decisions:7)/**미구현**(`docs/release-checklist.md:29` "cargo-dist 미도입 — 수동 sha256.sum 절차가 이 항목을 선행"). provenance — **결정됨**: attestation 단독(decisions:8)/미구현. self-replace — **결정됨**(decisions:12)/**미구현**(Cargo.toml 의존성에 없음)" | `docs/decisions-2026-11-08.md:7-10·12`(현재 tip 재인용), `docs/release-checklist.md:29`(cargo-dist 미도입 명시), `Cargo.toml`(단일 crate — [[bin]]·[package] 1건, self-replace 의존성 부재), 검토 §8 X4·§9.2 |
| 2 (X9) | §격리 시험 명령 예시(L114) — "moai-cli setup --dry-run --json, moai-cli project init --apply-plan <id> --json, moai-cli maintenance update --dry-run --json, … doctor --json, capabilities --json" | 존재하지 않는 내부 subcommand 체계 | 실제 subcommand로 교체: "`moai-cli project-plan <project> --json`(stdin profile) → `moai-cli project-apply <project> --plan-digest <d> --permission-reference <r>`, `moai-cli update-plan <home> <stage>` → `moai-cli update-apply <home> <stage> --plan-digest <d> --permission-reference <r>`, `moai-cli claude-doctor`·`moai-cli codex-doctor`, `moai-cli install-inspect --json`. **`capabilities`는 RC1에 없다.** 이후 검증된 명령 전환 단계에서만 실행 이름을 moai로 바꾼다는 문장은 유지(X12와 정렬)" | `src/main.rs:613-640`(OPERATIONS 실측 목록 — project-plan/apply·update-plan/apply·claude-doctor·codex-doctor·install-inspect 존재, capabilities 부재), 검토 §8 X9 |
| 3 | §3 crate 구조 절 | 다중 crate 구조로 읽히는 서술 | "RC1은 **단일 crate**(`moai-cli` 1개 — Cargo.toml에 [[bin]]·[package] 각 1건)이다. host adapter·forge·claims 등 모듈 분리는 분리 후보이며 RC1에서 수행하지 않는다" | `Cargo.toml`(tip 관측 — [package] 1건), 검토 §3.6.3 |
| 4 | §"Host 전환" 단락 — "P17에 따라 성공 host는 유지하고 실패 host의 안전한 이번 변경만 복구한다." | "P17"이 정의 없이 사용됨 | P17 정의를 넣는다: "P17(설계 01-page D72/P17): UPDATE 실패 시 **성공 host는 유지**하고 실패 host의 **이번 변경만** 복구한다(사용자 편집·이전 버전으로의 전면 롤백 아님). RC1 실물: 실패 host는 `pending` receipt + known-good 사본을 남기고(04-B #2091), `recovery-plan`/`recovery-apply`가 되돌린다(04-C #2000)" | `docs/rust-product-path.md` UPDATE·Recovery 절(tip — "hosts that already committed are never rolled back", pending+known-good), `src/update.rs`(failed host pending 유지), 검토 §3.6.4 |
| 5 (X12) | §9 전환 절 머리(L735 전후 "설치 탐지: …" 단락들) | 전환 단계가 RC1 내 계획처럼 읽힘 | 절 머리에 배너: "**RC1 비범위 — 이 절 전체(설치 탐지·분리 시험·Host 전환·성공 gate·명령 전환)는 04-D 후속의 계획이다. RC1에서 실기·설치·삭제·전환은 수행하지 않았다(본문 말미 명시 유지).**" | 검토 §3.6.5·§8 X12, `docs/decisions-2026-11-08.md:89-91` |
| 6 | §말미 시험 ID(L751) — "필수 시험 ID는 RL-GENERATION-ISOLATION, RL-HOST-EXPLICIT-SWITCH, RL-COMMAND-CUTOVER, RL-CUTOVER-RESTORE다. 실기·설치·삭제·전환은 아직 수행하지 않았다." | RC1 대응 시험·not_run 여부 미기재 | 각 ID에 RC1 상태 추가: "RL-GENERATION-ISOLATION — **not_run**(RC1 receipt 없음; 부분 관측만 존재: T1 fixture가 가짜 Go 상태 digest 불변을 단언 — AC-08C-1 receipt 내부, AC-08C-7 sandbox가 실제 HOME 쓰기 차단). RL-HOST-EXPLICIT-SWITCH·RL-COMMAND-CUTOVER·RL-CUTOVER-RESTORE — **not_run, RC1 범위 밖**(04-D 후속). Go updater characterization의 실물 시험은 `tests/install_cli.rs`(존재 확인)" | `tests/install_cli.rs`(tip 존재), `docs/e2e-receipts/receipts.jsonl`(RL-* receipt 없음), `docs/support-matrix.md`, 검토 §3.6.6 |
| 7 (X11) | L737 — "새 Rust RC의 실행 파일·업데이트 경로·상태는 기존 Go 설치와 분리한다. …" 및 L739 — "테스트 완료 후 기존 Go 설치를 제거하고 …" | [83f] L18·L2611과 같은 문단 | **한 교체문 · 세 위치 적용** — 홈·프로젝트 root 분리 문단(L737)은 group3-83f162ef.md X11 절의 공통 교체문을 동일 적용, Go 전환 문장(L739)은 같은 문서 표 5번의 X12 배너와 83f 표 8번 교체문을 동일 적용 | 검토 §8 X11·X12 |

## 시험 ID 행 RC1 상태 (AC-2)

| ID | RC1 상태 | 근거 |
| --- | --- | --- |
| RL-GENERATION-ISOLATION | not_run (부분 관측: AC-08C-1·7 receipt 내부) | receipts.jsonl에 RL-* 없음 |
| RL-HOST-EXPLICIT-SWITCH · RL-COMMAND-CUTOVER · RL-CUTOVER-RESTORE | not_run, RC1 범위 밖 | decisions:89-91(04-D 후속) |
| DT-* (Desktop) | not_run | Desktop은 RC1 후속 |

## 비고

- 가설 단독 근거 0, 개인 절대 경로·비밀 0건(AC-3).

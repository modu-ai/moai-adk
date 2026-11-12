# group3-83f162ef 수정안 — MoAI 홈 설치와 프로젝트 생성 기능 계약

- 대상: `references/chatgpt-space/page_83f162ef14d08191917913738184f167.md` (읽기 전용 입력)
- 검토 근거: `review-group3-install-host-plan-run-release.md` §1·§8(X1·X5·X8·X11·X13·X14)·§9.3
- 고정 근거 tip: **86762b8e1** (처음 작성 기준은 604792666). P09 현재 교체문은 통합 base **8c5732615**, #2016 상태는 2026-11-10 독립감사 `6759943258`·리더 회신으로 보강했다. 기존 고정 근거와 현재 보강은 구분한다.
- 우선순위: S3(§1 경로)·S8(F ID), 검토 §9.3 "실물과 같은 것은 installations/뿐"

## 수정 항목

| # | 위치 | 현재 | 교체 | 근거 (tip 86762b8e1 재확인) |
| --- | --- | --- | --- | --- |
| 1 (S3, X1) | §1 "파일 또는 경로" 표(문서 머리 L37 `<state_root>/skills/moai/<version>/SKILL.md` 행 이후의 설치 경로 표) | 홈 설치 경로를 `<state_root>/skills/moai/<version>/` 등 가칭 체계로 표기 | 표의 RC1 실물 행을 교체: "설치 receipt `~/.moai-cli/installations/<install_id>.json`, host projection receipt `~/.moai-cli/installations/<install_id>/hosts/<host>.json`, known-good 사본 `~/.moai-cli/installations/<install_id>/hosts/<host>/known-good/<sha256 hex>`. receipt·outbox·claim journal은 프로젝트·clone 쪽이 기본이다. #2175의 `layout-plan`·`layout-apply`로 명시 opt-in하면 `~/.moai-cli/projects/<project_id>/{receipts, forge/outbox, claims}`를 권위 store로 쓴다. 결정안(제안) 상태와 기본값은 유지하며, 홈 패키지 루트·registry·runs 등 더 넓은 배치는 후속(P09)이다" | 고정 근거: `86762b8e1:src/install.rs:3-5`, `86762b8e1:src/update.rs`, `86762b8e1:src/claim.rs:5-9`, `86762b8e1:src/apply.rs:29-31`. 현재 opt-in: `8c5732615:docs/home-layout-p09.md:3-44`, `8c5732615:src/layout.rs`; 검토 §8 X1 |
| 2 | §6 hook 절 (문서 후반 "MoAI는 꼭 필요한 최소 hook만 host별 공식 event·payload·결과·timeout 계약에 맞춰 제공한다" 단락) | hook 설계가 RC1 시험 대상처럼 읽힘 | 절 머리에 아래 「hook·TOML 공통 배너」를 추가한다 | `docs/decisions-2026-11-08.md` L66·L67(02-B·02-C: rc.1 hook 0개·TOML 없음), L121(toml_edit 채택), `src/hooks.rs:1-20`(rc.2 I1 #2238 hook 생성·`host_loaded: not_run`), `docs/rust-product-path.md:504`(Codex config writer — 라이브러리만, CLI op 없음)·`:508`(hook 생성 절), 검토 §1.6.2·§8 X8 |
| 3 | §F 기능 ID 목록(L941-1046 전후) — F20·F22·F23·F24·F25 | 상위 §20과 같은 번호·다른 기능(예: [83f] F23 "서명·버전" vs 상위 F23 Desktop) | F ID 목록 머리에 경고·대응 추가: "**경고: F20·F22–F25는 상위 §20과 같은 번호·다른 기능이다**(예: 이 문서 F23=릴리스 서명·버전, 상위 §20 F23=Desktop). 누락 검사는 번호만 하지 말고 이름·source Page를 함께 대조한다([66a] §"기능 누락 검사"와 동일 규칙)." | 검토 §8 X13, 상위 §20 L2407-2434(검토 인용), [66a] page "기능 누락 검사는 번호만으로 하지 않는다" 단락 |
| 4 | §7 시험 ID 표(HOME-*·SETUP-*·GENERATION-* 행) | 시험 ID별 RC1 상태 미표기 | 각 행에 RC1 상태 추가: "HOME-*/SETUP-*/GENERATION-*: RC1 receipt 대응 없음(not_applicable). RC1의 실측 receipt는 AC-08C-1(passed — 임시 HOME SETUP→UPDATE→중단→복구 전 경로)·AC-08C-2(passed)·HA-02(passed)·AC-08C-7(passed, macOS)·AC-08C-8(passed, T1 절반)·AC-08C-5(passed) 6건과 AC-08C-3·4·6 not_run 3건뿐이다" | `docs/support-matrix.md`(지원 6행·not_run 표), `docs/e2e-receipts/receipts.jsonl`(9행, 2026-11-10 재확인 — 변화 없음) |
| 5 | §예시 JSON(L1101 전후) — `"profile": "personal-local"`, `"task_backend": "host-session"` | `src/profile.rs`가 관리하지 않는 키 | 키를 관리 키로 교체하거나 표시: `"hosts"`·`"development_methods"`·`"package_pin"` 등 관리 키로 바꾸고, 예시 구조가 스키마 미정이면 JSON 상단에 `"//": "스키마 미정 — 아래 키는 RC1 profile 관리 키(PROFILE_KEYS)와 무관한 예시"` 표시 | `src/profile.rs:36-50`(PROFILE_KEYS 13개: schema_version·project_id·purpose·tools·development_methods·hosts·package_pin·interview_revision·decisions + 추가형 github_issues·leader·context_max_bytes·archive_grace_days — #2254에서 옛 10개 목록을 대체). 2026-11-12 tip `325cc2bf03`에서는 추가형 `language`(#2293)·`detected`(#2496)가 더해져 15개다(`325cc2bf03:src/profile.rs:36-52`). 교체문은 키 개수를 적지 않으므로 그대로 유효하다, 검토 §1.6.5 |
| 6 | 상태 어휘(L1223) — "discovered → selected → planned → previewed → authorized → staged → applied → host_verified …" | 8단계 어휘가 RC1 receipt·admission·claim 상태와 대응 없음(S4) | 단락 뒤에 대응표 추가(S4 — DR-07 부록 대응표 참조): "RC1 대응 — 설치/UPDATE 결과 상태는 receipt `pending`·`committed`·`recovered`(update.rs), PLAN 승인은 `ready`·`stale`·`blocked`(plan.rs, 매 요청 재계산·무저장), 실행 claim은 ACTIVE_STATES 8개(claim.rs). 위 8단계는 설계 어휘이며 RC1 상태로 직접 승격되지 않는다" | `src/update.rs:253-255`(status 주석), `src/plan.rs:3-6`, `src/claim.rs:58-67`, DR-07 부록(상태 어휘 대응표 소유) |
| 7 (X11) | L18(문서 머리 상태 root 단락) · L2597(홈 hook 소유권 단락) · L2611(Go 전환 단락) | 세 문서에 붙여 넣은 같은 문단 — [25f]·[487]·[dff]의 동일 문단과 어긋남 | **한 교체문(아래)을 세 위치에 동일 적용** + 이 문서의 L2611(Go 전환) 문장은 X12에 따라 별도 교체(아래 8번) — 공통 문단 교체문은 group3-25fa9b23.md·group3-487fa473.md에도 동일 인용됨 | 검토 §8 X11 행 |
| 8 (X12) | L2611 — "테스트 완료 후 기존 Go 설치를 제거하고 새 버전을 moai 명령으로 전환하는 계획을 따른다." | Go 전환을 RC1 내 계획처럼 서술 | "**Go 전환(cutover)은 RC1 범위 밖(04-D)** — RC1에서는 기존 Go 상태·설치본의 자동 쓰기·이동·삭제가 0이며, 위 문장은 전환 계획의 요구사항 나열로만 읽는다. #2016은 2026-11-11에 재설계 범위(합성 HOME 한정 백업·선택 이전·복구와 CLI 연결, PR #2449)로 닫혔고, 실제 HOME 이전과 Go `moai update` 브리지는 #2459(HOLD, 사용자 결정 대기)로 분리됐다. 실행 시점·owner·검증은 그 Issue와 최신 사용자 결정으로 별도 확인한다. 어느 쪽도 실제 Go 자산 변경 승인은 아니다." | 과거 범위 근거: `86762b8e1:docs/decisions-2026-11-08.md` L94 ②·L90. 현재 Issue 상태(2026-11-12 재확인): #2016 CLOSED 2026-11-11T09:58Z(리더 종료 6792799989, 대조표 6792572670, 병합 `313c7f186f`), #2459 OPEN·`needs:decision`, 검토 §8 X12 |

## hook·TOML 공통 배너 (이 문서 2번 · dff 1·2·6번 · 487 7번 · c4f9 6번이 인용)

> "**rc.1 기준: Claude hook 0개·Codex hook 0개·TOML 처리 없음**(02-B·02-C 임시 결정). **rc.2에서 바뀜**: `hooks-plan`·`hooks-apply`·`hook-receipt`가 Claude `.claude/settings.json`의 `hooks` 키와 Codex `.codex/hooks.json`에 고정 hook 집합(`SessionStart`·`PreToolUse`·`PostToolUse`·`Stop`)을 생성한다(#2238). host가 생성된 hook을 실제로 읽는지는 **not_run**이다. Codex `.codex/config.toml` 편집은 `toml_edit` 0.25.3 라이브러리로만 있고 CLI 명령은 아직 없다(#2240). 이 절은 HOOK-EVENT-PROBE 설계이며 수용 시험이 아니다(§7 HOOK-EVENT-PROBE도 수용 시험 아님)."

## X11 공통 문단 — 한 교체문 · 세 위치 적용

대상 위치: **[83f] L18 = [25f] L737 = [487] L156**(홈·프로젝트 root 분리), **[83f] L2597 = [dff] L827 = [487] L792**(홈 hook 소유권), **[83f] L2611 = [25f] L739**(Go 전환).

공통 교체문(홈·프로젝트 root 분리 문단용 — 세 위치 동일 적용):

> "Rust RC의 사용자 홈 상태 root는 `~/.moai-cli/`(설치 receipt `installations/<install_id>.json`, host receipt `…/hosts/<host>.json`, known-good `…/known-good/<hex>`)다. **기본 배치**는 claim 저널이 공통 dir `.moai-cli/claims/`, 프로젝트 데이터가 `<project>/.moai-cli/`(receipts·forge outbox)다. **구현된 명시 opt-in(#2175)**: `layout-plan`·`layout-apply`로 전환하면 `~/.moai-cli/projects/<project_id>/{receipts, forge/outbox, claims}`를 권위 store로 쓰며 프로젝트 원본은 보존한다. 결정안(제안) 상태와 프로젝트 쪽 기본값은 바꾸지 않는다. 기존 Go의 `~/.moai/`·`<project>/.moai/`는 보존하며 새 RC가 기본 쓰기 대상으로 사용하지 않는다. worktree는 `<primary>/.moai-cli/worktrees/<issue>-<logical_key>`(`git worktree add --lock`만, 삭제 금지). `~/moai/` 경로와 `config.toml` 문장은 RC1 실물에 없는 초기 구현안 잔재다. 홈 패키지 루트·registry·runs 등 더 넓은 배치는 후속(P09)다."

Go 전환 문장(L2611·[25f] L739)은 위 표 8번의 X12 교체문을 세 위치 동일 적용. hook 소유권 문단([83f] L2597 = [dff] L827 = [487] L792)은 위 「hook·TOML 공통 배너」를 세 위치 머리에 동일 추가.

## 비고

- 가설 단독 근거 0, 개인 절대 경로·비밀 0건(AC-3).

## 갱신 기록 (#2207, tip 86762b8e1)

- 2번: "RC1 hook 0·TOML 미구현" 배너를 rc.1/rc.2 구분 공통 배너로 바꿨다. rc.2 hook 생성(#2238)과 toml_edit(L121)가 들어왔기 때문이다. 근거 "No hooks are generated." 문장은 지금 `docs/rust-product-path.md`에 없다.
- 5번: `MANAGED_KEYS` 10개 → `PROFILE_KEYS` 13개(`src/profile.rs:36-50`).
- 8번: 근거 "decisions:89-91"은 잘못된 줄이었다(그 줄은 레인 대기·제약 해제·worktree 도구 행). L94 ②·L90으로 고쳤다.
- 줄 번호 이동: `src/claim.rs:11-12`→`5-9`, `:50-59`→`58-67`.
- 6번: 셀 안의 `|`가 표를 깨뜨려 상태 값 구분을 `·`로 바꿨다(뜻은 같음).

## 독립감사 후 보강 (#2207, base 8c5732615)

- P09 opt-in의 세 store·권위 포인터·원본 보존과 기본 배치를 구분했다(`8c5732615:docs/home-layout-p09.md:3-44`). X11 공통 문단을 인용하는 [25f]·[487]도 같은 구분을 따른다.
- #2016 OPEN 재개 상태를 적되 과거 결정 기록을 삭제하거나 Go 자산 변경 승인을 추정하지 않았다.

## 현재 tip 재확인 (#2207, 2026-11-12, tip `325cc2bf03`)

- 인용 대조(`86762b8e1` → `325cc2bf03`, 같은 줄 범위의 글자 비교): 11건 중 그대로 6, 줄만 이동 3, 내용 바뀜 1, 모호 0, 고정 tip에 없던 파일 1(뒤 base `8c5732615` 근거). 인용은 각자 적힌 고정 tip 기준이라 그대로 두고, 교체문에 옮겨지는 사실만 다시 확인했다.
- 8번 교체문·근거: #2016 상태를 위와 같이 고쳤다.
- 5번 근거: PROFILE_KEYS가 지금 15개(`language` #2293·`detected` #2496 추가)임을 덧붙였다. 교체문은 키 개수를 적지 않아 바꾸지 않았다.
- 교체문에 들어간 다음 사실은 tip에서 다시 확인했고 그대로다: claim lease 기본 4시간·heartbeat 없음 (`src/claim.rs` `DEFAULT_LEASE`), `Budget`은 `max_attempts` 하나·기본값 없음(`src/run.rs`), E2E receipt 9건(6 passed + 3 not_run, `docs/e2e-receipts/receipts.jsonl`), GitLab은 capability 선언만 (`src/forge/gitlab.rs`, 호출 경로 없음), hook 생성의 host 로딩 `not_run`(`docs/rust-product-path.md` `hooks.apply` 행), `capabilities` 명령 없음, Codex config CLI 없음(#2237 PR #2534 미병합), #2197·#2014 OPEN. OPERATIONS 전수는 지금 99개다(고정 근거 `86762b8e1`에서는 70개). 70을 적은 근거 열은 그 고정 tip 기준이며, 교체문에는 개수가 없다.

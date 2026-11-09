# group3-83f162ef 수정안 — MoAI 홈 설치와 프로젝트 생성 기능 계약

- 대상: `references/chatgpt-space/page_83f162ef14d08191917913738184f167.md` (읽기 전용 입력)
- 검토 근거: `review-group3-install-host-plan-run-release.md` §1·§8(X1·X5·X8·X11·X13·X14)·§9.3
- 적용 tip: **604792666** (2026-11-09 재확인)
- 우선순위: S3(§1 경로)·S8(F ID), 검토 §9.3 "실물과 같은 것은 installations/뿐"

## 수정 항목

| # | 위치 | 현재 | 교체 | 근거 (tip 604792666 재확인) |
| --- | --- | --- | --- | --- |
| 1 (S3, X1) | §1 "파일 또는 경로" 표(문서 머리 L37 `<state_root>/skills/moai/<version>/SKILL.md` 행 이후의 설치 경로 표) | 홈 설치 경로를 `<state_root>/skills/moai/<version>/` 등 가칭 체계로 표기 | 표의 RC1 실물 행을 교체: "설치 receipt `~/.moai-cli/installations/<install_id>.json`, host projection receipt `~/.moai-cli/installations/<install_id>/hosts/<host>.json`, known-good 사본 `~/.moai-cli/installations/<install_id>/hosts/<host>/known-good/<sha256 hex>`. 홈 패키지 루트·registry·runs 레이아웃은 후속(P09) — RC1에서 경로 상수가 없다" | `src/install.rs:3-5`(installations/<install_id>.json), `src/update.rs`(hosts/<host>.json·known-good — #2091·#2103 병합분), `src/claim.rs:11-12`(claims common dir), `src/apply.rs:29-31`(receipts 프로젝트 로컬 + @MX:UPGRADE), 검토 §8 X1 |
| 2 | §6 hook 절 (문서 후반 "MoAI는 꼭 필요한 최소 hook만 host별 공식 event·payload·결과·timeout 계약에 맞춰 제공한다" 단락) | hook 설계가 RC1 시험 대상처럼 읽힘 | 절 머리에 한 줄 추가: "**RC1: Claude hook 0·Codex hook 0·TOML 처리 미구현** — 이 절은 HOOK-EVENT-PROBE 설계이며 RC1 시험이 아니다(§7 HOOK-EVENT-PROBE도 RC1 시험 아님)." | `docs/rust-product-path.md` Codex adapter 절 "No hooks are generated."(tip 재확인), 검토 §1.6.2·§8 X8, decisions(hook 0·MCP 다음 버전) |
| 3 | §F 기능 ID 목록(L941-1046 전후) — F20·F22·F23·F24·F25 | 상위 §20과 같은 번호·다른 기능(예: [83f] F23 "서명·버전" vs 상위 F23 Desktop) | F ID 목록 머리에 경고·대응 추가: "**경고: F20·F22–F25는 상위 §20과 같은 번호·다른 기능이다**(예: 이 문서 F23=릴리스 서명·버전, 상위 §20 F23=Desktop). 누락 검사는 번호만 하지 말고 이름·source Page를 함께 대조한다([66a] §"기능 누락 검사"와 동일 규칙)." | 검토 §8 X13, 상위 §20 L2407-2434(검토 인용), [66a] page "기능 누락 검사는 번호만으로 하지 않는다" 단락 |
| 4 | §7 시험 ID 표(HOME-*·SETUP-*·GENERATION-* 행) | 시험 ID별 RC1 상태 미표기 | 각 행에 RC1 상태 추가: "HOME-*/SETUP-*/GENERATION-*: RC1 receipt 대응 없음(not_applicable). RC1의 실측 receipt는 AC-08C-1(passed — 임시 HOME SETUP→UPDATE→중단→복구 전 경로)·AC-08C-2(passed)·HA-02(passed)·AC-08C-7(passed, macOS)·AC-08C-8(passed, T1 절반)·AC-08C-5(passed) 6건과 AC-08C-3·4·6 not_run 3건뿐이다" | `docs/support-matrix.md`(지원 6행·not_run 표), `docs/e2e-receipts/receipts.jsonl`(9행, 2026-11-09 읽음) |
| 5 | §예시 JSON(L1101 전후) — `"profile": "personal-local"`, `"task_backend": "host-session"` | `src/profile.rs`가 관리하지 않는 키 | 키를 관리 키로 교체하거나 표시: `"hosts"`·`"development_methods"`·`"package_pin"` 등 관리 키로 바꾸고, 예시 구조가 스키마 미정이면 JSON 상단에 `"//": "스키마 미정 — 아래 키는 RC1 MANAGED_KEYS와 무관한 예시"` 표시 | `src/profile.rs:31-42`(MANAGED_KEYS 10개: schema_version·project_id·purpose·tools·development_methods·hosts·package_pin·interview_revision·decisions·github_issues), 검토 §1.6.5 |
| 6 | 상태 어휘(L1223) — "discovered → selected → planned → previewed → authorized → staged → applied → host_verified …" | 8단계 어휘가 RC1 receipt·admission·claim 상태와 대응 없음(S4) | 단락 뒤에 대응표 추가(S4 — DR-07 부록 대응표 참조): "RC1 대응 — 설치/UPDATE 결과 상태는 receipt `pending|committed|recovered`(update.rs), PLAN 승인은 `ready|stale|blocked`(plan.rs, 매 요청 재계산·무저장), 실행 claim은 ACTIVE_STATES 8개(claim.rs). 위 8단계는 설계 어휘이며 RC1 상태로 직접 승격되지 않는다" | `src/update.rs:253-255`(status 주석), `src/plan.rs:3-4`, `src/claim.rs:50-59`, DR-07 부록(상태 어휘 대응표 소유) |
| 7 (X11) | L18(문서 머리 상태 root 단락) · L2597(홈 hook 소유권 단락) · L2611(Go 전환 단락) | 세 문서에 붙여 넣은 같은 문단 — [25f]·[487]·[dff]의 동일 문단과 어긋남 | **한 교체문(아래)을 세 위치에 동일 적용** + 이 문서의 L2611(Go 전환) 문장은 X12에 따라 별도 교체(아래 8번) — 공통 문단 교체문은 group3-25fa9b23.md·group3-487fa473.md에도 동일 인용됨 | 검토 §8 X11 행 |
| 8 (X12) | L2611 — "테스트 완료 후 기존 Go 설치를 제거하고 새 버전을 moai 명령으로 전환하는 계획을 따른다." | Go 전환을 RC1 내 계획처럼 서술 | "**Go 전환(cutover)은 RC1 범위 밖(04-D 후속)** — RC1에서는 기존 Go 상태·설치본의 자동 쓰기·이동·삭제가 0이며, 위 문장은 전환 계획의 요구사항 나열로만 읽는다. 실행 시점·owner·검증은 decisions:91 ②·89가 아니라 후속 Issue에서 정한다." | `docs/decisions-2026-11-08.md:89-91`(04-D RC1 밖·Go 상태 무변경), 검토 §8 X12 |

## X11 공통 문단 — 한 교체문 · 세 위치 적용

대상 위치: **[83f] L18 = [25f] L737 = [487] L156**(홈·프로젝트 root 분리), **[83f] L2597 = [dff] L827 = [487] L792**(홈 hook 소유권), **[83f] L2611 = [25f] L739**(Go 전환).

공통 교체문(홈·프로젝트 root 분리 문단용 — 세 위치 동일 적용):

> "Rust RC의 사용자 홈 상태 root는 `~/.moai-cli/`(설치 receipt `installations/<install_id>.json`, host receipt `…/hosts/<host>.json`, known-good `…/known-good/<hex>`, claim 저널은 공통 dir `.moai-cli/claims/`)이고, 프로젝트 데이터는 `<project>/.moai-cli/`(receipts·forge outbox)로 분리된다. 기존 Go의 `~/.moai/`·`<project>/.moai/`는 보존하며 새 RC가 기본 쓰기 대상으로 사용하지 않는다. worktree는 `<primary>/.moai-cli/worktrees/<issue>-<logical_key>`(`git worktree add --lock`만, 삭제 금지). `~/moai/` 경로와 `config.toml` 문장은 RC1 실물에 없는 초기 구현안 잔재다. 홈 패키지 루트·registry·runs 레이아웃은 후속(P09)."

Go 전환 문장(L2611·[25f] L739)은 위 표 8번의 X12 교체문을 세 위치 동일 적용. hook 소유권 문단([83f] L2597 = [dff] L827 = [487] L792)은 표 2번의 "RC1: hook 0" 한 줄을 세 위치 머리에 동일 추가.

## 비고

- 가설 단독 근거 0, 개인 절대 경로·비밀 0건(AC-3).

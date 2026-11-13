# 수정안 — MoAI ADK 2026.11.0-rc.1 온라인 우선 최종 계획 (page_43faf222)

- 원본: `references/chatgpt-space/page_43faf222414c81919de71f7d7830e26c.md` (완전본 5,523행; 아래 모든 L은 1,753행 이후 완전본 기준)
- 검토 출처: `references/chatgpt-space/review-group1-rc1-plan.md` §6 개선 제안 (영향 순) 1~12
- 근거 기준: **tip `86762b8e1`**(기존 고정 인용 보존, 처음 작성 기준은 604792666). 아래 명령 목록 정책·P09 현재 상태는 **base `8c5732615`**(2026-11-10, 독립감사 `6759943258`)에서 보강했으며 해당 인용에는 새 접두사를 적었다.
- 상위 문서 우선 규칙: 상위 §20(L2115-2116)에 따라 충돌 항목은 이 문서가 '역사'로 남는다 — 교체문은 결정을 되돌리지 않고 어디가 현행인지만 표시한다.

## 수정 항목 (영향 순)

| # | 위치 | 현재 | 교체 | 근거 |
|---|---|---|---|---|
| 1 | L3781 | `Issue의 assignee·label·댓글을 원자적 lock으로 사용하지 않는다.` | S1 표기 추가: `…원자적 lock으로 사용하지 않는다. RC1 결정(2026-11-08 16:4x)은 이 원칙과 원자 primitive 부재 전제를 유지하면서, Issue 댓글 id의 단조성에 기대는 eventual claim(최소 댓글 id 승자 + `Moai-Claim:` 토큰 + 리더 읽기 측 검증 + 만료 시 운영자 판단)을 가정에 기댄 선택으로 채택했다 — 댓글이 lock이 되는 것은 아니다. lease는 생략 시 4시간, heartbeat는 없고, 인계는 리더의 epoch+1 수동 인계뿐이다(2026-11-09).` | `86762b8e1:docs/decisions-2026-11-08.md` L21(P08 ⑤+①)·L102(00:5x 잔여: 수동 인계, lease 4h, heartbeat 없음), `86762b8e1:src/claim.rs:15-30`(⑤ primary·자동 takeover 금지·lease 기본 4h) |
| 2 | L5164 | `Git tag 후보는 v2026.11.0-rc.1, package version은 2026.11.0-rc.1, 채널은 next 또는 명시적인 RC 채널이다. 정확한 tag prefix는 기존 release automation과 충돌 여부를 확인한 뒤 고정한다.` | R5 교체: `Git tag 후보는 moai-cli-v2026.11.0-rc.1이며 게시는 --prerelease --latest=false로 한다. Go stable updater의 /releases/latest와 v/go-v 접두사 필터 어느 쪽에도 걸리지 않는다(04-A·E6 특성화 시험). 무접두사 v2026.11.0-rc.1은 Go RC 채널 필터에 걸려 검사 오류나 "업데이트 있음"으로 읽히므로 금지된다. 접미사 없는 정식판은 기본 게시(make_latest 기본 true)로 /releases/latest를 가져가 Go stable 사용자에게서 더 새 Go 릴리스를 가리므로, 처음부터 latest가 되지 않는 게시 경로가 필요하다(운영자 승인 항목). tag ruleset moai-cli-v*는 사용자 승인 1회로 만드는 별도 카드다.` L5444 "새 repo는 선택안" 행에도 결정 경계 표시(같은 repo 확정)를 추가한다 | `86762b8e1:docs/release-checklist.md:8`(tag 확정)·`:9·48·108`(tag ruleset 미적용)·`:46`(`/releases/latest` 실측 기록)·`:121`(`--prerelease --latest=false` 게시 절차), `86762b8e1:docs/decisions-2026-11-08.md` L9(15:5x tag namespace)·L107(01:1x tag ruleset 카드, 사용자 승인 1회), `86762b8e1:docs/release-attestation.md`(운영자 승인 항목 6 — 정식판 latest 격리), `86762b8e1:tests/release_suite.rs`(E6 #2198 — 평문 `v*`·정식판 latest 위험 재현), rc.2 E5 #2197(진행 중) |
| 3 | L4394-4410 | `moai-cli setup --dry-run --json` 등 내부 CLI 17줄 | S3 교체: 목록 앞에 배너를 두고 목록 유지 — `이 목록은 미출하 계약안이다. RC1이 출하한 내부 명령은 docs/design-review/appendix/interfaces.md의 "RC1 OPERATIONS 전수" 목록(코드와 시험으로 대조)이 전부이며 이 목록과 이름이 하나도 같지 않다.` 배너에 전수 목록이나 개수를 옮겨 적지 않고 그 문서를 가리킨다(rc.2 카드가 계속 더하기 때문) | 기존 실측: `86762b8e1:docs/design-review/appendix/interfaces.md` § "RC1 OPERATIONS 전수", `86762b8e1:src/main.rs:1866`(`operations_doc_matches_code`). 현재 목록·개수 미기재 정책: `8c5732615:docs/design-review/appendix/interfaces.md:20-95` |
| 4 | L4479·L4483·L4542·L5485 | 홈 경로 용어 4종(`<state_root>`·`~/.moai-cli/`·`<moai_home>`·`workspace_root`) 상호 정의 없음 | I1 교체: §20 머리에 정의 절을 하나 두고 — `RC1 현행: 홈 상태 root는 ~/.moai-cli(installations/, hosts/, bin/)이며, 기본 receipt·outbox·claim journal은 프로젝트·clone 쪽이다. #2175의 layout-plan·layout-apply로 명시적으로 opt-in하면 ~/.moai-cli/projects/<project_id>/{receipts, forge/outbox, claims}가 권위 store가 된다. 결정안(제안) 상태와 프로젝트 쪽 기본값은 유지하며, 이 구현을 P09 전체 기본 배치의 승인으로 읽지 않는다. 홈 패키지 루트·registry·runs 등 더 넓은 배치는 후속이다. <state_root>·<moai_home>·workspace_root는 정의 없이 쓰지 않는다` — 나머지 절은 참조만 한다 | 고정 근거: `86762b8e1:docs/rust-product-path.md:215`, `86762b8e1:src/install.rs:3-5`, `86762b8e1:docs/decisions-2026-11-08.md` L10. 현재 opt-in·기본값·제안 경계: `8c5732615:docs/home-layout-p09.md:3-44`, `8c5732615:src/layout.rs` |
| 5 | L2713 | `description에 맞는 자연어 요청의 선택률을 측정` — 임계값·표본·host 없음 | I2 교체: `선택률은 (표본: 테스트 저장소 N개·host별 분리), (임계값: baseline 확보 뒤 고정, 그 전까지는 수치 없이 기록만), (측정 방법: fixture 문장 집합에 대한 route 일치율)로 측정한다. baseline 전까지 이 acceptance는 기록 전용이다.` | 사본 43fa 실측 L2713(검토 §2 I2) |
| 6 | L5298·L5311·L5345 | M3·M4 완료 조건 정성 문장 + `정량 목표는 baseline 확보 뒤 고정` | I3 교체: M3·M4 행에 `측정 지표(이름 나열: stale/false PASS 차단 건수·비용·비밀 누수 0 등)를 명시하고, 수치는 baseline 확보 뒤 고정한다 — 그 전까지는 기록 전용`을 덧붙인다 | 사본 43fa 실측 L5298·L5311·L5345(검토 §2 I3) |
| 7 | L3347-3390·L3394·L3590 | 구독 크레딧 금액($100/$200/$500 …)·`현재 Decisions API는 실제 GA 제품이며 … gpt-6-luna` — 출처 URL·확인 시각 없음 | I5·I6 + AC-3 교체: `이 숫자·GA 문장은 2026-11-09 공식 페이지 조사(docs/judge-backend-research-2026-11-09.md)와 대조한다. 조사 표와 같은 값(예: OpenAI Pro $100·$200·$500 tier, Anthropic Max $100·Ultra $200, Decisions API GA·gpt-6-luna)은 그 조사의 URL·확인일을 붙여 남기고, 조사와 다르거나 조사에 없는 값은 "미결(외부 확인 필요)"로 바꾼다. 어떤 값도 모델 기억으로 채우지 않으며 vendor가 바꾸면 다시 읽는다.` §7 절 머리에 `RC1 비범위(허브 보류, 판단 backend는 조사 뒤 구현 카드 blocked)` 배너를 둔다 | `86762b8e1:docs/judge-backend-research-2026-11-09.md:12-55`(F6 #2204 — backend 요금·구독 credit·Decisions GA, 확인일 2026-11-09·URL), `86762b8e1:docs/decisions-2026-11-08.md` L24(16:4x 허브 보류)·L51(MCP 다음 버전)·L110(01:1x P10), 사본 43fa 실측 L3347-3390·L3590 |
| 8 | L5469-5471 | §27 수치(blob 17,478 등)·`별도 목록으로 제공`(대상 없음) | I7 교체: `수치 중 1,107(legacy asset)은 .moai-cli/manifests/legacy-asset-map.json으로 재현된다. 나머지 수치는 RC1 docs/baseline-rc1.md의 측정값으로 교체하거나 재현 명령을 붙인다. "별도 목록"의 실제 경로를 명시한다.` | `86762b8e1:.moai-cli/manifests/legacy-asset-map.json`(assets 1,107 — 검토 I7 인용), `86762b8e1:docs/baseline-rc1.md:69-72` |
| 9 | L1896·L2476·L5181 | 단계 목록 3종 불일치(5단계 / 8 intent / workflows 7종 READ 없음) | I9·P4 교체: 세 목록을 skill의 route 표(`skills/moai/references/routes.json`)로 통일한다 — 진입 단계 READ·SETUP·PROJECT·UPDATE, 개발 5단계 PLAN → RUN → VERIFY → AUDIT → SYNC, 그리고 `--mode-agent`·`--mode-leader`로만 고르는 모드 route MODE-AGENT·MODE-LEADER. route는 rc.2 카드가 더하므로 개수를 옮겨 적지 않고 그 표를 가리킨다. workflows 목록에 READ를 포함한다(`skills/moai/workflows/read.md` 존재). "5단계"라고 쓰는 곳은 개발 5단계를 뜻한다고 밝힌다. 상위 §2의 UPDATE 누락은 상위를 고친다 | `86762b8e1:skills/moai/SKILL.md:10`(당시 8 routes), 2026-11-12 tip `325cc2bf03:skills/moai/references/routes.json`(11 route: READ·SETUP·PROJECT·UPDATE·PLAN·RUN·VERIFY·AUDIT·SYNC·MODE-LEADER·MODE-AGENT — PROJECT는 #2496, 모드 route는 #2286·#2497)·`325cc2bf03:skills/moai/SKILL.md:10-11`, `86762b8e1:skills/moai/workflows/read.md`(존재), `86762b8e1:docs/decisions-2026-11-08.md` L116(01:2x 5단계 PLAN → RUN → VERIFY → AUDIT → SYNC, READ·SETUP·UPDATE는 진입 단계) |
| 10 | L4162-4174 | §14 `@MX` 6종만 서술(상위 §9 `@MOAI:`·Code Atlas와 무관) | P3 교체: §14 머리에 `상위 §9의 @MOAI: namespace는 @MX:와 함께 읽는 방식(dual-read)으로 결정됐고 전환 시점은 정하지 않았다. 묶음 명칭은 "MoAI Atlas"다(명칭만 승인 — 패키지명·CLI 서브커맨드·태그 문법은 미확정). rc.2 atlas-tag-scan이 두 namespace를 같은 규칙으로 읽는다. 본 절의 @MX 6종이 현행 태그다`를 붙인다 | `86762b8e1:docs/decisions-2026-11-08.md` L100(00:5x `@MOAI:` dual-read)·L119(01:2x 묶음 명칭 "MoAI Atlas", 나머지 미확정), `86762b8e1:src/atlas_tag.rs:1-25`(C2 #2178 dual-read 문법 v1), `86762b8e1:docs/rust-product-path.md` @MX 절 |
| 11 | L5236 | RC1 필수 gate(두 host 실제 발견·호출 등) — 충족 상태 링크 없음 | R10 교체: `RC1 gate 상태는 docs/support-matrix.md(certified not_run)·docs/e2e-receipts/receipts.jsonl(#2014)에 실시간으로 남긴다. 미충족(T2 실측·실제 GitHub 쓰기)은 문서 결함이 아니라 진행 상태다. 첫 RC 인증 조합은 macOS arm64 × Claude Code·Codex다(2026-11-09).` | `86762b8e1:docs/support-matrix.md:5`(certified not_run), `86762b8e1:docs/p07-decision.md:23`(Live writes not_run), `86762b8e1:docs/decisions-2026-11-08.md` L120(04:54 P04 첫 RC 인증 조합) |
| 12 | L2639 | `CWD/상위의 CLAUDE.md·.claude/CLAUDE.md·CLAUDE.local.md가 있으면 기본 AGENTS 로딩을 억제할 수 있다` | R12 교체: `억제 규칙의 정확한 범위(HOME 쪽 포함 여부)는 02-B 확정 규칙(#2044 재감사 결과)을 따른다 — 본 문장을 그 규칙과 대조해 갱신한다.` | `86762b8e1:docs/decisions-2026-11-08.md` L74(19:4x — #2044 HOME/.claude/CLAUDE.md 억제 오판 수정 지시) |

## 비고

- 기록 정리 대상(사용자 선택 아님): `docs/release-checklist.md:48`·`:108`은 `moai-cli-v*` tag ruleset을 아직 "HOLD — 사용자 결정 대상"으로 적는다. decisions L107(01:1x)은 이를 "생성 카드(사용자 승인 1회)"로 정했고 그 카드(E5 #2197)가 진행 중이다.

## 갱신 기록 (#2207, tip 86762b8e1)

- 2번: "tag ruleset 추가는 운영자 HOLD"는 L107(ruleset 생성 카드, 사용자 승인 1회 — E5 #2197)로 바뀌었다. 정식판이 `/releases/latest`를 가져가는 위험(E4 운영자 승인 항목 6, E6 시험)을 더했다. release-checklist 근거 줄을 옮겼다(`:43`→`46`, `:45`→`48`, `:62`→`108`, 게시 절차 `:121` 추가).
- 10번: "@MOAI·Code Atlas는 미결 후보"는 L100(`@MOAI:` dual-read 결정)·L119("MoAI Atlas" 명칭 승인)로 낡았다. 결정된 것과 미확정인 것을 나눠 적었다.
- 3번: "OPERATIONS 47개" 나열을 지우고 전수 목록 문서(70개)를 가리키게 했다.
- 7번: F6 조사(#2204, `docs/judge-backend-research-2026-11-09.md`)가 금액·GA를 확인했으므로 "전부 미결" 대신 조사와 대조하는 교체문으로 바꿨다.
- 9번·11번: L116(5단계 이름·진입 단계)·L120(P04 첫 인증 조합)을 반영했다.
- 1번: lease 4h·heartbeat 없음·수동 인계(L102)를 더했다. 줄 번호 이동: `src/claim.rs:15-21`→`15-30`, `docs/rust-product-path.md:186`→`215`.
- 근거 접두사 `604792666:`을 `86762b8e1:`로 바꿨다.

## 독립감사 후 보강 (#2207, base 8c5732615)

- 붙여 넣을 배너에서 OPERATIONS 개수를 삭제했다. 과거 갱신 기록의 수치는 당시 고정 tip의 이력일 뿐 현재 전수 개수가 아니다.
- P09의 구현된 명시 opt-in과 프로젝트 쪽 기본값·결정안 상태·후속 전체 배치를 분리했다.

## 현재 tip 재확인 (#2207, 2026-11-12, tip `325cc2bf03`)

- 인용 대조(`86762b8e1` → `325cc2bf03`, 같은 줄 범위의 글자 비교): 16건 중 그대로 8, 줄만 이동 2, 내용 바뀜 5, 모호 0, 고정 tip에 없던 파일 1(뒤 base `8c5732615` 근거). 인용은 각자 적힌 고정 tip 기준이라 그대로 두고, 교체문에 옮겨지는 사실만 다시 확인했다.
- 9번 교체문·근거: "8 route로 통일"을 route 표(`skills/moai/references/routes.json`, 지금 11개 — PROJECT #2496, 모드 route #2286·#2497 추가)를 가리키는 문장으로 바꿨다. 3번과 같이 개수를 옮겨 적지 않는다.
- 교체문에 들어간 다음 사실은 tip에서 다시 확인했고 그대로다: claim lease 기본 4시간·heartbeat 없음 (`src/claim.rs` `DEFAULT_LEASE`), `Budget`은 `max_attempts` 하나·기본값 없음(`src/run.rs`), E2E receipt 9건(6 passed + 3 not_run, `docs/e2e-receipts/receipts.jsonl`), GitLab은 capability 선언만 (`src/forge/gitlab.rs`, 호출 경로 없음), hook 생성의 host 로딩 `not_run`(`docs/rust-product-path.md` `hooks.apply` 행), `capabilities` 명령 없음, Codex config CLI 없음(#2237 PR #2534 미병합), #2197·#2014 OPEN. OPERATIONS 전수는 지금 99개다(고정 근거 `86762b8e1`에서는 70개). 70을 적은 근거 열은 그 고정 tip 기준이며, 교체문에는 개수가 없다.

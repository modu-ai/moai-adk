# group3-dff088ea 수정안 — MoAI Host adapter와 hook 이식성

- 대상: `references/chatgpt-space/page_dff088ea648081918e2eb28ab4a3e77c.md` (읽기 전용 입력)
- 검토 근거: `review-group3-install-host-plan-run-release.md` §2·§8(X8·X10·X11)·§9.4·§9.6
- 적용 tip: **86762b8e1** (2026-11-10 재확인, #2207 — 처음 작성 기준은 604792666)
- 우선순위: S5(§9 HA별 RC1 상태)

## 수정 항목

| # | 위치 | 현재 | 교체 | 근거 (tip 86762b8e1 재확인) |
| --- | --- | --- | --- | --- |
| 1 (S5) | §9 HA 행들(HA-01~HA-10) | HA별 RC1 상태 미표기 — 검토 §9.4 "HA-05·07·08·09가 시험 불가인데 그 표시가 없다" | 각 HA 행 머리에 RC1 상태 열 추가. 판정 규칙(2026-11-10 receipts.jsonl·support-matrix 기준): **HA-02 = passed**(receipt `docs/e2e-receipts/receipts.jsonl`, ran_at 2026-11-08T15:04:16Z, T1 fixture — UPDATE 재실행 diff 0). **hook 계열 HA(HA-05 event sentinel, HA-07 Codex hooks.json 중복·hash trust, HA-08 PermissionRequest) = not_run** — rc.1에는 hook이 없었고, rc.2에서 hook 생성은 들어왔지만(#2238) host가 실제로 읽고 호출하는지는 측정하지 않았다. **MCP 계열 HA(HA-09) = not_applicable**(MCP는 다음 버전 — 시험 대상 자체가 없다). **나머지 HA = not_run**(RC1 receipt 없음 — `docs/support-matrix.md` 지원 6행은 AC-08C-1·2·5·7·8·HA-02뿐) | `docs/e2e-receipts/receipts.jsonl`(9행 — HA-02 passed만 존재), `docs/support-matrix.md`(지원 6행 + not_run 표), `docs/decisions-2026-11-08.md` L51(HOLD-C1 MCP 다음 버전)·L67(rc.1 hook 0개), `src/hooks.rs:1-20`(rc.2 hook 생성, `host_loaded: not_run`), 사본 dff L592-709(HA 정의), 검토 §8 X8 |
| 2 (X8) | §5 hook 절과 §7 "MCP는 선택적 1급 adapter" 절 머리 | hook·MCP가 RC1 시험 대상처럼 서술 | §5 hook 절 머리에는 group3-83f162ef.md 「hook·TOML 공통 배너」를 넣는다. §7 MCP 절 머리에는 "**MCP는 다음 버전(RC1 범위 밖). 이 절은 설계이며 수용 시험 대상이 아니다.**" 배너를 넣는다 | `docs/decisions-2026-11-08.md` L51(HOLD-C1)·L66·L67·L121, `src/hooks.rs:1-20`, `docs/rust-product-path.md:508`(hook 생성 절), 검토 §8 X8·§9.4 |
| 3 | §검증 목록(문서 후반, L942) — "SOURCE-DEFAULT, STYLE-MODE, STYLE-CODING-BOUNDARY, INSTRUCTION-BUDGET, STATUS-FIELD-PARITY, USER-UI-PRESERVE, HOOK-LOAD-EVENT, CLOUD-GATE를 각각 실제 앱 시험으로 확인한다. 20개 후속 시험은 모두 미실행 상태이며…" | 이름을 댄 8개 + "20개" 주장의 불일치(8 vs 20) | ID 목록을 완성하거나 수를 고친다: "후속 시험 ID 전부를 나열한다(이름이 확인된 8개: SOURCE-DEFAULT·STYLE-MODE·STYLE-CODING-BOUNDARY·INSTRUCTION-BUDGET·STATUS-FIELD-PARITY·USER-UI-PRESERVE·HOOK-LOAD-EVENT·CLOUD-GATE — 나머지 N개는 ID 미정). 전부 미실행이며 RC1 receipt도 없다(not_run)." — 20의 근거가 없으면 "20개"를 "ID가 확인된 8개 + 미정 나머지"로 교체 | 검토 §2.6.3·§2.7("20개" 집계 기준 확인 불가), 본문 관측(2026-11-09, 위 인용) |
| 4 (X10) | 문서 머리(L18) "기존 Claude Code와 Codex Desktop 각각에서 …" 및 Desktop 관련 절 | "Codex Desktop"과 "Codex CLI" 혼용 — RC1 실측 스크립트는 CLI만 | 표기 구분 + 실측 부재 명시: "이 문서의 Codex Desktop 표현은 앱 UI 경로다. RC1의 실측 스크립트는 Codex CLI(`tests/e2e/host-codex.sh`, `codex exec --json --ephemeral --ignore-user-config --sandbox read-only`)이고 Desktop·CLI 모두 **not_run**(운영자 실측 대기)이다. Desktop 전용 시험은 존재하지 않는다." | `tests/e2e/host-codex.sh`(tip 관측), `docs/support-matrix.md`(Codex not_run 행), `docs/decisions-2026-11-08.md` L66, 검토 §8 X10 |
| 5 | §수치들 — "12 events, 65.6KB, 22.5KB, 32 KiB, v2.1.277" vs "13 harness 소비자, Codex 기본 14항목, t1909" | 검증 수치와 미검증 수치(가설)의 구분 없음 | 미검증 수치에 표기 추가: "13 harness 소비자·Codex 기본 14항목·t1909는 **미검증(가설 — Go 소스 검색 0건)**, 12 events·65.6KB·22.5KB·32 KiB·v2.1.277은 검증됨(공식 문서 재조회 2026-11-08)" | 검토 §2.6.5·§2.7(검색 0건 명시) |
| 6 (X11) | L827 — "MoAI는 꼭 필요한 최소 hook만 host별 공식 event·payload·결과·timeout 계약에 맞춰 제공한다. …" 단락 | [83f] L2597·[487] L792와 같은 문단 — 세 위치 동기 수리 대상 | **한 교체문 · 세 위치 적용**(교체문 본문은 group3-83f162ef.md의 X11 절 인용) + 이 문서 위치에는 group3-83f162ef.md 「hook·TOML 공통 배너」를 단락 머리에 동일 추가 | 검토 §8 X11 행("한 곳을 고치면 세 곳을 같이 고쳐야 한다") |

## 시험 ID 행 RC1 상태 (AC-2)

| ID | RC1 상태 | 근거 |
| --- | --- | --- |
| HA-02 | passed | receipts.jsonl(2026-11-08T15:04:16Z, T1 fixture)·support-matrix 지원 행 |
| HA-05·HA-07·HA-08(hook 계열) | not_run | rc.2 hook 생성(#2238)은 있으나 host 로딩·호출 미실측(`host_loaded: not_run`) |
| HA-09(MCP 계열) | not_applicable | MCP 다음 버전(L51 HOLD-C1) — 시험 대상 부재 |
| HA-01·03·04·06·10 | not_run | receipts.jsonl에 없음(지원표 not_run 원칙) |
| AC-08C-3·4·6 | not_run | receipts.jsonl not_run 3행(운영자 실측 대기) |

## 비고

- 가설 단독 근거 0(미검증 수치는 "미검증" 표기로 구분 — AC-3 충족), 개인 절대 경로·비밀 0건.

## 갱신 기록 (#2207, tip 86762b8e1)

- 1번·시험 ID 표: hook 계열을 HA-05·07·08로 바로잡았다(사본 L644·670·683 정의 확인). rc.2 hook 생성(#2238)이 들어와 "시험 대상 없음(not_applicable)"이 아니라 "host 로딩 미실측(not_run)"이 됐다. MCP 계열은 HA-09 하나만 not_applicable로 남겼다.
- 2번·6번: "hook 0개" 배너를 83f의 rc.1/rc.2 공통 배너 인용으로 바꿨다. 근거 "No hooks are generated." 문장은 지금 `docs/rust-product-path.md`에 없다.
- decisions 인용을 줄 번호(L51·L66·L67·L121)로 명시했다.

## 현재 tip 재확인 (#2207, 2026-11-12, tip `325cc2bf03`)

- 인용 대조(`86762b8e1` → `325cc2bf03`, 같은 줄 범위의 글자 비교): 2건 중 그대로 1, 줄만 이동 1, 내용 바뀜 0, 모호 0, 고정 tip에 없던 파일 0(뒤 base `8c5732615` 근거). 인용은 각자 적힌 고정 tip 기준이라 그대로 두고, 교체문에 옮겨지는 사실만 다시 확인했다.
- 교체문에서 바뀐 사실 없음.
- 교체문에 들어간 다음 사실은 tip에서 다시 확인했고 그대로다: claim lease 기본 4시간·heartbeat 없음 (`src/claim.rs` `DEFAULT_LEASE`), `Budget`은 `max_attempts` 하나·기본값 없음(`src/run.rs`), E2E receipt 9건(6 passed + 3 not_run, `docs/e2e-receipts/receipts.jsonl`), GitLab은 capability 선언만 (`src/forge/gitlab.rs`, 호출 경로 없음), hook 생성의 host 로딩 `not_run`(`docs/rust-product-path.md` `hooks.apply` 행), `capabilities` 명령 없음, Codex config CLI 없음(#2237 PR #2534 미병합), #2197·#2014 OPEN. OPERATIONS 전수는 지금 99개다(고정 근거 `86762b8e1`에서는 70개). 70을 적은 근거 열은 그 고정 tip 기준이며, 교체문에는 개수가 없다.

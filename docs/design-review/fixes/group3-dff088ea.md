# group3-dff088ea 수정안 — MoAI Host adapter와 hook 이식성

- 대상: `references/chatgpt-space/page_dff088ea648081918e2eb28ab4a3e77c.md` (읽기 전용 입력)
- 검토 근거: `review-group3-install-host-plan-run-release.md` §2·§8(X8·X10·X11)·§9.4·§9.6
- 적용 tip: **604792666** (2026-11-09 재확인)
- 우선순위: S5(§9 HA별 RC1 상태)

## 수정 항목

| # | 위치 | 현재 | 교체 | 근거 (tip 604792666 재확인) |
| --- | --- | --- | --- | --- |
| 1 (S5) | §9 HA 행들(HA-01~HA-10) | HA별 RC1 상태 미표기 — 검토 §9.4 "HA-05·07·08·09가 시험 불가인데 그 표시가 없다" | 각 HA 행 머리에 RC1 상태 열 추가. 판정 규칙(2026-11-09 receipts.jsonl·support-matrix 기준): **HA-02 = passed**(receipt `docs/e2e-receipts/receipts.jsonl`, ran_at 2026-11-08T15:04:16Z, T1 fixture — UPDATE 재실행 diff 0). **hook·MCP 관련 HA(§5·§9의 HA-05·07·09) = not_applicable**(RC1: hook 0개·MCP 다음 버전 — 시험 대상 자체가 없다). **나머지 HA = not_run**(RC1 receipt 없음 — `docs/support-matrix.md` 지원 6행은 AC-08C-1·2·5·7·8·HA-02뿐) | `docs/e2e-receipts/receipts.jsonl`(9행 — HA-02 passed만 존재), `docs/support-matrix.md`(지원 6행 + not_run 표), decisions(MCP 다음 버전·hook 0), 검토 §8 X8 |
| 2 (X8) | §5 hook 절과 §7 "MCP는 선택적 1급 adapter" 절 머리 | hook·MCP가 RC1 시험 대상처럼 서술 | 두 절 머리에 배너: "**RC1 범위 밖 — hook 0개(Claude·Codex 모두), MCP는 다음 버전. 이 절은 설계이며 RC1 수용 시험 대상이 아니다.**" | `docs/rust-product-path.md` Codex adapter 절 "No hooks are generated."(tip 재확인), decisions(MCP 다음 버전), 검토 §8 X8·§9.4 |
| 3 | §검증 목록(문서 후반, L942) — "SOURCE-DEFAULT, STYLE-MODE, STYLE-CODING-BOUNDARY, INSTRUCTION-BUDGET, STATUS-FIELD-PARITY, USER-UI-PRESERVE, HOOK-LOAD-EVENT, CLOUD-GATE를 각각 실제 앱 시험으로 확인한다. 20개 후속 시험은 모두 미실행 상태이며…" | 이름을 댄 8개 + "20개" 주장의 불일치(8 vs 20) | ID 목록을 완성하거나 수를 고친다: "후속 시험 ID 전부를 나열한다(이름이 확인된 8개: SOURCE-DEFAULT·STYLE-MODE·STYLE-CODING-BOUNDARY·INSTRUCTION-BUDGET·STATUS-FIELD-PARITY·USER-UI-PRESERVE·HOOK-LOAD-EVENT·CLOUD-GATE — 나머지 N개는 ID 미정). 전부 미실행이며 RC1 receipt도 없다(not_run)." — 20의 근거가 없으면 "20개"를 "ID가 확인된 8개 + 미정 나머지"로 교체 | 검토 §2.6.3·§2.7("20개" 집계 기준 확인 불가), 본문 관측(2026-11-09, 위 인용) |
| 4 (X10) | 문서 머리(L18) "기존 Claude Code와 Codex Desktop 각각에서 …" 및 Desktop 관련 절 | "Codex Desktop"과 "Codex CLI" 혼용 — RC1 실측 스크립트는 CLI만 | 표기 구분 + 실측 부재 명시: "이 문서의 Codex Desktop 표현은 앱 UI 경로다. RC1의 실측 스크립트는 Codex CLI(`tests/e2e/host-codex.sh`, `codex exec --json --ephemeral --ignore-user-config --sandbox read-only`)이고 Desktop·CLI 모두 **not_run**(운영자 실측 대기)이다. Desktop 전용 시험은 존재하지 않는다." | `tests/e2e/host-codex.sh`(tip 관측), `docs/support-matrix.md`(Codex not_run 행), decisions:66, 검토 §8 X10 |
| 5 | §수치들 — "12 events, 65.6KB, 22.5KB, 32 KiB, v2.1.277" vs "13 harness 소비자, Codex 기본 14항목, t1909" | 검증 수치와 미검증 수치(가설)의 구분 없음 | 미검증 수치에 표기 추가: "13 harness 소비자·Codex 기본 14항목·t1909는 **미검증(가설 — Go 소스 검색 0건)**, 12 events·65.6KB·22.5KB·32 KiB·v2.1.277은 검증됨(공식 문서 재조회 2026-11-08)" | 검토 §2.6.5·§2.7(검색 0건 명시) |
| 6 (X11) | L827 — "MoAI는 꼭 필요한 최소 hook만 host별 공식 event·payload·결과·timeout 계약에 맞춰 제공한다. …" 단락 | [83f] L2597·[487] L792와 같은 문단 — 세 위치 동기 수리 대상 | **한 교체문 · 세 위치 적용**(교체문 본문은 group3-83f162ef.md의 X11 절 인용) + 이 문서 위치에는 표 2번의 "RC1 범위 밖 — hook 0개" 배너를 단락 머리에 동일 추가 | 검토 §8 X11 행("한 곳을 고치면 세 곳을 같이 고쳐야 한다") |

## 시험 ID 행 RC1 상태 (AC-2)

| ID | RC1 상태 | 근거 |
| --- | --- | --- |
| HA-02 | passed | receipts.jsonl(2026-11-08T15:04:16Z, T1 fixture)·support-matrix 지원 행 |
| HA-05·HA-07·HA-09(hook·MCP 계열) | not_applicable | RC1 hook 0·MCP 다음 버전 — 시험 대상 부재 |
| HA-01·03·04·06·08·10 | not_run | receipts.jsonl에 없음(지원표 not_run 원칙) |
| AC-08C-3·4·6 | not_run | receipts.jsonl not_run 3행(운영자 실측 대기) |

## 비고

- 가설 단독 근거 0(미검증 수치는 "미검증" 표기로 구분 — AC-3 충족), 개인 절대 경로·비밀 0건.

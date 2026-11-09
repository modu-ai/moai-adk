# MCP 설계 준비 — 정본·결정의 요구, 빈 부분, 사용자 결정 질문 (2026-11-09)

- Issue: #2243 [rc.2][I6] (상위 EP-I #2237). 범위는 리더 댓글 6732287215(사용자 직접 지시 ADK_MCP_CLEAN_REDESIGN_20261109_01)대로 **설계 준비만**이다. 코드·`.mcp.json`·설정은 하나도 만들지 않는다.
- 작성: agent-1 [0dd806], 세션 `113e4ddd-566b-4bd7-b27d-f6cc8c19dcf1`, 기준 tip `cf4e662aff5b33ceb96c47757eafa9eb6f250f20`
- 이 문서는 근거를 모으고 질문을 정리한다. MCP를 할지, 어떤 모양으로 할지는 정하지 않는다.

## 1. 범위와 방법

읽은 자료(모두 읽기 전용, grep으로 찾은 행을 직접 열어 확인):

| 자료 | 위치 | 무결성 |
| --- | --- | --- |
| 설계 정본(2026-11-09 Downloads 내보내기, text-only) | `references/chatgpt-space/downloads-design-2026-11-09-textonly.md`(저장소에 넣지 않는 원문 사본, 862행) | sha256 `b4242abda1b116f7a38ac81530961ce53a0ded9216a95f92b23de8ba8be7d542`. `docs/design-review/appendix/state-vocabulary-crosswalk.md:66`에 기록된 `b4242abd…7d542`와 같음 |
| 정본 §20이 계승한다고 밝힌 RC1 상세 페이지 | RC1 handoff 묶음 `design-snapshot/`(2026-11-08 조회 사본) | 아래 5개 파일의 sha256이 묶음의 `SHA256SUMS.txt`와 같음: `00-page_43faf…` `3763028f…`, `02-page_83f1…` `71cf9942…`, `10-page_dff0…` `e3587548…`, `12-page_25fa…` `32de4736…`, `21-page_20fc…` `bda30b9d…` |
| 결정 기록 | `docs/decisions-2026-11-08.md` | tip `cf4e662af` |
| RC1 Issue 본문 거울 | `.moai-cli/reports/issue-bodies/1987-00-D.md` | tip `cf4e662af` |
| Rust RC 트리 | `src/` | tip `cf4e662af` |

일부러 읽지 않거나 근거로 쓰지 않은 자료(ADK_MCP_CLEAN_REDESIGN):

- 과거 Go MCP 구현(`moai mcp-server`, `internal/cli/mcp_server.go` 등)
- #2176 / PR #2230 parity 조사와 `docs/design-review/mcp-parity-rc1-vs-go-2026-11-09.md`
- `design-snapshot/21-page_20fc…` "MoAI MCP와 기존 기능 재사용 설계": 제목과 머리말(원문 수정 시각 2026-11-08T04:07:04Z)만 읽었다. 정본 §20 계승 목록에 없고, 제목부터 기존 기능 재사용이 목적이라 근거에서 뺐다. 이 페이지를 어떻게 다룰지는 §7 Q4에서 묻는다.

## 2. 정본 본문의 MCP 관련 행

정본 본문에서 `MCP`가 나오는 곳은 두 행이다(`grep -n -i mcp` → 644, 718). 리더 실측(댓글 6732287215 3항)과 같다. MoAI MCP 서버를 정의하는 절은 본문에 없다. MCP 결정과 맞닿는 행은 다음과 같다.

| 행 | 절 | 내용(요지·짧은 인용) |
| --- | --- | --- |
| L31 | §1 통합 결정 | 온라인 협업은 기존 GitHub·GitLab·CI를 쓰고 "새 MoAI 계정·서버·control plane·필수 DB 서버를 요구하지 않는다" |
| L73 | §2 구조 | Host Runtime Adapter가 "발견·호출·세션·이벤트·도구·승인·취소·hook 차이"를 소유한다 |
| L162 | §4 Claude와 Codex | 호출: "slash·plugin namespace는 실제 설치 상태에 따라 검증" |
| L631–644 | §14 Pi | Pi SDK/RPC 표의 "MCP·Codemode \| 현재 CLI에는 내장 확장 제공 \| SDK는 자동 적재가 아니므로 factory·binding 구성 시험". Pi runtime 안의 MCP이고, MoAI가 노출하는 MCP 서버가 아니다 |
| L660–662 | §14 Pi 보안 | 확장·tool_call 신뢰에 대해 정본이 적은 유일한 부분이다. 확장 allowlist, 고정 version, 실행 직전 policy 검증, 승인 인자와 실제 실행값 binding, 독립 audit log, OS 격리를 요구한다 |
| L718 | §17 수락 시험 | PI 시험군: "session switch·MCP SDK 차이 시험" |
| L726–744 | §18 구현 순서 | 1번이 공통 계약(identity·status·digest·approval·receipt schema, no-mutation read API)이다. L738: "새 조정 서버는 만들지 않는다" |
| L746–763 | §18 남은 결정 | MCP 항목이 없다 |
| L783 | §20 기존 설계 통합 | "현재 제품 결정은 [2026.11.0 rc1 온라인 우선 최종 계획](page_43faf…), 상세 Core Brain과 아래 기능별 문서를 통합한다. 최신 요구와 충돌하는 과거 선택은 역사로 남기며 현재 구현·승인으로 복원하지 않는다." 그 아래 계승 목록 12개 페이지 중에 Host adapter(page_dff0…), 릴리스와 Desktop(page_25fa…), 홈 설치와 SETUP(page_83f1…), SYNC와 Issues(page_0d45…), 검증 감사 보안(page_c4f9…)이 있다 |

## 3. 정본이 계승한 페이지의 MCP 정의

**리더 전제 보정:** "정본에 MoAI MCP 서버 정의 없음"은 정본 본문에 대해서는 맞다. 그러나 정본 L783은 RC1 최종 계획과 상세 페이지를 통합한다고 적고, 그 페이지들에는 MCP를 **선택적 1급 adapter**로 정의한 행이 있다. 같은 L783이 "충돌하는 과거 선택은 역사로 남긴다"고도 적기 때문에, 이 정의가 지금도 유효한지는 문서만으로 정해지지 않는다(§7 Q1).

| 페이지 | §20 계승 목록 | 행 | 내용(요지·짧은 인용) |
| --- | --- | --- | --- |
| `00-page_43faf…` RC1 온라인 우선 최종 계획 | L783이 "현재 제품 결정"으로 지목 | L86 | "MCP \| 선택적 1급 adapter \| 기존 구조화 호출 소비자를 보존·매핑하고 CLI 직접 실행과 함께 capability를 인증한다. remote portable 경로에 MCP 서버를 필수화하지 않는다" |
| 〃 | 〃 | L788 | "local stdio, remote API, MCP, browser는 같은 권한 의무를 지킨다" |
| 〃 | 〃 | L1240 | 구조 트리 `mcp/  # 선택적 1급 transport, 기존 소비자 compatibility` |
| 〃 | 〃 | L1290, L1335, L1341 | M0에 "MCP 소비자 mapping"이 있다. "자체 MCP 서버 호스팅이나 remote MCP 의존을 필수화하지 않음". 기존 소비자를 `port/wrap/retire/defer`로 분류한다 |
| 〃 | 〃 | L1385–1389 | MCP의 typed JSON 유지(TOON 일괄 교체 금지), "MCP structuredContent는 JSON object 계약을 유지" |
| `10-page_dff0…` Host adapter와 hooks | 있음 | L15 | "service-free는 MCP까지 제거한다는 뜻이 아니다. 기존 MCP 소비자를 지원하는 얇은 adapter는 선택적 1급 연결면으로 유지하되 자체 hosted backend·상주 운영 서비스를 제품 기본으로 만들지 않는다" |
| 〃 | 〃 | L127–144 (§7 "MCP는 선택적 1급 adapter") | CLI와 MCP는 같은 domain reducer·policy·evidence admission을 부르는 transport다(L129). L131은 기존 `mcp__moai__*` 소비자 13개 보고를 M0에서 재확인하라고 하고 "소비자 존재를 무시한 MCP 제거·일괄 배포 제외는 허용하지 않는다"고 적는다. L140은 새 remote MCP service를 기본 설치하지 않는다고 한다. L142는 CLI/MCP 동일 verdict golden fixture를, L144는 CLI에서 거부된 작업의 MCP 우회를 금지하고 MCP 등록·credential·지속 접근을 SETUP/UPDATE와 다른 승인 범위로 둔다 |
| 〃 | 〃 | L166 | 시험 HA-09 "기존 MCP consumers golden fixtures \| consumers별 유지/bridge/미지원 판정과 근거, CLI/MCP parity" |
| `12-page_25fa…` 릴리스와 Desktop | 있음 | L30, L38 | "CLI/MCP adapter … 같은 domain 결과". "MCP는 optional first-class adapter다. … 신규 remote MCP 서버 운영은 제품 기본이 아니며 MCP transport 제거도 기본이 아니다" |
| 〃 | 〃 | L141, L153, L212 | MCP review/factory/audit 소비자의 call graph parity, M0 "MCP call graph", RL-08 "기존 MCP 소비자·Codex 재지원 spike" |
| `02-page_83f1…` 홈 설치와 SETUP | 있음 | L128 | "stdout JSON/명시 JSONL과 stderr 진단, MCP structuredContent, native host schema는 각각의 계약을 유지한다" |
| 〃 | 〃 | L137 | "MCP \| 기존 38개 tool의 typed schema와 선택 소비자 계약 \| 노출 allowlist와 실제 action 권한 별도. 선택하지 않은 도구 자동 활성화 금지" |
| `06-page_0d45…` SYNC와 Issues | 있음 | L93, L95 | 실제 Rust argv와 MCP tool 이름은 adapter schema로 생성하며 이미 있는 명령이라고 안내하지 않는다. 기존 MCP 소비자는 새 실행 계약과 매핑한 뒤 선택 adapter로 유지한다 |
| `07-page_c4f9…` 검증 감사 보안 | 있음 | L74 | "새 MCP server나 OAuth grant는 단순한 skill 추가와 다른 지속 접근 행위다" |
| 〃 | 〃 | L90 | 내부 경계 `policy.authorize`·`verify.execute`·`evidence.admit`·`audit.reduce`·`review.propose`. "CLI·MCP·remote adapter는 같은 schema를 전달" |
| `01-page_cfa4…` (결정 체크리스트) | 목록에 없음(RC1 결정 페이지) | L240, L264 | D77(MCP typed JSON 유지, 결정 완료). **P06 미결**: "내부 typed protocol의 최종 schema·error/status·compatibility version 규칙과 MCP/CLI parity 범위" |

정리하면, 계승 페이지들은 MCP를 **선택적** 1급 adapter로 두고 다음 다섯 가지를 요구한다.

1. CLI와 같은 domain·policy·evidence를 쓴다.
2. MCP 서버나 remote 서비스를 필수로 만들지 않는다.
3. MCP 등록은 따로 승인받는 지속 접근이다.
4. CLI에서 거부된 작업을 MCP로 우회하지 않는다.
5. 기존 소비자를 매핑하거나 보존한다.

5번은 사용자 지시 ADK_MCP_CLEAN_REDESIGN("과거 구현을 참고·복사·이식·재사용 근거로 쓰지 않음")과 정면으로 부딪친다(§7 Q3).

## 4. 결정 기록과 RC1 원장

| 출처 | 내용 |
| --- | --- |
| `docs/decisions-2026-11-08.md` L16 (16:3x) | 통신 허브 전송 조합: "H-B 파일 큐(내구층) + H-A 로컬 브로커(실시간층); **H-C MCP 도구층은 후속**" |
| 〃 L51 (18:2x, 사용자 결정) | "**HOLD-C1 = A CLI만, MCP는 다음 버전**(AC-01C-2/3/4 not_applicable)". 같은 행에 "P06 revision·HOLD-C3는 MCP 후속과 함께 HOLD 유지"가 있다 |
| 〃 L99 (00:4x 11-09) | rc.2 발행 틀: "비범위 영역(§14 Pi/Desktop·F21·§16·MCP)도 전부 발행하되 '선행조건 아님/blocked by HOLD' 표시" |
| `.moai-cli/reports/issue-bodies/1987-00-D.md` L68 | **REQ-00D-07**(정의만 있고 구현은 없음): "WHERE MCP tool이 같은 operation을 노출하면, tool의 `structuredContent`는 CLI stdout envelope의 `result`와 byte-동등 … `outputSchema`는 같은 JSON Schema 파일 … 실행 오류는 `isError: true`로, 호출 형식 오류는 JSON-RPC protocol error로" |
| 〃 L80 | 00D-AC-6: host native 포맷 무변경. "`.claude/`, `.codex/`, `.mcp.json`에 변경 0" |
| 〃 L33, L52 | MCP 서버 구현과 CLI/MCP parity 실측은 01-C 소관이고, 지원 범위는 P06 HOLD다. MCP revision은 2026-07-28로 적혀 있지만 "조사 본문은 2025-06-18·2025-11-25 기준이라 차이 재확인 필요"라고 남아 있다 |

## 5. 현재 RC 트리 측정

```
$ grep -rn -i mcp src | wc -l            # tip cf4e662af
0
$ grep -rln -i mcp src tests schemas
tests/product_cli.rs
tests/fixtures/github/issues-closed.json
tests/fixtures/github/prs.json
tests/fixtures/github/issues-open.json
schemas/protocol/v1/variants/hold-00d-3-error-B.schema.json
```

Rust 제품 코드(`src/`)에 MCP 구현은 없다. 나머지 일치는 시험 fixture와 schema variant 파일에 나오는 문자열이다(이번에 내용은 열어 보지 않았다). HOLD-C1과 맞는 상태다.

## 6. 빈 부분 — 어느 정본·결정도 정하지 않은 것

| # | 빈 부분 | 이유 |
| --- | --- | --- |
| G1 | MCP를 rc.2(또는 2026.11.0)에 넣을지 | HOLD-C1 "다음 버전"이 어느 버전을 뜻하는지 적혀 있지 않다 |
| G2 | 노출 범위(어떤 operation을 tool로 내보낼지, 읽기 전용부터인지) | 정본 본문에 없다. 계승 페이지는 "기존 소비자 매핑"으로 범위를 정하는데, 그 방식은 사용자 지시와 충돌한다 |
| G3 | transport(stdio / Streamable HTTP)와 프로세스 수명 | 정본에 없다. 계승 페이지는 "remote 필수화 금지"만 적는다 |
| G4 | host 등록 방식(Claude 프로젝트 `.mcp.json`, Codex 설정의 MCP 항목)과 그 생성을 MoAI가 맡을지 | 정본에 없다. 00D-AC-6은 `.mcp.json` 변경 0을 요구한다. 07-page L74와 10-page L144는 등록을 별도 승인 범위로 둔다 |
| G5 | 결과 계약(REQ-00D-07의 byte-동등 parity를 그대로 쓸지, 새로 정할지) | P06 미결. 정본 §18 L730은 공통 schema를 먼저 고정하라고만 한다 |
| G6 | 승인·권한 경계를 실제로 어떻게 묶을지(tool annotation, host 승인, policy 검증) | 07-page L74와 10-page L144는 원칙만 있다. 정본 L660–662의 확장 보안 요구를 MoAI MCP에도 적용할지는 적혀 있지 않다 |
| G7 | 대상 MCP spec revision | 00-D 거울은 2026-07-28을 현행으로 적었지만 재확인이 필요하다고 남겼다. 이 문서는 spec을 다시 조회하지 않았다. 구현에 들어갈 때 `/latest`와 changelog로 재확인해야 한다 |
| G8 | parity 시험 방법(CLI와 같은 verdict) | 10-page L142와 HA-09가 golden fixture를 말하는데, 그 fixture의 원본이 기존 소비자다(G2와 같은 충돌) |
| G9 | Codex MCP 클라이언트가 실제로 지원하는 기능 | 정본에 없다. 참고로 통신 허브 조사 `docs/session-hub-research-2026-11-08.md:54`는 Codex의 prompts·sampling·roots·범용 서버 알림을 "미확인"으로 적었다(조사 기록이고 결정은 아니다) |

## 7. 사용자 결정 질문

각 질문은 서로 다른 출처가 다른 답을 가리키는 지점이다. 레인과 리더는 답을 정하지 않는다.

| # | 질문 | 부딪치는 출처 |
| --- | --- | --- |
| Q1 | 정본 L783의 "통합한다"와 "충돌하는 과거 선택은 역사로 남긴다" 가운데 무엇이 MCP 정의에 적용되는가? 즉 계승 페이지의 "MCP는 선택적 1급 adapter"(00-page L86, 10-page §7, 12-page L38)는 지금도 유효한 설계인가, 역사인가? | 정본 L783 ↔ 사용자 지시 ADK_MCP_CLEAN_REDESIGN(정본 기준 처음부터 재설계) |
| Q2 | HOLD-C1 "MCP는 다음 버전"에서 다음 버전은 rc.2인가, 2026.11.0 다음 릴리스인가? rc.2가 아니면 이 카드와 EP-I의 MCP 항목은 그 버전까지 blocked로 둘 것인가? | decisions L51 ↔ decisions L99(rc.2에 MCP 영역도 발행, blocked 표시) |
| Q3 | 기존 Go 시대 `mcp__moai__*` 소비자를 다루는 원칙은 무엇인가? 사용자 지시는 과거 구현을 근거로 쓰지 말라고 한다. 계승 페이지는 "소비자 존재를 무시한 MCP 제거·일괄 배포 제외는 허용하지 않는다"고 한다. (a) 소비자를 무시하고 새로 설계 (b) 소비자 목록만 요구 입력으로 쓰고 구현은 새로 (c) 다른 방식 | ADK_MCP_CLEAN_REDESIGN ↔ 10-page L131·L144, 00-page L86·L1341 |
| Q4 | `21-page_20fc…` "MoAI MCP와 기존 기능 재사용 설계"는 폐기인가, 참고 입력인가? (§20 계승 목록에 없고, 이 문서는 근거로 쓰지 않았다) | 정본 §20 목록 ↔ RC1 handoff 묶음에 포함 |
| Q5 | MCP를 진행한다면 결과 계약은 REQ-00D-07(CLI envelope `result`와 byte-동등, 같은 `outputSchema` 파일)인가, 새 계약인가? P06을 이 결정과 함께 풀 것인가? | REQ-00D-07 ↔ P06 미결(01-page L264) |
| Q6 | host 등록(`.mcp.json`, Codex MCP 항목)을 MoAI가 생성하는 일이 범위에 들어갈 수 있는가? 들어간다면 00D-AC-6(`.mcp.json` 변경 0)을 바꾸는 결정이 따로 필요한가? | 00D-AC-6 ↔ 카드 원래 범위(`.mcp.json` plan/apply, 리더 댓글로 철회) |
| Q7 | transport와 배포 형태: 로컬 stdio만인가, Streamable HTTP도인가? 어느 경우든 "서버 필수화 금지"(정본 L31·L738, 00-page L1335)를 어떤 기준으로 확인할 것인가? | 정본 L31·L738 ↔ 빈 부분 G3 |

## 8. 이 카드의 결과와 남은 일

- 코드·설정·`.mcp.json` 생성 0. 카드 본문의 AC-1(fixture plan/apply 보존 시험)과 AC-2(노출 op 목록 + schema 파일)는 리더 댓글 6732287215로 철회됐다. 이 문서는 그 자리에 근거 목록과 질문 목록을 남긴다.
- 상태: 설계 빈 부분(G1–G9)과 RC1 결정 충돌(decisions L51 HOLD-C1)이 겹쳐 있어 사용자 결정을 기다린다. 구현 착수 여부는 §7 답에 달려 있다.

## Gaps (미검증)

- 계승 페이지는 2026-11-08 조회 사본이다. 원문 Pages가 그 뒤 바뀌었는지는 확인하지 않았다.
- 정본은 Downloads 내보내기 사본이다. sha256은 기록값과 같다. 원 URL과 내보낸 시각은 crosswalk L66 기록에 의존한다.
- MCP spec(`modelcontextprotocol.io`)과 Claude Code·Codex의 MCP 문서는 이번에 다시 조회하지 않았다(G7, G9).
- `21-page_20fc…`는 머리말만 읽었다. 본문에 다른 결정이 있는지는 확인하지 않았다.
- host 실세션은 실행하지 않았다.

## Residual-risk

- grep 기반이라 "MCP"라는 글자 없이 같은 개념을 말하는 행(예: "구조화 호출", "typed tool")은 일부만 잡았을 수 있다. §2와 §3의 표는 `mcp` 일치 행과, 그 주변에서 직접 읽은 행에 한정된다.
- 계승 페이지의 행 번호는 스냅샷 파일 기준이다. 원문 Pages의 위치와 다를 수 있다.

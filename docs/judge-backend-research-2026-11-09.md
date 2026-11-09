# 판단 backend 조사 — Jev · OpenAI Decisions · Claude/Codex 구독 credit (2026-11-09)

Issue #2204 (rc.2 F6, 상위 #2158). 근거: 설계 §18 '남은 결정' 8항, 상세 설계 ⑤ `page_e3612b0b…` L147-151([e361] 수정안 1·2행), 사용자 결정 `docs/decisions-2026-11-08.md` 01:1x(11-09) "P10 판단 backend" 행.

## 0. 범위와 선언

- **비용 호출 0 (AC-2).** 이 조사는 공개 문서 페이지를 HTTP GET으로 읽기만 했다. 어떤 vendor API도 호출하지 않았고(Jev `/v1/systemone`, OpenAI `/v1/decisions`, Claude API 모두 0회), API key 발급·credit claim·계정 연결·결제를 하지 않았다. 숫자는 전부 페이지 문구이며 측정값이 아니다.
- **확인일.** 표의 "확인일"은 페이지를 읽은 날짜 2026-11-09(UTC, 02:46Z 전후)다. 페이지가 자체 날짜를 표시하면 따로 적었다. 대부분의 페이지는 "last updated"를 표시하지 않는다.
- **REQ-F6-01.** 과거 문구(상세 설계 ⑤ L147-151, 2025년 프로모션 문구, 블로그·aggregator 글)만으로 혜택을 확정하지 않는다. 공식 페이지에서 읽은 것만 "확인", 공식 페이지가 말하지 않으면 "미확인", 읽지 못한 페이지는 "읽지 못함(사유)"으로 적는다. 제3자 글은 출처로 쓰지 않았다.
- **현행(RC1)은 바뀌지 않는다.** 판단은 실행 중인 코딩 에이전트(`native`)가 하고, 별도 판단 backend는 P10 HOLD다(`src/decision.rs:5-6`, `docs/rust-product-path.md` Decision record 절). 이 문서는 backend 구현 카드의 선행 조사이며 코드·설정을 바꾸지 않는다.

## 1. backend별 현행 API·요금 (AC-1)

| backend | 현행 API | 요금(페이지 문구) | 상태 | 확인일 · URL |
|---|---|---|---|---|
| TypeSafe Jev (System One) | `POST https://api.typesafe.ai/v1/systemone`, `Authorization: Bearer <API_KEY>`(키는 TypeSafe console). `GET /v1/models`. primitive: Choice·Score·Noul·Box(Box는 Jev 1.14). 한도: 요청당 64k, 질문당 32k 토큰(`TypeSafe-Danger-Zone-Long-Context: true`로 262k). | 입력 $0.042/MTok($42/BTok), "Output tokens are free". 무료 tier·구독 credit: 미확인(페이지에 없음) | 홈페이지 "early access"; 모델 `jev-1.14.0`(2026-10-14부터), `jev-1.13.0`; `jev-latest` → `jev-1.14.0` | 2026-11-09 · https://docs.typesafe.ai/models , https://docs.typesafe.ai/introduction/quickstart , https://docs.typesafe.ai/concepts/system-one , https://typesafe.ai (페이지 표시 published 2026-11-06) |
| OpenAI Decisions API | `POST https://api.openai.com/v1/decisions`, 예시는 `Authorization: Bearer $OPENAI_API_KEY`. 질문 유형 predicate·choice·score, 요청당 질문 1–400, `refusal` 응답 가능. 사용량 `GET /v1/organization/usage/decisions`. ChatGPT 로그인 경로: 미확인(페이지에 없음) | uncached 입력 $0.10/MTok, "Output tokens are not charged"; cache 할인·regional/long-context 배수는 별도 pricing 페이지(숫자 미확인) | "generally available"(앞선 public beta 언급, 날짜 미표시); 지원 모델 `gpt-6-luna` 하나 | 2026-11-09 · https://developers.openai.com/api/docs/guides/decisions , https://developers.openai.com/api/reference/resources/decisions/methods/create · 읽지 못함: https://openai.com/index/decisions-api/ (HTTP 403) |
| Claude API (판단을 Claude 모델에 맡길 경우) | Claude API(Console 조직 API key) | 현행 모델 입력/출력 $/MTok 예: Opus 5.5 $4/$20, Sonnet 5.5 $2/$10, Haiku 5.5 $0.10/$0.50(100k 이하 prompt), Haiku 4.5 $1/$5; Batch 50% 할인 | 판단 전용 endpoint 없음(범용 Messages) | 2026-11-09 · https://platform.claude.com/docs/en/about-claude/pricing |
| native (현행) | 실행 중인 host 에이전트(Claude Code·Codex)가 판단, `decision-record`로 기록 | 별도 호출 없음(host 구독·사용량 안) | RC1 출하 | `src/decision.rs:1-6`, `docs/rust-product-path.md` |

## 2. 구독 연계 API credit (AC-1)

| 항목 | Anthropic (Claude 구독) | OpenAI (ChatGPT 구독) |
|---|---|---|
| 자격 | Max, Ultra, Team(할인 Team 포함). "Free, Pro, and Enterprise plans aren't eligible." iOS/Android 구매도 자격, claim은 claude.ai 웹 | Pro($100·$200·$500 tier), Business(유료 Standard·Premium seat). Plus·Go·Free는 "ineligible". 무료 체험·무상 구독 제외 |
| 월 금액 | Max $100, Ultra $200, Team Standard $20/seat·Premium $100/seat(pooled, 상한 $500) | Pro $100/$200/$500(tier별), Business Standard $20/seat·Premium $100/seat(workspace당 월 $500 상한) |
| claim | claude.ai Settings > Billing(Team은 Organization settings > Billing) → "API credits"에서 Console 조직 연결·약관 수락 → Console Billing "Promotional credits" | ChatGPT Settings → Billing(Business는 workspace billing/Admin Console) 또는 API Platform Billing에서 구독 선택 → "Confirm to activate the benefit"; 이후 갱신 때 자동 지급 |
| 역할 | 구독자(Team은 Primary Owner·Owner) + Console 조직 Owner·Admin·Billing; 플랜 1개 ↔ 조직 1개, 연결 조직 변경은 직접 불가 | 구독/workspace와 API 조직 둘 다 소유, 같은 email |
| 대기 | 적격 플랜 7일 뒤 | 구독 7일 이상; 첫 지급은 남은 주기 비례 |
| 만료 | 주기 말 만료, rollover 없음 | 주기 말 만료, rollover 없음 |
| 연간 결제 | "On annual plans, credits arrive monthly"(만료도 월 단위) | Business 연간 선불: 1년치 한 번 지급, 연간 주기 말 만료. Pro 연간: 미확인(페이지에 없음) |
| 적용 상품 | Claude API, Managed Agents, Agent SDK, Console Playground: Yes. Claude Code: No. Claude 앱 extra usage: No. AWS/Bedrock/Google Cloud/Foundry 경유: No | 연결 조직의 API Platform 사용(Responses, Agents, **Decisions API**, Realtime, Images, 내장 도구 등). "Codex usage through API keys is not supported by this benefit." ChatGPT·로그인 Codex 사용 한도를 늘리지 않음 |
| `claude -p`(API key) | 미확인(두 페이지 모두 언급 없음). Claude Code 자체는 No | 해당 없음 |
| auto-reload | "Auto-reload is based only on your purchased balance"; 포함 credit을 먼저 쓰고 구매 credit | 무료 credit을 먼저 쓰고 구매 credit; 소진 시 HTTP 429 `credit_balance_exhausted`; Auto Recharge와 무료 grant의 상호작용: 미확인 |
| 소진 시 | API 요청 중단, "Usage is never charged to your Claude plan" | 429(위) |
| 확인일 · URL | 2026-11-09 · https://platform.claude.com/docs/en/about-claude/api-credits-for-subscribers , https://support.claude.com/en/articles/8977456-how-do-i-pay-for-my-claude-api-usage (페이지 "Updated over a week ago") | 2026-11-09 · https://developers.openai.com/api/docs/guides/chatgpt-api-credits · 읽지 못함: https://help.openai.com/en/articles/20001626-api-credits-benefit-for-eligible-chatgpt-subscriptions (HTTP 403) |

TypeSafe(Jev): 구독 상품·credit 프로그램은 미확인(읽은 페이지에 없음).

Codex 과금 경로(같은 날 확인, https://developers.openai.com/codex/pricing → 308 → https://learn.chatgpt.com/docs/pricing): ChatGPT 로그인은 Free·Go·Plus·Pro·Business·Edu·Enterprise 플랜에 포함되고 ChatGPT와 사용량을 공유한다. API key는 "Pay for Codex usage based on API pricing", 클라우드 기능 없음. "API token prices are separate from subscription usage". 플랜별 사용 한도 숫자는 페이지 후반이라 미확인.

## 3. 상세 설계 ⑤ L147-151 가설 대조 (REQ-F6-01)

[e361] 수정안 1행이 "미확정 가설"로 돌려 둔 문장들을 위 공식 페이지와 맞춘다. 이 표는 2026-11-09 시점 문구 대조이며 vendor가 바꾸면 다시 읽는다.

| 설계 가설 | 대조 결과 | 근거 |
|---|---|---|
| OpenAI credit은 API-key Codex를 제외 | 일치(확인) | OpenAI 표 "적용 상품" |
| Claude는 직접 API-key `claude -p`·Agent SDK 예외 | Agent SDK는 적용 Yes(확인). `claude -p`는 미확인. Claude Code는 No | Anthropic 표 "적용 상품" |
| annual 지급과 monthly 지급이 다름 | Anthropic: 연간도 월 지급(확인). OpenAI: Business 연간 선불은 1년치 일괄(확인), Pro 연간은 미확인 | 두 표 "연간 결제" |
| auto-reload와 구매 잔액은 별개 | Anthropic: 구매 잔액만 기준(확인). OpenAI: 사용 순서만 확인, auto-reload 상호작용 미확인 | 두 표 "auto-reload" |
| 양사 자격·claim·계정 연결·유효기간·적용 상품이 다름 | 다름(확인): Pro는 Anthropic 부적격·OpenAI 적격, 연결 단위(Console 조직 vs API 조직), 역할 조건이 다르다 | 2절 전체 |
| "OpenAI Decisions"(endpoint·문서 버전 미상) | 현재 `POST /v1/decisions`, GA, `gpt-6-luna`로 식별(확인). 문서 버전 표시는 없음 | 1절 |

설계가 보존할 구분 축(auth_mode, launcher_origin, billing_product, organization/project, credit_eligibility)은 위 결과와 맞는다: 같은 vendor 안에서도 로그인 경로와 API key 경로의 과금·credit 적용이 다르다.

## 4. Jev — Go 기준선 사실

Go 저장소 `moai-adk-go` @ `4fa2aa9db2`(2026-11-07, 읽기 전용, grep/sed로 확인):

- `internal/jev/jev.go:1-3` 단일 client 구현("TypeSafe System One call path": 요청 구성·credential 읽기·HTTP·응답 해석·사용량 기록).
- `:73` `ModelID = "jev-latest"`(alias, 버전 아님), `:80` `VerifiedInterpretedModel = "jev-1.13.0"`(2026-10-15 측정), `:97` `EndpointURL = "https://api.typesafe.ai/v1/systemone"`(컴파일 상수).
- `:11-17` 불가용은 값(no signal)이지 오류가 아니다; `:19-22` gate가 패키지 안에 있어 비활성이면 요청을 만들지 않는다; `:24-36` display-only(파일·queue·git 변경 없음), 순위 예외는 `workflow.jev.enabled` 뒤.
- `:135-136` credential 없음 → `no-credential`; `:196-197` 입력 토큰만 과금된다는 전제로 질문을 한 요청에 묶는다.
- `internal/config/workflow_jev_test.go:13-19` `workflow.jev.enabled` 기본 false.

현행 페이지와의 차이(사실만): 공식 `jev-latest`는 지금 `jev-1.14.0`을 가리킨다(1절). Go 기준선의 검증 모델은 `jev-1.13.0`이므로 Go client는 응답 모델 id가 다르면 change notice를 한 번 낸다(`jev.go:76-79` 주석, `:765` `changeNotice`). Rust RC에는 jev 코드가 없다(`src/`에 0건, P10 HOLD).

## 5. 결론 — 구현 카드에 넘길 것

- 세 외부 backend 모두 **API key 과금 경로**다. 판단 backend를 켜면 외부 호출 비용이 생기므로, 사용자 결정 01:1x(11-09) "비용 상한 미설정 = 외부 backend 호출 0" 기본값과 함께 opt-in이어야 한다.
- 구독 credit으로 상쇄 가능한 경로: OpenAI Decisions는 Pro/Business credit 적용 상품에 명시돼 있다. Claude API 판단은 Max/Ultra/Team credit 적용(Claude API·Agent SDK). 단, 두 credit 모두 host 코딩 에이전트(Claude Code, API-key Codex)의 사용에는 쓰이지 않는다. Jev는 credit 프로그램 미확인.
- 구현 카드는 이 결과에 blocked로 발행한다(01:1x). 구현 전 다시 확인할 미확인 항목: `claude -p` API-key 사용의 credit 적용, OpenAI Pro 연간 지급, OpenAI auto-recharge와 무료 grant 관계, Decisions cache·long-context 단가, Jev 무료 tier·가격 보장, 읽지 못한 두 페이지(OpenAI 블로그·Help Center, HTTP 403).
- 정책은 외부 사실이므로 구현은 정책 값을 코드 상수로 굳히지 않고 확인일과 URL을 함께 기록한다(Go Jev가 모델 id를 alias와 응답 attribution으로 나눈 방식과 같은 이유).

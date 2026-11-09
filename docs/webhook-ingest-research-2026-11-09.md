# webhook 수신 조사 — 수신자 없는 설계와 GitHub webhook의 양립 조건 (#2174, REQ-B7-01)

상위: #2154 (EP-B). 설계 근거: `references/chatgpt-space/page_0d4538a27fb88191a19ba261bd5a3442.md` §6 L164(사본 sha256 `e24298b229e7618d6d06f1533d50d6a2dad45c1807bb31be5ad8d1487185a1bd`, 2026-11-09 읽음).

## 조사 방법

- 외부 사실은 GitHub 공식 문서(docs.github.com, 원본 저장소 `github/docs` main branch)에서 확인했습니다. 확인 경로는 Context7 문서 미러(`/github/docs`, 2026-11-09 질의)이며 각 사실에 공식 URL과 확인일을 붙였습니다(REQ-B7-01).
- 설계 전제는 위 사본 인용이며 외부 사실이 아닙니다.
- 확인 날짜: **2026-11-09**(아래 모든 외부 사실에 동일 적용, 별도 표기 없으면 미확인 아님).

## 설계 전제 (내부 근거)

설계 §6 L164: "Webhook은 이미 존재하는 수신 환경을 사용자가 선택한 경우만 추가한다. 서명 검증, delivery dedup, 권한·actor 확인, 역순·누락 처리를 갖추고 poll 복구를 유지한다. 전달된 텍스트나 event가 새로운 실행 권한을 만들지 않는다. GitHub 실패 webhook의 자동 재전송을 전제하지 않는다."

RC1 전제(동일 파일): 실행 중인 host만 활성 Issue·PR/MR·CI를 polling하고, 프로세스가 종료되면 poll도 중지하며 다음 실행에서 reconciliation한다. 항상 켜진 서비스가 없다는 점을 UI에 표시한다.

## 외부 사실 (URL + 확인일 2026-11-09)

| # | 사실 | URL |
| --- | --- | --- |
| F1 | GitHub은 실패한 webhook delivery를 **자동 재전송하지 않는다**. 재전송은 사용자가 수동으로(REST API·UI) 하거나, 스스로 스케줄 스크립트(REST API로 실패 delivery 조회 후 재전송)를 돌려야 한다 | https://docs.github.com/webhooks/using-webhooks/handling-failed-webhook-deliveries |
| F2 | 수신 서버는 delivery 수신 후 **10초**(GitHub.com 기준; GHES는 30초) 안에 2XX로 응답해야 한다. 넘으면 연결을 끊고 실패로 기록한다. 공식 권장: 즉시 응답하고 payload는 비동기 큐에서 처리 | https://docs.github.com/webhooks/using-webhooks/best-practices-for-using-webhooks |
| F3 | webhook **배달 순서는 보장되지 않는다** — 이벤트 발생 순서와 다르게 도달할 수 있으므로, 선후 관계가 필요하면 payload 안의 타임스탬프를 사용해야 한다 | https://docs.github.com/webhooks/testing-and-troubleshooting-webhooks/troubleshooting-webhooks |
| F4 | dedup 키: `X-GitHub-Delivery` 헤더의 GUID로 delivery를 식별하며, **같은 delivery의 재전송에서도 GUID가 불변**이다 — 재전송·중복 도달 판별에 이 GUID를 사용한다 | https://docs.github.com/webhooks/webhook-events-and-payloads , https://docs.github.com/webhooks/using-webhooks/automatically-redelivering-failed-deliveries-for-a-repository-webhook |
| F5 | 서명 검증: secret 설정 시 `X-Hub-Signature-256`(요청 본문의 HMAC-SHA256 hex digest, prefix `sha256=`)를 보내며, 수신자는 constant-time 비교로 검증해야 한다. SHA-1 `X-Hub-Signature`는 호환용으로만 제공되고 256 사용이 권장된다 | https://docs.github.com/webhooks/using-webhooks/validating-webhook-deliveries |
| F6 | delivery 헤더: `X-GitHub-Hook-ID`·`X-GitHub-Event`·`X-GitHub-Delivery`·`X-GitHub-Hook-Installation-Target-Type/ID`·`User-Agent: GitHub-Hookshot/*` | https://docs.github.com/webhooks/webhook-events-and-payloads |
| F7 | 실패 delivery의 REST 조회·재전송 엔드포인트가 repo/org/App webhook 단위로 존재한다( Marketplace·Sponsors·global webhook에는 조회 엔드포인트 없음) | https://docs.github.com/rest/repos/webhooks , https://docs.github.com/webhooks/using-webhooks/handling-failed-webhook-deliveries |
| F8 | 개발용 포워더(smee.io)로 로컬 수신을 시험할 수 있다 — 운영 수신 환경이 아니며 본 조사의 전제(사용자 기존 수신 환경)와는 별개 | https://docs.github.com/webhooks/using-webhooks/handling-webhook-deliveries |

## 미확인 항목

| 항목 | 상태 |
| --- | --- |
| GitHub 실패 delivery에 대한 재시도 상한·백오프 존재 여부 (자동 재전송이 "없음"은 확인했으나, 일부 통합 문서에서 말하는 내부 재시도 동작은 공식 문서만으로 단정 불가) | 미확인 — 수신자 구현 시 수동 재전송·poll 복구만 전제 |
| GHES 환경에서의 실측 (30초 타임아웃 등 조건부 문서는 확인했으나 GHES 인스턴스 실측은 없음) | 미확인 |
| 사용자 "기존 수신 환경"의 실제 형태 분포(역방향 프록시·relay 서비스·사내 서버 등) | 미확인 — 사용자 질의 대상 |

## service-free 전제와의 양립 조건

RC1 전제는 "항상 켜진 서비스가 없다"는 것이고 webhook은 **수신자가 필요한** 기술입니다. 양립은 다음 조건이 모두 참일 때만 성립합니다:

1. **수신 환경은 사용자 소유**: 사용자가 이미 운영하는 공개 도달 엔드포인트(기존 서버·역방향 프록시·relay)가 있어야 합니다. moai-cli가 리스너·터널·relay를 대신 띄우거나 상시 상주하는 것은 전제 위반 — 이 조건이 충족되지 않으면 webhook은 **채택하지 않고 polling만 유지**합니다(F1: 자동 재전송도 없으므로 수신 환경 부재 시 polling이 유일한 신뢰 경로).
2. **10초 안내-응답과 비동기 처리(F2)**: 수신 엔드포인트는 10초 안에 2XX로 응답하고 처리는 분리해야 합니다. moai-cli는 상주 프로세스가 아니므로, 수신 환경이 payload를 **디스크 등에 적어 두고** 다음 moai-cli 실행이 읽는 파일 기반 전달이 양립 가능한 최소 형태입니다(수신 환경이 파일을 남기는 것은 "사용자가 선택한 수신 환경"의 일부).
3. **dedup은 GUID 기반(F4)**: 재전송·중복 도달에서도 GUID가 불변이므로, 수신 환경이 남긴 delivery 기록의 GUID를 처리 이력 키로 삼습니다.
4. **역순 처리(F3)**: 도달 순서에 의미를 두지 않고 payload 타임스탬프와 moai-cli 실행 시 reconciliation으로 순서를 복구합니다. "역순·누락 처리를 갖춘다"는 설계 문구의 구현 형태.
5. **서명 검증 필수(F5)**: `X-Hub-Signature-256` HMAC-SHA256 검증과 constant-time 비교는 수신 환경 측 구현 요건입니다(moai-cli가 검증기를 제공하는 것은 가능 — 상주와 무관).
6. **자동 재전송 부재(F1) + poll 복구 유지(설계)**: 실패 delivery는 GitHub이 다시 보내지 않으므로, **poll이 여전히 정합성의 최종 권위**입니다. webhook은 도달 시점을 앞당기는 부채널이며 빠졌으면 poll이 뒤따라 잡습니다.
7. **권한 불생성(설계)**: webhook이 전달한 텍스트·event는 새로운 실행 권한을 만들지 않습니다 — 수신 내용은 05-B가 이미 승인한 plan의 실행 트리거 역할만 하고, 승인 스냅샷 없는 이벤트는 처리하지 않습니다.

## polling과의 역할 분담

| | polling (00-F/05-C 현행) | webhook (조건부 부채널) |
| --- | --- | --- |
| 역할 | 정합성의 최종 권위 — reconciliation으로 누락·역순·실패를 전부 복구 | 도달 지연을 줄이는 알림 트리거 |
| 요구 | 실행 중 host만 존재(기존 전제) | 사용자 기존 수신 환경 + 서명 검증 + GUID dedup + 비동기 기록 |
| 실패 시 | 다음 poll에서 재조정 | 자동 재전송 없음(F1) → poll이 커버 |
| 우선순위 | 필수 | 선택(수신 환경이 있을 때만) |

## 결론 (채택/보류)

- **채택**: "사용자가 기존 수신 환경을 이미 갖춘 경우"에 한정한 webhook 지원을 설계에 반영합니다. 요건은 위 양립 조건 1~7(서명 검증 F5, GUID dedup F4, 비동기 기록 F2, 역순 무시 F3, 권한 불생성, poll 복구 유지)이며, 이는 설계 §6 L164 문구의 조건과 1:1로 대응합니다.
- **보류**: (1) moai-cli 자체 수신기·터널·relay 제공 — service-free 전제 위반; (2) 자동 재전송 가정 — F1로 공식 문서가 부정; (3) webhook 단독 운용 — poll 복구 없는 webhook은 F1 때문에 누락 복구가 불가능; (4) 토큰 사용량 추정 — receipt에 원천이 없어 추정 대신 absent 표기(#2191 D5 보고와 정합).
- **구현 카드 전제**: 구현 착수 전 사용자의 수신 환경 존재 여부를 확인하는 질의가 선행합니다. 미확인 상태에서는 polling 단독이 기본이며, 이 문서의 조건표가 구현 카드의 입력입니다.

## not_run

- 실제 webhook 수신 환경에서의 E2E(사용자 환경 필요) — 미실행
- GHES 인스턴스 실측 — 미실행
- GitLab webhook (별도 설계 절) — 이 문서 범위 밖

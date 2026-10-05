# SPEC-MOAI-GPT-AUTH-001 — 수용 기준

**AC-GA-001** (REQ-GA-001) — Given 공식 broker와 비어 있는 MoAI 전용 scratch 및 미로그인 canonical 저장소가 있을 때, When 로그인을 완료할 때, Then 로그인 상태가 MoAI 저장소에만 반영되고 구독 방식이 표시되며, 성공한 인증으로 실제 GPT 요청이 수행된다. 실제 공급자 판정과 mock 판정을 별도 기록한다.

**AC-GA-002** (REQ-GA-002) — Given 공식 broker와 작업 ID를 기록하는 MoAI wrapper 및 별도 mock broker가 있을 때,
When 다른 작업 ID의 완료·중복 완료·취소 뒤 완료·실패 notification을 각각 주입하면, Then 잘못된 결과의 canonical publish는
0이고 현재 작업의 정상 완료만 1회다. 공식 broker의 PKCE state/verifier/S256·callback 검증은 공급자 소관으로 기록한다.
실제 새 로그인 preflight에서 관측 가능한 state·callback 거절을 별도 확인하며 MoAI mock 성공을 broker 내부 PKCE 검증으로
세지 않는다. 관측할 수 없는 내부 검증은 미검증 Gap으로 명시한다.

**AC-GA-003** (REQ-GA-003) — Given 유효 기존 로그인과 새 로그인 대기가 있을 때, When 취소와 제한 만료를 각각 발생시키고 뒤늦은 callback을 보낼 때, Then listener와 대기 프로세스가 종료되고 기존 credential 세대·해시는 그대로이며 추가 token 교환은 0이다.

**AC-GA-004** (REQ-GA-004) — Given 격리 저장소와 쓰기 실패·손상 파일·과도한 권한 픽스처가 있을 때, When 로그인 저장과 상태 읽기를 실행할 때, Then 정상 저장은 원자적이며 소유자 외 접근이 제한된다. 실패 시 성공 표시·부분 credential 노출이 없고 로그·argv·오류에 공급한 비밀 표지가 없다.

**AC-GA-005** (REQ-GA-005) — Given 동일 만료 credential을 공유하는 별도 OS 프로세스 두 개와 지연 가능한 token mock이 있을 때, When 동시에 갱신을 요청할 때, Then 동일 세대의 refresh는 한 번만 교환되고 둘은 최신 세대를 읽는다. 교환 중 새 로그인 또는 로그아웃이 완료되면 지연 결과는 해당 새 상태를 덮어쓰지 못한다.

**AC-GA-006** (REQ-GA-006) — Given credential 해석 뒤 송신 직전에서 멈춘 요청과 logout 작업이 있을 때, When 로그아웃을 완료한 뒤 요청을 풀 때, Then 옛 credential을 가진 외부 송신은 0이다. 송신을 먼저 허용한 대조군은 별도 in-flight로 기록되어 취소 결과가 표시된다. generation 검사만 하고 무조건 보내는 뮤턴트는 실패한다.

**AC-GA-007** (REQ-GA-007) — Given provider·endpoint·세대와 credential을 다르게 조합한 요청들이 있을 때, When 각 요청의 credential 적용과 송신을 실행할 때, Then 허용 조합만 해당 endpoint에 송신되며 부재·잘못된 provider·폐기 세대·redirect·실패한 refresh는 비밀 전달과 fallback 모두 0이다.

**AC-GA-008** (REQ-GA-008) — Given 구독 credential과 API 키를 각각 가진 격리 설정이 있을 때, When 각 인증 경로로 요청하고 상대 endpoint 및 임의 endpoint도 지정할 때, Then 선택한 경로만 성공하고 다른 경로는 송신 전에 거절된다. 사용자에게 기록되는 방식이 실제 수신 mock과 일치한다.

**AC-GA-009** (REQ-GA-009) — Given CODEX_HOME·Claude 전역 설정에 표지 파일을 둔 격리 환경이 있을 때, When 로그인·상태 조회·refresh·logout을 수행할 때, Then 표지 경로의 내용 해시·존재·권한이 그대로이고 credential 자동 읽기·복사 시도 계수도 0이다.

**AC-GA-010** (REQ-GA-010) — Given 실제 GPT 구독 인증과 코어·PICKER 통합 세션이 있을 때, When 같은 Claude Code에서 네 GPT ID를 차례로 선택하고 각 모델에 이전 대화 회상과 도구 왕복을 요청할 때, Then 각 응답의 실제 모델 ID·대화 맥락·tool 결과·stream 종료가 검증된다. 401·429·권한 부재는 미완료로 기록하며 다른 인증 방식으로 통과시키지 않는다.


## 기존 AC의 broker 트랜잭션 보강

- AC-GA-003·004: partial/truncated scratch 파일과 broker 강제 종료를 재현한다. 종료 확인 전 파일을 읽어 publish하는
  뮤턴트, canonical을 broker에 직접 쓰게 하는 뮤턴트는 실패한다. 취소 시 기존 세대·내용이 같고 잔여 scratch는
  canonical로 채택되지 않는다. 사용자 keyring fallback도 0이다.
- AC-GA-005: 서로 다른 프로세스의 refresh 중 logout을 완료시킨 뒤 늦은 정상 결과를 보낸다. 결과는 새 tombstone을
  덮어쓰지 못한다. RPC success이지만 파일이 그대로이거나 오래된 token·last_refresh만 바뀐 mock은 refresh PASS가 아니다.
  실제 갱신의 token/만료·공급자 수용 근거를 확인하며 비밀값은 출력하지 않는다.
- AC-GA-006·007: first-write 직전·중간·완료 뒤 barrier를 각각 두어 logout과 직렬화한다. tombstone 뒤 옛 token
  header 송신은 0이다. header write만 끝나고 response SSE가 끝나지 않는 대조군에서도 logout은 response 종료를 기다리지
  않는다. DNS/connect/write timeout과 transport 미종료는 성공으로 표시하지 않는다. 원격 revoke 실패에도 canonical은
  tombstone이며 원격 실패가 따로 표시된다. 단순 Generation→Apply 호출만 하는 뮤턴트는 실패한다.
- AC-GA-008: 구독 endpoint와 API endpoint의 실제 수신 경로·필수 헤더·지원 필드 preflight를 분리 기록한다.
  app-server 추론 성공은 gateway의 Responses HTTP 성공을 대체하지 않는다.
- AC-GA-009: 초기 scratch seed는 빈 상태, refresh seed는 MoAI canonical만이다. 기존 사용자 CODEX_HOME·Claude 파일과
  keyring의 자동 읽기/변경·credential 복사 시도는 0이다. 환경 override·symlink 경로 이탈은 broker 시작 전 거절한다.

## 공통 실계정 시험 조건

Claude 실서비스 시험은 사용자 지시에 따라 **2026-09-11 19:00 Asia/Seoul 이후** 실행한다. Claude 시험 입력은
`claude-opus-5`와 `claude-sonnet-5`이며 계정의 실제 사용 가능 여부는 별도 관측한다. Sonnet 4.5 과거 캡처는 대체 근거가 아니다.
GPT 시험의 upstream ID는 `gpt-6-astra`, `gpt-5.6-sol`, `gpt-5.6-terra`, `gpt-5.6-luna`다. 임의 `gpt-6` 별칭을 만들지 않는다.
공식 `gpt-5.6` 별칭을 UI에 추가하더라도 공개 근거와 canonical 매핑 시험이 먼저 필요하며 본 초안의 필수 범위에는 넣지 않는다.
429는 실패한 인증 방식의 증명도 성공도 아니다. Claude OAuth의 코어 T09가 INCONCLUSIVE이면 그 상태를 유지한다.

## 근거와 완료 판단

사용자 범위·공식 모델 문서 위치는 `../../reports/SPEC-MOAI-GATEWAY-001/goal-execution-20260911.md`, 코어 경계는
`../SPEC-MOAI-GATEWAY-001/spec.md` 및 `design.md`가 기준이다. 코어의 기존 lint·mock PASS는 이 SPEC의 완료 근거가 아니다.
각 AC별 실행 명령·그 출력·HEAD·사용한 모델·mock/실계정 구분·Gap·잔여 위험을 보고한다. 관련 구현·통합 시험·독립 감사와
문서 동기화 전 완료로 표시하지 않는다. 이번 계획 작성은 numeric card 생성이나 기존 worktree의 새 카드 착수가 아니다.

## 0.2.0 관측의 귀속

AC-GA-001·008·010의 직접 구독 사전 관측은 `../../reports/SPEC-MOAI-GATEWAY-001/auth-responses-runtime-observation.md`에 있다.
네 GPT 함수 후속의 직접 200은 실제 Claude Code 통합·회상·도구 실행을 대신하지 않는다. SSE 형식 대조군은 코어
AC-MG-007·008과 함께 판정하며 malformed·빈 출력·오류를 성공으로 치환하지 않는다. 출력 상한은 승인된 구독 서버 정책과 API 키 매핑을 구분한다. AC-GA-004·005의 Windows 실행은 GitHub runner의 해당 HEAD·run/pass 이벤트·artifact로 판정하고
cross-compile·합성 checker를 native PASS로 세지 않는다. 저장 완료는 승인된 plan B/E의 Windows API 기준으로 판정하며 전원 장애 보장으로 확대하지 않는다.

## 승인된 Windows API 저장 기준 — 기존 AC-GA-004·005 보강

Given Windows GitHub runner에 소유자 전용 protected 루트와 안전한 inherited broker 파일 대조군이 있을 때,
When 로그인·저장·읽기·refresh·logout을 실행하면, Then file flush·same-volume write-through 교체·최종 파일/부모
identity·ACL·내용 readback을 모두 통과한 상태만 완료로 표시한다. broad/NULL ACL·잘못된 소유자·reparse·부모 교체·
잘못된 파일형식·write/flush/replace/readback 실패는 성공이나 부분 credential 노출 없이 명시 오류다.
교체 전 실패는 이전 canonical을 보존하고, 교체 뒤 실패는 상태 변경 가능성을 표시하며 확인되지 않은 credential 송신은 0이다.

Given 별도 프로세스 잠금과 각 저장 경계에서 강제 종료 가능한 시험이 있을 때, When 경합·취소·강제 종료 뒤 다시 열면,
Then 잠금은 회수되고 상태는 완전한 이전/새 상태 또는 명시 오류이며 partial state를 채택하지 않는다. 지연 refresh는
최신 로그인/tombstone을 덮어쓰지 못한다. 이는 프로세스 crash 복구 시험이며 전원 장애 뒤 디렉터리 엔트리 보존의 증거가 아니다.
필수 test의 run/pass와 해당 HEAD의 run ID/artifact를 확인하고 missing/skip을 PASS로 세지 않는다. 기존 필수 시험을 삭제하지 않는다.

AC-GA-008: Given 같은 양의 정수 max_tokens 입력과 구독/API 키 경로가 있을 때, When 요청하면, Then 고정 구독 경로만
max_output_tokens를 생략하고 API 키 경로는 값을 보존한다. 잘못된 입력은 두 경로 송신 0이며 구독 서버 출력 정책과
응답 byte/취소 한도를 구분하여 기록한다. 다른 endpoint·인증 fallback·입력 잘림을 허용하는 변경은 아니다.

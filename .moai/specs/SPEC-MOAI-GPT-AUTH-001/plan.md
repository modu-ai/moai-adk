# SPEC-MOAI-GPT-AUTH-001 — 구현 계획

## A. 선택한 경로와 신뢰 경계

Tier M, 0.1.0 초안 보강이다. 이미 설치된 공식 Codex `app-server`를 인증 브로커로 재사용한다. 등록된 OAuth client를
MoAI가 복제 구현하지 않는다. 이전 custom client 등록 질문은 선택적인 별도 경로이며 브로커 계획의 인위적 blocker로
유지하지 않는다. 브로커가 하는 일은 새 ChatGPT 로그인·취소·갱신·로그아웃이며 추론은 코어 gateway의 Responses HTTP다.
app-server의 turn 실행으로 Claude Code 내 GPT 추론 구현을 대신하지 않는다.

브로커는 매 작업의 새 MoAI 소유 scratch `CODEX_HOME`에서만 실행한다. `cli_auth_credentials_store="file"`을 명시하여
해당 scratch의 `auth.json`만 사용한다. 초기 로그인 scratch는 비어 있고, refresh scratch의 seed는 MoAI가 앞서 명시
로그인으로 취득한 canonical credential만 허용한다. 사용자 CODEX_HOME·Claude 설정·keyring에서 기존 인증을 가져오지
않는다. 경로 소유권·권한·symlink 이탈과 환경변수 override를 실행 전에 검사한다. 관리 정책이 file 저장을 금지하면
기존 사용자 저장소로 fallback하지 않고 명시 실패한다.

브로커 RPC 후보는 `account/login/start`의 chatgpt 방식, login cancel, logout,
`account/read(refreshToken:true)`다. `account/read`는 토큰 export 응답이 아니다. token 값은 **종료가 확인된**
브로커의 private scratch 파일에서만 읽는다. 설치본 0.154.0의 버전·help 확인은 부모 관측이며 실제 RPC 호환 검증은
아직 없다. 공개 소스 HEAD `5a9eb14`는 그 설치본과 같은 빌드라는 근거가 없다.

## B. MoAI 저장 트랜잭션

1. canonical 저장소는 MoAI 소유 단일 원자 상태다. credential·세대·tombstone을 함께 교체한다. POSIX는 성공 write와
   fsync·rename·디렉터리 내구성 확인 전 저장 성공을 표시하지 않는다. Windows는 사용자 “Windows API 기준으로 진행”
   승인에 따라 file flush → 같은 volume native write-through 교체 → 최종 ACL/소유자·파일/부모 identity·내용 readback을
   모두 확인한 뒤 저장 완료를 표시한다. 이는 POSIX directory fsync와 동등한 전원 장애 내구성의 보장이 아니다.
   현재 사용자 전용 권한·실패 단계 구분·scratch cleanup 및 GitHub native 프로세스 crash 복구 시험이 필요하다.
2. 별도 프로세스 간 operation lock이 로그인·refresh 브로커를 직렬화한다. 짧은 state lock 아래 현재 세대와 상태를 읽고
   private scratch를 준비한 뒤 state lock은 놓는다. 네트워크가 끝날 때까지 state lock을 붙잡지 않는다.
3. 브로커 완료·종료를 확인한 뒤 파일의 형식·인증 방식·계정 일관성·필수 token 필드·유효기간을 검증한다. refresh는
   실제 갱신 결과의 새 token/만료 정보와 공급자 수용 근거로 확인하며 RPC success나 last_refresh 시각만으로 성공 처리하지
   않는다. 갱신 필요가 없는 읽기는 refresh 성공으로 기록하지 않는다. 확인 불가·401·429·부분 파일은 publish하지 않는다.
4. state lock 아래 시작 세대와 현재 세대가 같고 tombstone이 없을 때만 새 credential과 증가 세대를 원자 publish한다.
   취소·실패는 기존 상태를 보존한다. 로그인은 자신이 시작한 상태가 미로그인 tombstone이더라도 시작 세대가 같으면 새
   로그인으로 publish할 수 있다. 중간 logout/다른 login이 만든 세대 변화는 모든 지연 결과를 폐기한다.
5. logout은 operation lock의 장기 broker 작업을 기다리지 않고 state lock 아래 tombstone·새 세대를 먼저 publish한다.
   알려진 송신 lease의 취소를 요청한다. remote logout/revocation은 MoAI 소유 snapshot만 가진 별도 scratch broker로
   시도한다. 로컬 폐기 완료와 remote revoke 성공·실패·미지원은 분리 표시하고 remote 실패로 token을 복원하지 않는다.
   logout RPC가 서버 측 revoke까지 보장하는지는 runtime/source preflight에서 확인하기 전 주장하지 않는다.
6. scratch broker는 항상 종료를 확인하고 private 파일을 정리한다. 강제 종료·부분 write·cleanup 실패는 관측하고
   canonical 파일을 브로커가 직접 열지 못하게 한다. crash 뒤 남은 scratch를 유효 canonical 인증으로 채택하지 않는다.

기존 primitive를 우선 재사용한다. POSIX canonical의 최종 교체는 `internal/atomicfile.Replace`로 한다.
Windows AUTH canonical 교체에만 `MoveFileEx(REPLACE_EXISTING|WRITE_THROUGH)`의 같은 디렉터리·같은 volume 경로를
허용한다. COPY_ALLOWED·cross-volume fallback을 사용하지 않는다. 현재 atomicfile API의 다른 호출자와 의미는 바꾸지 않는다.
AUTH writer가 임시 파일 생성·소유자 전용 권한·file flush·부모 identity·최종 readback·실패 정리를 맡는다.
교체 전 실패는 이전 canonical을 보존하며, 교체 뒤 확인 실패는 무조건 rollback됐다고 보고하지 않는다.
미확인 상태를 credential 송신에 사용하지 않고 명시 오류로 끝낸다. `internal/lockfile`은 Windows에서 프로세스 내부 mutex라는 문서상 범위가 있어 GA-005의
프로세스 간 보장을 충족하는 근거로 사용할 수 없다. AUTH 전용의 최소 cross-platform lock을 두되 Unix native flock과
`internal/homestate/admission_lock_windows.go`의 native LockFileEx 패턴을 재사용한다. context/timeout·잠금 해제·crash
복구를 시험하고 관련 없는 전역 lockfile API를 개편하지 않는다. 기존 primitive의 범위를 넘어선 일반 결함을 주장하지 않는다.

## C. 송신 직전 세대와 logout 경합

코어 `CredentialRef`의 Provider/Generation/Apply/Redacted 네 메서드는 그대로 연결하지만 Generation 호출만으로
REQ-GA-006을 증명하지 못한다. run 단계는 다음 AUTH 전용 송신 경계를 코어 transport와 연결해야 한다. 이 계획은
코어 인터페이스 파일을 수정한 것이 아니며 연결이 없으면 AUTH 완료를 판정하지 않는다.

`SendAuthorized(ctx, expectedGeneration, request, transport)`는 짧은 프로세스 간 state lock 아래 tombstone·세대·
provider·endpoint를 다시 검사하고 token 적용과 **최초 HTTP request header write의 완료 또는 실패**까지 lock을 유지한다.
transport는 first-write 완료를 동기적으로 알리고 lock을 그 시점에 해제한다. 그 뒤 response stream 전체를 기다리며 lock을
유지하지 않는다. DNS/connect/TLS/request write에는 명시된 양수 deadline을 적용한다. 단순 goroutine enqueue나
RoundTrip 호출 시작을 송신 완료로 오인하지 않는다. write가 끝나거나 실패하기 전 lock을 놓는 구현은 금지한다.

logout은 이 경계와 직렬화된 tombstone publish가 끝나야 로컬 완료다. 이미 header가 전송된 요청은 in-flight이며 취소를
시도하되 소급 미송신으로 기록하지 않는다. deadline 후에도 transport 종료를 확인하지 못하면 logout은 완료를 거짓 표시하지
않고 pending/error를 반환한다. 다른 프로세스의 lease와 crash 복구도 검증한다. 응답이 끝나지 않는 SSE가 logout을
무한히 붙잡는 설계는 허용하지 않는다. 정확한 deadline·플랫폼 잠금 수단·취소 전달은 run에서 시험 전에 고정한다.

## D. endpoint 및 브로커 사전 검증

구독은 `https://chatgpt.com/backend-api/codex/responses`, OpenAI API 키는 `https://api.openai.com/v1/responses`로
구분하는 공식 근거가 있다. 출처는 [OpenAI agent loop 설명](https://openai.com/index/unrolling-the-codex-agent-loop/)이다.
임의 사용자 endpoint나 redirect에 credential을 싣지 않는다. 구독 endpoint의 필수 헤더·account 식별·지원 request 필드,
저장/stream 제약은 소스 기준선과 실제 동작을 대조한 뒤 allowlist를 고정한다. 공개 API schema를 그대로 지원한다고 가정하지
않는다. 이 preflight와 실제 새 로그인은 2026-09-11 19:00 Asia/Seoul 이후 실시한다.

[Codex 인증 문서](https://learn.chatgpt.com/docs/auth)는 file 방식이 CODEX_HOME/auth.json을 쓴다고 설명한다.
공개 소스의 `login/src/auth/storage.rs:206` 이후는 truncate/write 방식이며, `login/src/auth/manager.rs`의 semaphore는
프로세스 내부 동기화다. `app-server/src/request_processors/account_processor.rs:1017` 이후 refresh 오류를 구분하지만
응답 경로에서 결과를 무시하는 호출이 있어 RPC 응답만으로 성공을 판정하지 않는다. 기준은 공개 소스 `5a9eb14`이며 실제
설치본과의 일치는 미검증이다. app-server 프로토콜·파일 저장·취소·성공 notification은 설치본을 실행해 별도로 확인한다.

## E. 구현·검증 순서

1. High — 독립 계획 감사 후 broker 버전/프로토콜·scratch 격리·endpoint 사전 측정으로 경계를 고정한다.
2. High — mocked broker로 원자 publish·부분 write·실패 refresh·cancel·late completion·logout 세대 시험을 먼저 만든다.
3. High — 별도 OS 프로세스의 operation/state lock과 first-write/로그아웃 경합·timeout·crash 시험을 통과시킨다.
4. High — 공식 broker 새 로그인과 갱신을 검증하고 코어 CredentialRef/SendAuthorized 및 사용자 명령을 연결한다.
5. High — 코어·PICKER와 네 GPT 모델의 실제 Claude Code 도구 왕복·stream을 검증한다.
6. Medium — 변경 범위 시험·Windows 권한/잠금 실행 증거·독립 감사·문서 동기화로 닫는다.

PKCE 생성·교환은 공식 broker의 소관이다. MoAI mock 시험은 broker 작업 ID·notification·취소·파일 publish 경계만
검증하며 Codex 내부 PKCE unit coverage로 주장하지 않는다. 실제 callback 음성 시험이 불가능한 부분은 정확히 Gap으로
기록하고 공급자 내부 구현을 시험했다고 말하지 않는다. 숫자 카드 생성·구현 착수·로그인 성공은 이 문서 작성의 결과가 아니다.

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

## E. 19시 이후 사전 관측과 보존할 결정 경계 (0.2.0)

`../../reports/SPEC-MOAI-GATEWAY-001/auth-responses-runtime-observation.md`는 부모의 fresh MoAI 로그인 뒤
Store.SendAuthorized를 통한 네 GPT 직접 text·함수 후속 HTTP 200과 Luna no-tool opaque 보존 후속을 기록한다.
Content-Type 부재/빈 completed.output의 엄격한 형식 처리는 코어 design §4.3 소관이다. Claude Code 통합 AC-GA-010은
이 direct probe로 PASS하지 않는다. 기존 사용자 credential을 복사한 근거도 아니다.

사용자 결정(2026-09-11): “구독 서버 출력 정책으로 진행”. 고정 구독 AuthPKCE 경로는 outgoing max_output_tokens를
생략하고 입력 max_tokens의 양의 정수 검증을 유지한다. API 키 경로의 매핑은 그대로다. 코어 design §4.3과
AC-MG-007·008·010의 경로별 대조군을 따른다. 별도 API 과금 fallback은 금지한다.

사용자 결정(2026-09-11): “Windows API 기준으로 진행”. 위 B.1의 플랫폼별 저장 완료와 Windows 전용 native 교체
예외를 채택한다. `../../reports/SPEC-MOAI-GATEWAY-001/windows-store-contract-gap.md`의 선행 차이를 이 계약으로
정리하며 실제 native PASS는 GitHub runner의 HEAD·run ID·필수 test stream·artifact로 판정한다.

MoAI가 만드는 루트·scratch는 current-user SID 소유의 protected DACL로 생성한다. broker 자식의 inherited flag만으로
수용/거절하지 않고 검증된 보호 부모·실제 handle의 owner·effective DACL·reparse/파일형식·부모 identity를 확인한다.
현재 사용자 외 읽기·쓰기·삭제·권한 변경 허용, NULL/부재 DACL·잘못된 소유자·경로 교체는 거절한다. 이미 공개됐을 수
있는 broker 파일을 사후 chmod/ACL 수정으로 안전한 것으로 채택하지 않는다. native 관리자의 소유권 강탈까지
막는다는 보장은 아니다. protected/inherited 정상·음성군은 Windows CI에서 직접 실행한다.

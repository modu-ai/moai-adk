# App Server 전환 수용 기준 제안

t649 하위 설계 검토용 초안. canonical AC 추가나 완료 판정이 아니다. 각각 Given/When/Then으로 독립 판정한다.
검증은 설치 Codex 버전·schema hash·MoAI HEAD와 작업 diff·auth 모드·실행 명령·비밀값 없는 원출력을 함께 기록한다.

## AS-1 기반과 인증

**AC-AS-001** Given 설치 Codex의 schema와 기능 협상이 있을 때, When App Server를 시작하고 initialize·thread/start를
수행하면, Then 동적 도구 capability를 확인하고 unsupported 버전은 명시 오류로 끝난다. 자동 설치나 direct backend 우회는 없다.

**AC-AS-002** Given 구독 managed 계정과 API 키 전용 프로필을 각각 준비했을 때, When 로그인·상태·생성·갱신·로그아웃을
수행하면, Then 선택한 auth 모드로만 처리된다. 구독에서 MoAI의 토큰 파일 읽기·refresh·direct backend 요청은 0이고,
구독 만료/한도 오류가 API 과금 경로로 넘어가지 않는다. 같은 프로필을 경쟁 로그인으로 덮어쓰지 않는다.

**AC-AS-003** Given Codex native 도구가 가능한 기본 환경과 MoAI 제한 환경이 있을 때, When shell·file·MCP·agent 실행을
각각 유도하면, Then 제한 환경의 native 실행 계수는 0이고 실제 Claude 도구만 호출된다. 모델 도구 inventory도 대조한다.
approval never·readonly 또는 단순 설정 파일 존재로 이 AC를 통과시키지 않는다.

## AS-2 도구 실행권과 ToolSearch

**AC-AS-004** Given 실제 Claude 초기 요청의 일반·deferred 도구 N개와 각각의 schema가 있을 때, When A 방식으로
thread/start에 등록하면, Then 누락·이름 충돌·schema 변형 없이 N개가 대응된다. 최초 요청에 전체 schema가 없으면 A는
Gap/FAIL로 기록하고 B와의 비교를 보고한다. B는 native schema 수 N→1의 변경을 명시한다.

**AC-AS-005** Given App Server dynamic call이 발생했을 때, When Claude가 tool_use를 받아 승인·실행하고 다음 HTTP로
tool_result를 보내면, Then App Server의 같은 thread/turn/call에 결과가 한 번만 전달되고 최종 답변이 Claude 화면에 보인다.
텍스트·로컬 이미지·도구 실패를 각각 판정하며 unsupported 결과는 명시 오류다.

**AC-AS-006** Given ToolSearch가 `tool_reference`와 설명을 반환하는 실제 Claude 요청이 있을 때, When 검색 결과를
처리하고 발견된 도구를 호출하면, Then 400 없이 타입·이름·등록 schema가 검증되고 도구가 Claude에서 정확히 한 번 실행된다.
잘못된 tool_reference·미등록 이름·중복 발견·schema 변경·다른 대화의 참조는 송신 전 거절한다.
기존 textContent/receipt projection 거절 보고는 회귀 픽스처의 유래이며, 최신 partial fix 상태는 실행으로 재확인한다.

## AS-3 수명·권한·복구

**AC-AS-007** Given 도구 대기 때문에 첫 HTTP 응답이 끝난 turn이 있을 때, When 다음 HTTP 결과 요청이 오면,
Then App Server process와 pending RPC가 살아 있고 원 turn을 잇는다. 동시 대화 A/B의 결과를 바꾸면 둘 다 서로의 call을
소비하지 않는다. 같은 결과 재전송은 도구 재실행 없이 일관된 처리 또는 명시 중복 오류다.

**AC-AS-008** Given pending call의 실행 전·실행 후·결과 저장 후 각 지점에서 프로세스를 강제 종료했을 때,
When 재개하면, Then 실행 여부가 불명확한 도구를 자동 재실행하지 않는다. 복구 불가능한 active RPC ID는 명시 실패와
복구 안내를 제공한다. 성공 상태를 합성하거나 새 process에 옛 RPC 응답을 주입하지 않는다.

**AC-AS-009** Given 활동 중 turn 또는 병렬 tool call이 있을 때, When 사용자가 취소하거나 연결이 끊기면,
Then turn/interrupt와 대기 호출 정리가 해당 turn에만 적용되고 뒤늦은 결과가 다른 turn을 진행시키지 않는다.
EOF·RPC 오류·출력 상한 초과 때 성공 Messages terminal이 없어야 한다.

## AS-4 이력·모델·압축·fork

**AC-AS-010** Given GPT 대화를 종료하고 새 MoAI process로 시작했을 때, When 소유 thread ID로 resume하면,
Then 합성 사실·도구 결과가 유지된 정상 답변이 나온다. 다른 계정·다른 대화·변조 mapping은 거절한다.
thread/resume.history 또는 raw reasoning을 직접 가져와 재구축한 결과로 통과할 수 없다.

**AC-AS-011** Given 같은 GPT thread가 idle이고 허용 모델이 있을 때, When `/model`로 GPT-6 Astra와 5.6을 각각
선택하면, Then 실제 turn model이 선택 ID와 일치하며 App Server 경고·실패가 명시된다. 다른 제공자와 bare gpt-6는
거절된다. active tool 대기 중 전환은 잘못된 turn에 적용되지 않는다. family 호환 성공은 실제 계정별 별도 증거가 필요하다.

**AC-AS-012** Given 실제 Claude `/compact` 또는 자동 압축의 캡처된 표면이 있을 때, When 압축을 수행하면,
Then Codex contextCompaction 완료와 Claude 표시 이력이 정합하고 중복 입력·이중 압축·reasoning 유실이 없다.
단순 compact ack를 완료로 세지 않는다. 압축 표면을 식별할 수 없으면 자동 지원은 Gap으로 남긴다.

**AC-AS-013** Given 실제 Claude Agent가 병렬 자식 둘을 만들었을 때, When 각 자식 요청과 fork/resume를 처리하면,
Then 부모 및 자식 thread 소유권·결과·취소가 섞이지 않는다. agent 식별자를 얻지 못하면 텍스트나 모델 ID로 추정하여
같은 thread에 넣지 않는다. 각 자식의 tool 실행도 Claude가 담당한다.

## AS-5 제품 검증

**AC-AS-014** Given 세 MoAI launcher와 오염된 전역 모델 기본값이 있을 때, When 실제 PTY `/model`의 Default·현재 행·
목록·직접 입력·Enter·s·재개를 수행하면, Then cc는 Claude, glm은 GLM, gpt는 GPT만 선택·요청하며 공유 설정이 변하지 않는다.
구독과 API 인증 표시는 실제 선택과 일치해야 한다. Codex TUI를 열어 성공한 것은 이 AC의 증거가 아니다.

**AC-AS-015** Given 설치 App Server model metadata 및 토큰 사용량이 있을 때, When 한도 경계 안팎의 실제 입력을
보내면, Then 현재 경로 한도와 오류를 정확히 표시하고 direct endpoint 921k나 미검증 1M을 수용 보장으로 쓰지 않는다.
872k metadata도 실제 수용 시험과 별개로 표기한다. 구독 출력은 서버 정책이며 local 취소·출력 바이트 한도는 동작한다.

**AC-AS-016** Given 기존 review RPC 소비자와 Windows GitHub CI가 있을 때, When 공통 RPC 추출의 회귀 및
Windows process 종료·재개·파일 flush/원자교체·권한 시험을 실행하면, Then 기존 review 의미와 승인된 Windows API 계약이 유지된다.
Windows cross compile만으로 runtime AC를 통과시키지 않는다.

## 완료 판정

AS-1→AS-2→AS-3→AS-4→AS-5 순서의 증거를 모으고 독립 검토를 통과해야 한다. 하나의 실제 도구 왕복이나 로그인
성공으로 전체를 완료하지 않는다. mock·단위·프로토콜·제품 PTY·실제 구독/API·Windows 결과는 각각 명시한다.
계정·CI·실험 API 제약으로 수행하지 못한 항목은 Gap으로 남기며 기능을 완료로 광고하지 않는다.

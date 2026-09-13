# SPEC-MOAI-GATEWAY-TEAMMATE-001 — 수용 기준

**AC-GT-001** (REQ-GT-001) — Given 같은 tmux 세션에 lead A/B와 비gateway pane C가 있을 때, When A/B에서 동시에 teammate를 생성하고 각 요청을 보낼 때, Then A pane은 A gateway만 B pane은 B gateway만 사용하고 C의 환경·설정 해시는 그대로다. 교차 token 요청은 인증 거절된다.

**AC-GT-002** (REQ-GT-002) — Given 계수기와 비밀 표지를 둔 pane bootstrap fixture가 있을 때, When 정상 생성·실패·재전송을 수행할 때, Then tmux global/session env·argv·shell history·로그의 비밀 표지는 0이다. 소비한 bootstrap은 재사용할 수 없고 임시 인증 자료는 정리된다.

**AC-GT-003** (REQ-GT-003) — Given tmux 세션 env에 stale GLM키·타 gateway token·CLAUDE_CONFIG_DIR가 있는 fixture일 때, When A의 pane을 생성하고 요청할 때, Then A의 명시 구성만 쓰고 stale endpoint의 요청 계수는 0이다. 원본 tmux 세션 환경은 수정되지 않으며 A 아닌 pane의 프로필도 유지된다.

**AC-GT-004** (REQ-GT-004) — Given 각 소유 pane이 실행 중인 A/B lead가 있을 때, When A의 정상 종료·오류·강제 종료를 각각 발생시킬 때, Then A pane 후속 송신은 실패하고 A 소유 프로세스가 정리되며 B/C는 유지된다. 이미 전송된 요청은 별도 in-flight로 기록한다.

**AC-GT-005** (REQ-GT-005) — Given Claude Opus5/Sonnet5·GLM tier·네 GPT ID를 설정한 실제 teammate가 있을 때, When 각 정책으로 요청하고 알 수 없는 ID도 요청할 때, Then 본문 canonical ID와 실제 수신 provider가 설정과 일치하며 알 수 없는 ID와 실패 상황의 다른 provider 송신은 0이다.

**AC-GT-006** (REQ-GT-006) — Given 동시 bootstrap과 PID/주소/소유 ID 재사용 및 생성 실패 fixture가 있을 때, When 재개·두 번 소비·지연 생성 완료·부모 종료 뒤 소비를 수행할 때, Then 옛 권한은 거절되고 다른 소유자의 pane을 정리하지 않는다. 일시 파일·pane 누수가 드러나며 각 cleanup 결과가 기록된다.

**AC-GT-007** (REQ-GT-007) — Given 설치 Claude Code와 pane 실행 연결 후보가 있을 때, When 실제 teammate 생성과 bootstrap 전달을 관측할 때, Then 지원된 연결 경로·버전·argv·env 수신·반환 pane ID·수명 증거가 확보된다. 연동 불가/미관측이면 pane 활성화는 거절되고 in-process 제한을 유지한다.

**AC-GT-008** (REQ-GT-008) — Given 원본 permission·profile·MCP와 fallbackModel 오염 fixture가 있을 때, When pane 실행 후 503·인증 오류·제한 모델 요청을 보낼 때, Then 원본 파일 해시는 같고 허용 정책만 사용하며 무단 provider fallback은 0이다. 지원 불가 정책은 명시 오류이고 credential 자동 복사는 없다.

**AC-GT-009** (REQ-GT-009) — Given 실제 A/B/C tmux 공존 환경과 통합 AUTH·PICKER가 있을 때, When pane별 대화·도구 왕복·stream·parent 종료 시나리오를 실행할 때, Then 모든 지정 모델·소유권·수명 판정이 실행 증거로 충족된다. local mock 및 provider 계정 미검증은 별도 Gap으로 남는다.


## 검증과 완료 조건

실제 Claude TUI·client·provider 시험은 2026-09-11 19:00 Asia/Seoul 이후다. Claude 입력은 `claude-opus-5`·
`claude-sonnet-5`, GPT는 `gpt-6-astra`·`gpt-5.6-sol`·`gpt-5.6-terra`·`gpt-5.6-luna`다. 실제 요청 ID·provider·인증·
대화 맥락·도구 왕복·stream 종료를 증거로 남긴다. Sonnet 4.5 과거 캡처와 mock 응답은 실제 새 모델 성공을 대신하지 않는다.
클라이언트 fallbackModel/--fallback-model 및 provider별 effort·format·context 정책도 음성 fixture와 실제 요청으로 검사한다.
설정 필드 삭제나 다른 모델로 자동 변경하여 시험을 통과시키지 않는다. 429·권한 부재는 미완료이며 인증 실패 확정도 아니다.
각 AC의 명령·출력·HEAD·버전·mock/실계정 구분·Gap·잔여 위험을 보고한다. 독립 계획 감사·실행·독립 코드 감사·문서 동기화가
완료 기준이며 원격 출시·push·PR·병합·worktree 제거는 기존 별도 지시를 따른다. 이번 문서 작성은 숫자 카드 생성·DB 변경이나
새 카드 구현 착수를 뜻하지 않는다. 원본 코어·AUTH·PICKER 문서와 다른 세션의 작업을 수정하지 않는다.

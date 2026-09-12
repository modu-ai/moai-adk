# SPEC-MOAI-GATEWAY-TEAMMATE-001 — 구현 계획

## Tier와 선행 조건

Tier L의 근거와 현재 코드 관측은 research.md에 있다. design.md의 최소안과 미결 게이트를 구현 전에 독립 감사한다.

## 실행 순서

1. High — 실제 Claude teammate pane 생성 경로가 pane별 MoAI bootstrap을 연결하는지 읽기 조사와 오후7시 이후 TUI로
   확인한다. seam이 없으면 design 대안을 운영자에게 제시하고 다른 독립 시험 준비만 진행한다.
2. High — 소유 ID·단일 소비·stale env·부모 종료·동시 A/B/C 격리의 실패 시험과 bootstrap 경계를 작성한다.
3. High — 확인된 seam에만 최소 helper·liveness를 연결한다. shared tmux env를 인증 저장소로 쓰지 않는다.
4. High — 실제 model/provider/tool/stream 및 강제 종료 시험으로 pane 지원 활성화 게이트를 판정한다.
5. Medium — 문서와 지원 플랫폼·실패 안내·회귀를 검증한 뒤 독립 감사·동기화한다.

## 검증과 완료 조건

실제 Claude TUI·client·provider 시험은 2026-09-11 19:00 Asia/Seoul 이후다. Claude 입력은 `claude-opus-5`·
`claude-sonnet-5`, GPT는 `gpt-6-astra`·`gpt-5.6-sol`·`gpt-5.6-terra`·`gpt-5.6-luna`다. 실제 요청 ID·provider·인증·
대화 맥락·도구 왕복·stream 종료를 증거로 남긴다. Sonnet 4.5 과거 캡처와 mock 응답은 실제 새 모델 성공을 대신하지 않는다.
클라이언트 fallbackModel/--fallback-model 및 provider별 effort·format·context 정책도 음성 fixture와 실제 요청으로 검사한다.
설정 필드 삭제나 다른 모델로 자동 변경하여 시험을 통과시키지 않는다. 429·권한 부재는 미완료이며 인증 실패 확정도 아니다.
각 AC의 명령·출력·HEAD·버전·mock/실계정 구분·Gap·잔여 위험을 보고한다. 독립 계획 감사·실행·독립 코드 감사·문서 동기화가
완료 기준이며 원격 출시·push·PR·병합·worktree 제거는 기존 별도 지시를 따른다. 이번 문서 작성은 숫자 카드 생성·DB 변경이나
새 카드 구현 착수를 뜻하지 않는다. 원본 코어·AUTH·PICKER 문서와 다른 세션의 작업을 수정하지 않는다.

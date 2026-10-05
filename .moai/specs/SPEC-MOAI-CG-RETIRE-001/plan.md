# SPEC-MOAI-CG-RETIRE-001 — 구현 계획

## Tier와 선행 조건

Tier L의 근거와 현재 코드 관측은 research.md에 있다. design.md의 최소안과 미결 게이트를 구현 전에 독립 감사한다.

## 실행 순서

1. High — live CG 실행·backend 판정·현재 문서/template·테스트와 역사 보존 목록을 경로별로 분류한다.
2. High — design SP-B1의 `moai migrate cg` 명령·두 target·정확한 YAML delta·새 reader/첫 guard와 AC 고정 fixture를
   먼저 테스트하고 구현한다. typed 전체 저장은 사용하지 않는다. TEAMMATE가 필요한 동등성은
   해당 실제 게이트 전 완료로 표시하지 않는다.
3. High — 현재 cg 등록과 모드 적용 연결을 제거하고 지원 launcher 대조·금지 조합을 시험한다.
4. Medium — 같은 변경에서 4locale 현재 문서·template·README·도움말을 고치고 빌드·렌더·링크를 검증한다.
5. High — 통합 launch/이전·독립 감사·동기화로 닫는다.

## 검증과 완료 조건

실제 Claude TUI·client·provider 시험은 2026-09-11 19:00 Asia/Seoul 이후다. Claude 입력은 `claude-opus-5`·
`claude-sonnet-5`, GPT는 `gpt-6-astra`·`gpt-5.6-sol`·`gpt-5.6-terra`·`gpt-5.6-luna`다. 실제 요청 ID·provider·인증·
대화 맥락·도구 왕복·stream 종료를 증거로 남긴다. Sonnet 4.5 과거 캡처와 mock 응답은 실제 새 모델 성공을 대신하지 않는다.
클라이언트 fallbackModel/--fallback-model 및 provider별 effort·format·context 정책도 음성 fixture와 실제 요청으로 검사한다.
설정 필드 삭제나 다른 모델로 자동 변경하여 시험을 통과시키지 않는다. 429·권한 부재는 미완료이며 인증 실패 확정도 아니다.
각 AC의 명령·출력·HEAD·버전·mock/실계정 구분·Gap·잔여 위험을 보고한다. 독립 계획 감사·실행·독립 코드 감사·문서 동기화가
완료 기준이며 원격 출시·push·PR·병합·worktree 제거는 기존 별도 지시를 따른다. 이번 문서 작성은 숫자 카드 생성·DB 변경이나
새 카드 구현 착수를 뜻하지 않는다. 원본 코어·AUTH·PICKER 문서와 다른 세션의 작업을 수정하지 않는다.

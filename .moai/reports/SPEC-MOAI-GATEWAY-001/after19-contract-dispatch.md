# 19시 이후 관측에 따른 계약 보강 인계

부모의 작성 인계 메모다. SPEC 본문이나 감사 판정이 아니다. 기준은 moai-proxy-unified / WT-unified-gateway / HEAD 81c1d58f9이며 source_session_id는 01a08e7b-6aa0-7361-ab7e-ea8da1f02228이다.

## 이미 결정되거나 직접 확인한 입력

- 사용자는 Windows 실행 증거를 GitHub CI에서 확보하도록 결정했다. 원격 push·PR·병합의 별도 지시 경계는 유지한다.
- MoAI 전용 OAuth 등록 정보는 없다. 새 MoAI scratch에서 공식 Codex broker 로그인은 generation 1, LOGIN_PUBLISHED, exit 0이었다. 기존 사용자 credential 복사를 승인한 것이 아니다.
- M0 정상 응답과 직접 refresh 증거는 `m0-after19-observation.md`, `m0-refresh-transport-observation.md`다. 부모가 두 보고서와 직접 refresh JSON을 읽었다. 별도 X-MoAI-Session-Token과 OAuth Bearer/beta의 공존, 실제 refresh POST 200의 반환 토큰과 같은 본문 재전송 200의 hash 일치가 확인되었다. 제품 gateway의 전체 수용 판정과 구별한다.
- 실제 Claude TUI의 네 GPT 표시·선택 요청 증거는 `picker-reasoning-runtime-observation.md`, `picker-followup-runtime-observation.md`다. s는 세션 선택, Enter는 기본값 선택이다. Default의 화면 표시와 현재 요청은 Sol이며 직후 settings.json은 model 없는 빈 객체였다. 재시작 적용은 미측정이다.
- 합성 redacted_thinking 68바이트와 hash 결합 tool ID 172바이트의 실제 Read 왕복은 첫 picker 보고서에 있다. 실제 opaque·resume·유실 탐지의 증거를 대신하지 않는다.
- 요청의 실제 raw 정책은 본 대화 adaptive/high/max_tokens 32000/context_management clear_thinking keep all, 제목 요청 thinking 부재/high/title JSON schema다. Opus·Sonnet의 별도 원본 정책은 M0 보고서에 있다.

## GPT 구독 프로토콜에서 확인 중인 차이

관측 작성자는 gateway_translation이며 private auth-live probe와 증거를 소유한다. 최종 보고서를 받은 뒤 아래 항목의 원문과 해시를 인계에 연결해야 한다.

- 네 GPT exact ID의 text 요청은 HTTP 200, Content-Type 없음, 무압축 UTF8 SSE, completed 이벤트 판독 성공이다. 공식 Codex source의 해당 stream 경로도 body eventsource 판독을 사용한다. 이 관측이 기존 제품 media 계약을 자동 변경하지 않는다.
- Astra는 max_output_tokens를 HTTP 400으로 거절했고 안전한 분류는 unsupported_parameter/max_output_tokens다. 부모는 사용자에게 구독 서버 출력 정책을 사용하는 계약 변경 여부를 질문했다. **답변 대기 중이며 이 값을 조용히 생략하는 제품 변경은 아직 승인하지 않는다.** 바이트 제한이나 입력 토큰 추정을 생성 토큰 상한과 같다고 쓰지 않는다.
- 실제 stream은 output_item.done에 완료된 함수 호출을 보내고 response.completed.output은 빈 배열일 수 있었다. ordered item 이벤트의 독립 검증 및 공식 client 처리와 대조하고, 중복·순서·ID·잘림 오류를 성공으로 바꾸지 않는 계약이 필요하다. 현재 도구 왕복 관측 중이다.
- high 요청이 opaque item을 항상 반환한다는 보장은 아직 없다. 실제 반환 item의 원문 바이트 보존·다음 요청 수용과, 반환하지 않은 경우를 분리한다.

## Windows 계약 검토 입력

gateway_windows_store가 AUTH plan의 internal/atomicfile.Replace 강제와 Windows native write-through 후보의 차이, 디렉터리 내구성 성공 기준의 미정의를 직접 확인했다. 공식 API 근거와 정확한 인용을 별도 보고서로 작성 중이다. Windows 실제 시험을 CI로 옮기는 결정만으로 내구성 요구를 삭제하지 않는다. 독립적인 LockFileEx·취소·프로세스 종료 시험 준비는 계속한다.

## 작성·검증 경계

manager-spec은 원문 보고서와 실제 승인 답변을 읽고 필요한 본문 계약을 보강한다. 미관측 항목을 PASS로 변경하거나 기존 AC를 제거하여 통과시키지 않는다. 코드가 존재하는 부분의 수정 계약과 아직 구현하지 않은 opaque 계약을 분리하여 감사 범위를 명시한다. 실제 TUI resume·foreign-provider 전환·carrier 유실의 송신 0은 구현 후 별도 검증할 항목이다.

제품 진입 factory, PICKER의 사용자 설정 보존, TEAMMATE의 native 수명·인증 연결, 실제 네 GPT의 Claude Code 왕복, Windows CI는 전체 완료까지 계속 남아 있다. 이 인계는 범위 축소나 목표 완료가 아니다.

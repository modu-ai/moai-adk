# t649 MoAI 실행 경로 E2E 하네스

`gateway_probe.py`는 지정한 MoAI 바이너리를 PTY로 실행한다. 원본 터미널 출력, 토큰, 응답 본문은 저장하지 않는다. 합성 표식의 최종 응답 포함 여부, 실제 assistant 모델, Read 호출, 세션 ID, 오류 유형, 모델 선택기의 번호 행만 JSON으로 남긴다. 유료 추론은 `--allow-live`가 있어야 실행한다.

## 실행

기존 인증을 제품 저장소를 통해 사용한다. `--profile`은 이미 존재하는 프로필만 지정한다. 작업 디렉터리는 합성 검증용으로 별도 준비하며 신뢰/온보딩 화면을 자동 승인하지 않는다. `--settings`로 hooks를 끄고 MCP 서버를 빈 목록으로 제한한다.

```bash
python3 gateway_probe.py --binary /tmp/moai-t649 --cwd /path/to/trusted-synthetic-project --output picker-gpt.json --mode gpt --model gpt-5.6-sol --journey picker --seconds 22 --expected-models gpt-6-astra gpt-5.6-sol gpt-5.6-terra gpt-5.6-luna
python3 gateway_probe.py --binary /tmp/moai-t649 --cwd /path/to/trusted-synthetic-project --output answer.json --mode gpt --model gpt-6-astra --journey answer --allow-live
python3 gateway_probe.py --binary /tmp/moai-t649 --cwd /path/to/trusted-synthetic-project --output tool.json --mode gpt --model gpt-6-astra --journey tool --allow-live
python3 gateway_probe.py --binary /tmp/moai-t649 --cwd /path/to/trusted-synthetic-project --output resume.json --mode gpt --model gpt-6-astra --journey resume --resume-session UUID_FROM_ANSWER --allow-live
```

cc/glm도 동일한 picker 명령에서 mode/model/expected-models를 실제 제품 목록으로 바꾼다. 목록은 추측해서 채우지 않는다. picker는 `/model` 뒤 위쪽 행으로 이동하여 `s`로 현재 세션에만 적용한다. 성공에는 목록의 정확한 집합 일치와 세션 한정 선택 확인 문구가 모두 필요하다. Default 행은 별도로 기록한다. live 응답 성공에는 exit=0, result 이벤트의 합성 표식, 오류 없음이 필요하며 tool은 Read 이벤트도 필요하다. 실제 모델 ID는 별도 확인한다.

## 하네스 검증과 한계

`harness-selftest.json`은 기존 직접 Claude 실행 화면 세 건의 행 추출/선택 확인 파서와 합성 자식 프로세스 검증이다. MoAI 제품 E2E 통과 증거가 아니다. 살아 있는 합성 프로세스를 제한 시간으로 종료한 별도 검사에서 exit=-15, timeout=true, cleanup_errors=[]를 관측했다. 이미 종료된 그룹에 신호를 보낼 때 macOS가 권한 오류를 반환할 수 있어 cleanup_errors로 드러낸다. 모든 실행은 finally에서 그룹 SIGTERM/SIGKILL을 시도한다. 신뢰/로그인 화면, 관리 설정 우선권, 실제 모델 사용 권한은 제품 실행에서 검증해야 한다.

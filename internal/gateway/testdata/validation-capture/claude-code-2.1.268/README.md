# Claude Code 2.1.268 실제 요청 픽스처

2026-09-11 억제 플래그 없이 로컬 mock에서 캡처한 역사적 요청이다. Sonnet 4.5 모델명은 당시 관측값이며 후속 실측 모델 기준이 아니다. 후속 실측은 Opus 5와 Sonnet 5를 쓴다.

request-003은 /model 선택 검증 요청이며 나머지 네 건은 첫 turn, 후속 turn, 제목 생성의 음성 대조군이다. 원본은 `.moai/state/verify/01a08e7b-6aa0-7361-ab7e-ea8da1f02228/gateway-entry/m1-raw/`에 보존한다.

manifest.json은 원본과 익명화 픽스처 SHA-256 및 변경된 JSON 경로를 기록한다. 재귀 순회로 문자열 내 기계 로컬 임시 경로를 /fixture/project로, metadata.user_id 내부 device_id·session_id를 fixture 값으로 치환했다. 빈 account_uuid는 그대로다. JSON 직렬화 공백은 바뀌었으나 모델, stream 키 유무·값, max_tokens 정수, 메시지 수·역할, tools 키 유무·배열 구조는 보존했다. 원본 파일은 수정하지 않았다.

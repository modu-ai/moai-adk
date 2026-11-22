# `host-wake`

operation `host.wake` · 코드 `src/ops/host_wake.rs`

`host-wake NAME --text T [--idle] [--config-dir D]` 는 흐름의 깨우기 하나를 호스트 어댑터로 판정한다(SPEC-20261122-session-messages). 설정 루트(CLAUDE_CONFIG_DIR, 없으면 ~/.claude)의 세션 레지스트리(sessions/*.json)를 읹는데, 레지스트리 폴더가 없거나 읽을 수 없으면 내부 오류가 아니라 무채널(no channel) 거부 모양으로 답한다. 목록에 이름이 없으면 다른 루트, 채널이 없으면 무채널, 소켓이 없으면 무수신함, 같은 이름이 둘 이상이면 모호함, `--idle` 는 지원하지 않음(일회 알림은 모델 세션 도구 길)으로 한 줄 거부 사유를 답한다.

송신(D-8, 리더 조사로 잰 전선 형식): 받는 쪽이 받아들이는 최소 줄 `{"type":"user","message":{"role":"user","content":"…"}}` 에 줄바꿈을 붙여 소켓에 한 번 쓰고 닫는다 — 답은 queued(줄이 이 프로세스를 떠났다), `socket_write: true`(그 한 줄 외에는 아무것도 쓰지 않는다). 인증 줄 없이(남의 토큰은 읽지 않음), from-mode 도 밝히지 않는다(꾸미지 않음). bypass 로 도는 레인은 이런 메시지를 보류하니 moai 는 설치·업데이트 기본값으로 crossSessionInbound accept 를 사용자 설정에 쓴다(미리보기 기본 항목 + 이 컴퓨터의 세션들이 서로 검토 없이 메시지를 주고받게 된다는 한 줄과 끄는 법). 보류·전달 통지는 보낸 쪽 소켓으로 오는데 이 명령은 소켓이 없어 답은 queued 모양이고, 시험은 시험이 직접 임시 소켓(및 임시 세션)에만 보낸다.

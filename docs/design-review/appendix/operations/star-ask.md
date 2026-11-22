# `star-ask`

operation `star.ask` · 코드 `src/ops/star_ask.rs`

`star-ask [--dismiss]` 는 설치·업데이트 끝의 정중한 스타 한 줄을 판정한다(SPEC-20261122-star-and-reports). 사용자의 gh 로그인으로 정확히 한 번 GET(`gh api /user/starred/modu-ai/moai-adk`)을 읽어 상태를 묻고, 모르거나(gh 없음·실패) 스타하지 않았으면(404) 정중한 한 줄을 답한다. 스타했으면(204) 줄은 절대 나오지 않고, `--dismiss` 로 해지 마커를 홈 moai 상태에 남기면 그 뒤로는 상태와 무관하게 나오지 않는다. 확인 경로에는 쓰기 호출이 없다(시험이 스탠드인 gh 로 기록해 증명). 아무것도 쓰지 않는다(해지 마커 외).

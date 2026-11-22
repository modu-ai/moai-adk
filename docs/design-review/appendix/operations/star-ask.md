# `star-ask`

operation `star.ask` · 코드 `src/ops/star_ask.rs`

`star-ask [--dismiss]` 는 정중한 스타 한 줄을 판정한다(SPEC-20261122-star-and-reports). `--dismiss` 없이 부르면: 해지 마커가 있으면 네트워크 호출 없이 바로 dismissed 로 답하고, 없으면 사용자의 gh 로그인으로 정확히 한 번 GET(실제 호출: `gh api -X GET -i https://api.github.com/user/starred/modu-ai/moai-adk` — URL 고정, 본문·추가 헤더 없음, gh 자신의 인증만, 60초 상한)을 읽어 상태를 묻고, 모르거나 스타하지 않았으면 정중한 한 줄을 답한다. 줄은 공용 UI 텍스트 표(`src/ui_text.rs`)의 영어·한국어 쌍에서 스트림 언어로 고른다(모든 화면과 같은 경로; 파이프·스크립트는 영어). `--dismiss` 를 붙이면 해지 마커를 현행 세대 홈 폴더에 쓰고, 어떤 결과든 같은 한 모양 `{written, reason}` 으로 답한다 — 쓰기 성공이면 `written: true`·`reason: null`, 홈이 은퇴 폴더(.moai-cli)를 가리키면 `written: false` 와 그 이유, 쓰기가 거부되면 `written: false` 와 그 이유 — 종료 코드는 셋 다 0 이고 다른 실행을 막지 않는다. HOME 이 절대 경로가 아니면 거부한다. 확인 경로에 쓰기 호출은 없다. 아무것도 설치·업데이트 훅에서 이 명령을 부르지 않는다 — 명시적 호출만.

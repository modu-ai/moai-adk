# `star-ask`

operation `star.ask` · 코드 `src/ops/star_ask.rs`

`star-ask [--dismiss]` 는 정중한 스타 한 줄을 판정한다(SPEC-20261122-star-and-reports). `--dismiss` 없이 부르면: 해지 마커가 있으면 네트워크 호출 없이 바로 dismissed 로 답하고(B1), 없으면 사용자의 gh 로그인으로 정확히 한 번 GET(`gh api https://api.github.com/user/starred/modu-ai/moai-adk` — URL 고정, 본문·헤더 없음, gh 자신의 인증만, 60초 상한)을 읽어 상태를 묻고, 모르거나 스타하지 않았으면 정중한 한 줄을 스트림 언어로 답한다(영어·한국어 쌍). `--dismiss` 를 붙이면 해지 마커를 현행 세대 홈 폴더에 쓰고 그 결과를 성실히 답한다(쓰기 성공 true, 은퇴 폴더면 false 이유와 함께, 쓰기 거부면 그 이유). HOME 이 절대 경로가 아니면 거부한다(N2). 확인 경로에 쓰기 호출은 없다. 아무것도 설치·업데이트 훅에서 이 명령을 부르지 않는다 — 명시적 호출만.

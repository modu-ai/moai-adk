# SPEC-MOAI-CG-RETIRE-001 — 조사

## 직접 읽은 기준선

HEAD `81c1d58f9`의 `internal/cli/cg.go:57`은 rootCmd.AddCommand(cgCmd)를 호출하며 runCG가 존재한다.
`launcher.go:311` applyCGMode는 GLM 키를 읽고 `:362` 이후 설정 변경과 tmux 주입을 수행한다.
`internal/config/team_mode.go`는 TeamModeCG를 GLM backend signal 값으로 정의한다. `internal/tmux/cg_detect_ssot_test.go`
등에는 기존 CG 성공/감지 계약이 있다. `docs-site/hugo.toml:94,103,112,121`은 ko/en/ja/zh contentDir를 둔다.
코어 spec의 CG 제외 범위에는 문서/template/테스트/교차 SPEC 스윕이 기록되어 있다. 그 과거 개수는 이번 측정값이 아니다.

따라서 CLI·config·tmux·template·4locale 문서의 여러 영역을 다루며 Tier L이다. 실행·테스트는 이번 작성 중 하지 않았다.


## SP-B1 추가 판독

`internal/config/types.go:265` LLMConfig에는 team_mode, dormant mode와 기존 GLM/profile 값이 있다. gateway teammate
두 필드와 ReadGatewayTeammatePolicy/GuardLegacyCG는 이번 계획의 새 구현 대상이며 기존 live reader가 아니다.
이전 명령은 기존 migrate root 아래에 cg를 추가하는 것으로 확정했다. 손실 없는 편집은 typed 저장을 기반으로 삼지 않는다.

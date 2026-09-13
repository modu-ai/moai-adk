# SPEC-MOAI-GATEWAY-TEAMMATE-001 — 조사

## 직접 읽은 기준선과 한계

HEAD `81c1d58f9`에서 `internal/cli/spawn.go:78` 이후는 호출자의 tmux 세션에 new-window를 만들며 cwd와 명령 문자열을
넘기고 pane ID를 받는다. `internal/tmux/session.go:166`은 split-window 뒤 sendKeys 방식으로 명령을 넣는다.
같은 파일 InjectEnv는 set-environment를 사용하며 InjectSensitiveEnv도 전송 수단을 바꿀 뿐 세션 env 대상이다.
따라서 현재 이 함수들이 pane 전용 인증 격리를 제공한다고 판단하지 않는다. 이는 코드 판독이며 실유출 재현이 아니다.

설치 man `/opt/homebrew/Cellar/tmux/3.6a/share/man/man1/tmux.1:3104,3157,3509`를 읽었다.
new-window와 split-window는 명령·인수와 -e environment를 받으며 new-window 설명은 새 창 환경을 지정한다고 적는다.
실제 tmux·Claude TUI·provider는 실행하지 않았다. tmux가 이를 지원한다는 문서만으로 Claude teammate 연동을 보장하지 않는다.

CLI·tmux·gateway auth/lifetime·hook과 실제 동시 pane의 새 신뢰 경계를 함께 바꾸므로 Tier L이다. 코어 spec Out of Scope의
same-session ownership·모델·수명·15키 이전 논의는 범위 입력이며 현재 코드 검증 결과와 구분한다.

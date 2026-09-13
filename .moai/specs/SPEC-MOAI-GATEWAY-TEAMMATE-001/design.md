# SPEC-MOAI-GATEWAY-TEAMMATE-001 — 설계

## 최소 연결 후보와 미결 게이트

tmux 세션 env를 공유 저장소로 쓰지 않는다. MoAI가 제어하는 pane 시작 명령에 bootstrap helper를 연결하여 새 프로세스에서
정리된 env를 조립하는 방식이 우선 후보다. `new-window`/`split-window`의 `-e`는 비밀이 아닌 owner 식별에만 사용할 수
있으며 token 값을 argv로 전달하지 않는다. private 단일 소비 bootstrap 자료에는 해당 gateway의 접근 token만 담고 provider
원본 credential은 담지 않는다. 파일 방식 후보는 사용자 전용 디렉터리·소유권·symlink/중복 소비·부모 생존·만료 검사를
통과해야 한다. 파일 경로를 안다는 것만으로 다른 owner의 bootstrap을 소비하지 못하게 연결·소비 권한을 검증한다.

중요한 미결은 Claude Code가 실제 teammate를 만들 때 이 helper를 pane별로 연결할 지원 seam이 있는지다. 현재 MoAI의
spawn 함수가 명령을 받는다는 사실은 Claude Code 내부 teammate 생성에 도달한다는 증거가 아니다. 실제 client argv와 pane
생성·hook 경로를 관측해 지원 seam을 먼저 확인한다. PATH의 tmux를 몰래 가로채는 wrapper나 비공식 experimental
`--teammate-mode` 플래그를 자동 채택하지 않는다. 코어 wording-only 결정은 pane 격리 승인·검증으로 해석하지 않는다.

지원 seam이 없으면 대안은 사용자가 명시 실행하는 MoAI 소유 pane 작업 경로 또는 lead별 별도 tmux server 경계다.
전자는 Claude native teammate와 기능 동등성을 추가 확인해야 하며, 후자는 같은 tmux 세션 공존 경험을 바꾸므로 운영자
결정이 필요하다. 현재 둘 다 채택하지 않는다. 이 경우 실행 gateway 코어는 계속 in-process로 제한하고 본 SPEC 전체는
미완료다. 숫자 카드·새 worktree 생성이나 사용자 경험 변경을 이 계획에서 실행하지 않는다.

부모 연결은 PID 하나가 아니라 launch instance identity·유효 gateway token·liveness 채널을 결합한다. bootstrap과 실제
요청은 부모 종료 직전/직후를 판정한다. 소유한 pane ID/프로세스만 닫고 tmux 세션 전체나 다른 세션 env는 지우지 않는다.
비gateway pane C가 상속한 옛 사용자 GLM 환경을 정리했다고 주장하지 않는다. 보장 범위는 gateway가 소유한 pane이다.
공유 stale tmux env를 강제 삭제하는 기존 후보 대신 무변경·자식 경계 정리를 채택한다.

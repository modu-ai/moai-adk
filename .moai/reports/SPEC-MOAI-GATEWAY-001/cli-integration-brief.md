# CLI 통합 구현 인계 메모

지정 WT `moai-proxy-unified`, HEAD `81c1d58f9`에서 부모가 `launcher.go:130-220`, `:620-850`, `:1145-1183`과 `cc.go:80-160`을 직접 읽었다. 아래는 구현 지시와 경계이며 완료 보고가 아니다.

- `unifiedLaunchDefault(profileName, modeOverride, extraArgs)`가 모드를 해석하지만 마지막 `launchClaude(profileName, extraArgs)`에는 모드가 전달되지 않는다. gateway 통합에서는 명시 인자로 전달한다. 전역 current-provider 변수로 보충하지 않는다.
- 기존 `applyCCMode`·`applyGLMMode`를 gateway 경로에 그대로 호출하면 tmux·GLM process env·설정 부작용이 다시 생긴다. gateway 전용 준비 경로를 공통 plan으로 조립하고 기존 프로필 해석·worktree·spawn·kanban/factory 진입은 보존한다.
- `launchClaudeDefault`는 profile 기본값·프로젝트 DO_CLAUDE 값·명시 모델을 차례로 읽는다. gateway의 초기 모델 우선순위를 명시하고 프로필의 정상 설정을 무음으로 버리지 않는다. 사용자 `--model=value`·`--` 인수도 계약에 맞게 판정한다.
- `--continue`는 별도 `exec.Command.Run` 경로다. 그 뒤 정상 exec 경로에만 env를 추가하면 continue가 gateway를 우회한다. 최초 continue, 정상 종료, exit 1 뒤 새 세션 fallback 모두 같은 준비된 gateway·env·overlay를 사용하도록 시험한다. 재개 대화 기록은 보존한다.
- `appendCrossSessionSettings`가 앞에서 `--settings`를 만들 수 있고 사용자가 직접 전달할 수도 있다. gateway overlay를 더할 때 기존 설정과 권한 의미를 누락하거나 조용히 바꾸지 않는다. 실제 설정 우선순위와 `/model` 저장 격리는 PICKER의 사전 측정 게이트다.
- POSIX의 `execOrSpawnClaudeFunc` seam과 실제 `syscall.Exec`, Windows의 기존 spawn/wait·종료 코드·profile lease·session PID 계약을 유지한다. launcher가 exec 뒤 defer를 수행한다고 가정하지 않는다.
- gateway용 자식 env는 상속 14키와 provider별 `Z_AI_API_KEY`를 먼저 정리한 뒤 loopback 주소·세션 인증·초기 provider를 추가한다. GLM 슬롯은 저장소에서 해석한 high/medium/low/fable 값만 재주입한다. tmux 세션 env를 고치지 않는다.
- 새 provider signal은 config 상수와 kanban/backend UI에서 같은 어휘를 쓴다. SessionStart·SessionEnd의 gateway guard는 초기 provider를 현재 요청 provider로 오인하지 않는다.
- `gpt`는 닫힌 하위 동사와 기존 spawn/프로필/작업 트리 표면을 제공한다. login/logout은 AUTH 구현을 연결하며 사용자 오류에 내부 SPEC 번호를 노출하지 않는다. `gg`는 만들지 않는다.

M0 세션 인증 운반 키, AUTH 실제 로그인, PICKER 설정 저장 격리, opaque reasoning 운반은 아직 미검증이다. 이 메모는 그 결정을 대신하지 않는다. 일반 진입 코드를 썼다는 이유로 core·AUTH·PICKER의 통합 게이트를 열지 않는다.

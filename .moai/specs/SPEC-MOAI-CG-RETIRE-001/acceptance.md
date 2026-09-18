# SPEC-MOAI-CG-RETIRE-001 — 수용 기준

**AC-CR-001** (REQ-CR-001) — Given CG 명령과 help·잘못된 인수·-k·-f·--spawn 조합이 있을 때, When 각 실행을 명령 seam에서 호출할 때, Then cg는 실행 가능 목록에서 사라지고 폐기 안내 또는 명시 오류가 발생하며 exec·tmux·credential write 계수는 0이다.

**AC-CR-002** (REQ-CR-002) — Given team_mode cg와 관련 없는 사용자 키가 있는 격리 프로젝트일 때, When 새 launcher가 이전 구성을 읽거나 무인 실행할 때, Then 원본 해시가 그대로이고 미이전 오류가 나타나며 다른 provider 송신은 0이다. 명시 정상 구성 대조군은 실행 분기에 닿는다.

**AC-CR-003** (REQ-CR-003) — Given 명시 이전 선택과 원본 snapshot이 있을 때, When 정상 이전·취소·동시 수정·write 실패·재실행을 각각 수행할 때, Then 선택한 필드만 변경되고 나머지 값과 원본 snapshot이 보존된다. 동시 변경은 덮어쓰지 않고 실패하며 재실행은 같은 결과다.

**AC-CR-004** (REQ-CR-004) — Given CG 전용 호출 경계에 계수기를 둔 현재 launcher들이 있을 때, When cc·glm·gpt 및 legacy cg 입력을 각각 실행할 때, Then 지원 경로는 새 provider 판정으로 실행되고 CG 주입 계수는 0이다. 과거 backend 기록의 cg 문자열은 현재 실행 provider로 오해하지 않고 읽을 수 있다.

**AC-CR-005** (REQ-CR-005) — Given kanban/factory·프로필·worktree·spawn 입력 행렬이 있을 때, When 지원 입력과 CG legacy 입력을 각각 파싱·실행 seam으로 보낼 때, Then 지원 동작은 보존되고 폐기 입력이 다른 유료 provider로 바뀌지 않는다. 권한·사용자 argv 의미도 유지된다.

**AC-CR-006** (REQ-CR-006) — Given 현재 문서/template 목록과 역사 보존 목록이 있을 때, When 4locale 빌드·링크 검사와 template 렌더·도움말 비교를 수행할 때, Then 새 실행 안내는 지원 launcher/이전 정책과 일치하고 CG 실행을 권하지 않는다. 역사 파일과 사용자 파일의 전후 해시는 같다.

**AC-CR-007** (REQ-CR-007) — Given 이전 CG 성공 테스트 및 지원 launcher 대조 테스트가 있을 때, When 현재 범위 단위·통합 테스트를 실행할 때, Then CG 실행·자동 provider 변경 뮤턴트는 실패하고 지원 launcher 대조군은 통과한다. 단순 문자열 검색 결과와 runtime 계수를 별도 보고한다.

**AC-CR-008** (REQ-CR-008) — Given leader Claude/teammate GLM 의도가 있는 이전 프로젝트일 때, When 동등 구성을 선택하고 실제 TEAMMATE 통합 시험을 수행할 때, Then leader·pane의 요청 ID·provider·auth·lifetime이 선택한 구성과 일치한다. TEAMMATE 게이트 미충족이면 이전 전체 완료는 미완료다.


## AC-CR-002·003 고정 fixture (SP-B1)

원본 `.moai/config/sections/llm.yaml`:

```yaml
llm:
  team_mode: cg # 기존 혼합 역할
  glm_env_var: MY_GLM_KEY
  glm:
    models: {high: glm-5.2, medium: glm-4.7, low: glm-4.5-air, fable: glm-4.7}
  unknown_keep: yes # 사용자 주석 보존
```

`moai migrate cg --target claude-only --apply --accept-role-change` 후 기대값:

```yaml
llm:
  team_mode: claude # 기존 혼합 역할
  glm_env_var: MY_GLM_KEY
  glm:
    models: {high: glm-5.2, medium: glm-4.7, low: glm-4.5-air, fable: glm-4.7}
  unknown_keep: yes # 사용자 주석 보존
  gateway:
    teammate_mode: in-process
    teammate_provider: inherit
```

원본 backup bytes와 SHA-256은 정확히 같아야 하며 출력에는 역할 변화가 표시된다. formatting 전체 동일성을 요구하지 않지만
두 주석·unknown 값·GLM 모델·credential 참조는 그대로여야 한다. `--target claude-glm --apply`의 기대 delta는 위 결과에서
`teammate_mode: tmux`, `teammate_provider: glm`뿐이며 TEAMMATE 실제 capability가 없으면 원본·backup 모두 변화 0이다.

- AC-CR-002: 원본으로 cc/glm/gpt, --model 지정, continue/resume/spawn, -k/-f를 각각 실행하면 **첫 guard**에서 거절되고
  typed backend 해석·credential 조회·worktree 생성·tmux 실행·exec 계수는 0이다. 같은 fixture의 claude-only 결과는
  명시 launcher의 준비 seam까지 도달해야 한다. hybrid 결과는 cc+실제 capability에서만 도달하고 glm/gpt는 거절한다.
- AC-CR-003: preview, --apply만, target만, claude-only 수락 누락, 알 수 없는 target, 독립 mode:glm, 중복 team_mode,
  충돌 gateway 값을 각각 시험한다. preview는 변화0의 정상 종료, 잘못된 apply는 변화0의 오류다.
  동일 target 재실행은 unchanged이며 반대 target은 오류다. 잠금 전후 원본 변경은 덮어쓰기0이다. 저장 실패 주입은
  backup/원본 상태와 오류를 읽어 확인하고 readback 실패를 applied 성공으로 처리하는 뮤턴트는 적색이다.
- 새 reader/guard를 빼고 typed `IsGLMBackend`로 먼저 진입하는 뮤턴트, team_mode만 claude로 바꾸고 역할 정책을
  누락하는 뮤턴트, 의미 변화 수락 없이 apply하는 뮤턴트가 각각 실패해야 한다. reader 존재 검색만으로 통과하지 않는다.

## 검증과 완료 조건

실제 Claude TUI·client·provider 시험은 2026-09-11 19:00 Asia/Seoul 이후다. Claude 입력은 `claude-opus-5`·
`claude-sonnet-5`, GPT는 `gpt-6-astra`·`gpt-5.6-sol`·`gpt-5.6-terra`·`gpt-5.6-luna`다. 실제 요청 ID·provider·인증·
대화 맥락·도구 왕복·stream 종료를 증거로 남긴다. Sonnet 4.5 과거 캡처와 mock 응답은 실제 새 모델 성공을 대신하지 않는다.
클라이언트 fallbackModel/--fallback-model 및 provider별 effort·format·context 정책도 음성 fixture와 실제 요청으로 검사한다.
설정 필드 삭제나 다른 모델로 자동 변경하여 시험을 통과시키지 않는다. 429·권한 부재는 미완료이며 인증 실패 확정도 아니다.
각 AC의 명령·출력·HEAD·버전·mock/실계정 구분·Gap·잔여 위험을 보고한다. 독립 계획 감사·실행·독립 코드 감사·문서 동기화가
완료 기준이며 원격 출시·push·PR·병합·worktree 제거는 기존 별도 지시를 따른다. 이번 문서 작성은 숫자 카드 생성·DB 변경이나
새 카드 구현 착수를 뜻하지 않는다. 원본 코어·AUTH·PICKER 문서와 다른 세션의 작업을 수정하지 않는다.

# SPEC-MOAI-CG-RETIRE-001 — 설계

## 구조 결정

CG는 다른 launcher의 자동 별칭으로 남기지 않는다. root 등록·applyCGMode의 실행 연결을 제거하고 폐기 토큰에 대한 짧은
진단만 허용한다. `TeamModeCG`의 runtime provider 의미는 제거하되 legacy 데이터 판독·이전 식별은 별도 경계로 보존한다.
`sessionEnvHasGLM`·`hasGLMEnv`는 호출자를 조사하여 CG 전용 도달 경로만 제거하고 일반 tmux 기능을 정리하지 않는다.

## SP-B1 확정 계약 (0.1.0 보강)

명령은 기존 migrate root 아래 `moai migrate cg`다. 프로젝트 루트의 `.moai/config/sections/llm.yaml`만 대상으로 한다.
`--target`은 `claude-only` 또는 `claude-glm` 두 값이다. 기본은 preview이며 `--apply`와 target 없이는 쓰지 않는다.
`--target claude-only --apply`에는 `--accept-role-change`도 필요하다. 이 플래그는 기존 GLM teammate 자동 배정을
제거한다는 명시 수락이며 무인 실행에도 동일하다. 의미 변화는 stdout에 표시하되 credential 값은 출력하지 않는다.
`--target claude-glm`은 기존 역할 보존이므로 해당 플래그를 받지 않는다. 알 수 없는 옵션·누락 값은 변경 전 오류다.

| 명령/입력 | 정확한 저장 결과 | reader와 guard |
|---|---|---|
| `moai migrate cg` 또는 `--target <값>` | 원본 유지, preview만 출력 | target 미선택이면 두 선택의 역할 변화·게이트를 표시 |
| `moai migrate cg --target claude-only --apply --accept-role-change` | `llm.team_mode: claude`; `llm.gateway`가 없던 원본에 `teammate_mode: in-process`, `teammate_provider: inherit` 두 키 생성 | 새 gateway 정책 reader가 모델별 자식 정책을 읽음; 혼합 강제 배정 없음 |
| `moai migrate cg --target claude-glm --apply` | `llm.team_mode: claude`; `llm.gateway.teammate_mode: tmux`, `teammate_provider: glm`; 기존 `llm.glm.models`·credential 참조 유지 | 새 reader가 lead=Claude, pane=코어 GLM tier 정책을 TEAMMATE 연결에 전달; 실제 capability 없으면 **쓰기 0** |
| target/의미 변화 수락 없음, unknown key/value 충돌, TEAMMATE 미충족 | 변경·backup 생성 0, 명시 오류 | `team_mode: cg` 미이전 guard 유지 |

`llm.gateway.teammate_mode` 허용값은 `in-process|tmux`, `teammate_provider`는 `inherit|glm`이다. 두 키는 이 SPEC에서 새로
정의하며 현재 reader가 이미 있다는 주장은 아니다. 유효 조합은 `in-process/inherit`, `tmux/glm` 둘뿐이다. existing gateway
객체의 다른 키는 보존하고 위 두 키가 이미 다른 값이면 덮어쓰지 않고 충돌 오류다. 기존 `llm.mode`가 비어 있거나 없을
때만 이 이전을 허용한다. `mode: glm` 같은 독립 signal이 있으면 자동 삭제하지 않고 충돌 오류를 반환한다. GLM-only/GPT-only
target은 이 좁은 이전 명령에 없다. 사용자는 claude-only의 의미 변화 수락 뒤 지원 launcher를 명시 선택할 수 있다.

**새 소비 경계.** 공통 gateway launch 준비에 `ReadGatewayTeammatePolicy`와 `GuardLegacyCG`를 추가한다(구현 예정).
원본 YAML을 읽은 직후, typed decode·template backend 해석·apply mode·profile lease·worktree/spawn·kanban/factory
dispatch·credential 조회보다 앞에서 guard를 수행한다. `llm.team_mode`가 `cg`이면 cc/glm/gpt와 모든 continue/resume/spawn
진입을 거절하고 위 preview 명령을 안내한다. 이 거절은 `--model` 지정으로 해제되지 않는다. 중복 YAML key·alias로 모호한
team_mode/gateway 경로도 실패한다. 루트 경로를 찾기 위한 읽기 외에는 부작용이 없어야 한다.

이전 후 `claude/in-process/inherit`은 cg guard를 해제하고 명시 launcher를 그대로 사용한다. `claude/tmux/glm`은
`moai cc`만 허용하며 glm/gpt 진입은 역할 정책 충돌로 거절한다. 이 조합은 apply 시점과 매 launch 시점 모두
TEAMMATE의 native pane 연결·소유권·인증·수명 지원이 확인된 설치 버전에 한해 허용한다. 사용자가 쓰는 `verified:true`
같은 설정으로 우회하지 않는다. 독립 TEAMMATE 통합 게이트가 닫히면 원본 hybrid apply는 쓰기 0이고, 이미 이전된 hybrid
파일도 launch 실패다. 이 reader/guard는 코어와의 연결을 별도로 구현·시험해야 하며 문서만으로 도달성을 주장하지 않는다.

**손실 없는 저장.** typed `saveLLMSection`으로 전체 문서를 다시 만들지 않는다. YAML node로 대상 scalar와 새 mapping만
편집하여 알 수 없는 키의 값·credential 참조·주석을 보존한다. duplicate key·지원하지 않는 alias 편집은 명시 실패한다.
원본 bytes·SHA-256을 확보하고 exclusive lock 아래 재측정한다. 변경된 다른 세션 파일은 덮어쓰지 않는다. backup은
`.moai/backups/cg-migration/<원본SHA256>.yaml`에 owner-only 권한으로 원본 bytes를 독점 생성한다. 기존 같은 이름 파일이
있으면 bytes가 일치할 때만 재사용한다. 임시 쓰기·검증·atomic replace·readback이 모두 성공한 뒤 applied를 표시한다.
preview·취소·사전 게이트 실패는 backup도 만들지 않는다. write 도중 실패 시 backup은 남을 수 있으나 원본은 유효한 이전
또는 새 상태여야 하고 부분 YAML은 금지한다. readback 실패는 성공으로 표시하지 않고 보존 backup 경로를 안내한다.
같은 target의 결과가 이미 저장되어 있으면 검증 뒤 unchanged를 출력한다. 반대 target으로의 재이전은 거절한다.

4locale 현재 페이지·template·README는 현재 실행 의미를 서술한다. 오래된 release note·SPEC·감사 근거는 history allowlist로
보존한다. 검색 결과는 분류 후보일 뿐 제거 근거가 아니며 각 live 호출·렌더 결과를 확인한다.

# CG-RETIRE·TEAMMATE 형제 계획 검토안

## Claim

각 Tier L 6종 문서를 작성했다. CG는 live CLI·config·tmux·template·4locale 문서 영역의 철거/이전이고 TEAMMATE는
CLI·tmux·gateway 인증/수명·hook의 새 pane 신뢰 경계이므로 Tier L이다. 분량을 줄이려고 실제 범위를 Tier M으로 낮추지 않았다.

## Evidence 및 Baseline-attribution

지정 WT에서 `git rev-parse --short HEAD` 출력은 `81c1d58f9`였다. cg.go root 등록, applyCGMode의 설정/tmux 주입,
TeamModeCG 정의, spawn.go new-window, session.go split-window를 sed/rg로 읽었다. 각 research.md에 경로를 적었다.
설치 tmux 3.6a man의 new-window/split-window -e와 명령 인수 문법을 읽었으며 tmux/Claude를 실행하지 않았다.
SPEC ID Bash regex는 두 ID 모두 PASS였다. 역사적인 문서 스윕 개수를 현재 측정값으로 복제하지 않았다.

## 결정과 미결

CG: 제거된 명령을 다른 provider로 자동 alias하지 않는다. team_mode cg는 명시 이전 전 실행 거절하며 원본·모르는 키·
credential 참조를 보존한다. 역사 자료는 고치지 않는다. 동등 CG 팀 구성을 유지하려면 TEAMMATE 통합 게이트가 필요하다.

TEAMMATE: 공유 tmux env를 수정하지 않는 pane bootstrap을 우선 후보로 한다. tmux 문법의 존재는 Claude native teammate
연결 증거가 아니다. 지원 seam 실측이 필수다. seam이 없다면 명시 MoAI pane 작업 또는 lead별 tmux server는 대안이며
동등성/UX 변화 결정 전 채택하지 않는다. 코어의 in-process 제한을 pane 격리 완료로 세지 않는다.

## Gaps와 잔여 위험

실제 teammate seam·scope별 env·parent 강제 종료·same-session 공존·모든 목표 model/auth는 미실행이다. CG 현재 참조 전체
분류·4locale 빌드·명시 이전의 실제 정책 UI 역시 run에서 확인해야 한다. 이 초안은 구현·카드 생성·원격 착지가 아니다.


## 직접 실행한 구조 검사

```text
$ /tmp/moai-gateway-81c1d58f9 spec lint SPEC-MOAI-CG-RETIRE-001
✓ No findings — all SPEC documents are valid
$ /tmp/moai-gateway-81c1d58f9 spec lint SPEC-MOAI-GATEWAY-TEAMMATE-001
✓ No findings — all SPEC documents are valid
```

각 exit 0. 부모가 같은 HEAD에서 빌드한 기존 바이너리를 사용했고 이번 작성자는 재빌드하지 않았다. 이 출력은 구조 판정이며
semantic 감사·live 제거·pane 보안·사용자 기능 검증의 PASS가 아니다.


## SP-B1 한정 보강

CG 0.1.0의 design 표와 AC-CR-002/003에 migrate cg preview/apply, claude-only의 역할 변화 수락,
claude-glm의 실제 TEAMMATE 게이트, 정확한 YAML 경로·값, 첫 launch guard·새 reader, 원본 backup과 before/after fixture를
고정했다. 새 reader는 아직 구현되지 않았다. TEAMMATE 및 코어 문서는 수정하지 않았다.

보고만 하는 코어 경계: design §7.1은 Backend를 launcher 초기 provider로 정의하지만 §7.3은 FactoryRunStart의
command backend를 그대로 둔다. 이 작성 작업은 그 차이를 해결하거나 factory provider attribution이 완료됐다고 주장하지 않는다.

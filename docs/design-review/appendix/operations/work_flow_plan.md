# `work-flow-plan`

operation `work_flow_plan` · 코드 `src/ops/work_flow_plan.rs`

`work-flow-plan --project <폴더>` 는 M-Flow 첫 설정의 `work_flow` 질문 근거를 읽기 전용으로 내보낸다(SPEC-20261120-mflow-integration-branch-choice). 후보 가지마다 `canonical` 플래그를 붙이는데, 원격의 풀 요청 후보 이름공간(`refs/remotes/origin/pr/<n>` 따위)은 `canonical: false` 와 candidate 라벨이 붙어 보이기만 하고 통합 대상으로 추천되지 않는다. 추천은 값 하나: 프로필에 `work_flow.integration_branch` 가 기록돼 있으면 그 값(이유 문구가 "recorded" 라고 말한다 — 다시 실행해도 저장된 답을 제안한다), 없으면 `origin/HEAD` 가 가리키는 원격 기본 가지, 그것도 없으면 현재 가지. 기록된 대상이 저장소에 없으면 `missing-target` 상태와 안전한 선택지 문구를, 다른 워크트리가 이미 그 가지를 체크아웃 중이면 `occupied-worktree` 상태와 그 경로를 내보낸다. 가지·워크트리·파일·ref 를 하나도 만들거나 옮기거나 지우지 않고, 네트워크 명령도 돌리지 않는다. 카드 본문이 요구한 "값 하나로 저장" 의 두 번째 값(원격 최종 대상)과 묶음 `--tip` 의 프로필 기본은 이 카드의 뒤 조각(다른 카드의 예약이 풀린 뒤)이다.

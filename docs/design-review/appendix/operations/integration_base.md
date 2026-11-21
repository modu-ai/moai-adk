# `integration-base`

operation `integration.base` · 코드 `src/ops/integration_base.rs`(해석은 `src/spec_take.rs` 의 `integration_base`·`worktree_base`) · SPEC-20261121-integration-base-sha (조각 S1)

`integration-base <project> [--worktree PATH]` 는 M-Flow 레인이 출발하는 **로컬 통합 브랜치의 정확한 커밋**을 알려 주는 읽기 전용 명령이다. 아무것도 쓰지 않는다(`writes:false`).

**읽는 곳.** 프로젝트 프로필의 `work_flow` 하나다. 흐름이 `m-flow` 이고 `integration_branch` 가 있으면 `refs/heads/<branch>` 가 가리키는 40자리 커밋을 `sha` 로, 출처를 `source:"local"` 로 낸다. 해석은 `refs/heads/` 아래에서만 하므로 같은 이름의 태그나 원격 추적 ref(`origin/<branch>`)가 대신 잡히지 않는다. 그래서 로컬과 원격이 갈라져도 항상 로컬 커밋이다.

**로컬 브랜치가 없을 때.** 원격 추적 ref 를 읽지 않고 멈춘다(`not_found`). 오류 글에 없는 브랜치 이름, "origin 에서 읽지 않았다"는 말, 다음 단계(`git branch <branch> <commit>` 로 직접 만들거나 프로필의 `integration_branch` 를 바꾼 뒤 다시 실행)가 들어 있고 커밋 id 는 출력하지 않는다.

**정책이 없을 때.** 프로필을 읽을 수 없거나(`profile_unreadable`), `work_flow` 가 없거나(`no_flow`), 흐름이 m-flow 가 아니거나(`flow_not_m_flow`), 통합 브랜치가 비어 있으면(`no_integration_branch`) `policy:"none"` 과 그 이유를 내고 종료 코드 0 이다. 그런 프로젝트의 기준은 이전처럼 사람이 정한다.

**`--worktree PATH`.** 그 워크트리의 `head`, 브랜치 이름, 로컬 통합 커밋이 `head` 의 조상(또는 같음)인지(`contains_base`), 그 브랜치에 기록된 커밋(`branch.<name>.moai-base-sha`, 없으면 `recorded_sha:null`), 기록이 로컬 통합 커밋과 같은지(`matches_recorded`)를 `worktree` 에 낸다. 커밋이 없는 경로는 `invalid_input` 으로 거절한다. 정책이 없으면 `worktree` 는 `null` 이다.

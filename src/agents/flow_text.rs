//! The text of the flow rule file `.claude/rules/moai/flow.md`, one per work flow the profile can
//! name (`profile::WORK_FLOWS`). Each text states the flow's rules as short facts, never longer
//! than [`MAX_LINES`] lines and [`MAX_BYTES`] bytes, and has no frontmatter, so no `paths` key
//! exists and Claude Code loads the file at session start like `CLAUDE.md`.

/// The most lines a flow text holds (REQ-10 of the layout SPEC).
pub const MAX_LINES: usize = 80;
/// The most bytes a flow text holds.
pub const MAX_BYTES: usize = 5 * 1024;

/// The flow text for `flow` (`m-flow`, `github-flow`, `trunk-based` or `git-flow`) in English or
/// Korean, with the integration branch where the flow has one; `None` for any other value.
pub fn render(flow: &str, integration_branch: Option<&str>, ko: bool) -> Option<String> {
    let branch = integration_branch.map(|b| format!(" `{b}`")).unwrap_or_default();
    let body: String = match (flow, ko) {
        ("m-flow", false) => format!(
            "# Work flow: m-flow\n\n\
MoAI's work flow for several AI agents that work at the same time on one project.\n\n\
- One integration branch{branch}. Every change starts from its HEAD, never from the remote default branch when that differs (Claude Code worktrees need `worktree.baseRef: head`).\n\
- One SPEC, one worktree, one branch `spec/<slug>-<yyyymmdd>` and one pull request.\n\
- Write the tests first. When the work is done, push the branch and open the pull request.\n\
- An agent other than the author audits the exact head.\n\
- The leader merges after CI passes and the audit says PASS.\n\
- No card number goes into a branch name, a commit or a pull request.\n\
- The local cards stay each person's own to-do list.\n"
        ),
        ("m-flow", true) => format!(
            "# 작업 흐름: m-flow\n\n\
AI 에이전트 여럿이 한 프로젝트에서 동시에 일할 때 쓰는 MoAI의 작업 흐름이에요.\n\n\
- 통합 브랜치는 하나예요{branch}. 모든 변경은 그 HEAD에서 시작하고, 원격 기본 브랜치와 다르면 기본 브랜치에서 시작하지 않아요(Claude Code 워크트리는 `worktree.baseRef: head` 가 필요해요).\n\
- SPEC 하나에 워크트리 하나, 브랜치 `spec/<slug>-<yyyymmdd>` 하나, 풀 리퀘스트 하나예요.\n\
- 테스트를 먼저 써요. 일이 끝나면 브랜치를 푸시하고 풀 리퀘스트를 열어요.\n\
- 작성자가 아닌 에이전트가 정확한 head를 감사해요.\n\
- 리더는 CI가 통과하고 감사가 PASS일 때 병합해요.\n\
- 카드 번호는 브랜치 이름, 커밋, 풀 리퀘스트에 넣지 않아요.\n\
- 로컬 카드는 각자의 할 일 목록으로 남아요.\n"
        ),
        ("github-flow", false) => format!(
            "# Work flow: github-flow\n\n\
- Short branches start from the default branch{branch}.\n\
- One pull request per change; merge after review and CI.\n\
- The default branch is always deployable.\n"
        ),
        ("github-flow", true) => format!(
            "# 작업 흐름: github-flow\n\n\
- 짧은 브랜치는 기본 브랜치{branch}에서 시작해요.\n\
- 변경 하나에 풀 리퀘스트 하나이고, 리뷰와 CI 뒤에 병합해요.\n\
- 기본 브랜치는 언제나 배포할 수 있는 상태예요.\n"
        ),
        ("trunk-based", false) => format!(
            "# Work flow: trunk-based\n\n\
- Everyone commits to one trunk{branch} in small steps, often directly or through very short branches.\n\
- Unfinished work hides behind flags.\n\
- The trunk stays green after every commit.\n"
        ),
        ("trunk-based", true) => format!(
            "# 작업 흐름: trunk-based\n\n\
- 모두가 하나의 트렁크{branch}에 작은 단위로 커밋해요. 직접 커밋하거나 아주 짧은 브랜치를 써요.\n\
- 끝나지 않은 일은 플래그 뒤에 숨겨요.\n\
- 커밋마다 트렁크는 초록 상태를 지켜요.\n"
        ),
        ("git-flow", false) => String::from(
            "# Work flow: git-flow\n\n\
- `develop` is the integration branch; `feature/*` branches start from it.\n\
- `release/*` and `hotfix/*` branches carry a release and a fix.\n\
- `main` holds only released versions.\n"
        ),
        ("git-flow", true) => String::from(
            "# 작업 흐름: git-flow\n\n\
- `develop` 이 통합 브랜치이고, `feature/*` 브랜치는 거기서 시작해요.\n\
- `release/*` 와 `hotfix/*` 브랜치가 릴리스와 수정을 맡아요.\n\
- `main` 에는 릴리스한 버전만 있어요.\n"
        ),
        _ => return None,
    };
    debug_assert!(body.lines().count() <= MAX_LINES && body.len() <= MAX_BYTES);
    Some(body)
}

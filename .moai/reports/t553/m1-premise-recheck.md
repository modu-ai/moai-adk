# t553 M1 — 전제 재현 시도의 결과: 전제가 스테일하다

측정 트리: `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t553`
측정 SHA: `3ac58b5a1` (= `origin/develop`), 브랜치 `WT-mcp-project-root`
측정 일자: 2026-09-08

## Claim

카드 t553 이 인용한 GH #1640 의 전제 — "문서화된 `project_root` 파라미터가
실구현되지 않았다" — 는 `origin/develop` 및 `origin/main` 에서 **더 이상 참이
아니다**. 구현·문서·테스트가 모두 존재하며 서로 일치한다.

동시에 **제보자는 제보 시점에 옳았다**: 수리는 어떤 릴리스 태그에도 실려
있지 않다.

## Evidence

### E1 — 구현 실재

```
$ /usr/bin/grep -rn 'project_root' --include='*.go' internal/ | /usr/bin/grep -v '_test\.go:'
internal/cli/mcp_project_root.go:34:const projectRootArg = "project_root"
internal/cli/mcp_project_root.go:83:// A non-empty project_root that cannot be a project root is REJECTED rather than
internal/cli/mcp_project_root.go:181:  return "", fmt.Errorf("project_root %q cannot be canonicalized: %w", raw, err)
internal/cli/mcp_project_root.go:187:  return "", fmt.Errorf("project_root %q has no .moai directory, so it is not a MoAI project root", raw)
(외 다수)
```

카드가 [HARD] 로 요구한 두 성질이 모두 구현돼 있다 — canonicalize
(`mcp_project_root.go:181`, symlink 해소) 와 조용한 폴백 대신 거부
(`:160`/`:163`/`:187`).

### E2 — 도구 커버리지 12/28, 문서와 정확히 일치

선언 지점 12개를 도구 이름에 매핑한 결과:

| 도구 | 의미론 |
|---|---|
| spec_progress, verify_snapshot, verify_trend, spec_audit, spec_drift, codex_audit, glm_audit, graph_file_api, graph_find_code, graph_trace_calls, graph_shortest_path | fallback |
| audit_multi | passthrough |

이 트리의 `.claude/rules/moai/core/moai-mcp-tools.md:22` 는 **"Twelve tools"**
로 시작하며 위 12개를 정확히 열거한다. 즉 문서↔코드 불일치는 **없다**.

주의 — primary 체크아웃의 같은 파일은 `Nine tools` 라고 적혀 있다
(`/Users/goos/MoAI/moai-adk-go/.claude/rules/moai/core/moai-mcp-tools.md:22`).
이는 결함이 아니라 **트리 차이**다(main 이 develop 보다 뒤). 세션에
로드된 CLAUDE.md 문맥은 primary 사본이므로, 그것만 읽으면 "문서는 9개인데
코드는 12개" 라는 존재하지 않는 드리프트를 보고하게 된다.

### E3 — 테스트 초록

```
$ go test ./internal/cli/ -run 'ProjectRoot|project_root|Worktree' -count=1
ok  github.com/modu-ai/moai-adk/internal/cli  16.218s
```

### E4 — 수리 이력

```
$ git log --reverse --format='%h %ad %s' --date=short -- internal/cli/mcp_project_root.go
1f9deed0c 2026-08-23 feat(SPEC-MCP-WORKTREE-ROOT-001): M1 project_root on the three SPEC tools
2c0efade0 2026-08-23 feat(SPEC-MCP-WORKTREE-ROOT-001): M2 project_root on the codex path and through audit_multi
34de07740 2026-08-23 fix(SPEC-MCP-WORKTREE-ROOT-001): M5 post-repair check, and the defect it found in the repair
db1ac0afa            fix(cli): canonicalize project_root so a symlink cannot outlive a boundary (card t183)
21734f9e9            feat(mcp): verify tools honor project_root + fallback provenance in catalog responses (t236)
```

`.moai/specs/SPEC-MCP-WORKTREE-ROOT-001/spec.md` frontmatter: `status: completed`.

### E5 — 착지 범위

```
$ git merge-base --is-ancestor 1f9deed0c origin/main  → 0 (포함)
$ git merge-base --is-ancestor 21734f9e9 origin/main  → 1 (미포함, develop 전용)
$ git merge-base --is-ancestor 1f9deed0c v3.1.2       → 1 (미포함)
$ git tag --list 'v3*' --sort=-v:refname | head -1    → v3.1.2
```

### E6 — 카드 문안의 변수명 부정확

카드는 "스테일 `MOAI_PROJECT_DIR` 에 고정된다" 고 적었으나, 폴백이 읽는
변수는 `CLAUDE_PROJECT_DIR` 이다 (`internal/cli/session.go`
`resolveProjectDirWithSource`). `MOAI_PROJECT_DIR` 은 `internal/hook/cwd_changed.go:83`
이 유일하게 **생산**하며, 같은 파일 `:70` 주석이 "No Go code consumes
MOAI_PROJECT_DIR yet (verified 2026-09-02)" 라고 적는다. 실질(스테일 env
폴백)은 옳고 변수명만 틀렸다.

## Baseline-attribution

위 모든 측정은 이 트리(`.claude/worktrees/t553`)에서 HEAD `3ac58b5a1` 에
대해 이번 실행에서 수행했다. 다른 트리·다른 시점의 수치를 옮겨 적지
않았다. `grep` 은 `/usr/bin/grep` 절대경로로 재측정했다 — 셸의 `grep` 은
ugrep 래퍼로 프로덕션 `.go` 파일을 조용히 건너뛰었고, 그 판독만 믿었다면
"테스트에만 존재, 구현 없음" 이라는 거짓 확인이 나왔을 것이다.

## Gaps (관측하지 않은 것)

- 제보자 환경의 실제 서버 빌드 버전을 확인하지 않았다(#1640 본문 미판독 —
  `gh issue view 1640` 는 메타데이터만 조회했다).
- 미선언 16개 도구(goal_*, session_*, codex_job_*, glm_job_*, audit_cache,
  codex_setup, codex_task, glm_task)가 `project_root` 를 **가져야 하는지**
  는 판정하지 않았다. 독트린은 12개만 명시하므로 현재로선 의도된 범위로
  읽히지만, 그 의도를 SPEC 본문에서 확인하지 않았다.
- 런타임 재현(실제 워크트리에서 MCP 호출을 날려 primary 를 읽는지)은
  하지 않았다 — 실행 중인 서버는 `91d25bc61` 빌드이며 이 트리가 아니다.

## Residual-risk

- 미선언 16개 중 트리 의존적인 것이 있다면(예: `goal_arm` 의 상태 파일
  경로) 같은 결함이 그 표면에 남아 있을 수 있다. 위 Gaps 2번이 그 판정을
  덮지 않는다.
- primary/develop 문서 차이는 무해하지만, 다음 세션이 primary 사본만 읽고
  드리프트를 오진할 재발 가능성이 있다.

## 부수 관측 (이 카드와 별건)

세션 프롬프트가 지정한 스크래치패드 경로
(`/private/tmp/claude-501/<project-key>/<session>/scratchpad`)에 쓰려 하자
`Path traversal detected: file is outside project directory` 로 차단됐다.
스크래치패드는 정의상 프로젝트 밖에 있으므로, 「프로젝트 디렉터리 밖 쓰기
금지」 가드와 「임시 파일은 스크래치패드에」 지침이 서로 부딪힌다. 이번에는
증거 디렉터리로 대체해 결과가 오히려 나았지만(게시문이 추적 경로에 남았다),
가드가 지시된 경로를 막는 형태 자체는 별도 판정 대상이다.

## 게시 결과 (2026-09-08)

GH #1640 에 게시했다 — https://github.com/modu-ai/moai-adk/issues/1640#issuecomment-5579350783
게시 본문은 같은 디렉터리 `gh1640-comment-posted.md` 가 그대로 보관한다.
재판독 검증: 댓글 수 1, author `GoosLab`, `2026-09-08T04:41:49Z`, 본문 4613자,
이슈 state `OPEN` 유지(릴리스 전이므로 닫지 않았다).

게시 전에 #1640 본문을 실제로 읽은 결과, 위 Gaps 1번이 닫히면서 **결함이
둘**임이 드러났다 — M1 이 측정한 것은 결함 2(catalog tools 의 project_root)
뿐이고, 본문의 결함 1(worktree 이동 시 env 스탬프 스테일)은 별도 축이다.
결함 1 의 수리는 `f1b379434` (2026-09-02, t236) 이며 `origin/main` 에
미포함(develop 전용)이다. 또한 `git show a1b1ca696:internal/cli/session.go`
`:243` 실측으로, 제보 시점에도 폴백이 읽던 변수는 `CLAUDE_PROJECT_DIR`
이었음을 확인했다.

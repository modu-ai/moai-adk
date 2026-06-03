# Memory Directory Resolution Doctrine (Gap C)

> SessionEnd handoff 영속화(persistence)의 메모리 디렉터리 해석이 **cwd별(per-cwd)로 동작하는 것은 의도된 설계**임을 명문화한다. git-root 정규화는 회귀(regression)를 유발하므로 채택하지 않는다.
>
> 코드 위치: `internal/hook/session_end.go` `resolveMemoryDir` / `projectSlug`, `internal/hook/handoff/persist.go` warn 분기.

---

## 1. 핵심 원칙 — per-cwd by design

`resolveMemoryDir(homeDir, projectDir)`는 세션의 raw `CWD`를 `projectSlug()`로 슬러그화하여 `~/.claude/projects/{slug}/memory/` 경로를 해석한다. **git-root 정규화를 수행하지 않는다.** 따라서:

- main repo 세션과 git-worktree 세션은 **서로 다른 메모리 디렉터리**를 해석한다.
- 이는 결함이 아니라 Claude Code 네이티브 메모리 모델과의 정합이다.

근거 (실증):

- `~/.claude/projects/` 디렉터리는 cwd별 해시 디렉터리를 보유한다 — Claude Code 자체가 per-cwd로 동작한다.
- MoAI의 SessionEnd handoff 쓰기를 git-root로 강제 정규화하면, MoAI 쓰기(main-root)가 Claude Code의 네이티브 auto-load(cwd 기준)와 **desync**된다. 이는 "고침"이 아니라 회귀다.
- `CLAUDE_PROJECT_DIR` 기반 해석으로 바꾸는 대안도 단독으로는 divergence를 해소하지 못함이 검증되었다.

## 2. 변경 금지 (no-logic-change 계약)

다음은 **변경 금지** 대상이다 (문서/주석만 추가, 동작 불변):

- `resolveMemoryDir` 의 경로 해석 로직 (homeDir/projectDir 검증 + `filepath.Abs` + `projectSlug` join)
- `projectSlug` 의 슬러그 인코딩 규칙 (`/`, `\`, `.` → `-`)
- 위 동작을 인코딩한 테스트 `internal/hook/resolve_memory_dir_test.go` (assertion 불변)

이 doctrine은 미래 메인테이너가 "per-cwd divergence는 버그다"라고 오인하여 git-root 정규화를 도입하는 것을 방지한다.

## 3. L3 `--worktree` resume와의 관계

per-cwd 동작 때문에, L3 `/moai plan --worktree`로 초기화한 SPEC을 이어받을 때는 **반드시 worktree cwd 안에서** 새 세션을 시작해야 메모리 디렉터리가 일관되게 해석된다.

- paste-ready resume의 **Block 0 (cwd anchoring)** 가 이 재앵커링을 담당한다.
- 상세: `.claude/rules/moai/workflow/session-handoff.md` § Worktree-Anchored Resume Pattern.

worktree cwd에서 메모리 디렉터리가 아직 생성되지 않았다면, `persist.go`의 `os.Stat(memoryDir)` 미스 분기가 발동하여 영속화를 건너뛴다 (정상 동작, 에러 아님). warn 메시지는 이 per-cwd divergence 가능성을 힌트로 제공한다.

## 4. Cross-References

- `internal/hook/session_end.go` — `resolveMemoryDir` (per-cwd 의도 @MX:NOTE 보유) + `projectSlug`
- `internal/hook/handoff/persist.go` — memory-dir 미존재 warn 분기 (worktree divergence 힌트)
- `internal/hook/resolve_memory_dir_test.go` — per-cwd 동작 인코딩 테스트 (불변)
- `.claude/rules/moai/workflow/session-handoff.md` § Worktree-Anchored Resume Pattern — Block 0 재앵커링
- `.claude/rules/moai/workflow/moai-memory.md` — Claude Code 네이티브 auto-memory feature (저장 위치 = git repo root 기준, but Claude Code 자체는 per-cwd 해시 디렉터리 사용)

---

Status: Active — Gap C 의사결정 기록 (document-only, 동작 불변)

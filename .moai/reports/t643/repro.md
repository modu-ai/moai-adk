# t643 재현 — F1·F7 (정정 전)

- 카드: t643 (Class B, 문서 정정, plan 생략). 근거: t610 sync 감사 `.moai/reports/t610/sync-audit.md`의 F1(should-fix)·F7(advisory)
- 워크트리: `.claude/worktrees/t643`, 브랜치 `WT-toolchain-doc-fixes`
- 기반 정렬: `origin/develop`(`d060e0d13`)에서 워크트리를 만들었다 → `git status --short` 출력 없음 → `git merge --no-ff 296ba7aa5` 실행.
  - 기반 정렬 병합 커밋: `8addcf742`
  - `git rev-parse HEAD HEAD^1 HEAD^2` → `8addcf7429e57de988037528901826e2fd9b14cc` / `d060e0d136f173f07353e46a709382d08ae25dd2` / `296ba7aa5682da20ef906dde6b2764431442b348`
- 측정 트리: HEAD `8addcf742` (정정 전)

## F1 — codemaps 머리말의 측정 귀속

감사 원문(`sync-audit.md` F1 행): 머리말은 모든 수치를 t592 `e7bd89ee3` 트리에서 쟀다고 단언한다. 그런데 같은 머리말 블록의 `**Go**: 1.26.8`은 그 트리에서 잰 값이 아니다.

`.moai/project/codemaps/overview.md` 3-8행 (정정 전):

```text
> `/moai codemaps`로 생성된 아키텍처 지도입니다. 모든 수치는 아래 트리에서 직접 잰 것이고,
> 다른 트리·다른 시점에서 옮겨온 값은 없습니다.

**모듈**: `github.com/modu-ai/moai-adk` · **Go**: 1.26.8
**측정 트리**: worktree `.claude/worktrees/t592`, 브랜치 `WT-home-state-rollout`, HEAD `e7bd89ee3`
```

`.moai/project/codemaps/modules.md` 6-7행 (정정 전): 같은 `**Go**: 1.26.8` 줄과 같은 `**측정 트리**` 줄이 있다.

측정 트리의 실제 값:

```text
$ git show e7bd89ee3:go.mod
module github.com/modu-ai/moai-adk

go 1.26.4
...
```

재현: 머리말이 지목한 트리의 `go.mod:3`은 `go 1.26.4`이고, 문서는 `1.26.8`을 적고 있다. 이 값은 t610 커밋 `41f445fa5`가 `go.mod:3`을 바꾼 뒤의 값이다. 값 자체는 현재 트리 기준으로 맞지만, 머리말의 귀속과는 어긋난다. **재현됨.**

## F7 — CHANGELOG 도달성 문구

감사 원문(`sync-audit.md` F7 행): "reachable from the `go.mod` build directive"는 부정확하다. govulncheck의 도달성은 코드 호출 경로를 기준으로 판정하지, go 지시어를 기준으로 하지 않는다.

```text
$ grep -n -o 'reachable from the[^;]*;' CHANGELOG.md
12:reachable from the `go.mod` build directive at go1.26.4 — GO-2026-6218, ..., GO-2026-5026 — once the `go` directive on `go.mod:3` is raised to `go 1.26.8`;
```

govulncheck가 도달성을 무엇으로 판정하는지는 기준선 로그의 `Example traces found` 블록으로 확인한다. 이 블록에는 모듈 코드에서 취약 심벌까지의 호출 경로가 적혀 있다. 명령과 출력은 `repro-traces.txt`에 있다. **재현됨.**

## 형제 문구 확인

같은 문구가 다른 문서에 있는지는 보고서 디렉터리를 뺀 `*.md`에서 찾았다. 명령과 결과는 `repro-siblings.txt`에 있다.

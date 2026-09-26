# t1219 판정서 — 세션 컨텍스트 이중 로드와 CLAUDE.local.md 모순

- 카드: t1219 (Tier M · 클래스 C · plan)
- 트리: `.claude/worktrees/t1219`, 브랜치 `WT-local-doc-dedup`, 기준 develop `e464fd5d0`
- 측정 환경: Claude Code 2.1.283, 2026-09-26

## (1)(2) 이중 로드 — 원인 귀속

### Claim
이중 로드는 워크트리 세션 일반의 성질이 아니다. **워크트리 안에서 시작한 세션**은 지침 파일과 프로젝트 스킬 디렉터리를 워크트리 사본 한 벌만 읽는다. 이중 로드는 **primary 체크아웃에서 시작한 뒤 `EnterWorktree` 로 옮겨 온 세션**에서 생긴다. 시작 시 primary 사본을 싣고, 이동 뒤 워크트리 안의 파일을 읽을 때 워크트리 사본(중첩 메모리·rules)을 추가로 싣는다.

### Evidence
1. 워크트리 cwd 에서 띄운 headless 세션(`claude -p "reply ok" --max-turns 1 --debug-file …`, exit 0):
   - 프로젝트 스킬 경로가 하나뿐이다 — `probe-skills-dirs.txt`:
     `project=[/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t1219/.claude/skills]`
   - 로드된 CLAUDE.local.md 는 워크트리 사본 하나다. InstructionsLoaded 훅이 보고한 경로가 워크트리 사본뿐이다 — `probe-memory-files.txt`. primary 사본 경로는 권한 규칙 문자열(`Read(`/`Edit(`)에만 나타난다.
2. 이 세션(primary 에서 시작 → `EnterWorktree`)에서 워크트리 안의 `CLAUDE.local.md` 를 Read 하자, 워크트리 사본 `.claude/rules/local/gitflow-lane-protocol.md` 전문이 컨텍스트에 새로 첨부됐다. primary 사본은 세션 시작 때 이미 실려 있었다.
3. 크기: primary 사본 52,280 B, 워크트리(develop) 사본 62,301 B(`wc -c`). 두 사본은 내용이 다르다(CLAUDE.local.md §0 이 설명하는 의도된 분기).

### Baseline-attribution
위 명령과 출력은 이 실행에서, 이 트리(`e464fd5d0`)를 대상으로 얻었다.

### Gaps
- 두 번째 경로(primary 시작 → EnterWorktree)의 **스킬 목록 이중 나열**은 이 세션 컨텍스트에서 직접 세지 못했다. 메모리·rules 는 관측했고, 스킬 목록은 추론이다.
- 이중 로드의 토큰 수(카드의 약 30k·17~20k)는 재측정하지 않았다.
- `moai cc -w <name>` 런처 경로는 "워크트리 안에서 시작" 쪽이라 한 벌만 로드할 것으로 보이나, 런처 자체로는 재지 않았다.

### Residual-risk
headless `-p` 모드와 대화형 모드가 중첩 메모리를 싣는 시점이 다를 수 있다.

### 해법 방향(SPEC 에서 확정)
레인은 `EnterWorktree` 대신 **런처로 워크트리 안에서 시작**하도록 배차 절차를 바꾸는 것이 1순위 후보다. 대안은 세션 중 이동이 불가피할 때 `/clear` 로 끊는 것이다.

## (3) CLAUDE.local.md 행동 모순 — 수정 완료

| 위치 | 이전 | 이후 |
|---|---|---|
| §1 Development Cycle | `Git commit from local root` | 카드 워크트리에서 커밋, primary 금지(AGENTS.md §2·§3) |
| §6 첫 [HARD] 줄 | §4 명령·근거 재서술(중복) | §4 Before Commit 을 단일 출처로 가리킴 |
| §6 `-count=1`·`-race`·`vet` | `./...` 전체 | `./internal/<pkg>/...` 변경 패키지 |
| §13 Unit tests | `go test ./...` | 변경 패키지만, 전체 스위트 금지(§4) |

검증: `grep -nE 'go (test|vet)( -[a-z=0-9]+)* \./\.\.\.|local root' CLAUDE.local.md` → 265행 1건만 남는다. §4 의 금지 문장 자체이며, 이 적중이 검색이 제대로 걸린다는 양성 대조다.

## (4) 제안 — kanban-dispatch.md 리드 전용 주입

`kanban-dispatch.md`(상시 로드)는 첫 줄에서 리드가 아닌 세션에서는 "inert" 라고 스스로 밝힌다. SessionStart 훅이 이미 Kanban 역할을 알고 있으므로, 이 파일을 `paths:` 스코프로 내리고 리드 역할에서만 SessionStart additionalContext 로 본문(또는 stub)을 주입하는 방식을 제안한다. HARD 조항이 리드에게 빠짐없이 닿는지 기계적으로 검사하는 장치가 선행돼야 한다. t1175·t1226(상시 로드 다이어트)과 같은 파일군이므로 그 착지 뒤 별도 카드로 다룬다.

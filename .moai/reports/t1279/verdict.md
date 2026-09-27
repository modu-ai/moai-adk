# t1279 판정서 — 세션 지침 이중 로드의 원인과 해법 비교

- 카드: t1279 (t1219 후속, plan)
- 트리: `.claude/worktrees/t1279`, 브랜치 `WT-session-double-load`, 기준 develop `b59a5d69c`
- 측정 환경: Claude Code 2.1.283, 2026-09-27
- 판정 원천: 모든 로드 판정은 **세션 전사**(`Contents of <경로>` 줄)로 했다. 디버그 로그는 쓰지 않았다(t1219 정정의 교훈).
- 측정 상한(착수 전 선언): 프로브 6회, 각 1턴, `--model haiku`. 실제 사용 5회.
- 격리: 모든 프로브는 세션 scratchpad 안의 임시 저장소에서 돌렸다. 이 저장소의 primary 체크아웃이나 t1279 트리에서는 프로브를 띄우지 않았다. `MOAI_KANBAN*` 변수는 빈 값으로 덮어썼다.
- 원시 증거: `probes-transcript.txt` (세션별 로드 파일 목록 + 첫 턴 usage)

## ① 원인 귀속

### Claim 1 — 상위 디렉터리 탐색이 원인이다
세션은 시작 디렉터리에서 위로 올라가며 `CLAUDE.md`·`CLAUDE.local.md` 를 모두 싣는다. git 저장소 경계에서 멈추지 않는다.

- 프로브 1(`de2dbb1a`): `outer/`(git 저장소) 안에 `outer/inner/`(별개 git 저장소)를 두고 inner 에서 시작했다. outer 와 inner 의 두 파일이 모두 로드됐다(4건).
- 프로브 2(`45a671b6`): outer `CLAUDE.md` 를 inner 와 바이트 동일하게 만들었다(`cmp` 동일). 여전히 4건이 로드됐다. 따라서 **내용 기반 중복 제거는 없다.**

### Claim 2 — git 워크트리에서는 main 의 CLAUDE.md 만 빠지고 CLAUDE.local.md 는 실린다
- 프로브 3(`f592469f`): main 저장소 안에 `git worktree add .claude/worktrees/w1` 로 워크트리를 만들고 w1 에서 시작했다. 로드된 것은 `w1/CLAUDE.local.md`, `w1/CLAUDE.md`, `main/CLAUDE.local.md` 이고, **`main/CLAUDE.md` 는 없다.** t1219 실측(`79ed0553`)과 같은 모양이다.
- 왜 CLAUDE.md 만 빠지는지는 관측하지 않았다. 워크트리가 같은 저장소의 추적 파일을 이미 갖고 있어 건너뛰는 것으로 보이지만, 이는 추정이다.

### Claim 3 — main 밖에 둔 워크트리는 main 의 파일을 싣지 않는다
- 프로브 4(`e8a442d6`): 같은 main 저장소의 워크트리를 main 폴더 **밖**(`../outside-w2`)에 만들고 시작했다. 로드된 것은 `outside-w2/CLAUDE.local.md`, `outside-w2/CLAUDE.md` 뿐이다. main 의 파일은 0건이다.
- 양성 대조: 같은 main 을 대상으로 한 프로브 3 이 main 파일을 로드했으므로, 이 0건은 측정 실패가 아니라 부재다.

### Claim 4 — 비용은 세션당 약 20.9k 입력 토큰이다
- 프로브 5(`27c54282`): 프로브 3 과 같은 구조에서 main `CLAUDE.local.md` 만 실제 primary 사본(52,280 B)으로 바꿨다.
- 첫 턴 입력 합계(input + cache_creation + cache_read): 프로브 3 = 41,354, 프로브 5 = 62,209. 차이는 **20,855 토큰**이다(마커 한 줄 파일과 실제 사본의 차이).

### (2) 스킬 목록
- t1219 Evidence 1: 워크트리에서 시작한 세션의 프로젝트 스킬 경로는 하나였다(디버그 로그 판독 — 이 항목은 전사 판정이 아님).
- t1219 Evidence 4: primary 에서 시작해 `EnterWorktree` 로 옮긴 세션에서는 워크트리 범위 스킬 약 40건이 추가 첨부됐다(대화 컨텍스트 직접 관측).
- 즉 스킬 목록 중복은 **세션 도중 이동** 경로의 문제이고, CLAUDE.local.md 중복(상위 탐색)과는 원인이 다르다.

### Gaps
- 스킬 목록 중복의 토큰 수는 재지 못했다. SKILL.md 설명이 여러 줄 블록이라 frontmatter 첫 줄 합산(2,406 B, 45개)은 실제 목록 크기를 과소평가한다. 이 수치는 근거로 쓰지 않는다.
- 워크트리에서 시작한 세션의 스킬 경로 단일성은 디버그 로그로만 봤다. 전사 기반 재확인이 필요하다.
- 서브에이전트가 부모의 지침을 물려받는 경로(t1219 감사 관측)는 이번에 재지 않았다.
- `main/CLAUDE.md` 가 빠지는 규칙의 정확한 조건은 미관측이다.

### Residual-risk
- 프로브는 haiku·headless·1턴이다. 대화형 세션이나 다른 모델에서 로드 규칙이 다를 가능성은 배제하지 못했다.
- Claude Code 버전이 바뀌면 탐색 규칙이 달라질 수 있다.

## ② 해법 비교

| 해법 | (1) CLAUDE.local.md 중복 | (2) 스킬 목록 중복 | 비용·위험 |
|---|---|---|---|
| **A. L2 워크트리 이전**(카드 워크트리를 primary 폴더 밖 `~/.moai/worktrees/` 로) | 해소(Claim 3) | 해소 안 됨 — 이동 경로의 문제 | 워크트리 체계·런처·폐기 절차·독트린 다수 변경. `moai worktree done` 이 L2 를 닫는 기존 경로는 있다. 사용자 표면 변경 |
| **B. primary CLAUDE.local.md 경량화** | 완화 — 남는 크기만큼 계속 중복 | 해소 안 됨 | 문서만 바꾼다. 다만 §0 은 primary 사본을 develop 판으로 유지하도록 정하고 있어, 경량화하려면 그 규칙과 먼저 맞춰야 한다 |
| **C. `--setting-sources` 조정** | 역효과 가능 — t1219 감사(N1)에서 `user,project` 는 CLAUDE.local.md 로드를 통째로 끔. 워크트리 사본까지 잃는다 | 무관 | 채택 불가로 판단 |
| **D. 세션 도중 이동 금지**(레인은 런처로 워크트리 안에서 시작, 불가피하면 이동 후 `/clear`) | 해소 안 됨(Claim 2) | 해소 가능 — t1219 Evidence 1·4 | 레인 운영 절차 변경. `/clear` 가 스킬 목록을 되돌리는지는 미측정 |

## ③ 리드 결정(2026-09-27): A + D — 사전 조건 ② 확인 결과: 블로커

리드는 A+D 를 정하면서, A 를 확정하기 전에 L2 경로에서 도구들이 동작하는지 먼저 확인하도록 했다. 코드와 도구 계약을 읽어 확인했으며, 실행 프로브는 돌리지 않았다. 확인 결과 막히는 지점이 셋이다.

| 도구 | L2 에서 | 근거 |
|---|---|---|
| **생성**(`moai worktree new`) | **불가** — 절대 경로·구분자를 거부하고 L1 leaf 이름만 받는다 | `internal/cli/worktree/new.go:50-58` `validateNewWorktreeName` |
| **WorktreeCreate 훅**(`isolation: worktree` 서브에이전트) | **L1 고정** — `.claude/worktrees/<name>` 에만 만든다 | `internal/hook/worktree_create.go:41` `agentWorktreeParentDir` |
| **EnterWorktree 전환** | **부분 불가** — 이미 워크트리 안에 있으면 전환 대상이 `.claude/worktrees/` 아래여야 한다. launch 디렉터리에서 첫 진입할 때만 `git worktree list` 의 임의 경로가 허용된다 | Claude Code EnterWorktree 도구 계약(설명문) |
| 재진입(`moai cc -w <abs>`) | 가능 — `~/.moai/worktrees/` 접두를 받는다 | `internal/cli/launcher.go:958-991` `resolveWorktreeL2Path` |
| `moai worktree done` | L2 대상. 코드 주석에 따르면 L2 흐름은 변경 없이 유지된다 | `internal/cli/worktree/done.go:186` |
| `moai integration acquire/release`, git -C 절대 경로 가드 | **미확인** | — |

**Claim:** A 를 그대로 채택하면 카드 워크트리를 만들 인가된 경로가 없다. 레인이 맨손 `git worktree add` 로 돌아가게 되는데, 이는 독트린이 금지한다. 레인 사이의 전환도 막힌다. 따라서 A 는 **생성 경로 신설(Go 변경)** 을 필수 선행으로 갖는다.

**Gaps:** 위 표는 실행 프로브가 아니라 코드·계약 판독이다. integration·가드는 확인하지 않았다.

## ④ 최종 결정(리드, 2026-09-27): A 보류 · D 만 SPEC 화 · B 는 범위 밖

**A 보류 사유:** 위 §③ 표에 정리했다. 확인한 7개 도구 가운데 불가 3곳, 동작 2가지, 미확인 2가지다.
- 불가(생성·전환 좌표 3곳): `new.go:50-58` · `worktree_create.go:41` · EnterWorktree 전환 계약.
- 동작 2가지: `moai cc -w <abs>` 재진입과 `moai worktree done`.
- 미확인 2가지: integration acquire/release 와 `git -C` 가드.
- 그 밖에, 레인이 develop 통합 워크트리로 들어가 병합하는 흐름이 깨질 위험이 있다.

**A 를 되살릴 조건(둘 다 충족해야 한다):**
1. L2 워크트리를 만드는 인가 경로가 생긴다(`moai worktree new` 의 L2 모드 또는 동등한 경로). WorktreeCreate 훅도 L2 를 지원해야 한다.
2. 레인 전환 흐름에 대안이 생긴다. 이미 워크트리 안에 있는 세션이 다른 L2 트리나 통합 워크트리로 옮기는 인가 경로가 필요하다(예: 런처 재기동 절차). Claude Code 의 EnterWorktree 제약이 풀리는 것도 이 조건을 충족한다.

**B(primary CLAUDE.local.md 경량화)는 이 카드 범위에서 뺀다.** t1243·t1259(CLAUDE.local.md → AGENTS.local.md, 4만 자 이하 이관)가 같은 문제를 다룬다. 그 작업이 착지하면 이중 로드의 비용(Claim 4, 약 20.9k 토큰)도 함께 줄어든다.

**D 범위:** 세션 도중 워크트리 이동을 금지한다(런처로 시작하고, 불가피하면 이동 뒤 `/clear`). 문서는 로컬·템플릿 양쪽에 둔다. 가드로 차단할지는 SPEC 에서 선택지로 제시한다. 스킬 목록 중복의 토큰 비용 측정을 포함한다.

**권고(이전, 참고용):** (1)은 **A** 가 유일하게 원인을 제거한다. (2)는 **D** 가 담당한다. 둘은 서로 다른 원인을 겨냥하므로 함께 쓴다.

A 는 사용자 표면(워크트리 위치, 런처 동작)을 바꾸는 선택이다. 그래서 리드에게 올려 결정을 받는다. 결정 전에 SPEC 을 쓰면 t1219 처럼 전제가 뒤집힐 수 있으므로, SPEC 작성은 결정 뒤로 미룬다.

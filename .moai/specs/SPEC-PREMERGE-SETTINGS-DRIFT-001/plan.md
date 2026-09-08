# plan.md — SPEC-PREMERGE-SETTINGS-DRIFT-001

카드 t488 · 브랜치 `WT-premerge-drift-assert` · 워크트리 `.claude/worktrees/t488`

읽은 파일에서 확인한 사실만 적는다. 코드 주장에는 경로와 줄번호를 붙였고, 줄번호는 본 워크트리 트리 기준이다.

## §A 맥락

t487(SPEC-SETTINGS-ORIGIN-001)은 dirty `.claude/settings.json`의 작성자를 귀속 불가로 종결하면서 재발 방지 권고 하나를 남겼다: 병합 창 진입 직전에 drift를 단정하고, 검출되면 보존 + 리드 보고, 자동 복구 금지. 그 권고를 구현하는 카드가 t488이다.

기존 자산이 이미 필요한 조각을 대부분 갖고 있다.

- `moai integration` 세 동사(`status` / `acquire` / `release`)가 레인의 창 진입 표면이다(`internal/cli/integration.go:167` 이하).
- `integrationLockRoot()`(`internal/cli/integration.go:46`)는 **워크트리에서 호출해도 primary 체크아웃 루트를 돌려주는** 이미 검증된 해석기다(`CLAUDE_PROJECT_DIR` 우선, 실패 시 `git rev-parse --git-common-dir`의 부모, 그래도 실패하면 cwd). 보존 사본을 리드가 읽을 수 있는 곳에 떨어뜨리는 문제는 이 함수로 이미 풀려 있다.
- `acquire`에 `--card` 플래그가 있다(`:275`). 보존 파일 이름에 카드 id를 넣을 수 있다.
- lock 레코드는 `<root>/.moai/state/integration-lock.json`이며(`internal/kanban/integration_lock.go:151`), `.moai/state/`는 gitignore 대상이다(`.gitignore:224`, `:311`).

## §B 설계 결정 4건

되돌리기 어려운 순서로 먼저 놓는다. 각 항목은 고른 안과 기각한 안의 근거를 함께 적는다.

### D1 — 단정이 사는 자리: 술어 1개 · 호출자 2개

**고른 안**: 술어를 `internal/kanban`에 함수 하나로 두고, 그것을 부르는 CLI 표면을 둘 만든다 — (a) `moai integration acquire`의 전제(창 기록 **전에** 실행), (b) 새 읽기 동사 `moai integration preflight`(단독 실행).

기각한 안:

- **(a) 단독** — `acquire` 안에만 두면 사람이 손으로 확인할 수 없고, 리드가 보고서에 넣을 표면도 없다. REQ-PSD-011/012가 요구하는 표면이 없어진다.
- **(b) 단독** — 별도 동사만 두면 실행 여부가 사회적 약속이 된다. `internal/hook/integration_lock_guard.go` 머리말이 명시적으로 적어 둔 t181의 결함("announcement is a social protocol, so a lane that skips it is stopped by nothing")과 같은 모양이다. 통합 락 자체가 그 결함을 닫으려고 만들어졌는데, 그 위에 같은 결함을 다시 얹는 셈이다.
- **(c) doctrine 문구만** — 뮤턴트가 존재할 코드가 없으므로 REQ-PSD-013을 만족시킬 수 없다. 문서 갱신은 이 카드에도 들어가지만 그것만으로는 부족하다.

`acquire`가 옳은 지점인 이유는 cwd다. `kanban-dispatch.md` § Integration into the release branch is self-served와 `CLAUDE.local.md` §4.1의 절차상 레인은 **자기 카드 워크트리에서** `moai integration acquire --name <lane>`을 실행한 다음 release 워크트리로 들어간다. 즉 acquire 시점의 cwd가 곧 병합될 트리다.

다만 그것은 절차상 그렇다는 것이지 코드가 강제하는 사실이 아니다. 그래서 REQ-PSD-004는 대상 트리를 "호출자의 작업 트리 최상위"로 못박고, 출력은 어느 트리를 쟀는지 경로로 밝힌다. 다른 트리에서 실행하면 그 트리를 재고, 잰 대상을 말한다 — 조용히 엉뚱한 트리를 재고 통과시키지 않는다.

배치: `resolveIntegrationTarget()`(`:117`)이 푸는 문제(창이 어느 브랜치를 잠그는가)와는 다른 축이다. 그 함수는 **잠글 대상**을 고르고, drift 검사는 **호출자가 서 있는 트리**를 본다. 섞지 않는다.

### D2 — 적중 시 거절인가 경고인가

**고른 안(운영자 결정 2026-09-06)**: 계열 **(가) 기본 OFF + 켜면 거절**. `workflow.settings_drift_gate.enabled`의 기본값은 `false`이고, 그 플래그가 게이트하는 것은 **거절 층 하나**다. 검출·보존·원장 기록은 플래그 값과 무관하게 `acquire`마다 항상 돈다(REQ-PSD-016). 우회는 새 플래그 `--allow-settings-drift` 하나이며, 우회 사실은 lock 레코드와 출력 양쪽에 남는다.

운영자 근거 — 옮겨 적는다:

- t334는 **인스턴스 둘, 실측된 피해 0**이다. 배포 사용자에게 기본 거절을 보내기에는 근거가 얇다.
- 9일의 실명은 "막지 않아서"가 아니라 **"아무도 보지 않아서"** 생겼다. (가)는 그것을 고친다 — 검출·보존·원장이 여전히 매 `acquire`마다 돌기 때문이다. 게이트는 t334를 잡는다. 다만 유지자가 켜지 않는 한 병합을 멈추지 않는다.

t487 §Q3의 "리드에게 blocker 보고"와 어긋나지 않는다는 점을 적어 둔다 — 그 문구가 요구하는 것은 **보고**이고, (가)에서도 보고는 매번 나간다. opt-in이 되는 것은 멈춤뿐이다. 이 저장소는 그 층을 로컬 config에서 켠다.

설계 근거(운영자 답 이전부터 유효했고 답에 의해 바뀌지 않은 것):

- **저장소의 네 가드와 같은 계열이다.** `BranchGuard` / `AgentModelGuard` / `AgentStopGuard` / `IntegrationLock`은 전부 기본 `false`이고(`internal/config/defaults.go:877-900`), 주석이 "관찰과 권고는 항상 돌고 거절만 opt-in"이라고 명시한다. 이 게이트는 그 정격을 그대로 따른다 — 새 계열을 열지 않는다.
- `acquire`에는 이미 "기본 거절 + 기록되는 우회"의 선례가 있다: 살아 있는 보유자가 있으면 거절하고 `--force`로만 빼앗으며, 빼앗은 사실을 출력한다("recorded, never silent" — `internal/cli/integration.go:264-268`; 거절 자체는 `kanban.AcquireIntegrationLock`의 `ErrIntegrationLockHeld`로 난다 — `internal/kanban/integration_lock.go:53`).
- **`--force`에 얹지 않는 이유**(REQ-PSD-010): `--force`는 "보유자에게서 창을 빼앗는다"는 축이다. 한 플래그로 두 결정을 표현하면 기록이 어느 쪽을 의도했는지 말하지 못한다. 리드 지시도 이 중복 적재를 기각했다.
- 관찰층을 게이트하지 않는 것은 `AgentStopGuardConfig` 주석의 규율을 그대로 따른 것이다(`internal/config/types.go:670-679`: 기록과 감사는 플래그와 무관하게 돈다).

**기본 자세 — 종결(운영자 결정 2026-09-06).** 이 결정에 이르기까지 제시한 지형을 기록으로 남긴다. 축은 둘이고 이 트리에서 관측되는 계열은 셋이다.

| 계열 | 실측 | 선례 수 |
|---|---|---|
| (가) 기본 OFF + 거절 | `BranchGuard` / `IntegrationLock` / `AgentModelGuard` / `AgentStopGuard` 전부 `Enabled: false`(`internal/config/defaults.go:877-900`). 주석이 명시한다 — 관찰과 권고는 항상 돌고 **거절만** opt-in이다 | 4 |
| (나) 기본 ON + 권고 | `AstGrepGate{Enabled: true, BlockOnError: false, WarnOnlyMode: true}`(`defaults.go:607-611`, 주석: *"findings reported, commits never blocked; blocking is opt-in via gate.yaml"*), `GraphFreshness{Enabled: true, Blocking: false}`(`defaults.go:616-618`) | 2 |
| (다) 기본 ON + 거절 | `internal/config`에서 `Blocking: true` / `BlockOnError: true` 히트 **0건**(파이프 없이 측정) | **0** |

즉 이 저장소에서 실제로 갈리는 축은 켜짐/꺼짐이 아니라 **차단이냐 권고냐**다 — 기본 ON인 것은 전부 권고이고, 거절하는 것은 전부 기본 OFF다. 초안이 고른 것은 (다), 선례 0건인 계열이었다.

**운영자는 (가)를 골랐다.** 선례 4건의 계열이고, 위 근거대로 검출·보존·원장은 그대로 돈다.

이 답에 묶여 있던 다섯 곳을 한 번에 처리했다.

| 걸린 곳 | 결정 결과 |
|---|---|
| 기본값 상수 | `workflow.settings_drift_gate.enabled` 기본 `false` |
| M5의 `acquire` 동작 | 기본은 창을 내주고 보고만 한다. 켠 설정에서만 거절 |
| `AC-PSD-009` / `AC-PSD-010` 전제 | 둘 다 Given에 `enabled: true`를 명시(기본값이 아니므로 픽스처가 켠다) |
| M2 뮤턴트 2(판정 반전)의 포착 픽스처 | 거절 층을 켠 설정에서 구성한다 |
| `AC-PSD-009`의 대조 없는 단정 | 부재 단정 규율에 맞춰 양성 대조를 붙였다. 조항의 예외 조항은 삭제 |

여기에 결정이 새로 요구한 것이 하나 있다 — **REQ-PSD-016**. 운영자 근거의 핵심이 "검출·보존·원장은 여전히 매 `acquire`마다 돈다"인데, 킬 스위치를 게이트 **전체**에 건 구현은 그 전제를 무너뜨리면서 다른 AC를 전부 통과한다. 그래서 요구로 못박고 `AC-PSD-013`으로 반증한다. 산문으로 둘 근거가 아니다.

술어·보존·원장·`preflight`, 그리고 `AC-PSD-011`/`AC-PSD-012`(판정층에 걸어 둔 안전 단정)는 이 결정과 무관하게 그대로다.

**킬 스위치 설정 키 — `workflow.settings_drift_gate.enabled` 새 블록으로 확정한다.** 기존 `workflow.integration_lock` 아래 하위 키로 넣는 안은 기각했고, 근거는 그 키가 스스로 밝힌 계약이다. `internal/config/types.go:650-659`의 `IntegrationLockConfig` 주석은 그 플래그가 게이트하는 것이 **PreToolUse 훅의 deny 층 하나**임을 명시하고(*"only the DENY layer is gated, exactly as BranchGuard gates only its deny"*), 레코드와 `moai integration` CLI는 그 플래그와 무관하다고 못박는다. 새 게이트의 거절은 다른 표면(`moai integration acquire`)에서 다른 시점에 일어난다. 하나의 `integration_lock.enabled`가 두 층을 게이트하면 설정이 어느 쪽을 끄려던 것인지 말하지 못하고, 그것은 REQ-PSD-010이 `--force`에 우회를 얹지 않는 이유와 **같은 모양의 중복 적재**다. 플래그 축에서 세운 논거를 설정 키 축에서 버릴 이유가 없다.

이 결정은 위 기본 자세 항목과 독립이다 — (가)/(나)/(다) 어느 답에서도 킬 스위치는 필요하고, 새 블록이라는 답은 바뀌지 않는다.

### D3 — 보존 사본이 떨어지는 자리

**고른 안**: `<integrationLockRoot()>/.moai/state/settings-drift/` 아래.

- 파일 이름: `settings.json.<card-or-branch>.<UTC RFC3339 밀리초 압축형>.<sha256 앞 8자>[.<n>]` — 예: `settings.json.t488.20260906T031500.123Z.4f455d94`.
- **충돌 규칙**: 위 세 성분이 같은 이름이 이미 있으면 덮어쓰지 않고 `.2`, `.3` … 순번 접미를 붙인다. 시각 성분을 밀리초로 올려도 충돌 확률만 줄 뿐 0이 되지 않으며(같은 카드·같은 내용으로 연달아 부르면 같은 밀리초에 들어갈 수 있다), "덮어쓰지 않는다"를 **결정적으로** 만드는 것은 순번 접미뿐이다. 초 해상도 + 접미 규칙 없음이던 종전 규칙은 `AC-PSD-007(c)`를 간헐 실패시키는 형태였다.
- 원장: 같은 디렉터리의 `ledger.jsonl`에 한 줄. 필드: 측정 시각, 카드 id, 브랜치, 원본 워크트리 절대경로, 원본 파일 경로, 보존 사본 경로, sha256, 바이트 크기, 일치 줄 수, 우회 여부.

근거:

- 리드는 레인의 워크트리 안에 있지 않다. `integrationLockRoot()`가 이미 "모든 linked worktree에서 보이는 하나의 디렉터리"를 푸는 함수이고(`internal/cli/integration.go:36-45`의 주석이 그 성질을 명시한다), lock 레코드가 이미 그 아래 산다(`internal/kanban/integration_lock.go:151`). 같은 자리를 쓰면 리드는 이미 아는 경로만 보면 된다.
- `.moai/state/`는 gitignore 대상이므로(`.gitignore:224`, `:311`) 보존 사본은 untracked로 남는다. 이것이 의도다 — REQ-PSD-007대로 자동 커밋하지 않는다. 이 파일은 토큰·절대경로·tmux pane id를 담을 수 있어서, 저장소 이력에 넣으려면 t487 L2가 했던 것처럼 시크릿 스캔을 먼저 사람이 해야 한다. 승격 대상 자리는 `.moai/reports/<card-id>/preserved-copies/`이며 그 이동은 사람의 손이다.
- 이름에 시각과 sha256을 넣는 이유: 두 번째 인스턴스가 첫 번째를 덮어쓰면 안 된다. 운영자가 t334를 대조 표본으로 남기라고 한 것과 같은 이유다.
- 카드 id는 `acquire --card`에서 온다(`:275`). 비어 있으면 브랜치 이름으로 떨어지고, 그것도 비면 `unknown`을 쓰되 원장에는 워크트리 절대경로가 있으므로 추적은 끊기지 않는다.

### D4 — 창 밖에서도 부를 수 있는가

**고른 안**: 부를 수 있다. `moai integration preflight [--card <id>] [--json]`.

- 기본 출력은 사람이 읽는 한 줄 + 적중 시 보존 경로·sha256.
- `--json`은 `{"status": "clean"|"drift"|"undetermined", "match_count": int, "worktree": ..., "path": ..., "sha256": ..., "preserved": ..., "error": ...}`.
  - `status`가 판정 carrier이고 **항상** 있다. 불리언 `drift` 필드를 판정 carrier로 쓰지 않는 이유는 REQ-PSD-014가 요구하는 세 번째 상태를 두 값으로는 표현할 수 없기 때문이다 — 불리언을 쓰면 "측정 못 함"이 `false`로 접히거나 필드가 생략돼 소비자의 기본값이 그것을 통과로 읽는다. 어느 쪽이든 9일 실명의 재현이다.
  - `match_count`는 `clean`/`drift`에서만 나온다. `undetermined`에서는 **생략한다** — 0을 채우면 그 값이 곧 통과 신호가 된다.
  - `error`는 `undetermined`에서만 나오고 실패 사유를 담는다.
- 적중 시 보존과 원장 기록을 수행한다. 즉 순수 읽기 전용은 아니며, 쓰기는 **primary 체크아웃의 상태 디렉터리**로만 나간다. 대상 워크트리에는 아무것도 쓰지 않는다(REQ-PSD-006).
- 종료 코드는 편의 신호(적중 시 비-0)로 두되, 어떤 AC도 여기에 걸지 않는다(REQ-PSD-003). t474가 관측한 "뮤턴트에서도 `grep`이 0을 돌려줘 게이트가 뒤집힌" 사례가 근거다.

`status`가 아니라 새 동사인 이유: `status`는 "누가 창을 갖고 있나"를 답하는 기존 계약이고, 여기에 drift 판정을 섞으면 그 출력의 의미가 둘이 된다.

## §C 사전 확인(run 착수 전)

- `git rev-parse --show-toplevel`이 `.claude/worktrees/t488`이고 브랜치가 `WT-premerge-drift-assert`인지.
- `internal/cli/integration.go` / `internal/kanban/integration_lock.go` / `internal/config/{types,defaults}.go`가 위 인용 줄번호대로인지(줄은 드리프트한다 — 착수 시점에 재확인하고, 다르면 인용을 고친다).
- `.gitignore`의 `.moai/state/` 제외가 유지되는지(보존 사본이 untracked로 남는다는 전제).

## §D 제약

spec.md §4가 정본이다. run 단계에서 특히 걸리는 두 가지만 다시 적는다: **t334 워크트리 무수정**, **로컬 전체 스위트 금지**(건드린 패키지만).

## §E 자기 검증

- 두 방향 뮤턴트 검증(hit / pass)이 실제 출력으로 인용됐는가.
- `--no-optional-locks`가 **실행 경계에 기록된** argv 단정으로 고정됐는가(행동만으로는 신뢰성 있게 구별되지 않는다 — 인덱스 쓰기 여부는 stat 캐시 상태에 좌우된다 — 그래서 문자열 단정을 쓴다).
- 판정이 종료 코드에 걸려 있지 않은가.
- 보존 사본이 대상 워크트리 밖으로만 쓰이는가.

## §F 마일스톤

**M1 — 술어와 그 테스트 (우선순위 High).**
`internal/kanban`에 drift 술어를 둔다. 서명 개형: 트리 경로와 **명령 실행기**를 받아 `(matchCount int, rawOutput string, err error)`를 돌려준다.

실행기를 서명에 드러내는 이유는 `AC-PSD-005`의 관측 지점 때문이다. argv는 정확히 한 곳에서 구성되고, 실행기는 그 구성 결과를 그대로 받으며 자기에게 넘어온 argv를 기록한다. 테스트는 **기록된 argv**를 읽어 단정한다. argv 빌더를 따로 노출해 그 반환값을 부르는 형태는 금지한다 — 실행 지점이 `exec.Command("git", …)`를 따로 적어 두면 빌더와 실행 경로가 갈리고, 어느 한쪽에서 플래그를 빼도 테스트가 초록으로 남는다(t477이 관측한 "틀린 모양이 통과하는" 공허한 매치). 이 기록은 `AC-PSD-007(d)`와 `AC-PSD-008`의 "대상 트리를 수정하는 명령이 없다"도 함께 기계화한다.

argv는 `git --no-optional-locks status --porcelain -- .claude/settings.json` 고정. `t.TempDir()` 안에 픽스처 5종을 구성한다.

| 픽스처 | 구성 | 기대 |
|---|---|---|
| F1 clean | `.claude/settings.json` 커밋 후 무수정 | `matchCount == 0` |
| F2 dirty | 같은 파일을 커밋 후 수정 | `matchCount == 1` |
| F3 다른 파일만 dirty | `.claude/settings.json`은 깨끗, `README.md`만 수정 | `matchCount == 0` (경로 오지정 뮤턴트 검출) |
| F4 둘 다 dirty | 두 파일 모두 수정 | `matchCount == 1` (경로 인자 누락 뮤턴트 검출) |
| F5 저장소 아님 | `git init` 없이 파일만 둔 디렉터리 | 오류가 나고, 통과 판정으로 읽히지 않는다 (REQ-PSD-014) |

픽스처의 `git init`은 `-b main`으로 초기 브랜치를 못박는다 — SPEC-GIT-STATUS-FIXTURE-001(카드 t474)이 실측한 대로, 못박지 않으면 ambient `init.defaultBranch`에 따라 CI와 로컬이 갈린다. `user.name`/`user.email`도 픽스처 로컬 config로 지정한다.

**M2 — 뮤턴트 반증 (우선순위 High).**
아래 **여섯** 뮤턴트를 각각 뒤집어 실제로 잡히는지 확인하고 판정서에 인용한다(다섯은 M1의 테스트가, 여섯 번째는 `AC-PSD-007(d-3)`이 잡는다). "잡을 것이다"가 아니라 뒤집어 보고 RED를 관측한다.

| 뮤턴트 | 층 | 잡는 픽스처/단정 |
|---|---|---|
| 술어 삭제(항상 0) | 술어 | F2 / `AC-PSD-001` |
| 판정 반전(`drift = matchCount > 0`을 뒤집음) | **게이트** | **`AC-PSD-009`** — 거절 경로에서 `integration-lock.json`이 생성되지 않음 |
| 경로를 `.claude/settings.local.json`으로 오지정 | 술어 | F2(0이 나옴) / `AC-PSD-003` |
| 경로 인자 자체 누락(트리 전체 status) | 술어 | F4(2가 나옴) / `AC-PSD-004` |
| `--no-optional-locks` 제거 | 술어 | 실행 경계에 기록된 argv 문자열 단정 / `AC-PSD-005` |
| **게이트 경로에서 실행기 우회**(보존·커밋을 `exec.Command`로 직접 실행) | **게이트** | **`AC-PSD-007(d-3)`** — 기록을 읽지 않는 직접 관측(양쪽 루트의 HEAD + 원격 ref)에서 잡힌다 |

판정 반전이 **F1이 아닌** 이유를 명시해 둔다. F1이 단정하는 것은 술어의 반환값(`matchCount == 0`)이고, 판정 반전은 술어가 아니라 게이트에서 일어난다. 게이트를 뒤집어도 F1의 `matchCount`는 그대로 0이므로 F1은 초록으로 남는다. 틀린 좌표는 틀린 실험을 부른다 — 구현자가 F1으로 이 뮤턴트를 뒤집어 보면 초록이 나오고, 그때 "잡히지 않는다"고 결론짓거나 억지로 빨간불을 만들게 된다. `AC-PSD-009`의 전제는 §B D2의 운영자 결정으로 정해졌다 — 거절은 기본값이 아니므로, 이 뮤턴트의 포착 픽스처는 `workflow.settings_drift_gate.enabled: true`를 **명시적으로 켠** 설정에서 구성한다. 기본 설정으로 구성하면 거절 경로 자체가 돌지 않아 뮤턴트가 잡히지 않는다.

마지막 행은 REQ-PSD-013이 게이트 경로 전체로 넓어지면서 생긴 뮤턴트다. 기록 위의 부재 단정((d-1)/(d-2))만으로는 잡히지 않으므로 — 술어는 실행기를 통과하고 커밋만 우회하면 기록은 멀쩡하다 — 기록을 전혀 읽지 않는 `AC-PSD-007(d-3)`이 유일한 포착기다. 이 행이 없으면 DoD의 뮤턴트 셈에서 빠져 아무도 (d-3)을 뒤집어 보지 않은 채 카드가 닫힌다.

마지막 항목이 별도인 이유: 플래그를 빼도 **출력과 반환값은 같다**. 다른 것은 부작용뿐이다 — `--no-optional-locks` 없는 `git status`는 인덱스를 쓸 수 있고, 그것이 카드 t485가 실측해 이 카드의 근거가 된 현상이다. 부작용이 다르므로 행동 기반 검증이 *원리적으로* 배제되지는 않는다(픽스처의 `.git/index` mtime/inode를 술어 실행 전후로 재는 형태가 성립할 수 있다). 그러나 `git status`가 인덱스를 실제로 다시 쓰는지는 stat 캐시가 낡았는지에 달려 있어, 갓 만든 픽스처에서는 쓸 수도 쓰지 않을 수도 있다 — 정확히 말하면 **신뢰성 있게는 구별되지 않는다**. 그래서 실행 경계에 기록된 argv를 문자열로 단정한다. t477의 교훈(제거 뮤턴트만으로는 약하고 "틀린 모양" 뮤턴트가 필요하다)이 여기 그대로 적용된다.

**M3 — 보존과 원장 (우선순위 High).**
D3의 디렉터리·파일명·충돌 접미 규칙, `ledger.jsonl` 한 줄 추가(append), sha256 계산. 테스트: 적중 시 보존 파일이 원본과 바이트 일치하고, 원본이 수정되지 않았으며, 같은 카드·같은 내용으로 **연속 두 번** 적중시켜도 보존 파일이 두 개로 남는 것(두 실행이 같은 밀리초에 들어가도 성립해야 한다 — 접미 규칙만이 이것을 결정적으로 만든다). 보존 디렉터리를 쓰기 불가로 만든 픽스처에서 drift 판정이 통과로 뒤집히지 않는 **동작**(REQ-PSD-015). 그 동작의 **증거**는 `preflight --json`으로 잡히므로 `AC-PSD-012`는 M4가 청구한다 — `preflight`는 M4에서 생기고, M3 시점에는 AC-012의 When 절을 실행할 표면이 없다.

이 마일스톤의 적중 픽스처에는 `t.TempDir()` 안의 로컬 bare 저장소를 `origin`으로 붙이고, **브랜치를 실제로 push해 둔다**. `AC-PSD-007(d-3)`이 자동 push를 실행기 기록과 무관하게 반증하려면 비교할 원격 ref가 있어야 하는데, 원격만 만들고 push하지 않으면 `git ls-remote --heads`가 `rc=0`에 **빈 출력**을 내고(실측) 전후가 둘 다 비어 동일성 단정이 공허하게 통과한다. 종료 코드로는 이 퇴화를 구별할 수 없으므로 픽스처가 사전 출력을 비어 있지 않게 만드는 것이 유일한 방어다.

**이것이 이 픽스처에서 가장 나오기 쉬운 실수 모양이다** — 원격을 만드는 것은 눈에 보이는 단계이고 브랜치를 push하는 것은 잊기 쉬운 단계인데, 잊어도 오류가 나지 않는다. 그래서 픽스처 구성 직후 `ls-remote` 출력이 비어 있지 않음을 한 번 확인하고 넘어간다. 이 확인이 없으면 `AC-PSD-007(d-3)`의 push 반증은 조용히 아무것도 검사하지 않는 상태로 남는다.

**M4 — `preflight` 동사 (우선순위 High).**
D4의 표면. `--json`의 `status`·`match_count`·`sha256` 필드 포함(`status`는 항상, `match_count`는 `undetermined`에서 생략). 기존 세 동사의 출력 계약은 건드리지 않는다. 안전 불변식 두 개의 반증이 **모두 이 표면에 걸려 있다**:

- 술어 실패(F5)를 통과로 표현하지 않는 것 — REQ-PSD-014 / `AC-PSD-011`. `status`가 `"undetermined"`이고 `match_count`는 생략되며 `error`가 사유를 담는다. `clean`으로 접지도, 0으로 채우지도 않는다.
- 보존 실패가 판정을 뒤집지 않는 것 — REQ-PSD-015 / `AC-PSD-012`. 보존 디렉터리를 쓰기 불가로 만든 픽스처에서 `status`가 여전히 `"drift"`이고 `match_count`가 1이다(M3이 넣은 동작을 여기서 관측한다). `match_count`는 여기서 그대로 남는다 — 측정 자체는 성공했고 실패한 것은 보존뿐이므로 `undetermined`가 아니라 `drift` 상태이고, `match_count` 생략은 `undetermined`에만 걸린다.

**우선순위가 Medium이 아니라 High인 이유**: REQ-PSD-014/015는 게이트가 가장 위험하게 무력화되는 두 경로이고, 그 반증의 유일한 관측 지점이 이 표면이다. 이 마일스톤이 뒤로 밀리면 두 안전 단정이 함께 밀린다.

**M5 — `acquire` 전제와 우회 (우선순위 Medium).**
창 기록 **전에** 술어를 돌린다. 여기서부터 두 갈래이고, 갈리는 것은 거절뿐이다.

- **기본 설정(`enabled: false`)** — 적중해도 창을 기록한다. 그러나 검출·보존·원장은 그대로 돌고, 출력에 drift 사실과 보존 경로가 나간다(REQ-PSD-016 / `AC-PSD-013`). 이것이 운영자가 고른 자세다.
- **켠 설정(`enabled: true`)** — 적중이면 lock을 쓰지 않고 거절하며, 보존 경로·sha256·리드 보고 문구를 출력한다(`AC-PSD-009`). `--allow-settings-drift`가 있으면 기록하되 우회 사실을 lock 레코드 필드와 출력에 남긴다(`AC-PSD-010`).

테스트에서 확인할 것 둘: 켠 설정의 거절 경로에서 `integration-lock.json`이 **생성되지 않는다**(거절인데 창이 잡히면 최악이다), 그리고 기본 설정에서 보존 사본과 원장 한 줄이 **여전히 생긴다**(킬 스위치를 게이트 전체에 걸면 여기서 잡힌다).

**M6 — doctrine 문구 (우선순위 Low).**
`.claude/rules/moai/workflow/kanban-dispatch.md` § Integration into the release branch is self-served와 `CLAUDE.local.md` §4.1의 레인 절차에 이 전제를 한 줄로 넣는다. `.claude/rules/**`는 `moai update`의 관리 대상 뿌리이므로 Template-First 규율이 걸린다(`CLAUDE.local.md` §2/§2.3 — 로컬만 고치면 다음 update에서 템플릿판으로 덮인다). 미러 대상은 착수 시점 확인 사항이 아니라 이미 실측으로 정해져 있다.

| 파일 | 템플릿 미러 | 처분 |
|---|---|---|
| `.claude/rules/moai/workflow/kanban-dispatch.md` | `internal/template/templates/.claude/rules/moai/workflow/kanban-dispatch.md` — **존재**(34268 bytes) | 로컬과 미러를 함께 고친다 |
| `CLAUDE.local.md` | **없음** | 로컬만 고친다 — 그 파일 자신이 Local-Only 목록에 자기를 올려 둔 의도적 로컬 전용 파일이다 |

즉 M6의 실제 작업은 3파일이 아니라 **2파일 + 1미러 = 3개 쓰기**이며, 그중 템플릿 쓰기는 하나다. 넣을 문구는 카드 id(`t488`)와 SPEC ID 없이 중립으로 쓴다(`CLAUDE.local.md` §2.1 템플릿 중립성). 중립으로 쓰면 로컬과 미러의 해당 절이 동일해지므로 `diff` 한 줄로 미러 동일성을 확인할 수 있다.

## §G 안티패턴

- 종료 코드로 판정하기 — t474에서 게이트가 뒤집힌 모양 그대로다.
- 제거 뮤턴트만으로 만족하기 — "틀린 모양" 뮤턴트가 없으면 공허한 초록을 못 잡는다(t477).
- 실제 워크트리(특히 t334)를 픽스처로 쓰기 — 대조 표본을 오염시키고, cross-tree `git`은 가드가 거부한다.
- 검출된 파일을 "고쳐 주기" — REQ-PSD-006 위반이자 데이터 파괴다.
- 감시 대상을 조용히 넓히기 — 운영자가 고르지 않은 확대다.

## §H 교차 참조

- `.moai/reports/t487/verdict.md` §Q3 — 원문 권고와 그 premise.
- `.moai/specs/SPEC-GIT-STATUS-FIXTURE-001/` — git 픽스처의 초기 브랜치 못박기 선례.
- `internal/hook/integration_lock_guard.go` — fail-open 규율과 deny sentinel 형태.
- `.claude/rules/moai/core/verification-claim-integrity.md` — 관측하지 않은 주장 금지(뮤턴트를 "잡을 것이다"로 적지 않는 이유).

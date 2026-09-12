# t594 Verdict — release-update 하네스 수리 2건

카드: t594 · release-update 하네스 수리 2건
브랜치: `WT-release-update-harness` · 워크트리: `.claude/worktrees/t594` · 베이스: `eabce7444`

---

## Claim

1. **건 (1) — GLM 백엔드 웹 라우팅.** `hns-release-update-specialist.md` Phase 3의 `[HARD]` doc-fetch 조항이
   세션 백엔드별 도구를 명시한다: Claude 백엔드는 `WebFetch`, GLM 백엔드는 `mcp__web_reader__webReader`.
   frontmatter `tools:`가 두 MCP 도구를 싣고 있어 GLM 분기가 실행 가능하다.
2. **건 (2) — 셸 분해 독트린.** Phase 2에 "계수는 plain 단일 명령으로, 복합 `for` 루프나 런타임 생성
   `sed`/`awk` 프로그램 금지" 독트린이 존재한다.
3. **범위 무변경.** 변경 파일은 dev-only 하네스 1본뿐이며 템플릿 미러는 생성되지 않았다(생성하면 독트린 위반).

---

## Evidence

모든 명령은 이 워크트리(`.claude/worktrees/t594`, HEAD `eabce7444` + 워킹트리 변경)에서 실행했다.

### 변경 범위

```
$ git diff --stat
 .../harness/hns-release-update-specialist.md       | 37 +++++++++++++++++-----
 1 file changed, 29 insertions(+), 8 deletions(-)

$ git status --porcelain
 M .claude/agents/harness/hns-release-update-specialist.md
```

### 에이전트 린트 (frontmatter 파싱 + 규칙)

```
$ moai agent lint
Summary: 25 total (0 errors, 25 warnings)
exit=0
```

경고 25건은 **전부 기존 것**이며 이 편집과 무관하다. 대상 파일에 붙은 유일한 경고는 LR-05
(`isolation: worktree` 부재)인데, 그 필드는 건드리지 않았음을 양변 측정으로 보인다:

```
$ git show eabce7444:.claude/agents/harness/hns-release-update-specialist.md | grep -c "^isolation:"
0
$ grep -c "^isolation:" .claude/agents/harness/hns-release-update-specialist.md
0
```

같은 LR-05 경고가 편집하지 않은 `hook-ci-specialist.md` / `quality-specialist.md` /
`workflow-specialist.md`에도 동일하게 붙는다 — 파일 단위 결함이 아니라 하네스 전역의 기존 상태다.
경고 총계는 최종 편집 전후 모두 25로 동일했다.

### 템플릿 가드 (dev-only 격리 · 중립성 · 누출)

```
$ go test ./internal/template/ -run 'Leak|Neutrality|Harness|Agent|Command' -count=1
ok  	github.com/modu-ai/moai-adk/internal/template	1.807s

$ go test ./internal/template/ ./internal/config/ -run 'Agent|Frontmatter|Tools' -count=1
ok  	github.com/modu-ai/moai-adk/internal/template	2.696s
ok  	github.com/modu-ai/moai-adk/internal/config	0.490s
```

dev-only 격리는 디렉터리 부재로 성립한다(`agent-authoring.md`: 이 경로는 템플릿에 **존재해서는 안 된다**):

```
$ ls internal/template/templates/.claude/agents/harness/
ls: internal/template/templates/.claude/agents/harness/: No such file or directory
```

### 교차참조 무결성

`Option C` 라벨 참조가 끊기지 않았음(라벨 `A/B/C`는 보존, 대시 뒤 설명만 변경):

```
$ grep -rn "Option C" .claude/agents/harness/hns-release-update-specialist.md
85:**Option C — web fallback** (backend-routed exactly as Phase 3: ...
177:If Option C: also create child SPEC stub directories (orchestrator-direct).
```

---

## Baseline-attribution

- 워크트리 `.claude/worktrees/t594`, 브랜치 `WT-release-update-harness`, 베이스 커밋 `eabce7444`
  (`git rev-parse --short HEAD` = `eabce7444`, `git rev-parse --show-toplevel` =
  `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t594`).
- 위 측정은 전부 이 실행에서, 이 트리를 상대로, 커밋 전 더티 상태에서 수행했다.
- 변경 파일: `.claude/agents/harness/hns-release-update-specialist.md` 1본.
- 착수 전 교차확인: `moai todo pr t594 --json` → `[{"card_id":"t594","outcome":"no-link"}]`,
  `git log -S'WebFetch' --all -- <파일>` → f55aefef3(개명) 1건뿐, `git log -S'web_reader' --all -- <파일>` → 0건.

---

## 카드 문구 대비 편차 (리드 결정으로 기록)

### 편차 1 — 건 (1)의 처방 변경: curl 폴백 → MCP 라우팅

카드 본문은 "GLM 백엔드용 **curl 폴백** 명시"를 지시했으나, 그대로 쓰면
always-loaded `[HARD]` 규칙과 충돌한다. `glm-web-tooling.md`가 못박은 바:

- §32/§40 — GLM 세션에서 내장 `WebSearch`/`WebFetch`는 **PROHIBITED**
- §37 — `WebFetch` → `mcp__web_reader__webReader` **MUST** 교체
- §171 — 내장 `WebFetch` 호출은 이름 붙은 안티패턴 `AP-GWT-002`

규칙과 카드 문구가 충돌하면 규칙이 이긴다(리드 판정). 따라서 curl이 아니라 MCP 라우팅으로 구현했다.

### 편차 2 — frontmatter `tools:` 확대 (범위 포함, 리드 결정)

문구만 고치면 "적혀 있지만 못 하는 절차"가 된다 — 편집 전 `tools:`에는 MCP 웹 도구가 없어
GLM 분기가 원리상 실행 불가였다. 두 도구를 추가했다:

- `mcp__web_reader__webReader` — Phase 3 + Phase 1 Option C의 fetch 경로
- `mcp__web_search_prime__webSearchPrime` — Phase 1 Option C의 last-resort 검색 경로
  (이 에이전트는 `WebSearch`를 선언·사용하므로 같은 `[HARD]` 금지에 걸린다. fetch만 고치고
  search를 두면 반쪽 수리가 되므로 함께 라우팅했다 — 도구 2개 + 한 절 수정의 최소 범위)

### 편차 3 — 건 (2) 재정의: "레시피 교체" → "독트린 추가" (리드 승인)

**카드의 전제가 거짓이다.** "Phase 2의 sed 범위 계수 레시피"는 존재하지 않는다. release-update
하네스 4본 전수 측정:

| 파일 | `sed` 매치 |
|---|---|
| `.claude/agents/harness/hns-release-update-specialist.md` | 0 |
| `.claude/commands/harness/release-update.md` | 0 |
| `.claude/commands/harness/release-update/manifest.json` | 0 |
| `.claude/workflows/hns-release-update-run.js` | 1 — **오탐** (40행 주석의 `never relea`**`sed`**) |

즉 실제 `sed` 0건. Phase 2 본문은 티어 분류표와 출력 형식뿐이고 셸 레시피 자체가 없었다.
실제로 거부된 것은 **스윕 수행 중 에이전트가 즉석에서 조립한** 복합 명령이다
(`.moai/research/cc-update-2.1.263-to-2.1.267.md` E15). 교체할 대상이 없으므로 독트린을 **추가**했다.

### 편차 4 — 카드 근거 경로 정정

카드는 근거를 `.moai/research/cc-update-2.1.263-to-2.1.267.md REQ-HRR-006`이라 적었으나:

- 그 research 문서에 `REQ-HRR-006` 문자열은 **없다**.
- `REQ-HRR-006`은 `SPEC-HARNESS-EVO-RUN-REPORT-001`의 **findings 방출 메커니즘** id이지
  finding id가 아니다.
- 두 finding의 실제 기록처는 **`.moai/state/last-cc-version.json:29`** 의 note 필드
  ("Harness improvement findings 2 (REQ-HRR-006, conf 0.70 each): specialist Phase 3 WebFetch
  mandate vs GLM-backend curl reality; Phase 2 sed-range counting recipe trips the worktree
  guard … — carded as t594").

카드 본문 자체는 큐 이력이라 고치지 않는다(리드 판정). **다음 사람이 같은 근거로 같은 카드를
다시 내지 않도록** 여기에 남긴다.

---

## Gaps — 관측하지 않은 것

- **런타임 미검증.** GLM 백엔드에서 이 specialist를 실제로 돌려 `mcp__web_reader__webReader`가
  호출되는지는 확인하지 않았다. 문서·frontmatter 변경이며, 실행 경로는 다음 실제 스윕이 판정한다.
- **MCP 서버 가용성 미측정.** `tools:` 등재는 허용목록일 뿐이다. 해당 도구는 `moai glm tools enable
  webreader` / `websearch`로 등록돼 있어야 실제로 존재한다. 미등록 세션에서의 동작(도구 부재 시
  폴백 형태)은 재지 않았다.
- **전체 테스트 수트 미실행.** 로컬 부하 규율(CLAUDE.local.md §4/§6)에 따라 영향 패키지만 돌렸다.
  전 패키지 판정은 CI 몫이다. 단, 이 변경은 Go 코드를 건드리지 않는 문서 변경이다.
- **Phase 2 독트린의 효과 미측정.** "거부-재시도 1회/스윕"이 0으로 떨어지는지는 다음 스윕에서만
  관측 가능하다. 이 카드는 독트린을 심었을 뿐 감소를 입증하지 않았다.
- **다른 하네스 specialist 미점검.** `hns-github-specialist` / `hns-release-specialist` 등이 같은
  GLM 라우팅 결함을 갖는지는 이 카드 범위 밖이다(아래 참조).

---

## Residual-risk

- **레포 전역 동일 결함.** `git grep -l "web_reader\|glm-web-tooling" eabce7444 -- .claude/agents/`
  → **0건**. 이 파일을 고친 지금도 나머지 모든 에이전트는 GLM 라우팅을 싣고 있지 않다. 이 카드는
  한 파일만 고쳤으므로 전역 결함은 남아 있다 — 리드가 별도 카드(에이전트 전수 점검 축)로 발행 목록에 올렸다.
- **dev-only 자산이라 CI 보호가 얇다.** 이 파일은 템플릿 미러가 없어 template-neutrality CI 가드의
  대상이 아니다. 회귀는 `moai agent lint`와 사람 눈에만 걸린다.
- **`moai update` 삭제 위험 없음(측정).** `.claude/agents/harness/`는 `IsUserOwnedNamespace`
  (`internal/cli/update/plan/plan.go:152`) 보호 대상이라 `CleanMoaiManagedPaths` wipe 범위 밖이다.
  근거는 문서 인용이 아니라 코드다 — `internal/cli/update_characterization_test.go:197`이
  `.claude/agents/harness/my-specialist.md` → `ClassPreserveUserOwned`를 고정하고 있다.
- **Phase 3 표의 백엔드 판별은 산문이다.** 어느 백엔드인지는 에이전트가 `ANTHROPIC_BASE_URL`로
  판단해야 하며(`glm-web-tooling.md` §18), 이 문서는 판별 절차를 기계화하지 않는다 — 오판 가능성은 남는다.

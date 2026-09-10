# t196 카드 전제 재측정

- 측정 트리: `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t196`
- 측정 시점 HEAD: `297a21ea73b24e6605280625e576555e4316263e`
- 브랜치: `WT-codex-skill-neutral`
- 측정 주체: manager-spec (t196-spec). 리드가 건네준 값은 **근거로 채택하지 않고 전부 이 트리에서 다시 쟀다.**

**[iter-2 보충] 리드의 대조 트리는 방증 자격이 없다 — 리드가 값을 철회했다.**

리드가 잰 `main 48239c7dc` 는 배포선(`origin/develop`)보다 **686커밋 뒤진 트리**다:

```
$ git rev-list --count --left-right origin/develop...48239c7dc
686	0
```

오른쪽 `0` 은 그 트리가 배포선의 **엄격한 조상**이라는 뜻이다 — 고유 커밋이 하나도 없다. 아무도 배포하지 않는 지점에서 잰 값이므로 **대조 근거로 세우지 않는다.** 병기는 유지한다(갈린 사실과 사유가 기록으로 남는 편이 낫다) — 다만 아래 표시된 자리는 전부 **무효 방증**이다.

**대신 두 ref 를 직접 읽어 다시 세웠다.** 값이 겹치는 자리들은 이제 리드의 철회된 측정이 아니라 **본 SPEC 이 두 ref 에 각각 대고 읽은 결과**에 근거한다. 이 보충 이후의 모든 ref 대조는 체크아웃 grep 이 아니라 `git ls-tree <ref>` / `git grep <pattern> <ref> -- <paths>` 형태로 수행했다 — **어느 트리를 잰 것인지가 명령 자체에 박히므로 라벨을 잃을 수 없다.** 이번 철회가 정확히 라벨을 잃은 사고였다.

값이 실제로 갈린 자리는 **§4 의 `SPEC-CODEX-*` 개수 하나뿐**이며, 거기서는 `origin/develop` 쪽 값이 정본이다.

---

## 1. 카드 전제 4건

### ① "미러된 스킬 21종" — 반증 (FALSIFIED)

```
$ find internal/template/templates/.claude/skills -mindepth 2 -maxdepth 2 -name SKILL.md | wc -l
      34
$ find internal/template/templates/.claude/skills -mindepth 1 -maxdepth 1 -type d | wc -l
      34
```

배포 대상 스킬은 **34종**이다. 21 은 어느 단위로도 재현되지 않는다.

~~리드의 `main` 트리도 34.~~ **[iter-2 무효 방증]** — 리드가 값을 철회했다(§머리말). 대신 두 ref 를 직접 읽었다:

```
$ git ls-tree -d --name-only 297a21ea7 -- internal/template/templates/.claude/skills/ | wc -l
      34
$ git ls-tree -d --name-only 48239c7dc -- internal/template/templates/.claude/skills/ | wc -l
      34
```

두 ref 가 34 로 일치한다 — 이 수치는 686커밋 구간에서 움직이지 않았다. **이 대조는 리드의 철회된 측정이 아니라 본 SPEC 의 ref 판독에 근거한다.**

두 단위(디렉터리 수, `SKILL.md` 보유 수)가 모두 34 로 일치하므로 "`SKILL.md` 없는 스킬 디렉터리"는 없다.

이 34 라는 값은 `SPEC-CODEX-SKILLS-CANONICAL-001` §A.1 이 이미 확정한 값이며, 그 SPEC 의 HISTORY 는 **자기보다 앞선 두 값(카드의 32, 선행 실측의 36)을 둘 다 틀렸다고 정정**한 기록을 남겼다. 36 은 `ls | wc -l` 이 별칭 때문에 long-format 으로 돌아 `.` `..` 두 줄이 더해진 값이었다. 본 재측정은 `find` 로 독립 수행했고 같은 34 에 도달했다 — 세 번째 오답이 되지 않았다.

### ② "9종이 Claude 전용 도구명 참조" — 반증 (FALSIFIED)

```
$ grep -rlE 'AskUserQuestion|Agent\(|Skill\(|TaskCreate|TaskUpdate|TaskList|TaskGet' \
    internal/template/templates/.claude/skills --include='SKILL.md' | wc -l
      14
```

**14종**이다.

~~리드의 `main` 트리도 14.~~ **[iter-2 무효 방증]** — 리드가 값을 철회했다(§머리말). 대신 두 ref 를 직접 읽었다:

```
$ git grep -lE 'AskUserQuestion|Agent\(|Skill\(|TaskCreate|TaskUpdate|TaskList|TaskGet' \
    297a21ea7 -- 'internal/template/templates/.claude/skills/*/SKILL.md' | wc -l
      14
$ git grep -lE 'AskUserQuestion|Agent\(|Skill\(|TaskCreate|TaskUpdate|TaskList|TaskGet' \
    48239c7dc -- 'internal/template/templates/.claude/skills/*/SKILL.md' | wc -l
      14
```

두 ref 가 14 로 일치한다. **본 SPEC 의 ref 판독에 근거한 대조다.**

목록(`297a21ea7`):

```
moai/SKILL.md
moai-domain-html-report/SKILL.md
moai-foundation-cc/SKILL.md
moai-foundation-core/SKILL.md
moai-foundation-quality/SKILL.md
moai-harness-learner/SKILL.md
moai-kanban-foreman/SKILL.md
moai-meta-harness/SKILL.md
moai-ref-api-patterns/SKILL.md
moai-ref-owasp-checklist/SKILL.md
moai-ref-react-patterns/SKILL.md
moai-ref-seo/SKILL.md
moai-ref-ui-polish/SKILL.md
moai-workflow-spec/SKILL.md
```

식별자별 분해 — **단위는 "일치하는 줄 수"이지 출현 횟수가 아니다** (`grep -c` 는 줄을 센다):

| 식별자 | 일치 줄을 가진 파일 | 파일별 일치 줄 수 |
|---|---|---|
| `AskUserQuestion` | 7 | moai 12 · moai-harness-learner 10 · moai-meta-harness 4 · moai-kanban-foreman 2 · moai-workflow-spec 2 · moai-foundation-cc 1 · moai-foundation-core 1 |
| `Agent(` | 11 | moai 7 · moai-foundation-core 5 · moai-foundation-cc 4 · moai-ref-api-patterns 3 · moai-ref-react-patterns 3 · moai-meta-harness 2 · moai-ref-ui-polish 2 · 나머지 4파일 각 1 |
| `Skill(` | 2 | moai-foundation-quality 2 · moai 1 |
| `TaskCreate`/`TaskUpdate`/`TaskList`/`TaskGet` | 1 | moai 3줄 (출현 6회: TaskCreate 2 · TaskUpdate 2 · TaskList 1 · TaskGet 1) |

리드가 "최다 위반 파일 = `moai/SKILL.md`, 7개 식별자 전부" 라고 건넸다. **직접 재측정해 확인했다** — `moai/SKILL.md` 는 7개 식별자를 모두 포함한다.

### ③ "3종이 `${CLAUDE_SKILL_DIR}` 의존" — 단위 불일치 (UNIT MISMATCH). 두 값 모두 유효

| 단위 | 값 | 명령 |
|---|---|---|
| `SKILL.md` 파일만 | **3 스킬** | `grep -rl CLAUDE_SKILL_DIR … --include='SKILL.md'` |
| 스킬 디렉터리 전체 | **4 스킬 / 9 파일 / 46 줄** | `grep -rl CLAUDE_SKILL_DIR …` / `grep -rn … \| wc -l` |

`SKILL.md` 단위(3): `moai` · `moai-workflow-testing` · `moai-domain-svg-infographic`
디렉터리 단위가 추가로 잡는 것(+1): `moai-workflow-project` (`references/navigator.md`, `references/navigator-audit.md`)

```
$ grep -rl 'CLAUDE_SKILL_DIR' internal/template/templates/.claude/skills
moai-domain-svg-infographic/SKILL.md
moai-workflow-project/references/navigator-audit.md
moai-workflow-project/references/navigator.md
moai-workflow-testing/SKILL.md
moai/SKILL.md
moai/workflows/harness-build-entry.md
moai/workflows/harness-builder.md
moai/workflows/project/meta-harness.md
moai/workflows/run/task-decomposition.md
$ grep -rn 'CLAUDE_SKILL_DIR' internal/template/templates/.claude/skills | wc -l
      46
```

**넓은 단위가 판정 단위여야 한다** — 실행이 실제로 깨지는 자리가 그 안에 있다(§3).

**[iter-2] "3 대 4 에 트리 차이가 섞였을 수 있다"는 추측은 반증됐다.** 두 ref 를 같은 두 단위로 각각 읽었다:

```
$ git grep -l 'CLAUDE_SKILL_DIR' 297a21ea7 -- 'internal/template/templates/.claude/skills/*/SKILL.md' | wc -l
       3
$ git grep -l 'CLAUDE_SKILL_DIR' 48239c7dc -- 'internal/template/templates/.claude/skills/*/SKILL.md' | wc -l
       3
$ git grep -l 'CLAUDE_SKILL_DIR' 297a21ea7 -- 'internal/template/templates/.claude/skills/' | wc -l
       9
$ git grep -l 'CLAUDE_SKILL_DIR' 48239c7dc -- 'internal/template/templates/.claude/skills/' | wc -l
       9
```

| | `297a21ea7` | `48239c7dc` |
|---|---|---|
| `SKILL.md` 단위 | 3 | 3 |
| 디렉터리 단위(파일) | 9 | 9 |

**두 단위 모두 두 ref 에서 같다 — 트리는 이 차분에 전혀 기여하지 않는다.** 3 과 4 의 차이는 전적으로 세는 단위의 차이이며, 위의 "단위 불일치, 둘 다 유효" 판정이 정확하다. 이 항목은 정정하지 않는다.

### ④ "에이전트 TOML 11종 전부" — 확인 (CONFIRMED), 그리고 강화됨

```
$ find internal/template/templates/.codex/agents -name '*.toml' | wc -l
      11
$ grep -rlE 'AskUserQuestion|Agent\(|Skill\(' internal/template/templates/.codex/agents --include='*.toml' | wc -l
      11
$ find internal/template/templates/.claude/agents -name '*.md' | wc -l
      11
$ grep -rlE 'AskUserQuestion|Agent\(|Skill\(|TaskCreate|TaskUpdate|TaskList|TaskGet' \
    internal/template/templates/.claude/agents --include='*.md' | wc -l
      11
```

TOML 11종 **전부** Claude 도구명을 담고 있고, 그 출처인 정본 `.md` 11종도 **전부** 담고 있다. 카드가 맞다.

---

## 2. 카드가 놓친 구조 — 미러에는 코덱스판 스킬 본문이 없다

```
$ find internal/template/templates/.codex -mindepth 1 -maxdepth 1
internal/template/templates/.codex/agents
```

`.codex/` 아래에는 `agents/` 하나뿐이다. **`.codex/skills/` 는 존재하지 않는다.**

`internal/template/skill_mirror.go` 의 파일 머리말이 그 이유를 직접 적고 있다 — 스킬은 배포 시점에 `.agents/skills/<name>` 을 `../../.claude/skills/<name>` 를 가리키는 **상대 심볼릭 링크**로 만들어 코덱스에 노출한다(링크 생성이 불가한 플랫폼에서는 실 디렉터리 복사로 폴백). `mirrorLinkTarget` 이 그 경로를 만들고, `WithSkillMirror` 가 `Deploy` 당 한 번 호출한다.

**귀결 — 처방이 갈린다.**

| 축 | 코덱스가 읽는 실물 | 편집 지점 | 폭발 반경 |
|---|---|---|---|
| 스킬 본문 | `.claude/skills/**` 와 **바이트 동일** (심볼릭 링크) | 정본 `.claude/skills/**` 자체 | **넓다 — Claude 쪽 변경이다** |
| 에이전트 | `.codex/agents/moai/*.toml` (생성물) | 생성기 `internal/template/agentemit/` | 좁다 — 코덱스 쪽 산출물이다 |

"코덱스 스킬을 중립화한다"는 문장은 성립하지 않는다. 중립화할 코덱스판 스킬 텍스트가 **없다**. 스킬 축의 편집은 전부 Claude 쪽 정본 편집이며, 카드가 "코덱스 배선 손질"로 읽히도록 쓰여 있는 것은 이 지점에서 틀렸다.

에이전트 축은 진짜 이중 발행이다. `internal/template/agentemit/emit.go` 의 `EmitAll` 이 정본 `.md` 를 파싱해 TOML 을 결정적으로 렌더한다. 다만 `developer_instructions` 본문은 **정본 `.md` 본문을 그대로 옮긴다**(파일 머리말: "the emitter never re-renders the neutral layer"). 즉 생성기는 변환 지점으로 **쓸 수 있으나 현재 의도적으로 안 쓰고 있다.**

---

## 3. `${CLAUDE_SKILL_DIR}` — 깨짐의 두 등급

`$CLAUDE_SKILL_DIR` 은 Claude Code 런타임이 채우는 변수이며, 이 저장소는 그것을 **의도적으로 미치환 통과**시킨다:

```
$ sed -n '38,42p' internal/template/renderer.go
var claudeCodePassthroughTokens = []string{
	"$CLAUDE_PROJECT_DIR",
	"$CLAUDE_SKILL_DIR",
```

`internal/cli/codex_launcher.go` 에는 환경변수를 설정하는 코드가 없다(`grep -n 'Env\|env'` → 출력 없음). 코덱스 아래에서 이 변수는 **미설정**이며 빈 문자열로 전개된다.

### 3-1. 시끄러운 깨짐 (HARD) — 6줄

셸/노드 호출의 인자로 쓰여, 빈 전개가 곧 잘못된 경로가 되고 실행이 실패한다.

```
$ grep -rn 'CLAUDE_SKILL_DIR' internal/template/templates/.claude/skills | grep -E ':(bash|node) '
moai-workflow-project/references/navigator.md:88:bash "${CLAUDE_SKILL_DIR}/scripts/navigator-regen.sh"
moai-domain-svg-infographic/SKILL.md:235:node ${CLAUDE_SKILL_DIR}/scripts/check-svg.mjs diagram.svg          # human-readable diagnostics
moai-domain-svg-infographic/SKILL.md:236:node ${CLAUDE_SKILL_DIR}/scripts/check-svg.mjs diagram.svg --json   # machine-readable
moai-domain-svg-infographic/SKILL.md:237:node ${CLAUDE_SKILL_DIR}/scripts/check-svg.mjs diagram.svg --strict # warnings also fail
moai-domain-svg-infographic/SKILL.md:269:node ${CLAUDE_SKILL_DIR}/scripts/render.mjs diagram.svg --out diagram.png            # 2x default
moai-domain-svg-infographic/SKILL.md:270:node ${CLAUDE_SKILL_DIR}/scripts/render.mjs diagram.svg --out diagram.png --scale 3
```

**6줄** / 3개 스크립트 대상 / 2개 스킬.

리드는 이 집합을 "4 sites" 로 건넸다. **줄 단위로는 6이다** — 리드가 든 파일 위치는 3곳이고, 그중 두 곳이 여러 줄을 담는다. 어느 쪽이 틀렸다기보다 단위가 다르다: 파일 위치 3 / 줄 6 / 스크립트 대상 3. 판정에 쓸 단위는 **줄 6** 이다(고칠 대상이 줄이므로).

**[iter-2] 이 차분도 단위 차이이지 트리 차이가 아니다.** 다만 리드의 `4` 는 철회된 트리에서 온 값이므로 **대조 대상으로 세우지 않는다** — 위 6줄은 본 SPEC 이 이 트리에서 잰 값이고, 그것이 판정 단위다.

### 3-2. 조용한 깨짐 (SOFT) — 40줄

나머지 40줄(= 46 − 6)은 산문 경로 지시다. 대표형:

```
moai/SKILL.md:125:For detailed orchestration: Read ${CLAUDE_SKILL_DIR}/workflows/plan.md
```

`moai/SKILL.md` 안에만 이런 `Read ${CLAUDE_SKILL_DIR}/workflows/*.md` 줄이 **19줄**(125~274행 구간) 있다. 여기서 변수가 비면 실행이 실패하지 않는다 — 모델이 추론으로 메운다. 산출물은 plan-phase 를 돈 것처럼 보이지만 실제로는 워크플로 본문을 읽지 않은 것이다.

---

## 4. 선행 카드·SPEC 상태

- 카드 t88 은 큐에 없다(종결됨).
- 선행 `SPEC-CODEX-*` 는 **9건, 전부 `status: completed`**.

**[iter-2 최종 정정] 이 자리에 있던 워킹트리 글롭은 제거했다 — 인쇄된 출력을 더 이상 만들지 않기 때문이다.**

종전 판본은 `grep -m1 '^status:' .moai/specs/SPEC-CODEX-*/spec.md` 를 출력 9줄과 함께 인쇄했다. 그 명령은 **지금 10줄을 낸다** — 열 번째가 이 SPEC 자신(`SPEC-CODEX-SKILL-NEUTRAL-001`, `status: draft`)이다. 글롭이 세는 집합에 세는 주체가 들어갔고, 그 일은 이 보고서를 쓴 뒤 `spec.md` 를 쓰는 순간 일어났다. **인쇄된 명령 옆에 그 명령이 내지 않는 출력을 남겨 두지 않는다.**

근거는 ref 판독으로 대체한다 — 이 형태는 **자기포함에 구조적으로 면역**이다. 이 SPEC 의 디렉터리는 미추적이므로 어떤 ref 에도 들어가지 않는다:

```
$ git ls-tree -d --name-only origin/develop -- .moai/specs/ | grep -c CODEX
9
$ git ls-tree -d --name-only 297a21ea7    -- .moai/specs/ | grep -c CODEX
9
```

선행 9건의 `status` 는 각 `spec.md` 를 직접 읽어 확인했으며 전부 `completed` 다: `DUAL-AGENTS-001` · `HOOK-ADAPTER-001` · `INIT-001` · `LAUNCHER-001` · `PHASE2-001` · `SESSION-MSG-001` · `SKILLS-CANONICAL-001` · `VERDICT-SYNTH-001` · `WIRING-001`.

**[iter-2 정정] 이것이 리드 값과 실제로 갈린 유일한 자리이고, 여기서는 9 가 정본이다.**

종전 서술은 "두 값은 각각 자기 트리에서 참이다"였다. 형식상 맞지만 **오해를 부르는 문장이었다** — 두 값이 대등한 것처럼 읽힌다. 한쪽 트리는 배포선이 아니다.

ref 에 직접 대고 읽었다:

```
$ git ls-tree -d --name-only origin/develop -- .moai/specs/ | grep -c CODEX
9
$ git ls-tree -d --name-only 297a21ea7    -- .moai/specs/ | grep -c CODEX
9
$ git ls-tree -d --name-only 48239c7dc    -- .moai/specs/ | grep -c CODEX
7
```

`48239c7dc` 는 `origin/develop` 보다 686커밋 뒤진 **엄격한 조상**이다(§머리말). 7 은 그 낡은 지점의 값이고, 배포선과 본 SPEC 의 베이스가 모두 **9** 다. **9 가 정본이며 7 은 채택하지 않는다.**

리드는 이 값을 철회했다. 갈린 사실은 기록으로 남기되 대등한 두 값으로 제시하지 않는다.

---

## 5. 설계에 쓰이는 부수 실측

### 5-1. 중립 어휘는 이미 하나 존재한다 (재사용 근거)

`internal/template/agentemit/agents-codex.yaml` 의 `tool_classes` 가 Claude 도구 토큰을 하네스 중립 클래스로 이미 매핑하고 있다:

```
Read: file-read      Write: file-write     Bash: shell
Grep: file-read      Edit: file-write      WebFetch: web
Glob: file-read                            WebSearch: web
TaskCreate/TaskUpdate/TaskList/TaskGet: task-list
Skill: skill-loader
Agent: subagent-spawn
DesignSync: design-sync
```

다만 이 어휘는 **`tools:` 프론트매터에만** 걸린다. 본문 산문은 이 매핑을 통과하지 않는다 — 그것이 정확히 이 카드가 남긴 구멍이다.

### 5-2. `AGENTS.md` 는 어휘를 이미 선언했고, 담을 여유도 있다

```
$ grep -n 'question channel' AGENTS.md
14:mechanisms (the question channel, subagent spawning, skills, session handoff) stay there.
```

`AGENTS.md` 는 이미 능력을 중립 이름으로 부르고("question channel", "subagent spawning") 그 구현을 `CLAUDE.md` 로 미룬다. 없는 것은 **하네스가 그 능력을 못 가졌을 때 무엇을 하는가**를 적은 결속표다.

용량:

```
$ wc -c AGENTS.md internal/template/templates/AGENTS.md
   14229 AGENTS.md
   14229 internal/template/templates/AGENTS.md
$ grep -n 'CodexContractByteCeiling *=' internal/config/token_budget_guard.go
41:const CodexContractByteCeiling = 24576
```

여유 **10,347 바이트**. `TestCodexContractByteCeiling` 이 두 사본 모두를 실 트리에서 감시하며, 초과 시 빌드를 실패시킨다(코덱스가 꼬리를 조용히 자르므로 다른 신호가 없다).

### 5-3. 미러가 심볼릭 링크라는 사실이 경로 축의 해답을 준다

미러가 `.agents/skills/<name>` → `../../.claude/skills/<name>` 이므로, **프로젝트 루트 기준 상대 경로 `.claude/skills/<name>/...` 은 두 하네스에서 동일하게 해석된다.** 코덱스용 대체 환경변수를 새로 만들 필요가 없다 — 변수를 없애는 쪽이 성립한다.

이 값은 측정된 사실(`skill_mirror.go` 의 `mirrorLinkTarget`)에서 나온 것이지 추정이 아니다. 다만 **복사 폴백 모드에서도 성립하는지는 이 보고서가 관측하지 않았다** — run-phase 가 판정할 몫으로 남긴다(§Gaps).

---

## 6. 형제 카드 표면 겹침

카드 t391(`moai codex` 런처 동작)이 lane-3 에서 돌고 있고 같은 `.codex/` 배선을 읽는다. 겹치는 표면은 `internal/cli/codex_launcher.go` 다 — 본 카드는 **거기에 손대지 않는다**(§5-3 이 환경변수 신설을 불필요하게 만들었으므로 겹침이 설계상 해소된다). 표면 겹침 사실만 기록하고 그것을 피해 설계하지 않았다.

---

## Gaps — 관측하지 않은 것

- 코덱스 CLI 를 **실행해서** 빈 `${CLAUDE_SKILL_DIR}` 의 실제 거동을 확인하지 않았다. 미설정이라는 판단은 코드(`renderer.go` 통과 목록 + `codex_launcher.go` 환경변수 부재)에서 나온 것이지 런타임 관측이 아니다.
- 복사 폴백 모드(`MirrorModeCopy`)에서 `.claude/skills/<name>/...` 상대 경로가 성립하는지 관측하지 않았다.
- 코덱스가 `AskUserQuestion` 같은 이름을 만났을 때 **무엇을 하는지** 관측하지 않았다. "조용히 개선한다"는 실패 모드는 추론이며, 관측된 사실이 아니다. 이 추론이 요구사항의 우선순위를 정하고 있으므로 run-phase 에서 최소 1건 관측으로 확인해야 한다.
- 로컬 `.claude/skills/` (44종, dev-only `hns-*` 10종 포함)는 배포 대상이 아니므로 세지 않았다. 본 보고서의 모든 스킬 수치는 **템플릿 배포 대상** 단위다.

## Residual-risk — 관측했음에도 틀릴 수 있는 것

- `grep -rlE 'Agent\('` 은 산문 안의 `Agent(` 를 도구 호출로 세지만, `Agent(general-purpose)` 처럼 **역할 서술**로 쓰인 자리도 함께 센다. 14 라는 파일 수는 상한이며, 실제로 "코덱스에서 실행 불가한 지시"인 자리는 그보다 적을 수 있다. run-phase 는 파일 수가 아니라 **자리별**로 판정해야 한다.
- 46 / 6 / 40 이라는 줄 수는 이 트리의 현재 상태다. `.claude/skills/**` 는 활발히 편집되는 표면이므로 run-phase 진입 시 재측정 없이 이 값을 인용하면 스테일이다.

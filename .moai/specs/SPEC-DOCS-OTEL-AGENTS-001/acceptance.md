---
id: SPEC-DOCS-OTEL-AGENTS-001
acceptance_version: "0.2.0"
created: 2026-09-25
updated: 2026-09-25
author: manager-spec
---

# Acceptance — SPEC-DOCS-OTEL-AGENTS-001

## §A 검증 모델

docs 전용 SPEC 이다. `go test` 는 이 편집에 대해 아무것도 증명하지 않으므로 게이트가 아니다. 모든 AC 는 실행 가능한 명령과 기대 출력으로 판정한다. 부재를 단언하는 AC 에는 양성 대조를 붙인다 — 같은 패턴이 base `a520187f1` 에서는 실제로 적중함을 `git show a520187f1:<path>` 로 보여, 0건이 패턴의 실명이 아니라 편집의 결과임을 확인한다.

공통 정의(모든 AC 가 같은 셸에서 쓴다):

- `$SCRATCH` — 세션 스크래치 디렉터리.
- `$C` — 카드 커밋. `C=$(git log --no-merges --format=%H --grep=t1181 a520187f1..HEAD -- docs-site)`. develop 흡수 뒤에도 병합 커밋과 다른 카드의 커밋을 배제하고 이 카드의 docs-site 커밋만 가리킨다. AC-DOA-008 이 이 값이 정확히 한 개임을 먼저 단언한다.
- `after <key>` — plan.md M1 의 `<!-- after:<key> -->` 표지 뒤 ```` ```text ```` 펜스 안 한 줄을 뽑는 헬퍼:

  ```bash
  after() { awk -v k="$1" '$0=="<!-- after:" k " -->"{f=1;next} f&&/^```text$/{g=1;next} g&&/^```$/{exit} g{print}' .moai/specs/SPEC-DOCS-OTEL-AGENTS-001/plan.md; }
  ```

## §D AC Matrix

### AC-DOA-001 — 텔레메트리 행 제거, 인접 행 보존 (blocks)

- **Given** base 에서 `git grep -c 'CLAUDE_CODE_ENABLE_TELEMETRY' a520187f1 -- docs-site/content` 가 en/ja/zh `advanced/settings-json.md` 각 `:1` 을 출력하는 상태(양성 대조),
- **When** run-phase 커밋 `$C` 가 들어가면,
- **Then** `grep -rn 'CLAUDE_CODE_ENABLE_TELEMETRY\|OTEL_' docs-site/content` 는 출력 없이 exit 1 이고, `git show --numstat --format= "$C" -- docs-site/content/{en,ja,zh}/advanced/settings-json.md` 가 세 파일 각각 `0	1` 을 출력한다. 카드 커밋이 이 세 파일에서 지운 것이 텔레메트리 행 한 줄뿐이므로 앞뒤 행(`CLAUDE_AUTOCOMPACT_PCT_OVERRIDE`, `CLAUDE_CODE_DISABLE_BACKGROUND_TASKS`)이 보존됐음도 함께 증명된다. (maps REQ-DOA-001)

### AC-DOA-002 — `/agents` 표 행 정정, 옛 서술자 제거 (blocks)

- **Given** 네 로케일 `claude-code/foundations/commands.md` 의 `/agents` 표 행과, base 에서 옛 서술자가 각 파일에 한 번씩 있는 상태 — `git show a520187f1:<file> | grep -cF '<옛 서술자>'` → `1`(양성 대조). 옛 서술자: ko `서브에이전트 관리 (v2.1.198`, en `Manage subagent configuration`, ja `サブエージェント管理 UI`, zh `子智能体管理 UI`,
- **When** 편집이 들어가면,
- **Then** 각 파일에서 ``grep '^| `/agents`' <file> | grep -c 'v2.1.281'`` → `1`, 파이프 끝을 `grep -c 'v2.1.198'` 로 바꾸면 → `1`, `grep -c '/help'` 로 바꾸면 → `1` 이고(대조: base 의 `v2.1.281`·`/help` 는 네 파일 모두 `0`), 작업 트리에서 `grep -cF '<옛 서술자>' <file>` → `0` 이다. (maps REQ-DOA-002)

### AC-DOA-003 — 낡은 2026-07 단서 제거 (blocks)

- **Given** base 에서 `git show a520187f1:<file> | grep '/agents' | grep -c '2026-07'` 가 en/ja/zh `commands.md` 와 ja/zh `sub-agents.md` 다섯 파일 각각 `1` 인 상태(양성 대조),
- **When** 편집이 들어가면,
- **Then** 같은 다섯 파일에서 `grep '/agents' <file> | grep -c '2026-07'` → `0` 이다. (maps REQ-DOA-003)

### AC-DOA-004 — ko/en `/agents` 절 제목·첫 문단 정정 (blocks)

- **Given** base 에서 ko `commands.md` 의 `grep -c '^### /agents — 서브에이전트 관리$'` → `1`, `grep -c '살펴보는 명령'` → `1`, en `commands.md` 의 `grep -c '^### /agents — Managing Subagents$'` → `1`, `grep -c 'a command to inspect'` → `1` 인 상태(각각 `git show a520187f1:<file> | …` 로 재는 양성 대조),
- **When** 편집이 들어가면,
- **Then** 작업 트리에서 네 grep 이 각각 `0` 이고, ko/en 파일에서 `grep -A2 '^### /agents' <file> | grep -c 'v2.1.281'` → `1` 이다(대조: base `0`). (maps REQ-DOA-004)

### AC-DOA-005 — ko/en 절의 보존 대상 유지 (blocks)

- **Given** ko/en `commands.md` 의 `### /agents` 절,
- **When** 편집이 들어가면,
- **Then** ko 에서 `grep -c '폴더에 마크다운 파일을 직접 만들기'` → `1`, en 에서 `grep -c 'Create a markdown file directly under'` → `1`, 두 파일 모두 `grep -c 'CLAUDE_CODE_MAX_SUBAGENT_SPAWN_DEPTH=1'` → `1`, `grep -c '^description:.*/agents'` → `1` 이다(결정 D2 — description 유지). 이 절에서 바뀐 행이 표 행·제목·문단 셋뿐이라는 것은 AC-DOA-008 의 커밋 단위 numstat(`3	3`)과 AC-DOA-007 의 행 전체 대조가 함께 증명한다. (maps REQ-DOA-004)

### AC-DOA-006 — ja/zh 서브에이전트 문장 정렬 (blocks)

- **Given** base 에서 `git show a520187f1:docs-site/content/ja/claude-code/agentic/sub-agents.md | grep -c '対話的に生成'` → `1`, zh 는 `grep -c '交互式生成'` → `1` 이고, 두 파일 각각 `git show a520187f1:<file> | grep '/agents' | grep -c 'v2.1.198'` → `0` 인 상태(양성 대조 — base 의 해당 문장은 `v` 없는 `CC 2.1.198` 을 쓴다),
- **When** 편집이 들어가면,
- **Then** 작업 트리에서 앞의 두 grep 이 각각 `0` 이고, 두 파일 각각 `grep '/agents' <file> | grep -c 'v2.1.198'` → `1` 이다. 계수를 `/agents` 가 있는 행으로 한정하므로 ja 파일의 다른 네 행에 있는 `v2.1.198`(base 파일 전체 계수 `4`)은 판정에 끼지 않는다. (maps REQ-DOA-005)

### AC-DOA-007 — 확정 문구와 행 단위 정확 일치 (blocks)

- **Given** plan.md M1 의 AFTER 블록 10개(`ko-row`, `en-row`, `ja-row`, `zh-row`, `ko-head`, `ko-para`, `en-head`, `en-para`, `ja-sub`, `zh-sub`)와 §A 의 `after` 헬퍼, 그리고 base 에서 각 블록이 대상 파일에 `0` 번 나타나는 상태(`git show a520187f1:<file> | grep -cxF -e "$(after <key>)"` → `0`, 양성 대조),
- **When** 편집이 들어가면,
- **Then** 각 키에 대해 `after <key> | wc -l` → `1` 이고, 대상 파일에서 `grep -cxF -e "$(after <key>)" <file>` → `1` 이다. 대상 파일: `*-row`·`*-head`·`*-para` 는 해당 로케일 `claude-code/foundations/commands.md`, `*-sub` 는 해당 로케일 `claude-code/agentic/sub-agents.md`. (maps REQ-DOA-002, REQ-DOA-004, REQ-DOA-005)

### AC-DOA-008 — 범위와 단일 커밋 (blocks)

- **Given** run-phase 커밋이 들어간 브랜치(develop 흡수 여부와 무관),
- **When** `git log --no-merges --format=%H --grep=t1181 a520187f1..HEAD -- docs-site | wc -l` 과 `git show --numstat --format= "$C" -- docs-site/ | sort -k3` 를 실행하면,
- **Then** 앞의 계수는 `1` 이고, 뒤의 출력은 정확히 아래 아홉 행이다(열 구분은 탭).

  ```text
  0	1	docs-site/content/en/advanced/settings-json.md
  3	3	docs-site/content/en/claude-code/foundations/commands.md
  0	1	docs-site/content/ja/advanced/settings-json.md
  1	1	docs-site/content/ja/claude-code/agentic/sub-agents.md
  1	1	docs-site/content/ja/claude-code/foundations/commands.md
  3	3	docs-site/content/ko/claude-code/foundations/commands.md
  0	1	docs-site/content/zh/advanced/settings-json.md
  1	1	docs-site/content/zh/claude-code/agentic/sub-agents.md
  1	1	docs-site/content/zh/claude-code/foundations/commands.md
  ```

  그리고 `git log -1 --format=%B "$C" | grep -c t1181` → `1` 이상이다. (maps REQ-DOA-006, REQ-DOA-007)

### AC-DOA-009 — 경고 없는 Hugo 빌드, 배포·push 없음 (blocks)

- **Given** base 기준선(이 SPEC 작성 실행에서 측정: exit 0, `WARN` 0건),
- **When** `hugo --source docs-site --minify --gc --destination "$SCRATCH/hugo-after" > "$SCRATCH/hugo-after.log" 2>&1; echo "exit=$?"` 를 실행하면,
- **Then** `exit=0`, `grep -c WARN "$SCRATCH/hugo-after.log"` → `0`, `git status --short docs-site` → 빈 출력(빌드가 트리에 쓰지 않음), `git ls-remote --heads origin WT-docs-otel-env` → 빈 출력(push 없음). (maps REQ-DOA-006, REQ-DOA-007)

## §E 엣지 케이스

- develop 흡수로 행 번호가 움직이거나 develop 이 docs-site 의 다른 파일을 바꾼 경우: AC-DOA-001·007·008 은 카드 커밋 `$C` 단위(`git show`)로, 나머지는 행 번호가 아니라 내용으로 판정하므로 영향이 없다. 흡수된 develop 이 대상 아홉 파일 중 하나를 바꿨다면 AC-DOA-002~007 의 작업 트리 계수는 그 변경까지 포함해 읽히므로, 그때는 흡수 전후 값을 함께 기록한다.
- `WARN` 0건은 경고가 생길 수 있는 조건을 이 실행이 만들어 보인 것이 아니다 — 경고 검출 자체의 양성 대조는 없다(아래 Gaps 로 보고).

## §F Definition of Done

- AC-DOA-001~009 전부 PASS, 명령과 출력 원문이 `progress.md` §E.2 와 `.moai/reports/t1181/verdict.md` 에 실림.
- 아홉 파일 한 커밋, 배포·push 없음.
- 보고서의 Gaps 에 다음을 적는다: "Hugo `WARN` 검출기의 양성 대조 없음", "렌더된 HTML 을 눈으로 확인하지 않았다면 그 사실", "배포 부재는 push 부재로 대리 판정했다 — push 가 없으면 Vercel 빌드 트리거도 없다고 간주한다".

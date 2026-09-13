# t523 — 71건 전수 분류표 (12개 미러 스킬 디렉터리)

card: t523 · tree: `.claude/worktrees/t523` · branch `WT-mirror-skill-neutrality` · base `7e1859c28`
분류 기준: SPEC-CODEX-BODY-NEUTRALITY-001 REQ-CBN-002 논리(지시의 주어가 읽는 에이전트 자신이면 지시, 오케스트레이터·타 시스템이면 산문) + 착지된 `templates/AGENTS.md` 중립 결속표 어휘.

## 분포 재측정 (이 트리, 이 실행)

```
$ grep -rhoE 'AskUserQuestion|Agent\(|Task(Create|Update|List|Get)|DesignSync|Skill\(' <dir>/ | wc -l
moai 413 · moai-foundation-cc 120 · moai-foundation-core 105   (상위 3본 = 638)
12개 디렉터리 합계 = 71   ← 본 카드 작업 집합
```

t497 측정(ace1c5440, 719/639) 대비 `moai` 디렉터리 1건 감소(414→413). 12디렉터리 71건은 불변.

## 표본 확인 — 상위 3본 제외 전제에 대한 판정 (§B.6 선행 조건)

| 디렉터리 | 표본 | 판정 |
|---|---|---|
| moai-foundation-cc (120) | 5줄 (SKILL.md:66, reference.md:149, claude-code-custom-slash-commands:664, sub-agent-integration:14, sub-agent-examples:251) | **전부 산문** — Claude Code 공식 문서 계열 참조물. 제외 지지 |
| moai-foundation-core (105) | 3줄 (SKILL.md:129, examples.md:309, execution-rules.md:14) | **산문(교육 콘텐츠), 단 지시 인접** — "ALWAYS delegate via Agent(), NEVER execute directly" 같은 원칙 진술이 행위 규칙 형태. 주어는 MoAI 원칙이라 산문 판정하나 경계 |
| moai (413) | 15줄 (workflows/ + SKILL.md 계통 표본) | **부분 반증** — 실제 행위 지시 다수 확인: `loop.md:194`("AskUserQuestion required") · `mode-detection.md:64`("[HARD] Present detection result via AskUserQuestion") · `run/phase-execution.md:361`("execute TaskCreate") · `harness.md:190`(ToolSearch→AskUserQuestion) · `codemaps.md:234`("Delegate ... the single Agent() spawn") 외. `SKILL.md:8` `allowed-tools:` 프론트매터(하네스 메타데이터)도 존재 |

→ 「상위 3본 = 전부 산문」 전제는 moai-foundation-cc에선 성립, moai 디렉터리에선 **성립하지 않는다**.
  본 카드의 범위(12개/71건)는 운영자가 확정한 것이므로 그대로 수행하되, moai 디렉터리 413건의
  지시 전수 분류는 **후속 카드 후보**로 리드·운영자에게 보고한다.

## 71건 전수 분류

### 중립화 (지시 → 능력 클래스 어휘로 개정, 13건)

| 파일 | 행 | 원문 → 개정 |
|---|---|---|
| moai-foundation-thinking/modules/first-principles.md | 79 | "Use AskUserQuestion to clarify..." → "use the harness's `question-channel` capability ... (where the harness lacks it, name the open question in the report instead of asking)" |
| 〃 | 127 | "Use AskUserQuestion to explore..." → "Use the question channel to explore..." |
| 〃 | 132-135 | 4개 불릿 "Use AskUserQuestion to ..." → "Use the question channel to ..." |
| moai-foundation-thinking/modules/trade-off-analysis.md | 46 | "Use AskUserQuestion to understand user priorities" → "Use the harness's `question-channel` capability ... (결속표 대체 행동 병기)" |
| 〃 | 136-139 | 4개 불릿 "Use AskUserQuestion to ..." → "Use the question channel to ..." |
| moai-foundation-thinking/modules/assumption-matrix.md | 71 | "use AskUserQuestion to:" → "use the harness's `question-channel` capability to (...):" |
| moai-workflow-spec/references/requirement-clarification.md | 7 | "via AskUserQuestion" → "via the harness's `question-channel` capability (...)" |

개정 문체는 같은 트리에 착지된 상위 SPEC의 산출물(manager-develop.md:104 "through the
harness's `task-list` capability. A harness with no `task-list` records ... as prose")의
결을 따른다. 어휘는 `agents-codex.yaml` tool_classes의 `question-channel`(REQ-CBN-005 —
새 어휘 금지), 대체 행동은 `templates/AGENTS.md` 결속표 question-channel 행의 것.

### 산문 — 보존 (57건)

| 디렉터리 | 건수 | 대표 근거 |
|---|---|---|
| moai-foundation-thinking | 9 | 절제목 3("## Integration with AskUserQuestion": fp:129·toa:133·am:69 — REQ-CBN-010 논리: 산문 문면 불변), 예시 서술 5(philosopher-examples: "AskUserQuestion applied:", "confirmed via..."), 예시 프레임 1(toa:141) |
| moai-meta-harness | 14 | 48·91·92는 **v4 Builder(주어=Claude 오케스트레이터 세션)의 내부 동작 묘사** — 읽는 에이전트에 내리는 지시가 아니라 재분류(초기 지시 판정 정정). 나머지 11건: 보존 에이전트 나열·@MX 주석·교차참조 |
| moai-harness-learner | 11 | 전부 주어=오케스트레이터("the orchestrator surfaces them via AskUserQuestion") |
| moai-workflow-spec | 9 | "Orchestrator runs AskUserQuestion rounds"(주어=오케스트레이터), "Assigned: Agent(general-purpose)" 참조표, worktree-runtime 서술 |
| moai-ref-api-patterns / moai-ref-react-patterns | 3+3 | "spawned via Agent(general-purpose)" — 스킬 적용 대상 묘사 |
| moai-workflow-project | 2 | "the orchestrator passes...", JSON 스키마 서술 |
| moai-ref-ui-polish | 2 | "equivalently available as a per-spawn Agent(general-purpose)" — 호출 면 묘사 |
| moai-kanban-foreman | 2 | **:18 `disallowed-tools:` 프론트매터는 하네스 메타데이터 — 제거 시 Claude 동작이 바뀌므로 보존**(행동 보존), :40은 그 설명 산문 |
| moai-ref-seo | 1 | 호출 면 묘사 |
| moai-ref-owasp-checklist | 1 | 호출 면 묘사 |
| moai-domain-html-report | 1 | "an `Agent()` spawn prompt" — 맥락 열거 |

합계 검산: 중립화 13 + 산문·메타데이터 보존 58 = **71** ✓

## Gaps

- 상위 3본(638건)은 표본(23줄)만 분류했다 — moai 디렉터리의 지시 전수는 이 카드 범위 밖이며
  후속 카드 후보로 보고한다.
- `moai/SKILL.md:8`의 `allowed-tools:` 프론트매터류 하네스 메타데이터의 codex 미러 노출
  적합성은 본 분류 축이 아니다(선언은 지시도 산문도 아님) — 별도 판정 대상으로 기록만 한다.

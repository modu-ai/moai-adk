# t497 M2 — 배포 TOML 본문 84건 전수 분류

card: t497 · tree: `.claude/worktrees/t497` · branch `WT-codex-neutrality`
측정 HEAD: `7b4ba4491` (이 트리, 이 실행) · base `ace1c5440`

SPEC-CODEX-BODY-NEUTRALITY-001 M2 산출. 모집단은 `.codex/agents/moai/*.toml` **본문뿐**이다
(`.claude/agents/moai/*.md` 전수 grep 은 프론트매터 `tools:` CSV 때문에 과다계상 —
`Task*` 가 4 가 아니라 48 로 나온다).

**단위: 발생 84 / 서로 다른 줄 81.** 한 줄이 두 토큰을 싣는 경우가 3줄 있어
(`e2e-tester.toml:139` · `manager-docs.toml:23` · `plan-auditor.toml:146`) 그 3줄은 표에서
2행씩 차지하되 좌표 집합으로는 1개다.

## 모집단 재측정 (이 실행)

```
$ grep -rhoE 'AskUserQuestion|TaskCreate|TaskUpdate|TaskList|TaskGet|DesignSync|Skill\(|Agent\(' \
    internal/template/templates/.codex/agents/moai/*.toml | wc -l
84
$ grep -rnoE '<같은 패턴>' …/*.toml | cut -d: -f1-2 | sort -u | wc -l
81
```

## 판정 규칙 (REQ-CBN-002 · `spec.md` §B.4)

**이 에이전트에게 도구를 호출하라고 지시하는 줄은 `directive`. 그 도구가 무엇인지·누가
쓰는지를 설명하는 줄은 `prose`.** 지시의 주어가 이 에이전트가 아니라 오케스트레이터이면
`prose`(subject=`orchestrator`)다. 어떤 행위도 지시하지 않는 금지 서술은 `prohibition`,
행위 주어 자체가 없는 서술(정의·제목·교차 참조·모드 기술)은 `n/a` 다 —
금지 서술이나 순수 서술을 앞의 두 값 중 하나로 밀어 넣으면 분류가 거짓이 된다.

## 집계

| 축 | 값 |
|---|---|
| verdict | directive **57** · prose **27** (합 84) |
| subject | this-agent **57** · orchestrator **17** · n/a **9** · prohibition **1** |
| directive 토큰별 | `Skill(` 45 · `Agent(` 7 · `TaskUpdate` 3 · `TaskCreate` 1 · `DesignSync` 1 |

**지시 부류는 넷이며 처분이 각각 다르다.**

1. **skill-loader 45건** — 문면을 **손대지 않는다**(REQ-CBN-011). progressive disclosure
   절감이 사라지고, 코덱스 쪽 대체 행동이 실재하기 때문이다(미러가 같은 파일을
   `.agents/skills/<name>/SKILL.md` 로 읽게 한다). 덮개는 `AGENTS.md` 두 사본의 1문장.
2. **task-list 4건 / 3줄** — 도구 이름을 능력 클래스 이름으로 고쳐 쓴다.
3. **subagent-spawn 7건** — `manager-lead.toml` 의 `Agent(` 줄 중 directive 로 판정된 전부.
   능력 이름 + 코덱스 쪽 대체 행동을 본문에 직접 적는다. 결속행은 만들지 않는다
   (`subagent-spawn` 은 능력 **존재**로 확정 — M1).
4. **design-sync 1건** — 사다리 문면은 유지하고 부재 시 행동을 인접 문단으로 붙인다.

## `manager-lead.toml` `Agent(` 모집단 10줄의 전수 판정 (AC-CBN-013 의 N)

이 파일의 `Agent(` 모집단은 **10줄**(좌표 `7 23 29 37 57 59 130 172 193 261`)이고,
그중 directive 는 **N = 7**(`29 37 57 59 172 193 261`), prose 는 3(`7 23 130`)이다.
`spec.md` §B.4 가 경계 표본으로 든 `37 57 59 193` 네 좌표는 **전부** directive 집합 안에 있다.

**SPEC 이 M2 에 판정을 위임한 두 줄의 판정과 근거.**

- **`:172` → directive · this-agent.** "At Tier M/L milestones, every AC … is re-run by a second
  read-only `Agent(general-purpose)`" — `:57` 이 능력 목록으로 적은 행위를 절차 구역에서
  다시 적은 줄이다. 수동태라 주어가 표면에 없지만, 재실행 스폰을 수행하는 주체는 peer
  cross-validation 을 오케스트레이션하는 이 에이전트다. REQ-CBN-002 의 주어 규칙은
  「오케스트레이터가 주어면 prose」이고 여기서 오케스트레이터는 주어가 아니므로 this-agent.
  `:57` 을 directive 로 두고 `:172` 를 prose 로 두면 같은 행위에 두 판정을 주게 된다.
- **`:261` → directive · this-agent.** "Domain consultation … → leaf worker as
  `Agent(general-purpose)` with domain whitelist" — 이 에이전트의 위임 라우팅 목록 항목이고,
  화살표 오른쪽이 이 에이전트가 취할 행동이다. 형태가 닮은 `manager-develop.toml:64-66` 이
  prose 인 것과 갈리는 지점은 **스폰 주체**다: manager-develop 은 `Agent` 도구가 없어 그 표가
  오케스트레이터의 라우팅을 기술하는 데 반해, manager-lead 는 카탈로그에서 유일한
  Agent-carrier 라 같은 문장이 자기 행위 지시로 읽힌다.

`spec.md` §B.4 가 판정하지 않고 남긴 나머지 넷도 여기서 닫는다 — `:7`(보드 기제 서술,
행위자는 plan 레인 세션) · `:23`(Role A/B 대조표의 정의 칸) · `:130`(리프 워커 정의 + CI 가드
사실)은 prose, `:29`(리드 자세 문단의 배경 스폰 지시)는 `:59` 와 같은 행위이므로 directive 다.

## 전수 분류표

| file | line | token | verdict | subject | rationale |
|---|---|---|---|---|---|
| internal/template/templates/.codex/agents/moai/builder-harness.toml | 54 | Agent( | prose | orchestrator | OUT OF SCOPE 라우팅 서술 — 어디로 보내는지의 기술이며 이 에이전트에게 스폰하라고 지시하지 않는다 (builder-harness 는 Agent 도구를 갖지 않는다) |
| internal/template/templates/.codex/agents/moai/builder-harness.toml | 131 | Skill( | directive | this-agent | `Skill(...)` 지시. 이 에이전트에게 스킬을 읽으라고 지시하는 줄이므로 지시다. 처분은 문면 개정이 아니라 AGENTS.md 두 사본의 덮개 1문장 (REQ-CBN-011) |
| internal/template/templates/.codex/agents/moai/builder-harness.toml | 132 | Skill( | directive | this-agent | `Skill(...)` 지시. 이 에이전트에게 스킬을 읽으라고 지시하는 줄이므로 지시다. 처분은 문면 개정이 아니라 AGENTS.md 두 사본의 덮개 1문장 (REQ-CBN-011) |
| internal/template/templates/.codex/agents/moai/e2e-tester.toml | 132 | Skill( | directive | this-agent | `Skill(...)` 지시. 이 에이전트에게 스킬을 읽으라고 지시하는 줄이므로 지시다. 처분은 문면 개정이 아니라 AGENTS.md 두 사본의 덮개 1문장 (REQ-CBN-011) |
| internal/template/templates/.codex/agents/moai/e2e-tester.toml | 133 | Skill( | directive | this-agent | `Skill(...)` 지시. 이 에이전트에게 스킬을 읽으라고 지시하는 줄이므로 지시다. 처분은 문면 개정이 아니라 AGENTS.md 두 사본의 덮개 1문장 (REQ-CBN-011) |
| internal/template/templates/.codex/agents/moai/e2e-tester.toml | 139 | TaskCreate | directive | this-agent | 이 에이전트에게 여정 추적을 도구 이름으로 하라고 지시하는 줄. task-list 클래스로 고쳐 쓴다 |
| internal/template/templates/.codex/agents/moai/e2e-tester.toml | 139 | TaskUpdate | directive | this-agent | 같은 줄의 두 번째 발생. 같은 판정, 같은 처분 |
| internal/template/templates/.codex/agents/moai/manager-design.toml | 24 | DesignSync | prose | n/a | 결합 범위 서술 — 무엇에만 결합하는지의 기술이지 호출 지시가 아니다 |
| internal/template/templates/.codex/agents/moai/manager-design.toml | 27 | DesignSync | prose | n/a | 원격 조작이 어디를 통해 일어나는지의 기술. 행위 지시 없음 |
| internal/template/templates/.codex/agents/moai/manager-design.toml | 109 | DesignSync | directive | this-agent | 우선순위 사다리의 (1) 기본값 — 이 에이전트에게 무엇을 먼저 시도하라고 지시한다. 사다리 문면은 유지하고 design-sync 부재 시 행동을 인접 문단으로 붙인다 (REQ-CBN-009) |
| internal/template/templates/.codex/agents/moai/manager-design.toml | 112 | DesignSync | prose | n/a | 절 제목. 계약 서술 구역의 머리말이며 지시가 아니다 |
| internal/template/templates/.codex/agents/moai/manager-design.toml | 114 | DesignSync | prose | n/a | 계약 결합 범위 서술 — :24 와 같은 형태 |
| internal/template/templates/.codex/agents/moai/manager-design.toml | 132 | DesignSync | prose | n/a | 서버 등록 상태에 대한 사실 서술 |
| internal/template/templates/.codex/agents/moai/manager-design.toml | 191 | Skill( | directive | this-agent | `Skill(...)` 지시. 이 에이전트에게 스킬을 읽으라고 지시하는 줄이므로 지시다. 처분은 문면 개정이 아니라 AGENTS.md 두 사본의 덮개 1문장 (REQ-CBN-011) |
| internal/template/templates/.codex/agents/moai/manager-design.toml | 192 | Skill( | directive | this-agent | `Skill(...)` 지시. 이 에이전트에게 스킬을 읽으라고 지시하는 줄이므로 지시다. 처분은 문면 개정이 아니라 AGENTS.md 두 사본의 덮개 1문장 (REQ-CBN-011) |
| internal/template/templates/.codex/agents/moai/manager-design.toml | 193 | Skill( | directive | this-agent | `Skill(...)` 지시. 이 에이전트에게 스킬을 읽으라고 지시하는 줄이므로 지시다. 처분은 문면 개정이 아니라 AGENTS.md 두 사본의 덮개 1문장 (REQ-CBN-011) |
| internal/template/templates/.codex/agents/moai/manager-design.toml | 194 | Skill( | directive | this-agent | `Skill(...)` 지시. 이 에이전트에게 스킬을 읽으라고 지시하는 줄이므로 지시다. 처분은 문면 개정이 아니라 AGENTS.md 두 사본의 덮개 1문장 (REQ-CBN-011) |
| internal/template/templates/.codex/agents/moai/manager-develop.toml | 9 | Agent( | prose | orchestrator | NOT for 서술의 라우팅 절 — 도메인 작업을 어디로 보내는지의 기술이며 주어가 이 에이전트가 아니다 |
| internal/template/templates/.codex/agents/moai/manager-develop.toml | 64 | Agent( | prose | orchestrator | 위임 라우팅 표의 목적지 칸. 오케스트레이터가 어디로 보내는지의 서술 (spec.md §B.4 판정) |
| internal/template/templates/.codex/agents/moai/manager-develop.toml | 65 | Agent( | prose | orchestrator | 같은 표의 목적지 칸 (spec.md §B.4 판정) |
| internal/template/templates/.codex/agents/moai/manager-develop.toml | 66 | Agent( | prose | orchestrator | 같은 표의 목적지 칸 (spec.md §B.4 판정) |
| internal/template/templates/.codex/agents/moai/manager-develop.toml | 90 | TaskUpdate | directive | this-agent | RED 단계에서 이 에이전트에게 상태 기록을 도구 이름으로 하라고 지시. task-list 클래스로 고쳐 쓴다 |
| internal/template/templates/.codex/agents/moai/manager-develop.toml | 115 | TaskUpdate | directive | this-agent | 변경 루프 5단계에서 이 에이전트에게 태스크 상태 갱신을 도구 이름으로 지시. task-list 클래스로 고쳐 쓴다 |
| internal/template/templates/.codex/agents/moai/manager-develop.toml | 213 | Skill( | directive | this-agent | `Skill(...)` 지시. 이 에이전트에게 스킬을 읽으라고 지시하는 줄이므로 지시다. 처분은 문면 개정이 아니라 AGENTS.md 두 사본의 덮개 1문장 (REQ-CBN-011) |
| internal/template/templates/.codex/agents/moai/manager-develop.toml | 214 | Skill( | directive | this-agent | `Skill(...)` 지시. 이 에이전트에게 스킬을 읽으라고 지시하는 줄이므로 지시다. 처분은 문면 개정이 아니라 AGENTS.md 두 사본의 덮개 1문장 (REQ-CBN-011) |
| internal/template/templates/.codex/agents/moai/manager-develop.toml | 215 | Skill( | directive | this-agent | `Skill(...)` 지시. 이 에이전트에게 스킬을 읽으라고 지시하는 줄이므로 지시다. 처분은 문면 개정이 아니라 AGENTS.md 두 사본의 덮개 1문장 (REQ-CBN-011) |
| internal/template/templates/.codex/agents/moai/manager-develop.toml | 216 | Skill( | directive | this-agent | `Skill(...)` 지시. 이 에이전트에게 스킬을 읽으라고 지시하는 줄이므로 지시다. 처분은 문면 개정이 아니라 AGENTS.md 두 사본의 덮개 1문장 (REQ-CBN-011) |
| internal/template/templates/.codex/agents/moai/manager-develop.toml | 217 | Skill( | directive | this-agent | `Skill(...)` 지시. 이 에이전트에게 스킬을 읽으라고 지시하는 줄이므로 지시다. 처분은 문면 개정이 아니라 AGENTS.md 두 사본의 덮개 1문장 (REQ-CBN-011) |
| internal/template/templates/.codex/agents/moai/manager-develop.toml | 218 | Skill( | directive | this-agent | `Skill(...)` 지시. 이 에이전트에게 스킬을 읽으라고 지시하는 줄이므로 지시다. 처분은 문면 개정이 아니라 AGENTS.md 두 사본의 덮개 1문장 (REQ-CBN-011) |
| internal/template/templates/.codex/agents/moai/manager-develop.toml | 219 | Skill( | directive | this-agent | `Skill(...)` 지시. 이 에이전트에게 스킬을 읽으라고 지시하는 줄이므로 지시다. 처분은 문면 개정이 아니라 AGENTS.md 두 사본의 덮개 1문장 (REQ-CBN-011) |
| internal/template/templates/.codex/agents/moai/manager-develop.toml | 220 | Skill( | directive | this-agent | `Skill(...)` 지시. 이 에이전트에게 스킬을 읽으라고 지시하는 줄이므로 지시다. 처분은 문면 개정이 아니라 AGENTS.md 두 사본의 덮개 1문장 (REQ-CBN-011) |
| internal/template/templates/.codex/agents/moai/manager-docs.toml | 21 | Skill( | directive | this-agent | `Skill(...)` 지시. 이 에이전트에게 스킬을 읽으라고 지시하는 줄이므로 지시다. 처분은 문면 개정이 아니라 AGENTS.md 두 사본의 덮개 1문장 (REQ-CBN-011) |
| internal/template/templates/.codex/agents/moai/manager-docs.toml | 23 | Agent( | prose | orchestrator | OUT OF SCOPE 라우팅 — 구현/배포/보안을 어디로 보내는지의 서술 |
| internal/template/templates/.codex/agents/moai/manager-docs.toml | 23 | Agent( | prose | orchestrator | 같은 줄의 두 번째 발생 — 디자인 시스템 문서의 협업 상대를 지목하는 서술 |
| internal/template/templates/.codex/agents/moai/manager-docs.toml | 187 | Skill( | directive | this-agent | `Skill(...)` 지시. 이 에이전트에게 스킬을 읽으라고 지시하는 줄이므로 지시다. 처분은 문면 개정이 아니라 AGENTS.md 두 사본의 덮개 1문장 (REQ-CBN-011) |
| internal/template/templates/.codex/agents/moai/manager-docs.toml | 188 | Skill( | directive | this-agent | `Skill(...)` 지시. 이 에이전트에게 스킬을 읽으라고 지시하는 줄이므로 지시다. 처분은 문면 개정이 아니라 AGENTS.md 두 사본의 덮개 1문장 (REQ-CBN-011) |
| internal/template/templates/.codex/agents/moai/manager-docs.toml | 189 | Skill( | directive | this-agent | `Skill(...)` 지시. 이 에이전트에게 스킬을 읽으라고 지시하는 줄이므로 지시다. 처분은 문면 개정이 아니라 AGENTS.md 두 사본의 덮개 1문장 (REQ-CBN-011) |
| internal/template/templates/.codex/agents/moai/manager-docs.toml | 190 | Skill( | directive | this-agent | `Skill(...)` 지시. 이 에이전트에게 스킬을 읽으라고 지시하는 줄이므로 지시다. 처분은 문면 개정이 아니라 AGENTS.md 두 사본의 덮개 1문장 (REQ-CBN-011) |
| internal/template/templates/.codex/agents/moai/manager-docs.toml | 191 | Skill( | directive | this-agent | `Skill(...)` 지시. 이 에이전트에게 스킬을 읽으라고 지시하는 줄이므로 지시다. 처분은 문면 개정이 아니라 AGENTS.md 두 사본의 덮개 1문장 (REQ-CBN-011) |
| internal/template/templates/.codex/agents/moai/manager-git.toml | 176 | Skill( | directive | this-agent | `Skill(...)` 지시. 이 에이전트에게 스킬을 읽으라고 지시하는 줄이므로 지시다. 처분은 문면 개정이 아니라 AGENTS.md 두 사본의 덮개 1문장 (REQ-CBN-011) |
| internal/template/templates/.codex/agents/moai/manager-git.toml | 177 | Skill( | directive | this-agent | `Skill(...)` 지시. 이 에이전트에게 스킬을 읽으라고 지시하는 줄이므로 지시다. 처분은 문면 개정이 아니라 AGENTS.md 두 사본의 덮개 1문장 (REQ-CBN-011) |
| internal/template/templates/.codex/agents/moai/manager-git.toml | 178 | Skill( | directive | this-agent | `Skill(...)` 지시. 이 에이전트에게 스킬을 읽으라고 지시하는 줄이므로 지시다. 처분은 문면 개정이 아니라 AGENTS.md 두 사본의 덮개 1문장 (REQ-CBN-011) |
| internal/template/templates/.codex/agents/moai/manager-git.toml | 179 | Skill( | directive | this-agent | `Skill(...)` 지시. 이 에이전트에게 스킬을 읽으라고 지시하는 줄이므로 지시다. 처분은 문면 개정이 아니라 AGENTS.md 두 사본의 덮개 1문장 (REQ-CBN-011) |
| internal/template/templates/.codex/agents/moai/manager-git.toml | 180 | Skill( | directive | this-agent | `Skill(...)` 지시. 이 에이전트에게 스킬을 읽으라고 지시하는 줄이므로 지시다. 처분은 문면 개정이 아니라 AGENTS.md 두 사본의 덮개 1문장 (REQ-CBN-011) |
| internal/template/templates/.codex/agents/moai/manager-lead.toml | 7 | Agent( | prose | n/a | -k/-f 보드 기제 서술. 팬아웃 행위자는 운영자가 띄운 plan 레인 세션이지 이 에이전트가 아니다 |
| internal/template/templates/.codex/agents/moai/manager-lead.toml | 23 | Agent( | prose | n/a | Role A/B 대조표의 정의 칸 — Workers 가 무엇인지의 정의이며 스폰 지시가 아니다 |
| internal/template/templates/.codex/agents/moai/manager-lead.toml | 29 | Agent( | directive | this-agent | 리드 자세 문단이 이 에이전트에게 병렬 작업을 배경 스폰으로 던지라고 지시한다 (GLM 위험 조항이 스폰 방법까지 못박는다). :59 와 같은 행위이므로 같은 판정 |
| internal/template/templates/.codex/agents/moai/manager-lead.toml | 37 | Agent( | directive | this-agent | 이 에이전트가 자기 행위로 리프 워커를 스폰하라는 지시 (spec.md §B.4 경계 표본) |
| internal/template/templates/.codex/agents/moai/manager-lead.toml | 57 | Agent( | directive | this-agent | AC PASS 재검증을 위해 이 에이전트가 두 번째 읽기 전용 워커를 스폰하라는 지시 (경계 표본) |
| internal/template/templates/.codex/agents/moai/manager-lead.toml | 59 | Agent( | directive | this-agent | 병렬 작업을 배경 스폰으로 배차하라는 지시 (경계 표본) |
| internal/template/templates/.codex/agents/moai/manager-lead.toml | 130 | Agent( | prose | n/a | 리프 워커가 무엇인지의 정의 + CI 가드 사실. 스폰 지시가 아니다 |
| internal/template/templates/.codex/agents/moai/manager-lead.toml | 172 | Agent( | directive | this-agent | **M2 판정** — :57 이 능력 목록으로 적은 행위를 절차로 다시 적은 줄. 재실행 스폰의 행위자는 이 에이전트이며, 수동태라도 REQ-CBN-002 의 주어 규칙상 orchestrator 가 아니다 → this-agent |
| internal/template/templates/.codex/agents/moai/manager-lead.toml | 193 | Agent( | directive | this-agent | 리드 세션이 부관을 배경 스폰으로 띄우라는 지시 (경계 표본) |
| internal/template/templates/.codex/agents/moai/manager-lead.toml | 261 | Agent( | directive | this-agent | **M2 판정** — 이 에이전트의 위임 라우팅 목록에서 도메인 자문을 리프 워커 스폰으로 처리하라는 지시. manager-develop:64-66 과 달리 이 에이전트가 스폰 주체다 → this-agent |
| internal/template/templates/.codex/agents/moai/manager-lead.toml | 275 | Skill( | directive | this-agent | `Skill(...)` 지시. 이 에이전트에게 스킬을 읽으라고 지시하는 줄이므로 지시다. 처분은 문면 개정이 아니라 AGENTS.md 두 사본의 덮개 1문장 (REQ-CBN-011) |
| internal/template/templates/.codex/agents/moai/manager-lead.toml | 276 | Skill( | directive | this-agent | `Skill(...)` 지시. 이 에이전트에게 스킬을 읽으라고 지시하는 줄이므로 지시다. 처분은 문면 개정이 아니라 AGENTS.md 두 사본의 덮개 1문장 (REQ-CBN-011) |
| internal/template/templates/.codex/agents/moai/manager-lead.toml | 277 | Skill( | directive | this-agent | `Skill(...)` 지시. 이 에이전트에게 스킬을 읽으라고 지시하는 줄이므로 지시다. 처분은 문면 개정이 아니라 AGENTS.md 두 사본의 덮개 1문장 (REQ-CBN-011) |
| internal/template/templates/.codex/agents/moai/manager-spec.toml | 29 | Skill( | directive | this-agent | `Skill(...)` 지시. 이 에이전트에게 스킬을 읽으라고 지시하는 줄이므로 지시다. 처분은 문면 개정이 아니라 AGENTS.md 두 사본의 덮개 1문장 (REQ-CBN-011) |
| internal/template/templates/.codex/agents/moai/manager-spec.toml | 43 | Agent( | prose | orchestrator | 동사가 recommend 다 — 이 에이전트는 스폰하지 않고 오케스트레이터에게 권고한다 |
| internal/template/templates/.codex/agents/moai/manager-spec.toml | 159 | Agent( | prose | orchestrator | 동사가 recommend 다 — 스폰 주체는 오케스트레이터이며 이 에이전트는 키워드로 후보를 지목할 뿐이다 |
| internal/template/templates/.codex/agents/moai/manager-spec.toml | 160 | Agent( | prose | orchestrator | 동사가 recommend 다 — 스폰 주체는 오케스트레이터이며 이 에이전트는 키워드로 후보를 지목할 뿐이다 |
| internal/template/templates/.codex/agents/moai/manager-spec.toml | 161 | Agent( | prose | orchestrator | 동사가 recommend 다 — 스폰 주체는 오케스트레이터이며 이 에이전트는 키워드로 후보를 지목할 뿐이다 |
| internal/template/templates/.codex/agents/moai/manager-spec.toml | 162 | Agent( | prose | orchestrator | 동사가 recommend 다 — 스폰 주체는 오케스트레이터이며 이 에이전트는 키워드로 후보를 지목할 뿐이다 |
| internal/template/templates/.codex/agents/moai/manager-spec.toml | 218 | Skill( | directive | this-agent | `Skill(...)` 지시. 이 에이전트에게 스킬을 읽으라고 지시하는 줄이므로 지시다. 처분은 문면 개정이 아니라 AGENTS.md 두 사본의 덮개 1문장 (REQ-CBN-011) |
| internal/template/templates/.codex/agents/moai/manager-spec.toml | 219 | Skill( | directive | this-agent | `Skill(...)` 지시. 이 에이전트에게 스킬을 읽으라고 지시하는 줄이므로 지시다. 처분은 문면 개정이 아니라 AGENTS.md 두 사본의 덮개 1문장 (REQ-CBN-011) |
| internal/template/templates/.codex/agents/moai/manager-spec.toml | 220 | Skill( | directive | this-agent | `Skill(...)` 지시. 이 에이전트에게 스킬을 읽으라고 지시하는 줄이므로 지시다. 처분은 문면 개정이 아니라 AGENTS.md 두 사본의 덮개 1문장 (REQ-CBN-011) |
| internal/template/templates/.codex/agents/moai/manager-spec.toml | 221 | Skill( | directive | this-agent | `Skill(...)` 지시. 이 에이전트에게 스킬을 읽으라고 지시하는 줄이므로 지시다. 처분은 문면 개정이 아니라 AGENTS.md 두 사본의 덮개 1문장 (REQ-CBN-011) |
| internal/template/templates/.codex/agents/moai/manager-spec.toml | 222 | Skill( | directive | this-agent | `Skill(...)` 지시. 이 에이전트에게 스킬을 읽으라고 지시하는 줄이므로 지시다. 처분은 문면 개정이 아니라 AGENTS.md 두 사본의 덮개 1문장 (REQ-CBN-011) |
| internal/template/templates/.codex/agents/moai/manager-spec.toml | 223 | Skill( | directive | this-agent | `Skill(...)` 지시. 이 에이전트에게 스킬을 읽으라고 지시하는 줄이므로 지시다. 처분은 문면 개정이 아니라 AGENTS.md 두 사본의 덮개 1문장 (REQ-CBN-011) |
| internal/template/templates/.codex/agents/moai/manager-spec.toml | 224 | Skill( | directive | this-agent | `Skill(...)` 지시. 이 에이전트에게 스킬을 읽으라고 지시하는 줄이므로 지시다. 처분은 문면 개정이 아니라 AGENTS.md 두 사본의 덮개 1문장 (REQ-CBN-011) |
| internal/template/templates/.codex/agents/moai/plan-auditor.toml | 146 | AskUserQuestion | prose | orchestrator | "The orchestrator MUST resolve each marked topic via AskUserQuestion" — 주어가 이 에이전트가 아니다 (AC-CBN-002) |
| internal/template/templates/.codex/agents/moai/plan-auditor.toml | 146 | AskUserQuestion | prose | orchestrator | 같은 줄의 두 번째 발생 (ToolSearch preload 인용). 같은 판정 (AC-CBN-002) |
| internal/template/templates/.codex/agents/moai/plan-auditor.toml | 501 | Skill( | directive | this-agent | `Skill(...)` 지시. 이 에이전트에게 스킬을 읽으라고 지시하는 줄이므로 지시다. 처분은 문면 개정이 아니라 AGENTS.md 두 사본의 덮개 1문장 (REQ-CBN-011) |
| internal/template/templates/.codex/agents/moai/plan-auditor.toml | 503 | Skill( | directive | this-agent | `Skill(...)` 지시. 이 에이전트에게 스킬을 읽으라고 지시하는 줄이므로 지시다. 처분은 문면 개정이 아니라 AGENTS.md 두 사본의 덮개 1문장 (REQ-CBN-011) |
| internal/template/templates/.codex/agents/moai/super-advisor.toml | 7 | Agent( | prose | orchestrator | 이 에이전트가 어떻게 스폰되는지의 서술 — 스폰 주체는 오케스트레이터다 |
| internal/template/templates/.codex/agents/moai/super-advisor.toml | 59 | Agent( | prose | orchestrator | "the orchestrator spawns" — 주어가 명시적으로 오케스트레이터다 |
| internal/template/templates/.codex/agents/moai/super-advisor.toml | 62 | AskUserQuestion | prose | orchestrator | "the orchestrator ... escalates to the user via AskUserQuestion" — 주어가 이 에이전트가 아니다 (AC-CBN-002) |
| internal/template/templates/.codex/agents/moai/super-advisor.toml | 125 | Skill( | directive | this-agent | `Skill(...)` 지시. 이 에이전트에게 스킬을 읽으라고 지시하는 줄이므로 지시다. 처분은 문면 개정이 아니라 AGENTS.md 두 사본의 덮개 1문장 (REQ-CBN-011) |
| internal/template/templates/.codex/agents/moai/super-advisor.toml | 131 | Agent( | prose | n/a | 패턴 근거 문서를 가리키는 교차 참조 줄. 행위 지시 없음 |
| internal/template/templates/.codex/agents/moai/sync-auditor.toml | 131 | AskUserQuestion | prose | prohibition | "no sync-auditor path invokes AskUserQuestion" — 어떤 행위도 지시하지 않는 금지 서술. this-agent/orchestrator 어느 쪽도 아니다 (AC-CBN-002) |
| internal/template/templates/.codex/agents/moai/sync-auditor.toml | 159 | Skill( | directive | this-agent | `Skill(...)` 지시. 이 에이전트에게 스킬을 읽으라고 지시하는 줄이므로 지시다. 처분은 문면 개정이 아니라 AGENTS.md 두 사본의 덮개 1문장 (REQ-CBN-011) |
| internal/template/templates/.codex/agents/moai/sync-auditor.toml | 160 | Skill( | directive | this-agent | `Skill(...)` 지시. 이 에이전트에게 스킬을 읽으라고 지시하는 줄이므로 지시다. 처분은 문면 개정이 아니라 AGENTS.md 두 사본의 덮개 1문장 (REQ-CBN-011) |
| internal/template/templates/.codex/agents/moai/sync-auditor.toml | 161 | Skill( | directive | this-agent | `Skill(...)` 지시. 이 에이전트에게 스킬을 읽으라고 지시하는 줄이므로 지시다. 처분은 문면 개정이 아니라 AGENTS.md 두 사본의 덮개 1문장 (REQ-CBN-011) |
| internal/template/templates/.codex/agents/moai/sync-auditor.toml | 162 | Skill( | directive | this-agent | `Skill(...)` 지시. 이 에이전트에게 스킬을 읽으라고 지시하는 줄이므로 지시다. 처분은 문면 개정이 아니라 AGENTS.md 두 사본의 덮개 1문장 (REQ-CBN-011) |

## Gaps

- 이 표는 **본문 축**만 분류한다. 프론트매터 `tools:` 축과 MCP/effort/sandbox 축은
  매니페스트가 이미 처분했으므로 모집단이 아니다(`spec.md` §A.4).
- 미러 스킬 77파일(719 매치)은 이 카드의 모집단이 아니다 — 운영자가 2026-09-07 에
  후속 카드 분리로 확정했다(`spec.md` §D).

## Residual-risk

`:172` 를 directive 로 판정한 것은 수동태 문장의 행위자 귀속 판단이다. 이 판단이 틀렸다면
그 줄은 개정되지 않아야 했던 산문이 되고, 개정 결과는 절차 서술이 능력 이름을 부르는
형태가 된다 — 문서를 망가뜨리지는 않지만 필요 없는 변경이다. 반대 방향(prose 로 두는 것)의
실패는 코덱스 하네스에 Claude 서브에이전트 스폰을 지시하는 줄이 그대로 남는 것이므로,
불확실할 때 directive 로 기우는 쪽이 보수적이다.

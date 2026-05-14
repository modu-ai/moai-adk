# Harness System Audit — 2026-05-14

> **Audit scope**: MoAI-ADK harness 서브시스템 end-to-end fact-check after PR #908 (`/moai harness` 슬래시 명령 thin-wrapper 추가).
> **Audit goal**: 사용자가 요청한 4가지 핵심 ("revfactory/harness 활용, 논문 기반 베스트프랙티스, 프로젝트 맞춤형 자동 생성, 자가 학습 진화")의 실제 작동 여부 검증.
> **Status**: 인프라 90%+ 완성, 일부 wiring 미완 + 외부 레퍼런스 흡수 미흡 → 후속 작업 후보 식별.

---

## 1. 요약

| 영역 | 인프라 구현 | 실제 작동 | 후속 필요 |
|------|------------|----------|----------|
| revfactory/harness 흡수 | ✅ Apache-2.0 attribution + 7-Phase MoAI 적응 | ✅ meta-harness SKILL 15.6KB 동작 | — |
| 학술 논문/베스트프랙티스 | ⚠️ Karpathy + revfactory 2건만 | ⚠️ Anthropic/OpenAI/AutoGen/MetaGPT/AgentBench 등 흡수 미흡 | **GAP-3** |
| 프로젝트 맞춤형 생성 | ✅ 16Q Interview + `ScaffoldHarnessDir` (7-8 file) | ⚠️ dev project에서 의도된 미실행 (`.moai/harness/main.md` 부재) | **GAP-2** |
| 5-Layer Safety | ✅ L1-L5 모두 구현 + e2e_ios_test PASS | ✅ 정상 작동 | — |
| 4-Tier 자가 학습 | ✅ Observer + Learner + Applier + Retention | ⚠️ `usage-log.jsonl` 1 entry (2026-04-27) → hook event drought | **GAP-4** |
| 관리 CLI (`moai harness`) | ✅ 4 verb (status/apply/rollback/disable) | ✅ `/moai harness` 슬래시 명령 PR #908에서 추가 | — |

---

## 2. GAP-2: 프로젝트 맞춤형 하네스 자동 생성 — 진단

### 2.1 트리거 경로 (정상 구현됨)

```
/moai project Phase 5
  → AskUserQuestion Round 1-4 (16Q in conversation_language)
  → Buffer.Append() × 16 (in-memory, REQ-PH-010 HARD)
  → Q16 = "Confirm" → Buffer.Commit()
  → moai-meta-harness skill 호출
  → ScaffoldHarnessDir(harnessDir, ScaffoldOpts{Domain, SpecID, IncludeDesignExtension})
  → 7-8 file 생성 in .moai/harness/:
    • main.md (CLAUDE.md @import 진입점)
    • plan-extension.md (manager-spec chain)
    • run-extension.md (manager-tdd/ddd chain)
    • sync-extension.md (manager-docs chain)
    • design-extension.md (Q13 = "Advanced"일 때만, 8번째)
    • chaining-rules.yaml (machine-readable)
    • interview-results.md (Buffer.WriteResultsToFile)
    • README.md (5-Layer 설명)
```

**증거**: `internal/harness/layer5.go:38-67` (`ScaffoldHarnessDir` + `scaffoldFiles`), `internal/harness/e2e_ios_test.go:103` (`WriteChainingRules`), `.claude/skills/moai/workflows/project.md:770-820` (Phase 5 16Q skeleton).

### 2.2 현재 프로젝트 상태

`/Users/goos/MoAI/moai-adk-go/.moai/harness/` 디렉터리에는 다음만 존재:

```
README.md             8.6KB (4/28)  ← scaffolder가 항상 생성하는 7개 중 1개
usage-log.jsonl       152B  (4/27)  ← observer가 1 entry만 기록
```

**누락된 6 file** (main.md, plan/run/sync-extension.md, chaining-rules.yaml, interview-results.md).

### 2.3 원인 분석

이 프로젝트는 moai-adk-go **dev project** (CLAUDE.local.md §1 "Primary work location: internal/template/templates/"). 의도된 dogfood 범위 밖.

- README.md는 SPEC-V3R3-HARNESS-LEARNING-001 구현 시 manually committed (`.moai/harness/README.md`)
- usage-log.jsonl 1 entry는 4/27 일회성 테스트 흔적
- `/moai project` Phase 5 (16Q Interview)를 dev project에서 한 번도 완주한 적 없음 → `ScaffoldHarnessDir`가 호출된 적 없음

### 2.4 진단 결과

- **버그 아님** — dev project가 user-facing harness flow를 invoke하지 않은 결과
- **권고 (후속 SPEC 후보)**:
  1. user project에서 end-to-end smoke test (16Q → main.md 확인)
  2. dev project에 dogfood 인터뷰 실행 + 결과 commit (만약 의미있다면)
  3. 또는 `/moai harness generate` 신규 verb 도입 (16Q 인터뷰만 stand-alone 실행 가능하게)

---

## 3. GAP-3: 외부 학술/오픈소스 레퍼런스 — 통합 후보

### 3.1 현재 attribution (NOTICE.md)

- **revfactory/harness** (Apache-2.0): 7-Phase workflow, 1차 흡수 완료
- **Karpathy coding principles**: 4 principles + 8 anti-patterns 흡수 완료

### 3.2 흡수 후보 (2026 외부 자료 수집 결과)

| # | 자료 | 핵심 기여 | 통합 후보 위치 |
|---|------|----------|---------------|
| 1 | **OpenAI "harness engineering" formalization** (2026) — 3-7명 팀이 5개월에 ~1M LOC 생성, Codex agent + harness | "harness, not the model, is the binding constraint" → MoAI orchestrator 철학 강화 | `.claude/skills/moai-meta-harness/SKILL.md` §Why Harness Matters 추가 |
| 2 | **Anthropic Harness Engineering Playbook** (Nov 2025 + Mar 2026 연결 논문 2편) | Long-running agent 세션 (multi-hour autonomous) — 5-Layer 추상화 (LLM core + gateway/session + memory + instructions/tools + triggers/outputs) | meta-harness SKILL의 5-Layer Safety와 비교 도해 추가 |
| 3 | **Agent Harness for LLM Agents: A Survey** (Preprints.org 202604.0428, GloriaaaM/LLM-Agent-Harness-Survey HuggingFace) | 학술 survey + dataset → 5-architectural-layer characterization | `.claude/rules/moai/NOTICE.md` attribution 추가 + meta-harness Bibliography 신설 |
| 4 | **AutoGen / Microsoft Agent Framework 1.0** (April 2026 GA) | DevUI 브라우저 debugger, declarative YAML agent definitions, A2A + MCP integration | meta-harness가 생성하는 my-harness 에이전트 패턴 비교 + DevUI 차용 검토 |
| 5 | **MetaGPT role-based message passing** | PM / Architect / Engineer / QA agent assignment, shared message pool | MoAI manager-* / expert-* / builder-* 역할 분리 패턴 cross-reference |
| 6 | **HITL Protocol v0.8** (Feb 2026) | HTTP 202 + review URL — agent ↔ human ↔ messaging channel 표준 | AskUserQuestion bridge 패턴 비교 + ~15 LOC reference implementations |
| 7 | **375 GitHub issues fault taxonomy** (Mar 2026 empirical study) | 5-category fault catalog (init / role deviation / memory / orchestration / tool integration) | `.claude/rules/moai/quality/` 신규 file (fault taxonomy reference) |
| 8 | **AgentBench / SWE-bench / WebArena integrity issues** (2026) | Gold-answer leak (WebArena file:// URL), METR reward-hacking 30%+, SWE-bench contamination | 12-mechanism leniency prevention 보강 (Mechanism 6+: contamination check) |
| 9 | **LLM-as-Judge bias study** (Red Hat 8-stage maturity, $0.64/run) | 50%+ error rate, position/length/agreeableness bias, llama-3-3-70b catches all | evaluator-active 4-dimension scoring + ensemble (multiple judges + minority-veto) 검토 |
| 10 | **NIST AI Agent Standards Initiative** (Feb 2026) | 산업 표준 발판 | MoAI compliance roadmap 신설 (장기) |
| 11 | **awesome-harness-engineering** GitHub list (ai-boost) | 큐레이션 — patterns, evals, memory, MCP, permissions, observability | 외부 reference list로 `.moai/research/harness-external-refs.md` 신설 |

### 3.3 권고 흡수 우선순위

- **P0**: #1 (OpenAI formalization), #2 (Anthropic Playbook), #9 (LLM-as-Judge bias) → MoAI 핵심 철학과 직접 연결
- **P1**: #3 (학술 Survey), #7 (Fault Taxonomy), #8 (Benchmark Integrity) → quality gates 보강
- **P2**: #4 (AutoGen DevUI), #5 (MetaGPT role), #6 (HITL Protocol), #10 (NIST), #11 (awesome list) → 비교 학습 / 장기

---

## 4. GAP-4: 자가 학습 진화 — 진단

### 4.1 Observer wiring (정상 구현됨)

```
PostToolUse hook event
  → internal/cli/hook.go:452 NewObserverWithRetention(logPath, retention)
  → obs.RecordEvent(EventTypeAgentInvocation, subject, "")
  → append JSONL line to .moai/harness/usage-log.jsonl
```

**증거**: `internal/cli/hook.go:452-461`, `internal/harness/observer.go:16` (`@MX:ANCHOR RecordEvent is the entry point called from all hook paths`).

### 4.2 학습 파이프라인 (정상 구현됨)

```
Observer (Tier 1) → 1x event
  ↓
Learner.AggregatePatterns
  ↓ pattern count >= 3
Heuristic (Tier 2)
  ↓ pattern count >= 5, confidence >= 0.80
Rule (Tier 3) → Proposal Generation
  ↓ Tier 4 (10x or high-confidence)
Auto-Update Proposal (Tier 4)
  ↓ 5-Layer Safety pipeline:
    L1 FrozenGuard → L2 Canary → L3 Contradiction → L4 RateLimit → L5 HumanApproval
  ↓ moai-harness-learner skill → AskUserQuestion
Apply (graduate) → .moai/harness/learning-history/snapshots/<date>/
```

**증거**: `internal/harness/learner.go`, `applier.go`, `layer1.go ~ layer5.go`.

### 4.3 현재 작동 상태

- `harness.yaml` `learning.enabled: true` ✅
- `usage-log.jsonl`: **1 entry (2026-04-27)** — 그 후 18일간 추가 entry 없음
- 진단: hook event가 거의 fire되지 않거나, observer가 silent fail

### 4.4 가능한 원인

| 원인 | 검증 방법 |
|------|----------|
| ① `settings.json` PostToolUse hook 등록 누락 | `cat .claude/settings.json \| jq '.hooks.PostToolUse'` 확인 |
| ② Hook 등록은 되어있지만 `learning.enabled: true` 우회 path | `internal/cli/hook.go:452` 도달 전 early return 검사 |
| ③ Observer가 silent fail (`RecordEvent` return error 무시) | `internal/cli/hook.go:461 \_ = obs.RecordEvent(...)` 에러 swallow 확인 |
| ④ Retention pruning이 너무 aggressive | `retention.go` 1시간 cooldown + 90 day window 확인 |
| ⑤ Dev project에서 `_test.go` 환경 등으로 hook 비활성화 | 환경변수 `MOAI_HARNESS_DISABLE` 검사 |

### 4.5 다음 단계 (후속 SPEC 후보)

- **Smoke test SPEC**: `/moai project` 16Q 한 번 실행 + observer 활성화 확인 + 24h 후 usage-log 누적 검증
- **Hook diagnostic verb**: `/moai harness diagnose` 신규 verb (`moai harness diagnose` CLI 신설 → hook registration / retention status / observer health)
- **Observer error reporting**: `internal/cli/hook.go:461` `_ =` 제거하고 error를 separately log

---

## 5. 권고 사항 (Post-Audit Action Plan)

이번 audit 결과를 기반으로 다음 4가지 follow-up 후보를 식별:

### 5.1 SPEC-V3R3-HARNESS-OBS-FIX-001 (P1, 소형)

`internal/cli/hook.go:461`의 observer error swallow 제거 + diagnostic logging 추가 + smoke test (Bash 1회 invocation 후 usage-log entry 증가 검증).

### 5.2 SPEC-V3R3-HARNESS-EXT-REFS-001 (P1, 중형)

외부 학술/오픈소스 레퍼런스 #1-#11 중 P0/P1 (#1, #2, #3, #7, #8, #9) 흡수.
- meta-harness SKILL.md 본문에 references section 추가
- NOTICE.md attribution 확장
- `.moai/research/harness-external-refs.md` 신설

### 5.3 SPEC-V3R3-HARNESS-DIAGNOSE-001 (P2, 소형)

`/moai harness diagnose` verb 추가 — hook registration + observer health + learning subsystem state 통합 진단.

### 5.4 SPEC-V3R3-HARNESS-DOGFOOD-001 (P2, 소형)

Dev project에서 16Q 인터뷰 1회 실행 (또는 단순화된 4Q dogfood interview) + `.moai/harness/main.md` 등 7 file 생성 + commit. moai-adk-go 자체가 user를 위한 harness 모범 사례를 보여줌.

---

## 6. 사용자 요청 5축 vs 실제 상태 매트릭스

| 사용자 goal directive 키워드 | 실제 상태 | 후속 SPEC |
|-----------------------------|----------|----------|
| "revfactory/harness 이용해서" | ✅ Apache-2.0 attribution + 7-Phase MoAI 적응 완료 | — |
| "각종 하네스 논문과 베스트프랙티스" | ⚠️ #1-#11 흡수 미흡 | **5.2** |
| "claude code에서 하네스를 생성하는 커맨드" | ✅ `/moai harness` 슬래시 명령 (PR #908) + `/moai project` 트리거 | (선택) generate verb 분리 |
| "프로젝트 맞춤형으로 하네스가 제작" | ✅ 16Q Interview + `ScaffoldHarnessDir` 구현 / 단 dev project에서 dogfood 미실행 | **5.4** |
| "관리/수정" | ✅ 4 verb CLI (status/apply/rollback/disable) | — |
| "자생적으로 자가 학습을 통한 스스로 진화" | ⚠️ 인프라 구현 / 단 observer event drought (18일간 0 entries) | **5.1** |

---

## 7. Cross-references

- PR #908: https://github.com/modu-ai/moai-adk/pull/908 (GAP-1 해결, `/moai harness` slash 명령)
- SPEC-V3R3-HARNESS-001 (Meta-Harness Skill, completed)
- SPEC-V3R3-HARNESS-LEARNING-001 (Self-learning, completed, PR #728)
- SPEC-V3R3-PROJECT-HARNESS-001 (16Q Interview + 5-Layer, completed)
- SPEC-V3R2-HRN-001/002/003 (Routing + Memory + Hierarchical Scoring)
- `.claude/rules/moai/NOTICE.md` (revfactory/harness + Karpathy attribution)
- `.moai/research/evolution-log.md` (자가 학습 진화 기록 — 현재 비어있음)

## 8. Sources (외부 자료)

- [Best AI Agent Harness Tools and Frameworks 2026 — Atlan](https://atlan.com/know/best-ai-agent-harness-tools-2026/)
- [awesome-harness-engineering — ai-boost](https://github.com/ai-boost/awesome-harness-engineering)
- [Agent Harness for LLM Agents: A Survey — Preprints.org 202604.0428](https://www.preprints.org/manuscript/202604.0428)
- [LLM-Agent-Harness-Survey dataset — HuggingFace](https://huggingface.co/datasets/GloriaaaM/LLM-Agent-Harness-Survey)
- [Anthropic Harness Engineering Playbook — Medium (Joe Njenga, Mar 2026)](https://medium.com/ai-software-engineer/anthropics-new-harness-engineering-playbook-shows-why-your-ai-agents-keep-failing-174a5575ff92)
- [What Is Harness Engineering? Complete Guide for AI Agent Development 2026 — NxCode](https://www.nxcode.io/resources/news/what-is-harness-engineering-complete-guide-2026)
- [Harness Engineering 2026 Discipline — agent-engineering.dev](https://www.agent-engineering.dev/article/harness-engineering-in-2026-the-discipline-that-makes-ai-agents-production-ready)
- [How to Build an Agent Evaluation Framework — Galileo](https://galileo.ai/blog/agent-evaluation-framework-metrics-rubrics-benchmarks)
- [Inside Anthropic's 2026 Developer Conference — Every](https://every.to/chain-of-thought/inside-anthropic-s-2026-developer-conference)
- [AI Agent Framework Scorecard 2026 — Rapid Claw](https://rapidclaw.dev/blog/ai-agent-benchmarks-2026)
- [Codex vs Claude Code 2026 — MindStudio](https://www.mindstudio.ai/blog/codex-vs-claude-code-2026)

---

Version: 1.0.0
Audit Date: 2026-05-14
Auditor: MoAI orchestrator (post PR #908 follow-up)
Status: COMPLETE — 4 follow-up SPEC 후보 식별, 사용자 결정 대기

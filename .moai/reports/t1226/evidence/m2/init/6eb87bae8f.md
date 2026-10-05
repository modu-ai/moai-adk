# M2 attempt — CLAUDE.md | 2. Request Processing Pipeline

surface = init
pre_chars = 998
post_chars = 839
chars = 159
post_hash = 93de7321ea747c584768af09d1908ad3e36fd076073a6f2a9d6fb5b7a22b4477

post_hash is the frozen-multiset sha256 of the whole post-attempt surface (every attempted M2 row replaced, every ADMIT M1 row removed), measured with the AC-ALH-004 pipeline on $SCRATCH/post-init.

## before

````
## 2. Request Processing Pipeline

**Analyze-First** is the default main-session orchestration behavior: every request — in any input language, with or without a `/moai` subcommand — flows through one ordered pipeline, beginning with intent analysis (classify meaning, language-independent, never keyword-gated). The structured Intent Router lives in the `/moai` skill (`.claude/skills/moai/SKILL.md`).

Five ordered stages: ① intent analysis → ② context-sufficiency check (insufficient → Rule 5 Context-First Discovery rounds, §7) → ③ execution-plan composition (`orchestration-mode-selection.md`; surfaced before execution per Approach-First, §7 Rule 1) → ④ **approval gates**, incl. the **Implementation Kickoff Approval** human gate at plan→run (§8; the progression axis is post-approval, never a bypass) → ⑤ execute → verify → iterate against acceptance criteria (an armed `/moai goal` is the termination judge).

Report: consolidate agent results in the user's `conversation_language`.

---

````

## after

````
## 2. Request Processing Pipeline

**Analyze-First** (default main-session behavior): every request — any input language, with or without a `/moai` subcommand — runs one ordered pipeline, starting from intent classified by meaning (language-independent, never keyword-gated). Intent Router: `.claude/skills/moai/SKILL.md`.

① intent analysis → ② context-sufficiency check (insufficient → Rule 5 Discovery rounds, §7) → ③ execution-plan composition (`orchestration-mode-selection.md`; surfaced before execution, §7 Rule 1) → ④ **approval gates**, incl. the **Implementation Kickoff Approval** human gate at plan→run (§8; the progression axis is post-approval, never a bypass) → ⑤ execute → verify → iterate against acceptance criteria (an armed `/moai goal` judges termination). Report agent results in the user's `conversation_language`.

````

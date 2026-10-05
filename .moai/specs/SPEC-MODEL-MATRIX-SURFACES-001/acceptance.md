# Acceptance Criteria — SPEC-MODEL-MATRIX-SURFACES-001

10 acceptance criteria across two milestones. Every criterion states observable evidence; no criterion is satisfied by assertion alone.

Nothing in this SPEC has landed, so every criterion below is unmet at authoring — that is the expected starting state, not a defect.

---

## §D AC Matrix

### §D.1 M4 — Init wizard question

**AC-MPMS-001** (REQ-MPMS-001, 002)
- **Given** the `model_policy` question
- **When** its title, description, and option text are read
- **Then** they use subscription-tier framing and contain no reference to `haiku`
- **Verify**:
  ```bash
  grep -in "haiku" internal/cli/wizard/questions.go
  # expect: no match in the model_policy question
  ```

**AC-MPMS-002** (REQ-MPMS-003)
- **Given** each of the question's options
- **When** the wizard answer is persisted
- **Then** the resulting `llm.profile` is one of `max`, `medium`, `low`
- **Verify**: Go test iterating every option value through the persistence path and asserting the written profile

**AC-MPMS-003** (REQ-MPMS-003, backward compatibility)
- **Given** a legacy stored answer of `high`
- **When** it is normalized
- **Then** it resolves to `max` without error
- **Verify**: Go test on the normalizer

**AC-MPMS-004** (REQ-MPMS-004, 005)
- **Given** `conversation_language` set to `ko`, then `ja`, then `zh`
- **When** the wizard renders the `model_policy` question
- **Then** the title, description, and every option label and description render in that language, with no English fallback
- **Verify**: Go test asserting a translation entry exists for every string of the question in all three locales

**AC-MPMS-005** (REQ-MPMS-006)
- **Given** the option descriptions
- **When** they are read
- **Then** each names the subscription tier it targets and none asserts — or implies — that a higher profile produces stronger results
- **Verify**: manual read against `plan.md` §B.2. The implication half cannot be grepped; a reviewer states what a first-time reader would conclude about the ordering

### §D.2 M5 — Web console cleanup

**AC-MPMS-006** (REQ-MPMS-007)
- **Given** the web console agent-frontmatter model selector
- **When** its option set is enumerated
- **Then** `haiku` is absent
- **Verify**: Go test asserting the option slice excludes the haiku constant

**AC-MPMS-007** (REQ-MPMS-008)
- **Given** the v4manifest tier-suggestion table
- **When** the lightblue tier is looked up
- **Then** it yields `{sonnet, low}`
- **Verify**: Go test on the suggestion table; a second assertion confirms the neighbouring `agentTiers` badge distribution is unchanged, so the criterion cannot pass on a widened edit

**AC-MPMS-008** (REQ-MPMS-009)
- **Given** the four locale files
- **When** `agentfm.tier.desc` is read in each
- **Then** the key exists in all four and its text describes the effort-reapplication behaviour `SPEC-MODEL-MATRIX-CONFIG-001` restores
- **Verify**: Go test asserting key presence in all 4 locales, plus a manual read of the wording

**AC-MPMS-009** (REQ-MPMS-010)
- **Given** the four locale files
- **When** they are searched for the orphaned key family
- **Then** no `mp.` key remains, and a consumer grep run **before** the removal confirmed none was in use
- **Verify**:
  ```bash
  grep -rn '"mp\.' internal/web/
  # expect: no matches
  ```
  plus the pre-removal consumer-grep output recorded in `progress.md` — a post-removal grep alone cannot distinguish "was orphaned" from "was live and is now broken"

**AC-MPMS-010** (REQ-MPMS-011)
- **Given** the four locale files after M5
- **When** their key sets are compared
- **Then** all four contain the same key set
- **Verify**: Go test computing the 4 key sets and asserting equality

---

## §E Severity classification

| Severity | ACs | Meaning |
|---|---|---|
| **Must-pass (user-facing)** | AC-MPMS-001 … 005 | A violation ships misleading information to users at the first surface they meet |
| **Must-pass** | AC-MPMS-006, 007, 009 | A No-Haiku policy violation on a user-facing surface, or a removal that breaks a live key |
| **Should-pass** | AC-MPMS-008, 010 | Cleanup and parity hygiene; a violation is cosmetic debt, not a defect |

AC-MPMS-009's Must-pass placement is deliberate: the *removal* is cleanup, but removing a live key is a silent breakage, which is why its verification demands pre-removal evidence rather than a post-removal grep.

---

## §F Verification-method distribution

| Method | Count | ACs |
|---|---|---|
| Go test (behavioral) | 6 | AC-MPMS-002, 003, 004, 006, 007, 010 |
| Shell command (absence/presence) | 2 | AC-MPMS-001, 009 |
| Manual read (copy quality) | 2 | AC-MPMS-005, plus the wording half of 008 |

AC-MPMS-005 is manual by necessity: no grep can verify that a paragraph avoids *implying* an ordering it never states. It is satisfied by a reviewer reading the text against `plan.md` §B.2, not by a token match.

---

## §G Indirect-verification notes

Two criteria use a proxy rather than direct observation:

- **AC-MPMS-004** proxies "renders localized" through translation-entry presence rather than a rendered-output capture. A present-but-wrong-language entry would evade it; accepted, since catching that needs a human read of each locale anyway.
- **AC-MPMS-009** cannot be satisfied by its own grep alone. A post-removal zero-match result is identical whether the family was orphaned or live-and-now-broken, so the criterion requires the pre-removal consumer grep as recorded evidence. This is the one criterion here whose verification order is load-bearing.

---

## §H Closure gates

This SPEC may close when:

1. All Must-pass criteria pass with cited evidence.
2. Any failing Should-pass criterion is recorded as debt with an owner.
3. The wizard option-value clarification (`plan.md` §B.1) is resolved and the resolution recorded.
4. The `mp.*` family's consumer status is **measured**, not inherited, and the measurement recorded in `progress.md`.
5. The handoff to `SPEC-MODEL-MATRIX-DOCS-001` is recorded: both cleared surfaces now need haiku-residual guard coverage.
6. `go test ./internal/cli/... ./internal/web/... ./internal/harness/...` are green.

---

## §I Forward-looking checks

Conditions that would indicate this SPEC's work has regressed after closure:

- `haiku` reappears in the web-console model option set or the v4manifest tier suggestions.
- The wizard question regains a model-class framing or a performance-ordering claim.
- A locale file gains a key absent from the other three.
- The `model_policy` question loses a locale entry and silently falls back to English.
- The v4manifest `agentTiers` badge distribution changes as a side effect of a tier-suggestion edit.

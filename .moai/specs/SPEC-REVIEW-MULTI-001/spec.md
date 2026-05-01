---
id: SPEC-REVIEW-MULTI-001
status: draft
version: "0.1.0"
priority: High
labels: [review, multi-agent, parallel, verification, severity, wave-2, tier-1]
issue_number: null
scope: [review.md, team/review.md]
blockedBy: []
dependents: []
related_specs: [SPEC-V3R3-PATTERNS-001, SPEC-EVAL-LOOP-001]
created_at: 2026-04-30
updated_at: 2026-04-30
author: manager-spec
wave: 2
tier: 1
---

# SPEC-REVIEW-MULTI-001: Code Review Multi-Agent 패턴 도입

## HISTORY

- 2026-04-30 v0.1.0: 최초 작성. Wave 2 / Tier 1. Anthropic "Code Review (with Claude)" 3-stage 패턴 (parallel detection → verification → severity ranking)을 `/moai review`에 도입.

---

## 1. Goal (목적)

Anthropic blog "Code Review (with Claude)"의 3-stage multi-agent 패턴 (parallel detection → verification → severity ranking)을 `/moai review` 워크플로우에 도입한다. 단일 agent 검토의 false-positive 문제를 verification stage로 해소하고, 병렬 detection으로 large PR (>1,000 LOC) coverage를 84% 수준 (Anthropic baseline)에 근접시킨다.

### 1.1 배경

- Anthropic: "We found that a single agent reviewing in isolation produced too many false positives. The verification step cut false positives by more than half."
- 현재 `/moai review`는 manager-quality 단일 위임 (4 perspective sequential) → 한 관점 지배, false-positive filter 없음
- Severity ranking이 manager-quality 내부에 묻혀 있어 일관성 부재

### 1.2 비목표 (Non-Goals)

- 신규 specialized review agent 생성 (기존 expert-* + general-purpose 조합 사용)
- review 외 다른 워크플로우의 multi-agent 일반화 (review 한정)
- review 결과의 자동 fix 적용 (review = 분석, fix는 별도 워크플로우)
- Anthropic baseline (84%/31%) 재현 의무 (참고 metric으로만 사용)
- chunked large-PR 처리 알고리즘 (별도 SPEC, 본 SPEC은 분기 정책만)

---

## 2. Scope (범위)

### 2.1 In Scope

- `.claude/skills/moai/workflows/review.md` 재작성:
  - Stage 1 (Parallel Detection): 4 agent 병렬 (expert-security, expert-performance, manager-quality, expert-refactoring)
  - Stage 2 (Verification): finding당 verifier agent (general-purpose)
  - Stage 3 (Severity Ranking): ranker agent (general-purpose)
  - Large PR (>1,000 LOC) 분기: depth 증가 (expert-debug 추가)
  - Small PR (<50 LOC) 분기: 기존 단일 agent 경로 유지 (token 절약)
- `.claude/skills/moai/team/review.md` 강화: 3-stage with team mode
- Template-First 동기화

### 2.2 Exclusions (What NOT to Build)

- 신규 review agent 신설 (general-purpose 활용)
- review 외 워크플로우의 multi-agent 일반화
- 자동 fix 적용
- chunked large-PR 처리 (token 한계 대응은 별도 SPEC)
- review report의 publish (GitHub PR comment posting 등)
- review 결과의 SPEC 자동 생성

---

## 3. Environment (환경)

- 런타임: moai-adk-go (Go 1.23+)
- Claude Code v2.1.111+
- 영향 디렉터리: `.claude/skills/moai/workflows/`, `.claude/skills/moai/team/`
- 템플릿 동기화: `internal/template/templates/`

---

## 4. Assumptions (가정)

- A1: 4 detection agent을 병렬로 spawn해도 read-only이므로 file race 없음
- A2: verifier는 general-purpose context에서 detection agent와 충분히 독립적
- A3: `git diff --shortstat`으로 PR LOC 측정 가능
- A4: small PR (<50 LOC) 단일 agent 경로 유지가 token 비용 효율적
- A5: ranker가 프로젝트 risk profile (.moai/project/tech.md, structure.md)을 인용 가능

---

## 5. Requirements (EARS Format)

### 5.1 Ubiquitous Requirements

- **REQ-RM-001**: THE `/moai review` WORKFLOW SHALL implement 3 distinct stages: Parallel Detection, Verification, Severity Ranking.
- **REQ-RM-002**: THE STAGE 1 (PARALLEL DETECTION) SHALL dispatch 4 agents simultaneously: expert-security, expert-performance, manager-quality, expert-refactoring.
- **REQ-RM-003**: THE STAGE 2 (VERIFICATION) SHALL spawn an independent verifier agent for each candidate finding from Stage 1.
- **REQ-RM-004**: THE STAGE 3 (SEVERITY RANKING) SHALL spawn a dedicated ranker agent that classifies verified findings into Critical / High / Medium / Low.
- **REQ-RM-005**: THE FINAL REVIEW REPORT SHALL group findings by severity in descending order with file:line references for each finding.

### 5.2 Event-Driven Requirements

- **REQ-RM-006**: WHEN `/moai review` is invoked, THE WORKFLOW SHALL measure PR LOC via `git diff --shortstat` to determine routing.
- **REQ-RM-007**: WHEN PR size is less than 50 LOC, THE WORKFLOW SHALL execute single-agent review (legacy manager-quality path) for token efficiency, with multi-agent path opt-in via `--full` flag.
- **REQ-RM-008**: WHEN PR size is between 50 and 1,000 LOC, THE WORKFLOW SHALL execute the 3-stage multi-agent review by default.
- **REQ-RM-009**: WHEN PR size exceeds 1,000 LOC, THE WORKFLOW SHALL execute the 3-stage review AND additionally include expert-debug in Stage 1 detection agents.
- **REQ-RM-010**: WHEN all Stage 1 detection agents complete, THE WORKFLOW SHALL aggregate findings into a unified candidate list before invoking Stage 2.

### 5.3 State-Driven Requirements

- **REQ-RM-011**: WHILE Stage 1 detection agents are running in parallel, THE ORCHESTRATOR SHALL collect their results without enforcing inter-agent communication.
- **REQ-RM-012**: WHILE Stage 2 verification is in progress, THE ORCHESTRATOR SHALL track which findings are verified vs dropped, with rationale captured for each drop.

### 5.4 Conditional Requirements

- **REQ-RM-013**: WHERE the Stage 2 verifier executes the 4-step false-positive filter algorithm defined in §5.6 against a candidate finding AND the algorithm returns `verified: false` with `confidence < 0.50`, THEN THE FINDING SHALL be dropped from the candidate list. The drop reason SHALL record which of the 4 algorithm steps failed (reproducer attempt / pattern reproduction / OWASP-CWE mapping / confidence floor) and the verifier-emitted rationale.
- **REQ-RM-014**: WHERE the project's `.moai/project/tech.md` declares a sensitive domain (auth, payment, public_api), THEN THE RANKER SHALL elevate Security findings by one severity level (e.g., Medium → High).
- **REQ-RM-015**: IF the same finding is reported by multiple Stage 1 agents, THEN THE RANKER SHALL deduplicate by file:line + symptom signature, retaining all originating agents in metadata.
- **REQ-RM-016**: THE STAGE 1 DETECTION AGENTS AND THE STAGE 2 VERIFIER AGENT SHALL run in read-only mode (Claude Code `mode: "plan"` or equivalent permission restriction that denies Write/Edit) AND SHALL NOT use `isolation: "worktree"`. Rationale: per CLAUDE.md §14 [HARD] "Read-only teammates (role_profiles: researcher, analyst, reviewer) MUST NOT use isolation: 'worktree'". Review agents are categorized as reviewer role; therefore worktree isolation is prohibited regardless of any opt-in flag. Stage 3 ranker is similarly read-only and follows the same constraint.
- **REQ-RM-017**: IF Stage 1 detection agents collectively produce zero candidate findings, THEN THE WORKFLOW SHALL skip Stage 2 and Stage 3, producing a "no findings" report.

### 5.5 Unwanted (Negative) Requirements

- **REQ-RM-018**: THE STAGE 1 DETECTION AGENTS SHALL NOT communicate with each other during parallel execution; their context isolation is essential for independence.
- **REQ-RM-019**: THE STAGE 2 VERIFIER SHALL NOT be the same agent that produced the finding (independence requirement).
- **REQ-RM-020**: THE WORKFLOW SHALL NOT proceed past Stage 2 with unverified findings included in the final report.

### 5.6 Verification Algorithm (False-Positive Filter)

The Stage 2 verifier MUST execute the following deterministic 4-step algorithm against each candidate finding produced by Stage 1. The algorithm output drives REQ-RM-013 (drop decision). The verifier is invoked as a `general-purpose` agent in read-only mode with the originating agent identity stripped from its prompt (REQ-RM-019, REQ-RM-016).

**Algorithm input**:
- `finding.file` (string, file path)
- `finding.line` (integer, line number)
- `finding.severity` (enum: Critical / High / Medium / Low — preliminary, may be re-ranked in Stage 3)
- `finding.description` (string, symptom and suspected cause)
- `finding.category` (enum: security / performance / quality / refactoring / debug)
- The PR diff hunk(s) covering `finding.file:finding.line`

**Step 1 — Reproducer attempt** (mandatory):
- The verifier attempts to construct a minimal reproducer: a hypothetical input, code path, or call sequence that would exercise the alleged defect.
- A reproducer is considered VALID when it (a) names a concrete entry point in the diff or surrounding code, (b) describes inputs or state that would trigger the symptom, and (c) explains the observable failure.
- If a valid reproducer is constructed: emit `step1.verified = true`, `step1.confidence = 0.40`. Otherwise: `step1.verified = false`, `step1.confidence = 0.00`.

**Step 2 — Pattern reproduction via AST/grep** (mandatory):
- The verifier independently searches the diff hunks for the symptom pattern using language-aware AST inspection or `git grep` over the file at `finding.file:finding.line`.
- For example, an alleged "unparameterized SQL string concatenation" finding is checked by searching for string concatenation flowing into a SQL execution call within the same scope.
- If the pattern is independently reproducible: emit `step2.verified = true`, `step2.confidence += 0.30`. Otherwise: `step2.verified = false`, `step2.confidence += 0.00`.

**Step 3 — OWASP / CWE mapping (Security findings only)** (conditional):
- IF `finding.category == "security"`, the verifier attempts to map the symptom to an OWASP Top 10 category or CWE identifier (e.g., CWE-89 for SQL injection, CWE-79 for XSS).
- A successful mapping requires a CWE/OWASP identifier AND a one-sentence justification anchored in the diff.
- If mapping succeeds: emit `step3.verified = true`, `step3.confidence += 0.20`. If mapping fails or no obvious match exists: `step3.verified = false`, `step3.confidence += 0.00`.
- For non-security findings (`category != "security"`), `step3.verified = null` (not applicable) and `step3.confidence += 0.00`. The aggregate confidence floor is then satisfied by Steps 1 and 2 alone.

**Step 4 — Confidence floor decision** (mandatory):
- Compute `total_confidence = step1.confidence + step2.confidence + step3.confidence` (capped at 1.00).
- The verifier emits `verified = true` IF `total_confidence >= 0.50` AND at least Step 1 OR Step 2 returned `true`. Otherwise `verified = false`.
- If `verified = false`, the finding is dropped per REQ-RM-013, and the drop reason records `failed_steps` (which steps returned `false`) and `total_confidence`.

**Algorithm output**:
```yaml
verifier_output:
  finding_id: F-NNN
  verified: true | false
  total_confidence: 0.00..1.00
  step1: {verified: true|false, confidence: 0.0..0.40, reproducer: "..."}
  step2: {verified: true|false, confidence: 0.0..0.30, ast_or_grep_evidence: "..."}
  step3: {verified: true|false|null, confidence: 0.0..0.20, cwe_owasp: "CWE-89" | null}
  rationale: "Concise summary of pass/fail reasoning across the 4 steps"
  drop_reason: "<populated only when verified=false>"
```

This algorithm is the canonical false-positive filter; the verifier prompt MUST encode the 4 steps and the confidence floor verbatim. Stage 2 implementations MUST NOT substitute subjective heuristics for the algorithm.

---

## 6. Success Criteria

| Criterion | Measurement | Target |
|-----------|-------------|--------|
| Detection coverage vs single-agent baseline | sample 5 PRs comparison | > 80% more findings |
| False-positive rate (verifier-dropped) | dropped / total candidates | < 15% |
| Large PR (>1,000 LOC) finding rate | sample 5 large PRs | reasonable (Anthropic baseline ~84%) |
| Token cost vs single-agent | per-PR token | < +200% (4x agents + verifiers + ranker) |
| Wall-clock time vs single-agent | parallel stage 1 timing | < +50% (parallelism benefit) |
| Template-First sync | `make build` diff | clean |

---

## 7. Acceptance References

See `acceptance.md` for Given-When-Then scenarios and Definition of Done.

---

## 8. Constraints

- C1: 신규 specialized review agent 신설 금지 (general-purpose 활용)
- C2: review 외 워크플로우 변경 금지
- C3: small PR (<50 LOC) 단일 agent 경로 유지 (token 절약)
- C4: Stage 1 detection agent 간 통신 금지 (independence 핵심)
- C5: verifier는 originating agent와 다른 컨텍스트여야 함

---

## 9. Frontmatter Field Semantics (Wave 2 Tier 1 Standard)

This section defines the canonical meaning of inter-SPEC reference fields used in `.moai/specs/*/spec.md` frontmatter. All 5 SPECs in Wave 2 Tier 1 (EVAL-LOOP-001, LOOP-TERM-001, EVAL-RUBRIC-001, REVIEW-MULTI-001, SKILL-TEST-001) follow this standard.

| Field | Semantic | Blocking? |
|-------|----------|-----------|
| `blockedBy: [SPEC-X-001, ...]` | This SPEC's implementation cannot start until the listed SPECs are completed. HARD dependency. | Yes |
| `dependents: [SPEC-Y-001, ...]` | The listed SPECs are blocked by this SPEC (inverse of `blockedBy`). Forward declarations to future SPECs are allowed. | Yes (transitively) |
| `related_specs: [SPEC-Z-001, ...]` | Semantic association only; reference for context. NOT blocking. Cross-references for design coherence. | No |

### Application to this SPEC

- `blockedBy: []` — No prior SPEC must be completed first.
- `dependents: []` — No SPEC currently waits on this one for unblocking.
- `related_specs: [SPEC-V3R3-PATTERNS-001, SPEC-EVAL-LOOP-001]` — Shares Generator-Verifier and parallel multi-agent design themes; not blocked by or blocking this SPEC.

End of spec.md (SPEC-REVIEW-MULTI-001 v0.1.0).

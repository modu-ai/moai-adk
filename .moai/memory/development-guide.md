# MoAI-ADK Development Guide (TRUST 5 Principles)

> "No spec, no code. No tests, no implementation."

This development guide is the unified guardrail for all agents and developers working on the MoAI-ADK project. It applies to Claude Code sessions, Headless CLI automation, and both team and individual workflows. Korean is the default communication language for the project.

---

## 0. Scope · Base Loop

- Applies to: `/moai:0-project → /moai:3-sync`, `/moai:git:*`, and any work performed via Headless CLI (`codex exec`, `gemini -p`).
- Base action loop: **Problem Definition → Small & Safe Change → Change Review → Refactor → Docs/TAG Sync**.
- All changes must follow AGENTS rules, the 16-Core TAG system, the Waiver process, structured logging, and security policies.

---

## TRUST 5 — Core Engineering Principles

### **T** - **Test First**

1. **Red → Green → Refactor** (Kent Beck)
   - RED: Write a failing test first and confirm it fails.
   - GREEN: Make it pass with the minimum code and commit.
   - REFACTOR: Refactor only when all tests pass.
2. **Test Policy**
   - New features require new tests; bug fixes require regression tests.
   - Tests must be deterministic and isolated; replace external systems with fakes/contracts.
   - Include at least one success and one failure path in E2E.
3. **Coverage Goal**: Recommended ≥ 85%. If lower, record a remediation plan or a Waiver.

### **R** - **Readable**

1. **Clean Code Rules** (Robert C. Martin)
   - Clear names, small functions (≤ 50 LOC), single responsibility, minimal parameters.
   - Remove duplication; reveal intent with structure; keep comments minimal (intent only).
   - Follow team formatting rules; add vertical whitespace between related concepts.
2. **Variable Roles** (Sajaniemi): Use the 11 roles explicitly, e.g., Fixed Value, Stepper, Flag, Walker, etc.
3. **Side-Effect Isolation**: Separate I/O, network, and global state into boundaries; prefer guard clauses; symbolize constants.

### **U** - **Unified**

1. **Complexity Recommendation**: Default `simplicity_threshold = 5` per module. If exceeding, write a Waiver (reason, risk, mitigation, expiry conditions).
2. **Fowler’s Two Hats**: Separate “feature work” and “refactoring”; never wear both hats at once.
3. **Refactoring Signals**: Long functions, large classes, long parameter lists, duplication, feature envy. Plan refactors when smells appear.
4. **Architecture Principles**: Separate Domain/Application/Infrastructure layers, follow DIP, interface-first design, document API/data contracts.

### **S** - **Secured**

1. Collect structured logs (JSONL) and key metrics (latency, error rate, throughput, resources); mask PII/secrets with `***redacted***`.
2. Record all significant events for audit; on hook failure, trip a circuit breaker and switch to safe mode (additional approvals, restricted Bash).
3. Pre-block dangerous commands (`rm -rf`, disallowed network) and manage allowlists via `policy_block.py`.
4. **Security/Quality**: Validate/normalize/encode inputs, structured (JSON) logging, mask sensitive info, apply least privilege.

### **T** - **Trackable**

1. Maintain semantic versioning (MAJOR.MINOR.BUILD); make Git history traceable via `@TAG`s and commit messages.
2. Use `/moai:git:*` commands for checkpoints, branches, commits, and sync; auto-tag before/after risky ops.
3. Follow `/moai:3-sync` checklist when moving Draft PR → Ready.

---

## Article I — Mindset & Decision Loop

1. **Senior Engineer Mindset**: Decide based on facts, not assumptions; compare at least two alternatives (pros/cons/risks) and choose the simplest viable solution.
2. **Whole-Context Reading**: Before edits, read related files/definitions/references/tests/docs/flags end-to-end and jot down 1–3 lines on impact.
3. **Korean Communication**: Korean is the default for team/AI communication; quote originals when needed and add explanations.

---

## Article II — Workflow Guardrails

1. **Pre-work**: Before coding, clarify `Background/Problem/Goals/Non-Goals/Constraints` and check required SPEC/TAGs.
2. **Small & Safe Changes**: Minimize scope of PRs/commits/files; create checkpoints before risky operations.
3. **Review & Docs**: After changes, run `/moai:3-sync` to update the living docs/TAGs; capture summaries/reviews/refactor plans.

---

## Article III — 16-Core @TAG System (Traceability)

1. Maintain the 16-Core @TAG chain: Primary (@REQ → @DESIGN → @TASK → @TEST), Steering, Implementation, Quality.
2. Keep `.moai/indexes/tags.json` and `.moai/reports/sync-report.md` up to date.
3. When reporting headless analysis (`gemini -p`) or implementation (`codex exec`), state the TAGs used and chain status.

---

## Article IV — Review & Refactoring Discipline

1. **Rule of Three**: On the third repetition of a pattern, plan a refactor.
2. **Preparatory Refactoring**: Prepare the environment to make the change easy, then apply the change.
3. **Litter-Pickup**: Fix small smells immediately; if scope grows, split into a separate task.

---

## Article V — Microservice/API Patterns (Olaf Zimmermann)

1. **Foundation**: Choose the appropriate frontend integration strategy among BFF, API Gateway, and Client-Side Composition.
2. **Design**: Specify patterns such as Request/Response, Request-Acknowledge, Event Message, and keep contracts documented.
3. **Quality**: Apply performance patterns (Pagination, Wish List, Conditional Request) and security patterns (Rate Limiting, Circuit Breaker).
4. **Evolution**: Manage compatibility via explicit version IDs, "Two in Production", Consumer-Driven Contracts, Published Language.

---

## Article VI — Exceptions & Waivers

- When deviating from or exceeding recommendations, write a Waiver and attach it to PR/Issue/ADR.
- Waiver must include: reason, alternatives considered, risks/mitigations, temporary/permanent status, expiry conditions, approver.

---

## Operational Appendix A — Work Loop & Checklist

1. **Preparation**
   - Write Background/Problem/Goals/Non-Goals/Constraints
   - Read all related files/tests/docs/flags end-to-end
   - Draft an alternatives comparison table
2. **Execution**
   - Create required SPEC/TAGs
   - Make small changes with per-change checkpoints
   - Follow the TDD cycle; run tests/linters
3. **Wrap-up**
   - Run `/moai:3-sync` → update TAG index and docs
   - Record logs and a summary for analysis/implementation commands (`codex exec`, `gemini -p`)
   - Leave TODO/refactor items and request review

---

## Operational Appendix B — Sajaniemi’s Variable Roles

| Role               | Description                         | Example                               |
| ------------------ | ----------------------------------- | ------------------------------------- |
| Fixed Value        | Constant after initialization       | `const MAX_SIZE = 100`                |
| Stepper            | Changes sequentially                | `for (let i = 0; i < n; i++)`         |
| Flag               | Boolean state indicator             | `let isValid = true`                  |
| Walker             | Traverses a data structure          | `while (node) { node = node.next; }`  |
| Most Recent Holder | Holds the most recent value         | `let lastError`                       |
| Most Wanted Holder | Holds optimal/maximum value         | `let bestScore = -Infinity`           |
| Gatherer           | Accumulator                         | `sum += value`                        |
| Container          | Stores multiple values              | `const list = []`                     |
| Follower           | Previous value of another variable  | `prev = curr; curr = next;`           |
| Organizer          | Reorganizes data                    | `const sorted = array.sort()`         |
| Temporary          | Temporary storage                   | `const temp = a; a = b; b = temp;`    |

---

## Operational Appendix C — Refactoring Quick Reference

- **Extract Method**: Reveal intent and remove duplication
- **Rename Variable**: Use meaningful names
- **Move Method**: Move to the appropriate object
- **Replace Temp with Query**: Prefer query over temps
- **Introduce Parameter Object**: Group related parameters
- **Matt Beck Rule**: "Do not implement while tests are failing"

---

## Operational Appendix D — TDD & Microservice Patterns

- **TDD Rules**: Write tests first, confirm failure, implement minimally, refactor only when all tests pass.
- **Microservice Quality Patterns**: Apply Pagination, Conditional Request, Rate Limiting, Circuit Breaker.
- **API Documentation**: Maintain OpenAPI/Swagger; verify both sides via Consumer-Driven Contracts.

---

This guide provides standards to execute the MoAI-ADK 4-stage pipeline, Git automation, Headless CLI automation, and team/individual collaboration safely and consistently. All contributors should link this document into session-start memory (e.g., `CLAUDE.md`) for constant reference.

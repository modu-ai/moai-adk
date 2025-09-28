# MoAI-ADK Development Guide (SPEC-First TDD Principles)

> "No SPEC, no code. No tests, no implementation. SPEC-First TDD, Language Agnostic."

This development guide is the unified guardrail for all agents and developers working with the MoAI-ADK universal development toolkit. **MoAI-ADK itself is built with TypeScript**, but **supports all major programming languages** for user projects. It follows the **SPEC-First TDD methodology** with 16-Core @TAG v4.0 distributed traceability (94% optimized). Korean is the default communication language for the project.

---

## 0. SPEC-First TDD Workflow

**Core Development Loop (3-Stage)**:
1. **SPEC Creation** (`/moai:1-spec`) → 명세 없이는 코드 없음
2. **TDD Implementation** (`/moai:2-build`) → 테스트 없이는 구현 없음
3. **Documentation Sync** (`/moai:3-sync`) → 추적성 없이는 완성 없음

**On-Demand Quality Assurance**:
- **Debug & Validation** (`@agent-debug-helper`) → 디버깅이 필요할 때 debug 에이전트 호출해서 사용

All changes must follow the 16-Core TAG v4.0 distributed system, SPEC-driven requirements, and language-appropriate TDD practices.

---

## TRUST 5 — SPEC-First TDD Engineering Principles

### **T** - **Test-Driven Development (SPEC-Based)**

1. **SPEC → Test → Code** (SPEC-First TDD Cycle)
   - **SPEC**: Create detailed SPEC with `@REQ`, `@DESIGN`, `@TASK` tags first
   - **RED**: Write failing tests based on SPEC requirements and confirm failure
   - **GREEN**: Implement minimum code to pass tests and fulfill SPEC
   - **REFACTOR**: Improve code quality while maintaining SPEC compliance
2. **Language-Specific TDD Implementation**
   - **Python**: pytest + SPEC-driven test cases (type hints with mypy)
   - **TypeScript/JavaScript**: Vitest + SPEC-based test suites (strict typing)
   - **Java**: JUnit + SPEC annotations (behavior-driven tests)
   - **Go**: go test + SPEC table-driven tests (interface compliance)
   - **Rust**: cargo test + SPEC documentation tests (trait validation)
   - **C++**: GoogleTest + SPEC template tests (concept validation)
   - **C#**: xUnit + SPEC attribute tests (contract validation)
3. **SPEC-TDD Integration**: Each test must trace back to specific SPEC requirements via @TAG references.

### **R** - **Requirements-Driven Readable Code**

1. **SPEC-Aligned Clean Code**
   - Functions directly implement SPEC requirements (≤ 50 LOC per function)
   - Variable names reflect SPEC terminology and domain language
   - Code structure mirrors SPEC design decisions
   - Comments only for SPEC clarifications and @TAG references
2. **Language-Specific SPEC Implementation**
   - **Python**: Type hints reflecting SPEC interfaces + mypy validation
   - **TypeScript**: Strict interfaces matching SPEC contracts + Biome
   - **Java**: Classes implementing SPEC components + strong typing
   - **Go**: Interfaces fulfilling SPEC requirements + gofmt
   - **Rust**: Types embodying SPEC safety requirements + rustfmt
3. **SPEC Traceability**: Every code element should be traceable to SPEC via @TAG comments.

### **U** - **Unified SPEC Architecture**

1. **SPEC-Driven Complexity Management**: Each SPEC defines complexity thresholds. Exceeding requires new SPEC or Waiver with clear justification.
2. **SPEC Implementation Phases**: Separate SPEC creation from implementation; never modify SPEC during TDD cycle.
3. **Cross-Language SPEC Compliance**:
   - **Python**: Modules following SPEC component boundaries
   - **TypeScript**: Interfaces implementing SPEC contracts
   - **Java**: Packages aligned with SPEC architecture
   - **Go**: Packages respecting SPEC interface definitions
   - **Rust**: Crates embodying SPEC module separation
4. **SPEC-Driven Architecture**: Domain boundaries defined by SPEC, not language conventions. Use 16-Core @TAG v4.0 distributed system for cross-language traceability (487KB, 45ms load).

### **S** - **SPEC-Compliant Security**

1. **SPEC Security Requirements**: Every SPEC must define security requirements, data sensitivity, and access controls explicitly.
2. **Security-by-Design**: Security controls implemented during TDD phase, not retrofitted after completion.
3. **Language-Agnostic Security Patterns**:
   - Input validation based on SPEC interface definitions
   - Audit logging for SPEC-defined critical operations
   - Access control following SPEC permission models
   - Secret management per SPEC environment requirements
4. **MoAI-ADK Security**: TypeScript policy-block hooks enforce SPEC security rules across all language implementations.

### **T** - **SPEC Traceability**

1. **SPEC-to-Code Traceability**: Every code change must reference SPEC ID and specific requirement via 16-Core @TAG v4.0 distributed system (category-based storage).
2. **3-Stage Workflow Tracking**:
   - `/moai:1-spec`: SPEC creation with @REQ, @DESIGN, @TASK tags
   - `/moai:2-build`: TDD implementation with @TEST, @FEATURE tags
   - `/moai:3-sync`: Documentation sync with @DOCS, @TAG tags
   - `@agent-debug-helper`: 온디맨드 디버깅 with @PERF, @SEC tags
3. **Cross-Language TAG Consistency**: Distributed JSONL stores maintain unified traceability across all programming languages with 94% storage optimization.

---

## Article I — SPEC-First Mindset

1. **SPEC-Driven Decisions**: All technical decisions must reference existing SPEC or create new SPEC. No implementation without clear requirements.
2. **SPEC-Context Reading**: Before any code changes, read relevant SPEC documents, understand @TAG relationships, and verify compliance.
3. **SPEC Communication**: Korean is default for communication; all SPEC documents use clear Korean with technical terms in English.

---

## Article II — SPEC-TDD Workflow

1. **SPEC-First**: Create or reference SPEC before any code. Use `/moai:1-spec` to define requirements, design, and tasks clearly.
2. **TDD Implementation**: Follow Red-Green-Refactor strictly. Use `/moai:2-build` with language-appropriate testing frameworks.
3. **Traceability Sync**: Run `/moai:3-sync` to update documentation and maintain @TAG relationships across SPEC and code.

---

## Article III — 16-Core @TAG System v4.0 (Distributed Traceability)

1. Maintain the 16-Core @TAG chain: Primary (@REQ → @DESIGN → @TASK → @TEST), Steering, Implementation, Quality.
2. **Distributed Storage Structure v4.0**:
   - **Categories**: `.moai/indexes/categories/*.jsonl` (category-based JSONL files)
   - **Relations**: `.moai/indexes/relations/chains.jsonl` (chain mappings)
   - **Cache**: `.moai/indexes/cache/summary.json` (fast access summary)
   - **Performance**: 487KB total, 45ms load time, 94% size reduction
3. Keep sync reports updated: `.moai/reports/sync-report.md`


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
   - Record logs and a summary for analysis/implementation commands 

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

This guide provides SPEC-First TDD standards to execute the MoAI-ADK 3-stage pipeline (`/moai:1-spec` → `/moai:2-build` → `/moai:3-sync`) with universal language support and 16-Core @TAG v4.0 distributed traceability (94% optimized). Use `@agent-debug-helper` when issues arise. All contributors should follow SPEC-driven development with language-appropriate TDD practices.

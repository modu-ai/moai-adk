# Progressive Disclosure - Content Architecture

Purpose: Three-tier knowledge delivery system balancing immediate value with comprehensive depth through strategic content structuring and file organization.

Version: 1.0.0
Last Updated: 2025-11-25

---

## Quick Reference (30 seconds)

Progressive Disclosure = Layered content delivery matching user expertise and time investment.

Three Levels:
1. Quick Reference (30 seconds) - Core principles, immediate value
2. Implementation Guide (5 minutes) - Workflows, patterns, examples
3. Advanced Patterns (10+ minutes) - Deep dives, edge cases, optimization

SKILL.md Structure (≤500 lines hard limit):
```markdown
## Quick Reference (30s) # 80-120 lines
## Implementation Guide (5min) # 180-250 lines
## Advanced Patterns (10+min) # 80-140 lines
## Works Well With # 10-20 lines
```

Token Efficiency:
- Level 1: 1,000 tokens → Immediate value
- Level 2: 3,000 tokens → Practical guidance
- Level 3: 5,000 tokens → Expert knowledge

File Overflow: When SKILL.md > 500 lines → Split to modules/

---

## Implementation Guide (5 minutes)

### Level 1: Quick Reference Design

Purpose: Deliver maximum value in minimum time (30 seconds reading).

Content Requirements:
- Core concepts in 1-2 sentences
- Key patterns/workflows as diagrams
- Essential syntax/commands
- Quick decision matrix
- Cross-references to deeper content

Structure Template:

```markdown
## Quick Reference (30 seconds)

[Skill Name] is [one-sentence definition].

Core Principles:
1. Principle 1 - Brief explanation
2. Principle 2 - Brief explanation
3. Principle 3 - Brief explanation

Quick Access:
- Pattern A → [Link to module](modules/pattern-a.md)
- Pattern B → [Link to module](modules/pattern-b.md)
- Pattern C → [Link to module](modules/pattern-c.md)

Use Cases:
- Scenario 1
- Scenario 2
- Scenario 3

Quick Syntax:
```python
# Minimal working example
result = function(params)
```
```

Example - Quality Framework:

```markdown
## Quick Reference (30 seconds)

TRUST 5 is MoAI-ADK's comprehensive quality assurance framework enforcing five pillars:

1. Test-first(T) - ≥85% coverage, ANALYZE-PRESERVE-IMPROVE cycle
2. Readable(R) - Clear naming, ≤10 cyclomatic complexity
3. Unified(U) - Consistent patterns, architecture compliance
4. Secured(S) - OWASP Top 10 compliance, security validation
5. Trackable(T) - Clear commits, requirement traceability

Integration Points:
- Pre-commit hooks → Automated validation
- CI/CD pipelines → Quality gate enforcement
- quality-gate agent → TRUST 5 validation
- /moai:2-run → Enforces ≥85% coverage

Quick Validation:
```python
validations = [
 test_coverage >= 85, # T
 complexity <= 10, # R
 consistency > 90, # U
 security_score == 100,# S
 has_clear_commits # T
]
```
```

Token Budget: ~1,000 tokens

---

### Level 2: Implementation Guide Design

Purpose: Step-by-step practical guidance for implementation (5 minutes reading).

Content Requirements:
- Detailed workflow explanations
- Pattern implementations with code
- Common scenarios with examples
- Decision trees and flowcharts
- Troubleshooting basics

Structure Template:

```markdown
## Implementation Guide (5 minutes)

### Pattern 1: [Pattern Name]

Purpose: [What this pattern achieves]

Workflow:
```
Step 1: [Action]
 ↓
Step 2: [Action]
 ↓
Step 3: [Action]
```

Implementation:
```python
# Complete working example
def implement_pattern():
 # Step 1
 result1 = step_one()
 
 # Step 2
 result2 = step_two(result1)
 
 # Step 3
 return finalize(result2)
```

Common Scenarios:

Scenario A: [Description]
```python
# Specific implementation for scenario A
solution_a()
```

Scenario B: [Description]
```python
# Specific implementation for scenario B
solution_b()
```

---

### Pattern 2: [Pattern Name]

[Repeat structure]
```

Example - DDD Workflow:

```markdown
## Implementation Guide (5 minutes)

### Phase 1: SPEC Generation

Purpose: Define clear, testable requirements in EARS format before coding.

Workflow:
```bash
# 1. Generate SPEC
/moai:1-plan "Implement user authentication with JWT tokens"

# 2. spec-builder creates:
.moai/specs/SPEC-001/
 spec.md # EARS format requirements
 acceptance.md # Acceptance criteria
 complexity.yaml # Complexity analysis

# 3. Execute /clear (mandatory)
/clear # Saves 45-50K tokens, prepares clean context
```

EARS Format Structure:
```markdown
### SPEC-001-REQ-01: User Registration (Ubiquitous)
Pattern: Ubiquitous
Statement: The system SHALL register users with email and password validation.

Acceptance Criteria:
- Email format validated (RFC 5322)
- Password strength: ≥8 chars, mixed case, numbers, symbols
- Duplicate email rejected with clear error
- Success returns user ID and confirmation email sent

Test Coverage Target: ≥90%
```

---

### Phase 2: Domain-Driven Development

ANALYZE-PRESERVE-IMPROVE Cycle:

```python
# RED: Write failing test first
def test_register_user():
 result = register_user("user@example.com", "SecureP@ssw0rd")
 assert result.success is True

# GREEN: Minimal implementation
def register_user(email, password):
 return RegistrationResult(success=True, user=User())

# REFACTOR: Improve quality
def register_user(email: str, password: str) -> RegistrationResult:
 """Register new user with email and password.
 
 Implements SPEC-001-REQ-01
 """
 # Validation, hashing, database operations
 return RegistrationResult(success=True, user=user)
```
```

Token Budget: ~3,000 tokens

---

### Level 3: Advanced Patterns Design

Purpose: Expert-level knowledge, edge cases, optimization (10+ minutes reading).

Content Requirements:
- Complex scenarios and edge cases
- Performance optimization techniques
- Integration patterns
- Architecture considerations
- Best practices and anti-patterns

Structure Template:

```markdown
## Advanced Implementation (10+ minutes)

### Advanced Pattern 1: [Complex Pattern]

Context: When to use this advanced pattern

Implementation:
```python
# Complex implementation with edge case handling
class AdvancedPattern:
 def __init__(self):
 self.setup_complex_state()
 
 def handle_edge_case_1(self):
 # Detailed edge case handling
 pass
 
 def handle_edge_case_2(self):
 # Another edge case
 pass
 
 def optimize_performance(self):
 # Performance optimization
 pass
```

Edge Cases:
- Case 1: [Description and solution]
- Case 2: [Description and solution]
- Case 3: [Description and solution]

Performance Considerations:
- Optimization technique 1
- Optimization technique 2
- Benchmarking approach

---

### Integration Patterns

[Advanced integration examples]

---

### Anti-Patterns to Avoid

Anti-Pattern 1: [Description]
```python
# BAD: Anti-pattern example
bad_implementation()
```

Solution:
```python
# GOOD: Correct implementation
good_implementation()
```
```

Token Budget: ~5,000 tokens

---

## Advanced Implementation (10+ minutes)

### 500-Line SKILL.md Limit Enforcement

Critical Rule: SKILL.md MUST be ≤500 lines.

Line Budget Breakdown:

```
SKILL.md (500 lines maximum)
 Frontmatter (4-6 lines)
 name, description, tools

 Quick Reference (80-120 lines)
 Core concepts (30-40)
 Quick access (20-30)
 Use cases (15-20)
 Quick syntax (15-30)

 Implementation Guide (180-250 lines)
 Pattern 1 (60-80)
 Pattern 2 (60-80)
 Pattern 3 (60-90)

 Advanced Patterns (80-140 lines)
 Advanced pattern 1 (40-60)
 Advanced pattern 2 (40-60)
 Edge cases (20-40)

 Works Well With (10-20 lines)
 Agents (3-5)
 Skills (3-5)
 Commands (2-4)
 Memory (2-4)
```

Overflow Handling Strategy:

```python
class SKILLMDValidator:
 """Validate and enforce 500-line SKILL.md limit."""
 
 MAX_LINES = 500
 
 def validate_skill(self, skill_path: str) -> dict:
 """Validate SKILL.md compliance."""
 
 skill_file = f"{skill_path}/SKILL.md"
 
 # Count lines
 with open(skill_file) as f:
 lines = f.readlines()
 
 line_count = len(lines)
 
 if line_count > self.MAX_LINES:
 return {
 "valid": False,
 "line_count": line_count,
 "overflow": line_count - self.MAX_LINES,
 "action": "SPLIT_REQUIRED",
 "recommendation": self._generate_split_recommendation(lines)
 }
 
 return {
 "valid": True,
 "line_count": line_count,
 "remaining": self.MAX_LINES - line_count
 }
 
 def _generate_split_recommendation(self, lines: list) -> dict:
 """Generate file splitting recommendation."""
 
 sections = self._analyze_sections(lines)
 
 recommendations = []
 
 # Check Advanced Patterns section size
 advanced = sections.get("Advanced Patterns", 0)
 if advanced > 100:
 recommendations.append({
 "target": "modules/advanced-patterns.md",
 "content": "Advanced Patterns section",
 "lines_saved": advanced - 20 # Keep brief intro
 })
 
 # Check code examples
 example_lines = self._count_code_blocks(lines)
 if example_lines > 100:
 recommendations.append({
 "target": "examples.md",
 "content": "Code examples",
 "lines_saved": example_lines - 30 # Keep key examples
 })
 
 # Check reference links
 reference_lines = self._count_references(lines)
 if reference_lines > 50:
 recommendations.append({
 "target": "reference.md",
 "content": "External references",
 "lines_saved": reference_lines - 10
 })
 
 return {
 "recommendations": recommendations,
 "total_lines_saved": sum(r["lines_saved"] for r in recommendations),
 "resulting_size": len(lines) - sum(r["lines_saved"] for r in recommendations)
 }
 
 def auto_split_skill(self, skill_path: str):
 """Automatically split SKILL.md into modules."""
 
 skill_file = f"{skill_path}/SKILL.md"
 
 with open(skill_file) as f:
 content = f.read()
 
 # Extract sections
 sections = self._extract_sections(content)
 
 # Keep core sections in SKILL.md
 core_content = {
 "frontmatter": sections["frontmatter"],
 "quick_reference": sections["quick_reference"],
 "implementation_guide": sections["implementation_guide"],
 "advanced_intro": self._create_brief_intro(sections["advanced_patterns"]),
 "works_well_with": sections["works_well_with"]
 }
 
 # Move overflow to modules
 modules_dir = f"{skill_path}/modules"
 os.makedirs(modules_dir, exist_ok=True)
 
 # Advanced patterns → modules/advanced-patterns.md
 with open(f"{modules_dir}/advanced-patterns.md", "w") as f:
 f.write(sections["advanced_patterns"])
 
 # Examples → examples.md
 if "examples" in sections:
 with open(f"{skill_path}/examples.md", "w") as f:
 f.write(sections["examples"])
 
 # References → reference.md
 if "references" in sections:
 with open(f"{skill_path}/reference.md", "w") as f:
 f.write(sections["references"])
 
 # Rewrite SKILL.md with core content + cross-references
 with open(skill_file, "w") as f:
 f.write(self._assemble_core_skill(core_content))

# Usage
validator = SKILLMDValidator()

# Validate skill
result = validator.validate_skill(".claude/skills/moai-foundation-core")

if not result["valid"]:
 print(f" SKILL.md exceeds limit: {result['line_count']} lines")
 print(f" Overflow: {result['overflow']} lines")
 print(f" Recommendation: {result['recommendation']}")
 
 # Auto-split
 validator.auto_split_skill(".claude/skills/moai-foundation-core")
 print(" Skill automatically split into modules")
```

### Progressive Loading Strategy

Token-Efficient Content Access:

```python
class ProgressiveContentLoader:
 """Load skill content progressively based on user needs."""
 
 def __init__(self, skill_path: str):
 self.skill_path = skill_path
 self.loaded_levels = set()
 
 def load_level_1(self) -> str:
 """Load Quick Reference only (30 seconds, ~1K tokens)."""
 
 if "level_1" in self.loaded_levels:
 return # Already loaded
 
 with open(f"{self.skill_path}/SKILL.md") as f:
 content = f.read()
 
 # Extract only Quick Reference section
 quick_ref = self._extract_section(content, "Quick Reference")
 
 self.loaded_levels.add("level_1")
 return quick_ref
 
 def load_level_2(self) -> str:
 """Load Implementation Guide (~3K additional tokens)."""
 
 if "level_2" not in self.loaded_levels:
 with open(f"{self.skill_path}/SKILL.md") as f:
 content = f.read()
 
 impl_guide = self._extract_section(content, "Implementation Guide")
 self.loaded_levels.add("level_2")
 return impl_guide
 
 def load_level_3(self) -> str:
 """Load Advanced Patterns (~5K additional tokens)."""
 
 if "level_3" not in self.loaded_levels:
 # Check if in SKILL.md or split to module
 advanced_path = f"{self.skill_path}/modules/advanced-patterns.md"
 
 if os.path.exists(advanced_path):
 # Load from module
 with open(advanced_path) as f:
 advanced = f.read()
 else:
 # Load from SKILL.md
 with open(f"{self.skill_path}/SKILL.md") as f:
 content = f.read()
 advanced = self._extract_section(content, "Advanced Patterns")
 
 self.loaded_levels.add("level_3")
 return advanced
 
 def load_examples(self) -> str:
 """Load examples.md if exists."""
 examples_path = f"{self.skill_path}/examples.md"
 if os.path.exists(examples_path):
 with open(examples_path) as f:
 return f.read()
 
 def load_on_demand(self, user_expertise: str, time_available: int) -> str:
 """Load appropriate level based on user context."""
 
 if time_available <= 30: # seconds
 return self.load_level_1()
 
 elif time_available <= 300: # 5 minutes
 return self.load_level_1() + "\n\n" + self.load_level_2()
 
 else: # 10+ minutes
 return (
 self.load_level_1() + "\n\n" +
 self.load_level_2() + "\n\n" +
 self.load_level_3()
 )

# Usage
loader = ProgressiveContentLoader(".claude/skills/moai-foundation-core")

# User with 30 seconds
quick_help = loader.load_level_1() # ~1K tokens

# User with 5 minutes
practical_guide = loader.load_on_demand(expertise="intermediate", time_available=300)
# ~4K tokens (Level 1 + Level 2)

# Expert user with time
comprehensive = loader.load_on_demand(expertise="expert", time_available=900)
# ~9K tokens (All levels)
```

### Cross-Reference Architecture

Effective Cross-Linking:

```markdown
<!-- In SKILL.md -->

## Quick Reference (30 seconds)

Quick Access:
- TRUST 5 Framework → [Module](modules/trust-5-framework.md)
- SPEC-First DDD → [Module](modules/spec-first-ddd.md)
- Delegation Patterns → [Module](modules/delegation-patterns.md)

Detailed Examples: [examples.md](examples.md)
External Resources: [reference.md](reference.md)

---

## Implementation Guide (5 minutes)

### Pattern 1: Quality Gates

For advanced TRUST 5 patterns, see [Advanced TRUST 5](modules/trust-5-framework.md#advanced-implementation).

---

## Advanced Patterns (10+ minutes)

Brief Introduction: Advanced patterns split to dedicated modules for depth.

Available Modules:
- [trust-5-framework.md](modules/trust-5-framework.md) - Quality assurance
- [spec-first-ddd.md](modules/spec-first-ddd.md) - Development workflow
- [delegation-patterns.md](modules/delegation-patterns.md) - Agent orchestration
- [token-optimization.md](modules/token-optimization.md) - Budget management
- [progressive-disclosure.md](modules/progressive-disclosure.md) - Content structure
- [modular-system.md](modules/modular-system.md) - File organization
```

---

## Works Well With

Skills:
- moai-foundation-modular-system - File organization patterns
- moai-foundation-token-optimization - Content efficiency
- moai-cc-skill-factory - Skill creation with progressive structure

Agents:
- skill-factory - Create skills with progressive disclosure
- docs-manager - Generate documentation with layered structure

Commands:
- /moai:1-plan - Generate SPEC with progressive detail
- /moai:3-sync - Create docs with layered structure

---

Version: 1.0.0
Last Updated: 2025-11-25
Status: Production Ready

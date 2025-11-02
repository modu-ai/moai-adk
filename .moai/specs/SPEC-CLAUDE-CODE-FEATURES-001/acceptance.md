# SPEC-CLAUDE-CODE-FEATURES-001: Acceptance Test Scenarios

## Overview
12 comprehensive test scenarios in Given-When-Then (GWT) format to validate all 6 features of Claude Code v2.0.30+ integration.

---

## Feature 1: Enhanced Code Analysis (@CODE) - Scenarios 1-2

### Test Scenario 1: Code Complexity Analysis
**Feature**: Enhanced Code Analysis
**Priority**: High
**Status**: Pending

**Given**: 
- A Python function with nested loops and conditional branches
- File path: `src/example.py`
- Function name: `process_data()`
- Code:
  ```python
  def process_data(items):
      result = []
      for item in items:
          if item > 0:
              for sub_item in item.values():
                  if sub_item.is_valid():
                      result.append(process(sub_item))
      return result
  ```

**When**: 
- Code analysis is triggered with `@CODE` tag
- Analysis engine parses the function

**Then**:
- Cyclomatic complexity score: 4
- Cognitive complexity score: 6
- Quality rating: Medium
- Optimization suggestions provided:
  - Reduce nested loops
  - Extract validation logic
  - Use list comprehension

**Expected Output**: Structured analysis report with scores and suggestions

---

### Test Scenario 2: Multi-File Code Quality Assessment
**Feature**: Enhanced Code Analysis
**Priority**: High
**Status**: Pending

**Given**:
- Multiple Python files in project:
  - `core.py` (500 lines)
  - `utils.py` (300 lines)
  - `handlers.py` (400 lines)
- All files have `@CODE` tags

**When**:
- Batch analysis is executed on all files
- Quality metrics are aggregated

**Then**:
- Individual file reports generated
- Project-wide quality score calculated
- Hotspots identified (top 3 most complex functions)
- Linting issues enumerated
- Overall project quality rating provided

**Expected Output**: Comprehensive quality dashboard

---

## Feature 2: Automated Test Generation (@TEST) - Scenarios 3-4

### Test Scenario 3: SPEC-to-Test Conversion
**Feature**: Automated Test Generation
**Priority**: High
**Status**: Pending

**Given**:
- SPEC document with EARS-format requirements:
  ```
  GIVEN a user with valid credentials
  WHEN login is attempted
  THEN access is granted and session created
  ```
- Feature branch: `feature/SPEC-AUTH-001`
- Test framework: pytest

**When**:
- Test generation is triggered
- Generator reads SPEC scenarios

**Then**:
- Test file created: `tests/test_auth.py`
- Test cases generated:
  - `test_login_with_valid_credentials()`
  - `test_login_with_invalid_password()`
  - `test_login_with_nonexistent_user()`
  - `test_session_created_after_login()`
- Tests include @TEST tags for traceability
- Tests are immediately runnable
- Parameterized tests for variations

**Expected Output**: `tests/test_auth.py` with 4+ test cases

---

### Test Scenario 4: Edge Case Identification & Testing
**Feature**: Automated Test Generation
**Priority**: High
**Status**: Pending

**Given**:
- SPEC with core requirements
- Feature implementation in `src/payment.py`:
  - `process_payment(amount, card_info)` function
  - Valid amount range: 0.01 - 999,999.99
  
**When**:
- Test generation analyzes implementation
- Edge cases are identified

**Then**:
- Generated tests include:
  - Boundary tests: `0.00`, `0.01`, `999,999.99`, `1,000,000.00`
  - Type tests: string, None, negative values
  - Format tests: invalid card formats
  - Concurrency tests: parallel payment requests
- All edge cases have corresponding test cases
- Test names clearly indicate what is being tested

**Expected Output**: `tests/test_payment.py` with 12+ test cases covering edge cases

---

## Feature 3: Documentation Sync (@DOC) - Scenarios 5-6

### Test Scenario 5: Automatic Docstring Generation
**Feature**: Documentation Sync
**Priority**: Medium
**Status**: Pending

**Given**:
- Python function without docstring:
  ```python
  def calculate_discount(price, quantity, customer_tier):
      base_discount = 0
      if quantity > 100:
          base_discount = 10
      if customer_tier == "premium":
          base_discount += 5
      return price * (1 - base_discount/100)
  ```
- File tagged with @DOC

**When**:
- Documentation sync is executed
- Docstring generator analyzes function

**Then**:
- Docstring is generated with:
  - Function description
  - Parameter documentation (type hints inferred)
  - Return value documentation
  - Example usage
  - Raises documentation if exceptions possible
- Generated docstring follows Google style guide
- Docstring is inserted into source file

**Expected Output**: Function with complete docstring

---

### Test Scenario 6: README & Changelog Synchronization
**Feature**: Documentation Sync
**Priority**: Medium
**Status**: Pending

**Given**:
- Implemented feature: `calculate_discount()` in `src/pricing.py`
- File tagged with @DOC
- Existing `README.md` with API section
- Existing `CHANGELOG.md`
- New version: 0.8.0

**When**:
- Documentation sync is triggered
- Changes are detected and documented

**Then**:
- `README.md` updated:
  - New function added to API section
  - Example usage included
  - Return to table of contents works
- `CHANGELOG.md` updated:
  - Entry: "Add `calculate_discount()` for flexible pricing"
  - Version: 0.8.0
  - Date: Current date
  - Category: Features
- Links are validated
- Formatting is consistent

**Expected Output**: Updated README.md and CHANGELOG.md

---

## Feature 4: Git Workflow Automation (GitFlow) - Scenarios 7-8

### Test Scenario 7: Branch Naming Validation
**Feature**: Git Workflow Automation
**Priority**: High
**Status**: Pending

**Given**:
- Feature branch naming convention: `feature/SPEC-{ID}`
- Bugfix branch naming convention: `bugfix/SPEC-{ID}`
- Release branch naming convention: `release/v{VERSION}`
- Invalid branch names:
  - `my-feature` (missing SPEC reference)
  - `feature/my-awesome-feature` (no SPEC ID)
  - `Feature/SPEC-001` (uppercase F)

**When**:
- Developer attempts to push invalid branch
- GitFlow validator is triggered

**Then**:
- Invalid branch: Push is blocked
- Error message displayed: "Branch must follow convention: feature/SPEC-{ID}"
- Valid branch: `feature/SPEC-001` is accepted
- Valid branch: `feature/SPEC-CLAUDE-001` is accepted
- Pre-commit hook prevents local commit on invalid branch

**Expected Output**: Branch validation enforcement via pre-commit hook

---

### Test Scenario 8: Commit Message Linting
**Feature**: Git Workflow Automation
**Priority**: High
**Status**: Pending

**Given**:
- Commit message convention:
  - Emoji prefix: 🔴 (RED/test), 🟢 (GREEN/impl), ♻️ (REFACTOR)
  - Format: `emoji MESSAGE with @TAG reference`
  - Example: `🔴 RED: Test login validation with @TEST:COMMIT-001-RED`
- Invalid messages:
  - `Fixed bug in auth` (no emoji)
  - `red test login` (wrong format)
  - `🔴 test` (no tag reference)

**When**:
- Developer commits with invalid message
- Commit message linter runs

**Then**:
- Invalid message: Commit is rejected
- Error message: "Commit must include emoji, message, and @TAG reference"
- Valid message: `🔴 RED: Test login validation with @TEST:COMMIT-001-GREEN` is accepted
- Suggestion provided for fixing message

**Expected Output**: Enforced commit message format via pre-commit hook

---

## Feature 5: SPEC-First Development (@SPEC) - Scenarios 9-10

### Test Scenario 9: SPEC-to-Code Traceability
**Feature**: SPEC-First Development
**Priority**: High
**Status**: Pending

**Given**:
- SPEC: `SPEC-TRACE-001` with requirement:
  - REQ-TRACE-001: "System shall validate user credentials"
  - REQ-TRACE-002: "System shall create session on successful login"
- Code implementation:
  ```python
  def validate_credentials(username, password):  # @SPEC:TRACE-001-REQ-001
      pass

  def create_session(user_id):  # @SPEC:TRACE-001-REQ-002
      pass
  ```

**When**:
- Traceability matrix is generated
- Code is analyzed for @SPEC tags

**Then**:
- Traceability matrix shows:
  - REQ-TRACE-001 → validate_credentials() ✓
  - REQ-TRACE-002 → create_session() ✓
  - All requirements have corresponding code
  - All code references requirements
- Missing references identified
- Orphaned code flagged

**Expected Output**: Traceability matrix in `SPEC-TRACE-001/traceability.md`

---

### Test Scenario 10: SPEC Compliance Validation
**Feature**: SPEC-First Development
**Priority**: High
**Status**: Pending

**Given**:
- SPEC: `SPEC-PAYMENT-001` with 5 acceptance criteria
- Implementation covering:
  - 4/5 criteria implemented and tested
  - 1/5 criteria pending
- Test coverage: 80%

**When**:
- SPEC compliance check is executed
- Coverage analysis runs

**Then**:
- Compliance report shows:
  - 4/5 criteria met: ✓
  - 1/5 criteria pending: ⏳
  - Overall completion: 80%
  - Missing implementation flagged
  - Missing tests identified
- Recommendation: "Implement criterion 5 and add tests before merge"
- PR approval blocked until compliance reaches 100%

**Expected Output**: Compliance report blocking non-compliant merge

---

## Feature 6: Checkpoint & Rollback System - Scenarios 11-12

### Test Scenario 11: Checkpoint Creation & Recovery
**Feature**: Checkpoint & Rollback System
**Priority**: High
**Status**: Pending

**Given**:
- Working directory with experimental code
- Current state: Feature partially implemented
- Checkpoint creation: `git checkpoint "WIP: Auth feature"`
- File changes made after checkpoint:
  - `src/auth.py` modified
  - `tests/test_auth.py` added
  - `config.json` modified

**When**:
- Rollback to checkpoint is executed
- `git rollback <checkpoint-hash>`

**Then**:
- Working directory state restored to checkpoint
- All post-checkpoint changes reverted:
  - `src/auth.py` restored to checkpoint version
  - `tests/test_auth.py` deleted
  - `config.json` restored
- Checkpoint history preserved in Git tags
- No data loss
- Timestamp in checkpoint tag: `moai_cp/20251102_144530`

**Expected Output**: Working directory reverted to checkpoint state

---

### Test Scenario 12: Checkpoint Management & History
**Feature**: Checkpoint & Rollback System
**Priority**: Medium
**Status**: Pending

**Given**:
- Repository with 5 checkpoints:
  - `moai_cp/20251102_100000` - "Start implementation"
  - `moai_cp/20251102_110000` - "RED tests passing"
  - `moai_cp/20251102_120000` - "GREEN implementation"
  - `moai_cp/20251102_130000` - "REFACTOR code cleanup"
  - `moai_cp/20251102_140000` - "All tests passing"
- Current time: 20251102_150000

**When**:
- Checkpoint list command is executed
- `git checkpoint list` or Alfred checkpoint management

**Then**:
- Checkpoint history displayed:
  - Shows all 5 checkpoints with timestamps
  - Shows checkpoint messages
  - Shows relative time (2 hours ago, 1 hour ago, etc.)
  - Shows which checkpoint is current (if any)
- Search functionality: `git checkpoint list --grep "test"`
- Latest checkpoint highlighted
- Diff between checkpoints available: `git checkpoint diff cp1 cp2`

**Expected Output**: Formatted checkpoint history with management options

---

## Test Execution Plan

### Phase 1: Unit Tests (Week 1-2)
- Scenarios 1-2: Code analysis
- Scenarios 3-4: Test generation
- Run: `pytest tests/test_features.py -v`

### Phase 2: Integration Tests (Week 3)
- Scenarios 5-12: All features
- Run: `pytest tests/test_integration.py -v`

### Phase 3: E2E Tests (Week 4)
- All 12 scenarios with real workflows
- Run: `pytest tests/test_e2e.py -v`

### Success Criteria
- All 12 scenarios passing
- Code coverage >= 85%
- Performance targets met
- Zero breaking changes
- Documentation complete

---

## Sign-Off

- **Acceptance Owner**: GOOS
- **Test Reviewer**: Alfred SuperAgent
- **Quality Gate**: TRUST 5 Principles
- **Target Completion**: End of Week 4

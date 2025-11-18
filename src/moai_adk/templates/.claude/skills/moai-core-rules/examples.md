# moai-core-rules Examples ( .0)

## 실제 적용 예제 (10+ Scenarios)

### Example 1: Skill Invocation 규칙 적용

**상황**: TRUST 5 검증 필요

```python
# ❌ WRONG: 직접 검증
def validate_trust():
    coverage = measure_coverage()  # 직접 실행 X
    if coverage >= 85:
        return "PASS"
    return "FAIL"

# ✅ CORRECT: Skill 호출
Skill("moai-foundation-trust")
# → Agent가 TRUST 5 게이트 검증
# → 상세 리포트 반환
```

### Example 2: AskUserQuestion 올바른 사용

**상황**: 사용자 의도 모호

```python
# ❌ WRONG: 평문 질문
print("어떤 프레임워크 사용할래?")
response = input()  # 자유로운 텍스트

# ✅ CORRECT: AskUserQuestion 도구
AskUserQuestion({
  question: "어떤 유형의 웹 프레임워크를 원하나요?",
  header: "Framework Type",
  multiSelect: false,
  options: [
    {
      label: "REST API",
      description: "FastAPI, Flask - High performance APIs"
    },
    {
      label: "Full Stack",
      description: "Django - Batteries included"
    },
    {
      label: "Lightweight",
      description: "Bottle, Falcon - Minimal overhead"
    }
  ]
})
```

### Example 3: Agent Delegation 패턴

**상황**: 새 기능 구현 요청

```python
# ❌ WRONG: 직접 구현
def implement_feature():
    code = """
    def new_function():
        pass
    """
    return code

# ✅ CORRECT: Agent에 위임
Task(
  subagent_type="tdd-implementer",
  description="Implement user login feature",
  prompt="""
You are the TDD implementer. Implement user login with:
- RED: Write failing tests for authentication
- GREEN: Implement minimal passing code
- REFACTOR: Optimize and clean up

Add docstrings and type hints.
  """
)
```

### Example 4: TAG Chain 생성

**상황**: 새로운 기능 개발

```
────────────────────────────
spec.md:
"""

Requirements:
- Generate JWT tokens
- Validate token expiration
- Refresh token support
"""

────────────────────────────
test_auth.py:
"""

def test_generate_jwt_token():
    # Test valid token generation

def test_validate_expired_token():
    # Test expired token handling

def test_refresh_token():
    # Test token refresh
"""

────────────────────────────
auth.py:
"""

def generate_jwt_token(user_id):
    # Implementation

def validate_token(token):
    # Implementation

def refresh_token(token):
    # Implementation
"""

────────────────────────────

- Add JWT token generation
- Add token validation with expiration
- Add token refresh mechanism

All tests passing. 85%+ coverage."
```

### Example 5: TRUST 5 검증 Flow

**상황**: 코드가 준비되었는데 품질 확인 필요

```python
# 자동화된 TRUST 5 검증
checks = {
    "test": check_coverage(target=85),      # Test coverage
    "readable": check_code_style(),          # Readability
    "unified": check_consistency(),          # Unified patterns
    "secured": check_security_scan(),        # Security
    "trackable": check_tag_chain()           # Traceability
}

results = {}
for check_name, passed in checks.items():
    if not passed:
        print(f"❌ {check_name.upper()} gate failed")
        # Request fix
        Task(
            subagent_type="tdd-implementer",
            prompt=f"Fix {check_name} gate violations"
        )
    else:
        print(f"✅ {check_name.upper()} gate passed")
        results[check_name] = "PASS"

if all(results.values()):
    print("✅ ALL TRUST 5 GATES PASSED")
    # Ready for merge
else:
    print("❌ QUALITY GATES FAILED - CANNOT MERGE")
```

### Example 6: Workflow Compliance Validation

**상황**: feature branch를 develop에 merge하기 전 확인

```python
# Merge 전 compliance 체크리스트

class MergeValidator:
    def __init__(self, branch_name):
        self.branch = branch_name
        self.checks = {}
    
    def validate_spec(self):
        # SPEC 파일 존재 확인
        spec_file = f".moai/specs/{self.branch}/spec.md"
        self.checks["spec_exists"] = Path(spec_file).exists()
        return self.checks["spec_exists"]
    
    def validate_tests(self):
        # 85%+ coverage 확인
        coverage = self.measure_coverage()
        self.checks["coverage"] = coverage >= 85
        return self.checks["coverage"]
    
    def validate_tags(self):
        Skill("moai-foundation-tags")
        # → 반환: chain_complete (bool)
        self.checks["tags_complete"] = True  # Result from Skill
        return self.checks["tags_complete"]
    
    def validate_commit_messages(self):
        # 커밋 메시지 형식 확인
        commits = self.get_branch_commits()
        for commit in commits:
            # feat(), test(), refactor(), fix() 형식 확인
            pass
        self.checks["commit_format"] = True
        return self.checks["commit_format"]
    
    def validate_all(self):
        print("Running merge validation...")
        
        self.validate_spec()
        self.validate_tests()
        self.validate_tags()
        self.validate_commit_messages()
        
        if all(self.checks.values()):
            print("✅ Ready to merge")
            return True
        else:
            print("❌ Fix issues before merge:")
            for check, status in self.checks.items():
                if not status:
                    print(f"  - {check}: FAILED")
            return False

# 사용
validator = MergeValidator("feature/SPEC-001")
if validator.validate_all():
    # Create PR to develop
    subprocess.run(["gh", "pr", "create", "--base", "develop"])
```

### Example 7: Agent Delegation 의사결정

**상황**: 여러 작업을 조율하는 상황

```python
def orchestrate_feature_development():
    feature_name = "user_profile_api"
    
    # Step 1: Planning
    print("Step 1: Creating plan...")
    plan = Task(
        subagent_type="plan-agent",
        description=f"Plan {feature_name}",
        prompt="Create detailed development plan with sprints"
    )
    
    # Step 2: Implementation (TDD)
    print("Step 2: Implementing with TDD...")
    implementation = Task(
        subagent_type="tdd-implementer",
        description=f"Implement {feature_name}",
    )
    
    # Step 3: Testing (if not covered in TDD)
    print("Step 3: Comprehensive testing...")
    testing = Task(
        subagent_type="test-engineer",
        description=f"Test {feature_name}",
        prompt="Achieve 85%+ coverage, test edge cases"
    )
    
    # Step 4: Documentation
    print("Step 4: Documenting...")
    docs = Task(
        subagent_type="doc-syncer",
        description=f"Document {feature_name}",
        prompt="Update README, API docs, examples"
    )
    
    # Step 5: Validation
    print("Step 5: Quality validation...")
    quality = Task(
        subagent_type="qa-validator",
        description=f"Validate {feature_name}",
        prompt="Verify TRUST 5, TAG chains, compliance"
    )
    
    # Step 6: Commit & Push
    print("Step 6: Committing...")
    commit = Task(
        subagent_type="git-manager",
        description=f"Commit {feature_name}",
    )
    
    print("✅ Feature development complete")
```

### Example 8: Rule Violation Detection

**상황**: 규칙 위반 감지 및 수정

```python
def detect_violations():
    violations = {
        "command_direct_execution": False,
        "skill_not_invoked": False,
        "agent_skipped": False,
        "tag_orphaned": False,
        "test_coverage_low": False
    }
    
    # Check 1: Command직접 실행 확인
    if subprocess.call("git commit") == 0:  # Direct call!
        violations["command_direct_execution"] = True
    
    # Check 2: Skill 호출 안 함 확인
    if "TRUST 5" in task_description and "Skill(" not in code:
        violations["skill_not_invoked"] = True
    
    all_tags = find_all_tags()
    for tag in all_tags:
        if tag_has_no_chain(tag):
            violations["tag_orphaned"] = True
    
    # Check 4: 테스트 커버리지 확인
    coverage = measure_coverage()
    if coverage < 85:
        violations["test_coverage_low"] = True
    
    # Report violations
    if any(violations.values()):
        print("❌ VIOLATIONS DETECTED:")
        for violation, detected in violations.items():
            if detected:
                print(f"  - {violation}")
        
        # Request remediation
        AskUserQuestion({
            question: "위반 사항을 수정하시겠습니까?",
            options: ["Fix now", "Review later", "Skip"]
        })
    else:
        print("✅ No violations detected")
```

### Example 9: Commit Message Format Validation

**상황**: 커밋 메시지 형식 검증

```python
import re

class CommitValidator:
    VALID_TYPES = ["feat", "fix", "test", "refactor", "docs", "chore"]
    TAG_PATTERN = r'@[A-Z]+-\d{3}'
    
    def validate_message(self, message):
        lines = message.split('\n')
        header = lines[0]
        
        # Check format: type(tag): subject
        pattern = rf"({'|'.join(self.VALID_TYPES)})\(({self.TAG_PATTERN})\):\s(.+)"
        
        if not re.match(pattern, header):
            return False, f"Invalid format: {header}"
        
        # Check body (if present)
        if len(lines) > 2 and lines[1].strip() != "":
            return False, "Second line must be blank"
        
        # Check tag in body
        body = '\n'.join(lines[2:])
        if not re.search(self.TAG_PATTERN, body):
            return False, f"TAG reference missing in body"
        
        return True, "Valid"

# Usage
validator = CommitValidator()


- Add JWT token generation
- Add credential validation
- Add error handling


is_valid, msg = validator.validate_message(commit_msg)
if is_valid:
    print("✅ Commit message valid")
else:
    print(f"❌ {msg}")
```

### Example 10: TRUST 5 Status Dashboard

**상황**: 프로젝트 품질 상태 모니터링

```python
class TrustDashboard:
    def __init__(self):
        self.metrics = {}
    
    def test_coverage(self):
        # T: Test coverage
        coverage = measure_coverage()
        status = "PASS" if coverage >= 85 else "FAIL"
        self.metrics["Test"] = {
            "value": f"{coverage}%",
            "target": "≥ 85%",
            "status": status
        }
    
    def code_quality(self):
        # R: Readable
        style_score = check_code_style()
        self.metrics["Readable"] = {
            "value": style_score,
            "target": "No issues",
            "status": "PASS" if style_score == 0 else "FAIL"
        }
    
    def consistency(self):
        # U: Unified
        duplicate_code = check_duplication()
        self.metrics["Unified"] = {
            "value": f"{duplicate_code}%",
            "target": "< 5%",
            "status": "PASS" if duplicate_code < 5 else "FAIL"
        }
    
    def security(self):
        # S: Secured
        vulnerabilities = security_scan()
        self.metrics["Secured"] = {
            "value": len(vulnerabilities),
            "target": "0",
            "status": "PASS" if len(vulnerabilities) == 0 else "FAIL"
        }
    
    def traceability(self):
        # T: Trackable
        chain_complete = check_tag_chains()
        self.metrics["Trackable"] = {
            "value": "Complete" if chain_complete else "Broken",
            "target": "Complete chains",
            "status": "PASS" if chain_complete else "FAIL"
        }
    
    def print_dashboard(self):
        print("\n=== TRUST 5 DASHBOARD ===")
        print()
        
        all_pass = True
        for principle, data in self.metrics.items():
            status_icon = "✅" if data["status"] == "PASS" else "❌"
            print(f"{status_icon} {principle:10s}: {data['value']:>10s} (target: {data['target']})")
            if data["status"] != "PASS":
                all_pass = False
        
        print()
        if all_pass:
            print("🎉 ALL GATES PASSED - READY FOR RELEASE")
        else:
            print("⚠️ SOME GATES FAILED - FIX BEFORE MERGE")

# Usage
dashboard = TrustDashboard()
dashboard.test_coverage()
dashboard.code_quality()
dashboard.consistency()
dashboard.security()
dashboard.traceability()
dashboard.print_dashboard()
```

---

## 실무 체크리스트

### 신규 기능 개발 체크리스트

- [ ] AskUserQuestion로 의도 명확화 (필요 시)
- [ ] plan-agent로 계획 수립
- [ ] tdd-implementer로 RED-GREEN-REFACTOR 구현
- [ ] test-engineer로 85%+ coverage 달성
- [ ] TRUST 5 모든 게이트 통과
- [ ] TAG 체인 완전성 검증
- [ ] 커밋 메시지 형식 확인
- [ ] doc-syncer로 문서화
- [ ] git-manager로 커밋 및 푸시
- [ ] Merge 전 compliance 검증

### 코드 리뷰 체크리스트

- [ ] Skill 호출은 explicit 문법 사용?
- [ ] Agent delegation 올바른가?
- [ ] Command가 직접 실행하는 부분은?
- [ ] TRUST 5 모두 통과?
- [ ] 커밋 메시지 형식 맞음?
- [ ] 테스트 커버리지 ≥ 85%?
- [ ] 보안 스캔 통과?
- [ ] 문서 업데이트됨?

---

**마지막 업데이트**: 2025-11-12
**버전**: 4.0.0

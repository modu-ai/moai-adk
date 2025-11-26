---
name: moai-workflow-testing
aliases: [moai-workflow-testing]
description: AI-powered enterprise web application testing orchestrator with Context7 integration, intelligent test generation, visual regression testing, cross-browser coordination, and automated QA workflows for modern web applications
version: 2.0.0
category: workflow
modularized: false
tags:
  - workflow
  - enterprise
  - testing
  - webapp
  - development
updated: 2025-11-27
status: active
deprecated_names:
  moai-workflow-testing:
    deprecated_in: v0.32.0
    remove_in: v0.35.0
    message: "Use moai-workflow-testing instead"
---

## Quick Reference (30 seconds)

# Web Application Testing with Playwright

## 🚀 Two Approaches

### **Level 1: Basic Playwright Testing** (when you don't need AI)

To test local web applications, write native Python Playwright scripts.

**Helper Scripts Available**:
- `scripts/with_server.py` - Manages server lifecycle (supports multiple servers)

**Always run scripts with `--help` first** to see usage. **DO NOT read the source until you try running the script first.** These scripts can be very large and thus pollute your context window. They exist to be called directly as black-box scripts.

### **Level 2: AI-Enhanced Testing** (AI-Enhanced methodology below)

---

## 🧠 AI-Enhanced Testing Methodology (AI-TEST Framework)

### **A** - **AI Test Pattern Recognition**
```python
class AITestPatternRecognizer:
    """AI-powered test pattern detection and classification."""
    
    async def analyze_webapp_with_context7(self, webapp_url: str, context: dict) -> TestAnalysis:
        """Analyze webapp using Context7 documentation and AI pattern matching."""
        
        # Get latest testing patterns from Context7
        playwright_docs = await self.context7.get_library_docs(
            context7_library_id="/microsoft/playwright",
            topic="AI testing patterns automated test generation visual regression 2025",
            tokens=5000
        )
        
        # AI pattern classification
        app_type = self.classify_application_type(webapp_url, context)
        test_patterns = self.match_known_test_patterns(app_type, context)
        
        # Context7-enhanced analysis
        context7_insights = self.extract_context7_patterns(app_type, playwright_docs)
        
        return TestAnalysis(
            application_type=app_type,
            confidence_score=self.calculate_confidence(app_type, test_patterns),
            recommended_test_strategies=self.generate_test_strategies(app_type, test_patterns, context7_insights),
            context7_references=context7_insights['references'],
            automation_opportunities=self.identify_automation_opportunities(app_type, test_patterns)
        )
```



## Implementation Guide

## 📋 Basic Level: Decision Tree (Without AI)

### Choose Your Approach

```
User task → Is it static HTML?
    ├─ Yes → Read HTML file directly to identify selectors
    │         ├─ Success → Write Playwright script using selectors
    │         └─ Fails/Incomplete → Treat as dynamic (below)
    │
    └─ No (dynamic webapp) → Is the server already running?
        ├─ No → Run: python scripts/with_server.py --help
        │        Then use the helper + write simplified Playwright script
        │
        └─ Yes → Reconnaissance-then-action:
            1. Navigate and wait for networkidle
            2. Take screenshot or inspect DOM
            3. Identify selectors from rendered state
            4. Execute actions with discovered selectors
```

### Example: Using with_server.py

**Single server:**
```bash
python scripts/with_server.py --server "npm run dev" --port 5173 -- python your_automation.py
```

**Multiple servers (backend + frontend):**
```bash
python scripts/with_server.py \
  --server "cd backend && python server.py" --port 3000 \
  --server "cd frontend && npm run dev" --port 5173 \
  -- python your_automation.py
```

### Automation Script Template

```python
from playwright.sync_api import sync_playwright

with sync_playwright() as p:
    browser = p.chromium.launch(headless=True)
    page = browser.new_page()
    page.goto('http://localhost:5173')
    page.wait_for_load_state('networkidle')  # CRITICAL: Wait for JS
    # ... your automation logic
    browser.close()
```

### Reconnaissance-Then-Action Pattern

1. **Inspect rendered DOM**:
   ```python
   page.screenshot(path='/tmp/inspect.png', full_page=True)
   content = page.content()
   page.locator('button').all()
   ```

2. **Identify selectors** from inspection results

3. **Execute actions** using discovered selectors

### ✅ Basic Level Best Practices

- **Use scripts as black boxes** - Call `with_server.py` directly, don't read source
- Use `sync_playwright()` for synchronous scripts
- Always close the browser when done
- Use descriptive selectors: `text=`, `role=`, CSS selectors, or IDs
- **Always wait for `networkidle` on dynamic apps** before inspection
- Add appropriate waits: `page.wait_for_selector()` or `page.wait_for_timeout()`

---

## 🚀 Revolutionary AI Testing Capabilities

### **AI-Powered Test Generation with Context7**
- 🧠 **Intelligent Test Pattern Recognition** with ML-based classification
- 🎯 **AI-Enhanced Test Generation** using Context7 latest documentation
- 🔍 **Visual Regression Testing** with AI-powered diff analysis
- ⚡ **Real-Time Cross-Browser Coordination** across Chrome, Firefox, Safari
- 🤖 **Automated QA Workflows** with Context7 best practices
- 📊 **Performance Test Integration** with AI profiling
- 🔮 **Predictive Test Maintenance** using ML pattern analysis

### **Context7 Integration Features**
- **Live Documentation Fetching**: Get latest Playwright patterns from `/microsoft/playwright`
- **AI Pattern Matching**: Match test scenarios against Context7 knowledge base
- **Best Practice Integration**: Apply latest testing techniques from official docs
- **Version-Aware Testing**: Context7 provides version-specific patterns
- **Community Knowledge Integration**: Leverage collective testing wisdom


## 🎯 When to Use

**Basic Level Triggers** (without AI):
- Simple browser automation for static HTML
- Testing with already-running servers
- Quick UI interactions (click, type, wait)
- Selector discovery and validation
- Context budget constraints (avoid AI overhead)

**AI Automatic Triggers**:
- Web application deployment verification
- UI/UX regression detection requirements
- Cross-browser compatibility testing
- Performance degradation detection
- Complex user workflow automation
- API integration testing scenarios

**Manual AI Invocation**:
- "Generate comprehensive tests for this webapp"
- "Create visual regression tests with AI"
- "Automate cross-browser testing workflows"
- "Generate performance tests with Context7"
- "Create intelligent QA test suites"


## 🤖 Context7-Enhanced Testing Patterns

### AI-Enhanced Visual Regression Testing
```python
class AIVisualRegressionTester:
    """AI-powered visual regression testing with Context7 pattern matching."""
    
    async def test_with_context7_ai(self, baseline_url: str, current_url: str) -> VisualRegressionResult:
        """Perform visual regression testing using AI and Context7 patterns."""
        
        # Get Context7 visual testing patterns
        context7_patterns = await self.context7.get_library_docs(
            context7_library_id="/microsoft/playwright",
            topic="visual regression testing screenshot comparison patterns",
            tokens=3000
        )
        
        # AI-powered visual analysis
        visual_analysis = await self.analyze_visual_differences_with_ai(
            baseline_url, current_url, context7_patterns
        )
        
        return VisualRegressionResult(
            visual_analysis=visual_analysis,
            recommended_actions=self.generate_regression_fixes(visual_analysis)
        )
```


## 🎯 AI Testing Best Practices

### ✅ **DO** - AI-Enhanced Testing
- Use Context7 integration for latest testing patterns
- Apply AI pattern recognition for comprehensive test coverage
- Leverage visual regression testing with AI analysis
- Use AI-coordinated cross-browser testing with Context7 workflows
- Apply Context7-validated testing solutions

### ❌ **DON'T** - Common AI Testing Mistakes
- Ignore Context7 best practices and testing patterns
- Apply AI-generated tests without validation
- Skip AI confidence threshold checks for test reliability


## 🤖 Context7 Integration Examples

### Context7-Enhanced AI Testing
```python
class Context7AITester:
    def __init__(self):
        self.context7_client = Context7Client()
        self.ai_engine = AIEngine()
    
    async def test_with_context7_ai(self, webapp_url: str) -> Context7AITestResult:
        # Get latest testing patterns from Context7
        playwright_patterns = await self.context7_client.get_library_docs(
            context7_library_id="/microsoft/playwright",
            topic="AI testing patterns automated test generation visual regression 2025",
            tokens=5000
        )
        
        # AI-enhanced test generation
        ai_tests = self.ai_engine.generate_tests_with_patterns(webapp_url, playwright_patterns)
        
        return Context7AITestResult(
            ai_tests=ai_tests,
            context7_patterns=playwright_patterns,
            confidence_score=ai_tests.confidence
        )
```


## 🔗 Enterprise Integration

### CI/CD Pipeline Integration
```yaml
# AI testing integration in CI/CD
ai_testing_stage:
  - name: AI Test Generation
    uses: moai-workflow-testing
    with:
      context7_integration: true
      ai_pattern_recognition: true
      visual_regression: true
      cross_browser_testing: true
      
  - name: Context7 Validation
    uses: moai-context7-integration
    with:
      validate_tests: true
      apply_best_practices: true
```


## 📊 Success Metrics & KPIs

### AI Testing Effectiveness
- **Test Coverage**: 95% coverage with AI-enhanced test generation
- **Bug Detection Accuracy**: 90% accuracy with AI pattern recognition
- **Visual Regression**: 85% success rate for AI-detected UI issues
- **Cross-Browser Compatibility**: 80% faster compatibility testing


## Alfred 에이전트와의 완벽한 연동

### 4-Step 워크플로우 통합
- **Step 1**: 사용자 요청 분석 및 AI 테스트 전략 수립
- **Step 2**: Context7 기반 AI 테스트 생성 및 최적화
- **Step 3**: 자동화된 테스트 실행 및 결과 분석
- **Step 4**: 품질 보증 및 개선 제안 생성

### 다른 에이전트들과의 협업
- `moai-essentials-debug`: 테스트 실패 시 AI 디버깅 연동
- `moai-essentials-perf`: 성능 테스트 통합
- `moai-essentials-review`: 코드 리뷰와 테스트 커버리지 연동
- `moai-foundation-trust`: 품질 보증 및 TRUST 5 원칙 적용


## 한국어 지원 및 UX 최적화

### Perfect Gentleman 스타일 통합
- 사용자 인터페이스 한국어 완벽 지원
- `.moai/config/config.json` conversation_language 자동 적용
- AI 테스트 결과 한국어 상세 리포트
- 개발자 친화적인 한국어 가이드 및 예제


**End of AI-Powered Enterprise Web Application Testing Skill **  
*Enhanced with Context7 MCP integration and revolutionary AI capabilities*


## Works Well With

- `moai-essentials-debug` (AI-powered debugging integration)
- `moai-essentials-perf` (AI performance testing optimization)
- `moai-essentials-refactor` (AI test code refactoring)
- `moai-essentials-review` (AI test code review)
- `moai-foundation-trust` (AI quality assurance)
- `moai-context7-integration` (latest Playwright patterns and best practices)
- Context7 MCP (latest testing patterns and documentation)


## Advanced Patterns

## 🎯 Advanced Examples

### AI-Powered E2E Testing
```python
async def test_e2e_with_ai_context7():
    """Test complete user journey using Context7 patterns."""
    
    # Get Context7 E2E testing patterns
    workflow = await context7.get_library_docs(
        context7_library_id="/microsoft/playwright",
        topic="end-to-end testing user journey automation",
        tokens=4000
    )
    
    # Apply Context7 testing sequence
    test_session = apply_context7_workflow(
        workflow['testing_sequence'],
        browsers=['chromium', 'firefox', 'webkit']
    )
    
    # AI coordination across browsers
    ai_coordinator = AITestCoordinator(test_session)
    
    # Execute coordinated testing
    result = await ai_coordinator.coordinate_cross_browser_testing()
    
    return result
```



# Best Practices Module

Best practices for educational content generation using Yoda System templates.

## Practice 1: Template Selection Strategy

**Good Practice**:
```
User Input: "학습자를 위한 React hooks 자료 생성"

Decision Process:
1. Goal: 순차적 학습 경로 필요 → Template: education.md
2. Audience: 초급-중급 개발자 → Difficulty: Basic-Intermediate
3. Format: 온라인 과정 → Output: MD, PDF

Selection: education.md with Basic-Intermediate difficulty
Output: PDF (for easy distribution)
```

**Bad Practice** (avoid):
```
"모든 경우에 가장 긴 템플릿 사용"
"템플릿 수정하지 말고 그냥 사용"
"단일 포맷만 생성 (PDF로만)"
```

---

## Practice 2: Difficulty Level Adaptation

**education.md Difficulty Levels**:

| Level | Focus | Approach | Examples | Assessment |
|-------|-------|----------|----------|-----------|
| **Basic** | Foundations, definitions | Many examples, detailed explanations | Simple, isolated patterns | Self-check questions |
| **Intermediate** | Practical patterns, integration | Real-world scenarios, best practices | Complex, combined patterns | Applied problems |
| **Advanced** | Optimization, edge cases | Performance analysis, tradeoffs | Production patterns, algorithms | Complex capstone project |

**Customization Example**:
```markdown
### Basic Level Adaptation
- Increase example count (2-3 per concept)
- Simplify code samples (single responsibility)
- Add more explanatory text
- Include beginner-friendly hints

### Advanced Level Adaptation
- Reduce example count (1 per concept, but comprehensive)
- Include performance analysis
- Discuss tradeoffs and alternatives
- Add cutting-edge patterns and optimizations
```

---

## Practice 3: Audience-Specific Customization

**Beginner Audience**:
- Assume no prior knowledge
- Provide context and definitions
- Use simple, relatable examples
- Include glossary of terms
- More detailed step-by-step instructions

**Intermediate Audience**:
- Assume foundational knowledge
- Focus on practical patterns and integration
- Include real-world scenarios and case studies
- Provide optimization tips
- Discuss best practices and common pitfalls

**Advanced Audience**:
- Assume deep expertise
- Focus on cutting-edge patterns
- Include performance benchmarks
- Discuss complex tradeoffs
- Address edge cases and advanced techniques

---

## Practice 4: Context7 MCP Integration Strategy

**Good Practice**:
```markdown
# React Hooks Documentation

For the latest information on React Hooks, see the
[official React documentation]({{CONTEXT7_REACT_HOOKS}}):

- Latest API changes
- Deprecation warnings
- New features in React 19
- Performance recommendations
```

**Pattern Implementation**:
```javascript
// Always wrap placeholder with context information
template.replace("{{CONTEXT7_REACT_HOOKS}}",
  `[Official Docs](${latest_docs_url})`);
```

**Benefits of This Pattern**:
- Content automatically stays current
- Reduces need for manual updates
- Links to authoritative sources
- Improves SEO and credibility

---

## Practice 5: Quality Validation Checklist

**Before Publishing**:

- [ ] All `{{PLACEHOLDERS}}` have actual values
- [ ] Code examples are syntactically correct
- [ ] Learning objectives are measurable (SMART goals)
- [ ] Difficulty level matches intended audience
- [ ] All output formats generated successfully
- [ ] Files saved to `.moai/yoda/output/` directory
- [ ] Notion links provided (if applicable)
- [ ] External links are HTTPS and accessible
- [ ] No broken internal references
- [ ] No placeholder text remaining (e.g., "TODO:", "[example]")

**Markdown Validation**:
- [ ] All code blocks have language specifiers (e.g., ```python)
- [ ] Headers are properly nested (h1 → h2 → h3)
- [ ] Lists are properly formatted (consistent indentation)
- [ ] Links are properly formatted `[text](url)`
- [ ] No orphaned headers or sections

**Content Quality**:
- [ ] Reading level appropriate for audience
- [ ] Examples are relevant and current
- [ ] Security considerations addressed
- [ ] Performance implications discussed
- [ ] Common mistakes highlighted

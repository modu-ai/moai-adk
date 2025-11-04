# Reference

## Output Format Guidelines

### Screen Output (User-Facing)
- Use status indicators: ✅ ⚠️ ❌ ⏳
- Keep concise and scannable
- Focus on progress and next steps
- Use plain text, not markdown formatting

### Internal Documentation (Files)
- Use markdown for structure
- Include detailed analysis and rationale
- Store in proper `.moai/` directories
- Follow document management rules

## Location Rules

| Output Type | Location | Format | When to Use |
|-------------|----------|--------|-------------|
| Progress Updates | Screen | Plain text | Real-time status |
| Analysis Results | `.moai/docs/` or `.moai/analysis/` | Markdown | Detailed findings |
| Sync Reports | `.moai/reports/` | Markdown | Workflow completion |
| User Documentation | Project root | Markdown | Official docs only |

## Content Guidelines

### Screen Output Structure
1. Status summary with indicators
2. Key findings or decisions
3. Next steps or action items
4. Keep under 10 lines when possible

### File Documentation Structure
1. Clear title and purpose
2. Detailed analysis with rationale
3. Technical specifications
4. References and next steps

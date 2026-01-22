# Review Mode Context

# Context injection mode for code review and analysis

# Version: 1.0.0

# Last Updated: 2026-01-22

---

mode: "review"
name: "Review Mode"
description: "Quality-focused context for code review and analysis"

# Primary Focus

focus:

- "Code quality assessment"
- "Security vulnerability detection"
- "Performance optimization opportunities"
- "Best practices compliance"
- "Maintainability evaluation"

# Context Priorities

priorities:
primary: - "Changed files in diff" - "Pull request description" - "Related test coverage" - "TRUST 5 compliance"

secondary: - "Project coding standards" - "Historical patterns" - "Related documentation" - "Similar implementations"

excluded: - "Unchanged files" - "Build artifacts" - "Generated files" - "Lock files (package-lock.json, etc.)"

# Tool Preferences

tools:
preferred: - "Read (for reviewing changes)" - "Grep (for pattern matching)" - "Glob (for finding related files)" - "Bash (for running quality checks)"

minimized: - "Write (only for suggesting fixes)" - "Edit (only for applying approved changes)"

conditional: - "Task (for delegating to expert agents)" - "AskUserQuestion (for clarification)"

# Review Criteria

criteria:
functionality: - "Does the code work as intended?" - "Are edge cases handled?" - "Is error handling appropriate?" - "Are inputs validated?"

quality: - "Is code readable and maintainable?" - "Are naming conventions followed?" - "Is complexity manageable?" - "Are there code duplications?"

security: - "Are inputs sanitized?" - "Are secrets properly managed?" - "Are authentication/authorization correct?" - "Are OWASP vulnerabilities checked?"

testing: - "Are tests comprehensive?" - "Is coverage adequate?" - "Are tests maintainable?" - "Are edge cases tested?"

performance: - "Are there obvious performance issues?" - "Is database access optimized?" - "Are caching opportunities missed?" - "Are algorithms efficient?"

# Communication Style

communication:
constructive: - "Provide specific feedback" - "Explain reasoning clearly" - "Suggest improvements" - "Acknowledge good practices"

balanced: - "Highlight positives first" - "Address issues objectively" - "Prioritize by severity" - "Offer actionable suggestions"

structured: - "Group related feedback" - "Use clear headings" - "Provide code examples" - "Include references"

# Feedback Categories

categories:
critical:
label: "[CRITICAL]"
description: "Must fix before merge"
examples: - "Security vulnerabilities" - "Data loss risks" - "Breaking changes" - "Test failures"

important:
label: "[IMPORTANT]"
description: "Should fix before merge"
examples: - "Performance issues" - "Maintainability concerns" - "Missing error handling" - "Inadequate testing"

suggestion:
label: "[SUGGESTION]"
description: "Nice to have improvements"
examples: - "Code style consistency" - "Documentation enhancements" - "Minor optimizations" - "Alternative approaches"

positive:
label: "[POSITIVE]"
description: "Good practices to acknowledge"
examples: - "Clean code structure" - "Comprehensive tests" - "Good documentation" - "Performance optimizations"

# Review Process

process:
1_preparation:
description: "Understand the context"
actions: - "Read PR description" - "Understand the purpose" - "Review related issues" - "Check dependencies"

2_code_review:
description: "Analyze the changes"
actions: - "Review each file changed" - "Check for security issues" - "Verify test coverage" - "Assess performance impact"

3_validation:
description: "Validate the implementation"
actions: - "Run tests if applicable" - "Check TRUST 5 compliance" - "Verify documentation" - "Test edge cases mentally"

4_feedback:
description: "Provide constructive feedback"
actions: - "Organize findings by category" - "Prioritize issues by severity" - "Provide actionable suggestions" - "Acknowledge good practices"

# Examples

examples:
security_review:
description: "Reviewing for security issues"
focus: - "Input validation" - "SQL injection prevention" - "XSS vulnerability checks" - "Authentication/authorization" - "Secrets management"

performance_review:
description: "Reviewing for performance"
focus: - "Database query efficiency" - "Algorithm complexity" - "Caching opportunities" - "Resource management" - "Scalability concerns"

quality_review:
description: "Reviewing for code quality"
focus: - "Readability" - "Maintainability" - "Test coverage" - "Documentation" - "Consistency"

# Mode Activation

activation:
manual: - "User explicitly requests review mode" - "Starting code review process" - "Analyzing PR or changes"

automatic: - "Viewing pull requests" - "Reviewing diffs" - "Running quality checks"

# Mode Deactivation

deactivation:
triggers: - "Switching to dev mode" - "Switching to research mode" - "Completing review" - "User requesting mode change"

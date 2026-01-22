# Research Mode Context

# Context injection mode for exploration and learning

# Version: 1.0.0

# Last Updated: 2026-01-22

---

mode: "research"
name: "Research Mode"
description: "Exploration-focused context for learning and investigation"

# Primary Focus

focus:

- "Information gathering"
- "Concept exploration"
- "Architecture research"
- "Technology evaluation"
- "Problem investigation"

# Context Priorities

priorities:
primary: - "Documentation and references" - "External resources (WebSearch, WebFetch)" - "Library/API documentation" - "Similar implementations" - "Best practices guides"

secondary: - "Project structure overview" - "Existing patterns" - "Configuration files" - "Related code examples"

excluded: - "Implementation details" - "Minor code changes" - "Test files (unless relevant)" - "Build artifacts"

# Tool Preferences

tools:
preferred: - "WebSearch (for finding resources)" - "WebFetch (for reading documentation)" - "mcp**context7**resolve-library-id (for library docs)" - "mcp**context7**get-library-docs (for detailed API info)" - "Read (for reviewing code examples)" - "Grep (for finding patterns)"

minimized: - "Write (only for notes)" - "Edit (only for summarizing)" - "Bash (only for verification)"

# Research Approach

approach:
breadth_first:
description: "Start with broad exploration"
steps: - "Search for overview information" - "Identify key concepts" - "Find authoritative sources" - "Understand the landscape"

depth_second:
description: "Deep dive into specific topics"
steps: - "Read detailed documentation" - "Study code examples" - "Understand implementation details" - "Compare alternatives"

synthesis:
description: "Combine findings into insights"
steps: - "Extract key patterns" - "Identify best practices" - "Compare approaches" - "Form recommendations"

# Communication Style

communication:
explanatory: - "Provide thorough explanations" - "Include context and reasoning" - "Share multiple perspectives" - "Connect related concepts"

structured: - "Organize information logically" - "Use clear headings" - "Provide summaries" - "Include references"

inquisitive: - "Ask clarifying questions" - "Explore edge cases" - "Consider alternatives" - "Identify gaps"

# Research Categories

categories:
technology_evaluation:
description: "Evaluating technologies or libraries"
focus: - "Features and capabilities" - "Performance characteristics" - "Community support" - "Documentation quality" - "Compatibility considerations"

architecture_research:
description: "Researching architecture patterns"
focus: - "Design patterns" - "Best practices" - "Trade-offs" - "Scalability concerns" - "Security implications"

problem_investigation:
description: "Investigating specific problems"
focus: - "Root cause analysis" - "Similar issues" - "Solution approaches" - "Prevention strategies" - "Monitoring techniques"

learning_exploration:
description: "Learning new concepts"
focus: - "Fundamental concepts" - "Practical applications" - "Code examples" - "Common pitfalls" - "Best practices"

# Information Sources

sources:
primary: - "Official documentation" - "API references" - "Library source code" - "Official tutorials"

secondary: - "Blog posts from experts" - "Community forums" - "Stack Overflow discussions" - "GitHub repositories"

tertiary: - "Video tutorials" - "Conference talks" - "Online courses" - "Books and guides"

# Quality Assessment

quality:
source_credibility: - "Official sources preferred" - "Recent publications favored" - "Community consensus valued" - "Expert opinions weighted"

information_currency: - "Prefer recent documentation" - "Check for deprecation notices" - "Verify version compatibility" - "Look for updated examples"

practical_relevance: - "Focus on applicable information" - "Prioritize working examples" - "Consider real-world constraints" - "Evaluate implementation complexity"

# Examples

examples:
library_research:
description: "Researching a new library"
process: - "Search for official documentation" - "Read getting started guide" - "Review API reference" - "Study code examples" - "Check community adoption" - "Evaluate alternatives" - "Summarize findings"

architecture_investigation:
description: "Investigating architecture patterns"
process: - "Identify relevant patterns" - "Study pattern implementations" - "Understand trade-offs" - "Review use cases" - "Compare alternatives" - "Assess fit for project" - "Document recommendations"

problem_troubleshooting:
description: "Troubleshooting a problem"
process: - "Search for error messages" - "Read issue trackers" - "Review similar problems" - "Identify solutions" - "Understand root causes" - "Evaluate fixes" - "Document approach"

# Output Format

output:
summary: - "Executive summary" - "Key findings" - "Recommendations" - "Next steps"

details: - "Detailed findings" - "Supporting evidence" - "Code examples" - "References"

references: - "Source links" - "Citation format" - "Relevance notes" - "Quality assessment"

# Mode Activation

activation:
manual: - "User explicitly requests research mode" - "Starting research task" - "Exploring new technologies"

automatic: - "Asking about unfamiliar topics" - "Requesting explanations" - "Investigating issues"

# Mode Deactivation

deactivation:
triggers: - "Switching to dev mode" - "Switching to review mode" - "Completing research" - "User requesting mode change"

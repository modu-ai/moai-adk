# Template Details Module

Complete specifications for all three Yoda System templates.

## education.md - Theory-Focused Learning Guide

**File**: `templates/education.md`
**Lines**: 464
**Code Blocks**: 13
**Sections**: 8 major (Overview, Core Concepts, Practice Problems, References, Summary)

### When to Use

- Creating online courses or e-learning content
- Writing technical tutorials or textbooks
- Developing comprehensive learning paths
- Building self-paced training materials
- Creating detailed reference guides

### Structure Overview

```
1. Overview (Learning objectives, prerequisites, duration)
2. Core Concepts (3 sections: Basic → Practical → Advanced)
3. Practice Problems (2 basic + 2 advanced with solutions)
4. Reference Materials (Links, resources, community info)
5. Summary (Learning path, next steps)
```

### Key Features

- ✅ **Progressive learning path**: Structured progression from basic to advanced
- ✅ **Built-in assessment**: Understanding confirmation questions at each stage
- ✅ **Detailed code examples**: Syntax-correct, well-commented implementations
- ✅ **Context7 integration points**: Automatic documentation injection placeholders
- ✅ **Security & performance**: Explicit coverage of both non-functional requirements
- ✅ **Multiple difficulty levels**: Content adapts from Basic to Advanced

### Output Formats

- **Primary**: Markdown (.md)
- **Generated**: PDF (via pandoc or wkhtmltopdf)
- **Generated**: DOCX (via pandoc)

### Best For

- Self-paced learning environments
- Online education platforms
- Formal training programs
- Technical documentation
- Reference materials

### Customization Variables

```
Core variables (required):
- {{TOPIC}}                 - Main subject
- {{STUDY_TIME}}           - Duration estimate
- {{DIFFICULTY_LEVEL}}     - Basic/Intermediate/Advanced

Content variables:
- {{CLASS_NAME}}           - Python class names
- {{PRACTICAL_CLASS}}      - Real-world implementation
- {{ADVANCED_CLASS}}       - High-performance version
- {{PROBLEM_DESCRIPTIONS}} - Practice problem details

Integration variables:
- {{LIBRARY_NAME}}         - For Context7 MCP injection
- {{REFERENCE_URLS}}       - External documentation links
```

---

## presentation.md - Visual Presentation Slides

**File**: `templates/presentation.md`
**Lines**: 762
**Code Blocks**: 10
**Slides**: 18 (complete presentation)
**Layouts**: 7 types (title, content, case_study, content_with_code, closing)

### When to Use

- Creating conference presentations
- Preparing team talks or seminars
- Designing webinar slide decks
- Building sales or pitch materials
- Creating visual overviews for decision makers

### Structure Overview

```
Slide 1:   Title Slide (topic, instructor, audience, time)
Slide 2:   Presentation Overview (journey outline)
Slide 3:   Learning Objectives (4 goals, expectations)
Slide 4:   Background & Context (trends, problems, solutions)
Slides 5-7: Core Concepts (3 major topics with details)
Slide 8:   Practical Patterns (3 implementation patterns)
Slides 9-10: Case Studies (2 detailed real-world examples)
Slide 11:  Success/Failure Patterns (best practices & pitfalls)
Slide 12:  Advanced Topics (optimization strategies)
Slide 13:  Trends & Future (current state, forecasts)
Slide 14:  Action Plan (3-phase implementation roadmap)
Slide 15:  FAQ (5 common questions with answers)
Slide 16:  Q&A Session (facilitation guide)
Slide 17:  Key Takeaways (3 main messages + next steps)
Slide 18:  Closing (thanks, resources, contact info)
```

### Key Features

- ✅ **Complete 18-slide structure**: Ready to present as-is
- ✅ **Multiple slide layouts**: Title, content, code, case study, closing
- ✅ **Speaker notes**: Detailed guidance for each slide
- ✅ **Case studies included**: 2 real-world examples with metrics
- ✅ **Action-oriented**: Practical implementation roadmap
- ✅ **FAQ integration**: Common questions pre-addressed

### Output Formats

- **Primary**: Markdown (.md) with YAML frontmatter
- **Generated**: PPTX (via pandoc or Marp)
- **Generated**: PDF (via pandoc or wkhtmltopdf)

### Best For

- Public speaking and presentations
- Team knowledge sharing
- Executive summaries
- Conference talks
- Visual learning audiences

### Customization Variables

```
Presentation metadata:
- {{TOPIC}}                    - Main subject
- {{INSTRUCTOR}}               - Speaker name
- {{AUDIENCE}}                 - Target listeners
- {{DIFFICULTY_LEVEL}}         - Basic/Intermediate/Advanced
- {{PRESENTATION_TIME}}        - Total duration (e.g., "60 minutes")

Content variables:
- {{CORE_MESSAGE}}             - Key takeaway
- {{TREND_1/2/3}}              - Market/industry trends
- {{CONCEPT_1/2/3_TITLE}}      - Main concepts
- {{CASE_STUDY_1/2_TITLE}}     - Real-world examples
- {{SUCCESS/FAILURE_PATTERNS}}  - Best practices

Metrics and data:
- {{EFFICIENCY_IMPROVEMENT}}   - Performance gains
- {{COST_SAVING}}              - Financial benefits
- {{METRIC_IMPROVEMENTS}}      - Quantified results

Integration:
- {{LIBRARY_NAME}}             - For Context7 MCP injection
```

---

## workshop.md - Hands-On Workshop Materials

**File**: `templates/workshop.md`
**Lines**: 928
**Code Blocks**: 53
**Exercises**: 2 main (30-45 min each)
**Team Project**: 1 (4-6 hours, 4-stage evaluation)

### When to Use

- Conducting hands-on workshops
- Running intensive bootcamps
- Teaching practical skills
- Team training programs
- Building real-world competency

### Structure Overview

```
1. Workshop Goals (5 objectives, skills, understanding)
2. Prerequisites & Requirements (OS, software, knowledge)
3. Environment Setup (3 steps, 30 minutes total)
4. Hands-On Lab 1 (3 steps, 45 minutes)
   - Step 1: Basic implementation
   - Step 2: Enhancement
   - Step 3: Testing/validation
5. Hands-On Lab 2 (3-4 steps, 45 minutes, more advanced)
6. Team Project (4-6 hours)
   - MVP requirements (3+ features)
   - Team roles (leader, developers, QA)
   - Evaluation criteria (40+ points)
   - Implementation timeline
7. Troubleshooting (5+ common issues)
8. Additional Resources (advanced topics, community)
9. Completion Checklist (workshop validation)
```

### Key Features

- ✅ **Complete environment setup**: Step-by-step with validation
- ✅ **Progressive hands-on exercises**: Basic → Advanced
- ✅ **Team-based capstone project**: Real-world scenario
- ✅ **Comprehensive troubleshooting**: 5+ scenarios with solutions
- ✅ **Validation checkpoints**: At each major step
- ✅ **Real-world context**: Practical, production-relevant examples

### Output Formats

- **Primary**: Markdown (.md)
- **Generated**: DOCX (via pandoc)
- **Generated**: PDF (via pandoc or wkhtmltopdf)

### Best For

- Practical skill development
- Hands-on team training
- Intensive bootcamp programs
- Building real competency
- Project-based learning

### Customization Variables

```
Workshop setup:
- {{TOPIC}}                    - Main subject
- {{INSTRUCTOR}}               - Facilitator
- {{AUDIENCE}}                 - Target group
- {{TOTAL_DURATION}}           - Total hours (e.g., "6 hours")
- {{DIFFICULTY_LEVEL}}         - Basic/Intermediate/Advanced

Requirements:
- {{SUPPORTED_OS}}             - Operating systems
- {{CPU_REQUIREMENT}}          - Processor specs
- {{RAM_REQUIREMENT}}          - Memory needs
- {{SOFTWARE_1/2/3}}           - Required tools
- {{PREREQUISITE_1/2/3}}       - Background knowledge

Lab exercises:
- {{LAB_1/2_TITLE}}            - Exercise names
- {{LAB_1/2_GOAL_1/2/3}}       - Learning objectives
- {{LAB_1/2_STEP_1/2/3_*}}     - Step descriptions and code
- {{LAB_1/2_TIME}}             - Duration per lab
- {{LAB_1/2_DIFFICULTY}}       - Difficulty level

Team project:
- {{PROJECT_NAME}}             - Project title
- {{PROJECT_REQUIREMENT_1/2/3}}- MVP features
- {{TEAM_SIZE}}                - Team composition
- {{PROJECT_DURATION}}         - Project timeline
- {{EVALUATION_CRITERIA}}      - Grading rubric

Troubleshooting:
- {{TROUBLESHOOT_1-5_TITLE}}   - Problem descriptions
- {{TROUBLESHOOT_1-5_ERROR}}   - Error messages
- {{TROUBLESHOOT_1-5_SOLUTION}}- Solutions
```

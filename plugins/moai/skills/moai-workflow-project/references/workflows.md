# Core Workflows — Step-by-Step

Three workflows: project initialization, documentation generation, template optimization.

## Complete Project Initialization Workflow

Step 1: Initialize the project management system by specifying the project directory path.

Step 2: Execute complete setup with the following configuration parameters:

- Language setting: Specify primary language code (e.g., "en", "ko")
- User name: Developer or team name for personalization
- Domains: Project domains list (backend, frontend, mobile)
- Project type: Classification (e.g., web_application)
- Optimization enabled: Set true to enable template optimization during init

Interview questions: [schemas/tab_schema.json](../schemas/tab_schema.json) defines the
tab/batch/question structure the initialization and settings interviews walk through — read it
and follow its tabs and `field` keys when collecting configuration, so the questions asked and the
config values they write stay consistent with `schemas/config-schema.json` and the shipped
`git-strategy.yaml` template. When the schema and the live config templates disagree, the config
templates are authoritative — fix the schema instead of asking a dead key.

Step 3: Review initialization results:

- Language configuration with token cost analysis
- Documentation structure creation status
- Template analysis and optimization report
- Multilingual documentation setup confirmation

## Documentation Generation from SPEC Workflow

Step 1: Prepare SPEC data structure:

| Field | Content |
|-------|---------|
| Identifier | Unique SPEC ID (e.g., SPEC-001) |
| Title | Feature or component title |
| Description | Brief description of implementation |
| Requirements | List of specific requirements |
| Status | Planned / In Progress / Complete |
| Priority | High / Medium / Low |
| API Endpoints | List of endpoint definitions (path, method, description) |

Step 2: Generate comprehensive documentation from the SPEC data.

Step 3: Review generated documentation:

- Feature documentation with requirements
- API documentation with endpoint details
- Updated project documentation files
- Multilingual versions if configured

## Template Performance Optimization Workflow

Step 1: Analyze current templates to gather metrics.

Step 2: Configure optimization options:

- Backup first: true to create backup before optimization
- Apply size optimizations: reduce file sizes
- Apply performance optimizations: improve loading times
- Apply complexity optimizations: simplify template structures
- Preserve functionality: ensure all features remain intact

Step 3: Execute optimization and review results:

- Size reduction percentage achieved
- Performance improvement metrics
- Backup creation confirmation
- Detailed optimization report

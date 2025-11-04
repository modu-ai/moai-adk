# Examples

## Reporting Style Examples

### User-Facing Output (Screen)
```
✅ Phase 1: Requirements Analysis — Completed
✅ Phase 2: Architecture Design — In Progress  
⏳ Phase 3: Implementation — Pending

Key Decisions:
- Selected microservices architecture for scalability
- Using PostgreSQL for primary data store
- Implementing JWT authentication

Next Steps: Complete API design then begin implementation
```

### Internal Documentation (Files)
```
# Implementation Analysis

## Architecture Decisions
- Rationale: Microservices chosen for team scalability
- Trade-offs: Increased complexity vs. independent deployment
- Alternatives considered: Monolith, modular monolith

## Implementation Progress
- Completed: User service schema design
- In Progress: API contract definition
- Blocked: Authentication service dependencies

## Risk Assessment
- Technical: Service communication complexity
- Timeline: 2 weeks buffer needed for integration testing
```

## Output Location Examples

### Correct Usage
```python
# Screen output (user-facing)
print("✅ Task completed successfully")

# Internal documentation (files)
with open(".moai/reports/analysis-report.md", "w") as f:
    f.write("# Analysis Results\n\n...")
```

### Incorrect Usage
```python
# WRONG - Creating report in project root
with open("ANALYSIS_REPORT.md", "w") as f:
    f.write("This violates reporting rules...")

# CORRECT - Using proper location
with open(".moai/docs/analysis-report.md", "w") as f:
    f.write("This follows the rules...")
```

# moai_adk.core.quality.quality_gates

Quality gates automation for MoAI-ADK.

Integrates with linting tools (black, mypy, flake8) and provides
automated quality gate enforcement for continuous integration.

@FEATURE:QUALITY-GATES Quality gate automation and enforcement

## Functions

### __init__

Initialize quality gates - RED phase implementation.

```python
__init__(self, project_path)
```

## Classes

### QualityGateError

Quality gate exception.

### QualityGates

Manages automated quality gate enforcement.

@DESIGN:QUALITY-GATES-ARCH-001 Quality gates architecture

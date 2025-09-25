# moai_adk.resources.templates.scripts.validate_stage

MoAI-ADK Stage Validator
4단계 파이프라인의 각 Gate 검수 자동화

## Functions

### print_validation_results

검증 결과 출력

```python
print_validation_results(results, verbose)
```

### main

```python
main()
```

### __init__

```python
__init__(self, project_root)
```

### validate_specify_stage

SPECIFY Gate 검수

```python
validate_specify_stage(self)
```

### validate_plan_stage

PLAN Gate 검수

```python
validate_plan_stage(self)
```

### validate_tasks_stage

TASKS Gate 검수

```python
validate_tasks_stage(self)
```

### validate_implement_stage

IMPLEMENT Gate 검수

```python
validate_implement_stage(self)
```

### validate_all_stages

모든 단계 검수

```python
validate_all_stages(self)
```

## Classes

### StageValidator

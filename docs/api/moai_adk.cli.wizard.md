# moai_adk.cli.wizard

@FEATURE:WIZARD-001 Interactive setup wizard for MoAI-ADK projects.

@TASK:WIZARD-UI-001 Handles user input collection through a step-by-step wizard interface.

## Functions

### __init__

```python
__init__(self)
```

### run_wizard

Run the complete 10-step interactive wizard.

```python
run_wizard(self)
```

### _collect_product_vision

1-3단계: 제품 비전 수집

```python
_collect_product_vision(self)
```

### _collect_tech_stack

4-5단계: 기술 스택 수집

```python
_collect_tech_stack(self)
```

### _collect_quality_standards

6-7단계: 품질 기준 수집

```python
_collect_quality_standards(self)
```

### _collect_advanced_settings

8-10단계: 고급 설정 수집

```python
_collect_advanced_settings(self)
```

### _show_summary

설정 요약 출력

```python
_show_summary(self)
```

## Classes

### InteractiveWizard

@TASK:WIZARD-MAIN-001 Interactive setup wizard for MoAI-ADK projects.

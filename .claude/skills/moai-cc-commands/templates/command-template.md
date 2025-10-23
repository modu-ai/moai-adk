---
name: {{command_name}}
description: {{command_description}}
argument-hint: "{{command_args}}"
tools: {{command_tools}}
model: {{command_model}}
language: {{conversation_language_name}}
project: {{project_name}}
---

# {{command_title}}

{{command_description}}

## Usage

`/{{command_name}} {{command_args}}` — {{command_usage_example}}

## Agent Orchestration

1. **Phase 1**: Call {{primary_agent}} for {{phase1_task}}
2. **Phase 2**: Call {{secondary_agent}} for {{phase2_task}}
3. **Phase 3**: Call {{tertiary_agent}} for {{phase3_task}}

## Success Criteria

- {{success_criterion_1}}
- {{success_criterion_2}}
- {{success_criterion_3}}

## Language

> All output in {{conversation_language_name}}

# Malformed YAML Fixture

YAML 파싱 오류 케이스 픽스처. 로더가 파싱 오류를 반환해야 한다.

## Entries

```yaml
- id: CONST-V3R2-001
  zone: Frozen
  file: .claude/rules/moai/workflow/spec-workflow.md
  anchor: "#phase-overview"
  clause: "SPEC+EARS format"
  canary_gate: not-a-boolean
  invalid_field: [unclosed bracket
```

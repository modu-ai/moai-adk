# Duplicate IDs Fixture

중복 ID 오류 케이스 픽스처. 로더가 fatal error를 반환해야 한다.

## Entries

```yaml
- id: CONST-V3R2-042
  zone: Frozen
  file: .claude/rules/moai/core/moai-constitution.md
  anchor: "#quality-gates"
  clause: "TRUST 5"
  canary_gate: true

- id: CONST-V3R2-042
  zone: Evolvable
  file: .claude/rules/moai/core/agent-common-protocol.md
  anchor: "#time-estimation"
  clause: "Never use time predictions in plans or reports."
  canary_gate: false
```

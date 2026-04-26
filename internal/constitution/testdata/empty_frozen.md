# Empty Frozen Fixture

Frozen zone 엔트리가 없는 registry 픽스처.
LoadRegistry는 성공하지만 FilterByZone(ZoneFrozen)이 빈 slice를 반환해야 한다.
doctor check에서 warn-level 경고를 발행해야 한다.

## Entries

```yaml
- id: CONST-V3R2-001
  zone: Evolvable
  file: .claude/rules/moai/core/agent-common-protocol.md
  anchor: "#time-estimation"
  clause: "Never use time predictions in plans or reports."
  canary_gate: false

- id: CONST-V3R2-002
  zone: Evolvable
  file: .claude/rules/moai/core/agent-common-protocol.md
  anchor: "#output-format"
  clause: "User-Facing: Always use Markdown formatting."
  canary_gate: false
```

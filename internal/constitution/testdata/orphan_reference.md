# Orphan Reference Fixture

존재하지 않는 파일을 참조하는 orphan 케이스 픽스처.
로더가 panic 없이 Orphan=true로 마킹하고 나머지 엔트리를 로드해야 한다.

## Entries

```yaml
- id: CONST-V3R2-001
  zone: Frozen
  file: .claude/rules/moai/workflow/spec-workflow.md
  anchor: "#phase-overview"
  clause: "SPEC+EARS format"
  canary_gate: true

- id: CONST-V3R2-099
  zone: Frozen
  file: this-file-does-not-exist.md
  anchor: "#non-existent"
  clause: "Orphan clause for testing"
  canary_gate: true
```

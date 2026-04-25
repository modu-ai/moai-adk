# Overflow Mirror Fixture

design mirror 범위(051-099, 49 slots)를 초과하는 51개 엔트리를 포함하는 픽스처.
REQ-CON-001-021의 auto-extend(100-149) 로직 검증용.
loader가 overflow 감지 후 100-149 범위로 auto-extend하고 warning을 발행해야 한다.

## Entries

```yaml
- id: CONST-V3R2-001
  zone: Frozen
  file: .claude/rules/moai/workflow/spec-workflow.md
  anchor: "#phase-overview"
  clause: "SPEC+EARS format"
  canary_gate: true

- id: CONST-V3R2-051
  zone: Frozen
  file: .claude/rules/moai/design/constitution.md
  anchor: "#2-frozen-vs-evolvable-zones"
  clause: "Mirror entry 051"
  canary_gate: true

- id: CONST-V3R2-052
  zone: Frozen
  file: .claude/rules/moai/design/constitution.md
  anchor: "#2-frozen-vs-evolvable-zones"
  clause: "Mirror entry 052"
  canary_gate: true

- id: CONST-V3R2-053
  zone: Frozen
  file: .claude/rules/moai/design/constitution.md
  anchor: "#2-frozen-vs-evolvable-zones"
  clause: "Mirror entry 053"
  canary_gate: true

- id: CONST-V3R2-054
  zone: Frozen
  file: .claude/rules/moai/design/constitution.md
  anchor: "#2-frozen-vs-evolvable-zones"
  clause: "Mirror entry 054"
  canary_gate: true

- id: CONST-V3R2-055
  zone: Frozen
  file: .claude/rules/moai/design/constitution.md
  anchor: "#2-frozen-vs-evolvable-zones"
  clause: "Mirror entry 055"
  canary_gate: true

- id: CONST-V3R2-056
  zone: Frozen
  file: .claude/rules/moai/design/constitution.md
  anchor: "#2-frozen-vs-evolvable-zones"
  clause: "Mirror entry 056"
  canary_gate: true

- id: CONST-V3R2-057
  zone: Frozen
  file: .claude/rules/moai/design/constitution.md
  anchor: "#2-frozen-vs-evolvable-zones"
  clause: "Mirror entry 057"
  canary_gate: true

- id: CONST-V3R2-058
  zone: Frozen
  file: .claude/rules/moai/design/constitution.md
  anchor: "#2-frozen-vs-evolvable-zones"
  clause: "Mirror entry 058"
  canary_gate: true

- id: CONST-V3R2-059
  zone: Frozen
  file: .claude/rules/moai/design/constitution.md
  anchor: "#2-frozen-vs-evolvable-zones"
  clause: "Mirror entry 059"
  canary_gate: true

- id: CONST-V3R2-060
  zone: Frozen
  file: .claude/rules/moai/design/constitution.md
  anchor: "#2-frozen-vs-evolvable-zones"
  clause: "Mirror entry 060"
  canary_gate: true

- id: CONST-V3R2-061
  zone: Frozen
  file: .claude/rules/moai/design/constitution.md
  anchor: "#2-frozen-vs-evolvable-zones"
  clause: "Mirror entry 061"
  canary_gate: true

- id: CONST-V3R2-062
  zone: Frozen
  file: .claude/rules/moai/design/constitution.md
  anchor: "#2-frozen-vs-evolvable-zones"
  clause: "Mirror entry 062"
  canary_gate: true

- id: CONST-V3R2-063
  zone: Frozen
  file: .claude/rules/moai/design/constitution.md
  anchor: "#2-frozen-vs-evolvable-zones"
  clause: "Mirror entry 063"
  canary_gate: true

- id: CONST-V3R2-064
  zone: Frozen
  file: .claude/rules/moai/design/constitution.md
  anchor: "#2-frozen-vs-evolvable-zones"
  clause: "Mirror entry 064"
  canary_gate: true

- id: CONST-V3R2-065
  zone: Frozen
  file: .claude/rules/moai/design/constitution.md
  anchor: "#2-frozen-vs-evolvable-zones"
  clause: "Mirror entry 065"
  canary_gate: true

- id: CONST-V3R2-066
  zone: Frozen
  file: .claude/rules/moai/design/constitution.md
  anchor: "#2-frozen-vs-evolvable-zones"
  clause: "Mirror entry 066"
  canary_gate: true

- id: CONST-V3R2-067
  zone: Frozen
  file: .claude/rules/moai/design/constitution.md
  anchor: "#2-frozen-vs-evolvable-zones"
  clause: "Mirror entry 067"
  canary_gate: true

- id: CONST-V3R2-068
  zone: Frozen
  file: .claude/rules/moai/design/constitution.md
  anchor: "#2-frozen-vs-evolvable-zones"
  clause: "Mirror entry 068"
  canary_gate: true

- id: CONST-V3R2-069
  zone: Frozen
  file: .claude/rules/moai/design/constitution.md
  anchor: "#2-frozen-vs-evolvable-zones"
  clause: "Mirror entry 069"
  canary_gate: true

- id: CONST-V3R2-070
  zone: Frozen
  file: .claude/rules/moai/design/constitution.md
  anchor: "#2-frozen-vs-evolvable-zones"
  clause: "Mirror entry 070"
  canary_gate: true

- id: CONST-V3R2-071
  zone: Frozen
  file: .claude/rules/moai/design/constitution.md
  anchor: "#2-frozen-vs-evolvable-zones"
  clause: "Mirror entry 071"
  canary_gate: true

- id: CONST-V3R2-072
  zone: Frozen
  file: .claude/rules/moai/design/constitution.md
  anchor: "#2-frozen-vs-evolvable-zones"
  clause: "Mirror entry 072"
  canary_gate: true

- id: CONST-V3R2-073
  zone: Frozen
  file: .claude/rules/moai/design/constitution.md
  anchor: "#2-frozen-vs-evolvable-zones"
  clause: "Mirror entry 073"
  canary_gate: true

- id: CONST-V3R2-074
  zone: Frozen
  file: .claude/rules/moai/design/constitution.md
  anchor: "#2-frozen-vs-evolvable-zones"
  clause: "Mirror entry 074"
  canary_gate: true

- id: CONST-V3R2-075
  zone: Frozen
  file: .claude/rules/moai/design/constitution.md
  anchor: "#2-frozen-vs-evolvable-zones"
  clause: "Mirror entry 075"
  canary_gate: true

- id: CONST-V3R2-076
  zone: Frozen
  file: .claude/rules/moai/design/constitution.md
  anchor: "#2-frozen-vs-evolvable-zones"
  clause: "Mirror entry 076"
  canary_gate: true

- id: CONST-V3R2-077
  zone: Frozen
  file: .claude/rules/moai/design/constitution.md
  anchor: "#2-frozen-vs-evolvable-zones"
  clause: "Mirror entry 077"
  canary_gate: true

- id: CONST-V3R2-078
  zone: Frozen
  file: .claude/rules/moai/design/constitution.md
  anchor: "#2-frozen-vs-evolvable-zones"
  clause: "Mirror entry 078"
  canary_gate: true

- id: CONST-V3R2-079
  zone: Frozen
  file: .claude/rules/moai/design/constitution.md
  anchor: "#2-frozen-vs-evolvable-zones"
  clause: "Mirror entry 079"
  canary_gate: true

- id: CONST-V3R2-080
  zone: Frozen
  file: .claude/rules/moai/design/constitution.md
  anchor: "#2-frozen-vs-evolvable-zones"
  clause: "Mirror entry 080"
  canary_gate: true

- id: CONST-V3R2-081
  zone: Frozen
  file: .claude/rules/moai/design/constitution.md
  anchor: "#2-frozen-vs-evolvable-zones"
  clause: "Mirror entry 081"
  canary_gate: true

- id: CONST-V3R2-082
  zone: Frozen
  file: .claude/rules/moai/design/constitution.md
  anchor: "#2-frozen-vs-evolvable-zones"
  clause: "Mirror entry 082"
  canary_gate: true

- id: CONST-V3R2-083
  zone: Frozen
  file: .claude/rules/moai/design/constitution.md
  anchor: "#2-frozen-vs-evolvable-zones"
  clause: "Mirror entry 083"
  canary_gate: true

- id: CONST-V3R2-084
  zone: Frozen
  file: .claude/rules/moai/design/constitution.md
  anchor: "#2-frozen-vs-evolvable-zones"
  clause: "Mirror entry 084"
  canary_gate: true

- id: CONST-V3R2-085
  zone: Frozen
  file: .claude/rules/moai/design/constitution.md
  anchor: "#2-frozen-vs-evolvable-zones"
  clause: "Mirror entry 085"
  canary_gate: true

- id: CONST-V3R2-086
  zone: Frozen
  file: .claude/rules/moai/design/constitution.md
  anchor: "#2-frozen-vs-evolvable-zones"
  clause: "Mirror entry 086"
  canary_gate: true

- id: CONST-V3R2-087
  zone: Frozen
  file: .claude/rules/moai/design/constitution.md
  anchor: "#2-frozen-vs-evolvable-zones"
  clause: "Mirror entry 087"
  canary_gate: true

- id: CONST-V3R2-088
  zone: Frozen
  file: .claude/rules/moai/design/constitution.md
  anchor: "#2-frozen-vs-evolvable-zones"
  clause: "Mirror entry 088"
  canary_gate: true

- id: CONST-V3R2-089
  zone: Frozen
  file: .claude/rules/moai/design/constitution.md
  anchor: "#2-frozen-vs-evolvable-zones"
  clause: "Mirror entry 089"
  canary_gate: true

- id: CONST-V3R2-090
  zone: Frozen
  file: .claude/rules/moai/design/constitution.md
  anchor: "#2-frozen-vs-evolvable-zones"
  clause: "Mirror entry 090"
  canary_gate: true

- id: CONST-V3R2-091
  zone: Frozen
  file: .claude/rules/moai/design/constitution.md
  anchor: "#2-frozen-vs-evolvable-zones"
  clause: "Mirror entry 091"
  canary_gate: true

- id: CONST-V3R2-092
  zone: Frozen
  file: .claude/rules/moai/design/constitution.md
  anchor: "#2-frozen-vs-evolvable-zones"
  clause: "Mirror entry 092"
  canary_gate: true

- id: CONST-V3R2-093
  zone: Frozen
  file: .claude/rules/moai/design/constitution.md
  anchor: "#2-frozen-vs-evolvable-zones"
  clause: "Mirror entry 093"
  canary_gate: true

- id: CONST-V3R2-094
  zone: Frozen
  file: .claude/rules/moai/design/constitution.md
  anchor: "#2-frozen-vs-evolvable-zones"
  clause: "Mirror entry 094"
  canary_gate: true

- id: CONST-V3R2-095
  zone: Frozen
  file: .claude/rules/moai/design/constitution.md
  anchor: "#2-frozen-vs-evolvable-zones"
  clause: "Mirror entry 095"
  canary_gate: true

- id: CONST-V3R2-096
  zone: Frozen
  file: .claude/rules/moai/design/constitution.md
  anchor: "#2-frozen-vs-evolvable-zones"
  clause: "Mirror entry 096"
  canary_gate: true

- id: CONST-V3R2-097
  zone: Frozen
  file: .claude/rules/moai/design/constitution.md
  anchor: "#2-frozen-vs-evolvable-zones"
  clause: "Mirror entry 097"
  canary_gate: true

- id: CONST-V3R2-098
  zone: Frozen
  file: .claude/rules/moai/design/constitution.md
  anchor: "#2-frozen-vs-evolvable-zones"
  clause: "Mirror entry 098"
  canary_gate: true

- id: CONST-V3R2-099
  zone: Frozen
  file: .claude/rules/moai/design/constitution.md
  anchor: "#2-frozen-vs-evolvable-zones"
  clause: "Mirror entry 099"
  canary_gate: true

- id: CONST-V3R2-100
  zone: Frozen
  file: .claude/rules/moai/design/constitution.md
  anchor: "#2-frozen-vs-evolvable-zones"
  clause: "Mirror entry 100 - overflow auto-extended"
  canary_gate: true

- id: CONST-V3R2-101
  zone: Frozen
  file: .claude/rules/moai/design/constitution.md
  anchor: "#2-frozen-vs-evolvable-zones"
  clause: "Mirror entry 101 - overflow auto-extended"
  canary_gate: true
```

# Decisions — fixture bundle (#2619 AC-3)

A fenced example is never the table:

```markdown
| id | what | options | default | reversible | category | deadline | decider | answer | evidence |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| X | example | A: a; B: b | Z | maybe | none | soon | nobody |  | - |
```

| id | what | options | default | reversible | category | deadline | decider | answer | evidence |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| D1 | open-PR limit | A: 8; B: 4 | A | yes | none | 2026-11-13T00:00:00Z | leader | A | #1975 c6840449133 |
| D2 | default deadline for user decisions | A: 24 h; B: 72 h | A | yes | none | 2026-11-14T00:00:00Z | user |  | #2619 c6840519246 |
| D3 | publish the docs site | A: now; B: later | B | no | external_contract | 2026-11-14T00:00:00Z | user |  | #2298 |
| D4 | apply the tag ruleset | A: now; B: later | B | yes | guarded | 2026-11-14T00:00:00Z | user |  | #2197 |
| D5 | latency target | A: record only; B: 200 ms | A | yes | none | 2026-11-20T00:00:00Z | user |  | #2539 |

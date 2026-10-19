# t1865 — MCP reconnection check (board d-20261018T235443Z-39d8)

Requested by the factory leader: confirm the `moai` MCP server reconnection by one direct call from this session, not from the operator's report.

| Field | Value |
|---|---|
| Session | lane-31 (card t1865) |
| Tool | `mcp__moai__session_list` (argument `spec_id` = `SPEC-UPDATE-OWNERSHIP-REPAIR-001`) |
| Calls | exactly one |
| UTC time | 2026-10-18T23:55:08Z (clock read right after the call returned) |
| Result | success |
| Response | `{"count":0,"sessions":[],"spec_id":"SPEC-UPDATE-OWNERSHIP-REPAIR-001"}` |
| Server build named by the reconnect notice | `moai_cp/20261007_195927`, commit `7705b365e2`, built 2026-10-18T10:45:08Z, pid 60306 |

Reading: the call returned a well-formed answer, so the MCP server answers tool calls again in this session. `count: 0` is the registry's active-session count for this SPEC (the lane's own session is not registered under that SPEC filter), not an error.

Consequence for the card: the required `mcp__moai__audit_multi` call of the final sync audit is available again. No command, lease or input in flight was restarted; no duplicate audit was started. The final audit still waits for the chunk-13 verification, the SPEC wording review and the sync-document correction, and starts only when the factory admission conditions (load, free memory, landing slot) hold.

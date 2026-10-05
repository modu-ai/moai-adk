# Vendored codex app-server response schemas

These JSON Schema files are copies of the response schemas that `codex-cli 0.160.0`
generates for its app-server protocol. `TestManagedServerRequestPolicyMatchesCodexSchema`
(`internal/cli/managed_hardening_test.go`) validates every answer the managed Factory
owner writes to a server-originated request against them, so the answer shapes are
checked against the real protocol rather than against our own reading of enum names.

Generated with (codex-cli 0.160.0), then the files below copied unchanged out of `<dir>`:

```
codex app-server generate-json-schema --out <dir>
```

Files: `ApplyPatchApprovalResponse.json`, `ExecCommandApprovalResponse.json`,
`CommandExecutionRequestApprovalResponse.json`, `FileChangeRequestApprovalResponse.json`,
`PermissionsRequestApprovalResponse.json`, `McpServerElicitationRequestResponse.json`,
`DynamicToolCallResponse.json`, `JSONRPCError.json`.

Regenerate them, and re-run the test, whenever the minimum supported codex version moves.

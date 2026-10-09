# Vendored codex app-server response schemas

These JSON Schema files are copies of the response schemas that `codex-cli 0.161.0`
generates for its app-server protocol. `TestManagedServerRequestPolicyMatchesCodexSchema`
(`internal/cli/managed_hardening_test.go`) validates every answer the managed Factory
owner writes to a server-originated request against them, so the answer shapes are
checked against the real protocol rather than against our own reading of enum names.

Generated with (codex-cli 0.161.0), then the files below copied unchanged out of `<dir>`:

```
npx -y @openai/codex@0.161.0 app-server generate-json-schema --out <dir>
```

The generator emits 39 top-level files (request-side Params companions, JSON-RPC
envelope types, and two consolidated bundles) plus `v1/`/`v2/` split layouts —
only the eight consumed Response schemas are vendored here. Re-vendor the rest
per the discipline below if a consumer for them appears.

Files: `ApplyPatchApprovalResponse.json`, `ExecCommandApprovalResponse.json`,
`CommandExecutionRequestApprovalResponse.json`, `FileChangeRequestApprovalResponse.json`,
`PermissionsRequestApprovalResponse.json`, `McpServerElicitationRequestResponse.json`,
`DynamicToolCallResponse.json`, `JSONRPCError.json`.

`resume-help.txt` is the verbatim `codex resume --help` output of the same binary
(captured 2026-10-09). `TestManagedCodexRemoteSupportProbe`
(`internal/cli/managed_codex_tui_test.go`) reads it as the feature-detection
baseline for the managed TUI's `--remote` capability probe.

Regenerate them, and re-run the tests, whenever the minimum supported codex version moves.

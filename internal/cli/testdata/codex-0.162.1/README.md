# Vendored codex app-server response schemas

These JSON Schema files are copies of the response schemas that `codex-cli 0.162.1`
generates for its app-server protocol. `TestManagedServerRequestPolicyMatchesCodexSchema`
(`internal/cli/managed_hardening_test.go`) validates every answer the managed Factory
owner writes to a server-originated request against them, so the answer shapes are
checked against the real protocol rather than against our own reading of enum names.

Generated with (codex-cli 0.162.1), then the files below copied unchanged out of `<dir>`:

```
npx -y @openai/codex@0.162.1 app-server generate-json-schema --out <dir>
```

Refresh check (2026-10-15, card t1754): the eight files below were compared byte-for-byte
against a local `codex app-server generate-json-schema` run of codex-cli 0.162.1, and all
eight are identical to the 0.161.0 generation this directory carried before the refresh.
The generator emits 37 top-level files (request-side Params companions, JSON-RPC
envelope types, and two consolidated bundles) plus `v1/`/`v2/` split layouts —
only the eight consumed Response schemas are vendored here. Re-vendor the rest
per the discipline below if a consumer for them appears.

Files: `ApplyPatchApprovalResponse.json`, `ExecCommandApprovalResponse.json`,
`CommandExecutionRequestApprovalResponse.json`, `FileChangeRequestApprovalResponse.json`,
`PermissionsRequestApprovalResponse.json`, `McpServerElicitationRequestResponse.json`,
`DynamicToolCallResponse.json`, `JSONRPCError.json`.

`resume-help.txt` is the verbatim `codex resume --help` output of codex-cli 0.162.1
(captured 2026-10-15). It is byte-identical to the 0.161.0 capture of 2026-10-09 that
this directory carried before the refresh. `TestManagedCodexRemoteSupportProbe`
(`internal/cli/managed_codex_tui_test.go`) reads it as the feature-detection
baseline for the managed TUI's `--remote` capability probe.

Compatibility check (2026-10-17, card t1754): the eight vendored schemas and `resume-help.txt` were
compared byte-for-byte against `codex app-server generate-json-schema` and `codex resume --help` of the
installed codex-cli 0.163.0, and all nine are identical. The files stay labelled 0.162.1 because that is
the version they were captured from.

Regenerate them, and re-run the tests, whenever the minimum supported codex version moves.

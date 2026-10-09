# Vendored codex login status sample

`login-status-not-logged-in.txt` is the verbatim stderr capture of the
**unauthenticated** `codex login status` output of codex-cli 0.161.0.
It is the minimum mandatory observation of REQ-CONF-007
(SPEC-CODEX-CONFORMANCE-001): the stage-2 grammar
(`codexAuthStatusLine`, `internal/cli/mcp_codex.go`) must classify it by
the same rule as the 0.160-era output.

Capture provenance (2026-10-09, fixture regeneration environment):

```
CODEX_HOME=<empty temp dir> npx -y @openai/codex@0.161.0 login status
```

- binary version: `codex-cli 0.161.0` (`--version`, exit 0)
- stderr: `Not logged in` + newline (the entire capture; the file's exact bytes)
- stdout: empty
- exit code: 1

Sanitization: the unauthenticated output carries no token and no account
identifier — the capture is the logged-OUT shape, so nothing needed
redaction. The keyring-logged-in variant could not be captured (no
credentials available in the generation environment) and is recorded as
the run phase's only eligible verification gap (progress.md §E.2).

Consumer: `TestClassifyCodexAuth_Codex0161CapturedOutputIsAGap`
(`internal/cli/codex_auth_ladder_test.go`), which asserts the file's
bytes stay exactly this capture and that both the pure parser and the
full two-stage ladder classify it `codexAuthUnknown` — a gap, never a
"not authenticated" provider verdict.

# Gateway product binding verification

## Claim

The CLI now wires `moai gpt` to a private gateway child. The parent creates an
exact four-model GPT catalog and a random session token, while the child creates
the verified OpenAI adapter and resolves credentials from the MoAI-owned
`gateway-auth` store at request time. The parent and child environments remove
credential-shaped inherited variables and routing overrides. Claude-only and
GLM launch registration remains on its existing path.

The binding also opens a private conversation family index below
`<MoAIHome>/state/gateway-conversations`. New launches receive a generated
`--session-id`; exact `--resume` and `--continue`/fork selection is resolved by
the family manager before the child launch plan is returned. The index stores
paths and UUID metadata only; it does not copy native transcripts, provider
payloads, or credentials.

## Evidence

The following focused test was run against this worktree:

```text
$ go test -race ./internal/cli -run 'TestGPTBinding|TestGatewayGPTModels|TestGatewayChildEnvironment|TestProductionGatewayFactory|TestGPTClosedVerbs' -count=1 -timeout=120s
ok  github.com/modu-ai/moai-adk/internal/cli  2.577s
```

The focused static checks were run as one read-only batch:

```text
$ go vet ./internal/cli

```

The binding tests assert the four exact route IDs (`gpt-6-astra`,
`gpt-5.6-sol`, `gpt-5.6-terra`, and `gpt-5.6-luna`), OpenAI/PKCE identity,
272000 context capability, streaming/tools capability, and the removal of
`Z_AI_API_KEY`, `OPENAI_API_KEY`, `ANTHROPIC_API_KEY`, `CODEX_HOME`, and
`ANTHROPIC_BASE_URL` from the child environment.

`TestGPTBindingBootstrapsPrivateFamilyAndSessionArgs` creates a completed
family record in the owned state directory, then verifies that an exact
`--resume <UUID>` request reaches the child plan with that UUID. The companion
new-session assertion verifies `--session-id <UUID>` is generated before child
startup.

## Baseline-attribution

The evidence was measured on branch `WT-unified-gateway`, worktree
`/Users/goos/MoAI/moai-adk-go/.claude/worktrees/moai-proxy-unified`, during
this implementation turn. No commit or remote operation was performed.

## Gaps

The test process did not replace itself with a real Claude Code executable and
did not send a request to the live GPT subscription endpoint. The private child
factory requires an installed `codex` executable and a logged-in owned store at
runtime. Windows execution and GitHub CI results are not represented by this
macOS test output. The external provider acceptance verifier is limited to
owned OpenAI credential/provider and generation checks; no unverified endpoint
was invented.

## Residual-risk

An expired credential can refresh only when the existing broker and the
provider acceptance verifier both succeed; the current verifier deliberately
fails closed for foreign or inaccessible references. The full product path
still requires a real Claude Code process/provider run and the planned GitHub
Windows CI run before operational acceptance.

# Retained conversation family verification

## Claim

The new `internal/gateway/conversation` package provides a bounded, private
family index with pre-launch UUID descriptors for new, exact resume, continue,
selection, and fork. It validates completion evidence before resume, keeps fork
receipt candidates in a separate child store, and provides a process-owned
family lease. No launcher, provider, credential, transcript-copy, or CLI code
was changed.

## Evidence

The focused specification tests were run after implementation:

```text
$ go test ./internal/gateway/conversation -count=1
ok  github.com/modu-ai/moai-adk/internal/gateway/conversation  0.430s
```

The same tests were run with the race detector:

```text
$ go test -race ./internal/gateway/conversation -count=1
ok  github.com/modu-ai/moai-adk/internal/gateway/conversation  1.497s
```

Static checks and a Windows cross compilation were run:

```text
$ go vet ./internal/gateway/conversation

$ GOOS=windows GOARCH=amd64 go test ./internal/gateway/conversation -c -o /tmp/conversation-windows.test.exe
```

The focused tests cover incomplete and completed transcript admission, exact
resume arguments, project-scoped continuation, fork UUID/argument formation,
fork family lease exclusivity, and secure namespace decision cases. The
implementation stores only family/UUID/project/CWD/config and receipt paths,
transcript relative identity, and completion sequence in its index. Index writes
use a private temporary file, `Sync`, and same-directory rename; entries and
transcript input are size bounded.

## Baseline-attribution

Measured against the current isolated worktree:

```text
worktree: /Users/goos/MoAI/moai-adk-go/.claude/worktrees/moai-proxy-unified
branch: WT-unified-gateway
HEAD: 81c1d58f9
family.go: d7491223100cbaba8b2311cdb298b5eb0cc88a2ce90173375db6adc73e6a0116
family_test.go: 61514bd2879aa284e91b41abafd46f8e2624e010350e22af13bce737d7d1045e
```

The two owned paths were untracked before this verification. Existing receipt
and opaque packages were read and reused; no receipt implementation was
modified.

## Gaps

Actual Claude Code launcher activation, provider egress, native transcript
schema produced by a live client, cross-process lease recovery, Windows native
execution, GitHub CI evidence, and production fork/resume integration were not
observed by this package-level verification. The package does not yet connect
to the launcher or the opaque codec. The secure namespace function records the
approved selection matrix but does not read a keychain or credential.

## Residual-risk

The family lease is process-owned in this package and therefore does not yet
coordinate independent OS processes; launcher integration must add a durable
OS-level lease or use the existing private locking seam before claiming
cross-process exclusivity. Transcript validation accepts the explicit
preflight record contract used by these tests and must be checked against a
live Claude client capture before egress is enabled. Cross compilation proves
build compatibility only, not Windows ACL, identity, crash recovery, or
write-through behavior.

# SPEC-V3R2-RT-003 Acceptance Criteria — Given/When/Then

> Detailed acceptance scenarios for each AC declared in `spec.md` §6 + plan §1.4 traceability matrix.
> Companion to `spec.md` v0.1.0, `research.md` v0.1.0, `plan.md` v0.1.0, `tasks.md` v0.1.0.

## HISTORY

| Version | Date       | Author                        | Description                                                            |
|---------|------------|-------------------------------|------------------------------------------------------------------------|
| 0.1.0   | 2026-05-10 | manager-spec (Plan workflow)  | Initial G/W/T conversion of 16 ACs (AC-V3R2-RT-003-01 through -16) + 3 perf budget ACs (AC-17/18/19) + 1 LSP carve-out baseline AC (AC-21). Total 20 ACs. Each maps to T-RT003-NN tasks per plan §1.4 traceability matrix. |

---

## Scope

본 문서는 `spec.md` §6 의 16 ACs 와 plan §1.4 traceability matrix 의 추가 ACs 를 Given/When/Then 형식으로 변환한다. Test mapping 은 `internal/sandbox/*_test.go` 또는 manual verification 으로 표시.

**Notation**:
- **Test mapping**: Go test function 또는 manual verification step.
- **Sentinel**: 음의 path test 가 기대하는 literal error string / sentinel error.
- **OS**: Linux / macOS / All / CI (per-OS gating).
- **Maps to REQ**: spec.md §5 의 EARS REQ ID.

---

## AC-V3R2-RT-003-01 — macOS sandbox-exec denies write outside scope

Maps to: REQ-V3R2-RT-003-010, REQ-V3R2-RT-003-013.
OS: macOS.

### Happy path

- **Given** an agent with `sandbox: seatbelt` running on macOS, with `SandboxOptions.WritableScope = ["/tmp/agent-worktree"]`
- **When** the agent invokes `Bash(touch /etc/passwd)` via `Launcher.Exec`
- **Then** the command exits non-zero (EPERM 1)
- **And** stdout contains a SystemMessage line naming the sandbox-denied path `/etc/passwd`
- **And** `/etc/passwd` is not modified on the host

### Edge case — write inside scope succeeds

- **Given** same configuration
- **When** the agent invokes `Bash(touch /tmp/agent-worktree/test.txt)`
- **Then** the command exits 0
- **And** the file exists at `/tmp/agent-worktree/test.txt`

### Test mapping

- `internal/sandbox/seatbelt_test.go::TestSeatbelt_FileWriteScopeEPERM` (macOS-gated)
- Sentinel: `*ExitError` with code 1 + SystemMessage matching `^sandbox-denied: /etc/passwd`

---

## AC-V3R2-RT-003-02 — Linux bubblewrap blocks network egress to non-allowlist

Maps to: REQ-V3R2-RT-003-011, REQ-V3R2-RT-003-014.
OS: Linux.

### Happy path

- **Given** an agent with `sandbox: bubblewrap` on Linux, default network allowlist (8 hosts)
- **When** the agent invokes `Bash(curl -sS -o /dev/null -w "%{http_code}" https://evil.example.com)`
- **Then** the command fails (curl exit code 6 = could not resolve host, or 7 = could not connect)
- **And** SystemMessage names the blocked host `evil.example.com`

### Edge case — allowlist host succeeds

- **Given** same configuration
- **When** the agent invokes `Bash(curl -sS -o /dev/null -w "%{http_code}" https://github.com)`
- **Then** the command exits 0 with HTTP 200

### Test mapping

- `internal/sandbox/bubblewrap_test.go::TestBubblewrap_NetworkBlocked` (Linux-gated; `MOAI_TEST_NETWORK=1` for the curl-based assertions; otherwise UNIX socket check)
- Sentinel: SystemMessage matching `^sandbox-network-denied: evil.example.com`

---

## AC-V3R2-RT-003-03 — Default sandbox per role profile is OS-resolved (not "none")

Maps to: REQ-V3R2-RT-003-003.
OS: All.

### Happy path

- **Given** `internal/template/templates/.moai/config/sections/workflow.yaml` with `role_profiles.implementer.sandbox: ""` (empty placeholder)
- **And** `defaults.go::NewDefaultConfig()` resolves implementer/tester/designer roles to OS-appropriate sandbox
- **When** an agent with `role_profile: implementer` is spawned on macOS
- **Then** `Launcher.resolveBackend(declared)` returns `seatbelt` (not `none`)
- **And** equivalent test on Linux returns `bubblewrap`

### Edge case — researcher role retains "none"

- **Given** same config
- **When** agent with `role_profile: researcher` spawns
- **Then** resolved sandbox is `none` (read-only role, no isolation needed)

### Test mapping

- `internal/sandbox/launcher_test.go::TestLauncher_ResolveBackend_AllScenarios` (4 cases: macOS+CI, macOS, Linux+CI, Linux × per-role 7 = 28 sub-cases; first AC checks 6 implementer/tester/designer cases)
- `internal/config/types_test.go::TestRoleProfile_Sandbox_DefaultByRole` (after T-RT003-29)

---

## AC-V3R2-RT-003-04 — `moai doctor sandbox` reports backend availability + per-agent resolved

Maps to: REQ-V3R2-RT-003-005, REQ-V3R2-RT-003-032, REQ-V3R2-RT-003-050.
OS: All.

### Happy path (macOS, all backends present)

- **Given** macOS host with `/usr/bin/sandbox-exec` present, `bwrap` not in PATH, `docker` daemon running
- **When** user runs `moai doctor sandbox`
- **Then** stdout includes (in any order):
  ```
  bubblewrap: unavailable (bwrap not in PATH)
  seatbelt: available (/usr/bin/sandbox-exec)
  docker: available (Docker version X.Y.Z)
  ```
- **And** for each agent in `.claude/agents/moai/*.md`, a line `<agent-name>: declared=<value>, effective=<value>` is printed

### Edge case — `--profile <agent>` flag

- **Given** same host
- **When** user runs `moai doctor sandbox --profile expert-backend`
- **Then** stdout prints the SBPL profile (macOS) or bwrap argument list (Linux) that would be applied to expert-backend

### Edge case — backend missing prompts AskUser via orchestrator (REQ-050)

- **Given** Linux host without `bwrap`
- **When** an implementer agent attempts to spawn (orchestrator-side workflow)
- **Then** `Launcher.Available()` returns false → orchestrator (per plan §3.3 + tasks T-RT003-50 doc-only) presents AskUserQuestion: install bwrap / opt into `sandbox: none`
- **Note**: the AskUser flow is owner = orchestrator (not subagent); RT-003 provides the detection + error; the prompt is in MIG-001 / orchestrator scope.

### Test mapping

- `internal/cli/doctor_sandbox_test.go::TestDoctorSandbox_AvailabilityReport` (mock `exec.LookPath`)
- `internal/cli/doctor_sandbox_test.go::TestDoctorSandbox_ProfileFlag`
- Manual verification step recorded in `progress.md` (T-RT003-47)

---

## AC-V3R2-RT-003-05 — Backend unavailable causes spawn failure (no silent unsandboxed execution)

Maps to: REQ-V3R2-RT-003-012.
OS: All.

### Happy path

- **Given** Linux host with `bwrap` not in PATH and agent with `sandbox: bubblewrap`
- **When** `Launcher.Exec(ctx, sandbox=bubblewrap, opts)` is called
- **Then** function returns `(nil, *SandboxBackendUnavailable)`
- **And** the wrapped error message names "bwrap" (the missing backend)
- **And** no command is executed (verified by checking that no side-effect occurred)

### Edge case — fallback to `none` is NOT silent

- **Given** same Linux host with no bwrap
- **When** orchestrator catches `*SandboxBackendUnavailable` and chooses to retry with `sandbox: none`
- **Then** RT-003 launcher does NOT auto-fallback; orchestrator must explicitly opt-in via separate `Launcher.Exec(ctx, sandbox=none, opts)` call (and that call may itself fail per AC-08 if `sandbox.required: true`)

### Test mapping

- `internal/sandbox/launcher_test.go::TestLauncher_BackendUnavailable`
- Sentinel: `errors.Is(err, ErrSandboxBackendUnavailable)` returns `true`

---

## AC-V3R2-RT-003-06 — Env scrubbing: `ANTHROPIC_API_KEY` empty in child process

Maps to: REQ-V3R2-RT-003-006.
OS: All.

### Happy path

- **Given** parent env has `ANTHROPIC_API_KEY=sk-ant-12345` and `sandbox: seatbelt|bubblewrap`
- **And** agent frontmatter has no `sandbox.env_passthrough`
- **When** `Launcher.Exec(ctx, sandbox, opts)` invokes `Bash(echo "$ANTHROPIC_API_KEY")`
- **Then** stdout is empty line (`""\n`) — env var was scrubbed before reaching child

### Edge case — 6 default scrubbed patterns

- **Given** parent env has all 6: `AWS_ACCESS_KEY_ID`, `GITHUB_TOKEN`, `ANTHROPIC_API_KEY`, `OPENAI_API_KEY`, `NPM_TOKEN`, `GH_TOKEN` populated
- **When** child invokes `env | grep -E '^(AWS_|GITHUB_TOKEN|ANTHROPIC_API_KEY|OPENAI_API_KEY|NPM_TOKEN|GH_TOKEN)'`
- **Then** zero output lines

### Edge case — `AWS_*` prefix-match (not over-greedy)

- **Given** parent env has `AWS_REGION=us-east-1` (legitimate, scrubbed) and `AWSOME_VAR=hello` (false-positive candidate)
- **When** child invokes `echo "$AWSOME_VAR"`
- **Then** stdout is `hello\n` (NOT scrubbed) — confirms `strings.HasPrefix(k, "AWS_")` strict match

### Test mapping

- `internal/sandbox/env_test.go::TestEnvScrub_DefaultDenylist`
- `internal/sandbox/env_test.go::TestEnvScrub_AWSPrefixOnly`

---

## AC-V3R2-RT-003-07 — `sandbox.env_passthrough: [GH_TOKEN]` preserves variable

Maps to: REQ-V3R2-RT-003-031.
OS: All.

### Happy path

- **Given** agent frontmatter `sandbox.env_passthrough: ["GH_TOKEN"]` and parent env `GH_TOKEN=abc123`
- **When** `Launcher.Exec` invokes `Bash(echo "$GH_TOKEN")` after passthrough wiring
- **Then** stdout is `abc123\n` (preserved despite default scrubbing)

### Edge case — non-listed vars still scrubbed

- **Given** same configuration, parent env also has `NPM_TOKEN=xyz`
- **When** child invokes `echo "$NPM_TOKEN"`
- **Then** stdout is empty (passthrough is list-explicit, not pattern)

### Test mapping

- `internal/sandbox/env_test.go::TestEnvScrub_PassthroughPreserved`

---

## AC-V3R2-RT-003-08 — `security.yaml` `sandbox.required: true` + agent `sandbox: none` (no justification) → spawn fails

Maps to: REQ-V3R2-RT-003-020, REQ-V3R2-RT-003-043.
OS: All.

### Happy path

- **Given** `.moai/config/sections/security.yaml` has `sandbox.required: true`
- **And** agent frontmatter has `sandbox: none` and no `sandbox.justification`
- **When** `Launcher.Exec(ctx, sandbox=none, opts)` is called with `opts.SecurityRequired=true, opts.HasJustification=false`
- **Then** function returns `(nil, *SandboxRequired)` and message names the agent file path
- **And** no command executes

### Edge case — `sandbox: none` + `sandbox.justification: "..."` allows spawn

- **Given** same security.yaml
- **And** agent frontmatter has `sandbox: none` and `sandbox.justification: "dogfooding legacy workflow X"`
- **When** spawn proceeds
- **Then** spawn succeeds; `moai doctor sandbox` displays the justification next to the agent name

### Edge case — CI lint catches missing justification

- **Given** agent file with `sandbox: none` and no `sandbox.justification`
- **When** `moai cli agent_lint <agent-file>` runs
- **Then** lint fails with rule key `AGENT_LINT_NO_SANDBOX_NO_JUSTIFICATION`
- **And** exit code is 1

### Test mapping

- `internal/sandbox/launcher_test.go::TestLauncher_SandboxRequired_NoJustification`
- `internal/cli/agent_lint_test.go::TestAgentLint_NoSandboxNoJustification_Fails`
- `internal/cli/agent_lint_test.go::TestAgentLint_NoSandboxWithJustification_Passes`

---

## AC-V3R2-RT-003-09 — `permissionMode: plan` results in fully read-only mount

Maps to: REQ-V3R2-RT-003-022.
OS: All.

### Happy path (macOS)

- **Given** agent has `permissionMode: plan` (RT-002) and `sandbox: seatbelt`
- **When** `Launcher.Exec` generates the SBPL profile
- **Then** the SBPL contains `(deny file-write* (subpath "/"))` and `(allow file-read* ...)`
- **And** `WritableScope` was emptied before profile generation

### Edge case (Linux)

- **Given** same config, `sandbox: bubblewrap`
- **When** profile is generated
- **Then** bwrap args contain only `--ro-bind` (no `--bind`)

### Test mapping

- `internal/sandbox/profile_test.go::TestProfile_PlanMode_AllPathsReadOnly_SBPL` (macOS)
- `internal/sandbox/profile_test.go::TestProfile_PlanMode_AllPathsReadOnly_Bwrap` (Linux)

---

## AC-V3R2-RT-003-10 — `security.yaml` `sandbox.network_allowlist` extends defaults

Maps to: REQ-V3R2-RT-003-008, REQ-V3R2-RT-003-030.
OS: All.

### Happy path

- **Given** `security.yaml` has `sandbox.network_allowlist: ["internal.company.com"]`
- **When** `Launcher.Exec` resolves effective allowlist
- **Then** the resulting list contains all 8 default hosts + `internal.company.com` (additive, not replacement)

### Edge case — connection to allowlist host succeeds

- **Given** Linux host with bwrap + above config
- **When** sandboxed process invokes `curl https://internal.company.com`
- **Then** connection succeeds (assuming valid DNS + reachable internal endpoint, gated `MOAI_TEST_NETWORK=1`)

### Test mapping

- `internal/sandbox/launcher_test.go::TestLauncher_NetworkAllowlist_Extension` (verify list construction; gated network test optional)

---

## AC-V3R2-RT-003-11 — `CI=1` env auto-selects docker backend

Maps to: REQ-V3R2-RT-003-015.
OS: CI.

### Happy path

- **Given** Linux host with `CI=1` env set, `bwrap` available, `docker` available
- **And** agent declared `sandbox: bubblewrap`
- **When** `Launcher.resolveBackend(declared=bubblewrap)` is called
- **Then** returns `docker` (CI override wins)

### Edge case — explicit `--sandbox bubblewrap` flag overrides CI=1

- **Given** same env + explicit launcher flag (programmatic `Launcher.Exec(forceBackend=bubblewrap)`)
- **When** dispatch happens
- **Then** returns `bubblewrap` (explicit flag > CI=1 > declared > OS default)

### Test mapping

- `internal/sandbox/launcher_test.go::TestLauncher_CIOverride`
- `internal/sandbox/docker_test.go::TestDocker_ExecHello` (`MOAI_TEST_DOCKER=1` gated)

---

## AC-V3R2-RT-003-12 — Setuid escalation (`sudo`, `su`) is denied

Maps to: REQ-V3R2-RT-003-040.
OS: macOS / Linux.

### Happy path (Linux)

- **Given** sandboxed process with `sandbox: bubblewrap`
- **When** the process invokes `sudo systemctl restart foo`
- **Then** `sudo` fails (bwrap `--unshare-user` strips setuid bit; `sudo: effective uid is not 0`)
- **And** SystemMessage names the attempted setuid escalation

### Happy path (macOS)

- **Given** sandboxed process with `sandbox: seatbelt`
- **When** process invokes `sudo ls /etc`
- **Then** SBPL `(deny process-exec* (literal "/usr/bin/sudo"))` rejects → EPERM

### Test mapping

- `internal/sandbox/bubblewrap_test.go::TestBubblewrap_SetuidDenied` (Linux)
- `internal/sandbox/seatbelt_test.go::TestSeatbelt_SetuidDenied` (macOS)
- Sentinel: `errors.Is(err, ErrSandboxSetuidDenied)` returns `true`

---

## AC-V3R2-RT-003-13 — Profile generation produces invalid SBPL/bwrap → `SandboxProfileInvalid`

Maps to: REQ-V3R2-RT-003-041.
OS: All.

### Happy path

- **Given** `SandboxOptions.WritableScope = ["/tmp\x00null-byte"]` (intentionally null-byte injected)
- **When** `generateSBPL(opts)` is called
- **Then** function returns `("", *SandboxProfileInvalid)` and error names the offending field "WritableScope[0]" + reason "null byte not permitted"

### Edge case — empty WritableScope is valid

- **Given** `WritableScope = []` (empty, valid for plan mode)
- **When** generator is called
- **Then** returns valid SBPL with no `file-write*` allow clauses (plan-mode profile)

### Test mapping

- `internal/sandbox/profile_test.go::TestProfile_GenerateSBPL_NullByteRejected`
- `internal/sandbox/profile_test.go::TestProfile_GenerateBwrapArgs_NullByteRejected`

---

## AC-V3R2-RT-003-14 — Output truncation at 16 MiB + SystemMessage

Maps to: REQ-V3R2-RT-003-042.
OS: All.

### Happy path

- **Given** sandboxed process produces 32 MiB to stdout (e.g., `dd if=/dev/zero bs=1M count=32 status=none`)
- **When** `Launcher.Exec` returns
- **Then** returned `[]byte` length is exactly 16,777,216 (16 MiB)
- **And** SystemMessage names "output truncated at 16 MiB" + the agent name

### Edge case — small output (< 16 MiB) untruncated

- **Given** process produces 1 MiB
- **When** Exec returns
- **Then** length is 1,048,576 (full size, no truncation)

### Test mapping

- `internal/sandbox/launcher_test.go::TestLauncher_OutputTruncation_SystemMessage`
- `internal/sandbox/launcher_test.go::TestLauncher_OutputTruncation16MiB`
- Sentinel: `errors.Is(err, ErrSandboxOutputTruncated)` returns `true`; output buffer length == 16,777,216

---

## AC-V3R2-RT-003-15 — `sandbox: none` with valid `sandbox.justification` allows spawn + lint pass

Maps to: REQ-V3R2-RT-003-033.
OS: All.

### Happy path

- **Given** agent file with frontmatter:
  ```yaml
  sandbox: none
  sandbox.justification: "dogfooding legacy workflow X — needs unrestricted Bash for v2.x compat"
  ```
- **And** `security.yaml` has `sandbox.required: false` (default)
- **When** `moai cli agent_lint <agent-file>` runs
- **Then** lint exits 0 (PASS)

### Edge case — `moai doctor sandbox` displays justification

- **Given** same agent
- **When** user runs `moai doctor sandbox`
- **Then** the agent's row shows `<name>: sandbox=none, justification="dogfooding..."`

### Test mapping

- `internal/cli/agent_lint_test.go::TestAgentLint_NoSandboxWithJustification_Passes`
- `internal/cli/doctor_sandbox_test.go::TestDoctorSandbox_DisplaysJustification`

---

## AC-V3R2-RT-003-16 — Permission allow + sandbox deny → sandbox wins + SystemMessage divergence

Maps to: REQ-V3R2-RT-003-051.
OS: All.

### Happy path

- **Given** RT-002 mock `permission.Decide(action="Bash(rm -rf /tmp/test)") = "allow"`
- **And** sandbox profile denies writes to `/tmp` (only `WritableScope=["/home/user/worktree"]`)
- **When** `Launcher.Exec(ctx, sandbox=bubblewrap, action="Bash(rm -rf /tmp/test)")` is called
- **Then** sandbox blocks → returns `(nil, *SandboxFileWriteEPERM)` (or similar)
- **And** SystemMessage line names: `permission-sandbox divergence: action=allowed, sandbox=denied; sandbox verdict wins`
- **And** `/tmp/test` is not deleted on host

### Test mapping

- `internal/sandbox/launcher_test.go::TestLauncher_PermissionDenyDivergence` (mock-based; real RT-002 wiring deferred to RT-002 merge)

---

## AC-V3R2-RT-003-17 — Bubblewrap startup p99 ≤ 50ms

Maps to: spec.md §7 Constraints.
OS: Linux.

### Happy path

- **Given** Linux host with bwrap available, `BenchmarkSandbox_BwrapHello` runs 1000 iterations of `bwrap ... -- /bin/echo hi`
- **When** benchmark completes
- **Then** p99 latency ≤ 50ms
- **And** `progress.md` records the measured p99 number

### Test mapping

- `internal/sandbox/launcher_bench_test.go::BenchmarkSandbox_BwrapHello`
- Recording: `progress.md` "M6 benchmark results" subsection

---

## AC-V3R2-RT-003-18 — Seatbelt startup p99 ≤ 50ms

Maps to: spec.md §7 Constraints.
OS: macOS.

### Happy path

- **Given** macOS host, `BenchmarkSandbox_SeatbeltHello` runs 1000 iterations of `sandbox-exec -p <profile> /bin/echo hi`
- **When** benchmark completes
- **Then** p99 latency ≤ 50ms

### Test mapping

- `internal/sandbox/launcher_bench_test.go::BenchmarkSandbox_SeatbeltHello`

---

## AC-V3R2-RT-003-19 — Steady-state syscall overhead ≤ 10% vs unsandboxed baseline

Maps to: spec.md §7 Constraints.
OS: All.

### Happy path

- **Given** bench compares 1000 iterations of `sh -c "echo hi"` with sandbox vs without
- **When** benchmark completes
- **Then** sandboxed walltime ≤ 1.10 × unsandboxed walltime

### Test mapping

- `internal/sandbox/launcher_bench_test.go::BenchmarkSandbox_NestedExec` (gated per OS)

---

## AC-V3R2-RT-003-21 — LSP carve-out clause baseline (alpha.2 deferred validation)

Maps to: REQ-V3R2-RT-003-021.
OS: All (baseline); alpha.2 manual validation per master §12 Q7.

### Happy path (baseline only)

- **Given** generator produces SBPL/bwrap args for an implementer agent
- **When** profile is inspected
- **Then** profile includes:
  - macOS SBPL: `(allow file-read* (subpath "<HOME>/.cache"))` and `(allow file-write* (subpath "<HOME>/.cache"))` and `(allow file-write* (subpath "/tmp"))`
  - Linux bwrap: `--bind <HOME>/.cache <HOME>/.cache` and `--tmpfs /tmp`
- **And** the carve-out is unconditional (always present in implementer profiles, regardless of `permissionMode`)

### Deferred (alpha.2)

- Full runtime validation that pyright/tsserver/rust-analyzer/gopls actually start and serve `moai_lsp_*` ACI commands inside the sandbox profile is deferred to alpha.2 manual verification per master §12 Q7. This AC only checks the profile clause is generated.

### Test mapping

- `internal/sandbox/profile_test.go::TestProfile_LSPCarveOut_SBPL` (macOS)
- `internal/sandbox/profile_test.go::TestProfile_LSPCarveOut_BwrapArgs` (Linux)

---

## Coverage matrix (REQ → AC summary)

총 33 EARS REQs × 20 ACs cover. plan §1.4 traceability matrix 와 1:1 정합:

| REQ category | REQ count | Mapped AC count | AC IDs |
|--------------|-----------|-----------------|--------|
| Ubiquitous (5.1) | 8 | 9 | AC-01 (write scope EPERM via macOS), AC-02 (network block via Linux), AC-03 (default mapping), AC-04 (doctor report), AC-06 (env scrub), AC-09 (plan mode read-only), AC-10 (allowlist extension), AC-13 (profile checksum), AC-21 (LSP carve-out baseline) |
| Event-Driven (5.2) | 6 | 6 | AC-01 (macOS), AC-02 (Linux), AC-05 (backend unavailable), AC-09 (write scope), AC-11 (CI override), AC-14 (output truncation) |
| State-Driven (5.3) | 3 | 3 | AC-08 (sandbox required), AC-09 (plan mode), AC-21 (LSP carve-out) |
| Optional (5.4) | 4 | 4 | AC-04 (doctor profile flag), AC-07 (env passthrough), AC-10 (allowlist extension), AC-15 (justification) |
| Unwanted (5.5) | 4 | 3 | AC-12 (setuid), AC-13 (invalid profile), AC-14 (output truncation) — REQ-043 covered by AC-08 |
| Complex (5.6) | 2 | 2 | AC-04 + AC-16 |
| Performance budget | 0 (REQ implicit in spec §7) | 3 | AC-17/18/19 |
| **Total** | **33** | **20** (= 16 baseline ACs + 3 perf + 1 LSP baseline; some ACs cover multiple REQ IDs) |

→ 100% REQ coverage. plan-auditor PASS criterion #2 충족.

---

## Cross-test gating summary

per-OS gating 정책 명시 (M1 RED 이후 모든 milestone 일관 적용):

| Test pattern | Gate condition | Skip behavior |
|--------------|----------------|---------------|
| `TestBubblewrap_*`, `TestProfile_*BwrapArgs*` | `runtime.GOOS == "linux"` && `exec.LookPath("bwrap") == nil` | `t.Skipf("bwrap unavailable: %v", err)` |
| `TestSeatbelt_*`, `TestProfile_*SBPL*` | `runtime.GOOS == "darwin"` | `t.Skipf("sandbox-exec requires macOS")` |
| `TestDocker_*` | `os.Getenv("MOAI_TEST_DOCKER") == "1"` && docker daemon ping | `t.Skip("Docker tests gated; set MOAI_TEST_DOCKER=1")` |
| `TestEnvScrub_*`, `TestSandbox_EnumExhaustive` | (none — All OS) | (always run) |
| `TestLauncher_ResolveBackend_AllScenarios` | (none — uses mocks; runs on All) | (always run; CI=1 simulated via env var injection in test) |

→ developer ergonomic: macOS dev 가 `go test ./internal/sandbox/` 실행 시 ~70% test 가 RUN/GREEN, ~30% skip (Linux+Docker). CI Linux runner 에서 ~80% RUN/GREEN. CI 매트릭스가 macOS + Linux 양쪽 보장 (master §10 R3 mitigation).

---

End of acceptance.md v0.1.0.

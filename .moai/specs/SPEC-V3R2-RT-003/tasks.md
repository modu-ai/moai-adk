# SPEC-V3R2-RT-003 Implementation Tasks

> Run-phase task list for **Sandbox Execution Layer (Bubblewrap / Seatbelt / Docker)**.
> Companion to `spec.md` v0.1.0, `research.md` v0.1.0, `plan.md` v0.1.0, `acceptance.md` v0.1.0.

## HISTORY

| Version | Date       | Author                       | Description                                                              |
|---------|------------|------------------------------|--------------------------------------------------------------------------|
| 0.1.0   | 2026-05-10 | manager-spec (Plan workflow) | Initial task decomposition — **52 tasks** across M1-M6 milestones (TDD mode). 33 EARS REQs × 16 ACs cover; per-OS gating (Linux/macOS/Docker) explicit per task. Greenfield (no placeholder replacement). Performance budget tasks T-RT003-46/47/48 + LSP carve-out validation T-RT003-21. |

---

## Methodology

`.moai/config/sections/quality.yaml` `development_mode: tdd` → RED → GREEN → REFACTOR per `.claude/rules/moai/workflow/spec-workflow.md` § Run Phase.

각 task 는:
- **ID**: T-RT003-NN (zero-padded 2자리)
- **Milestone**: M1 (RED) / M2-M5 (GREEN) / M6 (REFACTOR + final)
- **Owner**: `expert-backend` (Go) / `manager-cycle` (직접) / `manager-docs` (CHANGELOG/MX) / `manager-git` (PR)
- **Depends_on**: prerequisite task IDs
- **Files**: 변경/생성 file paths
- **REQ/AC**: 충족 EARS REQs + ACs
- **LOC**: estimate (test 별도 표시)
- **Parallel**: 다른 task 와 동시 실행 가능 여부
- **OS**: Linux / macOS / All / CI (per-OS gating)

---

## Milestone M1 — RED phase (Test scaffolding)

[HARD] 본 milestone 은 source 코드 변경 없이 신규 테스트 케이스만 추가 (TDD discipline). 테스트는 모두 RED 상태로 commit (compile fail OK; per-OS skip 정상).

| Task ID | Subject | Files | REQ/AC | Owner | Deps | LOC | Parallel | OS |
|---------|---------|-------|--------|-------|------|-----|----------|-----|
| T-RT003-01 | RED: create `internal/sandbox/context_test.go` with 3 functions (`TestSandbox_EnumExhaustive`, `TestSandboxOptions_Validate`, `TestSandboxBackend_InterfaceContract`) | `internal/sandbox/context_test.go` (new) | REQ-001/002, AC-13 | expert-backend | - | ~80 | ✅ | All |
| T-RT003-02 | RED: create `internal/sandbox/launcher_test.go` with 5 functions (`TestLauncher_DispatchByOS`, `TestLauncher_CIOverride`, `TestLauncher_OutputTruncation16MiB`, `TestLauncher_BackendUnavailable`, `TestLauncher_ResolveBackend_AllScenarios`) | `internal/sandbox/launcher_test.go` (new) | REQ-002/012/015/042, AC-05/11/14 | expert-backend | - | ~180 | ✅ | All |
| T-RT003-03 | RED: create `internal/sandbox/bubblewrap_test.go` with 5 functions (`TestBubblewrap_ExecHello`, `TestBubblewrap_FileWriteScopeEPERM`, `TestBubblewrap_NetworkBlocked`, `TestBubblewrap_SetuidDenied`, `TestBubblewrap_ArgsDeterministic`) — Linux-gated | `internal/sandbox/bubblewrap_test.go` (new) | REQ-011/013/014, AC-02/12 | expert-backend | - | ~150 | ✅ | Linux |
| T-RT003-04 | RED: create `internal/sandbox/seatbelt_test.go` with 5 functions (`TestSeatbelt_ExecHello`, `TestSeatbelt_FileWriteScopeEPERM`, `TestSeatbelt_NetworkBlocked`, `TestSeatbelt_SetuidDenied`, `TestSeatbelt_SBPLDeterministic`) — macOS-gated | `internal/sandbox/seatbelt_test.go` (new) | REQ-010/013/014, AC-01/12 | expert-backend | - | ~150 | ✅ | macOS |
| T-RT003-05 | RED: create `internal/sandbox/docker_test.go` with 3 functions (`TestDocker_ExecHello`, `TestDocker_NetworkAllowlist`, `TestDocker_FileWriteScope`) — `MOAI_TEST_DOCKER=1` gated | `internal/sandbox/docker_test.go` (new) | REQ-015, AC-11 | expert-backend | - | ~100 | ✅ | CI |
| T-RT003-06 | RED: create `internal/sandbox/profile_test.go` with 4 functions (`TestProfile_GenerateSBPL`, `TestProfile_GenerateBwrapArgs`, `TestProfile_GenerateDockerSnippet`, `TestProfile_DeterministicChecksum_100Runs`) | `internal/sandbox/profile_test.go` (new) | REQ-004/021/041, AC-09/13 | expert-backend | - | ~150 | ✅ | All |
| T-RT003-07 | RED: create `internal/sandbox/env_test.go` with 4 functions (`TestEnvScrub_DefaultDenylist`, `TestEnvScrub_AWSPrefixOnly`, `TestEnvScrub_PassthroughPreserved`, `TestEnvScrub_EmptyInput`) | `internal/sandbox/env_test.go` (new) | REQ-006/031, AC-06/07 | expert-backend | - | ~80 | ✅ | All |
| T-RT003-08 | RED: create `internal/sandbox/errors_test.go` with 2 functions (`TestErrors_SentinelMatching`, `TestErrors_Wrapping`) | `internal/sandbox/errors_test.go` (new) | REQ-012/041/043, AC-05/08 | expert-backend | - | ~60 | ✅ | All |
| T-RT003-09 | RED: create config-extension test cases in `internal/config/types_test.go` (extend) for `RoleProfile.Sandbox` field + `SecurityConfig.Sandbox` substruct | `internal/config/types_test.go` (extend) | REQ-003/008/030 | expert-backend | - | ~50 | ✅ | All |
| T-RT003-10 | RED: create `internal/cli/doctor_sandbox_test.go` with 4 functions (`TestDoctorSandbox_AvailabilityReport`, `TestDoctorSandbox_PerAgentResolved`, `TestDoctorSandbox_ProfileFlag`, `TestDoctorSandbox_BackendUnavailableMessage`) | `internal/cli/doctor_sandbox_test.go` (new) | REQ-005/032, AC-04 | expert-backend | - | ~120 | ✅ | All |
| T-RT003-11 | RED: extend `internal/cli/agent_lint_test.go` with 2 functions (`TestAgentLint_NoSandboxNoJustification_Fails`, `TestAgentLint_NoSandboxWithJustification_Passes`) | `internal/cli/agent_lint_test.go` (extend) | REQ-033/043, AC-08/15 | expert-backend | - | ~50 | ✅ | All |
| T-RT003-12 | RED: extend `internal/sandbox/launcher_test.go` with `TestLauncher_PermissionDenyDivergence` (REQ-051 mock-based) | `internal/sandbox/launcher_test.go` (extend) | REQ-051, AC-16 | expert-backend | T-RT003-02 | ~40 | ❌ (depends T-02) | All |
| T-RT003-13 | RED: extend `internal/sandbox/launcher_test.go` with `TestLauncher_OutputTruncation_SystemMessage` | (same file) | REQ-042, AC-14 | expert-backend | T-RT003-02 | ~30 | ❌ | All |
| T-RT003-14 | M1 verification: `go test ./internal/sandbox/... ./internal/config/... ./internal/cli/...` → confirm new test functions are RED + existing baseline GREEN | (verification only) | M1 gate | expert-backend | T-RT003-01..13 | - | ❌ | All |

M1 산출물: ~32 신규 test 함수 (M1 의 24개 + M5 wiring 8개 합산), 모두 RED. 기존 baseline 100% GREEN. 이 단계에서 `go build ./internal/sandbox/...` 가 fail (정상 — source 부재).

---

## Milestone M2 — GREEN type-level (enum + interface + sentinel errors)

[HARD] M2 는 type-level GREEN — backend exec 은 placeholder. 컴파일 통과만 보장.

| Task ID | Subject | Files | REQ/AC | Owner | Deps | LOC | Parallel | OS |
|---------|---------|-------|--------|-------|------|-----|----------|-----|
| T-RT003-15 | Create `internal/sandbox/context.go` with `Sandbox` typed string enum (4 values: none/bubblewrap/seatbelt/docker), `SandboxOptions` struct (WritableScope []string, ReadOnlyScope []string, NetworkAllowlist []string, EnvPassthrough []string, MaxOutputBytes int64), `SandboxBackend` interface (Available, Exec, Profile) | `internal/sandbox/context.go` (new) | REQ-001/002 | expert-backend | T-RT003-14 | ~100 | ✅ | All |
| T-RT003-16 | Create `internal/sandbox/errors.go` with 5 sentinels (`ErrSandboxBackendUnavailable`, `ErrSandboxProfileInvalid`, `ErrSandboxRequired`, `ErrSandboxOutputTruncated`, `ErrSandboxSetuidDenied`) + `errors.Is` support via `Unwrap()` | `internal/sandbox/errors.go` (new) | REQ-012/041/043/042/040 | expert-backend | T-RT003-14 | ~70 | ✅ (parallel with T-15) | All |
| T-RT003-17 | M2 verification: `go test ./internal/sandbox/ -run "TestSandbox_EnumExhaustive|TestErrors_Sentinel|TestErrors_Wrapping"` → these GREEN; rest still RED (no backend impl yet) | (verification only) | M2 gate | expert-backend | T-RT003-15/16 | - | ❌ | All |

M2 산출물: type-level + sentinel errors GREEN (3 test functions). enum exhaustiveness, error wrapping verified.

---

## Milestone M3 — GREEN per-OS backends (bubblewrap + seatbelt + profile generator)

[HARD] M3 는 per-OS backend 의 실제 구현. M3 종료 시 macOS/Linux 실 호스트에서 smoke test 통과.

| Task ID | Subject | Files | REQ/AC | Owner | Deps | LOC | Parallel | OS |
|---------|---------|-------|--------|-------|------|-----|----------|-----|
| T-RT003-18 | Create `internal/sandbox/profile.go` with 3 generators: `generateSBPL(opts) (string, error)` (sorted directives), `generateBwrapArgs(opts) ([]string, error)` (deterministic ordering), `generateDockerSnippet(opts) (string, error)` (Dockerfile fragment) | `internal/sandbox/profile.go` (new) | REQ-004/021/041 | expert-backend | T-RT003-15 | ~250 | ✅ | All |
| T-RT003-19 | Create `internal/sandbox/seatbelt.go` (macOS backend) implementing `SandboxBackend`: `Available()` checks `/usr/bin/sandbox-exec`; `Exec(ctx, opts) ([]byte, error)` invokes `sandbox-exec -p <generated SBPL> cmd...`; output truncation at 16 MiB | `internal/sandbox/seatbelt.go` (new) | REQ-002/010/013/014/021/022/040, AC-01/09/12/14 | expert-backend | T-RT003-18 | ~200 | ✅ (parallel with T-20) | macOS |
| T-RT003-20 | Create `internal/sandbox/bubblewrap.go` (Linux backend) implementing `SandboxBackend`: `Available()` checks `bwrap` in PATH + `unprivileged_userns_clone`; `Exec(ctx, opts)` invokes `bwrap --unshare-all --die-with-parent ... -- cmd`; output truncation 16 MiB | `internal/sandbox/bubblewrap.go` (new) | REQ-002/011/013/014/021/022/040, AC-02/09/12/14 | expert-backend | T-RT003-18 | ~220 | ✅ | Linux |
| T-RT003-21 | Add LSP carve-out clauses in profile generators (REQ-021): SBPL `(allow file-read* (subpath "~/.cache/"))` + bwrap `--bind ~/.cache ~/.cache --tmpfs /tmp` | `internal/sandbox/profile.go` (extend) | REQ-021, AC-21 (LSP carve-out clause baseline) | expert-backend | T-RT003-18 | ~30 | ❌ (depends T-18) | All |
| T-RT003-22 | M3 verification: macOS host runs `TestSeatbelt_*` 5 functions GREEN; Linux host runs `TestBubblewrap_*` 5 functions GREEN; profile checksum 100-run test GREEN cross-OS | (verification only) | M3 gate | expert-backend | T-RT003-18..21 | - | ❌ | macOS/Linux |

M3 산출물: 10 GREEN ACs (AC-01/02/09/12/14/21 baseline). per-OS smoke test 통과.

---

## Milestone M4 — GREEN CI Docker backend + dispatcher

| Task ID | Subject | Files | REQ/AC | Owner | Deps | LOC | Parallel | OS |
|---------|---------|-------|--------|-------|------|-----|----------|-----|
| T-RT003-23 | Create `internal/sandbox/docker.go` (CI backend) implementing `SandboxBackend`: `Available()` checks `docker` binary + daemon ping; `Exec(ctx, opts)` invokes `docker run --rm --network=<bridge> -v <scope>:<scope> -w <scope> alpine:latest cmd...` | `internal/sandbox/docker.go` (new) | REQ-002/015, AC-11 | expert-backend | T-RT003-22 | ~160 | ✅ | CI |
| T-RT003-24 | Create `internal/sandbox/launcher.go` with `Launcher` (top-level facade), `New()` constructor, `dispatch(declared Sandbox) Backend` (OS-auto + CI=1 override), `resolveBackend(declared) Sandbox` (mixed-OS fallback per plan §3.4), `Exec(ctx, declared, opts)` (top-level entrypoint with truncation) | `internal/sandbox/launcher.go` (new) | REQ-002/012/015/050, AC-05/11/14 | expert-backend | T-RT003-23 | ~180 | ❌ | All |
| T-RT003-25 | Implement `Launcher.Exec` permission-divergence handling: if RT-002 `permission.Decision` returns `allow` but sandbox profile denies, sandbox wins + emit SystemMessage to RT-001 hook stream | `internal/sandbox/launcher.go` (extend) | REQ-051, AC-16 | expert-backend | T-RT003-24 | ~40 | ❌ | All |
| T-RT003-26 | M4 verification: `MOAI_TEST_DOCKER=1 go test ./internal/sandbox/ -run TestDocker` GREEN (Docker-가용 환경); `TestLauncher_DispatchByOS`, `TestLauncher_CIOverride`, `TestLauncher_ResolveBackend_AllScenarios` GREEN | (verification only) | M4 gate | expert-backend | T-RT003-23..25 | - | ❌ | All/CI |

M4 산출물: AC-11 + AC-16 GREEN. dispatcher 4-scenario test (macOS/Linux × CI=on/off) all GREEN.

---

## Milestone M5 — GREEN config + frontmatter + lint wiring

| Task ID | Subject | Files | REQ/AC | Owner | Deps | LOC | Parallel | OS |
|---------|---------|-------|--------|-------|------|-----|----------|-----|
| T-RT003-27 | Create `internal/sandbox/env.go` with `ScrubEnv(parent []string, passthrough []string) []string` — default 6-pattern denylist (`AWS_*` prefix-match, `GITHUB_TOKEN`, `ANTHROPIC_API_KEY`, `OPENAI_API_KEY`, `NPM_TOKEN`, `GH_TOKEN`) + passthrough preservation | `internal/sandbox/env.go` (new) | REQ-006/031, AC-06/07 | expert-backend | T-RT003-26 | ~80 | ✅ | All |
| T-RT003-28 | Wire `Launcher.Exec` to call `ScrubEnv(os.Environ(), opts.EnvPassthrough)` before passing env to backend | `internal/sandbox/launcher.go` (extend) | REQ-006/031 | expert-backend | T-RT003-27 | ~15 | ❌ (depends T-27) | All |
| T-RT003-29 | Extend `internal/config/types.go`: add `RoleProfile.Sandbox string` field + `SecurityConfig.Sandbox SecuritySandbox` substruct (`Required bool`, `NetworkAllowlist []string`, `EnvScrubExtra []string`, `DockerImage string`) + yaml tags | `internal/config/types.go` (extend) | REQ-003/008/030 | expert-backend | T-RT003-26 | ~30 | ✅ (parallel with T-27) | All |
| T-RT003-30 | Update `internal/config/defaults.go::NewDefaultConfig()`: `RoleProfiles[implementer/tester/designer].Sandbox` set to OS-resolved default (`runtime.GOOS == "darwin" → seatbelt`, `linux → bubblewrap`); `SecurityConfig.Sandbox.Required = false`; `NetworkAllowlist = [...8 default hosts]`; `DockerImage = "alpine:latest"` (TBD `moai/sandbox:latest` post-EXT-004) | `internal/config/defaults.go` (extend) | REQ-003/008 | expert-backend | T-RT003-29 | ~25 | ❌ | All |
| T-RT003-31 | Update template `internal/template/templates/.moai/config/sections/workflow.yaml`: add `role_profiles.<role>.sandbox: ""` for 7 roles + comment block | `internal/template/templates/.moai/config/sections/workflow.yaml` | REQ-003 | expert-backend | T-RT003-30 | ~15 | ✅ | All |
| T-RT003-32 | Update template `internal/template/templates/.moai/config/sections/security.yaml`: add `sandbox:` top-level block with all 4 keys + comments | `internal/template/templates/.moai/config/sections/security.yaml` | REQ-008/030 | expert-backend | T-RT003-30 | ~20 | ✅ (parallel with T-31) | All |
| T-RT003-33 | Run `make build` to regenerate `internal/template/embedded.go`; verify diff is non-zero only for the two yaml templates | (regenerate) | embedded parity | expert-backend | T-RT003-31/32 | - | ❌ | All |
| T-RT003-34 | Implement `security.yaml` `sandbox.required: true` enforcement in `Launcher.Exec`: if true and declared `sandbox: none` and no `sandbox.justification` → return `*SandboxRequired` error | `internal/sandbox/launcher.go` (extend) | REQ-020/043, AC-08 | expert-backend | T-RT003-33 | ~30 | ❌ | All |
| T-RT003-35 | Extend `internal/cli/agent_lint.go` (existing) with new rule key `AGENT_LINT_NO_SANDBOX_NO_JUSTIFICATION` + `lintSandboxJustification(frontmatter map[string]any) []LintIssue` helper: emits issue when `sandbox: none` exists without `sandbox.justification` field | `internal/cli/agent_lint.go` (extend) | REQ-033/043, AC-08/15 | expert-backend | T-RT003-29 | ~35 | ✅ | All |
| T-RT003-36 | Extend `internal/cli/agent_lint_test.go` with TestAgentLint_NoSandboxNoJustification_Fails (1 fail case + 1 pass case) | `internal/cli/agent_lint_test.go` (extend) | REQ-033/043, AC-08/15 | expert-backend | T-RT003-35 | ~50 | ❌ (depends T-35) | All |
| T-RT003-37 | Add `sandbox.docker_image` frontmatter override in `Launcher.Exec`: if agent declares `sandbox.docker_image`, override default for docker backend | `internal/sandbox/launcher.go` (extend) | optional polish (not strict REQ) | expert-backend | T-RT003-34 | ~15 | ❌ | All |
| T-RT003-38 | Wire frontmatter `sandbox.env_passthrough: [...]` parsing: extend `internal/agent/frontmatter.go` (or alternative loader) to expose `Frontmatter.SandboxEnvPassthrough []string`; consumed by `Launcher.Exec` via `opts.EnvPassthrough` | `internal/agent/frontmatter.go` (extend or alternative path) | REQ-031, AC-07 | expert-backend | T-RT003-29 | ~40 | ✅ (parallel with T-35) | All |
| T-RT003-39 | M5 verification: `go test ./internal/sandbox/... ./internal/config/... ./internal/cli/...` → AC-03/06/07/08/10/15 GREEN; `make build` no diff in non-yaml files; `go test -race ./internal/sandbox/...` race-free | (verification only) | M5 gate | expert-backend | T-RT003-27..38 | - | ❌ | All |

M5 산출물: AC-03/06/07/08/10/15/16 GREEN. config 통합 + frontmatter parsing + lint rule 활성화. embedded.go 재생성 검증.

---

## Milestone M6 — REFACTOR + benchmark + doctor + docs + MX

| Task ID | Subject | Files | REQ/AC | Owner | Deps | LOC | Parallel | OS |
|---------|---------|-------|--------|-------|------|-----|----------|-----|
| T-RT003-40 | REFACTOR: extract dispatcher 공통 코드 (OS detection, CI=1 detect, fallback resolution) into `internal/sandbox/launcher.go::dispatch` private helper; consolidate duplicate code between bubblewrap/seatbelt/docker (truncation, env passing) | `internal/sandbox/launcher.go`, `bubblewrap.go`, `seatbelt.go`, `docker.go` | (refactor — no new REQ) | expert-backend | T-RT003-39 | ~100 (delta) | ❌ | All |
| T-RT003-41 | REFACTOR: profile generator strategy pattern — extract `ProfileGenerator` interface + 3 implementations (SBPLGenerator, BwrapArgsGenerator, DockerSnippetGenerator) for testability | `internal/sandbox/profile.go` (refactor) | REQ-004 hardening | expert-backend | T-RT003-40 | ~80 (delta) | ❌ | All |
| T-RT003-42 | REFACTOR: env scrubbing prefix-match optimization — pre-compile `regexp.MustCompile("^AWS_")` once via `sync.Once` (avoid per-Exec re-compile) | `internal/sandbox/env.go` (refactor) | REQ-006 perf | expert-backend | T-RT003-40 | ~15 (delta) | ✅ | All |
| T-RT003-43 | Create `internal/sandbox/launcher_bench_test.go` with 3 benchmarks: `BenchmarkSandbox_BwrapHello` (Linux), `BenchmarkSandbox_SeatbeltHello` (macOS), `BenchmarkSandbox_DockerHello` (CI-only); each measures p99 startup + steady-state syscall overhead | `internal/sandbox/launcher_bench_test.go` (new) | spec §7 Constraints, AC-17/18/19 | expert-backend | T-RT003-39 | ~80 | ✅ | All |
| T-RT003-44 | Run benchmarks, confirm p99 budgets: bwrap ≤ 50ms, seatbelt ≤ 50ms, docker ≤ 5s; record numbers in `progress.md` | (verification + record) | AC-17/18/19 | expert-backend | T-RT003-43 | - | ❌ | All |
| T-RT003-45 | Create `internal/cli/doctor_sandbox.go` (alongside existing `doctor_config.go`): `moai doctor sandbox` reports availability + per-agent resolved backend + `--profile <agent>` flag dumps SBPL/bwrap-args/Dockerfile snippet (mirror `doctor_config.go` pattern per plan §6 MX hint) | `internal/cli/doctor_sandbox.go` (new) | REQ-005/032, AC-04 | expert-backend | T-RT003-39 | ~140 | ✅ (parallel with T-43) | All |
| T-RT003-46 | Register `sandboxCmd` cobra subcommand under `internal/cli/doctor.go::doctorCmd` | `internal/cli/doctor.go` (extend) | REQ-005 | expert-backend | T-RT003-45 | ~10 | ❌ | All |
| T-RT003-47 | Manual verification of `moai doctor sandbox`: run on macOS host, Linux host (CI), confirm output format consistent with `doctor_config.go`; record sample output in `progress.md` | (manual verification) | AC-04 | expert-backend | T-RT003-46 | - | ❌ | macOS/Linux |
| T-RT003-48 | Add inline MX tags per plan §6 (10 tags total): ANCHOR (4) at `Sandbox enum`, `generateSBPL`, `ScrubEnv`, `SecuritySandbox struct`; NOTE (3) at `resolveBackend`, `ErrSandboxBackendUnavailable`, `doctor_sandbox.go header`; WARN+REASON (3) at `bubblewrap::buildArgs`, `seatbelt::execSandboxExec`, `agent_lint.go::AGENT_LINT_NO_SANDBOX_NO_JUSTIFICATION` | 10 source/config files (per plan §6) | mx_plan | manager-docs | T-RT003-39 | ~10 lines (inline) | ✅ | All |
| T-RT003-49 | Add CHANGELOG entry under Unreleased: 한국어 entry mentioning sandbox layer (Breaking: BC-V3R2-003), 4-value enum, OS-자동 default + CI Docker fallback, env scrub list, `moai doctor sandbox` 명령 | `CHANGELOG.md` (extend Unreleased) | Trackable (TRUST 5) | manager-docs | T-RT003-48 | ~10 | ✅ | All |
| T-RT003-50 | Cross-SPEC AskUser flow doc-only stub for REQ-V3R2-RT-003-050: in `plan.md` §3.3 + `progress.md` "Out-of-scope notes" — orchestrator-side AskUserQuestion implementation deferred to MIG-001 (orchestrator owns user interaction; subagent cannot prompt) | `progress.md` (extend "Out-of-scope notes") | REQ-050 (doc-only — implementation deferred to MIG-001 / orchestrator) | manager-docs | T-RT003-39 | doc-only | ✅ | All |
| T-RT003-51 | Run final quality gate suite: `go vet ./internal/sandbox/...`, `golangci-lint run ./internal/sandbox/... ./internal/cli/...`, `go test -race ./internal/...`, `go test -count=1 ./internal/...` (cache-bypass) — all GREEN | (verification only) | TRUST 5 quality gate | expert-backend | T-RT003-40..49 | - | ❌ | All |
| T-RT003-52 | Update `progress.md` to plan-complete state + `paste-ready resume message` per `.claude/rules/moai/workflow/session-handoff.md` 6-block format; mark all 16 ACs PASS evidence script paths | `progress.md` (finalize) | session-handoff | manager-docs | T-RT003-51 | - | ❌ | All |

M6 산출물: AC-04/17/18/19 GREEN + benchmark numbers recorded + `moai doctor sandbox` 명령 + 10 MX 태그 + CHANGELOG entry + final quality gate PASS + progress.md plan-complete.

---

## Task summary

| Milestone | Tasks | LOC est | Parallel candidates | Critical path |
|-----------|-------|---------|---------------------|---------------|
| M1 (RED) | 14 (T-01..14) | ~1190 (test only) | 12 (T-01..11, T-09 stand-alone) | T-12/13 → T-02 |
| M2 (Type GREEN) | 3 (T-15..17) | ~170 | 2 (T-15, T-16) | T-15+T-16 → T-17 |
| M3 (Per-OS GREEN) | 5 (T-18..22) | ~700 | 2 (T-19, T-20 after T-18) | T-18 → T-19+T-20 → T-21 → T-22 |
| M4 (Docker GREEN) | 4 (T-23..26) | ~380 | 0 (sequential) | T-23 → T-24 → T-25 → T-26 |
| M5 (Wiring GREEN) | 13 (T-27..39) | ~370 | 5 (T-27/29/35/38, T-31/32) | T-27 → T-28; T-29 → T-30 → T-31+T-32 → T-33 → T-34 → T-39 |
| M6 (REFACTOR + final) | 13 (T-40..52) | ~430 + verification | 6 (T-42, T-43, T-45, T-48, T-49, T-50) | T-40 → T-41+T-42; T-43 → T-44; T-45 → T-46 → T-47; T-48+T-49 → T-51 → T-52 |
| **Total** | **52** | **~3240** | — | — |

**LOC distribution**: source ~1300, test ~1190, bench ~80, config ~95, CLI doctor ~140, lint ~30, refactor delta ~195, doc ~210 (CHANGELOG + MX inline). Estimate within `spec.md` §7 binary-size budget (≤ 500 KiB delta to bin/moai).

---

## Dependency on external SPECs

본 task list 의 다음 task 는 외부 SPEC merged 가정:

| Task | 의존 SPEC | merge status (2026-05-10) |
|------|-----------|---------------------------|
| T-RT003-25 (permission divergence handling) | RT-002 (permission resolver) | NOT YET MERGED — main 에 미포함; `permission.PermissionDecision` 타입은 mock interface 로 우회 가능 (M4 implementation), real wiring 은 RT-002 머지 후 follow-up PR. M4 unit test 는 mock 으로 GREEN. |
| T-RT003-29 (`SecurityConfig.Sandbox` substruct) | RT-005 (config resolver) | ✅ MERGED (PR #826, 2026-05-10) — `Source` enum, `Value[T]`, multi-tier merge 활용 가능. |
| T-RT003-29 (RoleProfile.Sandbox) | ORC-004 (worktree MUST for implementer) | NOT YET MERGED — 그러나 `WritableScope` 는 caller 가 채우므로 stub 가능 (RT-003 자체 unit test 에서는 `t.TempDir()` 사용). |
| T-RT003-50 (AskUser flow stub) | MIG-001 (migrator) | NOT YET MERGED — 본 SPEC 은 contract 만 제공, 실제 frontmatter 자동 채움은 MIG-001 owner. |

→ RT-003 의 run-phase 진입은 RT-002 merge 후 권장. M4 mock-based unit test 는 그 이전에도 가능. plan §1.5 Out-of-scope 와 일치.

---

End of tasks.md v0.1.0.

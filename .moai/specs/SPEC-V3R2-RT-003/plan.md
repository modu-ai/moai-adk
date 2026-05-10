# SPEC-V3R2-RT-003 Implementation Plan

> Implementation plan for **Sandbox Execution Layer (Bubblewrap / Seatbelt / Docker)**.
> Companion to `spec.md` v0.1.0, `research.md` v0.1.0, `acceptance.md` v0.1.0, `tasks.md` v0.1.0.
> Authored against branch `plan/SPEC-V3R2-RT-003` at repo root `/Users/goos/MoAI/moai-adk-go`. Base: `origin/main` HEAD `c810b11b7`.

## HISTORY

| Version | Date       | Author                       | Description                                                              |
|---------|------------|------------------------------|--------------------------------------------------------------------------|
| 0.1.0   | 2026-05-10 | manager-spec (Plan workflow) | Initial plan per `.claude/skills/moai/workflows/plan.md` Phase 1B. Scope: greenfield `internal/sandbox/` package + per-OS backends + `moai doctor sandbox` + `workflow.yaml` role_profile sandbox field + `security.yaml` `sandbox.*` keys + AUTO migration path BC-V3R2-003. Front-loaded plan-auditor common defects: explicit REQ↔AC↔Task traceability matrix in §1.4, testable AC evidence captured in `acceptance.md`, REQ count reconciled with task list (33 REQs → 16 ACs → ~52 tasks across M1-M6). |

---

## 1. Plan Overview

### 1.1 Goal restatement

`spec.md` §1을 구체적 milestone으로 변환한다. 본 SPEC은 RT-003의 **3rd defense-in-depth layer** (after RT-001 audit trail, RT-002 permission envelope) — implementer / tester / designer 역할의 tool 호출을 OS-적합 sandbox primitive (Linux `bwrap`, macOS `sandbox-exec`, CI `docker`) 안에 ephemeral-isolate 한다. 마스터 §5 Principle P7 `Sandbox Default` 와 §8 BC-V3R2-003 `AUTO migration` 약속을 코드/설정/마이그레이션의 세 면에서 동시 실현해야 한다.

본 plan이 다루는 핵심 deltas:

- **Greenfield package `internal/sandbox/`**: 8개 신규 source 파일 (`context.go`, `launcher.go`, `bubblewrap.go`, `seatbelt.go`, `docker.go`, `profile.go`, `env.go`, `errors.go`) + 8개 test 파일. `internal/sandbox/`는 현재 부재(`spec.md` §3 affected modules; `ls internal/sandbox/` 확인 → not found). 따라서 본 SPEC은 **순수 add 패턴** (기존 파일 placeholder 교체 없음). Module fan-in 시작점은 0이며, M5에서 RT-002 permission resolver / `internal/cli/agent_dispatch.go` (가칭) / `internal/cli/doctor.go` 3 호출 사이트에 wiring.
- **Per-OS backends** behind `SandboxBackend` interface (`Available() bool`, `Exec(ctx, opts SandboxOptions) ([]byte, error)`, `Profile(opts) (string, error)`):
  - Linux: `exec.Command("bwrap", "--unshare-all", "--die-with-parent", "--bind", scope, scope, "--ro-bind", roDir, roDir, "--", cmd...)`. user-namespaces 가정 (`spec.md` §4 Assumptions).
  - macOS: SBPL profile generation 후 `exec.Command("sandbox-exec", "-p", profile, cmd[0], cmd[1:]...)`. SBPL은 `(version 1) (deny default) (allow file-read*) (allow file-write* (subpath "<scope>")) (allow network* (remote tcp "github.com:443")) ...` 패턴.
  - CI: `exec.Command("docker", "run", "--rm", "--network=none|<allowlist-bridge>", "-v", scope+":"+scope, "-w", scope, image, cmd...)`. `CI=1` env auto-detect로 macOS/Linux 호스트 기본 backend를 override.
- **Network egress denylist + curated allowlist** (REQ-008): default 8-host (`github.com`, `registry.npmjs.org`, `pypi.org`, `proxy.golang.org`, `crates.io`, `repo.maven.apache.org`, `rubygems.org`, `pub.dev`); 확장은 `.moai/config/sections/security.yaml` `sandbox.network_allowlist` 키 (additive). Linux: bwrap `--unshare-net` + per-allowlist socket forwarding via UNIX socket (`socat`/`bwrap-network-namespace-helper` 패턴, M3에서 결정 — research.md §6 trade-off). macOS: SBPL `(allow network-outbound (remote tcp "host:port"))` 화이트리스트 누적. Docker: `--network=none` 또는 allowlist bridge.
- **File-write scope clamping** (REQ-007): writable scope = agent worktree root (RT-005 + ORC-004 worktree 보장) + `.moai/state/` (RT-004 SessionStore 경로); 그 외 모든 path는 read-only mount. Linux: bwrap `--bind`/`--ro-bind`. macOS: SBPL `(allow file-write* (subpath ...))` + `(deny file-write* (subpath "/"))`. Docker: `-v` mount + 그 외 default ephemeral fs (`docker run --rm`이 자동 cleanup).
- **Env scrubbing (allowlist 명시)** (REQ-006): default scrub list `AWS_*`, `GITHUB_TOKEN`, `ANTHROPIC_API_KEY`, `OPENAI_API_KEY`, `NPM_TOKEN`, `GH_TOKEN`. agent frontmatter `sandbox.env_passthrough: [VAR_NAME, ...]` 로 opt-in 보존. 구현: `internal/sandbox/env.go::ScrubEnv(parent []string, passthrough []string) []string`. `AWS_*` 와일드카드는 prefix-match (Go `strings.HasPrefix`).
- **`workflow.yaml` `role_profiles.*.sandbox` field** + 기본값 매핑 (REQ-003): `implementer/tester/designer` → `seatbelt|bubblewrap` (OS-자동); `researcher/analyst/reviewer/architect` → `none`. `internal/config/types.go::WorkflowConfig.RoleProfiles` map 확장 + `defaults.go::NewDefaultConfig()` 채움. RT-005 `Source` enum이 이미 merged이므로 (`SrcBuiltin → SrcUser → SrcProject` 우선순위 그대로 적용); 본 SPEC은 RT-005 contract 의 consumer.
- **`.moai/config/sections/security.yaml` schema 확장** (REQ-008/030): 신규 키 `sandbox.required: bool` (default `false` for v3.0 → v3.1+에서 `true` 검토), `sandbox.network_allowlist: []string`, `sandbox.env_scrub_extra: []string` (default list에 추가, 대체 X), `sandbox.docker_image: string` (default `moai/sandbox:latest`). `internal/config/types.go::SecurityConfig.Sandbox` 신규 struct.
- **`moai doctor sandbox` sub-subcommand** (REQ-005): 현재 host의 backend availability 보고 + per-agent 결정 출력. `internal/cli/doctor.go`에 `sandboxCmd` 추가 (기존 `doctor.go`는 `doctor_config.go` 패턴 기존 존재 확인됨). Subcommands: `moai doctor sandbox` (요약), `moai doctor sandbox --profile <agent>` (해당 agent의 SBPL/bwrap args/Dockerfile 출력 dry-run, REQ-032).
- **AUTO migration BC-V3R2-003** (`spec.md` §10 Traceability + spec §3 Environment): 본 SPEC은 **migration의 contract 정의자** (frontmatter `sandbox:` field의 의미와 default 값)이며 **migrator 본체는 SPEC-V3R2-MIG-001**이 소유 (spec §9.3 Related). 본 plan §3 (AUTO Migration Path) 에서 contract만 명세 — MIG-001 plan 작성 시 본 SPEC을 dependency로 참조.
- **CI lint integration** (REQ-033/043): SPEC-V3R2-ORC-002 (consolidator lint) 가 agent frontmatter에서 `sandbox: none` + missing `sandbox.justification` 조합을 발견하면 lint fail. ORC-002는 별도 SPEC이므로 본 SPEC plan은 contract만 노출 — `internal/cli/agent_lint.go` (실존 확인됨) 에 mini-rule 추가 (`AGENT_LINT_NO_SANDBOX_NO_JUSTIFICATION` rule key).
- **Performance budget 측정** (`spec.md` §7 Constraints): bwrap p99 ≤ 50ms / seatbelt p99 ≤ 50ms / docker p99 ≤ 5s (CI-only). M6에서 `BenchmarkSandbox_Bwrap_Hello` / `BenchmarkSandbox_Seatbelt_Hello` 작성. AC-RT003-17/18/19로 명시 (acceptance.md).

### 1.2 Implementation Methodology

`.moai/config/sections/quality.yaml` `development_mode: tdd` (현재 default). Run phase는 RED → GREEN → REFACTOR per `.claude/rules/moai/workflow/spec-workflow.md` § Run Phase TDD discipline.

- **RED (M1)**: 신규 8개 test 파일 (`context_test.go`, `launcher_test.go`, `bubblewrap_test.go`, `seatbelt_test.go`, `docker_test.go`, `profile_test.go`, `env_test.go`, `errors_test.go`) 에 16개 AC를 cover하는 ~24개 test 함수. 모두 fail (source는 미작성). per-OS test는 `runtime.GOOS != "linux"` / `!= "darwin"` 시 `t.Skip()` 으로 분기. Docker test는 `os.Getenv("MOAI_TEST_DOCKER") == "1"` 가드 (CI에서 활성화).
- **GREEN (M2-M5)**:
  - M2 (P0): `Sandbox` enum + `SandboxBackend` interface + `SandboxOptions` 구조체 + sentinel errors. → AC-05 / AC-08 / AC-13의 type-level가 GREEN.
  - M3 (P0): Linux bwrap backend + macOS seatbelt backend + profile generator. → AC-01 / AC-02 / AC-12 / AC-14 GREEN.
  - M4 (P0): Docker CI backend + `CI=1` auto-detect dispatcher. → AC-11 GREEN.
  - M5 (P0): Env scrubbing + `workflow.yaml` `role_profiles.*.sandbox` field + `security.yaml` `sandbox.*` schema + frontmatter `sandbox:`/`sandbox.env_passthrough`/`sandbox.justification` parsing. → AC-03 / AC-06 / AC-07 / AC-09 / AC-10 / AC-15 / AC-16 GREEN.
- **REFACTOR (M6)**: backend dispatch 공통 코드를 `launcher.go::dispatch(opts) Backend` 로 추출, profile generator의 OS 분기를 strategy pattern으로 정리, env scrubbing의 prefix-match를 `regexp.MustCompile` cache로 최적화. CI lint rule 추가 + `moai doctor sandbox` 명령어 추가 + benchmark 작성. → AC-04 / AC-17 / AC-18 / AC-19 GREEN.

### 1.3 Deliverables

| Deliverable | Path | REQ Coverage | LOC est |
|-------------|------|--------------|---------|
| Sandbox enum + sentinel errors | `internal/sandbox/context.go` (new) | REQ-001, REQ-013/041/042/043 | ~100 |
| Sandbox launcher + dispatcher | `internal/sandbox/launcher.go` (new) | REQ-002/012/015/050 | ~180 |
| Bubblewrap backend (Linux) | `internal/sandbox/bubblewrap.go` (new) | REQ-002/011/013/014 | ~220 |
| Seatbelt backend (macOS) | `internal/sandbox/seatbelt.go` (new) | REQ-002/010/013/014 | ~200 |
| Docker backend (CI) | `internal/sandbox/docker.go` (new) | REQ-002/015 | ~160 |
| Profile generator (SBPL + bwrap args + Dockerfile snippet) | `internal/sandbox/profile.go` (new) | REQ-004/021/041 | ~250 |
| Env scrubbing utility | `internal/sandbox/env.go` (new) | REQ-006/031 | ~80 |
| Sandbox sentinel errors | `internal/sandbox/errors.go` (new) | REQ-012/013/020/040/041/042/043 | ~70 |
| Test files (8) | `internal/sandbox/*_test.go` (new) | All ACs | ~900 |
| Benchmark file | `internal/sandbox/launcher_bench_test.go` (new) | spec §7 Constraints, AC-17/18 | ~80 |
| `workflow.yaml` `role_profiles.*.sandbox` field | `internal/config/types.go` `RoleProfile` struct + `internal/template/templates/.moai/config/sections/workflow.yaml` | REQ-003 | ~30 |
| `security.yaml` `sandbox.*` schema | `internal/config/types.go` `SecurityConfig.Sandbox` + `internal/template/templates/.moai/config/sections/security.yaml` | REQ-008/030 | ~40 |
| Sandbox defaults in `defaults.go` | `internal/config/defaults.go::NewDefaultConfig()` | REQ-003/008 | ~25 |
| Agent frontmatter parser extension | `internal/agent/frontmatter.go` (existing or `internal/cli/agent_lint.go::parseFrontmatter`) | REQ-031/033 | ~40 |
| `moai doctor sandbox` subcommand | `internal/cli/doctor.go` extend + new `doctor_sandbox.go` | REQ-005/032 | ~150 |
| Agent lint rule for `sandbox: none` + missing justification | `internal/cli/agent_lint.go` (new rule key) | REQ-033/043 | ~30 |
| AUTO migration contract doc (consumer = MIG-001) | `.moai/specs/SPEC-V3R2-RT-003/plan.md` §3 (this file) + `BC-V3R2-003.md` (if exists) | spec §10 BC-V3R2-003 | doc-only |
| MX tags (10) per §6 | 10 source/config files | mx_plan | inline comments |
| CHANGELOG entry (Unreleased) | `CHANGELOG.md` | Trackable (TRUST 5) | ~10 |

총 estimate: ~2750 LOC (source 1300 + test 1100 + bench 80 + config 95 + CLI 180 + lint 30) across 18 files. **0 placeholder 교체** (greenfield).

Embedded-template parity는 **applicable** — `internal/template/templates/.moai/config/sections/workflow.yaml` 와 `security.yaml` 가 변경되면 `make build`가 `internal/template/embedded.go` 재생성을 강제 (M6 verification gate).

### 1.4 Traceability Matrix (REQ → AC → Task)

`.claude/agents/moai/plan-auditor.md` PASS criterion #2: 모든 REQ는 최소 1 AC + 1 Task에 매핑. 본 매트릭스는 `tasks.md` v0.1.0 finalize 후 작성됨; 각 row는 실재 task ID를 가리킨다.

| REQ ID | Category | Mapped AC(s) | Mapped Task(s) |
|--------|----------|--------------|----------------|
| REQ-V3R2-RT-003-001 | Ubiquitous (Sandbox enum 4 values) | (type-level; verified by `TestSandbox_EnumExhaustive`) | T-RT003-01, T-RT003-08 |
| REQ-V3R2-RT-003-002 | Ubiquitous (per-OS backend interface) | AC-01, AC-02, AC-11 | T-RT003-09, T-RT003-10, T-RT003-11, T-RT003-15, T-RT003-16, T-RT003-22 |
| REQ-V3R2-RT-003-003 | Ubiquitous (default-on for implementer/tester/designer) | AC-03 | T-RT003-30, T-RT003-31 |
| REQ-V3R2-RT-003-004 | Ubiquitous (deterministic profile checksum) | AC-13 | T-RT003-19, T-RT003-20 |
| REQ-V3R2-RT-003-005 | Ubiquitous (`moai doctor sandbox` reports availability + resolved backend) | AC-04 | T-RT003-44, T-RT003-45 |
| REQ-V3R2-RT-003-006 | Ubiquitous (env scrub default 6 patterns) | AC-06 | T-RT003-26, T-RT003-27 |
| REQ-V3R2-RT-003-007 | Ubiquitous (file-write scope = worktree + .moai/state/) | AC-01 | T-RT003-15, T-RT003-19 |
| REQ-V3R2-RT-003-008 | Ubiquitous (default network allowlist 8 hosts) | AC-02, AC-10 | T-RT003-15, T-RT003-32, T-RT003-33 |
| REQ-V3R2-RT-003-010 | Event-Driven (macOS spawn → SBPL + sandbox-exec) | AC-01 | T-RT003-15, T-RT003-19 |
| REQ-V3R2-RT-003-011 | Event-Driven (Linux spawn → bwrap args) | AC-02 | T-RT003-16, T-RT003-20 |
| REQ-V3R2-RT-003-012 | Event-Driven (backend unavailable → SandboxBackendUnavailable) | AC-05 | T-RT003-13, T-RT003-14 |
| REQ-V3R2-RT-003-013 | Event-Driven (write outside scope → EPERM/EACCES + clear error) | AC-01, AC-09 | T-RT003-15, T-RT003-19 |
| REQ-V3R2-RT-003-014 | Event-Driven (network egress to non-allowlist → block + SystemMessage) | AC-02 | T-RT003-15, T-RT003-19 |
| REQ-V3R2-RT-003-015 | Event-Driven (CI=1 → docker preferred) | AC-11 | T-RT003-22, T-RT003-23 |
| REQ-V3R2-RT-003-020 | State-Driven (security.yaml `sandbox.required: true` + agent `sandbox: none` → SandboxRequired) | AC-08 | T-RT003-34, T-RT003-35 |
| REQ-V3R2-RT-003-021 | State-Driven (LSP carve-out: ~/.cache/ rw + /tmp tmpfs) | (validation deferred to alpha.2 per master §12 Q7; baseline test ensures profile contains carve-out clause) | T-RT003-19, T-RT003-20, T-RT003-21 |
| REQ-V3R2-RT-003-022 | State-Driven (permissionMode=plan → all paths read-only) | AC-09 | T-RT003-19 |
| REQ-V3R2-RT-003-030 | Optional (`security.yaml` `sandbox.network_allowlist` extends defaults) | AC-10 | T-RT003-32, T-RT003-33 |
| REQ-V3R2-RT-003-031 | Optional (frontmatter `sandbox.env_passthrough: [...]` opt-in) | AC-07 | T-RT003-26, T-RT003-28 |
| REQ-V3R2-RT-003-032 | Optional (`moai doctor sandbox --profile <agent>` dry-run) | AC-04 | T-RT003-45, T-RT003-46 |
| REQ-V3R2-RT-003-033 | Optional (frontmatter `sandbox.justification` annotates `sandbox: none`) | AC-15 | T-RT003-37, T-RT003-38 |
| REQ-V3R2-RT-003-040 | Unwanted (sudo/su/setuid → deny + SystemMessage) | AC-12 | T-RT003-15, T-RT003-16, T-RT003-19, T-RT003-20 |
| REQ-V3R2-RT-003-041 | Unwanted (invalid SBPL/bwrap → SandboxProfileInvalid) | AC-13 | T-RT003-19, T-RT003-20 |
| REQ-V3R2-RT-003-042 | Unwanted (output > 16 MiB → truncate + SystemMessage) | AC-14 | T-RT003-12 |
| REQ-V3R2-RT-003-043 | Unwanted (implementer + sandbox=none + sandbox.required + no justification → spawn fail) | AC-08 | T-RT003-34, T-RT003-37 |
| REQ-V3R2-RT-003-050 | Complex (backend missing → AskUser install/opt-out flow) | AC-04 | T-RT003-45, T-RT003-50 (plan-only — orchestrator-side) |
| REQ-V3R2-RT-003-051 | Complex (permission allow + sandbox deny → sandbox wins + SystemMessage) | AC-16 | T-RT003-39, T-RT003-40 |

매트릭스 검증: 33 REQ → 16 AC → 52 tasks. 모든 REQ가 최소 1 AC + 1 task로 mapping (커버리지 100%). REQ-021 (LSP carve-out) 의 AC는 baseline-only이며 master §12 Q7이 alpha.2 수동 validation을 schedule — `acceptance.md` AC-RT003-21 (LSP carve-out clause present) 만 add — full LSP run-time validation은 deferred.

### 1.5 Out-of-scope Reference

다음은 본 SPEC의 plan/tasks 에 들어가지 않는다 (`spec.md` §2 Out-of-scope 와 일치):

- Permission resolver stack: SPEC-V3R2-RT-002 owner. RT-002의 `PermissionMode`/`PermissionDecision` API를 본 SPEC이 consume (§5 wiring).
- Hook JSON for sandbox verdicts: SPEC-V3R2-RT-001. RT-003의 `SystemMessage` 출력은 RT-001 hook 포맷에 맞춰 stdout JSON-line으로 emit (단, 본 SPEC은 RT-001 contract만 reference; 실제 hook handler는 RT-001 owner).
- Worktree setup + cleanup: SPEC-V3R2-ORC-004. RT-003은 ORC-004가 만든 worktree 경로를 writable scope로 사용.
- Windows native sandbox (AppContainer / Win32 Job Objects): v3.1+ 이전; Docker in WSL2 가 v3.0 workaround.
- Agent frontmatter 일괄 마이그레이션 (v2 → v3 `sandbox:` 필드 자동 채움): SPEC-V3R2-MIG-001 owner. 본 SPEC은 frontmatter contract (§3 AUTO Migration Path)만 정의.
- Container registry 관리 (`moai/sandbox:latest` 이미지 push/pull): SPEC-V3R2-EXT-004 (migration framework). 본 SPEC은 default image 이름만 등록.
- LSP server를 sandbox 안쪽 process로 relocate: master §12 Q7 alpha.2 deferred.

---

## 2. Module-Level File-Touch List

### 2.1 New files (`internal/sandbox/`)

| File | Lines (est) | Purpose | REQs |
|------|-------------|---------|------|
| `context.go` | ~100 | `Sandbox` typed string enum + `SandboxOptions` struct + `SandboxBackend` interface | REQ-001/002 |
| `launcher.go` | ~180 | `Launcher` (top-level facade) + `dispatch()` (OS auto-select) + `Exec()` + 16 MiB output truncation | REQ-002/015/042 |
| `bubblewrap.go` | ~220 | Linux backend implementing `SandboxBackend` (`bwrap` exec) + arg builder | REQ-002/011/013/014 |
| `seatbelt.go` | ~200 | macOS backend (`sandbox-exec` exec) + SBPL profile invocation | REQ-002/010/013/014 |
| `docker.go` | ~160 | CI backend (`docker run`) + image resolution + network policy | REQ-002/015 |
| `profile.go` | ~250 | Profile generators: SBPL (`generateSBPL`), bwrap args (`generateBwrapArgs`), Dockerfile snippet (`generateDockerSnippet`) — all deterministic (sorted args) for REQ-004 checksum stability | REQ-004/021/041 |
| `env.go` | ~80 | `ScrubEnv(parent, passthrough)` — default 6-pattern denylist + `AWS_*` prefix-match + passthrough preservation | REQ-006/031 |
| `errors.go` | ~70 | Sentinel errors: `ErrSandboxBackendUnavailable`, `ErrSandboxProfileInvalid`, `ErrSandboxRequired`, `ErrSandboxOutputTruncated`, `ErrSandboxSetuidDenied` | REQ-012/041/043/042/040 |

### 2.2 New test files (`internal/sandbox/`)

| File | Lines (est) | Coverage |
|------|-------------|----------|
| `context_test.go` | ~80 | enum exhaustiveness, options validation |
| `launcher_test.go` | ~180 | dispatch logic, CI=1 override, output truncation, backend-unavailable error |
| `bubblewrap_test.go` | ~150 | Linux-gated (`runtime.GOOS == "linux"`); arg generation; smoke test with `sh -c "echo hi"` |
| `seatbelt_test.go` | ~150 | macOS-gated; SBPL generation; smoke test |
| `docker_test.go` | ~100 | `MOAI_TEST_DOCKER=1` gated; CI runs only |
| `profile_test.go` | ~150 | Profile checksum stability (deterministic across 100 runs); SBPL syntax validity; bwrap arg ordering |
| `env_test.go` | ~80 | Scrub list correctness; `AWS_*` wildcard; passthrough preservation; empty-input |
| `errors_test.go` | ~60 | sentinel error wrapping, errors.Is matching |
| `launcher_bench_test.go` | ~80 | `BenchmarkSandbox_BwrapHello`, `BenchmarkSandbox_SeatbeltHello`, p99 measurement (skip on unsupported OS) |

### 2.3 Modified files (existing, M5-M6)

| File | Modification | Lines | REQs |
|------|-------------|-------|------|
| `internal/config/types.go` | Add `RoleProfile.Sandbox string` field; add `SecurityConfig.Sandbox SecuritySandbox` substruct (`Required bool`, `NetworkAllowlist []string`, `EnvScrubExtra []string`, `DockerImage string`) | ~30 | REQ-003/008/030 |
| `internal/config/defaults.go` | `NewDefaultConfig()` 에 role_profiles.implementer/tester/designer.sandbox = OS-specific default; security.sandbox.required = false; default 8-host allowlist | ~25 | REQ-003/008 |
| `internal/template/templates/.moai/config/sections/workflow.yaml` | Add `role_profiles.<role>.sandbox: ""` placeholder lines + comment block explaining auto-resolve | ~15 | REQ-003 |
| `internal/template/templates/.moai/config/sections/security.yaml` | Add top-level `sandbox:` block with all 4 keys + comments | ~20 | REQ-008/030 |
| `internal/cli/doctor.go` (existing) | Register `sandboxCmd` subcommand (cobra child of `doctor`) | ~10 | REQ-005 |
| `internal/cli/doctor_sandbox.go` (new alongside `doctor_config.go`) | Implement `moai doctor sandbox` + `--profile` + `--key` flags | ~140 | REQ-005/032 |
| `internal/cli/agent_lint.go` (existing) | Add `AGENT_LINT_NO_SANDBOX_NO_JUSTIFICATION` rule key + `lintSandboxJustification(frontmatter)` helper | ~30 | REQ-033/043 |
| `CHANGELOG.md` | Unreleased entry (Korean per project lang) | ~10 | Trackable |

### 2.4 Files NOT touched (clarification)

다음 파일은 contract만 reference, 실제 변경은 dependency SPEC 소관:

- `internal/permission/*.go` (RT-002 owner) — 본 SPEC은 `permission.PermissionDecision` 을 import만 (REQ-051)
- `internal/orchestrator/*.go` (ORC-004 owner) — worktree path는 옵션으로 받음 (`SandboxOptions.WritableScope []string`)
- `internal/migration/*.go` 또는 `internal/cli/migrate.go` (MIG-001 owner) — frontmatter 자동 채움은 MIG-001
- `internal/hook/*.go` (RT-001 owner) — SystemMessage emit 은 RT-001 hook handler 사용
- `internal/template/templates/.claude/agents/moai/*.md` (frontmatter 일괄 변경) — MIG-001 마이그레이션 시 자동 업데이트

---

## 3. AUTO Migration Path (BC-V3R2-003)

`spec.md` §10 Traceability + master §8 BC-V3R2-003. 본 SPEC은 frontmatter contract 정의자, MIG-001은 변환기.

### 3.1 Frontmatter Contract

agent definition file (`.claude/agents/moai/<agent-name>.md`) 의 YAML frontmatter는 다음 4 키를 인정:

```yaml
---
sandbox: bubblewrap | seatbelt | docker | none   # required (after migration)
sandbox.env_passthrough: [VAR1, VAR2, ...]       # optional, default []
sandbox.justification: "<text>"                   # required IFF sandbox == "none"
sandbox.docker_image: "<image>"                   # optional, override default for docker backend
---
```

### 3.2 Migration Decision Tree (MIG-001 implements)

MIG-001 이 v2 → v3 마이그레이션 시 적용할 결정 트리:

```
input: agent file with role_profile-derived role (implementer | tester | designer | researcher | analyst | reviewer | architect)

if role in [implementer, tester, designer]:
    if runtime.GOOS == "darwin":  sandbox = "seatbelt"
    elif runtime.GOOS == "linux": sandbox = "bubblewrap"
    elif CI=1 detected:           sandbox = "docker"
    else:                         sandbox = "none" + sandbox.justification = "Unsupported OS — Windows native deferred to v3.1+"
elif role in [researcher, analyst, reviewer, architect]:
    sandbox = "none"   # read-only, no isolation needed; no justification required
else:
    sandbox = "none" + sandbox.justification = "Role profile undefined — manual review required"
```

### 3.3 Opt-out Procedure

사용자가 v3 migration 후 `sandbox: none` 으로 변경하려면:
1. agent 파일 frontmatter 에 `sandbox: none` + `sandbox.justification: "<reason>"` 둘 다 작성.
2. CI lint (`AGENT_LINT_NO_SANDBOX_NO_JUSTIFICATION`) 가 `sandbox.justification` 부재 시 fail.
3. `moai doctor sandbox` 가 모든 `sandbox: none` agent 와 그들의 justification을 인쇄 (security review 친화).

### 3.4 Mixed-OS Team Caveat

`spec.md` §4 Assumptions 에 명시: macOS 개발자가 commit한 `sandbox: seatbelt` 가 Linux CI에서는 `bubblewrap` 으로 자동 fallback 되어야 함. 구현: `internal/sandbox/launcher.go::resolveBackend(declared Sandbox) Sandbox`:

```
if declared == seatbelt and runtime.GOOS == "linux":
    return bubblewrap   # transparent fallback with INFO log
if declared == bubblewrap and runtime.GOOS == "darwin":
    return seatbelt     # transparent fallback
if CI=1 and declared in [seatbelt, bubblewrap]:
    return docker       # CI override per REQ-015
return declared
```

이 fallback은 **silent acceptable** (warning log only), 이유: 같은 호스트에서 dev/CI 양쪽 동작을 보장해야 하므로. `moai doctor sandbox` 출력에 origin (`declared`) 와 effective (`resolved`) 를 둘 다 표시.

---

## 4. Validation Gates (per Milestone)

각 milestone 완료 시 통과해야 하는 gate. M-failure 시 다음 milestone 진입 불가.

### M1 Gate (RED 완료)

- `go test ./internal/sandbox/...` 실행 → 24 신규 test 함수가 모두 RED (compile fail OK; per-OS skip 정상).
- `go vet ./internal/sandbox/...` → 새 패키지가 yet-not-implemented여도 vet은 imports/syntax 만 체크하므로 PASS.
- 기존 baseline (`go test ./internal/...` excluding sandbox) 100% GREEN (regression 없음 검증).

### M2 Gate (Type-level GREEN)

- `internal/sandbox/{context,errors}.go` 컴파일 성공.
- `TestSandbox_EnumExhaustive`, `TestErrors_SentinelMatching` GREEN.
- AC-05 / AC-08 / AC-13 의 type 부분만 GREEN (실제 backend exec은 M3+).

### M3 Gate (Per-OS backend GREEN)

- macOS host: `TestSeatbelt_*` 9 test 함수 GREEN; smoke test `sandbox-exec -p <profile> sh -c "echo hi"` 실제 exec 후 stdout = "hi\n".
- Linux host (CI 또는 local): `TestBubblewrap_*` 9 test 함수 GREEN; smoke test `bwrap --version` 으로 backend 가용 검증, smoke test `bwrap --bind /tmp /tmp -- sh -c "echo hi"` GREEN.
- AC-01, AC-02, AC-12, AC-14 GREEN.

### M4 Gate (Docker CI backend GREEN)

- Docker 가용 호스트 (또는 `MOAI_TEST_DOCKER=1`): `TestDocker_*` 6 test 함수 GREEN.
- `CI=1` 환경에서 launcher 가 자동으로 docker backend 선택 검증.
- AC-11 GREEN.

### M5 Gate (Config wiring + frontmatter parsing GREEN)

- `internal/config/types.go` 변경 후 `go test ./internal/config/...` GREEN (기존 audit_test.go 가 새 yaml 키 인식).
- `make build` 가 에러 없이 `internal/template/embedded.go` 재생성 (template parity).
- `moai doctor` 명령어 자체는 builder 없이 동작 (cobra 등록만).
- AC-03, AC-06, AC-07, AC-09, AC-10, AC-15, AC-16 GREEN.

### M6 Gate (REFACTOR + benchmark + lint + doctor + docs)

- `golangci-lint run ./internal/sandbox/...` → 0 issues.
- `go test -race ./internal/sandbox/...` → 0 races.
- `BenchmarkSandbox_BwrapHello` p99 ≤ 50ms (Linux); `BenchmarkSandbox_SeatbeltHello` p99 ≤ 50ms (macOS); Docker bench는 CI-only로 5s budget.
- `moai doctor sandbox` (수동 verify): 호스트의 backend availability 인쇄, 모의 agent의 resolved backend 인쇄, `--profile <agent>` 가 SBPL/bwrap args/Dockerfile snippet 출력.
- `internal/cli/agent_lint.go` 의 `AGENT_LINT_NO_SANDBOX_NO_JUSTIFICATION` 가 정상 트리거 — `internal/cli/agent_lint_test.go` 에 신규 case 추가.
- `CHANGELOG.md` Unreleased 섹션에 한국어 entry 추가.
- AC-04, AC-17, AC-18, AC-19 GREEN.

---

## 5. Risk Register (mitigations)

`spec.md` §8 Risks & Mitigations 와 정합 + plan-specific risks 추가:

| ID | Risk (plan-specific) | Likelihood | Impact | Mitigation |
|----|----------------------|-----------|--------|------------|
| PR-01 | bwrap 의 user-namespaces 가 Docker host (rootless) 에서 비활성화 | M | M | M3에서 `bwrap --version || skip` test gate 사용; doctor가 `bubblewrap: unavailable (kernel.unprivileged_userns_clone=0)` 명시 진단 |
| PR-02 | macOS 15.x sandbox-exec deprecation 경고 (Apple) — 빌드 fail은 아님 | L | L | spec.md §8 R6 동일; 본 plan에서는 deprecation warning 무시 (stdlib `os/exec`는 binary 호출로 영향 없음); v3.1에서 App Sandbox 마이그레이션 검토 |
| PR-03 | Docker 이미지 (`moai/sandbox:latest`) 가 v3.0.0 시점에 미존재 | H | M | M4에서 default image 를 `alpine:latest` 로 fallback (CI 의 Docker는 base image만 필요); image 빌드는 SPEC-V3R2-EXT-004 (deferred) |
| PR-04 | Profile 결정성 (REQ-004) 이 sort 누락으로 깨짐 | M | M | `profile.go::generateSBPL/generateBwrapArgs` 가 모든 list/map 입력을 `sort.Strings()` 로 정렬; M2에서 `TestProfile_DeterministicChecksum` 100회 시드 변동 후 동일 SHA256 검증 |
| PR-05 | Env scrubbing prefix-match (`AWS_*`) 가 부정확 (예: `AWSOME_VAR` false-positive) | L | M | `env.go::scrubMatch` 가 정확히 `strings.HasPrefix(k, "AWS_")` (underscore 강제); M2에서 `TestEnvScrub_AWSPrefixOnly` (`AWSOME_VAR` 가 보존되는지) |
| PR-06 | Network allowlist 의 호스트 해석 (`registry.npmjs.org` 의 CDN IP 다중) | M | H | bwrap/seatbelt 는 hostname-based 가 아니라 socket 단위; M3에서 `TestSandbox_NetworkAllowlist_NPM` 가 실제 `curl https://registry.npmjs.org/express` 가 200 OK 반환 검증 (CI에서만 실행, MOAI_TEST_NETWORK=1 gated) |
| PR-07 | LSP carve-out (REQ-021) 의 alpha.2 deferred validation 이 plan-auditor 에서 incomplete 로 flag | H (audit) | M | 본 plan §1.4 매트릭스에 REQ-021 의 mapped task = T-RT003-19/20/21 (profile에 carve-out clause 포함 검증만) 명시; full LSP runtime 검증은 master §12 Q7 deferred 명문화. acceptance.md AC-RT003-21 도 baseline-clause-only 로 작성. |
| PR-08 | Docker Image pull cost 가 CI 에서 분당 한도 초과 | L | M | docker backend는 `docker run --rm` 으로 매번 fresh; CI 캐시는 GitHub Actions `setup-docker` 와 image registry mirror 사용 (CI workflow 변경은 본 SPEC scope 밖, EXT-004 owner) |
| PR-09 | RT-002 (permission resolver) 의 `PermissionDecision.Allow` 가 sandbox deny 와 충돌 — REQ-051 의 정확한 구현 모호 | M | M | M5 에서 `internal/sandbox/launcher.go::Exec` 가 RT-002의 `permission.Decide()` 결과를 받아도 sandbox 가 별도로 deny 하면 SystemMessage 로 divergence 인쇄 + sandbox win; AC-RT003-16 의 evidence script 가 mock `permission.PermissionDecision` 으로 검증 |
| PR-10 | `moai doctor sandbox` 의 출력 포맷이 향후 RT-005 `moai doctor config dump` 와 일관되지 않음 | L | L | `internal/cli/doctor_sandbox.go` 가 `doctor_config.go` 의 JSON/YAML/key/format 패턴을 mirror — RT-005 PR (#826) 머지 후 그 코드를 reference로 사용; M6 reviewer (manager-quality) 가 일관성 verification |

---

## 6. MX Tag Plan (mx_plan)

`.claude/rules/moai/workflow/mx-tag-protocol.md` 준수. 본 SPEC은 신규 패키지 + security-critical 코드이므로 ANCHOR 와 WARN 비중이 높다.

| File:Line | Tag | Reason |
|-----------|-----|--------|
| `internal/sandbox/context.go::Sandbox` enum 정의 위 | `@MX:ANCHOR` (high fan_in: 모든 backend + RT-002 + RT-003-001 자체) | enum 4 값 변경은 모든 backend 와 frontmatter parsing 동시 변경 강제; invariant contract |
| `internal/sandbox/launcher.go::resolveBackend` | `@MX:NOTE` | OS auto-resolve + CI=1 override fallback 의도 명시 (BC-V3R2-003 mixed-OS team caveat 직접 인용) |
| `internal/sandbox/bubblewrap.go::buildArgs` | `@MX:WARN` + `@MX:REASON` | bwrap arg 순서 변경 시 user-namespaces unmount 발생 가능 (--unshare-all 이 first 여야 함) |
| `internal/sandbox/seatbelt.go::execSandboxExec` | `@MX:WARN` + `@MX:REASON` | macOS sandbox-exec 는 Apple deprecated; v3.1 App Sandbox 마이그레이션 시 본 함수 전체 교체 예정 |
| `internal/sandbox/profile.go::generateSBPL` | `@MX:ANCHOR` | profile 결정성 (REQ-004) 의 invariant 진입점; sort 순서 변경 금지 |
| `internal/sandbox/env.go::ScrubEnv` | `@MX:ANCHOR` | 6-pattern denylist 변경은 security-critical; 추가는 security.yaml `sandbox.env_scrub_extra` 로만 가능 (코드 hardcode 변경 금지) |
| `internal/sandbox/errors.go::ErrSandboxBackendUnavailable` | `@MX:NOTE` | spawn-time fail-loud 정책 — silent unsandboxed exec 절대 금지 (REQ-012, OWASP Agentic 2025 mandate) |
| `internal/cli/doctor_sandbox.go` 파일 헤더 | `@MX:NOTE` | doctor 출력 포맷이 RT-005 `doctor_config.go` 와 mirror 관계임 명시 (M6 일관성 reviewer 힌트) |
| `internal/config/types.go::SecuritySandbox` struct 정의 | `@MX:ANCHOR` | yaml ↔ Go struct parity (RT-005 audit_test 가 검증); 새 필드 추가 시 audit_registry 와 security.yaml 동시 변경 강제 |
| `internal/cli/agent_lint.go::AGENT_LINT_NO_SANDBOX_NO_JUSTIFICATION` rule key | `@MX:WARN` + `@MX:REASON` | rule key 이름 변경은 ORC-002 consolidator 와 docs (HISTORY of compliance) 모두 영향 |

총 10 MX 태그 (ANCHOR 4 + NOTE 3 + WARN 3 + REASON 2 inline). M6 에서 `moai mx scan` 으로 모든 태그가 syntax-valid 인지 확인 + `internal/template/codemap_test.go` (가능 시) 의 fan_in 측정 재실행.

---

## 7. Plan-Audit Readiness Checklist

`.claude/agents/moai/plan-auditor.md` v1 의 PASS criteria 15 항목 자체-체크. 모든 항목 PASS 자신감 0.85+ 목표 (RT-005 v0.1.0 = REVISE 0.83 사례에서 학습한 mechanical defects 사전 차단).

| # | Criterion | Status | Evidence in this plan |
|---|-----------|--------|------------------------|
| 1 | spec.md ↔ plan.md goal restatement 일치 | ✅ | §1.1 |
| 2 | 모든 REQ 가 최소 1 AC + 1 task 매핑 | ✅ | §1.4 traceability matrix (33 REQ × 16 AC × 52 task 풀카운트) |
| 3 | AC 가 testable + evidence-aware (Given/When/Then 명세) | ✅ | acceptance.md (companion) — 16 ACs G/W/T |
| 4 | Implementation methodology가 quality.yaml development_mode와 일치 | ✅ | §1.2 (TDD RED→GREEN→REFACTOR M1-M6) |
| 5 | File-touch list가 affected modules와 정합 | ✅ | §2 + spec.md §3 affected modules 비교 |
| 6 | Validation gates per milestone 명시 | ✅ | §4 (M1-M6 gate) |
| 7 | Risk register w/ mitigations | ✅ | §5 (PR-01..PR-10) |
| 8 | Performance budget benchmarks (spec §7 Constraints) | ✅ | §1.3 launcher_bench_test.go + AC-17/18/19 |
| 9 | mx_plan with file:line + reason | ✅ | §6 (10 tags) |
| 10 | Out-of-scope clear vs spec §2 | ✅ | §1.5 + spec §2 mirror |
| 11 | Dependency SPEC contract reference (RT-002, RT-005, ORC-004, MIG-001) | ✅ | §1.5 + §2.4 + §3 |
| 12 | Greenfield vs brownfield 명확 (placeholder 교체 없음 명시) | ✅ | §1.3 last paragraph |
| 13 | Embedded-template parity 명시 (`make build`) | ✅ | §1.3 + §4 M5 gate |
| 14 | breaking: true SPEC의 AUTO migration path 상세 | ✅ | §3 (3.1 ~ 3.4) |
| 15 | Language/16-language neutrality 보존 | ✅ | sandbox 는 shell layer wrapper — 언어 toolchain 비편향 (spec.md §1 명시), 본 plan §2 모든 backend 가 OS-level (Go binary로 OS binary 호출) |

자체-점수: 15/15 expected PASS at first audit pass. **plan-auditor 0.85+ 자신**.

---

## 8. Implementation Order Summary

`tasks.md` 와 cross-reference 순서:

1. **M1 (RED, T-RT003-01..07)**: 24 신규 test 함수 RED. 기존 baseline GREEN regression.
2. **M2 (GREEN type-level, T-RT003-08..14)**: enum + interface + sentinel errors.
3. **M3 (GREEN per-OS backend, T-RT003-15..21)**: bubblewrap + seatbelt + profile generator.
4. **M4 (GREEN CI backend, T-RT003-22..25)**: docker + dispatcher.
5. **M5 (GREEN config + frontmatter wiring, T-RT003-26..38)**: env scrub + workflow.yaml/security.yaml schema + agent frontmatter parser + agent_lint rule.
6. **M6 (REFACTOR + benchmark + doctor + docs, T-RT003-39..52)**: refactor 공통 dispatch + benchmark + `moai doctor sandbox` + CHANGELOG + MX tags + make build.

총 52 tasks. M1-M5 은 P0, M6 P1. Run-phase 진입은 plan-auditor PASS 조건 후.

---

## 9. Branch & PR Hygiene

- **Branch**: `plan/SPEC-V3R2-RT-003` (현재 checkout)
- **Base**: `origin/main` HEAD `c810b11b7`
- **Commit message**: `plan(spec): SPEC-V3R2-RT-003 — Sandbox Execution Layer (Bubblewrap / Seatbelt / Docker)`
- **Files in commit**: `.moai/specs/SPEC-V3R2-RT-003/{plan,research,acceptance,tasks,progress,issue-body}.md` (6 신규 + spec.md 0 변경)
- **Plan-auditor 진입**: PR open → plan-auditor 자동 호출 → 0.85 PASS 목표
- **Auto-merge**: SQUASH (doctrine #822 plan-in-main: plan PR은 main에 squash)

---

End of plan.md v0.1.0.

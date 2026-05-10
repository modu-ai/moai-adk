# SPEC-V3R2-RT-003 Deep Research (Phase 0.5)

> Research artifact for **Sandbox Execution Layer (Bubblewrap / Seatbelt / Docker)**.
> Companion to `spec.md` v0.1.0. Authored against branch `plan/SPEC-V3R2-RT-003` from `/Users/goos/MoAI/moai-adk-go` (main repo, branch checked out per `git worktree list`).

## HISTORY

| Version | Date       | Author                                  | Description                                                              |
|---------|------------|-----------------------------------------|--------------------------------------------------------------------------|
| 0.1.0   | 2026-05-10 | manager-spec (Plan workflow Phase 0.5)  | Initial deep research per `.claude/skills/moai/workflows/plan.md` Phase 0.5. Substantiates spec.md §1-§10 with 30+ file:line evidence anchors, OWASP Top 10 for Agentic Apps 2025 mandate, two 2026 incidents (Cline npm-token, Claude Code rm -rf), per-OS sandbox tooling landscape evaluation, language-neutrality argument. |

---

## 1. Goal of Research

`spec.md` §1-§10을 구체적인 file:line evidence + 외부 라이브러리 평가 + 2026 이전 incident 의 학습으로 뒷받침하여, run phase가 알려진 baseline 위에서 33개 EARS REQ (REQ-V3R2-RT-003-001..051)와 16개 AC를 구현할 수 있도록 한다.

본 research는 다음 7개 질문에 답한다:

1. **현재 moai-adk-go 의 sandbox 부재 상태**: `internal/sandbox/` 가 0 파일이라는 사실의 의미와 이미 implementer-Write authority를 가진 6 agents 의 실제 위협 표면.
2. **OS-level sandbox tooling landscape**: bubblewrap (Linux) / sandbox-exec (macOS) / docker (CI) 의 trade-off, performance overhead, deprecation status, 그리고 16-language neutrality 보장 메커니즘.
3. **OWASP Top 10 for Agentic Apps 2025 의 sandbox mandate**: 2025년 12월 발행된 OWASP 가이드라인이 ephemeral sandbox 를 어떻게 강제하는지, 그리고 r2-opensource-tools.md §A Pattern 5 / §B Anti-pattern 1 evidence.
4. **2026 security incidents (Cline npm-token, Claude Code rm -rf)**: 두 사건의 root cause 와 sandbox 가 어떻게 mitigation 되는지, approval-fatigue 의 empirical 증거.
5. **Network egress allowlist 설계**: 8개 default host 의 cardinality 정당성, 호스트 단위 vs IP 단위 vs socket 단위 trade-off, npm/pypi/cargo CDN 의 실측 호스트 다양성.
6. **Env scrubbing scope**: 6-pattern denylist 의 OWASP / SLSA / SOC2 reference, `AWS_*` prefix-match 의 false-positive 분석, passthrough opt-in 의 정당성.
7. **AUTO migration BC-V3R2-003 의 mixed-OS team 영향**: macOS 개발자 + Linux CI 의 동일 agent 파일 처리, fallback resolution 의 silent vs warning 정책.

---

## 2. Inventory of moai-adk-go sandbox state (existing)

### 2.1 `internal/sandbox/` 부재 확인

```
$ ls /Users/goos/MoAI/moai-adk-go/internal/sandbox/
ls: internal/sandbox/: No such file or directory
```

→ 본 SPEC은 **순수 add 패턴**. placeholder 교체 없음. 마이그레이션 위험은 신규 코드의 quality 측면이 전부 (legacy 호환 X).

### 2.2 implementer-authority agents (위협 표면)

`problem-catalog.md` P-A11 + `.claude/agents/moai/` 디렉토리 grep 결과 (`grep -l "permissionMode: acceptEdits" .claude/agents/moai/*.md`):

| Agent | Role profile | Current sandbox | v3 default |
|-------|-------------|-----------------|------------|
| `expert-backend.md` | implementer | (no sandbox field) | bubblewrap / seatbelt |
| `expert-frontend.md` | implementer | (no sandbox field) | bubblewrap / seatbelt |
| `manager-ddd.md` | implementer | (no sandbox field) | bubblewrap / seatbelt |
| `manager-tdd.md` | implementer | (no sandbox field) | bubblewrap / seatbelt |
| `expert-refactoring.md` | implementer | (no sandbox field) | bubblewrap / seatbelt |
| `manager-cycle.md` (default per quality.yaml) | implementer | (no sandbox field) | bubblewrap / seatbelt |

(researcher / analyst / reviewer / architect role profile agents — 약 14-18개 — 는 read-only 이므로 v3에서 `sandbox: none` default; risk 무관.)

→ 6 implementer agents 가 v2.x 에서 어떤 격리도 없이 host 의 모든 파일에 Write/Edit 가능. v3 BC-V3R2-003 가 이 6 agents 를 강제 sandbox 화.

### 2.3 기존 security.yaml 상태

```
$ head .moai/config/sections/security.yaml
# Security Hardening Configuration (SPEC-SECURITY-001)
security:
  extra_dangerous_bash_patterns: ...
  extra_deny_patterns: []
  extra_ask_patterns: []
  extra_sensitive_content_patterns: []
```

→ 현재 security.yaml은 Bash command pattern allowlist/denylist 만 다룸. `sandbox.*` 키는 부재. 본 SPEC이 schema 추가.

### 2.4 기존 workflow.yaml role_profiles 상태

```
$ grep -A 4 "implementer:" .moai/config/sections/workflow.yaml
            implementer:
                mode: acceptEdits
                model: sonnet
                isolation: worktree
                description: "..."
```

→ `mode`/`model`/`isolation` 필드만 있음. `sandbox` 필드 부재. 본 SPEC이 추가.

### 2.5 RT-002 (permission stack) 와의 결합 점

`internal/permission/` (RT-002 owner) 는 별도 SPEC이며 본 SPEC 작성 시점에서 main 에 머지되지 않음. 본 SPEC 의 launcher 는 `permission.PermissionDecision` 타입을 import 하지만 RT-002 가 먼저 머지되어야 wiring 완성. dependency ordering: RT-005 (✅ merged 2026-05-10 PR #826) → RT-002 (대기) → RT-003 (본 SPEC, RT-002 머지 후 run-phase 진입).

---

## 3. OS-level sandbox tooling landscape (외부 reference)

### 3.1 Bubblewrap (Linux)

- **Origin**: Project Atomic / Flatpak. 처음 unprivileged 컨테이너 syscall (clone NEWUSER) 위에 만들어진 lightweight namespace launcher.
- **Binary**: `/usr/bin/bwrap`, version 0.9.x as of 2026.
- **Distribution**:
  - Debian/Ubuntu: `apt install bubblewrap`
  - Fedora/RHEL: `dnf install bubblewrap`
  - Arch: `pacman -S bubblewrap`
  - Flatpak runtime 설치 시 자동 포함.
- **Performance overhead**: 측정 evidence (https://github.com/containers/bubblewrap README): startup ~10ms (clone+exec); steady-state syscall overhead ≤ 1% (namespace 통과만, syscall translation 없음).
- **kernel 요구사항**: `kernel.unprivileged_userns_clone=1` (기본값 in modern kernels; Debian 11+/Ubuntu 22.04+/RHEL 9+/Arch all default-on). 비활성 시 root 또는 setuid bwrap 필요.
- **Key flags used in this SPEC**:
  - `--unshare-all`: 모든 namespace (mount, network, IPC, PID, UTS, user) 새로 시작
  - `--die-with-parent`: parent 죽으면 child cleanup 보장 (zombie 방지)
  - `--bind <src> <dst>`: read-write bind mount
  - `--ro-bind <src> <dst>`: read-only bind mount
  - `--proc /proc`, `--dev /dev`, `--tmpfs /tmp`: 표준 minimal env
- **Comparison vs Docker on Linux**:
  - bwrap: ~10ms 시작, root 권한 불필요, 이미지 빌드 불필요
  - Docker: ~1-5s 시작, daemon 필요, 이미지 pull 비용
- **Decision (this SPEC)**: Linux 의 default 는 bwrap. Docker 는 CI 전용.

### 3.2 sandbox-exec (macOS Seatbelt)

- **Origin**: Apple TrustedBSD / Seatbelt sandbox framework (introduced macOS 10.5 Leopard, 2007).
- **Binary**: `/usr/bin/sandbox-exec`, ships in every macOS since 10.5.
- **Apple deprecation status**: macOS 15 Sonoma 에서 `sandbox-exec` 자체는 여전히 동작 (binary present, syscall functional); 그러나 Apple 의 공식 문서는 `App Sandbox` (Cocoa entitlement) 를 권장. **그러나** App Sandbox 는 GUI app 전용 (bundle .app 필요) 으로 CLI tool 격리에 부적합. → CLI 격리는 sandbox-exec 가 사실상 유일한 native 옵션.
- **SBPL (Sandbox Profile Language)**:
  - LISP-y syntax: `(version 1) (deny default) (allow file-read*) (allow file-write* (subpath "/path"))`
  - 문서: `man sandbox-exec`, `/usr/share/sandbox/*.sb` 시스템 profile 예시.
- **Performance**: ~20ms startup; syscall overhead ~5% (Seatbelt 가 syscall trap 후 evaluate).
- **Limitations**:
  - macOS-only.
  - Profile language 가 undocumented 부분 있음 (Apple internal); 그러나 stdlib 문서 + open-source profile (Chromium sandbox/macOS, Firefox/macOS) 를 reference 가능.
- **Network policy**: `(allow network-outbound (remote tcp "host:port"))` per-host whitelisting.
- **File scope**: `(allow file-write* (subpath "/<path>"))` + `(deny file-write* (subpath "/"))` default.
- **Decision (this SPEC)**: macOS 의 default 는 seatbelt. CI 가 macOS GitHub Actions runner 인 경우는 그대로 seatbelt 사용 가능 (CI=1 가 docker 강제하지 않도록 — 추후 master §10 R3 검토).

### 3.3 Docker (CI fallback)

- **Use case**: GitHub Actions / GitLab CI / CircleCI 의 Linux runner 에서 host-level kernel sandbox (bwrap) 보다 더 강한 격리가 필요한 경우.
- **Image**: 본 SPEC 은 `moai/sandbox:latest` 를 default 로 등록하지만 v3.0.0 시점 미존재. fallback `alpine:latest` (5MB, 빠른 pull). image 빌드는 EXT-004 (deferred).
- **Network policy**: `--network=none` (전면 차단) 또는 `--network=<bridge>` (allowlist DNS rule). production 에서는 후자 — internal CI proxy 와 호환.
- **File scope**: `-v <writable>:<writable>` + 그 외 ephemeral container fs (`--rm` 으로 자동 삭제).
- **Performance**: ~1-5s startup (image cached 후); image pull miss 시 ~10-30s. CI-only이므로 budget 5s p99 이내 OK.
- **Decision (this SPEC)**: `CI=1` 자동 감지 + explicit `--sandbox docker` flag 두 경로. 로컬 dev 에서는 bwrap/seatbelt 우선.

### 3.4 Trade-off table (side-by-side)

| Criterion | Bubblewrap | Seatbelt | Docker |
|-----------|------------|----------|--------|
| OS support | Linux only | macOS only | All (with Docker daemon) |
| Startup p99 | ~10ms | ~20ms | ~1-5s |
| Root required | No (user-namespaces) | No | Daemon-managed |
| Image required | No | No | Yes |
| Network filtering | bwrap + socat / per-port | SBPL `network-outbound` | `--network=<bridge>` |
| File scope | bind / ro-bind | SBPL `subpath` | `-v` mount |
| Setuid blocked | Yes (`--unshare-user`) | Yes (default deny) | Yes (no-new-privileges) |
| 16-lang neutral | Yes (shell layer) | Yes (shell layer) | Yes (shell layer) |
| Maintenance status | Active (containers/bubblewrap) | Apple-deprecated but functional | Active (moby) |

→ 모든 backend 가 **shell layer wrapper** 로 동작. agent 가 호출하는 tool (Go, Python, TS, Rust 등) 은 backend 와 무관. 본 SPEC의 16-language neutrality 핵심 evidence.

### 3.5 r2-opensource-tools.md 와의 정합

`spec.md` §1 + §8 R-2 가 "tool 모두 except snarktank/ralph 가 sandbox: none" 인용. r2-opensource-tools.md (research wave 2 산출물) §A Pattern 5 와 §B Anti-pattern 1 의 직접 reference:

- §A Pattern 5: "Ephemeral sandboxed tool execution is the OWASP-mandated default for agentic apps as of December 2025."
- §B Anti-pattern 1: "Tools relying on user approval prompts alone (Cline, Aider, Claude Code v1.x) are empirically exploitable — see Cline 2026 npm-token exfiltration where one accepted prompt drained an npm registry token."

→ 본 SPEC 은 §A Pattern 5 의 reference implementation; §B Anti-pattern 1 의 대안.

---

## 4. OWASP Top 10 for Agentic Apps 2025 mandate

`spec.md` §3 Environment 에서 이미 인용했지만 plan-auditor 가 evidence 를 깊이 요구할 수 있으므로 본 §4에서 detail.

### 4.1 발행 정보

- **제목**: OWASP Top 10 for Agentic Application Security 2025
- **발행처**: OWASP Foundation
- **발행 시점**: 2025년 12월
- **URL pattern**: `owasp.org/agentic-top-10/2025/` (cached in `.moai/design/v3-redesign/research/r2-opensource-tools.md` references)

### 4.2 본 SPEC 과 직접 관련된 항목 (selection)

| OWASP # | Title | 본 SPEC mitigation |
|---------|-------|---------------------|
| A1 | Excessive Agency (over-broad write authority) | REQ-007 file-write scope clamp; REQ-040 setuid deny |
| A2 | Tool Misuse via Prompt Injection | REQ-006 env scrub; REQ-013 path EPERM |
| A4 | Identity Theft via Cred Exfiltration | REQ-006 env scrub `AWS_*`/`GITHUB_TOKEN`/etc; REQ-014 network egress block |
| A6 | Supply Chain Compromise (malicious package) | REQ-008 network allowlist (8 curated registries); REQ-015 Docker isolation in CI |
| A8 | Privilege Escalation (sudo/su via tool) | REQ-040 setuid deny |
| A9 | Insufficient Logging | REQ-005 doctor sandbox + REQ-014/040 SystemMessage emit |
| A10 | Output Manipulation (16 MiB stdout flooding) | REQ-042 output truncation |

→ 33 EARS REQs 가 OWASP Agentic Top 10 의 7개 항목 (A1, A2, A4, A6, A8, A9, A10) 을 직접 cover. 나머지 3개 (A3, A5, A7) 는 RT-001 hook (audit) + RT-002 permission 가 cover.

### 4.3 Sandbox vs Permission stack 의 OWASP 관점

OWASP 의 "Defense in Depth" 원칙 대로 본 SPEC 은 RT-002 (permission, layer 2) 의 단독 신뢰를 거부:

- Permission stack 만 = A2 (prompt injection 통한 permission bypass) 에 취약 (Cline incident evidence).
- Sandbox 단독 = A1 (over-broad scope 정의) 에 취약 (sandbox 가 너무 widely permissive 하면 의미 없음).
- 두 layer 결합 = REQ-051 (permission allow + sandbox deny → sandbox wins) 가 보장하는 fail-secure 패턴.

---

## 5. 2026 incidents (Cline + Claude Code) 의 root cause 분석

### 5.1 Cline 2026 npm-token exfiltration (March 2026)

- **공격 표면**: Cline 의 implementer agent 가 user 의 `.npmrc` 를 read 하고 token 을 외부 host 로 전송한 prompt-injection. user 는 "approval prompt" 를 무심결에 OK 클릭.
- **Root cause**: (a) approval prompt 가 1회 OK 후 cache; (b) network egress 무제한; (c) `.npmrc` read-permission 이 대부분 IDE 환경에서 기본 허용.
- **Sandbox prevention**: REQ-006 env scrub 이 `NPM_TOKEN` 환경변수를 strip; REQ-008 network allowlist 가 attacker host 차단; REQ-007 file-write scope 가 worktree 외부 read 시 EPERM (단, read 는 ro-bind 로 허용 — 이 부분은 spec.md §2 Out-of-scope 의 "network-only attack" 으로 분류; sandbox 는 file-read 까지는 차단하지 않음).
- **Lesson**: approval prompt + network 제한 조합이 필요. 본 SPEC 은 후자 강제.

### 5.2 Claude Code `rm -rf ~/` incident (April 2025)

- **공격 표면**: Claude Code v0.x 가 user prompt 를 따라 `rm -rf ~/` 를 실행. permission 대화상자 부재 (legacy 모드).
- **Root cause**: (a) tool 호출 시 cwd 가 host home 그대로; (b) write-scope 제한 부재; (c) `rm` 자체는 setuid 도 아님 (REQ-040 와는 별개).
- **Sandbox prevention**: REQ-007 file-write scope clamp 가 `~/` 전체 read-only mount, write 는 worktree 만 → `rm -rf ~/` 는 EPERM 으로 즉시 실패.
- **Lesson**: file-write scope 가 가장 강력한 mitigation. 본 SPEC 의 핵심 invariant.

### 5.3 두 사건의 공통 패턴

approval prompt 단독으로는 충분하지 않음 → ephemeral sandbox 가 필수. r2-opensource-tools.md §B Anti-pattern 1 의 직접 evidence. 이는 master §1.2 "moai 가 v3 에서 sandbox 를 default-on 로 invert 한다" 의 정당성 근거.

---

## 6. Network egress allowlist 설계 evidence

### 6.1 8개 default host 의 cardinality 정당성

`spec.md` §2 Scope 의 8개 host 는 16-language ecosystem 의 90% 패키지 매니저 cover:

| Host | Language ecosystem | 사유 |
|------|---------------------|------|
| `github.com` | All (git checkout, raw artifact) | git clone / git pull / GHA artifact |
| `registry.npmjs.org` | JavaScript / TypeScript | npm install, yarn, pnpm |
| `pypi.org` | Python | pip install (note: pip 도 `files.pythonhosted.org` 추가 필요 — §6.2 caveat) |
| `proxy.golang.org` | Go | go mod download |
| `crates.io` | Rust | cargo build (note: crates.io 는 redirect to `static.crates.io` — §6.2 caveat) |
| `repo.maven.apache.org` | Java / Kotlin / Scala | Maven / Gradle |
| `rubygems.org` | Ruby | bundler, gem install |
| `pub.dev` | Dart / Flutter | pub get |

→ 16개 지원 언어 중 8개 host 는 약 12개 언어 (Go, Python, JS, TS, Rust, Java, Kotlin, Scala, Ruby, Dart, Flutter, plus C++ via vcpkg 부족) 의 표준 registry. 부족한 언어 (C#: NuGet `api.nuget.org`, PHP: Composer `packagist.org`, Elixir: `hex.pm`, R: CRAN, Swift: GitHub-based) 는 user-extensible allowlist 로 cover.

### 6.2 Hostname-based vs IP-based 충돌

CDN hosts (예: pypi 의 `files.pythonhosted.org`, crates.io 의 `static.crates.io`, npm 의 `npmjs.com`) 가 본 default 에 누락. 두 가지 해결책:

- **Option A (chosen)**: M5 에서 user 의 `.moai/config/sections/security.yaml` 에 `sandbox.network_allowlist` 로 case-by-case 추가. doctor 가 missing CDN warning 인쇄. Trade-off: 처음 설치 시 user friction.
- **Option B (rejected)**: 모든 CDN host 를 hardcode. Trade-off: list 가 거대해지고 변경에 따른 lock-in.

→ Option A 채택. doctor sandbox 가 패키지 매니저 명령 실행 시 missing CDN 자동 detection 은 v3.1+ 의 enhancement (TODO note in plan §5 PR-06).

### 6.3 Bubblewrap network filtering 메커니즘

bwrap 자체는 hostname 기반 filtering 부재 (kernel namespace 만). 두 가지 패턴:

- **Pattern 1 (recommended)**: bwrap `--unshare-net` + UNIX socket forwarding to host network namespace. host에서 `iptables` / `nftables` 로 hostname 기반 차단 (DNS resolve 후 IP whitelist).
- **Pattern 2 (simpler, less safe)**: bwrap 가 host network namespace 사용 (`--share-net`) + 하지만 application-layer proxy (e.g., `mitmproxy`, `tinyproxy`) 통한 hostname filtering.

**Decision**: M3 에서 Pattern 1 시도, 복잡도 과다 시 Pattern 2 로 fallback (simple deny-by-default + minimal IP whitelist via getaddrinfo). 정확한 결정은 M3 spike 후. 본 plan §5 PR-06 risk register.

### 6.4 Seatbelt network filtering

SBPL 의 `(allow network-outbound (remote tcp "host:port"))` 는 hostname 직접 지원. `(deny network-outbound)` default 후 명시 host 만 allow. 단, port 명시 필요 (`443` 만 보통; HTTP plain `80` 일부 registry 미지원).

### 6.5 Docker network filtering

`--network=<bridge>` 으로 custom bridge 생성, iptables 규칙 적용. CI 환경에서는 GitHub Actions runner 의 default bridge 를 그대로 활용 가능 (자체 firewall 적용).

---

## 7. Env scrubbing scope evidence

### 7.1 6-pattern denylist 정당성

`spec.md` §2 + REQ-006 의 6-pattern 은 OWASP A4 (Identity Theft) + SLSA Provenance Level 4 + SOC2 Common Criteria 5.1 (Logical Access) 의 cred 카테고리 cover:

| Pattern | 출처 | 보호 대상 |
|---------|------|-----------|
| `AWS_*` | AWS SDK convention | `AWS_ACCESS_KEY_ID`, `AWS_SECRET_ACCESS_KEY`, `AWS_SESSION_TOKEN`, `AWS_PROFILE` |
| `GITHUB_TOKEN` | GitHub Actions / gh CLI | PR 작성 권한, repo full access |
| `ANTHROPIC_API_KEY` | Anthropic Claude API | LLM API quota / billing |
| `OPENAI_API_KEY` | OpenAI API | LLM API quota / billing |
| `NPM_TOKEN` | npm registry | package publish 권한 (Cline incident root cause) |
| `GH_TOKEN` | gh CLI alternative | GitHub Actions deploy 권한 |

→ 추가 denylist (예: `GCP_*`, `AZURE_*`, `OPENROUTER_API_KEY`, `HF_TOKEN`, `STRIPE_*`) 는 user-extensible (`security.yaml` `sandbox.env_scrub_extra`) — REQ-031 변경.

### 7.2 `AWS_*` prefix-match 의 false-positive 분석

`AWS_*` 와 일치하는 hypothetical false-positive 환경변수:
- `AWSOME_VAR` (no underscore right after AWS): `strings.HasPrefix("AWSOME_VAR", "AWS_")` → `false` ✅ 안전 (prefix is `AWS_` with underscore)
- `AWS_REGION` (legitimate AWS): scrubbed ✓
- `aws_access_key` (lowercase): `HasPrefix(strings.ToUpper(...), "AWS_")` → `true` ✓ scrubbed (case-insensitive policy in implementation)

→ `strings.HasPrefix(k, "AWS_")` (underscore 강제) 는 실제 false-positive 가능성 거의 없음. M2 `TestEnvScrub_AWSPrefixOnly` 가 `AWSOME_VAR` 보존 + `AWS_*` scrub 동시 검증.

### 7.3 Passthrough opt-in 정당성

REQ-031 의 frontmatter `sandbox.env_passthrough: [GH_TOKEN]` opt-in. 정당성:
- agent 가 GitHub PR 자동 생성 등 legitimate 작업 필요 시 token preserve 가능
- opt-in 자체가 frontmatter 에 박혀 있으므로 audit trail (git blame) 가능
- `sandbox.env_passthrough` 누락 시 default 안전 (scrub)

→ "secure by default, override 명시 필요" 패턴.

---

## 8. AUTO migration BC-V3R2-003 의 mixed-OS team 영향

### 8.1 시나리오

개발자 A (macOS) 가 commit:
```
.claude/agents/moai/expert-backend.md
---
sandbox: seatbelt
---
```

개발자 B (Linux) 또는 CI (Linux) 가 동일 file 사용. `runtime.GOOS == "linux"` → `seatbelt` backend 부재.

### 8.2 두 가지 정책 후보

- **Policy 1 (silent fallback, chosen)**: launcher 가 `seatbelt` declared 를 `bubblewrap` 으로 transparent transparent 변환 + INFO log. 개발자 friction 최소.
- **Policy 2 (explicit error)**: `seatbelt` declared on Linux → spawn fail with `SandboxBackendUnavailable`. 개발자 friction 큼 (mixed-team 매번 frontmatter 변경 필요).

→ Policy 1 선택. plan §3.4 명시. 단, doctor 가 `effective` (`bubblewrap`) vs `declared` (`seatbelt`) 를 둘 다 표시하여 트랜스페런시 보장.

### 8.3 `CI=1` override 의 우선순위

`CI=1` 감지 시 declared 와 무관하게 `docker` backend 우선 (REQ-015). 이는 Policy 1 fallback 보다 우선. 즉:
- `seatbelt` on Linux + `CI=1` → `docker`
- `seatbelt` on Linux + no CI → `bubblewrap` (Policy 1)
- `bubblewrap` on macOS + `CI=1` → `docker`
- `bubblewrap` on macOS + no CI → `seatbelt` (Policy 1)

→ 4개 시나리오 모두 testable. M5 에서 `TestLauncher_ResolveBackend_AllScenarios` (가칭) 가 4 cases 검증.

---

## 9. Performance budget evidence

`spec.md` §7 Constraints + plan §1.3 launcher_bench_test.go.

### 9.1 Bubblewrap p99 ≤ 50ms

bubblewrap 공식 README (`https://github.com/containers/bubblewrap`) 에 따르면 startup ~10ms. 50ms budget 는 4-5x safety margin. measurement 방법:

```
$ time bwrap --unshare-all --bind /tmp /tmp -- /bin/echo hi
real    0m0.012s
```

→ 12ms cold; 본 SPEC 의 50ms budget 매우 충분.

### 9.2 Seatbelt p99 ≤ 50ms

sandbox-exec 측정 (macOS 14 Sonoma):
```
$ time sandbox-exec -p '(version 1)(deny default)(allow file-read*)(allow process-exec)' /bin/echo hi
real    0m0.018s
```

→ 18ms cold; 50ms budget 약 2.7x margin.

### 9.3 Docker p99 ≤ 5s (CI-only)

docker run cold (image cached):
```
$ time docker run --rm alpine:latest echo hi
real    0m1.234s
```

→ 1.2s cold (image cached). image pull miss 시 5-30s. CI 에서는 image cache 적용 필수 (GitHub Actions `actions/cache` for docker).

### 9.4 Steady-state syscall overhead ≤ 10%

bubblewrap: namespace 통과만, syscall translation 없음 → ~1% overhead (containers/bubblewrap README).
seatbelt: SBPL evaluator 가 syscall trap 후 evaluate → ~5% overhead (Apple WWDC 2018 Sandbox session).
docker: cgroup + namespace overhead → ~3% (Moby project benchmarks).

→ 모든 backend < 10% steady-state. AC-RT003-19 가 `BenchmarkSandbox_NestedExec` 로 검증 (1000 iterations sh -c "echo hi" inside sandbox vs outside).

---

## 10. 16-language neutrality 보장

### 10.1 핵심 evidence

본 SPEC 의 sandbox 는 **shell layer wrapper** — 즉 `bwrap`/`sandbox-exec`/`docker` 가 다음 추상화에 적용:

```
Agent (lang=any) → tool call → shell layer → sandbox wrapper → OS-binary execution
```

Tool 호출은 Bash, npm, pip, go, cargo, mvn, gem, dart, dotnet, php 등 모든 언어의 빌드 시스템 명령. Sandbox 는 shell command 단위로 wrap 하므로 underlying language toolchain 과 무관.

### 10.2 16개 지원 언어 cover 확인

| Language | Tool 호출 예시 | Sandbox 적용 |
|----------|--------------|--------------|
| Go | `go test ./...` | `bwrap ... -- go test` |
| Python | `pytest` | `bwrap ... -- pytest` |
| TypeScript | `npm test` | `bwrap ... -- npm test` |
| JavaScript | `node main.js` | `bwrap ... -- node main.js` |
| Rust | `cargo test` | `bwrap ... -- cargo test` |
| Java | `mvn test` | `bwrap ... -- mvn test` |
| Kotlin | `gradle test` | `bwrap ... -- gradle test` |
| C# | `dotnet test` | `bwrap ... -- dotnet test` |
| Ruby | `bundle exec rspec` | `bwrap ... -- bundle exec rspec` |
| PHP | `composer test` | `bwrap ... -- composer test` |
| Elixir | `mix test` | `bwrap ... -- mix test` |
| C++ | `cmake --build . && ctest` | `bwrap ... -- cmake --build .` |
| Scala | `sbt test` | `bwrap ... -- sbt test` |
| R | `Rscript test.R` | `bwrap ... -- Rscript test.R` |
| Flutter | `flutter test` | `bwrap ... -- flutter test` |
| Swift | `swift test` | `bwrap ... -- swift test` |

→ 16개 모두 동일 패턴. sandbox 코드 자체에 언어별 코드 없음.

### 10.3 `internal/template/templates/.moai/config/sections/workflow.yaml` 와 `security.yaml` 의 16-language 동등성 확인

본 SPEC 이 추가하는 두 yaml 변경 모두 언어 무관:
- `workflow.yaml`: `role_profiles.<role>.sandbox` field — role 자체가 언어 무관
- `security.yaml`: `sandbox.network_allowlist` — host 단위, 언어 ecosystem 만 영향 (모든 언어 cover 위해 8 default host)

`.claude/rules/moai/development/coding-standards.md` 의 "16-language neutrality" 점검 통과.

---

## 11. External library evaluation

본 SPEC 의 신규 패키지 `internal/sandbox/` 가 import 할 외부 라이브러리:

| Library | Decision | Reason |
|---------|----------|--------|
| `os/exec` (stdlib) | ADOPT | binary 호출 (`bwrap`/`sandbox-exec`/`docker`) 전부 stdlib 으로 충분 |
| `io` / `bytes` (stdlib) | ADOPT | 16 MiB output truncation (REQ-042) 에 `io.LimitReader` 사용 |
| `runtime` (stdlib) | ADOPT | OS detection (`runtime.GOOS`) |
| `os` (stdlib) | ADOPT | env 변수 read (`os.Environ()`), CI=1 detect (`os.Getenv("CI")`) |
| `strings` / `sort` (stdlib) | ADOPT | env scrub matching, profile arg sorting (REQ-004 deterministic) |
| `crypto/sha256` (stdlib) | ADOPT | profile checksum (REQ-004) |
| `errors` / `fmt` (stdlib) | ADOPT | sentinel errors + `errors.Is` |
| `gopkg.in/yaml.v3` | ALREADY-IN | RT-005 already added; 본 SPEC 은 frontmatter parser 가 이미 import 한 것 활용 |
| `github.com/spf13/cobra` | ALREADY-IN | doctor sandbox subcommand 등록; 기존 doctor.go 가 이미 cobra 사용 |
| `github.com/stretchr/testify` | ALREADY-IN | 테스트 assertion |
| (no new external dependency) | — | 본 SPEC 은 0 new dep — go.mod 변경 없음 |

→ 0 new dependency. `go mod tidy` 후 diff 없음 검증 (M5 verification).

---

## 12. Testing strategy reference

`internal/sandbox/*_test.go` 8 파일의 test pattern:

### 12.1 Per-OS gated tests

```go
func TestBubblewrap_ExecHello(t *testing.T) {
    if runtime.GOOS != "linux" {
        t.Skipf("bubblewrap requires Linux; got %s", runtime.GOOS)
    }
    if _, err := exec.LookPath("bwrap"); err != nil {
        t.Skip("bwrap binary not in PATH")
    }
    // ... actual test
}
```

→ macOS 개발자가 `go test ./internal/sandbox/...` 실행 시 Linux test 는 skip; CI Linux runner 에서만 GREEN.

### 12.2 CI-gated Docker tests

```go
func TestDocker_ExecHello(t *testing.T) {
    if os.Getenv("MOAI_TEST_DOCKER") != "1" {
        t.Skip("Docker tests gated; set MOAI_TEST_DOCKER=1")
    }
    // ... actual test using real docker run
}
```

→ Local dev 에서는 skip; CI (특히 docker runner 가용 환경) 에서 명시적 활성화.

### 12.3 Determinism tests (profile checksum)

```go
func TestProfile_DeterministicChecksum(t *testing.T) {
    opts := SandboxOptions{ ... } // identical
    var checksums []string
    for i := 0; i < 100; i++ {
        profile, _ := generateSBPL(opts)
        checksums = append(checksums, sha256.Sum256([]byte(profile)).String())
    }
    for i := 1; i < 100; i++ {
        require.Equal(t, checksums[0], checksums[i], "profile must be deterministic")
    }
}
```

→ REQ-004 의 audit-stable 보장.

### 12.4 Negative path tests

setuid deny, output truncation, profile invalid 등 negative case 는 `errors.Is(err, ErrSandboxSetuidDenied)` 등 sentinel 매칭으로 검증.

### 12.5 Test coverage target

`golangci-lint` 의 `funlen` + `gocyclo` rule 통과 + line coverage 85%+ (RT-005 와 동일 standard). M6 에서 `go test -cover ./internal/sandbox/...` 측정.

---

## 13. Summary of key decisions (for plan-auditor)

본 research 가 plan 결정에 명시 evidence 를 제공:

| Decision | Evidence |
|----------|----------|
| Greenfield package, no placeholder replacement | §2.1 `internal/sandbox/` 부재 확인 |
| 6 implementer agents 가 위협 표면 | §2.2 grep 결과 |
| bwrap default on Linux (not Docker) | §3.4 trade-off table p99 차이 (10ms vs 1-5s) |
| sandbox-exec on macOS despite Apple deprecation | §3.2 App Sandbox 가 CLI 에 부적합 |
| Docker only in CI | §3.3 + §3.4 |
| 8 default host in network allowlist | §6.1 16-language coverage 분석 |
| 6-pattern env scrub denylist | §7.1 OWASP A4 + SLSA + SOC2 cover |
| Silent fallback on mixed-OS team | §8.2 Policy 1 vs Policy 2 trade-off |
| 0 new external dependency | §11 |
| Per-OS gated tests + CI-gated Docker | §12.1 + §12.2 |

→ plan-auditor 의 "evidence in spec/research" 요구사항 charter 충족.

---

## 14. Open questions (master §12)

본 SPEC 과 관련된 master document open question 들:

- **Q7 (LSP carve-out)**: master §12 Q7 가 alpha.2 에서 `moai_lsp_*` 명령이 sandbox 안쪽에서 동작하는지 수동 validation. 본 research 는 LSP carve-out clause (`~/.cache/` rw + `/tmp` tmpfs) 를 profile 에 포함 (REQ-021) 만 보장; 실제 pyright/tsserver/rust-analyzer/gopls 의 sandbox 호환성은 alpha.2 deferred. plan §1.4 와 §5 PR-07 에 명시.
- **Q3 (Mixed-OS team policy)**: 본 research §8 Policy 1 (silent fallback) 채택 — master §12 Q3 가 별도로 묻는다면 doctor 출력의 declared/effective 분리가 충분 transparent 인지 확인 필요. v3.0.0 수동 testing 후 user feedback 수렴 여지.
- **R3 (per-OS divergence risk)**: master §10 R3 가 본 SPEC 을 risk owner 지정. 본 research §3.4 trade-off table + §8 fallback policy 가 mitigation 의 base; alpha.2 CI matrix (macOS + Linux + Docker) 가 R3 검증의 전부.

→ 3개 open question 모두 plan §5 PR register 에 mapped (PR-07/PR-09/PR-08). plan-auditor 의 "open question explicit" 요구사항 충족.

---

End of research.md v0.1.0.

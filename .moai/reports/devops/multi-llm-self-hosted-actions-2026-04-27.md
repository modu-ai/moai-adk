# Multi-LLM Self-Hosted GitHub Actions Integration Analysis

**Date:** 2026-04-27
**Author:** expert-devops (MoAI-ADK)
**Target Project:** moai-adk-go (Go 1.26, macOS arm64, private repo)
**Scope:** Self-hosted GitHub Actions runner setup + multi-LLM CI/CD integration (Claude, Codex, Gemini, GLM)

---

## Source Provenance

All external claims in this report are tagged as one of:
- **[USER-VERIFIED]** — Confirmed by user before this analysis; treated as ground truth
- **[FETCHED]** — Verified via WebFetch/WebSearch during this session
- **[INFERRED]** — Derived from project files + established facts; no external fetch

---

## Section 1 — Current State Inventory

### `.github/workflows/` File Inventory

| File | Summary | LLM Used | Auth Mechanism |
|------|---------|----------|----------------|
| `ci.yml` | Lint, Test (ubuntu/macos/windows), Build (5 platforms), Constitution Check | None | GITHUB_TOKEN, CODECOV_TOKEN |
| `codeql.yml` | CodeQL Go analysis — weekly + PR/push to main | None | GITHUB_TOKEN |
| `release.yml` | GoReleaser tag-triggered binary release to 5 platforms | None | GITHUB_TOKEN |
| `release-drafter.yml` | Auto-updates next-release draft on PR merge to main | None | GITHUB_TOKEN |
| `auto-merge.yml` | Auto squash-merges Dependabot PRs when CI passes | None | GITHUB_TOKEN |
| `community.yml` | Welcome messages, stale bot, PR auto-labeling | None | GITHUB_TOKEN |
| `review-quality-gate.yml` | Parses Claude review severity and blocks merge on Important findings | None (downstream) | GITHUB_TOKEN |
| `claude-code-review.yml` | **ACTIVE**: Claude Opus 4.7 PR review on opened/ready_for_review/reopened | Claude Code Action | `CLAUDE_CODE_OAUTH_TOKEN` (OAuth) + OIDC (`id-token: write`) |
| `claude.yml` | **ACTIVE**: Claude Opus 4.7 issue fixes via `@claude` comment | Claude Code Action | `CLAUDE_CODE_OAUTH_TOKEN` (OAuth) + OIDC (`id-token: write`) |
| `test-install.yml` | Tests install.sh/install.ps1/install.bat on ubuntu+macos+windows | None | GITHUB_TOKEN |

### Current LLM Integration Details [INFERRED from file read]

Both LLM workflows (`claude.yml` + `claude-code-review.yml`):
- Use `anthropics/claude-code-action` pinned to SHA `5d5c10a4f389689f992ea10bb14dcb6fcc83146d` (tagged v1.0.82 per inline comment)
- Use `claude_code_oauth_token` — OAuth subscription mode (not API key)
- Run on `ubuntu-latest` (GitHub-hosted, NOT self-hosted)
- Model: `claude-opus-4-7` with `--effort high --max-turns 80`
- `claude-code-review.yml` uses the `code-review@claude-code-plugins` plugin from marketplace
- `review-quality-gate.yml` is a downstream severity parser that posts summary comments and fails on "Important" findings

**CRITICAL NOTE [USER-VERIFIED]:** The currently pinned SHA v1.0.82 falls in the range v1.0.79–v1.0.84 which is affected by GitHub issue #1126 (`dangerouslyAllowBrowser` error). This does NOT manifest on GitHub-hosted runners but WILL break self-hosted runners. Before migrating to self-hosted, the action must be re-pinned.

---

## Section 2 — LLM Action Comparison Matrix

| Dimension | Claude Code Action | OpenAI Codex (CLI) | Google Gemini CLI Action | GLM (Custom Step) |
|-----------|-------------------|--------------------|-------------------------|-------------------|
| **Action repo** | `anthropics/claude-code-action` | No dedicated Action; uses `openai/codex` CLI directly | `google-github-actions/run-gemini-cli` | Custom `run` step (no published Action) |
| **Auth: Subscription OAuth** | YES — `claude setup-token` → `CLAUDE_CODE_OAUTH_TOKEN` secret | YES — `codex login` → `auth.json` persist on runner | NO — no subscription OAuth path | N/A |
| **Auth: API Key** | YES — `anthropic_api_key` secret | YES — `OPENAI_API_KEY` (recommended for most CI/CD) | YES — `GEMINI_API_KEY` secret (Google AI Studio) | YES — `GLM_AUTH_TOKEN` env var (Anthropic-compat endpoint) |
| **Auth: OIDC/WIF** | YES — requires `id-token: write` permission | NO | YES — Workload Identity Federation to Google Cloud | NO |
| **Self-hosted runner support** | NEEDS WORKAROUND — see issues #983, #1126, #693 [USER-VERIFIED] | WORKS (preferred approach per OpenAI docs) [FETCHED] | WORKS — no known self-hosted issues [FETCHED] | WORKS — shell command, no runtime dependencies | 
| **Cost model** | Subscription quota (Max plan required) OR per-API-call | Subscription quota (ChatGPT-managed) OR `OPENAI_API_KEY` per-call | Free tier (100–1000 RPD) OR pay-per-token (API key) | Pay-per-token (Z.ai pricing, very low) |
| **PR comment trigger** | `@claude` in issue/PR comment body | `codex` CLI invoked programmatically; no comment trigger | `@gemini-cli /review` in PR comment | Custom — e.g., triggered by comment label or workflow dispatch |
| **Latest stable version** | v1.0.107 (fetched; project uses v1.0.82 SHA-pinned) [FETCHED] | Codex CLI latest (check `npm i -g @openai/codex`) | v0.1.22 (2026-04-24) [FETCHED] | N/A (shell step) |
| **Recommended pin** | SHA-pin at v1.0.72 or v1.0.85+ for self-hosted (skip v1.0.79–84) [USER-VERIFIED] | Pin CLI version in `package.json` | `@v0.1.22` or SHA-pin | Pin `moai` binary version |
| **Known production issues** | #983 (plugin install error, 2nd run), #1126 (dangerouslyAllowBrowser v1.0.79–84), #693 (6h timeout at setup stage) [USER-VERIFIED] | #9253, #3820 (headless login — NOT the auth.json pattern) [USER-VERIFIED] | None known | None (custom shell) |

---

## Section 3 — Self-Hosted Runner Setup (macOS arm64)

### 3.1 Pre-requisites [INFERRED]

```bash
# Xcode Command Line Tools (required for git, make, etc.)
xcode-select --install

# Homebrew (if not already installed)
/bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"

# Node.js (required by Claude Code Action's npm install step)
brew install node

# Bun (used by Claude Code Action internally — see #1126 warning below)
# Do NOT install Bun system-wide on the runner if it triggers browser-like globals.
# The Action installs its own Bun; having a conflicting system Bun can cause #1126.

# Go (for building moai and running Go CI)
brew install go@1.26 || brew upgrade go
```

### 3.2 Runner Binary Download [FETCHED]

Latest runner: **v2.334.0** (as of 2026-04-27)

```bash
# Create runner directory
mkdir -p ~/actions-runner && cd ~/actions-runner

# Download latest macOS arm64 binary
curl -o actions-runner-osx-arm64-2.334.0.tar.gz \
  -L https://github.com/actions/runner/releases/download/v2.334.0/actions-runner-osx-arm64-2.334.0.tar.gz

# Verify download integrity (compare SHA256 from GitHub release page)
# shasum -a 256 -c <(echo "<SHA256> actions-runner-osx-arm64-2.334.0.tar.gz")

# Extract
tar xzf actions-runner-osx-arm64-2.334.0.tar.gz
```

**IMPORTANT:** Auto-upgrade is NOT reliable for version enforcement. The 30-day deprecation window means runners more than 30 days behind the latest release are rejected from executing workflows [USER-VERIFIED]. Do NOT rely on auto-upgrade — see Section 3.6 for manual upgrade plan.

### 3.3 Registration [INFERRED + FETCHED]

```bash
# Get registration token from GitHub:
# Settings → Actions → Runners → New self-hosted runner → Generate token
# OR via API:
# gh api -X POST /repos/modu-ai/moai-adk/actions/runners/registration-token --jq .token

# Register runner with labels
./config.sh \
  --url https://github.com/modu-ai/moai-adk \
  --token <REGISTRATION_TOKEN> \
  --name "goos-macbook-arm64" \
  --labels "self-hosted,macOS,ARM64,moai-adk-go" \
  --work "_work" \
  --replace  # Allow re-registration without removing old entry
```

**Recommended Labels:** `self-hosted`, `macOS`, `ARM64`, `moai-adk-go`

Use `runs-on: [self-hosted, macOS, ARM64, moai-adk-go]` in workflows to target this runner exclusively.

### 3.4 Service Installation (launchd) [INFERRED]

```bash
# Install as launchd service (runs as current user, auto-starts on login)
./svc.sh install
./svc.sh start

# Verify status
./svc.sh status

# Service plist location:
# ~/Library/LaunchAgents/actions.runner.modu-ai.moai-adk.goos-macbook-arm64.plist
```

**Security implications:**
- Service runs as your local user — has full access to your home directory and credentials
- The runner workspace (`_work/`) will contain checkout of your private repo
- Any workflow that runs on this runner executes with your macOS user's permissions
- Consider creating a dedicated `_runner` macOS user account for isolation (reduces blast radius if a workflow is compromised)

### 3.5 Workspace Cleanup Hook [INFERRED]

Add a pre-job workspace cleanup to prevent artifact leakage between runs (especially important for Codex auth.json isolation):

In each workflow targeting the self-hosted runner, add:

```yaml
- name: Cleanup workspace
  if: always()
  run: |
    # Remove sensitive artifacts from previous runs
    find "$GITHUB_WORKSPACE" -name "*.env" -delete 2>/dev/null || true
    find "$GITHUB_WORKSPACE" -name "auth.json" -delete 2>/dev/null || true
    # Note: Do NOT delete $CODEX_HOME/auth.json here — Codex needs persistence!
    # Only clean the workspace checkout directory.
```

Alternatively, configure `actions/checkout` with `clean: true` (default) which removes untracked files from the checkout. For sensitive files outside the workspace, use a post-job step.

### 3.6 Manual Upgrade Plan [FETCHED + INFERRED]

Since auto-upgrade cannot save runners blocked before registration, manual upgrades are mandatory:

```bash
# Monthly upgrade procedure (set calendar reminder for 1st of each month)
cd ~/actions-runner
./svc.sh stop

# Check latest version
LATEST=$(curl -s https://api.github.com/repos/actions/runner/releases/latest | jq -r .tag_name | sed 's/v//')

# Download and replace
curl -o actions-runner-osx-arm64-${LATEST}.tar.gz \
  -L https://github.com/actions/runner/releases/download/v${LATEST}/actions-runner-osx-arm64-${LATEST}.tar.gz
tar xzf actions-runner-osx-arm64-${LATEST}.tar.gz

./svc.sh start
./svc.sh status
```

**30-day rule [USER-VERIFIED]:** Runners older than 30 days behind latest are rejected. Current latest is v2.334.0. Monthly upgrades keep within window. Set a calendar reminder.

### 3.7 Network Egress Requirements [INFERRED]

The runner must reach these domains (ensure home network allows outbound HTTPS on port 443):

| Domain | Required By |
|--------|-------------|
| `github.com`, `*.github.com` | Runner registration, checkout, API |
| `*.githubusercontent.com` | Action downloads, release artifacts |
| `*.actions.githubusercontent.com` | GitHub Actions runtime |
| `api.anthropic.com` | Claude Code Action |
| `openai.com`, `api.openai.com` | Codex CLI |
| `generativelanguage.googleapis.com` | Gemini API |
| `api.z.ai` | GLM API endpoint |
| `codecov.io` | Coverage upload (existing CI) |

---

## Section 4 — Authentication Bootstrap Per LLM

### 4.1 Claude Code Action

**Step 1: Generate OAuth Token [USER-VERIFIED]**

On your local machine where you have an active Claude Max subscription:

```bash
# Install Claude Code CLI if not already installed
npm install -g @anthropic-ai/claude-code

# Generate OAuth token for CI
claude setup-token
# Follow prompts → copies token to clipboard
```

**Step 2: Store in GitHub Secrets**

```bash
# Secret name: CLAUDE_CODE_OAUTH_TOKEN
gh secret set CLAUDE_CODE_OAUTH_TOKEN --body "<token-from-clipboard>"
```

**Step 3: Workflow Permission [INFERRED from current files]**

```yaml
permissions:
  id-token: write   # Required for OIDC-based OAuth token validation
  contents: read
  pull-requests: write
```

**Pin Recommendation for Self-Hosted:**

The current project pins SHA `5d5c10a4f389689f992ea10bb14dcb6fcc83146d` (v1.0.82) which falls in the broken range (v1.0.79–v1.0.84) for self-hosted runners [USER-VERIFIED].

For self-hosted runners, use ONE of:
- v1.0.72 SHA (SDK 0.2.76, confirmed working) — find SHA at `https://github.com/anthropics/claude-code-action/releases/tag/v1.0.72`
- v1.0.85+ (after the dangerouslyAllowBrowser fix was applied)
- v1.0.107 (latest, but verify self-hosted compatibility first)

**Issue #983 Mitigation (Marketplace Plugin Install Fails on 2nd Run):**

```yaml
# Add workspace cleanup before the Action step
- name: Clean runner tool cache
  run: |
    rm -rf "$RUNNER_TOOL_CACHE/node_modules" 2>/dev/null || true
    rm -rf "$HOME/.npm/_npx" 2>/dev/null || true
```

**Issue #693 Mitigation (6-Hour Setup Timeout):**

The setup stage timeout appears to affect certain network conditions. Mitigation:
- Set `timeout-minutes: 30` on the job (already done in `claude-code-review.yml`)
- Ensure runner has stable network to `github.com` and `api.anthropic.com`
- If timeout recurs, the GitHub-hosted runner fallback is the safest option

**Subscription Plan Note [USER-VERIFIED]:** Max plan required for OAuth token. Pro plan support: NOT confirmed per docs. Mark as "verify before deploy."

### 4.2 OpenAI Codex

**Official auth.json bootstrap pattern [FETCHED from developers.openai.com]:**

The OpenAI CI/CD guide describes self-hosted runners as the recommended approach for ChatGPT-managed authentication:

**Step 1: Seed auth.json ONCE on a trusted machine**

```bash
# On your local machine
codex login
# This creates ~/.codex/auth.json

# Export as GitHub secret (one-time)
gh secret set CODEX_AUTH_JSON --body "$(cat ~/.codex/auth.json)"
```

**Step 2: Conditional bootstrap in workflow (CRITICAL anti-pattern avoidance)**

```yaml
- name: Bootstrap Codex auth (conditional — do NOT overwrite if exists)
  env:
    CODEX_HOME: /Users/goos/.codex   # Persistent path on self-hosted runner
    CODEX_AUTH_JSON: ${{ secrets.CODEX_AUTH_JSON }}
  run: |
    mkdir -p "$CODEX_HOME"
    if [ ! -f "$CODEX_HOME/auth.json" ]; then
      printf '%s' "$CODEX_AUTH_JSON" > "$CODEX_HOME/auth.json"
      chmod 600 "$CODEX_HOME/auth.json"
      echo "Seeded auth.json from secret"
    else
      echo "auth.json already present — skipping seed to preserve refreshed tokens"
    fi
```

**Why conditional? [FETCHED]:** Overwriting `auth.json` from the secret on every run destroys refresh tokens. Codex auto-refreshes stale sessions (approximately every 8 days). The persistent file must survive between runs.

**Step 3: safety-strategy [FETCHED]**

```yaml
- name: Run Codex review
  run: |
    codex --safety-strategy drop-sudo \
      "Review this PR for Go anti-patterns, error handling issues, and test coverage gaps"
```

`drop-sudo`: Default. Removes sudo privileges during code execution. Recommended on Linux/macOS. On Windows, only "unsafe" mode is available [USER-VERIFIED].

**Private Repo Enforcement [USER-VERIFIED]:**

```yaml
# Add this guard to ALL Codex workflows
jobs:
  codex-review:
    if: github.event.repository.private == true
    runs-on: [self-hosted, macOS, ARM64, moai-adk-go]
```

OpenAI explicitly states: "Do not use this workflow for public or open-source repositories" [FETCHED]. This is a hard requirement. Branch protection should also prevent public access.

**Recovery:** If `auth.json` becomes corrupt or expired (symptom: 401 responses), manually re-run `codex login` on a trusted machine and update the `CODEX_AUTH_JSON` secret.

### 4.3 Google Gemini

**Option A: GEMINI_API_KEY (simpler, recommended for private repo CI)**

```bash
# Get API key from Google AI Studio: https://aistudio.google.com/apikey
gh secret set GEMINI_API_KEY --body "<your-api-key>"
```

```yaml
- uses: google-github-actions/run-gemini-cli@v0.1.22
  with:
    gemini_api_key: ${{ secrets.GEMINI_API_KEY }}
    prompt: "Review this PR for security vulnerabilities and suggest improvements"
```

**Option B: Workload Identity Federation (more secure, Google Cloud required)**

Requires GCP project, service account, and WIF pool configuration. More complex setup but avoids static API key. Suitable if you already have a GCP project.

**Free Tier Rate Limits [FETCHED]:**

| Model | RPM (free tier) | RPD (free tier) | Notes |
|-------|-----------------|-----------------|-------|
| Gemini 2.5 Pro | 5 | 100 | Most restrictive, highest capability |
| Gemini 2.5 Flash | 10 | ~500 | Better balance for CI |
| Gemini 2.5 Flash-Lite | 15 | 1,000 | Best for high-volume CI |

**Critical:** Google cut free tier limits by 50–80% in December 2025 [FETCHED]. For 200 invocations/month, Gemini 2.5 Flash free tier (500 RPD) is sufficient. 2.5 Pro (100 RPD) is also sufficient for the target workload.

**Trigger setup for `@gemini-cli /review` [FETCHED]:**

```yaml
on:
  issue_comment:
    types: [created]

jobs:
  gemini-review:
    if: |
      github.event.issue.pull_request &&
      contains(github.event.comment.body, '@gemini-cli') &&
      (github.event.comment.author_association == 'OWNER' ||
       github.event.comment.author_association == 'MEMBER')
    uses: google-github-actions/run-gemini-cli@v0.1.22
```

### 4.4 GLM

**Two paths evaluated:**

**Path A: Custom workflow step invoking `moai glm` binary (Recommended)**

This is the cleanest path since `moai glm` already handles all compatibility env vars:

```yaml
- name: Setup GLM credentials
  env:
    GLM_AUTH_TOKEN: ${{ secrets.GLM_AUTH_TOKEN }}
  run: |
    mkdir -p ~/.moai
    cat > ~/.moai/.env.glm << EOF
    ANTHROPIC_BASE_URL=https://api.z.ai/v1
    ANTHROPIC_AUTH_TOKEN=${GLM_AUTH_TOKEN}
    DISABLE_BETAS=1
    DISABLE_PROMPT_CACHING=1
    EOF
    chmod 600 ~/.moai/.env.glm

- name: Run GLM code review
  run: |
    # moai glm reads ~/.moai/.env.glm and sets required env vars
    moai glm -- claude -p "Review this diff for Go best practices" \
      --model glm-4.7 \
      --max-turns 20
```

**Path B: anthropics/claude-code-action with overridden env (Advanced)**

```yaml
- name: Run GLM via Claude Code Action (env override)
  uses: anthropics/claude-code-action@<SHA>
  env:
    ANTHROPIC_BASE_URL: https://api.z.ai/v1
    ANTHROPIC_AUTH_TOKEN: ${{ secrets.GLM_AUTH_TOKEN }}
    CLAUDE_CODE_DISABLE_EXPERIMENTAL_BETAS: "1"
    DISABLE_PROMPT_CACHING: "1"
  with:
    # Do NOT use claude_code_oauth_token with GLM — API key mode only
    anthropic_api_key: ${{ secrets.GLM_AUTH_TOKEN }}
    claude_args: |
      --model glm-4.7
      --max-turns 20
```

**Compatibility env vars required [PROJECT MEMORY - glm_compatibility.md]:**

```
CLAUDE_CODE_DISABLE_EXPERIMENTAL_BETAS=1  # strips beta headers that GLM rejects
DISABLE_PROMPT_CACHING=1                   # strips prompt-caching-scope headers
```

Both are automatically injected by `moai glm` (SPEC-GLM-001). For Path B, inject manually as shown above.

**Token storage:**
- Create secret: `gh secret set GLM_AUTH_TOKEN --body "<z.ai-api-token>"`
- The runner-side `~/.moai/.env.glm` file is persistent but the secret provides the authoritative source

**Free GLM models for low-cost review:**
- `GLM-4.7-Flash` and `GLM-4.5-Flash` are genuinely free [FETCHED from docs.z.ai]
- For production-quality review: `GLM-4.7` at $0.38/$1.74 per M tokens is very cost-effective

---

## Section 5 — Integration Architecture

### Architecture A: Comment-Driven Multi-LLM Router (Recommended)

**Concept:** A single dispatcher workflow receives `@<llm-name>` comments and routes to the appropriate LLM.

**Trigger model:** PR and issue comment-driven.

**Routing rules:**

| Trigger Comment | LLM | Task Focus |
|-----------------|-----|------------|
| `@claude review` (existing) | Claude Opus 4.7 | Deep code review, SPEC compliance, security |
| `@gemini-cli /review` | Gemini 2.5 Flash | General review, second opinion, security triage |
| `@glm review` | GLM-4.7 | Cost-optimized bulk review, formatting checks |
| Automatic (PR open) | Claude Opus 4.7 | Primary review (existing claude-code-review.yml) |

**Workflow YAML skeleton:**

```yaml
name: Multi-LLM Router

on:
  issue_comment:
    types: [created]

concurrency:
  group: multi-llm-router-${{ github.event.issue.number }}-${{ github.event.comment.id }}
  cancel-in-progress: false

jobs:
  # Route to Gemini
  gemini-review:
    if: |
      github.event.issue.pull_request &&
      contains(github.event.comment.body, '@gemini-cli') &&
      (github.event.comment.author_association == 'OWNER' ||
       github.event.comment.author_association == 'MEMBER')
    runs-on: ubuntu-latest   # Gemini works fine on GitHub-hosted
    timeout-minutes: 15
    permissions:
      contents: read
      pull-requests: write
    steps:
      - uses: actions/checkout@v5
      - uses: google-github-actions/run-gemini-cli@v0.1.22
        with:
          gemini_api_key: ${{ secrets.GEMINI_API_KEY }}
          prompt: "@gemini-cli /review — Review PR #${{ github.event.issue.number }} for security, performance, and Go best practices."

  # Route to GLM
  glm-review:
    if: |
      github.event.issue.pull_request &&
      contains(github.event.comment.body, '@glm') &&
      (github.event.comment.author_association == 'OWNER' ||
       github.event.comment.author_association == 'MEMBER')
    runs-on: [self-hosted, macOS, ARM64, moai-adk-go]
    timeout-minutes: 30
    permissions:
      contents: read
      pull-requests: write
    steps:
      - uses: actions/checkout@v5
      - name: Setup Go
        uses: actions/setup-go@v6
        with:
          go-version: "1.26"
          cache: true
      - name: Setup GLM credentials
        env:
          GLM_AUTH_TOKEN: ${{ secrets.GLM_AUTH_TOKEN }}
        run: |
          mkdir -p ~/.moai
          printf 'ANTHROPIC_BASE_URL=https://api.z.ai/v1\nANTHROPIC_AUTH_TOKEN=%s\nDISABLE_BETAS=1\nDISABLE_PROMPT_CACHING=1\n' \
            "$GLM_AUTH_TOKEN" > ~/.moai/.env.glm
          chmod 600 ~/.moai/.env.glm
      - name: GLM Review
        run: |
          moai glm -- claude -p "Review PR diff for Go anti-patterns and test coverage gaps" \
            --model glm-4.7 --max-turns 15
```

**Architecture A Concurrency control:**

Each LLM workflow uses independent concurrency groups scoped to PR number + comment ID, with `cancel-in-progress: false` to avoid cancelling in-flight LLM sessions.

**Output strategy:**
- Claude: Check Run (existing behavior via claude-code-action)
- Gemini: PR comment (via google-github-actions/run-gemini-cli)
- GLM: PR comment (via `gh pr comment` in workflow step)

---

### Architecture B: Parallel Automated LLM Panel (for thorough SPEC-level PRs)

**Concept:** On PRs touching `.moai/specs/` or key source paths, automatically run 2+ LLMs in parallel for broader coverage.

**Trigger:** `pull_request` with path filter for high-value changes.

```yaml
name: LLM Panel Review

on:
  pull_request:
    types: [opened, ready_for_review]
    paths:
      - 'internal/**'
      - '.moai/specs/**'
      - 'cmd/**'

concurrency:
  group: llm-panel-${{ github.event.pull_request.number }}
  cancel-in-progress: true

jobs:
  # Job 1: Claude (primary — already in claude-code-review.yml, keep separate)
  # Not duplicated here; claude-code-review.yml handles it

  # Job 2: Gemini second opinion (parallel, free tier)
  gemini-panel:
    runs-on: ubuntu-latest
    timeout-minutes: 15
    permissions:
      contents: read
      pull-requests: write
    steps:
      - uses: actions/checkout@v5
        with:
          fetch-depth: 2
      - name: Get PR diff
        id: diff
        run: |
          git diff HEAD~1..HEAD --stat > diff_stat.txt
          echo "Diff stat generated"
      - uses: google-github-actions/run-gemini-cli@v0.1.22
        with:
          gemini_api_key: ${{ secrets.GEMINI_API_KEY }}
          prompt: |
            You are a Go expert. Review the following PR for moai-adk-go.
            Focus on: security vulnerabilities, API contract changes, test coverage gaps.
            Be concise. Post findings as a PR comment.
```

**Architecture B** is more expensive (2 LLMs per PR) but provides cross-validation. Recommended only for PRs above a complexity threshold (e.g., `> 500 lines changed`).

---

## Section 6 — Security Analysis

### 6.1 Private Repo Enforcement

**MANDATORY [USER-VERIFIED]:** The Codex auth.json pattern must ONLY be used in private repositories. `auth.json` is equivalent to a password.

Current project status: Private repo (has `CODECOV_TOKEN`, `CLAUDE_CODE_OAUTH_TOKEN` secrets — private repo pattern) [INFERRED].

Branch protection enforcement (per CLAUDE.local.md §18.7):

```bash
# Verify branch protection is active
gh api repos/modu-ai/moai-adk/branches/main/protection | jq .
```

If the branch protection rule from §18.7 is not yet applied (it was pending admin action), apply it before enabling Codex OAuth workflow.

### 6.2 Secret Rotation Strategy

| Secret | Rotation Trigger | Rotation Method | Frequency |
|--------|-----------------|-----------------|-----------|
| `CLAUDE_CODE_OAUTH_TOKEN` | Subscription change / suspected compromise | `claude setup-token` re-run | When compromised or on subscription renewal |
| `CODEX_AUTH_JSON` | 401 errors from Codex, compromised runner | `codex login` on trusted machine + `gh secret set` | When compromised; check monthly |
| `GEMINI_API_KEY` | Suspected exposure / quota anomaly | Google AI Studio → Create new key → delete old | Every 6 months or on compromise |
| `GLM_AUTH_TOKEN` | Z.ai account changes | Z.ai dashboard → regenerate token | Every 6 months or on compromise |

### 6.3 Workspace Cleanup Between Jobs

For self-hosted runners, sensitive files can persist between job runs. Standard mitigation:

```yaml
# At start of each job targeting self-hosted runner
- name: Pre-job cleanup
  run: |
    rm -f "$GITHUB_WORKSPACE/../**/*.env" 2>/dev/null || true
    git clean -fdx 2>/dev/null || true

# actions/checkout already sets clean: true by default
# which removes untracked files from the workspace
```

For GLM credentials, do NOT delete `~/.moai/.env.glm` between runs if you want the token to persist. The token is re-created from `secrets.GLM_AUTH_TOKEN` on each run anyway, so cleaning it is safe.

### 6.4 Sudo Management — Codex `drop-sudo` [FETCHED]

The Codex `--safety-strategy drop-sudo` option removes sudo privileges before executing any AI-generated code. On macOS:

```bash
# What drop-sudo does on macOS:
# 1. Removes runner user from sudoers for the duration of the Codex session
# 2. Prevents privilege escalation via AI-generated shell commands
# 3. The runner user process itself still runs with normal user permissions

# If using a dedicated _runner user (recommended):
# - The _runner user should NOT be in the sudo group at all
# - drop-sudo becomes a no-op but provides defense-in-depth
```

For Windows self-hosted runners: only `unsafe` mode is available [USER-VERIFIED]. Avoid running Codex on Windows self-hosted runners.

### 6.5 Supply Chain Risk — Action SHA Pinning [INFERRED]

Current state: `claude.yml` and `claude-code-review.yml` already pin to SHA (good practice).
`ci.yml`, `release.yml`, etc. use floating tags (`@v5`, `@v6`, `@v7`) — acceptable for trusted publishers (GitHub Actions, GoReleaser) but higher risk for third-party actions.

For new LLM actions, always SHA-pin:

```bash
# Get SHA for a specific tag
gh api repos/google-github-actions/run-gemini-cli/git/refs/tags/v0.1.22 \
  --jq '.object.sha'
```

### 6.6 Audit Log

- GitHub Actions logs: `github.com/modu-ai/moai-adk/actions` — retained 90 days by default
- Self-hosted runner logs: `~/actions-runner/_diag/` on the macOS machine
- Upgrade logs: `~/actions-runner/_update/` for auto-update attempts
- GLM API usage: Z.ai dashboard
- Gemini API usage: Google AI Studio dashboard

---

## Section 7 — Cost Analysis

### Workload Assumptions

- 50 PRs/month, 4 review rounds each = **200 LLM invocations/month**
- Input: ~5,000 tokens/invocation (PR diff + context)
- Output: ~2,000 tokens/invocation (review comments)
- Total per invocation: 7,000 tokens

### Monthly Cost Per LLM

**Claude Opus 4.7 (current, subscription OAuth):**
- Claude Max subscription: ~$100/month (includes CI quota)
- Current allocation: 200 invocations × 7K tokens = 1.4M tokens/month
- Verdict: Within Max plan quota for typical usage; overage risk at high PR volume
- Note [USER-VERIFIED]: Pro plan support NOT confirmed — must use Max plan

**Claude Opus 4.7 (API key mode, fallback):**
- Input: $15.00/M tokens × 1M input = $15.00
- Output: $75.00/M tokens × 0.4M output = $30.00
- Total: ~$45/month (API key mode, no subscription needed)

**Google Gemini 2.5 Flash (free tier):**
- Free tier RPD: ~500 requests/day [FETCHED] → 15,000/month
- 200 invocations/month << 500 RPD → **FREE**
- Paid overage: $0.15/M input + $0.60/M output (verify at AI Studio)

**GLM-4.7 (pay-per-token) [FETCHED from docs.z.ai]:**
- Input: $0.38/M × 1M = $0.38
- Output: $1.74/M × 0.4M = $0.70
- Total: ~**$1.08/month**

**GLM-4.7-Flash (free model) [FETCHED from docs.z.ai]:**
- Cost: $0.00/month (genuinely free, no trial limitations)
- Quality: lower than GLM-4.7 but suitable for formatting/simple checks

**OpenAI Codex (ChatGPT-managed subscription):**
- INCONCLUSIVE — Codex subscription pricing not publicly listed for CI/CD usage. Uses ChatGPT Pro/Team/Enterprise allocation. Estimate: included in ChatGPT Plus ($20/month) or Team ($25/user/month). For 200 invocations/month this is within typical quota but verify at `openai.com/codex`.

**Self-Hosted Runner (macOS arm64, GOOS' machine):**
- Machine cost: $0 (existing hardware)
- Electricity: ~$0.05/hour × assume 4 hours active/month = **~$0.20/month**

**GitHub-hosted macOS arm64 runner (for comparison):**
- Rate: $0.16/min (macOS arm64 is premium tier at 5x Linux rate: Linux $0.008 × 5 = $0.04/min... INCONCLUSIVE — verify exact macOS arm64 rate at github.com/billing/plans)
- 200 invocations × average 5 minutes = 1,000 minutes × $0.08/min (user's estimate) = **$80/month**

**CodeRabbit Pro (existing alternative) [FETCHED]:**
- Monthly billing: **$30/seat/month**
- Annual billing: **$24/seat/month**
- For 1 developer (GOOS): $24–$30/month

### 12-Month TCO Comparison Table

| Scenario | Month 1 Setup | Monthly Ongoing | 12-Month TCO | Notes |
|----------|--------------|-----------------|--------------|-------|
| **Current: Claude only (Max sub)** | $0 | ~$100 | ~$1,200 | Max plan subscription; no self-hosted |
| **Claude Max + Gemini free** | $0 | ~$100 | ~$1,200 | No additional LLM cost; Gemini free tier |
| **Claude Max + GLM-4.7** | $0 | ~$101 | ~$1,212 | Negligible GLM cost |
| **Self-hosted + Claude Max + Gemini + GLM** | $0 hardware | ~$101 | ~$1,212 | Full multi-LLM; self-hosted saves compute cost |
| **CodeRabbit Pro (replace Claude)** | $0 | $30 | $360 | Lower cost, less control, SaaS dependency |
| **API key Claude Opus 4.7 only (no Max)** | $0 | ~$45 | ~$540 | No subscription; pay-per-use |
| **GitHub-hosted macOS runners for LLM** | $0 | ~$80 runners + $100 Max | ~$2,160 | Expensive; macOS runner premium |

**TCO Summary:** Multi-LLM on self-hosted (Claude Max + Gemini free + GLM) has the same TCO as the current single-Claude setup (~$1,212/year) because the Max subscription dominates cost and Gemini/GLM add negligible marginal cost.

---

## Section 8 — Risk Register

| # | Risk | Severity | Probability | Mitigation |
|---|------|----------|-------------|------------|
| R1 | Codex auth.json corruption → 401 errors → CI breakage | P1 | Medium | Weekly health-check workflow; manual re-seed procedure documented in Section 4.2 |
| R2 | Claude Code Action SDK regression (type: #1126) breaks self-hosted | P1 | Medium (known pattern) | SHA-pin to v1.0.72 or v1.0.85+ for self-hosted; maintain separate SHA for GitHub-hosted. Test on self-hosted before deploying. |
| R3 | Runner version drift past 30-day deprecation window | P1 | High (if no calendar reminder set) | Monthly calendar reminder on 1st of month; `~/actions-runner/run.sh --version` check in monitoring |
| R4 | Single-machine SPOF — GOOS' macOS is the only self-hosted runner | P2 | High (unavoidable with 1 machine) | Keep GitHub-hosted fallback for non-LLM CI jobs. For LLM jobs, set `continue-on-error: true` on self-hosted steps with GitHub-hosted fallback job. |
| R5 | Network egress blocked — home network firewall blocks one of 5 LLM endpoints | P2 | Low | Test each endpoint before deployment: `curl -v https://api.anthropic.com/v1/messages` etc. |
| R6 | Subscription account suspension cascades to CI breakage | P2 | Low | Monitor subscription status; keep API key fallback (`ANTHROPIC_API_KEY`) as emergency credential |
| R7 | Public contributor attempts to abuse self-hosted runner | P0 | Very Low (private repo) | Private repo enforcement on all self-hosted workflows. `if: github.event.repository.private == true` guard. |
| R8 | Codex auth.json used concurrently in parallel jobs → token corruption | P1 | Medium (if parallel jobs enabled) | Never run concurrent Codex jobs sharing the same auth.json path. Serialize Codex jobs per `concurrency:` group. |
| R9 | Gemini free tier quota exhausted by external abuse / rate spike | P2 | Low | Add billing alert in Google AI Studio. Fallback: reduce Gemini invocations, use paid tier at $0.15/M input. |
| R10 | GLM API endpoint unavailability (Z.ai maintenance) | P3 | Low | GLM is optional/supplementary. `continue-on-error: true` on GLM jobs. Claude remains primary. |
| R11 | CLAUDE_CODE_OAUTH_TOKEN expiry breaks existing claude.yml + claude-code-review.yml | P1 | Low (tokens long-lived) | Monitor for 401 in workflow logs; `claude setup-token` rotation procedure in Section 4.1 |

---

## Section 9 — Implementation Plan

### Priority P1: Setup Runner + Verify Registration

**Prerequisite:** macOS arm64 machine available, GitHub admin access.

**Deliverable:** Runner registered, visible in `Settings → Actions → Runners`, status "Idle".

**Steps:**
1. Download v2.334.0 binary (Section 3.2)
2. Register with labels `[self-hosted, macOS, ARM64, moai-adk-go]` (Section 3.3)
3. Install as launchd service (Section 3.4)
4. Verify: `gh api repos/modu-ai/moai-adk/actions/runners | jq '.runners[] | {name, status, labels}'`

**Acceptance test:**

```yaml
# Create .github/workflows/runner-smoke-test.yml temporarily
name: Runner Smoke Test
on: workflow_dispatch
jobs:
  test:
    runs-on: [self-hosted, macOS, ARM64, moai-adk-go]
    steps:
      - run: echo "Runner $(hostname) is working" && uname -a && go version
```

**Rollback:** `./svc.sh stop && ./config.sh remove --token <removal-token>`

---

### Priority P1: Migrate claude-code-review.yml to Self-Hosted + Fix Version Pin

**Prerequisite:** Runner successfully registered (above).

**Deliverable:** `claude-code-review.yml` running on self-hosted, review working on test PR.

**Steps:**
1. Change `runs-on: ubuntu-latest` → `runs-on: [self-hosted, macOS, ARM64, moai-adk-go]`
2. Update pinned SHA from v1.0.82 to v1.0.72 (safe for self-hosted) OR v1.0.107 (verify first)

   ```bash
   # Get SHA for v1.0.72
   gh api repos/anthropics/claude-code-action/git/refs/tags/v1.0.72 \
     --jq '.object.sha'
   ```

3. Add workspace cleanup step before Claude action (issue #983 mitigation)
4. Test on a draft PR

**Acceptance test:** Claude posts review comment on test PR within 5 minutes.

**Rollback:** Revert `runs-on` to `ubuntu-latest`. Zero downtime — previous SHA still available.

---

### Priority P2: Add Gemini CLI Workflow

**Prerequisite:** `GEMINI_API_KEY` secret created in repo.

**Deliverable:** `gemini-review.yml` deployed, `@gemini-cli /review` works on test PR.

**Steps:**
1. Get API key from Google AI Studio
2. `gh secret set GEMINI_API_KEY --body "<key>"`
3. Create `.github/workflows/gemini-review.yml` using Architecture A skeleton
4. Test by posting `@gemini-cli /review` on a test PR

**Acceptance test:** Gemini posts review within 3 minutes of comment.

**Rollback:** Delete `gemini-review.yml`. No secrets to remove urgently (key can be revoked in AI Studio).

---

### Priority P2: Add GLM Workflow

**Prerequisite:** Z.ai API token, `moai` binary available on self-hosted runner.

**Deliverable:** `glm-review.yml` deployed, `@glm review` triggers GLM response.

**Steps:**
1. Get Z.ai API token from `https://api.z.ai`
2. `gh secret set GLM_AUTH_TOKEN --body "<token>"`
3. Verify `moai` binary is on PATH on self-hosted runner: `which moai`
4. Create `.github/workflows/glm-review.yml`
5. Test on a draft PR

**Acceptance test:** GLM posts review within 5 minutes of trigger.

**Rollback:** Delete `glm-review.yml`. Revoke token in Z.ai dashboard if needed.

---

### Priority P2: Codex OAuth Bootstrap (Optional — Evaluate After Gemini/GLM)

**Prerequisite:** OpenAI subscription with ChatGPT-managed auth, macOS self-hosted runner.

**Deliverable:** `codex-review.yml` with auth.json bootstrap pattern, `codex --version` works on runner.

**Steps:**
1. `npm install -g @openai/codex` on self-hosted runner
2. `codex login` on local machine → save `auth.json`
3. `gh secret set CODEX_AUTH_JSON --body "$(cat ~/.codex/auth.json)"`
4. Create workflow with conditional bootstrap (Section 4.2)
5. Private repo guard: `if: github.event.repository.private == true`

**Acceptance test:** Codex runs without 401 errors on a test PR. `auth.json` persists across two consecutive runs.

**Rollback:** Delete `codex-review.yml`. The auth.json on runner can stay (harmless); revoke via OpenAI dashboard.

---

### Priority P3: Cost Dashboard / Observability

**Prerequisite:** All LLM workflows deployed.

**Deliverable:** Monthly cost summary posted to a GitHub Issue or Slack channel.

**Steps:**
1. Z.ai dashboard: enable usage notifications
2. Google AI Studio: enable quota alerts
3. GitHub Actions: enable billing alerts at `Settings → Billing → Spend limits`
4. Optional: Create a monthly `workflow_dispatch` job that posts a cost summary comment

**Acceptance test:** Receive at least one cost alert/summary without manual intervention.

---

## Section 10 — Final Recommendation

**Verdict: GO-WITH-CAVEATS**

The multi-LLM self-hosted CI/CD integration is technically feasible and economically attractive for moai-adk-go. The dominant cost driver is the Claude Max subscription (already in use) — adding Gemini and GLM costs less than $2/month extra. However, three caveats are hard requirements:

**Caveats:**

1. **Private repo only for Codex OAuth.** OpenAI explicitly prohibits the auth.json pattern for public or open-source repositories. moai-adk-go appears to be private, but enforce this with the `if: github.event.repository.private == true` guard in all Codex workflows.

2. **Monthly runner upgrade discipline required.** The self-hosted runner on GOOS' macOS machine must be manually upgraded within 30 days of each new runner release. There is no automatic enforcement mechanism — this requires a calendar reminder and consistent execution.

3. **Version pin must be updated before self-hosted migration.** The current SHA `5d5c10a4f389689f992ea10bb14dcb6fcc83146d` (v1.0.82) falls in the known-broken range for self-hosted runners (v1.0.79–v1.0.84, issue #1126). Re-pin to v1.0.72 or v1.0.85+ before migrating `claude-code-review.yml` to self-hosted.

**Recommended starting set:** Claude Code Action (already deployed, just migrate runner) + Gemini free tier (zero marginal cost, comment-driven). Codex and GLM are secondary additions once the runner is stable.

**SPEC recommendation: YES — Create SPEC-CI-MULTI-LLM-001**

Rationale: This work involves 4+ new workflow files, runner infrastructure setup, 4 new secrets, and operational runbooks. It directly maps to the SPEC-First DDD workflow requirement of "tracked implementation with acceptance criteria." The implementation plan (Section 9) naturally becomes the SPEC task list. Without a SPEC, this work is untracked and the operational decisions (version pins, rotation strategy, rollback plans) will not be captured in a structured way. SPEC scope: 8 tasks (P1: runner setup + runner smoke test + claude migration; P2: Gemini + GLM + Codex + secrets; P3: cost dashboard), complexity level: `standard` harness.

---

## Sources

All URLs listed below were actually fetched or searched during this session. No URLs were generated from training data.

- [OpenAI Codex CI/CD Auth Guide](https://developers.openai.com/codex/auth/ci-cd-auth) — Auth.json bootstrap pattern [FETCHED]
- [GitHub Self-Hosted Runner v2.329.0 Enforcement Paused](https://github.blog/changelog/2026-03-13-self-hosted-runner-minimum-version-enforcement-paused/) — Runner version policy [FETCHED]
- [Google Gemini CLI GitHub Action (google-github-actions/run-gemini-cli)](https://github.com/google-github-actions/run-gemini-cli) — v0.1.22, auth methods, trigger syntax [FETCHED]
- [Anthropic Claude Code Action (anthropics/claude-code-action)](https://github.com/anthropics/claude-code-action) — Current version, OAuth setup [FETCHED]
- [GitHub Actions Runner Latest Release (v2.334.0)](https://api.github.com/repos/actions/runner/releases/latest) — Download URL for macOS arm64 [FETCHED]
- [Google AI Studio / Gemini API Rate Limits 2026](https://ai.google.dev/gemini-api/docs/rate-limits) — Free tier RPM/RPD [SEARCHED]
- [Z.ai GLM Pricing](https://docs.z.ai/guides/overview/pricing) — GLM model pricing table [FETCHED]
- [CodeRabbit Pricing 2026](https://www.coderabbit.ai/pricing) — $24–$30/seat/month [SEARCHED]
- [Gemini CLI Quota and Pricing](https://geminicli.com/docs/resources/quota-and-pricing/) — Free tier limits for CI/CD context [FETCHED]

---

*Report generated by expert-devops agent (MoAI-ADK) on 2026-04-27.*
*Methodology: Sequential Thinking MCP + parallel WebFetch/WebSearch verification + local file reads.*
*Decision-critical facts are tagged [USER-VERIFIED], [FETCHED], or [INFERRED].*

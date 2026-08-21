# t170 — Lens: masking, secret detection, path sanitization

Read-only investigation of what already exists in `moai-adk-go` for the card's three
[HARD] security clauses. Every claim below is quoted from source read in this worktree
(`/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t170`).

**Headline: there is no text scrubber in this repo.** Zero hits for a `Redact*` function
anywhere in `internal/`, `pkg/`, `cmd/`. Everything that exists is either (a) a *display*
masker for one already-isolated value, (b) a *detector* that denies rather than rewrites,
or (c) an env-var *dropper*. None of the three rewrites a body of text before it leaves
the process. The pattern sets, however, are reusable and are the real find.

---

## 1. Existing redaction / masking / sanitization code

Search: `func .*(Redact|Mask|Sanitiz|Scrub)` over `--include="*.go" internal/ pkg/ cmd/`,
excluding `_test.go`. Eight hits, all read:

| Location | What it does | Reusable for the card? |
|---|---|---|
| `internal/github/secret.go:144` `MaskSecret` | value-level, first char + last 4 | display only |
| `internal/cli/glm.go:454` `maskAPIKey` | value-level, 4+4 | display only |
| `internal/cli/glm_tools.go:992` `maskPartial` | value-level, first 4 | display only |
| `internal/sandbox/env.go:51` `ScrubEnv` | drops env vars from a child env | **partially** — see below |
| `internal/runtime/audit_cache.go:215` `SanitizeMarkdown` | escapes markdown metachars | no (different concern) |
| `internal/harness/applier.go:803` `sanitizeSurfaceToken` | filename safety | no |
| `internal/astgrep/sarif.go:244` `sanitizeTagValue` | SARIF tag safety | no |
| `internal/hook/worktree_create.go:183` `sanitizeWorktreeBranchSuffix` | branch-name safety | no |

### The three value-level maskers — all identical in shape, all display-only

```go
// internal/github/secret.go:144
func MaskSecret(value string) string {
	if len(value) <= 4 {
		return "***"
	}
	return value[:1] + "..." + value[len(value)-4:]
}
```

```go
// internal/cli/glm.go:454
func maskAPIKey(key string) string {
	if len(key) <= 8 {
		return "****"
	}
	return key[:4] + "****" + key[len(key)-4:]
}
```

```go
// internal/cli/glm_tools.go:992
func maskPartial(token string) string {
	if len(token) <= 8 {
		return "****"
	}
	return token[:4] + "****"
}
```

Critical limitation for the card: **each takes a value the caller has already isolated.**
`MaskSecret` is called at `internal/github/secret.go:88` on a value the caller passed in;
`maskAPIKey` at `glm.go:449` on the key the user typed. None of them *finds* a secret
inside arbitrary text. The card needs the finding step, which is where §3's pattern sets
come in. There are three near-duplicate implementations with three different output
shapes (`***` / `****` / `x...yyyy`), so a scrubber that adopts one of them should
consolidate rather than add a fourth.

### `ScrubEnv` — the one genuinely reusable mechanism, for env values only

```go
// internal/sandbox/env.go:31-37
var defaultDenyList = []string{
	"GITHUB_TOKEN",
	config.EnvAnthropicAPIKey,
	"OPENAI_API_KEY",
	"NPM_TOKEN",
	"GH_TOKEN",
}
```

Rules, from the doc comment at `internal/sandbox/env.go:41-48`: passthrough wins, then
`AWS_` prefix removed, then denylist removed, else keep. Config-extensible via
`security.sandbox.env_scrub_extra` (declared in `.moai/config/sections/security.yaml`).

It **removes** variables from a `[]string` of `KEY=VALUE` pairs; it does not mask a value
found inside prose. For the card's "mask env values" requirement it supplies the
**denylist vocabulary** (which names are sensitive) but not the mechanism.

---

## 2. `.moai/config/sections/security.yaml`

Full current contents (995 bytes):

```yaml
security:
    extra_ask_patterns: []
    extra_dangerous_bash_patterns:
        - curl\s+.*\|\s*(ba)?sh
        - wget\s+.*\|\s*(ba)?sh
        - chmod\s+777
        - rm\s+-rf\s+/[^.]
        - mkfs\.
        - dd\s+if=.*of=/dev/
    permission:
        strict_mode: false
        pre_allowlist: []
        session_rules: []
    sandbox:
        required: false
        network_allowlist: []
        env_scrub_extra: []
        docker_image: "alpine:latest"
    extra_deny_patterns: []
    extra_sensitive_content_patterns: []
```

Every key is an **`extra_*` extension point** — the built-in defaults live in Go and the
YAML only appends. Schema: `internal/hook/security/config.go:18`
(`ExtraSensitiveContentPatterns []string` with tag `yaml:"extra_sensitive_content_patterns"`).
Merge semantics: `internal/hook/pre_tool.go:300-318` `MergeExtraPatterns`, whose comment
states *"Extra patterns extend (never replace) the built-in defaults per SOLID
O-Principle (REQ-SEC-003, REQ-SEC-004)"*.

**Yes — a security-pattern list is already declared in Go**, and it is the most important
find on this lens. `internal/hook/pre_tool.go:262-273`:

```go
	// Content patterns that indicate sensitive data
	sensitiveContentPatterns := []string{
		`-----BEGIN\s+(RSA\s+)?PRIVATE\s+KEY-----`,
		`-----BEGIN\s+CERTIFICATE-----`,
		`sk-[a-zA-Z0-9]{32,}`,       // OpenAI API keys
		`ghp_[a-zA-Z0-9]{36}`,       // GitHub tokens
		`gho_[a-zA-Z0-9]{36}`,       // GitHub OAuth tokens
		`glpat-[a-zA-Z0-9\-]{20}`,   // GitLab tokens
		`xox[baprs]-[a-zA-Z0-9\-]+`, // Slack tokens
		`AKIA[0-9A-Z]{16}`,          // AWS access keys
		`ya29\.[a-zA-Z0-9_\-]+`,     // Google OAuth tokens
	}
```

Compiled case-insensitively by `compilePatterns` (`pre_tool.go:294`:
`regexp.Compile("(?i)" + p)`), which **skips** an uncompilable pattern with a `slog.Warn`
rather than failing — a fail-open posture worth matching.

How they are used today (`pre_tool.go:895-910`) — **deny, not rewrite**:

```go
	if toolName == "Write" || toolName == "Edit" {
		contentField := "content"
		if toolName == "Edit" {
			contentField = "new_string"
		}
		content, ok := parsed[contentField].(string)
		if ok && content != "" {
			for _, pattern := range h.policy.SensitiveContentPatterns {
				if pattern.MatchString(content) {
					return DecisionDeny, "Content contains sensitive data (credentials, API keys, or certificates)"
				}
			}
		}
	}
```

`MatchString` — a boolean. To mask, the card needs `ReplaceAllString` /
`ReplaceAllStringFunc` over the same compiled set. The `SecurityPolicy` struct
(`pre_tool.go:73-98`) is exported and `DefaultSecurityPolicy()` (`pre_tool.go:210`) is
exported, so `hook.DefaultSecurityPolicy().SensitiveContentPatterns` is directly
importable as the scrubber's pattern set **today, with no refactor**. That also means a
project's `extra_sensitive_content_patterns` extends the scrubber for free.

Also present and relevant to clause (b): `denyPatterns` at `pre_tool.go:216-230` names the
*file* classes that are secret-bearing — `secrets?\.(json|ya?ml|toml)$`, `\.ssh/.*`,
`id_rsa.*`, `id_ed25519.*`, `\.pem$`, `\.key$`, `\.crt$`.

---

## 3. Existing secret-pattern detection elsewhere

### ast-grep ruleset — present in this worktree at `.moai/astgrep-rules/security/`

(Note the path: `.moai/astgrep-rules/`, **not** `.moai/config/astgrep-rules/` — moved per
`CLAUDE.local.md` §2.3 because `moai update` wipes `.moai/config` wholesale. Files:
`secrets.yml`, `credentials.yml`, `injection.yml`, `web.yml`, plus `crypto.yml` in the
primary checkout.)

`.moai/astgrep-rules/security/credentials.yml` carries the **densest single pattern** in
the repo, repeated per language (go / python / javascript / typescript):

```yaml
rule:
  kind: interpreted_string_literal
  regex: "^\"(sk-|AKIA[0-9A-Z]{16}|ghp_[0-9A-Za-z]{36}|xox[baprs]-|AIza[0-9A-Za-z_-]{35})"
```

Its header comment states the design rationale, which the card should inherit:
*"The prefix + minimum-length anchoring keeps the match precise: a benign string that
merely contains these characters mid-value is not flagged."*

It adds **`AIza[0-9A-Za-z_-]{35}`** (Google API key) which the Go `sensitiveContentPatterns`
list does **not** have. Conversely the Go list has PEM headers, `gho_`, `glpat-`, and
`ya29.` which the ast-grep rule lacks. **Neither set is a superset of the other** — the
card's scrubber should take the union.

`.moai/astgrep-rules/security/secrets.yml` is narrower and Go-AST-shaped (not text
regex), so it is not a pattern source:

```yaml
id: sec-hardcoded-api-key
rule:
  pattern: const $NAME = "sk-$$$REST"
---
id: sec-hardcoded-jwt-signing-key
rule:
  pattern: SignedString([]byte("$HARDCODED"))
```

### Pre-commit hook

`lefthook.yml` exists (210 bytes) and `internal/cli/hook_install.go` installs a git hook,
but **no secret-scanning step**: I found no `gitleaks`, `trufflehog`, or equivalent
anywhere. The gate is `moai gate` (lint+format+type+test).

### CI guard

`.github/workflows/template-neutrality-check.yaml` guards *template neutrality* (SPEC IDs,
internal dates, commit SHAs, macOS-bias paths) — a content-class guard, not a secret
scanner. Sibling `internal/template/internal_content_leak_test.go` owns dates + commit
hashes. **Interesting as a precedent for clause (b)**: it is exactly the shape of
"classify content and refuse to ship it", and it already includes an
"absolute-home-path" forbidden class (C6 macOS-bias paths per
`.moai/docs/template-internal-isolation-doctrine.md §25.1`).

### Policy precedent for clause (b)

`SECURITY.md` already carries the human-facing form of the rule the card wants mechanized:

> 1. **Do NOT** open a public GitHub issue for security vulnerabilities.
> 2. Email the security report to the maintainers via GitHub Security Advisories

So clause (b) is not a new policy — it is the mechanization of an existing documented one,
and its refusal message should point at the GitHub Security Advisories flow named there.

---

## 4. Home / absolute path handling

**There is no `~`-collapsing helper. This is a real gap.**

`internal/paths/paths.go` (the SSOT, 5.1 KB) resolves *outward* only — it builds absolute
paths from `$HOME`, never shortens one for display:

```go
// internal/paths/paths.go:49-57
// Home returns $HOME when non-empty, else os.UserHomeDir() (REQ-MHP-002).
// @MX:ANCHOR: [AUTO] home-resolution SSOT — HOME-first contract for every ~/.moai consumer
func Home() (string, error) {
	if h := os.Getenv("HOME"); h != "" {
		return h, nil
	}
	return os.UserHomeDir()
}
```

plus `MoaiHome()`, `StateDir()`, `CacheDir()`, `ReleasesDir()`, `WorktreesDir()`,
`ProfilesDir()`, `GlmEnvFile()`, `UserSettingsFile()`, `UserConfigSectionsDir()` — all
constructors.

A repo-wide grep for tilde substitution (`strings.Replace.*home`, `"~"`) returns exactly
two hits, neither a display helper:

- `internal/core/git/branch.go:179` — `strings.Contains(name, "~")`, a git refname
  validity check.
- `internal/shell/detect.go:135` — `home = "~"` as a *last-resort fallback* when both
  `$HOME` and `$USERPROFILE` are empty, then joined into a config-file path. The opposite
  direction.

Statusline does not collapse either: `internal/statusline/builder.go:47-49` carries a
`HomeDir` field used only for the usage cache location.

**Verdict: a `CollapseHome(path string) string` (or a `SanitizePaths(text string) string`
that rewrites every `$HOME`-prefixed absolute path to `~/…`) must be written new.** It
should read `$HOME` through `paths.Home()` so an overridden `HOME` is honored in lockstep
with every other consumer — that contract is explicitly asserted at `paths.go:8`.

Note the harder half: the card says mask absolute home paths *in submitted text*, which
means finding them inside prose, not converting one known path. That is at minimum
`strings.ReplaceAll(text, home, "~")`, and the same for the project root and the worktree
path if those are also to be hidden.

---

## 5. Local queuing precedent

Three candidates, in descending order of fit.

### (a) `internal/kanban.BacklogStore` — the strongest precedent

File: `.moai/state/kanban/backlog.json`, path built by `internal/cli/todo.go:42-44`:

```go
func todoBacklogPath(root string) string {
	return filepath.Join(root, ".moai", "state", "kanban", "backlog.json")
}
```

Shape (`internal/kanban/backlog_store.go`): a `BacklogStore` over one JSON file with a
sibling lock file, a lock-free `Load()` (:152), a locked read-modify-write `Mutate()`
(:190) that is *the* guarded mutation entry point, and a same-directory temp + atomic
rename write (:290 `writeAtomic`, delegating to `atomicfile.Replace`):

```go
	tmp, err := os.CreateTemp(dir, ".backlog-*.tmp")
	…
	if err := atomicfile.Replace(tmpName, s.path); err != nil {
```

This is the right convention for a **retryable queue with mutation** (enqueue on send
failure, dequeue on later success). It is cross-process safe
(`internal/kanban/board_lock.go`) and already has `internal/atomicfile` (`Replace`,
`ReadFile`, `Claim`) underneath.

### (b) `.moai/lessons-inbox.jsonl` — the strongest precedent for *append-only*

`internal/hook/failure_observer.go:128-156`:

```go
func appendLessonsInboxStub(root, eventKey, summary, source string) {
	inboxPath := filepath.Join(root, ".moai", "lessons-inbox.jsonl")
	stub := lessonsInboxStub{
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		EventKey:  eventKey,
		Summary:   summary,
		Source:    source,
	}
	…
	f, err := os.OpenFile(inboxPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o600)
```

Doc comment states the three properties the card should copy verbatim:
*"append-only JSONL (EC-4: concurrent Stop hooks tolerate interleaving; no
read-modify-write). Permissions 0o600 are consistent with sibling state files. Fail-open:
errors are logged and swallowed (a learning-loop write must never block the session)."*

It also carries a bounded-summary helper (`truncateSummary`, :159, 200 runes + `…`) —
directly reusable shape for bounding a queued payload.

### (c) `.moai/state/handoff/pending.json` — the "claim-and-consume" precedent

`internal/cli/handoff.go:159` `saveHandoff` → `handoff.SavePending`; consumption at
`internal/hook/handoff_inject.go:116` does a **claim-rename** `pending.json →
consumed/<ts>-<nonce>.json` where *"Exactly one concurrent caller gets true"* (:252), plus
a staleness TTL (`internal/config/defaults.go:285` `DefaultHandoffStaleTTL`).

If the card's queue needs "send it once, then archive it, never twice", this is the
precedent — the claim-rename + `consumed/` audit dir is exactly a send-once queue.

**Recommendation for clause (c):** append-only JSONL at `.moai/state/` following (b)'s
`0o600` + fail-open + mkdir-parent contract for the *masking log*, and (a)'s locked
`BacklogStore` shape (or (c)'s claim-rename) for the *retry queue*, since a retry queue
needs deletion on success and a bare append-only file does not.

---

## 6. Logging precedent for a locally-persisted audit trail

Two established shapes; the repo uses both.

### Line-oriented `.moai/logs/*.log` — the closest analogue to a "masking log"

`internal/config/log.go:21-65` is the cleanest complete example:

```go
func logTierReadFailure(source Source, path string, err error) {
	const logDir = ".moai/logs"
	const logFile = ".moai/logs/config.log"

	if mkErr := os.MkdirAll(logDir, 0o755); mkErr != nil { … return }
	f, openErr := os.OpenFile(filepath.Clean(logFile), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	…
	entry := fmt.Sprintf("%s | source=%s | path=%s | err=%v\n",
		time.Now().Format(time.RFC3339), source.String(), path, err)
```

Doc comment: *"best-effort logger — if the log file cannot be opened or written, the error
falls back to slog.Warn on stderr and never panics."* Note `0o644` here vs `0o600` for the
lessons inbox; **a masking log records what a secret looked like enough to have been
masked, so it should take `0o600`.**

Siblings in the same convention, all `.moai/logs/`:

- `.moai/logs/status-transition-audit.log` — written by a hook, **parsed back** by
  `internal/spec/audit_transition.go:135`. Its parse contract is the one to copy for any
  log meant to be read again: *"an absent, corrupt, or unparseable log degrades to zero
  findings and NEVER returns an error"*, with unparseable lines counted
  (`SkippedUnparseable`) rather than fatal.
- `.moai/logs/lifecycle-close.log` (`internal/spec/closer.go:537`)
- `.moai/logs/autonomy-downgrade.log` (`internal/core/project/autonomy_bundle.go:73`)
- `.moai/logs/branch-guard-audit.log` (per `main-checkout-branch-guard.md`)
- `.moai/logs/config.log`, `.moai/logs/navigator-sync.log` (`internal/cli/navigator_route.go:13`)

### JSONL ledgers

`.moai/observability/hook-metrics.jsonl` (declared in
`.moai/config/sections/observability.yaml`), `.moai/logs/trace-*.jsonl`
(`internal/config/defaults.go:180`; pruned at SessionEnd), `.moai/logs/update-cleanup-{ts}.jsonl`
(`internal/cli/update_cleanup.go:414`), `.moai/lessons-inbox.jsonl`,
`.moai/logs/agent-model-audit.jsonl`.

Two things that bear on the card:

- `internal/cli/update_cleanup.go:133`, `internal/worktree/state_guard.go:54`, and
  `internal/cli/codex_review_gate.go:39` all **exclude `.moai/logs/`** from their sweeps.
  A masking log placed there inherits that protection.
- `internal/telemetry/` writes hook traces and contains **zero** redaction —
  `grep -niE "redact|mask|sanitiz|scrub|secret|token" internal/telemetry/*.go` returns
  nothing. The existing trace/log surface is itself unmasked; the card's scrubber is a
  candidate consumer for it later, but that is out of scope here.

---

## Verdict per [HARD] clause

### (a) Mask tokens / secret values / absolute home paths / env values before submission

**Extend + write new — a split verdict, and the split matters.**

| Sub-requirement | Verdict | Name |
|---|---|---|
| Secret **pattern set** | **REUSE, as-is** | `hook.DefaultSecurityPolicy().SensitiveContentPatterns` (`internal/hook/pre_tool.go:262-273`) — exported, case-insensitive, config-extensible via `extra_sensitive_content_patterns` |
| Pattern set **completeness** | **EXTEND** | union with `AIza[0-9A-Za-z_-]{35}` from `.moai/astgrep-rules/security/credentials.yml`, absent from the Go list |
| The **masking transform** | **WRITE NEW** | nothing in the repo rewrites text. `MatchString` → needs `ReplaceAllStringFunc`. The three value-level maskers supply the *output shape* only, and should be consolidated rather than four-peated |
| **Env values** | **EXTEND** | `sandbox.defaultDenyList` (`internal/sandbox/env.go:31-37`) + `security.sandbox.env_scrub_extra` give the name vocabulary; `ScrubEnv` itself removes vars from a `[]string` and cannot mask a value found in prose |
| **Absolute home paths** | **WRITE NEW** | confirmed gap — no `~`-collapsing helper exists anywhere (§4). Build it on `paths.Home()` so an overridden `HOME` is honored per the `paths.go:8` contract |

### (b) Classify security-vulnerability content and exclude it from automatic submission

**Write new — but the policy and one structural precedent already exist.**

- Nothing today classifies *report content* as vulnerability disclosure. Found nothing.
- `SECURITY.md` already states the human rule ("Do NOT open a public GitHub issue…";
  route to GitHub Security Advisories) — the refusal path should cite it, not invent one.
- Structural precedent for "classify content, refuse to ship": the template-neutrality
  guard (`.github/workflows/template-neutrality-check.yaml` +
  `internal/template/internal_content_leak_test.go`, content classes C1-C8 in
  `.moai/docs/template-internal-isolation-doctrine.md §25.1`). Same shape, different
  classifier.
- Weak signal reusable as one input: `denyPatterns` (`pre_tool.go:216-230`) already names
  secret-bearing *file* classes (`\.pem$`, `\.key$`, `\.ssh/.*`, `id_rsa.*`), so a report
  quoting those paths is a candidate flag.
- **Gap named plainly:** there is no vulnerability-content classifier, no CVE/CWE text
  detector, and no allow/deny vocabulary for it. This clause is the most net-new work on
  the card.

### (c) Keep a masking log locally + queue on send failure

**Reuse conventions; write the two artifacts.**

- **Masking log** — reuse the `.moai/logs/*.log` convention. Closest analogue:
  `internal/config/log.go` (mkdir 0o755 → `O_APPEND|O_CREATE|O_WRONLY` → RFC3339
  `|`-delimited entry → best-effort, `slog.Warn` on failure, never panics). Take `0o600`
  from `appendLessonsInboxStub` rather than config.log's `0o644`, since the log's subject
  is secret-adjacent. If the log will be read back, copy the graceful parse contract of
  `internal/spec/audit_transition.go` (absent/corrupt → zero findings, never an error).
  `.moai/logs/` is already excluded from three cleanup sweeps.
- **Retry queue** — reuse `internal/kanban.BacklogStore` (`internal/kanban/backlog_store.go`)
  as the shape: one JSON file under `.moai/state/<subsystem>/`, sibling lock, `Mutate()`
  read-modify-write, `writeAtomic` via `internal/atomicfile.Replace`. A bare append-only
  JSONL is *not* sufficient for a retry queue, because a queue needs deletion on success.
  If send-once semantics are wanted instead, the handoff claim-rename
  (`internal/hook/handoff_inject.go:116,252` + `consumed/` archive +
  `DefaultHandoffStaleTTL`) is the exact precedent.
- **Gap named:** no existing subsystem queues a *failed network send* for retry.
  `internal/resilience/circuit.go` is a circuit breaker, not a queue; `internal/telemetry`
  has an async recorder but no on-disk failure spool. The queue is new code; only its
  conventions are borrowed.

---

## Where I found nothing (explicit)

1. No `Redact*` function, and no function that rewrites arbitrary text to remove secrets.
2. No `~`-collapsing / home-abbreviating path display helper.
3. No secret scanner in the pre-commit hook or in CI (no gitleaks/trufflehog).
4. No vulnerability-content classifier.
5. No redaction in `internal/telemetry` — the existing hook-trace surface is unmasked.
6. No on-disk retry spool for a failed outbound send anywhere in the repo.

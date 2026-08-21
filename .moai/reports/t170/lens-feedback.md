# t170 — lens: the existing `/moai feedback` path, end to end

Read-only measurement. Every claim below cites `file:line` and quotes the literal text read. Working dir: `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t170` (HEAD `4b2f203fe`).

## Summary

- The feedback path is **entirely prose** — a skill body the orchestrator follows. There is no Go code that files an issue; `gh issue create` appears only as an instruction in `.claude/skills/moai/workflows/feedback.md:118` and never with `--title` / `--body` / `--label` flags spelled out. Any `auto_submit` behavior is therefore an edit to prose plus (optionally) a config key that only prose reads.
- **The card's central assumption is false as stated.** There is no confirm-before-submit question today. The skill invokes `AskUserQuestion` exactly twice: once to *collect* type/title/description/priority (`feedback.md:52`) and once *after* submission (`feedback.md:156`). Nothing sits between the collection round and `gh issue create`. The en docs page states it outright: "Once you answer, a GitHub issue is created automatically" (`docs-site/content/en/utility-commands/moai-feedback.md:179`). So `auto_submit=true` would not skip a confirmation — it would have to skip the **data-collection** round, which is a different and much larger change (it must then synthesize title/description/type/priority without asking).
- The masking surface is small and already fenced: the skill has a [HARD] ban on attaching arbitrary file contents (`feedback.md:139`). The one genuinely unbounded text source is the orchestrator-passed "last-failed-command / error context" (`feedback.md:137`).
- Adding `feedback.auto_submit` to config touches **9 files minimum** (struct, default, accessor, template yaml + 2 non-template copies, key inventory, web settings schema, 4-locale i18n) because two anti-rot guards enforce registration: `TestShippedConfigKeysHaveReaders` (`internal/config/shipped_key_reader_test.go:70`) and the i18n-completeness check (`internal/web/schema_label_test.go:96`).
- Existing feedback config tests would **not** break — they only assert `Repository` (`internal/config/feedback_config_test.go`).

---

## 1. What the workflow does today, and what files the issue

**Command entry.** `.claude/commands/moai/feedback.md:1-7` is a 7-line shim:

```
allowed-tools: Skill
---
Use Skill("moai") with arguments: feedback $ARGUMENTS
```

Template twin at `internal/template/templates/.claude/commands/moai/feedback.md.tmpl:7` (same body; frontmatter localized via `{{if eq .ConversationLanguage ...}}`).

**Routing.** `.claude/skills/moai/SKILL.md:70` `- **feedback** (aliases: fb): GitHub issue creation`; `SKILL.md:240` `Agents: orchestrator-direct (records feedback via gh CLI)`; `SKILL.md:241` points at `workflows/feedback.md`.

**Step by step** (`.claude/skills/moai/workflows/feedback.md`):

| Step | Line | What happens |
|---|---|---|
| Precondition | 36 | `[HARD] Before issue creation, run `gh auth status`.` Fallback on unauth/rate-limit: report, guide, offer local draft at `.moai/state/feedback-draft-<timestamp>.md` (40) |
| Phase 1 Step 1 | 52 | `[HARD] Collect the feedback fields — type, title, and description — in ONE AskUserQuestion round` — plus priority as a 4th question (61) |
| Phase 1 Step 3 | 71 | Duplicate search: `gh issue list --repo <resolved-target> --search "<title keywords>" --state open` |
| — | 76 | `[HARD] The duplicate-detection step ... MUST NOT prompt the user inline; it only produces the candidate-report.` |
| Phase 2 | 82 | `[HARD] Create the GitHub issue orchestrator-direct via the `gh` CLI` |
| Phase 2 | 118 | `The orchestrator executes directly: `gh issue create --repo <resolved-target>`` |
| Result | 143-144 | `[HARD] Provide user with the created issue URL.` / `[HARD] Confirm successful feedback submission to user.` |
| Post | 156 | `Use AskUserQuestion after successful submission:` (Continue / Submit another / View issue) |

**Where is the user asked to confirm submission? Nowhere.** `grep -n AskUserQuestion .claude/skills/moai/workflows/feedback.md` → lines `52`, `156`, `178` only. Line 178 is the summary restating line 52. There is no gate between the candidate-report (76) and `gh issue create` (118) — line 76 explicitly says the duplicate step "MUST NOT prompt", and delegates the decision to "the orchestrator ... at its own level" without prescribing a question.

**The exact invocation quoted (feedback.md:118):**

```
The orchestrator executes directly: `gh issue create --repo <resolved-target>`, where `<resolved-target>` is the resolved feedback target repository (config `feedback.repository`, default `modu-ai/moai-adk`).
```

That is the whole thing. **Gap:** no flag set (`--title`, `--body`, `--label`, `--web`) is specified anywhere in the skill; the body template is described in prose (120-125) and the labels in prose (94-96). So there is no literal command string in the repo to modify — an implementer adding `auto_submit` has no existing invocation line to branch on and would be writing the flags for the first time.

**No MCP tool, no Go path.** `grep -rIn "gh issue create"` over the repo returns only: the 4 docs-site locale pages, this skill, and SPEC/history artifacts. No `internal/**` Go file constructs an issue. The `mcp__moai__*` catalogue (`.claude/rules/moai/core/moai-mcp-tools.md`) has no issue-creation tool.

**Note (mirror drift, minor):** `.claude/skills/moai/workflows/feedback.md` and its template twin differ by exactly one line — the source has `Last Updated: 2026-02-07` at line 184, the template does not (`diff` output: `184d183`). `feedback.md` is *not* in the byte-parity allowlist (`internal/template/rule_template_mirror_test.go:109-127`), so nothing enforces it; the Template-First rule (CLAUDE.local.md §2) still applies to any edit.

---

## 2. How the issue body is assembled — every text source that can reach a public issue

Body composition is prose-specified at `feedback.md:120-125`:

```
Issue body uses a consistent template in the user's conversation_language, including:
- Feedback type header (translated)
- Description content (user's original text)
- Priority level (translated)
- Tool-diagnostic information (see Diagnostic Attachment below)
```

**Complete source inventory** — the load-bearing answer for masking:

| # | Text source | Origin | Line | Bounded? |
|---|---|---|---|---|
| 1 | Title | user free text via the round's "Other" option | 58 | user-authored, unbounded |
| 2 | Description | user free text via "Other" | 59 | user-authored, unbounded |
| 3 | Type | fixed enum (Bug Report / Feature Request / Question) or `$ARGUMENTS` | 50, 54-57 | bounded |
| 4 | Priority | fixed enum (Low/Medium/High) | 61-65 | bounded |
| 5 | `conversation_language` | config | 90 | bounded |
| 6 | MoAI version | `moai version` — **guaranteed** | 131 | command output |
| 7 | OS | `uname` (OS name / release) — **guaranteed** | 132 | command output |
| 8 | Go toolchain version | `go version` — best-effort | 135 | command output |
| 9 | **Last-failed-command / error context** | passed in by the orchestrator | 137 | **unbounded** |
| 10 | Title keywords | derived from #1, leaves the machine as a `gh issue list --search` query | 71, 74 | derived from user text |

Two [HARD] fences already exist:

- `feedback.md:129` — "Two items are GUARANTEED (always collected)" (#6, #7).
- `feedback.md:139` — `[HARD] Diagnostic attachment is restricted to tool-diagnostic information only. The workflow MUST NOT attach arbitrary user file contents to the issue body — no reading or embedding of source files, configuration files, or any user-supplied file contents beyond the tool-diagnostic set above.`

Source #9 is the one the fence does not cover. Its own text says the boundary is delegated: "appended additively ONLY when the orchestrator passes it into the feedback invocation. The workflow never reads session error history itself." (`feedback.md:137`). Whatever the orchestrator hands over — a command line, a stack trace, an env dump — is appended verbatim. This is the surface any masking work has to target.

Also note #1/#2 are verbatim-preserved by policy: `feedback.md:104` — "Body content: User-provided text preserved verbatim (not translated, even if language differs from conversation_language)". A masking pass therefore has to be an explicit exception to that clause, not an implicit one.

Secondary egress: the duplicate search (`feedback.md:71`) sends title-derived keywords to GitHub *before* any submission decision. If `auto_submit` is meant to change what leaves the machine, this call already leaves it.

---

## 3. Config shape today, and every file an `auto_submit bool` would touch

### Current shape (one key, one field)

```go
// internal/config/types.go:1307-1314
// FeedbackConfig represents the /moai feedback workflow configuration section.
// SPEC-INVOCATION-MODEL-001: the feedback target repository is a config value so
// fork maintainers can redirect feedback away from the default tool channel.
type FeedbackConfig struct {
	// Repository is the "owner/repo" GitHub slug the feedback workflow targets.
	// Default DefaultFeedbackRepository (the remote MoAI-ADK tool repo).
	Repository string `yaml:"repository"`
}
```

- `internal/config/types.go:31` — ``Feedback FeedbackConfig `yaml:"feedback"` `` on `Config`.
- `internal/config/types.go:1316-1319` — ``feedbackFileWrapper{ Feedback FeedbackConfig `yaml:"feedback"` }``.
- `internal/config/defaults.go:212` — `DefaultFeedbackRepository = "modu-ai/moai-adk"`.
- `internal/config/defaults.go:436` — `Feedback: NewDefaultFeedbackConfig(),` in the default-Config constructor.
- `internal/config/defaults.go:451-455` — `func NewDefaultFeedbackConfig() FeedbackConfig { return FeedbackConfig{ Repository: DefaultFeedbackRepository } }`.
- `internal/config/loader.go:77` — `l.loadFeedbackSection(sectionsDir, cfg)`.
- `internal/config/loader.go:280-291` — the loader; wrapper is **seeded with the populated default** (`wrapper := &feedbackFileWrapper{Feedback: cfg.Feedback}`, line 281), which is the partial-override contract: a yaml omitting a key keeps the compiled default. A new bool inherits this behavior for free.
- `internal/config/slice.go:31` — `"feedback": (*Loader).loadFeedbackSection,` (lazy slice loader map).
- `internal/config/audit_registry.go:41` — `"feedback": "FeedbackConfig", // loaded via Loader.Load → loadFeedbackSection`.
- `internal/config/feedback_accessors.go:15-20` — the only accessor:

```go
func (c *Config) FeedbackRepository() string {
	if c.Feedback.Repository == "" {
		return DefaultFeedbackRepository
	}
	return c.Feedback.Repository
}
```

**Finding worth flagging:** `FeedbackRepository()` has **zero production callers**. `grep -rIn "FeedbackRepository()"` over the repo returns only `feedback_accessors.go:15` (the definition) and `feedback_config_test.go:29,64,97,...` (tests) — plus SPEC prose. Consistently, the key is classed `R` (reserved, `evidence: none`) in the inventory (`internal/config/testdata/shipped_key_inventory.yaml:362-364`). The consumer is the skill prose. An `auto_submit` key would land in exactly the same position: a Go field nothing in Go reads, consumed only by the workflow body.

### Every file an `auto_submit bool` would need

| # | File | Line | Edit |
|---|---|---|---|
| 1 | `internal/config/types.go` | 1310-1314 | add ``AutoSubmit bool `yaml:"auto_submit"` `` to `FeedbackConfig` |
| 2 | `internal/config/defaults.go` | ~212 | add `DefaultFeedbackAutoSubmit = false` const (repo convention: no bare literals — CLAUDE.local.md §14) |
| 3 | `internal/config/defaults.go` | 451-455 | set the field in `NewDefaultFeedbackConfig()` |
| 4 | `internal/config/feedback_accessors.go` | after 20 | add `func (c *Config) FeedbackAutoSubmit() bool` (matches the existing accessor pattern; note a `false` default makes the empty-fallback trick used for `Repository` a no-op) |
| 5 | `internal/template/templates/.moai/config/sections/feedback.yaml` | 6 | ship the key — **required**, this is the only tree `enumerateShippedKeys` scans (`shipped_key_reader_test.go:157` → `internal/template/templates/.moai/config/sections`) |
| 6 | `internal/config/testdata/shipped_key_inventory.yaml` | after 364 | add `- path: "feedback.auto_submit"` + class/evidence, or `TestShippedConfigKeysHaveReaders` FAILs (`shipped_key_reader_test.go:107-111`, sentinel text "REQ-CKH-008 anti-rot") |
| 7 | `internal/settings/schema_sections.go` | 395-396 | add `s(SectionFeedback, "feedback", TypeBool, "feedback", "auto_submit"),` next to the existing `TypeText ... "repository"` line — required for the web console to render/persist it |
| 8 | `internal/web/assets/i18n.js` | 523-524, 1235-1236, 1852-1853, 2469-2470 | add `f.feedback.auto_submit.title` + `.desc` in **all four** locales — `internal/web/schema_label_test.go:96` fails otherwise (`i18nKeyInAllLocales`) |
| 9 | `.moai/config/sections/feedback.yaml` + `internal/settings/testdata/sections/feedback.yaml` | 6 | the two non-template copies (the local project config and the settings-package fixture) — neither is scanned by the anti-rot guard, but both are byte-identical to the template today and would drift |

Loader (`loader.go:280`), slice map (`slice.go:31`), and audit registry (`audit_registry.go:41`) need **no** change — they register the section, not its keys.

Not required but likely wanted:
- `.claude/skills/moai/workflows/feedback.md` + its template mirror — the actual behavior change.
- `docs-site/content/{ko,en,ja,zh}/utility-commands/moai-feedback.md` — the 4-locale obligation (CLAUDE.local.md §17); the "Feedback Settings" section already documents `feedback.repository` in all four (`en:91`, `ko:91`, `ja:95`, `zh:95`).
- Template neutrality: the CI guard forbids SPEC IDs / REQ tokens in `internal/template/templates/**` comments (CLAUDE.local.md §2.1) — the existing `feedback.yaml` comment is already neutral; keep the new comment neutral too.

---

## 4. Does an existing test pin the feedback config surface?

Yes — `internal/config/feedback_config_test.go`, six tests, **all of them about `Repository` only**:

| Test | Line | Asserts |
|---|---|---|
| `TestFeedbackRepositoryDefault` | 12 | absent `feedback.yaml` → `FeedbackRepository() == DefaultFeedbackRepository`; and `loader.LoadedSections()["feedback"]` is **false** (34-35) |
| `TestFeedbackRepositoryOverride` | 44 | writes real `feedback:\n    repository: myfork/moai-adk\n` (53) through `Loader.Load()`, expects `myfork/moai-adk` (64) and section marked loaded (68) |
| `TestFeedbackRepositoryMissingKey` | 76 | `feedback:\n    unrelated_key: value\n` (86) → default (97) |
| `TestFeedbackRepositoryEmptyString` | 105 | `repository: ""` → default (125) |
| `TestFeedbackRepositoryInvalidYAML` | 133 | `repository: [unterminated` → default (154), section **not** marked loaded (157) |
| `TestNewDefaultFeedbackConfig` | 164 | `NewDefaultFeedbackConfig().Repository == DefaultFeedbackRepository` (168) |

**Would adding a field break them? No.** None constructs `FeedbackConfig` with positional/unkeyed literals; none compares the struct by value; none asserts a field count. `TestFeedbackRepositoryMissingKey` even proves the wrapper tolerates unknown keys. Adding `AutoSubmit` is additive here.

Tests that **would** fail without the corresponding registration (all outside this file):

- `internal/config/shipped_key_reader_test.go:70` `TestShippedConfigKeysHaveReaders` — fails on an untriaged shipped key (item 6 above). Its non-vacuity floor `minimumShippedKeys = 875` (line 53) is a lower bound only, so a +1 key is fine.
- `internal/web/schema_label_test.go:96` — fails if the new schema field's i18n key is missing in any locale (item 8).
- `internal/settings/schema_sections_test.go:452-468` holds a `"feedback.repository": "modu-ai/moai-adk"` expectation (line 460) in `TestSchemaCurrentValuesReadsAllSections`; it is a per-key map, not exhaustive, so a new key does not fail it — but adding the new key's expected value there is the consistent move.
- **Gap:** I did not run any test. All statements about pass/fail above are read from source, not observed. No `go test` was executed.

---

## 5. Existing bug-vs-enhancement label logic

Yes — prose only, at `.claude/skills/moai/workflows/feedback.md:92-96`:

```
### GitHub Issue Labels

- Bug Report: labels "bug"
- Feature Request: labels "enhancement"
- Question: labels "question"
```

Reinforced at `feedback.md:105`: `- **Labels**: English only (GitHub standard: "bug", "enhancement", "question")`.

The mapping input is the Type field collected at `feedback.md:54-57` (Bug Report / Feature Request / Question). There is **no Go code, no config key, and no `--label` flag text** implementing this — the orchestrator is expected to translate the enum to a label at `gh issue create` time.

---

## Contradictions with the card's assumptions

1. **"the feedback skill currently asks a confirmation question that `auto_submit=true` would skip" — false.** Verified by enumeration, not by restatement: `AskUserQuestion` appears at `feedback.md:52` (collect), `156` (post-submission), `178` (summary of 52). Nothing between collection and `gh issue create` (118). The docs agree: `docs-site/content/en/utility-commands/moai-feedback.md:179` "Once you answer, a GitHub issue is created automatically". What `auto_submit` would actually have to suppress is the **collection** round — a materially different design, since the title/description then have to come from somewhere else (`$ARGUMENTS`, per `feedback.md:50`, resolves only the *type*).

2. **There is an adjacent, weaker confirmation obligation the card does not mention.** `.claude/rules/moai/core/askuser-protocol.md` § Socratic Interview Structure item 7 requires "explicit final confirmation via `AskUserQuestion` before irreversible actions" as a general orchestrator rule, and `.claude/skills/moai/SKILL.md:360` exempts `feedback` only from Step 2.8 (requirement analysis), not from that. So the *skill* prescribes no confirm, while a general rule arguably implies one — meaning behavior today may vary by session. An `auto_submit` option would need to say which of the two it overrides. **Gap:** I could not determine what the orchestrator actually does in practice; that is a runtime-behavior question, not answerable from source.

3. **"config option" implies a Go-enforced switch — it would not be.** `feedback.repository` has zero production Go readers (§3); the config layer only *carries* the value. `auto_submit` would be the same: a bool the skill prose is trusted to honor. Whether that satisfies the card's intent is a design question the card does not settle.

4. **The card frames `auto_submit` as filing "automatically instead of asking", but the path already has a non-optional pre-submission network call** — the duplicate search at `feedback.md:71` sends title keywords to GitHub before any decision point. Any privacy/masking framing has to account for it.

5. **No existing `auto_submit` anywhere.** `grep -rIn "auto_submit\|AutoSubmit\|autoSubmit"` over the whole worktree (excluding `.git`) returns zero matches — this is a genuinely new key, not a revival.

## Gaps (explicitly not determined)

- No commands were run beyond `grep`/`diff`/`awk`/`git log`. No `go test`, no `gh`, no build. Every "would fail / would not fail" claim is source-read, not observed.
- The literal `gh issue create` flag set does not exist in the repo; I could not report what flags are used because none are written down (`feedback.md:118` is the whole invocation).
- Whether the orchestrator currently inserts an ad-hoc confirmation in practice (per the general askuser rule) is unknown — unobservable from source.
- I did not audit `internal/web` render/persist code paths beyond the schema and i18n registration points, so there may be further web-console surfaces a new bool field touches.

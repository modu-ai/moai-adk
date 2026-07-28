package web

// i18n_untranslated_allowlist_test.go — the untranslated-value governance surface.
//
// This file is the SINGLE owning artifact for the web console i18n catalogue's
// intentional-untranslated-value allowlist (SPEC-I18N-GOVERNANCE-001). It is a
// test-binary-only artifact (package web, build-tag none, consumed by tests) —
// no production (non-_test.go) symbol depends on it, so the shipped console
// binary is byte-unaffected.
//
// ── Governance contract ────────────────────────────────────────────────────
//
// Location : internal/web/i18n_untranslated_allowlist_test.go (this file).
//            Deliberately OUTSIDE internal/web/assets/ so the /static/ handler
//            never serves it (REQ-I18NGOV-001, B6).
//
// Entry    : Each i18nUntranslatedAllowEntry carries three fields — the exact
// format     catalogue Key, a Reason drawn from the closed taxonomy below, and
//            a free-text Justification naming WHY the value is locale-invariant
//            (REQ-I18NGOV-002).
//
// Who adds : Any contributor may add an entry. The reviewer SHALL, before
// an entry   accepting one, assert (a) the Reason fits the closed taxonomy,
//            (b) the Justification cites mechanical evidence (the config value,
//            CLI token, product/model name, or initialism the literal must
//            match byte-for-byte), and (c) the orphan check
//            (TestI18nAllowlistNoOrphans) still passes — i.e. the key exists in
//            the catalogue and is genuinely identical to `en` in at least one
//            locale (REQ-I18NGOV-015). A stale entry surviving the deletion or
//            translation of its key MUST be removed in the same change.
//
// Ruling   : For a borderline value, classify against the closed taxonomy. If
// procedure  no taxonomy reason applies, the value is untranslated prose and
//            the fix is a translation in i18n.js, NOT an allowlist entry
//            (REQ-I18NGOV-023). Record the outcome as either an entry here or
//            a translation fix.
//
// Bound    : The allowlist SHALL carry no more than 30 entries
// (REQ-I18NGOV-016). The 23 measured legitimate members (the exact-identity
//            set at run-phase minus lang.opt.* which is handled structurally
//            by the endonym invariant, and minus the translation-fix rulings)
//            sit under that bound with headroom for legitimate growth.
//
// Wildcards: Forbidden (REQ-I18NGOV-004). Every entry names one exact key.
//
// lang.opt: The lang.opt.* family is NEVER allowlisted (REQ-I18NGOV-014); it
//            is governed by the bidirectional endonym invariant
//            (TestI18nEndonymInvariants).

// i18nUntranslatedReason is the closed taxonomy of reasons a catalogue value
// is intentionally left locale-invariant. The set is closed at compile time:
// an invented reason fails to assign to the type, so go vet catches it before
// any test runs (REQ-I18NGOV-003, AC-I18NGOV-004).
type i18nUntranslatedReason string

const (
	// reasonTechnicalIdentifier — the value must match a configuration value,
	// CLI token, or machine-emitted string byte-for-byte (e.g. an enum literal
	// like "auto", "manual", "html+md", or a Severity value like "MUST-FIX").
	reasonTechnicalIdentifier i18nUntranslatedReason = "technical-identifier"
	// reasonProperNoun — a product, brand, vendor, model, or convention name
	// that is locale-invariant (e.g. "Opus 5", "Angular", "MoAI-Loop").
	reasonProperNoun i18nUntranslatedReason = "proper-noun"
	// reasonAcronym — a locale-invariant initialism (e.g. "LLM").
	reasonAcronym i18nUntranslatedReason = "acronym"
)

// i18nUntranslatedAllowEntry is one allowlist row (REQ-I18NGOV-002).
type i18nUntranslatedAllowEntry struct {
	Key           string
	Reason        i18nUntranslatedReason
	Justification string
}

// i18nUntranslatedAllowlist is the complete set of intentionally-untranslated
// catalogue keys. Entries are scoped per key and apply to every non-`en`
// locale (REQ-I18NGOV-005): after the lang.opt.* family is handled structurally
// the three locales' identity sets are byte-identical, so per-locale narrowing
// would triple the artifact with no added signal. The residual risk (a global
// entry masking a genuinely-untranslated copy of a key in one locale) is
// accepted and bounded by the orphan check plus the 30-entry cap.
//
// The entries below were derived by ruling on every member of the re-measured
// exact-identity set (run-phase measurement: 26 keys per locale; minus one
// lang.opt.<self> handled structurally, minus three translation-fix rulings
// for mp.tier.empty / mp.col.effort / f.model.desc). See the run-phase report
// for the two judgment rulings (board.badge.mustfix, mp.tier.empty).
var i18nUntranslatedAllowlist = []i18nUntranslatedAllowEntry{
	// ── technical-identifier : Severity / badge literals ────────────────────
	{
		Key:    "board.badge.mustfix",
		Reason: reasonTechnicalIdentifier,
		Justification: "the Severity value emitted by internal/spec/audit.go and " +
			"matched verbatim by internal/web/board.go and internal/cli/spec_audit.go; " +
			"the console badge must read the same token the CLI and its JSON output emit " +
			"so a user grepping `moai spec audit --json` finds the badge.",
	},

	// ── technical-identifier : cacheStrategy TTL enum ──────────────────────
	{
		Key:           "f.cacheStrategy.session_ttl.opt.1h",
		Reason:        reasonTechnicalIdentifier,
		Justification: "the literal matches the session_ttl config enum value `1h` byte-for-byte.",
	},
	{
		Key:           "f.cacheStrategy.session_ttl.opt.5m",
		Reason:        reasonTechnicalIdentifier,
		Justification: "the literal matches the session_ttl config enum value `5m` byte-for-byte.",
	},
	{
		Key:           "f.cacheStrategy.session_ttl.opt.off",
		Reason:        reasonTechnicalIdentifier,
		Justification: "the literal matches the session_ttl config enum value `off` byte-for-byte.",
	},

	// ── technical-identifier : handoff.mode enum ───────────────────────────
	{
		Key:           "f.handoff.mode.opt.auto",
		Reason:        reasonTechnicalIdentifier,
		Justification: "the literal matches the handoff.mode config value `auto` (handoff.yaml).",
	},
	{
		Key:           "f.handoff.mode.opt.manual",
		Reason:        reasonTechnicalIdentifier,
		Justification: "the literal matches the handoff.mode config value `manual` (handoff.yaml).",
	},

	// ── technical-identifier : report.format enum ──────────────────────────
	{
		Key:           "f.report.format.opt.html+md",
		Reason:        reasonTechnicalIdentifier,
		Justification: "the literal matches the report.format config value `html+md` byte-for-byte.",
	},
	{
		Key:           "f.report.format.opt.md",
		Reason:        reasonTechnicalIdentifier,
		Justification: "the literal matches the report.format config value `md` byte-for-byte.",
	},

	// ── proper-noun : git-convention names ─────────────────────────────────
	{
		Key:           "f.git_convention.opt.angular",
		Reason:        reasonProperNoun,
		Justification: "the Angular framework commit-convention name; locale-invariant proper noun.",
	},
	{
		Key:           "f.git_convention.opt.conventional-commits",
		Reason:        reasonProperNoun,
		Justification: "the Conventional Commits specification name; locale-invariant proper noun.",
	},
	{
		Key:           "f.git_convention.opt.karma",
		Reason:        reasonProperNoun,
		Justification: "the Karma test-runner convention name; locale-invariant proper noun.",
	},

	// ── proper-noun : model names (f.model.opt.*) ──────────────────────────
	{
		Key:           "f.model.opt.fable",
		Reason:        reasonProperNoun,
		Justification: "Anthropic model name `Fable 5`; locale-invariant proper noun.",
	},
	{
		Key:           "f.model.opt.fable[1m]",
		Reason:        reasonProperNoun,
		Justification: "Anthropic model name `Fable 5` with the 1M-context variant suffix; proper noun.",
	},
	{
		Key:           "f.model.opt.haiku",
		Reason:        reasonProperNoun,
		Justification: "Anthropic model name `Haiku 4.5`; locale-invariant proper noun.",
	},
	{
		Key:           "f.model.opt.opus",
		Reason:        reasonProperNoun,
		Justification: "Anthropic model name `Opus 5`; locale-invariant proper noun.",
	},
	{
		Key:           "f.model.opt.opus[1m]",
		Reason:        reasonProperNoun,
		Justification: "Anthropic model name `Opus 5` with the 1M-context variant suffix; proper noun.",
	},
	{
		Key:           "f.model.opt.sonnet",
		Reason:        reasonProperNoun,
		Justification: "Anthropic model name `Sonnet 5`; locale-invariant proper noun.",
	},
	{
		Key:           "f.model.opt.sonnet[1m]",
		Reason:        reasonProperNoun,
		Justification: "Anthropic model name `Sonnet 5` with the 1M-context variant suffix; proper noun.",
	},

	// ── proper-noun : model names (f.v4.model.opt.*) ───────────────────────
	{
		Key:           "f.v4.model.opt.haiku",
		Reason:        reasonProperNoun,
		Justification: "Anthropic model family name `Haiku`; locale-invariant proper noun.",
	},
	{
		Key:           "f.v4.model.opt.opus",
		Reason:        reasonProperNoun,
		Justification: "Anthropic model family name `Opus`; locale-invariant proper noun.",
	},
	{
		Key:           "f.v4.model.opt.sonnet",
		Reason:        reasonProperNoun,
		Justification: "Anthropic model family name `Sonnet`; locale-invariant proper noun.",
	},

	// ── acronym ────────────────────────────────────────────────────────────
	{
		Key:           "sec.launch.title",
		Reason:        reasonAcronym,
		Justification: "the section title is the initialism `LLM`; locale-invariant acronym.",
	},

	// ── proper-noun : product name ─────────────────────────────────────────
	{
		Key:           "sec.ralph.title",
		Reason:        reasonProperNoun,
		Justification: "the MoAI-Loop product name; locale-invariant proper noun.",
	},
}

// i18nEnExemptPrefixRegistry is the explicit registry of key prefixes that are
// deliberately absent from the `en` locale (REQ-I18NGOV-019, REQ-I18NGOV-020).
// Each member carries a justification naming the surface that supplies the
// English baseline instead. Adding a prefix is a reviewed act: a new
// unexplained reverse-asymmetry fails TestI18nKeyCoverageReverse.
var i18nEnExemptPrefixRegistry = []struct {
	Prefix        string
	Justification string
}{
	{
		Prefix:        "agentdesc.",
		Justification: "English reads the agent .md frontmatter `description` as the " +
			"server-rendered baseline (the single source); applyI18n leaves a node " +
			"untouched when the key is absent, so duplicating the .md text into the " +
			"dictionary would create a second surface that silently goes stale. " +
			"Enforced structurally by TestD3AgentDescIsEnExempt (C1).",
	},
}

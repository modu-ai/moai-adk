package web

// i18n_untranslated_allowlist_test.go — the untranslated-value allowlist for
// internal/web/assets/i18n.js.
//
// ─── Governance contract (REQ-I18NGOV-001 / 006 / 023) ────────────────────────
//
// Location:   this file (internal/web/i18n_untranslated_allowlist_test.go),
//             co-located with the console package and OUTSIDE
//             internal/web/assets/ so the /static/ handler never serves it
//             (REQ-I18NGOV-001). Test-binary only — the shipped console binary
//             is byte-unaffected (C4).
//
// Entry form: each entry carries three fields —
//               • Key          the exact catalogue key (no wildcards / globs /
//                              regex metacharacters; REQ-I18NGOV-004). One
//                              entry applies to EVERY non-en locale
//                              (REQ-I18NGOV-005).
//               • Reason       one of the closed taxonomy below
//                              (REQ-I18NGOV-003). An invented reason fails to
//                              compile, not merely fails a test.
//               • Justification why the value is locale-invariant, citing the
//                              mechanical evidence (config value, CLI token,
//                              model/vendor/convention name, …).
//
// Who adds:   a contributor who has measured a key's value identical to its en
//             counterpart under NFC + trim + casefold normalization in at least
//             one non-en locale, and who can classify the value against the
//             closed taxonomy with mechanical evidence.
//
// Reviewer assertion: before accepting an entry the reviewer confirms that (a)
//             the key is genuinely locale-invariant by intent rather than an
//             untranslated defect, (b) no taxonomy-short English-prose value is
//             being rubber-stamped (prose whose siblings are translated is a
//             translation fix, not an allowlist entry), and (c) the entry cap
//             of 30 (REQ-I18NGOV-016) is not exceeded.
//
// Ruling procedure (REQ-I18NGOV-023) for a borderline value:
//             1. Re-measure the key against en under NFC + trim + casefold.
//             2. Attempt to classify against the closed taxonomy:
//                  technical-identifier | proper-noun | acronym.
//             3. If a taxonomy reason fits with mechanical evidence, record the
//                entry here with the justification.
//             4. If no taxonomy reason fits (the value is English prose, a
//                sentence, or a label whose siblings are translated), the value
//                is a translation defect: fix it in i18n.js instead.
//
// Anti-rubber-stamp: the detector is a pure function over (catalogue,
//             allowlist); a negative-control test proves a non-allowlisted
//             untranslated key fails. An orphan check removes any entry whose
//             key was deleted or later translated.
//
// Owning surface pointer: internal/web/assets/i18n.js header names this file
//             as the governance surface (REQ-I18NGOV-021).

// i18nAllowReason is the closed taxonomy of reasons a catalogue value is
// intentionally locale-invariant (REQ-I18NGOV-003). It is a Go type so that an
// invented reason fails to compile before any test runs.
type i18nAllowReason string

const (
	// reasonTechnicalIdentifier: the value must match a configuration value,
	// CLI token, or machine-emitted string byte-for-byte.
	reasonTechnicalIdentifier i18nAllowReason = "technical-identifier"
	// reasonProperNoun: a product, brand, vendor, model, or convention name.
	reasonProperNoun i18nAllowReason = "proper-noun"
	// reasonAcronym: a locale-invariant initialism.
	reasonAcronym i18nAllowReason = "acronym"
)

// i18nAllowReasons is the closed set — membership is asserted by
// TestI18nAllowlistShape so adding a fourth reason is a reviewed act.
var i18nAllowReasons = map[i18nAllowReason]struct{}{
	reasonTechnicalIdentifier: {},
	reasonProperNoun:          {},
	reasonAcronym:             {},
}

// i18nAllowEntry is one allowlist row (REQ-I18NGOV-002).
type i18nAllowEntry struct {
	Key           string
	Reason        i18nAllowReason
	Justification string
}

// i18nUntranslatedAllowlist is the single, exhaustive registry of catalogue
// keys whose non-en value is intentionally identical to the en value. Every
// member was re-measured at run-phase entry; the two identity-set members that
// admit no taxonomy reason — mp.tier.empty and f.model.desc (English prose whose
// siblings are translated) — were translated in i18n.js rather than listed here.
//
// The lang.opt.* endonym family is NOT listed here: it is handled structurally
// by the endonym invariants (REQ-I18NGOV-012/013) and is explicitly forbidden
// from this allowlist (REQ-I18NGOV-014).
var i18nUntranslatedAllowlist = []i18nAllowEntry{
	// Severity literal — the console badge must read the token the CLI/JSON emit.
	{
		Key:           "board.badge.mustfix",
		Reason:        reasonTechnicalIdentifier,
		Justification: "the MUST-FIX Severity value emitted by internal/spec/audit.go and matched by internal/web/board.go and internal/cli/spec_audit.go; the console badge must read the same token a user greps from `moai spec audit --json`.",
	},
	// session_ttl duration / sentinel literals (config enum values).
	{
		Key:           "f.cacheStrategy.session_ttl.opt.1h",
		Reason:        reasonTechnicalIdentifier,
		Justification: "duration literal matching the session_ttl configuration value byte-for-byte.",
	},
	{
		Key:           "f.cacheStrategy.session_ttl.opt.5m",
		Reason:        reasonTechnicalIdentifier,
		Justification: "duration literal matching the session_ttl configuration value byte-for-byte.",
	},
	{
		Key:           "f.cacheStrategy.session_ttl.opt.off",
		Reason:        reasonTechnicalIdentifier,
		Justification: "sentinel value that disables the session cache; translating it would diverge from the config enum.",
	},
	// Git-convention brand names.
	{
		Key:           "f.git_convention.opt.angular",
		Reason:        reasonProperNoun,
		Justification: "Angular commit-convention brand name.",
	},
	{
		Key:           "f.git_convention.opt.conventional-commits",
		Reason:        reasonProperNoun,
		Justification: "Conventional Commits convention name.",
	},
	{
		Key:           "f.git_convention.opt.karma",
		Reason:        reasonProperNoun,
		Justification: "Karma test-runner convention name.",
	},
	// handoff.mode config enum values.
	{
		Key:           "f.handoff.mode.opt.auto",
		Reason:        reasonTechnicalIdentifier,
		Justification: "handoff.mode configuration enum value.",
	},
	{
		Key:           "f.handoff.mode.opt.manual",
		Reason:        reasonTechnicalIdentifier,
		Justification: "handoff.mode configuration enum value.",
	},
	// Anthropic model names (product/brand names).
	{
		Key:           "f.model.opt.fable",
		Reason:        reasonProperNoun,
		Justification: "Anthropic Fable model name.",
	},
	{
		Key:           "f.model.opt.fable[1m]",
		Reason:        reasonProperNoun,
		Justification: "Anthropic Fable model name, 1M-context variant selector.",
	},
	{
		Key:           "f.model.opt.haiku",
		Reason:        reasonProperNoun,
		Justification: "Anthropic Haiku model name.",
	},
	{
		Key:           "f.model.opt.opus",
		Reason:        reasonProperNoun,
		Justification: "Anthropic Opus model name.",
	},
	{
		Key:           "f.model.opt.opus[1m]",
		Reason:        reasonProperNoun,
		Justification: "Anthropic Opus model name, 1M-context variant selector.",
	},
	{
		Key:           "f.model.opt.sonnet",
		Reason:        reasonProperNoun,
		Justification: "Anthropic Sonnet model name.",
	},
	{
		Key:           "f.model.opt.sonnet[1m]",
		Reason:        reasonProperNoun,
		Justification: "Anthropic Sonnet model name, 1M-context variant selector.",
	},
	// report.format config enum literals.
	{
		Key:           "f.report.format.opt.html+md",
		Reason:        reasonTechnicalIdentifier,
		Justification: "report.format configuration enum literal emitted by the HTML report skill.",
	},
	{
		Key:           "f.report.format.opt.md",
		Reason:        reasonTechnicalIdentifier,
		Justification: "report.format configuration enum literal.",
	},
	// Anthropic model family names (v4 surface).
	{
		Key:           "f.v4.model.opt.haiku",
		Reason:        reasonProperNoun,
		Justification: "Anthropic Haiku model family name.",
	},
	{
		Key:           "f.v4.model.opt.opus",
		Reason:        reasonProperNoun,
		Justification: "Anthropic Opus model family name.",
	},
	{
		Key:           "f.v4.model.opt.sonnet",
		Reason:        reasonProperNoun,
		Justification: "Anthropic Sonnet model family name.",
	},
	// Locale-invariant initialism.
	{
		Key:           "sec.launch.title",
		Reason:        reasonAcronym,
		Justification: "LLM is a locale-invariant initialism.",
	},
	// Product feature name.
	{
		Key:           "sec.ralph.title",
		Reason:        reasonProperNoun,
		Justification: "MoAI-Loop product feature name.",
	},
}

// i18nMaxAllowlistEntries bounds the allowlist so blanket growth fails loudly
// rather than accreting silently (REQ-I18NGOV-016). Derived from the 23 measured
// legitimate members plus headroom.
const i18nMaxAllowlistEntries = 30

// i18nExemptPrefix is one row of the en-exempt prefix registry
// (REQ-I18NGOV-020): a key prefix that is legitimately present in a non-en
// locale but absent from en, with the surface that supplies the English
// baseline instead.
type i18nExemptPrefix struct {
	Prefix        string
	Justification string
}

// i18nEnExemptPrefixes is the explicit, enumerated registry of key prefixes
// that may appear in non-en locales without an en counterpart. Its sole initial
// member is agentdesc. (REQ-I18NGOV-020, C1): English reads the agent .md
// frontmatter description as the server-rendered baseline, and applyI18n guards
// its assignment on a non-empty string so an absent key leaves that baseline
// intact. Adding a prefix is a reviewed act, not a silent one.
var i18nEnExemptPrefixes = []i18nExemptPrefix{
	{
		Prefix:        "agentdesc.",
		Justification: "English reads the agent .md frontmatter description (the SSOT) as the server-rendered baseline; applyI18n leaves the node untouched when the key is absent, so an en copy would duplicate the .md text into a second surface that silently goes stale.",
	},
}

package web

// SPEC-I18N-GOVERNANCE-001 — Untranslated-value allowlist artifact (M1).
//
// Governance contract
// -------------------
// Location     : this file (internal/web/i18n_untranslated_allowlist_test.go),
//                 package `web`, OUTSIDE internal/web/assets/ so it is never
//                 served by the /static/ handler.
// Consumed by  : tests only — no runtime (non-_test.go) symbol references it,
//                 so the shipped console binary is byte-unaffected.
// Entry format : each entry is one exact catalogue key plus a reason drawn
//                 from the closed `untranslatedReason` taxonomy below and a
//                 free-text justification naming why the value is
//                 locale-invariant. Wildcards, globs, prefixes, and regex
//                 entries are rejected (one exact key per entry).
// Who adds one : any contributor, in the same change that introduces the
//                 locale-invariant value.
// Reviewer assertion: before accepting an entry, the reviewer confirms the
//                 value really is locale-invariant (it matches a config enum,
//                 CLI token, machine-emitted string, product/model/convention
//                 name, or locale-invariant initialism) AND that translating
//                 it would break a correspondence a user or tool relies on.
// Ruling procedure for a borderline value:
//                 1. classify against the closed taxonomy;
//                 2. cite mechanical evidence for the classification
//                    (the emitting source, the sibling values, the grep that
//                    proves the token correspondence);
//                 3. record the outcome — either an allowlist entry with the
//                    taxonomy reason + justification, or a translation fix in
//                    i18n.js. A value may NOT be left ungoverned.
// Scope        : entries are per key and apply to every non-`en` locale.
//                 After the `lang.opt.*` family is excluded structurally
//                 (endonym invariants), the three non-`en` locales' identical
//                 sets are byte-identical, so per-locale narrowing would
//                 triple the artifact with no added signal. Per-locale
//                 narrowing is introduced only when a measurement
//                 demonstrates divergence.
// Bound        : at most 30 entries (i18nAllowlistMaxEntries), derived from
//                 the measured legitimate set plus headroom, so blanket
//                 growth fails loudly rather than accreting silently.

// untranslatedReason is the closed reason taxonomy for an allowlist entry.
// It is a named type (not a bare string) so an invented reason fails to
// compile rather than merely failing a test.
type untranslatedReason string

const (
	// reasonTechnicalIdentifier: the value must match a configuration value,
	// CLI token, or machine-emitted string byte-for-byte.
	reasonTechnicalIdentifier untranslatedReason = "technical-identifier"
	// reasonProperNoun: a product, brand, vendor, model, or convention name.
	reasonProperNoun untranslatedReason = "proper-noun"
	// reasonAcronym: a locale-invariant initialism.
	reasonAcronym untranslatedReason = "acronym"
)

// i18nAllowlistMaxEntries is the upper bound on allowlist size, derived from
// the measured legitimate set plus headroom (REQ-I18NGOV-016).
const i18nAllowlistMaxEntries = 30

// untranslatedAllowEntry is one allowlist ruling: an exact catalogue key plus
// its taxonomy reason and a free-text justification.
type untranslatedAllowEntry struct {
	key          string
	reason       untranslatedReason
	justification string
}

// untranslatedAllowlist is the set of catalogue keys intentionally left equal
// to their `en` value across every non-`en` locale. Every member was measured
// byte-identical to `en` in ko/ja/zh at the worktree base (SPEC §A.1) and
// ruled locale-invariant per the governance contract above.
//
// Two members of the measured identical set are NOT here because they are
// defects the mechanism exists to find, not exemptions:
//   - mp.tier.empty — English prose; translated in i18n.js (sibling
//     mp.tier.default is translated in every locale).
//   - f.model.desc  — English prose; translated in i18n.js (sibling f.*.desc
//     values are translated in every locale).
//
// mp.col.effort is included: the effort axis labels a control whose values
// (max/high/medium/low) are locale-invariant machine tokens, so the axis name
// is a technical reference to those tokens. The non-`en` value was capitalized
// to "Effort" in i18n.js so the entry is byte-identical (not merely
// case-fold-equal), satisfying the orphan check.
var untranslatedAllowlist = []untranslatedAllowEntry{
	// --- technical identifiers (config enums + CLI/machine tokens) ---
	{
		key:           "board.badge.mustfix",
		reason:        reasonTechnicalIdentifier,
		justification: "Literal MUST-FIX emitted by internal/spec/audit.go Severity and matched by internal/web/board.go and internal/cli/spec_audit.go; the badge must read the same token the CLI and its JSON output emit so a user grepping `moai spec audit --json` finds it.",
	},
	{
		key:           "f.cacheStrategy.session_ttl.opt.1h",
		reason:        reasonTechnicalIdentifier,
		justification: "Config enum value (1h) for f.cacheStrategy.session_ttl; matches the YAML token a user writes.",
	},
	{
		key:           "f.cacheStrategy.session_ttl.opt.5m",
		reason:        reasonTechnicalIdentifier,
		justification: "Config enum value (5m) for f.cacheStrategy.session_ttl; matches the YAML token a user writes.",
	},
	{
		key:           "f.cacheStrategy.session_ttl.opt.off",
		reason:        reasonTechnicalIdentifier,
		justification: "Config enum value (off) for f.cacheStrategy.session_ttl; matches the YAML token a user writes.",
	},
	{
		key:           "f.handoff.mode.opt.auto",
		reason:        reasonTechnicalIdentifier,
		justification: "Config enum value (auto) for handoff.mode; matches the YAML token.",
	},
	{
		key:           "f.handoff.mode.opt.manual",
		reason:        reasonTechnicalIdentifier,
		justification: "Config enum value (manual) for handoff.mode; matches the YAML token.",
	},
	{
		key:           "f.report.format.opt.html+md",
		reason:        reasonTechnicalIdentifier,
		justification: "Config enum value (html+md) for report.format; matches the YAML token and the CLI-accepted value.",
	},
	{
		key:           "f.report.format.opt.md",
		reason:        reasonTechnicalIdentifier,
		justification: "Config enum value (md) for report.format; matches the YAML token and the CLI-accepted value.",
	},
	{
		key:           "mp.col.effort",
		reason:        reasonTechnicalIdentifier,
		justification: "Column header for the effort axis whose values (max/high/medium/low) are locale-invariant machine tokens; the label is a technical reference to those tokens. Non-en value capitalized to 'Effort' in i18n.js so the entry is byte-identical, not merely case-fold-equal.",
	},

	// --- proper nouns (product / model / convention names) ---
	{
		key:           "f.git_convention.opt.angular",
		reason:        reasonProperNoun,
		justification: "Angular — a proper-noun commit convention name.",
	},
	{
		key:           "f.git_convention.opt.conventional-commits",
		reason:        reasonProperNoun,
		justification: "Conventional Commits — a proper-noun commit convention name.",
	},
	{
		key:           "f.git_convention.opt.karma",
		reason:        reasonProperNoun,
		justification: "Karma — a proper-noun commit convention name.",
	},
	{
		key:           "f.model.opt.fable",
		reason:        reasonProperNoun,
		justification: "Fable — a proper-noun model name.",
	},
	{
		key:           "f.model.opt.fable[1m]",
		reason:        reasonProperNoun,
		justification: "Fable (1M context variant) — a proper-noun model name.",
	},
	{
		key:           "f.model.opt.haiku",
		reason:        reasonProperNoun,
		justification: "Haiku — a proper-noun model name.",
	},
	{
		key:           "f.model.opt.opus",
		reason:        reasonProperNoun,
		justification: "Opus — a proper-noun model name.",
	},
	{
		key:           "f.model.opt.opus[1m]",
		reason:        reasonProperNoun,
		justification: "Opus (1M context variant) — a proper-noun model name.",
	},
	{
		key:           "f.model.opt.sonnet",
		reason:        reasonProperNoun,
		justification: "Sonnet — a proper-noun model name.",
	},
	{
		key:           "f.model.opt.sonnet[1m]",
		reason:        reasonProperNoun,
		justification: "Sonnet (1M context variant) — a proper-noun model name.",
	},
	{
		key:           "f.v4.model.opt.haiku",
		reason:        reasonProperNoun,
		justification: "Haiku (v4 builder variant) — a proper-noun model name.",
	},
	{
		key:           "f.v4.model.opt.opus",
		reason:        reasonProperNoun,
		justification: "Opus (v4 builder variant) — a proper-noun model name.",
	},
	{
		key:           "f.v4.model.opt.sonnet",
		reason:        reasonProperNoun,
		justification: "Sonnet (v4 builder variant) — a proper-noun model name.",
	},
	{
		key:           "sec.ralph.title",
		reason:        reasonProperNoun,
		justification: "MoAI-Loop — a proper-noun product/feature name.",
	},

	// --- acronyms (locale-invariant initialisms) ---
	{
		key:           "sec.launch.title",
		reason:        reasonAcronym,
		justification: "LLM — a locale-invariant initialism.",
	},
}

// enExemptPrefixRegistry lists key prefixes whose keys are intentionally
// absent from the `en` block (reverse-coverage exemption, REQ-I18NGOV-019 /
// REQ-I18NGOV-020). Each member names the surface that supplies the English
// baseline instead of the dictionary.
var enExemptPrefixRegistry = []struct {
	prefix        string
	justification string
}{
	{
		prefix:        "agentdesc.",
		justification: "English reads the agent .md frontmatter description as the server-rendered baseline (the SSOT); applyI18n leaves a node untouched when the key is absent, so an en copy would duplicate the .md text into a second surface that silently goes stale. Enforced by TestD3AgentDescIsEnExempt.",
	},
}

// isEnExempt reports whether a key absent from `en` matches a declared
// en-exempt prefix (REQ-I18NGOV-019). A key matches by exact prefix; no
// wildcard or regex matching is performed.
func isEnExempt(key string) bool {
	for _, e := range enExemptPrefixRegistry {
		if len(key) >= len(e.prefix) && key[:len(e.prefix)] == e.prefix {
			return true
		}
	}
	return false
}

// allowlistReasonFor returns the reason classification for key and whether key
// is allowlisted. Used by the orphan check and the negative control.
func allowlistReasonFor(key string) (untranslatedReason, bool) {
	for _, e := range untranslatedAllowlist {
		if e.key == key {
			return e.reason, true
		}
	}
	return "", false
}

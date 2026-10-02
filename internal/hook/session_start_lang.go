package hook

// session_start_lang.go holds the two locale helpers the factory notices, the
// stale-run notice, factory_messages.go, and the SessionStart handler share.
// They moved here from session_start_kanban_i18n.go (SPEC-LAUNCHER-ENTRY-FLAGS-001
// M5b first step) so that file could be deleted without taking them along.

// langEnglish is the fallback locale and the language of every agent-facing
// copy. A locale absent from a locale table (factoryLocales, staleRunLocales)
// resolves here rather than yielding an empty notice — an English instruction
// the operator can still act on beats no instruction at all.
const langEnglish = "en"

// operatorLang reports the locale the operator reads, or langEnglish when the
// configuration is unavailable.
//
// Fail-open like the rest of the hook: a nil provider, a nil config, or an empty
// conversation_language degrades to English rather than failing the session
// start. The message-table resolvers absorb an unknown value, so no validation
// happens here — the locale is passed through as configured.
func operatorLang(cfg ConfigProvider) string {
	if cfg == nil {
		return langEnglish
	}
	c := cfg.Get()
	if c == nil || c.Language.ConversationLanguage == "" {
		return langEnglish
	}
	return c.Language.ConversationLanguage
}

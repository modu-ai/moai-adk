package web

// glm_tier_test.go — SPEC-WEB-CONSOLE-REDESIGN-001 M4 guards
// (AC-WCR-030..034) for the 3rd Party LLM tab.
//
// The tab's hard problem is honesty, not wiring. Four per-tier reasoning-effort
// values are editable and stored, but NONE of them reaches the runtime: the GLM
// launcher injects exactly one session-global ANTHROPIC_REASONING_EFFORT derived
// from the LLM tab's session-wide effort_level preference, which is unrelated to
// the GLM tier map. These tests pin that the UI says so and never implies the
// per-tier values are applied.

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/glmcred"
	"github.com/modu-ai/moai-adk/internal/settings"
	"github.com/modu-ai/moai-adk/internal/template"
)

// glmTierKeys are the internal tier keys. REQ-WCR-032 pins that these stay put
// while the DISPLAY labels move to the Claude-facing names.
var glmTierKeys = []string{"high", "medium", "low", "fable"}

// fieldByName looks a FieldDef up by its logical name.
func fieldByName(t *testing.T, name string) settings.FieldDef {
	t.Helper()
	for _, f := range settings.AllFields() {
		if f.Name == name {
			return f
		}
	}
	t.Fatalf("field %q is absent from the schema", name)
	return settings.FieldDef{}
}

func optionValues(f settings.FieldDef) []string {
	out := make([]string, 0, len(f.Options))
	for _, o := range f.Options {
		out = append(out, o.Value)
	}
	return out
}

// TestGLMModelSelectOptions verifies AC-WCR-030: the four tier fields are
// closed-set selects over exactly {glm-5.3-flash, glm-5.3, glm-5.1, glm-4.7,
// glm-4.5-air} — flash first (the default), glm-5.3 retained as selectable.
func TestGLMModelSelectOptions(t *testing.T) {
	want := []string{"glm-5.3-flash", "glm-5.3", "glm-5.1", "glm-4.7", "glm-4.5-air"}

	// The set the schema renders is derived, not re-declared. Assert the derived
	// accessor equals the SPEC set first, so a drift in the underlying constants
	// fails here rather than silently changing every tier widget.
	if got := config.ValidGLMModels(); !equalStrings(got, want) {
		t.Errorf("config.ValidGLMModels() = %v, want %v", got, want)
	}

	for _, tier := range glmTierKeys {
		name := "llm.glm.models." + tier
		f := fieldByName(t, name)
		if f.Type != settings.TypeSelect {
			t.Errorf("%s Type = %q, want select (a closed model set gets a closed widget)", name, f.Type)
		}
		if got := optionValues(f); !equalStrings(got, want) {
			t.Errorf("%s options = %v, want %v", name, got, want)
		}
		if f.Validate == nil {
			t.Errorf("%s has no Validate — the closed set is not enforced on save", name)
		} else if f.Validate("not-a-glm-model") {
			t.Errorf("%s Validate accepted an out-of-set model", name)
		}
	}
}

// TestGLMFlashOptionLabelsAllLocales pins the flash option labels: the two
// option-key families (tier-slot selects + audit GLM-model select) must each
// carry a non-empty "glm-5.3-flash" label in all four locales — exactly 8
// entries across the dictionary, one per (key family × locale).
func TestGLMFlashOptionLabelsAllLocales(t *testing.T) {
	dict := readEmbeddedAsset(t, "i18n.js")
	cat, err := parseI18nCatalogue(dict)
	if err != nil {
		t.Fatalf("parse i18n catalogue: %v", err)
	}
	keyFamilies := []string{
		"f.llm.glm.models.opt.",
		"f.workflow.audit.glm.model.opt.",
	}
	total := 0
	for _, fam := range keyFamilies {
		key := fam + "glm-5.3-flash"
		for _, loc := range []string{"en", "ko", "ja", "zh"} {
			v, ok := cat[loc][key]
			if !ok {
				t.Errorf("locale %q is missing the flash option label %q (blank/untranslated render risk)", loc, key)
				continue
			}
			if strings.TrimSpace(v) == "" {
				t.Errorf("locale %q defines %q as an EMPTY label", loc, key)
			}
			total++
		}
	}
	if total != 8 {
		t.Errorf("flash option label count = %d, want exactly 8 (2 key families x 4 locales)", total)
	}
}

// TestGLMEffortTierDefaults verifies AC-WCR-031: option set {Max, High, None}
// and per-tier defaults fable=Max, high=Max, medium=High, low=None.
//
// The three option values are the z.ai canonical reasoning states, derived from
// the template package's GLMState* constants — the same domain the runtime
// collapse function produces. "Max"/"High"/"None" are the DISPLAY labels for
// those states, carried by i18n, not a second value vocabulary.
func TestGLMEffortTierDefaults(t *testing.T) {
	wantOpts := []string{
		template.GLMStateMax,
		template.GLMStateHigh,
		template.GLMStateLow,
	}
	wantDefault := map[string]string{
		"high":   template.GLMStateHigh,
		"medium": template.GLMStateHigh,
		"low":    template.GLMStateLow,
		"fable":  template.GLMStateMax,
	}

	for _, tier := range glmTierKeys {
		name := "llm.glm.effort." + tier
		f := fieldByName(t, name)
		if f.Type != settings.TypeSelect {
			t.Errorf("%s Type = %q, want select", name, f.Type)
		}
		if got := optionValues(f); !equalStrings(got, wantOpts) {
			t.Errorf("%s options = %v, want %v", name, got, wantOpts)
		}
		if f.Default != wantDefault[tier] {
			t.Errorf("%s Default = %q, want %q", name, f.Default, wantDefault[tier])
		}
	}

	// The default must actually reach the widget: with no value on disk, the
	// rendered select preselects the default rather than falling back to the
	// first option for every tier.
	html := renderConsolePage(t)
	for _, tier := range []string{"medium", "low"} {
		name := "llm.glm.effort." + tier
		want := wantDefault[tier]
		if !strings.Contains(html, `<option value="`+want+`" selected`) {
			t.Errorf("%s: rendered page does not preselect the default %q", name, want)
		}
	}
}

// TestGLMTierLabelVsKey verifies AC-WCR-032: the form keys stay high/medium/
// low/fable while the visible labels are the Claude-facing tier names.
func TestGLMTierLabelVsKey(t *testing.T) {
	html := renderConsolePage(t)
	dict := readEmbeddedAsset(t, "i18n.js")

	for _, tier := range glmTierKeys {
		for _, prefix := range []string{"llm.glm.models.", "llm.glm.effort."} {
			name := prefix + tier
			if !strings.Contains(html, `name="`+name+`"`) {
				t.Errorf("form key %q is absent from the rendered page — the internal key changed", name)
			}
		}
	}

	// The i18n title for each tier field must carry the Claude-facing name.
	wantLabel := map[string]string{
		"high":   "Opus",
		"medium": "Sonnet",
		"low":    "Haiku",
		"fable":  "Fable",
	}
	for tier, claudeName := range wantLabel {
		for _, prefix := range []string{"f.llm.glm.models.", "f.llm.glm.effort."} {
			key := prefix + tier + ".title"
			line := i18nEnLine(t, dict, key)
			if line == "" {
				t.Errorf("i18n key %q is missing from the en locale", key)
				continue
			}
			if !strings.Contains(line, claudeName) {
				t.Errorf("i18n %q = %q, want it to name the Claude-facing tier %q", key, line, claudeName)
			}
		}
	}
}

// TestGLMEffortScopeBadge verifies the honesty requirement (AC-WCR-033
// lineage, updated by RC3 glm-settings-persist): the tab names WHAT the
// per-tier effort now does — applies at the next moai glm launch to the main
// session's slot, overriding the session-wide effort_level preference, with
// sub-agents keeping the session-wide value — and no longer marks the four
// per-tier fields stored-only.
func TestGLMEffortScopeBadge(t *testing.T) {
	html := renderConsolePage(t)
	dict := readEmbeddedAsset(t, "i18n.js")

	const badgeKey = "sec.llm.effortnote"
	if !strings.Contains(html, `data-i18n="`+badgeKey+`"`) {
		t.Errorf("3rd Party LLM panel does not render the effort-scope note (%s)", badgeKey)
	}
	// The note must render exactly once — a per-field repetition is the shape
	// this requirement exists to avoid.
	if n := strings.Count(html, `data-i18n="`+badgeKey+`"`); n != 1 {
		t.Errorf("effort-scope note renders %d times, want exactly 1 (panel header level)", n)
	}

	// The note must exist in all four locales.
	for _, loc := range []string{"en", "ko", "ja", "zh"} {
		if !i18nLocaleHasKey(t, dict, loc, badgeKey) {
			t.Errorf("locale %q is missing the effort-scope note key %q", loc, badgeKey)
		}
	}

	// The en text must name the application point (the moai glm launch), the
	// preference it overrides (effort_level), and the sub-agent carve-out —
	// and must not regress into the retired stored-only claim.
	en := i18nEnLine(t, dict, badgeKey)
	for _, want := range []string{"moai glm", "effort_level", "Sub-agents"} {
		if !strings.Contains(en, want) {
			t.Errorf("effort-scope note does not name %q: %q", want, en)
		}
	}
	for _, forbidden := range []string{"stored only", "applied per tier", "each tier is applied", "per-tier effort is applied"} {
		if strings.Contains(strings.ToLower(en), forbidden) {
			t.Errorf("effort-scope note carries the false claim %q: %q", forbidden, en)
		}
	}

	// The per-tier effort fields are load-bearing now — the stored-only badge
	// must be gone. (One data-store-only render anywhere fails the probe; the
	// schema sets StoreOnly on no field since RC3.)
	if strings.Contains(html, `data-store-only="llm.glm.effort.`) {
		t.Errorf("a per-tier effort field still renders the stored-only badge — the value is applied at the next moai glm launch")
	}
}

// TestGLMKeyNeverEchoedByDefault is the regression pin for the SPEC-GLM-KEY-
// INPUT-001 never-echo contract: adding the reveal control must not put the
// stored key into the default page render.
func TestGLMKeyNeverEchoedByDefault(t *testing.T) {
	dir := withSandboxedGLMHome(t)
	if err := glmcred.Save(sentinelKey); err != nil {
		t.Fatalf("seed credential: %v", err)
	}
	_ = dir

	a := newTestApp(t)
	body := renderSettingsGET(a)
	if strings.Contains(body, sentinelKey) {
		t.Error("the stored GLM key appears in the default page render — the never-echo contract regressed")
	}
	if !strings.Contains(body, `value=""`) {
		t.Error("the GLM key input lost its unconditional empty value attribute")
	}
	// The reveal control must be present but must not carry the key.
	if !strings.Contains(body, glmKeyRevealPath) {
		t.Errorf("the page does not offer the reveal control (%s)", glmKeyRevealPath)
	}
}

// TestGLMKeyRevealLoopbackOnly verifies AC-WCR-034: an explicit same-origin
// loopback POST returns the plaintext once; a non-loopback origin is refused.
func TestGLMKeyRevealLoopbackOnly(t *testing.T) {
	withSandboxedGLMHome(t)
	if err := glmcred.Save(sentinelKey); err != nil {
		t.Fatalf("seed credential: %v", err)
	}
	h := newTestApp(t).routes()

	t.Run("loopback returns plaintext", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, glmKeyRevealPath, nil)
		req.Header.Set("Sec-Fetch-Site", "same-origin")
		req.Host = "127.0.0.1:8080"
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("reveal status = %d, want 200", rec.Code)
		}
		body, _ := io.ReadAll(rec.Body)
		if strings.TrimSpace(string(body)) != sentinelKey {
			t.Errorf("reveal body = %q, want the stored key", string(body))
		}
	})

	t.Run("non-loopback host refused", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, glmKeyRevealPath, nil)
		req.Header.Set("Sec-Fetch-Site", "same-origin")
		req.Host = "evil.example.com"
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)
		if rec.Code == http.StatusOK {
			t.Error("a non-loopback Host was served the plaintext key")
		}
		body, _ := io.ReadAll(rec.Body)
		if strings.Contains(string(body), sentinelKey) {
			t.Error("the plaintext key leaked to a non-loopback origin")
		}
	})

	t.Run("cross-site refused", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, glmKeyRevealPath, nil)
		req.Header.Set("Sec-Fetch-Site", "cross-site")
		req.Host = "127.0.0.1:8080"
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)
		if rec.Code == http.StatusOK {
			t.Error("a cross-site POST was served the plaintext key")
		}
	})

	t.Run("GET refused", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, glmKeyRevealPath, nil)
		req.Host = "127.0.0.1:8080"
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)
		if rec.Code == http.StatusOK {
			t.Error("GET was served the plaintext key — the reveal must be an explicit POST action")
		}
	})
}

// TestGLMKeyRevealAbsentCredential pins the no-credential path: the endpoint
// reports "not configured" rather than an empty 200 a caller could mistake for
// an empty key.
func TestGLMKeyRevealAbsentCredential(t *testing.T) {
	withSandboxedGLMHome(t)
	h := newTestApp(t).routes()

	req := httptest.NewRequest(http.MethodPost, glmKeyRevealPath, nil)
	req.Header.Set("Sec-Fetch-Site", "same-origin")
	req.Host = "127.0.0.1:8080"
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Errorf("reveal with no stored key: status = %d, want 404", rec.Code)
	}
}

// equalStrings compares two string slices element-wise.
func equalStrings(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// i18nEnLine returns the en-locale value for key, or "" when absent. The
// catalogue is parsed rather than substring-matched so a key defined only in a
// non-en locale is not mistaken for an en definition.
func i18nEnLine(t *testing.T, dict, key string) string {
	t.Helper()
	cat, err := parseI18nCatalogue(dict)
	if err != nil {
		t.Fatalf("parse i18n catalogue: %v", err)
	}
	return cat["en"][key]
}

// i18nLocaleHasKey reports whether the given locale defines key.
func i18nLocaleHasKey(t *testing.T, dict, locale, key string) bool {
	t.Helper()
	cat, err := parseI18nCatalogue(dict)
	if err != nil {
		t.Fatalf("parse i18n catalogue: %v", err)
	}
	_, ok := cat[locale][key]
	return ok
}

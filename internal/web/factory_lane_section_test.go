package web

// factory_lane_section_test.go — SPEC-WEB-CONSOLE-015 M3: the rendered lane
// section, the corrected note banner, and the locale coverage of every string
// this SPEC introduces.
//
// The banner and the marker hover text are asserted for SURVIVAL as well as
// for removal: both greps below are satisfied by deleting the strings outright,
// which REQ-WC15-052 forbids — the explanation is what stops a reader misreading
// an honest blank as a bug.

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/kanban"
	"github.com/modu-ai/moai-adk/internal/session"
)

// kanbanBodyFor renders GET /kanban for a console served from projectRoot.
func kanbanBodyFor(t *testing.T, projectRoot string) string {
	t.Helper()
	a := newApp(Config{ProjectRoot: projectRoot, ProfileName: "default"})
	a.recordLastProfile = func(string) error { return nil }
	req := httptest.NewRequest(http.MethodGet, "/kanban", nil)
	rec := httptest.NewRecorder()
	a.routes().ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("GET /kanban status = %d, want 200\nbody:\n%s", rec.Code, rec.Body.String())
	}
	return rec.Body.String()
}

// TestLaneSectionRendersCompleteRow — AC-WC15-044: a resolved lane row carries
// its lane number, card identifier, SPEC identifier, session state and stage.
func TestLaneSectionRendersCompleteRow(t *testing.T) {
	root := t.TempDir()
	pid := os.Getpid()
	writeFactoryRegistry(t, root, map[string]int{"lane-2": pid})
	writeActiveSessions(t, root, []session.Entry{liveEntry("sess-lane-2", pid)})
	writeKanbanRecord(t, root, kanban.Record{
		SessionID: "sess-lane-2", SpecID: "SPEC-EXAMPLE-001", Role: "lane",
		Backend: kanban.BackendGLM, Lane: 2, CardID: "t207",
	})

	body := kanbanBodyFor(t, root)
	for _, want := range []string{`data-lane="2"`, "t207", "SPEC-EXAMPLE-001", "state--live", "stage--active"} {
		if !strings.Contains(body, want) {
			t.Errorf("lane row missing %q", want)
		}
	}
}

// TestLaneSectionMarksEstimatedStageOnlyWhenEstimated — AC-WC15-045: the
// estimated marker is on the estimated row and NOT on the unestimated one.
// The second half is what stops the marker rendering unconditionally, which
// would pass the first half while observing nothing.
func TestLaneSectionMarksEstimatedStageOnlyWhenEstimated(t *testing.T) {
	// estimateStage returns estimated=true for a live/stale session and
	// estimated=false only when there is no session — which is also the
	// unresolved case, so the two halves are asserted on the view model the
	// section renders from.
	live := SessionVM{State: StateLive}
	if stage, estimated := estimateStage(live); !estimated || stage != StageActive {
		t.Fatalf("live session: stage=%q estimated=%v, want active/true", stage, estimated)
	}
	none := SessionVM{}
	if stage, estimated := estimateStage(none); estimated || stage != StageBlocked {
		t.Fatalf("absent session: stage=%q estimated=%v, want blocked/false", stage, estimated)
	}

	root := t.TempDir()
	pid := os.Getpid()
	writeFactoryRegistry(t, root, map[string]int{"lane-1": pid})
	writeActiveSessions(t, root, []session.Entry{liveEntry("sess-1", pid)})
	writeKanbanRecord(t, root, kanban.Record{SessionID: "sess-1", Role: "lane", Lane: 1, CardID: "t1"})

	if body := kanbanBodyFor(t, root); !strings.Contains(body, `data-i18n="mark.estimated"`) {
		t.Error("estimated lane row carries no estimated marker")
	}

	// An unresolved lane asserts no stage at all, so it carries no stage mark
	// to mislabel — the row shows the unresolved marker instead.
	bare := t.TempDir()
	writeFactoryRegistry(t, bare, map[string]int{"lane-1": 999101})
	body := kanbanBodyFor(t, bare)
	if !strings.Contains(body, "lane-unresolved") {
		t.Error("unresolved lane row carries no unresolved marker")
	}
	if strings.Contains(body, `data-i18n="mark.estimated"`) {
		t.Error("unresolved lane row carries an estimated stage marker it cannot know")
	}
}

// TestLaneSectionPresentWithNoRegistry — AC-WC15-046: an absent registry and a
// malformed one both yield status 200 AND a factory section rendering zero
// lanes. "The page still loads" is preserved; "the section handled the failure"
// is the new half.
func TestLaneSectionPresentWithNoRegistry(t *testing.T) {
	t.Run("absent", func(t *testing.T) {
		body := kanbanBodyFor(t, t.TempDir())
		if !strings.Contains(body, `data-i18n="kanban.lanes"`) {
			t.Error("factory section absent from the markup")
		}
		if !strings.Contains(body, `data-i18n="kanban.noLanes"`) {
			t.Error("zero-lane state not rendered")
		}
	})
	t.Run("malformed", func(t *testing.T) {
		root := t.TempDir()
		path := kanban.FactoryRegistryPath(root)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatalf("mkdir: %v", err)
		}
		if err := os.WriteFile(path, []byte("{not json"), 0o600); err != nil {
			t.Fatalf("write: %v", err)
		}
		body := kanbanBodyFor(t, root)
		if !strings.Contains(body, `data-i18n="kanban.lanes"`) {
			t.Error("factory section absent from the markup")
		}
		if !strings.Contains(body, `data-i18n="kanban.noLanes"`) {
			t.Error("zero-lane state not rendered")
		}
	})
}

// TestLaneSectionRendersUnresolvedRows — AC-WC15-043b / AC-WC15-047 at the
// render surface: an unresolvable lane is present, numbered, and marked — never
// dropped, never blank, and never carrying another lane's record.
func TestLaneSectionRendersUnresolvedRows(t *testing.T) {
	root := t.TempDir()
	pid := os.Getpid()
	writeFactoryRegistry(t, root, map[string]int{"lane-1": pid, "lane-5": pid})
	writeActiveSessions(t, root, []session.Entry{liveEntry("sess-dup", pid)})
	writeKanbanRecord(t, root, kanban.Record{
		SessionID: "sess-dup", SpecID: "SPEC-EXAMPLE-001", Role: "lane",
		Lane: 1, CardID: "t999",
	})

	body := kanbanBodyFor(t, root)
	for _, want := range []string{`data-lane="1"`, `data-lane="5"`} {
		if !strings.Contains(body, want) {
			t.Errorf("ambiguous lane row %q dropped from the markup", want)
		}
	}
	if strings.Contains(body, "t999") {
		t.Error("an ambiguous pid attributed a card id to a lane row")
	}
	if !strings.Contains(body, `data-lane-unresolved="ambiguous"`) {
		t.Error("ambiguous rows do not name their unresolved reason")
	}
}

// TestKanbanNoteBannerCorrected — AC-WC15-052 removal + survival + translation.
// The removal half alone is satisfiable by deletion, which the requirement
// forbids, so survival and the translation key are asserted beside it.
func TestKanbanNoteBannerCorrected(t *testing.T) {
	sources := map[string]string{
		"screens.templ": readSource(t, "screens.templ"),
		"widgets.templ": readSource(t, "widgets.templ"),
	}
	for _, stale := range []string{
		"are not recorded yet",
		"kanban.Record is extended",
		"kanban.Record extension required",
	} {
		for name, src := range sources {
			if strings.Contains(src, stale) {
				t.Errorf("%s still carries the falsified string %q", name, stale)
			}
		}
	}

	// Survival: the banner is present in the rendered page and carries a key.
	body := kanbanBodyFor(t, t.TempDir())
	if !strings.Contains(body, `data-i18n="kanban.note"`) {
		t.Error("the kanban note banner is absent or carries no translation key")
	}

	// Survival: the not-recorded marker still carries non-empty hover text, and
	// that text now carries a translation key of its own.
	if !strings.Contains(sources["widgets.templ"], `data-i18n-title="mark.notRecorded"`) {
		t.Error("the not-recorded marker carries no translated hover text")
	}
	if !strings.Contains(body, `data-i18n-title="mark.notRecorded"`) {
		t.Error("the rendered marker carries no translated hover text")
	}
	if strings.Contains(body, `title=""`) {
		t.Error("a marker rendered with empty hover text")
	}
}

// readSource reads one internal/web source file for a source-level assertion.
func readSource(t *testing.T, name string) string {
	t.Helper()
	body, err := os.ReadFile(name)
	if err != nil {
		t.Fatalf("read %s: %v", name, err)
	}
	return string(body)
}

// TestSPECIntroducedKeysResolveInEveryLocale — AC-WC15-050: the key set this
// SPEC introduces is NON-EMPTY and every member resolves in all four maps. An
// empty set would mean the strings were hard-coded in English instead.
func TestSPECIntroducedKeysResolveInEveryLocale(t *testing.T) {
	introduced := []string{
		"kanban.lanes",
		"kanban.noLanes",
		"kanban.lane",
		"kanban.laneCard",
		"kanban.laneSpec",
		"kanban.laneUnresolved",
		"kanban.laneUnresolved.no-session",
		"kanban.laneUnresolved.ambiguous",
		"kanban.laneUnresolved.no-record",
		"kanban.note",
		"mark.notRecorded",
	}
	if len(introduced) == 0 {
		t.Fatal("this SPEC introduces user-visible strings; an empty key set means they were hard-coded")
	}

	cat := shippedI18nCatalogue(t)
	for _, key := range introduced {
		for _, loc := range i18nLocaleOrder {
			v, ok := cat[loc][key]
			if !ok || strings.TrimSpace(v) == "" {
				t.Errorf("key %q missing from locale %q", key, loc)
			}
		}
		// A non-en value identical to en would need an allowlist entry, which
		// this SPEC is forbidden from adding — so require a real translation.
		for _, loc := range i18nNonEnLocales {
			if cat[loc][key] == cat["en"][key] {
				t.Errorf("key %q is untranslated in %q (identical to en)", key, loc)
			}
		}
	}
}

// TestLaneSectionAddsNoTransportSurface — REQ-WC15-001 at the markup level: the
// lane section rides the kanban area the page already declares; it introduces
// no second live area and therefore no new event name.
func TestLaneSectionAddsNoTransportSurface(t *testing.T) {
	root := t.TempDir()
	writeFactoryRegistry(t, root, map[string]int{"lane-1": 999102})
	body := kanbanBodyFor(t, root)

	areas := strings.Count(body, `data-live="`)
	plain := kanbanBodyFor(t, t.TempDir())
	if areas != strings.Count(plain, `data-live="`) {
		t.Errorf("live-area count changed with lanes present: %d vs %d",
			areas, strings.Count(plain, `data-live="`))
	}
	if strings.Contains(body, `data-live="factory"`) || strings.Contains(body, `data-live="lane"`) {
		t.Error("the lane section declared a new live area — transport is out of scope")
	}
	_ = time.Now
}

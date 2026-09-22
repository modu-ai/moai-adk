package web

// shell_chrome_states_test.go — TG-4 (SPEC-WEB-CONSOLE-018 REQ-008).
//
// Defect answer: these tests catch navigation losing its active-tab marking,
// the live-state cluster emitting a stale state, the save cluster dropping the
// failure surface (the SPEC-WEB-CONSOLE-017 observability contract), or a
// widget drawing a recorded value over an unrecorded one.

import (
	"strings"
	"testing"
)

// tg4ShellVM is a shell state with the settings-area fields populated.
func tg4ShellVM(area string) ShellVM {
	return ShellVM{
		Area: area, Tab: "identity", Title: "T", Crumb: "c",
		Host: "127.0.0.1:3041", Profile: "default", Project: "proj",
		ProjectPath: "/tmp/proj", Lang: "en", Live: "on", RenderedAt: "12:00:00",
	}
}

// TestShellMarksExactlyOneActiveTab pins the rail contract: the rail carries
// exactly three nav rows (overview / todo / settings), and for each of those
// areas exactly one row carries aria-current — zero means the reader loses
// their place, two means a doubled marking. The read-only screen areas
// (kanban / specs / monitor) own no rail row today, so the contract for them is
// zero: if a row is ever added for one of those areas, this half of the pin is
// the one to update, together with the row itself.
func TestShellMarksExactlyOneActiveTab(t *testing.T) {
	for _, area := range []string{"overview", "todo", "settings"} {
		html := renderTempl(t, Shell(tg4ShellVM(area)))
		if n := strings.Count(html, `aria-current="page"`); n != 1 {
			t.Errorf("area %q: %d nav rows marked active, want exactly 1:\n%s", area, n, html)
		}
	}
	for _, area := range []string{"kanban", "specs", "monitor"} {
		html := renderTempl(t, Shell(tg4ShellVM(area)))
		if n := strings.Count(html, `aria-current="page"`); n != 0 {
			t.Errorf("area %q: %d nav rows marked active, want 0 (no rail row owns this area):\n%s", area, n, html)
		}
	}
}

// TestNavRowCurrentAndEcho pins the navRow branches: the active row keeps its
// English echo span, and both rows render their i18n key — a renamed area id
// would silently orphan its translation and its icon.
func TestNavRowCurrentAndEcho(t *testing.T) {
	active := renderTempl(t, navRow(tg4ShellVM("todo"), "todo", "Todo", "/todo"))
	if !strings.Contains(active, `data-i18n="nav.todo"`) {
		t.Errorf("the active row lost its i18n key:\n%s", active)
	}
	idle := renderTempl(t, navRow(tg4ShellVM("todo"), "kanban", "Kanban", "/kanban"))
	if strings.Contains(idle, `aria-current="page"`) {
		t.Errorf("an inactive row carried the current marker:\n%s", idle)
	}
}

// TestTopbarAreaBranches pins the topbar: settings renders the save cluster,
// every other area renders the live indicator, and a crumb renders only when
// the viewmodel carries one.
func TestTopbarAreaBranches(t *testing.T) {
	settingsTop := renderTempl(t, topbar(tg4ShellVM("settings")))
	if !strings.Contains(settingsTop, `data-save-state`) || strings.Contains(settingsTop, "data-live-indicator") {
		t.Errorf("the settings topbar did not render the save cluster:\n%s", settingsTop)
	}
	monitorTop := renderTempl(t, topbar(tg4ShellVM("monitor")))
	if !strings.Contains(monitorTop, "data-live-indicator") || strings.Contains(monitorTop, "data-save-state") {
		t.Errorf("a read-only screen rendered the save cluster instead of live state:\n%s", monitorTop)
	}
	crumbless := renderTempl(t, topbar(func() ShellVM { vm := tg4ShellVM("monitor"); vm.Crumb = ""; return vm }()))
	if strings.Contains(crumbless, `top__crumb`) {
		t.Errorf("an empty crumb rendered a crumb element:\n%s", crumbless)
	}
}

// TestLiveStateOnOff pins the live indicator branches: on renders the
// timestamp marker, off renders the degraded role=status polling notice — a
// lost stream that still reads "Live" lies about the data's freshness.
func TestLiveStateOnOff(t *testing.T) {
	on := renderTempl(t, liveState("on", "12:00:05"))
	for _, want := range []string{`data-live-indicator="on"`, `data-live-rendered-at="12:00:05"`, `12:00:05`} {
		if !strings.Contains(on, want) {
			t.Errorf("live-on missing %q:\n%s", want, on)
		}
	}
	off := renderTempl(t, liveState("off", ""))
	for _, want := range []string{`data-live-indicator="off"`, `role="status"`, `live--off`} {
		if !strings.Contains(off, want) {
			t.Errorf("live-off missing %q:\n%s", want, off)
		}
	}
	if strings.Contains(off, "12:00:05") {
		t.Errorf("a lost stream rendered a freshness timestamp:\n%s", off)
	}
}

// TestSaveClusterAllStates pins every save-state branch: clean, dirty with the
// field count, saving with the disabled submit, saved with the timestamp, and
// error with the server message — falling back to the generic wording only when
// the server gave none. The error branch is the SPEC-WEB-CONSOLE-017 surface:
// a save failure with no visible explanation silently did nothing.
func TestSaveClusterAllStates(t *testing.T) {
	render := func(vm ShellVM) string { return renderTempl(t, saveCluster(vm)) }

	cleanVM := tg4ShellVM("settings")
	cleanVM.SaveState = "clean"
	clean := render(cleanVM)
	if !strings.Contains(clean, `data-save-state="clean"`) || !strings.Contains(clean, `data-i18n="save.clean"`) {
		t.Errorf("clean state lost its branch:\n%s", clean)
	}

	dirtyVM := tg4ShellVM("settings")
	dirtyVM.SaveState = "dirty"
	dirtyVM.DirtyCount = 3
	dirty := render(dirtyVM)
	if !strings.Contains(dirty, `data-save-state="dirty"`) || !strings.Contains(dirty, "3") {
		t.Errorf("dirty state lost its field count:\n%s", dirty)
	}

	savingVM := tg4ShellVM("settings")
	savingVM.SaveState = "saving"
	saving := render(savingVM)
	if !strings.Contains(saving, `data-save-state="saving"`) || !strings.Contains(saving, "disabled") {
		t.Errorf("saving state did not disable the submit:\n%s", saving)
	}

	savedVM := tg4ShellVM("settings")
	savedVM.SaveState = "saved"
	savedVM.SavedAt = "12:01:00"
	saved := render(savedVM)
	if !strings.Contains(saved, `data-save-state="saved"`) || !strings.Contains(saved, "12:01:00") {
		t.Errorf("saved state lost its timestamp:\n%s", saved)
	}

	errVM := tg4ShellVM("settings")
	errVM.SaveState = "error"
	errVM.SaveMessage = "coverage gate failed"
	errored := render(errVM)
	if !strings.Contains(errored, `data-save-state="error"`) || !strings.Contains(errored, "coverage gate failed") || !strings.Contains(errored, `role="alert"`) {
		t.Errorf("error state lost the server message:\n%s", errored)
	}

	errVM.SaveMessage = ""
	generic := render(errVM)
	if !strings.Contains(generic, `data-i18n="save.error"`) {
		t.Errorf("a message-less error lost the generic fallback:\n%s", generic)
	}
}

// TestSaveClusterDirtyAndErrorUsePrimaryButton pins the variant mapping: dirty
// and error states escalate the save button to the primary variant so the
// pending action is visible; calm states keep the outline.
func TestSaveClusterDirtyAndErrorUsePrimaryButton(t *testing.T) {
	vm := tg4ShellVM("settings")
	clean := renderTempl(t, saveCluster(vm))
	if !strings.Contains(clean, "btn--outline") {
		t.Errorf("clean state lost the outline variant:\n%s", clean)
	}
	vm.SaveState = "error"
	errored := renderTempl(t, saveCluster(vm))
	if !strings.Contains(errored, "btn--primary") {
		t.Errorf("error state did not escalate the save button:\n%s", errored)
	}
}

// TestProfilePopRenameTargets pins the popover gating: rename/delete forms
// render only when an eligible target exists — offering a rename the server
// will refuse invites a round-trip that cannot succeed.
func TestProfilePopRenameTargets(t *testing.T) {
	vm := tg4ShellVM("settings")
	vm.Profiles = []ProfileVM{{Name: "default", Current: true}, {Name: "work", Current: false}}
	vm.RenameTargets = []string{"work"}
	with := renderTempl(t, profilePop(vm))
	for _, want := range []string{`action="/profile/rename"`, `action="/profile/delete"`, `action="/profile/create"`, `>work<`} {
		if !strings.Contains(with, want) {
			t.Errorf("popover with targets missing %q:\n%s", want, with)
		}
	}
	vm.RenameTargets = nil
	without := renderTempl(t, profilePop(vm))
	if strings.Contains(without, "/profile/rename") || strings.Contains(without, "/profile/delete") {
		t.Errorf("popover rendered rename/delete with no eligible target:\n%s", without)
	}
}

// TestPanelMetaVariants pins the panel header: a metaKey renders the meta as an
// i18n composite with its params, a plain meta renders verbatim, an empty meta
// renders nothing, and an empty help renders no help paragraph.
func TestPanelMetaVariants(t *testing.T) {
	keyed := renderTempl(t, panel("Sessions", "registry 5", "panelMeta.registry", "5", "", "roomy"))
	for _, want := range []string{`data-i18n="panelMeta.registry"`, `data-i18n-params="5"`, `registry 5`} {
		if !strings.Contains(keyed, want) {
			t.Errorf("keyed meta missing %q:\n%s", want, keyed)
		}
	}
	plain := renderTempl(t, panel("Sessions", "registry 5", "", "", "", "roomy"))
	if !strings.Contains(plain, `<span class="panel__meta">registry 5</span>`) {
		t.Errorf("plain meta lost its verbatim render:\n%s", plain)
	}
	bare := renderTempl(t, panel("Sessions", "", "", "", "", "dense"))
	if strings.Contains(bare, "panel__meta") || strings.Contains(bare, "panel__help") {
		t.Errorf("an empty meta/help rendered chrome:\n%s", bare)
	}
}

// TestNoteBannerVariants pins the banner kinds: warn renders the alert icon and
// role=alert, info renders the note icon and role=note, and a key attaches the
// i18n hook while a keyless banner renders its text verbatim.
func TestNoteBannerVariants(t *testing.T) {
	warn := renderTempl(t, noteBanner("warn", "careful", "note.warn"))
	for _, want := range []string{`banner--warn`, `role="alert"`, `data-i18n="note.warn"`} {
		if !strings.Contains(warn, want) {
			t.Errorf("warn banner missing %q:\n%s", want, warn)
		}
	}
	info := renderTempl(t, noteBanner("info", "fyi", ""))
	for _, want := range []string{`role="note"`, `>fyi<`} {
		if !strings.Contains(info, want) {
			t.Errorf("info banner missing %q:\n%s", want, info)
		}
	}
	if strings.Contains(info, `data-i18n=`) {
		t.Errorf("a keyless banner emitted an i18n hook:\n%s", info)
	}
}

// TestGaugeNegativeIsMissing pins the gauge: a negative percentage draws the
// missing glyph (unrecorded), never a 0% meter — an unrecorded context drawn as
// 0% reads as an empty context.
func TestGaugeNegativeIsMissing(t *testing.T) {
	none := renderTempl(t, gauge("CW", -1))
	if !strings.Contains(none, `class="missing"`) {
		t.Errorf("an unrecorded gauge did not draw the missing glyph:\n%s", none)
	}
	if strings.Contains(none, `role="meter"`) {
		t.Errorf("an unrecorded gauge drew a meter:\n%s", none)
	}
	measured := renderTempl(t, gauge("CW", 42))
	if !strings.Contains(measured, `aria-valuenow="42"`) || !strings.Contains(measured, "width:42%") {
		t.Errorf("a measured gauge lost its value:\n%s", measured)
	}
}

// TestStageMarkEstimateTag pins the estimate tag: an estimated stage carries
// the est wrapper and its tag, an exact stage renders the inner mark alone —
// a guess presented as fact is exactly what the tag exists to prevent.
func TestStageMarkEstimateTag(t *testing.T) {
	estimated := renderTempl(t, stageMark(StageActive, true))
	if !strings.Contains(estimated, `class="est"`) || !strings.Contains(estimated, `data-i18n="mark.estimated"`) {
		t.Errorf("an estimated stage lost its estimate tag:\n%s", estimated)
	}
	exact := renderTempl(t, stageMark(StageBlocked, false))
	if strings.Contains(exact, `class="est"`) {
		t.Errorf("an exact stage carried the estimate tag:\n%s", exact)
	}
	if !strings.Contains(exact, `stage--blocked`) {
		t.Errorf("the exact stage lost its class:\n%s", exact)
	}
}

// TestStateMarkLabelBranch pins the state mark: a label renders the i18n hook
// beside the dot, an empty label renders the dot alone.
func TestStateMarkLabelBranch(t *testing.T) {
	labeled := renderTempl(t, stateMark(StateLive, "Active"))
	if !strings.Contains(labeled, `data-i18n="state.live"`) {
		t.Errorf("a labeled state mark lost its i18n hook:\n%s", labeled)
	}
	bare := renderTempl(t, stateMark(StateLive, ""))
	if strings.Contains(bare, "data-i18n") {
		t.Errorf("an unlabeled state mark emitted an i18n hook:\n%s", bare)
	}
}

// TestBackendBadgeBranches pins the backend vocabulary: empty draws the missing
// glyph, glm carries the metered marker, claude the flat-rate marker — a
// metered backend labelled flat-rate misprices the session.
func TestBackendBadgeBranches(t *testing.T) {
	if html := renderTempl(t, backendBadge("")); !strings.Contains(html, `class="missing"`) {
		t.Errorf("an unrecorded backend did not draw the missing glyph:\n%s", html)
	}
	glm := renderTempl(t, backendBadge("glm"))
	if !strings.Contains(glm, "backend--metered") || !strings.Contains(glm, `data-i18n="backend.metered"`) {
		t.Errorf("glm backend lost its metered marker:\n%s", glm)
	}
	claude := renderTempl(t, backendBadge("claude"))
	if !strings.Contains(claude, `data-i18n="backend.flat"`) {
		t.Errorf("claude backend lost its flat-rate marker:\n%s", claude)
	}
}

// TestBadgeKindsAndMissing pins the badge kind classes and the missing glyph's
// not-recorded explanation.
func TestBadgeKindsAndMissing(t *testing.T) {
	if html := renderTempl(t, badge("danger", "MUST-FIX")); !strings.Contains(html, "badge--danger") {
		t.Errorf("danger badge lost its class:\n%s", html)
	}
	if html := renderTempl(t, badge("other", "x")); strings.Contains(html, "badge--") {
		t.Errorf("an unknown kind invented a variant class:\n%s", html)
	}
	if html := renderTempl(t, missing()); !strings.Contains(html, `data-i18n-title="mark.notRecorded"`) {
		t.Errorf("the missing glyph lost its explanation:\n%s", html)
	}
}

// TestLaneUnresolvedRendersReason pins the unresolved-lane widget: the reason
// rides both the attribute and the per-reason i18n key — a bare "unresolved"
// hides whether the lane has no session or no attributable record.
func TestLaneUnresolvedRendersReason(t *testing.T) {
	html := renderTempl(t, laneUnresolved("no-session"))
	for _, want := range []string{`data-lane-unresolved="no-session"`, `kanban.laneUnresolved.no-session`} {
		if !strings.Contains(html, want) {
			t.Errorf("unresolved lane missing %q:\n%s", want, html)
		}
	}
}

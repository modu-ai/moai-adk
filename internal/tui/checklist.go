// Package tui — checklist.go
//
// SPEC-TUI-CHECKLIST-001: a group-checklist progress renderer. A command opens
// one item at a time under a titled group; the running item is one line that
// is redrawn in place, and a finished item is one line that is never redrawn.
// Failure, warning and "keep" detail is expanded beneath the item; everything
// else stays collapsed.
//
// The renderer knows nothing about any command. Group keys, item keys, the
// header key and every summary field come from the caller, and every string it
// draws is a "checklist.*" catalogue key resolved through Translate. It always
// draws in rich mode: deciding whether to draw at all (TTY, CI, --verbose,
// TERM=dumb) belongs to the caller, and the fallback output is the caller's
// existing code path.
//
// It reuses the package primitives — Glyph* runes, Spinner, Theme tokens,
// downsample, progressClearPrefix and the CJK-aware StringWidth — and starts no
// goroutine or ticker: the spinner advances only when the caller reports an
// event, exactly like the stateless Spinner it wraps.
package tui

import (
	"fmt"
	"io"
	"strings"
	"time"

	"charm.land/lipgloss/v2"

	tuiinternal "github.com/modu-ai/moai-adk/internal/tui/internal"
)

const (
	// checklistItemIndent is the indent of an item line and of the header's
	// environment line.
	checklistItemIndent = "  "
	// checklistDetailIndent is the indent of expanded detail lines.
	checklistDetailIndent = "    "
	// checklistGroupMarker is the group-title and header marker. It sits in
	// the status-glyph whitelist (✓ ✗ ! · ● ○ ◆ ◇).
	checklistGroupMarker = "◆"
	// checklistLabelColumn is the display width a skipped item's label is
	// padded to, so the trailing reason lines up whatever the label's script.
	checklistLabelColumn = 30
	// checklistSeparator joins the fields of the environment line and the
	// summary headline.
	checklistSeparator = " · "
)

// EnvSummary is the environment line drawn under the header, for example
// "darwin/arm64 · v3.3.0 · ~/work/demo". Empty fields are omitted.
type EnvSummary struct {
	Platform string
	Version  string
	Dir      string
}

// ChecklistOptions configures NewChecklist.
type ChecklistOptions struct {
	// Out receives every byte the renderer draws. The caller picks the
	// channel.
	Out io.Writer
	// Theme provides the colour tokens. nil selects LightTheme, matching the
	// other stateless primitives; MonochromeTheme draws no SGR colour.
	Theme *Theme
	// Lang is the catalogue language. An empty or unsupported value falls back
	// to en inside Translate.
	Lang string
	// Clock supplies the time for the elapsed field. nil selects time.Now;
	// golden tests inject a fixed clock.
	Clock func() time.Time
}

// ChecklistSummary is the block that ends a completed run. Pointer counts are
// optional: nil omits the field. The warning count and the elapsed time are
// always drawn; the elapsed time is measured on the injected clock from the
// moment the header was drawn.
type ChecklistSummary struct {
	// Failed selects the "failed" headline instead of "done".
	Failed bool
	// CreatedFiles and CreatedDirs are the init-shaped counts.
	CreatedFiles *int
	CreatedDirs  *int
	// Changed and Preserved are the update-shaped counts.
	Changed   *int
	Preserved *int
	// Warnings is the warning count the caller defines (decision D-9).
	Warnings int
	// DeployMode, BackupPath, RecoverCmd and NextSteps are optional lines.
	DeployMode string
	BackupPath string
	RecoverCmd string
	NextSteps  string
	// OutsideWarnings lists warnings recorded while no item was open.
	OutsideWarnings []string
}

// Checklist draws a group checklist. It is not safe for concurrent use.
//
// @MX:ANCHOR: [AUTO] Checklist is the one group-checklist renderer; the init
// session and reporter (and the update session in a follow-up card) call it.
// @MX:REASON: [AUTO] A single renderer keeps the redraw rule, the glyph
// whitelist and the collapse rules in one place for every command that wires
// a checklist.
type Checklist struct {
	out   io.Writer
	theme Theme
	lang  string
	clock func() time.Time

	headerKey    string
	env          EnvSummary
	hasHeader    bool
	headerDrawn  bool
	started      time.Time
	currentGroup string
	open         *ChecklistItem
	// folded holds the entries of finished items that finish did not draw, in
	// item order, until FlushFolded writes them.
	folded []string
}

// NewChecklist builds a renderer that draws to opts.Out.
func NewChecklist(opts ChecklistOptions) *Checklist {
	th := LightTheme()
	if opts.Theme != nil {
		th = *opts.Theme
	}
	clock := opts.Clock
	if clock == nil {
		clock = time.Now
	}
	return &Checklist{
		out:   opts.Out,
		theme: th,
		lang:  opts.Lang,
		clock: clock,
	}
}

// Header records the header line and the environment line. Nothing is drawn
// until the first item opens (or the summary is drawn), so a run that opens no
// item draws no header.
func (c *Checklist) Header(titleKey string, env EnvSummary) {
	c.headerKey = titleKey
	c.env = env
	c.hasHeader = true
}

// Item opens the item itemKey under groupKey and draws its running line. An
// item that is still open is closed as done first, so call sites never pair
// every open with a close. A group title is drawn when the group changes.
func (c *Checklist) Item(groupKey, itemKey string) *ChecklistItem {
	if c.open != nil {
		c.open.Done()
	}
	c.drawHeader()
	if groupKey != c.currentGroup {
		c.currentGroup = groupKey
		c.writeln(c.paint(c.theme.Accent, checklistGroupMarker+" "+c.tr("checklist.group."+groupKey)))
	}
	it := &ChecklistItem{
		c:     c,
		label: c.tr("checklist.item." + groupKey + "." + itemKey),
	}
	c.open = it
	it.drawRunning(false)
	return it
}

// Open returns the item that is currently open, or nil when none is. A
// command-side session uses it to route captured output to the open item.
func (c *Checklist) Open() *ChecklistItem { return c.open }

// Summary closes any open item and draws the summary block.
func (c *Checklist) Summary(s ChecklistSummary) {
	if c.open != nil {
		c.open.Done()
	}
	c.drawHeader()

	elapsed := c.clock().Sub(c.started)
	head := c.tr("checklist.summary.done")
	headColor := c.theme.Success
	if s.Failed {
		head = c.tr("checklist.summary.failed")
		headColor = c.theme.Danger
	}
	parts := []string{c.paint(headColor, head)}
	for _, f := range []struct {
		key string
		n   *int
	}{
		{"checklist.summary.created_files", s.CreatedFiles},
		{"checklist.summary.created_dirs", s.CreatedDirs},
		{"checklist.summary.changed", s.Changed},
		{"checklist.summary.preserved", s.Preserved},
	} {
		if f.n != nil {
			parts = append(parts, fmt.Sprintf(c.tr(f.key), *f.n))
		}
	}
	parts = append(parts,
		fmt.Sprintf(c.tr("checklist.summary.warnings"), s.Warnings),
		fmt.Sprintf(c.tr("checklist.summary.elapsed"), formatChecklistElapsed(elapsed)),
	)

	c.writeln("")
	c.writeln(checklistItemIndent + strings.Join(parts, checklistSeparator))
	for _, l := range []struct {
		key string
		val string
	}{
		{"checklist.summary.deploy_mode", s.DeployMode},
		{"checklist.summary.backup", s.BackupPath},
		{"checklist.summary.recover", s.RecoverCmd},
		{"checklist.summary.next_steps", s.NextSteps},
	} {
		if l.val != "" {
			c.writeln(checklistItemIndent + fmt.Sprintf(c.tr(l.key), l.val))
		}
	}
	if len(s.OutsideWarnings) > 0 {
		c.writeln(checklistItemIndent + c.tr("checklist.summary.warnings_outside"))
		for _, w := range s.OutsideWarnings {
			c.writeln(checklistDetailIndent + c.glyph(c.theme.Warning, GlyphWarn) + " " + w)
		}
	}
}

// drawHeader draws the header and environment lines once. It is a no-op when
// no header was recorded or it was already drawn.
func (c *Checklist) drawHeader() {
	if c.headerDrawn {
		return
	}
	c.headerDrawn = true
	c.started = c.clock()
	if !c.hasHeader {
		return
	}
	c.writeln(c.paint(c.theme.Accent, checklistGroupMarker+" "+c.tr(c.headerKey)))
	var env []string
	for _, p := range []string{c.env.Platform, c.env.Version, c.env.Dir} {
		if p != "" {
			env = append(env, p)
		}
	}
	if len(env) > 0 {
		c.writeln(checklistItemIndent + c.paint(c.theme.Dim, strings.Join(env, checklistSeparator)))
	}
}

// FlushFolded writes, once and in item order, the entries of finished items
// that were never drawn, and forgets them. Finished item lines are not redrawn.
// An aborted run calls it so no captured line is lost (REQ-TCL-016).
func (c *Checklist) FlushFolded() {
	for _, l := range c.folded {
		c.writeln(checklistDetailIndent + l)
	}
	c.folded = nil
}

// tr resolves a catalogue key in the configured language.
func (c *Checklist) tr(key string) string { return Translate(key, c.lang) }

// writeln writes one finished line.
func (c *Checklist) writeln(s string) { _, _ = io.WriteString(c.out, s+"\n") }

// paint colours s with the given theme token. An empty token (MonochromeTheme)
// draws plain text so no SGR sequence is produced.
func (c *Checklist) paint(color, s string) string {
	if color == "" {
		return s
	}
	return downsample(lipgloss.NewStyle().Foreground(lipgloss.Color(color)).Render(s))
}

// glyph renders one status glyph in the given token colour.
func (c *Checklist) glyph(color string, r rune) string { return c.paint(color, string(r)) }

// formatChecklistElapsed renders an elapsed duration with one decimal, for
// example 1.4s.
func formatChecklistElapsed(d time.Duration) string {
	return fmt.Sprintf("%.1fs", d.Seconds())
}

// checklistEntryKind classifies a line recorded against an open item.
type checklistEntryKind int

const (
	entryDetail checklistEntryKind = iota
	entryWarn
	entryKeep
)

type checklistEntry struct {
	kind checklistEntryKind
	text string
}

// ChecklistItem is one open (or finished) checklist item.
type ChecklistItem struct {
	c       *Checklist
	label   string
	entries []checklistEntry
	done    bool
}

// Update advances the spinner one frame and redraws only this item's line.
func (i *ChecklistItem) Update() {
	if i.done {
		return
	}
	i.drawRunning(true)
}

// Detail records a collapsed detail line. It is shown only if the item fails.
func (i *ChecklistItem) Detail(line string) { i.record(entryDetail, line) }

// Warn records a warning. It is expanded beneath the item and the item ends
// with the warning glyph.
func (i *ChecklistItem) Warn(msg string) { i.record(entryWarn, msg) }

// Keep records a notice that stays visible even when the item succeeds. It
// changes neither the item glyph nor any count.
func (i *ChecklistItem) Keep(line string) { i.record(entryKeep, line) }

func (i *ChecklistItem) record(kind checklistEntryKind, text string) {
	if i.done {
		return
	}
	i.entries = append(i.entries, checklistEntry{kind: kind, text: text})
}

// Done finishes the item: a success glyph, or the warning glyph when a warning
// was recorded. Warnings and kept notices are expanded; plain details stay
// collapsed.
func (i *ChecklistItem) Done() {
	i.assertNotDone()
	warned := false
	for _, e := range i.entries {
		if e.kind == entryWarn {
			warned = true
			break
		}
	}
	glyph := i.c.glyph(i.c.theme.Success, GlyphDone)
	if warned {
		glyph = i.c.glyph(i.c.theme.Warning, GlyphWarn)
	}
	i.finish(glyph, nil, func(e checklistEntry) bool { return e.kind != entryDetail })
}

// Fail finishes the item as failed: the failure glyph, every detail line
// recorded while the item was open, then msg.
func (i *ChecklistItem) Fail(msg string) {
	i.assertNotDone()
	i.finish(i.c.glyph(i.c.theme.Danger, GlyphErr), []string{msg}, func(checklistEntry) bool { return true })
}

// Skip finishes the item as skipped, with the catalogue text reasonKey as the
// trailing reason. Kept notices are still expanded.
func (i *ChecklistItem) Skip(reasonKey string) {
	i.assertNotDone()
	reason := i.c.paint(i.c.theme.Dim, i.c.tr(reasonKey))
	i.label += strings.Repeat(" ", max(2, checklistLabelColumn-tuiinternal.StringWidth(i.label))) + reason
	i.finish(i.c.glyph(i.c.theme.Dim, GlyphSkip), nil, func(e checklistEntry) bool { return e.kind == entryKeep })
}

// finish redraws the item line once in its terminal state and expands the
// entries selected by show, followed by tail.
func (i *ChecklistItem) finish(glyph string, tail []string, show func(checklistEntry) bool) {
	i.done = true
	c := i.c
	if c.open == i {
		c.open = nil
	}
	_, _ = io.WriteString(c.out, progressClearPrefix+checklistItemIndent+glyph+" "+i.label+"\n")
	for _, e := range i.entries {
		if show(e) {
			c.writeln(checklistDetailIndent + e.text)
		} else {
			c.folded = append(c.folded, e.text)
		}
	}
	for _, t := range tail {
		c.writeln(checklistDetailIndent + t)
	}
}

// drawRunning draws the in-flight line without a trailing newline. A redraw
// clears the line first; the first draw follows a newline and needs no clear.
func (i *ChecklistItem) drawRunning(redraw bool) {
	prefix := ""
	if redraw {
		prefix = progressClearPrefix
	}
	_, _ = io.WriteString(i.c.out, prefix+checklistItemIndent+Spinner(i.label, &i.c.theme))
}

// assertNotDone panics on a second terminal transition, matching
// ProgressLineHandle.
func (i *ChecklistItem) assertNotDone() {
	if i.done {
		panic("tui: ChecklistItem terminal state already reached (Done/Fail/Skip already called)")
	}
}

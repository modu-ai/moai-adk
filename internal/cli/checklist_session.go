package cli

// checklist_session.go — the CLI side of the group-checklist progress surface
// (SPEC-TUI-CHECKLIST-001, design.md §3). The renderer lives in
// internal/tui/checklist.go and knows nothing about any command; this file
// decides WHETHER the checklist is drawn, captures the command's own output
// while an item is open, routes warnings to the open item, and closes the
// surface when the run ends.
//
// An inactive session (any activation condition false) is an empty no-op value:
// every method returns at once, the command writers are left alone, and the
// command runs on its existing output path byte for byte (REQ-TCL-012).

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/mattn/go-isatty"
	"github.com/spf13/cobra"

	"github.com/modu-ai/moai-adk/internal/cli/printer"
	"github.com/modu-ai/moai-adk/internal/cli/wizard"
	"github.com/modu-ai/moai-adk/internal/profile"
	"github.com/modu-ai/moai-adk/internal/tui"
	"github.com/modu-ai/moai-adk/pkg/version"
)

// Test seams (package variables, design.md §3.1). A test that changes one runs
// without t.Parallel and restores it with t.Cleanup.
var (
	// checklistIsTTYFn reports whether one writer is a terminal. It is asked
	// once per process stream, so the stdout answer never stands in for the
	// stderr one.
	checklistIsTTYFn = func(w io.Writer) bool {
		f, ok := w.(*os.File)
		return ok && isatty.IsTerminal(f.Fd())
	}
	// checklistProcessStreamsFn returns the process stdout and stderr, the
	// comparison target of activation condition 2.
	checklistProcessStreamsFn = func() (stdout, stderr io.Writer) { return os.Stdout, os.Stderr }
	// checklistForceOnFn makes the whole predicate true; not even --verbose can
	// turn the checklist off once it returns true.
	checklistForceOnFn = func() bool { return false }
	// checklistSessionStartFn is a test-only injection point. Production value:
	// nil. When set it is called once per run right after the warning collector
	// is created — on the active and the inactive path alike, before the first
	// item.
	checklistSessionStartFn func(cmd *cobra.Command, c *warnCollector)
	// checklistClockFn supplies the summary's elapsed time. Production value:
	// nil (time.Now).
	checklistClockFn func() time.Time
)

// checklistInputs are the five activation inputs (condition 1 splits into the
// two process streams, so the struct has six fields).
type checklistInputs struct {
	StdoutTTY      bool // condition 1: process stdout is a terminal
	StderrTTY      bool // condition 1: process stderr is a terminal
	ProcessStreams bool // condition 2: both command writers are the process streams
	CIUnset        bool // condition 3: CI is empty
	TermNotDumb    bool // condition 4: TERM is not "dumb"
	VerboseOff     bool // condition 5: --verbose is off
}

// checklistActive is the activation predicate: every condition must hold.
func checklistActive(in checklistInputs) bool {
	return in.StdoutTTY && in.StderrTTY && in.ProcessStreams && in.CIUnset && in.TermNotDumb && in.VerboseOff
}

// checklistEnv collects the activation inputs. Commands call only this; they
// never read the inputs themselves.
func checklistEnv(cmd *cobra.Command, verbose bool) checklistInputs {
	procOut, procErr := checklistProcessStreamsFn()
	return checklistInputs{
		StdoutTTY:      checklistIsTTYFn(procOut),
		StderrTTY:      checklistIsTTYFn(procErr),
		ProcessStreams: sameWriter(cmd.OutOrStdout(), procOut) && sameWriter(cmd.ErrOrStderr(), procErr),
		CIUnset:        os.Getenv("CI") == "",
		TermNotDumb:    os.Getenv("TERM") != "dumb",
		VerboseOff:     !verbose,
	}
}

// sameWriter compares two writers by identity. A writer whose dynamic type is
// not comparable is never "the same".
func sameWriter(a, b io.Writer) (same bool) {
	defer func() {
		if recover() != nil {
			same = false
		}
	}()
	return a == b
}

func init() {
	// init gains --verbose with this card: it forces the fallback output path
	// (REQ-TCL-013). update registers its own copy (update.go).
	initCmd.Flags().Bool("verbose", false, "Print plain progress output instead of the group checklist")
}

// outsideWarning is a warning recorded while no item was open. shown says its
// sentence already reached the real writer as an immediate line.
type outsideWarning struct {
	msg   string
	shown bool
}

// checklistSession is one run's checklist state. Methods are safe on a nil or
// inactive session.
type checklistSession struct {
	mu     sync.Mutex
	active bool

	cmd  *cobra.Command
	args []string

	origOut io.Writer
	origErr io.Writer
	outW    *sessionWriter
	errW    *sessionWriter

	lang          string
	cl            *tui.Checklist
	curKey        string // "group.item" of the open item, "" when none
	outside       []outsideWarning
	summaryDrawn  bool
	ended         bool
	restoreWizard func()
}

// beginChecklistSession decides activation and, when active, puts the session
// writers on the command. It never returns nil.
func beginChecklistSession(cmd *cobra.Command, args []string) *checklistSession {
	s := &checklistSession{cmd: cmd, args: args}
	if !checklistForceOnFn() && !checklistActive(checklistEnv(cmd, getBoolFlag(cmd, "verbose"))) {
		return s
	}
	s.active = true
	s.origOut, s.origErr = cmd.OutOrStdout(), cmd.ErrOrStderr()
	s.outW = &sessionWriter{s: s, real: s.origOut}
	s.errW = &sessionWriter{s: s, real: s.origErr}
	cmd.SetOut(s.outW)
	cmd.SetErr(s.errW)
	s.lang = checklistProfileLang()
	s.followWizardLang()
	return s
}

// isActive reports whether the checklist is on for this run.
func (s *checklistSession) isActive() bool { return s != nil && s.active }

// collector wraps pr in the warning collector, binds the session to it and
// calls the test injection point. The inactive session is bound too, so the
// collector behaves exactly as before.
func (s *checklistSession) collector(pr printer.Printer) *warnCollector {
	c := newWarnCollector(pr)
	c.sess = s
	if checklistSessionStartFn != nil {
		checklistSessionStartFn(s.cmd, c)
	}
	return c
}

// checklistProfileLang is the catalogue language before any wizard answer: the
// profile locale, else English.
func checklistProfileLang() string {
	prefs, err := profile.ReadPreferences(profile.GetCurrentName())
	if err != nil || prefs.ConversationLang == "" {
		return "en"
	}
	return prefs.ConversationLang
}

// followWizardLang records the conversation language the wizard returns, so the
// checklist — created only at the first item, after the wizard — speaks it. The
// wizard itself is untouched: the seam is wrapped, not changed.
func (s *checklistSession) followWizardLang() {
	prev := runWizardFn
	runWizardFn = func(root, convLang, userName string) (*wizard.WizardResult, error) {
		res, err := prev(root, convLang, userName)
		if err == nil && res != nil && res.ConversationLang != "" {
			s.mu.Lock()
			if s.cl == nil {
				s.lang = res.ConversationLang
			}
			s.mu.Unlock()
		}
		return res, err
	}
	s.restoreWizard = func() { runWizardFn = prev }
}

// checklistLocked returns the renderer, creating it (and recording the header)
// on first use. The header is drawn by the renderer when the first item opens.
func (s *checklistSession) checklistLocked() *tui.Checklist {
	if s.cl == nil {
		th := resolveTheme()
		s.cl = tui.NewChecklist(tui.ChecklistOptions{
			Out:   s.origErr,
			Theme: &th,
			Lang:  s.lang,
			Clock: checklistClockFn,
		})
		s.cl.Header("checklist.header.init", tui.EnvSummary{
			Platform: runtime.GOOS + "/" + runtime.GOARCH,
			Version:  version.GetVersion(),
			Dir:      checklistProjectDir(s.cmd, s.args),
		})
	}
	return s.cl
}

// openLocked returns the open item, or nil.
func (s *checklistSession) openLocked() *tui.ChecklistItem {
	if s.cl == nil {
		return nil
	}
	return s.cl.Open()
}

// checklistProjectDir is the project directory shown in the environment line,
// resolved the way runInit resolves it and shortened with ~ under the home.
func checklistProjectDir(cmd *cobra.Command, args []string) string {
	root := getStringFlag(cmd, "root")
	switch {
	case root != "":
	case len(args) > 0 && args[0] != ".":
		if abs, err := filepath.Abs(args[0]); err == nil {
			root = abs
		} else {
			root = args[0]
		}
	default:
		root, _ = os.Getwd()
	}
	if home, err := os.UserHomeDir(); err == nil && home != "" {
		if rest, ok := strings.CutPrefix(root, home+string(filepath.Separator)); ok {
			return "~" + string(filepath.Separator) + rest
		}
	}
	return root
}

// openItem opens group/item, closing a still-open item as done first.
func (s *checklistSession) openItem(group, item string) {
	if !s.isActive() {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.openItemLocked(group, item)
}

func (s *checklistSession) openItemLocked(group, item string) {
	if s.ended {
		return
	}
	s.flushPendingLocked()
	s.checklistLocked().Item(group, item)
	s.curKey = group + "." + item
}

// openItemOnce opens group/item unless it is already the open item (the
// executor reports "Validation" twice in an explicit-mode run; one item is one
// line).
func (s *checklistSession) openItemOnce(group, item string) {
	if !s.isActive() {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.openLocked() != nil && s.curKey == group+"."+item {
		return
	}
	s.openItemLocked(group, item)
}

// skipItem opens group/item and closes it as skipped at once.
func (s *checklistSession) skipItem(group, item string) {
	if !s.isActive() {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.ended {
		return
	}
	s.openItemLocked(group, item)
	if it := s.openLocked(); it != nil {
		it.Skip("checklist.status.skipped")
		s.curKey = ""
	}
}

// closeOpenItem closes the open item as done.
func (s *checklistSession) closeOpenItem() {
	if !s.isActive() {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.flushPendingLocked()
	if it := s.openLocked(); it != nil {
		it.Done()
		s.curKey = ""
	}
}

// failOpenItem closes the open item as failed and expands everything recorded
// against it.
func (s *checklistSession) failOpenItem(err error) {
	if !s.isActive() {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.flushPendingLocked()
	if it := s.openLocked(); it != nil {
		msg := ""
		if err != nil {
			msg = err.Error()
		}
		it.Fail(msg)
		s.curKey = ""
	}
}

// detail records a collapsed detail line against the open item.
func (s *checklistSession) detail(line string, advance bool) {
	if !s.isActive() {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if it := s.openLocked(); it != nil {
		if advance {
			it.Update()
		}
		if line != "" {
			it.Detail(line)
		}
	}
}

// flushPendingLocked records the partial (unterminated) lines both session
// writers hold against the open item, so a later item change cannot strand
// them.
func (s *checklistSession) flushPendingLocked() {
	for _, w := range []*sessionWriter{s.outW, s.errW} {
		if w != nil {
			w.flushPartialLocked()
		}
	}
}

// captureWarning records msg as a warning of the open item and reports whether
// an item was open. The warning's own sentence is shown once, under the item.
func (s *checklistSession) captureWarning(msg string) bool {
	if !s.isActive() {
		return false
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	it := s.openLocked()
	if it == nil {
		return false
	}
	it.Warn(severityWarnLine(msg))
	return true
}

// severityWarnLine is the line shape of a warning entry: the severity glyph and
// the sentence.
func severityWarnLine(msg string) string {
	return severityGlyph(sevWarn, resolveTheme()) + " " + msg
}

// noteOutside records a warning raised while no item was open. shown is true
// when the sentence already reached the real writer as an immediate line.
func (s *checklistSession) noteOutside(msg string, shown bool) {
	if !s.isActive() {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.outside = append(s.outside, outsideWarning{msg: msg, shown: shown})
}

// summary draws the summary block. It lists the warnings recorded outside any
// item, so the end of the run does not write them again.
func (s *checklistSession) summary(sum tui.ChecklistSummary) {
	if !s.isActive() {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.ended {
		return
	}
	s.flushPendingLocked()
	for _, o := range s.outside {
		sum.OutsideWarnings = append(sum.OutsideWarnings, o.msg)
	}
	s.checklistLocked().Summary(sum)
	s.curKey = ""
	s.summaryDrawn = true
}

// end finishes the run: it closes the open item (failed when the run errored,
// done otherwise), writes the collector warnings that never reached the real
// writer, and hands the original command writers back. It is the first defer
// of runInit, so it runs last.
func (s *checklistSession) end(err error) {
	if !s.isActive() {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.ended {
		return
	}
	s.flushPendingLocked()
	if it := s.openLocked(); it != nil {
		if err != nil {
			it.Fail(err.Error())
		} else {
			it.Done()
		}
		s.curKey = ""
	}
	if !s.summaryDrawn {
		th := resolveTheme()
		for _, o := range s.outside {
			if !o.shown {
				emitSeverityLine(s.origErr, sevWarn, th, "%s", o.msg)
			}
		}
	}
	s.ended = true
	s.cmd.SetOut(s.origOut)
	s.cmd.SetErr(s.origErr)
	if s.restoreWizard != nil {
		s.restoreWizard()
	}
}

// preservedKind selects how a preserved line is recorded against the open item.
type preservedKind int

const (
	preservedNone preservedKind = iota
	// preservedNotice is an informational line that stays visible under a
	// successful item and leaves the item glyph alone.
	preservedNotice
	// preservedWarn is a warning or error line: visible under the item, and the
	// item ends with the warning glyph.
	preservedWarn
)

// beginPreserved marks the lines w receives until the returned function is
// called as preserved lines (REQ-TCL-022). On a writer that is not a session
// writer it does nothing, so the call sites cost the fallback path nothing.
func beginPreserved(w io.Writer, kind preservedKind) (end func()) {
	sw, ok := w.(*sessionWriter)
	if !ok {
		return func() {}
	}
	sw.s.mu.Lock()
	prev := sw.keep
	sw.keep = kind
	sw.s.mu.Unlock()
	return func() {
		sw.s.mu.Lock()
		defer sw.s.mu.Unlock()
		sw.flushPartialLocked()
		sw.keep = prev
	}
}

// warningReceiver is implemented by the session writers: emitSeverityLine hands
// a warn/err line to the open item through it instead of writing it.
type warningReceiver interface {
	ReceiveWarning(line string) bool
}

// sessionWriter replaces a command writer while the session is active. Bytes
// written while no item is open pass straight through; bytes written while an
// item is open are recorded against it line by line.
type sessionWriter struct {
	s       *checklistSession
	real    io.Writer
	pending []byte
	keep    preservedKind
}

func (w *sessionWriter) Write(p []byte) (int, error) {
	w.s.mu.Lock()
	defer w.s.mu.Unlock()
	it := w.s.openLocked()
	if it == nil {
		return w.real.Write(p)
	}
	w.pending = append(w.pending, p...)
	for {
		i := indexByte(w.pending, '\n')
		if i < 0 {
			break
		}
		line := strings.TrimSuffix(string(w.pending[:i]), "\r")
		w.pending = w.pending[i+1:]
		w.recordLocked(it, line)
	}
	return len(p), nil
}

func indexByte(b []byte, c byte) int {
	for i, x := range b {
		if x == c {
			return i
		}
	}
	return -1
}

// recordLocked files one complete line against it by the writer's current
// preserved state.
func (w *sessionWriter) recordLocked(it *tui.ChecklistItem, line string) {
	switch w.keep {
	case preservedNotice:
		it.Keep(line)
	case preservedWarn:
		it.Warn(line)
	case preservedNone:
		it.Detail(line)
	}
}

// flushPartialLocked records an unterminated trailing line against the open
// item; with no open item the bytes go to the real writer unchanged.
func (w *sessionWriter) flushPartialLocked() {
	if len(w.pending) == 0 {
		return
	}
	line := string(w.pending)
	w.pending = nil
	if it := w.s.openLocked(); it != nil {
		w.recordLocked(it, line)
		return
	}
	_, _ = io.WriteString(w.real, line)
}

// ReceiveWarning implements warningReceiver.
func (w *sessionWriter) ReceiveWarning(line string) bool {
	w.s.mu.Lock()
	defer w.s.mu.Unlock()
	it := w.s.openLocked()
	if it == nil {
		return false
	}
	w.flushPartialLocked()
	it.Warn(line)
	return true
}

var _ warningReceiver = (*sessionWriter)(nil)

// checklistReporter adapts the executor's step events to checklist items
// (design.md §3.6). Events carry no item key of their own: the step name picks
// the templates item, and a repeated name keeps its single line.
type checklistReporter struct {
	s *checklistSession
}

func newChecklistReporter(s *checklistSession) *checklistReporter {
	return &checklistReporter{s: s}
}

func (r *checklistReporter) StepStart(name, message string) {
	switch name {
	case "Detection":
		r.s.openItemOnce(clGroupTemplates, clItemTemplatesDetect)
	case "Methodology Detection", "Validation":
		r.s.openItemOnce(clGroupTemplates, clItemTemplatesValidate)
	case "Initialization":
		r.s.openItemOnce(clGroupTemplates, clItemTemplatesDeploy)
	default:
		r.s.detail(fmt.Sprintf("%s: %s", name, message), false)
	}
}

func (r *checklistReporter) StepUpdate(message string) { r.s.detail(message, true) }

func (r *checklistReporter) StepComplete(message string) { r.s.detail(message, false) }

func (r *checklistReporter) StepError(err error) { r.s.failOpenItem(err) }

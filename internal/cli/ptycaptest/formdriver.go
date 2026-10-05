package ptycaptest

// The form driver: huh v2 forms driven programmatically with bubbletea v2
// messages (the M2a spike technique) — no TTY, no form.Run, cross-platform.
// The wizard package's unified-form tests and the cli package's profile-wizard
// golden tests share this one implementation.

import (
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/huh/v2"
)

// FormDriver drives a huh v2 form via Update messages, executing returned
// commands with a bounded timeout so sleeping tick commands (cursor blink)
// are abandoned instead of recursing.
type FormDriver struct {
	t *testing.T
	m huh.Model
}

// NewFormDriver initializes f and sizes the viewport to 80x40.
func NewFormDriver(t *testing.T, f *huh.Form) *FormDriver {
	t.Helper()
	d := &FormDriver{t: t, m: f}
	d.Drain(f.Init())
	d.Send(tea.WindowSizeMsg{Width: 80, Height: 40})
	return d
}

// Send delivers one message to the form and drains the returned command.
func (d *FormDriver) Send(msg tea.Msg) {
	nm, cmd := d.m.Update(msg)
	d.m = nm
	d.Drain(cmd)
}

// Drain executes a tea.Cmd with a bounded timeout: huh's internal routing
// messages return instantly; sleeping tick commands are abandoned.
func (d *FormDriver) Drain(cmd tea.Cmd) {
	if cmd == nil {
		return
	}
	msg := execBoundedCmd(cmd)
	if msg == nil {
		return
	}
	if batch, ok := msg.(tea.BatchMsg); ok {
		for _, c := range batch {
			d.Drain(c)
		}
		return
	}
	d.Send(msg)
}

// execBoundedCmd runs a tea.Cmd with a timeout: huh's internal routing
// messages return instantly; sleeping tick commands are abandoned.
func execBoundedCmd(cmd tea.Cmd) tea.Msg {
	ch := make(chan tea.Msg, 1)
	go func() { ch <- cmd() }()
	select {
	case msg := <-ch:
		return msg
	case <-time.After(50 * time.Millisecond):
		return nil
	}
}

// TypeText types a string rune by rune.
func (d *FormDriver) TypeText(s string) {
	for _, r := range s {
		d.Send(tea.KeyPressMsg{Code: r, Text: string(r)})
	}
}

// Enter presses enter.
func (d *FormDriver) Enter() { d.Send(tea.KeyPressMsg{Code: tea.KeyEnter}) }

// Down presses the down arrow.
func (d *FormDriver) Down() { d.Send(tea.KeyPressMsg{Code: tea.KeyDown}) }

// Backspace presses the delete key.
func (d *FormDriver) Backspace() {
	d.Send(tea.KeyPressMsg{Code: tea.KeyBackspace})
}

// View renders the current frame.
func (d *FormDriver) View() string { return d.m.(*huh.Form).View() }

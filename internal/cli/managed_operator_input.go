package cli

// managed_operator_input.go — SPEC-FACTORY-MANAGED-CARD-CHILD-001: the lane
// loop's managed card-child launch and the operator-input pump the loop's
// successive managed sessions share.
//
// Why a pump: the managed owner reads operator lines on a goroutine that a
// finished session never releases, and its scanner and buffered channel
// consume lines ahead of the driver. Handing every session the process stdin
// would let an ended session swallow the next card's input. The pump reads the
// source itself, hands each session an adapter that returns ONE line per Read
// and shuts itself after the line that ends the session (/exit or /quit), and
// keeps every later line for the next session. The owner and its driver are
// not touched; the plain `moai codex` divert keeps os.Stdin.

import (
	"bufio"
	"io"
	"os"
	"strings"
	"sync"
	"sync/atomic"
)

// managedLaneOperatorSource is the operator-input source the lane loop's pump
// reads (default: the process stdin). Tests point it at their own reader.
var managedLaneOperatorSource io.Reader = os.Stdin

// managedOperatorPumpsCreated counts pumps built, so a test can show the
// switch-off path builds none.
var managedOperatorPumpsCreated atomic.Int64

// managedLanePump holds the pump of the running lane loop. The loop clears it
// when it returns (endManagedLanePump), so a later loop call reads a fresh
// source; a pump is built only when the first managed session asks for input.
var managedLanePump struct {
	mu sync.Mutex
	p  *managedOperatorPump
}

func acquireManagedLanePump() *managedOperatorPump {
	managedLanePump.mu.Lock()
	defer managedLanePump.mu.Unlock()
	if managedLanePump.p == nil {
		managedLanePump.p = newManagedOperatorPump(managedLaneOperatorSource)
	}
	return managedLanePump.p
}

func endManagedLanePump() {
	managedLanePump.mu.Lock()
	p := managedLanePump.p
	managedLanePump.p = nil
	managedLanePump.mu.Unlock()
	if p != nil {
		p.stop()
	}
}

// defaultManagedCodexCardLaunch is the default body of the lane loop's managed
// launch seam: the plain managed owner, fed through an adapter of the loop's
// operator-input pump that is closed when the session returns.
func defaultManagedCodexCardLaunch(bin string, args, env []string, dir string) error {
	in := acquireManagedLanePump().attach()
	defer func() { _ = in.Close() }()
	return runManagedFactoryCodex(bin, args, env, dir, in)
}

// managedEndToken reports whether line (with or without its line ending) is a
// token that ends a managed session: the same two values, compared the same
// way, as the delivery driver's own check. The loop-level test runs both real
// tokens through the real driver, so a drift between the two reads red.
func managedEndToken(line string) bool {
	line = strings.TrimSuffix(strings.TrimSuffix(line, "\n"), "\r")
	return line == "/exit" || line == "/quit"
}

// managedOperatorPump is the lane loop's one reader of the operator input. It
// holds at most one line read ahead of the sessions.
// @MX:NOTE: operator lines belong to the session that RECEIVES them; the pump
// only guarantees that lines after a session's end token reach the next one.
// @MX:SPEC: SPEC-FACTORY-MANAGED-CARD-CHILD-001
type managedOperatorPump struct {
	src io.Reader

	mu         sync.Mutex
	cond       *sync.Cond
	started    bool
	stopped    bool
	hasPending bool
	pending    string
	srcDone    bool
}

// newManagedOperatorPump builds a pump over src; it reads nothing until the
// first adapter is attached.
func newManagedOperatorPump(src io.Reader) *managedOperatorPump {
	managedOperatorPumpsCreated.Add(1)
	p := &managedOperatorPump{src: src}
	p.cond = sync.NewCond(&p.mu)
	return p
}

// stop releases a reader goroutine waiting to hand over a line and refuses
// further input. A goroutine blocked inside the source's Read is not
// interruptible; it ends with the process (or the source).
func (p *managedOperatorPump) stop() {
	p.mu.Lock()
	p.stopped = true
	p.cond.Broadcast()
	p.mu.Unlock()
}

// attach returns the reader one managed session reads its operator input from
// and starts the source reader on first use.
func (p *managedOperatorPump) attach() io.ReadCloser {
	p.mu.Lock()
	if !p.started {
		p.started = true
		go p.run()
	}
	p.mu.Unlock()
	return &managedOperatorAdapter{p: p}
}

// run reads the source line by line, keeping at most one line pending.
// @MX:WARN: a goroutine over a blocking source read cannot be cancelled.
// @MX:REASON: os.Stdin has no read deadline; the goroutine ends with the source or the process, and holds one line at most.
func (p *managedOperatorPump) run() {
	reader := bufio.NewReader(p.src)
	for {
		p.mu.Lock()
		for p.hasPending && !p.stopped {
			p.cond.Wait()
		}
		stopped := p.stopped
		p.mu.Unlock()
		if stopped {
			return
		}
		line, err := reader.ReadString('\n')
		p.mu.Lock()
		if line != "" && !p.stopped {
			p.pending, p.hasPending = line, true
		}
		if err != nil {
			p.srcDone = true
		}
		p.cond.Broadcast()
		p.mu.Unlock()
		if err != nil {
			return
		}
	}
}

// managedOperatorAdapter is one session's view of the pump: one line per Read,
// end of input after the session's end token or after Close.
type managedOperatorAdapter struct {
	p *managedOperatorPump
	// guarded by p.mu
	rest        []byte
	endsSession bool
	closed      bool
	latched     bool
}

func (a *managedOperatorAdapter) Read(b []byte) (int, error) {
	if len(b) == 0 {
		return 0, nil
	}
	p := a.p
	p.mu.Lock()
	defer p.mu.Unlock()
	for {
		if a.closed || a.latched {
			return 0, io.EOF
		}
		if len(a.rest) > 0 {
			n := copy(b, a.rest)
			a.rest = a.rest[n:]
			if len(a.rest) == 0 && a.endsSession {
				a.latched = true
			}
			return n, nil
		}
		if p.hasPending {
			line := p.pending
			p.pending, p.hasPending = "", false
			p.cond.Broadcast()
			a.rest = []byte(line)
			a.endsSession = managedEndToken(line)
			continue
		}
		if p.srcDone || p.stopped {
			return 0, io.EOF
		}
		p.cond.Wait()
	}
}

// Close releases a Read blocked on the pump with end of input.
func (a *managedOperatorAdapter) Close() error {
	a.p.mu.Lock()
	a.closed = true
	a.p.cond.Broadcast()
	a.p.mu.Unlock()
	return nil
}

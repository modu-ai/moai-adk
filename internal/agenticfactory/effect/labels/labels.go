// Package labels applies a labels manifest to a code host in six stages:
// check -> dry-run (plan) -> diff -> operator approval -> idempotent apply -> readback verification.
//
// Safety shape: writes are off by default (AllowWrite=false makes Apply refuse before any call);
// apply needs an approval bound to the digest of the exact plan the operator saw, and re-plans just
// before writing so a remote that changed in between refuses the run; there is no delete operation
// at all (labels the manifest does not mention are only reported); every write is an effect intent
// (idempotent key, lost responses reconciled by looking the label up); a rate limit or a permission
// denial stops the run and is never retried; success is claimed only after the remote is read back.
// The host is an interface so GitHub and GitLab adapters sit behind the same boundary; only the
// GitHub adapter exists here, and nothing in this package talks to a network by itself.
package labels

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"sort"
	"strings"
	"time"
	"unicode"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/core"
	"github.com/modu-ai/moai-adk/internal/agenticfactory/effect"
	"github.com/modu-ai/moai-adk/internal/agenticfactory/forge"
)

var (
	ErrWriteDisabled = errors.New("labels: writes are disabled (AllowWrite=false)")
	ErrNotApproved   = errors.New("labels: apply needs an operator approval bound to the plan digest")
	ErrPlanChanged   = errors.New("labels: the remote no longer matches the approved plan; re-run the dry-run")
)

// Label is one label. Color is six lower-case hex digits without '#'.
type Label struct {
	Name        string
	Color       string
	Description string
}

// Manifest is the desired set.
type Manifest struct{ Labels []Label }

// Host is the neutral label API of one repository.
type Host interface {
	List(ctx context.Context) ([]Label, error)
	Create(ctx context.Context, l Label) error
	Update(ctx context.Context, name string, l Label) error
}

var colorRe = regexp.MustCompile(`^[0-9a-f]{6}$`)

func hasControl(s string) bool {
	for _, r := range s {
		if unicode.IsControl(r) {
			return true
		}
	}
	return false
}

// Check validates and normalises a manifest (stage 1). It needs no network.
func Check(m Manifest) (Manifest, error) {
	if len(m.Labels) == 0 {
		return Manifest{}, errors.New("labels: manifest has no labels")
	}
	seen := map[string]bool{}
	out := Manifest{}
	for i, l := range m.Labels {
		l.Name = strings.TrimSpace(l.Name)
		l.Color = strings.ToLower(strings.TrimPrefix(strings.TrimSpace(l.Color), "#"))
		switch {
		case l.Name == "":
			return Manifest{}, fmt.Errorf("labels: label %d has an empty name", i)
		case hasControl(l.Name) || hasControl(l.Description):
			return Manifest{}, fmt.Errorf("labels: label %q has control characters", l.Name)
		case !colorRe.MatchString(l.Color):
			return Manifest{}, fmt.Errorf("labels: label %q color must be six hex digits", l.Name)
		case seen[strings.ToLower(l.Name)]:
			return Manifest{}, fmt.Errorf("labels: duplicate label %q (names compare case-insensitively)", l.Name)
		}
		seen[strings.ToLower(l.Name)] = true
		out.Labels = append(out.Labels, l)
	}
	return out, nil
}

// OpKind is what an operation does. There is deliberately no delete.
type OpKind string

const (
	OpCreate OpKind = "create"
	OpUpdate OpKind = "update"
)

// Op is one planned change.
type Op struct {
	Kind OpKind
	Name string
	From *Label
	To   Label
}

// Plan is the dry-run result.
type Plan struct {
	Repo      string
	Ops       []Op
	Unmanaged []string // remote labels the manifest does not mention; reported, never touched
}

// Diff renders the plan as stable text for the operator to read.
func (p Plan) Diff() string {
	var b strings.Builder
	for _, o := range p.Ops {
		switch o.Kind {
		case OpCreate:
			fmt.Fprintf(&b, "+ create %s color=%s description=%q\n", o.Name, o.To.Color, o.To.Description)
		case OpUpdate:
			fmt.Fprintf(&b, "~ update %s: color %s -> %s; description %q -> %q\n", o.Name, o.From.Color, o.To.Color, o.From.Description, o.To.Description)
		}
	}
	for _, u := range p.Unmanaged {
		fmt.Fprintf(&b, "? unmanaged %s (left alone)\n", u)
	}
	return b.String()
}

// Digest identifies the exact plan; an approval is only valid for this digest.
func (p Plan) Digest() string {
	h := sha256.Sum256([]byte(p.Repo + "\n" + p.Diff()))
	return hex.EncodeToString(h[:])
}

// Approval is the operator's consent for one plan.
type Approval struct {
	PlanDigest string
	By         string
}

// Result is what Apply reports.
type Result struct {
	Verified      bool
	Mismatches    []string
	Stopped       bool // a rate limit, denial or failure ended the run early
	ReadbackLists int
}

// Applier runs the stages against a Host.
type Applier struct {
	Host       Host
	AllowWrite bool
	Repo       string
	Store      effect.IntentStore
	Timeouts   effect.Timeouts
	Backoff    forge.Backoff         // optional; with it a rate limit sets NotBefore
	Now        func() time.Time      // optional
	Pace       func(context.Context) // optional; called between writes (documented >=1 s gap)
}

func same(a, b Label) bool { return a.Color == b.Color && a.Description == b.Description }

// Plan runs stages 1-3: check, dry-run against the remote, diff. It only reads.
func (a *Applier) Plan(ctx context.Context, m Manifest) (Plan, error) {
	m, err := Check(m)
	if err != nil {
		return Plan{}, err
	}
	remote, err := a.Host.List(ctx)
	if err != nil {
		return Plan{}, fmt.Errorf("labels: reading the remote: %w", err)
	}
	have := map[string]Label{}
	for _, l := range remote {
		have[strings.ToLower(l.Name)] = l
	}
	p := Plan{Repo: a.Repo}
	want := map[string]bool{}
	for _, l := range m.Labels {
		key := strings.ToLower(l.Name)
		want[key] = true
		r, ok := have[key]
		switch {
		case !ok:
			p.Ops = append(p.Ops, Op{Kind: OpCreate, Name: l.Name, To: l})
		case !same(r, l):
			rr := r
			p.Ops = append(p.Ops, Op{Kind: OpUpdate, Name: l.Name, From: &rr, To: l})
		}
	}
	for k, l := range have {
		if !want[k] {
			p.Unmanaged = append(p.Unmanaged, l.Name)
		}
	}
	sort.Slice(p.Ops, func(i, j int) bool { return p.Ops[i].Name < p.Ops[j].Name })
	sort.Strings(p.Unmanaged)
	return p, nil
}

func (a *Applier) now() time.Time {
	if a.Now != nil {
		return a.Now()
	}
	return time.Now()
}

type opPayload struct {
	Kind OpKind `json:"kind"`
	Name string `json:"name"`
	To   struct {
		Name, Color, Description string
	} `json:"to"`
}

// Apply runs stages 4-6: approval check, idempotent writes, readback. It refuses before any call
// when writes are disabled or the approval is missing, and before any write when the remote moved.
func (a *Applier) Apply(ctx context.Context, m Manifest, ap Approval) (Result, error) {
	if !a.AllowWrite {
		return Result{}, ErrWriteDisabled
	}
	if strings.TrimSpace(ap.By) == "" || strings.TrimSpace(ap.PlanDigest) == "" {
		return Result{}, ErrNotApproved
	}
	m, err := Check(m)
	if err != nil {
		return Result{}, err
	}
	plan, err := a.Plan(ctx, m)
	if err != nil {
		return Result{}, err
	}
	if plan.Digest() != ap.PlanDigest {
		return Result{}, ErrPlanChanged
	}

	ex := &effect.Executor{Store: a.Store, Send: opSender{a}, Lookup: opLookup{a}, Timeouts: a.Timeouts, Now: a.Now}
	var res Result
	for i, op := range plan.Ops {
		current := op.Name // the name the host knows today (an update renames only the casing)
		if op.From != nil {
			current = op.From.Name
		}
		raw, _ := json.Marshal(opPayload{Kind: op.Kind, Name: current, To: struct{ Name, Color, Description string }{op.To.Name, op.To.Color, op.To.Description}})
		req := effect.Request{Kind: effect.LabelApply, Target: a.Repo + "/" + strings.ToLower(op.Name), Digest: digestOf(op), Payload: string(raw)}
		// An op still in the plan means the remote does not match, so an old "confirmed" record for the
		// same op is stale: arm it again rather than trusting it.
		if old, ok, _ := a.Store.Get(effect.Key(req.Kind, req.Target, req.Digest)); ok && old.Status == effect.StatusConfirmed {
			old.Status, old.Attempts, old.ProviderID = effect.StatusIntended, 0, ""
			// A conflict means another executor moved the record; Execute follows what is stored.
			if _, err := a.Store.PutIf(old); err != nil && !errors.Is(err, core.ErrRevisionConflict) {
				return res, err
			}
		}
		if i > 0 && a.Pace != nil {
			a.Pace(ctx)
		}
		got, err := ex.Execute(ctx, req)
		if err != nil {
			res.Stopped = true
			return res, err
		}
		if got.Status != effect.StatusConfirmed {
			res.Stopped = true
			return res, fmt.Errorf("labels: %s %s did not land (%s): %s", op.Kind, op.Name, got.Status, got.LastError)
		}
	}

	remote, err := a.Host.List(ctx)
	res.ReadbackLists++
	if err != nil {
		return res, fmt.Errorf("labels: readback failed: %w", err)
	}
	have := map[string]Label{}
	for _, l := range remote {
		have[strings.ToLower(l.Name)] = l
	}
	for _, l := range m.Labels {
		if r, ok := have[strings.ToLower(l.Name)]; !ok {
			res.Mismatches = append(res.Mismatches, l.Name+": missing on the remote")
		} else if !same(r, l) {
			res.Mismatches = append(res.Mismatches, l.Name+": remote differs after apply")
		}
	}
	res.Verified = len(res.Mismatches) == 0
	return res, nil
}

func digestOf(o Op) string {
	h := sha256.Sum256([]byte(fmt.Sprintf("%s\x00%s\x00%s\x00%s\x00%s", o.Kind, o.Name, o.To.Name, o.To.Color, o.To.Description)))
	return hex.EncodeToString(h[:])
}

type opSender struct{ a *Applier }

func (s opSender) Send(ctx context.Context, in effect.Intent) (string, error) {
	var p opPayload
	if err := json.Unmarshal([]byte(in.Payload), &p); err != nil {
		return "", fmt.Errorf("%w: bad payload", effect.ErrNotSent)
	}
	l := Label{Name: p.To.Name, Color: p.To.Color, Description: p.To.Description}
	var err error
	switch p.Kind {
	case OpCreate:
		err = s.a.Host.Create(ctx, l)
	case OpUpdate:
		err = s.a.Host.Update(ctx, p.Name, l)
	default:
		return "", fmt.Errorf("%w: unknown operation %q", effect.ErrNotSent, p.Kind)
	}
	if err != nil {
		return "", forge.ForEffect(err, s.a.now(), s.a.Backoff)
	}
	return strings.ToLower(l.Name), nil
}

type opLookup struct{ a *Applier }

func (l opLookup) Lookup(ctx context.Context, in effect.Intent) (effect.LookupResult, error) {
	var p opPayload
	if err := json.Unmarshal([]byte(in.Payload), &p); err != nil {
		return effect.LookupResult{}, err
	}
	remote, err := l.a.Host.List(ctx)
	if err != nil {
		return effect.LookupResult{}, err
	}
	want := Label{Name: p.To.Name, Color: p.To.Color, Description: p.To.Description}
	for _, r := range remote {
		if strings.EqualFold(r.Name, want.Name) && same(r, want) {
			return effect.LookupResult{Found: true, ProviderID: strings.ToLower(want.Name), Authoritative: true}, nil
		}
	}
	return effect.LookupResult{Authoritative: true}, nil
}

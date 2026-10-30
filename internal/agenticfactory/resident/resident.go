// Package resident wires the confirmed agentic-factory settings (internal/config) into the components of
// the resident monitor and runs its loop (Issue #1948). It carries no number of its own: every limit,
// period and name reaches a component from the settings value, and the two things whose source is still
// undecided (OD-17) - the loop's tick period and where a webhook secret is read from - are injected by
// the caller, never read from a key.
//
// It builds and runs read-side and queue-side machinery only. It opens no network connection, starts no
// process and enables no write path: ghintake's AllowWrite stays false, the gh runner it is given is the
// caller's read-only one, and the webhook handler is returned for the caller to mount.
package resident

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/claim"
	"github.com/modu-ai/moai-adk/internal/agenticfactory/core"
	"github.com/modu-ai/moai-adk/internal/agenticfactory/effect"
	"github.com/modu-ai/moai-adk/internal/agenticfactory/effect/ctxledger"
	"github.com/modu-ai/moai-adk/internal/agenticfactory/effect/docstore"
	"github.com/modu-ai/moai-adk/internal/agenticfactory/effect/ghidentity"
	"github.com/modu-ai/moai-adk/internal/agenticfactory/effect/ghintake"
	"github.com/modu-ai/moai-adk/internal/agenticfactory/effect/labels"
	"github.com/modu-ai/moai-adk/internal/agenticfactory/forge"
	"github.com/modu-ai/moai-adk/internal/agenticfactory/forge/webhook"
	"github.com/modu-ai/moai-adk/internal/agenticfactory/monitor/adapt"
	"github.com/modu-ai/moai-adk/internal/agenticfactory/monitor/notify"
	"github.com/modu-ai/moai-adk/internal/agenticfactory/monitor/notify/ghpr"
	"github.com/modu-ai/moai-adk/internal/agenticfactory/runner"
	"github.com/modu-ai/moai-adk/internal/config"
)

// The bound on each external effect. The leader decided the values for OD-20 (the key names, and so
// whether they become settings, stay open); until then they are these named constants, not literals at the
// call sites. An effect that does not answer within its bound is treated as possibly delivered: the
// executor records it as sent and reconciles it by looking, it never sends again.
const (
	// A gh subprocess that writes: the bound of the repository's existing gh subprocess
	// (internal/cli/todo_pr.go todoPRSubprocessTimeout).
	issueCreateTimeout   = 30 * time.Second
	ledgerCommentTimeout = 30 * time.Second
	labelApplyTimeout    = 30 * time.Second
	// A gh read: the bound of the delivery edge's gh call (internal/cli/factory_card_pr.go factoryGHTimeout).
	checkObserveTimeout = 10 * time.Second
	// Local git work. The repository has no precedent for these two; the leader set them, and an expiry is
	// safe because it takes the same reconcile-by-lookup path.
	branchPrepareTimeout   = 60 * time.Second
	worktreePrepareTimeout = 120 * time.Second
)

func defaultTimeouts() effect.Timeouts {
	return effect.Timeouts{
		IssueCreate: issueCreateTimeout, BranchPrepare: branchPrepareTimeout, WorktreePrepare: worktreePrepareTimeout,
		CheckObserve: checkObserveTimeout, LedgerComment: ledgerCommentTimeout, LabelApply: labelApplyTimeout,
	}
}

// Deps is what the settings cannot provide.
type Deps struct {
	// Store is the claim store, opened by the caller.
	Store *claim.Store
	// GHRead runs gh for reads only (ghpr and adapt refuse every other verb before calling it).
	GHRead ghpr.Runner
	// Now is the clock every component reads; no component reads the wall clock.
	Now func() time.Time
	// Secret returns the signing secret of one webhook route, read from wherever the operator keeps it
	// (OD-17). It is required when the settings name any route.
	Secret func(repo string) []byte
	// Transport answers the read-only host lookups behind the permission decision (the acting user's id
	// and an Issue's author). It is nil until a gh-backed transport is wired; while it is nil the Runner has
	// no Facts, and every authority check ends as an unknown identity - the fail-closed state, not an error.
	Transport forge.Transport
}

// Resident is the wired set.
type Resident struct {
	Settings config.AgenticFactorySettings

	// Runner is the one Runner this process builds. Its Facts are the host lookups of Deps.Transport; its
	// Attributor is left at the lock constant (REQ-AF-011): identity.Book is not connected.
	Runner  *runner.Runner
	Service *notify.Service
	Inbox   notify.Inbox
	Webhook *webhook.Handler // mount it; the route for a repository is /hook/<owner>/<name>

	Backoff forge.Backoff
	Poll    forge.Adaptive
	Budget  *forge.Budget
	Cursors docstore.Cursors

	deps     Deps
	queues   map[string]*docstore.Queue // by the repository a webhook route names
	timeouts effect.Timeouts
}

// ErrNotConfigured is returned when the settings say Agentic Factory is not configured.
var ErrNotConfigured = config.ErrAgenticFactoryNotConfigured

// BuildFromProject reads the settings under projectRoot and wires them. A project that has not
// configured Agentic Factory gets ErrNotConfigured and nothing is built.
func BuildFromProject(projectRoot string, d Deps) (*Resident, error) {
	s, err := config.LoadAgenticFactorySettings(projectRoot)
	if err != nil {
		return nil, err
	}
	return Build(s, d)
}

// Build wires validated settings. It repeats the checks whose failure would only show at the first use.
func Build(s config.AgenticFactorySettings, d Deps) (*Resident, error) {
	switch {
	case d.Store == nil:
		return nil, errors.New("resident: Deps.Store is required")
	case d.GHRead == nil:
		return nil, errors.New("resident: Deps.GHRead is required")
	case d.Now == nil:
		return nil, errors.New("resident: Deps.Now is required")
	case len(s.WebhookRoutes) > 0 && d.Secret == nil:
		return nil, errors.New("resident: Deps.Secret is required while webhook routes are configured (where the secret is read from is undecided, OD-17)")
	}
	r := &Resident{Settings: s, deps: d, queues: map[string]*docstore.Queue{}, timeouts: defaultTimeouts()}
	r.Backoff = forge.Backoff{
		MinSecondaryWait: s.GitHubBackoff.MinSecondaryWait, Base: s.GitHubBackoff.Base,
		Max: s.GitHubBackoff.Max, MaxAttempts: s.GitHubBackoff.MaxAttempts,
	}
	if err := r.Backoff.Validate(); err != nil {
		return nil, err
	}
	r.Poll = forge.Adaptive{Min: s.CollectorPoll.Min, Max: s.CollectorPoll.Max, Safety: s.CollectorPoll.Safety}
	r.Budget = &forge.Budget{ReserveCalls: s.BudgetReserve, ReportResource: s.BudgetReportResource, Now: d.Now}
	r.Cursors = docstore.Cursors{Docs: d.Store, Kind: docstore.CollectorKind, MaxDocBytes: s.CursorMaxDocBytes, MaxChunks: s.CursorMaxChunks}

	r.Inbox = notify.Inbox{Docs: d.Store, Kind: docstore.CollectorKind,
		MaxDocBytes: s.NotifyInbox.MaxDocBytes, SeenLimit: s.NotifyInbox.SeenLimit, MaxIDBytes: s.NotifyInbox.MaxIDBytes}

	h := &webhook.Handler{Provider: s.WebhookProvider, MaxBody: int64(s.WebhookMaxBody)}
	for _, repo := range s.WebhookRoutes {
		q := &docstore.Queue{Docs: d.Store, Kind: docstore.CollectorKind, Key: repo + "/hooks",
			MaxDocBytes: s.WebhookQueue.MaxDocBytes, SeenLimit: s.WebhookQueue.SeenLimit, MaxIDBytes: s.WebhookQueue.MaxIDBytes}
		if err := q.Validate(); err != nil {
			return nil, fmt.Errorf("resident: webhook queue of %s: %w", repo, err)
		}
		r.queues[core.NormalizeRepo(s.WebhookProvider, repo)] = q
		h.Routes = append(h.Routes, webhook.Route{Repo: repo, Queue: q, Secret: func() []byte { return d.Secret(repo) }})
	}
	r.Webhook = h

	r.Runner = runner.New(d.Store)
	if d.Transport != nil {
		r.Runner.Facts = r.hostFacts(d.Transport)
	}

	number := func(issue int) (int, bool, error) {
		ref, err := core.NewIssueRef(issue)
		if err != nil {
			return 0, false, err
		}
		rec, err := d.Store.Load(ref)
		if errors.Is(err, core.ErrNotHeld) {
			return 0, false, nil
		}
		if err != nil {
			return 0, false, err
		}
		if rec.Receipt == nil || rec.Receipt.PR <= 0 {
			return 0, false, nil
		}
		return rec.Receipt.PR, true, nil
	}
	r.Service = &notify.Service{
		Records: d.Store, Waits: r.Runner, Repairs: d.Store,
		PRs:      ghpr.Source{Repo: s.PRSourceRepository, Run: d.GHRead, Timeout: checkObserveTimeout, Number: number},
		Notifier: r.Inbox, Escalator: r.Inbox,
		State: r.Cursors, Now: d.Now, ReNotifyAfter: s.NotifyRenotifyAfter,
	}
	return r, nil
}

// hostFacts connects the GitHub identity reader to the Runner's permission decision: the acting user is the
// reader's CurrentUserID, an Issue's author and real repository are its IssueFacts (fields copied into
// runner.HostIssue), and who owns the work now is read from the claim store. Every lookup is bounded by the
// named gh-read timeout (checkObserveTimeout). Reads only; nothing here writes to the host.
func (r *Resident) hostFacts(t forge.Transport) runner.HostFacts {
	id := r.Identity(&forge.Client{T: t, B: r.Budget, Backoff: r.Backoff, Now: r.deps.Now})
	return runner.HostFacts{
		Repo:    r.Settings.PRSourceRepository,
		Timeout: r.timeouts.CheckObserve,
		User:    id.CurrentUserID,
		Issue: func(ctx context.Context, repo string, issue int) (runner.HostIssue, error) {
			f, err := id.IssueFacts(ctx, repo, issue)
			if err != nil {
				return runner.HostIssue{}, err
			}
			return runner.HostIssue{AuthorID: f.AuthorID, ActualRepo: f.ActualRepo, Redirected: f.Redirected}, nil
		},
		Owner: runner.StoreOwner(r.deps.Store),
	}
}

// Timeouts is the bound on each external effect.
func (r *Resident) Timeouts() effect.Timeouts { return r.timeouts }

// Executor is the effect executor with these timeouts and the intents kept in the claim store.
func (r *Resident) Executor(send effect.Sender, lookup effect.Lookuper) *effect.Executor {
	return &effect.Executor{Store: docstore.Store{C: r.deps.Store}, Send: send, Lookup: lookup, Timeouts: r.timeouts, Now: r.deps.Now, Admit: ghintake.Admit}
}

// NewWait is a wait record whose retry limit and escalation target come from the settings; the owner
// supplies the rest.
func (r *Resident) NewWait(owner, nextAction, diagnosis string, recheckAt time.Time) core.Wait {
	return core.Wait{
		Owner: owner, NextAction: nextAction, RecheckAt: recheckAt, Diagnosis: diagnosis,
		MaxRetries: r.Settings.WaitMaxRetries, Escalate: r.Settings.WaitEscalateTo,
	}
}

// Collector is the single collector of one webhook-fed repository path: its backoff-driven client, its
// polling bounds, its cursor store and its durable queue all come from the settings.
func (r *Resident) Collector(repo, path string, t forge.Transport, differ forge.Differ) (*forge.Collector, error) {
	q, ok := r.queues[core.NormalizeRepo(r.Settings.WebhookProvider, repo)]
	if !ok {
		return nil, fmt.Errorf("resident: no webhook route is configured for %s", repo)
	}
	return &forge.Collector{
		Client: &forge.Client{T: t, B: r.Budget, Backoff: r.Backoff, Now: r.deps.Now},
		Path:   path, Interval: r.Poll, Differ: differ,
		Durable: q, Cursors: r.Cursors, CursorKey: repo + path,
	}, nil
}

// Identity is the host-identity lookup with the configured cache age.
func (r *Resident) Identity(c *forge.Client) *ghidentity.GitHub {
	return &ghidentity.GitHub{Client: c, MaxAge: r.Settings.IdentityMaxAge, Now: r.deps.Now}
}

// Labels is the label host with the configured page size. Writes stay off.
func (r *Resident) Labels(t forge.Transport) labels.GitHubHost {
	return labels.GitHubHost{Repo: r.Settings.PRSourceRepository, T: t, B: r.Budget, PerPage: r.Settings.LabelsPerPage, Backoff: r.Backoff, Now: r.deps.Now}
}

// IntakeGH is the gh face of intake with the configured list limit. AllowWrite is false: enabling it is
// a separate, explicit decision of the caller.
func (r *Resident) IntakeGH(run ghintake.Runner) ghintake.GH {
	return ghintake.GH{Repo: r.Settings.PRSourceRepository, Run: run, ListLimit: r.Settings.IntakeListLimit}
}

// IntakeDuplicateLimit is the most duplicate candidates a search returns.
func (r *Resident) IntakeDuplicateLimit() int { return r.Settings.IntakeDuplicateSearchLimit }

// PRReader is the monitor's read-only PR reader with the configured list limit.
func (r *Resident) PRReader(run adapt.GHRunner) adapt.GHReader {
	return adapt.GHReader{Run: run, PRListLimit: r.Settings.MonitorPRListLimit}
}

// Ledger is the Issue context ledger with the configured size cap.
func (r *Resident) Ledger(remote ctxledger.Remote, store effect.IntentStore) *ctxledger.Ledger {
	return &ctxledger.Ledger{Remote: remote, Store: store, MaxBytes: r.Settings.ContextSummaryMaxBytes, Timeouts: r.timeouts, Backoff: r.Backoff, Now: r.deps.Now}
}

// Pass is the outcome of one pass of the loop.
type Pass struct {
	At      time.Time
	Actions []notify.Action
	Err     error
}

// Run does one pass at once and then one per tick, until ctx ends or ticks is closed. The tick period is
// the caller's (OD-17); passes never overlap. A failing pass is reported and the loop goes on.
func (r *Resident) Run(ctx context.Context, ticks <-chan time.Time, report func(Pass)) error {
	pass := func() {
		acts, err := r.Service.Tick(ctx)
		if report != nil {
			report(Pass{At: r.deps.Now(), Actions: acts, Err: err})
		}
	}
	pass()
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case _, ok := <-ticks:
			if !ok {
				return nil
			}
			if ctx.Err() != nil {
				return ctx.Err()
			}
			pass()
		}
	}
}

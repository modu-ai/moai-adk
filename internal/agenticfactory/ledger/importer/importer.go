// Package importer reads the old ledgers (the backlog queue and the factory card table) and reports how their
// rows would move into the local AgenticFactory ledger (Issue #1948, local-only AF redesign).
//
// This is the dry-run half: it opens the sources read-only (mode=ro, query_only), builds a Plan of ledger.Card
// proposals and says what it did not understand, and never writes anything — not to the sources, not to the
// ledger. Applying a plan is a separate step that only exists after the plan has been read and accepted.
package importer

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/url"
	"path/filepath"
	"sort"
	"strings"

	_ "modernc.org/sqlite" // the pure-Go "sqlite" database/sql driver the old ledgers use

	"github.com/modu-ai/moai-adk/internal/agenticfactory/core"
	"github.com/modu-ai/moai-adk/internal/agenticfactory/ledger"
)

// Sources names the old ledger files. An empty path skips that source.
type Sources struct {
	BacklogDB string // backlog.db: the operator queue (items)
	FactoryDB string // factory.db: the card table (cards)
}

// Proposal is one card the ledger would receive.
type Proposal struct {
	Source   string // "backlog", "factory" or "backlog+factory" when both described the card
	SourceID string
	Card     ledger.Card
	Notes    []string // what was decided and why, including what the target cannot hold
}

// Skip is a source row the importer did not turn into a card, with the reason.
type Skip struct {
	Source   string
	SourceID string
	Reason   string
}

// Plan is the dry-run result.
type Plan struct {
	Proposals []Proposal
	Skips     []Skip
	// Counts has one entry per source state, e.g. "backlog:picked", so the totals can be compared with what
	// the operator expects.
	Counts map[string]int
}

// Priority bands. A picked backlog item was approved for dispatch, a queued one was only captured, and the
// ledger has no field for that distinction, so it is kept as ordering: picked cards claim before queued ones,
// and within each the old queue order stands.
const (
	priorityPickedBase = 1_000_000
	prioritySpacing    = 1
)

// DefaultClass is the card class an imported card gets: C, the most careful route (plan, run and sync), since
// nothing in the old rows says how risky the work is. Triage after the import may lower it.
const DefaultClass = "C"

// Dry reads the sources and returns the plan. ctx cancels the reads.
func Dry(ctx context.Context, src Sources) (Plan, error) {
	p := Plan{Counts: map[string]int{}}
	byID := map[string]*Proposal{}
	var order []string

	if src.BacklogDB != "" {
		items, err := readBacklog(ctx, src.BacklogDB)
		if err != nil {
			return Plan{}, fmt.Errorf("importer: backlog %s: %w", src.BacklogDB, err)
		}
		for rank, it := range items {
			p.Counts["backlog:"+it.state]++
			prop, skip := fromBacklog(it, len(items)-rank)
			if skip != nil {
				p.Skips = append(p.Skips, *skip)
				continue
			}
			byID[it.id] = prop
			order = append(order, it.id)
		}
	}
	if src.FactoryDB != "" {
		cards, err := readFactory(ctx, src.FactoryDB)
		if err != nil {
			return Plan{}, fmt.Errorf("importer: factory %s: %w", src.FactoryDB, err)
		}
		for _, c := range cards {
			p.Counts["factory:"+c.state]++
			if existing, ok := byID[c.id]; ok {
				refine(existing, c)
				continue
			}
			prop, skip := fromFactory(c)
			if skip != nil {
				p.Skips = append(p.Skips, *skip)
				continue
			}
			byID[c.id] = prop
			order = append(order, c.id)
		}
	}
	for _, id := range order {
		p.Proposals = append(p.Proposals, *byID[id])
	}
	return p, nil
}

type backlogItem struct{ id, text, addedAt, specID, state string }

type factoryCard struct {
	id, state, stage, specID, worktree, owner, updatedAt, runID string
}

func readOnlyDSN(path string) string {
	v := url.Values{}
	v.Add("mode", "ro")
	v.Add("_pragma", "query_only(ON)")
	v.Add("_pragma", "busy_timeout(2000)")
	p := filepath.ToSlash(path)
	if !strings.HasPrefix(p, "/") {
		p = "/" + p
	}
	return (&url.URL{Scheme: "file", Path: p, RawQuery: v.Encode()}).String()
}

func openRO(path string) (*sql.DB, error) {
	db, err := sql.Open("sqlite", readOnlyDSN(path))
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)
	return db, nil
}

func columns(ctx context.Context, db *sql.DB, table string) (map[string]bool, error) {
	rows, err := db.QueryContext(ctx, "PRAGMA table_info("+table+")")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	cols := map[string]bool{}
	for rows.Next() {
		var cid int
		var name, typ string
		var notnull, pk int
		var dflt sql.NullString
		if err := rows.Scan(&cid, &name, &typ, &notnull, &dflt, &pk); err != nil {
			return nil, err
		}
		cols[name] = true
	}
	return cols, rows.Err()
}

func readBacklog(ctx context.Context, path string) ([]backlogItem, error) {
	db, err := openRO(path)
	if err != nil {
		return nil, err
	}
	defer db.Close()
	cols, err := columns(ctx, db, "items")
	if err != nil {
		return nil, err
	}
	if !cols["id"] || !cols["text"] || !cols["state"] {
		return nil, errors.New("the items table is missing or lacks id, text or state")
	}
	spec := "''"
	if cols["spec_id"] {
		spec = "COALESCE(spec_id,'')"
	}
	added := "''"
	if cols["added_at"] {
		added = "added_at"
	}
	rows, err := db.QueryContext(ctx, "SELECT id, text, "+added+", "+spec+", state FROM items ORDER BY seq")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []backlogItem
	for rows.Next() {
		var it backlogItem
		if err := rows.Scan(&it.id, &it.text, &it.addedAt, &it.specID, &it.state); err != nil {
			return nil, err
		}
		out = append(out, it)
	}
	return out, rows.Err()
}

// readFactory returns the latest row per card id (a card can appear under several runs). Columns an older
// schema lacks read as empty.
func readFactory(ctx context.Context, path string) ([]factoryCard, error) {
	db, err := openRO(path)
	if err != nil {
		return nil, err
	}
	defer db.Close()
	cols, err := columns(ctx, db, "cards")
	if err != nil {
		return nil, err
	}
	if !cols["card_id"] || !cols["state"] {
		return nil, errors.New("the cards table is missing or lacks card_id or state")
	}
	sel := func(c string) string {
		if cols[c] {
			return "COALESCE(" + c + ",'')"
		}
		return "''"
	}
	q := "SELECT card_id, state, " + sel("stage") + ", " + sel("spec_id") + ", " + sel("worktree_path") + ", " +
		sel("owner_label") + ", " + sel("updated_at") + ", " + sel("run_id") + " FROM cards ORDER BY " + sel("updated_at") + ", card_id"
	rows, err := db.QueryContext(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	latest := map[string]factoryCard{}
	var ids []string
	for rows.Next() {
		var c factoryCard
		if err := rows.Scan(&c.id, &c.state, &c.stage, &c.specID, &c.worktree, &c.owner, &c.updatedAt, &c.runID); err != nil {
			return nil, err
		}
		if _, seen := latest[c.id]; !seen {
			ids = append(ids, c.id)
		}
		latest[c.id] = c // ordered by updated_at ascending, so the last one read is the newest
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	sort.Strings(ids)
	out := make([]factoryCard, 0, len(ids))
	for _, id := range ids {
		out = append(out, latest[id])
	}
	return out, nil
}

func titleOf(text, specID string) string {
	t := strings.TrimSpace(text)
	if specID != "" {
		t += " (spec " + specID + ")"
	}
	return t
}

func baseCard(id, title string) ledger.Card {
	return ledger.Card{ID: id, Title: title, Spec: ledger.Spec{Class: DefaultClass}}
}

// fromBacklog maps one operator-queue row. weight is the row's place from the end of the queue, so the old
// order survives as priority.
func fromBacklog(it backlogItem, weight int) (*Proposal, *Skip) {
	if strings.TrimSpace(it.id) == "" {
		return nil, &Skip{Source: "backlog", SourceID: it.id, Reason: "the row has no id"}
	}
	card := baseCard(it.id, titleOf(it.text, it.specID))
	prop := &Proposal{Source: "backlog", SourceID: it.id, Card: card}
	note := func(s string) { prop.Notes = append(prop.Notes, s) }
	switch it.state {
	case "queued":
		prop.Card.Position = core.Position{Stage: core.StageReady}
		prop.Card.Priority = weight * prioritySpacing
	case "picked":
		prop.Card.Position = core.Position{Stage: core.StageReady}
		prop.Card.Priority = priorityPickedBase + weight*prioritySpacing
		note("picked (approved for dispatch) is kept as a higher priority band: the ledger has no field for it")
	case "hold":
		pos, err := core.FromLegacy(core.Legacy{State: core.StateOnHold, Return: core.StateQueued, Resume: "the operator releases the hold"})
		if err != nil {
			return nil, &Skip{Source: "backlog", SourceID: it.id, Reason: "hold does not map: " + err.Error()}
		}
		prop.Card.Position = pos
		prop.Card.Priority = weight * prioritySpacing
		note("hold imports as READY parked (operator-hold); the original hold reason was not stored in the old queue")
	case "dropped":
		prop.Card.Position = core.Position{Stage: core.StageCancelled}
		note("dropped imports as CANCELLED; rows are never deleted")
	default:
		return nil, &Skip{Source: "backlog", SourceID: it.id, Reason: fmt.Sprintf("unknown queue state %q", it.state)}
	}
	if it.specID != "" {
		note("spec id " + it.specID + " is carried in the title: the ledger card has no spec-id field")
	}
	return prop, nil
}

// fromFactory maps a factory.db card that no backlog row described.
func fromFactory(c factoryCard) (*Proposal, *Skip) {
	if strings.TrimSpace(c.id) == "" {
		return nil, &Skip{Source: "factory", SourceID: c.id, Reason: "the row has no card id"}
	}
	title := c.id
	if c.specID != "" {
		title = c.id + " (spec " + c.specID + ")"
	}
	prop := &Proposal{Source: "factory", SourceID: c.id, Card: baseCard(c.id, title)}
	applyFactoryState(prop, c)
	if prop.Card.Position.Stage == "" {
		return nil, &Skip{Source: "factory", SourceID: c.id, Reason: fmt.Sprintf("unknown card state %q", c.state)}
	}
	return prop, nil
}

// refine lets a factory.db row settle a card the backlog already described: a card the old factory finished or
// abandoned is done or cancelled whatever the queue still says; an in-flight card keeps the queue's position and
// gains a note about where it was.
func refine(prop *Proposal, c factoryCard) {
	prop.Source = "backlog+factory"
	switch c.state {
	case "done", "merged-local", "merged-pr", "abandoned":
		applyFactoryState(prop, c)
	default:
		prop.Notes = append(prop.Notes, fmt.Sprintf("the factory has it in state %q (stage %q, owner %q); the import starts it again at READY, its old worktree %q is left in place",
			c.state, c.stage, c.owner, c.worktree))
	}
}

func applyFactoryState(prop *Proposal, c factoryCard) {
	switch c.state {
	case "done", "merged-local", "merged-pr":
		prop.Card.Position = core.Position{Stage: core.StageDone}
		prop.Notes = append(prop.Notes, fmt.Sprintf("factory state %q imports as DONE; no merge evidence is stored in the ledger for it", c.state))
	case "abandoned":
		prop.Card.Position = core.Position{Stage: core.StageCancelled}
		prop.Notes = append(prop.Notes, "factory state abandoned imports as CANCELLED")
	case "picked", "assigned", "leased", "plan", "plan-audit", "run", "sync", "sync-audit", "merge-ready", "pr-open", "kickoff", "queued":
		prop.Card.Position = core.Position{Stage: core.StageReady}
		prop.Notes = append(prop.Notes, fmt.Sprintf("factory state %q (stage %q) restarts at READY; its old worktree %q is left in place", c.state, c.stage, c.worktree))
	}
}

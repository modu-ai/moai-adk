package adapt

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/monitor"
)

// GHRunner runs `gh` with the given arguments. The production runner is a thin os/exec wrapper
// supplied by the caller; GHReader never builds a shell string.
type GHRunner func(ctx context.Context, args ...string) ([]byte, error)

// GHReader implements monitor.PRReader with read-only gh calls. Verdict and Repair come from the
// evidence store and the dispatch record, which gh does not know.
type GHReader struct {
	Run     GHRunner
	Verdict func(issue int, head string) string
	Repair  func(issue int) bool
	// PRListLimit is the most pull requests one list call reads. It comes from configuration and is
	// required: gh's own default is silent, and a list that reaches the limit may be cut short, which is
	// reported as an error and never as a count or a "not found".
	PRListLimit int
}

func (g GHReader) listLimit() (string, int, error) {
	if g.PRListLimit <= 0 {
		return "", 0, fmt.Errorf("adapt: GHReader.PRListLimit must come from configuration")
	}
	return strconv.Itoa(g.PRListLimit), g.PRListLimit, nil
}

// readOnlyGH admits only `pr list` and `issue view` with no web/edit flags.
func readOnlyGH(args []string) bool {
	if len(args) < 2 {
		return false
	}
	switch args[0] + " " + args[1] {
	case "pr list", "issue view":
	default:
		return false
	}
	for _, a := range args {
		if a == "--web" || a == "-w" {
			return false
		}
	}
	return true
}

func (g GHReader) run(ctx context.Context, args ...string) ([]byte, error) {
	if !readOnlyGH(args) {
		return nil, fmt.Errorf("adapt: refusing non-read-only gh call %v", args)
	}
	return g.Run(ctx, args...)
}

type ghPR struct {
	Number            int    `json:"number"`
	State             string `json:"state"`
	HeadRefOid        string `json:"headRefOid"`
	StatusCheckRollup []struct {
		Name       string `json:"name"`
		Conclusion string `json:"conclusion"`
		Status     string `json:"status"`
		// StatusContext entries (commit statuses) carry context/state instead of name/status/conclusion.
		Context string `json:"context"`
		State   string `json:"state"`
	} `json:"statusCheckRollup"`
	ClosingIssuesReferences []struct {
		Number int `json:"number"`
	} `json:"closingIssuesReferences"`
}

// ReadPR returns the newest PR that closes the Issue (open first, else the newest merged one).
func (g GHReader) ReadPR(ctx context.Context, issue int) (monitor.PRInfo, bool, error) {
	limit, max, err := g.listLimit()
	if err != nil {
		return monitor.PRInfo{}, false, err
	}
	out, err := g.run(ctx, "pr", "list", "--state", "all", "--search", "#"+strconv.Itoa(issue), "--limit", limit,
		"--json", "number,state,headRefOid,statusCheckRollup,closingIssuesReferences")
	if err != nil {
		return monitor.PRInfo{}, false, err
	}
	var prs []ghPR
	if err := json.Unmarshal(out, &prs); err != nil {
		return monitor.PRInfo{}, false, fmt.Errorf("adapt: gh pr list output: %w", err)
	}
	if len(prs) >= max {
		return monitor.PRInfo{}, false, fmt.Errorf("adapt: gh pr list returned %d pull requests, the configured limit: the result may be cut short", len(prs))
	}
	var pick *ghPR
	for i := range prs {
		p := &prs[i]
		if !closes(p, issue) || p.State == "CLOSED" {
			continue
		}
		// An open PR outranks a merged one; among PRs in the same class the higher number wins.
		if pick == nil || betterPR(p, pick) {
			pick = p
		}
	}
	if pick == nil {
		return monitor.PRInfo{}, false, nil
	}
	info := monitor.PRInfo{Number: pick.Number, Open: pick.State == "OPEN", Merged: pick.State == "MERGED", HeadSHA: pick.HeadRefOid}
	for _, c := range pick.StatusCheckRollup {
		name, concl := c.Name, strings.ToLower(c.Conclusion)
		switch {
		case c.State != "" && c.Status == "":
			name, concl = c.Context, statusContextConclusion(c.State)
		case c.Status != "COMPLETED":
			concl = "pending"
		}
		info.Checks = append(info.Checks, monitor.CheckRun{Name: name, Conclusion: concl, HeadSHA: pick.HeadRefOid})
	}
	if g.Verdict != nil {
		info.ReviewVerdict = g.Verdict(issue, pick.HeadRefOid)
	}
	if g.Repair != nil {
		info.RepairDispatched = g.Repair(issue)
	}
	if info.Merged {
		iss, err := g.run(ctx, "issue", "view", strconv.Itoa(issue), "--json", "state")
		if err != nil {
			return monitor.PRInfo{}, false, err
		}
		var st struct {
			State string `json:"state"`
		}
		if err := json.Unmarshal(iss, &st); err != nil {
			return monitor.PRInfo{}, false, fmt.Errorf("adapt: gh issue view output: %w", err)
		}
		info.IssueClosed = st.State == "CLOSED"
	}
	return info, true, nil
}

// statusContextConclusion maps a commit-status state onto the check conclusions Phase reads.
func statusContextConclusion(state string) string {
	switch strings.ToUpper(state) {
	case "SUCCESS":
		return "success"
	case "FAILURE", "ERROR":
		return "failure"
	}
	return "pending" // PENDING, EXPECTED, or anything unknown: not yet a result
}

func betterPR(p, than *ghPR) bool {
	if po, to := p.State == "OPEN", than.State == "OPEN"; po != to {
		return po
	}
	return p.Number > than.Number
}

func closes(p *ghPR, issue int) bool {
	for _, r := range p.ClosingIssuesReferences {
		if r.Number == issue {
			return true
		}
	}
	return false
}

// CountOpenPRs counts open PRs in the repository.
func (g GHReader) CountOpenPRs(ctx context.Context) (int, error) {
	limit, max, err := g.listLimit()
	if err != nil {
		return 0, err
	}
	out, err := g.run(ctx, "pr", "list", "--state", "open", "--limit", limit, "--json", "number")
	if err != nil {
		return 0, err
	}
	var prs []struct{ Number int }
	if err := json.Unmarshal(out, &prs); err != nil {
		return 0, fmt.Errorf("adapt: gh pr list output: %w", err)
	}
	if len(prs) >= max {
		return 0, fmt.Errorf("adapt: gh pr list returned %d pull requests, the configured limit: the count may be too low", len(prs))
	}
	return len(prs), nil
}

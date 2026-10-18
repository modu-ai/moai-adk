package cli

// factory_age_alarm.go — SPEC-CARD-AGE-ALARM-001 (card t1838) M3: the
// display-only work-item-age alarm on `moai factory status`. The whole path
// is rendering: it resolves the flag, reads the store and the recorded
// worktrees, and writes nothing — no card-row write, no event append, no
// notification, no automatic action of any kind (REQ-CAA-007; quarantine is
// G2, a separate future card).

import (
	"context"
	"fmt"
	"io"
	"os/exec"
	"sort"
	"strings"
	"time"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/gitenv"
	"github.com/modu-ai/moai-adk/internal/homestate"
)

// factoryAgeAlarmGitTimeout bounds the one git invocation per candidate card
// (spec.md §Residual risk: a bounded per-card subprocess cost inside an
// on-demand command). A bound, not an age threshold.
const factoryAgeAlarmGitTimeout = 10 * time.Second

// factoryAgeAlarmView is one row of the additive age_alarms block
// (REQ-CAA-006): the card identity, its class and band, the class age, and
// the three signal ages of REQ-CAA-004. Ages render as Go duration strings
// (the lease_duration precedent); an unknown signal renders "unknown".
type factoryAgeAlarmView struct {
	Card     string `json:"card"`
	Run      string `json:"run"`
	State    string `json:"state"`
	Class    string `json:"class"`
	Band     string `json:"band"`
	Age      string `json:"age"`
	Stage    string `json:"stage_age"`
	Commit   string `json:"commit_age"`
	Activity string `json:"activity_age"`

	age time.Duration // sort key, not serialized
}

// computeFactoryAgeAlarms is the alarm driver of the status command. It
// performs no work beyond resolving the flag while the alarm is disabled
// (REQ-CAA-001); enabled, it resolves the bands (a refused key degrades to
// exactly one warning line naming the key and no rows — the card rows and
// every other cell survive, REQ-CAA-002), reads the events once, classifies
// and ages every covered card, bands them, and sorts oldest-first. It writes
// nothing.
func computeFactoryAgeAlarms(ctx context.Context, root, runID string, db *homestate.FactoryDB, cards []homestate.Card, now time.Time) ([]factoryAgeAlarmView, string) {
	keys := config.LoadCardAgeAlarm(root)
	if !keys.Enabled {
		return nil, ""
	}
	bands, err := config.ResolveCardAgeAlarm(keys)
	if err != nil {
		return nil, fmt.Sprintf("age-alarm warning: %v", err)
	}
	events, err := db.LoadCardEvents(ctx, runID)
	if err != nil {
		return nil, fmt.Sprintf("age-alarm warning: events read failed: %v", err)
	}
	byCard := make(map[[2]string][]homestate.CardEvent)
	for _, e := range events {
		key := [2]string{e.RunID, e.CardID}
		byCard[key] = append(byCard[key], e)
	}
	var rows []factoryAgeAlarmView
	for _, c := range cards {
		class := homestate.AgeClassOf(c)
		if class == "" {
			continue
		}
		ages := homestate.ComputeCardAges(c, byCard[[2]string{c.RunID, c.CardID}], factoryAgeAlarmCommitAt, now)
		if row, ok := bandAgeAlarm(c, class, ages, bands); ok {
			rows = append(rows, row)
		}
	}
	sort.SliceStable(rows, func(i, j int) bool { return rows[i].age > rows[j].age })
	return rows, ""
}

// bandAgeAlarm renders one card's alarm row, reporting false when the card
// crosses no threshold of its class (REQ-CAA-006: a card below every
// threshold renders no alarm row; an unknown class age renders none either —
// it cannot cross anything).
func bandAgeAlarm(c homestate.Card, class string, ages homestate.CardAges, bands config.CardAgeAlarmSettings) (factoryAgeAlarmView, bool) {
	sig, warn, alert := ages.NoProgress, bands.StallWarn, bands.StallAlert
	if class == homestate.AgeClassGateWait {
		sig, warn, alert = ages.GateWait, bands.GateWarn, bands.GateAlert
	}
	if !sig.Known || sig.Age < warn {
		return factoryAgeAlarmView{}, false
	}
	band := "warn"
	if sig.Age >= alert {
		band = "alert"
	}
	return factoryAgeAlarmView{
		Card: c.CardID, Run: c.RunID, State: c.State, Class: class, Band: band,
		Age:      sig.Age.String(),
		Stage:    ageAlarmSignal(ages.Stage),
		Commit:   ageAlarmSignal(ages.Commit),
		Activity: ageAlarmSignal(ages.Activity),
		age:      sig.Age,
	}, true
}

// ageAlarmSignal renders one signal age; unknown stays unknown.
func ageAlarmSignal(s homestate.AgeSignal) string {
	if !s.Known {
		return "unknown"
	}
	return s.Age.String()
}

// writeFactoryAgeAlarmText renders the alarm section of the status text: one
// `age-alarm` line per row (the caller sorted them oldest-first) and, when a
// key was refused, exactly one warning line. Both shapes are tag-prefixed
// with the literal token `age-alarm` (plan.md M3 (2), iter1 D3 — the token a
// reader greps for). Nothing prints when the alarm is off and the keys are
// sound, so the disabled output is byte-identical (REQ-CAA-001).
func writeFactoryAgeAlarmText(w io.Writer, rows []factoryAgeAlarmView, warning string) {
	for _, r := range rows {
		_, _ = fmt.Fprintf(w, "age-alarm card=%s run=%s state=%s class=%s band=%s age=%s stage=%s commit=%s activity=%s\n",
			r.Card, r.Run, r.State, r.Class, r.Band, r.Age, r.Stage, r.Commit, r.Activity)
	}
	if warning != "" {
		_, _ = fmt.Fprintln(w, warning)
	}
}

// factoryAgeAlarmCommitAt reads the committer timestamp of the head commit of
// the worktree at path — REQ-CAA-004's commit signal: one bounded, read-only
// git invocation against the card's recorded WorktreePath, reading timestamps
// git already recorded. The invocation is repo-isolated and helper-free: the
// child environment is gitenv.Env (an inherited GIT_DIR / GIT_WORK_TREE would
// outrank -C and redirect the read into a foreign repository, sync-audit iter1
// F1) and signature rendering is pinned OFF so no repo-configured helper
// program executes during the read (F2). A missing or unreadable worktree, a
// failed or timed-out invocation, and an unparseable stamp all render unknown
// — never a fabricated value, never a failed status command.
func factoryAgeAlarmCommitAt(worktreePath string) (time.Time, bool) {
	ctx, cancel := context.WithTimeout(context.Background(), factoryAgeAlarmGitTimeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, "git", "-C", worktreePath, "-c", "log.showSignature=false", "log", "-1", "--format=%cI")
	// gitenv.Env, not os.Environ — the house pattern of internal/core/git
	// execGit: a lane or hook process exports GIT_DIR / GIT_WORK_TREE, and an
	// inherited copy outranks -C.
	cmd.Env = append(gitenv.Env(), "GIT_TERMINAL_PROMPT=0")
	out, err := cmd.Output()
	if err != nil {
		return time.Time{}, false
	}
	at, err := time.Parse(time.RFC3339, strings.TrimSpace(string(out)))
	if err != nil {
		return time.Time{}, false
	}
	return at, true
}

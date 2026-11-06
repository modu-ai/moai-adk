package config

// agentic_factory_settings.go — reader and validator for the NEW file
// .moai/config/sections/agentic-factory.yaml (SPEC-AGENTIC-FACTORY-001
// design.md §6.1, leader decision DEC-AF-25).
//
// No defaults. Every key below is required: an absent (or null) key is an error
// whose message carries the full dotted key name, and a rejected value is never
// clamped, coerced or replaced. This file only reads; it never writes.
//
// Two-stage reading:
//
//	stage 1 — the file is missing, or workflow.agentic_factory.enabled is
//	          absent / null / false: Agentic Factory is NOT configured. This is
//	          not an error, and no other key is read or validated. A non-bool
//	          enabled is an error.
//	stage 2 — enabled is true: every other required key must be present and in
//	          range. Any failure is a validation error naming the key, and the
//	          result is also reported as not configured.
//
// Keys deliberately NOT read here, because their names are still undecided:
// webhook.routes[].secret_ref and loop.tick_period. A webhook secret VALUE must never live in this file; a
// key named "secret" under webhook.routes[] is refused (the value is never
// echoed). Unknown keys under workflow.agentic_factory are ignored (forward
// compatible); workflow.yaml is never consulted.
//
// provisional: key names follow design.md §6.1 and may be renamed in one place.

import (
	"errors"
	"fmt"
	"math/big"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

const (
	agenticFactorySettingsFile = "agentic-factory.yaml"
	afsPrefix                  = "workflow.agentic_factory."
)

// ErrAgenticFactoryNotConfigured is returned by LoadAgenticFactorySettings in
// stage 1 (file missing, or enabled absent/null/false). It is not a validation
// failure: errors.Is(err, ErrInvalidConfig) is false for it.
var ErrAgenticFactoryNotConfigured = errors.New("agentic factory is not configured")

// AgenticFactoryBackoff is the GitHub provider backoff policy.
type AgenticFactoryBackoff struct {
	MinSecondaryWait time.Duration
	Base             time.Duration
	Max              time.Duration
	MaxAttempts      int
}

// AgenticFactoryPoll is the collector polling policy. Only presence and
// parseability are checked: the rules Min>0, Max>=Min and Safety>0 are named in
// design.md §6.1 as "needed" but the SPEC has NOT adopted them, so this reader
// does not enforce them.
type AgenticFactoryPoll struct {
	Min    time.Duration
	Max    time.Duration
	Safety time.Duration
}

// AgenticFactoryQueueLimits sizes one doc-backed queue (webhook queue or agent
// inbox; they are separate instances, DEC-AF-25 D4).
type AgenticFactoryQueueLimits struct {
	MaxDocBytes int // >= 1
	SeenLimit   int
	MaxIDBytes  int
}

// AgenticFactorySettings holds every confirmed key of agentic-factory.yaml.
type AgenticFactorySettings struct {
	ClaimStoreDir string

	GitHubBackoff AgenticFactoryBackoff
	CollectorPoll AgenticFactoryPoll

	BudgetReserve        int
	BudgetReportResource string // empty allowed (default bucket)

	IdentityMaxAge             time.Duration // 0 allowed
	LabelsPerPage              int
	IntakeDuplicateSearchLimit int
	IntakeListLimit            int // >= 1
	MonitorPRListLimit         int // >= 1

	CursorMaxDocBytes int // >= 1
	CursorMaxChunks   int

	WebhookQueue AgenticFactoryQueueLimits
	NotifyInbox  AgenticFactoryQueueLimits

	WebhookMaxBody      int
	WebhookProvider     string
	WebhookRoutes       []string // repositories, one per route
	NotifyRenotifyAfter time.Duration

	WaitMaxRetries int
	WaitEscalateTo string

	PRSourceRepository string

	ContextSummaryMaxBytes int // <=0 means no cap downstream
}

// AgenticFactoryConfiguredResult is the judgment "is Agentic Factory configured".
// Stage-1 not-configured has Err==nil and a human Reason; a stage-2 failure has
// Err!=nil (naming the key) and a Reason.
type AgenticFactoryConfiguredResult struct {
	Configured bool
	Reason     string
	Err        error
}

// AgenticFactoryConfigured judges whether Agentic Factory is configured for
// projectRoot. It never writes.
func AgenticFactoryConfigured(projectRoot string) AgenticFactoryConfiguredResult {
	_, reason, err := readAgenticFactorySettings(projectRoot)
	if err != nil {
		return AgenticFactoryConfiguredResult{Reason: "invalid agentic-factory.yaml: " + err.Error(), Err: err}
	}
	if reason != "" {
		return AgenticFactoryConfiguredResult{Reason: reason}
	}
	return AgenticFactoryConfiguredResult{Configured: true}
}

// LoadAgenticFactorySettings returns the settings only when they are fully
// valid. Stage 1 yields ErrAgenticFactoryNotConfigured (wrapped with the
// reason); a validation failure yields an error naming the key.
func LoadAgenticFactorySettings(projectRoot string) (AgenticFactorySettings, error) {
	s, reason, err := readAgenticFactorySettings(projectRoot)
	if err != nil {
		return AgenticFactorySettings{}, err
	}
	if reason != "" {
		return AgenticFactorySettings{}, fmt.Errorf("%w: %s", ErrAgenticFactoryNotConfigured, reason)
	}
	return s, nil
}

// readAgenticFactorySettings returns (settings, "", nil) when configured,
// (zero, reason, nil) in stage 1, and (zero, "", err) on any error.
func readAgenticFactorySettings(projectRoot string) (AgenticFactorySettings, string, error) {
	dir := filepath.Join(projectRoot, ".moai", "config", "sections")
	path := filepath.Join(dir, agenticFactorySettingsFile)

	var doc yaml.Node
	found, err := loadYAMLFile(dir, agenticFactorySettingsFile, &doc)
	if err != nil {
		return AgenticFactorySettings{}, "", fmt.Errorf("%s: %w", path, err)
	}
	if !found {
		return AgenticFactorySettings{}, agenticFactorySettingsFile + " not found at " + path, nil
	}

	r := &afsReader{path: path, root: &doc}

	// Stage 1.
	enabledKey := afsPrefix + "enabled"
	en, err := afsLookup(r.root, enabledKey)
	if err != nil {
		return AgenticFactorySettings{}, "", r.wrapLookup(enabledKey, err)
	}
	if en == nil || agenticFactoryUnset(en) {
		return AgenticFactorySettings{}, enabledKey + " is not set", nil
	}
	var enabled bool
	if en.Kind != yaml.ScalarNode || en.Tag != "!!bool" || en.Decode(&enabled) != nil {
		return AgenticFactorySettings{}, "", agenticFactoryFileError(path, enabledKey, "must be a bool (true or false)", agenticFactoryNodeKind(en))
	}
	if !enabled {
		return AgenticFactorySettings{}, enabledKey + " is false", nil
	}

	// Stage 2.
	var s AgenticFactorySettings
	r.guardSecrets()

	s.ClaimStoreDir = r.str("claim_store.dir", false)

	b := &s.GitHubBackoff
	b.MinSecondaryWait = r.duration("provider.github.backoff.min_secondary_wait")
	b.Base = r.duration("provider.github.backoff.base")
	b.Max = r.duration("provider.github.backoff.max")
	b.MaxAttempts = r.integer("provider.github.backoff.max_attempts")
	r.check(b.MinSecondaryWait > 0, "provider.github.backoff.min_secondary_wait", "must be > 0", b.MinSecondaryWait)
	r.check(b.Base > 0, "provider.github.backoff.base", "must be > 0", b.Base)
	r.check(b.Max > 0, "provider.github.backoff.max", "must be > 0", b.Max)
	r.check(b.Max >= b.MinSecondaryWait, "provider.github.backoff.max", "must be >= provider.github.backoff.min_secondary_wait", b.Max)
	r.check(b.Max >= b.Base, "provider.github.backoff.max", "must be >= provider.github.backoff.base", b.Max)
	r.check(b.MaxAttempts > 0, "provider.github.backoff.max_attempts", "must be an integer > 0", b.MaxAttempts)

	// Presence and parseability only: Min>0, Max>=Min and Safety>0 are not
	// adopted by the SPEC (design.md §6.1), so no range rule is applied.
	s.CollectorPoll.Min = r.duration("collector.poll.min")
	s.CollectorPoll.Max = r.duration("collector.poll.max")
	s.CollectorPoll.Safety = r.duration("collector.poll.safety")

	s.BudgetReserve = r.integer("budget.reserve") // presence only
	s.BudgetReportResource = r.str("budget.report_resource", true)

	s.IdentityMaxAge = r.duration("identity.max_age")
	r.check(s.IdentityMaxAge >= 0, "identity.max_age", "must be >= 0", s.IdentityMaxAge)

	s.LabelsPerPage = r.integer("labels.per_page")
	r.check(s.LabelsPerPage > 0, "labels.per_page", "must be an integer > 0", s.LabelsPerPage)
	s.IntakeDuplicateSearchLimit = r.integer("intake.duplicate_search_limit")
	r.check(s.IntakeDuplicateSearchLimit > 0, "intake.duplicate_search_limit", "must be an integer > 0", s.IntakeDuplicateSearchLimit)
	s.IntakeListLimit = r.integer("intake.list_limit")
	r.check(s.IntakeListLimit >= 1, "intake.list_limit", "must be an integer >= 1", s.IntakeListLimit)
	s.MonitorPRListLimit = r.integer("monitor.pr_list_limit")
	r.check(s.MonitorPRListLimit >= 1, "monitor.pr_list_limit", "must be an integer >= 1", s.MonitorPRListLimit)

	s.CursorMaxDocBytes = r.integer("cursor.max_doc_bytes")
	r.check(s.CursorMaxDocBytes >= 1, "cursor.max_doc_bytes", "must be an integer >= 1", s.CursorMaxDocBytes)
	s.CursorMaxChunks = r.integer("cursor.max_chunks")
	r.check(s.CursorMaxChunks >= 0, "cursor.max_chunks", "must be an integer >= 0", s.CursorMaxChunks)

	// The webhook delivery id reference length is 36 (design section 6.1), but no
	// minimum is adopted for it: only the inbox ids are evidenced, so 0 is the floor here.
	s.WebhookQueue = r.queue("queue.webhook", 0)
	s.NotifyInbox = r.queue("notify.inbox", afsInboxMinIDBytes)

	s.WebhookMaxBody = r.integer("webhook.max_body")
	r.check(s.WebhookMaxBody >= 0, "webhook.max_body", "must be an integer >= 0", s.WebhookMaxBody)
	s.WebhookProvider = r.str("webhook.provider", false)
	s.WebhookRoutes = r.routes()

	s.NotifyRenotifyAfter = r.duration("notify.renotify_after")
	r.check(s.NotifyRenotifyAfter > 0, "notify.renotify_after", "must be > 0", s.NotifyRenotifyAfter)

	s.WaitMaxRetries = r.integer("wait.max_retries")
	r.check(s.WaitMaxRetries >= 1, "wait.max_retries", "must be an integer >= 1", s.WaitMaxRetries)
	s.WaitEscalateTo = r.str("wait.escalate_to", false)

	s.PRSourceRepository = r.str("pr_source.repository", false)
	r.check(agenticFactoryRepoValid(s.PRSourceRepository), "pr_source.repository", agenticFactoryRepoMessageOne, nil)

	s.ContextSummaryMaxBytes = r.integer("context.summary_max_bytes") // presence only; <=0 = no cap downstream

	if r.err != nil {
		return AgenticFactorySettings{}, "", r.err
	}
	return s, "", nil
}

const agenticFactoryRepoMessageOne = "must be a repository path: owner/name or group/subgroup/project (>= 2 non-empty '/'-separated segments, no wildcard, no URL scheme, no whitespace)"

// afsReader reads keys relative to workflow.agentic_factory. The first failure
// is sticky: later reads return zero values and the first error is reported.
type afsReader struct {
	path string
	root *yaml.Node
	err  error
}

func (r *afsReader) fail(key, msg string, got any) {
	if r.err == nil {
		r.err = agenticFactoryFileError(r.path, afsPrefix+key, msg, got)
	}
}

// check records a failure when ok is false.
func (r *afsReader) check(ok bool, key, msg string, got any) {
	if r.err == nil && !ok {
		r.fail(key, msg, got)
	}
}

func (r *afsReader) wrapLookup(key string, err error) error {
	return agenticFactoryFileError(r.path, key, err.Error(), nil)
}

// node returns the node for a required key, or nil after recording an error.
func (r *afsReader) node(rel string) *yaml.Node {
	if r.err != nil {
		return nil
	}
	n, err := afsLookup(r.root, afsPrefix+rel)
	if err != nil {
		r.err = r.wrapLookup(afsPrefix+rel, err)
		return nil
	}
	if n == nil || agenticFactoryUnset(n) {
		r.fail(rel, "missing required key (no default exists)", nil)
		return nil
	}
	return n
}

func (r *afsReader) str(rel string, allowEmpty bool) string {
	n := r.node(rel)
	if n == nil {
		return ""
	}
	if n.Kind != yaml.ScalarNode || n.Tag != "!!str" {
		r.fail(rel, "must be a string", agenticFactoryNodeKind(n))
		return ""
	}
	if !allowEmpty && n.Value == "" {
		r.fail(rel, "must be a non-empty string", nil)
		return ""
	}
	return n.Value
}

func (r *afsReader) integer(rel string) int {
	n := r.node(rel)
	if n == nil {
		return 0
	}
	var v int
	if n.Kind != yaml.ScalarNode || n.Tag != "!!int" || n.Decode(&v) != nil {
		r.fail(rel, "must be an integer", agenticFactoryNodeKind(n))
		return 0
	}
	return v
}

// duration reads a YAML string parsed by time.ParseDuration; a bare number is rejected.
func (r *afsReader) duration(rel string) time.Duration {
	n := r.node(rel)
	if n == nil {
		return 0
	}
	if n.Kind != yaml.ScalarNode || n.Tag != "!!str" {
		r.fail(rel, "must be a duration string such as 30s (a bare number is rejected)", agenticFactoryNodeKind(n))
		return 0
	}
	d, err := time.ParseDuration(n.Value)
	if err != nil {
		r.fail(rel, "must be a duration string such as 30s", n.Value)
		return 0
	}
	return d
}

// afsInboxMinIDBytes is the smallest notify.inbox.max_id_bytes the agent inbox can work
// with: its message ids are 32 bytes (messageID in internal/agenticfactory/monitor/notify/inbox.go,
// hex of 16 bytes). Below it every Notify fails as "not delivered" and burns retries.
// This duplicates that value and must change together with it.
const afsInboxMinIDBytes = 32

// queue reads and validates one doc-backed queue block. minIDBytes > 0 sets a floor
// for max_id_bytes (0 is then rejected too); 0 keeps the generic rule.
func (r *afsReader) queue(rel string, minIDBytes int) AgenticFactoryQueueLimits {
	q := AgenticFactoryQueueLimits{
		MaxDocBytes: r.integer(rel + ".max_doc_bytes"),
		SeenLimit:   r.integer(rel + ".seen_limit"),
		MaxIDBytes:  r.integer(rel + ".max_id_bytes"),
	}
	// 0 is rejected for max_doc_bytes: it cannot be told apart from "no limit".
	r.check(q.MaxDocBytes >= 1, rel+".max_doc_bytes", "must be an integer >= 1", q.MaxDocBytes)
	r.check(q.SeenLimit >= 0, rel+".seen_limit", "must be an integer >= 0", q.SeenLimit)
	r.check(q.MaxIDBytes >= 0, rel+".max_id_bytes", "must be an integer >= 0", q.MaxIDBytes)
	if minIDBytes > 0 {
		r.check(q.MaxIDBytes >= minIDBytes, rel+".max_id_bytes", fmt.Sprintf("must be an integer >= %d (agent inbox message ids are %d bytes)", minIDBytes, minIDBytes), q.MaxIDBytes)
	}
	r.check(q.SeenLimit == 0 || q.MaxIDBytes > 0, rel+".max_id_bytes", "must be > 0 when "+afsPrefix+rel+".seen_limit > 0", q.MaxIDBytes)
	// The inequality is always evaluated (big-int safe against overflow).
	if r.err == nil {
		lhs := new(big.Int).Mul(big.NewInt(int64(q.SeenLimit)), new(big.Int).Add(big.NewInt(int64(q.MaxIDBytes)), big.NewInt(3)))
		lhs.Mul(lhs, big.NewInt(2))
		if lhs.Cmp(big.NewInt(int64(q.MaxDocBytes))) > 0 {
			r.fail(rel+".seen_limit", "seen_limit*(max_id_bytes+3)*2 must be <= "+afsPrefix+rel+".max_doc_bytes", q.SeenLimit)
		}
	}
	return q
}

// routes reads webhook.routes[].repository. The key must exist; an empty list is accepted.
func (r *afsReader) routes() []string {
	n := r.node("webhook.routes")
	if n == nil {
		return nil
	}
	if n.Kind != yaml.SequenceNode {
		r.fail("webhook.routes", "must be a list", agenticFactoryNodeKind(n))
		return nil
	}
	out := make([]string, 0, len(n.Content))
	for i, e := range n.Content {
		key := fmt.Sprintf("webhook.routes[%d].repository", i)
		if e.Kind != yaml.MappingNode {
			r.fail(fmt.Sprintf("webhook.routes[%d]", i), "must be a map with a repository", agenticFactoryNodeKind(e))
			return nil
		}
		repo, err := afsChild(e, "repository")
		if err != nil {
			r.fail(key, err.Error(), nil)
			return nil
		}
		if repo == nil || agenticFactoryUnset(repo) {
			r.fail(key, "missing required key (no default exists)", nil)
			return nil
		}
		if repo.Kind != yaml.ScalarNode || repo.Tag != "!!str" || !agenticFactoryRepoValid(repo.Value) {
			r.fail(key, agenticFactoryRepoMessageOne, agenticFactoryNodeKind(repo))
			return nil
		}
		out = append(out, repo.Value)
	}
	return out
}

// guardSecrets refuses a key named "secret" anywhere under webhook.routes[].
// The value is never echoed: a committed config must hold a reference, not a secret.
func (r *afsReader) guardSecrets() {
	n, err := afsLookup(r.root, afsPrefix+"webhook.routes")
	if err != nil || n == nil || n.Kind != yaml.SequenceNode {
		return // shape errors are reported by routes()
	}
	for i, e := range n.Content {
		found, cycle := afsHasSecretKey(e)
		if cycle {
			r.fail(fmt.Sprintf("webhook.routes[%d]", i),
				"a YAML alias refers back to itself (a cycle), which this file does not allow (the value was not echoed)", nil)
			return
		}
		if found {
			r.fail(fmt.Sprintf("webhook.routes[%d].secret", i),
				"a secret value must not be stored in this file; use a reference to where the secret is read from (the value was not echoed)", nil)
			return
		}
	}
}

// afsHasSecretKey reports whether a mapping key named "secret" exists under n,
// following YAML aliases. cycle is true when an alias leads back to a node that
// is still being walked (a self-referencing anchor); the walk then stops. A
// non-cyclic anchor used several times is not a cycle.
func afsHasSecretKey(n *yaml.Node) (found, cycle bool) {
	return afsWalkSecret(n, map[*yaml.Node]bool{}, map[*yaml.Node]bool{})
}

// afsWalkSecret keeps onPath (nodes on the current descent, for cycle detection)
// and clean (nodes fully walked without a finding, so a shared anchor is walked once).
func afsWalkSecret(n *yaml.Node, onPath, clean map[*yaml.Node]bool) (found, cycle bool) {
	if n == nil {
		return false, false
	}
	if n.Kind == yaml.AliasNode {
		n = n.Alias
		if n == nil {
			return false, false
		}
	}
	if onPath[n] {
		return false, true
	}
	if clean[n] || (n.Kind != yaml.MappingNode && n.Kind != yaml.SequenceNode) {
		return false, false
	}
	onPath[n] = true
	defer delete(onPath, n)
	if n.Kind == yaml.MappingNode {
		for i := 0; i+1 < len(n.Content); i += 2 {
			if n.Content[i].Value == "secret" {
				return true, false
			}
			if f, c := afsWalkSecret(n.Content[i+1], onPath, clean); f || c {
				return f, c
			}
		}
	} else {
		for _, e := range n.Content {
			if f, c := afsWalkSecret(e, onPath, clean); f || c {
				return f, c
			}
		}
	}
	clean[n] = true
	return false, false
}

// afsLookup walks a dotted path from the document root. It returns (nil, nil)
// when any segment is absent or null, and an error when an intermediate value
// is not a map or a key is duplicated.
func afsLookup(root *yaml.Node, dotted string) (*yaml.Node, error) {
	cur := root
	if cur.Kind == yaml.DocumentNode {
		if len(cur.Content) == 0 {
			return nil, nil
		}
		cur = cur.Content[0]
	}
	segs := strings.Split(dotted, ".")
	for i, seg := range segs {
		if cur.Kind == yaml.AliasNode {
			cur = cur.Alias
		}
		if cur.Kind == 0 || (cur.Kind == yaml.ScalarNode && cur.Tag == "!!null") {
			return nil, nil
		}
		if cur.Kind != yaml.MappingNode {
			where := "the document root"
			if i > 0 {
				where = strings.Join(segs[:i], ".")
			}
			return nil, fmt.Errorf("%s must be a map to hold %q", where, seg)
		}
		next, err := afsChild(cur, seg)
		if err != nil {
			return nil, err
		}
		if next == nil {
			return nil, nil
		}
		cur = next
	}
	if cur.Kind == yaml.AliasNode {
		cur = cur.Alias
	}
	return cur, nil
}

// afsChild returns the value of key in a mapping node, nil when absent, and an
// error when the key appears twice (yaml.v3 does not reject duplicates when
// decoding into a Node, and which one wins is not guessed).
func afsChild(m *yaml.Node, key string) (*yaml.Node, error) {
	var found *yaml.Node
	for i := 0; i+1 < len(m.Content); i += 2 {
		if m.Content[i].Value != key {
			continue
		}
		if found != nil {
			return nil, fmt.Errorf("duplicate key %q", key)
		}
		found = m.Content[i+1]
	}
	return found, nil
}

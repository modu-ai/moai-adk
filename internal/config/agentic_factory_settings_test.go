package config

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"gopkg.in/yaml.v3"
)

// afsDoc is a nested YAML document under construction.
type afsDoc = map[string]any

// afsValidDoc returns a fresh, fully valid document.
func afsValidDoc() afsDoc {
	return afsDoc{"workflow": afsDoc{"agentic_factory": afsDoc{
		"enabled":     true,
		"claim_store": afsDoc{"dir": "state/claims"},
		"provider":    afsDoc{"github": afsDoc{"backoff": afsDoc{"min_secondary_wait": "60s", "base": "1s", "max": "5m", "max_attempts": 5}}},
		"collector":   afsDoc{"poll": afsDoc{"min": "10s", "max": "5m", "safety": "15m"}},
		"budget":      afsDoc{"reserve": 10, "report_resource": "core"},
		"identity":    afsDoc{"max_age": "10m"},
		"labels":      afsDoc{"per_page": 50},
		"intake":      afsDoc{"duplicate_search_limit": 20, "list_limit": 100},
		"monitor":     afsDoc{"pr_list_limit": 200},
		"cursor":      afsDoc{"max_doc_bytes": 65536, "max_chunks": 4},
		"queue":       afsDoc{"webhook": afsDoc{"max_doc_bytes": 262144, "seen_limit": 100, "max_id_bytes": 36}},
		"notify":      afsDoc{"inbox": afsDoc{"max_doc_bytes": 262144, "seen_limit": 50, "max_id_bytes": 32}, "renotify_after": "30m"},
		"webhook":     afsDoc{"max_body": 1048576, "provider": "github", "routes": []any{afsDoc{"repository": "Owner/Name"}}},
		"wait":        afsDoc{"max_retries": 3, "escalate_to": "operator"},
		"pr_source":   afsDoc{"repository": "owner/name"},
		"context":     afsDoc{"summary_max_bytes": 4096},
	}}}
}

func afsAF(d afsDoc) afsDoc { return d["workflow"].(afsDoc)["agentic_factory"].(afsDoc) }

// afsSet sets a dotted path (relative to workflow.agentic_factory).
func afsSet(d afsDoc, rel string, v any) {
	segs := strings.Split(rel, ".")
	cur := afsAF(d)
	for _, s := range segs[:len(segs)-1] {
		cur = cur[s].(afsDoc)
	}
	cur[segs[len(segs)-1]] = v
}

// afsDel removes a dotted path.
func afsDel(d afsDoc, rel string) {
	segs := strings.Split(rel, ".")
	cur := afsAF(d)
	for _, s := range segs[:len(segs)-1] {
		cur = cur[s].(afsDoc)
	}
	delete(cur, segs[len(segs)-1])
}

func afsWriteRaw(t *testing.T, raw string) string {
	t.Helper()
	root := t.TempDir()
	dir := filepath.Join(root, ".moai", "config", "sections")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "agentic-factory.yaml"), []byte(raw), 0o644); err != nil {
		t.Fatal(err)
	}
	return root
}

func afsWriteDoc(t *testing.T, d afsDoc) string {
	t.Helper()
	b, err := yaml.Marshal(d)
	if err != nil {
		t.Fatal(err)
	}
	return afsWriteRaw(t, string(b))
}

// afsWantInvalid asserts both entry points report a stage-2 style failure naming key.
func afsWantInvalid(t *testing.T, root, key string) {
	t.Helper()
	quoted := `field "` + afsPrefix + key + `"`
	_, err := LoadAgenticFactorySettings(root)
	if err == nil {
		t.Fatalf("expected an error naming %s, got nil", key)
	}
	if !strings.Contains(err.Error(), quoted) {
		t.Fatalf("error does not name %s: %v", quoted, err)
	}
	if !errors.Is(err, ErrInvalidConfig) || errors.Is(err, ErrAgenticFactoryNotConfigured) {
		t.Fatalf("error must wrap ErrInvalidConfig and not be a not-configured outcome: %v", err)
	}
	res := AgenticFactoryConfigured(root)
	if res.Configured || res.Err == nil || res.Reason == "" || !strings.Contains(res.Err.Error(), quoted) {
		t.Fatalf("Configured result for %s = %+v", key, res)
	}
}

func afsWantValid(t *testing.T, root string) AgenticFactorySettings {
	t.Helper()
	s, err := LoadAgenticFactorySettings(root)
	if err != nil {
		t.Fatalf("expected valid settings, got: %v", err)
	}
	if res := AgenticFactoryConfigured(root); !res.Configured || res.Err != nil {
		t.Fatalf("Configured = %+v", res)
	}
	return s
}

func afsWantNotConfigured(t *testing.T, root string) {
	t.Helper()
	_, err := LoadAgenticFactorySettings(root)
	if !errors.Is(err, ErrAgenticFactoryNotConfigured) || errors.Is(err, ErrInvalidConfig) {
		t.Fatalf("loader err = %v, want ErrAgenticFactoryNotConfigured only", err)
	}
	res := AgenticFactoryConfigured(root)
	if res.Configured || res.Err != nil || res.Reason == "" {
		t.Fatalf("result = %+v, want not configured with Err==nil and a Reason", res)
	}
}

// afsRequired lists every required key with the kind of value it takes.
var afsRequired = []struct{ key, kind string }{
	{"claim_store.dir", "str"},
	{"provider.github.backoff.min_secondary_wait", "dur"},
	{"provider.github.backoff.base", "dur"},
	{"provider.github.backoff.max", "dur"},
	{"provider.github.backoff.max_attempts", "int"},
	{"collector.poll.min", "dur"},
	{"collector.poll.max", "dur"},
	{"collector.poll.safety", "dur"},
	{"budget.reserve", "int"},
	{"budget.report_resource", "str"},
	{"identity.max_age", "dur"},
	{"labels.per_page", "int"},
	{"intake.duplicate_search_limit", "int"},
	{"intake.list_limit", "int"},
	{"monitor.pr_list_limit", "int"},
	{"cursor.max_doc_bytes", "int"},
	{"cursor.max_chunks", "int"},
	{"queue.webhook.max_doc_bytes", "int"},
	{"queue.webhook.seen_limit", "int"},
	{"queue.webhook.max_id_bytes", "int"},
	{"notify.inbox.max_doc_bytes", "int"},
	{"notify.inbox.seen_limit", "int"},
	{"notify.inbox.max_id_bytes", "int"},
	{"webhook.max_body", "int"},
	{"webhook.provider", "str"},
	{"webhook.routes", "list"},
	{"notify.renotify_after", "dur"},
	{"wait.max_retries", "int"},
	{"wait.escalate_to", "str"},
	{"pr_source.repository", "repo"},
	{"context.summary_max_bytes", "int"},
}

func TestAgenticFactorySettings_ValidLoads(t *testing.T) {
	s := afsWantValid(t, afsWriteDoc(t, afsValidDoc()))
	want := AgenticFactorySettings{
		ClaimStoreDir:              "state/claims",
		GitHubBackoff:              AgenticFactoryBackoff{MinSecondaryWait: 60 * time.Second, Base: time.Second, Max: 5 * time.Minute, MaxAttempts: 5},
		CollectorPoll:              AgenticFactoryPoll{Min: 10 * time.Second, Max: 5 * time.Minute, Safety: 15 * time.Minute},
		BudgetReserve:              10,
		BudgetReportResource:       "core",
		IdentityMaxAge:             10 * time.Minute,
		LabelsPerPage:              50,
		IntakeDuplicateSearchLimit: 20,
		IntakeListLimit:            100,
		MonitorPRListLimit:         200,
		CursorMaxDocBytes:          65536,
		CursorMaxChunks:            4,
		WebhookQueue:               AgenticFactoryQueueLimits{MaxDocBytes: 262144, SeenLimit: 100, MaxIDBytes: 36},
		NotifyInbox:                AgenticFactoryQueueLimits{MaxDocBytes: 262144, SeenLimit: 50, MaxIDBytes: 32},
		WebhookMaxBody:             1048576,
		WebhookProvider:            "github",
		WebhookRoutes:              []string{"Owner/Name"},
		NotifyRenotifyAfter:        30 * time.Minute,
		WaitMaxRetries:             3,
		WaitEscalateTo:             "operator",
		PRSourceRepository:         "owner/name",
		ContextSummaryMaxBytes:     4096,
	}
	if !reflect.DeepEqual(s, want) {
		t.Fatalf("settings mismatch:\n got %+v\nwant %+v", s, want)
	}
}

func TestAgenticFactorySettings_MissingKey(t *testing.T) {
	for _, k := range afsRequired {
		t.Run(k.key, func(t *testing.T) {
			d := afsValidDoc()
			afsDel(d, k.key)
			afsWantInvalid(t, afsWriteDoc(t, d), k.key)
		})
		t.Run(k.key+"/null", func(t *testing.T) {
			d := afsValidDoc()
			afsSet(d, k.key, nil)
			afsWantInvalid(t, afsWriteDoc(t, d), k.key)
		})
	}
}

func TestAgenticFactorySettings_WrongType(t *testing.T) {
	bad := map[string][]any{
		"dur":  {30, 1.5, true, "not-a-duration", []any{"1s"}, 0},
		"int":  {"5", 1.5, true, []any{1}, afsDoc{"a": 1}},
		"str":  {5, true, []any{"x"}},
		"list": {"a/b", 5, afsDoc{"a": 1}},
		"repo": {5, true, []any{"a/b"}},
	}
	for _, k := range afsRequired {
		for i, v := range bad[k.kind] {
			t.Run(k.key+"/"+string(rune('a'+i)), func(t *testing.T) {
				d := afsValidDoc()
				afsSet(d, k.key, v)
				afsWantInvalid(t, afsWriteDoc(t, d), k.key)
			})
		}
	}
}

func TestAgenticFactorySettings_Ranges(t *testing.T) {
	tests := []struct {
		name string
		set  map[string]any
		key  string // key the error must name; empty = must load
	}{
		{"per_page 0", map[string]any{"labels.per_page": 0}, "labels.per_page"},
		{"per_page -1", map[string]any{"labels.per_page": -1}, "labels.per_page"},
		{"per_page 1", map[string]any{"labels.per_page": 1}, ""},
		{"duplicate_search_limit 0", map[string]any{"intake.duplicate_search_limit": 0}, "intake.duplicate_search_limit"},
		{"duplicate_search_limit 1", map[string]any{"intake.duplicate_search_limit": 1}, ""},
		{"max_age 0s", map[string]any{"identity.max_age": "0s"}, ""},
		{"max_age -1s", map[string]any{"identity.max_age": "-1s"}, "identity.max_age"},
		{"min_secondary_wait 0s", map[string]any{"provider.github.backoff.min_secondary_wait": "0s", "provider.github.backoff.max": "5m"}, "provider.github.backoff.min_secondary_wait"},
		{"base 0s", map[string]any{"provider.github.backoff.base": "0s"}, "provider.github.backoff.base"},
		{"base negative", map[string]any{"provider.github.backoff.base": "-1s"}, "provider.github.backoff.base"},
		{"max 0s", map[string]any{"provider.github.backoff.max": "0s", "provider.github.backoff.base": "0s"}, "provider.github.backoff.base"},
		{"max below base", map[string]any{"provider.github.backoff.base": "2s", "provider.github.backoff.max": "1s", "provider.github.backoff.min_secondary_wait": "500ms"}, "provider.github.backoff.max"},
		{"max equals base", map[string]any{"provider.github.backoff.base": "2s", "provider.github.backoff.max": "2s", "provider.github.backoff.min_secondary_wait": "500ms"}, ""},
		{"max below min_secondary_wait", map[string]any{"provider.github.backoff.base": "1s", "provider.github.backoff.max": "30s", "provider.github.backoff.min_secondary_wait": "60s"}, "provider.github.backoff.max"},
		{"max equals min_secondary_wait", map[string]any{"provider.github.backoff.base": "1s", "provider.github.backoff.max": "60s", "provider.github.backoff.min_secondary_wait": "60s"}, ""},
		{"max_attempts 0", map[string]any{"provider.github.backoff.max_attempts": 0}, "provider.github.backoff.max_attempts"},
		{"max_attempts 1", map[string]any{"provider.github.backoff.max_attempts": 1}, ""},
		{"cursor max_doc_bytes 0 rejected", map[string]any{"cursor.max_doc_bytes": 0}, "cursor.max_doc_bytes"},
		{"cursor max_doc_bytes 1 accepted", map[string]any{"cursor.max_doc_bytes": 1}, ""},
		{"intake list_limit 0", map[string]any{"intake.list_limit": 0}, "intake.list_limit"},
		{"intake list_limit -1", map[string]any{"intake.list_limit": -1}, "intake.list_limit"},
		{"intake list_limit 1", map[string]any{"intake.list_limit": 1}, ""},
		{"monitor pr_list_limit 0", map[string]any{"monitor.pr_list_limit": 0}, "monitor.pr_list_limit"},
		{"monitor pr_list_limit -1", map[string]any{"monitor.pr_list_limit": -1}, "monitor.pr_list_limit"},
		{"monitor pr_list_limit 1", map[string]any{"monitor.pr_list_limit": 1}, ""},
		{"cursor max_doc_bytes -1", map[string]any{"cursor.max_doc_bytes": -1}, "cursor.max_doc_bytes"},
		{"cursor max_chunks 0", map[string]any{"cursor.max_chunks": 0}, ""},
		{"cursor max_chunks -1", map[string]any{"cursor.max_chunks": -1}, "cursor.max_chunks"},
		{"webhook max_body 0", map[string]any{"webhook.max_body": 0}, ""},
		{"webhook max_body -1", map[string]any{"webhook.max_body": -1}, "webhook.max_body"},
		{"webhook provider empty", map[string]any{"webhook.provider": ""}, "webhook.provider"},
		{"claim_store dir empty", map[string]any{"claim_store.dir": ""}, "claim_store.dir"},
		{"renotify_after 0s", map[string]any{"notify.renotify_after": "0s"}, "notify.renotify_after"},
		{"renotify_after negative", map[string]any{"notify.renotify_after": "-1s"}, "notify.renotify_after"},
		{"renotify_after 1ns", map[string]any{"notify.renotify_after": "1ns"}, ""},
		{"max_retries 0", map[string]any{"wait.max_retries": 0}, "wait.max_retries"},
		{"max_retries 1", map[string]any{"wait.max_retries": 1}, ""},
		{"escalate_to empty", map[string]any{"wait.escalate_to": ""}, "wait.escalate_to"},
		{"pr_source no slash", map[string]any{"pr_source.repository": "single"}, "pr_source.repository"},
		{"pr_source wildcard", map[string]any{"pr_source.repository": "owner/*"}, "pr_source.repository"},
		{"pr_source url", map[string]any{"pr_source.repository": "https://x/y"}, "pr_source.repository"},
		{"pr_source empty", map[string]any{"pr_source.repository": ""}, "pr_source.repository"},
		{"report_resource empty allowed", map[string]any{"budget.report_resource": ""}, ""},
		{"reserve negative (presence only)", map[string]any{"budget.reserve": -3}, ""},
		{"summary_max_bytes negative (presence only)", map[string]any{"context.summary_max_bytes": -1}, ""},
		{"summary_max_bytes zero (presence only)", map[string]any{"context.summary_max_bytes": 0}, ""},
		// Poll range rules are NOT adopted by the SPEC: these must all load.
		{"poll min 0s accepted", map[string]any{"collector.poll.min": "0s"}, ""},
		{"poll max below min accepted", map[string]any{"collector.poll.min": "5m", "collector.poll.max": "1s"}, ""},
		{"poll safety 0s accepted", map[string]any{"collector.poll.safety": "0s"}, ""},
		{"poll negative accepted", map[string]any{"collector.poll.safety": "-1s"}, ""},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			d := afsValidDoc()
			for k, v := range tc.set {
				afsSet(d, k, v)
			}
			root := afsWriteDoc(t, d)
			if tc.key == "" {
				afsWantValid(t, root)
				return
			}
			afsWantInvalid(t, root, tc.key)
		})
	}
}

func TestAgenticFactorySettings_QueueLimits(t *testing.T) {
	// Only the webhook queue is table-tested here: it carries no minimum max_id_bytes.
	// The inbox has its own minimum and its own test below.
	for _, q := range []string{"queue.webhook"} {
		set := func(doc, seen, id any) afsDoc {
			d := afsValidDoc()
			afsSet(d, q+".max_doc_bytes", doc)
			afsSet(d, q+".seen_limit", seen)
			afsSet(d, q+".max_id_bytes", id)
			return d
		}
		// seen*(id+3)*2: 10*(7+3)*2 = 200.
		tests := []struct {
			name          string
			doc, seen, id int
			key           string
		}{
			{"equality accepted", 200, 10, 7, ""},
			{"one byte over rejected", 199, 10, 7, q + ".seen_limit"},
			// seen 1, id 4: 1*(4+3)*2 = 14 is the smallest doc cap that holds.
			{"small caps equality accepted", 14, 1, 4, ""},
			{"small caps one byte over rejected", 13, 1, 4, q + ".seen_limit"},
			{"doc 0 rejected", 0, 0, 0, q + ".max_doc_bytes"},
			{"doc 0 rejected even with a large seen_limit", 0, 1000000, 36, q + ".max_doc_bytes"},
			{"doc 1 accepted when seen 0", 1, 0, 0, ""},
			{"doc 1 rejected when seen 1", 1, 1, 1, q + ".seen_limit"},
			{"seen 0 id 0 accepted", 10, 0, 0, ""},
			{"seen positive id zero rejected", 100, 5, 0, q + ".max_id_bytes"},
			{"negative doc", -1, 0, 0, q + ".max_doc_bytes"},
			{"negative seen", 10, -1, 0, q + ".seen_limit"},
			{"negative id", 10, 0, -1, q + ".max_id_bytes"},
			{"overflow-sized limits rejected", 1 << 62, 1 << 62, 1 << 62, q + ".seen_limit"},
		}
		for _, tc := range tests {
			t.Run(q+"/"+tc.name, func(t *testing.T) {
				root := afsWriteDoc(t, set(tc.doc, tc.seen, tc.id))
				if tc.key == "" {
					afsWantValid(t, root)
					return
				}
				afsWantInvalid(t, root, tc.key)
			})
		}
	}
}

func TestAgenticFactorySettings_InboxMinIDBytes(t *testing.T) {
	const key = "notify.inbox.max_id_bytes"
	set := func(doc, seen, id int) string {
		d := afsValidDoc()
		afsSet(d, "notify.inbox.max_doc_bytes", doc)
		afsSet(d, "notify.inbox.seen_limit", seen)
		afsSet(d, key, id)
		return afsWriteDoc(t, d)
	}
	// seen*(id+3)*2: 1*(32+3)*2 = 70, 1*(33+3)*2 = 72.
	tests := []struct {
		name          string
		doc, seen, id int
		key           string
	}{
		{"id 0 rejected", 262144, 0, 0, key},
		{"id 0 rejected with seen_limit", 262144, 5, 0, key},
		{"id 1 rejected", 262144, 1, 1, key},
		{"id 16 rejected", 262144, 1, 16, key},
		{"id 16 rejected with seen 0", 262144, 0, 16, key},
		{"id 31 rejected", 262144, 1, 31, key},
		{"id 31 rejected with seen 0", 262144, 0, 31, key},
		{"id 32 accepted at equality", 70, 1, 32, ""},
		{"id 32 one byte over rejected", 69, 1, 32, "notify.inbox.seen_limit"},
		{"id 32 accepted with seen 0", 1, 0, 32, ""},
		{"id 33 accepted", 72, 1, 33, ""},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			root := set(tc.doc, tc.seen, tc.id)
			if tc.key == "" {
				afsWantValid(t, root)
				return
			}
			afsWantInvalid(t, root, tc.key)
		})
	}
	t.Run("error states the minimum and its reason", func(t *testing.T) {
		_, err := LoadAgenticFactorySettings(set(262144, 1, 31))
		if err == nil || !strings.Contains(err.Error(), "must be an integer >= 32 (agent inbox message ids are 32 bytes)") {
			t.Fatalf("error = %v", err)
		}
	})
	t.Run("webhook queue keeps no minimum", func(t *testing.T) {
		d := afsValidDoc()
		afsSet(d, "queue.webhook.max_doc_bytes", 262144)
		afsSet(d, "queue.webhook.seen_limit", 100)
		afsSet(d, "queue.webhook.max_id_bytes", 16)
		afsWantValid(t, afsWriteDoc(t, d))
	})
}

func TestAgenticFactorySettings_Routes(t *testing.T) {
	t.Run("empty list accepted", func(t *testing.T) {
		d := afsValidDoc()
		afsSet(d, "webhook.routes", []any{})
		s := afsWantValid(t, afsWriteDoc(t, d))
		if len(s.WebhookRoutes) != 0 {
			t.Fatalf("routes = %v", s.WebhookRoutes)
		}
	})
	t.Run("two routes keep order", func(t *testing.T) {
		d := afsValidDoc()
		afsSet(d, "webhook.routes", []any{afsDoc{"repository": "a/b"}, afsDoc{"repository": "c/d/e"}})
		s := afsWantValid(t, afsWriteDoc(t, d))
		if strings.Join(s.WebhookRoutes, ",") != "a/b,c/d/e" {
			t.Fatalf("routes = %v", s.WebhookRoutes)
		}
	})
	bad := map[string]any{
		"missing repository": afsDoc{"other": "x"},
		"null repository":    afsDoc{"repository": nil},
		"invalid repository": afsDoc{"repository": "no-slash"},
		"non-string":         afsDoc{"repository": 5},
		"entry not a map":    "a/b",
	}
	for name, entry := range bad {
		t.Run(name, func(t *testing.T) {
			d := afsValidDoc()
			afsSet(d, "webhook.routes", []any{afsDoc{"repository": "ok/ok"}, entry})
			root := afsWriteDoc(t, d)
			_, err := LoadAgenticFactorySettings(root)
			if err == nil || !strings.Contains(err.Error(), afsPrefix+"webhook.routes[1]") {
				t.Fatalf("err = %v, want it to name webhook.routes[1]", err)
			}
			if res := AgenticFactoryConfigured(root); res.Configured || res.Err == nil {
				t.Fatalf("result = %+v", res)
			}
		})
	}
}

func TestAgenticFactorySettings_Stage1NotConfigured(t *testing.T) {
	garbage := func() afsDoc {
		d := afsValidDoc()
		for _, k := range afsRequired {
			afsSet(d, k.key, "garbage")
		}
		return d
	}
	t.Run("no file", func(t *testing.T) {
		afsWantNotConfigured(t, t.TempDir())
	})
	t.Run("empty file", func(t *testing.T) {
		afsWantNotConfigured(t, afsWriteRaw(t, ""))
	})
	t.Run("no workflow block", func(t *testing.T) {
		afsWantNotConfigured(t, afsWriteRaw(t, "other: 1\n"))
	})
	t.Run("enabled absent with garbage elsewhere", func(t *testing.T) {
		d := garbage()
		afsDel(d, "enabled")
		afsWantNotConfigured(t, afsWriteDoc(t, d))
	})
	t.Run("enabled null with garbage elsewhere", func(t *testing.T) {
		d := garbage()
		afsSet(d, "enabled", nil)
		afsWantNotConfigured(t, afsWriteDoc(t, d))
	})
	t.Run("enabled false with garbage elsewhere", func(t *testing.T) {
		d := garbage()
		afsSet(d, "enabled", false)
		afsWantNotConfigured(t, afsWriteDoc(t, d))
	})
	t.Run("enabled false with only enabled", func(t *testing.T) {
		afsWantNotConfigured(t, afsWriteRaw(t, "workflow:\n  agentic_factory:\n    enabled: false\n"))
	})
	t.Run("enabled true only in workflow.yaml is ignored", func(t *testing.T) {
		root := t.TempDir()
		dir := filepath.Join(root, ".moai", "config", "sections")
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatal(err)
		}
		raw := "workflow:\n  agentic_factory:\n    enabled: true\n"
		if err := os.WriteFile(filepath.Join(dir, "workflow.yaml"), []byte(raw), 0o644); err != nil {
			t.Fatal(err)
		}
		afsWantNotConfigured(t, root)
	})
	for name, v := range map[string]any{"string": "true", "int": 1, "list": []any{true}, "map": afsDoc{"a": 1}} {
		t.Run("enabled non-bool "+name, func(t *testing.T) {
			d := afsValidDoc()
			afsSet(d, "enabled", v)
			root := afsWriteDoc(t, d)
			_, err := LoadAgenticFactorySettings(root)
			if err == nil || errors.Is(err, ErrAgenticFactoryNotConfigured) || !strings.Contains(err.Error(), `field "`+afsPrefix+`enabled"`) {
				t.Fatalf("err = %v", err)
			}
			if res := AgenticFactoryConfigured(root); res.Configured || res.Err == nil {
				t.Fatalf("result = %+v", res)
			}
		})
	}
	t.Run("workflow is not a map", func(t *testing.T) {
		root := afsWriteRaw(t, "workflow: 5\n")
		res := AgenticFactoryConfigured(root)
		if res.Configured || res.Err == nil {
			t.Fatalf("result = %+v", res)
		}
	})
}

func TestAgenticFactorySettings_ParseFailureIsError(t *testing.T) {
	root := afsWriteRaw(t, "workflow: [unclosed\n")
	_, err := LoadAgenticFactorySettings(root)
	if !errors.Is(err, ErrInvalidYAML) || errors.Is(err, ErrAgenticFactoryNotConfigured) {
		t.Fatalf("err = %v, want ErrInvalidYAML", err)
	}
	if res := AgenticFactoryConfigured(root); res.Configured || res.Err == nil || res.Reason == "" {
		t.Fatalf("result = %+v", res)
	}
}

func TestAgenticFactorySettings_DuplicateKeyRejected(t *testing.T) {
	root := afsWriteRaw(t, "workflow:\n  agentic_factory:\n    enabled: false\n    enabled: true\n")
	res := AgenticFactoryConfigured(root)
	if res.Configured || res.Err == nil {
		t.Fatalf("result = %+v", res)
	}
	// The duplicate itself must be the reported cause, not a later missing key.
	if !strings.Contains(res.Err.Error(), "duplicate key") || !strings.Contains(res.Err.Error(), afsPrefix+"enabled") {
		t.Fatalf("err = %v, want a duplicate-key error naming enabled", res.Err)
	}
}

func TestAgenticFactorySettings_NeverWrites(t *testing.T) {
	for name, d := range map[string]afsDoc{"valid": afsValidDoc(), "invalid": func() afsDoc { x := afsValidDoc(); afsSet(x, "labels.per_page", 0); return x }()} {
		t.Run(name, func(t *testing.T) {
			root := afsWriteDoc(t, d)
			dir := filepath.Join(root, ".moai", "config", "sections")
			file := filepath.Join(dir, "agentic-factory.yaml")
			before, _ := os.ReadFile(file)
			infoBefore, _ := os.Stat(file)
			_, _ = LoadAgenticFactorySettings(root)
			_ = AgenticFactoryConfigured(root)
			after, _ := os.ReadFile(file)
			infoAfter, _ := os.Stat(file)
			if string(before) != string(after) || !infoBefore.ModTime().Equal(infoAfter.ModTime()) {
				t.Fatal("file changed after reading")
			}
			entries, _ := os.ReadDir(dir)
			if len(entries) != 1 {
				t.Fatalf("directory has %d entries after reading, want 1", len(entries))
			}
		})
	}
}

func TestAgenticFactorySettings_SecretGuard(t *testing.T) {
	const marker = "not-a-real-value"
	cases := map[string]any{
		"direct":      afsDoc{"repository": "o/n", "secret": marker},
		"nested":      afsDoc{"repository": "o/n", "auth": afsDoc{"secret": marker}},
		"in a list":   afsDoc{"repository": "o/n", "auth": []any{afsDoc{"secret": marker}}},
		"empty value": afsDoc{"repository": "o/n", "secret": ""},
	}
	for name, entry := range cases {
		t.Run(name, func(t *testing.T) {
			d := afsValidDoc()
			afsSet(d, "webhook.routes", []any{afsDoc{"repository": "ok/ok"}, entry})
			root := afsWriteDoc(t, d)
			_, err := LoadAgenticFactorySettings(root)
			if err == nil {
				t.Fatal("expected an error for a secret key")
			}
			if strings.Contains(err.Error(), marker) {
				t.Fatalf("error echoes the secret value: %v", err)
			}
			if !strings.Contains(err.Error(), "webhook.routes[1].secret") || !strings.Contains(err.Error(), "reference") {
				t.Fatalf("error should name the key and tell the operator to use a reference: %v", err)
			}
			res := AgenticFactoryConfigured(root)
			if res.Configured || res.Err == nil || strings.Contains(res.Reason, marker) {
				t.Fatalf("result = %+v", res)
			}
		})
	}
	t.Run("a secret key outside routes is not this guard's concern", func(t *testing.T) {
		d := afsValidDoc()
		afsAF(d)["unknown"] = afsDoc{"secret": marker}
		afsWantValid(t, afsWriteDoc(t, d))
	})
	t.Run("guard does not apply when not configured", func(t *testing.T) {
		d := afsValidDoc()
		afsSet(d, "enabled", false)
		afsSet(d, "webhook.routes", []any{afsDoc{"secret": marker}})
		afsWantNotConfigured(t, afsWriteDoc(t, d))
	})
}

func TestAgenticFactorySettings_UnknownKeysIgnored(t *testing.T) {
	d := afsValidDoc()
	af := afsAF(d)
	af["future_key"] = afsDoc{"nested": []any{1, 2}}
	af["loop"] = afsDoc{"tick_period": "1s"}
	af["webhook"].(afsDoc)["routes"] = []any{afsDoc{"repository": "o/n", "secret_ref": "some-name"}}
	d["other_top_level"] = "x"
	s := afsWantValid(t, afsWriteDoc(t, d))
	if len(s.WebhookRoutes) != 1 {
		t.Fatalf("routes = %v", s.WebhookRoutes)
	}
}

//go:build afred

// screen_redaction_afred_test.go: B4-pre. KNOWN DEFECT REPRODUCTION, BEFORE THE
// REPAIR (RED). Excluded from the default test run by the afred build tag (the
// same convention as the lane-17 reproductions) so the default run stays green;
// run it with:
//
//	go test -tags afred -run '^TestScreenTextIsRefined$' ./internal/web/
//
// Contract: every text that reaches the Agentic Factory panels passes through
// one refiner. One value of each category — an e-mail address, a home
// directory path, a credential token — is put into every text path in turn;
// the rendered page and the JSON API must never carry the original. The test
// does not depend on which refiner masks it. When the core refiner lands, the
// build tag is dropped and this becomes an ordinary test.
//
// Sentinel on failure: AF_SCREEN_REFINE.
package web

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/ctxmodel"
)

var screenPayloads = map[string]string{
	"email":     "alice.private@example.com",
	"home-path": "/Users/alice/secret-project/notes",
	"token":     "ghp_" + strings.Repeat("Z", 36),
}

// screenCase puts one payload into one text path and returns the app to render.
type screenCase struct {
	name  string
	build func(t *testing.T, p string) *app
}

func monitorRowsApp(t *testing.T, rows string) *app { return prApp(t, rows, nil) }

func screenCases() []screenCase {
	return []screenCase{
		{"intake request: ready path", func(t *testing.T, p string) *app {
			a := newApp(Config{ProjectRoot: t.TempDir()})
			a.intakeRequests = func(context.Context) ([]byte, error) {
				return []byte(`[{"key":"k1","stage":"READY_ACKED","ready":{"host":"h","slot":"s","path":` + jsonQ(p) + `,"head":"abc"}}]`), nil
			}
			return a
		}},
		{"intake request: blocked detail", func(t *testing.T, p string) *app {
			a := newApp(Config{ProjectRoot: t.TempDir()})
			a.intakeRequests = func(context.Context) ([]byte, error) {
				return []byte(`[{"key":"k2","stage":"ACCEPTED","blocked":{"step":"s","detail":` + jsonQ(p) + `,"resume":` + jsonQ(p) + `}}]`), nil
			}
			return a
		}},
		{"intake request: branch, evidence, names", func(t *testing.T, p string) *app {
			a := newApp(Config{ProjectRoot: t.TempDir()})
			a.intakeRequests = func(context.Context) ([]byte, error) {
				return []byte(`[{"key":"k3","stage":"ACCEPTED","agent_name":` + jsonQ(p) + `,"slot":` + jsonQ(p) + `,"acked_by":` + jsonQ(p) +
					`,"branch":{"ref":` + jsonQ(p) + `},"evidence":[{"stage":"x","ref":` + jsonQ(p) + `}]}]`), nil
			}
			return a
		}},
		{"intake request: read error note", func(t *testing.T, p string) *app {
			a := newApp(Config{ProjectRoot: t.TempDir()})
			a.intakeRequests = func(context.Context) ([]byte, error) { return nil, errors.New("open " + p) }
			return a
		}},
		{"monitor rows: read error note", func(t *testing.T, p string) *app {
			return prApp(t, "", errors.New("open "+p))
		}},
		{"monitor rows: malformed note", func(t *testing.T, p string) *app {
			return prApp(t, "{"+p, nil)
		}},
		{"monitor rows: agent, state, stage, gate detail", func(t *testing.T, p string) *app {
			return monitorRowsApp(t, `{"rows":[{"Issue":1,"State":{"value":`+jsonQ(p)+`,"confirmed":true},"Stage":{"value":`+jsonQ(p)+`,"confirmed":true},`+
				`"agent":{"value":`+jsonQ(p)+`,"confirmed":true},"gate":{"value":{"reason":"r","detail":`+jsonQ(p)+`,"resume_condition":`+jsonQ(p)+`},"confirmed":true},`+
				`"wait":{"value":{"owner":`+jsonQ(p)+`,"next_action":`+jsonQ(p)+`,"diagnosis":`+jsonQ(p)+`,"escalate_to":`+jsonQ(p)+`},"confirmed":true}}]}`)
		}},
		{"monitor rows: context item owner and link", func(t *testing.T, p string) *app {
			return monitorRowsApp(t, `{"rows":[{"Issue":1,"context":{"value":[{"section":"spec","link":`+jsonQ(p)+`,"owner":`+jsonQ(p)+`,"target_head":`+jsonQ(p)+`,"hash":`+jsonQ(p)+`,"state":`+jsonQ(p)+`,"stage":`+jsonQ(p)+`}],"confirmed":true}}]}`)
		}},
		{"monitor intake item: block detail", func(t *testing.T, p string) *app {
			return monitorRowsApp(t, `{"rows":[],"intake":{"value":[{"key":"k","stage":"ACCEPTED","agent":`+jsonQ(p)+`,"slot":`+jsonQ(p)+`,"block":{"step":"s","detail":`+jsonQ(p)+`,"resume":`+jsonQ(p)+`}}],"confirmed":true}}`)
		}},
		{"context model: path, revision, verified, owner, title", func(t *testing.T, p string) *app {
			a := monitorRowsApp(t, cmRows)
			a.contextModel = func(context.Context, int) ([]byte, error) {
				l := &ctxmodel.Ledger{Issue: 7, Items: []ctxmodel.Item{{Section: ctxmodel.SectionSpec, Title: p, Owner: p, Path: p, Revision: p, Verified: p, VerifiedAt: p, State: ctxmodel.LinkState(p)}}}
				return l.Marshal()
			}
			return a
		}},
		{"context model: read error note", func(t *testing.T, p string) *app {
			a := monitorRowsApp(t, cmRows)
			a.contextModel = func(context.Context, int) ([]byte, error) { return nil, errors.New("open " + p) }
			return a
		}},
		{"browser: selection read error note", func(t *testing.T, p string) *app {
			a, _ := browserApp(t, "", errors.New("read "+p))
			return a
		}},
		{"browser: unrecognised saved value", func(t *testing.T, p string) *app {
			a, _ := browserApp(t, p, nil)
			return a
		}},
	}
}

func jsonQ(s string) string {
	b, _ := json.Marshal(s)
	return string(b)
}

func TestScreenTextIsRefined(t *testing.T) {
	for _, c := range screenCases() {
		for cat, payload := range screenPayloads {
			t.Run(c.name+"/"+cat, func(t *testing.T) {
				a := c.build(t, payload)
				m, err := a.buildMonitor(time.Now())
				if err != nil {
					t.Fatalf("AF_SCREEN_REFINE: buildMonitor: %v", err)
				}
				html := renderTempl(t, Monitor(tg1ShellVM(), m))
				if strings.Contains(html, payload) {
					t.Errorf("AF_SCREEN_REFINE: the %s reached the rendered page", cat)
				}
				for _, path := range []string{"/api/pr-tracking", "/api/intake-requests", "/api/browser"} {
					req := httptest.NewRequest(http.MethodGet, path, nil)
					req.Host = "127.0.0.1:8080"
					rec := httptest.NewRecorder()
					a.routes().ServeHTTP(rec, req)
					if strings.Contains(rec.Body.String(), payload) {
						t.Errorf("AF_SCREEN_REFINE: the %s reached %s", cat, path)
					}
				}
			})
		}
	}
}

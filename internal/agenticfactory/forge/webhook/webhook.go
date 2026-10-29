// Package webhook is the receiving end of the push channel (Issue #1948, the inbound side of forge): an http.Handler that turns a
// signed GitHub webhook delivery into one queued change, and answers the host only after the change is
// durably queued. It reads; it never writes to a forge. One route serves one repository, with its own
// secret and its own queue, so deliveries for different repositories cannot meet in a queue or a cursor. It only
// receives: it sends nothing, which is why it lives beside forge and not under effect.
package webhook

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/core"
	"github.com/modu-ai/moai-adk/internal/agenticfactory/effect/docstore"
	"github.com/modu-ai/moai-adk/internal/agenticfactory/forge"
)

// Acceptor queues one change and calls ack — the acknowledgement to the host — only after the change is
// durably stored. docstore.Queue is the production implementation.
type Acceptor interface {
	Accept(id string, ch forge.Change, ack func()) error
}

// Route is one repository's endpoint.
type Route struct {
	// Repo is "owner/name" as the host reports it in the payload's repository.full_name.
	Repo string
	// Secret returns the shared signing secret, read from configuration at the time of each delivery. An
	// empty secret refuses every delivery: an unsigned channel is never accepted.
	Secret func() []byte
	Queue  Acceptor
}

// Handler serves POST /hook/<owner>/<name>. Provider selects how repository names compare (see
// core.NormalizeRepo). MaxBody is the largest body read, from configuration; zero refuses every delivery.
type Handler struct {
	Provider string
	MaxBody  int64
	Routes   []Route
}

var _ http.Handler = (*Handler)(nil)

const (
	signatureHeader = "X-Hub-Signature-256"
	deliveryHeader  = "X-GitHub-Delivery"
	eventHeader     = "X-GitHub-Event"
	pathPrefix      = "/hook/"
)

// Sign is the signature GitHub sends for body under secret ("sha256=" and the hex HMAC-SHA256).
func Sign(secret, body []byte) string {
	m := hmac.New(sha256.New, secret)
	m.Write(body)
	return "sha256=" + hex.EncodeToString(m.Sum(nil))
}

func (h *Handler) route(path string) (Route, bool) {
	repo, ok := strings.CutPrefix(path, pathPrefix)
	if !ok || strings.Count(repo, "/") != 1 {
		return Route{}, false
	}
	want := core.NormalizeRepo(h.Provider, repo)
	for _, r := range h.Routes {
		if core.NormalizeRepo(h.Provider, r.Repo) == want {
			return r, true
		}
	}
	return Route{}, false
}

type payload struct {
	Action string `json:"action"`
	Issue  *struct {
		Number int `json:"number"`
	} `json:"issue"`
	PullRequest *struct {
		Number int `json:"number"`
	} `json:"pull_request"`
	Repository struct {
		FullName string `json:"full_name"`
	} `json:"repository"`
}

func kindOf(action string) forge.ChangeKind {
	switch action {
	case "opened":
		return forge.ChangeAdded
	case "deleted":
		return forge.ChangeRemoved
	}
	return forge.ChangeModified
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", http.MethodPost)
		http.Error(w, "POST only", http.StatusMethodNotAllowed)
		return
	}
	rt, ok := h.route(r.URL.Path)
	if !ok {
		http.NotFound(w, r)
		return
	}
	if h.MaxBody <= 0 {
		http.Error(w, "the body limit is not configured", http.StatusServiceUnavailable)
		return
	}
	var secret []byte
	if rt.Secret != nil {
		secret = rt.Secret()
	}
	if len(secret) == 0 || rt.Queue == nil {
		http.Error(w, "this route is not configured", http.StatusInternalServerError)
		return
	}
	body, err := io.ReadAll(http.MaxBytesReader(w, r.Body, h.MaxBody))
	if err != nil {
		var tooBig *http.MaxBytesError
		if errors.As(err, &tooBig) {
			http.Error(w, "body too large", http.StatusRequestEntityTooLarge)
			return
		}
		http.Error(w, "unreadable body", http.StatusBadRequest)
		return
	}
	// The signature is checked before anything in the body is read.
	if !hmac.Equal([]byte(Sign(secret, body)), []byte(r.Header.Get(signatureHeader))) {
		http.Error(w, "bad signature", http.StatusUnauthorized)
		return
	}
	id := r.Header.Get(deliveryHeader)
	if id == "" {
		http.Error(w, "no delivery id", http.StatusBadRequest)
		return
	}
	event := r.Header.Get(eventHeader)
	if event == "ping" {
		w.WriteHeader(http.StatusOK)
		return
	}
	var p payload
	if err := json.Unmarshal(body, &p); err != nil {
		http.Error(w, "payload is not JSON", http.StatusBadRequest)
		return
	}
	// A correctly signed delivery for another repository must not enter this repository's queue.
	if core.NormalizeRepo(h.Provider, p.Repository.FullName) != core.NormalizeRepo(h.Provider, rt.Repo) {
		http.Error(w, "the payload is for another repository", http.StatusBadRequest)
		return
	}
	var number int
	switch {
	case event == "issues" && p.Issue != nil:
		number = p.Issue.Number
	case event == "pull_request" && p.PullRequest != nil:
		number = p.PullRequest.Number
	default: // an event this channel does not carry: acknowledged so the host does not retry it
		w.WriteHeader(http.StatusAccepted)
		return
	}
	if number <= 0 {
		http.Error(w, "no record number", http.StatusBadRequest)
		return
	}
	acked := false
	err = rt.Queue.Accept(id, forge.Change{Key: strconv.Itoa(number), Kind: kindOf(p.Action)}, func() {
		acked = true
		w.WriteHeader(http.StatusAccepted)
	})
	switch {
	case err == nil || acked:
	case errors.Is(err, docstore.ErrQueueFull):
		http.Error(w, "queue full; redeliver later", http.StatusServiceUnavailable)
	case errors.Is(err, docstore.ErrBadDeliveryID), errors.Is(err, docstore.ErrNoDeliveryID):
		http.Error(w, "bad delivery id", http.StatusBadRequest)
	default:
		http.Error(w, "not queued; redeliver", http.StatusInternalServerError)
	}
}

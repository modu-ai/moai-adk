package ghexec

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/forge"
)

// Transport is a forge.Transport over `gh api -i`. It is read-only: a GET without a body is the only
// request it makes, and every call still passes through Exec's allowlist.
//
// gh follows a redirect and prints only the final answer, so Redirected is always false here. A moved
// repository is caught one level up instead: the identity reader takes the repository the response names
// (its repository_url), and the permission guard denies when that differs from the one asked about
// (core.Collaboration.repoAllowed).
type Transport struct{ Exec Exec }

var _ forge.Transport = Transport{}

// Do sends one GET.
//
//   - A printed status line becomes the Response, whatever the status: the forge client classifies 2xx,
//     304, 404 and 5xx itself.
//   - 401, and 403 without rate-limit evidence, are permission answers: forge.ErrDenied.
//   - 429, and a 403 that carries a retry-after, a zero remaining count or the secondary-limit message,
//     are rate limits: the response goes through with its headers unchanged so the client can wait.
//   - Output that is not a response and a failure with no HTTP answer are errors, never a guess.
func (t Transport) Do(ctx context.Context, r forge.Request) (forge.Response, error) {
	if m := strings.ToUpper(r.Method); (m != "" && m != "GET") || len(r.Body) > 0 {
		return forge.Response{}, fmt.Errorf("%w: gh api transport is read-only (%s with %d body bytes)", ErrNotAllowed, r.Method, len(r.Body))
	}
	args := []string{"api", "-i", r.Path}
	keys := make([]string, 0, len(r.Header))
	for k := range r.Header {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		v := r.Header[k]
		if strings.ContainsAny(k+v, "\r\n\x00") || strings.Contains(k, ":") {
			return forge.Response{}, fmt.Errorf("%w: header %q is not a clean header", ErrNotAllowed, k)
		}
		args = append(args, "-H", k+": "+v)
	}
	out, err := t.Exec.Run(ctx, args...)
	res, perr := parseResponse(out)
	switch {
	case perr == nil:
		// A printed answer is the answer, even when gh also exited non-zero for its status.
	case err != nil:
		var se *forge.StatusError
		if errors.As(err, &se) { // no headers were printed, but the exit text named the status
			res, perr = forge.Response{Status: se.Status}, nil
		} else {
			return forge.Response{}, err
		}
	default:
		return forge.Response{}, perr
	}
	switch {
	case res.Status == 429, res.Status == 403 && rateLimitEvidence(res):
		return res, nil
	case res.Status == 401, res.Status == 403:
		return forge.Response{}, fmt.Errorf("%w: http %d", forge.ErrDenied, res.Status)
	}
	return res, nil
}

// rateLimitEvidence reports a 403 that is a limit rather than a permission answer.
func rateLimitEvidence(r forge.Response) bool {
	if r.Header["retry-after"] != "" || strings.TrimSpace(r.Header["x-ratelimit-remaining"]) == "0" {
		return true
	}
	return bytes.Contains(bytes.ToLower(r.Body), []byte("rate limit"))
}

// parseResponse reads the `gh api -i` preamble: a status line, header lines, a blank line, the body.
func parseResponse(out []byte) (forge.Response, error) {
	head, body, found := splitHead(out)
	if !found {
		head, body = out, nil
	}
	lines := strings.Split(strings.ReplaceAll(string(head), "\r\n", "\n"), "\n")
	if len(lines) == 0 || !strings.HasPrefix(lines[0], "HTTP/") {
		return forge.Response{}, errors.New("ghexec: gh api printed no HTTP status line")
	}
	fields := strings.Fields(lines[0])
	if len(fields) < 2 {
		return forge.Response{}, fmt.Errorf("ghexec: unreadable status line %q", lines[0])
	}
	status, err := strconv.Atoi(fields[1])
	if err != nil || status < 100 || status > 599 {
		return forge.Response{}, fmt.Errorf("ghexec: unreadable status code in %q", lines[0])
	}
	res := forge.Response{Status: status, Header: map[string]string{}, Body: body}
	for _, l := range lines[1:] {
		name, val, ok := strings.Cut(l, ":")
		if !ok {
			continue
		}
		name = strings.ToLower(strings.TrimSpace(name))
		val = strings.TrimSpace(val)
		if prev, dup := res.Header[name]; dup {
			val = prev + ", " + val
		}
		res.Header[name] = val
	}
	return res, nil
}

// splitHead splits at the first blank line (CRLF or LF).
func splitHead(out []byte) (head, body []byte, found bool) {
	for _, sep := range []string{"\r\n\r\n", "\n\n"} {
		if i := bytes.Index(out, []byte(sep)); i >= 0 {
			return out[:i], out[i+len(sep):], true
		}
	}
	return nil, nil, false
}

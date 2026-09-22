package cli

import (
	"context"
	"fmt"
	"net"
	"net/url"
	"time"

	"github.com/modu-ai/moai-adk/internal/cli/uikit"
	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/jev"
	"github.com/modu-ai/moai-adk/internal/jevcred"
)

// jevCheckName is the `moai doctor --check "<name>"` address of the Jev
// readiness check (REQ-JEVC-022).
const jevCheckName = "Jev"

// jevProbeTimeout bounds the reachability probe. The doctor path is
// latency-sensitive, so the probe is time-boxed and a deadline overrun reports
// "unreachable" rather than delaying the run — an advisory check never blocks
// the path it runs on.
const jevProbeTimeout = 3 * time.Second

// jevEndpointProbe answers "is the endpoint reachable" WITHOUT asking the model
// anything (REQ-JEVC-022). It opens a TCP connection to the endpoint host and
// closes it: no HTTP request is written, so no judgment request can be sent
// even by accident. Replaced in tests.
var jevEndpointProbe = dialJevEndpoint

func dialJevEndpoint(ctx context.Context) error {
	u, err := url.Parse(jev.EndpointURL)
	if err != nil {
		return fmt.Errorf("endpoint URL is unparseable: %w", err)
	}
	host := u.Host
	if u.Port() == "" {
		port := "443"
		if u.Scheme == "http" {
			port = "80"
		}
		host = net.JoinHostPort(u.Hostname(), port)
	}
	d := net.Dialer{Timeout: jevProbeTimeout}
	conn, err := d.DialContext(ctx, "tcp", host)
	if err != nil {
		return err
	}
	return conn.Close()
}

// checkJev reports Jev readiness: whether the capability is enabled, whether a
// credential is present, and whether the endpoint is reachable — without
// sending a judgment request (REQ-JEVC-022).
//
// The check never fails the doctor run. Every absence here is graceful
// degradation by design (REQ-JEVC-007): a disabled capability is a normal
// state, and an absent credential or an unreachable endpoint means consumers
// get no signal, which each of them already handles. Reporting any of these as
// a failure would put a permanent red row in front of every user who never
// opted in.
//
// The gate is read FIRST and short-circuits: while the capability is disabled
// no network call is made at all (REQ-JEVC-017), so the probe is not reached
// even when a credential happens to be present.
func checkJev(root string, verbose bool) DiagnosticCheck {
	check := DiagnosticCheck{Name: jevCheckName, Status: uikit.CheckOK}

	enabled, readErr := jevEnabled(root)
	if readErr != nil {
		check.Message = "gate unreadable — treated as disabled; no network call made"
		if verbose {
			check.Detail = readErr.Error()
		}
		return check
	}
	if !enabled {
		check.Message = "disabled (workflow.jev.enabled: false) — no request constructed, no network call"
		if verbose {
			check.Detail = "set workflow.jev.enabled to true in .moai/config/sections/workflow.yaml to opt in; " +
				"a credential at ~/.moai/.env.typesafe is then also required"
		}
		return check
	}

	view := jevcred.View()
	credential := "no credential at ~/.moai/.env.typesafe"
	if view.Configured {
		credential = "credential configured"
		if view.Hint != "" {
			// The bounded disclosure, unchanged here: the final four
			// characters and never the credential (REQ-JEVC-020).
			credential = fmt.Sprintf("credential configured (…%s)", view.Hint)
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), jevProbeTimeout)
	defer cancel()
	probeErr := jevEndpointProbe(ctx)

	reach := "endpoint reachable"
	if probeErr != nil {
		reach = "endpoint unreachable"
	}

	check.Message = fmt.Sprintf("enabled; %s; %s", credential, reach)
	if !view.Configured || probeErr != nil {
		// Advisory, never a failure: the capability simply produces no signal.
		check.Status = uikit.CheckWarn
	}
	if verbose {
		detail := fmt.Sprintf("model %s; endpoint %s; readiness only — no judgment request is sent by this check",
			jev.ModelID, jev.EndpointURL)
		if probeErr != nil {
			detail += "; probe: " + probeErr.Error()
		}
		check.Detail = detail
	}
	return check
}

// jevEnabled resolves workflow.jev.enabled for the project at root. A project
// whose config cannot be read is reported as disabled with the read error
// surfaced — never as an error that stops the doctor run.
func jevEnabled(root string) (bool, error) {
	cfg, err := config.NewConfigManager().Load(root)
	if err != nil {
		return false, err
	}
	if cfg == nil {
		return false, fmt.Errorf("config resolved to nil for %s", root)
	}
	return cfg.Workflow.Jev.Enabled, nil
}

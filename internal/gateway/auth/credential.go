// Package auth defines provider-neutral references resolved at send time.
package auth

import (
	"errors"
	"net/http"
)

type ProviderID string

const (
	ProviderAnthropic ProviderID = "anthropic"
	ProviderOpenAI    ProviderID = "openai"
	ProviderZAI       ProviderID = "zai"
)

var ErrCredentialAbsent = errors.New("gateway credential absent")
var ErrCredentialChanged = errors.New("gateway credential generation changed")
var ErrWrongProvider = errors.New("gateway credential provider mismatch")

// CredentialRef does not expose credential bytes. Implementations must reject
// foreign-provider destinations in Apply. Senders must recheck Generation before
// sending; a changed generation invalidates a previously resolved request.
type CredentialRef interface {
	Provider() ProviderID
	Generation() (uint64, error)
	Apply(req *http.Request) error
	Redacted() string
}

// Absent explicitly rejects unimplemented or unauthenticated provider routes.
// It performs no network activity and never implements OAuth passthrough.
type Absent struct{ ProviderID ProviderID }

func (a Absent) Provider() ProviderID        { return a.ProviderID }
func (a Absent) Generation() (uint64, error) { return 0, ErrCredentialAbsent }
func (a Absent) Apply(*http.Request) error   { return ErrCredentialAbsent }
func (a Absent) Redacted() string            { return string(a.ProviderID) + ":absent" }

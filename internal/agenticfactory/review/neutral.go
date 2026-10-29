package review

// ProviderOperator marks a review the operator recorded through the
// ControlService evidence path (design §11.1 (b)).
const ProviderOperator = "operator"

// AdmitNeutral decides whether a code or security review record satisfies the
// merge gate without requiring Codex (design §11.1). It adds to Admit: the
// provider must be allowed, the record must be a code or security review, an
// operator record must carry the operator role, and the producer must differ
// from every writer session (current and earlier attempts).
func (r Record) AdmitNeutral(candidateHead string, writerSessions []string, codexOptedIn bool) Reason {
	switch r.Provider {
	case ProviderGLM:
		return ReasonProducer
	case ProviderCodex:
		if !codexOptedIn {
			return ReasonProducer
		}
	case ProviderOperator:
		if r.Producer.Role != "operator" {
			return ReasonProducer
		}
	}
	if r.Kind != KindCodeReview && r.Kind != KindSecurityReview {
		return ReasonProducer
	}
	if len(writerSessions) == 0 {
		return r.Admit(candidateHead, "")
	}
	for _, w := range writerSessions {
		if reason := r.Admit(candidateHead, w); reason != ReasonNone {
			return reason
		}
	}
	return ReasonNone
}

// SeparateReviewers reports whether the code review and the security review
// came from different sessions; one session holds one role in a run.
func SeparateReviewers(code, security Record) bool {
	return code.Producer.Session != "" && code.Producer.Session != security.Producer.Session
}

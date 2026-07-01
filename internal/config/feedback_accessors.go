package config

// feedback_accessors.go — resolver for the /moai feedback target repository.
//
// SPEC-INVOCATION-MODEL-001: the feedback workflow targets a configurable
// GitHub repository (the remote MoAI-ADK tool repo by default, overridable by
// fork maintainers via .moai/config/sections/feedback.yaml).

// FeedbackRepository returns the resolved feedback target repository slug.
//
// Resolution: the loaded Feedback.Repository value when non-empty; otherwise the
// default tool feedback channel (DefaultFeedbackRepository). The empty-fallback
// covers EC-1 (a feedback.yaml with the feedback: section present but the
// repository: key missing or explicitly empty).
func (c *Config) FeedbackRepository() string {
	if c.Feedback.Repository == "" {
		return DefaultFeedbackRepository
	}
	return c.Feedback.Repository
}

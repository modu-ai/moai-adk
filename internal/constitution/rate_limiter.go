package constitution

import (
	"os"
	"time"
)

const (
	// rateLimitMaxPerWeek is the maximum number of amendments in 7 days.
	rateLimitMaxPerWeek = 3
	// rateLimitCooldownHours is the minimum interval between amendments (hours).
	rateLimitCooldownHours = 24
	// rateLimitMaxActiveLearnings is the maximum number of active learnings.
	rateLimitMaxActiveLearnings = 50
	// rateLimitRollbackCooldownDays is the re-amendment cooldown for rolled-back rules (days).
	rateLimitRollbackCooldownDays = 30
)

// rateLimiter is the implementation of the RateLimiter interface.
// Limits amendment frequency.
type rateLimiter struct {
	// now is the current time function (injected for testability).
	now func() time.Time
}

// NewRateLimiter creates a RateLimiter.
func NewRateLimiter() RateLimiter {
	return &rateLimiter{
		now: time.Now,
	}
}

// Admit checks if the proposal is within rate limits.
// Implements SPEC-V3R2-CON-002 REQ-CON-002-007 Layer 4.
//
// Rate limiter rules:
// 1. Maximum 3 amendments in 7 days
// 2. 24-hour cooldown between amendments
// 3. Maximum 50 active learnings
// 4. Rolled-back rules have 30-day cooldown
func (l *rateLimiter) Admit(proposal *AmendmentProposal, evolutionLogPath string) error {
	logs, err := LoadEvolutionLogs(evolutionLogPath)
	if err != nil {
		// No file → first amendment → allow
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	now := l.now()

	// 1. Rollback check: verify if the rule was rolled back in the last 30 days
	for _, log := range logs {
		if log.RuleID == proposal.RuleID && log.RolledBack {
			if log.RollbackAt != nil {
				rollbackTime := log.RollbackAt
				cooldownEnd := rollbackTime.AddDate(0, 0, rateLimitRollbackCooldownDays)
				if now.Before(cooldownEnd) {
					return &ErrRolledBack{
						RuleID:       log.RuleID,
						RolledBackAt: *rollbackTime,
						CooldownDays: rateLimitRollbackCooldownDays,
					}
				}
			}
		}
	}

	// 2. 7-day window check: verify amendment count in last 7 days
	oneWeekAgo := now.AddDate(0, 0, -7)
	recentCount := 0
	var lastAmendmentTime time.Time

	for _, log := range logs {
		if log.ApprovedAt.After(oneWeekAgo) {
			recentCount++
			if log.ApprovedAt.After(lastAmendmentTime) {
				lastAmendmentTime = log.ApprovedAt
			}
		}
	}

	if recentCount >= rateLimitMaxPerWeek {
		return &ErrRateLimitExceeded{
			MaxPerWeek:    rateLimitMaxPerWeek,
			CooldownHours: rateLimitCooldownHours,
			NextAllowedAt: lastAmendmentTime.AddDate(0, 0, 7).Add(time.Duration(rateLimitCooldownHours) * time.Hour),
		}
	}

	// 3. Cooldown check: verify 24 hours have passed since last amendment
	if !lastAmendmentTime.IsZero() {
		cooldownEnd := lastAmendmentTime.Add(time.Duration(rateLimitCooldownHours) * time.Hour)
		if now.Before(cooldownEnd) {
			return &ErrRateLimitExceeded{
				MaxPerWeek:    rateLimitMaxPerWeek,
				CooldownHours: rateLimitCooldownHours,
				NextAllowedAt: cooldownEnd,
			}
		}
	}

	// 4. Active learnings cap: check count of entries with rolled_back=false
	activeCount := 0
	for _, log := range logs {
		if !log.RolledBack {
			activeCount++
		}
	}

	if activeCount >= rateLimitMaxActiveLearnings {
		return &ErrRateLimitExceeded{
			MaxPerWeek:    rateLimitMaxPerWeek,
			CooldownHours: rateLimitCooldownHours,
			NextAllowedAt: now.Add(time.Hour * 24), // Arbitrary future time
		}
	}

	return nil
}

// rateLimiter satisfies the RateLimiter interface.
var _ RateLimiter = (*rateLimiter)(nil)


package constitution

import (
	"os"
	"time"
)

const (
	// rateLimitMaxPerWeek는 7일 간 최대 amendment 횟수이다.
	rateLimitMaxPerWeek = 3
	// rateLimitCooldownHours는 amendment 간 최소 간격(시간)이다.
	rateLimitCooldownHours = 24
	// rateLimitMaxActiveLearnings는 최대 활성 learning 수이다.
	rateLimitMaxActiveLearnings = 50
	// rateLimitRollbackCooldownDays는 rollback된 rule의 재-amendment 쿨다운(일)이다.
	rateLimitRollbackCooldownDays = 30
)

// rateLimiter는 RateLimiter interface의 구현이다.
// Amendment 빈도를 제한한다.
type rateLimiter struct {
	// now는 현재 시간 함수이다 (테스트 가능성을 위해 주입).
	now func() time.Time
}

// NewRateLimiter는 RateLimiter를 생성한다.
func NewRateLimiter() RateLimiter {
	return &rateLimiter{
		now: time.Now,
	}
}

// Admit는 proposal이 rate limit 내에 있는지 확인한다.
// SPEC-V3R2-CON-002 REQ-CON-002-007 Layer 4 구현.
//
// Rate limiter 규칙:
// 1. 7일 간 최대 3회 amendment
// 2. Amendment 간 24시간 쿨다운
// 3. 최대 50개 활성 learnings
// 4. Rollback된 rule은 30일 쿨다운
func (l *rateLimiter) Admit(proposal *AmendmentProposal, evolutionLogPath string) error {
	logs, err := LoadEvolutionLogs(evolutionLogPath)
	if err != nil {
		// 파일 없음 → 첫 amendment → 허용
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	now := l.now()

	// 1. Rollback check: 해당 rule이 최근 30일 내에 rollback되었는지 확인
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

	// 2. 7-day window check: 최근 7일 내 amendment 횟수 확인
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

	// 3. Cooldown check: 마지막 amendment로부터 24시간 경과 확인
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

	// 4. Active learnings cap: rolled_back=false인 엔트리 수 확인
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
			NextAllowedAt: now.Add(time.Hour * 24), // 임의의 미래 시간
		}
	}

	return nil
}

// rateLimiter는 RateLimiter interface를 만족한다.
var _ RateLimiter = (*rateLimiter)(nil)


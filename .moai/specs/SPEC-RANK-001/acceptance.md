# SPEC-RANK-001: Acceptance Criteria

## Test Scenarios

### TS-001: Token Collection Hook

#### Scenario 1.1: Successful Session Submission
```gherkin
Given a user has registered with moai-adk rank register
And the user is working in a MoAI project (contains .moai folder)
When the Claude Code session ends
Then the session hook should execute
And token usage data should be extracted from .jsonl logs
And an HMAC-signed request should be sent to /api/v1/sessions
And the server should return 200 OK
And the session should appear in the user's dashboard
```

#### Scenario 1.2: Non-MoAI Project Ignored
```gherkin
Given a user has registered with moai-adk rank register
And the user is working in a non-MoAI project (no .moai folder)
When the Claude Code session ends
Then the session hook should NOT collect any data
And no API request should be made
```

#### Scenario 1.3: Network Failure Retry
```gherkin
Given a user has a valid session to submit
And the network is temporarily unavailable
When the session hook attempts to submit
Then the submission should be queued locally
And the system should retry up to 3 times with exponential backoff
And when network is restored, queued sessions should be submitted
```

---

### TS-002: User Registration

#### Scenario 2.1: Fresh Registration
```gherkin
Given a user has NOT previously registered
When the user runs "moai-adk rank register"
Then the system should open a browser for GitHub OAuth
And after successful authentication
Then an API key should be generated (256-bit entropy)
And the API key should be stored in OS secure storage
And the user should see a confirmation message with profile link
```

#### Scenario 2.2: Already Registered
```gherkin
Given a user has previously registered
When the user runs "moai-adk rank register"
Then the system should display current registration status
And offer option to re-authenticate or regenerate API key
```

---

### TS-003: Leaderboard Display

#### Scenario 3.1: Public Leaderboard View
```gherkin
Given any user visits rank.mo.ai.kr
When the page loads
Then the leaderboard should display within 2 seconds
And show rankings with: rank, username, avatar, composite score, total tokens, sessions
And period filter should default to "all-time"
And users with privacy_mode should show as "Anonymous"
```

#### Scenario 3.2: Authenticated User Highlight
```gherkin
Given an authenticated user visits the leaderboard
When the page loads
Then the user's own row should be visually highlighted
And if not in current view, a "Jump to my rank" button should appear
```

#### Scenario 3.3: Period Filtering
```gherkin
Given a user is viewing the leaderboard
When the user selects "weekly" filter
Then the leaderboard should update to show this week's rankings
And scores should reflect only this week's activity
```

---

### TS-004: User Dashboard

#### Scenario 4.1: Dashboard Overview
```gherkin
Given an authenticated user visits /dashboard
When the page loads
Then the user should see:
  - Current rank and percentile
  - Total tokens (all time)
  - This week's tokens
  - Token usage chart (last 30 days)
  - Recent sessions list
```

#### Scenario 4.2: API Key Regeneration
```gherkin
Given an authenticated user is on the dashboard
When the user clicks "Regenerate API Key"
Then a confirmation dialog should appear
And after confirmation:
  - A new API key should be generated
  - The old key should be immediately invalidated
  - The new key should be displayed (one-time view)
  - The event should be logged in audit log
```

#### Scenario 4.3: Privacy Mode Toggle
```gherkin
Given an authenticated user is on the dashboard
When the user enables privacy mode
Then their username should not appear on the public leaderboard
And their stats should still be counted in global totals
And they should see their own rank (not publicly visible)
```

---

### TS-005: Composite Score Calculation

#### Scenario 5.1: Score Formula Verification
```gherkin
Given a user with:
  - Total tokens: 1,000,000
  - Output/Input ratio: 1.5
  - Session count: 50
  - Streak days: 15
When the composite score is calculated
Then the score should be:
  - Token component: log10(1,000,001) / 10 * 0.40 = 0.240
  - Efficiency component: (1.5 / 2.0) * 0.25 = 0.1875
  - Session component: log10(51) / 3 * 0.20 = 0.114
  - Streak component: (15 / 30) * 0.15 = 0.075
  - Total: (0.240 + 0.1875 + 0.114 + 0.075) * 1000 = 616.5
```

#### Scenario 5.2: Daily Aggregation
```gherkin
Given sessions are submitted throughout the day
When midnight UTC occurs
Then daily aggregates should be calculated for all users
And rankings should be updated based on new aggregates
```

---

### TS-006: CLI Status Commands

#### Scenario 6.1: Rank Status
```gherkin
Given a registered user runs "moai-adk rank status"
When the command executes
Then the output should display:
  - Current rank: #X (top Y%)
  - Composite score: XXX
  - This week: X,XXX tokens across N sessions
  - Streak: X days
```

#### Scenario 6.2: CLI Leaderboard
```gherkin
Given a registered user runs "moai-adk rank leaderboard --limit 10"
When the command executes
Then a formatted table should display:
  - Top 10 users with rank, username, score
  - Current user's position if not in top 10
```

---

### TS-007: Security Scenarios

#### Scenario 7.1: Invalid HMAC Signature
```gherkin
Given an attacker attempts to submit session data
With an invalid or missing HMAC signature
When the request reaches /api/v1/sessions
Then the server should return 401 Unauthorized
And the attempt should be logged in security audit log
```

#### Scenario 7.2: Replay Attack Prevention
```gherkin
Given an attacker captures a valid signed request
And replays it after 5 minutes
When the request reaches the server
Then the server should reject it due to timestamp expiry
And return 401 Unauthorized
```

#### Scenario 7.3: Rate Limit Exceeded
```gherkin
Given a user has submitted 100 sessions in the past hour
When they attempt to submit another session
Then the server should return 429 Too Many Requests
And include Retry-After header
```

#### Scenario 7.4: API Key Compromise Response
```gherkin
Given a user suspects their API key is compromised
When they regenerate their API key from the dashboard
Then all previous API key hashes should be invalidated
And all pending sessions with old key should be rejected
And the user should receive the new key for CLI configuration
```

---

### TS-008: Privacy Scenarios

#### Scenario 8.1: No Code Content Collected
```gherkin
Given the session hook analyzes .jsonl logs
When extracting usage data
Then ONLY the following fields should be extracted:
  - input_tokens, output_tokens
  - cache_creation_input_tokens, cache_read_input_tokens
  - timestamp, model
Then NO code content, prompts, or responses should be accessed
```

#### Scenario 8.2: Project Path Anonymization
```gherkin
Given a session is submitted with project path
When the server processes it
Then the path should be hashed with HMAC(user_salt, path)
And only the first 16 characters of hash should be stored
And the original path should NEVER be logged or stored
```

---

## Checklist Summary

### Functional Requirements
- [ ] FR-001: Token Collection Hook
- [ ] FR-002: User Registration
- [ ] FR-003: Leaderboard Display
- [ ] FR-004: User Dashboard
- [ ] FR-005: Composite Score Calculation
- [ ] FR-006: CLI Status Commands

### Non-Functional Requirements
- [ ] NFR-001: Security (API key hash, HMAC, rate limiting)
- [ ] NFR-002: Performance (< 2s page load, < 500ms API)
- [ ] NFR-003: Privacy (no code collection, path hashing)
- [ ] NFR-004: Availability (99.9% uptime, offline CLI)

### Security Checklist
- [ ] API key 256-bit entropy
- [ ] Local secure storage (Keychain/libsecret)
- [ ] HMAC request signing
- [ ] Server-side hash recalculation
- [ ] Rate limiting active
- [ ] Audit logging implemented
- [ ] Security headers configured
- [ ] RLS policies in database

---

**Test Coverage Target**: 85%+
**Last Updated**: 2026-01-11

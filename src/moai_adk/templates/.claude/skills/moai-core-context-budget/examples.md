# Context Budget Optimization Examples

Practical strategies for managing context window tokens efficiently in LLM conversations.

---

## Progressive File Viewing

### View File Outline First

```python
# ❌ BAD: Load entire 5000-line file immediately
view_file("src/backend/main.py")  # ~25,000 tokens

# ✅ GOOD: Start with outline to understand structure
view_file_outline("src/backend/main.py")  # ~500 tokens

# Output shows:
# - Classes: UserController, AuthService, DatabaseManager
# - Functions: init_app(), setup_routes(), handle_error()
# - Lines: 1-5000

# Then view only the specific function you need
view_code_item("src/backend/main.py", ["UserController.create_user"])  # ~200 tokens
```

### Targeted Line Range Viewing

```python
# ❌ BAD: View entire file to see one function
view_file("utils/helpers.py")  # 10,000 tokens

# ✅ GOOD: View specific line range
# (After checking outline or search results)
view_file("utils/helpers.py", start_line=145, end_line=180)  # 140 tokens
```

**Token Savings**: ~99% reduction (10,000 → 140 tokens)

---

## Conversation Summarization

### Memory File Structure

```markdown
<!-- .moai/memory/session-summary.md -->

# Session Summary (Last Updated: 2025-11-20 14:48)

## Current Task

Implementing user authentication with JWT tokens

## Key Decisions

- Using bcrypt for password hashing (cost factor: 12)
- JWT expiry: 7 days for access tokens, 30 days for refresh tokens
- Storing refresh tokens in Redis with user ID as key

## Files Modified

- `src/auth/jwt.ts` - JWT generation and verification
- `src/auth/password.ts` - Password hashing utilities
- `src/middleware/auth.ts` - Authentication middleware

## Next Steps

1. Implement token refresh endpoint
2. Add unit tests for auth middleware
3. Update API documentation

## Gotchas

- Remember to set `httpOnly` flag on JWT cookies
- Use constant-time comparison for token validation
```

**Token Usage**:

- Full conversation history: ~50,000 tokens
- Summarized memory file: ~500 tokens
- **Savings**: 99% reduction

### Archive Old Sessions

```bash
# Move completed session summaries to archive
.moai/memory/
├── session-summary.md           # Current session (~500 tokens)
├── architecture.md              # System design (~1,000 tokens)
├── gotchas.md                   # Common pitfalls (~300 tokens)
└── archive/                     # Closed sessions (not loaded)
    ├── feature-auth-20251115.md
    ├── refactor-db-20251110.md
    └── bugfix-cors-20251105.md
```

**Strategy**: Only load current session + reference docs (~2,000 tokens total)

---

## JIT (Just-In-Time) File Loading

### Load Files On-Demand

```python
# Task: Add email verification to signup flow

# ❌ BAD: Load all potentially related files upfront
view_file("src/models/user.ts")          # 2,000 tokens
view_file("src/services/email.ts")       # 3,000 tokens
view_file("src/controllers/auth.ts")     # 5,000 tokens
view_file("src/utils/templates.ts")      # 4,000 tokens
view_file("config/email.json")           # 500 tokens
# Total: 14,500 tokens (before starting work!)

# ✅ GOOD: Load incrementally as needed

# 1. Start with outline to understand structure
view_file_outline("src/controllers/auth.ts")  # 200 tokens

# 2. View only the signup function
view_code_item("src/controllers/auth.ts", ["AuthController.signup"])  # 300 tokens

# 3. Now I see it calls EmailService.sendVerification
view_code_item("src/services/email.ts", ["EmailService.sendVerification"])  # 250 tokens

# 4. Finally, view the email template
view_file("templates/verification-email.html")  # 150 tokens

# Total: 900 tokens (94% savings!)
```

---

## Selective Code Sharing

### Minimize Code When Seeking Review

```typescript
// ❌ BAD: Share entire file with all imports, comments, and types
// file: src/auth/jwt.ts (2,500 tokens)

import * as jwt from 'jsonwebtoken';
import { config } from '../config';
import { User } from '../models/user';
import { logger } from '../utils/logger';
import { DatabaseService } from '../services/database';
import { CacheService } from '../services/cache';

/**
 * JWT Service for token generation and verification.
 * Supports both access tokens (short-lived) and refresh tokens (long-lived).
 *
 * Access tokens:
 * - Expire in 7 days
 * - Stored in httpOnly cookies
 * - Used for API authentication
 *
 * Refresh tokens:
 * - Expire in 30 days
 * - Stored in Redis
 * - Used to generate new access tokens
 */
export class JWTService {
  private secret: string;
  private accessTokenExpiry: string;
  private refreshTokenExpiry: string;

  constructor() {
    this.secret = config.jwt.secret;
    this.accessTokenExpiry = '7d';
    this.refreshTokenExpiry = '30d';
  }

  /**
   * Generates an access token for the given user ID.
   * @param userId - User's unique identifier
   * @returns Signed JWT token
   */
  generateAccessToken(userId: string): string {
    return jwt.sign(
      { userId, type: 'access' },
      this.secret,
      { expiresIn: this.accessTokenExpiry }
    );
  }

  // ... (30 more lines)
}


// ✅ GOOD: Share only the relevant function (~200 tokens)
// Asking: "Is this JWT generation logic secure?"

generateAccessToken(userId: string): string {
  return jwt.sign(
    { userId, type: 'access' },
    this.secret,
    { expiresIn: this.accessTokenExpiry }
  );
}

// Using: bcrypt cost 12, JWT expiry 7d
// Question: Should I add more claims (e.g., roles)?
```

**Token Savings**: 92% reduction (2,500 → 200 tokens)

---

## Search vs. View

### Use Search for Discovery

```python
# Task: Find all usages of deprecated function `oldAuth()`

# ❌ BAD: View all files that might contain it
view_file("src/controllers/user.ts")      # 3,000 tokens
view_file("src/controllers/admin.ts")     # 2,500 tokens
view_file("src/middleware/auth.ts")       # 1,500 tokens
view_file("src/services/auth.ts")         # 4,000 tokens
# Total: 11,000 tokens

# ✅ GOOD: Search first, then view only matches
grep_search(
  SearchPath="src/",
  Query="oldAuth",
  MatchPerLine=True
)
# Returns: 3 matches in 2 files (~100 tokens for search results)

# Then view only the matching line ranges:
# File: src/controllers/user.ts, Line 145
view_file("src/controllers/user.ts", start_line=140, end_line=150)  # 40 tokens

# File: src/middleware/auth.ts, Lines 78, 92
view_file("src/middleware/auth.ts", start_line=75, end_line=95)  # 80 tokens

# Total: 220 tokens (98% savings!)
```

---

## Batch File Operations

### Group Related Reads

```python
# Task: Understand authentication flow

# ❌ BAD: Scattered file views with explanations in between
view_file_outline("src/middleware/auth.ts")   # 200 tokens
# "OK I see authenticate() middleware..."
view_code_item("src/middleware/auth.ts", ["authenticate"])  # 300 tokens
# "Now let me check how JWT is verified..."
view_code_item("src/services/jwt.ts", ["verifyToken"])  # 250 tokens
# "And how the user is loaded..."
view_code_item("src/models/user.ts", ["User.findById"])  # 200 tokens

# Conversation overhead: ~400 tokens
# Total: 1,350 tokens


# ✅ GOOD: Batch all reads upfront, analyze together
# Step 1: Gather all relevant code
view_code_item("src/middleware/auth.ts", ["authenticate"])  # 300 tokens
view_code_item("src/services/jwt.ts", ["verifyToken"])      # 250 tokens
view_code_item("src/models/user.ts", ["User.findById"])     # 200 tokens

# Step 2: Analyze the complete flow (one conversation turn)
# "Now I have the full auth flow: middleware → JWT verify → load user"

# Total: 750 tokens (44% savings due to less back/forth)
```

---

## Low Budget Recovery

### When <20% Tokens Remaining (~40K tokens left)

#### Strategy 1: Summarize Conversation

```markdown
<!-- Prompt to Claude -->

Please summarize our conversation so far into a bullet-point memory file:

- Current task
- Key decisions made
- Files we've modified
- Next steps

Save to .moai/memory/session-summary.md
```

**Result**: Frees ~30,000 tokens from conversation history

#### Strategy 2: Close Unnecessary Files

```python
# If you've viewed many files but only need a few:

# Review what's in context:
# - src/models/* (10 files, ~15,000 tokens)
# - src/controllers/* (8 files, ~12,000 tokens)
# - config/* (5 files, ~3,000 tokens)

# Keep only what's needed for current task:
# "For the email verification task, I only need:
# - src/controllers/auth.ts
# - src/services/email.ts
# Rest can be re-loaded if needed"

# Freed: ~27,000 tokens
```

#### Strategy 3: Split Task

```markdown
<!-- Current task is too large -->

Task: "Implement complete authentication system with OAuth, 2FA, and passwordless login"

# This would require viewing 50+ files (~100,000 tokens)

<!-- Split into smaller tasks -->

Session 1: "Implement basic JWT authentication" (~20,000 tokens)
Session 2: "Add OAuth integration" (~25,000 tokens)
Session 3: "Implement 2FA with TOTP" (~20,000 tokens)
Session 4: "Add passwordless (magic link) login" (~15,000 tokens)

# Each session fits comfortably within budget
```

---

## Token Estimation

### Quick Token Calculator

```python
def estimate_tokens(text: str) -> int:
    """
    Rough token estimation.
    Rule of thumb: ~4 characters per token for English.
    """
    return len(text) // 4


# Examples
code = '''
def hello():
    print("Hello, world!")
'''
print(estimate_tokens(code))  # ~10 tokens

large_file = open("src/main.py").read()
print(estimate_tokens(large_file))  # e.g., ~15,000 tokens
```

### Token Budget Tracker

```python
class TokenBudgetTracker:
    """Track token usage during conversation."""

    def __init__(self, max_tokens=200_000):
        self.max_tokens = max_tokens
        self.used_tokens = 0

    def add_file(self, filepath: str, content: str):
        """Record tokens used by viewing a file."""
        tokens = estimate_tokens(content)
        self.used_tokens += tokens
        print(f"Added {filepath}: {tokens:,} tokens")

    def add_message(self, message: str):
        """Record tokens in a message."""
        tokens = estimate_tokens(message)
        self.used_tokens += tokens

    def remaining(self) -> int:
        """Get remaining tokens."""
        return self.max_tokens - self.used_tokens

    def percentage_used(self) -> float:
        """Get percentage of budget used."""
        return (self.used_tokens / self.max_tokens) * 100

    def status(self):
        """Print current status."""
        remaining = self.remaining()
        pct = self.percentage_used()

        print(f"\n{'='*50}")
        print(f"Token Budget Status")
        print(f"{'='*50}")
        print(f"Used:      {self.used_tokens:,} tokens ({pct:.1f}%)")
        print(f"Remaining: {remaining:,} tokens")

        if pct > 80:
            print("⚠️  WARNING: >80% budget used! Consider summarizing.")
        elif pct > 60:
            print("ℹ️  INFO: >60% budget used. Monitor usage.")
        else:
            print("✓ Budget healthy")
        print(f"{'='*50}\n")


# Usage
tracker = TokenBudgetTracker()

tracker.add_file("src/main.py", open("src/main.py").read())
tracker.add_file("src/utils.py", open("src/utils.py").read())
tracker.add_message("Implement user authentication with JWT")

tracker.status()
# Output:
# ==================================================
# Token Budget Status
# ==================================================
# Used:      15,234 tokens (7.6%)
# Remaining: 184,766 tokens
# ✓ Budget healthy
# ==================================================
```

---

## Real-World Scenarios

### Scenario 1: Large Codebase Exploration

**Task**: Understand authentication flow in a 100-file codebase

**Approach**:

1. Search for "auth" → Find 15 relevant files (100 tokens)
2. View outlines of top 5 files (500 tokens)
3. View specific auth functions (1,500 tokens)
4. Summarize findings in memory file (200 tokens)

**Total**: ~2,300 tokens (vs. 150,000 if viewing all files)

### Scenario 2: Code Review

**Task**: Review 10 PRs with 2,000 lines of changes total

**Approach**:

1. Use `git diff` to see changes only (not full files) (~5,000 tokens)
2. Review each PR's diff individually
3. Summarize review comments in a document (~500 tokens)

**Total**: ~5,500 tokens (vs. 50,000 viewing all full files)

### Scenario 3: Multi-Day Project

**Day 1**: Implement feature, use 120K tokens

- End of day: Summarize session → 500 tokens
- Archive conversation history

**Day 2**: Continue feature, start fresh

- Load summary from Day 1 → 500 tokens
- Have 199.5K tokens available for new work

**Result**: Effectively "reset" context budget while maintaining continuity

---

**See also**: [reference.md](./reference.md) for advanced token optimization techniques and monitoring strategies

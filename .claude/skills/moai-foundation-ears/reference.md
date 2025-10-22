# EARS Reference Documentation

> Official EARS methodology created by Alistair Mavin and Rolls-Royce PLC team (2009)

---

## What is EARS?

The **Easy Approach to Requirements Syntax (EARS)** is a lightweight mechanism to structure natural language requirements using consistent patterns. It reduces ambiguity, improves readability, and produces testable requirements without requiring specialized tools.

**Key Benefits**:
- ✅ Reduces ambiguity in natural language
- ✅ Easy to learn (minimal training overhead)
- ✅ No specialized tools required
- ✅ Effective for non-native English speakers
- ✅ Produces testable, verifiable requirements

**Industry Adoption**: Airbus, Bosch, Dyson, Honeywell, Intel, NASA, Rolls-Royce, Siemens, and universities worldwide.

---

## Generic EARS Syntax

All EARS requirements follow this foundational structure:

```
While <optional pre-condition>, when <optional trigger>, the <system name> shall <system response>
```

**Component Rules**:
- **Pre-conditions**: Zero or many (use "While" or "Where")
- **Trigger**: Zero or one (use "When" or "If")
- **System name**: Exactly one
- **System response**: One or many

---

## The Five EARS Patterns

### 1. Ubiquitous Requirements

**Always active with no keywords** — describes fundamental properties or constraints.

**Format**:
```
The <system name> shall <system response>
```

**Examples**:
```
✅ The mobile phone shall have a mass of less than 150 grams.
✅ The API shall respond within 200 milliseconds.
✅ The application shall support UTF-8 character encoding.
```

**When to use**:
- Fundamental system properties (size, weight, capacity)
- Performance constraints (latency, throughput)
- Universal design requirements (accessibility, compatibility)

---

### 2. State-Driven Requirements

**Active while specified conditions remain true** — uses keyword **"While"**.

**Format**:
```
While <precondition(s)>, the <system name> shall <system response>
```

**Examples**:
```
✅ While there is no card in the ATM, the ATM shall display 'insert card to begin'.
✅ While the battery level is below 20%, the device shall enable power-saving mode.
✅ While the user is not authenticated, the dashboard shall display only public data.
```

**When to use**:
- Continuous conditions (temperature, battery level, connection status)
- User states (authenticated, anonymous, admin)
- System modes (idle, active, maintenance)

---

### 3. Event-Driven Requirements

**Specify system response to triggering events** — uses keyword **"When"**.

**Format**:
```
When <trigger>, the <system name> shall <system response>
```

**Examples**:
```
✅ When 'mute' is selected, the laptop shall suppress all audio output.
✅ When the user clicks 'Submit', the form shall validate all input fields.
✅ When the connection is lost, the application shall retry up to 3 times.
```

**When to use**:
- User actions (button clicks, form submissions)
- System events (startup, shutdown, reconnection)
- External triggers (API calls, scheduled jobs)

---

### 4. Optional Feature Requirements

**Apply only when specific features are included** — uses keyword **"Where"**.

**Format**:
```
Where <feature is included>, the <system name> shall <system response>
```

**Examples**:
```
✅ Where the car has a sunroof, the car shall have a sunroof control panel on the driver door.
✅ Where multi-language support is enabled, the application shall display a language selector.
✅ Where dark mode is available, the UI shall persist the user's theme preference.
```

**When to use**:
- Optional modules or plugins
- Feature flags or toggles
- Conditional functionality based on licensing or configuration

---

### 5. Unwanted Behavior Requirements

**Specify responses to undesired situations** — uses keywords **"If"** and **"Then"**.

**Format**:
```
If <trigger>, then the <system name> shall <system response>
```

**Examples**:
```
✅ If an invalid credit card number is entered, then the website shall display 'please re-enter credit card details'.
✅ If authentication fails 3 times, then the system shall lock the account for 15 minutes.
✅ If the API rate limit is exceeded, then the server shall return HTTP 429 status code.
```

**When to use**:
- Error handling (invalid input, network failures)
- Security responses (failed authentication, unauthorized access)
- Constraint violations (rate limits, resource exhaustion)

---

## Complex Requirements (Combined Patterns)

Simple patterns can be combined to express richer behaviors.

**Format**:
```
While <precondition(s)>, When <trigger>, the <system name> shall <system response>
```

**Examples**:
```
✅ While the aircraft is on ground, when reverse thrust is commanded, the engine control system shall enable reverse thrust.

✅ While the user is authenticated, when the session expires, the application shall redirect to the login page.

✅ Where multi-factor authentication is enabled, when the user logs in, the system shall send a verification code to the registered email.
```

**Pattern Combination Rules**:
- Start with preconditions ("While" or "Where")
- Add trigger ("When" or "If/Then")
- End with system response
- Maintain temporal logic (chronological order)

---

## EARS Best Practices

### Writing Guidelines

✅ **DO**:
- Use active voice ("the system shall...")
- Be specific and measurable
- Keep requirements atomic (one requirement per statement)
- Use consistent terminology throughout
- Include quantifiable criteria when possible

❌ **DON'T**:
- Use passive voice ("data shall be processed...")
- Write ambiguous or vague requirements
- Combine multiple requirements in one statement
- Use synonyms for the same concept
- Include implementation details (HOW) instead of behavior (WHAT)

### Common Pitfalls

| Anti-Pattern | Problem | Fix |
|-------------|---------|-----|
| "The system should..." | Non-binding language | Use "shall" (mandatory) |
| "The system may..." | Unclear optionality | Use "Where" pattern for optional features |
| Multiple "shall" statements | Atomic requirement violated | Split into separate requirements |
| Implementation specifics | Over-constrains design | Focus on behavior, not implementation |

### Clarity Checklist

Before finalizing an EARS requirement, verify:
- [ ] Pattern matches requirement type (Ubiquitous/State/Event/Optional/Unwanted)
- [ ] Keywords are correct ("While"/"When"/"Where"/"If-Then")
- [ ] System name is explicit and consistent
- [ ] Response is measurable or verifiable
- [ ] No ambiguous terms (e.g., "fast", "robust", "user-friendly")
- [ ] Terminology is consistent with other requirements

---

## Integration with MoAI-ADK

### SPEC File Structure

EARS requirements integrate into `.moai/specs/SPEC-<ID>/spec.md`:

```markdown
---
id: SPEC-AUTH-001
version: 0.1.0
status: draft
created: 2025-10-22
---

# @SPEC:AUTH-001 | User Authentication

## Requirements (EARS Format)

### Ubiquitous
- The authentication service shall support JWT tokens with HS256 signing.
- The API shall rate-limit login attempts to 5 per minute per IP address.

### Event-Driven
- When the user submits valid credentials, the system shall return a JWT token.
- When the JWT token expires, the API shall return HTTP 401 status.

### Unwanted Behavior
- If invalid credentials are provided, then the system shall return HTTP 401 with message "Invalid credentials".
- If the account is locked, then the system shall return HTTP 403 with message "Account locked".

## HISTORY

### v0.1.0 (2025-10-22)
- **INITIAL**: Draft JWT authentication SPEC with EARS requirements.
```

### TAG Integration

Link EARS requirements to code and tests:

```python
# @CODE:AUTH-001 | SPEC: SPEC-AUTH-001.md | TEST: tests/auth/test_jwt.py

def authenticate_user(username: str, password: str) -> str:
    """
    EARS Requirement: Event-Driven
    'When the user submits valid credentials, the system shall return a JWT token.'
    """
    # Implementation
    pass
```

---

## Quick Reference Card

| Pattern | Keyword | Use Case | Example Start |
|---------|---------|----------|---------------|
| **Ubiquitous** | None | Fundamental properties | "The system shall..." |
| **State-Driven** | While | Continuous conditions | "While <state>, the system shall..." |
| **Event-Driven** | When | Triggered actions | "When <event>, the system shall..." |
| **Optional** | Where | Feature-specific | "Where <feature>, the system shall..." |
| **Unwanted** | If/Then | Error handling | "If <error>, then the system shall..." |

---

## Resources

**Official Sources**:
- Alistair Mavin's EARS homepage: https://alistairmavin.com/ears/
- Original IEEE paper (2009): "Easy Approach to Requirements Syntax"
- QRA Corporation definitive guide: https://qracorp.com/guides_checklists/the-easy-approach-to-requirements-syntax-ears/

**Training Materials**:
- INCOSE Requirements Working Group presentations
- Jama Software requirements engineering guide
- Medium article by Oguz Senna (ParamTech)

**Adopting Organizations**:
Airbus, Bosch, Dyson, Honeywell, Intel, NASA, Rolls-Royce, Siemens

---

**Last Updated**: 2025-10-22
**Version**: 1.0.0
**Maintained by**: MoAI-ADK Foundation Team

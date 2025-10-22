# EARS Working Examples

> Real-world requirement authoring scenarios using EARS patterns

_Last updated: 2025-10-22_

---

## Example 1: RESTful API Authentication (Full SPEC)

### Context
Building a REST API with JWT-based authentication and rate limiting.

### EARS Requirements by Pattern

#### Ubiquitous Requirements
```
1. The authentication service shall support JWT tokens with HS256 signing algorithm.
2. The API shall enforce HTTPS for all authentication endpoints.
3. The system shall store passwords using bcrypt with cost factor 12.
4. The API shall rate-limit login attempts to 5 per minute per IP address.
```

#### Event-Driven Requirements
```
5. When the user submits valid credentials, the system shall return a JWT token with 24-hour expiration.
6. When the JWT token expires, the API shall return HTTP 401 status with message "Token expired".
7. When the user requests logout, the system shall invalidate the token server-side.
```

#### Unwanted Behavior Requirements
```
8. If invalid credentials are provided, then the system shall return HTTP 401 with message "Invalid credentials".
9. If the account is locked, then the system shall return HTTP 403 with message "Account locked. Contact support.".
10. If the JWT signature is invalid, then the API shall return HTTP 401 with message "Invalid token".
```

#### State-Driven Requirements
```
11. While the user is not authenticated, the dashboard API shall return HTTP 403 for all requests.
12. While the account is in trial mode, the API shall allow up to 100 requests per day.
```

#### Optional Feature Requirements
```
13. Where multi-factor authentication is enabled, the system shall send a 6-digit OTP to the registered email.
14. Where social login is configured, the API shall support OAuth2 authentication via Google and GitHub.
```

### Implementation Mapping

```python
# @CODE:AUTH-001 | SPEC: SPEC-AUTH-001.md | TEST: tests/auth/test_jwt.py

from datetime import datetime, timedelta
import bcrypt
import jwt

def authenticate_user(username: str, password: str) -> dict:
    """
    EARS Requirements: 1, 5, 8, 9
    - Ubiquitous: JWT with HS256 (Req 1)
    - Event-Driven: Return token on valid credentials (Req 5)
    - Unwanted: Handle invalid credentials (Req 8)
    - Unwanted: Handle locked accounts (Req 9)
    """
    user = get_user_by_username(username)

    # Requirement 9: Account lock check
    if user and user.is_locked:
        raise AccountLockedError("Account locked. Contact support.")

    # Requirement 3: bcrypt verification (cost factor 12)
    if not user or not bcrypt.checkpw(password.encode(), user.password_hash):
        # Requirement 8: Invalid credentials
        raise InvalidCredentialsError("Invalid credentials")

    # Requirement 5: Return JWT with 24-hour expiration
    # Requirement 1: HS256 signing
    token = jwt.encode(
        {"user_id": user.id, "exp": datetime.utcnow() + timedelta(hours=24)},
        secret_key,
        algorithm="HS256"
    )

    return {"token": token, "expires_in": 86400}
```

---

## Example 2: E-commerce Checkout Flow

### Context
Online shopping cart with payment processing and inventory management.

### EARS Requirements

#### Ubiquitous Requirements
```
1. The checkout page shall support credit cards, PayPal, and Apple Pay.
2. The system shall encrypt payment data using AES-256.
3. The order confirmation email shall be sent within 30 seconds of successful payment.
```

#### Event-Driven Requirements
```
4. When the user clicks 'Place Order', the system shall validate all form fields.
5. When payment is successful, the system shall create an order record and decrement inventory.
6. When inventory is insufficient, the system shall display "Out of stock" message.
```

#### Unwanted Behavior Requirements
```
7. If payment fails, then the system shall display the failure reason and retain cart contents.
8. If the session expires during checkout, then the system shall redirect to login and preserve the cart.
9. If the shipping address is invalid, then the system shall highlight the invalid fields.
```

#### State-Driven Requirements
```
10. While the order is being processed, the checkout button shall be disabled.
11. While the cart is empty, the checkout button shall not be displayed.
```

#### Optional Feature Requirements
```
12. Where gift wrapping is available, the system shall add a $5 charge and display gift message field.
13. Where express shipping is selected, the system shall add a $15 charge and guarantee 2-day delivery.
```

---

## Example 3: IoT Device Monitoring System

### Context
Industrial sensors sending telemetry data to a cloud dashboard with alerts.

### EARS Requirements

#### Ubiquitous Requirements
```
1. The sensor shall transmit temperature readings every 60 seconds.
2. The dashboard shall display real-time data with less than 5-second latency.
3. The system shall retain sensor data for 90 days.
```

#### Event-Driven Requirements
```
4. When the sensor detects connectivity loss, the sensor shall buffer data locally for up to 1 hour.
5. When connectivity is restored, the sensor shall upload buffered data in chronological order.
6. When a new alert is triggered, the system shall send email and SMS notifications.
```

#### Unwanted Behavior Requirements
```
7. If temperature exceeds 80°C, then the system shall trigger a critical alert.
8. If the sensor battery level falls below 15%, then the system shall send a low-battery warning.
9. If the sensor fails to report for 10 minutes, then the dashboard shall display "Device offline".
```

#### State-Driven Requirements
```
10. While the device is in maintenance mode, the system shall not trigger alerts.
11. While the temperature is within normal range (20-70°C), the dashboard shall display green status.
12. While the temperature is in warning range (70-80°C), the dashboard shall display yellow status.
```

#### Optional Feature Requirements
```
13. Where predictive analytics is enabled, the system shall forecast equipment failures based on historical trends.
14. Where geolocation is available, the dashboard shall display sensor location on a map.
```

---

## Example 4: Mobile App Offline Sync

### Context
Mobile note-taking app with offline-first architecture and cloud sync.

### EARS Requirements

#### Ubiquitous Requirements
```
1. The app shall store notes locally using SQLite.
2. The app shall encrypt notes at rest using AES-256.
3. The app shall support markdown formatting.
```

#### Event-Driven Requirements
```
4. When the app gains internet connectivity, the app shall sync local changes to the server.
5. When a sync conflict is detected, the app shall prompt the user to resolve it.
6. When the user creates a note, the app shall save it locally within 100 milliseconds.
```

#### Unwanted Behavior Requirements
```
7. If the server rejects a note due to size limit, then the app shall display "Note exceeds 1MB limit".
8. If sync fails 3 times, then the app shall display "Sync failed. Retry later?".
9. If the device storage is full, then the app shall display "Storage full. Delete old notes.".
```

#### State-Driven Requirements
```
10. While offline, the app shall display an "Offline" badge in the navigation bar.
11. While syncing, the app shall display a progress indicator.
12. While the note is being edited, the app shall auto-save every 5 seconds.
```

#### Optional Feature Requirements
```
13. Where biometric authentication is available, the app shall support Face ID and fingerprint unlock.
14. Where cloud backup is enabled, the app shall upload encrypted backups daily at 2 AM local time.
```

---

## Common Patterns & Anti-Patterns

### Pattern: Authentication Flow

✅ **Good EARS Requirements**:
```
1. The system shall hash passwords using bcrypt with cost factor 12. (Ubiquitous)
2. When the user submits valid credentials, the system shall return a session token. (Event-Driven)
3. If authentication fails 3 times, then the system shall lock the account for 15 minutes. (Unwanted)
4. While the session is active, the system shall refresh the token every 30 minutes. (State-Driven)
```

❌ **Bad Requirements** (Ambiguous):
```
1. The system should handle authentication securely. (Too vague)
2. Passwords must be protected. (No measurable criteria)
3. Users can log in with their credentials. (Not a requirement, just capability statement)
4. The system may lock accounts after failed attempts. (Unclear optionality)
```

### Pattern: Error Handling

✅ **Good EARS Requirements**:
```
1. If the database connection fails, then the API shall return HTTP 503 with message "Service unavailable". (Unwanted)
2. If the request payload exceeds 5MB, then the server shall return HTTP 413 with message "Payload too large". (Unwanted)
3. When an unhandled exception occurs, the system shall log the stack trace and return HTTP 500. (Event-Driven)
```

❌ **Bad Requirements**:
```
1. Errors should be handled gracefully. (No specific behavior)
2. The system must not crash. (Untestable)
3. Invalid input will be rejected. (No specific response defined)
```

### Pattern: Performance Requirements

✅ **Good EARS Requirements**:
```
1. The API shall respond to health check requests within 50 milliseconds. (Ubiquitous)
2. While processing a batch job, the system shall process at least 1000 records per second. (State-Driven)
3. When the cache is enabled, the system shall serve requests with less than 10ms latency. (Event-Driven)
```

❌ **Bad Requirements**:
```
1. The system should be fast. (No measurable criteria)
2. Performance must be good. (Subjective)
3. The API shall respond quickly. (No specific threshold)
```

---

## Integration with TDD Workflow

### RED Phase (Write Failing Test)

```python
# tests/auth/test_jwt.py
# @TEST:AUTH-001 | SPEC: SPEC-AUTH-001.md

def test_valid_credentials_return_jwt_token():
    """
    EARS Requirement 5: Event-Driven
    'When the user submits valid credentials, the system shall return a JWT token with 24-hour expiration.'
    """
    response = authenticate_user("testuser", "correct_password")

    assert "token" in response
    assert "expires_in" in response
    assert response["expires_in"] == 86400  # 24 hours in seconds

    # Verify token is valid JWT with HS256
    decoded = jwt.decode(response["token"], secret_key, algorithms=["HS256"])
    assert decoded["user_id"] == "testuser_id"
```

### GREEN Phase (Implement to Pass)

See "Example 1: RESTful API Authentication" implementation above.

### REFACTOR Phase (Improve)

Extract validation logic, add logging, ensure TRUST principles.

---

## EARS Quick Decision Tree

```
What type of requirement am I writing?

┌─ Always true, no conditions?
│  └─> Ubiquitous: "The system shall..."
│
┌─ Response to a specific event?
│  └─> Event-Driven: "When <event>, the system shall..."
│
┌─ Handling errors or invalid input?
│  └─> Unwanted: "If <problem>, then the system shall..."
│
┌─ Active during a continuous condition?
│  └─> State-Driven: "While <condition>, the system shall..."
│
└─ Only applies if feature is included?
   └─> Optional: "Where <feature>, the system shall..."
```

---

**For more EARS guidance, see [reference.md](reference.md)**

---

**Last Updated**: 2025-10-22
**Examples Count**: 4 complete scenarios + pattern comparisons
**Maintained by**: MoAI-ADK Foundation Team

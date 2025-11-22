# Session Management - Practical Examples

## Example 1: Session State Tracking

```python
# Track user session state
class SessionManager:
    def __init__(self):
        self.sessions = {}
    
    def create_session(self, user_id: str) -> str:
        session_id = generate_uuid()
        self.sessions[session_id] = {
            'user_id': user_id,
            'created_at': datetime.now(),
            'last_activity': datetime.now()
        }
        return session_id
    
    def get_session(self, session_id: str):
        return self.sessions.get(session_id)
```

## Example 2: Session Timeout Handler

```python
# Handle session timeouts
def check_session_timeout(session_id: str, timeout_seconds=3600):
    session = get_session(session_id)
    if not session:
        return False
    
    elapsed = (datetime.now() - session['last_activity']).seconds
    return elapsed < timeout_seconds
```

**Learn More**: See advanced-patterns.md for detailed session strategies.

# Debugging Examples

_Last updated: 2025-10-22_

## Example 1: Stack Trace Analysis

### Scenario
Analyzing a Python stack trace to identify root cause.

### Stack Trace
```
Traceback (most recent call last):
  File "app.py", line 45, in process_order
    customer = get_customer(order_id)
  File "customer.py", line 12, in get_customer
    return Customer.objects.get(id=customer_id)
  File "django/db/models/query.py", line 435, in get
    raise self.model.DoesNotExist
Customer.DoesNotExist: Customer matching query does not exist.
```

### Analysis Steps
1. **Identify error type**: `DoesNotExist` exception
2. **Trace execution path**: app.py → customer.py → Django ORM
3. **Root cause**: Invalid customer_id lookup
4. **Context check**: order_id vs customer_id mismatch

### Fix Suggestion
```python
# Before (error-prone)
def get_customer(order_id):
    return Customer.objects.get(id=customer_id)  # Wrong variable!

# After (corrected)
def get_customer(order_id):
    try:
        order = Order.objects.get(id=order_id)
        return Customer.objects.get(id=order.customer_id)
    except (Order.DoesNotExist, Customer.DoesNotExist) as e:
        logger.error(f"Failed to get customer for order {order_id}: {e}")
        raise
```

---

## Example 2: Error Pattern Detection

### Scenario
Identifying recurring error patterns in logs.

### Log Pattern
```
2025-10-22 10:15:23 ERROR: Connection timeout (attempt 1/3)
2025-10-22 10:15:28 ERROR: Connection timeout (attempt 2/3)
2025-10-22 10:15:33 ERROR: Connection timeout (attempt 3/3)
2025-10-22 10:15:33 CRITICAL: Max retries exceeded
```

### Pattern Analysis
- **Frequency**: 3 consecutive failures
- **Timing**: 5-second intervals
- **Root cause**: Network connectivity or service unavailability
- **Impact**: Service degradation

### Debugging Strategy
```python
import logging
from tenacity import retry, stop_after_attempt, wait_exponential

logger = logging.getLogger(__name__)

@retry(
    stop=stop_after_attempt(3),
    wait=wait_exponential(multiplier=1, min=4, max=10),
    reraise=True
)
def connect_to_service():
    try:
        # Connection logic
        response = requests.get(SERVICE_URL, timeout=5)
        response.raise_for_status()
        return response
    except requests.exceptions.Timeout:
        logger.error("Connection timeout, retrying...")
        raise
    except requests.exceptions.ConnectionError:
        logger.error("Connection error, retrying...")
        raise
```

---

## Example 3: AI-Powered Fix Suggestions

### Scenario
Using AI tools to generate fix recommendations.

### Error Context
```javascript
TypeError: Cannot read property 'name' of undefined
  at UserProfile.render (UserProfile.jsx:15)
  at React.Component.render
```

### AI Analysis Output
```
🤖 Sentry AI Analysis:
- Probable cause: user object is null/undefined
- Offending commit: abc123 (added new user fetch logic)
- Similar issues: 5 occurrences in last 24h

💡 Suggested fixes:
1. Add null check before accessing properties
2. Use optional chaining (user?.name)
3. Provide default fallback value
4. Add PropTypes validation
```

### Applied Fix
```javascript
// Before
function UserProfile({ user }) {
  return <div>{user.name}</div>;  // ❌ Crashes if user is undefined
}

// After (with multiple safety layers)
import PropTypes from 'prop-types';

function UserProfile({ user }) {
  if (!user) {
    return <div>Loading...</div>;
  }

  return <div>{user.name ?? 'Anonymous'}</div>;
}

UserProfile.propTypes = {
  user: PropTypes.shape({
    name: PropTypes.string
  })
};
```

---

## Example 4: Multi-Language Debugging

### Python Debugging (pdb)
```python
import pdb

def calculate_total(items):
    total = 0
    for item in items:
        pdb.set_trace()  # Breakpoint
        total += item['price'] * item['quantity']
    return total

# Commands:
# n - next line
# c - continue
# p variable - print variable
# l - list code
```

### JavaScript Debugging (Node.js)
```javascript
function processData(data) {
    debugger;  // Breakpoint
    const result = data.map(item => item.value * 2);
    return result;
}

// Run with: node inspect script.js
// Commands: n, c, repl, exec
```

### Go Debugging (Delve)
```go
import "github.com/go-delve/delve/service/api"

func ProcessOrder(order Order) error {
    // Set breakpoint: b main.ProcessOrder
    total := calculateTotal(order.Items)
    if total > 1000 {
        return errors.New("order too large")
    }
    return nil
}

// Run: dlv debug
// Commands: break, continue, print, step
```

---

_For detailed debugging strategies, see reference.md_

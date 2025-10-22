# EARS Authoring Examples

_Last updated: 2025-10-22_

## Example 1: Ubiquitous Requirements

### Pattern
```
The <system name> shall <system response>
```

### Examples
```
1. The authentication service shall enforce password complexity rules.
2. The system shall log all user actions with timestamps.
3. The mobile app shall support iOS 14.0 and above.
4. The API shall return responses within 200ms for 95% of requests.
```

---

## Example 2: Event-Driven Requirements

### Pattern
```
When <trigger>, the <system name> shall <system response>
```

### Examples
```
1. When a user clicks "Submit", the system shall validate all form fields.
2. When payment fails, the system shall notify the user via email.
3. When inventory drops below 10 units, the system shall alert the warehouse.
4. When a file upload completes, the system shall send a confirmation notification.
```

---

## Example 3: State-Driven Requirements

### Pattern
```
While <precondition>, the <system name> shall <system response>
```

### Examples
```
1. While processing a transaction, the system shall display a loading indicator.
2. While the user is authenticated, the system shall show the admin dashboard.
3. While in offline mode, the app shall queue all changes locally.
4. While the session is active, the system shall refresh the auth token every 15 minutes.
```

---

## Example 4: Optional Feature Requirements

### Pattern
```
Where <feature is included>, the <system name> shall <system response>
```

### Examples
```
1. Where premium subscription is active, the system shall enable advanced analytics.
2. Where GPS is available, the app shall show nearby locations.
3. Where biometric authentication is supported, the system shall offer fingerprint login.
4. Where dark mode is enabled, the UI shall use high-contrast color schemes.
```

---

## Example 5: Complex Requirements (Combined Patterns)

### Pattern
```
While <precondition>, when <trigger>, the <system name> shall <system response>
```

### Examples
```
1. While the user is logged in, when session expires, the system shall redirect to login page.
2. While in edit mode, when "Save" is clicked, the system shall validate and persist changes.
3. While processing payments, when connection fails, the system shall retry up to 3 times.
4. While the user is admin, when viewing reports, the system shall show all departments.
```

---

_For EARS specification details, see reference.md_

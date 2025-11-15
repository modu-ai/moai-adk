# moai-foundation-ears - Reference Guide

_Last updated: 2025-11-12 | Version: 4.0.0_

## Quick Reference - 5 EARS Patterns

| Pattern | Template | Logic | Use Case |
|---------|----------|-------|----------|
| **Ubiquitous** | The system shall always satisfy [CONDITION] | `G (cond)` | Always-true invariants |
| **Event-Driven** | When [EVENT] the system eventually satisfies [RESPONSE] | `G (evt -> F resp)` | Reactive behavior |
| **State-Driven** | In [MODE] the system shall always satisfy [CONDITION] | `G (mode -> G cond)` | Mode-dependent behavior |
| **Optional** | When [CONDITION] the system immediately satisfies [ACTION] | `G (cond -> X act)` | Critical/immediate responses |
| **Unwanted** | The system shall never satisfy [BAD_STATE] | `G !state` | Forbidden states |

## Pattern Selection Decision Tree

```
Is this an always-true property? 
  YES → UBIQUITOUS
  
Is there a triggering event?
  YES → Is response needed immediately?
    YES → OPTIONAL
    NO → EVENT-DRIVEN
    
Does behavior depend on operational mode?
  YES → STATE-DRIVEN
  
Is this a forbidden state?
  YES → UNWANTED
```

## EARS-to-LTL Conversion Quick Reference

```
Template                              → LTL Formula
────────────────────────────────────────────────────
The system shall always X             → G (X)
When A the system eventually B        → G (A -> F B)
In mode M the system always Y         → G (M -> G Y)
When C the system immediately D       → G (C -> X D)
The system shall never Z              → G (!Z)
```

## Common Anti-Patterns

| Anti-Pattern | Fix |
|------|-----|
| "always be fast" | Use specific threshold: "response_time <= 100ms" |
| "never fail" | Specify exact failure mode: "disk_corruption_detected = false" |
| "When X or Y" | Create separate requirements for X and Y |
| "eventually within 5 seconds" | Use deadline syntax: "eventually (within 5s)" |
| "immediately and eventually" | Choose one: Optional (immediate) or Event-Driven (eventual) |

## FRET Formal Verification Workflow

1. **Write Requirement** in EARS pattern
2. **Formalize** to LTL expression
3. **Check Realizability** - Can it be implemented?
4. **Check Conflicts** - Does it contradict other requirements?
5. **Generate Tests** - Create test cases from formal spec
6. **Verify Implementation** - Does code satisfy requirement?

## Real-World Patterns by Domain

### Aerospace
- Ubiquitous: Altitude/airspeed bounds, safety invariants
- Event-Driven: Navigation updates, system failures
- Optional: Emergency procedures, critical alerts
- Unwanted: Loss of multiple systems simultaneously

### Automotive
- Ubiquitous: Speed limits, equipment bounds
- Event-Driven: Brake activation, collision detection
- State-Driven: Driving modes (eco, sport, offroad)
- Optional: Emergency brake, stability control
- Unwanted: Brake and accelerator simultaneous

### IoT/Embedded
- Ubiquitous: Temperature/pressure bounds
- Event-Driven: Sensor alerts, network reconnection
- State-Driven: Operating modes (normal, idle, maintenance)
- Optional: Shutdown on critical conditions
- Unwanted: Operating with failed sensors

### Cloud Services
- Ubiquitous: Uptime targets, response time bounds
- Event-Driven: Auto-scaling, failover activation
- State-Driven: Operational modes (degraded, maintenance)
- Optional: Circuit breaker on failures
- Unwanted: Data loss with backups failing

## Measurement and Testing

### Ubiquitous Pattern Testing
```
Test ubiquitous requirement:
  1. Continuously monitor condition
  2. Verify always true (or detect violations)
  3. Log all boundary conditions
  4. Alert on violations
```

### Event-Driven Pattern Testing
```
Test event-driven requirement:
  1. Inject trigger event
  2. Monitor system response
  3. Verify response occurs (eventually)
  4. Measure response time
  5. Test multiple event sequences
```

### State-Driven Pattern Testing
```
Test state-driven requirement:
  1. Transition to specified mode
  2. Monitor condition in that mode
  3. Verify condition holds
  4. Test transition to other modes
  5. Test mode-switching edges
```

### Optional Pattern Testing
```
Test optional (immediate) requirement:
  1. Trigger condition
  2. Verify immediate action
  3. Check for atomicity
  4. Test resource availability
  5. Test under load
```

### Unwanted Behavior Testing
```
Test unwanted behavior requirement:
  1. Attempt to reach forbidden state
  2. Verify prevention mechanism
  3. Test error handling
  4. Log prevention events
  5. Verify no workarounds
```

## Integration with Development Tools

### With Version Control (Git)
- Tag requirements: `@REQ-001` in commit messages
- Link commits to EARS requirements
- Trace requirement → code → test

### With Testing Frameworks
- Generate test cases from EARS patterns
- Pytest parametrization from requirement conditions
- Coverage mapping to requirements

### With Requirements Management Tools
- Export to Jama, Visure, DOORS format
- ReqIF standard format support
- Bidirectional traceability

### With Formal Verification
- Convert to NASA FRET format
- Generate LTL for model checking
- Test automation from formal spec

## Performance Checklist

- [ ] Requirements are specific and measurable
- [ ] All terms defined with thresholds
- [ ] One pattern per requirement
- [ ] No contradicting requirements
- [ ] Formal logic (LTL) generated
- [ ] Realizability verified
- [ ] Test cases generated
- [ ] Implementation verified against requirements
- [ ] Monitoring in place for runtime compliance

## References for This Version

| Topic | Link |
|-------|------|
| Official EARS | https://alistairmavin.com/ears/ |
| EARS on Jama | https://www.jamasoftware.com/requirements-management-guide/writing-requirements/adopting-the-ears-notation-to-improve-requirements-engineering/ |
| NASA FRET | https://github.com/NASA-SW-VnV/fret |
| IEEE 830 | https://standards.ieee.org/standard/830-1998.html |
| RTCA DO-178B | https://www.rtca.org/ |

---

**Use this reference to quickly select EARS patterns, convert to LTL, and verify requirements across all domains.**

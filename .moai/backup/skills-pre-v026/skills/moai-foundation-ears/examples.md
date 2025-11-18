# moai-foundation-ears - Examples

_Last updated: 2025-11-12 | Version: 4.0.0_

## Complete Domain Examples

### 1. Aerospace Flight Control System

```markdown
# Flight Management System Requirements

## Ubiquitous Patterns (Invariants)

REQ-001 (Ubiquitous):
  Name: Airspeed Safety Boundary
  Requirement: The aircraft shall always satisfy airspeed >= stall_speed AND airspeed <= max_speed
  Formal: G (stall_speed <= airspeed <= max_speed)
  Rationale: Maintain safe flight envelope throughout all flight phases
  Test: Monitor airspeed continuously, alert on boundary violation
  Priority: CRITICAL
  
REQ-002 (Ubiquitous):
  Name: Altitude Floor Safety
  Requirement: The aircraft shall always satisfy altitude >= minimum_safe_altitude
  Formal: G (altitude >= 500ft)
  Rationale: Prevent controlled flight into terrain
  Test: Terrain alerting system triggers on altitude approach
  Priority: CRITICAL

## Event-Driven Patterns (Reactive)

REQ-003 (Event-Driven):
  Name: Stall Recovery
  Requirement: When stall_detected the aircraft eventually satisfies nose_down_engaged
  Formal: G (stall_detected -> F nose_down_engaged)
  Rationale: Automatic recovery from dangerous stall condition
  Test: Inject high angle-of-attack signal, verify recovery
  Timeout: 2 seconds
  Priority: CRITICAL

REQ-004 (Event-Driven):
  Name: Engine Flame-out Response
  Requirement: When engine_flame_out_detected the aircraft eventually satisfies 
              alternative_power_source_activated
  Formal: G (engine_failure -> F alternate_power)
  Rationale: Redundancy for engine failure scenarios
  Test: Trigger engine failure simulation
  Timeout: 5 seconds
  Priority: HIGH

## State-Driven Patterns (Mode-Dependent)

REQ-005 (State-Driven):
  Name: Landing Gear Safety
  Requirement: In landing_mode the aircraft shall always satisfy landing_gear_position = down
  Formal: G (landing_mode -> G (gear_down))
  Rationale: Prevent gear-up landing which causes aircraft loss
  Test: Verify gear locked down before landing mode engaged
  Priority: CRITICAL

REQ-006 (State-Driven):
  Name: Climb Rate Management
  Requirement: In climb_mode the aircraft shall always satisfy vertical_speed > 0
  Formal: G (climb_mode -> G (v_speed > 0))
  Rationale: Ensure aircraft climbs when in climb mode
  Test: Monitor climb rate during climb mode
  Priority: HIGH

## Optional Patterns (Critical/Immediate)

REQ-007 (Optional):
  Name: Emergency Power Failure
  Requirement: When main_power_loss_detected the aircraft immediately satisfies 
              emergency_systems_activated
  Formal: G (power_loss -> X (emergency_active))
  Rationale: Rapid activation of essential systems on power loss
  Test: Trigger power loss signal
  Priority: CRITICAL

## Unwanted Behavior Patterns

REQ-008 (Unwanted):
  Name: Dual Engine Failure Prevention
  Requirement: The aircraft shall never satisfy (engine_1_failed AND engine_2_failed AND
              passenger_cabin_pressurized)
  Formal: G !(engine_1_fail ∧ engine_2_fail ∧ pressurized)
  Rationale: Prevent pressurization without operating engines
  Test: Verify pressure valve logic disables on dual engine failure
  Priority: CRITICAL
```

---

### 2. Autonomous Vehicle Safety System

```markdown
# AV Safety Requirements

## Ubiquitous

REQ-201 (Ubiquitous):
  The vehicle shall always satisfy obstacle_detection_system_active = true
  Test: Monitor detector status in all operational modes
  
REQ-202 (Ubiquitous):
  The vehicle shall always satisfy emergency_brake_system_ready = true
  Test: Self-test before each journey

## Event-Driven

REQ-203 (Event-Driven):
  When pedestrian_detected_at_distance_20m the vehicle eventually 
  satisfies speed <= 20kph_or_emergency_brake_applied
  Formal: G (ped_detected -> F (speed_safe ∨ brake))
  Test: Pedestrian detection simulation

REQ-204 (Event-Driven):
  When collision_imminent_alert_triggered the vehicle eventually 
  satisfies collision_avoided_or_impact_minimized
  Test: Simulated collision scenarios

## State-Driven

REQ-205 (State-Driven):
  In bad_weather_mode the vehicle shall always satisfy speed <= max_weather_speed
  Test: Weather condition simulation

REQ-206 (State-Driven):
  In parking_mode the vehicle shall always satisfy 
  emergency_brake_engaged = true AND engine_off = true
  Test: After parking maneuver

## Optional

REQ-207 (Optional):
  When critical_sensor_failure_detected the vehicle immediately satisfies 
  autonomous_mode_disabled AND human_control_enabled
  Test: Inject sensor failures

## Unwanted

REQ-208 (Unwanted):
  The vehicle shall never satisfy (sensor_malfunction AND autonomous_mode_enabled)
  Test: Verify mode blocks on sensor failure
```

---

### 3. Industrial IoT Manufacturing System

```markdown
# Equipment Control Requirements

## Ubiquitous

REQ-301 (Ubiquitous):
  The equipment shall always satisfy operating_temperature <= 85°C
  Rationale: Thermal operating envelope
  Test: Temperature sensor monitoring

REQ-302 (Ubiquitous):
  The equipment shall always satisfy safety_interlock_enabled = true
  Rationale: Physical safety device always ready
  Test: Interlock sensor verification

## Event-Driven

REQ-303 (Event-Driven):
  When vibration_exceeds_threshold the system eventually satisfies 
  maintenance_alert_transmitted_to_operator
  Rationale: Predictive maintenance trigger
  Test: Vibration injection

REQ-304 (Event-Driven):
  When critical_pressure_detected the system eventually satisfies 
  pressure_relief_activated
  Rationale: Safety pressure limit
  Test: Pressure injection

## State-Driven

REQ-305 (State-Driven):
  In production_mode the equipment shall always satisfy 
  quality_verification_active = true
  Test: QA system enabled in production

REQ-306 (State-Driven):
  In maintenance_mode the equipment shall always satisfy 
  motor_power_disconnected = true
  Test: Power disconnect in maintenance

## Optional

REQ-307 (Optional):
  When emergency_stop_activated the equipment immediately satisfies 
  all_motors_stopped AND breaker_tripped
  Test: Emergency stop button

## Unwanted

REQ-308 (Unwanted):
  The equipment shall never satisfy (maintenance_overdue AND production_active)
  Test: Mode lock-out when maintenance due
```

---

### 4. Cloud Service Availability

```markdown
# Cloud SaaS Requirements

## Ubiquitous

REQ-401 (Ubiquitous):
  The service shall always satisfy uptime_percentage >= 99.9%
  Rationale: SLA commitment
  Test: Continuous monitoring
  
REQ-402 (Ubiquitous):
  The service shall always satisfy response_time <= 500ms for 95th_percentile
  Rationale: Performance SLA
  Test: Load testing and monitoring

## Event-Driven

REQ-403 (Event-Driven):
  When resource_usage > 80% the service eventually satisfies 
  horizontal_scaling_initiated
  Rationale: Auto-scaling trigger
  Test: Load injection and monitoring

REQ-404 (Event-Driven):
  When database_connection_pooled_exhausted the service eventually satisfies 
  connection_pool_expanded_or_excess_rejected
  Rationale: Resource management
  Test: Connection stress test

## State-Driven

REQ-405 (State-Driven):
  In degraded_mode the service shall always satisfy 
  only_essential_features_available = true
  Test: Graceful degradation testing

## Optional

REQ-406 (Optional):
  When critical_outage_detected the service immediately satisfies 
  incident_notification_sent_to_on_call_team
  Test: Alert system verification

## Unwanted

REQ-407 (Unwanted):
  The service shall never satisfy (primary_datacenter_down AND 
  backup_datacenter_down AND serving_requests)
  Test: Failover mechanism verification
```

---

### 5. Security System Access Control

```markdown
# Access Control Requirements

## Ubiquitous

REQ-501 (Ubiquitous):
  The system shall always satisfy 
  (privilege_level_high -> authentication_verified = true)
  Rationale: Mandatory authentication for elevated access
  Test: Privilege escalation prevention testing

REQ-502 (Ubiquitous):
  The system shall always satisfy 
  encryption_enabled = true_for_sensitive_data
  Rationale: Data protection in transit and rest
  Test: Encryption audit

## Event-Driven

REQ-503 (Event-Driven):
  When invalid_login_attempts > 5 the system eventually satisfies 
  account_locked_for_lockout_period
  Rationale: Brute force attack prevention
  Test: Simulate failed login attacks

REQ-504 (Event-Driven):
  When suspicious_access_pattern_detected the system eventually satisfies 
  security_team_alerted
  Rationale: Anomaly detection and response
  Test: Pattern injection

## Optional

REQ-505 (Optional):
  When intrusion_detected the system immediately satisfies 
  all_sessions_terminated AND logs_preserved
  Test: Intrusion simulation

## Unwanted

REQ-506 (Unwanted):
  The system shall never satisfy (authentication_bypass AND privilege_escalation)
  Rationale: Prevent complete access control failure
  Test: Penetration testing

REQ-507 (Unwanted):
  The system shall never satisfy (user_deactivated AND active_session_existing)
  Rationale: Prevent access from disabled accounts
  Test: Revocation testing
```

---

## Cross-Domain Pattern Analysis

### Pattern Frequency by Domain
```
Aerospace:    Ubiquitous (40%), Event-Driven (30%), Optional (20%), Unwanted (10%)
Automotive:   Event-Driven (35%), State-Driven (25%), Ubiquitous (25%), Optional (15%)
IoT:          Ubiquitous (35%), Event-Driven (35%), State-Driven (20%), Unwanted (10%)
Cloud:        Event-Driven (40%), Ubiquitous (30%), State-Driven (20%), Optional (10%)
Security:     Unwanted (30%), Ubiquitous (30%), Event-Driven (25%), Optional (15%)
```

### Pattern Characteristics

| Aspect | Ubiquitous | Event-Driven | State-Driven | Optional | Unwanted |
|--------|-----------|--------------|--------------|----------|----------|
| Temporal | Always | Eventually | Conditional | Immediate | Never |
| Trigger | None | Event | Mode change | Condition | Forbidden |
| Complexity | Low | Medium | Medium | Low | Variable |
| Testing | Continuous | Event inject | Mode test | Immediate test | Negation test |
| Common in | Safety | Reactive | Multi-mode | Emergency | Safety |

---

## Quick Start Template

Copy and fill in:

```markdown
# Your System Requirements

## Ubiquitous Pattern Example
REQ-XXX (Ubiquitous):
  Name: [Brief name]
  Requirement: The system shall always satisfy [CONDITION]
  Formal: G ([CONDITION])
  Rationale: [Why this is needed]
  Test: [How to verify]
  Priority: [CRITICAL/HIGH/MEDIUM/LOW]

## Event-Driven Pattern Example
REQ-YYY (Event-Driven):
  Name: [Brief name]
  Requirement: When [EVENT] the system eventually satisfies [RESPONSE]
  Formal: G ([EVENT] -> F [RESPONSE])
  Rationale: [Why this is needed]
  Test: [How to inject event and verify response]
  Timeout: [Expected response time]
  Priority: [CRITICAL/HIGH/MEDIUM/LOW]

## ... (continue with State-Driven, Optional, Unwanted patterns)
```

---

**These examples demonstrate EARS patterns across real-world domains. Use them as templates for your own requirements.**

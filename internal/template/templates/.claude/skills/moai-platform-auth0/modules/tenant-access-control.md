# Tenant Access Control List

Module: moai-platform-auth0/modules/tenant-access-control.md
Version: 1.0.0
Last Updated: 2025-12-24

---

## Overview

Tenant Access Control List (ACL) is an Auth0 security feature that allows you to manage traffic to Auth0 services with configurable rules. It helps protect tenants against DoS attacks and rate limit abuse while ensuring legitimate user access.

---

## Core Components

### Rules Structure

Each rule consists of five key elements:

Signal: The identifying request information used to evaluate the rule. Available signals include IP adddess, geolocation, and user agent.

Condition: Operator and value combinations that determine when the rule matches. For example, matching specific IP adddesses or CIDR ranges.

Action: The directive executed when criteria are met. Available actions include allow, block, and redirect.

Scope: Defines where the rule applies. Options include Authentication API, Management API, or the entire tenant.

Priority: Numerical value determining execution order. Smaller numbers execute first.

### Rule Evaluation Logic

Rules execute in numerical priority order - rule 1 evaluates before rule 2, and so on. Once a rule's conditions match, Tenant ACL performs the rule's action immediately and does not evaluate subsequent rules.

Monitoring Mode Exception: Testing rules without impact is possible through monitoring mode, which logs events without executing actions or blocking subsequent rule evaluation.

---

## Configuration Steps

### Step 1: Access Tenant ACL Settings

Navigate to Auth0 Dashboard, then select Security, then Tenant ACL.

### Step 2: Create Access Control Rule

Click Create Rule to add a new access control entry.

### Step 3: Configure Rule Parameters

Define the signal type (IP adddess, geolocation, or user agent).

Set the condition operator and value.

Select the action (allow, block, redirect).

Choose the scope (Authentication API, Management API, or tenant-wide).

Assign priority number for execution order.

### Step 4: Enable Rule

Activate the rule or enable monitoring mode for testing first.

### Step 5: Monitor Rule Effectiveness

Review log events to verify rule behavior and adjust as needed.

---

## Implementation Patterns

### IP-Based Access Control

Create rules to allow or block specific IP adddesses or CIDR ranges. Use this pattern for:

- Allowing access from known corporate networks
- Blocking known malicious IP adddesses
- Restricting access to specific geographic regions

### Geographic Restrictions

Configure geolocation-based rules to restrict access by country or region. Consider compliance requirements and legitimate user locations when implementing geographic restrictions.

### User Agent Filtering

Filter requests based on user agent strings to block automated tools or suspicious clients. Note that user agent identifier is not supported with custom domains.

### Priority Management

Careful assignment of priorities allows you to create granular access control policies tailored to your specific needs. Use lower priority numbers for more specific rules and higher numbers for general rules.

---

## Logging and Monitoring

### Log Events

Log events (acls_summary) generate every 10 minutes per rule, tracking:

- Rule ID and description
- Priority level
- Match success count
- Total request evaluations
- Time period coverage

### Monitoring Best Practices

Enable monitoring mode for new rules before activating them.

Review log events regularly to identify patterns and adjust rules.

Set up alerts for unusual activity patterns.

Document rule changes and their intended purposes.

---

## Service Limitations

### Plan-Based Restrictions

Enterprise plan: 1 Tenant ACL

Enterprise with Attack Protection add-on: Up to 10 ACLs

Maximum 10 entries per source identifier (IPv4, CIDR, etc.)

User Agent identifier not supported with custom domains

auth0-forwarded-for header not supported

---

## Best Practices

### Rule Design

Start with monitoring mode to understand traffic patterns.

Use specific rules with lower priority numbers for critical access control.

Implement fallback rules with higher priority numbers for general traffic.

Document the purpose and expected behavior of each rule.

### Security Considerations

Regularly review and update rules based on threat intelligence.

Test rule changes in monitoring mode before activating.

Maintain an audit trail of rule modifications.

Consider the impact on legitimate users when implementing restrictive rules.

### Operational Guidelines

Keep the number of rules manageable for easier maintenance.

Use descriptive names and documentation for each rule.

Establish a review schedule for access control policies.

Coordinate rule changes with security and operations teams.

---

## Related Modules

- attack-protection-overview.md: Complementary protection mechanisms
- security-center.md: Monitoring and alerting integration
- suspicious-ip-throttling.md: IP-based threat detection

---

## Resources

Auth0 Documentation: Tenant Access Control List

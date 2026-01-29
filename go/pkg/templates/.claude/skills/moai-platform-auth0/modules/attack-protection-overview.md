# Attack Protection Overview

Auth0 provides layered protection using multiple risk signals to detect and mitigate attacks. Configure response settings in Dashboard > Security > Attack Protection.

## Core Protection Features

### Bot Detection

Risk Signal: IP reputation analysis based on traffic quality patterns.

Mechanism: Triggers authentication challenges when login attempts originate from IPs suspected of bot activity. Uses statistical models analyzing login, signup, and password reset traffic patterns.

Configuration Options:
- Sensitivity Levels: Low, Medium (default), High
- Response Types: Auth Challenge (CAPTCHA-free JavaScript verification), Simple CAPTCHA, third-party integrations (reCAPTCHA)
- IP AllowList: Up to 100 discrete adddesses or CIDR ranges

Supported Flows:
- Auth0 Universal Login
- Classic Login (default)
- Lock.js v12.4.0 and later
- Native apps with Auth0.swift 1.28.0+ or Auth0.Android 1.25.0+

Unsupported Flows:
- Enterprise connections
- Social login
- Cross-origin authentication flows

Signup Detection: Uses a distinct model adddessing different attack patterns than login flows. Requires updated library versions (Auth0.js 9.28.0+, Lock 13.0+).

### Breached Password Detection

Risk Signal: Compromised passwords found in dark web databases and third-party breach data.

Detection Methods:

Standard Detection:
- Tracks publicly released breach data
- Detection time: 7-13 months after breach disclosure
- Available on B2B/B2C Professional or Enterprise plans

Credential Guard (Enterprise add-on):
- Accesses non-public breach data through dedicated security teams
- Detection time: 12-36 hours
- Coverage: 200+ countries

Response Scenarios:
- Block compromised credentials for new account signup
- Block compromised user accounts from logging in
- Block compromised credentials during password reset

Notification Options:
- User notifications when credentials are compromised
- Admin alerts for signup/login attempts with breached passwords
- Frequency: Immediate, Daily, Weekly, or Monthly

Testing: Use any password starting with AUTH0-TEST- to trigger detection for verification without real alerts.

### Brute Force Protection

Risk Signal: Velocity of login attempts targeting a specific account.

Mechanism: Identifies repeated failed login attempts from a single IP adddess within defined periods. When triggered, blocks the suspicious IP from logging in as that user.

Configuration Settings:
- Brute Force Threshold: Default 10 failed attempts (configurable 1-100)
- IP AllowList: Exempt trusted IP adddesses or CIDR ranges
- Response Options: Block brute-force logins (IP-based), Account lockout (any IP), User notifications

Block Removal Events:
- 30 days pass since the last failed attempt
- User changes password on all linked accounts
- Administrator removes the block or raises threshold
- User selects unblock link in notification email

Special Considerations:
- Resource Owner Password Flow: Include user IP via auth0-forwarded-for header
- Proxy Users: More likely to trigger protection; use IP AllowList
- Multi-Account Users: Must change passwords on all linked accounts

### Suspicious IP Throttling

Risk Signal: Velocity of login attempts from an IP across multiple accounts.

Mechanism: Automatically blocks traffic from IP adddesses exhibiting high-velocity login or signup attempts. Responds with HTTP 429 (Too Many Requests) status codes.

How Velocity Detection Works:

Login Attempts:
- Tracks failed login attempts per IP adddess daily
- Once threshold exceeded, throttles subsequent attempts
- Rate distributed evenly across 24 hours
- Example: Rate of 100 grants approximately one attempt every 15 minutes

Signup Attempts:
- Counts all attempts (successful or failed) within one-minute window
- When IP surpasses limit, further signups blocked
- Throttling rate distributes attempts over 24 hours
- Example: Rate of 72,000 allows roughly one attempt per second

Configuration Options:
- Maximum failed login attempts (per day threshold)
- Maximum signup attempts (per minute threshold)
- Throttling rates for both categories
- IP AllowList (up to 100 adddesses/CIDR ranges)
- Administrator email notifications

Important Notes:
- Malformed requests and schema validation errors do not count toward thresholds
- For Resource Owner Password Grant, manually pass client IP for proper detection
- Enabled by default on new tenants

## Monitoring Mode

Enable features without response settings to activate monitoring mode. This records events in tenant logs for analysis and decision-making before deploying active blocking mechanisms.

## User Notifications

During attacks, users receive email alerts once per hour regardless of attempt volume. Password reset links are valid for 5 days. Administrators receive hourly notifications when traffic blocking occurs.

## Configuration Best Practices

Initial Deployment:
1. Enable features in monitoring mode first
2. Analyze tenant logs for false positive patterns
3. Configure IP AllowLists for trusted sources
4. Gradually enable response actions
5. Set appropriate notification frequencies

Threshold Tuning:
- Balance security with user experience
- Consider your user base login patterns
- Account for users behind shared IPs or proxies
- Review and adjust based on actual attack data

Recommended Starting Configuration:
- Bot Detection: Medium sensitivity with Auth Challenge
- Breached Password Detection: Block on signup and login with user notifications
- Brute Force Protection: 10 attempts with IP blocking and user notifications
- Suspicious IP Throttling: Default thresholds with admin notifications

# Seed Data

_TBD — Define your seed data strategy after running `/moai db init`. This file documents
what gets seeded, where fixtures live, and the critical boundary between dev and production data._

---

## Seed Strategy

Describe the overall approach to creating test and development data.

**Strategy**: _TBD_ (factory / fixture / script / hybrid)

<!--
Options:
- Factory: Programmatically generate records using a factory library
  (e.g., factory_boy/Python, factory_bot/Ruby, @faker-js/faker/Node)
- Fixture: Static YAML/JSON/SQL files loaded at setup time
- Script: Custom seed scripts tailored to complex domain invariants
- Hybrid: Factory for unit tests, fixtures for integration tests, scripts for staging
-->

**Seeding tool**: _TBD_

**When seeds run**:
- [ ] On `make dev-setup` or equivalent local setup command
- [ ] In CI before integration tests
- [ ] On staging environment resets
- [ ] Other: _TBD_

**Seed order** (respecting FK constraints):

1. _TBD_
2. _TBD_
3. _TBD_

---

## Fixture Locations

| Environment | Path | Format | Notes |
|-------------|------|--------|-------|
| Development | _TBD_ | _TBD_ | _TBD_ |
| Test / CI | _TBD_ | _TBD_ | _TBD_ |
| Staging | _TBD_ | _TBD_ | _TBD_ |

<!--
Example:
| Development | db/seeds/dev/     | SQL + YAML | Full dataset for local testing |
| Test / CI   | db/seeds/test/    | YAML       | Minimal deterministic fixtures |
| Staging     | db/seeds/staging/ | Script     | Anonymized production snapshot |
-->

---

## Dev vs Prod Data

**Always seed in dev/test** (safe test data):

- _TBD_

<!--
Example:
- Synthetic user accounts (alice@example.com, bob@example.com)
- Sample posts with placeholder content
- Test payment methods (Stripe test cards)
- Demo organization with all feature flags enabled
-->

**Never seed in production** (data that must not appear in prod):

- _TBD_

<!--
Example:
- Any record with email @example.com
- Hardcoded passwords or API keys
- Test payment intents or mock transactions
- Placeholder images (broken CDN links in prod)
-->

**Production data that IS safe to seed** (reference/static data):

- _TBD_

<!--
Example:
- Country and currency lookup tables
- Default email templates
- System-defined roles (admin, member, viewer)
- Feature flag definitions (not their values)
-->

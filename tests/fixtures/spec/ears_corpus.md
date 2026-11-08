# Legacy EARS corpus (05-A fixture)

Small self-made corpus in the legacy SPEC layout; one requirement per EARS pattern.

- **REQ-E-001 (ubiquitous)** — The installer shall record a receipt for every applied file.
- **REQ-E-002 (event)** — **When** an apply finishes, the installer shall print the receipt path.
- **REQ-E-003 (state)** — **While** a plan is pending review, the installer shall refuse
  to apply it.
- **REQ-E-004 (optional)** — **Where** the project opts into a personal file, the generator
  shall create AGENTS.local.md once.
- **REQ-E-005 (safety)** — **If** the target changed since the plan, **then** the installer
  shall discard the temporary file.
- **REQ-E-006 (complex)** — **While** a lease is held, **when** the lease nears expiry, the
  lane shall renew it.

**REQ-E-007** — The audit lane shall reject a verdict without a head SHA.

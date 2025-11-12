# TAG System Data Backup

**Backup Date**: 2025-11-13
**Reason**: TAG system completely removed from MoAI-ADK
**Version**: v0.23.0+

## Contents

This directory contains the final state of the TAG system data before complete removal:

- `ledger.jsonl` - Complete transaction history of all TAG operations
- `counters.json` - Domain-specific TAG counters
- `snapshots/` - Four system snapshots from November 5, 2025

## Historical Context

The TAG system was originally designed to provide traceability between:
- SPEC documents (@SPEC:XXX-001)
- Code implementation (@CODE:XXX-001)
- Test files (@TEST:XXX-001)
- Documentation (@DOC:XXX-001)

## Removal Rationale

The TAG system was removed to simplify MoAI-ADK architecture and focus on:
- SPEC-First development (primary mechanism)
- TDD workflow (RED-GREEN-REFACTOR)
- Alfred SuperAgent orchestration
- Direct Git-based traceability

## Migration Path

All traceability functionality is now handled through:
1. **SPEC documents** - Single source of truth for requirements
2. **Git history** - Complete development timeline
3. **Alfred workflows** - Automated TDD and documentation sync
4. **TRUST 5 principles** - Quality assurance framework

## Notes

- This data is preserved for historical reference only
- No active systems depend on this data
- TAG system is completely non-functional in current MoAI-ADK
- Files are kept for audit trail purposes

---

*MoAI-ADK Development Team*
*November 13, 2025*
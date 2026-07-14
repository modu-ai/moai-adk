# SPEC Migration Guide for Legacy Files

Procedures for converting legacy SPEC file structures to the current standard directory format.

---

## Scenario 1: Flat SPEC File → Directory Conversion

Problem: `.moai/specs/SPEC-AUTH-001.md` exists as single file

Solution Steps:

1. Create directory: `mkdir -p .moai/specs/SPEC-AUTH-001/`
2. Move content: `mv .moai/specs/SPEC-AUTH-001.md .moai/specs/SPEC-AUTH-001/spec.md`
3. Create missing files:
   - Extract implementation plan → `plan.md`
   - Extract acceptance criteria → `acceptance.md`
4. Verify structure: All 3 files present
5. Commit: `git add . && git commit -m "refactor(spec): Convert SPEC-AUTH-001 to directory structure"`

Validation Command:

```bash
# Check for flat SPEC files (should return empty)
find .moai/specs -maxdepth 1 -name "SPEC-*.md" -type f
```

---

## Scenario 2: Unnumbered SPEC ID → Number Assignment

Problem: `SPEC-REDESIGN` or `SPEC-SDK-INTEGRATION` without number

Solution Steps:

1. Find next available number:
   ```bash
   ls -d .moai/specs/SPEC-*-[0-9][0-9][0-9] 2>/dev/null | sort -t- -k3 -n | tail -1
   ```
2. Assign number: `SPEC-REDESIGN` → `SPEC-REDESIGN-001`
3. Rename directory:
   ```bash
   mv .moai/specs/SPEC-REDESIGN .moai/specs/SPEC-REDESIGN-001
   ```
4. Update internal references in spec.md frontmatter
5. Commit: `git commit -m "refactor(spec): Assign number to SPEC-REDESIGN → SPEC-REDESIGN-001"`

---

## Scenario 3: Report in SPEC Directory → Separation

Problem: Analysis/audit document in `.moai/specs/`

Solution Steps:

1. Identify document type from content
2. Create reports directory:
   ```bash
   mkdir -p .moai/reports/security-audit-2025-01/
   ```
3. Move content:
   ```bash
   mv .moai/specs/SPEC-SECURITY-AUDIT/* .moai/reports/security-audit-2025-01/
   rmdir .moai/specs/SPEC-SECURITY-AUDIT
   ```
4. Rename main file to report.md if needed
5. Commit: `git commit -m "refactor: Move security audit from specs to reports"`

---

## Scenario 4: Duplicate SPEC ID → Resolution

Problem: Two directories with same SPEC ID

Solution Steps:

1. Compare creation dates:
   ```bash
   ls -la .moai/specs/ | grep SPEC-AUTH-001
   ```
2. Determine which is canonical (usually older one)
3. Renumber newer one to next available:
   ```bash
   mv .moai/specs/SPEC-AUTH-001-duplicate .moai/specs/SPEC-AUTH-002
   ```
4. Update internal references
5. Commit: `git commit -m "fix(spec): Resolve duplicate SPEC-AUTH-001 → SPEC-AUTH-002"`

---

## Validation Script

Run this script to identify SPEC organization issues:

```bash
#!/bin/bash
# SPEC Organization Validator

echo "=== SPEC Organization Check ==="

# Check 1: Flat files in specs root
echo -e "\n[Check 1] Flat SPEC files (should be empty):"
find .moai/specs -maxdepth 1 -name "SPEC-*.md" -type f

# Check 2: Directories without required files
echo -e "\n[Check 2] SPEC directories missing required files:"
for dir in .moai/specs/SPEC-*/; do
  if [ -d "$dir" ]; then
    missing=""
    [ ! -f "${dir}spec.md" ] && missing="${missing}spec.md "
    [ ! -f "${dir}plan.md" ] && missing="${missing}plan.md "
    [ ! -f "${dir}acceptance.md" ] && missing="${missing}acceptance.md "
    [ -n "$missing" ] && echo "$dir: Missing $missing"
  fi
done

# Check 3: SPECs without numbers
echo -e "\n[Check 3] SPECs without proper numbering:"
ls -d .moai/specs/SPEC-*/ 2>/dev/null | grep -v -E 'SPEC-[A-Z]+-[0-9]{3}'

# Check 4: Potential reports in specs
echo -e "\n[Check 4] Potential reports in specs (check manually):"
grep -l -r "findings\|recommendations\|audit\|analysis" .moai/specs/*/spec.md 2>/dev/null

echo -e "\n=== Check Complete ==="
```

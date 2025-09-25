# moai_adk.core.docs.release_notes_converter

@FEATURE:RELEASE-NOTES-001 Release notes converter

Converts sync-report files into structured release notes.
Supports version-based organization and changelog generation.

@REQ:RELEASE-NOTES-001 → @TASK:RELEASE-NOTES-001

## Functions

### __init__

Initialize release notes converter

Args:
    project_root: Root directory of the project

```python
__init__(self, project_root)
```

### parse_sync_report

@TASK:RELEASE-NOTES-002 Parse sync-report.md file

```python
parse_sync_report(self)
```

### extract_version_info

@TASK:RELEASE-NOTES-003 Extract version information

```python
extract_version_info(self)
```

### categorize_changes

@TASK:RELEASE-NOTES-004 Categorize changes by TAG type

```python
categorize_changes(self, report_data)
```

### generate_changelog

@TASK:RELEASE-NOTES-005 Generate changelog format

```python
generate_changelog(self)
```

### get_version_timeline

@TASK:RELEASE-NOTES-006 Get sorted version timeline

```python
get_version_timeline(self)
```

### generate_release_notes

@TASK:RELEASE-NOTES-007 Generate release notes markdown file

```python
generate_release_notes(self, docs_dir)
```

### sort_key

```python
sort_key(version)
```

## Classes

### ReleaseNotesConverter

@FEATURE:RELEASE-NOTES-002 Convert sync reports to release notes

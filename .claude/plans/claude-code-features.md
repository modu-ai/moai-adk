# Claude Code Feature Parity Implementation Plan

**Date**: 2026-01-21
**Task**: Claude Code 공식 문서(settings, memory, statusline) 기반 MoAI-ADK 개선
**Priority**: Medium

---

## Executive Summary

Claude Code 공식 문서를 분석한 결과, MoAI-ADK에 다음 4가지 주요 기능을 추가/개선해야 합니다:

1. **Statusline Cost Tracking** - API 사용량 및 비용 표시
2. **Statusline Context Window 개선** - 백분율 정보 추가
3. **CLAUDE.md @path Import** - 외부 파일 import 문법 지원
4. **.claude/rules/ 디렉토리** - Path-specific rules 시스템

---

## Detailed Implementation Plan

### 1. Statusline Cost Tracking (Priority: High)

**Current State**:
- `extract_context_window()` 함수는 context_window 정보를 추출
- `cost` 객체는 session_context에 포함되지만 추출하지 않음

**Implementation**:

**File: `src/moai_adk/statusline/main.py`**

```python
def extract_cost_info(session_context: dict) -> dict:
    """
    Extract cost information from session context.

    Args:
        session_context: Context passed from Claude Code via stdin

    Returns:
        Dict with cost information (empty if not available)
    """
    cost_info = session_context.get("cost", {})
    if not cost_info:
        return {}

    return {
        "total_cost_usd": cost_info.get("total_cost_usd", 0.0),
        "total_duration_ms": cost_info.get("total_duration_ms", 0),
        "total_api_duration_ms": cost_info.get("total_api_duration_ms", 0),
        "total_lines_added": cost_info.get("total_lines_added", 0),
        "total_lines_removed": cost_info.get("total_lines_removed", 0),
    }
```

**File: `src/moai_adk/statusline/renderer.py`**

Update `StatuslineData` dataclass:
```python
@dataclass
class StatuslineData:
    # ... existing fields ...
    cost_total_usd: float = 0.0
    cost_lines_added: int = 0
    cost_lines_removed: int = 0
```

**File: `.moai/config/statusline-config.yaml`**

Add cost display configuration:
```yaml
display:
  cost: true  # 💰 API cost tracking
```

---

### 2. Statusline Context Window 개선 (Priority: High)

**Current State**:
- `extract_context_window()`는 토큰 수만 추출
- `used_percentage`, `remaining_percentage` 미사용

**Implementation**:

**File: `src/moai_adk/statusline/main.py`**

Update `extract_context_window()`:
```python
def extract_context_window(session_context: dict) -> dict:
    """
    Extract and format context window usage from session context.

    Returns dict with:
    - formatted: "15K/200K" string
    - used_percentage: 42.5 (from Claude Code)
    - remaining_percentage: 57.5 (from Claude Code)
    - current_usage: detailed breakdown
    """
    context_info = session_context.get("context_window", {})
    if not context_info:
        return {"formatted": "", "used_percentage": 0, "remaining_percentage": 100}

    # Use pre-calculated percentages from Claude Code
    used_pct = context_info.get("used_percentage", 0)
    remaining_pct = context_info.get("remaining_percentage", 100)

    # Format token usage
    current_tokens = context_info.get("total_input_tokens", 0)
    display_size = min(context_info.get("context_window_size", 200000), 200000)

    formatted = f"{format_token_count(current_tokens)}/{format_token_count(display_size)}"

    return {
        "formatted": formatted,
        "used_percentage": used_pct,
        "remaining_percentage": remaining_pct,
        "current_usage": context_info.get("current_usage", {}),
    }
```

---

### 3. CLAUDE.md @path Import System (Priority: High)

**Current State**:
- Template processor: `{{VARIABLE}}` 치환만 지원
- `@path/to/file` 문법 미구현

**Implementation**:

**File: `src/moai_adk/core/context_manager.py`** (new module)

```python
"""
CLAUDE.md Import Processor

Processes @path/to/file imports in CLAUDE.md files.

Based on Claude Code official documentation:
- Import syntax: @path/to/file or @~/absolute/path
- Recursive imports (max 5-depth)
- Ignored in code blocks/spans
- Both relative and absolute paths supported
"""

import re
from pathlib import Path
from typing import List, Set, Tuple

IMPORT_PATTERN = r'@([^\s\])`>@]+)'  # Matches @path but not in code
CODE_BLOCK_PATTERN = r'```[\s\S]*?```'  # Matches code blocks
INLINE_CODE_PATTERN = r'`[^`]+`'  # Matches inline code

class ClaudeMDImporter:
    """Process @path imports in CLAUDE.md files."""

    MAX_RECURSION_DEPTH = 5

    def __init__(self, base_path: Path):
        self.base_path = base_path.resolve()
        self.import_cache: dict[str, str] = {}
        self.recursion_stack: List[str] = []

    def process_imports(self, content: str, depth: int = 0) -> Tuple[str, List[str]]:
        """
        Process @path imports in CLAUDE.md content.

        Args:
            content: CLAUDE.md content
            depth: Current recursion depth

        Returns:
            Tuple of (processed_content, list_of_imported_files)
        """
        if depth >= self.MAX_RECURSION_DEPTH:
            return content, []

        # Track recursion to prevent cycles
        self.recursion_stack.append(str(self.base_path))

        imported_files = []

        # Remove code blocks (imports ignored in code)
        code_blocks = {}
        content = self._extract_and_replace_code_blocks(content, code_blocks)

        # Process imports
        lines = content.split('\n')
        processed_lines = []

        for line in lines:
            processed_line, imported = self._process_line(line, depth)
            processed_lines.append(processed_line)
            imported_files.extend(imported)

        # Restore code blocks
        content = '\n'.join(processed_lines)
        content = self._restore_code_blocks(content, code_blocks)

        self.recursion_stack.pop()
        return content, imported_files

    def _extract_and_replace_code_blocks(self, content: str, store: dict) -> str:
        """Extract code blocks and replace with placeholders."""
        idx = 0
        for match in re.finditer(CODE_BLOCK_PATTERN, content):
            placeholder = f"__CODE_BLOCK_{idx}__"
            store[placeholder] = match.group(0)
            content = content[:match.start()] + placeholder + content[match.end():]
            idx += 1
        return content

    def _restore_code_blocks(self, content: str, store: dict) -> str:
        """Restore code blocks from placeholders."""
        for placeholder, code_block in store.items():
            content = content.replace(placeholder, code_block)
        return content

    def _process_line(self, line: str, depth: int) -> Tuple[str, List[str]]:
        """Process a single line for imports."""
        imported_files = []
        processed_line = line

        for match in re.finditer(IMPORT_PATTERN, line):
            import_path = match.group(1)
            # Skip if in inline code (already replaced code blocks)
            if '`' in import_path:
                continue

            imported_content, import_files = self._load_import(import_path, depth)
            if imported_content:
                processed_line = processed_line.replace(match.group(0), imported_content)
                imported_files.extend(import_files)

        return processed_line, imported_files

    def _load_import(self, import_path: str, depth: int) -> Tuple[str, List[str]]:
        """Load and process an imported file."""
        # Resolve path
        if import_path.startswith('~/'):
            # User home directory import
            resolved_path = Path(import_path).expanduser()
        else:
            # Relative path from base_path
            resolved_path = self.base_path / import_path

        # Check cache
        cache_key = str(resolved_path)
        if cache_key in self.import_cache:
            return self.import_cache[cache_key], []

        # Check for circular imports
        if cache_key in self.recursion_stack:
            return f"[Circular import detected: {import_path}]", []

        # Read file
        if not resolved_path.exists():
            return f"[Import not found: {import_path}]", []

        try:
            content = resolved_path.read_text(encoding='utf-8', errors='replace')

            # Recursive processing
            importer = ClaudeMDImporter(resolved_path.parent)
            importer.recursion_stack = self.recursion_stack.copy()
            processed_content, nested_imports = importer.process_imports(content, depth + 1)

            # Cache result
            self.import_cache[cache_key] = processed_content

            return processed_content, [cache_key] + nested_imports

        except Exception as e:
            return f"[Import error: {import_path} - {e}]", []
```

**File: `src/moai_adk/core/template/processor.py`**

Update `_copy_claude_md()`:
```python
def _copy_claude_md(self, silent: bool = False) -> None:
    """Copy CLAUDE.md with import processing."""
    # ... existing code ...

    # Read template content
    content = src.read_text(encoding="utf-8", errors="replace")

    # Process @path imports
    from moai_adk.core.context_manager import ClaudeMDImporter
    importer = ClaudeMDImporter(self.template_root)
    processed_content, imported_files = importer.process_imports(content)

    # Apply variable substitution
    if self.context:
        processed_content, warnings = self._substitute_variables(processed_content)

    dst.write_text(processed_content, encoding="utf-8", errors="replace")
```

---

### 4. .claude/rules/ 디렉토리 시스템 (Priority: High)

**Current State**:
- `.claude/rules/` 디렉토리 구조 미존재
- Path-specific rules (YAML frontmatter) 미구현

**Implementation**:

**File: `src/moai_adk/core/rules_loader.py`** (new module)

```python
"""
.claude/rules/ Directory Loader

Loads path-specific rules from .claude/rules/*.md files.

Based on Claude Code official documentation:
- All .md files in .claude/rules/ are automatically loaded
- YAML frontmatter with `paths` field for path-specific rules
- Glob patterns supported: **/*.ts, src/**/*.{ts,tsx}
- Subdirectories supported
- Symlinks supported
"""

import re
import yaml
from pathlib import Path
from typing import List, Optional, Tuple
from dataclasses import dataclass

@dataclass
class Rule:
    """A rule loaded from .claude/rules/."""
    name: str
    content: str
    paths: Optional[List[str]]  # None = applies to all files
    source_file: Path

class RulesLoader:
    """Load rules from .claude/rules/ directory."""

    def __init__(self, project_path: Path):
        self.project_path = project_path.resolve()
        self.rules_dir = self.project_path / ".claude" / "rules"
        self.user_rules_dir = Path.home() / ".claude" / "rules"

    def load_rules(self, file_path: Optional[Path] = None) -> List[Rule]:
        """
        Load rules applicable to a specific file.

        Args:
            file_path: File to check for path-specific rules

        Returns:
            List of applicable rules (user rules first, then project rules)
        """
        rules = []

        # Load user rules (~/.claude/rules/)
        rules.extend(self._load_rules_from_dir(self.user_rules_dir, file_path))

        # Load project rules (.claude/rules/)
        if self.rules_dir.exists():
            rules.extend(self._load_rules_from_dir(self.rules_dir, file_path))

        return rules

    def _load_rules_from_dir(self, rules_dir: Path, file_path: Optional[Path]) -> List[Rule]:
        """Load all rules from a directory."""
        rules = []

        # Find all .md files (including subdirectories)
        for md_file in rules_dir.rglob("*.md"):
            # Skip symlinks that create circular references
            if md_file.is_symlink():
                try:
                    resolved = md_file.resolve()
                    # Check if this is within the rules directory (prevent loops)
                    resolved.relative_to(rules_dir)
                except (ValueError, RuntimeError):
                    continue

            rule = self._parse_rule_file(md_file)
            if rule:
                # Check if rule applies to this file
                if file_path is None or self._rule_applies(rule, file_path):
                    rules.append(rule)

        return rules

    def _parse_rule_file(self, file_path: Path) -> Optional[Rule]:
        """Parse a rule file with YAML frontmatter."""
        try:
            content = file_path.read_text(encoding="utf-8", errors="replace")

            # Extract YAML frontmatter
            frontmatter_match = re.match(r'^---\n(.*?)\n---', content, re.DOTALL)
            paths = None

            if frontmatter_match:
                yaml_content = frontmatter_match.group(1)
                try:
                    frontmatter_data = yaml.safe_load(yaml_content)
                    if frontmatter_data and isinstance(frontmatter_data, dict):
                        paths = frontmatter_data.get('paths')
                        if paths and not isinstance(paths, list):
                            paths = None
                except yaml.YAMLError:
                    pass

            # Remove frontmatter from content
            if frontmatter_match:
                content = content[frontmatter_match.end():].lstrip()

            return Rule(
                name=file_path.stem,
                content=content,
                paths=paths,
                source_file=file_path
            )

        except Exception:
            return None

    def _rule_applies(self, rule: Rule, file_path: Path) -> bool:
        """Check if a rule applies to a specific file."""
        if rule.paths is None:
            return True  # Applies to all files

        # Convert file_path to relative path from project root
        try:
            relative_path = file_path.relative_to(self.project_path)
            path_str = str(relative_path).replace('\\', '/')
        except ValueError:
            # File is outside project, use absolute path
            path_str = str(file_path).replace('\\', '/')

        # Check each pattern
        for pattern in rule.paths:
            if self._matches_pattern(pattern, path_str):
                return True

        return False

    def _matches_pattern(self, pattern: str, path: str) -> bool:
        """Check if a path matches a glob pattern."""
        # Convert glob pattern to regex
        regex_pattern = pattern.replace('*', '.*').replace('?', '.')
        regex_pattern = f'^{regex_pattern}$'

        # Handle brace expansion: {src,lib}/**/*.ts
        if '{' in pattern:
            # Simple brace expansion
            base, rest = pattern.split('{', 1)
            options_part, suffix = rest.split('}', 1)
            options = options_part.split(',')

            for option in options:
                expanded_pattern = f"{base}{option}{suffix}"
                if self._matches_pattern(expanded_pattern, path):
                    return True
            return False

        return re.match(regex_pattern, path) is not None
```

**Integration with CLAUDE.md loading:**

Rules should be automatically included when CLAUDE.md is loaded. Update the context loader to include rules:

```python
def load_context_with_rules(project_path: Path) -> dict:
    """Load CLAUDE.md with .claude/rules/ included."""
    context = {}

    # Load CLAUDE.md
    claude_md = project_path / "CLAUDE.md"
    if claude_md.exists():
        # Process imports
        from moai_adk.core.context_manager import ClaudeMDImporter
        importer = ClaudeMDImporter(project_path)
        content, _ = importer.process_imports(claude_md.read_text())
        context["claude_md"] = content

    # Load applicable rules
    from moai_adk.core.rules_loader import RulesLoader
    rules_loader = RulesLoader(project_path)
    rules = rules_loader.load_rules()  # No file_path = load all rules

    context["rules"] = "\n\n".join([rule.content for rule in rules])

    return context
```

---

## Execution Order

1. **Phase 1: Statusline improvements** (Lower risk, visible benefit)
   - Cost tracking
   - Context window percentage display

2. **Phase 2: CLAUDE.md @path import** (Medium risk, high value)
   - Import processor
   - Template integration

3. **Phase 3: .claude/rules/ directory** (Higher complexity)
   - Rules loader
   - Path-specific matching
   - Integration with context loader

---

## Files to Create

- `src/moai_adk/core/context_manager.py` - CLAUDE.md import processor
- `src/moai_adk/core/rules_loader.py` - Rules directory loader

## Files to Modify

- `src/moai_adk/statusline/main.py` - Cost extraction, context window improvements
- `src/moai_adk/statusline/renderer.py` - Cost display
- `src/moai_adk/core/template/processor.py` - Import processing integration
- `.moai/config/statusline-config.yaml` - Cost display configuration

---

## Testing Strategy

1. **Unit Tests**:
   - `test_import_processor.py` - @path import logic
   - `test_rules_loader.py` - Path-specific rules
   - `test_cost_extraction.py` - Cost parsing

2. **Integration Tests**:
   - Test CLAUDE.md with imports
   - Test .claude/rules/ with various patterns

3. **Manual Tests**:
   - Create sample CLAUDE.md with @path imports
   - Create .claude/rules/ with path-specific rules
   - Verify statusline displays cost and context info

---

## Rollout Plan

1. **Branch**: `feature/claude-code-parity`
2. **PR Description**: Detail all 4 improvements with examples
3. **Migration Guide**: Document new features for users

---

## Completion Criteria

- [ ] Statusline displays cost information (total_cost_usd, lines added/removed)
- [ ] Statusline displays context window percentage
- [ ] CLAUDE.md supports @path/to/file imports
- [ ] .claude/rules/ directory with path-specific rules works
- [ ] All tests passing
- [ ] Documentation updated

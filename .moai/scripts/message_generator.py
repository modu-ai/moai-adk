#!/usr/bin/env python3
"""
Message Generator Module for MoAI Commit Helper
Handles automatic commit message generation with smart analysis

@TASK:MESSAGE-GEN-001
@FEATURE:AUTO-MESSAGE-001
@API:GENERATE-MESSAGE
"""

from typing import Any


class MessageGenerator:
    """Smart commit message generator following conventional commits"""

    def generate_smart_message(self, files: list[dict[str, Any]]) -> str:
        """Generate smart commit message based on file changes"""
        if not files:
            return "🔧 Minor updates"

        # Classify changes by type
        added = [f for f in files if f["type"] == "added"]
        modified = [f for f in files if f["type"] == "modified"]
        deleted = [f for f in files if f["type"] == "deleted"]

        # Classify by file extensions
        py_files = [f for f in files if f["filename"].endswith(".py")]
        md_files = [f for f in files if f["filename"].endswith(".md")]
        json_files = [f for f in files if f["filename"].endswith(".json")]

        return self._generate_message_by_pattern(
            added, modified, deleted, py_files, md_files, json_files
        )

    def generate_context_message(self, context: str, files: list[dict[str, Any]]) -> str:
        """Generate context-based commit message"""
        context_lower = context.lower()

        if "fix" in context_lower or "bug" in context_lower:
            return f"🐛 Fix: {context}"
        elif "feat" in context_lower or "feature" in context_lower:
            return f"✨ Feature: {context}"
        elif "test" in context_lower:
            return f"🧪 Test: {context}"
        elif "doc" in context_lower:
            return f"📚 Docs: {context}"
        elif "refactor" in context_lower:
            return f"♻️ Refactor: {context}"
        else:
            return f"🔧 {context}"

    def generate_template_suggestions(self) -> list[dict[str, Any]]:
        """Generate template-based message suggestions"""
        return [
            {"type": "feature", "message": "✨ feat: ", "confidence": 0.6},
            {"type": "fix", "message": "🐛 fix: ", "confidence": 0.6},
            {"type": "docs", "message": "📚 docs: ", "confidence": 0.6},
            {"type": "refactor", "message": "♻️ refactor: ", "confidence": 0.6},
            {"type": "test", "message": "🧪 test: ", "confidence": 0.6},
            {"type": "chore", "message": "🔧 chore: ", "confidence": 0.6},
        ]

    def calculate_confidence(self, files: list[dict[str, Any]]) -> float:
        """Calculate confidence score for message suggestions"""
        if not files:
            return 0.0

        # High confidence for single file changes
        if len(files) == 1:
            return 0.9

        # High confidence for same type files
        file_types = set(
            f["filename"].split(".")[-1] for f in files if "." in f["filename"]
        )
        if len(file_types) == 1:
            return 0.8

        # Medium confidence for mixed changes
        return 0.6

    def _generate_message_by_pattern(
        self,
        added: list[dict[str, Any]],
        modified: list[dict[str, Any]],
        deleted: list[dict[str, Any]],
        py_files: list[dict[str, Any]],
        md_files: list[dict[str, Any]],
        json_files: list[dict[str, Any]],
    ) -> str:
        """Generate message based on file change patterns"""
        # Only additions
        if len(added) > 0 and len(modified) == 0 and len(deleted) == 0:
            return self._handle_additions(added)

        # Only modifications
        elif len(modified) > 0 and len(added) == 0 and len(deleted) == 0:
            return self._handle_modifications(py_files, md_files, json_files)

        # Only deletions
        elif len(deleted) > 0:
            return f"🗑️ Remove {len(deleted)} files"

        # Mixed changes
        else:
            return self._handle_mixed_changes(added, modified, deleted)

    def _handle_additions(self, added: list[dict[str, Any]]) -> str:
        """Handle addition-only changes"""
        if len(added) == 1:
            return f"✨ Add {added[0]['filename']}"
        else:
            return f"✨ Add {len(added)} new files"

    def _handle_modifications(
        self,
        py_files: list[dict[str, Any]],
        md_files: list[dict[str, Any]],
        json_files: list[dict[str, Any]],
    ) -> str:
        """Handle modification-only changes"""
        if len(py_files) > len(md_files):
            return "🔧 Update Python modules"
        elif len(md_files) > 0:
            return "📚 Update documentation"
        elif len(json_files) > 0:
            return "🔧 Update configuration"
        else:
            return "🔧 Update files"

    def _handle_mixed_changes(
        self,
        added: list[dict[str, Any]],
        modified: list[dict[str, Any]],
        deleted: list[dict[str, Any]],
    ) -> str:
        """Handle mixed change types"""
        total = len(added) + len(modified) + len(deleted)
        if total <= 3:
            return f"🔧 Update {total} files"
        else:
            return f"♻️ Refactor multiple files ({total} files)"
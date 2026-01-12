# type: ignore
#!/usr/bin/env python3
"""
Enhanced Output Style Detector for Claude Code Statusline

This module provides real-time detection of Claude Code's current output style
by analyzing session context, environment variables, and behavioral indicators.

Key improvements:
1. Real-time session context analysis
2. Multiple detection methods with priority ordering
3. Behavioral pattern recognition
4. Robust error handling and graceful degradation
"""

import json
import os
import sys
import time
from pathlib import Path
from typing import Any, Dict, Optional


class OutputStyleDetector:
    """
    Enhanced output style detector with multiple detection strategies.

    Detection Priority:
    1. Session context (most reliable)
    2. Environment variables
    3. Behavioral analysis
    4. Settings file (fallback)
    5. Heuristics (last resort)
    """

    # Style mapping for consistent display format
    STYLE_MAPPING = {
        # Internal style names
        "streaming": "R2-D2",
        "explanatory": "Explanatory",
        "concise": "Concise",
        "detailed": "Detailed",
        "yoda": "🧙 Yoda Master",
        "yoda-master": "🧙 Yoda Master",
        "tutorial": "🧙 Yoda Master",
        # Display values with emojis
        "🤖 R2-D2": "R2-D2",
        "R2-D2": "R2-D2",
        "🧙 Explanatory": "Explanatory",
        "Explanatory": "Explanatory",
        "🧙 Concise": "Concise",
        "Concise": "Concise",
        "🧙 Detailed": "Detailed",
        "Detailed": "Detailed",
        "🧙 Yoda Master": "🧙 Yoda Master",
        "Yoda Master": "🧙 Yoda Master",
    }

    def __init__(self):
        self.cache = {}
        self.cache_ttl: float = 5.0  # Cache for 5 seconds to balance performance and accuracy

    def detect_from_session_context(self, session_data: Dict[str, Any]) -> Optional[str]:
        """
        Detect output style from Claude Code session context.

        This is the most reliable method as it uses real-time session data.
        """
        try:
            # Method 1: Check explicit outputStyle in session
            if "outputStyle" in session_data:
                style = session_data["outputStyle"]
                if style:
                    return self._normalize_style(style)

            # Method 2: Check model configuration for style indicators
            model_info = session_data.get("model", {})
            if isinstance(model_info, dict):
                model_name = model_info.get("name", "").lower()
                display_name = model_info.get("display_name", "").lower()

                # Check for style indicators in model names
                for text in [model_name, display_name]:
                    if "explanatory" in text:
                        return "Explanatory"
                    elif "yoda" in text or "tutorial" in text:
                        return "🧙 Yoda Master"
                    elif "concise" in text:
                        return "Concise"
                    elif "detailed" in text:
                        return "Detailed"
                    elif "streaming" in text or "r2d2" in text:
                        return "R2-D2"

            # Method 3: Check conversation patterns
            messages = session_data.get("messages", [])
            if messages:
                # Analyze recent message patterns for style indicators
                recent_messages = messages[-3:]  # Last 3 messages
                return self._analyze_message_patterns(recent_messages)

        except Exception as e:
            # Log error but don't fail - continue to next detection method
            print(f"Session context detection error: {e}", file=sys.stderr)

        return None

    def detect_from_environment(self) -> Optional[str]:
        """
        Detect output style from environment variables.
        """
        try:
            # Check for explicit environment variable
            if "CLAUDE_OUTPUT_STYLE" in os.environ:
                env_style = os.environ["CLAUDE_OUTPUT_STYLE"]
                if env_style:
                    return self._normalize_style(env_style)

            # Check for Claude Code session indicators
            if "CLAUDE_SESSION_ID" in os.environ:
                # Could integrate with Claude Code session management
                pass

            # Check for process name or command line indicators
            try:
                # This could be enhanced with process inspection if needed
                pass
            except Exception:
                pass

        except Exception as e:
            print(f"Environment detection error: {e}", file=sys.stderr)

        return None

    def detect_from_behavioral_analysis(self) -> Optional[str]:
        """
        Analyze behavioral patterns to infer current output style.

        This method uses heuristics based on file system state and recent activity.
        """
        try:
            cwd = Path.cwd()

            # Check for active Yoda session indicators
            moai_dir = cwd / ".moai"
            if moai_dir.exists():
                # Look for recent Yoda-related activity
                yoda_files = list(moai_dir.rglob("*yoda*"))
                if yoda_files:
                    # Check if any Yoda files are recently modified
                    recent_yoda = any(f.stat().st_mtime > (time.time() - 300) for f in yoda_files)  # Last 5 minutes
                    if recent_yoda:
                        return "🧙 Yoda Master"

            # Check for extensive documentation (might indicate Explanatory mode)
            docs_dir = cwd / "docs"
            if docs_dir.exists():
                md_files = list(docs_dir.rglob("*.md"))
                if len(md_files) > 10:  # Heuristic threshold
                    return "Explanatory"

            # Check for TODO/task tracking patterns
            todo_file = cwd / ".moai" / "current_session_todo.txt"
            if todo_file.exists():
                content = todo_file.read_text()
                if "plan" in content.lower() or "phase" in content.lower():
                    return "Explanatory"

        except Exception as e:
            print(f"Behavioral analysis error: {e}", file=sys.stderr)

        return None

    def detect_from_settings(self) -> Optional[str]:
        """
        Detect output style from settings.json file.

        This is the least reliable method as it may not reflect current session.
        """
        try:
            settings_path = Path.cwd() / ".claude" / "settings.json"
            if settings_path.exists():
                with open(settings_path, "r", encoding="utf-8") as f:
                    settings = json.load(f)
                    output_style = settings.get("outputStyle", "")

                    if output_style:
                        return self._normalize_style(output_style)

        except Exception as e:
            print(f"Settings file detection error: {e}", file=sys.stderr)

        return None

    def _normalize_style(self, style: str) -> str:
        """
        Normalize style name to consistent display format.
        """
        if not style:
            return "Unknown"

        # Direct mapping lookup
        if style in self.STYLE_MAPPING:
            return self.STYLE_MAPPING[style]

        # Case-insensitive lookup
        style_lower = style.lower()
        for key, value in self.STYLE_MAPPING.items():
            if key.lower() == style_lower:
                return value

        # Pattern-based normalization
        if "r2d2" in style_lower or "streaming" in style_lower:
            return "R2-D2"
        elif "yoda" in style_lower or "master" in style_lower:
            return "🧙 Yoda Master"
        elif "explanatory" in style_lower:
            return "Explanatory"
        elif "concise" in style_lower:
            return "Concise"
        elif "detailed" in style_lower:
            return "Detailed"

        # Extract from emoji-prefixed values
        if "🤖" in style:
            return "R2-D2"
        elif "🧙" in style:
            # Extract the part after emoji
            parts = style.split("🧙", 1)
            if len(parts) > 1:
                extracted = parts[1].strip()
                return self._normalize_style(extracted)

        # Fallback: capitalize first letter
        return style.title() if style else "Unknown"

    def _analyze_message_patterns(self, messages: list) -> Optional[str]:
        """
        Analyze recent message patterns for style indicators.
        """
        try:
            if not messages:
                return None

            # Look for style indicators in recent responses
            full_text = " ".join(msg.get("content", "") for msg in messages[-3:] if msg.get("role") == "assistant")

            if not full_text:
                return None

            # Style heuristics based on response patterns
            text_lower = full_text.lower()

            # Yoda Master indicators
            yoda_indicators = [
                "young padawan",
                "the force",
                "master",
                "wisdom",
                "patience",
            ]
            yoda_count = sum(1 for indicator in yoda_indicators if indicator in text_lower)

            if yoda_count >= 2:
                return "🧙 Yoda Master"

            # Explanatory indicators
            if len(full_text) > 2000:  # Long responses
                explanatory_count = sum(
                    1
                    for phrase in [
                        "let me explain",
                        "here's how",
                        "the reason is",
                        "to understand",
                    ]
                    if phrase in text_lower
                )
                if explanatory_count >= 2:
                    return "Explanatory"

            # Concise indicators
            if len(full_text) < 500:
                return "Concise"

            # Default fallback
            return "R2-D2"

        except Exception as e:
            print(f"Message pattern analysis error: {e}", file=sys.stderr)
            return None

    def get_output_style(self, session_context: Optional[Dict[str, Any]] = None) -> str:
        """
        Get the current output style using all available detection methods.

        Args:
            session_context: Optional session context from Claude Code

        Returns:
            Normalized output style string
        """
        # Use cache if available and fresh
        cache_key = f"{id(session_context)}_{hash(str(session_context))}"
        current_time = time.time()

        if cache_key in self.cache:
            cached_style, cached_time = self.cache[cache_key]
            if current_time - cached_time < self.cache_ttl:
                return cached_style

        # Detection methods in priority order
        detection_methods = [
            (
                "Session Context",
                lambda: self.detect_from_session_context(session_context or {}),
            ),
            ("Environment", self.detect_from_environment),
            ("Behavioral Analysis", self.detect_from_behavioral_analysis),
            ("Settings File", self.detect_from_settings),
        ]

        for method_name, method_func in detection_methods:
            try:
                style = method_func()
                if style and style != "Unknown":
                    # Cache the result
                    self.cache[cache_key] = (style, current_time)
                    return style
            except Exception as e:
                print(f"{method_name} detection failed: {e}", file=sys.stderr)
                continue

        # Default fallback - Mr. Alfred is the default output style
        return "Mr. Alfred"


def safe_collect_output_style() -> str:
    """
    Legacy compatibility function that maintains the original interface.

    This function provides backward compatibility with the existing statusline
    system while using the enhanced detection capabilities.

    Returns:
        Detected output style string
    """
    try:
        # Read session context from stdin if available
        session_context = {}
        try:
            # Handle Docker/non-interactive environments by checking TTY
            input_data = sys.stdin.read() if not sys.stdin.isatty() else "{}"
            if input_data.strip():
                session_context = json.loads(input_data)
        except (json.JSONDecodeError, EOFError):
            pass

        # Use the enhanced detector
        detector = OutputStyleDetector()
        return detector.get_output_style(session_context)

    except Exception as e:
        print(f"Output style detection failed: {e}", file=sys.stderr)
        return "Mr. Alfred"


# For backward compatibility
if __name__ == "__main__":
    # If run as a script, output the detected style
    print(safe_collect_output_style())

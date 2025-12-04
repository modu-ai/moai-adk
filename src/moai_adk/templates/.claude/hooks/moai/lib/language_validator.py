#!/usr/bin/env python3
"""Language Configuration Validator for MoAI-ADK

Validates language configuration settings and ensures proper functionality.
Provides detailed diagnostics for language-related issues.
"""

import json
import sys
from pathlib import Path
from typing import Any, Dict, List


def load_config() -> Dict[str, Any]:
    """Load configuration from .moai/config/config.json

    Returns:
        Configuration dictionary or empty dict if not found
    """
    config_path = Path(".moai/config/config.json")

    if not config_path.exists():
        return {"error": "Configuration file not found"}

    try:
        with open(config_path, "r", encoding="utf-8") as f:
            return json.load(f)
    except json.JSONDecodeError as e:
        return {"error": f"JSON decode error: {e}"}
    except Exception as e:
        return {"error": f"File read error: {e}"}


def validate_language_config(config: Dict[str, Any]) -> Dict[str, Any]:
    """Validate language configuration

    Args:
        config: Configuration dictionary

    Returns:
        Validation result with status and details
    """
    warnings: List[str] = []
    errors: List[str] = []
    result: Dict[str, Any] = {
        "valid": True,
        "warnings": warnings,
        "errors": errors,
        "language_info": {},
    }

    # Check if language section exists
    if "language" not in config:
        result["valid"] = False
        errors.append("Missing 'language' configuration section")
        return result

    lang_config = config.get("language", {})

    # Validate conversation_language
    conversation_lang = lang_config.get("conversation_language", "en")
    if not conversation_lang:
        result["valid"] = False
        errors.append("conversation_language is empty or missing")
    elif conversation_lang not in ["ko", "en", "ja", "zh"]:
        warnings.append(
            f"conversation_language '{conversation_lang}' is not officially supported (supported: ko, en, ja, zh)"
        )

    # Validate conversation_language_name
    lang_name = lang_config.get("conversation_language_name", "")
    if not lang_name:
        result["valid"] = False
        errors.append("conversation_language_name is empty or missing")
    elif conversation_lang == "ko" and lang_name != "Korean":
        warnings.append(f"conversation_language_name '{lang_name}' doesn't match 'Korean' for Korean language")
    elif conversation_lang == "en" and lang_name != "English":
        warnings.append(f"conversation_language_name '{lang_name}' doesn't match 'English' for English language")
    elif conversation_lang == "ja" and lang_name != "Japanese":
        warnings.append(f"conversation_language_name '{lang_name}' doesn't match 'Japanese' for Japanese language")
    elif conversation_lang == "zh" and lang_name != "Chinese":
        warnings.append(f"conversation_language_name '{lang_name}' doesn't match 'Chinese' for Chinese language")

    # Validate agent_prompt_language
    agent_lang = lang_config.get("agent_prompt_language", conversation_lang)
    if agent_lang != conversation_lang:
        warnings.append(
            f"agent_prompt_language '{agent_lang}' differs from conversation_language '{conversation_lang}'"
        )

    # Store language info
    result["language_info"] = {
        "conversation_language": conversation_lang,
        "conversation_language_name": lang_name,
        "agent_prompt_language": agent_lang,
        "primary_language": conversation_lang,
        "is_korean": conversation_lang == "ko",
        "is_english": conversation_lang == "en",
        "is_japanese": conversation_lang == "ja",
        "is_chinese": conversation_lang == "zh",
        "is_multilingual": conversation_lang in ["ko", "ja", "zh"],
    }

    return result


def validate_output_style_compatibility() -> Dict[str, Any]:
    """Validate if output style supports language configuration

    Returns:
        Compatibility check results
    """
    recommendations: List[str] = []
    result: Dict[str, Any] = {
        "r2d2_exists": False,
        "language_support_present": False,
        "config_reading_present": False,
        "recommendations": recommendations,
    }

    # Check if R2-D2 output style exists
    r2d2_path = Path(".claude/output-styles/moai/r2d2.md")
    result["r2d2_exists"] = r2d2_path.exists()

    if r2d2_path.exists():
        try:
            with open(r2d2_path, "r", encoding="utf-8") as f:
                content = f.read()

            # Check for language support indicators
            language_keywords = [
                "conversation_language",
                "language.conversation_language",
                "한국어",
                "Korean",
                "Language Configuration",
            ]

            result["language_support_present"] = any(keyword in content for keyword in language_keywords)
            result["config_reading_present"] = ".moai/config/config.json" in content

            if not result["language_support_present"]:
                recommendations.append("Add language support documentation to R2-D2 output style")

            if not result["config_reading_present"]:
                recommendations.append("Add config.json reading instructions to R2-D2 output style")

        except Exception as e:
            result["error"] = f"Error reading R2-D2 output style: {e}"
    else:
        recommendations.append("R2-D2 output style not found - ensure proper installation")

    return result


def validate_session_start_hook() -> Dict[str, Any]:
    """Validate if SessionStart hook displays language info

    Returns:
        Hook validation results
    """
    recommendations: List[str] = []
    result: Dict[str, Any] = {
        "hook_exists": False,
        "language_display_present": False,
        "recommendations": recommendations,
    }

    hook_path = Path(".claude/hooks/moai/session_start__show_project_info.py")
    result["hook_exists"] = hook_path.exists()

    if hook_path.exists():
        try:
            with open(hook_path, "r", encoding="utf-8") as f:
                content = f.read()

            # Check for language display functionality
            language_indicators = [
                "get_language_info",
                "conversation_language",
                "language_name",
                "Language:",
            ]

            result["language_display_present"] = any(indicator in content for indicator in language_indicators)

            if not result["language_display_present"]:
                recommendations.append("Add language info display to SessionStart hook")

        except Exception as e:
            result["error"] = f"Error reading SessionStart hook: {e}"
    else:
        recommendations.append("SessionStart hook not found - ensure proper installation")

    return result


def generate_validation_report() -> str:
    """Generate comprehensive language validation report

    Returns:
        Formatted validation report string
    """
    report_lines = ["🌐 Language Configuration Validation Report", "=" * 50]

    # Load and validate config
    config = load_config()
    if "error" in config:
        report_lines.append(f"❌ Config Error: {config['error']}")
        return "\n".join(report_lines)

    validation = validate_language_config(config)
    language_info = validation["language_info"]

    # Current configuration status
    report_lines.extend(
        [
            "",
            "📋 Current Configuration:",
            f"  • Language: {language_info.get('conversation_language_name', 'Unknown')} "
            f"({language_info.get('conversation_language', 'Unknown')})",
            f"  • Agent Language: {language_info.get('agent_prompt_language', 'Unknown')}",
            f"  • Korean Mode: {'✅ Active' if language_info.get('is_korean') else '❌ Inactive'}",
            f"  • Config Valid: {'✅ Valid' if validation['valid'] else '❌ Invalid'}",
        ]
    )

    # Errors and warnings
    if validation["errors"]:
        report_lines.extend(["", "❌ Errors:"])
        for error in validation["errors"]:
            report_lines.append(f"  • {error}")

    if validation["warnings"]:
        report_lines.extend(["", "⚠️ Warnings:"])
        for warning in validation["warnings"]:
            report_lines.append(f"  • {warning}")

    # Output style compatibility
    report_lines.extend(["", "🔧 Output Style Compatibility:"])
    output_check = validate_output_style_compatibility()
    report_lines.extend(
        [
            f"  • R2-D2 Output Style: {'✅ Found' if output_check['r2d2_exists'] else '❌ Missing'}",
            f"  • Language Support: {'✅ Present' if output_check['language_support_present'] else '❌ Missing'}",
            f"  • Config Reading: {'✅ Present' if output_check['config_reading_present'] else '❌ Missing'}",
        ]
    )

    if output_check.get("recommendations"):
        report_lines.extend(["", "  Recommendations:"])
        for rec in output_check["recommendations"]:
            report_lines.append(f"    • {rec}")

    # SessionStart hook check
    report_lines.extend(["", "🎣 SessionStart Hook:"])
    hook_check = validate_session_start_hook()
    report_lines.extend(
        [
            f"  • Hook Exists: {'✅ Found' if hook_check['hook_exists'] else '❌ Missing'}",
            f"  • Language Display: {'✅ Present' if hook_check['language_display_present'] else '❌ Missing'}",
        ]
    )

    if hook_check.get("recommendations"):
        report_lines.extend(["", "  Recommendations:"])
        for rec in hook_check["recommendations"]:
            report_lines.append(f"    • {rec}")

    # Overall status
    report_lines.extend(["", "📊 Overall Status:"])
    all_good = (
        validation["valid"]
        and not validation["errors"]
        and output_check["r2d2_exists"]
        and output_check["language_support_present"]
        and hook_check["hook_exists"]
        and hook_check["language_display_present"]
    )

    if all_good:
        report_lines.append("✅ Language configuration is fully functional!")
        if language_info.get("is_korean"):
            report_lines.append("🇰🇷 Korean language responses should work correctly.")
        else:
            report_lines.append("🌐 English language responses configured.")
    else:
        report_lines.append("⚠️ Issues found that may prevent proper language functionality.")
        report_lines.append("   See recommendations above for fixes.")

    return "\n".join(report_lines)


def main() -> None:
    """Main entry point for language configuration validator"""
    try:
        report = generate_validation_report()
        print(report)

        # Exit with appropriate code
        config = load_config()
        if "error" in config:
            sys.exit(2)  # Config error
        else:
            validation = validate_language_config(config)
            if not validation["valid"]:
                sys.exit(1)  # Config invalid
            else:
                sys.exit(0)  # Success

    except Exception as e:
        print(f"❌ Language validation error: {e}")
        sys.exit(3)


if __name__ == "__main__":
    main()

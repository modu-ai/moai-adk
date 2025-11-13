#!/usr/bin/env python3
"""Agent Context Engineering utilities

Advanced JIT (Just-in-Time) Retrieval with Expert Agent Delegation
Intelligently analyzes user prompts to recommend specialist agents and skills
"""

import json
import re
from pathlib import Path
from typing import Dict, List, Optional, Tuple, Any

from .context import get_jit_context


def load_agent_skills_mapping() -> Dict[str, Any]:
    """Load agent-skills mapping configuration

    Returns:
        Mapping configuration dictionary
    """
    try:
        mapping_file = Path(__file__).parent.parent / "config" / "agent_skills_mapping.json"
        if mapping_file.exists():
            with open(mapping_file, 'r', encoding='utf-8') as f:
                return json.load(f)
    except Exception:
        pass

    return {
        "agent_skills_mapping": {},
        "prompt_patterns": {}
    }


def analyze_prompt_intent(prompt: str, mapping: Dict[str, Any]) -> Optional[Dict[str, Any]]:
    """Analyze user prompt intent

    Args:
        prompt: User prompt
        mapping: Agent-skills mapping

    Returns:
        Analyzed intent information or None
    """
    prompt_lower = prompt.lower()

    # Calculate score for each pattern
    pattern_scores = []

    for pattern_name, pattern_config in mapping.get("prompt_patterns", {}).items():
        score = 0
        matched_keywords = []

        # Keyword matching
        for keyword in pattern_config.get("keywords", []):
            if keyword.lower() in prompt_lower:
                score += 1
                matched_keywords.append(keyword)

        # Regex pattern matching
        if pattern_config.get("regex_patterns"):
            for regex_pattern in pattern_config["regex_patterns"]:
                if re.search(regex_pattern, prompt_lower):
                    score += 2  # Regex matching has higher weight

        if score > 0:
            pattern_scores.append({
                "pattern": pattern_name,
                "score": score,
                "matched_keywords": matched_keywords,
                "config": pattern_config
            })

    # Return pattern with highest score
    if pattern_scores:
        pattern_scores.sort(key=lambda x: x["score"], reverse=True)
        best_match = pattern_scores[0]

        return {
            "intent": best_match["pattern"],
            "confidence": min(best_match["score"] / 3.0, 1.0),  # Maximum 3 keyword matches
            "matched_keywords": best_match["matched_keywords"],
            "primary_agent": best_match["config"].get("primary_agent"),
            "secondary_agents": best_match["config"].get("secondary_agents", []),
            "recommended_skills": best_match["config"].get("skills", []),
            "context_files": best_match["config"].get("context_files", [])
        }

    return None


def get_agent_delegation_context(prompt: str, cwd: str) -> Dict[str, Any]:
    """Create agent delegation context based on prompt

    Args:
        prompt: User prompt
        cwd: Current working directory

    Returns:
        Agent delegation context information
    """
    mapping = load_agent_skills_mapping()
    intent_analysis = analyze_prompt_intent(prompt, mapping)

    # Get existing JIT context
    existing_context = get_jit_context(prompt, cwd)

    # Agent delegation information
    agent_context = {
        "intent_detected": intent_analysis is not None,
        "traditional_context": existing_context
    }

    if intent_analysis:
        # Verify file existence
        valid_context_files = []
        cwd_path = Path(cwd)

        for context_file in intent_analysis["context_files"]:
            file_path = cwd_path / context_file
            if file_path.exists():
                valid_context_files.append(context_file)

        # Create skills reference paths
        skill_references = []
        for skill in intent_analysis["recommended_skills"]:
            skill_ref = f".claude/skills/{skill}/reference.md"
            skill_path = cwd_path / skill_ref
            if skill_path.exists():
                skill_references.append(skill_ref)

        agent_context.update({
            "primary_agent": intent_analysis["primary_agent"],
            "secondary_agents": intent_analysis["secondary_agents"],
            "recommended_skills": intent_analysis["recommended_skills"],
            "skill_references": skill_references,
            "context_files": valid_context_files,
            "confidence": intent_analysis["confidence"],
            "intent": intent_analysis["intent"],
            "matched_keywords": intent_analysis["matched_keywords"]
        })

    return agent_context


def format_agent_delegation_message(context: Dict[str, Any]) -> Optional[str]:
    """Format agent delegation message

    Args:
        context: Agent delegation context

    Returns:
        Formatted message or None
    """
    if not context.get("intent_detected"):
        return None

    messages = []

    # Basic information
    primary_agent = context.get("primary_agent")
    confidence = context.get("confidence", 0)
    intent = context.get("intent", "")
    matched_keywords = context.get("matched_keywords", [])

    if primary_agent and confidence > 0.5:
        messages.append(f"🎯 Recommended Expert Agent: {primary_agent}")
        messages.append(f"📋 Task Intent: {intent}")

        if matched_keywords:
            messages.append(f"🔍 Detected Keywords: {', '.join(matched_keywords)}")

        # Recommended skills
        skills = context.get("recommended_skills", [])
        if skills:
            messages.append(f"⚡ Recommended Skills: {', '.join(skills[:3])}")  # Display max 3

        # Secondary agents
        secondary_agents = context.get("secondary_agents", [])
        if secondary_agents:
            messages.append(f"🤝 Collaboration Agents: {', '.join(secondary_agents[:2])}")  # Display max 2

        # Context files
        context_files = context.get("context_files", [])
        skill_references = context.get("skill_references", [])

        all_files = context_files + skill_references
        if all_files:
            messages.append(f"📚 Auto-loaded Context: {len(all_files)} files")

    return "\n".join(messages) if messages else None


def get_enhanced_jit_context(prompt: str, cwd: str) -> Tuple[List[str], Optional[str]]:
    """Get enhanced JIT context

    Args:
        prompt: User prompt
        cwd: Current working directory

    Returns:
        (Context file list, system message)
    """
    agent_context = get_agent_delegation_context(prompt, cwd)

    # Combine all context files
    context_files = []

    # Add existing context
    traditional_context = agent_context.get("traditional_context", [])
    context_files.extend(traditional_context)

    # Add agent context files
    agent_context_files = agent_context.get("context_files", [])
    for file in agent_context_files:
        if file not in context_files:
            context_files.append(file)

    # Add skills reference files
    skill_references = agent_context.get("skill_references", [])
    for skill_ref in skill_references:
        if skill_ref not in context_files:
            context_files.append(skill_ref)

    # Generate system message
    system_message = format_agent_delegation_message(agent_context)

    return context_files, system_message


__all__ = [
    "load_agent_skills_mapping",
    "analyze_prompt_intent",
    "get_agent_delegation_context",
    "format_agent_delegation_message",
    "get_enhanced_jit_context"
]
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
    """에이전트-Skills 매핑 설정 로드

    Returns:
        매핑 설정 딕셔너리
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
    """사용자 프롬프트 의도 분석

    Args:
        prompt: 사용자 프롬프트
        mapping: 에이전트-Skills 매핑

    Returns:
        분석된 의도 정보 또는 None
    """
    prompt_lower = prompt.lower()

    # 각 패턴에 대해 점수 계산
    pattern_scores = []

    for pattern_name, pattern_config in mapping.get("prompt_patterns", {}).items():
        score = 0
        matched_keywords = []

        # 키워드 매칭
        for keyword in pattern_config.get("keywords", []):
            if keyword.lower() in prompt_lower:
                score += 1
                matched_keywords.append(keyword)

        # 정규식 패턴 매칭
        if pattern_config.get("regex_patterns"):
            for regex_pattern in pattern_config["regex_patterns"]:
                if re.search(regex_pattern, prompt_lower):
                    score += 2  # 정규식 매칭은 더 높은 가중치

        if score > 0:
            pattern_scores.append({
                "pattern": pattern_name,
                "score": score,
                "matched_keywords": matched_keywords,
                "config": pattern_config
            })

    # 가장 높은 점수의 패턴 반환
    if pattern_scores:
        pattern_scores.sort(key=lambda x: x["score"], reverse=True)
        best_match = pattern_scores[0]

        return {
            "intent": best_match["pattern"],
            "confidence": min(best_match["score"] / 3.0, 1.0),  # 최대 3개 키워드 매칭
            "matched_keywords": best_match["matched_keywords"],
            "primary_agent": best_match["config"].get("primary_agent"),
            "secondary_agents": best_match["config"].get("secondary_agents", []),
            "recommended_skills": best_match["config"].get("skills", []),
            "context_files": best_match["config"].get("context_files", [])
        }

    return None


def get_agent_delegation_context(prompt: str, cwd: str) -> Dict[str, Any]:
    """프롬프트 기반 에이전트 위임 컨텍스트 생성

    Args:
        prompt: 사용자 프롬프트
        cwd: 현재 작업 디렉토리

    Returns:
        에이전트 위임 컨텍스트 정보
    """
    mapping = load_agent_skills_mapping()
    intent_analysis = analyze_prompt_intent(prompt, mapping)

    # 기존 JIT 컨텍스트 가져오기
    existing_context = get_jit_context(prompt, cwd)

    # 에이전트 위임 정보
    agent_context = {
        "intent_detected": intent_analysis is not None,
        "traditional_context": existing_context
    }

    if intent_analysis:
        # 파일 존재 확인
        valid_context_files = []
        cwd_path = Path(cwd)

        for context_file in intent_analysis["context_files"]:
            file_path = cwd_path / context_file
            if file_path.exists():
                valid_context_files.append(context_file)

        # Skills 참조 경로 생성
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
    """에이전트 위임 메시지 포맷팅

    Args:
        context: 에이전트 위임 컨텍스트

    Returns:
        포맷된 메시지 또는 None
    """
    if not context.get("intent_detected"):
        return None

    messages = []

    # 기본 정보
    primary_agent = context.get("primary_agent")
    confidence = context.get("confidence", 0)
    intent = context.get("intent", "")
    matched_keywords = context.get("matched_keywords", [])

    if primary_agent and confidence > 0.5:
        messages.append(f"🎯 전문가 에이전트 추천: {primary_agent}")
        messages.append(f"📋 작업 의도: {intent}")

        if matched_keywords:
            messages.append(f"🔍 인식된 키워드: {', '.join(matched_keywords)}")

        # 추천 Skills
        skills = context.get("recommended_skills", [])
        if skills:
            messages.append(f"⚡ 추천 Skills: {', '.join(skills[:3])}")  # 최대 3개만 표시

        # 보조 에이전트
        secondary_agents = context.get("secondary_agents", [])
        if secondary_agents:
            messages.append(f"🤝 협업 에이전트: {', '.join(secondary_agents[:2])}")  # 최대 2개만 표시

        # 컨텍스트 파일
        context_files = context.get("context_files", [])
        skill_references = context.get("skill_references", [])

        all_files = context_files + skill_references
        if all_files:
            messages.append(f"📚 자동 로딩된 컨텍스트: {len(all_files)}개 파일")

    return "\n".join(messages) if messages else None


def get_enhanced_jit_context(prompt: str, cwd: str) -> Tuple[List[str], Optional[str]]:
    """향상된 JIT 컨텍스트 가져오기

    Args:
        prompt: 사용자 프롬프트
        cwd: 현재 작업 디렉토리

    Returns:
        (컨텍스트 파일 리스트, 시스템 메시지)
    """
    agent_context = get_agent_delegation_context(prompt, cwd)

    # 모든 컨텍스트 파일 결합
    context_files = []

    # 기존 컨텍스트 추가
    traditional_context = agent_context.get("traditional_context", [])
    context_files.extend(traditional_context)

    # 에이전트 컨텍스트 파일 추가
    agent_context_files = agent_context.get("context_files", [])
    for file in agent_context_files:
        if file not in context_files:
            context_files.append(file)

    # Skills 참조 파일 추가
    skill_references = agent_context.get("skill_references", [])
    for skill_ref in skill_references:
        if skill_ref not in context_files:
            context_files.append(skill_ref)

    # 시스템 메시지 생성
    system_message = format_agent_delegation_message(agent_context)

    return context_files, system_message


__all__ = [
    "load_agent_skills_mapping",
    "analyze_prompt_intent",
    "get_agent_delegation_context",
    "format_agent_delegation_message",
    "get_enhanced_jit_context"
]
#!/usr/bin/env python3
"""
JIT Skill Filter Strategy Implementation
=========================================

MoAI-ADK의 Phase별 필수 Skill만 필터링하여 토큰 사용을 최적화합니다.
- RED Phase: 테스트 관련 6개 Skill만 로드
- GREEN Phase: 언어별 3개 Skill만 로드
- REFACTOR Phase: 리팩토링 관련 4개 Skill만 로드

실행: uv run .moai/scripts/jit-skill-filter.py [phase] [language]
예시: uv run .moai/scripts/jit-skill-filter.py RED python
"""

import json
import sys
from pathlib import Path
from dataclasses import dataclass, field
from typing import List, Dict, Set
from enum import Enum

# ============================================================================
# 데이터 구조 정의
# ============================================================================

class Phase(Enum):
    """TDD Phase 정의"""
    SPEC = "SPEC"
    RED = "RED"
    GREEN = "GREEN"
    REFACTOR = "REFACTOR"


@dataclass
class PhaseSkillProfile:
    """Phase별 필수 Skill 프로필"""
    phase: Phase
    required_skills: List[str]
    optional_skills: List[str] = field(default_factory=list)
    language_specific: bool = False
    expected_tokens: int = 0
    description: str = ""


@dataclass
class SkillInfo:
    """Skill 메타데이터"""
    name: str
    path: str
    category: str  # "foundation", "domain", "lang", "essentials", "core"
    estimated_tokens: int
    description: str


@dataclass
class FilterResult:
    """필터링 결과"""
    phase: str
    language: str
    required_skills: List[SkillInfo]
    total_tokens: int
    filtered_out_count: int
    efficiency_percentage: float
    recommendations: List[str]


# ============================================================================
# Phase별 Skill 프로필 정의
# ============================================================================

PHASE_SKILL_PROFILES = {
    Phase.SPEC: PhaseSkillProfile(
        phase=Phase.SPEC,
        required_skills=[
            "moai-foundation-specs",
            "moai-core-agent-factory",
            "moai-foundation-trust",
        ],
        description="SPEC 명세 생성 - 기초 스킬만 필요",
        expected_tokens=30000,
    ),
    Phase.RED: PhaseSkillProfile(
        phase=Phase.RED,
        required_skills=[
            "moai-domain-testing",  # 테스트 프레임워크
            "moai-foundation-trust",  # TRUST 5 원칙
            "moai-essentials-review",  # 코드 리뷰
            "moai-core-code-reviewer",  # 코드 검토
            "moai-essentials-debug",  # 디버깅
        ],
        language_specific=True,
        description="RED Phase - 테스트 작성 (6개 Skill)",
        expected_tokens=25000,
    ),
    Phase.GREEN: PhaseSkillProfile(
        phase=Phase.GREEN,
        required_skills=[
            "moai-essentials-review",  # 코드 리뷰
        ],
        language_specific=True,
        description="GREEN Phase - 최소 구현 (3개 Skill: lang + domain)",
        expected_tokens=25000,
    ),
    Phase.REFACTOR: PhaseSkillProfile(
        phase=Phase.REFACTOR,
        required_skills=[
            "moai-essentials-refactor",  # 리팩토링
            "moai-essentials-review",  # 코드 리뷰
            "moai-core-code-reviewer",  # 코드 검토
            "moai-essentials-debug",  # 디버깅
        ],
        description="REFACTOR Phase - 코드 품질 (4개 Skill)",
        expected_tokens=20000,
    ),
}

# 언어별 Skill 매핑
LANGUAGE_SKILLS = {
    "python": "moai-lang-python",
    "typescript": "moai-lang-typescript",
    "javascript": "moai-lang-javascript",
    "go": "moai-lang-go",
    "rust": "moai-lang-rust",
    "java": "moai-lang-java",
    "csharp": "moai-lang-csharp",
    "kotlin": "moai-lang-kotlin",
    "swift": "moai-lang-swift",
    "dart": "moai-lang-dart",
    "shell": "moai-lang-shell",
    "cpp": "moai-lang-cpp",
    "c": "moai-lang-c",
}

# Domain별 Skill 매핑
DOMAIN_SKILLS = {
    "backend": "moai-domain-backend",
    "frontend": "moai-domain-frontend",
    "mobile": "moai-domain-mobile-app",
    "devops": "moai-domain-devops",
    "database": "moai-domain-database",
    "testing": "moai-domain-testing",
    "security": "moai-security-api",
    "monitoring": "moai-domain-monitoring",
}


# ============================================================================
# Skill 메타데이터 수집
# ============================================================================

def collect_skill_metadata(project_dir: Path) -> Dict[str, SkillInfo]:
    """프로젝트의 모든 Skill 메타데이터 수집"""
    skills = {}
    skills_dir = project_dir / ".claude" / "skills"

    if not skills_dir.exists():
        return skills

    for skill_dir in skills_dir.iterdir():
        if not skill_dir.is_dir() or "template" in str(skill_dir):
            continue

        skill_file = skill_dir / "SKILL.md"
        if not skill_file.exists():
            continue

        # 토큰 추정 (간단히 파일 크기 기반)
        try:
            size_bytes = skill_file.stat().st_size
            estimated_tokens = max(1000, size_bytes // 4)  # 약 4 bytes = 1 token
        except:
            estimated_tokens = 2000

        # Skill 카테고리 식별
        skill_name = skill_dir.name
        if "lang-" in skill_name:
            category = "lang"
        elif "domain-" in skill_name:
            category = "domain"
        elif "essentials-" in skill_name:
            category = "essentials"
        elif "foundation-" in skill_name:
            category = "foundation"
        elif "core-" in skill_name:
            category = "core"
        elif "security-" in skill_name:
            category = "security"
        else:
            category = "other"

        skills[skill_name] = SkillInfo(
            name=skill_name,
            path=str(skill_file.relative_to(project_dir)),
            category=category,
            estimated_tokens=estimated_tokens,
            description=f"{category.upper()} Skill",
        )

    return skills


# ============================================================================
# Skill 필터링
# ============================================================================

def filter_skills_for_phase(
    phase: Phase,
    language: str,
    domain: str,
    all_skills: Dict[str, SkillInfo],
) -> FilterResult:
    """
    Phase별로 필요한 Skill만 필터링

    Args:
        phase: TDD Phase
        language: 프로젝트 언어 (python, typescript, etc.)
        domain: 프로젝트 도메인 (backend, frontend, etc.)
        all_skills: 모든 Skill 메타데이터

    Returns:
        FilterResult: 필터링된 Skill 목록과 분석
    """

    profile = PHASE_SKILL_PROFILES.get(phase)
    if not profile:
        raise ValueError(f"Unknown phase: {phase}")

    # Phase별 필수 Skill 수집
    required_skills_list: List[SkillInfo] = []

    for skill_name in profile.required_skills:
        if skill_name in all_skills:
            required_skills_list.append(all_skills[skill_name])

    # 언어별 Skill 추가 (GREEN, RED 등 필요한 경우)
    if profile.language_specific:
        lang_skill_name = LANGUAGE_SKILLS.get(language.lower())
        if lang_skill_name and lang_skill_name in all_skills:
            required_skills_list.append(all_skills[lang_skill_name])

        # Domain 기반 Skill 추가 (GREEN phase)
        if phase == Phase.GREEN:
            domain_skill_name = DOMAIN_SKILLS.get(domain.lower())
            if domain_skill_name and domain_skill_name in all_skills:
                required_skills_list.append(all_skills[domain_skill_name])

    # 토큰 계산
    total_tokens = sum(s.estimated_tokens for s in required_skills_list)
    filtered_out_count = len(all_skills) - len(required_skills_list)

    # 효율성 계산
    all_tokens = sum(s.estimated_tokens for s in all_skills.values())
    efficiency = (total_tokens / all_tokens * 100) if all_tokens > 0 else 0

    # 권장사항 생성
    recommendations = []

    # 예산 준수 확인
    budget = profile.expected_tokens
    if total_tokens > budget * 1.1:  # 10% 마진
        recommendations.append(
            f"⚠️ 예산 초과: {total_tokens:,}K > {budget:,}K "
            f"({(total_tokens/budget*100):.1f}%)"
        )
    elif total_tokens <= budget * 0.7:  # 30% 이상 미달
        recommendations.append(
            f"✓ 효율적: {total_tokens:,}K / {budget:,}K ({(total_tokens/budget*100):.1f}%)"
        )
    else:
        recommendations.append(
            f"✓ 준수: {total_tokens:,}K / {budget:,}K ({(total_tokens/budget*100):.1f}%)"
        )

    # 대체 Skills 제안
    optional_skills = [
        s for s in all_skills.values()
        if s.name not in [req.name for req in required_skills_list]
        and total_tokens + s.estimated_tokens <= budget * 1.1
    ]

    if optional_skills:
        optional_skills.sort(key=lambda s: s.estimated_tokens)
        recommendations.append(
            f"Optional: {', '.join([s.name for s in optional_skills[:3]])}"
        )

    return FilterResult(
        phase=phase.value,
        language=language,
        required_skills=required_skills_list,
        total_tokens=total_tokens,
        filtered_out_count=filtered_out_count,
        efficiency_percentage=efficiency,
        recommendations=recommendations,
    )


# ============================================================================
# 보고서 생성 및 출력
# ============================================================================

def print_filter_report(filter_result: FilterResult):
    """필터링 결과 보고서 출력"""
    print("\n" + "=" * 80)
    print(f"JIT Skill Filter Report - {filter_result.phase} Phase ({filter_result.language})")
    print("=" * 80)

    print(f"\nRequired Skills: {len(filter_result.required_skills)}")
    print("-" * 80)

    total_estimated = 0
    for skill in filter_result.required_skills:
        print(f"  ✓ {skill.name:<35} {skill.estimated_tokens:>6,} tokens ({skill.category})")
        total_estimated += skill.estimated_tokens

    print("-" * 80)
    print(f"  Total: {total_estimated:>59,} tokens")
    print(f"  Filtered out: {filter_result.filtered_out_count} skills")
    print(f"  Efficiency: {filter_result.efficiency_percentage:.1f}% (Context saved)")

    print("\nRecommendations:")
    print("-" * 80)
    for rec in filter_result.recommendations:
        print(f"  {rec}")

    print("\n" + "=" * 80)


def save_filter_config(
    project_dir: Path,
    filter_results: Dict[str, FilterResult],
):
    """필터링 결과를 JSON으로 저장"""
    config = {
        "version": "1.0.0",
        "timestamp": __import__("datetime").datetime.now().isoformat(),
        "phases": {},
    }

    for phase_name, result in filter_results.items():
        config["phases"][phase_name] = {
            "phase": result.phase,
            "required_skills": [s.name for s in result.required_skills],
            "total_tokens": result.total_tokens,
            "efficiency_percentage": result.efficiency_percentage,
        }

    config_path = project_dir / ".moai" / "config" / "jit-skill-filter.json"
    config_path.parent.mkdir(parents=True, exist_ok=True)

    with open(config_path, "w") as f:
        json.dump(config, f, indent=2)

    print(f"\n✓ 필터링 설정 저장: {config_path}")

    return config_path


# ============================================================================
# 메인 실행
# ============================================================================

def main():
    """메인 실행 함수"""

    if len(sys.argv) < 2:
        # 대화형 모드
        project_dir = Path(__file__).parent.parent.parent
        print("=" * 80)
        print("JIT Skill Filter - Interactive Mode")
        print("=" * 80)

        # Skill 메타데이터 수집
        print("\n수집 중: Skill 메타데이터...")
        all_skills = collect_skill_metadata(project_dir)
        print(f"  ✓ {len(all_skills)}개 Skill 발견")

        # Phase별 필터링
        filter_results = {}

        for phase in Phase:
            print(f"\n분석 중: {phase.value} Phase...")

            # 언어 선택 (기본값: python)
            language = "python"

            filter_result = filter_skills_for_phase(
                phase=phase,
                language=language,
                domain="backend",
                all_skills=all_skills,
            )

            filter_results[phase.value] = filter_result
            print_filter_report(filter_result)

        # 설정 저장
        save_filter_config(project_dir, filter_results)

    else:
        # 명령행 인자 모드
        phase_str = sys.argv[1].upper()
        language = sys.argv[2].lower() if len(sys.argv) > 2 else "python"
        domain = sys.argv[3].lower() if len(sys.argv) > 3 else "backend"

        try:
            phase = Phase[phase_str]
        except KeyError:
            print(f"Error: Unknown phase '{phase_str}'")
            print(f"Available phases: {', '.join([p.name for p in Phase])}")
            sys.exit(1)

        project_dir = Path(__file__).parent.parent.parent

        # Skill 메타데이터 수집
        all_skills = collect_skill_metadata(project_dir)

        # 필터링
        filter_result = filter_skills_for_phase(
            phase=phase,
            language=language,
            domain=domain,
            all_skills=all_skills,
        )

        # 보고서 출력
        print_filter_report(filter_result)


if __name__ == "__main__":
    main()

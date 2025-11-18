#!/usr/bin/env python3
"""
JIT Context Strategy Validation Script
======================================

MoAI-ADK의 Phase별 JIT (Just-In-Time) Context Strategy를 실제로 검증합니다.
- Phase별 컨텍스트 로딩 효율 측정
- 토큰 사용량 예산 vs 실제 비교
- /clear 명령어의 토큰 절약 효과 분석
- Dynamic Context Loading 성능 평가

실행: uv run .moai/scripts/jit-context-validation.py
"""

import os
import json
import time
from pathlib import Path
from dataclasses import dataclass, asdict, field
from typing import List, Dict, Tuple
import re

# ============================================================================
# 데이터 구조 정의
# ============================================================================

@dataclass
class ContextMetric:
    """개별 문서/파일의 컨텍스트 메트릭"""
    name: str
    path: str
    size_bytes: int
    lines: int
    category: str  # "spec", "skill", "agent", "foundation"
    phase: str = ""  # "SPEC", "RED", "GREEN", "REFACTOR"
    estimated_tokens: int = field(default=0, init=False)

    def __post_init__(self):
        # 라인 수로부터 토큰 추정 (평균 약 1 줄 = 1.3 토큰)
        self.estimated_tokens = int(self.lines * 1.3)


@dataclass
class PhaseContextAnalysis:
    """Phase별 컨텍스트 분석 결과"""
    phase_name: str  # "SPEC", "RED", "GREEN", "REFACTOR"
    budget_tokens: int
    actual_tokens: int
    documents: List[ContextMetric] = field(default_factory=list)
    efficiency: float = field(default=0, init=False)  # 0-100%
    compliance: bool = field(default=True, init=False)  # 예산 준수 여부

    def __post_init__(self):
        self.calculate_efficiency()

    def calculate_efficiency(self):
        """효율성 계산 (실제 / 예산)"""
        if self.budget_tokens == 0:
            return 0
        self.efficiency = min(100, (self.actual_tokens / self.budget_tokens) * 100)
        self.compliance = self.actual_tokens <= self.budget_tokens


@dataclass
class ClearEffectAnalysis:
    """'/clear' 명령어 효과 분석"""
    before_tokens: int
    after_tokens: int
    saved_tokens: int = 0
    savings_percentage: float = 0

    def __post_init__(self):
        self.saved_tokens = self.before_tokens - self.after_tokens
        if self.before_tokens > 0:
            self.savings_percentage = (self.saved_tokens / self.before_tokens) * 100


@dataclass
class ValidationReport:
    """전체 검증 보고서"""
    timestamp: str
    project_name: str
    project_version: str

    # 기준선 데이터
    total_specs: int
    total_skills: int
    total_agents: int
    total_context_size_mb: float

    # Phase별 분석
    phase_analyses: List[PhaseContextAnalysis] = field(default_factory=list)

    # /clear 효과
    clear_effect: ClearEffectAnalysis = None

    # 최적화 추천
    recommendations: List[str] = field(default_factory=list)

    # 성능 지표
    avg_token_efficiency: float = 0
    bottleneck_phases: List[str] = field(default_factory=list)


# ============================================================================
# 데이터 수집 함수
# ============================================================================

def get_file_metrics(file_path: Path) -> Tuple[int, int]:
    """파일의 바이트 크기와 라인 수 반환"""
    try:
        size_bytes = file_path.stat().st_size
        with open(file_path, 'r', encoding='utf-8', errors='ignore') as f:
            lines = len(f.readlines())
        return size_bytes, lines
    except Exception:
        return 0, 0


def collect_spec_metrics(project_dir: Path) -> List[ContextMetric]:
    """SPEC 문서 메트릭 수집"""
    specs = []
    specs_dir = project_dir / ".moai" / "specs"

    if not specs_dir.exists():
        return specs

    for spec_dir in specs_dir.iterdir():
        if not spec_dir.is_dir():
            continue

        spec_file = spec_dir / "spec.md"
        if spec_file.exists():
            size_bytes, lines = get_file_metrics(spec_file)

            # SPEC ID 추출
            spec_id = spec_dir.name

            # Phase 추정 (SPEC 파일의 상태에서)
            phase = "SPEC"  # 모든 SPEC 파일은 SPEC phase로 분류

            spec_metric = ContextMetric(
                name=spec_id,
                path=str(spec_file.relative_to(project_dir)),
                size_bytes=size_bytes,
                lines=lines,
                category="spec",
                phase=phase
            )
            specs.append(spec_metric)

    return specs


def collect_skill_metrics(project_dir: Path) -> List[ContextMetric]:
    """Skill 문서 메트릭 수집 (템플릿 제외)"""
    skills = []
    skills_dir = project_dir / ".claude" / "skills"

    if not skills_dir.exists():
        return skills

    for skill_dir in skills_dir.iterdir():
        if not skill_dir.is_dir() or "template" in str(skill_dir):
            continue

        skill_file = skill_dir / "SKILL.md"
        if skill_file.exists():
            size_bytes, lines = get_file_metrics(skill_file)
            skill_name = skill_dir.name

            # Skill 카테고리 식별
            if "lang-" in skill_name:
                category = "skill-lang"
            elif "domain-" in skill_name:
                category = "skill-domain"
            elif "moai-" in skill_name:
                category = "skill-moai"
            else:
                category = "skill-other"

            skill_metric = ContextMetric(
                name=skill_name,
                path=str(skill_file.relative_to(project_dir)),
                size_bytes=size_bytes,
                lines=lines,
                category=category
            )
            skills.append(skill_metric)

    return skills


def collect_agent_metrics(project_dir: Path) -> List[ContextMetric]:
    """Agent 설정 메트릭 수집"""
    agents = []
    agents_dir = project_dir / ".claude" / "agents" / "moai"

    if not agents_dir.exists():
        return agents

    for agent_file in agents_dir.glob("*.md"):
        size_bytes, lines = get_file_metrics(agent_file)
        agent_name = agent_file.stem

        agent_metric = ContextMetric(
            name=agent_name,
            path=str(agent_file.relative_to(project_dir)),
            size_bytes=size_bytes,
            lines=lines,
            category="agent"
        )
        agents.append(agent_metric)

    return agents


def collect_foundation_metrics(project_dir: Path) -> List[ContextMetric]:
    """기초 문서 (CLAUDE.md, config 등) 메트릭 수집"""
    foundation = []

    # CLAUDE.md
    claude_md = project_dir / "CLAUDE.md"
    if claude_md.exists():
        size_bytes, lines = get_file_metrics(claude_md)
        foundation.append(ContextMetric(
            name="CLAUDE.md",
            path="CLAUDE.md",
            size_bytes=size_bytes,
            lines=lines,
            category="foundation"
        ))

    # CLAUDE.local.md
    claude_local = project_dir / "CLAUDE.local.md"
    if claude_local.exists():
        size_bytes, lines = get_file_metrics(claude_local)
        foundation.append(ContextMetric(
            name="CLAUDE.local.md",
            path="CLAUDE.local.md",
            size_bytes=size_bytes,
            lines=lines,
            category="foundation"
        ))

    return foundation


# ============================================================================
# Phase별 컨텍스트 분석
# ============================================================================

def analyze_phase_context(
    phase: str,
    budget: int,
    all_metrics: List[ContextMetric]
) -> PhaseContextAnalysis:
    """
    Phase별 컨텍스트 로딩 효율 분석

    Phase별 필요 문서:
    - SPEC: foundation + 해당 SPEC 파일
    - RED: foundation + 해당 SPEC + test-related skills
    - GREEN: foundation + 해당 SPEC + language-specific skills
    - REFACTOR: foundation + 해당 SPEC + refactor skills
    """

    # Phase별 필요 문서 카테고리
    phase_requirements = {
        "SPEC": ["foundation", "spec"],
        "RED": ["foundation", "spec", "skill-moai"],  # 테스트 스킬
        "GREEN": ["foundation", "spec", "skill-lang"],  # 언어별 스킬
        "REFACTOR": ["foundation", "spec", "skill-moai"]  # 리팩토링 스킬
    }

    required_categories = phase_requirements.get(phase, [])

    # 필요한 문서만 필터링
    phase_docs = [
        m for m in all_metrics
        if m.category in required_categories
    ]

    # 토큰 합계 계산
    total_tokens = sum(m.estimated_tokens for m in phase_docs)

    analysis = PhaseContextAnalysis(
        phase_name=phase,
        budget_tokens=budget,
        actual_tokens=total_tokens,
        documents=phase_docs
    )
    return analysis


# ============================================================================
# /clear 효과 분석
# ============================================================================

def estimate_clear_effect(all_metrics: List[ContextMetric]) -> ClearEffectAnalysis:
    """
    /clear 명령어의 예상 토큰 절약 효과 분석

    /clear는 컨텍스트를 완전히 초기화하므로:
    - 이전: 모든 누적된 컨텍스트
    - 이후: 최소 기초 문서만 (약 5K 토큰)
    """

    # 전체 컨텍스트
    total_tokens = sum(m.estimated_tokens for m in all_metrics)

    # /clear 후 유지되는 최소 컨텍스트 (약 5K 토큰)
    remaining_tokens = 5000

    effect = ClearEffectAnalysis(
        before_tokens=total_tokens,
        after_tokens=remaining_tokens
    )

    return effect


# ============================================================================
# 최적화 권장사항
# ============================================================================

def generate_recommendations(
    phase_analyses: List[PhaseContextAnalysis],
    all_metrics: List[ContextMetric]
) -> Tuple[List[str], List[str]]:
    """
    JIT 최적화 권장사항 생성

    Returns:
        (recommendations, bottleneck_phases)
    """
    recommendations = []
    bottlenecks = []

    # Phase별 효율성 분석
    for analysis in phase_analyses:
        if analysis.efficiency > 80:
            recommendations.append(
                f"{analysis.phase_name}: 효율적 (사용률 {analysis.efficiency:.1f}%)"
            )
        elif analysis.efficiency > 60:
            recommendations.append(
                f"{analysis.phase_name}: 개선 권장 (사용률 {analysis.efficiency:.1f}%) "
                f"- 불필요한 Skill 제외 고려"
            )
            bottlenecks.append(analysis.phase_name)
        else:
            recommendations.append(
                f"{analysis.phase_name}: 즉시 최적화 필요 (사용률 {analysis.efficiency:.1f}%)"
            )
            bottlenecks.append(analysis.phase_name)

    # 상위 크기 파일 식별
    large_files = sorted(
        all_metrics,
        key=lambda m: m.estimated_tokens,
        reverse=True
    )[:5]

    if large_files:
        recommendations.append("\n최상위 5개 대용량 문서:")
        for f in large_files:
            recommendations.append(
                f"  - {f.name}: {f.estimated_tokens} 토큰 ({f.lines} 줄)"
            )

    # Skill 최적화 권장
    skills = [m for m in all_metrics if m.category.startswith("skill-")]
    if len(skills) > 100:
        recommendations.append(
            f"\nSkill 정리 권장: {len(skills)}개 Skill 중 미사용 Skill 검토"
        )

    # /clear 활용 권장
    recommendations.append(
        "\n토큰 효율 최적화:\n"
        "  1. 각 Phase 완료 후 /clear 실행 (45-50K 토큰 절약)\n"
        "  2. Phase별 필수 Skill만 활성화\n"
        "  3. 대규모 SPEC은 sub-SPEC으로 분할 권장"
    )

    return recommendations, bottlenecks


# ============================================================================
# 보고서 생성
# ============================================================================

def generate_report(
    project_dir: Path,
    all_metrics: List[ContextMetric],
    phase_analyses: List[PhaseContextAnalysis],
    clear_effect: ClearEffectAnalysis
) -> ValidationReport:
    """종합 검증 보고서 생성"""

    # 메타데이터
    config_path = project_dir / ".moai" / "config" / "config.json"
    project_name = "MoAI-ADK"
    project_version = "0.26.0"

    if config_path.exists():
        try:
            with open(config_path) as f:
                config = json.load(f)
                project_name = config.get("project", {}).get("name", project_name)
                project_version = config.get("moai", {}).get("version", project_version)
        except:
            pass

    # 카테고리별 집계
    specs = [m for m in all_metrics if m.category == "spec"]
    skills = [m for m in all_metrics if m.category.startswith("skill-")]
    agents = [m for m in all_metrics if m.category == "agent"]

    # 전체 토큰
    total_tokens = sum(m.estimated_tokens for m in all_metrics)
    total_mb = sum(m.size_bytes for m in all_metrics) / (1024 * 1024)

    # 평균 효율
    avg_efficiency = sum(a.efficiency for a in phase_analyses) / len(phase_analyses) if phase_analyses else 0

    # 병목 Phase
    bottlenecks = [a.phase_name for a in phase_analyses if not a.compliance]

    # 권장사항
    recommendations, bottleneck_phases = generate_recommendations(phase_analyses, all_metrics)

    report = ValidationReport(
        timestamp=time.strftime("%Y-%m-%d %H:%M:%S"),
        project_name=project_name,
        project_version=project_version,
        total_specs=len(specs),
        total_skills=len(skills),
        total_agents=len(agents),
        total_context_size_mb=total_mb,
        phase_analyses=phase_analyses,
        clear_effect=clear_effect,
        recommendations=recommendations,
        avg_token_efficiency=avg_efficiency,
        bottleneck_phases=bottleneck_phases
    )

    return report


# ============================================================================
# 메인 실행
# ============================================================================

def main():
    """메인 실행 함수"""

    print("=" * 80)
    print("JIT Context Strategy Validation - 시작")
    print("=" * 80)

    project_dir = Path(__file__).parent.parent.parent

    print(f"\n📊 프로젝트: {project_dir.name}")
    print(f"📂 작업 디렉토리: {project_dir}\n")

    # Step 1: 메트릭 수집
    print("Step 1: 메트릭 수집 중...")
    specs = collect_spec_metrics(project_dir)
    skills = collect_skill_metrics(project_dir)
    agents = collect_agent_metrics(project_dir)
    foundation = collect_foundation_metrics(project_dir)

    all_metrics = specs + skills + agents + foundation

    print(f"  ✓ SPEC: {len(specs)}개")
    print(f"  ✓ Skills: {len(skills)}개")
    print(f"  ✓ Agents: {len(agents)}개")
    print(f"  ✓ Foundation: {len(foundation)}개")
    print(f"  ✓ 총 문서: {len(all_metrics)}개\n")

    # Step 2: Phase별 분석
    print("Step 2: Phase별 컨텍스트 분석 중...")
    phase_budgets = {
        "SPEC": 50000,
        "RED": 60000,
        "GREEN": 60000,
        "REFACTOR": 50000
    }

    phase_analyses = []
    for phase, budget in phase_budgets.items():
        analysis = analyze_phase_context(phase, budget, all_metrics)
        phase_analyses.append(analysis)

        status = "✓ 준수" if analysis.compliance else "✗ 초과"
        print(f"  {phase}: {analysis.actual_tokens:,} / {budget:,} 토큰 ({analysis.efficiency:.1f}%) {status}")

    print()

    # Step 3: /clear 효과 분석
    print("Step 3: /clear 효과 분석 중...")
    clear_effect = estimate_clear_effect(all_metrics)

    print(f"  절약 전: {clear_effect.before_tokens:,} 토큰")
    print(f"  절약 후: {clear_effect.after_tokens:,} 토큰")
    print(f"  절약량: {clear_effect.saved_tokens:,} 토큰 ({clear_effect.savings_percentage:.1f}%)\n")

    # Step 4: 보고서 생성
    print("Step 4: 종합 보고서 생성 중...")
    report = generate_report(
        project_dir,
        all_metrics,
        phase_analyses,
        clear_effect
    )

    # 보고서 저장
    report_dir = project_dir / ".moai" / "reports"
    report_dir.mkdir(parents=True, exist_ok=True)

    report_file = report_dir / f"jit-context-validation-{time.strftime('%Y%m%d-%H%M%S')}.json"

    with open(report_file, 'w', encoding='utf-8') as f:
        json.dump(asdict(report), f, indent=2, ensure_ascii=False)

    print(f"  ✓ 보고서 저장: {report_file}\n")

    # Step 5: 권장사항 출력
    print("=" * 80)
    print("최적화 권장사항")
    print("=" * 80)
    for rec in report.recommendations:
        print(rec)

    print("\n" + "=" * 80)
    print("검증 완료")
    print("=" * 80)

    return report


if __name__ == "__main__":
    report = main()

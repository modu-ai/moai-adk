"""
Interactive setup wizard for MoAI-ADK projects.

Handles user input collection through a step-by-step wizard interface.
"""

import sys
from typing import Dict, Any

import click
from colorama import Fore, Style

try:
    from ..utils.logger import get_logger
except ImportError:
    from moai_adk.utils.logger import get_logger

logger = get_logger(__name__)


class InteractiveWizard:
    """Interactive setup wizard for MoAI-ADK projects."""
    
    def __init__(self):
        self.answers = {}
        self.tech_stack = []
    
    def run_wizard(self) -> Dict[str, Any]:
        """Run the complete 10-step interactive wizard."""
        
        click.echo(f"\n{Fore.CYAN}🗿 MoAI-ADK 대화형 초기화 마법사{Style.RESET_ALL}")
        click.echo("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
        
        # Step 1-3: 제품 비전 설정
        self._collect_product_vision()
        
        # Step 4-5: 기술 스택 설정  
        self._collect_tech_stack()
        
        # Step 6-7: 품질 기준 설정
        self._collect_quality_standards()
        
        # Step 8-10: 고급 설정 (선택사항)
        if click.confirm(f"\n{Fore.YELLOW}고급 설정을 진행하시겠습니까? (보안, 운영, 리스크 관리){Style.RESET_ALL}"):
            self._collect_advanced_settings()
        
        # 설정 요약 및 확인
        self._show_summary()
        
        if click.confirm(f"\n{Fore.GREEN}이 설정으로 프로젝트를 초기화하시겠습니까?{Style.RESET_ALL}"):
            return self.answers
        else:
            click.echo(f"{Fore.YELLOW}초기화가 취소되었습니다.{Style.RESET_ALL}")
            sys.exit(0)
    
    def _collect_product_vision(self):
        """1-3단계: 제품 비전 수집"""
        click.echo(f"\n{Fore.BLUE}📋 1단계: 제품 비전 설정{Style.RESET_ALL}")
        click.echo("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
        
        # Q1: 핵심 문제
        while True:
            problem = click.prompt(f"{Fore.WHITE}Q1. 이 프로젝트가 해결하려는 핵심 문제는 무엇인가요?{Style.RESET_ALL}")
            if len(problem) >= 20:
                break
            click.echo(f"{Fore.YELLOW}⚠️  더 구체적으로 설명해주세요 (최소 20자). 대상, 원인, 빈도를 포함하여 작성해주세요.{Style.RESET_ALL}")
        
        self.answers["core_problem"] = problem
        
        # Q2: 목표 사용자
        users = click.prompt(
            f"{Fore.WHITE}Q2. 목표 사용자는 누구인가요?{Style.RESET_ALL}",
            type=click.Choice(["개발자", "일반 사용자", "관리자", "B2B 고객", "API 사용자", "기타"], case_sensitive=False)
        )
        self.answers["target_users"] = users
        
        # Q3: 6개월 목표  
        while True:
            goal = click.prompt(f"{Fore.WHITE}Q3. 6개월 후 달성하고 싶은 구체적인 목표는?{Style.RESET_ALL}")
            if any(metric in goal.lower() for metric in ["mau", "사용자", "응답시간", "오류율", "%", "개", "명"]):
                break
            click.echo(f"{Fore.YELLOW}⚠️  측정 가능한 KPI를 포함해주세요 (예: MAU 1000명, 응답시간 500ms 이하){Style.RESET_ALL}")
        
        self.answers["goal"] = goal
        
        # Q4: 핵심 기능 3가지
        click.echo(f"\n{Fore.WHITE}Q4. 핵심 기능 3가지를 우선순위대로 입력해주세요:{Style.RESET_ALL}")
        features = []
        for i in range(3):
            feature = click.prompt(f"  {i+1}순위 기능")
            features.append(feature)
            
            if i < 2 and not click.confirm(f"    {i+2}순위 기능을 추가하시겠습니까?"):
                break
        
        self.answers["core_features"] = features
        
        click.echo(f"{Fore.GREEN}✅ 제품 비전 설정 완료{Style.RESET_ALL}")
    
    def _collect_tech_stack(self):
        """4-5단계: 기술 스택 수집"""
        click.echo(f"\n{Fore.BLUE}🔧 2단계: 기술 스택 설정{Style.RESET_ALL}")
        click.echo("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
        
        # Q5: 주요 기술 스택
        click.echo(f"{Fore.WHITE}Q5. 주요 기술 스택을 선택해주세요:{Style.RESET_ALL}")
        
        categories = {
            "웹": ["React", "Vue", "Angular", "Svelte", "Next.js", "Nuxt.js"],
            "모바일": ["React Native", "Flutter", "SwiftUI", "Kotlin"],
            "백엔드": ["FastAPI", "Django", "Flask", "Express", "Spring Boot"],
            "데이터베이스": ["PostgreSQL", "MySQL", "MongoDB", "Redis", "SQLite"],
            "인프라": ["Docker", "Kubernetes", "AWS", "GCP", "Azure"]
        }
        
        selected_tech = []
        for category, options in categories.items():
            click.echo(f"\n{Fore.CYAN}{category}:{Style.RESET_ALL}")
            for i, option in enumerate(options, 1):
                click.echo(f"  {i}. {option}")
            
            choices = click.prompt(
                f"선택 (번호 입력, 여러 개는 쉼표로 구분, 건너뛰려면 엔터)", 
                default="", 
                show_default=False
            )
            
            if choices:
                for choice in choices.split(","):
                    try:
                        idx = int(choice.strip()) - 1
                        if 0 <= idx < len(options):
                            selected_tech.append(options[idx])
                    except ValueError:
                        pass
        
        self.tech_stack = selected_tech
        self.answers["tech_stack"] = selected_tech
        
        # Q6: 팀 숙련도
        skill_level = click.prompt(
            f"{Fore.WHITE}Q6. 팀의 기술 숙련도는?{Style.RESET_ALL}",
            type=click.Choice(["초급", "중급", "고급"], case_sensitive=False)
        )
        self.answers["skill_level"] = skill_level
        
        click.echo(f"{Fore.GREEN}✅ 기술 스택 설정 완료{Style.RESET_ALL}")
    
    def _collect_quality_standards(self):
        """6-7단계: 품질 기준 수집"""
        click.echo(f"\n{Fore.BLUE}🧪 3단계: 품질 기준 설정{Style.RESET_ALL}")
        click.echo("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
        
        # Q7: 품질 목표
        coverage = click.prompt(
            f"{Fore.WHITE}Q7. 테스트 커버리지 목표는? (%){Style.RESET_ALL}",
            type=click.IntRange(60, 100),
            default=80
        )
        self.answers["test_coverage"] = coverage
        
        performance = click.prompt(f"{Fore.WHITE}API 응답시간 목표는? (ms){Style.RESET_ALL}", default="500")
        self.answers["performance_target"] = performance
        
        click.echo(f"{Fore.GREEN}✅ 품질 기준 설정 완료{Style.RESET_ALL}")
    
    def _collect_advanced_settings(self):
        """8-10단계: 고급 설정 수집"""
        click.echo(f"\n{Fore.BLUE}🛡️ 4단계: 고급 설정{Style.RESET_ALL}")
        click.echo("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
        
        # 보안 설정
        security_features = []
        if click.confirm(f"{Fore.WHITE}사용자 인증이 필요한가요?{Style.RESET_ALL}"):
            security_features.append("authentication")
        
        if click.confirm(f"{Fore.WHITE}API 키 관리가 필요한가요?{Style.RESET_ALL}"):
            security_features.append("api_key_management")
        
        if click.confirm(f"{Fore.WHITE}데이터 암호화가 필요한가요?{Style.RESET_ALL}"):
            security_features.append("encryption")
        
        self.answers["security_features"] = security_features
        
        # 운영 설정
        monitoring = click.confirm(f"{Fore.WHITE}모니터링 및 로깅 설정이 필요한가요?{Style.RESET_ALL}")
        self.answers["monitoring"] = monitoring
        
        ci_cd = click.confirm(f"{Fore.WHITE}CI/CD 파이프라인 설정이 필요한가요?{Style.RESET_ALL}")
        self.answers["ci_cd"] = ci_cd
        
        click.echo(f"{Fore.GREEN}✅ 고급 설정 완료{Style.RESET_ALL}")
    
    def _show_summary(self):
        """설정 요약 출력"""
        click.echo(f"\n{Fore.CYAN}📋 설정 요약{Style.RESET_ALL}")
        click.echo("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
        
        click.echo(f"{Fore.WHITE}제품 정보:{Style.RESET_ALL}")
        click.echo(f"  • 핵심 문제: {self.answers.get('core_problem', 'N/A')[:50]}...")
        click.echo(f"  • 목표 사용자: {self.answers.get('target_users', 'N/A')}")
        click.echo(f"  • 6개월 목표: {self.answers.get('goal', 'N/A')[:50]}...")
        
        click.echo(f"\n{Fore.WHITE}기술 설정:{Style.RESET_ALL}")
        click.echo(f"  • 기술 스택: {', '.join(self.answers.get('tech_stack', []))}")
        click.echo(f"  • 팀 숙련도: {self.answers.get('skill_level', 'N/A')}")
        
        click.echo(f"\n{Fore.WHITE}품질 기준:{Style.RESET_ALL}")
        click.echo(f"  • 테스트 커버리지: {self.answers.get('test_coverage', 80)}%")
        click.echo(f"  • 성능 목표: {self.answers.get('performance_target', '500')}ms")
        
        if self.answers.get('security_features'):
            click.echo(f"\n{Fore.WHITE}보안 기능:{Style.RESET_ALL}")
            for feature in self.answers.get('security_features', []):
                click.echo(f"  • {feature}")
#!/usr/bin/env python3
"""
Research Integration for moai-project-config-manager
Provides intelligent configuration analysis and optimization recommendations
"""

from pathlib import Path
import json
import re
from typing import Dict, List, Any
from datetime import datetime


class ConfigurationResearchEngine:
    """Advanced research engine for configuration optimization"""

    def __init__(self):
        self.research_strategies = {
            'LANGUAGE_OPTIMIZATION': self.analyze_language_config,
            'WORKFLOW_OPTIMIZATION': self.analyze_workflow_config,
            'REPORTING_OPTIMIZATION': self.analyze_reporting_config,
            'DOMAIN_OPTIMIZATION': self.analyze_domain_config,
            'PERFORMANCE_OPTIMIZATION': self.analyze_performance_config,
            'SECURITY_OPTIMIZATION': self.analyze_security_config
        }

    def analyze_configuration(self, config: Dict[str, Any]) -> Dict[str, Any]:
        """Comprehensive configuration analysis using all research strategies"""
        research_findings = {
            'timestamp': datetime.now().isoformat(),
            'analysis_type': 'COMPREHENSIVE_CONFIG_ANALYSIS',
            'findings': [],
            'recommendations': [],
            'risk_assessment': [],
            'optimization_opportunities': []
        }

        # Apply each research strategy
        for strategy_name, strategy_func in self.research_strategies.items():
            try:
                result = strategy_func(config)
                if result:
                    research_findings['findings'].extend(result.get('findings', []))
                    research_findings['recommendations'].extend(result.get('recommendations', []))
                    research_findings['risk_assessment'].extend(result.get('risk_assessment', []))
                    research_findings['optimization_opportunities'].extend(result.get('optimization_opportunities', []))
            except Exception as e:
                research_findings['findings'].append({
                    'category': 'ANALYSIS_ERROR',
                    'message': f"Strategy {strategy_name} failed: {str(e)}"
                })

        # Prioritize and deduplicate recommendations
        research_findings['recommendations'] = self.prioritize_recommendations(research_findings['recommendations'])
        research_findings['optimization_opportunities'] = self.rank_optimizations(research_findings['optimization_opportunities'])

        return research_findings

    def analyze_language_config(self, config: Dict[str, Any]) -> Dict[str, Any]:
        """Analyze language configuration for optimization opportunities"""
        findings = []
        recommendations = []
        risk_assessment = []
        optimizations = []

        if 'language' in config:
            lang = config['language']

            # Conversation language analysis
            if 'conversation_language' in lang:
                current_lang = lang['conversation_language']
                findings.append({
                    'category': 'LANGUAGE_ANALYSIS',
                    'analysis': f'Current conversation language: {current_lang}',
                    'impact': 'HIGH'
                })

                # Language optimization recommendations
                if current_lang == 'en':
                    recommendations.extend([
                        'Consider multilingual team support if working with international developers',
                        'Evaluate agent prompt language for global standard vs localized performance'
                    ])
                    optimizations.append({
                        'name': 'MULTILINGUAL_SUPPORT',
                        'priority': 'MEDIUM',
                        'potential_benefit': 'Expanded team collaboration',
                        'effort': 'LOW'
                    })

            # Agent prompt language analysis
            if 'agent_prompt_language' in lang:
                prompt_lang = lang['agent_prompt_language']
                findings.append({
                    'category': 'AGENT_LANGUAGE_ANALYSIS',
                    'analysis': f'Agent prompt language: {prompt_lang}',
                    'impact': 'MEDIUM'
                })

                if prompt_lang == 'english':
                    recommendations.extend([
                        'Localized prompts may improve performance for specific languages',
                        'English provides consistency for international teams'
                    ])

        return {
            'findings': findings,
            'recommendations': recommendations,
            'risk_assessment': risk_assessment,
            'optimization_opportunities': optimizations
        }

    def analyze_workflow_config(self, config: Dict[str, Any]) -> Dict[str, Any]:
        """Analyze Git workflow configuration"""
        findings = []
        recommendations = []
        risk_assessment = []
        optimizations = []

        if 'github' in config:
            github = config['github']

            # Git workflow analysis
            if 'spec_git_workflow' in github:
                workflow = github['spec_git_workflow']
                findings.append({
                    'category': 'WORKFLOW_ANALYSIS',
                    'analysis': f'Current Git workflow: {workflow}',
                    'impact': 'HIGH'
                })

                # Workflow-specific recommendations
                workflow_recommendations = {
                    'feature_branch': [
                        'Best practice for team collaboration',
                        'Reduces merge conflicts through isolation',
                        'Enables comprehensive code review'
                    ],
                    'develop_direct': [
                        'Fast development for solo projects',
                        'Higher risk of conflicts in team environments',
                        'Streamlined but less controlled process'
                    ],
                    'per_spec': [
                        'Flexible approach for varied project types',
                        'Complex to manage at scale',
                        'Requires consistent specification structure'
                    ]
                }

                recommendations.extend(workflow_recommendations.get(workflow, []))

            # Auto-delete branches analysis
            if 'auto_delete_branches' in github:
                auto_delete = github['auto_delete_branches']
                findings.append({
                    'category': 'BRANCH_MANAGEMENT_ANALYSIS',
                    'analysis': f'Auto-delete branches: {auto_delete}',
                    'impact': 'MEDIUM'
                })

                if auto_delete:
                    recommendations.extend([
                        'Maintains clean repository',
                        'Reduces storage costs',
                        'Requires careful branch naming convention'
                    ])
                    optimizations.append({
                        'name': 'BRANCH_LIFECYCLE_OPTIMIZATION',
                        'priority': 'MEDIUM',
                        'potential_benefit': 'Reduced repository clutter',
                        'effort': 'LOW'
                    })

        return {
            'findings': findings,
            'recommendations': recommendations,
            'risk_assessment': risk_assessment,
            'optimization_opportunities': optimizations
        }

    def analyze_reporting_config(self, config: Dict[str, Any]) -> Dict[str, Any]:
        """Analyze report generation configuration"""
        findings = []
        recommendations = []
        risk_assessment = []
        optimizations = []

        if 'report_generation' in config:
            report = config['report_generation']

            # Reporting configuration analysis
            if report.get('enabled', False):
                findings.append({
                    'category': 'REPORTING_ENABLED',
                    'analysis': 'Report generation is enabled',
                    'impact': 'MEDIUM'
                })

                if report.get('auto_create', False):
                    findings.append({
                        'category': 'AUTO_REPORTING_ANALYSIS',
                        'analysis': 'Automatic report creation enabled',
                        'impact': 'LOW'
                    })
                    recommendations.extend([
                        'Consider minimal reports for token efficiency',
                        'Enable reports for team transparency when needed'
                    ])
                    optimizations.append({
                        'name': 'REPORTING_EFFICIENCY',
                        'priority': 'HIGH',
                        'potential_benefit': 'Token savings while maintaining visibility',
                        'effort': 'LOW'
                    })
                else:
                    recommendations.extend([
                        'Enable auto-creation for consistent documentation',
                        'Use minimal option for better performance'
                    ])
            else:
                findings.append({
                    'category': 'REPORTING_DISABLED',
                    'analysis': 'Report generation is disabled',
                    'impact': 'MEDIUM'
                })
                recommendations.extend([
                    'Consider enabling reports for team transparency',
                    'Use minimal reporting for critical workflows'
                ])

        return {
            'findings': findings,
            'recommendations': recommendations,
            'risk_assessment': risk_assessment,
            'optimization_opportunities': optimizations
        }

    def analyze_domain_config(self, config: Dict[str, Any]) -> Dict[str, Any]:
        """Analyze project domain configuration"""
        findings = []
        recommendations = []
        risk_assessment = []
        optimizations = []

        if 'stack' in config and 'selected_domains' in config['stack']:
            domains = config['stack']['selected_domains']
            findings.append({
                'category': 'DOMAIN_ANALYSIS',
                'analysis': f'Selected domains: {domains}',
                'impact': 'MEDIUM'
            })

            # Domain-specific recommendations
            domain_specific_recommendations = {
                'frontend': [
                    'Consider React/Vue/Angular optimization patterns',
                    'Evaluate mobile development requirements'
                ],
                'backend': [
                    'API design best practices critical',
                    'Consider microservices architecture if growing'
                ],
                'data': [
                    'Database optimization essential',
                    'Consider caching strategies'
                ],
                'devops': [
                    'CI/CD automation recommended',
                    'Infrastructure as code patterns helpful'
                ],
                'security': [
                    'Security audit essential',
                    'Consider penetration testing'
                ]
            }

            for domain in domains:
                if domain in domain_specific_recommendations:
                    recommendations.extend(domain_specific_recommendations[domain])
                    optimizations.append({
                        'name': f'{domain.upper()}_OPTIMIZATION',
                        'priority': 'HIGH' if domain in ['security', 'backend'] else 'MEDIUM',
                        'potential_benefit': 'Domain-specific performance',
                        'effort': 'MEDIUM'
                    })

        return {
            'findings': findings,
            'recommendations': recommendations,
            'risk_assessment': risk_assessment,
            'optimization_opportunities': optimizations
        }

    def analyze_performance_config(self, config: Dict[str, Any]) -> Dict[str, Any]:
        """Analyze configuration for performance implications"""
        findings = []
        recommendations = []
        risk_assessment = []
        optimizations = []

        # Check for potential performance issues
        if 'language' in config and config['language'].get('conversation_language') == 'en':
            recommendations.extend([
                'English provides consistent performance',
                'Monitor for token optimization opportunities'
            ])
            optimizations.append({
                'name': 'TOKEN_OPTIMIZATION',
                'priority': 'HIGH',
                'potential_benefit': 'Reduced API costs',
                'effort': 'LOW'
            })

        # Reporting performance impact
        if 'report_generation' in config and config['report_generation'].get('enabled', True):
            report_level = config['report_generation'].get('user_choice', 'Minimal')
            if report_level == 'Enable':
                recommendations.extend([
                    'Full reports may impact token usage',
                    'Consider minimal reports for production'
                ])
                optimizations.append({
                    'name': 'REPORTING_PERFORMANCE',
                    'priority': 'HIGH',
                    'potential_benefit': 'Token efficiency',
                    'effort': 'LOW'
                })

        return {
            'findings': findings,
            'recommendations': recommendations,
            'risk_assessment': risk_assessment,
            'optimization_opportunities': optimizations
        }

    def analyze_security_config(self, config: Dict[str, Any]) -> Dict[str, Any]:
        """Analyze configuration for security implications"""
        findings = []
        recommendations = []
        risk_assessment = []
        optimizations = []

        # GitHub security analysis
        if 'github' in config:
            github = config['github']

            if not github.get('auto_delete_branches', False):
                risk_assessment.append({
                    'category': 'BRANCH_MANAGEMENT_RISK',
                    'severity': 'MEDIUM',
                    'description': 'Auto-delete disabled may lead to branch accumulation'
                })
                recommendations.extend([
                    'Enable auto-delete to maintain security',
                    'Implement branch lifecycle management'
                ])

        # Configuration security
        if 'user' in config and 'nickname' in config['user']:
            nickname = config['user']['nickname']
            if len(nickname) > 20:
                risk_assessment.append({
                    'category': 'IDENTIFICATION_RISK',
                    'severity': 'LOW',
                    'description': 'Long nickname may cause display issues'
                })
                recommendations.extend([
                    'Keep nickname under 20 characters',
                    'Use professional identifiers'
                ])

        return {
            'findings': findings,
            'recommendations': recommendations,
            'risk_assessment': risk_assessment,
            'optimization_opportunities': optimizations
        }

    def prioritize_recommendations(self, recommendations: List[str]) -> List[str]:
        """Prioritize recommendations based on impact and feasibility"""
        # Simple prioritization logic
        high_priority = []
        medium_priority = []
        low_priority = []

        for rec in recommendations:
            if any(keyword in rec.lower() for keyword in ['security', 'high impact', 'critical']):
                high_priority.append(rec)
            elif any(keyword in rec.lower() for keyword in ['performance', 'optimization', 'efficiency']):
                medium_priority.append(rec)
            else:
                low_priority.append(rec)

        return high_priority + medium_priority + low_priority

    def rank_optimizations(self, optimizations: List[Dict]) -> List[Dict]:
        """Rank optimizations by priority and potential benefit"""
        # Sort by priority and potential benefit
        return sorted(optimizations, key=lambda x: (
            {'HIGH': 3, 'MEDIUM': 2, 'LOW': 1}.get(x['priority'], 1),
            {'HIGH': 3, 'MEDIUM': 2, 'LOW': 1}.get(x['potential_benefit'], 1)
        ), reverse=True)


def main():
    """Main function to demonstrate research integration"""
    engine = ConfigurationResearchEngine()

    # Load configuration
    config_path = Path(__file__).parent / "../../../.moai/config.json"
    if config_path.exists():
        try:
            with open(config_path, 'r') as f:
                config = json.load(f)

            # Perform comprehensive analysis
            research_results = engine.analyze_configuration(config)

            # Display results
            print("🔬 CONFIGURATION RESEARCH ANALYSIS")
            print("=" * 50)

            print("\n📊 FINDINGS:")
            for finding in research_results['findings']:
                print(f"  • {finding['category']}: {finding['analysis']}")

            print("\n💡 RECOMMENDATIONS:")
            for i, rec in enumerate(research_results['recommendations'], 1):
                print(f"  {i}. {rec}")

            print("\n⚠️ RISK ASSESSMENT:")
            for risk in research_results['risk_assessment']:
                print(f"  • {risk['category']}: {risk['description']} (Severity: {risk['severity']})")

            print("\n🚀 OPTIMIZATION OPPORTUNITIES:")
            for opt in research_results['optimization_opportunities']:
                print(f"  • {opt['name']}: {opt['potential_benefit']} (Priority: {opt['priority']})")

        except Exception as e:
            print(f"❌ Error loading configuration: {e}")
    else:
        print("⚠️ Configuration file not found")


if __name__ == "__main__":
    main()
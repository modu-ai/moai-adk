#!/usr/bin/env python3
"""
MoAI-ADK Agent System Validation Script
설계 문서와 실제 구현 간의 일관성을 자동으로 검증하는 도구
"""

import os
import json
import re
from pathlib import Path
from typing import Dict, List, Set, Tuple, Optional
from datetime import datetime
import yaml


class AgentSystemValidator:
    """MoAI-ADK 에이전트 시스템 검증기"""

    def __init__(self, project_root: str = None):
        self.project_root = Path(project_root or os.getcwd()).resolve()
        
        # MoAI-ADK 디렉토리 구조 감지
        if (self.project_root / "moai_adk").exists():
            # 상위 디렉토리에서 실행된 경우
            self.design_doc = self.project_root / "MoAI-ADK-Design-Final.md"
            self.src_templates = self.project_root / "moai_adk" / "src" / "moai_adk" / "templates"
            self.dist_templates = self.project_root / "moai_adk" / "dist" / "templates"
        elif (self.project_root / "src" / "moai_adk" / "templates").exists():
            # moai_adk 디렉토리에서 실행된 경우
            self.design_doc = self.project_root.parent / "MoAI-ADK-Design-Final.md"
            self.src_templates = self.project_root / "src" / "moai_adk" / "templates"
            self.dist_templates = self.project_root / "dist" / "templates"
        else:
            # 기본 경로
            self.design_doc = self.project_root / "MoAI-ADK-Design-Final.md"
            self.src_templates = self.project_root / "src" / "moai_adk" / "templates"
            self.dist_templates = self.project_root / "dist" / "templates"
        
        # 검증 결과 저장
        self.validation_results = {
            "timestamp": datetime.now().isoformat(),
            "design_agents": {},
            "src_agents": {},
            "dist_agents": {},
            "commands": {},
            "discrepancies": [],
            "summary": {}
        }

    def log_error(self, category: str, message: str):
        """검증 오류 로그"""
        self.validation_results["discrepancies"].append({
            "category": category,
            "message": message,
            "timestamp": datetime.now().isoformat()
        })
        print(f"❌ [{category}] {message}")

    def log_success(self, message: str):
        """검증 성공 로그"""
        print(f"✅ {message}")

    def log_warning(self, message: str):
        """검증 경고 로그"""
        print(f"⚠️  {message}")

    def extract_agents_from_design(self) -> Dict[str, Dict]:
        """설계 문서에서 에이전트 정보 추출"""
        if not self.design_doc.exists():
            self.log_error("DESIGN_DOC", f"Design document not found: {self.design_doc}")
            return {}

        with open(self.design_doc, 'r', encoding='utf-8') as f:
            content = f.read()

        agents = {}
        
        # YAML frontmatter에서 에이전트 추출
        yaml_match = re.search(r'^---\s*\n(.*?)\n---', content, re.MULTILINE | re.DOTALL)
        if yaml_match:
            try:
                yaml_content = yaml.safe_load(yaml_match.group(1))
                if 'agents' in yaml_content:
                    for agent_key, agent_info in yaml_content['agents'].items():
                        agents[agent_key] = {
                            'name': agent_info.get('name', agent_key),
                            'description': agent_info.get('description', ''),
                            'responsibility': agent_info.get('responsibility', ''),
                            'source': 'yaml_frontmatter'
                        }
            except yaml.YAMLError as e:
                self.log_error("YAML_PARSE", f"Failed to parse YAML frontmatter: {e}")

        # 테이블에서 에이전트 목록 추출
        table_pattern = r'\|\s*에이전트\s*\|\s*역할\s*\|(.*?)\n\n'
        table_match = re.search(table_pattern, content, re.DOTALL)
        if table_match:
            table_content = table_match.group(1)
            rows = re.findall(r'\|\s*([^|]+)\s*\|\s*([^|]+)\s*\|', table_content)
            for agent_name, role in rows:
                agent_name = agent_name.strip()
                role = role.strip()
                if agent_name not in ['---', '']:
                    agent_key = agent_name.lower().replace(' ', '-').replace('_', '-')
                    if agent_key not in agents:
                        agents[agent_key] = {
                            'name': agent_name,
                            'description': role,
                            'responsibility': '',
                            'source': 'table'
                        }

        # 상세 정의에서 에이전트 추출
        detail_pattern = r'### (\d+)\.\s*([^#\n]+)\n\n(.*?)(?=\n### |\n## |\Z)'
        detail_matches = re.findall(detail_pattern, content, re.DOTALL)
        for match in detail_matches:
            number, name, description = match
            agent_key = name.strip().lower().replace(' ', '-').replace('_', '-')
            if agent_key in agents:
                agents[agent_key]['detailed_description'] = description.strip()
                agents[agent_key]['has_detailed_definition'] = True
            else:
                agents[agent_key] = {
                    'name': name.strip(),
                    'description': '',
                    'detailed_description': description.strip(),
                    'responsibility': '',
                    'source': 'detailed_definition',
                    'has_detailed_definition': True
                }

        self.validation_results["design_agents"] = agents
        return agents

    def scan_template_agents(self, template_dir: Path) -> Dict[str, Dict]:
        """템플릿 디렉토리에서 에이전트 파일 스캔"""
        if not template_dir.exists():
            self.log_error("TEMPLATE_DIR", f"Template directory not found: {template_dir}")
            return {}

        agents = {}
        agents_dir = template_dir / ".claude" / "agents" / "moai"
        
        if not agents_dir.exists():
            self.log_error("AGENTS_DIR", f"Agents directory not found: {agents_dir}")
            return {}

        for agent_file in agents_dir.glob("*.md"):
            agent_key = agent_file.stem
            try:
                with open(agent_file, 'r', encoding='utf-8') as f:
                    content = f.read()
                
                # 첫 번째 줄에서 제목 추출
                lines = content.split('\n')
                title = lines[0].strip('#').strip() if lines else agent_key
                
                # 설명 추출 (첫 번째 문단)
                description = ""
                for line in lines[1:]:
                    if line.strip() and not line.startswith('#'):
                        description = line.strip()
                        break

                agents[agent_key] = {
                    'name': title,
                    'description': description,
                    'file_path': str(agent_file),
                    'file_size': agent_file.stat().st_size,
                    'content_preview': content[:200] + "..." if len(content) > 200 else content
                }
            except Exception as e:
                self.log_error("FILE_READ", f"Failed to read agent file {agent_file}: {e}")

        return agents

    def scan_commands(self, template_dir: Path) -> Dict[str, Dict]:
        """커맨드 파일 스캔"""
        if not template_dir.exists():
            return {}

        commands = {}
        commands_dir = template_dir / ".claude" / "commands" / "moai"
        
        if not commands_dir.exists():
            return {}

        for command_file in commands_dir.glob("*.md"):
            command_key = command_file.stem
            try:
                with open(command_file, 'r', encoding='utf-8') as f:
                    content = f.read()
                
                commands[command_key] = {
                    'file_path': str(command_file),
                    'file_size': command_file.stat().st_size,
                    'content_preview': content[:150] + "..." if len(content) > 150 else content
                }
            except Exception as e:
                self.log_error("COMMAND_READ", f"Failed to read command file {command_file}: {e}")

        return commands

    def validate_consistency(self) -> Dict:
        """전체 일관성 검증"""
        print("🔍 MoAI-ADK Agent System Validation Started")
        print("=" * 60)

        # 1. 설계 문서 분석
        print("📄 Analyzing design document...")
        design_agents = self.extract_agents_from_design()
        
        # 2. src 템플릿 스캔
        print("📁 Scanning src templates...")
        src_agents = self.scan_template_agents(self.src_templates)
        src_commands = self.scan_commands(self.src_templates)
        
        # 3. dist 템플릿 스캔
        print("📁 Scanning dist templates...")
        dist_agents = self.scan_template_agents(self.dist_templates)
        dist_commands = self.scan_commands(self.dist_templates)

        self.validation_results["src_agents"] = src_agents
        self.validation_results["dist_agents"] = dist_agents
        self.validation_results["commands"] = {
            "src": src_commands,
            "dist": dist_commands
        }

        # 4. 일관성 검증
        print("\n🔍 Validating consistency...")
        self._validate_agent_counts(design_agents, src_agents, dist_agents)
        self._validate_agent_completeness(design_agents, src_agents, dist_agents)
        self._validate_sync_status(src_agents, dist_agents, src_commands, dist_commands)
        
        # 5. 요약 생성
        self._generate_summary()
        
        return self.validation_results

    def _validate_agent_counts(self, design: Dict, src: Dict, dist: Dict):
        """에이전트 개수 검증"""
        design_count = len(design)
        src_count = len(src)
        dist_count = len(dist)
        
        print(f"\n📊 Agent Counts:")
        print(f"   Design Document: {design_count} agents")
        print(f"   Src Templates:   {src_count} agents")
        print(f"   Dist Templates:  {dist_count} agents")
        
        if design_count == src_count == dist_count:
            self.log_success(f"Agent counts are consistent ({design_count})")
        else:
            self.log_error("COUNT_MISMATCH", 
                         f"Agent count mismatch: Design({design_count}), Src({src_count}), Dist({dist_count})")

    def _validate_agent_completeness(self, design: Dict, src: Dict, dist: Dict):
        """에이전트 완성도 검증"""
        design_keys = set(design.keys())
        src_keys = set(src.keys())
        dist_keys = set(dist.keys())
        
        # 설계 문서에만 있는 에이전트
        design_only = design_keys - src_keys - dist_keys
        if design_only:
            for agent in design_only:
                self.log_error("MISSING_IMPLEMENTATION", 
                             f"Agent '{agent}' defined in design but not implemented")
        
        # src에만 있는 에이전트
        src_only = src_keys - design_keys
        if src_only:
            for agent in src_only:
                self.log_error("UNDOCUMENTED_AGENT", 
                             f"Agent '{agent}' in src but not in design document")
        
        # dist에만 있는 에이전트
        dist_only = dist_keys - design_keys
        if dist_only:
            for agent in dist_only:
                self.log_error("UNDOCUMENTED_AGENT", 
                             f"Agent '{agent}' in dist but not in design document")
        
        # 상세 정의 누락 검증
        detailed_agents = {k: v for k, v in design.items() if v.get('has_detailed_definition', False)}
        undefined_agents = design_keys - set(detailed_agents.keys())
        
        if undefined_agents:
            for agent in undefined_agents:
                self.log_warning(f"Agent '{agent}' lacks detailed definition in design document")

    def _validate_sync_status(self, src_agents: Dict, dist_agents: Dict, src_commands: Dict, dist_commands: Dict):
        """src와 dist 동기화 상태 검증"""
        src_agent_keys = set(src_agents.keys())
        dist_agent_keys = set(dist_agents.keys())
        
        missing_in_dist = src_agent_keys - dist_agent_keys
        missing_in_src = dist_agent_keys - src_agent_keys
        
        if missing_in_dist:
            for agent in missing_in_dist:
                self.log_error("SYNC_ISSUE", f"Agent '{agent}' in src but missing in dist")
        
        if missing_in_src:
            for agent in missing_in_src:
                self.log_error("SYNC_ISSUE", f"Agent '{agent}' in dist but missing in src")
        
        # 커맨드 동기화 검증
        src_cmd_keys = set(src_commands.keys())
        dist_cmd_keys = set(dist_commands.keys())
        
        cmd_missing_in_dist = src_cmd_keys - dist_cmd_keys
        cmd_missing_in_src = dist_cmd_keys - src_cmd_keys
        
        if cmd_missing_in_dist:
            for cmd in cmd_missing_in_dist:
                self.log_error("SYNC_ISSUE", f"Command '{cmd}' in src but missing in dist")
        
        if cmd_missing_in_src:
            for cmd in cmd_missing_in_src:
                self.log_error("SYNC_ISSUE", f"Command '{cmd}' in dist but missing in src")
        
        if not missing_in_dist and not missing_in_src and not cmd_missing_in_dist and not cmd_missing_in_src:
            self.log_success("Src and dist templates are in sync")

    def _generate_summary(self):
        """검증 결과 요약 생성"""
        total_issues = len(self.validation_results["discrepancies"])
        
        issue_categories = {}
        for issue in self.validation_results["discrepancies"]:
            category = issue["category"]
            issue_categories[category] = issue_categories.get(category, 0) + 1
        
        self.validation_results["summary"] = {
            "total_issues": total_issues,
            "issue_categories": issue_categories,
            "validation_status": "PASS" if total_issues == 0 else "FAIL",
            "design_agent_count": len(self.validation_results["design_agents"]),
            "src_agent_count": len(self.validation_results["src_agents"]),
            "dist_agent_count": len(self.validation_results["dist_agents"])
        }

    def save_report(self, output_file: str = None) -> str:
        """검증 리포트 저장"""
        if not output_file:
            timestamp = datetime.now().strftime("%Y%m%d_%H%M%S")
            output_file = f"validation_report_{timestamp}.json"
        
        output_path = self.project_root / output_file
        
        try:
            with open(output_path, 'w', encoding='utf-8') as f:
                json.dump(self.validation_results, f, indent=2, ensure_ascii=False)
            
            print(f"\n📋 Validation report saved to: {output_path}")
            return str(output_path)
        except Exception as e:
            self.log_error("REPORT_SAVE", f"Failed to save report: {e}")
            return ""

    def print_summary(self):
        """검증 결과 요약 출력"""
        print("\n" + "=" * 60)
        print("📋 VALIDATION SUMMARY")
        print("=" * 60)
        
        summary = self.validation_results["summary"]
        status = summary["validation_status"]
        
        if status == "PASS":
            print("✅ VALIDATION PASSED")
        else:
            print("❌ VALIDATION FAILED")
        
        print(f"\n📊 Statistics:")
        print(f"   Design Document Agents: {summary['design_agent_count']}")
        print(f"   Src Template Agents:    {summary['src_agent_count']}")
        print(f"   Dist Template Agents:   {summary['dist_agent_count']}")
        print(f"   Total Issues Found:     {summary['total_issues']}")
        
        if summary['total_issues'] > 0:
            print(f"\n🏷️  Issue Categories:")
            for category, count in summary['issue_categories'].items():
                print(f"   {category}: {count} issues")
        
        print("\n" + "=" * 60)


def main():
    """메인 실행 함수"""
    import argparse
    
    parser = argparse.ArgumentParser(description="MoAI-ADK Agent System Validator")
    parser.add_argument("--project-root", help="Project root directory")
    parser.add_argument("--save-report", help="Save report to file")
    parser.add_argument("--verbose", "-v", action="store_true", help="Verbose output")
    
    args = parser.parse_args()
    
    validator = AgentSystemValidator(args.project_root)
    
    try:
        # 검증 실행
        results = validator.validate_consistency()
        
        # 요약 출력
        validator.print_summary()
        
        # 리포트 저장
        if args.save_report:
            validator.save_report(args.save_report)
        
        # 종료 코드 설정
        exit_code = 0 if results["summary"]["validation_status"] == "PASS" else 1
        exit(exit_code)
        
    except KeyboardInterrupt:
        print("\n⚠️  Validation interrupted by user")
        exit(130)
    except Exception as e:
        print(f"❌ Validation failed with error: {e}")
        exit(1)


if __name__ == "__main__":
    main()
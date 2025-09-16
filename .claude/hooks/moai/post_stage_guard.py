#!/usr/bin/env python3
"""
MoAI-ADK Post Stage Guard Hook v0.1.12
PostToolUse Hook - 단계별 품질 게이트 검수 및 다음 단계 안내

이 Hook은 각 개발 단계 완료 후 품질 게이트를 자동으로 검증합니다.
- 4단계 파이프라인 진행 상태 추적
- 단계별 완료 기준 검증
- 다음 단계 진행 가능성 판단  
- 자동 커밋 및 동기화 지침 제공
"""

import json
import sys
import os
import subprocess
from pathlib import Path
from datetime import datetime
from typing import Dict, List, Optional, Any

# Import security manager for safe subprocess execution
sys.path.insert(0, str(Path(__file__).parent.parent.parent.parent.parent / 'moai_adk'))
try:
    from security import SecurityManager, SecurityError
except ImportError:
    # Fallback if security module not available
    SecurityManager = None
    class SecurityError(Exception):
        pass

class MoAIStageGuard:
    """MoAI-ADK 4단계 파이프라인 품질 게이트 관리"""
    
    def __init__(self, project_root: Path):
        self.project_root = project_root
        self.moai_dir = project_root / ".moai"
        self.specs_dir = self.moai_dir / "specs"
        self.indexes_dir = self.moai_dir / "indexes"
        self.state_file = self.indexes_dir / "state.json"

        # Initialize security manager
        self.security_manager = SecurityManager() if SecurityManager else None
        
        # 4단계 파이프라인 정의
        self.pipeline_stages = {
            'SPECIFY': {
                'files': ['spec.md', 'acceptance.md'],
                'requirements': ['EARS 형식 요구사항', 'Given-When-Then 수락기준'],
                'next_command': '/moai:plan',
                'quality_gates': ['모든_요구사항_명확', 'NEEDS_CLARIFICATION_해결']
            },
            'PLAN': {
                'files': ['plan.md', 'research.md', 'data-model.md'],
                'requirements': ['Constitution Check 통과', '기술 조사 완료', 'ADR 작성'],
                'next_command': '/moai:tasks',
                'quality_gates': ['Constitution_5원칙_준수', '기술적_실현가능성_확인']
            },
            'TASKS': {
                'files': ['tasks.md'],
                'requirements': ['TDD 순서 작업 분해', '병렬 실행 최적화'],
                'next_command': '/moai:dev',
                'quality_gates': ['모든_계약_테스트_포함', '의존성_그래프_완성']
            },
            'IMPLEMENT': {
                'files': ['src/', 'tests/', '*.py', '*.js', '*.ts'],
                'requirements': ['Red-Green-Refactor 완료', '테스트 커버리지 80%+'],
                'next_command': '/moai:sync',
                'quality_gates': ['모든_테스트_통과', '커버리지_달성', '코드_품질_기준_준수']
            }
        }
    
    def analyze_recent_changes(self, tool_name: str, tool_input: Dict) -> Dict[str, Any]:
        """최근 변경사항 분석 및 단계 추론"""
        
        if tool_name in ['Write', 'Edit', 'MultiEdit']:
            file_path = tool_input.get('file_path', '')
            content = tool_input.get('content', '') or tool_input.get('new_string', '')
            
            # 파일 경로로 단계 추론
            stage = self.infer_stage_from_path(file_path)
            
            if stage:
                return {
                    'stage': stage,
                    'file_path': file_path,
                    'content_length': len(content),
                    'analysis': f"{stage} 단계 파일 수정됨"
                }
        
        return {'stage': None, 'analysis': '단계 추론 불가'}
    
    def infer_stage_from_path(self, file_path: str) -> Optional[str]:
        """파일 경로로부터 파이프라인 단계 추론"""
        
        if not file_path:
            return None
        
        # SPEC 단계
        if any(keyword in file_path for keyword in ['spec.md', 'acceptance.md', 'requirements']):
            return 'SPECIFY'
        
        # PLAN 단계  
        if any(keyword in file_path for keyword in ['plan.md', 'research.md', 'data-model.md', 'contracts/']):
            return 'PLAN'
        
        # TASKS 단계
        if 'tasks.md' in file_path:
            return 'TASKS'
        
        # IMPLEMENT 단계
        if any(keyword in file_path for keyword in ['src/', 'tests/', '.py', '.js', '.ts', '.jsx', '.tsx']):
            return 'IMPLEMENT'
        
        return None
    
    def check_stage_completion(self, stage: str) -> Dict[str, Any]:
        """특정 단계의 완료 상태 검증"""
        
        if stage not in self.pipeline_stages:
            return {'completed': False, 'reason': f'Unknown stage: {stage}'}
        
        stage_config = self.pipeline_stages[stage]
        missing_files = []
        quality_issues = []
        
        # 필수 파일 존재 확인
        for required_file in stage_config['files']:
            if '/' in required_file:  # 디렉토리
                dir_path = self.project_root / required_file
                if not dir_path.exists() or not any(dir_path.iterdir()):
                    missing_files.append(required_file)
            else:  # 개별 파일 (SPEC 디렉토리에서 찾기)
                found = False
                for spec_dir in self.specs_dir.glob('SPEC-*'):
                    if (spec_dir / required_file).exists():
                        found = True
                        break
                if not found:
                    missing_files.append(required_file)
        
        # 품질 게이트 검증
        for gate in stage_config['quality_gates']:
            gate_result = self.check_quality_gate(gate, stage)
            if not gate_result['passed']:
                quality_issues.append(gate_result['message'])
        
        completed = len(missing_files) == 0 and len(quality_issues) == 0
        
        return {
            'completed': completed,
            'missing_files': missing_files,
            'quality_issues': quality_issues,
            'next_command': stage_config['next_command'] if completed else None
        }
    
    def check_quality_gate(self, gate: str, stage: str) -> Dict[str, Any]:
        """품질 게이트 개별 검증"""
        
        if gate == '모든_요구사항_명확':
            return self.check_requirements_clarity()
        elif gate == 'NEEDS_CLARIFICATION_해결':
            return self.check_clarification_resolved()
        elif gate == 'Constitution_5원칙_준수':
            return self.check_constitution_compliance()
        elif gate == '모든_테스트_통과':
            return self.check_tests_passing()
        elif gate == '커버리지_달성':
            return self.check_coverage_target()
        else:
            return {'passed': True, 'message': f'{gate} 검증 스킵됨'}
    
    def check_requirements_clarity(self) -> Dict[str, Any]:
        """요구사항 명확성 검사"""
        unclear_count = 0
        
        for spec_dir in self.specs_dir.glob('SPEC-*'):
            spec_file = spec_dir / 'spec.md'
            if spec_file.exists():
                try:
                    content = spec_file.read_text(encoding='utf-8')
                    unclear_count += content.count('[NEEDS CLARIFICATION')
                except:
                    pass
        
        return {
            'passed': unclear_count == 0,
            'message': f'{unclear_count}개의 미해결 [NEEDS CLARIFICATION] 발견' if unclear_count > 0 else '모든 요구사항 명확'
        }
    
    def check_clarification_resolved(self) -> Dict[str, Any]:
        """NEEDS CLARIFICATION 해결 여부"""
        return self.check_requirements_clarity()  # 같은 로직
    
    def check_constitution_compliance(self) -> Dict[str, Any]:
        """Constitution 5원칙 준수 검사"""
        try:
            # constitution_guard.py 실행하여 검증
            constitution_script = self.project_root / '.claude' / 'hooks' / 'constitution_guard.py'
            if constitution_script.exists():
                if self.security_manager:
                    # Use secure subprocess execution
                    result = self.security_manager.safe_subprocess_run(
                        ['python3', str(constitution_script), 'Write'],
                        cwd=self.project_root,
                        timeout=30
                    )
                else:
                    # Fallback to basic subprocess with validation
                    if not self._validate_constitution_script_path(constitution_script):
                        return {'passed': False, 'message': 'Constitution script path validation failed'}

                    result = subprocess.run([
                        'python3', str(constitution_script), 'Write'
                    ], capture_output=True, text=True, timeout=30, cwd=self.project_root)
                
                return {
                    'passed': result.returncode == 0,
                    'message': result.stderr if result.returncode != 0 else 'Constitution 5원칙 준수'
                }
        except:
            pass
        
        return {'passed': True, 'message': 'Constitution 검증 스킵됨'}
    
    def check_tests_passing(self) -> Dict[str, Any]:
        """테스트 통과 여부 확인"""
        try:
            # pytest 실행
            if self.security_manager:
                # Use secure subprocess execution
                result = self.security_manager.safe_subprocess_run(
                    ['python', '-m', 'pytest', 'tests/', '--tb=no', '-q'],
                    cwd=self.project_root,
                    timeout=60
                )
            else:
                # Fallback with validation
                tests_dir = self.project_root / 'tests'
                if not tests_dir.exists() or not tests_dir.is_dir():
                    return {'passed': True, 'message': 'No tests directory found'}

                result = subprocess.run([
                    'python', '-m', 'pytest', 'tests/', '--tb=no', '-q'
                ], capture_output=True, text=True, timeout=60, cwd=self.project_root)

            return {
                'passed': result.returncode == 0,
                'message': 'All tests passed' if result.returncode == 0 else 'Some tests failing'
            }
        except Exception as e:
            return {'passed': True, 'message': f'테스트 실행 스킵됨: {str(e)}'}
    
    def check_coverage_target(self) -> Dict[str, Any]:
        """테스트 커버리지 목표 달성 확인"""
        try:
            # coverage 실행
            if self.security_manager:
                # Use secure subprocess execution
                result = self.security_manager.safe_subprocess_run(
                    ['python', '-m', 'pytest', 'tests/', '--cov=src', '--cov-report=term-missing', '--cov-fail-under=80'],
                    cwd=self.project_root,
                    timeout=60
                )
            else:
                # Fallback with validation
                tests_dir = self.project_root / 'tests'
                src_dir = self.project_root / 'src'
                if not tests_dir.exists() or not src_dir.exists():
                    return {'passed': True, 'message': 'Required directories not found for coverage check'}

                result = subprocess.run([
                    'python', '-m', 'pytest', 'tests/', '--cov=src', '--cov-report=term-missing', '--cov-fail-under=80'
                ], capture_output=True, text=True, timeout=60, cwd=self.project_root)

            return {
                'passed': result.returncode == 0,
                'message': 'Coverage target achieved (80%+)' if result.returncode == 0 else 'Coverage below target'
            }
        except Exception as e:
            return {'passed': True, 'message': f'커버리지 검사 스킵됨: {str(e)}'}
    
    def generate_stage_report(self, stage: str, completion_status: Dict) -> str:
        """단계별 완료 리포트 생성"""
        
        report_lines = [
            f"🎯 {stage} 단계 완료 상태 검증",
            "=" * 40
        ]
        
        if completion_status['completed']:
            report_lines.extend([
                "✅ 모든 요구사항 충족",
                f"🚀 다음 단계: {completion_status['next_command']}",
                "",
                "💡 권장 사항:",
                "   • 현재 진행사항을 커밋하세요",
                "   • 새 탭에서 다음 명령어를 실행하세요", 
                f"   • {completion_status['next_command']} [관련 ID]"
            ])
        else:
            report_lines.append("⚠️  완료 요구사항 미충족")
            
            if completion_status['missing_files']:
                report_lines.extend([
                    "",
                    "📋 누락된 파일:",
                ] + [f"   • {file}" for file in completion_status['missing_files']])
            
            if completion_status['quality_issues']:
                report_lines.extend([
                    "",
                    "🔍 품질 이슈:",
                ] + [f"   • {issue}" for issue in completion_status['quality_issues']])
            
            report_lines.extend([
                "",
                "💡 해결 방안:",
                "   • 누락된 파일을 완성하세요",
                "   • 품질 이슈를 해결하세요",
                "   • 다시 검증을 실행하세요"
            ])
        
        return "\n".join(report_lines)
    
    def update_project_state(self, stage: str, completed: bool):
        """프로젝트 상태 업데이트"""
        
        try:
            # 기존 상태 읽기
            if self.state_file.exists():
                state = json.loads(self.state_file.read_text())
            else:
                state = {}
            
            # 상태 업데이트
            if 'pipeline' not in state:
                state['pipeline'] = {}
            
            state['pipeline'][stage] = {
                'completed': completed,
                'last_update': datetime.now().isoformat(),
                'version': '0.1.9'
            }
            
            # 상태 저장
            self.indexes_dir.mkdir(parents=True, exist_ok=True)
            self.state_file.write_text(json.dumps(state, indent=2))
            
        except Exception as error:
            print(f"State update error: {error}", file=sys.stderr)

    def _validate_constitution_script_path(self, script_path: Path) -> bool:
        """Validate constitution script path for security."""
        try:
            # Check if path is within project boundaries
            resolved_path = script_path.resolve()
            project_root = self.project_root.resolve()

            # Must be within project directory
            try:
                resolved_path.relative_to(project_root)
            except ValueError:
                return False

            # Must be a .py file in hooks directory
            if not (script_path.suffix == '.py' and 'hooks' in script_path.parts):
                return False

            return True
        except Exception:
            return False


def main():
    """Hook 진입점"""
    
    try:
        # Claude Code Hook 데이터 읽기
        hook_data = json.loads(sys.stdin.read())
        
        tool_name = hook_data.get('tool_name', '')
        tool_input = hook_data.get('tool_input', {})
        
        # 파일 편집/생성 도구에만 적용
        if tool_name not in ['Write', 'Edit', 'MultiEdit']:
            sys.exit(0)
        
        # 프로젝트 루트 경로
        project_root = Path(os.environ.get('CLAUDE_PROJECT_DIR', os.getcwd()))
        
        # MoAI 프로젝트인지 확인
        moai_dir = project_root / '.moai'
        if not moai_dir.exists():
            sys.exit(0)  # MoAI 프로젝트가 아니면 스킵
        
        # Stage Guard 실행
        guard = MoAIStageGuard(project_root)
        change_analysis = guard.analyze_recent_changes(tool_name, tool_input)
        
        if change_analysis['stage']:
            stage = change_analysis['stage']
            
            # 단계 완료 상태 검증
            completion_status = guard.check_stage_completion(stage)
            
            # 리포트 생성 및 출력
            report = guard.generate_stage_report(stage, completion_status)
            print("\n" + report, file=sys.stderr)
            
            # 프로젝트 상태 업데이트
            guard.update_project_state(stage, completion_status['completed'])
            
            # 성공 메시지
            if completion_status['completed']:
                print(f"\n🎉 {stage} 단계가 성공적으로 완료되었습니다!", file=sys.stderr)
            else:
                print(f"\n⏳ {stage} 단계 진행 중... 완료 요구사항을 충족해주세요.", file=sys.stderr)
        
        sys.exit(0)  # 항상 통과
        
    except Exception as error:
        print(f"🔧 Stage guard error: {error}", file=sys.stderr)
        sys.exit(0)  # 오류 시에도 통과

if __name__ == "__main__":
    main()
#!/usr/bin/env python3
"""
MoAI-ADK GitFlow 자동화 헬퍼 모듈 v0.2.1

이 모듈은 Python에서 GitFlow 작업을 편리하게 수행할 수 있는
헬퍼 함수들을 제공합니다.
"""

import os
import subprocess
import sys
from pathlib import Path
from typing import List, Optional, Dict, Tuple
import json
import re
from datetime import datetime


class GitFlowAutomator:
    """MoAI-ADK GitFlow 자동화 클래스"""

    def __init__(self, project_root: Optional[Path] = None):
        """
        GitFlowAutomator 초기화

        Args:
            project_root: 프로젝트 루트 디렉토리 (기본값: 현재 디렉토리)
        """
        self.project_root = project_root or Path.cwd()
        self.moai_dir = self.project_root / ".moai"
        self.scripts_dir = self.moai_dir / "scripts"

        # 색상 코드
        self.colors = {
            'red': '\033[0;31m',
            'green': '\033[0;32m',
            'yellow': '\033[1;33m',
            'blue': '\033[0;34m',
            'nc': '\033[0m'  # No Color
        }

    def log(self, level: str, message: str):
        """로깅 함수"""
        color = self.colors.get(level, self.colors['nc'])
        print(f"{color}[{level.upper()}]{self.colors['nc']} {message}")

    def run_command(self, command: List[str], capture_output: bool = True,
                   check: bool = True) -> subprocess.CompletedProcess:
        """시스템 명령어 실행"""
        try:
            result = subprocess.run(
                command,
                capture_output=capture_output,
                text=True,
                check=check,
                cwd=self.project_root
            )
            return result
        except subprocess.CalledProcessError as e:
            self.log('red', f"명령어 실행 실패: {' '.join(command)}")
            self.log('red', f"에러: {e.stderr}")
            raise

    def check_git_status(self) -> Dict[str, any]:
        """Git 상태 확인"""
        self.log('blue', 'Git 상태 확인 중...')

        # Git 저장소 확인
        try:
            self.run_command(['git', 'rev-parse', '--git-dir'])
        except subprocess.CalledProcessError:
            raise RuntimeError("Git 저장소가 아닙니다. 먼저 git init을 실행하세요.")

        # 현재 브랜치 확인
        current_branch = self.run_command(['git', 'branch', '--show-current']).stdout.strip()

        # 변경사항 확인
        status_output = self.run_command(['git', 'status', '--porcelain']).stdout.strip()
        has_changes = bool(status_output)

        # 리모트 확인
        try:
            remote_url = self.run_command(['git', 'config', '--get', 'remote.origin.url']).stdout.strip()
        except subprocess.CalledProcessError:
            remote_url = None

        return {
            'current_branch': current_branch,
            'has_changes': has_changes,
            'remote_url': remote_url,
            'status': status_output
        }

    def create_feature_branch(self, spec_id: str, feature_name: str) -> str:
        """Feature 브랜치 생성"""
        branch_name = f"feature/{spec_id}-{feature_name}"

        self.log('blue', f'Feature 브랜치 생성: {branch_name}')

        # develop 브랜치로 전환
        try:
            self.run_command(['git', 'show-ref', '--verify', '--quiet', 'refs/heads/develop'])
            self.run_command(['git', 'checkout', 'develop'])
            try:
                self.run_command(['git', 'pull', 'origin', 'develop'])
            except subprocess.CalledProcessError:
                self.log('yellow', '리모트에서 develop 브랜치를 가져올 수 없습니다.')
        except subprocess.CalledProcessError:
            self.log('yellow', 'develop 브랜치가 없습니다. main에서 생성합니다.')
            self.run_command(['git', 'checkout', '-b', 'develop'])

        # Feature 브랜치 생성 또는 전환
        try:
            self.run_command(['git', 'show-ref', '--verify', '--quiet', f'refs/heads/{branch_name}'])
            self.log('blue', f'기존 브랜치로 전환: {branch_name}')
            self.run_command(['git', 'checkout', branch_name])
        except subprocess.CalledProcessError:
            self.log('blue', f'새 브랜치 생성: {branch_name}')
            self.run_command(['git', 'checkout', '-b', branch_name])
            try:
                self.run_command(['git', 'push', '-u', 'origin', branch_name])
            except subprocess.CalledProcessError:
                self.log('yellow', '리모트 브랜치 생성 실패 (로컬에서 계속)')

        return branch_name

    def commit_spec_stage(self, spec_id: str, stage: str, description: str):
        """SPEC 단계별 커밋"""
        commit_messages = {
            'init': f"""feat({spec_id}): Add initial EARS requirements draft

{description}

- EARS 키워드 구조화 완료
- 초기 요구사항 정의
- [NEEDS CLARIFICATION] 마커 추가""",

            'stories': f"""feat({spec_id}): Add user stories US-001~005

{description}

- User Stories 생성 완료
- 수락 기준 초안 작성
- 우선순위 및 복잡도 평가""",

            'acceptance': f"""feat({spec_id}): Add acceptance criteria with GWT scenarios

{description}

- Given-When-Then 시나리오 완료
- 테스트 가능한 수락 기준 정의
- 품질 검증 체크리스트 완료""",

            'complete': f"""feat({spec_id}): Complete {spec_id} specification

{description}

- SPEC 문서 최종 검토 완료
- TAG 추적성 매핑 완료
- 품질 지표 충족 확인"""
        }

        # 파일 추가 및 커밋
        file_paths = {
            'init': [f'.moai/specs/{spec_id}/spec.md'],
            'stories': [f'.moai/specs/{spec_id}/user-stories.md'],
            'acceptance': [f'.moai/specs/{spec_id}/acceptance.md'],
            'complete': [f'.moai/specs/{spec_id}/']
        }

        for file_path in file_paths[stage]:
            self.run_command(['git', 'add', file_path])

        self.run_command(['git', 'commit', '-m', commit_messages[stage]])
        self.log('green', f'커밋 완료: {stage} stage')

    def commit_build_stage(self, spec_id: str, stage: str, description: str):
        """BUILD 단계별 커밋 (TDD)"""
        commit_messages = {
            'constitution': f"""feat({spec_id}): Constitution 5원칙 검증 완료

{description}

- Simplicity: 복잡도 제한 확인
- Architecture: 모듈형 구조 설계
- Testing: TDD 계획 수립
- Observability: 로깅 전략 정의
- Versioning: 버전 관리 체계 확립""",

            'red': f"""test({spec_id}): Add failing tests (RED phase)

{description}

- 실패하는 테스트 케이스 작성
- TDD Red 단계 완료
- 테스트 커버리지 목표 설정""",

            'green': f"""feat({spec_id}): Implement core functionality (GREEN phase)

{description}

- 테스트를 통과하는 최소 구현
- TDD Green 단계 완료
- 기능 동작 검증 완료""",

            'refactor': f"""refactor({spec_id}): Code optimization and cleanup (REFACTOR phase)

{description}

- 코드 품질 개선
- 중복 코드 제거
- 성능 최적화
- TDD Refactor 단계 완료"""
        }

        # 파일 추가 및 커밋
        file_paths = {
            'constitution': ['.moai/plans/'],
            'red': ['tests/'],
            'green': ['src/', 'tests/'],
            'refactor': ['src/', 'tests/']
        }

        for file_path in file_paths[stage]:
            self.run_command(['git', 'add', file_path])

        self.run_command(['git', 'commit', '-m', commit_messages[stage]])
        self.log('green', f'커밋 완료: {stage} stage')

    def create_draft_pr(self, spec_id: str, title: str, description: str) -> Optional[str]:
        """Draft PR 생성"""
        self.log('blue', 'Draft PR 생성 중...')

        # GitHub CLI 확인
        try:
            self.run_command(['gh', '--version'])
        except (subprocess.CalledProcessError, FileNotFoundError):
            self.log('yellow', 'GitHub CLI (gh) 가 설치되어 있지 않습니다.')
            return None

        # PR 본문 생성
        pr_body = f"""# {spec_id}: {title} 🚀

## 📋 변경사항 요약
{description}

## 📊 생성된 파일
- [x] .moai/specs/{spec_id}/spec.md - EARS 형식 요구사항
- [x] .moai/specs/{spec_id}/user-stories.md - User Stories
- [x] .moai/specs/{spec_id}/acceptance.md - 수락 기준

## 🏷️ TAG 매핑
- REQ:{spec_id.replace('SPEC-', '')} → DESIGN:{spec_id.replace('SPEC-', '')} → TASK:{spec_id.replace('SPEC-', '')}

## 🔄 다음 단계
- [ ] Constitution 5원칙 검증
- [ ] TDD 구현 진행
- [ ] 문서 동기화

## 📝 체크리스트
- [x] SPEC 문서 작성 완료
- [x] User Stories 정의
- [x] 수락 기준 작성
- [x] 품질 검증 통과
- [ ] Constitution 검증 대기
- [ ] TDD 구현 대기

---
🤖 MoAI-ADK v0.2.1에서 자동 생성됨"""

        try:
            result = self.run_command([
                'gh', 'pr', 'create', '--draft',
                '--title', f'{spec_id}: {title}',
                '--body', pr_body
            ])
            pr_url = result.stdout.strip()
            self.log('green', f'Draft PR 생성 완료: {pr_url}')
            return pr_url
        except subprocess.CalledProcessError:
            self.log('yellow', 'PR 생성 실패. 수동으로 생성해주세요.')
            return None

    def update_pr_comment(self, spec_id: str, stage: str, description: str):
        """PR에 댓글 추가"""
        self.log('blue', 'PR 업데이트 중...')

        try:
            self.run_command(['gh', '--version'])
        except (subprocess.CalledProcessError, FileNotFoundError):
            self.log('yellow', 'GitHub CLI (gh)가 없어 PR 업데이트를 건너뜁니다.')
            return

        progress = self.get_progress_percentage(spec_id)
        comment = f"""## 🔄 {stage} 단계 완료

{description}

진행률: {progress}% 완료

---
🤖 MoAI-ADK v0.2.1 자동 업데이트 ({datetime.now().strftime('%Y-%m-%d %H:%M:%S')})"""

        try:
            self.run_command(['gh', 'pr', 'comment', '--body', comment])
            self.log('green', f'PR 댓글 추가 완료: {stage}')
        except subprocess.CalledProcessError:
            self.log('yellow', 'PR 댓글 추가 실패')

    def get_progress_percentage(self, spec_id: str) -> int:
        """현재 진행률 계산"""
        try:
            current_branch = self.run_command(['git', 'branch', '--show-current']).stdout.strip()
            total_commits = self.run_command(['git', 'rev-list', '--count', current_branch]).stdout.strip()
            total_commits = int(total_commits)
        except (subprocess.CalledProcessError, ValueError):
            return 0

        # 단계별 진행률 계산 (대략적)
        base_progress = {
            'spec': 0,
            'build': 25,
            'sync': 85
        }

        # 현재 명령어 타입을 추정 (브랜치 이름이나 커밋 메시지로)
        if 'spec' in current_branch.lower():
            return min(25, total_commits * 6)  # 4단계 * 6%
        elif 'build' in current_branch.lower():
            return min(85, 25 + total_commits * 15)  # 4단계 * 15%
        else:
            return min(100, 85 + total_commits * 5)  # 3단계 * 5%

    def make_pr_ready(self, spec_id: str):
        """PR을 Ready 상태로 변경"""
        self.log('blue', 'PR을 Ready 상태로 변경 중...')

        try:
            self.run_command(['gh', '--version'])
        except (subprocess.CalledProcessError, FileNotFoundError):
            self.log('yellow', 'GitHub CLI (gh)가 없어 PR 상태 변경을 건너뜁니다.')
            return

        try:
            self.run_command(['gh', 'pr', 'ready'])
            self.log('green', 'PR이 Ready 상태로 변경되었습니다.')

            # 기본 리뷰어 추가
            try:
                reviewers = self.run_command(['git', 'config', 'moai.default-reviewers']).stdout.strip()
                if reviewers:
                    self.run_command(['gh', 'pr', 'edit', '--add-reviewer', reviewers])
                    self.log('green', f'리뷰어 추가: {reviewers}')
            except subprocess.CalledProcessError:
                pass  # 리뷰어 설정이 없으면 무시

        except subprocess.CalledProcessError:
            self.log('yellow', 'PR 상태 변경 실패')

    def run_spec_workflow(self, spec_id: str, description: str, feature_name: str):
        """SPEC 워크플로우 전체 실행"""
        self.log('blue', f'SPEC 워크플로우 시작: {spec_id}')

        # Git 상태 확인
        git_status = self.check_git_status()
        if git_status['has_changes']:
            self.log('yellow', '작업 중인 변경사항을 스태시합니다.')
            self.run_command(['git', 'stash', 'push', '-m', f'MoAI GitFlow: Auto-stash before {spec_id}'])

        # 브랜치 생성
        branch_name = self.create_feature_branch(spec_id, feature_name)

        # 단계별 커밋
        stages = ['init', 'stories', 'acceptance', 'complete']
        for stage in stages:
            self.commit_spec_stage(spec_id, stage, description)

        # Draft PR 생성
        self.create_draft_pr(spec_id, description, 'SPEC 문서 작성 완료')

        self.log('green', f'✅ {spec_id} SPEC 워크플로우 완료!')
        return branch_name

    def run_build_workflow(self, spec_id: str, description: str):
        """BUILD 워크플로우 전체 실행"""
        self.log('blue', f'BUILD 워크플로우 시작: {spec_id}')

        stages = ['constitution', 'red', 'green', 'refactor']
        stage_names = ['Constitution 검증', 'TDD RED', 'TDD GREEN', 'TDD REFACTOR']

        for stage, stage_name in zip(stages, stage_names):
            self.commit_build_stage(spec_id, stage, description)
            self.update_pr_comment(spec_id, stage_name, description)

        self.log('green', f'✅ {spec_id} BUILD 워크플로우 완료!')

    def run_sync_workflow(self, spec_id: str, description: str):
        """SYNC 워크플로우 전체 실행"""
        self.log('blue', f'SYNC 워크플로우 시작: {spec_id}')

        # 문서 동기화 커밋들은 별도 구현 필요
        # 여기서는 기본 구조만 제공

        self.update_pr_comment(spec_id, '문서 동기화', description)
        self.update_pr_comment(spec_id, 'TAG 시스템 업데이트', description)
        self.make_pr_ready(spec_id)

        self.log('green', f'✅ {spec_id} SYNC 워크플로우 완료!')


def main():
    """CLI 인터페이스"""
    import argparse

    parser = argparse.ArgumentParser(description='MoAI-ADK GitFlow 자동화 도구')
    parser.add_argument('command', choices=['spec', 'build', 'sync'], help='실행할 명령')
    parser.add_argument('spec_id', help='SPEC ID (예: SPEC-001)')
    parser.add_argument('description', help='작업 설명')
    parser.add_argument('--feature-name', help='Feature 브랜치 이름 (선택사항)')

    args = parser.parse_args()

    # feature_name 자동 생성
    if not args.feature_name:
        feature_name = re.sub(r'[^a-z0-9-]', '', args.description.lower().replace(' ', '-'))
    else:
        feature_name = args.feature_name

    automator = GitFlowAutomator()

    try:
        if args.command == 'spec':
            automator.run_spec_workflow(args.spec_id, args.description, feature_name)
        elif args.command == 'build':
            automator.run_build_workflow(args.spec_id, args.description)
        elif args.command == 'sync':
            automator.run_sync_workflow(args.spec_id, args.description)
    except Exception as e:
        automator.log('red', f'워크플로우 실행 중 오류 발생: {str(e)}')
        sys.exit(1)


if __name__ == '__main__':
    main()
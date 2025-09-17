#!/usr/bin/env python3
"""
MoAI-ADK Session Start Notice Hook - v0.1.12
세션 시작 시 프로젝트 상태 알림 및 컨텍스트 정보 제공

SessionStart Hook으로 현재 MoAI 프로젝트 상태를 분석하고 
개발자에게 유용한 정보를 제공합니다.
"""

import json
import sys
import os
from pathlib import Path
from typing import Dict, List, Any, Optional
from datetime import datetime

class SessionNotifier:
    """MoAI-ADK 세션 시작 알림 시스템"""
    
    def __init__(self, project_root: Path):
        self.project_root = project_root
        self.moai_config_path = project_root / ".moai" / "config.json"
        self.state_path = project_root / ".moai" / "indexes" / "state.json"
        self.tags_path = project_root / ".moai" / "indexes" / "tags.json"
    
    def get_project_status(self) -> Dict[str, Any]:
        """프로젝트 전체 상태 분석"""
        status = {
            "project_name": self.project_root.name,
            "moai_version": self.get_moai_version(),
            "initialized": self.is_moai_project(),
            "constitution_status": self.check_constitution_status(),
            "pipeline_stage": self.get_current_pipeline_stage(),
            "specs_count": self.count_specs(),
            "incomplete_specs": self.get_incomplete_specs(),
            "active_tasks": self.get_active_tasks(),
            "last_activity": self.get_last_activity(),
            "tag_health": self.analyze_tag_health()
        }
        
        return status
    
    def is_moai_project(self) -> bool:
        """MoAI 프로젝트 초기화 확인"""
        required_dirs = [
            ".moai",
            ".moai/steering",
            ".moai/specs",
            ".claude/commands/moai",
            ".claude/agents/moai"
        ]
        
        return all((self.project_root / dir_path).exists() for dir_path in required_dirs)
    
    def check_constitution_status(self) -> Dict[str, Any]:
        """Constitution 상태 확인"""
        constitution_path = self.project_root / ".moai" / "memory" / "constitution.md"
        checklist_path = self.project_root / ".moai" / "memory" / "constitution_update_checklist.md"
        
        return {
            "exists": constitution_path.exists(),
            "checklist_ready": checklist_path.exists(),
            "last_modified": self.get_file_mtime(constitution_path) if constitution_path.exists() else None
        }
    
    def get_current_pipeline_stage(self) -> Dict[str, Any]:
        """현재 파이프라인 단계 분석"""

        # steering 문서 먼저 체크
        if not self.has_steering_docs():
            return {"stage": "INIT", "description": "프로젝트 셋업 필요 (steering 문서 생성)"}

        specs_dir = self.project_root / ".moai" / "specs"

        if not specs_dir.exists():
            return {"stage": "SPECIFY", "description": "첫 번째 요구사항 작성 필요"}

        # 템플릿 디렉토리와 샘플 파일들 제외하고 실제 SPEC만 검사
        spec_dirs = [
            d for d in specs_dir.iterdir()
            if (d.is_dir()
                and not d.name.startswith("_")  # _templates 제외
                and not d.name.endswith("-sample")  # 샘플 파일 제외
                and d.name.startswith("SPEC-")  # SPEC- 패턴만 포함
            )
        ]

        if not spec_dirs:
            return {"stage": "SPECIFY", "description": "첫 번째 요구사항 작성 필요"}
        
        # 가장 최근 SPEC 디렉토리 분석
        latest_spec = max(spec_dirs, key=lambda d: d.stat().st_mtime)
        
        has_spec = (latest_spec / "spec.md").exists()
        has_plan = (latest_spec / "plan.md").exists()
        has_tasks = (latest_spec / "tasks.md").exists()
        
        if has_tasks:
            return {"stage": "IMPLEMENT", "description": f"구현 진행 중: {latest_spec.name}", "spec_id": latest_spec.name}
        elif has_plan:
            return {"stage": "TASKS", "description": f"작업 분해 필요: {latest_spec.name}", "spec_id": latest_spec.name}
        elif has_spec:
            return {"stage": "PLAN", "description": f"계획 수립 필요: {latest_spec.name}", "spec_id": latest_spec.name}
        else:
            return {"stage": "SPECIFY", "description": f"SPEC 작성 미완료: {latest_spec.name}", "spec_id": latest_spec.name}
    
    def count_specs(self) -> Dict[str, int]:
        """SPEC 개수 통계"""
        specs_dir = self.project_root / ".moai" / "specs"

        if not specs_dir.exists():
            return {"total": 0, "complete": 0, "incomplete": 0}

        # 실제 SPEC 디렉토리만 필터링 (템플릿, 샘플 제외)
        spec_dirs = [
            d for d in specs_dir.iterdir()
            if (d.is_dir()
                and not d.name.startswith("_")  # _templates 제외
                and not d.name.endswith("-sample")  # 샘플 파일 제외
                and d.name.startswith("SPEC-")  # SPEC- 패턴만 포함
            )
        ]

        total = len(spec_dirs)
        complete = 0

        for spec_dir in spec_dirs:
            # spec.md와 acceptance.md가 모두 있고 완성된 경우
            spec_file = spec_dir / "spec.md"
            acceptance_file = spec_dir / "acceptance.md"

            if spec_file.exists() and acceptance_file.exists():
                try:
                    # spec.md 내용 확인 (빈 파일이 아닌지)
                    with open(spec_file, 'r', encoding='utf-8') as f:
                        spec_content = f.read().strip()

                    # [NEEDS CLARIFICATION] 마커가 없고 실제 내용이 있는 경우만 완료로 처리
                    if spec_content and '[NEEDS CLARIFICATION' not in spec_content and len(spec_content) > 100:
                        complete += 1
                except:
                    pass

        return {
            "total": total,
            "complete": complete,
            "incomplete": total - complete
        }
    
    def get_incomplete_specs(self) -> List[str]:
        """미완료 SPEC 목록"""
        specs_dir = self.project_root / ".moai" / "specs"
        incomplete = []

        if not specs_dir.exists():
            return incomplete

        # 실제 SPEC 디렉토리만 필터링
        spec_dirs = [
            d for d in specs_dir.iterdir()
            if (d.is_dir()
                and not d.name.startswith("_")  # _templates 제외
                and not d.name.endswith("-sample")  # 샘플 파일 제외
                and d.name.startswith("SPEC-")  # SPEC- 패턴만 포함
            )
        ]

        for spec_dir in spec_dirs:
            spec_file = spec_dir / "spec.md"
            acceptance_file = spec_dir / "acceptance.md"

            # 파일이 없거나 미완료인 경우
            is_incomplete = False

            if not spec_file.exists() or not acceptance_file.exists():
                is_incomplete = True
            else:
                try:
                    with open(spec_file, 'r', encoding='utf-8') as f:
                        content = f.read().strip()

                    # [NEEDS CLARIFICATION] 마커가 있거나 내용이 부족한 경우
                    if '[NEEDS CLARIFICATION' in content or len(content) < 100:
                        is_incomplete = True
                except:
                    is_incomplete = True

            if is_incomplete:
                incomplete.append(spec_dir.name)

        return incomplete
    
    def get_active_tasks(self) -> Dict[str, Any]:
        """활성 작업 현황"""
        tasks_info = {"total": 0, "pending": 0, "in_progress": 0, "completed": 0}
        
        specs_dir = self.project_root / ".moai" / "specs"
        
        if not specs_dir.exists():
            return tasks_info
        
        for spec_dir in specs_dir.iterdir():
            if spec_dir.is_dir():
                tasks_file = spec_dir / "tasks.md"
                if tasks_file.exists():
                    try:
                        with open(tasks_file, 'r', encoding='utf-8') as f:
                            content = f.read()
                            # 간단한 작업 상태 파싱 (실제로는 더 정교한 파싱 필요)
                            tasks_info["total"] += content.count("T00")
                            tasks_info["completed"] += content.count("✅")
                            tasks_info["in_progress"] += content.count("🚧")
                    except:
                        pass
        
        tasks_info["pending"] = tasks_info["total"] - tasks_info["completed"] - tasks_info["in_progress"]
        return tasks_info
    
    def get_last_activity(self) -> Optional[str]:
        """최근 활동 시간"""
        if self.state_path.exists():
            try:
                with open(self.state_path, 'r', encoding='utf-8') as f:
                    state = json.load(f)
                    return state.get("last_activity")
            except:
                pass

        return None

    def get_last_commit_info(self) -> Optional[Dict[str, str]]:
        """최근 커밋 정보 조회"""
        try:
            import subprocess

            # Git 저장소인지 확인
            result = subprocess.run(
                ["git", "rev-parse", "--git-dir"],
                cwd=self.project_root,
                capture_output=True,
                text=True
            )
            if result.returncode != 0:
                return None

            # 최근 커밋 정보 가져오기
            result = subprocess.run(
                ["git", "log", "-1", "--pretty=format:%H|%s|%an|%ad", "--date=relative"],
                cwd=self.project_root,
                capture_output=True,
                text=True
            )

            if result.returncode == 0 and result.stdout.strip():
                parts = result.stdout.strip().split("|")
                if len(parts) >= 4:
                    return {
                        "hash": parts[0],
                        "message": parts[1],
                        "author": parts[2],
                        "date": parts[3]
                    }

        except Exception:
            pass

        return None

    def get_working_directory_status(self) -> Dict[str, Any]:
        """작업 디렉토리 상태 분석"""
        try:
            import subprocess

            # Git 상태 확인
            result = subprocess.run(
                ["git", "status", "--porcelain"],
                cwd=self.project_root,
                capture_output=True,
                text=True
            )

            if result.returncode == 0:
                lines = result.stdout.strip().split('\n') if result.stdout.strip() else []

                status = {
                    "clean": len(lines) == 0,
                    "modified": 0,
                    "added": 0,
                    "deleted": 0,
                    "untracked": 0
                }

                for line in lines:
                    if line.startswith(' M'):
                        status["modified"] += 1
                    elif line.startswith('A '):
                        status["added"] += 1
                    elif line.startswith(' D'):
                        status["deleted"] += 1
                    elif line.startswith('??'):
                        status["untracked"] += 1

                return status

        except Exception:
            pass

        return {"clean": True, "modified": 0, "added": 0, "deleted": 0, "untracked": 0}

    def get_smart_recommendations(self, pipeline: Dict[str, Any], git_status: Dict[str, Any],
                                specs: Dict[str, int], tasks: Dict[str, Any],
                                incomplete: List[str]) -> List[str]:
        """상황에 맞는 스마트 추천 생성"""
        recommendations = []

        # 1. Git 상태 기반 추천
        if not git_status["clean"]:
            recommendations.append("git add . && git commit -m 'WIP: 진행 중인 작업 임시 저장'  # 변경사항 커밋")

        # 2. 파이프라인 단계별 기본 추천
        if pipeline["stage"] == "INIT":
            if not self.has_steering_docs():
                recommendations.append("/moai:1-project init  # 프로젝트 초기화 및 steering 문서 생성")
            else:
                recommendations.append("/moai:2-spec '첫 번째 기능 요구사항'  # 첫 SPEC 작성")

        elif pipeline["stage"] == "SPECIFY":
            if incomplete:
                recommendations.append(f"/moai:2-spec {incomplete[0]}  # 미완료 SPEC 완성")
            else:
                recommendations.append("/moai:2-spec '새로운 기능 요구사항'  # 새 SPEC 작성")

        elif pipeline["stage"] == "PLAN":
            spec_id = pipeline.get("spec_id", "SPEC-001")
            recommendations.append(f"/moai:3-plan {spec_id}  # Constitution 검증 및 계획 수립")

        elif pipeline["stage"] == "TASKS":
            spec_id = pipeline.get("spec_id", "SPEC-001")
            recommendations.append(f"/moai:4-tasks {spec_id}  # TDD 작업 분해")

        elif pipeline["stage"] == "IMPLEMENT":
            if tasks["pending"] > 0:
                recommendations.append("/moai:5-dev T001  # 다음 작업 구현 (Red-Green-Refactor)")
            else:
                recommendations.append("/moai:6-sync  # 문서 동기화 및 TAG 정리")

        # 3. 상황별 추가 추천
        if specs["incomplete"] > 0:
            recommendations.append(f"# 📝 {specs['incomplete']}개의 미완료 SPEC이 있습니다")

        if tasks["in_progress"] > 1:
            recommendations.append("# ⚠️ 동시에 진행 중인 작업이 여러 개입니다. 집중하세요!")

        # 4. 품질 개선 추천
        specs_dir = self.project_root / ".moai" / "specs"
        if specs_dir.exists():
            spec_dirs = list(specs_dir.glob("SPEC-*/"))
            if len(spec_dirs) >= 3:
                recommendations.append("python .moai/scripts/check-traceability.py  # TAG 추적성 검증")

        return recommendations[:3]  # 최대 3개까지만 추천
    
    def analyze_tag_health(self) -> Dict[str, Any]:
        """TAG 시스템 건강도 분석"""
        if not self.tags_path.exists():
            return {"status": "not_initialized", "total_tags": 0}
        
        try:
            with open(self.tags_path, 'r', encoding='utf-8') as f:
                tags_data = json.load(f)
                
                total_tags = len(tags_data.get("tags", {}))
                orphan_tags = len(tags_data.get("orphan_tags", []))
                broken_chains = len(tags_data.get("broken_chains", []))
                
                health_score = max(0, 100 - (orphan_tags * 5) - (broken_chains * 10))
                
                return {
                    "status": "healthy" if health_score >= 80 else "needs_attention",
                    "health_score": health_score,
                    "total_tags": total_tags,
                    "orphan_tags": orphan_tags,
                    "broken_chains": broken_chains
                }
        except:
            return {"status": "error", "total_tags": 0}
    
    def has_steering_docs(self) -> bool:
        """steering 문서 존재 여부 확인"""
        steering_dir = self.project_root / ".moai" / "steering"
        if not steering_dir.exists():
            return False

        # vision.md, architecture.md, techstack.md 중 하나라도 있으면 True
        steering_files = ["vision.md", "architecture.md", "techstack.md"]
        return any((steering_dir / f).exists() for f in steering_files)

    def get_moai_version(self) -> str:
        """MoAI 버전 동적 조회"""
        version_path = self.project_root / ".moai" / "version.json"
        try:
            if version_path.exists():
                with open(version_path, 'r', encoding='utf-8') as f:
                    version_data = json.load(f)
                    return version_data.get("package_version", "unknown")
        except:
            pass
        return "unknown"

    def get_file_mtime(self, file_path: Path) -> Optional[str]:
        """파일 수정 시간"""
        try:
            if file_path.exists():
                mtime = file_path.stat().st_mtime
                return datetime.fromtimestamp(mtime).isoformat()
        except:
            pass
        return None
    
    def generate_notice(self) -> str:
        """세션 시작 알림 메시지 생성"""
        status = self.get_project_status()
        
        if not status["initialized"]:
            return self.generate_init_notice()
        
        return self.generate_status_notice(status)
    
    def generate_init_notice(self) -> str:
        """프로젝트 초기화 안내 메시지"""
        return f"""
🗿 MoAI-ADK 프로젝트가 감지되지 않았습니다.

📋 초기화 방법:
  1. 새 프로젝트: moai init project-name
  2. 기존 프로젝트: moai init .
  3. 대화형 설정: /moai:1-project init

💡 MoAI-ADK는 Spec-First TDD 개발을 지원합니다.
   Constitution 5원칙과 16-Core TAG 시스템으로 품질을 보장합니다.
"""
    
    def generate_status_notice(self, status: Dict[str, Any]) -> str:
        """프로젝트 상태 알림 메시지"""
        pipeline = status["pipeline_stage"]
        specs = status["specs_count"]
        incomplete = status["incomplete_specs"]
        tasks = status["active_tasks"]
        tag_health = status["tag_health"]
        
        message_parts = [
            f"🗿 MoAI-ADK 프로젝트: {status['project_name']}",
            ""
        ]
        
        # 파이프라인 상태
        stage_emoji = {
            "INIT": "🚀",
            "SPECIFY": "📝", 
            "PLAN": "📋",
            "TASKS": "⚡",
            "IMPLEMENT": "🔧"
        }
        
        current_emoji = stage_emoji.get(pipeline["stage"], "📍")
        message_parts.append(f"{current_emoji} 현재 단계: {pipeline['stage']} - {pipeline['description']}")
        
        # SPEC 통계
        if specs["total"] > 0:
            message_parts.append(f"📊 SPEC 현황: {specs['complete']}/{specs['total']} 완료")
            
            if incomplete:
                message_parts.append(f"⚠️  명확화 필요: {', '.join(incomplete[:3])}" + 
                                   ("..." if len(incomplete) > 3 else ""))
        
        # 작업 현황
        if tasks["total"] > 0:
            message_parts.append(f"🔧 작업 현황: {tasks['completed']} 완료, {tasks['in_progress']} 진행 중, {tasks['pending']} 대기")
        
        # TAG 건강도
        if tag_health["status"] != "not_initialized":
            if tag_health["health_score"] < 80:
                message_parts.append(f"🏷️  TAG 건강도: {tag_health['health_score']}% (개선 권장)")
            else:
                message_parts.append(f"🏷️  TAG 건강도: {tag_health['health_score']}% ✅")
        
        # 작업 디렉토리 상태
        git_status = self.get_working_directory_status()
        if not git_status["clean"]:
            status_parts = []
            if git_status["modified"] > 0:
                status_parts.append(f"수정 {git_status['modified']}개")
            if git_status["added"] > 0:
                status_parts.append(f"추가 {git_status['added']}개")
            if git_status["deleted"] > 0:
                status_parts.append(f"삭제 {git_status['deleted']}개")
            if git_status["untracked"] > 0:
                status_parts.append(f"미추적 {git_status['untracked']}개")

            if status_parts:
                message_parts.append(f"📝 작업 상태: {', '.join(status_parts)}")

        # 마지막 활동 정보
        last_commit = self.get_last_commit_info()
        if last_commit:
            message_parts.append(f"📅 마지막 커밋: {last_commit['hash'][:8]} - {last_commit['message']}")
            message_parts.append(f"   {last_commit['date']} ({last_commit['author']})")

        # 다음 단계 추천
        message_parts.extend(["", "💡 다음 단계 추천:"])

        recommendations = self.get_smart_recommendations(pipeline, git_status, specs, tasks, incomplete)
        for rec in recommendations:
            message_parts.append(f"   > {rec}")
        
        
        return "\n".join(message_parts)

def handle_session_start():
    """SessionStart Hook 메인 핸들러"""
    try:
        # 현재 디렉토리에서 시작해서 프로젝트 루트 찾기
        current_dir = Path.cwd()
        project_root = current_dir
        
        # .claude 또는 .moai 디렉토리를 찾을 때까지 상위로 올라가기
        max_depth = 10
        depth = 0
        
        while depth < max_depth:
            if (project_root / '.claude').exists() or (project_root / '.moai').exists():
                break
            
            parent = project_root.parent
            if parent == project_root:  # 루트 디렉토리에 도달
                break
                
            project_root = parent
            depth += 1
        
        # MoAI 관련 디렉토리가 있는지 확인
        has_claude = (project_root / '.claude').exists()
        has_moai = (project_root / '.moai').exists()
        
        if not (has_claude or has_moai):
            # 일반 프로젝트인 경우 간단한 안내만
            return
        
        notifier = SessionNotifier(project_root)
        notice = notifier.generate_notice()
        
        # 표준 출력으로 알림 출력 (Claude Code에서 사용자에게 표시됨)
        print(notice)
        
    except KeyboardInterrupt:
        pass
    except Exception as e:
        # 에러가 발생해도 세션을 방해하지 않음
        print(f"🗿 MoAI-ADK 상태 확인 중 오류가 발생했습니다: {e}", file=sys.stderr)

if __name__ == "__main__":
    handle_session_start()

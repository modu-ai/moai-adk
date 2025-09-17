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
        specs_dir = self.project_root / ".moai" / "specs"
        
        if not specs_dir.exists():
            return {"stage": "INIT", "description": "프로젝트 초기화 필요"}

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
            return {"stage": "SPECIFY", "description": "첫 번째 SPEC 생성 필요"}
        
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
        
        spec_dirs = [d for d in specs_dir.iterdir() if d.is_dir()]
        total = len(spec_dirs)
        complete = 0
        
        for spec_dir in spec_dirs:
            if (spec_dir / "spec.md").exists() and (spec_dir / "acceptance.md").exists():
                complete += 1
        
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
        
        for spec_dir in specs_dir.iterdir():
            if spec_dir.is_dir():
                spec_file = spec_dir / "spec.md"
                if spec_file.exists():
                    # [NEEDS CLARIFICATION] 마커 체크
                    try:
                        with open(spec_file, 'r', encoding='utf-8') as f:
                            content = f.read()
                            if '[NEEDS CLARIFICATION' in content:
                                incomplete.append(spec_dir.name)
                    except:
                        pass
        
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
        
        # 다음 단계 제안
        message_parts.extend(["", "💡 다음 단계:"])
        
        if pipeline["stage"] == "INIT":
            message_parts.append("   > /moai:1-project init  # 프로젝트 초기화")
        elif pipeline["stage"] == "SPECIFY":
            if self.has_steering_docs():
                message_parts.append("   > /moai:2-spec '기능 설명'  # 첫 SPEC 작성")
            else:
                message_parts.append("   > /moai:1-project init  # steering 문서 생성 필요")
        elif pipeline["stage"] == "PLAN":
            spec_id = pipeline.get("spec_id", "SPEC-001")
            message_parts.append(f"   > /moai:3-plan {spec_id}  # Constitution Check")
        elif pipeline["stage"] == "TASKS":
            spec_id = pipeline.get("spec_id", "SPEC-001")
            message_parts.append(f"   > /moai:4-tasks {spec_id}  # TDD 작업 분해")
        elif pipeline["stage"] == "IMPLEMENT":
            message_parts.append("   > /moai:5-dev T001  # Red-Green-Refactor 구현")
        
        # 유용한 명령어
        message_parts.extend([
            "",
            "🛠️  유용한 명령어:",
            "   > /moai:sync  # 문서 동기화",
            "   > python scripts/validate_stage.py  # 품질 검증",
            "   > python scripts/repair_tags.py  # TAG 자동 복구"
        ])
        
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

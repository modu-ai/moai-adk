#!/usr/bin/env python3
"""
MoAI-ADK Session Start Notice Hook - v0.1.0
세션 시작 시 프로젝트 상태 알림 및 컨텍스트 정보 제공

SessionStart Hook으로 현재 MoAI 프로젝트 상태를 분석하고
개발자에게 유용한 정보를 제공합니다.

@REQ:HOOK-SESSION-START-001
@FEATURE:SESSION-NOTICE-001
@API:HOOK-INTERFACE-001
@DESIGN:PROJECT-STATUS-001
@TECH:SESSIONSTART-HOOK-001
"""

import json
import sys
import os
import subprocess
from pathlib import Path
from typing import Dict, List, Any, Optional
from datetime import datetime

class SessionNotifier:
    """MoAI-ADK 세션 시작 알림 시스템

    @FEATURE:SESSION-NOTICE-001
    @API:SESSION-NOTIFIER-001
    """
    
    def __init__(self, project_root: Path):
        self.project_root = project_root
        self.moai_config_path = project_root / ".moai" / "config.json"
        self.state_path = project_root / ".moai" / "indexes" / "state.json"
        self.tags_path = project_root / ".moai" / "indexes" / "tags.json"
    
    def get_project_status(self) -> Dict[str, Any]:
        """프로젝트 전체 상태 분석

        @FEATURE:PROJECT-STATUS-001 @DATA:STATUS-ANALYSIS-001
        """
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

    def analyze_existing_project(self) -> Dict[str, Any]:
        """기존 프로젝트 구조 분석 및 초기화 전략 제안"""
        analysis = {
            "detected_language": None,
            "test_dirs": [],
            "config_files": [],
            "code_files": [],
            "complexity_score": 0,
            "suggested_specs": [],
            "initialization_strategy": "basic"
        }

        # 언어 감지
        language_indicators = {
            "python": ["*.py", "pyproject.toml", "requirements.txt", "setup.py"],
            "javascript": ["*.js", "package.json", "yarn.lock", "*.ts"],
            "go": ["*.go", "go.mod", "go.sum"],
            "rust": ["*.rs", "Cargo.toml", "Cargo.lock"],
            "java": ["*.java", "pom.xml", "build.gradle", "*.kt"],
            "csharp": ["*.cs", "*.csproj", "*.sln"]
        }

        for lang, patterns in language_indicators.items():
            for pattern in patterns:
                if list(self.project_root.glob(f"**/{pattern}")):
                    analysis["detected_language"] = lang
                    break
            if analysis["detected_language"]:
                break

        # 테스트 디렉토리 감지
        test_patterns = ["test*", "*test*", "tests", "spec", "specs"]
        for pattern in test_patterns:
            test_dirs = list(self.project_root.glob(f"**/{pattern}/"))
            analysis["test_dirs"].extend([str(d.relative_to(self.project_root)) for d in test_dirs])

        # 설정 파일 감지
        config_patterns = [
            "*.toml", "*.json", "*.yaml", "*.yml", "*.ini", "*.cfg",
            "Dockerfile", "docker-compose.*", "Makefile", ".gitignore"
        ]
        for pattern in config_patterns:
            config_files = list(self.project_root.glob(pattern))
            analysis["config_files"].extend([f.name for f in config_files])

        # 코드 파일 분석
        if analysis["detected_language"]:
            lang = analysis["detected_language"]
            if lang == "python":
                code_files = list(self.project_root.glob("**/*.py"))
            elif lang in ["javascript", "typescript"]:
                code_files = list(self.project_root.glob("**/*.js")) + list(self.project_root.glob("**/*.ts"))
            elif lang == "go":
                code_files = list(self.project_root.glob("**/*.go"))
            elif lang == "rust":
                code_files = list(self.project_root.glob("**/*.rs"))
            elif lang == "java":
                code_files = list(self.project_root.glob("**/*.java"))
            elif lang == "csharp":
                code_files = list(self.project_root.glob("**/*.cs"))
            else:
                code_files = []

            # .git, node_modules, __pycache__ 등 제외
            excluded_dirs = {".git", "node_modules", "__pycache__", ".venv", "venv", "target", "build", "dist"}
            code_files = [f for f in code_files if not any(excluded in f.parts for excluded in excluded_dirs)]

            analysis["code_files"] = [str(f.relative_to(self.project_root)) for f in code_files[:20]]  # 최대 20개만

            # 복잡도 점수 계산 (파일 수, 총 라인 수 기반)
            total_lines = 0
            for file in code_files[:50]:  # 최대 50개 파일만 분석
                try:
                    with open(file, 'r', encoding='utf-8', errors='ignore') as f:
                        total_lines += len(f.readlines())
                except:
                    pass

            analysis["complexity_score"] = min(100, (len(code_files) * 10 + total_lines) // 100)

        # 제안 SPEC 생성
        analysis["suggested_specs"] = self.generate_suggested_specs(analysis)

        # 초기화 전략 결정
        if analysis["complexity_score"] > 50:
            analysis["initialization_strategy"] = "complex"
        elif analysis["test_dirs"]:
            analysis["initialization_strategy"] = "tdd_ready"
        elif analysis["detected_language"]:
            analysis["initialization_strategy"] = "language_specific"

        return analysis

    def generate_suggested_specs(self, analysis: Dict[str, Any]) -> List[str]:
        """프로젝트 분석 결과를 바탕으로 초기 SPEC 제안"""
        suggestions = []

        lang = analysis.get("detected_language")
        complexity = analysis.get("complexity_score", 0)

        if lang == "python":
            if complexity > 30:
                suggestions.extend([
                    "기존 Python 코드 리팩토링 및 모듈 분리",
                    "pytest 기반 테스트 인프라 구축",
                    "타입 힌트 및 docstring 표준화"
                ])
            else:
                suggestions.extend([
                    "Python 프로젝트 품질 개선",
                    "단위 테스트 작성 및 커버리지 확보"
                ])
        elif lang in ["javascript", "typescript"]:
            if "package.json" in analysis.get("config_files", []):
                suggestions.extend([
                    "JavaScript/TypeScript 프로젝트 현대화",
                    "Jest 테스트 환경 구축",
                    "ESLint/Prettier 코드 품질 도구 설정"
                ])
        elif lang == "go":
            suggestions.extend([
                "Go 프로젝트 구조 최적화",
                "go test 기반 테스트 스위트 구축"
            ])

        if not analysis.get("test_dirs"):
            suggestions.append("TDD 개발 환경 구축 및 테스트 전략 수립")

        if complexity > 50:
            suggestions.append("대규모 코드베이스 점진적 리팩토링 계획")

        return suggestions[:3]  # 최대 3개 제안
    
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

        # 모든 SPEC의 상태 분석
        specs_analysis = []
        for spec_dir in spec_dirs:
            spec_file = spec_dir / "spec.md"
            plan_file = spec_dir / "plan.md"
            tasks_file = spec_dir / "tasks.md"

            status = "empty"
            needs_clarification = False

            if spec_file.exists():
                try:
                    with open(spec_file, 'r', encoding='utf-8') as f:
                        content = f.read().strip()

                    if '[NEEDS CLARIFICATION' in content:
                        needs_clarification = True
                        status = "needs_clarification"
                    elif len(content) > 500:
                        if tasks_file.exists():
                            status = "has_tasks"
                        elif plan_file.exists():
                            status = "has_plan"
                        else:
                            status = "spec_complete"
                    else:
                        status = "spec_incomplete"
                except:
                    status = "error"

            specs_analysis.append({
                "name": spec_dir.name,
                "status": status,
                "needs_clarification": needs_clarification,
                "mtime": spec_dir.stat().st_mtime
            })

        # 우선순위: 명확화 필요 > 미완료 SPEC > 완료된 SPEC 중 다음 단계
        clarification_needed = [s for s in specs_analysis if s["needs_clarification"]]
        if clarification_needed:
            spec = clarification_needed[0]
            return {"stage": "SPECIFY", "description": f"명확화 필요: {spec['name']}", "spec_id": spec['name']}

        incomplete_specs = [s for s in specs_analysis if s["status"] in ["empty", "spec_incomplete"]]
        if incomplete_specs:
            spec = incomplete_specs[0]
            return {"stage": "SPECIFY", "description": f"SPEC 작성 미완료: {spec['name']}", "spec_id": spec['name']}

        # 다음 단계가 필요한 SPEC 찾기
        spec_complete = [s for s in specs_analysis if s["status"] == "spec_complete"]
        if spec_complete:
            spec = max(spec_complete, key=lambda s: s["mtime"])
            return {"stage": "PLAN", "description": f"계획 수립 필요: {spec['name']}", "spec_id": spec['name']}

        has_plan = [s for s in specs_analysis if s["status"] == "has_plan"]
        if has_plan:
            spec = max(has_plan, key=lambda s: s["mtime"])
            return {"stage": "TASKS", "description": f"작업 분해 필요: {spec['name']}", "spec_id": spec['name']}

        has_tasks = [s for s in specs_analysis if s["status"] == "has_tasks"]
        if has_tasks:
            spec = max(has_tasks, key=lambda s: s["mtime"])
            return {"stage": "IMPLEMENT", "description": f"구현 진행 중: {spec['name']}", "spec_id": spec['name']}

        # 모든 SPEC이 완료된 경우
        return {"stage": "SYNC", "description": "문서 동기화 및 품질 검증 필요"}
    
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
            # spec.md 파일 존재 여부와 내용 확인
            spec_file = spec_dir / "spec.md"

            if spec_file.exists():
                try:
                    # spec.md 내용 확인 (빈 파일이 아닌지)
                    with open(spec_file, 'r', encoding='utf-8') as f:
                        spec_content = f.read().strip()

                    # [NEEDS CLARIFICATION] 마커가 없고 실제 내용이 있는 경우만 완료로 처리
                    if spec_content and '[NEEDS CLARIFICATION' not in spec_content and len(spec_content) > 500:
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

            # 파일이 없거나 미완료인 경우
            is_incomplete = False

            if not spec_file.exists():
                is_incomplete = True
            else:
                try:
                    with open(spec_file, 'r', encoding='utf-8') as f:
                        content = f.read().strip()

                    # [NEEDS CLARIFICATION] 마커가 있거나 내용이 부족한 경우
                    if '[NEEDS CLARIFICATION' in content or len(content) < 500:
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

    def get_next_pending_task(self) -> Optional[str]:
        """대기 중인 다음 작업 ID 찾기"""
        specs_dir = self.project_root / ".moai" / "specs"

        if not specs_dir.exists():
            return None

        for spec_dir in specs_dir.iterdir():
            if spec_dir.is_dir() and spec_dir.name.startswith("SPEC-"):
                tasks_file = spec_dir / "tasks.md"
                if tasks_file.exists():
                    try:
                        with open(tasks_file, 'r', encoding='utf-8') as f:
                            content = f.read()

                        # 간단한 작업 ID 추출 (T001, T002 등)
                        import re
                        pending_tasks = re.findall(r'(T\d{3})', content)
                        completed_tasks = re.findall(r'(T\d{3}).*✅', content)

                        # 완료되지 않은 첫 번째 작업 찾기
                        for task_id in pending_tasks:
                            if task_id not in completed_tasks:
                                return task_id
                    except:
                        pass

        return None

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

    def get_checkpoint_watcher_status(self) -> Dict[str, Any]:
        """자동 체크포인트 워처 상태 조회"""
        script = self.project_root / ".moai" / "scripts" / "checkpoint_watcher.py"
        if not script.exists():
            return {"available": False, "status": "missing"}

        try:
            result = subprocess.run(
                [sys.executable or "python3", str(script), "status"],
                cwd=self.project_root,
                capture_output=True,
                text=True,
                timeout=5,
            )
        except Exception as exc:
            return {"available": True, "status": "error", "message": str(exc)}

        output = (result.stdout or "").strip()
        errors = (result.stderr or "").strip()
        if result.returncode != 0:
            message = errors.splitlines()[0] if errors else (output or f"exit code {result.returncode}")
            lowered_err = errors.lower()
            if "filesystemeventhandler" in lowered_err or "watchdog" in lowered_err:
                message = "watchdog 모듈 미설치로 상태 확인 실패"
            return {"available": True, "status": "error", "message": message}

        lowered = output.lower()
        if "running" in lowered or "✅" in output:
            state = "running"
        elif "not running" in lowered or "❌" in output:
            state = "stopped"
        else:
            state = "unknown"
        message = output if output else (errors or "")
        return {"available": True, "status": state, "message": message}

    def get_smart_recommendations(self, pipeline: Dict[str, Any], git_status: Dict[str, Any],
                                specs: Dict[str, int], tasks: Dict[str, Any],
                                incomplete: List[str]) -> List[str]:
        """상황에 맞는 스마트 추천 생성"""
        recommendations = []

        # 시간 기반 컨텍스트 확인
        hour = datetime.now().hour
        is_work_hours = 9 <= hour <= 18

        # 최근 활동 분석
        last_commit = self.get_last_commit_info()
        recent_activity = last_commit and "minutes" in (last_commit.get("date", "") or "")

        # 1. 우선순위 알림 (긴급한 것부터)
        # Git 상태가 더러우면 먼저 정리
        if not git_status["clean"]:
            total_changes = git_status["modified"] + git_status["added"] + git_status["deleted"] + git_status["untracked"]
            if total_changes > 10:
                recommendations.append("git add . && git commit -m 'WIP: 대량 변경사항 임시 저장'  # ⚠️ 많은 변경사항 커밋 권장")
            else:
                recommendations.append("git add . && git commit -m 'WIP: 진행 중인 작업 임시 저장'  # 변경사항 커밋")

        # 2. 파이프라인 단계별 상황 인식 추천
        if pipeline["stage"] == "INIT":
            if not self.has_steering_docs():
                recommendations.append("moai init .  # 프로젝트 초기화 및 기본 설정")
            else:
                recommendations.append("/moai:1-spec '첫 번째 기능 요구사항'  # 첫 SPEC 작성")

        elif pipeline["stage"] == "SPECIFY":
            spec_id = pipeline.get("spec_id")
            if spec_id and "명확화 필요" in pipeline["description"]:
                # 명확화 필요한 SPEC 우선 처리
                recommendations.append(f"/moai:1-spec {spec_id}  # 🔍 명확화 마커 해결 (우선순위 높음)")
            elif spec_id:
                recommendations.append(f"/moai:1-spec {spec_id}  # SPEC 작성 완료")
            else:
                # 병렬 처리 제안
                if specs["total"] > 0:
                    recommendations.append("/moai:1-spec --project  # 🚀 프로젝트 전반 SPEC 대화형 생성")
                else:
                    recommendations.append("/moai:1-spec '새로운 기능 요구사항'  # 첫 SPEC 작성")

        elif pipeline["stage"] == "PLAN":
            spec_id = pipeline.get("spec_id", "SPEC-001")
            # Constitution 검증 필요성 강조
            recommendations.append(f"/moai:2-build {spec_id}  # Constitution 검증 및 TDD 구현 시작")

            # 계획 단계에서 추가 도움
            if not recent_activity and not is_work_hours:
                recommendations.append("# 💡 계획 단계는 충분한 시간을 가지고 진행하세요")

        elif pipeline["stage"] == "TASKS":
            spec_id = pipeline.get("spec_id", "SPEC-001")
            recommendations.append(f"/moai:2-build {spec_id}  # TDD 작업 분해 및 구현")

            # 작업 분해 후 즉시 구현 제안
            if specs["complete"] > 0:
                recommendations.append("# 다음: SPEC 완료 후 /moai:2-build로 구현 시작")

        elif pipeline["stage"] == "IMPLEMENT":
            if tasks["pending"] > 0:
                # 첫 번째 대기 중인 작업 찾기
                next_task = self.get_next_pending_task()
                if next_task:
                    recommendations.append(f"/moai:2-build  # 다음 작업 구현 (Red-Green-Refactor)")
                else:
                    recommendations.append("/moai:2-build  # 다음 작업 구현 (Red-Green-Refactor)")

                # 집중도 향상 제안
                if tasks["in_progress"] > 1:
                    recommendations.append("# ⚠️ 한 번에 하나의 작업에 집중하세요!")
            else:
                recommendations.append("/moai:3-sync  # 모든 작업 완료! 문서 동기화")

        elif pipeline["stage"] == "SYNC":
            recommendations.append("/moai:3-sync  # 문서 동기화 및 TAG 정리")

            # 추적성 검증 우선순위
            tag_health = self.analyze_tag_health()
            if tag_health.get("health_score", 100) < 80:
                recommendations.append("python .moai/scripts/check-traceability.py --repair  # TAG 추적성 복구")
            else:
                recommendations.append("python .moai/scripts/check-traceability.py  # TAG 추적성 검증")

        # 3. 상황별 지능형 추천
        # 미완료 작업이 많으면 집중 권고
        if specs["incomplete"] > 2:
            recommendations.append(f"# 📝 {specs['incomplete']}개의 미완료 SPEC - 우선순위를 정하고 집중하세요")
        elif specs["incomplete"] > 0:
            recommendations.append(f"# 📝 {specs['incomplete']}개의 미완료 SPEC이 있습니다")

        # 작업시간 외 권고사항
        if not is_work_hours and recent_activity:
            recommendations.append("# 🌙 늦은 시간 작업 중 - 충분한 휴식을 취하세요")

        # 4. 품질 및 성능 개선 추천
        specs_dir = self.project_root / ".moai" / "specs"
        if specs_dir.exists():
            spec_dirs = list(specs_dir.glob("SPEC-*/"))

            # 프로젝트 규모에 따른 추천
            if len(spec_dirs) >= 5:
                recommendations.append("# 🎯 대규모 프로젝트 - 정기적인 TAG 검증 권장")
            elif len(spec_dirs) >= 3:
                recommendations.append("python .moai/scripts/check-traceability.py  # TAG 추적성 검증")

        # 5. 개발 효율성 팁
        if len(recommendations) < 3:
            # 개발 팁 추가
            if pipeline["stage"] == "IMPLEMENT":
                recommendations.append("# 💡 TDD: Red → Green → Refactor 사이클을 지키세요")
            elif pipeline["stage"] == "PLAN":
                recommendations.append("# 💡 Constitution 5원칙을 염두에 두고 계획하세요")

        return recommendations[:3]  # 최대 3개까지만 추천

    def get_contextual_actions(self, pipeline: Dict[str, Any], git_status: Dict[str, Any],
                              specs: Dict[str, int], tasks: Dict[str, Any]) -> List[Dict[str, str]]:
        """현재 상황에 맞는 즉시 실행 가능한 액션 제공"""
        actions = []

        # 1. 긴급 상황 처리 (Git 상태 등)
        if not git_status["clean"]:
            total_changes = sum([git_status[k] for k in ["modified", "added", "deleted", "untracked"]])
            if total_changes > 10:
                actions.append({
                    "priority": "urgent",
                    "emoji": "💾",
                    "title": "대량 변경사항 커밋",
                    "command": "git add . && git commit -m 'WIP: 대량 변경사항 임시 저장'",
                    "description": f"{total_changes}개 파일 변경 - 안전을 위해 즉시 커밋 권장"
                })
            elif total_changes > 0:
                actions.append({
                    "priority": "high",
                    "emoji": "📝",
                    "title": "변경사항 커밋",
                    "command": "/moai:commit",
                    "description": f"{total_changes}개 파일 변경됨"
                })

        # 2. 테스트 실패 감지 및 수정 제안
        failed_tests = self.detect_test_failures()
        if failed_tests:
            actions.append({
                "priority": "high",
                "emoji": "🔴",
                "title": "테스트 실패 수정",
                "command": "/moai:fix-tests",
                "description": f"{len(failed_tests)}개 테스트 실패"
            })

        # 3. 파이프라인 단계별 주요 액션
        stage = pipeline["stage"]
        spec_id = pipeline.get("spec_id")

        if stage == "INIT":
            actions.append({
                "priority": "normal",
                "emoji": "🚀",
                "title": "프로젝트 초기화",
                "command": "moai init .",
                "description": "MoAI-ADK 프로젝트 설정"
            })

        elif stage == "SPECIFY":
            if spec_id and "명확화 필요" in pipeline.get("description", ""):
                actions.append({
                    "priority": "high",
                    "emoji": "🔍",
                    "title": "명확화 해결",
                    "command": f"/moai:1-spec {spec_id}",
                    "description": "SPEC 명확화 마커 해결"
                })
            else:
                actions.append({
                    "priority": "normal",
                    "emoji": "📝",
                    "title": "SPEC 작성",
                    "command": "/moai:1-spec '새로운 기능 요구사항'",
                    "description": "새 요구사항 명세 작성"
                })

        elif stage == "PLAN":
            actions.append({
                "priority": "normal",
                "emoji": "📋",
                "title": "구현 계획 수립",
                "command": f"/moai:2-build {spec_id or 'SPEC-001'}",
                "description": "TDD 구현 계획 및 시작"
            })

        elif stage == "IMPLEMENT":
            if tasks["pending"] > 0:
                actions.append({
                    "priority": "normal",
                    "emoji": "🔧",
                    "title": "다음 작업 구현",
                    "command": "/moai:2-build",
                    "description": f"{tasks['pending']}개 작업 대기 중"
                })

        elif stage == "SYNC":
            actions.append({
                "priority": "normal",
                "emoji": "🔄",
                "title": "문서 동기화",
                "command": "/moai:3-sync",
                "description": "문서 업데이트 및 TAG 정리"
            })

        # 4. 시간 기반 제안
        hour = datetime.now().hour
        session_duration = self.get_session_duration_minutes()

        if session_duration and session_duration > 120:  # 2시간 이상
            actions.append({
                "priority": "low",
                "emoji": "☕",
                "title": "휴식 권장",
                "command": "/moai:save-session",
                "description": f"{session_duration//60}시간 작업 중 - 휴식 후 재개"
            })

        if not (9 <= hour <= 18) and session_duration and session_duration > 60:
            actions.append({
                "priority": "low",
                "emoji": "🌙",
                "title": "작업 저장",
                "command": "/moai:save-session",
                "description": "늦은 시간 작업 - 안전하게 저장"
            })

        # 5. 품질 개선 제안
        if specs["total"] > 3:
            tag_health = self.analyze_tag_health()
            if tag_health.get("health_score", 100) < 80:
                actions.append({
                    "priority": "medium",
                    "emoji": "🏷️",
                    "title": "TAG 추적성 복구",
                    "command": "python .moai/scripts/check-traceability.py --repair",
                    "description": f"TAG 건강도 {tag_health.get('health_score', 0)}%"
                })

        # 우선순위별 정렬 및 최대 5개까지 반환
        priority_order = {"urgent": 0, "high": 1, "medium": 2, "normal": 3, "low": 4}
        actions.sort(key=lambda x: priority_order.get(x["priority"], 5))

        return actions[:5]

    def detect_test_failures(self) -> List[str]:
        """최근 테스트 실패 감지"""
        failures = []
        try:
            # 최근 테스트 결과 파일들 확인
            test_result_patterns = [
                ".pytest_cache/v/cache/lastfailed",
                "test-results.xml",
                "coverage.xml"
            ]

            for pattern in test_result_patterns:
                result_files = list(self.project_root.glob(f"**/{pattern}"))
                for file in result_files:
                    if file.exists():
                        # 파일이 최근에 수정되었는지 확인 (30분 이내)
                        if (datetime.now().timestamp() - file.stat().st_mtime) < 1800:
                            failures.append(str(file.relative_to(self.project_root)))
        except:
            pass

        return failures

    def get_session_duration_minutes(self) -> Optional[int]:
        """세션 지속 시간을 분 단위로 반환"""
        try:
            session_file = self.project_root / ".moai" / "session" / "current.json"
            if session_file.exists():
                with open(session_file, 'r', encoding='utf-8') as f:
                    session_data = json.load(f)
                    start_time = datetime.fromisoformat(session_data.get("start_time", ""))
                    duration = datetime.now() - start_time
                    return duration.seconds // 60
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

        # 실제 파일명에 맞춰 수정: product.md, structure.md, tech.md
        steering_files = ["product.md", "structure.md", "tech.md"]
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

    def generate_progress_bar(self, current: int, total: int, width: int = 10, filled: str = "█", empty: str = "░") -> str:
        """ASCII 진행률 바 생성"""
        if total == 0:
            return empty * width

        filled_width = int(width * current / total)
        return filled * filled_width + empty * (width - filled_width)

    def generate_dashboard(self, status: Dict[str, Any]) -> str:
        """실시간 대시보드 생성"""
        dashboard_lines = []

        # 헤더
        dashboard_lines.append("🗿 MoAI-ADK Dashboard")
        dashboard_lines.append("━" * 30)

        pipeline = status["pipeline_stage"]
        specs = status["specs_count"]
        tasks = status["active_tasks"]
        tag_health = status["tag_health"]

        # SPEC 진행률
        if specs["total"] > 0:
            spec_progress = self.generate_progress_bar(specs["complete"], specs["total"])
            spec_percentage = int(100 * specs["complete"] / specs["total"])
            dashboard_lines.append(f"📝 SPEC Progress: {spec_progress} {spec_percentage}% ({specs['complete']}/{specs['total']} 완료)")
        else:
            dashboard_lines.append("📝 SPEC Progress: 시작 전")

        # 작업 현황
        if tasks["total"] > 0:
            task_progress = self.generate_progress_bar(tasks["completed"], tasks["total"])
            task_percentage = int(100 * tasks["completed"] / tasks["total"]) if tasks["total"] > 0 else 0
            dashboard_lines.append(f"🔧 Task Progress: {task_progress} {task_percentage}% ({tasks['completed']}/{tasks['total']} 완료)")

            if tasks["in_progress"] > 0:
                dashboard_lines.append(f"⚡ 진행 중: {tasks['in_progress']}개 작업")
        else:
            dashboard_lines.append("🔧 Task Progress: 작업 없음")

        # TAG 건강도
        if tag_health["status"] != "not_initialized":
            health_score = tag_health.get("health_score", 0)
            health_bar = self.generate_progress_bar(health_score, 100)
            health_status = "✅" if health_score >= 80 else "⚠️"
            dashboard_lines.append(f"🏷️  TAG Health: {health_bar} {health_score}% {health_status}")
        else:
            dashboard_lines.append("🏷️  TAG Health: 미초기화")

        # 현재 단계
        stage_indicator = {
            "INIT": "🚀",
            "SPECIFY": "📝",
            "PLAN": "📋",
            "TASKS": "⚡",
            "IMPLEMENT": "🔧",
            "SYNC": "🔄"
        }
        current_emoji = stage_indicator.get(pipeline["stage"], "📍")
        dashboard_lines.append(f"{current_emoji} 현재 단계: {pipeline['stage']}")

        # 작업 시간 추적 (세션 지속시간)
        session_time = self.get_session_duration()
        if session_time:
            dashboard_lines.append(f"⏱️  세션 시간: {session_time}")

        return "\n".join(dashboard_lines)

    def get_session_duration(self) -> Optional[str]:
        """현재 세션 지속 시간 계산"""
        try:
            # 세션 시작 시간을 저장할 파일
            session_file = self.project_root / ".moai" / "session" / "current.json"
            if session_file.exists():
                with open(session_file, 'r', encoding='utf-8') as f:
                    session_data = json.load(f)
                    start_time = datetime.fromisoformat(session_data.get("start_time", ""))
                    duration = datetime.now() - start_time

                    hours = duration.seconds // 3600
                    minutes = (duration.seconds % 3600) // 60

                    if hours > 0:
                        return f"{hours}시간 {minutes}분"
                    else:
                        return f"{minutes}분"
        except:
            pass
        return None

    def save_session_start(self):
        """세션 시작 시간 저장"""
        try:
            session_dir = self.project_root / ".moai" / "session"
            session_dir.mkdir(parents=True, exist_ok=True)

            session_file = session_dir / "current.json"

            # 기존 세션 데이터가 있으면 로드
            existing_data = {}
            if session_file.exists():
                try:
                    with open(session_file, 'r', encoding='utf-8') as f:
                        existing_data = json.load(f)
                except:
                    pass

            # 현재 상태 수집
            status = self.get_project_status()
            pipeline = status["pipeline_stage"]

            session_data = {
                "start_time": existing_data.get("start_time", datetime.now().isoformat()),
                "last_activity": datetime.now().isoformat(),
                "project_name": self.project_root.name,
                "session_id": existing_data.get("session_id", datetime.now().strftime("%Y%m%d_%H%M%S")),
                "current_stage": pipeline["stage"],
                "current_spec": pipeline.get("spec_id"),
                "last_task": self.get_next_pending_task(),
                "specs_progress": status["specs_count"],
                "tasks_progress": status["active_tasks"],
                "git_status": self.get_working_directory_status()
            }

            with open(session_file, 'w', encoding='utf-8') as f:
                json.dump(session_data, f, ensure_ascii=False, indent=2)
        except:
            pass

    def get_session_context(self) -> Optional[Dict[str, Any]]:
        """이전 세션 컨텍스트 조회"""
        try:
            session_file = self.project_root / ".moai" / "session" / "current.json"
            if session_file.exists():
                with open(session_file, 'r', encoding='utf-8') as f:
                    return json.load(f)
        except:
            pass
        return None

    def generate_session_continuity_actions(self) -> List[Dict[str, str]]:
        """이전 세션 연속성 기반 액션 생성"""
        actions = []
        context = self.get_session_context()

        if not context:
            return actions

        last_activity = context.get("last_activity")
        if last_activity:
            try:
                last_time = datetime.fromisoformat(last_activity)
                time_diff = datetime.now() - last_time

                # 최근 활동이 30분 이내면 연속 작업 제안
                if time_diff.total_seconds() < 1800:  # 30분
                    current_spec = context.get("current_spec")
                    last_task = context.get("last_task")

                    if current_spec and last_task:
                        actions.append({
                            "priority": "high",
                            "emoji": "🔄",
                            "title": "이전 작업 계속",
                            "command": f"/moai:resume {current_spec} {last_task}",
                            "description": f"{current_spec} - {last_task} 작업 이어서 진행"
                        })
                    elif current_spec:
                        actions.append({
                            "priority": "normal",
                            "emoji": "📝",
                            "title": "SPEC 작업 재개",
                            "command": f"/moai:1-spec {current_spec}",
                            "description": f"{current_spec} 명세 작업 계속"
                        })

                # 하루 이상 지났으면 상태 동기화 제안
                elif time_diff.days >= 1:
                    actions.append({
                        "priority": "medium",
                        "emoji": "🔄",
                        "title": "프로젝트 상태 동기화",
                        "command": "/moai:3-sync",
                        "description": f"{time_diff.days}일 전 마지막 활동 - 상태 동기화 권장"
                    })

            except:
                pass

        return actions

    def save_work_context(self, spec_id: Optional[str] = None, task_id: Optional[str] = None):
        """현재 작업 컨텍스트 저장"""
        try:
            session_dir = self.project_root / ".moai" / "session"
            session_dir.mkdir(parents=True, exist_ok=True)

            context_file = session_dir / "work_context.json"

            context_data = {
                "timestamp": datetime.now().isoformat(),
                "current_spec": spec_id,
                "current_task": task_id,
                "git_branch": self.get_current_git_branch(),
                "uncommitted_changes": not self.get_working_directory_status()["clean"],
                "session_duration": self.get_session_duration_minutes()
            }

            with open(context_file, 'w', encoding='utf-8') as f:
                json.dump(context_data, f, ensure_ascii=False, indent=2)

        except:
            pass

    def get_current_git_branch(self) -> Optional[str]:
        """현재 Git 브랜치 조회"""
        try:
            import subprocess
            result = subprocess.run(
                ["git", "branch", "--show-current"],
                cwd=self.project_root,
                capture_output=True,
                text=True
            )
            if result.returncode == 0:
                return result.stdout.strip()
        except:
            pass
        return None

    def check_constitution_violations(self) -> List[Dict[str, str]]:
        """Constitution 5원칙 위반 사항 실시간 검증"""
        violations = []

        try:
            # Article I: Simplicity (단순성의 원칙) 검증
            violations.extend(self.check_simplicity_violations())

            # Article II: Architecture (아키텍처의 원칙) 검증
            violations.extend(self.check_architecture_violations())

            # Article III: Testing (테스트의 원칙) 검증
            violations.extend(self.check_testing_violations())

            # Article IV: Observability (관찰가능성의 원칙) 검증
            violations.extend(self.check_observability_violations())

            # Article V: Versioning (버전관리의 원칙) 검증
            violations.extend(self.check_versioning_violations())

        except Exception:
            pass

        return violations

    def check_simplicity_violations(self) -> List[Dict[str, str]]:
        """단순성 원칙 위반 검사"""
        violations = []

        # 모듈 수 확인 (최대 3개)
        module_count = self.count_project_modules()
        if module_count > 3:
            violations.append({
                "article": "Article I: Simplicity",
                "rule": "모듈 수 ≤ 3개",
                "violation": f"현재 {module_count}개 모듈 감지",
                "severity": "high",
                "fix_command": "/moai:refactor --consolidate-modules"
            })

        # 파일 크기 확인 (300 LOC 제한)
        large_files = self.find_large_files(300)
        if large_files:
            violations.append({
                "article": "Article I: Simplicity",
                "rule": "파일 크기 ≤ 300 LOC",
                "violation": f"{len(large_files)}개 파일이 제한 초과",
                "severity": "medium",
                "fix_command": f"/moai:refactor {large_files[0]} --split"
            })

        return violations

    def check_architecture_violations(self) -> List[Dict[str, str]]:
        """아키텍처 원칙 위반 검사"""
        violations = []

        # 계층 분리 확인
        if not self.has_layered_architecture():
            violations.append({
                "article": "Article II: Architecture",
                "rule": "계층형 아키텍처 준수",
                "violation": "Domain/Application/Infrastructure 분리 없음",
                "severity": "high",
                "fix_command": "/moai:1-spec '계층형 아키텍처 리팩토링'"
            })

        return violations

    def check_testing_violations(self) -> List[Dict[str, str]]:
        """테스트 원칙 위반 검사"""
        violations = []

        # 테스트 커버리지 확인
        coverage = self.get_test_coverage()
        if coverage is not None and coverage < 85:
            violations.append({
                "article": "Article III: Testing",
                "rule": "테스트 커버리지 ≥ 85%",
                "violation": f"현재 커버리지 {coverage}%",
                "severity": "high",
                "fix_command": "/moai:2-build --focus-tests"
            })

        # 테스트 파일 존재 여부
        if not self.has_test_files():
            violations.append({
                "article": "Article III: Testing",
                "rule": "TDD 필수",
                "violation": "테스트 파일 없음",
                "severity": "critical",
                "fix_command": "/moai:1-spec 'TDD 테스트 인프라 구축'"
            })

        return violations

    def check_observability_violations(self) -> List[Dict[str, str]]:
        """관찰가능성 원칙 위반 검사"""
        violations = []

        # 구조화된 로깅 확인
        if not self.has_structured_logging():
            violations.append({
                "article": "Article IV: Observability",
                "rule": "구조화된 로깅 의무화",
                "violation": "JSON 로깅 구조 없음",
                "severity": "medium",
                "fix_command": "/moai:1-spec '구조화된 로깅 시스템 구축'"
            })

        return violations

    def check_versioning_violations(self) -> List[Dict[str, str]]:
        """버전관리 원칙 위반 검사"""
        violations = []

        # 시맨틱 버저닝 확인
        if not self.has_semantic_versioning():
            violations.append({
                "article": "Article V: Versioning",
                "rule": "시맨틱 버저닝 의무화",
                "violation": "MAJOR.MINOR.BUILD 형식 없음",
                "severity": "low",
                "fix_command": "/moai:3-sync --setup-versioning"
            })

        return violations

    def count_project_modules(self) -> int:
        """프로젝트 모듈 수 계산"""
        try:
            # src, lib, components 등 주요 디렉토리 확인
            module_dirs = [
                "src", "lib", "components", "modules", "packages",
                "services", "controllers", "models", "views"
            ]

            count = 0
            for dir_name in module_dirs:
                module_path = self.project_root / dir_name
                if module_path.exists() and module_path.is_dir():
                    # 실제 코드 파일이 있는지 확인
                    if list(module_path.glob("**/*.py")) or list(module_path.glob("**/*.js")) or list(module_path.glob("**/*.go")):
                        count += 1

            return max(1, count)  # 최소 1개 모듈
        except:
            return 1

    def find_large_files(self, max_lines: int) -> List[str]:
        """제한을 초과하는 큰 파일들 찾기"""
        large_files = []
        try:
            # 주요 코드 파일 패턴
            patterns = ["**/*.py", "**/*.js", "**/*.ts", "**/*.go", "**/*.rs", "**/*.java"]

            for pattern in patterns:
                for file_path in self.project_root.glob(pattern):
                    # 제외할 디렉토리
                    if any(excluded in file_path.parts for excluded in [".git", "node_modules", "__pycache__", ".venv"]):
                        continue

                    try:
                        with open(file_path, 'r', encoding='utf-8', errors='ignore') as f:
                            line_count = len(f.readlines())
                            if line_count > max_lines:
                                large_files.append(str(file_path.relative_to(self.project_root)))
                    except:
                        continue

                if large_files:
                    break  # 첫 번째 언어에서 발견되면 중단

        except:
            pass

        return large_files[:5]  # 최대 5개만 반환

    def has_layered_architecture(self) -> bool:
        """계층형 아키텍처 존재 여부 확인"""
        try:
            # Domain, Application, Infrastructure 패턴 확인
            architectural_patterns = [
                ["domain", "application", "infrastructure"],
                ["models", "services", "controllers"],
                ["core", "application", "infrastructure"],
                ["entities", "usecases", "gateways"]
            ]

            for pattern in architectural_patterns:
                if all((self.project_root / dir_name).exists() for dir_name in pattern):
                    return True

            return False
        except:
            return False

    def get_test_coverage(self) -> Optional[float]:
        """테스트 커버리지 조회"""
        try:
            # coverage.xml 파일이나 .coverage 파일 확인
            coverage_files = [
                ".coverage",
                "coverage.xml",
                "htmlcov/index.html",
                "coverage-report.json"
            ]

            for coverage_file in coverage_files:
                file_path = self.project_root / coverage_file
                if file_path.exists():
                    # 간단한 커버리지 추출 (실제로는 더 정교한 파싱 필요)
                    try:
                        with open(file_path, 'r', encoding='utf-8') as f:
                            content = f.read()
                            # 간단한 패턴 매칭으로 커버리지 추출
                            import re
                            match = re.search(r'(\d+)%', content)
                            if match:
                                return float(match.group(1))
                    except:
                        continue

            return None
        except:
            return None

    def has_test_files(self) -> bool:
        """테스트 파일 존재 여부 확인"""
        try:
            test_patterns = [
                "test_*.py", "*_test.py", "test*.py",
                "*.test.js", "*.spec.js", "*test*.js",
                "*_test.go", "*test*.go"
            ]

            for pattern in test_patterns:
                if list(self.project_root.glob(f"**/{pattern}")):
                    return True

            return False
        except:
            return False

    def has_structured_logging(self) -> bool:
        """구조화된 로깅 존재 여부 확인"""
        try:
            # 로깅 라이브러리나 JSON 로깅 패턴 확인
            logging_indicators = [
                "import logging",
                "import structlog",
                "json.dumps",
                "logger.info",
                "console.log"
            ]

            code_files = list(self.project_root.glob("**/*.py")) + list(self.project_root.glob("**/*.js"))

            for file_path in code_files[:20]:  # 최대 20개 파일만 확인
                try:
                    with open(file_path, 'r', encoding='utf-8', errors='ignore') as f:
                        content = f.read()
                        if any(indicator in content for indicator in logging_indicators):
                            return True
                except:
                    continue

            return False
        except:
            return False

    def has_semantic_versioning(self) -> bool:
        """시맨틱 버저닝 설정 여부 확인"""
        try:
            version_files = [
                "pyproject.toml",
                "package.json",
                "Cargo.toml",
                "version.py",
                "__init__.py"
            ]

            for version_file in version_files:
                file_path = self.project_root / version_file
                if file_path.exists():
                    try:
                        with open(file_path, 'r', encoding='utf-8') as f:
                            content = f.read()
                            # MAJOR.MINOR.PATCH 패턴 확인
                            import re
                            if re.search(r'\d+\.\d+\.\d+', content):
                                return True
                    except:
                        continue

            return False
        except:
            return False

    def get_team_status(self) -> Dict[str, Any]:
        """팀 상태 정보 조회"""
        try:
            team_status_file = self.project_root / ".moai" / "team" / "status.json"
            if team_status_file.exists():
                with open(team_status_file, 'r', encoding='utf-8') as f:
                    return json.load(f)
        except:
            pass

        return {"team_members": [], "active_specs": {}, "conflicts": []}

    def update_team_status(self, member_name: str, current_spec: Optional[str] = None, current_task: Optional[str] = None):
        """개인 상태를 팀 상태에 업데이트"""
        try:
            team_dir = self.project_root / ".moai" / "team"
            team_dir.mkdir(parents=True, exist_ok=True)

            team_status_file = team_dir / "status.json"

            # 기존 팀 상태 로드
            team_status = self.get_team_status()

            # 현재 사용자 정보 업데이트
            member_info = {
                "name": member_name,
                "last_activity": datetime.now().isoformat(),
                "current_spec": current_spec,
                "current_task": current_task,
                "git_branch": self.get_current_git_branch(),
                "session_duration": self.get_session_duration_minutes()
            }

            # 기존 멤버 정보 업데이트 또는 새 멤버 추가
            updated = False
            for i, member in enumerate(team_status["team_members"]):
                if member["name"] == member_name:
                    team_status["team_members"][i] = member_info
                    updated = True
                    break

            if not updated:
                team_status["team_members"].append(member_info)

            # active_specs 업데이트
            if current_spec:
                if current_spec not in team_status["active_specs"]:
                    team_status["active_specs"][current_spec] = []

                # 현재 멤버가 이미 해당 SPEC에 있는지 확인
                if member_name not in team_status["active_specs"][current_spec]:
                    team_status["active_specs"][current_spec].append(member_name)

            # 충돌 감지
            team_status["conflicts"] = self.detect_team_conflicts(team_status)

            # 팀 상태 저장
            with open(team_status_file, 'w', encoding='utf-8') as f:
                json.dump(team_status, f, ensure_ascii=False, indent=2)

        except:
            pass

    def detect_team_conflicts(self, team_status: Dict[str, Any]) -> List[Dict[str, Any]]:
        """팀 작업 충돌 감지"""
        conflicts = []

        try:
            # 같은 SPEC을 동시에 작업하는 멤버들 감지
            for spec_id, members in team_status.get("active_specs", {}).items():
                if len(members) > 1:
                    # 최근 30분 이내에 활동한 멤버들만 확인
                    active_members = []
                    for member_data in team_status.get("team_members", []):
                        if member_data["name"] in members:
                            try:
                                last_activity = datetime.fromisoformat(member_data["last_activity"])
                                if (datetime.now() - last_activity).total_seconds() < 1800:  # 30분
                                    active_members.append(member_data["name"])
                            except:
                                continue

                    if len(active_members) > 1:
                        conflicts.append({
                            "type": "spec_conflict",
                            "spec_id": spec_id,
                            "members": active_members,
                            "description": f"{len(active_members)}명이 동시에 {spec_id} 작업 중",
                            "severity": "medium"
                        })

            # 같은 파일/모듈 수정 가능성 감지
            file_conflicts = self.detect_file_conflicts(team_status)
            conflicts.extend(file_conflicts)

        except:
            pass

        return conflicts

    def detect_file_conflicts(self, team_status: Dict[str, Any]) -> List[Dict[str, Any]]:
        """파일 수정 충돌 감지"""
        conflicts = []

        try:
            # Git 브랜치별 작업자 매핑
            branch_workers = {}
            for member in team_status.get("team_members", []):
                branch = member.get("git_branch")
                if branch and branch != "main" and branch != "master":
                    if branch not in branch_workers:
                        branch_workers[branch] = []
                    branch_workers[branch].append(member["name"])

            # 여러 명이 같은 브랜치에서 작업하는 경우
            for branch, workers in branch_workers.items():
                if len(workers) > 1:
                    conflicts.append({
                        "type": "branch_conflict",
                        "branch": branch,
                        "members": workers,
                        "description": f"브랜치 '{branch}'에서 {len(workers)}명 동시 작업",
                        "severity": "high"
                    })

        except:
            pass

        return conflicts

    def generate_team_status_message(self) -> List[str]:
        """팀 상태 메시지 생성"""
        lines = []
        team_status = self.get_team_status()

        if not team_status.get("team_members"):
            return lines

        # 활성 팀원 표시
        active_members = []
        for member in team_status["team_members"]:
            try:
                last_activity = datetime.fromisoformat(member["last_activity"])
                if (datetime.now() - last_activity).total_seconds() < 3600:  # 1시간 이내
                    active_members.append(member)
            except:
                continue

        if active_members:
            lines.append("👥 팀 상태:")
            for member in active_members[:3]:  # 최대 3명만 표시
                status_parts = []
                if member.get("current_spec"):
                    status_parts.append(f"SPEC: {member['current_spec']}")
                if member.get("git_branch") and member["git_branch"] not in ["main", "master"]:
                    status_parts.append(f"브랜치: {member['git_branch']}")

                status_text = " | ".join(status_parts) if status_parts else "대기 중"
                lines.append(f"   • {member['name']}: {status_text}")

        # 충돌 경고
        conflicts = team_status.get("conflicts", [])
        high_conflicts = [c for c in conflicts if c["severity"] == "high"]
        if high_conflicts:
            lines.append("   ⚠️ 충돌 감지:")
            for conflict in high_conflicts[:2]:  # 최대 2개만
                lines.append(f"      • {conflict['description']}")

        return lines

    def generate_notice(self) -> str:
        """세션 시작 알림 메시지 생성"""
        if os.environ.get("MOAI_SESSION_NOTICE_VERBOSE") == "1":
            status = self.get_project_status()
            if not status["initialized"]:
                return self.generate_simple_init_notice()
            return self.generate_simple_status_notice(status)

        return self.generate_quick_notice()

    def generate_quick_notice(self) -> str:
        """가벼운 요약만 제공하는 빠른 알림"""
        lines = [f"🗿 MoAI-ADK 프로젝트: {self.project_root.name}"]

        branch = self.get_current_git_branch()
        if branch:
            lines.append(f"🌿 현재 브랜치: {branch}")

        specs = self.count_specs()
        if specs["total"]:
            lines.append(
                f"📝 SPEC 진행률: {specs['complete']}/{specs['total']} (미완료 {specs['incomplete']}개)"
            )

        incomplete_specs = self.get_incomplete_specs()
        if incomplete_specs:
            lines.append(
                "⚠️  명확화 필요: " + ", ".join(incomplete_specs[:2]) + ("..." if len(incomplete_specs) > 2 else "")
            )

        git_status = self.get_working_directory_status()
        if not git_status["clean"]:
            total_changes = sum(
                git_status[k] for k in ["modified", "added", "deleted", "untracked"]
            )
            lines.append(f"📝 변경사항: {total_changes}개 파일")

        watcher = self.get_checkpoint_watcher_status()
        if watcher.get("available"):
            status = watcher.get("status")
            message = watcher.get("message") or "상태를 판별할 수 없습니다"
            if status == "running":
                lines.append("✅ 자동 체크포인트 워처 실행 중")
            elif status == "stopped":
                lines.append(
                    "⚠️ 자동 체크포인트 워처 미기동 → `python .moai/scripts/checkpoint_watcher.py start` 실행 권장"
                )
            elif status == "error":
                lines.append(f"⚠️ 워처 오류: {message}")
            else:
                lines.append(f"ℹ️ 워처 상태 확인 필요: {message}")
        else:
            lines.append("ℹ️ 자동 체크포인트 워처 스크립트를 찾을 수 없습니다.")

        lines.append("💡 상세 상태는 `MOAI_SESSION_NOTICE_VERBOSE=1` 환경변수 설정 후 재시작하거나 `/moai:status` 명령으로 확인하세요.")

        return "\n".join(lines)

    def generate_simple_init_notice(self) -> str:
        """간단한 프로젝트 초기화 안내 메시지 - 동적 정보 포함"""
        lines = []

        # Git 상태 정보
        current_branch = self.get_current_git_branch()
        if current_branch:
            lines.append(f"🌿 현재 브랜치: {current_branch}")

        # 마지막 커밋 정보
        last_commit = self.get_last_commit_info()
        if last_commit:
            lines.append(f"📅 마지막 커밋: {last_commit['hash'][:8]} - {last_commit['message'][:60]}")
            lines.append(f"👤 {last_commit['author']} ({last_commit['date']})")

        # 변경사항
        git_status = self.get_working_directory_status()
        if not git_status["clean"]:
            total_changes = sum([git_status[k] for k in ["modified", "added", "deleted", "untracked"]])
            if total_changes > 5:
                lines.append(f"📝 ⚠️  변경사항: {total_changes}개 파일 (커밋 권장)")
            else:
                lines.append(f"📝 변경사항: {total_changes}개 파일")

        watcher = self.get_checkpoint_watcher_status()
        if watcher.get("available"):
            status = watcher.get("status")
            message = watcher.get("message") or "상태를 판별할 수 없습니다"
            if status == "running":
                lines.append("✅ 자동 체크포인트 워처 실행 중")
            elif status == "stopped":
                lines.append("⚠️ 자동 체크포인트 워처 미기동 → `python .moai/scripts/checkpoint_watcher.py start`")
            elif status == "error":
                lines.append(f"⚠️ 워처 오류: {message}")
            else:
                lines.append(f"ℹ️ 워처 상태 확인 필요: {message}")

        # 언어 감지 및 도구 정보
        analysis = self.analyze_existing_project()
        if analysis["detected_language"]:
            lines.append(f"🌐 감지된 언어: {analysis['detected_language']}")

            # 권장 도구 정보 추가
            test_tool = self.get_recommended_test_tool(analysis['detected_language'])
            lint_tool = self.get_recommended_lint_tool(analysis['detected_language'])
            format_tool = self.get_recommended_format_tool(analysis['detected_language'])
            lines.append(f"🧪 권장 도구: test={test_tool}, lint={lint_tool}, format={format_tool}")

            # 테스트 파일 존재 여부
            if not analysis["test_dirs"]:
                lines.append("⚠️  테스트 디렉토리 없음 - TDD 환경 구축 필요")
            else:
                lines.append(f"🧪 테스트 디렉토리: {len(analysis['test_dirs'])}개")

        # 프로젝트 초기화가 필요한 경우만 안내
        if not lines:  # 아무 정보가 없는 경우에만
            lines.append("🗿 MoAI-ADK 프로젝트 초기화 권장")

        return "\n".join(lines)

    def generate_simple_status_notice(self, status: Dict[str, Any]) -> str:
        """스마트한 프로젝트 상태 알림 메시지"""
        lines = []

        # 현재 브랜치
        current_branch = self.get_current_git_branch()
        if current_branch:
            lines.append(f"🌿 현재 브랜치: {current_branch}")

        # 마지막 커밋 정보
        last_commit = self.get_last_commit_info()
        if last_commit:
            commit_msg = last_commit['message'][:60] + ("..." if len(last_commit['message']) > 60 else "")
            lines.append(f"📅 마지막 커밋: {last_commit['hash'][:8]} - {commit_msg}")
            lines.append(f"👤 {last_commit['author']} ({last_commit['date']})")

        # 변경사항
        git_status = self.get_working_directory_status()
        if not git_status["clean"]:
            total_changes = sum([git_status[k] for k in ["modified", "added", "deleted", "untracked"]])
            if total_changes > 10:
                lines.append(f"📝 ⚠️  대량 변경: {total_changes}개 파일 (즉시 커밋 권장)")
            elif total_changes > 5:
                lines.append(f"📝 ⚠️  변경사항: {total_changes}개 파일 (커밋 권장)")
            else:
                lines.append(f"📝 변경사항: {total_changes}개 파일")

        watcher = self.get_checkpoint_watcher_status()
        if watcher.get("available"):
            status = watcher.get("status")
            message = watcher.get("message") or "상태를 판별할 수 없습니다"
            if status == "running":
                lines.append("✅ 자동 체크포인트 워처 실행 중")
            elif status == "stopped":
                lines.append("⚠️ 자동 체크포인트 워처가 꺼져 있습니다 → `python .moai/scripts/checkpoint_watcher.py start`")
            elif status == "error":
                lines.append(f"⚠️ 워처 오류: {message}")
            else:
                lines.append(f"ℹ️ 워처 상태 확인 필요: {message}")

        # 현재 작업 상태 및 다음 액션 제안
        pipeline = status["pipeline_stage"]
        if pipeline.get("spec_id"):
            stage_emoji = {
                "SPECIFY": "📝", "PLAN": "📋", "TASKS": "⚡",
                "IMPLEMENT": "🔧", "SYNC": "🔄"
            }.get(pipeline["stage"], "📍")

            lines.append(f"{stage_emoji} 현재 작업: {pipeline['spec_id']} - {pipeline['description']}")

            # 다음 액션 제안
            if pipeline["stage"] == "SPECIFY":
                lines.append("💡 다음: /moai:1-spec 으로 명세 작성 완료")
            elif pipeline["stage"] == "PLAN":
                lines.append("💡 다음: /moai:2-build 로 TDD 구현 시작")
            elif pipeline["stage"] == "IMPLEMENT":
                lines.append("💡 다음: /moai:2-build 로 구현 계속")
            elif pipeline["stage"] == "SYNC":
                lines.append("💡 다음: /moai:3-sync 로 문서 동기화")

        # SPEC 진행률 정보
        specs = status["specs_count"]
        if specs["total"] > 0:
            progress = f"{specs['complete']}/{specs['total']}"
            if specs["incomplete"] > 0:
                lines.append(f"📊 SPEC 진행률: {progress} ({specs['incomplete']}개 미완료)")
            else:
                lines.append(f"📊 SPEC 진행률: {progress} (모두 완료!)")

        # 테스트 실패 감지
        failed_tests = self.detect_test_failures()
        if failed_tests:
            lines.append(f"🔴 테스트 실패 감지: {len(failed_tests)}개 파일")

        # Constitution 위반 간단 체크
        violations = self.check_constitution_violations()
        critical_violations = [v for v in violations if v["severity"] in ["critical", "high"]]
        if critical_violations:
            lines.append(f"⚠️  Constitution 위반: {len(critical_violations)}개 (수정 필요)")

        return "\n".join(lines) if lines else "🗿 MoAI-ADK 프로젝트 준비 완료"

    def get_recommended_test_tool(self, language: str) -> str:
        """언어별 추천 테스트 도구"""
        tools = {
            "python": "pytest",
            "javascript": "jest",
            "typescript": "jest",
            "go": "go test",
            "rust": "cargo test",
            "java": "junit",
            "csharp": "dotnet test"
        }
        return tools.get(language, "test")

    def get_recommended_lint_tool(self, language: str) -> str:
        """언어별 추천 린트 도구"""
        tools = {
            "python": "ruff",
            "javascript": "eslint",
            "typescript": "eslint",
            "go": "gofmt",
            "rust": "rustfmt",
            "java": "spotless",
            "csharp": "dotnet format"
        }
        return tools.get(language, "lint")

    def get_recommended_format_tool(self, language: str) -> str:
        """언어별 추천 포맷 도구"""
        tools = {
            "python": "black",
            "javascript": "prettier",
            "typescript": "prettier",
            "go": "gofmt",
            "rust": "rustfmt",
            "java": "google-java-format",
            "csharp": "dotnet format"
        }
        return tools.get(language, "format")

    def generate_init_notice(self) -> str:
        """프로젝트 초기화 안내 메시지"""
        # 기존 프로젝트 분석
        analysis = self.analyze_existing_project()

        lines = ["🗿 MoAI-ADK 프로젝트가 감지되지 않았습니다.", ""]

        # 기존 프로젝트 구조 분석 결과 표시
        if analysis["detected_language"]:
            lines.extend([
                f"🔍 프로젝트 분석 결과:",
                f"   📁 언어: {analysis['detected_language']}",
                f"   📊 복잡도: {analysis['complexity_score']}/100",
                f"   📂 코드 파일: {len(analysis['code_files'])}개"
            ])

            if analysis["test_dirs"]:
                lines.append(f"   🧪 테스트 디렉토리: {', '.join(analysis['test_dirs'])}")
            else:
                lines.append("   ⚠️  테스트 디렉토리 없음 - TDD 환경 구축 필요")

            lines.append("")

        # 초기화 전략별 맞춤형 가이드
        strategy = analysis["initialization_strategy"]

        if strategy == "complex":
            lines.extend([
                "🎯 복잡한 프로젝트가 감지되었습니다:",
                "  1. moai init . --complex  # 점진적 리팩토링 모드",
                "  2. /moai:1-spec '기존 코드 리팩토링 및 모듈 분리'",
                "  3. Constitution 5원칙 단계별 적용"
            ])
        elif strategy == "tdd_ready":
            lines.extend([
                "🧪 테스트 환경이 준비된 프로젝트입니다:",
                "  1. moai init . --tdd  # 기존 테스트 통합",
                "  2. /moai:1-spec '테스트 커버리지 확대 및 품질 개선'",
                "  3. 즉시 TDD 사이클 시작 가능"
            ])
        elif strategy == "language_specific":
            lang = analysis["detected_language"]
            if lang == "python":
                lines.extend([
                    "🐍 Python 프로젝트 최적화:",
                    "  1. moai init . --python  # pytest, ruff, black 설정",
                    "  2. /moai:1-spec 'Python 코드 품질 개선 및 타입 힌트'",
                    "  3. 즉시 사용 가능한 도구: pytest, ruff, black"
                ])
            elif lang in ["javascript", "typescript"]:
                lines.extend([
                    "⚡ JavaScript/TypeScript 프로젝트:",
                    "  1. moai init . --js  # Jest, ESLint, Prettier 설정",
                    "  2. /moai:1-spec 'JS/TS 프로젝트 현대화'",
                    "  3. 즉시 사용: npm test, eslint, prettier"
                ])
        else:
            lines.extend([
                "📋 표준 초기화 방법:",
                "  1. 새 프로젝트: moai init project-name",
                "  2. 기존 프로젝트: moai init .",
                "  3. 대화형 설정: /moai:1-spec '첫 번째 기능 요구사항'"
            ])

        # 제안 SPEC 표시
        if analysis["suggested_specs"]:
            lines.extend([
                "",
                "💡 추천 초기 SPEC:",
            ])
            for i, spec in enumerate(analysis["suggested_specs"], 1):
                lines.append(f"  {i}. {spec}")

        lines.extend([
            "",
            "🗿 MoAI-ADK 특징:",
            "   • Spec-First TDD 자동화 (Git을 몰라도 프로급 워크플로우)",
            "   • Constitution 5원칙으로 품질 보장",
            "   • 16-Core TAG 시스템으로 완전 추적성"
        ])

        return "\n".join(lines)
    
    def generate_status_notice(self, status: Dict[str, Any]) -> str:
        """프로젝트 상태 알림 메시지"""
        pipeline = status["pipeline_stage"]
        specs = status["specs_count"]
        incomplete = status["incomplete_specs"]
        tasks = status["active_tasks"]
        tag_health = status["tag_health"]
        git_status = self.get_working_directory_status()

        # 세션 시작 시간 저장
        self.save_session_start()

        message_parts = [f"🗿 MoAI-ADK 프로젝트: {status['project_name']}", ""]

        # 실시간 대시보드 통합
        dashboard = self.generate_dashboard(status)
        message_parts.extend(dashboard.split('\n'))
        message_parts.append("")

        # 현재 상태 요약
        current_emoji = {
            "INIT": "🚀", "SPECIFY": "📝", "PLAN": "📋",
            "TASKS": "⚡", "IMPLEMENT": "🔧", "SYNC": "🔄"
        }.get(pipeline["stage"], "📍")

        message_parts.append(f"{current_emoji} 현재: {pipeline['description']}")

        # 미완료 SPEC 강조
        if incomplete:
            message_parts.append(f"⚠️  명확화 필요: {', '.join(incomplete[:3])}" +
                               ("..." if len(incomplete) > 3 else ""))

        # 작업 디렉토리 상태
        if not git_status["clean"]:
            total_changes = sum([git_status[k] for k in ["modified", "added", "deleted", "untracked"]])
            if total_changes > 10:
                message_parts.append(f"📝 ⚠️  대량 변경: {total_changes}개 파일 (즉시 커밋 권장)")
            else:
                message_parts.append(f"📝 변경사항: {total_changes}개 파일")

        # 마지막 활동
        last_commit = self.get_last_commit_info()
        if last_commit:
            message_parts.append(f"📅 최근: {last_commit['hash'][:8]} - {last_commit['message'][:50]}")

        # Quick Actions 섹션
        message_parts.extend(["", "⚡ Quick Actions:"])

        # 세션 연속성 액션 우선 표시
        continuity_actions = self.generate_session_continuity_actions()
        contextual_actions = self.get_contextual_actions(pipeline, git_status, specs, tasks)

        # 연속성 액션을 우선순위에 따라 통합
        all_actions = continuity_actions + contextual_actions

        # 중복 제거 및 우선순위 정렬
        unique_actions = []
        seen_commands = set()
        for action in all_actions:
            if action["command"] not in seen_commands:
                unique_actions.append(action)
                seen_commands.add(action["command"])

        # 우선순위별 정렬
        priority_order = {"urgent": 0, "high": 1, "medium": 2, "normal": 3, "low": 4}
        unique_actions.sort(key=lambda x: priority_order.get(x["priority"], 5))
        actions = unique_actions[:5]

        for action in actions:
            priority_indicator = {
                "urgent": "🚨", "high": "🔥", "medium": "📋",
                "normal": "💡", "low": "💭"
            }.get(action["priority"], "💡")

            message_parts.append(
                f"   {priority_indicator} {action['emoji']} {action['title']}: {action['command']}"
            )
            if action.get("description"):
                message_parts.append(f"      └─ {action['description']}")

        # 컨텍스트 팁
        message_parts.extend(["", "💡 Pro Tips:"])

        # 시간 기반 팁
        hour = datetime.now().hour
        if 9 <= hour <= 11:
            message_parts.append("   🌅 Morning Focus: 새로운 SPEC 작성에 최적한 시간")
        elif 14 <= hour <= 16:
            message_parts.append("   ⚡ Afternoon Power: 구현 작업에 집중하기 좋은 시간")
        elif hour >= 18:
            message_parts.append("   🌙 Evening Review: 코드 리뷰와 문서 정리 시간")

        # Constitution 원칙 리마인더
        if pipeline["stage"] == "IMPLEMENT":
            message_parts.append("   🏛️ Constitution: TDD Red-Green-Refactor 사이클 준수")
        elif pipeline["stage"] == "SPECIFY":
            message_parts.append("   🏛️ Constitution: 단순성 원칙 - 모듈 수 ≤ 3개 유지")

        # Constitution 위반 사항 실시간 체크
        violations = self.check_constitution_violations()
        if violations:
            critical_violations = [v for v in violations if v["severity"] == "critical"]
            high_violations = [v for v in violations if v["severity"] == "high"]

            if critical_violations or high_violations:
                message_parts.extend(["", "⚠️ Constitution 위반 감지:"])

                for violation in (critical_violations + high_violations)[:3]:  # 최대 3개만
                    severity_emoji = {"critical": "🚨", "high": "🔥", "medium": "⚠️", "low": "💡"}[violation["severity"]]
                    message_parts.append(f"   {severity_emoji} {violation['article']}: {violation['violation']}")
                    message_parts.append(f"      🔧 자동 수정: {violation['fix_command']}")

        # 팀 상태 표시
        team_status_lines = self.generate_team_status_message()
        if team_status_lines:
            message_parts.extend([""] + team_status_lines)

        # 팀 상태 업데이트 (현재 사용자 정보)
        try:
            import getpass
            current_user = getpass.getuser()
            current_spec = pipeline.get("spec_id")
            self.update_team_status(current_user, current_spec)
        except:
            pass

        return "\n".join(message_parts)

def handle_session_start():
    """SessionStart Hook 메인 핸들러"""
    try:
        # 현재 디렉토리에서 시작해서 프로젝트 루트 찾기
        project_root = _locate_project_root(Path.cwd())
        
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

def _locate_project_root(start: Path) -> Path:
    project_root = start
    depth = 0
    while depth < 10:
        if (project_root / '.claude').exists() or (project_root / '.moai').exists():
            break
        parent = project_root.parent
        if parent == project_root:
            break
        project_root = parent
        depth += 1
    return project_root


def _run_diagnostics() -> None:
    project_root = _locate_project_root(Path.cwd())
    notifier = SessionNotifier(project_root)
    status = notifier.get_project_status()
    if not status["initialized"]:
        print(notifier.generate_simple_init_notice())
    else:
        print(notifier.generate_simple_status_notice(status))


if __name__ == "__main__":
    if len(sys.argv) > 1 and sys.argv[1] == "--diagnostics":
        _run_diagnostics()
    else:
        handle_session_start()

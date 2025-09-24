#!/usr/bin/env python3
"""
MoAI Workflow Orchestrator - 워크플로우 단계별 순서 보장

MoAI 명령어들의 실행 순서를 조율하여 Git 충돌을 방지하고
올바른 워크플로우 진행을 보장합니다.

@TASK:WORKFLOW-ORCHESTRATION-001
"""

import json
import os
import time
from pathlib import Path
from datetime import datetime
from enum import Enum
from typing import Optional, Dict, List
import logging

class WorkflowStage(Enum):
    """워크플로우 단계"""
    PROJECT = "0-project"
    SPEC = "1-spec"
    BUILD = "2-build"
    SYNC = "3-sync"
    DEBUG = "4-debug"

class WorkflowState(Enum):
    """워크플로우 상태"""
    READY = "ready"
    IN_PROGRESS = "in_progress"
    COMPLETED = "completed"
    FAILED = "failed"
    BLOCKED = "blocked"

class WorkflowOrchestrator:
    """워크플로우 오케스트레이션 관리자"""

    def __init__(self, project_root: str = None):
        self.project_root = Path(project_root or os.getcwd())
        self.workflow_dir = self.project_root / ".moai" / "workflow"
        self.workflow_dir.mkdir(parents=True, exist_ok=True)

        self.state_file = self.workflow_dir / "workflow_state.json"
        self.history_file = self.workflow_dir / "workflow_history.json"

        # 로깅 설정
        logging.basicConfig(level=logging.INFO)
        self.logger = logging.getLogger(__name__)

        # 단계별 의존성 정의
        self.stage_dependencies = {
            WorkflowStage.PROJECT: [],  # 의존성 없음
            WorkflowStage.SPEC: [WorkflowStage.PROJECT],  # 프로젝트 먼저
            WorkflowStage.BUILD: [WorkflowStage.SPEC],    # 명세 먼저
            WorkflowStage.SYNC: [WorkflowStage.BUILD],    # 빌드 먼저
            WorkflowStage.DEBUG: []  # 언제든 실행 가능
        }

    def can_execute_stage(self, stage: WorkflowStage, spec_id: str = None) -> tuple[bool, str]:
        """
        워크플로우 단계 실행 가능 여부 확인

        Returns:
            (can_execute: bool, reason: str)
        """
        current_state = self._get_current_state()

        # 의존성 확인
        for dependency in self.stage_dependencies[stage]:
            if not self._is_stage_completed(dependency, spec_id):
                return False, f"의존 단계 미완료: {dependency.value}"

        # 동시 실행 차단 확인
        if self._has_conflicting_stage_running(stage):
            return False, f"충돌 단계 실행 중: {self._get_running_stage()}"

        # SPEC별 순서 확인 (SPEC-ID 있는 경우)
        if spec_id and not self._is_spec_ready_for_stage(spec_id, stage):
            return False, f"SPEC {spec_id}가 {stage.value} 단계 준비 미완료"

        return True, "실행 가능"

    def start_stage(self, stage: WorkflowStage, agent_name: str,
                   spec_id: str = None, description: str = "") -> bool:
        """워크플로우 단계 시작"""

        # 실행 가능성 확인
        can_execute, reason = self.can_execute_stage(stage, spec_id)
        if not can_execute:
            self.logger.error(f"❌ {stage.value} 시작 불가: {reason}")
            return False

        # 단계 상태 업데이트
        stage_info = {
            "stage": stage.value,
            "agent": agent_name,
            "spec_id": spec_id,
            "description": description,
            "state": WorkflowState.IN_PROGRESS.value,
            "started_at": datetime.now().isoformat(),
            "pid": os.getpid()
        }

        self._update_stage_state(stage_info)
        self.logger.info(f"🚀 {stage.value} 시작: {agent_name} ({spec_id or 'N/A'})")

        return True

    def complete_stage(self, stage: WorkflowStage, agent_name: str,
                      spec_id: str = None, success: bool = True) -> bool:
        """워크플로우 단계 완료"""

        current_state = self._get_current_state()
        stage_key = self._get_stage_key(stage, spec_id)

        if stage_key not in current_state:
            self.logger.warning(f"⚠️ 완료할 단계가 없음: {stage.value}")
            return False

        # 단계 완료 처리
        final_state = WorkflowState.COMPLETED if success else WorkflowState.FAILED

        stage_info = current_state[stage_key].copy()
        stage_info.update({
            "state": final_state.value,
            "completed_at": datetime.now().isoformat(),
            "duration_seconds": self._calculate_duration(stage_info)
        })

        self._update_stage_state(stage_info, remove_after=True)

        # 히스토리 저장
        self._save_to_history(stage_info)

        status_emoji = "✅" if success else "❌"
        self.logger.info(f"{status_emoji} {stage.value} 완료: {agent_name}")

        return True

    def get_workflow_status(self) -> Dict:
        """현재 워크플로우 상태 조회"""
        current_state = self._get_current_state()
        recent_history = self._get_recent_history(10)

        return {
            "current_stages": current_state,
            "recent_history": recent_history,
            "next_available": self._get_next_available_stages(),
            "blocked_stages": self._get_blocked_stages()
        }

    def get_spec_workflow_status(self, spec_id: str) -> Dict:
        """특정 SPEC의 워크플로우 상태 조회"""
        history = self._get_spec_history(spec_id)

        completed_stages = []
        current_stage = None

        for record in history:
            stage = WorkflowStage(record["stage"])
            if record["state"] == WorkflowState.COMPLETED.value:
                completed_stages.append(stage)
            elif record["state"] == WorkflowState.IN_PROGRESS.value:
                current_stage = stage

        # 다음 단계 결정
        next_stage = self._determine_next_stage(completed_stages)

        return {
            "spec_id": spec_id,
            "completed_stages": [s.value for s in completed_stages],
            "current_stage": current_stage.value if current_stage else None,
            "next_stage": next_stage.value if next_stage else None,
            "progress_percentage": self._calculate_progress(completed_stages)
        }

    def force_reset_stage(self, stage: WorkflowStage, spec_id: str = None) -> bool:
        """강제로 단계 상태 초기화 (응급용)"""
        current_state = self._get_current_state()
        stage_key = self._get_stage_key(stage, spec_id)

        if stage_key in current_state:
            # 강제 중단 기록
            stage_info = current_state[stage_key].copy()
            stage_info.update({
                "state": WorkflowState.FAILED.value,
                "completed_at": datetime.now().isoformat(),
                "force_reset": True,
                "reset_reason": "강제 초기화"
            })

            self._save_to_history(stage_info)

            # 현재 상태에서 제거
            del current_state[stage_key]
            self._save_current_state(current_state)

            self.logger.warning(f"🚨 {stage.value} 강제 초기화 완료")
            return True

        return False

    def wait_for_stage_completion(self, stage: WorkflowStage, spec_id: str = None,
                                timeout: int = 600) -> bool:
        """특정 단계의 완료 대기"""
        start_time = time.time()
        stage_key = self._get_stage_key(stage, spec_id)

        while time.time() - start_time < timeout:
            current_state = self._get_current_state()

            if stage_key not in current_state:
                # 현재 실행 중이 아니므로 완료된 것으로 간주
                return True

            stage_info = current_state[stage_key]
            if stage_info["state"] in [WorkflowState.COMPLETED.value, WorkflowState.FAILED.value]:
                return stage_info["state"] == WorkflowState.COMPLETED.value

            self.logger.info(f"⏳ {stage.value} 완료 대기 중... ({spec_id or 'global'})")
            time.sleep(5)

        self.logger.error(f"❌ {stage.value} 완료 대기 타임아웃")
        return False

    # Private methods
    def _get_current_state(self) -> Dict:
        """현재 워크플로우 상태 로드"""
        try:
            if self.state_file.exists():
                with open(self.state_file, 'r') as f:
                    return json.load(f)
        except Exception as e:
            self.logger.error(f"상태 로드 실패: {e}")
        return {}

    def _save_current_state(self, state: Dict):
        """현재 워크플로우 상태 저장"""
        try:
            with open(self.state_file, 'w') as f:
                json.dump(state, f, indent=2)
        except Exception as e:
            self.logger.error(f"상태 저장 실패: {e}")

    def _update_stage_state(self, stage_info: Dict, remove_after: bool = False):
        """단계 상태 업데이트"""
        current_state = self._get_current_state()
        stage_key = self._get_stage_key(
            WorkflowStage(stage_info["stage"]),
            stage_info.get("spec_id")
        )

        if remove_after:
            current_state.pop(stage_key, None)
        else:
            current_state[stage_key] = stage_info

        self._save_current_state(current_state)

    def _get_stage_key(self, stage: WorkflowStage, spec_id: str = None) -> str:
        """단계 키 생성"""
        if spec_id:
            return f"{stage.value}:{spec_id}"
        return stage.value

    def _is_stage_completed(self, stage: WorkflowStage, spec_id: str = None) -> bool:
        """단계 완료 여부 확인"""
        history = self._get_recent_history(50)
        stage_key = self._get_stage_key(stage, spec_id)

        for record in reversed(history):
            record_key = self._get_stage_key(
                WorkflowStage(record["stage"]),
                record.get("spec_id")
            )

            if record_key == stage_key:
                return record["state"] == WorkflowState.COMPLETED.value

        return False

    def _has_conflicting_stage_running(self, stage: WorkflowStage) -> bool:
        """충돌 단계 실행 중 확인"""
        current_state = self._get_current_state()

        # Git 작업이 필요한 단계들 (동시 실행 금지)
        git_stages = {WorkflowStage.SPEC, WorkflowStage.BUILD, WorkflowStage.SYNC}

        if stage in git_stages:
            for key, stage_info in current_state.items():
                running_stage = WorkflowStage(stage_info["stage"])
                if (running_stage in git_stages and
                    stage_info["state"] == WorkflowState.IN_PROGRESS.value):
                    return True

        return False

    def _get_running_stage(self) -> Optional[str]:
        """현재 실행 중인 단계 조회"""
        current_state = self._get_current_state()

        for stage_info in current_state.values():
            if stage_info["state"] == WorkflowState.IN_PROGRESS.value:
                return stage_info["stage"]

        return None

    def _save_to_history(self, stage_info: Dict):
        """히스토리 저장"""
        try:
            history = self._get_recent_history(100)
            history.append(stage_info)

            # 최근 100개만 유지
            if len(history) > 100:
                history = history[-100:]

            with open(self.history_file, 'w') as f:
                json.dump(history, f, indent=2)

        except Exception as e:
            self.logger.error(f"히스토리 저장 실패: {e}")

    def _get_recent_history(self, limit: int = 10) -> List[Dict]:
        """최근 히스토리 조회"""
        try:
            if self.history_file.exists():
                with open(self.history_file, 'r') as f:
                    history = json.load(f)
                    return history[-limit:] if limit else history
        except Exception as e:
            self.logger.error(f"히스토리 로드 실패: {e}")
        return []

    def _get_spec_history(self, spec_id: str) -> List[Dict]:
        """특정 SPEC 히스토리 조회"""
        all_history = self._get_recent_history(0)  # 전체 히스토리
        return [
            record for record in all_history
            if record.get("spec_id") == spec_id
        ]

    def _calculate_duration(self, stage_info: Dict) -> float:
        """단계 지속 시간 계산"""
        try:
            start = datetime.fromisoformat(stage_info["started_at"])
            end = datetime.now()
            return (end - start).total_seconds()
        except Exception:
            return 0.0

    def _determine_next_stage(self, completed_stages: List[WorkflowStage]) -> Optional[WorkflowStage]:
        """다음 실행 가능한 단계 결정"""
        workflow_order = [
            WorkflowStage.PROJECT,
            WorkflowStage.SPEC,
            WorkflowStage.BUILD,
            WorkflowStage.SYNC
        ]

        for stage in workflow_order:
            if stage not in completed_stages:
                return stage

        return None  # 모든 단계 완료

    def _calculate_progress(self, completed_stages: List[WorkflowStage]) -> int:
        """진행률 계산"""
        total_stages = 4  # PROJECT, SPEC, BUILD, SYNC
        return int((len(completed_stages) / total_stages) * 100)

    def _get_next_available_stages(self) -> List[str]:
        """다음 실행 가능한 단계들 조회"""
        available = []

        for stage in WorkflowStage:
            can_execute, _ = self.can_execute_stage(stage)
            if can_execute:
                available.append(stage.value)

        return available

    def _get_blocked_stages(self) -> List[Dict]:
        """차단된 단계들과 사유 조회"""
        blocked = []

        for stage in WorkflowStage:
            can_execute, reason = self.can_execute_stage(stage)
            if not can_execute:
                blocked.append({
                    "stage": stage.value,
                    "reason": reason
                })

        return blocked

    def _is_spec_ready_for_stage(self, spec_id: str, stage: WorkflowStage) -> bool:
        """SPEC이 특정 단계 실행 준비가 되었는지 확인"""
        # SPEC별 워크플로우 상태 확인
        spec_status = self.get_spec_workflow_status(spec_id)

        if stage == WorkflowStage.BUILD:
            # BUILD는 SPEC이 완료되어야 함
            return WorkflowStage.SPEC.value in spec_status["completed_stages"]

        elif stage == WorkflowStage.SYNC:
            # SYNC는 BUILD가 완료되어야 함
            return WorkflowStage.BUILD.value in spec_status["completed_stages"]

        return True  # 다른 단계는 특별한 조건 없음

# CLI 인터페이스
if __name__ == "__main__":
    import sys

    if len(sys.argv) < 2:
        print("사용법: workflow_orchestrator.py <command> [args...]")
        print("명령어: start, complete, status, spec-status, wait, reset")
        sys.exit(1)

    orchestrator = WorkflowOrchestrator()
    command = sys.argv[1]

    if command == "start":
        if len(sys.argv) < 4:
            print("사용법: start <stage> <agent_name> [spec_id] [description]")
            sys.exit(1)

        stage = WorkflowStage(sys.argv[2])
        agent_name = sys.argv[3]
        spec_id = sys.argv[4] if len(sys.argv) > 4 else None
        description = sys.argv[5] if len(sys.argv) > 5 else ""

        success = orchestrator.start_stage(stage, agent_name, spec_id, description)
        sys.exit(0 if success else 1)

    elif command == "complete":
        if len(sys.argv) < 4:
            print("사용법: complete <stage> <agent_name> [spec_id] [success=true]")
            sys.exit(1)

        stage = WorkflowStage(sys.argv[2])
        agent_name = sys.argv[3]
        spec_id = sys.argv[4] if len(sys.argv) > 4 else None
        success = sys.argv[5].lower() != "false" if len(sys.argv) > 5 else True

        success = orchestrator.complete_stage(stage, agent_name, spec_id, success)
        sys.exit(0 if success else 1)

    elif command == "status":
        status = orchestrator.get_workflow_status()
        print(json.dumps(status, indent=2))

    elif command == "spec-status":
        if len(sys.argv) < 3:
            print("사용법: spec-status <spec_id>")
            sys.exit(1)

        spec_id = sys.argv[2]
        status = orchestrator.get_spec_workflow_status(spec_id)
        print(json.dumps(status, indent=2))

    else:
        print(f"알 수 없는 명령어: {command}")
        sys.exit(1)
"""
Tests for StatuslineRenderer - Compact 모드에서 기본 정보 포맷팅 기능

"""

from dataclasses import dataclass

import pytest


@dataclass
class StatuslineData:
    """상태줄 데이터 구조"""

    model: str
    duration: str
    directory: str
    version: str
    branch: str
    git_status: str
    active_task: str


class TestStatuslineRendererCompactMode:
    """Compact 모드 렌더러 테스트"""

    def test_compact_render_basic(self):
        """
        GIVEN: StatuslineRenderer 인스턴스와 기본 StatuslineData
        WHEN: render() 메서드를 Compact 모드로 호출
        THEN: 80자 이내이고 "|" 구분자가 포함되며 필수 정보가 모두 포함됨
        Format: 🤖 Model | Ver Version | Git: Branch | GitStatus | Task
        """
        from moai_adk.statusline.renderer import StatuslineRenderer

        renderer = StatuslineRenderer()
        data = StatuslineData(
            model="H 4.5",
            duration="5m",
            directory="MoAI-ADK",
            version="0.20.1",
            branch="feature/SPEC-AUTH-001",
            git_status="+2 M1",
            active_task="[PLAN]",
        )

        result = renderer.render(data, mode="compact")

        # 80자 이내 확인
        assert len(result) <= 80, f"Statusline length {len(result)} exceeds 80 chars: {result}"

        # "|" 구분자 포함 확인
        assert "|" in result, "Statusline must contain pipe separator"

        # 필수 정보가 모두 포함되는지 확인
        assert "H 4.5" in result or "Haiku" in result, "Model info missing"
        assert "0.20.1" in result or "v0.20.1" in result, "Version missing"
        assert "feature/SPEC-AUTH-001" in result or "AUTH-001" in result or "feature" in result, "Branch info missing"
        assert "+2" in result or "M1" in result or "+2 M1" in result, "Git status missing"
        assert "[PLAN]" in result or "PLAN" in result, "Active task missing"

    def test_compact_render_length_constraint(self):
        """
        GIVEN: 다양한 길이의 입력 데이터
        WHEN: render()를 Compact 모드로 호출
        THEN: 항상 80자 이내를 유지함
        """
        from moai_adk.statusline.renderer import StatuslineRenderer

        renderer = StatuslineRenderer()

        test_cases = [
            StatuslineData(
                model="Haiku 4.5",
                duration="1h 30m",
                directory="/Users/goos/MoAI/MoAI-ADK",
                version="v0.20.1",
                branch="feature/SPEC-VERY-LONG-SPEC-NAME-HERE",
                git_status="+10 M5 ?2",
                active_task="[RUN-GREEN]",
            ),
            StatuslineData(
                model="S 4.5",
                duration="30s",
                directory="proj",
                version="0.20.1",
                branch="main",
                git_status="",
                active_task="[IDLE]",
            ),
        ]

        for data in test_cases:
            result = renderer.render(data, mode="compact")
            assert len(result) <= 80, f"Exceeded 80 chars: {len(result)} - {result}"

    def test_compact_render_information_order(self):
        """
        GIVEN: 알려진 값의 StatuslineData
        WHEN: render()를 Compact 모드로 호출
        THEN: 🤖 Model | Ver Version | Git: Branch | Status | Task 순서를 준수
        """
        from moai_adk.statusline.renderer import StatuslineRenderer

        renderer = StatuslineRenderer()
        data = StatuslineData(
            model="H 4.5",
            duration="5m",
            directory="MoAI-ADK",
            version="0.20.1",
            branch="develop",
            git_status="+1",
            active_task="[PLAN]",
        )

        result = renderer.render(data, mode="compact")

        # 정보 순서 검증 (위치 기반)
        model_pos = result.find("H 4.5") if "H 4.5" in result else result.find("Haiku")
        version_pos = result.find("0.20.1") if "0.20.1" in result else result.find("v0.20.1")
        branch_pos = result.find("develop")
        git_pos = result.find("+1")
        task_pos = result.find("[PLAN]") if "[PLAN]" in result else result.find("PLAN")

        # 모든 주요 요소가 찾았는지 확인
        positions = [model_pos, version_pos, branch_pos, git_pos, task_pos]
        assert all(p >= 0 for p in positions), f"Not all elements found in: {result}"

        # 순서 검증: 모델 < 버전 < branch < git < task
        assert model_pos < version_pos < branch_pos, "Model, version, branch order incorrect"
        assert branch_pos < git_pos or branch_pos < task_pos, "Branch position incorrect"


class TestStatuslineRendererDataHandling:
    """데이터 처리 및 엣지 케이스 테스트"""

    def test_render_with_update_indicator(self):
        """
        GIVEN: 업데이트 가능한 상태의 버전 데이터 (버전 + 업데이트 표시)
        WHEN: render()를 호출
        THEN: 버전과 업데이트 아이콘이 함께 표시됨
        """
        from moai_adk.statusline.renderer import StatuslineRenderer

        renderer = StatuslineRenderer()
        data = StatuslineData(
            model="H 4.5",
            duration="5m",
            directory="MoAI-ADK",
            version="0.20.1 ⬆️ 0.21.0",
            branch="main",
            git_status="",
            active_task="[IDLE]",
        )

        result = renderer.render(data, mode="compact")

        assert "0.20.1" in result, "Current version not displayed"
        assert "⬆️" in result or "0.21.0" in result, "Update indicator or new version not shown"
        assert len(result) <= 80, f"Update indicator version exceeds 80 chars: {len(result)}"

    def test_render_empty_git_status(self):
        """
        GIVEN: 깨끗한 Git 상태 (변경 없음)
        WHEN: render()를 호출
        THEN: 공백이나 "clean" 표시로 처리됨
        """
        from moai_adk.statusline.renderer import StatuslineRenderer

        renderer = StatuslineRenderer()
        data = StatuslineData(
            model="H 4.5",
            duration="10m",
            directory="proj",
            version="0.20.1",
            branch="develop",
            git_status="",
            active_task="[IDLE]",
        )

        result = renderer.render(data, mode="compact")

        # 오류 없이 렌더링되어야 함
        assert len(result) <= 80
        assert "develop" in result


if __name__ == "__main__":
    pytest.main([__file__, "-v"])

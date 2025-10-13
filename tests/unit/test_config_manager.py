# @TEST:CORE-TEMPLATE-001 | SPEC: SPEC-CORE-TEMPLATE-001.md
"""ConfigManager 테스트 스위트

config.json 읽기, 쓰기, 업데이트, 깊은 병합 기능을 검증합니다.
"""

import json
import tempfile
from pathlib import Path

import pytest

from moai_adk.core.template.config import ConfigManager


class TestConfigManager:
    """ConfigManager 클래스 테스트"""

    def test_load_returns_default_when_file_not_exists(self, tmp_path: Path) -> None:
        """파일이 없을 때 기본 설정 반환"""
        config_path = tmp_path / "nonexistent" / "config.json"
        manager = ConfigManager(config_path)

        config = manager.load()

        assert config["mode"] == "personal"
        assert config["locale"] == "ko"
        assert config["moai"]["version"] == "0.3.0"

    def test_save_creates_directory_and_file(self, tmp_path: Path) -> None:
        """save()가 디렉토리와 파일을 생성"""
        config_path = tmp_path / "new_dir" / "config.json"
        manager = ConfigManager(config_path)

        manager.save({"test": "value"})

        assert config_path.exists()
        with open(config_path, encoding="utf-8") as f:
            data = json.load(f)
        assert data["test"] == "value"

    def test_save_preserves_korean_characters(self, tmp_path: Path) -> None:
        """save()가 한글을 보존 (ensure_ascii=False)"""
        config_path = tmp_path / "config.json"
        manager = ConfigManager(config_path)

        manager.save({"projectName": "테스트 프로젝트"})

        content = config_path.read_text(encoding="utf-8")
        assert "테스트 프로젝트" in content
        assert "\\u" not in content  # 유니코드 이스케이프 없음

    def test_load_reads_existing_file(self, tmp_path: Path) -> None:
        """load()가 기존 파일을 올바르게 읽음"""
        config_path = tmp_path / "config.json"
        test_data = {"mode": "team", "locale": "en"}
        config_path.write_text(json.dumps(test_data), encoding="utf-8")

        manager = ConfigManager(config_path)
        config = manager.load()

        assert config == test_data

    def test_update_merges_nested_dicts(self, tmp_path: Path) -> None:
        """update()가 중첩 딕셔너리를 병합"""
        config_path = tmp_path / "config.json"
        manager = ConfigManager(config_path)

        # 초기 설정 저장
        manager.save({"git": {"enabled": True, "autoCommit": True}})

        # autoCommit만 업데이트
        manager.update({"git": {"autoCommit": False}})

        config = manager.load()
        assert config["git"]["enabled"] is True  # 유지됨
        assert config["git"]["autoCommit"] is False  # 변경됨

    def test_deep_merge_overwrites_non_dict_values(self, tmp_path: Path) -> None:
        """_deep_merge가 딕셔너리가 아닌 값은 덮어씀"""
        config_path = tmp_path / "config.json"
        manager = ConfigManager(config_path)

        manager.save({"locale": "ko", "mode": "personal"})
        manager.update({"locale": "en"})

        config = manager.load()
        assert config["locale"] == "en"

    def test_deep_merge_adds_new_keys(self, tmp_path: Path) -> None:
        """_deep_merge가 새로운 키를 추가"""
        config_path = tmp_path / "config.json"
        manager = ConfigManager(config_path)

        manager.save({"git": {"enabled": True}})
        manager.update({"git": {"branchPrefix": "feature/"}})

        config = manager.load()
        assert config["git"]["enabled"] is True
        assert config["git"]["branchPrefix"] == "feature/"

    def test_deep_merge_three_levels(self, tmp_path: Path) -> None:
        """_deep_merge가 3단계 중첩 딕셔너리를 병합"""
        config_path = tmp_path / "config.json"
        manager = ConfigManager(config_path)

        manager.save({
            "constitution": {
                "enforce_tdd": True,
                "test_coverage_target": 85
            }
        })
        manager.update({
            "constitution": {
                "test_coverage_target": 90,
                "enforce_spec": True
            }
        })

        config = manager.load()
        assert config["constitution"]["enforce_tdd"] is True  # 유지
        assert config["constitution"]["test_coverage_target"] == 90  # 변경
        assert config["constitution"]["enforce_spec"] is True  # 추가

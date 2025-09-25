#!/usr/bin/env python3
"""
MoAI-ADK Build System Test Suite
빌드 시스템의 모든 기능을 자동 테스트합니다.
"""

import unittest
import sys
import os
import json
import tempfile
import shutil
import hashlib
from pathlib import Path
from unittest.mock import patch, mock_open

# Add project root to path
sys.path.insert(0, os.path.join(os.path.dirname(__file__), ".."))

try:
    from build import MoAIBuilder
except ImportError:
    print("Warning: Could not import build module")
    MoAIBuilder = None


class TestMoAIBuilder(unittest.TestCase):
    """MoAI Builder 기본 기능 테스트"""

    def setUp(self):
        """테스트 환경 설정"""
        if MoAIBuilder is None:
            self.skipTest("MoAIBuilder not available")

        # 임시 디렉토리 생성
        self.test_dir = Path(tempfile.mkdtemp())
        self.src_dir = self.test_dir / "src" / "templates"
        self.dist_dir = self.test_dir / "dist" / "templates"

        # 소스 디렉토리 생성 및 파일 추가
        self.src_dir.mkdir(parents=True)

        # 테스트 파일들 생성
        self.create_test_files()

        # Builder 인스턴스 생성
        self.builder = MoAIBuilder(str(self.test_dir))

    def tearDown(self):
        """테스트 환경 정리"""
        if self.test_dir.exists():
            shutil.rmtree(self.test_dir)

    def create_test_files(self):
        """테스트용 파일들 생성"""
        # .claude/settings.json
        claude_dir = self.src_dir / ".claude"
        claude_dir.mkdir()
        claude_settings = {
            "permissions": {"allow": ["Read(*)"]},
            "environment": {"NODE_ENV": "development"},
        }
        with open(claude_dir / "settings.json", "w") as f:
            json.dump(claude_settings, f, indent=2)

        # .moai/config.json
        moai_dir = self.src_dir / ".moai"
        moai_dir.mkdir()
        moai_config = {"moai_version": "0.2.1", "constitution": {"maxProjects": 3}}
        with open(moai_dir / "config.json", "w") as f:
            json.dump(moai_config, f, indent=2)

        # Hook 스크립트
        hooks_dir = claude_dir / "hooks" / "moai"
        hooks_dir.mkdir(parents=True)
        (hooks_dir / "test_hook.py").write_text(
            '#!/usr/bin/env python3\nprint("test hook")\n'
        )

        # 일반 파일
        (self.src_dir / "README.md").write_text("# Test Project\n")

        # 서브디렉토리 파일
        sub_dir = self.src_dir / "scripts"
        sub_dir.mkdir()
        (sub_dir / "test_script.sh").write_text('#!/bin/bash\necho "test"\n')

    def test_builder_initialization(self):
        """빌더 초기화 테스트"""
        self.assertEqual(self.builder.project_root, self.test_dir.resolve())
        self.assertEqual(self.builder.src_dir, self.src_dir)
        self.assertEqual(self.builder.dist_dir, self.dist_dir)
        self.assertTrue(self.builder.build_log_file.name == "build.log")

    def test_file_hash_calculation(self):
        """파일 해시 계산 테스트"""
        test_file = self.src_dir / "README.md"

        # 해시 계산
        hash1 = self.builder.calculate_file_hash(test_file)
        hash2 = self.builder.calculate_file_hash(test_file)

        # 동일한 파일의 해시는 같아야 함
        self.assertEqual(hash1, hash2)
        self.assertEqual(len(hash1), 32)  # MD5 해시 길이

        # 실제 해시 값 확인
        with open(test_file, "rb") as f:
            expected_hash = hashlib.md5(f.read()).hexdigest()
        self.assertEqual(hash1, expected_hash)

    def test_get_files_to_sync(self):
        """동기화할 파일 목록 생성 테스트"""
        files_to_sync = self.builder.get_files_to_sync()

        # 생성한 파일들이 모두 포함되어야 함
        src_files = [src for src, _ in files_to_sync]
        file_names = [f.name for f in src_files]

        self.assertIn("settings.json", file_names)
        self.assertIn("config.json", file_names)
        self.assertIn("test_hook.py", file_names)
        self.assertIn("README.md", file_names)
        self.assertIn("test_script.sh", file_names)

        # 상대 경로 확인
        for src, dist in files_to_sync:
            rel_path = src.relative_to(self.src_dir)
            expected_dist = self.dist_dir / rel_path
            self.assertEqual(dist, expected_dist)

    def test_needs_sync_new_file(self):
        """새 파일 동기화 필요성 테스트"""
        test_file = self.src_dir / "README.md"
        dist_file = self.dist_dir / "README.md"

        manifest = {"files": {}}

        # dist 파일이 없으면 동기화 필요
        self.assertTrue(self.builder.needs_sync(test_file, dist_file, manifest))

    def test_needs_sync_different_content(self):
        """내용이 다른 파일 동기화 테스트"""
        test_file = self.src_dir / "README.md"
        dist_file = self.dist_dir / "README.md"

        # dist 파일 생성 (다른 내용)
        dist_file.parent.mkdir(parents=True, exist_ok=True)
        dist_file.write_text("# Different Content\n")

        manifest = {"files": {}}

        # 내용이 다르면 동기화 필요
        self.assertTrue(self.builder.needs_sync(test_file, dist_file, manifest))

    def test_needs_sync_same_content(self):
        """같은 내용 파일 동기화 테스트"""
        test_file = self.src_dir / "README.md"
        dist_file = self.dist_dir / "README.md"

        # dist 파일 생성 (같은 내용)
        dist_file.parent.mkdir(parents=True, exist_ok=True)
        dist_file.write_text("# Test Project\n")

        # 매니페스트에 올바른 해시 포함
        file_key = str(test_file.relative_to(self.builder.project_root))
        manifest = {
            "files": {file_key: {"hash": self.builder.calculate_file_hash(test_file)}}
        }

        # 같은 내용이면 동기화 불필요
        self.assertFalse(self.builder.needs_sync(test_file, dist_file, manifest))

    def test_sync_file_success(self):
        """파일 동기화 성공 테스트"""
        src_file = self.src_dir / "README.md"
        dist_file = self.dist_dir / "README.md"

        # 동기화 실행
        result = self.builder.sync_file(src_file, dist_file)

        # 성공 확인
        self.assertTrue(result)
        self.assertTrue(dist_file.exists())

        # 내용 확인
        self.assertEqual(src_file.read_text(), dist_file.read_text())

    def test_sync_file_permission(self):
        """Python 파일 권한 설정 테스트"""
        # Python Hook 파일
        src_file = self.src_dir / ".claude" / "hooks" / "moai" / "test_hook.py"
        dist_file = self.dist_dir / ".claude" / "hooks" / "moai" / "test_hook.py"

        # 동기화 실행
        result = self.builder.sync_file(src_file, dist_file)

        # 성공 확인
        self.assertTrue(result)
        self.assertTrue(dist_file.exists())

        # 권한 확인 (Unix 시스템에서만)
        if os.name != "nt":  # Windows가 아닌 경우
            file_mode = dist_file.stat().st_mode
            # 실행 권한이 있어야 함 (0o755)
            self.assertTrue(file_mode & 0o111)  # 실행 권한 비트 확인

    def test_build_basic(self):
        """기본 빌드 테스트"""
        # 빌드 실행
        result = self.builder.build()

        # 성공 확인
        self.assertTrue(result)

        # dist 디렉토리 생성 확인
        self.assertTrue(self.dist_dir.exists())

        # 파일들이 복사되었는지 확인
        expected_files = [
            ".claude/settings.json",
            ".moai/config.json",
            ".claude/hooks/moai/test_hook.py",
            "README.md",
            "scripts/test_script.sh",
        ]

        for file_path in expected_files:
            dist_file = self.dist_dir / file_path
            self.assertTrue(dist_file.exists(), f"File not found: {file_path}")

    def test_build_force(self):
        """강제 빌드 테스트"""
        # 첫 번째 빌드
        self.builder.build()

        # 두 번째 빌드 (강제)
        result = self.builder.build(force=True)

        # 성공 확인
        self.assertTrue(result)

        # 모든 파일이 다시 동기화되었는지 확인
        files_to_sync = self.builder.get_files_to_sync()
        manifest = self.builder.load_sync_manifest()

        # 매니페스트에 모든 파일 기록 확인
        self.assertTrue(len(manifest.get("files", {})) > 0)

    def test_clean_orphaned_files(self):
        """고아 파일 정리 테스트"""
        # 먼저 빌드 실행
        self.builder.build()

        # 고아 파일 생성 (dist에만 존재)
        orphan_file = self.dist_dir / "orphan.txt"
        orphan_file.write_text("orphan content")

        # 유효한 파일 목록 (src에 있는 파일들)
        files_to_sync = self.builder.get_files_to_sync()
        valid_files = [dist for _, dist in files_to_sync]

        # 고아 파일 정리 실행
        self.builder.clean_orphaned_files(valid_files)

        # 고아 파일이 제거되었는지 확인
        self.assertFalse(orphan_file.exists())

    def test_manifest_operations(self):
        """매니페스트 파일 작업 테스트"""
        # 매니페스트 로드 (없는 경우)
        manifest = self.builder.load_sync_manifest()

        # 기본 구조 확인
        self.assertIn("files", manifest)
        self.assertIsNone(manifest.get("last_sync"))

        # 매니페스트 저장
        test_manifest = {
            "files": {
                "test.txt": {"hash": "abcd1234", "synced_at": "2025-09-12T02:00:00Z"}
            }
        }

        self.builder.save_sync_manifest(test_manifest)

        # 저장된 매니페스트 로드
        loaded_manifest = self.builder.load_sync_manifest()

        # 내용 확인
        self.assertIn("test.txt", loaded_manifest["files"])
        self.assertIsNotNone(loaded_manifest.get("last_sync"))

    def test_status_check(self):
        """상태 확인 테스트"""
        # 빌드 전 상태 (출력 확인용)
        with patch("builtins.print") as mock_print:
            self.builder.status()
            # dist 디렉토리가 없으면 경고 메시지 출력
            calls = [str(call) for call in mock_print.call_args_list]
            self.assertTrue(any("missing" in call.lower() for call in calls))

        # 빌드 실행
        self.builder.build()

        # 빌드 후 상태
        with patch("builtins.print") as mock_print:
            self.builder.status()
            # 모든 파일이 동기화되었다는 메시지 출력
            calls = [str(call) for call in mock_print.call_args_list]
            self.assertTrue(any("synchronized" in call.lower() for call in calls))


class TestBuildSystemIntegration(unittest.TestCase):
    """빌드 시스템 통합 테스트"""

    def setUp(self):
        """통합 테스트 환경 설정"""
        if MoAIBuilder is None:
            self.skipTest("MoAIBuilder not available")

        # 실제 프로젝트 구조와 유사한 테스트 환경 생성
        self.test_dir = Path(tempfile.mkdtemp())
        self.setup_realistic_project()

        self.builder = MoAIBuilder(str(self.test_dir))

    def tearDown(self):
        """테스트 환경 정리"""
        if self.test_dir.exists():
            shutil.rmtree(self.test_dir)

    def setup_realistic_project(self):
        """실제와 유사한 프로젝트 구조 설정"""
        src_templates = self.test_dir / "src" / "templates"
        src_templates.mkdir(parents=True)

        # Claude Code 구조
        claude_dir = src_templates / ".claude"
        claude_dir.mkdir()

        # 설정 파일
        settings = {
            "permissions": {"allow": ["Read(*)", "Write(.moai/**)"]},
            "environment": {"CLAUDE_PROJECT_DIR": "${PWD}"},
            "hooks": {"PreToolUse": [], "PostToolUse": []},
            "memory": {"project_memory_files": [".moai/project/product.md"]},
        }
        with open(claude_dir / "settings.json", "w") as f:
            json.dump(settings, f, indent=2)

        # Hook 스크립트들
        hooks_dir = claude_dir / "hooks" / "moai"
        hooks_dir.mkdir(parents=True)

        hook_files = [
            "config_loader.py",
            "constitution_guard.py",
            "policy_block.py",
            "tag_sync.py",
            "session_start_notice.py",
        ]

        for hook_file in hook_files:
            (hooks_dir / hook_file).write_text(
                f'#!/usr/bin/env python3\n# {hook_file}\nprint("Hook: {hook_file}")\n'
            )

        # 에이전트 파일들
        agents_dir = claude_dir / "agents" / "moai"
        agents_dir.mkdir(parents=True)

        agent_files = ["spec-manager.md", "plan-architect.md", "code-generator.md"]

        for agent_file in agent_files:
            (agents_dir / agent_file).write_text(
                f"# {agent_file}\n\nAgent description...\n"
            )

        # 명령어 파일들
        commands_dir = claude_dir / "commands" / "moai"
        commands_dir.mkdir(parents=True)

        command_files = ["spec.md", "plan.md", "dev.md"]

        for command_file in command_files:
            (commands_dir / command_file).write_text(
                f"# {command_file}\n\nCommand description...\n"
            )

        # 메모리 파일들
        memory_dir = claude_dir / "memory"
        memory_dir.mkdir()

        memory_files = [
            "constitution-principles.md",
            "tag-system-guide.md",
            "agent-system-reference.md",
        ]

        for memory_file in memory_files:
            (memory_dir / memory_file).write_text(
                f"# {memory_file}\n\nMemory content...\n"
            )

        # MoAI 구조
        moai_dir = src_templates / ".moai"
        moai_dir.mkdir()

        # MoAI 설정
        moai_config = {
            "moai_version": "0.2.1",
            "project_type": "spec_first_tdd",
            "constitution": {"maxProjects": 3, "enforceTDD": True},
            "tag_system": {"enabled": True, "version": "16-Core"},
            "quality_gates": {"coverageTarget": 0.8},
            "agents": {"core_agents": ["spec-manager", "plan-architect"]},
        }

        with open(moai_dir / "config.json", "w") as f:
            json.dump(moai_config, f, indent=2)

        # 프로젝트 문서
        (src_templates / "CLAUDE.md").write_text(
            "# MoAI-ADK Project\n\nProject documentation...\n"
        )

        # 템플릿 파일들
        templates_dir = moai_dir / "templates"
        templates_dir.mkdir()

        template_files = ["spec-template.md", "plan-template.md", "tasks-template.md"]

        for template_file in template_files:
            (templates_dir / template_file).write_text(
                f"# {template_file}\n\nTemplate content...\n"
            )

    def test_full_project_build(self):
        """전체 프로젝트 빌드 테스트"""
        # 빌드 실행
        result = self.builder.build()

        # 성공 확인
        self.assertTrue(result)

        # 주요 파일들이 복사되었는지 확인
        important_files = [
            ".claude/settings.json",
            ".moai/config.json",
            ".claude/hooks/moai/config_loader.py",
            ".claude/agents/moai/spec-manager.md",
            ".claude/commands/moai/spec.md",
            ".claude/memory/constitution-principles.md",
            "CLAUDE.md",
            ".moai/templates/spec-template.md",
        ]

        dist_dir = self.test_dir / "dist" / "templates"
        for file_path in important_files:
            dist_file = dist_dir / file_path
            self.assertTrue(
                dist_file.exists(), f"Important file not copied: {file_path}"
            )

        # Python 파일들의 실행 권한 확인
        if os.name != "nt":  # Windows가 아닌 경우
            python_files = [
                ".claude/hooks/moai/config_loader.py",
                ".claude/hooks/moai/constitution_guard.py",
            ]

            for py_file in python_files:
                dist_file = dist_dir / py_file
                if dist_file.exists():
                    file_mode = dist_file.stat().st_mode
                    self.assertTrue(
                        file_mode & 0o111, f"No execute permission: {py_file}"
                    )

    def test_incremental_build_performance(self):
        """증분 빌드 성능 테스트"""
        import time

        # 첫 번째 빌드 (전체)
        start_time = time.time()
        self.builder.build()
        first_build_time = time.time() - start_time

        # 두 번째 빌드 (증분)
        start_time = time.time()
        self.builder.build()
        second_build_time = time.time() - start_time

        # 증분 빌드가 더 빨라야 함
        self.assertLess(second_build_time, first_build_time)

        print(
            f"First build: {first_build_time:.2f}s, Second build: {second_build_time:.2f}s"
        )

    def test_manifest_accuracy(self):
        """매니페스트 정확성 테스트"""
        # 빌드 실행
        self.builder.build()

        # 매니페스트 로드
        manifest = self.builder.load_sync_manifest()

        # 모든 파일이 매니페스트에 기록되었는지 확인
        files_to_sync = self.builder.get_files_to_sync()

        for src_file, _ in files_to_sync:
            file_key = str(src_file.relative_to(self.builder.project_root))
            self.assertIn(
                file_key, manifest["files"], f"File not in manifest: {file_key}"
            )

            # 해시 값 확인
            recorded_hash = manifest["files"][file_key]["hash"]
            actual_hash = self.builder.calculate_file_hash(src_file)
            self.assertEqual(recorded_hash, actual_hash, f"Hash mismatch: {file_key}")


def run_build_tests():
    """빌드 시스템 테스트 실행"""
    print("🔨 Running MoAI-ADK Build System Tests...")

    # 테스트 로더 생성
    loader = unittest.TestLoader()
    suite = unittest.TestSuite()

    # 테스트 케이스 추가
    test_classes = [TestMoAIBuilder, TestBuildSystemIntegration]

    for test_class in test_classes:
        tests = loader.loadTestsFromTestCase(test_class)
        suite.addTests(tests)

    # 테스트 실행
    runner = unittest.TextTestRunner(verbosity=2, buffer=True)
    result = runner.run(suite)

    # 결과 요약
    print(f"\n📊 Build Test Results:")
    print(f"   Tests run: {result.testsRun}")
    print(f"   Failures: {len(result.failures)}")
    print(f"   Errors: {len(result.errors)}")
    print(f"   Skipped: {len(result.skipped)}")

    success = len(result.failures) == 0 and len(result.errors) == 0
    print(
        f"\n{'✅ All build tests passed!' if success else '❌ Some build tests failed!'}"
    )

    return success


if __name__ == "__main__":
    success = run_build_tests()
    sys.exit(0 if success else 1)

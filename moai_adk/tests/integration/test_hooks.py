#!/usr/bin/env python3
"""
MoAI-ADK Hook System Test Suite
Hook 시스템의 모든 구성 요소를 자동 테스트합니다.
"""

import unittest
import sys
import os
import json
import tempfile
import shutil
from pathlib import Path
from unittest.mock import patch, MagicMock

# Add src path for imports
sys.path.insert(0, os.path.join(os.path.dirname(__file__), '..', 'src'))
sys.path.insert(0, os.path.join(os.path.dirname(__file__), '..', 'src', 'templates', '.claude', 'hooks', 'moai'))

# Import hook modules
try:
    from config_loader import MoAIConfigLoader, get_config
except ImportError:
    print("Warning: Could not import config_loader")
    MoAIConfigLoader = None


class TestConfigLoader(unittest.TestCase):
    """config_loader.py 테스트"""
    
    def setUp(self):
        """테스트 환경 설정"""
        if MoAIConfigLoader is None:
            self.skipTest("config_loader not available")
        
        # 임시 디렉토리 생성
        self.test_dir = Path(tempfile.mkdtemp())
        self.claude_dir = self.test_dir / '.claude'
        self.moai_dir = self.test_dir / '.moai'
        
        self.claude_dir.mkdir()
        self.moai_dir.mkdir()
        
        # 테스트용 설정 파일 생성
        self.claude_config = {
            "permissions": {"allow": ["Read(*)", "Write(.moai/**)"]},
            "environment": {"CLAUDE_PROJECT_DIR": "${PWD}"},
            "hooks": {"PreToolUse": []},
            "moai_integration": {
                "config_path": "${CLAUDE_PROJECT_DIR}/.moai/config.json",
                "enabled": True
            }
        }
        
        self.moai_config = {
            "moai_version": "0.1.15",
            "constitution": {
                "maxProjects": 3,
                "enforceTDD": True
            },
            "tag_system": {
                "enabled": True,
                "categories": {
                    "SPEC": {"tags": ["REQ", "DESIGN", "TASK"], "required": True}
                }
            },
            "quality_gates": {
                "coverageTarget": 0.8
            }
        }
        
        # 설정 파일 저장
        with open(self.claude_dir / 'settings.json', 'w') as f:
            json.dump(self.claude_config, f, indent=2)
        
        with open(self.moai_dir / 'config.json', 'w') as f:
            json.dump(self.moai_config, f, indent=2)
    
    def tearDown(self):
        """테스트 환경 정리"""
        if self.test_dir.exists():
            shutil.rmtree(self.test_dir)
    
    def test_config_loader_initialization(self):
        """ConfigLoader 초기화 테스트"""
        loader = MoAIConfigLoader(str(self.test_dir))
        self.assertEqual(loader.project_root, self.test_dir.resolve())
        self.assertTrue(loader.claude_config_path.exists())
        self.assertTrue(loader.moai_config_path.exists())
    
    def test_claude_config_loading(self):
        """Claude 설정 로드 테스트"""
        loader = MoAIConfigLoader(str(self.test_dir))
        config = loader.claude_config
        
        self.assertIn('permissions', config)
        self.assertIn('environment', config)
        self.assertIn('moai_integration', config)
    
    def test_moai_config_loading(self):
        """MoAI 설정 로드 테스트"""
        loader = MoAIConfigLoader(str(self.test_dir))
        config = loader.moai_config
        
        self.assertEqual(config['moai_version'], '0.1.0')
        self.assertIn('constitution', config)
        self.assertIn('tag_system', config)
    
    def test_constitution_config_access(self):
        """Constitution 설정 접근 테스트"""
        loader = MoAIConfigLoader(str(self.test_dir))
        constitution = loader.get_constitution_config()
        
        self.assertEqual(constitution['maxProjects'], 3)
        self.assertTrue(constitution['enforceTDD'])
    
    def test_tag_system_config_access(self):
        """TAG 시스템 설정 접근 테스트"""
        loader = MoAIConfigLoader(str(self.test_dir))
        tag_config = loader.get_tag_system_config()
        
        self.assertTrue(tag_config['enabled'])
        self.assertIn('SPEC', tag_config['categories'])
    
    def test_feature_enabled_check(self):
        """기능 활성화 상태 확인 테스트"""
        loader = MoAIConfigLoader(str(self.test_dir))
        
        self.assertTrue(loader.is_feature_enabled('tag_system'))
        self.assertFalse(loader.is_feature_enabled('non_existent_feature'))
    
    def test_supported_tags_retrieval(self):
        """지원 태그 목록 조회 테스트"""
        loader = MoAIConfigLoader(str(self.test_dir))
        tags = loader.get_supported_tags()
        
        self.assertIn('REQ', tags)
        self.assertIn('DESIGN', tags)
        self.assertIn('TASK', tags)
    
    def test_missing_config_handling(self):
        """설정 파일 누락 처리 테스트"""
        # MoAI 설정 파일 삭제
        (self.moai_dir / 'config.json').unlink()
        
        loader = MoAIConfigLoader(str(self.test_dir))
        config = loader.moai_config
        
        # 빈 딕셔너리 반환 확인
        self.assertEqual(config, {})
    
    def test_invalid_json_handling(self):
        """잘못된 JSON 처리 테스트"""
        # 잘못된 JSON 작성
        with open(self.moai_dir / 'config.json', 'w') as f:
            f.write('{ invalid json }')
        
        loader = MoAIConfigLoader(str(self.test_dir))
        config = loader.moai_config
        
        # 빈 딕셔너리 반환 확인
        self.assertEqual(config, {})


class TestConstitutionGuard(unittest.TestCase):
    """Constitution Guard 테스트"""
    
    def setUp(self):
        """테스트 환경 설정"""
        self.test_dir = Path(tempfile.mkdtemp())
        self.moai_dir = self.test_dir / '.moai'
        self.moai_dir.mkdir()
        
        # 테스트용 Constitution 설정
        self.constitution_config = {
            "constitution": {
                "maxProjects": 3,
                "requireLibraries": True,
                "enforceTDD": True,
                "requireObservability": True,
                "versioningScheme": "MAJOR.MINOR.BUILD"
            }
        }
        
        with open(self.moai_dir / 'config.json', 'w') as f:
            json.dump(self.constitution_config, f)
    
    def tearDown(self):
        """테스트 환경 정리"""
        if self.test_dir.exists():
            shutil.rmtree(self.test_dir)
    
    def test_simplicity_check_basic(self):
        """기본 Simplicity 원칙 테스트"""
        # package.json 파일 생성 (2개)
        (self.test_dir / 'package.json').touch()
        (self.test_dir / 'frontend' / 'package.json').parent.mkdir()
        (self.test_dir / 'frontend' / 'package.json').touch()
        
        # ConstitutionGuard import 시도
        try:
            sys.path.insert(0, os.path.join(os.path.dirname(__file__), '..', 'src', 'templates', '.claude', 'hooks', 'moai'))
            from constitution_guard import ConstitutionGuard
            
            guard = ConstitutionGuard(self.test_dir)
            result, message = guard.check_simplicity()
            
            # 2개 프로젝트이므로 통과해야 함
            self.assertTrue(result)
            self.assertIn("2", message)
        except ImportError:
            self.skipTest("ConstitutionGuard not available")
    
    def test_simplicity_check_violation(self):
        """Simplicity 원칙 위반 테스트"""
        # package.json 파일 생성 (4개 - 위반)
        packages = ['package.json', 'frontend/package.json', 'backend/package.json', 'mobile/package.json']
        for pkg in packages:
            pkg_path = self.test_dir / pkg
            pkg_path.parent.mkdir(parents=True, exist_ok=True)
            pkg_path.touch()
        
        try:
            sys.path.insert(0, os.path.join(os.path.dirname(__file__), '..', 'src', 'templates', '.claude', 'hooks', 'moai'))
            from constitution_guard import ConstitutionGuard
            
            guard = ConstitutionGuard(self.test_dir)
            result, message = guard.check_simplicity()
            
            # 4개 프로젝트이므로 실패해야 함
            self.assertFalse(result)
            self.assertIn("4", message)
            self.assertIn("3", message)  # 최대 허용 개수
        except ImportError:
            self.skipTest("ConstitutionGuard not available")


class TestBuildSystem(unittest.TestCase):
    """Build System 테스트"""
    
    def setUp(self):
        """테스트 환경 설정"""
        self.test_dir = Path(tempfile.mkdtemp())
        self.src_dir = self.test_dir / 'src' / 'templates'
        self.dist_dir = self.test_dir / 'dist' / 'templates'
        
        # 디렉토리 생성
        self.src_dir.mkdir(parents=True)
        
        # 테스트 파일 생성
        (self.src_dir / 'test1.txt').write_text('content1')
        (self.src_dir / 'subdir').mkdir()
        (self.src_dir / 'subdir' / 'test2.txt').write_text('content2')
    
    def tearDown(self):
        """테스트 환경 정리"""
        if self.test_dir.exists():
            shutil.rmtree(self.test_dir)
    
    def test_build_system_import(self):
        """빌드 시스템 모듈 임포트 테스트"""
        try:
            sys.path.insert(0, os.path.join(os.path.dirname(__file__), '..'))
            from build import MoAIBuilder
            
            builder = MoAIBuilder(str(self.test_dir))
            self.assertEqual(builder.project_root, self.test_dir.resolve())
            self.assertEqual(builder.src_dir, self.src_dir)
            self.assertEqual(builder.dist_dir, self.dist_dir)
        except ImportError:
            self.skipTest("Build system not available")
    
    def test_file_sync_detection(self):
        """파일 동기화 감지 테스트"""
        try:
            sys.path.insert(0, os.path.join(os.path.dirname(__file__), '..'))
            from build import MoAIBuilder
            
            builder = MoAIBuilder(str(self.test_dir))
            files_to_sync = builder.get_files_to_sync()
            
            # 2개 파일이 감지되어야 함
            self.assertEqual(len(files_to_sync), 2)
            
            file_names = [str(src.name) for src, _ in files_to_sync]
            self.assertIn('test1.txt', file_names)
            self.assertIn('test2.txt', file_names)
        except ImportError:
            self.skipTest("Build system not available")
    
    def test_hash_calculation(self):
        """파일 해시 계산 테스트"""
        try:
            sys.path.insert(0, os.path.join(os.path.dirname(__file__), '..'))
            from build import MoAIBuilder
            
            builder = MoAIBuilder(str(self.test_dir))
            
            test_file = self.src_dir / 'test1.txt'
            hash1 = builder.calculate_file_hash(test_file)
            hash2 = builder.calculate_file_hash(test_file)
            
            # 같은 파일의 해시는 동일해야 함
            self.assertEqual(hash1, hash2)
            self.assertTrue(len(hash1) == 32)  # MD5 해시 길이
        except ImportError:
            self.skipTest("Build system not available")


class TestIntegration(unittest.TestCase):
    """통합 테스트"""
    
    def setUp(self):
        """통합 테스트 환경 설정"""
        self.test_dir = Path(tempfile.mkdtemp())
        
        # 완전한 MoAI-ADK 구조 생성
        self.setup_complete_structure()
    
    def tearDown(self):
        """테스트 환경 정리"""
        if self.test_dir.exists():
            shutil.rmtree(self.test_dir)
    
    def setup_complete_structure(self):
        """완전한 MoAI-ADK 구조 설정"""
        # 디렉토리 구조
        dirs = [
            '.claude', '.claude/hooks/moai', '.claude/commands/moai', '.claude/agents/moai', '.claude/memory',
            '.moai', '.moai/steering', '.moai/specs', '.moai/indexes',
            'src/templates', 'dist/templates'
        ]
        
        for d in dirs:
            (self.test_dir / d).mkdir(parents=True, exist_ok=True)
        
        # 기본 설정 파일
        claude_settings = {
            "permissions": {"allow": ["Read(*)", "Write(.moai/**)"]},
            "environment": {"CLAUDE_PROJECT_DIR": str(self.test_dir)},
            "moai_integration": {"enabled": True}
        }
        
        moai_config = {
            "moai_version": "0.1.15",
            "constitution": {"maxProjects": 3, "enforceTDD": True},
            "tag_system": {"enabled": True, "categories": {"SPEC": {"tags": ["REQ"]}}},
            "quality_gates": {"coverageTarget": 0.8}
        }
        
        with open(self.test_dir / '.claude' / 'settings.json', 'w') as f:
            json.dump(claude_settings, f, indent=2)
        
        with open(self.test_dir / '.moai' / 'config.json', 'w') as f:
            json.dump(moai_config, f, indent=2)
    
    def test_full_integration(self):
        """전체 시스템 통합 테스트"""
        # config_loader 테스트
        if MoAIConfigLoader is not None:
            loader = MoAIConfigLoader(str(self.test_dir))
            
            # 설정 로드 확인
            self.assertTrue(len(loader.claude_config) > 0)
            self.assertTrue(len(loader.moai_config) > 0)
            
            # Constitution 설정 확인
            constitution = loader.get_constitution_config()
            self.assertEqual(constitution.get('maxProjects'), 3)
            
            # TAG 시스템 설정 확인
            tag_config = loader.get_tag_system_config()
            self.assertTrue(tag_config.get('enabled', False))


def run_hook_tests():
    """Hook 테스트 실행"""
    print("🧪 Running MoAI-ADK Hook System Tests...")
    
    # 테스트 로더 생성
    loader = unittest.TestLoader()
    suite = unittest.TestSuite()
    
    # 테스트 케이스 추가
    test_classes = [
        TestConfigLoader,
        TestConstitutionGuard,
        TestBuildSystem,
        TestIntegration
    ]
    
    for test_class in test_classes:
        tests = loader.loadTestsFromTestCase(test_class)
        suite.addTests(tests)
    
    # 테스트 실행
    runner = unittest.TextTestRunner(verbosity=2, buffer=True)
    result = runner.run(suite)
    
    # 결과 요약
    print(f"\n📊 Test Results:")
    print(f"   Tests run: {result.testsRun}")
    print(f"   Failures: {len(result.failures)}")
    print(f"   Errors: {len(result.errors)}")
    print(f"   Skipped: {len(result.skipped)}")
    
    if result.failures:
        print("\n❌ Failures:")
        for test, traceback in result.failures:
            print(f"   {test}: {traceback}")
    
    if result.errors:
        print("\n💥 Errors:")
        for test, traceback in result.errors:
            print(f"   {test}: {traceback}")
    
    if result.skipped:
        print(f"\n⏭️ Skipped: {len(result.skipped)} tests")
    
    success = len(result.failures) == 0 and len(result.errors) == 0
    print(f"\n{'✅ All tests passed!' if success else '❌ Some tests failed!'}")
    
    return success


if __name__ == '__main__':
    success = run_hook_tests()
    sys.exit(0 if success else 1)
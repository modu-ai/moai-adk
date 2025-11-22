# Claude Code Hooks - 실전 예제

## 예제 1: Pre-commit 훅

```python
# hooks/pre_commit_hook.py
from moai_cc_hooks import Hook, HookRegistry

class PreCommitHook(Hook):
    """커밋 전 검증 훅"""
    
    event = "pre-commit"
    priority = 100
    
    async def execute(self, context):
        """커밋 전 검증"""
        
        # 1. 린팅 검증
        lint_result = await self.run_linter()
        if not lint_result.passed:
            return {
                "success": False,
                "message": f"린팅 실패: {lint_result.errors}"
            }
        
        # 2. 테스트 실행
        test_result = await self.run_tests()
        if not test_result.passed:
            return {
                "success": False,
                "message": f"테스트 실패: {test_result.failures}"
            }
        
        # 3. 타입 검증
        type_check = await self.run_type_checker()
        if not type_check.passed:
            return {
                "success": False,
                "message": f"타입 검증 실패: {type_check.errors}"
            }
        
        return {"success": True}
    
    async def run_linter(self):
        # 린팅 로직
        pass
    
    async def run_tests(self):
        # 테스트 로직
        pass
    
    async def run_type_checker(self):
        # 타입 검증 로직
        pass

# 등록
registry = HookRegistry()
registry.register(PreCommitHook())
```

## 예제 2: Post-merge 훅

```python
# hooks/post_merge_hook.py
from moai_cc_hooks import Hook

class PostMergeHook(Hook):
    """병합 후 자동화 훅"""
    
    event = "post-merge"
    
    async def execute(self, context):
        """병합 후 처리"""
        
        merged_branch = context['source_branch']
        
        # 1. 의존성 업데이트
        await self.update_dependencies()
        
        # 2. 문서 생성
        await self.generate_documentation()
        
        # 3. 배포 트리거
        if merged_branch == "main":
            await self.trigger_deployment()
        
        return {"success": True}
```

## 예제 3: Custom 훅

```python
# hooks/custom_validation_hook.py
from moai_cc_hooks import Hook

class CustomValidationHook(Hook):
    """커스텀 검증 훅"""
    
    event = "pre-push"
    
    async def execute(self, context):
        """푸시 전 커스텀 검증"""
        
        # SPEC 파일 존재 확인
        spec_files = await self.find_spec_files()
        if not spec_files:
            return {
                "success": False,
                "message": "SPEC 파일이 필요합니다"
            }
        
        # 테스트 커버리지 확인
        coverage = await self.check_coverage()
        if coverage < 80:
            return {
                "success": False,
                "message": f"테스트 커버리지가 부족합니다: {coverage}%"
            }
        
        return {"success": True}
```

## 예제 4: 훅 체인

```python
# hooks/hook_chain.py
from moai_cc_hooks import HookChain

class ValidationChain(HookChain):
    """검증 체인"""
    
    async def execute(self):
        """검증 체인 실행"""
        
        hooks = [
            LintingHook(),
            TestingHook(),
            SecurityScanHook(),
            CoverageCheckHook()
        ]
        
        results = []
        for hook in hooks:
            result = await hook.execute()
            results.append(result)
            
            # 실패 시 중단
            if not result.success:
                return {
                    "success": False,
                    "failed_at": hook.name,
                    "results": results
                }
        
        return {"success": True, "results": results}
```

## 예제 5: 비동기 훅

```python
# hooks/async_notification_hook.py
from moai_cc_hooks import Hook

class AsyncNotificationHook(Hook):
    """비동기 알림 훅"""
    
    event = "post-deploy"
    async_execution = True
    
    async def execute(self, context):
        """배포 후 비동기 알림"""
        
        # 슬랙 알림
        await self.send_slack_notification(
            f"배포 완료: {context['version']}"
        )
        
        # 메일 알림
        await self.send_email_notification(
            recipients=context['team'],
            message=f"버전 {context['version']} 배포됨"
        )
        
        # 모니터링 업데이트
        await self.update_monitoring(
            version=context['version'],
            status="deployed"
        )
        
        return {"success": True}
```

## 예제 6: 조건부 훅

```python
# hooks/conditional_hook.py
from moai_cc_hooks import Hook

class ConditionalHook(Hook):
    """조건부 실행 훅"""
    
    event = "pre-build"
    
    async def should_execute(self, context):
        """실행 조건 확인"""
        
        # main 브랜치에서만 실행
        if context['branch'] != 'main':
            return False
        
        # 코드 변경이 있을 때만 실행
        if not context['files_changed']:
            return False
        
        return True
    
    async def execute(self, context):
        """조건 충족 시 실행"""
        
        if not await self.should_execute(context):
            return {"skipped": True}
        
        await self.run_full_build()
        
        return {"success": True}
```

---

**Last Updated**: 2025-11-22  
**Status**: Production Ready

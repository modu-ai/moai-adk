# Claude Code Commands - 실전 예제

## 예제 1: 프로젝트 초기화 명령어

```python
# commands/project_init.py
from moai_cc_commands import Command, CommandRegistry

class ProjectInitCommand(Command):
    """프로젝트 초기화 명령어"""
    
    name = "project-init"
    description = "프로젝트 초기화 및 구성"
    
    def setup_parameters(self):
        self.add_parameter("project_name", "프로젝트 이름")
        self.add_parameter("--template", "프로젝트 템플릿 선택")
        self.add_parameter("--skip-git", "Git 초기화 건너뛰기")
    
    async def execute(self, params):
        # 프로젝트 디렉토리 생성
        await self.create_directory(params.project_name)
        
        # 템플릿 적용
        if params.template:
            await self.apply_template(params.template)
        
        # Git 저장소 초기화
        if not params.skip_git:
            await self.init_git_repo()
        
        return {
            "status": "success",
            "project": params.project_name,
            "message": "프로젝트 초기화 완료"
        }

# 등록
registry = CommandRegistry()
registry.register(ProjectInitCommand())
```

## 예제 2: 워크플로우 오케스트레이션

```python
# workflows/deployment.py
from moai_cc_commands import Workflow, WorkflowStep

class DeploymentWorkflow(Workflow):
    """배포 워크플로우"""
    
    name = "deploy"
    description = "애플리케이션 배포"
    
    def setup_steps(self):
        # Step 1: 빌드
        self.add_step(WorkflowStep(
            name="build",
            command="build",
            description="애플리케이션 빌드"
        ))
        
        # Step 2: 테스트
        self.add_step(WorkflowStep(
            name="test",
            command="test",
            description="테스트 실행",
            depends_on=["build"]
        ))
        
        # Step 3: 배포
        self.add_step(WorkflowStep(
            name="deploy",
            command="deploy",
            description="프로덕션 배포",
            depends_on=["test"],
            on_failure="rollback"
        ))
    
    async def on_step_complete(self, step_name, result):
        """단계 완료 시 처리"""
        if step_name == "test" and result.status == "failed":
            await self.notify_team(f"테스트 실패: {result.error}")
```

## 예제 3: 파라미터 검증

```python
# validators/project_validator.py
from moai_cc_commands import Validator

class ProjectNameValidator(Validator):
    """프로젝트 이름 검증"""
    
    def validate(self, value):
        # 빈 값 확인
        if not value:
            raise ValueError("프로젝트 이름이 필요합니다")
        
        # 길이 확인
        if len(value) < 3:
            raise ValueError("프로젝트 이름은 3자 이상이어야 합니다")
        
        # 문자 확인
        if not value.isalnum():
            raise ValueError("프로젝트 이름은 알파벳과 숫자만 사용 가능합니다")
        
        return value
```

## 예제 4: 명령어 체이닝

```python
# commands/chained_commands.py
from moai_cc_commands import CommandChain

async def run_deployment_pipeline():
    """배포 파이프라인 실행"""
    
    chain = CommandChain()
    
    # 체인 구성
    chain.add_command("/moai:2-run", {"spec_id": "SPEC-001"})
    chain.add_command("/moai:3-sync", {"spec_id": "SPEC-001"})
    chain.add_command("/deploy", {"environment": "production"})
    
    # 순차 실행
    result = await chain.execute()
    
    return result
```

## 예제 5: CLI 헬프 시스템

```python
# cli/help_system.py
from moai_cc_commands import Command

class HelpCommand(Command):
    """헬프 명령어"""
    
    name = "help"
    description = "명령어 도움말"
    
    def format_help(self, command):
        """명령어 헬프 포맷"""
        return f"""
        명령어: {command.name}
        설명: {command.description}
        
        파라미터:
        {self._format_parameters(command)}
        
        사용 예:
        {command.usage_example}
        """
    
    async def execute(self, params):
        command_name = params.get("command")
        command = self.registry.get(command_name)
        
        if not command:
            return {"error": f"명령어를 찾을 수 없습니다: {command_name}"}
        
        return {"help": self.format_help(command)}
```

## 예제 6: 에러 처리 및 재시도

```python
# error_handling/retry_logic.py
from moai_cc_commands import CommandExecutor

class ResilientCommandExecutor(CommandExecutor):
    """탄력적 명령어 실행기"""
    
    async def execute_with_retry(self, command, params, max_retries=3):
        """재시도 로직 포함 실행"""
        
        last_error = None
        
        for attempt in range(max_retries):
            try:
                result = await command.execute(params)
                return result
                
            except Exception as e:
                last_error = e
                
                # 재시도 전 대기
                if attempt < max_retries - 1:
                    wait_time = 2 ** attempt  # 지수 백오프
                    await asyncio.sleep(wait_time)
        
        # 모든 재시도 실패
        raise last_error
```

## 예제 7: 진행상황 추적

```python
# progress/progress_tracker.py
from moai_cc_commands import ProgressTracker

class WorkflowProgressTracker(ProgressTracker):
    """워크플로우 진행상황 추적"""
    
    async def track_workflow(self, workflow):
        """워크플로우 진행상황 추적"""
        
        total_steps = len(workflow.steps)
        
        for idx, step in enumerate(workflow.steps):
            # 진행상황 업데이트
            progress = (idx / total_steps) * 100
            
            self.update(
                current=idx + 1,
                total=total_steps,
                percentage=progress,
                message=f"실행 중: {step.name}"
            )
            
            # 단계 실행
            result = await step.execute()
            
            self.log_result(step.name, result)
```

---

**Last Updated**: 2025-11-22  
**Status**: Production Ready

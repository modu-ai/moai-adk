# Claude Code Commands - 참고 자료

## API Reference

### Command 클래스

```python
class Command:
    """기본 명령어 클래스"""
    
    name: str  # 명령어 이름
    description: str  # 명령어 설명
    
    def setup_parameters(self) -> None:
        """파라미터 설정"""
    
    async def execute(self, params: CommandParams) -> dict:
        """명령어 실행"""
    
    def validate(self, params: CommandParams) -> bool:
        """파라미터 검증"""
```

### CommandRegistry

```python
registry = CommandRegistry()
registry.register(command)  # 명령어 등록
command = registry.get(name)  # 명령어 조회
```

### Workflow 클래스

```python
class Workflow:
    """워크플로우 기본 클래스"""
    
    name: str
    description: str
    
    def add_step(self, step: WorkflowStep) -> None:
        """단계 추가"""
    
    async def execute(self) -> WorkflowResult:
        """워크플로우 실행"""
```

## 명령어 목록

| 명령어 | 설명 | 파라미터 |
|--------|------|---------|
| `/moai:0-project` | 프로젝트 초기화 | 없음 |
| `/moai:1-plan` | SPEC 생성 | 설명 문자열 |
| `/moai:2-run` | TDD 구현 | SPEC ID |
| `/moai:3-sync` | 문서 동기화 | SPEC ID |

## 에러 코드

| 코드 | 설명 |
|------|------|
| `CMD_001` | 명령어를 찾을 수 없음 |
| `CMD_002` | 파라미터 검증 실패 |
| `CMD_003` | 명령어 실행 실패 |
| `WF_001` | 워크플로우 실행 실패 |

---

**Last Updated**: 2025-11-22

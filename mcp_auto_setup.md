# MoAI-ADK MCP 자동 설정 방안

## 🎯 목표
MoAI-ADK 패키지 설치 시 3가지 핵심 MCP 서버를 자동으로 설치하고 설정하도록 개선

## 📋 구현 계획

### 단계 1: MCP 서버 탐지 기능
```python
# src/moai_adk/core/mcp/setup.py
class MCPAutoSetup:
    def detect_installed_mcp_servers(self) -> dict:
        """설치된 MCP 서버 탐지"""
        return {
            "context7": self.check_npm_package("@upstash/context7-mcp"),
            "figma": self.check_npm_package("figma-mcp-pro"),
            "playwright": self.check_npm_package("@playwright/mcp")
        }

    def auto_install_mcp_servers(self, servers: list) -> bool:
        """선택된 MCP 서버 자동 설치"""
        for server in servers:
            if server == "context7":
                self.install_package("@upstash/context7-mcp")
            elif server == "figma":
                self.install_package("figma-mcp-pro")
            elif server == "playwright":
                self.install_package("@playwright/mcp")
```

### 단계 2: 설정 파일 자동 생성
```python
def generate_mcp_config(self, servers: dict) -> dict:
    """MCP 설정 자동 생성"""
    mcp_config = {}

    if servers.get("context7", {}).get("installed"):
        mcp_config["context7"] = {
            "command": "node",
            "args": [f"{self.get_npm_global_path()}/@upstash/context7-mcp/dist/index.js"],
            "env": {}
        }

    if servers.get("figma", {}).get("installed"):
        mcp_config["figma"] = {
            "command": "node",
            "args": [f"{self.get_npm_global_path()}/figma-mcp-pro/dist/index.js"],
            "env": {
                "FIGMA_ACCESS_TOKEN": "${FIGMA_ACCESS_TOKEN}"
            }
        }

    if servers.get("playwright", {}).get("installed"):
        mcp_config["playwright"] = {
            "command": "node",
            "args": [f"{self.get_npm_global_path()}/@playwright/mcp/dist/index.js"],
            "env": {}
        }

    return mcp_config
```

### 단계 3: moai-adk init 명령어 확장
```python
# src/moai_adk/cli/commands/init.py 기존 코드에 추가
@click.option(
    "--with-mcp",
    multiple=True,
    type=click.Choice(["context7", "figma", "playwright"]),
    help="Install MCP servers automatically"
)
@click.option(
    "--mcp-auto",
    is_flag=True,
    help="Auto-install all recommended MCP servers"
)
def init(
    project_name: str,
    with_mcp: tuple,
    mcp_auto: bool,
    **kwargs
):
    # 기존 초기화 로직...

    # MCP 자동 설정
    if mcp_auto or with_mcp:
        mcp_setup = MCPAutoSetup()

        servers_to_install = ["context7", "figma", "playwright"] if mcp_auto else list(with_mcp)

        with Progress(SpinnerColumn(), TextColumn("[progress.description]")) as progress:
            task = progress.add_task("설치 중인 MCP 서버 탐지...", total=None)

            # 설치 상태 확인
            installed_servers = mcp_setup.detect_installed_mcp_servers()

            for server in servers_to_install:
                progress.update(task, description=f"{server} MCP 서버 설치 중...")
                if not installed_servers.get(server, {}).get("installed"):
                    mcp_setup.auto_install_mcp_servers([server])

            progress.update(task, description="MCP 설정 파일 생성 중...")

            # 설정 파일 업데이트
            final_config = mcp_setup.generate_mcp_config(installed_servers)
            mcp_setup.update_settings_file(final_config)

            progress.update(task, description="✅ MCP 설정 완료")
```

## 🔧 사용자 인터페이스

### CLI 사용 예시
```bash
# 추천 MCP 모두 자동 설치
moai-adk init my-project --mcp-auto

# 특정 MCP만 선택 설치
moai-adk init my-project --with-mcp context7 --with-mcp figma

# 기존 프로젝트에 MCP 추가
moai-adk mcp-setup --auto
moai-adk mcp-setup --add context7 playwright
```

### 대화형 설정
```python
@click.command()
@click.option("--auto", is_flag=True, help="Auto-setup recommended MCP servers")
def mcp_setup(auto: bool):
    """Setup MCP servers for MoAI-ADK project"""

    if auto:
        servers = ["context7", "figma", "playwright"]
    else:
        # 대화형 선택
        servers = prompt_for_mcp_servers()

    setup = MCPAutoSetup()
    setup.configure_servers(servers)
```

## 📦 패키지 설치 시 자동화

### setup.py 수정
```python
# setup.py
entry_points={
    "console_scripts": [
        "moai-adk=moai_adk.cli.main:cli",
    ],
    # MCP 설치 후크 추가
    "moai_adk.mcp": [
        "context7=@upstash/context7-mcp",
        "figma=figma-mcp-pro",
        "playwright=@playwright/mcp",
    ]
}
```

## 🎯 기대 효과

1. **일관성**: 모든 MoAI-ADK 프로젝트가 동일한 MCP 환경 보장
2. **간편성**: 한 번의 명령어로 MCP 자동 설치 및 설정
3. **유연성**: 필요한 MCP만 선택적으로 설치 가능
4. **안정성**: 공식 패키지 사용으로 신뢰성 보장

## 📋 구현 순서

1. **Week 1**: MCP 탐지 및 설치 기능 구현
2. **Week 2**: CLI 명령어 확장 및 UI 개발
3. **Week 3**: 테스트 및 문서화 완성